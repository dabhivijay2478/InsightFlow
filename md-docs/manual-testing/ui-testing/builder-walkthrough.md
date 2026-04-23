# Pipeline Builder — Universal UI Walkthrough

Panel-by-panel instructions that apply to every pipeline test. Source-specific stream details are in the per-source files.

---

## Step 1 — Navigate to Pipeline Builder

1. Log in → go to **Workspace → Data Pipelines**
2. Click **"+ New Pipeline"**
3. Enter pipeline name e.g. `Stripe → PostgreSQL (charges)` and optional description
4. Click **"Create"** → builder canvas opens

---

## Step 2 — Source Panel Configuration

### 2a. Add Source
1. Click **"Add Source"** (left panel)
2. Select **source type** from the dropdown (Stripe / Shopify / HubSpot / GitHub / Notion / PostgreSQL / MySQL / MariaDB / SQLite / CockroachDB)
3. Fill in all credential fields (see per-source file for exact fields)
4. Click **"Test Connection"**
   - ✅ Green: proceed
   - ❌ Red: check credentials, network, firewall — error message shown inline

### 2b. Select Streams
1. Stream list loads after successful connection
2. Toggle **checkbox** to enable each stream
3. At least one stream must be selected — **Save** button grays out if none selected

### 2c. Sync Mode per Stream
For each enabled stream, set sync mode via the **Sync Mode** dropdown on the stream row:

| Option | When to use |
|--------|-------------|
| **Full Table** | No reliable cursor field; small tables; reference/lookup data |
| **Incremental** | Large tables with `updated_at` / `created_at` / integer sequence column |

### 2d. Cursor Field (Incremental only)
When **Incremental** selected on a stream:
1. **Cursor Field** dropdown appears immediately below the sync mode
2. Select the watermark column (e.g. `updated_at`, `created`, `id`)
3. ⚠️ Selecting a non-monotonic column (e.g. boolean, status) will cause missed records

**Test cases for cursor field:**

| Test | Action | Expected |
|------|--------|---------|
| Valid timestamp cursor | Select `updated_at` | Dropdown accepts; field highlighted |
| Valid integer cursor | Select `id` | Dropdown accepts |
| No cursor selected (Incremental mode) | Leave blank | Save blocked: "cursor field required" error |
| Wrong column type | Select a TEXT column | UI should warn; if not — document as known gap |

---

## Step 3 — Destination Panel Configuration

### 3a. Add Destination
1. Click **"Add Destination"** (right panel)
2. Select **destination type** (PostgreSQL / MySQL / MariaDB / SQLite / CockroachDB)
3. Fill in connection fields:

| Destination | Required fields |
|-------------|----------------|
| PostgreSQL | host, port, database, username, password, ssl_mode |
| MySQL | host, port, database, username, password |
| MariaDB | host, port, database, username, password |
| SQLite | absolute file path |
| CockroachDB | host, port, database, username, password, ssl_mode |

4. Click **"Test Connection"**
   - ✅ Green: proceed
   - ❌ Red: error shown — check host/credentials/firewall

### 3b. Map Streams to Destination Tables
For each selected source stream:
1. **Schema field**: enter destination schema name (e.g. `analytics`)
   - SQLite: use `main`
2. **Table field**: enter destination table name (e.g. `stripe_charges`)
3. Destination table must already exist with correct DDL (see `pipelines/` docs)

### 3c. Write Disposition
| Mode | Behaviour |
|------|-----------|
| **Merge (Upsert)** | INSERT … ON CONFLICT UPDATE — requires PK |
| **Append** | INSERT only — no PK check |

Select based on whether the destination table has a primary key.

---

## Step 4 — Normalisation Panel

### 4a. Open Normalisation
1. Click **"Normalisation"** tab (middle panel) — or select from the stream row options
2. Panel shows rule list (empty by default)

### 4b. Add Rename Rule
1. Click **"+ Add Rule"**
2. **Rule type**: `Rename`
3. **Column**: select source column from dropdown (e.g. `user_id`)
4. **New name**: type destination column name (e.g. `customer_ref`)
5. Click **"Save Rule"**

