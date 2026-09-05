# Stripe Source — Manual Testing Guide

**Streams:** 63 registered streams across stable and feature-gated families

**Credential:** Secret or restricted key (`sk_test_...`, `sk_live_...`, `rk_test_...`, or `rk_live_...`)

**DuckDB prefix:** `stripe__`

The authoritative catalog and strategy matrix are documented in
`apps/server/arcyria-elt/saas_sources/stripe_analytics/README.md`. The table
below is the original core regression set, not the complete discovery catalog.

---

## Credential Setup

In MantrixFlow Connections → New source → Stripe:

```json
{ "api_key": "sk_test_51..." }
```

Test connection must return ✅ before building any pipeline.

---

## Core Regression Streams

| Stream | DuckDB staging name | Key columns | Supports INCREMENTAL |
|--------|-------------------|-------------|---------------------|
| `stripe.customers` | `stripe__customers` | `id`, `email`, `name`, `currency`, `created` | ✅ `created` |
| `stripe.charges` | `stripe__charges` | `id`, `amount`, `currency`, `status`, `customer`, `created` | ✅ `created` |
| `stripe.invoices` | `stripe__invoices` | `id`, `customer`, `amount_due`, `amount_paid`, `status`, `created` | ✅ `created` |
| `stripe.subscriptions` | `stripe__subscriptions` | `id`, `customer`, `status`, `plan`, `current_period_start`, `current_period_end` | ✅ `created` |
| `stripe.products` | `stripe__products` | `id`, `name`, `active`, `description`, `created` | ✅ `created` |
| `stripe.prices` | `stripe__prices` | `id`, `product`, `unit_amount`, `currency`, `type`, `recurring` | ✅ `created` |
| `stripe.events` | `stripe__events` | `id`, `type`, `data`, `created`, `livemode` | ✅ `created` |
| `stripe.balance_transactions` | `stripe__balance_transactions` | `id`, `amount`, `fee`, `net`, `currency`, `type`, `created` | ✅ `created` |
| `stripe.coupons` | `stripe__coupons` | `id`, `name`, `percent_off`, `amount_off`, `currency`, `duration`, `created` | ✅ `created` |
| `stripe.payment_intents` | `stripe__payment_intents` | `id`, `amount`, `currency`, `status`, `customer`, `created` | ✅ `created` |
| `stripe.payment_methods` | `stripe__payment_methods` | `id`, `type`, `customer`, `card`, `created` | ✅ `created` |
| `stripe.refunds` | `stripe__refunds` | `id`, `charge`, `amount`, `currency`, `status`, `created` | ✅ `created` |
| `stripe.disputes` | `stripe__disputes` | `id`, `charge`, `amount`, `currency`, `status`, `reason`, `created` | ✅ `created` |
| `stripe.payouts` | `stripe__payouts` | `id`, `amount`, `currency`, `status`, `arrival_date`, `created` | ✅ `created` |
| `stripe.plans` | `stripe__plans` | `id`, `product`, `amount`, `currency`, `interval`, `created` | ✅ `created` |
| `stripe.tax_rates` | `stripe__tax_rates` | `id`, `display_name`, `percentage`, `inclusive`, `active`, `created` | ✅ `created` |
| `stripe.credit_notes` | `stripe__credit_notes` | `id`, `invoice`, `amount`, `currency`, `status`, `created` | ✅ `created` |
| `stripe.promotion_codes` | `stripe__promotion_codes` | `id`, `code`, `coupon`, `active`, `created` | ✅ `created` |
| `stripe.setup_intents` | `stripe__setup_intents` | `id`, `customer`, `status`, `payment_method`, `created` | ✅ `created` |

---

## Scenario S-STR-1 — Full Table Sync, Single Stream (`customers`)

**Source panel:** `stripe.customers`  
**Sync mode:** `FULL_TABLE`  
**Destination table (pre-create):**
```sql
CREATE TABLE analytics.stripe_customers_hd (
    id       TEXT PRIMARY KEY,
    email    TEXT,
    name     TEXT,
    currency TEXT,
    created  BIGINT
);
```

**dbt SQL:**
```sql
SELECT
    id,
    email,
    name,
    currency,
    created
FROM {{ source('raw', 'stripe__customers') }}
```

**Expected run result:**
- `rows_written > 0`
- `delivered_tables = ["analytics.stripe_customers_hd"]`

