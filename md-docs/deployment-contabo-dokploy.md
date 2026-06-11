# MantrixFlow Contabo + Dokploy Cloud CI/CD Setup Guide

This guide deploys the Go API and Python ELT server to an existing Contabo VPS
through Dokploy Cloud. GitHub Actions owns validation, image publishing,
infrastructure reconciliation, deployment, and post-deployment health checks.

After the one-time account and secret bootstrap, deployments happen only by
merging or pushing to `mantrixflow-contabo`. Do not deploy from the Dokploy
dashboard.

## 1. Architecture

```text
GitHub repositories
  |
  |-- mantrixflow-infra
  |     `-- GitHub Actions -> SSH + Cloudflare API + Dokploy Cloud API
  |
  |-- cloud.api.mantrixflow.com
  |     `-- GitHub Actions -> GHCR -> Dokploy Cloud -> go-api
  |
  `-- cloud.api.etl.server.mantrixflow.com
        `-- GitHub Actions -> GHCR -> Dokploy Cloud -> elt-server

Dokploy Cloud control plane
  `-- remote deploy server: existing Contabo VPS
        |-- Traefik: HTTPS -> go-api:8080
        |-- go-api -> http://elt-server:8000
        `-- elt-server -> http://go-api:8080/api/v1/internal/elt-callback
```

The VPS does not host the Dokploy control plane. Do not install self-hosted
Dokploy, expose port `3000`, create `deploy.mantrixflow.com`, or use Terraform
for this Contabo target.

## 2. Repository And Branch Ownership

| Repository | Contabo workflow | Responsibility |
| --- | --- | --- |
| `dabhivijay2478/mantrixflow-infra` | `.github/workflows/deploy-contabo.yml` | VPS hardening, Cloudflare DNS, Dokploy Cloud server and application reconciliation |
| `dabhivijay2478/cloud.api.mantrixflow.com` | `.github/workflows/ci-contabo.yml` | API tests, GHCR publish, API-only deployment and health verification |
| `dabhivijay2478/cloud.api.etl.server.mantrixflow.com` | `.github/workflows/ci-contabo.yml` | ELT tests, GHCR publish, ELT-only deployment and health verification |

All Contabo workflows use only `mantrixflow-contabo`.

| Event | Result |
| --- | --- |
| Pull request targeting `mantrixflow-contabo` | Contabo validation only; no deployment |
| Push or merge to `mantrixflow-contabo` | Validation followed by the repository's Contabo deployment |
| Push to `main`, `master`, or `mantrixflow` | Existing non-Contabo workflows only |
| Dokploy native GitHub auto-deploy | Disabled |

Existing AWS workflows and infrastructure remain separate and unchanged.

## 3. Prerequisites

Create or collect:

- Existing Contabo VPS running Ubuntu 22.04 or newer.
- Non-root SSH user with passwordless `sudo`.
- SSH private/public key pair authorized for that user.
- Dokploy Cloud account, organization ID, account URL, and API key.
- Cloudflare zone containing `mantrixflow.com`.
- Cloudflare API token allowed to edit DNS and zone SSL settings.
- GitHub account/package token that can read private GHCR packages.
- Production API and ELT environment values.

The VPS must initially accept SSH on port `22`. Ports `80` and `443` must be
available for Dokploy/Traefik. The automation closes all other inbound ports.

Verify initial SSH access before running CI/CD:

```bash
ssh -i ~/.ssh/mantrixflow-contabo ubuntu@CONTABO_IP \
  'sudo -n true && echo "SSH and sudo are ready"'
```

Do not put private keys or runtime secrets into repository files.

## 4. Create Access Credentials

### 4.1 Contabo SSH Key

Generate a dedicated key if one does not already exist:

```bash
ssh-keygen -t ed25519 -C mantrixflow-contabo-deploy \
  -N "" -f ~/.ssh/mantrixflow-contabo
ssh-copy-id -i ~/.ssh/mantrixflow-contabo.pub ubuntu@CONTABO_IP
```

This dedicated automation key must be non-interactive because the workflow
does not supply an SSH key passphrase. Store it only in the protected GitHub
environment secret.

GitHub secret values:

```text
CONTABO_SSH_KEY        = complete private key file
CONTABO_SSH_PUBLIC_KEY = complete public key line
CONTABO_SSH_HOST       = Contabo public IPv4 address
CONTABO_SSH_USER       = ubuntu
```

