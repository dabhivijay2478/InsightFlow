# CockroachDB Source — UI Testing (3 Streams × 5 Destinations)

> Universal builder steps in `builder-walkthrough.md`.

---

## Phase 1 — Source Panel (CockroachDB)

### Credential Fields
| Field | Value |
|-------|-------|
| **Host** | `src-host` |
| **Port** | `26257` |
| **Database** | `defaultdb` |
| **Username** | `root` |
| **Password** | `***` (empty for local dev) |
| **SSL Mode** | `disable` (local) / `verify-full` (CockroachDB Cloud) |

**Test Connection → ✅**
- ❌ `connection refused`: wrong port (CockroachDB uses 26257, not 5432)
- ❌ `SSL handshake error`: cert not provided for `verify-full` mode
- ❌ `user root has no password`: set password or use cert auth

### SSL Test Cases
| SSL Mode | Cert provided | Expected |
|----------|--------------|---------|
| `disable` | N/A | ✅ local dev |
| `require` | No cert | ✅ (no cert verification) |
| `verify-full` | CA cert pasted | ✅ Cloud connection |
| `verify-full` | No cert | ❌ `certificate verification failed` |

---

## Step 2b — Stream Selection & Sync Mode

Source tables (pre-create):
```sql
USE defaultdb;
CREATE SCHEMA IF NOT EXISTS public;
CREATE TABLE public.accounts (id UUID DEFAULT gen_random_uuid() PRIMARY KEY, email STRING, plan STRING, balance DECIMAL(12,2), created_at TIMESTAMPTZ, updated_at TIMESTAMPTZ);
CREATE TABLE public.transactions (id UUID DEFAULT gen_random_uuid() PRIMARY KEY, account_id UUID, amount DECIMAL(12,2), type STRING, description STRING, created_at TIMESTAMPTZ);
CREATE TABLE public.sessions (id UUID DEFAULT gen_random_uuid() PRIMARY KEY, user_id UUID, ip_address STRING, user_agent STRING, active BOOL, started_at TIMESTAMPTZ, ended_at TIMESTAMPTZ);
```

| Stream | Sync Mode | Cursor Field |
|--------|-----------|-------------|
| `public.accounts` | INCREMENTAL | `updated_at` |
| `public.transactions` | INCREMENTAL | `created_at` |
| `public.sessions` | INCREMENTAL | `started_at` |

---

## Phase 2 — Stream→Table Mapping

| Stream | Table |
|--------|-------|
| `public.accounts` | `account_ledger` |
| `public.transactions` | `txn_facts` |
| `public.sessions` | `session_registry` |

---

## Phase 3 — Normalisation Rules

### `public.accounts`
| Rule | Column | Target |
|------|--------|--------|
| Rename | `id` | `account_id` |
| Rename | `plan` | `plan_tier` |
| Rename | `balance` | `balance_usd` |

### `public.transactions`
| Rule | Column | Target |
|------|--------|--------|
| Rename | `id` | `txn_id` |
| Rename | `account_id` | `account_ref` |
| Rename | `amount` | `txn_amount` |
| Rename | `type` | `txn_type` |
| Rename | `description` | `note` |

### `public.sessions`
| Rule | Column | Target |
|------|--------|--------|
| Rename | `id` | `session_id` |
| Rename | `user_id` | `user_ref` |
| Rename | `user_agent` | `browser_info` |
| Rename | `active` | `is_active` |

---

## Phase 4 — dbt SQL

### Stream 1 — `public.accounts` → `account_ledger`
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
**Preview check**:
- `plan_tier`: `"enterprise"` / `"pro"` etc.
- `tier_rank`: integer 1–4
- `is_paid`: `true`/`false`
- `plan` column absent (renamed)

---

### Stream 2 — `public.transactions` → `txn_facts`
```sql
SELECT
    id                                    AS txn_id,
    account_id                            AS account_ref,
    amount                                AS txn_amount,
    type                                  AS txn_type,
    type = 'credit'                       AS is_credit,
    description                           AS note,
    created_at::DATE                      AS txn_date,
    DATE_TRUNC('month', created_at)::TEXT AS txn_month
FROM {{ source('raw', 'public__transactions') }}
WHERE amount > 0
```
**Preview check**: `is_credit` boolean; `txn_month` e.g. `'2024-01-01 00:00:00+00'`

---

### Stream 3 — `public.sessions` → `session_registry`
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
**Preview check**: `duration_mins` integer or NULL; `is_active` boolean

---

## Phase 5 — Preview Checks

| Column | Expected |
|--------|---------|
| `tier_rank` | Integer 1–4 |
| `is_paid` | `true`/`false` |
| `is_credit` | `true`/`false` |
| `txn_month` | ISO date string |
| `duration_mins` | Integer or NULL |

---

## Phase 5b — Destination Type Check

| Destination | `account_id` | `is_paid` | `balance_usd` |
|-------------|-------------|----------|--------------|
| PostgreSQL | `UUID` | `BOOLEAN` | `NUMERIC` |
| MySQL | `VARCHAR(36)` | `TINYINT(1)` | `DECIMAL` |
| MariaDB | `VARCHAR(36)` | `TINYINT(1)` | `DECIMAL` |
| SQLite | `TEXT` | `INTEGER` | `REAL` |
| CockroachDB | `UUID` | `BOOL` | `DECIMAL` |

---

## Phase 6 — Schedule Tests

| Test | Config | Expected |
|------|--------|---------|
| Every 15 min | Cron `*/15 * * * *` | Runs 4×/hour |
| Daily 03:00 UTC | Cron `0 3 * * *` | Nightly sync |
| None | None | Manual only |

---

## Phase 7 — Failure Scenarios

| Scenario | Expected |
|---------|---------|
| `plan = NULL` | `tier_rank` ELSE → 1; `is_paid` = NULL (fix with COALESCE) |
| `ended_at IS NULL` (active session) | `duration_mins` = NULL; `ended_on` = NULL |
| `amount = 0` | WHERE filters row |
| CockroachDB source SSL required cloud | Add `"ssl_mode": "verify-full"` + CA cert path |
| Same cluster source+dest | Works with different schemas |

---

## Phase 8 — Verify

```sql
-- PostgreSQL destination
SELECT plan_tier, tier_rank, is_paid FROM analytics.account_ledger GROUP BY plan_tier, tier_rank, is_paid;
-- enterprise→4→true, pro→3→true, starter→2→true, free→1→false

SELECT account_id FROM analytics.account_ledger LIMIT 3;   -- UUID type
SELECT is_credit FROM analytics.txn_facts GROUP BY is_credit; -- true/false
SELECT txn_month FROM analytics.txn_facts GROUP BY txn_month ORDER BY txn_month;
SELECT txn_id, COUNT(*) FROM analytics.txn_facts GROUP BY txn_id HAVING COUNT(*)>1; -- 0
```

---

## Incremental Re-run Test

1. Run pipeline — note `txn_facts` row count = N
2. Insert new transaction: `INSERT INTO public.transactions VALUES (gen_random_uuid(), ..., NOW());`
3. Re-run pipeline
4. ✅ `txn_facts` count = N+1
5. ✅ `account_ledger` unchanged (no new accounts)
