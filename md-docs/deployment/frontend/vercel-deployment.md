# MantrixFlow Frontend Deployment

The frontend deploys through Vercel Git integration, not GitHub Actions.

## Vercel Project

- Git repository: `dabhivijay2478/InsightFlow-app`
- Framework preset: Next.js
- Install command: `bun install --frozen-lockfile`
- Build command: `bun run build`
- Production branch: `mantrixflow`
- Production domain: `cloud.mantrixflow.com`

## Production Environment Variables

Store runtime values in Vercel project settings:

- `NEXT_PUBLIC_API_URL=https://cloud.api.mantrixflow.com`
- `NEXT_PUBLIC_SUPABASE_URL`
- `NEXT_PUBLIC_SUPABASE_ANON_KEY`
- `NEXT_PUBLIC_APP_URL=https://cloud.mantrixflow.com`
- `NEXT_PUBLIC_SITE_URL=https://cloud.mantrixflow.com`

Optional:

- `GOOGLE_FONTS_API_KEY`
- `SLACK_PROXY_TARGET_URL=https://cloud.api.mantrixflow.com`

Do not store `SUPABASE_SERVICE_ROLE_KEY`, internal tokens, test tokens, localhost URLs, ngrok URLs, or `NEXT_PUBLIC_PYTHON_SERVICE_URL` in Vercel production.

GitHub Actions is only the CI gate for lint/build before merge. Vercel creates preview deployments for pull requests and the production deployment after merge to `mantrixflow`.
