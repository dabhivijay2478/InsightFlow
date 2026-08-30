# MantrixFlow content source of truth

Date: 2026-08-13
Status: current repository-verified content baseline

Use this document when writing product copy. If the implementation changes, update this document from working application behavior and backend contracts before updating marketing or documentation.

## Evidence priority

1. Working application behavior
2. Backend API contracts
3. Database and plan configuration
4. Frontend implementation
5. Public documentation
6. Marketing copy
7. Historical plans, drafts, and reports

Runtime code alone does not make a feature publicly available. A capability must also be enabled in the current customer workflow.

## Product terminology

| Preferred term | Meaning | Avoid as a synonym |
| --- | --- | --- |
| Connection | Saved credentials and configuration for a system | Integration, connector |
| Source | A connection MantrixFlow reads from | Dataset |
| Destination | A connection MantrixFlow writes to | Target system |
| Pipeline | A configured data movement workflow | Flow, job |
| Stream | A selectable source object or table | Dataset, object unless required by the connector |
| Schema | Discovered structure and field definitions | Structure |
| Data Preview | A sample of source data | Sample view |
| Mapping | Source-to-destination field configuration | Binding |
| Transformation | SQL or Normalisation logic applied to data | Processing rule |
| Pipeline Run | One execution of a pipeline | Job, sync, execution |
| Run History | Previous pipeline runs and outcomes | Execution history |
| Workspace | The organization-scoped product area | Team |
| Workspace Member | A person with workspace access | Team member, workspace user |
| AI Copilot | Planned guided AI product surface | Agent, assistant unless technically distinct |
| Connector | A supported system type in the catalog | Connection |
| Integration | An operational link such as Slack or GitHub | Data connector |

Button labels should describe the specific action: **Create Connection**, **Test Connection**, **Save Connection**, **Refresh Schema**, **Preview Data**, **Create Pipeline**, **Validate Pipeline**, **Run Pipeline**, **Retry Run**, **Pause Pipeline**, and **Delete Pipeline**.

## Available now

### Data connectors

| Connector | Role | Stage | Authentication/content notes |
| --- | --- | --- | --- |
| PostgreSQL | Source and destination | Available | Preserve connection, network, permissions, schema, preview, and destination requirements in connector documentation |
| MySQL | Source and destination | Available | Runtime-capability gated; preserve database, grants, schema discovery, Upsert, and explicit table-contract requirements |
| Airtable | Source and destination | Available | Runtime-capability gated; PAT scopes, base/table IDs, writable field mapping, and existing-table delivery contract apply |
| Asana | Source | Available | PAT-only; eight streams; pipeline-specific workspace/project scope; tasks support `modified_at` incremental extraction; not CDC |
| HubSpot | Source | Available | Uses private-app token authentication; preserve required scopes and stream limitations |
| Stripe | Source | Available | Supports 34 tested streams for PostgreSQL delivery; preserve authentication, supported objects, pagination, empty-stream behavior, and limitation guidance |

The catalog may contain additional connector definitions for future work. Do not describe them as selectable production workflows unless the frontend enablement gate and end-to-end customer path are active.

### Core workflow

- Create, test, save, edit, and delete connections.
- Discover source schemas and preview source data.
- Create and configure pipelines.
- Select streams and configure field mappings.
- Configure SQL transformations and Normalisation.
- Use Full Table and Incremental sync where offered by the selected source and configuration.
- Validate and run pipelines.
- View run history, run details, statuses, and errors.
- Retry supported failed runs.
- Manage workspace members subject to role permissions.
- View billing usage and enforced plan limits.

Keep connector-specific and mode-specific limitations in the relevant reference page. Do not generalize support across all connectors.

### Operational integrations

- Slack is an operational integration, not a data source connector.
- GitHub pipeline export and synchronization is an operational integration, not a GitHub data source.

### Plans and billing

The backend plan configuration is authoritative for plan names and core limits.

| Plan | Monthly price | Annual monthly equivalent | Pipelines | Rows | Storage |
| --- | ---: | ---: | ---: | ---: | ---: |
| Free | $0 | $0 | 5 | 25,000 | 1 GB |
| Plus | $29 | $22 | 20 | 1,000,000 | 5 GB |
| Pro | $129 | $97 | 50 | 10,000,000 | 20 GB |
| Enterprise | Custom | Custom | Unlimited | Unlimited | Custom |

