# Live PostgreSQL to PostgreSQL Data-Type UI Test

Use this guide to verify the current MantrixFlow UI can move live PostgreSQL data into another PostgreSQL destination while preserving or intentionally casting common data types.

This test uses one source table with many PostgreSQL types, a pre-created destination table, dbt SQL in the pipeline builder, and UI-only pipeline execution.

## What This Test Proves

- Source and destination PostgreSQL connections can be created and tested from the UI.
- The pipeline builder discovers a real PostgreSQL source stream.
- dbt SQL can read from the staged source table using `{{ source('raw', 'schema__table') }}`.
- Destination writes land in the exact pre-created table.
- UUID, integer, numeric, boolean, date, timestamp, timestamptz, text, JSONB, arrays, and bytea-derived values behave correctly.
- Incremental/upsert reruns update existing rows instead of duplicating them.

## Safety Rules

- Use test databases or dedicated test schemas only.
- Do not point source and destination to the same schema/table.
- Do not paste DB passwords into screenshots, tickets, or logs.
- The ELT engine does not create destination tables. Pre-create every destination table before clicking Run.
- RDS, Supabase, and managed Postgres usually need SSL enabled in the connection form.

## Test Names

| Purpose | Value |
| --- | --- |
| Source schema | `mxf_live_p2p_src_types` |
| Source table | `orders_type_matrix` |
| Destination schema | `mxf_live_p2p_dest_types` |
| Destination table | `orders_type_matrix_curated` |
| First sync mode | `FULL_TABLE` |
| Rerun sync mode | `INCREMENTAL` |
| Incremental key | `updated_at` |
| Write mode | `upsert` / merge |
| Primary key | `id` |

You can use two separate PostgreSQL servers, or one PostgreSQL server with two isolated schemas. Separate source and destination databases are closer to production, but same-server schema isolation is fine for this UI test.

## Step 1 - Create The Source Schema And Table

Run this SQL in the source PostgreSQL database.

```sql
CREATE SCHEMA IF NOT EXISTS mxf_live_p2p_src_types;

DROP TABLE IF EXISTS mxf_live_p2p_src_types.orders_type_matrix;

CREATE TABLE mxf_live_p2p_src_types.orders_type_matrix (
    id uuid PRIMARY KEY,
    order_no bigint NOT NULL,
    small_count smallint,
    item_count integer,
    amount numeric(12,2),
    tax_rate real,
    risk_score double precision,
    is_paid boolean,
    order_date date,
    order_time time,
    created_at timestamp,
    updated_at timestamptz NOT NULL,
    customer_email varchar(255),
    notes text,
    metadata jsonb,
    tags text[],
    payload json,
    raw_bytes bytea
);

INSERT INTO mxf_live_p2p_src_types.orders_type_matrix (
    id,
    order_no,
    small_count,
    item_count,
    amount,
    tax_rate,
    risk_score,
    is_paid,
    order_date,
    order_time,
    created_at,
    updated_at,
    customer_email,
    notes,
    metadata,
    tags,
    payload,
    raw_bytes
) VALUES
(
    '11111111-1111-1111-1111-111111111111',
    1001,
    2,
    3,
    199.95,
    0.075,
    0.9821,
    true,
    DATE '2026-06-01',
    TIME '10:15:30',
    TIMESTAMP '2026-06-01 10:15:30',
    TIMESTAMPTZ '2026-06-01 10:15:30+05:30',
    'alpha@example.com',
    'first order with json, array, and bytes',
    '{"city":"Mumbai","channel":"web","tier":"growth","priority":true}'::jsonb,
    ARRAY['new','priority'],
    '{"source":"rds","version":1}'::json,
    decode('DEADBEEF', 'hex')
),
(
    '22222222-2222-2222-2222-222222222222',
    1002,
    0,
    1,
    49.00,
    0.050,
    0.4210,
    false,
    DATE '2026-06-02',
    TIME '14:05:00',
    TIMESTAMP '2026-06-02 14:05:00',
    TIMESTAMPTZ '2026-06-02 14:05:00+00',
    'beta@example.com',
    'second order with unicode: namaste',
    '{"city":"Bengaluru","channel":"partner","tier":"starter","priority":false}'::jsonb,
    ARRAY['partner'],
    '{"source":"manual","version":2}'::json,
    decode('AABBCC', 'hex')
),
(
    '33333333-3333-3333-3333-333333333333',
    1003,
    NULL,
    NULL,
    NULL,
    NULL,
    NULL,
    NULL,
    DATE '2026-06-03',
    NULL,
    TIMESTAMP '2026-06-03 08:00:00',
    TIMESTAMPTZ '2026-06-03 08:00:00-07',
    NULL,
    NULL,
    '{"city":null,"channel":"api","tier":"pro"}'::jsonb,
    ARRAY[]::text[],
    '{"source":"null-case"}'::json,
    NULL
);
```

