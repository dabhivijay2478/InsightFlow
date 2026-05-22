# Production Environment Inventory

This inventory was checked against the real local env files. Secret values were intentionally not copied.

For the redacted production placement of the provided frontend, API, and ELT env blocks, see [production env mapping](production-env-mapping.md).

Checked files:

- `.env`
- `apps/app/.env`
- `apps/server/elt-server/.env`
- `apps/server/.env.production.example`

Sanitized production example files were added separately for the split repos:

- `apps/app/.env.production.example`
- `apps/server/main-server/.env.production.example`
- `apps/server/elt-server/.env.production.example`

## Manual Env Rule

Do not copy a whole local `.env` file into production. Local env files include test IDs, local service URLs, and development-only flags.

Production values should be placed in one of these systems:

- Vercel production env vars for frontend runtime values.
- Infra repo GitHub environment secrets for Terraform inputs.
- AWS SSM SecureString parameters written by Terraform for ECS runtime secrets.
- API and ELT repo GitHub environment secrets only for AWS OIDC role ARNs.

Use separate live Supabase and Unosend projects/accounts for production. The local values are development credentials and should not be promoted to live.

## Root `.env`

No parseable `KEY=value` entries were found in the root `.env`.

## `apps/app/.env`

Keys present locally:

- `AGENT_AI_MODEL`
- `AGENT_AI_MODEL_PROVIDER`
- `AI_MODEL`
- `AI_MODEL_PROVIDER`
- `AI_PIPELINE_BUILDER_MODEL`
- `AI_PIPELINE_BUILDER_MODEL_PROVIDER`
- `ANTHROPIC_MODEL`
- `GOOGLE_FONTS_API_KEY`
- `INTERNAL_TOKEN`
- `NEON_RDS_KEEP_PIPELINES`
- `NEON_RDS_KEEP_SCHEMAS`
- `NEON_RDS_ROW_COUNT`
- `NEON_RDS_RUN_TIMEOUT_MS`
- `NEON_RDS_WRITE_MODE`
- `NEXT_PUBLIC_ALLOWED_ORIGINS`
- `NEXT_PUBLIC_API_URL`
- `NEXT_PUBLIC_APP_URL`
- `NEXT_PUBLIC_PYTHON_SERVICE_URL`
- `NEXT_PUBLIC_SITE_URL`
- `NEXT_PUBLIC_SUPABASE_ANON_KEY`
- `NEXT_PUBLIC_SUPABASE_URL`
- `OLLAMA_BASE_URL`
- `OPENROUTER_APP_NAME`
- `OPENROUTER_MODEL`
- `OPENROUTER_SITE_URL`
- `P2P_LOAD_ROW_COUNT`
- `PIPELINE_ASSISTANT_AI_MODEL`
- `PIPELINE_ASSISTANT_AI_MODEL_PROVIDER`
- `SLACK_PROXY_TARGET_URL`
- `SUPABASE_SERVICE_ROLE_KEY`
- `TEST_AUTH_TOKEN`
- `TEST_ORGANIZATION_ID`
- `TEST_ORG_ID`
- `TEST_POSTGRES_DESTINATION_IDS`
- `TEST_POSTGRES_SOURCE_IDS`

Vercel production should normally include:

- `NEXT_PUBLIC_API_URL`
- `NEXT_PUBLIC_APP_URL`
- `NEXT_PUBLIC_SITE_URL`
- `NEXT_PUBLIC_SUPABASE_URL`
- `NEXT_PUBLIC_SUPABASE_ANON_KEY`

Optional Vercel production values:

- `GOOGLE_FONTS_API_KEY`
- `SLACK_PROXY_TARGET_URL`

Do not set these in production unless a specific production feature requires them:

- `SUPABASE_SERVICE_ROLE_KEY`
- `NEXT_PUBLIC_PYTHON_SERVICE_URL`
- `TEST_*`
- `NEON_RDS_*`
- `P2P_LOAD_ROW_COUNT`
- local AI provider test variables
- commented optional AI provider secrets such as `ANTHROPIC_API_KEY`, `OPENROUTER_API_KEY`, `AI_GATEWAY_API_KEY`, and `OPENAI_COMPATIBLE_API_KEY` unless you intentionally enable those server-side Vercel features

Browser code must call the Go API only. It should not call the internal ELT service directly.

## `apps/server/elt-server/.env`

Keys present locally:

- `CALLBACK_TOKEN`
- `CALLBACK_URL`
- `DEFAULT_SYNC_TIMEOUT_SECONDS`
- `ENCRYPTION_KEY`
- `ETL_INTERNAL_TOKEN`
- `LOG_LEVEL`
- `MAX_CONCURRENT_RUNS`
- `MAX_TAPS_PER_SOURCE`
- `PORT`
- `SUPABASE_URL`

Production notes:

- The Python ELT code expects `ELT_INTERNAL_TOKEN`. The local file contains `ETL_INTERNAL_TOKEN`, which is a typo for production. Terraform and CDK use `ELT_INTERNAL_TOKEN`.
- Production ELT also receives `SUPABASE_SERVICE_ROLE_KEY` from SSM, even though it is not present in the local ELT `.env`.
- `CALLBACK_URL` is set by CDK to `https://api.mantrixflow.com/api/v1/internal/elt-callback`.

## `apps/server/.env.production.example`

Keys present in the production example:

- `API_PUBLIC_URL`
- `APP_WEB_URL`
- `BILLING_GRACE_PERIOD_DAYS`
- `CALLBACK_TOKEN`
- `CORS_ALLOWED_ORIGINS`
- `DATABASE_DIRECT_URL`
- `DATABASE_URL`
- `DEFAULT_SYNC_TIMEOUT_SECONDS`
- `DIRECT_URL`
- `DODO_API_CALLS_EVENT_NAME`
- `DODO_API_CALLS_METER_ID`
- `DODO_API_CALLS_USAGE_BILLING_ENABLED`
- `DODO_PAYMENTS_API_KEY`
- `DODO_PRODUCT_COLLECTION_ID`
- `DODO_PRODUCT_ENTERPRISE`
- `DODO_PRODUCT_GROWTH_ANNUAL`
- `DODO_PRODUCT_GROWTH_MONTHLY`
- `DODO_PRODUCT_PRO_ANNUAL`
- `DODO_PRODUCT_PRO_MONTHLY`
- `DODO_ROWS_DELIVERED_EVENT_NAME`
- `DODO_ROWS_DELIVERED_METER_ID`
- `DODO_WEBHOOK_SECRET`
- `ELT_PYTHON_SERVICE_URL`
- `ELT_STAGING_DISPATCH_HEADROOM_MULTIPLIER`
- `ENCRYPTION_KEY`
- `ENCRYPTION_MASTER_KEY`
- `ELT_INTERNAL_TOKEN`
- `ENVIRONMENT`
- `GITHUB_API_BASE_URL`
- `GITHUB_APP_ID`
- `GITHUB_APP_PRIVATE_KEY`
- `GITHUB_APP_SLUG`
- `GITHUB_WEBHOOK_SECRET`
- `INTERNAL_TOKEN`
- `LOG_LEVEL`
- `MAX_CONCURRENT_RUNS`
- `MAX_TAPS_PER_SOURCE`
- `ORG_INVITE_MODE`
- `PGMQ_PARALLEL_WORKERS`
- `PIPELINE_MAX_CONCURRENT`
- `PIPELINE_MAX_PER_DAY`
- `PIPELINE_MAX_PER_HOUR`
- `PIPELINE_MAX_PER_ORG_CONCURRENT`
- `PIPELINE_MAX_PER_ORG_PER_HOUR`
- `PIPELINE_MAX_PER_SOURCE_CONCURRENT`
- `PORT`
- `PUBLIC_APP_URL`
- `SLACK_API_BASE_URL`
- `SLACK_CLIENT_ID`
- `SLACK_CLIENT_SECRET`
- `SLACK_OAUTH_REDIRECT_BASE_URL`
- `SLACK_SIGNING_SECRET`
- `STAGING_DISK_LIMIT_GB`
- `STAGING_ROOT`
- `SUPABASE_ANON_KEY`
- `SUPABASE_JWT_SECRET`
- `SUPABASE_SERVICE_ROLE_KEY`
- `SUPABASE_URL`
- `UNOSEND_API_KEY`
- `UNOSEND_FROM`
- `UNOSEND_LOGO_URL`
- `UNOSEND_TEMPLATE_FIRST_SUCCESS`
- `UNOSEND_TEMPLATE_INCREMENTAL_INITIAL_COMPLETE`
- `UNOSEND_TEMPLATE_INCREMENTAL_SETUP_COMPLETE`
- `UNOSEND_TEMPLATE_MEMBER_REMOVED`
- `UNOSEND_TEMPLATE_ONBOARDING_DAY3_NUDGE`
- `UNOSEND_TEMPLATE_ONBOARDING_DAY7_NUDGE`
- `UNOSEND_TEMPLATE_ORG_INVITE`
- `UNOSEND_TEMPLATE_PAYMENT_FAILED`
- `UNOSEND_TEMPLATE_PIPELINE_DISABLED`
- `UNOSEND_TEMPLATE_PIPELINE_PARTIAL_SUCCESS`
- `UNOSEND_TEMPLATE_PIPELINE_QUEUED`
- `UNOSEND_TEMPLATE_PIPELINE_RECOVERED`
- `UNOSEND_TEMPLATE_PIPELINE_RUN_FAILED`
- `UNOSEND_TEMPLATE_PIPELINE_STARTING`
- `UNOSEND_TEMPLATE_TRIAL_ENDS_1_DAY`
- `UNOSEND_TEMPLATE_TRIAL_ENDS_7_DAYS`
- `UNOSEND_TEMPLATE_TRIAL_EXPIRED`
- `UNOSEND_TEMPLATE_TRIAL_STARTED`
- `UNOSEND_TEMPLATE_WEEKLY_DIGEST`

