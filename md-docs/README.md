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

- [`billing-dodo-billingsdk.md`](./billing-dodo-billingsdk.md) — current billing architecture: Dodo hosted checkout for paid purchases/upgrades, scheduled downgrades, Dodo portal for payment/receipts, and BillingSDK-style settings UI.
- [`dodo-payments-setup.md`](./dodo-payments-setup.md) — Dodo product IDs, checkout, webhooks, environment variables, and troubleshooting.
- [`dodo-payments-flowchart.md`](./dodo-payments-flowchart.md) — **Mermaid flow charts:** app → Go API → Dodo hosted checkout, portal, invoices, and webhook reconciliation.

## Email

- [`autosend-email-system-plan.md`](./autosend-email-system-plan.md) - AutoSend architecture, Supabase SMTP setup, backend API setup, Dodo integration setup, and duplicate-prevention rules.
- [`autosend-email-catalog.md`](./autosend-email-catalog.md) - Owner-by-owner email catalog for Supabase Auth, Go backend, and Dodo Payments.
- [`autosend-template-copy.md`](./autosend-template-copy.md) - Subject lines, preview text, CTA labels, and body copy for every AutoSend template.
- [`autosend-template-design-guide.md`](./autosend-template-design-guide.md) - Responsive MantrixFlow email layout rules and base shell.
- [`autosend-template-id-map.md`](./autosend-template-id-map.md) - AutoSend-created template IDs and env mapping.
- [`dodo-autosend-transformations.md`](./dodo-autosend-transformations.md) - Dodo to AutoSend JavaScript transformation examples.
- [`autosend-dodo-supabase-production-runbook.md`](./autosend-dodo-supabase-production-runbook.md) - Production setup, Dodo integration steps, Supabase SMTP timeout troubleshooting, and rollout checklist.
- [`autosend-production-deployment-guide.md`](./autosend-production-deployment-guide.md) - Production deployment checklist and exact env/dashboard values for AutoSend, Supabase, backend, Dodo, and frontend.
- [`supabase-auth-email-templates.md`](./supabase-auth-email-templates.md) - Paste-ready Supabase Auth HTML templates for confirmation, invite, magic link, email change, password reset, and reauthentication.

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
- [`oria-adk-to-ai-sdk-migration.md`](./oria-adk-to-ai-sdk-migration.md) - ADK → Vercel AI SDK migration report (architecture, removed files, test status).
- [`oria-agent-setup.md`](./oria-agent-setup.md) - Private Oria runtime setup: Next.js OpenRouter + Go tools/persistence, env vars, verification.
- [`oria-agent-testing-guide.md`](./oria-agent-testing-guide.md) - Master index for six release test corpora (~5,100 prompts, all 73 agents).
- [`oria-test-prompts-release1-read.md`](./oria-test-prompts-release1-read.md) - Release 1 read specialists + Oria root (850 prompts).
- [`oria-test-prompts-release2-action.md`](./oria-test-prompts-release2-action.md) - Release 2 action specialists: pipeline build, connections, transforms, runs (881 prompts).
- [`oria-test-prompts-release3-automation.md`](./oria-test-prompts-release3-automation.md) - Release 3 automation specialists (852 prompts).
- [`oria-test-prompts-release4-intelligence.md`](./oria-test-prompts-release4-intelligence.md) - Release 4 intelligence specialists (852 prompts).
- [`oria-test-prompts-release5-enterprise.md`](./oria-test-prompts-release5-enterprise.md) - Release 5 enterprise specialists (861 prompts).
- [`oria-test-prompts-release6-platform.md`](./oria-test-prompts-release6-platform.md) - Release 6 platform specialists (852 prompts).
- [`../apps/mantrixflow-docs/user-guide/oria-copilot.mdx`](../apps/mantrixflow-docs/user-guide/oria-copilot.mdx) - Public Oria Copilot guide covering navigation, context, example questions, history, safety boundaries, and troubleshooting without exposing internal capability names.
- [`ai-copilot-phase-1.md`](./ai-copilot-phase-1.md) - Workspace Copilot Release 1 architecture, provider setup, 12-agent/tool registries, redaction, persistence, and testing.

## Testing

- [`frontend-refactor-audit.md`](./frontend-refactor-audit.md) — complete
  frontend size/duplication audit, target architecture, safety baseline, and
  incremental feature-by-feature refactoring plan.
- [`pipeline-all-streams-chrome-e2e-report-2026-07-19.md`](./pipeline-all-streams-chrome-e2e-report-2026-07-19.md)
  — final live-Chrome all-streams PostgreSQL, HubSpot, and Stripe pipeline report,
  including Neon cleanup, GitHub/YAML round trip, fixes, and destination evidence.
- [`pipeline-new-layout-e2e-report-2026-07-18.md`](./pipeline-new-layout-e2e-report-2026-07-18.md)
  — previous API-focused report with historical Chrome/GitHub blockers.
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

- [`../apps/mantrixflow-docs/connections/sources/productivity/airtable.mdx`](../apps/mantrixflow-docs/connections/sources/productivity/airtable.mdx) — public Airtable source guide: Personal Access Token scopes, base/table discovery, stable staging names, Full Table behavior, and troubleshooting.
- [`../apps/mantrixflow-docs/connections/destinations/airtable.mdx`](../apps/mantrixflow-docs/connections/destinations/airtable.mdx) — public Airtable destination guide: existing-table contract, writable field mapping, merge keys, Upsert behavior, verification, and errors.
- [`../apps/mantrixflow-docs/connections/sources/database/mysql.mdx`](../apps/mantrixflow-docs/connections/sources/database/mysql.mdx) — public MySQL source guide: least-privilege grants, discovery, Full Table and Incremental modes, types, and verification.
- [`../apps/mantrixflow-docs/connections/destinations/mysql.mdx`](../apps/mantrixflow-docs/connections/destinations/mysql.mdx) — public MySQL destination guide: destination DDL, writer grants, Upsert contract, verification queries, and troubleshooting.
- [`../apps/mantrixflow-docs/example/pipelines/airtable-and-mysql.mdx`](../apps/mantrixflow-docs/example/pipelines/airtable-and-mysql.mdx) — end-to-end Airtable → MySQL, MySQL → Airtable, and Airtable → Airtable setup and run guide.
- [`../docs/airtable-source-destination-connector-audit.md`](../docs/airtable-source-destination-connector-audit.md) — implementation and live-Chrome verification audit for the Airtable bidirectional connector and seven-route matrix.
- [`saas-sources-group2.md`](./saas-sources-group2.md) — detailed SaaS
  source implementation guide (dlt verified sources: HubSpot, Stripe, …).
- [`stripe-connector-ai-prompts.md`](./stripe-connector-ai-prompts.md) —
  AI prompt/reference pack for designing or extending the Stripe source
  connector.
- [`hubspot-postgres-full-phase-plan.md`](./hubspot-postgres-full-phase-plan.md) —
  implemented production architecture for the HubSpot source to existing PostgreSQL
  destination flow across registration, discovery, preflight, dlt/DuckDB,
  UI SQL/dbt, delivery, checkpoints, AI safety, testing, and release gates.
- [`salesforce-postgres-setup.md`](./salesforce-postgres-setup.md) —
  Salesforce source to PostgreSQL destination setup guide, including OAuth,
  dynamic discovery, Bulk API, CDC streamer, and destination table contract.
- [`salesforce-postgres-testing-guide.md`](./salesforce-postgres-testing-guide.md)
  — Salesforce to PostgreSQL automated and manual testing guide.
