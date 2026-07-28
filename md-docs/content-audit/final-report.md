# MantrixFlow content audit and controlled cleanup — final report

Date: 2026-07-28
Scope: marketing website, public documentation, web application, and repository content used to verify claims

## Outcome

The audited customer-facing content is shorter, more consistent, and aligned with the implemented product. Public URLs, application routes, API contracts, plan values, permission checks, legal pages, and product workflows were preserved. No indexed public page was deleted.

The complete evidence set is available in this directory:

- [`audit-report.md`](./audit-report.md) — initial inventory, classification, conflicts, and proposed actions
- [`generated/content-inventory.csv`](./generated/content-inventory.csv) — baseline item-by-item inventory
- [`generated/after-content-inventory.csv`](./generated/after-content-inventory.csv) — post-cleanup item-by-item inventory
- [`generated/duplicate-content.csv`](./generated/duplicate-content.csv) and [`generated/after-duplicate-content.csv`](./generated/after-duplicate-content.csv)
- [`generated/link-inventory.csv`](./generated/link-inventory.csv) and [`generated/after-link-inventory.csv`](./generated/after-link-inventory.csv)
- [`generated/baseline-manifest.json`](./generated/baseline-manifest.json) and [`generated/after-baseline-manifest.json`](./generated/after-baseline-manifest.json)
- [`source-of-truth.md`](./source-of-truth.md) — final terminology, availability, connector, plan, and content-placement rules
- [`../../scripts/content-audit.mjs`](../../scripts/content-audit.mjs) — repeatable inventory generator

The inventory covers routes, layouts, content-bearing components, dialog/sheet and tooltip hosts, public articles, and internal reference articles. Generated classifications are conservative; the reviewed decisions in `audit-report.md` take precedence.

## Before and after

| Surface | Files before | Files after | Words before | Words after | Change |
| --- | ---: | ---: | ---: | ---: | ---: |
| Marketing website | 76 | 73 | 37,051 | 31,845 | -5,206 (-14.1%) |
| Web application | 273 | 273 | 148,741 | 147,337 | -1,404 (-0.9%) |
| Public documentation | 52 | 53 | 11,862 | 11,348 | -514 (-4.3%) |
| Total customer-facing | 401 | 399 | 197,654 | 190,530 | -7,124 (-3.6%) |

The application count includes implementation code, so its word reduction intentionally remains modest. The cleanup targeted visible copy without deleting behavior. Public documentation added a canonical Troubleshooting page while reducing repeated explanations overall.

The post-cleanup snapshot was generated before this final report and source-of-truth file were added. Those internal audit documents do not affect customer-facing counts.

## Classification and action summary

### Kept

- Legal, privacy, subprocessors, billing consequences, destructive-action warnings, security guidance, permissions, credentials, network requirements, connector limitations, troubleshooting, and technical reference material.
- Existing canonical URLs and direct-link connector status pages.
- Complete technical guidance where shortening could mislead users.

### Kept and improved or shortened

- Homepage hierarchy, connector illustration, navigation, footer, pricing, integrations, comparison, and AI Copilot status copy.
- Public documentation introduction, getting started, connections, pipeline, transformation, billing, FAQ, and status pages.
- Application navigation, onboarding, connection and pipeline empty states, workspace-member terminology, notifications, analytics, billing, and common action labels.

### Merged or moved to a canonical explanation

- Repeated destination DDL and PostgreSQL pipeline walkthroughs now defer to a canonical guide.
- Repeated Normalisation and transformation examples now link to the primary transformation documentation.
- Repeated connector availability notices are shorter and route users to the canonical connector overview.
- Detailed technical explanations remain in documentation rather than marketing or task-focused application copy.

### Archived

- The legacy `/workspace/data-sources` implementation was reduced to its existing redirect to `/workspace/connections`; the old URL and navigation behavior remain intact.
- Historical AWS/ECS and CDK documents now carry archive warnings and point to current Hetzner/Terraform guidance.

### Removed

