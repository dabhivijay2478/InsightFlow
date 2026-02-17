# MANTrixFlow — Meltano-as-an-Engine Refactoring Plan

> **Scope**: Migrate from manual Singer tap execution to a **Metadata-Driven Orchestration** model with Meltano CLI as the execution engine. This plan avoids removing necessary code; it adds orchestration layers and refactors where needed.

---

## 1. Current State Summary

### What Exists Today

| Component | Location | Purpose |
|-----------|----------|---------|
| **Manual tap execution** | `main.py` | Imports and runs `tap_postgres.main()`, `tap_mysql`, `tap_mongodb` directly via subprocess |
| **Singer config builder** | `utils.py` `build_singer_config()` | Maps `connection_config` → Singer tap config |
| **Discovery** | `main.py` `_run_discovery_sync()` | Runs `tap --discover`, parses output |
| **Collect** | `main.py` `_run_collect_sync()` | Runs tap with catalog, parses RECORD/STATE |
| **Emit** | `main.py` `_emit_to_sql()`, `_emit_to_mongodb()` | Direct SQLAlchemy / PyMongo writes |
| **Pipeline endpoint** | `main.py` `/run-meltano-pipeline` | Collect → (optional transform) → Emit; **does not use Meltano CLI** |
| **CDC logic** | `connectors/tap-postgres/.../logical_replication.py` | WAL-based CDC; `connectors/tap-mysql/.../binlog.py`; MongoDB oplog |
| **meltano.yml** | `apps/etl/meltano.yml` | Static job `mongodb-to-postgres`; env vars from `.env` |
| **Utils** | `utils.py` | `parse_discovery_output`, `parse_singer_stream`, `merge_state`, `select_catalog_streams`, etc. |

### What to Keep (No Removal)

- **Connector implementations** (`tap-postgres`, `tap-mysql`, `tap-mongodb`) — required by Singer/Meltano
- **CDC strategies** (`logical_replication.py`, `binlog.py`, `oplog.py`) — used by taps internally
- **`transformer.py`** — safe sandbox for user transforms; can be invoked by Meltano or kept as API
- **`utils.py`** — `parse_discovery_output`, `parse_singer_stream`, `merge_state`, `chunked`, `temporary_json_file` remain useful; `build_singer_config` may be repurposed
- **API surface** — `/discover-schema`, `/collect`, `/emit`, `/transform`, `/test-connection`, `/run-meltano-pipeline` — **contract preserved**; implementation may delegate to Meltano
- **NestJS** — Management layer unchanged; continues to store pipeline definitions and call FastAPI

### What Is Redundant / Can Be Replaced (Not Deleted Until New Path Works)

- Direct `_tap_command()` / `_run_discovery_sync` / `_run_collect_sync` — **replaced by** `meltano invoke <tap> --discover` and `meltano run <job>`
- Custom `_emit_to_sql` / `_emit_to_mongodb` in `run-meltano-pipeline` — **replaced by** Meltano loaders when using full `tap → target → dbt` flow
- Hardcoded `source_type` branches in `_tap_command()` — **replaced by** generic plugin mapping

---

## 2. Target Architecture

### Layered Model

```
┌─────────────────────────────────────────────────────────────────┐
│  MANAGEMENT (NestJS)                                             │
│  - Auth, orgs, pipeline definitions, schedules                    │
│  - Sends: job_id, connection_config, source_type, dest_type      │
└──────────────────────────────┬──────────────────────────────────┘
                               │ HTTP
┌──────────────────────────────▼──────────────────────────────────┐
│  ORCHESTRATION (FastAPI)                                         │
│  - Maps connection_config → MELTANO_EXTRACTOR_* env vars         │
│  - Runs: meltano run <job> or meltano invoke <tap> --discover     │
│  - Captures exit codes, state, errors                            │
└──────────────────────────────┬──────────────────────────────────┘
                               │ subprocess
┌──────────────────────────────▼──────────────────────────────────┐
│  EXECUTION (Meltano + dbt)                                        │
│  - Tap (extract) → Target (load) → dbt (transform)               │
│  - State in meltano.db; CDC LSN/Binlog/Oplog handled by taps      │
└─────────────────────────────────────────────────────────────────┘
```

### Single Source of Truth: `meltano.yml`

- All extractors, loaders, transformers declared in YAML
- New connectors = add plugin block + optional mapping in DB
- Meltano handles install, versioning, venv isolation

---

## 3. Implementation Phases

### Phase 1: Meltano Job Runner Utility (No Breaking Changes)

