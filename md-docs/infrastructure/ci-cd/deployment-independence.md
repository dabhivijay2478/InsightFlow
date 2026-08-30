# Deployment independence checks

- Go CI changes only the Go application in self-hosted Dokploy.
- ELT CI changes only the ELT application.
- Simulation-manager CI changes only its single-replica application.
- Simulation-runtime CI publishes an image and does not deploy persistent ELT.
- Terraform creates static OVH infrastructure and never updates application images.
- Database and PgBouncer remain separate, manually activated projects.
- Frontend deploys only through Vercel.
- Rollback selects a previous immutable digest for one application without a rebuild.

Run these checks during staging installation and after changing Dokploy application
IDs, protected workflow environments, or remote-server assignments.

Use `../scripts/check-github-environments.sh` to compare the live protected
environment names with the workflow contract. The command is read-only and
requires an authenticated GitHub CLI session.
