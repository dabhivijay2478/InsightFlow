# Pipeline 12 — HubSpot → MySQL

**Streams:** 14 | **Destination:** MySQL

> dbt SQL identical to `11-hubspot-to-postgres.md`.

---

## Connections
```json
{ "credentials": { "client_id":"..","client_secret":"..","refresh_token":"..","access_token":"..","token_expires_in":1800 } }
{ "host":"..","port":3306,"database":"analytics","username":"writer","password":"..." }
```

---

## All 14 Stream DDLs (MySQL)

```sql
CREATE TABLE analytics.hs_contacts (
    contact_id  VARCHAR(255) PRIMARY KEY,
    email VARCHAR(255), full_name VARCHAR(500), phone VARCHAR(50),
    lead_status VARCHAR(100), lifecycle VARCHAR(100), owner_id VARCHAR(255),
    created_on DATE, updated_on DATE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE analytics.hs_companies (
    company_id  VARCHAR(255) PRIMARY KEY,
    company_name VARCHAR(500), website VARCHAR(255), sector VARCHAR(255),
    city VARCHAR(255), country VARCHAR(100),
    annual_rev DECIMAL(15,2), headcount INT, created_on DATE
) ENGINE=InnoDB;

CREATE TABLE analytics.hs_deals (
    deal_id   VARCHAR(255) PRIMARY KEY,
    deal_name VARCHAR(500),
    deal_value DECIMAL(12,2), stage VARCHAR(100), pipeline VARCHAR(255),
    close_date DATE, owner_id VARCHAR(255), days_to_close INT
) ENGINE=InnoDB;

CREATE TABLE analytics.hs_owners (
    owner_id   VARCHAR(255) PRIMARY KEY,
    first_name VARCHAR(255), last_name VARCHAR(255),
    email VARCHAR(255), user_id BIGINT, created_on DATE
) ENGINE=InnoDB;

CREATE TABLE analytics.hs_tickets (
    ticket_id   VARCHAR(255) PRIMARY KEY,
    ticket_name VARCHAR(500), category VARCHAR(100), priority VARCHAR(50),
    ticket_stage VARCHAR(100), owner_id VARCHAR(255),
    created_on DATE, closed_on DATE
) ENGINE=InnoDB;

CREATE TABLE analytics.hs_calls (
    call_id     VARCHAR(255) PRIMARY KEY,
    title       VARCHAR(500), call_status VARCHAR(50),
    direction   VARCHAR(50),  duration_ms INT, called_at DATETIME
) ENGINE=InnoDB;

CREATE TABLE analytics.hs_emails (
    email_id    VARCHAR(255) PRIMARY KEY,
    subject     VARCHAR(500), direction VARCHAR(50),
    send_status VARCHAR(50),  sent_at DATETIME
) ENGINE=InnoDB;

CREATE TABLE analytics.hs_meetings (
    meeting_id    VARCHAR(255) PRIMARY KEY,
    title         VARCHAR(500), outcome VARCHAR(100),
    started_at    DATETIME, ended_at DATETIME, duration_mins INT
) ENGINE=InnoDB;

CREATE TABLE analytics.hs_notes (
    note_id   VARCHAR(255) PRIMARY KEY,
    body      TEXT, owner_id VARCHAR(255), created_on DATE
) ENGINE=InnoDB;

CREATE TABLE analytics.hs_tasks (
    task_id   VARCHAR(255) PRIMARY KEY,
    subject   VARCHAR(500), task_type VARCHAR(100),
    status    VARCHAR(50),  priority VARCHAR(50),
    due_on    DATE,         created_on DATE
) ENGINE=InnoDB;

CREATE TABLE analytics.hs_nps (
    submission_id VARCHAR(255) PRIMARY KEY,
    survey_type   VARCHAR(100), rating VARCHAR(50),
    nps_score     INT, comment TEXT, submitted_at DATETIME
) ENGINE=InnoDB;

CREATE TABLE analytics.hs_pipelines (
    pipeline_id    VARCHAR(255) PRIMARY KEY,
    pipeline_label VARCHAR(255), object_type VARCHAR(100),
    is_default     TINYINT(1),   created_on DATE
) ENGINE=InnoDB;

CREATE TABLE analytics.hs_pipeline_stages (
    stage_id      VARCHAR(255) PRIMARY KEY,
    pipeline_ref  VARCHAR(255), stage_label VARCHAR(255),
    display_order INT,          is_closed TINYINT(1),
    probability   DECIMAL(5,4)
) ENGINE=InnoDB;

CREATE TABLE analytics.hs_associations (
    assoc_id   VARCHAR(255) PRIMARY KEY,
    from_id    VARCHAR(255), to_id VARCHAR(255),
    assoc_type VARCHAR(100), created_on DATE
) ENGINE=InnoDB;
```

---

## Verify
```sql
SELECT is_default FROM analytics.hs_pipelines;          -- 0 or 1
SELECT deal_value FROM analytics.hs_deals LIMIT 5;      -- decimal
SELECT called_at FROM analytics.hs_calls LIMIT 3;       -- DATETIME
SELECT contact_id, COUNT(*) cnt FROM analytics.hs_contacts GROUP BY contact_id HAVING cnt>1;
```
