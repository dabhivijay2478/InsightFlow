# Production Env Mapping

This maps the provided frontend, Go API, and ELT env blocks into the production deployment model. Values are redacted on purpose. The provided env blocks are treated as local/dev only; live must use a separate Supabase project and a separate production Unosend setup.

## Immediate Security Reset

Several real-looking local/dev credentials were pasted while preparing this mapping. Before using production, rotate or replace any credential that could reach a real service:

- Supabase database password, service role key, and JWT secret.
- Go/ELT internal tokens and callback token.
- Encryption material.
- Google Fonts API key, or restrict it to the production project.
- Unosend API key.
- Dodo API key and webhook secret.
- Slack client secret and signing secret.
- GitHub App private key and webhook secret.
- Any test auth token that was generated from the same Supabase project.

The Supabase anon key is designed to be public, but it should still only be used with correct RLS policies.

## Vercel Production Env

Set these in the Vercel project for `dabhivijay2478/InsightFlow-app`, production branch `production`:

```bash
NEXT_PUBLIC_SUPABASE_URL=<supabase-project-url>
NEXT_PUBLIC_SUPABASE_ANON_KEY=<supabase-anon-key>
NEXT_PUBLIC_API_URL=https://api.mantrixflow.com
NEXT_PUBLIC_APP_URL=https://app.mantrixflow.com
NEXT_PUBLIC_SITE_URL=https://app.mantrixflow.com
```

Optional Vercel variables:

```bash
GOOGLE_FONTS_API_KEY=<restricted-google-fonts-key>
SLACK_PROXY_TARGET_URL=https://api.mantrixflow.com
```

Do not put these in Vercel:

- `SUPABASE_SERVICE_ROLE_KEY`
- `INTERNAL_TOKEN`
- `NEXT_PUBLIC_PYTHON_SERVICE_URL`
- `TEST_*`
- `NEON_RDS_*`
- `P2P_LOAD_ROW_COUNT`
- localhost or ngrok URLs
- local Ollama/OpenRouter demo variables unless you intentionally run server-side AI features in Vercel

The browser-facing app should call the Go API only.

## Infra Repo Secrets

Set these in `dabhivijay2478/mantrixflow-infra`, environment `production-infra`. Terraform writes them to SSM SecureString for ECS.

Required:

```bash
AWS_INFRA_DEPLOY_ROLE_ARN=<aws-oidc-role-arn>
CLOUDFLARE_API_TOKEN=<cloudflare-token>
CLOUDFLARE_ZONE_ID=<cloudflare-zone-id>
DATABASE_URL=<supabase-transaction-pooler-url>
DATABASE_DIRECT_URL=<supabase-direct-or-session-url>
DIRECT_URL=<same-value-as-database-direct-url>
SUPABASE_URL=<supabase-project-url>
SUPABASE_ANON_KEY=<supabase-anon-key>
SUPABASE_SERVICE_ROLE_KEY=<rotated-service-role-key>
SUPABASE_JWT_SECRET=<supabase-jwt-secret>
ENCRYPTION_KEY=<new-shared-fernet-or-64-hex-key>
ELT_INTERNAL_TOKEN=<new-random-token>
CALLBACK_TOKEN=<new-random-token>
INTERNAL_TOKEN=<same-as-callback-token-or-dedicated-internal-token>
```

Encryption rule:

- Generate a new value. Do not reuse the local internal token.
- The Go API receives this as `ENCRYPTION_MASTER_KEY`.
- The ELT service receives the same value as `ENCRYPTION_KEY`.
- It must be either 64 hex characters or a 44-character Fernet base64 key so both services can use it safely.

Token rule:

- Use `ELT_INTERNAL_TOKEN`, not `ETL_INTERNAL_TOKEN`.
- `ELT_INTERNAL_TOKEN`, `CALLBACK_TOKEN`, and encryption material must be different values.
- `INTERNAL_TOKEN` should normally match `CALLBACK_TOKEN` because the Go middleware validates internal callback routes against the callback token first.

