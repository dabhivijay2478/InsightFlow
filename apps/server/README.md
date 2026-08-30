# MantrixFlow Server (Go API + Python ELT)

The backend uses independent OVHcloud targets managed by self-hosted Dokploy:

| Service | Persistent target | Exposure |
| --- | --- | --- |
| Dokploy control plane | Dedicated OVHcloud VPS-1 | Public management HTTPS |
| Go API and simulation manager | OVHcloud VPS-1 | Public HTTPS for Go only |
| Production Python ELT | OVHcloud VPS-2 | Private gRPC and health traffic |
| Future PostgreSQL and PgBouncer | Separate OVHcloud VPS-2 | Prepared; Supabase remains active |
| Microsandbox | On-demand OVHcloud Public Cloud host | Private, temporary simulation compute |

Use [`apps/mantrixflow-infra/DEPLOYMENT.md`](../mantrixflow-infra/DEPLOYMENT.md)
for production operations. Self-hosted Dokploy runs on its own VPS and manages
the other servers over dedicated SSH. Go, ELT, and future database deployments
are independent, and production ELT never runs inside Microsandbox.

First-time infrastructure, private routing, dynamic-host trust, and protected
GitHub setup are documented in
[`apps/mantrixflow-infra/docs/setup-guide.md`](../mantrixflow-infra/docs/setup-guide.md)
and [`apps/mantrixflow-infra/docs/github-actions.md`](../mantrixflow-infra/docs/github-actions.md).

## Local development

Run services separately:

```bash
# Go API
cd apps/server/main-server && go run ./cmd/server

# Python ELT
cd apps/server/elt-server
./.venv/bin/python -m uvicorn api.main:app --host 127.0.0.1 --port 8000 --loop asyncio
```

See each service README and `.env.production.example` for configuration. The
optional Compose file in this directory is for local validation only and is
not the production topology.

## Production boundaries

- Vercel hosts the Next.js frontend.
- The browser calls only the Go API for business operations.
- Go calls production ELT over authenticated private gRPC.
- Supabase PostgreSQL and Auth remain active in this phase.
- The Go simulation manager creates temporary OVHcloud Public Cloud hosts and
  one Microsandbox microVM per simulation run.
- Internal gRPC, PostgreSQL, PgBouncer, worker endpoints, and Microsandbox
  control interfaces are never public.

Provider-specific deployment material outside
[`apps/mantrixflow-infra`](../mantrixflow-infra/) is historical and must not be
used as current production guidance.
