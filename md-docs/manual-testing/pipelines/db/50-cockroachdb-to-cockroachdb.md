# Pipeline 50 — CockroachDB Source → CockroachDB Destination

**Source streams:** 3 | **Destination:** CockroachDB (different cluster/database)

> dbt SQL identical to `46-cockroachdb-to-postgres.md`. Types identical to PostgreSQL.

---

## Connections
```json
{ "host":"src-host","port":26257,"database":"defaultdb","username":"root","password":"","ssl_mode":"disable" }
{ "host":"dest-host","port":26257,"database":"defaultdb","username":"root","password":"","ssl_mode":"disable" }
```
```sql
-- On destination cluster:
CREATE DATABASE IF NOT EXISTS analytics;
CREATE SCHEMA IF NOT EXISTS analytics.analytics;
-- Or use defaultdb with analytics schema:
USE defaultdb;
CREATE SCHEMA IF NOT EXISTS analytics;
```

---

## Destination DDLs (CockroachDB)

```sql
CREATE TABLE IF NOT EXISTS analytics.account_ledger (
    account_id  UUID        PRIMARY KEY,
    email       STRING,
    plan_tier   STRING,
    balance_usd DECIMAL(12,2),
    tier_rank   INT8,
    is_paid     BOOL,
    joined_on   DATE,
    updated_on  DATE
);
CREATE TABLE IF NOT EXISTS analytics.txn_facts (
    txn_id      UUID PRIMARY KEY,
    account_ref UUID,
    txn_amount  DECIMAL(12,2),
    txn_type    STRING,
    is_credit   BOOL,
    note        STRING,
    txn_date    DATE,
    txn_month   STRING
);
CREATE TABLE IF NOT EXISTS analytics.session_registry (
    session_id    UUID PRIMARY KEY,
    user_ref      UUID,
    ip_address    STRING,
    browser_info  STRING,
    is_active     BOOL,
    duration_mins INT8,
    started_on    DATE,
    ended_on      DATE
);
```

---

## CockroachDB-to-CockroachDB Notes
- UUID native on both sides — no VARCHAR conversion needed
- `is_paid`, `is_credit`, `is_active` → `BOOL`: `true`/`false`
- Useful for cross-region replication testing or cluster migration

## Verify
```sql
SELECT plan_tier, tier_rank, is_paid FROM analytics.account_ledger GROUP BY plan_tier, tier_rank, is_paid;
-- enterprise→4→true, pro→3→true, starter→2→true, free→1→false
SELECT account_id FROM analytics.account_ledger LIMIT 3;  -- UUID type
SELECT txn_id, COUNT(*) FROM analytics.txn_facts GROUP BY txn_id HAVING COUNT(*) > 1;
-- Must be empty (no duplicates)
```

---

## Edge Cases

| Scenario | Expected |
|---------|---------|
| Same cluster, different schema | Works; schema isolation prevents collision |
| Same cluster, same schema | Phase 0 may fail on self-reference — use different schema |
| SSL required on source cluster | Add `"ssl_mode": "verify-full"` + cert |
| `plan = NULL` source rows | `COALESCE(plan, 'free')` in dbt SQL guards tier_rank |
