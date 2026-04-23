# Pipeline 04 — Stripe → SQLite

**Source streams:** 19 | **Destination:** SQLite | **dbt engine:** DuckDB

> dbt SQL identical to `01-stripe-to-postgres.md`.  
> SQLite uses dynamic type affinity — all types are advisory.

---

## Connections

### Source — Stripe
```json
{ "client_secret": "sk_live_..." }
```

### Destination — SQLite
```json
{ "database": "/absolute/path/to/analytics.db" }
```
```bash
sqlite3 /absolute/path/to/analytics.db
```

---

## SQLite Type Notes

| Intended type | SQLite column declaration | Actual storage |
|--------------|--------------------------|----------------|
| TIMESTAMPTZ | `TEXT` | ISO 8601 string |
| NUMERIC/DECIMAL | `REAL` | 64-bit float |
| BOOLEAN | `INTEGER` | 0 or 1 |
| TEXT/VARCHAR | `TEXT` | String |
| INTEGER/BIGINT | `INTEGER` | 64-bit int |

> ⚠️ SQLite ignores declared column types at write time. The ELT server writes actual Python types — booleans become 0/1 integers, datetimes become ISO strings.

---

## All 19 Stream DDLs

```sql
-- Stream 1: stripe.customers
CREATE TABLE IF NOT EXISTS stripe_customers (
    customer_id      TEXT    PRIMARY KEY,
    email_address    TEXT,
    display_name     TEXT,
    billing_currency TEXT,
    is_delinquent    INTEGER,   -- 0/1
    first_seen_at    TEXT,      -- ISO datetime
    metadata_keys    TEXT
);

-- Stream 2: stripe.charges
CREATE TABLE IF NOT EXISTS stripe_charges (
    charge_id        TEXT    PRIMARY KEY,
    customer_ref     TEXT,
    amount_dollars   REAL,
    currency         TEXT,
    charge_status    TEXT,
    card_brand       TEXT,
    card_last4       TEXT,
    refunded         INTEGER,   -- 0/1
    charged_at       TEXT
);

-- Stream 3: stripe.invoices
CREATE TABLE IF NOT EXISTS stripe_invoices (
    invoice_id       TEXT PRIMARY KEY,
    customer_ref     TEXT,
    subscription_ref TEXT,
    amount_due       REAL,
    amount_paid      REAL,
    invoice_status   TEXT,
    period_from      TEXT,   -- DATE as TEXT
    period_to        TEXT,
    issued_at        TEXT
);

-- Stream 4: stripe.subscriptions
CREATE TABLE IF NOT EXISTS stripe_subscriptions (
    sub_id           TEXT PRIMARY KEY,
    customer_ref     TEXT,
    plan_id          TEXT,
    plan_interval    TEXT,
    plan_amount      REAL,
    sub_status       TEXT,
    trial_ends_on    TEXT,
    renews_on        TEXT,
    cancelled_at     TEXT
);

-- Stream 5: stripe.products
CREATE TABLE IF NOT EXISTS stripe_products (
    product_id    TEXT PRIMARY KEY,
    product_name  TEXT,
    description   TEXT,
    is_active     INTEGER,
    unit_label    TEXT,
    created_at    TEXT,
    updated_at    TEXT
);

-- Stream 6: stripe.prices
CREATE TABLE IF NOT EXISTS stripe_prices (
    price_id       TEXT PRIMARY KEY,
    product_ref    TEXT,
    unit_amount    REAL,
    currency       TEXT,
    billing_scheme TEXT,
    price_type     TEXT,
    interval_unit  TEXT,
    interval_count INTEGER,
    is_active      INTEGER,
    created_at     TEXT
);

-- Stream 7: stripe.payment_intents
CREATE TABLE IF NOT EXISTS stripe_payment_intents (
    intent_id        TEXT PRIMARY KEY,
    customer_ref     TEXT,
    amount_dollars   REAL,
    currency         TEXT,
    intent_status    TEXT,
    payment_method   TEXT,
    is_captured      INTEGER,
    created_at       TEXT
);

-- Stream 8: stripe.payment_methods
CREATE TABLE IF NOT EXISTS stripe_payment_methods (
    pm_id         TEXT PRIMARY KEY,
    customer_ref  TEXT,
    pm_type       TEXT,
    card_brand    TEXT,
    card_last4    TEXT,
    card_expiry   TEXT,
    added_at      TEXT
);

-- Stream 9: stripe.refunds
CREATE TABLE IF NOT EXISTS stripe_refunds (
    refund_id      TEXT PRIMARY KEY,
    charge_ref     TEXT,
    refund_amount  REAL,
    currency       TEXT,
    refund_reason  TEXT,
    refund_status  TEXT,
    refunded_at    TEXT
);

-- Stream 10: stripe.disputes
CREATE TABLE IF NOT EXISTS stripe_disputes (
    dispute_id      TEXT PRIMARY KEY,
    charge_ref      TEXT,
    dispute_amount  REAL,
    currency        TEXT,
    dispute_reason  TEXT,
    dispute_status  TEXT,
    is_won          INTEGER,
    created_at      TEXT
);

-- Stream 11: stripe.events
CREATE TABLE IF NOT EXISTS stripe_events (
    event_id      TEXT PRIMARY KEY,
    event_type    TEXT,
    object_type   TEXT,
    object_id     TEXT,
    api_version   TEXT,
    is_live       INTEGER,
    occurred_at   TEXT
);

-- Stream 12: stripe.customers_balance_transactions
CREATE TABLE IF NOT EXISTS stripe_balance_txns (
    txn_id        TEXT PRIMARY KEY,
    customer_ref  TEXT,
    amount        REAL,
    currency      TEXT,
    txn_type      TEXT,
    description   TEXT,
    txn_at        TEXT
);

-- Stream 13: stripe.coupons
CREATE TABLE IF NOT EXISTS stripe_coupons (
    coupon_id     TEXT PRIMARY KEY,
    coupon_name   TEXT,
    discount_type TEXT,
    pct_off       REAL,
    amount_off    REAL,
    currency      TEXT,
    is_valid      INTEGER,
    max_uses      INTEGER,
    use_count     INTEGER,
    created_at    TEXT
);

-- Stream 14: stripe.promotion_codes
CREATE TABLE IF NOT EXISTS stripe_promo_codes (
    promo_id      TEXT PRIMARY KEY,
    coupon_ref    TEXT,
    promo_code    TEXT,
    is_active     INTEGER,
    max_uses      INTEGER,
    use_count     INTEGER,
    expires_on    TEXT,
    created_at    TEXT
);

-- Stream 15: stripe.tax_rates
CREATE TABLE IF NOT EXISTS stripe_tax_rates (
    rate_id       TEXT PRIMARY KEY,
    display_name  TEXT,
    jurisdiction  TEXT,
    rate_pct      REAL,
    is_inclusive  INTEGER,
    is_active     INTEGER,
    created_at    TEXT
);

-- Stream 16: stripe.setup_intents
CREATE TABLE IF NOT EXISTS stripe_setup_intents (
    intent_id      TEXT PRIMARY KEY,
    customer_ref   TEXT,
    payment_method TEXT,
    intent_status  TEXT,
    usage          TEXT,
    created_at     TEXT
);

-- Stream 17: stripe.credit_notes
CREATE TABLE IF NOT EXISTS stripe_credit_notes (
    note_id        TEXT PRIMARY KEY,
    customer_ref   TEXT,
    invoice_ref    TEXT,
    note_amount    REAL,
    currency       TEXT,
    note_status    TEXT,
    note_type      TEXT,
    created_at     TEXT
);

-- Stream 18: stripe.early_fraud_warnings
CREATE TABLE IF NOT EXISTS stripe_fraud_warnings (
    warning_id      TEXT PRIMARY KEY,
    charge_ref      TEXT,
    fraud_type      TEXT,
    actionable      INTEGER,
    created_at      TEXT
);

-- Stream 19: stripe.customers_bank_accounts
CREATE TABLE IF NOT EXISTS stripe_bank_accounts (
    account_id    TEXT PRIMARY KEY,
    customer_ref  TEXT,
    bank_name     TEXT,
    routing_num   TEXT,
    last4         TEXT,
    account_type  TEXT,
    acct_status   TEXT,
    currency      TEXT
);
```

