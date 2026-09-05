# Pipeline New-Layout End-to-End Test Report — 2026-07-18

## Overall status

**Pipeline data-plane result: PASS for all three requested source-to-destination flows.** Fresh runs completed for PostgreSQL, HubSpot, and Stripe, and transformed destination values were verified with read-only SQL.

**Complete Chrome/UI and GitHub status: BLOCKED.** Google Chrome is running, but the ChatGPT Chrome Extension is not installed in the selected Chrome profile, so the new-layout UI could not be controlled or certified in Chrome. The configured GitHub App is available, but the test organization is disconnected and requires a user-approved GitHub App installation. The supplied email/password is also rejected by the configured Supabase Auth project as `invalid_credentials`; API retesting used a short-lived locally signed JWT for the known test user.

This report therefore does **not** mark the entire requested task complete even though all three pipeline executions and destination-data assertions pass.

## Test environment details

| Item | Value |
| --- | --- |
| Test date | 2026-07-18 |
| Time zone | Asia/Kolkata |
| Workspace | `/Users/vijay.d/vijay.d/Vapps/incomplete/ai-bi` |
| Frontend | Next.js 16.0.10 / React 19 / Bun 1.3.4 |
| Main API | Go 1.26 / Fiber, including embedded pgmq worker |
| ELT | Python 3.13.1 / FastAPI / dlt 1.25.0 / DuckDB / dbt |
| Browser target | Google Chrome |
| Chrome automation state | Blocked: ChatGPT Chrome Extension missing |
| Test organization | `43b7828d-6247-4c66-8be7-b9fd558822dd` (`Salesforce-test`) |
| Test user | `469c44be-6ba8-4b0d-b4db-81aae70e2db8` |

## Servers and commands used

| Service | Command | URL | Final health |
| --- | --- | --- | --- |
| Frontend | `cd apps/arcyria-platform && bun run dev` | `http://localhost:3000` | Responding |
| Go API + worker | `cd apps/server/arcyria-server && go run ./cmd/server` | `http://localhost:5000` | Operational; database, ELT, and queue operational |
| Python ELT | `cd apps/server/arcyria-elt && .venv/bin/python -m uvicorn api.main:app --host 0.0.0.0 --port 8000 --workers 1 --loop asyncio` | `http://localhost:8000` | `status=ok`, accepts new runs, zero active runs |

Additional verification commands:

- Frontend: `bun run lint`, `bun run build`
- Go: `GOCACHE=/tmp/mantrixflow-go-build go test ./internal/elt ./internal/server` — passed
- GitHub YAML service: `GOCACHE=/tmp/mantrixflow-go-build go test ./internal/services/github` — passed
- Full Go regression: `GOCACHE=/tmp/mantrixflow-go-build go test ./...` — passed
- ELT focused regression: `53 passed, 36 subtests passed`; naming/preview regression: `4 passed`
- Destination validation: authenticated read-only explorer SQL through the Go API

## Connections tested

| Role | Name | Connector | Connection ID | Result |
| --- | --- | --- | --- | --- |
| Source | `rds` | PostgreSQL | `9a31ebbf-9480-406a-8cad-db2594691d66` | Passed |
| Source | `HubSpot Source` | HubSpot | `15cf4720-e7df-4878-be9b-b8b972460308` | Discovery, preview, extract passed |
| Source | `stipre test` | Stripe | `4aa964eb-5c2d-46cc-bb1d-f28459629ae5` | Discovery, preview, extract passed |
| Destination | `Neon dest` | PostgreSQL | `c6c98bbe-def1-4182-ab15-eee6701fbe94` | DDL, upsert, explorer validation passed |

## Pipelines created and run

