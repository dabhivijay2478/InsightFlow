# MantrixFlow AI Copilot — Phase 1

Phase 1 is one workspace Copilot interface backed by 12 logical, read-only
agents. It can inspect, analyze, validate, explain, and recommend. It cannot
create, update, delete, execute, retry, cancel, publish, save, or schedule
anything.

## User interface and navigation

Users open Oria from the workspace sidebar at `/agents`. The Oria header has
two compact actions:

- **plus** — start a new conversation;
- **clock** — open Chat History;

Oria chooses the appropriate read-only capability from the user's question and
attached workspace context. Internal routing and capability names are not
presented as a user-selectable catalog.

See
[`Use Oria Copilot`](../apps/mantrixflow-docs/user-guide/oria-copilot.mdx) for
the public user flow, example questions, context guidance, history, safety
boundaries, and troubleshooting.

## Request path

```text
Oria workspace
  → Next.js /api/copilot/chat
  → environment-selected AI SDK provider
  → one routed Phase 1 agent
  → typed read-only tool
  → Go /api/v1/organizations/:organizationId/agent/context
  → organization-scoped Postgres metadata
  → Python ELT only when Go invokes an existing validation/test operation
```

The compatibility route at `/api/pipelines/:id/chat` converts the old request
to a resource reference and calls the same handler. It no longer imports a
provider directly, trusts a browser pipeline object, or emits `<action>` tags.

## Provider configuration

Set `AI_PROVIDER=vercel_gateway` or `AI_PROVIDER=openrouter`. Model identifiers
are never hardcoded; configure at least one of `AI_MODEL_DEFAULT`,
`AI_MODEL_FAST`, `AI_MODEL_REASONING`, or `AI_MODEL_CODE`. Missing tiers use the
first configured model.

Vercel AI Gateway requires `AI_GATEWAY_API_KEY`. OpenRouter requires
`OPENROUTER_API_KEY`; `OPENROUTER_APP_URL` and `OPENROUTER_APP_NAME` provide
supported application attribution. Provider keys and model configuration are
server-only.

## Agent and tool registries

The registry lives in `apps/app/features/ai-copilot/config/agent-registry.ts`.
Each definition records its model tier, supported context, intent, read risk,
step limit, and exact tool allowlist. Agents cannot delegate.

The tool registry lives in
`apps/app/features/ai-copilot/server/tools/tool-registry.ts`. Every tool:

- uses a strict Zod input schema;
- is bound to the authenticated organization and trusted resource;
- calls the Go control plane;
- returns a typed result and evidence envelope;
- blocks duplicate successful calls and repeated failures;
- enforces request, result-size, step, and total-call limits.

SQL validation and read-only connection tests are invoked by Go. The Next.js
route never calls the Python ELT service.

## Data sent to the model

The model receives the selected user message, recent UI messages, user role,
organization name/plan, route, a minimal sanitized resource summary, retrieval
timestamps, and evidence identifiers. It may receive schema names, table and
column names, statuses, run counts, sanitized errors, and SQL supplied for
read-only analysis.

It never receives passwords, API keys, bearer/session/refresh tokens, cookies,
private keys, credentialed connection strings, raw connection configuration,
raw authorization headers, payment data, or raw customer rows.

## Persistence and audit

The Go migration system owns `agent_threads`, `agent_messages`, `agent_runs`,
`agent_tool_calls`, and `agent_audit_events`. These public-schema tables are
API-only: RLS is enabled and `anon`/`authenticated` table privileges are
revoked. Go organization middleware and user filters protect every read.

Message content is stored only when `AI_STORE_MESSAGE_CONTENT=true`. Audit
metadata is recursively sanitized, and raw prompts/tool payloads are never
sent to PostHog.

## Testing

Run deterministic unit checks without provider keys:

```bash
cd apps/app
bun test features/ai-copilot/__tests__
bun run lint
bun run typecheck
bun run build
```

Run Go checks:

```bash
cd apps/server/main-server
go test ./...
```

Real-provider smoke tests are optional and must be explicitly enabled with
`TEST_RUN_AI_PROVIDER_SMOKE=true`. Test Vercel AI Gateway and OpenRouter
separately with their own configured model identifiers and keys.
