# Pipeline E2E & Unit Tests — 50 Source × Destination Pipelines (All Streams)

Create Vitest test suites covering exactly 50 pipeline combos (25 SaaS→DB + 25 DB→DB) with **every available SaaS stream** and Docker DB containers, validating every source→destination pair end-to-end.

---

## All Available Streams Per Source (from `preview.py` + `saas_runner.py`)

### Stripe (19 streams)
`customers`, `charges`, `invoices`, `subscriptions`, `products`, `prices`, `events`, `balance_transactions`, `coupons`, `payment_intents`, `payment_methods`, `refunds`, `disputes`, `payouts`, `plans`, `tax_rates`, `credit_notes`, `promotion_codes`, `setup_intents`

DuckDB names → `stripe__customers`, `stripe__charges`, `stripe__invoices`, …

### Shopify (14 streams)
`products`, `orders`, `customers`, `draft_orders`, `custom_collections`, `smart_collections`, `pages`, `blogs`, `articles`, `locations`, `price_rules`, `themes`, `countries`, `collects`

DuckDB names → `shopify__products`, `shopify__orders`, `shopify__customers`, …

### HubSpot (10 production streams; PostgreSQL destination only)
`contacts`, `companies`, `deals`, `tickets`, `owners`, `deal_pipelines`, `ticket_pipelines`, `products`, `line_items`, `quotes`

DuckDB names → `hubspot__contacts`, `hubspot__companies`, `hubspot__deals`, …

### GitHub (12 streams)
`issues`, `pull_requests`, `stargazers`, `events`, `commits`, `branches`, `releases`, `tags`, `contributors`, `milestones`, `labels`, `forks`

DuckDB names → `github__issues`, `github__pull_requests`, `github__stargazers`, …

### Notion (3 streams)
`databases`, `pages`, `users`

DuckDB names → `notion__databases`, `notion__pages`, `notion__users`

### DB Sources (schema.table format)
- **PostgreSQL**: `public.users`, `public.orders`, `public.payments`
- **MySQL**: `mydb.products`, `mydb.inventory`, `mydb.categories`
- **MariaDB**: `app.events`, `app.logs`, `app.sessions`
- **SQLite**: `main.tasks`, `main.notes`, `main.tags`
- **CockroachDB**: `public.sessions`, `public.accounts`, `public.transactions`

DuckDB names → `public__users`, `mydb__products`, `app__events`, `main__tasks`, etc.

---

## The 50-Pipeline Matrix

### SaaS → DB (25 pipelines) — tests ALL streams per source

| # | Source | All Streams Tested | → Destination |
|---|--------|--------------------|---------------|
| 1 | Stripe | all 19 streams | → PostgreSQL |
| 2 | Stripe | all 19 streams | → MySQL |
| 3 | Stripe | all 19 streams | → CockroachDB |
| 4 | Stripe | all 19 streams | → MariaDB |
| 5 | Stripe | all 19 streams | → SQLite |
| 6 | Shopify | all 14 streams | → PostgreSQL |
| 7 | Shopify | all 14 streams | → MySQL |
| 8 | Shopify | all 14 streams | → CockroachDB |
| 9 | Shopify | all 14 streams | → MariaDB |
| 10 | Shopify | all 14 streams | → SQLite |
| 11 | HubSpot | all 10 production streams | → PostgreSQL |
| 12 | HubSpot | unsupported destination reference | → MySQL (unsupported) |
| 13 | HubSpot | unsupported destination reference | → CockroachDB (unsupported) |
| 14 | HubSpot | unsupported destination reference | → MariaDB (unsupported) |
| 15 | HubSpot | unsupported destination reference | → SQLite (unsupported) |
| 16 | GitHub | all 12 streams | → PostgreSQL |
| 17 | GitHub | all 12 streams | → MySQL |
| 18 | GitHub | all 12 streams | → CockroachDB |
| 19 | GitHub | all 12 streams | → MariaDB |
| 20 | GitHub | all 12 streams | → SQLite |
| 21 | Notion | all 3 streams | → PostgreSQL |
| 22 | Notion | all 3 streams | → MySQL |
| 23 | Notion | all 3 streams | → CockroachDB |
| 24 | Notion | all 3 streams | → MariaDB |
| 25 | Notion | all 3 streams | → SQLite |

### DB → DB (25 pipelines) — tests ALL streams per source

