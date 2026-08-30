# OVH, Dokploy, and Microsandbox setup guide

This is the one-time setup guide for the current MantrixFlow architecture.
Vercel Sandbox is not used. Production applications run on persistent OVH VPSs;
the Go simulation manager creates an hourly OVH Public Cloud host only when a
simulation needs Microsandbox.

## Responsibility map

```text
Vercel                 Next.js frontend
Supabase               current PostgreSQL, Auth, RLS, Realtime, and PGMQ
Terraform + Tigris     persistent OVH infrastructure and remote state
Self-hosted Dokploy    deploys immutable Go/ELT/manager images to OVH VPSs
Go simulation manager creates/reconciles/deletes temporary OVH hosts
Microsandbox           one isolated microVM per simulation run
Python runtime image   same production ELT implementation, synthetic endpoints
Tigris                 large simulation evidence and Dokploy backups
```

Terraform must never create a per-run host. Dokploy must never create or manage
a Microsandbox microVM. Production Python ELT must never be routed through the
simulation path.

## 1. Accounts, tools, and domains

Prepare:

- an OVH account with VPS ordering, Public Cloud, IAM tag, network, firewall,
  instance, image, flavor, and SSH-key permissions;
- a pre-existing OVH Public Cloud project for simulation resources;
- Terraform 1.13.5, GitHub CLI, `jq`, Docker, OpenSSH, and WireGuard tooling;
- private, versioned Tigris buckets for Terraform state, simulation artifacts,
  and Dokploy backups;
- DNS names for the frontend, Go API, and Dokploy management interface;
- the three GitHub repositories listed in
  [GitHub repositories and Actions environments](../ci-cd/github-actions.md).

Recommended production DNS:

```text
cloud.mantrixflow.com       -> Vercel frontend
cloud.api.mantrixflow.com   -> Go VPS public edge
deploy.mantrixflow.com      -> Dokploy VPS public edge
```

Do not create a public ELT, database, simulation-manager, SSH, or Microsandbox
control endpoint.

## 2. Create separate SSH identities

Use three independent key pairs:

1. **Infrastructure key** — GitHub protected apply workflow to persistent VPSs.
2. **Dokploy deployment key** — Dokploy to the Go, ELT, and future DB servers.
3. **Simulation client key** — simulation-manager container to temporary hosts.

Only public keys enter Terraform. Store the infrastructure private key only in
the Terraform apply environments. Import the Dokploy private key only into the
self-hosted Dokploy credential store. Mount the simulation private key only into
the simulation-manager container.

The dynamic-host SSH path uses strict host checking. Build the OVH simulation
image so its SSH host certificate is signed by a dedicated host CA. Put only
that CA public key in the manager's dedicated known-hosts file, for example:

```text
@cert-authority * ssh-ed25519 REPLACE_WITH_SIMULATION_HOST_CA_PUBLIC_KEY
```

The wildcard is acceptable only in this dedicated file when that CA signs
simulation hosts exclusively. Never place the host CA private key on a VPS,
inside Dokploy, in GitHub, or in a container image.

## 3. Create Tigris storage

Create private buckets with separate scoped credentials:

```text
mantrixflow-terraform-state
mantrixflow-simulation-artifacts
mantrixflow-dokploy-backups
mantrixflow-database-backups   # future-only
```

Enable object versioning on state and backup buckets. Keep Terraform state,
Dokploy backups, artifacts, and future WAL-G credentials distinct. The default
S3 endpoint used by the repository is `https://fly.storage.tigris.dev`; use the
endpoint assigned to the actual account if different.

The state bucket must exist before `terraform init`. Terraform uses:

```text
mantrixflow/ovh/staging.tfstate
mantrixflow/ovh/production.tfstate
```

## 4. Configure GitHub environments

Follow [github-actions.md](../ci-cd/github-actions.md) and add required reviewers before
uploading values. The infrastructure apply environments must be protected
because a new `ovh_vps` starts a billable service immediately.

After authentication, run:

```bash
cd apps/mantrixflow-infra
./scripts/check-github-environments.sh
```

The checker reads names and workflow status only. It cannot and must not recover
secret values.

## 5. Select OVH products and simulation resources

Resolve account-specific catalog identifiers instead of copying example text:

