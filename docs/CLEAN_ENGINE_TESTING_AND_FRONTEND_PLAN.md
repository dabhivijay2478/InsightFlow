# MANTrixFlow Clean Engine — Testing & Frontend Refinement Plan

> **Context**: The migration to a pure **Clean Engine** architecture (Meltano + Singer SDK) is fundamentally complete on `fix-missing-data`. Legacy database connection code is deprecated; orchestration is enforced via the Meltano CLI. This plan covers backend verification, frontend refinements, and deployment strategy.

---

## Phase Status

| Phase | Status | Notes |
|-------|--------|-------|
| **Phase 1: Backend Verification** | ✅ Done | state_id, slim requirements, verification script |
| **Phase 2: Frontend Refactoring** | ✅ Done | Connection logic, Transform UI, Lifecycle tracking |
| **Phase 3: Deployment** | ✅ Done | Dockerfile slimmed, Meltano-only |

**All phases complete.** Optional: run `./scripts/verify-clean-engine-phase1.sh` and manual discovery/CDC tests.

### Frontend Implementation Plan (CLEAN_ENGINE_FRONTEND_IMPLEMENTATION_PLAN.md)

| FE Phase | Status |
|----------|--------|
| 1. Remove Legacy UI | ✅ Done |
| 2. Metadata-Driven Connection Forms | ✅ Done |
| 3. dbt Model Selector | ✅ Done |
| 4. Pipeline Run Tracking & CDC Feedback | ✅ Done |
| 5. Full Sync vs CDC Toggle | ✅ Done |
| 6. Cleanup & Polish | ✅ Done |

---

## 1. Backend Verification & Testing Steps ✅ DONE

### 1.1 Plugin Installation Check

**Goal**: Ensure Meltano plugins are installed and Python 3.12 compatibility fixes are active.

| Step | Action | Verification |
|------|--------|---------------|
| 1 | Run `meltano install` inside Docker or local env | `.meltano/` contains isolated venvs for tap-postgres, tap-mysql, tap-mongodb, target-postgres, dbt-postgres |
| 2 | Verify `meltano config` succeeds | No import or plugin errors |
| 3 | Check `meltano.yml` jobs | `postgres-to-postgres`, `mysql-to-postgres`, `mongodb-to-postgres` jobs exist |

**Files**: `apps/etl/meltano.yml`, `apps/etl/Dockerfile` (Stage 1: `meltano install`)

---

### 1.2 Discovery Validation

**Goal**: Confirm `run_discovery` invokes Meltano and parses the catalog correctly.

| Step | Action | Verification |
|------|--------|---------------|
| 1 | Call `POST /discover-schema/postgresql` with valid connection config | Returns `columns`, `primary_keys`, `schemas` |
| 2 | Repeat for `POST /discover-schema/mysql` and `POST /discover-schema/mongodb` | Same structure |
| 3 | Inspect `orchestration/discovery.py` | Uses `meltano invoke tap-postgres --discover` (or tap-mysql, tap-mongodb) |
| 4 | Verify `utils.extract_schema` / `catalog_to_schemas` parse JSON catalog | No raw Singer catalog leaked to API |

**Files**: `apps/etl/orchestration/discovery.py`, `apps/etl/main.py` (discover-schema routes)

---

### 1.3 Job Execution & State Persistence

**Goal**: Ensure `state_id` is passed to Meltano and CDC state (LSN/Binlog) is persisted.

| Step | Action | Verification |
|------|--------|---------------|
| 1 | Run `POST /run-meltano-pipeline` with `state_id: "pipeline_<id>"` | Request accepted |
| 2 | Check backend logs | `pipeline_runner.py` passes `--state-id` to `meltano run` |
| 3 | Inspect `meltano.db` or configured state backend | LSN (Postgres) or Binlog position (MySQL) saved after run |
| 4 | Run incremental sync with same `state_id` | Only new/changed rows processed |

**Files**: `apps/etl/orchestration/pipeline_runner.py`, `apps/etl/main.py`, `apps/api/.../python-etl.service.ts` (passes `state_id`)

---

### 1.4 CDC Error Handling

