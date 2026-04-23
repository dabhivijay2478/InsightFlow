# Pipeline 32 — MySQL Source → MySQL Destination

**Source streams:** 3 | **Destination:** MySQL (different host/db)

> dbt SQL identical to `31-mysql-to-postgres.md`. DDL uses MySQL types.

---

## Connections
```json
{ "host":"src-host","port":3306,"database":"mydb","username":"reader","password":"..." }
{ "host":"dest-host","port":3306,"database":"analytics","username":"writer","password":"..." }
```

---

## Destination DDLs (MySQL)

```sql
CREATE TABLE analytics.product_master (
    product_id    INT          PRIMARY KEY,
    product_name  VARCHAR(255),
    product_sku   VARCHAR(100),
    unit_price    DECIMAL(10,2),
    is_active     TINYINT(1),
    added_on      DATE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE analytics.stock_ledger (
    inventory_id   INT PRIMARY KEY,
    product_ref    INT,
    qty_on_hand    INT,
    warehouse_code VARCHAR(100),
    stock_status   VARCHAR(20),
    refreshed_on   DATE
) ENGINE=InnoDB;

CREATE TABLE analytics.category_tree (
    category_id   INT PRIMARY KEY,
    category_name VARCHAR(255),
    url_slug      VARCHAR(255),
    parent_ref    INT,
    is_root       TINYINT(1),
    added_on      DATE
) ENGINE=InnoDB;
```

---

## Verify
```sql
SELECT is_active FROM analytics.product_master LIMIT 5;      -- 0 or 1
SELECT stock_status FROM analytics.stock_ledger LIMIT 5;     -- text label
SELECT is_root FROM analytics.category_tree LIMIT 5;         -- 0 or 1
```
