# Pipeline-by-Pipeline Step-by-Step Test Cases

Every pipeline, every stream. Each test case covers **all four** column-mapping patterns:

- **JSON → flat TEXT/VARCHAR** — extract specific keys from JSON/JSONB into flat columns
- **Column name mapping** — source column names differ from destination column names
- **Type coercion** — source data type differs from destination data type
- **Different column sets** — source has X columns, destination has Y different columns

---

## How to read each test case

```
Step 1  Pre-create the destination table (DDL shows intentionally different column names/types)
Step 2  Create source + destination connections in MantrixFlow UI
Step 3  Build the pipeline — select stream in Source Panel, set sync mode
Step 4  Set normalisation rules (if any)
Step 5  Write dbt SQL in Destination Panel (applies JSON extraction, aliases, casts)
Step 6  Click Validate SQL → confirm green
Step 7  Click Run
Step 8  Verify destination table in the DB
```

---
---

# PART A — SaaS Sources

---

## Pipeline 1 — Stripe → PostgreSQL

---

### Stream: `stripe.customers` → `analytics.customer_profiles`

> Source columns (15+): `id`, `email`, `name`, `description`, `phone`, `currency`, `balance`, `created`, `delinquent`, `discount`, `invoice_prefix`, `livemode`, `metadata`, `shipping`, `sources`, `subscriptions`  
> Destination needs: 5 columns, renamed, `created` converted to TIMESTAMPTZ

**Step 1 — Pre-create destination table**
```sql
CREATE TABLE analytics.customer_profiles (
    customer_id       TEXT        PRIMARY KEY,  -- source: id
    email_address     TEXT,                      -- source: email (renamed)
    display_name      TEXT,                      -- source: name (renamed)
    billing_currency  TEXT,                      -- source: currency (renamed)
    first_seen_at     TIMESTAMPTZ                -- source: created BIGINT → TIMESTAMPTZ
);
```

**Step 2 — Connections**  
Source: Stripe API key. Destination: PostgreSQL analytics DB.

**Step 3 — Source Panel**  
Stream: `stripe.customers` | Sync mode: `FULL_TABLE`

**Step 4 — Normalisation rules**  
None needed (aliases handled in dbt SQL).

**Step 5 — dbt SQL**
```sql
SELECT
    id                              AS customer_id,
    email                           AS email_address,
    name                            AS display_name,
    currency                        AS billing_currency,
    TO_TIMESTAMP(created)::TIMESTAMPTZ AS first_seen_at
FROM {{ source('raw', 'stripe__customers') }}
WHERE email IS NOT NULL
  AND livemode = true
```
> Source `created` is Unix epoch integer → destination `TIMESTAMPTZ`. 10 source columns excluded (description, phone, balance, delinquent, discount, invoice_prefix, metadata, shipping, sources, subscriptions).

**Step 6 — Validate SQL** → ✅ green (5 output cols match destination 5 cols)

**Step 7 — Run**

**Step 8 — Verify**
```sql
SELECT customer_id, email_address, billing_currency, first_seen_at
FROM analytics.customer_profiles LIMIT 5;
-- first_seen_at must be TIMESTAMPTZ (e.g. 2024-01-15 10:30:00+00), not an integer
-- display_name column exists, not 'name'
SELECT COUNT(*) FROM analytics.customer_profiles;
```

---

### Stream: `stripe.charges` → `analytics.payment_ledger`

> Source `amount` is integer (cents). Source `created` is Unix epoch. Source has `payment_intent`, `metadata` (JSON), `payment_method_details` (JSON) — not needed in destination.

**Step 1 — Pre-create destination table**
```sql
CREATE TABLE analytics.payment_ledger (
    payment_ref       TEXT           PRIMARY KEY,  -- source: id
    amount_dollars    NUMERIC(10,2),               -- source: amount INT (cents) / 100
    payment_currency  TEXT,                         -- source: currency
    payment_status    TEXT,                         -- source: status
    customer_ref      TEXT,                         -- source: customer
    card_brand        TEXT,                         -- from JSON: payment_method_details->'card'->>'brand'
    card_last4        TEXT,                         -- from JSON: payment_method_details->'card'->>'last4'
    charged_at        TIMESTAMPTZ                   -- source: created BIGINT → TIMESTAMPTZ
);
```

**Step 3 — Source Panel** Stream: `stripe.charges` | Sync: `INCREMENTAL` | Key: `created`

**Step 5 — dbt SQL**
```sql
SELECT
    id                                                     AS payment_ref,
    amount / 100.0                                         AS amount_dollars,
    currency                                               AS payment_currency,
    status                                                 AS payment_status,
    customer                                               AS customer_ref,
    payment_method_details->'card'->>'brand'               AS card_brand,
    payment_method_details->'card'->>'last4'               AS card_last4,
    TO_TIMESTAMP(created)::TIMESTAMPTZ                     AS charged_at
FROM {{ source('raw', 'stripe__charges') }}
WHERE status = 'succeeded'
```
> JSON: `payment_method_details` (nested JSON) → 2 flat TEXT columns. Integer cents → NUMERIC dollars. 8+ source columns excluded.

**Step 8 — Verify**
```sql
SELECT amount_dollars, card_brand, card_last4, charged_at
FROM analytics.payment_ledger LIMIT 5;
-- amount_dollars must be decimal (e.g. 29.99, not 2999)
-- card_brand must be plain text (e.g. 'visa'), not a JSON object
```

---

### Stream: `stripe.invoices` → `analytics.invoice_summary`

**Step 1 — Pre-create destination table**
```sql
CREATE TABLE analytics.invoice_summary (
    invoice_ref     TEXT        PRIMARY KEY,  -- source: id
    customer_ref    TEXT,                      -- source: customer
    due_amount      NUMERIC(10,2),             -- source: amount_due (cents) / 100
    paid_amount     NUMERIC(10,2),             -- source: amount_paid (cents) / 100
    invoice_status  TEXT,                      -- source: status (renamed)
    issued_on       DATE,                      -- source: created BIGINT → DATE
    period_start    TIMESTAMPTZ,               -- source: period_start BIGINT → TIMESTAMPTZ
    period_end      TIMESTAMPTZ                -- source: period_end BIGINT → TIMESTAMPTZ
);
```

**Step 5 — dbt SQL**
```sql
SELECT
    id                                        AS invoice_ref,
    customer                                  AS customer_ref,
    amount_due  / 100.0                       AS due_amount,
    amount_paid / 100.0                       AS paid_amount,
    status                                    AS invoice_status,
    TO_TIMESTAMP(created)::DATE               AS issued_on,
    TO_TIMESTAMP(period_start)::TIMESTAMPTZ   AS period_start,
    TO_TIMESTAMP(period_end)::TIMESTAMPTZ     AS period_end
FROM {{ source('raw', 'stripe__invoices') }}
WHERE status IN ('paid', 'open', 'void')
```
> Unix epoch × 3 columns converted. Cents → dollars × 2. Column renamed (`status` → `invoice_status`).

**Step 8 — Verify**
```sql
SELECT due_amount, paid_amount, issued_on, period_start
FROM analytics.invoice_summary LIMIT 5;
-- issued_on must be DATE only (no time), period_start must be TIMESTAMPTZ
```

---

### Stream: `stripe.events` → `analytics.stripe_event_log`

> `data` column is a large JSONB blob. Destination needs only 3 scalar keys from it.

