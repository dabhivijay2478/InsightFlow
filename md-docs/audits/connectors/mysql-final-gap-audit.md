# MySQL Source + Destination final gap audit

Audit completed: 2026-08-10 (Asia/Kolkata)

Overall status: **NOT PRODUCTION-READY**

The MySQL connector itself moved real data successfully in all three required directions, including two-run cursor-based incremental extraction and composite-key MySQL upsert. The repository cannot be marked production-ready yet because the complete frontend lint gate is failing on pre-existing unrelated AI/onboarding files, Firefox/WebKit Playwright binaries are not installed in this environment, the ELT image build did not progress past Docker Hub base-image metadata, and live PostHog/vector-memory inspection was not available. No MySQL-critical runtime failure remains in the tested paths.

## Repository SHAs

All repositories were fetched and fast-forwarded before the audit. The final landing branches follow the repository defaults requested by the user: frontend uses `main`; Go and ELT use `mantrixflow`. No feature branch was created. Frontend `mantrixflow` (`0b3edf3b89d5dd64ac267b7db31d5b221b0bd732`) was already an ancestor of frontend `main`, so no additional merge commit was required.

- Frontend target HEAD before landing changes (`main`): `9ebf526d0d39ab4c9897b1937c64910df141bd26`
- Frontend original audited HEAD (`mantrixflow`): `0b3edf3b89d5dd64ac267b7db31d5b221b0bd732`
- Go HEAD before changes: `a87b43870c196f81e11d61b46b89ccb7f72b5a70`
- ELT HEAD before changes: `56bb14aa771f1da50176d63259c689e35cbbed87`

Repositories:

- Frontend: `dabhivijay2478/cloud.mantrixflow.com`
- Go API: `dabhivijay2478/main-server-mantrixflow.com`
- Python ELT: `dabhivijay2478/etl-server-mantrixflow.com`

## 1. Existing implementation found

### Frontend

- The active registry is `features/connections/data/connectors.ts`.
- Canonical ID is `mysql`; category is Database; direction is Source + Destination; default port is 3306; CDC is false.
- `ENABLED_CONNECTOR_IDS` already contains `mysql` on the latest branch.
- Visibility and direct-route access fail closed until `/api/v1/connectors/health` reports the requested MySQL role available.
- The existing form already uses React Hook Form, Zod, shared shadcn fields, required create fields, port bounds, disabled/required SSL choices, and blank-password preservation on edit.
- Saved MySQL connections already appear in source and destination selectors.
- MySQL destination namespace already defaults to the selected connection database while retaining `destinationSchema` on the wire.

### Go API and Oria

- Canonical MySQL validation, role validation, port validation, SSL-shape validation, encryption, masking, secret-preserving edit merge, organization-scoped draft testing, connector health propagation, namespace defaults, safe ELT errors, and redaction already exist.
- The organization join is retained for secret-bearing saved-connection operations.
- Oria uses the generic specialists and tools; no MySQL-specific agent or endpoint exists.
- Existing tests reject credentials/DSNs from agent input, tool output, callbacks, action audit metadata, and prompt context.

### Python ELT

- `mysql+pymysql`, `utf8mb4`, bounded connect timeout, nested SSL handling, `SELECT 1` connection testing, sanitized error categorization, schema discovery, system-database filtering, bounded preview, dialect quoting, MySQL datetime normalization, explicit DDL, and upsert-only MySQL delivery already exist.
- MySQL delivery uses SQLAlchemy's MySQL insert API and requires the actual primary key, including composite keys.
- Pipeline delivery does not create or alter destination tables.
- MySQL health is available only when the dialect, DDL, and delivery implementations load.

## 2. Missing implementation found

One production-facing frontend defect was reproduced in Chrome: the password manager could repopulate an edit-password field after interaction, defeating the intended blank-means-preserve behavior.

Fix:

- Changed the shared password input autocomplete contract to `new-password`.
- Added 1Password and LastPass ignore hints and `data-form-type="other"`.
- Added Playwright coverage for the edit-password autocomplete contract.

Test gaps fixed:

- Added Go tests for nested MySQL secret encryption/masking and two-organization denial.
- Expanded ELT safe error-code coverage.
- Expanded the real Docker incremental test to 100 initial rows, five updates, five inserts, then a zero-change run.
- Added real control-plane DDL and all-type MySQL upsert coverage.
- Added real composite-primary-key MySQL upsert coverage.
- Added wrong-user, wrong-password, missing-database, refused-port, unresolved-host, TLS, and restricted-user integration assertions.
- Moved disposable database data directories to bounded `tmpfs`; this fixed repeat-suite Docker VM disk exhaustion and prevents anonymous database volumes from accumulating test data.

## 3. Dead or duplicate registry paths

The latest frontend branch had already removed the obsolete runtime definitions:

- `config/connectors.ts`
- `config/database-registry.ts`
- `config/connector-types.ts`

No active caller imports them. The feature registry is the single active connector catalog. MariaDB remains hidden and was not enabled.

## 4. Frontend enablement changes

No MySQL enablement code change was needed on the latest branch. Verified resolved capability:

```text
id=mysql
enabled=true
source=true
destination=true
defaultPort=3306
cdc=false
```

Chrome verified:

- MySQL appears in New Connection.
- Badge is Source & Destination.
- Runtime health reports source and destination available.
- Source and destination saved connections are selectable.
- MySQL target label is Target database and defaults to `defaultdb`.

## 5. Credential contract

Verified canonical fields: host, port, database, username, password, and structured SSL.

- Create requires a real password.
- Edit renders an empty password field and explains that blank preserves the stored value.
- A deliberately wrong draft password produced a sanitized authentication error and was not saved.
- Reloading the edit form cleared the draft; a new connection test with the password left blank succeeded using the stored encrypted secret.
- Trusted internal resolution recovered secrets only in memory for final database assertions.
- GET/edit UI never displayed a stored plaintext password.

## 6. SSL fix and verification

No production SSL builder change was needed. The current implementation passes MySQL SSL configuration through SQLAlchemy `connect_args` for PyMySQL rather than copying PostgreSQL `sslmode` query behavior.

Verified by tests:

- disabled: no MySQL SSL argument
- required: PyMySQL SSL enabled
- verification: CA/reject-unauthorized settings affect the PyMySQL connection arguments
- invalid CA: `MYSQL_TLS_FAILED`
- TLS material is omitted from safe responses

The live local MySQL destination also connected successfully with the UI set to Required. The source connected successfully with SSL Disabled.

## 7. Source discovery results

Real Chrome discovery against disposable MySQL 8 returned one application table and 15 real columns, including:

- composite primary key: `tenant_id, id`
- `varchar(255)` with `utf8mb4_unicode_ci`
- `text`, `decimal(20,6)`, `double`, `tinyint`, `date`, `time`, `datetime`, `json`, `varbinary(64)`, and `blob`
- verified incremental cursor: `updated_at`

`information_schema`, `mysql`, `performance_schema`, and `sys` were not shown as selectable source data.

## 8. Preview results

Chrome preview returned bounded rows from the real source table. Values included Gujarati, Hindi, emoji, accented Latin, nested JSON, nulls, binary/blob values, large decimals, dates, and microsecond datetimes. Identifiers were generated by the server/ELT path; no arbitrary browser SQL or browser-to-ELT request was used.

## 9. Full-table results

The Docker integration matrix passed real FULL_TABLE movement:

- MySQL -> PostgreSQL
- PostgreSQL -> MySQL
- MySQL -> MySQL

The maintained three-path integration test passed inside both the targeted matrix and the complete ELT suite.

## 10. Incremental results

Cursor-based MySQL incremental extraction is verified and remains distinct from CDC.

Chrome MySQL -> PostgreSQL:

- run 1: 10,000 rows, success, 0 failed
- source change: five existing rows updated and five rows inserted
- run 2: 10 rows, success, 0 failed
- run 3 without changes: 0 rows, success, 0 failed
- destination rows belonging to the current seed: 10,005
- duplicate composite keys: 0

Chrome MySQL -> MySQL:

- run 1: 10,005 rows, success, 0 failed
- source change: five existing rows updated and five rows inserted
- run 2: 10 rows, success, 0 failed
- final destination rows: 10,010

The Docker incremental test independently passed 100 -> 10 -> 0 with a final 105-row upserted table and advancing checkpoints.

CDC/binlog remains disabled (`cdc=false`).

## 11. Destination results

Destination tables were explicitly created or verified from the authenticated Chrome control-plane UI before runs.

- PostgreSQL -> MySQL: 3 rows written; second FULL_TABLE/upsert run wrote 3 and final count remained 3.
- Stripe -> MySQL: 1 row written; final count 1.
- HubSpot -> MySQL: initial current run failed because HubSpot added `user_id_including_inactive` to a `SELECT *` model. The saved model was changed to an explicit published column list matching the destination contract. Rerun wrote 1 row successfully; final count 1.
- MySQL -> MySQL: 10,005 then 10 rows; final count 10,010.

No destination database was created during a run.

## 12. Write-mode results

- MySQL is upsert-only.
- No append fallback was observed or permitted.
- Single-key upsert passed in PostgreSQL/Stripe/HubSpot destination paths.
- Composite `(tenant_id,id)` upsert passed in Chrome and Docker.
- Updated rows were replaced and new keys inserted without duplicates.
- No `_dlt_*` tables were produced.

## 13. Type compatibility

Verified MySQL destination mapping and round trip:

| Contract | Result |
| --- | --- |
| INT/BIGINT | passed |
| DECIMAL(20,6) and DECIMAL(38,10) | exact Decimal values preserved |
| VARCHAR/TEXT | passed |
| BOOLEAN/TINYINT(1) | passed |
| DATE | passed |
| DATETIME(6)/TIMESTAMP | microseconds preserved; timezone-aware input normalized |
| JSON | nested object, array, null, boolean, number, and Unicode passed |
| binary/blob | passed |
| NULL | passed |
| utf8mb4 | English, Gujarati, Hindi, emoji, and accented Latin passed |
| composite primary key | passed |

Example exact decimal preserved: `12345678901234.123456`.

## 14. Security results

- Organization-isolation test denies the same connection ID from a different organization.
- Secret-bearing internal resolution requires the internal token and organization ID.
- Frontend never calls ELT directly.
- No direct client access to `data_source_connections` was added.
- No Supabase schema/RLS migration was needed.
- Oria remains on generic connector tools and receives IDs, connector type, safe errors, and schema evidence only.

## 15. Secret-redaction results

Sentinel: `MYSQL_SECRET_SHOULD_NEVER_LEAK_8737`.

- Browser failed-auth response displayed only `MySQL authentication failed.`
- Go encryption tests prove password, CA, and client-key plaintext are absent from encrypted config and masked response JSON.
- Go callback/action/chat/tool tests passed for DSN, password, prompt, audit, and tool-result redaction.
- ELT unit and real integration failures returned stable safe codes without the sentinel, SQLAlchemy URL, PyMySQL text, or stack trace.
- The sentinel was never saved to the connection; reloading restored a blank edit field and the stored password still authenticated.

Live third-party PostHog delivery and vector-memory storage were not directly inspectable in this local environment, so this item prevents a READY declaration even though repository security tests passed.

## 16. End-to-end results

Real Chrome flow passed:

- connections catalog visibility and Source & Destination badge
- saved MySQL source and destination editing
- real connection test
- blank-password preservation
- real discovery and preview
- database namespace default
- authenticated destination DDL
- MySQL -> PostgreSQL run and incremental reruns
- PostgreSQL -> MySQL run and upsert rerun
- Stripe -> MySQL run
- HubSpot -> MySQL failure diagnosis, model correction, publish, and successful rerun
- creation and execution of a new MySQL -> MySQL pipeline

The user-provided Aiven hostname did not resolve in DNS during the audit. It was not used. Disposable local MySQL 8 source/destination servers with runtime-generated credentials were used instead.

## 17. Validation commands and results

### Frontend

- `bun install --frozen-lockfile`: passed; no lockfile changes.
- targeted `biome check`: passed.
- `bun run typecheck`: passed.
- `bun run build`: passed; existing DuckDB WASM dynamic-dependency warning only.
- `bunx playwright test tests/connections/mysql-connector.spec.ts --project=chromium`: 8 passed.
- broad multi-browser Playwright command: Chromium/mobile Chromium passed; Firefox/WebKit/mobile Safari could not launch because their local binaries are not installed.
- `bun run lint`: failed with 23 errors and 8 warnings in pre-existing unrelated AI-copilot/onboarding files.
- full non-writing format audit: failed on pre-existing unrelated AI-copilot/onboarding formatting.
- maintained-file audit: largest file is `components/ui/ai-prompt-box.tsx` at 669 lines; `features/team/components/team-screen.tsx` is 546 lines. Largest Connections file is below 500 lines. Neither existing violation was expanded.

### Go

- `gofmt`: passed.
- `go test ./...`: passed.
- `go vet ./...`: passed.
- `go build ./...`: passed.
- `go test -race ./...`: passed.

### ELT

- non-integration: 219 passed.
- targeted Docker MySQL matrix: 5 passed.
- full suite after tmpfs fix: 224 passed, 4 dependency deprecation warnings.
- ELT Docker image build: did not complete in this environment; BuildKit remained at Docker Hub metadata resolution for `python:3.13-slim`.

## 18. Remaining limitations and production readiness

Intentional limitations:

- MySQL CDC/binlog is unsupported and remains disabled.
- Destination databases and tables must be explicitly provisioned; the delivery runner will not create or alter them.
- Upsert requires an actual primary key; there is no append fallback.

Release blockers for a READY declaration:

1. Fix or baseline the existing full frontend lint/format failures.
2. Install Firefox and WebKit Playwright binaries and run the configured cross-browser projects.
3. Complete the ELT Docker image build in an environment with working Docker Hub metadata access.
4. Execute the live PostHog/vector-memory secret-absence assertions in the deployment environment.

Until those checks pass, the audit status remains **NOT PRODUCTION-READY**, despite successful real MySQL movement in both source and destination directions.
