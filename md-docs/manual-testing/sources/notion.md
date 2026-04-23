# Notion Source — Manual Testing Guide

**Streams:** 3  
**Credential:** Integration token  
**DuckDB prefix:** `notion__`

---

## Credential Setup

```json
{ "token": "secret_..." }
```

The integration must be connected to the Notion workspace pages you want to sync.

---

## All 3 Streams

| Stream | DuckDB staging name | Key columns | INCREMENTAL key |
|--------|-------------------|-------------|----------------|
| `notion.databases` | `notion__databases` | `id`, `title`, `description`, `created_time`, `last_edited_time`, `url` | `last_edited_time` |
| `notion.pages` | `notion__pages` | `id`, `url`, `created_time`, `last_edited_time`, `parent`, `properties`, `archived` | `last_edited_time` |
| `notion.users` | `notion__users` | `id`, `name`, `type`, `avatar_url`, `person` | — |

---

## Scenario S-NOT-1 — Full Table Sync: `databases`

**Destination DDL:**
```sql
CREATE TABLE analytics.notion_databases_hd (
    id               TEXT PRIMARY KEY,
    title            TEXT,
    url              TEXT,
    created_time     TIMESTAMPTZ,
    last_edited_time TIMESTAMPTZ
);
```

**dbt SQL:**
```sql
SELECT
    id,
    title->0->>'plain_text'   AS title,
    url,
    created_time,
    last_edited_time
FROM {{ source('raw', 'notion__databases') }}
WHERE archived = false
```

> ℹ️ `title` in Notion API is a rich-text array. `->0->>'plain_text'` extracts the first text element.

**Verify:** Each row is one Notion database; `title` is readable text.

---

## Scenario S-NOT-2 — Pages with Properties Extraction

Notion `pages.properties` is a deeply nested JSON object (each key is a Notion column). Extract 3 specific property values:

**dbt SQL:**
```sql
SELECT
    id,
    url,
    properties->'Status'->'select'->>'name'              AS status,
    properties->'Priority'->'select'->>'name'            AS priority,
    properties->'Assignee'->'people'->0->>'name'         AS assignee,
    created_time,
    last_edited_time
FROM {{ source('raw', 'notion__pages') }}
WHERE archived = false
  AND properties->'Status' IS NOT NULL
```

**Destination DDL:**
```sql
CREATE TABLE analytics.notion_pages_flat (
    id               TEXT PRIMARY KEY,
    url              TEXT,
    status           TEXT,
    priority         TEXT,
    assignee         TEXT,
    created_time     TIMESTAMPTZ,
    last_edited_time TIMESTAMPTZ
);
```

**Verify:** `status`, `priority`, `assignee` columns contain values from the Notion database columns.  
`properties` JSON blob must NOT appear as a column in destination.

---

## Scenario S-NOT-3 — Incremental Sync: `pages`

**Sync mode:** `INCREMENTAL`, replication key `last_edited_time`

**dbt SQL:**
```sql
SELECT
    id,
    url,
    archived,
    created_time,
    last_edited_time
FROM {{ source('raw', 'notion__pages') }}
```

**Run 1:** All pages.  
**Run 2 (edit one page in Notion):** Only the edited page appears.  `rows_written = 1`.

---

## Scenario S-NOT-4 — Users Master Table

**dbt SQL:**
```sql
SELECT
    id,
    name,
    type,
    avatar_url,
    person->>'email'  AS email
FROM {{ source('raw', 'notion__users') }}
WHERE type = 'person'
```

**Destination DDL:**
```sql
CREATE TABLE analytics.notion_users_hd (
    id         TEXT PRIMARY KEY,
    name       TEXT,
    type       TEXT,
    avatar_url TEXT,
    email      TEXT
);
```

**Verify:** Only `person` type users (not bots); `email` extracted from nested `person` JSON.

---

## Scenario S-NOT-5 — Multi-Stream: `databases` + `pages` + `users`

All 3 Notion streams in one pipeline:

**Model 1:**
```sql
SELECT id, title->0->>'plain_text' AS title, url, created_time
FROM {{ source('raw', 'notion__databases') }}
```

**Model 2:**
```sql
SELECT id, url, archived, created_time, last_edited_time
FROM {{ source('raw', 'notion__pages') }}
```

**Model 3:**
```sql
SELECT id, name, type, person->>'email' AS email
FROM {{ source('raw', 'notion__users') }}
```

**Expected:** `dbt_models_run = 3`.

---

## Scenario S-NOT-6 — Normalisation: Rename `last_edited_time` → `updated_at`

**Rule:**
```json
{ "rule_type": "rename", "table": "notion.pages", "column": "last_edited_time", "destination_name": "updated_at" }
```

**dbt SQL:**
```sql
SELECT id, url, archived, created_time, updated_at
FROM {{ source('raw', 'notion__pages') }}
```

---

## Scenario S-NOT-7 — Error Path: Notion token missing workspace access

1. Create a Notion integration token that has **no** workspace access.
2. Add the connection in MantrixFlow.
3. Build a pipeline for `notion.databases`.
4. Click **Run**.

**Expected:** Phase 1 extraction fails with a Notion API `401` or `403` error. Run status: `failed`.

---

## All 3 Streams — Quick Smoke Test Checklist

| Stream | DuckDB ref | Expected rows |
|--------|-----------|---------------|
| databases | `notion__databases` | ≥ 1 |
| pages | `notion__pages` | ≥ 1 |
| users | `notion__users` | ≥ 1 |
