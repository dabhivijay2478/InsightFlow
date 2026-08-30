# Consolidated Real-Connector Stream Coverage

| Connector | Discovered | Selected | Executed | Latest UI result |
| :--- | ---: | ---: | ---: | :--- |
| Stripe | 34 | 34 | 34 | Success |
| HubSpot | 10 | 10 | 10 | Success |
| PostgreSQL broad-type source | 6 tables | 6 | 6 | Success |
| PostgreSQL scale source | 1 table | 1 | 1 | Success twice |

## Stripe

All 34 discovered streams were selected and executed. Nineteen discovered cursor-capable streams used incremental mode and the 15 cursorless streams used full-table mode. Empty streams completed normally.

## HubSpot

All 10 discovered streams were selected. Each stream has a retained validated/published model and pre-created destination table. Incremental object streams and full-table owner/pipeline-stage streams completed together in the final run.

## PostgreSQL

The full-sync pipeline covered six tables and a 31-column broad PostgreSQL type fixture. The scale pipeline retained 10,000 destination rows after an 8,000-row initial load and a 2,025-row incremental upsert.

Detailed inventories are in `STRIPE_STREAM_COVERAGE.md`, `HUBSPOT_STREAM_COVERAGE.md`, and `POSTGRES_TYPE_COMPATIBILITY.md`.
