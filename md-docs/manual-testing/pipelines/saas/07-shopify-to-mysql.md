# Pipeline 07 — Shopify → MySQL

**Streams:** 14 | **Destination:** MySQL

> dbt SQL identical to `06-shopify-to-postgres.md`. Only destination DDL differs.

---

## Connections
```json
{ "shop": "your-store.myshopify.com", "api_password": "shppa_..." }
{ "host":"..","port":3306,"database":"analytics","username":"writer","password":"..." }
```

---

## All 14 Stream DDLs (MySQL)

```sql
CREATE TABLE analytics.shopify_products (
    product_id    BIGINT       PRIMARY KEY,
    product_name  VARCHAR(500),
    brand         VARCHAR(255),
    category      VARCHAR(255),
    is_published  TINYINT(1),
    listed_on     DATE,
    updated_on    DATE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE analytics.shopify_orders (
    order_id       BIGINT       PRIMARY KEY,
    customer_email VARCHAR(255),
    order_total    DECIMAL(10,2),
    currency       VARCHAR(10),
    order_status   VARCHAR(50),
    fulfil_status  VARCHAR(50),
    ship_city      VARCHAR(255),
    ship_country   VARCHAR(10),
    item_count     INT,
    ordered_on     DATE
) ENGINE=InnoDB;

CREATE TABLE analytics.shopify_customers (
    customer_id    BIGINT       PRIMARY KEY,
    email          VARCHAR(255),
    full_name      VARCHAR(500),
    order_count    INT,
    lifetime_spend DECIMAL(10,2),
    is_verified    TINYINT(1),
    registered_on  DATE
) ENGINE=InnoDB;

CREATE TABLE analytics.shopify_draft_orders (
    draft_id      BIGINT       PRIMARY KEY,
    customer_ref  VARCHAR(50),
    draft_total   DECIMAL(10,2),
    currency      VARCHAR(10),
    draft_status  VARCHAR(50),
    created_on    DATE,
    has_invoice   TINYINT(1)
) ENGINE=InnoDB;

CREATE TABLE analytics.shopify_abandoned (
    checkout_id    BIGINT       PRIMARY KEY,
    customer_email VARCHAR(255),
    cart_total     DECIMAL(10,2),
    currency       VARCHAR(10),
    item_count     INT,
    abandoned_at   DATETIME
) ENGINE=InnoDB;

CREATE TABLE analytics.shopify_collections (
    collection_id   BIGINT  PRIMARY KEY,
    collection_name VARCHAR(500),
    url_handle      VARCHAR(255),
    is_published    TINYINT(1),
    published_on    DATE
) ENGINE=InnoDB;

CREATE TABLE analytics.shopify_collects (
    collect_id      BIGINT PRIMARY KEY,
    collection_ref  BIGINT,
    product_ref     BIGINT,
    sort_value      VARCHAR(100),
    added_on        DATE
) ENGINE=InnoDB;

CREATE TABLE analytics.shopify_variants (
    variant_id    BIGINT       PRIMARY KEY,
    product_ref   BIGINT,
    variant_title VARCHAR(255),
    sku           VARCHAR(100),
    unit_price    DECIMAL(10,2),
    inventory_qty INT,
    created_on    DATE
) ENGINE=InnoDB;

CREATE TABLE analytics.shopify_inventory_items (
    item_id       BIGINT PRIMARY KEY,
    sku           VARCHAR(100),
    is_tracked    TINYINT(1),
    cost_price    DECIMAL(10,2),
    created_on    DATE
) ENGINE=InnoDB;

CREATE TABLE analytics.shopify_locations (
    location_id   BIGINT PRIMARY KEY,
    location_name VARCHAR(255),
    address1      VARCHAR(500),
    city          VARCHAR(255),
    country_code  VARCHAR(10),
    is_active     TINYINT(1),
    is_legacy     TINYINT(1)
) ENGINE=InnoDB;

CREATE TABLE analytics.shopify_pages (
    page_id      BIGINT PRIMARY KEY,
    page_title   VARCHAR(500),
    url_handle   VARCHAR(255),
    author_name  VARCHAR(255),
    is_published TINYINT(1),
    published_on DATE
) ENGINE=InnoDB;

CREATE TABLE analytics.shopify_price_rules (
    rule_id        BIGINT       PRIMARY KEY,
    rule_title     VARCHAR(500),
    value_type     VARCHAR(50),
    discount_value DECIMAL(10,2),
    valid_from     DATE,
    valid_to       DATE,
    usage_limit    INT
) ENGINE=InnoDB;

CREATE TABLE analytics.shopify_themes (
    theme_id    BIGINT PRIMARY KEY,
    theme_name  VARCHAR(255),
    theme_role  VARCHAR(50),
    is_live     TINYINT(1),
    created_on  DATE
) ENGINE=InnoDB;

CREATE TABLE analytics.shopify_articles (
    article_id    BIGINT PRIMARY KEY,
    article_title VARCHAR(500),
    author_name   VARCHAR(255),
    blog_id       BIGINT,
    is_published  TINYINT(1),
    published_on  DATE
) ENGINE=InnoDB;
```

---

## Verify
```sql
SELECT is_published, COUNT(*) FROM analytics.shopify_products GROUP BY is_published;
SELECT order_total, ship_city FROM analytics.shopify_orders LIMIT 5;
SELECT customer_id, COUNT(*) cnt FROM analytics.shopify_customers GROUP BY customer_id HAVING cnt>1;
```
