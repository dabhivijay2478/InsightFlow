# Pipeline 14 — HubSpot → SQLite

> Retired for the HubSpot beta. The supported connector path is dlt-based
> HubSpot → existing PostgreSQL only. This file is retained as historical
> reference and must not be used as an acceptance test.

**Streams:** 14 | **Destination:** SQLite

> dbt SQL identical to `11-hubspot-to-postgres.md`.

---

## Connections
```json
{ "credentials": { "client_id":"..","client_secret":"..","refresh_token":"..","access_token":"..","token_expires_in":1800 } }
{ "database": "/absolute/path/to/analytics.db" }
```

---

## All 14 Stream DDLs (SQLite)

```sql
CREATE TABLE IF NOT EXISTS hs_contacts (
    contact_id TEXT PRIMARY KEY, email TEXT, full_name TEXT, phone TEXT,
    lead_status TEXT, lifecycle TEXT, owner_id TEXT,
    created_on TEXT, updated_on TEXT
);
CREATE TABLE IF NOT EXISTS hs_companies (
    company_id TEXT PRIMARY KEY, company_name TEXT, website TEXT, sector TEXT,
    city TEXT, country TEXT, annual_rev REAL, headcount INTEGER, created_on TEXT
);
CREATE TABLE IF NOT EXISTS hs_deals (
    deal_id TEXT PRIMARY KEY, deal_name TEXT,
    deal_value REAL, stage TEXT, pipeline TEXT,
    close_date TEXT, owner_id TEXT, days_to_close INTEGER
);
CREATE TABLE IF NOT EXISTS hs_owners (
    owner_id TEXT PRIMARY KEY, first_name TEXT, last_name TEXT,
    email TEXT, user_id INTEGER, created_on TEXT
);
CREATE TABLE IF NOT EXISTS hs_tickets (
    ticket_id TEXT PRIMARY KEY, ticket_name TEXT, category TEXT, priority TEXT,
    ticket_stage TEXT, owner_id TEXT, created_on TEXT, closed_on TEXT
);
CREATE TABLE IF NOT EXISTS hs_calls (
    call_id TEXT PRIMARY KEY, title TEXT, call_status TEXT,
    direction TEXT, duration_ms INTEGER, called_at TEXT
);
CREATE TABLE IF NOT EXISTS hs_emails (
    email_id TEXT PRIMARY KEY, subject TEXT, direction TEXT,
    send_status TEXT, sent_at TEXT
);
CREATE TABLE IF NOT EXISTS hs_meetings (
    meeting_id TEXT PRIMARY KEY, title TEXT, outcome TEXT,
    started_at TEXT, ended_at TEXT, duration_mins INTEGER
);
CREATE TABLE IF NOT EXISTS hs_notes (
    note_id TEXT PRIMARY KEY, body TEXT, owner_id TEXT, created_on TEXT
);
CREATE TABLE IF NOT EXISTS hs_tasks (
    task_id TEXT PRIMARY KEY, subject TEXT, task_type TEXT,
    status TEXT, priority TEXT, due_on TEXT, created_on TEXT
);
CREATE TABLE IF NOT EXISTS hs_nps (
    submission_id TEXT PRIMARY KEY, survey_type TEXT, rating TEXT,
    nps_score INTEGER, comment TEXT, submitted_at TEXT
);
CREATE TABLE IF NOT EXISTS hs_pipelines (
    pipeline_id TEXT PRIMARY KEY, pipeline_label TEXT, object_type TEXT,
    is_default INTEGER, created_on TEXT
);
CREATE TABLE IF NOT EXISTS hs_pipeline_stages (
    stage_id TEXT PRIMARY KEY, pipeline_ref TEXT, stage_label TEXT,
    display_order INTEGER, is_closed INTEGER, probability REAL
);
CREATE TABLE IF NOT EXISTS hs_associations (
    assoc_id TEXT PRIMARY KEY, from_id TEXT, to_id TEXT,
    assoc_type TEXT, created_on TEXT
);
```

---

## Verify
```bash
DB=/absolute/path/to/analytics.db
sqlite3 $DB "SELECT typeof(deal_value), deal_value FROM hs_deals LIMIT 3;"   # real
sqlite3 $DB "SELECT is_default FROM hs_pipelines LIMIT 5;"                   # 0 or 1
sqlite3 $DB "SELECT called_at FROM hs_calls LIMIT 3;"                        # ISO string
```
