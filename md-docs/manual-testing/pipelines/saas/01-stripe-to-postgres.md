# Pipeline 01 — Stripe → PostgreSQL

**Source streams:** 19 | **Destination:** PostgreSQL | **dbt engine:** DuckDB

---

## Connections

### Source — Stripe
```json
{ "client_secret": "sk_live_..." }
```

### Destination — PostgreSQL
```json
{ "host": "...", "port": 5432, "database": "analytics", "username": "writer", "password": "...", "ssl_mode": "disable" }
```
```sql
CREATE SCHEMA IF NOT EXISTS analytics;
```

---

## Type Reference (Stripe API → PostgreSQL)

| Stripe type | PostgreSQL type |
|------------|----------------|
| Unix epoch integer | `TIMESTAMPTZ` via `TO_TIMESTAMP(col)` |
| Amount integer (cents) | `NUMERIC(10,2)` via `/ 100.0` |
| Boolean | `BOOLEAN` |
| String | `TEXT` |
| JSONB object | `TEXT` (extract keys) or `JSONB` |
| Nullable | `NULL` handled with `COALESCE` / `NULLIF` |

---

## Stream 1 — `stripe.customers` → `analytics.stripe_customers`

### Step 1 — Destination DDL
```sql
CREATE TABLE analytics.stripe_customers (
    customer_id      TEXT        PRIMARY KEY,   -- source: id
    email_address    TEXT,                       -- source: email (renamed)
    display_name     TEXT,                       -- source: name (renamed)
    billing_currency TEXT,                       -- source: currency (renamed)
    is_delinquent    BOOLEAN,                    -- source: delinquent
    first_seen_at    TIMESTAMPTZ,                -- source: created INT → TIMESTAMPTZ
    metadata_keys    TEXT                        -- from JSON: metadata (slim TEXT, not full blob)
);
```

### Step 3 — Source Panel
`stripe.customers` | `FULL_TABLE`

### Step 5 — dbt SQL
```sql
SELECT
    id                                AS customer_id,
    email                             AS email_address,
    name                              AS display_name,
    currency                          AS billing_currency,
    delinquent                        AS is_delinquent,
    TO_TIMESTAMP(created)::TIMESTAMPTZ AS first_seen_at,
    NULLIF(metadata::TEXT, '{}')      AS metadata_keys
FROM {{ source('raw', 'stripe__customers') }}
WHERE email IS NOT NULL
```

### Step 8 — Verify
```sql
SELECT customer_id, email_address, billing_currency, first_seen_at FROM analytics.stripe_customers LIMIT 5;
-- first_seen_at must be TIMESTAMPTZ, not an integer
-- column 'id', 'name', 'currency', 'created' must NOT appear
SELECT COUNT(*) FROM analytics.stripe_customers;
```

---

## Stream 2 — `stripe.charges` → `analytics.stripe_charges`

### Step 1 — Destination DDL
```sql
CREATE TABLE analytics.stripe_charges (
    charge_id        TEXT        PRIMARY KEY,   -- source: id
    customer_ref     TEXT,                       -- source: customer (renamed)
    amount_dollars   NUMERIC(10,2),              -- source: amount (cents)/100
    currency         TEXT,
    charge_status    TEXT,                       -- source: status (renamed)
    card_brand       TEXT,                       -- from JSON: payment_method_details->'card'->>'brand'
    card_last4       TEXT,                       -- from JSON: payment_method_details->'card'->>'last4'
    refunded         BOOLEAN,
    charged_at       TIMESTAMPTZ                 -- source: created → TIMESTAMPTZ
);
```

### Step 5 — dbt SQL
```sql
SELECT
    id                                              AS charge_id,
    customer                                        AS customer_ref,
    amount / 100.0                                  AS amount_dollars,
    currency,
    status                                          AS charge_status,
    payment_method_details->'card'->>'brand'        AS card_brand,
    payment_method_details->'card'->>'last4'        AS card_last4,
    refunded,
    TO_TIMESTAMP(created)::TIMESTAMPTZ              AS charged_at
FROM {{ source('raw', 'stripe__charges') }}
WHERE status = 'succeeded'
```

