# Dodo Payments — MantrixFlow Setup Guide

This guide walks you through connecting MantrixFlow to Dodo Payments end-to-end: creating products, configuring environment variables, registering webhooks, and verifying the integration.

---

## Starter plan — no Dodo setup

**Starter is the default for every new organization.** It is free, has no checkout flow, and does **not** use Dodo product IDs or webhooks.

| What you need | Starter |
|---|---|
| Dodo account | No |
| `DODO_*` env vars | Only required if you sell Growth/Pro |
| User flow | Sign up in the app → workspace is on `plan=starter` until checkout upgrades it |

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

The Go API reads **`ENVIRONMENT`** (see `internal/config/config.go`; default is `development` if unset).

That value controls which **Dodo API profile** the SDK uses (`internal/services/billing/dodo_client.go`):

| `ENVIRONMENT` | Dodo SDK mode | Typical `DODO_PAYMENTS_API_KEY` prefix | Product IDs & webhook |
|---|---|---|---|
| **`production`** (any casing, e.g. `Production`) | **Live** (`WithEnvironmentLiveMode`) | `sk_live_…` | From the Dodo dashboard **Live** environment |
| **Anything else** (`development`, `dev`, `local`, `test`, `staging`, or empty default) | **Test** (`WithEnvironmentTestMode`) | `sk_test_…` | From the Dodo dashboard **Test** environment |

### Testing billing locally or on staging

1. In the Dodo dashboard, switch to **Test** mode (sandbox).
2. Create **test** products and copy **test** product IDs into `DODO_PRODUCT_*`.
3. Create a **test** webhook endpoint → copy its **test** signing secret into `DODO_WEBHOOK_SECRET`.
4. Set `DODO_PAYMENTS_API_KEY` to your **test** key (`sk_test_…`).
5. Keep **`ENVIRONMENT` unset** or set it to `development` / `staging` — **not** `production`.

Real charges and live bank traffic only happen when **`ENVIRONMENT=production`** **and** you use **`sk_live_…`** and live product IDs.

> **Important:** Test and live each have their own API keys, product IDs, and webhook secrets. Mixing (e.g. test key + live product ID) will fail.

---

## Step 1 — Create Products in the Dodo Dashboard

