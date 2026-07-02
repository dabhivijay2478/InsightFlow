# AutoSend Production Deployment Guide

Use this guide when deploying MantrixFlow email to production.

## What Goes Where

| Area | What to configure |
| --- | --- |
| AutoSend Dashboard | Verified domain, API key, SMTP key, templates |
| Supabase Dashboard | Auth SMTP settings and Auth email HTML templates |
| Go backend server | `AUTOSEND_*` env vars for backend queued emails |
| Dodo Dashboard | AutoSend integration API key and transformation code |
| Frontend hosting | Supabase URL/key, API URL, site URL |

## 1. AutoSend Dashboard

Project: `MantrixFlow`

Domain:

```txt
mantrixflow.com
```

Verified sender:

```txt
no-reply@mantrixflow.com
```

Reply/support sender:

```txt
support@mantrixflow.com
```

Create two separate secrets:

| Secret | Used by | Notes |
| --- | --- | --- |
| AutoSend REST API key | Go backend and Dodo AutoSend integration | Used for `https://api.autosend.com/v1/mails/send`. |
| AutoSend SMTP key | Supabase Auth SMTP | Do not use the REST API key as SMTP password. |

## 2. Supabase Production Settings

Supabase Dashboard -> Authentication -> SMTP:

```txt
Enable custom SMTP: On
Host: smtp.autosend.com
Port: 465
Username: autosend
Password: <AutoSend SMTP key>
Sender email: no-reply@mantrixflow.com
Sender name: MantrixFlow
```

Supabase Dashboard -> Authentication -> URL Configuration:

```txt
Site URL: https://cloud.mantrixflow.com
Redirect URLs:
https://cloud.mantrixflow.com/auth/callback
https://cloud.mantrixflow.com/**
```

For local development, also keep:

```txt
http://localhost:3000/auth/callback
http://localhost:3000/**
```

Supabase Dashboard -> Authentication -> Emails:

Use the templates in [`supabase-auth-email-templates.md`](./supabase-auth-email-templates.md). Copy only the raw HTML inside each code block, not the Markdown heading or backticks.

## 3. Backend Production Env

Set these in the backend server secret store. In this repo, the full example lives in:

```txt
apps/server/main-server/.env.production.example
```

### Safe CI Rollout With A Separate Secret

The running Go API must receive variables named `AUTOSEND_*`, but the GitHub/CI secret that stores them can have a separate name. This lets you add AutoSend without editing the existing multiline production secret immediately.

Recommended new GitHub environment secret:

```txt
HETZNER_API_ENV_AUTOSEND
```

If deploying through Oracle/OCI instead of Hetzner, use:

```txt
ORACLE_API_ENV_AUTOSEND
```

Put only the AutoSend block in that separate secret:

```bash
AUTOSEND_API_KEY=<AutoSend REST API key>
AUTOSEND_API_BASE_URL=https://api.autosend.com/v1
AUTOSEND_LOGO_URL=https://d1v739xuxrzdgy.cloudfront.net/orgs/6a158503d8ca8942a075ce15/projects/6a158503d8ca8942a075ce18/media/1ce9920218c57a04ed.png
AUTOSEND_FROM=MantrixFlow <no-reply@mantrixflow.com>
AUTOSEND_REPLY_TO=support@mantrixflow.com
AUTOSEND_SMTP_HOST=smtp.autosend.com
AUTOSEND_SMTP_PORT=465
AUTOSEND_SMTP_USER=autosend
AUTOSEND_SMTP_PASS=<AutoSend SMTP key>
AUTOSEND_TEMPLATE_ORG_INVITE=A-13f385ebe4e515c5fdba
AUTOSEND_TEMPLATE_MEMBER_REMOVED=A-fe931f5dd54404c7bb14
AUTOSEND_TEMPLATE_WORKSPACE_ROLE_CHANGED=A-daa3404f2b8e99fd5a49
AUTOSEND_TEMPLATE_PIPELINE_CREATED=A-beebe2fb91a3a44ed5ed
AUTOSEND_TEMPLATE_PIPELINE_QUEUED=A-48ab3a0f90992ccfdce9
AUTOSEND_TEMPLATE_PIPELINE_STARTING=A-54ff7dac742baf2ef9c5
AUTOSEND_TEMPLATE_PIPELINE_RUN_SUCCESS=A-208bea0f6a9b177efbd6
AUTOSEND_TEMPLATE_PIPELINE_RUN_FAILED=A-a2f768b21cca086526ee
AUTOSEND_TEMPLATE_PIPELINE_PARTIAL_SUCCESS=A-35ddb7838e6afe10acc4
AUTOSEND_TEMPLATE_FIRST_SUCCESS=A-b3150f6991578e8be9cd
AUTOSEND_TEMPLATE_PIPELINE_RECOVERED=A-7fe39607039b04690897
AUTOSEND_TEMPLATE_PIPELINE_DISABLED=A-6645a104c84e5806de0a
AUTOSEND_TEMPLATE_PIPELINE_SCHEDULE_CHANGED=A-d9f57966d6384d910cdc
AUTOSEND_TEMPLATE_INCREMENTAL_SETUP_COMPLETE=A-8cfe848d77d53e70bf0a
AUTOSEND_TEMPLATE_INCREMENTAL_INITIAL_COMPLETE=A-dadb53596fd59af0d9b6
AUTOSEND_TEMPLATE_USAGE_WARNING_80=A-9d98e50ad517101c1403
AUTOSEND_TEMPLATE_USAGE_LIMIT_REACHED=A-a0159d272ea193527d5e
AUTOSEND_TEMPLATE_WEEKLY_DIGEST=A-828f9c40d0b5df1ff40c
AUTOSEND_TEMPLATE_REENGAGEMENT_14_DAYS=A-08eaa38f8e2989c3be4b
AUTOSEND_TEMPLATE_ONBOARDING_DAY3_NUDGE=A-29dfa1c5633e88a43d72
AUTOSEND_TEMPLATE_ONBOARDING_DAY7_NUDGE=A-3d107ea5bb257bd254b5
AUTOSEND_TEMPLATE_CONNECTION_CREATED=A-ac28cb59277d6534a744
AUTOSEND_TEMPLATE_CONNECTION_ERROR=A-158ad0b25cc108964cdf
AUTOSEND_TEMPLATE_PIPELINE_DELETED=A-eabb1c6d3c8602baef39
AUTOSEND_TEMPLATE_TRIAL_STARTED=A-5b250383e6f9aa3e6e5a
AUTOSEND_TEMPLATE_TRIAL_ENDS_7_DAYS=A-48833eea5afa2659fb3d
AUTOSEND_TEMPLATE_TRIAL_ENDS_1_DAY=A-6e65de00867a30b038b7
AUTOSEND_TEMPLATE_TRIAL_EXPIRED=A-6c7ccf48715ad79280f5
AUTOSEND_TEMPLATE_PAYMENT_FAILED=A-26b35ffc124523857bf4
```

