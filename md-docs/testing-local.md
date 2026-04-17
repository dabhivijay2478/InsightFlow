# Local testing (Docker DBs + smoke tests)

This doc is a lightweight local testing checklist for the **current** services:

- Go API: `apps/server/main-server`
- Python ELT: `apps/server/elt-server`
- App: `apps/app`

## Docker test databases

If this repo contains `docker-compose.test.yml`, use it to boot local DBs:

```bash
docker compose -f docker-compose.test.yml up -d
```

## Go API tests (main-server)

```bash
cd apps/server/main-server
go test ./...
```

## ELT service smoke test

Start the ELT:

```bash
cd apps/server/elt-server
./.venv/bin/python -m uvicorn api.main:app --host 0.0.0.0 --port 8000 --loop asyncio
```

Then:

- `GET http://localhost:8000/health`

## End-to-end sanity

With Go API running on `:5000` and ELT on `:8000`:

- `GET http://localhost:5000/api/v1/health`
- Run a small pipeline from the app (or hit the API endpoint that enqueues a run) and confirm the internal callback is received.

