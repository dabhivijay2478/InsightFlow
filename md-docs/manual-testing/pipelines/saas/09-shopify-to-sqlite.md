# Pipeline 09 — Shopify → SQLite

**Streams:** 14 | **Destination:** SQLite

> dbt SQL identical to `06-shopify-to-postgres.md`. Types stored as TEXT/REAL/INTEGER.

---

## Connections
```json
{ "shop": "your-store.myshopify.com", "api_password": "shppa_..." }
{ "database": "/absolute/path/to/analytics.db" }
```

---

## All 14 Stream DDLs (SQLite)

```sql
CREATE TABLE IF NOT EXISTS shopify_products (
    product_id   INTEGER PRIMARY KEY,
    product_name TEXT, brand TEXT, category TEXT,
    is_published INTEGER,   -- 0/1
    listed_on    TEXT, updated_on TEXT
);

CREATE TABLE IF NOT EXISTS shopify_orders (
    order_id       INTEGER PRIMARY KEY,
    customer_email TEXT,
    order_total    REAL,
    currency       TEXT,
    order_status   TEXT, fulfil_status TEXT,
    ship_city      TEXT, ship_country TEXT,
    item_count     INTEGER, ordered_on TEXT
);

CREATE TABLE IF NOT EXISTS shopify_customers (
    customer_id    INTEGER PRIMARY KEY,
    email          TEXT, full_name TEXT,
    order_count    INTEGER,
    lifetime_spend REAL,
    is_verified    INTEGER, registered_on TEXT
);

CREATE TABLE IF NOT EXISTS shopify_draft_orders (
    draft_id     INTEGER PRIMARY KEY,
    customer_ref TEXT,
    draft_total  REAL,
    currency     TEXT, draft_status TEXT,
    created_on   TEXT, has_invoice INTEGER
);

CREATE TABLE IF NOT EXISTS shopify_abandoned (
    checkout_id    INTEGER PRIMARY KEY,
    customer_email TEXT,
    cart_total     REAL,
    currency       TEXT,
    item_count     INTEGER, abandoned_at TEXT
);

CREATE TABLE IF NOT EXISTS shopify_collections (
    collection_id   INTEGER PRIMARY KEY,
    collection_name TEXT, url_handle TEXT,
    is_published    INTEGER, published_on TEXT
);

CREATE TABLE IF NOT EXISTS shopify_collects (
    collect_id     INTEGER PRIMARY KEY,
    collection_ref INTEGER, product_ref INTEGER,
    sort_value     TEXT, added_on TEXT
);

CREATE TABLE IF NOT EXISTS shopify_variants (
    variant_id    INTEGER PRIMARY KEY,
    product_ref   INTEGER,
    variant_title TEXT, sku TEXT,
    unit_price    REAL,
    inventory_qty INTEGER, created_on TEXT
);

CREATE TABLE IF NOT EXISTS shopify_inventory_items (
    item_id    INTEGER PRIMARY KEY,
    sku        TEXT,
    is_tracked INTEGER, cost_price REAL, created_on TEXT
);

CREATE TABLE IF NOT EXISTS shopify_locations (
    location_id   INTEGER PRIMARY KEY,
    location_name TEXT, address1 TEXT,
    city TEXT, country_code TEXT,
    is_active INTEGER, is_legacy INTEGER
);

CREATE TABLE IF NOT EXISTS shopify_pages (
    page_id      INTEGER PRIMARY KEY,
    page_title   TEXT, url_handle TEXT,
    author_name  TEXT,
    is_published INTEGER, published_on TEXT
);

CREATE TABLE IF NOT EXISTS shopify_price_rules (
    rule_id        INTEGER PRIMARY KEY,
    rule_title     TEXT, value_type TEXT,
    discount_value REAL,
    valid_from     TEXT, valid_to TEXT,
    usage_limit    INTEGER
);

CREATE TABLE IF NOT EXISTS shopify_themes (
    theme_id   INTEGER PRIMARY KEY,
    theme_name TEXT, theme_role TEXT,
    is_live    INTEGER, created_on TEXT
);

CREATE TABLE IF NOT EXISTS shopify_articles (
    article_id    INTEGER PRIMARY KEY,
    article_title TEXT, author_name TEXT,
    blog_id       INTEGER,
    is_published  INTEGER, published_on TEXT
);
```

---

## Verify
```bash
DB=/absolute/path/to/analytics.db
sqlite3 $DB "SELECT typeof(order_total), order_total FROM shopify_orders LIMIT 3;"
# real|29.99
sqlite3 $DB "SELECT is_published FROM shopify_products LIMIT 5;"
# 0 or 1
sqlite3 $DB "SELECT abandoned_at FROM shopify_abandoned LIMIT 3;"
# ISO string
```
