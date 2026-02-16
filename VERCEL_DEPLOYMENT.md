# Vercel Deployment Guide: ETL, API (NestJS), and Meltano

This guide covers deploying the ai-bi monorepo to Vercel: **ETL (FastAPI)**, **API (NestJS)**, and **Meltano pipelines** (dynamic mode).

## Architecture

```
┌─────────────────────┐     ┌─────────────────────┐     ┌─────────────────────┐
│  App / Website      │────▶│  API (NestJS)       │────▶│  ETL (FastAPI)      │
│  Next.js            │     │  api.ai-bi.vercel.app│     │  etl.ai-bi.vercel.app│
│  app.ai-bi.vercel.app│     │                     │     │                     │
└─────────────────────┘     │  - Pipelines        │     │  - /collect         │
                            │  - Data sources     │     │  - /emit            │
                            │  - runMeltanoPipeline│    │  - /run-meltano-pipeline│
                            └─────────────────────┘     └─────────────────────┘
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
- ETL uses Singer taps (tap-postgres, tap-mongodb) for collect + emit

This avoids Meltano's `meltano install` (large, and `target-mongodb` fails on Python 3.12).

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

### Environment Variables

| Variable | Required | Description |
|----------|----------|-------------|
| `SUPABASE_JWT_SECRET` | Yes | JWT secret for auth validation |
| `CORS_ORIGINS` | No | Comma-separated origins, or `*` |
| `LOG_LEVEL` | No | `INFO` (default) |
| `PYTHON_VERSION` | No | `3.11` (set in vercel.json) |

### Function Timeouts

ETL operations (collect, emit, run-meltano-pipeline) can be slow. Set `maxDuration`:

- **Hobby**: 10s (may timeout on large syncs)
- **Pro**: 60s
- **Enterprise**: 300s

Configure in `apps/etl/vercel.json` or Vercel Dashboard → Settings → Functions.

### Deploy

```bash
cd apps/etl
vercel link
vercel env add SUPABASE_JWT_SECRET production
vercel --prod
```

---

## 2. API (NestJS) Deployment

**Root directory:** `apps/api`

### Vercel Project Settings

| Setting | Value |
|---------|-------|
| Framework Preset | **NestJS** |
| Root Directory | `apps/api` |
| Build Command | `bun run build:deploy` |
| Install Command | `bun install` |

### Environment Variables

| Variable | Required | Description |
|----------|----------|-------------|
| `DATABASE_URL` | Yes | Postgres connection (Supabase) |
| `ETL_PYTHON_SERVICE_URL` | Yes | ETL Vercel URL, e.g. `https://ai-bi-etl.vercel.app` |
| `ETL_PYTHON_SERVICE_TOKEN` | Yes | Same as `SUPABASE_JWT_SECRET` (or `SUPABASE_SERVICE_ROLE_KEY`) |
| `SUPABASE_SERVICE_ROLE_KEY` | Yes | Supabase service role key |
| `SUPABASE_URL` | Yes | Supabase project URL |
| `ALLOWED_ORIGINS` | Yes | Comma-separated: `https://app.ai-bi.vercel.app,https://api.ai-bi.vercel.app` |
| `FRONTEND_URL` | No | Frontend URL |
| `APP_URL` | No | `https://api.ai-bi.vercel.app` |

### Build & Migrations

`build:deploy` runs `nest build` then `db:migrate`. Ensure `DATABASE_URL` is set so migrations succeed.

### Deploy

```bash
cd apps/api
vercel link
vercel env add DATABASE_URL production
vercel env add ETL_PYTHON_SERVICE_URL production  # Your ETL URL
vercel env add ETL_PYTHON_SERVICE_TOKEN production
vercel --prod
```

---

## 3. App / Website (Next.js) Deployment

**Root directory:** `apps/app` or `apps/website`

| Setting | Value |
|---------|-------|
| Framework Preset | Next.js |
| Root Directory | `apps/app` (or `apps/website`) |
| Build Command | `bun run build` |
| Install Command | `bun install` |

Set `NEXT_PUBLIC_API_URL` to your API URL.

---

## 4. Monorepo Setup (Multiple Projects)

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

After deployment, set:

**API project:**
```
ETL_PYTHON_SERVICE_URL=https://ai-bi-etl.vercel.app
```

**App project:**
```
NEXT_PUBLIC_API_URL=https://ai-bi-api.vercel.app
```

---

## 6. Meltano Pipelines (Dynamic Mode)

Pipelines run via the API → ETL flow:

1. User triggers pipeline in the app
2. API fetches `source_connection_config` and `dest_connection_config` from DB
3. API calls `POST https://ai-bi-etl.vercel.app/run-meltano-pipeline`
4. ETL runs collect → (transform) → emit

No Meltano binary or `meltano.yml` config needed on Vercel.

---

## 7. Troubleshooting

### ETL: ModuleNotFoundError tap_postgres

Ensure `installCommand` includes the tap connectors. Check `apps/etl/vercel.json`.

### ETL: Function timeout

Increase `maxDuration` in `vercel.json` (Pro: 60s, Enterprise: 300s). Or reduce sync batch size.

### API: ETL_PYTHON_SERVICE_URL must be valid

Set `ETL_PYTHON_SERVICE_URL` to your deployed ETL URL. Must include `https://`.

### API: Migration fails during build

Set `DATABASE_URL` in Vercel before first deploy. Migrations run in `build:deploy`.

---

## 8. Quick Reference

| Service | URL pattern | Health check |
|---------|-------------|--------------|
| ETL | `https://[project].vercel.app` | `GET /health` |
| API | `https://[project].vercel.app` | `GET /api` |
| Meltano pipeline | `POST /run-meltano-pipeline` (on ETL) | — |