**Step 1 — Pre-create destination table**
```sql
CREATE TABLE analytics.stripe_event_log (
    event_ref     TEXT        PRIMARY KEY,  -- source: id
    event_type    TEXT,                      -- source: type (renamed)
    object_type   TEXT,                      -- from JSON: data->>'object'
    object_id     TEXT,                      -- from JSON: data->>'id'
    is_live       BOOLEAN,                   -- source: livemode (same type, renamed)
    occurred_at   TIMESTAMPTZ                -- source: created BIGINT → TIMESTAMPTZ
    -- data blob NOT stored: gateway, api_version, pending_webhooks, request excluded
);
```

**Step 5 — dbt SQL**
```sql
SELECT
    id                                 AS event_ref,
    type                               AS event_type,
    data->>'object'                    AS object_type,
    data->>'id'                        AS object_id,
    livemode                           AS is_live,
    TO_TIMESTAMP(created)::TIMESTAMPTZ AS occurred_at
FROM {{ source('raw', 'stripe__events') }}
WHERE livemode = true
```
> JSON key extraction: 2 flat TEXT columns from `data` JSONB. 4 source columns excluded.

**Step 8 — Verify**
```sql
SELECT event_type, object_type, object_id FROM analytics.stripe_event_log LIMIT 5;
-- object_type = 'charge', 'customer', etc. (plain string, not JSON)
-- 'data' column must NOT exist in destination
```

---

### Stream: `stripe.subscriptions` → `analytics.subscription_state`

**Step 1 — Pre-create destination table**
```sql
CREATE TABLE analytics.subscription_state (
    sub_id           TEXT        PRIMARY KEY,  -- source: id
    customer_ref     TEXT,                      -- source: customer
    plan_id          TEXT,                      -- from JSON: plan->>'id'
    plan_interval    TEXT,                      -- from JSON: plan->>'interval'
    amount_monthly   NUMERIC(10,2),             -- from JSON: plan->>'amount' (cents) / 100
    sub_status       TEXT,                      -- source: status (renamed)
    trial_ends_on    DATE,                      -- source: trial_end BIGINT → DATE
    renews_on        DATE                       -- source: current_period_end BIGINT → DATE
);
```

**Step 5 — dbt SQL**
```sql
SELECT
    id                                                       AS sub_id,
    customer                                                 AS customer_ref,
    plan->>'id'                                              AS plan_id,
    plan->>'interval'                                        AS plan_interval,
    CAST(plan->>'amount' AS NUMERIC) / 100.0                 AS amount_monthly,
    status                                                   AS sub_status,
    CASE WHEN trial_end IS NOT NULL
         THEN TO_TIMESTAMP(trial_end)::DATE ELSE NULL END    AS trial_ends_on,
    TO_TIMESTAMP(current_period_end)::DATE                   AS renews_on
FROM {{ source('raw', 'stripe__subscriptions') }}
```
> JSON extraction from `plan` nested object → 3 flat columns. Cents conversion. 2 epoch → DATE conversions.

---

### Streams: `stripe.payment_methods`, `stripe.refunds`, `stripe.disputes` — Quick-map steps

**`stripe.payment_methods` → `analytics.card_vault`**
```sql
-- Step 1: Destination
CREATE TABLE analytics.card_vault (
    pm_id       TEXT PRIMARY KEY,   -- source: id
    customer_ref TEXT,              -- source: customer
    card_brand  TEXT,               -- from JSON: card->>'brand'
    card_last4  TEXT,               -- from JSON: card->>'last4'
    card_expiry TEXT,               -- derived: card->>'exp_month' || '/' || card->>'exp_year'
    added_at    TIMESTAMPTZ         -- source: created → TIMESTAMPTZ
);
-- Step 5: dbt SQL
SELECT
    id                                                            AS pm_id,
    customer                                                      AS customer_ref,
    card->>'brand'                                                AS card_brand,
    card->>'last4'                                                AS card_last4,
    (card->>'exp_month') || '/' || (card->>'exp_year')           AS card_expiry,
    TO_TIMESTAMP(created)::TIMESTAMPTZ                            AS added_at
FROM {{ source('raw', 'stripe__payment_methods') }}
WHERE type = 'card'
```

**`stripe.refunds` → `analytics.refund_log`**
```sql
CREATE TABLE analytics.refund_log (
    refund_ref     TEXT PRIMARY KEY,  -- source: id
    charge_ref     TEXT,              -- source: charge
    refund_amount  NUMERIC(10,2),     -- source: amount (cents)/100
    refund_reason  TEXT,              -- source: reason
    refund_status  TEXT,              -- source: status
    refunded_at    TIMESTAMPTZ        -- source: created → TIMESTAMPTZ
);
SELECT
    id                                 AS refund_ref,
    charge                             AS charge_ref,
    amount / 100.0                     AS refund_amount,
    reason                             AS refund_reason,
    status                             AS refund_status,
    TO_TIMESTAMP(created)::TIMESTAMPTZ AS refunded_at
FROM {{ source('raw', 'stripe__refunds') }}
```

**`stripe.tax_rates` → `analytics.tax_rate_master`**
```sql
CREATE TABLE analytics.tax_rate_master (
    rate_id       TEXT    PRIMARY KEY,
    display_name  TEXT,
    rate_pct      NUMERIC(6,4),   -- source: percentage (float string) → NUMERIC
    is_inclusive  BOOLEAN,
    is_active     BOOLEAN
);
SELECT
    id                              AS rate_id,
    display_name,
    CAST(percentage AS NUMERIC)     AS rate_pct,
    inclusive                       AS is_inclusive,
    active                          AS is_active
FROM {{ source('raw', 'stripe__tax_rates') }}
```

---

## Pipeline 2 — Shopify → PostgreSQL

---

### Stream: `shopify.products` → `analytics.product_catalog`

**Step 1 — Pre-create destination table**
```sql
CREATE TABLE analytics.product_catalog (
    product_id    BIGINT      PRIMARY KEY,  -- source: id
    product_name  TEXT,                      -- source: title (renamed)
    brand         TEXT,                      -- source: vendor (renamed)
    category      TEXT,                      -- source: product_type (renamed)
    is_published  BOOLEAN,                   -- derived: status = 'active'
    listed_on     DATE,                      -- source: published_at → DATE
    last_modified DATE                       -- source: updated_at → DATE
);
```

**Step 5 — dbt SQL**
```sql
SELECT
    id                               AS product_id,
    title                            AS product_name,
    vendor                           AS brand,
    product_type                     AS category,
    status = 'active'                AS is_published,
    published_at::DATE               AS listed_on,
    updated_at::DATE                 AS last_modified
FROM {{ source('raw', 'shopify__products') }}
WHERE status != 'archived'
```
> 3 columns renamed. `status` TEXT → `is_published` BOOLEAN (derived). 2 TIMESTAMPTZ → DATE.

**Step 8 — Verify**
```sql
SELECT product_name, brand, category, is_published, listed_on
FROM analytics.product_catalog LIMIT 5;
-- is_published must be true/false, not 'active'/'draft'
-- listed_on must be DATE only
```

---

### Stream: `shopify.orders` → `analytics.order_register`

> `line_items` is a JSON array. `billing_address`, `shipping_address` are JSON objects.

