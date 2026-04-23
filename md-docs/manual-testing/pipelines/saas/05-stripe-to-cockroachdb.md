# Pipeline 05 — Stripe → CockroachDB

**Source streams:** 19 | **Destination:** CockroachDB | **dbt engine:** DuckDB

> CockroachDB is PostgreSQL-wire-compatible. DDLs nearly identical to `01-stripe-to-postgres.md`.  
> dbt SQL is identical to `01-stripe-to-postgres.md`.

---

## Connections

### Source — Stripe
```json
{ "client_secret": "sk_live_..." }
```

### Destination — CockroachDB
```json
{ "host": "...", "port": 26257, "database": "defaultdb", "username": "root", "password": "", "ssl_mode": "disable" }
```
```sql
CREATE SCHEMA IF NOT EXISTS analytics;
```

---

## CockroachDB vs PostgreSQL Differences

| Feature | CockroachDB | PostgreSQL |
|---------|-------------|-----------|
| `TEXT` alias | `STRING` (same) | `TEXT` |
| `BIGINT` alias | `INT8` (same) | `BIGINT` |
| `TIMESTAMPTZ` | ✅ identical | ✅ |
| `NUMERIC` | ✅ identical | ✅ |
| `BOOLEAN` | ✅ identical | ✅ |
| `JSONB` | ✅ identical | ✅ |
| UUID auto-gen | `gen_random_uuid()` | `gen_random_uuid()` |
| Upsert | `INSERT … ON CONFLICT DO UPDATE` | same |

---

## All 19 Stream DDLs

