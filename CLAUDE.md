# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Product Overview

MantrixFlow is a B2B SaaS ETL platform. Users connect source databases, configure data pipelines with transforms, and sync data to destination databases.

## Monorepo Structure

```
apps/
├── arcyria-platform/ ← Next.js 16 + React 19 frontend (App Router)
├── server/
│   ├── arcyria-server/ ← Go API server (Fiber + GORM) — ACTIVE
│   └── arcyria-elt/    ← Python FastAPI ELT server (DuckDB-staged ELT) — ACTIVE
└── arcyria-website/ ← Marketing site (Next.js)
apps/arcyria-docs/ ← Mintlify docs — separate repository; clone the current docs remote into this path
cloud.api.mantrixflow.com/  ← NestJS API (DEPRECATED, reference only)
```

**Documentation:** The Arcyria Mintlify source lives in **`apps/arcyria-docs/`** as a separate repository. Its GitHub remote still uses the legacy `MantrixFlow-Docs` slug until that external repository is renamed. Preview with `cd apps/arcyria-docs && npm install && npm run dev`. Edit MDX and `docs.json` only in that repository. The marketing site uses `NEXT_PUBLIC_DOCS_URL` and redirects `/docs/*` to the docs host.

**Important:** `cloud.api.mantrixflow.com` is the old NestJS API being replaced by `apps/server/arcyria-server`. Use it only as a reference. Do not write new NestJS code. Do not import from it.

## Architecture

### Data Flow

```
Next.js App → Go API (/api/v1) → pgmq → ETL Server (dlt) → Destinations
                   ↑                    ↓
            Supabase JWT         Callback to Go API
```

### ETL Architecture (dlt-based)

The ETL service uses **dlt (data load tools)** directly — **not Singer taps or Meltano**:

```
Go API enqueues → ETL /sync (HTTP 202) → dlt pipeline runs async → callback to Go API
                       ↓
            dlt.sources.sql_table (source) → dlt.destinations.postgres (dest)
```

**Key patterns:**
- **Credential encryption:** AES-256 Fernet (`github.com/fernet/fernet-go`). Go encrypts, Python decrypts. Same key, same format.
- **Queue:** pgmq on Supabase. Go API enqueues pipeline runs; worker goroutine polls every 5 seconds and dispatches to ETL.
- **Scheduling:** pg_cron for scheduled pipeline runs.
- **Auth:** Supabase JWT for all user requests (`Authorization: Bearer {token}`). ETL ↔ Go callbacks use `X-Callback-Token` / `X-ETL-Token` headers.
- **ETL is internal only** — no public access, never called directly by frontend.

### dlt Usage

- **Sources:** `dlt.sources.sql_database` for FULL_TABLE and INCREMENTAL; `pg_replication` module for LOG_BASED (PostgreSQL CDC)
- **Destinations:** `dlt.destinations.postgres`, `dlt.destinations.mssql`, `dlt.destinations.sqlalchemy` (generic for MySQL, MariaDB, SQLite, Oracle, CockroachDB)
- **Transforms:** User-provided Python scripts wrapped as `add_map()` on dlt resources
- **Write dispositions:** `append`, `merge` (upsert), `replace` (overwrite)
- **Backends:** `sqlalchemy` (default) or native dlt backends
- **State management:** dlt pipeline state stored in `.dlt/pipelines/` under temp directory; checkpoint posted back to Go API

### Replication Methods

| Method | Source Support | dlt Resource | Checkpoint |
|--------|---------------|--------------|------------|
| FULL_TABLE | All SQL | `sql_table()` | None |
| INCREMENTAL | All SQL | `sql_table(incremental=)` | `last_value` of replication key |
| LOG_BASED | PostgreSQL only | `pg_replication()` | `last_commit_lsn` |

## Commands

### API (Go) — `apps/server/arcyria-server/`

```bash
cd apps/server/arcyria-server
go run ./cmd/server              # Start dev server (default port 5000)
go test ./...
```

