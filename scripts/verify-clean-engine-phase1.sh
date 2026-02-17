#!/usr/bin/env bash
# Phase 1: Backend Verification for Clean Engine
# Run from repo root. Requires: cd apps/etl, meltano installed (or Docker).
# See docs/CLEAN_ENGINE_TESTING_AND_FRONTEND_PLAN.md

set -e
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

echo "=== Clean Engine Phase 1 Verification ==="
echo ""

# ---------------------------------------------------------------------------
# 1.1 Plugin Installation Check
# ---------------------------------------------------------------------------
echo "--- 1.1 Plugin Installation Check ---"
cd apps/etl

if ! command -v meltano &>/dev/null; then
  echo "WARN: meltano not in PATH. Install with: pip install meltano"
  echo "      Or use Docker: docker build -t ai-bi-etl -f apps/etl/Dockerfile apps/etl"
  echo "      Skipping meltano checks."
else
  # Upgrade psutil before meltano install to avoid macOS + Python 3.12 bug:
  # "SystemError: cpu_count_logical returned a result with an exception set"
  # Use python3 -m pip so we upgrade in the same env meltano uses (usually system Python)
  echo "Upgrading psutil (fixes macOS/Python 3.12 cpu_count_logical bug)..."
  python3 -m pip install --upgrade "psutil>=6.0.0" --quiet 2>/dev/null || true
  echo "Running: meltano install"
  MELTANO_OUT=$(meltano install 2>&1) || MELTANO_RC=$?
  echo "$MELTANO_OUT" | tail -25
  if [ "${MELTANO_RC:-0}" -ne 0 ]; then
    echo ""
    echo "WARN: meltano install failed. Try: python3 -m pip install --upgrade 'psutil>=6.0.0'"
    echo "  Or use Docker: docker build -t ai-bi-etl -f apps/etl/Dockerfile apps/etl"
    echo "  Continuing verification..."
  fi

  echo ""
  echo "Running: meltano config"
  meltano config 2>/dev/null || true

  echo ""
  echo "Checking jobs in meltano.yml..."
  for job in postgres-to-postgres mysql-to-postgres mongodb-to-postgres; do
    if grep -q "name: $job" meltano.yml 2>/dev/null; then
      echo "  ✓ Job defined: $job"
    else
      echo "  ✗ Job missing: $job"
    fi
  done

  if [ -d ".meltano" ]; then
    echo ""
    echo "  .meltano/ directory exists with plugin venvs"
  fi
fi

cd "$ROOT"
echo ""

# ---------------------------------------------------------------------------
# 1.2 Discovery Validation (requires ETL service + DB)
# ---------------------------------------------------------------------------
echo "--- 1.2 Discovery Validation ---"
echo "Manual steps (requires ETL on :8001 + test DBs):"
echo "  1. Start ETL: cd apps/etl && ./run.sh"
echo "  2. POST /discover-schema/postgresql with connection_config"
echo "  3. Verify response has columns, primary_keys, schemas"
echo "  4. Repeat for mysql, mongodb"
echo ""

# ---------------------------------------------------------------------------
# 1.3 State Persistence
# ---------------------------------------------------------------------------
echo "--- 1.3 State Persistence ---"
echo "Verify state_id is passed:"
echo "  - API pipeline.service passes stateId: pipeline_<id>"
echo "  - ETL main.py accepts state_id in payload"
echo "  - pipeline_runner.py passes --state-id to meltano run"
echo "  - Check meltano.db or state backend after run"
echo ""

# ---------------------------------------------------------------------------
# 1.4 CDC Error Handling
# ---------------------------------------------------------------------------
echo "--- 1.4 CDC Error Handling ---"
echo "Manual test: Use DB account without REPLICATION permission."
echo "  - Pipeline should fail with 502"
echo "  - detail should contain _infer_user_message output (e.g. pg_create_logical_replication_slot)"
echo ""

echo ""
echo "=== Phase 3: Docker validation (optional) ==="
echo "Build ETL image: docker build -t ai-bi-etl -f apps/etl/Dockerfile apps/etl"
echo "Run: docker run -p 8001:8001 -e ETL_AUTH_TOKEN=test ai-bi-etl"
echo ""
echo "=== Phase 1 verification script complete ==="
echo "See docs/CLEAN_ENGINE_TESTING_AND_FRONTEND_PLAN.md for full checklist."
