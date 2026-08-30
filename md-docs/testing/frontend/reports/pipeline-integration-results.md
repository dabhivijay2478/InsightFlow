# Full-Scale Pipeline Integration Results

All runs below were initiated and verified through the real Chrome UI. Resources remain available for inspection.

| Pipeline | Pipeline ID | Coverage | Latest outcome |
| :--- | :--- | :--- | :--- |
| Postgres Full Sync | `01132030-99a5-4885-b777-f1e4231daf73` | 6 tables / 31 type-bearing fixture columns | Success; all 6 extracted and 3 type rows delivered |
| Postgres Incremental Sync | `67c49a7b-f564-4fa4-a352-55d1bc799519` | 8,000 initial + 2,000 inserts + 25 updates | Success; 8,000 initial and 2,025 incremental |
| Stripe All Streams | `397fc4bf-3021-44bf-b2e2-18fa1e9b867d` | 34/34 streams | Success; 1 destination row; 38 seconds |
| HubSpot All Streams | `d6189dd7-5092-47b2-ae55-fd34aef1d952` | 10/10 streams and models | Success; 17 destination rows; 179 seconds |

## Destination verification

- PostgreSQL incremental destination: 10,000 rows retained.
- Stripe destination: representative `stripe_account` row retained after all 34 streams executed.
- HubSpot destination: 17 rows retained across owners, contacts, companies, deals, tickets, deal pipelines, ticket pipelines, products, line items, and quotes.
- Historical failed runs are intentionally retained with the successful reruns for diagnostic traceability.
