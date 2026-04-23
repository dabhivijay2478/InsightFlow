# Shopify Source — UI Testing (All 14 Streams × 5 Destinations)

> Universal builder steps in `builder-walkthrough.md`.

---

## Phase 1 — Source Panel (Shopify)

### Credential Fields
| Field | Value |
|-------|-------|
| **Shop URL** | `your-store.myshopify.com` (no https://) |
| **API Password** | `shppa_...` |

**Test Connection → ✅**
- ❌ `401`: wrong password or expired token
- ❌ `404`: shop URL incorrect

---

## Step 2b — Stream Selection & Sync Mode

| Stream | Sync Mode | Cursor Field |
|--------|-----------|-------------|
| `products` | INCREMENTAL | `updated_at` |
| `orders` | INCREMENTAL | `updated_at` |
| `customers` | INCREMENTAL | `updated_at` |
| `draft_orders` | INCREMENTAL | `updated_at` |
| `abandoned_checkouts` | INCREMENTAL | `created_at` |
| `collections` | FULL TABLE | — |
| `collects` | FULL TABLE | — |
| `product_variants` | INCREMENTAL | `updated_at` |
| `inventory_items` | INCREMENTAL | `updated_at` |
| `locations` | FULL TABLE | — |
| `pages` | INCREMENTAL | `updated_at` |
| `price_rules` | INCREMENTAL | `updated_at` |
| `themes` | FULL TABLE | — |
| `articles` | INCREMENTAL | `updated_at` |

---

## Phase 2 — Destination Stream→Table Mapping

| Stream | Table name (all 5 destinations) |
|--------|--------------------------------|
| products | `shopify_products` |
| orders | `shopify_orders` |
| customers | `shopify_customers` |
| draft_orders | `shopify_draft_orders` |
| abandoned_checkouts | `shopify_abandoned` |
| collections | `shopify_collections` |
| collects | `shopify_collects` |
| product_variants | `shopify_variants` |
| inventory_items | `shopify_inventory_items` |
| locations | `shopify_locations` |
| pages | `shopify_pages` |
| price_rules | `shopify_price_rules` |
| themes | `shopify_themes` |
| articles | `shopify_articles` |

Schema: `analytics` (or `main` for SQLite)

---

## Phase 3 — Normalisation Rules

### `shopify.orders`
| Rule | Column | Target |
|------|--------|--------|
| Rename | `id` | `order_id` |
| Rename | `financial_status` | `order_status` |
| Rename | `fulfillment_status` | `fulfil_status` |
| Exclude | `line_items` | — |
| Exclude | `shipping_lines` | — |
| Exclude | `tax_lines` | — |
| Exclude | `note_attributes` | — |

### `shopify.customers`
| Rule | Column | Target |
|------|--------|--------|
| Rename | `id` | `customer_id` |
| Rename | `orders_count` | `order_count` |
| Rename | `total_spent` | `lifetime_spend` |
| Cast | `verified_email` | Boolean |
| Cast | `total_spent` | Numeric |
| Exclude | `addresses` | — |
| Exclude | `default_address` | — |

### `shopify.products`
| Rule | Column | Target |
|------|--------|--------|
| Rename | `id` | `product_id` |
| Rename | `title` | `product_name` |
| Rename | `vendor` | `brand` |
| Rename | `product_type` | `category` |
| Exclude | `variants` | — |
| Exclude | `images` | — |
| Exclude | `options` | — |

---

## Phase 4 — dbt SQL (stream-by-stream)

### Stream 1 — `shopify.products`
```sql
SELECT
    id                         AS product_id,
    title                      AS product_name,
    vendor                     AS brand,
    product_type               AS category,
    status = 'active'          AS is_published,
    published_at::DATE         AS listed_on,
    updated_at::DATE           AS updated_on
FROM {{ source('raw', 'shopify__products') }}
WHERE status != 'archived'
```
**Preview check**: `is_published` = true/false; `listed_on` = DATE; no `variants` column

---

### Stream 2 — `shopify.orders`
```sql
SELECT
    id                                    AS order_id,
    email                                 AS customer_email,
    CAST(total_price AS NUMERIC)          AS order_total,
    currency,
    financial_status                      AS order_status,
    fulfillment_status                    AS fulfil_status,
    shipping_address->>'city'             AS ship_city,
    shipping_address->>'country_code'     AS ship_country,
    JSON_ARRAY_LENGTH(line_items)         AS item_count,
    created_at::DATE                      AS ordered_on
FROM {{ source('raw', 'shopify__orders') }}
WHERE financial_status = 'paid'
```
**Preview check**: `ship_city` plain string; `order_total` decimal; `item_count` integer

---

### Stream 3 — `shopify.customers`
```sql
SELECT
    id                                                           AS customer_id,
    email,
    TRIM(CONCAT(COALESCE(first_name,''),' ',COALESCE(last_name,''))) AS full_name,
    orders_count                                                 AS order_count,
    CAST(NULLIF(total_spent,'') AS NUMERIC)                      AS lifetime_spend,
    verified_email                                               AS is_verified,
    created_at::DATE                                             AS registered_on
FROM {{ source('raw', 'shopify__customers') }}
WHERE email IS NOT NULL
```
**Preview check**: `full_name` = combined string; `lifetime_spend` = decimal

---

### Stream 4 — `shopify.draft_orders`
```sql
SELECT
    id                                  AS draft_id,
    customer->>'id'                     AS customer_ref,
    CAST(total_price AS NUMERIC)        AS draft_total,
    currency,
    status                              AS draft_status,
    created_at::DATE                    AS created_on,
    invoice_url IS NOT NULL             AS has_invoice
FROM {{ source('raw', 'shopify__draft_orders') }}
```

---

### Stream 5 — `shopify.abandoned_checkouts`
```sql
SELECT
    id                               AS checkout_id,
    email                            AS customer_email,
    CAST(total_price AS NUMERIC)     AS cart_total,
    currency,
    JSON_ARRAY_LENGTH(line_items)    AS item_count,
    created_at                       AS abandoned_at
FROM {{ source('raw', 'shopify__abandoned_checkouts') }}
WHERE completed_at IS NULL
```

---

### Stream 6 — `shopify.collections`
```sql
SELECT
    id                             AS collection_id,
    title                          AS collection_name,
    handle                         AS url_handle,
    published_at IS NOT NULL       AS is_published,
    published_at::DATE             AS published_on
FROM {{ source('raw', 'shopify__collections') }}
```

---

### Stream 7 — `shopify.collects`
```sql
SELECT
    id               AS collect_id,
    collection_id    AS collection_ref,
    product_id       AS product_ref,
    sort_value,
    created_at::DATE AS added_on
FROM {{ source('raw', 'shopify__collects') }}
```

---

### Stream 8 — `shopify.product_variants`
```sql
SELECT
    id                           AS variant_id,
    product_id                   AS product_ref,
    title                        AS variant_title,
    sku,
    CAST(price AS NUMERIC)       AS unit_price,
    inventory_quantity           AS inventory_qty,
    created_at::DATE             AS created_on
FROM {{ source('raw', 'shopify__product_variants') }}
```

---

### Stream 9 — `shopify.inventory_items`
```sql
SELECT
    id                                        AS item_id,
    sku,
    tracked                                   AS is_tracked,
    CAST(NULLIF(cost,'') AS NUMERIC)          AS cost_price,
    created_at::DATE                          AS created_on
FROM {{ source('raw', 'shopify__inventory_items') }}
```
**Preview check**: `cost_price` NULL if cost is empty string; `is_tracked` boolean

---

### Stream 10 — `shopify.locations`
```sql
SELECT
    id       AS location_id,
    name     AS location_name,
    address1, city, country_code,
    active   AS is_active,
    legacy   AS is_legacy
FROM {{ source('raw', 'shopify__locations') }}
```

---

### Stream 11 — `shopify.pages`
```sql
SELECT
    id                          AS page_id,
    title                       AS page_title,
    handle                      AS url_handle,
    author                      AS author_name,
    published_at IS NOT NULL    AS is_published,
    published_at::DATE          AS published_on
FROM {{ source('raw', 'shopify__pages') }}
```

---

### Stream 12 — `shopify.price_rules`
```sql
SELECT
    id                             AS rule_id,
    title                          AS rule_title,
    value_type,
    CAST(value AS NUMERIC)         AS discount_value,
    starts_at::DATE                AS valid_from,
    ends_at::DATE                  AS valid_to,
    usage_limit
FROM {{ source('raw', 'shopify__price_rules') }}
```

---

### Stream 13 — `shopify.themes`
```sql
SELECT
    id               AS theme_id,
    name             AS theme_name,
    role             AS theme_role,
    role = 'main'    AS is_live,
    created_at::DATE AS created_on
FROM {{ source('raw', 'shopify__themes') }}
```

---

### Stream 14 — `shopify.articles`
```sql
SELECT
    id                          AS article_id,
    title                       AS article_title,
    author                      AS author_name,
    blog_id,
    published_at IS NOT NULL    AS is_published,
    published_at::DATE          AS published_on
FROM {{ source('raw', 'shopify__articles') }}
```

---

## Phase 5 — Preview Checks

| Stream | Column | Expected in preview |
|--------|--------|-------------------|
| orders | `ship_city` | Plain string e.g. `"New York"` |
| orders | `order_total` | Decimal e.g. `99.00` |
| orders | `item_count` | Integer e.g. `3` |
| customers | `full_name` | `"Jane Smith"` (combined) |
| inventory_items | `cost_price` | NULL or decimal; NOT empty string |
| price_rules | `discount_value` | Negative decimal e.g. `-10.00` |

---

## Phase 6 — Schedule Tests

| Test | Config | Expected |
|------|--------|---------|
| Hourly sync | Cron `0 * * * *` | Runs every hour |
| Daily | Cron `0 2 * * *` | Runs at 02:00 UTC |
| None | None | Manual only |

---

## Phase 7 — Run Status & Failure Scenarios

| Scenario | How to trigger | Expected |
|---------|---------------|---------|
| `total_price` TEXT cast failure | Insert `'N/A'` in Shopify test order | Phase 2 ❌ cast error; fix with `NULLIF` |
| `shipping_address` NULL | Digital product order | `ship_city` = NULL — Phase 3 ✅ |
| Empty line_items | Draft order with 0 items | `item_count` = 0 — ✅ |
| Destination table missing | Drop table before run | Phase 0 ❌ |

---

## Phase 8 — Destination Verify

```sql
-- Shopify orders
SELECT order_total, ship_city, item_count FROM analytics.shopify_orders LIMIT 5;
SELECT order_id, COUNT(*) FROM analytics.shopify_orders GROUP BY order_id HAVING COUNT(*)>1;
-- 0 rows (no duplicates)

-- Shopify customers full_name check
SELECT full_name FROM analytics.shopify_customers WHERE full_name LIKE ' %' OR full_name LIKE '% ';
-- 0 rows (TRIM applied)
```
