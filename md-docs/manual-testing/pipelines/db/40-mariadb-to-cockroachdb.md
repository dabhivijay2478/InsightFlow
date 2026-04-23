# Pipeline 40 — MariaDB Source → CockroachDB Destination

**Source streams:** 3 | **Destination:** CockroachDB

> dbt SQL identical to `36-mariadb-to-postgres.md`.

---

## Connections
```json
{ "host":"src-host","port":3306,"database":"app","username":"reader","password":"..." }
{ "host":"dest-host","port":26257,"database":"defaultdb","username":"root","password":"","ssl_mode":"disable" }
```
```sql
CREATE SCHEMA IF NOT EXISTS analytics;
```

---

## Destination DDLs (CockroachDB)

```sql
CREATE TABLE IF NOT EXISTS analytics.event_facts (
    event_id    INT8    PRIMARY KEY,
    actor_id    STRING, action STRING,
    client_ip   STRING, browser STRING,
    http_status INT8,   is_success BOOL,
    event_date  DATE,   event_hour INT8
);
CREATE TABLE IF NOT EXISTS analytics.log_facts (
    log_id      INT8    PRIMARY KEY,
    log_level   STRING, severity INT8,
    log_message STRING,
    request_id  STRING, ctx_user_id STRING,
    logged_on   DATE,   log_hour INT8
);
CREATE TABLE IF NOT EXISTS analytics.session_facts (
    session_id    STRING  PRIMARY KEY,
    user_ref      STRING, ip_address STRING,
    is_active     BOOL,   duration_mins INT8,
    started_on    DATE,   ended_on DATE
);
```

---

## Verify
```sql
SELECT is_success FROM analytics.event_facts LIMIT 5;              -- true/false
SELECT log_level, severity FROM analytics.log_facts GROUP BY log_level, severity;
SELECT is_active FROM analytics.session_facts LIMIT 5;             -- true/false
```
