# Tigris Storage Setup Guide

Tigris is the private S3-compatible object store for MantrixFlow. It stores
large simulation artifacts now and will store WAL-G database backups after the
future OVHcloud PostgreSQL cutover. Active ELT DuckDB staging remains on the
dedicated ELT VPS at `/var/lib/mantrixflow/staging`.

## Buckets and access

Create private, purpose-separated buckets:

```text
mantrixflow-terraform-state
mantrixflow-dokploy-backups
mantrixflow-simulation-artifacts
mantrixflow-database-backups
```

The database-backup bucket is future-only while Supabase remains active. Do
not make any bucket public. Create separate scoped access keys for Terraform
state, Dokploy recovery, simulation artifacts, and the future database backup
process; never reuse production connector credentials.

Use the S3-compatible endpoint:

```text
https://fly.storage.tigris.dev
```

Configure the simulation manager in self-hosted Dokploy with:

```text
SIMULATION_ARTIFACT_ENDPOINT=https://fly.storage.tigris.dev
SIMULATION_ARTIFACT_REGION=auto
SIMULATION_ARTIFACT_BUCKET=mantrixflow-simulation-artifacts
SIMULATION_ARTIFACT_ACCESS_KEY_ID=<scoped key>
SIMULATION_ARTIFACT_SECRET_ACCESS_KEY=<scoped secret>
```

Terraform state uses the S3 backend with separate production and staging keys,
S3 lockfiles, and protected GitHub environment credentials. Configure Dokploy's
built-in Web Server backup with its dedicated bucket so `/etc/dokploy` and the
control-plane PostgreSQL database can be recovered independently.

Configure future WAL-G values only when the self-hosted database migration is
explicitly approved. The prepared variable names and scripts are documented in
[`../apps/mantrixflow-infra/backup/README.md`](../apps/mantrixflow-infra/backup/README.md).

## Verification

For simulation artifacts:

1. Run an approved canary simulation.
2. Confirm PostgreSQL contains only artifact metadata and object references.
3. Confirm the referenced private object exists in the simulation bucket.
4. Confirm an unauthenticated request cannot read it.

For the future database service:

1. Run a full backup from the isolated database VPS.
2. Confirm WAL/full-backup objects appear under the configured prefix.
3. Restore into a disposable PostgreSQL instance.
4. Validate schema, PGMQ queues, and representative application reads.

Snapshots are additional protection; they do not replace database-level
backups or restore drills.

## Retention and rotation

- Apply lifecycle rules matching the evidence and backup retention policy.
- Do not persist active DuckDB working files directly in object storage.
- Rotate one scoped key at a time, update the corresponding self-hosted Dokploy
  application, verify access, and only then revoke the previous key.
- Never expose Tigris secrets to the frontend or bake them into an image.
