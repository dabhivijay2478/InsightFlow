# Strapi source connector audit

Date: 2026-08-20

## Status

**NOT READY** — a gated Beta implementation exists, but no live disposable Strapi 5 application or real MantrixFlow destination E2E was available. Mock HTTP tests, real dlt normalization into an isolated DuckDB file, connection-test behavior, preview behavior, and green repository suites are not treated as production proof.

The connector is source-only and read-only. No Strapi destination, reverse ETL, mutation, media download, schema mutation, admin API access, or GraphQL operation was added. At the user's explicit direction on 2026-08-20, Strapi is present in the frontend `ENABLED_CONNECTOR_IDS` set and appears under Available Now when runtime health reports the source available. This early UI enablement does not change the production-readiness result.

## Repository and upstream audit

All three nested repositories were fetched. Their `main` and `mantrixflow` histories were compared without merging or rewriting history. Work is on `feat/strapi-source-connector` in each nested repository.

| Repository | Pre-implementation HEAD | Selected baseline | Branch finding |
| --- | --- | --- | --- |
| Frontend | `a30abf35de18cf6f325d16646cfceadb8a0fc93c` | `main` | Current `main` contains the newer connector architecture and is 25 commits ahead of `mantrixflow`. |
| Go API | `6334d25a220196890742a5fee04b6310a4190fa9` | `mantrixflow` | Current `mantrixflow` contains the newer active control-plane architecture; histories are divergent. |
| Python ELT | `c93874ebd4f57c902a27cc319caa649db6120568` | `mantrixflow` | Current `mantrixflow` contains the newer staged SaaS runtime; histories are divergent. |

Development used dlt `1.30.0`. `dlt init strapi duckdb` was run only in a disposable directory. The generated source was pinned at upstream `dlt-hub/verified-sources` commit `3957506893a7da821dbcc6acd51c7ca4475d1f53` and exposed this signature:

```python
@dlt.source
def strapi_source(
    endpoints: List[str],
    api_secret_key: str = dlt.secrets.value,
    domain: str = dlt.secrets.value,
) -> Iterable[DltResource]:
    ...
```

