# Agent Setup Guide

## Prerequisites

- Go main server running with database access.
- Next.js app running with authenticated workspace sessions.
- ELT server reachable from Go.
- Anthropic key available to the Next.js server process.
- At least one pipeline with a configured destination connection.

## Environment Variables

Set these for the Next.js app:

```bash
NEXT_PUBLIC_API_URL=http://localhost:5000
ANTHROPIC_API_KEY=...
ANTHROPIC_MODEL=claude-sonnet-4-20250514
INTERNAL_TOKEN=shared-local-internal-token
```

`ANTHROPIC_MODEL` is optional. If unset, the code defaults to `claude-sonnet-4-20250514`.

Set the same internal token for the Go main server:

```bash
INTERNAL_TOKEN=shared-local-internal-token
```

The public embed chat is served by Next.js. Next.js calls Go internal routes with:

```http
X-Internal-Token: shared-local-internal-token
```

Do not expose `INTERNAL_TOKEN` or destination connection credentials to browser code.

## Database Setup

The Go AutoMigrate now includes:

- `pipeline_agents`
- `pipeline_agent_conversations`

The RLS script includes org-scoped policies for `pipeline_agents` and keeps `pipeline_agent_conversations` closed to direct client access.

Files:

- `apps/server/main-server/internal/models/agent.go`
- `apps/server/main-server/internal/database/migrate.go`
- `apps/server/main-server/sql/supabase_rls.sql`

## Local Startup

Start the backend services using the repo's normal local commands.

Typical process:

1. Start the ELT server.
2. Start the Go main server.
3. Start the Next.js app.
4. Sign in to the app.
5. Open an existing pipeline.
6. Go to `/workspace/data-pipelines/:pipelineId/agent`.

## Create an Agent

1. Enter an agent name.
2. Add a description shown to embedded users.
3. Select allowed destination tables.
4. Add allowed domains, one per line or comma-separated.
5. Save the agent.
6. Copy the script or React snippet.

Allowed domains must match the browser `Origin` exactly after normalization:

```text
https://dashboard.example.com
http://localhost:3000
```

Paths are ignored, but scheme and host must match.

## Embed Snippets

Vanilla script:

```html
<script>
(function(w,d,s,o,f,js,fjs){w['MantrixAgent']=o;w[o]=w[o]||function(){(w[o].q=w[o].q||[]).push(arguments)};js=d.createElement(s);fjs=d.getElementsByTagName(s)[0];js.id=o;js.src=f;js.async=1;fjs.parentNode.insertBefore(js,fjs);}(window,document,'script','mantrix','https://YOUR_APP_ORIGIN/agent.js'));
mantrix('init', 'agent_abc123def456');
</script>
```

React:

```tsx
import { MantrixAgent } from "@mantrixflow/agent-sdk";

export function App() {
  return <MantrixAgent agentId="agent_abc123def456" apiBaseUrl="https://YOUR_APP_ORIGIN" />;
}
```

## Smoke Tests

Backend:

```bash
cd apps/server/main-server
GOCACHE=/tmp/mxf-go-build-cache go test ./internal/server ./internal/database ./internal/models
```

Frontend touched-file lint:

```bash
cd apps/app
bun run biome check public/agent.js 'app/api/agents/[agentKey]/chat/route.ts' 'app/api/agents/[agentKey]/info/route.ts' 'app/api/pipelines/[id]/agent/chat/route.ts' 'app/workspace/data-pipelines/[id]/agent/page.tsx'
```

Manual smoke:

1. Save an agent with `http://localhost:3000` in allowed domains.
2. Ask a question in the builder test panel.
3. Embed the script in a local page served from an allowed origin.
4. Ask a question from the widget.
5. Confirm disallowed origins receive a forbidden response.
6. Confirm a query against a non-allowlisted table is rejected.

## Common Failure Modes

- `Origin is not allowed for this agent`: add the exact scheme and host to allowed domains.
- `No tables are allowlisted for this agent`: select at least one destination table and save.
- `Only SELECT queries are allowed`: the generated SQL included DML/DDL or multiple statements.
- `INTERNAL_TOKEN must be set`: Next.js cannot call Go internal routes.
- Empty or generic answers: destination schemas may not be discoverable, or the allowed table list is too narrow.