### 4.2 Dokploy Cloud API

In Dokploy Cloud, create an API key that can manage SSH keys, remote servers,
projects, environments, applications, domains, mounts, and deployments.

Record:

```text
DOKPLOY_API_URL         = Dokploy Cloud account URL, without trailing /api
DOKPLOY_API_KEY         = Dokploy Cloud API key
DOKPLOY_ORGANIZATION_ID = Dokploy Cloud organization ID
```

Do not manually create `go-api` or `elt-server`. The infra workflow creates and
reconciles them on the correct Contabo remote server.

### 4.3 GHCR Token

Create a GitHub personal access token for the account that owns the packages.
It needs `read:packages` so Dokploy can pull private images. Repository
workflows publish images using their built-in `GITHUB_TOKEN`.

Record:

```text
GHCR_USERNAME = dabhivijay2478
GHCR_TOKEN    = token with read:packages
```

### 4.4 Cloudflare Token

Create a Cloudflare API token scoped to the `mantrixflow.com` zone with:

```text
Zone / DNS / Edit
Zone / Zone Settings / Edit
Zone / Zone / Read
```

Record the token and zone ID:

```text
CLOUDFLARE_API_TOKEN
CLOUDFLARE_ZONE_ID
```

The workflow manages only `cloud.api.mantrixflow.com` and Full strict SSL. It
does not create `deploy.mantrixflow.com`.

## 5. Prepare Runtime Environment Bundles

The infra workflow stores complete multiline runtime configurations in GitHub
environment secrets and sends them to Dokploy Cloud.

Generate strong shared tokens:

```bash
openssl rand -hex 32
openssl rand -hex 32
openssl rand -hex 32
```

Use matching values for `ELT_INTERNAL_TOKEN` and `CALLBACK_TOKEN` in both
bundles. Use the API's expected encryption key names; the API and ELT
encryption material must be compatible with the application implementation.

### 5.1 `CONTABO_API_ENV`

Create one multiline secret containing the complete API production
environment:

```dotenv
PORT=8080
ENVIRONMENT=production
LOG_LEVEL=info
DATABASE_URL=
DATABASE_DIRECT_URL=
SUPABASE_URL=
SUPABASE_ANON_KEY=
SUPABASE_SERVICE_ROLE_KEY=
ENCRYPTION_MASTER_KEY=
ELT_PYTHON_SERVICE_URL=http://elt-server:8000
ELT_INTERNAL_TOKEN=
CALLBACK_TOKEN=
INTERNAL_TOKEN=
API_PUBLIC_URL=https://cloud.api.mantrixflow.com
APP_WEB_URL=https://app.mantrixflow.com
CORS_ALLOWED_ORIGINS=https://app.mantrixflow.com
```

Add enabled Dodo Payments, Slack, GitHub App, PostHog, email, and other API
runtime variables to the same secret.

### 5.2 `CONTABO_ELT_ENV`

Create one multiline secret containing:

```dotenv
PORT=8000
ENVIRONMENT=production
LOG_LEVEL=INFO
ENCRYPTION_KEY=
CALLBACK_URL=http://go-api:8080/api/v1/internal/elt-callback
ELT_INTERNAL_TOKEN=
CALLBACK_TOKEN=
MAX_CONCURRENT_RUNS=2
STAGING_ROOT=/var/mantrixflow/staging
STAGING_DISK_LIMIT_GB=20
```

Keep `MAX_CONCURRENT_RUNS=2` until memory and workload testing supports a
higher value.

## 6. Configure GitHub Environments

In each repository, open:

```text
Settings -> Environments -> New environment -> production-contabo
```

For fully automated deployment, do not configure required reviewers on this
environment. Branch protection should control merges instead. Configure the
environment deployment branch rule to allow only `mantrixflow-contabo`.

### 6.1 Infra Repository Secrets

Add these environment secrets to `dabhivijay2478/mantrixflow-infra`:

```text
CONTABO_SSH_HOST
CONTABO_SSH_USER
CONTABO_SSH_KEY
CONTABO_SSH_PUBLIC_KEY
CLOUDFLARE_API_TOKEN
CLOUDFLARE_ZONE_ID
DOKPLOY_API_URL
DOKPLOY_API_KEY
DOKPLOY_ORGANIZATION_ID
GHCR_USERNAME
GHCR_TOKEN
CONTABO_API_ENV
CONTABO_ELT_ENV
```

