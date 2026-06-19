# Hetzner Server Setup Guide

This guide explains how to create the MantrixFlow production server in Hetzner
Cloud and connect it to the existing GitHub Actions CI/CD deployment.

Use this when creating the server from the Hetzner Cloud UI. The app deployment
still happens through GitHub Actions. Do not manually copy application code to
the server.

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

Hetzner is different from AWS for this setup:

1. First you buy/create the Hetzner server in Hetzner Cloud.
2. Then you copy the server ID and IPv4 into GitHub environment secrets.
3. Then GitHub Actions bootstraps and deploys onto that existing server.

Terraform does **not** create or buy the server now. It only manages DNS/state
and validates the existing server. Keep the exact server name:

```text
mantrixflow-production
```

The API and ELT deploy workflows find the server by this name.

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

## 3. Create Server In Hetzner UI

Go to **Servers > Create Server**.

### Type

Choose:

```text
Shared Resources
```

Do not choose dedicated resources for MVP cost control.

### Location

Choose:

```text
Helsinki
```

Keep future volumes, firewalls, and networks in the same location or network
zone.

### Image

Choose:

```text
OS Images > Ubuntu > Ubuntu 24.04 LTS
```

Do not choose prebuilt Docker apps. The repo bootstrap scripts install Docker
and Caddy consistently.

### Server Type

Choose:

```text
CX33
```

This gives:

```text
4 vCPU
8 GB RAM
80 GB SSD
20 TB traffic
Intel/AMD x86_64
```

For better production headroom choose CX43 instead.

### Networking

Keep public networking enabled:

```text
IPv4: enabled
IPv6: enabled
Private network: not required for single-server setup
```

Hetzner charges for IPv4. That is expected because Cloudflare DNS needs a public
origin IP unless you use a more complex IPv6-only setup.

### SSH Keys

Select:

```text
mantrixflow-hetzner
```

Do not use password login for production.

### Volumes

Do not create a separate volume for the MVP.

The CX33 internal 80 GB disk is enough to start. ELT staging is stored at:

```text
/var/mantrixflow/staging
```

Tigris is used for durable staging backups, not active DuckDB storage.

### Firewalls

Create or select a firewall with only these inbound rules:

| Port | Source | Purpose |
| ---: | --- | --- |
| 22 | your IP or GitHub deploy access | SSH |
| 80 | 0.0.0.0/0, ::/0 | HTTP for Caddy certificate flow |
| 443 | 0.0.0.0/0, ::/0 | HTTPS API |

Do not expose:

```text
8080
8000
3000
```

API and ELT communicate over Docker private networking.

### Backups

Disable Hetzner backups for MVP cost control:

```text
Backups: off
```

Backups add 20% to the server price. Use Tigris for staging backup and keep
database backups at the database provider.

### Placement Groups

Skip placement groups for a single-server MVP.

### Labels

Add labels:

```text
app=mantrixflow
environment=production
managed-by=github-actions
```

### Cloud Config

Leave cloud config empty if Terraform/GitHub Actions will bootstrap the server.

If you are creating manually and want a minimal first boot setup, paste:

```yaml
#cloud-config
package_update: true
package_upgrade: true
packages:
  - curl
  - git
  - ufw
  - fail2ban
  - ca-certificates
  - gnupg
runcmd:
  - ufw allow 22/tcp
  - ufw allow 80/tcp
  - ufw allow 443/tcp
  - ufw --force enable
  - systemctl enable --now fail2ban
  - install -d -o root -g root -m 0755 /var/mantrixflow
  - install -d -o 1001 -g 1001 -m 0750 /var/mantrixflow/staging
```

The full repo bootstrap still runs later from GitHub Actions.

### Name

Set the server name exactly:

```text
mantrixflow-production
```

## 4. After Server Creation

Open the server details page and copy:

```text
HETZNER_SERVER_ID      numeric Hetzner server ID
HETZNER_SERVER_IPV4    public IPv4 address
```

You will add both values to GitHub environment secrets.

Test SSH:

```bash
ssh -i ~/.ssh/mantrixflow_hetzner root@SERVER_IPV4
```

Then check:

```bash
ls -ld /var/mantrixflow/staging
```

Expected ownership:

```text
1001 1001
```

## 5. Tigris Setup

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

## 6. Cloudflare DNS

Create or update:

```text
Type: A
Name: cloud.api
Value: SERVER_IPV4
Proxy: DNS only for first validation
TTL: Auto
```

After Caddy has issued TLS successfully, you can enable the orange-cloud proxy
if desired.

Do not point `cloud.mantrixflow.com` here. The frontend remains separate.

## 7. GitHub Environments

Each repository should have only one environment:

```text
production-hetzner
```

### Infra Repo Secrets

Repository: `mantrixflow-infra`

```text
HETZNER_API_TOKEN
HETZNER_SERVER_ID
HETZNER_SERVER_IPV4
HETZNER_SSH_PUBLIC_KEY
HETZNER_SSH_PRIVATE_KEY
CLOUDFLARE_API_TOKEN
CLOUDFLARE_ZONE_ID
TIGRIS_ACCESS_KEY_ID
TIGRIS_SECRET_ACCESS_KEY
TIGRIS_ENDPOINT
TIGRIS_BUCKET
```

### API Repo Secrets

Repository: `cloud.api.mantrixflow.com`

```text
HETZNER_API_TOKEN
HETZNER_SERVER_ID
HETZNER_SERVER_IPV4
HETZNER_SSH_PRIVATE_KEY
GHCR_READ_TOKEN
INTERNAL_TOKEN
HETZNER_API_ENV
```

### ELT Repo Secrets

Repository: `cloud.api.etl.server.mantrixflow.com`

```text
HETZNER_API_TOKEN
HETZNER_SERVER_ID
HETZNER_SERVER_IPV4
HETZNER_SSH_PRIVATE_KEY
GHCR_READ_TOKEN
INTERNAL_TOKEN
HETZNER_ELT_ENV
```

## 8. Deploy Order

Deploy in this order:

1. Infra repo: push to `main`.
2. ELT repo: push to `mantrixflow`.
3. API repo: push to `mantrixflow`.

Pull requests run validation only. Deployment happens on push to the deployment
branches.

## 9. Capacity Settings For CX33

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

## 10. Verify Deployment

```bash
curl -fsS https://cloud.api.mantrixflow.com/health
curl -fsS https://cloud.api.mantrixflow.com/api/v1/health
```

On the server:

```bash
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

## 11. Upgrade From CX33 To CX43

When you need more concurrency:

1. In Hetzner, power off the server.
2. Resize from CX33 to CX43.
3. Power on the server.
4. Increase env concurrency from 1 to 2.
5. Redeploy ELT and API through GitHub Actions.

Do not jump to high concurrency immediately. Watch memory, disk, and failed
pipeline recovery first.
