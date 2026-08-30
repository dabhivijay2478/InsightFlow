# Production architecture

```text
GitHub Actions --immutable images--> GHCR --deploy API--> self-hosted Dokploy (OVH VPS-1)
                                                              |--SSH--> Go (OVH VPS-1)
                                                              `--SSH--> ELT (OVH VPS-2)

Vercel Next.js --HTTPS/REST+SSE--> Go
                                      |--private TLS gRPC--> Python ELT
                                      |--SQL/PGMQ--> Supabase PostgreSQL (current)
                                      `--OVH API--> temporary Public Cloud host
                                                           `--Microsandbox--> one microVM/run

Future only: Go --private PgBouncer--> PostgreSQL+PGMQ (prepared OVH VPS-2)
```

Terraform owns the four persistent VPSs, their OVH edge firewalls and IAM tags,
plus the shared OVH Public Cloud region, subnet, and SSH key. It never deploys
containers or creates a host for an individual simulation request.

Dokploy is a dedicated control plane. Go, ELT, and future database workloads run
on remote deployment servers and continue running if Dokploy, GitHub Actions, or
GHCR is temporarily unavailable.

The Go manager owns dynamic-host states from `STOPPED` through `TERMINATING`,
starts one isolated Microsandbox microVM per run, and deletes the outer OVH host
after its idle timeout. Supabase remains the production database and Auth service.

Vercel Sandbox is not used in the current architecture. The required private
route, KVM-capable image, SSH host CA, GitHub environment matrix, and staged
rollout are documented in [setup-guide.md](../setup/ovh-dokploy-microsandbox.md).
