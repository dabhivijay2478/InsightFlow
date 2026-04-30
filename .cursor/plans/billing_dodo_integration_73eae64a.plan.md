---
name: Billing Dodo Integration
overview: Fix the row-count bug (use Phase 3 delivered rows, not Phase 1 staged rows) and implement full Dodo Payments billing integration — checkout, webhooks, customer portal, subscription lifecycle (upgrade/downgrade), and a wired-up frontend billing page.
todos:
  - id: install-sdk
    content: "Install dodopayments-go SDK: go get github.com/dodopayments/dodopayments-go"
    status: completed
  - id: models-usage
    content: Add CountSource field + UsageEventTypeRowDelivered constant to models/usage.go
    status: completed
  - id: models-org
    content: Add Dodo billing fields (DodoCustomerID, DodoSubID, PlanPeriod, PlanStatus, PlanStartedAt, PlanExpiresAt) to models/organization.go
    status: completed
  - id: config-dodo
    content: Add Dodo env vars to config struct (Config) and Load() in config/config.go
    status: completed
  - id: billing-service
    content: Create internal/services/billing/dodo_client.go and billing_service.go with CreateOrGetCustomer, CreateCheckoutSession, CreatePortalSession, IngestRowUsage, UpdateOrgPlan
    status: completed
  - id: fix-callback
    content: "Fix callback.go: use phase3_rows_delivered instead of rowsUpserted for usage recording; call Billing.IngestRowUsage on success"
    status: completed
  - id: usage-service
    content: Add RecordDeliveredRows to usage_service.go; update sumUsage to filter count_source=phase3_delivered for row_delivered events
    status: completed
  - id: billing-http
    content: Create internal/server/billing_http.go with checkout, webhook, subscription, portal, cancel handlers including HMAC signature verification
    status: completed
  - id: wire-state-routes
    content: Add Billing field to state.go + NewState(); mount billing routes in routes.go
    status: completed
  - id: frontend-service-hooks
    content: Create lib/api/services/billing.service.ts, lib/api/hooks/use-billing.ts, lib/api/types/billing.ts
    status: completed
  - id: frontend-billing-tab
    content: "Update settings/page.tsx billing tab: upgrade button, manage billing button, cancel button, past-due banner, plan_period/plan_expires_at in stat cards"
    status: completed
  - id: frontend-success-page
    content: Create app/workspace/billing/success/page.tsx as post-checkout landing page
    status: completed
isProject: false
---

# Billing — Dodo Payments + Row-Count Bug Fix

## Current state

- `internal/server/callback.go` calls `s.Usage.RecordRows(... rowsUpserted)` where `rowsUpserted` = Phase 1 `rows_written` from dlt — **wrong source**.
- `organizations` model has `StripeCustomerID` / `StripeSubID` columns and no Dodo fields.
- `models.UsageEvent` has no `count_source` column.
- No Dodo client, no billing routes, no checkout or webhook handler exist.
- Frontend Billing tab renders usage bars but has no upgrade/portal/past-due UI.

---

## Files to create / modify

### Go API

**1. `internal/config/config.go`** — add Dodo fields to `Config` struct and `Load()`:
```go
DodoAPIKey             string
DodoWebhookSecret      string
DodoProductGrowthMonthly  string
DodoProductGrowthAnnual   string
DodoProductProMonthly     string
DodoProductProAnnual      string
DodoProductEnterprise     string
```
Load from env: `DODO_PAYMENTS_API_KEY`, `DODO_WEBHOOK_SECRET`, `DODO_PRODUCT_*`.

**2. `internal/models/organization.go`** — add Dodo fields (replacing Stripe):
```go
DodoCustomerID    *string    `gorm:"column:dodo_customer_id;size:255"`
DodoSubID         *string    `gorm:"column:dodo_subscription_id;size:255"`
PlanPeriod        string     `gorm:"column:plan_period;default:'monthly'"`
PlanStatus        string     `gorm:"column:plan_status;default:'active'"`
PlanStartedAt     *time.Time `gorm:"column:plan_started_at"`
PlanExpiresAt     *time.Time `gorm:"column:plan_expires_at"`
```
Keep `StripeCustomerID`/`StripeSubID` for now (do not drop — may hold existing data).

**3. `internal/models/usage.go`** — add `CountSource` to `UsageEvent`:
```go
CountSource string `gorm:"column:count_source;size:30;not null;default:'phase3_delivered'"`
```
Add new constant `UsageEventTypeRowDelivered = "row_delivered"`.

**4. `internal/services/billing/dodo_client.go`** *(new)*
Initialize `dodopayments.NewClient()` using `option.WithBearerToken` and live/test mode. Returns `*dodopayments.Client`. Import: `github.com/dodopayments/dodopayments-go`.

