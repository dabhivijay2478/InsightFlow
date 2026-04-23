# Pipeline 34 — MySQL Source → SQLite Destination

**Source streams:** 3 | **Destination:** SQLite

> dbt SQL identical to `31-mysql-to-postgres.md`.

---

## Connections
```json
{ "host":"src-host","port":3306,"database":"mydb","username":"reader","password":"..." }
{ "database": "/absolute/path/to/analytics.db" }
```

---

## Destination DDLs (SQLite)

```sql
CREATE TABLE IF NOT EXISTS product_master (
    product_id INTEGER PRIMARY KEY, product_name TEXT, product_sku TEXT,
    unit_price REAL, is_active INTEGER, added_on TEXT
);
CREATE TABLE IF NOT EXISTS stock_ledger (
    inventory_id INTEGER PRIMARY KEY, product_ref INTEGER,
    qty_on_hand INTEGER, warehouse_code TEXT,
    stock_status TEXT, refreshed_on TEXT
);
CREATE TABLE IF NOT EXISTS category_tree (
    category_id INTEGER PRIMARY KEY, category_name TEXT, url_slug TEXT,
    parent_ref INTEGER, is_root INTEGER, added_on TEXT
);
```

---

## Verify
```bash
DB=/absolute/path/to/analytics.db
sqlite3 $DB "SELECT typeof(unit_price), unit_price FROM product_master LIMIT 3;"
sqlite3 $DB "SELECT is_active FROM product_master LIMIT 5;"
sqlite3 $DB "SELECT stock_status FROM stock_ledger GROUP BY stock_status;"
```
