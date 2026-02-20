# Database Migrations Status

## Missing Migration (Now Fixed)

### `dbt_models` column on `pipeline_destination_schemas`

**Status**: ✅ **Applied** (via Supabase MCP)

**Issue**: The schema and application expected `dbt_models` (jsonb) for storing selected dbt model names, but the column was missing in the database.

**Migration**: `0021_add_dbt_models_to_destination_schemas.sql`

```sql
ALTER TABLE pipeline_destination_schemas
ADD COLUMN IF NOT EXISTS dbt_models jsonb;

COMMENT ON COLUMN pipeline_destination_schemas.dbt_models IS 'Selected dbt model names; empty or null means run all models';
```

---

## Migration Files (apps/api)

| File | Purpose |
|------|---------|
| 0000_closed_centennial.sql | Initial schema |
| 0001_greedy_lucky_pierre.sql | |
| 0013_refactor_to_dynamic_data_sources.sql | Dynamic data sources |
| 0014_pipeline_lifecycle.sql | Pipeline lifecycle |
| 0015_pipeline_scheduling.sql | Scheduling |
| 0016_pipeline_incremental_sync_fixes.sql | Incremental sync |
| 0017_add_polling_trigger_type.sql | Polling trigger |
| 0018_add_transform_script.sql | transform_script column |
| 0019_remove_column_mappings.sql | Remove column_mappings |
| 0020_add_migration_state.sql | migration_state on pipelines |
| 0021_add_dbt_models_to_destination_schemas.sql | **dbt_models** column |

---

## How to Run Migrations

**Local / Non-Supabase**:
```bash
cd apps/api
pnpm db:migrate
# or
pnpm db:migrate:kit
```

**Supabase**: Migrations can be applied via Supabase Dashboard SQL editor or MCP. The `dbt_models` migration has been applied to the connected Supabase project.

---

## Drizzle Journal Note

The `meta/_journal.json` only lists 0000 and 0001. Migrations 0013–0021 may have been added manually. The custom `migrate.ts` script uses `drizzle-orm/migrator` which reads SQL files from the migrations folder. Ensure all migrations run in order when using `pnpm db:migrate`.
