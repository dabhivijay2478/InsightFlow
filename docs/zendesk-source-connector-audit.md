# Zendesk source connector audit

Date: 2026-08-20

## Status

**NOT READY** — a Beta implementation exists, but no live Zendesk-to-supported-destination MantrixFlow E2E run was possible without a Zendesk tenant and OAuth client. At the user's explicit direction on 2026-08-20, Zendesk is present in the frontend `ENABLED_CONNECTOR_IDS` set and appears in Available Now when both ELT runtime health and Go OAuth configuration report Zendesk available. Mocked tests, imports, or preview success are not treated as production proof.

The connector is strictly source-only and read-only. No Zendesk destination, reverse ETL operation, or Zendesk mutation was added.

## Repository and upstream references

The three repositories were fetched and their `main` and `mantrixflow` histories were inspected without merging or rewriting history. Work was based on each repository's newest connector implementation branch and placed on `feat/zendesk-source-connector`.

| Repository | Pre-implementation HEAD | Selected baseline |
| --- | --- | --- |
| Frontend | `7ae48da6a3d3deff51a269a0bf2025c1949f1533` | `main` |
| Go API | `f5209a70381dd755c00adf10a33df8332a677ff7` | `mantrixflow` |
| Python ELT | `2ed87a8cc4cba98b10dfb64f84ec63748c24eb43` | `mantrixflow` |

The implementation uses dlt `1.30.0` and vendors/adapts the official `dlt-hub/verified-sources` Zendesk source at commit `3957506893a7da821dbcc6acd51c7ca4475d1f53`. The source was inspected with `dlt init zendesk duckdb` in a disposable directory. Runtime code never runs `dlt init`.

The upstream source graph, API endpoints, pagination types, primary keys, incremental cursors, and resource write dispositions are retained. One integration setting is intentionally changed: external scheduler joining is disabled because MantrixFlow supplies explicit backfill dates and owns scheduling; dlt 1.30 otherwise rejects normal runs that lack an injected external interval.

## Product and capability contract

- Canonical connector ID: `zendesk`
- Category: SaaS & APIs
- Direction: source only
- Discovery: supported
- Preview: supported
- Incremental extraction: supported
- CDC: false
- Release stage: Beta, frontend-enabled and runtime-health gated
- Authentication advertised: OAuth only (`read` scope)
- Enabled service: Zendesk Support
- Conversations API / Chat: not used or advertised; it is a different live-chat API and does not replace the Support Ticketing API
- Talk: source code vendored, not advertised, and disabled pending resource and live tests

Go rejects a Zendesk destination deterministically with `Zendesk is currently supported as a source connector only.` Python is the authoritative Zendesk resource registry; Go contains no Zendesk stream list, and the frontend renders generic discovery metadata rather than hardcoding resources.

## Credential and security contract

Canonical stored configuration:

```json
{
  "subdomain": "acme",
  "auth_mode": "oauth",
  "access_token": "<encrypted>",
  "refresh_token": "<encrypted>",
  "scope": "read",
  "token_expires_at": "<RFC3339 timestamp>",
  "refresh_token_expires_at": "<RFC3339 timestamp>"
}
```

The browser sends only connection name and subdomain to the organization-scoped OAuth start endpoint. Go creates a short-lived, one-use, hashed state, exchanges the returned code using a confidential client, verifies the Bearer token against the Support API, and stores encrypted tokens. The ELT accepts only the canonical OAuth access token. Manual connection creation/test routes reject Zendesk credentials so there is no second API-token or pasted-token path.

`access_token` and `refresh_token` are encrypted at rest, decrypted only for authorized ELT calls, redacted from generic configuration, and excluded from agent-visible context. Go refreshes an access token within five minutes of expiry under a row lock and immediately persists both rotated tokens because Zendesk invalidates the old token pair. Python connector exceptions carry only stable codes and safe messages, and the SaaS runner scrubs the runtime credential before logs, callbacks, PostHog exception capture, or pipeline events.

Subdomains are converted to a single DNS label under the trusted `zendesk.com` suffix. The Go and Python boundaries reject HTTP, arbitrary hosts, paths, queries, fragments, userinfo, ports, IP literals, and reserved local/metadata labels.

