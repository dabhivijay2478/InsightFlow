# GitHub Source Connector Audit

## Status

**CURRENT STATUS: AVAILABLE NOW — source only; manual UI and live destination validation pending**

Available Now is the owner's release classification as of 2026-08-27. It
records code support and does not claim independent production certification.

The existing connector was audited and hardened on the repositories' current
branches. No branch was created, no branch was switched, and no merge was
performed. Deterministic GitHub tests pass, but no authenticated live
GitHub-to-destination matrix was available in this workspace. At the user's
explicit direction, the hardened source connector is enabled in the frontend
as an Available Now connector while that live validation remains an acknowledged release
limitation.

## Repository baselines

| Repository | Current branch | HEAD before changes | Newest audited product branch |
| --- | --- | --- | --- |
| Frontend (`apps/app`) | `main` | `bfa5488f6bc4d394863cd888223a5e82477ffa74` | `main` |
| Go API (`apps/server/main-server`) | `mantrixflow` | `71f863f0acdb8773fc4f725bbd9b00a8f90ac21a` | `mantrixflow` |
| ELT (`apps/server/elt-server`) | `mantrixflow` | `00d9978237b14afb22cbd15986c1baa4e419391e` | `mantrixflow` |

Remote refs were fetched and `main`/`mantrixflow` were inspected independently.
Frontend `main` and backend `mantrixflow` are the newest relevant branches by
commit date. Existing separate GitHub platform-integration code lives in the Go
and frontend repositories and was not merged into the data connector.

## Connector contract

| Property | Contract |
| --- | --- |
| Canonical ID | `github` |
| Display/category | GitHub / SaaS & APIs |
| Direction | Source only |
| Discovery / preview | Yes / Yes |
| Incremental | Conditional, resource-specific |
| CDC | No |
| Release stage | Available Now |
| Supported host | GitHub.com only |
| Repository scope | One repository per connection/pipeline |
| Mutation behavior | None; read-only APIs only |

GitHub is absent from the destination registry and destination selection. The
Go capability map is authoritative for source/destination roles. The frontend
definition is reused rather than duplicated and now carries availability and
`cdcCapable: false` metadata.

## Existing implementation and current dlt boundary

The Python implementation under `saas_sources/github/` was retained. A
disposable `dlt init github duckdb` audit was run with dlt 1.30.0. The generated
`.dlt/.sources` file identified verified-source commit
`3957506893a7da821dbcc6acd51c7ca4475d1f53` dated 2026-07-03.

Before hardening, MantrixFlow's `helpers.py` and `queries.py` matched that
verified source byte-for-byte. MantrixFlow already added stable `github__*`
table names and ten REST resources. The hardening preserves the verified
resource/query architecture while adding one shared request-policy client,
explicit pagination completion, partial-error rejection, repository provenance,
and current API fields.

| Resources | Classification |
| --- | --- |
| `issues`, `pull_requests`, `stargazers`, `repo_events` | Derived from the current dlt verified GitHub source |
| `commits`, `releases`, `contributors`, `milestones`, `labels`, `forks`, `branches`, `tags`, `events` | MantrixFlow extensions |

The custom REST resources are not described as upstream verified resources.

## Authentication and credential security

- The Available Now data connector supports Personal Access Token authentication only.
- PAT values are opaque; no `ghp_` or `github_pat_` prefix is required.
- `credential` and `api_key` remain accepted as legacy input aliases. Go
  canonicalizes them to encrypted `access_token` storage.
- Saved tokens are masked on reads and preserved during masked edits.
- Connection test performs `GET /user` and, when owner/repository are present,
  a repository metadata read. It returns safe account and repository metadata.
- Authentication success does not claim every stream permission is available;
  resource-specific required permissions are exposed by ELT discovery.
- Exact configured credentials are scrubbed in addition to prefix-shaped token
  strings. Tests cover `GITHUB_SECRET_NEVER_LEAK_8737`, `ghp_TESTSECRET8737`,
  and `github_pat_TESTSECRET8737`.

MantrixFlow already has a separate GitHub App installation integration for
repository/pipeline operations. Its installation tokens and permissions are
not automatically reused by this data connector. The data connector does not
advertise GitHub App auth until a shared, tested execution-token boundary is
implemented.

## Repository discovery and identity

ELT discovery now pages `/user/repos`, returning only safe repository fields:
repository ID, owner, name, full-name snapshot, private/archived flags, and
default branch. Manual legacy `github_owner`/`github_repo` fields remain valid.

