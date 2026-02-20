# Meltano Environment Variable Fix Plan

## Problem

The Singer taps (e.g. `tap-postgres`) receive **empty configuration** and raise `AssertionError` because Meltano does not recognize the environment variables our orchestrator generates.

**Observed:** Meltano logs "Skipping parse of env var settings" and the tap receives empty config.

**Root cause (env var approach):** Meltano requires a specific env var format:
```
MELTANO_<ROLE>_<PLUGIN_NAME_UPPER>_CONFIG_<SETTING_NAME_UPPER>
```

The `_CONFIG_` segment is **mandatory**. Without it, Meltano ignores the variable.

---

## Required Format

| Component | Example |
|-----------|---------|
| Role | `MELTANO_EXTRACTOR` or `MELTANO_LOADER` |
| Plugin name | `TAP_POSTGRES` (dashes → underscores) |
| **CONFIG** | **Required segment** |
| Setting | `SQLALCHEMY_URL`, `DEFAULT_REPLICATION_METHOD` |

**Correct examples:**
- `MELTANO_EXTRACTOR_TAP_POSTGRES_CONFIG_SQLALCHEMY_URL`
- `MELTANO_EXTRACTOR_TAP_MONGODB_CONFIG_MONGODB_CONNECTION_STRING`
- `MELTANO_LOADER_TARGET_POSTGRES_CONFIG_SQLALCHEMY_URL`

**Incorrect (will be ignored):**
- `MELTANO_EXTRACTOR_TAP_POSTGRES_SQLALCHEMY_URL` ❌

---

## Implementation Plan

### Step 1: Verify `connection_mapper.py` ✅ (Already done)

- [x] `_config_dict_to_env_vars` includes `_CONFIG` in env_prefix
- [x] `_plugin_name_to_env_prefix` converts `tap-postgres` → `TAP_POSTGRES`
- [x] `connection_config_to_meltano_env` uses correct prefix for extractor/loader

### Step 2: Use `--config` file instead of env vars ✅ DONE

**File:** `apps/etl/orchestration/discovery.py`

**Problem:** Meltano logs "Skipping parse of env var settings" and ignores env vars entirely.

**Solution:** Pass config via `--config` file. The Singer taps accept `--config /path/to/config.json` directly.

**Change:**
- Build tap config with `connection_config_to_tap_config()` (produces sqlalchemy_url, etc.)
- Write to temp JSON file
- Run: `meltano invoke tap-postgres --config /path/config.json --discover`
- No env vars needed; config is passed directly to the tap

### Step 3: Centralize env var building (optional but recommended)

To avoid future drift, add a helper in `connection_mapper.py`:

```python
def build_meltano_config_env_key(
    role: Literal["extractor", "loader"],
    plugin_name: str,
    setting_name: str,
) -> str:
    """Build Meltano config env key: MELTANO_<ROLE>_<PLUGIN>_CONFIG_<SETTING>."""
    role_prefix = "MELTANO_EXTRACTOR" if role == "extractor" else "MELTANO_LOADER"
    plugin = _plugin_name_to_env_prefix(plugin_name)
    return f"{role_prefix}_{plugin}_CONFIG_{setting_name.upper()}"
```

Then use it in `discovery.py` instead of inline string building.

### Step 4: Add unit tests

**File:** `apps/etl/tests/test_connection_mapper.py` (create if missing)

Test cases:
1. Postgres extractor env vars include `_CONFIG_`
2. MongoDB extractor env vars include `_CONFIG_`
3. Loader env vars include `_CONFIG_`
4. `_plugin_name_to_env_prefix` converts `tap-postgres` → `TAP_POSTGRES`

### Step 5: Verify env vars reach subprocess

**File:** `apps/etl/orchestration/meltano_runner.py`

- `run_meltano_job` and `run_meltano_invoke` merge `env_overrides` into `os.environ.copy()`
- Confirm no key transformation (e.g. lowercasing) occurs
- Env keys must remain uppercase as generated

### Step 6: End-to-end verification

1. Restart FastAPI ETL server
2. Call Test Connection endpoint with valid Postgres credentials
3. Call Discovery endpoint
4. Run a pipeline via `POST /run-meltano-pipeline`
5. Confirm no `AssertionError` about missing connection details

---

## Files to Modify

| File | Change |
|------|--------|
| `orchestration/connection_mapper.py` | ✅ Already has `_CONFIG` |
| `orchestration/discovery.py` | Add `_CONFIG_` to source_config env keys |
| `orchestration/connection_mapper.py` | (Optional) Add `build_meltano_config_env_key` helper |
| `tests/test_connection_mapper.py` | Add unit tests for env var format |

---

## Checklist

- [x] Fix `discovery.py` source_config env key format
- [ ] (Optional) Add `build_meltano_config_env_key` helper
- [ ] Add unit tests
- [ ] Restart ETL server and run Test Connection
- [ ] Run Discovery and verify tap receives config
- [ ] Run pipeline and verify end-to-end success
