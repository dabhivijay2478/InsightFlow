# Airtable Source + Destination Connector Audit

## Status

**AIRTABLE STATUS: READY**

The connector implementation, deterministic suites, and the complete authenticated Chrome/UI matrix now pass. The matrix used saved local connections and an isolated Airtable destination table created for this verification; no credentials or customer values are recorded in this report.

## Repository baselines

| Repository | Required base | Implementation branch | HEAD before changes |
| --- | --- | --- | --- |
| Frontend (`apps/app`) | `main` | `codex/feat/airtable-source-destination-connector` | `a04a72531ad12f0bc7ee25d4cf2c6462424cd093` |
| Go API (`apps/server/main-server`) | `mantrixflow` | `codex/feat/airtable-source-destination-connector` | `daf0a5a19f5027ad49d875f4f754a088c3fb5dec` |
| ELT (`apps/server/elt-server`) | `mantrixflow` | `codex/feat/airtable-source-destination-connector` | `6d53fd5e7c42c3ff09ff54d0dd418af9014244a5` |

The frontend already contained an inactive Airtable catalog entry and legacy SaaS schema. Go had no complete Airtable capability/credential/runtime path. ELT had no production Airtable source or destination implementation.

## Architecture

### Source

- Uses the generated official dlt verified Airtable source (`airtable_source` / `airtable_resource`) backed by `pyairtable`.
- The MantrixFlow adapter preserves the official extraction/pagination path, normalizes Airtable records, adds `_airtable_record_id` and `_airtable_created_time`, and stages selected table IDs under deterministic DuckDB relation names.
- Discovery is dynamic: accessible bases, existing tables, fields, field IDs, types, writability, and FULL_TABLE capability come from Airtable metadata APIs.
- Preview is limited to 50 records. Source extraction is verified with 251 records across three API pages.
- Supported source mode: `FULL_TABLE` only. Incremental and CDC are deliberately disabled because the verified source does not provide a tested incremental contract here.

### Destination

- Uses a real dlt `@dlt.destination` reverse-ETL sink. There is no fake `dlt.destinations.airtable` import and the Airtable write occurs inside dlt's load step.
- Targets an existing base, existing table, and existing fields only. No destination schema/table/field creation occurs during a run.
- Supported destination mode: `UPSERT` only. Append, replace, and delete are rejected.
- Destination batches contain at most 10 records and loads run sequentially with one parallel load job.
- Stable Airtable field IDs are persisted per destination-owned transformation. Preflight resolves current field names, rejects missing/calculated/read-only fields, and requires every merge field to be mapped.
- Destination base/table selection is stored on the pipeline destination, never on the shared connection. Field mapping and merge keys are stored per transformation assignment.

## Capability model and frontend gating

- Canonical connector ID: `airtable`.
- Go registry: source and destination entries are present, both using dynamic discovery.
- Frontend badge/capability: Source + Destination, PAT authentication, CDC false.
- Frontend availability is runtime-gated. Airtable is selectable only when ELT health reports `source=true`, `destination=true`, and `available=true`.
- The source pipeline UI provides Connection → Base → Tables → FULL_TABLE selection.
- The destination UI provides Connection → Base → Existing Table → UPSERT → stable field mapping → explicit merge fields.
- The selected source `base_id` is normalized to `airtable_base_id` at the Go→ELT dispatch boundary; a regression test covers graph persistence through runtime dispatch.

## Credential contract and PAT scopes

- Canonical stored credential key: `access_token`.
- `credential` and `api_key` remain accepted only as backward-compatible input aliases and are canonicalized before encryption.
- Active UI terminology is Personal Access Token, not API key.
- The global connection form never asks for a base or table.
- PAT values are encrypted at rest, decrypted only for trusted ELT calls, masked on reads, and preserved during masked edits.
- Source tokens require record-read and schema-read scopes. Destination tokens require record-write and schema-read scopes. A bidirectional connection requires all relevant scopes and access to each selected base.

## Rate limits, retries, and observability

- Shared per-base rolling limiter: 5 Airtable API requests/second.
- Airtable upsert batch size: 10.
- `429` and `5xx` failures are retryable; `401`, `403`, `404`, and `422` are terminal and mapped to sanitized messages.
- Process-local safe counters are exposed only on authenticated ELT health: connection tests, discovery, source records, destination records/created/updated, API requests, rate limits, retries, and terminal errors.
- Counters contain no PATs, record contents, field values, or customer identifiers.

## Type compatibility

- Source rows preserve Airtable values as JSON-compatible scalars, arrays, and objects while adding stable record lineage columns.
- Destination coercion handles `Decimal`, dates, datetimes, Airtable-compatible lists/objects, and null values before calling `pyairtable`.
- Computed/read-only field types are discoverable for source use but excluded from destination mapping.
- Linked records and attachments are preserved in their Airtable API shapes; human-name lookup and binary attachment upload are intentionally outside this release.

## Security review

- No PAT is returned by connection display APIs.
- No PAT is stored in pipeline source/destination options, GitHub YAML, field mappings, or schema metadata.
- Airtable exception messages are sanitized before callback delivery; destination logs record exception types rather than record bodies.
- Existing agent/audit sanitizers remove `access_token`, token, secret, credential, and API-key fields. New tests verify canonical encryption/masking/edit preservation.
- Existing pipeline tables remain under their current tenant-scoped RLS policies. Only JSONB columns were added to existing org-owned pipeline tables, so no new RLS ownership pattern or public grant was introduced.

