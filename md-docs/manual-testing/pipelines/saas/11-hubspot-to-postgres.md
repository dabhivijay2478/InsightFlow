# Pipeline 11 — HubSpot → PostgreSQL

**Streams:** 14 | **Destination:** PostgreSQL

---

## Connections
```json
{ "credentials": { "client_id":"..","client_secret":"..","refresh_token":"..","access_token":"..","token_expires_in":1800 } }
{ "host":"..","port":5432,"database":"analytics","username":"writer","password":"..","ssl_mode":"disable" }
```
```sql
CREATE SCHEMA IF NOT EXISTS analytics;
```

---

## Stream 1 — `hubspot.contacts` → `analytics.hs_contacts`

### Step 1 — DDL
```sql
CREATE TABLE analytics.hs_contacts (
    contact_id   TEXT        PRIMARY KEY,    -- source: id
    email        TEXT,
    full_name    TEXT,                        -- derived: CONCAT(firstname,' ',lastname)
    phone        TEXT,
    lead_status  TEXT,                        -- source: hs_lead_status (hs_ stripped)
    lifecycle    TEXT,                        -- source: lifecyclestage
    owner_id     TEXT,                        -- source: hubspot_owner_id
    created_on   DATE,
    updated_on   DATE
);
```
### Step 4 — Normalisation
Rename `hs_lead_status` → `lead_status`, `lifecyclestage` → `lifecycle`, `hubspot_owner_id` → `owner_id`
### Step 5 — dbt SQL
```sql
SELECT
    id                                                           AS contact_id,
    email,
    TRIM(CONCAT(COALESCE(firstname,''),' ',COALESCE(lastname,''))) AS full_name,
    phone,
    lead_status, lifecycle, owner_id,
    CAST(createdate AS DATE)       AS created_on,
    CAST(lastmodifieddate AS DATE) AS updated_on
FROM {{ source('raw', 'hubspot__contacts') }}
WHERE email IS NOT NULL
```
### Step 8 — Verify
```sql
SELECT contact_id, full_name, lead_status FROM analytics.hs_contacts LIMIT 5;
-- full_name: single string; hs_lead_status must NOT appear
```

---

## Stream 2 — `hubspot.companies` → `analytics.hs_companies`

### Step 1 — DDL
```sql
CREATE TABLE analytics.hs_companies (
    company_id   TEXT PRIMARY KEY,
    company_name TEXT,                        -- source: name
    website      TEXT,                        -- source: domain
    sector       TEXT,                        -- source: industry
    city         TEXT, country TEXT,
    annual_rev   NUMERIC(15,2),               -- source: annualrevenue TEXT → NUMERIC
    headcount    INTEGER,                     -- source: numberofemployees TEXT → INT
    created_on   DATE
);
```
### Step 5 — dbt SQL
```sql
SELECT
    id AS company_id, name AS company_name, domain AS website,
    industry AS sector, city, country,
    CAST(NULLIF(annualrevenue,'') AS NUMERIC)     AS annual_rev,
    CAST(NULLIF(numberofemployees,'') AS INTEGER) AS headcount,
    CAST(createdate AS DATE)                      AS created_on
FROM {{ source('raw', 'hubspot__companies') }}
WHERE name IS NOT NULL
```

---

## Stream 3 — `hubspot.deals` → `analytics.hs_deals`

### Step 1 — DDL
```sql
CREATE TABLE analytics.hs_deals (
    deal_id       TEXT PRIMARY KEY,
    deal_name     TEXT,
    deal_value    NUMERIC(12,2),              -- source: amount TEXT → NUMERIC
    stage         TEXT,                       -- source: dealstage
    pipeline      TEXT,
    close_date    DATE,
    owner_id      TEXT,
    days_to_close INTEGER
);
```
### Step 5 — dbt SQL
```sql
SELECT
    id AS deal_id, dealname AS deal_name,
    CAST(NULLIF(amount,'') AS NUMERIC) AS deal_value,
    dealstage AS stage, pipeline,
    closedate::DATE AS close_date,
    hubspot_owner_id AS owner_id,
    CASE WHEN closedate IS NOT NULL AND createdate IS NOT NULL
         THEN DATEDIFF('day', CAST(createdate AS DATE), CAST(closedate AS DATE))
         ELSE NULL END AS days_to_close
FROM {{ source('raw', 'hubspot__deals') }}
WHERE amount IS NOT NULL AND amount != ''
```

---

## Stream 4 — `hubspot.owners` → `analytics.hs_owners`

