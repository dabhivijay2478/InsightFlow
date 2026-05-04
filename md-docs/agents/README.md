# Mantrixflow Agents Docs

This folder documents the standalone Agent Platform implementation: setup, runtime flows, security boundaries, and the different ways a pipeline-aware worker agent is used.

## Documents

- [flow-charts.md](./flow-charts.md) - architecture and sequence diagrams for agent publishing, authenticated worker chat, public embed chat, run tools, and query execution.
- [workflows.md](./workflows.md) - the different agent workflows and which routes/components participate in each one.
- [setup-guide.md](./setup-guide.md) - local setup, required environment variables, migration notes, and smoke tests.
- [model-providers.md](./model-providers.md) - switch AI runtime between Anthropic, Vercel AI Gateway, OpenRouter, Ollama, or another OpenAI-compatible endpoint.
- [how-it-works.md](./how-it-works.md) - practical explanation of how requests move through Next.js, Go, ELT, and the destination database.

## Core Invariant

An embedded agent is public by `agent_key`, but it is not trusted by itself. Security comes from:

- `allowed_domains` checked against the browser `Origin`.
- destination `allowed_tables` and source `allowed_source_tables` checked against every referenced SQL table.
- read-only SQL validation before execution.
- source and destination credentials staying only inside Go and ELT server calls.
- public chat never receiving pipeline-run tools.
- public chat going through Next.js with a Go internal token, never direct browser-to-internal-Go calls.