Every extracted root row receives `_repository_id` and
`_repository_full_name`. This prevents provenance loss and gives rename/transfer
reconciliation a stable identity. Connection testing returns the repository ID,
but the current frontend does not yet persist that returned value automatically;
this remains a release blocker for fully automatic rename handling.

Private-resource `404` is reported as not found **or inaccessible**, never as a
definite non-existent repository. Archived repositories are accepted. Deleted
or access-revoked repositories fail repository metadata validation.

## Authoritative resource registry

Python `saas_sources/github/registry.py` is the implementation authority. ELT
discovery exposes sanitized stream metadata to Go and the frontend; Go's old
synthetic GitHub stream list was removed.

| Resource | API | Primary key | Mode | Origin / tier | Important limitation |
| --- | --- | --- | --- | --- | --- |
| Issues | GraphQL | repository ID + number | Full Table | verified / Available Now | nested connections first 100 |
| Pull Requests | GraphQL | repository ID + number | Full Table | verified / Available Now | nested connections first 100 |
| Stargazers | GraphQL | repository ID + user | Full Table | verified / Available Now | not advertised incremental |
| Commits | REST | SHA | Full Table | extension / Available Now | force-push semantics |
| Releases | REST | ID | Full Table | extension / Available Now | metadata only; no asset download |
| Contributors | REST | ID | Full Table | extension / Available Now | not a canonical user table |
| Milestones | REST | ID | Full Table | extension / Available Now | includes closed (`state=all`) |
| Labels | REST | ID | Full Table | extension / Available Now | snapshot |
| Forks | REST | ID | Full Table | extension / Available Now | credential visibility applies |
| Branches | REST | repository ID + name | Full Table | extension / Available Now | single-repository scope |
| Tags | REST | repository ID + name | Full Table | extension / Available Now | single-repository scope |
| Events | REST | ID | Full Table | extension / Available Now | recent-window snapshot |
| Repo Events | REST | ID, cursor `created_at` | Incremental | verified / experimental | legacy, dynamic tables, unadvertised |

Unknown resources now fail with `Unsupported GitHub resource '<name>'.`. Empty
selection fails with `At least one GitHub resource must be selected.`. The old
silent issues/pull-request fallback was removed.

## GraphQL behavior

- Root pagination follows opaque `pageInfo.endCursor` only while
  `pageInfo.hasNextPage` is true.
- Missing/repeated cursors fail instead of looping.
- HTTP 200 responses containing `errors[]` fail the resource; partial data is
  never reported as complete.
- Query cost is accumulated against a per-run GraphQL cost budget.
- The root issue/PR schema now includes node ID, labels, assignees, milestone,
  and PR draft/base/head/merge metadata where the GraphQL type supports it.
- The verified source's cost-optimized comment-reaction query is retained.

Nested comments, reactions, labels, and assignees still request only the first
100 per parent. Recursive nested pagination was not implemented and the limit
is surfaced in discovery and documentation.

## REST behavior and rate limits

- REST pagination follows GitHub's `Link` `rel=next` contract.
- Empty page length is no longer the pagination authority.
- Repeated paging URLs and page-budget exhaustion fail clearly.
- REST calls centralize `Accept: application/vnd.github+json`,
  `X-GitHub-Api-Version: 2026-03-10`, and
  `User-Agent: MantrixFlow-GitHub-Connector/1.0`.
- The client records `x-ratelimit-limit`, `remaining`, `used`, `reset`, and
  `resource` without logging URLs, tokens, or response content.
- `403`/`429` honor `Retry-After`, then primary reset time, then bounded
  secondary-limit exponential backoff.
- `401`, permission `403`, inaccessible `404`, rate limits, transient 5xx, and
  invalid responses have distinct secret-safe errors.
- Requests, pages, retries, runtime, and GraphQL cost are bounded. Resource
  execution is serial by default; unlimited concurrency was not introduced.

Conditional ETag requests remain an optional future optimization.

## Preview, schema, and table semantics

Preview is capped at 50 root records, performs no dlt state mutation, and uses
the same REST shapes/provenance conventions as extraction where practical.
Issues, PRs, and stargazers use lightweight preview calls to avoid loading large
nested graphs. Discovery marks limitations and returns current top-level schema
hints; nested/dynamic tables still depend on runtime dlt normalization.