### 6.2 API Repository Secrets

Add these environment secrets to
`dabhivijay2478/cloud.api.mantrixflow.com`:

```text
DOKPLOY_API_URL
DOKPLOY_API_KEY
GHCR_USERNAME
GHCR_TOKEN
```

### 6.3 ELT Repository Secrets

Add the same four environment secrets to
`dabhivijay2478/cloud.api.etl.server.mantrixflow.com`:

```text
DOKPLOY_API_URL
DOKPLOY_API_KEY
GHCR_USERNAME
GHCR_TOKEN
```

## 7. Configure Branch Protection

In all three repositories, protect `mantrixflow-contabo`:

```text
Settings -> Branches or Rulesets -> New branch rule
```

Recommended settings:

- Require pull requests before merging.
- Require the Contabo validation check to pass.
- Require branches to be up to date before merging.
- Block force pushes and branch deletion.
- Optionally require one reviewer.

Do not add environment required reviewers when deployment must remain fully
automatic after merge.

## 8. What The Infra Workflow Creates

When infra `mantrixflow-contabo` is merged, GitHub Actions:

1. Connects to the existing VPS over SSH.
2. Updates Ubuntu and installs `curl`, `fail2ban`, `git`, `jq`, `ufw`,
   `unzip`, and `wget`.
3. Configures UFW to allow only `22`, `80`, and `443`.
4. Creates `/var/mantrixflow/staging` owned by UID/GID `1001`, mode `0750`.
5. Creates/reuses Dokploy Cloud SSH key `mantrixflow-contabo-deploy`.
6. Creates/reuses remote deploy server `mantrixflow-contabo`.
7. Calls Dokploy Cloud `server.setup` when needed and polls `server.validate`.
8. Creates overlay network `mantrixflow-internal`.
9. Creates/reuses project `MantrixFlow Contabo` and environment `production`.
10. Creates/reconciles `go-api` and `elt-server` on the Contabo `serverId`.
11. Configures GHCR credentials, runtime environments, aliases, limits,
    rollback settings, API domain, and ELT staging mount.
12. Removes any accidental public ELT domains.
13. Points `cloud.api.mantrixflow.com` to the Contabo IP.

Managed application settings:

| Application | Exposure | Alias | Limits |
| --- | --- | --- | --- |
| `go-api` | HTTPS `cloud.api.mantrixflow.com`, target `8080` | `go-api` | 1 CPU / 512 MB |
| `elt-server` | Private network only; no domain or published port | `elt-server` | 2 CPU / 4 GB |

## 9. First Deployment

Before the first Contabo deployment, merge the AWS low-capacity fallback CDK
profile into infra `main` and confirm both AWS ECS services remain healthy.
This reduces AWS runtime capacity while preserving it as a rollback target.

The first deployment changes `cloud.api.mantrixflow.com` DNS during the infra
workflow. Schedule the first cutover when a short API interruption is
acceptable, or temporarily restore the previous DNS target until API
deployment completes.

Use pull requests and merge in this order:

1. Merge infra `mantrixflow-contabo`.
2. Confirm the infra GitHub Actions run succeeds.
3. Merge ELT `mantrixflow-contabo`.
4. Confirm the ELT image is published and deployment succeeds.
5. Merge API `mantrixflow-contabo`.
6. Confirm the API image is published and health verification succeeds.

Do not click deploy in Dokploy Cloud. The GitHub Actions workflows call the
Dokploy API after successful CI.

Expected GHCR images:

```text
ghcr.io/dabhivijay2478/cloud.api.mantrixflow.com:<commit-sha>
ghcr.io/dabhivijay2478/cloud.api.mantrixflow.com:contabo-latest
ghcr.io/dabhivijay2478/cloud.api.etl.server.mantrixflow.com:<commit-sha>
ghcr.io/dabhivijay2478/cloud.api.etl.server.mantrixflow.com:contabo-latest
```

## 10. Verify Production

Verify the public API:

```bash
curl --fail --show-error https://cloud.api.mantrixflow.com/health
curl --fail --show-error https://cloud.api.mantrixflow.com/api/v1/health | jq
```

The detailed health response must report database and `elt_server` as
operational.

Verify ELT is not publicly reachable:

```bash
curl --connect-timeout 5 http://CONTABO_IP:8000/health
```

