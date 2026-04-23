# MariaDB Source — UI Testing (3 Streams × 5 Destinations)

> Universal builder steps in `builder-walkthrough.md`.

---

## Phase 1 — Source Panel (MariaDB)

### Credential Fields
Identical to MySQL (MariaDB uses MySQL protocol):

| Field | Value |
|-------|-------|
| **Host** | `src-host` |
| **Port** | `3306` |
| **Database** | `app` |
| **Username** | `reader` |
| **Password** | `***` |

**Test Connection → ✅**

### MariaDB vs MySQL UI Difference
- In the source type dropdown, select **MariaDB** (not MySQL) — uses MariaDB-specific driver
- Test: select MySQL type for a MariaDB server → verify connection still works (wire-compatible) but note this as a gap if it doesn't

---

## Step 2b — Stream Selection & Sync Mode

Source tables:
```sql
CREATE TABLE app.events (id BIGINT AUTO_INCREMENT PRIMARY KEY, user_id VARCHAR(100), action VARCHAR(100), payload JSON, occurred_at DATETIME(6)) ENGINE=InnoDB;
CREATE TABLE app.logs (id BIGINT AUTO_INCREMENT PRIMARY KEY, level ENUM('DEBUG','INFO','WARN','ERROR'), message TEXT, context JSON, logged_at DATETIME(6)) ENGINE=InnoDB;
CREATE TABLE app.sessions (id VARCHAR(64) PRIMARY KEY, user_id VARCHAR(100), ip_address VARCHAR(45), active TINYINT(1), started_at DATETIME(6), ended_at DATETIME(6)) ENGINE=InnoDB;
```

| Stream | Sync Mode | Cursor Field |
|--------|-----------|-------------|
| `app.events` | INCREMENTAL | `occurred_at` |
| `app.logs` | INCREMENTAL | `logged_at` |
| `app.sessions` | INCREMENTAL | `started_at` |

**Cursor field test**: `DATETIME(6)` columns must appear in cursor dropdown — verify microsecond precision preserved.

---

## Phase 2 — Stream→Table Mapping

| Stream | Table |
|--------|-------|
| `app.events` | `event_facts` |
| `app.logs` | `log_facts` |
| `app.sessions` | `session_facts` |

---

## Phase 3 — Normalisation Rules

### `app.events`
| Rule | Column | Target |
|------|--------|--------|
| Rename | `id` | `event_id` |
| Rename | `user_id` | `actor_id` |
| Exclude | `payload` | — (JSON keys extracted in dbt) |

### `app.logs`
| Rule | Column | Target |
|------|--------|--------|
| Rename | `id` | `log_id` |
| Rename | `level` | `log_level` |
| Rename | `message` | `log_message` |
| Exclude | `context` | — (JSON keys extracted in dbt) |

### `app.sessions`
| Rule | Column | Target |
|------|--------|--------|
| Rename | `id` | `session_id` |
| Rename | `user_id` | `user_ref` |
| Cast | `active` | Boolean |

---

## Phase 4 — dbt SQL

### Stream 1 — `app.events` → `event_facts`
```sql
SELECT
    id AS event_id, user_id AS actor_id, action,
    payload->>'ip_address'                                       AS client_ip,
    payload->>'browser'                                          AS browser,
    CAST(NULLIF(payload->>'status_code','') AS INTEGER)          AS http_status,
    CAST(NULLIF(payload->>'status_code','') AS INTEGER) BETWEEN 200 AND 299 AS is_success,
    occurred_at::DATE                                            AS event_date,
    EXTRACT(hour FROM occurred_at)::INTEGER                      AS event_hour
FROM {{ source('raw', 'app__events') }}
WHERE payload IS NOT NULL
```
**Preview check**: `client_ip` plain string; `is_success` boolean; `payload` column absent; `event_hour` 0–23

---

### Stream 2 — `app.logs` → `log_facts`
```sql
SELECT
    id AS log_id, level AS log_level,
    CASE level
        WHEN 'ERROR' THEN 4
        WHEN 'WARN'  THEN 3
        WHEN 'INFO'  THEN 2
        ELSE              1
    END                                         AS severity,
    message                                     AS log_message,
    context->>'request_id'                      AS request_id,
    context->>'user_id'                         AS ctx_user_id,
    logged_at::DATE                             AS logged_on,
    EXTRACT(hour FROM logged_at)::INTEGER       AS log_hour
FROM {{ source('raw', 'app__logs') }}
```
**Preview check**: `severity` = 1/2/3/4 (integer); `log_level` = ENUM text; `context` column absent

---

### Stream 3 — `app.sessions` → `session_facts`
```sql
SELECT
    id AS session_id, user_id AS user_ref, ip_address,
    active = 1 AS is_active,
    CASE WHEN ended_at IS NOT NULL
         THEN DATEDIFF('minute', started_at, ended_at)
         ELSE NULL END AS duration_mins,
    started_at::DATE AS started_on,
    ended_at::DATE   AS ended_on
FROM {{ source('raw', 'app__sessions') }}
```
**Preview check**: `is_active` boolean; `duration_mins` integer or NULL

---

## Phase 5 — Preview Checks

| Stream | Column | Expected |
|--------|--------|---------|
| events | `client_ip` | `"192.168.1.1"` |
| events | `is_success` | `true`/`false` |
| logs | `severity` | Integer 1–4 |
| logs | `log_level` | `"ERROR"` / `"INFO"` / etc. |
| sessions | `duration_mins` | Integer or NULL |

---

## Phase 6 — Destination Type Verification

| Destination | `is_active` | `severity` | `client_ip` |
|-------------|------------|-----------|------------|
| PostgreSQL | `BOOLEAN` | `INTEGER` | `TEXT` |
| MySQL | `TINYINT(1)` | `INT` | `VARCHAR(45)` |
| MariaDB | `TINYINT(1)` | `INT` | `VARCHAR(45)` |
| SQLite | `INTEGER` | `INTEGER` | `TEXT` |
| CockroachDB | `BOOL` | `INT8` | `STRING` |

---

## Phase 7 — Failure Scenarios

| Scenario | Expected |
|---------|---------|
| `payload = NULL` | WHERE filter — row excluded |
| `status_code = 'abc'` | CAST + NULLIF → NULL; `is_success` = NULL |
| `active = 2` | `active = 1` → false |
| ENUM value out of range | MariaDB rejects at source — never reaches pipeline |
| `DATETIME(6)` precision: microseconds | Stored in DuckDB as TIMESTAMPTZ microsecond precision — ✅ |

---

## Phase 8 — Verify

```sql
SELECT log_level, severity FROM analytics.log_facts GROUP BY log_level, severity;
-- ERROR→4, WARN→3, INFO→2, DEBUG→1
SELECT is_success FROM analytics.event_facts GROUP BY is_success;    -- true/false
SELECT event_id, COUNT(*) FROM analytics.event_facts GROUP BY event_id HAVING COUNT(*)>1; -- 0
```
