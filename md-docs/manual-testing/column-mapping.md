# Column Mapping, Type Coercion & JSON Extraction — Manual Testing Guide

This guide covers every scenario where source and destination columns **do not match**:

- Source has a JSON/JSONB column → destination stores only selected keys as flat `TEXT`/`VARCHAR` columns
- Source column names differ from destination column names
- Source data type differs from destination data type
- Source has more (or fewer) columns than the destination

All transformations happen in the **dbt SQL model** (Phase 2 / DuckDB). The normalisation rules panel handles pre-processing (rename, cast, exclude) before dbt runs.

---

## Part 1 — JSON Column in Source → Flat TEXT/VARCHAR Columns in Destination

### CM-1 — JSONB (PostgreSQL source) → multiple TEXT columns (PostgreSQL dest)

**Source table:** `public.events`
```
payload (JSONB) contains 10 keys:
  user_id, action, ip_address, browser, os,
  duration_ms, session_id, referrer, page_url, status_code
```

**Destination only needs 3 keys as flat columns:**
```sql
CREATE TABLE analytics.events_flat (
    id          UUID    PRIMARY KEY,
    user_id     TEXT,               -- extracted from payload->>'user_id'
    action      TEXT,               -- extracted from payload->>'action'
    ip_address  TEXT,               -- extracted from payload->>'ip_address'
    occurred_at TIMESTAMPTZ
);
```

**dbt SQL:**
```sql
SELECT
    id,
    payload->>'user_id'     AS user_id,
    payload->>'action'      AS action,
    payload->>'ip_address'  AS ip_address,
    occurred_at
FROM {{ source('raw', 'public__events') }}
WHERE payload->>'action' IS NOT NULL
```

**Verify:**
```sql
-- Only 4 columns in destination (id, user_id, action, ip_address, occurred_at)
SELECT column_name FROM information_schema.columns
WHERE table_name = 'events_flat' ORDER BY ordinal_position;

-- payload blob must NOT be a column
-- browser, os, duration_ms, session_id, referrer, page_url, status_code must NOT appear
SELECT user_id, action, ip_address FROM analytics.events_flat LIMIT 5;
```

---

### CM-2 — JSON (MySQL / MariaDB source) → flat VARCHAR columns (MySQL dest)

**Source:** `app.events.payload` (JSON column, MariaDB)

dbt runs inside DuckDB, so use DuckDB JSON operators regardless of source DB:

```sql
SELECT
    id,
    payload->>'user_id'                     AS user_id,
    payload->>'action'                       AS action,
    CAST(payload->>'status_code' AS INTEGER) AS status_code,
    occurred_at
FROM {{ source('raw', 'app__events') }}
WHERE payload->>'action' IS NOT NULL
```

**Destination DDL (MySQL):**
```sql
CREATE TABLE analytics.events_flat (
    id          BIGINT       PRIMARY KEY,
    user_id     VARCHAR(36),
    action      VARCHAR(100),
    status_code INT,
    occurred_at DATETIME
) ENGINE=InnoDB;
```

**Verify:**
```sql
SELECT user_id, action, status_code FROM analytics.events_flat LIMIT 5;
-- No JSON blob column; status_code is integer
```

---

### CM-3 — 10-key JSON → 3 TEXT columns + 1 INTEGER column

**Source column:** `orders.metadata` (JSONB):
```json
{
  "payment_method": "stripe",
  "gateway":        "payments-v2",
  "currency":       "usd",
  "tax_rate":       0.08,
  "discount_code":  "SAVE20",
  "shipping_provider": "fedex",
  "tracking_number":   "1Z999...",
  "warehouse_id":      "WH-42",
  "region":            "us-east",
  "channel":           "web"
}
```

**Destination needs only:** `payment_method` (TEXT), `currency` (TEXT), `channel` (TEXT), `tax_rate` (NUMERIC)

```sql
-- Destination DDL
CREATE TABLE analytics.orders_meta_slim (
    id             UUID    PRIMARY KEY,
    payment_method TEXT,
    currency       TEXT,
    channel        TEXT,
    tax_rate       NUMERIC(5,4),
    placed_at      TIMESTAMPTZ
);
```

**dbt SQL:**
```sql
SELECT
    id,
    metadata->>'payment_method'              AS payment_method,
    metadata->>'currency'                    AS currency,
    metadata->>'channel'                     AS channel,
    CAST(metadata->>'tax_rate' AS NUMERIC)   AS tax_rate,
    placed_at
FROM {{ source('raw', 'public__orders') }}
WHERE metadata IS NOT NULL
  AND metadata->>'payment_method' IS NOT NULL
```

