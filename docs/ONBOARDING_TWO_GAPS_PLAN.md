# Onboarding — Two Reported Gaps: Plan & Verification

> **Purpose**: Address the two reported frontend gaps (Onboarding Persistence, Sync Mode Toggle). **Both are already implemented** in the current codebase.

---

## Summary

| Gap | Reported | Actual Status | Evidence |
|-----|----------|---------------|----------|
| **1. Onboarding Persistence** | "Never calls API" | ✅ **Implemented** | `createConnection.mutateAsync` at line 201 |
| **2. Sync Mode Toggle** | "No UI for sync method" | ✅ **Implemented** | RadioGroup at lines 398–436 |

---

## Gap 1: Onboarding Persistence

### Reported Issue
> The `onSubmit` function calls `addDataSource(dataSource)` which only updates local state. Must call `DataSourcesService.createDataSource(data)` to persist to DB.

### Actual Implementation

**File**: `apps/app/app/onboarding/connect/[connector]/page.tsx`

**Line 84**: `const createConnection = useCreateConnection(organizationId);`

**Lines 195–201**:
```typescript
const connectionData: CreateConnectionDto = {
  name: data.name || `${connector} Connection`,
  connection_type: connector as CreateConnectionDto["connection_type"],
  config: config as unknown as CreateConnectionDto["config"],
};

const created = await createConnection.mutateAsync(connectionData);
```

**What `createConnection.mutateAsync` does** (via `DataSourcesService.createConnection`):
1. **Creates data source** in DB via `createDataSource()`
2. **Creates connection** via `createOrUpdateConnection()`
3. Returns the created object with `id` from the API

**Lines 203–212**: `addDataSource` uses `created.id` from the API response, not a local `ds_${Date.now()}`.

### Verification
```bash
grep -n "createConnection\|mutateAsync" apps/app/app/onboarding/connect/\[connector\]/page.tsx
# Expected: line 84 (createConnection), line 201 (mutateAsync)
```

---

## Gap 2: Sync Mode Toggle (UI)

### Reported Issue
> The form does not render a toggle for Full Sync vs Log-Based CDC. `getDefaultValues` has `sync_mode` but no UI component.

### Actual Implementation

**File**: `apps/app/app/onboarding/connect/[connector]/page.tsx`

**Lines 398–436** (inside the database connector form):
```tsx
{/* Sync Mode toggle for database connectors (Phase 5) */}
{isDatabaseConnector && (
  <FormField
    control={form.control}
    name="sync_mode"
    render={({ field: formField }) => (
      <FormItem className="rounded-lg border p-4">
        <FormLabel>Sync Mode</FormLabel>
        <FormControl>
          <RadioGroup
            value={formField.value || "full"}
            onValueChange={formField.onChange}
            className="flex-col gap-2 pt-2"
          >
            <div className="flex items-center space-x-2">
              <RadioGroupItem value="full" id="sync-full" />
              <FormLabel htmlFor="sync-full" className="cursor-pointer font-normal">
                Full Sync
              </FormLabel>
            </div>
            <div className="flex items-center space-x-2">
              <RadioGroupItem value="incremental" id="sync-cdc" />
              <FormLabel htmlFor="sync-cdc" className="cursor-pointer font-normal">
                Log-Based CDC
              </FormLabel>
            </div>
          </RadioGroup>
        </FormControl>
        <p className="text-xs text-muted-foreground mt-2">
          {connector === "postgres" ? "CDC requires wal_level=logical..." : ...}
        </p>
      </FormItem>
    )}
  />
)}
```

**Lines 169–173**: `sync_mode` is included in the submitted config:
```typescript
if (["postgres", "mysql", "mongodb"].includes(connector)) {
  config.sync_mode = data.sync_mode || "full";
}
```

### Verification
```bash
grep -n "sync_mode\|RadioGroup\|Log-Based CDC" apps/app/app/onboarding/connect/\[connector\]/page.tsx
# Expected: sync_mode in getDefaultValues, FormField, config; RadioGroup with Full Sync / Log-Based CDC
```

---

## Corrected "All Clear" Checklist

| Feature | Status | Verification |
|---------|--------|--------------|
| **Legacy Removal** | ✅ Perfect | All legacy DB drivers and `/transform` routes removed |
| **dbt Model Selection** | ✅ Perfect | FE → NestJS → FastAPI → Meltano |
| **CDC State Management** | ✅ Perfect | `state_id` passed through service layers |
| **Onboarding Persistence** | ✅ **Done** | `createConnection.mutateAsync` persists to DB |
| **Sync Mode Toggle** | ✅ **Done** | RadioGroup (Full Sync / Log-Based CDC) in database connector form |

---

## If Reviewer Still Sees Gaps

1. **Branch**: Confirm they are on `/fix/missing-data` (or the branch that includes these changes).
2. **File**: Ensure they are looking at `apps/app/app/onboarding/connect/[connector]/page.tsx`.
3. **Lines**: Persistence at ~201; Sync Mode UI at ~398–436.

---

## Conclusion

Both reported gaps are implemented. No code changes are required. The onboarding flow:

1. Persists connections via `createConnection.mutateAsync` (API).
2. Shows a Sync Mode toggle (Full Sync / Log-Based CDC) for postgres, mysql, mongodb.
