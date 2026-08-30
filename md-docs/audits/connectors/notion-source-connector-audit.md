# Notion Source Connector Audit

Audit date: 2026-08-13

## Production-readiness decision

**AVAILABLE NOW — source only; independent certification and manual UI
validation pending.**

The modern source implementation and deterministic local validation are
complete. Available Now is the owner's release classification; it does not
replace the outstanding credential-backed Notion-to-destination evidence.

## Repository baselines

| Layer | Branch | HEAD before changes | Upstream comparison |
| --- | --- | --- | --- |
| Frontend | `main` | `fabf48cb3f70a682edda37d169797761a7e64de0` | identical to `origin/main`; the repository's `mantrixflow` branch is 15 commits behind this current branch |
| Go control plane | `mantrixflow` | `4abebf0687085bb24376c55157054a8f771da08e` | identical to `origin/mantrixflow` |
| Python ELT | `mantrixflow` | `b27d985622652ab8146bc1b46cab20db24204943` | identical to `origin/mantrixflow` |

All three remotes were refreshed before the final upstream comparison. The parent monorepo pointer was `390645738f4a26baa42df1b226f12491e535217f`.

## Existing implementation reused

The existing `saas_sources/notion` package, dlt source entry points, SaaS runner dispatch, connection forms, icon, Go connector registry, preview route, discovery proxy, encryption layer, and DuckDB-staged ELT path were upgraded in place. No parallel `notion_v2` source was created.

The existing flow remains:

```text
Notion API → dlt resources → DuckDB staging → dbt → supported destination
```

Notion is not registered as a destination and no write-capable Notion endpoint is used.

## Official contract verification

