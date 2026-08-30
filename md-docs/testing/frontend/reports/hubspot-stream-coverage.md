# HubSpot All-Stream Chrome E2E Coverage

Pipeline: `HubSpot All Streams` (`d6189dd7-5092-47b2-ae55-fd34aef1d952`)

Final result: **10/10 selected streams and 10/10 published models completed successfully**. The latest UI run delivered 17 rows in 179 seconds. All resources and history remain retained.

| Stream | Mode | Destination columns | Retained rows | Key |
| :--- | :--- | ---: | ---: | :--- |
| `contacts` | Incremental | 410 | 2 | `id` |
| `companies` | Incremental | 245 compatible | 1 | `id` |
| `deals` | Incremental | 195 compatible | 1 | `id` |
| `tickets` | Incremental | 233 | 1 | `id` |
| `owners` | Full table | 11 | 1 | `id` |
| `deal_pipelines` | Full table | 13 | 6 | `pipeline_id`, `stage_id` |
| `ticket_pipelines` | Full table | 13 | 4 | `pipeline_id`, `stage_id` |
| `products` | Incremental | 71 | 1 | `id` |
| `line_items` | Incremental | 141 | 0 | `id` |
| `quotes` | Incremental | 201 compatible | 0 | `id` |

## Compatibility coverage

- The strict invariant requiring one output model per selected stream was validated.
- `companies`, `deals`, and `quotes` contained discovered property names longer than PostgreSQL's 63-character identifier limit.
- Their published projections use a sparse-safe column-name filter and retain every runtime field whose identifier is PostgreSQL-compatible.
- Pipeline-stage streams use their real composite keys because the authoritative catalog does not expose a synthetic `id`.
- Empty line-item and quote streams completed successfully without fabricated records.

No credential value is recorded here.
