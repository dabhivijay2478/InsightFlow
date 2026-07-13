# HubSpot UI — Production Acceptance Guide

Use only the dlt-based HubSpot source with an existing PostgreSQL destination.
The canonical catalog has ten streams: contacts, companies, deals, tickets,
owners, deal pipelines, ticket pipelines, products, line items, and quotes.

1. Create a HubSpot connection with a read-only private app token. Confirm the
   saved UI masks the token and the connection test returns no CRM values.
2. Open discovery. Verify available streams can be selected and missing-scope
   streams remain visible but disabled with a permission badge.
3. Select one or more streams. Verify contacts/companies/deals/tickets/products/
   line items/quotes recommend incremental; owners and both pipeline streams
   remain full snapshot.
4. Configure the UTC start/end window, bounded lookback, and custom-property
   toggle. Sensitive fields must never be selectable.
5. Preview a stream. Email local parts, phones, and note/content/body fields must
   be masked.
6. Save explicit UI SQL/dbt for every selected stream. The destination panel
   exposes only upsert and requires a pre-existing PostgreSQL table with a PK.
7. Run and inspect phase status, delivered rows, cleanup, warnings, and committed
   versus blocked streams. Missing table/column/type/PK must hard fail.
8. Use the reset icon on a selected stream, confirm the warning, and verify only
   that stream restarts from its configured lower bound.
9. Paste a marker `pat-...` token into AI chat and verify the request is rejected
   locally with `SECRET_DETECTED` before model or usage calls.

The production regression covers live contacts, companies, deals, tickets, and
owners E2E plus the controlled remaining-stream matrix and destination checks.