**Goal**: Users receive actionable guidance when replication is not configured.

| Step | Action | Verification |
|------|--------|---------------|
| 1 | Use a DB account without replication permissions (e.g. no `REPLICATION` for Postgres) | Pipeline fails with 502 |
| 2 | Inspect error response | `detail` contains `_infer_user_message` output (e.g. `SELECT pg_create_logical_replication_slot(...)`) |
| 3 | Verify frontend displays this message | `PipelineRunTracker` / run detail shows `errorMessage` with user-friendly text |

**Files**: `apps/etl/orchestration/meltano_runner.py` (`_infer_user_message`), `apps/etl/main.py` (502 detail), `apps/app/components/data-pipelines/pipeline-run-tracker.tsx`

---

## 2. Frontend Refactoring for the New Engine ✅ DONE

### 2.A. Centralize Connection Logic

**Current state**:
- `connector-metadata.ts` (API) defines `CONNECTOR_METADATA` with `uiSchema` for postgresql, mysql, mongodb
- `ConnectionSheet` uses `connectionSchemas` from `constants.ts` (hardcoded) or `connectionSchemasOverride` (from API metadata)
- `useConnectorMetadata` fetches metadata from API; `ConnectionSheet` can use it via `connectionSchemasOverride`

| Task | Owner | Priority | Details |
|------|-------|----------|---------|
| **Remove legacy toggles** | Frontend | High | Remove any UI that switches between "Legacy" and "Meltano" modes. Single path only. |
| **Metadata-driven forms** | Frontend | High | Ensure `ConnectionSheet` always uses `useConnectorMetadata` when available. Map `CONNECTOR_METADATA.uiSchema` → `ConnectionSchema` format. Add new connector = new entry in `connector-metadata.ts` + `meltano.yml`; zero frontend form changes. |
| **Deprecate hardcoded constants** | Frontend | Medium | Gradually replace `connectionSchemas` in `constants.ts` with API-driven schema. Keep as fallback only if API unavailable. |

**Files**: `apps/app/components/data-sources/connection-sheet.tsx`, `apps/app/components/data-sources/connector-metadata-utils.ts`, `apps/api/src/modules/data-sources/connector-metadata.ts`

---

### 2.B. Update Transformation UI

**Current state**:
- `transform-step.tsx` uses a Python code editor for `transformScript`
- `/run-meltano-pipeline` rejects `transform_script` (Clean Engine uses dbt only)
- `destination-schema.service.ts` no longer requires `transformScript`

| Task | Owner | Priority | Details |
|------|-------|----------|---------|
| **Replace Python transform with dbt** | Frontend | Medium | Remove or hide the Python `transformScript` code editor. Add UI for selecting/defining dbt models (e.g. dropdown of models in `transform/models/`, or link to dbt project config). |
| **Update pipeline wizard** | Frontend | Medium | `transform-step.tsx`: Remove `transformScript` requirement. Show "Transformations are handled by dbt in the Meltano job" or dbt model selector. |
| **Destination schema UI** | Frontend | Low | `destination-schemas/page.tsx`: Make `transformScript` column optional or remove. Show "dbt" or "Meltano" instead of script status. |

**Files**: `apps/app/app/workspace/data-pipelines/new/transform-step.tsx`, `apps/app/app/workspace/data-pipelines/new/page.tsx`, `apps/app/app/workspace/destination-schemas/page.tsx`

---

### 2.C. Lifecycle & Status Tracking

**Current state**:
- ETL returns `user_message` in HTTP 502 `detail` when pipeline fails
- API `python-etl.service` throws; `extractPythonError` preserves `response.data.detail`
- Pipeline run stores `errorMessage` in DB; `PipelineRunTracker` displays `run.errorMessage`
- Socket.io `run_update` emits `error_message`; detail page shows it in toast

