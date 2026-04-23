# Shopify Source — Manual Testing Guide

**Streams:** 14  
**Credential:** Shop domain + access token  
**DuckDB prefix:** `shopify__`

---

## Credential Setup

```json
{
  "shop": "your-shop.myshopify.com",
  "access_token": "shpat_..."
}
```

---

## All 14 Streams

| Stream | DuckDB staging name | Key columns | INCREMENTAL key |
|--------|-------------------|-------------|----------------|
| `shopify.products` | `shopify__products` | `id`, `title`, `vendor`, `product_type`, `status`, `published_at`, `created_at`, `updated_at` | `updated_at` |
| `shopify.orders` | `shopify__orders` | `id`, `email`, `financial_status`, `fulfillment_status`, `total_price`, `currency`, `created_at` | `updated_at` |
| `shopify.customers` | `shopify__customers` | `id`, `email`, `first_name`, `last_name`, `orders_count`, `total_spent`, `created_at` | `updated_at` |
| `shopify.draft_orders` | `shopify__draft_orders` | `id`, `email`, `status`, `total_price`, `currency`, `created_at` | `updated_at` |
| `shopify.custom_collections` | `shopify__custom_collections` | `id`, `title`, `handle`, `published`, `updated_at` | `updated_at` |
| `shopify.smart_collections` | `shopify__smart_collections` | `id`, `title`, `handle`, `published`, `updated_at` | `updated_at` |
| `shopify.pages` | `shopify__pages` | `id`, `title`, `handle`, `body_html`, `published_at`, `created_at`, `updated_at` | `updated_at` |
| `shopify.blogs` | `shopify__blogs` | `id`, `title`, `handle`, `commentable`, `created_at`, `updated_at` | `updated_at` |
| `shopify.articles` | `shopify__articles` | `id`, `title`, `blog_id`, `author`, `published_at`, `created_at`, `updated_at` | `updated_at` |
| `shopify.locations` | `shopify__locations` | `id`, `name`, `address1`, `city`, `country`, `active` | — |
| `shopify.price_rules` | `shopify__price_rules` | `id`, `title`, `value_type`, `value`, `starts_at`, `ends_at`, `created_at` | `updated_at` |
| `shopify.themes` | `shopify__themes` | `id`, `name`, `role`, `created_at`, `updated_at` | `updated_at` |
| `shopify.countries` | `shopify__countries` | `id`, `name`, `code`, `tax`, `tax_name` | — |
| `shopify.collects` | `shopify__collects` | `id`, `collection_id`, `product_id`, `position`, `created_at` | — |

---

## Scenario S-SHP-1 — Full Table Sync: `products`

**Destination DDL:**
```sql
CREATE TABLE analytics.shopify_products_hd (
    id           BIGINT PRIMARY KEY,
    title        TEXT,
    vendor       TEXT,
    product_type TEXT,
    status       TEXT,
    created_at   TIMESTAMPTZ
);
```

**dbt SQL:**
```sql
SELECT
    id,
    title,
    vendor,
    product_type,
    status,
    created_at
FROM {{ source('raw', 'shopify__products') }}
```

**Verify:** `SELECT COUNT(*) FROM analytics.shopify_products_hd;` → row count matches Shopify admin product count.

---

## Scenario S-SHP-2 — Incremental Sync: `orders`

**Source panel:** `shopify.orders`, sync mode `INCREMENTAL`, replication key `updated_at`

**dbt SQL:**
```sql
SELECT
    id,
    email,
    financial_status,
    fulfillment_status,
    total_price,
    currency,
    created_at
FROM {{ source('raw', 'shopify__orders') }}
WHERE financial_status = 'paid'
```

**Run 1:** All paid orders synced.  
**Run 2:** Only orders updated since last run.

---

## Scenario S-SHP-3 — Customer Lifetime Value Aggregate

**dbt SQL:**
```sql
SELECT
    id              AS customer_id,
    email,
    first_name,
    last_name,
    orders_count,
    CAST(total_spent AS NUMERIC) AS lifetime_value,
    created_at
FROM {{ source('raw', 'shopify__customers') }}
WHERE orders_count > 0
ORDER BY lifetime_value DESC
```

**Verify:** `lifetime_value` is numeric; only customers with at least one order.

---

## Scenario S-SHP-4 — Multi-Stream: `products` + `orders` + `customers`

Three streams in one pipeline, three separate dbt SQL models:

**Model 1:**
```sql
SELECT id, title, status, vendor FROM {{ source('raw', 'shopify__products') }}
```

**Model 2:**
```sql
SELECT id, email, total_price, currency, created_at FROM {{ source('raw', 'shopify__orders') }}
```

**Model 3:**
```sql
SELECT id, email, orders_count, total_spent FROM {{ source('raw', 'shopify__customers') }}
```

**Expected:** `dbt_models_run = 3`, three chips in delivery row.

---

## Scenario S-SHP-5 — Normalisation: Rename `total_price` → `order_total`

**Rule:**
```json
{ "rule_type": "rename", "table": "shopify.orders", "column": "total_price", "destination_name": "order_total" }
```

**dbt SQL:**
```sql
SELECT id, email, order_total, currency FROM {{ source('raw', 'shopify__orders') }}
```

---

## Scenario S-SHP-6 — Normalisation: Exclude `body_html` from pages

**Rule:**
```json
{ "rule_type": "exclude", "table": "shopify.pages", "column": "body_html" }
```

**dbt SQL:**
```sql
SELECT id, title, handle, published_at FROM {{ source('raw', 'shopify__pages') }}
```

**Verify:** `body_html` absent from destination.

---

## Scenario S-SHP-7 — Draft Orders Status Filter

**dbt SQL:**
```sql
SELECT
    id,
    email,
    status,
    CAST(total_price AS NUMERIC) AS total,
    currency,
    created_at
FROM {{ source('raw', 'shopify__draft_orders') }}
WHERE status = 'open'
```

---

## Scenario S-SHP-8 — Collections (both custom + smart in same pipeline)

Two streams: `shopify.custom_collections` + `shopify.smart_collections`

**Model 1:**
```sql
SELECT id, title, handle, published, updated_at
FROM {{ source('raw', 'shopify__custom_collections') }}
```

**Model 2:**
```sql
SELECT id, title, handle, published, updated_at
FROM {{ source('raw', 'shopify__smart_collections') }}
```

---

## All 14 Streams — Quick Smoke Test Checklist

| Stream | DuckDB ref | Expected rows |
|--------|-----------|---------------|
| products | `shopify__products` | ≥ 1 |
| orders | `shopify__orders` | ≥ 1 |
| customers | `shopify__customers` | ≥ 1 |
| draft_orders | `shopify__draft_orders` | ≥ 0 |
| custom_collections | `shopify__custom_collections` | ≥ 0 |
| smart_collections | `shopify__smart_collections` | ≥ 0 |
| pages | `shopify__pages` | ≥ 0 |
| blogs | `shopify__blogs` | ≥ 0 |
| articles | `shopify__articles` | ≥ 0 |
| locations | `shopify__locations` | ≥ 1 |
| price_rules | `shopify__price_rules` | ≥ 0 |
| themes | `shopify__themes` | ≥ 1 |
| countries | `shopify__countries` | ≥ 1 |
| collects | `shopify__collects` | ≥ 0 |
