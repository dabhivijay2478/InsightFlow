# Pipeline 41 — SQLite Source → PostgreSQL Destination

**Source streams:** 3 | **Destination:** PostgreSQL

---

## Connections

### Source — SQLite
```json
{ "database": "/absolute/path/to/source.db" }
```
```sql
-- Source tables:
CREATE TABLE main.tasks (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT, status TEXT, priority INTEGER,
    due_date TEXT,       -- stored as ISO string 'YYYY-MM-DD'
    created_at TEXT,     -- stored as ISO string 'YYYY-MM-DDTHH:MM:SS'
    updated_at TEXT
);
CREATE TABLE main.notes (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    task_id INTEGER, content TEXT, created_at TEXT
);
CREATE TABLE main.tags (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT, color TEXT
);
```

### Destination — PostgreSQL
```json
{ "host":"dest-host","port":5432,"database":"analytics","username":"writer","password":"..","ssl_mode":"disable" }
```
```sql
CREATE SCHEMA IF NOT EXISTS analytics;
```

---

## Stream 1 — `main.tasks` → `analytics.task_board`

### Step 1 — DDL
```sql
CREATE TABLE analytics.task_board (
    task_id      INTEGER     PRIMARY KEY,
    task_title   TEXT,                        -- source: title (renamed)
    task_status  TEXT,                        -- source: status (renamed)
    urgency      TEXT,                        -- derived: CASE priority (INT→TEXT label)
    due_on       DATE,                        -- source: due_date TEXT → DATE
    created_on   TIMESTAMPTZ,                 -- source: created_at TEXT → TIMESTAMPTZ
    updated_on   TIMESTAMPTZ                  -- source: updated_at TEXT → TIMESTAMPTZ
);
```
### Step 3 — Panel: `main.tasks` | `INCREMENTAL` | key: `updated_at`
### Step 5 — dbt SQL
```sql
SELECT
    id                              AS task_id,
    title                           AS task_title,
    status                          AS task_status,
    CASE priority
        WHEN 3 THEN 'high'
        WHEN 2 THEN 'medium'
        ELSE       'low'
    END                             AS urgency,
    CAST(due_date  AS DATE)         AS due_on,
    CAST(created_at AS TIMESTAMPTZ) AS created_on,
    CAST(updated_at AS TIMESTAMPTZ) AS updated_on
FROM {{ source('raw', 'main__tasks') }}
WHERE title IS NOT NULL AND TRIM(title) != ''
```
### Step 8 — Verify
```sql
SELECT task_id, task_title, urgency, due_on, created_on FROM analytics.task_board LIMIT 5;
-- urgency: 'high'/'medium'/'low' (not 3/2/1)
-- due_on: DATE; created_on: TIMESTAMPTZ
-- 'priority' column must NOT exist
```

---

## Stream 2 — `main.notes` → `analytics.task_notes`

### Step 1 — DDL
```sql
CREATE TABLE analytics.task_notes (
    note_id    INTEGER     PRIMARY KEY,
    task_ref   INTEGER,                       -- source: task_id (renamed)
    note_body  TEXT,                          -- source: content (renamed)
    added_on   DATE                           -- source: created_at TEXT → DATE
);
```
### Step 3 — Panel: `main.notes` | `FULL_TABLE`
### Step 5 — dbt SQL
```sql
SELECT
    id       AS note_id,
    task_id  AS task_ref,
    content  AS note_body,
    CAST(created_at AS DATE) AS added_on
FROM {{ source('raw', 'main__notes') }}
WHERE content IS NOT NULL AND TRIM(content) != ''
```

---

## Stream 3 — `main.tags` → `analytics.tag_master`

### Step 1 — DDL
```sql
CREATE TABLE analytics.tag_master (
    tag_id    INTEGER PRIMARY KEY,
    tag_name  TEXT,                           -- source: name (renamed)
    hex_color TEXT                            -- source: color (normalized with # prefix)
);
```
### Step 5 — dbt SQL
```sql
SELECT
    id   AS tag_id,
    name AS tag_name,
    CASE WHEN color LIKE '#%' THEN color
         ELSE '#' || color
    END  AS hex_color
FROM {{ source('raw', 'main__tags') }}
WHERE name IS NOT NULL
```

---

## Edge Cases

| Scenario | Expected |
|---------|---------|
| SQLite TEXT date `'2024-1-5'` (no zero-padding) | `CAST` may fail → store NULL; fix source data |
| `created_at = 'now'` literal string | CAST → NULL — guard with `TRY_CAST` if DuckDB supports |
| `priority = NULL` | CASE ELSE → 'low' — acceptable default |
| Empty `content` after TRIM | Filtered by WHERE |
| Source file not mounted / path wrong | Phase 1 fails: `unable to open database file` |