| # | Source | All Streams Tested | → Destination |
|---|--------|--------------------|---------------|
| 26 | PostgreSQL | public.users, public.orders, public.payments | → PostgreSQL |
| 27 | PostgreSQL | public.users, public.orders, public.payments | → MySQL |
| 28 | PostgreSQL | public.users, public.orders, public.payments | → CockroachDB |
| 29 | PostgreSQL | public.users, public.orders, public.payments | → MariaDB |
| 30 | PostgreSQL | public.users, public.orders, public.payments | → SQLite |
| 31 | MySQL | mydb.products, mydb.inventory, mydb.categories | → PostgreSQL |
| 32 | MySQL | mydb.products, mydb.inventory, mydb.categories | → MySQL |
| 33 | MySQL | mydb.products, mydb.inventory, mydb.categories | → CockroachDB |
| 34 | MySQL | mydb.products, mydb.inventory, mydb.categories | → MariaDB |
| 35 | MySQL | mydb.products, mydb.inventory, mydb.categories | → SQLite |
| 36 | MariaDB | app.events, app.logs, app.sessions | → PostgreSQL |
| 37 | MariaDB | app.events, app.logs, app.sessions | → MySQL |
| 38 | MariaDB | app.events, app.logs, app.sessions | → CockroachDB |
| 39 | MariaDB | app.events, app.logs, app.sessions | → MariaDB |
| 40 | MariaDB | app.events, app.logs, app.sessions | → SQLite |
| 41 | SQLite | main.tasks, main.notes, main.tags | → PostgreSQL |
| 42 | SQLite | main.tasks, main.notes, main.tags | → MySQL |
| 43 | SQLite | main.tasks, main.notes, main.tags | → CockroachDB |
| 44 | SQLite | main.tasks, main.notes, main.tags | → MariaDB |
| 45 | SQLite | main.tasks, main.notes, main.tags | → SQLite |
| 46 | CockroachDB | public.sessions, public.accounts, public.transactions | → PostgreSQL |
| 47 | CockroachDB | public.sessions, public.accounts, public.transactions | → MySQL |
| 48 | CockroachDB | public.sessions, public.accounts, public.transactions | → CockroachDB |
| 49 | CockroachDB | public.sessions, public.accounts, public.transactions | → MariaDB |
| 50 | CockroachDB | public.sessions, public.accounts, public.transactions | → SQLite |

---

## Credential & Connection Strategy

- **SaaS sources** → Real JSON config files in `__tests__/fixtures/credentials/` (gitignored)
  - `stripe.json` — `{ "credential": "sk_test_..." }`
  - `shopify.json` — `{ "credential": "shpat_...", "shop_url": "https://..." }`
  - `hubspot.json` — `{ "credential": "pat-..." }`
  - `github.json` — `{ "credential": "ghp_...", "github_owner": "...", "github_repo": "..." }`
  - `notion.json` — `{ "credential": "secret_..." }`
- **DB sources/destinations** → Docker containers via `testcontainers`
  - PostgreSQL (`postgres:16-alpine`), MySQL (`mysql:8`), MariaDB (`mariadb:11`), CockroachDB (`cockroachdb/cockroach:latest-v24.1`), SQLite (temp file)
- Tests skip gracefully if credentials/Docker unavailable

---

## Plan

### Step 1 — Install Vitest + testcontainers & create config

- `bun add -D vitest @vitest/coverage-v8 testcontainers`
- Create `vitest.config.ts` with `@/*` path alias
- Add `"test": "vitest run"`, `"test:watch": "vitest"` scripts to `package.json`
- Add `__tests__/fixtures/credentials/` to `.gitignore`

### Step 2 — Create shared fixtures & factory helpers

**File**: `__tests__/fixtures/pipeline-fixtures.ts`

- All stream lists from above (complete, not abbreviated)
- `SAAS_SOURCES`, `DB_SOURCES`, `DB_DESTINATIONS` constants
- Per-SaaS stream map with ALL streams:
  ```ts
  SAAS_STREAMS = {
    stripe: ['customers','charges','invoices','subscriptions','products','prices','events','balance_transactions','coupons','payment_intents','payment_methods','refunds','disputes','payouts','plans','tax_rates','credit_notes','promotion_codes','setup_intents'],
    shopify: ['products','orders','customers','draft_orders','custom_collections','smart_collections','pages','blogs','articles','locations','price_rules','themes','countries','collects'],
    hubspot: ['contacts','companies','deals','tickets','owners','deal_pipelines','ticket_pipelines','products','line_items','quotes'],
    github: ['issues','pull_requests','stargazers','events','commits','branches','releases','tags','contributors','milestones','labels','forks'],
    notion: ['databases','pages','users'],
  }
  ```
