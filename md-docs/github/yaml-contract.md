# YAML Contract

Pipeline YAML is graph-aligned and versioned as `kind: mantrixflow.pipeline`.

```yaml
version: 1
kind: mantrixflow.pipeline
name: Orders Sync
description: Sync orders into analytics
schedule:
  type: hourly
  value: ""
  timezone: UTC

source:
  connection: Production Postgres
  connector_type: postgres
  selected_streams:
    - stream_key: public.orders
      sync_mode: INCREMENTAL
      replication_method: INCREMENTAL
      replication_key: updated_at
      duckdb_table_name: public__orders

destinations:
  - id: destination-1
    label: Analytics Warehouse
    connection: Analytics Postgres
    connector_type: postgres
    dest_schema: analytics
    write_mode: upsert
    sync_mode: INCREMENTAL
    replication_key: updated_at
    dbt_config:
      mode: ui_sql
      target_schema: analytics
      run_tests: false
      sql_models:
        - source_stream_key: public.orders
          duckdb_source_table: public__orders
          output_table: fct_orders
          dest_table: analytics.fct_orders
          sql: |
            SELECT *
            FROM {{ source('raw', 'public__orders') }}
```

## Export Rules

- Serialize connection names, connector types, stream keys, SQL models, schedules, destination schema, write mode, replication keys, and normalisation rules.
- Never serialize connection IDs, passwords, tokens, SSH values, private keys, credentials, or encrypted blobs.
- Source stream keys must remain `schema.table`.
- `duckdb_source_table` and `duckdb_table_name` must match the strict ELT staging contract: `schema__table`.
- SQL models must live under each destination node's `dbt_config.sql_models[]`.

## Import Rules

- `version` must be `1`.
- `source.connection` and every `destinations[].connection` resolve by name inside the org.
- Duplicate matching names fail with an ambiguous connection error. The importer never guesses.
- Every SQL model must reference a stream in `source.selected_streams[]`.
- Every SQL model must include `source_stream_key`, `output_table`, and `sql`.
- SQL is validated through the existing ELT `/validate-sql` path when source column hints are available. If an organization has not cached source column metadata yet, import still enforces the YAML/connection/model contract and resumes ELT SQL validation once hints exist.
- Valid imports rebuild `pipeline_graph` and update the primary source/destination schema mirrors.

## Runtime Rule

The DB graph remains the runtime source used by pipeline execution. YAML is the reviewed audit source that can update the DB only after validation.