| Task | Owner | Priority | Details |
|------|-------|----------|---------|
| **Display `user_message` in errors** | Frontend | Medium | Ensure failed runs show the ETL's `user_message` (CDC guidance, timeout, etc.). Currently `run.errorMessage` is populated from API throw; verify the API passes `detail` through to `errorMessage` in run record. |
| **Real-time progress** | Frontend | Medium | Poll or subscribe to run status. Display `rows_read` and `rows_written` from `run_update` or run detail. Detail page already shows `rows_written` in toasts; ensure `PipelineRunTracker` shows progress during run. |
| **Structured error display** | Frontend | Low | Consider a dedicated "CDC Setup" or "Replication" error block when `user_message` contains replication-slot/binlog/oplog guidance. |

**Files**: `apps/app/components/data-pipelines/pipeline-run-tracker.tsx`, `apps/app/app/workspace/data-pipelines/[id]/page.tsx`, `apps/api/.../pipeline.service.ts`, `apps/api/.../python-etl.service.ts`

---

## 3. Deployment Strategy: Docker Integration ✅ DONE

**Current state**:
- `Dockerfile` has two stages: meltano-builder (install plugins) and runtime (FastAPI + Meltano)
- Copies `meltano.yml`, `transform/`, `.meltano/` from builder
- `requirements.txt` slimmed (no tap deps; Meltano manages in isolated venvs)
- Removed redundant `connectors/` copy and `pip install -e ./connectors/...` (Clean Engine uses Meltano only)

| Task | Owner | Priority | Status |
|------|-------|----------|--------|
| Slim `requirements.txt` | Backend | High | ✅ Done |
| Dockerfile validation | Backend | High | ✅ Done |
| Isolated environments | Backend | Medium | ✅ Meltano install creates plugin venvs |

**Files**: `apps/etl/requirements.txt`, `apps/etl/Dockerfile`

---

## 4. Summary: Next Steps

| Task | Owner | Priority | Status |
|------|-------|----------|--------|
| Slim `requirements.txt` | Backend | High | ✅ Done |
| Validate `state_id` persistence | Backend | High | ✅ Done |
| Plugin installation check | Backend | High | ✅ Done (script) |
| Discovery validation | Backend | High | ✅ Done (manual steps documented) |
| CDC error handling | Backend | Medium | ✅ Done (manual steps documented) |
| Remove legacy toggles | Frontend | High | ✅ N/A (no toggle UI) |
| Metadata-driven connection forms | Frontend | High | ✅ Done (data-sources uses useConnectorMetadata) |
| **Replace Python Transform UI with dbt** | Frontend | Medium | ✅ Done |
| Update UI to use `user_message` errors | Frontend | Medium | ✅ Done |
| Real-time progress display | Frontend | Medium | ✅ Done |

---

## 5. Optional: Metadata-Driven React Component Example

A metadata-driven `ConnectionForm` component would:

1. Fetch `CONNECTOR_METADATA` from `GET /connector-metadata` (or equivalent)
2. Map `uiSchema` to form fields (Input, Select, Password, etc.)
3. Render fields dynamically; no hardcoded field list per connector
4. Validate using `requiredFields` and optional constraints

**Proposed structure**:
```tsx
// ConnectionFormDynamic.tsx
// - useConnectorMetadata(sourceType) → ConnectorMetadata
// - metadata.uiSchema.forEach(field => renderField(field))
// - onSubmit: build connection_config from form values keyed by field.key
```

This would live alongside or replace the current `ConnectionSheet` form logic. The API's `connector-metadata.ts` already provides the schema; the frontend needs to consume it and render accordingly.

---

## 6. Phase 1 Verification Script

Run from repo root:

```bash
./scripts/verify-clean-engine-phase1.sh
```

This script:
- Runs `meltano install` and `meltano config` (if meltano in PATH)
- Verifies jobs in meltano.yml (postgres-to-postgres, mysql-to-postgres, mongodb-to-postgres)
- Documents manual steps for discovery, state persistence, and CDC error handling

---

## 7. References

- `docs/CLEAN_ENGINE_MIGRATION_PLAN.md` — Architecture and migration steps
- `docs/MANTRIXFLOW_MELTANO_REFACTOR_PLAN.md` — Meltano integration design
- `apps/etl/orchestration/meltano_runner.py` — `_infer_user_message` for CDC errors
- `apps/api/src/modules/data-sources/connector-metadata.ts` — Connector metadata registry
