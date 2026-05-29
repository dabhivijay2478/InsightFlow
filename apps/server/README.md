# MantrixFlow Server (Go API + Python ELT)

Production runs on **AWS ECS Fargate** (`ap-south-1`), not a VPS. Use the infra repo guide and service READMEs for day-to-day work.

## Production (AWS)

| Guide | Purpose |
| --- | --- |
| [`apps/mantrixflow-infra/DEPLOYMENT.md`](../mantrixflow-infra/DEPLOYMENT.md) | ECS, ALB, SSM secrets, GitHub Actions, smoke checks |
| [`md-docs/aws-ses-setup.md`](../../md-docs/aws-ses-setup.md) | Transactional email (Go API `EMAIL_PROVIDER=ses`) |
| [`md-docs/deployment-vercel.md`](../../md-docs/deployment-vercel.md) | Frontend on Vercel |
| [`md-docs/betterstack-status-page-creation.md`](../../md-docs/betterstack-status-page-creation.md) | Status page + monitors |

**Production URLs**

```text
App:  https://cloud.mantrixflow.com          (Vercel)
API:  https://cloud.api.mantrixflow.com      (ALB → ECS Go API :8080)
ELT:  http://elt-service:8000                (internal Service Connect only)
```

Secrets: AWS SSM `/mantrixflow/production/*` — see `.env.production.example` for names.

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

`docker-compose.prod.yml` runs API + ELT on one machine for integration tests. **Not** used for AWS production.

```bash
cd apps/server
cp .env.production.example .env.production   # fill values for a test stack
docker compose -f docker-compose.prod.yml up -d --build
```

## Layout

```text
apps/server/
├── docker-compose.prod.yml      # optional local stack
├── .env.production.example      # env names (production values → SSM on AWS)
├── main-server/                 # Go API (Fiber)
└── etl-server/                  # Python ELT (FastAPI + dlt)
```

## Architecture (AWS)

```text
Internet
   │
   ├── cloud.mantrixflow.com ──────────► Vercel (Next.js)
   │
   └── cloud.api.mantrixflow.com ──────► ALB → ECS Go API :8080
                                              │
                                              └── Service Connect
                                                    elt-service:8000 (ELT)
```

- API is public (TLS at ALB; Cloudflare DNS-only on `cloud.api`).
- ELT has no public hostname; health is exposed via API `GET /status` → `components.elt_server`.
- Product email: **AWS SES** (`md-docs/aws-ses-setup.md`).
