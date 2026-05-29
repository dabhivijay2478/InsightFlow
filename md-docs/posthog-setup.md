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

## Step 6 — Optional: Error tracking alerts → Slack

Skip if you only want the PostHog UI for analytics and issues.  
This uses PostHog’s **3-step wizard**: **Destination** → **Trigger** → **Configure** ([docs](https://posthog.com/docs/error-tracking/alerts)).

Slack alerts notify your team in-channel. They do **not** update [status.mantrixflow.com](https://status.mantrixflow.com) — that stays on Better Stack ([betterstack-setup.md](./betterstack-setup.md)).

### Start the wizard

1. **Error tracking** → **Configuration** → **Alerting**.
2. **New notification** (or **+ New alert**).

---

### Step 6.1 — Destination (you chose Slack)

Wizard step **1 — Destination**:

PostHog’s **new** error-tracking alert wizard often shows only:

- **Slack**
- **Discord**
- **Microsoft Teams**

**HTTP Webhook is usually not on this screen.** That is normal — you are not missing a step. For MantrixFlow, **Slack is enough** for team notifications.

1. Select **Slack**.
2. If prompted to connect Slack the first time:
   - **Install PostHog to Slack** / **Connect workspace** → approve in Slack.
   - Pick the workspace MantrixFlow uses (e.g. engineering).
3. Click **Continue** (step 1 shows a green check when done).

---

### Step 6.2 — Trigger

Wizard step **2 — Trigger** — screen title: **“What should trigger the alert?”**

Pick what you want notified about. For MantrixFlow production, use **one notification per trigger** (simplest), or combine if the UI allows multiple:

| Trigger card | When to use | MantrixFlow recommendation |
| --- | --- | --- |
| **Issue created** | New error issue first seen | **Enable** — new bugs after deploy |
| **Issue reopened** | Resolved issue came back | **Enable** — regressions |
| **Issue spiking** | Existing issue volume jumps | **Optional** — good after bad deploys ([spike docs](https://posthog.com/docs/error-tracking/spikes)) |

**Suggested minimum:** create **two** alerts (same Slack destination):

1. Notification A: **Issue created** → channel e.g. `#mantrixflow-alerts`
2. Notification B: **Issue reopened** → same channel

Add a third for **Issue spiking** if you want deploy/regression volume warnings without waiting for a brand-new issue fingerprint.

Click **Continue** after selecting the trigger for this notification.

---

### Step 6.3 — Configure

Wizard step **3 — Configure**:

| Field | What to set |
| --- | --- |
| **Slack channel** | e.g. `#engineering`, `#mantrixflow-alerts`, or `#incidents` |
| **Name** | e.g. `MantrixFlow — issue created` / `MantrixFlow — issue reopened` |
| **Filters** (if shown) | Optional: property `service` = `main-server` (API only) or leave empty for all services |

Use filters when you want separate channels per service:

| Filter | Example channel |
| --- | --- |
| `service` = `main-server` | `#api-alerts` |
| `service` = `elt-server` | `#elt-alerts` |
| `service` = `app-server` | `#frontend-alerts` |

Finish:

1. **Test function** (or **Send test**) — confirm a message appears in Slack.
2. **Create & enable**.

Repeat **New notification** for each trigger (created / reopened / spiking) if you did not combine them in one rule.

---

### Slack vs public status page

| Channel | What it does |
| --- | --- |
| **Slack (this step)** | Team notifications for PostHog **Issues** |
| **Better Stack** | Public uptime + `status.mantrixflow.com` |

No change to MantrixFlow code is required for Slack alerts.

---

## Step 7 — Optional: **HTTP Webhook** destination (real PostHog UI)

Use this **only** if you want `$exception` events sent to the MantrixFlow Go API (e.g. to drive Better Stack status reports).  
**Skip entirely** if Slack alerts (Step 6) are enough.

The **Error tracking → Alerting** wizard (Slack / Discord / Teams) does **not** show HTTP Webhook. Webhook lives under **Data pipeline** as destination type **HTTP Webhook**.

#### Open the destination

1. Left sidebar → **Data pipeline** (may appear under **Tools** in the sidebar).
2. **+ New** → **Destination**.
3. Search and open **HTTP Webhook** (you land on the configuration page titled **HTTP Webhook**).

---

#### Fill the form (matches current PostHog UI)

##### Status

| UI section | What to do |
| --- | --- |
| **Enable destination** | Turn **on** (destination must be enabled to send) |

##### Filters

| UI field | What to do |
| --- | --- |
| **Filter out internal and test users** | Optional — turn **on** if you use PostHog’s internal/test user filters |
| **Match events and actions** | **Add a match** so only errors fire the webhook: event name **equals** `$exception` |
| **Trigger options** | Leave default unless PostHog shows a specific trigger mode you need |
| **Matching events** | Shows e.g. *“would have triggered 0 times in the last 7 days”* — **normal until** `$exception` events exist in this project. After app/API send errors (Step 4), refresh; count should rise. If still 0, check project key and **Match events** filter |

To alert only the Go API (optional second destination):

- Add property filter: `service` **equals** `main-server`  
  Repeat with `elt-server` / `app-server` if you want separate destinations per service.

##### Webhook URL

| UI field | Value |
| --- | --- |
| **Webhook URL** | `https://cloud.api.mantrixflow.com/api/v1/internal/incident-webhook` |

##### Method

| UI field | Value |
| --- | --- |
| **Method** (optional) | `POST` |

##### JSON Body

| UI field | Value |
| --- | --- |
| **JSON Body** (optional) | Paste (Hog replaces `{...}` at send time): |

```json
{
  "hook": "PostHog HTTP Webhook",
  "event": "alert.triggered",
  "data": {
    "name": "{event.properties.$exception_message}",
    "tags": ["service:{event.properties.service}"]
  }
}
```

If `$exception_message` is empty in tests, PostHog still sends the event; the Go API accepts other shapes too.

##### Headers

| UI field | Value |
| --- | --- |
| **Content-Type** | `application/json` |
| **Headers** (optional) | Add a row: name `X-Internal-Token`, value = your `INTERNAL_TOKEN` (from secrets / [observability-deployment.md](./observability-deployment.md) — do not paste in git) |

Without `X-Internal-Token`, the Go API returns **401**.

##### Log responses

| UI field | What to do |
| --- | --- |
| **Log responses** (optional) | Turn **on** while testing (see HTTP status in PostHog logs); turn off later if noisy |

##### Edit source

| UI field | What to do |
| --- | --- |
| **Edit source** | Leave default unless you need custom Hog code — not required for MantrixFlow |

##### Testing

| UI field | What to do |
| --- | --- |
| **Testing** | Click **test your function with an example event** → confirm **200** from `cloud.api.mantrixflow.com` when API is deployed and token is correct |

Then **Save** / enable the destination (wording may be **Create & enable** depending on create vs edit flow).

---

#### After save

| Check | Expected |
| --- | --- |
| Real `$exception` from production | Webhook runs when errors hit PostHog with matching filters |
| **Matching events** still 0 | No `$exception` in last 7 days — deploy app with `phc_...` key first |
| Better Stack status page | Needs `BETTERSTACK_*` in SSM + [betterstack-setup.md](./betterstack-setup.md) — webhook alone is not enough |

**Slack (Step 6)** and **HTTP Webhook (Step 7)** are independent — use Slack for team chat, Webhook only if you need the Go API bridge.

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
| Slack test message missing | Slack app not installed; wrong channel; workspace not connected |
| No Webhook on Error tracking wizard | Expected — use **Data pipeline** → **HTTP Webhook** (Step 7) |
| Matching events “0 in 7 days” | No `$exception` yet — use app/API with project key; fix **Match events** filter |
| Webhook test fails | API not deployed; missing/wrong `X-Internal-Token` header; wrong URL |
| Webhook 200 but status page unchanged | Configure Better Stack SSM — webhook only hits Go API |
| Events without `service` | Filter only backend issues; browser events may omit `service` until you add it in custom captures |

---

## Related (not PostHog UI)

- [observability-deployment.md](./observability-deployment.md) — put `phc_...` into Vercel / AWS SSM and redeploy
- [betterstack-setup.md](./betterstack-setup.md) — public status page (separate product)
