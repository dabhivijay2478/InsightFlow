# MantrixFlow Slack App Guide

This is the single setup and review guide for the MantrixFlow Slack app.

MantrixFlow uses Slack as a Slack-native admin surface for pipeline operations:
OAuth install, App Home, writable direct messages, slash commands, confirm/cancel
buttons, pipeline lifecycle alerts, channel-level command access, optional
personal identity linking, and saved-connection pipeline creation. Go remains
the source of truth for OAuth exchange, request verification, authorization,
encrypted token/webhook storage, Slack actions, pipeline creation, queueing, and
alerts.

## 1. Architecture

Slack Dashboard must point to the public Next.js app host:

```text
Slack -> Next /api/slack/* -> Go /api/v1/slack/*
```

Local defaults:

```text
Next app: http://localhost:3000
Go API:   http://localhost:5000
```

For local Slack testing, expose the frontend port, not the Go port:

```bash
cd apps/app
bun run dev

ngrok http 3000
```

The current local ngrok host is:

```text
https://b7c9-2409-40c1-5011-3ea3-2836-61f9-1495-f47c.ngrok-free.app
```

If ngrok changes, replace this host everywhere in Slack Dashboard and env vars.

## 2. Slack URLs

Use these URLs in Slack Dashboard for local dev:

```text
Direct install URL: https://b7c9-2409-40c1-5011-3ea3-2836-61f9-1495-f47c.ngrok-free.app/api/slack/install
OAuth callback:     https://b7c9-2409-40c1-5011-3ea3-2836-61f9-1495-f47c.ngrok-free.app/api/slack/oauth/callback
Slash commands:     https://b7c9-2409-40c1-5011-3ea3-2836-61f9-1495-f47c.ngrok-free.app/api/slack/commands
Interactivity:      https://b7c9-2409-40c1-5011-3ea3-2836-61f9-1495-f47c.ngrok-free.app/api/slack/actions
Events:             https://b7c9-2409-40c1-5011-3ea3-2836-61f9-1495-f47c.ngrok-free.app/api/slack/events
```

Production shape:

```text
Direct install URL: https://YOUR_APP_HOST/api/slack/install
OAuth callback:     https://YOUR_APP_HOST/api/slack/oauth/callback
Slash commands:     https://YOUR_APP_HOST/api/slack/commands
Interactivity:      https://YOUR_APP_HOST/api/slack/actions
Events:             https://YOUR_APP_HOST/api/slack/events
```

Keep `APP_WEB_URL` as the browser app host, not the Slack callback path. In
local dev it should stay on localhost so the OAuth result returns to the same
host where you are already signed in.

## 3. Environment Variables

Go API:

```bash
APP_WEB_URL=http://localhost:3000
PUBLIC_APP_URL=http://localhost:3000
SLACK_OAUTH_REDIRECT_BASE_URL=https://b7c9-2409-40c1-5011-3ea3-2836-61f9-1495-f47c.ngrok-free.app
SLACK_CLIENT_ID=123456789.123456789
SLACK_CLIENT_SECRET=your-slack-client-secret
SLACK_SIGNING_SECRET=your-slack-signing-secret
SLACK_API_BASE_URL=https://slack.com/api
CORS_ALLOWED_ORIGINS=http://localhost:3000,http://127.0.0.1:3000,https://b7c9-2409-40c1-5011-3ea3-2836-61f9-1495-f47c.ngrok-free.app
```

Next app:

```bash
NEXT_PUBLIC_APP_URL=http://localhost:3000
NEXT_PUBLIC_SITE_URL=http://localhost:3000
NEXT_PUBLIC_API_URL=http://localhost:5000
SLACK_PROXY_TARGET_URL=http://localhost:5000
NEXT_PUBLIC_ALLOWED_ORIGINS=https://cloud.mantrixflow.com,https://cloud.api.mantrixflow.com,https://cloud.api.etl.server.mantrixflow.com,https://b7c9-2409-40c1-5011-3ea3-2836-61f9-1495-f47c.ngrok-free.app
```

`SLACK_PROXY_TARGET_URL` is optional when `NEXT_PUBLIC_API_URL` already points to
the Go API.