### Step 8 — Verify
```sql
SELECT amount_dollars, card_brand, card_last4 FROM analytics.stripe_charges LIMIT 5;
-- amount_dollars: decimal like 29.99, NOT 2999
-- card_brand: plain 'visa', NOT a JSON object
-- payment_method_details column must NOT exist
```

---

## Stream 3 — `stripe.invoices` → `analytics.stripe_invoices`

### Step 1 — Destination DDL
```sql
CREATE TABLE analytics.stripe_invoices (
    invoice_id       TEXT        PRIMARY KEY,
    customer_ref     TEXT,                       -- source: customer
    subscription_ref TEXT,                       -- source: subscription
    amount_due       NUMERIC(10,2),              -- source: amount_due (cents)/100
    amount_paid      NUMERIC(10,2),              -- source: amount_paid (cents)/100
    invoice_status   TEXT,                       -- source: status (renamed)
    period_from      DATE,                       -- source: period_start INT → DATE
    period_to        DATE,                       -- source: period_end INT → DATE
    issued_at        TIMESTAMPTZ
);
```

### Step 5 — dbt SQL
```sql
SELECT
    id                                        AS invoice_id,
    customer                                  AS customer_ref,
    subscription                              AS subscription_ref,
    amount_due  / 100.0                       AS amount_due,
    amount_paid / 100.0                       AS amount_paid,
    status                                    AS invoice_status,
    TO_TIMESTAMP(period_start)::DATE          AS period_from,
    TO_TIMESTAMP(period_end)::DATE            AS period_to,
    TO_TIMESTAMP(created)::TIMESTAMPTZ        AS issued_at
FROM {{ source('raw', 'stripe__invoices') }}
WHERE status IN ('paid','open','void','uncollectible')
```

### Step 8 — Verify
```sql
SELECT period_from, period_to, amount_due, invoice_status FROM analytics.stripe_invoices LIMIT 5;
-- period_from/period_to must be DATE only (no time)
```

---

## Stream 4 — `stripe.subscriptions` → `analytics.stripe_subscriptions`

### Step 1 — Destination DDL
```sql
CREATE TABLE analytics.stripe_subscriptions (
    sub_id           TEXT        PRIMARY KEY,
    customer_ref     TEXT,
    plan_id          TEXT,                       -- from JSON: plan->>'id'
    plan_interval    TEXT,                       -- from JSON: plan->>'interval'
    plan_amount      NUMERIC(10,2),              -- from JSON: plan->>'amount' (cents)/100
    sub_status       TEXT,                       -- source: status (renamed)
    trial_ends_on    DATE,                       -- source: trial_end INT → DATE
    renews_on        DATE,                       -- source: current_period_end INT → DATE
    cancelled_at     TIMESTAMPTZ
);
```

### Step 5 — dbt SQL
```sql
SELECT
    id                                                             AS sub_id,
    customer                                                       AS customer_ref,
    plan->>'id'                                                    AS plan_id,
    plan->>'interval'                                              AS plan_interval,
    CAST(NULLIF(plan->>'amount','') AS NUMERIC) / 100.0            AS plan_amount,
    status                                                         AS sub_status,
    CASE WHEN trial_end IS NOT NULL
         THEN TO_TIMESTAMP(trial_end)::DATE END                    AS trial_ends_on,
    TO_TIMESTAMP(current_period_end)::DATE                         AS renews_on,
    CASE WHEN canceled_at IS NOT NULL
         THEN TO_TIMESTAMP(canceled_at)::TIMESTAMPTZ END           AS cancelled_at
FROM {{ source('raw', 'stripe__subscriptions') }}
```

### Step 8 — Verify
```sql
SELECT plan_id, plan_interval, plan_amount, sub_status FROM analytics.stripe_subscriptions LIMIT 5;
-- plan_id: plain string like 'price_xxx', NOT full JSON
-- plan_amount: decimal dollars
```

---

## Stream 5 — `stripe.products` → `analytics.stripe_products`

### Step 1 — Destination DDL
```sql
CREATE TABLE analytics.stripe_products (
    product_id    TEXT      PRIMARY KEY,
    product_name  TEXT,                  -- source: name (renamed)
    description   TEXT,
    is_active     BOOLEAN,               -- source: active (renamed)
    unit_label    TEXT,
    created_at    TIMESTAMPTZ,
    updated_at    TIMESTAMPTZ
);
```