**5. `internal/services/billing/billing_service.go`** *(new)*

Methods:
- `CreateOrGetCustomer(ctx, orgID, email, orgName)` → creates Dodo customer if `dodo_customer_id` is null, stores and returns ID
- `CreateCheckoutSession(ctx, orgID, plan, period)` → maps plan+period to product ID via config, calls `client.CheckoutSessions.Create()`, returns `checkout_url`
- `CreatePortalSession(ctx, orgID)` → calls `client.Customers.CreatePortalSession()`, returns `portal_url`
- `IngestRowUsage(ctx, orgID, runID, rowsDelivered)` → calls `client.UsageEvents.Ingest()` with `event_name=rows_delivered`, only when `rowsDelivered > 0`
- `UpdateOrgPlan(ctx, orgID, plan, period, subID, status, expiresAt)` → DB update

Plan mapping:
```go
func (s *BillingService) productID(plan, period string) (string, error) {
  switch plan + ":" + period {
  case "growth:monthly": return s.cfg.DodoProductGrowthMonthly, nil
  case "growth:annual":  return s.cfg.DodoProductGrowthAnnual, nil
  case "pro:monthly":    return s.cfg.DodoProductProMonthly, nil
  case "pro:annual":     return s.cfg.DodoProductProAnnual, nil
  case "enterprise:any": return s.cfg.DodoProductEnterprise, nil
  }
  return "", fmt.Errorf("unknown plan/period")
}
```

For plan-from-productID (webhook direction):
```go
func (s *BillingService) planFromProductID(productID string) (plan, period string) {
  switch productID {
  case s.cfg.DodoProductGrowthMonthly: return "growth", "monthly"
  case s.cfg.DodoProductGrowthAnnual:  return "growth", "annual"
  // ...
  default: return "starter", ""
  }
}
```

**6. `internal/server/callback.go`** — **Bug fix**

Change row recording from Phase 1 `rowsUpserted` to Phase 3 `phase3_rows_delivered`:

```go
// Extract phase3_rows_delivered from body
phase3Rows := num(body["phase3_rows_delivered"])
if phase3Rows == 0 {
    phase3Rows = num(body["delivery_rows"])
}

// Only record when delivery succeeded and rows > 0
if hasCallbackRun && (status == "completed") && phase3Rows > 0 && s.Usage != nil {
    _ = s.Usage.RecordDeliveredRows(
        callbackRun.OrganizationID,
        callbackRun.PipelineID,
        runUUID,
        int64(phase3Rows),
    )
}
// Ingest to Dodo for overage billing
if hasCallbackRun && status == "completed" && phase3Rows > 0 && s.Billing != nil {
    go s.Billing.IngestRowUsage(context.Background(),
        callbackRun.OrganizationID.String(),
        runID,
        int64(phase3Rows),
    )
}
```

**7. `internal/services/usage/usage_service.go`** — add `RecordDeliveredRows`:
```go
func (s *Service) RecordDeliveredRows(orgID, pipelineID, runID uuid.UUID, rowCount int64) error {
    // Writes event_type=row_delivered, count_source=phase3_delivered
}
```
Update `sumUsage` for `row_delivered` type to filter `count_source = 'phase3_delivered'`. Update `CurrentUsage` to use `row_delivered` for `RowsProcessed` metric.

**8. `internal/server/billing_http.go`** *(new)* — Fiber handlers:

- `POST /api/v1/billing/checkout` (JWT authed) → calls `BillingService.CreateCheckoutSession`
- `POST /api/v1/billing/webhook` (NO JWT — raw body, HMAC verify) → signature check using `DODO_WEBHOOK_SECRET`, dispatch by `event.type`:
  - `subscription.active` → upgrade plan
  - `subscription.plan_changed` → upgrade or downgrade plan
  - `subscription.cancelled` → set `plan_status=cancelled`, `plan_expires_at=next_billing_date`
  - `subscription.renewed` → update `plan_expires_at`
  - `subscription.on_hold` → set `plan_status=past_due`
  - `subscription.expired` → downgrade to starter
  - `payment.succeeded` / `payment.failed` → log billing event
- `GET /api/v1/billing/subscription` (JWT authed) → return org plan/status/period/expiry
- `POST /api/v1/billing/portal` (JWT authed) → return portal URL
- `POST /api/v1/billing/cancel` (JWT authed) → call Dodo `subscriptions.Cancel()`

