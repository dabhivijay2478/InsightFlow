#!/usr/bin/env bash
# One-time: stop the monorepo from indexing apps/api and give apps/api its own .git
# pushing to https://github.com/dabhivijay2478/cloud.api.mantrixflow.com
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
CLOUD_API_URL="${CLOUD_API_URL:-https://github.com/dabhivijay2478/cloud.api.mantrixflow.com.git}"

cd "$ROOT"

echo "=== 1. Remove apps/api from monorepo Git index (files stay on disk) ==="
if git ls-files --error-unmatch apps/api >/dev/null 2>&1; then
  git rm -r --cached apps/api
  echo "OK: staged 'git rm --cached'. Next: git commit -m 'chore: apps/api is standalone (cloud.api repo)'"
else
  echo "apps/api is not in the monorepo index (already standalone or missing)."
fi

echo "=== 2. Standalone Git repo inside apps/api ==="
mkdir -p apps/api
cd apps/api

if [ -d .git ]; then
  echo ".git already exists — skip init."
else
  git init
  git branch -M main
  echo "Initialized new repo in apps/api"
fi

if git remote get-url origin >/dev/null 2>&1; then
  git remote set-url origin "$CLOUD_API_URL"
else
  git remote add origin "$CLOUD_API_URL"
fi

git fetch origin main 2>/dev/null || echo "Note: no remote main yet (first push OK)"

echo "=== 3. Commit local tree if there are changes ==="
git add -A
if git diff --cached --quiet 2>/dev/null; then
  echo "Nothing new to commit in apps/api."
else
  git commit -m "chore: sync cloud API workspace" || true
fi

echo "=== 4. Push to cloud.api (set upstream) ==="
git push -u origin main || {
  echo "Push failed. If histories differ, try:"
  echo "  git pull origin main --rebase --allow-unrelated-histories"
  echo "  git push -u origin main"
  exit 1
}

echo "Done. Open mantrixflow.code-workspace so Cloud API shows its own Source Control."
