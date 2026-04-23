# SQLite Source — Manual Testing Guide

**Streams:** 3 (`main.tasks`, `main.notes`, `main.tags`)  
**Credential format:** file path to `.db` file  
**DuckDB naming:** `main__tasks`, `main__notes`, `main__tags`

---

## Credential Setup

SQLite does not use a network connection. The ELT server must have direct filesystem access to the `.db` file:

```json
{
  "database": "/absolute/path/to/your/database.db"
}
```

> ⚠️ The ELT server process must have read permissions on the file path. Relative paths are not supported.

---

## Required Source Tables

```sql
-- Create and seed the SQLite database file:
-- sqlite3 /path/to/database.db

CREATE TABLE IF NOT EXISTS main.tasks (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    title       TEXT NOT NULL,
    description TEXT,
    status      TEXT DEFAULT 'todo' CHECK(status IN ('todo','in_progress','done')),
    priority    INTEGER DEFAULT 1,
    due_date    TEXT,
    created_at  TEXT DEFAULT (datetime('now')),
    updated_at  TEXT DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS main.notes (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    task_id     INTEGER REFERENCES main.tasks(id),
    content     TEXT NOT NULL,
    created_at  TEXT DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS main.tags (
    id    INTEGER PRIMARY KEY AUTOINCREMENT,
    name  TEXT NOT NULL UNIQUE,
    color TEXT DEFAULT '#808080'
);

-- Seed data
INSERT INTO main.tasks (title, status, priority) VALUES
    ('Deploy to production', 'in_progress', 3),
    ('Write unit tests', 'todo', 2),
    ('Review PR', 'done', 2);

INSERT INTO main.notes (task_id, content) VALUES
    (1, 'Check rollback plan first'),
    (1, 'Update deployment docs after');

INSERT INTO main.tags (name, color) VALUES
    ('backend', '#0000ff'),
    ('frontend', '#ff0000'),
    ('urgent', '#ff8800');
```

---

## Stream Reference

| Stream key | DuckDB staging name | Key columns | INCREMENTAL key |
|-----------|-------------------|-------------|----------------|
| `main.tasks` | `main__tasks` | `id`, `title`, `status`, `priority`, `due_date`, `created_at`, `updated_at` | `updated_at` |
| `main.notes` | `main__notes` | `id`, `task_id`, `content`, `created_at` | `created_at` |
| `main.tags` | `main__tags` | `id`, `name`, `color` | — |

> ℹ️ SQLite stores all dates as TEXT. The dbt SQL model (running in DuckDB) can cast with `CAST(created_at AS TIMESTAMPTZ)`.

---

## Scenario S-SQ-1 — Full Table Sync: `main.tasks`

**Destination DDL:**
```sql
CREATE TABLE analytics.sqlite_tasks_hd (
    id          INTEGER PRIMARY KEY,
    title       TEXT,
    status      TEXT,
    priority    INTEGER,
    due_date    TEXT,
    created_at  TIMESTAMPTZ
);
```

**dbt SQL:**
```sql
SELECT
    id,
    title,
    status,
    priority,
    due_date,
    CAST(created_at AS TIMESTAMPTZ)  AS created_at
FROM {{ source('raw', 'main__tasks') }}
```

**Verify:** `SELECT status, COUNT(*) FROM analytics.sqlite_tasks_hd GROUP BY status;`

---

## Scenario S-SQ-2 — Tasks with Notes JOIN

Both `main.tasks` AND `main.notes` in Source Panel:

**dbt SQL:**
```sql
SELECT
    t.id          AS task_id,
    t.title,
    t.status,
    n.id          AS note_id,
    n.content     AS note_content,
    CAST(n.created_at AS TIMESTAMPTZ)  AS note_created_at
FROM {{ source('raw', 'main__tasks') }} AS t
LEFT JOIN {{ source('raw', 'main__notes') }} AS n
    ON t.id = n.task_id
```

---

## Scenario S-SQ-3 — Incremental Sync: `main.tasks`

**Sync mode:** `INCREMENTAL`, replication key `updated_at`

> ⚠️ SQLite stores `updated_at` as TEXT (`"2024-01-15 10:30:00"`). Ensure the ELT cursor comparison works with ISO 8601 string comparison.

**dbt SQL:**
```sql
SELECT
    id,
    title,
    status,
    priority,
    CAST(updated_at AS TIMESTAMPTZ)  AS updated_at
FROM {{ source('raw', 'main__tasks') }}
```

**Run 1:** All tasks.  
**Run 2 (update a task status):** Only the updated task.

---

## Scenario S-SQ-4 — Task Priority Grouping

**dbt SQL:**
```sql
SELECT
    status,
    priority,
    COUNT(*)                                  AS task_count,
    CAST(MAX(updated_at) AS TIMESTAMPTZ)     AS last_updated
FROM {{ source('raw', 'main__tasks') }}
GROUP BY status, priority
ORDER BY priority DESC, status
```

---

## Scenario S-SQ-5 — Tags Reference Table (no PK warning expected)

`main.tags` has an `id` INTEGER PK, so this will use merge mode.

**dbt SQL:**
```sql
SELECT id, name, color
FROM {{ source('raw', 'main__tags') }}
ORDER BY name
```

**Expected:** Merge mode (PK = `id`). No `no_pk_warnings`.

---

## Scenario S-SQ-6 — Normalisation: Cast TEXT dates

**Rule:**
```json
{ "rule_type": "cast", "table": "main.tasks", "column": "created_at", "cast_to": "timestamptz" }
```

**dbt SQL:**
```sql
SELECT id, title, status, created_at FROM {{ source('raw', 'main__tasks') }}
```

---

## Scenario S-SQ-7 — Error: File path not accessible

Set `database = "/nonexistent/path/app.db"`.  
**Expected:** Phase 1 extraction fails with a file-not-found error. Run status: `failed`.

---

## All 3 Streams — Quick Smoke Test Checklist

| Stream | DuckDB ref | Expected rows |
|--------|-----------|---------------|
| tasks | `main__tasks` | ≥ 1 |
| notes | `main__notes` | ≥ 1 |
| tags | `main__tags` | ≥ 1 |
