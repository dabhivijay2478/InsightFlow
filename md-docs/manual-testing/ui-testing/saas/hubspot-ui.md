# HubSpot Source — UI Testing (All 14 Streams × 5 Destinations)

> Universal builder steps in `builder-walkthrough.md`.

---

## Phase 1 — Source Panel (HubSpot)

### Credential Fields
| Field | Value |
|-------|-------|
| **Client ID** | `xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx` |
| **Client Secret** | `xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx` |
| **Refresh Token** | `na1-xxxx-...` |
| **Access Token** | `eyJ...` (auto-refreshed) |
| **Token Expires In** | `1800` (seconds) |

**Test Connection → ✅**
- ❌ `401`: expired refresh token — re-authenticate OAuth
- ❌ `403`: missing scope — check HubSpot app scopes include `crm.objects.contacts.read` etc.

---

## Step 2b — Stream Selection & Sync Mode

| Stream | Sync Mode | Cursor Field |
|--------|-----------|-------------|
| `contacts` | INCREMENTAL | `lastmodifieddate` |
| `companies` | INCREMENTAL | `hs_lastmodifieddate` |
| `deals` | INCREMENTAL | `hs_lastmodifieddate` |
| `owners` | FULL TABLE | — |
| `tickets` | INCREMENTAL | `hs_lastmodifieddate` |
| `calls` | INCREMENTAL | `hs_timestamp` |
| `emails` | INCREMENTAL | `hs_timestamp` |
| `meetings` | INCREMENTAL | `hs_timestamp` |
| `notes` | INCREMENTAL | `hs_timestamp` |
| `tasks` | INCREMENTAL | `hs_lastmodifieddate` |
| `feedback_submissions` | INCREMENTAL | `hs_submission_timestamp` |
| `pipelines` | FULL TABLE | — |
| `pipeline_stages` | FULL TABLE | — |
| `associations` | FULL TABLE | — |

---

## Phase 2 — Stream→Table Mapping

| Stream | Table |
|--------|-------|
| contacts | `hs_contacts` |
| companies | `hs_companies` |
| deals | `hs_deals` |
| owners | `hs_owners` |
| tickets | `hs_tickets` |
| calls | `hs_calls` |
| emails | `hs_emails` |
| meetings | `hs_meetings` |
| notes | `hs_notes` |
| tasks | `hs_tasks` |
| feedback_submissions | `hs_nps` |
| pipelines | `hs_pipelines` |
| pipeline_stages | `hs_pipeline_stages` |
| associations | `hs_associations` |

---

## Phase 3 — Normalisation Rules

### `hubspot.contacts`
| Rule | Column | Target |
|------|--------|--------|
| Rename | `hs_lead_status` | `lead_status` |
| Rename | `lifecyclestage` | `lifecycle` |
| Rename | `hubspot_owner_id` | `owner_id` |
| Exclude | `associatedcompanyid` | — |
| Exclude | `hs_analytics_source` | — |

### `hubspot.deals`
| Rule | Column | Target |
|------|--------|--------|
| Rename | `dealname` | `deal_name` |
| Rename | `dealstage` | `stage` |
| Rename | `hubspot_owner_id` | `owner_id` |
| Cast | `amount` | Numeric |
| Exclude | `associations` | — |

### `hubspot.companies`
| Rule | Column | Target |
|------|--------|--------|
| Rename | `name` | `company_name` |
| Rename | `domain` | `website` |
| Rename | `industry` | `sector` |
| Cast | `annualrevenue` | Numeric |
| Cast | `numberofemployees` | Integer |

### `hubspot.feedback_submissions`
| Rule | Column | Target |
|------|--------|--------|
| Rename | `hs_survey_type` | `survey_type` |
| Exclude | `hs_response_group` | — |

---

## Phase 4 — dbt SQL (stream-by-stream)

### Stream 1 — `hubspot.contacts`
```sql
SELECT
    id                                                              AS contact_id,
    email,
    TRIM(CONCAT(COALESCE(firstname,''),' ',COALESCE(lastname,''))) AS full_name,
    phone,
    lead_status, lifecycle, owner_id,
    CAST(createdate AS DATE)        AS created_on,
    CAST(lastmodifieddate AS DATE)  AS updated_on
FROM {{ source('raw', 'hubspot__contacts') }}
WHERE email IS NOT NULL
```

---

### Stream 2 — `hubspot.companies`
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
**Preview check**: `annual_rev` decimal or NULL; `headcount` integer or NULL

---

### Stream 3 — `hubspot.deals`
```sql
SELECT
    id AS deal_id, dealname AS deal_name,
    CAST(NULLIF(amount,'') AS NUMERIC)  AS deal_value,
    dealstage AS stage, pipeline,
    closedate::DATE                     AS close_date,
    hubspot_owner_id                    AS owner_id,
    CASE WHEN closedate IS NOT NULL AND createdate IS NOT NULL
         THEN DATEDIFF('day', CAST(createdate AS DATE), CAST(closedate AS DATE))
         ELSE NULL END                  AS days_to_close
FROM {{ source('raw', 'hubspot__deals') }}
WHERE amount IS NOT NULL AND amount != ''
```

---

### Stream 4 — `hubspot.owners`
```sql
SELECT
    id AS owner_id, firstName AS first_name, lastName AS last_name,
    email, userId AS user_id,
    CAST(createdAt AS DATE) AS created_on
FROM {{ source('raw', 'hubspot__owners') }}
```

---

