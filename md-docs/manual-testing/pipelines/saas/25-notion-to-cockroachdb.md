# Pipeline 25 — Notion → CockroachDB

**Streams:** 3 | **Destination:** CockroachDB

> dbt SQL identical to `21-notion-to-postgres.md`.

---

## Connections
```json
{ "credentials": { "auth_type": "token", "token": "secret_..." } }
{ "host":"..","port":26257,"database":"defaultdb","username":"root","password":"","ssl_mode":"disable" }
```
```sql
CREATE SCHEMA IF NOT EXISTS analytics;
```

---

## All 3 Stream DDLs (CockroachDB)

```sql
CREATE TABLE IF NOT EXISTS analytics.notion_databases (
    db_id STRING PRIMARY KEY, db_title STRING, db_url STRING,
    created_on DATE, modified_on DATE
);
CREATE TABLE IF NOT EXISTS analytics.notion_pages (
    page_id STRING PRIMARY KEY, page_url STRING,
    status STRING, priority STRING, assignee STRING,
    due_date DATE, created_on DATE, modified_on DATE
);
CREATE TABLE IF NOT EXISTS analytics.notion_users (
    user_id STRING PRIMARY KEY, display_name STRING, user_type STRING,
    email STRING, avatar_url STRING
);
```

---

## Verify
```sql
SELECT db_title FROM analytics.notion_databases LIMIT 5;
SELECT status, due_date FROM analytics.notion_pages LIMIT 5;
SELECT db_id, COUNT(*) FROM analytics.notion_databases GROUP BY db_id HAVING COUNT(*)>1;
```
