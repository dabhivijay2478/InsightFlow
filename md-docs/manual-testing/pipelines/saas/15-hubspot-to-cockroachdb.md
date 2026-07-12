# Pipeline 15 — HubSpot → CockroachDB

> Retired for the HubSpot beta. The supported connector path is dlt-based
> HubSpot → existing PostgreSQL only. This file is retained as historical
> reference and must not be used as an acceptance test.

**Streams:** 14 | **Destination:** CockroachDB

> dbt SQL identical to `11-hubspot-to-postgres.md`.

---

## Connections
```json
{ "credentials": { "client_id":"..","client_secret":"..","refresh_token":"..","access_token":"..","token_expires_in":1800 } }
{ "host":"..","port":26257,"database":"defaultdb","username":"root","password":"","ssl_mode":"disable" }
```
```sql
CREATE SCHEMA IF NOT EXISTS analytics;
```

---

## All 14 Stream DDLs (CockroachDB)

```sql
CREATE TABLE IF NOT EXISTS analytics.hs_contacts (
    contact_id STRING PRIMARY KEY, email STRING, full_name STRING, phone STRING,
    lead_status STRING, lifecycle STRING, owner_id STRING,
    created_on DATE, updated_on DATE
);
CREATE TABLE IF NOT EXISTS analytics.hs_companies (
    company_id STRING PRIMARY KEY, company_name STRING, website STRING, sector STRING,
    city STRING, country STRING, annual_rev DECIMAL(15,2), headcount INT8, created_on DATE
);
CREATE TABLE IF NOT EXISTS analytics.hs_deals (
    deal_id STRING PRIMARY KEY, deal_name STRING,
    deal_value DECIMAL(12,2), stage STRING, pipeline STRING,
    close_date DATE, owner_id STRING, days_to_close INT8
);
CREATE TABLE IF NOT EXISTS analytics.hs_owners (
    owner_id STRING PRIMARY KEY, first_name STRING, last_name STRING,
    email STRING, user_id INT8, created_on DATE
);
CREATE TABLE IF NOT EXISTS analytics.hs_tickets (
    ticket_id STRING PRIMARY KEY, ticket_name STRING, category STRING, priority STRING,
    ticket_stage STRING, owner_id STRING, created_on DATE, closed_on DATE
);
CREATE TABLE IF NOT EXISTS analytics.hs_calls (
    call_id STRING PRIMARY KEY, title STRING, call_status STRING,
    direction STRING, duration_ms INT8, called_at TIMESTAMPTZ
);
CREATE TABLE IF NOT EXISTS analytics.hs_emails (
    email_id STRING PRIMARY KEY, subject STRING, direction STRING,
    send_status STRING, sent_at TIMESTAMPTZ
);
CREATE TABLE IF NOT EXISTS analytics.hs_meetings (
    meeting_id STRING PRIMARY KEY, title STRING, outcome STRING,
    started_at TIMESTAMPTZ, ended_at TIMESTAMPTZ, duration_mins INT8
);
CREATE TABLE IF NOT EXISTS analytics.hs_notes (
    note_id STRING PRIMARY KEY, body STRING, owner_id STRING, created_on DATE
);
CREATE TABLE IF NOT EXISTS analytics.hs_tasks (
    task_id STRING PRIMARY KEY, subject STRING, task_type STRING,
    status STRING, priority STRING, due_on DATE, created_on DATE
);
CREATE TABLE IF NOT EXISTS analytics.hs_nps (
    submission_id STRING PRIMARY KEY, survey_type STRING, rating STRING,
    nps_score INT8, comment STRING, submitted_at TIMESTAMPTZ
);
CREATE TABLE IF NOT EXISTS analytics.hs_pipelines (
    pipeline_id STRING PRIMARY KEY, pipeline_label STRING, object_type STRING,
    is_default BOOL, created_on DATE
);
CREATE TABLE IF NOT EXISTS analytics.hs_pipeline_stages (
    stage_id STRING PRIMARY KEY, pipeline_ref STRING, stage_label STRING,
    display_order INT8, is_closed BOOL, probability DECIMAL(5,4)
);
CREATE TABLE IF NOT EXISTS analytics.hs_associations (
    assoc_id STRING PRIMARY KEY, from_id STRING, to_id STRING,
    assoc_type STRING, created_on DATE
);
```

---

## Verify
```sql
SELECT is_default FROM analytics.hs_pipelines LIMIT 5;    -- true/false
SELECT deal_value FROM analytics.hs_deals LIMIT 5;        -- decimal
SELECT deal_id, COUNT(*) FROM analytics.hs_deals GROUP BY deal_id HAVING COUNT(*)>1;
```
