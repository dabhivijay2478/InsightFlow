# How Agents Work

## Short Version

A Mantrixflow Custom Agent is a model-powered pipeline worker tied to one pipeline. It can answer questions from explicitly allowlisted source/destination tables, and authenticated workspace users can also ask it to trigger pipeline runs or inspect run status. When the selected model generates SQL, Go validates and executes that SQL through the scoped connection in a guarded read-only path.

## Main Pieces

| Piece | Responsibility |
| --- | --- |
| Agent Platform Page | Configure publisher settings, source/destination table allowlists, run permissions, allowed domains, and test chat. |
| Next.js AI Routes | Own AI streaming and tool orchestration with the Vercel AI SDK. |
| Go API | Own tenancy, persistence, credentials, SQL validation, allowlists, origin checks, rate limits, and ELT calls. |
| ELT Server | Executes discover/query requests using decrypted source or destination credentials provided by Go. |
| SDK Widget | Renders the embedded chat UI and stores a browser session id. |

## Data Stored

`pipeline_agents` stores the deployable agent config:

- `pipeline_id`
- `org_id`
- `name`
- `description`
- `personality`
- `allowed_tables` for destination queries
- `allowed_source_tables` for source queries
- `allowed_domains`
- `can_trigger_pipeline_runs`
- `allow_public_source_queries`
- `agent_key`
- `is_active`

`pipeline_agent_conversations` stores public browser session state:

- `agent_id`
- `session_id`
- `messages`
- `rate_window_start`
- `rate_window_count`

Conversation rows are not directly readable by normal clients.

## Why Next.js Owns AI Calls

The implementation keeps AI orchestration in Next.js because the app already uses the Vercel AI SDK. That gives the UI streaming behavior without pushing AI-specific streaming code into Go.

The model provider is selected by `apps/app/lib/ai/model-provider.ts`, so the same routes can use Anthropic, Vercel AI Gateway, OpenRouter, Ollama, or another OpenAI-compatible endpoint. Next.js does not decrypt destination credentials and does not execute SQL. It only:

1. Builds the prompt.
2. Defines tools such as `execute_query`, `run_pipeline`, and `get_run_status`.
3. Streams AI output.
4. Calls Go when a tool needs data.

For local demos, set `AI_MODEL_PROVIDER=ollama` and a tool-call-capable local model such as `qwen3.5:4b`. The security behavior remains the same because all data access still goes through Go.

## Why Go Owns Data Access

Go already owns org tenancy, connection credential handling, ELT proxying, and persistence. For agents, Go is the trust boundary.

Go checks:

- Is this pipeline in the organization?
- Does this agent exist and belong to the pipeline?
- Is the public agent active?
- Is the browser `Origin` allowlisted?
- Is the session under 100 requests/hour?
- Is the SQL read-only?
- Are all referenced tables in the scope's allowlist?
- Is public source querying explicitly enabled before serving source-scope embedded requests?
- Does the query have a row limit no higher than 10,000?

Only after those checks does Go decrypt the matching source or destination credentials and call ELT.

## Public Embed Security Model

The `agent_key` is public. It is safe to put in HTML because it only identifies the agent.

The real controls are:

- `allowed_domains`: rejects unknown browser origins.
- `allowed_tables`: rejects destination SQL outside the selected table set.
- `allowed_source_tables`: rejects source SQL outside the selected table set.
- `allow_public_source_queries`: keeps embedded source access off unless explicitly enabled.
- `X-Internal-Token`: lets Next.js call Go internal endpoints without exposing those endpoints directly to browsers.
- read-only SQL validation: rejects mutating SQL before ELT execution.
- rate limiting: limits each browser session to 100 requests per hour.
- no public run tools: embeds cannot trigger pipeline runs or inspect private run APIs.

## Query Example

Visitor asks:

```text
What was revenue last week?
```

The selected model may call:

```sql
SELECT
  SUM(total) AS revenue,
  COUNT(*) AS orders
FROM analytics.order_history
WHERE created_at >= NOW() - INTERVAL '7 days'
```

Go normalizes referenced tables and the requested scope:

```text
analytics.order_history
```

If `analytics.order_history` is in `allowed_tables`, Go sends a limited query to ELT:

```sql
SELECT * FROM (
  SELECT
    SUM(total) AS revenue,
    COUNT(*) AS orders
  FROM analytics.order_history
  WHERE created_at >= NOW() - INTERVAL '7 days'
) AS __mxf_agent_query LIMIT 10000
```

The selected model then formats the result:

```text
Revenue last week was $45,231 across 234 orders.
```

## Authenticated Pipeline Worker Example

Workspace user asks:

```text
Run the pipeline and tell me whether the latest order total changed.
```

The authenticated agent route may call:

```text
run_pipeline()
get_run_status(runId)
execute_query({ "scope": "destination", "sql": "SELECT SUM(total) FROM analytics.order_history" })
```

The public embedded route never receives `run_pipeline` or `get_run_status`, even if the agent config allows authenticated users to trigger runs.

## Chart Responses

If a chart is useful, the selected model appends:

```text
<chart_data>{"type":"bar","labels":["Jan","Feb"],"datasets":[{"label":"Revenue","data":[45000,62000]}]}</chart_data>
```

The UI:

1. Extracts the JSON.
2. Hides the raw marker from visible chat text.
3. Renders a chart only when the data is valid.

The Agent Platform uses Recharts. The static vanilla widget renders a lightweight inline bar chart to avoid adding host-page dependencies.

## Agent vs AI Pipeline Builder

The Agent Platform is for publishing and operating pipeline-aware workers after a pipeline exists.

The AI Pipeline Builder is for creating a pipeline through chat before the pipeline exists. It also uses Next.js AI orchestration, but its tools create source schemas, destination schemas, and the pipeline graph instead of operating published pipeline workers.
