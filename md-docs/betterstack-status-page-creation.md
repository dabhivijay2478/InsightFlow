# Better Stack Status Page — AWS Production Setup

Step-by-step guide for **AWS ECS Fargate** production. Authoritative infra guide: [`apps/mantrixflow-infra/DEPLOYMENT.md`](../apps/mantrixflow-infra/DEPLOYMENT.md).

Code wiring: [`posthog-betterstack-setup.md`](./posthog-betterstack-setup.md).

---

## Production architecture (AWS)

```text
Users
  |
  |-- https://cloud.mantrixflow.com     -> Vercel (InsightFlow-app, branch mantrixflow)
  |
  |-- https://cloud.api.mantrixflow.com -> Cloudflare DNS-only -> ALB -> ECS Go API :8080
  |                                      GET /health, GET /status (public)
  |
  |-- https://status.mantrixflow.com    -> Better Stack (CNAME in Cloudflare)
  |
  `-- Python ELT: INTERNAL ONLY
        http://elt-service:8000  (ECS Service Connect, not on the internet)
```

| Item | Value |
| --- | --- |
| AWS region | `ap-south-1` |
| ECS cluster | `mantrixflow-cluster` |
| API service | `mantrixflow-api-service` |
| ELT service | `mantrixflow-elt-service` (no ALB, no public URL) |
| SSM prefix | `/mantrixflow/production/` |
| API public URL | `https://cloud.api.mantrixflow.com` |
| App URL | `https://cloud.mantrixflow.com` |

Repos (from infra `DEPLOYMENT.md`):

| Component | GitHub repo | Branch |
| --- | --- | --- |
| Infra | `dabhivijay2478/mantrixflow-infra` | `main` |
| Frontend | `dabhivijay2478/InsightFlow-app` | `mantrixflow` |
| Go API | `dabhivijay2478/main-server-mantrixflow.com` | `mantrixflow` |
| Python ELT | `dabhivijay2478/etl-server-mantrixflow.com` | `mantrixflow` |

---

## What to monitor on Better Stack (free plan)

On the free tier you only get **URL** monitors (e.g. **URL becomes unavailable**). You do **not** get a public URL for ELT on AWS.

| Better Stack monitor | URL | Why |
| --- | --- | --- |
| **MantrixFlow App** | `https://cloud.mantrixflow.com` | Vercel frontend |
| **MantrixFlow API** | `https://cloud.api.mantrixflow.com/status` | Public; includes `components.elt_server` in JSON |
| ~~ELT~~ | ~~no public host~~ | ELT is `http://elt-service:8000` inside the VPC only |

**Status page rows:** add **App** + **API** (linked to the two monitors). For a third row **ELT Pipeline**, either:

- Rely on API `/status` (recommended — no extra monitor), or
- Add a **Manually tracked** resource in Better Stack (no HTTP check), updated when you know ELT is down.

Do **not** use `api.mantrixflow.com`, `elt.mantrixflow.com`, or `https://cloud.api.etl.server.mantrixflow.com` unless you deliberately add those DNS records in infra.

---

## Prerequisites

1. AWS bootstrap finished per [`DEPLOYMENT.md`](../apps/mantrixflow-infra/DEPLOYMENT.md) Steps 16–18.
2. This passes:

```bash
curl -f https://cloud.api.mantrixflow.com/health
curl -sS https://cloud.api.mantrixflow.com/status | head -c 500
curl -sS -o /dev/null -w "%{http_code}\n" https://cloud.mantrixflow.com/
```

3. Better Stack account (free tier is fine).
4. Cloudflare access for `status.mantrixflow.com` CNAME.

If `curl https://cloud.api.mantrixflow.com/health` fails with SSL handshake errors through Cloudflare, see **Production Smoke Checks** in `DEPLOYMENT.md` (API DNS record must be **DNS only**, grey cloud).

---

## Step 1 — Better Stack API token