When you run Go from `apps/server/main-server`, `config.Load()` reads:

```text
apps/server/main-server/.env
apps/app/.env
```

The easiest local setup is to keep Slack env vars in `apps/app/.env`, then
restart both dev servers after changes.

## 4. Create Or Update The Slack App

Go to `https://api.slack.com/apps`.

1. Create New App -> From scratch, or open the existing MantrixFlow app.
2. Choose the development workspace.
3. Open Basic Information and copy:
   - Client ID
   - Client Secret
   - Signing Secret
4. Add those values to the Go API env vars.
5. Restart the Go API and Next app.

## 5. OAuth And Marketplace Install

Open OAuth & Permissions.

Add this redirect URL for local dev:

```text
https://b7c9-2409-40c1-5011-3ea3-2836-61f9-1495-f47c.ngrok-free.app/api/slack/oauth/callback
```

For production:

```text
https://YOUR_APP_HOST/api/slack/oauth/callback
```

Required bot scopes:

```text
commands
chat:write
incoming-webhook
im:history
im:write
```

Optional advanced email-sync scopes:

```text
users:read
users:read.email
```

Keep the primary Marketplace install minimal. Required-scope installs use the
OAuth-selected channel for command access. Add `users:read` and
`users:read.email` only if you later want optional owner-controlled email sync.

Slack shows "Channel for webhook" because the app requests `incoming-webhook`.
That channel is not a per-user mapping step. In MantrixFlow V1 it is the
workspace channel where alerts post and where everyone in that channel can run
slash commands.

For Marketplace/public distribution, open Manage Distribution and set:

```text
Direct Install URL:
https://b7c9-2409-40c1-5011-3ea3-2836-61f9-1495-f47c.ngrok-free.app/api/slack/install
```

The direct install URL must return an HTTP `302` to Slack OAuth.

Reinstall the app after changing scopes, App Home, slash commands, events, or
interactivity settings.

## 6. App Home And Messages Tab

Open App Home and enable:

```text
Home Tab: on
Messages Tab: on
Allow users to send Slash commands and messages from the Messages tab: on
```

In manifest form:

```yaml
features:
  app_home:
    home_tab_enabled: true
    messages_tab_enabled: true
    messages_tab_read_only_enabled: false
```

This fixes Slack's "Sending messages to this app has been turned off" screen.

## 7. Slash Commands

Create these commands:

```text
/pipeline    -> https://b7c9-2409-40c1-5011-3ea3-2836-61f9-1495-f47c.ngrok-free.app/api/slack/commands
/connection  -> https://b7c9-2409-40c1-5011-3ea3-2836-61f9-1495-f47c.ngrok-free.app/api/slack/commands
/mantrixflow -> https://b7c9-2409-40c1-5011-3ea3-2836-61f9-1495-f47c.ngrok-free.app/api/slack/commands
```

Supported commands:

```text
/pipeline list
/pipeline status <name-or-id>
/pipeline run <name-or-id>
/pipeline pause <name-or-id>
/pipeline resume <name-or-id>
/pipeline cancel <name-or-id>
/pipeline runs <name-or-id>
/pipeline validate <name-or-id>
/pipeline delete <name-or-id>
/pipeline create
/connection list
/connection test <name-or-id>
/connection delete <name-or-id>
/connection create
/mantrixflow link
/mantrixflow help
```

| Command | Role Required | Description |
|---------|--------------|-------------|
| `/pipeline list` | VIEWER | List all pipelines |
| `/pipeline status <name>` | VIEWER | Show pipeline status |
| `/pipeline run <name>` | EDITOR | Run a pipeline (with confirmation) |
| `/pipeline pause <name>` | EDITOR | Pause a running pipeline |
| `/pipeline resume <name>` | EDITOR | Resume a paused pipeline |
| `/pipeline cancel <name>` | EDITOR | Cancel an active run |
| `/pipeline runs <name>` | VIEWER | List recent pipeline runs |
| `/pipeline validate <name>` | EDITOR | Validate pipeline configuration |
| `/pipeline delete <name>` | OWNER | Delete a pipeline |
| `/pipeline create` | EDITOR | Open pipeline builder modal |
| `/connection list` | VIEWER | List all connections |
| `/connection test <name>` | EDITOR | Test a connection |
| `/connection delete <name>` | OWNER | Delete a connection |
| `/connection create` | EDITOR | Open connection setup modal |
| `/mantrixflow link` | ANY | Link your Slack user |
| `/mantrixflow help` | ANY | Show help |

