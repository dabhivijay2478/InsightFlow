# HubSpot Source — Manual Testing Guide

Status: production

Runtime: dlt only

Supported destination for this connector release: PostgreSQL only

## Connection

Create a read-only HubSpot private app and save its access token in the
`credential` field. Grant only the scopes required by the selected streams.
Never paste the token into pipeline JSON, SQL, logs, or AI chat. Rotate it on
your organization's security schedule.

The connection test is permission-aware: a valid token with missing optional
scopes remains connected with limited permissions. Discovery marks each stream
as available, missing scope, unavailable, or unknown and never returns CRM rows.

## Stable catalog

| Stream | dlt resource | DuckDB relation | PK | Mode |
| --- | --- | --- | --- | --- |
| `hubspot.contacts` | `contacts` | `raw.hubspot__contacts` | `id` | Incremental |
| `hubspot.companies` | `companies` | `raw.hubspot__companies` | `id` | Incremental |
| `hubspot.deals` | `deals` | `raw.hubspot__deals` | `id` | Incremental |
| `hubspot.tickets` | `tickets` | `raw.hubspot__tickets` | `id` | Incremental |
| `hubspot.owners` | `owners` | `raw.hubspot__owners` | `id` | Full snapshot |
| `hubspot.deal_pipelines` | `pipelines_deals` | `raw.hubspot__deal_pipelines` | `pipeline_id, stage_id` | Full snapshot |
| `hubspot.ticket_pipelines` | `pipelines_tickets` | `raw.hubspot__ticket_pipelines` | `pipeline_id, stage_id` | Full snapshot |
| `hubspot.products` | `products` | `raw.hubspot__products` | `id` | Incremental |
| `hubspot.line_items` | `line_items` | `raw.hubspot__line_items` | `id` | Incremental |
| `hubspot.quotes` | `quotes` | `raw.hubspot__quotes` | `id` | Incremental |

No other HubSpot streams are supported in this release. Property history,
archived/deleted capture, activities, forms, custom objects, OAuth, webhooks,
and write-back are not supported.

## Required checks

1. Test the connection and run live discovery.
2. Confirm at least one available stream is selected. Empty selection must fail.
3. Set a UTC start date for the first incremental run; optionally set an end
   date and a lookback from 0 through 604800 seconds (default 3600).
4. Keep custom properties enabled only when needed. Sensitive properties are
   excluded; visible fallback warnings mean the run used standard properties.
5. Confirm preview masks email local parts, phone numbers, and free-text
   note/content/body fields.
6. Map every source to an explicit UI SQL/dbt model.
7. Point every model at a pre-existing PostgreSQL table with a primary or stable
   merge key. Upsert is the only accepted delivery mode.
8. Run twice. The second incremental run must use the overlapped change window,
   deduplicate by ID/update time, and commit only successfully delivered streams.
9. Verify callback metadata includes selected streams, phase status,
   `phase3_rows_delivered`, cleanup status, and committed/blocked checkpoint
   streams without credentials or CRM values.
10. Test `DELETE .../sync-state/hubspot.contacts` as an editor and verify other
    stream checkpoints remain unchanged.

HubSpot calculated-property-only changes may not advance the object update
timestamp. Scheduled calculated-property reconciliation is a future phase and
is not an MVP guarantee.
