# PostHog + Better Stack Observability Setup

This document describes how to configure PostHog observability and Better Stack status page automation for MantrixFlow.

## Overview

| Layer | What's integrated |
|---|---|
| Next.js app (Vercel) | PostHog autocapture, session replay, frontend errors, pageviews |
| Go API (ECS) | PostHog server-side error capture via Fiber middleware |
| Python ELT (ECS) | PostHog exception capture, pipeline lifecycle events |
| Status page | Better Stack `status.mantrixflow.com` — monitors **App** (`cloud.mantrixflow.com`) + **API** (`cloud.api.mantrixflow.com/status`). ELT is internal-only on AWS (`http://elt-service:8000`); health appears in API `/status` → `components.elt_server`. |

---

## Phase 1 — PostHog Frontend (Next.js App)

### Required env vars (Vercel dashboard)

```
NEXT_PUBLIC_POSTHOG_KEY=phc_...
NEXT_PUBLIC_POSTHOG_HOST=https://us.i.posthog.com
```

### What's wired

- `PostHogProvider` in `app/layout.tsx` — initialises posthog-js with session replay and exception capture.
- `PostHogPageView` — captures `$pageview` on every Next.js navigation.
- `instrumentation.ts` — `onRequestError` hook captures server-side exceptions from Route Handlers and Server Components to PostHog using `posthog-node`.
- `middleware.ts` — Supabase session refresh (replaces previous `proxy.ts` which lacked a default export).

### Session replay

All inputs are masked by default (`maskAllInputs: true`). To mask additional elements (e.g. sensitive data in text), add `data-ph-mask` to the element.

---

## Phase 2 — PostHog Go API

### Required env vars (AWS SSM)

```
POSTHOG_API_KEY=phc_...
POSTHOG_HOST=https://us.i.posthog.com
```

### What's wired

- `internal/observability/posthog.go` — singleton PostHog client, `CaptureException`, `Capture`, and `ErrorMiddleware` helpers.
- `ErrorMiddleware` attached in `main.go` — captures all HTTP 5xx responses as PostHog `$exception` events with path, method, and user ID context.
- `client.Close()` deferred in `main.go` — flushes queued events on shutdown.

---

## Phase 3 — PostHog Python ELT

### Required env vars (AWS SSM)

```
POSTHOG_API_KEY=phc_...
POSTHOG_HOST=https://us.i.posthog.com
```

### What's wired

- `core/observability.py` — singleton Posthog client with `capture_exception` and `capture_pipeline_event` helpers.
- FastAPI `exception_handler(Exception)` in `api/main.py` — captures unhandled exceptions.
- `runner/paths/duckdb_staged.py` — captures `pipeline_run_started`, `pipeline_run_succeeded`, `pipeline_run_partial`, `pipeline_run_failed` events with `pipeline_id`, `org_id`, `duration_seconds`, `rows_read`, `rows_written`.
- `lifespan` cleanup — calls `posthog.shutdown()` before process exit to flush queued events.

---

## Phase 4 — Enhanced Health Endpoints

### Go API

**`GET /status`** (public, no auth) — returns aggregated component health for Better Stack polling:

```json
{
  "status": "success",
  "data": {
    "status": "operational",
    "components": {
      "api": { "status": "operational", "latency_ms": 0 },
      "database": { "status": "operational", "latency_ms": 8 },
      "queue": { "status": "operational", "detail": "active_runs=1" },
      "elt_server": { "status": "operational", "latency_ms": 42 }
    },
    "version": "2.0.0",
    "timestamp": "2026-05-28T12:00:00Z"
  },
  "message": "OK"
}
```

**`GET /api/v1/health`** — same structure plus admission stats (queue depth, rate limit counters).

### Python ELT

**`GET /health`** — enhanced with disk usage and concurrency info:

```json
{
  "status": "ok",
  "version": "1.0.0",
  "dlt_version": "1.23.0",
  "active_runs": 1,
  "max_concurrent_runs": 4,
  "disk_free_gb": 18.4,
  "disk_total_gb": 50.0,
  "disk_pct_free": 36.8
}
```

