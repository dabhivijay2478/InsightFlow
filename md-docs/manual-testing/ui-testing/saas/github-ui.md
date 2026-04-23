# GitHub Source — UI Testing (All 12 Streams × 5 Destinations)

> Universal builder steps in `builder-walkthrough.md`.

---

## Phase 1 — Source Panel (GitHub)

### Credential Fields
| Field | Value |
|-------|-------|
| **Personal Access Token** | `ghp_...` |
| **Owner** | `myorg` or `myusername` |
| **Repository** | `myrepo` |

Multiple repos: add additional `owner/repo` pairs in the repos list panel.

**Test Connection → ✅**
- ❌ `401`: bad token
- ❌ `404`: repo not found or private (token lacks `repo` scope)
- ❌ `403`: rate limited — use fine-grained token with appropriate permissions

---

## Step 2b — Stream Selection & Sync Mode

| Stream | Sync Mode | Cursor Field |
|--------|-----------|-------------|
| `issues` | INCREMENTAL | `updated_at` |
| `pull_requests` | INCREMENTAL | `updated_at` |
| `commits` | INCREMENTAL | `committed_date` |
| `events` | INCREMENTAL | `created_at` |
| `stargazers` | INCREMENTAL | `starred_at` |
| `releases` | INCREMENTAL | `created_at` |
| `contributors` | FULL TABLE | — |
| `milestones` | INCREMENTAL | `updated_at` |
| `labels` | FULL TABLE | — |
| `forks` | INCREMENTAL | `created_at` |
| `projects` | INCREMENTAL | `updated_at` |
| `branches` | FULL TABLE | — |

---

## Phase 2 — Stream→Table Mapping

| Stream | Table |
|--------|-------|
| issues | `gh_issues` |
| pull_requests | `gh_pull_requests` |
| commits | `gh_commits` |
| events | `gh_events` |
| stargazers | `gh_stars` |
| releases | `gh_releases` |
| contributors | `gh_contributors` |
| milestones | `gh_milestones` |
| labels | `gh_labels` |
| forks | `gh_forks` |
| projects | `gh_projects` |
| branches | `gh_branches` |

---

## Phase 3 — Normalisation Rules

### `github.issues`
| Rule | Column | Target |
|------|--------|--------|
| Rename | `id` | `issue_id` |
| Rename | `number` | `issue_number` |
| Exclude | `user` | — (extracted via dbt JSON) |
| Exclude | `labels` | — (first label extracted in dbt) |
| Exclude | `assignees` | — |
| Exclude | `pull_request` | — |

### `github.commits`
| Rule | Column | Target |
|------|--------|--------|
| — | *(dlt loads `commit__message`, `commit__author__name`, …)* | Use these in dbt (no JSON `commit` column) |
| Exclude | `parents` | — (nested table `github__commits__parents` if needed) |

### `github.events`
| Rule | Column | Target |
|------|--------|--------|
| Rename | `id` | `event_id` |
| Rename | `type` | `event_type` |
| Exclude | `actor` | — (JSON key extracted) |
| Exclude | `repo` | — |
| Exclude | `payload` | — |

---

## Phase 4 — dbt SQL (stream-by-stream)

### Stream 1 — `github.issues`
```sql
SELECT
    id              AS issue_id,
    number          AS issue_number,
    title, state,
    user->>'login'       AS author_login,
    labels->0->>'name'   AS label_name,
    state = 'closed'     AS is_closed,
    created_at::DATE     AS opened_on,
    closed_at::DATE      AS closed_on
FROM {{ source('raw', 'github__issues') }}
```
**Preview check**: `author_login` plain string; `label_name` first label or NULL; `user` column absent

---

### Stream 2 — `github.pull_requests`
```sql
SELECT
    id AS pr_id, number AS pr_number, title, state,
    user->>'login'         AS author_login,
    head->>'ref'           AS head_branch,
    base->>'ref'           AS base_branch,
    merged_at IS NOT NULL  AS is_merged,
    created_at::DATE       AS opened_on,
    merged_at::DATE        AS merged_on
FROM {{ source('raw', 'github__pull_requests') }}
```

---

