# MantrixFlow content audit and controlled cleanup plan

Date: 2026-07-22  
Status: audit complete enough to begin controlled copy changes; product-confirmation items remain frozen

## Scope and audit trail

This audit covers the marketing website, the public Mintlify documentation, the authenticated web application, and the internal repository documents used to verify claims.

No marketing, documentation, legal, or application copy was changed before this report and the generated baseline were created.

Generated evidence:

- [`generated/content-inventory.csv`](./generated/content-inventory.csv) — one row per route, layout, content-bearing component, dialog/tooltip host, public documentation article, and internal reference article
- [`generated/baseline-manifest.json`](./generated/baseline-manifest.json) — SHA-256, source word count, line count, and byte count for every inventoried file
- [`generated/duplicate-content.csv`](./generated/duplicate-content.csv) — exact repeated content blocks for human review
- [`generated/link-inventory.csv`](./generated/link-inventory.csv) — internal, external, contact, and local-file links
- [`generated/summary.json`](./generated/summary.json) — inventory totals
- [`../../scripts/content-audit.mjs`](../../scripts/content-audit.mjs) — repeatable audit generator

The generated classifications are conservative first-pass classifications. Every `REQUIRES PRODUCT CONFIRMATION` item is frozen until implementation evidence exists. Human-reviewed overrides in this report take precedence over a generated classification.

## Inventory summary

| Surface | Inventoried files | Current state |
| --- | ---: | --- |
| Marketing website | 76 | 11 implemented routes, shared navigation/footer, and reusable marketing sections |
| Web application | 273 | 46 application routes plus layouts, content-bearing components, dialogs, tooltips, notifications, errors, and shared content sources |
| Public documentation | 52 | 50 MDX pages, Mintlify navigation, and repository guidance; about 11,400 MDX words |
| Internal reference | 137 | About 133,000 words of engineering guidance, plans, runbooks, QA procedures, and reports |
| Raw internal test evidence | 157 files | About 122,000 words; not public documentation |

The inventory contains 57 page routes, 5 layouts, 18 dialog/sheet hosts, 5 explicit tooltip hosts, 188 Markdown/MDX articles, and 64 external links that require a network check.

## Source-of-truth findings

### Verified available now

| Capability | Evidence | Content decision |
| --- | --- | --- |
| PostgreSQL source and destination | Enabled frontend connector; complete create/test/save path; backend runtime | Present as live |
| HubSpot source to PostgreSQL | Enabled frontend connector; backend registry/runtime; public connector guide | Present as available; private-app token authentication |
| Stripe source to PostgreSQL | Enabled frontend connector with `beta` release stage; public guide | Present as Beta only |
| Pipeline builder | Current app routes and structured pipeline APIs | Present as available |
| Schema discovery and source preview | Current app services and builder UI | Present as available |
| Full Table and Incremental sync | Current app types, UI, backend dispatch, and checkpoint behavior | Present as available with current restrictions |
| SQL transformations and Normalisation | Current destination-owned transformation UI and backend | Present as available |
| Pipeline runs and run history | Current app and backend routes | Present as available |
| Slack operational integration | Settings UI, Next proxy, Go OAuth/events/commands routes | Present as an integration, not a data connector |
| GitHub pipeline export and sync | Settings UI, app services, Go GitHub and YAML routes | Present as an integration, not a data connector |
| Free, Plus, Pro, Enterprise plans | Backend `internal/config/plans.go` and billing service | Use these names only |
| Core plan prices and limits | Backend and marketing pricing page agree | Free: 5 pipelines/25K rows; Plus: $29, 20/1M; Pro: $129, 50/10M; Enterprise: custom/unlimited rows and pipelines |

### Not available as a self-serve cloud workflow

The frontend catalog contains many future connectors, but only PostgreSQL, HubSpot, and Stripe pass the current enablement gate. MySQL, MariaDB, SQL Server, Oracle, SQLite, CockroachDB, warehouses, Shopify, GitHub data source, Notion, Salesforce, Slack data source, and other catalog entries must not be presented as currently selectable end-to-end workflows.

