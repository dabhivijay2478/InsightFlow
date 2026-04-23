# Pipeline 26 — PostgreSQL Source → PostgreSQL Destination

**Source streams:** 3 | **Destination:** PostgreSQL

---

## Connections

### Source — PostgreSQL
```json
{ "host":"src-host","port":5432,"database":"mydb","username":"reader","password":"..","ssl_mode":"disable" }
```
```sql
-- Source tables required:
CREATE TABLE public.users (
    id UUID PRIMARY KEY, email TEXT, first_name TEXT, last_name TEXT,
    metadata JSONB, created_at TIMESTAMPTZ, updated_at TIMESTAMPTZ
);
CREATE TABLE public.orders (
    id UUID PRIMARY KEY, user_id UUID, amount NUMERIC(10,2),
    status TEXT, metadata JSONB, placed_at TIMESTAMPTZ
);
CREATE TABLE public.payments (
    id UUID PRIMARY KEY, order_id UUID, method TEXT,
    amount NUMERIC(10,2), status TEXT, paid_at TIMESTAMPTZ
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

## Stream 1 — `public.users` → `analytics.dim_users`

### Step 1 — DDL
```sql
CREATE TABLE analytics.dim_users (
    user_uuid      UUID        PRIMARY KEY,   -- source: id (renamed)
    email_address  TEXT,                       -- source: email (renamed)
    full_name      TEXT,                       -- derived: CONCAT(first_name,' ',last_name)
    city           TEXT,                       -- from JSON: metadata->>'city'
    country        TEXT,                       -- from JSON: metadata->>'country'
    account_tier   TEXT,                       -- from JSON: metadata->>'tier'
    registered_on  DATE,                       -- source: created_at → DATE
    last_active    DATE                        -- source: updated_at → DATE
    -- first_name, last_name, metadata excluded
);
```
### Step 3 — Panel: `public.users` | `INCREMENTAL` | key: `updated_at`
### Step 4 — Normalisation
```json
[
  { "rule_type": "exclude", "table": "public.users", "column": "metadata" },
  { "rule_type": "exclude", "table": "public.users", "column": "first_name" },
  { "rule_type": "exclude", "table": "public.users", "column": "last_name" }
]
```
### Step 5 — dbt SQL
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
### Step 8 — Verify
```sql
SELECT user_uuid, email_address, full_name, city, account_tier FROM analytics.dim_users LIMIT 5;
-- full_name: single string; city/country: plain text; metadata col must NOT exist
SELECT COUNT(*) FROM analytics.dim_users WHERE city IS NULL;  -- allowed, not every user has city
```

---

## Stream 2 — `public.orders` → `analytics.fact_orders`

### Step 1 — DDL
```sql
CREATE TABLE analytics.fact_orders (
    order_id        UUID        PRIMARY KEY,
    customer_ref    UUID,                      -- source: user_id (renamed)
    order_amount    NUMERIC(10,2),
    order_status    TEXT,                      -- source: status (renamed)
    payment_method  TEXT,                      -- from JSON: metadata->>'payment_method'
    channel         TEXT,                      -- from JSON: metadata->>'channel'
    currency        TEXT,                      -- from JSON: metadata->>'currency'
    is_high_value   BOOLEAN,                   -- derived: amount > 100
    placed_on       DATE                       -- source: placed_at → DATE
    -- metadata blob excluded
);
```
### Step 3 — Panel: `public.orders` | `INCREMENTAL` | key: `placed_at`
### Step 5 — dbt SQL
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
### Step 8 — Verify
```sql
SELECT order_id, payment_method, channel, is_high_value FROM analytics.fact_orders LIMIT 5;
-- is_high_value: true/false; metadata column must NOT exist
SELECT COUNT(*) FROM analytics.fact_orders WHERE is_high_value = true;
```

---

## Stream 3 — `public.payments` → `analytics.fact_payments`

### Step 1 — DDL
```sql
CREATE TABLE analytics.fact_payments (
    payment_id    UUID        PRIMARY KEY,
    order_ref     UUID,                       -- source: order_id (renamed)
    pay_method    TEXT,                       -- source: method (renamed)
    pay_amount    NUMERIC(10,2),
    pay_status    TEXT,                       -- source: status (renamed)
    is_successful BOOLEAN,                   -- derived: status = 'paid'
    paid_on       DATE                        -- source: paid_at → DATE
);
```
### Step 3 — Panel: `public.payments` | `INCREMENTAL` | key: `paid_at`
### Step 5 — dbt SQL
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
### Step 8 — Verify
```sql
SELECT pay_method, pay_amount, is_successful FROM analytics.fact_payments LIMIT 5;
-- is_successful: true/false
-- pay_method uses renamed column (not 'method')
```

---

## Edge Cases

| Scenario | Expected |
|---------|---------|
| `metadata` is NULL | WHERE filters row; excluded from output |
| `metadata->>'city'` missing key | city = NULL |
| `first_name` NULL | COALESCE → '' → full_name trim to last_name only |
| INCREMENTAL: `updated_at` not indexed on source | Slow sync; add index to source |
| Same source and destination host | Pipeline still works; schema isolation prevents collisions |
