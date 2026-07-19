# YAML Contract

Pipeline YAML is graph-aligned and versioned as `kind: mantrixflow.pipeline`.

```yaml
version: 2
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
    transformations:
      - key: orders_clean
        name: Clean orders
        input_streams:
          - public.orders
        output_table: fct_orders
        destination_table: analytics.fct_orders
        write_mode: upsert
        upsert_keys:
          - id
        sql: |
          SELECT *
          FROM {{ source('raw', 'public__orders') }}
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
- In version 2, transformation definitions live under their owning destination's `transformations[]`.
- `dbt_config.sql_models[]` remains part of the runtime-compatible destination configuration.

## Import Rules

- `version: 2` is the current contract. Version 1 remains accepted during the compatibility window.
- `source.connection` and every `destinations[].connection` resolve by name inside the org.
- Duplicate matching names fail with an ambiguous connection error. The importer never guesses.
- Every destination-owned transformation must reference streams in `source.selected_streams[]`.
- Every version 2 transformation must include `key`, `name`, `input_streams`, `output_table`, `destination_table`, and `sql`.
- `destination_table` must use `schema.table` form.
- SQL is validated through the existing ELT `/validate-sql` path when source column hints are available. If an organization has not cached source column metadata yet, import still enforces the YAML/connection/model contract and resumes ELT SQL validation once hints exist.
- Valid imports rebuild `pipeline_graph` and update the primary source/destination schema mirrors.

## Runtime Rule

The DB graph remains the runtime source used by pipeline execution. YAML is the reviewed audit source that can update the DB only after validation.