- Per-DB stream map: 3 streams each
- `loadSaaSCredential(source)` → reads from `__tests__/fixtures/credentials/{source}.json`
- `buildMockSourceNode()`, `buildMockDestinationNode()`, `buildMockPipelineGraph()`
- `getDuckDBRef(connectorType, streamKey)` → expected DuckDB table name

**File**: `__tests__/fixtures/credentials/.gitkeep` + template JSON files

### Step 3 — Docker DB helper

**File**: `__tests__/e2e/docker-db-setup.ts`

- `startDockerDb(connectorType)` → spins up container, returns `{host, port, user, password, database}`
- `stopDockerDb(container)`
- SQLite: temp file
- `describe.skipIf(!isDockerAvailable())` guard

### Step 4 — Unit tests: `schema-table.test.ts` (~15 tests)

**File**: `lib/pipelines/__tests__/schema-table.test.ts`

- `duckdbTableNameForStream` for all DB schemas
- `duckdbTableNameForStream` for SaaS-style qualified keys (`stripe.charges` → `stripe__charges`)
- `parseQualifiedTable` valid, invalid, edge cases
- `buildSourceStreamConfig` for DB and SaaS stream keys

### Step 5 — Structured pipeline API tests

- Source connections can be shared while stream selections remain pipeline-specific.
- Transformation drafts require validation and preview before publication.
- Destination assignments reference published transformation revisions without copying SQL.
- One pipeline can fan out to several tested destinations with independent checkpoints.
- Legacy graph payloads are migrated once and cannot be written by current APIs.

### Step 6 — Structured pipeline workspace tests

- Directory rows open the pipeline Overview without action-menu propagation.
- Source, Transformations, Destinations, Runs, General, and GitHub remain URL-addressable.
- Transformation and destination editors use full pages at mobile widths.
- Tables provide server pagination and compact mobile rows.
- Legacy builder URLs redirect to the corresponding structured page.

### Step 7 — Unit tests: `credentialForm.test.ts` (~10 tests)

**File**: `app/workspace/connections/__tests__/credentialForm.test.ts`

- `buildTestDto()` for each SaaS and DB
- `buildCreateDto()` role=source for SaaS, role=destination for DB
- SQLite file path handling

### Step 8 — E2E: 50-pipeline matrix test (ALL streams per pipeline)

**File**: `__tests__/e2e/pipeline-combinations.test.ts`

The main test file — `describe.each` + `it.each`:

```ts
describe.each(SAAS_SOURCES)('%s → all destinations', (source) => {
  it.each(DB_DESTINATIONS)('pipeline %s → %s', (source, dest) => {
    // Tests ALL streams for this source (19 for Stripe, 14 for Shopify, etc.)
    for (const stream of SAAS_STREAMS[source]) {
      // verify duckdb ref, SQL model, graph, dest_table format
    }
  })
})
describe.each(DB_SOURCES)('%s → all destinations', (source) => {
  it.each(DB_DESTINATIONS)('pipeline %s → %s', (source, dest) => {
    for (const stream of DB_STREAMS[source]) {
      // verify duckdb ref, SQL model, graph, dest_table format
    }
  })
})
```

Each of the 50 tests validates **every stream** for that source:
1. ✅ Credential shape loads correctly (JSON for SaaS, Docker config for DB)
2. ✅ Every stream config produces correct `duckdb_table_name` (e.g. `stripe__balance_transactions`)
3. ✅ `buildDefaultSqlModel` generates SQL with correct `{{ source('raw', '...') }}` per stream
4. ✅ `normalizePipelineGraph` round-trips source→destination graph
5. ✅ `dest_table` is in `schema.table` format for every stream
6. ✅ `output_table` naming convention correct per stream
7. ✅ Full graph has valid edges connecting source to destination node

---

## Data Type & dbt Transformation Edge Cases

### Supported Data Types (from `delivery_row_coercion.py` + `normalisation_handler.py`)

| Category | DB Types Covered |
|----------|-----------------|
| **UUID** | `uuid`, `uniqueidentifier` |
| **JSON** | `json`, `jsonb` |
| **Text** | `text`, `varchar`, `character varying`, `char`, `citext`, `tinytext`, `mediumtext`, `longtext` |
| **Integer** | `integer`, `int`, `bigint`, `smallint`, `tinyint`, `int2`, `int4`, `int8`, `serial`, `bigserial` |
| **Decimal** | `decimal`, `numeric`, `numeric(10,2)` |
| **Float** | `float`, `double`, `real`, `double precision` |
| **Boolean** | `boolean`, `bool` |
| **Timestamp** | `timestamp`, `timestamptz`, `datetime`, `timestamp with time zone`, `timestamp without time zone` |
| **Date** | `date` |
| **Time** | `time`, `timetz` |
| **Binary** | `binary`, `varbinary`, `blob`, `bytea`, `bytes`, `tinyblob`, `mediumblob`, `longblob` |

