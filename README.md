# Arcyria

Arcyria is a private, multi-tenant data operations platform. It combines
managed data connections, DuckDB-staged ELT pipelines, SQL transformations,
run monitoring, billing, the Oria AI operator, and isolated simulation runs.

Some package names, image names, database fields, and Git remotes still use the
historical `MantrixFlow` identifier. Those identifiers remain part of the
current code and must not be renamed without a coordinated migration.

## Current architecture

```text
Browser
  ├─ Arcyria website (Next.js)
  └─ Arcyria platform (Next.js on Vercel)
       ├─ Supabase Auth
       ├─ product API requests ────────────────┐
       └─ Oria chat (server-side OpenRouter)   │
                                                ▼
                                  Go API and queue workers
                                  ├─ authorization and org scoping
                                  ├─ encrypted connection metadata
                                  ├─ billing, integrations, and audit
                                  ├─ Supabase Postgres, RLS, Realtime, PGMQ
                                  └─ private ELT dispatch
                                                │
                                                ▼
                                  Python ELT (FastAPI + gRPC)
                                  ├─ extract into ephemeral DuckDB
                                  ├─ run dbt-duckdb models
                                  ├─ upsert into existing destinations
                                  └─ callback run evidence to Go

Simulation API → PGMQ → Go simulation manager → on-demand OVH host
                                            → Microsandbox microVM
                                            → Python simulation runtime
```

The browser sends authenticated product operations to the Go API. The main
exception is Oria chat orchestration: a Next.js server route calls OpenRouter,
while Go remains the owner of tool execution, permissions, persistence,
action confirmation, usage, and audit records.

Production infrastructure encoded in this workspace uses Vercel for the
Next.js applications, Supabase for the current database and identity layer,
and persistent OVH VPSs managed through self-hosted Dokploy for Go and Python
services. Tigris-compatible object storage holds Terraform state, backups, and
large simulation evidence. Simulation compute is created on demand in OVH
Public Cloud and is separate from production ELT execution.

## Repository map

Each entry below is an independent Git repository with its own CI/CD and
license, checked out together in this engineering workspace.

| Repository | Path | Responsibility |
| --- | --- | --- |
| Platform | [`apps/arcyria-platform`](apps/arcyria-platform/README.md) | Authenticated Next.js product, Oria orchestration, and operator UI |
| Go API | [`apps/server/arcyria-server`](apps/server/arcyria-server/README.md) | Control plane, authorization, persistence, queues, billing, integrations, and simulation manager |
| Python ELT | [`apps/server/arcyria-elt`](apps/server/arcyria-elt/README.md) | Private extraction, DuckDB/dbt execution, destination delivery, and simulation runtime |
| Website | [`apps/arcyria-website`](apps/arcyria-website/README.md) | Public Arcyria marketing, product, connector, and legal pages |
| Public docs | [`apps/arcyria-docs`](apps/arcyria-docs/README.md) | Mintlify customer documentation and connector guides |
| Infrastructure | [`apps/arcyria-infra`](apps/arcyria-infra/README.md) | OVH Terraform, Dokploy Compose definitions, networking, and bootstrap scripts |
| Engineering docs | [`md-docs`](md-docs/README.md) | Maintained architecture, deployment, operations, integration, audit, and testing guides |

## Runtime boundaries

- Supabase authenticates users; Go enforces organization membership and role
  permissions on product resources.
- Connection credentials are encrypted by the Go service and must never be
  logged, returned to the browser, or stored in pipeline metadata.
- ELT REST and gRPC operations require service credentials. The frontend never
  calls the ELT service directly.
- Pipeline jobs are durable PGMQ messages. The Go worker performs admission and
  disk checks before dispatching to Python.
- The only active pipeline execution path is `duckdb_staged`.
- Source and destination objects use `schema.table`; DuckDB staging uses
  collision-safe `schema__table` names.
- Pipeline delivery is upsert-only. A destination must already exist, expose
  compatible columns, and provide a stable primary or merge key.
- Destination provisioning is an explicit setup operation; the pipeline runner
  never creates a client table during delivery.
- Cursor state is captured before ephemeral DuckDB and dbt artifacts are
  removed, and callback metadata is persisted with the run.
- Production ELT and simulation execution are separate paths. Production
  credentials are not copied into simulation microVMs.

## Local development

Install Bun, Go as declared by the Go module, Python 3.13, and any native
database drivers required by the connectors you are testing. Configure each
service from its checked-in environment example; never commit populated env
files.

Start the three core services in separate terminals:

```bash
cd apps/server/arcyria-elt
python3.13 -m venv .venv
.venv/bin/pip install -r requirements.txt -r requirements-dev.txt
.venv/bin/python -m uvicorn api.main:app --host 127.0.0.1 --port 8000 --loop asyncio
```

```bash
cd apps/server/arcyria-server
go run ./cmd/server
```

```bash
cd apps/arcyria-platform
bun install
bun run dev
```

Default local ports are `3000` for the platform, `5000` for Go, `8000` for ELT
REST, and `8001` when the ELT gRPC server is configured. Go and ELT intentionally
fail startup when required internal tokens are missing.

## Verification

Run checks from the repository that owns the changed code:

```bash
# Platform
cd apps/arcyria-platform
bun run lint
bun run typecheck
bun run build

# Go API
cd apps/server/arcyria-server
go test ./...
go build ./cmd/server ./cmd/simulation-worker

# Python ELT
cd apps/server/arcyria-elt
.venv/bin/python -m pytest tests

# Website
cd apps/arcyria-website
bun run check
bun run test
bun run build

# Public docs
cd apps/arcyria-docs
npm run broken-links
```

Validate links across the combined workspace with:

```bash
node scripts/check-markdown-links.mjs
```

## Engineering references

- [Engineering documentation index](md-docs/README.md)
- [Production architecture](md-docs/infrastructure/architecture/production-architecture.md)
- [Strict ELT pipeline guide](md-docs/architecture/elt/strict-pipeline-guide.md)
- [Simulation platform](md-docs/architecture/simulation/platform.md)
- [Oria runtime setup](md-docs/ai/oria/agent-setup.md)
- [Local development and testing](md-docs/testing/local-development.md)

Automated checks prove repository behavior, not external certification or a
production service-level commitment. Security, privacy, residency, retention,
availability, and compliance claims require explicit legal approval before
publication.

## License

Copyright © 2026 Arcyria. All rights reserved.

This project is proprietary software and is not open source.
Unauthorized copying, modification, distribution, or use is prohibited.
