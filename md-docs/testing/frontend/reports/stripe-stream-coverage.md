# Stripe All-Stream Chrome E2E Coverage

Pipeline: `Stripe All Streams` (`397fc4bf-3021-44bf-b2e2-18fa1e9b867d`)

Final result: **34/34 selected streams completed successfully**. The latest UI run delivered one representative account row in 38 seconds. The pipeline, connection, model, destination table, data, and run history remain retained.

## Incremental streams

The following 19 streams used their discovered `created` cursor:

`balance_transactions`, `charges`, `customers`, `disputes`, `events`, `payment_intents`, `payouts`, `refunds`, `setup_intents`, `subscriptions`, `invoices`, `invoice_items`, `credit_notes`, `quotes`, `checkout_sessions`, `reviews`, `early_fraud_warnings`, `files`, `file_links`

## Full-table streams

The following 15 cursorless streams used full-table mode:

`account`, `payment_methods`, `subscription_items`, `credit_note_line_items`, `products`, `prices`, `plans`, `coupons`, `promotion_codes`, `tax_rates`, `quote_line_items`, `checkout_session_line_items`, `payment_links`, `webhook_endpoints`, `tax_ids`

## Executed stream inventory

`account`, `balance_transactions`, `charges`, `customers`, `disputes`, `events`, `payment_intents`, `payment_methods`, `payouts`, `refunds`, `setup_intents`, `subscriptions`, `subscription_items`, `invoices`, `invoice_items`, `credit_notes`, `credit_note_line_items`, `products`, `prices`, `plans`, `coupons`, `promotion_codes`, `tax_rates`, `quotes`, `quote_line_items`, `checkout_sessions`, `checkout_session_line_items`, `payment_links`, `reviews`, `early_fraud_warnings`, `files`, `file_links`, `webhook_endpoints`, `tax_ids`

## Runtime evidence

- All 34 resources emitted stream start/completion events.
- Empty streams completed normally and were materialized.
- Populated source streams included account, customers, events, payment intents, setup intents, invoices, products, prices, plans, payment links, and files.
- The published output model uses an explicit stable account projection to prevent dynamic Stripe fields from exceeding the pre-created destination contract.

No credential value is recorded here.
