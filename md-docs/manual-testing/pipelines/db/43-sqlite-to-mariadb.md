# Pipeline 43 — SQLite Source → MariaDB Destination

**Source streams:** 3 | **Destination:** MariaDB

> dbt SQL identical to `41-sqlite-to-postgres.md`. DDL same as `42-sqlite-to-mysql.md`.

---

## Connections
```json
{ "database": "/absolute/path/to/source.db" }
{ "host":"dest-host","port":3306,"database":"analytics","username":"writer","password":"..." }
```

---

## Destination DDLs (MariaDB)

```sql
CREATE TABLE analytics.task_board (
    task_id     INT          PRIMARY KEY,
    task_title  VARCHAR(500),
    task_status VARCHAR(50),
    urgency     VARCHAR(10),
    due_on      DATE,
    created_on  DATETIME(6),
    updated_on  DATETIME(6)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE analytics.task_notes (
    note_id   INT PRIMARY KEY,
    task_ref  INT,
    note_body LONGTEXT,
    added_on  DATE
) ENGINE=InnoDB;

CREATE TABLE analytics.tag_master (
    tag_id    INT PRIMARY KEY,
    tag_name  VARCHAR(255),
    hex_color VARCHAR(10)
) ENGINE=InnoDB;
```