**Goal**: Add a generic runner that executes `meltano run <job>` with dynamic env overrides.

**Tasks**:

1. **Create** `apps/etl/orchestration/meltano_runner.py`:
   - `run_meltano_job(job_name: str, env_overrides: dict) -> (exit_code, stdout, stderr)`
   - Use `asyncio.create_subprocess_exec` for non-blocking execution
   - Inject `env_overrides` into subprocess env (Meltano convention: `MELTANO_EXTRACTOR_<TAP>_CONFIG_<KEY>`)
   - Parse exit code; on non-zero, return structured error (e.g. replication slot missing)

2. **Create** `apps/etl/orchestration/connection_mapper.py`:
   - `connection_config_to_meltano_env(source_type: str, connection_config: dict, plugin_name: str) -> dict`
   - Maps generic keys (`db_host`, `host`, `port`, etc.) to `MELTANO_EXTRACTOR_*_CONFIG_*`
   - Reuse logic from `build_singer_config` where applicable
   - Support Postgres, MySQL, MongoDB; extendable for future connectors

3. **Keep** existing `main.py` endpoints working; do not change behavior yet.

**Deliverables**: New modules only; no removals.

---

### Phase 2: Discovery via Meltano

**Goal**: Replace `_run_discovery_sync` with `meltano invoke <tap> --discover`.

**Tasks**:

1. **Extend** `meltano.yml`:
   - Ensure all taps (tap-postgres, tap-mysql, tap-mongodb) are defined
   - Use custom/local taps from `connectors/` if Meltano Labs variants differ

2. **Add** `orchestration/discovery.py`:
   - `run_discovery(source_type: str, connection_config: dict) -> catalog_dict`
   - Resolves tap name from `source_type` (e.g. postgresql → tap-postgres)
   - Builds env via `connection_mapper`
   - Runs `meltano invoke <tap> --discover` with env overrides
   - Parses JSON catalog (reuse `parse_discovery_output` from utils)

3. **Refactor** `main.py` `/discover-schema/{source_type}`:
   - Call new `run_discovery()` instead of `_run_discovery_sync`
   - Keep request/response schema identical
   - Retain `extract_schema`, `catalog_to_schemas` logic

4. **Deprecate** (comment, do not delete) `_tap_command`, `_run_discovery_sync`; keep for fallback during testing.

**Deliverables**: Discovery path uses Meltano; old path kept as backup.

**Phase 2 DONE** (implemented):
- `meltano.yml`: Added tap-mysql extractor
- `orchestration/meltano_runner.py`: Added `run_meltano_invoke()` for `meltano invoke <tap> --discover`
- `orchestration/discovery.py`: Added `run_discovery()` — builds env, runs invoke, parses catalog
- `main.py`: `/discover-schema` uses `run_discovery()` first; falls back to `_run_discovery_sync` on failure
- Set `USE_MELTANO_DISCOVERY=false` to force legacy path

---

### Phase 3: Pipeline Execution via Meltano

**Goal**: Run full `tap → target → dbt` via `meltano run <job>`.

**Tasks**:

1. **Extend** `meltano.yml`:
   - Add **jobs** for each direction: `postgres-to-postgres`, `postgres-to-mongodb`, `mongodb-to-postgres`, `mysql-to-postgres`, etc.
   - Add **dbt** as transformer plugin; wire into job tasks
   - Ensure targets (target-postgres, target-mongodb if available) are configured

2. **Create** job definitions that accept env overrides:
   - Source and destination configs injected at runtime
   - No hardcoded `POSTGRES_URL`, `MONGODB_URI` in job execution

3. **Implement** `orchestration/pipeline_runner.py`:
   - `run_pipeline_job(job_name, source_config, dest_config, checkpoint?, state_id?)`
   - Build env from `connection_mapper` for both extractor and loader
   - Handle **state**: pass `--state-id` or use Meltano state backend if supported
   - For CDC: ensure state (LSN, binlog position, oplog ts) is persisted by Meltano
   - Run `meltano run <tap> <target> [dbt]` with env
   - Parse output for row counts, errors; return structured response

4. **Refactor** `main.py` `/run-meltano-pipeline`:
   - Map `RunMeltanoPipelineRequest` → job name + env
   - Call `run_pipeline_job()` instead of manual collect + emit
   - Keep `RunMeltanoPipelineResponse` schema; populate from Meltano output
   - If `transform_script` is provided: either (a) keep current in-memory transform step, or (b) add dbt model that wraps it ( Phase 4 )