Some of these connectors have runtime code or internal QA documents. That does not override the current application enablement gate.

### Requires product confirmation

The following claims must not be removed from legal agreements or silently converted into promises. They must be hidden from availability copy or marked clearly as planned until verified:

- MCP server and `skill.md` agent integration
- public agent embeds and public `agent_key`
- multi-provider model selection
- agent domain/table allowlists and 100-requests-per-session limit
- agent tools that create or run pipelines
- public React agent SDK workflow
- broad AI Copilot availability in the web application
- SSO/SAML availability
- formal SLA response or uptime commitments
- custom connector delivery terms
- dedicated infrastructure availability
- plan-specific run-history retention promises
- automatic row-overage charging versus hard run pause at the plan limit
- API-call allowances and API-call overage pricing
- “200+ connectors” target timing

## Marketing website inventory and recommendations

| Route or surface | Purpose | Accuracy | Duplication/detail | Status | Proposed action |
| --- | --- | --- | --- | --- | --- |
| `/` | Explain value and start flow | Partly accurate | Agent sections duplicate `/agents`; connector visual implies unavailable connectors are live | KEEP AND IMPROVE | Keep one message and CTA; show only enabled connectors; remove unsupported agent availability from homepage |
| `/connectors` | Connector availability catalog | Mostly accurate | Intro repeats availability copy | KEEP AND IMPROVE | Keep indexed page; make availability labels the source of truth; ensure Stripe Beta is visible |
| `/integrations` | Slack and GitHub operational integrations | Implemented | 537-line page repeats feature/trust descriptions | SHORTEN | Keep functional claims; reduce repeated cards and decorative explanations |
| `/agents` | Agent product marketing | Unsupported as presented | Repeats homepage agent sections and internal agent drafts | REQUIRES PRODUCT CONFIRMATION | Preserve URL and SEO metadata; replace current-availability claims with an honest development-status page |
| `/pricing` | Plans, limits, checkout links | Core prices/limits verified | Feature matrix includes unverified promises | KEEP AND IMPROVE | Keep verified limits and billing terms; label or remove unverified feature promises without changing prices |
| `/compare` | Competitor price comparison | Unsupported and time-sensitive | Unsourced “cheapest,” “4–8x,” and competitor price claims | UPDATE | Preserve URL; replace competitor assertions with a factual MantrixFlow plan/ownership comparison unless dated sources are added |
| `/legal` | Legal index | Accurate | Low duplication | KEEP | Keep |
| `/privacy` | Privacy and data handling | Requires legal review for retention/region wording | Necessary detail | REQUIRES PRODUCT CONFIRMATION | Do not shorten substantive terms; standardize brand spelling only after review |
| `/terms` | Service terms | Requires legal review | Necessary detail | REQUIRES PRODUCT CONFIRMATION | Do not remove or rewrite legal obligations without approval |
| `/subprocessors` | Provider disclosure | Requires operational/legal verification | Necessary detail | REQUIRES PRODUCT CONFIRMATION | Preserve; verify provider list and regions before copy edits |
| Navbar/mobile navigation | Primary navigation | Agents is already marked “Soon” only in mobile detail, not desktop | Product and top-level lists duplicate | KEEP AND IMPROVE | Use one consistent availability label; keep URLs |
| Footer | Core links and legal access | Omits Legal Hub and Subprocessors | Concise | KEEP AND IMPROVE | Add existing legal links without adding marketing copy |
| Metadata/sitemap/robots | SEO discovery | Agents and comparison metadata repeat unsupported claims | Concise | UPDATE | Preserve canonicals and URLs; rewrite claims; retain legal and connector routes |
| Empty route directories | Unimplemented future routes | No page implementation | Not navigable or in sitemap | ARCHIVE | Remove empty directories only; no route or indexed page exists to lose |

