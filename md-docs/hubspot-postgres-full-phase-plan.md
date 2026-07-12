# HubSpot to PostgreSQL Full-Phase Connector Plan

Status: implemented beta; live E2E, canary, and release gates remain  
Scope: HubSpot source only, PostgreSQL destination only  
Target release: stable MVP after every acceptance gate in section 18 passes  
Last repository audit: 2026-07-11

Implementation note (2026-07-11): the beta code now includes the versioned
ten-stream registry and aliases, live permission/property discovery, bounded
dlt incremental resources, isolated DuckDB staging, selected-stream enforcement,
per-stream candidate/commit checkpoints and audited reset, mandatory SQL/dbt,
existing-table Postgres preflight, repository-wide public upsert enforcement,
Phase 3 usage/callback metadata, finalized cleanup status, credential scrubbing,
masked HubSpot previews, and AI token rejection. Estuary remains reference-only
and no Estuary runtime or dependency was added. The catalog deliberately remains
`beta`; live HubSpot E2E, cross-tenant security validation, canary telemetry, and
the complete CI matrix are release operations that require configured external
accounts and are not claimed complete by this implementation.

This plan is subordinate to the repository's strict ELT invariants. Where the
requested feature set and the current product contract differ, the stricter
rule wins: dlt extraction, isolated DuckDB staging, mandatory UI SQL/dbt,
delivery only to existing tables, public upsert-only delivery, per-stream
checkpoint safety, authenticated callbacks, and no credential disclosure.

## 1. Current state summary

### 1.1 What already exists

| Area | Current implementation | Assessment |
| --- | --- | --- |
| Catalog | HubSpot is a source-only SaaS connector in the Go and frontend catalogs. | Present, but metadata is too shallow for a stable capability contract. |
| Credentials | Go encrypts the SaaS `credential` field before storing it in `data_source_connections.config`, decrypts server-side, and masks it in UI responses. | Correct foundation; add HubSpot-specific leak tests and token rotation guidance. |
| Connection test | ELT calls `GET /crm/v3/objects/contacts?limit=1` with a bearer token. | Present, but incorrectly couples token validity to contact scope and returns only a generic result. |
| Discovery | Go returns a static list of 14 HubSpot resources. | Not production discovery: no live permissions, fields, primary keys, sync modes, or limitations. |
| Preview | ELT has direct, capped HubSpot REST previews for 14 resources. | Useful foundation; it bypasses the dlt record-shaping path and can expose raw CRM values. |
| Extraction | `_build_hubspot_source` calls the local dlt HubSpot source with `include_history=False`, `include_custom_props=True`, then applies `with_resources(...)` when selections exist. | Selected resources work when supplied, but an empty selection currently selects the whole source. |
| Streams | The local source includes CRM objects, owners, pipelines/stage timing, properties, and activity resources. | Wider than the desired MVP, with naming mismatches and no single typed registry. |
| Staging | SaaS runs use a per-run DuckDB file and the common `duckdb_staged` path. | Correct architecture; HubSpot resource names still need canonical `schema__table` mapping and lineage columns. |
| Transform | SaaS delivery hard-fails when `dbt_config.sql_models` is absent and runs UI SQL/dbt in DuckDB. | Meets the mandatory transform rule. |
| Delivery | Common delivery verifies destination existence and model columns, then uses dlt to write. | Correct base, but public append/replace options and no-PK append fallback violate strict invariant 9. |
| Callback | The callback already supports the invariant audit fields: `delivery_outputs`, `staging_size_bytes`, `dbt_models_run`, and `no_pk_warnings`. | Extend with phase, stream, safe warning, and committed-checkpoint detail. |
| Checkpoint | The SaaS runner restores and extracts a dlt pipeline-state envelope. Stripe adds an explicit per-stream envelope; HubSpot does not. | HubSpot incremental state and per-stream commit semantics are missing. |
| Cleanup | DuckDB and temporary work directories are removed in `finally`. | Correct base; callback-visible cleanup status is calculated before `finally` and needs contract correction. |
| AI | Pipeline chat receives connector names, selected streams, modes, destinations, and run summary. | No HubSpot-specific context store or vector tables exist; no central PII classification boundary exists. |
| Tests/docs | HubSpot manual guides and matrix fixtures exist. | Several guides claim incremental behavior that the current local source does not implement; they must be corrected before stable release. |

### 1.2 Repository paths to extend

Do not create a parallel runtime. Extend these existing ownership points:

- Frontend catalog, credential, discovery, preview, SQL, approval, and run UI:
  `apps/app/app/workspace/connections/**` and
  `apps/app/app/workspace/data-pipelines/[id]/builder/**`.
- Go catalog, connection proxy, dispatch, plan guard, callback, audit, and
  optional approved-DDL control plane:
  `apps/server/main-server/internal/{connectorsdata,server,models}/**`.
- HubSpot dlt source, discovery, preflight, staging, checkpoint, callback, and
  tests: `apps/server/elt-server/{saas_sources/hubspot,api/routes,runner,models,tests}/**`.
- RLS changes, if AI-context tables are added:
  `apps/server/main-server/sql/supabase_rls.sql` and the RLS guide.

### 1.3 Release-blocking gaps

1. A single typed HubSpot stream registry is missing across Go, Python, and TS.
2. Empty stream selection can result in all dlt resources being selected.
3. Live discovery, per-stream permissions, and selected property discovery are
   missing.
4. The local HubSpot object resources do not apply dlt incremental cursors or
   HubSpot `updatedAt` filters; current “incremental” runs can still scan all
   records.
5. HubSpot checkpoint state is not explicitly per stream and cannot commit only
   successful stream deliveries.
6. `partial_success` is treated as completed by the Go callback, which can
   advance an unsafe aggregate checkpoint.
7. Public append/replace controls and the delivery no-PK append fallback conflict
   with the authoritative upsert-only rule.
8. Callback and usage accounting do not yet expose one canonical Phase 3 row
   counter (`phase3_rows_delivered`).
9. Preview and AI-context PII minimization are not HubSpot-aware.
10. Existing manuals overstate incremental guarantees and use inconsistent
    stream names such as `pipelines`, `pipeline_stages`, and
    `feedback_submissions`.

## 2. Reference comparison: dlt vs Estuary vs MantrixFlow

| Concern | dlt HubSpot source | Estuary HubSpot real-time connector | MantrixFlow MVP decision |
| --- | --- | --- | --- |
| Role | Embeddable Python source/resources. | Managed capture connector with discovered bindings. | Embed and harden dlt; Estuary is a behavior benchmark, not a runtime dependency. |
| Authentication | Static API/private-app token in the verified source. | OAuth2 credentials. | Private app token for MVP; OAuth after MVP. |
| Resource selection | `with_resources(...)` selects named dlt resources. | Catalog bindings select discovered resources. | Validate selections against one registry, then call `with_resources`; empty is an error. |
| Custom properties | Supported; too many properties can exceed HubSpot's request-length limit. | Captures broad schemas and supports calculated-property refresh. | Default on, preflight size, chunk when possible, and fall back per stream with an explicit warning. |
| Property history | Optional separate resources. | Optional capture setting. | Off for MVP; later separate history tables with volume approval. |
| Incremental | dlt supplies state primitives, but the current local HubSpot object resources do not bind an incremental cursor. | Uses `updatedAt`; also schedules full calculated-property refresh because calculated changes do not update that timestamp. | Build HubSpot-specific incremental dlt resources with windowed API reads, per-stream watermarks, overlap, and calculated-property limitations. |
| Stream breadth | Local source and current upstream examples emphasize core CRM objects. | Broader dynamic catalog including forms, lists, custom objects, marketing, and workflows. | Ten stable MVP streams; use Estuary's breadth and calculated-property behavior as Phase 2 input. |
| Staging/delivery | Can load directly to a destination. | Captures into Estuary collections. | Always dlt to isolated DuckDB, then UI SQL/dbt, then existing Postgres tables. |
| Schema changes | Destination behavior depends on dlt configuration. | Materialization contract is platform-managed. | ELT runner never creates a client target. AI may draft DDL; a separate audited approval path may execute it. |

