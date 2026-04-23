# Pipeline 38 — MariaDB Source → MariaDB Destination

**Source streams:** 3 | **Destination:** MariaDB (different host/db)

> dbt SQL identical to `36-mariadb-to-postgres.md`. DDL same as `37-mariadb-to-mysql.md`.

---

## Connections
```json
{ "host":"src-host","port":3306,"database":"app","username":"reader","password":"..." }
{ "host":"dest-host","port":3306,"database":"analytics","username":"writer","password":"..." }
```

---

## Destination DDLs (MariaDB)

```sql
CREATE TABLE analytics.event_facts (
    event_id   BIGINT  PRIMARY KEY,
    actor_id   VARCHAR(100), action VARCHAR(100),
    client_ip  VARCHAR(45), browser VARCHAR(255),
    http_status INT, is_success TINYINT(1),
    event_date DATE, event_hour INT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE analytics.log_facts (
    log_id     BIGINT  PRIMARY KEY,
    log_level  VARCHAR(10), severity INT,
    log_message LONGTEXT,
    request_id VARCHAR(255), ctx_user_id VARCHAR(100),
    logged_on DATE, log_hour INT
) ENGINE=InnoDB;

CREATE TABLE analytics.session_facts (
    session_id    VARCHAR(64) PRIMARY KEY,
    user_ref      VARCHAR(100), ip_address VARCHAR(45),
    is_active     TINYINT(1), duration_mins INT,
    started_on    DATE, ended_on DATE
) ENGINE=InnoDB;
```
