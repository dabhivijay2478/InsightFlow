# GitHub Version Control For Pipeline Configs

This package documents the MantrixFlow GitHub version-control implementation for pipeline configs.

Runtime execution still reads `public.pipelines.pipeline_graph`. GitHub stores the reviewed, versioned YAML source, and validated GitHub changes sync back into the DB graph.

## Files

- [backend-plan.md](backend-plan.md) - Go API, DB models, GitHub App flow, webhooks, rollback.
- [yaml-contract.md](yaml-contract.md) - graph-aligned YAML shape and import/export invariants.
- [frontend-plan.md](frontend-plan.md) - Settings, pipeline builder, history drawer, list status.
- [test-plan.md](test-plan.md) - unit, route, frontend, and manual verification.

## Implemented Surface

- `pipelines` GitHub metadata columns: repository, path, branch, sync mode/status, last SHA/time, error, PR URL.
- `org_github_integrations`: org-level GitHub App installation metadata and AES-GCM encrypted token cache.
- `github_install_states`: signed, short-lived install state for callback validation.
- Raw GitHub REST client with App JWT, installation token refresh, contents, refs, PRs, commits, and repo listing.
- YAML export/import preserving strict ELT graph contracts.
- Webhook signature verification for `push` and `pull_request.closed`.
- Settings integration drawer, pipeline GitHub settings tab, Git history drawer, and pipeline list Git status.

## Invariants

- No credentials, connection IDs, or token values are serialized into YAML.
- Imported YAML resolves connection names inside the organization; duplicate names fail clearly.
- SQL models remain destination-node-owned under `dbt_config.sql_models[]`.
- Source stream keys use `schema.table`; DuckDB staging uses `schema__table`.
- Rollback writes a normal commit or bidirectional PR. It never force-pushes.