**Step 1 — Pre-create destination table**
```sql
CREATE TABLE analytics.order_register (
    order_id         BIGINT      PRIMARY KEY,  -- source: id
    customer_email   TEXT,                      -- source: email
    order_total      NUMERIC(10,2),             -- source: total_price TEXT → NUMERIC
    currency         TEXT,
    order_status     TEXT,                      -- source: financial_status (renamed)
    fulfil_status    TEXT,                      -- source: fulfillment_status (renamed)
    ship_city        TEXT,                      -- from JSON: shipping_address->>'city'
    ship_country     TEXT,                      -- from JSON: shipping_address->>'country_code'
    item_count       INTEGER,                   -- from JSON: line_items array length
    ordered_on       DATE                       -- source: created_at → DATE
);
```

**Step 5 — dbt SQL**
```sql
SELECT
    id                                            AS order_id,
    email                                         AS customer_email,
    CAST(total_price AS NUMERIC)                  AS order_total,
    currency,
    financial_status                              AS order_status,
    fulfillment_status                            AS fulfil_status,
    shipping_address->>'city'                     AS ship_city,
    shipping_address->>'country_code'             AS ship_country,
    JSON_ARRAY_LENGTH(line_items)                 AS item_count,
    created_at::DATE                              AS ordered_on
FROM {{ source('raw', 'shopify__orders') }}
WHERE financial_status = 'paid'
```
> JSON extraction: 2 scalar keys from `shipping_address`. JSON array length from `line_items`. TEXT → NUMERIC cast. TIMESTAMPTZ → DATE.

---

### Stream: `shopify.customers` → `analytics.shopper_profiles`

**Step 1 — Pre-create destination table**
```sql
CREATE TABLE analytics.shopper_profiles (
    shopper_id      BIGINT      PRIMARY KEY,  -- source: id
    email           TEXT,
    full_name       TEXT,                      -- derived: CONCAT(first_name, ' ', last_name)
    order_count     INTEGER,                   -- source: orders_count (renamed)
    lifetime_spend  NUMERIC(10,2),             -- source: total_spent TEXT → NUMERIC
    is_verified     BOOLEAN,                   -- source: verified_email
    registered_on   DATE                       -- source: created_at → DATE
);
```

**Step 5 — dbt SQL**
```sql
SELECT
    id                                              AS shopper_id,
    email,
    TRIM(CONCAT(
        COALESCE(first_name, ''),
        ' ',
        COALESCE(last_name, '')
    ))                                              AS full_name,
    orders_count                                    AS order_count,
    CAST(total_spent AS NUMERIC)                    AS lifetime_spend,
    verified_email                                  AS is_verified,
    created_at::DATE                                AS registered_on
FROM {{ source('raw', 'shopify__customers') }}
WHERE email IS NOT NULL
```
> Two source columns merged into one (`full_name`). `verified_email` BOOL → `is_verified` BOOL renamed. TEXT → NUMERIC.

---

### Remaining Shopify streams — Quick-map

| Stream | Source JSON column | Extracted keys | Key renames | Type changes |
|--------|-------------------|----------------|-------------|--------------|
| `draft_orders` | — | — | `total_price`→`draft_total` (NUMERIC), `status`→`draft_status` | TEXT→NUMERIC |
| `pages` | — | — | `title`→`page_title`, `handle`→`url_slug`, `published_at`→`published_on` (DATE) | TIMESTAMPTZ→DATE |
| `articles` | — | — | `title`→`article_title`, `author`→`author_name`, `published_at`→`published_on` (DATE) | TIMESTAMPTZ→DATE |
| `locations` | — | — | `name`→`location_name`, `address1`→`street`, `country`→`country_name` | — |
| `price_rules` | — | — | `value`→`discount_value` (NUMERIC), `starts_at`→`valid_from`, `ends_at`→`valid_to` | TEXT→NUMERIC, TIMESTAMPTZ→DATE |
| `themes` | — | — | `name`→`theme_name`, `role`→`theme_role` | — |
| `collects` | — | — | `collection_id`→`collection_ref`, `product_id`→`product_ref` | — |

---

## Pipeline 3 — HubSpot → PostgreSQL

> The authoritative beta guide is
> [`pipelines/saas/11-hubspot-to-postgres.md`](pipelines/saas/11-hubspot-to-postgres.md).
> Only its ten-stream dlt catalog is supported. Any calls, emails, meetings, or
> feedback-submission examples remaining in this legacy walkthrough are
> historical and must not be used for acceptance.

---

### Stream: `hubspot.contacts` → `analytics.crm_contacts`

**Step 1 — Pre-create destination table**
```sql
CREATE TABLE analytics.crm_contacts (
    contact_id    TEXT        PRIMARY KEY,  -- source: id
    email         TEXT,
    full_name     TEXT,                      -- derived: CONCAT(firstname, ' ', lastname)
    phone         TEXT,                      -- source: phone
    lead_status   TEXT,                      -- source: hs_lead_status (hs_ stripped)
    lifecycle     TEXT,                      -- source: lifecyclestage (renamed)
    owner_id      TEXT,                      -- source: hubspot_owner_id (renamed)
    created_date  DATE,                      -- source: createdate → DATE
    updated_date  DATE                       -- source: lastmodifieddate → DATE
);
```

**Step 4 — Normalisation rules**
```json
[
  { "rule_type": "rename", "table": "hubspot.contacts", "column": "hs_lead_status",     "destination_name": "lead_status" },
  { "rule_type": "rename", "table": "hubspot.contacts", "column": "lifecyclestage",      "destination_name": "lifecycle" },
  { "rule_type": "rename", "table": "hubspot.contacts", "column": "hubspot_owner_id",    "destination_name": "owner_id" }
]
```

**Step 5 — dbt SQL**
```sql
SELECT
    id                                             AS contact_id,
    email,
    TRIM(CONCAT(
        COALESCE(firstname, ''),
        ' ',
        COALESCE(lastname, '')
    ))                                             AS full_name,
    phone,
    lead_status,
    lifecycle,
    owner_id,
    CAST(createdate AS DATE)                       AS created_date,
    CAST(lastmodifieddate AS DATE)                 AS updated_date
FROM {{ source('raw', 'hubspot__contacts') }}
WHERE email IS NOT NULL
```

---

### Stream: `hubspot.deals` → `analytics.crm_deals`

**Step 1 — Pre-create destination table**
```sql
CREATE TABLE analytics.crm_deals (
    deal_id       TEXT        PRIMARY KEY,  -- source: id
    deal_name     TEXT,                      -- source: dealname (renamed)
    deal_value    NUMERIC(12,2),             -- source: amount TEXT → NUMERIC (renamed)
    stage         TEXT,                      -- source: dealstage (renamed)
    pipeline      TEXT,
    close_date    DATE,                      -- source: closedate → DATE
    owner_id      TEXT,                      -- source: hubspot_owner_id (renamed)
    days_to_close INTEGER                    -- derived: DATEDIFF days
);
```

**Step 5 — dbt SQL**
```sql
SELECT
    id                                                    AS deal_id,
    dealname                                              AS deal_name,
    CAST(NULLIF(amount, '') AS NUMERIC)                   AS deal_value,
    dealstage                                             AS stage,
    pipeline,
    closedate::DATE                                       AS close_date,
    hubspot_owner_id                                      AS owner_id,
    CASE
        WHEN closedate IS NOT NULL AND createdate IS NOT NULL
        THEN DATEDIFF('day', CAST(createdate AS DATE), CAST(closedate AS DATE))
        ELSE NULL
    END                                                   AS days_to_close
FROM {{ source('raw', 'hubspot__deals') }}
WHERE amount IS NOT NULL AND amount != ''
```
> TEXT → NUMERIC with `NULLIF` guard for empty strings. Derived `days_to_close` INTEGER. Multiple renames.

