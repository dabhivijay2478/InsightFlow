# Asana source connector audit

Date: 2026-08-13

Production readiness: **READY for the documented PAT-based source workflow**. Deterministic validation is supplemented by a retained customer-UI run of all eight streams to PostgreSQL, MySQL, and Airtable. The connector remains source-only, PAT-only, and subject to the limitations documented below.

## Baseline revisions

| Repository | `mantrixflow` HEAD before implementation |
|---|---|
| Frontend (`cloud.mantrixflow.com`) | `0b3edf3b89d5dd64ac267b7db31d5b221b0bd732` |
| Go (`main-server-mantrixflow.com`) | `f68049f9beba45eb43a17cc129383110180577db` |
| ELT (`etl-server-mantrixflow.com`) | `a213a5fd6eb24099496d50f596cc47d77323e7df` |

All three repositories were fetched, checked out at `mantrixflow`, pulled, and then moved to `codex/feat/asana-source-connector`. No Asana connector implementation existed in the reviewed baseline.

## Initial architecture matrix

| Layer | Initial status | Existing | Missing before this change |
|---|---|---|---|
| Frontend connector registry | Missing | Consolidated connector model | Asana entry and enable gate |
| Source connection schema | Missing | Shared SaaS credential form | Opaque PAT field and mapping |
| Source picker | Missing | Capability-filtered catalog | Asana source card |
| Runtime capability gate | Partial | Health-backed runtime connectors | Asana runtime probe |
| Go connector catalog | Missing | Embedded source/destination registry | Source-only Asana catalog |
| Go validation | Missing | Connector/role validation | Asana allowlist and destination rejection |
| SaaS classification | Missing | Shared SaaS classification | `asana` |
| Credential encryption | Partial | Generic encrypted SaaS secrets | Canonical `access_token` aliases |
| Test connection | Missing | Go-to-ELT proxy | Real `users/me` call |
| ELT dependency | Missing | `dlt>=1.23,<2` | Compatible Asana SDK |
| dlt verified source | Missing | dlt runtime | Imported verified resource graph |
| Resource discovery | Missing | Authoritative ELT discovery path | Workspaces, projects, and stream metadata |
| Preview | Missing | Saved-connection preview proxy | Bounded Asana previews |
| FULL_TABLE | Missing | Shared dlt staging | Stateless full reads and dispositions |
| INCREMENTAL | Missing | dlt checkpoint extraction | Task `modified_at` state and `gid` merge |
| Pagination | Missing | None for Asana | SDK offset pagination at page size 100 |
| Rate limiting | Missing | None for Asana | SDK Retry-After handling and retry ceiling |
| Nested data | Missing | dlt normalization | Verified-source fields and child tables |
| Custom fields | Missing | dlt normalization | Project/task custom fields |
| Transformations | Partial | Shared DuckDB/dbt path | Asana admission to that path |
| Oria support | Partial | Generic organization-scoped connection context | Asana registry visibility and secret proof |
| Secret redaction | Partial | Generic config redaction | Asana-specific encryption/redaction tests |
| Integration tests | Missing | Other connector suites | SDK-boundary, dlt, Go, and frontend tests |

## Verified source and dependencies

Development inspection used `dlt init asana_dlt duckdb`. Production does not run `dlt init`.

- Upstream source: dlt verified source `asana_dlt`
- Upstream commit: `3957506893a7da821dbcc6acd51c7ca4475d1f53`
- Upstream timestamp: 2026-07-03
- Repository import: `saas_sources.asana.verified_source`
- dlt contract retained: `@dlt.source`, `@dlt.resource`, `@dlt.transformer`, `dlt.defer`, resource dependencies, dispositions, task incremental cursor, and task primary key
- Python dependency: `asana==3.2.2`

The upstream template declared `asana<5.0.0`, but its `Client.access_token(...)` integration matches the stable 3.x SDK API. The connector pins 3.2.2 instead of allowing a future incompatible major version.

## Product and control-plane implementation

The frontend registers canonical ID `asana`, display name `Asana`, category `saas`, dedicated Asana brand icon, runtime availability, source capability only, and CDC disabled. The connection form accepts an opaque Personal Access Token; it does not validate a prefix. Test and persistence payloads use canonical `access_token`.

