# Oracle Cloud Production Deployment

This guide deploys MantrixFlow to two Oracle Cloud ARM VMs using a strict
zero-cost infrastructure profile. Terraform owns the
infrastructure; the API and ELT repositories independently publish immutable
GHCR images and deploy them through OCI Run Command.

The repository intentionally refuses larger compute or storage values. This
prevents a normal Terraform apply from silently creating OCI resources above
the Always Free allocation. Infrastructure CI creates a saved Terraform plan
and applies that exact plan only after the zero-cost static guard passes.

## Zero-Cost Boundary

This setup uses only the following Oracle Always Free allocations:

| Resource | Configured usage | Always Free boundary |
| --- | ---: | ---: |
| Ampere A1 compute | 2 OCPUs / 12 GB total | 2 OCPUs / 12 GB total |
| Block storage | 190 GB total | 200 GB total |
| Network Load Balancer | 1 | 1 |
| VCN | 1 | Up to 2 for Free Tier tenancies |
| Object Storage state | Tiny Terraform state bucket | Stay below 20 GB and 50,000 requests/month |
| Vault | Default vault, software key, 3 secrets | Software keys and up to 150 secrets |
| Outbound transfer | Application traffic | Stay below 10 TB/month |

Cloudflare DNS-only records and Caddy/Let's Encrypt are free. No NAT Gateway,
paid database, paid load balancer, Kubernetes cluster, paid logging sink,
private Vault, or paid support product is provisioned.

Important limitations:

- Always Free resources must be created in the OCI tenancy's **home region**.
  This guide uses `ap-mumbai-1`; do not continue if Mumbai is not the home
  region.
- The Terraform profile reserves 10 GB of the 200 GB block-storage allocation.
  It still assumes the tenancy has no other boot or block volumes consuming the
  allowance.
- A Free Tier account is safer than a Pay As You Go account because paid
  resources cannot continue after trial credits expire. Do not upgrade the
  tenancy for this setup.
- GitHub Actions and GHCR are outside OCI. Use public repositories/packages or
  set GitHub Actions and Packages budgets/spending limits to zero. Private
  repositories can consume included minutes/storage and then incur charges.
- The domain registration itself is not free. This guide assumes
  `mantrixflow.com` is already owned.
- Supabase, email, Anthropic, PostHog, Slack, and payment-provider usage are
  application integrations, not provisioned infrastructure. Keep them on their
  free plans or leave optional API keys unset.

## Ownership

| Repository | Production responsibility |
| --- | --- |
| `dabhivijay2478/mantrixflow-infra` | OCI networking, VMs, NLB, Vault, Cloudflare DNS |
| `dabhivijay2478/cloud.api.mantrixflow.com` | Go API image and API deployment |
| `dabhivijay2478/cloud.api.etl.server.mantrixflow.com` | ELT image and ELT deployment |

Development continues on `mantrixflow-contabo`. Open a pull request into
`mantrixflow-oracle`; only a merge or direct push to `mantrixflow-oracle`
deploys production.

## Architecture

- API VM: Oracle Linux 9 ARM, 1 OCPU, 4 GB RAM, 50 GB boot volume.
- ELT VM: Oracle Linux 9 ARM, 1 OCPU, 8 GB RAM, 50 GB boot plus 90 GB staging volume.
- OCI Network Load Balancer: public TCP 80/443 to Caddy on the API VM.
- API private port: `8080`, allowed only from the ELT NSG.
- ELT private port: `8000`, allowed only from the API NSG.
- No public SSH, Dokploy, port 3000, API internal port, or ELT port.
- Cloudflare record `cloud.api.mantrixflow.com` is DNS-only. Caddy obtains TLS.

The database-backed deployment lease serializes deployments across the two
service repositories. New runs queue during a deployment, active runs drain,
and expired leases automatically stop blocking dispatch.

## One-Time OCI Bootstrap

1. Create or keep an Oracle **Free Tier** tenancy. Do not upgrade to Pay As You
   Go. Confirm `ap-mumbai-1` is the tenancy home region.
