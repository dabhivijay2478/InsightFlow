# SQLite Source — UI Testing (3 Streams × 5 Destinations)

> Universal builder steps in `builder-walkthrough.md`.

---

## Phase 1 — Source Panel (SQLite)

### Credential Fields
| Field | Value |
|-------|-------|
| **Database path** | `/absolute/path/to/source.db` |

**Test Connection → ✅**
- ❌ `unable to open database file`: path wrong or file doesn't exist
- ❌ `permission denied`: file not readable by the ELT server process
- ❌ `database is locked`: another process has exclusive lock

### Path Test Cases
| Path entered | Expected |
|-------------|---------|
| `/correct/path/source.db` | ✅ |
| `./relative/path.db` | ❌ relative path rejected or fails |
| `/nonexistent/file.db` | ❌ `no such file` |
| `/read-only/file.db` (read-only) | ✅ source read is OK; only write needs permission |

---

## Step 2b — Stream Selection & Sync Mode

Source tables (pre-create in SQLite):
```sql
CREATE TABLE IF NOT EXISTS main.tasks (id INTEGER PRIMARY KEY AUTOINCREMENT, title TEXT, status TEXT, priority INTEGER, due_date TEXT, created_at TEXT, updated_at TEXT);
CREATE TABLE IF NOT EXISTS main.notes (id INTEGER PRIMARY KEY AUTOINCREMENT, task_id INTEGER, content TEXT, created_at TEXT);
CREATE TABLE IF NOT EXISTS main.tags (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT, color TEXT);
```

| Stream | Sync Mode | Cursor Field |
|--------|-----------|-------------|
| `main.tasks` | INCREMENTAL | `updated_at` |
| `main.notes` | FULL TABLE | — |
| `main.tags` | FULL TABLE | — |

**Cursor field note**: `updated_at` is stored as TEXT in SQLite. UI must show it in cursor dropdown — verify the ELT server accepts TEXT columns as incremental cursors (lexicographic comparison works for ISO 8601 strings).

---

## Phase 2 — Stream→Table Mapping

| Stream | Table |
|--------|-------|
| `main.tasks` | `task_board` |
| `main.notes` | `task_notes` |
| `main.tags` | `tag_master` |

---

## Phase 3 — Normalisation Rules

### `main.tasks`
| Rule | Column | Target |
|------|--------|--------|
| Rename | `id` | `task_id` |
| Rename | `title` | `task_title` |
| Rename | `status` | `task_status` |
| Exclude | `priority` | — (converted to text label in dbt) |

### `main.notes`
| Rule | Column | Target |
|------|--------|--------|
| Rename | `id` | `note_id` |
| Rename | `task_id` | `task_ref` |
| Rename | `content` | `note_body` |

### `main.tags`
| Rule | Column | Target |
|------|--------|--------|
| Rename | `id` | `tag_id` |
| Rename | `name` | `tag_name` |

---

## Phase 4 — dbt SQL

### Stream 1 — `main.tasks` → `task_board`
```sql
SELECT
    id                               AS task_id,
    title                            AS task_title,
    status                           AS task_status,
    CASE priority
        WHEN 3 THEN 'high'
        WHEN 2 THEN 'medium'
        ELSE       'low'
    END                              AS urgency,
    CAST(due_date   AS DATE)         AS due_on,
    CAST(created_at AS TIMESTAMPTZ)  AS created_on,
    CAST(updated_at AS TIMESTAMPTZ)  AS updated_on
FROM {{ source('raw', 'main__tasks') }}
WHERE title IS NOT NULL AND TRIM(title) != ''
```
**Preview check**:
- `urgency`: `'high'`/`'medium'`/`'low'` — NOT `3`/`2`/`1`
- `due_on`: DATE type (or NULL)
- `created_on`: TIMESTAMPTZ
- `priority` column absent

---

### Stream 2 — `main.notes` → `task_notes`
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

### Stream 3 — `main.tags` → `tag_master`
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
**Preview check**: `hex_color` always starts with `#`

---

## Phase 5 — Preview Checks

| Column | Expected |
|--------|---------|
| `urgency` | `'high'`/`'medium'`/`'low'` string |
| `due_on` | DATE or NULL |
| `hex_color` | `"#ff0000"` — starts with `#` |
| `note_body` | Non-empty text |

---

## Phase 5b — Destination-Specific Type Check

| Destination | `created_on` | `urgency` | `task_id` |
|-------------|-------------|----------|----------|
| PostgreSQL | `TIMESTAMPTZ` | `TEXT` | `INTEGER` |
| MySQL | `DATETIME` | `VARCHAR(10)` | `INT` |
| MariaDB | `DATETIME(6)` | `VARCHAR(10)` | `INT` |
| SQLite (dest) | `TEXT` | `TEXT` | `INTEGER` |
| CockroachDB | `TIMESTAMPTZ` | `STRING` | `INT8` |

---

## Phase 6 — Schedule Tests

| Test | Config | Expected |
|------|--------|---------|
| Hourly | Cron `0 * * * *` | Runs at top of each hour |
| Every 15 min | Cron `*/15 * * * *` | 4 runs/hour |
| None | None | Manual only |

---

## Phase 7 — Failure Scenarios

| Scenario | Expected |
|---------|---------|
| SQLite TEXT date `'2024-1-5'` (no zero-pad) | CAST may fail → NULL; fix source data |
| `created_at = 'now'` literal | CAST → NULL — guard with NULLIF |
| `priority = NULL` | CASE ELSE → `'low'` — acceptable |
| Empty `content` string | Filtered by `TRIM(content) != ''` |
| Source file locked by another process | Phase 1 ❌ `database is locked` |

---

## Phase 8 — Verify (PostgreSQL dest)

```sql
SELECT urgency, COUNT(*) FROM analytics.task_board GROUP BY urgency;
-- high/medium/low only
SELECT created_on FROM analytics.task_board LIMIT 3;        -- TIMESTAMPTZ
SELECT hex_color FROM analytics.tag_master WHERE hex_color NOT LIKE '#%'; -- 0 rows
SELECT task_id, COUNT(*) FROM analytics.task_board GROUP BY task_id HAVING COUNT(*)>1; -- 0
```
