# Dodo AutoSend Transformations

These examples are for the Dodo Payments AutoSend integration. Dodo owns customer-facing payment, subscription, refund, dispute, and invoice emails. The backend should only reconcile billing state from Dodo webhooks.

## Dodo Handler Shape

Dodo transformations run inside a JavaScript handler. Paste the shared helpers, then branch by `webhook.eventType`, and always return `webhook`.

```js
function handler(webhook) {
  // Paste shared helpers above this handler, then paste the event blocks here.
  if (webhook.eventType === "payment.succeeded") {
    // payment.succeeded block
  }

  return webhook;
}
```

## Ready-To-Paste `payment.succeeded`

This is the official Dodo handler shape adapted for MantrixFlow and the current AutoSend template.

```js
/**
 * @param webhook the webhook object
 * @param webhook.method destination method. Allowed values: "POST", "PUT"
 * @param webhook.url current destination address
 * @param webhook.eventType current webhook Event Type
 * @param webhook.payload JSON payload
 * @param webhook.cancel whether to cancel dispatch of the given webhook
 */
function handler(webhook) {
  if (webhook.eventType === "payment.succeeded") {
    const p = webhook.payload.data || {};
    const customer = p.customer || {};
    const amountRaw = Number(p.amount || p.total || 0);
    const amount = amountRaw > 1000 ? (amountRaw / 100).toFixed(2) : amountRaw.toFixed(2);
    const currency = String(p.currency || "USD").toUpperCase();

    webhook.url = "https://api.autosend.com/v1/mails/send";
    webhook.payload = {
      to: {
        email: customer.email,
        name: customer.name || "there",
      },
      from: {
        email: "support@mantrixflow.com",
        name: "MantrixFlow",
      },
      subject: "Payment received for MantrixFlow",
      templateId: "A-c3684aac9fd1839d7392",
      dynamicData: {
        customer_name: customer.name || "there",
        amount,
        currency,
        payment_id: p.payment_id || p.id || "",
        receipt_url: p.receipt_url || p.invoice_url || "https://cloud.mantrixflow.com/workspace/settings",
        date: p.created_at ? new Date(p.created_at).toLocaleDateString() : new Date().toLocaleDateString(),
      },
      replyTo: {
        email: "support@mantrixflow.com",
        name: "MantrixFlow Support",
      },
      trackingOpen: true,
      trackingClick: true,
    };
  }

  return webhook;
}
```

Change only the sender email if the verified AutoSend sender is different. The `dynamicData` keys must stay snake_case because they match the current AutoSend template placeholders.

## Shared Helpers

Paste helpers at the top of each Dodo transformation if the integration editor does not support shared code.

