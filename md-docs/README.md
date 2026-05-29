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

- [`dodo-payments-setup.md`](./dodo-payments-setup.md) — **Single** Dodo guide: products, first-time **checkout**, portal-only plan changes (**change-plan API not used**), collection **fallback** (`DODO_PRODUCT_COLLECTION_ID`), webhooks, troubleshooting ([Dodo Product Collections](https://docs.dodopayments.com/features/product-collections)).
- [`dodo-payments-flowchart.md`](./dodo-payments-flowchart.md) — **Mermaid flow charts:** app → Go API → Dodo (checkout vs portal) and webhook reconciliation.

## Observability

- [`betterstack-status-page-creation.md`](./betterstack-status-page-creation.md) — **Start here.** Better Stack monitors + status page for **AWS ECS** production (`cloud.mantrixflow.com`, `cloud.api.mantrixflow.com/status`), SSM secrets, Vercel PostHog, custom domain `status.mantrixflow.com`.
- [`posthog-betterstack-setup.md`](./posthog-betterstack-setup.md) — PostHog + Better Stack code paths, health endpoints, webhooks, widgets, and **AWS SSM + Vercel** deployment.

## Deployment (AWS production)

- [`../apps/mantrixflow-infra/DEPLOYMENT.md`](../apps/mantrixflow-infra/DEPLOYMENT.md) — **Start here.** ECS Fargate, ALB, SSM, GitHub Actions, smoke checks.
- [`deployment-vercel.md`](./deployment-vercel.md) — Next.js frontend on Vercel (`cloud.mantrixflow.com`).
- [`aws-ses-setup.md`](./aws-ses-setup.md) — AWS SES for Go API product email and Supabase Auth SMTP (DNS, DKIM, production access).

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

## Connectors

- [`saas-sources-group2.md`](./saas-sources-group2.md) — detailed SaaS
  source implementation guide (dlt verified sources: HubSpot, Stripe, …).

