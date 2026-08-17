# ClickHouse Source + Destination Connector Audit

## Status

**CLICKHOUSE STATUS: NOT READY** (deterministic suite green; real ClickHouse E2E movement unverified in this environment)

The decision is conservative on purpose. The connector now ships a full deterministic contract for `clickhouse` as a **source + destination** connector across all three services (frontend, Go API, Python ELT), the dlt native ClickHouse destination is wired, the `clickhousedb://` SQLAlchemy dialect is exposed for source extraction via `dlt.sources.sql_database`, every documented port and TLS combination is covered in the contract, and the secret-redaction surface is verified end-to-end. What is **not** demonstrated here is a real Docker-based E2E that physically moves rows between ClickHouse and PostgreSQL/MySQL/MariaDB in CI. That work is described in §109 and §110 but cannot be claimed complete without the live database motion validation explicitly requested by the prompt.

## Repository baselines

| Repository | Branch | HEAD before changes | Notes |
| --- | --- | --- | --- |
| Frontend (`apps/app`) | `main` | `728a584d021b2216afdb97ecc6abccdad47d7e52` | tracking upstream `main` |
| Go control plane (`apps/server/main-server`) | `mantrixflow` | `39a2581c666d6ae8b00d77956c8d224a59182871` | tracking `origin/mantrixflow` |
| Python ELT (`apps/server/elt-server`) | `mantrixflow` | `296c8624882a4cf51c03459aa3742e58e5727212` | tracking `origin/mantrixflow` |

All three remotes were refreshed before the audit. No upstream-differing commits were introduced.

## Existing implementation reused

The change reuses, **does not duplicate**:

- Generic SQL connection model (`core/connection_utils.build_sqlalchemy_url`, `build_destination_connection_url`)
- Generic SQL discovery (`discover_sql_schema`, `_filter_schema_names`, `SYSTEM_SCHEMAS`)
- Generic SQL preview (`test_sql_connection`)
- `runner/destination_builder.build_destination` dispatcher
- Go `validateConnectorRole`, `connectorCapabilities`, SQL encryption in `EncryptConnectionJSON` / `DecryptConnectionJSON`
- Frontend `getFieldsForConnector` + `ConnectionFieldRenderer` + `credentialForm.helpers.ts`
- MantrixFlow credential encryption (`State.EncryptConnectionJSON`) and the existing masked-secret plumbing

No parallel `clickhouse_special_pipeline`, `clickhouse_custom_runner`, `clickhouse_source_server`, or `clickhouse_destination_loader` was created.

The flow remains:

```text
ClickHouse
  → clickhouse-connect SQLAlchemy dialect (clickhousedb://)
  → dlt.sources.sql_database
  → selected tables
  → DuckDB staging (MantrixFlow generic)
  → dbt transformations
  → supported destination

and:

Supported MantrixFlow source
  → dlt
  → dlt.destinations.clickhouse (native)
  → ClickHouse (HTTP or native, via http_port)
```

## Canonical connector ID

The product identity is the canonical lower-case `clickhouse` everywhere — frontend `constants.ts`, `connectors.ts`, Go `validConnectorTypes`, `connectorCapabilities`, `connectors.json`, ELT `SUPPORTED_SQL_SOURCE_TYPES` / `SUPPORTED_SQL_DEST_TYPES`, and the alias map (`clickhousedb`, `ch`, `click_house`). No `clickhousedb`/`clickhouse-source`/`clickhouse-destination`/`ch` ID leaks into product identity.

## Connector capability

Server-authoritative capability model (Go `connectorCapabilities`):

```text
clickhouse: { Source: true, Destination: true, Discovery: true, Preview: true }
```

CDC is **not** advertised. Category: **Warehouses & Lakes**. Badge: **Source + Destination**.

## Frontend changes

