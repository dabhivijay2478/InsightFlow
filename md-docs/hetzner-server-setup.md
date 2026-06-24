# Hetzner Server Setup Guide

This guide explains how to create the MantrixFlow production server through
GitHub Actions CI/CD using Terraform and Hetzner Cloud.

Do not manually copy application code to the server. The infra workflow creates
and bootstraps the server; the API and ELT workflows deploy containers.

## Recommended Server

Start with **CX33** if cost is the priority.

| Type | CPU | RAM | Disk | Traffic | Approx price | Use |
| --- | ---: | ---: | ---: | ---: | --- |
| CX23 | 2 vCPU | 4 GB | 40 GB | 20 TB | $6.49/mo | Too small for API + ELT production |
| CX33 | 4 vCPU | 8 GB | 80 GB | 20 TB | $8.99/mo | MVP start, 1-2 ELT pipelines |
| CX43 | 8 vCPU | 16 GB | 160 GB | 20 TB | $18.49/mo | Better for 2-3 ELT pipelines |
| CX53 | 16 vCPU | 32 GB | 320 GB | 20 TB | higher | Later scale-up |

For the current MVP, choose:

```text
Server type: CX33
Location: Helsinki
Image: Ubuntu 24.04 LTS
Architecture: Intel/AMD x86_64
```

Use **CX43** if you want safer memory headroom for ELT. CX33 is cheaper, but
keep concurrency conservative.

## Important First-Time Flow

Use CI/CD for the server:

1. Create Hetzner, Tigris, Cloudflare, and GHCR secrets in GitHub.
2. Push or manually run the infra workflow on `main`.
3. Terraform creates the Hetzner CX33 in Helsinki.
4. GitHub Actions bootstraps Docker, Caddy, firewall, and Tigris backup.
5. ELT and API workflows deploy onto the current server resolved from Hetzner.

The default server name is:

```text
mantrixflow-production
```

API and ELT deploy workflows find the server by this name each run. If you
change it, set `HETZNER_SERVER_NAME` in all three repos.

## 1. Create Project

1. Open Hetzner Cloud Console.
2. Create or open project: `MantrixFlow`.
3. Go to **Security > API Tokens**.
4. Create a read/write token.
5. Save it as GitHub environment secret:

```text
HETZNER_API_TOKEN
```

Add this secret in all three repositories that deploy:

```text
mantrixflow-infra
cloud.api.mantrixflow.com
cloud.api.etl.server.mantrixflow.com
```

## 2. Add SSH Key

Create an SSH key locally:

```bash
ssh-keygen -t ed25519 -f ~/.ssh/mantrixflow_hetzner -C mantrixflow-hetzner
```

In Hetzner:

1. Go to **Security > SSH Keys**.
2. Click **Add SSH key**.
3. Name: `mantrixflow-hetzner`.
4. Paste the contents of:

```bash
cat ~/.ssh/mantrixflow_hetzner.pub
```

In GitHub environment `production-hetzner`, add:

```text
HETZNER_SSH_PUBLIC_KEY   = contents of ~/.ssh/mantrixflow_hetzner.pub
HETZNER_SSH_PRIVATE_KEY  = contents of ~/.ssh/mantrixflow_hetzner
```

## 3. Server Created By Terraform

Do not create the server manually in Hetzner UI. Terraform creates:

```text
Name:        mantrixflow-production
Type:        cx33
Location:    hel1 (Helsinki)
Image:       ubuntu-24.04
Public IP:   IPv4 + IPv6
Firewall:    22, 80, 443
Labels:      application=mantrixflow, environment=production
```

Optional infra repo secrets can override the default:

```text
HETZNER_SERVER_NAME
HETZNER_SERVER_TYPE
HETZNER_LOCATION
```

For the current target, leave those unset or set:

```text
HETZNER_SERVER_NAME=mantrixflow-production
HETZNER_SERVER_TYPE=cx33
HETZNER_LOCATION=hel1
```

## 4. Tigris Setup

Create Tigris before running the first infra workflow. The full setup is in
[`tigris-storage-setup.md`](./tigris-storage-setup.md).

1. Sign in to Tigris.
2. Create bucket:

```text
mantrixflow-production
```

3. Create access keys.
4. Copy these values:

```text
TIGRIS_ACCESS_KEY_ID
TIGRIS_SECRET_ACCESS_KEY
TIGRIS_ENDPOINT
TIGRIS_BUCKET
```

Typical endpoint:

```text
https://fly.storage.tigris.dev
```

Tigris is used for:

1. Terraform state.
2. Hourly backups from `/var/mantrixflow/staging`.

Tigris is **not** used as active DuckDB storage. DuckDB staging stays on the
server SSD.

## 5. Cloudflare DNS

Do not create the DNS record manually unless you need a temporary test. The
infra workflow creates/updates:

```text
Type: A
Name: cloud.api
Value: Terraform-created Hetzner server IPv4
Proxy: DNS only for first validation
TTL: Auto
```

After Caddy has issued TLS successfully, you can enable the orange-cloud proxy
if desired.

Do not point `cloud.mantrixflow.com` here. The frontend remains separate.

