# Pipeline 45 — SQLite Source → CockroachDB Destination

**Source streams:** 3 | **Destination:** CockroachDB

> dbt SQL identical to `41-sqlite-to-postgres.md`.

---

## Connections
```json
{ "database": "/absolute/path/to/source.db" }
{ "host":"dest-host","port":26257,"database":"defaultdb","username":"root","password":"","ssl_mode":"disable" }
```
```sql
CREATE SCHEMA IF NOT EXISTS analytics;
```

---

## Destination DDLs (CockroachDB)

```sql
CREATE TABLE IF NOT EXISTS analytics.task_board (
    task_id    INT8    PRIMARY KEY,
    task_title STRING, task_status STRING, urgency STRING,
    due_on     DATE,   created_on TIMESTAMPTZ, updated_on TIMESTAMPTZ
);
CREATE TABLE IF NOT EXISTS analytics.task_notes (
    note_id  INT8 PRIMARY KEY,
    task_ref INT8, note_body STRING, added_on DATE
);
CREATE TABLE IF NOT EXISTS analytics.tag_master (
    tag_id INT8 PRIMARY KEY, tag_name STRING, hex_color STRING
);
```

---

## Verify
```sql
SELECT urgency FROM analytics.task_board GROUP BY urgency;        -- high/medium/low
SELECT created_on FROM analytics.task_board LIMIT 3;              -- TIMESTAMPTZ
SELECT hex_color FROM analytics.tag_master WHERE hex_color NOT LIKE '#%';
-- Must return 0 rows
```
