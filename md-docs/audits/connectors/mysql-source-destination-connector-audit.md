# MySQL source and destination connector audit

Date: 2026-08-08  
Branch in each repository: `feat/mysql-source-destination-connector`  
Overall release status: **not production-ready**

The connector implementation and disposable-container data-movement matrix pass. The remaining release gate is a single authenticated, real local browser → Go API → ELT run that creates and edits saved connections, discovers and previews a source, provisions a destination table, completes a pipeline run, and verifies the final databases. The UI/API health chain was verified and the movement paths were exercised through the production ELT runner, but those two checks were not combined into one authenticated browser session in this audit.

## Repository baselines

The latest `origin/mantrixflow` was fetched before the feature branches were created.

| Repository | Fetched `origin/mantrixflow` SHA |
| --- | --- |
| Frontend (`apps/arcyria-platform`) | `0d74c16ad9254936409b82e8df27a98185544535` |
| Go API (`apps/server/arcyria-server`) | `4df49ce22b9d641604d68cb9075e0af28cf43d27` |
| Python ELT (`apps/server/arcyria-elt`) | `6a8e4c6a2f0ecae330fdd8642607cc76c1019f33` |

## Implemented contract

Canonical connector ID is `mysql`; the default port is `3306`. MySQL is source- and destination-capable, supports `FULL_TABLE` and cursor-based `INCREMENTAL`, and has no CDC/binlog capability. MariaDB remains hidden. PostgreSQL/CockroachDB delivery stays on the existing PostgreSQL path.

Canonical saved configuration:

```json
{
  "host": "mysql.example",
  "port": 3306,
  "database": "analytics",
  "username": "etl_user",
  "password": "create-only-or-replacement-secret",
  "ssl": {
    "enabled": false,
    "rejectUnauthorized": false
  }
}
```

Create requires a real password. Edit and draft testing preserve the encrypted stored value when `password` is omitted, blank, or exactly `***`. The mask is not submitted by the frontend. Optional `dataSourceId` resolution uses the existing organization-owned connection join before any stored secret is decrypted. Masked responses include secret-presence flags and never return plaintext.

`/api/v1/connectors/health` retains its existing top-level fields and carries a connector-capability map with `source`, `destination`, `available`, and optional `reason`. MySQL is exposed by the frontend only when both runtime roles are explicitly available; unknown or unavailable health hides the catalog item and rejects direct create/edit routes.

MySQL namespaces default to the configured database through discovery, preview, pipeline dispatch, destination provisioning, and the UI's existing `destinationSchema` compatibility field. PostgreSQL continues to default to `public`.

Stable MySQL failure codes are:

- `MYSQL_AUTH_FAILED`
- `MYSQL_UNREACHABLE`
- `MYSQL_DATABASE_UNAVAILABLE`
- `MYSQL_TIMEOUT`
- `MYSQL_PERMISSION_DENIED`
- `MYSQL_TLS_FAILED`

## Frontend audit

The active `features/connections` catalog is now the single connector registry. It owns role, default-port, availability, and CDC metadata. The unused `config/connectors.ts`, `config/database-registry.ts`, and `config/connector-types.ts` registries were removed, eliminating 1,037 lines of duplicate/dead configuration.

The MySQL credential flow uses React Hook Form, Zod, and existing shared fields. Create validation covers name, host, database, username, password, and the `1–65535` port range. SSL is limited to disabled or required. Edit copy explains that a blank password preserves the existing value. Test/create/edit/discovery/preview calls remain browser → Go API only.

Extracted focused modules:

- `features/connections/components/connectionForm.schema.ts`
- `features/connections/hooks/use-connector-health.ts`

No table engine was added or migrated. No internal raw-anchor navigation was introduced or needed correction. No empty catch block, commented-out implementation, duplicate dialog, or new custom shadcn-equivalent component was introduced.

Frontend architecture metrics:

| Metric | Result |
| --- | --- |
| Changed/new maintained files | 18 changed/deleted plus 3 new |
| Largest touched maintained source | `features/pipelines/containers/destination-editor-screen.tsx`, 471 lines |
| MySQL form | `features/connections/components/CredentialForm.tsx`, 355 lines |
| Active connector registry | `features/connections/data/connectors.ts`, 339 lines |
| Touched files over 500 lines | 0 |
| Unrelated pre-existing files over 500 lines | `components/ui/ai-prompt-box.tsx` (669), `features/team/components/team-screen.tsx` (546) |

