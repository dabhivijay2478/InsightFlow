# Pure Meltano Engine Alignment Plan

> **Context**: The backend (FastAPI ETL) strictly enforces Meltano orchestration. Legacy paths for direct data movement and custom Python transformations return **410 Gone**. This plan removes all legacy code and standardizes on the Meltano/Singer SDK path.

---

## Current State vs. Target

| Area | Current | Target | Status |
|------|---------|--------|--------|
| NestJS `transform()` / `emit()` | ✅ Already absent | No action | ✅ Done |
| NestJS `runMeltanoPipeline` state_id | ✅ Passed as `pipeline_${id}` | Verify strict enforcement | ✅ Done |
| NestJS `transform_script` in payload | Removed | Remove from payload | ✅ Done |
| Frontend `transform()` / `emit()` | ✅ Already absent | No action | ✅ Done |
| Transform step `transformScript` | Removed | Remove entirely; dbt-only | ✅ Done |
| Transform step `_handleAutoGenerate` etc. | ✅ Removed (Phase 6) | No action | ✅ Done |
| Onboarding forms | ✅ Metadata-driven + Sync toggle | No action | ✅ Done |
| PipelineRunTracker `user_message` | ✅ Displays | No action | ✅ Done |

---

## Phase 1: NestJS API Cleanup

### 1.1 Remove `transform_script` from Payload ✅ Done

**File**: `apps/api/src/modules/data-pipelines/services/python-etl.service.ts`

| Task | Action | Status |
|------|--------|--------|
| Remove `transformScript` param | Delete from `runMeltanoPipeline` options | ✅ |
| Remove from request body | Stop sending `transform_script` to ETL | ✅ |
| Rationale | ETL returns 410 when `transform_script` is provided; Clean Engine uses dbt only | — |

**Current** (lines 258, 282, 307):
```ts
transformScript?: string;
// ...
transformScript,
// ...
transform_script: transformScript ?? null,
```

**Target**: Remove all three. Keep `state_id` (already passed correctly).

### 1.2 Enforce `state_id` (Verification) ✅ Done

**File**: `apps/api/src/modules/data-pipelines/services/pipeline.service.ts`

| Task | Action | Status |
|------|--------|--------|
| Verify | `stateId: \`pipeline_${pipeline.id}\`` is always passed (line 600) | ✅ Done |
| Optional | Add assertion that `stateId` is never null/undefined before calling | Skipped (not needed) |

**Status**: ✅ Verified. `stateId` is always passed.

---

## Phase 2: Frontend Client Cleanup

### 2.1 Python ETL Service

**File**: `apps/app/lib/api/services/python-etl.service.ts`

| Task | Status |
|------|--------|
| `transform()` method | ✅ Does not exist |
| `emit()` method | ✅ Does not exist |
| `discoverSchema` | ✅ Keep |
| `collect` | ✅ Keep |
| `deltaCheck` | ✅ Keep |
| `testConnection` | ✅ Keep |
| `getDbtModels` | ✅ Keep |

**Action**: None. Client is already clean.

---

## Phase 3: Transform Step UI Refactoring

**File**: `apps/app/app/workspace/data-pipelines/new/transform-step.tsx`

### 3.1 Remove `transformScript` State and Logic ✅ Done

| Task | Action | Status |
|------|--------|--------|
| Remove state | Delete `transformScript`, `setTransformScript` | ✅ |
| Remove from `TransformConfig` | Removed from interface | ✅ |
| Remove from `handleAddTransform` | No longer passed to `newTransform` | ✅ |
| Remove from `handleEditTransform` | Removed `setTransformScript` | ✅ |
| Remove from form UI | No Python editor exists; ensure no hidden inputs | ✅ (no editor) |

### 3.2 Standardize on dbt Models

| Task | Status |
|------|--------|
| dbt model selector | ✅ Already present (badge-based multi-select) |
| `dbtModels` in `TransformConfig` | ✅ Already stored |
| Empty selection = run all | ✅ Already implemented |

**Action**: Remove `transformScript` state and all references; keep dbt model selector as primary.

### 3.3 Clean State (Already Done)

| Task | Status |
|------|--------|
| `_transformMode` | ✅ Removed |
| `_handleAutoGenerate` | ✅ Removed |
| `_handleFieldMapping` | ✅ Removed |
| `_handleRemoveMapping` | ✅ Removed |
| `PythonScriptEditor` | ✅ Removed |

---

## Phase 4: Onboarding UI

**File**: `apps/app/app/onboarding/connect/[connector]/page.tsx`

| Task | Status |
|------|--------|
| Eliminate hardcoded forms | ✅ Uses `schema.fields` from metadata |
| Dynamic registry | ✅ `useConnectorMetadata` + `buildConnectionSchemasFromMetadata` |
| Sync Mode toggle | ✅ Full Sync vs Log-Based CDC with `sync_mode` |