### Step 5 — dbt SQL
```sql
SELECT
    id                                   AS product_id,
    name                                 AS product_name,
    description,
    active                               AS is_active,
    unit_label,
    TO_TIMESTAMP(created)::TIMESTAMPTZ   AS created_at,
    TO_TIMESTAMP(updated)::TIMESTAMPTZ   AS updated_at
FROM {{ source('raw', 'stripe__products') }}
WHERE active = true
```

---

## Stream 6 — `stripe.prices` → `analytics.stripe_prices`

### Step 1 — Destination DDL
```sql
CREATE TABLE analytics.stripe_prices (
    price_id       TEXT        PRIMARY KEY,
    product_ref    TEXT,                       -- source: product (renamed)
    unit_amount    NUMERIC(10,2),              -- source: unit_amount (cents)/100
    currency       TEXT,
    billing_scheme TEXT,
    price_type     TEXT,                       -- source: type (renamed)
    interval_unit  TEXT,                       -- from JSON: recurring->>'interval'
    interval_count INTEGER,                    -- from JSON: recurring->>'interval_count' → INT
    is_active      BOOLEAN,
    created_at     TIMESTAMPTZ
);
```

### Step 5 — dbt SQL
```sql
SELECT
    id                                           AS price_id,
    product                                      AS product_ref,
    unit_amount / 100.0                          AS unit_amount,
    currency,
    billing_scheme,
    type                                         AS price_type,
    recurring->>'interval'                       AS interval_unit,
    CAST(recurring->>'interval_count' AS INTEGER) AS interval_count,
    active                                       AS is_active,
    TO_TIMESTAMP(created)::TIMESTAMPTZ           AS created_at
FROM {{ source('raw', 'stripe__prices') }}
```

### Step 8 — Verify
```sql
SELECT price_id, unit_amount, interval_unit, interval_count FROM analytics.stripe_prices LIMIT 5;
-- interval_unit: 'month' / 'year' (plain string, not JSON)
-- interval_count: INTEGER
-- unit_amount: decimal dollars
```

---

## Stream 7 — `stripe.payment_intents` → `analytics.stripe_payment_intents`

### Step 1 — Destination DDL
```sql
CREATE TABLE analytics.stripe_payment_intents (
    intent_id        TEXT        PRIMARY KEY,
    customer_ref     TEXT,
    amount_dollars   NUMERIC(10,2),
    currency         TEXT,
    intent_status    TEXT,                       -- source: status
    payment_method   TEXT,                       -- source: payment_method
    is_captured      BOOLEAN,                    -- derived: status = 'succeeded'
    created_at       TIMESTAMPTZ
);
```

### Step 5 — dbt SQL
```sql
SELECT
    id                                   AS intent_id,
    customer                             AS customer_ref,
    amount / 100.0                       AS amount_dollars,
    currency,
    status                               AS intent_status,
    payment_method,
    status = 'succeeded'                 AS is_captured,
    TO_TIMESTAMP(created)::TIMESTAMPTZ   AS created_at
FROM {{ source('raw', 'stripe__payment_intents') }}
```

---

## Stream 8 — `stripe.payment_methods` → `analytics.stripe_payment_methods`

### Step 1 — Destination DDL
```sql
CREATE TABLE analytics.stripe_payment_methods (
    pm_id         TEXT        PRIMARY KEY,
    customer_ref  TEXT,
    pm_type       TEXT,                       -- source: type (renamed)
    card_brand    TEXT,                       -- from JSON: card->>'brand'
    card_last4    TEXT,                       -- from JSON: card->>'last4'
    card_expiry   TEXT,                       -- derived: month/year concat
    added_at      TIMESTAMPTZ
);
```

### Step 5 — dbt SQL
```sql
SELECT
    id                                              AS pm_id,
    customer                                        AS customer_ref,
    type                                            AS pm_type,
    card->>'brand'                                  AS card_brand,
    card->>'last4'                                  AS card_last4,
    (card->>'exp_month')||'/'||(card->>'exp_year')  AS card_expiry,
    TO_TIMESTAMP(created)::TIMESTAMPTZ              AS added_at
FROM {{ source('raw', 'stripe__payment_methods') }}
WHERE type = 'card'
  AND card IS NOT NULL
```

