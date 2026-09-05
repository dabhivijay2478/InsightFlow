# MongoDB source and destination connector audit

Audit date: 2026-08-25

Decision: **AVAILABLE NOW**. This owner release classification records code
support; it is not independent production certification. The implemented paths
are production-oriented and the standalone MongoDB paths exercised below move
real data successfully. Manual UI validation and the complete deployment matrix
remain pending.

## 1. Frontend HEAD SHA

`2508dc6ac7bcf9dd74726c3f19a9dc0fd2364fb4` (`apps/arcyria-platform`, merged locally into `main`).

## 2. Go HEAD SHA

`69028a3d1ed783c5d214bd0f72a44d743135da21` (`apps/server/arcyria-server`, merged locally into `mantrixflow`).

## 3. ELT HEAD SHA

`686186cb83c69431a7c363124ecde3d9cdd01254` (`apps/server/arcyria-elt`, merged locally into `mantrixflow`).

The feature branches were merged locally by fast-forward: frontend into `main`, and both backend services into `mantrixflow`. No histories were rebased, force-pushed, or otherwise rewritten.

## 4. Pre-existing MongoDB code found

The frontend already had MongoDB type remnants, schema/preview response normalization, pipeline payload compatibility, icon support, and an explicit comment saying new MongoDB connections were disabled. Frontend and Go security code already recognized credential-bearing MongoDB URI patterns. Go had an alias allowlist entry but no connector registry definition or complete connection lifecycle. ELT had no MongoDB runtime implementation.

## 5. Branches inspected

The frontend baseline was inspected from `main`; Go and ELT baselines were inspected from `mantrixflow`, as required by the repository divergence notes. Existing connector implementations and the Airtable, Asana, DuckDB, and MariaDB feature branches were used as architectural precedents. The feature branch was created independently in each nested repository.

## 6. Connector-family architecture

The canonical ID is `mongodb`, aliases normalize to that ID, and the connector family is `document`. Direction is source and destination. MongoDB does not enter SQL or SaaS family branches. Central source/destination normalization and role-aware family resolution are in `core/connector_support.py`.

## 7. Frontend registration

The connector catalog now exposes an explicitly enabled Available Now MongoDB card with source and destination roles. A focused field definition collects connection name, full URI, and database. Create/edit/test payloads preserve a masked URI on edit, validate `mongodb://` and `mongodb+srv://`, and reuse the existing connection form. Explicit frontend availability was enabled only after the real MongoDB-to-PostgreSQL, PostgreSQL-to-MongoDB, and MongoDB-to-MongoDB tests passed, so a transient health-response failure no longer moves the connector back to Coming Soon.

The public documentation includes separate MongoDB source and destination setup
guides plus a MongoDB/PostgreSQL pipeline reference. The marketing website
catalog includes MongoDB as an Available Now bidirectional database connector.

Frontend restructuring audit: no pages, tables, dialogs, links, or shared table implementations were added; no duplicate implementations, empty catches, commented-out code, or dead code were introduced. The largest changed frontend file is `features/data-sources/types/data-sources.ts` at 499 lines. Pre-existing files over 500 lines are `features/team/components/team-screen.tsx` (546) and `features/ai-copilot/server/agent/orchestrator.ts` (526).

## 8. Go registry and capability

The connector registry declares canonical ID `mongodb`, family `document`, source/destination direction, dynamic discovery, inferred schema, preview, FULL_TABLE, INCREMENTAL, and CDC disabled. Public destination capability advertises UPSERT only, matching the strict MantrixFlow delivery invariant. The control plane validates roles and URI schemes, encrypts/decrypts the full URI as one credential, preserves a masked value during edit, dispatches source and destination options, and rejects the same logical database/collection target.

## 9. ELT document connector routing

Generic test, discovery, table discovery, preview, health, sync, source-builder, destination-builder, and DuckDB-staged runtime routes now resolve MongoDB through the document family. MongoDB source staging, transformation, checkpoint extraction, and destination delivery stay inside the existing five-phase runtime. Document destination delivery uses transformed DuckDB output and does not route through SQL destination helpers.

