# MANTrixFlow — Clean Engine Migration Plan

> **Goal**: Transition to a **pure Meltano + Singer SDK** architecture with zero legacy fallback. FastAPI becomes a strict orchestrator; no direct DB access, no manual tap imports, no SQLAlchemy/PyMongo in the ETL service.

---

## 1. Architecture Principle: Pure SDK Model

```
┌─────────────────────────────────────────────────────────────────────────┐
│  FastAPI (Orchestrator Only)                                             │
│  • Maps connection_config → MELTANO_EXTRACTOR_* / MELTANO_LOADER_* env   │
│  • Calls meltano invoke / meltano run                                    │
│  • Returns structured success/error — NO fallback to legacy code         │
└────────────────────────────────────┬────────────────────────────────────┘
                                     │ subprocess
┌────────────────────────────────────▼────────────────────────────────────┐
│  Meltano (Execution Layer)                                               │
│  • tap-postgres | tap-mysql | tap-mongodb (extract)                      │
│  • target-postgres | target-mysql | target-mongodb (load)                │
│  • dbt-postgres (transform)                                              │
│  • State in meltano.db; CDC via Singer SDK                               │
└─────────────────────────────────────────────────────────────────────────┘
```

### Non-Negotiables

| Rule | Current Violation | Target |
|------|-------------------|--------|
| **Zero Direct DB Access** | `sqlalchemy`, `pymongo`, `psycopg2`, `pymysql` in main.py for emit/test | Remove; Meltano loaders handle all writes |
| **Zero Manual Tap Imports** | `_tap_command()`, `tap_postgres.main()`, `tap_mysql`, `tap_mongodb` | Remove; use `meltano invoke <tap>` only |
| **No Legacy Fallback** | `USE_MELTANO_*` toggles, `_run_*_legacy`, fallback on Meltano failure | Remove; fail with structured error |
| **Mandatory Isolation** | Shared venv with tap libs | Meltano-managed venvs per plugin |

---

## 2. Refactoring Steps for `main.py`

### 2.1 Remove Legacy Toggles and Fallback Logic

**Delete**:
- `USE_MELTANO_DISCOVERY` env check and legacy branch
- `USE_MELTANO_COLLECT` env check and legacy branch
- `USE_MELTANO_PIPELINE` env check and `use_legacy` logic
- All `except ... fallback to legacy` blocks

**Replace with**: Single path per endpoint; on failure, return HTTP 502 with structured error.

---

### 2.2 Refactor `/discover-schema/{source_type}`

**Current flow**:
1. If `USE_MELTANO_DISCOVERY=true` → try `run_discovery()`
2. On failure → fall back to `_run_discovery_sync()` (direct tap)

**Target flow**:
1. Call `run_discovery(source_type, connection_config)` only
2. On failure → `HTTPException(502, detail=result.user_message or result.stderr)`
3. Remove `_run_discovery_sync`, `_run_discovery_source_type`, `_tap_command`, `_tap_env`, `_tap_pythonpath`, `_catalog_arg_name`

**Dependencies to keep**: `orchestration.discovery.run_discovery`, `utils.extract_schema`, `utils.catalog_to_schemas`, `utils.build_singer_config` (only for source_config merging if needed — or move to connection_mapper)

---

### 2.3 Refactor `/collect/{source_type}`

**Current flow**:
1. If `USE_MELTANO_COLLECT=true` → try `run_collect_via_meltano()`
2. On failure → fall back to `_run_discovery_sync` + `_run_collect_sync`

**Target flow**:
1. Call `run_collect_via_meltano()` only
2. On failure → `HTTPException(502, detail=...)`
3. Remove `_run_collect_sync`, `_run_discovery_sync` (if not already removed)

**Note**: `/collect` is a "sample data" endpoint. If Meltano doesn't support single-stream invoke with catalog override, consider:
- **Option A**: Require full pipeline run for data movement; deprecate `/collect` for raw extraction
- **Option B**: Use `meltano invoke tap-X --config ... --catalog ...` (if supported) — verify in `orchestration/collect.py`

