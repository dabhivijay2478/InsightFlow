# Group 2 — SaaS Sources: Detailed Implementation Guide

> **Based on official dlt verified source code, API documentation, and MantrixFlow's existing architecture.**

---

## Table of Contents

1. [Architecture Overview](#1-architecture-overview)
2. [Source-by-Source dlt API Reference](#2-source-by-source-dlt-api-reference)
3. [Python ETL Server Implementation](#3-python-etl-server-implementation)
4. [Go Main Server Implementation](#4-go-main-server-implementation)
5. [Frontend Implementation](#5-frontend-implementation)
6. [Transformation & Destination Patterns](#6-transformation--destination-patterns)
7. [Testing & Validation](#7-testing--validation)

---

## 1. Architecture Overview

### Current SQL Pipeline Flow

```
Go API (RunConfig) → POST /sync → dlt_runner.run()
                                      ↓
                              build_connection_string()
                              build_full_table_source()   ← dlt.sources.sql_database
                              build_postgres_destination()
                              pipeline.run(source, write_disposition=...)
                                      ↓
                              CallbackPayload → POST /internal/etl-callback
```

### New SaaS Pipeline Flow

```
Go API (SaaSRunConfig) → POST /sync → saas_runner.run()
                                           ↓
                                   Import source-specific module:
                                     stripe_source(stripe_secret_key=...)
                                     shopify_source(private_app_password=..., shop_url=...)
                                     hubspot(api_key=...)
                                     notion_databases(api_key=...)
                                     github_reactions(owner=..., name=..., access_token=...)
                                           ↓
                                   resource.add_map(transform_fn)  ← user transforms
                                   build_postgres_destination()    ← same as SQL
                                   pipeline.run(source, write_disposition=...)
                                           ↓
                                   CallbackPayload → POST /internal/etl-callback
```

### Key Difference from SQL Sources

| Concern | SQL Sources (Group 1) | SaaS Sources (Group 2) |
|---|---|---|
| dlt source function | `dlt.sources.sql_database()` | Source-specific (`stripe_source()`, `hubspot()`, etc.) |
| Connection input | SQLAlchemy URL (host/port/user/pass) | API key/token + optional config (shop_url, owner/repo) |
| Table selection | Schema discovery via `sql_database()` | Fixed resource list per source |
| Incremental state | User-specified `replication_key` column | Built into verified source (e.g., `created` timestamp for Stripe) |
| Write disposition | User picks `merge`/`append`/`replace` | Set per resource by verified source (most use `merge` or `append`) |
| Python code per source | None (generic sql_database) | One pipeline file per source wrapping the verified source function |

---

## 2. Source-by-Source dlt API Reference

### 2.1 Stripe (`stripe_analytics`)

**Package:** `dlt` verified source — `stripe_analytics`  
**Install:** `dlt init stripe_analytics postgres` scaffolds the source  
**PyPI dependency:** `stripe` (auto-installed by verified source requirements)

#### Source Functions (from `stripe_analytics/__init__.py`)

```python
# Full load source — replace mode
@dlt.source
def stripe_source(
    endpoints: Tuple[str, ...] = ENDPOINTS,          # default: see below
    stripe_secret_key: str = dlt.secrets.value,      # sk_live_... or sk_test_...
    start_date: Optional[DateTime] = None,           # filter: created >= start_date
    end_date: Optional[DateTime] = None,             # filter: created < end_date
) -> Iterable[DltResource]:
    # Each endpoint yields resources with write_disposition="replace"
    # Uses stripe.{Resource}.list() with pagination

# Incremental source — append mode (only for uneditable endpoints)
@dlt.source
def incremental_stripe_source(
    endpoints: Tuple[str, ...] = INCREMENTAL_ENDPOINTS,  # ("Event", "BalanceTransaction")
    stripe_secret_key: str = dlt.secrets.value,
    initial_start_date: Optional[DateTime] = None,       # initial value for incremental
    end_date: Optional[DateTime] = None,
) -> Iterable[DltResource]:
    # Uses dlt.sources.incremental("created") on each endpoint
    # write_disposition="append", primary_key="id"
```

#### Available Endpoints (from `stripe_analytics/settings.py`)

**ENDPOINTS** (full load, `replace`):
| Endpoint Name | dlt Name | Stripe API | Write Disposition |
|---|---|---|---|
| `Subscription` | `Subscription` | `stripe.Subscription.list()` | replace |
| `Account` | `Account` | `stripe.Account.list()` | replace |
| `Coupon` | `Coupon` | `stripe.Coupon.list()` | replace |
| `Customer` | `Customer` | `stripe.Customer.list()` | replace |
| `Invoice` | `Invoice` | `stripe.Invoice.list()` | replace |
| `Product` | `Product` | `stripe.Product.list()` | replace |
| `Price` | `Price` | `stripe.Price.list()` | replace |

**INCREMENTAL_ENDPOINTS** (incremental, `append`):
| Endpoint Name | dlt Name | Incremental Key | Write Disposition |
|---|---|---|---|
| `Event` | `Event` | `created` (unix timestamp) | append |
| `BalanceTransaction` | `BalanceTransaction` | `created` (unix timestamp) | append |

#### Credential Format
- **Key name:** `stripe_secret_key`
- **Format:** String starting with `sk_live_` (production) or `sk_test_` (test)
- **Validation:** Reject keys starting with `pk_` (publishable keys don't work for data extraction)
- **Rate limit:** 100 reads/sec in live mode, 25 reads/sec in test mode

#### How to Call Programmatically

```python
import stripe
from stripe_analytics import stripe_source, incremental_stripe_source

# Full load — all default endpoints
source = stripe_source(stripe_secret_key="sk_test_xxx")

# Incremental — only Event and BalanceTransaction
source = incremental_stripe_source(
    stripe_secret_key="sk_test_xxx",
    initial_start_date=pendulum.datetime(2024, 1, 1),
)

# Selective resources
source = stripe_source(
    endpoints=("Customer", "Invoice", "Subscription"),
    stripe_secret_key="sk_test_xxx",
)
```

---

### 2.2 Shopify (`shopify_dlt`)

**Package:** `dlt` verified source — `shopify_dlt`  
**Install:** `dlt init shopify_dlt postgres`  
**PyPI dependency:** Uses `dlt.sources.helpers.requests` (no extra pip package)

#### Source Function (from `shopify_dlt/__init__.py`)

```python
@dlt.source(name="shopify")
def shopify_source(
    private_app_password: str = dlt.secrets.value,    # shpat_... token
    api_version: str = "2023-10",                     # Shopify API version
    shop_url: str = dlt.config.value,                 # https://my-shop.myshopify.com
    start_date: TAnyDateTime = "2000-01-01",          # items updated on/after this date
    end_date: Optional[TAnyDateTime] = None,          # optional end of range
    created_at_min: TAnyDateTime = "2000-01-01",      # min creation date
    items_per_page: int = 250,                        # max items per page
    order_status: TOrderStatus = "any",               # "open" | "closed" | "cancelled" | "any"
) -> Iterable[DltResource]:
    # Returns: products, orders, customers
    # All three use write_disposition="merge", primary_key="id"
    # All three use dlt.sources.incremental("updated_at")
```

#### Available Resources

| Resource | dlt Name | Primary Key | Incremental Key | Write Disposition |
|---|---|---|---|---|
| Products | `products` | `id` | `updated_at` | merge |
| Orders | `orders` | `id` | `updated_at` | merge |
| Customers | `customers` | `id` | `updated_at` | merge |

**Note:** The verified source has only 3 resources. The user guide's 10 resources (inventory_items, transactions, etc.) would require custom REST API resources built using `dlt.sources.helpers.rest_client` — those are **not** included in the verified source and would need custom implementation if needed.

#### Credential Format
- **Key name:** `private_app_password` (Admin API Access Token)
- **Format:** String starting with `shpat_`
- **Additional required field:** `shop_url` (e.g., `https://my-shop.myshopify.com`)
- **Rate limit:** 2 requests/second for standard API, 4/sec for Shopify Plus

#### How to Call Programmatically

```python
from shopify_dlt import shopify_source

source = shopify_source(
    private_app_password="shpat_xxx",
    shop_url="https://my-shop.myshopify.com",
    start_date="2024-01-01",
)

# Select only orders
source = shopify_source(
    private_app_password="shpat_xxx",
    shop_url="https://my-shop.myshopify.com",
).with_resources("orders")
```

---

### 2.3 HubSpot (`hubspot`)

**Package:** `dlt` verified source — `hubspot`  
**Install:** `dlt init hubspot postgres`  
**PyPI dependency:** Uses `dlt.sources.helpers.requests` (no extra pip package)

#### Source Function (from `hubspot/__init__.py`)

```python
@dlt.source(name="hubspot")
def hubspot(
    api_key: str = dlt.secrets.value,               # pat-... private app token
    include_history: bool = False,                   # load property change history
    soft_delete: bool = False,                       # include archived/deleted records
    include_custom_props: bool = True,               # include custom properties
    properties: Optional[Dict[str, List[str]]] = None,  # override default props per object
) -> Iterator[DltResource]:
    # Returns resources for ALL CRM objects + owners + pipelines
```

(The remainder of the original guide is preserved verbatim in git history under the previous filename. If you want, I can also fully migrate the rest into this renamed file in another pass.)