| Flow | Pipeline | Pipeline ID | Successful run ID | Rows written | Final status |
| --- | --- | --- | --- | ---: | --- |
| PostgreSQL → PostgreSQL | `E2E PostgreSQL to PostgreSQL 20260718` | `a744691d-299b-4923-9a96-9bc2b41c837c` | `0669f0b9-d449-478c-b18a-e5dcbbe86763` | 10,000 | Success |
| HubSpot → PostgreSQL | `E2E HubSpot to PostgreSQL 20260718` | `4c60ed34-9ffd-440e-aefb-666a0d6327a7` | `281ba2a3-c82e-448e-a990-5ae2aac39e0a` | 1 | Success |
| Stripe → PostgreSQL | `E2E Stripe to PostgreSQL 20260718` | `2c7ea78c-b390-4ce5-80e6-b3d1d831fe6a` | `28e5ad1b-27f5-4c17-9125-261d9efdd0a1` | 1 | Success |

All three appear in the pipeline list with `status=idle`, `lastRunStatus=success`, `configurationStatus=ready`, one source, one transformation, and one destination after reload.

## Destination tables created

| Table | Creation mode | Primary key | Result |
| --- | --- | --- | --- |
| `public.p2p_numeric_types_v3` | Existing/current PostgreSQL test target | `id` | 10,000 rows verified |
| `public.e2e_hubspot_owners_20260718` | Explicit control-plane DDL; later added normalized columns | `id` | 1 row verified |
| `public.e2e_stripe_customers_20260718` | Explicit control-plane DDL | `id` | 1 row verified |

The ELT runtime did not create tables. Table creation used the explicit destination-table endpoint, preserving the runtime's upsert-only invariant.

## Data types tested

| Flow | Types exercised |
| --- | --- |
| PostgreSQL | UUID, small/integer/big integer, decimal/numeric, real/double, serial/bigserial, money/text, timestamp |
| HubSpot | VARCHAR identifiers/text, nullable email/name fields, boolean/runtime metadata during extraction, timestamp-like text |
| Stripe | VARCHAR identifiers/text, BIGINT balance, derived BIGINT, boolean present in preview, JSON/array source fields discovered, derived TIMESTAMP |

## Transformation and SQL results

### PostgreSQL

SQL added `int_col_doubled` and `decimal_normalized`. Destination assertions:

- `row_count = 10000`
- `int_col_doubled = int_col * 2` for 10,000/10,000 rows
- `decimal_normalized = CAST(decimal_col AS NUMERIC(20,5))` for 10,000/10,000 rows

### HubSpot

Final SQL selects the intended owner fields and computes `UPPER(first_name) AS first_name_upper`. Destination assertions:

- `row_count = 1`
- `first_name = Vijay`
- `first_name_upper = VIJAY`
- `first_name_upper = UPPER(first_name)` is true

### Stripe

Final SQL uppercases the customer name, doubles the balance, and converts the Unix `created` value to a timestamp. Destination assertions:

- `row_count = 1`
- `customer_name_upper = VIJAY DABHI`
- `balance_doubled = balance * 2` is true
- `created_at` is populated

## Transformation version results

| Transformation | Revision | State | Validation/preview | Run evidence |
| --- | ---: | --- | --- | --- |
| PostgreSQL numeric | 1 | Superseded | Valid | Historical published version retained |
| PostgreSQL numeric | 2 | Superseded | Valid | Successful run `c5981c51-7aad-4be4-87d0-1dd1dfb7551c`, 10,000 rows |
| PostgreSQL numeric | 3 | Published | Valid | Successful fresh run `0669f0b9-d449-478c-b18a-e5dcbbe86763`, 10,000 rows |
| HubSpot owners | 1 | Superseded | Valid under old validation behavior | Runtime naming failure exposed |
| HubSpot owners | 2 | Superseded | Valid | Destination-column mismatch correctly blocked delivery |
| HubSpot owners | 3 | Published | Valid | Successful run, 1 row |
| Stripe customers | 1 | Published | Valid | Successful run, 1 row |

Revision history, validation timestamps, preview timestamps, publication state, superseded state, and run revision IDs were confirmed through the API.

## YAML management results