### Step 1 — DDL
```sql
CREATE TABLE analytics.hs_owners (
    owner_id     TEXT PRIMARY KEY,
    first_name   TEXT, last_name TEXT,
    email        TEXT,
    user_id      BIGINT,
    created_on   DATE
);
```
### Step 5 — dbt SQL
```sql
SELECT
    id AS owner_id, firstName AS first_name, lastName AS last_name,
    email, userId AS user_id,
    CAST(createdAt AS DATE) AS created_on
FROM {{ source('raw', 'hubspot__owners') }}
```

---

## Stream 5 — `hubspot.tickets` → `analytics.hs_tickets`

### Step 1 — DDL
```sql
CREATE TABLE analytics.hs_tickets (
    ticket_id    TEXT PRIMARY KEY,
    ticket_name  TEXT,
    category     TEXT,                        -- source: hs_ticket_category
    priority     TEXT,                        -- source: hs_ticket_priority
    ticket_stage TEXT,
    owner_id     TEXT,
    created_on   DATE, closed_on DATE
);
```
### Step 5 — dbt SQL
```sql
SELECT
    id AS ticket_id, subject AS ticket_name,
    hs_ticket_category AS category,
    hs_ticket_priority AS priority,
    hs_pipeline_stage  AS ticket_stage,
    hubspot_owner_id   AS owner_id,
    CAST(createdate AS DATE) AS created_on,
    CAST(closed_date AS DATE) AS closed_on
FROM {{ source('raw', 'hubspot__tickets') }}
```

---

## Stream 6 — `hubspot.calls` → `analytics.hs_calls`

### Step 1 — DDL
```sql
CREATE TABLE analytics.hs_calls (
    call_id     TEXT PRIMARY KEY,
    title       TEXT,
    call_status TEXT,
    direction   TEXT,
    duration_ms INTEGER,                      -- source: hs_call_duration TEXT → INT
    called_at   TIMESTAMPTZ
);
```
### Step 5 — dbt SQL
```sql
SELECT
    id AS call_id, hs_call_title AS title,
    hs_call_status AS call_status,
    hs_call_direction AS direction,
    CAST(NULLIF(hs_call_duration,'') AS INTEGER) AS duration_ms,
    CAST(hs_timestamp AS TIMESTAMPTZ) AS called_at
FROM {{ source('raw', 'hubspot__calls') }}
```

---

## Stream 7 — `hubspot.emails` → `analytics.hs_emails`

### Step 1 — DDL
```sql
CREATE TABLE analytics.hs_emails (
    email_id    TEXT PRIMARY KEY,
    subject     TEXT,
    direction   TEXT,                         -- source: hs_email_direction
    send_status TEXT,                         -- source: hs_email_status
    sent_at     TIMESTAMPTZ
);
```
### Step 5 — dbt SQL
```sql
SELECT
    id AS email_id, hs_email_subject AS subject,
    hs_email_direction AS direction,
    hs_email_status AS send_status,
    CAST(hs_timestamp AS TIMESTAMPTZ) AS sent_at
FROM {{ source('raw', 'hubspot__emails') }}
```

---

## Stream 8 — `hubspot.meetings` → `analytics.hs_meetings`

### Step 1 — DDL
```sql
CREATE TABLE analytics.hs_meetings (
    meeting_id TEXT PRIMARY KEY,
    title      TEXT,
    outcome    TEXT,
    started_at TIMESTAMPTZ,
    ended_at   TIMESTAMPTZ,
    duration_mins INTEGER                     -- derived: DATEDIFF minutes
);
```
### Step 5 — dbt SQL
```sql
SELECT
    id AS meeting_id, hs_meeting_title AS title,
    hs_meeting_outcome AS outcome,
    CAST(hs_meeting_start_time AS TIMESTAMPTZ) AS started_at,
    CAST(hs_meeting_end_time   AS TIMESTAMPTZ) AS ended_at,
    CASE WHEN hs_meeting_end_time IS NOT NULL AND hs_meeting_start_time IS NOT NULL
         THEN DATEDIFF('minute',
              CAST(hs_meeting_start_time AS TIMESTAMPTZ),
              CAST(hs_meeting_end_time AS TIMESTAMPTZ))
         ELSE NULL END AS duration_mins
FROM {{ source('raw', 'hubspot__meetings') }}
```

---

## Stream 9 — `hubspot.notes` → `analytics.hs_notes`

### Step 1 — DDL
```sql
CREATE TABLE analytics.hs_notes (
    note_id    TEXT PRIMARY KEY,
    body       TEXT,                          -- source: hs_note_body
    owner_id   TEXT,
    created_on DATE
);
```
### Step 5 — dbt SQL
```sql
SELECT
    id AS note_id, hs_note_body AS body,
    hubspot_owner_id AS owner_id,
    CAST(hs_timestamp AS DATE) AS created_on
FROM {{ source('raw', 'hubspot__notes') }}
```

---

