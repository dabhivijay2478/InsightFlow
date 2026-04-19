# ELT Architecture Addendum — Strict ELT migration (post-audit)

**Scope:** this addendum supersedes the pipeline section of the knowledge
graph in [`graph.html`](./graph.html) / [`graph.json`](./graph.json) and the
runtime section of [`GRAPH_REPORT.md`](./GRAPH_REPORT.md). The graph itself
has NOT been regenerated — only this narrative reflects the post-migration
strict-ELT flow.

For the active enforcement rules and the human guide, see:

- [`md-docs/strict-elt-pipeline-guide.md`](../md-docs/strict-elt-pipeline-guide.md)
  — step-by-step guide (schema.table contract, 5 phases, UI flow, invariants).
- [`.cursor/rules/strict-elt-invariants.mdc`](../.cursor/rules/strict-elt-invariants.mdc)
  — the 12 invariants, always-applied.
- [`.cursor/rules/elt-flow-diagram.mdc`](../.cursor/rules/elt-flow-diagram.mdc)
  — the 5-phase reference diagram.
- [`.cursor/rules/agents-and-orchestration.mdc`](../.cursor/rules/agents-and-orchestration.mdc)
  — agent onboarding + specialized review personas (orchestrator, code, architecture, compliance, docs, ELT invariants).

## 1. Active services (authoritative paths)

| Concern | Path | Notes |
|---|---|---|
| Frontend (Next.js 16 + React 19) | `apps/app` | pipeline builder + run drawers |
| Go API (Fiber) | `apps/server/main-server` | orchestration, credentials, pgmq dispatch, callback |
| Python ELT (FastAPI + dlt + dbt-duckdb) | `apps/server/elt-server` | 5-phase staged runner |

The graph may still reference `apps/api` / `apps/new-etl`. Treat those as
historic artefacts — the active code lives under the paths above.

## 2. The single execution path

There is exactly **one** execution path for every run: `duckdb_staged`.

```
user → Next.js builder → Go API /run → pgmq → Go worker
     → Go dispatcher (disk guard + RunConfig) → POST /sync
     → Python ELT (Phase 0 → 1 → 2 → 3 → 4 → 5)
     → Go /internal/elt-callback → Supabase Realtime → Run Status Drawer
```

The phases are mandatory and ordered: pre-flight → extract+stage → dbt
transform → deliver → cleanup → callback. The callback always runs, even
on failure.

## 3. Contract between Go API and Python ELT

The dispatch payload is a strict `RunConfig`. Two fields replace earlier
loose shapes:

```jsonc
{
  "selected_streams": [
    {
      "stream_key": "public.users",
      "replication_method": "INCREMENTAL",
      "replication_key": "updated_at",
      "duckdb_table_name": "public__users"
    }
  ],
  "dbt_config": {
    "mode": "ui_sql",
    "sql_models": [
      {
        "source_stream_key": "public.users",
        "duckdb_source_table": "public__users",
        "output_table": "dim_users",
        "dest_table": "analytics.dim_users",
        "sql": "SELECT ... FROM {{ source('raw', 'public__users') }}"
      }
    ]
  }
}
```

Go builders that MUST be used to produce this payload:

- `duckdbTableName(streamKey)` — converts `schema.table` → `schema__table`.
- `buildSourceStreamConfigs(exec)` — builds the `selected_streams` list.
- `buildStrictDBTSQLModels(dbtConfig, exec, destSchema, srcSchema)` — fills
  `source_stream_key`, `duckdb_source_table`, `output_table`, `dest_table`,
  `sql`.

All three live in
[`apps/server/main-server/internal/server/pipeline_bundle.go`](../apps/server/main-server/internal/server/pipeline_bundle.go).

TypeScript mirror: `duckdbTableNameForStream`, `buildSourceStreamConfig` in
[`apps/app/lib/pipelines/schema-table.ts`](../apps/app/lib/pipelines/schema-table.ts).
The builder graph normaliser
[`apps/app/app/workspace/data-pipelines/[id]/builder/shared/pipelineGraph.ts`](../apps/app/app/workspace/data-pipelines/[id]/builder/shared/pipelineGraph.ts)
uses these helpers so the on-disk graph JSON always satisfies the shape.