```js
const FROM = { email: "support@mantrixflow.com", name: "MantrixFlow" };
const REPLY_TO = { email: "support@mantrixflow.com", name: "MantrixFlow Support" };
const BILLING_URL = "https://cloud.mantrixflow.com/workspace/settings";

const templates = {
  payment_succeeded: "A-c3684aac9fd1839d7392",
  payment_failed: "A-59e0d279cdb764c41e95",
  subscription_active: "A-394fbb29881d9a416302",
  subscription_renewed: "A-5f1b6453e6d6ed0fe2e7",
  subscription_plan_changed: "A-29e8495496ae69515bbe",
  subscription_cancelled: "A-4b18e2e41ce0adaa5b9b",
  subscription_on_hold: "A-7f203558a2c98cff9b5b",
  subscription_expired: "A-6f2ab20a0e24a1bc0696",
  refund_succeeded: "A-2d0a43d72d7a51fe9005",
  dispute_opened: "A-886448badef5328a784f",
  invoice_available: "A-567959b2859c430bcf68"
};

function valueAt(obj, paths, fallback = "") {
  for (const path of paths) {
    const value = path.split(".").reduce((acc, part) => (acc && acc[part] !== undefined ? acc[part] : undefined), obj);
    if (value !== undefined && value !== null && String(value).trim() !== "") return value;
  }
  return fallback;
}

function customer(event) {
  const email = valueAt(event, ["payload.data.customer.email", "payload.data.customer_email", "data.customer.email", "data.customer_email", "customer.email", "customer_email"]);
  const name = valueAt(event, ["payload.data.customer.name", "payload.data.customer_name", "data.customer.name", "data.customer_name", "customer.name", "customer_name"], "there");
  return { email, name };
}

function money(event) {
  const amountRaw = Number(valueAt(event, ["payload.data.amount", "payload.data.total", "data.amount", "data.total", "amount"], 0));
  const currency = String(valueAt(event, ["payload.data.currency", "data.currency", "currency"], "USD")).toUpperCase();
  const amount = amountRaw > 1000 ? (amountRaw / 100).toFixed(2) : amountRaw.toFixed(2);
  return { amount, currency };
}

function send(webhook, templateKey, subject, dynamicData, toOverride) {
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
send(webhook, "payment_succeeded", "Payment received for MantrixFlow", {
  customer_name: c.name,
  amount: m.amount,
  currency: m.currency,
  payment_id: valueAt(webhook, ["payload.data.payment_id", "payload.data.id", "data.payment_id", "data.id", "id"]),
  receipt_url: valueAt(webhook, ["payload.data.receipt_url", "payload.data.invoice_url", "data.receipt_url", "data.invoice_url"], BILLING_URL),
  date: new Date().toISOString()
}, c);
```

## `payment.failed`

```js
const c = customer(webhook);
const m = money(webhook);
send(webhook, "payment_failed", "Payment failed for MantrixFlow", {
  customer_name: c.name,
  amount: m.amount,
  currency: m.currency,
  payment_id: valueAt(webhook, ["payload.data.payment_id", "payload.data.id", "data.payment_id", "data.id", "id"]),
  failure_reason: valueAt(webhook, ["payload.data.failure_reason", "payload.data.error_message", "payload.data.status", "data.failure_reason", "data.error_message", "data.status"], "The payment could not be completed."),
  billing_url: BILLING_URL
}, c);
```

## `subscription.active`

```js
const c = customer(webhook);
send(webhook, "subscription_active", "Your MantrixFlow subscription is active", {
  customer_name: c.name,
  plan_name: valueAt(webhook, ["payload.data.product.name", "payload.data.plan.name", "payload.data.product_name", "data.product.name", "data.plan.name", "data.product_name"], "MantrixFlow plan"),
  billing_period: valueAt(webhook, ["payload.data.billing_period", "payload.data.interval", "data.billing_period", "data.interval"], "current billing period"),
  billing_url: BILLING_URL
}, c);
```

## Renewal or Plan Change

Use the event names provided by your Dodo webhook configuration for renewal and plan-change events.

```js
const c = customer(webhook);
send(webhook, "subscription_renewed", "MantrixFlow subscription renewed", {
  customer_name: c.name,
  plan_name: valueAt(webhook, ["payload.data.product.name", "payload.data.plan.name", "data.product.name", "data.plan.name"], "MantrixFlow plan"),
  renewed_at: valueAt(webhook, ["payload.data.renewed_at", "payload.created_at", "data.renewed_at", "created_at"], new Date().toISOString()),
  next_billing_date: valueAt(webhook, ["payload.data.next_billing_date", "payload.data.current_period_end", "data.next_billing_date", "data.current_period_end"], "your next renewal date"),
  billing_url: BILLING_URL
}, c);
```

```js
const c = customer(webhook);
send(webhook, "subscription_plan_changed", "MantrixFlow plan changed", {
  customer_name: c.name,
  old_plan_name: valueAt(webhook, ["payload.data.previous_product.name", "payload.data.old_plan.name", "data.previous_product.name", "data.old_plan.name"], "previous plan"),
  new_plan_name: valueAt(webhook, ["payload.data.product.name", "payload.data.new_plan.name", "data.product.name", "data.new_plan.name"], "new plan"),
  effective_date: valueAt(webhook, ["payload.data.effective_date", "payload.created_at", "data.effective_date", "created_at"], new Date().toISOString()),
  billing_url: BILLING_URL
}, c);
```

