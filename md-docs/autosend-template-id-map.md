# AutoSend Template ID Map

Generated from `apps/server/main-server/internal/email/templates/*.html` after deleting and recreating MantrixFlow-managed templates via the AutoSend API.

## Deleted Before Recreate

| AutoSend template name | Old template ID |
| --- | --- |
| `MantrixFlow / member_removed` | `A-8faa245feda927e90e1a` |
| `MantrixFlow / organization_invite` | `A-c54473fc25f390eb9a1c` |

## Current Templates

| Template key | AutoSend template name | Template ID | Env var | Type |
| --- | --- | --- | --- | --- |
| `organization_invite` | `MantrixFlow / organization_invite` | `A-13f385ebe4e515c5fdba` | `AUTOSEND_TEMPLATE_ORG_INVITE` | transactional |
| `member_removed` | `MantrixFlow / member_removed` | `A-fe931f5dd54404c7bb14` | `AUTOSEND_TEMPLATE_MEMBER_REMOVED` | transactional |
| `workspace_role_changed` | `MantrixFlow / workspace_role_changed` | `A-daa3404f2b8e99fd5a49` | `AUTOSEND_TEMPLATE_WORKSPACE_ROLE_CHANGED` | transactional |
| `pipeline_created` | `MantrixFlow / pipeline_created` | `A-beebe2fb91a3a44ed5ed` | `AUTOSEND_TEMPLATE_PIPELINE_CREATED` | transactional |
| `pipeline_queued` | `MantrixFlow / pipeline_queued` | `A-48ab3a0f90992ccfdce9` | `AUTOSEND_TEMPLATE_PIPELINE_QUEUED` | transactional |
| `pipeline_starting` | `MantrixFlow / pipeline_starting` | `A-54ff7dac742baf2ef9c5` | `AUTOSEND_TEMPLATE_PIPELINE_STARTING` | transactional |
| `pipeline_run_success` | `MantrixFlow / pipeline_run_success` | `A-208bea0f6a9b177efbd6` | `AUTOSEND_TEMPLATE_PIPELINE_RUN_SUCCESS` | transactional |
| `pipeline_run_failed` | `MantrixFlow / pipeline_run_failed` | `A-a2f768b21cca086526ee` | `AUTOSEND_TEMPLATE_PIPELINE_RUN_FAILED` | transactional |
| `pipeline_partial_success` | `MantrixFlow / pipeline_partial_success` | `A-35ddb7838e6afe10acc4` | `AUTOSEND_TEMPLATE_PIPELINE_PARTIAL_SUCCESS` | transactional |
| `first_success` | `MantrixFlow / first_success` | `A-b3150f6991578e8be9cd` | `AUTOSEND_TEMPLATE_FIRST_SUCCESS` | transactional |
| `pipeline_recovered` | `MantrixFlow / pipeline_recovered` | `A-7fe39607039b04690897` | `AUTOSEND_TEMPLATE_PIPELINE_RECOVERED` | transactional |
| `pipeline_disabled` | `MantrixFlow / pipeline_disabled` | `A-6645a104c84e5806de0a` | `AUTOSEND_TEMPLATE_PIPELINE_DISABLED` | transactional |
| `pipeline_schedule_changed` | `MantrixFlow / pipeline_schedule_changed` | `A-d9f57966d6384d910cdc` | `AUTOSEND_TEMPLATE_PIPELINE_SCHEDULE_CHANGED` | transactional |
| `incremental_setup_complete` | `MantrixFlow / incremental_setup_complete` | `A-8cfe848d77d53e70bf0a` | `AUTOSEND_TEMPLATE_INCREMENTAL_SETUP_COMPLETE` | transactional |
| `incremental_initial_complete` | `MantrixFlow / incremental_initial_complete` | `A-dadb53596fd59af0d9b6` | `AUTOSEND_TEMPLATE_INCREMENTAL_INITIAL_COMPLETE` | transactional |
| `usage_warning_80` | `MantrixFlow / usage_warning_80` | `A-9d98e50ad517101c1403` | `AUTOSEND_TEMPLATE_USAGE_WARNING_80` | transactional |
| `usage_limit_reached` | `MantrixFlow / usage_limit_reached` | `A-a0159d272ea193527d5e` | `AUTOSEND_TEMPLATE_USAGE_LIMIT_REACHED` | transactional |
| `weekly_digest` | `MantrixFlow / weekly_digest` | `A-828f9c40d0b5df1ff40c` | `AUTOSEND_TEMPLATE_WEEKLY_DIGEST` | marketing |
| `reengagement_14_days` | `MantrixFlow / reengagement_14_days` | `A-08eaa38f8e2989c3be4b` | `AUTOSEND_TEMPLATE_REENGAGEMENT_14_DAYS` | marketing |
| `onboarding_day3_nudge` | `MantrixFlow / onboarding_day3_nudge` | `A-29dfa1c5633e88a43d72` | `AUTOSEND_TEMPLATE_ONBOARDING_DAY3_NUDGE` | marketing |
| `onboarding_day7_nudge` | `MantrixFlow / onboarding_day7_nudge` | `A-3d107ea5bb257bd254b5` | `AUTOSEND_TEMPLATE_ONBOARDING_DAY7_NUDGE` | marketing |
| `connection_created` | `MantrixFlow / connection_created` | `A-ac28cb59277d6534a744` | `AUTOSEND_TEMPLATE_CONNECTION_CREATED` | transactional |
| `connection_error` | `MantrixFlow / connection_error` | `A-158ad0b25cc108964cdf` | `AUTOSEND_TEMPLATE_CONNECTION_ERROR` | transactional |
| `pipeline_deleted` | `MantrixFlow / pipeline_deleted` | `A-eabb1c6d3c8602baef39` | `AUTOSEND_TEMPLATE_PIPELINE_DELETED` | transactional |
| `trial_started` | `MantrixFlow / trial_started` | `A-5b250383e6f9aa3e6e5a` | `AUTOSEND_TEMPLATE_TRIAL_STARTED` | transactional |
| `trial_ends_7_days` | `MantrixFlow / trial_ends_7_days` | `A-48833eea5afa2659fb3d` | `AUTOSEND_TEMPLATE_TRIAL_ENDS_7_DAYS` | transactional |
| `trial_ends_1_day` | `MantrixFlow / trial_ends_1_day` | `A-6e65de00867a30b038b7` | `AUTOSEND_TEMPLATE_TRIAL_ENDS_1_DAY` | transactional |
| `trial_expired` | `MantrixFlow / trial_expired` | `A-6c7ccf48715ad79280f5` | `AUTOSEND_TEMPLATE_TRIAL_EXPIRED` | transactional |
| `payment_failed` | `MantrixFlow / payment_failed` | `A-26b35ffc124523857bf4` | `AUTOSEND_TEMPLATE_PAYMENT_FAILED` | transactional |