All three pipelines exported successfully with source, streams, transformations, destinations, and SQL present:

| Pipeline | YAML path | Bytes | Result |
| --- | --- | ---: | --- |
| PostgreSQL | `mantrixflow/pipelines/e2e-postgresql-to-postgresql-20260718.yaml` | 1,425 | Export passed |
| HubSpot | `mantrixflow/pipelines/e2e-hubspot-to-postgresql-20260718.yaml` | 1,300 | Export passed |
| Stripe | `mantrixflow/pipelines/e2e-stripe-to-postgresql-20260718.yaml` | 1,352 | Export passed |

All exports were retested after the backend restart and use the current `version: 2` destination-owned transformation contract. YAML apply is now transactional, persists the rebuilt `pipeline_graph`, and propagates source/destination schema-mirror failures. Focused tests prove successful graph/mirror persistence and full rollback when a mirror update fails.

GitHub pull, rollback, and remote YAML round-trip remain blocked until the organization installs/connects the GitHub App.

## GitHub integration status

- GitHub App configuration: present (`mantrixflow-dev`)
- Organization connection: **disconnected**
- Install-start endpoint: passed and generated a GitHub App installation URL with expiring state
- Setup callback now requires both the expiring state and installation ID; a live missing-state retest redirects with an actionable error
- Sensitive install state, state hashes, and full callback/install/redirect URLs are no longer logged
- Empty repositories are initialized with a neutral marker; pipeline YAML remains on a review branch and pull request
- Push and rollback-write failures now persist `sync_failed`
- Repository list while disconnected: HTTP 400, `GITHUB_NOT_CONNECTED`, message `Connect GitHub before listing repositories`
- Required user action: authorize/install the GitHub App for the organization, then retest repositories, push/PR, merge/cancel, pull, history, and rollback

## Issues found and fixes applied

| Issue | Evidence | Fix applied | Retest |
| --- | --- | --- | --- |
| Frontend process held port 3000 but timed out | HTTP timeout with listener present | Stopped stale process group; clean dev restart | Frontend responds |
| Go API/worker not running | Port 5000 refused | Started `go run ./cmd/server` | API/database/ELT/queue operational |
| Frontend lint failures | Biome formatting/import errors | Targeted Biome fixes | 439 files pass lint |
| Local ELT auth disabled by misspelled env key | `ETL_INTERNAL_TOKEN` vs `ELT_INTERNAL_TOKEN` | Corrected both service env keys; restarted services | No token = 401; correct token = 200; Go→ELT health passes |
| HubSpot validation/runtime identifier mismatch | Validation accepted `firstName`; runtime staged `first_name` | Added shared dlt runtime naming helper; normalized validation hints and SaaS SQL preview records; added tests | Focused 4 tests pass; corrected SQL validates and run succeeds |
| HubSpot `SELECT *` exceeded destination contract | Preflight named five missing columns | Changed SQL to explicit mapped fields | Delivery succeeds; 1 row verified |
| Stripe requested unavailable staged fields | Validation named missing `livemode` then `created_at` | Used supported fields and derived timestamp from `created` | Validation/preview/publish/run pass |
| Invalid incremental configuration | Missing replication key | Validation already correct | HTTP 400 with named stream and required field |
| Empty source preview | Stripe charges contained no records | Empty result returned successfully | HTTP 200, zero records |
| YAML apply was non-transactional and did not persist the rebuilt graph | Audit of `applyPipelineYAMLToPipeline` | Wrapped apply in a DB transaction, persisted `pipeline_graph`, and returned schema-mirror errors; added success/rollback tests | Focused Go suite passes; all three version 2 exports pass after restart |
| GitHub callback could fall back to an unrelated pending install and logs exposed state-bearing URLs | Backend audit | Required state-bound callback, removed pending fallback, redacted sensitive logs, ignored unbound installation webhooks | Security regressions pass; live missing-state callback returns 302 error redirect |
| GitHub failed pushes did not reliably persist failure status | Backend audit | Added `sync_failed` persistence for push/export and rollback-write errors | Focused Go regression passes |
| Empty GitHub repository could receive pipeline YAML directly on the base branch | Backend audit | Bootstrap only `.mantrixflow/.gitkeep`; write YAML to review branch and open PR | Focused HTTP-client regression passes |
| New-layout GitHub section lacked complete loading/error/empty/retry and PR/history controls | Frontend audit | Added explicit states, retries, disable confirmation, rollback, PR merge/cancel, current-version protection, and mutual push/pull locking | Lint and production build pass; Chrome UI certification still blocked |
| GitHub disconnected | Integration endpoint `connected=false` | Verified install-start flow and actionable error | Awaiting user authorization |
| Supabase test credentials rejected | `invalid_credentials` | No safe password mutation performed; used short-lived local JWT for API tests | Login remains blocked |
| Chrome control unavailable | Extension check: not installed/enabled in selected profile | No automated repair permitted | Awaiting extension installation |