**Verify:**
```sql
SELECT COUNT(*), MAX(created) FROM analytics.stripe_customers_hd;
```

---

## Scenario S-STR-2 — Incremental Sync (`charges`)

**Source panel:** `stripe.charges`, sync mode `INCREMENTAL`, replication key `created`

**dbt SQL:**
```sql
SELECT
    id,
    amount,
    currency,
    status,
    customer,
    created
FROM {{ source('raw', 'stripe__charges') }}
```

**Run 1:** All charges synced → `rows_written = N`.  
**Run 2 (no new charges):** `rows_written = 0`.  
**Run 3 (new charge in Stripe):** `rows_written = 1`.

---

## Scenario S-STR-3 — JSON Key Filtering (`events.data`)

Stripe `events.data` is a JSON blob with many nested keys. Extract only what matters:

**dbt SQL:**
```sql
SELECT
    id,
    type                    AS event_type,
    data->>'object'         AS object_type,
    livemode,
    created
FROM {{ source('raw', 'stripe__events') }}
WHERE type LIKE 'charge.%'
```

**Destination DDL:**
```sql
CREATE TABLE analytics.stripe_charge_events (
    id          TEXT PRIMARY KEY,
    event_type  TEXT,
    object_type TEXT,
    livemode    BOOLEAN,
    created     BIGINT
);
```

**Verify:** Only `charge.*` events; `data` JSON blob not stored in destination.

---

## Scenario S-STR-4 — Revenue Aggregate by Currency

**dbt SQL:**
```sql
SELECT
    currency,
    COUNT(*)                AS charge_count,
    SUM(amount)             AS total_cents,
    AVG(amount)             AS avg_cents,
    MIN(created)            AS first_charge_ts,
    MAX(created)            AS last_charge_ts
FROM {{ source('raw', 'stripe__charges') }}
WHERE status = 'succeeded'
GROUP BY currency
```

**Destination DDL (no PK — append mode):**
```sql
CREATE TABLE analytics.stripe_revenue_by_currency (
    currency       TEXT,
    charge_count   BIGINT,
    total_cents    BIGINT,
    avg_cents      NUMERIC,
    first_charge_ts BIGINT,
    last_charge_ts  BIGINT
);
```

**Expected:** `no_pk_warnings` populated. Amber banner in drawer. Rows show correct revenue totals.

---

## Scenario S-STR-5 — Multi-Stream Pipeline (3 streams in one pipeline)

**Source panel streams:** `stripe.customers`, `stripe.charges`, `stripe.invoices`  
**Three separate dbt SQL models, one per stream.**

**Model 1 — customers:**
```sql
SELECT id, email, currency FROM {{ source('raw', 'stripe__customers') }}
```

**Model 2 — charges:**
```sql
SELECT id, amount, currency, status, customer FROM {{ source('raw', 'stripe__charges') }}
```

**Model 3 — invoices:**
```sql
SELECT id, customer, amount_due, status, created FROM {{ source('raw', 'stripe__invoices') }}
```

**Expected:** `delivered_tables = ["analytics.customers_hd", "analytics.charges_hd", "analytics.invoices_hd"]`, `dbt_models_run = 3`.

---

## Scenario S-STR-6 — Normalisation: Rename `created` → `created_ts`

**Normalisation rule:**
```json
{ "rule_type": "rename", "table": "stripe.customers", "column": "created", "destination_name": "created_ts" }
```

**dbt SQL:**
```sql
SELECT id, email, name, currency, created_ts
FROM {{ source('raw', 'stripe__customers') }}
```

**Verify:** Destination column is `created_ts`; not `created`.

---

## Scenario S-STR-7 — Normalisation: Exclude `sources` and `subscriptions`

Stripe customers carry nested `sources` and `subscriptions` objects that are useless in a flat table.

**Rules:**
```json
[
  { "rule_type": "exclude", "table": "stripe.customers", "column": "sources" },
  { "rule_type": "exclude", "table": "stripe.customers", "column": "subscriptions" }
]
```

**Verify:** Neither `sources` nor `subscriptions` columns appear in destination.

---

## Scenario S-STR-8 — Type Cast: `amount` (integer cents) → `decimal` dollars

**Rule:**
```json
{ "rule_type": "cast", "table": "stripe.charges", "column": "amount", "cast_to": "decimal" }
```

