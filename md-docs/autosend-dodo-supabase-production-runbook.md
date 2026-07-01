# AutoSend, Dodo, And Supabase Production Runbook

This runbook covers the production setup for MantrixFlow email after moving to AutoSend.

## Ownership

| Email group | Owner | Transport | Where to configure |
| --- | --- | --- | --- |
| Supabase Auth | Supabase | AutoSend SMTP | Supabase Dashboard Auth SMTP settings |
| Product and pipeline emails | Go backend | AutoSend REST API | Backend environment and `email_jobs` worker |
| Billing/customer payment emails | Dodo Payments | Dodo AutoSend integration | Dodo Dashboard Webhooks integration |

Do not send Dodo customer payment, subscription, refund, dispute, or invoice emails from the Go backend. The backend Dodo webhook still reconciles plans, subscriptions, checkout intents, invoice rows, and payment state.

## Current AutoSend Template Count

AutoSend now has 40 MantrixFlow templates:

| Group | Count | Notes |
| --- | ---: | --- |
| Backend templates | 29 | Used by `AUTOSEND_TEMPLATE_*` env vars in the Go backend. |
| Dodo templates | 11 | Used by Dodo transformation `templateId` values. |

AutoSend's pricing page currently lists the Hobby plan template limit as 50 templates, so the current 40-template setup fits. If the dashboard still shows only 25 templates, refresh the template list, check project selection, and search for `MantrixFlow Dodo /`.

## Dodo AutoSend Setup

1. Open Dodo Dashboard.
2. Go to `Developer` -> `Webhooks`.
3. Add an endpoint or integration and choose `AutoSend`.
4. Paste the AutoSend REST API key from AutoSend `Settings` -> `API Keys`.
5. Add the JavaScript transformation from [`dodo-autosend-transformations.md`](./dodo-autosend-transformations.md).
6. In Dodo test mode, test at least:
   - `payment.succeeded`
   - `payment.failed`
   - `subscription.active`
7. Create the integration only after the test email arrives and the template variables render correctly.

The Dodo integration should send payloads to:

```txt
https://api.autosend.com/v1/mails/send
```

The payload must include `to`, `from`, `subject`, `templateId`, and `dynamicData`.

## Supabase Auth SMTP Setup

Configure Supabase Auth with AutoSend SMTP:

| Supabase field | Production value |
| --- | --- |
| Host | `smtp.autosend.com` |
| Port | `587` with STARTTLS, or `465` with implicit TLS |
| Username | `autosend` |
| Password | AutoSend SMTP key, not the REST API key |
| Sender email | A verified sender, for example `support@mantrixflow.com` |
| Sender name | `MantrixFlow` |

Important: AutoSend REST API keys and SMTP keys are different. A REST API key used as the SMTP password can cause auth email failures.

## Signup 504 Timeout Diagnosis

The Supabase log from July 1, 2026 shows `POST /signup`, `user_confirmation_requested`, then `504 request_timeout` / `context deadline exceeded` after roughly 10 seconds. Supabase Auth sends the confirmation email inside the signup request, so a slow or failing custom SMTP call can make signup time out.

Check these first:

1. AutoSend sender domain is verified with SPF and DKIM.
2. Supabase uses the AutoSend SMTP key, not the AutoSend REST API key.
3. Supabase host is exactly `smtp.autosend.com`.
4. Port and TLS mode match:
   - `587` = STARTTLS
   - `465` = implicit TLS
5. Sender email belongs to the verified AutoSend domain.
6. AutoSend Email Activity shows the attempted SMTP message.
7. Supabase Auth redirect URLs use the correct production URL, not only `localhost`.

To isolate the issue quickly, temporarily switch Supabase back to a known-good SMTP provider or Supabase default email and retry signup. If signup becomes fast again, the failure is definitely in the AutoSend SMTP configuration path.

## Production Deploy Checklist

Backend:

1. Set `AUTOSEND_API_KEY` in the backend secret store.
2. Set `AUTOSEND_API_BASE_URL=https://api.autosend.com/v1`.
3. Set `AUTOSEND_FROM="MantrixFlow <support@mantrixflow.com>"`.
4. Set all backend `AUTOSEND_TEMPLATE_*` variables from [`autosend-template-id-map.md`](./autosend-template-id-map.md).
5. Keep Dodo payment/subscription customer emails out of backend queues.

Supabase:

1. Add AutoSend SMTP settings.
2. Confirm signup, invite, magic link, recovery, and email-change delivery.
3. Configure production `Site URL` and `Redirect URLs`.

Dodo:

1. Add the official AutoSend integration.
2. Paste the Dodo transformation code and template IDs.
3. Test Dodo events in test mode.
4. Confirm backend webhook reconciliation still points to MantrixFlow's billing webhook endpoint.

Security:

1. Rotate the AutoSend REST API key because it was pasted into chat.
2. Store the new key only in the secret manager and Dodo integration settings.
3. Create a separate AutoSend SMTP key for Supabase and store it only in Supabase Auth SMTP settings.

## Sources

- [Dodo AutoSend integration](https://docs.dodopayments.com/integrations/autosend)
- [AutoSend API reference](https://docs.autosend.com/api-reference/introduction)
- [AutoSend SMTP quickstart](https://docs.autosend.com/quickstart/smtp)
- [AutoSend pricing](https://autosend.com/pricing?volume=3k)
- [Supabase custom SMTP](https://supabase.com/docs/guides/auth/auth-smtp)