## 10. Official dlt MongoDB source version

Source base: the official `dlt-hub/verified-sources` MongoDB template fetched with `dlt 1.30.0 init mongodb duckdb` on 2026-08-25. The adapted source is `apps/server/arcyria-elt/document_sources/mongodb/verified_source.py`. Runtime dlt version is 1.30.0.

## 11. Source customizations

The verified-source contracts (`mongodb` and standalone `mongodb_collection`) are retained. MantrixFlow adds validated settings, bounded connection/socket/server-selection timeouts, bounded 1,000-document chunks, explicit cursor and client cleanup, loss-aware BSON conversion, deterministic cursor plus `_id` sorting, projection/filter guards, dynamic resource selection, and runtime checkpoint wiring.

## 12. PyMongo version

The dependency contract is `pymongo>=4.7.0,<5.0.0`; the verified environment used PyMongo 4.17.0.

## 13. Connection URI architecture

The connector accepts one opaque `mongodb://` or `mongodb+srv://` URI plus an optional explicit database. URI query options, authentication database, replica-set, read preference, and TLS options pass directly to PyMongo. The database is resolved explicitly or from the URI. The password is not split into independent UI fields.

## 14. Atlas support

Atlas-compatible SRV URI parsing and PyMongo configuration are implemented. Live Atlas connectivity is **NOT TESTED**, so Atlas is not included in the production-readiness claim.

## 15. TLS and network behavior

TLS, certificate, SRV, replica-set, and network options are delegated to PyMongo from the URI. Timeouts are bounded and classified into safe authentication, DNS/SRV, TLS, timeout, Atlas/network-access, and generic connectivity messages. Live TLS, replica-set, sharded, DNS/SRV, and firewall failure scenarios are **NOT TESTED**.

## 16. Credential security

The full URI is treated as a secret. Go encrypts it as a single credential value, masked edit values do not overwrite the stored URI, displays use a masked URI, callback errors redact full MongoDB URI forms, Python Pydantic secret fields use `repr=False`, and Python exception scrubbing removes complete URIs. Frontend structured-field policy marks `connection_url` sensitive and free-text redaction recognizes both MongoDB URI schemes. Connector tests verify that classified failures do not echo a URI or password. No test secret was committed.

## 17. Database discovery

Discovery pings the deployment, returns only accessible databases, respects an explicitly selected database, filters system databases by default, and closes the client. Database discovery passed against a real standalone MongoDB 8.0 container.

## 18. Collection discovery

Collections are discovered dynamically. Results distinguish normal collections, views, and time-series collections, include safe index metadata, and flag destination-writability. Empty collections are supported. Real tests covered a normal collection, an empty collection, a view, and a time-series collection.

## 19. Schema inference

Schema inference samples a bounded number of documents and reports observed BSON types, mixed types, nullability, missing counts, arrays, nested objects, `_id` primary-key semantics, index metadata, and sample size. It does not scan or materialize an entire collection.

## 20. BSON type handling

The converter recursively handles ObjectId, Decimal128, timezone-aware and naive dates, BSON Timestamp, Binary/bytes, Regex, DBRef, Code, MinKey, MaxKey, arrays, nested documents, booleans, integers, doubles, strings, nulls, and mixed schemas. Exotic BSON values use loss-aware canonical Extended JSON structures rather than lossy stringification.

## 21. ObjectId handling

ObjectIds normalize to stable 24-character strings for staging and SQL portability. When `_id` is the incremental cursor, a persisted string checkpoint is converted back to a BSON ObjectId before the next MongoDB query. `_id` remains the default primary/merge key.

## 22. Decimal128 handling

Decimal128 converts to Python `Decimal`, preserving exact decimal semantics through staging. Real MongoDB-to-PostgreSQL movement and the BSON unit suite exercised Decimal128 values.

## 23. Nested document normalization

