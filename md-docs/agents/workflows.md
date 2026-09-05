# Agent Workflows

## 1. Standalone Agent Platform Configuration

Purpose: create or update the agent for one existing pipeline.

Main files:

- `apps/arcyria-platform/app/workspace/agents/page.tsx`
- `apps/arcyria-platform/app/workspace/agents/AgentsPlatformClient.tsx`
- `apps/arcyria-platform/components/agents/*`
- `apps/arcyria-platform/lib/api/services/pipeline-agents.service.ts`
- `apps/server/arcyria-server/internal/server/agent_http.go`
- `apps/server/arcyria-server/internal/models/agent.go`

Route path:

- UI: `/workspace/agents`
- Deep link: `/workspace/agents?pipelineId=:id`
- Compatibility redirect: `/workspace/data-pipelines/:id/agent`
- Go: `/api/v1/organizations/:organizationId/pipelines/:id/agent`

What happens:

1. `/workspace/agents` lists persisted agents from `pipeline_agents`; it does not invent draft agents for every pipeline.
2. `/workspace/agents/new` lets the user select one real existing pipeline for a new worker.
3. The platform loads pipeline details.
4. It discovers source and destination tables for the selected pipeline.
5. The user chooses source/destination allowlists, run permissions, public source-query behavior, and allowed domains.
6. The Go API stores the config in `pipeline_agents`.
7. Go generates a public `agent_key` like `agent_abc123def456`.

## 2. Authenticated Test Chat

Purpose: let workspace users test the agent before embedding it.

Main files:

- `apps/arcyria-platform/app/workspace/agents/AgentsPlatformClient.tsx`
- `apps/arcyria-platform/components/agents/AgentChatShell.tsx`
- `apps/arcyria-platform/app/api/pipelines/[id]/agent/chat/route.ts`
- `apps/server/arcyria-server/internal/server/agent_http.go`

Route path:

- Next: `/api/pipelines/:id/agent/chat`
- Go query: `/api/v1/organizations/:organizationId/pipelines/:id/agent/query`

What happens:

1. The platform chat sends AI SDK messages to Next.js.
2. Next.js loads the agent config using the user's JWT.
3. The selected model receives a prompt with pipeline context, latest run status, table allowlists, and tools.
4. Authenticated tools are `execute_query`, `run_pipeline`, and `get_run_status`.
5. Query tool calls go to Go, not directly to the database.
6. Go validates SQL, scope, and table allowlists before calling ELT.
7. The response streams back to the central chat surface.

## 3. Public Embedded Chat

Purpose: run the same agent from an external website.

Main files:

- `apps/arcyria-platform/public/agent.js`
- `packages/agent-sdk/src/index.tsx`
- `packages/agent-sdk/src/loader.tsx`
- `apps/arcyria-platform/app/api/agents/[agentKey]/chat/route.ts`
- `apps/arcyria-platform/app/api/agents/[agentKey]/info/route.ts`
- `apps/server/arcyria-server/internal/server/agent_http.go`

Route path:

- Widget info: `/api/agents/:agentKey/info`
- Widget chat: `/api/agents/:agentKey/chat`
- Go internal session: `/api/v1/internal/agents/:agentKey/session`
- Go internal query: `/api/v1/internal/agents/:agentKey/query`
- Go internal conversation save: `/api/v1/internal/agents/:agentKey/conversation`

What happens:

1. The host page loads `/agent.js` or renders `<MantrixAgent />`.
2. The widget creates a browser session id in `localStorage`.
3. The widget calls the public Next route with `agent_key`, messages, session id, and browser `Origin`.
4. Next calls Go internal routes using `X-Internal-Token`.
5. Go checks that the agent exists, is active, and allows the request origin.
6. Go enforces the session rate limit.
7. The selected model can only query through the guarded Go query endpoint.
8. The public model route never receives `run_pipeline` or `get_run_status`.
9. Source queries are denied unless `allow_public_source_queries=true`.
10. Next saves the updated conversation through the internal route.

## 4. Read-Only Query Execution

Purpose: let the selected model answer questions without giving it broad database access.

Main files:

- `apps/server/arcyria-server/internal/server/agent_http.go`
- `apps/server/arcyria-server/internal/server/agent_http_test.go`

Guard order:

1. Reject empty SQL.
2. Reject non-SELECT/non-WITH SQL.
3. Reject multiple top-level statements.
4. Extract referenced tables from `FROM` and `JOIN`.
5. Normalize table names to `schema.table`.
6. Reject any table outside the requested scope's allowlist:
   - destination scope uses `pipeline_agents.allowed_tables`.
   - source scope uses `pipeline_agents.allowed_source_tables`.
7. Append or clamp `LIMIT 10000`.
8. Execute through ELT with `timeout_ms: 30000`.

Public source-scope execution has one extra check: `allow_public_source_queries` must be true. Authenticated workspace chat can query source tables if the table is explicitly allowlisted.

## 5. Pipeline Worker Tools

Purpose: let an authenticated workspace user operate the pipeline from the agent chat.

Main files:

- `apps/arcyria-platform/app/api/pipelines/[id]/agent/chat/route.ts`
- existing org-scoped pipeline run APIs

Available only in authenticated chat:

- `run_pipeline`
- `get_run_status`
- `execute_query`

Not available in public embeds:

- `run_pipeline`
- `get_run_status`

## 6. Chart And Run Status Rendering

Purpose: show simple business charts and pipeline run cards when the selected model includes structured markers.

Main files:

- `apps/arcyria-platform/components/agents/AgentChart.tsx`
- `apps/arcyria-platform/components/agents/AgentRunStatusCard.tsx`
- `apps/arcyria-platform/components/agents/AgentMessageList.tsx`
- `packages/agent-sdk/src/index.tsx`
- `apps/arcyria-platform/public/agent.js`

Expected response marker:

```json
{
  "type": "bar",
  "labels": ["Jan", "Feb", "Mar"],
  "datasets": [
    { "label": "Revenue", "data": [12000, 18000, 22000] }
  ]
}
```

The model wraps the JSON in:

```text
<chart_data>{...}</chart_data>
```

The frontend strips the marker from visible text and renders the chart only when valid `chart_data` exists.

Run status cards use the same marker pattern:

```text
<run_status>{"status":"completed","runId":"...","startedAt":"...","completedAt":"..."}</run_status>
```