## Test results

### Deterministic CI

- Airtable-focused ELT/dlt tests: **14 passed**.
- Latest focused Airtable/dbt/staging regression set: **39 passed**.
- Full ELT pytest suite: **241 passed**.
- Full Go suite (`go test ./...`): **passed**.
- Full Go race suite (`go test -race ./...`): **passed**.
- Go vet (`go vet ./...`): **passed**.
- Go build (`go build ./...`): **passed**.
- Frontend changed-file Biome check: **passed**.
- Frontend TypeScript check: **passed**.
- Frontend production build: **passed** with the pre-existing DuckDB WASM dynamic-dependency warning.
- Repository-wide frontend lint: **not clean due to 90 pre-existing errors and 15 warnings outside the Airtable change**; the 22 changed Airtable/pipeline files pass Biome.
- Largest changed frontend file: **490 lines**. No changed frontend file exceeds 500 lines. Two unchanged pre-existing files remain above the limit: `team-screen.tsx` (546) and `orchestrator.ts` (509).

The deterministic tests cover capability/registry contracts, PAT encryption/masking/edit preservation, dynamic base/table/field discovery, preview normalization, 251-record pagination, FULL_TABLE extraction, stable field mapping, read-only rejection, 10-record dlt batching, UPSERT, retry classification, throttling, empty-table schema materialization, and fake Airtable→Airtable data movement through both dlt integrations.

### Secure real integration

The complete matrix was created, configured, previewed, published, and run through the authenticated local Chrome UI. Source previews were non-empty before the Stripe, HubSpot, PostgreSQL, and MySQL streams were selected.

| Route | Source stream | Destination | Final UI result |
| --- | --- | --- | --- |
| Airtable → PostgreSQL | Airtable Customers | isolated PostgreSQL table | **success — 11 written, 0 failed** |
| Airtable → MySQL | Airtable Customers | isolated MySQL table | **success — 11 written, 0 failed** |
| Airtable → Airtable | Airtable Customers | `MantrixFlow Matrix Sink 20260810` | **success — 11 written, 0 failed** |
| PostgreSQL → Airtable | `mantrix_source.customers` | isolated Airtable table | **success — 100 written, 0 failed** |
| MySQL → Airtable | `defaultdb.mf_matrix_source_20260810` | isolated Airtable table | **success — 3 written, 0 failed** |
| Stripe → Airtable | `stripe.customers` | isolated Airtable table | **success — 1 written, 0 failed** |
| HubSpot → Airtable | `hubspot.owners` | isolated Airtable table | **success — 1 written, 0 failed** |

The Airtable grid showed **119 records** after delivery: 3 default blank records plus 116 delivered records. Distinct `airtable-`, `postgres-`, `mysql-`, `stripe-`, and `hubspot-` merge-key prefixes were visible in the real table. Re-running the three Airtable-source destinations concurrently also completed successfully after dbt execution was isolated from dlt's process-global context.

Real-flow failures found and corrected during this verification included:

- source-only `/full` loading incorrectly requiring a destination;
- empty Airtable primary-field values breaking extraction instead of using the Airtable record ID;
- mixed-case Airtable stream names disagreeing with generated dbt source names;
- sparse Airtable fields and non-actionable dbt errors;
- targeted Airtable dispatch losing base/table and assignment mapping metadata;
- Airtable catalog refresh leaving an assignment pointed at the logical model instead of the selected table ID;
- concurrent destination dbt processes cross-selecting another run's locked DuckDB file.

## Known limitations

- Initial destination release supports existing bases/tables/fields and UPSERT only.
- Linked-record values must already be shaped as Airtable record-ID arrays; lookup-by-human-name is not performed.
- Attachment writes expect Airtable-compatible attachment objects/URLs; binary upload orchestration is not included.
- Incremental source sync, CDC, append, replace, delete, schema creation, and field creation are intentionally unavailable.
- `pyairtable~=2.1` is the verified compatible dependency range; installed test version was 2.3.7 with the existing dlt 1.x runtime.

## Documentation and website coverage

Public documentation now includes:

- Airtable source setup, Personal Access Token access, base/table discovery,
  stable staging names, Full Table behavior, verification, and troubleshooting;
- Airtable destination setup, existing-table requirements, writable field
  mapping, merge keys, Upsert semantics, type handling, and rate-limit errors;
- MySQL source setup, least-privilege grants, schema discovery, Full Table and
  Incremental modes, type guidance, and verification queries;
- MySQL destination setup, explicit table DDL, writer grants, Upsert behavior,
  duplicate checks, and destination troubleshooting; and
- an end-to-end Airtable → MySQL, MySQL → Airtable, and Airtable → Airtable
  guide with sample data, SQL models, field mappings, rerun checks, and the
  authenticated seven-route validation matrix.

The marketing website connector catalog now marks Airtable and MySQL as live
source-and-destination connectors, includes the Airtable brand icon, links live
connector cards to their setup guides, and removes PostgreSQL-only availability
copy from connector, comparison, navigation, and product sections.

## Production readiness decision

The connector is **READY** for the documented existing-table, UPSERT-only release scope. Deterministic suites pass, the full authenticated seven-route UI matrix passes, delivered records are visible in Airtable, concurrent destination execution is regression-covered, and no credentials were exposed in UI errors, run details, callbacks, or this audit.