The current row limit is enforced as a hard limit. Do not promise automatic overage charging, API-call allowances, retention periods, SSO, SLAs, custom connector delivery, or dedicated infrastructure unless the applicable contract and implementation have been confirmed.

## Availability labels

- **Available**: enabled in the current customer workflow.
- **Enterprise-only**: available only when the backend entitlement and customer contract both support it.
- **In development**: active product work without a supported public workflow.
- **Planned**: intended future work without a delivery promise.
- **Roadmap target**: an aspiration with an explicit date, not current availability.
- **Not currently supported**: not available in the current customer workflow.

The target of 200 connectors by December 2026, if mentioned, must be called a **roadmap target**. It must never be written as current inventory.

## Not publicly available or not verified

Do not present the following as available without new implementation and product evidence:

- Public/self-serve AI Copilot workflows
- Public agents, agent embeds, or public `agent_key`
- Public MCP or `skill.md` integrations
- Multi-provider AI model selection
- AI tools that create or run pipelines
- A public React agent SDK workflow
- A supported public external API and plan-level API allowances
- SSO/SAML, formal SLA commitments, dedicated infrastructure, or custom connector delivery terms
- MariaDB, SQL Server, Oracle, SQLite, CockroachDB, warehouse, Shopify, Salesforce, Notion, or GitHub data-source workflows

AI Copilot marketing must remain labeled **In development** until a supported customer workflow is enabled. Internal routes, meters, experiments, and planning documents are not proof of public availability.

## Content placement rules

### Marketing

Describe the user problem, core workflow, currently supported systems, verified benefits, pricing, and a primary call to action. Keep implementation detail in documentation. Availability labels must be visible where a future capability is mentioned.

### Application

Show the minimum copy needed to complete the current task. Preserve safety warnings, billing consequences, permission restrictions, and destructive-action consequences. Use a short helper, tooltip, advanced section, or documentation link for optional detail.

### Documentation

Preserve complete setup, permissions, credential, network, schema, preview, mapping, transformation, sync, checkpoint, retry, security, billing, troubleshooting, migration, and limitation guidance. Keep one canonical explanation and link to it instead of repeating it.

### Legal and security

Do not simplify substantive legal, privacy, retention, deletion, subprocessors, security, cancellation, or contractual language without legal or product confirmation.

## Infrastructure documentation

Current infrastructure guidance is the OVHcloud Terraform and self-hosted Dokploy documentation in `apps/mantrixflow-infra`. AWS/ECS and CDK deployment documents are historical references and must retain an archive warning; they are not current deployment instructions.

## Change control

Before publishing a new claim:

1. Identify its implementation or contract evidence.
2. Confirm its availability stage and entitlement.
3. Check for conflicting pricing, plan, connector, and limitation copy.
4. Update the canonical documentation page first.
5. Link secondary pages to the canonical explanation.
6. Preserve route, canonical URL, redirects, structured data, and legal context.

## July 2026 connector verification

The current Stripe availability claim is backed by the retained real-browser
pipeline verification completed on 26 July 2026:

- all 34 discoverable Stripe streams were selected and completed;
- PostgreSQL delivery completed successfully;
- valid empty streams completed without false failures;
- connection, discovery, preview, transformation, validation, delivery, and
  run-history paths were exercised in the customer UI.

The retained evidence is in
`md-docs/testing/frontend/reports/stripe-stream-coverage.md` and
`md-docs/testing/frontend/reports/final-test-summary.md`.

## August 2026 Asana verification

Asana availability is backed by the retained customer-UI verification completed
on 13 August 2026:

- a Personal Access Token was tested, saved, masked, and reused through the
  customer connection flow;
- workspace and project scope was discovered and saved on the pipeline;
- all eight streams (`workspaces`, `projects`, `sections`, `tags`, `tasks`,
  `stories`, `teams`, and `users`) were extracted;
- PostgreSQL, MySQL, and Airtable delivery runs completed successfully with zero
  failed rows;
- direct MySQL counts matched the final run's 20,471 delivered rows;
- the source remains PAT-only, source-only, and non-CDC; stories require a Full
  Table task traversal.