Commands work from the OAuth-selected Slack channel without manual user mapping.
Personal App Home and DM workflows can still use `/mantrixflow link` as an
optional identity link. Run, pause, test, and create actions re-check the
connected organization and use the MantrixFlow owner who installed Slack as the
channel actor. Delete actions require OWNER role.

## 8. Interactivity

Open Interactivity & Shortcuts.

Enable interactivity and set:

```text
https://b7c9-2409-40c1-5011-3ea3-2836-61f9-1495-f47c.ngrok-free.app/api/slack/actions
```

This is required for confirm/cancel buttons, App Home buttons, modal opens,
modal updates, and pipeline builder submissions.

## 9. Event Subscriptions

Open Event Subscriptions.

Enable events and set:

```text
https://b7c9-2409-40c1-5011-3ea3-2836-61f9-1495-f47c.ngrok-free.app/api/slack/events
```

Subscribe to bot events:

```text
app_home_opened
message.im
app_uninstalled
```

`app_home_opened` publishes the MantrixFlow Home dashboard. `message.im` powers
the deterministic DM assistant. `app_uninstalled` marks the integration
disconnected and clears encrypted Slack secrets.

During Slack URL verification, the endpoint must respond with the raw
`challenge` value and `text/plain`. The Next proxy also answers
`url_verification` directly before forwarding normal events to Go, so Slack
Dashboard verification does not depend on Go's signing-secret env being perfect
yet.

## 10. Copyable App Manifest

Use this as a local/staging manifest, replacing the ngrok host when needed:

```yaml
display_information:
  name: MantrixFlow
  description: Slack-native data pipeline admin for MantrixFlow.
  long_description: MantrixFlow brings pipeline operations, connection testing, App Home status, direct-message guidance, saved-connection pipeline creation, and lifecycle alerts into Slack. The app uses OAuth and stores installed Slack secrets encrypted in MantrixFlow. Connection credentials are captured only in MantrixFlow web/OAuth flows.
  background_color: "#111827"
features:
  app_home:
    home_tab_enabled: true
    messages_tab_enabled: true
    messages_tab_read_only_enabled: false
  bot_user:
    display_name: MantrixFlow
    always_online: false
  slash_commands:
    - command: /pipeline
      description: List, check, run, pause, or create MantrixFlow pipelines.
      usage_hint: list | status <name-or-id> | run <name-or-id> | pause <name-or-id> | create
      should_escape: false
      url: https://b7c9-2409-40c1-5011-3ea3-2836-61f9-1495-f47c.ngrok-free.app/api/slack/commands
    - command: /connection
      description: List, test, or create MantrixFlow connections.
      usage_hint: list | test <name-or-id> | create
      should_escape: false
      url: https://b7c9-2409-40c1-5011-3ea3-2836-61f9-1495-f47c.ngrok-free.app/api/slack/commands
    - command: /mantrixflow
      description: Link your Slack user to MantrixFlow.
      usage_hint: link | help
      should_escape: false
      url: https://b7c9-2409-40c1-5011-3ea3-2836-61f9-1495-f47c.ngrok-free.app/api/slack/commands
oauth_config:
  redirect_urls:
    - https://b7c9-2409-40c1-5011-3ea3-2836-61f9-1495-f47c.ngrok-free.app/api/slack/oauth/callback
  scopes:
    bot:
      - commands
      - chat:write
      - incoming-webhook
      - im:history
      - im:write
settings:
  interactivity:
    is_enabled: true
    request_url: https://b7c9-2409-40c1-5011-3ea3-2836-61f9-1495-f47c.ngrok-free.app/api/slack/actions
  event_subscriptions:
    request_url: https://b7c9-2409-40c1-5011-3ea3-2836-61f9-1495-f47c.ngrok-free.app/api/slack/events
    bot_events:
      - app_home_opened
      - message.im
      - app_uninstalled
  org_deploy_enabled: false
  socket_mode_enabled: false
  token_rotation_enabled: false
```

