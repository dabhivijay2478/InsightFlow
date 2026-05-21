# Test Plan

## Go Unit Tests

- YAML export/import round trip for:
  - multiple selected streams
  - multiple destination nodes
  - destination-owned `dbt_config.sql_models[]`
- Import fails when a connection name is duplicated in the org.
- Exported YAML contains no credential-looking fields:
  - password
  - token
  - secret
  - credential
  - private key
  - SSH material
- Webhook HMAC validation accepts valid signatures and rejects invalid signatures.
- GitHub client request construction with `httptest.Server`:
  - auth header
  - API version header
  - content encoding
  - refs and PR payloads

## Go Route Tests

- Start integration creates hashed install state.
- Callback consumes state and upserts `org_github_integrations`.
- Enable/disable pipeline GitHub config.
- Export returns valid YAML.
- Pull rejects invalid YAML and marks `sync_failed`.
- Push always creates a new `mantrixflow/pipelines/...` branch and pull request.
- No push path commits directly to the configured/base branch.
- Rollback applies YAML and creates a review pull request.
- `pull_request.closed` merged clears PR URL.
- `pull_request.closed` unmerged marks `conflict`.

## Frontend Tests

- Settings GitHub drawer:
  - not connected state
  - connected account/repo state
  - connect button calls start endpoint
  - disconnect confirmation/mutation
- Pipeline GitHub tab:
  - validates repository/path before save
  - saves selected sync mode
  - export/push/pull buttons call the right mutations
- History drawer:
  - renders commit list
  - disables rollback on current commit
  - confirms before rollback
- Pipeline list:
  - No Git, Synced, PR open, Conflict, Sync failed statuses render from persisted fields.

## Manual Verification

1. Configure `GITHUB_APP_ID`, `GITHUB_APP_SLUG`, `GITHUB_APP_PRIVATE_KEY`, and `GITHUB_WEBHOOK_SECRET`.
2. Connect the GitHub App from Settings.
3. Enable GitHub sync for a pipeline and push YAML.
4. Edit YAML in GitHub, merge to configured branch, and verify `pipeline_graph` updates.
5. Save a builder change in bidirectional mode and verify a PR is created.
6. Close the PR without merging and verify conflict banner/status.
7. Use history rollback and verify a new pull request is created.
