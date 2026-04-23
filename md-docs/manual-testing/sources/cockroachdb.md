# CockroachDB Source — Manual Testing Guide

**Streams:** 3 (`public.sessions`, `public.accounts`, `public.transactions`)  
**Credential format:** PostgreSQL-compatible wire protocol  
**DuckDB naming:** `public__sessions`, `public__accounts`, `public__transactions`

---

## Credential Setup

CockroachDB is PostgreSQL-wire-compatible. Use the postgres connector type:

```json
{
  "host": "your-cockroachdb-host",
  "port": 26257,
  "database": "defaultdb",
  "username": "root",
  "password": "",
  "ssl_mode": "disable"
}
```

For CockroachDB Cloud (Serverless/Dedicated):
```json
{
  "host": "your-cluster.cockroachlabs.cloud",
  "port": 26257,
  "database": "defaultdb",
  "username": "your-user",
  "password": "your-password",
  "ssl_mode": "verify-full"
}
```

---

## Required Source Tables

```sql
-- public.sessions
CREATE TABLE IF NOT EXISTS public.sessions (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID NOT NULL,
    ip_address  STRING,
    user_agent  STRING,
    started_at  TIMESTAMPTZ DEFAULT NOW(),
    ended_at    TIMESTAMPTZ,
    active      BOOL DEFAULT true
);

-- public.accounts
CREATE TABLE IF NOT EXISTS public.accounts (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email       STRING NOT NULL UNIQUE,
    plan        STRING DEFAULT 'free' CHECK (plan IN ('free','starter','pro','enterprise')),
    balance     DECIMAL(12, 2) DEFAULT 0.00,
    created_at  TIMESTAMPTZ DEFAULT NOW(),
    updated_at  TIMESTAMPTZ DEFAULT NOW()
);

-- public.transactions
CREATE TABLE IF NOT EXISTS public.transactions (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    account_id  UUID REFERENCES public.accounts(id),
    amount      DECIMAL(12, 2) NOT NULL,
    type        STRING CHECK (type IN ('credit','debit')),
    description STRING,
    created_at  TIMESTAMPTZ DEFAULT NOW()
);
```

Seed with at least 20 rows per table.

---

## Stream Reference

| Stream key | DuckDB staging name | Key columns | INCREMENTAL key |
|-----------|-------------------|-------------|----------------|
| `public.sessions` | `public__sessions` | `id`, `user_id`, `ip_address`, `started_at`, `ended_at`, `active` | `started_at` |
| `public.accounts` | `public__accounts` | `id`, `email`, `plan`, `balance`, `created_at`, `updated_at` | `updated_at` |
| `public.transactions` | `public__transactions` | `id`, `account_id`, `amount`, `type`, `description`, `created_at` | `created_at` |

---

## Scenario S-CDB-1 — Full Table Sync: `public.accounts`

**Destination DDL:**
```sql
CREATE TABLE analytics.crdb_accounts_hd (
    id         UUID PRIMARY KEY,
    email      TEXT,
    plan       TEXT,
    balance    NUMERIC(12, 2),
    created_at TIMESTAMPTZ
);
```

**dbt SQL:**
```sql
SELECT
    id,
    email,
    plan,
    balance,
    created_at
FROM {{ source('raw', 'public__accounts') }}
```

**Verify:** `SELECT plan, COUNT(*) FROM analytics.crdb_accounts_hd GROUP BY plan;`

---

## Scenario S-CDB-2 — Incremental Sync: `public.transactions`

**Sync mode:** `INCREMENTAL`, replication key `created_at`

**dbt SQL:**
```sql
SELECT
    id,
    account_id,
    amount,
    type,
    description,
    created_at
FROM {{ source('raw', 'public__transactions') }}
```

**Run 1:** All transactions.  
**Run 2 (add 5 new transactions):** Only the 5 new rows.

---

## Scenario S-CDB-3 — Account Balance Summary

**dbt SQL:**
```sql
SELECT
    plan,
    COUNT(*)           AS account_count,
    SUM(balance)       AS total_balance,
    AVG(balance)       AS avg_balance,
    MAX(balance)       AS max_balance
FROM {{ source('raw', 'public__accounts') }}
GROUP BY plan
ORDER BY total_balance DESC
```

---

## Scenario S-CDB-4 — Transaction Flow: Credits vs Debits

**dbt SQL:**
```sql
SELECT
    account_id,
    SUM(CASE WHEN type = 'credit' THEN amount ELSE 0 END)  AS total_credits,
    SUM(CASE WHEN type = 'debit'  THEN amount ELSE 0 END)  AS total_debits,
    SUM(CASE WHEN type = 'credit' THEN amount ELSE -amount END) AS net_balance
FROM {{ source('raw', 'public__transactions') }}
GROUP BY account_id
```

---

## Scenario S-CDB-5 — Session Activity Analysis

**dbt SQL:**
```sql
SELECT
    id,
    user_id,
    ip_address,
    started_at,
    ended_at,
    active,
    CASE
        WHEN ended_at IS NOT NULL
        THEN DATEDIFF('minute', started_at, ended_at)
        ELSE NULL
    END AS duration_minutes
FROM {{ source('raw', 'public__sessions') }}
ORDER BY started_at DESC
```

---

## Scenario S-CDB-6 — JOIN: accounts + transactions

Both streams in one pipeline:

**dbt SQL:**
```sql
SELECT
    a.id          AS account_id,
    a.email,
    a.plan,
    t.id          AS transaction_id,
    t.amount,
    t.type        AS transaction_type,
    t.created_at  AS transaction_at
FROM {{ source('raw', 'public__accounts') }} AS a
INNER JOIN {{ source('raw', 'public__transactions') }} AS t
    ON a.id = t.account_id
WHERE t.amount > 0
```

---

## Scenario S-CDB-7 — Normalisation: Rename `plan` → `subscription_plan`

**Rule:**
```json
{ "rule_type": "rename", "table": "public.accounts", "column": "plan", "destination_name": "subscription_plan" }
```

**dbt SQL:**
```sql
SELECT id, email, subscription_plan, balance, created_at
FROM {{ source('raw', 'public__accounts') }}
```

---

## Scenario S-CDB-8 — Normalisation: Exclude `description` from transactions

**Rule:**
```json
{ "rule_type": "exclude", "table": "public.transactions", "column": "description" }
```

**dbt SQL:**
```sql
SELECT id, account_id, amount, type, created_at
FROM {{ source('raw', 'public__transactions') }}
```

---

## Scenario S-CDB-9 — SSL Verify-Full (CockroachDB Cloud)

Set `ssl_mode = "verify-full"` in connection config.  
**Expected:** Connection test passes with CA cert chain validated. Phase 1 extracts over TLS.

---

## All 3 Streams — Quick Smoke Test Checklist

| Stream | DuckDB ref | Expected rows |
|--------|-----------|---------------|
| sessions | `public__sessions` | ≥ 1 |
| accounts | `public__accounts` | ≥ 1 |
| transactions | `public__transactions` | ≥ 1 |
