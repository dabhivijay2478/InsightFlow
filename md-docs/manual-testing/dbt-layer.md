# dbt Layer — Manual Testing Guide

All pipeline SQL models run through **dbt-duckdb** in Phase 2. This guide covers every SQL pattern you must verify manually.

---

## How the dbt Layer Works

```
Source data extracted into DuckDB staging file
        ↓
dbt-duckdb reads from {{ source('raw', '<duckdb_table_name>') }}
        ↓
SQL model transforms data → output lands in DuckDB analytics schema
        ↓
Delivery phase writes DuckDB output → destination DB table
```

Key facts:
- Mode is always `"ui_sql"` — dbt project is generated on the fly from the SQL models you write in the builder.
- `sources.yml` is auto-generated; each stream is registered as `raw.<duckdb_table_name>`.
- Output table name is auto-generated (`dim_{pipelineId}_{nodeId}_{streamKey}`), not user-editable.
- Destination table (`dest_table = "schema.table"`) **must pre-exist** in the destination DB.

---

## DuckDB Source Reference

Always use the double-underscore staging name inside `{{ source('raw', '...') }}`.

| Stream key (source panel) | dbt source reference |
|--------------------------|---------------------|
| `stripe.customers` | `{{ source('raw', 'stripe__customers') }}` |
| `stripe.charges` | `{{ source('raw', 'stripe__charges') }}` |
| `stripe.invoices` | `{{ source('raw', 'stripe__invoices') }}` |
| `stripe.subscriptions` | `{{ source('raw', 'stripe__subscriptions') }}` |
| `stripe.products` | `{{ source('raw', 'stripe__products') }}` |
| `stripe.prices` | `{{ source('raw', 'stripe__prices') }}` |
| `stripe.events` | `{{ source('raw', 'stripe__events') }}` |
| `stripe.balance_transactions` | `{{ source('raw', 'stripe__balance_transactions') }}` |
| `stripe.coupons` | `{{ source('raw', 'stripe__coupons') }}` |
| `stripe.payment_intents` | `{{ source('raw', 'stripe__payment_intents') }}` |
| `stripe.payment_methods` | `{{ source('raw', 'stripe__payment_methods') }}` |
| `stripe.refunds` | `{{ source('raw', 'stripe__refunds') }}` |
| `stripe.disputes` | `{{ source('raw', 'stripe__disputes') }}` |
| `stripe.payouts` | `{{ source('raw', 'stripe__payouts') }}` |
| `stripe.plans` | `{{ source('raw', 'stripe__plans') }}` |
| `stripe.tax_rates` | `{{ source('raw', 'stripe__tax_rates') }}` |
| `stripe.credit_notes` | `{{ source('raw', 'stripe__credit_notes') }}` |
| `stripe.promotion_codes` | `{{ source('raw', 'stripe__promotion_codes') }}` |
| `stripe.setup_intents` | `{{ source('raw', 'stripe__setup_intents') }}` |
| `shopify.products` | `{{ source('raw', 'shopify__products') }}` |
| `shopify.orders` | `{{ source('raw', 'shopify__orders') }}` |
| `shopify.customers` | `{{ source('raw', 'shopify__customers') }}` |
| `shopify.draft_orders` | `{{ source('raw', 'shopify__draft_orders') }}` |
| `shopify.custom_collections` | `{{ source('raw', 'shopify__custom_collections') }}` |
| `shopify.smart_collections` | `{{ source('raw', 'shopify__smart_collections') }}` |
| `shopify.pages` | `{{ source('raw', 'shopify__pages') }}` |
| `shopify.blogs` | `{{ source('raw', 'shopify__blogs') }}` |
| `shopify.articles` | `{{ source('raw', 'shopify__articles') }}` |
| `shopify.locations` | `{{ source('raw', 'shopify__locations') }}` |
| `shopify.price_rules` | `{{ source('raw', 'shopify__price_rules') }}` |
| `shopify.themes` | `{{ source('raw', 'shopify__themes') }}` |
| `shopify.countries` | `{{ source('raw', 'shopify__countries') }}` |
| `shopify.collects` | `{{ source('raw', 'shopify__collects') }}` |
| `hubspot.contacts` | `{{ source('raw', 'hubspot__contacts') }}` |
| `hubspot.companies` | `{{ source('raw', 'hubspot__companies') }}` |
| `hubspot.deals` | `{{ source('raw', 'hubspot__deals') }}` |
| `hubspot.tickets` | `{{ source('raw', 'hubspot__tickets') }}` |
| `hubspot.products` | `{{ source('raw', 'hubspot__products') }}` |
| `hubspot.line_items` | `{{ source('raw', 'hubspot__line_items') }}` |
| `hubspot.quotes` | `{{ source('raw', 'hubspot__quotes') }}` |
| `hubspot.calls` | `{{ source('raw', 'hubspot__calls') }}` |
| `hubspot.emails` | `{{ source('raw', 'hubspot__emails') }}` |
| `hubspot.meetings` | `{{ source('raw', 'hubspot__meetings') }}` |
| `hubspot.notes` | `{{ source('raw', 'hubspot__notes') }}` |
| `hubspot.tasks` | `{{ source('raw', 'hubspot__tasks') }}` |
| `hubspot.feedback_submissions` | `{{ source('raw', 'hubspot__feedback_submissions') }}` |
| `hubspot.owners` | `{{ source('raw', 'hubspot__owners') }}` |
| `github.issues` | `{{ source('raw', 'github__issues') }}` |
| `github.pull_requests` | `{{ source('raw', 'github__pull_requests') }}` |
| `github.stargazers` | `{{ source('raw', 'github__stargazers') }}` |
| `github.events` | `{{ source('raw', 'github__events') }}` |
| `github.commits` | `{{ source('raw', 'github__commits') }}` |
| `github.branches` | `{{ source('raw', 'github__branches') }}` |
| `github.releases` | `{{ source('raw', 'github__releases') }}` |
| `github.tags` | `{{ source('raw', 'github__tags') }}` |
| `github.contributors` | `{{ source('raw', 'github__contributors') }}` |
| `github.milestones` | `{{ source('raw', 'github__milestones') }}` |
| `github.labels` | `{{ source('raw', 'github__labels') }}` |
| `github.forks` | `{{ source('raw', 'github__forks') }}` |
| `notion.databases` | `{{ source('raw', 'notion__databases') }}` |
| `notion.pages` | `{{ source('raw', 'notion__pages') }}` |
| `notion.users` | `{{ source('raw', 'notion__users') }}` |
| `public.users` | `{{ source('raw', 'public__users') }}` |
| `public.orders` | `{{ source('raw', 'public__orders') }}` |
| `public.payments` | `{{ source('raw', 'public__payments') }}` |
| `mydb.products` | `{{ source('raw', 'mydb__products') }}` |
| `mydb.inventory` | `{{ source('raw', 'mydb__inventory') }}` |
| `mydb.categories` | `{{ source('raw', 'mydb__categories') }}` |
| `app.events` | `{{ source('raw', 'app__events') }}` |
| `app.logs` | `{{ source('raw', 'app__logs') }}` |
| `app.sessions` | `{{ source('raw', 'app__sessions') }}` |
| `main.tasks` | `{{ source('raw', 'main__tasks') }}` |
| `main.notes` | `{{ source('raw', 'main__notes') }}` |
| `main.tags` | `{{ source('raw', 'main__tags') }}` |
| `public.sessions` | `{{ source('raw', 'public__sessions') }}` |
| `public.accounts` | `{{ source('raw', 'public__accounts') }}` |
| `public.transactions` | `{{ source('raw', 'public__transactions') }}` |

