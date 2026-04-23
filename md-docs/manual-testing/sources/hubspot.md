# HubSpot Source — Manual Testing Guide

**Streams:** 14  
**Credential:** Private app access token  
**DuckDB prefix:** `hubspot__`

---

## Credential Setup

```json
{ "access_token": "pat-na1-..." }
```

---

## All 14 Streams

| Stream | DuckDB staging name | Key columns | INCREMENTAL key |
|--------|-------------------|-------------|----------------|
| `hubspot.contacts` | `hubspot__contacts` | `id`, `email`, `firstname`, `lastname`, `phone`, `hs_lead_status`, `createdate`, `lastmodifieddate` | `lastmodifieddate` |
| `hubspot.companies` | `hubspot__companies` | `id`, `name`, `domain`, `industry`, `city`, `country`, `createdate`, `hs_lastmodifieddate` | `hs_lastmodifieddate` |
| `hubspot.deals` | `hubspot__deals` | `id`, `dealname`, `amount`, `dealstage`, `pipeline`, `closedate`, `createdate` | `hs_lastmodifieddate` |
| `hubspot.tickets` | `hubspot__tickets` | `id`, `subject`, `hs_pipeline_stage`, `hs_pipeline`, `createdate`, `hs_lastmodifieddate` | `hs_lastmodifieddate` |
| `hubspot.products` | `hubspot__products` | `id`, `name`, `price`, `description`, `createdate`, `hs_lastmodifieddate` | `hs_lastmodifieddate` |
| `hubspot.line_items` | `hubspot__line_items` | `id`, `name`, `quantity`, `amount`, `price`, `hs_product_id`, `createdate` | `hs_lastmodifieddate` |
| `hubspot.quotes` | `hubspot__quotes` | `id`, `hs_title`, `hs_status`, `hs_expiration_date`, `hs_amount_billed_to_company`, `createdate` | `hs_lastmodifieddate` |
| `hubspot.calls` | `hubspot__calls` | `id`, `hs_call_title`, `hs_call_status`, `hs_call_duration`, `hs_timestamp`, `createdate` | `hs_lastmodifieddate` |
| `hubspot.emails` | `hubspot__emails` | `id`, `hs_email_subject`, `hs_email_status`, `hs_email_direction`, `hs_timestamp`, `createdate` | `hs_lastmodifieddate` |
| `hubspot.meetings` | `hubspot__meetings` | `id`, `hs_meeting_title`, `hs_meeting_outcome`, `hs_timestamp`, `hs_meeting_start_time`, `createdate` | `hs_lastmodifieddate` |
| `hubspot.notes` | `hubspot__notes` | `id`, `hs_note_body`, `hs_timestamp`, `createdate` | `hs_lastmodifieddate` |
| `hubspot.tasks` | `hubspot__tasks` | `id`, `hs_task_subject`, `hs_task_status`, `hs_task_priority`, `hs_timestamp`, `createdate` | `hs_lastmodifieddate` |
| `hubspot.feedback_submissions` | `hubspot__feedback_submissions` | `id`, `hs_survey_type`, `hs_response`, `hs_submission_timestamp`, `createdate` | `hs_lastmodifieddate` |
| `hubspot.owners` | `hubspot__owners` | `id`, `email`, `firstName`, `lastName`, `userId`, `createdAt`, `updatedAt` | `updatedAt` |

---

## Scenario S-HS-1 — Full Table Sync: `contacts`

**Destination DDL:**
```sql
CREATE TABLE analytics.hubspot_contacts_hd (
    id             TEXT PRIMARY KEY,
    email          TEXT,
    firstname      TEXT,
    lastname       TEXT,
    hs_lead_status TEXT,
    createdate     TIMESTAMPTZ
);
```

**dbt SQL:**
```sql
SELECT
    id,
    email,
    firstname,
    lastname,
    hs_lead_status,
    createdate
FROM {{ source('raw', 'hubspot__contacts') }}
WHERE email IS NOT NULL
```

**Verify:** `SELECT COUNT(*) FROM analytics.hubspot_contacts_hd;` → matches HubSpot contact count.

---

## Scenario S-HS-2 — Incremental Sync: `deals`

**Sync mode:** `INCREMENTAL`, replication key `hs_lastmodifieddate`

**dbt SQL:**
```sql
SELECT
    id,
    dealname,
    CAST(amount AS NUMERIC)  AS deal_amount,
    dealstage,
    pipeline,
    closedate,
    createdate
FROM {{ source('raw', 'hubspot__deals') }}
WHERE amount IS NOT NULL
```

