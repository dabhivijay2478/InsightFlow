# MariaDB Source — Manual Testing Guide

**Streams:** 3 (`app.events`, `app.logs`, `app.sessions`)  
**Credential format:** Same as MySQL (MariaDB is protocol-compatible)  
**DuckDB naming:** `app__events`, `app__logs`, `app__sessions`

---

## Credential Setup

```json
{
  "host": "your-mariadb-host",
  "port": 3306,
  "database": "app",
  "username": "your_user",
  "password": "your_password",
  "ssl_mode": "disable"
}
```

---

## Required Source Tables

```sql
-- app.events
CREATE TABLE IF NOT EXISTS app.events (
    id          BIGINT AUTO_INCREMENT PRIMARY KEY,
    user_id     VARCHAR(36),
    action      VARCHAR(100) NOT NULL,
    payload     JSON,
    occurred_at DATETIME DEFAULT NOW()
);

-- app.logs
CREATE TABLE IF NOT EXISTS app.logs (
    id          BIGINT AUTO_INCREMENT PRIMARY KEY,
    level       ENUM('DEBUG','INFO','WARN','ERROR') NOT NULL,
    message     TEXT NOT NULL,
    context     JSON,
    logged_at   DATETIME DEFAULT NOW()
);

-- app.sessions
CREATE TABLE IF NOT EXISTS app.sessions (
    id          VARCHAR(128) PRIMARY KEY,
    user_id     VARCHAR(36),
    ip_address  VARCHAR(45),
    user_agent  TEXT,
    started_at  DATETIME DEFAULT NOW(),
    ended_at    DATETIME,
    active      TINYINT(1) DEFAULT 1
);
```

Seed with at least 20 rows per table for meaningful tests.

---

## Stream Reference

| Stream key | DuckDB staging name | Key columns | INCREMENTAL key |
|-----------|-------------------|-------------|----------------|
| `app.events` | `app__events` | `id`, `user_id`, `action`, `payload`, `occurred_at` | `occurred_at` |
| `app.logs` | `app__logs` | `id`, `level`, `message`, `context`, `logged_at` | `logged_at` |
| `app.sessions` | `app__sessions` | `id`, `user_id`, `ip_address`, `started_at`, `ended_at`, `active` | `started_at` |

---

## Scenario S-MDB-1 — Full Table Sync: `app.events`

**Destination DDL:**
```sql
CREATE TABLE analytics.app_events_hd (
    id          BIGINT PRIMARY KEY,
    user_id     TEXT,
    action      TEXT,
    occurred_at TIMESTAMPTZ
);
```

**dbt SQL:**
```sql
SELECT
    id,
    user_id,
    action,
    occurred_at
FROM {{ source('raw', 'app__events') }}
```

---

## Scenario S-MDB-2 — JSON Key Filtering from `events.payload`

`app.events.payload` is a JSON object with 10 keys. Extract only `ip_address`, `browser`, `status_code`:

**dbt SQL:**
```sql
SELECT
    id,
    user_id,
    action,
    JSON_UNQUOTE(JSON_EXTRACT(payload, '$.ip_address'))  AS ip_address,
    JSON_UNQUOTE(JSON_EXTRACT(payload, '$.browser'))     AS browser,
    CAST(JSON_EXTRACT(payload, '$.status_code') AS UNSIGNED) AS status_code,
    occurred_at
FROM {{ source('raw', 'app__events') }}
WHERE JSON_EXTRACT(payload, '$.status_code') IS NOT NULL
```

> ℹ️ MariaDB uses `JSON_EXTRACT` / `JSON_UNQUOTE` instead of PostgreSQL `->>'key'` syntax.  
> dbt runs against DuckDB, which supports `->>'key'` syntax regardless of source DB.

**dbt SQL (DuckDB syntax inside the SQL model — use `->>'key'`):**
```sql
SELECT
    id,
    user_id,
    action,
    payload->>'ip_address'              AS ip_address,
    payload->>'browser'                 AS browser,
    CAST(payload->>'status_code' AS INTEGER) AS status_code,
    occurred_at
FROM {{ source('raw', 'app__events') }}
WHERE payload->>'status_code' IS NOT NULL
```

**Destination DDL:**
```sql
CREATE TABLE analytics.events_payload_flat (
    id          BIGINT PRIMARY KEY,
    user_id     TEXT,
    action      TEXT,
    ip_address  TEXT,
    browser     TEXT,
    status_code INTEGER,
    occurred_at TIMESTAMPTZ
);
```

---

## Scenario S-MDB-3 — Log Level Aggregate

**dbt SQL:**
```sql
SELECT
    level,
    COUNT(*)                              AS log_count,
    MIN(logged_at)                        AS first_log,
    MAX(logged_at)                        AS last_log
FROM {{ source('raw', 'app__logs') }}
GROUP BY level
```

**Destination DDL (no PK — append):**
```sql
CREATE TABLE analytics.log_summary (
    level      TEXT,
    log_count  BIGINT,
    first_log  TIMESTAMPTZ,
    last_log   TIMESTAMPTZ
);
```

**Expected:** `no_pk_warnings` in callback.

---

## Scenario S-MDB-4 — Incremental Sync: `app.sessions`

**Sync mode:** `INCREMENTAL`, replication key `started_at`

**dbt SQL:**
```sql
SELECT
    id,
    user_id,
    ip_address,
    started_at,
    ended_at,
    active = 1  AS is_active
FROM {{ source('raw', 'app__sessions') }}
```

---

## Scenario S-MDB-5 — Session Duration Calculation

**dbt SQL:**
```sql
SELECT
    id,
    user_id,
    ip_address,
    started_at,
    ended_at,
    CASE
        WHEN ended_at IS NOT NULL
        THEN DATEDIFF('second', started_at, ended_at)
        ELSE NULL
    END AS duration_seconds
FROM {{ source('raw', 'app__sessions') }}
WHERE active = 0
  AND ended_at IS NOT NULL
```

---

## Scenario S-MDB-6 — Error Log Context JSON Filtering

`app.logs.context` is a JSON object. Extract `request_id` and `user_id` for ERROR-level logs:

**dbt SQL:**
```sql
SELECT
    id,
    level,
    message,
    context->>'request_id'   AS request_id,
    context->>'user_id'      AS user_id,
    logged_at
FROM {{ source('raw', 'app__logs') }}
WHERE level = 'ERROR'
  AND context IS NOT NULL
```

---

## Scenario S-MDB-7 — Multi-Stream: All 3 MariaDB Tables

**Model 1:**
```sql
SELECT id, user_id, action, occurred_at FROM {{ source('raw', 'app__events') }}
```

**Model 2:**
```sql
SELECT id, level, message, logged_at FROM {{ source('raw', 'app__logs') }}
```

**Model 3:**
```sql
SELECT id, user_id, ip_address, started_at, active FROM {{ source('raw', 'app__sessions') }}
```

---

## Scenario S-MDB-8 — Normalisation: Rename ENUM → text label

**Rule:**
```json
{ "rule_type": "rename", "table": "app.logs", "column": "level", "destination_name": "log_level" }
```

**dbt SQL:**
```sql
SELECT id, log_level, message, logged_at FROM {{ source('raw', 'app__logs') }}
```