**Key env vars:**
- `DATABASE_URL` — Supabase Postgres (required)
- `SUPABASE_JWT_SECRET` or `JWT_SECRET` — JWT validation (required)
- `ENCRYPTION_MASTER_KEY` — 32+ chars, shared with ETL (required)
- `ETL_PYTHON_SERVICE_URL` — internal ETL URL (required)
- `API_PUBLIC_URL` — public callback base URL (required in prod)
- `ETL_INTERNAL_TOKEN` / `CALLBACK_TOKEN` — internal auth secrets

### ELT (Python) — `apps/server/arcyria-elt/`

```bash
cd apps/server/arcyria-elt
python3 -m venv .venv
source .venv/bin/activate
pip install -r requirements.txt
uvicorn api.main:app --host 0.0.0.0 --port 8000
```

**Key env vars:**
- `ENCRYPTION_KEY` — same as Go server (Fernet format)
- `CALLBACK_URL` — Go server callback endpoint
- `CALLBACK_TOKEN` — for posting to Go API
- `ETL_INTERNAL_TOKEN` — validates incoming requests
- `MAX_CONCURRENT_RUNS` — default 10-20 per instance

### App (Next.js) — `apps/arcyria-platform/`

```bash
cd apps/arcyria-platform
bun install
bun run dev                      # Port 3000
bun run build
bun run start
bun run lint
bun run format
```

**Key env vars:**
- `NEXT_PUBLIC_API_URL` — Go API origin only (required)
- `NEXT_PUBLIC_SUPABASE_URL` (required)
- `NEXT_PUBLIC_SUPABASE_ANON_KEY` (required)

## Database Schema

Tables in Supabase Postgres:
- `etl_source_connections` — encrypted source DB credentials
- `etl_dest_connections` — encrypted destination DB credentials
- `etl_pipelines` — pipeline config with `pipeline_graph` JSONB (nodes, edges, branches)
- `etl_pipeline_runs` — run history with per-branch results
- `etl_sync_states` — incremental cursor state per pipeline

## API Routes Reference

### Go API (`apps/server/arcyria-server/`) — `/api/v1/`

All routes require Supabase JWT except `/api/v1/internal/*`:
- `GET /health` — health check
- `POST /test-connection` — validate DB credentials (calls ETL)
- `POST /discover` — reflect schema/tables/columns (calls ETL)
- `POST /preview` — sample records via dlt (calls ETL)
- `POST /sync` — enqueue pipeline run via pgmq, return job_id
- `POST /cleanup/connection` — drop CDC replication slot
- `POST /internal/etl-callback` — ETL posts completion results
- `POST /internal/checkpoint/:pipelineId` — ETL posts progress updates

### ELT Server (`apps/server/arcyria-elt/`)

All routes validate `X-ETL-Token`:
- `GET/POST /health` — capacity info
- `POST /test-connection` — SQLAlchemy SELECT 1
- `POST /discover` — SQLAlchemy inspect() — returns schemas/tables/columns
- `POST /preview` — dlt `sql_table` with `add_limit(5)` — no destination write
- `POST /sync` — async dlt sync, returns 202 immediately, posts callback when done
- `POST /cdc/verify` — PostgreSQL logical replication checks (slot, publication)
- `POST /cleanup/connection` — drop replication slot

## ETL Request/Response Formats

### POST /sync Request

```json
{
  "run_id": "uuid",
  "pipeline_id": "uuid",
  "org_id": "uuid",
  "connector_type": "postgres",
  "replication_method": "FULL_TABLE|INCREMENTAL|LOG_BASED",
  "source_host": "...",
  "source_port": 5432,
  "source_user": "...",
  "source_password": "...",
  "source_database": "...",
  "source_schema": "public",
  "dest_host": "...",
  "dest_port": 5432,
  "dest_user": "...",
  "dest_password": "...",
  "dest_database": "...",
  "dest_schema": "public",
  "selected_streams": ["table1", "table2"],
  "stream_configs": {"table1": {"dest_table": "custom_name"}},
  "replication_key": "updated_at",  // for INCREMENTAL
  "replication_slot_name": "slot1", // for LOG_BASED
  "emit_method": "append|merge|overwrite",
  "transform_script": "def transform(r): return r",
  "callback_url": "https://api.example.com/api/v1/internal/etl-callback"
}
```