- VPS-1 plan: 2 vCPU/4 GB class for Dokploy and Go;
- VPS-2 plan: 4 vCPU/8 GB class for ELT and the prepared future database;
- one supported VPS image and datacenter;
- a KVM-capable hourly Public Cloud flavor for the temporary simulation host;
- a regional Public Cloud image that exposes `/dev/kvm`, accepts the simulation
  SSH key, presents the signed host certificate, and permits cloud-init;
- production and staging Public Cloud regions, private networks, and subnets.

Terraform creates the shared simulation region/network/subnet/SSH-key resources.
The Go manager receives their output identifiers and creates the hourly instance
at runtime.

## 6. Plan and create persistent infrastructure

Populate the protected environment names from the checked-in `tfvars` examples.
For an existing server, import it before planning. Never approve a plan that
duplicates an existing server.

Pull requests run format, backend-free initialization, validation, and trusted
state-backed plans. From `main`, manually run **Terraform Apply** with:

```text
environment: staging
bootstrap_hosts: false
```

Review outputs and OVH billing, then repeat for production. Expected persistent
servers are three in staging and four in production. The staging database is
disabled by default.

Verify each persistent VPS SSH fingerprint out of band. Add the exact lines to
`OVH_SSH_KNOWN_HOSTS`, supply the role-specific base64 WireGuard configs, and
rerun the apply workflow with `bootstrap_hosts: true`.

## 7. Configure the static WireGuard overlay

The production address plan is:

| Role | Address |
| --- | --- |
| Dokploy | `10.20.0.10` |
| Go + simulation manager | `10.20.0.20` |
| Python ELT | `10.20.0.30` |
| future database | `10.20.0.40` |

Staging uses `10.21.0.0/24`. Generate a unique WireGuard private key on each
host and create one full `wg0.conf` per role. Encode each complete file into its
matching protected apply secret. The bootstrap workflow installs it as mode
`0600` and enables `wg-quick@wg0`.

Verify from the Go VPS:

```bash
ping -c 2 10.20.0.30
nc -vz 10.20.0.30 8001
```

The database addresses may remain unreachable until the future services are
explicitly enabled.

## 8. Prove the dynamic-host private route

Literal OVH VPSs are not automatically attached to a Public Cloud private
network. The Go VPS at `10.20.0.20` must nevertheless reach the private address
assigned to a temporary host in the production simulation subnet, normally
`10.30.0.0/24`.

Provide one reviewed route before enabling the worker:

- an OVH-supported gateway/router between the VPS overlay and Public Cloud
  private subnet; or
- a WireGuard peer built into the dedicated simulation image/bootstrap.

The current Microsandbox cloud-init installs Microsandbox and UFW; it does not
create that cross-network WireGuard route. Do not fall back to public runtime
ports. `SIMULATION_CONTROL_PLANE_CIDR` must be the actual private source CIDR
seen by the temporary host. Its UFW policy allows runtime ports `20000:40000`
only from that CIDR.

Before production, launch one disposable staging instance with the selected
image, flavor, network, and SSH key. Verify from the Go VPS:

```bash
ssh -i /run/secrets/simulation-host-ssh-key \
  -o BatchMode=yes \
  -o StrictHostKeyChecking=yes \
  -o UserKnownHostsFile=/run/secrets/simulation-host-known-hosts \
  ubuntu@PRIVATE_INSTANCE_IP \
  'test -c /dev/kvm && sudo -u mantrixflow-simulation env HOME=/var/lib/mantrixflow-simulation /var/lib/mantrixflow-simulation/.local/bin/msb doctor'
```

Delete the disposable instance immediately after the test.

## 9. Initialize self-hosted Dokploy

Point the management DNS record at the dedicated Dokploy VPS. Tunnel to the
initial UI because port 3000 is not public:

```bash
ssh -L 3000:127.0.0.1:3000 root@DOKPLOY_PUBLIC_IP
```

Complete administrator creation at `http://127.0.0.1:3000`, configure the
management domain and HTTPS, and create a scoped API key for GitHub Actions.
Import the dedicated Dokploy deployment private key and register remote servers
by their WireGuard addresses.

Create separate applications:

```text
go-api                 -> Go VPS, one or more API replicas within current limits
simulation-manager     -> Go VPS, exactly one replica
python-elt             -> ELT VPS
database/pgbouncer     -> future DB VPS, stopped
```

Configure the Tigris backup destination and immediately test a full Dokploy
backup. Keep normal application workloads off the Dokploy VPS.

