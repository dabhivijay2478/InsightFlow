# Group 2 — SaaS Sources: Detailed Implementation Guide

> **Based on official dlt verified source code, API documentation, and MantrixFlow's existing architecture.**

---

## Table of Contents

1. [Architecture Overview](#1-architecture-overview)
2. [Source-by-Source dlt API Reference](#2-source-by-source-dlt-api-reference)
3. [Python ETL Server Implementation](#3-python-etl-server-implementation)
4. [Go Arcyria Server Implementation](#4-go-arcyria-server-implementation)
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

#### Available Resources

**CRM Objects** (all use `write_disposition="merge"`, `primary_key="id"`):
| Resource | dlt Name | HubSpot API Endpoint | Incremental |
|---|---|---|---|
| Contacts | `contacts` | `/crm/v3/objects/contacts` | Yes (via HubSpot lastmodifieddate) |
| Companies | `companies` | `/crm/v3/objects/companies` | Yes |
| Deals | `deals` | `/crm/v3/objects/deals` | Yes |
| Tickets | `tickets` | `/crm/v3/objects/tickets` | Yes |
| Products | `products` | `/crm/v3/objects/products` | Yes |
| Quotes | `quotes` | `/crm/v3/objects/quotes` | Yes |

**Property History** (when `include_history=True`, `write_disposition="append"`):
| Resource | dlt Name |
|---|---|
| Contacts History | `contacts_property_history` |
| Companies History | `companies_property_history` |
| Deals History | `deals_property_history` |
| Tickets History | `tickets_property_history` |
| Products History | `products_property_history` |
| Quotes History | `quotes_property_history` |

**Other Resources**:
| Resource | dlt Name | Write Disposition |
|---|---|---|
| Owners | `owners` | merge |
| Deal Pipelines | `pipelines_deals` | merge |
| Ticket Pipelines | `pipelines_tickets` | merge |
| Deal Stage Timing | `stages_timing_deals` | merge |
| Ticket Stage Timing | `stages_timing_tickets` | merge |
| Properties Labels | `properties` | replace |

**Web Analytics** (standalone resource):
```python
@dlt.resource
def hubspot_events_for_objects(
    object_type: THubspotObjectType,  # "company" | "contact" | "deal" | "ticket" | "product" | "quote"
    object_ids: List[str],
    api_key: str = dlt.secrets.value,
    start_date: pendulum.DateTime = STARTDATE,
) -> DltResource:
    # Incremental via "occurredAt" field
    # write_disposition="append", primary_key="id"
```

#### Required HubSpot Scopes Per Resource

| Resource | Required Scope |
|---|---|
| contacts | `crm.objects.contacts.read` |
| companies | `crm.objects.companies.read` |
| deals | `crm.objects.deals.read` |
| tickets | `tickets` |
| products | `crm.objects.products.read` |
| quotes | `crm.quotes.read` |
| owners | `crm.objects.owners.read` |
| pipelines | `crm.objects.deals.read` (for deal pipelines) |

#### Credential Format
- **Key name:** `api_key` (Private App Token)
- **Format:** String starting with `pat-`
- **Rate limit:** 100 requests per 10 seconds (private apps), 150/10sec on Enterprise

#### How to Call Programmatically

```python
from hubspot import hubspot, hubspot_events_for_objects

# All CRM objects with history
source = hubspot(
    api_key="pat-xxx",
    include_history=True,
    include_custom_props=True,
)

# Only contacts and deals
source = hubspot(api_key="pat-xxx").with_resources("contacts", "deals")

# Custom properties
source = hubspot(
    api_key="pat-xxx",
    properties={
        "contact": ["email", "firstname", "lastname", "custom_field"],
        "deal": ["dealname", "amount", "closedate"],
    },
)
```

---

### 2.4 Notion (`notion`)

**Package:** `dlt` verified source — `notion`  
**Install:** `dlt init notion postgres`  
**PyPI dependency:** Uses `dlt.sources.helpers.requests` (no extra pip package)

#### Source Functions (from `notion/__init__.py`)

```python
@dlt.resource
def notion_pages(
    page_ids: Optional[List[str]] = None,   # specific page IDs, or None for all
    api_key: str = dlt.secrets.value,       # secret_... integration token
) -> Iterator[TDataItems]:
    # Fetches pages and their content blocks
    # No write_disposition set → defaults to "append"

@dlt.source
def notion_databases(
    database_ids: Optional[List[Dict[str, str]]] = None,  # [{"id": "...", "use_name": "..."}]
    api_key: str = dlt.secrets.value,
) -> Iterator[DltResource]:
    # If database_ids is None, discovers all databases shared with the integration
    # Each database becomes a separate resource named after the database title
    # write_disposition="replace", primary_key="id"
```

#### Available Resources

| Resource | dlt Function | Description | Write Disposition |
|---|---|---|---|
| Database rows | `notion_databases()` | Each Notion database → one dlt resource | replace |
| Pages (content blocks) | `notion_pages()` | Page content as block objects | append |

**Important Notion behavior:**
- `notion_databases()` auto-discovers all databases the integration has access to
- Each database produces a resource dynamically named after the database title
- Column names vary per workspace (Notion properties become columns)
- The integration must be explicitly shared with each page/database from Notion's UI

#### Credential Format
- **Key name:** `api_key` (Integration Token)
- **Format:** String starting with `secret_`
- **Rate limit:** 3 requests/second

#### How to Call Programmatically

```python
from notion import notion_databases, notion_pages

# All databases the integration can see
source = notion_databases(api_key="secret_xxx")

# Specific databases
source = notion_databases(
    database_ids=[
        {"id": "abc123", "use_name": "tasks"},
        {"id": "def456", "use_name": "projects"},
    ],
    api_key="secret_xxx",
)

# Pages with content blocks
pages = notion_pages(api_key="secret_xxx")
```

---

### 2.5 GitHub (`github`)

**Package:** `dlt` verified source — `github`  
**Install:** `dlt init github postgres`  
**PyPI dependency:** Uses `dlt.sources.helpers.requests` (no extra pip package)

#### Source Functions (from `github/__init__.py`)

```python
@dlt.source
def github_reactions(
    owner: str,                              # repo owner (e.g., "dlt-hub")
    name: str,                               # repo name (e.g., "dlt")
    access_token: str = dlt.secrets.value,   # GitHub PAT
    items_per_page: int = 100,
    max_items: Optional[int] = None,         # limit total items
) -> Sequence[DltResource]:
    # Returns: issues, pull_requests
    # write_disposition="replace" — full load each time
    # Uses GraphQL API for efficient data retrieval

@dlt.source(max_table_nesting=2)
def github_repo_events(
    owner: str,
    name: str,
    access_token: Optional[str] = None,     # optional for public repos
) -> DltResource:
    # Returns: repo_events (dispatched to table per event type)
    # write_disposition not set → defaults to "append"
    # Uses dlt.sources.incremental("created_at")
    # GitHub limits to 300 events for public repos

@dlt.source
def github_stargazers(
    owner: str,
    name: str,
    access_token: str = dlt.secrets.value,
    items_per_page: int = 100,
    max_items: Optional[int] = None,
) -> Sequence[DltResource]:
    # Returns: stargazers
    # write_disposition="replace"
```

#### Available Resources

| Resource | dlt Source Function | Includes | Write Disposition | API Type |
|---|---|---|---|---|
| Issues | `github_reactions()` | reactions, comments, comment reactions | replace | GraphQL |
| Pull Requests | `github_reactions()` | reactions, comments, comment reactions | replace | GraphQL |
| Repo Events | `github_repo_events()` | all event types (push, star, fork, etc.) | append (incremental) | REST |
| Stargazers | `github_stargazers()` | user + starred date | replace | GraphQL |

**Child tables generated by dlt:**
- `issues__reactions`
- `issues__comments`
- `issues__comments__reactions`
- `issues__labels`
- `issues__assignees`
- `pull_requests__reactions`
- `pull_requests__comments`
- `pull_requests__reviews` (if included in query)

#### Credential Format
- **Key name:** `access_token` (Personal Access Token)
- **Format:** Classic PAT (`ghp_...`) or fine-grained PAT (`github_pat_...`)
- **Additional required fields:** `owner` (string), `name` (string) — the `owner/repo` format
- **Required scopes:** `repo` (or `public_repo` for public repos), `read:org` for team data
- **Rate limit:** 5,000 requests/hour (authenticated), 60/hour (unauthenticated)

#### How to Call Programmatically

```python
from github import github_reactions, github_repo_events, github_stargazers

# Issues + PRs with reactions
source = github_reactions(
    owner="dlt-hub",
    name="dlt",
    access_token="ghp_xxx",
)

# Just issues
source = github_reactions(
    owner="dlt-hub",
    name="dlt",
    access_token="ghp_xxx",
).with_resources("issues")

# Repo events (incremental)
source = github_repo_events(
    owner="dlt-hub",
    name="dlt",
    access_token="ghp_xxx",
)
```

---

## 3. Python ETL Server Implementation

### 3.1 New Dependencies (`requirements.txt`)

Add to `apps/server/etl-server/requirements.txt`:

```
# SaaS source dependencies
stripe>=7.0.0
```

**Note:** Shopify, HubSpot, Notion, and GitHub verified sources use only `dlt.sources.helpers.requests` — no extra pip packages needed beyond `dlt` itself. The `stripe` package is the only additional dependency because the Stripe verified source calls `stripe.{Resource}.list()` directly.

### 3.2 Scaffolded Source Files

Create directory: `apps/server/etl-server/saas_sources/`

#### File Structure

```
apps/server/etl-server/
├── saas_sources/
│   ├── __init__.py
│   ├── stripe_analytics/        ← from dlt init stripe_analytics postgres
│   │   ├── __init__.py          ← stripe_source(), incremental_stripe_source()
│   │   ├── helpers.py           ← pagination(), stripe_get_data()
│   │   └── settings.py          ← ENDPOINTS, INCREMENTAL_ENDPOINTS
│   ├── shopify_dlt/             ← from dlt init shopify_dlt postgres
│   │   ├── __init__.py          ← shopify_source()
│   │   ├── helpers.py           ← ShopifyApi client
│   │   ├── settings.py          ← API version, pagination defaults
│   │   └── exceptions.py
│   ├── hubspot/                 ← from dlt init hubspot postgres
│   │   ├── __init__.py          ← hubspot(), hubspot_events_for_objects()
│   │   ├── helpers.py           ← fetch_data(), pagination()
│   │   ├── settings.py          ← CRM endpoints, default properties
│   │   └── utils.py             ← chunk_properties()
│   ├── notion/                  ← from dlt init notion postgres
│   │   ├── __init__.py          ← notion_databases(), notion_pages()
│   │   ├── settings.py          ← API_URL
│   │   └── helpers/
│   │       ├── __init__.py
│   │       ├── client.py        ← NotionClient
│   │       └── database.py      ← NotionDatabase
│   └── github/                  ← from dlt init github postgres
│       ├── __init__.py          ← github_reactions(), github_repo_events(), github_stargazers()
│       ├── helpers.py           ← GraphQL + REST API clients
│       ├── queries.py           ← GraphQL query strings
│       └── settings.py          ← API base URLs
```

#### Scaffold Commands

```bash
cd apps/server/etl-server

# Create temp directory for scaffolding
mkdir -p /tmp/dlt_scaffold && cd /tmp/dlt_scaffold

# Scaffold each source (use postgres as destination)
dlt init stripe_analytics postgres
dlt init shopify_dlt postgres
dlt init hubspot postgres
dlt init notion postgres
dlt init github postgres

# Copy source directories into the ETL server
cp -r stripe_analytics/ /path/to/apps/server/etl-server/saas_sources/
cp -r shopify_dlt/      /path/to/apps/server/etl-server/saas_sources/
cp -r hubspot/          /path/to/apps/server/etl-server/saas_sources/
cp -r notion/           /path/to/apps/server/etl-server/saas_sources/
cp -r github/           /path/to/apps/server/etl-server/saas_sources/

# Cleanup
rm -rf /tmp/dlt_scaffold
```

**Important:** After copying, you do NOT need to modify the verified source code itself. All credential injection happens at call time via function arguments.

### 3.3 SaaS Run Config Model

Create `apps/server/etl-server/models/saas_run_config.py`:

```python
"""SaaSRunConfig — POST /sync request body for SaaS/API sources."""

from __future__ import annotations
from typing import Optional
from pydantic import BaseModel


class SaaSRunConfig(BaseModel):
    """Configuration for a SaaS source pipeline sync run."""

    run_id: str
    pipeline_id: str
    org_id: str
    source_type: str              # "stripe" | "shopify" | "hubspot" | "notion" | "github"

    # Credentials (decrypted by Go server before sending)
    credential: str               # API key / token (never logged)

    # Source-specific config
    shop_url: str | None = None          # Shopify only
    github_owner: str | None = None      # GitHub only
    github_repo: str | None = None       # GitHub only
    notion_database_ids: list[dict] | None = None  # Notion only: [{"id": "...", "use_name": "..."}]

    # Resource selection
    selected_resources: list[str]        # which resources to sync

    # Sync mode
    sync_mode: str = "incremental"       # "full_refresh" | "incremental"

    # Destination credentials (same as SQL — already decrypted)
    dest_host: str
    dest_port: int = 5432
    dest_user: str
    dest_password: str
    dest_database: str
    dest_schema: str = "public"
    dest_ssl_mode: str = "require"

    # Transform (same as SQL sources)
    transform_script: str | None = None
    on_transform_error: str = "fail"
    column_map: dict[str, str] | None = None
    drop_columns: list[str] | None = None

    # Configuration
    emit_method: str = "merge"
    cleanup_dlt_artifacts: bool = False

    # Optional override for callback URL
    callback_url: str | None = None
```

### 3.4 SaaS Pipeline Runner

Create `apps/server/etl-server/runner/saas_runner.py`:

```python
"""SaaS source pipeline runner — one function per source type."""

from __future__ import annotations

import logging
import tempfile
import time
from pathlib import Path
from typing import Any

import dlt

from models.callback_payload import CallbackPayload
from models.saas_run_config import SaaSRunConfig
from runner.connection_builder import build_connection_string
from runner.destination_builder import build_postgres_destination
from runner.metrics_extractor import extract_metrics
from runner.transform_handler import compile_transform, make_add_map_fn
from core.transform_utils import apply_record_transform

logger = logging.getLogger("etl.saas_runner")

_TMPFS_ROOT = Path(tempfile.gettempdir())
_TMPFS_PREFIX = "mxf_saas_"

# -------------------------------------------------------------------
# Source builders: each returns a dlt source/resource for the given type
# -------------------------------------------------------------------


def _build_stripe_source(config: SaaSRunConfig) -> Any:
    """Build Stripe dlt source from config."""
    from saas_sources.stripe_analytics import stripe_source, incremental_stripe_source
    from saas_sources.stripe_analytics.settings import ENDPOINTS, INCREMENTAL_ENDPOINTS

    all_endpoints = set(ENDPOINTS) | set(INCREMENTAL_ENDPOINTS)
    selected = tuple(r for r in config.selected_resources if r in all_endpoints)

    if config.sync_mode == "full_refresh":
        return stripe_source(
            endpoints=selected or ENDPOINTS,
            stripe_secret_key=config.credential,
        )

    # For incremental, split endpoints into full-load and incremental
    inc_set = set(INCREMENTAL_ENDPOINTS)
    inc_endpoints = tuple(r for r in selected if r in inc_set)
    full_endpoints = tuple(r for r in selected if r not in inc_set)

    sources = []
    if full_endpoints:
        sources.append(stripe_source(
            endpoints=full_endpoints,
            stripe_secret_key=config.credential,
        ))
    if inc_endpoints:
        sources.append(incremental_stripe_source(
            endpoints=inc_endpoints,
            stripe_secret_key=config.credential,
        ))

    return sources if len(sources) > 1 else sources[0] if sources else stripe_source(
        stripe_secret_key=config.credential
    )


def _build_shopify_source(config: SaaSRunConfig) -> Any:
    """Build Shopify dlt source from config."""
    from saas_sources.shopify_dlt import shopify_source

    if not config.shop_url:
        raise ValueError("shop_url is required for Shopify source")

    source = shopify_source(
        private_app_password=config.credential,
        shop_url=config.shop_url,
    )

    if config.selected_resources:
        source = source.with_resources(*config.selected_resources)

    return source


def _build_hubspot_source(config: SaaSRunConfig) -> Any:
    """Build HubSpot dlt source from config."""
    from saas_sources.hubspot import hubspot

    source = hubspot(
        api_key=config.credential,
        include_history=False,
        include_custom_props=True,
    )

    if config.selected_resources:
        source = source.with_resources(*config.selected_resources)

    return source


def _build_notion_source(config: SaaSRunConfig) -> Any:
    """Build Notion dlt source from config."""
    from saas_sources.notion import notion_databases

    return notion_databases(
        database_ids=config.notion_database_ids,
        api_key=config.credential,
    )


def _build_github_source(config: SaaSRunConfig) -> Any:
    """Build GitHub dlt source from config."""
    from saas_sources.github import github_reactions, github_repo_events, github_stargazers

    if not config.github_owner or not config.github_repo:
        raise ValueError("github_owner and github_repo are required for GitHub source")

    sources = []

    # Map selected resources to the correct source function
    reaction_resources = {"issues", "pull_requests"}
    event_resources = {"repo_events"}
    stargazer_resources = {"stargazers"}

    selected = set(config.selected_resources)

    if selected & reaction_resources:
        src = github_reactions(
            owner=config.github_owner,
            name=config.github_repo,
            access_token=config.credential,
        )
        # Filter to only selected reaction resources
        wanted = selected & reaction_resources
        if wanted != reaction_resources:
            src = src.with_resources(*wanted)
        sources.append(src)

    if selected & event_resources:
        sources.append(github_repo_events(
            owner=config.github_owner,
            name=config.github_repo,
            access_token=config.credential,
        ))

    if selected & stargazer_resources:
        sources.append(github_stargazers(
            owner=config.github_owner,
            name=config.github_repo,
            access_token=config.credential,
        ))

    if not sources:
        # Default to issues + PRs
        sources.append(github_reactions(
            owner=config.github_owner,
            name=config.github_repo,
            access_token=config.credential,
        ))

    return sources if len(sources) > 1 else sources[0]


# Registry of source builders
_SOURCE_BUILDERS = {
    "stripe": _build_stripe_source,
    "shopify": _build_shopify_source,
    "hubspot": _build_hubspot_source,
    "notion": _build_notion_source,
    "github": _build_github_source,
}

SUPPORTED_SAAS_TYPES = set(_SOURCE_BUILDERS.keys())


# -------------------------------------------------------------------
# Apply transforms — same pattern as dlt_runner.py
# -------------------------------------------------------------------

def _apply_transforms(source: Any, config: SaaSRunConfig, drop_counter: list) -> None:
    """Apply column map, drop columns, and user transform to all resources in source."""
    transform_fn = compile_transform(config.transform_script)

    # Column map / drop
    if config.column_map or config.drop_columns:
        drop_set = set(config.drop_columns or [])
        cmap = config.column_map

        def _col_map(item: Any) -> Any:
            if isinstance(item, dict):
                return apply_record_transform(item, column_map=cmap, drop_columns=drop_set)
            if isinstance(item, list):
                return [apply_record_transform(r, column_map=cmap, drop_columns=drop_set)
                        for r in item if isinstance(r, dict)]
            return item

        if hasattr(source, "resources"):
            for resource in source.resources.values():
                resource.add_map(_col_map)
                resource.apply_hints(columns={}, primary_key=[], schema_contract="evolve")
        elif hasattr(source, "add_map"):
            source.add_map(_col_map)

    # User transform script
    if transform_fn:
        add_map_fn = make_add_map_fn(transform_fn, config.on_transform_error, drop_counter)
        if hasattr(source, "resources"):
            for resource in source.resources.values():
                resource.add_map(add_map_fn)
                resource.add_filter(lambda item: item is not None)
                resource.apply_hints(columns={}, primary_key=[], schema_contract="evolve")
        elif hasattr(source, "add_map"):
            source.add_map(add_map_fn)
            source.add_filter(lambda item: item is not None)


# -------------------------------------------------------------------
# Main runner
# -------------------------------------------------------------------

async def run(config: SaaSRunConfig) -> CallbackPayload:
    """Execute a SaaS source pipeline run and return CallbackPayload."""
    started_at = time.time()
    drop_counter = [0]
    work_dir = _TMPFS_ROOT / f"{_TMPFS_PREFIX}{config.run_id}"

    try:
        work_dir.mkdir(parents=True, exist_ok=True)

        # 1. Build source
        builder = _SOURCE_BUILDERS.get(config.source_type)
        if not builder:
            raise ValueError(
                f"Unsupported SaaS source type '{config.source_type}'. "
                f"Supported: {', '.join(sorted(SUPPORTED_SAAS_TYPES))}"
            )
        source_or_sources = builder(config)

        # 2. Build destination
        dest_creds = {
            "host": config.dest_host,
            "port": config.dest_port,
            "user": config.dest_user,
            "password": config.dest_password,
            "database": config.dest_database,
            "ssl_mode": config.dest_ssl_mode,
        }
        destination = build_postgres_destination(dest_creds)

        # 3. Create pipeline
        pipeline = dlt.pipeline(
            pipeline_name=f"mxf_saas_{config.pipeline_id[:8]}",
            pipelines_dir=str(work_dir),
            destination=destination,
            dataset_name=config.dest_schema,
            dev_mode=False,
        )

        # 4. Run source(s)
        write_disposition = config.emit_method if config.sync_mode == "full_refresh" else None

        if isinstance(source_or_sources, list):
            for src in source_or_sources:
                _apply_transforms(src, config, drop_counter)
                run_kwargs = {}
                if write_disposition:
                    run_kwargs["write_disposition"] = "replace"
                pipeline.run(src, **run_kwargs)
        else:
            _apply_transforms(source_or_sources, config, drop_counter)
            run_kwargs = {}
            if write_disposition:
                run_kwargs["write_disposition"] = "replace"
            pipeline.run(source_or_sources, **run_kwargs)

        # 5. Extract metrics
        metrics = extract_metrics(pipeline, started_at, drop_counter)
        rows_written = metrics.get("rows_written", 0)

        return CallbackPayload(
            run_id=config.run_id,
            job_id=config.run_id,
            pipeline_id=config.pipeline_id,
            organization_id=config.org_id,
            status="completed",
            rows_upserted=rows_written,
            source_tool=f"dlt_saas_{config.source_type}",
            dest_tool="dlt_dest_postgres",
            **metrics,
        )

    except Exception as exc:
        duration = time.time() - started_at
        error_msg = str(exc)
        # Scrub credentials from error messages
        if config.credential and config.credential in error_msg:
            error_msg = error_msg.replace(config.credential, "***")
        logger.error("SaaS run %s failed: %s", config.run_id, error_msg, exc_info=True)
        return CallbackPayload(
            run_id=config.run_id,
            job_id=config.run_id,
            pipeline_id=config.pipeline_id,
            organization_id=config.org_id,
            status="failed",
            error_message=error_msg,
            duration_seconds=duration,
            source_tool=f"dlt_saas_{config.source_type}",
            dest_tool="dlt_dest_postgres",
        )

    finally:
        import shutil
        if work_dir.exists():
            shutil.rmtree(work_dir, ignore_errors=True)
```

### 3.5 Sync Route Update

Update `apps/server/etl-server/api/routes/sync.py` to route SaaS sources:

```python
# In the sync route handler, after detecting source_type:

from runner.saas_runner import SUPPORTED_SAAS_TYPES
from runner.saas_runner import run as saas_run
from models.saas_run_config import SaaSRunConfig

# In the dispatch logic:
if source_type in SUPPORTED_SAAS_TYPES:
    saas_config = SaaSRunConfig(**payload)
    result = await saas_run(saas_config)
else:
    # Existing SQL pipeline path
    config = RunConfig(**payload) if is_run_config else _legacy_to_run_config(payload)
    result = await run(config)
```

### 3.6 Connector Support Update

Update `apps/server/etl-server/core/connector_support.py`:

```python
# Add SaaS source types
SUPPORTED_SAAS_SOURCE_TYPES = {
    "stripe",
    "shopify",
    "hubspot",
    "notion",
    "github",
}

SAAS_SOURCE_TYPE_ALIASES = {
    "stripe_analytics": "stripe",
    "shopify_dlt": "shopify",
    "github_reactions": "github",
    "github_repo_events": "github",
    "notion_databases": "notion",
}

def normalize_saas_source_type(source_type: str | None = None) -> str:
    """Normalize SaaS source connector aliases."""
    normalized = (source_type or "").strip().lower()
    normalized = SAAS_SOURCE_TYPE_ALIASES.get(normalized, normalized)
    if normalized not in SUPPORTED_SAAS_SOURCE_TYPES:
        raise ValueError(
            f"Unsupported SaaS source type '{normalized}'. "
            f"Supported: {', '.join(sorted(SUPPORTED_SAAS_SOURCE_TYPES))}"
        )
    return normalized

def is_saas_source(source_type: str) -> bool:
    """Check if a source type is a SaaS/API source."""
    normalized = (source_type or "").strip().lower()
    normalized = SAAS_SOURCE_TYPE_ALIASES.get(normalized, normalized)
    return normalized in SUPPORTED_SAAS_SOURCE_TYPES
```

---

## 4. Go Arcyria Server Implementation

### 4.1 SaaS Source Registry

Create a new registry file in the Go server that defines each SaaS source's metadata, credential fields, and available resources:

```go
// internal/saas_registry/registry.go

package saas_registry

type AuthField struct {
    Name        string `json:"name"`
    Label       string `json:"label"`
    Placeholder string `json:"placeholder"`
    HelpText    string `json:"help_text"`
    Required    bool   `json:"required"`
    Secret      bool   `json:"secret"`
}

type ResourcePreview struct {
    Name                string   `json:"name"`
    Description         string   `json:"description"`
    Incremental         bool     `json:"incremental"`
    WriteDisposition    string   `json:"write_disposition"`   // "merge" | "append" | "replace"
    RecommendedFreq     string   `json:"recommended_frequency"`
    VolumeEstimate      string   `json:"volume_estimate"`     // "small" | "medium" | "large"
    KeyFields           []string `json:"key_fields"`
    ChildTables         []string `json:"child_tables,omitempty"`
    RelatedResources    []string `json:"related_resources,omitempty"`
    RequiredScope       string   `json:"required_scope,omitempty"` // HubSpot scopes
}

type SaaSSourceEntry struct {
    SourceType      string            `json:"source_type"`
    DisplayName     string            `json:"display_name"`
    Icon            string            `json:"icon"`
    Category        string            `json:"category"`
    AuthFields      []AuthField       `json:"auth_fields"`
    Resources       []ResourcePreview `json:"resources"`
    DocsURL         string            `json:"docs_url"`
    RateLimitNote   string            `json:"rate_limit_note"`
    CredentialHint  string            `json:"credential_hint"` // e.g. "Starts with sk_live_ or sk_test_"
    Notes           string            `json:"notes,omitempty"`
}
```

### 4.2 Resource Registry Data (Per Source)

The full registry should be populated with the data from Section 2. Here's a summary of what goes in:

**Stripe** — 9 resources (7 ENDPOINTS + 2 INCREMENTAL_ENDPOINTS), 1 auth field (stripe_secret_key)  
**Shopify** — 3 resources (products, orders, customers), 2 auth fields (private_app_password, shop_url)  
**HubSpot** — 12+ resources (6 CRM objects + history + owners + pipelines), 1 auth field (api_key)  
**Notion** — 2 resource types (databases, pages), 1 auth field (api_key)  
**GitHub** — 4 resource groups (issues, pull_requests, repo_events, stargazers), 2 auth fields (access_token, owner/repo)

### 4.3 Job Payload for SaaS Sources

The Go server should send a different payload shape for SaaS sources. The key fields:

```json
{
  "run_id": "uuid",
  "pipeline_id": "uuid",
  "org_id": "uuid",
  "source_type": "stripe",
  "credential": "sk_test_...",
  "selected_resources": ["Customer", "Invoice", "Subscription"],
  "sync_mode": "incremental",
  "dest_host": "...",
  "dest_port": 5432,
  "dest_user": "...",
  "dest_password": "...",
  "dest_database": "...",
  "dest_schema": "stripe_data",
  "transform_script": null,
  "emit_method": "merge",
  "callback_url": "https://api.example.com/api/v1/internal/etl-callback"
}
```

**Source-specific extra fields:**
- Shopify: `"shop_url": "https://my-shop.myshopify.com"`
- GitHub: `"github_owner": "dlt-hub"`, `"github_repo": "dlt"`
- Notion: `"notion_database_ids": [{"id": "...", "use_name": "tasks"}]`

### 4.4 Credential Validation (Basic Format Check)

Before saving/dispatching, validate the credential format:

```go
func ValidateSaaSCredential(sourceType, credential string) error {
    switch sourceType {
    case "stripe":
        if !strings.HasPrefix(credential, "sk_") {
            return fmt.Errorf("Stripe secret key must start with 'sk_live_' or 'sk_test_'")
        }
        if strings.HasPrefix(credential, "pk_") {
            return fmt.Errorf("publishable keys (pk_) cannot be used for data extraction")
        }
    case "shopify":
        if !strings.HasPrefix(credential, "shpat_") {
            return fmt.Errorf("Shopify Admin API token must start with 'shpat_'")
        }
    case "hubspot":
        if !strings.HasPrefix(credential, "pat-") {
            return fmt.Errorf("HubSpot private app token must start with 'pat-'")
        }
    case "notion":
        if !strings.HasPrefix(credential, "secret_") {
            return fmt.Errorf("Notion integration token must start with 'secret_'")
        }
    case "github":
        // Classic PATs start with ghp_, fine-grained with github_pat_
        // Also accept older tokens that don't have a prefix
        if credential == "" {
            return fmt.Errorf("GitHub personal access token is required")
        }
    }
    return nil
}
```

---

## 5. Frontend Implementation

### 5.1 Connector Catalog Updates

Add to `apps/arcyria-platform/app/workspace/connections/data/connectors.ts`:

```typescript
// Wave 2 — SaaS Sources
{
  id: "stripe",
  displayName: "Stripe",
  category: "saas",
  icon: "stripe",
  sourceCapable: true,
  destCapable: false,
  wave: 2,
  popular: true,
},
{
  id: "shopify",
  displayName: "Shopify",
  category: "saas",
  icon: "shopify",
  sourceCapable: true,
  destCapable: false,
  wave: 2,
},
{
  id: "hubspot",
  displayName: "HubSpot",
  category: "saas",
  icon: "hubspot",
  sourceCapable: true,
  destCapable: false,
  wave: 2,
},
{
  id: "notion",
  displayName: "Notion",
  category: "saas",
  icon: "notion",
  sourceCapable: true,
  destCapable: false,
  wave: 2,
},
{
  id: "github",
  displayName: "GitHub",
  category: "saas",
  icon: "github",
  sourceCapable: true,
  destCapable: false,
  wave: 2,
},
```

### 5.2 Connection Field Configs

Add to `apps/arcyria-platform/app/workspace/connections/data/connectionFields.ts`:

```typescript
// SaaS connection fields — API key/token only
{
  connectorId: "stripe",
  fields: [
    {
      name: "connectionName",
      label: "Connection Name",
      type: "text",
      placeholder: "Production Stripe",
      required: true,
      gridCol: "full",
    },
    {
      name: "credential",
      label: "Secret Key",
      type: "password",
      placeholder: "sk_live_...",
      required: true,
      gridCol: "full",
    },
  ],
},
{
  connectorId: "shopify",
  fields: [
    {
      name: "connectionName",
      label: "Connection Name",
      type: "text",
      placeholder: "My Shopify Store",
      required: true,
      gridCol: "full",
    },
    {
      name: "shopUrl",
      label: "Shop URL",
      type: "text",
      placeholder: "https://my-store.myshopify.com",
      required: true,
      gridCol: "full",
    },
    {
      name: "credential",
      label: "Admin API Access Token",
      type: "password",
      placeholder: "shpat_...",
      required: true,
      gridCol: "full",
    },
  ],
},
{
  connectorId: "hubspot",
  fields: [
    {
      name: "connectionName",
      label: "Connection Name",
      type: "text",
      placeholder: "Production HubSpot",
      required: true,
      gridCol: "full",
    },
    {
      name: "credential",
      label: "Private App Token",
      type: "password",
      placeholder: "pat-...",
      required: true,
      gridCol: "full",
    },
  ],
},
{
  connectorId: "notion",
  fields: [
    {
      name: "connectionName",
      label: "Connection Name",
      type: "text",
      placeholder: "Notion Workspace",
      required: true,
      gridCol: "full",
    },
    {
      name: "credential",
      label: "Integration Token",
      type: "password",
      placeholder: "secret_...",
      required: true,
      gridCol: "full",
    },
  ],
},
{
  connectorId: "github",
  fields: [
    {
      name: "connectionName",
      label: "Connection Name",
      type: "text",
      placeholder: "My GitHub Repo",
      required: true,
      gridCol: "full",
    },
    {
      name: "githubOwner",
      label: "Repository Owner",
      type: "text",
      placeholder: "octocat",
      required: true,
      gridCol: "half",
    },
    {
      name: "githubRepo",
      label: "Repository Name",
      type: "text",
      placeholder: "Hello-World",
      required: true,
      gridCol: "half",
    },
    {
      name: "credential",
      label: "Personal Access Token",
      type: "password",
      placeholder: "ghp_...",
      required: true,
      gridCol: "full",
    },
  ],
},
```

### 5.3 Resource Picker Component

The resource picker is a multi-select list where each item has:
- Checkbox for selection
- Resource name
- Incremental/Full Refresh badge
- Info icon that opens the data preview panel

This should be a new component: `apps/arcyria-platform/components/connections/resource-picker.tsx`

### 5.4 Data Preview Panel

The data preview panel drawer is a Sheet component that opens from the right side. It shows:
- Resource name + badges
- Description text
- Key fields table
- Related resources
- Volume estimate
- Child tables notice

All content is **static** — loaded from the Go server's SaaS registry (fetched once on connection setup) or hardcoded in the frontend.

---

## 6. Transformation & Destination Patterns

### 6.1 How Transforms Work with SaaS Sources

User transforms work **identically** for SaaS and SQL sources. The pattern is:

```python
# After building the source, before pipeline.run():
source_or_resource.add_map(transform_fn)       # row-by-row transformation
source_or_resource.add_filter(predicate_fn)     # row-by-row filtering
```

**Key dlt APIs for transforms:**

| Method | Purpose | Signature |
|---|---|---|
| `resource.add_map(fn)` | Transform each data item | `fn(item: dict) -> dict` |
| `resource.add_filter(fn)` | Keep only items where fn returns True | `fn(item: dict) -> bool` |
| `resource.add_yield_map(fn)` | One item → many items (fan-out) | `fn(item: dict) -> Iterator[dict]` |
| `resource.add_limit(n)` | Stop after n items (for preview) | `n: int` |

**Example: Anonymize Stripe customer emails before loading:**

```python
def anonymize_email(record):
    if "email" in record:
        record["email"] = hash(record["email"])
    return record

source = stripe_source(stripe_secret_key="sk_test_xxx")
for resource in source.resources.values():
    resource.add_map(anonymize_email)

pipeline.run(source, destination=postgres_dest)
```

**Example: Filter HubSpot contacts to only include those with email:**

```python
source = hubspot(api_key="pat-xxx").with_resources("contacts")
source.resources["contacts"].add_filter(lambda c: c.get("email") is not None)
```

### 6.2 Write Dispositions Per Source

Each verified source has its own default write dispositions per resource. The user CAN override at `pipeline.run()` time:

```python
# Default: uses each resource's built-in write_disposition
pipeline.run(source)

# Override: force full refresh (replace) for all resources
pipeline.run(source, write_disposition="replace")
```

Per-source defaults:

| Source | Resource | Default Write Disposition | Can Override? |
|---|---|---|---|
| Stripe | ENDPOINTS (Customer, Invoice, etc.) | `replace` | Yes |
| Stripe | INCREMENTAL_ENDPOINTS (Event, BalanceTransaction) | `append` | Yes |
| Shopify | products, orders, customers | `merge` | Yes |
| HubSpot | CRM objects (contacts, deals, etc.) | `merge` | Yes |
| HubSpot | Property history | `append` | Yes |
| Notion | Database rows | `replace` | Yes |
| GitHub | issues, pull_requests, stargazers | `replace` | Yes |
| GitHub | repo_events | `append` (incremental) | Yes |

### 6.3 Destination Configuration

All SaaS sources use the same PostgreSQL destination as SQL sources. The existing `build_postgres_destination()` function works unchanged:

```python
from runner.destination_builder import build_postgres_destination

dest_creds = {
    "host": config.dest_host,
    "port": config.dest_port,
    "user": config.dest_user,
    "password": config.dest_password,
    "database": config.dest_database,
    "ssl_mode": config.dest_ssl_mode,
}
destination = build_postgres_destination(dest_creds)
```

The destination handles:
- Schema creation (dataset_name → PostgreSQL schema)
- Table creation (auto-created from dlt resource names)
- Nested table flattening (e.g., `orders__line_items`)
- Type inference from API response shapes
- Merge/append/replace write dispositions

### 6.4 Schema Handling for SaaS Sources

Unlike SQL sources where the schema is known from `sql_database()` introspection, SaaS sources have **inferred schemas**. dlt handles this automatically:

1. **First run:** dlt creates tables and infers column types from the data
2. **Subsequent runs:** dlt evolves the schema if new fields appear (schema_contract="evolve")
3. **Nested objects:** Automatically flattened into child tables (configurable via `max_table_nesting`)

**Controlling nesting depth:**

```python
# Default: unlimited nesting (all nested arrays become child tables)
pipeline.run(source)

# Limit to 1 level of nesting
source = github_reactions(owner="x", name="y", access_token="z")
# github_reactions already sets max_table_nesting=2 for repo_events
```

**Schema contracts:**

```python
# Let schema evolve freely (new columns, new tables)
pipeline.run(source, schema_contract="evolve")

# Freeze schema — error if new columns appear
pipeline.run(source, schema_contract="freeze")
```

### 6.5 Destination-Side Table Names

dlt generates table names from resource names. For SaaS sources:

| Source | Resource | Destination Table | Child Tables |
|---|---|---|---|
| Stripe | `Customer` | `customer` | — |
| Stripe | `Invoice` | `invoice` | `invoice__lines` |
| Stripe | `Event` | `event` | `event__data` |
| Shopify | `orders` | `orders` | `orders__line_items` |
| Shopify | `products` | `products` | `products__variants` |
| HubSpot | `contacts` | `contacts` | `contacts__deals` (associations) |
| HubSpot | `deals` | `deals` | — |
| Notion | (database title) | `{database_title}` | varies by properties |
| GitHub | `issues` | `issues` | `issues__reactions`, `issues__comments`, `issues__labels` |
| GitHub | `pull_requests` | `pull_requests` | `pull_requests__reactions`, `pull_requests__comments` |

The user can override table names via `apply_hints`:

```python
source.resources["Customer"].apply_hints(table_name="stripe_customers")
```

Or via `stream_configs` in the payload (same as SQL sources):

```json
{
  "stream_configs": {
    "Customer": {"dest_table": "stripe_customers"}
  }
}
```

---

## 7. Testing & Validation

### 7.1 Test Checklist Per Source

For each SaaS source, verify:

| Test | What to Check |
|---|---|
| **Auth validation** | Go server rejects invalid credential format (e.g., `pk_` for Stripe) |
| **Auth error handling** | ETL server returns clear error for expired/invalid credentials (not generic failure) |
| **Resource selection** | Selecting subset of resources loads only those tables |
| **All resources deselected** | UI blocks save when no resources selected |
| **Incremental state** | Second run loads only new/updated records (where supported) |
| **Full refresh** | `sync_mode=full_refresh` replaces destination tables |
| **Schema evolution** | New fields in API responses create new columns (not errors) |
| **Transform compatibility** | `add_map()` transforms work on SaaS resource records |
| **Credential rotation** | Updating API key works on next run without re-setup |
| **Rate limit handling** | Verified source auto-retries on rate limit (all verified sources do this) |
| **Error propagation** | API errors surface in CallbackPayload.error_message |
| **Credential scrubbing** | API keys never appear in logs or error messages |

### 7.2 Rate Limits Per Source

| Source | Rate Limit | Backoff Strategy |
|---|---|---|
| Stripe | 100 reads/sec (live), 25/sec (test) | Stripe SDK auto-retries |
| Shopify | 2 req/sec (standard), 4/sec (Plus) | REST API link pagination respects limits |
| HubSpot | 100 req/10 sec (private apps) | dlt.sources.helpers.requests retries |
| Notion | 3 req/sec | Client retries on 429 |
| GitHub | 5,000 req/hour (authenticated) | GraphQL cost-based, REST respects x-ratelimit headers |

### 7.3 Manual Test Commands

```bash
# Test Stripe pipeline locally
cd apps/server/etl-server
python -c "
from saas_sources.stripe_analytics import stripe_source
import dlt
source = stripe_source(
    endpoints=('Customer',),
    stripe_secret_key='sk_test_xxx',
)
pipeline = dlt.pipeline(pipeline_name='test_stripe', destination='duckdb')
info = pipeline.run(source)
print(info)
"

# Test Shopify pipeline locally
python -c "
from saas_sources.shopify_dlt import shopify_source
import dlt
source = shopify_source(
    private_app_password='shpat_xxx',
    shop_url='https://test-store.myshopify.com',
).with_resources('products')
pipeline = dlt.pipeline(pipeline_name='test_shopify', destination='duckdb')
info = pipeline.run(source)
print(info)
"
```

---

## Appendix: File Change Summary

| File | Action | Description |
|---|---|---|
| `etl-server/requirements.txt` | Edit | Add `stripe>=7.0.0` |
| `etl-server/saas_sources/` | Create | Directory with 5 scaffolded verified source packages |
| `etl-server/models/saas_run_config.py` | Create | Pydantic model for SaaS sync requests |
| `etl-server/runner/saas_runner.py` | Create | SaaS pipeline runner with per-source builders |
| `etl-server/core/connector_support.py` | Edit | Add `SUPPORTED_SAAS_SOURCE_TYPES`, `is_saas_source()` |
| `etl-server/api/routes/sync.py` | Edit | Route SaaS source types to `saas_runner.run()` |
| `arcyria-server/internal/saas_registry/` | Create | Go source registry with resource previews |
| `arcyria-server/internal/server/` | Edit | Add SaaS connection form + job payload serializer |
| `app/components/connections/resource-picker.tsx` | Create | Multi-select resource picker component |
| `app/components/connections/data-preview-panel.tsx` | Create | Resource preview drawer component |
| `app/config/connectors.ts` | Edit | Add SaaS source schemas |
| `app/config/database-registry.ts` | Edit | Add SaaS entries (or separate `saas-registry.ts`) |
| `app/workspace/connections/data/connectors.ts` | Edit | Add 5 SaaS connector entries |
| `app/workspace/connections/data/connectionFields.ts` | Edit | Add 5 SaaS field configs |
