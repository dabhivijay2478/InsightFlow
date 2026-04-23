# PostgreSQL Source — Manual Testing Guide

**Streams:** 3 (`public.users`, `public.orders`, `public.payments`)  
**Credential format:** host + port + database + username + password  
**DuckDB naming:** `{schema}__{table}` (e.g., `public__users`)

---

## Credential Setup

```json
{
  "host": "your-postgres-host",
  "port": 5432,
  "database": "your_database",
  "username": "your_user",
  "password": "your_password",
  "ssl_mode": "disable"
}
```

`ssl_mode` options: `disable`, `require`, `verify-full`

---

## Required Source Tables

Create these in the source PostgreSQL database before running any pipeline:

```sql
-- public.users
CREATE TABLE IF NOT EXISTS public.users (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email       TEXT NOT NULL UNIQUE,
    first_name  TEXT,
    last_name   TEXT,
    created_at  TIMESTAMPTZ DEFAULT NOW(),
    updated_at  TIMESTAMPTZ DEFAULT NOW()
);

-- public.orders
CREATE TABLE IF NOT EXISTS public.orders (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID REFERENCES public.users(id),
    amount      NUMERIC(10, 2) NOT NULL,
    status      TEXT NOT NULL DEFAULT 'pending',
    metadata    JSONB,
    placed_at   TIMESTAMPTZ DEFAULT NOW()
);

-- public.payments
CREATE TABLE IF NOT EXISTS public.payments (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id    UUID REFERENCES public.orders(id),
    method      TEXT NOT NULL,
    amount      NUMERIC(10, 2) NOT NULL,
    status      TEXT NOT NULL,
    paid_at     TIMESTAMPTZ DEFAULT NOW()
);
```

Seed with at least 10 rows per table for meaningful tests.

---

## Stream Reference

| Stream key | DuckDB staging name | Columns | INCREMENTAL key |
|-----------|-------------------|---------|----------------|
| `public.users` | `public__users` | `id`, `email`, `first_name`, `last_name`, `created_at`, `updated_at` | `updated_at` |
| `public.orders` | `public__orders` | `id`, `user_id`, `amount`, `status`, `metadata`, `placed_at` | `placed_at` |
| `public.payments` | `public__payments` | `id`, `order_id`, `method`, `amount`, `status`, `paid_at` | `paid_at` |

---

## Scenario S-PG-1 — Full Table Sync: `public.users`

**Destination DDL:**
```sql
CREATE TABLE analytics.pg_users_hd (
    id         UUID PRIMARY KEY,
    email      TEXT,
    first_name TEXT,
    last_name  TEXT,
    created_at TIMESTAMPTZ
);
```

**dbt SQL:**
```sql
SELECT
    id,
    email,
    first_name,
    last_name,
    created_at
FROM {{ source('raw', 'public__users') }}
```

**Verify:**
```sql
SELECT COUNT(*) FROM analytics.pg_users_hd;
-- Must equal row count in public.users
```

---

## Scenario S-PG-2 — Incremental Sync: `public.orders`

**Sync mode:** `INCREMENTAL`, replication key `placed_at`

**dbt SQL:**
```sql
SELECT
    id,
    user_id,
    amount,
    status,
    placed_at
FROM {{ source('raw', 'public__orders') }}
WHERE status IN ('paid', 'shipped', 'delivered')
```

**Run 1:** All matching orders.  
**Run 2 (add new order in source):** Only the new order.

---

## Scenario S-PG-3 — JSONB Key Filtering: `orders.metadata`

`orders.metadata` stores 10 keys. Deliver only `payment_method`, `currency`, `channel`:

```sql
SELECT
    id,
    amount,
    status,
    metadata->>'payment_method'  AS payment_method,
    metadata->>'currency'        AS currency,
    metadata->>'channel'         AS channel,
    placed_at
FROM {{ source('raw', 'public__orders') }}
WHERE metadata IS NOT NULL
  AND metadata->>'currency' IS NOT NULL
```

