# DuckDB Source and Destination Connector Audit

Audit date: 2026-08-20

Production-readiness decision: **NOT READY**

DuckDB is explicitly enabled in the frontend catalog by user request. The local runtime primitives work, but this UI enablement does not add the managed-file lifecycle required to make a worker-produced `.duckdb` file a durable, organization-isolated product destination.

## 1. Frontend HEAD SHA

`cb58f15bef1eb488c55483ecbdf3b07e4e89308e` (`origin/main` after fetch; active branch baseline)

## 2. Go HEAD SHA

`fe1731bec5757e7255a0c7ada9c8d698b134f88d` (`origin/mantrixflow` after fetch; active branch baseline)

## 3. ELT HEAD SHA

`adefaf0cb78ac9860a65a0f33aa293ed417e6188` (`origin/mantrixflow` after fetch; active branch baseline)

## 4. Existing DuckDB support found

- The frontend already had one `duckdb` connector definition with source and destination metadata.
- ELT connector normalization already recognized `duckdb` and `duck_db` for both roles.
- The generic dlt SQL source builders already used `dlt.sources.sql_database` / `sql_table` and could be reused.
- The destination factory already selected dlt's native DuckDB destination, but defaulted to `:memory:`.
- The ELT environment already contained dlt 1.30.0, DuckDB 1.5.2, SQLAlchemy 2.0.49, duckdb-engine 0.17.0, and the staged internal DuckDB path.
- Existing strict delivery and disk guards were preserved.

## 5. Missing pieces found

- No managed file-artifact table, object-storage service, upload/select UX, or service-to-service materializer exists.
- No organization ownership check for a `file_ref` exists because there is no artifact repository.
- No exclusive artifact write lease, optimistic version check, stale-writer rejection, atomic publication, rollback, retention, or download contract exists.
- The frontend now exposes source/destination fields for an existing managed `file_ref`, but no upload/select artifact picker or create-output workflow exists.
- No browser-to-Go-to-ELT DuckDB E2E is possible.
- The public delivery path is intentionally upsert-only into pre-existing tables and supports PostgreSQL, MySQL/MariaDB, and Airtable—not DuckDB.
- Native dlt DuckDB loads create dlt state tables. Publishing that directly would violate the repository rule that customer destinations contain no `_dlt_*` artifacts.
- dlt 1.30 supports DuckDB `merge` strategies such as delete-insert, but does not list DuckDB among destinations that support the literal `upsert` merge strategy. Silently treating append or delete-insert as product UPSERT would be incorrect.

## 6. Frontend changes

- Reused the existing `duckdb` definition.
- Added DuckDB to `ENABLED_CONNECTOR_IDS` and set `availability: "enabled"` so the card appears under Available Now and opens the connection route.
- Declared `cdcCapable: false`.
- Added Connection Name, Managed DuckDB File Reference, and Default Schema fields for both roles, including client validation and role-aware test payloads.
- No unsafe host, port, credential, or server-path form was added.

## 7. Go registry changes

- Added one source and one destination registry record with canonical ID/type `duckdb` and alias `duck_db`.
- Added source, destination, discovery, and preview capability metadata; CDC remains false.
- Added canonical normalization and default `main` namespace handling.
- Added validation at create, update, ad-hoc test, and connector-test boundaries.
- Public configuration accepts an opaque `file_ref` or a safe create-output file name only. Server paths, URLs, and materialized paths are rejected.
- Added full and incremental dispatch protection against using the same artifact as DuckDB source and destination.
- Registry write modes advertise only the repository-supported `UPSERT` contract; they do not falsely advertise native dlt append/replace as production delivery.

## 8. ELT changes

- Removed the DuckDB `:memory:` source/destination fallback and require a materialized file.
- Open source engines read-only and dispose connection-test, discovery, and preview engines.
- Added safe DuckDB error codes/messages that do not return paths.
- Added `main` namespace behavior, schema filtering, table/view discovery, and relation metadata.
- Added a local SQLAlchemy dialect compatibility layer for duckdb-engine 0.17 + SQLAlchemy 2.0.49 + DuckDB 1.5 reflection.
- Kept runtime health hard-disabled with the missing artifact-storage/lease/versioning reason.

## 9. dlt source implementation

The connector continues to use the generic dlt SQL source architecture. Real-file tests exercise `build_full_table_source` and `build_incremental_source`, which delegate to `dlt.sources.sql_database` / `sql_table` with batching rather than loading a database into a Python list.

Status: **PASS as an ELT primitive; FAIL as a production connector because artifact materialization is absent.**

## 10. dlt destination implementation

