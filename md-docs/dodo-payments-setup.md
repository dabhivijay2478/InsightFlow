# Dodo Payments Setup For MantrixFlow

This guide covers the current MVP billing setup.

## Current Billing Model

- Plans: Starter, Growth, Pro, Enterprise
- Starter: free fallback plan
- Growth and Pro: self-serve paid subscriptions
- Enterprise: coming soon, not self-serve
- Checkout: Dodo hosted checkout for first paid purchases and paid upgrades
- Paid upgrades: hosted checkout creates a replacement subscription; webhook activation supersedes the old subscription
- Downgrades: scheduled for the next billing cycle
- Portal: Dodo customer portal for payment methods and hosted receipts
- Invoices: listed in MantrixFlow from Dodo-backed invoice/payment rows

Do not configure or use product collection checkout for the MVP flow. Do not expose a manual cancel button in the app. API-call billing is postponed until developer/API features exist.

## Dodo Dashboard Setup

Create these subscription products in Dodo:

| Product | Interval | Env var |
| --- | --- | --- |
| MantrixFlow Growth Monthly | Monthly | `DODO_PRODUCT_GROWTH_MONTHLY` |
| MantrixFlow Growth Annual | Annual | `DODO_PRODUCT_GROWTH_ANNUAL` |
| MantrixFlow Pro Monthly | Monthly | `DODO_PRODUCT_PRO_MONTHLY` |
| MantrixFlow Pro Annual | Annual | `DODO_PRODUCT_PRO_ANNUAL` |

Enterprise is not self-serve, so it does not need a checkout product for this MVP.

## Environment Variables

Required backend variables:

```env
DODO_PAYMENTS_API_KEY=
DODO_WEBHOOK_SECRET=
DODO_PRODUCT_GROWTH_MONTHLY=
DODO_PRODUCT_GROWTH_ANNUAL=
DODO_PRODUCT_PRO_MONTHLY=
DODO_PRODUCT_PRO_ANNUAL=
API_PUBLIC_URL=https://cloud.api.mantrixflow.com
APP_WEB_URL=https://cloud.mantrixflow.com
```

Do not set `DODO_PRODUCT_COLLECTION_ID` for the current flow.

Legacy metering variables such as `DODO_API_CALLS_USAGE_BILLING_ENABLED`, `DODO_API_CALLS_EVENT_NAME`, and `DODO_API_CALLS_METER_ID` should remain unset or disabled.

## Webhooks

Configure the Dodo webhook endpoint:

```text
https://cloud.api.mantrixflow.com/api/v1/billing/webhook
```

The webhook handler verifies Standard Webhooks signatures and reconciles subscription/payment state into Postgres.

Expected lifecycle handling:

- Successful first checkout activates Growth or Pro.
- Successful paid upgrade activates the replacement subscription and supersedes the old subscription.
- Failed renewal or expired subscription keeps paid access until the current period ends, then falls back to Starter.
- Duplicate webhook delivery is idempotent.

## App Billing Flow

### Starter To Growth Or Pro

1. User clicks Growth or Pro in Settings -> Billing.
2. App calls `POST /api/v1/organizations/:orgId/billing/checkout`.
3. API creates a checkout intent and Dodo hosted checkout session.
4. User completes checkout in Dodo.
5. Dodo webhook activates the paid plan.

### Growth To Pro

Dodo can reject a second subscription for the same customer with:

```text
CUSTOMER_HAS_EXISTING_SUBSCRIPTION
```

MantrixFlow still uses hosted checkout for paid upgrades. The API creates an upgrade intent and a replacement checkout customer for that Dodo session, then links the new subscription back to the organization through checkout metadata.

1. User clicks Pro.
2. App calls `POST /billing/checkout`.
3. API creates an upgrade intent with the old subscription ID.
4. API creates a Dodo hosted checkout session for Pro.
5. User completes payment in Dodo.
6. Dodo webhook activates Pro in MantrixFlow.
7. MantrixFlow supersedes the old Growth subscription to prevent duplicate renewal.

### Pro To Growth

Downgrades are scheduled and do not reduce access immediately.

1. User clicks Growth.
2. API creates a `downgrade_renewal` checkout intent.
3. Current Pro access remains active until `plan_expires_at`.
4. Billing UI shows Growth starts at the next billing cycle.
5. If renewal is not completed, the organization falls back to Starter.

### Starter Fallback

Users do not manually downgrade to Starter. Starter is automatic when payment cannot be collected or a paid renewal expires after the current paid period.

## API Reference

| Method | Route | Purpose |
| --- | --- | --- |
| `GET` | `/api/v1/organizations/:orgId/billing/subscription` | Current plan, period, subscription id, and pending downgrade |
| `POST` | `/api/v1/organizations/:orgId/billing/checkout` | Hosted checkout for first paid purchase, paid upgrade, or scheduled downgrade |
| `POST` | `/api/v1/organizations/:orgId/billing/portal` | Dodo customer portal |
| `GET` | `/api/v1/organizations/:orgId/billing/invoices` | Dodo-backed invoice/payment list |
| `GET` | `/api/v1/organizations/:orgId/billing/upcoming-charges` | Base subscription estimate |
| `DELETE` | `/api/v1/organizations/:orgId/billing/scheduled-change` | Cancel pending scheduled downgrade |
| `POST` | `/api/v1/billing/webhook` | Public Dodo webhook endpoint |

## Verification Checklist

1. Starter to Growth opens hosted checkout.
2. Dodo webhook activates Growth.
3. Growth to Pro opens hosted checkout and webhook activation supersedes the old subscription.
4. Pro to Growth schedules the downgrade and keeps Pro active until period end.
5. Enterprise shows Coming soon and cannot be selected.
6. Starter selection is disabled for paid users and described as automatic fallback.
7. Manage billing opens the Dodo portal.
8. Invoice history lists Dodo-backed rows and hosted invoice links.
9. Failed/expired renewal falls back to Starter only after the current period ends.
