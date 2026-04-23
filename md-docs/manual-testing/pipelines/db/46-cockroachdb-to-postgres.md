# Pipeline 46 — CockroachDB Source → PostgreSQL Destination

**Source streams:** 3 | **Destination:** PostgreSQL

---

## Connections

### Source — CockroachDB
```json
{ "host":"src-host","port":26257,"database":"defaultdb","username":"root","password":"","ssl_mode":"disable" }
```
```sql
-- Source tables:
CREATE TABLE public.accounts (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    email STRING, plan STRING,
    balance DECIMAL(12,2), created_at TIMESTAMPTZ, updated_at TIMESTAMPTZ
);
CREATE TABLE public.transactions (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    account_id UUID, amount DECIMAL(12,2),
    type STRING, description STRING, created_at TIMESTAMPTZ
);
CREATE TABLE public.sessions (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    user_id UUID, ip_address STRING, user_agent STRING,
    active BOOL, started_at TIMESTAMPTZ, ended_at TIMESTAMPTZ
);
```

### Destination — PostgreSQL
```json
{ "host":"dest-host","port":5432,"database":"analytics","username":"writer","password":"..","ssl_mode":"disable" }
```
```sql
CREATE SCHEMA IF NOT EXISTS analytics;
```

---

## Stream 1 — `public.accounts` → `analytics.account_ledger`

### Step 1 — DDL
```sql
CREATE TABLE analytics.account_ledger (
    account_id   UUID        PRIMARY KEY,
    email        TEXT,
    plan_tier    TEXT,                        -- source: plan (renamed)
    balance_usd  NUMERIC(12,2),               -- source: balance (renamed)
    tier_rank    INTEGER,                     -- derived: CASE plan → 1-4
    is_paid      BOOLEAN,                     -- derived: plan != 'free'
    joined_on    DATE,                        -- source: created_at → DATE
    updated_on   DATE
);
```
### Step 3 — Panel: `public.accounts` | `INCREMENTAL` | key: `updated_at`
### Step 5 — dbt SQL
```sql
SELECT
    id                              AS account_id,
    email,
    plan                            AS plan_tier,
    balance                         AS balance_usd,
    CASE plan
        WHEN 'enterprise' THEN 4
        WHEN 'pro'        THEN 3
        WHEN 'starter'    THEN 2
        ELSE                   1
    END                             AS tier_rank,
    plan != 'free'                  AS is_paid,
    created_at::DATE                AS joined_on,
    updated_at::DATE                AS updated_on
FROM {{ source('raw', 'public__accounts') }}
```
### Step 8 — Verify
```sql
SELECT plan_tier, tier_rank, is_paid FROM analytics.account_ledger GROUP BY plan_tier, tier_rank, is_paid;
-- enterprise→4→true, pro→3→true, starter→2→true, free→1→false
SELECT account_id FROM analytics.account_ledger LIMIT 3;  -- UUID
```

---

## Stream 2 — `public.transactions` → `analytics.txn_facts`

### Step 1 — DDL
```sql
CREATE TABLE analytics.txn_facts (
    txn_id      UUID        PRIMARY KEY,
    account_ref UUID,                         -- source: account_id (renamed)
    txn_amount  NUMERIC(12,2),                -- source: amount (renamed)
    txn_type    TEXT,                         -- source: type (renamed)
    is_credit   BOOLEAN,                      -- derived: type = 'credit'
    note        TEXT,                         -- source: description (renamed)
    txn_date    DATE,
    txn_month   TEXT                          -- derived: DATE_TRUNC month label
);
```
### Step 5 — dbt SQL
```sql
SELECT
    id                                  AS txn_id,
    account_id                          AS account_ref,
    amount                              AS txn_amount,
    type                                AS txn_type,
    type = 'credit'                     AS is_credit,
    description                         AS note,
    created_at::DATE                    AS txn_date,
    DATE_TRUNC('month', created_at)::TEXT AS txn_month
FROM {{ source('raw', 'public__transactions') }}
WHERE amount > 0
```
### Step 8 — Verify
```sql
SELECT txn_type, is_credit FROM analytics.txn_facts GROUP BY txn_type, is_credit;
-- credit→true, debit→false
SELECT txn_month FROM analytics.txn_facts GROUP BY txn_month ORDER BY txn_month;
-- e.g. '2024-01-01 00:00:00+00' or '2024-01-01'
```

---

## Stream 3 — `public.sessions` → `analytics.session_registry`

### Step 1 — DDL
```sql
CREATE TABLE analytics.session_registry (
    session_id    UUID        PRIMARY KEY,
    user_ref      UUID,                       -- source: user_id (renamed)
    ip_address    TEXT,
    browser_info  TEXT,                       -- source: user_agent (renamed)
    is_active     BOOLEAN,
    duration_mins INTEGER,                    -- derived: DATEDIFF minutes
    started_on    DATE,
    ended_on      DATE
);
```
### Step 5 — dbt SQL
```sql
SELECT
    id                                                    AS session_id,
    user_id                                               AS user_ref,
    ip_address,
    user_agent                                            AS browser_info,
    active                                                AS is_active,
    CASE WHEN ended_at IS NOT NULL
         THEN DATEDIFF('minute', started_at, ended_at)
         ELSE NULL END                                    AS duration_mins,
    started_at::DATE                                      AS started_on,
    ended_at::DATE                                        AS ended_on
FROM {{ source('raw', 'public__sessions') }}
```

---

## Edge Cases

| Scenario | Expected |
|---------|---------|
| `plan = NULL` | tier_rank ELSE → 1; is_paid `NULL != 'free'` → NULL — use `COALESCE(plan,'free')` |
| `ended_at IS NULL` (active session) | duration_mins = NULL; ended_on = NULL |
| `amount = 0` | Filtered by WHERE `amount > 0` |
| CockroachDB source SSL required | Add `"ssl_mode": "verify-full"` and cert paths |
