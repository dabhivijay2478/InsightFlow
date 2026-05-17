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
- Prefer consolidated single guide files over multiple separate markdown files for the same topic. Confidence: 0.65

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
