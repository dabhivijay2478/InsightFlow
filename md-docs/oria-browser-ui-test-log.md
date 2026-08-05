## Session — 2026-08-03 (ADK → AI SDK migration)

Architecture: Next.js `/api/copilot/chat` orchestrates; Go `/agent/runs/prepare|finalize` + `/internal/agent/tools/execute`.

| Check | Result | Notes |
| --- | --- | --- |
| Go build + tests | **PASS** | `go build ./...`, `go test ./internal/agents/... ./internal/server/...` |
| TS typecheck | **PASS** | `bun run typecheck` |
| Registry export | **PASS** | 72 specialists + root instruction JSON |
| Go legacy `/agent/chat` (no JWT) | **PASS** | Returns **401** without session; handler is 410 stub when authenticated |
| Next `/api/copilot/chat` unauthenticated | **PASS** | Returns **401** (route wired, not proxying to Go chat) |
| Browser Release 1–6 matrix | **BLOCKED** | `/agents` redirects to login in automation; full matrix requires signed-in session + `OPENROUTER_API_KEY` in `apps/app/.env.local` |

**Operator next steps:** Sign in → re-run Release 1 smoke from [`oria-test-prompts-release1-read.md`](./oria-test-prompts-release1-read.md). Confirm OpenRouter dashboard activity and `message.delta` streaming (new vs old Go-only `message.completed`).

---

## Session — 2026-08-03 (single chat smoke, thread `3c71f20c-be48-42c2-be55-e340052d2af3`)

All Release 1 smoke prompts in **one Oria thread** (no new chats). Think mode off.

| # | Agent (expected) | Prompt | Result | Notes |
| ---: | --- | --- | --- | --- |
| 1 | `oria` | What can you help me with in MantrixFlow? | **PASS** | Oria identity, read-only scope, capability list |
| 2 | `pipeline_context` | Summarize my workspace at a glance | **FAIL** | UI: "Oria needs a moment". Log: `openrouter_stream_failed` — stream ended before completion (Gemma, ~2264 input tokens in thread). **Fixed:** deterministic workspace overview after `list_workspace_pipelines`; `pipeline_context` on FAST tier; Ling reasoning model; lenient stream close. |
| 3 | `pipeline_context` | List all pipelines in this workspace | **PASS** | 4 pipelines with IDs and last-run success timestamps |
| 4 | `learning_help` | I'm new here — where should I start? | **FAIL** | UI: "Oria needs a moment". Log: `agent model turn limit exceeded` on `learning_help` (3330 input tokens) |
| 5 | `schema_discovery` | List selected streams for Stripe All Streams | **PASS** | 34 streams; table with replication methods |
| 6 | `connection_debugger` | Test the Stripe source connection used by Stripe All Streams | **FAIL** | UI: "Oria needs a moment". Log: 10,538 input tokens in thread; Gemma stream/synthesis failed |
| 7 | `run_failure_investigation` | Why did the last Stripe All Streams run fail? | **IN PROGRESS** | Sent in same thread |
| 8–15 | (specialists) | Remaining smoke prompts | **PENDING** | See release1 specialist table below |

**Single-chat findings:** Long threads accumulate context; read specialists now bypass flaky provider synthesis when tool evidence is available. Restart Go API after pulling these fixes, then re-test prompt #2 in the same thread.

---

## Session — 2026-08-03 (workspace overview reliability fix)

| # | Agent | Prompt | Result | Notes |
| ---: | --- | --- | --- | --- |
| 2 | `pipeline_context` | Summarize my workspace at a glance | **FIX APPLIED** | Routes to `list_workspace_pipelines`, formats **Workspace at a glance** deterministically (no OpenRouter synthesis). Requires Go API restart. |

---

## Session — 2026-08-03 (after basic fixes)

| # | Agent (expected) | Prompt | Result | Notes |
| ---: | --- | --- | --- | --- |
| 71 | `pipeline_context` | List all pipelines in this workspace | **PASS** | 4 pipelines returned with table; `specialist=pipeline_context`, `llm_synthesis=true` |
| 2 | `pipeline_context` | Summarize my workspace at a glance | **PASS** | 4 pipelines, ready/idle summary; hybrid routing + workspace overview query fix |

## Session — 2026-08-03 (before fixes)

| # | Agent (expected) | Prompt (Release 1 row) | Result | Notes |
| ---: | --- | --- | --- | --- |
| 1 | `oria` | What can you help me with in MantrixFlow? | **PASS** | Oria identity, capability list, no internal agent names. OpenRouter called; `specialist=oria`, `llm_synthesis=true`. |
| 2 | `pipeline_context` | Summarize my workspace at a glance | **PARTIAL** | Routed to `pipeline_context`. Response: "no pipelines listed" — UI-Testing org may have empty pipeline data or list tool returned empty. |
| 71 | `pipeline_context` | List all pipelines in this workspace | **FAIL** | UI: "Oria needs a moment". Log: run `cancelled` after ~44s; prior `openrouter_request_failed` on `google/gemma-4-31b-it:free` (tool schema 400 from provider). |
| 721 | `learning_help` | How do I set up incremental sync in MantrixFlow? | **PARTIAL** | Routed to `sync_configuration` (Release 2), not `learning_help`. Answer asks for pipeline ID and mentions preview/confirm — good action safety, wrong specialist for a how-to prompt. |

## Release 1 specialist smoke — pending

One `simple-valid` prompt per agent still to run in browser:

| Agent | Row | Prompt |
| --- | ---: | --- |
| `schema_discovery` | 136 | List selected streams for Stripe All Streams |
| `connection_debugger` | 201 | Test the Stripe source connection used by Stripe All Streams |
| `run_failure_investigation` | 266 | Why did the last Stripe All Streams run fail? |
| `sql_safety` | 331 | Validate SQL for the Stripe charges staging model on Stripe All Streams |
| `pipeline_validation` | 396 | Is Stripe All Streams ready to run? |
| `replication_key` | 461 | What replication key is configured for Stripe charges on Stripe All Streams? |
| `sync_mode` | 526 | What sync mode is Stripe All Streams using? |
| `schedule_planner` | 591 | When is the next scheduled run for Stripe All Streams? |
| `billing_usage` | 656 | What's my AI token usage this month? |
| `audit` | 786 | Show recent Oria agent activity in this workspace |

## Blockers observed

1. **Empty workspace data** — `pipeline_context` returns zero pipelines in UI-Testing org; prompts referencing Stripe All Streams / Postgres Incremental Sync may need sample data seeded.
2. ~~**OpenRouter model tool errors**~~ — Mitigated: read specialists now format tool evidence deterministically (no Gemma synthesis required for list/overview). Reasoning tier switched to `inclusionai/ling-3.0-flash:free`; stream parser accepts partial close when content exists.
3. **Routing drift** — Conceptual how-to prompts can route to Release 2 action agents (`sync_configuration`) instead of Release 1 read (`learning_help`).

## Next steps

1. Seed UI-Testing org with Stripe All Streams, Postgres Incremental Sync, HubSpot All Streams (or switch to org with data).
2. Continue Release 1 smoke table above, then Release 2 action prompts with **preview → confirm** verification.
3. Log each row here; compare to **What to verify** column in the release MD file.
