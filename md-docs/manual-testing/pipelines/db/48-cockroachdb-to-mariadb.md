# Pipeline 48 — CockroachDB Source → MariaDB Destination

**Source streams:** 3 | **Destination:** MariaDB

> dbt SQL identical to `46-cockroachdb-to-postgres.md`. DDL same as `47-cockroachdb-to-mysql.md`.

---

## Connections
```json
{ "host":"src-host","port":26257,"database":"defaultdb","username":"root","password":"","ssl_mode":"disable" }
{ "host":"dest-host","port":3306,"database":"analytics","username":"writer","password":"..." }
```

---

## Destination DDLs (MariaDB)

```sql
CREATE TABLE analytics.account_ledger (
    account_id   VARCHAR(36) PRIMARY KEY,
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
    note        LONGTEXT,
    txn_date    DATE,
    txn_month   VARCHAR(30)
) ENGINE=InnoDB;

CREATE TABLE analytics.session_registry (
    session_id    VARCHAR(36) PRIMARY KEY,
    user_ref      VARCHAR(36),
    ip_address    VARCHAR(45),
    browser_info  LONGTEXT,
    is_active     TINYINT(1),
    duration_mins INT,
    started_on    DATE,
    ended_on      DATE
) ENGINE=InnoDB;
```