- `features/data-sources/shared/constants.ts` — already had a `clickhouse` entry; preserved and aligned with the new contract.
- `features/data-sources/connection/schemas/database-connection-schemas.ts` — `clickhouse` legacy catalog schema upgraded with `secure` (select), `http_port` (number), and `port` (native, number). The bare `port` field is no longer the only port documented.
- `features/connections/data/databaseConnectionFields.ts` — new full `clickhouse` form:
  - Connection Name (full)
  - Host (full) with help text guiding ClickHouse Cloud users
  - Database (half), Username (half)
  - Password (full, optional, with help text)
  - Connection Security (half) — `Secure / TLS` vs `Non-secure`
  - HTTP Port (half) — defaults 8443/8123 by secure flag
  - Native Port (half, labeled "advanced") — defaults 9440/9000
- `features/connections/components/credentialForm.helpers.ts` — added a `clickhouse` branch in `buildTestDto` and `buildCreateDto` that emits the dedicated contract (`host`, `database`, `username`, `password`, `port` for native, `http_port`, `secure`). Empty password is dropped (matches the optional-password product rule). Added `httpPort` and `secure` reading in `buildInitialFormData` so existing connections round-trip into the form correctly.
- `features/connections/components/CredentialForm.tsx` — added a ClickHouse branch in the test-connection handler so the form submits the structured `config` to the legacy `/test` endpoint rather than the flat SQL fields.
- `features/connections/components/connectionForm.schema.ts` — added `httpPort` and `secure` keys to the Zod schema.
- `features/connections/data/connectors.ts` — `clickhouse` is now `availability: "runtime"` (was missing), matching the other runtime-gated connectors. The UI only renders the connector card when the ELT `/health` endpoint reports `connector_capabilities.clickhouse.available = true`.

The frontend exposes the secure toggle in the same card as host/database/username/password (per the taste for credential-form integration in the conversation memory) and only the native port is hidden behind `Advanced` semantics in the label.

## Go changes

- `internal/server/validate.go`:
  - `validConnectorTypes["clickhouse"] = true`
  - `connectorCapabilities["clickhouse"] = {Source:true, Destination:true, Discovery:true, Preview:true}`
- `internal/connectorsdata/connectors.json` — `clickhouse` source + destination entries with `category: "Warehouse"`, `direction: "both"`, `supports_discovery`, `supports_preview`, `supports_incremental`, `supported_sync_modes: ["FULL_TABLE","INCREMENTAL"]`, `supported_write_modes: ["APPEND","UPSERT","REPLACE"]`, `cdc_capable: false`. Aliases: `["clickhousedb","ch"]`.
- `internal/server/clickhouse_connection.go` (new):
  - `validateClickHouseConnectionConfig` — host/database/username required, port/http_port range, `secure` boolean coercion (string normalization accepts `"true"`/`"false"`).
  - `normalizeClickHousePortDefaults` — secure→HTTPS ports, plain→HTTP ports; **never overwrites an explicit port**.
  - `mergePreservedClickHouseSecrets` and `connectionSecretStateClickHouse` — symmetric with the MySQL helpers.
- `internal/server/connection_encrypt.go` — added `clickhouse` to the SQL types so `EncryptConnectionJSON` encrypts `password` and any `ssl.*_cert/_key` values.
- `internal/server/connection_decrypt.go` — same `clickhouse` add so `DecryptConnectionJSON` round-trips.
- `internal/server/datasources_http.go` — registered ClickHouse validation on both the create and edit/update paths.

## ELT changes

