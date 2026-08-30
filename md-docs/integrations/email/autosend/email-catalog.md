# AutoSend Email Catalog

Preference gates:

- `always`: auth, security, billing-critical, and legal/account notices.
- `pipelineStatusEmails`: currently implemented through `email_preferences.pipeline_failure_emails`.
- `weeklyDigestEnabled`: weekly digest.
- `marketingEmails`: onboarding and re-engagement.

## Supabase Auth via AutoSend SMTP

| Template key | Trigger | Audience | Gate | Required variables | CTA | Dedupe | Priority |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `auth_signup_confirmation` | User signs up or confirms account | New user | always | `confirmation_url`, `email`, `site_url` | Confirm account | Supabase Auth token | high |
| `auth_workspace_invite` | Org invite through Supabase invite link | Invited member | always | `invite_url`, `org_name`, `inviter_email`, `email` | Accept invite | Supabase Auth token | high |
| `auth_magic_link` | Passwordless login requested | User | always | `magic_link`, `email`, `site_url` | Sign in | Supabase Auth token | high |
| `auth_password_recovery` | Reset requested | User | always | `recovery_url`, `email`, `site_url` | Reset password | Supabase Auth token | high |
| `auth_email_change` | Email change requested | User | always | `confirmation_url`, `new_email`, `old_email` | Confirm email | Supabase Auth token | high |
| `auth_reauthentication` | Security reauth requested | User | always | `confirmation_url`, `email` | Continue securely | Supabase Auth token | high |

## Go Backend via AutoSend API

| Template key | Trigger | Audience | Gate | Required variables | CTA | Dedupe | Priority |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `organization_invite` | Backend workspace invite fallback | Invited member | always | `org_name`, `inviter_email`, `role`, `accept_url`, `recipient_email` | Accept invite | invite ID | normal |
| `member_removed` | Member removed from workspace | Removed member | always | `first_name`, `org_name`, `dashboard_url` | Open workspace | `member_removed:{org}:{member}` | normal |
| `workspace_role_changed` | Member role changes | Changed member | always | `first_name`, `org_name`, `old_role`, `new_role`, `workspace_url` | Review workspace | `workspace_role_changed:{org}:{user}:{role}` | normal |
| `pipeline_created` | Pipeline created | Pipeline creator | pipelineStatusEmails | `first_name`, `org_name`, `pipeline_name`, `schedule_summary`, `pipeline_url` | View pipeline | `pipeline_created:{pipeline}` | normal |
| `pipeline_queued` | Run queued due capacity/limits | Triggering user | pipelineStatusEmails | `pipeline_name`, `queue_reason`, `queue_position`, `run_detail_url` | View run | `pipeline_email:{run}:pipeline_queued` | normal |
| `pipeline_starting` | Queued run promoted or immediate run starts | Triggering user | pipelineStatusEmails | `pipeline_name`, `source_stream`, `dest_table`, `run_detail_url` | View run | `pipeline_email:{run}:pipeline_starting` | normal |
| `pipeline_run_success` | Normal successful run | Triggering user or creator | pipelineStatusEmails | `pipeline_name`, `rows_upserted`, `duration_seconds`, `dest_table`, `pipeline_url` | View run | `pipeline_email:{run}:pipeline_run_success` | normal |
| `pipeline_run_failed` | Run failed | Triggering user or creator | pipelineStatusEmails | `pipeline_name`, `source_stream`, `dest_table`, `error_message`, `run_detail_url`, `edit_pipeline_url` | Fix pipeline | `pipeline_email:{run}:pipeline_run_failed` | high |
| `pipeline_partial_success` | Partial delivery or timeout partial success | Triggering user or creator | pipelineStatusEmails | `pipeline_name`, `rows_upserted`, `timeout_seconds`, `run_detail_url` | Review run | `pipeline_email:{run}:pipeline_partial_success` | normal |
| `first_success` | First successful run for pipeline | Pipeline creator | pipelineStatusEmails | `pipeline_name`, `rows_upserted`, `dest_table`, `duration_seconds`, `pipeline_url` | View pipeline | `pipeline_email:{run}:first_success` | normal |
| `pipeline_recovered` | Previous failed pipeline succeeds | Pipeline creator | pipelineStatusEmails | `pipeline_name`, `rows_upserted`, `duration_seconds`, `run_history_url` | View history | `pipeline_email:{run}:pipeline_recovered` | normal |
| `pipeline_disabled` | Pipeline paused after repeated failures | Pipeline creator | pipelineStatusEmails | `pipeline_name`, `failure_count`, `last_error_message`, `edit_pipeline_url` | Fix pipeline | `pipeline_disabled:{pipeline}:{count}` | high |
| `pipeline_schedule_changed` | Schedule created, updated, or removed | Pipeline creator | pipelineStatusEmails | `first_name`, `org_name`, `pipeline_name`, `old_schedule`, `new_schedule`, `pipeline_url` | View pipeline | `pipeline_schedule_changed:{pipeline}:{new_schedule}` | normal |
| `incremental_setup_complete` | Incremental sync setup completed | Connection owner | pipelineStatusEmails | `connection_name`, `create_pipeline_url` | Create pipeline | connection ID | normal |
| `incremental_initial_complete` | First incremental baseline load completed | Pipeline creator | pipelineStatusEmails | `pipeline_name`, `rows_upserted`, `dest_table`, `pipeline_url` | View pipeline | run ID | normal |
| `usage_warning_80` | Org reaches 80% row allowance | Owners/admins | always | `first_name`, `org_name`, `usage_percent`, `rows_used`, `row_limit`, `plan_name`, `billing_month`, `billing_url` | Manage plan | `usage_warning_80:{org}:{month}:{email}` | high |
| `usage_limit_reached` | Org reaches hard row cap | Owners/admins | always | `first_name`, `org_name`, `usage_percent`, `rows_used`, `row_limit`, `plan_name`, `billing_month`, `billing_url` | Upgrade plan | `usage_limit_reached:{org}:{month}:{email}` | high |
| `weekly_digest` | Monday morning summary | Opted-in users | weeklyDigestEnabled | `org_name`, `week_start_date`, `total_runs`, `success_rate`, `failed_runs`, `rows_synced`, `top_pipeline_name`, `analytics_url` | View activity | `weekly_digest:{user}:{date}` | low |
| `reengagement_14_days` | 14 days inactive with pipelines | Opted-in users | marketingEmails | `first_name`, `org_name`, `pipeline_count`, `dashboard_url` | Open pipelines | `reengagement_14_days:{user}:{org}` | low |
| `onboarding_day3_nudge` | Day 3 incomplete onboarding | Opted-in users | marketingEmails | `first_name`, `dashboard_url`, `organization_name` | Continue setup | `onboarding_day3_nudge:{user}` | low |
| `onboarding_day7_nudge` | Day 7 incomplete onboarding | Opted-in users | marketingEmails | `first_name`, `dashboard_url`, `organization_name` | Finish setup | `onboarding_day7_nudge:{user}` | low |
| `connection_created` | Source/destination connected | Connection creator | pipelineStatusEmails | `first_name`, `org_name`, `connection_name`, `connector_type`, `connection_url` | View connection | connection ID | normal |
| `connection_error` | Saved connection enters error state | Owner/admin | pipelineStatusEmails | `first_name`, `org_name`, `connection_name`, `error_message`, `connection_url` | Fix connection | connection status event | high |
| `pipeline_deleted` | Pipeline deleted | Pipeline creator/admin | pipelineStatusEmails | `first_name`, `org_name`, `pipeline_name`, `deleted_at`, `workspace_url` | Open pipelines | `pipeline_deleted:{pipeline}` | normal |