---

### Stream: `hubspot.companies` → `analytics.crm_companies`

**Step 1 — Pre-create destination table**
```sql
CREATE TABLE analytics.crm_companies (
    company_id    TEXT        PRIMARY KEY,
    company_name  TEXT,                -- source: name
    website       TEXT,                -- source: domain (renamed)
    sector        TEXT,                -- source: industry (renamed)
    city          TEXT,
    country       TEXT,
    annual_rev    NUMERIC(15,2),       -- source: annualrevenue TEXT → NUMERIC
    headcount     INTEGER,             -- source: numberofemployees TEXT → INTEGER
    created_date  DATE
);
```

**Step 5 — dbt SQL**
```sql
SELECT
    id                                                  AS company_id,
    name                                                AS company_name,
    domain                                              AS website,
    industry                                            AS sector,
    city,
    country,
    CAST(NULLIF(annualrevenue, '') AS NUMERIC)          AS annual_rev,
    CAST(NULLIF(numberofemployees, '') AS INTEGER)      AS headcount,
    CAST(createdate AS DATE)                            AS created_date
FROM {{ source('raw', 'hubspot__companies') }}
WHERE name IS NOT NULL
```

---

### Activity streams: `hubspot.calls`, `hubspot.emails`, `hubspot.meetings` — Quick-map

**Calls → `analytics.call_log`**
```sql
SELECT
    id                                   AS call_id,
    hs_call_title                        AS title,
    hs_call_status                       AS call_status,
    CAST(hs_call_duration AS INTEGER)    AS duration_ms,
    CAST(hs_timestamp AS TIMESTAMPTZ)    AS called_at
FROM {{ source('raw', 'hubspot__calls') }}
```
Destination: `call_id TEXT PK, title TEXT, call_status TEXT, duration_ms INTEGER, called_at TIMESTAMPTZ`

**Emails → `analytics.email_log`**
```sql
SELECT
    id                                   AS email_id,
    hs_email_subject                     AS subject,
    hs_email_status                      AS send_status,
    hs_email_direction                   AS direction,
    CAST(hs_timestamp AS TIMESTAMPTZ)    AS sent_at
FROM {{ source('raw', 'hubspot__emails') }}
```

**Meetings → `analytics.meeting_log`**
```sql
SELECT
    id                                             AS meeting_id,
    hs_meeting_title                               AS title,
    hs_meeting_outcome                             AS outcome,
    CAST(hs_meeting_start_time AS TIMESTAMPTZ)     AS started_at,
    CAST(hs_meeting_end_time AS TIMESTAMPTZ)       AS ended_at
FROM {{ source('raw', 'hubspot__meetings') }}
```

---

### Stream: `hubspot.feedback_submissions` → `analytics.nps_responses`

> `hs_response` is a JSON string with keys: `rating`, `comment`, `nps_score`.

**Step 1 — Pre-create destination table**
```sql
CREATE TABLE analytics.nps_responses (
    submission_id  TEXT        PRIMARY KEY,
    survey_type    TEXT,                    -- source: hs_survey_type (renamed + hs_ stripped)
    rating         TEXT,                    -- from JSON: hs_response->>'rating'
    nps_score      INTEGER,                 -- from JSON: hs_response->>'nps_score' → INTEGER
    comment        TEXT,                    -- from JSON: hs_response->>'comment'
    submitted_at   TIMESTAMPTZ             -- source: hs_submission_timestamp (renamed)
);
```

**Step 5 — dbt SQL**
```sql
SELECT
    id                                                        AS submission_id,
    hs_survey_type                                            AS survey_type,
    hs_response->>'rating'                                    AS rating,
    CAST(NULLIF(hs_response->>'nps_score', '') AS INTEGER)    AS nps_score,
    NULLIF(TRIM(hs_response->>'comment'), '')                 AS comment,
    CAST(hs_submission_timestamp AS TIMESTAMPTZ)              AS submitted_at
FROM {{ source('raw', 'hubspot__feedback_submissions') }}
WHERE hs_response IS NOT NULL
```
> JSON extraction: 3 keys from `hs_response`. `nps_score` TEXT → INTEGER. Empty comment → NULL.

---

## Pipeline 4 — GitHub → PostgreSQL

---

### Stream: `github.issues` → `analytics.issue_tracker`

**Step 1 — Pre-create destination table**
```sql
CREATE TABLE analytics.issue_tracker (
    issue_id      BIGINT      PRIMARY KEY,  -- source: id
    issue_number  INTEGER,                   -- source: number (renamed)
    title         TEXT,
    state         TEXT,
    author_login  TEXT,                      -- from JSON: user->>'login'
    label_names   TEXT,                      -- derived: first label name
    is_closed     BOOLEAN,                   -- derived: state = 'closed'
    opened_on     DATE,                      -- source: created_at → DATE
    closed_on     DATE                       -- source: closed_at → DATE (nullable)
);
```

**Step 5 — dbt SQL**
```sql
SELECT
    id                                      AS issue_id,
    number                                  AS issue_number,
    title,
    state,
    user->>'login'                          AS author_login,
    labels->0->>'name'                      AS label_names,
    state = 'closed'                        AS is_closed,
    created_at::DATE                        AS opened_on,
    closed_at::DATE                         AS closed_on
FROM {{ source('raw', 'github__issues') }}
```
> JSON: `user` nested object → `author_login` TEXT. JSON array first element: `labels->0->>'name'`. Derived BOOLEAN. Two TIMESTAMPTZ → DATE.

---

### Stream: `github.pull_requests` → `analytics.pr_tracker`

**Step 1 — Pre-create destination table**
```sql
CREATE TABLE analytics.pr_tracker (
    pr_id         BIGINT      PRIMARY KEY,
    pr_number     INTEGER,                   -- source: number
    title         TEXT,
    state         TEXT,
    author_login  TEXT,                      -- from JSON: user->>'login'
    head_branch   TEXT,                      -- from JSON: head->>'ref'
    base_branch   TEXT,                      -- from JSON: base->>'ref'
    is_merged     BOOLEAN,                   -- derived: merged_at IS NOT NULL
    opened_on     DATE,
    merged_on     DATE
);
```

**Step 5 — dbt SQL**
```sql
SELECT
    id                                  AS pr_id,
    number                              AS pr_number,
    title,
    state,
    user->>'login'                      AS author_login,
    head->>'ref'                        AS head_branch,
    base->>'ref'                        AS base_branch,
    merged_at IS NOT NULL               AS is_merged,
    created_at::DATE                    AS opened_on,
    merged_at::DATE                     AS merged_on
FROM {{ source('raw', 'github__pull_requests') }}
```
> JSON nested objects: `user`, `head`, `base` → 3 flat TEXT columns. Derived BOOLEAN. Nullable DATE.

---

### Stream: `github.events` → `analytics.repo_event_log`

> `payload` JSON varies by event type (huge object). `actor` and `repo` are JSON objects.

