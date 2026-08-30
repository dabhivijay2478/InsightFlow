# Self-hosted Dokploy operations

Dokploy runs only on its dedicated OVH VPS-1. Go, ELT, and the future database
are independent remote deployment servers connected with the dedicated Dokploy
SSH key. Do not place MantrixFlow workloads on the control-plane VPS.

## Installation and access

`scripts/bootstrap-dokploy.sh` downloads the official installer, verifies the
operator-supplied SHA-256, and installs a pinned `DOKPLOY_RELEASE_TAG`. Port 3000
is not public; complete initial setup through an SSH tunnel. Point the management
domain to the VPS, enable HTTPS, strong authentication, and 2FA when available.

Register each remote server using its WireGuard address where reachable. Import
the dedicated deployment private key into Dokploy; never use a personal developer
key. Configure GHCR read credentials and make applications pull immutable image
digests built by GitHub Actions.

Protected application environments require:

```text
DOKPLOY_URL=https://deploy.mantrixflow.com
DOKPLOY_API_KEY
DOKPLOY_GO_APPLICATION_ID
DOKPLOY_ELT_APPLICATION_ID
DOKPLOY_SIMULATION_MANAGER_APPLICATION_ID
GHCR_USERNAME
GHCR_READ_TOKEN
GO_API_HEALTH_URL
ELT_HEALTH_URL
SIMULATION_MANAGER_HEALTH_URL
```

The exact repository/environment ownership and a read-only checker are in
[github-actions.md](../ci-cd/github-actions.md). Runtime application values belong in
Dokploy and are listed in [setup-guide.md](../setup/ovh-dokploy-microsandbox.md); do not duplicate
production database or OVH runtime credentials into GitHub deployment environments.

The dynamic OVH simulation host is absent from Dokploy. The Go manager owns its
creation, Microsandbox bootstrap, health, idle accounting, and deletion.

## Tigris backup and recovery

Create a Tigris S3 destination in Dokploy with credentials separate from
Terraform state. Schedule the built-in full control-plane backup; it includes the
`dokploy-postgres` database and `/etc/dokploy`. Test the destination immediately.

Recovery procedure:

1. Recreate or repair the dedicated Dokploy VPS and rerun the pinned installer.
2. Configure the Tigris destination without committing credentials.
3. Restore the selected full backup from Dokploy's Web Server backup screen.
4. Update the server IP and management DNS if they changed.
5. Revalidate GHCR access, remote SSH connectivity, application definitions, and TLS.
6. Trigger a no-change deployment and readiness check for each application.

Existing containers on remote servers keep running while the Dokploy control plane
is unavailable. Recovery restores deployment management; it is not an application
or Supabase database restore.