Confirm the source data exists:

```sql
SELECT COUNT(*) AS source_rows
FROM mxf_live_p2p_src_types.orders_type_matrix;

SELECT order_no, amount, is_paid, metadata, tags
FROM mxf_live_p2p_src_types.orders_type_matrix
ORDER BY order_no;
```

Expected: `source_rows = 3`.

## Step 2 - Create The Destination Schema And Table

Run this SQL in the destination PostgreSQL database.

```sql
CREATE SCHEMA IF NOT EXISTS mxf_live_p2p_dest_types;

DROP TABLE IF EXISTS mxf_live_p2p_dest_types.orders_type_matrix_curated;

CREATE TABLE mxf_live_p2p_dest_types.orders_type_matrix_curated (
    id uuid PRIMARY KEY,
    order_no bigint NOT NULL,
    amount numeric(12,2),
    is_paid boolean,
    order_date date,
    created_at timestamp,
    updated_at timestamptz,
    customer_email text,
    city text,
    channel text,
    priority boolean,
    tag_count integer,
    raw_bytes_hex text,
    metadata jsonb
);
```

Confirm the destination table is empty:

```sql
SELECT COUNT(*) AS destination_rows
FROM mxf_live_p2p_dest_types.orders_type_matrix_curated;
```

Expected: `destination_rows = 0`.

## Step 3 - Create Or Verify UI Connections

Open the app:

```text
https://cloud.mantrixflow.com/workspace/connections
```

Create or verify two connections.

### Source Connection

1. Click `New connection`.
2. Choose `PostgreSQL`.
3. Use the source database host, port, database, username, and password.
4. Enable SSL if the database requires it.
5. Click `Test connection`.
6. Save the connection with a clear name, for example `Live P2P Source Types`.

### Destination Connection

1. Click `New connection`.
2. Choose `PostgreSQL`.
3. Use the destination database host, port, database, username, and password.
4. Enable SSL if the database requires it.
5. Click `Test connection`.
6. Save the connection with a clear name, for example `Live P2P Destination Types`.

If connection testing fails, fix credentials, SSL, network allowlists, or DB grants before continuing.

## Step 4 - Build The Pipeline In The UI

Open:

```text
https://cloud.mantrixflow.com/workspace/pipelines
```

1. Click `New pipeline`.
2. Name it `Live Postgres Type Matrix`.
3. Select the source PostgreSQL connection.
4. Select source schema `mxf_live_p2p_src_types`.
5. Select source table `orders_type_matrix`.
6. Set sync mode to `FULL_TABLE` for the first run.
7. Add a PostgreSQL destination node.
8. Select the destination PostgreSQL connection.
9. Set destination schema to `mxf_live_p2p_dest_types`.
10. Set destination table to `orders_type_matrix_curated`.
11. Set write mode to `upsert` / merge.
12. Confirm the primary key is `id`.

## Step 5 - Add The dbt SQL Model

In the SQL/dbt editor for this stream, use:

