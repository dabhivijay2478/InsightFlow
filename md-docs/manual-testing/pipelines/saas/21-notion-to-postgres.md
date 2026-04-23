# Pipeline 21 — Notion → PostgreSQL

**Streams:** 3 | **Destination:** PostgreSQL

---

## Connections
```json
{ "credentials": { "auth_type": "token", "token": "secret_..." } }
{ "host":"..","port":5432,"database":"analytics","username":"writer","password":"..","ssl_mode":"disable" }
```
```sql
CREATE SCHEMA IF NOT EXISTS analytics;
```

---

## Stream 1 — `notion.databases` → `analytics.notion_databases`

### Step 1 — DDL
```sql
CREATE TABLE analytics.notion_databases (
    db_id       TEXT        PRIMARY KEY,
    db_title    TEXT,                         -- from JSON array: title->0->>'plain_text'
    db_url      TEXT,
    created_on  DATE,
    modified_on DATE
);
```
### Step 3 — Panel: `notion.databases` | `INCREMENTAL` | key: `last_edited_time`
### Step 5 — dbt SQL
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
### Step 8 — Verify
```sql
SELECT db_id, db_title, created_on FROM analytics.notion_databases LIMIT 5;
-- db_title: plain string, NOT a JSON array
-- created_on: DATE only
```

---

## Stream 2 — `notion.pages` → `analytics.notion_pages`

### Step 1 — DDL
```sql
CREATE TABLE analytics.notion_pages (
    page_id      TEXT        PRIMARY KEY,
    page_url     TEXT,
    status       TEXT,                        -- from JSON: properties->'Status'->'select'->>'name'
    priority     TEXT,                        -- from JSON: properties->'Priority'->'select'->>'name'
    assignee     TEXT,                        -- from JSON: properties->'Assignee'->'people'->0->>'name'
    due_date     DATE,                        -- from JSON: properties->'Due Date'->'date'->>'start' → DATE
    created_on   DATE,
    modified_on  DATE
    -- properties blob NOT stored
);
```
### Step 3 — Panel: `notion.pages` | `INCREMENTAL` | key: `last_edited_time`
### Step 5 — dbt SQL
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
### Step 8 — Verify
```sql
SELECT page_id, status, priority, assignee, due_date FROM analytics.notion_pages LIMIT 5;
-- status: plain string 'In Progress', NOT nested JSON
-- 'properties' column must NOT exist
-- due_date: DATE type
```

---

## Stream 3 — `notion.users` → `analytics.notion_users`

### Step 1 — DDL
```sql
CREATE TABLE analytics.notion_users (
    user_id      TEXT PRIMARY KEY,
    display_name TEXT,                        -- source: name (renamed)
    user_type    TEXT,
    email        TEXT,                        -- from JSON: person->>'email'
    avatar_url   TEXT
);
```
### Step 3 — Panel: `notion.users` | `FULL_TABLE`
### Step 5 — dbt SQL
```sql
SELECT
    id              AS user_id,
    name            AS display_name,
    type            AS user_type,
    person->>'email' AS email,
    avatar_url
FROM {{ source('raw', 'notion__users') }}
WHERE type = 'person'
```
### Step 8 — Verify
```sql
SELECT user_id, display_name, email FROM analytics.notion_users LIMIT 5;
-- email: plain string; 'person' column must NOT exist
-- user_type: 'person' only (bot filtered out)
```

---

## Edge Cases

| Scenario | Expected |
|---------|---------|
| `title` empty array `[]` | `title->0->>'plain_text'` = NULL → filtered by WHERE |
| `properties->'Status'` NULL (page has no Status property) | `status` = NULL |
| `person` key missing for bot users | `email` = NULL; bots filtered by `WHERE type='person'` |
| `due_date->>'start'` NULL | `CAST(NULL AS DATE)` = NULL — OK |
| `archived = true` page | Filtered out by WHERE |
| Workspace access revoked mid-run | Phase 1 fails: Notion 403 |