The current [dlt Strapi documentation](https://dlthub.com/docs/dlt-ecosystem/verified-sources/strapi) lists Strapi under verified sources and documents the same signature. The generated upstream source package also labels the source community-maintained and not regularly tested. MantrixFlow therefore preserves the public signature and one-resource-per-endpoint graph, pins the inspected revision, and replaces the request helper with a security-hardened client. Runtime code never invokes `dlt init`. This is an adapted verified-source integration, not an unmodified upstream copy.

The current [Strapi 5 documentation](https://docs.strapi.io/) shows flattened Content API records with `id`, `documentId`, direct content fields, and `meta.pagination`. The connector detects the older `attributes` envelope and rejects it instead of silently normalizing it as Strapi 5.

## Capability contract

- Canonical ID: `strapi`
- Category: SaaS & APIs
- Direction: source only
- Discovery: configured-endpoint validation and authoritative ELT metadata
- Preview: supported and bounded to 50 root records
- FULL_TABLE: implemented using resource-scoped dlt `replace`
- Incremental: disabled pending live filter/key/state verification
- CDC: false
- Supported version: Strapi 5 only
- Release stage: Beta metadata, frontend-enabled and runtime-health gated
- Authentication: Strapi API Token using `Authorization: Bearer <token>`
- Recommended token: Read Only or custom `find`/`findOne` permissions

Go rejects a Strapi destination with `Strapi is currently supported as a source connector only.` The Go registry contains no destination entry. Python is the authoritative resource catalog, and no universal Strapi stream list is maintained in Go or the frontend.

## Credential and security contract

The trusted stored shape is:

```json
{
  "base_url": "https://cms.example.com",
  "credential": "<encrypted API token>"
}
```

Go canonicalizes frontend aliases once, encrypts `credential` with the existing connection encryption path, returns only the masked sentinel to the browser, and preserves the encrypted value when an edit submits the unchanged sentinel. The secret-removal path used for agent-visible configuration removes the credential. The source configuration and pipeline graph contain endpoint and runtime options only, never the API token.

The hosted-mode URL policy requires an HTTPS origin with no userinfo, path, query, or fragment. It rejects localhost, loopback, link-local, private, reserved, multicast, unspecified, and metadata-style hosts. Python resolves DNS before outbound requests and rejects the whole answer set when any address is non-public. Redirects must remain on the validated origin and under `/api/`, followed by another address validation. Deployment-level `STRAPI_ALLOW_INSECURE_HTTP` and `STRAPI_ALLOW_PRIVATE_NETWORKS` flags are explicit self-hosted escape hatches; hosted defaults remain restrictive.

Content endpoint input is not a URL. Only lowercase kebab-case Content API IDs such as `articles` are accepted. The server constructs `/api/{api_id}`. Paths, traversal, admin/plugin routes, query strings, arbitrary schemes, and full URLs are rejected.

The implementation covers the requested literal-IP, localhost, private-range, userinfo, file/FTP scheme, private DNS answer, and cross-origin redirect cases. A mocked DNS preflight cannot fully prove protection against every production DNS-rebinding/transport race; deployment egress controls remain a required defense in depth before readiness.

## Dynamic endpoints, connection test, discovery, and preview

Pipeline configuration stores endpoint objects in the frontend and canonical API IDs at the trusted Go boundary. Disabled entries are omitted and duplicates are removed. Only selected endpoints instantiate dlt resources, so unselected content types make zero normal extraction requests.

The connection test is two-stage:

1. Validate the URL/security policy and prove host reachability. A 4xx at the origin still establishes reachability.
2. If an endpoint is configured, request `/api/{api_id}` with `pageSize=1` to verify the token and that content type's `find` permission.

The result distinguishes `host_reachable` from `token_verified`; host reachability alone is never described as full source access.

Discovery validates only user-configured Content API IDs and returns authoritative Strapi 5 collection metadata. There is no route crawler. OpenAPI import and single-type discovery are not implemented. Preview uses the same hardened client and response-shape rules, is restricted to configured content types, caps output at 50 records, does not persist dlt state, does not add `populate=*`, and does not fetch media binaries.

## Extraction and normalization behavior

Collection requests use explicit page/pageSize pagination and returned `meta.pagination.page`/`pageCount` as authority. The default page size is 25, configurable up to 100. Request count, page count, timeout, and retry budgets are bounded. A 250-record mocked collection test proves three pages, all 250 records, no duplicates, and the correct final record.

The default request includes `status=published`; draft extraction is not advertised. Locale is preserved when Strapi returns it. Discovery reports `documentId + locale` only when both fields are present, or `documentId` when only that field is present, but this is metadata rather than a verified merge contract because incremental/merge mode is disabled. Numeric `id` remains in the source record.

Nested objects and arrays are passed unchanged to dlt normalization. A real dlt-to-DuckDB normalization test covers a Strapi 5 record containing an SEO component and ordered dynamic-zone blocks, and verifies that a child table is created. The connector does not stringify nested payloads or download media URLs. Relations, media metadata, rich-text blocks, custom data types, field removal/type-change behavior, locale variants, and schema evolution rely on generic dlt/staging behavior but have not been validated against a live Strapi instance.

FULL_TABLE applies `replace` to each selected resource only. Incremental extraction is deterministically rejected in Go and Python and is not offered by discovery. No `updatedAt` cursor, merge key, or failed-incremental checkpoint claim is made. This prevents an unverified Strapi 4/5 filter or identity strategy from entering production.

## Resilience and observability

The client retries only 429, 502, 503, 504, timeouts, and other network errors. Retries are bounded, use exponential backoff with jitter, and honor `Retry-After`. It does not retry 400, 401, 403, or 404. Redirects are bounded separately and do not consume the network retry allowance.

Safe process-local counters cover connection tests, endpoint validations, requests, rate limits, retries, extracted records, per-resource counts, and the existing staged failure path. Errors expose stable codes and sanitized messages and never include request headers, tokens, response bodies, or customer content.

Organization ownership continues to use the existing Go-authenticated connection, discovery, preview, and sync paths. Browser-to-Strapi and browser-to-Python paths were not added. The existing agent sanitizer is covered by a Strapi token-exclusion test. A complete live redaction sweep across browser logs, SSE, PostHog, audit events, Oria, OpenRouter, and vector memory was not possible and remains a readiness blocker.

## Code inventory

New ELT package:

- `saas_sources/strapi/__init__.py`
- `saas_sources/strapi/adapter.py`
- `saas_sources/strapi/client.py`
- `saas_sources/strapi/discovery.py`
- `saas_sources/strapi/errors.py`
- `saas_sources/strapi/security.py`
- `saas_sources/strapi/settings.py`
- `saas_sources/strapi/verified_source.py`

New focused tests:

- `apps/server/elt-server/tests/test_strapi_connector.py`
- `apps/server/main-server/internal/server/strapi_connector_test.go`

New frontend component:

- `apps/app/features/pipelines/operations/source/strapi-source-settings.tsx`

The frontend connector registry, connection form/schema/DTO mapping, icon, connection display, pipeline settings, discovery, preview, source types, and runtime gate metadata were extended. The Go connector registry, capability/role validation, encryption/display/edit preservation, connection and pipeline validation, runtime option dispatch, health, discovery, preview, and test routes were extended. The ELT SaaS allowlist/model/builder, generic routes, and DuckDB-staged disposition/metrics paths were extended. No new connector framework, normalization engine, table engine, Oria specialist, or destination was created.

## Test and validation evidence

- dlt source audit: `dlt init strapi duckdb` completed in a disposable directory at dlt 1.30.0; upstream revision pinned.
- ELT production and development requirements install: passed.
- ELT focused/impacted suite: 55 passed.
- ELT full suite: 384 passed, 17 skipped, 44 subtests passed. Warnings are existing Pydantic, DuckDB-engine, and dlt deprecations.
- Go focused Strapi tests: passed.
- Go full `go test ./...`: passed.
- Go full `go test -race ./...`: passed.
- Go `go vet ./...`: passed.
- Go `go build ./...`: passed; Go emitted a sandbox-denied module stat-cache warning after the successful build.
- Frontend `bun install --frozen-lockfile`: passed with no dependency or lockfile changes.
- Frontend changed-file Biome check: 22 files passed.
- Frontend TypeScript: passed.
- Frontend production build: passed with existing DuckDB-WASM dynamic-dependency and stale baseline-browser-data warnings.
- Frontend full lint: failed with 26 unrelated existing errors and 15 warnings in AI-copilot, onboarding, shared UI, and utility files; no Strapi-changed file was among them.
- Every changed frontend source file is at or below 500 lines. The largest is `features/data-sources/connection/schemas/saas-connection-schemas.ts` at 475 lines.
- Two unrelated maintained frontend files remain above 500 lines: `features/team/components/team-screen.tsx` (546) and `features/ai-copilot/server/agent/orchestrator.ts` (526).
- No frontend Playwright test was added or run because repository rules reserve frontend E2E testing for a separate scope; the requested live flow remains impossible without a live Strapi fixture and destination credentials.

Mocked Strapi tests cover source-only registration, URL and endpoint rejection, private DNS answers, redirect containment, safe redirect handling, host-only tests, sanitized 400/401/403/404/500 errors, bounded 429 handling with `Retry-After`, sanitized TLS failure, Strapi 5 flattening, explicit Strapi 4 rejection, 250-row pagination, resource selection, FULL_TABLE enforcement, and real dlt nested normalization. They do not cover the entire requested connection-reset/502/503/504, single type, live draft/publish, locale, relation, media, incremental, and schema-evolution matrix.

## Live and destination matrix

| Flow | Result |
| --- | --- |
| Live Strapi 5 authentication/read | NOT TESTED — no disposable Strapi application/token |
| Live Strapi 4 | NOT TESTED and not supported |
| Live draft/published behavior | NOT TESTED |
| Live locale and stable identity behavior | NOT TESTED |
| Live relation/component/dynamic-zone/media behavior | NOT TESTED |
| Strapi → PostgreSQL through MantrixFlow | NOT TESTED |
| Strapi → MySQL through MantrixFlow | NOT TESTED |
| Strapi → MariaDB through MantrixFlow | NOT TESTED |
| Strapi → ClickHouse through MantrixFlow | NOT TESTED |
| Strapi → customer DuckDB through MantrixFlow | NOT TESTED |
| Mocked Strapi payload → real dlt normalization → isolated DuckDB | PASS |

## Acceptance status and known limitations

Implemented and passing locally: canonical/source-only capability, dynamic validated endpoints, API-token form and canonical contract, encryption/masking/edit preservation, restrictive URL policy, authoritative ELT discovery, bounded preview, Strapi 5 shape preservation, explicit pagination, selected-resource FULL_TABLE, dlt normalization, feature gating, Go/Python registration, and deterministic destination/incremental rejection.

Readiness blockers:

1. No real disposable Strapi 5 test application was run, so actual authentication, permissions, pagination limits, draft/published semantics, locale variants, relations, components, dynamic zones, media metadata, custom types, and schema drift are unverified.
2. Incremental is intentionally disabled. `updatedAt` filtering, inclusive boundaries, `documentId + locale` identity, same-timestamp rows, and failed-run state recovery require real Strapi 5 tests before exposure.
3. Strapi 4, single types, explicit relation population, selected locales, draft ingestion, and OpenAPI discovery are not supported.
4. No live PostgreSQL, MySQL, MariaDB, ClickHouse, or customer DuckDB E2E was run. Staging/dbt/destination delivery therefore has no Strapi-specific production proof.
5. The full frontend lint baseline is not green, although all changed files, TypeScript, and the production build pass.
6. Complete multi-tenant adversarial and secret-redaction sweeps across every telemetry/agent surface were not run with a live connector.
7. Application-level DNS validation should be paired with hosted egress/network policy to close DNS-rebinding and transport-level SSRF races.

Production readiness remains **NOT READY**. The user explicitly enabled the Beta connector in the frontend before live E2E completion; runtime health still gates availability. Production promotion requires a disposable Strapi 5 fixture, completion of the live security/data-type matrix, and at least Strapi → PostgreSQL, Strapi → MySQL, and Strapi → MariaDB flows through the full MantrixFlow staging/dbt/delivery architecture.