**Run 1:** All deals.  
**Run 2:** Only deals modified since last run.

---

## Scenario S-HS-3 — Deal Pipeline Stage Aggregate

**dbt SQL:**
```sql
SELECT
    pipeline,
    dealstage,
    COUNT(*)               AS deal_count,
    SUM(CAST(amount AS NUMERIC)) AS total_value
FROM {{ source('raw', 'hubspot__deals') }}
WHERE amount IS NOT NULL
GROUP BY pipeline, dealstage
ORDER BY total_value DESC
```

**Destination DDL (no PK — append):**
```sql
CREATE TABLE analytics.deal_pipeline_summary (
    pipeline    TEXT,
    dealstage   TEXT,
    deal_count  BIGINT,
    total_value NUMERIC
);
```

**Expected:** `no_pk_warnings` in callback.

---

## Scenario S-HS-4 — Activity Streams (calls + emails + meetings)

Three streams in one pipeline:

**Model 1 — calls:**
```sql
SELECT
    id,
    hs_call_title     AS title,
    hs_call_status    AS status,
    hs_call_duration  AS duration_ms,
    hs_timestamp      AS activity_at
FROM {{ source('raw', 'hubspot__calls') }}
```

**Model 2 — emails:**
```sql
SELECT
    id,
    hs_email_subject    AS subject,
    hs_email_status     AS status,
    hs_email_direction  AS direction,
    hs_timestamp        AS activity_at
FROM {{ source('raw', 'hubspot__emails') }}
```

**Model 3 — meetings:**
```sql
SELECT
    id,
    hs_meeting_title        AS title,
    hs_meeting_outcome      AS outcome,
    hs_meeting_start_time   AS started_at
FROM {{ source('raw', 'hubspot__meetings') }}
```

---

## Scenario S-HS-5 — Normalisation: Rename HubSpot prefixed columns

HubSpot prefixes many columns with `hs_`. Strip the prefix for a cleaner destination schema.

**Rules:**
```json
[
  { "rule_type": "rename", "table": "hubspot.tickets", "column": "hs_pipeline_stage", "destination_name": "pipeline_stage" },
  { "rule_type": "rename", "table": "hubspot.tickets", "column": "hs_pipeline",       "destination_name": "pipeline" },
  { "rule_type": "rename", "table": "hubspot.tickets", "column": "hs_lastmodifieddate","destination_name": "updated_at" }
]
```

**dbt SQL:**
```sql
SELECT id, subject, pipeline_stage, pipeline, updated_at
FROM {{ source('raw', 'hubspot__tickets') }}
```

---

## Scenario S-HS-6 — Feedback Submissions JSON Response

`feedback_submissions.hs_response` is a JSON string. Extract rating:

**dbt SQL:**
```sql
SELECT
    id,
    hs_survey_type                       AS survey_type,
    hs_response->>'rating'               AS rating,
    hs_response->>'comment'              AS comment,
    hs_submission_timestamp              AS submitted_at
FROM {{ source('raw', 'hubspot__feedback_submissions') }}
WHERE hs_response IS NOT NULL
```

---

## Scenario S-HS-7 — Owners Master Table

**dbt SQL:**
```sql
SELECT
    id,
    email,
    "firstName"  AS first_name,
    "lastName"   AS last_name,
    "userId"     AS user_id,
    "createdAt"  AS created_at,
    "updatedAt"  AS updated_at
FROM {{ source('raw', 'hubspot__owners') }}
```

---

## All 14 Streams — Quick Smoke Test Checklist

| Stream | DuckDB ref | Expected rows |
|--------|-----------|---------------|
| contacts | `hubspot__contacts` | ≥ 1 |
| companies | `hubspot__companies` | ≥ 1 |
| deals | `hubspot__deals` | ≥ 0 |
| tickets | `hubspot__tickets` | ≥ 0 |
| products | `hubspot__products` | ≥ 0 |
| line_items | `hubspot__line_items` | ≥ 0 |
| quotes | `hubspot__quotes` | ≥ 0 |
| calls | `hubspot__calls` | ≥ 0 |
| emails | `hubspot__emails` | ≥ 0 |
| meetings | `hubspot__meetings` | ≥ 0 |
| notes | `hubspot__notes` | ≥ 0 |
| tasks | `hubspot__tasks` | ≥ 0 |
| feedback_submissions | `hubspot__feedback_submissions` | ≥ 0 |
| owners | `hubspot__owners` | ≥ 1 |
