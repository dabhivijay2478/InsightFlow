# AutoSend Template Copy

Use these subjects, preview lines, CTA labels, and body blocks when creating AutoSend templates. Variables use `{{variable_name}}`.

## Supabase Auth SMTP

| Template | Subject | Preview | CTA | Body |
| --- | --- | --- | --- | --- |
| `auth_signup_confirmation` | Confirm your MantrixFlow account | Finish setting up your workspace. | Confirm account | Thanks for signing up for MantrixFlow. Confirm this email address to finish creating your account. If you did not request this, you can ignore this email. |
| `auth_workspace_invite` | You are invited to join {{org_name}} | {{inviter_email}} invited you to MantrixFlow. | Accept invite | You have been invited to join {{org_name}} on MantrixFlow. Use the secure invite link below to accept and finish setting up your account. |
| `auth_magic_link` | Sign in to MantrixFlow | Use this secure link to continue. | Sign in | Use this link to sign in to MantrixFlow. The link expires soon and can only be used for this sign-in request. |
| `auth_password_recovery` | Reset your MantrixFlow password | A reset was requested for this account. | Reset password | We received a request to reset your password. Use the secure link below to choose a new one. If this was not you, ignore this email. |
| `auth_email_change` | Confirm your new email address | Confirm the email change for your account. | Confirm email | Confirm that you want to use {{new_email}} for your MantrixFlow account. |
| `auth_reauthentication` | Confirm it is you | Continue your security-sensitive action. | Continue securely | MantrixFlow needs to confirm it is you before continuing. Use the secure link below. |

## Backend Product Email

| Template | Subject | Preview | CTA | Body |
| --- | --- | --- | --- | --- |
| `organization_invite` | You're invited to join {{org_name}} on MantrixFlow | {{inviter_email}} invited you as {{role}}. | Accept invite | {{inviter_email}} invited {{recipient_email}} to join {{org_name}} on MantrixFlow. Accept the invite to access the workspace. |
| `member_removed` | Your access to {{org_name}} was removed | You no longer have access to this workspace. | Open workspace | Hi {{first_name}}, your access to {{org_name}} has been removed. If this looks wrong, contact a workspace owner. |
| `workspace_role_changed` | Your role changed in {{org_name}} | Your workspace permissions were updated. | Review workspace | Hi {{first_name}}, your role in {{org_name}} changed from {{old_role}} to {{new_role}}. The new role is active now. |
| `pipeline_created` | Pipeline created: {{pipeline_name}} | The pipeline is ready in {{org_name}}. | View pipeline | Hi {{first_name}}, {{pipeline_name}} was created in {{org_name}}. Current schedule: {{schedule_summary}}. |
| `pipeline_queued` | Pipeline queued: {{pipeline_name}} | The run is waiting for capacity. | View run | {{pipeline_name}} is queued. Reason: {{queue_reason}}. We will start it automatically when capacity is available. |
| `pipeline_starting` | Pipeline starting: {{pipeline_name}} | The run has started. | View run | {{pipeline_name}} is starting now. Source: {{source_stream}}. Destination: {{dest_table}}. |
| `pipeline_run_success` | Pipeline synced: {{pipeline_name}} | {{rows_upserted}} rows delivered. | View run | {{pipeline_name}} completed successfully. {{rows_upserted}} rows were delivered to {{dest_table}} in {{duration_seconds}} seconds. |
| `pipeline_run_failed` | Pipeline failed: {{pipeline_name}} | Review the safe error summary and fix the pipeline. | Fix pipeline | {{pipeline_name}} failed. Safe error summary: {{error_message}}. Review the run details before retrying. |
| `pipeline_partial_success` | Partial sync success: {{pipeline_name}} | Some rows were delivered before the run stopped. | Review run | {{pipeline_name}} partially completed. {{rows_upserted}} rows were delivered before the run stopped after {{timeout_seconds}} seconds. |
| `first_success` | First successful run: {{pipeline_name}} | Your first pipeline run delivered data. | View pipeline | Nice work. {{pipeline_name}} completed its first successful run and delivered {{rows_upserted}} rows to {{dest_table}}. |
| `pipeline_recovered` | {{pipeline_name}} is healthy again | The latest run completed successfully. | View history | {{pipeline_name}} recovered after a previous failure. The latest run delivered {{rows_upserted}} rows in {{duration_seconds}} seconds. |
| `pipeline_disabled` | Pipeline paused after repeated failures: {{pipeline_name}} | MantrixFlow paused this pipeline to protect your workspace. | Fix pipeline | {{pipeline_name}} was paused after {{failure_count}} repeated failures. Last safe error summary: {{last_error_message}}. |
| `pipeline_schedule_changed` | Schedule changed: {{pipeline_name}} | The pipeline schedule was updated. | View pipeline | Hi {{first_name}}, the schedule for {{pipeline_name}} changed from {{old_schedule}} to {{new_schedule}}. |
| `incremental_setup_complete` | Incremental setup complete for {{connection_name}} | You can now create incremental pipelines. | Create pipeline | Incremental setup is complete for {{connection_name}}. New pipelines can now use incremental sync where a replication key is available. |
| `incremental_initial_complete` | Initial incremental load complete: {{pipeline_name}} | Baseline rows are loaded. | View pipeline | {{pipeline_name}} completed its initial incremental baseline load with {{rows_upserted}} rows delivered to {{dest_table}}. |
| `usage_warning_80` | {{org_name}} is at {{usage_percent}}% of its row allowance | {{rows_used}} of {{row_limit}} rows used this month. | Manage plan | Hi {{first_name}}, {{org_name}} has used {{rows_used}} of {{row_limit}} rows on the {{plan_name}} plan for {{billing_month}}. |
| `usage_limit_reached` | {{org_name}} reached its row limit | New runs may be blocked until the plan is updated. | Upgrade plan | Hi {{first_name}}, {{org_name}} has reached its monthly row allowance of {{row_limit}} rows on the {{plan_name}} plan. |
| `weekly_digest` | Weekly digest - {{org_name}} | {{total_runs}} runs, {{success_rate}} success rate. | View activity | Here is the MantrixFlow summary for the week of {{week_start_date}}. {{total_runs}} runs completed with a {{success_rate}} success rate and {{rows_synced}} rows synced. |
| `reengagement_14_days` | Your pipelines are waiting in {{org_name}} | You still have {{pipeline_count}} pipelines ready. | Open pipelines | Hi {{first_name}}, {{org_name}} still has {{pipeline_count}} pipelines ready when you need them. Open the workspace to review recent activity or run a sync. |
| `onboarding_day3_nudge` | Quick check-in on your MantrixFlow setup | Continue setup when you are ready. | Continue setup | Hi {{first_name}}, your MantrixFlow workspace is ready. Continue setup to connect a source, choose a destination, and create your first pipeline. |
| `onboarding_day7_nudge` | Need help finishing your MantrixFlow setup? | Your workspace is still waiting. | Finish setup | Hi {{first_name}}, your MantrixFlow setup is still open. Finish the workspace setup to start syncing data. |
| `connection_created` | Connection ready: {{connection_name}} | {{connector_type}} is available in {{org_name}}. | View connection | Hi {{first_name}}, {{connection_name}} is connected and ready in {{org_name}}. |
| `connection_error` | Connection needs attention: {{connection_name}} | Review the saved connection. | Fix connection | Hi {{first_name}}, {{connection_name}} needs attention. Safe error summary: {{error_message}}. |
| `pipeline_deleted` | Pipeline deleted: {{pipeline_name}} | A pipeline was removed from {{org_name}}. | Open pipelines | Hi {{first_name}}, {{pipeline_name}} was deleted at {{deleted_at}}. This is an audit notice for your workspace. |

