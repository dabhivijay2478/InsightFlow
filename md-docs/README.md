# MantrixFlow documentation

This directory is the single home for maintained engineering guides, runbooks,
architecture notes, audits, and test reports. Documentation is grouped by
subject so service repositories do not grow their own competing `docs/` trees.

## Directory map

```text
md-docs/
├── architecture/     System, ELT, frontend, and simulation design
├── infrastructure/   OVH, Terraform, networking, storage, CI/CD, and database plans
├── deployment/       Current service runbooks plus isolated legacy guides
├── operations/       Observability and production operations
├── integrations/     Billing, connectors, email, and provider setup
├── ai/               Oria and Agent Grid implementation guides
├── audits/           Connector, frontend, product, and legal reviews
├── testing/          Local test guides and test reports
├── manual-testing/   Detailed connector and pipeline verification matrices
├── agents/           Custom Agent Builder documentation
├── github/           GitHub integration contracts and plans
└── content-audit/     Documentation inventory and generated audit evidence
```

Component `README.md` files remain beside their code as short entry points.
Generated Swagger files remain in `apps/server/arcyria-server/docs/` because that
directory is a generated Go package imported by the server. Customer-facing
MDX remains in `apps/arcyria-docs/`, and code-owned templates, migrations,
skills, and test fixtures remain beside their consumers.

## Start here

| Subject | Primary document |
| --- | --- |
| New OVH architecture setup | [OVH, Dokploy, and Microsandbox setup](./infrastructure/setup/ovh-dokploy-microsandbox.md) |
| Production deployment | [OVH and Microsandbox runbook](./deployment/infrastructure/ovh-microsandbox-runbook.md) |
| GitHub environments | [GitHub Actions environment matrix](./infrastructure/ci-cd/github-actions.md) |
| Production architecture | [Production architecture](./infrastructure/architecture/production-architecture.md) |
| Simulation platform | [Microsandbox platform](./architecture/simulation/platform.md) |
| Strict ELT | [Strict ELT pipeline guide](./architecture/elt/strict-pipeline-guide.md) |
| Local verification | [Local development and testing](./testing/local-development.md) |

## Infrastructure and deployment

- [OVH, Dokploy, and Microsandbox setup](./infrastructure/setup/ovh-dokploy-microsandbox.md)
- [OVH and Microsandbox deployment runbook](./deployment/infrastructure/ovh-microsandbox-runbook.md)
- [Go API and simulation-manager deployment](./deployment/services/go-api-simulation-manager.md)
- [Python ELT and simulation-runtime deployment](./deployment/services/python-elt-simulation-runtime.md)
- [GitHub Actions environments](./infrastructure/ci-cd/github-actions.md)
- [Deployment independence](./infrastructure/ci-cd/deployment-independence.md)
- [Private networking](./infrastructure/networking/private-network.md)
- [Dokploy operations](./infrastructure/operations/dokploy.md)
- [Backup and restore](./infrastructure/operations/backup-restore.md)
- [Tigris storage](./infrastructure/storage/tigris.md)
- [Supabase transition plan](./infrastructure/database/supabase-transition.md)
- [Supabase RLS guide](./infrastructure/database/supabase-rls-guide.md)

Frontend deployment remains separate under
[deployment/frontend](./deployment/frontend/). Superseded provider material is
isolated under [deployment/legacy](./deployment/legacy/); it is not the current
backend architecture.

## Architecture

- [Strict ELT pipeline guide](./architecture/elt/strict-pipeline-guide.md)
- [Source-to-destination ELT flow](./architecture/elt/source-to-destination-flow.md)
- [Simulation platform](./architecture/simulation/platform.md)
- [Simulation implementation report](./architecture/simulation/implementation-report.md)
- [Analytics page implementation](./architecture/frontend/analytics-page-implementation.md)

The non-negotiable ELT invariants remain executable repository rules in
[strict-elt-invariants.mdc](../.cursor/rules/strict-elt-invariants.mdc) and
[elt-flow-diagram.mdc](../.cursor/rules/elt-flow-diagram.mdc).

## Operations

- [Better Stack setup](./operations/observability/betterstack-setup.md)
- [PostHog setup](./operations/observability/posthog-setup.md)
- [PostHog integration reference](./operations/observability/posthog-full-integration.md)
- [Observability deployment](./operations/observability/deployment.md)

## Integrations

- [Billing documentation](./integrations/billing/)
- [Connector guides and plans](./integrations/connectors/)
- [Email and AutoSend documentation](./integrations/email/)
- [GitHub integration documentation](./github/)

Implementation and release reviews for connectors are kept separately under
[audits/connectors](./audits/connectors/) so setup guides are not confused with
point-in-time audit evidence.

## AI and agents

- [Oria documentation](./ai/oria/)
- [Agent Grid documentation](./ai/agent-grid/)
- [Custom Agent Builder index](./agents/README.md)

## Audits and reports

- [Connector audits](./audits/connectors/)
- [Frontend audits](./audits/frontend/)
- [Product audits](./audits/product/)
- [Legal reviews](./audits/legal/)
- [Infrastructure implementation reports](./infrastructure/reports/)
- [Test reports](./testing/reports/)
- [Content audit](./content-audit/)

## Testing

- [Local development and testing](./testing/local-development.md)
- [General test cases](./testing/test-cases.md)
- [Frontend testing documentation](./testing/frontend/README.md)
- [Manual testing library](./manual-testing/README.md)
- [Pipeline E2E reports](./testing/reports/)

## Documentation rules

1. Put new standalone engineering documentation in the closest category here.
2. Keep a component `README.md` short and link to the canonical guide here.
3. Do not create another repository-level `docs/` directory.
4. Keep public customer documentation in `apps/arcyria-docs/`.
5. Keep generated documentation and code-owned Markdown beside the generator or
   runtime that consumes it.
6. Use kebab-case filenames and update this index when adding a new category.
