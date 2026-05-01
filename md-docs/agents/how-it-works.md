# How Agents Work

## Short Version

A Mantrixflow Custom Agent is a Claude-powered data Q&A assistant tied to one pipeline. It knows only the tables selected in the Agent Builder. When a user asks a question, Claude may generate SQL, but Go validates and executes that SQL through the destination connection in a guarded read-only path.

## Main Pieces

| Piece | Responsibility |
| --- | --- |
| Agent Builder Page | Configure name, description, personality, allowed tables, allowed domains, and test chat. |
| Next.js AI Routes | Own Claude streaming and tool orchestration with the Vercel AI SDK. |
| Go API | Own tenancy, persistence, credentials, SQL validation, allowlists, origin checks, rate limits, and ELT calls. |
| ELT Server | Executes discover/query requests using decrypted destination credentials provided by Go. |
| SDK Widget | Renders the embedded chat UI and stores a browser session id. |

## Data Stored

`pipeline_agents` stores the deployable agent config:

- `pipeline_id`
- `org_id`
- `name`
- `description`
- `personality`
- `allowed_tables`
- `allowed_domains`
- `agent_key`
- `is_active`

`pipeline_agent_conversations` stores public browser session state:

- `agent_id`
- `session_id`
- `messages`
- `rate_window_start`
- `rate_window_count`

Conversation rows are not directly readable by normal clients.

## Why Next.js Owns Claude Calls

The implementation keeps AI orchestration in Next.js because the app already uses the Vercel AI SDK. That gives the UI streaming behavior without pushing AI-specific streaming code into Go.

Next.js does not decrypt destination credentials and does not execute SQL. It only:

1. Builds the prompt.
2. Defines the `execute_query` tool.
3. Streams Claude output.
4. Calls Go when a tool needs data.

## Why Go Owns Data Access

Go already owns org tenancy, connection credential handling, ELT proxying, and persistence. For agents, Go is the trust boundary.

Go checks:

- Is this pipeline in the organization?
- Does this agent exist and belong to the pipeline?
- Is the public agent active?
- Is the browser `Origin` allowlisted?
- Is the session under 100 requests/hour?
- Is the SQL read-only?
- Are all referenced tables in `allowed_tables`?
- Does the query have a row limit no higher than 10,000?

Only after those checks does Go decrypt destination credentials and call ELT.

## Public Embed Security Model

The `agent_key` is public. It is safe to put in HTML because it only identifies the agent.

The real controls are:

- `allowed_domains`: rejects unknown browser origins.
- `allowed_tables`: rejects SQL outside the selected table set.
- `X-Internal-Token`: lets Next.js call Go internal endpoints without exposing those endpoints directly to browsers.
- read-only SQL validation: rejects mutating SQL before ELT execution.
- rate limiting: limits each browser session to 100 requests per hour.

## Query Example

Visitor asks:

```text
What was revenue last week?
```

Claude may call:

```sql
SELECT
  SUM(total) AS revenue,
  COUNT(*) AS orders
FROM analytics.order_history
WHERE created_at >= NOW() - INTERVAL '7 days'
```

Go normalizes referenced tables:

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

Claude then formats the result:

```text
Revenue last week was $45,231 across 234 orders.
```

## Chart Responses

If a chart is useful, Claude appends:

```text
<chart_data>{"type":"bar","labels":["Jan","Feb"],"datasets":[{"label":"Revenue","data":[45000,62000]}]}</chart_data>
```

The UI:

1. Extracts the JSON.
2. Hides the raw marker from visible chat text.
3. Renders a chart only when the data is valid.

The builder page uses Recharts. The static vanilla widget renders a lightweight inline bar chart to avoid adding host-page dependencies.

## Agent vs AI Pipeline Builder

The Custom Agent Builder is for asking questions about data after a pipeline exists.

The AI Pipeline Builder is for creating a pipeline through chat before the pipeline exists. It also uses Next.js AI orchestration, but its tools create source schemas, destination schemas, and the pipeline graph instead of executing destination Q&A queries.

