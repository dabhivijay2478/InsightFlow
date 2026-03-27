# Vercel Deployment Guide: API, ETL, and App

Deploy the ai-bi stack: **API (Go / [Fiber](https://gofiber.io/) in `apps/api`)**, **ETL (FastAPI)**, and **App (Next.js)**. The NestJS API has been removed; HTTP routes live under **`/api/v1`**.

## Architecture

```
┌─────────────────────┐     ┌─────────────────────┐     ┌─────────────────────┐
│  App (Next.js)      │────▶│  API (Go / Fiber)      │────▶│  ETL (FastAPI)       │
│  app.ai-bi.vercel.app│     │  api.ai-bi.vercel.app │     │  etl.ai-bi.vercel.app│
└─────────────────────┘     │                      │     │                      │
                            │  - Pipelines         │     │  - /collect          │
                            │  - Data sources      │     │  - /emit             │
                            │  - runMeltanoPipeline│     │  - /run-meltano-pipeline│
                            └──────────────────────┘     └──────────────────────┘
                                     │
                                     ▼
                            ┌─────────────────────┐
                            │  Supabase / Postgres │
                            │  data_source_connections│
                            └─────────────────────┘
```

## Meltano on Vercel

**Meltano is NOT installed on Vercel.** All pipelines use **dynamic mode**:

- The API fetches connection config from `data_source_connections`
- The API calls `POST /run-meltano-pipeline` on the ETL with those configs
- ETL uses Singer taps (tap-postgres, tap-mongodb, tap-mysql) for collect + emit

---

## 1. ETL (FastAPI) Deployment

**Root directory:** `apps/etl`

### Vercel Project Settings

| Setting | Value |
|---------|-------|
| Framework Preset | Other |
| Root Directory | `apps/etl` |
| Build Command | *(auto)* |
| Output Directory | *(auto)* |
| Install Command | `pip install --upgrade pip setuptools wheel && pip install -r requirements.txt && pip install meltano && meltano install` |

### Deploy

```bash
cd apps/etl
vercel link
vercel env add SUPABASE_JWT_SECRET production
vercel env add CORS_ORIGINS production  # e.g. https://app.ai-bi.vercel.app,https://api.ai-bi.vercel.app
vercel --prod
```

### Function Timeouts

ETL operations can be slow. Set `maxDuration` in `apps/etl/vercel.json` or Vercel Dashboard:

- **Hobby**: 10s (may timeout on large syncs)
- **Pro**: 60s
- **Enterprise**: 300s

---

## 2. API (Go) Deployment

**Root directory:** `apps/api`

The service is a standard Go binary (`cmd/server`). Deploy it on any platform that runs containers or long-lived processes (Fly.io, Railway, Kubernetes, ECS, etc.). **Vercel** is optimized for serverless Node/Python; for Go on Vercel use a **Docker** deployment or run the API elsewhere and point `NEXT_PUBLIC_API_URL` at it.

### Build and run (any host)

```bash
cd apps/api
go build -o bin/server ./cmd/server
./bin/server   # listens on $PORT, default 8080
```

### Environment variables

Set the same variables listed in **§6 API (`apps/api`)** below. Required: `DATABASE_URL`, `SUPABASE_JWT_SECRET` (or `JWT_SECRET`), `ENCRYPTION_MASTER_KEY`, `ETL_PYTHON_SERVICE_URL`. Set **`API_PUBLIC_URL`** to the public origin of this API (no trailing slash); ETL callbacks use `{API_PUBLIC_URL}/api/v1/internal/etl-callback` and `{API_PUBLIC_URL}/api/v1/internal/checkpoint/{pipelineId}`.

### Optional: Vercel with Docker

If you use Vercel’s container runtime, add a `Dockerfile` in `apps/api` that builds `cmd/server` and set the project root to `apps/api`. Do not use the old NestJS `api/index.ts` handler.

`build:deploy` runs `nest build` then `db:migrate`. Ensure `DATABASE_URL` is set before first deploy.

---

## 3. App (Next.js) Deployment

**Root directory:** `apps/app`

### Vercel Project Settings

| Setting | Value |
|---------|-------|
| Framework Preset | Next.js |
| Root Directory | `apps/app` |
| Build Command | `bun run build` |
| Install Command | `bun install` |

### Deploy

```bash
cd apps/app
vercel link
vercel env add NEXT_PUBLIC_API_URL production       # e.g. https://ai-bi-api.vercel.app
vercel env add NEXT_PUBLIC_SUPABASE_URL production
vercel env add NEXT_PUBLIC_SUPABASE_ANON_KEY production
vercel env add NEXT_PUBLIC_SITE_URL production      # e.g. https://ai-bi-app.vercel.app
vercel env add SUPABASE_SERVICE_ROLE_KEY production
vercel --prod
```

---

## 4. Monorepo Setup (3 Projects)

Create **3 separate Vercel projects** from the same repo:

| Project | Root Directory | Branch |
|---------|----------------|--------|
| ai-bi-etl | `apps/etl` | main |
| ai-bi-api | `apps/api` | main |
| ai-bi-app | `apps/app` | main |

1. **New Project** → Import from Git
2. For each project, set **Root Directory** in Settings → General
3. Enable **Auto-deploy** on push to `main`

---

## 5. Cross-Service URLs

After deployment, set these so services can talk to each other:

**API project:**
```
ETL_PYTHON_SERVICE_URL=https://ai-bi-etl.vercel.app
ALLOWED_ORIGINS=https://ai-bi-app.vercel.app,https://ai-bi-api.vercel.app
```

For MantrixFlow domains, include the frontend origin:
```
ALLOWED_ORIGINS=https://cloud.mantrixflow.com,https://cloud.api.mantrixflow.com
```

**App project:**
```
NEXT_PUBLIC_API_URL=https://ai-bi-api.vercel.app
NEXT_PUBLIC_SITE_URL=https://ai-bi-app.vercel.app
```

The frontend must call the Go API with the **`/api/v1`** prefix (for example, `{NEXT_PUBLIC_API_URL}/api/v1/organizations/{orgId}/pipelines`).

---

## 6. Required Environment Variables Reference

### ETL (`apps/etl`)

| Variable | Required | Description |
|----------|----------|-------------|
| `SUPABASE_JWT_SECRET` | Yes | JWT secret for auth validation (same as Supabase JWT secret) |
| `CORS_ORIGINS` | No | Comma-separated origins, or `*` for all |
| `LOG_LEVEL` | No | `INFO` (default) |
| `PORT` | No | `8001` (default) |
| `TAP_TIMEOUT_SECONDS` | No | `1200` (default) |
| `EMIT_CHUNK_SIZE` | No | `1000` (default) |
| `PYTHON_VERSION` | No | `3.11` (set in vercel.json) |

### API (`apps/api`)

| Variable | Required | Description |
|----------|----------|-------------|
| `DATABASE_URL` | Yes | Postgres connection string (Supabase) |
| `DATABASE_DIRECT_URL` | No | Direct Postgres URL (port 5432) for pgmq worker if pooler is on 6543 |
| `SUPABASE_JWT_SECRET` or `JWT_SECRET` | Yes | Supabase JWT signing secret (HS256) |
| `ENCRYPTION_MASTER_KEY` | Yes | Min 32 chars; must match legacy Nest credential encryption |
| `ETL_PYTHON_SERVICE_URL` | Yes | ETL base URL, e.g. `https://ai-bi-etl.vercel.app` |
| `ETL_INTERNAL_TOKEN` | Yes* | Sent as `X-ETL-Token` to ETL (`SUPABASE_SERVICE_ROLE_KEY` or dedicated secret) |
| `API_PUBLIC_URL` | Yes | Public base URL of this API (callbacks; no path suffix) |
| `CALLBACK_TOKEN` / `INTERNAL_TOKEN` | Yes* | ETL → API: header `X-Callback-Token` or `X-Internal-Token` on `/api/v1/internal/*` |
| `SUPABASE_URL` | No | Supabase project URL |
| `SUPABASE_SERVICE_ROLE_KEY` | No | Fallback for token chaining |
| `ALLOW_SOURCE_DB_MUTATIONS_FOR_CDC` | No | `true` to allow CDC runs that touch source DB |
| `PORT` | No | HTTP listen port (default `8080`) |

\* At least one ETL auth secret and one internal callback token must match what the ETL service sends.

### App (`apps/app`)

| Variable | Required | Description |
|----------|----------|-------------|
| `NEXT_PUBLIC_API_URL` | Yes | API base URL, e.g. `https://ai-bi-api.vercel.app` |
| `NEXT_PUBLIC_SUPABASE_URL` | Yes | Supabase project URL |
| `NEXT_PUBLIC_SUPABASE_ANON_KEY` | Yes | Supabase anon key (public) |
| `NEXT_PUBLIC_SITE_URL` | Yes | App URL, e.g. `https://ai-bi-app.vercel.app` |
| `SUPABASE_SERVICE_ROLE_KEY` | Yes | Supabase service role key (server-side) |
| `NEXT_PUBLIC_PYTHON_SERVICE_URL` | No | ETL URL (if app calls ETL directly) |
| `GOOGLE_FONTS_API_KEY` | No | For Google Fonts API |

---

## 7. Troubleshooting

### ETL: Conflicting functions and builds / Unmatched function pattern

Use **`builds` only** (not `functions`) for Python. The `functions` pattern doesn't match Python serverless functions, and you cannot use both. Set `maxDuration` in the build config:

```json
"builds": [{"src": "api/index.py", "use": "@vercel/python", "config": {"maxDuration": 60}}]
```

See [Conflicting functions and builds](https://vercel.com/docs/errors/error-list#conflicting-functions-and-builds-configuration) and [Unmatched function pattern](https://vercel.com/docs/errors/error-list#unmatched-function-pattern).

### ETL: ModuleNotFoundError tap_postgres

Ensure `installCommand` includes `meltano install`. Check `apps/etl/vercel.json`.

### ETL: Function timeout

Increase `maxDuration` in `vercel.json` (Pro: 60s, Enterprise: 300s).

### API: ETL_PYTHON_SERVICE_URL must be valid

Set `ETL_PYTHON_SERVICE_URL` to your deployed ETL URL. Must include `https://`.

### API: Worker not processing queues

Ensure `DATABASE_DIRECT_URL` (or a session-mode URL on port **5432**) is set if `DATABASE_URL` uses Supavisor on **6543**. The pgmq worker uses `SessionDatabaseURL()`.

### API: ETL callbacks return 401

Match **`CALLBACK_TOKEN`** / **`INTERNAL_TOKEN`** on the API to the token the ETL sends (`X-Callback-Token` or `X-Internal-Token`).

### API: CORS

Add a CORS middleware in `apps/api` if browsers call the API cross-origin; configure allowed origins to match your frontend.

### App: NEXT_PUBLIC_API_URL not set

Set `NEXT_PUBLIC_API_URL` to your deployed API URL. Required for API calls.

---

## 8. Quick Reference

| Service | URL pattern | Health check |
|---------|-------------|--------------|
| ETL | `https://[project].vercel.app` | `GET /health` |
| API | `https://[project].vercel.app` | `GET /health` or `GET /api/v1/health` |
| App | `https://[project].vercel.app` | — |
| Meltano pipeline | `POST /run-meltano-pipeline` (on ETL) | — |
