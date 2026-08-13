# MantrixFlow — System Design & Technical Interview Guide

Use this document to present **MantrixFlow** in a system design or technical interview. It covers the three services you built: the **Next.js app**, the **Go API (main-server)**, and the **Python ELT server**. Tone is conversational — practice saying these sections out loud, not reading them word-for-word.

**Deeper references (optional):** [`md-docs/strict-elt-pipeline-guide.md`](md-docs/strict-elt-pipeline-guide.md), [`md-docs/source-to-destination-elt-flow.md`](md-docs/source-to-destination-elt-flow.md), [`apps/server/README.md`](apps/server/README.md).

---

## 30-second elevator pitch

> MantrixFlow is a multi-tenant SaaS for **ELT pipelines**: users connect a source (Postgres, MySQL, Stripe, HubSpot, etc.), map tables, write **dbt-style SQL** in the UI, and deliver transformed rows into **existing** destination tables. We never create tables in the customer database at runtime — delivery is **upsert/merge** when a primary key exists, otherwise append with a clear warning.
>
> Architecture: **Next.js** talks only to a **Go API** (auth, orchestration, encrypted credentials). Runs are **async** via **Postgres pgmq**. A **Python ELT worker** stages data in **ephemeral DuckDB**, runs **dbt-duckdb**, writes to the destination, then **callbacks** to Go so the UI updates in real time.

---

## 2-minute “what I built” story

1. **Product:** Pipeline builder on a canvas — one source, many destinations, each destination owns sync mode, SQL models, and final `schema.table` targets.
2. **Frontend:** Next.js 16, React 19, TanStack Query, Zustand for builder state, Monaco for SQL, Supabase auth in the browser.
3. **Control plane:** Go (Fiber) validates JWTs, enforces org roles, encrypts connection secrets (Fernet), builds a strict `RunConfig`, enqueues jobs, proxies discovery/preview/validate-sql to ELT without exposing ELT to the browser.
4. **Data plane:** Python FastAPI receives `POST /sync`, runs a **5-phase** DuckDB-staged pipeline (pre-flight → extract → dbt → deliver → cleanup + callback).
5. **Reliability:** Disk guard before dispatch and again in ELT; runs can sit in **`waiting`** and re-queue every 30s; concurrency capped with a worker semaphore; orphan/stale run cleanup.

---

## High-level architecture

```text
┌─────────────────────────────────────────────────────────────────────────┐
│  User (browser)                                                          │
│  apps/app — Next.js on Vercel (cloud.mantrixflow.com)                    │
│  • Supabase JWT in Authorization header                                  │
│  • All product APIs → Go /api/v1/...                                     │
└───────────────────────────────┬─────────────────────────────────────────┘
                                │ HTTPS + JWT
                                ▼
┌─────────────────────────────────────────────────────────────────────────┐
│  Go API — apps/server/main-server (ECS Fargate, public ALB)             │
│  • Auth, org scoping, billing (Dodo), pipelines, connections           │
│  • Encrypt/decrypt credentials; never return secrets in API/errors       │
│  • POST /pipelines/:id/run → pipeline_runs row + pgmq enqueue → 202      │
│  • Worker polls pgmq (~5s), disk check, POST /sync to ELT (internal)     │
│  • POST /internal/elt-callback ← ELT (X-Callback-Token)                  │
└───────────────┬─────────────────────────────┬───────────────────────────┘
                │ Postgres (Supabase)          │ Service Connect (private)
                │ • app data + RLS             ▼
                │                    ┌──────────────────────────────────────┐
                │                    │  Python ELT — apps/server/elt-server │
                │                    │  FastAPI :8000 (no public URL)       │
                │                    │  dlt + DuckDB + dbt-duckdb per run   │
                └────────────────────┤  Callback → Go with audit metadata │
                                     └──────────────────────────────────────┘
                                                │
                    ┌───────────────────────────┼───────────────────────────┐
                    ▼                           ▼                           ▼
            Customer source DB          Ephemeral .duckdb file      Customer dest DB
            (extract)                   (stage + dbt)               (upsert only)
```

**Design choice to mention:** The browser **never** calls ELT directly for authenticated flows. That keeps one security boundary, one credential path, and one contract owner (Go).

---

## End-to-end: user clicks Run

Walk interviewers through this sequence — it shows you understand the full system.