## Stream 10 — `hubspot.tasks` → `analytics.hs_tasks`

### Step 1 — DDL
```sql
CREATE TABLE analytics.hs_tasks (
    task_id    TEXT PRIMARY KEY,
    subject    TEXT,
    task_type  TEXT,                          -- source: hs_task_type
    status     TEXT,
    priority   TEXT,
    due_on     DATE,                          -- source: hs_task_due_date → DATE
    created_on DATE
);
```
### Step 5 — dbt SQL
```sql
SELECT
    id AS task_id, hs_task_subject AS subject,
    hs_task_type AS task_type,
    hs_task_status AS status,
    hs_task_priority AS priority,
    CAST(hs_task_due_date AS DATE) AS due_on,
    CAST(createdate AS DATE) AS created_on
FROM {{ source('raw', 'hubspot__tasks') }}
```

---

## Stream 11 — `hubspot.feedback_submissions` → `analytics.hs_nps`

### Step 1 — DDL
```sql
CREATE TABLE analytics.hs_nps (
    submission_id TEXT PRIMARY KEY,
    survey_type   TEXT,
    rating        TEXT,                       -- from JSON: hs_response->>'rating'
    nps_score     INTEGER,                    -- from JSON: hs_response->>'nps_score' → INT
    comment       TEXT,                       -- from JSON: hs_response->>'comment'
    submitted_at  TIMESTAMPTZ
);
```
### Step 5 — dbt SQL
```sql
SELECT
    id AS submission_id, hs_survey_type AS survey_type,
    hs_response->>'rating'                                    AS rating,
    CAST(NULLIF(hs_response->>'nps_score','') AS INTEGER)     AS nps_score,
    NULLIF(TRIM(hs_response->>'comment'),'')                  AS comment,
    CAST(hs_submission_timestamp AS TIMESTAMPTZ)              AS submitted_at
FROM {{ source('raw', 'hubspot__feedback_submissions') }}
WHERE hs_response IS NOT NULL
```

---

## Stream 12 — `hubspot.pipelines` → `analytics.hs_pipelines`

### Step 1 — DDL
```sql
CREATE TABLE analytics.hs_pipelines (
    pipeline_id    TEXT PRIMARY KEY,
    pipeline_label TEXT,                      -- source: label
    object_type    TEXT,
    is_default     BOOLEAN,
    created_on     DATE
);
```
### Step 5 — dbt SQL
```sql
SELECT
    id AS pipeline_id, label AS pipeline_label,
    objectType AS object_type,
    default AS is_default,
    CAST(createdAt AS DATE) AS created_on
FROM {{ source('raw', 'hubspot__pipelines') }}
```

---

## Stream 13 — `hubspot.pipeline_stages` → `analytics.hs_pipeline_stages`

### Step 1 — DDL
```sql
CREATE TABLE analytics.hs_pipeline_stages (
    stage_id     TEXT PRIMARY KEY,
    pipeline_ref TEXT,
    stage_label  TEXT,                        -- source: label
    display_order INTEGER,
    is_closed    BOOLEAN,
    probability  NUMERIC(5,4)                 -- source: probability TEXT → NUMERIC
);
```
### Step 5 — dbt SQL
```sql
SELECT
    id AS stage_id, pipelineId AS pipeline_ref,
    label AS stage_label, displayOrder AS display_order,
    closed AS is_closed,
    CAST(NULLIF(probability,'') AS NUMERIC) AS probability
FROM {{ source('raw', 'hubspot__pipeline_stages') }}
```

---

## Stream 14 — `hubspot.associations` → `analytics.hs_associations`

### Step 1 — DDL
```sql
CREATE TABLE analytics.hs_associations (
    assoc_id      TEXT PRIMARY KEY,
    from_id       TEXT,
    to_id         TEXT,
    assoc_type    TEXT,                       -- source: type
    created_on    DATE
);
```
### Step 5 — dbt SQL
```sql
SELECT
    id AS assoc_id,
    fromObjectId AS from_id,
    toObjectId   AS to_id,
    type         AS assoc_type,
    CAST(createdAt AS DATE) AS created_on
FROM {{ source('raw', 'hubspot__associations') }}
```

---

## Edge Cases

| Scenario | Expected |
|---------|---------|
| `amount = ''` in deals | `CAST(NULLIF(amount,'') AS NUMERIC)` = NULL |
| `hs_response` NULL in feedback | WHERE filters it out |
| `nps_score` non-numeric string | CAST fails → NULL with NULLIF guard |
| hs_ prefix column missing rename rule | Column appears as `hs_lead_status` in dest — fix: add normalisation rule |
| `hubspot_owner_id` renamed to `owner_id` | Both layers (normalisation + dbt) rename correctly |
