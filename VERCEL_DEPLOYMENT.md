# Vercel Deployment Guide: API, ETL, and App

Deploy the ai-bi monorepo to Vercel: **API (NestJS)**, **ETL (FastAPI)**, and **App (Next.js)**.

## Architecture

```
┌─────────────────────┐     ┌─────────────────────┐     ┌─────────────────────┐
│  App (Next.js)      │────▶│  API (NestJS)        │────▶│  ETL (FastAPI)       │
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
| Install Command | `pip install --upgrade pip setuptools wheel && pip install -r requirements.txt && pip install -e ./connectors/tap-postgres && pip install -e ./connectors/tap-mysql && pip install -e ./connectors/tap-mongodb` |

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

## 2. API (NestJS) Deployment

**Root directory:** `apps/api`

### Vercel Project Settings

| Setting | Value |
|---------|-------|
| Framework Preset | **Other** |
| Root Directory | `apps/api` |
| Build Command | `bun run build:deploy` |
| Install Command | `bun install` |

> **Note:** Do not use Framework Preset "NestJS". The project uses `api/index.ts` + `@vercel/node` with rewrites so all requests hit a single exported handler. The legacy NestJS preset expects `main.js` to export a handler, which causes "No exports found in module" errors.

### Deploy

```bash
cd apps/api
vercel link
vercel env add DATABASE_URL production
vercel env add ETL_PYTHON_SERVICE_URL production   # Your ETL URL, e.g. https://ai-bi-etl.vercel.app
vercel env add ETL_PYTHON_SERVICE_TOKEN production # Same as SUPABASE_JWT_SECRET or SUPABASE_SERVICE_ROLE_KEY
vercel env add SUPABASE_URL production
vercel env add SUPABASE_SERVICE_ROLE_KEY production
vercel env add ALLOWED_ORIGINS production
vercel --prod
```

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
| `DATABASE_DIRECT_URL` | No | Direct Postgres URL for pgmq/pg_cron (session mode) |
| `ETL_PYTHON_SERVICE_URL` | Yes | ETL base URL, e.g. `https://ai-bi-etl.vercel.app` |
| `ETL_PYTHON_SERVICE_TOKEN` | Yes | Token for ETL auth (use `SUPABASE_SERVICE_ROLE_KEY` or `SUPABASE_JWT_SECRET`) |
| `SUPABASE_URL` | Yes | Supabase project URL |
| `SUPABASE_SERVICE_ROLE_KEY` | Yes | Supabase service role key |
| `SUPABASE_ANON_KEY` | No | Supabase anon key (for some guards) |
| `ALLOWED_ORIGINS` | Yes | Comma-separated: `https://app.ai-bi.vercel.app,https://api.ai-bi.vercel.app` |
| `FRONTEND_URL` | No | Frontend URL for CORS/redirects |
| `SUPABASE_WEBHOOK_SECRET` | No | For Supabase user webhooks |

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

Ensure `installCommand` includes the tap connectors. Check `apps/etl/vercel.json`.

### ETL: Function timeout

Increase `maxDuration` in `vercel.json` (Pro: 60s, Enterprise: 300s).

### API: ETL_PYTHON_SERVICE_URL must be valid

Set `ETL_PYTHON_SERVICE_URL` to your deployed ETL URL. Must include `https://`.

### API: Migration fails during build

Set `DATABASE_URL` in Vercel before first deploy. Migrations run in `build:deploy`.

### API: "No exports found in module main.js"

The NestJS Framework Preset loads `main.js`, which does not export a handler. Fix:

1. Set Framework Preset to **Other** (not NestJS).
2. Ensure `apps/api/vercel.json` has `builds` + `rewrites`:
   ```json
   "builds": [{ "src": "api/index.ts", "use": "@vercel/node" }],
   "rewrites": [{ "source": "/(.*)", "destination": "/api" }]
   ```
3. The entrypoint `api/index.ts` re-exports the handler from `src/vercel.ts`.

### API: CORS blocked – "No Access-Control-Allow-Origin header"

1. Set `ALLOWED_ORIGINS` in the API Vercel project to include the frontend origin, e.g.:
   - `https://cloud.mantrixflow.com` (if frontend is on that domain)
   - `https://cloud.api.mantrixflow.com` (if calling API from same domain)
2. CORS is configured in `src/app-factory.ts` from `ALLOWED_ORIGINS`.

### App: NEXT_PUBLIC_API_URL not set

Set `NEXT_PUBLIC_API_URL` to your deployed API URL. Required for API calls.

---

## 8. Quick Reference

| Service | URL pattern | Health check |
|---------|-------------|--------------|
| ETL | `https://[project].vercel.app` | `GET /health` |
| API | `https://[project].vercel.app` | `GET /api` |
| App | `https://[project].vercel.app` | — |
| Meltano pipeline | `POST /run-meltano-pipeline` (on ETL) | — |