This example is the closest human-readable union of production variables, but infra Terraform and CDK are the executable production source of truth. It has been aligned with the deploy config so production uses `PORT=8080` and `ELT_INTERNAL_TOKEN`.

## Infra Repo GitHub Environment Secrets

Create these in `dabhivijay2478/mantrixflow-infra`, environment `production-infra`:

- `AWS_INFRA_DEPLOY_ROLE_ARN`
- `CLOUDFLARE_API_TOKEN`
- `CLOUDFLARE_ZONE_ID`
- `DATABASE_URL`
- `SUPABASE_URL`
- `SUPABASE_ANON_KEY`
- `SUPABASE_SERVICE_ROLE_KEY`
- `SUPABASE_JWT_SECRET`
- `ENCRYPTION_KEY`
- `ELT_INTERNAL_TOKEN`
- `CALLBACK_TOKEN`
- `INTERNAL_TOKEN`, optional; when omitted Terraform writes the callback token value

Recommended optional secrets:

- `DATABASE_DIRECT_URL`
- `DODO_PAYMENTS_API_KEY`
- `DODO_WEBHOOK_SECRET`
- `DODO_PRODUCT_GROWTH_MONTHLY`
- `DODO_PRODUCT_GROWTH_ANNUAL`
- `DODO_PRODUCT_PRO_MONTHLY`
- `DODO_PRODUCT_PRO_ANNUAL`
- `DODO_PRODUCT_ENTERPRISE`
- `DODO_PRODUCT_COLLECTION_ID`
- `DODO_ROWS_DELIVERED_METER_ID`
- `DODO_API_CALLS_METER_ID`
- `UNOSEND_API_KEY`
- `UNOSEND_LOGO_URL`
- `UNOSEND_FROM`
- `UNOSEND_TEMPLATE_ORG_INVITE`
- `UNOSEND_TEMPLATE_PIPELINE_RUN_FAILED`
- `UNOSEND_TEMPLATE_PIPELINE_RECOVERED`
- `UNOSEND_TEMPLATE_PIPELINE_DISABLED`
- `UNOSEND_TEMPLATE_FIRST_SUCCESS`
- `UNOSEND_TEMPLATE_INCREMENTAL_INITIAL_COMPLETE`
- `UNOSEND_TEMPLATE_PIPELINE_PARTIAL_SUCCESS`
- `UNOSEND_TEMPLATE_PIPELINE_QUEUED`
- `UNOSEND_TEMPLATE_PIPELINE_STARTING`
- `UNOSEND_TEMPLATE_INCREMENTAL_SETUP_COMPLETE`
- `UNOSEND_TEMPLATE_MEMBER_REMOVED`
- `UNOSEND_TEMPLATE_TRIAL_STARTED`
- `UNOSEND_TEMPLATE_TRIAL_ENDS_7_DAYS`
- `UNOSEND_TEMPLATE_TRIAL_ENDS_1_DAY`
- `UNOSEND_TEMPLATE_TRIAL_EXPIRED`
- `UNOSEND_TEMPLATE_PAYMENT_FAILED`
- `UNOSEND_TEMPLATE_WEEKLY_DIGEST`
- `UNOSEND_TEMPLATE_ONBOARDING_DAY3_NUDGE`
- `UNOSEND_TEMPLATE_ONBOARDING_DAY7_NUDGE`
- `SLACK_SIGNING_SECRET`
- `SLACK_CLIENT_ID`
- `SLACK_CLIENT_SECRET`
- `SLACK_BOT_TOKEN`
- `SLACK_OAUTH_REDIRECT_BASE_URL`
- `GH_APP_ID`
- `GH_APP_SLUG`
- `GH_APP_PRIVATE_KEY`
- `GH_WEBHOOK_SECRET`
- `API_CERTIFICATE_ARN`, only as a fallback override if Terraform output cannot be used.