Status changes to `"at_capacity"` when `active_runs >= max_concurrent_runs` and `"low_disk"` when `disk_pct_free < 10`.

---

## Phase 5 — Better Stack Status Page Automation

**UI walkthrough (monitors, status page, custom domain):** [betterstack-status-page-creation.md](./betterstack-status-page-creation.md)

Production URLs (not `api.mantrixflow.com`):

| Monitor | URL |
| --- | --- |
| App | `https://cloud.mantrixflow.com` |
| Go API | `https://cloud.api.mantrixflow.com/status` |
| ELT | No public URL on AWS — use API `/status` or a manual Better Stack resource |

### One-time setup (API)

```bash
# 1. Create an Uptime API token
#    Better Stack → Settings → API tokens → Team-based tokens

# 2. Get your status page ID
curl -H "Authorization: Bearer $BETTERSTACK_API_TOKEN" \
  https://uptime.betterstack.com/api/v2/status-pages | jq '.data[].id'

# 3. Get resource IDs for each component (App, API, ELT Pipeline)
curl -H "Authorization: Bearer $BETTERSTACK_API_TOKEN" \
  https://uptime.betterstack.com/api/v2/status-pages/$PAGE_ID/resources \
  | jq '.data[] | {id, public_name: .attributes.public_name}'
```

### Required env vars (AWS SSM — Go API only)

```
BETTERSTACK_API_TOKEN=...
BETTERSTACK_STATUS_PAGE_ID=...
BETTERSTACK_RESOURCE_APP=...       # status_page_resource_id for "App"
BETTERSTACK_RESOURCE_API=...       # status_page_resource_id for "main-server"
BETTERSTACK_RESOURCE_ELT=...       # status_page_resource_id for "etl-server"
```

### Automated status report creation (approach A: PostHog webhook)

1. In PostHog, create an **Action** matching error spike conditions (e.g. `$exception` events for a service).
2. Add a **Webhook destination**:
   - URL: `https://cloud.api.mantrixflow.com/api/v1/internal/incident-webhook`
   - Header: `X-Internal-Token: <INTERNAL_TOKEN>`
3. Tag the action payload with `service:api`, `service:app`, or `service:elt` to route to the correct Better Stack resource.
4. The Go webhook relay at `internal/server/incident_webhook.go` maps the tag to a resource ID and calls the Better Stack API.

On `alert.triggered` the relay creates a status report with `status: "downtime"`. On `alert.resolved` it creates a status report with `status: "resolved"` to close the incident on the status page.

### Automated status report creation (approach B: health polling)

Configure a Better Stack **HTTP monitor** to poll:
- `https://cloud.api.mantrixflow.com/status` every 60 s

Better Stack auto-creates status page reports when the endpoint returns non-200 or a degraded payload. Link the monitor to the relevant status page resource for automatic status page updates.

---

## Phase 6 — Status Page Widget

### App sidebar

`StatusIndicator` in `apps/app/components/shared/StatusIndicator.tsx` is embedded in the workspace sidebar footer. It:
- Polls `https://status.mantrixflow.com/index.json` (Better Stack public JSON endpoint) every 60 s.
- Reads `data.attributes.aggregate_state` — values: `operational`, `degraded`, `downtime`, `maintenance`.
- Shows a green pulsing dot when all systems are operational.
- Shows a yellow/red/blue dot with the status label otherwise.
- Links to `https://status.mantrixflow.com/` on click.

### Website footer

`StatusBadge` in `apps/website/components/status-badge.tsx` replaces the static "System Status" link in the footer. Same polling logic with a compact design matching the existing footer typography.

---

## Production Deployment (AWS ECS + Vercel)

Authoritative bootstrap: [`apps/mantrixflow-infra/DEPLOYMENT.md`](../apps/mantrixflow-infra/DEPLOYMENT.md).  
Better Stack UI steps: [betterstack-status-page-creation.md](./betterstack-status-page-creation.md).