### Stream 5 — `hubspot.tickets`
```sql
SELECT
    id AS ticket_id, subject AS ticket_name,
    hs_ticket_category  AS category,
    hs_ticket_priority  AS priority,
    hs_pipeline_stage   AS ticket_stage,
    hubspot_owner_id    AS owner_id,
    CAST(createdate AS DATE)   AS created_on,
    CAST(closed_date AS DATE)  AS closed_on
FROM {{ source('raw', 'hubspot__tickets') }}
```

---

### Stream 6 — `hubspot.calls`
```sql
SELECT
    id AS call_id, hs_call_title AS title,
    hs_call_status AS call_status,
    hs_call_direction AS direction,
    CAST(NULLIF(hs_call_duration,'') AS INTEGER) AS duration_ms,
    CAST(hs_timestamp AS TIMESTAMPTZ)            AS called_at
FROM {{ source('raw', 'hubspot__calls') }}
```

---

### Stream 7 — `hubspot.emails`
```sql
SELECT
    id AS email_id, hs_email_subject AS subject,
    hs_email_direction AS direction,
    hs_email_status    AS send_status,
    CAST(hs_timestamp AS TIMESTAMPTZ) AS sent_at
FROM {{ source('raw', 'hubspot__emails') }}
```

---

### Stream 8 — `hubspot.meetings`
```sql
SELECT
    id AS meeting_id, hs_meeting_title AS title,
    hs_meeting_outcome AS outcome,
    CAST(hs_meeting_start_time AS TIMESTAMPTZ) AS started_at,
    CAST(hs_meeting_end_time   AS TIMESTAMPTZ) AS ended_at,
    CASE WHEN hs_meeting_end_time IS NOT NULL AND hs_meeting_start_time IS NOT NULL
         THEN DATEDIFF('minute',
              CAST(hs_meeting_start_time AS TIMESTAMPTZ),
              CAST(hs_meeting_end_time   AS TIMESTAMPTZ))
         ELSE NULL END AS duration_mins
FROM {{ source('raw', 'hubspot__meetings') }}
```

---

### Stream 9 — `hubspot.notes`
```sql
SELECT
    id AS note_id, hs_note_body AS body,
    hubspot_owner_id AS owner_id,
    CAST(hs_timestamp AS DATE) AS created_on
FROM {{ source('raw', 'hubspot__notes') }}
```

---

### Stream 10 — `hubspot.tasks`
```sql
SELECT
    id AS task_id, hs_task_subject AS subject,
    hs_task_type     AS task_type,
    hs_task_status   AS status,
    hs_task_priority AS priority,
    CAST(hs_task_due_date AS DATE) AS due_on,
    CAST(createdate AS DATE)       AS created_on
FROM {{ source('raw', 'hubspot__tasks') }}
```

---

### Stream 11 — `hubspot.feedback_submissions`
```sql
SELECT
    id AS submission_id, hs_survey_type AS survey_type,
    hs_response->>'rating'                                AS rating,
    CAST(NULLIF(hs_response->>'nps_score','') AS INTEGER) AS nps_score,
    NULLIF(TRIM(hs_response->>'comment'),'')              AS comment,
    CAST(hs_submission_timestamp AS TIMESTAMPTZ)          AS submitted_at
FROM {{ source('raw', 'hubspot__feedback_submissions') }}
WHERE hs_response IS NOT NULL
```
**Preview check**: `rating` plain string; `nps_score` integer or NULL; `comment` plain text or NULL

---

### Stream 12 — `hubspot.pipelines`
```sql
SELECT
    id AS pipeline_id, label AS pipeline_label,
    objectType AS object_type,
    default AS is_default,
    CAST(createdAt AS DATE) AS created_on
FROM {{ source('raw', 'hubspot__pipelines') }}
```

---

### Stream 13 — `hubspot.pipeline_stages`
```sql
SELECT
    id AS stage_id, pipelineId AS pipeline_ref,
    label AS stage_label, displayOrder AS display_order,
    closed AS is_closed,
    CAST(NULLIF(probability,'') AS NUMERIC) AS probability
FROM {{ source('raw', 'hubspot__pipeline_stages') }}
```

---

### Stream 14 — `hubspot.associations`
```sql
SELECT
    id AS assoc_id, fromObjectId AS from_id,
    toObjectId AS to_id, type AS assoc_type,
    CAST(createdAt AS DATE) AS created_on
FROM {{ source('raw', 'hubspot__associations') }}
```

---

## Phase 5 — Preview Checks

| Stream | Column | Expected |
|--------|--------|---------|
| contacts | `full_name` | Combined string `"Jane Smith"` |
| contacts | `hs_lead_status` | Must NOT appear (renamed to `lead_status`) |
| deals | `deal_value` | Decimal or NULL (if empty string) |
| feedback | `nps_score` | Integer 0–10 or NULL |
| feedback | `hs_response` | Must NOT appear (extracted into columns) |
| companies | `annual_rev` | Decimal or NULL |

---

## Phase 7 — Failure Scenarios

| Scenario | Expected |
|---------|---------|
| `amount = ''` in deals | `NULLIF` returns NULL — Phase 2 ✅ |
| `hs_response` NULL | WHERE filters row — ✅ |
| Refresh token expired | Phase 1 ❌ `401` — re-auth required |
| Missing `crm.objects.deals.read` scope | Phase 1 ❌ `403` — update app permissions |

---

## Phase 8 — Verify

```sql
SELECT full_name FROM analytics.hs_contacts LIMIT 5;            -- combined string
SELECT lead_status FROM analytics.hs_contacts LIMIT 5;          -- NOT hs_lead_status
SELECT deal_value, days_to_close FROM analytics.hs_deals LIMIT 5;
SELECT nps_score, rating FROM analytics.hs_nps LIMIT 5;
```