## 10. Configure production applications

Use the checked-in environment examples as name references; replace every
placeholder in Dokploy and never commit populated files.

The Go API uses the Go private endpoint for ELT:

```text
PORT=5000
ELT_PYTHON_SERVICE_URL=http://10.20.0.30:8000
ELT_GRPC_ADDRESS=10.20.0.30:8001
SIMULATION_PLATFORM_ENABLED=false
SIMULATION_WORKER_ENABLED=false
```

The Python ELT callback returns privately to Go:

```text
CALLBACK_URL=http://10.20.0.20:5000/api/v1/internal/elt-callback
ELT_GRPC_LISTEN_ADDRESS=0.0.0.0:8001
```

Bind Go port 5000 and ELT ports 8000/8001 to their WireGuard addresses in
Dokploy, matching the checked-in Compose definitions. Do not bind those internal
ports to `0.0.0.0`. The Go host firewall admits private callbacks and manager
health only from the selected WireGuard CIDR.

Use identical `ELT_INTERNAL_TOKEN` values at the Go/ELT boundary and identical
`CALLBACK_TOKEN` values for ELT callbacks. Mount TLS certificates as read-only
files. Supabase remains the production database and Auth provider.

## 11. Prepare the simulation-manager application

Start from `environments/production/simulation-manager.env.example`. Populate
the shared-network outputs after Terraform apply:

```bash
terraform -chdir=infra/terraform/environments/production output -json simulation_shared
```

Set `OVH_SIMULATION_REGION`, flavor/image IDs, SSH key ID, and private-network
ID to the exact same regional resources. `OVH_ENDPOINT` here is the HTTPS OVH
API base URL; it is different from Terraform's provider endpoint name.

Mount the simulation SSH files into the container:

```text
/run/secrets/simulation-host-ssh-key
/run/secrets/simulation-host-known-hosts
```

Because the image runs as UID/GID `65532`, ensure those read-only mounts are
readable by that identity without making the private key world-readable.

Production requires Tigris artifacts, an immutable runtime digest, exactly one
manager replica, and `MAX_SIMULATION_CONCURRENCY=1`. Use a random service secret:

```bash
openssl rand -hex 32
```

Store its output only as `SIMULATION_GRPC_TOKEN_SECRET` in Dokploy. The runtime
image must be pullable from the temporary host. For the initial implementation,
make the dedicated GHCR runtime package public or configure registry access on
the base image without placing a token in cloud-init or command arguments.

## 12. Apply and verify simulation schema

Production application processes keep `DISABLE_AUTO_MIGRATE=true`. Apply the
additive simulation migrations as a separate audited operation using the direct
Supabase connection, never the transaction pooler:

```bash
cd apps/server/main-server
go run ./cmd/simulation-migrate --dry-run
SIMULATION_MIGRATION_DATABASE_URL='REPLACE_DIRECT_URL' \
  go run ./cmd/simulation-migrate --apply
SIMULATION_MIGRATION_DATABASE_URL='REPLACE_DIRECT_URL' \
  go run ./cmd/simulation-migrate --verify
```

Do not place the direct URL in shell history in real operations; inject it from
a protected secret manager. This adds simulation metadata and PGMQ structures
only. It does not migrate Supabase data to the prepared OVH database.

## 13. Publish and roll out Microsandbox

1. Run **Publish Simulation Runtime** in the ELT repository.
2. Record the signed `ghcr.io/...@sha256:...` value from the workflow summary.
3. Put that digest and matching Git SHA version into the simulation-manager
   Dokploy environment.
4. Deploy the manager while API rollout remains disabled.
5. Verify `/readyz` through the approved private or authenticated health path.
6. Set the Go API and manager to `SIMULATION_ROLLOUT_MODE=canary`, add one
   organization ID, and enable the platform. Only the manager sets
   `SIMULATION_WORKER_ENABLED=true`.
7. Run the protected **Production Simulation Canary** with the exact confirmation
   phrase from the workflow.
8. Confirm completion, cancellation, evidence checksum, sandbox destruction,
   serialized recovery runs, and eventual outer-host deletion after idle TTL.
9. Move to `general` only after staging and canary evidence is retained.

The repeatable deployment and rollback sequence is in
[DEPLOYMENT.md](../../deployment/infrastructure/ovh-microsandbox-runbook.md).
