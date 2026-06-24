# Dodo Payments × MantrixFlow — flow charts

Visual map of how **Dodo Payments** connects to the MantrixFlow stack. The current architecture is documented in [`billing-dodo-billingsdk.md`](./billing-dodo-billingsdk.md).

**Facts this diagram reflects**

- The **Next.js app** (`apps/app`) talks only to the **Go API** for billing—never to Dodo directly for authenticated flows.
- **First paid purchase:** `POST …/billing/checkout` → Dodo **hosted checkout**.
- **Existing subscription:** plan / interval changes happen through `POST …/billing/change-plan`, which calls Dodo server-side.
- **Portal:** `POST …/billing/portal` is for payment methods, receipts, and hosted invoice access.
- **Source of truth** for `organizations.plan` (and related fields) after Dodo events: **webhooks** (`POST /api/v1/billing/webhook`) plus optional reconciliation when the app calls `GET …/billing/subscription`.

---

## End-to-end (system view)

```mermaid
flowchart TB
  subgraph App["Next.js app (apps/app)"]
    UI[Settings / Billing / Analytics]
  end

  subgraph Go["Go API (apps/server/main-server)"]
    JWT[JWT + org role]
    CO[POST .../billing/checkout]
    CH[POST .../billing/change-plan]
    PO[POST .../billing/portal]
    SUB[GET .../billing/subscription]
    WH[POST /api/v1/billing/webhook]
    DB[(Postgres: organizations.plan, dodo_* ids)]
  end

  subgraph Dodo["Dodo Payments"]
    HC[Hosted checkout]
    CP[Customer portal]
    EV[Events: subscription.*, payment.*]
  end

  UI --> JWT
  JWT --> CO
  JWT --> CH
  JWT --> PO
  JWT --> SUB

  CO -->|first purchase only| HC
  HC -->|redirect + pay| Dodo
  CH -->|server-side subscription change| Dodo
  PO -->|session URL| CP
  CP -->|payment method / receipts| Dodo

  EV -->|HMAC verified| WH
  WH --> DB
  SUB --> DB
  SUB -.->|optional live enrich| Dodo

  DB --> UI
```

---

## Checkout vs change-plan vs portal

```mermaid
flowchart LR
  A[Org needs paid plan change] --> B{dodo_subscription_id set?}
  B -->|No| C[POST /billing/checkout]
  B -->|Yes| D[POST /billing/change-plan]
  C --> E[Dodo hosted checkout]
  D --> F[Dodo change-plan API]
  E --> G[Webhooks update org]
  F --> G
  H[Payment method or receipts] --> I[POST /billing/portal]
  I --> J[Dodo customer portal]
```

---

## Webhook reconciliation (sequence)

```mermaid
sequenceDiagram
  participant Dodo
  participant API as Go API /billing/webhook
  participant DB as Postgres

  Dodo->>API: subscription.active / plan_changed / payment.* / cancelled …
  API->>API: Verify Standard Webhooks signature
  API->>DB: Update plan, plan_period, plan_status, dodo ids, dates
  Note over API, App: Client refetches GET /billing/subscription and usage
```

---

## Quick reference

| Path | Mechanism | When plan row updates in DB |
|------|-----------|------------------------------|
| First subscription | Checkout session → pay on Dodo | After webhooks (and GET subscription may enrich) |
| Upgrade / downgrade / monthly ↔ annual | MantrixFlow API → Dodo change-plan | Immediately for upgrades; scheduled locally for downgrades; webhooks reconcile |
| Payment method, invoices | Portal | Profile in Dodo; org plan unchanged unless an event fires |
