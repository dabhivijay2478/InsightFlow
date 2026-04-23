# Pipeline 19 — GitHub → SQLite

**Streams:** 12 | **Destination:** SQLite

> dbt SQL identical to `16-github-to-postgres.md`.

---

## Connections
```json
{ "credentials": { "personal_access_token": "ghp_..." }, "repositories": [{ "owner": "myorg", "name": "myrepo" }] }
{ "database": "/absolute/path/to/analytics.db" }
```

---

## All 12 Stream DDLs (SQLite)

```sql
CREATE TABLE IF NOT EXISTS gh_issues (
    issue_id INTEGER PRIMARY KEY, issue_number INTEGER, title TEXT, state TEXT,
    author_login TEXT, label_name TEXT, is_closed INTEGER, opened_on TEXT, closed_on TEXT
);
CREATE TABLE IF NOT EXISTS gh_pull_requests (
    pr_id INTEGER PRIMARY KEY, pr_number INTEGER, title TEXT, state TEXT,
    author_login TEXT, head_branch TEXT, base_branch TEXT,
    is_merged INTEGER, opened_on TEXT, merged_on TEXT
);
CREATE TABLE IF NOT EXISTS gh_commits (
    sha TEXT PRIMARY KEY, message TEXT,
    author_name TEXT, author_email TEXT, committed_at TEXT
);
CREATE TABLE IF NOT EXISTS gh_events (
    event_id TEXT PRIMARY KEY, event_type TEXT,
    actor_login TEXT, actor_id INTEGER, repo_name TEXT,
    action TEXT, occurred_at TEXT
);
CREATE TABLE IF NOT EXISTS gh_stars (
    user_login TEXT PRIMARY KEY, user_id INTEGER, starred_at TEXT
);
CREATE TABLE IF NOT EXISTS gh_releases (
    release_id INTEGER PRIMARY KEY, version TEXT, release_name TEXT,
    is_prerelease INTEGER, is_draft INTEGER,
    author_login TEXT, published_on TEXT
);
CREATE TABLE IF NOT EXISTS gh_contributors (
    github_user_id INTEGER PRIMARY KEY, github_login TEXT,
    commit_count INTEGER, avatar_url TEXT, profile_url TEXT
);
CREATE TABLE IF NOT EXISTS gh_milestones (
    milestone_id INTEGER PRIMARY KEY, milestone_number INTEGER,
    title TEXT, state TEXT,
    open_issues INTEGER, closed_issues INTEGER, due_on TEXT, closed_on TEXT
);
CREATE TABLE IF NOT EXISTS gh_labels (
    label_id INTEGER PRIMARY KEY, label_name TEXT,
    hex_color TEXT, is_default INTEGER, description TEXT
);
CREATE TABLE IF NOT EXISTS gh_forks (
    fork_id INTEGER PRIMARY KEY, owner_login TEXT,
    fork_name TEXT, is_private INTEGER, forked_on TEXT
);
CREATE TABLE IF NOT EXISTS gh_projects (
    project_id INTEGER PRIMARY KEY, project_number INTEGER,
    project_name TEXT, project_state TEXT, created_on TEXT
);
CREATE TABLE IF NOT EXISTS gh_branches (
    branch_name TEXT PRIMARY KEY, commit_sha TEXT, is_protected INTEGER
);
```

---

## Verify
```bash
DB=/absolute/path/to/analytics.db
sqlite3 $DB "SELECT is_closed FROM gh_issues LIMIT 5;"         # 0 or 1
sqlite3 $DB "SELECT committed_at FROM gh_commits LIMIT 3;"     # ISO string
sqlite3 $DB "SELECT typeof(actor_id) FROM gh_events LIMIT 1;"  # integer
```
