# Stripe Source — UI Testing (All 19 Streams × 5 Destinations)

> Follow `builder-walkthrough.md` for universal steps.  
> This file documents Stripe-specific source config, stream-by-stream settings, and all 5 destination variants.

---

## Phase 1 — Source Panel (Stripe)

### Step 2a — Add Source
1. Click **"Add Source"** → select **"Stripe"**
2. Fill in credential fields:

| Field | Value | Notes |
|-------|-------|-------|
| **API Key** | `sk_live_...` or `sk_test_...` | Test key for non-prod environments |

3. Click **"Test Connection"**
   - ✅ Green: Stripe API reachable and key valid
   - ❌ `401`: wrong key
   - ❌ `403`: restricted key — missing `read` permission on required objects

### Step 2b — Stream Selection

Enable all 19 streams:

| Stream name | Display label | Incremental? | Cursor field |
|-------------|--------------|-------------|-------------|
| `customers` | Customers | ✅ INCREMENTAL | `created` |
| `charges` | Charges | ✅ INCREMENTAL | `created` |
| `invoices` | Invoices | ✅ INCREMENTAL | `created` |
| `subscriptions` | Subscriptions | ✅ INCREMENTAL | `created` |
| `products` | Products | ✅ INCREMENTAL | `created` |
| `prices` | Prices | ✅ INCREMENTAL | `created` |
| `payment_intents` | Payment Intents | ✅ INCREMENTAL | `created` |
| `payment_methods` | Payment Methods | FULL TABLE | — |
| `refunds` | Refunds | ✅ INCREMENTAL | `created` |
| `disputes` | Disputes | ✅ INCREMENTAL | `created` |
| `events` | Events | ✅ INCREMENTAL | `created` |
| `customers_balance_transactions` | Balance Transactions | ✅ INCREMENTAL | `created` |
| `coupons` | Coupons | ✅ INCREMENTAL | `created` |
| `promotion_codes` | Promotion Codes | ✅ INCREMENTAL | `created` |
| `tax_rates` | Tax Rates | FULL TABLE | — |
| `setup_intents` | Setup Intents | ✅ INCREMENTAL | `created` |
| `credit_notes` | Credit Notes | ✅ INCREMENTAL | `created` |
| `early_fraud_warnings` | Early Fraud Warnings | ✅ INCREMENTAL | `created` |
| `customers_bank_accounts` | Bank Accounts | FULL TABLE | — |

---

## Phase 2 — Destination Panel (per destination)

### Variant A — PostgreSQL

| Field | Value |
|-------|-------|
| Host | `your-pg-host` |
| Port | `5432` |
| Database | `analytics` |
| Username | `writer` |
| Password | `***` |
| SSL Mode | `disable` (dev) / `require` (prod) |

**Test Connection → ✅**

Stream → table mapping:

| Stream | Schema | Table |
|--------|--------|-------|
| customers | analytics | stripe_customers |
| charges | analytics | stripe_charges |
| invoices | analytics | stripe_invoices |
| subscriptions | analytics | stripe_subscriptions |
| products | analytics | stripe_products |
| prices | analytics | stripe_prices |
| payment_intents | analytics | stripe_payment_intents |
| payment_methods | analytics | stripe_payment_methods |
| refunds | analytics | stripe_refunds |
| disputes | analytics | stripe_disputes |
| events | analytics | stripe_events |
| customers_balance_transactions | analytics | stripe_balance_txns |
| coupons | analytics | stripe_coupons |
| promotion_codes | analytics | stripe_promo_codes |
| tax_rates | analytics | stripe_tax_rates |
| setup_intents | analytics | stripe_setup_intents |
| credit_notes | analytics | stripe_credit_notes |
| early_fraud_warnings | analytics | stripe_fraud_warnings |
| customers_bank_accounts | analytics | stripe_bank_accounts |

### Variant B — MySQL

Same stream→table mapping; connection fields:

| Field | Value |
|-------|-------|
| Host | `your-mysql-host` |
| Port | `3306` |
| Database | `analytics` |
| Username | `writer` |
| Password | `***` |

### Variant C — MariaDB
Same as MySQL variant (MariaDB uses MySQL protocol).

### Variant D — SQLite

| Field | Value |
|-------|-------|
| Database path | `/absolute/path/to/analytics.db` |

Schema: `main` for all tables.

### Variant E — CockroachDB

| Field | Value |
|-------|-------|
| Host | `your-crdb-host` |
| Port | `26257` |
| Database | `defaultdb` |
| Username | `root` |
| SSL Mode | `disable` (local) / `verify-full` (cloud) |

Schema: `analytics`

---

## Phase 3 — Normalisation Panel (stream-by-stream)

