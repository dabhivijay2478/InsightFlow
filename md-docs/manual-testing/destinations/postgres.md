# PostgreSQL Destination — Manual Testing Guide

PostgreSQL is the most common and fully-featured destination. Supports merge (upsert by PK) and append write modes.

---

## Credential Setup

```json
{
  "host": "your-postgres-host",
  "port": 5432,
  "database": "analytics_db",
  "username": "writer_user",
  "password": "your_password",
  "ssl_mode": "disable"
}
```

The destination user needs `INSERT`, `UPDATE`, `SELECT` on the target tables. It does **not** need `CREATE TABLE`.

---

## Destination Schema Setup

Create the analytics schema before any pipeline run:

```sql
CREATE SCHEMA IF NOT EXISTS analytics;
GRANT ALL ON SCHEMA analytics TO writer_user;
```

---

## Scenario D-PG-1 — Basic Delivery: merge mode (with PK)

**Pre-create destination table:**
```sql
CREATE TABLE analytics.customers_hd (
    id       TEXT PRIMARY KEY,
    email    TEXT,
    name     TEXT,
    currency TEXT,
    created  BIGINT
);
```

**Pipeline:** Stripe → customers → PostgreSQL  
**dbt SQL:**
```sql
SELECT id, email, name, currency, created
FROM {{ source('raw', 'stripe__customers') }}
```

**Expected:** Run completes; `rows_written > 0`; second run upserts (no duplicates).

**Verify (after 2 runs):**
```sql
SELECT COUNT(*) FROM analytics.customers_hd;
-- Count must equal source, not doubled
```

---

## Scenario D-PG-2 — Append Mode (no PK)

**Pre-create table without PK:**
```sql
CREATE TABLE analytics.daily_revenue (
    currency      TEXT,
    charge_count  BIGINT,
    total_cents   BIGINT,
    synced_at     TIMESTAMPTZ DEFAULT NOW()
);
```

**Expected:** Each run appends new rows. `no_pk_warnings` in callback. Amber banner in Run Status drawer.

**Verify:**
```sql
SELECT COUNT(*) FROM analytics.daily_revenue;
-- Increases by N each run (N = row count from dbt model)
```

---

## Scenario D-PG-3 — Column Mismatch Error

**Pre-create table with extra required column:**
```sql
CREATE TABLE analytics.orders_strict (
    id            UUID PRIMARY KEY,
    amount        NUMERIC,
    status        TEXT,
    extra_column  TEXT NOT NULL  -- not in dbt SQL output
);
```

**dbt SQL (missing `extra_column`):**
```sql
SELECT id, amount, status FROM {{ source('raw', 'public__orders') }}
```

**Expected:**
- Phase 0 pre-flight fails: `"column extra_column not present in analytics.orders_strict"`
- Run status: `failed`
- `rows_written = 0`

---

## Scenario D-PG-4 — Missing Destination Table Error

Do NOT create the destination table. Set destination to `analytics.does_not_exist`.

**Expected:**
- Phase 0 fails: `"table analytics.does_not_exist does not exist"`
- Run status: `failed`

---

## Scenario D-PG-5 — JSONB Destination Column

**Pre-create table with a `jsonb` column:**
```sql
CREATE TABLE analytics.events_raw_col (
    id         TEXT PRIMARY KEY,
    event_type TEXT,
    raw_data   JSONB,
    created_at BIGINT
);
```

**dbt SQL (store the full JSON object in one column):**
```sql
SELECT
    id,
    type         AS event_type,
    data         AS raw_data,
    created
FROM {{ source('raw', 'stripe__events') }}
```

**Verify:**
```sql
SELECT raw_data->>'object' FROM analytics.events_raw_col LIMIT 5;
```

---

## Scenario D-PG-6 — Large Volume Delivery (stress test)

**Stream:** Stripe `events` (typically the highest-volume stream)  
**Sync mode:** `FULL_TABLE`

**Expected:**
- Run completes within a reasonable time (< 5 min for < 100K rows)
- `staging_size_bytes` in callback is non-zero
- `rows_written` matches source count

---

## Scenario D-PG-7 — SSL Destination Connection

Set `ssl_mode = "require"` in destination connection.  
**Expected:** Phase 3 delivery uses TLS; run completes successfully.

---

## Scenario D-PG-8 — Verify Run Metadata in Callback

After a successful run, check `pipeline_runs.run_metadata` via the API or database:

```sql
SELECT
    run_metadata->>'delivered_tables'    AS tables,
    run_metadata->>'rows_written'        AS rows,
    run_metadata->>'staging_size_bytes'  AS staging_bytes,
    run_metadata->>'dbt_models_run'      AS models
FROM pipeline_runs
WHERE pipeline_id = '<your-pipeline-id>'
ORDER BY created_at DESC
LIMIT 1;
```

**Expected:** All 4 fields non-null for a successful run.

---

## Destination Verification Queries

```sql
-- Check row count
SELECT COUNT(*) FROM analytics.<dest_table>;

-- Check column list (no extra or missing columns)
SELECT column_name, data_type
FROM information_schema.columns
WHERE table_schema = 'analytics'
  AND table_name = '<dest_table>'
ORDER BY ordinal_position;

-- Check for duplicates (merge mode)
SELECT id, COUNT(*)
FROM analytics.<dest_table>
GROUP BY id
HAVING COUNT(*) > 1;
-- Must return 0 rows

-- Check no _dlt_ tables exist
SELECT table_name
FROM information_schema.tables
WHERE table_schema = 'analytics'
  AND table_name LIKE '_dlt_%';
-- Must return 0 rows (invariant: no dlt tables in destination)
```
