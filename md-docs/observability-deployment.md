# Observability deployment (AWS + Vercel)

> **Archived:** This AWS/ECS procedure is retained for historical reference and
> is not the current production deployment path. Use the
> [Hetzner deployment guide](../apps/mantrixflow-infra/DEPLOYMENT.md) for current
> infrastructure and inject the observability variables through that workflow.

Put **PostHog** and **Better Stack** secrets into production and redeploy.  
No PostHog or Better Stack UI steps here — do those first:

1. [betterstack-setup.md](./betterstack-setup.md) — monitors, status page, collect IDs  
2. [posthog-setup.md](./posthog-setup.md) — PostHog project UI (error tracking, optional webhook)  

Current infrastructure: [`apps/mantrixflow-infra/DEPLOYMENT.md`](../apps/mantrixflow-infra/DEPLOYMENT.md).

---

## Production map

```text
cloud.mantrixflow.com          → Vercel (PostHog browser SDK)
cloud.api.mantrixflow.com      → ECS Go API (PostHog server + webhook)
http://elt-service:8000        → ECS ELT internal (PostHog server only)
status.mantrixflow.com         → Better Stack (public JSON + monitors)
```

| Secret prefix | ` /mantrixflow/production/` |
| AWS region | `ap-south-1` |

---

## Step 1 — Vercel (PostHog frontend only)

Project: `dabhivijay2478/InsightFlow-app`, branch `mantrixflow`.

**Settings → Environment variables → Production:**

| Variable | Value |
| --- | --- |
| `NEXT_PUBLIC_POSTHOG_KEY` | Copy from [posthog-setup.md](./posthog-setup.md) |
| `NEXT_PUBLIC_POSTHOG_HOST` | `https://us.i.posthog.com` |
| `NEXT_PUBLIC_API_URL` | `https://cloud.api.mantrixflow.com` |

No Better Stack variables on Vercel (widgets call `status.mantrixflow.com` directly).

**Redeploy** the Vercel project after saving.

---

## Step 2 — AWS SSM parameters

ECS injects secrets listed in `apps/mantrixflow-infra/cdk/lib/config.ts`.

### PostHog (API + ELT tasks)

| SSM name | Example |
| --- | --- |
| `POSTHOG_API_KEY` | `phc_...` |
| `POSTHOG_HOST` | `https://us.i.posthog.com` |

### Better Stack (API task only)

| SSM name | Source |
| --- | --- |
| `BETTERSTACK_API_TOKEN` | [betterstack-setup.md](./betterstack-setup.md) Step 1 |
| `BETTERSTACK_STATUS_PAGE_ID` | Step 6 curl |
| `BETTERSTACK_RESOURCE_APP` | Step 6 curl (App resource `id`) |
| `BETTERSTACK_RESOURCE_API` | Step 6 curl (API resource `id`) |
| `BETTERSTACK_RESOURCE_ELT` | Optional — blank if unused |

These SSM values are for runtime API behavior. The API deploy workflow also
needs GitHub environment secrets named `BETTERSTACK_API_TOKEN`,
`BETTERSTACK_STATUS_PAGE_ID`, and `BETTERSTACK_RESOURCE_API` so CI can create a
maintenance report before deployment and update the same report after the
health check. CI/CD API calls must use
`https://uptime.betterstack.com/api/v2/`, never `https://betterstack.com`.
The workflow ends maintenance by PATCHing the report's `ends_at` to the current
time. Failed deploys then open a separate manual incident with the API resource
marked `downtime`.

### Webhook auth (already required for ELT)

| SSM name | Used by |
| --- | --- |
| `INTERNAL_TOKEN` | PostHog webhook `X-Internal-Token` or `?internal_token=` |

### Write with AWS CLI

```bash
export AWS_REGION=ap-south-1

put_ssm() {
  aws ssm put-parameter \
    --name "/mantrixflow/production/$1" \
    --value "$2" \
    --type SecureString \
    --overwrite \
    --region "$AWS_REGION"
}

# PostHog
put_ssm POSTHOG_API_KEY "phc_..."  # from posthog-setup.md
put_ssm POSTHOG_HOST "https://us.i.posthog.com"

# Better Stack
put_ssm BETTERSTACK_API_TOKEN "your-uptime-token"
put_ssm BETTERSTACK_STATUS_PAGE_ID "your-page-id"
put_ssm BETTERSTACK_RESOURCE_APP "resource-id-for-app"
put_ssm BETTERSTACK_RESOURCE_API "resource-id-for-api"
put_ssm BETTERSTACK_RESOURCE_ELT " "
```

