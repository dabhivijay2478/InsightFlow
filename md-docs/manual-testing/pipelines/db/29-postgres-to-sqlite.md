# Pipeline 29 — PostgreSQL Source → SQLite Destination

**Source streams:** 3 | **Destination:** SQLite

> dbt SQL identical to `26-postgres-to-postgres.md`.

---

## Connections
```json
{ "host":"src-host","port":5432,"database":"mydb","username":"reader","password":"..","ssl_mode":"disable" }
{ "database": "/absolute/path/to/analytics.db" }
```

---

## Destination DDLs (SQLite)

```sql
CREATE TABLE IF NOT EXISTS dim_users (
    user_uuid TEXT PRIMARY KEY, email_address TEXT, full_name TEXT,
    city TEXT, country TEXT, account_tier TEXT,
    registered_on TEXT, last_active TEXT
);
CREATE TABLE IF NOT EXISTS fact_orders (
    order_id TEXT PRIMARY KEY, customer_ref TEXT,
    order_amount REAL, order_status TEXT,
    payment_method TEXT, channel TEXT, currency TEXT,
    is_high_value INTEGER, placed_on TEXT
);
CREATE TABLE IF NOT EXISTS fact_payments (
    payment_id TEXT PRIMARY KEY, order_ref TEXT,
    pay_method TEXT, pay_amount REAL,
    pay_status TEXT, is_successful INTEGER, paid_on TEXT
);
```

---

## Verify
```bash
DB=/absolute/path/to/analytics.db
sqlite3 $DB "SELECT typeof(order_amount), order_amount FROM fact_orders LIMIT 3;"   # real
sqlite3 $DB "SELECT is_high_value FROM fact_orders LIMIT 5;"                        # 0 or 1
sqlite3 $DB "SELECT user_uuid FROM dim_users LIMIT 3;"                              # UUID string
```