| Step | Who | What happens |
|------|-----|----------------|
| 1 | App | User saves graph; `POST .../pipelines/:id/run` with optional `destinationNodeId` / `branchId`. |
| 2 | Go | Validates JWT + org role; admission checks (plan limits, concurrent runs); creates `pipeline_runs` (`pending` → `queued`). |
| 3 | Go | Enqueues pgmq message (`pipeline_jobs` or `incremental_sync`) with `pipeline_id`, `run_id`, `org_id`, target destination. |
| 4 | Go | Returns **HTTP 202** with `run_id` — UI opens run drawer, may subscribe to Supabase Realtime. |
| 5 | Worker | Polls pgmq; acquires semaphore slot (`MAX_CONCURRENT_RUNS`, default ~2). |
| 6 | Worker | Calls ELT `GET /disk-status`; needs `available_gb >= plan_limit × 2`. If not → run **`waiting`**, re-queue **30s** (not failed). |
| 7 | Worker | Loads pipeline bundle, decrypts source + destination configs, builds `RunConfig`: `selected_streams[]`, `dbt_config`, `last_state`, `delivery_table_map`. |
| 8 | Worker | `POST /sync` to ELT with `X-ETL-Token`. |
| 9 | ELT | Phase 0 pre-flight: disk, source tables, **dest table must exist**, column match, PK discovery. |
| 10 | ELT | Phase 1: dlt extract → DuckDB `raw.schema__table` (double underscore in staging). |
| 11 | ELT | Phase 2: generate temp dbt project from UI SQL; `dlt.dbt.package()` in same DuckDB. |
| 12 | ELT | Phase 3: deliver each model to `dest_schema.destination_table` — **no CREATE TABLE**. |
| 13 | ELT | Phase 4: `state_handler.extract()` **then** delete DuckDB + dbt temp dir (`finally`). |
| 14 | ELT | Phase 5: callback to Go with `delivery_outputs`, `staging_size_bytes`, `dbt_models_run`, `no_pk_warnings`, checkpoint. |
| 15 | Go | Updates `pipeline_runs`, saves checkpoint on pipeline, publishes status; worker may promote next queued run. |
| 16 | App | Run drawer shows Stage / dbt / Deliver phases from `run_metadata`. |

---

## Service 1 — App (`apps/app`)

### Role in the system

The app is the **operator console**: workspaces, connections, pipeline canvas builder, run status, analytics, billing UI, Slack/GitHub integrations. It is **not** a second backend — business rules for runs live in Go + ELT.

### Stack (say this clearly)

- **Next.js 16** (App Router), **React 19**, **TypeScript**
- **TanStack Query** for server state; **Zustand** for builder/canvas state
- **Tailwind** + **Radix/shadcn** UI
- **Monaco** for SQL in the pipeline builder
- **Supabase Auth** — session/JWT in the client
- **PostHog** for product analytics (optional in dev)

### Auth flow (common interview question)

1. User signs in via Supabase (email/OAuth).
2. App attaches `Authorization: Bearer <access_token>` to Go API calls.
3. Go `AuthJWT` middleware validates HS256 (secret) or JWKS from Supabase URL.
4. Routes under `/api/v1/organizations/:organizationId/...` are org-scoped; role checks (`OWNER`, `ADMIN`, `EDITOR`) on mutating routes.

**Talking point:** We treat Supabase as **identity**; MantrixFlow enforces **authorization** in Go (org membership, editor vs viewer).

### Pipeline builder mental model

Say this in plain language:

- **One source node** — pick connection, select tables as `schema.table`, preview raw rows.
- **One or more destination nodes** — each edge is source → destination; each destination is independent (own sync mode, SQL, schedule).
- **Destination owns the contract:** `dest_schema`, `replication_method` (`FULL_TABLE` | `INCREMENTAL`), `replication_key`, `normalisation_rules`, `dbt_config.sql_models[]`.
- **Two names for tables:** internal DuckDB/dbt name (`output_table`, staging uses `schema__table`) vs final client target (`dest_schema.destination_table`).

### Key files to name if asked “where is X?”

| Area | Path |
|------|------|
| Pipeline builder page | `app/workspace/data-pipelines/[id]/builder/` |
| Source config / preview | `panels/SourcePanel.tsx` |
| Destination + SQL + dest discover | `panels/DestinationPanel.tsx` |
| Run progress UI | `panels/RunDetailsDrawer.tsx` (or `RunStatusDrawer`) |
| Graph → API shape | `shared/pipelineGraph.ts` |
| API client | `lib/api/services/*.ts`, `lib/api/constants.ts` |
| Types | `lib/api/types/data-pipelines.ts` |

### API contract from the frontend