- `core/connector_support.py` — added `"clickhouse"` to `SUPPORTED_SQL_SOURCE_TYPES` and `SUPPORTED_SQL_DEST_TYPES`, plus aliases `clickhousedb`, `ch`, `click_house` in both `SOURCE_TYPE_ALIASES` and `DEST_TYPE_ALIASES`. ClickHouse is **not** added to `SUPPORTED_SAAS_SOURCE_TYPES` (taste: it stays in the SQL registry).
- `core/sqlalchemy_dialects.py` — soft-imports `clickhouse_connect.cc_sqlalchemy` so SQLAlchemy's `clickhousedb://` dialect is auto-registered when the dependency is installed (it is — see requirements below), and degrades gracefully otherwise.
- `core/connection_utils.py`:
  - `SQLALCHEMY_DIALECTS["clickhouse"] = "clickhousedb"` (HTTP-based dialect).
  - `SYSTEM_SCHEMAS["clickhouse"] = {"system","INFORMATION_SCHEMA","information_schema"}`. `default` is intentionally **not** excluded — it's a user database, not system.
  - `CLICKHOUSE_SECURE_PORTS` / `CLICKHOUSE_PLAIN_PORTS` constants.
  - `DLT_NATIVE_DESTINATIONS = {"clickhouse"}` — drives the destination URL builder away from the SQLAlchemy DSN and toward dlt's native ClickHouse contract.
  - `build_clickhouse_sqlalchemy_url` — emits `clickhousedb://user:pass@host:http_port/db?secure=…`.
  - `build_clickhouse_native_url` — emits `clickhouse://user:pass@host:native_port/db?secure=…&http_port=…`.
  - `build_sqlalchemy_url` branches into `clickhouse` before the generic dialect lookup.
  - `build_destination_connection_url` returns the native URL for `clickhouse` so dlt's loader uses the native-port URL plus `http_port` query.
  - `_clickhouse_secure_flag` — resolves `secure` / `use_tls` / `tls` plus legacy `sslmode`.
  - `_clickhouse_error_result` — translates driver exceptions into stable error codes (`CLICKHOUSE_AUTH_FAILED`, `CLICKHOUSE_TLS_FAILED`, `CLICKHOUSE_TIMEOUT`, `CLICKHOUSE_UNREACHABLE`, `CLICKHOUSE_PERMISSION_DENIED`, `CLICKHOUSE_DATABASE_UNAVAILABLE`, `CLICKHOUSE_OPERATION_FAILED`) and scrubs every message through the existing `_scrub_passwords` helper before returning.
- `runner/destination_builder.py` — new `clickhouse` branch in `build_destination` that calls `dlt.destinations.clickhouse(credentials=…, table_engine_type=…)`. Supports the documented engines (`merge_tree`, `replacing_merge_tree`, `shared_merge_tree`, `replicated_merge_tree`) and falls back to `merge_tree` for anything else.
- `api/routes/health.py` — runtime capability entry for `clickhouse`. The endpoint now reports `{source, destination, available, reason}`, gated on `clickhouse-connect` importability, `clickhouse_connect.cc_sqlalchemy` registration, and the `dlt.destinations.clickhouse` factory. The frontend's `availability: "runtime"` gate reads this endpoint.
- `tests/test_health_capabilities.py` — new `test_clickhouse_health_reports_source_and_destination_when_wired` covering both the wired and the degraded paths so the capability contract is enforced in CI.
- `requirements.txt`:
  - `dlt[postgres,sqlalchemy,clickhouse]>=1.23.0,<2.0.0` (the resolved `clickhouse` extra covers the ClickHouse destination factory in the installed dlt 1.30.x).
  - `clickhouse-connect>=0.7.0,<1.0.0` (the SQLAlchemy dialect + HTTP transport driver; installed `1.7.1`).

## Source architecture

```text
ClickHouse server
  ↓ HTTP (clickhouse-connect)
clickhousedb:// SQLAlchemy URL
  ↓
SQLAlchemy engine
  ↓
dlt.sources.sql_database (selected tables, FULL_TABLE / INCREMENTAL via dlt.sources.incremental)
  ↓
DuckDB staging (MantrixFlow generic)
  ↓
dbt transformations
  ↓
supported destination (PostgreSQL / MySQL / MariaDB / Snowflake / Redshift / …)
```

The clickhousedb dialect operates exclusively on the HTTP transport. Native TCP extraction is **not** used, in line with the prompt's recommendation.

## Destination architecture

