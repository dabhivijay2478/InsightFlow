# Dodo AutoSend Transformations

These examples are for the Dodo Payments AutoSend integration. Dodo owns customer-facing payment, subscription, refund, dispute, and invoice emails. The backend should only reconcile billing state from Dodo webhooks.

## Shared Helpers

Paste helpers at the top of each Dodo transformation if the integration editor does not support shared code.

```js
const FROM = { email: "support@mantrixflow.com", name: "MantrixFlow" };
const REPLY_TO = { email: "support@mantrixflow.com", name: "MantrixFlow Support" };
const BILLING_URL = "https://cloud.mantrixflow.com/workspace/settings";

const templates = {
  payment_succeeded: "tpl_payment_succeeded",
  payment_failed: "tpl_payment_failed",
  subscription_active: "tpl_subscription_active",
  subscription_renewed: "tpl_subscription_renewed",
  subscription_plan_changed: "tpl_subscription_plan_changed",
  subscription_cancelled: "tpl_subscription_cancelled",
  subscription_on_hold: "tpl_subscription_on_hold",
  subscription_expired: "tpl_subscription_expired",
  refund_succeeded: "tpl_refund_succeeded",
  dispute_opened: "tpl_dispute_opened",
  invoice_available: "tpl_invoice_available"
};

function valueAt(obj, paths, fallback = "") {
  for (const path of paths) {
    const value = path.split(".").reduce((acc, part) => (acc && acc[part] !== undefined ? acc[part] : undefined), obj);
    if (value !== undefined && value !== null && String(value).trim() !== "") return value;
  }
  return fallback;
}

function customer(event) {
  const email = valueAt(event, ["data.customer.email", "data.customer_email", "customer.email", "customer_email"]);
  const name = valueAt(event, ["data.customer.name", "data.customer_name", "customer.name", "customer_name"], "there");
  return { email, name };
}

function money(event) {
  const amountRaw = Number(valueAt(event, ["data.amount", "data.total", "amount"], 0));
  const currency = String(valueAt(event, ["data.currency", "currency"], "USD")).toUpperCase();
  const amount = amountRaw > 1000 ? (amountRaw / 100).toFixed(2) : amountRaw.toFixed(2);
  return { amount, currency };
}

function send(templateKey, subject, dynamicData, toOverride) {
  webhook.url = "https://api.autosend.com/v1/mails/send";
  webhook.payload = {
    to: toOverride || customer(webhook),
    from: FROM,
    subject,
    templateId: templates[templateKey],
    dynamicData,
    replyTo: REPLY_TO,
    trackingOpen: true,
    trackingClick: true
  };
}
```

## `payment.succeeded`

```js
const c = customer(webhook);
const m = money(webhook);
send("payment_succeeded", "Payment received for MantrixFlow", {
  customer_name: c.name,
  amount: m.amount,
  currency: m.currency,
  payment_id: valueAt(webhook, ["data.payment_id", "data.id", "id"]),
  receipt_url: valueAt(webhook, ["data.receipt_url", "data.invoice_url"], BILLING_URL),
  date: new Date().toISOString()
}, c);
```

## `payment.failed`

```js
const c = customer(webhook);
const m = money(webhook);
send("payment_failed", "Payment failed for MantrixFlow", {
  customer_name: c.name,
  amount: m.amount,
  currency: m.currency,
  payment_id: valueAt(webhook, ["data.payment_id", "data.id", "id"]),
  failure_reason: valueAt(webhook, ["data.failure_reason", "data.error_message", "data.status"], "The payment could not be completed."),
  billing_url: BILLING_URL
}, c);
```

## `subscription.active`

```js
const c = customer(webhook);
send("subscription_active", "Your MantrixFlow subscription is active", {
  customer_name: c.name,
  plan_name: valueAt(webhook, ["data.product.name", "data.plan.name", "data.product_name"], "MantrixFlow plan"),
  billing_period: valueAt(webhook, ["data.billing_period", "data.interval"], "current billing period"),
  billing_url: BILLING_URL
}, c);
```

## Renewal or Plan Change