The source tab stores workspace/project scope on the pipeline source, not on the shared credential. Users can select one workspace, one or more projects, archived-project visibility, and an optional initial task-sync timestamp. Discovery refreshes projects when workspace or archived visibility changes. Stream sync choices are restricted to authoritative `supported_sync_modes`.

Go adds `asana` to connector validation and SaaS classification, exposes one source catalog entry and no destination entry, encrypts and masks the PAT, preserves a masked PAT during edits, and rejects destination use with:

`Asana is currently supported as a source connector only.`

Pipeline dispatch validates the selected resources and applies per-resource defaults. Full-table-only resources cannot be configured as incremental. Task incremental mode sets `modified_at` as the replication key and `gid` as the primary key. Workspace/project options are forwarded only at discovery/preview/run time for the organization-scoped saved source.

## Authentication and security contract

- MVP authentication: Asana Personal Access Token
- Token semantics: permissions are those of the Asana user who created the PAT
- Storage: encrypted by the existing Go connection encryption layer
- Display/edit: masked as `***`; masked edits preserve the encrypted stored value
- Logging/callbacks: Python maps SDK failures to sanitized connector errors and scrubs configured secrets
- Oria/OpenRouter: agent-visible sanitization removes `access_token` and the token value; only connector metadata, status, and sanitized discovery summaries are exposed
- Browser: the PAT is sent to the Go API, never directly to Python ELT or Asana
- Mutations: none; the connector exposes no destination or Asana write API

## Discovery, preview, and resource inventory

Connection testing performs a real SDK `users.me` request. Discovery lists accessible workspaces. Projects are fetched only after a workspace selection and respect archived-project visibility. Preview uses the same SDK client/error mapping and caps responses at 50 records.

| Resource | Parent | Verified disposition | Primary key | Cursor | Product modes |
|---|---|---|---|---|---|
| `workspaces` | — | replace | — | — | FULL_TABLE |
| `projects` | workspaces | replace | — | — | FULL_TABLE |
| `sections` | projects | replace | — | — | FULL_TABLE |
| `tags` | workspaces | replace | — | — | FULL_TABLE |
| `tasks` | projects | merge | `gid` | `modified_at` | FULL_TABLE, INCREMENTAL |
| `stories` | tasks | append | — | — | FULL_TABLE |
| `teams` | workspaces | replace | — | — | FULL_TABLE |
| `users` | workspaces | replace | — | — | FULL_TABLE |

Dependency resources remain in the dlt graph when only a child is selected, but only selected public resources are loaded. Workspace and project filters constrain the parent graph before child extraction. Multi-project task duplicates converge through the task `gid` merge key.

## Full-table and incremental behavior

Task incremental extraction passes dlt state into the official `modified_since` request using cursor `modified_at`. The initial value defaults to `2010-01-01T00:00:00.000Z` and may be overridden per pipeline. Incremental state becomes authoritative only when all modeled destination deliveries complete; failed or partial Asana delivery returns the previous checkpoint and commits no stream.

Task FULL_TABLE binds a bounded, stateless dlt incremental range and applies replace semantics. This retains the official resource signature while preventing a previous full run from advancing state and turning later full runs into partial reads.

Stories depend on a complete task traversal. A combined tasks-plus-stories run must configure tasks as FULL_TABLE; the ELT builder rejects incremental tasks with stories instead of silently producing incomplete story history.

Asana task `modified_at` does not change for every project/container relationship move. Task incremental synchronization is therefore scheduled incremental extraction, **not CDC**. Periodic full reconciliation is recommended where complete membership/relationship accuracy is required.

## Pagination, retries, and concurrency

The shared Asana client sets page size 100, follows SDK-provided offset iterators until exhaustion, sets a bounded retry ceiling of five, and uses the SDK's exact `Retry-After` handling for HTTP 429 responses. HTTP 401/403/404 failures are terminal and sanitized; 429, retryable server errors, and timeouts are marked retryable. Connector-wide request concurrency is bounded to 10 and may be lowered with `ASANA_MAX_CONCURRENT_REQUESTS`.

