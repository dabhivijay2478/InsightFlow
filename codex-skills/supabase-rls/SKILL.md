---
name: supabase-rls
description: Use when adding or extending Supabase row-level security in this repo. Covers MantrixFlow's current-organization tenancy model, safe handling of secret-bearing tables, and the manual maintenance workflow for apps/server/main-server/sql/supabase_rls.sql.
---

# Supabase RLS

Use this skill when a table in `apps/server/main-server/internal/models/` changes or a new table needs client-safe Supabase access.

## Workflow

1. Read the relevant model file under `apps/server/main-server/internal/models/`.
2. Classify the table:
   - `org_owned_crud` or `org_owned_select` for direct `organization_id` ownership
   - `self_user` or `self_user_crud` for `auth.uid()` ownership
   - `custom` when org access comes from joins or nullable ownership columns
   - `no_client_access` for secret-bearing tables such as `data_source_connections`
3. Update `apps/server/main-server/sql/supabase_rls.sql` directly.
4. Read `apps/server/main-server/docs/RLS_GUIDE.md` and keep the new table aligned with its checklist.
5. For local application, use `APPLY_SUPABASE_RLS_ONCE=true` when starting the Go API. The server applies the SQL through GORM and records the checksum as a comment on `public.app_user_org_id()`.
6. If a new ownership pattern was introduced, update `references/schema-map.md` so later agents do not have to rediscover it.

## Project Rules

- Tenancy is based on `users.current_organization_id`, not `users.org_id`.
- `public.app_user_org_id()` must only return an org the user actively belongs to.
- The Go API writes through a direct database connection without `auth.uid()`, so org/user enforcement triggers must no-op when `auth.uid()` is null.
- Do not add `FORCE ROW LEVEL SECURITY`; the backend must keep bypassing RLS on its owner/service connection.
- Do not expose `data_source_connections.config` through a base-table select policy. Keep the base table closed unless a real client-safe view is needed.

## References

- Read `references/schema-map.md` for the current table map and policy rationale.
- Read `apps/server/main-server/docs/RLS_GUIDE.md` for the maintenance checklist and local apply flow.
- The file that gets applied to Supabase is `apps/server/main-server/sql/supabase_rls.sql`.