**Verify (7 keys must be absent):**
```sql
SELECT payment_method, currency, channel, tax_rate FROM analytics.orders_meta_slim LIMIT 5;
-- gateway, discount_code, shipping_provider, tracking_number,
-- warehouse_id, region must NOT appear as columns
```

---

### CM-4 — Nested JSON (2 levels deep) → flat TEXT columns

**Source:** `github.events.payload` — `payload.pull_request.head.ref` (2 levels deep)

```sql
SELECT
    id,
    type                                         AS event_type,
    actor->>'login'                              AS actor_login,
    payload->'pull_request'->'head'->>'ref'      AS head_branch,
    payload->'pull_request'->'base'->>'ref'      AS base_branch,
    payload->'pull_request'->>'state'            AS pr_state,
    created_at
FROM {{ source('raw', 'github__events') }}
WHERE type = 'PullRequestEvent'
  AND payload->'pull_request' IS NOT NULL
```

**Destination DDL:**
```sql
CREATE TABLE analytics.github_pr_events (
    id          TEXT PRIMARY KEY,
    event_type  TEXT,
    actor_login TEXT,
    head_branch TEXT,
    base_branch TEXT,
    pr_state    TEXT,
    created_at  TIMESTAMPTZ
);
```

**Verify:** `head_branch` and `base_branch` contain branch names (strings), not JSON objects.

---

### CM-5 — JSON Array → single TEXT (serialised subset)

When the destination has one TEXT column to hold a compact JSON subset:

```sql
SELECT
    id,
    type              AS event_type,
    -- Store a compact 2-key subset as JSON string
    JSON_OBJECT(
        'object', data->>'object',
        'currency', data->>'currency'
    )::TEXT           AS data_slim,
    created
FROM {{ source('raw', 'stripe__events') }}
```

**Destination DDL:**
```sql
CREATE TABLE analytics.stripe_events_slim (
    id          TEXT PRIMARY KEY,
    event_type  TEXT,
    data_slim   TEXT,   -- compact JSON string with only 2 keys
    created     BIGINT
);
```

---

### CM-6 — JSON array element extraction (first element only)

**Source:** `notion.databases.title` is a rich-text array:
```json
[{"type": "text", "plain_text": "My Project"}]
```

**dbt SQL:**
```sql
SELECT
    id,
    title->0->>'plain_text'  AS title_text,
    url,
    created_time
FROM {{ source('raw', 'notion__databases') }}
WHERE title->0->>'plain_text' IS NOT NULL
```

**Destination DDL:**
```sql
CREATE TABLE analytics.notion_databases_flat (
    id           TEXT PRIMARY KEY,
    title_text   TEXT,         -- plain text, not a JSON array
    url          TEXT,
    created_time TIMESTAMPTZ
);
```

---

## Part 2 — Source Column Names ≠ Destination Column Names

### CM-7 — Complete column rename: source names → destination names

**Source table** `public.users` has columns: `id`, `email`, `first_name`, `last_name`, `created_at`, `updated_at`  
**Destination table** uses different names:

```sql
CREATE TABLE analytics.user_profiles (
    user_uuid     UUID    PRIMARY KEY,   -- source: id
    email_address TEXT,                   -- source: email
    fname         TEXT,                   -- source: first_name
    lname         TEXT,                   -- source: last_name
    registered_at TIMESTAMPTZ,            -- source: created_at
    modified_at   TIMESTAMPTZ             -- source: updated_at
);
```

**dbt SQL (alias every column):**
```sql
SELECT
    id          AS user_uuid,
    email       AS email_address,
    first_name  AS fname,
    last_name   AS lname,
    created_at  AS registered_at,
    updated_at  AS modified_at
FROM {{ source('raw', 'public__users') }}
```

**Verify:**
```sql
SELECT column_name FROM information_schema.columns
WHERE table_name = 'user_profiles' ORDER BY ordinal_position;
-- Must show: user_uuid, email_address, fname, lname, registered_at, modified_at
-- Must NOT show: id, email, first_name, last_name, created_at, updated_at
```

---

### CM-8 — SaaS source names → business names (Stripe charges)

**Stripe `charges` columns:** `id`, `amount`, `currency`, `status`, `customer`, `created`  
**Destination business names:**

