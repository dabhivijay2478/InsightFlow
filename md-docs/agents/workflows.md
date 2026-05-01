# Agent Workflows

## 1. Agent Builder Configuration

Purpose: create or update the agent for one existing pipeline.

Main files:

- `apps/app/app/workspace/data-pipelines/[id]/agent/page.tsx`
- `apps/app/lib/api/services/pipeline-agents.service.ts`
- `apps/server/main-server/internal/server/agent_http.go`
- `apps/server/main-server/internal/models/agent.go`

Route path:

- UI: `/workspace/data-pipelines/:id/agent`
- Go: `/api/v1/organizations/:organizationId/pipelines/:id/agent`

What happens:

1. The page loads pipeline details.
2. It discovers destination tables from the pipeline destination connection.
3. The user chooses allowed tables and allowed domains.
4. The Go API stores the config in `pipeline_agents`.
5. Go generates a public `agent_key` like `agent_abc123def456`.

## 2. Authenticated Test Chat

Purpose: let workspace users test the agent before embedding it.

Main files:

- `apps/app/app/workspace/data-pipelines/[id]/agent/page.tsx`
- `apps/app/app/api/pipelines/[id]/agent/chat/route.ts`
- `apps/server/main-server/internal/server/agent_http.go`

Route path:

- Next: `/api/pipelines/:id/agent/chat`
- Go query: `/api/v1/organizations/:organizationId/pipelines/:id/agent/query`

What happens:

1. The test panel sends AI SDK messages to Next.js.
2. Next.js loads the agent config using the user's JWT.
3. Claude receives a prompt with the agent config and an `execute_query` tool.
4. Tool calls go to Go, not directly to the database.
5. Go validates SQL and table allowlist before calling ELT.
6. The response streams back to the test panel.

## 3. Public Embedded Chat

Purpose: run the same agent from an external website.

Main files:

- `apps/app/public/agent.js`
- `packages/agent-sdk/src/index.tsx`
- `packages/agent-sdk/src/loader.tsx`
- `apps/app/app/api/agents/[agentKey]/chat/route.ts`
- `apps/app/app/api/agents/[agentKey]/info/route.ts`
- `apps/server/main-server/internal/server/agent_http.go`

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
7. Claude can only query through the guarded Go query endpoint.
8. Next saves the updated conversation through the internal route.

## 4. Read-Only Query Execution

Purpose: let Claude answer questions without giving it broad database access.

Main files:

- `apps/server/main-server/internal/server/agent_http.go`
- `apps/server/main-server/internal/server/agent_http_test.go`

Guard order:

1. Reject empty SQL.
2. Reject non-SELECT/non-WITH SQL.
3. Reject multiple top-level statements.
4. Extract referenced tables from `FROM` and `JOIN`.
5. Normalize table names to `schema.table`.
6. Reject any table outside `pipeline_agents.allowed_tables`.
7. Append or clamp `LIMIT 10000`.
8. Execute through ELT with `timeout_ms: 30000`.

## 5. Chart Response Rendering

Purpose: show simple business charts when Claude includes chart data.

Main files:

- `apps/app/app/workspace/data-pipelines/[id]/agent/page.tsx`
- `packages/agent-sdk/src/index.tsx`
- `apps/app/public/agent.js`

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

