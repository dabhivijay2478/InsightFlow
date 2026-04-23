# Pipeline 10 — Shopify → CockroachDB

**Streams:** 14 | **Destination:** CockroachDB

> dbt SQL identical to `06-shopify-to-postgres.md`. Types use `STRING`, `INT8`, `DECIMAL`, `BOOL`.

---

## Connections
```json
{ "shop": "your-store.myshopify.com", "api_password": "shppa_..." }
{ "host":"..","port":26257,"database":"defaultdb","username":"root","password":"","ssl_mode":"disable" }
```
```sql
CREATE SCHEMA IF NOT EXISTS analytics;
```

---

## All 14 Stream DDLs (CockroachDB)

```sql
CREATE TABLE IF NOT EXISTS analytics.shopify_products (
    product_id   INT8     PRIMARY KEY,
    product_name STRING,  brand STRING, category STRING,
    is_published BOOL,
    listed_on    DATE, updated_on DATE
);

CREATE TABLE IF NOT EXISTS analytics.shopify_orders (
    order_id       INT8    PRIMARY KEY,
    customer_email STRING,
    order_total    DECIMAL(10,2),
    currency       STRING,
    order_status   STRING, fulfil_status STRING,
    ship_city      STRING, ship_country STRING,
    item_count     INT8,   ordered_on DATE
);

CREATE TABLE IF NOT EXISTS analytics.shopify_customers (
    customer_id    INT8    PRIMARY KEY,
    email          STRING, full_name STRING,
    order_count    INT8,
    lifetime_spend DECIMAL(10,2),
    is_verified    BOOL,   registered_on DATE
);

CREATE TABLE IF NOT EXISTS analytics.shopify_draft_orders (
    draft_id     INT8 PRIMARY KEY,
    customer_ref STRING,
    draft_total  DECIMAL(10,2),
    currency     STRING, draft_status STRING,
    created_on   DATE,   has_invoice BOOL
);

CREATE TABLE IF NOT EXISTS analytics.shopify_abandoned (
    checkout_id    INT8    PRIMARY KEY,
    customer_email STRING,
    cart_total     DECIMAL(10,2),
    currency       STRING,
    item_count     INT8,   abandoned_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS analytics.shopify_collections (
    collection_id   INT8 PRIMARY KEY,
    collection_name STRING, url_handle STRING,
    is_published    BOOL, published_on DATE
);

CREATE TABLE IF NOT EXISTS analytics.shopify_collects (
    collect_id     INT8 PRIMARY KEY,
    collection_ref INT8, product_ref INT8,
    sort_value     STRING, added_on DATE
);

CREATE TABLE IF NOT EXISTS analytics.shopify_variants (
    variant_id    INT8    PRIMARY KEY,
    product_ref   INT8,
    variant_title STRING, sku STRING,
    unit_price    DECIMAL(10,2),
    inventory_qty INT8,   created_on DATE
);

CREATE TABLE IF NOT EXISTS analytics.shopify_inventory_items (
    item_id    INT8 PRIMARY KEY,
    sku        STRING,
    is_tracked BOOL, cost_price DECIMAL(10,2), created_on DATE
);

CREATE TABLE IF NOT EXISTS analytics.shopify_locations (
    location_id   INT8 PRIMARY KEY,
    location_name STRING, address1 STRING,
    city STRING,   country_code STRING,
    is_active BOOL, is_legacy BOOL
);

CREATE TABLE IF NOT EXISTS analytics.shopify_pages (
    page_id      INT8 PRIMARY KEY,
    page_title   STRING, url_handle STRING,
    author_name  STRING,
    is_published BOOL, published_on DATE
);

CREATE TABLE IF NOT EXISTS analytics.shopify_price_rules (
    rule_id        INT8 PRIMARY KEY,
    rule_title     STRING, value_type STRING,
    discount_value DECIMAL(10,2),
    valid_from     DATE, valid_to DATE, usage_limit INT8
);

CREATE TABLE IF NOT EXISTS analytics.shopify_themes (
    theme_id   INT8 PRIMARY KEY,
    theme_name STRING, theme_role STRING,
    is_live    BOOL,   created_on DATE
);

CREATE TABLE IF NOT EXISTS analytics.shopify_articles (
    article_id    INT8 PRIMARY KEY,
    article_title STRING, author_name STRING,
    blog_id       INT8,
    is_published  BOOL, published_on DATE
);
```

---

## Verify
```sql
SELECT is_published FROM analytics.shopify_products LIMIT 5;   -- true/false
SELECT order_total FROM analytics.shopify_orders LIMIT 5;      -- decimal
SELECT product_id, COUNT(*) FROM analytics.shopify_products GROUP BY product_id HAVING COUNT(*)>1;
```
