# MantrixFlow Stripe Source Connector AI Prompts

**Document purpose:** AI prompt pack for designing or extending the Stripe
source connector in MantrixFlow.

**Last updated:** July 2026
**Project:** MantrixFlow
**Connector:** Stripe source

> This is a prompt/reference pack, not the authoritative runtime
> implementation spec. Runtime changes must still follow the strict ELT
> invariants, the active connector code, and the repo rules.

---

## 1. Master Prompt

Use this prompt first to get the overall architecture.

```text
You are an expert ELT data engineer building a production-grade Stripe source connector for MantrixFlow.

MantrixFlow core principles:
- Strict 5-phase ELT: Preflight -> Read Source -> Transform (SQL) -> Deliver (Upsert) -> Cleanup
- Upsert-only into user-owned destination tables
- Strong preflight validation for connection, permissions, schema match, and disk budget
- Isolated DuckDB staging
- Full audit metadata and lineage
- High observability with detailed phase logs

Requirements for the Stripe connector:
- Support key Stripe objects: charges, customers, subscriptions, invoices, refunds, balance_transactions, events, payouts, products, prices
- Incremental sync using created and updated timestamps where available
- Full refresh fallback
- Proper pagination with limit and starting_after
- Rate limit handling that respects Stripe limits and retry guidance
- Expand nested objects when useful, for example customer details in charges
- Schema discovery that can fetch and describe object structure dynamically

Output structure:
1. High-level architecture diagram in text
2. List of supported streams with incremental strategy
3. Preflight checks needed
4. Data extraction logic steps
5. Normalization and flattening approach
6. Error handling strategy
7. Security and credential management
8. Future extensibility points

Follow MantrixFlow coding style: clean, observable, defensive, with detailed logging.
Respect MantrixFlow 5-phase ELT, upsert-only delivery, and high observability standards.
```

---

## 2. Individual Stream Prompts

### Charges Stream

```text
Create detailed specifications for the Stripe "charges" stream connector.

Context:
- Use the current stable Stripe API behavior supported by the project
- Incremental mode primary, using created >= last_sync
- Support timestamp filters plus cursor-based pagination
- Fields to extract: id, amount, amount_captured, amount_refunded, currency, customer, status, created, paid, refunded, description, metadata, invoice, payment_intent, balance_transaction, receipt_email, receipt_url, failure_code, failure_message
- Expand customer, invoice, payment_intent, and balance_transaction only when useful and safe for payload size
- Handle deleted, disputed, failed, partially refunded, and fully refunded charges gracefully

Provide:
- API endpoint and parameters
- Incremental logic and state fields
- Schema definition with key fields and types
- Flattening rules for nested objects and metadata
- Sample JSON response shape
- Edge cases to handle
- Logging statements for observability

Respect MantrixFlow 5-phase ELT, upsert-only delivery, and high observability standards.
```

### Customers Stream

```text
Create detailed specifications for the Stripe "customers" stream connector.

Context:
- Incremental mode should use created >= last_sync unless an updated-like field is available through the selected API path
- Fields to extract: id, email, name, phone, description, currency, created, delinquent, livemode, metadata, address, shipping, invoice_settings, default_source, balance, tax_exempt
- Include a subscriptions summary only if it does not create excessive nested payload size
- Preserve metadata and optionally retain raw JSON for audit/debug workflows

Provide:
- API endpoint and parameters
- Incremental logic and state fields
- Schema definition with key fields and types
- Flattening rules for address, shipping, invoice_settings, and metadata
- Sample JSON response shape
- Edge cases for deleted customers, missing email, tax fields, and nested payment sources
- Logging statements for observability

Respect MantrixFlow 5-phase ELT, upsert-only delivery, and high observability standards.
```

### Subscriptions Stream

```text
Create detailed specifications for the Stripe "subscriptions" stream connector.

Context:
- Incremental mode should use created >= last_sync and consider status transition timestamps where available
- Fields to extract: id, customer, status, created, start_date, current_period_start, current_period_end, canceled_at, cancel_at_period_end, ended_at, trial_start, trial_end, collection_method, currency, metadata, latest_invoice, default_payment_method, items
- Expand customer, latest_invoice, and items.data.price/product only when useful
- Preserve plan/price relationships in a predictable flattened shape

Provide:
- API endpoint and parameters
- Incremental logic and state fields
- Schema definition with key fields and types
- Flattening rules for items, prices, products, and metadata
- Sample JSON response shape
- Edge cases for canceled, paused, trialing, incomplete, and multi-item subscriptions
- Logging statements for observability

Respect MantrixFlow 5-phase ELT, upsert-only delivery, and high observability standards.
```

### Invoices Stream

