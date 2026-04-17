# Agents + project skills (MantrixFlow)

This repo is set up for multiple coding agents (Cursor agents + Codex-style skills).

## Repo architecture (current)

- **App (Next.js)**: `apps/app`
- **Go API (Fiber)**: `apps/server/main-server`
- **Python ELT (FastAPI)**: `apps/server/elt-server`
- **Repo-level notes**: `md-docs/`

High-level flow:

`apps/app` → `apps/server/main-server` (`/api/v1/...`) → queue/worker → `apps/server/elt-server` → callback back to Go API.

## Cursor agents

Cursor agents live in `.github/agents/*.agent.md`:

- `orchestrator` — delegates to other agents and summarizes status
- `code-reviewer` — bugs/security review
- `architecture-reviewer` — structure/boundaries/scalability
- `compliance-reviewer` — SOC2/GDPR/security posture checks
- `doc-generator` — keeps docs aligned with code

These agents should treat `apps/server/main-server` + `apps/server/elt-server` as the active backend services.

## Project skills (Codex-style)

Project skills live in `codex-skills/`.

- `codex-skills/mantrixflow-core/SKILL.md` — shared repo conventions + “where to look” map
- `codex-skills/supabase-rls/SKILL.md` — Supabase RLS workflow for `apps/server/main-server/sql/supabase_rls.sql`

## Cursor skills (symlinks)

`.cursor/skills/` contains symlinks to your global skills library (FastAPI, senior-frontend, etc.).
Prefer **project skills** (`codex-skills/*`) when instructions must be repo-specific.

