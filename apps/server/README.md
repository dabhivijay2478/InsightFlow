# MantrixFlow Server (Go API + Python ELT)

The current backend deployment is a Terraform-provisioned **Hetzner** host with
Go and ELT on a private Docker network. Use the infrastructure guide and
service READMEs for day-to-day work.

## Current production guidance (Hetzner)

| Guide | Purpose |
| --- | --- |
| [`apps/mantrixflow-infra/DEPLOYMENT.md`](../mantrixflow-infra/DEPLOYMENT.md) | Terraform, Hetzner, Caddy, private Docker network, GitHub Actions, smoke checks |
| [`md-docs/aws-ses-setup.md`](../../md-docs/aws-ses-setup.md) | Transactional email (Go API `EMAIL_PROVIDER=ses`) |
| [`md-docs/deployment-vercel.md`](../../md-docs/deployment-vercel.md) | Frontend on Vercel |
| Observability | `betterstack-setup.md` → `posthog-setup.md` → `observability-deployment.md` |
| [`md-docs/betterstack-setup.md`](../../md-docs/betterstack-setup.md) | Better Stack UI |
| [`md-docs/posthog-setup.md`](../../md-docs/posthog-setup.md) | PostHog UI |
| [`md-docs/observability-deployment.md`](../../md-docs/observability-deployment.md) | Observability configuration and Vercel deploy |

**Production URLs**

```text
App:  https://cloud.mantrixflow.com          (Vercel)
API:  https://cloud.api.mantrixflow.com      (Caddy → Go API :8080)
ELT:  http://mantrixflow-elt:8000            (private Docker network only)
```

Secrets are injected from the protected `production-hetzner` GitHub
environments; see `.env.production.example` for names.

Deploy flow: push `mantrixflow` branch → **Deploy API Production** / **Deploy ELT Production** workflows (repos `main-server-mantrixflow.com`, `etl-server-mantrixflow.com`).

## Local development

Run services separately (preferred for ELT work):

```bash
# Go API
cd apps/server/main-server && go run ./cmd/server

# Python ELT
cd apps/server/elt-server
./.venv/bin/python -m uvicorn api.main:app --host 0.0.0.0 --port 8000 --loop asyncio
```

See `main-server/README.md` and `elt-server/README.md` for env files.

## Optional: Docker Compose (local / staging only)

`docker-compose.prod.yml` runs API + ELT on one machine and mirrors the current
single-host topology for local or staging validation.

```bash
cd apps/server
cp .env.production.example .env.production   # fill values for a test stack
docker compose -f docker-compose.prod.yml up -d --build
```

## Layout

```text
apps/server/
├── docker-compose.prod.yml      # optional local stack
├── .env.production.example      # deployment environment names
├── main-server/                 # Go API (Fiber)
└── etl-server/                  # Python ELT (FastAPI + dlt)
```

## Architecture (current)

```text
Internet
   │
   ├── cloud.mantrixflow.com ──────────► Vercel (Next.js)
   │
   └── cloud.api.mantrixflow.com ──────► Caddy → Go API :8080
                                                │
                                                └── private Docker network
                                                      mantrixflow-elt:8000
```

- API is public (TLS at Caddy; Cloudflare DNS-only on `cloud.api`).
- ELT has no public hostname; health is exposed via API `GET /status` → `components.elt_server`.
- Product email may use the configured provider; the AWS SES document is a
  provider-specific historical/setup reference, not the backend deployment.

## Historical deployment material

AWS/ECS, Oracle, Contabo, and Dokploy files are retained for history only. They
are not current production instructions. Do not delete them or use them for a
new deployment without an explicit architecture decision.
