# Dodo Payments — MantrixFlow Setup Guide

> **Current architecture.** Checkout is for the **first paid purchase** only. For organizations that already have a Dodo subscription (`dodo_subscription_id`), upgrades, downgrades, and monthly/annual switches use `POST …/billing/change-plan`, which calls Dodo server-side. The Dodo portal remains available through `POST …/billing/portal` for payment methods, receipts, and hosted invoice access. See [`billing-dodo-billingsdk.md`](./billing-dodo-billingsdk.md) for the current API and UI contract.

This is the repo guide for Dodo setup: products, **first-time checkout**, optional **product collection** fallback, customer portal, environment variables, webhooks, and verification. Align dashboard settings with [Dodo Product Collections](https://docs.dodopayments.com/features/product-collections) when using a collection. **Architecture diagrams:** [`dodo-payments-flowchart.md`](./dodo-payments-flowchart.md).

---

## Starter plan — no Dodo setup

**Starter is the default for every new organization.** It is free, has no checkout flow, and does **not** use Dodo product IDs or webhooks.


| What you need     | Starter                                                                        |
| ----------------- | ------------------------------------------------------------------------------ |
| Dodo account      | No                                                                             |
| `DODO_`* env vars | Only required if you sell Growth/Pro                                           |
| User flow         | Sign up in the app → workspace is on `plan=starter` until checkout upgrades it |


Paid checkout (`/workspace/billing/checkout`) and usage metering apply **after** the customer subscribes to Growth or Pro.

---

## Enterprise plan

Enterprise self-serve checkout is **not enabled** on the marketing site until launch. You can still set `DODO_PRODUCT_ENTERPRISE` when you create that product so the API accepts `plan=enterprise` from internal tools later.

---

## Prerequisites

- A [Dodo Payments](https://dodopayments.com) merchant account (business approved)
- Access to the MantrixFlow Go API `.env` file (`apps/server/main-server/.env`)
- A publicly reachable webhook URL for your deployment (or [ngrok](https://ngrok.com) for local testing)

---

## ENVIRONMENT and Dodo test vs live

The Go API reads `**ENVIRONMENT`** (see `internal/config/config.go`; default is `development` if unset).

That value controls which **Dodo API profile** the SDK uses (`internal/services/billing/dodo_client.go`):


| `ENVIRONMENT`                                                                          | Dodo SDK mode                        | Typical `DODO_PAYMENTS_API_KEY` prefix | Product IDs & webhook                        |
| -------------------------------------------------------------------------------------- | ------------------------------------ | -------------------------------------- | -------------------------------------------- |
| `**production`** (any casing, e.g. `Production`)                                       | **Live** (`WithEnvironmentLiveMode`) | `sk_live_…`                            | From the Dodo dashboard **Live** environment |
| **Anything else** (`development`, `dev`, `local`, `test`, `staging`, or empty default) | **Test** (`WithEnvironmentTestMode`) | `sk_test_…`                            | From the Dodo dashboard **Test** environment |


### Testing billing locally or on staging

1. In the Dodo dashboard, switch to **Test** mode (sandbox).
2. Create **test** products and copy **test** product IDs into `DODO_PRODUCT_`*.
3. Create a **test** webhook endpoint → copy its **test** signing secret into `DODO_WEBHOOK_SECRET`.
4. Set `DODO_PAYMENTS_API_KEY` to your **test** key (`sk_test_…`).
5. Keep `**ENVIRONMENT` unset** or set it to `development` / `staging` — **not** `production`.

Real charges and live bank traffic only happen when `**ENVIRONMENT=production`** **and** you use `**sk_live_…`** and live product IDs.

> **Important:** Test and live each have their own API keys, product IDs, and webhook secrets. Mixing (e.g. test key + live product ID) will fail.

---

## Step 1 — Create Products in the Dodo Dashboard

You need one product per plan × billing period combination. Log in to the [Dodo Dashboard](https://dashboard.dodopayments.com) → **Products** → **Create product**.


| Product name                                     | Type         | Price           | Billing interval            |
| ------------------------------------------------ | ------------ | --------------- | --------------------------- |
| MantrixFlow Growth (Monthly)                     | Subscription | $49 / month     | Monthly                     |
| MantrixFlow Growth (Annual)                      | Subscription | $37 / month     | Annual (billed $444/year)   |
| MantrixFlow Pro (Monthly)                        | Subscription | $199 / month    | Monthly                     |
| MantrixFlow Pro (Annual)                         | Subscription | $149 / month    | Annual (billed $1,788/year) |
| MantrixFlow Enterprise *(optional until launch)* | Subscription | As you price it | Monthly                     |


Enterprise: create the product **when you open sales**; until then you can skip `DODO_PRODUCT_ENTERPRISE` if you never call checkout with `plan=enterprise`.

After saving each product, copy the **Product ID** (format: `prd_xxxxxxxxxxxx`).

> **Overage note:** current production overage is MantrixFlow-owned local accounting plus a manual cycle-end invoice ledger, not automatic Dodo usage-meter invoicing. Keep Dodo usage meters disabled unless you intentionally re-enable that legacy path.

---

## Step 1b — Product collection & customer portal (optional)

Set `DODO_PRODUCT_COLLECTION_ID` only as a **fallback** when `DODO_PRODUCT_`* cannot be resolved for a given `plan` + `period` (see `internal/services/billing/billing_service.go`). **Normal behavior** is **single-product checkout** (`product_cart` with one line item) so the customer gets the exact SKU for the tier and billing interval they chose in the app.

**Important:** A **new checkout session** creates a **new** Dodo subscription. Customers who **already** have a subscription must **not** run checkout again — you can get **duplicate subscriptions**. **Existing subscribers** change tier or billing interval in **Manage billing** (`POST …/billing/portal`), with confirmation in Dodo’s UI; MantrixFlow waits for **webhooks** to refresh `organizations.plan`. Enable **Allow Subscription Updates** in Dodo so the portal can offer plan changes among products in your collection.

Downgrades and upgrades from the app UI now use `POST …/billing/change-plan`. Owners use **Manage billing** (portal) only for payment methods, receipts, and hosted invoice access.

### Create a collection

Dashboard → **Products** → **Collections** → **Create collection**. Add **every** subscription SKU that should be interchangeable (Growth/Pro × monthly/annual, plus Enterprise if applicable). Order matters: Dodo **pre-selects the first product** in collection checkout.

**Rule:** Each product can belong to **only one** collection.

### Descriptions & images (customer-facing)

- **Description:** One line of outcome, then key limits (pipelines, rows/month, API) — match your pricing page. Avoid internal jargon.
- **Image:** PNG/JPG; consistent aspect ratio across tiers; don’t rely on color alone (put the tier name in the product title).

### Enable portal plan changes

Dashboard → **Settings** → **Subscriptions**:


| Setting                        | Purpose                                                                                                          |
| ------------------------------ | ---------------------------------------------------------------------------------------------------------------- |
| **Allow Subscription Updates** | Customers can change plans in the **Customer Portal** (upgrade/downgrade among products in the same collection). |


> This toggle is **not** inside the collection editor; it is a business-level setting.

### MantrixFlow behavior


| Flow                                                                   | Implementation                                                                                                                                                                                                                                                                                                     |
| ---------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| **First paid plan (e.g. Starter → Growth)**                            | `POST …/billing/checkout` with body `plan` + `period` → Dodo **hosted checkout** (one `product_cart` line when `DODO_PRODUCT_`* are set; **collection** only if those env vars are missing and `**DODO_PRODUCT_COLLECTION_ID`** is set). Webhooks (and live subscription reconciliation) set `organizations.plan`. |
| **Upgrade / downgrade / switch monthly ↔ annual (already subscribed)** | `POST …/billing/change-plan` → MantrixFlow API calls Dodo subscription change-plan. Upgrades apply immediately; downgrades schedule for the next billing date. Webhooks reconcile the org afterward.                                                                                                      |
| **Cleaning up duplicate subs**                                         | Cancel the unwanted subscription in the Dodo dashboard (or ask the customer to remove it in Manage billing); keep the subscription id you want and align `organizations.dodo_subscription_id` via webhook or support.                                                                                              |


---

## Step 2 — Configure Environment Variables

Add these to `apps/server/main-server/.env`:

```env
# ─── App environment (controls Dodo live vs test SDK mode) ─────
# Use production ONLY for real billing (sk_live_ + live products).
# For local/staging Dodo sandbox: development, dev, local, test, or staging.
ENVIRONMENT=development

# ─── Dodo Payments ─────────────────────────────────────────────
# Test: sk_test_…   Live: sk_live_…  (must match ENVIRONMENT + dashboard mode)
DODO_PAYMENTS_API_KEY=sk_test_xxxxxxxxxxxxxxxxxxxxxxxx

# Webhook secret — from Dashboard → Webhooks → endpoint (Test or Live, must match key)
DODO_WEBHOOK_SECRET=whsec_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx

# Product IDs — copy from the SAME mode as the key (test dashboard → test prd_…)
DODO_PRODUCT_GROWTH_MONTHLY=prd_xxxxxxxxxxxxxxxx
DODO_PRODUCT_GROWTH_ANNUAL=prd_xxxxxxxxxxxxxxxx
DODO_PRODUCT_PRO_MONTHLY=prd_xxxxxxxxxxxxxxxx
DODO_PRODUCT_PRO_ANNUAL=prd_xxxxxxxxxxxxxxxx
# Optional until Enterprise checkout is enabled:
DODO_PRODUCT_ENTERPRISE=

# Optional fallback: collection checkout when plan+period cannot be mapped to DODO_PRODUCT_* alone
DODO_PRODUCT_COLLECTION_ID=

# Rows usage meter (billable overage)
DODO_ROWS_DELIVERED_EVENT_NAME=rows_delivered
DODO_ROWS_DELIVERED_METER_ID=

# Optional API-call usage meter (disabled unless you deliberately bill API overage)
DODO_API_CALLS_USAGE_BILLING_ENABLED=false
DODO_API_CALLS_EVENT_NAME=api_call
DODO_API_CALLS_METER_ID=
```

> Dodo **test vs live** follows `ENVIRONMENT`: see **ENVIRONMENT and Dodo test vs live** above.

## Step 3 — Register the Webhook Endpoint

In the Dodo Dashboard → **Webhooks** → **Add endpoint**.


| Field               | Value                                                                                                                                                                                                              |
| ------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| URL                 | `https://your-api-domain.com/api/v1/billing/webhook`                                                                                                                                                               |
| Events to subscribe | `subscription.active`, `subscription.plan_changed`, `subscription.renewed`, `subscription.cancelled`, `subscription.on_hold`, `subscription.expired`, `subscription.failed`, `payment.succeeded`, `payment.failed` |


After saving, copy the **Signing secret** and set it as `DODO_WEBHOOK_SECRET` in `.env`.

### Local Testing with ngrok

```bash
ngrok http 5000   # same port as the Go API (default PORT=5000)
```

In the **Dodo Dashboard → Webhooks**, set the endpoint URL to:

```text
https://YOUR-SUBDOMAIN.ngrok-free.app/api/v1/billing/webhook
```

Example (your tunnel will differ): `https://1b5f-110-227-207-99.ngrok-free.app/api/v1/billing/webhook`

Use the **HTTPS** forwarding URL from ngrok, not `http://localhost:5000`.

---

## Step 4 — Database Migration

GORM AutoMigrate runs on server startup and adds the new columns automatically:


| Table              | New columns                                                                                                    |
| ------------------ | -------------------------------------------------------------------------------------------------------------- |
| `organizations`    | `dodo_customer_id`, `dodo_subscription_id`, `plan_period`, `plan_status`, `plan_started_at`, `plan_expires_at` |
| `elt_usage_events` | `count_source`                                                                                                 |


No manual SQL is required. You can verify after first startup:

```sql
SELECT column_name FROM information_schema.columns
WHERE table_name = 'organizations'
  AND column_name LIKE 'dodo%'
  OR column_name LIKE 'plan_%';
```

---

## Step 5 — API Routes Summary


| Method   | Path                                                    | Auth                      | Description                                                                                                                                                                                                       |
| -------- | ------------------------------------------------------- | ------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `POST`   | `/api/v1/billing/webhook`                               | None (signature verified) | Dodo event receiver                                                                                                                                                                                               |
| `GET`    | `/api/v1/organizations/:orgId/billing/subscription`     | JWT                       | Current plan status                                                                                                                                                                                               |
| `POST`   | `/api/v1/organizations/:orgId/billing/checkout`         | JWT + OWNER               | Hosted checkout (`product_cart` SKU from `plan`+`period`, or collection **fallback** if `DODO_PRODUCT_COLLECTION_ID` is set and product env vars are missing) — **first subscription only**; avoid duplicate subs |
| `POST`   | `/api/v1/organizations/:orgId/billing/portal`           | JWT                       | Customer portal (plan changes with Dodo confirmation, payment method, invoices, wallet)                                                                                                                           |
| `DELETE` | `/api/v1/organizations/:orgId/billing/scheduled-change` | JWT + OWNER               | Cancel a pending scheduled plan change in Dodo                                                                                                                                                                    |
| `POST`   | `/api/v1/organizations/:orgId/billing/cancel`           | JWT + OWNER               | Cancel at period end                                                                                                                                                                                              |


### Checkout request body

```json
{ "plan": "growth", "period": "monthly" }
```

Valid `plan` values: `growth`, `pro`, `enterprise`  
Valid `period` values: `monthly`, `annual`

### Subscription response shape (GET subscription)

```json
{
  "plan": "growth",
  "plan_period": "monthly",
  "plan_status": "active",
  "plan_started_at": "2026-04-01T00:00:00Z",
  "plan_expires_at": "2026-05-01T00:00:00Z",
  "dodo_subscription_id": "sub_xxxxxxxxxxxxxxxx"
}
```

---

## Step 6 — Frontend Billing Flow

### Pricing page CTA (website)

The website pricing page (`apps/website/app/pricing/pricing-client.tsx`) links paid plan CTAs to:

```
https://cloud.mantrixflow.com/workspace/billing/checkout?plan=growth&period=monthly
```

The `period` parameter changes when the user toggles Monthly / Annual on the pricing page.

### App checkout page

`/workspace/billing/checkout?plan=<plan>&period=<period>` — the user lands here, the page calls the backend, receives a Dodo-hosted checkout URL, and redirects the browser immediately.

### App success page

After payment Dodo redirects the user to:

```
https://cloud.mantrixflow.com/workspace/billing/success
```

This page invalidates `billing/subscription` and `usage/current` queries, shows a confirmation, then auto-redirects to Settings → Billing.

### Past-due banner

A global dismissible banner is rendered in `WorkspaceLayout` whenever `plan_status === "past_due"`. It links to the billing settings portal.

---

## Step 7 — Verify the Integration

### End-to-end smoke test

1. **Sign up** or log in to the app.
2. Navigate to **Settings → Billing**.
3. Click **Upgrade plan** (Growth, Monthly).
4. Complete payment on the Dodo-hosted checkout page (use test card `4111 1111 1111 1111` in test mode).
5. You should land on `/workspace/billing/success` with a confirmation message.
6. The Settings → Billing tab should now show `Plan: Growth · Monthly`.

### Webhook replay (Dodo Dashboard)

1. Go to **Webhooks** → select your endpoint → **Deliveries**.
2. Find the `subscription.active` event and click **Replay**.
3. The Go API log should show a successful plan update.

### Check the database

```sql
SELECT id, plan, plan_status, plan_period, dodo_subscription_id
FROM organizations
WHERE dodo_customer_id IS NOT NULL
LIMIT 5;
```

---

## Step 8 — Legacy Dodo Usage-Based Overage Billing

> **Current production path:** overage is MantrixFlow-owned local usage
> accounting plus a manual cycle-end invoice ledger. Do not enable the Dodo
> usage-meter flow below unless you intentionally want to revive legacy
> automatic metered billing and have verified it will not double-charge.

Dodo usage-based billing works in three parts: the app sends usage events, a Dodo meter aggregates those events, and the product's usage pricing bills the overage at the subscription cycle invoice. Official Dodo docs: [Usage Based Billing](https://docs.dodopayments.com/features/usage-based-billing/introduction), [Meters](https://docs.dodopayments.com/features/usage-based-billing/meters), and [Hybrid Billing](https://docs.dodopayments.com/features/hybrid-billing).

### Recommended launch setup

Start with **rows overage only**:

- Bill rows delivered beyond the included allowance.
- Keep API calls as the existing daily hard limit.
- Enable API-call billing later only if you want a second metered line item.

### Create the rows-delivered meter

In the Dodo Dashboard, go to **Meters** (or **Products → Meters**, depending on the dashboard layout) → **Create Meter**.


| Dodo field                   | Value to enter               |
| ---------------------------- | ---------------------------- |
| Meter name                   | `MantrixFlow Rows Delivered` |
| Event name                   | `rows_delivered`             |
| Aggregation type             | `Sum`                        |
| Over property / Metadata key | `rows_delivered`             |
| Measurement unit             | `rows`                       |
| Filters                      | None for launch              |


After saving, copy the meter ID (`mtr_...`) if Dodo shows one. The app does **not** need the meter ID to route usage; Dodo matches usage by the exact case-sensitive **event name**. `DODO_ROWS_DELIVERED_METER_ID` is optional bookkeeping metadata so support can correlate ingested events with the dashboard meter.

Add the rows meter as **usage pricing** to each paid subscription product:


| Product        | Free threshold                    | Overage unit price  |
| -------------- | --------------------------------- | ------------------- |
| Growth monthly | `1,000,000` rows / billing cycle  | `$0.000005` per row |
| Pro monthly    | `10,000,000` rows / billing cycle | `$0.000005` per row |
| Growth annual  | See annual note below             | `$0.000005` per row |
| Pro annual     | See annual note below             | `$0.000005` per row |
| Enterprise     | Custom contract value             | Custom or unlimited |


`$0.000005` per row is the same as `$5 / 1,000,000 rows`.

**Annual product note:** the app now estimates row usage using the current subscription billing cycle. If the Dodo annual product bills usage on an annual cycle, annualize the included allowance:


| Annual product | Annual free threshold     |
| -------------- | ------------------------- |
| Growth annual  | `12,000,000` rows / year  |
| Pro annual     | `120,000,000` rows / year |


If you want annual subscribers to receive a fresh row allowance every month, configure the Dodo product/meter for a monthly usage billing cadence if your Dodo account supports that. Do not leave annual products with only the monthly threshold unless you intentionally want Growth annual to include only `1,000,000` rows for the whole year.

### Rows env values

Set these in `apps/server/main-server/.env`:

```env
DODO_ROWS_DELIVERED_EVENT_NAME=rows_delivered
DODO_ROWS_DELIVERED_METER_ID=mtr_xxxxxxxxxxxxxxxx
```

Leave `DODO_ROWS_DELIVERED_METER_ID` empty until the meter exists. Keep `DODO_ROWS_DELIVERED_EVENT_NAME=rows_delivered` unless you also rename the meter event name in Dodo.

### Rows event sent by the server

Row overages are sent to Dodo after each successful pipeline run:

```
Event name : rows_delivered
Event ID   : <pipeline_run_id>   (idempotent — replays are safe)
Customer ID: <dodo_customer_id>
Quantity   : phase3_rows_delivered  (Phase 3 only — failed runs send nothing)
Metadata   : org_id, run_id, rows_delivered, optional meter_id
```

To verify ingestion, check the Dodo Dashboard → **Usage** → filter by `rows_delivered`.

> Only rows **successfully delivered** to the destination table are counted. Staged rows that fail before delivery are never billed. This is enforced by the `count_source = 'phase3_delivered'` filter in the `elt_usage_events` table.

### Rows overage calculation

Formula:

```text
overage_rows = max(0, delivered_rows_in_billing_cycle - included_rows)
row_overage_charge = overage_rows * 0.000005
```

Examples:


| Plan   | Delivered rows in cycle | Included rows | Billable overage | Charge   |
| ------ | ----------------------- | ------------- | ---------------- | -------- |
| Growth | `1,600,000`             | `1,000,000`   | `600,000`        | `$3.00`  |
| Pro    | `12,000,000`            | `10,000,000`  | `2,000,000`      | `$10.00` |


If a Growth customer reaches the `1,000,000` row allowance in the first 10 days of a monthly cycle, the next delivered rows in that same cycle are billable immediately as metered overage. Dodo aggregates usage through the cycle and adds the charge to the subscription invoice according to the product's usage billing settings.

### Optional API-call usage billing

By default, API calls stay as an app-enforced daily guard and are **not** sent to Dodo for invoicing:

```env
DODO_API_CALLS_USAGE_BILLING_ENABLED=false
```

Enable this only if you want Dodo to invoice API-call overages as a second usage line item.

Create a second Dodo meter:


| Dodo field                   | Value to enter          |
| ---------------------------- | ----------------------- |
| Meter name                   | `MantrixFlow API Calls` |
| Event name                   | `api_call`              |
| Aggregation type             | `Count`                 |
| Over property / Metadata key | Leave empty for Count   |
| Measurement unit             | `calls`                 |
| Filters                      | None for launch         |


Attach it to the paid products only if you want API overage billed:


| Product        | Suggested free threshold                          | Overage unit price   |
| -------------- | ------------------------------------------------- | -------------------- |
| Growth monthly | `1,500,000` calls / billing cycle                 | `$0.000002` per call |
| Pro monthly    | `15,000,000` calls / billing cycle                | `$0.000002` per call |
| Growth annual  | `18,000,000` calls / year, if annual usage cycle  | `$0.000002` per call |
| Pro annual     | `180,000,000` calls / year, if annual usage cycle | `$0.000002` per call |


`$0.000002` per call is the same as `$0.002 / 1,000 API calls`. The monthly thresholds above convert the app's daily limits into a 30-day billing-cycle allowance: Growth `50,000/day × 30`, Pro `500,000/day × 30`.

Then set:

```env
DODO_API_CALLS_USAGE_BILLING_ENABLED=true
DODO_API_CALLS_EVENT_NAME=api_call
DODO_API_CALLS_METER_ID=mtr_xxxxxxxxxxxxxxxx
```

`DODO_API_CALLS_METER_ID` is also optional metadata. Dodo matches this meter by `api_call`.

---

## Step 9 — Upgrade / Downgrade Flows


| User action                                                     | How it works                                                                                                                                                                                                                                                        |
| --------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **New subscription (no active paid sub yet)**                   | `POST …/billing/checkout` — typically **single-product** cart from `DODO_PRODUCT_*`; **collection** only if SKUs cannot be resolved and `DODO_PRODUCT_COLLECTION_ID` is set. `organizations.plan` follows webhooks and GET subscription reconciliation. |
| **Upgrade / downgrade / monthly ↔ annual (already subscribed)** | `POST …/billing/change-plan` — MantrixFlow API calls Dodo change-plan server-side. Do **not** run checkout again (duplicate subscriptions).                                                                                                             |
| **Portal**                                                      | `POST …/billing/portal` — payment method, receipts, and hosted invoice access.                                                                                                                                                                          |
| **Cancel**                                                      | **Cancel plan** → `POST …/billing/cancel` → cancel at next billing date → webhooks                                                                                                                                                                                  |
| **Payment failure**                                             | `payment.failed` / `subscription.on_hold` → `past_due` UX                                                                                                                                                                                                           |
| **Plan expires**                                                | `subscription.expired` → reset starter, clear subscription id                                                                                                                                                                                                       |


---

## Environment Variable Reference


| Variable                               | Required | Description                                                                                                               |
| -------------------------------------- | -------- | ------------------------------------------------------------------------------------------------------------------------- |
| `DODO_PAYMENTS_API_KEY`                | Yes      | Bearer token for Dodo API calls                                                                                           |
| `DODO_WEBHOOK_SECRET`                  | Yes      | Webhook signing secret (`whsec_…`)                                                                                        |
| `DODO_PRODUCT_GROWTH_MONTHLY`          | Yes      | Product ID for Growth monthly                                                                                             |
| `DODO_PRODUCT_GROWTH_ANNUAL`           | Yes      | Product ID for Growth annual                                                                                              |
| `DODO_PRODUCT_PRO_MONTHLY`             | Yes      | Product ID for Pro monthly                                                                                                |
| `DODO_PRODUCT_PRO_ANNUAL`              | Yes      | Product ID for Pro annual                                                                                                 |
| `DODO_PRODUCT_ENTERPRISE`              | No*      | Product ID for Enterprise (*only if you enable enterprise checkout)                                                       |
| `DODO_PRODUCT_COLLECTION_ID`           | No       | **Fallback** collection when `plan`+`period` cannot be mapped to `DODO_PRODUCT_*` alone                                   |
| `DODO_ROWS_DELIVERED_EVENT_NAME`       | No       | Rows usage event name. Default: `rows_delivered`. Must match the Dodo rows meter event name exactly.                      |
| `DODO_ROWS_DELIVERED_METER_ID`         | No       | Optional Dodo rows meter ID (`mtr_...`) stored in event metadata for audit/debugging. Leave blank until the meter exists. |
| `DODO_API_CALLS_USAGE_BILLING_ENABLED` | No       | Enables Dodo API-call usage ingestion when `true`. Default: `false`; API calls remain a daily hard guard.                 |
| `DODO_API_CALLS_EVENT_NAME`            | No       | API-call usage event name. Default: `api_call`. Must match the optional Dodo API-call meter event name exactly.           |
| `DODO_API_CALLS_METER_ID`              | No       | Optional Dodo API-call meter ID (`mtr_...`) stored in event metadata for audit/debugging.                                 |
| `ENVIRONMENT`                          | No       | Set to `production` for live Dodo + `sk_live_…`. Use `development` (default) or `staging` for `sk_test_…`.                |


---

## Troubleshooting


| Symptom                                               | Likely cause                                                                                           | Fix                                                                                                                                                                                  |
| ----------------------------------------------------- | ------------------------------------------------------------------------------------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| Checkout button shows "Billing not configured"        | `DODO_PAYMENTS_API_KEY` is empty                                                                       | Set the env var and restart the Go server                                                                                                                                            |
| Webhook returns 401                                   | `DODO_WEBHOOK_SECRET` mismatch                                                                         | Copy the exact secret from the Dodo Dashboard; no extra spaces                                                                                                                       |
| Plan not updating after payment                       | Webhook not reaching the server                                                                        | Check ngrok / firewall; replay from Dodo Dashboard                                                                                                                                   |
| Checkout URL is empty string                          | Product ID env var not set **or** collection misconfigured                                             | Set `DODO_PRODUCT_`* for the requested plan when **not** using a collection; set `DODO_PRODUCT_COLLECTION_ID` for collection mode                                                    |
| Rows not billed to Dodo                               | Missing customer, meter event mismatch, product usage pricing not attached, or test/live mismatch      | Confirm the org has `dodo_customer_id`, the meter event name is `rows_delivered`, the meter is attached to the active product, and key/product/meter are all from the same Dodo mode |
| Dodo meter receives events but invoice has no overage | Free threshold not exceeded, wrong billing period, or usage meter not attached to the subscription SKU | Check the product's usage pricing threshold, annual vs monthly usage period, and the exact product ID on the active subscription                                                     |
| API calls not appearing in Dodo usage                 | API-call metering is disabled or event name mismatch                                                   | Set `DODO_API_CALLS_USAGE_BILLING_ENABLED=true`, create an `api_call` Count meter, attach it to paid products, then restart the Go server                                            |
| Dodo API errors (invalid product, auth)               | Test/live mismatch                                                                                     | Use `sk_test_` + test product IDs + test webhook when `ENVIRONMENT` is not `production`; use live only when `ENVIRONMENT=production`                                                 |
| Plan change in portal not reflected in app            | Webhook delivery or signing secret                                                                     | Replay `**subscription.plan_changed`** from Dodo; confirm `**DODO_WEBHOOK_SECRET**` matches the dashboard                                                                            |