## Dodo Payments AutoSend Integration

| Template key | Trigger | Audience | Gate | Required variables | CTA | Dedupe | Priority |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `payment_succeeded` | `payment.succeeded` | Customer | always | `customer_name`, `amount`, `currency`, `payment_id`, `receipt_url`, `date` | View receipt | Dodo event/payment ID | high |
| `payment_failed` | `payment.failed` | Customer | always | `customer_name`, `amount`, `currency`, `payment_id`, `billing_url`, `failure_reason` | Update billing | Dodo event/payment ID | high |
| `subscription_active` | `subscription.active` | Customer | always | `customer_name`, `plan_name`, `billing_period`, `billing_url` | Open workspace | Dodo event/subscription ID | high |
| `subscription_renewed` | Renewal event | Customer | always | `customer_name`, `plan_name`, `renewed_at`, `next_billing_date`, `billing_url` | Manage billing | Dodo event/subscription ID | normal |
| `subscription_plan_changed` | Upgrade/downgrade | Customer | always | `customer_name`, `old_plan_name`, `new_plan_name`, `effective_date`, `billing_url` | Review plan | Dodo event/subscription ID | high |
| `subscription_cancelled` | Cancelled or cancel-at-period-end | Customer | always | `customer_name`, `plan_name`, `access_until`, `billing_url` | Manage subscription | Dodo event/subscription ID | high |
| `subscription_on_hold` | Subscription/payment hold | Customer | always | `customer_name`, `plan_name`, `reason`, `billing_url` | Fix billing | Dodo event/subscription ID | high |
| `subscription_expired` | Access expired | Customer | always | `customer_name`, `plan_name`, `expired_at`, `billing_url` | Reactivate | Dodo event/subscription ID | high |
| `refund_succeeded` | Refund completed | Customer | always | `customer_name`, `amount`, `currency`, `refund_id`, `payment_id` | View billing | Dodo event/refund ID | normal |
| `dispute_opened` | Dispute/chargeback opened | Admin/internal | always | `amount`, `currency`, `payment_id`, `customer_email`, `dashboard_url` | Open Dodo | Dodo event/dispute ID | high |
| `invoice_available` | Invoice or receipt available | Customer | always | `customer_name`, `invoice_number`, `amount`, `currency`, `invoice_url` | View invoice | Dodo event/invoice ID | normal |
