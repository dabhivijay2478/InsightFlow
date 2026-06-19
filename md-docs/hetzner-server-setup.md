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

Create Tigris before running the first infra workflow.

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

### Infra Repo Secrets

Repository: `mantrixflow-infra`

```text
HETZNER_API_TOKEN
HETZNER_SSH_PUBLIC_KEY
HETZNER_SSH_PRIVATE_KEY
CLOUDFLARE_API_TOKEN
CLOUDFLARE_ZONE_ID
TIGRIS_ACCESS_KEY_ID
TIGRIS_SECRET_ACCESS_KEY
TIGRIS_ENDPOINT
TIGRIS_BUCKET
```

Optional infra repo secrets:

```text
HETZNER_SERVER_NAME
HETZNER_SERVER_TYPE
HETZNER_LOCATION
```

### API Repo Secrets

Repository: `cloud.api.mantrixflow.com`

```text
HETZNER_API_TOKEN
HETZNER_SSH_PRIVATE_KEY
GHCR_READ_TOKEN
INTERNAL_TOKEN
HETZNER_API_ENV
```

Optional API repo secret:

```text
HETZNER_SERVER_NAME
```

### ELT Repo Secrets

Repository: `cloud.api.etl.server.mantrixflow.com`

```text
HETZNER_API_TOKEN
HETZNER_SSH_PRIVATE_KEY
GHCR_READ_TOKEN
INTERNAL_TOKEN
HETZNER_ELT_ENV
```

Optional ELT repo secret:

```text
HETZNER_SERVER_NAME
```

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