```text
Create detailed specifications for the Stripe "invoices" stream connector.

Context:
- Incremental mode should use created >= last_sync, with care for mutable invoice status fields
- Fields to extract: id, customer, subscription, status, created, due_date, period_start, period_end, amount_due, amount_paid, amount_remaining, subtotal, total, currency, paid, attempt_count, hosted_invoice_url, invoice_pdf, metadata, lines
- Expand customer, subscription, payment_intent, charge, and lines only when useful
- Avoid exploding line items unless the design explicitly creates a child stream/table

Provide:
- API endpoint and parameters
- Incremental logic and state fields
- Schema definition with key fields and types
- Flattening rules for lines, customer, subscription, and metadata
- Sample JSON response shape
- Edge cases for draft, void, uncollectible, paid, partially paid, and deleted/nullable linked objects
- Logging statements for observability

Respect MantrixFlow 5-phase ELT, upsert-only delivery, and high observability standards.
```

### Refunds Stream

```text
Create detailed specifications for the Stripe "refunds" stream connector.

Context:
- Incremental mode should use created >= last_sync
- Fields to extract: id, charge, payment_intent, amount, currency, status, reason, receipt_number, created, balance_transaction, metadata, destination_details, failure_balance_transaction, failure_reason
- Expand charge and balance_transaction only when useful
- Preserve refund-to-charge lineage for downstream reconciliation

Provide:
- API endpoint and parameters
- Incremental logic and state fields
- Schema definition with key fields and types
- Flattening rules for destination_details, linked charge fields, and metadata
- Sample JSON response shape
- Edge cases for failed, pending, canceled, partial, and multi-refund charges
- Logging statements for observability

Respect MantrixFlow 5-phase ELT, upsert-only delivery, and high observability standards.
```

### Balance Transactions Stream

```text
Create detailed specifications for the Stripe "balance_transactions" stream connector.

Context:
- Incremental mode should use created >= last_sync
- Fields to extract: id, amount, fee, net, currency, type, reporting_category, source, status, created, available_on, description, exchange_rate, fee_details
- Support reconciliation use cases that join charges, refunds, payouts, disputes, and fees
- Preserve fee_details either as JSON or a child stream/table depending on destination needs

Provide:
- API endpoint and parameters
- Incremental logic and state fields
- Schema definition with key fields and types
- Flattening rules for fee_details and source references
- Sample JSON response shape
- Edge cases for pending availability, currency conversion, negative amounts, and fee-only transactions
- Logging statements for observability

Respect MantrixFlow 5-phase ELT, upsert-only delivery, and high observability standards.
```

### Events Stream

```text
Create detailed specifications for the Stripe "events" stream connector.

Context:
- Incremental mode should use created >= last_sync
- Fields to extract: id, type, api_version, created, livemode, pending_webhooks, request, data, account
- Treat data.object as raw JSON by default and flatten only safe common fields such as object id and object type
- Support event replay/audit use cases without leaking secrets or overly nested payloads into user-facing logs

Provide:
- API endpoint and parameters
- Incremental logic and state fields
- Schema definition with key fields and types
- Flattening rules for data.object, request, and account
- Sample JSON response shape
- Edge cases for old api_version events, redacted objects, deleted objects, and high-volume event backfills
- Logging statements for observability

Respect MantrixFlow 5-phase ELT, upsert-only delivery, and high observability standards.
```

### Payouts Stream

```text
Create detailed specifications for the Stripe "payouts" stream connector.

Context:
- Incremental mode should use created >= last_sync and optionally arrival_date for operational reporting
- Fields to extract: id, amount, currency, status, created, arrival_date, automatic, method, type, description, destination, balance_transaction, failure_balance_transaction, failure_code, failure_message, metadata
- Expand balance_transaction only when useful
- Preserve payout-to-balance-transaction lineage for reconciliation

Provide:
- API endpoint and parameters
- Incremental logic and state fields
- Schema definition with key fields and types
- Flattening rules for destination, balance transaction references, and metadata
- Sample JSON response shape
- Edge cases for failed, canceled, in_transit, paid, delayed, and manual payouts
- Logging statements for observability

Respect MantrixFlow 5-phase ELT, upsert-only delivery, and high observability standards.
```

### Products Stream

```text
Create detailed specifications for the Stripe "products" stream connector.

Context:
- Incremental mode should use created >= last_sync, with full refresh fallback for mutable catalog fields
- Fields to extract: id, name, active, description, created, updated, default_price, livemode, metadata, images, package_dimensions, shippable, statement_descriptor, tax_code, type, unit_label, url
- Expand default_price only when useful
- Preserve catalog lineage to prices and subscriptions

Provide:
- API endpoint and parameters
- Incremental logic and state fields
- Schema definition with key fields and types
- Flattening rules for package_dimensions, images, default_price, and metadata
- Sample JSON response shape
- Edge cases for archived products, missing default prices, deleted linked prices, and mutable names/descriptions
- Logging statements for observability

Respect MantrixFlow 5-phase ELT, upsert-only delivery, and high observability standards.
```

### Prices Stream