You need one product per plan × billing period combination. Log in to the [Dodo Dashboard](https://dashboard.dodopayments.com) → **Products** → **Create product**.

| Product name | Type | Price | Billing interval |
|---|---|---|---|
| MantrixFlow Growth (Monthly) | Subscription | $49 / month | Monthly |
| MantrixFlow Growth (Annual) | Subscription | $37 / month | Annual (billed $444/year) |
| MantrixFlow Pro (Monthly) | Subscription | $199 / month | Monthly |
| MantrixFlow Pro (Annual) | Subscription | $149 / month | Annual (billed $1,788/year) |
| MantrixFlow Enterprise *(optional until launch)* | Subscription | As you price it | Monthly |

Enterprise: create the product **when you open sales**; until then you can skip `DODO_PRODUCT_ENTERPRISE` if you never call checkout with `plan=enterprise`.

After saving each product, copy the **Product ID** (format: `prd_xxxxxxxxxxxx`).

> **Usage-based overage** (rows delivered): If you want to charge per-row overages automatically, attach a Meter to each paid product in the dashboard. The event name must be `rows_delivered` (this is the name sent by `IngestRowUsage`).

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
```

> Dodo **test vs live** follows `ENVIRONMENT`: see **ENVIRONMENT and Dodo test vs live** above.

---

## Step 3 — Register the Webhook Endpoint

In the Dodo Dashboard → **Webhooks** → **Add endpoint**.

| Field | Value |
|---|---|
| URL | `https://your-api-domain.com/api/v1/billing/webhook` |
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

| Table | New columns |
|---|---|
| `organizations` | `dodo_customer_id`, `dodo_subscription_id`, `plan_period`, `plan_status`, `plan_started_at`, `plan_expires_at` |
| `elt_usage_events` | `count_source` |

No manual SQL is required. You can verify after first startup:

```sql
SELECT column_name FROM information_schema.columns
WHERE table_name = 'organizations'
  AND column_name LIKE 'dodo%'
  OR column_name LIKE 'plan_%';
```

---

## Step 5 — API Routes Summary

| Method | Path | Auth | Description |
|---|---|---|---|
| `POST` | `/api/v1/billing/webhook` | None (signature verified) | Dodo event receiver |
| `GET` | `/api/v1/organizations/:orgId/billing/subscription` | JWT | Current plan status |
| `POST` | `/api/v1/organizations/:orgId/billing/checkout` | JWT + OWNER | Create hosted checkout |
| `POST` | `/api/v1/organizations/:orgId/billing/portal` | JWT | Customer portal link |
| `POST` | `/api/v1/organizations/:orgId/billing/cancel` | JWT + OWNER | Cancel at period end |

### Checkout request body

```json
{ "plan": "growth", "period": "monthly" }
```

Valid `plan` values: `growth`, `pro`, `enterprise`  
Valid `period` values: `monthly`, `annual`

### Subscription response

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

## Step 8 — Usage-Based Overage Billing

Row overages are automatically sent to Dodo after each successful pipeline run:

```
Event name : rows_delivered
Event ID   : <pipeline_run_id>   (idempotent — replays are safe)
Customer ID: <dodo_customer_id>
Quantity   : phase3_rows_delivered  (Phase 3 only — failed runs send nothing)
```

To verify ingestion, check the Dodo Dashboard → **Usage** → filter by `rows_delivered`.

> Only rows **successfully delivered** to the destination table are counted. Staged rows that fail before delivery are never billed. This is enforced by the `count_source = 'phase3_delivered'` filter in the `elt_usage_events` table.

---

## Step 9 — Upgrade / Downgrade Flows

| User action | How it works |
|---|---|
| **Upgrade** | User clicks "Upgrade plan" → POST `/billing/checkout` → Dodo checkout → `subscription.active` or `subscription.plan_changed` webhook → DB updated |
| **Downgrade** | User opens Customer Portal → selects lower plan → Dodo fires `subscription.plan_changed` → DB updated at period end |
| **Cancel** | User clicks "Cancel plan" → POST `/billing/cancel` → sets `cancel_at_next_billing_date=true` → Dodo fires `subscription.cancelled` → `plan_status=cancelled`, access until `plan_expires_at` |
| **Payment failure** | Dodo fires `subscription.on_hold` or `payment.failed` → `plan_status=past_due` → Global past-due banner shown |
| **Plan expires** | Dodo fires `subscription.expired` → plan reset to `starter`, `dodo_subscription_id` cleared |

---

## Environment Variable Reference

| Variable | Required | Description |
|---|---|---|
| `DODO_PAYMENTS_API_KEY` | Yes | Bearer token for Dodo API calls |
| `DODO_WEBHOOK_SECRET` | Yes | Webhook signing secret (`whsec_…`) |
| `DODO_PRODUCT_GROWTH_MONTHLY` | Yes | Product ID for Growth monthly |
| `DODO_PRODUCT_GROWTH_ANNUAL` | Yes | Product ID for Growth annual |
| `DODO_PRODUCT_PRO_MONTHLY` | Yes | Product ID for Pro monthly |
| `DODO_PRODUCT_PRO_ANNUAL` | Yes | Product ID for Pro annual |
| `DODO_PRODUCT_ENTERPRISE` | No* | Product ID for Enterprise (*only if you enable enterprise checkout) |
| `ENVIRONMENT` | No | Set to `production` for live Dodo + `sk_live_…`. Use `development` (default) or `staging` for `sk_test_…`. |

---

## Troubleshooting

| Symptom | Likely cause | Fix |
|---|---|---|
| Checkout button shows "Billing not configured" | `DODO_PAYMENTS_API_KEY` is empty | Set the env var and restart the Go server |
| Webhook returns 401 | `DODO_WEBHOOK_SECRET` mismatch | Copy the exact secret from the Dodo Dashboard; no extra spaces |
| Plan not updating after payment | Webhook not reaching the server | Check ngrok / firewall; replay from Dodo Dashboard |
| Checkout URL is empty string | Product ID env var not set | Set `DODO_PRODUCT_*` for the plan/period combination |
| Rows not billed to Dodo | `dodo_customer_id` is null | Customer was not created yet; trigger a checkout first |
| Dodo API errors (invalid product, auth) | Test/live mismatch | Use `sk_test_` + test product IDs + test webhook when `ENVIRONMENT` is not `production`; use live only when `ENVIRONMENT=production` |