Then merge it with the existing env during deployment:

```bash
printf '%s\n' "$HETZNER_API_ENV" > api.env
printf '\n%s\n' "$HETZNER_API_ENV_AUTOSEND" >> api.env
```

For OCI/Oracle deployments:

```bash
printf '%s\n' "$ORACLE_API_ENV" > api.env
printf '\n%s\n' "$ORACLE_API_ENV_AUTOSEND" >> api.env
```

This avoids touching the current production env secret while still giving the app the normal `AUTOSEND_*` runtime variables. After the rollout is stable, you can optionally fold the AutoSend block into the main `HETZNER_API_ENV` or `ORACLE_API_ENV` secret.

Required AutoSend backend env:

```bash
AUTOSEND_API_KEY=<AutoSend REST API key>
AUTOSEND_API_BASE_URL=https://api.autosend.com/v1
AUTOSEND_LOGO_URL=https://d1v739xuxrzdgy.cloudfront.net/orgs/6a158503d8ca8942a075ce15/projects/6a158503d8ca8942a075ce18/media/1ce9920218c57a04ed.png
AUTOSEND_FROM=MantrixFlow <no-reply@mantrixflow.com>
AUTOSEND_REPLY_TO=support@mantrixflow.com
```

The backend does not use SMTP for sending queued product emails, but keep these documented for operators:

```bash
AUTOSEND_SMTP_HOST=smtp.autosend.com
AUTOSEND_SMTP_PORT=465
AUTOSEND_SMTP_USER=autosend
AUTOSEND_SMTP_PASS=<AutoSend SMTP key>
```

Template IDs:

```bash
AUTOSEND_TEMPLATE_ORG_INVITE=A-13f385ebe4e515c5fdba
AUTOSEND_TEMPLATE_MEMBER_REMOVED=A-fe931f5dd54404c7bb14
AUTOSEND_TEMPLATE_WORKSPACE_ROLE_CHANGED=A-daa3404f2b8e99fd5a49
AUTOSEND_TEMPLATE_PIPELINE_CREATED=A-beebe2fb91a3a44ed5ed
AUTOSEND_TEMPLATE_PIPELINE_QUEUED=A-48ab3a0f90992ccfdce9
AUTOSEND_TEMPLATE_PIPELINE_STARTING=A-54ff7dac742baf2ef9c5
AUTOSEND_TEMPLATE_PIPELINE_RUN_SUCCESS=A-208bea0f6a9b177efbd6
AUTOSEND_TEMPLATE_PIPELINE_RUN_FAILED=A-a2f768b21cca086526ee
AUTOSEND_TEMPLATE_PIPELINE_PARTIAL_SUCCESS=A-35ddb7838e6afe10acc4
AUTOSEND_TEMPLATE_FIRST_SUCCESS=A-b3150f6991578e8be9cd
AUTOSEND_TEMPLATE_PIPELINE_RECOVERED=A-7fe39607039b04690897
AUTOSEND_TEMPLATE_PIPELINE_DISABLED=A-6645a104c84e5806de0a
AUTOSEND_TEMPLATE_PIPELINE_SCHEDULE_CHANGED=A-d9f57966d6384d910cdc
AUTOSEND_TEMPLATE_INCREMENTAL_SETUP_COMPLETE=A-8cfe848d77d53e70bf0a
AUTOSEND_TEMPLATE_INCREMENTAL_INITIAL_COMPLETE=A-dadb53596fd59af0d9b6
AUTOSEND_TEMPLATE_USAGE_WARNING_80=A-9d98e50ad517101c1403
AUTOSEND_TEMPLATE_USAGE_LIMIT_REACHED=A-a0159d272ea193527d5e
AUTOSEND_TEMPLATE_WEEKLY_DIGEST=A-828f9c40d0b5df1ff40c
AUTOSEND_TEMPLATE_REENGAGEMENT_14_DAYS=A-08eaa38f8e2989c3be4b
AUTOSEND_TEMPLATE_ONBOARDING_DAY3_NUDGE=A-29dfa1c5633e88a43d72
AUTOSEND_TEMPLATE_ONBOARDING_DAY7_NUDGE=A-3d107ea5bb257bd254b5
AUTOSEND_TEMPLATE_CONNECTION_CREATED=A-ac28cb59277d6534a744
AUTOSEND_TEMPLATE_CONNECTION_ERROR=A-158ad0b25cc108964cdf
AUTOSEND_TEMPLATE_PIPELINE_DELETED=A-eabb1c6d3c8602baef39
AUTOSEND_TEMPLATE_TRIAL_STARTED=A-5b250383e6f9aa3e6e5a
AUTOSEND_TEMPLATE_TRIAL_ENDS_7_DAYS=A-48833eea5afa2659fb3d
AUTOSEND_TEMPLATE_TRIAL_ENDS_1_DAY=A-6e65de00867a30b038b7
AUTOSEND_TEMPLATE_TRIAL_EXPIRED=A-6c7ccf48715ad79280f5
AUTOSEND_TEMPLATE_PAYMENT_FAILED=A-26b35ffc124523857bf4
```

Full template reference:

```txt
md-docs/autosend-template-id-map.md
```

## 4. Dodo Production Setup

Dodo Dashboard -> Developer -> Webhooks -> AutoSend integration:

```txt
AutoSend API key: <AutoSend REST API key>
```

Select only the Dodo events available in your dashboard, for example:

```txt
payment.succeeded
payment.failed
subscription.active
subscription.renewed
subscription.plan_changed
subscription.cancelled
subscription.on_hold
subscription.expired
refund.succeeded
dispute.opened
```

Do not select `invoice.available` if Dodo does not show it. Use `payment.succeeded` for receipt/invoice emails.

Paste transformation code from:

```txt
md-docs/dodo-autosend-transformations.md
```

Use:

```js
const FROM_EMAIL = "no-reply@mantrixflow.com";
const REPLY_EMAIL = "support@mantrixflow.com";
```

## 5. Frontend Production Env

For the app frontend deployment:

```bash
NEXT_PUBLIC_API_URL=https://cloud.api.mantrixflow.com
NEXT_PUBLIC_SUPABASE_URL=https://<production-ref>.supabase.co
NEXT_PUBLIC_SUPABASE_ANON_KEY=<production anon key>
NEXT_PUBLIC_SITE_URL=https://cloud.mantrixflow.com
```

For the marketing website:

```bash
NEXT_PUBLIC_SITE_URL=https://mantrixflow.com
NEXT_PUBLIC_DOCS_URL=https://docs.mantrixflow.com
```

## 6. Production Smoke Tests

Run these after deploy:

1. Supabase signup with a fresh email.
2. Supabase password reset.
3. Supabase magic link or OTP.
4. Backend test email dry run, then live send to an internal inbox.
5. Pipeline failure email.
6. Dodo test `payment.succeeded`.
7. Dodo test `payment.failed`.
8. Confirm AutoSend Email Activity shows the expected sender:

```txt
MantrixFlow <no-reply@mantrixflow.com>
```

## 7. Troubleshooting

| Symptom | Likely cause | Fix |
| --- | --- | --- |
| Supabase signup returns 504 | SMTP timeout | Use port `465`, SMTP key, sender `no-reply@mantrixflow.com`. |
| AutoSend shows no SMTP activity | Supabase cannot connect/authenticate | Recheck host, port, username, password, and sender domain. |
| Supabase preview shows `#` headings or backticks | Markdown pasted into HTML field | Paste only raw HTML from inside the code block. |
| Dodo email sends wrong template | Wrong `templateId` in transformation | Copy IDs from `autosend-template-id-map.md`. |
| Backend emails do not send | Missing `AUTOSEND_API_KEY` or template env | Check backend env and worker logs. |

## Sources

- [AutoSend API reference](https://docs.autosend.com/api-reference/introduction)
- [AutoSend SMTP quickstart](https://docs.autosend.com/quickstart/smtp)
- [Dodo AutoSend integration](https://docs.dodopayments.com/integrations/autosend)
- [Supabase Auth email templates](https://supabase.com/docs/guides/auth/auth-email-templates)
