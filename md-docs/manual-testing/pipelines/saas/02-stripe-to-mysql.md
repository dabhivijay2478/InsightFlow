# Pipeline 02 — Stripe → MySQL

**Source streams:** 19 | **Destination:** MySQL | **dbt engine:** DuckDB

> dbt SQL is identical to `01-stripe-to-postgres.md` for every stream.  
> Only the **destination DDL** changes to MySQL types.

---

## Connections

### Source — Stripe
```json
{ "client_secret": "sk_live_..." }
```

### Destination — MySQL
```json
{ "host": "...", "port": 3306, "database": "analytics", "username": "writer", "password": "..." }
```
```sql
CREATE DATABASE IF NOT EXISTS analytics;
GRANT INSERT, UPDATE, SELECT ON analytics.* TO 'writer'@'%';
```

---

## MySQL Type Map (Stripe API → MySQL)

| Stripe / DuckDB type | MySQL column type |
|---------------------|------------------|
| Unix epoch → TIMESTAMPTZ | `DATETIME` |
| Amount cents / 100 | `DECIMAL(10,2)` |
| Boolean | `TINYINT(1)` |
| String / TEXT | `VARCHAR(255)` or `TEXT` |
| JSONB (extracted key) | `VARCHAR(500)` |
| Integer | `BIGINT` |

---

## All 19 Stream DDLs

```sql
-- Stream 1: stripe.customers
CREATE TABLE analytics.stripe_customers (
    customer_id      VARCHAR(255) PRIMARY KEY,
    email_address    VARCHAR(255),
    display_name     VARCHAR(255),
    billing_currency VARCHAR(10),
    is_delinquent    TINYINT(1),
    first_seen_at    DATETIME,
    metadata_keys    TEXT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- Stream 2: stripe.charges
CREATE TABLE analytics.stripe_charges (
    charge_id        VARCHAR(255) PRIMARY KEY,
    customer_ref     VARCHAR(255),
    amount_dollars   DECIMAL(10,2),
    currency         VARCHAR(10),
    charge_status    VARCHAR(50),
    card_brand       VARCHAR(50),
    card_last4       VARCHAR(10),
    refunded         TINYINT(1),
    charged_at       DATETIME
) ENGINE=InnoDB;

-- Stream 3: stripe.invoices
CREATE TABLE analytics.stripe_invoices (
    invoice_id       VARCHAR(255) PRIMARY KEY,
    customer_ref     VARCHAR(255),
    subscription_ref VARCHAR(255),
    amount_due       DECIMAL(10,2),
    amount_paid      DECIMAL(10,2),
    invoice_status   VARCHAR(50),
    period_from      DATE,
    period_to        DATE,
    issued_at        DATETIME
) ENGINE=InnoDB;

-- Stream 4: stripe.subscriptions
CREATE TABLE analytics.stripe_subscriptions (
    sub_id           VARCHAR(255) PRIMARY KEY,
    customer_ref     VARCHAR(255),
    plan_id          VARCHAR(255),
    plan_interval    VARCHAR(50),
    plan_amount      DECIMAL(10,2),
    sub_status       VARCHAR(50),
    trial_ends_on    DATE,
    renews_on        DATE,
    cancelled_at     DATETIME
) ENGINE=InnoDB;

-- Stream 5: stripe.products
CREATE TABLE analytics.stripe_products (
    product_id    VARCHAR(255) PRIMARY KEY,
    product_name  VARCHAR(500),
    description   TEXT,
    is_active     TINYINT(1),
    unit_label    VARCHAR(100),
    created_at    DATETIME,
    updated_at    DATETIME
) ENGINE=InnoDB;

-- Stream 6: stripe.prices
CREATE TABLE analytics.stripe_prices (
    price_id       VARCHAR(255) PRIMARY KEY,
    product_ref    VARCHAR(255),
    unit_amount    DECIMAL(10,2),
    currency       VARCHAR(10),
    billing_scheme VARCHAR(50),
    price_type     VARCHAR(50),
    interval_unit  VARCHAR(20),
    interval_count INT,
    is_active      TINYINT(1),
    created_at     DATETIME
) ENGINE=InnoDB;

-- Stream 7: stripe.payment_intents
CREATE TABLE analytics.stripe_payment_intents (
    intent_id        VARCHAR(255) PRIMARY KEY,
    customer_ref     VARCHAR(255),
    amount_dollars   DECIMAL(10,2),
    currency         VARCHAR(10),
    intent_status    VARCHAR(50),
    payment_method   VARCHAR(255),
    is_captured      TINYINT(1),
    created_at       DATETIME
) ENGINE=InnoDB;

-- Stream 8: stripe.payment_methods
CREATE TABLE analytics.stripe_payment_methods (
    pm_id         VARCHAR(255) PRIMARY KEY,
    customer_ref  VARCHAR(255),
    pm_type       VARCHAR(50),
    card_brand    VARCHAR(50),
    card_last4    VARCHAR(10),
    card_expiry   VARCHAR(10),
    added_at      DATETIME
) ENGINE=InnoDB;

-- Stream 9: stripe.refunds
CREATE TABLE analytics.stripe_refunds (
    refund_id      VARCHAR(255) PRIMARY KEY,
    charge_ref     VARCHAR(255),
    refund_amount  DECIMAL(10,2),
    currency       VARCHAR(10),
    refund_reason  VARCHAR(100),
    refund_status  VARCHAR(50),
    refunded_at    DATETIME
) ENGINE=InnoDB;

-- Stream 10: stripe.disputes
CREATE TABLE analytics.stripe_disputes (
    dispute_id      VARCHAR(255) PRIMARY KEY,
    charge_ref      VARCHAR(255),
    dispute_amount  DECIMAL(10,2),
    currency        VARCHAR(10),
    dispute_reason  VARCHAR(100),
    dispute_status  VARCHAR(50),
    is_won          TINYINT(1),
    created_at      DATETIME
) ENGINE=InnoDB;

-- Stream 11: stripe.events
CREATE TABLE analytics.stripe_events (
    event_id      VARCHAR(255) PRIMARY KEY,
    event_type    VARCHAR(100),
    object_type   VARCHAR(100),
    object_id     VARCHAR(255),
    api_version   VARCHAR(20),
    is_live       TINYINT(1),
    occurred_at   DATETIME
) ENGINE=InnoDB;

-- Stream 12: stripe.customers_balance_transactions
CREATE TABLE analytics.stripe_balance_txns (
    txn_id        VARCHAR(255) PRIMARY KEY,
    customer_ref  VARCHAR(255),
    amount        DECIMAL(10,2),
    currency      VARCHAR(10),
    txn_type      VARCHAR(100),
    description   TEXT,
    txn_at        DATETIME
) ENGINE=InnoDB;

-- Stream 13: stripe.coupons
CREATE TABLE analytics.stripe_coupons (
    coupon_id     VARCHAR(255) PRIMARY KEY,
    coupon_name   VARCHAR(255),
    discount_type VARCHAR(50),
    pct_off       DECIMAL(5,2),
    amount_off    DECIMAL(10,2),
    currency      VARCHAR(10),
    is_valid      TINYINT(1),
    max_uses      INT,
    use_count     INT,
    created_at    DATETIME
) ENGINE=InnoDB;

-- Stream 14: stripe.promotion_codes
CREATE TABLE analytics.stripe_promo_codes (
    promo_id      VARCHAR(255) PRIMARY KEY,
    coupon_ref    VARCHAR(255),
    promo_code    VARCHAR(100),
    is_active     TINYINT(1),
    max_uses      INT,
    use_count     INT,
    expires_on    DATE,
    created_at    DATETIME
) ENGINE=InnoDB;

-- Stream 15: stripe.tax_rates
CREATE TABLE analytics.stripe_tax_rates (
    rate_id       VARCHAR(255) PRIMARY KEY,
    display_name  VARCHAR(255),
    jurisdiction  VARCHAR(100),
    rate_pct      DECIMAL(6,4),
    is_inclusive  TINYINT(1),
    is_active     TINYINT(1),
    created_at    DATETIME
) ENGINE=InnoDB;

-- Stream 16: stripe.setup_intents
CREATE TABLE analytics.stripe_setup_intents (
    intent_id      VARCHAR(255) PRIMARY KEY,
    customer_ref   VARCHAR(255),
    payment_method VARCHAR(255),
    intent_status  VARCHAR(50),
    usage          VARCHAR(50),
    created_at     DATETIME
) ENGINE=InnoDB;

-- Stream 17: stripe.credit_notes
CREATE TABLE analytics.stripe_credit_notes (
    note_id        VARCHAR(255) PRIMARY KEY,
    customer_ref   VARCHAR(255),
    invoice_ref    VARCHAR(255),
    note_amount    DECIMAL(10,2),
    currency       VARCHAR(10),
    note_status    VARCHAR(50),
    note_type      VARCHAR(50),
    created_at     DATETIME
) ENGINE=InnoDB;

-- Stream 18: stripe.early_fraud_warnings
CREATE TABLE analytics.stripe_fraud_warnings (
    warning_id      VARCHAR(255) PRIMARY KEY,
    charge_ref      VARCHAR(255),
    fraud_type      VARCHAR(100),
    actionable      TINYINT(1),
    created_at      DATETIME
) ENGINE=InnoDB;

-- Stream 19: stripe.customers_bank_accounts
CREATE TABLE analytics.stripe_bank_accounts (
    account_id    VARCHAR(255) PRIMARY KEY,
    customer_ref  VARCHAR(255),
    bank_name     VARCHAR(255),
    routing_num   VARCHAR(50),
    last4         VARCHAR(10),
    account_type  VARCHAR(50),
    acct_status   VARCHAR(50),
    currency      VARCHAR(10)
) ENGINE=InnoDB;
```