---

## Stream 9 — `stripe.refunds` → `analytics.stripe_refunds`

### Step 1 — Destination DDL
```sql
CREATE TABLE analytics.stripe_refunds (
    refund_id      TEXT        PRIMARY KEY,
    charge_ref     TEXT,
    refund_amount  NUMERIC(10,2),
    currency       TEXT,
    refund_reason  TEXT,
    refund_status  TEXT,
    refunded_at    TIMESTAMPTZ
);
```

### Step 5 — dbt SQL
```sql
SELECT
    id                                   AS refund_id,
    charge                               AS charge_ref,
    amount / 100.0                       AS refund_amount,
    currency,
    reason                               AS refund_reason,
    status                               AS refund_status,
    TO_TIMESTAMP(created)::TIMESTAMPTZ   AS refunded_at
FROM {{ source('raw', 'stripe__refunds') }}
```

---

## Stream 10 — `stripe.disputes` → `analytics.stripe_disputes`

### Step 1 — Destination DDL
```sql
CREATE TABLE analytics.stripe_disputes (
    dispute_id      TEXT        PRIMARY KEY,
    charge_ref      TEXT,
    dispute_amount  NUMERIC(10,2),
    currency        TEXT,
    dispute_reason  TEXT,                       -- source: reason (renamed)
    dispute_status  TEXT,                       -- source: status (renamed)
    is_won          BOOLEAN,                    -- derived: status = 'won'
    created_at      TIMESTAMPTZ
);
```

### Step 5 — dbt SQL
```sql
SELECT
    id                                   AS dispute_id,
    charge                               AS charge_ref,
    amount / 100.0                       AS dispute_amount,
    currency,
    reason                               AS dispute_reason,
    status                               AS dispute_status,
    status = 'won'                       AS is_won,
    TO_TIMESTAMP(created)::TIMESTAMPTZ   AS created_at
FROM {{ source('raw', 'stripe__disputes') }}
```

---

## Stream 11 — `stripe.events` → `analytics.stripe_events`

### Step 1 — Destination DDL
```sql
CREATE TABLE analytics.stripe_events (
    event_id      TEXT        PRIMARY KEY,
    event_type    TEXT,                       -- source: type (renamed)
    object_type   TEXT,                       -- from JSON: data->>'object'
    object_id     TEXT,                       -- from JSON: data->>'id'
    api_version   TEXT,
    is_live       BOOLEAN,                    -- source: livemode (renamed)
    occurred_at   TIMESTAMPTZ
    -- data blob NOT stored
);
```

### Step 5 — dbt SQL
```sql
SELECT
    id                                   AS event_id,
    type                                 AS event_type,
    data->>'object'                      AS object_type,
    data->>'id'                          AS object_id,
    api_version,
    livemode                             AS is_live,
    TO_TIMESTAMP(created)::TIMESTAMPTZ   AS occurred_at
FROM {{ source('raw', 'stripe__events') }}
WHERE livemode = true
```

### Step 8 — Verify
```sql
SELECT event_type, object_type, object_id FROM analytics.stripe_events LIMIT 5;
-- object_type = plain string 'charge', NOT a JSON blob
-- 'data' column must NOT exist in destination
```

---

## Stream 12 — `stripe.customers_balance_transactions` → `analytics.stripe_balance_txns`

### Step 1 — Destination DDL
```sql
CREATE TABLE analytics.stripe_balance_txns (
    txn_id        TEXT        PRIMARY KEY,
    customer_ref  TEXT,
    amount        NUMERIC(10,2),
    currency      TEXT,
    txn_type      TEXT,                       -- source: type (renamed)
    description   TEXT,
    txn_at        TIMESTAMPTZ
);
```

### Step 5 — dbt SQL
```sql
SELECT
    id                                   AS txn_id,
    customer                             AS customer_ref,
    amount / 100.0                       AS amount,
    currency,
    type                                 AS txn_type,
    description,
    TO_TIMESTAMP(created)::TIMESTAMPTZ   AS txn_at
FROM {{ source('raw', 'stripe__customers_balance_transactions') }}
```

---

## Stream 13 — `stripe.coupons` → `analytics.stripe_coupons`