```sql
-- Stream 1: stripe.customers
CREATE TABLE IF NOT EXISTS analytics.stripe_customers (
    customer_id      STRING      PRIMARY KEY,
    email_address    STRING,
    display_name     STRING,
    billing_currency STRING,
    is_delinquent    BOOL,
    first_seen_at    TIMESTAMPTZ,
    metadata_keys    STRING
);

-- Stream 2: stripe.charges
CREATE TABLE IF NOT EXISTS analytics.stripe_charges (
    charge_id        STRING      PRIMARY KEY,
    customer_ref     STRING,
    amount_dollars   DECIMAL(10,2),
    currency         STRING,
    charge_status    STRING,
    card_brand       STRING,
    card_last4       STRING,
    refunded         BOOL,
    charged_at       TIMESTAMPTZ
);

-- Stream 3: stripe.invoices
CREATE TABLE IF NOT EXISTS analytics.stripe_invoices (
    invoice_id       STRING      PRIMARY KEY,
    customer_ref     STRING,
    subscription_ref STRING,
    amount_due       DECIMAL(10,2),
    amount_paid      DECIMAL(10,2),
    invoice_status   STRING,
    period_from      DATE,
    period_to        DATE,
    issued_at        TIMESTAMPTZ
);

-- Stream 4: stripe.subscriptions
CREATE TABLE IF NOT EXISTS analytics.stripe_subscriptions (
    sub_id           STRING      PRIMARY KEY,
    customer_ref     STRING,
    plan_id          STRING,
    plan_interval    STRING,
    plan_amount      DECIMAL(10,2),
    sub_status       STRING,
    trial_ends_on    DATE,
    renews_on        DATE,
    cancelled_at     TIMESTAMPTZ
);

-- Stream 5: stripe.products
CREATE TABLE IF NOT EXISTS analytics.stripe_products (
    product_id    STRING PRIMARY KEY,
    product_name  STRING,
    description   STRING,
    is_active     BOOL,
    unit_label    STRING,
    created_at    TIMESTAMPTZ,
    updated_at    TIMESTAMPTZ
);

-- Stream 6: stripe.prices
CREATE TABLE IF NOT EXISTS analytics.stripe_prices (
    price_id       STRING PRIMARY KEY,
    product_ref    STRING,
    unit_amount    DECIMAL(10,2),
    currency       STRING,
    billing_scheme STRING,
    price_type     STRING,
    interval_unit  STRING,
    interval_count INT8,
    is_active      BOOL,
    created_at     TIMESTAMPTZ
);

-- Stream 7: stripe.payment_intents
CREATE TABLE IF NOT EXISTS analytics.stripe_payment_intents (
    intent_id        STRING PRIMARY KEY,
    customer_ref     STRING,
    amount_dollars   DECIMAL(10,2),
    currency         STRING,
    intent_status    STRING,
    payment_method   STRING,
    is_captured      BOOL,
    created_at       TIMESTAMPTZ
);

-- Stream 8: stripe.payment_methods
CREATE TABLE IF NOT EXISTS analytics.stripe_payment_methods (
    pm_id         STRING PRIMARY KEY,
    customer_ref  STRING,
    pm_type       STRING,
    card_brand    STRING,
    card_last4    STRING,
    card_expiry   STRING,
    added_at      TIMESTAMPTZ
);

-- Stream 9: stripe.refunds
CREATE TABLE IF NOT EXISTS analytics.stripe_refunds (
    refund_id      STRING PRIMARY KEY,
    charge_ref     STRING,
    refund_amount  DECIMAL(10,2),
    currency       STRING,
    refund_reason  STRING,
    refund_status  STRING,
    refunded_at    TIMESTAMPTZ
);

-- Stream 10: stripe.disputes
CREATE TABLE IF NOT EXISTS analytics.stripe_disputes (
    dispute_id      STRING PRIMARY KEY,
    charge_ref      STRING,
    dispute_amount  DECIMAL(10,2),
    currency        STRING,
    dispute_reason  STRING,
    dispute_status  STRING,
    is_won          BOOL,
    created_at      TIMESTAMPTZ
);

-- Stream 11: stripe.events
CREATE TABLE IF NOT EXISTS analytics.stripe_events (
    event_id      STRING PRIMARY KEY,
    event_type    STRING,
    object_type   STRING,
    object_id     STRING,
    api_version   STRING,
    is_live       BOOL,
    occurred_at   TIMESTAMPTZ
);

-- Stream 12: stripe.customers_balance_transactions
CREATE TABLE IF NOT EXISTS analytics.stripe_balance_txns (
    txn_id        STRING PRIMARY KEY,
    customer_ref  STRING,
    amount        DECIMAL(10,2),
    currency      STRING,
    txn_type      STRING,
    description   STRING,
    txn_at        TIMESTAMPTZ
);

-- Stream 13: stripe.coupons
CREATE TABLE IF NOT EXISTS analytics.stripe_coupons (
    coupon_id     STRING PRIMARY KEY,
    coupon_name   STRING,
    discount_type STRING,
    pct_off       DECIMAL(5,2),
    amount_off    DECIMAL(10,2),
    currency      STRING,
    is_valid      BOOL,
    max_uses      INT8,
    use_count     INT8,
    created_at    TIMESTAMPTZ
);

-- Stream 14: stripe.promotion_codes
CREATE TABLE IF NOT EXISTS analytics.stripe_promo_codes (
    promo_id      STRING PRIMARY KEY,
    coupon_ref    STRING,
    promo_code    STRING,
    is_active     BOOL,
    max_uses      INT8,
    use_count     INT8,
    expires_on    DATE,
    created_at    TIMESTAMPTZ
);

-- Stream 15: stripe.tax_rates
CREATE TABLE IF NOT EXISTS analytics.stripe_tax_rates (
    rate_id       STRING PRIMARY KEY,
    display_name  STRING,
    jurisdiction  STRING,
    rate_pct      DECIMAL(6,4),
    is_inclusive  BOOL,
    is_active     BOOL,
    created_at    TIMESTAMPTZ
);

-- Stream 16: stripe.setup_intents
CREATE TABLE IF NOT EXISTS analytics.stripe_setup_intents (
    intent_id      STRING PRIMARY KEY,
    customer_ref   STRING,
    payment_method STRING,
    intent_status  STRING,
    usage          STRING,
    created_at     TIMESTAMPTZ
);

-- Stream 17: stripe.credit_notes
CREATE TABLE IF NOT EXISTS analytics.stripe_credit_notes (
    note_id        STRING PRIMARY KEY,
    customer_ref   STRING,
    invoice_ref    STRING,
    note_amount    DECIMAL(10,2),
    currency       STRING,
    note_status    STRING,
    note_type      STRING,
    created_at     TIMESTAMPTZ
);

-- Stream 18: stripe.early_fraud_warnings
CREATE TABLE IF NOT EXISTS analytics.stripe_fraud_warnings (
    warning_id      STRING PRIMARY KEY,
    charge_ref      STRING,
    fraud_type      STRING,
    actionable      BOOL,
    created_at      TIMESTAMPTZ
);

-- Stream 19: stripe.customers_bank_accounts
CREATE TABLE IF NOT EXISTS analytics.stripe_bank_accounts (
    account_id    STRING PRIMARY KEY,
    customer_ref  STRING,
    bank_name     STRING,
    routing_num   STRING,
    last4         STRING,
    account_type  STRING,
    acct_status   STRING,
    currency      STRING
);
```

---

## CockroachDB Verification

```sql
-- Row counts
SELECT COUNT(*) FROM analytics.stripe_customers;
SELECT COUNT(*) FROM analytics.stripe_charges;

-- Boolean values
SELECT is_delinquent FROM analytics.stripe_customers LIMIT 5;
-- Returns: true / false (not 0/1)

-- TIMESTAMPTZ precision
SELECT first_seen_at FROM analytics.stripe_customers LIMIT 3;

-- Decimal precision
SELECT amount_dollars FROM analytics.stripe_charges LIMIT 5;
-- Must show 29.99, not 2999

-- Duplicate check
SELECT charge_id, COUNT(*) FROM analytics.stripe_charges
GROUP BY charge_id HAVING COUNT(*) > 1;

-- No _dlt_ tables
SELECT table_name FROM information_schema.tables
WHERE table_schema = 'analytics' AND table_name LIKE '_dlt_%';
```