The homepage currently presents Shopify as a source and MySQL/SQL Server as destinations in its primary live-looking diagram. That is an availability error, not merely a wording preference.

## Public documentation inventory and recommendations

### Current public hierarchy

The Mintlify navigation currently splits material among Introduction, Getting started, Pricing, Connections, Pipelines, Sync reference, Platform, a separate User guide, and Examples. This creates overlap between topic reference pages and screenshot-led user guides.

### High-priority conflicts

| Location | Conflict | Status | Proposed action |
| --- | --- | --- | --- |
| `pricing.mdx` | Uses obsolete Starter/Growth names, $49/$199 prices, 3 pipelines/10K Free limit, API allowances, real-time pipelines, MCP, and agent claims | UPDATE | Replace with backend-verified Free/Plus/Pro/Enterprise prices and limits; remove unverified allowances |
| `billing-breakdown.mdx` | Repeats obsolete plans and unverified API/row overage formulas | REQUIRES PRODUCT CONFIRMATION | Retain hosted-checkout, upgrade/downgrade, and billing-cycle guidance; remove formulas not proven by current customer flow |
| `getting-started/faq.mdx` | Says only PostgreSQL is live, conflicting with Connections Overview, HubSpot, and Stripe Beta | UPDATE | Correct live connector list and link to the single canonical connector table |
| `index.mdx` | Omits Stripe Beta and says broad chat/notifications are not live despite Slack integration being live | UPDATE | Use precise distinctions: connector, integration, and future AI capability |
| `connections/overview.mdx` | Correct high-level connector status but says “another available source,” which is less precise than the table | KEEP AND IMPROVE | Link connector names directly; use exact availability wording |
| Hidden MySQL/MariaDB/SQLite/CockroachDB/GitHub/Shopify/Notion pages | Direct-link status notices repeat identical paragraphs | MERGE WITH ANOTHER PAGE | Preserve URLs; shorten each to one status notice and canonical connector link |
| CDC pages | Direct-link status notices for unavailable change capture | MERGE WITH ANOTHER PAGE | Preserve URLs; keep one canonical status explanation and cross-links |

### Duplicate groups

- `getting-started/quick-start.mdx`, `example/pipelines/postgres-to-postgres.mdx`, and `connections/destinations/postgresql.mdx` repeat the same destination DDL.
- `pipelines/normalisation.mdx`, `pipelines/transformation-rules.mdx`, and `user-guide/transformations-and-branches.mdx` repeat the same rename/cast example.
- `pipelines/dbt-layer.mdx` and `pipelines/transformation-rules.mdx` overlap on SQL Layer behavior.
- `getting-started/introduction.mdx`, `index.mdx`, and `user-guide/index.mdx` repeat orientation and reading-order content.
- Source and destination status-notice pages contain repeated boilerplate.

### Proposed public hierarchy

```text
Getting Started
Connections
Connector Reference
Pipelines
Transformations
Scheduling and Sync Modes
Pipeline Runs
Workspace
Billing and Usage
Security and Data Handling
Troubleshooting
Examples
Changelog
```

AI Copilot and API Reference should not be added to public navigation until a supported public surface and contract are verified. Existing direct-link status pages remain reachable to protect inbound links and SEO.

## Application content inventory and recommendations

The per-file application inventory is in `generated/content-inventory.csv`. The following groups are the human-reviewed priorities.

