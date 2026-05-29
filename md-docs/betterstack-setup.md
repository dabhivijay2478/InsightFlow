# Better Stack setup

Configure **monitors** and the public status page at `https://status.mantrixflow.com`.  
This file is **Better Stack UI only** — no AWS or Vercel steps.

**After this guide:** put secrets in AWS and redeploy → [observability-deployment.md](./observability-deployment.md).

---

## Production URLs (what you will monitor)

| Monitor | URL |
| --- | --- |
| App | `https://cloud.mantrixflow.com` |
| API | `https://cloud.api.mantrixflow.com/status` |

Do **not** add a public ELT URL. On AWS, ELT is internal (`http://elt-service:8000`). API `/status` includes `components.elt_server`.

---

## Before you start

Production API and app should already respond:

```bash
curl -f https://cloud.api.mantrixflow.com/health
curl -sS -o /dev/null -w "%{http_code}\n" https://cloud.mantrixflow.com/
```

If API SSL fails, Cloudflare `cloud.api` must be **DNS only** (grey cloud). See [`apps/mantrixflow-infra/DEPLOYMENT.md`](../apps/mantrixflow-infra/DEPLOYMENT.md).

---

## Step 1 — API token (Better Stack UI)

1. [betterstack.com](https://betterstack.com) → your team.
2. **Settings** → **API tokens** → **Team-based tokens** → **Uptime** → create token.
3. Save the token — you will store it as `BETTERSTACK_API_TOKEN` in [observability-deployment.md](./observability-deployment.md).

---

## Step 2 — Create monitors (Better Stack UI)

**Uptime** → **Monitors** → **Create monitor**. Create **two** monitors.

Free plan: use **URL becomes unavailable** only.

### Quick reference

| Field | MantrixFlow App | MantrixFlow API |
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
| **Regions** | Asia | Asia |
| **Request headers** | *(empty)* | *(empty)* |

### Monitor 1 — App

Fill the **Create monitor** form:

- **URL to monitor:** `https://cloud.mantrixflow.com`
- **Pronounceable monitor name:** `MantrixFlow App`
- **Escalation:** notify primary responder (email at minimum)
- **Advanced:** values in the table above

### Monitor 2 — API

Create again with:

- **URL:** `https://cloud.api.mantrixflow.com/status`
- **Name:** `MantrixFlow API`
- Same timing/SSL/region settings as App

Confirm both show **Up** under **Monitors**.

---

## Step 3 — Status page (Better Stack UI)

1. **Status pages** → **Create status page**.
2. Company: `MantrixFlow`, subdomain: `mantrixflow`.
3. **Structure** → move **MantrixFlow App** and **MantrixFlow API** to **Selected**.
4. Public labels: **App**, **API**.
5. **Save changes**.

Copy the **status page ID** from the browser URL (for SSM later).

---

## Step 4 — Custom domain (Better Stack + Cloudflare)

1. Status page **Settings** → **Custom domain** → `status.mantrixflow.com`.
2. Cloudflare zone `mantrixflow.com`:

| Type | Name | Target | Proxy |
| --- | --- | --- | --- |
| CNAME | `status` | target from Better Stack | **DNS only** (grey) |

3. Test:

```bash
curl -sS https://status.mantrixflow.com/index.json | jq '.data.attributes.aggregate_state'
```

The app sidebar and website footer poll this JSON automatically.

---

## Step 5 — Automatic reports (Better Stack UI)

Status page **Settings** → enable **Automatic reports** so monitor failures update the public page without PostHog.

---

## Step 6 — Collect IDs (Better Stack API)

Run locally (token from Step 1):

```bash
export BETTERSTACK_API_TOKEN="your-token"

curl -s -H "Authorization: Bearer $BETTERSTACK_API_TOKEN" \
  https://uptime.betterstack.com/api/v2/status-pages \
  | jq '.data[] | {id, company: .attributes.company_name}'

export BETTERSTACK_STATUS_PAGE_ID="paste-page-id-here"

curl -s -H "Authorization: Bearer $BETTERSTACK_API_TOKEN" \
  "https://uptime.betterstack.com/api/v2/status-pages/${BETTERSTACK_STATUS_PAGE_ID}/resources" \
  | jq '.data[] | {id, public_name: .attributes.public_name}'
```

Save:

| Value | SSM name (deployment doc) |
| --- | --- |
| API token | `BETTERSTACK_API_TOKEN` |
| Page `id` | `BETTERSTACK_STATUS_PAGE_ID` |
| Resource `id` for App | `BETTERSTACK_RESOURCE_APP` |
| Resource `id` for API | `BETTERSTACK_RESOURCE_API` |
| ELT resource (optional) | `BETTERSTACK_RESOURCE_ELT` — skip if you have no ELT row |

**Next:** [observability-deployment.md](./observability-deployment.md) — write these into AWS SSM and redeploy the API.

---

## Troubleshooting (Better Stack only)

| Symptom | Fix |
| --- | --- |
| Monitor Down | Fix `curl` to the same URL; check Vercel / ECS / Cloudflare |
| Status page 404 | CNAME `status` not propagated or wrong target |
| Wrong component on page | **Structure** tab — link monitors to resources |
| SSL errors on API monitor | `cloud.api` DNS-only in Cloudflare |

---

## Related

- [posthog-setup.md](./posthog-setup.md) — PostHog (separate product)
- [observability-deployment.md](./observability-deployment.md) — AWS SSM + Vercel + redeploy