---

## Scenario DBT-1 — SELECT * (Default / Pass-through)

**When to use:** You want all columns from the source as-is.

**SQL:**
```sql
SELECT
    *
FROM {{ source('raw', 'stripe__customers') }}
```

**Validate:** Click **Validate SQL** in the builder.
- ✅ Green — all source columns match destination columns.
- ❌ Red with column name — destination table is missing that column. Add it with `ALTER TABLE`.

**Run result:**
- `dbt_models_run = 1`
- `rows_written > 0`
- Destination table: `SELECT * FROM analytics.stripe_customers_hd LIMIT 5;`

---

## Scenario DBT-2 — Column Selection (subset of fields)

**When to use:** Destination table has only selected columns; you don't want all source fields.

**SQL (Stripe customers → selected columns):**
```sql
SELECT
    id,
    email,
    name,
    currency,
    created
FROM {{ source('raw', 'stripe__customers') }}
```

**Destination DDL:**
```sql
CREATE TABLE analytics.stripe_customers_slim (
    id       TEXT PRIMARY KEY,
    email    TEXT,
    name     TEXT,
    currency TEXT,
    created  BIGINT
);
```

**Verify:** Only the 5 selected columns exist in destination; no extra columns.

---

## Scenario DBT-3 — JSON / JSONB Key Filtering

