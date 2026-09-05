# Playwright complete workflow E2E summary

## Scope

This document tracks the fixture-driven, UI-only Playwright suite in `apps/arcyria-platform/__tests__/playwright/workflow`. It replaces API-assisted setup in the new suite with visible connection, Explorer, pipeline, transformation, destination, run, lifecycle, and cleanup actions.

## Test environment

| Component | Local URL | Command |
| --- | --- | --- |
| Next.js frontend | `http://localhost:3000` | `bun run dev` from `apps/arcyria-platform` |
| Go main API | `http://localhost:5000` | `go run ./cmd/server` from `apps/server/arcyria-server` |
| Python ELT runner | `http://localhost:8000` | `./.venv/bin/python -m uvicorn api.main:app --host 0.0.0.0 --port 8000 --loop asyncio` from `apps/server/arcyria-elt` |
| Browser | Playwright Desktop Chrome / Chromium | `bun run test:e2e:headed` for visible local execution |

Playwright can start all three services with `E2E_START_SERVERS=1`. Failure screenshots, retained videos, traces, logs, HTML, and CI JUnit results are under `apps/arcyria-platform/.playwright`.

## Connections and pipelines

| Flow | Source | Destination | Stream mode | Destination strategy |
| --- | --- | --- | --- | --- |
| PostgreSQL full table | generated PostgreSQL source connection | generated PostgreSQL destination connection | full table | transformed mixed-data-type table |
| PostgreSQL incremental | generated PostgreSQL source connection | generated PostgreSQL destination connection | incremental on `updated_at` | upsert by `id` |
| HubSpot | generated HubSpot source connection | generated PostgreSQL destination connection | full table | identity transformation and table per displayed stream |
| Stripe | generated Stripe source connection | generated PostgreSQL destination connection | full table | identity transformation and table per displayed stream |

All connection, table, schema, transformation, and pipeline names include a unique run ID.

## PostgreSQL data types

JSON fixtures cover integer, bigint, numeric, decimal, varchar, text, boolean, date, timestamp, timestamptz, UUID, JSON, JSONB, text arrays, integer arrays, empty arrays, and nullable values. Source rows come from `postgres-records.json`; expected transformed subsets and row counts come from `expected-results.json`.

## Transformations and versions

- SQL is entered through the accessible plain-text SQL field synchronized with Monaco.
- A draft must save, validate, preview, and publish before pipeline validation.
- PostgreSQL full-table flow runs revision 1, compares destination rows with JSON, publishes revision 2 with equivalent updated SQL, reruns, and compares again.
- Revision history must expose both revisions and identify revision 2 as published.
- Invalid SQL must display validation failure and keep publication disabled.

## YAML and GitHub

The suite verifies that the relationship YAML preview renders for a configured pipeline and survives navigation/refresh. Repository push/pull is intentionally not performed by this suite because it is an external side effect and requires a dedicated repository selected by the test environment. GitHub CI runs the suite itself and stores Playwright evidence; repository synchronization can be added as an opt-in fixture when a disposable repository is supplied.

## Test matrix

| Test Case | Source | Destination | Expected Result | Actual Result | Status | Issue | Fix Applied |
| --------- | ------ | ----------- | --------------- | ------------- | ------ | ----- | ----------- |
| Required connection fields | PostgreSQL form | N/A | Required controls visible; save disabled | Implemented assertion | Implemented | None | N/A |
| Invalid credentials | Invalid PostgreSQL | N/A | Test fails; error visible; save disabled | Implemented assertion | Implemented | None | N/A |
| Duplicate connection name | PostgreSQL | N/A | Duplicate rejected with validation message | Implemented assertion | Implemented | None | N/A |
| Create/test PostgreSQL source | PostgreSQL | N/A | Connection saves and retests successfully | UI page object implemented | Implemented | None | N/A |
| Create/test PostgreSQL destination | N/A | PostgreSQL | Connection saves and retests successfully | UI page object implemented | Implemented | None | N/A |
| Create/test HubSpot source | HubSpot | N/A | Connection saves and retests successfully | UI page object implemented | Implemented | None | N/A |
| Create/test Stripe source | Stripe | N/A | Connection saves and retests successfully | UI page object implemented | Implemented | None | N/A |
| JSON source setup | PostgreSQL | N/A | Generated schema, two tables, fixture rows | UI-controlled E2E SQL panel implemented | Implemented | Explorer was read-only | Added non-production, server-restricted setup panel |
| JSON destination setup | N/A | PostgreSQL | Generated empty destination schema | UI-controlled E2E SQL panel implemented | Implemented | Explorer was read-only | Added non-production, server-restricted setup panel |
| PostgreSQL data types | PostgreSQL | PostgreSQL | All fixture columns discovered and transferred | Assertions implemented | Pending live run | Requires live databases | N/A |
| PostgreSQL full table | PostgreSQL | PostgreSQL | Validated run completes and rows match JSON | Assertions implemented | Pending live run | Requires live databases | N/A |
| Transformation version 1/2 | PostgreSQL | PostgreSQL | Both versions publish and produce expected rows | Assertions implemented | Pending live run | Requires live runner | N/A |
| Incremental append | PostgreSQL | PostgreSQL | Second run adds only new ID and final count is 3 | Assertions implemented | Pending live run | Requires live runner | N/A |
| HubSpot all streams | HubSpot | PostgreSQL | Every displayed stream selected, transformed, materialized, run completes | Dynamic loop implemented | Pending live run | Requires token scopes and source data | N/A |
| Stripe all streams | Stripe | PostgreSQL | Every displayed stream selected, transformed, materialized, run completes | Dynamic loop implemented | Pending live run | Requires key scopes and source data | N/A |
| Missing destination table | PostgreSQL | PostgreSQL | Run fails with details; explicit table creation fixes it | Failure/retry assertions implemented | Pending live run | Requires live runner | N/A |
| Incompatible destination types | PostgreSQL | PostgreSQL | Run fails; target correction and retry complete | Failure/retry assertions implemented | Pending live run | Requires live runner | N/A |
| Invalid SQL | PostgreSQL | PostgreSQL | Validation fails and publish remains disabled | Assertions implemented | Pending live run | Requires live validator | N/A |
| Edit/pause/resume/rerun | PostgreSQL | PostgreSQL | State persists and later run completes | Assertions implemented | Pending live run | Requires live runner | N/A |
| List/open/refresh | PostgreSQL | PostgreSQL | Edited pipeline lists, opens, and survives refresh | Assertions implemented | Pending live run | Requires live services | N/A |
| Relationship YAML | PostgreSQL | PostgreSQL | YAML preview is available after configuration | Assertion implemented | Pending live run | External repo push/pull not in scope | N/A |
| Cleanup | All | PostgreSQL | Only run-owned pipelines, schemas, tables, rows, and connections removed | Reverse-order teardown implemented | Pending live run | Requires delete permissions | N/A |

## Implementation verification

- Playwright discovers the setup project and eight workflow tests.
- The new suite compiles cleanly in the application TypeScript pass; unrelated pre-existing Vitest type errors remain outside this suite.
- `playwright.config.ts` retains traces, screenshots, and videos on failure and emits HTML/JUnit reports.
- `.github/workflows/playwright-e2e.yml` installs services and browser dependencies, starts the stack, runs the complete workflow, and uploads evidence.

## Remaining work before a green environment result

Run `bun run test:e2e` against the configured live source/destination databases and SaaS accounts. Any live connector permission limitations or backend defects must be recorded here with the failed trace, fix, and retest result. Do not change a pending row to passed without a successful Playwright execution for that scenario.
