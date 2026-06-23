# GitHub App Setup

This guide configures the GitHub App required for MantrixFlow pipeline config version control.

## 1. Create The GitHub App

In GitHub, create a new GitHub App for the MantrixFlow environment.

Recommended naming:
- Production: `MantrixFlow`
- Staging: `MantrixFlow Staging`
- Local/dev: `MantrixFlow Dev`

Set the **Setup URL**. This is the important redirect field for GitHub App installs:

```text
https://api.yourdomain.com/api/v1/github/callback
```

For local development, use the tunnel/public API URL:

```text
https://your-ngrok-host.ngrok-free.app/api/v1/github/callback
```

The ngrok tunnel must forward to the Go main server, usually `http://localhost:5000`, because GitHub calls backend callback and webhook routes. Keep `APP_WEB_URL` pointed at the frontend, usually `http://localhost:3000`, so the backend can redirect the browser back into MantrixFlow after the install completes.

Do not rely on the OAuth callback URL for this install flow. GitHub redirects users to the Setup URL after they install a GitHub App; the callback URL is used for GitHub App user authorization/OAuth flows.

If the user stays on GitHub after clicking Install, the Setup URL is missing or points to the wrong host.

For local reinstall/update testing, enable **Redirect on update**. Without it, GitHub shows "app was updated" and stays on GitHub after repository access changes.

## 2. Configure Permissions

Repository permissions:
- Contents: Read and write
- Pull requests: Read and write
- Metadata: Read-only

Subscribe to events:
- Installation
- Installation repositories
- Push
- Pull request

Webhook URL:

```text
https://api.yourdomain.com/api/v1/github/webhook
```

Webhook secret:
- Generate a random secret with `openssl rand -hex 32`.
- Store the same value in `GH_WEBHOOK_SECRET`.

## 3. Generate Private Key

From the GitHub App settings page, generate a private key.

Store it as `GH_APP_PRIVATE_KEY`.

Example `.env` shape:

```dotenv
GH_APP_ID=123456
GH_APP_SLUG=mantrixflow-dev
GH_APP_PRIVATE_KEY=-----BEGIN RSA PRIVATE KEY-----\n...\n-----END RSA PRIVATE KEY-----
GH_WEBHOOK_SECRET=generated-webhook-secret
GITHUB_API_BASE_URL=https://api.github.com
```

## 4. Configure API And App URLs

The API must know its public callback base and the browser app return URL:

```dotenv
API_PUBLIC_URL=https://api.yourdomain.com
APP_WEB_URL=https://app.yourdomain.com
```

For local testing with a tunnel:

```dotenv
API_PUBLIC_URL=https://your-ngrok-host.ngrok-free.app
APP_WEB_URL=http://localhost:3000
```

## 5. Install From MantrixFlow

1. Open MantrixFlow.
2. Go to Settings -> Integrations -> GitHub.
3. Click Connect GitHub.
4. Install the GitHub App on the account/org and selected repositories.
5. Return to MantrixFlow and confirm the connected account/repository count.

The install flow stores:
- GitHub installation ID
- account metadata
- encrypted installation token cache
- token expiry

Tokens expire quickly and are refreshed on demand.

## 6. Enable Pipeline Sync

For each pipeline:

1. Open Pipeline Builder.
2. Open Settings -> GitHub.
3. Select repository.
4. Set file path, usually `mantrixflow/pipelines/{pipeline-name}.yaml`.
5. Set branch, usually `main`.
6. Choose sync mode:
   - Manual export
   - Auto-sync from GitHub
   - Bidirectional PR workflow
7. Save GitHub config.
8. Create a GitHub PR or export the first YAML version.

## 7. Webhook Verification

To verify webhooks:

1. Edit and merge a pipeline YAML file in GitHub.
2. Confirm GitHub sends a `push` event to `/api/v1/github/webhook`.
3. Confirm the pipeline updates in MantrixFlow.
4. Confirm `github_last_synced_sha`, `github_last_synced_at`, and `github_sync_status` update on `public.pipelines`.

If a webhook fails validation, the API returns `401` and does not parse the payload.

## 8. Troubleshooting

Common checks:
- `GH_APP_ID` matches the app ID, not the installation ID.
- `GH_APP_SLUG` matches the URL slug under `https://github.com/apps/{slug}`.
- `GH_APP_PRIVATE_KEY` has valid PEM content.
- `GH_WEBHOOK_SECRET` matches GitHub App webhook settings exactly.
- GitHub App **Setup URL** is set to `{API_PUBLIC_URL}/api/v1/github/callback`.
- GitHub App **Redirect on update** is checked, especially when reinstalling or changing repositories in local dev.
- GitHub App **Webhook URL** is set to `{API_PUBLIC_URL}/api/v1/github/webhook`.
- `APP_WEB_URL` points to the browser app origin.
- `API_PUBLIC_URL` points to the public Go API origin.
- The GitHub App is installed on the target repository.
- Connection names in YAML are unique inside the organization.

Useful tables:

```sql
select organization_id, github_installation_id, github_account_login, disconnected_at
from org_github_integrations
order by updated_at desc;

select id, name, github_repo_owner, github_repo_name, github_file_path,
       github_branch, github_sync_mode, github_sync_status,
       github_last_synced_sha, github_sync_error
from pipelines
where github_file_path is not null;
```