This SQL runs in DuckDB during the dbt phase, so use DuckDB functions such as
`json_extract_string`, `TRY_CAST`, `json_array_length`, and `hex` instead of
PostgreSQL-only functions. PostgreSQL `text[]` columns are staged as DuckDB
`JSON`, so use `json_array_length(tags)` rather than `array_length(tags)`.

```sql
SELECT
    id,
    order_no,
    CAST(amount AS DECIMAL(12,2)) AS amount,
    is_paid,
    order_date,
    created_at,
    updated_at,
    CAST(customer_email AS TEXT) AS customer_email,
    json_extract_string(metadata, '$.city') AS city,
    json_extract_string(metadata, '$.channel') AS channel,
    COALESCE(TRY_CAST(json_extract_string(metadata, '$.priority') AS BOOLEAN), false) AS priority,
    COALESCE(json_array_length(tags), 0)::INTEGER AS tag_count,
    CASE WHEN raw_bytes IS NULL THEN NULL ELSE hex(raw_bytes) END AS raw_bytes_hex,
    metadata
FROM {{ source('raw', 'mxf_live_p2p_src_types__orders_type_matrix') }}
WHERE id IS NOT NULL
```

Then:

1. Click `Validate SQL`.
2. Click `Preview`.
3. Confirm the preview has the destination columns listed above.
4. Save the pipeline.

The DuckDB staging name for database sources is `{schema}__{table}`, so this source table becomes:

```text
mxf_live_p2p_src_types__orders_type_matrix
```

## Step 6 - Run And Watch The Status Drawer

1. Click `Run now`.
2. Open the run status drawer.
3. Confirm:
   - Extract/stage reads 3 rows.
   - dbt transform completes.
   - Delivery writes to `mxf_live_p2p_dest_types.orders_type_matrix_curated`.
   - Final run status is `completed`.

## Step 7 - Verify Destination Data

Run these checks in the destination PostgreSQL database.

```sql
SELECT COUNT(*) AS destination_rows
FROM mxf_live_p2p_dest_types.orders_type_matrix_curated;
```

Expected: `destination_rows = 3`.

Check destination column types:

```sql
SELECT column_name, data_type
FROM information_schema.columns
WHERE table_schema = 'mxf_live_p2p_dest_types'
  AND table_name = 'orders_type_matrix_curated'
ORDER BY ordinal_position;
```

Check representative values and runtime types:

```sql
SELECT
    order_no,
    amount,
    pg_typeof(amount) AS amount_type,
    is_paid,
    pg_typeof(is_paid) AS is_paid_type,
    order_date,
    pg_typeof(order_date) AS order_date_type,
    updated_at,
    pg_typeof(updated_at) AS updated_at_type,
    city,
    channel,
    priority,
    tag_count,
    raw_bytes_hex,
    metadata,
    pg_typeof(metadata) AS metadata_type
FROM mxf_live_p2p_dest_types.orders_type_matrix_curated
ORDER BY order_no;
```

Expected highlights:

| Check | Expected |
| --- | --- |
| Row count | `3` |
| `id` | `uuid` primary key |
| `amount` | `numeric` |
| `is_paid` | `boolean` |
| `order_date` | `date` |
| `updated_at` | `timestamp with time zone` |
| `metadata` | `jsonb` |
| `tag_count` | `2`, `1`, `0` |
| `raw_bytes_hex` | `deadbeef`, `aabbcc`, `NULL` |

Confirm the destination table stayed explicit and no accidental client destination table was created:

```sql
SELECT table_schema, table_name
FROM information_schema.tables
WHERE table_schema = 'mxf_live_p2p_dest_types'
ORDER BY table_name;
```

Expected: only `orders_type_matrix_curated` unless you intentionally created other test tables.

## Step 8 - Incremental Upsert Rerun

Run this SQL in the source PostgreSQL database.

