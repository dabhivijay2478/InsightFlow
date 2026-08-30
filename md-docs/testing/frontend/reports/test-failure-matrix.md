# Full-Scale E2E Failure and Resolution Matrix

| Area | Observed failure | Root cause | Resolution | Final status |
| :--- | :--- | :--- | :--- | :--- |
| PostgreSQL scale | 8,000-row delivery stalled | Per-row conflict-aware remote inserts | Bounded 1,000-row multi-value upserts | Passed |
| SaaS validation | Source hints missing after discovery | Authoritative ELT catalog was not cached | Persist Stripe/HubSpot discovery metadata | Passed |
| Stripe delivery | Destination column mismatch | Dynamic account properties exceeded stable table contract | Explicit stable account projection | Passed |
| HubSpot invariant | Pipeline rejected with incomplete model coverage | Fewer output models than selected streams | One validated/published model per stream | Passed |
| HubSpot destination | Missing stream table | Destination tables were not all pre-created | Pre-create all 10 strict destination tables | Passed |
| HubSpot identifiers | Long property names failed PostgreSQL compatibility | Identifier length exceeded 63 characters | Sparse-safe compatible-name projection | Passed |
| HubSpot sparse data | Explicit exclusion referenced an absent runtime field | Sparse records omitted discovered properties | Regex-based compatible column selection | Passed |
| HubSpot pipeline stages | Upsert key failed | Catalog has no synthetic `id` | Composite `(pipeline_id, stage_id)` key | Passed |

The failed attempts are intentionally retained in UI run history. Each row above was followed by a successful real Chrome rerun.