## 11. Install From MantrixFlow

1. Start Go:

```bash
cd apps/server/main-server
go run ./cmd/server
```

2. Start Next:

```bash
cd apps/app
bun run dev
```

3. Open:

```text
http://localhost:3000/workspace/settings?tab=integrations
```

4. Sign in on localhost.
5. Open the Slack row and click Add to Slack.
6. Approve the Slack OAuth screen.
7. Choose the Slack channel when Slack asks for incoming webhook access. This is
   the alert channel and the command-access channel.
8. You should land back on Settings -> Integrations with Slack connected.
9. Click Test alert.
10. In that Slack channel, run `/pipeline list` or `/connection list`.

Slack calls the ngrok callback, but Go redirects your browser back to localhost
settings after saving the installation.

## 12. Marketplace Install Flow

1. Slack Marketplace calls `GET /api/slack/install`.
2. Next proxies to Go and returns Slack's OAuth `302`.
3. Go stores the OAuth callback as a short-lived pending install.
4. The browser returns to MantrixFlow login/org-selection with a claim token.
5. A MantrixFlow owner binds the Slack workspace to one organization.

V1 rule: one Slack workspace maps to one MantrixFlow organization.

## 13. Channel Access And Optional Linking

Primary path:

```text
Install MantrixFlow -> choose a Slack channel -> use commands in that channel
```

Everyone in the OAuth-selected channel can use MantrixFlow slash commands. This
keeps setup simple and avoids manual Slack user mapping for normal channel
workflows.

Optional personal link:

```text
/mantrixflow link
```

The user opens a short-lived MantrixFlow link while signed in. Go creates a
`slack_connections` mapping for personal App Home and DM workflows. This is not
required for channel slash commands.

Optional email sync can be added later with `users:read` and
`users:read.email`, but it is not part of the default Marketplace-safe flow.

## 14. Slack-Native Admin Surface

App Home shows:

- connected org, access mode, MantrixFlow role used, and alert/command channel
- recent/running/failed pipelines with run/pause buttons
- saved source/destination connections with test buttons
- Add Connection and Create Pipeline actions
- owner-only sync users, test alert, disconnect, and settings controls

Messages tab supports deterministic intents:

```text
help
status
pipelines
connections
runs
alerts
run <pipeline>
pause <pipeline>
test <connection>
create pipeline
add connection
```

There is no Slack AI assistant in V1.

## 15. Pipeline Builder In Slack

Slack can create pipelines only from saved MantrixFlow connections. Slack never
collects database passwords, API tokens, DSNs, or raw connector credentials.
Connection setup opens a MantrixFlow web/OAuth deep link such as:

```text
/workspace/connections/new/postgres?role=source&source=slack
```

Wave 1 sources:

```text
postgres, mysql, mariadb, sqlite, cockroachdb, stripe, shopify, hubspot, github, notion
```

Wave 1 destinations:

```text
postgres, mysql, mariadb, sqlite, cockroachdb
```

The Slack modal captures:

- pipeline name
- source connection via Slack external select
- selected streams
- per-stream `FULL_TABLE` or `INCREMENTAL` replication method
- replication keys for incremental streams
- one or more destination connections
- destination schema and stream-to-table mapping
- write mode: `append`, `upsert`, `replace`
- normalisation rules: `rename` and `cast`
- dbt UI SQL models
- schedule: `none`, `minutes`, `hourly`, `daily`, `weekly`, `monthly`, `custom_cron`
- optional run now

Slack-created pipelines use the same web-builder `pipelineGraph` contract:

- source node with `connection_id`, `connector_type`, `connection_name`, and
  strict `selected_streams`
- destination node with `connection_id`, `connector_type`, `emit_method`,
  `replication_method`, `replication_key`, `dest_schema`,
  `normalisation_rules`, `dbt_config`, `delivery_table_map`, schedule fields,
  and `execution_path: "duckdb_staged"`