**Step 1 — Pre-create destination table**
```sql
CREATE TABLE analytics.repo_event_log (
    event_id      TEXT        PRIMARY KEY,
    event_type    TEXT,                      -- source: type (renamed)
    actor_login   TEXT,                      -- from JSON: actor->>'login'
    actor_id      INTEGER,                   -- from JSON: actor->>'id' → INTEGER
    repo_name     TEXT,                      -- from JSON: repo->>'name'
    action        TEXT,                      -- from JSON: payload->>'action'
    occurred_at   TIMESTAMPTZ                -- source: created_at
    -- payload blob excluded; actor/repo blobs excluded
);
```

**Step 5 — dbt SQL**
```sql
SELECT
    id                                       AS event_id,
    type                                     AS event_type,
    actor->>'login'                          AS actor_login,
    CAST(actor->>'id' AS INTEGER)            AS actor_id,
    repo->>'name'                            AS repo_name,
    payload->>'action'                       AS action,
    created_at                               AS occurred_at
FROM {{ source('raw', 'github__events') }}
WHERE type IN (
    'PushEvent','PullRequestEvent','IssuesEvent',
    'WatchEvent','ForkEvent','ReleaseEvent'
)
```
> JSON extraction from 3 separate JSON objects (`actor`, `repo`, `payload`). Integer cast from JSON string.

---

### Stream: `github.commits` → `analytics.commit_log`

> Staging: dlt flattens `commit` into `commit__message`, `commit__author__name`, `commit__author__date`, etc.

**Step 1 — Pre-create destination table**
```sql
CREATE TABLE analytics.commit_log (
    sha           TEXT        PRIMARY KEY,
    message       TEXT,                      -- from: commit__message
    author_name   TEXT,                      -- from: commit__author__name
    author_email  TEXT,                      -- from: commit__author__email
    committed_at  TIMESTAMPTZ                -- from: commit__author__date
);
```

**Step 5 — dbt SQL**
```sql
SELECT
    sha,
    commit__message                        AS message,
    commit__author__name                   AS author_name,
    commit__author__email                  AS author_email,
    CAST(commit__author__date AS TIMESTAMPTZ) AS committed_at
FROM {{ source('raw', 'github__commits') }}
WHERE commit__author__name IS NOT NULL
```
> dlt flattened columns. CAST if needed for TIMESTAMPTZ.

---

### Remaining GitHub streams — Quick-map

| Stream | Destination table | Key mapping |
|--------|------------------|-------------|
| `github.stargazers` | `analytics.star_log` | `user->>'login'` → `user_login TEXT`, `starred_at` → TIMESTAMPTZ |
| `github.releases` | `analytics.release_log` | `tag_name`→`version TEXT`, `prerelease`→`is_prerelease BOOLEAN`, `draft`→`is_draft BOOLEAN`, `published_at`→DATE |
| `github.contributors` | `analytics.contributor_board` | `login`→`github_login`, `contributions`→`commit_count INTEGER`, `id`→`github_user_id INTEGER` |
| `github.milestones` | `analytics.milestone_tracker` | `number`→`milestone_number`, `open_issues`→`open_count INTEGER`, `closed_issues`→`closed_count INTEGER`, `due_on`→DATE |
| `github.labels` | `analytics.label_master` | `name`→`label_name`, `color`→`hex_color`, `default`→`is_default BOOLEAN` |
| `github.forks` | `analytics.fork_registry` | `owner->>'login'`→`owner_login TEXT`, `id`→`fork_id BIGINT`, `created_at`→DATE |

---

## Pipeline 5 — Notion → PostgreSQL

---

### Stream: `notion.databases` → `analytics.notion_db_registry`

**Step 1 — Pre-create destination table**
```sql
CREATE TABLE analytics.notion_db_registry (
    db_id        TEXT        PRIMARY KEY,  -- source: id
    db_title     TEXT,                      -- from JSON array: title->0->>'plain_text'
    db_url       TEXT,                      -- source: url (renamed)
    created_on   DATE,                      -- source: created_time → DATE
    modified_on  DATE                       -- source: last_edited_time → DATE
);
```

**Step 5 — dbt SQL**
```sql
SELECT
    id                                           AS db_id,
    title->0->>'plain_text'                      AS db_title,
    url                                          AS db_url,
    CAST(created_time AS DATE)                   AS created_on,
    CAST(last_edited_time AS DATE)               AS modified_on
FROM {{ source('raw', 'notion__databases') }}
WHERE archived = false
  AND title->0->>'plain_text' IS NOT NULL
```
> JSON rich-text array first element. TIMESTAMPTZ → DATE × 2.

---

### Stream: `notion.pages` → `analytics.notion_pages_flat`

> `properties` is a deeply nested JSON object (every Notion column is a key).

**Step 1 — Pre-create destination table**
```sql
CREATE TABLE analytics.notion_pages_flat (
    page_id       TEXT        PRIMARY KEY,
    page_url      TEXT,                     -- source: url
    status        TEXT,                     -- from JSON: properties->'Status'->'select'->>'name'
    priority      TEXT,                     -- from JSON: properties->'Priority'->'select'->>'name'
    assignee      TEXT,                     -- from JSON: properties->'Assignee'->'people'->0->>'name'
    due_date      DATE,                     -- from JSON: properties->'Due Date'->'date'->>'start' → DATE
    created_on    DATE,                     -- source: created_time → DATE
    modified_on   DATE                      -- source: last_edited_time → DATE
    -- properties blob: NOT stored
);
```

**Step 5 — dbt SQL**
```sql
SELECT
    id                                                          AS page_id,
    url                                                         AS page_url,
    properties->'Status'->'select'->>'name'                     AS status,
    properties->'Priority'->'select'->>'name'                   AS priority,
    properties->'Assignee'->'people'->0->>'name'                AS assignee,
    CAST(
        properties->'Due Date'->'date'->>'start'
        AS DATE
    )                                                           AS due_date,
    CAST(created_time AS DATE)                                  AS created_on,
    CAST(last_edited_time AS DATE)                              AS modified_on
FROM {{ source('raw', 'notion__pages') }}
WHERE archived = false
```
> 3-level JSON extraction. JSON string → DATE cast. `properties` blob excluded entirely.

**Step 8 — Verify**
```sql
SELECT page_id, status, priority, assignee, due_date FROM analytics.notion_pages_flat LIMIT 5;
-- status must be plain text ('In Progress'), not JSON
-- due_date must be DATE type
-- 'properties' column must NOT exist
```

---

### Stream: `notion.users` → `analytics.notion_user_directory`

**Step 1 — Pre-create destination table**
```sql
CREATE TABLE analytics.notion_user_directory (
    user_id      TEXT  PRIMARY KEY,
    display_name TEXT,                 -- source: name (renamed)
    user_type    TEXT,                 -- source: type
    email        TEXT,                 -- from JSON: person->>'email'
    avatar_url   TEXT
);
```

**Step 5 — dbt SQL**
```sql
SELECT
    id                     AS user_id,
    name                   AS display_name,
    type                   AS user_type,
    person->>'email'       AS email,
    avatar_url
FROM {{ source('raw', 'notion__users') }}
WHERE type = 'person'
```

---
---

# PART B — DB Sources

---

## Pipeline 6 — PostgreSQL Source → PostgreSQL Destination

---

### Stream: `public.users` → `analytics.user_profiles`

