# MantrixFlow Slack OAuth Setup

This guide configures the OAuth-first Slack app used by MantrixFlow.

If `/api/v1/organizations/:orgId/slack/oauth/start` returns:

```json
{
  "error": {
    "code": "SLACK_UNAVAILABLE",
    "message": "Slack OAuth is not configured"
  }
}
```

the Go API did not load `SLACK_CLIENT_ID` and `SLACK_CLIENT_SECRET`.

## Local URLs

Run both services:

```bash
cd apps/server/main-server
go run ./cmd/server
```

```bash
cd apps/app
bun run dev
```

Local defaults:

```text
Next app: http://localhost:3000
Go API:   http://localhost:5000
```

Slack cannot call `localhost` from its cloud. For real Slack commands, actions,
events, and OAuth callbacks in local development, expose the Next app with a
public HTTPS tunnel:

```bash
ngrok http 3000
```

Use the HTTPS forwarding URL as `APP_WEB_URL` and in the Slack Dashboard URLs.

## Create The Slack App

Go to `https://api.slack.com/apps`:

1. Create New App -> From scratch.
2. Choose the development workspace.
3. Open Basic Information and copy:
   - Client ID
   - Client Secret
   - Signing Secret

## OAuth And Permissions

Open OAuth & Permissions.

Add this Redirect URL:

```text
https://YOUR_APP_HOST/api/slack/oauth/callback
```

For local dev with ngrok:

```text
https://YOUR_NGROK_HOST/api/slack/oauth/callback
```

Add required Bot Token Scopes:

```text
commands
chat:write
incoming-webhook
```

Optional advanced email-sync scopes:

```text
users:read
users:read.email
```

The `incoming-webhook` scope makes Slack ask the installer to select the alert
channel during OAuth. MantrixFlow stores that webhook encrypted in Go.

## Slash Commands

Open Slash Commands and create:

```text
/pipeline   -> https://YOUR_APP_HOST/api/slack/commands
/connection -> https://YOUR_APP_HOST/api/slack/commands
/mantrixflow -> https://YOUR_APP_HOST/api/slack/commands
```

For local dev, use the ngrok host instead of `localhost`.

## Interactivity

Open Interactivity & Shortcuts.

Enable interactivity and set:

```text
https://YOUR_APP_HOST/api/slack/actions
```

This is required for confirm/cancel buttons on run and pause commands.

## Event Subscriptions

Open Event Subscriptions.

Enable events and set:

```text
https://YOUR_APP_HOST/api/slack/events
```

The app uses this for Slack URL verification and the `app_uninstalled` event.
If Slack asks for verification, the Go endpoint returns the challenge after
validating Slack's signature.

Subscribe to this bot event:

```text
app_uninstalled
```

## Environment Variables

The Go API must load these values:

```bash
SLACK_CLIENT_ID=123456789.123456789
SLACK_CLIENT_SECRET=your-slack-client-secret
SLACK_SIGNING_SECRET=your-slack-signing-secret
APP_WEB_URL=https://YOUR_APP_HOST
```

For local dev without a tunnel, `APP_WEB_URL=http://localhost:3000` lets the
settings UI start OAuth, but Slack will not be able to call the callback. Use a
public HTTPS tunnel for an actual install.

Optional:

```bash
SLACK_API_BASE_URL=https://slack.com/api
```

The Next proxy uses:

```bash
NEXT_PUBLIC_API_URL=http://localhost:5000
SLACK_PROXY_TARGET_URL=http://localhost:5000
```

`SLACK_PROXY_TARGET_URL` is optional. If omitted, the proxy falls back to
`NEXT_PUBLIC_API_URL`.

### Where To Put Env Vars Locally

When you run Go from `apps/server/main-server`, `config.Load()` reads:

```text
apps/server/main-server/.env
apps/app/.env
```

The easiest local setup is to add the Slack vars to `apps/app/.env`, because the
Go server already loads `../../app/.env` from `apps/server/main-server`.
If you start Go from a different working directory, export the vars in your
shell or put them in the `.env` file for that working directory.

Example:

```bash
NEXT_PUBLIC_APP_URL=http://localhost:3000
NEXT_PUBLIC_API_URL=http://localhost:5000

APP_WEB_URL=https://YOUR_NGROK_HOST
SLACK_CLIENT_ID=123456789.123456789
SLACK_CLIENT_SECRET=your-slack-client-secret
SLACK_SIGNING_SECRET=your-slack-signing-secret
```

Restart the Go API after changing env vars.

## Install From MantrixFlow

1. Open `http://localhost:3000/workspace/settings?tab=integrations`.
2. Click Add to Slack.
3. Approve the Slack OAuth screen.
4. Choose the alert channel when Slack asks for incoming webhook access.
5. You should land back on Settings -> Integrations with Slack connected.
6. Click Sync users if needed.
7. Click Test alert.

## Expected Behavior

After OAuth succeeds:

- `slack_integrations` stores the bot token and incoming webhook URL encrypted.
- Required-scope installs rely on `/mantrixflow link` for Slack user mapping.
- Optional email sync can populate `slack_connections` by matching Slack user
  email to active MantrixFlow organization member email.
- Unmatched Slack users receive an ephemeral rejection when using commands.
- Alerts post to the OAuth-selected incoming webhook/channel.

## Troubleshooting

### `SLACK_UNAVAILABLE: Slack OAuth is not configured`

Go did not load `SLACK_CLIENT_ID` or `SLACK_CLIENT_SECRET`.

Fix:

1. Add both env vars to the Go-loaded env file.
2. Restart `go run ./cmd/server`.
3. Re-open Settings -> Integrations and click Add to Slack again.

### Slack says the redirect URL is invalid

The Redirect URL in Slack Dashboard must exactly match:

```text
https://YOUR_APP_HOST/api/slack/oauth/callback
```

Protocol, host, and path must match. A new ngrok URL means updating Slack
Dashboard and `APP_WEB_URL`.

### Slash commands do not reach local dev

Slack cannot reach `localhost`. Use a public HTTPS tunnel to the Next app and
set the Slash Commands, Interactivity, Events, and OAuth Redirect URLs to that
tunnel host.

### Commands say the Slack user is not linked

Run `/mantrixflow link` in Slack and open the short-lived link while signed in
to MantrixFlow. Owners can also enable email sync in Settings -> Integrations or
use the Advanced manual mapping fallback.
