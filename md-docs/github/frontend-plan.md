# Frontend Plan

## API Layer

Added:
- `lib/api/types/github.ts`
- `lib/api/services/github.service.ts`
- `lib/api/hooks/use-github.ts`

The hooks expose integration state, repository listing, pipeline Git config, export, push, pull, history, and rollback mutations.

## Settings Integrations

Settings now treats GitHub as a live integration:
- Owner/Admin can start the GitHub App install flow.
- Connected state shows account, repository count, installation ID, and install time.
- Drawer lists repositories available to the installation.
- Manage GitHub App opens the GitHub installation page.
- Disconnect marks the integration disconnected and removes cached token data server-side.

## Pipeline Settings Drawer

The drawer has `General` and `GitHub` tabs.

GitHub tab:
- repository selector from `GET /github/repos`
- YAML file path input
- branch input
- sync mode selector
- save config
- export YAML
- push to GitHub
- pull latest from GitHub
- PR link/status if bidirectional mode creates a PR

## Builder Action Bar

The builder action bar shows persisted Git status without remote API calls:
- No Git
- Synced
- Syncing
- Sync failed
- Conflict
- PR open

A History button opens the Git history drawer when GitHub config exists.

## Git History Drawer

The drawer lists commits returned by `GET /git-history`:
- message
- short SHA
- author
- relative time
- GitHub diff link
- current marker
- rollback button with confirmation

Rollback calls `POST /rollback`.

## Pipeline List

The list has a Git status column based only on persisted pipeline fields. It intentionally avoids per-row GitHub API calls.
