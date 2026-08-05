# Oria agent runtime setup

This is the private operator setup for the Oria agent. It intentionally does not
publish internal capability names, prompts, routing rules, or tool
implementations.

## Runtime ownership (post ADK migration)

- **Next.js** (`apps/app`) orchestrates LLM calls with Vercel AI SDK v7 +
  OpenRouter via `POST /api/copilot/chat`.
- **Go API** (`apps/server/main-server`) owns tool execution, run/message
  persistence, permissions, action preview/confirm, usage, and audit records.
- The browser must never receive `OPENROUTER_API_KEY`, a Supabase service-role
  key, or `INTERNAL_TOKEN`.
- Phase 1 tools remain read-only in Go. Startup assertions still apply to the
  registry.

```text
Browser → Next.js /api/copilot/chat
       → Go prepare run + thread history
       → ToolLoopAgent + OpenRouter (Next.js)
       → Go internal tool execute
       → SSE envelopes (unchanged UI contract)
       → Go finalize run
```

Legacy `POST /api/v1/organizations/:orgId/agent/chat` on Go returns **410 Gone**.

## 1. Next.js server environment (OpenRouter + loop limits)

Add these to `apps/app/.env.local` (development) or Vercel server env (production).
**Never** use `NEXT_PUBLIC_` for these.

```dotenv
OPENROUTER_API_KEY=replace-with-a-server-only-key
OPENROUTER_BASE_URL=https://openrouter.ai/api/v1
OPENROUTER_APP_NAME=MantrixFlow
OPENROUTER_APP_URL=http://localhost:3000

OPENROUTER_MODEL_ROUTER=inclusionai/ling-3.0-flash:free
OPENROUTER_MODEL_ROUTER_FALLBACKS=poolside/laguna-xs-2.1:free
OPENROUTER_MODEL_FAST=poolside/laguna-xs-2.1:free
OPENROUTER_MODEL_FAST_FALLBACKS=inclusionai/ling-3.0-flash:free
OPENROUTER_MODEL_REASONING=google/gemma-4-31b-it:free
OPENROUTER_MODEL_REASONING_FALLBACKS=google/gemma-4-26b-a4b-it:free
OPENROUTER_MODEL_CODE=cohere/north-mini-code:free
OPENROUTER_MODEL_CODE_FALLBACKS=inclusionai/ling-3.0-flash:free

AGENT_MAX_TRANSFERS=2
AGENT_MAX_TOOL_CALLS=5
AGENT_MAX_MODEL_TURNS=12
AGENT_REQUEST_TIMEOUT_MS=60000
AGENT_TOOL_TIMEOUT_MS=20000

# Must match Go API INTERNAL_TOKEN (tool bridge auth)
INTERNAL_TOKEN=replace-with-internal-token
```

Also set the normal frontend public vars: `NEXT_PUBLIC_API_URL`,
`NEXT_PUBLIC_SUPABASE_*`.

## 2. Go server environment (tools + persistence)

Go still requires agent tables, release flags, and action guard settings. OpenRouter
model IDs in Go `.env` are **no longer used** for chat orchestration but may remain
harmless until cleaned up.

```dotenv
AGENT_RUNTIME_ENABLED=true

AGENT_RELEASE2_ENABLED=true
AGENT_RELEASE3_ENABLED=true
AGENT_RELEASE4_ENABLED=true
AGENT_RELEASE5_ENABLED=true
AGENT_RELEASE6_ENABLED=true

AGENT_CONTEXT_MAX_BYTES=50000
AGENT_THREAD_MAX_MESSAGES=24
AGENT_ACTION_TOKEN_EXPIRY_SECONDS=900
AGENT_MAX_PENDING_ACTIONS_PER_THREAD=3
AGENT_MAX_PENDING_ACTIONS_PER_RESOURCE=1

INTERNAL_TOKEN=same-value-as-nextjs
```

The normal Go service variables are still required: `DATABASE_URL`, Supabase JWT
verification, encryption settings, and ELT configuration (see Go server README).

## 3. Start the services

```bash
cd apps/server/main-server && go run ./cmd/server
cd apps/app && bun run dev
```

Sign in, then open `http://localhost:3000/agents`.

## 4. Verification

1. Send a workspace question — response should stream with `message.delta` events.
2. OpenRouter dashboard should show activity for the configured models.
3. Tool calls appear as `working` / `tool.completed` SSE events.
4. Release 2 preview prompts emit `action.preview.ready` and
   `action.confirmation.required`.
5. Duplicate `requestId` returns HTTP 409 from prepare.

## 5. Registry export (prompts/tools)

To regenerate TypeScript registry JSON from Go source:

```bash
cd apps/server/main-server
go run ./cmd/export-oria-registry > ../app/features/ai-copilot/server/agent/registry/oria-registry.json
```

## 6. Troubleshooting

| Symptom | Check |
| --- | --- |
| 502 on chat | `OPENROUTER_API_KEY`, `NEXT_PUBLIC_API_URL`, Go API reachable |
| Tool failures | `INTERNAL_TOKEN` matches on Next.js and Go |
| 410 on Go `/agent/chat` | Expected — UI must use `/api/copilot/chat` |
| Step limit errors | Narrow the question; adjust `AGENT_MAX_MODEL_TURNS` |

See also [`oria-adk-to-ai-sdk-migration.md`](./oria-adk-to-ai-sdk-migration.md).