2. In OCI Console, open **Governance & Administration → Limits, Quotas and
   Usage** and confirm these resources are unused or available:

   ```text
   Ampere A1:             at least 2 OCPUs and 12 GB memory
   Block/boot volumes:    at least 190 GB
   Network Load Balancer: at least 1
   Object Storage:        enough capacity below the 20 GB free total
   ```

3. Create an OCI API user/group with permissions to manage resources in the
   tenancy, invoke instance-agent commands, and update Vault secret versions.
4. Create an API signing key and record the user OCID, tenancy OCID,
   fingerprint, and private key.
5. Create one OCI Object Storage bucket with versioning enabled. Use it only for
   Terraform state. Configure a lifecycle rule to delete non-current state
   versions after 30 days, and keep the bucket below 1 GB.
6. Create an Object Storage customer secret key for Terraform's S3-compatible
   backend.
7. Find the Mumbai availability-domain name that has A1 capacity.
8. Create a Cloudflare Free-plan API token with DNS edit permission for
   `mantrixflow.com`.
9. In GitHub billing settings, configure Actions and Packages budgets to prevent
   additional paid usage. Make GHCR container packages public when acceptable.

The state bucket is intentionally bootstrapped outside Terraform because
Terraform cannot store its own state in a bucket before that bucket exists.

## GitHub Environments

Create a protected `production-oracle` environment in all three repositories.
Require production approval when appropriate and disallow unprotected branches.

Infrastructure repository secrets:

```text
OCI_TENANCY_OCID
OCI_USER_OCID
OCI_FINGERPRINT
OCI_API_PRIVATE_KEY
OCI_AVAILABILITY_DOMAIN
OCI_STATE_BUCKET
OCI_OBJECT_STORAGE_ACCESS_KEY
OCI_OBJECT_STORAGE_SECRET_KEY
OCI_OBJECT_STORAGE_ENDPOINT
CLOUDFLARE_API_TOKEN
CLOUDFLARE_ZONE_ID
ORACLE_API_ENV
ORACLE_ELT_ENV
GHCR_READ_TOKEN
```

Infrastructure repository variable:

```text
ORACLE_ENABLE_DNS_CUTOVER=false
```

API and ELT repository secrets:

```text
OCI_TENANCY_OCID
OCI_USER_OCID
OCI_FINGERPRINT
OCI_API_PRIVATE_KEY
OCI_COMPARTMENT_OCID
INTERNAL_TOKEN
```

`INTERNAL_TOKEN` must match the API runtime environment. The GHCR token needs
read access to both private images.

## Runtime Environment Bundles

Store newline-delimited environment files in the infrastructure repository
secrets `ORACLE_API_ENV` and `ORACLE_ELT_ENV`. The workflow uploads them to OCI
Vault; Terraform state and OCI Run Command payloads never contain their values.

Required API production values include:

```dotenv
PORT=8080
ENVIRONMENT=production
ELT_PYTHON_SERVICE_URL=http://<ELT_PRIVATE_IP>:8000
API_PUBLIC_URL=https://cloud.api.mantrixflow.com
PIPELINE_MAX_CONCURRENT=2
PGMQ_PARALLEL_WORKERS=2
PIPELINE_MAX_PER_ORG_CONCURRENT=1
PIPELINE_MAX_PER_SOURCE_CONCURRENT=1
PIPELINE_MAX_PER_HOUR=120
PIPELINE_MAX_PER_DAY=2000
```

Also include database, Supabase, encryption, callback, internal, Dodo, Slack,
GitHub App, PostHog, and email variables used by the application.

Required ELT values:

```dotenv
PORT=8000
ENVIRONMENT=production
LOG_LEVEL=INFO
CALLBACK_URL=http://<API_PRIVATE_IP>:8080/api/v1/internal/elt-callback
MAX_CONCURRENT_RUNS=2
STAGING_ROOT=/var/mantrixflow/staging
STAGING_DISK_LIMIT_GB=70
```

Also include matching `ENCRYPTION_KEY`, `ELT_INTERNAL_TOKEN`, and
`CALLBACK_TOKEN`.