## 4. Python handlers (strict refactor)

```
apps/server/elt-server/runner/
├── paths/
│   └── duckdb_staged.py           # 5-phase orchestrator
└── handlers/
    ├── __init__.py                # facade package
    ├── disk_guard.py              # Phase 0 disk budget (gap 10)
    ├── preflight_handler.py       # Phase 0 source/dest/column checks (3, 4, 10)
    ├── state_handler.py           # extract + restore cursor (gap 5)
    ├── normalisation_handler.py   # type/rename normalisation facade
    ├── dbt_handler.py             # UI SQL → dbt project (gap 6)
    └── delivery_handler.py        # DuckDB → destination write (3, 4, 12)
```

`duckdb_staged.py` orchestrates; it **never** inlines pre-flight or state
logic. New phase logic MUST live in `runner/handlers/*`.

## 5. Callback payload (post-audit)

```jsonc
{
  "status": "completed",
  "rows_read": 12345,
  "rows_written": 12340,
  "duration_seconds": 18.4,
  "checkpoint": { "lastSyncValue": "2026-04-19T12:00:00Z" },
  "delivery_outputs": ["analytics.dim_users", "analytics.fct_orders"],
  "delivered_tables": 2,
  "delivery_failures": [],
  "dbt_run_status": "success",
  "dbt_models_run": 2,
  "staging_size_bytes": 148234234,
  "no_pk_warnings": []
}
```

The Go callback handler
[`apps/server/main-server/internal/server/callback.go`](../apps/server/main-server/internal/server/callback.go)
persists these fields into `pipeline_runs.run_metadata` (JSONB).
The frontend type [`PipelineRunMetadata`](../apps/app/lib/api/types/data-pipelines.ts)
and the `RunStatusDrawer` render them.

## 6. Disk-guard (gap 10)

Two enforcement points:

1. **Go dispatcher** (`apps/server/main-server/internal/server/staged_delivery.go`,
   `dispatch.go`, `dispatch_incremental.go`): calls `ELT.DiskStatus()` and
   requires `available_gb ≥ plan_limit × 2`. If the disk is busy it marks
   the run `waiting` (new status) and re-queues with a 30 s delay using
   `handleDiskBusy`. Retry count is NOT incremented so the run keeps trying
   until capacity frees.
2. **Python Phase 0** (`runner/handlers/disk_guard.ensure_disk_budget`):
   re-checks inside the ELT server so an overlapping run cannot race the
   dispatcher check.

## 7. What the graph snapshot does NOT reflect yet

- `apps/server/elt-server/runner/handlers/` as a first-class layer.
- `handleDiskBusy` + the `waiting` pgmq path.
- `no_pk_warnings` propagating through Python → Go metadata → `PipelineRun.runMetadata`.
- `useDiscoverDestinationTable` hook wired from `DestinationPanel.tsx`.
- The 3-phase `RunStatusDrawer` render with delivered table pills.

The next full graphify regen should pick these up automatically; until
then, this addendum is the source of truth.

## 8. Legacy removed / forbidden

The following MUST NOT reappear. They are blocked by
[`.cursor/rules/strict-elt-invariants.mdc`](../.cursor/rules/strict-elt-invariants.mdc)
and the **ELT invariants reviewer** persona in
[`.cursor/rules/agents-and-orchestration.mdc`](../.cursor/rules/agents-and-orchestration.mdc):

- `apps/server/elt-server/runner/paths/cdc_direct.py`
- `apps/server/elt-server/routes/cleanup.py`
- `apps/app/**/TransformDrawer*.tsx`
- `apps/app/**/TransformNode*.tsx`
- `apps/app/**/TransformMiniCard*.tsx`
- `transform_script` string
- `LOG_BASED` replication branches
- user-facing "ETL pipeline" / "ETL run" / "ETL server" labels (DB column
  names are exempt)
