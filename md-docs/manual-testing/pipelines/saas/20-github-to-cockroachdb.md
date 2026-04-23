# Pipeline 20 — GitHub → CockroachDB

**Streams:** 12 | **Destination:** CockroachDB

> dbt SQL identical to `16-github-to-postgres.md`.

---

## Connections
```json
{ "credentials": { "personal_access_token": "ghp_..." }, "repositories": [{ "owner": "myorg", "name": "myrepo" }] }
{ "host":"..","port":26257,"database":"defaultdb","username":"root","password":"","ssl_mode":"disable" }
```
```sql
CREATE SCHEMA IF NOT EXISTS analytics;
```

---

## All 12 Stream DDLs (CockroachDB)

```sql
CREATE TABLE IF NOT EXISTS analytics.gh_issues (
    issue_id INT8 PRIMARY KEY, issue_number INT8, title STRING, state STRING,
    author_login STRING, label_name STRING, is_closed BOOL, opened_on DATE, closed_on DATE
);
CREATE TABLE IF NOT EXISTS analytics.gh_pull_requests (
    pr_id INT8 PRIMARY KEY, pr_number INT8, title STRING, state STRING,
    author_login STRING, head_branch STRING, base_branch STRING,
    is_merged BOOL, opened_on DATE, merged_on DATE
);
CREATE TABLE IF NOT EXISTS analytics.gh_commits (
    sha STRING PRIMARY KEY, message STRING,
    author_name STRING, author_email STRING, committed_at TIMESTAMPTZ
);
CREATE TABLE IF NOT EXISTS analytics.gh_events (
    event_id STRING PRIMARY KEY, event_type STRING,
    actor_login STRING, actor_id INT8, repo_name STRING,
    action STRING, occurred_at TIMESTAMPTZ
);
CREATE TABLE IF NOT EXISTS analytics.gh_stars (
    user_login STRING PRIMARY KEY, user_id INT8, starred_at TIMESTAMPTZ
);
CREATE TABLE IF NOT EXISTS analytics.gh_releases (
    release_id INT8 PRIMARY KEY, version STRING, release_name STRING,
    is_prerelease BOOL, is_draft BOOL, author_login STRING, published_on DATE
);
CREATE TABLE IF NOT EXISTS analytics.gh_contributors (
    github_user_id INT8 PRIMARY KEY, github_login STRING,
    commit_count INT8, avatar_url STRING, profile_url STRING
);
CREATE TABLE IF NOT EXISTS analytics.gh_milestones (
    milestone_id INT8 PRIMARY KEY, milestone_number INT8,
    title STRING, state STRING,
    open_issues INT8, closed_issues INT8, due_on DATE, closed_on DATE
);
CREATE TABLE IF NOT EXISTS analytics.gh_labels (
    label_id INT8 PRIMARY KEY, label_name STRING,
    hex_color STRING, is_default BOOL, description STRING
);
CREATE TABLE IF NOT EXISTS analytics.gh_forks (
    fork_id INT8 PRIMARY KEY, owner_login STRING,
    fork_name STRING, is_private BOOL, forked_on DATE
);
CREATE TABLE IF NOT EXISTS analytics.gh_projects (
    project_id INT8 PRIMARY KEY, project_number INT8,
    project_name STRING, project_state STRING, created_on DATE
);
CREATE TABLE IF NOT EXISTS analytics.gh_branches (
    branch_name STRING PRIMARY KEY, commit_sha STRING, is_protected BOOL
);
```

---

## Verify
```sql
SELECT is_closed FROM analytics.gh_issues LIMIT 5;    -- true/false
SELECT committed_at FROM analytics.gh_commits LIMIT 3; -- TIMESTAMPTZ
SELECT issue_id, COUNT(*) FROM analytics.gh_issues GROUP BY issue_id HAVING COUNT(*)>1;
```
