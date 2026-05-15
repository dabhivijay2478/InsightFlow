# Slack Marketplace Review Checklist

Use this before submitting the MantrixFlow Slack app for Marketplace review.

## App Functionality

- Required Slack value is visible without optional scopes:
  - Pipeline lifecycle alerts post to the selected incoming webhook channel.
  - `/pipeline list`, `/pipeline status`, `/pipeline run`, `/pipeline pause`
    work for linked users.
  - `/connection list` and `/connection test` work for linked users.
  - `/mantrixflow link` maps a Slack user through a short-lived MantrixFlow URL.
- Run and pause commands use confirm/cancel buttons.
- Complex flows stay in the web app:
  - no pipeline creation from Slack
  - no connection creation or credentials in Slack
  - no query result browsing in Slack
  - no Slack AI agent behavior

## Install And OAuth

- Direct Install URL is configured:

```text
https://YOUR_APP_HOST/api/slack/install
```

- The Direct Install URL returns an HTTP `302` to Slack OAuth.
- OAuth callback is configured:

```text
https://YOUR_APP_HOST/api/slack/oauth/callback
```

- Required scopes are minimal:

```text
commands
chat:write
incoming-webhook
```

- Optional email sync scopes are only used from the advanced settings flow:

```text
users:read
users:read.email
```

- Test both install paths:
  - Add to Slack from MantrixFlow Settings -> Integrations.
  - Install from Slack Marketplace Direct Install URL, then bind to an org.

## Security And Privacy

- Go verifies Slack request signatures with `SLACK_SIGNING_SECRET`.
- Request verification uses the raw body and rejects stale timestamps.
- Bot tokens and webhook URLs are encrypted before persistence.
- API responses expose only presence booleans, never bot tokens or webhook URLs.
- Slash-command responses are ephemeral by default.
- Connection-test output is sanitized and never includes credentials or raw config.
- `app_uninstalled` marks the integration disconnected and clears encrypted
  Slack secrets.
- Direct client access to Slack tables remains closed in `supabase_rls.sql`.

## Listing Assets

- Public landing page explains:
  - Slack alerts for pipeline started, completed, failed, and slow runs.
  - Slash-command actions and role requirements.
  - How users link their Slack account.
- Public privacy policy URL is live.
- Public support URL or contact page is live.
- Short description is 10 words or fewer.
- Screenshots show real Slack alerts and command responses, not credentials.
- Demo video covers install, self-link, pipeline list/status, run confirm, and
  test alert.

## Staging And Resubmission

- Maintain a separate staging Slack app with the same manifest shape.
- Test new scopes, events, redirect URLs, and commands on staging before changing
  production.
- Any new Slack capability or new scope should be submitted for review before
  relying on it in production.

## Final Verification

Run:

```bash
cd apps/server/main-server
GOCACHE=$(pwd)/.gocache-test go test ./internal/server/... ./internal/database/...
```

```bash
cd apps/app
bun run lint
bun run build
```