After the first infrastructure apply, obtain private IPs from Terraform output,
update the two environment bundles, and rerun the infrastructure workflow.

## First Deployment

1. Merge the infrastructure repository into `mantrixflow-oracle` while
   `ORACLE_ENABLE_DNS_CUTOVER=false`.
2. Confirm Terraform creates the compartment, VCN, VMs, Vault, NLB, NSGs,
   staging volume, and DNS record.
3. Update the Vault environment bundles with Terraform's private IP outputs.
4. Merge ELT into `mantrixflow-oracle` and wait for its image deployment.
5. Merge API into `mantrixflow-oracle`.
6. Set `ORACLE_ENABLE_DNS_CUTOVER=true` and rerun the infrastructure workflow.
   Terraform now switches only `cloud.api.mantrixflow.com` to the Oracle NLB.
7. Confirm:

```bash
curl -fsS https://cloud.api.mantrixflow.com/health
curl -fsS https://cloud.api.mantrixflow.com/api/v1/health | jq .
```

The detailed health response must report both `database` and `elt_server` as
`operational`.

## Rollback

Run the relevant service workflow manually and provide a previously published
commit SHA as `image_tag`. The workflow drains pipelines, deploys the immutable
image, verifies health, and releases the lease.

Never retag or overwrite a commit SHA image.

## Verification

- Confirm `scripts/check-zero-cost-profile.sh` passes in infrastructure CI.
- Before every apply, review the Terraform plan and reject any resource not
  listed in the Zero-Cost Boundary table.
- Recheck OCI **Limits, Quotas and Usage** after apply. Total A1 usage must be
  2 OCPUs/12 GB and total boot/block storage must be no more than 190 GB.
- Confirm the key uses `SOFTWARE` protection and the vault type is `DEFAULT`,
  never `VIRTUAL_PRIVATE`.
- Confirm only one Network Load Balancer exists in the tenancy.
- Confirm the Object Storage state bucket remains below 1 GB and GHCR/GitHub
  spending budgets remain at zero.
- Confirm public connection attempts to ports `22`, `3000`, `8000`, and `8080`
  fail.
- Submit 50 test runs and verify no more than two are `running`; the rest
  must remain queued.
- Deploy while long-running pipelines are active and verify replacement waits.
- Kill the ELT container during a controlled run and verify the API requeues it,
  preserves the old checkpoint, and stops after three unsuccessful recoveries.
- Verify the ELT staging volume is mounted and owned by UID/GID `1001`.
- Run Terraform twice and confirm the second plan is empty.

## Disaster Recovery

- Terraform state bucket versioning protects infrastructure history.
- OCI Vault contains runtime environment bundles.
- GHCR commit-SHA images provide immutable rollback artifacts.
- Recreate VMs with Terraform, rerun the infrastructure workflow to install the
  deployment command, then manually dispatch the ELT and API workflows using
  the desired image SHAs.

Oracle Always Free A1 capacity is not guaranteed. Do not destroy healthy VMs
during an incident unless replacement capacity has been confirmed.

## Cost Incident Procedure

If OCI shows any forecasted or actual charge:

1. Disable merges to `mantrixflow-oracle`.
2. Set `ORACLE_ENABLE_DNS_CUTOVER=false`.
3. Inspect **Billing & Cost Management → Cost Analysis** by service.
4. Compare tenancy usage against the Zero-Cost Boundary table.
5. Delete only the unexpected paid resource. Do not destroy the two healthy A1
   VMs while investigating.
6. Keep the account as Free Tier and do not accept an upgrade prompt.

## Official Free-Tier References

- [Oracle Always Free resources](https://docs.oracle.com/en-us/iaas/Content/FreeTier/freetier_topic-Always_Free_Resources.htm)
- [Oracle Free Tier account behavior](https://docs.oracle.com/en-us/iaas/Content/FreeTier/freetier.htm)
- [Oracle Cloud Free Tier overview](https://www.oracle.com/cloud/free/)
- [Cloudflare Free plan](https://www.cloudflare.com/plans/free/)

Always verify these pages before increasing any resource because provider limits
and free-tier terms can change.
