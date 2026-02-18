# Response to Reviewer — Four Plumbing Gaps

> **Purpose**: Provide evidence that all four reported plumbing gaps are **already implemented** in the current codebase. Use this document to verify or share with reviewers.

---

## Why the Reviewer May See "Gaps"

Possible reasons for the discrepancy:
1. **Reviewer is on an older commit** — Ensure they pull the latest `main` (or your branch) with the Clean Engine implementation
2. **Different branch** — Changes may be on a feature branch not yet merged
3. **Different repo/fork** — Verify they're reviewing the same repository

---

## Evidence: All Four Gaps Are Implemented

### 1. dbt_models Payload (FastAPI) — ✅ IMPLEMENTED

**File**: `apps/etl/main.py`

**Request model** (line 151):
```python
class RunMeltanoPipelineRequest(BaseModel):
    # ... existing fields ...
    dbt_models: Optional[List[str]] = None  # Selected dbt models; empty/None = run all
    state_id: Optional[str] = None
```

**Endpoint passes to run_pipeline_job** (lines 413–424):
```python
result = await run_pipeline_job(
    job_name,
    source_type,
    payload.source_connection_config,
    dest_type,
    payload.dest_connection_config,
    state_id=payload.state_id,
    checkpoint=payload.checkpoint,
    sync_mode=payload.sync_mode,
    dbt_models=payload.dbt_models,   # ← HERE
    timeout_seconds=TAP_TIMEOUT_SECONDS,
)
```

---

### 2. NestJS Service & Pipeline Alignment — ✅ IMPLEMENTED

**File**: `apps/api/src/modules/data-pipelines/services/python-etl.service.ts`

**Method signature** (line 262): `dbtModels?: string[]`

**POST body** (line 311): `dbt_models: dbtModels ?? null`

**File**: `apps/api/src/modules/data-pipelines/services/pipeline.service.ts`

**Extract and pass** (lines 589–604):
```typescript
const dbtModels = (destinationSchema.dbtModels as string[] | null) ?? undefined;
const result = await this.pythonETLService.runMeltanoPipeline({
  // ...
  stateId: `pipeline_${pipeline.id}`,
  dbtModels: dbtModels?.length ? dbtModels : undefined,
});
```

**Note**: `dbt_models` are stored in `pipeline_destination_schemas.dbt_models` (destination schema), not `pipeline.transformations`. This is by design (Option A from the implementation plan).

---

### 3. Onboarding Persistence — ✅ IMPLEMENTED

**File**: `apps/app/app/onboarding/connect/[connector]/page.tsx`

**API call** (lines 84, 201):
```typescript
const createConnection = useCreateConnection(organizationId);
// ...
const created = await createConnection.mutateAsync(connectionData);
```

**What `createConnection` does** (see `apps/app/lib/api/services/data-sources.service.ts`):
1. Calls `DataSourcesService.createDataSource()` — creates data source in DB
2. Calls `DataSourcesService.createOrUpdateConnection()` — configures connection
3. Returns `created` with `id` from API

**Local store sync** (lines 203–212): `addDataSource` uses `created.id` from API, not `ds_${Date.now()}`.

---

### 4. CDC Replication Method Mapping — ✅ IMPLEMENTED

**File**: `apps/etl/orchestration/connection_mapper.py`

**_connection_config_to_extractor_config** (lines 164–168):
```python
# Map sync_mode to Meltano default_replication_method
if sync_mode == "incremental":
    config["default_replication_method"] = "LOG_BASED"
else:
    config["default_replication_method"] = "FULL_TABLE"
```

**_config_dict_to_env_vars** (lines 211–222): Converts `default_replication_method` → `MELTANO_EXTRACTOR_TAP_POSTGRES_DEFAULT_REPLICATION_METHOD` (and similarly for tap-mysql, tap-mongodb).

**sync_mode flow**: `run_pipeline_job` → `connection_config_to_meltano_env_for_pipeline(sync_mode=...)` → `connection_config_to_meltano_env(sync_mode=...)` → `_connection_config_to_extractor_config(sync_mode=...)`

---

## Corrected Final Status Summary

| Component | Status | Evidence |
|-----------|--------|----------|
| **Backend Requirements** | ✅ Perfect | requirements.txt slimmed |
| **Legacy Removal** | ✅ Perfect | Standalone `/transform` and `/emit` blocked |
| **UI Refactoring** | ✅ Excellent | PythonScriptEditor removed; dbt selector active |
| **dbt_models Plumbing** | ✅ **Done** | main.py:151,422; python-etl.service.ts:262,311; pipeline.service.ts:589,604 |
| **Onboarding Persistence** | ✅ **Done** | createConnection.mutateAsync → createDataSource + createOrUpdateConnection |
| **CDC Replication** | ✅ **Done** | connection_mapper.py:164-168, 211-222 |

---

## Verification Commands (for reviewer)

```bash
# 1. dbt_models in FastAPI
grep -n "dbt_models" apps/etl/main.py

# 2. NestJS python-etl.service
grep -n "dbtModels\|dbt_models" apps/api/src/modules/data-pipelines/services/python-etl.service.ts

# 3. pipeline.service
grep -n "dbtModels\|destinationSchema.dbtModels" apps/api/src/modules/data-pipelines/services/pipeline.service.ts

# 4. Onboarding API
grep -n "createConnection\|mutateAsync" apps/app/app/onboarding/connect/\[connector\]/page.tsx

# 5. CDC mapping
grep -n "default_replication_method\|sync_mode" apps/etl/orchestration/connection_mapper.py
```

---

## Conclusion

All four plumbing gaps are implemented in the current codebase. The Clean Engine is fully operational. If the reviewer still sees gaps, they should:

1. Pull the latest code
2. Run the verification commands above
3. Open the referenced files and line numbers
