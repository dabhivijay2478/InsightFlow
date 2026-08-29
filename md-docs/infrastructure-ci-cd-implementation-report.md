# Infrastructure and CI/CD implementation report

## 1. Existing capabilities reused

- Go Fiber control plane, organization/RBAC middleware, scheduler, admission
  control, PGMQ producers/workers, callbacks, and SSE.
- Python FastAPI operational surface, dlt/DuckDB ELT engine, TLS gRPC execution
  service, generated Go/Python protobuf types, cancellation, and streaming
  events.
- Simulation PGMQ flow, versioned runtime service, deterministic
  Postgres-to-Postgres world, same ELT implementation, evaluator, evidence,
  artifacts, orphan handling, and frontend SSE status UI.
- Normal PostgreSQL/GORM access through `DATABASE_URL`; Supabase-specific Auth,
  admin, Storage, RLS, and realtime integration is already localized enough to
  keep separate from a future PostgreSQL host move.
- Existing optimized Go and Python Dockerfiles.

## 2. Modified

- Replaced the shared Hetzner/rootless-Docker simulation path with an on-demand
  OVH outer-host lifecycle and Microsandbox microVM manager.
- Normalized immutable GHCR `sha-...` CI/CD and independent Dokploy deployments
  for Go, ELT, and the simulation manager.
- Added `MAX_ELT_CONCURRENCY` while retaining the legacy Python setting alias.
- Added public liveness/readiness endpoints and container probes.
- Converted the infrastructure repository from one shared host to independent
  Go, ELT, future DB, PgBouncer, and manager units.

## 3. Added

- Durable `simulation_hosts` metadata, run-to-host reference, state machine,
  OVH API signing/client, private-network attachment, readiness probe, idle
  termination, and reconciliation.
- Microsandbox 0.6.8 checksum-pinned bootstrap, strict SSH host verification,
  per-run TLS/token credentials, default-deny microVM networking, bounded
  resources, cleanup, and immutable runtime selection.
- PostgreSQL-portable repository contracts for organizations, pipelines, runs,
  checkpoints, and simulations, with a GORM/PostgreSQL implementation.
- Future-only PostgreSQL 17 + PGMQ, PgBouncer, WAL-G/Tigris, backup timer,
  guarded restore tool, and migration runbooks. These are not activated.
- OVH host bootstrap scripts, private-network/WireGuard template, Dokploy
  configuration, rollback, and deployment-independence documentation.

## 4. Go ownership

Go owns scheduling/admission, PGMQ production and simulation consumption,
frontend REST/SSE, ELT/runtime gRPC clients, OVH scaling decisions, outer-host
state, Microsandbox lifecycle, evaluator coordination, evidence metadata, and
cleanup/reconciliation.

## 5. Python ownership

Python continues to own the real dlt/DuckDB execution engine, connector and
transformation runtime, FastAPI health endpoints, TLS gRPC execution server,
simulation world/runtime implementation, and generated Python protobuf types.
Production and simulation use the same engine with different endpoints.

## 6. Infrastructure ownership

Infrastructure defines separate OVH/Dokploy targets, private networking,
firewalls, Compose deployment units, future database/PgBouncer preparation,
backup/restore operations, and secrets placement. The Go manager—not static
IaC—owns dynamic simulation instances.

## 7. Frontend ownership

No frontend rewrite was required. The existing Next.js application already
talks to Go and consumes browser-safe SSE simulation events. It remains an
independent Vercel deployment and never receives protobuf or internal service
credentials.

## 8. Migration and compatibility risks

- Supabase remains production. Starting the prepared database stack or changing
  `DATABASE_URL` is a later, separately approved migration.
- The chosen OVH flavor must expose KVM/nested virtualization. The private
  network must have gateway egress for checksum-pinned bootstrap downloads.
- Strict dynamic-host SSH requires a custom image using a trusted SSH host CA.
- The MVP supports one simulation manager replica and one active microVM.
- OVH flavor/image/network identifiers are regional and must be staging-tested.
- WAL-G restore drills and PGMQ version compatibility are mandatory before a
  future database cutover.
- Generated protobuf sources remain canonical in the Go repository; the Python
  repository checks generated modules but cross-repository generation must be
  coordinated when the contract changes.

## Acceptance mapping

Go, ELT, the manager, runtime image, database, and PgBouncer have independent
deployment definitions. Production ELT is never routed through Microsandbox.
PGMQ remains durable coordination, gRPC remains service transport, SSE remains
browser streaming, and large evidence remains object storage. No Kubernetes,
Nomad, service mesh, Redis, or additional sandbox provider was introduced.

## Verification record

- Go host, sandbox, storage, configuration, database, server, healthcheck, and
  worker packages pass focused tests and `go vet`; the on-demand lifecycle test
  covers provision, lease, idle, and termination transitions.
- The Python suite passes with `436 passed`, `7 skipped`, and one real-Postgres
  test deselected. That test cannot start PostgreSQL here because the execution
  sandbox denies the shared-memory operation used by `initdb`.
- All Go/Python GitHub workflows pass `actionlint`. Service deployments resolve
  the published GHCR digest and update only their own Dokploy application.
- All five Compose definitions render successfully, all bootstrap/backup scripts
  pass shell syntax validation, and all five Dockerfiles pass BuildKit checks.
- No OVH, GHCR, Dokploy, Supabase, or database production state was changed.
  Live staging proof still requires the protected credentials, regional OVH
  identifiers, trusted host-CA image, and private-network routes listed in the
  deployment runbook.
