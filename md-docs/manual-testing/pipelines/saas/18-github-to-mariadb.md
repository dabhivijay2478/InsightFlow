# Pipeline 18 — GitHub → MariaDB

**Streams:** 12 | **Destination:** MariaDB

> DDL same as `17-github-to-mysql.md` with `DATETIME(6)`. dbt SQL = `16-github-to-postgres.md`.

---

## Connections
```json
{ "credentials": { "personal_access_token": "ghp_..." }, "repositories": [{ "owner": "myorg", "name": "myrepo" }] }
{ "host":"..","port":3306,"database":"analytics","username":"writer","password":"..." }
```

---

## All 12 Stream DDLs (MariaDB)

```sql
CREATE TABLE analytics.gh_issues (
    issue_id BIGINT PRIMARY KEY, issue_number INT, title TEXT, state VARCHAR(20),
    author_login VARCHAR(255), label_name VARCHAR(255),
    is_closed TINYINT(1), opened_on DATE, closed_on DATE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE analytics.gh_pull_requests (
    pr_id BIGINT PRIMARY KEY, pr_number INT, title TEXT, state VARCHAR(20),
    author_login VARCHAR(255), head_branch VARCHAR(255), base_branch VARCHAR(255),
    is_merged TINYINT(1), opened_on DATE, merged_on DATE
) ENGINE=InnoDB;

CREATE TABLE analytics.gh_commits (
    sha TEXT PRIMARY KEY, message TEXT,
    author_name VARCHAR(255), author_email VARCHAR(255), committed_at DATETIME(6)
) ENGINE=InnoDB;

CREATE TABLE analytics.gh_events (
    event_id VARCHAR(255) PRIMARY KEY, event_type VARCHAR(100),
    actor_login VARCHAR(255), actor_id INT, repo_name VARCHAR(500),
    action VARCHAR(100), occurred_at DATETIME(6)
) ENGINE=InnoDB;

CREATE TABLE analytics.gh_stars (
    user_login VARCHAR(255) PRIMARY KEY, user_id INT, starred_at DATETIME(6)
) ENGINE=InnoDB;

CREATE TABLE analytics.gh_releases (
    release_id BIGINT PRIMARY KEY, version VARCHAR(100), release_name VARCHAR(500),
    is_prerelease TINYINT(1), is_draft TINYINT(1),
    author_login VARCHAR(255), published_on DATE
) ENGINE=InnoDB;

CREATE TABLE analytics.gh_contributors (
    github_user_id INT PRIMARY KEY, github_login VARCHAR(255),
    commit_count INT, avatar_url TEXT, profile_url TEXT
) ENGINE=InnoDB;

CREATE TABLE analytics.gh_milestones (
    milestone_id BIGINT PRIMARY KEY, milestone_number INT,
    title VARCHAR(500), state VARCHAR(20),
    open_issues INT, closed_issues INT, due_on DATE, closed_on DATE
) ENGINE=InnoDB;

CREATE TABLE analytics.gh_labels (
    label_id BIGINT PRIMARY KEY, label_name VARCHAR(255),
    hex_color VARCHAR(10), is_default TINYINT(1), description TEXT
) ENGINE=InnoDB;

CREATE TABLE analytics.gh_forks (
    fork_id BIGINT PRIMARY KEY, owner_login VARCHAR(255),
    fork_name VARCHAR(500), is_private TINYINT(1), forked_on DATE
) ENGINE=InnoDB;

CREATE TABLE analytics.gh_projects (
    project_id BIGINT PRIMARY KEY, project_number INT,
    project_name VARCHAR(500), project_state VARCHAR(20), created_on DATE
) ENGINE=InnoDB;

CREATE TABLE analytics.gh_branches (
    branch_name VARCHAR(255) PRIMARY KEY,
    commit_sha VARCHAR(100), is_protected TINYINT(1)
) ENGINE=InnoDB;
```
