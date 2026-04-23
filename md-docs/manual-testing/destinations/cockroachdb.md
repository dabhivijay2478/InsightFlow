# CockroachDB Destination — Manual Testing Guide

CockroachDB is PostgreSQL-wire-compatible. All PostgreSQL destination test patterns apply; differences and CockroachDB-specific behaviours are noted.

---

## Credential Setup

```json
{
  "host": "your-cockroachdb-host",
  "port": 26257,
  "database": "defaultdb",
  "username": "root",
  "password": "",
  "ssl_mode": "disable"
}
```

For CockroachDB Cloud:
```json
{
  "host": "your-cluster.cockroachlabs.cloud",
  "port": 26257,
  "database": "defaultdb",
  "username": "your_user",
  "password": "your_password",
  "ssl_mode": "verify-full"
}
```

---

## Destination Schema Setup

```sql
CREATE SCHEMA IF NOT EXISTS analytics;
GRANT CREATE, INSERT, UPDATE, SELECT ON SCHEMA analytics TO your_user;
```

---

## Scenario D-CDB-1 — Basic Delivery: merge mode (with PK)

**Pre-create table:**
```sql
CREATE TABLE IF NOT EXISTS analytics.stripe_customers_hd (
    id       STRING PRIMARY KEY,
    email    STRING,
    name     STRING,
    currency STRING,
    created  INT8
);
```

> ℹ️ CockroachDB uses `STRING` (alias of `TEXT`) and `INT8` (64-bit int). Both PostgreSQL `TEXT` and `BIGINT` map correctly.

**Expected:** Upsert by `id`; no duplicates on second run.

**Verify:**
```sql
SELECT COUNT(*) FROM analytics.stripe_customers_hd;
SELECT id, COUNT(*) FROM analytics.stripe_customers_hd GROUP BY id HAVING COUNT(*) > 1;
-- Second query must return 0 rows
```

---

## Scenario D-CDB-2 — Append Mode (no PK)

```sql
CREATE TABLE IF NOT EXISTS analytics.revenue_summary (
    currency      STRING,
    charge_count  INT8,
    total_cents   INT8
);
```

**Expected:** `no_pk_warnings` in callback. Rows grow each run.

---

## Scenario D-CDB-3 — UUID Primary Key

CockroachDB natively supports UUIDs and can auto-generate them. PostgreSQL source UUIDs land directly:

```sql
CREATE TABLE IF NOT EXISTS analytics.pg_users_hd (
    id         UUID PRIMARY KEY,
    email      STRING,
    first_name STRING,
    last_name  STRING,
    created_at TIMESTAMPTZ
);
```

**dbt SQL:**
```sql
SELECT id, email, first_name, last_name, created_at
FROM {{ source('raw', 'public__users') }}
```

**Verify:**
```sql
SELECT id, pg_typeof(id) FROM analytics.pg_users_hd LIMIT 1;
-- Returns: uuid
```

---

## Scenario D-CDB-4 — JSONB Column

CockroachDB supports native `JSONB`:

```sql
CREATE TABLE IF NOT EXISTS analytics.stripe_events_raw (
    id         STRING PRIMARY KEY,
    event_type STRING,
    raw_data   JSONB,
    created    INT8
);
```

**dbt SQL:**
```sql
SELECT id, type AS event_type, data AS raw_data, created
FROM {{ source('raw', 'stripe__events') }}
```

**Verify:**
```sql
SELECT raw_data->>'object' FROM analytics.stripe_events_raw LIMIT 5;
```

---

## Scenario D-CDB-5 — Column Mismatch Error

```sql
CREATE TABLE IF NOT EXISTS analytics.orders_strict (
    id            UUID PRIMARY KEY,
    amount        DECIMAL,
    required_col  STRING NOT NULL
);
```

**dbt SQL** (no `required_col`):
```sql
SELECT id, amount FROM {{ source('raw', 'public__orders') }}
```

**Expected:** Phase 0 fails: column mismatch.

---

## Scenario D-CDB-6 — Missing Table Error

Do NOT create table. Set destination to `analytics.missing_table`.  
**Expected:** Phase 0 fails: table does not exist.

---

## Scenario D-CDB-7 — Distributed Write Performance (stress)

**Stream:** Stripe `events` (high volume)  
**Sync mode:** `FULL_TABLE`

CockroachDB distributes writes across nodes. Verify:
- Run completes without timeout
- `rows_written` matches expected count
- No transaction retry errors in ELT server logs

---

## Scenario D-CDB-8 — Multi-Region Latency

If using CockroachDB Cloud multi-region, the destination host is a load balancer. Delivery latency may be higher than single-node.

**Expected:** Run still completes; `staging_size_bytes` and `delivered_tables` populated in callback.

---

## CockroachDB vs PostgreSQL Differences

| Feature | CockroachDB | PostgreSQL |
|---------|-------------|-----------|
| Default PK type | UUID (auto) | SERIAL or explicit |
| `JSONB` | ✅ Native | ✅ Native |
| `pg_typeof()` | ✅ Supported | ✅ Supported |
| `information_schema` | ✅ Supported | ✅ Supported |
| `UPSERT` syntax | ✅ (alias of `INSERT ON CONFLICT DO UPDATE`) | Via `ON CONFLICT DO UPDATE` |
| SSL default | Strongly recommended | Optional |
| Schema creation | `CREATE SCHEMA` | Same |

---

## Destination Verification Queries

```sql
-- Row count
SELECT COUNT(*) FROM analytics.<dest_table>;

-- Column list
SELECT column_name, data_type
FROM information_schema.columns
WHERE table_schema = 'analytics'
  AND table_name = '<dest_table>'
ORDER BY ordinal_position;

-- Duplicate check
SELECT id, COUNT(*)
FROM analytics.<dest_table>
GROUP BY id
HAVING COUNT(*) > 1;
-- Must return 0 rows

-- No _dlt_ tables
SELECT table_name
FROM information_schema.tables
WHERE table_schema = 'analytics'
  AND table_name LIKE '_dlt_%';
-- Must return 0 rows

-- Verify JSONB column
SELECT jsonb_typeof(raw_data) FROM analytics.<dest_table> LIMIT 1;
```
