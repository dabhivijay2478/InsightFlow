# UI Testing — Pipeline Builder

Step-by-step manual UI tests for every source, every destination, and every stream through the pipeline builder interface.

---

## File Index

### Universal
| File | Purpose |
|------|---------|
| [builder-walkthrough.md](builder-walkthrough.md) | Panel-by-panel guide — source config, normalisation, dbt SQL, preview, schedule, run |

### SaaS Sources
| File | Streams | Destinations tested |
|------|---------|-------------------|
| [saas/stripe-ui.md](saas/stripe-ui.md) | 19 Stripe streams | All 5 (PG, MySQL, MariaDB, SQLite, CockroachDB) |
| [saas/shopify-ui.md](saas/shopify-ui.md) | 14 Shopify streams | All 5 |
| [saas/hubspot-ui.md](saas/hubspot-ui.md) | 14 HubSpot streams | All 5 |
| [saas/github-ui.md](saas/github-ui.md) | 12 GitHub streams | All 5 |
| [saas/notion-ui.md](saas/notion-ui.md) | 3 Notion streams | All 5 |

### DB Sources
| File | Streams | Destinations tested |
|------|---------|-------------------|
| [db/postgres-ui.md](db/postgres-ui.md) | 3 PG tables | All 5 |
| [db/mysql-ui.md](db/mysql-ui.md) | 3 MySQL tables | All 5 |
| [db/mariadb-ui.md](db/mariadb-ui.md) | 3 MariaDB tables | All 5 |
| [db/sqlite-ui.md](db/sqlite-ui.md) | 3 SQLite tables | All 5 |
| [db/cockroachdb-ui.md](db/cockroachdb-ui.md) | 3 CockroachDB tables | All 5 |

---

## Universal 8-Phase UI Test Checklist

Every pipeline test must pass all 8 phases before marking ✅.

```
Phase 1  Source Panel        — credentials → test connection → streams → sync mode → cursor key
Phase 2  Destination Panel   — credentials → test connection → schema → table per stream
Phase 3  Normalisation Panel — rename / cast / exclude rules per stream
Phase 4  dbt SQL Panel       — SQL editor → validate → column preview
Phase 5  Preview Panel       — column names, types, sample rows correct
Phase 6  Schedule Panel      — None / Cron expression / interval saved correctly
Phase 7  Run & Status Drawer — all 6 pipeline phases green
Phase 8  Destination Verify  — correct rows, column names, data types in destination DB
```

---

## UI Element Reference

| Element | Location | Action |
|---------|----------|--------|
| **Add Source** button | Builder left panel | Opens source type selector |
| **Source type dropdown** | Source config drawer | Select Stripe / Shopify / etc. |
| **Test Connection** button | Source + Destination config | Shows ✅ green / ❌ red + error message |
| **Stream toggle** | Stream list | Checkbox per stream |
| **Sync Mode dropdown** | Per-stream row | `Full Table` / `Incremental` |
| **Cursor Field dropdown** | Appears when Incremental selected | Lists timestamp/int columns |
| **Add Destination** button | Builder right panel | Opens destination type selector |
| **Schema input** | Destination config | Enter destination schema name |
| **Table input** | Per-stream destination row | Enter destination table name |
| **Normalisation tab** | Middle builder panel | Opens rules editor |
| **+ Add Rule** button | Normalisation panel | Adds rename/cast/exclude row |
| **Rule type selector** | Per-rule row | `Rename` / `Cast` / `Exclude` |
| **Transform tab** | Middle builder panel | Opens dbt SQL editor |
| **SQL editor** | Transform panel | Write SELECT statement |
| **Validate SQL** button | SQL editor toolbar | Returns column list or error |
| **Preview tab** | Middle builder panel | Shows sample output rows |
| **Schedule tab** | Top pipeline bar | None / Cron / Interval |
| **Cron input** | Schedule panel | Enter cron expression |
| **Save Pipeline** button | Top bar | Persists all config |
| **Run Now** button | Top bar | Triggers immediate run |
| **Run status drawer** | Bottom slide-up | Shows 6-phase progress + logs |

---

## Sync Mode Test Matrix

| Stream type | Sync mode | Cursor field | Expected |
|------------|-----------|-------------|---------|
| Stripe charges | Incremental | `created` | Only new records since last run |
| Stripe customers | Full Table | — | Full re-sync every run |
| Shopify orders | Incremental | `updated_at` | Changed orders only |
| PG source table | Incremental | `updated_at` | Row-level watermark |
| Any stream | Full Table | — | Merge upsert on PK every time |

---

## Expected Run Status Phases

```
✅ Phase 0 — Pre-flight checks     (destination table exists, PK matches)
✅ Phase 1 — Extract + Stage       (source → DuckDB staging)
✅ Phase 2 — dbt Transform         (SQL applied in DuckDB)
✅ Phase 3 — Deliver               (DuckDB → destination DB upsert/append)
✅ Phase 4 — Cleanup               (staging tables dropped)
✅ Phase 5 — Callback              (run status written, webhook fired if set)
```

### Common Phase Failures

| Phase | Error | Fix |
|-------|-------|-----|
| Phase 0 | `relation "analytics.table" does not exist` | Create destination table first |
| Phase 0 | PK column mismatch | Ensure destination PK column matches dbt SQL output |
| Phase 1 | `401 Unauthorized` | Bad API credentials — re-enter and test connection |
| Phase 2 | `SQL compilation error` | Fix dbt SQL — validate before saving |
| Phase 3 | `column X does not exist in destination` | Add column to destination DDL |
| Phase 3 | `database is locked` (SQLite) | Only one concurrent run allowed |