## Infra Optional Integration Secrets

Use these if production billing, email, Slack, or GitHub App integration should be enabled:

```bash
DODO_PAYMENTS_API_KEY=<rotated-dodo-api-key>
DODO_WEBHOOK_SECRET=<rotated-dodo-webhook-secret>
DODO_PRODUCT_GROWTH_MONTHLY=<product-id>
DODO_PRODUCT_GROWTH_ANNUAL=<product-id>
DODO_PRODUCT_PRO_MONTHLY=<product-id>
DODO_PRODUCT_PRO_ANNUAL=<product-id>
DODO_PRODUCT_ENTERPRISE=<product-id-or-blank>
DODO_PRODUCT_COLLECTION_ID=<collection-id>
DODO_ROWS_DELIVERED_METER_ID=<meter-id>
DODO_API_CALLS_METER_ID=<meter-id-or-blank>
DODO_ROWS_DELIVERED_EVENT_NAME=rows_delivered
DODO_API_CALLS_USAGE_BILLING_ENABLED=false
DODO_API_CALLS_EVENT_NAME=api_call
UNOSEND_API_KEY=<rotated-unosend-api-key>
UNOSEND_LOGO_URL=<logo-url>
UNOSEND_FROM=support@mantrixflow.com
UNOSEND_TEMPLATE_ORG_INVITE=<template-id>
UNOSEND_TEMPLATE_PIPELINE_RUN_FAILED=<template-id>
UNOSEND_TEMPLATE_PIPELINE_RECOVERED=<template-id>
UNOSEND_TEMPLATE_PIPELINE_DISABLED=<template-id>
UNOSEND_TEMPLATE_FIRST_SUCCESS=<template-id>
UNOSEND_TEMPLATE_INCREMENTAL_INITIAL_COMPLETE=<template-id>
UNOSEND_TEMPLATE_PIPELINE_PARTIAL_SUCCESS=<template-id>
UNOSEND_TEMPLATE_INCREMENTAL_SETUP_COMPLETE=<template-id>
UNOSEND_TEMPLATE_MEMBER_REMOVED=<template-id>
UNOSEND_TEMPLATE_TRIAL_STARTED=<template-id>
UNOSEND_TEMPLATE_TRIAL_ENDS_7_DAYS=<template-id>
UNOSEND_TEMPLATE_TRIAL_ENDS_1_DAY=<template-id>
UNOSEND_TEMPLATE_TRIAL_EXPIRED=<template-id>
UNOSEND_TEMPLATE_PAYMENT_FAILED=<template-id>
UNOSEND_TEMPLATE_WEEKLY_DIGEST=<template-id>
UNOSEND_TEMPLATE_ONBOARDING_DAY3_NUDGE=<template-id>
UNOSEND_TEMPLATE_ONBOARDING_DAY7_NUDGE=<template-id>
UNOSEND_TEMPLATE_PIPELINE_QUEUED=<template-id>
UNOSEND_TEMPLATE_PIPELINE_STARTING=<template-id>
SLACK_CLIENT_ID=<slack-client-id>
SLACK_CLIENT_SECRET=<rotated-slack-client-secret>
SLACK_SIGNING_SECRET=<rotated-slack-signing-secret>
SLACK_OAUTH_REDIRECT_BASE_URL=https://app.mantrixflow.com
GH_APP_ID=<github-app-id>
GH_APP_SLUG=<github-app-slug>
GH_APP_PRIVATE_KEY=<rotated-github-app-private-key>
GH_WEBHOOK_SECRET=<rotated-github-webhook-secret>
```

Name corrections from the local env:

