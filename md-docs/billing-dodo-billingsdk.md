# MantrixFlow Billing: Dodo Hosted Checkout + Billing UI

This guide documents the current MVP billing architecture.

- Plans: `starter`, `growth`, `pro`, `enterprise`
- Self-serve paid plans: Growth and Pro
- Enterprise: coming soon, not self-serve
- New paid purchase: Dodo hosted checkout
- Paid upgrade: Dodo hosted checkout creates a replacement paid subscription; webhook activation supersedes the old subscription
- Downgrade: scheduled locally for the next billing cycle
- Payment methods and hosted receipts: Dodo customer portal
- Invoice list: MantrixFlow API sync/cache of Dodo invoice/payment rows
- Frontend: local shadcn/BillingSDK-style components under `apps/app/components/billingsdk`

No Solo tier, product collection checkout, API-call billing, or user-facing cancel button is part of this MVP model.

## Backend API

Authenticated organization owner routes:

| Route | Purpose |
| --- | --- |
| `GET /api/v1/organizations/:id/billing/subscription` | Current plan, Dodo subscription id, pending downgrade, and period dates |
| `POST /api/v1/organizations/:id/billing/checkout` | Starts hosted checkout for first paid subscriptions and paid upgrades, or schedules downgrades |
| `POST /api/v1/organizations/:id/billing/portal` | Opens Dodo-hosted payment method and receipt management |
| `DELETE /api/v1/organizations/:id/billing/scheduled-change` | Cancels a pending scheduled downgrade |
| `GET /api/v1/organizations/:id/billing/invoices` | Lists Dodo-backed invoice/payment history |
| `GET /api/v1/organizations/:id/billing/upcoming-charges` | Current base subscription estimate only |

Webhook route:

| Route | Purpose |
| --- | --- |
| `POST /api/v1/billing/webhook` | Dodo webhook receiver. Webhooks are the billing source of truth. |

## Checkout And Plan Changes

### Starter to Growth or Pro

1. User selects Growth or Pro in the billing page.
2. App calls `POST /billing/checkout`.
3. API creates a `billing_checkout_intents` row.
4. API creates a Dodo hosted checkout session for the target product.
5. User pays in Dodo.
6. Dodo webhook activates the organization plan.

### Growth to Pro

Dodo can reject hosted checkout for an existing customer that already has an active subscription unless multiple subscriptions per customer are enabled. MantrixFlow still keeps upgrades on hosted checkout by creating an upgrade checkout intent and a replacement checkout customer for the session. Intent metadata links the new subscription back to the organization and the old subscription.

For Growth to Pro:

1. App calls `POST /billing/checkout` with `plan=pro`.
2. API creates an upgrade checkout intent with `replace_subscription_id`.
3. API creates a Dodo hosted checkout session for the Pro product.
4. User pays in Dodo hosted checkout.
5. Dodo webhook activates Pro in MantrixFlow.
6. MantrixFlow supersedes the old Growth subscription so duplicate renewals cannot happen.

The redirect URL is not the billing source of truth. Only Dodo webhooks complete the upgrade.

### Pro to Growth

Downgrades are not applied immediately.

1. App calls `POST /billing/checkout` with the lower plan.
2. API creates a `downgrade_renewal` intent.
3. Current paid access remains active until `plan_expires_at`.
4. The UI shows the pending downgrade and effective date.
5. At renewal, the account either completes the lower paid plan flow or falls back to Starter if no valid paid renewal exists.

### Starter fallback

Users do not manually switch to Starter. Starter is an automatic fallback after the current paid period ends when payment cannot be collected, payment method is removed, or a paid renewal is not completed.

## Invoice And Portal Rules

- MantrixFlow lists invoice/payment rows in the billing page.
- "View invoice" opens the hosted Dodo invoice/receipt URL when Dodo provides one.
- "Manage billing" opens the Dodo customer portal.
- The app does not expose a manual cancel subscription button.

## Billing UI Components

The app uses local, adapted BillingSDK-style components. Do not run the full BillingSDK initializer.

```text
apps/app/components/billingsdk/update-plan-card.tsx
apps/app/components/billingsdk/usage-meter-circle.tsx
apps/app/components/billingsdk/upcoming-charges.tsx
apps/app/components/billingsdk/invoice-history.tsx
apps/app/components/billingsdk/trial-expiry-card.tsx
```

Browser code must call MantrixFlow API hooks only. It must never call Dodo directly.

## Sandbox Test Checklist

1. Starter to Growth opens Dodo hosted checkout and webhook activates Growth.
2. Growth to Pro opens Dodo hosted checkout; webhook activates Pro and supersedes the old Growth subscription.
3. Pro to Growth schedules a downgrade at next billing date.
4. Pending downgrade appears in the billing page with the effective date.
5. Enterprise is disabled and shows Coming soon.
6. Starter action is disabled for paid users and described as automatic fallback only.
7. Manage billing opens the Dodo customer portal.
8. Invoice history lists Dodo-backed rows and hosted invoice links.
9. Duplicate Dodo webhook events do not double-apply plan or invoice changes.
