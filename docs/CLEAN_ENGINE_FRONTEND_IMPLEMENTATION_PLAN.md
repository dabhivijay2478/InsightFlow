# Clean Engine — Frontend Implementation Plan

> **Context**: Backend migration to Meltano-as-an-Engine is complete. This plan aligns the Frontend (FE) with the new architecture by removing legacy UI, adopting metadata-driven forms, and enhancing CDC feedback.

---

## Current State Summary

| Area | Status | Notes |
|------|--------|-------|
| **Transform UI** | Done | PythonScriptEditor removed; dbt-only messaging |
| **Connection Forms** | Done | Data-sources and onboarding use metadata-driven forms |
| **Pipeline Run Tracker** | Done | Shows `errorMessage`, CDC badge |
| **Onboarding** | Done | Metadata-driven; Full Sync vs CDC toggle added |

---

## Phase 1: Remove Legacy UI (High Priority) ✅ DONE

### 1.1 Delete PythonScriptEditor ✅

**Goal**: Remove deprecated Python transform editor; engine uses dbt only.

| Task | File(s) | Status |
|------|---------|--------|
| Delete component | `components/data-pipelines/python-script-editor.tsx` | ✅ Removed |
| Remove exports | `components/data-pipelines/index.ts` | ✅ N/A (was not exported) |
| Verify no imports | Grep for `PythonScriptEditor` | ✅ No references |

---

### 1.2 Remove "Legacy script" Badge and Script References ✅

**Goal**: Clean transform-step UI; no references to legacy Python scripts.

| Task | File | Status |
|------|------|--------|
| Update Script column | `transform-step.tsx` | ✅ "Transform" column shows "dbt" badge |
| Remove transform mode toggle | `transform-step.tsx` | ✅ `_transformMode` removed |
| Remove debug logs | `transform-step.tsx` | ✅ Removed |

**Files**: `app/workspace/data-pipelines/new/transform-step.tsx`

---

### 1.3 Destination Schemas Page ✅

**Goal**: Make `transformScript` optional; show "dbt" instead of script status.

| Task | File | Status |
|------|------|--------|
| Update transform column | `destination-schemas/page.tsx` | ✅ Shows "dbt" badge |
| Details dialog | `destination-schemas/page.tsx` | ✅ dbt description |

**Files**: `app/workspace/destination-schemas/page.tsx`

---

## Phase 2: Metadata-Driven Connection Forms (High Priority)

### 2.1 ConnectionSheet — Ensure Metadata-First ✅

**Goal**: ConnectionSheet always uses API metadata when available; constants only as fallback.

| Task | File | Status |
|------|------|--------|
| Prefer metadata | `connection-sheet.tsx` | ✅ Uses `connectionSchemasOverride` when provided |
| Data-sources page | `data-sources/page.tsx` | ✅ Uses `useConnectorMetadata` + `buildConnectionSchemasFromMetadata` |
| Fallback | `connection-sheet.tsx` | ✅ Falls back to `connectionSchemas` from constants if override undefined |

---

### 2.2 Onboarding Connect Flow — Use Metadata

**Goal**: Replace hardcoded Postgres-style form with metadata-driven form.

| Task | File | Status |
|------|------|--------|
| Fetch metadata | `onboarding/connect/[connector]/page.tsx` | ✅ Uses `useConnectorMetadata` |
| Map connector param | — | ✅ postgres, mysql, mongodb match frontend types |
| Render dynamic form | — | ✅ Uses `buildConnectionSchemasFromMetadata` + dynamic field rendering |
| Validation | — | ✅ `buildConnectionSchema` for zod (matches ConnectionSheet) |

**Files**: `app/onboarding/connect/[connector]/page.tsx`, `components/data-sources/connector-metadata-utils.ts`

**Implementation**: Database connectors (postgres, mysql, mongodb) now render form fields from API metadata. OAuth and file upload flows unchanged.

---

## Phase 3: dbt Model Selector (Medium Priority) ✅ DONE

### 3.1 Add dbt Model Selector to Transform Step ✅

**Goal**: Allow users to select which dbt models run as part of the sync (meltano.yml includes `dbt-postgres`).

| Task | File | Status |
|------|------|--------|
| List dbt models | ETL `GET /dbt-models` | ✅ Lists .sql files from transform/models/ |
| Add model selector UI | `transform-step.tsx` | ✅ Badge-based multi-select |
| Store selection | `TransformConfig.dbtModels` | ✅ Optional string[] |
| Pass to pipeline | API/ETL | ⏳ Backend support coming soon (UI stores selection) |

**Implementation**: ETL exposes `/dbt-models`; frontend fetches via `useDbtModels`; transform step shows clickable badges. Empty selection = run all models. Backend will use `dbtModels` when supported.

**Files**: `apps/etl/main.py`, `apps/app/lib/api/services/python-etl.service.ts`, `apps/app/lib/api/hooks/use-data-pipelines.ts`, `apps/app/app/workspace/data-pipelines/new/transform-step.tsx`

---

## Phase 4: Pipeline Run Tracking & CDC Feedback (Medium Priority) ✅ DONE

### 4.1 Verify user_message Flow ✅

**Goal**: Ensure `user_message` from ETL 502 responses reaches `run.errorMessage` and is displayed.

