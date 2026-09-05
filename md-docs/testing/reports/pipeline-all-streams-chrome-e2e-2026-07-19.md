# Pipeline All-Streams Chrome E2E Report — 2026-07-19

## Overall result

**PASS.** The complete new-layout workflow was exercised in live local Google Chrome. PostgreSQL → PostgreSQL, HubSpot → PostgreSQL, and Stripe → PostgreSQL all completed successfully after the issues found during testing were fixed. All requested HubSpot streams (10) and Stripe streams (34) are configured, published, delivered to Neon, and represented by destination tables, including valid zero-row streams.

The final pipeline list shows all three pipelines as `ready` and `Healthy`:

| Flow | Streams | Published transformations | Destinations | Latest successful run | Rows written | Data validation |
| --- | ---: | ---: | ---: | --- | ---: | --- |
| PostgreSQL → PostgreSQL | 3 | 3 | 1 | `e2141a2e-46b6-45f9-9ff6-3b741f0dec1e` | 9 | Passed |
| HubSpot → PostgreSQL | 10 | 10 | 1 | `8a0003ce-a748-437a-b309-0a18d461d587` | 17 | Passed |
| Stripe → PostgreSQL | 34 | 34 | 1 | `2db775d5-6556-46a6-b23b-910adeacab76` | 25 | Passed |

## Test environment details

| Item | Value |
| --- | --- |
| Test date | 2026-07-19 |
| Time zone | Asia/Kolkata |
| Workspace | `/Users/vijay.d/vijay.d/Vapps/incomplete/ai-bi` |
| Browser | Live local Google Chrome controlled through the Chrome integration |
| Frontend | Next.js 16.0.10, React 19, Bun 1.3.4 |
| Main API | Go / Fiber with embedded queue worker |
| ELT | Python 3.13, FastAPI, dlt, DuckDB, dbt-compatible SQL execution |
| Organization | `Salesforce-test` (`43b7828d-6247-4c66-8be7-b9fd558822dd`) |
| Test account | `mantrixflow.testing@yopmail.com` |
| Destination | Neon PostgreSQL connection `c6c98bbe-def1-4182-ab15-eee6701fbe94` |

## Servers and commands used

| Service | Command | URL | Final result |
| --- | --- | --- | --- |
| Frontend | `cd apps/arcyria-platform && bun run dev` | `http://localhost:3000` | Responding |
| API and worker | `cd apps/server/arcyria-server && go run ./cmd/server` | `http://localhost:5000` | Operational |
| ELT | `cd apps/server/arcyria-elt && .venv/bin/python -m uvicorn api.main:app --host 0.0.0.0 --port 8000 --workers 1 --loop asyncio` | `http://localhost:8000` | Healthy; accepts runs |

Final health evidence:

- Frontend: HTTP response OK.
- API: `ok=true`, `status=operational`.
- Database: operational, 36 ms health latency.
- ELT component: operational, 3 ms health latency.
- Queue: operational, `active_runs=0`.
- ELT worker: `status=ok`, `accepts_new_run=true`, `active_runs=0`, 29.06 GB available.

Regression commands and results:

- `cd apps/server/arcyria-elt && .venv/bin/python -m pytest -q` — **189 passed, 8 skipped, 36 subtests passed**.
- `cd apps/server/arcyria-server && GOCACHE=/tmp/mantrixflow-go-build go test ./...` — **passed**.
- `cd apps/arcyria-platform && bun run lint` — **439 files checked, no fixes required**.
- `cd apps/arcyria-platform && bun run build` — **production build and 46-page static generation passed**.
- `go run ./cmd/e2e_all_streams -mode verify -provider <provider>` — exact destination count and data-contract verification.
- `go run ./cmd/e2e_all_streams -mode yaml -provider all` — all three version 2 YAML exports and GitHub connection verified.

## Neon cleanup

The destination was cleaned before creating the new schemas and tables, as requested.

| Check | Before | After cleanup |
| --- | ---: | ---: |
| Existing tables in `public` | 38 | 0 |
| Existing data size | 19,603,456 bytes | 0 bytes |

No new test table was created until the cleanup completed. The later verification used only the three isolated schemas documented below.

## Connections tested

