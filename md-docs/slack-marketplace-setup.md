# MantrixFlow Slack Marketplace Setup

This guide is for the distributed Slack app that can be submitted to the Slack
Marketplace. Slack must call the public Next.js app host; Next proxies the raw
request to the Go API.

## Public URLs

Production:

```text
Direct install URL: https://YOUR_APP_HOST/api/slack/install
OAuth callback:     https://YOUR_APP_HOST/api/slack/oauth/callback
Slash commands:     https://YOUR_APP_HOST/api/slack/commands
Interactivity:      https://YOUR_APP_HOST/api/slack/actions
Events:             https://YOUR_APP_HOST/api/slack/events
```

Local development with ngrok:

```bash
cd apps/app
bun run dev

ngrok http 3000
```

Use the ngrok HTTPS host everywhere in Slack Dashboard. The tunnel points to the
frontend port, not the Go API port:

```text
https://YOUR_NGROK_HOST/api/slack/install
https://YOUR_NGROK_HOST/api/slack/oauth/callback
```

The Go API still runs on `http://localhost:5000`; Next forwards Slack traffic to
`SLACK_PROXY_TARGET_URL || NEXT_PUBLIC_API_URL`.

## Slack App Configuration

Create or update the Slack app at `https://api.slack.com/apps`.

1. Enable public distribution under Manage Distribution.
2. Set the Direct Install URL to:

```text
https://YOUR_APP_HOST/api/slack/install
```

3. Add OAuth redirect URL:

```text
https://YOUR_APP_HOST/api/slack/oauth/callback
```

4. Add required bot scopes:

```text
commands
chat:write
incoming-webhook
```

5. Add optional bot scopes for advanced email sync:

```text
users:read
users:read.email
```

6. Add slash commands:

```text
/pipeline     -> https://YOUR_APP_HOST/api/slack/commands
/connection   -> https://YOUR_APP_HOST/api/slack/commands
/mantrixflow  -> https://YOUR_APP_HOST/api/slack/commands
```

7. Set Interactivity URL:

```text
https://YOUR_APP_HOST/api/slack/actions
```

8. Set Events Request URL:

```text
https://YOUR_APP_HOST/api/slack/events
```

9. Subscribe to the `app_uninstalled` bot event so MantrixFlow can clear local
   encrypted Slack secrets when a workspace removes the app.

## Environment Variables

Go API:

```bash
APP_WEB_URL=https://YOUR_APP_HOST
SLACK_CLIENT_ID=123456789.123456789
SLACK_CLIENT_SECRET=your-slack-client-secret
SLACK_SIGNING_SECRET=your-slack-signing-secret
SLACK_API_BASE_URL=https://slack.com/api
```

Next app:

```bash
NEXT_PUBLIC_API_URL=https://YOUR_GO_API_HOST
SLACK_PROXY_TARGET_URL=https://YOUR_GO_API_HOST
```

`SLACK_PROXY_TARGET_URL` is optional when `NEXT_PUBLIC_API_URL` already points
to the Go API.

## Install Flows

Owner-started install:

1. MantrixFlow owner opens Settings -> Integrations.
2. Owner clicks Add to Slack.
3. Slack OAuth asks for required scopes and alert channel.
4. Go stores the bot token and incoming webhook encrypted.

Marketplace-started install:

1. Slack Marketplace calls `GET /api/slack/install`.
2. Next proxies to Go and returns Slack's OAuth `302`.
3. Go stores the OAuth callback as a short-lived pending install.
4. The browser returns to Settings -> Integrations with a claim token.
5. A MantrixFlow owner binds the Slack workspace to one organization.

User mapping:

- Required-scope path: users run `/mantrixflow link` and open the short-lived
  self-link URL.
- Optional email sync path: owners enable email sync, approve `users:read` and
  `users:read.email`, then click Sync users.
- Owner fallback: Advanced manual mapping remains available for unmatched users.

## Dashboard Notes

- Slack Dashboard URLs must use the Next app host.
- Keep the Go API private behind the normal API domain.
- One Slack workspace can be connected to one MantrixFlow organization in V1.
- Do not include credentials, DSNs, or raw connection config in Slack messages.

## References

- Slack Marketplace distribution: https://docs.slack.dev/slack-marketplace/distributing-your-app-in-the-slack-marketplace/
- Slack app distribution: https://docs.slack.dev/app-management/distribution/
- Slack OAuth: https://docs.slack.dev/authentication/installing-with-oauth/
- Slack request verification: https://docs.slack.dev/authentication/verifying-requests-from-slack/
