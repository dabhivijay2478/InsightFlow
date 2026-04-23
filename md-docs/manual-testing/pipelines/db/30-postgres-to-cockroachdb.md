# Pipeline 30 — PostgreSQL Source → CockroachDB Destination

**Source streams:** 3 | **Destination:** CockroachDB

> dbt SQL identical to `26-postgres-to-postgres.md`. Types nearly identical to PostgreSQL.

---

## Connections
```json
{ "host":"src-host","port":5432,"database":"mydb","username":"reader","password":"..","ssl_mode":"disable" }
{ "host":"dest-host","port":26257,"database":"defaultdb","username":"root","password":"","ssl_mode":"disable" }
```
```sql
CREATE SCHEMA IF NOT EXISTS analytics;
```

---

## Destination DDLs (CockroachDB)

```sql
CREATE TABLE IF NOT EXISTS analytics.dim_users (
    user_uuid     UUID         PRIMARY KEY,
    email_address STRING,      full_name STRING,
    city STRING,  country STRING, account_tier STRING,
    registered_on DATE,        last_active DATE
);
CREATE TABLE IF NOT EXISTS analytics.fact_orders (
    order_id       UUID PRIMARY KEY,
    customer_ref   UUID,
    order_amount   DECIMAL(10,2),
    order_status   STRING,
    payment_method STRING,     channel STRING, currency STRING,
    is_high_value  BOOL,       placed_on DATE
);
CREATE TABLE IF NOT EXISTS analytics.fact_payments (
    payment_id    UUID PRIMARY KEY,
    order_ref     UUID,
    pay_method    STRING,
    pay_amount    DECIMAL(10,2),
    pay_status    STRING,
    is_successful BOOL,
    paid_on       DATE
);
```

---

## Notes
- `UUID` native type in CockroachDB — no VARCHAR(36) needed
- BOOL returns `true`/`false`, not 0/1

## Verify
```sql
SELECT is_high_value FROM analytics.fact_orders LIMIT 5;   -- true/false
SELECT user_uuid FROM analytics.dim_users LIMIT 3;         -- UUID type
```