| Role | Name | Type | ID | Chrome test/discovery |
| --- | --- | --- | --- | --- |
| Source | `rds` | PostgreSQL | `9a31ebbf-9480-406a-8cad-db2594691d66` | Passed; 3/3 selected streams rediscovered |
| Source | `HubSpot Source` | HubSpot | `15cf4720-e7df-4878-be9b-b8b972460308` | Passed; 10/10 selected streams rediscovered |
| Source | `stipre test` | Stripe | `4aa964eb-5c2d-46cc-bb1d-f28459629ae5` | Passed; 34/34 selected streams rediscovered |
| Destination | `Neon dest` | PostgreSQL | `c6c98bbe-def1-4182-ab15-eee6701fbe94` | Active; DDL and upsert passed |

## Pipelines created

| Flow | Pipeline | Pipeline ID | Source schema/namespace | Destination schema |
| --- | --- | --- | --- | --- |
| PostgreSQL → PostgreSQL | `Chrome E2E RDS Mixed Types 20260719` | `8397d5f2-206f-4044-bca0-d2e5ba2bd596` | `mxf_e2e_p2p_source_20260719_streams` | `mxf_e2e_p2p_dest_rds_20260719` |
| HubSpot → PostgreSQL | `Chrome E2E HubSpot All Streams 20260719` | `7c55980f-2f6f-4db5-8c36-2e390c1d20db` | `hubspot` | `mxf_e2e_p2p_dest_hubspot_20260719` |
| Stripe → PostgreSQL | `Chrome E2E Stripe All Streams 20260719` | `7ce3044d-0a35-40e6-b20a-1c1911373209` | `stripe` | `mxf_e2e_p2p_dest_stripe_20260719` |

## Source streams and destination tables

### PostgreSQL mixed-type streams

| Source table | Destination table | Rows | Primary key | Result |
| --- | --- | ---: | --- | --- |
| `mixed_scalar_types` | `rds_mixed_scalar_types` | 3 | `id` | Passed |
| `mixed_temporal_types` | `rds_mixed_temporal_types` | 3 | `id` | Passed |
| `mixed_document_types` | `rds_mixed_document_types` | 3 | `id` | Passed |

### HubSpot — all 10 streams

Destination schema: `mxf_e2e_p2p_dest_hubspot_20260719`.

| Stream | Destination table | Rows | Key |
| --- | --- | ---: | --- |
| `contacts` | `hubspot_contacts` | 2 | `id` |
| `companies` | `hubspot_companies` | 1 | `id` |
| `deals` | `hubspot_deals` | 1 | `id` |
| `tickets` | `hubspot_tickets` | 1 | `id` |
| `owners` | `hubspot_owners_curated` | 1 | `id` |
| `deal_pipelines` | `hubspot_deal_pipelines` | 6 | `pipeline_id, stage_id` |
| `ticket_pipelines` | `hubspot_ticket_pipelines` | 4 | `pipeline_id, stage_id` |
| `products` | `hubspot_products` | 1 | `id` |
| `line_items` | `hubspot_line_items` | 0 | `id` |
| `quotes` | `hubspot_quotes` | 0 | `id` |

Final row-count vector: **`[2, 1, 1, 1, 1, 6, 4, 1, 0, 0]`**. Exact transformed-data validation returned `valid=true`.

### Stripe — all 34 streams

Destination schema: `mxf_e2e_p2p_dest_stripe_20260719`.

| Streams with data | Rows |
| --- | ---: |
| `account` | 1 |
| `customers` | 1 |
| `events` | 15 |
| `payment_intents` | 1 |
| `setup_intents` | 1 |
| `invoices` | 1 |
| `products` | 1 |
| `prices` | 1 |
| `plans` | 1 |
| `payment_links` | 1 |
| `files` | 1 |

All other configured Stripe streams correctly produced zero-row destination tables: `balance_transactions`, `charges`, `disputes`, `payment_methods`, `payouts`, `refunds`, `subscriptions`, `subscription_items`, `invoice_items`, `credit_notes`, `credit_note_line_items`, `coupons`, `promotion_codes`, `tax_rates`, `quotes`, `quote_line_items`, `checkout_sessions`, `checkout_session_line_items`, `reviews`, `early_fraud_warnings`, `file_links`, `webhook_endpoints`, and `tax_ids`.

All 34 destination table names use the `stripe_` prefix. Final row-count vector:

`[1, 0, 0, 1, 0, 15, 1, 0, 0, 0, 1, 0, 0, 1, 0, 0, 0, 1, 1, 1, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 1, 0, 0, 0]`

The vector sums to **25**, all 34 tables exist, and every row has a non-null mapped `id`; `VERIFY_DATA provider=stripe valid=true`.

## Data types tested

| Fixture | Types and edge cases |
| --- | --- |
| Scalar | Boolean, small integer, integer, big integer, numeric/decimal, real, double precision, fixed character, varying character, text, timestamptz; Unicode (`café`), nulls, and boundary values |
| Temporal | Date, time with microseconds, timestamp, timestamptz, interval; leap day, Unix epoch, future date, offsets, negative duration, and null interval |
| Document/binary | JSON, JSONB, text arrays, integer arrays, bytea transformed to text, nullable text, empty arrays, nested documents |
| HubSpot | Large discovered text schemas, nullable CRM properties, booleans, timestamps, JSON property payloads, simple and compound keys |
| Stripe | Text/IDs, booleans, integers, timestamps, JSON objects/arrays, nullable fields, populated and zero-row streams |

Chrome source preview confirmed three temporal records including leap-day, epoch, future, timezone, negative-duration, and null cases. HubSpot preview returned two contacts with sensitive email values masked at both top level and inside serialized `properties`. Stripe preview returned the account record with nested settings data.

## Transformations and SQL results

| Flow/model | SQL behavior | Validation result |
| --- | --- | --- |
| RDS scalar | Typed pass-through with all mixed scalar columns | 3 rows; Unicode and null assertions passed |
| RDS temporal | Typed pass-through of temporal columns | 3 rows; leap day and null-duration assertions passed |
| RDS document | Explicit projection; `CAST(binary_value AS VARCHAR)` | 3 rows; nested JSON and empty-array marker assertions passed |
| HubSpot owners | Explicit fields plus `UPPER(first_name) AS first_name_upper` | Every output row satisfies `first_name_upper = UPPER(first_name)` |
| Other HubSpot models | Explicit discovered runtime-normalized fields | All 10 delivered successfully |
| Stripe models | Explicit discovery-schema projections, not runtime `SELECT *` | All 34 delivered; populated and zero-row tables passed |

SQL validation, preview, publish, and execution were exercised in Chrome. A SQL statement ending with a line comment initially broke preview because the wrapper closing parenthesis was swallowed. The preview wrapper now inserts newlines around SQL; the same revision then previewed and published successfully.

## Transformation version results

- RDS document revision 1 exposed the initial issue; revision 2 was corrected and used successfully.
- In Chrome, a harmless SQL comment was saved as revision 3, validated, previewed, published, and used by successful run `e2141a2e-46b6-45f9-9ff6-3b741f0dec1e`.
- The revision list shows revision 3 as Published and revisions 1/2 as Superseded, preserving history.
- Stripe destination detail shows all 34 transformations as **Published v2**.
- HubSpot shows **10 models, 10 published, 0 drafts**.
- Stripe shows **34 models, 34 published, 0 drafts**.
- Run history records the exact published revision count used by each run: 3 for RDS, 10 for HubSpot, and 34 for Stripe.

Both the previous corrected RDS revision and the updated Chrome revision have successful nine-row runs, confirming old and updated versions work.

## YAML management results

All exports use `version: 2` and `kind: mantrixflow.pipeline`, contain source connection, every selected stream, destination-owned transformations and SQL, destination tables, write mode, and keys.

| Pipeline | Export bytes | Streams represented | Result |
| --- | ---: | ---: | --- |
| HubSpot all streams | 8,533 | 10 | Valid |
| Stripe all streams | 23,682 | 34 | Valid |
| RDS mixed types | 3,237 | 3 | Valid |

Live GitHub/YAML round trip:

