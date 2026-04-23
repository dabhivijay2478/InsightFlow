# PostgreSQL Source — UI Testing (3 Streams × 5 Destinations)

> Universal builder steps in `builder-walkthrough.md`.

---

## Phase 1 — Source Panel (PostgreSQL)

### Credential Fields
| Field | Value |
|-------|-------|
| **Host** | `src-host` |
| **Port** | `5432` |
| **Database** | `mydb` |
| **Username** | `reader` |
| **Password** | `***` |
| **SSL Mode** | `disable` / `require` / `verify-full` |

**Test Connection → ✅**
- ❌ `connection refused`: wrong host/port
- ❌ `password authentication failed`: wrong credentials
- ❌ `SSL SYSCALL error`: SSL mode mismatch

### SSL Mode UI Tests
| SSL Mode selected | Server SSL config | Expected |
|------------------|------------------|---------|
| `disable` | SSL off | ✅ |
| `require` | SSL on (any cert) | ✅ |
| `verify-full` | Self-signed cert | ❌ cert verification fails |
| `disable` | SSL required server | ❌ connection rejected |

---

## Step 2b — Stream Selection & Sync Mode

Source tables must exist before running. Create:
```sql
CREATE TABLE public.users (id UUID PRIMARY KEY, email TEXT, first_name TEXT, last_name TEXT, metadata JSONB, created_at TIMESTAMPTZ, updated_at TIMESTAMPTZ);
CREATE TABLE public.orders (id UUID PRIMARY KEY, user_id UUID, amount NUMERIC(10,2), status TEXT, metadata JSONB, placed_at TIMESTAMPTZ);
CREATE TABLE public.payments (id UUID PRIMARY KEY, order_id UUID, method TEXT, amount NUMERIC(10,2), status TEXT, paid_at TIMESTAMPTZ);
```

| Stream (table) | Sync Mode | Cursor Field |
|----------------|-----------|-------------|
| `public.users` | INCREMENTAL | `updated_at` |
| `public.orders` | INCREMENTAL | `placed_at` |
| `public.payments` | INCREMENTAL | `paid_at` |

**Cursor field selection test**: Verify `updated_at`, `placed_at`, `paid_at` appear in dropdown (TIMESTAMPTZ columns should be available).

---

## Phase 2 — Stream→Table Mapping (per destination)

| Stream | Table |
|--------|-------|
| `public.users` | `dim_users` |
| `public.orders` | `fact_orders` |
| `public.payments` | `fact_payments` |

### Per-destination schema
| Destination | Schema / Catalog |
|-------------|----------------|
| PostgreSQL | `analytics` |
| MySQL | `analytics` |
| MariaDB | `analytics` |
| SQLite | `main` |
| CockroachDB | `analytics` |

---

## Phase 3 — Normalisation Rules

### `public.users`
| Rule | Column | Target |
|------|--------|--------|
| Rename | `id` | `user_uuid` |
| Rename | `email` | `email_address` |
| Exclude | `metadata` | — (JSON keys extracted in dbt) |
| Exclude | `first_name` | — (merged in dbt) |
| Exclude | `last_name` | — (merged in dbt) |

### `public.orders`
| Rule | Column | Target |
|------|--------|--------|
| Rename | `id` | `order_id` |
| Rename | `user_id` | `customer_ref` |
| Rename | `status` | `order_status` |
| Exclude | `metadata` | — |

### `public.payments`
| Rule | Column | Target |
|------|--------|--------|
| Rename | `id` | `payment_id` |
| Rename | `order_id` | `order_ref` |
| Rename | `method` | `pay_method` |
| Rename | `amount` | `pay_amount` |
| Rename | `status` | `pay_status` |

---

## Phase 4 — dbt SQL (stream-by-stream)