---

### 2.4 Refactor `/run-meltano-pipeline`

**Current flow**:
1. If `use_legacy` (no job, or source_table, or transform_script) → `_run_meltano_pipeline_legacy`
2. Else → `run_pipeline_job()`
3. On Meltano failure → fall back to legacy

**Target flow**:
1. Resolve `job_name = get_job_for_direction(source_type, dest_type)`
2. If `job_name is None` → `HTTPException(400, detail="No Meltano job for direction X→Y. Add job in meltano.yml.")`
3. If `payload.source_table` or `payload.transform_script` → `HTTPException(400, detail="Single-table sync and transform_script not supported in Clean Engine. Use full pipeline.")` — OR extend Meltano jobs to support catalog override
4. Call `run_pipeline_job()` only
5. On failure → `HTTPException(502, detail=result.user_message)`
6. Remove `_run_meltano_pipeline_legacy` entirely

**Scope decision**: If postgres-to-mongodb has no Meltano job (target-mongodb Python 3.12 issue), either:
- Add a Meltano-compatible target-mongodb (e.g. different variant)
- Or return 400 for that direction until supported

---

### 2.5 Refactor `/emit/{dest_type}` and `/test-connection`

**Current**: Direct SQLAlchemy/PyMongo for emit and test.

**Target**:
- **Emit**: Deprecate or remove. Data movement is exclusively via Meltano loaders. If needed for "ad-hoc emit" (e.g. from transform output), consider a Meltano job that reads from stdin or a temp file — complex; may defer.
- **Test connection**: Replace with `meltano invoke <tap> --discover` with a minimal config (or a dedicated test command). If discovery succeeds, connection works. No direct `create_engine` or `MongoClient`.

**Simpler approach for test-connection**: Run `run_discovery(source_type, connection_config)` with a 10s timeout. Success = connection OK.

---

### 2.6 Remove Direct DB / Tap Code from `main.py`

**Delete**:
- `_tap_command`, `_tap_env`, `_tap_pythonpath`, `_catalog_arg_name`
- `_run_discovery_sync`, `_run_discovery_source_type`, `_run_collect_sync`
- `_run_meltano_pipeline_legacy`
- `_mongo_client_kwargs`, `_build_sqlalchemy_url` (if only used for test/emit)
- `_emit_to_sql`, `_emit_to_mongodb` and all SQLAlchemy/PyMongo emit logic
- `_prepare_sql_table`, `_sanitize_record_for_sql`, `_objectid_to_uuid`, `_mongo_id_to_uuid`, `_is_uuid_column`, `_infer_sql_type`
- Imports: `sqlalchemy`, `pymongo`, `certifi` (if only for MongoClient)
- `CONNECTORS_DIR` if only used for `_tap_pythonpath`

**Keep**:
- Pydantic models, FastAPI app, CORS, JWT
- `orchestration.discovery.run_discovery`
- `orchestration.pipeline_runner.run_pipeline_job`, `get_job_for_direction`
- `orchestration.collect.run_collect_via_meltano` (if /collect stays)
- `transformer.safe_exec_transform`, `validate_transform_script` (only if transform API remains for non-pipeline use)
- `utils`: `extract_schema`, `catalog_to_schemas`, `parse_discovery_output`, `merge_state`, `select_catalog_streams`, `chunked`, `temporary_json_file`, `run_command` (if still needed), `ETLRuntimeError`
- `build_singer_config` — only if discovery/collect need it for source_config; else move to connection_mapper

---

## 3. Unified Connection Mapping

### 3.1 Standard Meltano Env Format

`connection_mapper.py` must output:

```
MELTANO_EXTRACTOR_<PLUGIN>_<KEY> = <value>
MELTANO_LOADER_<PLUGIN>_<KEY> = <value>
```

