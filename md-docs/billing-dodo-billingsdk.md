# MantrixFlow Billing: Dodo + BillingSDK UI

This guide documents the current billing architecture:

- Plans: `starter`, `growth`, `pro`, `enterprise`
- First paid purchase: Dodo hosted checkout
- Existing paid subscription changes: MantrixFlow API calls Dodo `change-plan`
- Payment method and receipts: Dodo customer portal
- Overage: local usage accounting plus manual cycle-end invoice ledger
- Frontend: local shadcn/BillingSDK-style components under `apps/app/components/billingsdk`

No Solo tier is part of this model.

## Backend API

Authenticated organization owner routes:

| Route | Purpose |
| --- | --- |
| `GET /api/v1/organizations/:id/billing/subscription` | Current plan, Dodo subscription id, pending change, overage setting |
| `POST /api/v1/organizations/:id/billing/checkout` | First paid subscription checkout for Growth or Pro |
| `POST /api/v1/organizations/:id/billing/change-plan` | Upgrade, downgrade, or monthly/annual switch for an existing subscription |
| `POST /api/v1/organizations/:id/billing/portal` | Dodo-hosted payment method, receipt, and invoice management |
| `POST /api/v1/organizations/:id/billing/cancel` | Cancel at period end |
| `DELETE /api/v1/organizations/:id/billing/scheduled-change` | Cancel a pending scheduled change |
| `PATCH /api/v1/organizations/:id/billing/overage-settings` | Enable or disable opt-in overage |
| `GET /api/v1/organizations/:id/billing/upcoming-charges` | Base subscription plus projected manual overage |
| `GET /api/v1/organizations/:id/billing/invoices` | Local overage invoice ledger |

Internal admin route:

| Route | Purpose |
| --- | --- |
| `POST /api/v1/internal/billing/overage-invoices/run` | Create one idempotent overage invoice for a billing cycle |

## Change-Plan Rules

Checkout is only for the first paid subscription. After an organization has a
`dodo_subscription_id`, plan changes go through `POST /billing/change-plan`.

- Growth to Pro: applied immediately through Dodo `change-plan`.
- Monthly to annual: handled by `change-plan`.
- Downgrades: scheduled for next billing date.
- Downgrade is blocked when current usage exceeds the target plan and overage is disabled.
- Enterprise is contact-sales only and is rejected by checkout/change-plan.

## Overage Rules

Overage is off by default.

When overage is off:

- Pipeline row pressure blocks new runs.
- API-call limits return `429`.
- Downgrades to a smaller plan are blocked if current usage exceeds the target limit.

When overage is on:

- Excess usage is allowed.
- Upcoming charges show projected manual overage.
- Cycle-end invoice generation creates one idempotent ledger row for the period.

Manual invoice math:

| Resource | Included Source | Overage Price |
| --- | --- | --- |
| Rows | Delivered Phase 3 rows in billing-cycle window | `$5 / 1M rows` |
| API calls | Local API-call logs, over daily allowance | `$2 / 1M API calls` |

Example: Growth includes `1,000,000` rows. If a cycle delivers `2,300,000`
rows, billable overage is `1,300,000` rows:

```text
1.3M rows * $5 / 1M = $6.50
```

The invoice job is idempotent on `(organization_id, period_start, period_end)`,
so reruns do not double-charge.

## BillingSDK UI

Do not run the full BillingSDK initializer. The app uses local components so the
dark MantrixFlow UI remains consistent:

```text
apps/app/components/billingsdk/update-plan-card.tsx
apps/app/components/billingsdk/cancel-subscription-card.tsx
apps/app/components/billingsdk/usage-meter-circle.tsx
apps/app/components/billingsdk/upcoming-charges.tsx
apps/app/components/billingsdk/invoice-history.tsx
apps/app/components/billingsdk/trial-expiry-card.tsx
```

Reference commands for future component refreshes:

```bash
npx @billingsdk/cli add update-plan-card
npx @billingsdk/cli add cancel-subscription-card
npx @billingsdk/cli add usage-meter-circle
npx @billingsdk/cli add upcoming-charges
npx @billingsdk/cli add invoice-history
npx @billingsdk/cli add trial-expiry-card
```

Use the generated source as reference only, then adapt into
`components/billingsdk/*`. Browser code must call MantrixFlow API hooks, never
Dodo directly.

## Sandbox Test Checklist

1. Starter to Growth checkout opens Dodo checkout and webhook activates Growth.
2. Growth to Pro calls `POST /billing/change-plan` and applies immediately.
3. Pro to Growth schedules a downgrade at next billing date.
4. Downgrade is blocked when usage exceeds target limits and overage is off.
5. Enable overage, exceed rows/API limits, and verify upcoming charges update.
6. Run the internal overage invoice job once and verify one ledger row.
7. Run the same invoice job again and verify no duplicate charge.
8. Cancel subscription and confirm the UI shows cancellation at period end.
9. Open Dodo portal and verify receipts/payment method access still works.