### `stripe.charges` normalisation rules
| Rule | Column | Target |
|------|--------|--------|
| Rename | `id` | `charge_id` |
| Rename | `customer` | `customer_ref` |
| Rename | `payment_method_details` | *(excluded — use JSON extract in dbt)* |
| Exclude | `payment_method_details` | — |
| Exclude | `metadata` | — |
| Cast | `refunded` | Boolean |

### `stripe.invoices` normalisation rules
| Rule | Column | Target |
|------|--------|--------|
| Rename | `id` | `invoice_id` |
| Rename | `customer` | `customer_ref` |
| Rename | `subscription` | `subscription_ref` |
| Exclude | `lines` | — |
| Exclude | `custom_fields` | — |
| Exclude | `metadata` | — |

### `stripe.subscriptions` normalisation rules
| Rule | Column | Target |
|------|--------|--------|
| Rename | `id` | `sub_id` |
| Rename | `customer` | `customer_ref` |
| Rename | `status` | `sub_status` |
| Exclude | `items` | — |
| Exclude | `metadata` | — |

### `stripe.customers` normalisation rules
| Rule | Column | Target |
|------|--------|--------|
| Rename | `id` | `customer_id` |
| Rename | `email` | `email_address` |
| Rename | `name` | `display_name` |
| Rename | `currency` | `billing_currency` |
| Cast | `delinquent` | Boolean |
| Exclude | `metadata` | — |

### All other streams
Use **rename only** for `id` → `<stream>_id`. Exclude `metadata` and any nested array columns.

---

## Phase 4 — dbt SQL Panel (stream-by-stream)

### Stream 1 — `stripe.customers`
**Source token**: `{{ source('raw', 'stripe__customers') }}`

```sql
SELECT
    id                                      AS customer_id,
    email                                   AS email_address,
    name                                    AS display_name,
    currency                                AS billing_currency,
    delinquent                              AS is_delinquent,
    TO_TIMESTAMP(created)::TIMESTAMPTZ      AS first_seen_at,
    NULLIF(TRIM(metadata::TEXT), '{}')      AS metadata_keys
FROM {{ source('raw', 'stripe__customers') }}
WHERE email IS NOT NULL
```

**Validate → expected columns**: `customer_id`, `email_address`, `display_name`, `billing_currency`, `is_delinquent`, `first_seen_at`, `metadata_keys`

---

### Stream 2 — `stripe.charges`
```sql
SELECT
    id                                           AS charge_id,
    customer                                     AS customer_ref,
    CAST(amount AS NUMERIC) / 100                AS amount_dollars,
    currency,
    status                                       AS charge_status,
    payment_method_details->'card'->>'brand'     AS card_brand,
    payment_method_details->'card'->>'last4'     AS card_last4,
    refunded,
    TO_TIMESTAMP(created)::TIMESTAMPTZ           AS charged_at
FROM {{ source('raw', 'stripe__charges') }}
WHERE status IN ('succeeded','failed','pending')
```

**Key JSON extractions to verify in preview**:
- `card_brand`: plain string `"visa"` — not `{"brand":"visa"}`
- `card_last4`: 4-char string
- `amount_dollars`: decimal `29.99` — not integer `2999`

---

### Stream 3 — `stripe.invoices`
```sql
SELECT
    id                                   AS invoice_id,
    customer                             AS customer_ref,
    subscription                         AS subscription_ref,
    CAST(amount_due   AS NUMERIC) / 100  AS amount_due,
    CAST(amount_paid  AS NUMERIC) / 100  AS amount_paid,
    status                               AS invoice_status,
    TO_TIMESTAMP(period_start)::DATE     AS period_from,
    TO_TIMESTAMP(period_end)::DATE       AS period_to,
    TO_TIMESTAMP(created)::TIMESTAMPTZ   AS issued_at
FROM {{ source('raw', 'stripe__invoices') }}
```

---

### Stream 4 — `stripe.subscriptions`
```sql
SELECT
    id                                   AS sub_id,
    customer                             AS customer_ref,
    plan->>'id'                          AS plan_id,
    plan->>'interval'                    AS plan_interval,
    CAST(plan->>'amount' AS NUMERIC)/100 AS plan_amount,
    status                               AS sub_status,
    TO_TIMESTAMP(trial_end)::DATE        AS trial_ends_on,
    TO_TIMESTAMP(current_period_end)::DATE AS renews_on,
    TO_TIMESTAMP(NULLIF(canceled_at,0))::TIMESTAMPTZ AS cancelled_at
FROM {{ source('raw', 'stripe__subscriptions') }}
```

**JSON key extraction from `plan` object**:
- `plan->>'id'`: string
- `plan->>'interval'`: `"month"` / `"year"`
- `plan->>'amount'`: cents integer → divide by 100

