# Pipeline 06 — Shopify → PostgreSQL

**Streams:** 14 | **Destination:** PostgreSQL

---

## Connections
```json
{ "shop": "your-store.myshopify.com", "api_password": "shppa_..." }
{ "host":"..","port":5432,"database":"analytics","username":"writer","password":"..","ssl_mode":"disable" }
```
```sql
CREATE SCHEMA IF NOT EXISTS analytics;
```

---

## Stream 1 — `shopify.products` → `analytics.shopify_products`

### Step 1 — DDL
```sql
CREATE TABLE analytics.shopify_products (
    product_id    BIGINT      PRIMARY KEY,   -- source: id
    product_name  TEXT,                       -- source: title
    brand         TEXT,                       -- source: vendor
    category      TEXT,                       -- source: product_type
    is_published  BOOLEAN,                    -- derived: status = 'active'
    listed_on     DATE,                       -- source: published_at → DATE
    updated_on    DATE                        -- source: updated_at → DATE
);
```
### Step 3 — Panel: `shopify.products` | `FULL_TABLE`
### Step 5 — dbt SQL
```sql
SELECT
    id                 AS product_id,
    title              AS product_name,
    vendor             AS brand,
    product_type       AS category,
    status = 'active'  AS is_published,
    published_at::DATE AS listed_on,
    updated_at::DATE   AS updated_on
FROM {{ source('raw', 'shopify__products') }}
WHERE status != 'archived'
```
### Step 8 — Verify
```sql
SELECT product_name, brand, is_published FROM analytics.shopify_products LIMIT 5;
-- is_published: true/false; listed_on: DATE only
```

---

## Stream 2 — `shopify.orders` → `analytics.shopify_orders`

### Step 1 — DDL
```sql
CREATE TABLE analytics.shopify_orders (
    order_id       BIGINT       PRIMARY KEY,
    customer_email TEXT,
    order_total    NUMERIC(10,2),             -- source: total_price TEXT → NUMERIC
    currency       TEXT,
    order_status   TEXT,                       -- source: financial_status
    fulfil_status  TEXT,                       -- source: fulfillment_status
    ship_city      TEXT,                       -- from JSON: shipping_address->>'city'
    ship_country   TEXT,                       -- from JSON: shipping_address->>'country_code'
    item_count     INTEGER,                    -- from JSON: JSON_ARRAY_LENGTH(line_items)
    ordered_on     DATE
);
```
### Step 5 — dbt SQL
```sql
SELECT
    id                                   AS order_id,
    email                                AS customer_email,
    CAST(total_price AS NUMERIC)         AS order_total,
    currency,
    financial_status                     AS order_status,
    fulfillment_status                   AS fulfil_status,
    shipping_address->>'city'            AS ship_city,
    shipping_address->>'country_code'    AS ship_country,
    JSON_ARRAY_LENGTH(line_items)        AS item_count,
    created_at::DATE                     AS ordered_on
FROM {{ source('raw', 'shopify__orders') }}
WHERE financial_status = 'paid'
```
### Step 8 — Verify
```sql
SELECT order_total, ship_city, item_count FROM analytics.shopify_orders LIMIT 5;
-- ship_city: plain string; item_count: integer; order_total: decimal
```

---

## Stream 3 — `shopify.customers` → `analytics.shopify_customers`

### Step 1 — DDL
```sql
CREATE TABLE analytics.shopify_customers (
    customer_id    BIGINT      PRIMARY KEY,
    email          TEXT,
    full_name      TEXT,                       -- derived: CONCAT(first_name,' ',last_name)
    order_count    INTEGER,                    -- source: orders_count
    lifetime_spend NUMERIC(10,2),              -- source: total_spent TEXT → NUMERIC
    is_verified    BOOLEAN,
    registered_on  DATE
);
```
### Step 5 — dbt SQL
```sql
SELECT
    id                                         AS customer_id,
    email,
    TRIM(CONCAT(COALESCE(first_name,''),' ',COALESCE(last_name,''))) AS full_name,
    orders_count                               AS order_count,
    CAST(total_spent AS NUMERIC)               AS lifetime_spend,
    verified_email                             AS is_verified,
    created_at::DATE                           AS registered_on
FROM {{ source('raw', 'shopify__customers') }}
WHERE email IS NOT NULL
```