```text
Create detailed specifications for the Stripe "prices" stream connector.

Context:
- Incremental mode should use created >= last_sync, with full refresh fallback for mutable active/lookup fields
- Fields to extract: id, product, active, billing_scheme, created, currency, custom_unit_amount, livemode, lookup_key, metadata, nickname, recurring, tax_behavior, tiers_mode, transform_quantity, type, unit_amount, unit_amount_decimal
- Expand product only when useful
- Preserve recurring interval, usage_type, and trial configuration for subscription analytics

Provide:
- API endpoint and parameters
- Incremental logic and state fields
- Schema definition with key fields and types
- Flattening rules for recurring, custom_unit_amount, transform_quantity, tiers, product, and metadata
- Sample JSON response shape
- Edge cases for one-time prices, recurring prices, tiered pricing, archived prices, and currency precision
- Logging statements for observability

Respect MantrixFlow 5-phase ELT, upsert-only delivery, and high observability standards.
```

---

## 3. Preflight Phase Prompt

```text
Design Preflight Phase (Phase 0) for the Stripe source in MantrixFlow.

Must check:
- API key validity and whether the key is live or test mode
- Required permissions and resource access for selected streams
- Account information, including account id and livemode
- Rate limit and retry readiness where observable
- Ability to list each selected resource with a minimal request
- Network connectivity and timeout behavior
- Destination tables already exist and match dbt model output columns
- Disk budget is available before starting DuckDB staging

Return clear, user-friendly error messages that explain exactly what is wrong and how to fix it.
Output format should match MantrixFlow's existing preflight style.

Respect MantrixFlow 5-phase ELT, upsert-only delivery, and high observability standards.
```

---

## 4. Error Handling and Observability Prompt

```text
Define comprehensive error handling and logging strategy for the Stripe connector.

Cover:
- Authentication errors
- Permission or resource access errors
- Rate limit errors with bounded exponential backoff and retry-after support
- Invalid request errors
- Network timeouts and transient Stripe API failures
- Object not found, deleted, or redacted responses
- Schema drift detection
- Data type conversion issues
- Destination schema mismatch discovered during preflight

For each error type, specify:
- Log level
- User-facing message
- Action taken: retry, skip, or fail
- Metadata to store for debugging without storing secrets

Respect MantrixFlow 5-phase ELT, upsert-only delivery, and high observability standards.
```

---

## 5. Schema and Normalization Prompt

```text
Design schema discovery and normalization strategy for Stripe objects in MantrixFlow.

Rules:
- Flatten nested objects reasonably, with a maximum of 2-3 levels by default
- Handle Stripe metadata as JSON plus optional flattened metadata fields
- Convert timestamps to proper datetime fields where the destination model expects datetime, while preserving original epoch values when useful
- Keep raw JSON version as optional _raw column during staging only when explicitly enabled
- Support user-defined field selection
- Avoid sending _dlt_* tables to the client destination

Output recommended flattened schema for:
- Charges
- Customers
- Subscriptions
- Invoices
- Refunds
- Balance transactions
- Events
- Payouts
- Products
- Prices

Respect MantrixFlow 5-phase ELT, upsert-only delivery, and high observability standards.
```

---

## 6. Incremental State Management Prompt

```text
Design incremental state management for the Stripe connector in MantrixFlow.

Requirements:
- Store last_sync timestamp per stream
- Handle both created and updated fields where the Stripe object supports them
- Support bookmark or cursor method for paginated reads
- Make state persistent and recoverable
- Handle backfill scenarios with explicit start and end bounds
- Provide a clear state reset option
- Include extracted checkpoint state in the callback before DuckDB cleanup

Provide detailed logic and edge case handling.

Respect MantrixFlow 5-phase ELT, upsert-only delivery, and high observability standards.
```

---

## 7. Testing Strategy Prompt

```text
Create a test strategy for the Stripe connector in MantrixFlow.

Cover:
- Unit tests for stream config parsing, Stripe credential validation, pagination, and state updates
- Preflight tests for invalid keys, publishable keys, inaccessible resources, destination table missing, and column mismatch
- Integration tests against Stripe test-mode data using safe test keys only
- End-to-end pipeline tests for full-table and incremental sync into existing destination tables
- Regression tests for strict ELT invariants: upsert-only delivery, no destination CREATE TABLE, no credential leakage, state extraction before DuckDB cleanup, and required callback metadata
- Manual testing scenarios for charges, customers, invoices, subscriptions, events, and revenue reconciliation

Return a prioritized test matrix with expected outcomes and observability checks.

Respect MantrixFlow 5-phase ELT, upsert-only delivery, and high observability standards.
```

---

## How to Use This Prompt Pack

1. Start with the master prompt.
2. Use the preflight and error-handling prompts to define guardrails.
3. Build or review streams one by one using the individual stream prompts.
4. Finish with schema/normalization, incremental state, and testing prompts.
5. Validate any resulting implementation against the strict ELT invariants before merging.

Always end connector implementation prompts with:

```text
Respect MantrixFlow 5-phase ELT, upsert-only delivery, and high observability standards.
```