---

## MySQL-Specific Verification Queries

```sql
-- Boolean stored as TINYINT(1) — verify correct values
SELECT is_delinquent, COUNT(*) FROM analytics.stripe_customers GROUP BY is_delinquent;
-- Must return only 0 and 1

-- DATETIME (no timezone) — verify no truncation
SELECT first_seen_at FROM analytics.stripe_customers LIMIT 3;
-- Returns: 2024-01-15 10:30:00 (no +00 offset)

-- Decimal precision
SELECT amount_dollars FROM analytics.stripe_charges WHERE amount_dollars > 0 LIMIT 5;
-- Must show e.g. 29.99, not 2999

-- Duplicate check (merge mode)
SELECT charge_id, COUNT(*) cnt FROM analytics.stripe_charges
GROUP BY charge_id HAVING cnt > 1;
-- Must return 0 rows

-- No _dlt_ tables
SELECT TABLE_NAME FROM information_schema.TABLES
WHERE TABLE_SCHEMA = 'analytics' AND TABLE_NAME LIKE '_dlt_%';
```

---

## MySQL Edge Cases

| Scenario | Expected |
|---------|---------|
| `TINYINT(1)` bool column receives `true`/`false` | Stored as `1`/`0` |
| Long `description` TEXT > 65535 chars | Use `MEDIUMTEXT` or `LONGTEXT` |
| `VARCHAR(255)` receives Stripe ID > 255 chars | Truncation error — increase to `VARCHAR(500)` |
| Missing table | Phase 0 fails: `Table 'analytics.stripe_charges' doesn't exist` |