- direct source-to-destination edges and branch metadata
- no transform node and no legacy ETL graph keys

Strict ELT invariants still apply: no runner-created destination tables, no
credential leakage, strict selected streams, dbt output-to-destination mapping,
and no legacy ETL paths.

## 16. Expected Behavior

After OAuth succeeds:

- `slack_integrations` stores the bot token and incoming webhook URL encrypted.
- The OAuth-selected Slack channel is enabled for alerts and slash commands.
- Users in that channel can run commands without manual mapping.
- `/mantrixflow link` remains optional for personal App Home and DM workflows.
- Commands outside the selected channel receive a helpful message pointing back
  to the connected channel.
- Alerts post to the OAuth-selected incoming webhook channel.
- Slack action failures do not fail pipeline execution.
- Slack lifecycle alerts are best-effort and deduped by run metadata markers.

## 17. Marketplace Review Checklist

Before submission:

- Direct Install URL is configured and returns a Slack OAuth `302`.
- OAuth callback is configured.
- Required scopes are limited to:
  - `commands`
  - `chat:write`
  - `incoming-webhook`
  - `im:history`
  - `im:write`
- Optional email-sync scopes are explained separately.
- App Home is enabled.
- Messages tab is enabled and writable.
- Events include `app_home_opened`, `message.im`, and `app_uninstalled`.
- Test both install paths:
  - Add to Slack from MantrixFlow Settings -> Integrations.
  - Direct install URL, then bind to an org.
- Public privacy policy URL is live.
- Public support URL or contact page is live.
- Screenshots show real Slack alerts and command responses, never credentials.
- Demo video covers install, App Home, Messages tab, self-link, pipeline
  list/status, run confirm, connection test, pipeline builder, and test alert.

Security checks:

- Go verifies Slack signatures with `SLACK_SIGNING_SECRET`.
- Verification uses the raw body and rejects stale timestamps.
- Token and webhook values are encrypted before persistence.
- API responses expose presence booleans, not secrets.
- Slash-command responses are ephemeral by default.
- Connection-test output is sanitized.
- Slack modal responses never echo secret values.
- `app_uninstalled` clears encrypted Slack secrets.
- Direct client access to Slack secret-bearing tables remains closed in
  `supabase_rls.sql`.

## 18. Troubleshooting

### `SLACK_UNAVAILABLE: Slack OAuth is not configured`

Go did not load `SLACK_CLIENT_ID` or `SLACK_CLIENT_SECRET`.

Fix:

1. Add both env vars to the Go-loaded env file.
2. Restart `go run ./cmd/server`.
3. Re-open Settings -> Integrations and click Add to Slack again.

### Slack says the redirect URL is invalid

The Redirect URL in Slack Dashboard must exactly match:

```text
https://b7c9-2409-40c1-5011-3ea3-2836-61f9-1495-f47c.ngrok-free.app/api/slack/oauth/callback
```

Protocol, host, and path must match. A new ngrok URL means updating Slack
Dashboard and `SLACK_OAUTH_REDIRECT_BASE_URL`, then restarting both Go and Next.

If Slack shows `Passed URI: https://YOUR_APP_HOST/api/slack/oauth/callback`,
the Go server is still running with a placeholder env value.

### Event Subscriptions says `challenge_failed`

Check:

1. Request URL is exactly:

```text
https://b7c9-2409-40c1-5011-3ea3-2836-61f9-1495-f47c.ngrok-free.app/api/slack/events
```

2. `SLACK_SIGNING_SECRET` matches Slack Dashboard -> Basic Information ->
   Signing Secret.
3. Both dev servers were restarted after env changes.

If ngrok shows `POST /api/slack/events -> 307 Temporary Redirect` followed by
`POST /auth/login`, the Next auth proxy is intercepting Slack. `/api/slack/*`
must remain public and bypass session middleware.

Quick local check:

```bash
curl -i \
  -X POST http://localhost:3000/api/slack/events \
  -H 'Content-Type: application/json' \
  --data '{"type":"url_verification","challenge":"local-ok"}'
```

