# MantrixFlow Billing: Dodo Hosted Checkout + Billing UI

This guide documents the current MVP billing architecture.

- Plans: `free`, `plus`, `pro`, `enterprise`
- Self-serve paid plans: Plus and Pro
- Enterprise: coming soon, not self-serve
- New paid purchase: Dodo hosted checkout
- Paid upgrade: Dodo hosted checkout creates a replacement paid subscription; webhook activation supersedes the old subscription
- Downgrade: scheduled locally for the next billing cycle
- Payment methods and hosted receipts: Dodo customer portal
- Invoice list: MantrixFlow API sync/cache of Dodo invoice/payment rows
- Frontend: local shadcn/BillingSDK-style components under `apps/arcyria-platform/components/billingsdk`

No Solo tier, product collection checkout, usage-based add-on billing, upcoming local charges, or user-facing cancel button is part of this MVP model.

## Backend API

Authenticated organization owner routes:

| Route | Purpose |
| --- | --- |
| `GET /api/v1/organizations/:id/billing/subscription` | Current plan, Dodo subscription id, pending downgrade, and period dates |
| `POST /api/v1/organizations/:id/billing/checkout` | Starts hosted checkout for first paid subscriptions and paid upgrades, or schedules downgrades |
| `POST /api/v1/organizations/:id/billing/portal` | Opens Dodo-hosted payment method and receipt management |
| `DELETE /api/v1/organizations/:id/billing/scheduled-change` | Cancels a pending scheduled downgrade |
| `GET /api/v1/organizations/:id/billing/invoices` | Lists Dodo-backed invoice/payment history |

Webhook route:

| Route | Purpose |
| --- | --- |
| `POST /api/v1/billing/webhook` | Dodo webhook receiver. Webhooks are the billing source of truth. |

## Checkout And Plan Changes

### Free to Plus or Pro

1. User selects Plus or Pro in the billing page.
2. App calls `POST /billing/checkout`.
3. API creates a `billing_checkout_intents` row.
4. API creates a Dodo hosted checkout session for the target product.
5. User pays in Dodo.
6. Dodo webhook activates the organization plan.

### Plus to Pro

Dodo can reject hosted checkout for an existing customer that already has an active subscription unless multiple subscriptions per customer are enabled. MantrixFlow still keeps upgrades on hosted checkout by creating an upgrade checkout intent and a replacement checkout customer for the session. Intent metadata links the new subscription back to the organization and the old subscription.

For Plus to Pro:

1. App calls `POST /billing/checkout` with `plan=pro`.
2. API creates an upgrade checkout intent with `replace_subscription_id`.
3. API creates a Dodo hosted checkout session for the Pro product.
4. User pays in Dodo hosted checkout.
5. Dodo webhook activates Pro in MantrixFlow.
6. MantrixFlow supersedes the old Plus subscription so duplicate renewals cannot happen.

The redirect URL is not the billing source of truth. Only Dodo webhooks complete the upgrade.

### Pro to Plus

Downgrades are not applied immediately.

1. App calls `POST /billing/checkout` with the lower plan.
2. API creates a `downgrade_renewal` intent.
3. Current paid access remains active until `plan_expires_at`.
4. The UI shows the pending downgrade and effective date.
5. At renewal, the account either completes the lower paid plan flow or falls back to Free if no valid paid renewal exists.

### Free fallback

Users do not manually switch to Free. Free is an automatic fallback after the current paid period ends when payment cannot be collected, payment method is removed, or a paid renewal is not completed.

## Invoice And Portal Rules

- MantrixFlow lists invoice/payment rows in the billing page.
- "View invoice" opens the hosted Dodo invoice/receipt URL when Dodo provides one.
- "Manage billing" opens the Dodo customer portal.
- The app does not expose a manual cancel subscription button.

## Billing UI Components

The app uses local, adapted BillingSDK-style components. Do not run the full BillingSDK initializer.

```text
apps/arcyria-platform/components/billingsdk/update-plan-card.tsx
apps/arcyria-platform/components/billingsdk/usage-meter-circle.tsx
apps/arcyria-platform/components/billingsdk/invoice-history.tsx
apps/arcyria-platform/components/billingsdk/trial-expiry-card.tsx
```

Browser code must call MantrixFlow API hooks only. It must never call Dodo directly.

## Sandbox Test Checklist

1. Free to Plus opens Dodo hosted checkout and webhook activates Plus.
2. Plus to Pro opens Dodo hosted checkout; webhook activates Pro and supersedes the old Plus subscription.
3. Pro to Plus schedules a downgrade at next billing date.
4. Pending downgrade appears in the billing page with the effective date.
5. Enterprise is disabled and shows Coming soon.
6. Free action is disabled for paid users and described as automatic fallback only.
7. Manage billing opens the Dodo customer portal.
8. Invoice history lists Dodo-backed rows and hosted invoice links.
9. Duplicate Dodo webhook events do not double-apply plan or invoice changes.
