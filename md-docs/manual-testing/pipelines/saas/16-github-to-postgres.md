# Pipeline 16 — GitHub → PostgreSQL

**Streams:** 12 | **Destination:** PostgreSQL

---

## Connections
```json
{ "credentials": { "personal_access_token": "ghp_..." }, "repositories": [{ "owner": "myorg", "name": "myrepo" }] }
{ "host":"..","port":5432,"database":"analytics","username":"writer","password":"..","ssl_mode":"disable" }
```
```sql
CREATE SCHEMA IF NOT EXISTS analytics;
```

---

## Stream 1 — `github.issues` → `analytics.gh_issues`

### Step 1 — DDL
```sql
CREATE TABLE analytics.gh_issues (
    issue_id     BIGINT      PRIMARY KEY,
    issue_number INTEGER,
    title        TEXT,
    state        TEXT,
    author_login TEXT,                        -- from JSON: user->>'login'
    label_name   TEXT,                        -- from JSON array: labels->0->>'name'
    is_closed    BOOLEAN,                     -- derived: state = 'closed'
    opened_on    DATE,
    closed_on    DATE
);
```
### Step 5 — dbt SQL
```sql
SELECT
    id            AS issue_id, number AS issue_number, title, state,
    user->>'login'      AS author_login,
    labels->0->>'name'  AS label_name,
    state = 'closed'    AS is_closed,
    created_at::DATE    AS opened_on,
    closed_at::DATE     AS closed_on
FROM {{ source('raw', 'github__issues') }}
```

---

## Stream 2 — `github.pull_requests` → `analytics.gh_pull_requests`

### Step 1 — DDL
```sql
CREATE TABLE analytics.gh_pull_requests (
    pr_id        BIGINT PRIMARY KEY,
    pr_number    INTEGER,
    title        TEXT,
    state        TEXT,
    author_login TEXT,                        -- from JSON: user->>'login'
    head_branch  TEXT,                        -- from JSON: head->>'ref'
    base_branch  TEXT,                        -- from JSON: base->>'ref'
    is_merged    BOOLEAN,
    opened_on    DATE,
    merged_on    DATE
);
```
### Step 5 — dbt SQL
```sql
SELECT
    id AS pr_id, number AS pr_number, title, state,
    user->>'login'     AS author_login,
    head->>'ref'       AS head_branch,
    base->>'ref'       AS base_branch,
    merged_at IS NOT NULL AS is_merged,
    created_at::DATE   AS opened_on,
    merged_at::DATE    AS merged_on
FROM {{ source('raw', 'github__pull_requests') }}
```

---

## Stream 3 — `github.commits` → `analytics.gh_commits`

### Step 1 — DDL
```sql
CREATE TABLE analytics.gh_commits (
    sha          TEXT PRIMARY KEY,
    message      TEXT,                        -- from JSON: commit->>'message'
    author_name  TEXT,                        -- from JSON: commit->'author'->>'name'
    author_email TEXT,                        -- from JSON: commit->'author'->>'email'
    committed_at TIMESTAMPTZ                  -- from JSON: commit->'author'->>'date' → TIMESTAMPTZ
);
```
### Step 5 — dbt SQL
```sql
SELECT
    sha,
    commit->>'message'                                  AS message,
    commit->'author'->>'name'                           AS author_name,
    commit->'author'->>'email'                          AS author_email,
    CAST(commit->'author'->>'date' AS TIMESTAMPTZ)      AS committed_at
FROM {{ source('raw', 'github__commits') }}
WHERE commit->'author'->>'name' IS NOT NULL
```
### Step 8 — Verify
```sql
SELECT sha, author_name, committed_at FROM analytics.gh_commits LIMIT 5;
-- committed_at: TIMESTAMPTZ; commit column must NOT exist
```

---

## Stream 4 — `github.events` → `analytics.gh_events`

### Step 1 — DDL
```sql
CREATE TABLE analytics.gh_events (
    event_id    TEXT PRIMARY KEY,
    event_type  TEXT,                         -- source: type
    actor_login TEXT,                         -- from JSON: actor->>'login'
    actor_id    INTEGER,                      -- from JSON: actor->>'id' → INT
    repo_name   TEXT,                         -- from JSON: repo->>'name'
    action      TEXT,                         -- from JSON: payload->>'action'
    occurred_at TIMESTAMPTZ
);
```
### Step 5 — dbt SQL
```sql
SELECT
    id AS event_id, type AS event_type,
    actor->>'login'              AS actor_login,
    CAST(actor->>'id' AS INTEGER) AS actor_id,
    repo->>'name'                AS repo_name,
    payload->>'action'           AS action,
    created_at                   AS occurred_at
FROM {{ source('raw', 'github__events') }}
WHERE type IN ('PushEvent','PullRequestEvent','IssuesEvent','WatchEvent','ForkEvent','ReleaseEvent')
```

---

## Stream 5 — `github.stargazers` → `analytics.gh_stars`

### Step 1 — DDL
```sql
CREATE TABLE analytics.gh_stars (
    user_login  TEXT PRIMARY KEY,             -- from JSON: user->>'login'
    user_id     INTEGER,                      -- from JSON: user->>'id' → INT
    starred_at  TIMESTAMPTZ
);
```
### Step 5 — dbt SQL
```sql
SELECT
    user->>'login'               AS user_login,
    CAST(user->>'id' AS INTEGER) AS user_id,
    starred_at
FROM {{ source('raw', 'github__stargazers') }}
```