### Step 1 — Destination DDL
```sql
CREATE TABLE analytics.stripe_coupons (
    coupon_id     TEXT        PRIMARY KEY,
    coupon_name   TEXT,                       -- source: name (renamed)
    discount_type TEXT,                       -- source: duration (renamed)
    pct_off       NUMERIC(5,2),               -- source: percent_off (renamed)
    amount_off    NUMERIC(10,2),              -- source: amount_off (cents)/100
    currency      TEXT,
    is_valid      BOOLEAN,                    -- source: valid (renamed)
    max_uses      INTEGER,                    -- source: max_redemptions
    use_count     INTEGER,                    -- source: times_redeemed
    created_at    TIMESTAMPTZ
);
```

### Step 5 — dbt SQL
```sql
SELECT
    id                                   AS coupon_id,
    name                                 AS coupon_name,
    duration                             AS discount_type,
    percent_off                          AS pct_off,
    COALESCE(amount_off,0) / 100.0       AS amount_off,
    currency,
    valid                                AS is_valid,
    max_redemptions                      AS max_uses,
    times_redeemed                       AS use_count,
    TO_TIMESTAMP(created)::TIMESTAMPTZ   AS created_at
FROM {{ source('raw', 'stripe__coupons') }}
```

---

## Stream 14 — `stripe.promotion_codes` → `analytics.stripe_promo_codes`

### Step 1 — Destination DDL
```sql
CREATE TABLE analytics.stripe_promo_codes (
    promo_id      TEXT        PRIMARY KEY,
    coupon_ref    TEXT,                       -- source: coupon->>'id' (JSON)
    promo_code    TEXT,                       -- source: code (renamed)
    is_active     BOOLEAN,
    max_uses      INTEGER,                    -- source: max_redemptions
    use_count     INTEGER,                    -- source: times_redeemed
    expires_on    DATE,
    created_at    TIMESTAMPTZ
);
```

### Step 5 — dbt SQL
```sql
SELECT
    id                                           AS promo_id,
    coupon->>'id'                                AS coupon_ref,
    code                                         AS promo_code,
    active                                       AS is_active,
    max_redemptions                              AS max_uses,
    times_redeemed                               AS use_count,
    CASE WHEN expires_at IS NOT NULL
         THEN TO_TIMESTAMP(expires_at)::DATE END AS expires_on,
    TO_TIMESTAMP(created)::TIMESTAMPTZ           AS created_at
FROM {{ source('raw', 'stripe__promotion_codes') }}
```

---

## Stream 15 — `stripe.tax_rates` → `analytics.stripe_tax_rates`

### Step 1 — Destination DDL
```sql
CREATE TABLE analytics.stripe_tax_rates (
    rate_id       TEXT        PRIMARY KEY,
    display_name  TEXT,
    jurisdiction  TEXT,
    rate_pct      NUMERIC(6,4),               -- source: percentage float→NUMERIC
    is_inclusive  BOOLEAN,
    is_active     BOOLEAN,
    created_at    TIMESTAMPTZ
);
```

### Step 5 — dbt SQL
```sql
SELECT
    id                                   AS rate_id,
    display_name,
    jurisdiction,
    CAST(percentage AS NUMERIC)          AS rate_pct,
    inclusive                            AS is_inclusive,
    active                               AS is_active,
    TO_TIMESTAMP(created)::TIMESTAMPTZ   AS created_at
FROM {{ source('raw', 'stripe__tax_rates') }}
```

---

## Stream 16 — `stripe.setup_intents` → `analytics.stripe_setup_intents`

### Step 1 — Destination DDL
```sql
CREATE TABLE analytics.stripe_setup_intents (
    intent_id      TEXT        PRIMARY KEY,
    customer_ref   TEXT,
    payment_method TEXT,
    intent_status  TEXT,
    usage          TEXT,
    created_at     TIMESTAMPTZ
);
```

### Step 5 — dbt SQL
```sql
SELECT
    id                                   AS intent_id,
    customer                             AS customer_ref,
    payment_method,
    status                               AS intent_status,
    usage,
    TO_TIMESTAMP(created)::TIMESTAMPTZ   AS created_at
FROM {{ source('raw', 'stripe__setup_intents') }}
```

---

## Stream 17 — `stripe.credit_notes` → `analytics.stripe_credit_notes`

