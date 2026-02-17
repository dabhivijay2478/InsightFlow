# Clean Engine — Final Gaps Implementation Plan

> **Context**: The Clean Engine is ~90% complete. This plan addresses three critical gaps and validation cleanup to reach production readiness.

## Implementation Status (Completed)

| Gap | Status | Notes |
|-----|--------|-------|
| Gap 4: Transform step validation | ✅ Done | `handleContinue` validates `destinationTable` per transformer; blocks continue with toast |
| Gap 1.4: Persist dbt_models | ✅ Done | Migration, DTOs, create flow, pipeline.service pass-through |
| Gap 1.1–1.3: dbt_models plumbing | ✅ Done | FastAPI, NestJS, pipeline runner; DBT_SELECT env for dbt |
| Gap 2: Onboarding persistence | ✅ Done | `useCreateConnection` + `CreateConnectionDto`; API persist + store sync |
| Gap 3: Sync mode → replication | ✅ Done | `sync_mode` → `default_replication_method` (FULL_TABLE/LOG_BASED) in connection_mapper |

---

## Gap 1: Missing `dbt_models` Plumbing

**Problem**: The frontend dbt Model Selector stores `dbtModels` in `TransformConfig`, but this is never persisted or passed to the ETL. Selected models are ignored during execution.

### 1.1 Add `dbt_models` to FastAPI Request Model

**File**: `apps/etl/main.py`

| Task | Action |
|------|--------|
| Add field | Add `dbt_models: Optional[List[str]] = None` to `RunMeltanoPipelineRequest` |
| Remove deprecated | Remove `transform_script` from `RunMeltanoPipelineRequest` (already returns 410; clean up model) |

```python
class RunMeltanoPipelineRequest(BaseModel):
    # ... existing fields ...
    dbt_models: Optional[List[str]] = None  # Selected dbt models; empty/None = run all
    state_id: Optional[str] = None
    # Remove: transform_script
```

### 1.2 Pass `dbt_models` Through Pipeline Runner

**File**: `apps/etl/orchestration/pipeline_runner.py`

| Task | Action |
|------|--------|
| Add param | Add `dbt_models: Optional[List[str]] = None` to `run_pipeline_job` |
| Pass to Meltano | If `dbt_models` is provided, add `--select` or equivalent to `run_args` for dbt; or set env var for dbt model selection |

**Note**: Meltano runs `tap → dbt → target`. The dbt-postgres utility runs all models in `transform/models/` by default. To run specific models, use `dbt run --select model1 model2` or `dbt run --select path.to.model`. Check dbt-postgres Meltano config for `run_args` or `command` override.

**File**: `apps/etl/orchestration/meltano_runner.py` (or pipeline_runner)

| Task | Action |
|------|--------|
| Model selection | If `dbt_models` provided, pass `--select "model1 model2"` to dbt step; else run all |

### 1.3 Add `dbt_models` to NestJS Service

**File**: `apps/api/src/modules/data-pipelines/services/python-etl.service.ts`

| Task | Action |
|------|--------|
| Add param | Add `dbtModels?: string[]` to `runMeltanoPipeline` options |
| Add to body | Add `dbt_models: dbtModels ?? null` to POST body |

### 1.4 Persist `dbt_models` and Pass from Pipeline Service

**Option A: Add `dbt_models` to `pipeline_destination_schemas`**

| Task | Action |
|------|--------|
| Migration | Add `dbt_models jsonb` column to `pipeline_destination_schemas` |
| DTO | Add `dbtModels?: string[]` to create/update destination schema DTOs |
| Create flow | In `new/page.tsx`, pass `dbtModels: firstTransformer.dbtModels` to `createDestinationSchema` |
| Run flow | In `pipeline.service.ts`, read `destinationSchema.dbtModels` and pass to `runMeltanoPipeline` |

**Option B: Add `dbt_models` to pipeline table**

| Task | Action |
|------|--------|
| Migration | Add `dbt_models jsonb` to `pipelines` table |
| Run flow | Read from pipeline, pass to `runMeltanoPipeline` |

**Recommended**: Option A — destination schema is the transform config holder; keeps pipeline lean.

---

## Gap 2: Onboarding Persistence Gap

**Problem**: `onSubmit` in onboarding only calls `addDataSource` (local Zustand store). Connections are lost on refresh.

### 2.1 Call API to Persist Connection

**File**: `apps/app/app/onboarding/connect/[connector]/page.tsx`

| Task | Action |
|------|--------|
| Import | Add `useCreateConnection` from `@/lib/api/hooks/use-data-sources` |
| Require org | Ensure `currentOrganization?.id` exists before submit; show error if not |
| Build config | Map form data to `CreateConnectionDto` (same logic as data-sources `handleConnect`) |
| Call API | Call `createConnection.mutateAsync(connectionData)` |
| Use returned ID | Use `result.id` (from API) as `dataSourceId` instead of `ds_${Date.now()}` |
| Update store | Call `addDataSource` with API-returned connection for local state sync |
| Handle errors | Show toast on failure; do not navigate on error |

### 2.2 Build Config from Form Data

Reuse logic from `data-sources/page.tsx` `handleConnect`:

- Map `name` → `connectionData.name`
- Map `connection_type` from connector (postgres, mysql, mongodb)
- Build `config` from all other form fields (host, port, database, etc.)
- Include `sync_mode` in config
- Handle SSL, MongoDB connection string vs individual fields

