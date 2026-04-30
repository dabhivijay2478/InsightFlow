---
name: dodo-payments
description: Use when building or reviewing Dodo Payments integrations across checkout, subscriptions, usage-based billing, seat-based billing, credit-based billing, BillingSDK UI, webhooks, and license keys.
---

# Dodo Payments

Use this skill for broad Dodo Payments implementation work. For detailed flows, prefer the narrower skills in this plugin:

- `checkout-integration`: hosted checkout, overlay checkout, payment links, customer handoff.
- `subscription-integration`: recurring products, trials, upgrades, downgrades, proration, on-demand charges.
- `usage-based-billing`: meters, events, aggregation, per-unit pricing, free thresholds.
- `seat-based-billing`: per-user pricing, add-ons, seat changes, proration, seat enforcement.
- `credit-based-billing`: credit entitlements, balances, rollover, expiration, overage, meter-based deduction.
- `webhook-integration`: event verification and payment, subscription, credit, refund, dispute, and license events.
- `billing-sdk`: React billing UI components and project setup.
- `license-keys`: software license keys, activation limits, validation, and access control.

Always check the latest official docs before implementing exact API shapes:

- Main site: https://dodopayments.com/
- Docs: https://docs.dodopayments.com/
- Dashboard: https://app.dodopayments.com/
- Agent skills: https://docs.dodopayments.com/developer-resources/agent-skills

This plugin also includes Dodo's documented MCP endpoints:

- `dodo-knowledge`: documentation search over Dodo Payments docs.
- `dodopayments`: Dodo Payments API operations through MCP.

## Implementation Posture

Treat Dodo Payments as the billing source of truth for payment state, subscription state, product catalog, credit entitlements, and webhook-delivered lifecycle events. Treat the application database as the product-access projection that is updated from verified checkout returns, API reads, and webhooks.

For SaaS and AI products, design around these stable boundaries:

- Product setup lives in Dodo: products, prices, add-ons, credit entitlements, meters, and license key settings.
- Checkout starts server-side where API keys are protected.
- Webhooks are required for durable provisioning and deprovisioning.
- Application access checks should read local projected state, not call billing APIs on every request.
- Usage and credit events need idempotency keys so retries do not double-charge or double-deduct.
- Customer-facing plan changes should preview charges or credits before mutation when Dodo supports previewing.

## Common Architecture

1. Create products and billing objects in the Dodo dashboard.
2. Store product IDs, add-on IDs, meter names, and entitlement IDs in app config.
3. Create checkout sessions from backend routes only.
4. On checkout return, show pending or success UI, but do not grant durable access solely from the return URL.
5. Verify Dodo webhooks and update local billing tables idempotently.
6. Gate app features from local entitlement, subscription, credit, seat, or license state.
7. Reconcile periodically by reading Dodo APIs for subscriptions, customers, and payment state.

## Billing Model Picker

Use subscriptions when customers buy recurring access to plans or memberships.

Use usage-based billing when the invoice should reflect metered activity such as API calls, tokens, storage, messages, or compute.

Use seat-based billing when price scales by users, team members, hosts, editors, or licenses. Model extra seats as add-ons and enforce seat limits inside the application.

Use credit-based billing when customers receive a prepaid or recurring balance, then consume credits through usage events or manual deductions.

Use license keys when access needs portable activation, device limits, downloads, plugins, CLIs, or offline-ish software authorization.

## Security Checklist

- Keep `DODO_PAYMENTS_API_KEY` server-side only.
- Store `DODO_PAYMENTS_WEBHOOK_SECRET` separately from the API key.
- Verify webhook signatures before reading event data.
- Make webhook processing idempotent by event ID.
- Never trust client-submitted prices, product IDs, credit amounts, or seat counts without server-side validation.
- Log billing state transitions with enough context to audit access changes.
