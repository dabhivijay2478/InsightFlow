# Go API and Microsandbox manager deployment

The Go API and simulation manager are independently deployable applications in
self-hosted Dokploy. Both run on the persistent OVH Go VPS; neither runs on the
Dokploy VPS. Temporary simulation hosts are created directly through the OVH
API and are never Dokploy applications.

The complete infrastructure sequence is maintained in:

- [`../../mantrixflow-infra/docs/setup-guide.md`](../../infrastructure/setup/ovh-dokploy-microsandbox.md)
- [`../../mantrixflow-infra/DEPLOYMENT.md`](../infrastructure/ovh-microsandbox-runbook.md)
- [`../../mantrixflow-infra/docs/github-actions.md`](../../infrastructure/ci-cd/github-actions.md)

## GitHub deployment units

| Workflow | GitHub environment | Dokploy application |
| --- | --- | --- |
| `Go API CI/CD` | `production-go-api` | `DOKPLOY_GO_APPLICATION_ID` |
| `Simulation Manager CI/CD` | `production-simulation-manager` | `DOKPLOY_SIMULATION_MANAGER_APPLICATION_ID` |
| `Production Simulation Canary` | `production-simulation-canary` | no deployment; invokes Go API |

Merges to `main` or `mantrixflow` publish immutable GHCR digests. A manager-only
change does not redeploy the normal Go API, and a normal API change does not
redeploy the manager when it matches the workflow path filters.

## Runtime separation

The Go API application:

```text
SIMULATION_PLATFORM_ENABLED=false  # change to true during canary rollout
SIMULATION_WORKER_ENABLED=false
PORT=5000
ELT_PYTHON_SERVICE_URL=http://10.20.0.30:8000
ELT_GRPC_ADDRESS=10.20.0.30:8001
```

The simulation-manager application:

```text
SIMULATION_PLATFORM_ENABLED=true
SIMULATION_WORKER_ENABLED=true
SIMULATION_SANDBOX_BACKEND=microsandbox
MAX_SIMULATION_CONCURRENCY=1
DISABLE_AUTO_MIGRATE=true
```

Use `environments/production/simulation-manager.env.example` in the
infrastructure repository for the complete manager variable list. Store real
values in Dokploy, not GitHub deployment secrets and not committed files.

## Simulation release order

1. Apply and verify the additive simulation migrations with
   `cmd/simulation-migrate` using an explicit direct Supabase URL.
2. Publish and sign the simulation runtime from the ELT repository.
3. Put the exact runtime digest and Git SHA version into the manager environment.
4. Deploy exactly one manager replica with rollout disabled.
5. Enable one canary organization in both the API and manager configuration.
6. Run the protected production canary.
7. Confirm microVM cleanup and eventual hourly OVH host deletion.

Application rollback selects the previous Go or manager digest in only the
affected Dokploy application. Disable simulation admission before rolling back
the manager or runtime.
