# Tigris Storage Setup Guide

This guide sets up Tigris for the Hetzner production deployment.

Tigris is used for:

1. Terraform/OpenTofu remote state for the infra repository.
2. Hourly backup uploads from the Hetzner server.

Tigris is not the active DuckDB staging disk. Active ELT staging stays on the
Hetzner server SSD at `/var/mantrixflow/staging`.

## 1. Create A Tigris Account

1. Open [Tigris](https://www.tigrisdata.com/).
2. Sign in or create an account.
3. Open the Tigris dashboard.

## 2. Create The Bucket

Create one bucket for MantrixFlow production:

```text
mantrixflow-production
```

Recommended settings:

| Setting | Value |
| --- | --- |
| Bucket name | `mantrixflow-production` |
| Storage class | Standard |
| Region/distribution | Tigris default/global |
| Public access | Private |

Keep this bucket private. GitHub Actions and the Hetzner server access it with
S3-compatible access keys.

## 3. Create Access Keys

In the Tigris dashboard:

1. Go to **Access Keys**.
2. Create a new key.
3. Name it:

```text
mantrixflow-production-ci
```

4. Copy both values immediately:

```text
Access key ID
Secret access key
```

The secret access key is shown once. Store it directly in GitHub secrets.

## 4. Endpoint And Bucket Values

Use these values for GitHub:

```text
TIGRIS_BUCKET=mantrixflow-production
TIGRIS_ENDPOINT=https://fly.storage.tigris.dev
```

Then add the generated key values:

```text
TIGRIS_ACCESS_KEY_ID=your_tigris_access_key_id
TIGRIS_SECRET_ACCESS_KEY=your_tigris_secret_access_key
```

## 5. Add Infra Repo Secrets

Repository:

```text
dabhivijay2478/mantrixflow-infra
```

Environment:

```text
production-hetzner
```

Branch rule:

```text
main
```

Add these environment secrets:

| Secret | Value |
| --- | --- |
| `TIGRIS_ACCESS_KEY_ID` | Tigris access key ID |
| `TIGRIS_SECRET_ACCESS_KEY` | Tigris secret access key |
| `TIGRIS_ENDPOINT` | `https://fly.storage.tigris.dev` |
| `TIGRIS_BUCKET` | `mantrixflow-production` |

Do not add Tigris secrets to the API or ELT repositories. Only the infra
workflow needs them.

## 6. What The Infra Workflow Does

On push to `main`, the infra workflow uses Tigris as the S3-compatible backend
for Terraform/OpenTofu state.

It also writes Tigris credentials into the server bootstrap flow so the server
can upload hourly backups from:

```text
/var/mantrixflow/staging
```

The backup timer to check after deployment is:

```bash
systemctl status mantrixflow-tigris-backup.timer
```

## 7. Verify From GitHub Actions

After the first infra deployment succeeds, confirm the Tigris bucket contains a
state object. In Tigris, open:

```text
mantrixflow-production
```

You should see Terraform/OpenTofu state files or prefixes created by the infra
workflow.

## 8. Verify From The Hetzner Server

SSH into the server:

```bash
ssh -i ~/.ssh/mantrixflow_hetzner root@SERVER_IPV4
```

Check the backup timer:

```bash
systemctl status mantrixflow-tigris-backup.timer
```

Run one manual backup check:

```bash
systemctl start mantrixflow-tigris-backup.service
journalctl -u mantrixflow-tigris-backup.service -n 100 --no-pager
```

Then confirm a backup object appears in the Tigris bucket.

## 9. Cost Guardrails

For MVP, keep Tigris small:

1. Terraform state is tiny.
2. Hourly backups should only include required staging backup data.
3. Do not store active DuckDB working files directly in Tigris.
4. Periodically delete old backups or add lifecycle rules when available.

The first 5 GB of Tigris standard storage is free. Above that, storage is billed
per GB, so keep staging backup retention conservative.

## 10. Secret Rotation

To rotate Tigris keys:

1. Create a new Tigris access key.
2. Update `TIGRIS_ACCESS_KEY_ID` and `TIGRIS_SECRET_ACCESS_KEY` in the infra
   repo `production-hetzner` environment.
3. Run the infra workflow from `main`.
4. Confirm Terraform state still loads successfully.
5. Confirm the server backup timer still uploads.
6. Delete the old Tigris access key.

Do not delete the old key before the new infra workflow has completed
successfully.
