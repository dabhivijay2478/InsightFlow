# PostHog setup (MantrixFlow — US cloud)

Everything in this file is done **inside PostHog** ([us.posthog.com](https://us.posthog.com)).  
No AWS, Vercel, or infra steps here — use [observability-deployment.md](./observability-deployment.md) only when you are ready to deploy env vars.

MantrixFlow code already ships PostHog SDKs (Next.js, Go, Python). This guide configures the **PostHog project** and verifies data in the product UI.

---

## Your MantrixFlow project

| Setting | Value |
| --- | --- |
| **Region** | US |
| **App URL** | [https://us.posthog.com](https://us.posthog.com) |
| **Ingestion host** | `https://us.i.posthog.com` |
| **Project API key** | `phc_oosh6d4WYMyG6U9dx9ErJdYU6yhCsGQdaN8eqCJ6UxXG` |

Use the **Project API key** (`phc_...`) — not a Personal API key.  
Treat the key as a secret: keep it in env files or a password manager; do not commit it to public repos.

Repo env names (for later deployment):

```bash
NEXT_PUBLIC_POSTHOG_KEY=phc_oosh6d4WYMyG6U9dx9ErJdYU6yhCsGQdaN8eqCJ6UxXG
NEXT_PUBLIC_POSTHOG_HOST=https://us.i.posthog.com
POSTHOG_API_KEY=phc_oosh6d4WYMyG6U9dx9ErJdYU6yhCsGQdaN8eqCJ6UxXG
POSTHOG_HOST=https://us.i.posthog.com
```

---

## What you should see in PostHog

| MantrixFlow surface | PostHog events | `service` on errors |
| --- | --- | --- |
| Browser app | `$pageview`, session replay, `$exception` | (browser) |
| Next.js server | `$exception` | `app-server` |
| Go API | `$exception` | `main-server` |
| Python ELT | `$exception`, pipeline events | `elt-server` |

PostHog recommends **one project** for marketing site + app + API so journeys stay in one place ([install docs](https://posthog.com/docs/getting-started/install)).

---

## Step 1 — Open the correct project (PostHog UI)

1. Go to [https://us.posthog.com](https://us.posthog.com) and sign in.
2. Top-left **project switcher** → select the MantrixFlow project.
3. Left sidebar → **Settings** (gear) → **Project** → **General** (or **Project details**).
4. Confirm **Project API key** is:

   `phc_oosh6d4WYMyG6U9dx9ErJdYU6yhCsGQdaN8eqCJ6UxXG`

If it does not match, you are in the wrong project — switch projects before continuing.

---

## Step 2 — Error tracking (PostHog UI)

PostHog docs: [Error tracking alerts](https://posthog.com/docs/error-tracking/alerts), [Next.js installation](https://posthog.com/docs/error-tracking/installation/nextjs).

1. Left sidebar → **Error tracking**.
2. Open **Configuration** (gear on that page, or the **Configuration** tab).
3. Under **Exception autocapture** / project settings:
   - Enable **Web** / JavaScript exception autocapture (captures `window.onerror`, unhandled rejections).
   - Leave **Capture console errors** off unless you explicitly want `console.error` as issues.
4. **Save** if the UI shows a save button.

### Optional: spike detection

Same **Configuration** area → **Spike detection** → review defaults (alerts when an existing issue spikes after a deploy). Enable if you want volume-based alerts in addition to new issues.

---

## Step 3 — Session replay & autocapture (PostHog UI)

1. Left sidebar → **Settings** → **Project** → **Product analytics** (or **Autocapture** / **Session replay** depending on UI version).
2. Confirm **Autocapture** is on for web (clicks, pageviews — MantrixFlow also sends `$pageview` from code).
3. **Session replay** → enable if you want recordings; MantrixFlow masks inputs in code (`maskAllInputs: true`).

### Firewall note (if replay/heatmaps fail)

If your WAF blocks PostHog, allow these **US** IPs ([install docs](https://posthog.com/docs/getting-started/install)):

`44.205.89.55`, `52.4.194.122`, `44.208.188.173`

---

## Step 4 — Verify events in PostHog (no code changes)

### 4.1 Live stream

1. Left sidebar → **Activity** (or **Events** → live view).
2. Open [https://cloud.mantrixflow.com](https://cloud.mantrixflow.com) in another tab (after deployment) or run the app locally with the project key in `.env.local`.
3. Click around the app — within ~1 minute you should see **`$pageview`** (and other autocapture events) in Activity.

### 4.2 Toolbar (quick test)

1. In PostHog, open **Launch toolbar** (or **Tools** → **Toolbar**) for your site URL `https://cloud.mantrixflow.com`.
2. Confirm the toolbar loads — proves the project key and host work in the browser.

### 4.3 Error tracking issues

1. **Error tracking** → **Issues**.
2. Trigger a test error (e.g. temporary throw in dev, or a known 5xx on API after deploy).
3. A new **issue** should appear with stack trace (source maps improve traces — configure later under **Error tracking** → **Source maps** if you use minified bundles).

### 4.4 Filter by `service` (PostHog UI)

In **Error tracking** → **Issues** or **Activity**, add a property filter:

- `service` = `main-server` (Go API)
- `service` = `elt-server` (ELT)
- `service` = `app-server` (Next server)

This matches what MantrixFlow sends from code.

---

## Step 5 — Identify users (PostHog UI + product)

PostHog chapter: [Identify users](https://posthog.com/docs/getting-started/identify).

When users log in, the app should call `posthog.identify()` (wire in product code if not already). In PostHog you verify:

1. **Persons** → search for a user after login.
2. **Activity** → events show a stable `distinct_id` (not only anonymous IDs).

No extra PostHog UI toggle is required for identify — it is event-driven.

---

## Step 6 — Optional: Error tracking alert → webhook

**Only step that touches MantrixFlow API.** Skip if you only want analytics + issues inside PostHog.

Use PostHog’s **Error tracking → Alerting** UI ([docs](https://posthog.com/docs/error-tracking/alerts)) — not legacy **Actions**, not **Trend → Alerts** (email only).

### 6.1 Create notification

1. **Error tracking** → **Configuration** → section **Alerting**.
2. **New notification**.
3. Choose **Webhook**.

### 6.2 Webhook URL

PostHog’s alerting screen has a **URL** field only (no custom headers). Use:

```text
https://cloud.api.mantrixflow.com/api/v1/internal/incident-webhook?internal_token=YOUR_INTERNAL_TOKEN
```

Replace `YOUR_INTERNAL_TOKEN` with the same secret the Go API uses for internal routes (from your secrets store — not stored in this doc).

The API accepts `?internal_token=` on this path.

### 6.3 Filters (optional)

On the same screen, filter issues by properties if needed (e.g. only `service: main-server`).

### 6.4 Test and enable

1. Click **Test function** — expect HTTP **200** from `cloud.api.mantrixflow.com`.
2. Click **Create & enable**.

Alerts fire when an **issue is created or reopened**. Uptime on [status.mantrixflow.com](https://status.mantrixflow.com) still comes from Better Stack monitors ([betterstack-setup.md](./betterstack-setup.md)).

---

## Already implemented in MantrixFlow (reference)

You do **not** re-run the install wizard for production if the repo is already wired:

| Layer | Path |
| --- | --- |
| Next.js provider | `apps/app/components/providers/posthog-provider.tsx` |
| Client init | `apps/app/lib/posthog/client.ts` |
| Server `onRequestError` | `apps/app/instrumentation.ts` |
| Go API | `apps/server/main-server/internal/observability/posthog.go` |
| ELT | `apps/server/elt-server/core/observability.py` |

New greenfield app install (local only):

```bash
cd apps/app
npx -y @posthog/wizard@latest --region us
```

---

## PostHog UI troubleshooting

| Symptom | What to check in PostHog |
| --- | --- |
| No events in Activity | Wrong project (Step 1 key); ad blocker on `us.i.posthog.com`; app not using `phc_oosh6d4...` |
| No issues | **Error tracking** → **Configuration** → exception autocapture off |
| Toolbar does not load | Site URL wrong; WAF blocking PostHog IPs |
| Webhook test fails | API not deployed; wrong `internal_token`; URL must be `cloud.api.mantrixflow.com` |
| Events without `service` | Filter only backend issues; browser events may omit `service` until you add it in custom captures |

---

## Related (not PostHog UI)

- [observability-deployment.md](./observability-deployment.md) — put `phc_...` into Vercel / AWS SSM and redeploy
- [betterstack-setup.md](./betterstack-setup.md) — public status page (separate product)