5. **Error handling**:
   - Detect "replication slot missing" / "wal_level" / "binlog" errors from stderr
   - Return user-friendly message: "Create replication slot: `SELECT pg_create_logical_replication_slot(...)`"
   - Reference `logical_replication.py` `locate_replication_slot` for slot naming

**Deliverables**: Pipeline runs via Meltano; `/run-meltano-pipeline` delegates to runner.

**Phase 3 DONE** (implemented):
- `meltano.yml`: Added jobs postgres-to-postgres, mysql-to-postgres
- `orchestration/meltano_runner.py`: Added `run_args` for `--state-id` support
- `orchestration/pipeline_runner.py`: `run_pipeline_job()`, `get_job_for_direction()`, `PipelineRunResult`
- `main.py`: `/run-meltano-pipeline` uses Meltano when job exists, no source_table, no transform_script; else legacy
- postgres-to-mongodb: always legacy (no target-mongodb)
- Set `USE_MELTANO_PIPELINE=false` to force legacy path

---

### Phase 4: dbt and State Management

**Goal**: Add dbt transformer and proper CDC state handling.

**Tasks**:

1. **Add dbt to meltano.yml**:
   - `transformers: - name: dbt-postgres` (or appropriate variant)
   - Configure project path, profiles
   - Add dbt task to each job after target

2. **State management**:
   - Use Meltano `meltano state` / job state if available
   - Ensure `--state-id` or equivalent is passed per pipeline run
   - Persist state in `meltano.db` or external store; NestJS can pass `job_id` / `pipeline_id` for correlation

3. **Optional**: Add Great Expectations or other Meltano utility for validation

**Deliverables**: dbt runs after load; CDC state resumes correctly.

**Phase 4 DONE** (implemented):
- `meltano.yml`: Added dbt-postgres utility; jobs now run tap → target → dbt-postgres
- `transform/`: dbt project (dbt_project.yml, models/example.sql, profiles/postgres/profiles.yml)
- `connection_mapper`: `_connection_config_to_dbt_postgres_env()` — adds DBT_POSTGRES_* when dest is postgres
- `RunMeltanoPipelineRequest`: Added optional `state_id` for CDC resume (NestJS can pass pipeline_id)
- Run `meltano install` to install dbt-postgres

---

### Phase 5: Generic Collect/Emit API (Optional)

**Goal**: Allow `/collect` and `/emit` to optionally use Meltano for consistency.

**Tasks**:

1. For `/collect`: Optionally invoke `meltano run <tap>` with catalog selection; stream RECORD/STATE
2. For `/emit`: When using full Meltano flow, emit is handled by target; standalone `/emit` can remain as-is for backward compatibility (manual SQL/Mongo writes)

**Decision**: Keep standalone `/collect` and `/emit` for flexibility (e.g. preview, one-off). Primary path is `/run-meltano-pipeline` via Meltano.

**Phase 5 DONE** (implemented):
- `orchestration/collect.py`: `run_collect_via_meltano()` — runs tap via meltano invoke with config, catalog, state
- `main.py` `/collect`: When `USE_MELTANO_COLLECT=true`, tries Meltano path first; falls back to legacy on failure
- `/emit`: Unchanged (standalone manual writes for flexibility)

---

### Phase 6: Docker and Metadata-Driven UI

**Goal**: Immutable runtime and future-proof connector registration.

**Tasks**:

1. **Dockerfile**:
   - Multi-stage build
   - Install Meltano CLI, run `meltano install` to bake plugins
   - Start FastAPI server
   - Isolated Python versions per plugin (Meltano default)

2. **Metadata-driven frontend** (NestJS + DB):
   - New table or config: `connector_metadata` (source_type, required_fields, optional_fields, ui_schema)
   - NestJS endpoint: `GET /connectors/metadata`
   - Frontend builds connection form from metadata; adding MySQL vs MongoDB = new DB row + meltano.yml plugin

**Deliverables**: Docker image with Meltano + plugins; metadata API for dynamic UI.

**Phase 6 DONE** (implemented):
- `apps/etl/Dockerfile`: Multi-stage build; Meltano + plugins baked in; FastAPI on port 8001
- `apps/etl/.dockerignore`: Excludes .venv, .meltano, tests
- `apps/api/src/modules/data-sources/connector-metadata.ts`: Static metadata (postgresql, mysql, mongodb) + ssl/tls
- `apps/api/src/modules/data-sources/connectors.controller.ts`: GET /connectors/metadata
- Build: `docker build -f apps/etl/Dockerfile -t mantrixflow-etl .` (from repo root)
- **Frontend wiring**: `ConnectorsService`, `useConnectorMetadata`, `buildConnectionSchemasFromMetadata`; ConnectionSheet uses `connectionSchemasOverride` when metadata is available; data-sources page fetches metadata and passes to ConnectionSheet

