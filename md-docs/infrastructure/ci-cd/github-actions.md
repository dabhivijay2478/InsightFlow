# GitHub repositories and Actions environments

This guide matches the workflow files currently present in the three backend
repositories. GitHub secret values cannot be read after creation; the included
checker compares names only and never prints secret contents.

## Repository and branch map

| Repository | Local path | Deployment branches |
| --- | --- | --- |
| `dabhivijay2478/mantrixflow-infra` | `apps/mantrixflow-infra` | `main` |
| `dabhivijay2478/main-server-mantrixflow.com` | `apps/server/main-server` | `main`, `mantrixflow` |
| `dabhivijay2478/etl-server-mantrixflow.com` | `apps/server/elt-server` | `main`, `mantrixflow` |

The repositories and branches above were verified through their configured Git
remotes. Authenticate GitHub CLI before checking protected configuration:

```bash
gh auth login -h github.com
gh auth refresh -h github.com -s repo,workflow
gh auth status -h github.com
```

Run the read-only audit from the infrastructure repository:

```bash
./scripts/check-github-environments.sh
```

The script verifies repository access, every environment below, required
variable and secret names, and the latest eight workflow runs. It exits without
changing GitHub. Override repository names with `GITHUB_OWNER`,
`INFRA_GITHUB_REPOSITORY`, `GO_GITHUB_REPOSITORY`, or
`ELT_GITHUB_REPOSITORY` when testing a fork.

## Environment protection

Create these deployment environments in GitHub repository settings before
adding values:

```text
mantrixflow-infra:
  terraform-staging-plan
  terraform-production-plan
  terraform-staging
  terraform-production

main-server-mantrixflow.com:
  production-go-api
  production-simulation-manager
  production-simulation-canary

etl-server-mantrixflow.com:
  production-elt
  production-simulation-runtime
```

Require approval on `terraform-staging`, `terraform-production`, and every
`production-*` environment. Restrict production environments to their intended
deployment branches. The two `*-plan` environments may omit approval, but must
still restrict access to trusted pull requests because they expose credentials.
Fork pull requests intentionally receive static Terraform checks only.

An environment can be created without values using:

```bash
gh api --method PUT \
  repos/dabhivijay2478/mantrixflow-infra/environments/terraform-production
```

Configure reviewers and branch protection in the GitHub UI so human identities
and team IDs are selected explicitly.

## Terraform environment variables

Add the following as GitHub **environment variables** to all four Terraform
environments. Staging and production values must remain separate.

| Name | Meaning |
| --- | --- |
| `OVH_ENDPOINT` | OVH provider endpoint name, normally `ovh-eu` |
| `OVH_SUBSIDIARY` | OVH account subsidiary |
| `OVH_PROJECT_ID` | Public Cloud project used for shared simulation resources |
| `OVH_VPS_DATACENTER` | Literal VPS datacenter code |
| `OVH_VPS_IMAGE_ID` | VPS image identifier |
| `OVH_VPS1_PLAN_CODE` | Account-specific 2 vCPU/4 GB plan code |
| `OVH_VPS2_PLAN_CODE` | Account-specific 4 vCPU/8 GB plan code |
| `SSH_ALLOWED_CIDRS_JSON` | JSON list such as `["203.0.113.10/32"]` |
| `SERVER_NAMES_JSON` | JSON object containing `dokploy`, `go`, `elt`, `database` |
| `OVH_TAGS_JSON` | JSON object containing non-secret resource tags |
| `DOKPLOY_DOMAIN` | `deploy.mantrixflow.com` or the staging equivalent |
| `OVH_SIMULATION_REGION` | Region for temporary simulation hosts |
| `OVH_SIMULATION_NETWORK_NAME` | Shared private-network name |
| `OVH_SIMULATION_NETWORK_CIDR` | Production `10.30.0.0/24`; staging `10.31.0.0/24` by default |
| `OVH_SIMULATION_SSH_KEY_NAME` | Project SSH-key resource name |

Add `DOKPLOY_RELEASE_TAG` only to the two apply environments. It must be a
pinned release compatible with the verified installer checksum.