The destination factory uses `dlt.destinations.duckdb` with an explicit persistent file path supplied by trusted runtime materialization. It rejects missing paths and `:memory:`. Real-file tests exercise append, merge/delete-insert, and replace.

Status: **PASS as an isolated dlt primitive; FAIL as a product destination because durable publication is absent and native dlt metadata conflicts with strict customer-destination invariants.**

## 11. duckdb-engine usage

DuckDB sources use the duckdb-engine SQLAlchemy dialect. A local dialect subclass reflects schemas, tables, views, columns, and primary keys through DuckDB metadata functions/information schema, avoiding duckdb-engine's incompatible PostgreSQL `pg_collation` reflection query in the installed version combination.

## 12. File/artifact architecture

The public contract now reserves opaque references matching `file_...` and never accepts a browser-selected server path. The only path-bearing keys are internal ELT materialization inputs. The storage/materialization implementation itself is intentionally not fabricated: there is no current repository abstraction that can durably resolve or publish those references.

Required next architecture:

1. Organization-scoped artifact metadata and private object storage.
2. Authorized source download to a per-run temporary path.
3. Destination checkout to a working copy plus exclusive lease and expected-version token.
4. Close/checkpoint all writers, validate the result, upload a new immutable version, then compare-and-swap the current version.
5. Retain the previous good version on every failure and clean temporary files.

## 13. Hosted vs self-hosted path policy

- Hosted public API: opaque managed references only; arbitrary paths and file URLs are rejected.
- Trusted internal ELT: accepts `_materialized_path` after a future authorized materializer resolves the reference.
- Self-hosted explicit local-path mode: not implemented. Adding it requires an explicit deployment policy and must not weaken hosted validation.

## 14. Security/path traversal review

- **PASS:** Go tests reject `/tmp/...`, `../...`, connection URLs, SQLAlchemy URLs, database/path/file fields, invalid references, unsafe create names, and same-file source/destination.
- **PASS:** ELT failures return stable codes without the source path.
- **FAIL:** Artifact ownership, signed access, storage RLS, lease authorization, and cross-organization denial cannot pass until an artifact service exists.

## 15. Source read-only behavior

**PASS at ELT runtime.** DuckDB source connection, test, discovery, preview, and extraction engines use `read_only=True`. A real-file test verifies an insert fails and the file remains readable after all engine handles are disposed.

## 16. Discovery

**PASS at ELT runtime.** Real-file discovery returns `main` and an additional `analytics` schema, excludes system schemas, returns tables and views, columns, primary keys, and incremental-key candidates.

Production/browser result: **FAIL / not runnable** without file selection and materialization.

## 17. Preview

**PASS at ELT runtime.** Real-file preview reads selected rows, respects the limit clamp, handles Unicode, and releases the file handle.

Production/browser result: **FAIL / not runnable** without managed artifacts.

## 18. FULL_TABLE

**PASS at isolated ELT level.** A real DuckDB file is extracted through the generic dlt SQL source into another real DuckDB file with batch size 1 and verified rows/types.

Product E2E: **FAIL / not run.**

## 19. INCREMENTAL

**PASS at isolated ELT level.** A timestamp cursor loads initial rows, a new source row is inserted, and a second run produces three distinct destination IDs.

Failed-run recovery and product checkpoint publication: **FAIL / not tested** because the artifact lifecycle is absent.

## 20. APPEND

**PASS only for the isolated native dlt destination.** Two append runs create and extend a real DuckDB table. It is not exposed as a production MantrixFlow write mode under the strict upsert-only delivery invariant.

Production result: **FAIL.**

## 21. UPSERT/MERGE

**PASS only for native dlt merge with primary key using its default delete-insert strategy.** Updated and new rows are verified. This is not represented as literal dlt `upsert`, and missing-key product validation cannot be routed to DuckDB today.

Production UPSERT result: **FAIL.** No silent fallback is implemented.

## 22. REPLACE

**PASS only for the isolated native dlt destination.** Replace removes earlier customer rows and leaves the replacement set.

Production result: **FAIL** because safe artifact version publication and the repository's strict delivery semantics are unresolved.

## 23. Concurrency/write leasing

**FAIL.** Same-artifact source/destination dispatch is rejected, but no exclusive write lease, renewal, timeout, fencing token, or stale-writer denial exists.

## 24. Artifact versioning

**FAIL.** There is no immutable version record, expected-version token, compare-and-swap publication, retention, or stale artifact protection.

## 25. Artifact persistence

**FAIL.** The factory writes a persistent local file for isolated tests, but no output survives worker recreation as a MantrixFlow-managed artifact.

## 26. Failure rollback

**FAIL.** Native dlt transactional behavior is insufficient for product publication. There is no working-copy-to-new-version promotion that preserves the previous good artifact after load/upload/publication failure.

