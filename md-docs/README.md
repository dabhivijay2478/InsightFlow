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

## Integrations

- [`slack-oauth-setup.md`](./slack-oauth-setup.md) — Slack OAuth app setup:
  scopes, callback/proxy URLs, local tunnel setup, env vars, install flow, and
  troubleshooting.
- [`slack-marketplace-setup.md`](./slack-marketplace-setup.md) — Slack
  Marketplace direct install, required vs optional scopes, self-link mapping,
  and production Dashboard URLs.
- [`slack-marketplace-review-checklist.md`](./slack-marketplace-review-checklist.md)
  — review checklist for security, OAuth, listing assets, and final verification.
- [`slack-app-manifest.yaml`](./slack-app-manifest.yaml) — placeholder Slack
  app manifest for production/staging Dashboard setup.

## Agents

- [`agents/README.md`](./agents/README.md) - Custom Agent Builder docs: flow charts, different workflows, setup guide, and how the embedded data Q&A agent works.

## Testing

- [`testing-local.md`](./testing-local.md) — manual UI testing guide,
  connector-by-connector checklists, and final-target verification for the
  local ELT stack.

## Connectors

- [`saas-sources-group2.md`](./saas-sources-group2.md) — detailed SaaS
  source implementation guide (dlt verified sources: HubSpot, Stripe, …).

## Deployment

- [`deployment-vercel.md`](./deployment-vercel.md) — Vercel notes for the
  Next.js frontend.
- [`deployment-contabo-dokploy.md`](./deployment-contabo-dokploy.md) — VPS
  deployment on Contabo using Dokploy + Docker Compose.