```sql
CREATE TABLE analytics.payment_ledger (
    payment_ref      TEXT    PRIMARY KEY,   -- source: id
    amount_cents     BIGINT,                -- source: amount
    payment_currency TEXT,                  -- source: currency
    payment_status   TEXT,                  -- source: status
    customer_ref     TEXT,                  -- source: customer
    charged_at       BIGINT                 -- source: created
);
```

**dbt SQL:**
```sql
SELECT
    id          AS payment_ref,
    amount      AS amount_cents,
    currency    AS payment_currency,
    status      AS payment_status,
    customer    AS customer_ref,
    created     AS charged_at
FROM {{ source('raw', 'stripe__charges') }}
WHERE status = 'succeeded'
```

---

### CM-9 — HubSpot hs_ prefix stripped

HubSpot columns are prefixed `hs_`. Destination uses clean names:

```sql
CREATE TABLE analytics.hs_deals_clean (
    deal_id       TEXT    PRIMARY KEY,     -- source: id
    deal_name     TEXT,                    -- source: dealname
    deal_amount   NUMERIC,                 -- source: amount
    stage         TEXT,                    -- source: dealstage
    pipeline      TEXT,                    -- source: pipeline
    close_date    TIMESTAMPTZ,             -- source: closedate
    created_date  TIMESTAMPTZ              -- source: createdate
);
```

**dbt SQL:**
```sql
SELECT
    id                       AS deal_id,
    dealname                 AS deal_name,
    CAST(amount AS NUMERIC)  AS deal_amount,
    dealstage                AS stage,
    pipeline,
    closedate                AS close_date,
    createdate               AS created_date
FROM {{ source('raw', 'hubspot__deals') }}
WHERE amount IS NOT NULL
```

---

### CM-10 — Normalisation rename + dbt alias (both layers)

You can rename in the normalisation panel first, then alias again in dbt SQL if needed.

**Normalisation rules:**
```json
[
  { "rule_type": "rename", "table": "stripe.customers", "column": "created", "destination_name": "created_unix" }
]
```

**dbt SQL (further alias to business name):**
```sql
SELECT
    id            AS customer_id,
    email         AS customer_email,
    name          AS customer_name,
    currency      AS billing_currency,
    created_unix  AS first_seen_ts    -- already renamed by normalisation rule
FROM {{ source('raw', 'stripe__customers') }}
```

**Destination DDL:**
```sql
CREATE TABLE analytics.customers_master (
    customer_id       TEXT PRIMARY KEY,
    customer_email    TEXT,
    customer_name     TEXT,
    billing_currency  TEXT,
    first_seen_ts     BIGINT
);
```

---

## Part 3 — Different Data Types: Source → Destination

### CM-11 — Source INTEGER (cents) → Destination NUMERIC (dollars)

**Source:** `stripe.charges.amount` = `2999` (integer, cents)  
**Destination:** `NUMERIC(10,2)` = `29.99` (dollars)

```sql
SELECT
    id,
    amount / 100.0          AS amount_dollars,
    currency,
    status,
    created
FROM {{ source('raw', 'stripe__charges') }}
```

**Destination DDL:**
```sql
CREATE TABLE analytics.charges_dollars (
    id             TEXT PRIMARY KEY,
    amount_dollars NUMERIC(10, 2),
    currency       TEXT,
    status         TEXT,
    created        BIGINT
);
```

**Verify:** `SELECT amount_dollars FROM analytics.charges_dollars LIMIT 5;` — values like `29.99`, not `2999`.

---

### CM-12 — Source TEXT → Destination INTEGER (safe cast)

**Source:** A column stored as `TEXT` that contains numeric strings (e.g., `"200"`, `"404"`).

```sql
SELECT
    id,
    CASE
        WHEN payload->>'status_code' ~ '^\d+$'
        THEN CAST(payload->>'status_code' AS INTEGER)
        ELSE NULL
    END  AS status_code,
    action,
    occurred_at
FROM {{ source('raw', 'public__events') }}
```

**Destination DDL:**
```sql
CREATE TABLE analytics.events_typed (
    id          UUID PRIMARY KEY,
    status_code INTEGER,
    action      TEXT,
    occurred_at TIMESTAMPTZ
);
```

**Edge case:** If source has `"abc"` in that column, the `CASE WHEN ~ '^\d+$'` guard returns NULL instead of crashing.

---

### CM-13 — Source BOOLEAN / TINYINT(1) → Destination INTEGER (0/1)

**Source (MySQL):** `active TINYINT(1)` (stored as 0 or 1)  
**Destination:** `is_active INTEGER`

