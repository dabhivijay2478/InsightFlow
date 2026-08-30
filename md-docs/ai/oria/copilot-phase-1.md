# Oria Copilot — AI SDK production architecture

Oria is MantrixFlow's workspace Copilot. **LLM orchestration** is owned by
`apps/app` (Vercel AI SDK v7 + OpenRouter). **Tool execution, persistence,
permissions, and action confirmation** remain in `apps/server/main-server`.

Internal capability identities, prompts, routing rules, and tool names are not
part of the public contract.

## Request path

```text
Oria UI
  → Next.js POST /api/copilot/chat (authenticated SSE)
  → Go POST …/agent/runs/prepare
  → ToolLoopAgent (root Oria or specialist) + OpenRouter
  → Go POST /internal/agent/tools/execute (per tool call)
  → Next.js SSE envelopes (unchanged contract)
  → Go POST …/agent/runs/:runId/finalize
```

The browser sends the latest user message, stable request/message IDs, route,
and resource references. Go validates membership, persists the user message and
run row, executes tools when asked by Next.js, and stores the final assistant
message on finalize. The browser and model never call Python ELT directly.

## Runtime invariants

- Routing is **model-driven** via `transfer_to_specialist` — no keyword routing
  or deterministic LLM bypass in Go.
- Loop guards: max model turns, tool calls, transfers, duplicate tool hash,
  request timeout (Next.js `loop-guard.ts` + AI SDK `stopWhen`).
- Go registry still defines tool allowlists per specialist; `ExecuteToolWithState`
  enforces permissions and release flags.
- Action flows: preview tools enqueue confirmation events; UI confirms via existing
  Go action HTTP handlers.
- SSE event types unchanged for `OriaChatTransport`.

## Server configuration

OpenRouter and loop limits live in **Next.js server env** (see
[`oria-agent-setup.md`](./agent-setup.md)). Go retains release flags and
action guard settings.

Legacy Go `POST …/agent/chat` returns 410 — do not wire new clients to it.

## API and SSE contract

The chat request contains `requestId`, optional `threadId`, stable `messageId`,
the latest `message`, optional validated internal hint, safe route/context
fields, and optional `thinkMode`.

Public SSE types include: `thread.ready`, `run.started`, `working`,
`message.delta`, `message.completed`, `tool.completed`, `action.preview.ready`,
`action.confirmation.required`, `error`, `run.completed`, `run.failed`.

## Migration reference

Full file list and test notes: [`oria-adk-to-ai-sdk-migration.md`](./adk-to-ai-sdk-migration.md).
