# Agent Platform And Ask AI Summary

## Current UX Direction

The Agent Platform is the front door for pipeline-aware workers. Users first open `/workspace/agents` to browse saved agents, then use `/workspace/agents/new` to create a worker for an existing pipeline, or `/workspace/agents/:pipelineId?tab=publish` to publish and configure it.

The Canvas Builder `Ask AI` panel remains focused on configuring the currently open pipeline. It should use the same guided mental model, but it should not feel like a separate product.

## Guided Flow

1. User asks to create or configure a pipeline worker.
2. `/workspace/agents/new` shows only real existing pipelines from the workspace.
3. If no pipeline exists, send the user to create a pipeline first.
4. User selects the pipeline that the worker will operate.
5. The flow shows real pipeline context: source, destination, selected streams, SQL models, and publish status.
6. User opens publisher settings for that pipeline.
7. User configures agent name, description, personality, source/destination allowlists, run permissions, public source-query setting, and allowed domains.
8. The Go API persists the worker in `pipeline_agents` and returns a public `agent_key`.
9. Workspace users test the worker at `/workspace/agents/:pipelineId`.
10. Embedded public users use the SDK/widget, which calls the public Next route and then guarded Go internal endpoints.

Pipeline creation is separate: Canvas Builder `Ask AI` and `/workspace/data-pipelines/new` create/configure pipelines. Agent Platform publishes workers for pipelines that already exist.

## Implementation Notes

- `/workspace/agents` is a clean saved-agent listing with a page header, search, status filter, sort control, and card grid.
- The grid lists only persisted `pipeline_agents`. It does not create fake draft agent cards from pipelines.
- `/workspace/agents/new` is a centered ChatGPT-style worker setup screen. It starts with one large personalized welcome message, then guides the user to select a real existing pipeline.
- The new-agent chat reads the username from the Zustand auth store and opens with only: `Hi {name}, welcome. How can I help you?`
- The composer is fixed at the bottom of the chat surface. Navigation actions stay at the top level; there are no separate bottom setup buttons like "manual builder" or "use existing connections."
- Pipeline cards use actual API fields: source connection, destination connection, selected streams, delivery targets, status, and updated date.
- The guided chat does not show technical trace panels. Loading is limited to simple user-facing text such as "Checking available connections..." or "Preparing transforms..."
- If the first user message is only a greeting, the screen replies conversationally and asks which existing pipeline should get a worker.
- New-agent setup does not write mock builder handoff state and does not create fake connections, schemas, tables, mappings, or pipeline drafts.
- After pipeline selection, the setup routes to `/workspace/agents/:pipelineId?tab=publish` for real publisher configuration.
- `/workspace/agents/:pipelineId` is the ChatGPT-style worker chat for running and querying a configured pipeline agent.
- `/workspace/agents/:pipelineId?tab=publish` is the publisher/settings view for table access, allowed domains, embed snippets, and canvas handoff. The old settings path redirects there.
- Empty states send users to the separate creation flow, or to pipeline creation if no pipelines exist.
- Filters support search, status, and sorting.
- The builder `Ask AI` panel has been restyled as a borderless guided setup surface.
- `components/ui/ai-prompt-box.tsx` is the reusable shadcn-compatible prompt input used by both agent and builder flows.
- `framer-motion`, `lucide-react`, Radix Dialog, and Radix Tooltip are available for the prompt component.
- Data Q&A prompts instruct the model to query live data first and never invent numbers or columns. If columns are unknown, the model must inspect the allowlisted table with `SELECT * ... LIMIT 5` before using column-specific SQL.

## Security And Product Boundary

- Agent Platform workers can be published and embedded.
- Public embedded agents cannot trigger pipeline runs.
- Authenticated workspace agents can trigger runs only when enabled.
- Source and destination queries remain scoped to explicit allowlists.
- Ask AI in the canvas only configures the current pipeline and applies proposed actions after user review.
