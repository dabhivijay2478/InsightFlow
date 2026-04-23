# GitHub Source — Manual Testing Guide

**Streams:** 12  
**Credential:** Personal access token + repository  
**DuckDB prefix:** `github__`

---

## Credential Setup

```json
{
  "access_token": "ghp_...",
  "repository": "owner/repo-name"
}
```

The `repository` field must be `owner/repo` format (e.g., `acme/backend-api`).

---

## All 12 Streams

| Stream | DuckDB staging name | Key columns | INCREMENTAL key |
|--------|-------------------|-------------|----------------|
| `github.issues` | `github__issues` | `id`, `number`, `title`, `state`, `user`, `labels`, `assignees`, `created_at`, `updated_at`, `closed_at` | `updated_at` |
| `github.pull_requests` | `github__pull_requests` | `id`, `number`, `title`, `state`, `user`, `head`, `base`, `merged_at`, `created_at`, `updated_at` | `updated_at` |
| `github.stargazers` | `github__stargazers` | `user`, `starred_at` | `starred_at` |
| `github.events` | `github__events` | `id`, `type`, `actor`, `repo`, `payload`, `created_at` | `created_at` |
| `github.commits` | `github__commits` | `sha`, `commit`, `author`, `committer`, `message`, `url` | — |
| `github.branches` | `github__branches` | `name`, `commit`, `protected` | — |
| `github.releases` | `github__releases` | `id`, `tag_name`, `name`, `body`, `draft`, `prerelease`, `created_at`, `published_at` | `created_at` |
| `github.tags` | `github__tags` | `name`, `commit`, `zipball_url`, `tarball_url` | — |
| `github.contributors` | `github__contributors` | `login`, `id`, `contributions`, `avatar_url`, `type` | — |
| `github.milestones` | `github__milestones` | `id`, `number`, `title`, `state`, `open_issues`, `closed_issues`, `created_at`, `updated_at` | `updated_at` |
| `github.labels` | `github__labels` | `id`, `name`, `color`, `description`, `default` | — |
| `github.forks` | `github__forks` | `id`, `full_name`, `owner`, `private`, `created_at`, `updated_at` | `created_at` |

---

## Scenario S-GH-1 — Full Table Sync: `issues`

**Destination DDL:**
```sql
CREATE TABLE analytics.github_issues_hd (
    id         BIGINT PRIMARY KEY,
    number     INTEGER,
    title      TEXT,
    state      TEXT,
    created_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ,
    closed_at  TIMESTAMPTZ
);
```

**dbt SQL:**
```sql
SELECT
    id,
    number,
    title,
    state,
    created_at,
    updated_at,
    closed_at
FROM {{ source('raw', 'github__issues') }}
ORDER BY number DESC
```

**Verify:** `SELECT state, COUNT(*) FROM analytics.github_issues_hd GROUP BY state;` — shows open/closed counts.

---

## Scenario S-GH-2 — Incremental Sync: `pull_requests`

**Sync mode:** `INCREMENTAL`, replication key `updated_at`

**dbt SQL:**
```sql
SELECT
    id,
    number,
    title,
    state,
    merged_at IS NOT NULL  AS is_merged,
    created_at,
    updated_at
FROM {{ source('raw', 'github__pull_requests') }}
```

**Destination DDL:**
```sql
CREATE TABLE analytics.github_prs_hd (
    id         BIGINT PRIMARY KEY,
    number     INTEGER,
    title      TEXT,
    state      TEXT,
    is_merged  BOOLEAN,
    created_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ
);
```

---

## Scenario S-GH-3 — Events JSON Payload Filtering

GitHub `events.payload` is a large JSON object (varies by event type). Extract only key fields:

**dbt SQL:**
```sql
SELECT
    id,
    type                          AS event_type,
    actor->>'login'               AS actor_login,
    repo->>'name'                 AS repo_name,
    payload->>'action'            AS action,
    created_at
FROM {{ source('raw', 'github__events') }}
WHERE type IN ('PushEvent', 'PullRequestEvent', 'IssuesEvent', 'WatchEvent')
```