| Removed content | Reason |
| --- | --- |
| Homepage agent/MCP availability sections | No supported public/self-serve workflow |
| Live-looking Shopify, MySQL, and SQL Server diagram placements | Connectors are not enabled in the current application workflow |
| Unsourced “cheapest,” “4–8x,” and competitor price claims | Time-sensitive and unsupported |
| Unverified support, SLA, SSO, retention, agent, API allowance, and overage promises | Not proven by current implementation or contract evidence |
| Duplicate marketing agent showcase and pricing-comparison components | No remaining consumer after consolidation |
| Dead legacy data-source rendering and client-fetch code | Route already redirects; code was unreachable |
| Repeated documentation paragraphs and examples | Canonical documentation now provides the detail |

No legal page or substantive legal clause was removed.

## Duplicate-content report

The baseline detector found 66 exact repeated blocks. The first post-cleanup snapshot found 64. The remaining matches are predominantly reusable labels, code examples, fixtures, and internal QA material where deduplication would reduce clarity or executable evidence. User-facing high-value duplication was addressed even when the wording was not byte-for-byte identical.

## Outdated-content report

Corrected or clearly archived:

- Starter/Growth plan names and obsolete $49/$199 prices.
- Obsolete Free limits of 3 pipelines and 10,000 rows.
- FAQ copy that treated PostgreSQL as the only available connector.
- Public API, MCP, agent, overage, retention, and Enterprise promises not supported by current evidence.
- Historical AWS/ECS deployment guidance presented without an archive boundary.
- Earlier application names such as Data Pipelines, Data Sources, Team, and Team Members where they referred to current product concepts.

Legal dates, regional commitments, and contractual language were not rewritten without confirmation.

## Unsupported-claim report

Removed from current-availability copy or relabeled:

- AI agents described as active or ready to deploy.
- Public embeds, agent keys, MCP, provider selection, and AI pipeline actions.
- “200+ connectors” as a current product capability.
- Unavailable connectors shown as live.
- Unsupported competitor pricing and superlative claims.

The `/agents` canonical URL remains, but the page now clearly says **AI Copilot — In development**.

## Conflicting-information report

Resolved customer-facing conflicts:

- Plan names: Free, Plus, Pro, Enterprise.
- Prices: $0, $29, $129, and custom; annual monthly equivalents $0, $22, $97, and custom.
- Free limits: 5 pipelines and 25,000 rows.
- Connector availability: PostgreSQL, HubSpot, and Stripe available.
- Slack and GitHub: operational integrations, not data connectors.
- AI Copilot: in development, not generally available.
- Row limits: described as enforced limits, without inventing an automatic overage workflow.

Any future contract-specific Enterprise capability must be confirmed before it is restored to general marketing or plan copy.

## Updated pages and components

### Marketing website

- Homepage shared sections and connector diagram
- Navbar and mobile navigation
- Footer and legal navigation
- AI Copilot, pricing, comparison, connectors, and integrations pages
- Shared marketing content and metadata
- Unused duplicate marketing components removed

### Public documentation

- Documentation navigation and hierarchy
- Home, introduction, quick start, FAQ, pricing, and billing
- Connection overview and connector-status pages
- PostgreSQL destination and example pipeline guidance
- Pipeline Normalisation, SQL transformation, and related user guides
- New canonical Troubleshooting index

### Web application

- Root metadata and onboarding
- Sidebar, global search, and product tour
- Connections and pipeline list copy and empty states
- Connection creation and notification labels
- Pipeline settings and workflow terminology
- Workspace members and roles copy
- Billing, notifications, analytics, and activity empty states
- Legacy data-source redirect implementation

## Content placement changes

- Marketing pages retain value, current capabilities, supported systems, verified pricing, and a primary action.
- Application screens use shorter labels and task-focused empty states. Safety, permission, billing, validation, and destructive-action context remains visible.
- Documentation remains the detailed source for credentials, scopes, connector limitations, schema discovery, preview, mappings, transformations, sync modes, checkpoints, retries, security, billing, and troubleshooting.
- No unnecessary tooltip was added to obvious labels. Longer optional explanations are linked or kept in canonical documentation instead.

## Broken links, SEO, and accessibility

### Links

- The public documentation link check completed with no broken links before testing was stopped.
- A verified internal reference to the retired CDK path was corrected.
- Canonical public URLs and direct-link connector pages were retained; no redirect migration was required.
- External links were inventoried. A complete live-network recheck is intentionally deferred with the remaining test suite.

