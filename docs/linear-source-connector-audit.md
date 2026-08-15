# Linear source connector audit

Audit date: 2026-08-14  
Decision: **NOT READY**  
Implemented release stage: **Beta**  
Direction: **SOURCE ONLY**

At the user's explicit request, Linear is present in the frontend enabled
allowlist as a clickable **Beta** source connector before credential-backed live
E2E certification. Mock/in-process validation is strong, but no real Linear data
was moved through MantrixFlow staging, transformation, and a real destination in
this environment. Enablement does not change this audit's **NOT READY**
production-certification decision.

## 1. Repository heads

All three repositories were fetched, checked out from `mantrixflow`, pulled with
`--ff-only`, and moved to `feat/linear-source-connector` before implementation.
The recorded HEADs are the synchronized base commits; the connector changes are
currently uncommitted.

| Repository | HEAD SHA |
| --- | --- |
| Frontend (`apps/app`) | `0b3edf3b89d5dd64ac267b7db31d5b221b0bd732` |
| Go API (`apps/server/main-server`) | `dcb0ff10793dbce03f4e8741d6719dd57994bb17` |
| ELT (`apps/server/elt-server`) | `987dc04f922d7b2c1a2229590253a571f302f54a` |

## 2. Existing Linear code found

No active Linear connector implementation existed at the synchronized heads:

- no frontend connector or credential registration;
- no Go connector allowlist, capability, encryption, or registry entry;
- no ELT `_SOURCE_BUILDERS` entry or Linear source package;
- no Linear-specific tests.

The existing reusable SaaS connection, schema-cache, pipeline source options,
staged DuckDB, callback, and secret-sanitization paths were extended rather than
duplicated.

## 3. dlt source strategy

The production source is repository-owned under `saas_sources/linear/`. It uses
actual dlt primitives:

- `@dlt.source(name="linear")`;
- `@dlt.resource(name="linear_resource")` as the reusable resource template;
- `dlt.sources.incremental` for verified `updatedAt` cursors;
- per-resource `merge` for incremental mode and `replace` for full-table mode;
- the existing SaaS runner and staged DuckDB delivery path.

There is no production-time `dlt init`, generated source code, custom JSON-file
handoff, or connector-specific destination DDL.

## 4. dlt Linear starter comparison

The dltHub Linear page describes a workspace/source-generation starter based on
`dlt init dlthub:linear duckdb`; it begins with generated REST-source work and
suggests teams/users without incremental loading. Linear is also absent from the
current verified-source catalog. It was therefore treated as starter/context,
not an official verified Linear source.

MantrixFlow's implementation differs intentionally: it is checked into the
repository, uses Linear GraphQL directly, implements eight audited resources,
Relay pagination, per-resource incremental modes, typed discovery, GraphQL error
handling, request and complexity budgets, and staged-state recovery.