Stable table names remain `github__issues`, `github__pull_requests`,
`github__commits`, and the existing `github__*` resource names. The legacy
`repo_events` verified resource still routes by event type. It is retained for
saved pipelines but hidden from new discovery because dynamic event tables are
not yet proven compatible with every downstream validation/delivery path.

## Incremental state

All advertised resources are Full Table. `repo_events` remains the only
incremental resource and continues to use dlt state with `created_at`. It is
not CDC, not an immutable audit log, and not historically complete because
GitHub exposes only a limited recent event window, commonly about 300 events.

State remains within the existing dlt pipeline/checkpoint isolation boundary.
A live page-success/page-failure/load-failure recovery test was not executed,
so `repo_events` remains unadvertised.

## Organization isolation, Oria, and observability

All public connection, discovery, preview, and pipeline operations continue to
flow through existing organization-scoped Go routes and the internal-auth ELT
boundary. Browser-supplied organization IDs are not used as standalone
authority. No new public ELT route or direct frontend-to-ELT call was added.

The existing GitHub platform integration, GitHub settings drawer, pipeline YAML
operations, PR operations, and webhooks were not renamed or reused. No
GitHub-specific agent was added. Connector credentials and issue/PR/comment
content remain excluded from agent-visible contexts.

Safe process-local counters cover connection tests, API/GraphQL requests,
rate-limit waits, GraphQL cost, and extracted preview records. Counters carry no
token, title, body, comment, email, or private repository content.

## Validation results

| Check | Result |
| --- | --- |
| Disposable current dlt source audit | Pass |
| Python GitHub + route/staging focused tests | 30 passed; final GitHub-only rerun 9 passed |
| Full ELT pytest | 393 passed, 17 skipped, 44 subtests passed |
| Go GitHub capability/security focused tests | Pass |
| Full Go test / vet / build | Pass |
| Full Go race suite | Pass |
| REST 250-row pagination fixture | Pass |
| GraphQL 250-root-node pagination fixture | Pass |
| HTTP 200 GraphQL error rejection | Pass |
| Retry-After / reset handling fixtures | Pass |
| Exact token redaction fixtures | Pass |
| Frontend changed-file format + Biome | Pass |
| Frontend TypeScript | Pass |
| Frontend production build | Pass with existing DuckDB WASM warning |
| Frontend GitHub contract test | 2 passed |
| Repository-wide frontend lint | 26 errors / 15 warnings in unrelated pre-existing files |
| GitHub → PostgreSQL | Not tested |
| GitHub → MySQL | Not tested |
| GitHub → MariaDB | Not tested |
| GitHub → ClickHouse | Not tested |
| GitHub → DuckDB | Not tested |
| Existing GitHub integration live regression | Not tested |

Mock tests intentionally keep actual dlt source construction. They cover the
resource contract, unknown/empty selection, opaque PAT behavior, centralized
headers, 250-item REST/GraphQL pagination, loop protection, HTTP 200 GraphQL
errors, request budgets, rate-limit waits, and secret scrubbing.

The largest changed frontend file is `connectors.ts` at 409 lines. No changed
frontend file exceeds 500 lines. Two unchanged pre-existing frontend files
remain above the limit: `features/team/components/team-screen.tsx` (546) and
`features/ai-copilot/server/agent/orchestrator.ts` (526). No page, table,
navigation, client-component, API/UI separation, empty-error, or commented-out
implementation exception was introduced by this connector change.

## Known limitations and production-readiness decision

- Live credentials and disposable destination infrastructure were not present,
  so real GitHub-to-PostgreSQL/MySQL/MariaDB/ClickHouse/DuckDB movement is not
  claimed. The local GitHub CLI account was configured but its token was
  invalid, and no GitHub data-connector token variable was present.
- Frontend repository-picker UX is not enabled; the backward-compatible manual
  owner/repository fields remain.
- GitHub App execution auth, automatic repository-ID persistence, multi-repo,
  nested GraphQL pagination, GHES, ETags, and `repo_events` recovery E2E remain.
- Fine-grained PAT compatibility must be recorded per resource with live tests;
  reactions remain classic-PAT only.
- Existing platform-integration UI/runtime needs an authenticated regression
  run before release.

The implementation is materially hardened and the connector is **ENABLED AS
AVAILABLE NOW**. Full production-readiness is not claimed until the required real
destination matrix and existing-integration regression pass.