### Step 9 — Unit tests: dbt SQL model transforms with data types (~25 tests)

**File**: `__tests__/unit/dbt-sql-transforms.test.ts`

Tests that custom dbt SQL models handle all data type scenarios:

**JSON extraction & flattening:**
- SaaS `metadata` JSON column → extract specific keys: `metadata->>'plan_id'`
- Nested JSON → `address->>'city'`, `address->>'zip_code'`
- JSON array length: `json_array_length(line_items)`
- SaaS auto-flattened columns: `address__city`, `metadata__plan__name` (dlt double-underscore)

**Type casting in SQL models:**
- `CAST(id AS VARCHAR)` — UUID → text
- `CAST(amount AS DECIMAL(10,2))` — integer → decimal
- `CAST(created_at AS DATE)` — timestamp → date
- `CAST(is_active AS INTEGER)` — boolean → integer
- `CAST(metadata AS TEXT)` — JSON → text (strip unwanted keys)
- `CAST(price AS DOUBLE)` — decimal → float

**Calculations & derived columns:**
- `price * quantity AS total_amount` — arithmetic
- `COALESCE(discount, 0) AS safe_discount` — NULL handling
- `CASE WHEN status = 'paid' THEN 1 ELSE 0 END AS is_paid` — conditional
- `CONCAT(first_name, ' ', last_name) AS full_name` — string concatenation
- `DATE_TRUNC('month', created_at) AS month` — date truncation

**Filtering:**
- `WHERE status != 'deleted'` — exclude soft-deleted
- `WHERE created_at > '2024-01-01'` — date filter
- `WHERE metadata IS NOT NULL` — NULL filter
- `WHERE amount > 0` — numeric filter

**Edge cases:**
- Empty JSON `{}` and `[]`
- NULL values for every data type
- Unicode strings in VARCHAR
- Very large BIGINT values
- Zero-datetime MySQL (`0000-00-00 00:00:00`)
- `NaN` / `Infinity` in float columns

### Step 10 — Integration tests: seeded Docker DB with all data types (pytest)

**File**: `apps/server/elt-server/tests/test_data_type_round_trip.py`

Seed Docker DBs with a test table containing **every data type**, then verify coercion + delivery:

**Seed table schema** (created in each Docker DB):
```sql
CREATE TABLE test_all_types (
  id UUID PRIMARY KEY,
  name VARCHAR(255),
  description TEXT,
  amount DECIMAL(12,4),
  quantity INTEGER,
  price DOUBLE PRECISION,
  is_active BOOLEAN,
  metadata JSON,        -- or JSONB for Postgres/Cockroach
  tags JSON,            -- array-like JSON
  created_at TIMESTAMP WITH TIME ZONE,
  updated_at TIMESTAMP WITHOUT TIME ZONE,
  birth_date DATE,
  login_time TIME,
  score SMALLINT,
  big_number BIGINT,
  raw_data BYTEA,       -- or BLOB for MySQL/MariaDB
  nullable_col VARCHAR(50)  -- always NULL
);
```

**Seed data** (one row per edge case):
1. Normal row — all fields populated with typical values
2. NULL row — all nullable fields NULL
3. JSON row — nested `{"address": {"city": "NYC", "zip": "10001"}, "tags": ["vip", "active"]}`
4. Empty JSON — `{}` and `[]`
5. Unicode row — `name = '日本語テスト'`, `description = 'Ñoño 🎉'`
6. Large numbers — `big_number = 9223372036854775807` (max bigint), `amount = 99999999.9999`
7. Edge timestamps — `created_at = '1970-01-01T00:00:00Z'`, `updated_at = '2099-12-31T23:59:59'`
8. Boolean variants — `is_active = true/false/NULL`

**Parameterized across all 5 DB types** (Postgres, MySQL, MariaDB, CockroachDB, SQLite):
- Seed → read via `coerce_row_for_destination` → verify types match
- Seed → read via `apply_normalisation_rules` with cast rules → verify output types
- Seed → verify `normalize_sql_type_token` maps every column correctly
- Seed → verify `build_dlt_columns_hint` produces correct native DDL per DB

**Per-DB type adaptations:**
- **Postgres/CockroachDB**: JSONB, UUID native, TIMESTAMPTZ, BYTEA
- **MySQL/MariaDB**: JSON (no JSONB), CHAR(36) for UUID, DATETIME, BLOB, zero-datetime `0000-00-00`
- **SQLite**: TEXT for JSON/UUID, INTEGER for booleans, REAL for floats, BLOB for binary