Set a variable interactively without placing it in shell history:

```bash
gh variable set OVH_PROJECT_ID \
  --repo dabhivijay2478/mantrixflow-infra \
  --env terraform-production
```

## Terraform environment secrets

All four Terraform environments require:

```text
OVH_APPLICATION_KEY
OVH_APPLICATION_SECRET
OVH_CONSUMER_KEY
TIGRIS_ACCESS_KEY_ID
TIGRIS_SECRET_ACCESS_KEY
TIGRIS_STATE_BUCKET
TIGRIS_ENDPOINT
INFRASTRUCTURE_SSH_PUBLIC_KEY
DOKPLOY_DEPLOY_PUBLIC_KEY
SIMULATION_SSH_PUBLIC_KEY
```

The two apply environments additionally require:

```text
INFRASTRUCTURE_SSH_PRIVATE_KEY
OVH_SSH_KNOWN_HOSTS
DOKPLOY_INSTALLER_SHA256
WIREGUARD_DOKPLOY_CONFIG_B64
WIREGUARD_GO_CONFIG_B64
WIREGUARD_ELT_CONFIG_B64
WIREGUARD_DATABASE_CONFIG_B64
```

Keep production and staging state credentials separate. `OVH_SSH_KNOWN_HOSTS`
must contain operator-verified persistent VPS host keys, not output captured
blindly with `ssh-keyscan`.

Set a secret interactively:

```bash
gh secret set OVH_APPLICATION_SECRET \
  --repo dabhivijay2478/mantrixflow-infra \
  --env terraform-production
```

## Application deployment environments

`production-go-api` in the Go repository requires:

```text
Variable: DOKPLOY_URL
Secrets:  DOKPLOY_API_KEY
          DOKPLOY_GO_APPLICATION_ID
          GHCR_USERNAME
          GHCR_READ_TOKEN
          GO_API_HEALTH_URL
```

`production-simulation-manager` in the Go repository requires:

```text
Variable: DOKPLOY_URL
Secrets:  DOKPLOY_API_KEY
          DOKPLOY_SIMULATION_MANAGER_APPLICATION_ID
          GHCR_USERNAME
          GHCR_READ_TOKEN
          SIMULATION_MANAGER_HEALTH_URL
```

`production-simulation-canary` in the Go repository requires:

```text
SIMULATION_CANARY_API_URL
SIMULATION_CANARY_AUTH_TOKEN
SIMULATION_CANARY_ORGANIZATION_ID
SIMULATION_CANARY_SAVED_TEST_ID
```

`production-elt` in the ELT repository requires:

```text
Variable: DOKPLOY_URL
Secrets:  DOKPLOY_API_KEY
          DOKPLOY_ELT_APPLICATION_ID
          GHCR_USERNAME
          GHCR_READ_TOKEN
          ELT_HEALTH_URL
```

`production-simulation-runtime` currently needs no custom secret. Its workflow
uses the automatic `GITHUB_TOKEN` and GitHub OIDC to publish, scan, sign, and
attest the runtime image. Keep the environment so approval can protect runtime
promotion.

`DOKPLOY_URL` must be the self-hosted HTTPS origin, never a hosted-control-plane
URL. Use a classic token or fine-grained token with `read:packages` for
`GHCR_READ_TOKEN`; scope it only to the packages Dokploy pulls. Health URLs must
not expose ELT gRPC, the simulation manager, or Microsandbox control ports.

## Checks and troubleshooting

```bash
# Show names only; GitHub never returns stored secret values.
gh variable list --repo OWNER/REPO --env ENVIRONMENT --json name
gh secret list --repo OWNER/REPO --env ENVIRONMENT --json name

# Inspect recent runs.
gh run list --repo OWNER/REPO --limit 20
gh run view RUN_ID --repo OWNER/REPO --log-failed
```

If the checker reports authentication failure, repair `gh` authentication
before diagnosing missing environments. If a deployment waits indefinitely,
check environment reviewers and branch restrictions. If a pull-request plan is
skipped, confirm whether it came from a fork; secret-backed plans intentionally
run only for trusted branches.