---

## Stream 6 — `github.releases` → `analytics.gh_releases`

### Step 1 — DDL
```sql
CREATE TABLE analytics.gh_releases (
    release_id    BIGINT  PRIMARY KEY,
    version       TEXT,                       -- source: tag_name
    release_name  TEXT,                       -- source: name
    is_prerelease BOOLEAN,
    is_draft      BOOLEAN,
    author_login  TEXT,                       -- from JSON: author->>'login'
    published_on  DATE
);
```
### Step 5 — dbt SQL
```sql
SELECT
    id AS release_id, tag_name AS version, name AS release_name,
    prerelease AS is_prerelease, draft AS is_draft,
    author->>'login'         AS author_login,
    published_at::DATE       AS published_on
FROM {{ source('raw', 'github__releases') }}
WHERE draft = false
```

---

## Stream 7 — `github.contributors` → `analytics.gh_contributors`

### Step 1 — DDL
```sql
CREATE TABLE analytics.gh_contributors (
    github_user_id INTEGER PRIMARY KEY,       -- source: id
    github_login   TEXT,                      -- source: login
    commit_count   INTEGER,                   -- source: contributions
    avatar_url     TEXT,
    profile_url    TEXT                       -- source: html_url
);
```
### Step 5 — dbt SQL
```sql
SELECT
    id AS github_user_id, login AS github_login,
    contributions AS commit_count,
    avatar_url,
    html_url AS profile_url
FROM {{ source('raw', 'github__contributors') }}
ORDER BY contributions DESC
```

---

## Stream 8 — `github.milestones` → `analytics.gh_milestones`

### Step 1 — DDL
```sql
CREATE TABLE analytics.gh_milestones (
    milestone_id     BIGINT  PRIMARY KEY,
    milestone_number INTEGER,
    title            TEXT,
    state            TEXT,
    open_issues      INTEGER,
    closed_issues    INTEGER,
    due_on           DATE,
    closed_on        DATE
);
```
### Step 5 — dbt SQL
```sql
SELECT
    id AS milestone_id, number AS milestone_number, title, state,
    open_issues, closed_issues,
    due_on::DATE    AS due_on,
    closed_at::DATE AS closed_on
FROM {{ source('raw', 'github__milestones') }}
```

---

## Stream 9 — `github.labels` → `analytics.gh_labels`

### Step 1 — DDL
```sql
CREATE TABLE analytics.gh_labels (
    label_id   BIGINT PRIMARY KEY,
    label_name TEXT,
    hex_color  TEXT,
    is_default BOOLEAN,
    description TEXT
);
```
### Step 5 — dbt SQL
```sql
SELECT
    id AS label_id, name AS label_name,
    '#' || color AS hex_color,
    default AS is_default, description
FROM {{ source('raw', 'github__labels') }}
```

---

## Stream 10 — `github.forks` → `analytics.gh_forks`

### Step 1 — DDL
```sql
CREATE TABLE analytics.gh_forks (
    fork_id    BIGINT PRIMARY KEY,
    owner_login TEXT,                         -- from JSON: owner->>'login'
    fork_name  TEXT,
    is_private BOOLEAN,
    forked_on  DATE
);
```
### Step 5 — dbt SQL
```sql
SELECT
    id AS fork_id,
    owner->>'login'  AS owner_login,
    full_name        AS fork_name,
    private          AS is_private,
    created_at::DATE AS forked_on
FROM {{ source('raw', 'github__forks') }}
```

---

## Stream 11 — `github.projects` → `analytics.gh_projects`

### Step 1 — DDL
```sql
CREATE TABLE analytics.gh_projects (
    project_id     BIGINT PRIMARY KEY,
    project_number INTEGER,
    project_name   TEXT,
    project_state  TEXT,
    created_on     DATE
);
```
### Step 5 — dbt SQL
```sql
SELECT
    id AS project_id, number AS project_number,
    name AS project_name, state AS project_state,
    created_at::DATE AS created_on
FROM {{ source('raw', 'github__projects') }}
```

---

## Stream 12 — `github.branches` → `analytics.gh_branches`

### Step 1 — DDL
```sql
CREATE TABLE analytics.gh_branches (
    branch_name    TEXT PRIMARY KEY,
    commit_sha     TEXT,                      -- from JSON: commit->>'sha'
    is_protected   BOOLEAN
);
```
### Step 5 — dbt SQL
```sql
SELECT
    name      AS branch_name,
    commit->>'sha' AS commit_sha,
    protected AS is_protected
FROM {{ source('raw', 'github__branches') }}
```

---

## Edge Cases

| Scenario | Expected |
|---------|---------|
| `labels` empty array `[]` | `labels->0->>'name'` = NULL |
| `merged_at` NULL (open PR) | `merged_on` = NULL |
| `commit->'author'->>'date'` bad format | CAST fails → NULL |
| `actor->>'id'` is quoted number `"123"` | `CAST(... AS INTEGER)` = 123 |
| `payload->>'action'` NULL for PushEvent | `action` = NULL — OK |