### 2.3 Edge Cases

| Case | Action |
|------|--------|
| No organization | Redirect to org step or show "Select organization first" |
| OAuth connectors | Keep current mock flow; persistence deferred |
| File upload | Keep current mock flow; persistence deferred |

---

## Gap 3: Sync Mode vs Replication Method Alignment

**Problem**: Frontend uses `sync_mode` ("full" | "incremental"). Meltano taps expect `default_replication_method` (FULL_TABLE, INCREMENTAL, LOG_BASED). For Postgres CDC, we need LOG_BASED.

### 3.1 Mapping

| Frontend `sync_mode` | Postgres | MySQL | MongoDB |
|----------------------|----------|-------|---------|
| `full` | FULL_TABLE | FULL_TABLE | FULL_TABLE |
| `incremental` | LOG_BASED (CDC) | LOG_BASED (binlog) | LOG_BASED (oplog) |

### 3.2 Connection Config

**File**: `apps/app` (ConnectionSheet, onboarding)

| Task | Status |
|------|--------|
| Store `sync_mode` | ✅ Already stored in config |
| Values | `"full"` \| `"incremental"` |

### 3.3 ETL: Use `sync_mode` from Connection Config or Payload

**File**: `apps/etl/orchestration/connection_mapper.py`

| Task | Action |
|------|--------|
| Add replication method | In `_connection_config_to_extractor_config`, when `connection_config` has `sync_mode` (or `replication_method`): |
| | - If `sync_mode == "incremental"` and source is postgresql → add `default_replication_method: "LOG_BASED"` |
| | - If `sync_mode == "incremental"` and source is mysql → add `default_replication_method: "LOG_BASED"` |
| | - If `sync_mode == "incremental"` and source is mongodb → metadata already has LOG_BASED |
| | - If `sync_mode == "full"` or absent → add `default_replication_method: "FULL_TABLE"` |
| Env vars | Ensure `_config_dict_to_env_vars` emits `MELTANO_EXTRACTOR_TAP_POSTGRES_DEFAULT_REPLICATION_METHOD` etc. |

**Alternative**: Pass `sync_mode` from `RunMeltanoPipelineRequest` into `run_pipeline_job` and merge into source_connection_config before calling connection_mapper. Pipeline has `syncMode`; API passes it. So we can use `payload.sync_mode` in `run_meltano_pipeline` and merge into the env build.

**Recommended**: Use `payload.sync_mode` (from pipeline) as the source of truth. Merge it into the extractor config when building env. The connection config's `sync_mode` can override when creating pipelines; for now pipeline.syncMode is the runtime source.

### 3.4 Pipeline Service Already Passes syncMode

**File**: `apps/api/src/modules/data-pipelines/services/pipeline.service.ts`

| Task | Status |
|------|--------|
| Pass syncMode | ✅ Passes `syncMode: syncType` to `runMeltanoPipeline` |

**File**: `apps/api/src/modules/data-pipelines/services/python-etl.service.ts`

| Task | Action |
|------|--------|
| Verify | Ensure `sync_mode` is sent in POST body (already is) |

**File**: `apps/etl/main.py`

| Task | Action |
|------|--------|
| Use payload.sync_mode | In `run_meltano_pipeline`, pass `sync_mode` to `run_pipeline_job` |
| Merge into env | In `connection_config_to_meltano_env_for_pipeline`, accept optional `sync_mode` and add `default_replication_method` to extractor config |

---

## Gap 4: Transform Step Validation Cleanup

**Problem**: `handleContinue` has legacy validation that checks for "transform scripts". Should validate destination table selection instead.

### 4.1 Update handleContinue

**File**: `apps/app/app/workspace/data-pipelines/new/transform-step.tsx`

| Task | Action |
|------|--------|
| Remove | Legacy "hasValidTransformers" check for transform scripts |
| Add | Validate that each transformer has `destinationTable` selected |
| Message | "Please select a destination table for each transformer" |
| Logic | `collectors.some(c => c.transformers?.some(t => !t.destinationTable))` → block continue |

---

## Execution Order

| Step | Gap | Effort | Dependencies |
|------|-----|--------|--------------|
| 1 | 4. Transform validation cleanup | 0.5h | None |
| 2 | 1.4 Persist dbt_models (migration + DTOs + create flow) | 1.5h | None |
| 3 | 1.1–1.3 dbt_models plumbing (FastAPI, NestJS, pipeline runner) | 1.5h | 2 |
| 4 | 2. Onboarding persistence | 1.5h | None |
| 5 | 3. Sync mode → replication method in ETL | 1h | None |

**Total**: ~6h

---

## Verification Checklist

- [ ] dbt model selection flows: UI → API → ETL → Meltano
- [ ] Onboarding connection persists to DB; survives refresh
- [ ] Sync mode "incremental" results in LOG_BASED for Postgres when pipeline runs
- [ ] Transform step blocks continue when destination table not selected
- [ ] `pnpm build` succeeds (fix onboarding resolver TypeScript if needed)

---

## References

- `docs/PURE_MELTANO_ENGINE_ALIGNMENT_PLAN.md` — Alignment status
- `apps/etl/orchestration/connection_mapper.py` — Env mapping
- `apps/etl/meltano.yml` — Tap defaults
- `apps/app/app/workspace/data-sources/page.tsx` — Config build logic for createConnection