**Terraform option:** set `posthog_*` and `betterstack_*` in `mantrixflow-infra` GitHub `production-infra` secrets and run **Deploy Infrastructure** on infra `main`.

---

## Step 3 — Redeploy ECS

New SSM keys appear in tasks only after the task definition references them (CDK secret list) and services roll:

| Order | Action |
| --- | --- |
| 1 | **Deploy Infrastructure** (if you changed `cdk/lib/config.ts`) |
| 2 | **Deploy API Production** — repo `main-server-mantrixflow.com`, branch `mantrixflow` |
| 3 | **Deploy ELT Production** — repo `etl-server-mantrixflow.com`, branch `mantrixflow` |

---

## Step 4 — Smoke tests

```bash
# Liveness
curl -f https://cloud.api.mantrixflow.com/health
curl -sS https://cloud.api.mantrixflow.com/status | jq '.data.status, .data.components.elt_server'
curl -sS -o /dev/null -w "%{http_code}\n" https://cloud.mantrixflow.com/

# Better Stack public page
curl -sS https://status.mantrixflow.com/index.json | jq '.data.attributes.aggregate_state'

# ECS
aws ecs describe-services \
  --cluster mantrixflow-cluster \
  --services mantrixflow-api-service mantrixflow-elt-service \
  --region ap-south-1 \
  --query 'services[].{name:serviceName,running:runningCount,status:status}'

# PostHog webhook (optional)
export INTERNAL_TOKEN="$(aws ssm get-parameter --name /mantrixflow/production/INTERNAL_TOKEN --with-decryption --region ap-south-1 --query Parameter.Value --output text)"
curl -sS -X POST "https://cloud.api.mantrixflow.com/api/v1/internal/incident-webhook" \
  -H "Content-Type: application/json" \
  -H "X-Internal-Token: $INTERNAL_TOKEN" \
  -d '{"event":"alert.triggered","data":{"name":"Smoke test","tags":["service:main-server"]}}'
```

Logs: `/ecs/mantrixflow/api`, `/ecs/mantrixflow/elt`.

---

## Optional — PostHog + Better Stack together

| Path | What updates the status page |
| --- | --- |
| Better Stack monitors | URL down/up (automatic reports) — **primary** |
| PostHog webhook | New `$exception` / issues → Go API → Better Stack API |

Webhook needs **both** `BETTERSTACK_*` in SSM and PostHog configured per [posthog-setup.md](./posthog-setup.md) Step 4.

**Service → status page row** (webhook routing):

| `service` tag | Better Stack resource |
| --- | --- |
| `app`, `app-server`, `frontend` | `BETTERSTACK_RESOURCE_APP` |
| `main-server`, `api` | `BETTERSTACK_RESOURCE_API` |
| `elt-server`, `elt` | `BETTERSTACK_RESOURCE_ELT` |

---

## Master checklist

- [ ] [betterstack-setup.md](./betterstack-setup.md) — 2 monitors Up, status page live, IDs copied  
- [ ] [posthog-setup.md](./posthog-setup.md) — `phc_` key, issues visible in PostHog  
- [ ] Vercel `NEXT_PUBLIC_POSTHOG_*` set and redeployed  
- [ ] SSM `POSTHOG_*`, `BETTERSTACK_*` written  
- [ ] API + ELT ECS redeployed  
- [ ] `curl` smoke tests pass  
- [ ] (Optional) PostHog webhook test returns success  

---

## Troubleshooting (deployment)

| Symptom | Check |
| --- | --- |
| PostHog still empty in prod | SSM `POSTHOG_API_KEY` + API/ELT redeploy |
| Webhook 401 | `INTERNAL_TOKEN` in header or URL query |
| Webhook skipped | `BETTERSTACK_*` missing or wrong resource IDs |
| Status widget broken | `status.mantrixflow.com` CNAME (grey cloud) |
| API monitor Down | `cloud.api` DNS-only; ECS task health |

---

## Related

- [betterstack-setup.md](./betterstack-setup.md)
- [posthog-setup.md](./posthog-setup.md)
- [deployment-vercel.md](./deployment-vercel.md)
- [apps/mantrixflow-infra/DEPLOYMENT.md](../apps/mantrixflow-infra/DEPLOYMENT.md)
# Historical AWS/ECS observability deployment

> **Historical only.** The current backend deployment is Hetzner/Terraform; use
> [`../apps/mantrixflow-infra/DEPLOYMENT.md`](../apps/mantrixflow-infra/DEPLOYMENT.md).
> The AWS SSM/ECS commands below are retained for migration history and must not
> be used as current production guidance.