Expected response:

```text
HTTP/1.1 200 OK
content-type: text/plain; charset=utf-8

local-ok
```

### Slack redirects to login after Allow

This usually means the install was started from one browser host but the OAuth
callback returned to another. Browser cookies are host-scoped, so the ngrok host
does not share your `localhost:3000` login.

Fix:

1. Set `APP_WEB_URL=http://localhost:3000`.
2. Set `SLACK_OAUTH_REDIRECT_BASE_URL` to the ngrok base host.
3. Include localhost and the ngrok host in `CORS_ALLOWED_ORIGINS`.
4. Restart Go and Next.
5. Start a fresh install from localhost Settings -> Integrations.

Owner-started OAuth stores the browser origin that created the install state, so
a localhost install returns to localhost settings even though Slack calls the
ngrok callback.

Do not reuse an old Slack approval URL. `Slack install link was already used`
means the one-time OAuth state was already consumed.

### Slash commands do not reach local dev

Slack cannot reach `localhost`. Use `ngrok http 3000` and set all Slack
Dashboard URLs to the ngrok HTTPS host.

### Commands say to use another channel

Run commands from the channel chosen during Slack OAuth. If you want a different
channel, reconnect Slack and choose the new channel on the Slack approval screen.

### /connection test shows channel access error

If `/connection test` returns "Use MantrixFlow from #channel-name" or asks you to
run `/mantrixflow link`, you need either:

1. Run `/mantrixflow link` to create a personal Slack user connection (allows
   personal App Home and DM workflows)
2. Use the command from the connected Slack channel (everyone in that channel
   can use MantrixFlow commands without linking)

### Personal App Home or DM asks for linking

Run `/mantrixflow link` only if you want personal App Home or DM workflows.
Channel slash commands do not require it.

## 19. Workspace-Only Private App Setup

If MantrixFlow is not published to the Slack App Directory (private/workspace-only),
you can still install it for your organization:

### Option 1: Direct Install from Slack Dashboard

1. Go to `https://api.slack.com/apps` and select your MantrixFlow app
2. Click "Install to Workspace" on the left sidebar
3. Authorize the app for your workspace
4. After install, go to MantrixFlow web app -> Settings -> Integrations
5. Connect the Slack workspace to your MantrixFlow organization

### Option 2: OAuth Install from MantrixFlow

1. In MantrixFlow, go to Settings -> Integrations
2. Click "Connect Slack"
3. You will be redirected to Slack for authorization
4. After authorization, Slack is connected

### For Workspace-Only Apps: The Connected Channel

When using a private/workspace-only app, commands work from the channel selected
during OAuth install. Everyone in that channel can use MantrixFlow commands
without running `/mantrixflow link`.

If you need to use commands from outside the connected channel:
1. Run `/mantrixflow link` to create a personal user connection
2. Then you can use commands from any channel or DM

### Connection Not Found Error

If `/connection test` returns "Connection not found":

1. First run `/connection list` to see available connections
2. Use the exact connection name or ID from the list
3. Example: `/connection test my-postgres-db`
4. Or use the full UUID: `/connection test 0a387af5-89e8-457a-a8a3-9a8dfb8e1b4b`

If no connections exist, create one:
1. Run `/connection create` to open the connection setup modal
2. Or go to MantrixFlow web app -> Connections -> New Connection

### First Time Setup Checklist

- [ ] Install MantrixFlow app from `https://api.slack.com/apps`
- [ ] Connect Slack from MantrixFlow Settings -> Integrations
- [ ] Create a connection in MantrixFlow (web app or `/connection create`)
- [ ] Run `/connection list` to verify connections exist
- [ ] Run `/connection test <name>` to test the connection
- [ ] Run `/pipeline list` to see available pipelines

## 20. Final Verification

Backend:

```bash
cd apps/server/main-server
GOCACHE=$(pwd)/.gocache-test go test ./internal/server/... ./internal/database/...
```

Frontend:

```bash
cd apps/app
bun run biome check
bun run build
```

Builder smoke test when available:

```bash
cd apps/app
bun run test:playwright:builder
```
