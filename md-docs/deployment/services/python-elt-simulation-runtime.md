# Python ELT and simulation runtime deployment

Production Python ELT runs on its dedicated OVH VPS-2 and is deployed by the
self-hosted Dokploy control plane. The simulation runtime is a separate immutable
image built from this same repository so simulation and production execute the
same ELT implementation.

The complete operator guides are:

- [`../../mantrixflow-infra/docs/setup-guide.md`](../../infrastructure/setup/ovh-dokploy-microsandbox.md)
- [`../../mantrixflow-infra/DEPLOYMENT.md`](../infrastructure/ovh-microsandbox-runbook.md)
- [`../../mantrixflow-infra/docs/github-actions.md`](../../infrastructure/ci-cd/github-actions.md)

## Production ELT

The `ELT CI/CD` workflow runs on `main` and `mantrixflow`, verifies protobuf
modules, compiles and tests the service, publishes an immutable Linux AMD64 GHCR
digest, updates only `DOKPLOY_ELT_APPLICATION_ID`, and checks the configured
health endpoint.

Configure the `production-elt` GitHub environment with the names documented in
the GitHub Actions guide. Runtime credentials belong in the Python ELT Dokploy
application, not in the deployment workflow.

The cross-host production addresses are:

```text
Go -> ELT:       10.20.0.30:8001 (TLS gRPC)
ELT -> Go:       http://10.20.0.20:5000/api/v1/internal/elt-callback
ELT staging:     /var/lib/mantrixflow/staging
```

Do not expose gRPC port 8001 publicly. Keep `ELT_INTERNAL_TOKEN`, callback token,
TLS key, Supabase service role, and encryption key out of image layers and logs.

## Simulation runtime

`Publish Simulation Runtime` packages the existing ELT implementation with the
deterministic Postgres world, evaluator, and evidence runtime. It does not deploy
or replace the production ELT application.

The workflow publishes:

```text
ghcr.io/OWNER/mantrixflow-simulation-runtime@sha256:...
```

It also scans the image, exports an SBOM, signs the digest with GitHub OIDC, and
creates provenance. Copy only the immutable digest and Git SHA version into the
simulation-manager Dokploy environment.

Rollback selects the preceding digest for only the affected unit. A production
ELT rollback does not roll back the simulation runtime, and a runtime rollback
does not redeploy production ELT.
