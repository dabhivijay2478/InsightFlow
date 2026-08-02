# Oria Copilot — Go ADK production architecture

Oria is MantrixFlow's read-only workspace Copilot. Production orchestration is
owned by `apps/server/main-server` and uses Google ADK Go v2.1.0 with
OpenRouter. Next.js authenticates users, proxies the unbuffered SSE response,
and renders the existing Oria experience.

Internal capability identities, prompts, routing rules, and tool names are not
part of the public contract. There is no Agent Catalog route, drawer, API, or
documentation surface.

## Request path

```text
Oria UI
  → Next.js POST /api/copilot/chat (authenticated streaming proxy)
  → Go POST /api/v1/organizations/:organizationId/agent/chat
  → Oria ADK coordinator
  → one organization-scoped read-only capability
  → direct Go context/tool service
  → PostgreSQL and, only for an existing safe operation, the Go ELT client
```

The browser sends the latest user message, stable request/message IDs, route,
and resource references. Go reloads authoritative thread history, recomputes
membership and role, validates every resource, executes ADK, persists progress,
and streams safe public events. The browser and model never call Python.

## Runtime invariants

- The ADK tree is built once at startup.
- Oria is the root coordinator and has no domain tools.
- Twelve normal `llmagent` children are registered through `SubAgents`.
- Peer transfers are disabled; transfer, model-turn, and tool-call budgets are
  enforced in callbacks.
- Cyclic transfers and repeated identical tool calls are rejected.
- Every registered Release 1 tool must pass the startup read-only assertion.
- Schema and run results use deterministic formatting when authoritative tool
  data is already sufficient, avoiding an unnecessary provider turn.
- Only retryable provider/network failures receive bounded, jittered retries.

## Server configuration

All values belong to the Go deployment. Do not add them to Next.js or expose
them with a `NEXT_PUBLIC_` prefix.

```dotenv
AGENT_RUNTIME_ENABLED=true
AGENT_RUNTIME_PROVIDER=openrouter
AGENT_ALLOW_MODEL_FALLBACK=false
OPENROUTER_API_KEY=...
OPENROUTER_BASE_URL=https://openrouter.ai/api/v1
OPENROUTER_APP_URL=https://cloud.mantrixflow.com
OPENROUTER_APP_NAME=MantrixFlow
OPENROUTER_MODEL_ROUTER=inclusionai/ling-3.0-flash:free
OPENROUTER_MODEL_ROUTER_FALLBACKS=poolside/laguna-xs-2.1:free
OPENROUTER_MODEL_FAST=poolside/laguna-xs-2.1:free
OPENROUTER_MODEL_FAST_FALLBACKS=inclusionai/ling-3.0-flash:free
OPENROUTER_MODEL_REASONING=google/gemma-4-31b-it:free
OPENROUTER_MODEL_REASONING_FALLBACKS=google/gemma-4-26b-a4b-it:free
OPENROUTER_MODEL_CODE=cohere/north-mini-code:free
OPENROUTER_MODEL_CODE_FALLBACKS=inclusionai/ling-3.0-flash:free
AGENT_MAX_TRANSFERS=2
AGENT_MAX_MODEL_TURNS=6
AGENT_MAX_TOOL_CALLS=5
AGENT_REQUEST_TIMEOUT_MS=60000
```

With fallback disabled, every model tier is required and a bad configuration
fails startup. `AGENT_RUNTIME_ENABLED` is the operational kill switch; there is
no public provider/runtime feature flag.

The explicit `OPENROUTER_MODEL_*_FALLBACKS` lists are independent of the
startup-only `AGENT_ALLOW_MODEL_FALLBACK` behavior. OpenRouter tries each list
in priority order when a free model is temporarily rate-limited or unavailable.

## API and SSE contract

The Go chat request contains `requestId`, optional `threadId`, stable
`messageId`, the latest `message`, optional validated internal hint, safe route
context, and up to three resource references. Request ID uniqueness makes the
operation idempotent: active duplicates return `409`, while completed
duplicates replay the persisted answer.

Every SSE envelope contains `id`, `type`, `sequence`, `threadId`, `runId`,
`timestamp`, and safe `data`. Public events describe only Oria phases, safe
workspace retrieval progress, text, citations, and errors. Transfers are
audited internally without disclosing capability identities or hidden
reasoning. Keepalive comments are emitted every 15 seconds.

## Persistence

The runtime reuses `agent_threads`, `agent_messages`, `agent_runs`,
`agent_tool_calls`, and `agent_audit_events`, and adds `agent_transfers` and
`agent_evidence`. Current messages use `oria.ui_message.v1`. Legacy
`assistant-ui/if` history remains readable without rewriting stored data.

Raw provider credentials, authorization headers, connection configurations,
full customer rows, secret-bearing prompts, raw tool inputs, and raw tool
results are never persisted. Tool records contain hashes and bounded sanitized
summaries. All agent tables are API-only with RLS enabled and browser grants
revoked.

## Observability

Runs record sanitized provider/model tier, request/thread/run IDs, selected
internal capability, durations, tokens, transfer/tool counts, provider status,
rate limiting, and failure codes. Prometheus text metrics are available only at
`GET /api/v1/internal/agent/metrics` with the configured internal token.

## Verification

```bash
cd apps/server/main-server
go test ./internal/agents/... ./internal/server
go test -race ./internal/agents/... ./internal/server

cd ../../app
bun run lint
bun run typecheck
bun run build
```

Normal CI includes 216 deterministic prompt cases and local OpenRouter adapter
tests. A live OpenRouter key is optional and must never be required by CI.

The public navigation and troubleshooting guide is
[`Use Oria Copilot`](../apps/mantrixflow-docs/user-guide/oria-copilot.mdx).