- Base URL: `NEXT_PUBLIC_API_URL` (origin only, e.g. `http://localhost:8080`).
- All calls: `/api/v1/organizations/:orgId/...`
- Response envelope: `{ status, data, message, error_code? }`
- ELT helpers (discover, preview, validate-sql) go through Go proxy: `/organizations/:orgId/elt/...` — **not** direct to port 8000 in production.

### What you would demo in an interview

1. Create source + destination connections (test connection).
2. Add `public.users` on source; on destination set final table `analytics.dim_users` (must exist).
3. Write SQL with `{{ source('raw', 'public__users') }}` — note double underscore in dbt source.
4. Validate SQL → column match hints.
5. Run → show three phases and delivered table chips in the drawer.

---

## Service 2 — Go main-server (`apps/server/main-server`)

### Role in the system

Go is the **control plane and trust boundary**:

- Public REST API for the app
- Credential encryption at rest
- Pipeline/run lifecycle in Postgres
- Job queue orchestration
- Internal-only bridge to ELT (token auth)
- Callback ingestion and realtime-friendly status updates

### Stack

- **Fiber** (HTTP), **GORM** + Postgres (Supabase-hosted)
- **pgmq** for durable job queues
- **zerolog** structured logging (no secrets in logs)
- **Fernet** (or configured master key) for connection JSON encryption
- Swagger stubs under `docs/`

### Major subsystems

| Subsystem | Responsibility | Key paths |
|-----------|----------------|-----------|
| HTTP routes | Org-scoped REST | `internal/server/routes.go` |
| Pipelines | CRUD, validate, run, runs list | `run_pipeline.go`, `pipeline_*.go` |
| Dispatch | Build RunConfig, call ELT | `dispatch.go`, `dispatch_incremental.go` |
| Bundle builder | Graph → streams + dbt models | `pipeline_bundle.go` |
| Queue worker | Poll pgmq, concurrency, disk guard | `internal/worker/worker.go`, `staged_delivery.go` |
| ELT client | HTTP to Python service | `internal/elt/client.go` |
| Callback | Persist run metadata, checkpoint | `callback.go` |
| ELT proxy | Discover, preview, validate-sql | routes under `/elt/*` |
| Billing | Dodo checkout/portal/webhooks | billing handlers in `routes.go` |

### Run lifecycle states (good for system design)

Explain states you actually use:

- `pending` / `queued` — accepted, waiting for worker
- `running` — ELT `/sync` in flight
- `waiting` — disk guard: staging volume too full; **retry**, not failure
- `completed` / `failed` / `cancelled` — terminal

**Why `waiting` matters:** On shared ELT disk, failing every run would spam errors; backing off preserves fairness and auto-recovers when space frees up.

### Queue design talking points

- **Queues:** `pipeline_jobs` (full sync), `incremental_sync`, shorter VT queue for incremental checks, email queue separate.
- **Poll interval:** ~5s (configurable `PGMQ_POLL_INTERVAL_MS`).
- **Concurrency:** Semaphore sized to `PipelineMaxConcurrent` — don’t dequeue more than you can run.
- **Idempotency / duplicates:** `archiveDuplicateRunDispatch` avoids double-dispatch for the same run.
- **Safety nets:** Orphan run cleanup, stale queue cleanup, `DrainPipelineQueue` when callback promotes next job.

### RunConfig — the contract between Go and Python

This is high-value in interviews. Go sends a single structured payload (Pydantic `RunConfig` on the Python side):

- **`selected_streams`:** array of objects, not strings:
  - `stream_key`: `"public.users"`
  - `replication_method`: `FULL_TABLE` | `INCREMENTAL`
  - `replication_key`: optional cursor column
  - `duckdb_table_name`: `"public__users"` (sanctioned converter in `pipeline_bundle.go`)
- **`dbt_config`:** `mode: "ui_sql"`, `sql_models[]` with `output_table`, `destination_table`, `sql`, and delivery mapping
- **`last_state`:** incremental checkpoint from prior run
- **Encrypted credentials** in payload — ELT decrypts in memory only

### Security answers (prepare these)

| Topic | What we do |
|-------|------------|
| Credentials | Encrypted in DB; decrypted only in Go/ELT process memory; masked in GET connection APIs |
| Logs/errors | `sanitizeELTError`; no passwords in responses |
| ELT exposure | Internal network + `X-ETL-Token`; callback uses `X-Callback-Token` |
| Multi-tenant | `organization_id` on all resources; Supabase **RLS** policies (`sql/supabase_rls.sql`) tested with real JWTs |
| Auth | Supabase JWT on `/api/v1/*`; webhooks (Dodo) verified by signature, no JWT |

