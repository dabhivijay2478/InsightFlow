# Simulation implementation report

## Reused

The durable PGMQ queue, versioned simulation protobufs, gRPC runtime, existing
Python ELT, deterministic Postgres world, evaluator, evidence persistence,
artifact storage, cleanup paths, and SSE browser DTOs remain the product core.

## Changed

The rootless Docker/shared-host manager was removed. The worker is now a
lightweight control service that signs OVH API calls, persists outer-host state,
bootstraps a pinned Microsandbox installation, creates one resource-bounded
microVM, and reconciles idle or missing infrastructure. No host is created at
manager startup.

## Added

- Additive `simulation_hosts` migrations, including OVH region/flavor,
  `BOOTSTRAPPING`, capacity, IP, and bounded failure metadata.
- OVH provider, host state store, lifecycle manager, and reconciliation.
- Strict SSH transport and checksum-pinned cloud-init.
- Per-run TLS/token generation and Microsandbox runtime manager.
- Independent manager Dockerfile, GHCR workflow, self-hosted Dokploy deployment, and
  on-demand canary.

## Compatibility and risk

Supabase remains the metadata database and Auth provider. Production requires
a pre-created OVH private network with a gateway, a KVM-capable flavor, and a
host-CA-enabled image. OVH API schema/flavor availability and nested KVM must be
proved in staging. Exactly one manager replica is supported in the MVP. Runtime
image promotion should use a digest after signature and vulnerability checks.

The one-time operator procedure, including the required Go-VPS-to-Public-Cloud
private route and dynamic SSH host-CA trust, is maintained in
[`../../../mantrixflow-infra/docs/setup-guide.md`](../../infrastructure/setup/ovh-dokploy-microsandbox.md).
