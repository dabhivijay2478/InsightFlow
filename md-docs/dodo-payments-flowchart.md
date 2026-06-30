# Dodo Payments and MantrixFlow Flow Charts

Visual map of how Dodo Payments connects to the MantrixFlow stack. The current architecture is documented in [`billing-dodo-billingsdk.md`](./billing-dodo-billingsdk.md).

## Facts This Diagram Reflects

- The Next.js app talks only to the Go API for billing.
- Free to paid uses Dodo hosted checkout.
- Paid upgrades use Dodo hosted checkout for a replacement subscription, then webhook reconciliation supersedes the old subscription.
- Downgrades are scheduled for the next billing cycle.
- The Dodo portal is used for payment methods and hosted receipts.
- Webhooks reconcile the database and are the billing source of truth.

## End-to-End System View

```mermaid
flowchart TB
  subgraph App["Next.js app"]
    UI["Settings / Billing"]
  end

  subgraph API["Go API"]
    CO["POST .../billing/checkout"]
    PO["POST .../billing/portal"]
    SUB["GET .../billing/subscription"]
    INV["GET .../billing/invoices"]
    WH["POST /api/v1/billing/webhook"]
    DB[("Postgres billing state")]
  end

  subgraph Dodo["Dodo Payments"]
    HC["Hosted checkout"]
    UP["Supersede old subscription"]
    CP["Customer portal"]
    EV["subscription/payment events"]
  end

  UI --> CO
  UI --> PO
  UI --> SUB
  UI --> INV

  CO -->|"Free to paid"| HC
  CO -->|"Paid upgrade"| HC
  CO -->|"Downgrade"| DB
  PO --> CP

  HC --> EV
  WH -->|"Upgrade cleanup"| UP
  EV -->|"HMAC verified"| WH
  WH --> DB
  DB --> SUB
  DB --> INV
```

## Checkout Decision

```mermaid
flowchart LR
  A["User selects a plan"] --> B{"Current plan"}
  B -->|"Free"| C["Create hosted checkout"]
  B -->|"Plus, selects Pro"| D["Create hosted checkout for replacement subscription"]
  B -->|"Pro, selects Plus"| E["Schedule downgrade for period end"]
  B -->|"Enterprise"| F["Disabled: coming soon"]
  C --> G["Dodo webhook activates plan"]
  D --> G
  G --> I["Supersede old paid subscription"]
  E --> H["Current plan stays active until period end"]
```

## Quick Reference

| Scenario | Mechanism | When organization plan changes |
| --- | --- | --- |
| Free to Plus/Pro | Hosted checkout | After Dodo success webhook |
| Plus to Pro | Hosted checkout replacement subscription | After Dodo success webhook; old subscription is superseded |
| Pro to Plus | Local scheduled downgrade | At next billing cycle |
| Payment method or invoice view | Dodo customer portal | No plan change unless Dodo emits a lifecycle event |
