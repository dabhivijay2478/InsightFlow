# Pipeline 42 — SQLite Source → MySQL Destination

**Source streams:** 3 | **Destination:** MySQL

> dbt SQL identical to `41-sqlite-to-postgres.md`.

---

## Connections
```json
{ "database": "/absolute/path/to/source.db" }
{ "host":"dest-host","port":3306,"database":"analytics","username":"writer","password":"..." }
```

---

## Destination DDLs (MySQL)

```sql
CREATE TABLE analytics.task_board (
    task_id     INT          PRIMARY KEY,
    task_title  VARCHAR(500),
    task_status VARCHAR(50),
    urgency     VARCHAR(10),
    due_on      DATE,
    created_on  DATETIME,
    updated_on  DATETIME
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE analytics.task_notes (
    note_id   INT PRIMARY KEY,
    task_ref  INT,
    note_body TEXT,
    added_on  DATE
) ENGINE=InnoDB;

CREATE TABLE analytics.tag_master (
    tag_id    INT PRIMARY KEY,
    tag_name  VARCHAR(255),
    hex_color VARCHAR(10)
) ENGINE=InnoDB;
```

---

## Verify
```sql
SELECT urgency FROM analytics.task_board GROUP BY urgency;      -- high/medium/low
SELECT created_on FROM analytics.task_board LIMIT 3;            -- DATETIME
SELECT hex_color FROM analytics.tag_master LIMIT 5;             -- starts with #
```
