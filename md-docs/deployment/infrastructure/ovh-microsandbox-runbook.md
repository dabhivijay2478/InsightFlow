# OVH and Microsandbox deployment runbook

Use [docs/setup-guide.md](../../infrastructure/setup/ovh-dokploy-microsandbox.md) for first-time account,
network, image, host-CA, Tigris, GitHub, and Dokploy setup. This runbook covers
repeatable deployment after those prerequisites are complete.

## Architecture guardrails

- Terraform manages persistent OVH VPSs, edge firewalls, tags, and shared
  simulation network/key resources.
- Self-hosted Dokploy deploys persistent Go, simulation-manager, and ELT
  containers to remote OVH VPSs.
- Go creates and deletes temporary OVH Public Cloud hosts through the OVH API.
- Microsandbox runs one microVM per simulation run on that temporary host.
- Supabase remains the production PostgreSQL/Auth/PGMQ service.
- The prepared OVH PostgreSQL/PgBouncer services remain stopped.
- Vercel Sandbox is not part of this deployment.

## 1. Pre-deployment checks

From the infrastructure repository:

```bash
./scripts/check-github-environments.sh
terraform fmt -check -recursive infra/terraform
for environment in staging production; do
  terraform -chdir="infra/terraform/environments/$environment" init -backend=false
  terraform -chdir="infra/terraform/environments/$environment" validate
done
for script in scripts/*.sh; do bash -n "$script"; done
```

Confirm:

- the pull request targets the correct repository and deployment branch;
- no populated `.env`, `terraform.tfvars`, key, certificate, or state file is
  staged;
- the chosen runtime image is an immutable digest;
- application and simulation migrations are reviewed separately;
- the simulation private route, KVM image, host certificate, and Tigris bucket
  passed staging validation.

## 2. Reconcile static infrastructure

Open a pull request in `dabhivijay2478/mantrixflow-infra`. Trusted pull requests
run state-backed plans for staging and production; forks run format/init/validate
without protected credentials.

Review each plan for:

- three staging VPSs and four production VPSs at the expected steady state;
- no replacement or duplicate of an imported persistent VPS;
- only OVH infrastructure resources;
- no Docker application and no per-run simulation instance;
- expected firewall, tag, network, subnet, and SSH-key changes.

After merge to `main`, manually run **Terraform Apply** for one environment.
GitHub environment approval is required. The workflow creates a saved plan,
rejects every delete/replacement action, and applies that exact plan.

Use `bootstrap_hosts=false` for infrastructure-only changes. Use
`bootstrap_hosts=true` only when verified host keys and all required role
WireGuard configs are present.

## 3. Deploy the production ELT image

Merge the ELT change to `main` or `mantrixflow` in
`dabhivijay2478/etl-server-mantrixflow.com`.

The **ELT CI/CD** workflow:

1. installs pinned Python dependencies;
2. validates generated protobuf modules;
3. compiles and tests the ELT implementation;
4. builds a Linux AMD64 image;
5. publishes the immutable GHCR digest;
6. updates only `DOKPLOY_ELT_APPLICATION_ID`;
7. checks the configured health endpoint.

Production ELT runs only on the ELT VPS. Verify private Go-to-ELT gRPC and the
ELT-to-Go callback after deployment. Do not expose port 8001 publicly.

## 4. Deploy the Go API image

Merge the Go API change to `main` or `mantrixflow` in
`dabhivijay2478/main-server-mantrixflow.com`.

The **Go API CI/CD** workflow runs formatting, vet, tests, builds the API and
worker binaries, validates the container, publishes an immutable digest, updates
only `DOKPLOY_GO_APPLICATION_ID`, and polls `GO_API_HEALTH_URL`.

Keep normal Go application deployment independent of simulation-manager
deployment. The Go API may expose simulation REST/SSE while
`SIMULATION_WORKER_ENABLED=false`.

## 5. Publish the simulation runtime

The runtime is built from the ELT repository so simulations execute the same
Python ELT implementation used in production.

Run or trigger **Publish Simulation Runtime**. The workflow builds Linux AMD64,
scans high/critical vulnerabilities, exports a CycloneDX SBOM, signs the digest
through GitHub OIDC, and records provenance. Copy the exact digest shown in the
workflow summary:

```text
SIMULATION_RUNTIME_IMAGE=ghcr.io/OWNER/mantrixflow-simulation-runtime@sha256:...
SIMULATION_RUNTIME_VERSION=sha-GIT_SHA
```

Update the simulation-manager Dokploy environment only after the workflow and
environment approval succeed. A tag without a digest is invalid in production.

## 6. Apply simulation migrations

Run the dedicated migration command before enabling a new simulation schema.
Use a protected direct Supabase URL and keep `DISABLE_AUTO_MIGRATE=true` in both
Go containers.

```bash
cd apps/server/arcyria-server
go run ./cmd/simulation-migrate --dry-run
SIMULATION_MIGRATION_DATABASE_URL="$DIRECT_SECRET_VALUE" \
  go run ./cmd/simulation-migrate --apply
SIMULATION_MIGRATION_DATABASE_URL="$DIRECT_SECRET_VALUE" \
  go run ./cmd/simulation-migrate --verify
```

Do not echo or commit the direct URL. The command applies only ordered additive
simulation migrations and verifies the resulting schema.

## 7. Deploy the simulation manager

Merge a manager-path change or manually run **Simulation Manager CI/CD** in the
Go repository. The workflow tests the simulation packages, publishes
`mantrixflow-simulation-manager`, updates only
`DOKPLOY_SIMULATION_MANAGER_APPLICATION_ID`, and checks readiness.

Deployment requirements:

- exactly one replica on the Go VPS;
- simulation SSH key and dedicated host-CA known-hosts file mounted read-only;
- OVH runtime credentials and exact regional identifiers stored in Dokploy;
- Tigris artifact credentials stored in Dokploy;
- `MAX_SIMULATION_CONCURRENCY=1`;
- `SIMULATION_SANDBOX_BACKEND=microsandbox`;
- immutable runtime digest;
- a working private route from the Go VPS to the Public Cloud simulation subnet.

No dynamic host should exist immediately after manager startup. The first
eligible queue item causes host creation.

## 8. Canary rollout

Use this order:

1. Go API: `SIMULATION_PLATFORM_ENABLED=true`, worker disabled, rollout
   `canary`, and one approved organization ID.
2. Manager: platform and worker enabled, the same canary rollout values.
3. Confirm manager readiness, database/PGMQ connectivity, runtime digest, and
   absence of an unexplained existing dynamic host.
4. Manually run **Production Simulation Canary** and type exactly
   `RUN ON-DEMAND OVH MICROSANDBOX CANARY`.
5. Keep recovery checks enabled.

The canary verifies a completed run, a cancelled run, evidence events, artifact
download/checksum, sandbox destruction, and two serialized runs. After it
passes, wait longer than `SIMULATION_HOST_IDLE_TIMEOUT` and confirm the outer
OVH instance is deleted by reconciliation.

Move to `SIMULATION_ROLLOUT_MODE=general` only after retaining the staging and
production canary evidence.

## 9. Verification

Verify every boundary independently:

```text
Frontend -> Go public HTTPS and SSE
Go -> Supabase database/PGMQ
Go -> ELT private TLS gRPC
ELT -> Go private authenticated callback
Manager -> OVH API
Manager -> temporary host private SSH
Temporary host -> /dev/kvm and pinned Microsandbox
Microsandbox -> signed runtime digest
Manager -> Tigris artifacts
```

Confirm no production connector credential, database URI, authorization header,
or private key appears in evidence, logs, or artifacts.

## 10. Rollback and emergency cleanup

Application rollback is per service: select the previous immutable digest in
the affected Dokploy application and redeploy. Do not roll back Go, ELT, and the
manager together unless their contracts require it.

For simulation rollback:

1. set Go and manager rollout to `disabled`;
2. stop new simulation admission;
3. allow active cleanup to finish;
4. run manager reconciliation;
5. confirm no managed `mantrixflow-sim-*` OVH instance remains;
6. redeploy the previous manager/runtime digest;
7. re-run schema verification; never automatically reverse an additive
   migration containing production metadata.

If the manager cannot clean a paid instance, delete only the exact
prefix-matched simulation instance after matching its provider ID to
`simulation_hosts`. Never delete persistent Terraform-managed VPSs as part of
simulation cleanup.

Dokploy recovery and Tigris restore steps are documented in
[docs/dokploy.md](../../infrastructure/operations/dokploy.md). The future PostgreSQL recovery procedure
is not a Supabase production restore and must not be used before database
migration approval.