**dbt SQL:**
```sql
SELECT
    id,
    amount / 100.0  AS amount_dollars,
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

**Verify:** `amount_dollars` is decimal, e.g., `29.99` not `2999`.

---

## Scenario S-STR-9 — Subscription Status Pivot

**dbt SQL:**
```sql
SELECT
    customer,
    SUM(CASE WHEN status = 'active'   THEN 1 ELSE 0 END) AS active_count,
    SUM(CASE WHEN status = 'canceled' THEN 1 ELSE 0 END) AS canceled_count,
    SUM(CASE WHEN status = 'trialing' THEN 1 ELSE 0 END) AS trial_count
FROM {{ source('raw', 'stripe__subscriptions') }}
GROUP BY customer
```

**Destination DDL:**
```sql
CREATE TABLE analytics.sub_status_per_customer (
    customer       TEXT,
    active_count   INTEGER,
    canceled_count INTEGER,
    trial_count    INTEGER
);
```

---

## Scenario S-STR-10 — Missing Destination Table (Error Path)

1. Do NOT create the destination table.
2. Set destination table to `analytics.stripe_customers_missing`.
3. Click **Run**.

**Expected:**
- Run status: `failed`
- Error message: `"table analytics.stripe_customers_missing does not exist"` (or similar)
- `rows_written = 0`

---

## Scenario S-STR-11 — Concurrent Run Rejection

1. Trigger a run on a Stripe pipeline (long-running with many events).
2. Immediately trigger a second run on the same pipeline.

**Expected:** Second trigger returns HTTP 409 or the UI shows *"A run is already in progress"*.

---

## Scenario S-STR-12 — Cron Schedule

**General tab:** Set schedule type = `custom_cron`, value = `0 */6 * * *` (every 6 hours in UTC).

**Verify:**
- Pipeline saved with `scheduleType = "custom_cron"`, `scheduleValue = "0 */6 * * *"`.
- Next scheduled run appears in the runs list after 6 hours.
- Manual **Run** button still works immediately.

---

## All 19 Streams — Quick Smoke Test Checklist

For each stream, run a `SELECT *` pipeline and confirm `rows_written > 0`:

| Stream | dbt SQL | Expected rows |
|--------|---------|---------------|
| customers | `SELECT * FROM {{ source('raw', 'stripe__customers') }}` | ≥ 1 |
| charges | `SELECT * FROM {{ source('raw', 'stripe__charges') }}` | ≥ 1 |
| invoices | `SELECT * FROM {{ source('raw', 'stripe__invoices') }}` | ≥ 1 |
| subscriptions | `SELECT * FROM {{ source('raw', 'stripe__subscriptions') }}` | ≥ 0 |
| products | `SELECT * FROM {{ source('raw', 'stripe__products') }}` | ≥ 1 |
| prices | `SELECT * FROM {{ source('raw', 'stripe__prices') }}` | ≥ 1 |
| events | `SELECT * FROM {{ source('raw', 'stripe__events') }}` | ≥ 1 |
| balance_transactions | `SELECT * FROM {{ source('raw', 'stripe__balance_transactions') }}` | ≥ 1 |
| coupons | `SELECT * FROM {{ source('raw', 'stripe__coupons') }}` | ≥ 0 |
| payment_intents | `SELECT * FROM {{ source('raw', 'stripe__payment_intents') }}` | ≥ 1 |
| payment_methods | `SELECT * FROM {{ source('raw', 'stripe__payment_methods') }}` | ≥ 1 |
| refunds | `SELECT * FROM {{ source('raw', 'stripe__refunds') }}` | ≥ 0 |
| disputes | `SELECT * FROM {{ source('raw', 'stripe__disputes') }}` | ≥ 0 |
| payouts | `SELECT * FROM {{ source('raw', 'stripe__payouts') }}` | ≥ 1 |
| plans | `SELECT * FROM {{ source('raw', 'stripe__plans') }}` | ≥ 0 |
| tax_rates | `SELECT * FROM {{ source('raw', 'stripe__tax_rates') }}` | ≥ 0 |
| credit_notes | `SELECT * FROM {{ source('raw', 'stripe__credit_notes') }}` | ≥ 0 |
| promotion_codes | `SELECT * FROM {{ source('raw', 'stripe__promotion_codes') }}` | ≥ 0 |
| setup_intents | `SELECT * FROM {{ source('raw', 'stripe__setup_intents') }}` | ≥ 0 |
