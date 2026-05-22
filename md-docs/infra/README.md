# MantrixFlow Production Infrastructure Docs

This folder is the operator guide for the separate production infrastructure repo and the three service repos.

## Manual Setup Count

One-time manual setup is **18 steps** total:

- **8 AWS actions**: admin login, state bucket, CDK bootstrap, GitHub OIDC provider, infra deploy role, API deploy role, ELT deploy role, and the first AWS bootstrap verification.
- **4 GitHub steps**: environments, branch rules, repo secrets, and first workflow runs.
- **3 external service steps**: Cloudflare token, Supabase project, Vercel project.
- **3 verification steps**: API health, ECS service stability, and first pipeline callback smoke.

The detailed guide groups the three AWS role creations into one numbered step, but they are counted separately here because they are three IAM roles with different trust policies.

After this bootstrap, normal production deploys require **0 manual AWS steps**. Production changes deploy by merging:

- Infra repo `main` updates shared AWS and Cloudflare resources.
- Frontend repo `production` deploys through Vercel.
- Go API repo `production` deploys only the API ECS service.
- ELT repo `production` deploys only the ELT ECS service.

## Docs

- [Production setup guide](production-setup-guide.md)
- [AWS GitHub OIDC role guide](aws-github-oidc-roles.md)
- [Environment inventory from real local env files](env-inventory.md)