Where `<PLUGIN>` is the Meltano plugin name in SCREAMING_SNAKE (e.g. `TAP_POSTGRES`).

### 3.2 Extend for New Connectors

- Add entry to `SOURCE_TYPE_TO_TAP` / `SOURCE_TYPE_TO_LOADER`
- Add mapping in `_connection_config_to_extractor_config` / `_connection_config_to_loader_config`
- Add plugin block in `meltano.yml`
- No changes to `main.py` endpoint logic

### 3.3 Remove `build_singer_config` Dependency

- `connection_mapper` should be the single source for config → env mapping
- `utils.build_singer_config` was for legacy tap config files; replace usages with `connection_config_to_meltano_env` + any source_config overlay (table, schema) as extra env vars

---

## 4. Pure SDK CDC & State Management

### 4.1 State Handling

- Always pass `--state-id` to `run_meltano_job` when resuming (e.g. `pipeline_run_id`)
- Meltano stores state in `meltano.db`; Singer taps emit STATE messages
- No custom state merge in main.py for pipeline path

### 4.2 CDC (Log-Based Replication)

- PostgreSQL WAL, MySQL Binlog, MongoDB Oplog — handled by taps internally
- No changes needed if taps are Singer SDK-based and support incremental

---

## 5. Requirements and Dependencies

### 5.1 `requirements.txt` — Slim Down

**Remove**:
- `psycopg2-binary`
- `pymysql`
- `pymongo`
- `sqlalchemy`
- `singer-python` (Meltano installs taps in their own venvs)
- `mysql-replication`, `strict-rfc3339`, `tzlocal`, `attrs` (tap deps; Meltano manages)
- `certifi` (if only for MongoClient; Meltano/taps may need it — verify)

**Keep**:
- `fastapi`, `uvicorn`, `pydantic`, `python-dotenv`, `python-multipart`
- `httpx` (if used for outbound calls)
- `meltano` (if running Meltano as library) — or rely on `meltano` CLI in PATH
- `pytest` (testing)

**Add** (if not present):
- `meltano` — for programmatic use (optional; can use subprocess to `meltano` CLI only)

### 5.2 Connectors Directory

- `connectors/tap-postgres`, `tap-mysql`, `tap-mongodb` — **keep** if Meltano project uses `meltano add` with local path
- Or remove and use Meltano Labs variants only (from meltano.yml)

---

## 6. Meltano Configuration

### 6.1 `meltano.yml`

- Ensure all extractors and loaders use `variant: meltanolabs` or SDK-based variants
- Add `target-mongodb` when a Python 3.12–compatible variant exists (or document postgres-to-mongodb as unsupported until then)
- Jobs: `postgres-to-postgres`, `mysql-to-postgres`, `mongodb-to-postgres` (and `postgres-to-mongodb` when target available)
- dbt: `dbt-postgres` as utility in jobs

### 6.2 Environment

- No `.env` connection strings for dynamic runs; all from API `connection_config` → `connection_mapper` → env overrides

---

## 7. Implementation Phases

### Phase A: Discovery & Pipeline (No Collect/Emit) ✅ DONE

1. ✅ Remove legacy discovery fallback; `/discover-schema` → `run_discovery` only
2. ✅ Remove legacy pipeline fallback; `/run-meltano-pipeline` → `run_pipeline_job` only, fail if no job
3. ✅ Remove `_run_meltano_pipeline_legacy`; reject 400 when no job, source_table, or transform_script
4. ✅ Remove `USE_MELTANO_DISCOVERY`, `USE_MELTANO_PIPELINE` env checks
5. ✅ **Phase C**: Removed `_tap_command`, `_run_discovery_sync`, `_run_collect_sync`; `/collect` now Meltano-only

### Phase B: Test Connection ✅ DONE