## Billing Trial Email Owned by Backend

| Template | Subject | Preview | CTA | Body |
| --- | --- | --- | --- | --- |
| `trial_started` | Your MantrixFlow trial for {{org_name}} has started | Trial access is active. | Open workspace | Hi {{first_name}}, your MantrixFlow trial for {{org_name}} has started and ends on {{trial_end_date}}. |
| `trial_ends_7_days` | {{org_name}} trial ends in 7 days | Keep your workspace running after trial. | Upgrade plan | {{org_name}} trial ends on {{trial_end_date}}. You currently have {{pipeline_count}} pipelines, {{connection_count}} connections, and {{rows_synced_total}} rows synced. |
| `trial_ends_1_day` | Last day of trial - {{org_name}} | Upgrade to keep pipelines running. | Upgrade plan | {{org_name}} trial ends on {{trial_end_date}}. Choose a plan to keep workspace access and scheduled pipelines active. |
| `trial_expired` | Your MantrixFlow trial has ended | Upgrade to resume paused workspace access. | Upgrade plan | {{org_name}} trial has ended. {{paused_pipeline_count}} pipelines may be paused until a plan is active. |

## Dodo Payments Email

| Template | Subject | Preview | CTA | Body |
| --- | --- | --- | --- | --- |
| `payment_succeeded` | Payment received for MantrixFlow | {{amount}} {{currency}} was paid successfully. | View receipt | Hi {{customer_name}}, we received your payment of {{amount}} {{currency}}. Payment ID: {{payment_id}}. |
| `payment_failed` | Payment failed for MantrixFlow | Update billing to keep access active. | Update billing | Hi {{customer_name}}, your payment of {{amount}} {{currency}} failed. Reason: {{failure_reason}}. |
| `subscription_active` | Your MantrixFlow subscription is active | {{plan_name}} is now active. | Open workspace | Hi {{customer_name}}, your {{plan_name}} subscription is active for {{billing_period}}. |
| `subscription_renewed` | MantrixFlow subscription renewed | Your next billing date is {{next_billing_date}}. | Manage billing | Hi {{customer_name}}, your {{plan_name}} subscription renewed on {{renewed_at}}. |
| `subscription_plan_changed` | MantrixFlow plan changed | {{old_plan_name}} changed to {{new_plan_name}}. | Review plan | Hi {{customer_name}}, your plan changed from {{old_plan_name}} to {{new_plan_name}} effective {{effective_date}}. |
| `subscription_cancelled` | MantrixFlow subscription cancelled | Access remains available until {{access_until}}. | Manage subscription | Hi {{customer_name}}, your {{plan_name}} subscription has been cancelled. Access remains available until {{access_until}}. |
| `subscription_on_hold` | MantrixFlow subscription on hold | Fix billing to restore full access. | Fix billing | Hi {{customer_name}}, your {{plan_name}} subscription is on hold. Reason: {{reason}}. |
| `subscription_expired` | MantrixFlow subscription expired | Reactivate to resume access. | Reactivate | Hi {{customer_name}}, your {{plan_name}} subscription expired on {{expired_at}}. |
| `refund_succeeded` | Refund completed for MantrixFlow | {{amount}} {{currency}} was refunded. | View billing | Hi {{customer_name}}, your refund of {{amount}} {{currency}} is complete. Refund ID: {{refund_id}}. |
| `dispute_opened` | Dispute opened for MantrixFlow payment | Admin review required. | Open Dodo | A dispute was opened for payment {{payment_id}} from {{customer_email}} for {{amount}} {{currency}}. |
| `invoice_available` | Your MantrixFlow invoice is available | Invoice {{invoice_number}} is ready. | View invoice | Hi {{customer_name}}, invoice {{invoice_number}} for {{amount}} {{currency}} is available. |