Architecture exceptions: none introduced. The two unrelated pre-existing size violations were not expanded. Full-repository lint remains red from unrelated existing AI-copilot/onboarding formatting and accessibility findings; focused lint for every touched frontend area passes.

## Go API and Oria audit

The Go API validates MySQL host, database, username, port, SSL shape, connector ID, and create/edit password semantics. Draft tests may merge non-secret request fields with a saved password only after an organization-scoped ownership lookup. Connection writes retain the existing recursive encryption path for passwords and nested TLS material.

Discovery, preview, schema evidence, dispatch, and incremental dispatch use connector-aware namespace defaults. Health capabilities are passed through from ELT and fail closed for MySQL when upstream health is unavailable.

ELT client failures now use a bounded typed upstream error. Response bodies are parsed only for bounded structured code/message fields; raw upstream bodies are not logged or reflected. Callback/audit/agent redaction covers DSNs, password-like keys, nested SSL certificate/key values, callback errors, SSE/public messages, tool results, and audit metadata.

Oria retains its existing agents and generic connector tools. Documentation now describes MySQL source/destination roles, database namespace semantics, upsert/primary-key requirements, lack of CDC, and safe error codes. Regression tests preserve schema evidence while rejecting credentials from prompts, tool results, public messages, SSE-compatible public text, and action audit data.

## Python ELT audit

Structured source and destination `ssl` objects survive `/sync`, `RunConfig`, and `SaaSRunConfig`; legacy `source_ssl_mode` and `dest_ssl_mode` remain accepted. SQLAlchemy engines use `mysql+pymysql`, `utf8mb4`, binary-prefix handling, bounded connect timeouts, and PyMySQL-compatible SSL arguments. Nested TLS material is scrubbed from errors and telemetry.

MySQL discovery excludes `information_schema`, `mysql`, `performance_schema`, and `sys` without falling back to them. It reports columns, native types, nullability, indexes, primary/composite keys, and incremental cursor candidates. Preview uses dialect quoting, supports unusual valid table names, and hard-caps results at 100 rows.

Destination databases must already exist. Explicit DDL runs only through the authenticated control-plane endpoint. The delivery runner never creates or alters destination tables. MySQL delivery uses SQLAlchemy's MySQL `insert().on_duplicate_key_update()`, bounded row batches, and the actual reflected primary key, including composite keys. Missing primary keys fail the run; append fallback is prohibited. PostgreSQL-only dlt cleanup is skipped for MySQL and the integration assertions verify no `_dlt_*` tables or columns are produced.

### Destination type matrix

| Logical/native family | MySQL destination type |
| --- | --- |
| Integer families | `BIGINT` |
| Float/real | `DOUBLE` |
| Decimal/numeric | `DECIMAL(38, 10)` |
| Boolean | `BOOLEAN` |
| JSON/object/list | `JSON` |
| Date | `DATE` |
| Time | `TIME` |
| Timestamp/datetime | `DATETIME(6)` |
| UUID | `CHAR(36)` |
| Binary/blob/bytea | `LONGBLOB` |
| String primary key | `VARCHAR(255)` |
| Other Unicode text | `TEXT` with table `utf8mb4` |
| Composite primary key | Table-level `PRIMARY KEY (...)` |

## Security and tenancy review

- No Supabase schema migration is required or included.
- No direct browser/client access to `data_source_connections` was added.
- Every saved-secret read continues through the Go API's `data_sources.organization_id` ownership join before decryption.
- No MySQL-specific public HTTP endpoint was added; generic test, discovery, preview, destination-DDL, and sync contracts are reused.
- Passwords and nested CA/client certificate/key fields use the existing encrypted JSON mechanism.
- Responses are masked and expose only boolean secret-presence metadata.
- DSNs, nested TLS material, passwords, raw ELT bodies, and driver exception text are excluded from Go logs, callbacks, audits, Oria prompts/results/messages, and ELT MySQL run failures.
- Runtime-generated integration credentials are random and disposable; no personal database was used.

The current RLS boundary remains appropriate: clients have no direct connection-secret table path, while organization isolation is enforced in the trusted Go service. No new RLS policy is needed.

## Data-movement acceptance evidence

Disposable `mysql:8` and PostgreSQL containers use runtime-generated credentials, separate `mantrixflow_source` and `mantrixflow_destination` MySQL databases/users, and Unicode/JSON/DECIMAL/NULL/datetime/composite-key fixtures.

The required integration file passes all three tests in one run:

1. MySQL → PostgreSQL `FULL_TABLE`, PostgreSQL → MySQL, and MySQL source database → distinct MySQL destination database, including a repeat run and duplicate prevention.
2. Two-run MySQL cursor-based `INCREMENTAL`, including checkpoint advancement, one update, one insert, and duplicate prevention.
3. TLS-required and TLS-disabled connectivity, restricted-user and bad-password categorization, system-schema filtering, an unusual table name, Unicode data, cursor/key discovery, preview bounded to 100, and absence of `_dlt_*` artifacts.

## Exact validation commands and results

### Frontend

| Command | Result |
| --- | --- |
| `bun install --frozen-lockfile` | Pass; 897 installs checked across 736 packages, no changes |
| `bun run typecheck` | Pass (`tsc --noEmit`) |
| `bunx biome check features/connections app/workspace/connections features/pipelines/containers/destination-editor-screen.tsx tests/connections/mysql-connector.spec.ts` | Pass; 41 files checked |
| `bun run build` | Pass; pre-existing DuckDB WASM critical-dependency warning only |
| `bunx playwright test tests/connections/mysql-connector.spec.ts --project=chromium` | Pass; 8 tests (2 auth setup + 6 MySQL cases) in 17.5s, including pipeline selectors and target-database defaulting |
| all configured Playwright browser projects | Chromium and mobile Chromium cases passed; Firefox/WebKit projects could not launch because their local Playwright browser binaries are not installed |
| full `bun run lint` | Fail; 23 unrelated pre-existing AI-copilot/onboarding formatting/accessibility errors plus existing warnings |
| maintained-file length audit | Pass for touched files; two documented unrelated pre-existing violations remain |
| `git diff --check` | Pass |

### Go API

| Command | Result |
| --- | --- |
| `gofmt -w` on changed Go sources | Pass |
| `go vet ./...` | Pass |
| `go test ./...` | Pass; all packages |
| `go test -race ./...` | Pass; all packages |
| `go build ./...` | Pass |
| `GOCACHE=/tmp/mantrixflow-go-build go test ./internal/elt ./internal/server ./pkg/response` | Pass |
| `git diff --check` | Pass |

### Python ELT

| Command | Result |
| --- | --- |
| `.venv/bin/python -m pytest -q -m 'not integration'` | Pass; 204 passed, 8 skipped, 3 deselected, 36 subtests in 5.94s |
| `.venv/bin/python -m pytest tests/test_health_capabilities.py tests/test_connection_utils.py tests/test_destination_ddl.py tests/test_delivery_handler.py -q` | Pass; 35 passed in 1.07s after final health/error-boundary hardening |
| `.venv/bin/python -m pytest tests/test_mysql_pipeline_integration.py -q -s` | Pass; 3 passed, 4 dependency deprecation warnings in 87.18s |
| `docker build -t mantrixflow-elt-mysql-check:local .` | Pass; final image manifest `eddf541eb01a16b78c8f8361a1dafc8fdb7cfa6fb2c33573a0115c79eccc49b2` |
| `git diff --check` | Pass |

Local health evidence:

- frontend responded on `localhost:3000`;
- Go `/api/v1/health` reported API, database, ELT, and queue operational;
- ELT `/health` reported MySQL source and destination available;
- Go `/api/v1/connectors/health` propagated MySQL `available/source/destination: true`.

## Remaining release gates and limits

1. Execute the authenticated real local browser → Go → ELT workflow as one continuous acceptance run and record connection create/edit, discovery, preview, destination provisioning, run completion, and final database assertions.
2. Resolve or explicitly waive the unrelated pre-existing frontend full-lint failures before treating full repository lint as green.
3. TLS-required coverage verifies encrypted transport with certificate verification disabled. A production CA-verified/client-certificate environment should be exercised before enabling deployments that require private CA or mutual TLS material.
4. Keep the three PRs draft and retain this **not production-ready** status until gate 1 and the required security acceptance review are signed off.

Rollout order remains ELT first, Go second, frontend last. The frontend's fail-closed runtime health gate prevents early exposure.

## Publication status

The three repositories are on `feat/mysql-source-destination-connector`, but commits, pushes, and draft PR creation are pending because the required GitHub CLI is not installed in the local environment (`gh: command not found`). No partial publication was attempted. After `gh` is installed and authenticated, open three draft PRs titled `feat: enable MySQL source and destination connector`, target each repository's `mantrixflow` branch, and cross-link their final URLs here and in each PR body.
