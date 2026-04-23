# Pipeline 47 — CockroachDB Source → MySQL Destination

**Source streams:** 3 | **Destination:** MySQL

> dbt SQL identical to `46-cockroachdb-to-postgres.md`.

---

## Connections
```json
{ "host":"src-host","port":26257,"database":"defaultdb","username":"root","password":"","ssl_mode":"disable" }
{ "host":"dest-host","port":3306,"database":"analytics","username":"writer","password":"..." }
```

---

## Destination DDLs (MySQL)

```sql
CREATE TABLE analytics.account_ledger (
    account_id   VARCHAR(36) PRIMARY KEY,     -- UUID as VARCHAR
    email        VARCHAR(255),
    plan_tier    VARCHAR(50),
    balance_usd  DECIMAL(12,2),
    tier_rank    INT,
    is_paid      TINYINT(1),
    joined_on    DATE,
    updated_on   DATE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE analytics.txn_facts (
    txn_id      VARCHAR(36) PRIMARY KEY,
    account_ref VARCHAR(36),
    txn_amount  DECIMAL(12,2),
    txn_type    VARCHAR(50),
    is_credit   TINYINT(1),
    note        TEXT,
    txn_date    DATE,
    txn_month   VARCHAR(30)
) ENGINE=InnoDB;

CREATE TABLE analytics.session_registry (
    session_id    VARCHAR(36) PRIMARY KEY,
    user_ref      VARCHAR(36),
    ip_address    VARCHAR(45),
    browser_info  TEXT,
    is_active     TINYINT(1),
    duration_mins INT,
    started_on    DATE,
    ended_on      DATE
) ENGINE=InnoDB;
```

---

## MySQL-Specific Notes
- UUID stored as `VARCHAR(36)` — no native UUID in MySQL
- `is_paid`, `is_credit`, `is_active` → `TINYINT(1)` stores 0/1

## Verify
```sql
SELECT plan_tier, tier_rank, is_paid FROM analytics.account_ledger GROUP BY plan_tier, tier_rank, is_paid;
SELECT account_id FROM analytics.account_ledger LIMIT 3;   -- UUID string '550e8400-...'
```
