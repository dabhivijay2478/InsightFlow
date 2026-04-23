# Pipeline 35 — MySQL Source → CockroachDB Destination

**Source streams:** 3 | **Destination:** CockroachDB

> dbt SQL identical to `31-mysql-to-postgres.md`.

---

## Connections
```json
{ "host":"src-host","port":3306,"database":"mydb","username":"reader","password":"..." }
{ "host":"dest-host","port":26257,"database":"defaultdb","username":"root","password":"","ssl_mode":"disable" }
```
```sql
CREATE SCHEMA IF NOT EXISTS analytics;
```

---

## Destination DDLs (CockroachDB)

```sql
CREATE TABLE IF NOT EXISTS analytics.product_master (
    product_id   INT8   PRIMARY KEY,
    product_name STRING, product_sku STRING,
    unit_price   DECIMAL(10,2),
    is_active    BOOL,  added_on DATE
);
CREATE TABLE IF NOT EXISTS analytics.stock_ledger (
    inventory_id INT8 PRIMARY KEY, product_ref INT8,
    qty_on_hand  INT8, warehouse_code STRING,
    stock_status STRING, refreshed_on DATE
);
CREATE TABLE IF NOT EXISTS analytics.category_tree (
    category_id  INT8 PRIMARY KEY, category_name STRING, url_slug STRING,
    parent_ref   INT8, is_root BOOL, added_on DATE
);
```

---

## Verify
```sql
SELECT is_active FROM analytics.product_master LIMIT 5;    -- true/false
SELECT stock_status FROM analytics.stock_ledger GROUP BY stock_status;
SELECT is_root FROM analytics.category_tree LIMIT 5;       -- true/false
```