**When to use:** Source has a JSON column with many keys; destination stores only specific keys as flat columns.

**Example — Stripe `events` stream has `data` (JSONB) with many keys; extract 3:**
```sql
SELECT
    id,
    type                         AS event_type,
    data->>'object'              AS object_type,
    data->>'currency'            AS currency,
    created
FROM {{ source('raw', 'stripe__events') }}
WHERE data->>'object' IS NOT NULL
```

**Destination DDL:**
```sql
CREATE TABLE analytics.stripe_events_flat (
    id          TEXT PRIMARY KEY,
    event_type  TEXT,
    object_type TEXT,
    currency    TEXT,
    created     BIGINT
);
```

**Verify:**
```sql
SELECT id, event_type, object_type, currency
FROM analytics.stripe_events_flat
LIMIT 10;
-- All 4 columns populated; no JSON blobs; extra source keys absent
```

---

## Scenario DBT-4 — JSON Key Filtering from DB Source (10 keys → 3)

**When to use:** DB source table has a `jsonb` column (e.g., `payload`) with 10 keys; destination keeps only 3.

Source column layout (e.g., `public.events.payload`):
```json
{
  "user_id": "u-123",
  "action": "click",
  "ip_address": "1.2.3.4",
  "browser": "Chrome",
  "os": "macOS",
  "duration_ms": 342,
  "session_id": "s-abc",
  "referrer": "google.com",
  "page_url": "/pricing",
  "status_code": 200
}
```

**dbt SQL (keeps only `user_id`, `action`, `ip_address`):**
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

**Destination DDL:**
```sql
CREATE TABLE analytics.events_filtered (
    id          UUID PRIMARY KEY,
    user_id     TEXT,
    action      TEXT,
    ip_address  TEXT,
    occurred_at TIMESTAMPTZ
);
```

**Verify:**
```sql
SELECT id, user_id, action, ip_address FROM analytics.events_filtered LIMIT 5;
-- browser, os, duration_ms, session_id, referrer, page_url, status_code must NOT appear
```

---

## Scenario DBT-5 — COALESCE / NULL Handling

**SQL:**
```sql
SELECT
    id,
    COALESCE(status, 'unknown')   AS status_safe,
    COALESCE(amount, 0.00)        AS amount_safe,
    placed_at
FROM {{ source('raw', 'public__orders') }}
```

**Verify:** No NULL values in `status_safe` or `amount_safe` columns in destination.

---

## Scenario DBT-6 — CAST / Type Coercion

**SQL:**
```sql
SELECT
    id,
    CAST(price AS BIGINT)            AS price_cents,
    CASE WHEN active = true THEN 1
         ELSE 0 END                  AS is_active_int,
    CAST(created_at AS DATE)         AS created_date
FROM {{ source('raw', 'shopify__products') }}
```

**Destination DDL:**
```sql
CREATE TABLE analytics.products_typed (
    id            TEXT PRIMARY KEY,
    price_cents   BIGINT,
    is_active_int INTEGER,
    created_date  DATE
);
```

**Verify:** `price_cents` is numeric, `is_active_int` is 0 or 1, `created_date` has no time component.

---

## Scenario DBT-7 — UPPER / String Transformation

**SQL:**
```sql
SELECT
    id,
    UPPER(email)   AS email_upper,
    LOWER(name)    AS name_lower,
    created
FROM {{ source('raw', 'stripe__customers') }}
```

**Verify:** `email_upper` is fully uppercase; `name_lower` is fully lowercase.

---

## Scenario DBT-8 — GROUP BY Aggregate (Metrics Table)

**SQL:**
```sql
SELECT
    currency,
    COUNT(*)         AS customer_count,
    MIN(created)     AS first_seen,
    MAX(created)     AS last_seen
FROM {{ source('raw', 'stripe__customers') }}
GROUP BY currency
```

**Destination DDL (no PK — append mode):**
```sql
CREATE TABLE analytics.customer_by_currency (
    currency       TEXT,
    customer_count BIGINT,
    first_seen     BIGINT,
    last_seen      BIGINT
);
```

**Expected callback:** `no_pk_warnings` array is populated. Amber banner visible in Run Status drawer.
**Verify:** One row per currency; `customer_count` is accurate.

---

## Scenario DBT-9 — Window Function (ROW_NUMBER / RANK)