### Stream 3 — `github.commits`
dlt stores **flattened** columns (e.g. `commit__message`), not a single JSON `commit` column—use these in dbt over `github__commits`.

```sql
SELECT
    sha,
    commit__message                         AS message,
    commit__author__name                    AS author_name,
    commit__author__email                   AS author_email,
    CAST(commit__author__date AS TIMESTAMPTZ) AS committed_at
FROM {{ source('raw', 'github__commits') }}
WHERE commit__author__name IS NOT NULL
```
**Preview check**: `/preview` flattens nested JSON like dlt (`commit__author__name`, …), matching **sync/DuckDB** dbt SQL.

---

### Stream 4 — `github.events`
```sql
SELECT
    id AS event_id, type AS event_type,
    actor->>'login'                AS actor_login,
    CAST(actor->>'id' AS INTEGER)  AS actor_id,
    repo->>'name'                  AS repo_name,
    payload->>'action'             AS action,
    created_at                     AS occurred_at
FROM {{ source('raw', 'github__events') }}
WHERE type IN ('PushEvent','PullRequestEvent','IssuesEvent','WatchEvent','ForkEvent','ReleaseEvent')
```
**Preview check**: `actor_login` string; `actor_id` integer; `actor` column absent

---

### Stream 5 — `github.stargazers`
```sql
SELECT
    user->>'login'               AS user_login,
    CAST(user->>'id' AS INTEGER) AS user_id,
    starred_at
FROM {{ source('raw', 'github__stargazers') }}
```

---

### Stream 6 — `github.releases`
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

### Stream 7 — `github.contributors`
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

### Stream 8 — `github.milestones`
```sql
SELECT
    id AS milestone_id, number AS milestone_number,
    title, state,
    open_issues, closed_issues,
    due_on::DATE    AS due_on,
    closed_at::DATE AS closed_on
FROM {{ source('raw', 'github__milestones') }}
```

---

### Stream 9 — `github.labels`
```sql
SELECT
    id AS label_id, name AS label_name,
    '#' || color AS hex_color,
    default AS is_default, description
FROM {{ source('raw', 'github__labels') }}
```
**Preview check**: `hex_color` starts with `#`; no duplicates for same label

---

### Stream 10 — `github.forks`
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

### Stream 11 — `github.projects`
```sql
SELECT
    id AS project_id, number AS project_number,
    name AS project_name, state AS project_state,
    created_at::DATE AS created_on
FROM {{ source('raw', 'github__projects') }}
```

---

### Stream 12 — `github.branches`
dlt flattens `commit.sha` → `commit__sha`.
```sql
SELECT
    name            AS branch_name,
    commit__sha     AS commit_sha,
    protected       AS is_protected
FROM {{ source('raw', 'github__branches') }}
```

---

## Phase 5 — Preview Checks

| Stream | Column | Expected |
|--------|--------|---------|
| issues | `author_login` | Plain string e.g. `"octocat"` |
| issues | `label_name` | First label string or NULL |
| commits | `commit__author__name` in DuckDB; preview may differ | Plain string in dbt SELECT |
| events | `actor_id` | Integer |
| labels | `hex_color` | String starting with `#` |
| branches | `commit_sha` | 40-char SHA string |

---

## Phase 7 — Failure Scenarios

| Scenario | Expected |
|---------|---------|
| `labels` empty array | `labels->0->>'name'` = NULL — ✅ |
| `merged_at` NULL (open PR) | `merged_on` = NULL; `is_merged` = false — ✅ |
| `commit__author__date` NULL or bad | CAST → NULL — ✅ |
| Token with no `repo` scope | Phase 1 ❌ `403` |
| 5000 req/hr rate limit hit | Phase 1 ❌ rate limit error |

---

## Phase 8 — Verify

```sql
SELECT issue_id, author_login, label_name FROM analytics.gh_issues LIMIT 5;
SELECT author_name, committed_at FROM analytics.gh_commits LIMIT 5;
SELECT hex_color FROM analytics.gh_labels WHERE hex_color NOT LIKE '#%';  -- 0 rows
SELECT sha, COUNT(*) FROM analytics.gh_commits GROUP BY sha HAVING COUNT(*)>1; -- 0 rows
```
