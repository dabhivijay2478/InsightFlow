# Pipeline 11 — HubSpot → PostgreSQL

Status: production live-test guide

Extraction: dlt → isolated per-run DuckDB

Transform: mandatory UI SQL/dbt

Delivery: existing PostgreSQL tables, upsert only

Follow the [HubSpot source guide](../../sources/hubspot.md) for the exact
ten-stream catalog, permissions, incremental settings, masking, and reset flow.

## Minimal contacts scenario

Create the destination before running:

```sql
CREATE SCHEMA IF NOT EXISTS analytics;
CREATE TABLE analytics.hubspot_contacts (
    contact_id text PRIMARY KEY,
    email text,
    first_name text,
    last_name text,
    phone text,
    lifecycle_stage text,
    owner_id text,
    created_at timestamptz,
    updated_at timestamptz,
    archived boolean,
    properties jsonb,
    _mantrixflow_run_id text,
    extracted_at timestamptz
);
```

Select `hubspot.contacts`, choose incremental mode with `updated_at`, and set a
bounded UTC start date. Save an explicit model based on
`{{ source('raw', 'hubspot__contacts') }}` whose output exactly matches the
existing table. Do not include `_dlt_*` columns. Use `NULLIF` and explicit casts
for dates/numbers. Set the destination to `analytics.hubspot_contacts`; the only
write mode is upsert.

Run once, update a contact in HubSpot, then run again. Verify:

- only selected dlt resources made HubSpot API calls;
- the second run used a bounded overlapped search window;
- PostgreSQL contains one row per `contact_id` and the changed row was updated;
- `phase3_rows_delivered` equals committed destination rows;
- the contacts checkpoint advanced only after transform and delivery succeeded;
- the temporary DuckDB and dbt directory were removed;
- preview, callback, logs, and AI context contain no token or unmasked phone;
- a missing table, column, type, or primary key fails without creating or
  altering a target and without falling back to append.

Repeat the controlled scenario for companies, deals, tickets, and owners as a
production regression. Cover the remaining five streams with the
mocked/integration matrix described in the full-phase plan.
