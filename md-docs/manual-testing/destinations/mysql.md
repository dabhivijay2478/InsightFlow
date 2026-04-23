# MySQL Destination — Manual Testing Guide

MySQL supports merge (upsert via `INSERT … ON DUPLICATE KEY UPDATE`) and append write modes.

---

## Credential Setup

```json
{
  "host": "your-mysql-dest-host",
  "port": 3306,
  "database": "analytics",
  "username": "writer_user",
  "password": "your_password",
  "ssl_mode": "disable"
}
```

Destination user needs `INSERT`, `UPDATE`, `SELECT` on the target schema. No `CREATE TABLE` required.

---

## Destination Schema Setup

```sql
CREATE DATABASE IF NOT EXISTS analytics;
GRANT INSERT, UPDATE, SELECT ON analytics.* TO 'writer_user'@'%';
FLUSH PRIVILEGES;
```

---

## Scenario D-MY-1 — Basic Delivery: merge mode (with PK)

**Pre-create destination table:**
```sql
CREATE TABLE analytics.stripe_customers (
    id       VARCHAR(255) PRIMARY KEY,
    email    VARCHAR(255),
    name     VARCHAR(255),
    currency VARCHAR(10),
    created  BIGINT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

**Expected:** Upsert on `id`; second run does not duplicate rows.

**Verify:**
```sql
SELECT COUNT(*) FROM analytics.stripe_customers;
-- Must not grow on second run (same data)
```

---

## Scenario D-MY-2 — Append Mode (no PK table)

**Pre-create table without PK:**
```sql
CREATE TABLE analytics.charges_log (
    currency     VARCHAR(10),
    charge_count BIGINT,
    total_cents  BIGINT
) ENGINE=InnoDB;
```

**Expected:** `no_pk_warnings` in callback; rows appended on each run.

---

## Scenario D-MY-3 — JSON Column Destination

MySQL 5.7+ and 8.0 support native JSON columns.

**Pre-create table:**
```sql
CREATE TABLE analytics.stripe_events_raw (
    id         VARCHAR(255) PRIMARY KEY,
    event_type VARCHAR(100),
    raw_data   JSON,
    created    BIGINT
) ENGINE=InnoDB;
```

**dbt SQL:**
```sql
SELECT id, type AS event_type, data AS raw_data, created
FROM {{ source('raw', 'stripe__events') }}
```

**Verify:**
```sql
SELECT JSON_EXTRACT(raw_data, '$.object') FROM analytics.stripe_events_raw LIMIT 5;
```

---

## Scenario D-MY-4 — Column Mismatch Error

**Pre-create table with `NOT NULL` column absent from dbt output:**
```sql
CREATE TABLE analytics.orders_strict (
    id       VARCHAR(36) PRIMARY KEY,
    amount   DECIMAL(10,2),
    required_col TEXT NOT NULL
) ENGINE=InnoDB;
```

**dbt SQL** (no `required_col`):
```sql
SELECT id, amount FROM {{ source('raw', 'public__orders') }}
```

**Expected:** Phase 0 fails with column mismatch error.

---

## Scenario D-MY-5 — Missing Destination Table Error

Do NOT create destination table. Set destination to `analytics.missing_table`.  
**Expected:** Phase 0 fails: table does not exist.

---

## Scenario D-MY-6 — TEXT vs VARCHAR Compatibility

MySQL `VARCHAR(255)` may truncate long text from source. Test with HubSpot `notes.hs_note_body` (can be very long):

**Pre-create table (use TEXT to be safe):**
```sql
CREATE TABLE analytics.hubspot_notes (
    id          VARCHAR(255) PRIMARY KEY,
    note_body   TEXT,
    created_at  DATETIME
) ENGINE=InnoDB;
```

**Verify:** `SELECT LENGTH(note_body) FROM analytics.hubspot_notes ORDER BY LENGTH(note_body) DESC LIMIT 5;`

---

## Destination Verification Queries

```sql
-- Row count
SELECT COUNT(*) FROM analytics.<dest_table>;

-- Column list
SELECT COLUMN_NAME, DATA_TYPE, CHARACTER_MAXIMUM_LENGTH
FROM information_schema.COLUMNS
WHERE TABLE_SCHEMA = 'analytics'
  AND TABLE_NAME = '<dest_table>'
ORDER BY ORDINAL_POSITION;

-- Duplicate check (merge mode)
SELECT id, COUNT(*) AS cnt
FROM analytics.<dest_table>
GROUP BY id
HAVING cnt > 1;
-- Must return 0 rows

-- No dlt_ tables
SELECT TABLE_NAME FROM information_schema.TABLES
WHERE TABLE_SCHEMA = 'analytics'
  AND TABLE_NAME LIKE '_dlt_%';
-- Must return 0 rows
```