## 6. GitHub Environments

Each repository should have only one environment:

```text
production-hetzner
```

Restrict the environment to the deployment branch:

| Repository | Environment branch rule |
| --- | --- |
| `mantrixflow-infra` | `main` |
| `cloud.api.mantrixflow.com` | `mantrixflow` |
| `cloud.api.etl.server.mantrixflow.com` | `mantrixflow` |

### Infra Repository Secrets

Repository: `mantrixflow-infra`

Required secrets:

| Secret | Value |
| --- | --- |
| `HETZNER_API_TOKEN` | Hetzner Cloud project API token with read/write access |
| `HETZNER_SSH_PUBLIC_KEY` | Full contents of `~/.ssh/mantrixflow_hetzner.pub` |
| `HETZNER_SSH_PRIVATE_KEY` | Full contents of `~/.ssh/mantrixflow_hetzner` |
| `CLOUDFLARE_API_TOKEN` | Cloudflare token with Zone DNS Edit and Zone Settings Edit |
| `CLOUDFLARE_ZONE_ID` | Cloudflare zone ID for `mantrixflow.com` |
| `TIGRIS_ACCESS_KEY_ID` | Tigris S3 access key |
| `TIGRIS_SECRET_ACCESS_KEY` | Tigris S3 secret key |
| `TIGRIS_ENDPOINT` | Usually `https://fly.storage.tigris.dev` |
| `TIGRIS_BUCKET` | Tigris bucket name, for example `mantrixflow-production` |

Optional infra repo secrets:

| Secret | Default | Use |
| --- | --- | --- |
| `HETZNER_SERVER_NAME` | `mantrixflow-production` | Server name API/ELT resolve dynamically |
| `HETZNER_SERVER_TYPE` | `cx33` | Use `cx43` later for more capacity |
| `HETZNER_LOCATION` | `hel1` | Helsinki region |

### API Repository Secrets

Repository: `cloud.api.mantrixflow.com`

Required secrets:

| Secret | Value |
| --- | --- |
| `HETZNER_API_TOKEN` | Same Hetzner API token |
| `HETZNER_SSH_PRIVATE_KEY` | Same private SSH key used by infra |
| `GHCR_READ_TOKEN` | GitHub classic PAT with `read:packages` |
| `INTERNAL_TOKEN` | Shared internal token, same value in API and ELT |
| `HETZNER_API_ENV` | Complete multiline production env file for the Go API |
| `GH_APP_ID` | GitHub App ID. Use `GH_*` because GitHub Actions reserves `GITHUB_*` secret names |
| `GH_APP_SLUG` | GitHub App slug from `https://github.com/apps/{slug}` |
| `GH_APP_PRIVATE_KEY` | Raw multiline GitHub App private key PEM, stored separately from `HETZNER_API_ENV` |
| `GH_WEBHOOK_SECRET` | GitHub App webhook secret |

Optional API repo secret:

| Secret | Default | Use |
| --- | --- | --- |
| `HETZNER_SERVER_NAME` | `mantrixflow-production` | Must match infra if overridden |

Use this shape for `HETZNER_API_ENV`:

```env
PORT=8080
ENVIRONMENT=production
LOG_LEVEL=info

API_PUBLIC_URL=https://cloud.api.mantrixflow.com
APP_WEB_URL=https://cloud.mantrixflow.com
PUBLIC_APP_URL=https://cloud.mantrixflow.com
CORS_ALLOWED_ORIGINS=https://cloud.mantrixflow.com

DATABASE_URL=...
DATABASE_DIRECT_URL=...
DIRECT_URL=...

SUPABASE_URL=...
SUPABASE_ANON_KEY=...
SUPABASE_SERVICE_ROLE_KEY=...

ENCRYPTION_MASTER_KEY=...
INTERNAL_TOKEN=...
CALLBACK_TOKEN=...

ELT_PYTHON_SERVICE_URL=http://mantrixflow-elt:8000

PIPELINE_MAX_CONCURRENT=1
PGMQ_PARALLEL_WORKERS=1
PIPELINE_MAX_PER_ORG_CONCURRENT=1
PIPELINE_MAX_PER_SOURCE_CONCURRENT=1
PIPELINE_MAX_PER_HOUR=...
PIPELINE_MAX_PER_DAY=...
PIPELINE_ORPHANED_RUN_MAX_AGE_SEC=...
PIPELINE_QUEUED_RUN_MAX_AGE_SEC=...

# Add enabled product integrations:
DODO_PAYMENTS_API_KEY=...
DODO_WEBHOOK_SECRET=...
DODO_PRODUCT_GROWTH_MONTHLY=...
DODO_PRODUCT_GROWTH_ANNUAL=...
DODO_PRODUCT_PRO_MONTHLY=...
DODO_PRODUCT_PRO_ANNUAL=...
SLACK_OAUTH_REDIRECT_BASE_URL=https://cloud.api.mantrixflow.com
POSTHOG_API_KEY=...
GH_API_BASE_URL=https://api.github.com
EMAIL_FROM=...
```