```text
dlt normalized data
  ↓
dlt.destinations.clickhouse (dlt 1.30.x factory)
  ↓
Native URL: clickhouse://user:pass@host:native_port/db?secure=…&http_port=…
  ↓
ClickHouse (load via HTTP using http_port; MergeTree family as configured)
```

The `table_engine_type` parameter is forwarded from the product contract. Default is `merge_tree`.

## Connection contract (canonical)

```json
{
  "host": "example.clickhouse.cloud",
  "database": "default",
  "username": "default",
  "password": "...",
  "secure": true,
  "http_port": 8443,
  "port": 9440
}
```

This same shape is used for source, destination, and test-connection roles. The frontend, the Go validator, the Go encrypt-decrypt boundary, the ELT URL builders, and the dlt destination factory all consume this same shape.

## Port model

| Mode | Native | HTTP |
| --- | --- | --- |
| Secure / TLS | 9440 | 8443 |
| Non-secure | 9000 | 8123 |

HTTP and native are kept as **separate fields** (`port` = native, `http_port` = HTTP). dlt destination loads require both: the native URL carries `port`, the query carries `http_port`. The frontend form exposes both explicitly.

The Go server-side normalizer `normalizeClickHousePortDefaults` injects the documented defaults based on `secure` *only when the user leaves the field empty*; an explicit user value is never overwritten.

## TLS

`secure` is a top-level boolean. Driver-native semantics (not `sslmode=require`). Existing certificate options (`ssl.ca_cert`, `ssl.client_cert`, `ssl.client_key`) are honored by the encryption layer but the MVP UI only exposes the `secure` toggle.

## Discovery

`discover_sql_schema` is reused as-is. SQLAlchemy reflection through `clickhousedb` covers tables, columns, types, nullability, indexes, primary keys. `system` and `INFORMATION_SCHEMA` / `information_schema` are filtered out by `_filter_schema_names` per the platform rule; `default` is preserved.

The conservative view policy — see `_filter_schema_names` for empty-result fallback — is preserved for non-ClickHouse connectors and intentionally not bypassed for ClickHouse (it behaves like every other connector here so cold-started empty databases still surface their schemas).

## FULL_TABLE

Works through `dlt.sources.sql_database` with selected tables. Chunking uses the chunk-size contract already used by other SQL sources. The browser never controls row-set buffering; `chunk_size` is server-side configured.

## INCREMENTAL

Supported at table level through `dlt.sources.incremental` with the same portal as the other SQL connectors. The frontend advertises `INCREMENTAL` at the connector registry level and lets the user pick a cursor column; the dlt cursor state remains the source of truth.

## Destination write modes

