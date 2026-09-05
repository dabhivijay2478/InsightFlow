# Deployment on Vercel (notes)

This repo is a monorepo with:

- **App (Next.js)**: `apps/arcyria-platform`
- **Go API (Fiber)**: `apps/server/arcyria-server`
- **Python ELT (FastAPI)**: `apps/server/arcyria-elt`

## Recommended approach

- Deploy **`apps/arcyria-platform`** on Vercel.
- Deploy **`apps/server/arcyria-server`** and **`apps/server/arcyria-elt`** on a platform that supports long-running services (container / VM / k8s). Vercel is primarily serverless; it can be a poor fit for queue workers and long-lived servers.

## App (Next.js) on Vercel

**Root Directory**: `apps/arcyria-platform`
**Install**: `bun install`  
**Build**: `bun run build`

### Required env vars

- `NEXT_PUBLIC_API_URL` — Go API origin only (e.g. `https://api.example.com`)
- `NEXT_PUBLIC_SUPABASE_URL`
- `NEXT_PUBLIC_SUPABASE_ANON_KEY`
- `NEXT_PUBLIC_SITE_URL` — app origin (used for auth redirects)

## Cross-service URLs (prod)

The app calls the Go API under `/api/v1/...`:

- Example: `${NEXT_PUBLIC_API_URL}/api/v1/health`

The Go API talks to the ELT service via:

- `ELT_PYTHON_SERVICE_URL` (Go API env var)

The ELT service calls back to the Go API via:

- `API_PUBLIC_URL` (Go API env var) → callbacks land at `/api/v1/internal/...`
