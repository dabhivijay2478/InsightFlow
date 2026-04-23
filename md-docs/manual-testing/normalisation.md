# Normalisation Rules — Manual Testing Guide

Normalisation rules are applied **before** dbt SQL executes. They modify the data in DuckDB staging so your SQL model can reference the renamed/cast columns.

---

## Rule Types

| Rule type | What it does | Applied at |
|-----------|-------------|-----------|
| `rename` | Renames a column in the DuckDB staging table | Phase 1 → Phase 2 |
| `cast` | Changes the data type of a column | Phase 1 → Phase 2 |
| `exclude` | Drops a column from the staging table | Phase 1 → Phase 2 |

All rules are set per-table (you specify `table = "stream_key"` and `column`).

---

## Where to Set Rules in the UI

1. Open pipeline builder.
2. Go to **Normalisation** panel (tab between Destination and Preview).
3. For each source stream, click **+ Add Rule**.
4. Fill in: rule type, source table (stream key), column name, and the destination name / cast type.
5. **Save** the pipeline — rules are persisted in `destination_schemas.normalisation_rules`.

---

## Scenario NORM-1 — Rename Column

**Use case:** Source has `created_at`; your destination table calls it `order_date`.

**Rule configuration:**
```json
{
  "rule_type": "rename",
  "table": "public.orders",
  "column": "created_at",
  "destination_name": "order_date"
}
```

**dbt SQL after rename (reference the NEW name):**
```sql
SELECT
    id,
    amount,
    status,
    order_date       -- renamed from created_at
FROM {{ source('raw', 'public__orders') }}
```

**Verify:**
1. Run pipeline.
2. In destination DB: `SELECT order_date FROM analytics.orders_hd LIMIT 5;` — column exists and contains timestamps.
3. Confirm `created_at` column name does NOT appear in destination.

---

## Scenario NORM-2 — Rename Multiple Columns (snake_case → camelCase)

**Rule configuration:**
```json
[
  { "rule_type": "rename", "table": "public.users", "column": "created_at",  "destination_name": "createdAt" },
  { "rule_type": "rename", "table": "public.users", "column": "updated_at",  "destination_name": "updatedAt" },
  { "rule_type": "rename", "table": "public.users", "column": "first_name",  "destination_name": "firstName" },
  { "rule_type": "rename", "table": "public.users", "column": "last_name",   "destination_name": "lastName" }
]
```

**dbt SQL:**
```sql
SELECT
    id,
    email,
    firstName,
    lastName,
    createdAt,
    updatedAt
FROM {{ source('raw', 'public__users') }}
```

**Destination DDL:**
```sql
CREATE TABLE analytics.users_camel (
    id         UUID PRIMARY KEY,
    email      TEXT,
    "firstName" TEXT,
    "lastName"  TEXT,
    "createdAt" TIMESTAMPTZ,
    "updatedAt" TIMESTAMPTZ
);
```

> ⚠️ If destination DB is PostgreSQL, camelCase column names need double-quotes in DDL.

---

## Scenario NORM-3 — Rename for SaaS Stream

**Use case:** Stripe `charges.amount` (integer cents) rename to `amount_cents`.

**Rule:**
```json
{
  "rule_type": "rename",
  "table": "stripe.charges",
  "column": "amount",
  "destination_name": "amount_cents"
}
```

**dbt SQL:**
```sql
SELECT
    id,
    amount_cents,
    currency,
    status,
    customer
FROM {{ source('raw', 'stripe__charges') }}
```

**Verify:** `SELECT amount_cents FROM analytics.charges_hd LIMIT 5;` — values are integer cent amounts (e.g., 2999 = $29.99).

---

## Scenario NORM-4 — CAST text → integer

**Use case:** Source stores a numeric value as `TEXT`; destination column is `INTEGER`.

**Rule:**
```json
{
  "rule_type": "cast",
  "table": "public.orders",
  "column": "status_code",
  "cast_to": "integer"
}
```

**dbt SQL:**
```sql
SELECT
    id,
    status_code,    -- now integer after cast
    amount,
    placed_at
FROM {{ source('raw', 'public__orders') }}
WHERE status_code IS NOT NULL
```

**Destination DDL:**
```sql
CREATE TABLE analytics.orders_typed (
    id          UUID PRIMARY KEY,
    status_code INTEGER,
    amount      NUMERIC,
    placed_at   TIMESTAMPTZ
);
```

**Verify:**
```sql
SELECT pg_typeof(status_code) FROM analytics.orders_typed LIMIT 1;
-- Returns: integer
```