---

### Stream 5 — `stripe.products`
```sql
SELECT
    id          AS product_id,
    name        AS product_name,
    description,
    active      AS is_active,
    unit_label,
    TO_TIMESTAMP(created)::TIMESTAMPTZ AS created_at,
    TO_TIMESTAMP(updated)::TIMESTAMPTZ AS updated_at
FROM {{ source('raw', 'stripe__products') }}
WHERE active = true
```

---

### Stream 6 — `stripe.prices`
```sql
SELECT
    id                                          AS price_id,
    product                                     AS product_ref,
    CAST(unit_amount AS NUMERIC) / 100          AS unit_amount,
    currency,
    billing_scheme,
    type                                        AS price_type,
    recurring->>'interval'                      AS interval_unit,
    CAST(recurring->>'interval_count' AS INTEGER) AS interval_count,
    active                                      AS is_active,
    TO_TIMESTAMP(created)::TIMESTAMPTZ          AS created_at
FROM {{ source('raw', 'stripe__prices') }}
```

---

### Stream 7 — `stripe.payment_intents`
```sql
SELECT
    id                                   AS intent_id,
    customer                             AS customer_ref,
    CAST(amount AS NUMERIC) / 100        AS amount_dollars,
    currency,
    status                               AS intent_status,
    payment_method,
    amount_capturable > 0                AS is_captured,
    TO_TIMESTAMP(created)::TIMESTAMPTZ   AS created_at
FROM {{ source('raw', 'stripe__payment_intents') }}
```

---

### Stream 8 — `stripe.payment_methods`
```sql
SELECT
    id                               AS pm_id,
    customer                         AS customer_ref,
    type                             AS pm_type,
    card->>'brand'                   AS card_brand,
    card->>'last4'                   AS card_last4,
    CONCAT(card->>'exp_month','/',card->>'exp_year') AS card_expiry,
    TO_TIMESTAMP(created)::TIMESTAMPTZ AS added_at
FROM {{ source('raw', 'stripe__payment_methods') }}
WHERE type = 'card'
```

---

### Stream 9 — `stripe.refunds`
```sql
SELECT
    id                                   AS refund_id,
    charge                               AS charge_ref,
    CAST(amount AS NUMERIC) / 100        AS refund_amount,
    currency,
    reason                               AS refund_reason,
    status                               AS refund_status,
    TO_TIMESTAMP(created)::TIMESTAMPTZ   AS refunded_at
FROM {{ source('raw', 'stripe__refunds') }}
```

---

### Stream 10 — `stripe.disputes`
```sql
SELECT
    id                                   AS dispute_id,
    charge                               AS charge_ref,
    CAST(amount AS NUMERIC) / 100        AS dispute_amount,
    currency,
    reason                               AS dispute_reason,
    status                               AS dispute_status,
    status = 'won'                       AS is_won,
    TO_TIMESTAMP(created)::TIMESTAMPTZ   AS created_at
FROM {{ source('raw', 'stripe__disputes') }}
```

---

### Stream 11 — `stripe.events`
```sql
SELECT
    id                                   AS event_id,
    type                                 AS event_type,
    data->'object'->>'object'            AS object_type,
    data->'object'->>'id'                AS object_id,
    api_version,
    livemode                             AS is_live,
    TO_TIMESTAMP(created)::TIMESTAMPTZ   AS occurred_at
FROM {{ source('raw', 'stripe__events') }}
WHERE type IS NOT NULL
```

---

### Stream 12 — `stripe.customers_balance_transactions`
```sql
SELECT
    id                                   AS txn_id,
    customer                             AS customer_ref,
    CAST(amount AS NUMERIC) / 100        AS amount,
    currency,
    type                                 AS txn_type,
    description,
    TO_TIMESTAMP(created)::TIMESTAMPTZ   AS txn_at
FROM {{ source('raw', 'stripe__customers_balance_transactions') }}
```

---

### Stream 13 — `stripe.coupons`
```sql
SELECT
    id                                          AS coupon_id,
    name                                        AS coupon_name,
    duration                                    AS discount_type,
    percent_off                                 AS pct_off,
    CAST(NULLIF(amount_off::TEXT,'') AS NUMERIC)/100 AS amount_off,
    currency,
    valid                                       AS is_valid,
    max_redemptions                             AS max_uses,
    times_redeemed                              AS use_count,
    TO_TIMESTAMP(created)::TIMESTAMPTZ          AS created_at
FROM {{ source('raw', 'stripe__coupons') }}
```

---