1. Connected GitHub App account `dabhivijay2478`.
2. Linked dedicated repository `dabhivijay2478/mantrixflow-git-testing`, branch `main`.
3. Configured `mantrixflow/pipelines/chrome-e2e-rds-mixed-types-20260719.yaml`.
4. Pushed from Chrome; the service created review pull request [#3](https://github.com/dabhivijay2478/mantrixflow-git-testing/pull/3).
5. Squash-merged from Chrome.
6. Sync history displayed commit `2d3fe1f724a8894ad271618d037acbc9d3825a5e` as Current.
7. Pulled the YAML back and applied it transactionally.
8. Retested the RDS configuration and destination data: 3 streams, 3 published models, 1 destination, `[3,3,3]`, `valid=true`.

YAML apply tests also cover successful graph/schema-mirror persistence and forced-failure rollback, so a failed import leaves the database unchanged.

## GitHub integration status

**Connected and working.** The App installation has read access to metadata and read/write access to code and pull requests. Repository discovery returned 60 accessible repositories. The dedicated testing repository was used to avoid production repository changes.

The existing installation required GitHub account confirmation. After confirmation, the signed, expiring callback completed successfully. The repository link, push, review PR, merge, history, and pull all passed.

## Chrome new-layout results

- Login/session: authenticated test workspace loaded successfully.
- New pipeline validation: name and source requirements enforced; Create button disabled while required fields were empty.
- Pipeline overview: shared source, destination-owned transformations, destination, readiness checks, activity metrics, and validation action rendered.
- Source: 3/3 RDS, 10/10 HubSpot, and 34/34 Stripe streams configured; fields, modes, cursors, keys, preview, test connection, discovery, and save controls present.
- Destinations: table schema, all target tables, upsert mode, keys, connection status, enabled state, and last-run time present.
- Transformations: list, destination isolation, editor, validation, preview, revision history, publish confirmation, and published/superseded states passed.
- Runs: running, failed, partial-success, success, rows, failures, duration, destination, trigger, and published revision counts rendered.
- Loading states: disabled Run button, `Cancel run`, preview loading, `Pushing…`, `Pulling…`, and repository loading observed.
- Empty states: impossible pipeline search produced `No pipelines found`; zero-row source streams completed without failure; no-sync-history state transitioned to a commit history.
- Retry/error states: historical failed/partial runs remained visible; affected flows were fixed and rerun successfully.
- Navigation and persistence: tabs, list/detail navigation, search debounce, empty search, settings save/reload, refresh, rerun, and final listing passed.

The console emitted a realtime subscription warning and used its polling fallback; run state still updated correctly. No functional console error remained.

## Issues found, fixes applied, and retests

| Issue | Fix applied | Retest result |
| --- | --- | --- |
| New structured pipeline creation could return a generic/zero-stream payload | Attached structured configuration immediately after creation | New layout renders correct relationships/counts |
| Partial delivery callback could be marked completed | Added centralized delivery-status mapping; partial is `partial_success` | Stripe partial run rendered correctly, later full run succeeded |
| HubSpot discovery names/defaults did not match runtime output | Normalized owner/pipeline fields, merged defaults, and text property types | All 10 streams and 17 rows passed |
| HubSpot nested preview JSON leaked an email masked at top level | Recursively mask serialized `properties` payloads; added regression test | Raw nested email absent; masked email visible |
| Stripe zero-row streams had no ephemeral raw DuckDB table | Create missing zero-row staging tables only, never destination tables | 23 zero-row streams delivered successfully |
| Stripe runtime `SELECT *` included columns outside destination mappings | Generate explicit SQL from stable discovery schemas | All 34 tables and 25 rows passed |
| Static schema materialization caused an empty-stream PK constraint failure | Removed materialization behavior while retaining stable column hints | Corrected Stripe run succeeded |
| SQL ending with `--` swallowed preview wrapper closing syntax | Wrap preview SQL with surrounding newlines | Chrome preview/validate/publish v3 passed |
| Structured destination list showed Last run as Never after delivery | Update enabled `pipeline_destinations.last_synced_at` in callback | Chrome shows a real relative time after RDS run |
| GitHub config polling could overwrite unsaved form state | Hydrate form only when saved scalar config values change | Saved configuration remains stable across refresh/polling |
| GitHub PR Merge/Cancel disabled themselves while status was syncing | Allow PR actions while awaiting review; disable only during PR mutation | PR #3 merged successfully from Chrome |
| YAML apply was non-transactional and could discard rebuilt graph/mirror errors | Transactional apply, persisted graph, propagated mirror errors | Success and rollback regression tests passed |
| GitHub callback and empty-repository paths needed stronger safety | State-bound callback, safer logs/webhooks, neutral base marker, YAML review PR | Full GitHub tests passed; live PR workflow passed |
| Temporary public Neon data consumed available storage | Dropped all 38 existing public tables before testing | 19.6 MB reclaimed; isolated schemas created |

## Test matrix

| Test Case | Source | Destination | Expected Result | Actual Result | Status | Issue | Fix Applied |
| --------- | ------ | ----------- | --------------- | ------------- | ------ | ----- | ----------- |
| Service startup | All | PostgreSQL | All required services running | Frontend, API, database, queue, and ELT operational | Pass | Services were initially stopped | Started/restarted required services |
| Neon cleanup | Existing Neon data | Neon | Existing data removed before tests | 38 tables / 19.6 MB reduced to zero | Pass | Destination lacked space | Cleaned `public` before DDL |
| New pipeline required fields | All | PostgreSQL | Name/source required | Create disabled until required values supplied | Pass | None | None |
| New-layout relationship hydration | All | PostgreSQL | New pipeline shows saved source/streams/destination | Correct counts render after create/reload | Pass | Structured data initially absent from create response | Attached structured configuration |
| Connection test | PostgreSQL | Neon | Connection succeeds | Chrome test and rediscovery passed | Pass | None | None |
| Connection test | HubSpot | Neon | Connection succeeds | Chrome test and 10/10 rediscovery passed | Pass | None | None |
| Connection test | Stripe | Neon | Connection succeeds | Chrome test and 34/34 rediscovery passed | Pass | None | None |
| Source fields | PostgreSQL | Neon | Mixed fields/types available | 3 streams; scalar, temporal, document fields visible | Pass | None | None |
| Source fields | HubSpot | Neon | All available fields available | Large schemas plus runtime-normalized fields visible | Pass | Discovery/runtime drift | Normalized discovery |
| Source fields | Stripe | Neon | Stable fields for all streams | 34 streams configured | Pass | Runtime-only field drift | Stable discovery projections |
| Source preview | PostgreSQL | Neon | Fixture rows preview | 3 temporal edge-case rows displayed | Pass | None | None |
| Source preview | HubSpot | Neon | Records preview without PII leak | 2 contacts; nested and top-level emails masked | Pass | Serialized properties leaked email | Recursive masking |
| Source preview | Stripe | Neon | Populated stream preview | Account record displayed | Pass | None | None |
| Empty source streams | Stripe | Neon | Zero rows handled successfully | 23 zero-row tables created/delivered | Pass | Missing staging table | Ephemeral zero-row raw table |
| Destination DDL | PostgreSQL | Neon | 3 explicit tables created | 3 tables, `[3,3,3]` | Pass | None | None |
| Destination DDL | HubSpot | Neon | 10 explicit tables created | 10 tables, row vector verified | Pass | Quote/pipeline schema mismatch | Discovery fix |
| Destination DDL | Stripe | Neon | 34 explicit tables created | 34 tables, row vector verified | Pass | Empty-stream and runtime drift failures | Staging/schema/SQL fixes |
| Mixed scalar types | PostgreSQL | Neon | Values preserved | Unicode, null, numeric and boundary checks pass | Pass | None | None |
| Mixed temporal types | PostgreSQL | Neon | Temporal values preserved | Leap/epoch/future/offset/null checks pass | Pass | None | None |
| JSON/array/binary types | PostgreSQL | Neon | Values transform correctly | Nested JSON, empty arrays, bytea cast pass | Pass | None | None |
| SQL validation | PostgreSQL | Neon | Valid SQL accepted | Validated and published | Pass | None | None |
| SQL line-comment preview | PostgreSQL | Neon | Trailing comment previews | v3 preview completed | Pass | Wrapper syntax swallowed | Newline wrapper |
| HubSpot SQL transform | HubSpot | Neon | Uppercase owner field correct | `first_name_upper = UPPER(first_name)` | Pass | None after discovery fix | Explicit normalized fields |
| Stripe SQL transforms | Stripe | Neon | Every model maps stable fields | 34 published v2 models run | Pass | `SELECT *` drift | Explicit projections |
| Transformation versions | PostgreSQL | Neon | Previous and updated versions retained/work | Corrected v2 and Chrome v3 both have successful runs | Pass | v1 exposed initial issue | Fixed, saved, validated, published v2/v3 |
| Pipeline run | PostgreSQL | Neon | 9 records transferred | Latest Chrome run wrote 9, zero failures | Pass | Earlier partial result misclassified | Callback status fix |
| Pipeline run | HubSpot | Neon | All streams transfer | 17 written, zero failures | Pass | Initial quote schema failure | Discovery/runtime fix |
| Pipeline rerun | HubSpot | Neon | Upsert rerun is idempotent | Second success wrote 17; counts unchanged | Pass | None | None |
| Pipeline run | Stripe | Neon | All 34 streams complete | 25 written, zero failures | Pass | Two failures and one partial during diagnosis | Empty-stream/schema/SQL fixes |
| Partial status | Stripe | Neon | Partial delivery not shown as success | Historical run shown `partial success` | Pass | Completion mapping wrong | Central status mapping |
| Destination validation | PostgreSQL | Neon | Exact transformed data valid | `valid=true` | Pass | None | None |
| Destination validation | HubSpot | Neon | Exact transformed data valid | `valid=true` | Pass | None | None |
| Destination validation | Stripe | Neon | Exact mapped IDs/counts valid | `valid=true` | Pass | None | None |
| Run loading/success states | All | Neon | Running, cancel, success visible | States observed in Chrome | Pass | None | None |
| Error/retry states | HubSpot/Stripe | Neon | Failures visible and rerunnable | Failed/partial history visible; corrected reruns pass | Pass | Connector/schema defects | Fixed and reran |
| Destination last run | PostgreSQL | Neon | Latest run time shown | Relative time shown after run | Pass | Display remained Never | Structured timestamp callback |
| Save/edit/refresh | All | Neon | Configuration persists | Settings save/reload, tabs, refresh passed | Pass | None | None |
| Listing/search/empty state | All | Neon | Pipelines list and filters work | Three healthy pipelines; debounce and no-results state pass | Pass | None | None |
| YAML export | All | Neon | Complete version 2 files | 10/34/3 streams represented; exports valid | Pass | None after fixes | Version 2 destination-owned contract |
| YAML apply atomicity | All | Neon | Failure rolls back | Success and forced-failure tests pass | Pass | Non-transactional writes | Transaction wrapper and propagated errors |
| GitHub connection | RDS | Neon | App connected | Connected as `dabhivijay2478` | Pass | Existing install required confirmation/callback | Completed state-bound callback |
| GitHub repository | RDS | Neon | Test repo linked | `mantrixflow-git-testing` linked | Pass | Form polling risk found | Stable scalar hydration; authenticated harness fallback |
| GitHub push/PR | RDS | Neon | YAML goes through review | PR #3 created | Pass | None | Safe review branch path |
| GitHub merge controls | RDS | Neon | Awaiting PR can merge/cancel | Merge enabled and succeeded | Pass | Circular syncing disable | PR-only pending condition |
| GitHub pull/history | RDS | Neon | Merged YAML pulls/applies | Commit shown Current; pull returned idle | Pass | None | None |
| Post-YAML data retest | PostgreSQL | Neon | Pipeline/data unchanged | 3/3/1 relationships; `[3,3,3]`, valid | Pass | None | None |
| Full regression suites | All | Neon | Tests/lint/build pass | Python, Go, lint, build all pass | Pass | None | None |

## Remaining issues

No functional blocker remains for the three requested flows.

Non-blocking observations:

1. The frontend realtime run subscription logged a warning and used the implemented polling fallback; live state still updated correctly.
2. The production build retains the existing DuckDB WASM dynamic-dependency warning and a stale `baseline-browser-mapping` data notice; compilation, type checking, and page generation pass.
3. GitHub's existing-installation chooser did not automatically return through the Setup URL after account confirmation, so the already verified installation was completed through the same signed local callback. The organization is connected and the full repository/YAML workflow now works.

## Final status by pipeline flow

| Flow | New layout | All streams | Destination DDL | SQL/version | Transfer | Data validation | YAML/GitHub | Final |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| PostgreSQL → PostgreSQL | Pass | 3/3 | Pass | Pass | 9 rows | Pass | Push/merge/pull pass | **PASS** |
| HubSpot → PostgreSQL | Pass | 10/10 | Pass | Pass | 17 rows | Pass | Export pass | **PASS** |
| Stripe → PostgreSQL | Pass | 34/34 | Pass | Pass | 25 rows | Pass | Export pass | **PASS** |

**Final status: all three requested source-to-Neon pipeline flows are successfully tested in Chrome with the new layout and flow.**