- Use `UNOSEND_TEMPLATE_INCREMENTAL_INITIAL_COMPLETE`, not `UNOSEND_TEMPLATE_LOG_BASED_INITIAL_COMPLETE`.
- Use `UNOSEND_TEMPLATE_INCREMENTAL_SETUP_COMPLETE`, not `UNOSEND_TEMPLATE_LOG_BASED_SETUP_COMPLETE`.
- Do not set `DODO_SKIP_WEBHOOK_VERIFY` in production.
- Do not set `ALLOW_SOURCE_DB_MUTATIONS=true` in production.
- Do not set `EMAIL_SIMULATE_DELIVERY_DOMAINS` in production unless you intentionally want simulated email delivery for QA domains.

## API Repo Secrets

Set only this in `dabhivijay2478/cloud.api.mantrixflow.com`, environment `production-api`:

```bash
AWS_API_DEPLOY_ROLE_ARN=<aws-oidc-role-arn>
```

No runtime application secrets are stored in the API repo. The ECS task reads runtime secrets from SSM.

## ELT Repo Secrets

Set only this in `dabhivijay2478/cloud.api.etl.server.mantrixflow.com`, environment `production-elt`:

```bash
AWS_ELT_DEPLOY_ROLE_ARN=<aws-oidc-role-arn>
```

No runtime application secrets are stored in the ELT repo. The ECS task reads runtime secrets from SSM.

## Production Runtime Values Set By CDK

The ECS task definitions set production-safe runtime values directly:

- API: `PORT=8080`, `ENVIRONMENT=production`, `API_PUBLIC_URL=https://api.mantrixflow.com`, `APP_WEB_URL=https://app.mantrixflow.com`, `CORS_ALLOWED_ORIGINS=https://app.mantrixflow.com`, `ELT_PYTHON_SERVICE_URL=http://elt-service.mantrixflow.local:8000`.
- API production limits: conservative pipeline limits, `PGMQ_PARALLEL_WORKERS=2`, `ELT_STAGING_DISPATCH_HEADROOM_MULTIPLIER=2`.
- ELT: `PORT=8000`, `CALLBACK_URL=https://api.mantrixflow.com/api/v1/internal/elt-callback`, `MAX_CONCURRENT_RUNS=2`, `MAX_TAPS_PER_SOURCE=3`, `STAGING_ROOT=/tmp/mxf-duckdb`, `STAGING_DISK_LIMIT_GB=50`.

Local values like `MAX_CONCURRENT_RUNS=10`, localhost URLs, ngrok URLs, and broad CORS lists should stay local.

## Exhaustive Pasted Env Coverage

Every key from the pasted frontend env is accounted for here:

| Key | Production handling |
| --- | --- |
| `NEXT_PUBLIC_SUPABASE_URL` | Vercel production env, live Supabase project |
| `NEXT_PUBLIC_SUPABASE_ANON_KEY` | Vercel production env, live Supabase project |
| `SUPABASE_SERVICE_ROLE_KEY` | Not frontend. Infra secret -> SSM for API/ELT only |
| `NEXT_PUBLIC_APP_URL` | Vercel production env, `https://app.mantrixflow.com` |
| `NEXT_PUBLIC_API_URL` | Vercel production env, `https://api.mantrixflow.com` |
| `NEXT_PUBLIC_SITE_URL` | Vercel production env, `https://app.mantrixflow.com` |
| `SLACK_PROXY_TARGET_URL` | Optional server-side Vercel env, `https://api.mantrixflow.com` |
| `GOOGLE_FONTS_API_KEY` | Optional server-side Vercel env, restricted key |
| `NEXT_PUBLIC_ALLOWED_ORIGINS` | Do not set in production frontend |
| `NEXT_PUBLIC_PYTHON_SERVICE_URL` | Do not set in production frontend |
| `NEXT_PUBLIC_PYTHON_ELT_SERVICE_URL` | Do not set in production frontend |
| `AI_MODEL_PROVIDER` | Optional server-side Vercel AI config; do not use `ollama` in Vercel |
| `AI_MODEL` | Optional server-side Vercel AI config |
| `OLLAMA_BASE_URL` | Local-only |
| `OLLAMA_API_KEY` | Local-only |
| `ANTHROPIC_API_KEY` | Optional server-side Vercel AI secret |
| `ANTHROPIC_MODEL` | Optional server-side Vercel AI config |
| `OPENROUTER_API_KEY` | Optional server-side Vercel AI secret |
| `OPENROUTER_MODEL` | Optional server-side Vercel AI config |
| `OPENROUTER_SITE_URL` | Optional server-side Vercel AI config, use app domain |
| `OPENROUTER_APP_NAME` | Optional server-side Vercel AI config |
| `AI_GATEWAY_API_KEY` | Optional server-side Vercel AI secret |
| `AI_GATEWAY_MODEL` | Optional server-side Vercel AI config |
| `OPENAI_COMPATIBLE_PROVIDER_NAME` | Optional server-side Vercel AI config |
| `OPENAI_COMPATIBLE_BASE_URL` | Optional server-side Vercel AI config |
| `OPENAI_COMPATIBLE_MODEL` | Optional server-side Vercel AI config |
| `OPENAI_COMPATIBLE_API_KEY` | Optional server-side Vercel AI secret |
| `AGENT_AI_MODEL_PROVIDER` | Optional server-side Vercel AI config |
| `AGENT_AI_MODEL` | Optional server-side Vercel AI config |
| `AI_PIPELINE_BUILDER_MODEL_PROVIDER` | Optional server-side Vercel AI config |
| `AI_PIPELINE_BUILDER_MODEL` | Optional server-side Vercel AI config |
| `PIPELINE_ASSISTANT_AI_MODEL_PROVIDER` | Optional server-side Vercel AI config |
| `PIPELINE_ASSISTANT_AI_MODEL` | Optional server-side Vercel AI config |
| `INTERNAL_TOKEN` | Optional server-side Vercel env only if a Next route calls Go internal routes; match API token |
| `TEST_ORGANIZATION_ID` | Local/test only |
| `TEST_ORG_ID` | Local/test only |
| `TEST_POSTGRES_SOURCE_IDS` | Local/test only |
| `TEST_POSTGRES_DESTINATION_IDS` | Local/test only |
| `P2P_LOAD_ROW_COUNT` | Local/test only |
| `P2P_KEEP_TEST_SCHEMAS` | Local/test only |
| `TEST_AUTH_TOKEN` | Local/test only |
| `NEON_RDS_ROW_COUNT` | Local/test only |
| `NEON_RDS_RUN_TIMEOUT_MS` | Local/test only |
| `NEON_RDS_WRITE_MODE` | Local/test only |
| `NEON_RDS_KEEP_SCHEMAS` | Local/test only |
| `NEON_RDS_KEEP_PIPELINES` | Local/test only |

Every key from the pasted Go API env is accounted for here:

| Key | Production handling |
| --- | --- |
| `DATABASE_URL` | Infra secret -> SSM |
| `DIRECT_URL` | Infra writes from `DATABASE_DIRECT_URL` as SSM alias |
| `DATABASE_DIRECT_URL` | Infra secret -> SSM |
| `SUPABASE_JWT_SECRET` | Infra secret -> SSM |
| `ENCRYPTION_MASTER_KEY` | Infra writes from `ENCRYPTION_KEY` secret |
| `ETL_PYTHON_SERVICE_URL` | Old/local spelling; do not set |
| `ELT_PYTHON_SERVICE_URL` | CDK runtime env, Service Connect URL |
| `ETL_INTERNAL_TOKEN` | Old/local typo; do not set |
| `ELT_INTERNAL_TOKEN` | Infra secret -> SSM |
| `API_PUBLIC_URL` | CDK runtime env, `https://api.mantrixflow.com` |
| `APP_WEB_URL` | CDK runtime env, `https://app.mantrixflow.com` |
| `PUBLIC_APP_URL` | Optional alias; production example sets app domain |
| `SLACK_OAUTH_REDIRECT_BASE_URL` | Infra secret -> SSM when Slack is enabled |
| `SUPABASE_URL` | Infra secret -> SSM |
| `SUPABASE_SERVICE_ROLE_KEY` | Infra secret -> SSM |
| `SUPABASE_ANON_KEY` | Infra secret -> SSM |
| `CORS_ALLOWED_ORIGINS` | CDK runtime env, app domain only |
| `PORT` | CDK runtime env, `8080` |
| `LOG_LEVEL` | CDK runtime env, `info` |
| `CALLBACK_TOKEN` | Infra secret -> SSM |
| `INTERNAL_TOKEN` | Infra writes to SSM, defaults to callback token when unset |
| `UNOSEND_API_KEY` | Infra secret -> SSM, live Unosend only |
| `UNOSEND_LOGO_URL` | Infra secret -> SSM |
| `UNOSEND_FROM` | Infra secret -> SSM |
| `UNOSEND_TEMPLATE_ORG_INVITE` | Infra secret -> SSM |
| `UNOSEND_TEMPLATE_PIPELINE_RUN_FAILED` | Infra secret -> SSM |
| `UNOSEND_TEMPLATE_PIPELINE_RECOVERED` | Infra secret -> SSM |
| `UNOSEND_TEMPLATE_PIPELINE_DISABLED` | Infra secret -> SSM |
| `UNOSEND_TEMPLATE_FIRST_SUCCESS` | Infra secret -> SSM |
| `UNOSEND_TEMPLATE_LOG_BASED_INITIAL_COMPLETE` | Old/local name; use `UNOSEND_TEMPLATE_INCREMENTAL_INITIAL_COMPLETE` |
| `UNOSEND_TEMPLATE_PIPELINE_PARTIAL_SUCCESS` | Infra secret -> SSM |
| `UNOSEND_TEMPLATE_LOG_BASED_SETUP_COMPLETE` | Old/local name; use `UNOSEND_TEMPLATE_INCREMENTAL_SETUP_COMPLETE` |
| `UNOSEND_TEMPLATE_MEMBER_REMOVED` | Infra secret -> SSM |
| `UNOSEND_TEMPLATE_TRIAL_STARTED` | Infra secret -> SSM |
| `UNOSEND_TEMPLATE_TRIAL_ENDS_7_DAYS` | Infra secret -> SSM |
| `UNOSEND_TEMPLATE_TRIAL_ENDS_1_DAY` | Infra secret -> SSM |
| `UNOSEND_TEMPLATE_TRIAL_EXPIRED` | Infra secret -> SSM |
| `UNOSEND_TEMPLATE_PAYMENT_FAILED` | Infra secret -> SSM |
| `UNOSEND_TEMPLATE_WEEKLY_DIGEST` | Infra secret -> SSM |
| `UNOSEND_TEMPLATE_ONBOARDING_DAY3_NUDGE` | Infra secret -> SSM |
| `UNOSEND_TEMPLATE_ONBOARDING_DAY7_NUDGE` | Infra secret -> SSM |
| `UNOSEND_TEMPLATE_PIPELINE_QUEUED` | Infra secret -> SSM |
| `UNOSEND_TEMPLATE_PIPELINE_STARTING` | Infra secret -> SSM |
| `ALLOW_SOURCE_DB_MUTATIONS_FOR_CDC` | Local-only / not read by current config |
| `ALLOW_SOURCE_DB_MUTATIONS` | Do not set in production |
| `PIPELINE_MAX_CONCURRENT` | CDK runtime env |
| `PIPELINE_MAX_PER_HOUR` | CDK runtime env |
| `PIPELINE_MAX_PER_DAY` | CDK runtime env |
| `PIPELINE_MAX_PER_ORG_CONCURRENT` | CDK runtime env |
| `PIPELINE_MAX_PER_ORG_PER_HOUR` | CDK runtime env |
| `PIPELINE_MAX_PER_SOURCE_CONCURRENT` | CDK runtime env |
| `PGMQ_PARALLEL_WORKERS` | CDK runtime env |
| `EMAIL_SIMULATE_DELIVERY_DOMAINS` | Do not set in production |
| `ORG_INVITE_MODE` | CDK runtime env, `supabase_link` |
| `DODO_PAYMENTS_API_KEY` | Infra secret -> SSM |
| `DODO_WEBHOOK_SECRET` | Infra secret -> SSM |
| `DODO_SKIP_WEBHOOK_VERIFY` | Do not set in production |
| `DODO_PRODUCT_GROWTH_MONTHLY` | Infra secret -> SSM |
| `DODO_PRODUCT_GROWTH_ANNUAL` | Infra secret -> SSM |
| `DODO_PRODUCT_PRO_MONTHLY` | Infra secret -> SSM |
| `DODO_PRODUCT_PRO_ANNUAL` | Infra secret -> SSM |
| `DODO_PRODUCT_ENTERPRISE` | Infra secret -> SSM, optional blank |
| `ENVIRONMENT` | CDK runtime env, `production` |
| `DODO_PRODUCT_COLLECTION_ID` | Infra secret -> SSM |
| `DODO_ROWS_DELIVERED_EVENT_NAME` | CDK runtime env, `rows_delivered` |
| `DODO_ROWS_DELIVERED_METER_ID` | Infra secret -> SSM |
| `DODO_API_CALLS_USAGE_BILLING_ENABLED` | CDK runtime env, `false` |
| `DODO_API_CALLS_EVENT_NAME` | CDK runtime env, `api_call` |
| `DODO_API_CALLS_METER_ID` | Infra secret -> SSM, optional blank |
| `SLACK_CLIENT_ID` | Infra secret -> SSM |
| `SLACK_CLIENT_SECRET` | Infra secret -> SSM |
| `SLACK_SIGNING_SECRET` | Infra secret -> SSM |
| `SLACK_API_BASE_URL` | CDK runtime env, `https://slack.com/api` |
| `NEXT_PUBLIC_API_URL` | Local compatibility only; API production uses `API_PUBLIC_URL` |
| `SLACK_PROXY_TARGET_URL` | Frontend/Vercel optional, not API ECS |
| `ELT_STAGING_DISPATCH_HEADROOM_MULTIPLIER` | CDK runtime env, `2` |
| `GITHUB_APP_ID` | Infra secret -> SSM |
| `GITHUB_APP_SLUG` | Infra secret -> SSM |
| `GITHUB_APP_PRIVATE_KEY` | Infra secret -> SSM |
| `GITHUB_WEBHOOK_SECRET` | Infra secret -> SSM |
| `GITHUB_API_BASE_URL` | CDK runtime env, `https://api.github.com` |

