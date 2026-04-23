# Pipeline 49 — CockroachDB Source → SQLite Destination

**Source streams:** 3 | **Destination:** SQLite

> dbt SQL identical to `46-cockroachdb-to-postgres.md`.

---

## Connections
```json
{ "host":"src-host","port":26257,"database":"defaultdb","username":"root","password":"","ssl_mode":"disable" }
{ "database": "/absolute/path/to/analytics.db" }
```

---

## Destination DDLs (SQLite)

```sql
CREATE TABLE IF NOT EXISTS account_ledger (
    account_id  TEXT PRIMARY KEY,     -- UUID as TEXT
    email       TEXT,
    plan_tier   TEXT,
    balance_usd REAL,
    tier_rank   INTEGER,
    is_paid     INTEGER,              -- 0/1
    joined_on   TEXT,
    updated_on  TEXT
);
CREATE TABLE IF NOT EXISTS txn_facts (
    txn_id      TEXT PRIMARY KEY,
    account_ref TEXT,
    txn_amount  REAL,
    txn_type    TEXT,
    is_credit   INTEGER,
    note        TEXT,
    txn_date    TEXT,
    txn_month   TEXT
);
CREATE TABLE IF NOT EXISTS session_registry (
    session_id    TEXT PRIMARY KEY,
    user_ref      TEXT,
    ip_address    TEXT,
    browser_info  TEXT,
    is_active     INTEGER,
    duration_mins INTEGER,
    started_on    TEXT,
    ended_on      TEXT
);
```

---

## Verify
```bash
DB=/absolute/path/to/analytics.db
sqlite3 $DB "SELECT plan_tier, tier_rank, is_paid FROM account_ledger GROUP BY plan_tier, tier_rank, is_paid;"
sqlite3 $DB "SELECT typeof(balance_usd), balance_usd FROM account_ledger LIMIT 3;"  # real
sqlite3 $DB "SELECT txn_month FROM txn_facts GROUP BY txn_month;"
```