### SEO

- Meaningful titles, meta descriptions, canonical URLs, sitemap/robots behavior, and indexed connector routes were preserved.
- Unsupported availability claims were removed without deleting useful landing pages.
- Audited marketing pages retained one H1 and showed no horizontal overflow during the completed visual checks.

### Accessibility

- Completed marketing checks found one H1, no unnamed buttons, and no missing image alt text on the sampled desktop and mobile pages.
- Completed public authentication checks found labeled controls, one H1, and no unnamed buttons or horizontal overflow.
- Dialog, form, keyboard, and table behavior was not intentionally changed by the content cleanup.

## Functional validation status

Testing was stopped at the user's request to conserve the weekly usage limit. No further test, build, lint, browser, or E2E command was run while completing this report.

Completed before that request:

| Check | Result |
| --- | --- |
| Marketing lint | Passed (150 files) |
| Marketing production build | Passed (14 routes) |
| Application unit tests | Passed (55 tests) |
| Application production build and TypeScript | Passed after the final application reliability changes |
| Public documentation broken-link check | Passed |
| Marketing desktop/mobile/tablet visual checks | Passed for the audited routes |
| Public authentication route visual/accessibility checks | Passed |
| Authentication E2E | Passed |
| Connection form/create workflow | Passed |
| PostgreSQL schema, table, and records UI workflow | Passed |

Not completed in this session:

- The full PostgreSQL lifecycle Playwright case did not achieve a final green run before testing was stopped. Backend evidence showed the created pipeline runs completed successfully; the remaining failure was in run-history polling/reload test behavior.
- App lint had passed earlier, but was not rerun after the final small reliability edits.
- The complete cross-surface external-link, accessibility, and E2E suites remain for the next testing session.

These deferred checks are not represented as passed.

## Non-content reliability changes made during verification

Verification exposed two application reliability issues that were corrected without changing product contracts:

- Pipeline stream drafts are protected from being overwritten by an equivalent background schema refetch.
- The query editor retains a usable plain-text fallback when Monaco initialization fails.

The related Playwright page object was updated to match current schema-sync
endpoints, settings tabs, and terminal status wording. A later connector-focused
real-Chrome run completed the PostgreSQL, Stripe, and HubSpot customer paths;
the broader Playwright suite remains a separate regression concern.

## Remaining content and technical debt

- Obtain legal/product confirmation for retention, regions, Enterprise commitments, SLA/SSO, overage behavior, public API availability, and future AI capabilities.
- Review or replace screenshots that show earlier pipeline-builder and connection UI once a current approved screenshot set exists.
- Decide whether low-value direct-link future connector status pages should remain indexed; preserve URLs and add redirects if that policy changes.
- Complete the deferred live external-link, accessibility, and broader
  Playwright checks in a fresh testing session.
- Connector verification resources are intentionally retained for inspection;
  review them explicitly before any future production cleanup.
- Re-run the repeatable content inventory after future product changes and treat unexpected plan or availability conflicts as release blockers.

## Completion assessment

The controlled content cleanup is complete. Important technical, billing, security, permission, limitation, troubleshooting, and legal context remains accessible; unsupported and contradictory customer-facing claims were removed or relabeled; routes and product behavior were preserved.

The connector-specific real-Chrome verification is complete. Remaining deferred
checks are the broader accessibility, external-link, and cross-product
regression items listed above; they are not represented as passed.

## 28 July connector documentation follow-up

The retained real-Chrome connector verification resolved the previous Stripe
availability uncertainty. The canonical record, marketing site, and public docs
now consistently present PostgreSQL, HubSpot, and Stripe as available without
an obsolete availability qualifier.

The public documentation now also includes:

- complete Stripe authentication, 34-stream, sync-mode, destination, error, and
  limitation guidance;
- expanded PostgreSQL permissions, discovery, type, Full Table, Incremental,
  security, and troubleshooting guidance;
- links from the HubSpot reference to a complete ten-stream example;
- reproducible PostgreSQL 8,000-row initial and 2,000-row incremental sample;
- Stripe-to-PostgreSQL and HubSpot-to-PostgreSQL pipeline examples.

Mintlify reported no broken internal links. Website lint, TypeScript, and the
production build completed successfully.
