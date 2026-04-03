# MantrixFlow RLS Schema Map

Current public tables created by GORM live in `apps/server/main-server/internal/models/`.

## Current org-scoped tables

- `organizations`
  Current-org read only. The active org is `users.current_organization_id`.
- `organization_members`
  Current-org read only. Used to verify active membership.
- `data_sources`
  Direct `organization_id` ownership. Safe for org-scoped CRUD.
- `data_source_schema_metadata`
  Direct `organization_id` ownership. Client read only.
- `pipelines`
  Direct `organization_id` ownership. Safe for org-scoped CRUD.
- `pipeline_runs`
  Direct `organization_id` ownership. Client read only for dashboards and Realtime.
- `activity_logs`
  Direct `organization_id` ownership. Client read only.

## Derived org-scoped tables

- `pipeline_source_schemas`
  Org access comes from `organization_id` when present, otherwise from `data_source_id -> data_sources.organization_id`.
- `pipeline_destination_schemas`
  Same pattern as `pipeline_source_schemas`.

## User-scoped tables

- `users`
  Self read/update only. `current_organization_id` may only point to an org the user actively belongs to.
- `email_preferences`
  Self CRUD by `user_id = auth.uid()`.

## Secret-bearing tables

- `data_source_connections`
  Contains encrypted `config`. Keep the base table closed to client selects. Only add a dedicated safe view later if a real client read path needs it.

## Trigger rule

`public.enforce_authenticated_org_id()` and `public.enforce_authenticated_user_id()` only rewrite ownership columns when `auth.uid()` is present. This is required because the Go API writes through a direct Postgres connection and does not have Supabase JWT context inside SQL triggers.

## Local apply flow

The committed SQL bundle lives at `apps/server/main-server/sql/supabase_rls.sql`. Local developers do not need `psql` for normal setup; the Go API can apply it once on startup with `APPLY_SUPABASE_RLS_ONCE=true`. The apply marker is stored as a checksum comment on `public.app_user_org_id()`.

See `apps/server/main-server/docs/RLS_GUIDE.md` for the full update checklist when adding a new table.
