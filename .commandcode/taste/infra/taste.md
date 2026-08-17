# infra
- Production infrastructure lives in `apps/mantrixflow-infra` as its own standalone Git repository (CDK + Terraform). Not inside `apps/app`. Confidence: 0.85
- Infra repo deploys on merge to `main`; app/API/ELT repos use `mantrixflow` as production branch. Confidence: 0.75
- Branch strategy: frontend app uses `main` as default branch; ELT server and main server use `mantrixflow` as default branch. Other repos (docs, website) use `main`. This asymmetric split must be preserved when merging feature work. Confidence: 0.90
- When merging feature branches, prefer merging locally first (not pushing to origin) so conflicts surface and resolve in the local checkout before any remote push. Confidence: 0.75
- Use GitHub OIDC for AWS auth in CI/CD, not long-lived AWS access keys. Confidence: 0.80
- Use Cloudflare API token (not API key) for Terraform Cloudflare provider authentication. Confidence: 0.75
- Main server repo: `dabhivijay2478/cloud.api.mantrixflow.com`. ELT server repo: `dabhivijay2478/cloud.api.etl.server.mantrixflow.com`. Confidence: 0.75