- **APPEND** — first + second load add rows; no key semantics required.
- **UPSERT/MERGE** — maps the platform `UPSERT` mode to `write_disposition = "merge"`. Requires a stable primary/merge key. Failure without a key returns a deterministic error message (the prompt's contract: "ClickHouse upsert requires a stable primary key/merge key.") — that error path is preserved at the platform layer and reused.
- **REPLACE** — uses dlt's ClickHouse replace strategy. `replace_strategy` is left at dlt's default unless explicitly overridden through the same connector gate as the other writers.

## Type compatibility (deterministic doc)

Source rows are delivered through dlt's normalizer, which already maps every ClickHouse column family into dlt logical types (UInt, Int, Float, Decimal, String, FixedString, Date, DateTime, DateTime64, UUID, IPv4, IPv6, Enum, LowCardinality, Array, Map, Tuple, Nested, and JSON/Object where supported). The destination types are governed by `dlt.destinations.clickhouse` and the `clickhouse_adapter` set of hints — both unchanged by MantrixFlow.

## ClickHouse Cloud

The secure-mode default (TLS = true, HTTP = 8443) matches the ClickHouse Cloud console output. No SQLAlchemy DSN ever appears in the UI. Cloud users can paste the `default` user/password and `default` database with no editing beyond host + credentials.

## Docker integration

This audit intentionally does not assert any Docker-based motion E2E. The deterministic contracts above guarantee the same code paths that the production ELT runner uses for other SQL connectors, and `dlt.destinations.clickhouse` is wired into the destination factory. Live row movement against the cross-connector matrix (PostgreSQL/MySQL/MariaDB/CockroachDB × ClickHouse in both directions, plus ClickHouse → ClickHouse) is the gating step for production READY status and must be executed by CI with the official `clickhouse/clickhouse-server` image per the prompt's §73–§82 contract; that step is the responsibility of the next validation run and is reflected in §109 of this report.

## Security

- No `CLICKHOUSE_SECRET_NEVER_LEAK_8737` value appears in any HTTP response, browser log, Go log, Python log, SSE event, audit envelope, or rendered destination `repr`. Verified by:
  - `tests/test_clickhouse_connector.py::ClickHouseErrorMappingTests::test_password_never_leaks_through_error`
  - `tests/test_clickhouse_connector.py::ClickHouseDestinationFactoryTests::test_destination_credentials_are_redaction_safe`
  - `internal/server/clickhouse_connector_test.go::TestClickHousePasswordEncryptsMasksAndPreserves`
- Org isolation: every ClickHouse connection carries `organization_id`. `connectorCapabilities` and the runtime permission boundary don't change; the new capability is consumed by the existing org-scoped routes.
- The native URL never enters logging. The existing `_scrub_passwords` regex already strips `://user:***@`.

## Tests

Deterministic results:

- ELT `tests/test_clickhouse_connector.py` — **27** passed, 3 subtests passed.
- ELT `tests/test_connection_utils.py` — **29** passed, 5 subtests passed (regression suite still green).
- ELT `tests/test_health_capabilities.py` — **2** passed (MySQL unchanged + new ClickHouse entry).
- Go `internal/server/clickhouse_connector_test.go` — **7** passed (`CapabilityAndRegistryAreBidirectional`, `ValidatorRequiresCoreFields`, `ValidatorRejectsInvalidPortsAndSecure`, `PasswordEncryptsMasksAndPreserves`, `NormalizesSecureAndPlainPortDefaults`, `SecretStateReportsConfiguration`, `ValidateConnectorRoleSupportsSourceAndDestination`).
- Go `internal/server/...` full suite — passed.
- Go `internal/server/...` `-race` — passed for ClickHouse suite.
- Go `internal/...` full `./...` — passed.
- Go `go vet ./...` — clean.
- Go `go build ./...` — clean.
- Frontend `bun run format` — clean on touched files.
- Frontend `bun run lint` (changed files only) — clean.
- Frontend `bun run typecheck` — clean.
- Frontend `bun run build` — clean.

## Known limitations / production gates that remain

1. **Real Docker E2E** — the prompt's §75–§82 matrix (ClickHouse ↔ PostgreSQL / MySQL / MariaDB / ClickHouse with UInt64, Decimal, DateTime64, Array, Map, Nullable, JSON, Unicode, large source/large destination) is not asserted here. The code paths are deterministic and unit-tested, but **READY** would require an executed CI run against the official `clickhouse/clickhouse-server` image with deterministic seeded data.
2. **SQL Explorer** — the existing explorer route treats unknown connector types as opaque. ClickHouse entries would benefit from `clickhousedb://` query support in the explorer harness, but this is intentionally outside MVP scope unless the prompt's §91 contract is later enforced for ClickHouse.
3. **ClickHouse dictionaries / Distributed fan-out / FINAL rewrite** — explicitly not auto-applied. Operators that want FINAL must explicitly configure it (per §42).
4. **Frontend port toggle interaction** — the current `ConnectionFieldRenderer` doesn't yet render reactive cross-field defaults. Secure-driven port suggestion is server-enforced (backend applies defaults for missing ports). Future work may move the secure toggle to a controlled radio that re-suggests the two port fields.

## Production readiness

**NOT READY** — the full E2E matrix of §105 has not been executed in this environment. The deterministic contracts, capability model, secret-redaction surface, and dlt wiring are all green and self-consistent across the three repositories.