That command must fail.

Verify server state:

```bash
ssh ubuntu@CONTABO_IP '
  sudo ufw status &&
  stat -c "%u:%g %a %n" /var/mantrixflow/staging &&
  sudo docker network inspect mantrixflow-internal >/dev/null &&
  echo "private network exists"
'
```

Expected staging ownership:

```text
1001:1001 750 /var/mantrixflow/staging
```

Run a non-destructive ELT job to verify API-to-ELT dispatch and ELT-to-API
callback behavior.

## 11. Normal CI/CD Releases

For any repository:

1. Create a feature branch from `mantrixflow-contabo`.
2. Open a pull request targeting `mantrixflow-contabo`.
3. Wait for the Contabo CI check.
4. Review and merge.
5. GitHub Actions deploys only that repository's ownership area.

API and ELT releases are independent. Keep cross-service contract changes
backward compatible or release them in multiple steps.

| Repository merged | Automatic result |
| --- | --- |
| Infra | Reconcile VPS, DNS, Dokploy server, project, apps, network, domain, and mount |
| API | Test, publish immutable API image, deploy only `go-api`, verify `/health` |
| ELT | Test, publish immutable ELT image, deploy only `elt-server`, verify detailed health |

## 12. Rollback

### Service rollback

Revert the bad service commit and merge the revert into
`mantrixflow-contabo`:

```bash
git checkout mantrixflow-contabo
git pull --ff-only
git revert BAD_COMMIT_SHA
git push origin mantrixflow-contabo
```

The service workflow builds and deploys the reverted code automatically.
Dokploy rollback settings also protect failed rolling updates.

### Infrastructure rollback

Revert the infra commit and merge/push it to `mantrixflow-contabo`. The infra
workflow reconciles the declared state.

For emergency cutover rollback, restore the previous
`cloud.api.mantrixflow.com` DNS target in Cloudflare, then correct the infra
configuration before the next infra merge.

## 13. Troubleshooting

### Infra workflow cannot SSH

- Confirm `CONTABO_SSH_HOST`, `CONTABO_SSH_USER`, and private key.
- Confirm the matching public key exists in the user's `authorized_keys`.
- Confirm Contabo firewall/security controls permit TCP `22`.
- Confirm the SSH user has passwordless `sudo`.

### Dokploy server validation fails

- Check Dokploy Cloud API URL/key/organization ID.
- Confirm ports `80` and `443` are free.
- Confirm the VPS has enough disk and memory.
- Inspect the `Reconcile Dokploy Cloud remote server` workflow step.

### GHCR image pull fails

- Confirm `GHCR_USERNAME` owns or can access the package.
- Confirm `GHCR_TOKEN` includes `read:packages`.
- Confirm private package access includes the required repositories.

### Service workflow refuses deployment

The workflow refuses to deploy when the application is not assigned to
Dokploy server `mantrixflow-contabo`. Run the infra workflow first. Do not
manually recreate applications on Dokploy Cloud's default server.

### API works but ELT is not operational

- Confirm both applications use `mantrixflow-internal`.
- Confirm aliases are exactly `go-api` and `elt-server`.
- Confirm shared tokens match.
- Confirm API uses `ELT_PYTHON_SERVICE_URL=http://elt-server:8000`.
- Confirm ELT uses
  `CALLBACK_URL=http://go-api:8080/api/v1/internal/elt-callback`.
- Confirm `/var/mantrixflow/staging` is owned by `1001:1001`.

### HTTPS or DNS fails

- Confirm `cloud.api.mantrixflow.com` points to the Contabo IP.
- Confirm inbound ports `80` and `443` are allowed.
- Confirm Dokploy has the API domain with Let's Encrypt enabled.
- Confirm Cloudflare SSL mode is Full strict.

## 14. Security And Operations Checklist

- Rotate Dokploy, GHCR, Cloudflare, and SSH credentials periodically.
- Keep secrets only in GitHub `production-contabo` environments.
- Keep Dokploy native auto-deploy disabled.
- Never publish ELT port `8000`.
- Never add a public ELT domain.
- Keep UFW limited to `22`, `80`, and `443`.
- Review GitHub Actions logs after every deployment.
- Monitor disk usage under `/var/mantrixflow/staging`.
- Back up required production data independently of application containers.
- Keep AWS deployment branches and workflows separate.