Documents and arrays normalize recursively. dlt performs relational staging normalization where needed; transformed MongoDB destination rows are converted back to BSON-safe documents. Destination mapping strips every `_dlt*` top-level field, and real destination tests assert that internal dlt fields are absent.

## 24. FULL_TABLE

FULL_TABLE uses the verified source and bounded batches, stages into the per-run DuckDB workspace, then transforms and delivers through the existing runtime. It passed in real MongoDB-to-MongoDB movement and as the source side of the MongoDB destination test.

## 25. INCREMENTAL

INCREMENTAL uses an explicit cursor field and dlt state. A real MongoDB-to-PostgreSQL run loaded two initial rows, resumed from its returned checkpoint, loaded a later row, and produced three destination rows without duplicates.

## 26. Cursor strategy

Queries use an inclusive cursor boundary (`$gte` for ascending/max state, `$lte` for descending/min state), optional exclusive end bounds, and stable `(cursor, _id)` ordering. Missing cursor values follow dlt incremental behavior and selected cursor fields are forced into inclusive projections.

## 27. Compound and tied cursor handling

For a non-`_id` cursor, results are sorted by cursor then `_id`. The inclusive boundary intentionally replays tied boundary rows; downstream UPSERT absorbs the replay, preventing gaps and duplicates. This favors correctness over avoiding a small boundary replay. Arbitrary multi-field compound checkpoints are not exposed in the UI.

## 28. State recovery

State is extracted from the dlt pipeline before the transient DuckDB staging file is deleted and returned only through the existing successful-run checkpoint path. A failed run does not replace the caller's last successful state. Existing strict runtime recovery tests pass; a live MongoDB failure-injection recovery test remains a follow-up.

## 29. Custom dlt MongoDB destination

The destination is a real custom `@dlt.destination` backed by PyMongo, located in `apps/server/arcyria-elt/destinations/mongodb`. It consumes transformed DuckDB rows in bounded batches, uses one sequential load job, skips dlt internal tables/fields, performs index preflight checks, and closes its client.

## 30. APPEND behavior

Internal APPEND uses unordered `insert_many` in bounded batches and passed a real dlt custom-destination unit integration. APPEND is not advertised by the public control plane because MantrixFlow's public delivery contract is UPSERT-only.

## 31. UPSERT behavior

UPSERT requires one or more non-null merge keys for every row. It builds unordered PyMongo `UpdateOne(..., upsert=True)` operations. `_id` is immutable and uses `$setOnInsert` when it is present but is not itself a merge key. PostgreSQL-to-MongoDB and MongoDB-to-MongoDB reruns passed without duplicate records.

## 32. REPLACE behavior

REPLACE is **DISABLED** at settings validation and again defensively in the sink because a portable atomic collection swap cannot be guaranteed. No configuration flag can enable it.

## 33. Idempotency

Public MongoDB delivery is UPSERT-only. Real PostgreSQL-to-MongoDB delivery reran after updating a source row and retained exactly two destination documents with the changed value. Real MongoDB-to-MongoDB delivery used `_id` UPSERT semantics.

## 34. Bulk writes

APPEND uses unordered `insert_many`; UPSERT uses unordered `bulk_write`. Batch size defaults to 1,000 and is bounded. Bulk exceptions are converted to terminal messages with attempted, succeeded, and failed counts; duplicate-key failures receive a specific safe message. A missing supporting merge-key index produces a performance warning.

## 35. Document-size handling

Every outgoing document is BSON-encoded before write and rejected before network I/O if it exceeds MongoDB's 16 MiB document limit. The safe error identifies the collection and bounded record key. The size guard passed unit coverage.

## 36. MongoDB to SQL destination results

- PostgreSQL: **PASS**, real MongoDB 8.0 to PostgreSQL 17, including incremental resume.
- MySQL: **NOT TESTED**.
- MariaDB: **NOT TESTED**.
- ClickHouse: **NOT TESTED**.
- DuckDB: **NOT TESTED as a final connector destination**; DuckDB is exercised as the mandatory staging layer in every real run.

## 37. SQL source to MongoDB results