> ℹ️ SQLite destination uses `main` schema. Set destination panel to `main.stripe_customers` etc.

---

## SQLite Verification Commands

```bash
DB=/absolute/path/to/analytics.db

# Row counts
sqlite3 $DB "SELECT COUNT(*) FROM stripe_customers;"
sqlite3 $DB "SELECT COUNT(*) FROM stripe_charges;"

# Type check (amount_dollars must be REAL)
sqlite3 $DB "SELECT typeof(amount_dollars), amount_dollars FROM stripe_charges LIMIT 3;"
# Expected: real|29.99

# Boolean stored as INTEGER
sqlite3 $DB "SELECT is_delinquent FROM stripe_customers LIMIT 5;"
# Expected: 0 or 1 only

# Datetime stored as TEXT
sqlite3 $DB "SELECT first_seen_at FROM stripe_customers LIMIT 3;"
# Expected: 2024-01-15T10:30:00+00:00

# Duplicate check
sqlite3 $DB "SELECT charge_id, COUNT(*) cnt FROM stripe_charges GROUP BY charge_id HAVING cnt > 1;"

# No _dlt_ tables
sqlite3 $DB ".tables"
# Must not show _dlt_xxx tables
```

---

## SQLite Edge Cases

| Scenario | Expected |
|---------|---------|
| `first_seen_at` stored as TEXT | ISO 8601 string — cast with `datetime(first_seen_at)` in SQLite queries |
| `amount_dollars` stored as REAL | `0.10 + 0.20` may show float rounding — expected SQLite behavior |
| File not writable | Phase 3 delivery fails: `unable to open database file` |
| Multiple concurrent runs | Second run may hit `database is locked` — expected |