Primary references:

- [dlt HubSpot verified source](https://dlthub.com/docs/dlt-ecosystem/verified-sources/hubspot)
  documents resource selection, history, custom properties, and the 2,000-character
  request issue.
- [dlt source/resource selection](https://dlthub.com/docs/general-usage/source)
  documents `with_resources(...)` behavior.
- [Estuary HubSpot real-time connector](https://docs.estuary.dev/reference/Connectors/capture-connectors/HubSpot-real-time/)
  documents its discovered resources and calculated-property refresh model.
- [HubSpot API usage limits](https://developers.hubspot.com/docs/developer-tooling/platform/usage-guidelines)
  is authoritative for 429 behavior and account/app limits.
- [HubSpot object API guide](https://developers.hubspot.com/docs/api-reference/latest/crm/using-object-apis)
  is authoritative for object IDs, properties, paging, and permission failures.

The key conclusion is architectural: replacing dlt with Estuary would violate the
MVP extraction contract and introduce a second orchestration plane. The useful
Estuary lessons are dynamic permission-aware discovery, a broader catalog,
calculated-property refresh, and explicit OAuth—not its runtime.

## 3. Phase -2: Connector registration

### 3.1 Canonical connector metadata

The server-owned connector catalog must expose this stable contract; frontend
catalog entries should be derived from or contract-tested against it:

```json
{
  "id": "hubspot",
  "label": "HubSpot",
  "type": "hubspot",
  "category": "SaaS",
  "direction": "source",
  "auth_modes": ["private_app_token"],
  "mvp_status": "stable",
  "supports_discovery": true,
  "supports_preview": true,
  "supports_incremental": true,
  "supports_custom_properties": true,
  "supports_property_history": false
}
```

`mvp_status` must remain `beta` or `in_progress` in production until section 18
passes. It becomes `stable` only as the release action, not merely when metadata
is merged.

### 3.2 Registry design

Create one versioned HubSpot stream registry in Python and export/contract-test
equivalent JSON through Go. Each stream record contains:

- canonical stream key and user label;
- dlt resource name and canonical DuckDB name;
- HubSpot object type/endpoint family;
- primary key and cursor field;
- default selected fields and property schema endpoint;
- required and alternative read scopes;
- supported and recommended sync modes;
- custom-property, history, archive, preview, and availability flags;
- account-tier or product limitations;
- release tier: `mvp`, `phase_2`, or `unsupported`.

Do not keep independent resource arrays in `validate.go`, `preview.py`,
`connector_support.py`, docs, and frontend constants. Generate them or enforce a
snapshot contract test so drift fails CI.

### 3.3 MVP and Phase 2 catalog

The exact lists are in section 11. The MVP has ten streams. The user-facing
pipeline stream names are `deal_pipelines` and `ticket_pipelines`; the registry
maps them to the current local dlt names `pipelines_deals` and
`pipelines_tickets`. This alias is applied only at the source boundary. DuckDB
and callbacks always use the canonical public stream name.

### 3.4 Known registration limitations

- Private app tokens are tied to one HubSpot account and manually rotated.
- Availability depends on account products, enabled objects, and granted scopes.
- Sensitive and highly sensitive properties are excluded from MVP discovery and
  extraction even if a token has a broader scope.
- Property history, archived/deleted records, associations, and calculated-value
  refresh are not MVP guarantees.
- Incremental support is stream-specific, not a connector-wide promise.

## 4. Phase -1: Connection setup and discovery

### 4.1 Connection setup flow

1. User chooses HubSpot as a source and sees “Private app access token.”
2. UI links to the HubSpot private/static-token app instructions and lists only
   the scopes required for the selected MVP streams.
3. The token is submitted over TLS to the authenticated, org-scoped Go API.
4. Go validates input shape, encrypts `credential`, writes the secret-bearing
   connection row, and never returns the token after save.
5. Go calls the internal ELT connection test using the decrypted credential only
   in memory and an authenticated internal request.
6. On success, UI receives safe connection metadata; on failure, it receives a
   classified, scrubbed message.
7. Discovery runs live probes and caches only safe schema/permission metadata.

Connection create/update must be atomic from the user's perspective: if the test
is configured as “test before save,” do not persist a failed token; if save-then-
test is retained, mark the connection `error` and keep the token encrypted.

### 4.2 HubSpot private app instructions

For the MVP UI and docs:

1. In HubSpot, open Settings, then Integrations, then Private Apps (or the current
   static-token app flow).
2. Create an app named for MantrixFlow and select read-only scopes for the streams
   being connected.
3. Never grant write scopes merely to make discovery pass.
4. Copy the access token once and paste it into MantrixFlow.
5. Reinstall/re-authorize the app after adding scopes, and rotate the token at
   least on the organization's security schedule; HubSpot currently recommends
   periodic rotation.

Required scopes must be rendered from the registry. Typical mappings include
`crm.objects.contacts.read`, `crm.objects.companies.read`,
`crm.objects.deals.read`, ticket access, `crm.objects.owners.read`,
`crm.objects.products.read`, `crm.objects.line_items.read`,
`crm.objects.quotes.read`, and the relevant pipeline/schema read scopes. Confirm
the exact current names against HubSpot's
[scope reference](https://developers.hubspot.com/docs/apps/developer-platform/build-apps/authentication/scopes)
when implementation begins; do not hard-code remembered scope names without a
registry test.

### 4.3 Safe connection test

Do not use contacts as the universal validity test. The test sequence is:

1. Validate token presence and reject obvious placeholders/masked values.
2. Call a lightweight, registry-backed read endpoint that is covered by the MVP
   base scope, such as owners when that scope is required.
3. Probe selected stream endpoints with `limit=1`, without requesting contact,
   company, or deal properties.
4. Classify 401 as invalid token, 403 as valid token/missing stream permission,
   429 as retryable rate limit, and other 5xx/network failures as retryable.
5. Never put the token in a URL path, query string, exception, structured log,
   trace attribute, or response.

Safe response:

```json
{
  "status": "connected",
  "connector": "hubspot",
  "auth_type": "private_app_token",
  "portal_id": "safe_account_identifier_if_available",
  "available_stream_count": 10,
  "token_visible_to_ai": false
}
```

Omit `portal_id` when it cannot be obtained without a broader permission or a
token-in-path introspection call. Do not infer it from CRM records. Never return
headers, tokens, or raw records.

### 4.4 Discovery response

Discovery is live, bounded, and tolerant of optional-stream failures:

```json
{
  "connector": "hubspot",
  "catalog_version": 1,
  "streams": [
    {
      "stream_key": "hubspot.contacts",
      "name": "contacts",
      "label": "Contacts",
      "primary_key": ["id"],
      "default_fields": ["id", "email", "firstname", "lastname", "createdAt", "updatedAt"],
      "custom_properties_supported": true,
      "supported_sync_modes": ["FULL_TABLE", "INCREMENTAL"],
      "recommended_sync_mode": "INCREMENTAL",
      "permission_status": "available",
      "limitations": ["Calculated-only changes require a later refresh strategy"]
    }
  ],
  "warnings": [],
  "partial": false
}
```

Discovery performs metadata calls only: object/schema and property endpoints,
pipeline definitions, and one-record permission probes. It does not load full
datasets. Property discovery is paginated, cached by connection/object with a
short TTL, and capped. Sensitive classifications, hidden fields, read-only flags,
and property types are retained in the server cache but filtered before AI use.

### 4.5 Discovery errors

- A 401 fails the whole discovery as `AUTH_ERROR`.
- A 403 marks only that stream `permission_status=missing_scope`; required MVP
  base permission failure can also set the connection to `limited`.
- A 404 or object-not-enabled response marks that stream `unavailable`.
- A 429 stops further probes, returns already discovered streams with
  `partial=true`, and includes safe retry timing.
- A timeout/5xx retries with jitter within a small request budget, then returns a
  partial catalog.
- A property endpoint failure keeps the stream but marks fields
  `metadata_unavailable`; it never silently claims custom properties are safe.

## 5. Phase 0: Preflight validation

Phase 0 runs before data extraction and before the run-specific DuckDB file is
created. Checks that require actual dbt output are repeated at the Phase 2/3
boundary. Every error has a stable code, retryability, safe user message, and
operator detail that contains no credential or raw CRM payload.

### 5.1 HubSpot checklist

- [ ] Encrypted connection exists; decrypted token is non-empty in server memory.
- [ ] Lightweight authentication call succeeds.
- [ ] Every selection resolves to an MVP registry entry; selection is non-empty.
- [ ] Every selected stream permission probe succeeds.
- [ ] Selected properties exist, are readable, and are not sensitive-denied.
- [ ] Encoded property requests fit the safe request budget or have a valid
      chunk plan.
- [ ] `include_history=false` for MVP.
- [ ] Each incremental stream has a supported cursor and a valid RFC3339 start.
- [ ] Lookback is within configured minimum/maximum bounds.
- [ ] Backfill lower bound is before the run's captured upper bound.
- [ ] Rate-limit retry policy, maximum attempts, request budget, and deadline are
      configured.
- [ ] Per-stream prior checkpoint shape and registry version are valid.

### 5.2 PostgreSQL checklist

- [ ] Destination connection succeeds with TLS policy enforced.
- [ ] Destination schema exists.
- [ ] Every `dbt_config.sql_models[].dest_table` is a valid `schema.table`.
- [ ] Every destination table already exists.
- [ ] Saved SQL validation output columns are a subset of destination columns;
      actual dbt output is rechecked before delivery.
- [ ] Every destination has a primary key or approved stable merge key; otherwise
      fail—never fall back to append.
- [ ] The database user can `SELECT` metadata and `INSERT`/`UPDATE` the target.
- [ ] Output-to-destination types are safely coercible; narrowing or lossy casts
      require an explicit SQL model cast.
- [ ] No output column begins with `_dlt_` and no client target is in a dlt
      internal schema.

Use a transaction/savepoint or catalog privilege inspection for write permission
checks; do not insert a real business row during preflight.

### 5.3 MantrixFlow checklist

- [ ] Exactly one source node and at least one Postgres destination node exist.
- [ ] Source is HubSpot and the selected target is Postgres.
- [ ] `selected_streams` is `SourceStreamConfig[]`, never a raw string list.
- [ ] Every entry has `stream_key`, replication method/key, and canonical
      `duckdb_table_name`.
- [ ] UI SQL/dbt models exist, parse, and reference selected sources only.
- [ ] Public emit method resolves to upsert/merge only.
- [ ] Organization row usage has plan headroom for the estimated delivery.
- [ ] Dispatcher and ELT Phase 0 disk checks pass.
- [ ] Callback URL and `X-Callback-Token` are configured.
- [ ] No active/queued run violates the pipeline concurrency/idempotency rule.
- [ ] Destination/model mapping is one-to-one or has explicit source dependency
      metadata for checkpoint commit decisions.

### 5.4 Retryability

Retry only 429, network timeout/reset, HubSpot 5xx, ELT capacity waiting, and
transient Postgres connection errors. Use exponential backoff with full jitter,
honor `Retry-After`, cap attempts and wall time, and preserve the same run id for
idempotent retries. Do not retry invalid tokens, missing scopes, unsupported
streams, invalid fields, unsafe SQL, missing destination tables/columns/PKs,
plan exhaustion, or malformed checkpoint state.

Representative user messages are in section 15.

## 6. Phase 1: HubSpot extraction and DuckDB staging

### 6.1 Source extraction plan

1. Convert validated `SourceStreamConfig[]` to canonical public resource names.
2. Resolve public names to local dlt resource names through the registry.
3. Build the HubSpot dlt source with the encrypted token decrypted only in
   memory, `include_history=False`, explicit selected properties, and the
   configured custom-property policy.
4. Require at least one resource, then call `with_resources(*resolved_names)`.
5. Wrap/bind each CRM resource with its stream-specific incremental reader when
   incremental is selected. All network records still enter through a dlt
   resource; no direct-to-DuckDB bypass is allowed.
6. Apply stable record shaping and lineage before dlt normalization.
7. Run dlt into the per-run isolated DuckDB dataset `raw`.
8. Record per-stream read/stage results and candidate checkpoints without
   committing them.

### 6.2 Canonical staging names

Strict invariant 8 overrides the prompt's `_raw` examples. Public stream keys
are `schema.table`; internal tables are `schema__table`:

| Stream key | DuckDB relation |
| --- | --- |
| `hubspot.contacts` | `raw.hubspot__contacts` |
| `hubspot.companies` | `raw.hubspot__companies` |
| `hubspot.deals` | `raw.hubspot__deals` |
| `hubspot.tickets` | `raw.hubspot__tickets` |
| `hubspot.owners` | `raw.hubspot__owners` |
| `hubspot.deal_pipelines` | `raw.hubspot__deal_pipelines` |
| `hubspot.ticket_pipelines` | `raw.hubspot__ticket_pipelines` |
| `hubspot.products` | `raw.hubspot__products` |
| `hubspot.line_items` | `raw.hubspot__line_items` |
| `hubspot.quotes` | `raw.hubspot__quotes` |

Only `duckdbTableName(...)` in Go and `duckdbTableNameForStream(...)` in TS may
perform this conversion. Python validates the supplied result against the same
algorithm; it does not invent a second naming scheme.

### 6.3 Raw record contract

Every staged object has, where available:

- `id` as text, preserving the original HubSpot object ID;
- source `createdAt` and `updatedAt` plus relevant property timestamps;
- `archived` as boolean;
- selected standard columns in stable snake_case;
- `properties` as JSON text/JSON for long-tail custom values rather than
  flattening every custom property into the stable contract;
- `extracted_at` as UTC timestamp;
- `_mantrixflow_run_id` as the run UUID/text;
- optional non-sensitive association IDs only when explicitly selected later.

Never stage tokens, request/response headers, raw error bodies, token metadata,
or internal dlt state as business columns. dlt's own metadata remains inside the
isolated DuckDB and is filtered before model delivery.

### 6.4 Incremental algorithm

For CRM object streams, capture an immutable run upper bound `T_end`, query
records in bounded `[T_start-lookback, T_end)` windows ordered by update time,
and page deterministically. If a HubSpot search window reaches an API result cap,
split the time window until it is safely pageable; if equal timestamps still
exceed a cap, use ID tie-breaking/batch reads or fail with a resumable error
rather than dropping records.

Deduplicate staged candidates by `(stream, id)`, keeping the greatest
`updatedAt`, then highest deterministic page/order key. Upsert to Postgres later
uses the destination PK, not the cursor.

Small reference streams such as deal/ticket pipelines may use snapshot refresh
into DuckDB each run while still producing a per-stream completion marker.
Owners can use `updatedAt` only after its paging and archived behavior are proven;
otherwise it remains full snapshot in MVP. The registry advertises only what is
actually implemented.

### 6.5 Candidate checkpoint envelope

```json
{
  "connector": "hubspot",
  "version": 1,
  "catalog_version": 1,
  "streams": {
    "contacts": {
      "cursor_field": "updatedAt",
      "committed_value": "2026-07-01T00:00:00Z",
      "candidate_value": "2026-07-11T10:30:00Z",
      "lookback_seconds": 3600,
      "last_success_at": "2026-07-11T10:30:00Z",
      "status": "candidate"
    }
  },
  "pipeline_state": {}
}
```

The callback sends both the prior committed state and candidate values. Go
merges only streams listed in `checkpoint_commit_streams`. A failed stream,
transform, or dependent delivery retains its prior committed value. Reset is an
org-scoped audited operation on one stream; reset-all requires a separate
confirmation.

### 6.6 Custom-property fallback

1. Preflight computes encoded property length and chunks compatible reads where
   the API/source can merge property subsets by ID.
2. If HubSpot still rejects a stream as too large, retry that stream once with
   custom properties disabled.
3. Keep required standard fields and IDs; do not restart successful unrelated
   streams.
4. Add the stream to `custom_properties_disabled_streams`, emit
   `hubspot_custom_properties_disabled`, and present a visible warning.
5. Set top-level `custom_properties_disabled=true` when any stream falls back.
6. Do not advance a stream checkpoint unless the fallback version transformed
   and delivered successfully.

This fallback prevents total run loss but is not silent success. The next run
must continue warning until the user selects fewer properties or disables custom
properties intentionally.

### 6.7 Property history

MVP rejects `include_history=true`. Phase 2 enables it per supported stream,
uses separate `raw.hubspot__<stream>_property_history` tables, estimates rows and
storage before execution, requires explicit high-volume approval, and maintains
a separate history checkpoint. History is never embedded into the base object
row or automatically sent to AI context.

## 7. Phase 2: SQL/dbt transformation

### 7.1 Model plan

Each selected source has an explicit UI SQL model. `source_stream_key`,
`duckdb_source_table`, `output_table`, and `dest_table` are required. Models use
`{{ source('raw', 'hubspot__contacts') }}`-style references and write into the
configured DuckDB analytics schema.

| Model | Source(s) | Required output key | Recommended destination |
| --- | --- | --- | --- |
| `dim_hubspot_contacts` | contacts | `contact_id` | `public.hubspot_contacts` |
| `dim_hubspot_companies` | companies | `company_id` | `public.hubspot_companies` |
| `fct_hubspot_deals` | deals | `deal_id` | `public.hubspot_deals` |
| `fct_hubspot_tickets` | tickets | `ticket_id` | `public.hubspot_tickets` |
| `dim_hubspot_owners` | owners | `owner_id` | `public.hubspot_owners` |
| `dim_hubspot_deal_pipelines` | deal pipelines and nested stages | `(pipeline_id, stage_id)` | `public.hubspot_deal_pipelines` |
| `dim_hubspot_ticket_pipelines` | ticket pipelines and nested stages | `(pipeline_id, stage_id)` | `public.hubspot_ticket_pipelines` |
| `dim_hubspot_pipelines` | optional union of both pipeline-stage models | `(object_type, pipeline_id, stage_id)` | Optional combined existing table |
| `fct_hubspot_line_items` | line items | `line_item_id` | `public.hubspot_line_items` |
| `dim_hubspot_products` | products | `product_id` | `public.hubspot_products` |
| `fct_hubspot_quotes` | quotes | `quote_id` | `public.hubspot_quotes` |

Pipeline resources are shaped as one row per stage so both the pipeline and
stage identifiers survive normalization. The requested generic
`dim_hubspot_pipelines` pattern must not collapse deal and ticket pipelines onto
`pipeline_id` alone; either keep two tables keyed by `(pipeline_id, stage_id)`
or use `(object_type, pipeline_id, stage_id)` in a combined table. If a pipeline
has no stages, emit a pipeline row with a documented sentinel/nullable-stage
contract only when the destination key supports it; do not invent a stage ID.

### 7.2 Recommended columns

| Entity | Recommended output |
| --- | --- |
| Contacts | `contact_id`, `email`, `first_name`, `last_name`, `phone`, `company`, `lifecycle_stage`, `hubspot_owner_id`, `created_at`, `updated_at`, `archived`, `properties`, `_mantrixflow_run_id`, `extracted_at` |
| Companies | `company_id`, `name`, `domain`, `industry`, `city`, `country`, `hubspot_owner_id`, `created_at`, `updated_at`, `archived`, `properties`, lineage |
| Deals | `deal_id`, `deal_name`, `amount`, `deal_stage`, `pipeline_id`, `close_date`, `hubspot_owner_id`, `created_at`, `updated_at`, `archived`, `properties`, lineage |
| Tickets | `ticket_id`, `subject`, `content`, `pipeline_id`, `pipeline_stage_id`, `hubspot_owner_id`, `created_at`, `updated_at`, `archived`, `properties`, lineage |
| Owners | `owner_id`, `user_id`, `first_name`, `last_name`, `email`, `archived`, `created_at`, `updated_at`, lineage |
| Pipelines | `object_type`, `pipeline_id`, `pipeline_label`, `stage_id`, `stage_label`, `stage_display_order`, `is_closed`, `probability`, `is_default`, `created_at`, `updated_at`, lineage |
| Products | `product_id`, `name`, `sku`, `price`, `description`, `created_at`, `updated_at`, `archived`, `properties`, lineage |
| Line items | `line_item_id`, `name`, `quantity`, `unit_price`, `amount`, `product_id`, `created_at`, `updated_at`, `properties`, lineage |
| Quotes | `quote_id`, `title`, `status`, `expiration_at`, `amount`, `created_at`, `updated_at`, `archived`, `properties`, lineage |

Money uses `DECIMAL/NUMERIC` after `NULLIF` and safe parsing, never binary float.
Timestamps normalize to UTC-compatible timestamp values. IDs remain text to
avoid accidental numeric truncation. Column names are snake_case; original IDs,
owner IDs, pipeline IDs, and stage IDs are never lossy-renamed.

### 7.3 SQL validation and safety

- Parse exactly one read-only `SELECT`/CTE statement.
- Reject DDL, DML, `COPY`, `ATTACH`, `INSTALL`, external scans, file/network
  functions, secrets, pragmas that change runtime state, and multi-statements.
- Resolve every source reference to a selected canonical DuckDB table.
- Reject `SELECT *` for saved production models unless expanded and frozen by
  the editor.
- Infer/describe output columns and types before save.
- Require the output key and compare it to the existing Postgres PK/merge key.
- Reject output `_dlt_*` columns and duplicate/case-colliding names.
- Flag unsupported Postgres-only functions because execution is dbt-duckdb.
- Require explicit casts for unsafe type conversions.
- Store a SQL hash and discovered output contract; revalidate when source schema,
  SQL, destination table, or registry version changes.

### 7.4 Preview

Preview must run the same record-shaping and SQL path as production against a
small, bounded dlt sample in disposable DuckDB. Return columns, types, masked
rows, validation findings, and truncation flags. Mask email local parts, phone
numbers, note/content/body fields, and properties classified as personal or
sensitive. Raw preview can be shown only in the authenticated operator UI under
an explicit product policy; it is never forwarded to Copilot by default.

## 8. Phase 3: Postgres delivery

### 8.1 Delivery contract

- Every `dest_table` is a pre-existing `schema.table`.
- Recheck target existence, columns, types, and PK after dbt and immediately
  before writing.
- Public delivery is upsert-only. The builder, API, GitHub YAML, and Slack flow
  must reject append, replace, and overwrite for this MVP.
- No-PK is a hard preflight failure, not an append warning/fallback.
- Deliver only final dbt model rows; never deliver raw or `_dlt_*` relations.
- Count usage only from successfully committed Phase 3 rows.
- A model delivery is idempotent by destination key and run retry.

This intentionally resolves a prompt-level ambiguity: append/overwrite were
requested as possible explicit modes, but strict invariant 9 requires public
upsert-only delivery. They remain unsupported until that invariant is formally
changed repository-wide.

### 8.2 Destination plan

Recommended existing targets are:

`hubspot_contacts`, `hubspot_companies`, `hubspot_deals`, `hubspot_tickets`,
`hubspot_owners`, `hubspot_deal_pipelines`, `hubspot_ticket_pipelines`,
`hubspot_products`, `hubspot_line_items`, and `hubspot_quotes` in the user's
chosen schema.

Each has the entity key in section 7 plus lineage columns. Tables containing
PII should use destination-appropriate grants and RLS if the destination is a
Supabase-exposed schema; connection ownership does not make destination rows
safe automatically.

For example, an AI proposal for `public.hubspot_contacts` should contain this
reviewable contract before it renders SQL: `contact_id text primary key`, the
standard contact fields from section 7, `created_at/updated_at/extracted_at`
as `timestamptz`, `archived boolean`, `properties jsonb`, and
`_mantrixflow_run_id text`. The proposal must explain nullable fields, show that
`contact_id` is the upsert key, and verify the dbt output contract against the
proposed columns.

### 8.3 Missing table and AI DDL approval

When a target is missing, the run stops with `DESTINATION_TABLE_MISSING`. The AI
planner may generate a versioned proposal containing:

- fully qualified table;
- quoted, typed columns and PK;
- nullable/default choices;
- optional indexes;
- expected model-to-column mapping;
- risk summary, SQL preview, and proposal hash.

UI actions are `Approve Create Table`, `Edit SQL`, and `Cancel`. Approval does
not let the ELT runner execute DDL. If automatic approved execution is added, it
must be a separate org-scoped Go schema-management endpoint with editor/owner
authorization, destination re-authentication, exact proposal hash, transaction,
audit log, idempotency key, and post-create introspection. Any ALTER proposal
uses the same flow. Copilot may draft but cannot self-approve.

### 8.4 Column mismatch and delivery failure

Column errors name every offending output column and the target. Missing target
columns can produce an AI ALTER proposal but never an automatic alteration.
Per-model delivery results contain target, status, rows attempted/delivered,
duration, safe error code, and retryability. Checkpoint commit is withheld for
the source stream(s) feeding a failed model. Independent successful stream/model
pairs may commit only when the callback and Go merge logic support partial
per-stream commit safely.

## 9. Phase 4: Callback, checkpoint, observability and cleanup

### 9.1 Callback contract

Retain current field names for compatibility and add aliases only at the Go
boundary. The canonical payload includes:

- correlation: `run_id`, `job_id`, `pipeline_id`, `organization_id`;
- connector identity: `source_connector=hubspot`,
  `destination_connector=postgres`, `source_tool`, `dest_tool`;
- status: overall `status`, `phase_status`, per-stream results;
- counts: `records_read`, `records_staged`, `records_transformed`,
  `records_delivered`, plus current `rows_read`, `rows_written`, and canonical
  `phase3_rows_delivered`;
- state: `checkpoint_state`/`checkpoint`, `checkpoint_commit_streams`,
  `checkpoint_blocked_streams`;
- config outcome: selected streams, custom-property fallback streams,
  `custom_properties_disabled`, `property_history_enabled`;
- dbt: `dbt_status`/`dbt_run_status`, `dbt_models_run`, tests and warnings;
- delivery: `delivery_outputs`, `delivery_failures`, `no_pk_warnings` (which
  should be empty because no-PK is a hard failure);
- runtime: `duration_ms`, `duration_seconds`, `staging_size_bytes`, cleanup
  status;
- safe `warnings` and classified `errors`.

Callbacks never contain tokens, authorization headers, raw HubSpot error bodies,
raw records, email/phone values, private notes, or destination passwords. Go
validates `X-Callback-Token`, run/pipeline/org correlation, known fields, bounded
array/string sizes, and legal status transitions before persistence.

### 9.2 Checkpoint commit rules

1. Extract dlt state and candidate per-stream state before DuckDB deletion.
2. A stream is committable only if extraction, staging, every dependent dbt
   model, and every dependent delivery succeeded.
3. A failed or skipped dependent step preserves the old committed watermark.
4. Go merges committable stream entries into the prior checkpoint atomically.
5. A callback retry is idempotent by run ID and checkpoint proposal version.
6. `partial_success` remains a distinct run outcome; it is not blindly converted
   to full success.
7. Manual per-stream reset writes an audit event and does not mutate other
   streams.

### 9.3 Events

Emit the requested events with `run_id`, `pipeline_id`, `organization_id`, stream
name when applicable, duration/counts, and safe status—never CRM values:

- `hubspot_preflight_started`, `hubspot_preflight_completed`;
- `hubspot_stream_started`, `hubspot_page_received`,
  `hubspot_stream_completed`, `hubspot_stream_failed`;
- `hubspot_custom_properties_disabled`;
- `hubspot_transform_started`, `hubspot_transform_completed`;
- `hubspot_delivery_started`, `hubspot_delivery_completed`;
- `hubspot_checkpoint_saved`;
- `hubspot_pipeline_completed`, `hubspot_pipeline_failed`.

For `hubspot_page_received`, record only page number/cursor hash and row count,
not the paging URL if it can carry sensitive query details.

### 9.4 Cleanup order

In the runner's `finally` path:

1. Extract state/candidate checkpoint if a pipeline exists and extraction ran.
2. Build the final safe callback payload state in memory.
3. Close DuckDB/dlt handles.
4. Delete the DuckDB file.
5. Remove temporary dbt project, staging folders, and extracted temp files.
6. Record final cleanup status and send/retry the authenticated callback through
   the existing callback mechanism.

If the current architecture sends the callback after `run()` returns, return a
payload whose cleanup status was finalized, not the pre-`finally` default.
Durable masked debug artifacts may be uploaded to Tigris before local deletion
only under the retention policy. Never upload the active DuckDB file or raw CRM
payload merely for convenience.

## 10. Phase 5: AI context, testing, docs and release readiness

### 10.1 Copilot capability plan

Copilot may assist with stream selection, field mapping, destination DDL
proposals, SQL generation, safe failure investigation, and optimization. It may
also prepare chat-with-data query plans and Slack incident summaries. Every
action is proposal-first; connection, DDL, run, reset, or destructive changes
use the existing product authorization and approval path.

Allowed context:

- connector/stream names and release capabilities;
- field names, safe types, nullability, classifications, and schema hashes;
- destination schema summaries without credentials;
- masked or synthetic sample values;
- safe run/phase/stream status and classified errors;
- mapping and SQL model summaries;
- aggregate counts, timings, and rate-limit metrics.

Denied context:

- HubSpot tokens or authorization headers;
- raw CRM payloads or full contact/company lists;
- unmasked emails, phones, names paired with contact details, private notes, or
  sensitive/highly-sensitive properties;
- destination credentials/DSNs;
- unfiltered log bundles or DuckDB files.

Apply allowlist construction at the server boundary; do not rely on a prompt
instruction to remove secrets after context assembly. Also scrub user-provided
chat text for token patterns before logging/vectorization, while warning the user
not to paste secrets.

### 10.2 Safe context persistence

If created, `agent_context_documents`, `agent_context_chunks`, and
`agent_context_embeddings` are org-owned tables with source/pipeline/run IDs,
classification, retention deadline, content hash, and embedding model/version.
Store sanitized text only; embeddings are treated as derived personal data, not
as anonymized data.

Supabase requirements:

- RLS on all three tables, org membership predicates, and real cross-tenant JWT
  tests;
- no direct client write unless explicitly required;
- explicit grants when the Data API setting requires them—Supabase announced in
  2026 that tables may no longer be automatically exposed to Data/GraphQL APIs;
- no `SECURITY DEFINER` shortcut for authorization;
- deletion cascades and retention jobs that delete chunks and embeddings.

Relevant current notice:
[Supabase breaking-change changelog](https://supabase.com/changelog?tags=breaking-change).

Tigris object layout uses opaque org/run IDs and stores only schema snapshots,
masked sample files, sanitized run bundles, SQL validation reports, and context
snapshots. Objects are encrypted, private, retention-tagged, and deleted on
organization/data-source erasure. Tigris remains artifact/backup storage, not
the active DuckDB filesystem.

### 10.3 Implementation work packages

1. **Registry and contracts:** typed catalog, aliases, metadata API, drift tests.
2. **Safe auth/discovery:** connection response, permission probes, property
   metadata, partial errors, masked preview.
3. **Incremental dlt resources:** windowed reads, overlap, backfill, dedupe,
   per-stream candidates.
4. **Strict pipeline integration:** canonical names, mandatory dbt, upsert-only,
   existing-table/PK/type preflight, Phase 3 counts.
5. **Checkpoint/callback:** partial stream semantics, atomic Go merge, cleanup
   status, metadata persistence and Realtime.
6. **AI and approved DDL:** allowlisted context, optional vector schema/RLS,
   proposal and approval boundary.
7. **Verification and release:** automated matrix, live sandbox, docs correction,
   canary, rollback.

## 11. MVP stream catalog

### 11.1 Stable MVP streams

| # | Canonical stream | dlt resource mapping | PK | MVP sync | Notes |
| --- | --- | --- | --- | --- | --- |
| 1 | `contacts` | `contacts` | `id` | Incremental | Custom properties supported; PII masking required. |
| 2 | `companies` | `companies` | `id` | Incremental | Preserve domain/industry and properties JSON. |
| 3 | `deals` | `deals` | `id` | Incremental | Decimal-safe amount; preserve pipeline/stage/owner. |
| 4 | `tickets` | `tickets` | `id` | Incremental | Content is private-by-default for AI. |
| 5 | `owners` | `owners` | `id` | Full snapshot initially | Incremental only after `updatedAt` behavior is proven. |
| 6 | `deal_pipelines` | `pipelines_deals` | `id` | Full snapshot | Canonical alias hides local legacy name. |
| 7 | `ticket_pipelines` | `pipelines_tickets` | `id` | Full snapshot | Canonical alias hides local legacy name. |
| 8 | `products` | `products` | `id` | Incremental | Product availability can depend on account setup. |
| 9 | `line_items` | `line_items` | `id` | Incremental | Preserve product/deal association only when selected. |
| 10 | `quotes` | `quotes` | `id` | Incremental | Account/product limitations exposed by discovery. |

### 11.2 Phase 2 streams

`forms`, `form_submissions`, `contact_lists`, `contact_list_memberships`,
`email_events`, `campaigns`, `engagements`, `calls`, `emails`, `meetings`,
`notes`, `tasks`, `properties`, and `custom_objects`.

Several exist partially in the local source today. They remain unsupported in a
stable catalog until their permissions, schemas, incremental strategy, PII
classification, checkpoints, tests, and docs meet the same contract as MVP.

### 11.3 Unsupported for MVP

- property-history resources and stage-timing history;
- archived/deleted-object capture and hard-delete propagation;
- feedback submissions, goals, workflows, marketing emails/events, orders,
  associations as standalone streams, and web analytics events;
- sensitive/highly-sensitive property extraction;
- custom object schemas and records;
- write-back to HubSpot;
- webhook/real-time capture;
- OAuth and multi-account token management.

“Unsupported” means hidden or clearly disabled, never listed as stable merely
because a local endpoint/resource exists.

## 12. Authentication and security plan

### 12.1 Secret lifecycle

- Accept the token only on authenticated org-scoped create/update/test routes.
- Encrypt at rest using the existing Go connection encryption service.
- Keep `data_source_connections` closed to client base-table access.
- Decrypt only for an in-scope internal ELT request and only in memory.
- Require `X-ETL-Token` on ELT routes and `X-Callback-Token` on callbacks.
- Mask saved values as `***`; never return a token after save.
- Rotate/revoke through connection update; invalidate cached discovery after
  rotation.
- Never persist a token in pipeline graph JSON, run metadata, pgmq diagnostic
  payloads exposed to users, Tigris, AI context, or vector tables.

### 12.2 Defense in depth

- Central secret scrubber handles exact credential replacement plus HubSpot token
  patterns before logs/errors/traces/callbacks.
- Do not log request headers or full outbound URLs.
- Bound and sanitize HubSpot error bodies; retain safe correlation IDs separately
  if useful.
- Apply least-privilege, read-only scopes and exclude sensitive scopes by policy.
- SSRF is prevented by fixed HubSpot API base URLs; users cannot supply an API
  host.
- All SQL identifiers are parsed/quoted and all data writes parameterized through
  the delivery layer.
- JWT/org ownership is checked on connection, discovery, preview, run, reset,
  AI-context, artifact, and DDL approval operations.
- RLS is verified with two real tenant JWTs for control-plane tables.

## 13. Incremental sync strategy

### 13.1 Stream classes

- **CRM updated objects:** contacts, companies, deals, tickets, products, line
  items, and quotes use `updatedAt`/verified object-specific update property,
  stable upper bounds, window splitting, overlap, and ID dedupe.
- **Reference snapshots:** deal pipelines and ticket pipelines are small full
  snapshots staged each run and upserted to the existing target.
- **Owners:** full snapshot for MVP unless tests prove a complete incremental
  contract including archived owners.

### 13.2 Backfill and scheduling

- First run requires a user backfill start date or a documented safe default.
- Capture `T_end` before extraction so new writes land in the next overlapped run.
- Default lookback is 3,600 seconds; make it bounded and per stream.
- Store UTC RFC3339 values and registry/cursor versions.
- Full refresh ignores committed watermarks but does not erase them until the
  full run delivers successfully; then it replaces that stream's state.
- Manual reset is per stream and offers a dry-run estimate.

### 13.3 Calculated and late-changing properties

HubSpot calculated-property changes may not advance object `updatedAt`, a
limitation also documented by Estuary. MVP documents this explicitly. Phase 2
adds an opt-in scheduled property refresh that emits partial property snapshots
and merges them through the dbt model, with a full-scan cost warning and separate
checkpoint/metadata.

## 14. Custom property and property history strategy

### 14.1 Custom properties

- Default `include_custom_props=true`, subject to account/property limits.
- Discovery classifies standard, custom, calculated, sensitive, hidden, and
  read-only properties.
- UI selects a bounded subset; “all custom properties” shows a request/volume
  estimate.
- Store long-tail values in `properties` JSON and promote only user-selected
  fields to typed model columns.
- Cache property definitions, not CRM values.
- Chunk property reads by encoded request size and merge by object ID.
- Use the explicit fallback in section 6.6 and persist warnings.

HubSpot exposes property-limit information, but presence of capacity does not
mean a single read request can safely include every property. See the
[HubSpot property limits guide](https://developers.hubspot.com/docs/api-reference/latest/crm/limits-tracking/guide).

### 14.2 History

- Off and rejected in MVP.
- Phase 2 enables per stream with separate tables, keys, checkpoints, row counts,
  retention, and AI exclusion.
- Preflight estimates amplification and requires approval over plan thresholds.
- A history failure never blocks unrelated base objects unless the user marked
  history required.

## 15. Error handling matrix

| Code | Example safe message | Retry | Checkpoint effect |
| --- | --- | --- | --- |
| `AUTH_ERROR` | HubSpot token is invalid. Reconnect HubSpot with a valid private app access token. | No | Commit none. |
| `PERMISSION_ERROR` | The token cannot read deals. Add the required CRM read scope and reconnect. | No | Block deals only; required preflight may stop run. |
| `STREAM_UNSUPPORTED` | `forms` is not supported in the HubSpot MVP. | No | No candidate created. |
| `STREAM_NOT_SELECTED` | Select at least one HubSpot stream before running. | No | No change. |
| `PROPERTY_INVALID` | Property `x` is not available for contacts. Refresh discovery or remove it. | No | Block contacts. |
| `REQUEST_TOO_LARGE` | Too many HubSpot custom properties were requested. Select fewer fields. | No after one safe fallback | Commit only if fallback delivered; warn. |
| `RATE_LIMITED` | HubSpot rate limit reached. This run will retry after the allowed delay. | Yes | Keep old state until success. |
| `HUBSPOT_TEMPORARY_ERROR` | HubSpot is temporarily unavailable. The run will retry. | Yes | Keep old state. |
| `BACKFILL_RANGE_INVALID` | The HubSpot backfill start must be earlier than the end time. | No | No change. |
| `CHECKPOINT_INVALID` | Saved contacts sync state is incompatible. Reset that stream or run a full refresh. | No | Preserve old state. |
| `PLAN_LIMIT_EXCEEDED` | This run exceeds available row usage. Reduce the range or upgrade the plan. | No | No change. |
| `ELT_CAPACITY_WAITING` | Staging capacity is busy. The run is waiting and will retry. | Yes | No change. |
| `DESTINATION_CONNECTION_ERROR` | PostgreSQL could not be reached. Check the destination connection. | Conditional | No affected commit. |
| `DESTINATION_SCHEMA_MISSING` | PostgreSQL schema `analytics` does not exist. Create or choose an existing schema. | No | No affected commit. |
| `DESTINATION_TABLE_MISSING` | PostgreSQL table `public.hubspot_deals` does not exist. Create it first or approve an AI table plan. | No | Block dependent stream. |
| `DESTINATION_COLUMN_MISMATCH` | Model column `deal_amount` is not present in `public.hubspot_deals`. | No | Block dependent stream. |
| `DESTINATION_TYPE_MISMATCH` | Model column `amount` cannot safely write to the destination type. Add an explicit SQL cast. | No | Block dependent stream. |
| `DESTINATION_PK_MISSING` | `public.hubspot_deals` needs a primary or approved merge key for upsert. | No | Block dependent stream. |
| `SQL_INVALID` | The SQL model is invalid for DuckDB. Fix the highlighted expression. | No | Block dependent stream. |
| `SQL_UNSAFE` | The model contains an operation that is not allowed in UI SQL. | No | Block dependent stream. |
| `DBT_FAILED` | Transformation failed in model `fct_hubspot_deals`. | No automatic retry unless classified transient | Block dependent stream. |
| `DELIVERY_FAILED` | Delivery to `public.hubspot_deals` failed. Review the safe run details. | Conditional | Block dependent stream. |
| `CALLBACK_FAILED` | Run completion could not be recorded; callback delivery will retry. | Yes | Do not commit until callback accepted. |
| `CLEANUP_FAILED` | Temporary run cleanup needs operator attention. | Operator retry | State may commit only if callback safely records the run; alert. |

Errors carry `code`, `phase`, optional `stream`, optional `dest_table`,
`retryable`, `safe_message`, and a server-only safe correlation ID. No raw
exception is shown to the frontend.

## 16. Testing plan

### 16.1 Unit tests

- registry uniqueness, aliases, release tiers, scopes, PKs, cursors, and DuckDB
  names;
- empty/unknown resource rejection and selected-resources-only dlt selection;
- token encryption, masking, exact/pattern scrubbers, and no response echo;
- connection-test 401/403/429/5xx classification;
- partial discovery and property classification/caps;
- custom property encoding, chunking, merge, and fallback;
- history-off validation;
- window split, pagination, lookback, dedupe, and equal-timestamp tie handling;
- checkpoint candidate/commit/blocked merge and per-stream reset;
- SQL source reference, safety, output/PK/type validation;
- missing target/no PK hard failures and absence of append fallback;
- callback shape, Phase 3 count, bounded metadata, and status transitions;
- cleanup ordering and finalized cleanup status.

### 16.2 Contract tests

- Python registry equals Go catalog/discovery and frontend connector capability
  snapshots.
- Go dispatch `SourceStreamConfig[]` and Pydantic `SaaSRunConfig` agree.
- SQL model fields include `source_stream_key`, `duckdb_source_table`, and
  `dest_table=schema.table`.
- Python callback fields are accepted and persisted by Go.
- No existing Stripe, Shopify, GitHub, Notion, SQL source, or destination
  snapshots change unexpectedly.

### 16.3 Integration tests

Using mocked HubSpot HTTP plus a real isolated DuckDB and Postgres:

- private token connection test and safe response;
- every MVP stream discovery with available, missing-scope, unavailable, and
  partial combinations;
- contacts, companies, deals, tickets, owners, both pipelines, products, line
  items, and quotes selected individually;
- multiple selected streams with proof that unselected resources make zero API
  calls;
- custom properties on/off and too-large fallback;
- full first run then incremental second run with overlap/dedupe;
- one stream failure with other streams succeeding and only safe checkpoint
  commits;
- transform failure and delivery failure preserve dependent watermark;
- 429 `Retry-After`, timeout, 5xx, and request-budget exhaustion;
- missing table/column/PK/type errors;
- callback retry idempotency and cleanup.

### 16.4 End-to-end tests

For contacts, companies, deals, tickets, and owners at minimum:

`HubSpot -> dlt -> per-run DuckDB -> UI SQL/dbt -> existing Postgres -> callback -> Realtime UI`.

Also cover multiple streams in one run, incremental second run, stream reset,
request-too-large fallback, missing target, invalid token, missing permission,
temporary API failure, and approved DDL proposal without runner-side creation.

### 16.5 Security and tenancy tests

- Seed marker tokens and assert absence from logs, traces, callbacks, errors,
  frontend payloads, pgmq diagnostics, AI prompts, vectors, and Tigris objects.
- Assert raw emails/phones/notes are masked or absent from AI context.
- Test org A cannot read/mutate org B's connection, discovery cache, pipeline,
  run, checkpoint, AI context, artifact, or DDL proposal with real JWTs.
- Verify callback/internal route rejection without correct internal token.
- Verify DDL requires a fresh authorized approval matching the proposal hash.
- Run Supabase security advisors if vector/RLS schema is added.

### 16.6 Regression and performance

- Run Python, Go, frontend unit/build, and Playwright suites.
- Run connector regression for Stripe, Shopify, GitHub, Notion, Postgres, and all
  active destinations.
- Load-test large property catalogs, dense equal-timestamp update windows,
  maximum safe pagination, parallel streams, DuckDB disk limits, and callback
  payload bounds.
- Confirm rate behavior remains below HubSpot account/app limits and error rate
  budgets.

## 17. Documentation checklist

- [ ] Private/static-token app creation with current HubSpot UI screenshots.
- [ ] Exact read scopes by MVP stream and account/product prerequisites.
- [ ] Token storage, rotation, revocation, masking, and AI exclusion.
- [ ] Stable, Phase 2, and unsupported stream tables.
- [ ] Discovery permissions and partial discovery behavior.
- [ ] Default/selected custom properties and request-size fallback.
- [ ] Property history off in MVP and Phase 2 cost warning.
- [ ] Full refresh, incremental windows/lookback, calculated-property limitation,
      and per-stream reset.
- [ ] Postgres TLS, existing-schema/table, PK, grants, and type requirements.
- [ ] AI DDL proposal and explicit approval workflow.
- [ ] One UI SQL/dbt model guide for each MVP stream, with DuckDB source names.
- [ ] Preview masking and SQL safety.
- [ ] Run phases, callback fields, delivered-row usage, warnings, and cleanup.
- [ ] Error code troubleshooting and retry expectations.
- [ ] AI allowed/denied context and artifact retention.
- [ ] Local automated commands, mocked integration setup, HubSpot sandbox/live
      test setup, canary, rollback, and incident runbook.
- [ ] Correct or retire current manuals that claim unsupported incremental or
      inconsistent stream names.

## 18. Acceptance criteria

The connector may be marked `stable` only when all are true:

- [ ] Catalog metadata matches section 3 and HubSpot is source-only.
- [ ] Postgres remains an existing-table destination and no runner path creates a
      client target.
- [ ] Private token connection test is scope-aware, safe, and returns no CRM data.
- [ ] Live discovery returns MVP fields, PKs, modes, permissions, and limitations;
      optional failures are partial.
- [ ] Exactly the ten MVP streams are stable and selected streams only execute.
- [ ] Empty selection fails before any HubSpot data request.
- [ ] All extraction flows through dlt into an isolated per-run DuckDB.
- [ ] Canonical `hubspot.stream` to `hubspot__stream` naming is enforced.
- [ ] UI SQL/dbt is mandatory, validated, previewable, and runs before delivery.
- [ ] Delivery is upsert-only to existing Postgres tables with PK/type/column
      validation and no no-PK append fallback.
- [ ] Contacts, companies, deals, tickets, and owners pass live E2E; all other MVP
      streams pass mocked and controlled integration tests.
- [ ] Incremental second runs read/deliver only the overlapped change set, dedupe,
      and do not miss equal-timestamp records.
- [ ] Checkpoint state is per stream; extraction, transform, or delivery failure
      never advances the affected stream.
- [ ] Full refresh and per-stream reset are safe and audited.
- [ ] Custom-property fallback is visible and does not silently claim full schema
      coverage.
- [ ] Property history remains off and rejected in MVP.
- [ ] Callback contains all strict audit fields, per-stream results, finalized
      cleanup, and Phase 3 delivery counts; Go persists them.
- [ ] DuckDB and temporary project files are removed on success and every failure
      path after state extraction.
- [ ] Tokens and disallowed CRM PII are absent from every tested sink, including AI
      and vector context.
- [ ] Cross-tenant RLS/auth tests pass.
- [ ] Existing connectors and destinations pass regression suites.
- [ ] Unit, contract, integration, E2E, security, build, and docs checks pass in CI.
- [ ] Canary telemetry shows acceptable error, rate-limit, duration, disk, and
      delivery correctness before general availability.

## 19. Known risks

| Risk | Impact | Mitigation |
| --- | --- | --- |
| Current local source is derived from a moving dlt verified source. | Upstream behavior and local patches can diverge. | Pin dependencies, record source commit/version, keep local contract tests, review upstream intentionally. |
| `updatedAt` does not capture calculated-only changes. | Stale calculated values. | Document MVP limitation; add scheduled calculated-property refresh in Phase 2. |
| Search/paging caps and equal timestamps. | Missed records in dense windows. | Window splitting, stable upper bound, ID tie-breaking, overlap, reconciliation tests. |
| Broad custom property catalogs. | Request rejection, high latency, schema churn. | Metadata caps, explicit selection, chunking, JSON long tail, visible fallback. |
| HubSpot account tiers/scopes vary. | Streams appear but are inaccessible. | Live permission-aware discovery and partial catalog. |
| Partial stream success is complex. | Unsafe checkpoint advance or confusing run status. | Explicit dependency graph, candidate/commit lists, atomic Go merge, distinct partial status. |
| dlt may attempt destination schema evolution/internal columns. | Violation of existing-table/no `_dlt_*` contract. | Pre/post introspection, exact column hints, strict integration tests, fail on attempted client-target DDL. |
| Existing UI/Slack/GitHub paths expose append/replace. | Strict invariant violation. | Repository-wide validation and migration to upsert-only before stable launch. |
| Preview and logs can contain CRM PII. | Privacy/security incident. | Same-path masked preview, field classification, allowlists, retention, sink tests. |
| Vector embeddings can retain personal information. | GDPR deletion and tenant-isolation risk. | Sanitize before embedding, classify as personal data, RLS, deletion cascade, retention. |
| AI-generated DDL can be wrong or overprivileged. | Data loss or security exposure. | Proposal hash, human approval, transaction, allowlisted DDL, audit, post-check, no runner DDL. |
| Callback loss after successful delivery. | Destination changed but control plane/checkpoint is stale. | Durable idempotent callback retry; upsert makes delivery replay safe; no checkpoint commit until acceptance. |
| Documentation currently overstates capabilities. | Operators select unsafe modes or trust false incremental behavior. | Documentation correction is a release gate, not follow-up work. |

## 20. Future improvements

1. HubSpot OAuth with refresh-token rotation, reconnect UX, and granular optional
   scopes.
2. Phase 2 catalog streams after each receives full registry, PII, checkpoint,
   and test coverage.
3. Property history with independent retention and cost approval.
4. Archived/deleted record capture with explicit soft-delete modeling.
5. Association streams and relationship models for contact-company-deal-ticket
   graphs.
6. Custom objects with collision-safe canonical names, inspired by Estuary's
   prefixed custom-object handling.
7. Scheduled calculated-property refresh and reconciliation scans.
8. Webhook-assisted near-real-time capture backed by periodic reconciliation,
   while preserving dlt as the extraction/staging boundary.
9. Adaptive rate controller using HubSpot headers, org-wide budgets, and stream
   priority.
10. Schema-drift diffing with AI migration proposals and explicit approval.
11. Safe synthetic preview generation so Copilot never needs real CRM values.
12. Per-stream pause/resume, replay windows, and checkpoint history/rollback.
13. Data-quality dbt tests, freshness SLAs, lineage, and field-level observability.
14. Customer-managed encryption keys and configurable regional artifact storage.
15. A reconciliation job comparing HubSpot counts/hash samples with Postgres
    delivered models without exposing row values.

## Delivery sequence and rollout gate

Implement sections 3–9 in the work-package order from section 10.3. Keep the
catalog status non-stable behind an organization feature flag until the complete
test matrix passes. Canary with internal/test HubSpot accounts first, then a
small opt-in cohort. Roll back by disabling new HubSpot runs while preserving
encrypted connections, checkpoints, and already delivered Postgres data; never
delete customer targets during connector rollback.