### Stream 14 — `stripe.promotion_codes`
```sql
SELECT
    id                                   AS promo_id,
    coupon->>'id'                        AS coupon_ref,
    code                                 AS promo_code,
    active                               AS is_active,
    max_redemptions                      AS max_uses,
    times_redeemed                       AS use_count,
    TO_TIMESTAMP(NULLIF(expires_at,0))::DATE AS expires_on,
    TO_TIMESTAMP(created)::TIMESTAMPTZ   AS created_at
FROM {{ source('raw', 'stripe__promotion_codes') }}
```

---

### Stream 15 — `stripe.tax_rates`
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

### Stream 16 — `stripe.setup_intents`
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

### Stream 17 — `stripe.credit_notes`
```sql
SELECT
    id                                   AS note_id,
    customer                             AS customer_ref,
    invoice                              AS invoice_ref,
    CAST(amount AS NUMERIC) / 100        AS note_amount,
    currency,
    status                               AS note_status,
    type                                 AS note_type,
    TO_TIMESTAMP(created)::TIMESTAMPTZ   AS created_at
FROM {{ source('raw', 'stripe__credit_notes') }}
```

---

### Stream 18 — `stripe.early_fraud_warnings`
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

### Stream 19 — `stripe.customers_bank_accounts`
```sql
SELECT
    id                AS account_id,
    customer          AS customer_ref,
    bank_name,
    routing_number    AS routing_num,
    last4,
    account_type,
    status            AS acct_status,
    currency
FROM {{ source('raw', 'stripe__customers_bank_accounts') }}
```

---

## Phase 5 — Preview Checks (per stream)

| Stream | Column to spot-check | Expected value format |
|--------|--------------------|--------------------|
| charges | `amount_dollars` | `29.99` (decimal, not `2999`) |
| charges | `card_brand` | `"visa"` (plain string) |
| subscriptions | `plan_id` | `"price_xxx"` (plain string, not JSON) |
| events | `object_type` | `"charge"` (plain string) |
| invoices | `period_from` | `2024-01-01` (DATE) |
| coupons | `amount_off` | NULL (if percent coupon) |
| promotion_codes | `coupon_ref` | plain string ID |

---

## Phase 6 — Schedule Tests

| Test | Config | Expected |
|------|--------|---------|
| Daily sync | Cron `0 6 * * *` UTC | Runs at 06:00 each day |
| Every 6 hours | Cron `0 */6 * * *` | 4 runs per day |
| Manual only | None | No auto-trigger; "Run Now" only |
| Invalid cron | `99 * * * *` | ❌ "Invalid cron expression" |

---

## Phase 7 — Run Status Drawer Checks

After clicking **Run Now**:

1. Phase 0 ✅ — "Destination table `analytics.stripe_charges` found, PK `charge_id` verified"
2. Phase 1 ✅ — "Extracted N rows from `stripe.charges`"
3. Phase 2 ✅ — "dbt model executed: N rows output"
4. Phase 3 ✅ — "Delivered N rows to `analytics.stripe_charges`"
5. Phase 4 ✅ — "Staging cleaned"
6. Phase 5 ✅ — "Run complete"

### Failure scenarios to test

| Scenario | How to trigger | Expected drawer message |
|---------|---------------|------------------------|
| Bad API key | Enter invalid key | Phase 1 ❌ `401 Unauthorized` |
| Missing dest table | Drop table before run | Phase 0 ❌ `relation does not exist` |
| Bad dbt SQL | Set `SELEECT *` (typo) | Phase 2 ❌ `syntax error` |
| PK column missing in dest | Remove PK from DDL | Phase 0 ❌ PK mismatch |
| Concurrent run | Click Run Now twice fast | Second run ❌ `run already in progress` |

---

## Phase 8 — Destination Verification (PostgreSQL example)

```sql
-- Row count
SELECT COUNT(*) FROM analytics.stripe_charges;

-- No duplicates
SELECT charge_id, COUNT(*) FROM analytics.stripe_charges GROUP BY charge_id HAVING COUNT(*) > 1;

-- Amount conversion correct
SELECT charge_id, amount_dollars FROM analytics.stripe_charges WHERE amount_dollars > 0 LIMIT 5;
-- Must show e.g. 29.99, NOT 2999

-- JSON extracted correctly
SELECT card_brand, card_last4 FROM analytics.stripe_charges LIMIT 5;
-- Must show plain strings: "visa", "4242"

-- No staging tables
SELECT table_name FROM information_schema.tables
WHERE table_schema = 'analytics' AND table_name LIKE '_dlt_%';
-- 0 rows
```

---

## Incremental Re-run Test (charges stream)

1. Run pipeline → note row count N
2. Create a new charge in Stripe test dashboard
3. Wait 30 seconds
4. Run pipeline again
5. ✅ Row count = N+1; only new charge delivered
6. ❌ If count = N → cursor not advancing (check cursor field = `created`)