Organization ownership remains enforced by the existing Go connection, discovery, preview, and pipeline routes. Zendesk dlt identity includes organization, source connection, pipeline, and Support service; dlt state remains separated per resource. The remote checkpoint is fetched by the server-authenticated pipeline ID, and a failed destination delivery retains the previous checkpoint rather than committing extracted candidate state.

## Support resource catalog

| Resource | Primary key | Cursor | Modes | dlt disposition |
| --- | --- | --- | --- | --- |
| `tickets` | `id` | `updated_at` | FULL_TABLE, INCREMENTAL | merge (replace for explicit bounded full-table run) |
| `ticket_events` | `id` | `timestamp` | FULL_TABLE, INCREMENTAL | append (replace for explicit bounded full-table run) |
| `ticket_metric_events` | `id` | `time` | FULL_TABLE, INCREMENTAL | append (replace for explicit bounded full-table run) |
| `ticket_fields` | none | none | FULL_TABLE | replace |
| `users` | `id` | none | FULL_TABLE | replace |
| `organizations` | `id` | none | FULL_TABLE | replace |
| `groups` | `id` | none | FULL_TABLE | replace |
| `brands` | `id` | none | FULL_TABLE | replace |
| `sla_policies` | `id` | none | FULL_TABLE | replace |
| `ticket_forms` | `id` | none | FULL_TABLE | replace |
| `ticket_metrics` | `id` | none | FULL_TABLE | replace |
| `satisfaction_ratings` | `id` | none | FULL_TABLE | replace |
| `tags` | none | none | FULL_TABLE | replace |

Selected resources are applied with dlt `with_resources`; unselected Support endpoints, Chat, and Talk are inactive and consume no API requests. The initial incremental product default is a bounded 90-day lookback. The pipeline UI supports an explicit initial date and optional exclusive end boundary. Currently, incremental-capable resources in one run must all use the same full-table/incremental mode; mixed modes must be scheduled as separate bounded runs.

Tickets retain custom fields as JSON with ticket identity, and dlt normalization remains responsible for nested arrays/objects and schema evolution. Ticket events and ticket metric events remain independent streams.

## Runtime behavior

- Connection test performs an authenticated read of `/api/v2/users/me.json` through the same Zendesk client.
- Discovery verifies the connection and returns service/resource metadata, primary keys, cursors, modes, recommendations, and permission status without secrets.
- Preview executes the selected official dlt resource with a 1–50 row limit, five-page cap, 100-request cap, and 60-second runtime cap. It does not persist the production pipeline state.
- Offset, cursor, incremental stream, and start-time pagination retain the verified-source semantics. Repeated pagination URLs fail safely.
- Request retries are bounded to connection/timeouts, 429, and 500/502/503/504. `Retry-After` is honored. 401, 403, 404, and other non-transient failures are not retried.
- Per-extraction request, page, and runtime budgets stop runaway jobs.
- Safe counters cover connection tests, requests, retries, rate limits, extracted records, per-resource records, and pipeline failures. Customer content and credentials are never metric labels.
- Existing DuckDB staging, dlt normalization, dbt SQL models, and generic destination delivery are reused. There is no Zendesk-specific destination logic.

## Code inventory

New ELT package: `saas_sources/zendesk/{__init__,adapter,api_helpers,client,credentials,discovery,errors,registry,settings,verified_source}.py`.

New connector tests: `apps/server/elt-server/tests/test_zendesk_connector.py`, `apps/server/main-server/internal/server/zendesk_connector_test.go`, and `apps/app/tests/connections/zendesk-connector.spec.ts`.

New integration files include the Go Zendesk connector/OAuth/state/refresh implementation, the frontend OAuth form/service and source settings, and the OAuth-only Python source package.

Existing files modified:

- Frontend connector, SaaS field, connection DTO/edit/display, source discovery/preview, pipeline source settings, and connector type registries.
- Go connector registry/capabilities, generic connection validation, encryption/decryption/display/edit preservation, pipeline dispatch identity/runtime options, and generic test routes.
- ELT SaaS model/builder/support list, generic sync/test/discover/preview/health routes, and transactional staged-run checkpoint handling.

No duplicate Zendesk catalog entry, custom table engine, Oria specialist, or destination implementation was added. The only Zendesk-specific HTTP endpoints are the OAuth start and callback required to establish the connection; generic manual credential-test/create/update endpoints reject Zendesk.

## Test evidence

Mocked tests keep the official dlt source/resource graph active. They cover:

- canonical credentials, subdomain normalization, and SSRF rejection;
- source-only capability and destination rejection;
- OAuth-only credential rejection, token encryption/masking/redaction, confidential authorization-code parameters, token exchange/rotation, and agent-context exclusion;
- connection test, discovery, preview, resource metadata, and resource selection;
- multi-page pagination, authoritative next links, repeated-link termination, budgets, 401/403/404/429/5xx classification, and safe errors;
- real dlt-to-DuckDB mocked-network loads: 100 initial tickets, 5 new tickets, 10 updated tickets, same-timestamp replay/deduplication, cursor advancement, and merge result;
- bounded full-table date filtering;
- custom fields, nested arrays/objects, nulls, HTML, multiline/long text, emoji, UTF-8, and non-English content;
- organization/connection/pipeline/service state identity and failed-delivery checkpoint rollback;
- a Playwright catalog contract proving Zendesk is selectable as a source, exposes only its OAuth connection form, and remains absent from destination selection when runtime health reports available.

Optional live ELT variables are `ZENDESK_TEST_SUBDOMAIN` and `ZENDESK_TEST_ACCESS_TOKEN`. The full application OAuth flow additionally requires `ZENDESK_OAUTH_CLIENT_ID`, `ZENDESK_OAUTH_CLIENT_SECRET`, and a registered callback URL. No credentials are committed.

Validation results:

- ELT Zendesk: 30 passed, 1 live OAuth test skipped, including isolated temporary DuckDB destinations.
- ELT full suite: 340 passed, 17 skipped, 44 subtests passed. Warnings are existing Pydantic/dlt deprecations plus dlt date-helper warnings inherited by the vendored source.
- Go focused Zendesk OAuth/connector and database migration packages: passed.
- Go full `go test ./...`: passed.
- Frontend TypeScript: passed.
- Frontend production build: passed with the existing DuckDB WASM dynamic-dependency warning and stale baseline-browser mapping notices.
- Frontend changed-file Biome check: 22 files passed.
- Frontend full lint: failed on 26 unrelated pre-existing diagnostics in AI-copilot/shared UI files; no Zendesk-changed file was among them.
- Zendesk Playwright source-only catalog/form contract was updated for OAuth-only fields. It was not run because repository rules reserve frontend E2E execution for a separate testing scope.
- File-length audit: every changed frontend source file is at or below 500 lines. The largest changed file is `features/data-sources/types/data-sources.ts` at 495 lines. Two unrelated existing files remain above the repository limit: `features/team/components/team-screen.tsx` (546) and `features/ai-copilot/server/agent/orchestrator.ts` (526).

## Live and destination matrix

| Flow | Result |
| --- | --- |
| Live Zendesk authentication/read | NOT TESTED — credentials unavailable |
| Zendesk → PostgreSQL through MantrixFlow | NOT TESTED |
| Zendesk → MySQL through MantrixFlow | NOT TESTED |
| Zendesk → MariaDB through MantrixFlow | NOT TESTED |
| Zendesk → ClickHouse through MantrixFlow | NOT TESTED |

## Known limitations and readiness blockers

1. A live Zendesk tenant and source-safe dataset are required to validate actual endpoint permissions, rate-limit headers, large-account pagination, and Zendesk payload drift.
2. The frontend was enabled at the user's explicit request before live E2E validation. At least one complete live MantrixFlow flow through staging, dbt, and a supported destination must still pass before declaring the connector production-ready. The requested four-destination matrix remains untested.
3. Conversations API/Chat and Talk are not advertised. Conversations API is for live chat messaging, not the Support ticketing extraction implemented here; either service needs a separately designed resource contract and live E2E coverage before exposure.
4. Mixed full-table and incremental modes among incremental-capable resources require separate runs.
5. Discovery reports per-resource permission as unknown until that resource is read; a selected inaccessible resource fails with a sanitized permission-specific error.
6. Existing API-token connections must be reconnected through OAuth. They intentionally fail instead of using a deprecated compatibility path.
7. The Playwright test verifies source-only catalog and form behavior. The requested live authorize/callback/discover/preview/run UI flow remains untested because a Zendesk OAuth tenant/client is absent.

Production readiness remains **NOT READY**. The next release step is to provide the optional Zendesk test credentials and a disposable destination, execute the live MantrixFlow E2E, and inspect secret-bearing surfaces before promoting the Beta connector to production-ready status.
