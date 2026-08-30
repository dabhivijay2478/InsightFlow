# Oria agent testing guide

Manual and browser test corpus for all **73 Oria agents** (1 root coordinator + 72 specialists across 6 releases).

## Corpus files (~5,000 prompts)

| Release | Category | File | Agents |
| ---: | --- | --- | ---: |
| Root + 1 | Read | [`oria-test-prompts-release1-read.md`](./test-prompts-release1-read.md) | `oria` + 12 read specialists |
| 2 | Action | [`oria-test-prompts-release2-action.md`](./test-prompts-release2-action.md) | 12 action specialists (pipeline build, connections, transforms, runs) |
| 3 | Automation | [`oria-test-prompts-release3-automation.md`](./test-prompts-release3-automation.md) | 12 automation specialists |
| 4 | Intelligence | [`oria-test-prompts-release4-intelligence.md`](./test-prompts-release4-intelligence.md) | 12 intelligence specialists |
| 5 | Enterprise | [`oria-test-prompts-release5-enterprise.md`](./test-prompts-release5-enterprise.md) | 12 enterprise specialists |
| 6 | Platform | [`oria-test-prompts-release6-platform.md`](./test-prompts-release6-platform.md) | 12 platform specialists |

Each file is **hand-crafted** per agent: realistic user language, MantrixFlow ELT context, and explicit **What to verify** notes. There is no prefix/suffix template spam.

## Prerequisites

1. Go API with `AGENT_RUNTIME_ENABLED=true` and `AGENT_LLM_SYNTHESIS=true` (see [`oria-agent-setup.md`](./agent-setup.md)).
2. Enable release flags for the layer you are testing (Release 2 for action agents, Release 3 for automation, etc.).
3. Sign in to a workspace with sample data (e.g. Stripe All Streams, Postgres Incremental Sync, HubSpot All Streams).
4. Open **`http://localhost:3000/agents`**.

## How to test in the browser

1. Pick a row from the prompt table (`Prompt` column).
2. Start a **new Oria chat** (or stay in-thread for follow-up rows).
3. Send the prompt exactly or as a close natural variant.
4. Compare the live answer to **What to verify**:
   - Public identity is always **Oria** (never internal agent or tool names).
   - Read paths use tool evidence; missing data → state clearly.
   - **Action paths** (Release 2): preview first, explicit confirmation before writes, no silent mutations.
   - **Automation** (Release 3): respect shadow mode and policy gates when enabled.
   - **Intelligence** (Release 4): forecasts include confidence/assumptions; no direct production writes.
   - No credentials, tokens, or raw secrets in responses.
5. For follow-up rows, keep the same thread and confirm context carries forward.

## Test categories (all releases)

| Category | Purpose |
| --- | --- |
| `simple-valid` | Happy path, one clear intent |
| `complex-valid` | Multi-resource or detailed config |
| `multi-step` / `follow-up` | Sequenced or contextual requests |
| `ambiguous` | Should clarify before acting |
| `missing-info` | Should ask for required fields |
| `invalid` | Should reject unsafe or invariant-breaking requests |
| `preview-required` | Action preview before execute (Release 2+) |
| `confirm-flow` / `deny-confirmation` | Confirmation token behavior |
| `unsupported` | Graceful decline |
| `edge` | Collisions, typos, limits, timeouts |

## Agent inventory by release

### Release 1 — Read (12)

`pipeline_context`, `schema_discovery`, `connection_debugger`, `run_failure_investigation`, `sql_safety`, `pipeline_validation`, `replication_key`, `sync_mode`, `schedule_planner`, `billing_usage`, `learning_help`, `audit`

### Release 2 — Action (12)

`pipeline_builder`, `connection_setup`, `stream_selection`, `schema_mapping`, `transformation_builder`, `replication_configuration`, `sync_configuration`, `schedule_manager`, `run_controller`, `failure_recovery`, `destination_manager`, `notification_manager`

### Release 3 — Automation (12)

`automation_policy_manager`, `pipeline_health_monitor`, `incident_triage`, `self_healing_recovery`, `schema_drift_manager`, `data_quality_guardian`, `sla_freshness_manager`, `backfill_coordinator`, `cost_usage_optimizer`, `capacity_concurrency_manager`, `change_rollout_manager`, `governance_compliance_monitor`

### Release 4 — Intelligence (12)

`operations_digital_twin`, `predictive_failure_forecaster`, `change_impact_simulator`, `lineage_dependency_analyzer`, `data_contract_manager`, `workload_demand_forecaster`, `cost_capacity_planner`, `migration_planner`, `resilience_dr_planner`, `anomaly_pattern_miner`, `portfolio_optimizer`, `knowledge_memory_curator`

### Release 5 — Enterprise (12)

`enterprise_mission_control`, `federated_workspace_manager`, `executive_intelligence`, `strategic_goal_planner`, `data_product_governor`, `change_approval_board`, `enterprise_risk_manager`, `compliance_evidence_manager`, `ai_model_governance`, `extension_governance`, `enterprise_finops_planner`, `continuity_command_center`

### Release 6 — Platform (12)

`connector_builder`, `connector_certification`, `api_sdk_manager`, `webhook_integration_manager`, `workflow_template_manager`, `environment_promotion_manager`, `release_deployment_coordinator`, `test_quality_orchestrator`, `support_case_investigator`, `customer_onboarding_manager`, `documentation_publisher`, `marketplace_operations_manager`

## OpenRouter verification

When testing with real LLM synthesis enabled, confirm each non-trivial prompt produces:

- `openrouter_request_start` / `openrouter_request_complete` in Go server logs
- `oria_run_complete … llm_synthesis=true` with specialist id and token counts
- Matching activity in the OpenRouter dashboard for the configured API key

## Related docs

- Runtime setup: [`oria-agent-setup.md`](./agent-setup.md)
- Migration report: [`oria-adk-to-ai-sdk-migration.md`](./adk-to-ai-sdk-migration.md)
- Browser test log: [`oria-browser-ui-test-log.md`](./browser-ui-test-log.md)