---

## Scenario NORM-5 — CAST text → decimal/numeric

**Rule:**
```json
{
  "rule_type": "cast",
  "table": "public.orders",
  "column": "amount",
  "cast_to": "decimal"
}
```

**Verify:** `SELECT amount, pg_typeof(amount) FROM analytics.orders_hd LIMIT 1;` — type is `numeric`.

---

## Scenario NORM-6 — CAST timestamp text → timestamptz

**Rule:**
```json
{
  "rule_type": "cast",
  "table": "app.events",
  "column": "occurred_at",
  "cast_to": "timestamptz"
}
```

**Verify:** `SELECT occurred_at FROM analytics.events_hd LIMIT 1;` — value is a proper timestamp, not a plain text string.

---

## Scenario NORM-7 — Exclude Column

**Use case:** Source has a `password_hash` column; you must not deliver it to the destination.

**Rule:**
```json
{
  "rule_type": "exclude",
  "table": "public.users",
  "column": "password_hash"
}
```

**dbt SQL (do NOT reference excluded column):**
```sql
SELECT
    id,
    email,
    created_at
FROM {{ source('raw', 'public__users') }}
```

**Verify:**
1. Run pipeline.
2. `SELECT column_name FROM information_schema.columns WHERE table_name = 'users_hd';`
3. `password_hash` must NOT appear in the result.

---

## Scenario NORM-8 — Exclude Multiple Sensitive Columns

**Rules:**
```json
[
  { "rule_type": "exclude", "table": "stripe.customers", "column": "default_source" },
  { "rule_type": "exclude", "table": "stripe.customers", "column": "sources" },
  { "rule_type": "exclude", "table": "stripe.customers", "column": "subscriptions" }
]
```

**dbt SQL (all excluded columns omitted from SELECT):**
```sql
SELECT
    id,
    email,
    name,
    currency,
    created
FROM {{ source('raw', 'stripe__customers') }}
```

**Verify:** Only the 5 listed columns appear in destination.

---

## Scenario NORM-9 — Combined Rename + Cast

**Rules (applied together):**
```json
[
  { "rule_type": "rename", "table": "public.orders", "column": "placed_at",    "destination_name": "order_date" },
  { "rule_type": "cast",   "table": "public.orders", "column": "amount",       "cast_to": "decimal" }
]
```

**dbt SQL:**
```sql
SELECT
    id,
    order_date,       -- renamed from placed_at
    amount,           -- cast to decimal
    status
FROM {{ source('raw', 'public__orders') }}
```

**Verify:** `order_date` contains timestamps; `amount` type is numeric.

---

## Scenario NORM-10 — Combined Rename + Exclude + dbt SQL

**Rules:**
```json
[
  { "rule_type": "rename",  "table": "stripe.customers", "column": "created",   "destination_name": "created_ts" },
  { "rule_type": "exclude", "table": "stripe.customers", "column": "sources" },
  { "rule_type": "exclude", "table": "stripe.customers", "column": "subscriptions" }
]
```

**dbt SQL:**
```sql
SELECT
    id,
    email,
    name,
    currency,
    created_ts          -- renamed from created
FROM {{ source('raw', 'stripe__customers') }}
WHERE currency = 'usd'
```

**Verify:** Only USD customers; `sources` and `subscriptions` absent; `created_ts` is populated.

---

## Round-Trip Verification Checklist

After setting normalisation rules and running:

- [ ] Open pipeline builder → Normalisation panel → confirm rules are saved (not empty).
- [ ] In dbt SQL editor, reference renamed columns by their **new** name.
- [ ] Click **Validate SQL** — no column-not-found errors.
- [ ] After run: `SELECT column_name FROM information_schema.columns WHERE table_name = '<dest_table>';` — excluded columns absent, renamed columns present.
- [ ] Check Run Status drawer: no `delivery_failures` entry.

---

## Error Cases

| Mistake | Symptom | Fix |
|---------|---------|-----|
| Reference old name in SQL after rename | dbt compile error: column not found | Use new name in SELECT |
| Cast to incompatible type (e.g., text `"abc"` → integer) | Phase 2 / dbt error: invalid cast | Wrap with `CASE WHEN … ELSE NULL END` |
| Exclude PK column then use merge mode | Phase 3 delivery error: PK not in model output | Don't exclude PK, or switch to append mode |
| Rename to a name that already exists in source | Column name collision in DuckDB staging | Choose a different destination name |