Every key from the pasted ELT env is accounted for here:

| Key | Production handling |
| --- | --- |
| `MAX_CONCURRENT_RUNS` | CDK runtime env |
| `MAX_TAPS_PER_SOURCE` | CDK runtime env |
| `DEFAULT_SYNC_TIMEOUT_SECONDS` | CDK runtime env |
| `LOG_LEVEL` | CDK runtime env |
| `PORT` | CDK runtime env, `8000` |
| `ENCRYPTION_KEY` | Infra writes from `ENCRYPTION_KEY` secret |
| `ETL_INTERNAL_TOKEN` | Old/local typo; do not set |
| `ELT_INTERNAL_TOKEN` | Infra secret -> SSM |
| `CALLBACK_URL` | CDK runtime env, `/api/v1/internal/elt-callback` |
| `CALLBACK_TOKEN` | Infra secret -> SSM |
| `SUPABASE_URL` | Infra secret -> SSM |
| `SUPABASE_SERVICE_ROLE_KEY` | Infra secret -> SSM, required even if local ELT env omitted it |

## Sanitized Example Files

Use these as shape references only. Do not paste real secrets into the repository:

- `apps/app/.env.production.example`
- `apps/server/main-server/.env.production.example`
- `apps/server/elt-server/.env.production.example`
- `apps/server/.env.production.example`