Keep `INTERNAL_TOKEN` inside `HETZNER_API_ENV` equal to the GitHub secret
`INTERNAL_TOKEN`.

Do not put GitHub App secrets inside `HETZNER_API_ENV`. Store them as separate
API repo environment secrets:

```text
GH_APP_ID
GH_APP_SLUG
GH_APP_PRIVATE_KEY
GH_WEBHOOK_SECRET
GH_API_BASE_URL
```

The API workflow removes stale `GITHUB_*` / `GH_*` GitHub App lines from
`HETZNER_API_ENV`, writes these separate `GH_*` secrets to
`/var/mantrixflow/env/api.secrets.env`, and Docker loads that file directly in
addition to `/var/mantrixflow/env/api.env`.

### ELT Repository Secrets

Repository: `cloud.api.etl.server.mantrixflow.com`

Required secrets:

| Secret | Value |
| --- | --- |
| `HETZNER_API_TOKEN` | Same Hetzner API token |
| `HETZNER_SSH_PRIVATE_KEY` | Same private SSH key used by infra |
| `GHCR_READ_TOKEN` | GitHub classic PAT with `read:packages` |
| `INTERNAL_TOKEN` | Shared internal token, same value in API and ELT |
| `HETZNER_ELT_ENV` | Complete multiline production env file for the Python ELT server |

Optional ELT repo secret:

| Secret | Default | Use |
| --- | --- | --- |
| `HETZNER_SERVER_NAME` | `mantrixflow-production` | Must match infra if overridden |

Use this shape for `HETZNER_ELT_ENV`:

```env
PORT=8000
ENVIRONMENT=production
LOG_LEVEL=INFO

ENCRYPTION_KEY=...
INTERNAL_TOKEN=...
CALLBACK_TOKEN=...
CALLBACK_URL=http://mantrixflow-api:8080/api/v1/internal/elt-callback

MAX_CONCURRENT_RUNS=1
MAX_TAPS_PER_SOURCE=1
DEFAULT_SYNC_TIMEOUT_SECONDS=...

STAGING_ROOT=/var/mantrixflow/staging
STAGING_DISK_LIMIT_GB=50

SUPABASE_URL=...
SUPABASE_SERVICE_ROLE_KEY=...
```

Keep `INTERNAL_TOKEN` inside `HETZNER_ELT_ENV` equal to the GitHub secret
`INTERNAL_TOKEN`, and keep `CALLBACK_TOKEN` equal to the API callback token.

### Secret Reuse Summary

Use the same values across repos for these:

| Secret/value | Repos |
| --- | --- |
| `HETZNER_API_TOKEN` | Infra, API, ELT |
| `HETZNER_SSH_PRIVATE_KEY` | Infra, API, ELT |
| `GHCR_READ_TOKEN` | API, ELT |
| `INTERNAL_TOKEN` | API, ELT, and inside both multiline env files |
| `CALLBACK_TOKEN` | Inside `HETZNER_API_ENV` and `HETZNER_ELT_ENV` |
| `HETZNER_SERVER_NAME` | All repos, only if overriding default |

## 7. Deploy Order

Deploy in this order:

1. Infra repo: push to `main`.
2. ELT repo: push to `mantrixflow`.
3. API repo: push to `mantrixflow`.

Pull requests run validation only. Deployment happens on push to the deployment
branches.

## 8. Capacity Settings For CX33

For CX33, keep settings conservative:

```env
PIPELINE_MAX_CONCURRENT=1
PGMQ_PARALLEL_WORKERS=1
PIPELINE_MAX_PER_ORG_CONCURRENT=1
PIPELINE_MAX_PER_SOURCE_CONCURRENT=1
MAX_CONCURRENT_RUNS=1
MAX_TAPS_PER_SOURCE=1
STAGING_DISK_LIMIT_GB=50
```

This can handle many users because jobs queue, but it should run only one heavy
ELT pipeline at a time.

If you upgrade to CX43, use:

```env
PIPELINE_MAX_CONCURRENT=2
PGMQ_PARALLEL_WORKERS=2
MAX_CONCURRENT_RUNS=2
STAGING_DISK_LIMIT_GB=100
```

Increase concurrency only after load testing.

## 9. Verify Deployment

```bash
curl -fsS https://cloud.api.mantrixflow.com/health
curl -fsS https://cloud.api.mantrixflow.com/api/v1/health
```

On the server:

```bash
ssh -i ~/.ssh/mantrixflow_hetzner root@SERVER_IPV4
docker ps
docker network inspect mantrixflow
systemctl status caddy
systemctl status mantrixflow-tigris-backup.timer
```

Public checks that should fail:

```bash
curl -i http://SERVER_IPV4:8080/health
curl -i http://SERVER_IPV4:8000/health
```

Only ports 80 and 443 should be public.

## 10. Upgrade From CX33 To CX43

When you need more concurrency:

1. Set infra secret `HETZNER_SERVER_TYPE=cx43`.
2. Run the infra workflow on `main`.
3. Increase env concurrency from 1 to 2.
4. Redeploy ELT and API through GitHub Actions.

Do not jump to high concurrency immediately. Watch memory, disk, and failed
pipeline recovery first.