## Cancellation, Hold, and Expiry

```js
const c = customer(webhook);
send(webhook, "subscription_cancelled", "MantrixFlow subscription cancelled", {
  customer_name: c.name,
  plan_name: valueAt(webhook, ["payload.data.product.name", "payload.data.plan.name", "data.product.name", "data.plan.name"], "MantrixFlow plan"),
  access_until: valueAt(webhook, ["payload.data.current_period_end", "payload.data.cancel_at", "data.current_period_end", "data.cancel_at"], "the end of the current period"),
  billing_url: BILLING_URL
}, c);
```

```js
const c = customer(webhook);
send(webhook, "subscription_on_hold", "MantrixFlow subscription on hold", {
  customer_name: c.name,
  plan_name: valueAt(webhook, ["payload.data.product.name", "payload.data.plan.name", "data.product.name", "data.plan.name"], "MantrixFlow plan"),
  reason: valueAt(webhook, ["payload.data.reason", "payload.data.status", "data.reason", "data.status"], "Billing needs attention."),
  billing_url: BILLING_URL
}, c);
```

```js
const c = customer(webhook);
send(webhook, "subscription_expired", "MantrixFlow subscription expired", {
  customer_name: c.name,
  plan_name: valueAt(webhook, ["payload.data.product.name", "payload.data.plan.name", "data.product.name", "data.plan.name"], "MantrixFlow plan"),
  expired_at: valueAt(webhook, ["payload.data.expired_at", "payload.data.current_period_end", "data.expired_at", "data.current_period_end"], new Date().toISOString()),
  billing_url: BILLING_URL
}, c);
```

## Refund, Dispute, and Invoice

```js
const c = customer(webhook);
const m = money(webhook);
send(webhook, "refund_succeeded", "Refund completed for MantrixFlow", {
  customer_name: c.name,
  amount: m.amount,
  currency: m.currency,
  refund_id: valueAt(webhook, ["payload.data.refund_id", "payload.data.id", "data.refund_id", "data.id", "id"]),
  payment_id: valueAt(webhook, ["payload.data.payment_id", "payload.data.payment.id", "data.payment_id", "data.payment.id"], "")
}, c);
```

```js
const m = money(webhook);
send(webhook, "dispute_opened", "Dispute opened for MantrixFlow payment", {
  amount: m.amount,
  currency: m.currency,
  payment_id: valueAt(webhook, ["payload.data.payment_id", "payload.data.payment.id", "data.payment_id", "data.payment.id"], ""),
  customer_email: customer(webhook).email,
  dashboard_url: valueAt(webhook, ["payload.data.dashboard_url", "data.dashboard_url"], "https://app.dodopayments.com/")
}, { email: "billing-alerts@mantrixflow.com", name: "MantrixFlow Billing" });
```

```js
const c = customer(webhook);
const m = money(webhook);
send(webhook, "invoice_available", "Your MantrixFlow invoice is available", {
  customer_name: c.name,
  invoice_number: valueAt(webhook, ["payload.data.invoice_number", "payload.data.invoice.id", "payload.data.id", "data.invoice_number", "data.invoice.id", "data.id"], ""),
  amount: m.amount,
  currency: m.currency,
  invoice_url: valueAt(webhook, ["payload.data.invoice_url", "payload.data.receipt_url", "data.invoice_url", "data.receipt_url"], BILLING_URL)
}, c);
```

## Production Notes

- Test at least `payment.succeeded`, `payment.failed`, and `subscription.active` in Dodo test mode before production.
- Keep `templateId` values in Dodo transformations synced with AutoSend template IDs.
- Do not expose API keys in transformation logs.
- If Dodo payload names differ from these examples, update only the `valueAt` path list, not the AutoSend payload shape.
