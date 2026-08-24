# Taste (Continuously Learned by [CommandCode][cmd])

[cmd]: https://commandcode.ai/

# agent-ui
- Integrate SSL/TLS mode selection as radio buttons directly within the credential form card, not as a separate step after credential submission. Confidence: 0.65

# bun
- Use `bun run` for all scripts (lint, build, biome check, dev) instead of npm/yarn equivalents. Confidence: 0.85
- Use `bunx` instead of `npx` for running CLI tools like shadcn. Confidence: 0.85
- shadcn CLI reads registry tokens (like EFFERD_REGISTRY_TOKEN) from the app's .env file, not from system environment variables — always ensure tokens are in the target app's .env before running shadcn add commands. Confidence: 0.85

# go-testing
- Run Go tests with explicit GOCACHE: `GOCACHE=$(pwd)/.gocache-test go test ./internal/server/... ./internal/database/...`. Confidence: 0.80

# documentation
- Internal project docs live in `md-docs/` folder; public-facing docs site lives in `mantrixflow-docs/` folder. Confidence: 0.70
- Prefer consolidated single guide files over multiple separate markdown files for the same topic. Do not create separate files for AWS OIDC, product setup, etc. — one deployment file under infra. Confidence: 0.80
- Use only `support@mantrixflow.com` for website terms/privacy pages and docs site contact email. Never use `security@mantrixflow.com`. Confidence: 0.85
- Do not mention internal implementation details (DuckDB, Go API, Python ELT, dbt) in public-facing docs site or website. Use generic terms: "staging"/"temporary storage" instead of DuckDB, "SQL" instead of "dbt SQL", omit internal service layers entirely. Confidence: 0.80
- After adding a new connector (e.g., MySQL, Notion, Asana, Linear, Stripe, HubSpot), always update both the docs site and marketing website to reflect the new source/destination in the same change cycle. Confidence: 0.80

# infra
See [infra/taste.md](infra/taste.md)
# github-integration
- GitHub App env vars use `GH_` prefix (e.g., `GH_APP_SLUG`), never `GITHUB_` prefix — GitHub Actions rejects `GITHUB_` prefixed secrets. Confidence: 0.80
- Pipeline push to GitHub always creates a new branch and PR; never push directly to main. Branch naming: `mantrixflow/{pipeline-name}-{timestamp}`. Confidence: 0.80
- Pull pipeline YAML from `mantrixflow/pipelines/` folder in repo, not root, to avoid conflicts with user data. Confidence: 0.75

# supabase
- Use Supabase new JWT signing keys (not legacy JWT secret). `SUPABASE_JWT_SECRET` is deprecated and should be deleted. Confidence: 0.75

# builder-ux
- Builder tour/tooltips only appear when a pipeline already exists; do not show on empty new pipeline page. Confidence: 0.70
- Do not add tooltips for source/destination drawer panels. Confidence: 0.70

# ux
- Use `confirmation-modal.tsx` (existing shared component) for destructive/dangerous actions with toast messages; do not create alternative confirmation dialogs. Confidence: 0.75

# env
- Environment variables for each service belong in `.env` files inside their respective server directories (app, main-server, elt-server). Confidence: 0.70

# architecture
See [architecture/taste.md](architecture/taste.md)
# debugging
- When fixing bugs, always find and fix the root cause. Do not apply patch fixes or workarounds. Confidence: 0.90
- When user reports a connector missing from the UI catalog (e.g., "i not see asana name"), first check if the issue is stale local services (ELT/Go) and restart them before assuming code-side gating. Confidence: 0.75
- When verifying pipeline fixes, restart services in this exact order: ELT server first (port 8000), then Go main server (port 5000), then frontend app (port 3000). Each must be healthy before the next starts. Confidence: 0.85
- For real data-movement verification, run end-to-end tests one pipeline at a time rather than in parallel — MySQL and other SQL connectors do not tolerate concurrent connections against the same source/destination reliably. Confidence: 0.70

# connector-rollout
- New connectors (Linear, Notion, Asana, Stripe, HubSpot, Airtable, MySQL) must be runtime health-gated: the frontend `/api/v1/connectors/health` endpoint reports per-connector source/destination availability, and the UI hides the connector card when the capability is unavailable. Confidence: 0.85
- When a connector is added but not showing in the UI, check that the runtime enablement allowlist (e.g., `ENABLED_CONNECTOR_IDS`) includes the connector — having it in the registry is not enough; the gate is explicit. Confidence: 0.80
- Frontend connector catalog entries must declare `availability: "runtime"` (not omit the field) when they need ELT-side health confirmation. An omitted `availability` field renders the connector even when the backend can't support it. The matching ELT `/health` endpoint must publish a `clickhouse` (or other connector) entry with `{source, destination, available, reason}` so the frontend gate has something to read. Confidence: 0.85

# browser-verification
- For connector and pipeline verification, use the user's existing signed-in Chrome session (Profile 1) via the Chrome-control skill rather than spinning up a fresh logged-out session — saved connections and organizations live in that session. Confidence: 0.90
- When running real-data pipeline tests, the agent must use the live UI in the user's Chrome (click through Connections → New Connection → Test → Save → Pipeline → Run), not only API-level checks. Mocked or skipped movement does not count as production verification. Confidence: 0.85
- After any frontend code change in the local app, reload the tab (`tab.reload()`) before re-verifying — hot reload is not assumed reliable. Confidence: 0.75

# cdc
- Product CDC (LOG_BASED, replication slots) is removed from active product behavior. pg_replication package may remain on disk but must not be imported or surfaced by active routes. Confidence: 0.75

# builder-ux
- Source drawer has Preview tab (first) and Config tab (second). Config is source-only: connection status, discover, refresh tables. No sync mode or replication key. Confidence: 0.70
- Destination drawer owns: sync mode, manual replication key, normalisation rules (destination-scoped), dbt SQL models per selected source table, delivery schema, emit method. Confidence: 0.70
