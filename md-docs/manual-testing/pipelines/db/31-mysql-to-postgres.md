# Pipeline 31 — MySQL Source → PostgreSQL Destination

**Source streams:** 3 | **Destination:** PostgreSQL

---

## Connections

### Source — MySQL
```json
{ "host":"src-host","port":3306,"database":"mydb","username":"reader","password":"..." }
```
```sql
-- Source tables:
CREATE TABLE mydb.products (
    id INT AUTO_INCREMENT PRIMARY KEY, name VARCHAR(255), sku VARCHAR(100),
    description TEXT, price DECIMAL(10,2), active TINYINT(1), created_at DATETIME
);
CREATE TABLE mydb.inventory (
    id INT AUTO_INCREMENT PRIMARY KEY, product_id INT,
    quantity INT, warehouse VARCHAR(100), updated_at DATETIME
);
CREATE TABLE mydb.categories (
    id INT AUTO_INCREMENT PRIMARY KEY, name VARCHAR(255), slug VARCHAR(255),
    parent_id INT, created_at DATETIME
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

## Stream 1 — `mydb.products` → `analytics.product_master`

### Step 1 — DDL
```sql
CREATE TABLE analytics.product_master (
    product_id    INTEGER     PRIMARY KEY,   -- source: id
    product_name  TEXT,                       -- source: name (renamed)
    product_sku   TEXT,                       -- source: sku (renamed)
    unit_price    NUMERIC(10,2),              -- source: price (renamed)
    is_active     BOOLEAN,                    -- source: active TINYINT(1) → BOOLEAN
    added_on      DATE                        -- source: created_at DATETIME → DATE
    -- description excluded
);
```
### Step 3 — Panel: `mydb.products` | `INCREMENTAL` | key: `created_at`
### Step 4 — Normalisation
```json
[
  { "rule_type": "cast",    "table": "mydb.products", "column": "active",      "cast_to": "boolean" },
  { "rule_type": "exclude", "table": "mydb.products", "column": "description" }
]
```
### Step 5 — dbt SQL
```sql
SELECT
    id               AS product_id,
    name             AS product_name,
    sku              AS product_sku,
    price            AS unit_price,
    active           AS is_active,
    created_at::DATE AS added_on
FROM {{ source('raw', 'mydb__products') }}
WHERE sku IS NOT NULL
```
### Step 8 — Verify
```sql
SELECT product_id, product_name, is_active FROM analytics.product_master LIMIT 5;
-- is_active: true/false (not 0/1)
-- 'description' column must NOT exist
\d analytics.product_master
```

---

## Stream 2 — `mydb.inventory` → `analytics.stock_ledger`

### Step 1 — DDL
```sql
CREATE TABLE analytics.stock_ledger (
    inventory_id   INTEGER     PRIMARY KEY,
    product_ref    INTEGER,                   -- source: product_id (renamed)
    qty_on_hand    INTEGER,                   -- source: quantity (renamed)
    warehouse_code TEXT,                      -- source: warehouse (renamed)
    stock_status   TEXT,                      -- derived: CASE WHEN quantity
    refreshed_on   DATE                       -- source: updated_at → DATE
);
```
### Step 5 — dbt SQL
```sql
SELECT
    id          AS inventory_id,
    product_id  AS product_ref,
    quantity    AS qty_on_hand,
    warehouse   AS warehouse_code,
    CASE
        WHEN quantity = 0   THEN 'out_of_stock'
        WHEN quantity < 10  THEN 'low_stock'
        ELSE                     'in_stock'
    END         AS stock_status,
    updated_at::DATE AS refreshed_on
FROM {{ source('raw', 'mydb__inventory') }}
```
### Step 8 — Verify
```sql
SELECT warehouse_code, stock_status, qty_on_hand FROM analytics.stock_ledger LIMIT 10;
-- stock_status: 'in_stock','low_stock','out_of_stock' only
```

---

## Stream 3 — `mydb.categories` → `analytics.category_tree`

### Step 1 — DDL
```sql
CREATE TABLE analytics.category_tree (
    category_id   INTEGER     PRIMARY KEY,
    category_name TEXT,                       -- source: name (renamed)
    url_slug      TEXT,                       -- source: slug (renamed)
    parent_ref    INTEGER,                    -- source: parent_id (renamed)
    is_root       BOOLEAN,                    -- derived: parent_id IS NULL
    added_on      DATE
);
```
### Step 5 — dbt SQL
```sql
SELECT
    id               AS category_id,
    name             AS category_name,
    slug             AS url_slug,
    parent_id        AS parent_ref,
    parent_id IS NULL AS is_root,
    created_at::DATE  AS added_on
FROM {{ source('raw', 'mydb__categories') }}
```

---

## Edge Cases

| Scenario | Expected |
|---------|---------|
| `active = 2` (unexpected TINYINT value) | Cast → boolean: `2 != 0` = true |
| `quantity = NULL` | stock_status falls to ELSE = 'in_stock' — add COALESCE if needed |
| `parent_id = 0` (MySQL convention for root) | `is_root = false` — adjust: `parent_id IS NULL OR parent_id = 0` |
| DATETIME timezone: MySQL stores in local time | Destination DATE reflects local calendar date |