### Production deployment (one sentence each)

- App: **Vercel** → `cloud.mantrixflow.com`
- API: **AWS ECS Fargate** behind ALB → `cloud.api.mantrixflow.com`
- ELT: **ECS Service Connect** `elt-service:8000` — no public DNS
- Secrets: **AWS SSM** `/mantrixflow/production/*`

See [`apps/mantrixflow-infra/DEPLOYMENT.md`](apps/mantrixflow-infra/DEPLOYMENT.md) if they go deeper on infra.

---

## Service 3 — ELT server (`apps/server/elt-server`)

### Role in the system

ELT is the **data plane**: it touches customer databases and ephemeral disk. It should stay **stateless across requests** except for per-run DuckDB files and temp dbt directories.

### Stack

- **FastAPI** + **Pydantic v2**
- **dlt** for extract/load patterns
- **DuckDB** per-run staging file
- **dbt-core** + **dbt-duckdb** for transforms
- **SQLAlchemy** for destination introspection and delivery

### Active API surface (internal)

| Method | Path | Purpose |
|--------|------|---------|
| GET | `/health` | Health |
| POST | `/test-connection` | Source or destination connectivity |
| POST | `/discover` | Source schema discovery |
| POST | `/preview` | Sample rows / SQL model preview |
| POST | `/validate-sql` | DuckDB in-memory SQL validation |
| POST | `/discover-table` | Destination table + columns + PKs |
| GET | `/disk-status` | Free space for dispatcher guard |
| POST | `/sync` | Start async pipeline run |

### Five phases (memorize this)

```
Phase 0 — Pre-flight (preflight_handler)
  • Disk budget: available >= plan_limit × 2
  • Source table exists; incremental cursor column exists if needed
  • Destination table EXISTS (hard fail — no CREATE TABLE)
  • Model columns ⊆ destination columns (named error per column)
  • PK discovery → merge vs append + no_pk_warnings

Phase 1 — Extract + Stage (duckdb_staged.py + dlt)
  • Restore cursor from last_state
  • Stage as raw.schema__table in DuckDB
  • Apply destination-scoped normalisation (rename/cast)

Phase 2 — dbt Transform (dbt_handler)
  • Generate temp dbt project from ui_sql models
  • sources.yml maps raw.schema__table
  • Outputs materialized in DuckDB (analytics schema internally)

Phase 3 — Deliver (delivery_handler)
  • Re-verify dest table + columns
  • Upsert/merge when PK exists; append + warning when not
  • Strip _dlt_* from customer DB (never leave staging artifacts)

Phase 4 — Cleanup (always finally)
  • state_handler.extract() BEFORE os.remove(duckdb)
  • Remove dbt temp workspace

Phase 5 — Callback
  • POST Go /internal/elt-callback with audit fields
```

### The 12 invariants (if they ask “what rules are non-negotiable?”)

Summarize from your strict ELT rules — interviewers like discipline:

1. DuckDB file **always** deleted in `finally`
2. Cursor state extracted **before** DuckDB delete
3. Destination tables **never** created by runner
4. Column mismatch fails with **named column**
5. No `_dlt_*` tables left in client destination
6. Credentials never in API responses or error messages
7. Cross-tenant **RLS** is real (tested with JWT)
8. Source `schema.table` vs staging `schema__table` — single converters in Go + TS
9. Public delivery is **upsert-only** (no user-selectable destructive modes)
10. Disk check at **dispatcher and Phase 0**; `waiting` + 30s re-queue
11. Callback includes `delivery_outputs`, `staging_size_bytes`, `dbt_models_run`, `no_pk_warnings`
12. No legacy ETL paths (CDC direct, transform_script, etc.)

### Handlers layout (shows clean architecture)

- Orchestrator: `runner/paths/duckdb_staged.py`
- Phase logic: `runner/handlers/` — `preflight_handler`, `state_handler`, `disk_guard`, `delivery_handler`, `normalisation_handler`
- Models: `models/run_config.py`, `models/callback_payload.py`

### Scaling the ELT tier

Talking points (even if not all implemented yet):

- Horizontal scale: **multiple ELT tasks** behind internal load balancer; runs are independent per DuckDB file.
- Bottleneck: **disk** on staging volume → disk guard + plan-based limits.
- Bottleneck: **destination DB write rate** → incremental sync, merge keys, concurrency cap at Go worker.
- Long runs: async `/sync` + callback; Go doesn’t hold HTTP open for the whole pipeline.