1. ✅ Replace direct DB test with `run_discovery(..., timeout_seconds=10)`
2. **Deferred to Phase D**: `create_engine`, `MongoClient` still used by `/emit`; remove when emit is deprecated

### Phase C: Collect ✅ DONE

1. ✅ Make `/collect` use `run_collect_via_meltano` only; remove legacy fallback
2. ✅ Removed `_tap_command`, `_tap_env`, `_tap_pythonpath`, `_catalog_arg_name`, `_run_discovery_sync`, `_run_collect_sync`
3. ✅ Removed unused imports: `asyncio`, `sys`, `ETLRuntimeError`, `parse_discovery_output`, `parse_singer_stream`, `run_command`, `temporary_json_file`

### Phase D: Emit & Dependencies ✅ DONE

1. ✅ Deprecate `/emit` — returns 410 Gone with message to use `/run-meltano-pipeline`
2. ✅ Remove `_emit_to_sql`, `_emit_to_mongodb`, SQLAlchemy, PyMongo, psycopg2, pymysql from main.py
3. ✅ Update `requirements.txt` — removed sqlalchemy, psycopg2-binary, pymysql, pymongo, certifi, singer-python, tap-specific libs
4. ✅ Refactor `utils._parse_mongo_uri` to use stdlib only (no pymongo)

### Phase E: Utils Cleanup ✅ DONE

1. ✅ Moved `build_singer_config` logic to `connection_config_to_tap_config` in connection_mapper
2. ✅ Moved `_parse_mongo_uri` to connection_mapper (used by `_build_mongodb_connection_config`)
3. ✅ Pruned unused helpers from utils: `run_command`, `ETLRuntimeError`, `chunked`, `build_singer_config`, `_parse_mongo_uri`

### Phase F: Transform Endpoint ✅ DONE

1. ✅ Deprecate `/transform` — returns 410 Gone; message: "Use POST /run-meltano-pipeline with dbt for transformations"
2. ✅ Removed transformer import from main.py; endpoint stub retained for backward compatibility
3. **Callers to update**: `apps/api` pipeline.service.ts, `apps/app` python-etl.service.ts — migrate to run-meltano-pipeline with dbt

---

## 8. Final Cleanup Checklist

- [ ] No `sqlalchemy`, `pymongo`, `psycopg2`, `pymysql` in main.py
- [ ] No `_tap_command`, `_run_discovery_sync`, `_run_collect_sync`, `_run_meltano_pipeline_legacy`
- [ ] No `USE_MELTANO_*` or `use_legacy` variables
- [ ] `/discover-schema` → `run_discovery` only; 502 on failure
- [ ] `/run-meltano-pipeline` → `run_pipeline_job` only; 400 if no job; 502 on failure
- [ ] `/test-connection` → `run_discovery` with short timeout
- [ ] `/collect` → `run_collect_via_meltano` only or deprecated
- [ ] `/emit` → deprecated or removed
- [ ] `requirements.txt` slimmed (no DB drivers, no singer-python)
- [ ] `connection_mapper` is the single config→env bridge
- [ ] `meltano.yml` has all jobs for supported directions
- [ ] dbt used only as Meltano transformer utility

---

## 9. Risk Mitigation

| Risk | Mitigation |
|------|-------------|
| Meltano invoke fails (e.g. venv not installed) | Document `meltano install` as required before first run; health check can run `meltano --version` |
| postgres-to-mongodb no job | Return 400 with clear message; add target-mongodb when available |
| Single-table sync / transform_script in pipeline | Either extend Meltano to support catalog override, or require full sync + separate transform API |
| Test connection timeout | Use 10s; discovery is lightweight |

---

## 10. Success Criteria

- FastAPI ETL service has **zero** direct database connections
- All data movement goes through Meltano `run` or `invoke`
- New connector = add to meltano.yml + connection_mapper; no main.py changes
- `requirements.txt` has no DB drivers or tap-specific libs
- Clear error responses when Meltano fails (no silent fallback)