Process-local, secret-free counters include connection tests, discovery/API requests, rate-limit/retry events, source records, per-resource preview records, and pipeline failures.

## Nested data and custom fields

The verified source requests project/task custom fields plus task memberships, tags, followers, dependencies, dependents, assignee, parent, workspace, and related nested fields. dlt normalization preserves nested objects and arrays as JSON/normalized child tables in the existing staging path. The connector does not recursively fetch unlimited subtasks and does not fetch or download attachments.

## Error mapping

| Condition | Safe code | Retryable |
|---|---|---|
| 401 / invalid token | `ASANA_AUTHENTICATION_FAILED` | No |
| plan restriction | `ASANA_PLAN_RESTRICTED` | No |
| 403 | `ASANA_PERMISSION_DENIED` | No |
| 404 | `ASANA_RESOURCE_NOT_FOUND` | No |
| 429 after SDK retries | `ASANA_RATE_LIMITED` | Yes |
| 5xx/retryable SDK error | `ASANA_TEMPORARILY_UNAVAILABLE` | Yes |
| timeout | `ASANA_TIMEOUT` | Yes |
| unknown SDK failure | `ASANA_REQUEST_FAILED` | No |

## Validation results

Deterministic tests do not require an Asana token and mock the official SDK boundary, not the dlt source.

- Frontend changed-file Biome check: passed
- Frontend TypeScript: passed after moving a stale generated `.next/dev` validator cache to `/tmp/mantrixflow-next-dev-stale-20260811`
- Frontend production build: passed; existing DuckDB WASM dynamic-import and browser-baseline warnings remain
- Frontend Chromium Playwright: 3 passed (source-only catalog/direct route, opaque PAT test/save, masked PAT edit preservation)
- Go `go test ./...`: passed
- Go `go test -race ./...`: passed after rerunning outside the filesystem/network sandbox so existing `httptest` listeners could bind
- Go `go vet ./...`: passed
- Go `go build ./...`: passed (sandbox prevented only the module stat-cache write; build exit was successful)
- Go `go mod tidy`: passed with no dependency-file changes
- ELT Asana suite: 13 passed, 1 optional-real test skipped
- ELT full suite: 244 passed, 1 skipped, 10 failed. All failures were existing destination/database integration cases unable to connect to their ephemeral PostgreSQL, MySQL, MariaDB, or CockroachDB services (plus the dependent MySQL pipeline cases); the Asana suite passed independently.
- Optional environment-driven real Asana test: skipped because `ASANA_TEST_ACCESS_TOKEN`, `ASANA_TEST_WORKSPACE_GID`, and `ASANA_TEST_PROJECT_GID` were absent
- Retained customer-UI Asana run: all eight streams succeeded to Airtable (20,497 rows), PostgreSQL (20,491 rows), and MySQL (20,471 rows), with zero failed rows in each final destination run
- Direct MySQL verification after removing the diagnostic probe row: 1 workspace, 27 projects, 159 sections, 10 tags, 6,503 tasks, 13,769 stories, 1 team, and 1 user (20,471 total)
- Final connector-focused ELT suite after delivery fixes: 19 passed, 1 skipped; broader selected regression suite: 48 passed, 1 skipped

The repository-wide frontend lint command remains red because of unrelated existing AI-copilot/onboarding formatting, import-order, image, array-key, and label diagnostics. All files changed for this connector pass Biome.

## Known limitations and readiness decision

1. PAT only; OAuth and service-account authorization are future work.
2. No CDC. Relationship/container moves may require a full reconciliation.
3. Stories require a full task dependency traversal and cannot accompany incremental tasks in one run.
4. Attachments are not downloaded.
5. Subtasks are not recursively expanded beyond the fields returned by the verified task resource.
6. Rate-limit behavior is validated at the SDK/error boundary; the retained live run did not intentionally force an Asana 429 response.
7. Airtable delivery uses an existing selected table and explicit writable field mapping; runtime delivery does not create Airtable tables.

Decision: **READY** for Full Table use across all eight streams and documented task Incremental use. The verified public destinations are PostgreSQL, MySQL, and Airtable. OAuth, CDC, attachment download, recursive subtasks, and incremental tasks combined with stories remain outside the supported contract.
