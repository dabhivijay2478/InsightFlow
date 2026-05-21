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
- Connected state shows the installed account and repository count.
- Drawer lists repositories available to the installation.
- Manage GitHub App opens the GitHub installation page.
- Disconnect uninstalls the GitHub App installation from GitHub, then removes cached token data server-side.

## Pipeline Settings Drawer

The drawer has `General` and `GitHub` tabs.

GitHub tab:
- repository selector from `GET /github/repos`
- YAML file path input
- branch input
- sync mode selector
- save config
- export YAML
- create GitHub PR
- pull latest from GitHub
- PR link/status for the generated review branch

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
