# MySQL Source — UI Testing (3 Streams × 5 Destinations)

> Universal builder steps in `builder-walkthrough.md`.

---

## Phase 1 — Source Panel (MySQL)

### Credential Fields
| Field | Value |
|-------|-------|
| **Host** | `src-host` |
| **Port** | `3306` |
| **Database** | `mydb` |
| **Username** | `reader` |
| **Password** | `***` |

**Test Connection → ✅**
- ❌ `Access denied for user`: wrong credentials
- ❌ `Can't connect to MySQL server`: wrong host/port/firewall
- ❌ `Unknown database`: database name incorrect

### SSL Test (MySQL)
| Action | Expected |
|--------|---------|
| Enable SSL toggle | Extra fields appear: CA cert, client cert, client key |
| Valid cert paths | ✅ Connects with SSL |
| Missing cert file | ❌ `SSL connection error` |

---

## Step 2b — Stream Selection & Sync Mode

Source tables:
```sql
CREATE TABLE mydb.products (id INT AUTO_INCREMENT PRIMARY KEY, name VARCHAR(255), sku VARCHAR(100), description TEXT, price DECIMAL(10,2), active TINYINT(1), created_at DATETIME);
CREATE TABLE mydb.inventory (id INT AUTO_INCREMENT PRIMARY KEY, product_id INT, quantity INT, warehouse VARCHAR(100), updated_at DATETIME);
CREATE TABLE mydb.categories (id INT AUTO_INCREMENT PRIMARY KEY, name VARCHAR(255), slug VARCHAR(255), parent_id INT, created_at DATETIME);
```

| Stream | Sync Mode | Cursor Field |
|--------|-----------|-------------|
| `mydb.products` | INCREMENTAL | `created_at` |
| `mydb.inventory` | INCREMENTAL | `updated_at` |
| `mydb.categories` | FULL TABLE | — |

---

## Phase 2 — Stream→Table Mapping

| Stream | Table |
|--------|-------|
| `mydb.products` | `product_master` |
| `mydb.inventory` | `stock_ledger` |
| `mydb.categories` | `category_tree` |

---

## Phase 3 — Normalisation Rules

### `mydb.products`
| Rule | Column | Target |
|------|--------|--------|
| Rename | `id` | `product_id` |
| Rename | `name` | `product_name` |
| Rename | `sku` | `product_sku` |
| Rename | `price` | `unit_price` |
| Cast | `active` | Boolean |
| Exclude | `description` | — |

**UI test**: Cast `active` (TINYINT) → Boolean. Preview should show `true`/`false`.

### `mydb.inventory`
| Rule | Column | Target |
|------|--------|--------|
| Rename | `id` | `inventory_id` |
| Rename | `product_id` | `product_ref` |
| Rename | `quantity` | `qty_on_hand` |
| Rename | `warehouse` | `warehouse_code` |

### `mydb.categories`
| Rule | Column | Target |
|------|--------|--------|
| Rename | `id` | `category_id` |
| Rename | `name` | `category_name` |
| Rename | `slug` | `url_slug` |
| Rename | `parent_id` | `parent_ref` |

---

## Phase 4 — dbt SQL

### Stream 1 — `mydb.products` → `product_master`
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
**Preview check**: `is_active` = `true`/`false` (not 0/1 after TINYINT cast); `description` absent

---

### Stream 2 — `mydb.inventory` → `stock_ledger`
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
**Preview check**: `stock_status` values = `'in_stock'`/`'low_stock'`/`'out_of_stock'`

---

### Stream 3 — `mydb.categories` → `category_tree`
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
**Preview check**: `is_root` = `true` for top-level categories

---

## Phase 5 — Preview Checks

| Column | Expected |
|--------|---------|
| `is_active` | `true`/`false` |
| `stock_status` | One of three string labels |
| `is_root` | `true`/`false` |
| `unit_price` | Decimal e.g. `29.99` |

---

## Phase 6 — Destination Type Verification

After delivery:

| Destination | `is_active` | `unit_price` | `stock_status` |
|-------------|------------|-------------|---------------|
| PostgreSQL | `BOOLEAN` true/false | `NUMERIC` | `TEXT` |
| MySQL | `TINYINT(1)` 0/1 | `DECIMAL` | `VARCHAR` |
| MariaDB | `TINYINT(1)` 0/1 | `DECIMAL` | `VARCHAR` |
| SQLite | `INTEGER` 0/1 | `REAL` | `TEXT` |
| CockroachDB | `BOOL` true/false | `DECIMAL` | `STRING` |

---

## Phase 7 — Failure Scenarios

| Scenario | Expected |
|---------|---------|
| `active = 2` (non-standard TINYINT) | Cast `active` → boolean: `2 != 0` = true |
| `quantity = NULL` | CASE ELSE → `'in_stock'` — add COALESCE to fix |
| `parent_id = 0` (MySQL root convention) | `is_root = false` — fix: `parent_id IS NULL OR parent_id = 0` |
| DATETIME timezone: local time stored | `added_on` DATE reflects local calendar — note in docs |
| Reader missing SELECT on `products` | Phase 1 ❌ `SELECT command denied` |

---

## Phase 8 — Verify

```sql
SELECT is_active, COUNT(*) FROM analytics.product_master GROUP BY is_active;
SELECT stock_status, COUNT(*) FROM analytics.stock_ledger GROUP BY stock_status;
SELECT is_root FROM analytics.category_tree LIMIT 5;
SELECT product_id, COUNT(*) FROM analytics.product_master GROUP BY product_id HAVING COUNT(*)>1; -- 0
```

---

## Incremental Re-run Test

1. Run pipeline — note `stock_ledger` row count
2. Update one inventory row: `UPDATE mydb.inventory SET quantity=0, updated_at=NOW() WHERE id=1;`
3. Re-run pipeline
4. ✅ Only that 1 row re-delivered; `stock_status` = `'out_of_stock'` updated in destination