## Env Block

```bash
AUTOSEND_TEMPLATE_ORG_INVITE=A-13f385ebe4e515c5fdba
AUTOSEND_TEMPLATE_MEMBER_REMOVED=A-fe931f5dd54404c7bb14
AUTOSEND_TEMPLATE_WORKSPACE_ROLE_CHANGED=A-daa3404f2b8e99fd5a49
AUTOSEND_TEMPLATE_PIPELINE_CREATED=A-beebe2fb91a3a44ed5ed
AUTOSEND_TEMPLATE_PIPELINE_QUEUED=A-48ab3a0f90992ccfdce9
AUTOSEND_TEMPLATE_PIPELINE_STARTING=A-54ff7dac742baf2ef9c5
AUTOSEND_TEMPLATE_PIPELINE_RUN_SUCCESS=A-208bea0f6a9b177efbd6
AUTOSEND_TEMPLATE_PIPELINE_RUN_FAILED=A-a2f768b21cca086526ee
AUTOSEND_TEMPLATE_PIPELINE_PARTIAL_SUCCESS=A-35ddb7838e6afe10acc4
AUTOSEND_TEMPLATE_FIRST_SUCCESS=A-b3150f6991578e8be9cd
AUTOSEND_TEMPLATE_PIPELINE_RECOVERED=A-7fe39607039b04690897
AUTOSEND_TEMPLATE_PIPELINE_DISABLED=A-6645a104c84e5806de0a
AUTOSEND_TEMPLATE_PIPELINE_SCHEDULE_CHANGED=A-d9f57966d6384d910cdc
AUTOSEND_TEMPLATE_INCREMENTAL_SETUP_COMPLETE=A-8cfe848d77d53e70bf0a
AUTOSEND_TEMPLATE_INCREMENTAL_INITIAL_COMPLETE=A-dadb53596fd59af0d9b6
AUTOSEND_TEMPLATE_USAGE_WARNING_80=A-9d98e50ad517101c1403
AUTOSEND_TEMPLATE_USAGE_LIMIT_REACHED=A-a0159d272ea193527d5e
AUTOSEND_TEMPLATE_WEEKLY_DIGEST=A-828f9c40d0b5df1ff40c
AUTOSEND_TEMPLATE_REENGAGEMENT_14_DAYS=A-08eaa38f8e2989c3be4b
AUTOSEND_TEMPLATE_ONBOARDING_DAY3_NUDGE=A-29dfa1c5633e88a43d72
AUTOSEND_TEMPLATE_ONBOARDING_DAY7_NUDGE=A-3d107ea5bb257bd254b5
AUTOSEND_TEMPLATE_CONNECTION_CREATED=A-ac28cb59277d6534a744
AUTOSEND_TEMPLATE_CONNECTION_ERROR=A-158ad0b25cc108964cdf
AUTOSEND_TEMPLATE_PIPELINE_DELETED=A-eabb1c6d3c8602baef39
AUTOSEND_TEMPLATE_TRIAL_STARTED=A-5b250383e6f9aa3e6e5a
AUTOSEND_TEMPLATE_TRIAL_ENDS_7_DAYS=A-48833eea5afa2659fb3d
AUTOSEND_TEMPLATE_TRIAL_ENDS_1_DAY=A-6e65de00867a30b038b7
AUTOSEND_TEMPLATE_TRIAL_EXPIRED=A-6c7ccf48715ad79280f5
AUTOSEND_TEMPLATE_PAYMENT_FAILED=A-26b35ffc124523857bf4
```
