# Pipeline 24 — Notion → SQLite

**Streams:** 3 | **Destination:** SQLite

> dbt SQL identical to `21-notion-to-postgres.md`.

---

## Connections
```json
{ "credentials": { "auth_type": "token", "token": "secret_..." } }
{ "database": "/absolute/path/to/analytics.db" }
```

---

## All 3 Stream DDLs (SQLite)

```sql
CREATE TABLE IF NOT EXISTS notion_databases (
    db_id TEXT PRIMARY KEY, db_title TEXT, db_url TEXT,
    created_on TEXT, modified_on TEXT
);
CREATE TABLE IF NOT EXISTS notion_pages (
    page_id TEXT PRIMARY KEY, page_url TEXT,
    status TEXT, priority TEXT, assignee TEXT,
    due_date TEXT, created_on TEXT, modified_on TEXT
);
CREATE TABLE IF NOT EXISTS notion_users (
    user_id TEXT PRIMARY KEY, display_name TEXT, user_type TEXT,
    email TEXT, avatar_url TEXT
);
```

---

## Verify
```bash
DB=/absolute/path/to/analytics.db
sqlite3 $DB "SELECT db_title FROM notion_databases LIMIT 5;"
sqlite3 $DB "SELECT status, due_date FROM notion_pages LIMIT 5;"
sqlite3 $DB "SELECT email FROM notion_users LIMIT 5;"
```
