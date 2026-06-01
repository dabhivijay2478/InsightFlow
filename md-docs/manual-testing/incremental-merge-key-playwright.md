# Incremental merge-key Playwright E2E

Last updated: 2026-06-01

End-to-end test for **incremental sync + merge on `id`** using real Postgres
connections. Seeds fake data in isolated schemas, runs the pipeline twice, and
asserts the destination ends with **1,500 rows** (1,000 baseline + 500 delta)
with **no duplicate primary keys**.

## Default connections

| Role | UUID |
| --- | --- |
| Source | `0a387af5-89e8-457a-a8a3-9a8dfb8e1b4b` |
| Destination | `efc546a0-caac-47e9-b1a3-e2c13dd4526d` |

Override with `INCR_MERGE_SOURCE_ID` / `INCR_MERGE_DESTINATION_ID`.

## Prerequisites

1. Go API (port **5000**), ELT server, and Next.js app running locally.
2. `apps/app/.env.local` with `TEST_ORGANIZATION_ID`, Supabase keys, `NEXT_PUBLIC_API_URL=http://localhost:5000` (or your public Go API URL).
3. `__tests__/fixtures/credentials/login.json` for a user in that org.
4. Main server env: `ALLOW_SOURCE_DB_MUTATIONS=true` (E2E SQL seed route).

### ELT callbacks (required)

Go dispatch sends ELT these URLs built from **`API_PUBLIC_URL`** on the main server:

- `GET {API_PUBLIC_URL}/api/v1/internal/checkpoint/...`
- `POST {API_PUBLIC_URL}/api/v1/internal/elt-callback`

They must hit the **Go API**, not the Next.js app.

| Symptom | Cause |
| --- | --- |
| Callback `HTTP 404` with ngrok HTML | ngrok tunnels **port 3000** (Next) instead of **5000** (Go) |
| Run stuck `running` in UI but rows landed in destination | Delivery succeeded; callback never updated `pipeline_runs` |
| Run 2 fails: `Expired: pending run was not dispatched` | Run 1 still `running` blocked dispatch; fixed in Go by reconciling stale runs — cancel run 1 after run 1 in E2E if needed |
| Second run does not pick up delta | Checkpoint 404 — cursor not saved (re-sync behavior may still reach 1.5k with merge) |

**Local fix:**

```bash
# apps/server/main-server/.env (or export before go run)
API_PUBLIC_URL=http://127.0.0.1:5000
```

**ngrok fix:** tunnel the Go port, not Next:

```bash
ngrok http 5000
# Set API_PUBLIC_URL to the https://….ngrok-free.app URL
```

Restart the Go API after changing `API_PUBLIC_URL`.

## Run

```bash
cd apps/app
bun run test:playwright:incr-merge
```

### Tune row counts

```bash
INCR_MERGE_BASELINE_ROWS=1000 INCR_MERGE_DELTA_ROWS=500 bun run test:playwright:incr-merge
```

### Keep artifacts for debugging

```bash
INCR_MERGE_KEEP_SCHEMAS=1 INCR_MERGE_KEEP_PIPELINES=1 bun run test:playwright:incr-merge
```

## What the test does

| Step | Action |
| --- | --- |
| 01 | Create `mxf_e2e_p2p_source_incrmerge_*` / `mxf_e2e_p2p_dest_incrmerge_*_d1` schemas; seed **1,000** source rows; create empty destination table with PK on `id` |
| 02 | Build pipeline (INCREMENTAL, `updated_at`, merge + `merge_key_columns: [id]`); run 1 → destination **1,000** |
| 03 | Insert **500** new source rows (`updated_at` in 2030, after baseline) |
| 04 | Run 2 → destination **1,500**, `COUNT(DISTINCT id) = 1500`, no duplicate `id` (re-seeds delta if step 03 was skipped) |
| 05 | Run 3 with no new source changes → count stays **1,500** (merge idempotency) |

The test can treat a run as complete when the **destination row count** is stable, even if the API status is still `running` (lost callback). Fix `API_PUBLIC_URL` for accurate run status in the product UI.

## Files

- Spec: `apps/app/__tests__/playwright/ui/incremental-merge-key-grow.spec.ts`
- SQL/graph helpers: `apps/app/__tests__/helpers/incremental-merge-e2e.ts`
- Report JSON: `md-docs/test-logs/incremental-merge-key-*.json`

## Expected failure modes

- Destination table missing PK → pre-flight fails before extract.
- Merge key `id` omitted from SQL model → pre-flight or delivery names the column.
- Merge disabled (append) with re-synced rows → row count grows past 1,500 on run 3.
- `API_PUBLIC_URL` points at Next.js → permanent callback 404 in ELT logs.