1. [betterstack.com](https://betterstack.com) → team **MantrixFlow**.
2. **Settings** → **API tokens** → **Team-based tokens** → **Uptime** → create token.
3. Save as `BETTERSTACK_API_TOKEN` (used in SSM and local `curl`).

---

## Step 2 — Create URL monitors (Better Stack UI)

Create **two** monitors (App + API). On the **free plan**, use **URL becomes unavailable** only; keyword matching and extra alert types require an upgrade.

**Navigation:** **Uptime** → **Monitors** → **Create monitor**.

### Quick reference (both monitors)

| Field | Monitor 1 — App | Monitor 2 — API |
| --- | --- | --- |
| **Alert us when** | URL becomes unavailable | URL becomes unavailable |
| **URL to monitor** | `https://cloud.mantrixflow.com` | `https://cloud.api.mantrixflow.com/status` |
| **Pronounceable monitor name** | `MantrixFlow App` | `MantrixFlow API` |
| **Check frequency** | 3 minutes | 3 minutes |
| **Recovery period** | 3 minutes | 3 minutes |
| **Confirmation period** | Immediate start | Immediate start |
| **SSL/TLS verification** | On | On |
| **HTTP method** | GET | GET |
| **Request timeout** | 30 seconds | 30 seconds |
| **Regions** | Asia (+ optional Europe / North America) | Asia (+ optional Europe / North America) |

Do **not** create a public ELT monitor — ELT is internal on AWS (`http://elt-service:8000`). API `/status` already reports `components.elt_server`.

---

### Monitor 1 — MantrixFlow App (full form)

Open **Create monitor** and fill each section in order.

#### What should we monitor?

| UI label | Value |
| --- | --- |
| **Alert us when** | **URL becomes unavailable** (free plan; ignore “Upgrade” on keyword/other types for now) |
| **URL to monitor** | `https://cloud.mantrixflow.com` |

Leave **import multiple monitors** empty unless you are bulk-importing.

#### How should we escalate incidents?

| UI label | Recommended |
| --- | --- |
| **Notify the primary responder** | On — uses your **primary on-call schedule** (configure under team settings if empty) |
| **Escalation policy** | Optional — add later if you use on-call rotations |
| **How should we notify the primary responder?** | Your choice: **Email**, **Push**, and/or **SMS** (at least email for production) |
| **If the primary responder doesn't acknowledge the incident** | **Immediately alert all other team members** (or tighten per your team policy) |

#### Advanced settings

Expand **Advanced settings** and set:

| UI label | Value | Notes |
| --- | --- | --- |
| **Pronounceable monitor name** | `MantrixFlow App` | Used for phone/voice alerts |
| **Recovery period** | `3 minutes` | Monitor must stay up this long before auto-resolve |
| **Confirmation period** | `Immediate start` | Start incident as soon as a check fails |
| **Check frequency** | `3 minutes` | Typical free-plan minimum |
| **Internet Protocol (IP) version** | **Both IPv4 and IPv6** | Default |
| **SSL/TLS verification** | **On** | Validates HTTPS cert for Vercel |
| **SSL expiration** | Don't check for SSL expiration | Upgrade-only on free |
| **Domain expiration** | Don't check for domain expiration | Upgrade-only on free |
| **HTTP method used to make the request** | **GET** | |
| **Request timeout** | `30 seconds` | Enough for Vercel cold start |
| **Request body** | *(leave empty)* | Only for POST/PUT/PATCH |
| **Request headers** | *(leave empty)* | No `Authorization` — app is public |
| **Basic HTTP authentication** | *(leave empty)* | |
| **Proxy host / port** | *(leave empty)* | Not needed on AWS/Vercel |
| **Maintenance** | Off unless you have a recurring window | |
| **Regions** | **Asia** (required — production near `ap-south-1`) | Optionally add **Europe** or **North America** for extra coverage |
| **Metadata** | Optional | e.g. `Key=service` `Value=app` |

Click **Create monitor** (or **Save**).

---

### Monitor 2 — MantrixFlow API (full form)

**Monitors** → **Create monitor** again. Same sections as Monitor 1; only these fields change:

#### What should we monitor?

| UI label | Value |
| --- | --- |
| **Alert us when** | **URL becomes unavailable** |
| **URL to monitor** | `https://cloud.api.mantrixflow.com/status` |

Why `/status` and not `/health`? `/status` returns JSON with API, database, queue, and **ELT** component health — one monitor covers backend + pipeline worker reachability.

Alternative (simpler): `https://cloud.api.mantrixflow.com/health` — only checks API process up, not DB/ELT detail.

#### Escalation

Same as Monitor 1 (primary responder + your notification channels).

#### Advanced settings

| UI label | Value |
| --- | --- |
| **Pronounceable monitor name** | `MantrixFlow API` |
| **Recovery period** | `3 minutes` |
| **Confirmation period** | `Immediate start` |
| **Check frequency** | `3 minutes` |
| **Internet Protocol (IP) version** | Both IPv4 and IPv6 |
| **SSL/TLS verification** | **On** |
| **SSL expiration** / **Domain expiration** | Don't check (free plan) |
| **HTTP method** | **GET** |
| **Request timeout** | `30 seconds` |
| **Request headers** | *(empty)* — public endpoint, no `Bearer` token |
| **Proxy / Basic auth** | *(empty)* |
| **Regions** | **Asia** (+ optional others) |
| **Metadata** | Optional — e.g. `service` = `api` |

Create the monitor.

#### Free plan: keyword check (optional, paid)

If you upgrade later, you can switch **Alert us when** to **URL contains a keyword** on the same URL with keyword `operational` (inside the JSON body). That catches “HTTP 200 but degraded” cases. On free tier, **URL becomes unavailable** (non-2xx / timeout) is enough.

---

### After both monitors exist

1. **Uptime** → **Monitors** — confirm **MantrixFlow App** and **MantrixFlow API** show **Up** (green).
2. If **Down**, fix AWS/Vercel/Cloudflare first (see Prerequisites smoke `curl` commands).
3. Proceed to **Step 3** to attach both monitors to the public status page.

---

## Step 3 — Create status page (Better Stack UI)

Per [Better Stack getting started](https://betterstack.com/docs/uptime/getting-started-with-status-pages/):

1. **Status pages** → **Create status page**.
2. Company name: `MantrixFlow`, subdomain: `mantrixflow`.
3. **Structure** tab → drag **MantrixFlow App** and **MantrixFlow API** from Available → Selected.
4. Public names: **App**, **API** (or **main-server**).
5. **Save changes**.

Note **status page ID** from the URL for `BETTERSTACK_STATUS_PAGE_ID`.

---

## Step 4 — Custom domain `status.mantrixflow.com`

1. Status page **Settings** → **Custom domain** → `status.mantrixflow.com`.
2. In **Cloudflare** (zone `mantrixflow.com`):

| Type | Name | Target | Proxy |
| --- | --- | --- | --- |
| CNAME | `status` | value from Better Stack | **DNS only** (grey) |

3. Verify:

```bash
curl -sS https://status.mantrixflow.com/index.json | jq '.data.attributes.aggregate_state'
```

App/website widgets poll this JSON (see `StatusIndicator.tsx`, `status-badge.tsx`).

---

## Step 5 — Automatic reports

Status page **Settings** → enable **Automatic reports** so monitor failures open/resolve public incidents without PostHog.

---

## Step 6 — Collect resource IDs (Better Stack API)

```bash
export BETTERSTACK_API_TOKEN="your-token"
export AWS_REGION=ap-south-1

curl -s -H "Authorization: Bearer $BETTERSTACK_API_TOKEN" \
  https://uptime.betterstack.com/api/v2/status-pages \
  | jq '.data[] | {id, company: .attributes.company_name}'

export BETTERSTACK_STATUS_PAGE_ID="your-page-id"

curl -s -H "Authorization: Bearer $BETTERSTACK_API_TOKEN" \
  "https://uptime.betterstack.com/api/v2/status-pages/${BETTERSTACK_STATUS_PAGE_ID}/resources" \
  | jq '.data[] | {id, public_name: .attributes.public_name}'
```

Map IDs:

| SSM parameter (under `/mantrixflow/production/`) | Env var in Go |
| --- | --- |
| `BETTERSTACK_API_TOKEN` | `BETTERSTACK_API_TOKEN` |
| `BETTERSTACK_STATUS_PAGE_ID` | `BETTERSTACK_STATUS_PAGE_ID` |
| `BETTERSTACK_RESOURCE_APP` | `BETTERSTACK_RESOURCE_APP` |
| `BETTERSTACK_RESOURCE_API` | `BETTERSTACK_RESOURCE_API` |
| `BETTERSTACK_RESOURCE_ELT` | Optional; leave empty if no ELT row/monitor |

---

## Step 7 — Store secrets in AWS SSM (production)

ECS reads secrets from **`/mantrixflow/production/<NAME>`** (see `cdk/lib/config.ts` `ssmPrefix`).

### Option A — AWS CLI (quick, after CDK lists the parameter names)

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

# Better Stack (Go API only)
put_ssm BETTERSTACK_API_TOKEN "your-betterstack-token"
put_ssm BETTERSTACK_STATUS_PAGE_ID "your-page-id"
put_ssm BETTERSTACK_RESOURCE_APP "resource-id-app"
put_ssm BETTERSTACK_RESOURCE_API "resource-id-api"
put_ssm BETTERSTACK_RESOURCE_ELT " "   # optional; blank if unused

# PostHog (Go API + ELT)
put_ssm POSTHOG_API_KEY "phc_your-project-key"
put_ssm POSTHOG_HOST "https://us.i.posthog.com"
```

### Option B — Infra Terraform (preferred long-term)

Add values to `mantrixflow-infra` GitHub environment `production-infra` secrets and extend `terraform/ssm.tf` / `cdk/lib/config.ts` (see observability section in `DEPLOYMENT.md`), then run **Deploy Infrastructure** on infra `main`.

### Redeploy API after new SSM keys

New SSM parameters are only injected when the ECS task definition references them. After adding names to CDK:

1. Merge infra change → **Deploy Infrastructure** (or register new task definition).
2. **Deploy API Production** on `main-server-mantrixflow.com` branch `mantrixflow`.
3. **Deploy ELT Production** if you added `POSTHOG_*` to ELT secrets.

---

## Step 8 — Vercel (frontend PostHog only)

Repo: `dabhivijay2478/InsightFlow-app`, branch `mantrixflow`.

Vercel → project → **Settings → Environment Variables** (Production):

| Variable | Value |
| --- | --- |
| `NEXT_PUBLIC_POSTHOG_KEY` | `phc_...` (Project API key) |
| `NEXT_PUBLIC_POSTHOG_HOST` | `https://us.i.posthog.com` |
| `NEXT_PUBLIC_API_URL` | `https://cloud.api.mantrixflow.com` |

Already required by infra (do not change to `api.mantrixflow.com`):

```text
NEXT_PUBLIC_API_URL=https://cloud.api.mantrixflow.com
```

Redeploy Vercel after env changes. No Better Stack vars on Vercel.

---

## Step 9 — PostHog webhook → Go API (optional)

1. PostHog → **Data pipelines** → **Destinations** → **Webhook**.
2. URL: `https://cloud.api.mantrixflow.com/api/v1/internal/incident-webhook`
3. Header: `X-Internal-Token: <same as /mantrixflow/production/INTERNAL_TOKEN in SSM>`
4. Body tag example: `"tags": ["service:api"]` (or `service:app`).

Get `INTERNAL_TOKEN` (do not print in logs):

```bash
aws ssm get-parameter \
  --name /mantrixflow/production/INTERNAL_TOKEN \
  --with-decryption \
  --region ap-south-1 \
  --query Parameter.Value \
  --output text
```

---

## Step 10 — Smoke test (AWS + Better Stack)

```bash
# API (same as DEPLOYMENT.md)
curl -f https://cloud.api.mantrixflow.com/health
curl -sS https://cloud.api.mantrixflow.com/status | jq '.data.status, .data.components.elt_server'

# ECS
aws ecs describe-services \
  --cluster mantrixflow-cluster \
  --services mantrixflow-api-service mantrixflow-elt-service \
  --region ap-south-1 \
  --query 'services[].{name:serviceName,status:status,running:runningCount}'

# Better Stack public page
curl -sS https://status.mantrixflow.com/index.json | jq '.data.attributes.aggregate_state'

# CloudWatch logs
# /ecs/mantrixflow/api
# /ecs/mantrixflow/elt
```

---

## Checklist

- [ ] `curl -f https://cloud.api.mantrixflow.com/health` succeeds
- [ ] Better Stack monitors: App + API (no public ELT URL)
- [ ] Status page Structure: App + API linked to monitors
- [ ] `status.mantrixflow.com` CNAME verified (grey cloud)
- [ ] SSM: `BETTERSTACK_*`, `POSTHOG_*` written under `/mantrixflow/production/`
- [ ] CDK secret list includes those names; API/ELT redeployed
- [ ] Vercel: `NEXT_PUBLIC_POSTHOG_*`, `NEXT_PUBLIC_API_URL=https://cloud.api.mantrixflow.com`
- [ ] PostHog webhook uses `cloud.api.mantrixflow.com` + `X-Internal-Token`

---

## Related docs

- [`apps/mantrixflow-infra/DEPLOYMENT.md`](../apps/mantrixflow-infra/DEPLOYMENT.md) — full AWS bootstrap, GitHub Actions, smoke checks
- [`posthog-betterstack-setup.md`](./posthog-betterstack-setup.md) — application code paths
- [`deployment-vercel.md`](./deployment-vercel.md) — Vercel frontend
- [`aws-ses-setup.md`](./aws-ses-setup.md) — product email on AWS SES
