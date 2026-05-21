# Backend Plan

## Data Model

Use `public.pipelines`, not the older `etl_pipelines` name.

Pipeline fields:
- `github_repo_owner`, `github_repo_name`, `github_repo_url`
- `github_file_path`, `github_branch`
- `github_sync_mode`: `manual`, `github_to_db`, `bidirectional`
- `github_sync_status`: `idle`, `syncing`, `sync_failed`, `conflict`
- `github_last_synced_sha`, `github_last_synced_at`
- `github_sync_error`, `github_pr_url`

Org integration tables:
- `org_github_integrations` stores GitHub App installation metadata, account metadata, encrypted token cache, token expiry, connected/disconnected timestamps.
- `github_install_states` stores a hash of the install state, organization, user, return path, expiry, and consumed time.

Secret handling uses the existing Go AES-GCM helper in `internal/crypto/encryption.go`.

## Routes

Public:
- `GET /api/v1/github/callback`
- `POST /api/v1/github/webhook`

Organization scoped:
- `POST /api/v1/organizations/:organizationId/github/integration/start`
- `GET /api/v1/organizations/:organizationId/github/integration`
- `DELETE /api/v1/organizations/:organizationId/github/integration`
- `GET /api/v1/organizations/:organizationId/github/repos`
- `GET /api/v1/organizations/:organizationId/github/repos/:owner/:repo/files`

Pipeline scoped:
- `GET /api/v1/organizations/:organizationId/pipelines/:id/github`
- `POST /api/v1/organizations/:organizationId/pipelines/:id/github`
- `DELETE /api/v1/organizations/:organizationId/pipelines/:id/github`
- `POST /api/v1/organizations/:organizationId/pipelines/:id/export`
- `POST /api/v1/organizations/:organizationId/pipelines/:id/push`
- `POST /api/v1/organizations/:organizationId/pipelines/:id/pull`
- `GET /api/v1/organizations/:organizationId/pipelines/:id/git-history`
- `POST /api/v1/organizations/:organizationId/pipelines/:id/rollback`

## GitHub App Flow

1. `integration/start` creates a random state and stores only its SHA-256 hash.
2. The API returns `https://github.com/apps/{GITHUB_APP_SLUG}/installations/new?state=...`.
3. Callback validates state and `installation_id`.
4. API creates a GitHub App JWT using `GITHUB_APP_ID` and `GITHUB_APP_PRIVATE_KEY`.
5. API fetches installation metadata and creates an installation token.
6. Token is encrypted and cached until `token_expires_at`.
7. Future GitHub calls refresh the token on demand when near expiry.

## GitHub Client

Raw HTTP with:
- `Accept: application/vnd.github+json`
- `X-GitHub-Api-Version: 2026-03-10`

Implemented operations:
- create installation token
- get installation metadata
- list installation repositories
- get file content at ref
- create/update file
- get/create refs
- create pull requests
- squash merge pull requests
- close/cancel pull requests
- list commits for file path
- list repository contents
- compare refs

## Webhooks

`X-Hub-Signature-256` is verified using HMAC-SHA256 over the raw body before parsing.

`push`:
- matches installation, repo owner/name, branch, configured path, and sync mode.
- handles only `github_to_db` and `bidirectional` pipelines.
- fetches YAML at the pushed commit SHA.
- validates and applies YAML to `pipeline_graph`.
- marks failed state and stores error if validation/import fails.

`pull_request.closed`:
- merged PR clears `github_pr_url` and returns status to `idle`.
- declined PR marks `conflict` and keeps DB runtime state.

The pipeline GitHub tab also exposes explicit app actions to squash-merge or cancel the currently stored PR URL.

## Rollback

Rollback fetches YAML at the requested commit SHA, validates it, applies it to DB, then writes a reviewed Git update:
- all MantrixFlow-originated GitHub writes create a new branch and pull request.
- never commit directly to the configured/base branch.
- branch names use `mantrixflow/pipelines/{pipeline-slug}-{pipeline-id-prefix}-{timestamp}`.
- generated YAML should live under `mantrixflow/pipelines/{pipeline-slug}.yaml` by default to avoid colliding with a user's existing `pipelines/` folder.
- user-entered file paths are normalized into `mantrixflow/pipelines/{file}.yaml` unless already under that namespace.

No force-pushes are allowed.