```sql
UPDATE mxf_live_p2p_src_types.orders_type_matrix
SET
    amount = 249.95,
    metadata = jsonb_set(metadata, '{channel}', '"mobile"', true),
    updated_at = NOW()
WHERE order_no = 1001;

INSERT INTO mxf_live_p2p_src_types.orders_type_matrix (
    id,
    order_no,
    small_count,
    item_count,
    amount,
    tax_rate,
    risk_score,
    is_paid,
    order_date,
    order_time,
    created_at,
    updated_at,
    customer_email,
    notes,
    metadata,
    tags,
    payload,
    raw_bytes
) VALUES (
    '44444444-4444-4444-4444-444444444444',
    1004,
    5,
    8,
    799.00,
    0.090,
    0.9950,
    true,
    CURRENT_DATE,
    CURRENT_TIME,
    NOW()::timestamp,
    NOW(),
    'delta@example.com',
    'incremental insert',
    '{"city":"Pune","channel":"api","tier":"pro","priority":true}'::jsonb,
    ARRAY['incremental','vip'],
    '{"source":"incremental"}'::json,
    decode('FACEFEED', 'hex')
);
```

In the UI:

1. Open the same pipeline.
2. Change sync mode to `INCREMENTAL`.
3. Set replication key to `updated_at`.
4. Keep write mode as `upsert` / merge.
5. Save.
6. Click `Run now`.

Verify the destination:

```sql
SELECT COUNT(*) AS destination_rows
FROM mxf_live_p2p_dest_types.orders_type_matrix_curated;
```

Expected: `destination_rows = 4`.

```sql
SELECT order_no, amount, channel, updated_at
FROM mxf_live_p2p_dest_types.orders_type_matrix_curated
WHERE order_no IN (1001, 1004)
ORDER BY order_no;
```

Expected:

- `1001` has `amount = 249.95` and `channel = mobile`.
- `1004` exists.

Confirm there are no duplicate primary keys:

```sql
SELECT id, COUNT(*)
FROM mxf_live_p2p_dest_types.orders_type_matrix_curated
GROUP BY id
HAVING COUNT(*) > 1;
```

Expected: no rows.

## Troubleshooting

| Symptom | Root cause | Fix |
| --- | --- | --- |
| `connection_type and config required` | UI/API payload mismatch or old deployed API | Deploy latest API/app and re-save the connection. |
| `Saved connection credentials could not be read` | Connection was saved with a different encryption key or stale secret | Re-save the connection; confirm `ENCRYPTION_MASTER_KEY` did not change unintentionally. |
| `relation does not exist` | Destination table was not pre-created or schema/table is wrong | Run the destination DDL again and verify exact schema/table names. |
| `permission denied for schema` | DB user lacks schema/table grants | Grant `USAGE` on schema and `SELECT` or write permissions on the table. |
| SQL preview cannot find source | Wrong DuckDB source name | Use `mxf_live_p2p_src_types__orders_type_matrix`. |
| Upsert fails | Destination table has no primary key | Ensure `id uuid PRIMARY KEY` exists on the destination table. |
| SSL connection error | Managed Postgres requires SSL | Enable SSL in the connection form. |
| Unsupported type error | Very specialized type was selected raw | Cast unusual types to text in dbt SQL before delivery. |

## Cleanup

When testing is complete, run these only in the dedicated test schemas.

Source database:

```sql
DROP SCHEMA IF EXISTS mxf_live_p2p_src_types CASCADE;
```

Destination database:

```sql
DROP SCHEMA IF EXISTS mxf_live_p2p_dest_types CASCADE;
```

## Acceptance Checklist

- Source connection test passes.
- Destination connection test passes.
- Source table discovery shows `mxf_live_p2p_src_types.orders_type_matrix`.
- SQL validation passes.
- Preview shows curated destination columns.
- First run completes with 3 rows written.
- Destination types match the expected types.
- Incremental rerun updates order `1001` and inserts order `1004`.
- Destination row count is 4 after the incremental rerun.
- No duplicate primary keys exist.
- No destination table is created implicitly by the ELT engine.