```sql
SELECT
    id,
    name,
    price,
    CASE WHEN active = 1 THEN 1 ELSE 0 END  AS is_active,
    created_at
FROM {{ source('raw', 'mydb__products') }}
```

---

### CM-14 — Source BOOLEAN → Destination TEXT label

```sql
SELECT
    id,
    name,
    CASE WHEN active = true THEN 'active' ELSE 'inactive' END  AS status_label,
    created_at
FROM {{ source('raw', 'mydb__products') }}
```

**Destination DDL:**
```sql
CREATE TABLE analytics.products_labelled (
    id           INTEGER PRIMARY KEY,
    name         TEXT,
    status_label TEXT,
    created_at   TIMESTAMPTZ
);
```

---

### CM-15 — Source TIMESTAMPTZ → Destination DATE (truncation)

```sql
SELECT
    id,
    email,
    created_at::DATE   AS signup_date,
    updated_at::DATE   AS last_activity_date
FROM {{ source('raw', 'public__users') }}
```

**Destination DDL:**
```sql
CREATE TABLE analytics.users_dates (
    id                 UUID PRIMARY KEY,
    email              TEXT,
    signup_date        DATE,
    last_activity_date DATE
);
```

**Verify:** `SELECT signup_date FROM analytics.users_dates LIMIT 3;` — values like `2024-01-15`, no time component.

---

### CM-16 — Source BIGINT (Unix timestamp) → Destination TIMESTAMPTZ

Stripe `created` is a Unix epoch integer. Convert to timestamp:

```sql
SELECT
    id,
    email,
    name,
    currency,
    TO_TIMESTAMP(created)::TIMESTAMPTZ  AS created_at
FROM {{ source('raw', 'stripe__customers') }}
```

**Destination DDL:**
```sql
CREATE TABLE analytics.stripe_customers_ts (
    id         TEXT PRIMARY KEY,
    email      TEXT,
    name       TEXT,
    currency   TEXT,
    created_at TIMESTAMPTZ
);
```

**Verify:** `SELECT created_at FROM analytics.stripe_customers_ts LIMIT 3;` — human-readable timestamps.

---

### CM-17 — Source TEXT (ISO date string) → Destination TIMESTAMPTZ (SQLite source)

SQLite stores dates as TEXT (`"2024-01-15 10:30:00"`). Cast in dbt:

```sql
SELECT
    id,
    title,
    status,
    CAST(created_at  AS TIMESTAMPTZ)  AS created_at,
    CAST(updated_at  AS TIMESTAMPTZ)  AS updated_at
FROM {{ source('raw', 'main__tasks') }}
```

---

### CM-18 — Source NUMERIC → Destination TEXT (for display/category)

```sql
SELECT
    account_id,
    CASE
        WHEN net_balance > 1000  THEN 'high_value'
        WHEN net_balance > 0     THEN 'positive'
        WHEN net_balance = 0     THEN 'zero'
        ELSE                          'negative'
    END  AS balance_tier,
    net_balance
FROM {{ source('raw', 'public__accounts') }}
```

---

### CM-19 — Source JSON string → Destination parsed NUMERIC

`stripe.tax_rates.percentage` is returned as a float string by the API. Cast safely:

```sql
SELECT
    id,
    display_name,
    CAST(percentage AS NUMERIC(5, 4))  AS tax_rate,
    inclusive,
    active,
    created
FROM {{ source('raw', 'stripe__tax_rates') }}
```

---

## Part 4 — Different Column Sets: Source Has More/Fewer Columns Than Destination

### CM-20 — Source has 15 columns; destination uses only 4

**Source:** `stripe.customers` — 15+ columns returned by API  
**Destination:** only needs `id`, `email`, `currency`, `created`

```sql
CREATE TABLE analytics.customers_minimal (
    id       TEXT PRIMARY KEY,
    email    TEXT,
    currency TEXT,
    created  BIGINT
);
```

**dbt SQL (select only 4):**
```sql
SELECT id, email, currency, created
FROM {{ source('raw', 'stripe__customers') }}
```

**Verify:** `\d analytics.customers_minimal` (psql) — only 4 columns present; all others absent.

---

### CM-21 — Source has 6 columns; destination has 8 (derived columns added)

**Source:** `public.orders` — `id`, `user_id`, `amount`, `status`, `metadata`, `placed_at`  
**Destination has 8:** original 4 + 4 derived

