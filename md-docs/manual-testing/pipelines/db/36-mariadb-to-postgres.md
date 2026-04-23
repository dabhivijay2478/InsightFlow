# Pipeline 36 — MariaDB Source → PostgreSQL Destination

**Source streams:** 3 | **Destination:** PostgreSQL

---

## Connections

### Source — MariaDB
```json
{ "host":"src-host","port":3306,"database":"app","username":"reader","password":"..." }
```
```sql
-- Source tables:
CREATE TABLE app.events (
    id BIGINT AUTO_INCREMENT PRIMARY KEY, user_id VARCHAR(100),
    action VARCHAR(100), payload JSON, occurred_at DATETIME(6)
) ENGINE=InnoDB;
CREATE TABLE app.logs (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    level ENUM('DEBUG','INFO','WARN','ERROR'),
    message TEXT, context JSON, logged_at DATETIME(6)
) ENGINE=InnoDB;
CREATE TABLE app.sessions (
    id VARCHAR(64) PRIMARY KEY, user_id VARCHAR(100),
    ip_address VARCHAR(45), active TINYINT(1),
    started_at DATETIME(6), ended_at DATETIME(6)
) ENGINE=InnoDB;
```

### Destination — PostgreSQL
```json
{ "host":"dest-host","port":5432,"database":"analytics","username":"writer","password":"..","ssl_mode":"disable" }
```
```sql
CREATE SCHEMA IF NOT EXISTS analytics;
```

---

## Stream 1 — `app.events` → `analytics.event_facts`

### Step 1 — DDL
```sql
CREATE TABLE analytics.event_facts (
    event_id    BIGINT      PRIMARY KEY,
    actor_id    TEXT,                         -- source: user_id (renamed)
    action      TEXT,
    client_ip   TEXT,                         -- from JSON: payload->>'ip_address'
    browser     TEXT,                         -- from JSON: payload->>'browser'
    http_status INTEGER,                      -- from JSON: payload->>'status_code' → INT
    is_success  BOOLEAN,                      -- derived: status_code 200-299
    event_date  DATE,
    event_hour  INTEGER                        -- derived: EXTRACT(hour)
);
```
### Step 3 — Panel: `app.events` | `INCREMENTAL` | key: `occurred_at`
### Step 5 — dbt SQL
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
### Step 8 — Verify
```sql
SELECT actor_id, action, client_ip, is_success, event_hour FROM analytics.event_facts LIMIT 5;
-- client_ip: plain string; is_success: true/false; payload must NOT exist
SELECT DISTINCT is_success FROM analytics.event_facts;
```

---

## Stream 2 — `app.logs` → `analytics.log_facts`

### Step 1 — DDL
```sql
CREATE TABLE analytics.log_facts (
    log_id      BIGINT      PRIMARY KEY,
    log_level   TEXT,                         -- source: level ENUM → TEXT
    severity    INTEGER,                      -- derived: ERROR=4, WARN=3, INFO=2, DEBUG=1
    log_message TEXT,                         -- source: message (renamed)
    request_id  TEXT,                         -- from JSON: context->>'request_id'
    ctx_user_id TEXT,                         -- from JSON: context->>'user_id'
    logged_on   DATE,
    log_hour    INTEGER
);
```
### Step 5 — dbt SQL
```sql
SELECT
    id AS log_id, level AS log_level,
    CASE level
        WHEN 'ERROR' THEN 4 WHEN 'WARN' THEN 3
        WHEN 'INFO'  THEN 2 ELSE 1
    END                                         AS severity,
    message                                     AS log_message,
    context->>'request_id'                      AS request_id,
    context->>'user_id'                         AS ctx_user_id,
    logged_at::DATE                             AS logged_on,
    EXTRACT(hour FROM logged_at)::INTEGER       AS log_hour
FROM {{ source('raw', 'app__logs') }}
```
### Step 8 — Verify
```sql
SELECT log_level, severity FROM analytics.log_facts GROUP BY log_level, severity;
-- ERROR→4, WARN→3, INFO→2, DEBUG→1
SELECT request_id, ctx_user_id FROM analytics.log_facts WHERE log_level='ERROR' LIMIT 5;
```

---

## Stream 3 — `app.sessions` → `analytics.session_facts`

### Step 1 — DDL
```sql
CREATE TABLE analytics.session_facts (
    session_id    TEXT        PRIMARY KEY,
    user_ref      TEXT,                       -- source: user_id (renamed)
    ip_address    TEXT,
    is_active     BOOLEAN,                    -- source: active TINYINT(1) → BOOLEAN
    duration_mins INTEGER,                    -- derived: DATEDIFF minutes
    started_on    DATE,
    ended_on      DATE
);
```
### Step 5 — dbt SQL
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

---

## Edge Cases

| Scenario | Expected |
|---------|---------|
| `payload` key missing | `client_ip` = NULL |
| `status_code` = `"abc"` non-numeric | CAST + NULLIF returns NULL; is_success = NULL |
| `active = 2` | `active = 1` → false → is_active = false |
| ENUM value not in list | MariaDB rejects at source insert; never reaches pipeline |