- PostgreSQL: **PASS**, real PostgreSQL 17 to MongoDB 8.0, including an idempotent update rerun.
- MySQL: **NOT TESTED**.
- MariaDB: **NOT TESTED**.
- ClickHouse: **NOT TESTED**.
- DuckDB: **NOT TESTED**.

## 38. MongoDB to MongoDB result

**PASS** against a real MongoDB 8.0 deployment, across different databases and collections. The destination contained the expected two records and no `_dlt*` fields. Same logical source and destination database/collection is rejected by the Go control plane.

## 39. Multi-tenant isolation

MongoDB runs retain the existing required organization ID, per-run ID, pipeline ID, isolated staging/work directories, encrypted connection record ownership, and control-plane authorization boundaries. Generic organization/RLS regression suites pass. A dedicated two-organization live MongoDB isolation scenario is **NOT TESTED**.

## 40. Oria integration

The connector uses the existing registry and pipeline contracts. The full connection URI is a sensitive structured field and the AI text redactor recognizes credential-bearing MongoDB URI forms, preventing it from being forwarded in Oria/OpenRouter context or vector-memory text. No MongoDB-specific agent tool receives the URI. A live Oria telemetry-capture test is **NOT TESTED**.

## 41. Secret-redaction results

Go encryption/redaction tests and Python URI/error-redaction tests pass. Static inspection covered callbacks, display DTOs, SSE-safe failures, structured audit/AI fields, and exception normalization. Test logs do not print the live URI. The full frontend lint command reports only unrelated pre-existing AI-copilot diagnostics; all ten changed frontend files pass Biome.

## 42. CDC and Change Streams status

**DISABLED / FUTURE**. The registry reports `cdc: false`; no Change Streams, resume tokens, replica-set requirement, or CDC UI is enabled. Supported source sync modes are FULL_TABLE and INCREMENTAL only.

## 43. Known limitations

- The complete MySQL, MariaDB, ClickHouse, and DuckDB cross-direction E2E matrix has not been run.
- Atlas, `mongodb+srv://` DNS connectivity, replica sets, sharded clusters, live TLS/certificate validation, and network-failure recovery have not been run against real deployments.
- Multi-field compound incremental checkpoints are not exposed; tied cursor boundaries are safely replayed and require UPSERT.
- Live failure injection for incremental state recovery and PyMongo partial bulk results remains untested, although both code paths and safe error contracts are covered by the broader suite.
- REPLACE and Change Streams are intentionally disabled.
- Full frontend repository lint is blocked by 24 errors and 15 warnings in unrelated pre-existing AI-copilot files. Changed-file lint, typecheck, and production build pass.

## 44. Production-readiness decision

**AVAILABLE NOW; independent certification pending.** Canonical registration,
document-family routing, extraction, destination delivery, discovery, preview,
schema inference, BSON conversion, sync modes, public Upsert, secret handling,
bounded processing, and the primary data-movement paths have automated evidence.
The missing topology, tenant-isolation, failure-injection, and manual UI checks
remain explicitly recorded.

## Verification record

- ELT full suite: `415 passed, 7 skipped, 44 subtests passed`; zero failures.
- Real E2E suite: `3 passed` against MongoDB 8.0 and PostgreSQL 17; zero failures.
- Go: `go vet ./...`, `go test ./...`, and `go build ./...` pass.
- Frontend changed-file Biome: 10 files pass.
- Frontend TypeScript: `tsc --noEmit` passes after Next generated types are present.
- Frontend production build: passes; existing DuckDB WASM dynamic-dependency and stale baseline-browser-data warnings remain.
- Frontend full lint: fails on 24 errors and 15 warnings in unrelated pre-existing AI-copilot files; none are in MongoDB-changed files.
- Public Mintlify docs: source guide, destination guide, and verified-path example added; `mintlify broken-links` passes.
- Marketing website: full Biome check, TypeScript, 15 tests, and production build pass. The largest changed website file is `lib/marketing-content.ts` at 483 lines.
- `git diff --check`: passes in all three nested repositories.
- Python compile audit: passes.