---

## Stream 4 — `shopify.draft_orders` → `analytics.shopify_draft_orders`

### Step 1 — DDL
```sql
CREATE TABLE analytics.shopify_draft_orders (
    draft_id      BIGINT      PRIMARY KEY,
    customer_ref  TEXT,                        -- from JSON: customer->>'id'
    draft_total   NUMERIC(10,2),
    currency      TEXT,
    draft_status  TEXT,                        -- source: status
    created_on    DATE,
    expires_on    DATE
);
```
### Step 5 — dbt SQL
```sql
SELECT
    id                               AS draft_id,
    customer->>'id'                  AS customer_ref,
    CAST(total_price AS NUMERIC)     AS draft_total,
    currency,
    status                           AS draft_status,
    created_at::DATE                 AS created_on,
    invoice_url IS NOT NULL          AS has_invoice
FROM {{ source('raw', 'shopify__draft_orders') }}
```

---

## Stream 5 — `shopify.abandoned_checkouts` → `analytics.shopify_abandoned`

### Step 1 — DDL
```sql
CREATE TABLE analytics.shopify_abandoned (
    checkout_id    BIGINT      PRIMARY KEY,
    customer_email TEXT,
    cart_total     NUMERIC(10,2),
    currency       TEXT,
    item_count     INTEGER,
    abandoned_at   TIMESTAMPTZ
);
```
### Step 5 — dbt SQL
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

## Stream 6 — `shopify.collections` → `analytics.shopify_collections`

### Step 1 — DDL
```sql
CREATE TABLE analytics.shopify_collections (
    collection_id   BIGINT  PRIMARY KEY,
    collection_name TEXT,                      -- source: title
    url_handle      TEXT,                      -- source: handle
    is_published    BOOLEAN,
    published_on    DATE
);
```
### Step 5 — dbt SQL
```sql
SELECT
    id                AS collection_id,
    title             AS collection_name,
    handle            AS url_handle,
    published_at IS NOT NULL   AS is_published,
    published_at::DATE         AS published_on
FROM {{ source('raw', 'shopify__collections') }}
```

---

## Stream 7 — `shopify.collects` → `analytics.shopify_collects`

### Step 1 — DDL
```sql
CREATE TABLE analytics.shopify_collects (
    collect_id      BIGINT  PRIMARY KEY,
    collection_ref  BIGINT,
    product_ref     BIGINT,
    sort_value      TEXT,
    added_on        DATE
);
```
### Step 5 — dbt SQL
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

## Stream 8 — `shopify.product_variants` → `analytics.shopify_variants`

### Step 1 — DDL
```sql
CREATE TABLE analytics.shopify_variants (
    variant_id    BIGINT      PRIMARY KEY,
    product_ref   BIGINT,
    variant_title TEXT,                        -- source: title
    sku           TEXT,
    unit_price    NUMERIC(10,2),               -- source: price TEXT → NUMERIC
    inventory_qty INTEGER,
    created_on    DATE
);
```
### Step 5 — dbt SQL
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

## Stream 9 — `shopify.inventory_items` → `analytics.shopify_inventory_items`

### Step 1 — DDL
```sql
CREATE TABLE analytics.shopify_inventory_items (
    item_id       BIGINT  PRIMARY KEY,
    sku           TEXT,
    is_tracked    BOOLEAN,
    cost_price    NUMERIC(10,2),               -- source: cost TEXT → NUMERIC
    created_on    DATE
);
```
### Step 5 — dbt SQL
```sql
SELECT
    id                          AS item_id,
    sku,
    tracked                     AS is_tracked,
    CAST(NULLIF(cost,'') AS NUMERIC) AS cost_price,
    created_at::DATE            AS created_on
FROM {{ source('raw', 'shopify__inventory_items') }}
```

