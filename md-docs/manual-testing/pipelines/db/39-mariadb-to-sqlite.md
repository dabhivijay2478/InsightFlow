# Pipeline 39 — MariaDB Source → SQLite Destination

**Source streams:** 3 | **Destination:** SQLite

> dbt SQL identical to `36-mariadb-to-postgres.md`.

---

## Connections
```json
{ "host":"src-host","port":3306,"database":"app","username":"reader","password":"..." }
{ "database": "/absolute/path/to/analytics.db" }
```

---

## Destination DDLs (SQLite)

```sql
CREATE TABLE IF NOT EXISTS event_facts (
    event_id INTEGER PRIMARY KEY, actor_id TEXT, action TEXT,
    client_ip TEXT, browser TEXT, http_status INTEGER,
    is_success INTEGER, event_date TEXT, event_hour INTEGER
);
CREATE TABLE IF NOT EXISTS log_facts (
    log_id INTEGER PRIMARY KEY, log_level TEXT, severity INTEGER,
    log_message TEXT, request_id TEXT, ctx_user_id TEXT,
    logged_on TEXT, log_hour INTEGER
);
CREATE TABLE IF NOT EXISTS session_facts (
    session_id TEXT PRIMARY KEY, user_ref TEXT, ip_address TEXT,
    is_active INTEGER, duration_mins INTEGER,
    started_on TEXT, ended_on TEXT
);
```

---

## Verify
```bash
DB=/absolute/path/to/analytics.db
sqlite3 $DB "SELECT log_level, severity FROM log_facts GROUP BY log_level, severity;"
sqlite3 $DB "SELECT is_success FROM event_facts LIMIT 5;"  # 0 or 1
```
