#!/usr/bin/env bash
# Push only apps/api to https://github.com/dabhivijay2478/cloud.api.mantrixflow.com
# Requires: git remote "cloud-api" (see repo .git/config)
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

if ! git remote get-url cloud-api >/dev/null 2>&1; then
  echo "Adding remote cloud-api..."
  git remote add cloud-api https://github.com/dabhivijay2478/cloud.api.mantrixflow.com.git
fi

BRANCH="${1:-main}"
TMP_BRANCH="cloud-api-export-$(date +%s)"

echo "Splitting subtree apps/api -> $TMP_BRANCH"
git subtree split --prefix=apps/api -b "$TMP_BRANCH"

echo "Pushing to cloud-api ($BRANCH)..."
git push cloud-api "$TMP_BRANCH:$BRANCH"

git branch -D "$TMP_BRANCH"
echo "Done. Remote: cloud-api -> $(git remote get-url cloud-api)"