## API Repo GitHub Environment Secrets

Create this in `dabhivijay2478/cloud.api.mantrixflow.com`, environment `production-api`:

- `AWS_API_DEPLOY_ROLE_ARN`

The API deploy workflow does not store application runtime secrets. ECS reads them from SSM.

## ELT Repo GitHub Environment Secrets

Create this in `dabhivijay2478/cloud.api.etl.server.mantrixflow.com`, environment `production-elt`:

- `AWS_ELT_DEPLOY_ROLE_ARN`

The ELT deploy workflow does not store application runtime secrets. ECS reads them from SSM.

## SSM Parameters Written By Terraform

Terraform writes SecureString parameters under `/mantrixflow/production`.

Required parameters:

- `/mantrixflow/production/DATABASE_URL`
- `/mantrixflow/production/SUPABASE_URL`
- `/mantrixflow/production/SUPABASE_ANON_KEY`
- `/mantrixflow/production/SUPABASE_SERVICE_ROLE_KEY`
- `/mantrixflow/production/SUPABASE_JWT_SECRET`
- `/mantrixflow/production/ENCRYPTION_MASTER_KEY`
- `/mantrixflow/production/ENCRYPTION_KEY`
- `/mantrixflow/production/ELT_INTERNAL_TOKEN`
- `/mantrixflow/production/CALLBACK_TOKEN`
- `/mantrixflow/production/INTERNAL_TOKEN`
- `/mantrixflow/production/DATABASE_DIRECT_URL`
- `/mantrixflow/production/DIRECT_URL`

Optional parameters are also written with a single blank value when not configured. This keeps ECS task definitions stable while still letting you add integrations later.