---

## 4. Connection Config → Meltano Env Mapping

### Generic Mapping Function

```python
# Pseudocode for connection_mapper.py
CONNECTOR_ENV_MAP = {
    "postgresql": {
        "tap": "tap-postgres",
        "config_map": {
            "host": "host",
            "port": "port",
            "user": "user",
            "password": "password",
            "dbname": "database",  # or dbname
            "connection_string": "sqlalchemy_url",
        },
    },
    "mysql": {
        "tap": "tap-mysql",
        "config_map": {...},
    },
    "mongodb": {
        "tap": "tap-mongodb",
        "config_map": {
            "connection_string_mongo": "mongodb_connection_string",
            "database": "database",
            ...
        },
    },
}

def connection_config_to_meltano_env(source_type: str, connection_config: dict, role: "extractor"|"loader") -> dict:
    cfg = CONNECTOR_ENV_MAP.get(source_type)
    env = {}
    for api_key, meltano_key in cfg["config_map"].items():
        val = connection_config.get(api_key)
        if val is not None:
            env_var = f"MELTANO_EXTRACTOR_{cfg['tap'].upper().replace('-','_')}_CONFIG_{meltano_key.upper()}"
            env[env_var] = str(val)
    return env
```

Exact env var names follow [Meltano env var convention](https://docs.meltano.com/concepts/configuration#environment-variables).

---

## 5. What NOT to Remove

| Item | Reason |
|------|--------|
| `connectors/tap-postgres`, `tap-mysql`, `tap-mongodb` | Meltano uses these; keep as local/custom plugins |
| `logical_replication.py`, `binlog.py`, `oplog.py` | CDC logic inside taps; required for Log-Based CDC |
| `transformer.py` | User transforms; may run pre/post Meltano or as dbt |
| `utils.py` (parse, merge_state, chunked, etc.) | Shared helpers; still used |
| `_emit_to_sql`, `_emit_to_mongodb` | Keep until Meltano pipeline path is stable; useful for standalone `/emit` |
| NestJS pipeline definitions, PythonETLService | Contract preserved |
| `run-meltano-pipeline` request/response schema | Backward compatibility |

---

## 6. Code Removal Strategy (Minimal)

- **Remove only after** new path is tested and deployed
- **Replace, don’t delete**: e.g. `_run_discovery_sync` → call `run_discovery()` but keep old function commented for one release
- **Unused code**: After Phase 3 is stable, remove dead branches (e.g. old `_tap_command` if fully replaced)
- **No removals** in Phase 1–2

---

## 7. File Structure (Proposed)

```
apps/etl/
├── main.py                 # Slimmed; delegates to orchestration
├── orchestration/
│   ├── __init__.py
│   ├── meltano_runner.py    # run_meltano_job, asyncio subprocess
│   ├── connection_mapper.py # connection_config → MELTANO_* env
│   ├── discovery.py         # meltano invoke <tap> --discover
│   └── pipeline_runner.py  # meltano run <job>
├── connectors/              # KEEP as-is (used by Meltano)
├── utils.py                 # KEEP; maybe extend
├── transformer.py           # KEEP
├── meltano.yml              # Central config; add jobs, dbt
└── api/
    └── index.py             # Vercel entry; unchanged
```

---

## 8. Reference Checklist

- [ ] `meltano.yml` — plugin names, job definitions
- [ ] `logical_replication.py` — replication slot naming, LSN handling
- [ ] `utils.build_singer_config` — field mapping for Postgres, MySQL, MongoDB
- [ ] Meltano env docs: `MELTANO_EXTRACTOR_*_CONFIG_*`
- [ ] dbt plugin setup in Meltano

---

## 9. Summary

| Phase | Focus | Removals |
|-------|-------|----------|
| 1 | Meltano runner + connection mapper | None |
| 2 | Discovery via Meltano | Deprecate `_run_discovery_sync` (keep as fallback) |
| 3 | Pipeline execution via Meltano | Deprecate manual collect+emit in run-meltano-pipeline |
| 4 | dbt + state | None |
| 5 | Optional collect/emit via Meltano | None |
| 6 | Docker + metadata API | None |

**Principle**: Add orchestration, refactor call sites, deprecate incrementally, remove only when redundant and verified.
