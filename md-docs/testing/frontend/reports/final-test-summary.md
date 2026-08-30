# MantrixFlow Full-Scale Real Chrome E2E Summary

Run date: 26 July 2026

## Final result

All requested connector pipelines were created, configured, run, diagnosed, fixed, and rerun through the real MantrixFlow UI in Chrome. The latest run for every requested pipeline is successful.

| Connector | Coverage | Latest result | Data result |
| :--- | :--- | :--- | :--- |
| PostgreSQL full sync | 6 tables, 31 type-bearing columns in the broad type fixture | Passed | All 6 extracted; 3 representative type rows delivered |
| PostgreSQL incremental | `large_dataset`, `updated_at` cursor, `id` key | Passed twice | 8,000 initial rows, then 2,025 changed rows; 10,000 retained destination rows |
| Stripe | 34 of 34 discovered streams | Passed | All 34 streams executed; 1 representative account row delivered |
| HubSpot | 10 of 10 discovered streams | Passed | 10 models/tables executed; 17 destination rows retained |

## PostgreSQL scale proof

- The broad type fixture used 30 explicitly declared PostgreSQL type columns plus its serial primary key.
- Initial UI scale run inserted 8,000 source rows.
- A delta added 2,000 new rows and updated 25 existing rows.
- The next UI incremental run processed 2,025 rows.
- Destination verification found 10,000 total rows, exactly 2,000 IDs above 8,000, and all 25 expected updated records.

## SaaS stream proof

- Stripe selected and executed every one of the 34 streams returned by authoritative discovery. Empty streams completed normally and were materialized without false failures.
- HubSpot selected all 10 streams. Strict model coverage was satisfied with one validated, previewed, published model and one pre-created destination table per stream.
- HubSpot destination verification found 17 total rows across the 10 retained tables.

## Fix-and-rerun proof

The retained run history includes the failures used to diagnose scale delivery, schema-catalog caching, dynamic SaaS fields, long PostgreSQL identifiers, sparse properties, and composite pipeline-stage keys. After the fixes, the final Chrome run for each connector completed successfully.

## Retention

No test connection, pipeline, model, run history, source fixture, destination table, or delivered record was deleted. Secrets and connection strings are excluded from this report.
