# Mantrixflow Agents Docs

This folder documents the Custom Agent Builder implementation: setup, runtime flows, security boundaries, and the different ways an agent is used.

## Documents

- [flow-charts.md](./flow-charts.md) - architecture and sequence diagrams for agent creation, test chat, public embed chat, and query execution.
- [workflows.md](./workflows.md) - the different agent workflows and which routes/components participate in each one.
- [setup-guide.md](./setup-guide.md) - local setup, required environment variables, migration notes, and smoke tests.
- [how-it-works.md](./how-it-works.md) - practical explanation of how requests move through Next.js, Go, ELT, and the destination database.

## Core Invariant

An embedded agent is public by `agent_key`, but it is not trusted by itself. Security comes from:

- `allowed_domains` checked against the browser `Origin`.
- `allowed_tables` checked against every referenced SQL table.
- read-only SQL validation before execution.
- destination credentials staying only inside Go and ELT server calls.
- public chat going through Next.js with a Go internal token, never direct browser-to-internal-Go calls.

