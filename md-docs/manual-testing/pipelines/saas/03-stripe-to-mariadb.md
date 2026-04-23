# Pipeline 03 — Stripe → MariaDB

**Source streams:** 19 | **Destination:** MariaDB | **dbt engine:** DuckDB

> MariaDB is MySQL-wire-compatible. DDLs are identical to `02-stripe-to-mysql.md`.  
> dbt SQL is identical to `01-stripe-to-postgres.md`.  
> This file documents MariaDB-specific differences only.

---

## Connections

### Source — Stripe
```json
{ "client_secret": "sk_live_..." }
```

### Destination — MariaDB
```json
{ "host": "...", "port": 3306, "database": "analytics", "username": "writer", "password": "..." }
```
```sql
CREATE DATABASE IF NOT EXISTS analytics;
GRANT INSERT, UPDATE, SELECT ON analytics.* TO 'writer'@'%';
```

---

## MariaDB vs MySQL DDL Differences

| Feature | MariaDB | MySQL |
|---------|---------|-------|
| JSON type | `LONGTEXT` with JSON functions (10.2+) | Native `JSON` |
| `BOOLEAN` alias | `TINYINT(1)` | `TINYINT(1)` |
| Default charset | `utf8mb4` from 10.4+ | Varies |
| `INSERT … ON DUPLICATE KEY UPDATE` | ✅ | ✅ |

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
    first_seen_at    DATETIME(6),
    metadata_keys    LONGTEXT
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
    charged_at       DATETIME(6)
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
    issued_at        DATETIME(6)
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
    cancelled_at     DATETIME(6)
) ENGINE=InnoDB;

-- Stream 5: stripe.products
CREATE TABLE analytics.stripe_products (
    product_id    VARCHAR(255) PRIMARY KEY,
    product_name  VARCHAR(500),
    description   LONGTEXT,
    is_active     TINYINT(1),
    unit_label    VARCHAR(100),
    created_at    DATETIME(6),
    updated_at    DATETIME(6)
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
    created_at     DATETIME(6)
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
    created_at       DATETIME(6)
) ENGINE=InnoDB;

-- Stream 8: stripe.payment_methods
CREATE TABLE analytics.stripe_payment_methods (
    pm_id         VARCHAR(255) PRIMARY KEY,
    customer_ref  VARCHAR(255),
    pm_type       VARCHAR(50),
    card_brand    VARCHAR(50),
    card_last4    VARCHAR(10),
    card_expiry   VARCHAR(10),
    added_at      DATETIME(6)
) ENGINE=InnoDB;

-- Stream 9: stripe.refunds
CREATE TABLE analytics.stripe_refunds (
    refund_id      VARCHAR(255) PRIMARY KEY,
    charge_ref     VARCHAR(255),
    refund_amount  DECIMAL(10,2),
    currency       VARCHAR(10),
    refund_reason  VARCHAR(100),
    refund_status  VARCHAR(50),
    refunded_at    DATETIME(6)
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
    created_at      DATETIME(6)
) ENGINE=InnoDB;

-- Stream 11: stripe.events
CREATE TABLE analytics.stripe_events (
    event_id      VARCHAR(255) PRIMARY KEY,
    event_type    VARCHAR(100),
    object_type   VARCHAR(100),
    object_id     VARCHAR(255),
    api_version   VARCHAR(20),
    is_live       TINYINT(1),
    occurred_at   DATETIME(6)
) ENGINE=InnoDB;

-- Stream 12: stripe.customers_balance_transactions
CREATE TABLE analytics.stripe_balance_txns (
    txn_id        VARCHAR(255) PRIMARY KEY,
    customer_ref  VARCHAR(255),
    amount        DECIMAL(10,2),
    currency      VARCHAR(10),
    txn_type      VARCHAR(100),
    description   LONGTEXT,
    txn_at        DATETIME(6)
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
    created_at    DATETIME(6)
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
    created_at    DATETIME(6)
) ENGINE=InnoDB;

-- Stream 15: stripe.tax_rates
CREATE TABLE analytics.stripe_tax_rates (
    rate_id       VARCHAR(255) PRIMARY KEY,
    display_name  VARCHAR(255),
    jurisdiction  VARCHAR(100),
    rate_pct      DECIMAL(6,4),
    is_inclusive  TINYINT(1),
    is_active     TINYINT(1),
    created_at    DATETIME(6)
) ENGINE=InnoDB;

-- Stream 16: stripe.setup_intents
CREATE TABLE analytics.stripe_setup_intents (
    intent_id      VARCHAR(255) PRIMARY KEY,
    customer_ref   VARCHAR(255),
    payment_method VARCHAR(255),
    intent_status  VARCHAR(50),
    usage          VARCHAR(50),
    created_at     DATETIME(6)
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
    created_at     DATETIME(6)
) ENGINE=InnoDB;

-- Stream 18: stripe.early_fraud_warnings
CREATE TABLE analytics.stripe_fraud_warnings (
    warning_id      VARCHAR(255) PRIMARY KEY,
    charge_ref      VARCHAR(255),
    fraud_type      VARCHAR(100),
    actionable      TINYINT(1),
    created_at      DATETIME(6)
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

## MariaDB-Specific Verification

```sql
-- DATETIME(6) stores microseconds — verify precision
SELECT first_seen_at FROM analytics.stripe_customers LIMIT 3;

-- TINYINT(1) for booleans
SELECT is_active, COUNT(*) FROM analytics.stripe_products GROUP BY is_active;
-- Returns 0 and 1 only

-- Duplicate check
SELECT charge_id, COUNT(*) cnt FROM analytics.stripe_charges
GROUP BY charge_id HAVING cnt > 1;

-- Verify no _dlt_ tables
SELECT TABLE_NAME FROM information_schema.TABLES
WHERE TABLE_SCHEMA = 'analytics' AND TABLE_NAME LIKE '_dlt_%';
```