**Verify**: column now labelled `customer_ref` in preview panel

### 4c. Add Cast Rule
1. Click **"+ Add Rule"**
2. **Rule type**: `Cast`
3. **Column**: select source column (e.g. `active`)
4. **Cast to**: select target type from dropdown (`Boolean`, `Integer`, `Text`, `Numeric`, `Date`, `Timestamp`)
5. Click **"Save Rule"**

**Verify**: column type changed in preview panel

### 4d. Add Exclude Rule
1. Click **"+ Add Rule"**
2. **Rule type**: `Exclude`
3. **Column**: select column to drop (e.g. `metadata`)
4. Click **"Save Rule"**

**Verify**: column does NOT appear in dbt SQL editor or preview

### 4e. Normalisation Edge Cases

| Test | Expected |
|------|---------|
| Rename to same name as existing column | UI shows conflict error |
| Cast TEXT column to Integer (non-numeric value) | Pipeline runs; NULL produced for non-numeric rows |
| Exclude the PK column | UI should warn — delivery will fail without PK |
| Add 0 rules | Columns pass through unchanged |
| Add rename + cast on same column | Rename applied first, then cast |

---

## Step 5 — dbt SQL Panel (Transform)

### 5a. Open SQL Editor
1. Click **"Transform"** tab (middle panel)
2. SQL editor opens — default template shown: `SELECT * FROM {{ source('raw', 'stream_name') }}`
3. **Source table token**: `{{ source('raw', '<schema>__<stream>') }}` — double-underscore separator

### 5b. Write the SELECT Statement
Replace default template with stream-specific SQL (see per-source UI files).

**Key patterns:**

```sql
-- JSON key extraction
payload->>'action'                    AS action_type

-- Nested JSON
commit->'author'->>'name'             AS author_name

-- Type cast
CAST(amount AS NUMERIC) / 100         AS amount_dollars

-- Unix epoch → timestamp
TO_TIMESTAMP(created)::TIMESTAMPTZ    AS created_at

-- Boolean derivation
status = 'active'                     AS is_active

-- CASE expression
CASE priority WHEN 3 THEN 'high' WHEN 2 THEN 'medium' ELSE 'low' END AS urgency

-- COALESCE / NULLIF guard
CAST(NULLIF(price, '') AS NUMERIC)    AS unit_price
```

### 5c. Validate SQL
1. Click **"Validate"** button in editor toolbar
2. ✅ Green: column list appears below editor
3. ❌ Red: error message with line number — fix and re-validate

**Validate test cases:**

| SQL | Expected validation |
|-----|-------------------|
| `SELECT *` | ✅ All source columns listed |
| Missing FROM | ❌ `syntax error at or near EOF` |
| Wrong source token | ❌ `table not found` |
| Invalid JSON operator `->>'key'` on non-JSON column | ❌ `operator does not exist` |
| Valid JSON path with null key | ✅ Validates; NULL at runtime |

### 5d. Column Preview in SQL Editor
After validation, editor shows:
- Output column names
- Inferred types

**Check**: output column names must match destination table DDL exactly.

---

## Step 6 — Preview Panel

### 6a. Open Preview
1. Click **"Preview"** tab (middle panel)
2. System runs dbt SQL against staged sample data (first ~100 rows)
3. Preview table renders with column headers and sample values

### 6b. What to Check

| Check | Expected |
|-------|---------|
| Column names | Match destination DDL column names exactly |
| JSON extracted columns | Show plain strings, not `{"key":"value"}` |
| Renamed columns | Old name absent; new name present |
| Excluded columns | Not visible in preview |
| CAST columns | Correct type shown (e.g. 29.99 not 2999) |
| Boolean derived columns | `true`/`false` strings |
| NULL values | Shown as empty/null cell |

### 6c. Preview Edge Cases

