# md-docs

Repository-level notes that are **broader than a single app**. Each guide has
one job; read them in order when onboarding.

## ELT product — start here

1. [`strict-elt-pipeline-guide.md`](./strict-elt-pipeline-guide.md) —
   authoritative step-by-step guide: `schema.table` contract, 5-phase flow,
   UI walkthrough, run-status drawer, 12 invariants, manual verification.
2. [`source-to-destination-elt-flow.md`](./source-to-destination-elt-flow.md)
   — deeper reference for the saved graph contract, normalisation rules,
   and cross-service payload shapes.
3. [`analytics-page-implementation.md`](./analytics-page-implementation.md) —
   repo-aligned implementation spec for the workspace Analytics page and its
   org-scoped analytics APIs.

The 12 non-negotiable invariants are enforced directly in
[`.cursor/rules/strict-elt-invariants.mdc`](../.cursor/rules/strict-elt-invariants.mdc),
and the ASCII flow diagram lives in
[`.cursor/rules/elt-flow-diagram.mdc`](../.cursor/rules/elt-flow-diagram.mdc).

## Billing

- [`billing-dodo-billingsdk.md`](./billing-dodo-billingsdk.md) — current billing architecture: first paid checkout, API-owned plan changes, Dodo portal for payment/receipts, local manual overage invoicing, and BillingSDK-style settings UI.
- [`dodo-payments-setup.md`](./dodo-payments-setup.md) — Dodo product IDs, checkout, webhooks, environment variables, and troubleshooting.
- [`dodo-payments-flowchart.md`](./dodo-payments-flowchart.md) — **Mermaid flow charts:** app → Go API → Dodo, API plan changes, portal, and webhook reconciliation.

## Observability (three separate guides)

Read in this order:

| Order | Doc | What you do there |
| --- | --- | --- |
| 1 | [`betterstack-setup.md`](./betterstack-setup.md) | Better Stack UI: monitors, status page, domain, API IDs |
| 2 | [`posthog-setup.md`](./posthog-setup.md) | PostHog UI: project key, error tracking, optional webhook |
| 3 | [`posthog-full-integration.md`](./posthog-full-integration.md) | Architecture, events catalog, quotas, flags, surveys, troubleshooting |
| 4 | [`observability-deployment.md`](./observability-deployment.md) | AWS SSM + Vercel env + ECS redeploy + smoke tests |

Legacy filenames redirect to the above (`betterstack-status-page-creation.md`, `posthog-betterstack-setup.md`).

## Deployment

- [`hetzner-server-setup.md`](./hetzner-server-setup.md) — Hetzner Cloud
  server creation guide for the CX33/CX43 single-server backend deployment.
- [`tigris-storage-setup.md`](./tigris-storage-setup.md) — Tigris setup for
  Terraform/OpenTofu state and Hetzner staging backups.
- [`deployment-oracle-cloud.md`](./deployment-oracle-cloud.md) — Complete
  Oracle Cloud two-VM Terraform, GitHub Actions, OCI Run Command, deployment,
  verification, rollback, and recovery guide.
- [`deployment-vercel.md`](./deployment-vercel.md) — Vercel frontend.
- [`aws-ses-setup.md`](./aws-ses-setup.md) — AWS SES email.

## Integrations

- [`slack-guide.md`](./slack-guide.md) — **Single** Slack guide: local ngrok
  setup, Slack Dashboard URLs, OAuth and Marketplace install, App Home,
  commands, events, copyable manifest, native builder behavior, review
  checklist, and troubleshooting.

## Agents

- [`agents/README.md`](./agents/README.md) - Custom Agent Builder docs: flow charts, different workflows, setup guide, and how the embedded data Q&A agent works.

## Testing

- [`testing-local.md`](./testing-local.md) — manual UI testing guide,
  connector-by-connector checklists, and final-target verification for the
  local ELT stack.
- [`manual-testing/slack-pipeline-e2e.md`](./manual-testing/slack-pipeline-e2e.md)
  — Slack-native pipeline creation test using Neon source and RDS Postgres
  destination, including strict destination table setup.
- [`manual-testing/postgres-to-postgres-data-types-ui.md`](./manual-testing/postgres-to-postgres-data-types-ui.md)
  — live PostgreSQL source to PostgreSQL destination UI test covering schema
  setup, wide data types, dbt SQL, full-table sync, and incremental upsert.

## Connectors

- [`saas-sources-group2.md`](./saas-sources-group2.md) — detailed SaaS
  source implementation guide (dlt verified sources: HubSpot, Stripe, …).
