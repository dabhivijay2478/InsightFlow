# 50-Pipeline Test Index

Each file = one pipeline. Every file covers:
- **Destination DDL** — column names intentionally differ from source
- **dbt SQL** — JSON key extraction, column aliases, type casts, derived columns
- **Step-by-step** — 8 numbered steps: DDL → connect → panel → normalisation → dbt → validate → run → verify
- **Edge cases** — nulls, empty strings, type mismatches, missing tables

---

## Part A — SaaS Sources (01–25)

| # | File | Source streams | Destinations |
|---|------|---------------|-------------|
| 01 | [01-stripe-to-postgres.md](saas/01-stripe-to-postgres.md) | 19 Stripe streams | PostgreSQL |
| 02 | [02-stripe-to-mysql.md](saas/02-stripe-to-mysql.md) | 19 Stripe streams | MySQL |
| 03 | [03-stripe-to-mariadb.md](saas/03-stripe-to-mariadb.md) | 19 Stripe streams | MariaDB |
| 04 | [04-stripe-to-sqlite.md](saas/04-stripe-to-sqlite.md) | 19 Stripe streams | SQLite |
| 05 | [05-stripe-to-cockroachdb.md](saas/05-stripe-to-cockroachdb.md) | 19 Stripe streams | CockroachDB |
| 06 | [06-shopify-to-postgres.md](saas/06-shopify-to-postgres.md) | 14 Shopify streams | PostgreSQL |
| 07 | [07-shopify-to-mysql.md](saas/07-shopify-to-mysql.md) | 14 Shopify streams | MySQL |
| 08 | [08-shopify-to-mariadb.md](saas/08-shopify-to-mariadb.md) | 14 Shopify streams | MariaDB |
| 09 | [09-shopify-to-sqlite.md](saas/09-shopify-to-sqlite.md) | 14 Shopify streams | SQLite |
| 10 | [10-shopify-to-cockroachdb.md](saas/10-shopify-to-cockroachdb.md) | 14 Shopify streams | CockroachDB |
| 11 | [11-hubspot-to-postgres.md](saas/11-hubspot-to-postgres.md) | 10 beta HubSpot streams | PostgreSQL |
| 12 | [12-hubspot-to-mysql.md](saas/12-hubspot-to-mysql.md) | Retired beta reference | Unsupported |
| 13 | [13-hubspot-to-mariadb.md](saas/13-hubspot-to-mariadb.md) | Retired beta reference | Unsupported |
| 14 | [14-hubspot-to-sqlite.md](saas/14-hubspot-to-sqlite.md) | Retired beta reference | Unsupported |
| 15 | [15-hubspot-to-cockroachdb.md](saas/15-hubspot-to-cockroachdb.md) | Retired beta reference | Unsupported |
| 16 | [16-github-to-postgres.md](saas/16-github-to-postgres.md) | 12 GitHub streams | PostgreSQL |
| 17 | [17-github-to-mysql.md](saas/17-github-to-mysql.md) | 12 GitHub streams | MySQL |
| 18 | [18-github-to-mariadb.md](saas/18-github-to-mariadb.md) | 12 GitHub streams | MariaDB |
| 19 | [19-github-to-sqlite.md](saas/19-github-to-sqlite.md) | 12 GitHub streams | SQLite |
| 20 | [20-github-to-cockroachdb.md](saas/20-github-to-cockroachdb.md) | 12 GitHub streams | CockroachDB |
| 21 | [21-notion-to-postgres.md](saas/21-notion-to-postgres.md) | 3 Notion streams | PostgreSQL |
| 22 | [22-notion-to-mysql.md](saas/22-notion-to-mysql.md) | 3 Notion streams | MySQL |
| 23 | [23-notion-to-mariadb.md](saas/23-notion-to-mariadb.md) | 3 Notion streams | MariaDB |
| 24 | [24-notion-to-sqlite.md](saas/24-notion-to-sqlite.md) | 3 Notion streams | SQLite |
| 25 | [25-notion-to-cockroachdb.md](saas/25-notion-to-cockroachdb.md) | 3 Notion streams | CockroachDB |

---

## Part B — DB Sources (26–50)