### Callback Payload (ETL → Go API)

```json
{
  "job_id": "uuid",
  "pipeline_id": "uuid",
  "organization_id": "uuid",
  "status": "completed|failed|interrupted",
  "rows_read": 1000,
  "rows_upserted": 1000,
  "rows_dropped": 0,
  "error": null,
  "duration_seconds": 45.2,
  "checkpoint": {
    "engine": "dlt",
    "version": 1,
    "replication_method": "INCREMENTAL",
    "resource": "table1",
    "last_value": "2024-01-15T00:00:00Z"
  },
  "source_tool": "dlt_sql_postgres",
  "dest_tool": "dlt_dest_postgres"
}
```

## Development Status

### Done
- Canvas View (React Flow, dark theme)
- Card View (fan-out layout)
- AI Chat Panel (Vercel AI SDK + ai-elements)
- Connections catalog page

### Needs Implementation (Go API `apps/server/arcyria-server/`)
All routes need implementation. Reference `cloud.api.mantrixflow.com` for expected behavior, then rewrite in Go/Fiber.

### Needs Work (Python ELT `apps/server/arcyria-elt/`)
Routes exist but may have issues:
- test-connection should use SQLAlchemy, not subprocess
- credential handling must be consistent
- callback must post branch results correctly

### Frontend API Wiring
Frontend uses mock data in many places. Replace `MOCK_*` constants with real TanStack Query hooks calling the Go API.

## Coding Standards

### TypeScript/React (`apps/arcyria-platform/`)
- English for all code/docs
- Explicit types (avoid `any`)
- JSDoc for public classes/methods
- PascalCase classes, camelCase variables, kebab-case files
- Early returns, avoid nesting
- RO-RO pattern (object for params/returns)
- TailwindCSS for all styling
- Accessibility attributes on interactive elements
- Event handlers: `handleClick`, `handleKeyDown`

### Go (`apps/server/arcyria-server/`)
- Fiber v2 routing (`app.Get`, `app.Post`, `app.Group`)
- Structured logging with zerolog
- Thin handlers, thick services
- GORM models in `internal/models/`

### Python (`apps/server/arcyria-elt/`)
- FastAPI with Pydantic v2
- dlt for all data movement
- SQLAlchemy for discovery/test-connection
- Async dlt execution, immediate HTTP 202 response

## Security Rules

- Credentials are **never logged**
- Credentials are **never returned** in API responses
- ETL server is **internal only** — no public access
- Go and Python share the same `ENCRYPTION_KEY` (Fernet format)
- All pipeline runs flow through pgmq — no direct HTTP trigger from frontend to ETL

## Testing

See `md-docs/testing-local.md` for local testing notes.

```bash
# Docker test databases (ports: 15432, 13306, 27018)
docker compose -f docker-compose.test.yml up -d

# Go API tests
cd apps/server/arcyria-server && go test ./...

# Load tests (k6)
k6 run tests/load-tests/k6-api.js
k6 run tests/load-tests/k6-etl.js
```

## Deployment

See `md-docs/deployment/infrastructure/ovh-microsandbox-runbook.md`,
`md-docs/deployment/frontend/vercel.md`, and
`md-docs/integrations/email/aws-ses-setup.md`.

- **App**: Vercel — `cloud.mantrixflow.com` (root: `apps/arcyria-platform`)
- **API + ELT**: independent OVH VPS targets managed by self-hosted Dokploy; API public at `cloud.api.mantrixflow.com`, ELT private only
- **Email**: AWS SES (Go API `EMAIL_PROVIDER=ses`)

**No Meltano:** The system uses dlt directly — no Singer taps, no Meltano pipelines. ETL fetches connection config from `data_source_connections` and runs dlt pipelines dynamically.