**Step 1 — Pre-create destination table**
```sql
CREATE TABLE analytics.user_profiles (
    user_uuid       UUID        PRIMARY KEY,  -- source: id (renamed)
    email_address   TEXT,                      -- source: email (renamed)
    fname           TEXT,                      -- source: first_name (renamed)
    lname           TEXT,                      -- source: last_name (renamed)
    full_name       TEXT,                      -- derived: CONCAT(first_name,' ',last_name)
    registered_on   DATE,                      -- source: created_at TIMESTAMPTZ → DATE
    last_active_on  DATE                       -- source: updated_at TIMESTAMPTZ → DATE
);
```

**Step 3 — Source Panel** Stream: `public.users` | Sync: `INCREMENTAL` | Key: `updated_at`

**Step 4 — Normalisation rules**
```json
[
  { "rule_type": "rename", "table": "public.users", "column": "first_name", "destination_name": "fname" },
  { "rule_type": "rename", "table": "public.users", "column": "last_name",  "destination_name": "lname" }
]
```

**Step 5 — dbt SQL**
```sql
SELECT
    id                                              AS user_uuid,
    email                                           AS email_address,
    fname,
    lname,
    TRIM(CONCAT(COALESCE(fname,''), ' ', COALESCE(lname,''))) AS full_name,
    created_at::DATE                                AS registered_on,
    updated_at::DATE                                AS last_active_on
FROM {{ source('raw', 'public__users') }}
```
> Normalisation renames applied first, then referenced in dbt SQL. Two source columns merged into one derived column.

**Step 8 — Verify**
```sql
SELECT user_uuid, email_address, full_name, registered_on FROM analytics.user_profiles LIMIT 5;
-- columns must NOT be: id, email, first_name, last_name, created_at, updated_at
-- full_name must be a single concatenated string
```

---

### Stream: `public.orders` → `analytics.order_facts`

> `metadata` JSONB column with 10 keys. Source has 6 columns. Destination has 9 (3 extracted from JSON + 2 type-converted).

**Step 1 — Pre-create destination table**
```sql
CREATE TABLE analytics.order_facts (
    order_id        UUID        PRIMARY KEY,  -- source: id
    customer_ref    UUID,                      -- source: user_id (renamed)
    order_amount    NUMERIC(10,2),             -- source: amount (same type, renamed)
    order_status    TEXT,                      -- source: status (renamed)
    payment_method  TEXT,                      -- from JSON: metadata->>'payment_method'
    currency        TEXT,                      -- from JSON: metadata->>'currency'
    sales_channel   TEXT,                      -- from JSON: metadata->>'channel'
    is_high_value   BOOLEAN,                   -- derived: amount > 100
    placed_on       DATE                       -- source: placed_at TIMESTAMPTZ → DATE
    -- user_id, metadata excluded from dest by name/exclusion
);
```

**Step 5 — dbt SQL**
```sql
SELECT
    id                                  AS order_id,
    user_id                             AS customer_ref,
    amount                              AS order_amount,
    status                              AS order_status,
    metadata->>'payment_method'         AS payment_method,
    metadata->>'currency'               AS currency,
    metadata->>'channel'                AS sales_channel,
    amount > 100                        AS is_high_value,
    placed_at::DATE                     AS placed_on
FROM {{ source('raw', 'public__orders') }}
WHERE metadata IS NOT NULL
```
> JSON: 3 keys from 10-key `metadata`. Derived BOOLEAN. TIMESTAMPTZ → DATE. `metadata` blob excluded.

**Step 8 — Verify**
```sql
SELECT order_id, payment_method, currency, sales_channel, is_high_value, placed_on
FROM analytics.order_facts LIMIT 5;
-- metadata column must NOT exist
-- is_high_value must be true/false
-- placed_on must be DATE
```

---

### Stream: `public.payments` → `analytics.payment_facts`

**Step 1 — Pre-create destination table**
```sql
CREATE TABLE analytics.payment_facts (
    payment_id    UUID        PRIMARY KEY,
    order_ref     UUID,                   -- source: order_id (renamed)
    pay_method    TEXT,                   -- source: method (renamed)
    pay_amount    NUMERIC(10,2),          -- source: amount (renamed)
    pay_status    TEXT,                   -- source: status (renamed)
    is_successful BOOLEAN,               -- derived: status = 'paid'
    paid_on       DATE                   -- source: paid_at → DATE
);
```

**Step 5 — dbt SQL**
```sql
SELECT
    id                          AS payment_id,
    order_id                    AS order_ref,
    method                      AS pay_method,
    amount                      AS pay_amount,
    status                      AS pay_status,
    status = 'paid'             AS is_successful,
    paid_at::DATE               AS paid_on
FROM {{ source('raw', 'public__payments') }}
```

---

## Pipeline 7 — MySQL Source → PostgreSQL Destination

---

### Stream: `mydb.products` → `analytics.product_master`

**Step 1 — Pre-create destination table**
```sql
CREATE TABLE analytics.product_master (
    product_id    INTEGER     PRIMARY KEY,  -- source: id
    product_name  TEXT,                      -- source: name (renamed)
    product_sku   TEXT,                      -- source: sku (renamed)
    unit_price    NUMERIC(10,2),             -- source: price (same type, renamed)
    is_active     BOOLEAN,                   -- source: active TINYINT(1) → BOOLEAN
    added_on      DATE                       -- source: created_at DATETIME → DATE
);
```

**Step 4 — Normalisation rules**
```json
[
  { "rule_type": "cast", "table": "mydb.products", "column": "active", "cast_to": "boolean" },
  { "rule_type": "exclude", "table": "mydb.products", "column": "description" }
]
```

**Step 5 — dbt SQL**
```sql
SELECT
    id                AS product_id,
    name              AS product_name,
    sku               AS product_sku,
    price             AS unit_price,
    active            AS is_active,
    created_at::DATE  AS added_on
FROM {{ source('raw', 'mydb__products') }}
WHERE sku IS NOT NULL
```
> Normalisation: TINYINT cast → BOOLEAN before dbt SQL. `description` excluded at normalisation layer.

**Step 8 — Verify**
```sql
SELECT is_active FROM analytics.product_master LIMIT 5;
-- must be true/false, not 1/0
SELECT column_name FROM information_schema.columns
WHERE table_name = 'product_master';
-- 'description' must NOT appear
```

---

### Stream: `mydb.inventory` → `analytics.stock_levels`

**Step 1 — Pre-create destination table**
```sql
CREATE TABLE analytics.stock_levels (
    inventory_id   INTEGER     PRIMARY KEY,
    product_ref    INTEGER,               -- source: product_id (renamed)
    qty_on_hand    INTEGER,               -- source: quantity (renamed)
    warehouse_code TEXT,                  -- source: warehouse (renamed)
    stock_status   TEXT,                  -- derived: CASE WHEN quantity
    refreshed_on   DATE                   -- source: updated_at → DATE
);
```

**Step 5 — dbt SQL**
```sql
SELECT
    id                                               AS inventory_id,
    product_id                                       AS product_ref,
    quantity                                         AS qty_on_hand,
    warehouse                                        AS warehouse_code,
    CASE
        WHEN quantity = 0  THEN 'out_of_stock'
        WHEN quantity < 10 THEN 'low_stock'
        ELSE                    'in_stock'
    END                                              AS stock_status,
    updated_at::DATE                                 AS refreshed_on
FROM {{ source('raw', 'mydb__inventory') }}
```
> Derived TEXT column from INTEGER comparison. Multiple renames. DATETIME → DATE.

---

### Stream: `mydb.categories` → `analytics.category_tree`