| # | File | Source streams | Destinations |
|---|------|---------------|-------------|
| 26 | [26-postgres-to-postgres.md](db/26-postgres-to-postgres.md) | 3 PG streams | PostgreSQL |
| 27 | [27-postgres-to-mysql.md](db/27-postgres-to-mysql.md) | 3 PG streams | MySQL |
| 28 | [28-postgres-to-mariadb.md](db/28-postgres-to-mariadb.md) | 3 PG streams | MariaDB |
| 29 | [29-postgres-to-sqlite.md](db/29-postgres-to-sqlite.md) | 3 PG streams | SQLite |
| 30 | [30-postgres-to-cockroachdb.md](db/30-postgres-to-cockroachdb.md) | 3 PG streams | CockroachDB |
| 31 | [31-mysql-to-postgres.md](db/31-mysql-to-postgres.md) | 3 MySQL streams | PostgreSQL |
| 32 | [32-mysql-to-mysql.md](db/32-mysql-to-mysql.md) | 3 MySQL streams | MySQL |
| 33 | [33-mysql-to-mariadb.md](db/33-mysql-to-mariadb.md) | 3 MySQL streams | MariaDB |
| 34 | [34-mysql-to-sqlite.md](db/34-mysql-to-sqlite.md) | 3 MySQL streams | SQLite |
| 35 | [35-mysql-to-cockroachdb.md](db/35-mysql-to-cockroachdb.md) | 3 MySQL streams | CockroachDB |
| 36 | [36-mariadb-to-postgres.md](db/36-mariadb-to-postgres.md) | 3 MariaDB streams | PostgreSQL |
| 37 | [37-mariadb-to-mysql.md](db/37-mariadb-to-mysql.md) | 3 MariaDB streams | MySQL |
| 38 | [38-mariadb-to-mariadb.md](db/38-mariadb-to-mariadb.md) | 3 MariaDB streams | MariaDB |
| 39 | [39-mariadb-to-sqlite.md](db/39-mariadb-to-sqlite.md) | 3 MariaDB streams | SQLite |
| 40 | [40-mariadb-to-cockroachdb.md](db/40-mariadb-to-cockroachdb.md) | 3 MariaDB streams | CockroachDB |
| 41 | [41-sqlite-to-postgres.md](db/41-sqlite-to-postgres.md) | 3 SQLite streams | PostgreSQL |
| 42 | [42-sqlite-to-mysql.md](db/42-sqlite-to-mysql.md) | 3 SQLite streams | MySQL |
| 43 | [43-sqlite-to-mariadb.md](db/43-sqlite-to-mariadb.md) | 3 SQLite streams | MariaDB |
| 44 | [44-sqlite-to-sqlite.md](db/44-sqlite-to-sqlite.md) | 3 SQLite streams | SQLite |
| 45 | [45-sqlite-to-cockroachdb.md](db/45-sqlite-to-cockroachdb.md) | 3 SQLite streams | CockroachDB |
| 46 | [46-cockroachdb-to-postgres.md](db/46-cockroachdb-to-postgres.md) | 3 CRDB streams | PostgreSQL |
| 47 | [47-cockroachdb-to-mysql.md](db/47-cockroachdb-to-mysql.md) | 3 CRDB streams | MySQL |
| 48 | [48-cockroachdb-to-mariadb.md](db/48-cockroachdb-to-mariadb.md) | 3 CRDB streams | MariaDB |
| 49 | [49-cockroachdb-to-sqlite.md](db/49-cockroachdb-to-sqlite.md) | 3 CRDB streams | SQLite |
| 50 | [50-cockroachdb-to-cockroachdb.md](db/50-cockroachdb-to-cockroachdb.md) | 3 CRDB streams | CockroachDB |

---

## Column Mapping Patterns Covered in Every File

| Pattern | Example |
|---------|---------|
| **JSON → flat TEXT** | `payment_method_details->'card'->>'brand'` → `card_brand TEXT` |
| **Column name mapping** | source `id` → dest `customer_id`; source `name` → dest `product_name` |
| **Type coercion** | Unix epoch `INT` → `TIMESTAMPTZ`; cents `INT` → `NUMERIC(10,2)`; `TINYINT(1)` → `BOOLEAN` |
| **Different column sets** | Source 19 cols → dest 7 cols; derived columns added (`is_active`, `urgency`, `full_name`) |

---

## Reference Files (full dbt SQL per source)

| Source | Full dbt SQL reference |
|--------|----------------------|
| Stripe | `saas/01-stripe-to-postgres.md` |
| Shopify | `saas/06-shopify-to-postgres.md` |
| HubSpot | `saas/11-hubspot-to-postgres.md` |
| GitHub | `saas/16-github-to-postgres.md` |
| Notion | `saas/21-notion-to-postgres.md` |
| PostgreSQL src | `db/26-postgres-to-postgres.md` |
| MySQL src | `db/31-mysql-to-postgres.md` |
| MariaDB src | `db/36-mariadb-to-postgres.md` |
| SQLite src | `db/41-sqlite-to-postgres.md` |
| CockroachDB src | `db/46-cockroachdb-to-postgres.md` |