| Surface | Finding | Status | Proposed action |
| --- | --- | --- | --- |
| Sidebar | “Data Pipelines” differs from the product term “Pipelines” | KEEP AND IMPROVE | Use “Pipelines” while keeping route unchanged |
| Connections | Main empty catalog message is clear; some nested empty states repeat generic “No data available” | KEEP AND IMPROVE | Standardize actionable empty states |
| Legacy `/workspace/data-sources` | Immediately redirects but retains a full legacy UI and old “Data Source” copy | ARCHIVE | Preserve redirect behavior; remove dead render-only copy only after route verification |
| Pipeline list | Generic “No pipelines found” does not distinguish an empty workspace from no search results | KEEP AND IMPROVE | Use task-focused empty and no-result states |
| Pipeline builder | Large amount of contextual guidance and repeated save/preview instructions | SHORTEN | Keep safety warnings and workflow prerequisites; move background detail to docs links |
| Transformations | Repeated “save, validate, preview” guidance appears in multiple panels | SHORTEN | Keep one ordered action hint at the point of use |
| Analytics | Several empty states use long operational explanations | SHORTEN | State what is missing and link to the relevant action |
| Notifications | Empty copy is informative but long | SHORTEN | Use one sentence and the next action |
| Team/workspace | “Team,” “members,” and “workspace users” vary | KEEP AND IMPROVE | Use “Workspace members” for the feature and “member” for a person |
| Billing | “AI agent interactions (coming soon)” coexists with marketing “ready to deploy” | UPDATE | Keep “coming soon” until the UI workflow exists |
| Errors | Service-layer messages can surface internal object names and inconsistent next actions | KEEP AND IMPROVE | Preserve technical codes, sanitize secrets, and standardize actionable fallback copy |
| Tooltips | Few explicit content tooltips; most are navigation labels | KEEP | Do not add tooltips to obvious controls; use docs links for long explanations |

## Duplicate-content report

The generated detector found 66 exact repeated blocks across public and internal surfaces. Many are legitimate code examples or internal QA fixtures. The cleanup targets only user-facing repetition.

High-value merge candidates:

- repeated connector status notices in public docs
- repeated destination DDL in three public docs pages
- repeated Normalisation examples in three public docs pages
- repeated homepage/Agents architecture and workflow copy
- repeated homepage connector availability descriptions
- repeated application member-role restrictions
- duplicate top-level empty-state phrases that do not distinguish no data from no search results

Internal QA SQL fixtures will not be deduplicated merely to reduce words; they are executable test evidence.

## Outdated-content report

Confirmed outdated:

- Public docs Starter/Growth plan names and $49/$199 prices
- Public docs 3-pipeline and 10K-row free limits
- Public docs API-call plan allowances and rates without a current public API contract
- Public docs MCP/agent plan claims
- Public docs FAQ that omits HubSpot and Stripe
- Dated local ngrok URL in internal Slack guides
- Historical connector plans presented beside current operational docs without a clear archive boundary

Potentially outdated and frozen pending confirmation:

- screenshot sets showing the earlier canvas-node layout or earlier connection forms
- legal “last updated” dates and infrastructure-region statements
- plan-specific run-history retention periods
- Enterprise SSO, SLA, custom connector, and dedicated infrastructure claims

## Unsupported-claim report

Remove from availability copy or relabel as planned:

- “Active Integration Layer” and “Ready to deploy” on `/agents`
- create-an-agent CTA
- public agent embeds and `agent_key`
- MCP workflow demonstrations represented as a current product
- multi-provider model configuration
- “200+ Community Connectors” without an explicit roadmap label and date
- homepage live-looking Shopify/MySQL/SQL Server connector diagram
- “cheapest at every tier,” “4–8x cheaper,” and “best market discount” comparison claims

## Conflicting-information report

| Topic | Conflict |
| --- | --- |
| Plans | Public docs use Starter/Growth; app/backend/website use Free/Plus |
| Prices | Public docs use $49/$199; app/backend/website use $29/$129 |
| Free limits | Public docs use 3 pipelines/10K rows; backend/website use 5/25K |
| Connectors | Docs FAQ says only PostgreSQL; docs overview and app enable HubSpot and Stripe Beta |
| HubSpot stage | App catalog and product docs treat HubSpot as production; two backend dispatch errors still call it beta |
| Agents | App billing dialog says coming soon; marketing says active and ready to deploy |
| Overage | Website FAQ says runs pause at 100%; other marketing and docs mention row-based overage; backend contains both limit and overage infrastructure |
| Integration vs connector | Homepage and content sometimes display GitHub/Slack/Shopify without distinguishing operational integration, data source, and future connector |

