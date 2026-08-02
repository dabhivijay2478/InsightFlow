# Oria agent runtime setup

This is the private operator setup for the Go ADK runtime. It intentionally
does not publish internal capability names, prompts, routing rules, or tool
implementations.

## Runtime ownership

- The Go API owns orchestration, OpenRouter access, tools, session history,
  usage, and audit records.
- The Next.js app authenticates users, proxies the Go SSE response, and renders
  Oria.
- The browser must never receive `OPENROUTER_API_KEY`, a Supabase service-role
  key, or `INTERNAL_TOKEN`.
- Phase 1 tools are read-only. Startup fails if a registered tool is
  mutation-capable.

## 1. Required Go server environment

Add these values once to `apps/server/main-server/.env`:

```dotenv
AGENT_RUNTIME_ENABLED=true
AGENT_RUNTIME_PROVIDER=openrouter

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
AGENT_ALLOW_MODEL_FALLBACK=false

AGENT_MAX_TRANSFERS=2
AGENT_MAX_TOOL_CALLS=5
AGENT_MAX_MODEL_TURNS=6
AGENT_MAX_OUTPUT_TOKENS=1200
AGENT_CONTEXT_MAX_BYTES=50000
AGENT_TOOL_RESULT_MAX_BYTES=20000
AGENT_REQUEST_TIMEOUT_MS=60000
AGENT_TOOL_TIMEOUT_MS=20000
```

Each selected model must support OpenRouter Chat Completions and the features
needed by its tier. Do not configure an embedding-only model as a chat tier.

`AGENT_ALLOW_MODEL_FALLBACK=false` is the production-safe default. It makes the
server fail at startup if any tier is missing. For local evaluation only, set it
to `true` and provide at least one model tier; the configured model will fill
missing tiers.

The `OPENROUTER_MODEL_*_FALLBACKS` values are separate, explicit provider
fallback lists. They are sent to OpenRouter in priority order and remain active
when `AGENT_ALLOW_MODEL_FALLBACK=false`. This protects free-model traffic from
temporary upstream `429`, provider downtime, and unavailable model capacity
without silently filling a missing tier at server startup. Use comma-separated
model IDs to add more than one fallback.

Free model access is still subject to the selected upstream provider's shared
capacity. An OpenRouter account with sufficient credits can have a higher daily
request allowance while an individual free model returns an upstream `429`.
This is why every production tier should have a compatible fallback.

Do not define the same variable twice. Most dotenv loaders use the last value,
so a later blank line such as `OPENROUTER_API_KEY=` can override a valid key.

The normal Go service variables are still required, including `DATABASE_URL`,
Supabase JWT verification settings, `INTERNAL_TOKEN`, encryption settings, and
the ELT service configuration documented in the Go server README.

## 2. Frontend environment

The frontend needs its normal Go API and Supabase authentication configuration.
Do not add OpenRouter keys, model IDs, provider variables, or a public Oria
feature flag to `apps/app/.env`.

## 3. Start the services

Start the Go API first:

```bash
cd apps/server/main-server
go run ./cmd/server
```

On first startup, GORM creates or upgrades the private agent tables. To apply
the canonical Supabase RLS bundle once during startup:

```dotenv
APPLY_SUPABASE_RLS_ONCE=true
```

After the checksum is current, the value can be removed or set to `false`.

Then start the frontend:

```bash
cd apps/app
bun run dev
```

Open `http://localhost:3000/agents` while signed in to a workspace.

## 4. Verify the runtime

Health:

```bash
curl http://localhost:5000/api/v1/health
```

Metrics require the server-only internal token:

```bash
curl -H "X-Internal-Token: $INTERNAL_TOKEN" \
  http://localhost:5000/api/v1/internal/agent/metrics
```

Automated verification:

```bash
cd apps/server/main-server
GOCACHE=/tmp/mantrixflow-go-build GOTOOLCHAIN=auto go test ./internal/agents/... ./internal/server ./internal/database ./internal/models
```

Manual multi-turn checks:

1. Attach a pipeline and ask `Give me the schema.`
2. Follow with `Show the last failed run.`
3. Refresh the page and repeat a contextual follow-up.
4. Open the same thread from history and confirm earlier messages load.
5. Confirm responses expose only Oria, generic working states, and safe
   citations—not internal routing identities or raw tool payloads.

## 5. Troubleshooting

### `relation "agent_evidence" does not exist`

The original migration allowed GORM to infer `agent_evidences` while indexes
used the canonical singular table. The current startup migration detects that
legacy plural table, renames it when safe, creates the canonical table, and only
then creates indexes. Pull the current code and run the Go server again.

### Runtime disabled

Confirm `AGENT_RUNTIME_ENABLED=true` exists in the Go server environment and
restart the Go process. Frontend flags do not enable the runtime.

### Startup reports missing models

Configure all four `OPENROUTER_MODEL_*` variables, or explicitly enable local
fallback. Check for duplicate blank assignments in the environment file.

### Provider authentication fails

Rotate any key that has been pasted into logs, screenshots, tickets, or chat.
Store the replacement only in the Go deployment environment.

### Responses time out

Keep total runtime within `AGENT_REQUEST_TIMEOUT_MS`. Verify that the selected
models support chat/tool requests and inspect sanitized run status and provider
status in the activity view or internal metrics.
