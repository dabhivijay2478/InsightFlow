---
name: mantrixflow-core
description: Use for any change in this repo to stay aligned with the current MantrixFlow stack (apps/app + apps/server/main-server + apps/server/elt-server), docs layout, and safety constraints around credentials + internal tokens.
---

# MantrixFlow core (repo skill)

## What this repo is

Monorepo with three active projects:

- `apps/app` — Next.js app (UI)
- `apps/server/main-server` — Go API (Fiber) + DB + queue worker
- `apps/server/elt-server` — Python ELT service (FastAPI)

Repo-level docs: `md-docs/`

## Core runtime rules (non-negotiable)

- **Frontend never calls ELT directly** for authenticated product operations. It should call the Go API.
- **Credentials are never logged** and never returned in API responses.
- **ELT is internal**: validate internal auth headers (e.g. `X-ETL-Token`) and callback auth (e.g. `X-Callback-Token`).
- Prefer **safe write modes** (`upsert/merge`) unless explicitly requested.

## “Where to change what”

- **New UI flow / component**: `apps/app/app/**` and `apps/app/components/**`
- **API route or behavior**: `apps/server/main-server/internal/server/**`
- **DB models**: `apps/server/main-server/internal/models/**`
- **Supabase RLS SQL**: `apps/server/main-server/sql/supabase_rls.sql` (see `codex-skills/supabase-rls`)
- **ELT routes**: `apps/server/elt-server/api/routes/**`
- **ELT runner logic**: `apps/server/elt-server/runner/**`

## Docs conventions (keep small + accurate)

- One `README.md` per project:
  - `apps/app/README.md`
  - `apps/server/main-server/README.md`
  - `apps/server/elt-server/README.md`
- Repo-wide notes go in `md-docs/` with kebab-case names and an index `md-docs/README.md`.

## Local run commands (dev)

App:

```bash
cd apps/app
bun install
bun run dev
```

Go API:

```bash
cd apps/server/main-server
go run ./cmd/server
```

ELT:

```bash
cd apps/server/elt-server
./.venv/bin/python -m uvicorn api.main:app --host 0.0.0.0 --port 8000 --loop asyncio
```