---

## Stream 10 — `shopify.locations` → `analytics.shopify_locations`

### Step 1 — DDL
```sql
CREATE TABLE analytics.shopify_locations (
    location_id   BIGINT  PRIMARY KEY,
    location_name TEXT,                        -- source: name
    address1      TEXT,
    city          TEXT,
    country_code  TEXT,
    is_active     BOOLEAN,
    is_legacy     BOOLEAN
);
```
### Step 5 — dbt SQL
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

## Stream 11 — `shopify.pages` → `analytics.shopify_pages`

### Step 1 — DDL
```sql
CREATE TABLE analytics.shopify_pages (
    page_id      BIGINT  PRIMARY KEY,
    page_title   TEXT,                         -- source: title
    url_handle   TEXT,                         -- source: handle
    author_name  TEXT,                         -- source: author
    is_published BOOLEAN,
    published_on DATE
);
```
### Step 5 — dbt SQL
```sql
SELECT
    id                   AS page_id,
    title                AS page_title,
    handle               AS url_handle,
    author               AS author_name,
    published_at IS NOT NULL  AS is_published,
    published_at::DATE        AS published_on
FROM {{ source('raw', 'shopify__pages') }}
```

---

## Stream 12 — `shopify.price_rules` → `analytics.shopify_price_rules`

### Step 1 — DDL
```sql
CREATE TABLE analytics.shopify_price_rules (
    rule_id        BIGINT      PRIMARY KEY,
    rule_title     TEXT,
    value_type     TEXT,
    discount_value NUMERIC(10,2),              -- source: value TEXT → NUMERIC
    valid_from     DATE,
    valid_to       DATE,
    usage_limit    INTEGER
);
```
### Step 5 — dbt SQL
```sql
SELECT
    id                              AS rule_id,
    title                           AS rule_title,
    value_type,
    CAST(value AS NUMERIC)          AS discount_value,
    starts_at::DATE                 AS valid_from,
    ends_at::DATE                   AS valid_to,
    usage_limit
FROM {{ source('raw', 'shopify__price_rules') }}
```

---

## Stream 13 — `shopify.themes` → `analytics.shopify_themes`

### Step 1 — DDL
```sql
CREATE TABLE analytics.shopify_themes (
    theme_id    BIGINT  PRIMARY KEY,
    theme_name  TEXT,
    theme_role  TEXT,
    is_live     BOOLEAN,
    created_on  DATE
);
```
### Step 5 — dbt SQL
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

## Stream 14 — `shopify.articles` → `analytics.shopify_articles`

### Step 1 — DDL
```sql
CREATE TABLE analytics.shopify_articles (
    article_id    BIGINT  PRIMARY KEY,
    article_title TEXT,
    author_name   TEXT,
    blog_id       BIGINT,
    is_published  BOOLEAN,
    published_on  DATE
);
```
### Step 5 — dbt SQL
```sql
SELECT
    id                   AS article_id,
    title                AS article_title,
    author               AS author_name,
    blog_id,
    published_at IS NOT NULL  AS is_published,
    published_at::DATE        AS published_on
FROM {{ source('raw', 'shopify__articles') }}
```

---

## Edge Cases

| Scenario | Expected |
|---------|---------|
| `total_price` is TEXT `"0.00"` | `CAST('0.00' AS NUMERIC)` = `0.00` ✅ |
| `shipping_address` NULL (digital orders) | `ship_city` = NULL — OK |
| `line_items` empty array `[]` | `JSON_ARRAY_LENGTH` = 0 |
| `status = 'archived'` filtered out | Only active + draft in destination |
| Destination table missing | Phase 0 fails |
| `total_spent` empty string | `CAST(NULLIF(total_spent,'') AS NUMERIC)` = NULL |