### Stream 1 — `public.users` → `dim_users`
```sql
SELECT
    id                                                            AS user_uuid,
    email                                                         AS email_address,
    TRIM(CONCAT(COALESCE(first_name,''),' ',COALESCE(last_name,''))) AS full_name,
    metadata->>'city'                                             AS city,
    metadata->>'country'                                          AS country,
    metadata->>'tier'                                             AS account_tier,
    created_at::DATE                                              AS registered_on,
    updated_at::DATE                                              AS last_active
FROM {{ source('raw', 'public__users') }}
WHERE email IS NOT NULL
```
**Validate → expected columns**: `user_uuid`, `email_address`, `full_name`, `city`, `country`, `account_tier`, `registered_on`, `last_active`

**Preview checks**:
- `full_name`: `"Jane Smith"` — NOT `first_name` + `last_name` separately
- `city`: plain string or NULL — NOT `{"city":"London"}`
- `metadata` column absent

---

### Stream 2 — `public.orders` → `fact_orders`
```sql
SELECT
    id                               AS order_id,
    user_id                          AS customer_ref,
    amount                           AS order_amount,
    status                           AS order_status,
    metadata->>'payment_method'      AS payment_method,
    metadata->>'channel'             AS channel,
    metadata->>'currency'            AS currency,
    amount > 100                     AS is_high_value,
    placed_at::DATE                  AS placed_on
FROM {{ source('raw', 'public__orders') }}
WHERE metadata IS NOT NULL
```
**Preview checks**:
- `payment_method`: `"card"` / `"bank_transfer"` (plain string)
- `is_high_value`: `true`/`false`
- `metadata` column absent

---

### Stream 3 — `public.payments` → `fact_payments`
```sql
SELECT
    id              AS payment_id,
    order_id        AS order_ref,
    method          AS pay_method,
    amount          AS pay_amount,
    status          AS pay_status,
    status = 'paid' AS is_successful,
    paid_at::DATE   AS paid_on
FROM {{ source('raw', 'public__payments') }}
```
**Preview checks**:
- `is_successful`: `true`/`false`
- `pay_method`: `"card"` (renamed from `method`)

---

## Phase 5 — Preview Checks (all 5 destinations)

The preview uses DuckDB staging regardless of destination — same columns expected:

| Column | Expected |
|--------|---------|
| `user_uuid` | UUID string `550e8400-...` |
| `full_name` | Combined `"Jane Smith"` |
| `city` | Plain string or NULL |
| `is_high_value` | `true`/`false` |
| `is_successful` | `true`/`false` |

---

## Phase 6 — Destination-Specific Type Check

After delivery, verify destination-specific storage:

| Destination | `user_uuid` type | `is_high_value` type |
|-------------|-----------------|---------------------|
| PostgreSQL | `UUID` | `BOOLEAN` (true/false) |
| MySQL | `VARCHAR(36)` | `TINYINT(1)` (0/1) |
| MariaDB | `VARCHAR(36)` | `TINYINT(1)` (0/1) |
| SQLite | `TEXT` | `INTEGER` (0/1) |
| CockroachDB | `UUID` | `BOOL` (true/false) |

---

## Phase 7 — Failure Scenarios

| Scenario | Expected |
|---------|---------|
| `metadata` NULL | WHERE filters row; not in `dim_users` |
| `metadata->>'city'` key missing | `city` = NULL — ✅ |
| INCREMENTAL cursor `updated_at` not indexed | Slow sync; add index to source |
| Source and dest on same host | Works; schema isolation prevents collision |
| Reader user lacks SELECT permission | Phase 1 ❌ `permission denied for table users` |

---

## Phase 8 — Verify (PostgreSQL dest)

```sql
SELECT user_uuid, email_address, full_name, city FROM analytics.dim_users LIMIT 5;
SELECT user_uuid, COUNT(*) FROM analytics.dim_users GROUP BY user_uuid HAVING COUNT(*)>1;  -- 0
SELECT payment_method, channel FROM analytics.fact_orders LIMIT 5;
SELECT is_successful, COUNT(*) FROM analytics.fact_payments GROUP BY is_successful;
-- true: N, false: M
```

---

## Incremental Re-run Test

1. Run pipeline — note row counts
2. Insert 1 new row to `public.orders` with `placed_at = NOW()`
3. Re-run pipeline
4. ✅ `fact_orders` count = previous + 1
5. ✅ `dim_users` count unchanged (no new users)