**Action**: None. Onboarding is aligned.

---

## Phase 5: Pipeline Detail & Edit Pages

### 5.1 Remove Transform Script UI ✅ Done

**Files**:
- `apps/app/app/workspace/data-pipelines/[id]/page.tsx`
- `apps/app/app/workspace/data-pipelines/[id]/edit/page.tsx`
- `apps/app/app/workspace/data-pipelines/new/page.tsx`

| Task | Action | Status |
|------|--------|--------|
| Pipeline detail page | Removed `transformScript` display section | ✅ |
| Edit page | Removed `transformScript` from sync logic | ✅ |
| New pipeline page | No longer passes `transformScript` to API | ✅ |

### 5.2 Destination Schemas Page ✅ Done

**File**: `apps/app/app/workspace/destination-schemas/page.tsx`

| Task | Action | Status |
|------|--------|--------|
| Transform column | Uses `accessorKey: "transformScript"` but cell renders "dbt" badge | ✅ Done |
| Remove `transformScript` column | N/A — column shows "dbt" only, not script content | ✅ Done |

---

## Phase 6: API & Schema Cleanup (Optional / Deferred)

| Area | Action | Priority | Status |
|------|--------|----------|--------|
| `destination-schema.service.ts` | Keep `transformScript` in schema for DB backward compat | Low | Deferred |
| `create-destination-schema.dto.ts` | Mark `transformScript` deprecated | Low | Deferred |
| `pipeline-destination-schemas.schema.ts` | Keep column; no migration needed | Low | ✅ N/A |
| `collector-step.tsx` | Remove `transformScript` from `CollectorConfig` | Low | ✅ Done |

**Rationale**: DB column and DTOs persist for backward compatibility. UI and API payload should stop using them.

---

## Phase 7: Verification & Final Cleanup

### 7.1 No Extra Dependencies ✅ Done

| Check | Action | Status |
|-------|--------|--------|
| Direct DB connections | No SQLAlchemy, PyMongo in frontend | ✅ Done |
| Legacy toggles | No "Legacy" vs "Meltano" mode switches in UI | ✅ Done |
| Note | `testConnectionLegacy` etc. are API compat layers, not engine toggles | ✅ OK |

### 7.2 User Messages ✅ Done

| Check | Status |
|------|--------|
| `PipelineRunTracker` | ✅ Displays `run.errorMessage` (user_message from API) |
| CDC guidance | ✅ CDC error block with copy button |
| 502 detail | ✅ API passes raw detail for CDC guidance |

---

## Execution Order

| Step | Phase | Effort | Status |
|------|-------|--------|--------|
| 1 | 1.1 Remove transform_script from NestJS | 0.5h | ✅ Done |
| 2 | 3.1 Remove transformScript from transform-step | 1h | ✅ Done |
| 3 | 5.1 Remove transformScript on detail/edit/new pages | 0.5h | ✅ Done |
| 4 | 6 collector-step: Remove transformScript from CollectorConfig | 0.25h | ✅ Done |
| 5 | 7.1 Verification | — | ✅ Done |

**All alignment tasks complete.**

---

## Summary: Done

| Done |
|------|
| NestJS transform_script removed from payload |
| state_id always passed |
| Transform step: transformScript removed entirely |
| Pipeline detail/edit/new: transformScript UI removed |
| CollectorConfig: transformScript removed |
| PythonScriptEditor removed |
| Onboarding metadata + Sync toggle |
| PipelineRunTracker user_message |
| Destination schemas shows "dbt" |
| No direct DB / Legacy toggles |

---

## Verification Checklist

| Item | Status |
|------|--------|
| NestJS `runMeltanoPipeline` does not send `transform_script` | ✅ Done |
| `state_id` is always passed for pipeline runs | ✅ Done |
| Transform step has no `transformScript` state or UI | ✅ Done |
| Pipeline detail and edit pages do not show transform script | ✅ Done |
| New pipeline page does not pass `transformScript` to API | ✅ Done |
| No `PythonScriptEditor` or legacy transform UI | ✅ Done |
| Onboarding uses metadata-driven forms + Sync Mode toggle | ✅ Done |
| `PipelineRunTracker` displays `user_message` for CDC errors | ✅ Done |
| Destination schemas Transform column shows "dbt" | ✅ Done |
| No direct DB connections (SQLAlchemy/PyMongo) in frontend | ✅ Done |
| No Legacy vs Meltano mode toggles in UI | ✅ Done |
| `pnpm build` succeeds | ⏳ (pre-existing resolver error) |

---

## References

- `docs/CLEAN_ENGINE_FRONTEND_IMPLEMENTATION_PLAN.md` — FE phases 1–6
- `docs/CLEAN_ENGINE_TESTING_AND_FRONTEND_PLAN.md` — Backend verification
- `apps/etl/main.py` (lines 398–401) — 410 Gone for `transform_script`