Use the event names provided by your Dodo webhook configuration for renewal and plan-change events.

```js
const c = customer(webhook);
send("subscription_renewed", "MantrixFlow subscription renewed", {
  customer_name: c.name,
  plan_name: valueAt(webhook, ["data.product.name", "data.plan.name"], "MantrixFlow plan"),
  renewed_at: valueAt(webhook, ["data.renewed_at", "created_at"], new Date().toISOString()),
  next_billing_date: valueAt(webhook, ["data.next_billing_date", "data.current_period_end"], "your next renewal date"),
  billing_url: BILLING_URL
}, c);
```

```js
const c = customer(webhook);
send("subscription_plan_changed", "MantrixFlow plan changed", {
  customer_name: c.name,
  old_plan_name: valueAt(webhook, ["data.previous_product.name", "data.old_plan.name"], "previous plan"),
  new_plan_name: valueAt(webhook, ["data.product.name", "data.new_plan.name"], "new plan"),
  effective_date: valueAt(webhook, ["data.effective_date", "created_at"], new Date().toISOString()),
  billing_url: BILLING_URL
}, c);
```

## Cancellation, Hold, and Expiry

```js
const c = customer(webhook);
send("subscription_cancelled", "MantrixFlow subscription cancelled", {
  customer_name: c.name,
  plan_name: valueAt(webhook, ["data.product.name", "data.plan.name"], "MantrixFlow plan"),
  access_until: valueAt(webhook, ["data.current_period_end", "data.cancel_at"], "the end of the current period"),
  billing_url: BILLING_URL
}, c);
```

```js
const c = customer(webhook);
send("subscription_on_hold", "MantrixFlow subscription on hold", {
  customer_name: c.name,
  plan_name: valueAt(webhook, ["data.product.name", "data.plan.name"], "MantrixFlow plan"),
  reason: valueAt(webhook, ["data.reason", "data.status"], "Billing needs attention."),
  billing_url: BILLING_URL
}, c);
```

```js
const c = customer(webhook);
send("subscription_expired", "MantrixFlow subscription expired", {
  customer_name: c.name,
  plan_name: valueAt(webhook, ["data.product.name", "data.plan.name"], "MantrixFlow plan"),
  expired_at: valueAt(webhook, ["data.expired_at", "data.current_period_end"], new Date().toISOString()),
  billing_url: BILLING_URL
}, c);
```

## Refund, Dispute, and Invoice

```js
const c = customer(webhook);
const m = money(webhook);
send("refund_succeeded", "Refund completed for MantrixFlow", {
  customer_name: c.name,
  amount: m.amount,
  currency: m.currency,
  refund_id: valueAt(webhook, ["data.refund_id", "data.id", "id"]),
  payment_id: valueAt(webhook, ["data.payment_id", "data.payment.id"], "")
}, c);
```

```js
const m = money(webhook);
send("dispute_opened", "Dispute opened for MantrixFlow payment", {
  amount: m.amount,
  currency: m.currency,
  payment_id: valueAt(webhook, ["data.payment_id", "data.payment.id"], ""),
  customer_email: customer(webhook).email,
  dashboard_url: valueAt(webhook, ["data.dashboard_url"], "https://app.dodopayments.com/")
}, { email: "billing-alerts@mantrixflow.com", name: "MantrixFlow Billing" });
```

```js
const c = customer(webhook);
const m = money(webhook);
send("invoice_available", "Your MantrixFlow invoice is available", {
  customer_name: c.name,
  invoice_number: valueAt(webhook, ["data.invoice_number", "data.invoice.id", "data.id"], ""),
  amount: m.amount,
  currency: m.currency,
  invoice_url: valueAt(webhook, ["data.invoice_url", "data.receipt_url"], BILLING_URL)
}, c);
```

## Production Notes

- Test at least `payment.succeeded`, `payment.failed`, and `subscription.active` in Dodo test mode before production.
- Keep `templateId` values in Dodo transformations synced with AutoSend template IDs.
- Do not expose API keys in transformation logs.
- If Dodo payload names differ from these examples, update only the `valueAt` path list, not the AutoSend payload shape.