API and ELT run on **ECS Fargate** in `ap-south-1`; secrets live in **SSM** at `/mantrixflow/production/<NAME>`.

### 1 — Vercel (frontend only)

Repo: `dabhivijay2478/InsightFlow-app`, branch `mantrixflow`.

| Variable | Value | Environments |
| --- | --- | --- |
| `NEXT_PUBLIC_POSTHOG_KEY` | `phc_...` | Production, Preview |
| `NEXT_PUBLIC_POSTHOG_HOST` | `https://us.i.posthog.com` | Production, Preview |
| `NEXT_PUBLIC_API_URL` | `https://cloud.api.mantrixflow.com` | Production (set by infra) |

Redeploy Vercel after changes. Better Stack vars are **not** on Vercel — widgets poll `https://status.mantrixflow.com/index.json`.

### 2 — AWS SSM (Go API + ELT)

After CDK lists these names in `apiSecretParameterNames` / `eltSecretParameterNames` (see `mantrixflow-infra/cdk/lib/config.ts`), write values:

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

# PostHog (API + ELT)
put_ssm POSTHOG_API_KEY "phc_your-project-token"
put_ssm POSTHOG_HOST "https://us.i.posthog.com"

# Better Stack (Go API only)
put_ssm BETTERSTACK_API_TOKEN "your-token"
put_ssm BETTERSTACK_STATUS_PAGE_ID "your-page-id"
put_ssm BETTERSTACK_RESOURCE_APP "resource-id-app"
put_ssm BETTERSTACK_RESOURCE_API "resource-id-api"
put_ssm BETTERSTACK_RESOURCE_ELT " "   # optional
```

Or set `posthog_*` / `betterstack_*` in infra Terraform (`terraform/ssm.tf`) and run **Deploy Infrastructure** on `mantrixflow-infra` `main`.

### 3 — Redeploy ECS

1. **Deploy Infrastructure** (if CDK secret list changed) so the task definition references new SSM keys.
2. **Deploy API Production** — `main-server-mantrixflow.com`, branch `mantrixflow`.
3. **Deploy ELT Production** — `etl-server-mantrixflow.com`, branch `mantrixflow` (PostHog only).

### 4 — Smoke checks

```bash
curl -f https://cloud.api.mantrixflow.com/health
curl -sS https://cloud.api.mantrixflow.com/status | jq '.data.components.elt_server'
curl -sS https://status.mantrixflow.com/index.json | jq '.data.attributes.aggregate_state'
```

CloudWatch: `/ecs/mantrixflow/api`, `/ecs/mantrixflow/elt`.

---

### Where to find the PostHog Project Token

1. Log in to [us.posthog.com](https://us.posthog.com).
2. Go to **Settings → Project → Project API key**.
3. Copy the value starting with `phc_`.

---

### Where to find the Better Stack values

See [betterstack-status-page-creation.md](./betterstack-status-page-creation.md) → Step 7 for the `curl` commands that return all IDs.

---

## Troubleshooting

| Symptom | Check |
|---|---|
| PostHog events not appearing | Verify `NEXT_PUBLIC_POSTHOG_KEY` is set in Vercel; check browser network tab for requests to `us.i.posthog.com` |
| Server errors not captured | Check SSM `POSTHOG_API_KEY`; CloudWatch `/ecs/mantrixflow/api` for "PostHog disabled" |
| Better Stack webhook failing | `X-Internal-Token` must match SSM `INTERNAL_TOKEN`; webhook URL `https://cloud.api.mantrixflow.com/api/v1/internal/incident-webhook` |
| Status widget shows unknown | `curl https://status.mantrixflow.com/index.json`; CNAME `status` → Better Stack (grey cloud) |
| No status report created | SSM `BETTERSTACK_*`; redeploy API after CDK secret list update |
| API healthy but ELT red on `/status` | ELT is internal — check `mantrixflow-elt-service` tasks and `http://elt-service:8000/health` from API task network |