## Proposed removal, merge, and movement plan

### Safe to change now

1. Correct public docs plan names, prices, and enforced limits.
2. Remove unsupported agent availability claims while preserving `/agents` and its canonical URL.
3. Remove unavailable connectors from the homepage live diagram; keep them in the status-labeled connector catalog.
4. Remove unsourced competitor price assertions while preserving `/compare`.
5. Correct the docs connector-list conflict.
6. Deduplicate public docs examples by keeping one canonical explanation and linking from secondary pages.
7. Reorganize Mintlify navigation without changing page URLs.
8. Standardize non-legal MantrixFlow spelling and core application terminology.
9. Shorten top-level application empty states and helper copy without changing event handlers, validation, permissions, API calls, routes, or state.
10. Fix verified local broken links.

### Must remain until confirmation

- legal terms, privacy commitments, subprocessors, retention, and deletion language
- Enterprise contractual promises
- exact overage charging behavior
- public API availability and allowances
- AI/agent product availability beyond the current hidden/backend interaction meter
- connector availability that is implemented in a runtime but disabled in the cloud UI

### Preserve even when long

- credential and permission requirements
- OAuth/private-app scopes
- network allowlisting requirements
- destination table prerequisites
- data type, mapping, SQL, sync, checkpoint, retry, and recovery behavior
- security and data-handling details
- destructive-action warnings
- billing consequences
- troubleshooting and known limitations

## Terminology guide

| Term | Meaning | Avoid when referring to this concept |
| --- | --- | --- |
| Connection | Reusable saved credentials and connector configuration | Integration when credentials are meant |
| Connector | A supported source/destination capability in the catalog | Connection when referring to the product capability |
| Source | The system a pipeline reads from | Integration |
| Destination | The system/table a pipeline writes to | Sink in user-facing copy |
| Pipeline | Saved source-to-destination workflow | Flow, job |
| Stream | A selectable source object/table/API resource | Dataset or object unless the API uses that distinct concept |
| Schema | Database namespace or discovered field structure, according to context | Dataset |
| Data Preview | A limited preview of source or transformed rows | Sample when referring to the named feature |
| Mapping | Explicit relationship between source/model fields and destination fields | Binding |
| Normalisation | Rename/cast rules applied before SQL | Transformation when the SQL layer is meant |
| Transformation | A saved SQL model owned by a destination | Filter node or dbt layer when referring to the current UI feature |
| Pipeline Run | One execution record for a pipeline | Job, sync, execution |
| Run History | List of pipeline runs | Execution history |
| Workspace | Organization-scoped application context | Tenant in user-facing copy |
| Workspace member | Person with a workspace role | Team user |
| Integration | Slack or GitHub operational connection | Data connector |
| AI Copilot | Reserved product term for a verified in-app assistant | Agent for unavailable public/MCP functionality |

## Safety gates for implementation

Before removing or merging a user-facing item:

1. Verify its implementation and inbound links.
2. Preserve the URL or add an explicit redirect.
3. Preserve legal, billing, security, permissions, and destructive-action context.
4. Keep one canonical technical explanation in public docs.
5. Update every link to the canonical page.
6. Run the relevant build, lint, link, and flow checks.
7. Compare source word counts against the baseline manifest.

## Implementation sequence

1. Fix verified claims and unsupported marketing availability copy.
2. Correct and reorganize public documentation.
3. Update shared terminology and top-level application copy.
4. Review dialogs, empty states, errors, notifications, and helper text feature by feature.
5. Validate internal/local links, then external links.
6. Run website, docs, and app builds/lint/tests.
7. Run critical browser flows and responsive/accessibility checks.
8. Regenerate the inventory and produce before/after counts and a final report.