**Destination DDL:**
```sql
CREATE TABLE analytics.orders_meta_flat (
    id             UUID PRIMARY KEY,
    amount         NUMERIC,
    status         TEXT,
    payment_method TEXT,
    currency       TEXT,
    channel        TEXT,
    placed_at      TIMESTAMPTZ
);
```

**Verify:**
```sql
SELECT payment_method, currency, channel FROM analytics.orders_meta_flat LIMIT 5;
-- metadata blob must NOT appear as a column
```

---

## Scenario S-PG-4 — JOIN: `orders` + `payments`

Both `public.orders` AND `public.payments` selected in Source Panel.

**dbt SQL (join in one model):**
```sql
SELECT
    o.id          AS order_id,
    o.amount      AS order_amount,
    o.status      AS order_status,
    p.id          AS payment_id,
    p.method      AS payment_method,
    p.status      AS payment_status,
    o.placed_at
FROM {{ source('raw', 'public__orders') }} AS o
LEFT JOIN {{ source('raw', 'public__payments') }} AS p
    ON o.id = p.order_id
```

**Destination DDL:**
```sql
CREATE TABLE analytics.orders_with_payments (
    order_id        UUID PRIMARY KEY,
    order_amount    NUMERIC,
    order_status    TEXT,
    payment_id      UUID,
    payment_method  TEXT,
    payment_status  TEXT,
    placed_at       TIMESTAMPTZ
);
```

---

## Scenario S-PG-5 — Normalisation: Rename `first_name`/`last_name` + Exclude `updated_at`

**Rules:**
```json
[
  { "rule_type": "rename",  "table": "public.users", "column": "first_name", "destination_name": "fname" },
  { "rule_type": "rename",  "table": "public.users", "column": "last_name",  "destination_name": "lname" },
  { "rule_type": "exclude", "table": "public.users", "column": "updated_at" }
]
```

**dbt SQL:**
```sql
SELECT id, email, fname, lname, created_at
FROM {{ source('raw', 'public__users') }}
```

---

## Scenario S-PG-6 — Cast `amount` text → decimal (if stored as text)

**Rule:**
```json
{ "rule_type": "cast", "table": "public.orders", "column": "amount", "cast_to": "decimal" }
```

---

## Scenario S-PG-7 — Revenue Summary Aggregate

**dbt SQL:**
```sql
SELECT
    status,
    COUNT(*)       AS order_count,
    SUM(amount)    AS total_revenue,
    AVG(amount)    AS avg_order_value,
    MIN(placed_at) AS first_order_at,
    MAX(placed_at) AS last_order_at
FROM {{ source('raw', 'public__orders') }}
GROUP BY status
```

---

## Scenario S-PG-8 — Window Function: Latest Payment per Order

**dbt SQL:**
```sql
SELECT
    order_id,
    method,
    amount,
    status,
    paid_at,
    ROW_NUMBER() OVER (PARTITION BY order_id ORDER BY paid_at DESC) AS rn
FROM {{ source('raw', 'public__payments') }}
```

Then wrap in a CTE to keep only `rn = 1`:
```sql
WITH ranked AS (
    SELECT
        order_id,
        method,
        amount,
        status,
        paid_at,
        ROW_NUMBER() OVER (PARTITION BY order_id ORDER BY paid_at DESC) AS rn
    FROM {{ source('raw', 'public__payments') }}
)
SELECT order_id, method, amount, status, paid_at
FROM ranked
WHERE rn = 1
```

---

## Scenario S-PG-9 — SSL Connection Test

Set `ssl_mode = "require"` in the source connection.  
**Expected:** Connection test passes; Phase 1 extraction uses TLS.

Set `ssl_mode = "verify-full"` with a bad CA cert.  
**Expected:** Connection test fails with SSL verification error.

---

## Scenario S-PG-10 — Error: Incremental Key Column Missing

Set sync mode `INCREMENTAL`, replication key `nonexistent_column`.  
**Expected:** Phase 0 fails: `"cursor column nonexistent_column not found in public.orders"`.
