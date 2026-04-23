# MySQL Source — Manual Testing Guide

**Streams:** 3 (`mydb.products`, `mydb.inventory`, `mydb.categories`)  
**Credential format:** host + port + database + username + password  
**DuckDB naming:** `mydb__products`, `mydb__inventory`, `mydb__categories`

---

## Credential Setup

```json
{
  "host": "your-mysql-host",
  "port": 3306,
  "database": "mydb",
  "username": "your_user",
  "password": "your_password",
  "ssl_mode": "disable"
}
```

---

## Required Source Tables

```sql
-- mydb.products
CREATE TABLE IF NOT EXISTS mydb.products (
    id          INT AUTO_INCREMENT PRIMARY KEY,
    name        VARCHAR(255) NOT NULL,
    sku         VARCHAR(100) UNIQUE,
    price       DECIMAL(10, 2) NOT NULL,
    active      TINYINT(1) DEFAULT 1,
    description TEXT,
    created_at  DATETIME DEFAULT NOW(),
    updated_at  DATETIME DEFAULT NOW() ON UPDATE NOW()
);

-- mydb.inventory
CREATE TABLE IF NOT EXISTS mydb.inventory (
    id          INT AUTO_INCREMENT PRIMARY KEY,
    product_id  INT REFERENCES mydb.products(id),
    quantity    INT NOT NULL DEFAULT 0,
    warehouse   VARCHAR(100),
    updated_at  DATETIME DEFAULT NOW() ON UPDATE NOW()
);

-- mydb.categories
CREATE TABLE IF NOT EXISTS mydb.categories (
    id          INT AUTO_INCREMENT PRIMARY KEY,
    name        VARCHAR(100) NOT NULL UNIQUE,
    slug        VARCHAR(100) UNIQUE,
    parent_id   INT,
    created_at  DATETIME DEFAULT NOW()
);
```

---

## Stream Reference

| Stream key | DuckDB staging name | Key columns | INCREMENTAL key |
|-----------|-------------------|-------------|----------------|
| `mydb.products` | `mydb__products` | `id`, `name`, `sku`, `price`, `active`, `created_at`, `updated_at` | `updated_at` |
| `mydb.inventory` | `mydb__inventory` | `id`, `product_id`, `quantity`, `warehouse`, `updated_at` | `updated_at` |
| `mydb.categories` | `mydb__categories` | `id`, `name`, `slug`, `parent_id`, `created_at` | `created_at` |

---

## Scenario S-MY-1 — Full Table Sync: `mydb.products`

**Destination DDL (PostgreSQL destination):**
```sql
CREATE TABLE analytics.mysql_products_hd (
    id          INTEGER PRIMARY KEY,
    name        TEXT,
    sku         TEXT,
    price       NUMERIC(10, 2),
    active      BOOLEAN,
    created_at  TIMESTAMPTZ
);
```

**dbt SQL:**
```sql
SELECT
    id,
    name,
    sku,
    price,
    active = 1  AS active,
    created_at
FROM {{ source('raw', 'mydb__products') }}
```

> ℹ️ MySQL `TINYINT(1)` for booleans — cast with `active = 1` to get a proper boolean.

**Verify:** `SELECT COUNT(*) FROM analytics.mysql_products_hd;` matches source row count.

---

## Scenario S-MY-2 — Incremental Sync: `mydb.inventory`

**Sync mode:** `INCREMENTAL`, replication key `updated_at`

**dbt SQL:**
```sql
SELECT
    id,
    product_id,
    quantity,
    warehouse,
    updated_at
FROM {{ source('raw', 'mydb__inventory') }}
WHERE quantity > 0
```

---

## Scenario S-MY-3 — Product + Inventory JOIN

Both `mydb.products` AND `mydb.inventory` in Source Panel:

**dbt SQL:**
```sql
SELECT
    p.id          AS product_id,
    p.name        AS product_name,
    p.sku,
    p.price,
    i.quantity,
    i.warehouse,
    i.updated_at
FROM {{ source('raw', 'mydb__products') }} AS p
LEFT JOIN {{ source('raw', 'mydb__inventory') }} AS i
    ON p.id = i.product_id
WHERE p.active = 1
```

---

## Scenario S-MY-4 — Category Hierarchy Flattening

MySQL `categories` has self-referential `parent_id`. Flatten parent → child:

**dbt SQL:**
```sql
SELECT
    c.id,
    c.name,
    c.slug,
    p.name  AS parent_name,
    c.created_at
FROM {{ source('raw', 'mydb__categories') }} AS c
LEFT JOIN {{ source('raw', 'mydb__categories') }} AS p
    ON c.parent_id = p.id
```

---

## Scenario S-MY-5 — Normalisation: Cast `active` TINYINT → boolean

**Rule:**
```json
{ "rule_type": "cast", "table": "mydb.products", "column": "active", "cast_to": "boolean" }
```

**dbt SQL:**
```sql
SELECT id, name, sku, price, active, created_at
FROM {{ source('raw', 'mydb__products') }}
```

---

## Scenario S-MY-6 — Normalisation: Rename + Exclude

**Rules:**
```json
[
  { "rule_type": "rename",  "table": "mydb.products", "column": "name",        "destination_name": "product_name" },
  { "rule_type": "exclude", "table": "mydb.products", "column": "description" }
]
```

**dbt SQL:**
```sql
SELECT id, product_name, sku, price, active, created_at
FROM {{ source('raw', 'mydb__products') }}
```

---

## Scenario S-MY-7 — Low Stock Alert Model

**dbt SQL:**
```sql
SELECT
    product_id,
    warehouse,
    quantity,
    CASE
        WHEN quantity = 0  THEN 'out_of_stock'
        WHEN quantity < 10 THEN 'low_stock'
        ELSE 'in_stock'
    END AS stock_status,
    updated_at
FROM {{ source('raw', 'mydb__inventory') }}
ORDER BY quantity ASC
```

---

## Scenario S-MY-8 — SSL + Authentication Error Paths

| Error | How to trigger | Expected |
|-------|---------------|---------|
| Wrong password | Set wrong `password` in connection | Connection test fails: auth error |
| Wrong database | Set `database = "nonexistent"` | Phase 1 fails: `Unknown database` |
| Port mismatch | Set `port = 3307` | Connection test fails: connection refused |
