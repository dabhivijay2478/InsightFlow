# AutoSend Email System Plan

This plan standardizes all MantrixFlow email on AutoSend while keeping ownership clear across Supabase Auth, the Go backend, and Dodo Payments.

## Architecture

| Owner | Transport | Scope |
| --- | --- | --- |
| Supabase Auth | AutoSend SMTP | Signup confirmation, magic links, password recovery, email change, reauthentication, workspace invites that use Supabase invite links. |
| Go backend | AutoSend Send API | Product, pipeline, workspace, usage, digest, onboarding, and trial lifecycle email. |
| Dodo Payments | Official Dodo AutoSend integration | Payment, subscription, refund, dispute, and invoice/customer billing event email. |

The backend must not duplicate customer-facing Dodo payment/subscription emails. Backend Dodo webhooks continue to reconcile organization plan, subscription IDs, checkout intents, invoice/payment status, and access state.

## Backend Environment

Use `AUTOSEND_*` names for all new configuration. Legacy `UNOSEND_*` names are still read as fallbacks during migration.

```bash
AUTOSEND_API_KEY=...
AUTOSEND_API_BASE_URL=https://api.autosend.com/v1
AUTOSEND_FROM="MantrixFlow <support@mantrixflow.com>"
AUTOSEND_REPLY_TO=support@mantrixflow.com
AUTOSEND_LOGO_URL=https://d1v739xuxrzdgy.cloudfront.net/orgs/6a158503d8ca8942a075ce15/projects/6a158503d8ca8942a075ce18/media/1ce9920218c57a04ed.png

AUTOSEND_TEMPLATE_PIPELINE_CREATED=...
AUTOSEND_TEMPLATE_PIPELINE_RUN_SUCCESS=...
AUTOSEND_TEMPLATE_PIPELINE_RUN_FAILED=...
AUTOSEND_TEMPLATE_USAGE_WARNING_80=...
AUTOSEND_TEMPLATE_WEEKLY_DIGEST=...
```

The Go backend queues mail in `email_jobs` with `provider = 'autosend'`, dedupe keys, retries, and backoff. Old `provider = 'unosend'` rows remain as audit history.

## Supabase SMTP Setup

In Supabase Dashboard, configure Auth SMTP with AutoSend:

| Supabase SMTP field | Value |
| --- | --- |
| Host | `smtp.autosend.com` |
| Port | `587` or `465` |
| Username | `autosend` |
| Password | AutoSend SMTP key, not the API key |
| Sender email | Verified domain sender, for example `support@mantrixflow.com` |
| Sender name | `MantrixFlow` |

Supabase Auth templates should stay plain and deliverability-focused. Use the AutoSend design guide only lightly for auth: logo/name, one CTA, fallback link, and security footer.

## Dodo Integration Setup

Use Dodo's official AutoSend integration from the Dodo dashboard. Configure the AutoSend API key in Dodo, then add JavaScript transformations that map Dodo events into AutoSend Send API payloads:

```js
webhook.url = "https://api.autosend.com/v1/mails/send";
webhook.payload = {
  to: { email: customerEmail, name: customerName },
  from: { email: "support@mantrixflow.com", name: "MantrixFlow" },
  subject: "Payment received",
  templateId: "tpl_payment_succeeded",
  dynamicData: {}
};
```

See `dodo-autosend-transformations.md` for transformation examples.

## Duplicate Prevention

- Auth/security/billing-critical email always sends when the owner system triggers it.
- Backend uses `email_jobs.dedupe_key` for lifecycle, usage, digest, onboarding, and re-engagement emails.
- Dodo owns payment/subscription/refund/dispute customer emails. Do not queue those from backend webhooks.
- Usage threshold dedupe keys include template, organization, billing month, and recipient email.
- Pipeline lifecycle dedupe keys include run ID or pipeline ID plus event type.
- Re-engagement is sent once per user and workspace unless the dedupe strategy is intentionally reset.

## Rollout Checklist

1. Verify sender domain in AutoSend.
2. Create AutoSend templates from `autosend-template-copy.md` and `autosend-template-design-guide.md`.
3. Add all `AUTOSEND_TEMPLATE_*` IDs to backend env.
4. Configure Supabase Auth SMTP with AutoSend SMTP credentials.
5. Configure Dodo AutoSend integration and test key Dodo transformations in test mode.
6. Run backend email smoke tests with `TEST_EMAIL=you@example.com go run ./cmd/emailtest -dry-run`.
7. Send a small live test for auth, backend, and Dodo email before production rollout.

## Sources

- [Dodo AutoSend integration](https://docs.dodopayments.com/integrations/autosend)
- [AutoSend Send API](https://docs.autosend.com/api-reference/mails/send)
- [AutoSend SMTP quickstart](https://docs.autosend.com/quickstart/smtp)
- [Supabase custom SMTP](https://supabase.com/docs/guides/auth/auth-smtp)
