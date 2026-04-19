# Agents (MantrixFlow)

**Authoritative agent onboarding and specialized review personas:**

[`.cursor/rules/agents-and-orchestration.mdc`](.cursor/rules/agents-and-orchestration.mdc)

**Stack map, local dev commands, and repo conventions:**

[`.cursor/rules/mantrixflow-repo.mdc`](.cursor/rules/mantrixflow-repo.mdc)

**Backend services (scoped rules — apply when editing those trees):**

- [`.cursor/rules/go-main-server.mdc`](.cursor/rules/go-main-server.mdc) — `apps/server/main-server` (Fiber, dispatch, callbacks, queue)
- [`.cursor/rules/python-elt-server.mdc`](.cursor/rules/python-elt-server.mdc) — `apps/server/elt-server` (FastAPI, DuckDB-staged runner)

**Supabase RLS workflow and schema map:**

[`.cursor/rules/supabase-rls.mdc`](.cursor/rules/supabase-rls.mdc)

**ELT (strict invariants + flow diagram):**

- [`.cursor/rules/strict-elt-invariants.mdc`](.cursor/rules/strict-elt-invariants.mdc) — 12 strict ELT invariants
- [`.cursor/rules/elt-flow-diagram.mdc`](.cursor/rules/elt-flow-diagram.mdc) — 5-phase flow

**Human-readable pipeline guide:** [`md-docs/strict-elt-pipeline-guide.md`](md-docs/strict-elt-pipeline-guide.md)

**Frontend structure / TS conventions:** [`.cursor/rules/nextjs-typescript-stack.mdc`](.cursor/rules/nextjs-typescript-stack.mdc)
