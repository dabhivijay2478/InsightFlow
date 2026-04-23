# Notion Source — UI Testing (All 3 Streams × 5 Destinations)

> Universal builder steps in `builder-walkthrough.md`.

---

## Phase 1 — Source Panel (Notion)

### Credential Fields
| Field | Value |
|-------|-------|
| **Auth type** | `token` |
| **Internal Integration Token** | `secret_...` |

**Test Connection → ✅**
- ❌ `401`: invalid or expired token
- ❌ `403`: integration not invited to target pages/databases — share workspace with integration in Notion settings

---

## Step 2b — Stream Selection & Sync Mode

| Stream | Sync Mode | Cursor Field |
|--------|-----------|-------------|
| `databases` | INCREMENTAL | `last_edited_time` |
| `pages` | INCREMENTAL | `last_edited_time` |
| `users` | FULL TABLE | — |

---

## Phase 2 — Stream→Table Mapping

| Stream | Table |
|--------|-------|
| databases | `notion_databases` |
| pages | `notion_pages` |
| users | `notion_users` |

Schema: `analytics` (all destinations); `main` for SQLite

---

## Phase 3 — Normalisation Rules

### `notion.databases`
| Rule | Column | Target |
|------|--------|--------|
| Rename | `id` | `db_id` |
| Rename | `url` | `db_url` |
| Exclude | `title` | — (first element extracted in dbt) |
| Exclude | `properties` | — |
| Exclude | `parent` | — |

### `notion.pages`
| Rule | Column | Target |
|------|--------|--------|
| Rename | `id` | `page_id` |
| Rename | `url` | `page_url` |
| Exclude | `properties` | — (specific keys extracted in dbt) |
| Exclude | `parent` | — |
| Exclude | `cover` | — |
| Exclude | `icon` | — |

### `notion.users`
| Rule | Column | Target |
|------|--------|--------|
| Rename | `id` | `user_id` |
| Rename | `name` | `display_name` |
| Rename | `type` | `user_type` |
| Exclude | `person` | — (email extracted in dbt) |
| Exclude | `bot` | — |

---

## Phase 4 — dbt SQL (stream-by-stream)

### Stream 1 — `notion.databases`
```sql
SELECT
    id                                       AS db_id,
    title->0->>'plain_text'                  AS db_title,
    url                                      AS db_url,
    CAST(created_time AS DATE)               AS created_on,
    CAST(last_edited_time AS DATE)           AS modified_on
FROM {{ source('raw', 'notion__databases') }}
WHERE archived = false
  AND title->0->>'plain_text' IS NOT NULL
```
**Preview check**: `db_title` plain string e.g. `"Tasks"` — NOT a JSON array; `properties` column absent

---

### Stream 2 — `notion.pages`
```sql
SELECT
    id                                                       AS page_id,
    url                                                      AS page_url,
    properties->'Status'->'select'->>'name'                  AS status,
    properties->'Priority'->'select'->>'name'                AS priority,
    properties->'Assignee'->'people'->0->>'name'             AS assignee,
    CAST(properties->'Due Date'->'date'->>'start' AS DATE)   AS due_date,
    CAST(created_time AS DATE)                               AS created_on,
    CAST(last_edited_time AS DATE)                           AS modified_on
FROM {{ source('raw', 'notion__pages') }}
WHERE archived = false
```
**Preview check**:
- `status`: plain string `"In Progress"` — NOT nested JSON
- `assignee`: person name or NULL
- `due_date`: DATE or NULL
- `properties` column must NOT appear

**Test with missing property**: Create a Notion page without a `Status` property → `status` = NULL in preview ✅

---

### Stream 3 — `notion.users`
```sql
SELECT
    id               AS user_id,
    name             AS display_name,
    type             AS user_type,
    person->>'email' AS email,
    avatar_url
FROM {{ source('raw', 'notion__users') }}
WHERE type = 'person'
```
**Preview check**: `email` plain string; `person` column absent; `user_type` = `"person"` only (bots filtered)

---

## Phase 5 — Preview Checks

| Stream | Column | Expected |
|--------|--------|---------|
| databases | `db_title` | `"My Database"` (plain text, not JSON) |
| pages | `status` | `"Done"` / `"In Progress"` / NULL |
| pages | `properties` | Must NOT appear |
| users | `email` | `"user@example.com"` |
| users | `person` | Must NOT appear |

---

## Phase 5 — Destination Variants

### Test A — PostgreSQL dest
- `db_title TEXT`, `status TEXT`, `due_date DATE`, `email TEXT`

### Test B — MySQL dest
- `db_title VARCHAR(500)`, `status VARCHAR(100)`, `due_date DATE`, `email VARCHAR(255)`

### Test C — MariaDB dest
- Same as MySQL

### Test D — SQLite dest
- All columns `TEXT`; `due_date` stored as ISO string; verify with `sqlite3 ... "SELECT due_date FROM notion_pages LIMIT 3;"`

### Test E — CockroachDB dest
- `db_title STRING`, `status STRING`, `due_date DATE`, `email STRING`

---

## Phase 6 — Schedule Tests

| Test | Config | Expected |
|------|--------|---------|
| Every 30 min | Cron `*/30 * * * *` | Runs twice per hour |
| Daily at midnight | Cron `0 0 * * *` | Nightly sync |
| None | None | Manual only |

---

## Phase 7 — Failure Scenarios

| Scenario | How to trigger | Expected |
|---------|---------------|---------|
| Integration not shared | Remove integration from Notion workspace | Phase 1 ❌ `403` |
| `title` empty array `[]` | Create database with no name | `db_title` = NULL; WHERE filters row |
| `archived = true` page | Archive a page in Notion | Row absent from output |
| `due_date` bad format | Set date as text in Notion | CAST → NULL |
| `person` key missing on bot | Bot user included | Filtered by `WHERE type='person'` |

---

## Phase 8 — Verify

```sql
-- PostgreSQL destination
SELECT db_id, db_title FROM analytics.notion_databases LIMIT 5;
-- db_title must be plain string

SELECT page_id, status, priority, assignee, due_date
FROM analytics.notion_pages LIMIT 5;
-- status/priority: plain strings or NULL

SELECT user_id, display_name, email
FROM analytics.notion_users WHERE user_type != 'person';
-- Must return 0 rows (bots filtered)

-- No staging tables
SELECT table_name FROM information_schema.tables
WHERE table_schema = 'analytics' AND table_name LIKE '_dlt_%';
-- 0 rows
```
