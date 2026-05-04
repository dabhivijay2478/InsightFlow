# Agent Platform And Ask AI Summary

## Current UX Direction

The Agent Platform is the front door for pipeline-aware workers. Users first open `/workspace/agents` to browse saved agents, then use `/workspace/agents/new` to create a worker for an existing pipeline, or `/workspace/agents/:pipelineId/settings` to publish and configure it.

The Canvas Builder `Ask AI` panel remains focused on configuring the currently open pipeline. It should use the same guided mental model, but it should not feel like a separate product.

## Guided Flow

1. User asks to create or configure a pipeline worker.
2. Show existing source and destination connections.
3. If no connection exists, send the user to create one.
4. User selects source and destination connections, or the assistant guides creation through Q&A.
5. Assistant previews source schemas, tables, and sample rows.
6. User selects tables and confirms.
7. User selects destination schema and target table behavior.
8. Assistant asks for column mapping and transformation intent.
9. If the destination table does not exist, assistant asks for confirmation before creating it.
10. Assistant shows transformation/loading steps and progress.
11. After confirmation, redirect to `/workspace/data-pipelines/:id/builder`.
12. Canvas opens with source, destination, selected streams, SQL models, mappings, and automation context filled.

## Implementation Notes

- `/workspace/agents` is a clean saved-agent listing with a page header, search, status filter, sort control, and card grid.
- `/workspace/agents/new` is a centered ChatGPT-style guided pipeline setup without an outer card shell, secondary progress header, or top flow buttons. It starts with one large personalized welcome message, then the agent asks for connections, tables, mappings, and builder handoff through the chat itself.
- The new-agent chat reads the username from the Zustand auth store and opens with only: `Hi {name}, welcome. How can I help you?`
- The composer is fixed at the bottom of the chat surface. Navigation actions stay at the top level; there are no separate bottom setup buttons like "manual builder" or "use existing connections."
- Existing source/destination connections appear only after the user sends a request. The agent lists them without preselecting one, then asks the user to choose or create connections through the conversation.
- The guided chat does not show technical trace panels. Loading is limited to simple user-facing text such as "Checking available connections..." or "Preparing transforms..."
- If the first user message is only a greeting, the agent replies conversationally and asks what pipeline the user wants to create, without showing connections yet.
- Table preview and table confirmation are separate chat turns. The user names tables from the preview, the agent repeats the selected tables, and the flow only continues after confirmation.
- On final confirmation, the new-agent flow writes a mock `mantrixflow.agentPipelineDraft` handoff into `sessionStorage` and redirects to the builder with `agentConfigured=1&autoConfigure=1`.
- The builder consumes that handoff, fills mock source streams, destination SQL models, schedule context, switches to canvas mode, opens Ask AI, lays out the canvas, and shows a success toast.
- `/workspace/agents/:pipelineId` is the ChatGPT-style worker chat for running and querying a configured pipeline agent.
- `/workspace/agents/:pipelineId/settings` is the publisher/settings page for table access, allowed domains, embed snippets, and canvas handoff.
- The grid shows saved agents plus clearly labeled draft worker cards for pipelines that do not have a saved agent yet. Empty states send users to the separate creation flow.
- Filters support search, status, and sorting.
- The builder `Ask AI` panel has been restyled as a borderless guided setup surface.
- `components/ui/ai-prompt-box.tsx` is the reusable shadcn-compatible prompt input used by both agent and builder flows.
- `framer-motion`, `lucide-react`, Radix Dialog, and Radix Tooltip are available for the prompt component.

## Security And Product Boundary

- Agent Platform workers can be published and embedded.
- Public embedded agents cannot trigger pipeline runs.
- Authenticated workspace agents can trigger runs only when enabled.
- Source and destination queries remain scoped to explicit allowlists.
- Ask AI in the canvas only configures the current pipeline and applies proposed actions after user review.