Webhook signature verification (Go HMAC):
```go
func verifyDodoSignature(body []byte, sig, ts, secret string) bool {
    signed := ts + "." + string(body)
    mac := hmac.New(sha256.New, []byte(secret))
    mac.Write([]byte(signed))
    expected := base64.StdEncoding.EncodeToString(mac.Sum(nil))
    parts := strings.Split(sig, ",")
    if len(parts) < 2 { return false }
    return hmac.Equal([]byte(expected), []byte(parts[1]))
}
```

**9. `internal/server/routes.go`** — mount billing routes:
```go
// Public webhook (no JWT)
v1.Post("/billing/webhook", s.DodoBillingWebhook)

// JWT-authed billing routes
authed.Post("/billing/checkout", s.CreateCheckoutSession)
authed.Get("/billing/subscription", s.GetSubscription)
authed.Post("/billing/portal", s.CreatePortalSession)
authed.Post("/billing/cancel", s.CancelSubscription)
```

**10. `internal/server/state.go`** — add `Billing *billing.Service` field and wire in `NewState()`.

**11. `internal/database/migrate.go`** — GORM AutoMigrate already runs for `Organization` and `UsageEvent`; the new columns are added via struct tags so AutoMigrate handles them. No manual SQL needed.

---

### Frontend (`apps/app`)

**12. `lib/api/services/billing.service.ts`** *(new)*
```typescript
export class BillingService {
  static async createCheckout(orgId: string, plan: string, period: string)
  static async getSubscription(orgId: string)
  static async createPortal(orgId: string)
  static async cancelSubscription(orgId: string)
}
```

**13. `lib/api/hooks/use-billing.ts`** *(new)*
```typescript
export function useSubscription(orgId?: string)     // GET billing/subscription
export function useCreateCheckout()                  // mutation
export function useCreatePortal()                    // mutation
export function useCancelSubscription()             // mutation
```

**14. `lib/api/types/billing.ts`** *(new)*
```typescript
export interface SubscriptionStatus {
  plan: string; plan_period: string; plan_status: string;
  plan_expires_at: string | null; dodo_subscription_id: string | null;
}
```

**15. `app/workspace/settings/page.tsx`** — update billing tab:
- Add "Upgrade Plan" button when `plan === 'starter'` or current plan has a next plan → calls `createCheckout` → redirects to `checkout_url`
- Add "Manage Billing" button → calls `createPortal` → opens portal URL
- Add "Cancel Plan" button (destructive, owner-only)
- Add past-due banner: when `plan_status === 'past_due'`, show red banner with "Update Payment Method" link
- Show `plan_period` and `plan_expires_at` in the stat cards

**16. `app/workspace/billing/success/page.tsx`** *(new)* — success landing page after Dodo checkout redirect. Invalidates `useCurrentUsage` and `useSubscription` queries, shows confirmation, redirects to settings billing tab.

**17. `app/workspace/layout.tsx` or root layout** — add `PastDueBanner` component that conditionally renders based on `useSubscription().data?.plan_status === 'past_due'`.

---

## Upgrade/Downgrade support

- **Upgrade**: Frontend shows "Upgrade to Pro" button → POST `/billing/checkout` with new plan → Dodo checkout page. Dodo fires `subscription.plan_changed` webhook → Go updates `orgs.plan`.
- **Downgrade**: Via Dodo customer portal (`POST /billing/portal`) — user selects lower plan inside Dodo's portal, Dodo fires `subscription.plan_changed` → Go maps new product ID back to plan name and updates DB.
- Both paths converge at the `subscription.plan_changed` webhook handler.

---

## Row count invariants enforced

| Scenario | Rows billed |
|---|---|
| Phase 3 success, `phase3_rows_delivered > 0` | `phase3_rows_delivered` |
| Run status `failed` | 0 — no event written |
| Phase 1 staged, Phase 3 failed | 0 — only `phase3_rows_delivered` counts |
| `partial_success` with `phase3_rows_delivered > 0` | `phase3_rows_delivered` |

`RecordDeliveredRows` writes `count_source = 'phase3_delivered'`. The `sumUsage` query filters on this column so old events with wrong source are excluded.

---

## Dependency install

```bash
cd apps/server/main-server
go get github.com/dodopayments/dodopayments-go
```

---

## Order of execution

1. Install Dodo SDK
2. Add `CountSource` to `UsageEvent` model + new `UsageEventTypeRowDelivered` constant
3. Add Dodo columns to `Organization` model
4. Add Dodo fields to `Config` + `Load()`
5. Create `billing_service.go` + `dodo_client.go`
6. Fix `callback.go` row-count bug + add `Billing.IngestRowUsage` call
7. Add `RecordDeliveredRows` + update `sumUsage` filter in `usage_service.go`
8. Create `billing_http.go` handlers
9. Wire `State` + `routes.go`
10. Frontend: service, hooks, types, billing tab updates, success page, past-due banner
