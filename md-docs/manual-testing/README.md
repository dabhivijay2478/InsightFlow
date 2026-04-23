# MantrixFlow — Manual Testing Guide

Complete manual QA reference for all pipeline sources, destinations, dbt layer, and normalisation rules.

---

## Folder Structure

```
manual-testing/
├── README.md                  ← this file (index + universal checklist)
├── dbt-layer.md               ← all dbt SQL scenarios & verification
├── normalisation.md           ← rename / cast / exclude rule testing
├── sources/
│   ├── stripe.md              ← 19 streams
│   ├── shopify.md             ← 14 streams
│   ├── hubspot.md             ← 14 streams
│   ├── github.md              ← 12 streams
│   ├── notion.md              ← 3 streams
│   ├── postgres.md            ← 3 source tables
│   ├── mysql.md               ← 3 source tables
│   ├── mariadb.md             ← 3 source tables
│   ├── sqlite.md              ← 3 source tables
│   └── cockroachdb.md         ← 3 source tables
└── destinations/
    ├── postgres.md            ← PostgreSQL destination guide
    ├── mysql.md               ← MySQL destination guide
    ├── mariadb.md             ← MariaDB destination guide
    ├── sqlite.md              ← SQLite destination guide
    └── cockroachdb.md         ← CockroachDB destination guide
```

---

## Services Required

| Service | Default port |
|---------|-------------|
| Next.js app | `http://localhost:3000` |
| Go API | `http://localhost:5000` |
| Python ELT server | `http://localhost:8000` |

Start all three before any manual test. See [`testing-local.md`](../testing-local.md) for start commands.

---

## Universal Pipeline Build Checklist

Follow this sequence for **every** source × destination scenario:

### Step 1 — Create connections

1. `Workspace ▸ Connections ▸ New source` → select connector, fill credentials, **Test connection**.
2. `Workspace ▸ Connections ▸ New destination` → select DB connector, fill credentials, **Test connection**.

### Step 2 — Pre-create destination table

The ELT engine **never creates tables**. You must run the DDL in the destination DB before triggering a run.

```sql
-- Example for PostgreSQL destination
CREATE TABLE IF NOT EXISTS analytics.stripe_customers_hd (
    id          TEXT PRIMARY KEY,
    email       TEXT,
    name        TEXT,
    created     BIGINT,
    currency    TEXT,
    synced_at   TIMESTAMPTZ DEFAULT NOW()
);
```

> ⚠️ If the destination table does not exist, Phase 0 fails with a named error and `rows_written = 0`.

### Step 3 — Build the pipeline

1. `Workspace ▸ Data pipelines ▸ New`.
2. **Config panel** — name the pipeline, set sync mode (`FULL_TABLE` or `INCREMENTAL`), set schedule.
3. **Source panel** — add each stream as `schema.table` (SaaS: `stripe.customers`; DB: `public.users`).
4. **Destination panel** — per stream, enter the pre-created destination table as `schema.table`.
5. Click **Discover** — column pills and PK hints must appear. ✅ If no pills appear, the destination table or connection is wrong.
6. **SQL model editor** — write dbt SQL using `{{ source('raw', 'duckdb_name') }}`. Click **Validate SQL**.
7. **Normalisation panel** *(optional)* — add rename / cast / exclude rules.
8. **Preview panel** — preview each stream's dbt output.
9. **Save**.

### Step 4 — Run and verify

1. Click **Run** in the pipeline list or builder header.
2. Open **Run Status drawer**:
   - ✅ **Extract + Stage** row shows `rows_read > 0`.
   - ✅ **dbt Transform** row shows `models_run = N` (one per stream).
   - ✅ **Deliver** row shows destination table chip(s).
3. In the destination DB:
   ```sql
   SELECT COUNT(*) FROM analytics.stripe_customers_hd;
   -- Must be > 0
   ```

---

## DuckDB Naming Reference

### SaaS sources — `{source}__{stream}`