### Step 11 — Integration tests: normalisation rules with all cast types (pytest)

**File**: `apps/server/elt-server/tests/test_normalisation_all_casts.py`

Test `cast_value()` and `apply_normalisation_rules_to_record()` for every supported cast target:

| Cast target | Input values tested |
|-------------|-------------------|
| `string` / `text` / `varchar` | UUID, int, float, bool, datetime, JSON dict, bytes, None |
| `integer` / `bigint` / `smallint` | "42", 3.14, True, Decimal("100"), None, "" |
| `float` / `double` / `real` | "3.14", 42, Decimal("1.5"), None |
| `decimal` / `numeric` | "12.34", 42, 3.14, None |
| `boolean` / `bool` | "true", "false", "1", "0", "yes", "no", 1, 0, None |
| `timestamp` / `datetime` / `timestamptz` | "2024-01-02T03:04:05Z", date, epoch int, None |
| `date` | "2024-06-15", datetime, None |
| `time` / `timetz` | "14:30:00", datetime, None |
| `json` / `jsonb` | '{"a":1}', dict, list, None, "" |
| `binary` / `bytea` / `blob` | "hello", bytes, None |
| `uuid` | UUID object, "550e8400-...", None |

Plus rename rules:
- Rename `created` → `created_at`
- Rename + cast combo: rename `ts` → `created_at` AND cast to `timestamp`
- Table-scoped rules: rule applies only to `orders` table, not `users`

### Step 12 — Python integration: sources.yml generation (pytest)

**File**: `apps/server/elt-server/tests/test_duckdb_staged_sources_yml.py`

- Test `_write_ui_sql_project()` for all 5 SaaS sources with ALL their streams
- Test `_write_ui_sql_project()` for all 5 DB sources
- Verify model SQL files contain correct cast/transform expressions
- ~10 parameterized pytest cases

### Step 13 — Verify & run all

- `bun run test` — all Vitest tests
- `pytest` — all Python tests
- Ensure 50 pipeline combos + data type edge cases + normalisation casts all pass
- Migrate old `bun:test` imports in existing `pipelineGraph.test.ts`

---

## Files Created/Modified

| File | Type | ~Tests |
|------|------|--------|
| `vitest.config.ts` | Config | — |
| `__tests__/fixtures/pipeline-fixtures.ts` | Fixtures (all 77 streams) | — |
| `__tests__/fixtures/credentials/*.json` | SaaS creds (gitignored) | — |
| `__tests__/e2e/docker-db-setup.ts` | Docker helper | — |
| `lib/pipelines/__tests__/schema-table.test.ts` | Unit | ~15 |
| `builder/shared/__tests__/pipelineGraph.test.ts` | Unit | ~20 |
| `builder/panels/__tests__/destinationPanel.test.ts` | Unit | ~15 |
| `connections/__tests__/credentialForm.test.ts` | Unit | ~10 |
| `__tests__/e2e/pipeline-combinations.test.ts` | E2E — 50 pipelines × all streams | **50** |
| `__tests__/unit/dbt-sql-transforms.test.ts` | **Unit — data type transforms** | **~25** |
| `elt-server/tests/test_data_type_round_trip.py` | **Integration — seeded Docker all types** | **~40** |
| `elt-server/tests/test_normalisation_all_casts.py` | **Integration — all cast rules** | **~30** |
| `elt-server/tests/test_duckdb_staged_sources_yml.py` | Python integration | ~10 |
| **Total** | | **~215** |

---

## Key Decisions

- **Vitest** replaces `bun:test`; migrate existing test
- **Exactly 50 pipeline combos**: 5 SaaS × 5 DB + 5 DB × 5 DB
- **ALL available streams tested per pipeline** (62 SaaS streams + 15 DB streams = 77 total)
- **ALL 11 data type categories** tested with edge cases (UUID, JSON, VARCHAR, INT, DECIMAL, FLOAT, BOOL, TIMESTAMP, DATE, TIME, BINARY)
- **dbt SQL transforms**: JSON extraction, type casting, calculations, filtering, NULL handling
- **Seeded Docker round-trip**: real rows with every data type through full coercion pipeline
- **Real SaaS creds** in gitignored `__tests__/fixtures/credentials/*.json`
- **Docker** for all 5 DB types via `testcontainers`
- **No UI code changes** — tests only
- Tests skip gracefully when creds/Docker unavailable