The current official [Notion versioning documentation](https://developers.notion.com/reference/versioning) identifies `2026-03-11` as the latest API version. The implementation centralizes that value as `NOTION_API_VERSION` and sends it with `Authorization`, `Accept`, and `Content-Type` on every request.

Notion's [2025-09-03 upgrade guide](https://developers.notion.com/guides/get-started/upgrade-guide-2025-09-03) separates database containers from data sources. Production row extraction now uses:

```text
GET  /v1/databases/{database_id}
GET  /v1/data_sources/{data_source_id}
POST /v1/data_sources/{data_source_id}/query
```

The deprecated `POST /v1/databases/{database_id}/query` path and all `Notion-Version: 2022-06-28` production references were removed.

The current [official dlt Notion verified-source documentation](https://dlthub.com/docs/dlt-ecosystem/verified-sources/notion) still exposes only `notion_databases`. Its current [client](https://raw.githubusercontent.com/dlt-hub/verified-sources/master/sources/notion/helpers/client.py) and [database adapter](https://raw.githubusercontent.com/dlt-hub/verified-sources/master/sources/notion/helpers/database.py) still use the old version and deprecated query route. MantrixFlow therefore preserves dlt source/resource semantics while maintaining a local modern HTTP adapter.

## Authentication and credential security

- Canonical secret field: `access_token`.
- Compatibility input fields: `credential` and `api_key`, normalized once by the Go encryption boundary and the Python route boundary.
- Tokens are opaque; no prefix validation is performed.
- Connection setup uses `GET /v1/users/me` and returns bounded bot/workspace metadata.
- Go tests verify encryption at rest, UI masking, edit preservation, and removal from agent-visible context using `NOTION_SECRET_NEVER_LEAK_8737`.
- Python errors are deterministic and do not retain upstream response bodies, headers, or tokens.
- Authorization headers and tokens are never logged.

## Capability and discovery contract

```text
canonical id  = notion
source        = true
destination   = false
discovery     = true
preview       = true
cdc           = false
```

Go now treats Notion discovery as authoritative in ELT and no longer carries the synthetic `databases/pages/users` list. Frontend static Notion resource metadata was also removed. ELT discovery uses Search only to populate the selection UI, retrieves each real data source schema, and returns:

- database ID and name;
- data-source ID and name;
- created/edited/trash metadata;
- property IDs, names, and types;
- primary key, cursor, and supported sync modes.

The frontend discovery normalizer preserves the database/data-source hierarchy metadata. Search is never used for row synchronization, consistent with Notion's [search limitations](https://developers.notion.com/reference/search-optimizations-and-limitations).

## Resource model

### Dynamic data-source rows

Each selected data source becomes one dlt resource. Canonical identity is its data-source UUID; mutable titles are labels only. Stable resource names use:

```text
data_source_<32 lowercase UUID hex characters>
```

The DuckDB table is `notion__<resource>`. Rows preserve Notion's page payload, including `id`, timestamps, `in_trash`, URL, parent, properties, and editor/creator objects. The dlt primary key is `id`.

### Full table

FULL_TABLE queries the selected data source and applies per-resource `replace` semantics in DuckDB staging. It does not affect other selected data sources.

### Incremental

INCREMENTAL uses a real dlt incremental step on `last_edited_time`, queries with an inclusive system timestamp `on_or_after`, sorts ascending, and merges by `id`. Candidate state is restored/extracted through the existing dlt state path and is not committed when destination delivery fails or is partial.

### Pages and blocks

Standalone pages require explicit page IDs; Search is not used for page synchronization. Page metadata and block rows are separate resources. Block children use cursor pagination and bounded breadth-first traversal with page ID, parent block ID, depth, and sibling position. Default limits are 20 levels and 100,000 blocks.

### Users

`GET /v1/users` is implemented as a paginated FULL_TABLE dlt resource with primary key `id`. Permission failure is isolated to selecting the users stream.

## Pagination, query completeness, and budgets

- Maximum page size: 100.
- Cursors are passed back unchanged and never parsed.
- Deterministic test coverage: 250 unique rows over three pages.
- `request_status`/incomplete markers cause a hard failure.
- The documented 10,000-result cap also causes a hard failure rather than false success.
- Automatic timestamp-window partitioning is not yet implemented; operators must narrow the incremental window. This is a production-readiness limitation for very large data sources.
- Average request rate is limited to three per second.
- `Retry-After` is honored for 429 responses.
- Only 429, 5xx, timeout, and network failures receive bounded retry/backoff.
- Defaults: 4 retries, 100,000 requests/run, 3,600 seconds/run.

## Property and schema behavior

Data-source discovery returns the authoritative `properties` map as ID/name/type descriptors. Row payloads preserve typed nested values, so dlt can normalize title, rich text, numbers, selects, status, dates, people, files, formulas, relations, rollups, IDs, and verification fields without blindly stringifying them. Relation IDs and rollup payloads remain unresolved by default, and file binaries are never downloaded.

Schema evolution uses dlt's existing evolve behavior. Incompatible downstream model/destination changes continue to fail through the strict preflight column checks. Signed file URLs may expire. Expanded page-property fetching is not implemented and remains an opt-in future enhancement.

## Observability and Oria

Safe connector metrics cover connection tests, discovery, API requests, rate limits, retries, extracted rows/pages/blocks, incomplete queries, and pipeline failures. No metric labels contain tokens or property values.

No dedicated Notion agent was created. Existing connection, schema, pipeline, and sync-mode specialists consume safe discovery metadata; the Go agent sanitizer removes connection secrets before model-visible context.

## Validation results

### Frontend

- Changed-file Biome check: passed (7 files).
- TypeScript: passed (`bun run typecheck`).
- Production build: passed (`bun run build`), with the pre-existing DuckDB WASM dynamic dependency warning.
- Repository-wide lint: failed on 90 pre-existing formatting/lint findings outside the changed files.
- Largest changed frontend file: `features/data-sources/types/data-sources.ts`, 493 lines.
- Existing maintained files above 500 lines: `features/team/components/team-screen.tsx` (546) and `features/ai-copilot/server/agent/orchestrator.ts` (509).

No tables were added or migrated, no internal links were changed, no empty error blocks were added, and no commented-out implementation was introduced.

### Go

- Notion contract/security tests: passed.
- `go vet ./...`: passed.
- `go test ./...`: passed after allowing existing `httptest` cases to bind a local port.
- Scoped `go test -race` for Notion/Asana/Airtable connector tests: passed.
- `go build ./...`: passed; the sandbox emitted a non-fatal module stat-cache warning.

### Python ELT

- Notion suite: 17 passed, 1 optional live test skipped.
- Notion plus staging-source suite: 32 passed before the optional live-test addition.
- Full suite: 273 passed, 1 skipped, 41 subtests passed, 2 unrelated MySQL integration tests failed with their existing `MYSQL_OPERATION_FAILED`/`AttributeError` delivery failure.
- The Notion suite includes a real dlt → DuckDB two-run incremental merge test, pagination, inclusive cursor, query-cap, retry, error redaction, discovery, preview isolation, legacy resolution, blocks, and failure-safe checkpoint tests.

## Real E2E status

The following optional secrets were absent:

```text
NOTION_TEST_ACCESS_TOKEN
NOTION_TEST_DATABASE_ID
NOTION_TEST_DATA_SOURCE_ID
NOTION_TEST_PAGE_ID
```

Consequently, real Notion → PostgreSQL, MySQL, MariaDB, and Airtable tests were
not run. The connector is explicitly enabled as Available Now, while these
credential-backed flows remain pending verification and must not be described
as independent production certification.

## Known limitations and next gate

1. No real Notion or external-destination E2E evidence is available in this environment.
2. Queries reaching the Notion 10,000-result cap fail safely; automatic time-window partitioning is not implemented.
3. The generic pipeline picker preserves database hierarchy metadata but does not yet render a dedicated nested Database → Data Source selector.
4. Explicit standalone page selection is supported by the runtime DTO but has no dedicated frontend selector.
5. Expanded page-property-item retrieval is not implemented; large relation/rollup/people/rich-text properties retain Notion's inline response limitations.
6. Repository-wide frontend lint and two unrelated MySQL integration tests remain red outside this connector diff.

The next verification gate is a controlled live workspace test covering
discovery, schema, preview, FULL_TABLE, INCREMENTAL failure/retry, and real
delivery to PostgreSQL, MySQL, and MariaDB. The public connector remains
Available Now and source-only while that manual evidence is pending.