### Step 1 — Destination DDL
```sql
CREATE TABLE analytics.stripe_credit_notes (
    note_id        TEXT        PRIMARY KEY,
    customer_ref   TEXT,
    invoice_ref    TEXT,
    note_amount    NUMERIC(10,2),
    currency       TEXT,
    note_status    TEXT,
    note_type      TEXT,
    created_at     TIMESTAMPTZ
);
```

### Step 5 — dbt SQL
```sql
SELECT
    id                                   AS note_id,
    customer                             AS customer_ref,
    invoice                              AS invoice_ref,
    amount / 100.0                       AS note_amount,
    currency,
    status                               AS note_status,
    type                                 AS note_type,
    TO_TIMESTAMP(created)::TIMESTAMPTZ   AS created_at
FROM {{ source('raw', 'stripe__credit_notes') }}
```

---

## Stream 18 — `stripe.early_fraud_warnings` → `analytics.stripe_fraud_warnings`

### Step 1 — Destination DDL
```sql
CREATE TABLE analytics.stripe_fraud_warnings (
    warning_id      TEXT        PRIMARY KEY,
    charge_ref      TEXT,
    fraud_type      TEXT,                       -- source: fraud_type
    actionable      BOOLEAN,
    created_at      TIMESTAMPTZ
);
```

### Step 5 — dbt SQL
```sql
SELECT
    id                                   AS warning_id,
    charge                               AS charge_ref,
    fraud_type,
    actionable,
    TO_TIMESTAMP(created)::TIMESTAMPTZ   AS created_at
FROM {{ source('raw', 'stripe__early_fraud_warnings') }}
```

---

## Stream 19 — `stripe.customers_bank_accounts` → `analytics.stripe_bank_accounts`

### Step 1 — Destination DDL
```sql
CREATE TABLE analytics.stripe_bank_accounts (
    account_id    TEXT        PRIMARY KEY,
    customer_ref  TEXT,
    bank_name     TEXT,
    routing_num   TEXT,                       -- source: routing_number (renamed)
    last4         TEXT,
    account_type  TEXT,                       -- source: account_holder_type (renamed)
    acct_status   TEXT,                       -- source: status (renamed)
    currency      TEXT
);
```

### Step 5 — dbt SQL
```sql
SELECT
    id                              AS account_id,
    customer                        AS customer_ref,
    bank_name,
    routing_number                  AS routing_num,
    last4,
    account_holder_type             AS account_type,
    status                          AS acct_status,
    currency
FROM {{ source('raw', 'stripe__customers_bank_accounts') }}
```

---

## Cross-Stream Edge Cases

### EC-1 — JSON key absent → NULL destination column
```sql
-- In charges stream: payment_method_details may be NULL for some charge types
SELECT
    id AS charge_id,
    COALESCE(payment_method_details->'card'->>'brand', 'unknown') AS card_brand,
    payment_method_details->'card'->>'last4'                       AS card_last4
FROM {{ source('raw', 'stripe__charges') }}
```
**Verify:** `SELECT COUNT(*) FROM analytics.stripe_charges WHERE card_brand = 'unknown';` — may be > 0 (bank debits have no card data).

### EC-2 — Empty `amount_off` in coupons (percent-off coupons)
```sql
-- COALESCE(amount_off, 0) prevents NULL arithmetic crash
COALESCE(amount_off, 0) / 100.0 AS amount_off
```

### EC-3 — Concurrent run rejection
Trigger two runs simultaneously. **Expected:** second run returns `409 Conflict`.

### EC-4 — Wrong API key
Set `client_secret = "sk_live_invalid"`. **Expected:** Phase 1 fails with Stripe 401.

### EC-5 — Incremental sync: `stripe.charges`
After initial FULL_TABLE run, switch to INCREMENTAL with replication key `created`.  
Add a new charge in Stripe. Run again. **Expected:** only the new charge row in staging (`rows_written = 1`).

### EC-6 — Destination column type mismatch
Create `amount_dollars INTEGER` instead of `NUMERIC`. dbt SQL outputs `NUMERIC`.  
**Expected:** Phase 0 fails with column type error.

### EC-7 — Destination table missing
Do not pre-create the table. **Expected:** Phase 0 fails: table does not exist.