**Destination DDL:**
```sql
CREATE TABLE analytics.github_events_flat (
    id          TEXT PRIMARY KEY,
    event_type  TEXT,
    actor_login TEXT,
    repo_name   TEXT,
    action      TEXT,
    created_at  TIMESTAMPTZ
);
```

**Verify:** Only 4 event types in destination; `payload` blob not present.

---

## Scenario S-GH-4 — Commits with Author Extraction

dlt **flattens** the API `commit` object into columns (`commit__message`, `commit__author__name`, …); there is no `commit` column in `github__commits`.

**dbt SQL:**
```sql
SELECT
    sha,
    commit__message                AS message,
    commit__author__name           AS author_name,
    commit__author__email         AS author_email,
    CAST(commit__author__date AS TIMESTAMPTZ) AS committed_at
FROM {{ source('raw', 'github__commits') }}
WHERE commit__author__name IS NOT NULL
```

**Destination DDL:**
```sql
CREATE TABLE analytics.github_commits_flat (
    sha          TEXT PRIMARY KEY,
    message      TEXT,
    author_name  TEXT,
    author_email TEXT,
    committed_at TEXT
);
```

---

## Scenario S-GH-5 — Contributor Leaderboard

**dbt SQL:**
```sql
SELECT
    login,
    id          AS github_user_id,
    contributions,
    type
FROM {{ source('raw', 'github__contributors') }}
ORDER BY contributions DESC
```

---

## Scenario S-GH-6 — Stargazer Growth (time series)

**dbt SQL:**
```sql
SELECT
    DATE_TRUNC('week', starred_at::TIMESTAMPTZ)  AS week,
    COUNT(*)                                      AS new_stars
FROM {{ source('raw', 'github__stargazers') }}
GROUP BY 1
ORDER BY 1
```

**Destination DDL (no PK — append):**
```sql
CREATE TABLE analytics.github_stars_weekly (
    week      TIMESTAMPTZ,
    new_stars BIGINT
);
```

---

## Scenario S-GH-7 — Normalisation: Rename `state` → `issue_state`

**Rule:**
```json
{ "rule_type": "rename", "table": "github.issues", "column": "state", "destination_name": "issue_state" }
```

**dbt SQL:**
```sql
SELECT id, number, title, issue_state, created_at FROM {{ source('raw', 'github__issues') }}
```

---

## Scenario S-GH-8 — Releases vs Pre-releases Filter

**dbt SQL:**
```sql
SELECT
    id,
    tag_name,
    name,
    draft,
    prerelease,
    created_at,
    published_at
FROM {{ source('raw', 'github__releases') }}
WHERE draft = false
  AND prerelease = false
```

---

## Scenario S-GH-9 — Multi-Stream: `issues` + `pull_requests` + `milestones`

Three streams in one pipeline:

**Model 1:**
```sql
SELECT id, number, title, state, created_at FROM {{ source('raw', 'github__issues') }}
```

**Model 2:**
```sql
SELECT id, number, title, state, merged_at, created_at FROM {{ source('raw', 'github__pull_requests') }}
```

**Model 3:**
```sql
SELECT id, number, title, state, open_issues, closed_issues FROM {{ source('raw', 'github__milestones') }}
```

---

## All 12 Streams — Quick Smoke Test Checklist

| Stream | DuckDB ref | Expected rows |
|--------|-----------|---------------|
| issues | `github__issues` | ≥ 0 |
| pull_requests | `github__pull_requests` | ≥ 0 |
| stargazers | `github__stargazers` | ≥ 0 |
| events | `github__events` | ≥ 0 |
| commits | `github__commits` | ≥ 1 |
| branches | `github__branches` | ≥ 1 |
| releases | `github__releases` | ≥ 0 |
| tags | `github__tags` | ≥ 0 |
| contributors | `github__contributors` | ≥ 1 |
| milestones | `github__milestones` | ≥ 0 |
| labels | `github__labels` | ≥ 0 |
| forks | `github__forks` | ≥ 0 |