**Step 1 — Pre-create destination table**
```sql
CREATE TABLE analytics.category_tree (
    category_id   INTEGER     PRIMARY KEY,
    category_name TEXT,                    -- source: name (renamed)
    url_slug      TEXT,                    -- source: slug (renamed)
    parent_ref    INTEGER,                 -- source: parent_id (renamed, nullable)
    is_root       BOOLEAN,                 -- derived: parent_id IS NULL
    added_on      DATE                     -- source: created_at → DATE
);
```

**Step 5 — dbt SQL**
```sql
SELECT
    id              AS category_id,
    name            AS category_name,
    slug            AS url_slug,
    parent_id       AS parent_ref,
    parent_id IS NULL AS is_root,
    created_at::DATE  AS added_on
FROM {{ source('raw', 'mydb__categories') }}
```

---

## Pipeline 8 — MariaDB Source → PostgreSQL Destination

---

### Stream: `app.events` → `analytics.event_facts`

> `payload` JSON with 10 keys. Destination extracts 4, adds derived columns.

**Step 1 — Pre-create destination table**
```sql
CREATE TABLE analytics.event_facts (
    event_id      BIGINT      PRIMARY KEY,
    actor_id      TEXT,                     -- source: user_id (renamed)
    action        TEXT,                     -- source: action
    client_ip     TEXT,                     -- from JSON: payload->>'ip_address'
    browser       TEXT,                     -- from JSON: payload->>'browser'
    http_status   INTEGER,                  -- from JSON: payload->>'status_code' → INTEGER
    is_success    BOOLEAN,                  -- derived: status_code between 200-299
    event_date    DATE,                     -- source: occurred_at DATETIME → DATE
    event_hour    INTEGER                   -- derived: EXTRACT(hour FROM occurred_at)
    -- os, duration_ms, session_id, referrer, page_url excluded from payload
);
```

**Step 5 — dbt SQL**
```sql
SELECT
    id                                                              AS event_id,
    user_id                                                         AS actor_id,
    action,
    payload->>'ip_address'                                          AS client_ip,
    payload->>'browser'                                             AS browser,
    CAST(NULLIF(payload->>'status_code','') AS INTEGER)             AS http_status,
    CAST(NULLIF(payload->>'status_code','') AS INTEGER) BETWEEN 200 AND 299 AS is_success,
    occurred_at::DATE                                               AS event_date,
    EXTRACT(hour FROM occurred_at)::INTEGER                         AS event_hour
FROM {{ source('raw', 'app__events') }}
WHERE payload IS NOT NULL
  AND payload->>'action' IS NOT NULL
```
> JSON: 2 TEXT keys + 1 INTEGER key from 10-key payload. 2 derived columns. 6 payload keys excluded.

---

### Stream: `app.logs` → `analytics.log_facts`

> `context` JSON object. Source ENUM → destination TEXT. Derived severity INTEGER.

**Step 1 — Pre-create destination table**
```sql
CREATE TABLE analytics.log_facts (
    log_id        BIGINT      PRIMARY KEY,
    log_level     TEXT,                     -- source: level ENUM → TEXT (renamed)
    severity      INTEGER,                  -- derived: ERROR=4, WARN=3, INFO=2, DEBUG=1
    log_message   TEXT,                     -- source: message (renamed)
    request_id    TEXT,                     -- from JSON: context->>'request_id'
    ctx_user_id   TEXT,                     -- from JSON: context->>'user_id'
    logged_on     DATE,                     -- source: logged_at → DATE
    log_hour      INTEGER                   -- derived: EXTRACT(hour FROM logged_at)
);
```

**Step 5 — dbt SQL**
```sql
SELECT
    id                                                                 AS log_id,
    level                                                              AS log_level,
    CASE level
        WHEN 'ERROR' THEN 4
        WHEN 'WARN'  THEN 3
        WHEN 'INFO'  THEN 2
        ELSE              1
    END                                                                AS severity,
    message                                                            AS log_message,
    context->>'request_id'                                             AS request_id,
    context->>'user_id'                                                AS ctx_user_id,
    logged_at::DATE                                                    AS logged_on,
    EXTRACT(hour FROM logged_at)::INTEGER                              AS log_hour
FROM {{ source('raw', 'app__logs') }}
```
> ENUM → TEXT (rename). TEXT ENUM → INTEGER severity (CASE). JSON: 2 keys extracted. 2 derived columns.

---

### Stream: `app.sessions` → `analytics.session_facts`

**Step 1 — Pre-create destination table**
```sql
CREATE TABLE analytics.session_facts (
    session_id     TEXT        PRIMARY KEY,
    user_ref       TEXT,                    -- source: user_id (renamed)
    ip_address     TEXT,
    is_active      BOOLEAN,                 -- source: active TINYINT(1) → BOOLEAN (renamed)
    duration_mins  INTEGER,                 -- derived: DATEDIFF minutes
    started_on     DATE,                    -- source: started_at → DATE
    ended_on       DATE                     -- source: ended_at → DATE (nullable)
);
```

**Step 5 — dbt SQL**
```sql
SELECT
    id                                                                AS session_id,
    user_id                                                           AS user_ref,
    ip_address,
    active = 1                                                        AS is_active,
    CASE
        WHEN ended_at IS NOT NULL
        THEN DATEDIFF('minute', started_at, ended_at)
        ELSE NULL
    END                                                               AS duration_mins,
    started_at::DATE                                                  AS started_on,
    ended_at::DATE                                                    AS ended_on
FROM {{ source('raw', 'app__sessions') }}
```

---

## Pipeline 9 — SQLite Source → PostgreSQL Destination

---

### Stream: `main.tasks` → `analytics.task_board`

**Step 1 — Pre-create destination table**
```sql
CREATE TABLE analytics.task_board (
    task_id        INTEGER     PRIMARY KEY,
    task_title     TEXT,                     -- source: title (renamed)
    task_status    TEXT,                     -- source: status (renamed)
    urgency        TEXT,                     -- derived: CASE priority
    due_on         DATE,                     -- source: due_date TEXT → DATE
    created_on     TIMESTAMPTZ,              -- source: created_at TEXT → TIMESTAMPTZ
    updated_on     TIMESTAMPTZ               -- source: updated_at TEXT → TIMESTAMPTZ
);
```

**Step 5 — dbt SQL**
```sql
SELECT
    id                                              AS task_id,
    title                                           AS task_title,
    status                                          AS task_status,
    CASE priority
        WHEN 3 THEN 'high'
        WHEN 2 THEN 'medium'
        ELSE       'low'
    END                                             AS urgency,
    CAST(due_date AS DATE)                          AS due_on,
    CAST(created_at AS TIMESTAMPTZ)                 AS created_on,
    CAST(updated_at AS TIMESTAMPTZ)                 AS updated_on
FROM {{ source('raw', 'main__tasks') }}
```
> SQLite TEXT dates → TIMESTAMPTZ/DATE. Integer priority → TEXT label. Multiple renames.

---

### Stream: `main.notes` → `analytics.task_notes`

**Step 1 — Pre-create destination table**
```sql
CREATE TABLE analytics.task_notes (
    note_id     INTEGER     PRIMARY KEY,
    task_ref    INTEGER,              -- source: task_id (renamed)
    note_body   TEXT,                 -- source: content (renamed)
    added_on    DATE                  -- source: created_at TEXT → DATE
);
```