```sql
CREATE TABLE analytics.orders_enriched (
    id              UUID    PRIMARY KEY,
    amount          NUMERIC,
    status          TEXT,
    placed_at       TIMESTAMPTZ,
    -- derived:
    amount_dollars  NUMERIC(10,2),
    payment_method  TEXT,
    currency        TEXT,
    is_high_value   BOOLEAN
);
```

**dbt SQL (adds 4 derived columns, drops `user_id` and `metadata`):**
```sql
SELECT
    id,
    amount,
    status,
    placed_at,
    amount / 100.0                            AS amount_dollars,
    metadata->>'payment_method'               AS payment_method,
    metadata->>'currency'                     AS currency,
    amount > 10000                            AS is_high_value
FROM {{ source('raw', 'public__orders') }}
WHERE metadata IS NOT NULL
```

---

### CM-22 — Source has 20 columns; destination has 5 with renamed + retyped columns

**Source:** `hubspot.contacts` (20+ properties)  
**Destination:** compact CRM mirror with 5 renamed columns

```sql
CREATE TABLE analytics.crm_contacts (
    crm_id      TEXT  PRIMARY KEY,   -- source: id
    email       TEXT,                 -- source: email (same)
    full_name   TEXT,                 -- derived: firstname || ' ' || lastname
    lead_status TEXT,                 -- source: hs_lead_status
    created_ts  TIMESTAMPTZ          -- source: createdate (type changed)
);
```

**dbt SQL:**
```sql
SELECT
    id                                              AS crm_id,
    email,
    CONCAT(firstname, ' ', lastname)               AS full_name,
    hs_lead_status                                  AS lead_status,
    CAST(createdate AS TIMESTAMPTZ)                AS created_ts
FROM {{ source('raw', 'hubspot__contacts') }}
WHERE email IS NOT NULL
```

---

### CM-23 — Source column split into two destination columns

**Source:** `public.users.email` contains `user@domain.com`  
**Destination has separate:** `email_user` (TEXT) + `email_domain` (TEXT)

```sql
SELECT
    id,
    SPLIT_PART(email, '@', 1)  AS email_user,
    SPLIT_PART(email, '@', 2)  AS email_domain,
    created_at
FROM {{ source('raw', 'public__users') }}
WHERE email LIKE '%@%'
```

**Destination DDL:**
```sql
CREATE TABLE analytics.users_email_split (
    id           UUID PRIMARY KEY,
    email_user   TEXT,
    email_domain TEXT,
    created_at   TIMESTAMPTZ
);
```

---

### CM-24 — Two source columns merged into one destination column

**Source:** `first_name` + `last_name`  
**Destination:** single `full_name` TEXT column

```sql
SELECT
    id,
    TRIM(CONCAT(
        COALESCE(firstname, ''),
        ' ',
        COALESCE(lastname, '')
    ))              AS full_name,
    email,
    createdate      AS created_at
FROM {{ source('raw', 'hubspot__contacts') }}
```

---

### CM-25 — Source has column; destination does not (exclusion via SELECT)

**Source:** `public.users` has `password_hash` (sensitive)  
**Destination DDL** does NOT have `password_hash` column:

```sql
CREATE TABLE analytics.users_public (
    id         UUID PRIMARY KEY,
    email      TEXT,
    first_name TEXT,
    last_name  TEXT,
    created_at TIMESTAMPTZ
    -- password_hash intentionally absent
);
```