| Test | Expected |
|------|---------|
| Empty source (0 rows) | Preview shows column headers, 0 data rows |
| Source column name has spaces | Column quoted in dbt SQL: `"column name"` |
| Preview truncated at 100 rows | Warning shown — not a bug |

---

## Step 7 — Schedule Panel

### 7a. Open Schedule
1. Click **"Schedule"** tab (top pipeline bar)
2. Three options: **None** / **Cron** / **Interval**

### 7b. None (Manual only)
- Select **None** → pipeline only runs via **"Run Now"** button
- Verify: no schedule badge shown on pipeline card

### 7c. Cron Schedule
1. Select **Cron**
2. Enter cron expression, e.g.:
   - `0 6 * * *` — daily at 06:00 UTC
   - `0 */6 * * *` — every 6 hours
   - `0 0 * * 1` — every Monday midnight
3. Select **timezone** from dropdown (default UTC)
4. Click **Save**
5. **Verify**: next scheduled run time shown (e.g. `Next run: tomorrow 06:00 UTC`)

**Cron test cases:**

| Expression | Expected display |
|-----------|----------------|
| `0 6 * * *` | "Daily at 06:00 UTC" |
| `*/15 * * * *` | "Every 15 minutes" |
| `0 0 1 * *` | "Monthly on 1st at 00:00 UTC" |
| `invalid cron` | ❌ "Invalid cron expression" error |
| `60 * * * *` | ❌ Minute out of range (0-59) |

### 7d. Interval Schedule
1. Select **Interval**
2. Enter number + unit (e.g. `6` `hours`, `30` `minutes`, `1` `day`)
3. Click **Save**

---

## Step 8 — Save, Run & Monitor

### 8a. Save Pipeline
1. Click **"Save Pipeline"** button (top bar)
2. ✅ Toast: "Pipeline saved"
3. Pipeline appears in pipeline list with correct name and status

### 8b. Run Now
1. Click **"Run Now"** button
2. Run status drawer slides up from bottom

### 8c. Run Status Drawer — Phase Monitoring

Watch each phase in the drawer:

```
◉ Phase 0 — Pre-flight checks
  → Verifying destination table exists
  → Verifying PK column match
  ✅ Pass / ❌ Fail + error message

◉ Phase 1 — Extract + Stage
  → Fetching from source API / DB
  → Loading into DuckDB staging
  ✅ N rows extracted

◉ Phase 2 — dbt Transform
  → Executing SQL model
  ✅ N rows transformed

◉ Phase 3 — Deliver
  → Upserting to destination DB
  ✅ N rows written

◉ Phase 4 — Cleanup
  → Dropping staging tables
  ✅ Done

◉ Phase 5 — Callback
  → Status written
  ✅ Complete
```

### 8d. Concurrent Run Rejection Test
1. Click **"Run Now"** on a running pipeline
2. ✅ Expected: second run blocked — "A run is already in progress" error shown in UI

### 8e. Run History
1. Navigate to **Pipeline → Runs** tab
2. Verify most recent run shows:
   - Status: `success`
   - Duration
   - Rows extracted / delivered

---

## Step 9 — Destination Verification

After a successful run, connect to destination DB and verify:

```sql
-- Row count (must be > 0)
SELECT COUNT(*) FROM analytics.table_name;

-- No duplicates (merge mode)
SELECT pk_col, COUNT(*) cnt FROM analytics.table_name GROUP BY pk_col HAVING cnt > 1;

-- No _dlt_ staging tables
SELECT table_name FROM information_schema.tables WHERE table_name LIKE '_dlt_%';

-- Sample data spot check
SELECT * FROM analytics.table_name LIMIT 5;
```

---

## Step 10 — Incremental Re-run Test

1. Run pipeline once (full extraction)
2. Insert 1–2 new rows at the source with `updated_at = NOW()`
3. Run pipeline again
4. ✅ Expected: only new rows delivered (check row count delta = 1–2)
5. ❌ If all rows re-delivered → incremental cursor not working — check cursor field selection