References: [dltHub Linear workspace starter](https://dlthub.com/workspace/source/linear),
[dlt verified sources](https://dlthub.com/docs/dlt-ecosystem/verified-sources).

## 5. GraphQL architecture

All traffic uses the single official endpoint
`https://api.linear.app/graphql`. Static operations live in `queries.py`; no
user text is interpolated into query documents. `LinearGraphQLClient` is the
single HTTP, retry, error, pagination, request-budget, runtime-budget, and
complexity-policy boundary shared by connection test, discovery, preview, and
sync.

Every response is checked for HTTP status, `errors[]`, required `data`, and the
expected connection path. A response containing both `data` and `errors` fails;
partial required-field success is never accepted.

The queries were compared against the current public Linear schema and official
developer documentation on 2026-08-14. References:
[GraphQL API](https://linear.app/developers/graphql),
[public schema](https://github.com/linear/linear/blob/master/packages/sdk/src/schema.graphql).

## 6. Authentication and OAuth readiness

The initial mode is a Personal API Key, treated as an opaque value with no
guessed prefix. The frontend sends `credential`; Go deterministically
canonicalizes `credential`/`api_key` to encrypted `access_token`; ELT accepts
that canonical value as `SaaSRunConfig.credential`.

No reusable platform OAuth framework was found, so a one-off Linear OAuth stack
was not created. The connector is marked Beta. **OAuth with read-only scopes is
required before GA multi-customer rollout.**

Linear authentication references:
[API keys and OAuth](https://linear.app/developers/graphql),
[OAuth 2.0](https://linear.app/developers/oauth-2-0-authentication).

## 7. Credential security

- Go encrypts the canonical `access_token` at rest using the existing connection
  encryption path.
- Masked secrets are preserved during edits and never accepted as a new key.
- Masked configuration is returned to the browser; decrypted credentials remain
  server-side.
- The API key is not part of pipeline connector options, selected streams, dlt
  state, callbacks, metrics labels, logs, audit events, Oria context, OpenRouter
  context, or vector memory.
- Linear errors expose fixed safe messages and never copy GraphQL error text.
- A Go test proves the secret is encrypted, masked, preserved, and removed by
  the agent-context sanitizer.

## 8. Connection test

The ELT test performs the real read-only query:

```graphql
query Viewer {
  viewer { id name }
}
```

Success requires a reachable GraphQL endpoint, valid authentication, and a
resolved viewer. Non-empty local input alone is never considered success.

## 9. Connector registry and frontend

Go registers `linear` as SaaS, source-capable, discovery-capable,
preview-capable, incremental-capable, destination-disabled, and non-CDC. The
source registry entry has no synthetic stream list; discovery is authoritative
from ELT. Destination-role requests receive the exact source-only rejection.

Frontend registration includes the Linear brand icon, Beta badge, runtime
availability, source-only capability, API-key form, setup guide, saved-connection
mapping, authoritative preview catalog, and pipeline source scope UI. Linear is
in `ENABLED_CONNECTOR_IDS` at the user's explicit request. The source card and
setup route are available when runtime health reports the source capability; it
is not shown for the unsupported destination role. A Playwright test confirms
the role boundary and API-key form.

## 10. Discovery architecture

ELT discovery authenticates with `Viewer`, then queries accessible teams and
projects using the same client and production query definitions used by sync.
It returns typed streams, columns, schemas, primary keys, supported sync modes,
team metadata, project metadata, and safe limitations. Go persists this payload
to the existing organization/data-source-scoped schema cache. No Go or frontend
fake Linear catalog exists.

Preview is capped at 50 records and one page, uses the production query and
normalizer, and does not load or mutate production dlt incremental state.

## 11. Pipeline source scope

Connection creation asks only for connection name and API key. Team/project
scope is stored in the pipeline source's `connectorOptions` as
`linear_team_ids` and `linear_project_ids`, not in credentials. The frontend
supports all accessible items through an empty selection or explicit team and
project selections.

Applicable resource filters are pushed into GraphQL variables. Issue, project,
cycle, workflow-state, label, and comment extraction does not fetch the entire
workspace and filter those records in Python. References:
[Linear filtering](https://linear.app/developers/filtering).

## 12. Resource catalog and field design

| Resource | Important fields/relationships | Modes | Primary key | Cursor |
| --- | --- | --- | --- | --- |
| `teams` | id, key, name, display/description, archive timestamps | FULL_TABLE | `id` | — |
| `users` | id, name, email, active/admin/guest/app, timestamps | FULL_TABLE | `id` | — |
| `workflow_states` | id, name, type, position, compact team | FULL_TABLE | `id` | — |
| `issue_labels` | id, name, group flag, compact team/parent | FULL_TABLE | `id` | — |
| `projects` | status, lead, lead team, bounded team IDs/names | FULL_TABLE, INCREMENTAL | `id` | `updatedAt` |
| `cycles` | dates, progress, compact team | FULL_TABLE, INCREMENTAL | `id` | `updatedAt` |
| `issues` | identifier/title/Markdown, label IDs, compact relations | FULL_TABLE, INCREMENTAL | `id` | `updatedAt` |
| `comments` | Markdown body, issue/project/parent IDs, compact user | FULL_TABLE, INCREMENTAL | `id` | `updatedAt` |

Discovery types distinguish strings, timestamps, booleans, integers, numerics,
and JSON arrays. Nested high-cardinality issue/comment collections are separate
top-level resources. Project team relations are capped at 50 and fail rather
than silently returning a partial relationship.

## 13. Pagination

Every resource uses forward Relay pagination with `$first`, `$after`, `nodes`,
and `pageInfo`. Cursors remain opaque. Page size is clamped to 1–100 (default
50); page count, total requests, and total runtime are bounded. Missing or
repeated cursors fail safely. A test extracts 175 unique records across four
pages. Reference: [Linear pagination](https://linear.app/developers/pagination).

## 14. Incremental design

Incremental resources use inclusive GraphQL filtering:
`updatedAt: { gte: <last value> }`. dlt's incremental state and `id` merge key
make equal-timestamp boundary rereads idempotent. Tests execute two real dlt
loads, reread equal timestamps, add a later row, retain Unicode and Markdown,
and produce exactly one row per ID.

Resources without a verified `updatedAt` cursor reject INCREMENTAL and remain
FULL_TABLE. Per-stream modes from the structured pipeline are preserved through
Go, ELT config mapping, and the staged runner.

## 15. State recovery

Linear incremental checkpoints are committed only after completed destination
delivery. Failed delivery returns the previous checkpoint and an empty committed
stream list. The existing staged path restores state before extraction and
cleans dlt artifacts before deleting the work directory. Tests prove failed
candidate state does not replace the prior checkpoint.

## 16. Deletion and archive limitations

`includeArchived` is configurable and archived records expose `archivedAt` when
the schema provides it. Polling does not provide hard-delete detection and this
connector is not CDC. Downstream consumers must not infer that an absent record
was deleted. No webhook is automatically created.

## 17. Rate-limit, complexity, and retry handling

The client observes response headers for request limit/remaining/reset,
complexity cost/limit/remaining/reset, and exposes only safe numeric metrics.
Per-request complexity is bounded below Linear's maximum; page size, nested
connections, request count, page count, and runtime are also bounded.

`RATELIMITED` GraphQL errors and HTTP 429 use authoritative reset metadata.
Retries are bounded exponential backoff with jitter. Only 429/RATELIMITED,
502/503/504, timeout, and connection failures are retried. Authentication,
permission, validation, not-found, and bad-variable failures are not retry
storms. Reference: [Linear rate limiting](https://linear.app/developers/rate-limiting).

## 18. Organization isolation

Linear uses the existing organization-scoped data-source, connection,
schema-cache, pipeline-source, preview, and run routes. Decryption occurs only
after organization/data-source lookup. Pipeline scope is stored on the
organization-owned pipeline source, while the reusable credential remains on
the organization-owned connection. No new unscoped route or storage table was
introduced.

## 19. Oria and secret redaction

No Linear-specific agent tool receives connection JSON. Existing safe connector
context may expose connector type, connection ID, selected team/project IDs,
resources, modes, and sanitized error categories. The agent sanitizer test
proves `access_token` and its value are absent. Static scans found no Linear
credential logging, dynamic query interpolation, or GraphQL mutation.

## 20. Test results

### Passed

- Linear-focused ELT: **17 passed, 1 skipped** (plus 9 deselected in the combined
  focused command).
- Linear core source file alone: **16 passed, 1 skipped**.
- Native dlt execution covers all eight resources against DuckDB.
- Go Linear contract: passed.
- Go full test suite: passed.
- Go race suite: passed.
- `go mod tidy`, `gofmt`, `go vet ./...`, and Go server build: passed.
- Frontend changed-file Biome check: 24 files clean.
- Frontend TypeScript: passed.
- Frontend production build: passed with pre-existing DuckDB dynamic-import and
  browser-baseline warnings.
- Playwright Linear Beta catalog and role gate: **3 passed** including setup.
- Credential-gated Linear live Playwright project: **1 skipped** without
  `LINEAR_TEST_API_KEY` and while the production allowlist remains closed.
- Python compile audit: passed.

The Linear tests cover static read-only operations, API-key header behavior,
HTTP 200 GraphQL partial failure, 401, 403, 502, timeouts, bounded retry,
RATELIMITED reset delay, request budget, complexity headers and rejection,
Relay pagination over 150 records, repeated cursors, server-side scope filters,
typed catalog metadata, all resources through real dlt, full/incremental modes,
equal timestamps, Markdown/Unicode, primary-key merge, preview isolation, and
failed checkpoint rollback.

### Failed or unavailable

- Real Linear connection: skipped; `LINEAR_TEST_API_KEY` is absent.
- Linear → PostgreSQL: not run.
- Linear → MySQL: not run.
- Linear → MariaDB: not run.
- Full ELT suite: **289 passed, 3 skipped, 2 failed, 41 subtests passed**. The
  two failures are pre-existing MySQL container delivery cases reporting an
  `AttributeError` through `MYSQL_OPERATION_FAILED`; Linear-focused tests pass.
- Full frontend lint: baseline failed with **26 errors and 8 warnings** in
  unrelated AI-copilot/onboarding files. The connector's changed-file lint is
  clean.
- Frontend file-size audit found two pre-existing maintained files above 500
  lines: `components/ui/ai-prompt-box.tsx` (669) and
  `features/team/components/team-screen.tsx` (546). The largest changed file is
  `features/data-sources/types/data-sources.ts` at 488 lines; the new Linear
  scope component is below 500 lines.

## 21. Known limitations

1. Personal API Key only; OAuth is required before GA.
2. No hard-delete detection or CDC.
3. Real API schema behavior, permissions, and measured query complexity remain
   unverified without a live credential.
4. No real destination data movement was executed.
5. No Linear-to-MySQL/MariaDB type matrix was executed.
6. Project-to-team relationships above 50 deliberately fail instead of loading
   partial relation data.
7. The Beta source setup is enabled at the user's explicit direction before live
   E2E certification; this is a documented production-readiness exception.
8. Unrelated baseline lint, file-size, and MySQL integration failures remain.

## 22. Production-readiness decision

**NOT READY.** The implementation is exposed as a Beta source connector at the
user's explicit direction, but it is not production-certified. To reconsider
readiness:

1. provide a controlled `LINEAR_TEST_API_KEY` or approved read-only OAuth token;
2. run connection, discovery, preview, selected team/project/resource, full and
   incremental sync against real Linear data;
3. prove Linear → staging → transformation → PostgreSQL, MySQL, and MariaDB with
   row/type/state checks;
4. resolve or explicitly accept the repository-wide lint/file-size and ELT
   MySQL baseline failures;
5. rerun the complete Playwright and live destination flow before promoting
   Linear beyond Beta;
6. implement platform OAuth before GA multi-customer rollout.

Suggested PR title: `feat: add Linear source connector with dlt and GraphQL`.