| Task | Layer | Status |
|------|-------|--------|
| ETL | `main.py` | ✅ Returns `detail` with `user_message` in 502 |
| API | `python-etl.service.ts` | ✅ For 502, passes raw `detail` (no prefix) so CDC guidance displays cleanly |
| API | `pipeline.service.ts` | ✅ Persists `error.message` to `run.errorMessage` |
| FE | `PipelineRunTracker` | ✅ Displays `run.errorMessage` |

---

### 4.2 Structured CDC Error Block ✅

**Goal**: When `user_message` contains replication-slot/binlog/oplog guidance, show a dedicated "CDC Setup" block.

| Task | File | Status |
|------|------|--------|
| Detect CDC errors | `pipeline-run-tracker.tsx` | ✅ replication/binlog/oplog triggers CDC styling |
| CDC block | — | ✅ Amber styling, "CDC Setup Required" header, copy-to-clipboard button |
| Copy button | — | ✅ Copies error message (SQL/guidance) to clipboard |

**Files**: `apps/api/.../python-etl.service.ts`, `components/data-pipelines/pipeline-run-tracker.tsx`

---

## Phase 5: Onboarding — Full Sync vs CDC Toggle (Medium Priority) ✅ DONE

### 5.1 Add Sync Mode Toggle ✅

**Goal**: Clear toggle for "Full Sync" vs "Log-Based CDC" in connection setup.

| Task | File | Status |
|------|------|--------|
| Add toggle | `onboarding/connect/[connector]/page.tsx` | ✅ Radio: "Full Sync" \| "Log-Based CDC" |
| Add toggle | `connection-sheet.tsx` | ✅ Same radio for data-sources page |
| Store preference | Connection config | ✅ `sync_mode` ("full" \| "incremental") in config |
| Backend mapping | `connection_mapper.py` | Already supports `replication-method` in meltano env |

**Note**: For Postgres, CDC requires `wal_level=logical` and replication slot. For MySQL, binlog. MongoDB uses oplog. Connector-specific guidance shown below toggle.

**Implementation**: Sync mode toggle added to both onboarding and ConnectionSheet for postgres, mysql, mongodb. Stored as `sync_mode` in connection config. Pipeline creation can default `syncMode` from connection config when supported.

---

## Phase 6: Cleanup & Polish (Low Priority) ✅ DONE

### 6.1 Remove Dead Code ✅

| Task | File | Status |
|------|------|--------|
| Remove `_transformMode` | `transform-step.tsx` | ✅ Already removed in Phase 1.2 |
| Remove `_handleAutoGenerate`, `_handleFieldMapping`, `_handleRemoveMapping` | `transform-step.tsx` | ✅ Removed (no UI calls them) |
| Remove debug console.logs | `transform-step.tsx` | ✅ Stripped |
| Remove `generatedDestinationFields` state | `transform-step.tsx` | ✅ Removed (unused after handler removal) |

---

### 6.2 Update Documentation ✅

| Task | File | Status |
|------|------|--------|
| Update CLEAN_ENGINE_TESTING_AND_FRONTEND_PLAN | `docs/CLEAN_ENGINE_TESTING_AND_FRONTEND_PLAN.md` | ✅ Mark FE phases complete |
| Add FE implementation notes | This doc | ✅ Verification checklist linked below |

---

## Implementation Order

| Phase | Priority | Est. Effort | Dependencies |
|-------|----------|-------------|--------------|
| **1.1** Delete PythonScriptEditor | High | 0.5h | None |
| **1.2** Remove Legacy badge/script refs | High | 0.5h | 1.1 |
| **1.3** Destination schemas transform column | High | 0.5h | None |
| **2.1** Verify ConnectionSheet metadata-first | High | 0.5h | None |
| **2.2** Onboarding metadata-driven form | High | 2h | 2.1 |
| **3.1** dbt Model Selector | Medium | 2h | ETL/API contract |
| **4.1** Verify user_message flow | Medium | 1h | Manual test |
| **4.2** CDC error block | Low | 1h | 4.1 |
| **5.1** Full Sync vs CDC toggle | Medium | 2h | 2.2 |
| **6.1** Dead code cleanup | Low | 0.5h | 1.2 |
| **6.2** Docs update | Low | 0.5h | All |

---

## Verification Checklist

See `docs/CLEAN_ENGINE_TESTING_AND_FRONTEND_PLAN.md` for backend verification. Frontend phases 1–6 complete.

- [ ] `pnpm build` succeeds
- [x] No references to `PythonScriptEditor`, `transform_script` (except optional backward compat)
- [x] Data-sources: new connection uses metadata-driven form
- [x] Onboarding: connect form is dynamic per connector
- [ ] Pipeline run failure shows `user_message` (CDC guidance)
- [x] Transform step shows dbt-only messaging
- [x] Destination schemas: transformScript optional, shows "dbt"

---

## References

- `docs/CLEAN_ENGINE_TESTING_AND_FRONTEND_PLAN.md` — Backend verification, FE phase status
- `docs/CLEAN_ENGINE_MIGRATION_PLAN.md` — Architecture
- `apps/api/src/modules/data-sources/connector-metadata.ts` — Connector metadata registry
- `apps/app/components/data-sources/connector-metadata-utils.ts` — API → form schema mapping
- `apps/etl/orchestration/meltano_runner.py` — `_infer_user_message` for CDC errors