---

## System design questions — suggested answers

### Why DuckDB staging instead of loading straight to the warehouse?

> We need a consistent place to run **user SQL (dbt)** between extract and deliver. DuckDB per run gives isolated, fast local analytics SQL without polluting the customer database with raw or `_dlt_*` tables. The destination only sees **final** model outputs.

### Why Go + Python instead of one service?

> **Separation of concerns:** Go owns auth, multi-tenant CRUD, queueing, and stable public APIs. Python owns the heavy data stack (dlt, DuckDB, dbt) where the ecosystem is stronger. The contract is a versioned `RunConfig` + callback payload.

### How do you handle failures?

> Pre-flight fails fast with user-safe messages. Disk busy → `waiting` + retry. ELT errors sanitized at Go callback. Partial delivery tracked in `delivery_failures[]`. Orphan `running` rows cleaned by worker tickers.

### How do incremental syncs work?

> Destination chooses `INCREMENTAL` + `replication_key`. Go passes `last_state` from prior checkpoint. ELT restores dlt cursor, extracts deltas, delivers merge/upsert when PK exists. New checkpoint returned in callback and stored on the pipeline.

### How is multi-tenancy enforced?

> Every resource keyed by `organization_id`. Supabase RLS on `pipelines`, `pipeline_runs`, `data_sources`, etc. Go never trusts client org id without JWT + membership check.

### What would you improve next?

Pick 1–2 honest items, e.g.:

- Stronger observability dashboards per phase latency
- Per-connector rate limits and backpressure metrics
- Read replicas for analytics API if query load grows

---

## Technical deep-dive questions

| Question | Short answer |
|----------|----------------|
| What is `schema.table` vs `schema__table`? | User and APIs use `schema.table` (e.g. `public.users`). DuckDB staging uses `schema__table` to avoid collisions across schemas. |
| Where is SQL validated? | ELT `/validate-sql` via Go proxy; in-memory DuckDB schema, no live source writes. |
| How are secrets stored? | Encrypted connection JSON in Postgres; Fernet master key in env/SSM. |
| Can the UI create destination tables? | Builder may help **design** tables via pipeline-destination-schema APIs; **runtime delivery never CREATE TABLE**. |
| What connectors are production-ready? | PostgreSQL, MySQL, and Airtable are available as sources and destinations. Asana, HubSpot, and Stripe are available as sources. MySQL, Airtable, and Asana remain runtime-capability gated. |
| How does realtime UI update? | Go persists run + metadata; Supabase Realtime publish on status (Run drawer polls/subscribes). |

---

## Local dev cheat sheet (if they ask “how do you run it?”)

```bash
# Terminal 1 — App
cd apps/app && bun install && bun run dev

# Terminal 2 — Go API
cd apps/server/main-server && go run ./cmd/server

# Terminal 3 — ELT
cd apps/server/elt-server
./.venv/bin/python -m uvicorn api.main:app --host 0.0.0.0 --port 8000 --loop asyncio
```

Env: app needs `NEXT_PUBLIC_API_URL`, Supabase URL/anon key; Go needs DB URL, JWT secret, encryption key, ELT URL + tokens — see each service’s env examples.

---

## Presentation checklist (day of interview)

- [ ] Draw the three-box diagram (App → Go → ELT → callback) on a whiteboard in under 60 seconds.
- [ ] Walk one **Run** click through 5 ELT phases without looking at notes.
- [ ] Explain **why** destination tables must pre-exist (customer control, no surprise DDL).
- [ ] Mention **disk guard** and `waiting` — shows you thought about ops, not just happy path.
- [ ] Name **one security** measure (encrypted credentials + internal ELT) and **one multi-tenant** measure (RLS + org routes).
- [ ] Have **one tradeoff** ready (DuckDB staging vs complexity; Go/Python split vs ops).
- [ ] Point to **real files** you wrote (builder panel, dispatch, `duckdb_staged.py`) — credibility matters.

---

## Repo map (quick)

```text
ai-bi/
├── README.md                 ← this interview guide
├── apps/app/                 ← Next.js product UI
├── apps/server/main-server/  ← Go API + worker
├── apps/server/elt-server/   ← Python ELT
├── apps/mantrixflow-infra/   ← AWS ECS deployment
└── md-docs/                  ← extended product & ops docs
```

Good luck — lead with the user problem (reliable ELT into **existing** tables), then show how the three services divide responsibility cleanly.
