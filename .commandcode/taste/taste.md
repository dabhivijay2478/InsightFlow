# Taste (Continuously Learned by [CommandCode][cmd])

[cmd]: https://commandcode.ai/

# agent-ui
- Integrate SSL/TLS mode selection as radio buttons directly within the credential form card, not as a separate step after credential submission. Confidence: 0.65

# bun
- Use `bun run` for all scripts (lint, build, biome check, dev) instead of npm/yarn equivalents. Confidence: 0.85
- Use `bunx` instead of `npx` for running CLI tools like shadcn. Confidence: 0.85

# go-testing
- Run Go tests with explicit GOCACHE: `GOCACHE=$(pwd)/.gocache-test go test ./internal/server/... ./internal/database/...`. Confidence: 0.80

# documentation
- Internal project docs live in `md-docs/` folder; public-facing docs site lives in `mantrixflow-docs/` folder. Confidence: 0.70
- Prefer consolidated single guide files over multiple separate markdown files for the same topic. Do not create separate files for AWS OIDC, product setup, etc. — one deployment file under infra. Confidence: 0.80
- Use only `support@mantrixflow.com` for website terms/privacy pages and docs site contact email. Never use `security@mantrixflow.com`. Confidence: 0.85
- Do not mention internal implementation details (DuckDB, Go API, Python ELT, dbt) in public-facing docs site or website. Use generic terms: "staging"/"temporary storage" instead of DuckDB, "SQL" instead of "dbt SQL", omit internal service layers entirely. Confidence: 0.80

# infra
- Production infrastructure lives in `apps/mantrixflow-infra` as its own standalone Git repository (CDK + Terraform). Not inside `apps/app`. Confidence: 0.85
- Infra repo deploys on merge to `main`; app/API/ELT repos use `mantrixflow` as production branch. Confidence: 0.75
- Use GitHub OIDC for AWS auth in CI/CD, not long-lived AWS access keys. Confidence: 0.80
- Use Cloudflare API token (not API key) for Terraform Cloudflare provider authentication. Confidence: 0.75
- Main server repo: `dabhivijay2478/cloud.api.mantrixflow.com`. ELT server repo: `dabhivijay2478/cloud.api.etl.server.mantrixflow.com`. Confidence: 0.75

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

# cdc
- Product CDC (LOG_BASED, replication slots) is removed from active product behavior. pg_replication package may remain on disk but must not be imported or surfaced by active routes. Confidence: 0.75

# builder-ux
- Source drawer has Preview tab (first) and Config tab (second). Config is source-only: connection status, discover, refresh tables. No sync mode or replication key. Confidence: 0.70
- Destination drawer owns: sync mode, manual replication key, normalisation rules (destination-scoped), dbt SQL models per selected source table, delivery schema, emit method. Confidence: 0.70