**dbt SQL (simply don't SELECT it):**
```sql
SELECT id, email, first_name, last_name, created_at
FROM {{ source('raw', 'public__users') }}
```

**Verify:** `SELECT column_name FROM information_schema.columns WHERE table_name = 'users_public';` — `password_hash` absent.

---

## Part 5 — Edge Cases & Error Paths

### CM-26 — NULL JSON key → NULL in destination TEXT column

```sql
SELECT
    id,
    payload->>'user_id'     AS user_id,     -- may be NULL if key missing
    COALESCE(payload->>'action', 'unknown') AS action,  -- safe default
    occurred_at
FROM {{ source('raw', 'public__events') }}
```

**Verify:** `SELECT COUNT(*) FROM analytics.events_flat WHERE user_id IS NULL;` — may be > 0; `action` must never be NULL.

---

### CM-27 — JSON key exists but value is empty string → normalise to NULL

```sql
SELECT
    id,
    NULLIF(TRIM(metadata->>'discount_code'), '')   AS discount_code,
    NULLIF(TRIM(metadata->>'channel'), '')          AS channel,
    placed_at
FROM {{ source('raw', 'public__orders') }}
```

**Verify:** `SELECT discount_code FROM analytics.orders_meta_slim WHERE discount_code = '';` — must return 0 rows.

---

### CM-28 — Type mismatch causes Phase 0 column validation failure

**Scenario:** Destination `amount_dollars` column is `INTEGER`, but dbt SQL outputs `NUMERIC(10,2)`.

```sql
-- Destination (wrong type)
CREATE TABLE analytics.bad_types (
    id     TEXT PRIMARY KEY,
    amount INTEGER   -- ← should be NUMERIC
);
```

**dbt SQL outputs NUMERIC:**
```sql
SELECT id, amount / 100.0 AS amount FROM {{ source('raw', 'stripe__charges') }}
```

**Expected:** Phase 0 pre-flight detects type mismatch → run `failed` with column type error.

---

### CM-29 — Extra column in destination not in dbt output

**Destination has a column the dbt SQL does not SELECT:**

```sql
CREATE TABLE analytics.orders_extra (
    id             UUID PRIMARY KEY,
    amount         NUMERIC,
    extra_required TEXT NOT NULL   -- dbt SQL doesn't produce this
);
```

**Expected:** Phase 0 fails: `"column extra_required not present in analytics.orders_extra"`.

---

### CM-30 — JSON key value is a nested object (not a scalar)

`metadata->>'payment_details'` returns a JSON string `{"card":"visa","last4":"4242"}`, not a flat scalar.

**dbt SQL (store as TEXT):**
```sql
SELECT
    id,
    metadata->>'payment_details'                          AS payment_details_json,
    metadata->'payment_details'->>'card'                  AS card_brand,
    metadata->'payment_details'->>'last4'                 AS card_last4
FROM {{ source('raw', 'public__orders') }}
```

**Destination DDL:**
```sql
CREATE TABLE analytics.orders_card_details (
    id                   UUID PRIMARY KEY,
    payment_details_json TEXT,    -- full nested JSON as TEXT
    card_brand           TEXT,    -- extracted scalar
    card_last4           TEXT     -- extracted scalar
);
```

**Verify:** `card_brand` = `"visa"` (string), not `{"card":"visa","last4":"4242"}`.

---

## Summary: Mapping Pattern Reference

| Pattern | Source type | Destination type | dbt SQL approach |
|---------|------------|-----------------|-----------------|
| JSON key → flat TEXT | `JSONB` | `TEXT` | `col->>'key'` |
| JSON nested key → TEXT | `JSONB` | `TEXT` | `col->'k1'->>'k2'` |
| JSON key → INTEGER | `JSONB` | `INTEGER` | `CAST(col->>'key' AS INTEGER)` |
| JSON key → NUMERIC | `JSONB` | `NUMERIC` | `CAST(col->>'key' AS NUMERIC)` |
| JSON key → BOOLEAN | `JSONB` | `BOOLEAN` | `(col->>'key')::BOOLEAN` |
| JSON first array element | `JSONB[]` | `TEXT` | `col->0->>'field'` |
| Integer cents → dollars | `BIGINT` | `NUMERIC(10,2)` | `amount / 100.0` |
| Unix ts → TIMESTAMPTZ | `BIGINT` | `TIMESTAMPTZ` | `TO_TIMESTAMP(col)` |
| TEXT ISO date → TIMESTAMPTZ | `TEXT` | `TIMESTAMPTZ` | `CAST(col AS TIMESTAMPTZ)` |
| TIMESTAMPTZ → DATE | `TIMESTAMPTZ` | `DATE` | `col::DATE` |
| BOOLEAN → INTEGER 0/1 | `BOOL` | `INTEGER` | `CASE WHEN col THEN 1 ELSE 0 END` |
| BOOLEAN → TEXT label | `BOOL` | `TEXT` | `CASE WHEN col THEN 'yes' ELSE 'no' END` |
| Rename column | any | any | `col AS new_name` |
| Merge two → one | `TEXT` + `TEXT` | `TEXT` | `CONCAT(a, ' ', b)` |
| Split one → two | `TEXT` | `TEXT` + `TEXT` | `SPLIT_PART(col, '@', 1)` |
| Subset select (20 → 4) | many | few | list only needed cols |
| Add derived columns | fewer | more | compute in SELECT |
| Exclude sensitive col | exists in source | absent in dest | don't SELECT it |
| NULL safe extraction | nullable `JSONB` | `TEXT` (nullable) | `COALESCE(col->>'k', NULL)` |
| Empty string → NULL | `TEXT` `''` | `TEXT` NULL | `NULLIF(TRIM(col->>'k'), '')` |
