# Pipeline 23 — Notion → MariaDB

**Streams:** 3 | **Destination:** MariaDB

> dbt SQL identical to `21-notion-to-postgres.md`.

---

## Connections
```json
{ "credentials": { "auth_type": "token", "token": "secret_..." } }
{ "host":"..","port":3306,"database":"analytics","username":"writer","password":"..." }
```

---

## All 3 Stream DDLs (MariaDB)

```sql
CREATE TABLE analytics.notion_databases (
    db_id       VARCHAR(50)  PRIMARY KEY,
    db_title    VARCHAR(500),
    db_url      LONGTEXT,
    created_on  DATE,
    modified_on DATE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE analytics.notion_pages (
    page_id     VARCHAR(50)  PRIMARY KEY,
    page_url    LONGTEXT,
    status      VARCHAR(100),
    priority    VARCHAR(100),
    assignee    VARCHAR(255),
    due_date    DATE,
    created_on  DATE,
    modified_on DATE
) ENGINE=InnoDB;

CREATE TABLE analytics.notion_users (
    user_id      VARCHAR(50)  PRIMARY KEY,
    display_name VARCHAR(255),
    user_type    VARCHAR(50),
    email        VARCHAR(255),
    avatar_url   LONGTEXT
) ENGINE=InnoDB;
```