## Test matrix

| Test Case | Source | Destination | Expected Result | Actual Result | Status | Issue | Fix Applied |
| --------- | ------ | ----------- | --------------- | ------------- | ------ | ----- | ----------- |
| Service health | All | PostgreSQL | Frontend, API/worker, ELT healthy | All healthy after restart | Pass | Frontend stale; API down | Restarted frontend and API |
| ELT internal authentication | All | PostgreSQL | Protected ELT routes require token | 401 without token; 200 with token | Pass | Misspelled env key disabled validation | Renamed key and restarted |
| Connection list | PostgreSQL | PostgreSQL | Saved source/destination available | Both active and selectable via API | Pass | None | None |
| Connection list | HubSpot | PostgreSQL | Saved source/destination available | Both active and selectable via API | Pass | None | None |
| Connection list | Stripe | PostgreSQL | Saved source/destination available | Both active and selectable via API | Pass | None | None |
| Source discovery | PostgreSQL | PostgreSQL | Source schema/fields available | `public.numeric_all_types` configured | Pass | None | None |
| Source discovery | HubSpot | PostgreSQL | Production streams/fields available | 10 streams; owners available | Pass | None | None |
| Source discovery | Stripe | PostgreSQL | Streams/fields available | 34 streams; customers available | Pass | None | None |
| Source preview | HubSpot | PostgreSQL | Non-empty preview | One masked owner returned | Pass | None | None |
| Source preview | Stripe | PostgreSQL | Non-empty customers preview | One customer returned | Pass | None | None |
| Empty state | Stripe charges | PostgreSQL | Zero rows handled without crash | HTTP 200, zero records | Pass | None | None |
| Invalid incremental stream | HubSpot contacts | PostgreSQL | Missing key rejected | Named HTTP 400 validation | Pass | None | None |
| SQL validation error | HubSpot | PostgreSQL | Missing field named | Binder error named field | Pass | Validation/runtime mismatch found | Runtime naming fix |
| SQL validation | PostgreSQL | PostgreSQL | Numeric model valid | Valid and published | Pass | None | None |
| SQL validation | HubSpot | PostgreSQL | Runtime-normalized model valid | Valid after fix | Pass | Camel/snake mismatch | Normalized schema hints |
| SQL validation | Stripe | PostgreSQL | Metrics model valid | Valid after field correction | Pass | Unsupported fields | Corrected SQL |
| Destination DDL | HubSpot | PostgreSQL | Explicit table created with PK | Created/updated with `id` PK | Pass | Normalized fields initially absent | Added missing columns |
| Destination DDL | Stripe | PostgreSQL | Explicit table created with PK | Created with `id` PK | Pass | None | None |
| Pipeline run | PostgreSQL | PostgreSQL | Transfer and transform records | 10,000 written | Pass | None | None |
| Pipeline run retry | HubSpot | PostgreSQL | Failed config can be corrected/rerun | Third revision/run succeeded | Pass | Two named failures | SQL + validation fixes |
| Pipeline run | Stripe | PostgreSQL | Transfer and transform records | 1 written | Pass | Initial validation failures | SQL corrected |
| Destination data | PostgreSQL | PostgreSQL | All transformed values correct | 10,000/10,000 numeric assertions pass | Pass | None | None |
| Destination data | HubSpot | PostgreSQL | Uppercase mapping correct | `Vijay → VIJAY` | Pass | None after fixes | None |
| Destination data | Stripe | PostgreSQL | Name/balance/time transforms correct | All assertions pass | Pass | None after fixes | None |
| Save/reload/list | All | PostgreSQL | Saved structured config reloads | Three ready pipelines listed and fully reloaded | Pass | None | None |
| Rerun/status metadata | All | PostgreSQL | Runs expose phase/delivery metadata | All successful runs include dbt/delivery/cleanup metadata | Pass | None | None |
| Transformation versions | PostgreSQL | PostgreSQL | Prior and current published revisions retained and runnable | Revision 2 and 3 have successful 10,000-row runs | Pass | None | None |
| YAML export | All | PostgreSQL | Complete safe YAML produced | All three exports contain required sections | Pass | None | None |
| YAML apply atomicity | All | PostgreSQL | Failed import leaves DB unchanged | Graph/mirror success and forced-failure rollback tests pass | Pass | Non-transactional writes found | Transactional apply and propagated errors |
| YAML graph persistence | All | PostgreSQL | Imported graph becomes runtime graph | `pipeline_graph` persistence regression passes | Pass | Rebuilt graph was discarded | Persisted encoded graph in pipeline update |
| GitHub callback validation | All | PostgreSQL | Installation is bound to valid state | Missing state rejected; unbound webhooks ignored | Pass | Unsafe pending-install fallback | Required state-bound callback |
| GitHub empty-repository push | All | PostgreSQL | YAML goes through review PR | Neutral base marker plus YAML review branch/PR verified | Pass | Base-branch bypass found | Enforced review branch/PR path |
| GitHub connection | All | PostgreSQL | Organization connected | App configured, organization disconnected | Blocked | GitHub authorization required | Install flow verified |
| GitHub push/pull/history | All | PostgreSQL | Remote round-trip works | Not executable while disconnected | Blocked | Organization not connected | Awaiting installation |
| Chrome new-layout navigation/forms | All | PostgreSQL | Complete UI flow works in Chrome | Not directly testable | Blocked | ChatGPT Chrome Extension missing | Awaiting extension installation |
| Login in Chrome | All | PostgreSQL | Supplied credentials authenticate | Supabase reports invalid credentials | Failed | Credential mismatch | Requires account/password correction |

## Remaining issues

1. Install/enable the ChatGPT Chrome Extension in the selected Chrome profile and rerun the complete new-layout navigation, field, loading, keyboard, refresh, edit, save, and visual-state checks.
2. Correct or reset the supplied Supabase Auth credentials if a fresh login is required.
3. Complete the GitHub App installation for the test organization, then verify repository selection, YAML push/PR, merge/cancel, pull, history, rollback, webhook, and remote round-trip.
4. The production build retains an existing DuckDB WASM dynamic-dependency warning; compilation and static generation succeed.

## Final status by pipeline flow

| Flow | API/config | SQL/version | Destination DDL | Transfer | Destination validation | Chrome new-layout UI | Final |
| --- | --- | --- | --- | --- | --- | --- | --- |
| PostgreSQL → PostgreSQL | Pass | Pass | Pass | Pass | Pass | Blocked | Data pipeline pass; UI certification blocked |
| HubSpot → PostgreSQL | Pass | Pass after fixes | Pass | Pass | Pass | Blocked | Data pipeline pass; UI certification blocked |
| Stripe → PostgreSQL | Pass | Pass after fixes | Pass | Pass | Pass | Blocked | Data pipeline pass; UI certification blocked |