**Step 5 — dbt SQL**
```sql
SELECT
    id            AS note_id,
    task_id       AS task_ref,
    content       AS note_body,
    CAST(created_at AS DATE) AS added_on
FROM {{ source('raw', 'main__notes') }}
WHERE content IS NOT NULL
  AND TRIM(content) != ''
```

---

### Stream: `main.tags` → `analytics.tag_master`

**Step 1 — Pre-create destination table**
```sql
CREATE TABLE analytics.tag_master (
    tag_id    INTEGER PRIMARY KEY,
    tag_name  TEXT,              -- source: name (renamed)
    hex_color TEXT               -- source: color (renamed, validated)
);
```

**Step 5 — dbt SQL**
```sql
SELECT
    id                                                   AS tag_id,
    name                                                 AS tag_name,
    CASE
        WHEN color LIKE '#%' THEN color
        ELSE CONCAT('#', color)
    END                                                  AS hex_color
FROM {{ source('raw', 'main__tags') }}
```
> Derived: ensure hex color always has `#` prefix.

---

## Pipeline 10 — CockroachDB Source → PostgreSQL Destination

---

### Stream: `public.accounts` → `analytics.account_ledger`

**Step 1 — Pre-create destination table**
```sql
CREATE TABLE analytics.account_ledger (
    account_id    UUID        PRIMARY KEY,
    email         TEXT,
    plan_tier     TEXT,                    -- source: plan (renamed)
    balance_usd   NUMERIC(12,2),           -- source: balance (renamed)
    tier_rank     INTEGER,                 -- derived: CASE plan
    is_paid       BOOLEAN,                 -- derived: plan != 'free'
    joined_on     DATE,                    -- source: created_at → DATE
    updated_on    DATE                     -- source: updated_at → DATE
);
```

**Step 5 — dbt SQL**
```sql
SELECT
    id                                            AS account_id,
    email,
    plan                                          AS plan_tier,
    balance                                       AS balance_usd,
    CASE plan
        WHEN 'enterprise' THEN 4
        WHEN 'pro'        THEN 3
        WHEN 'starter'    THEN 2
        ELSE                   1
    END                                           AS tier_rank,
    plan != 'free'                                AS is_paid,
    created_at::DATE                              AS joined_on,
    updated_at::DATE                              AS updated_on
FROM {{ source('raw', 'public__accounts') }}
```
> TEXT → INTEGER rank (CASE). Derived BOOLEAN. Multiple renames. TIMESTAMPTZ → DATE × 2.

---

### Stream: `public.transactions` → `analytics.txn_facts`

**Step 1 — Pre-create destination table**
```sql
CREATE TABLE analytics.txn_facts (
    txn_id       UUID        PRIMARY KEY,
    account_ref  UUID,                     -- source: account_id (renamed)
    txn_amount   NUMERIC(12,2),            -- source: amount (renamed)
    txn_type     TEXT,                     -- source: type (renamed)
    is_credit    BOOLEAN,                  -- derived: type = 'credit'
    note         TEXT,                     -- source: description (renamed)
    txn_date     DATE,                     -- source: created_at → DATE
    txn_month    TEXT                      -- derived: STRFTIME month label
);
```

**Step 5 — dbt SQL**
```sql
SELECT
    id                                              AS txn_id,
    account_id                                      AS account_ref,
    amount                                          AS txn_amount,
    type                                            AS txn_type,
    type = 'credit'                                 AS is_credit,
    description                                     AS note,
    created_at::DATE                                AS txn_date,
    DATE_TRUNC('month', created_at)::TEXT           AS txn_month
FROM {{ source('raw', 'public__transactions') }}
WHERE amount > 0
```

---

### Stream: `public.sessions` → `analytics.session_registry`

**Step 1 — Pre-create destination table**
```sql
CREATE TABLE analytics.session_registry (
    session_id    UUID        PRIMARY KEY,
    user_ref      UUID,                    -- source: user_id (renamed)
    ip_address    TEXT,
    browser_info  TEXT,                    -- source: user_agent (renamed)
    is_active     BOOLEAN,                 -- source: active (renamed)
    duration_mins INTEGER,                 -- derived: DATEDIFF minutes
    started_on    DATE,                    -- source: started_at → DATE
    ended_on      DATE                     -- source: ended_at → DATE
);
```

**Step 5 — dbt SQL**
```sql
SELECT
    id                                                               AS session_id,
    user_id                                                          AS user_ref,
    ip_address,
    user_agent                                                       AS browser_info,
    active                                                           AS is_active,
    CASE
        WHEN ended_at IS NOT NULL
        THEN DATEDIFF('minute', started_at, ended_at)
        ELSE NULL
    END                                                              AS duration_mins,
    started_at::DATE                                                 AS started_on,
    ended_at::DATE                                                   AS ended_on
FROM {{ source('raw', 'public__sessions') }}
```

---
---

# PART C — Universal Edge Case Steps (apply to any pipeline)

---

## Edge Case EC-1 — JSON key is NULL in some rows

**Step 5 — dbt SQL (safe extraction with COALESCE):**
```sql
SELECT
    id,
    COALESCE(payload->>'action', 'unknown')               AS action,
    payload->>'user_id'                                   AS user_id,  -- may be NULL
    NULLIF(TRIM(payload->>'ip_address'), '')               AS ip_address -- empty→NULL
FROM {{ source('raw', 'public__events') }}
```

**Step 8 — Verify:**
```sql
SELECT COUNT(*) FROM analytics.event_facts WHERE action = 'unknown';  -- safe default rows
SELECT COUNT(*) FROM analytics.event_facts WHERE ip_address = '';      -- must be 0
```

---

## Edge Case EC-2 — Source has column; destination intentionally has fewer

> Source: 20 columns. Destination: 4 columns only.

**Step 1:** Create destination table with only the 4 needed columns.  
**Step 5:** SELECT only those 4 columns in dbt SQL.  
**Step 6:** Validate SQL → green (4 outputs = 4 dest columns).  
**Step 8:** Verify: `\d analytics.<table>` shows exactly 4 columns.

---

## Edge Case EC-3 — Type mismatch → Phase 0 failure

**Step 1:** Create destination with wrong column type:
```sql
CREATE TABLE analytics.bad_type_test (id TEXT PRIMARY KEY, amount INTEGER);
```
**Step 5:** dbt SQL outputs `amount NUMERIC(10,2)` (decimal).  
**Step 7:** Run → **Expected:** Phase 0 fails. Error names the column and type conflict. `rows_written = 0`.

---

## Edge Case EC-4 — Extra NOT NULL dest column not in dbt SQL

**Step 1:**
```sql
CREATE TABLE analytics.strict_test (
    id TEXT PRIMARY KEY, amount NUMERIC, required_col TEXT NOT NULL
);
```
**Step 5:** dbt SQL: `SELECT id, amount FROM ...` (no `required_col`).  
**Step 7:** Run → **Expected:** Phase 0 fails: `"column required_col not present"`.

---

## Edge Case EC-5 — Empty string JSON value → NULL

**Step 5:**
```sql
SELECT
    id,
    NULLIF(TRIM(metadata->>'discount_code'), '')  AS discount_code,
    NULLIF(TRIM(metadata->>'channel'), '')         AS channel
FROM {{ source('raw', 'public__orders') }}
```
**Step 8:**
```sql
SELECT COUNT(*) FROM analytics.order_facts WHERE discount_code = '';
-- Must be 0
```
