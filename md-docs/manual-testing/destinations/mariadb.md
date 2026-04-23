# MariaDB Destination — Manual Testing Guide

MariaDB is MySQL-wire-compatible. All MySQL destination scenarios apply; differences noted below.

---

## Credential Setup

```json
{
  "host": "your-mariadb-dest-host",
  "port": 3306,
  "database": "analytics",
  "username": "writer_user",
  "password": "your_password",
  "ssl_mode": "disable"
}
```

---

## Destination Schema Setup

```sql
CREATE DATABASE IF NOT EXISTS analytics;
GRANT INSERT, UPDATE, SELECT ON analytics.* TO 'writer_user'@'%';
```

---

## Scenario D-MDB-1 — Basic Delivery with PK (merge)

**Pre-create table:**
```sql
CREATE TABLE analytics.shopify_products (
    id           BIGINT PRIMARY KEY,
    title        VARCHAR(500),
    vendor       VARCHAR(255),
    product_type VARCHAR(255),
    status       VARCHAR(50),
    created_at   DATETIME
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

**Expected:** Upsert on `id`; second run count stays the same.

**Verify:**
```sql
SELECT COUNT(*) FROM analytics.shopify_products;
SELECT status, COUNT(*) FROM analytics.shopify_products GROUP BY status;
```

---

## Scenario D-MDB-2 — Long TEXT columns (MariaDB LONGTEXT)

MariaDB supports `LONGTEXT` (up to 4 GB). Use it for HubSpot notes or GitHub commit messages:

```sql
CREATE TABLE analytics.github_commits (
    sha      VARCHAR(40) PRIMARY KEY,
    message  LONGTEXT,
    author   VARCHAR(255),
    email    VARCHAR(255),
    committed_at TEXT
) ENGINE=InnoDB;
```

**Verify:** `SELECT CHAR_LENGTH(message) FROM analytics.github_commits ORDER BY 1 DESC LIMIT 5;`

---

## Scenario D-MDB-3 — JSON Column (MariaDB JSON / LONGTEXT)

MariaDB 10.2+ supports `JSON` type (stored as LONGTEXT with JSON validation):

```sql
CREATE TABLE analytics.stripe_events_raw (
    id         VARCHAR(255) PRIMARY KEY,
    event_type VARCHAR(100),
    raw_data   LONGTEXT,
    created    BIGINT
) ENGINE=InnoDB;
```

**Verify:** `SELECT JSON_EXTRACT(raw_data, '$.object') FROM analytics.stripe_events_raw LIMIT 5;`

---

## Scenario D-MDB-4 — Append Mode (no PK)

```sql
CREATE TABLE analytics.log_summary (
    level      VARCHAR(10),
    log_count  BIGINT,
    first_log  DATETIME,
    last_log   DATETIME
) ENGINE=InnoDB;
```

**Expected:** Rows grow each run. `no_pk_warnings` in callback.

---

## Scenario D-MDB-5 — Column Mismatch Error

```sql
CREATE TABLE analytics.sessions_strict (
    id          VARCHAR(128) PRIMARY KEY,
    user_id     VARCHAR(36),
    required_col VARCHAR(50) NOT NULL
) ENGINE=InnoDB;
```

**dbt SQL** (no `required_col`):
```sql
SELECT id, user_id FROM {{ source('raw', 'app__sessions') }}
```

**Expected:** Phase 0 fails.

---

## Scenario D-MDB-6 — Missing Table Error

Do NOT create `analytics.missing_table`. Set destination to it.  
**Expected:** Phase 0 fails with table-not-found message.

---

## MariaDB vs MySQL Differences to Test

| Feature | MariaDB | MySQL |
|---------|---------|-------|
| JSON type | `LONGTEXT` with JSON functions | Native `JSON` type |
| AUTO_INCREMENT | Same as MySQL | Same |
| `INSERT … ON DUPLICATE KEY UPDATE` | ✅ Supported | ✅ Supported |
| `ENUM` in WHERE | Works | Works |
| Charset default | `utf8mb4` (10.4+) | Varies by version |

---

## Destination Verification Queries

```sql
-- Row count
SELECT COUNT(*) FROM analytics.<dest_table>;

-- Schema
SELECT COLUMN_NAME, COLUMN_TYPE
FROM information_schema.COLUMNS
WHERE TABLE_SCHEMA = 'analytics'
  AND TABLE_NAME = '<dest_table>'
ORDER BY ORDINAL_POSITION;

-- Duplicate check
SELECT id, COUNT(*) cnt FROM analytics.<dest_table>
GROUP BY id HAVING cnt > 1;
-- Must return 0

-- No _dlt_ tables
SELECT TABLE_NAME FROM information_schema.TABLES
WHERE TABLE_SCHEMA = 'analytics' AND TABLE_NAME LIKE '_dlt_%';
-- Must return 0 rows
```