**SQL:**
```sql
SELECT
    id,
    status,
    amount,
    ROW_NUMBER() OVER (PARTITION BY status ORDER BY placed_at DESC) AS rn,
    placed_at
FROM {{ source('raw', 'public__orders') }}
```

**Destination DDL:**
```sql
CREATE TABLE analytics.orders_ranked (
    id       UUID PRIMARY KEY,
    status   TEXT,
    amount   NUMERIC,
    rn       BIGINT,
    placed_at TIMESTAMPTZ
);
```

**Verify:** `rn = 1` rows are the most recent per status; all ranks are sequential.

---

## Scenario DBT-10 — JOIN Across Two Streams (Multi-stream pipeline)

**Requires:** Both `public.orders` AND `public.payments` selected in Source Panel.

**SQL model for orders + payments joined:**
```sql
SELECT
    o.id          AS order_id,
    o.amount      AS order_amount,
    o.status,
    p.id          AS payment_id,
    p.method      AS payment_method,
    o.placed_at
FROM {{ source('raw', 'public__orders') }} AS o
LEFT JOIN {{ source('raw', 'public__payments') }} AS p
    ON o.id = p.order_id
```

**Destination DDL:**
```sql
CREATE TABLE analytics.orders_with_payments (
    order_id       UUID PRIMARY KEY,
    order_amount   NUMERIC,
    status         TEXT,
    payment_id     UUID,
    payment_method TEXT,
    placed_at      TIMESTAMPTZ
);
```

**Verify:** Rows appear with matched payment data; unmatched orders have NULL payment columns.

---

## Scenario DBT-11 — DATE_TRUNC / Time Bucketing

**SQL:**
```sql
SELECT
    DATE_TRUNC('month', TO_TIMESTAMP(created)) AS month,
    COUNT(*)                                    AS charge_count,
    SUM(amount)                                 AS total_amount_cents
FROM {{ source('raw', 'stripe__charges') }}
GROUP BY 1
ORDER BY 1
```

**Destination DDL:**
```sql
CREATE TABLE analytics.charges_monthly (
    month             TIMESTAMPTZ,
    charge_count      BIGINT,
    total_amount_cents BIGINT
);
```

**Verify:** One row per calendar month; amounts are in cents (Stripe returns integers).

---

## Scenario DBT-12 — Incremental-Aware Filter (INCREMENTAL sync mode)

**SQL (only processes new rows that dlt extracted via cursor):**
```sql
SELECT
    id,
    customer_id,
    amount,
    currency,
    status,
    created
FROM {{ source('raw', 'stripe__charges') }}
WHERE created IS NOT NULL
```

**Source Panel:** Set sync mode to `INCREMENTAL`, replication key = `created`.

**Run 1:** All rows synced. `rows_written = N`.
**Run 2 (before new charges added):** `rows_written = 0` (no new rows since last cursor).
**Run 3 (after new charges added in Stripe):** Only new charges appear. `rows_written = M` where `M` = number of new charges.

---

## Scenario DBT-13 — Custom Filter / WHERE clause

**SQL:**
```sql
SELECT
    id,
    email,
    name,
    currency
FROM {{ source('raw', 'stripe__customers') }}
WHERE currency IN ('usd', 'eur', 'gbp')
  AND email IS NOT NULL
```

**Verify:** Destination contains only USD/EUR/GBP customers with non-null emails.

---

## Scenario DBT-14 — Validate SQL Error Cases

| Error type | SQL that triggers it | Expected inline error |
|-----------|---------------------|----------------------|
| Missing column | `SELECT nonexistent_col FROM {{ source('raw', 'stripe__customers') }}` | Column not found in dbt |
| Wrong source ref | `SELECT * FROM {{ source('raw', 'stripe__customer') }}` (typo) | Source not registered |
| Syntax error | `SELECT * FORM {{ source(...) }}` | dbt parse error with line number |
| Column mismatch | SQL outputs `currency` but destination has no `currency` column | `"column currency not present in analytics.hd"` |

---

## SQL Validation Checklist (Before Every Run)

- [ ] Every `{{ source('raw', '...') }}` name matches the DuckDB staging name exactly (double underscore).
- [ ] All SELECT output columns exist in the destination table (`ALTER TABLE` if needed).
- [ ] No `CREATE TABLE`, `DROP TABLE`, `INSERT INTO`, `UPDATE` or `DELETE` statements in the SQL.
- [ ] Primary key column is included in the SELECT if destination uses merge mode.
- [ ] Click **Validate SQL** and see green confirmation before saving.
