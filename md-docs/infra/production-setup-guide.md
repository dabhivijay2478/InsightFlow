# MantrixFlow Production Setup Guide

This guide is for the split-repo production deployment model:

- Infra repo: `dabhivijay2478/mantrixflow-infra`, branch `main`.
- Frontend repo: `dabhivijay2478/InsightFlow-app`, branch `production`, deployed by Vercel.
- Go API repo: `dabhivijay2478/cloud.api.mantrixflow.com`, branch `production`.
- ELT repo: `dabhivijay2478/cloud.api.etl.server.mantrixflow.com`, branch `production`.

Manual setup count: **18 one-time steps**. Of those, **8 are AWS actions** when each IAM role is counted separately. After that, production deploy is merge-driven.

## 1. Install Admin Tools

Install these on the admin machine used for bootstrap:

```bash
brew install awscli terraform node
npm install -g aws-cdk
brew install --cask docker
```

Minimum versions:

- Node.js 20+
- Terraform 1.10+
- AWS CDK v2
- Docker Desktop running

## 2. Configure AWS Admin Access

Use a temporary admin or break-glass identity only for bootstrap.

```bash
aws configure
aws sts get-caller-identity
export AWS_REGION=ap-south-1
export AWS_ACCOUNT_ID=$(aws sts get-caller-identity --query Account --output text)
```

Do not use long-lived AWS access keys in GitHub Actions. The workflows use GitHub OIDC roles.

## 3. Create The Free Terraform S3 Backend

Terraform is self-managed and free. It uses S3 state, not Terraform Cloud/HCP.

```bash
aws s3 mb s3://mantrixflow-tfstate --region ap-south-1
aws s3api put-bucket-versioning \
  --bucket mantrixflow-tfstate \
  --versioning-configuration Status=Enabled
aws s3api put-public-access-block \
  --bucket mantrixflow-tfstate \
  --public-access-block-configuration BlockPublicAcls=true,IgnorePublicAcls=true,BlockPublicPolicy=true,RestrictPublicBuckets=true
aws s3api put-bucket-encryption \
  --bucket mantrixflow-tfstate \
  --server-side-encryption-configuration '{"Rules":[{"ApplyServerSideEncryptionByDefault":{"SSEAlgorithm":"AES256"}}]}'
```

The infra Terraform backend uses `use_lockfile = true`, so no paid locking service is needed.

## 4. Bootstrap CDK

Run once per AWS account and region:

```bash
cd apps/mantrixflow-infra/cdk
cdk bootstrap aws://$AWS_ACCOUNT_ID/$AWS_REGION
```

This creates the CDK toolkit bucket and roles needed by later CDK deploys.

## 5. Create Cloudflare API Token

Cloudflare must be authoritative DNS for `mantrixflow.com`.

Create a Cloudflare API token with:

- Zone: read
- DNS: edit
- Zone settings: edit
- Rulesets: edit

Record these values for infra repo secrets:

- `CLOUDFLARE_API_TOKEN`
- `CLOUDFLARE_ZONE_ID`

## 6. Create Separate Live Supabase And Unosend

Create a separate live Supabase project. Do not reuse the local/dev Supabase project or copied local database credentials.

Collect:

- `DATABASE_URL`
- `DATABASE_DIRECT_URL`
- `SUPABASE_URL`
- `SUPABASE_ANON_KEY`
- `SUPABASE_SERVICE_ROLE_KEY`
- `SUPABASE_JWT_SECRET`

Enable Realtime for production tables as needed by the app.

Create a separate live Unosend setup for production email. Do not reuse local/dev Unosend keys or template IDs unless that workspace is intentionally production.

Collect:

- `UNOSEND_API_KEY`
- `UNOSEND_FROM`
- `UNOSEND_LOGO_URL`
- production template IDs for invite, pipeline lifecycle, trial, payment, digest, and onboarding emails

Use the corrected template variable names from [production env mapping](production-env-mapping.md).

## 7. Create GitHub OIDC Provider In AWS

Create the GitHub OIDC provider once in the AWS account.

```bash
aws iam create-open-id-connect-provider \
  --url https://token.actions.githubusercontent.com \
  --client-id-list sts.amazonaws.com
```

If the CLI in your environment requires a thumbprint, add:

```bash
--thumbprint-list 6938fd4d98bab03faadb97b34396831e3780aea1
```

AWS and the current Terraform AWS provider can validate GitHub OIDC through trusted CAs, so thumbprints may be optional.

## 8. Create The Three GitHub Deploy Roles

Create these IAM roles and save their ARNs:

- `mantrixflow-infra-github-actions`
- `mantrixflow-api-github-actions`
- `mantrixflow-elt-github-actions`