## 27. Disk-capacity handling

Existing ELT disk guards remain in place for staged runs. **FAIL for connector acceptance:** artifact checkout size, copy-on-write headroom, upload headroom, destination growth estimation, and cleanup/finalization failure scenarios are not integrated or tested.

## 28. Data type compatibility

Real-file tests cover BIGINT, `DECIMAL(38,10)`, timestamp, VARCHAR/list extraction, BLOB, Unicode, and NULL values. The compatibility dialect maps signed/unsigned integers, huge integers, floating point, decimal, text, binary, UUID, date/time/timestamp, interval, JSON, and arrays.

**Partial.** Deep STRUCT/MAP/UNION/ENUM round trips, overflow policy, timezone edge cases, NaN/Infinity, and cross-destination coercions are not exhaustively tested.

## 29. Large-file/performance behavior

The dlt SQL source and incremental builders stream/batch records, and the test forces one-row batches. **FAIL for acceptance:** no multi-gigabyte file, memory bound, worker restart, upload/download throughput, or copy amplification test was run.

## 30. Organization isolation

**FAIL.** Public paths are rejected, but there is no artifact repository against which Go can prove that a `file_ref` belongs to the authenticated organization.

## 31. Oria security

No DuckDB-specific Oria agent or prompt content was added. Paths are removed from ELT connection/discovery logs in favor of `managed-artifact`, display name, or opaque reference, and safe errors do not return paths. **Partial:** end-to-end proof that artifact content and storage credentials never reach Oria/OpenRouter requires the future artifact implementation and E2E tracing.

## 32. E2E results

Passed:

- DuckDB source file -> dlt generic SQL extraction -> real DuckDB test sink (FULL_TABLE and INCREMENTAL).
- Python records -> native dlt DuckDB file (append, merge/delete-insert, replace).
- Real-file test connection, multiple-schema/table/view discovery, preview, read-only enforcement, safe invalid-file handling.

Not run / failed acceptance:

- Browser -> Go -> ELT managed DuckDB source.
- Browser -> Go -> ELT -> durable downloadable DuckDB destination.
- DuckDB -> PostgreSQL, MySQL, or MariaDB.
- PostgreSQL, MySQL, or MariaDB -> DuckDB.
- Production DuckDB -> DuckDB with distinct managed artifacts.
- Concurrency, lease fencing, stale version, rollback, restart persistence, organization isolation, and large-file scenarios.

## 33. Known limitations

- Runtime health remains deliberately false even though the frontend catalog card is explicitly enabled.
- The artifact model and UX do not exist.
- Native dlt DuckDB state tables conflict with the repository prohibition on `_dlt_*` artifacts in customer destinations.
- The repository's public upsert-only, pre-existing-table delivery invariant conflicts with native dlt create/append/replace product semantics.
- Literal dlt `upsert` merge strategy is not supported by the installed DuckDB destination; tested merge uses delete-insert with a primary key.
- duckdb-engine reports that index reflection is unsupported.
- Cross-destination and performance coverage is absent.
- Frontend repository-wide lint has unrelated pre-existing failures; the changed connector registry file passes Biome.
- Existing maintained frontend files above 500 lines: `features/team/components/team-screen.tsx` (546) and `features/ai-copilot/server/agent/orchestrator.ts` (526). This change introduced none.

## 34. Production-readiness decision

**DUCKDB STATUS: NOT READY**

Frontend catalog enablement does not change the **NOT READY** production decision. Complete all of the following before treating DuckDB as an operational production connector:

- organization-scoped managed artifact upload/select/create/download;
- private storage authorization and ownership checks;
- per-artifact exclusive writer lease with fencing and version CAS;
- durable immutable versions and atomic publication/rollback;
- a sanctioned destination delivery design that reconciles native dlt metadata/write modes with strict ELT invariants;
- failed incremental recovery, disk-capacity, large-file, restart, and cross-destination tests;
- frontend connection UX and the full requested E2E matrix.

### Verification performed

- Frontend targeted format/check: pass.
- Frontend type-check: pass.
- Frontend production build: pass with existing DuckDB WASM dynamic-import warning.
- Frontend repository-wide lint: fail on 26 unrelated existing errors and 15 warnings; the changed file is clean.
- Go `go mod tidy`, `go fmt ./...`, `go vet ./...`, `go test ./...`, `go test -race ./...`, and `go build ./...`: pass.
- ELT `pytest`: 346 passed, 17 skipped, 44 subtests passed.
- Focused DuckDB real-file suite: 6 passed.
- Focused adjacent ELT regression suite: 46 passed plus 5 subtests.
- `git diff --check`: pass in all three repositories.