| Source | Stream | DuckDB staging name |
|--------|--------|---------------------|
| stripe | customers | `stripe__customers` |
| stripe | charges | `stripe__charges` |
| shopify | products | `shopify__products` |
| hubspot | contacts | `hubspot__contacts` |
| github | issues | `github__issues` |
| notion | pages | `notion__pages` |

### DB sources — `{schema}__{table}`

| Source | Stream key | DuckDB staging name |
|--------|-----------|---------------------|
| postgres | `public.users` | `public__users` |
| mysql | `mydb.products` | `mydb__products` |
| mariadb | `app.events` | `app__events` |
| sqlite | `main.tasks` | `main__tasks` |
| cockroachdb | `public.sessions` | `public__sessions` |

---

## Sync Modes

| Mode | Behaviour | Replication key required |
|------|-----------|--------------------------|
| `FULL_TABLE` | Drops and re-syncs all rows every run | No |
| `INCREMENTAL` | Syncs only rows where `column > last_cursor` | Yes — must exist in source |

---

## Write Dispositions

| Source has PK | Destination behaviour | Run callback field |
|--------------|----------------------|-------------------|
| Yes | `merge` — upsert by PK | `delivered_tables > 0` |
| No | `append` — inserts only | `no_pk_warnings` array populated |

> Amber banner in Run Status drawer = no PK warning. This is not a failure.

---

## Phase-by-Phase Failure Modes

| Phase | What fails | Run status | Error shown |
|-------|-----------|-----------|-------------|
| Phase 0 | Destination table missing | `failed` | `"table analytics.hd does not exist"` |
| Phase 0 | Column mismatch | `failed` | `"column X not present in analytics.hd"` |
| Phase 0 | Incremental key missing | `failed` | `"cursor column updated_at not found"` |
| Phase 1 | Source credentials invalid | `failed` | connector-specific auth error |
| Phase 2 | Invalid dbt SQL | `failed` | dbt compile error with line number |
| Phase 3 | Destination write error | `failed` | DB-level error with table name |

---

## Stream Matrix — All 62 Streams + 15 DB Tables

### SaaS

| Source | Count | All streams |
|--------|-------|-------------|
| Stripe | 19 | customers, charges, invoices, subscriptions, products, prices, events, balance_transactions, coupons, payment_intents, payment_methods, refunds, disputes, payouts, plans, tax_rates, credit_notes, promotion_codes, setup_intents |
| Shopify | 14 | products, orders, customers, draft_orders, custom_collections, smart_collections, pages, blogs, articles, locations, price_rules, themes, countries, collects |
| HubSpot | 14 | contacts, companies, deals, tickets, products, line_items, quotes, calls, emails, meetings, notes, tasks, feedback_submissions, owners |
| GitHub | 12 | issues, pull_requests, stargazers, events, commits, branches, releases, tags, contributors, milestones, labels, forks |
| Notion | 3 | databases, pages, users |

### DB Sources

| Source | Stream keys |
|--------|-------------|
| PostgreSQL | `public.users`, `public.orders`, `public.payments` |
| MySQL | `mydb.products`, `mydb.inventory`, `mydb.categories` |
| MariaDB | `app.events`, `app.logs`, `app.sessions` |
| SQLite | `main.tasks`, `main.notes`, `main.tags` |
| CockroachDB | `public.sessions`, `public.accounts`, `public.transactions` |

---

## Quick Scenario Index

| Scenario | Coverage file |
|----------|-------------|
| Full-table sync, all SaaS streams | `sources/{source}.md` |
| Incremental sync with cursor | `sources/{source}.md` |
| JSON column key filtering | `dbt-layer.md` |
| Column rename in destination | `normalisation.md` |
| Type cast (text → integer) | `normalisation.md` |
| Column exclusion | `normalisation.md` |
| dbt GROUP BY aggregate | `dbt-layer.md` |
| dbt window function | `dbt-layer.md` |
| Missing destination table | each destination guide |
| No-PK append warning | each destination guide |
| Multi-stream single pipeline | each source guide |
| Concurrent run rejection | each source guide |
| Cron schedule | each source guide |