Use [AWS GitHub OIDC role guide](aws-github-oidc-roles.md) for trust policies and starter permissions.

## 9. Configure Infra Repo GitHub Environment

In `dabhivijay2478/mantrixflow-infra`:

- Create environment `production-infra`.
- Restrict deployment branch to `main`.
- Add required reviewers if desired.
- Add secret `AWS_INFRA_DEPLOY_ROLE_ARN`.

Add these production secrets to the infra repo environment:

- `CLOUDFLARE_API_TOKEN`
- `CLOUDFLARE_ZONE_ID`
- live Supabase secrets
- `ENCRYPTION_KEY`
- `ELT_INTERNAL_TOKEN`
- `CALLBACK_TOKEN`
- `INTERNAL_TOKEN` optional; omit it to reuse `CALLBACK_TOKEN`
- live Unosend secrets
- Optional Dodo, Slack, and GitHub app secrets

## 10. Configure API Repo GitHub Environment

In `dabhivijay2478/cloud.api.mantrixflow.com`:

- Create environment `production-api`.
- Restrict deployment branch to `production`.
- Add secret `AWS_API_DEPLOY_ROLE_ARN`.

No app runtime secrets are stored here; runtime secrets come from SSM parameters managed by the infra repo.

## 11. Configure ELT Repo GitHub Environment

In `dabhivijay2478/cloud.api.etl.server.mantrixflow.com`:

- Create environment `production-elt`.
- Restrict deployment branch to `production`.
- Add secret `AWS_ELT_DEPLOY_ROLE_ARN`.

No app runtime secrets are stored here; runtime secrets come from SSM parameters managed by the infra repo.

## 12. Configure Branch Protection

Set these branch rules:

- Infra repo: protect `main`; require `Infrastructure CI`.
- Frontend repo: protect `production`; require frontend CI.
- API repo: protect `production`; require Go CI.
- ELT repo: protect `production`; require ELT CI.

Only merges into those branches should deploy production.

## 13. Configure Vercel Frontend

Create or update the Vercel project:

- Git repo: `dabhivijay2478/InsightFlow-app`
- Production branch: `production`
- Domain: `app.mantrixflow.com`
- Framework: Next.js
- Install command: `bun install --frozen-lockfile`
- Build command: `bun run build`

Set Vercel production environment variables from [env inventory](env-inventory.md). Do not store backend AWS secrets in Vercel.

## 14. First Infra Bootstrap Run

Merge or dispatch the infra workflow from `main`.

Expected first run behavior:

- Deploy ECR repositories.
- Initialize Terraform.
- Write SSM SecureString parameters.
- Create and validate ACM through Cloudflare.
- Stop at bootstrap image validation if API/ELT images do not exist yet.

This stop is acceptable on the very first run. It means shared prerequisites are ready.

## 15. Push First API Image Through GitHub Actions

In the API repo, run `Deploy API Production` manually:

- `build_image_tag`: `bootstrap-api`
- `push_image_only`: `true`

This builds/tests the API, pushes `mantrixflow-api:bootstrap-api`, and stops before ECS deploy.

## 16. Push First ELT Image Through GitHub Actions

In the ELT repo, run `Deploy ELT Production` manually:

- `build_image_tag`: `bootstrap-elt`
- `push_image_only`: `true`

This builds/tests ELT, pushes `mantrixflow-elt:bootstrap-elt`, and stops before ECS deploy.

## 17. Finish Infra Bootstrap

Rerun `Deploy Infrastructure` from the infra repo:

- `api_bootstrap_image_tag`: `bootstrap-api`
- `elt_bootstrap_image_tag`: `bootstrap-elt`

This creates ECS services, Service Connect, ALB, Cloudflare proxied API DNS, scaling, and logs.

## 18. Verify Production

Run:

```bash
curl -f https://api.mantrixflow.com/health
aws ecs describe-services \
  --cluster mantrixflow-cluster \
  --services mantrixflow-api-service mantrixflow-elt-service \
  --region ap-south-1
```

Then verify:

- Vercel production deployment for `app.mantrixflow.com`.
- API health check returns success.
- Go API can call ELT through `http://elt-service.mantrixflow.local:8000`.
- One small pipeline run posts a callback to `/api/v1/internal/elt-callback`.

## Normal Deploy After Bootstrap

- Infra changes: merge infra PR to `main`.
- Frontend changes: merge frontend PR to `production`; Vercel deploys.
- API changes: merge API PR to `production`; GitHub Actions deploys API only.
- ELT changes: merge ELT PR to `production`; GitHub Actions deploys ELT only.

API and ELT do not need to merge or deploy at the same time.
