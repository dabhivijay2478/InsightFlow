# Real Chrome Full-Scale Test Summary

This report covers the requested live connector validation, not mocked Playwright coverage.

| Scenario | Result | Retained proof |
| :--- | :--- | :--- |
| PostgreSQL broad-type full sync | Passed | 6 tables, 31 type-bearing fixture columns, 3 output rows |
| PostgreSQL initial scale sync | Passed | 8,000 delivered rows |
| PostgreSQL incremental scale sync | Passed | 2,000 inserts + 25 updates; 10,000 final rows |
| Stripe all-stream sync | Passed | 34/34 streams, 1 output row |
| HubSpot all-stream sync | Passed | 10/10 streams/models, 17 rows |

## Execution method

- Login, connection creation, discovery, stream selection, model authoring, validation, publishing, run triggering, and final status inspection were performed in real Chrome against the locally running MantrixFlow stack.
- Direct read-only destination checks were used only to verify exact retained row and schema results after the UI runs.
- Failures were diagnosed, fixed, and rerun until the latest status was successful.

## Data retention

No cleanup was performed. Connections, pipelines, models, tables, data, and run histories remain available. Credentials and connection strings are excluded from reports.
