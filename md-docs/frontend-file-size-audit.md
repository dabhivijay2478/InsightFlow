# MantrixFlow frontend file-size audit

Date: 2026-07-22  
Scope: maintained production source under `apps/app` with extensions `.ts`, `.tsx`, `.js`, `.jsx`, `.css`, `.scss`, and `.less`  
Excluded: `node_modules`, `.next`, build/output/coverage folders, public/vendor assets, generated files, lock files, and existing test/spec files. Tests are intentionally excluded because this task explicitly prohibits creating or restructuring test files.

## Baseline

- Maintained production source files scanned: **436**
- Maintained production source lines scanned: **73,948**
- Files above 500 lines: **28**
- Largest maintained source file: **3,496 lines**
- Existing UI foundation: shadcn/ui primitives, TanStack Query, TanStack Table, React Hook Form, Zod, Zustand, and `nuqs`
- Existing shared table foundation: `components/shared/data-table/*`

## Files above 500 lines

| Lines | File | Classification | Responsibilities and duplication | Recommended split | Estimated resulting files |
| ---: | --- | --- | --- | --- | ---: |
| 3,496 | `app/workspace/data-pipelines/[id]/_components/pipeline-operations-page.tsx` | PAGE TOO LARGE; MIXED RESPONSIBILITIES; DUPLICATED UI; DUPLICATED TABLE; INLINE TYPES; INLINE BUSINESS LOGIC; DUPLICATED STATE | Route state, 15+ queries/mutations, readiness rules, source preview, stream selection, transformation/destination/run tables, GitHub sync, dialogs, status formatting, pagination, and thousands of lines of JSX. Reimplements status badges and paging already available in shared components. | Thin page container; operations hook; query state; domain selectors; overview/source/transformation/destination/run/GitHub tab components; column files; controlled dialogs; formatters. Migrate its repeated tables to shared `DataTable`. | 24–30 |
| 1,886 | `app/workspace/settings/page.tsx` | PAGE TOO LARGE; MIXED RESPONSIBILITIES; DUPLICATED FORM; DUPLICATED API LOGIC; DUPLICATED STATE; INLINE BUSINESS LOGIC | Profile, workspace, billing, Slack, GitHub, appearance, plan selection, and integration mutations are coupled in one page. Repeats save/error/toast and derived integration-state logic. | Thin route; settings container; focused hooks for profile/workspace/billing/Slack/GitHub; one component per tab; integration row models; billing helpers. Reuse existing settings UI and shadcn form/dialog primitives. | 12–16 |
| 1,624 | `lib/api/services/data-sources.service.ts` | MIXED RESPONSIBILITIES; DUPLICATED API LOGIC; INLINE TYPES; INLINE BUSINESS LOGIC | Data-source CRUD, connection lifecycle, schema discovery/normalization, preview, query execution, sync jobs, health, logs, and metrics. Contains legacy aliases and large response normalizers. | Preserve `DataSourcesService` compatibility facade while extracting CRUD, connection, discovery, query, sync, monitoring, and discovery-normalizer modules. | 8–10 |
| 1,193 | `components/data-sources/constants.ts` | MIXED RESPONSIBILITIES; DUPLICATED FORM; INLINE VALIDATION | Connection field schemas and complete connector catalog are combined. Overlaps `config/connectors.ts`, `config/database-registry.ts`, and connection form metadata. | Connector-family schema modules plus catalog metadata and a compatibility export. Centralize field builders/types. | 6–8 |
| 975 | `lib/api/hooks/use-data-pipelines.ts` | MIXED RESPONSIBILITIES; DUPLICATED API LOGIC; DUPLICATED STATE | Query keys, CRUD, execution, run control, validation, dry-run, run queries, sync-state reset, and stats. Mutation invalidation logic is repeated. | Keys/cache helpers; CRUD hooks; execution hooks; run hooks; validation hooks; sync-state hooks; stats hook; compatibility exports. | 7–9 |
| 892 | `components/ui/onboarding.tsx` | COMPONENT TOO LARGE; MIXED RESPONSIBILITIES; DUPLICATED UI | Three independent compound component systems—onboarding flow, choice group, feature carousel—plus tips list and step indicator in one UI file. | Separate onboarding flow, choice group, feature carousel, tips list, and step indicator modules; keep a compatibility export module. | 5–6 |
| 843 | `app/workspace/page.tsx` | PAGE TOO LARGE; MIXED RESPONSIBILITIES; DUPLICATED TABLE; INLINE TYPES; INLINE BUSINESS LOGIC | Dashboard data fetching, metrics, usage formatting, gauge, pipeline performance, actions, recent activity, migrations, and inline column definitions. | Thin page; dashboard container/hook; metric, usage, performance, actions, activity, and migration sections; table column files; formatter utilities. | 9–12 |
| 824 | `components/data-sources/saas-source-preview.tsx` | COMPONENT TOO LARGE; MIXED RESPONSIBILITIES; DUPLICATED TABLE; INLINE TYPES; INLINE BUSINESS LOGIC | Large SaaS resource catalog, docs/labels, fetch behavior, selection, preview formatting, and manual table rendering. | Resource metadata modules, preview hook, formatter, toolbar/resource list, and TanStack preview table configuration. | 7–9 |
| 813 | `lib/actions/auth.ts` | MIXED RESPONSIBILITIES; DUPLICATED API LOGIC; INLINE VALIDATION | Login, signup, forgot/reset/change password, invite acceptance, Supabase setup, API validation, and repeated error/redirect handling. | Shared auth action utilities plus one action module per workflow; compatibility exports preserve current imports. | 7–8 |
| 794 | `config/transform-examples.ts` | MIXED RESPONSIBILITIES; DUPLICATED STATE | One very large connector-keyed configuration object. Source-specific examples have independent ownership. | Shared type/helper plus one examples module per connector family and a composed registry. | 7–10 |
| 783 | `app/workspace/team/page.tsx` | PAGE TOO LARGE; MIXED RESPONSIBILITIES; DUPLICATED TABLE; DUPLICATED STATE; INLINE TYPES; INLINE BUSINESS LOGIC | Organization/member fetching, URL state, ownership rules, role/remove/transfer mutations, three confirmations, inline columns, filters, and page UI. | Thin page; team hook/controller; types/mappers; permissions; columns; toolbar; table; controlled dialogs. Continue using shared `DataTable`. | 7–9 |
| 757 | `lib/api/types/data-pipelines.ts` | MIXED RESPONSIBILITIES; INLINE TYPES | DTOs, source/destination domain types, graph types, transformations, runs, metadata, validation, discovery, previews, and generic API response types. | Split CRUD/config, graph, transformation, destination, run, validation/discovery, and shared response types; compatibility exports retain API. | 7–8 |
| 733 | `app/workspace/data-pipelines/[id]/transformations/[transformationId]/page.tsx` | PAGE TOO LARGE; MIXED RESPONSIBILITIES; DUPLICATED FORM; DUPLICATED API LOGIC; DUPLICATED STATE; INLINE VALIDATION; INLINE BUSINESS LOGIC | Route params, queries/mutations, editor state, placeholder rendering, validation, preview, revision/publish controls, and complete form JSX. | Thin route; transformation editor hook/service calls; form/type/defaults; SQL utilities; editor header; fields; preview; revisions; publish dialog. | 7–9 |
| 727 | `components/product-tour/product-tour-provider.tsx` | COMPONENT TOO LARGE; MIXED RESPONSIBILITIES; DUPLICATED STATE; INLINE BUSINESS LOGIC | Storage keys, step builders, card positioning, completion hooks, auto-start orchestration, route parsing, provider shell, and tour card UI. | Storage utilities; step definitions; viewport utility; completion hooks; tour card; auto-start controller; provider. | 7–8 |
| 726 | `components/ui/sidebar.tsx` | COMPONENT TOO LARGE; MIXED RESPONSIBILITIES | shadcn sidebar compound primitive, provider/state, layout primitives, groups, menus, skeleton, submenus, and exports. | Preserve the public API while separating provider/context, layout primitives, group primitives, and menu primitives. | 4–5 |
| 668 | `lib/api/hooks/use-data-sources.ts` | MIXED RESPONSIBILITIES; DUPLICATED API LOGIC; DUPLICATED STATE | Connection CRUD/test, database/schema/table discovery, query execution, sync jobs, monitoring, and repeated invalidation logic. | Keys/cache helpers; connection hooks; discovery hooks; query hooks; sync hooks; monitoring hooks; compatibility exports. | 6–7 |
| 618 | `app/onboarding/welcome/page.tsx` | PAGE TOO LARGE; MIXED RESPONSIBILITIES; DUPLICATED FORM; DUPLICATED STATE; INLINE VALIDATION; INLINE BUSINESS LOGIC | Three onboarding steps, Zod schema, RHF setup, organization creation, slug logic, onboarding mutations, session handling, and navigation. | Thin route; schema/defaults; onboarding controller; welcome/org/complete step components; navigation; session utility. | 7–8 |
| 589 | `components/workspace/workspace-sidebar.tsx` | COMPONENT TOO LARGE; MIXED RESPONSIBILITIES; DUPLICATED STATE; INLINE BUSINESS LOGIC | Organization switcher, navigation definitions/rendering, account menu, API/store synchronization, role filtering, onboarding state, and responsive behavior. | Organization switcher; navigation config/menu; account menu; synchronization hook; sidebar shell. | 5–6 |
| 586 | `lib/api/services/data-pipelines.service.ts` | MIXED RESPONSIBILITIES; DUPLICATED API LOGIC | Pipeline CRUD, structured graph entities, run control, validation, runs, stats, and sync state in one service. | CRUD, graph/configuration, execution, run, validation, and stats/sync modules with compatibility facade. | 6–7 |
| 578 | `components/data-sources/connection-sheet.tsx` | COMPONENT TOO LARGE; MIXED RESPONSIBILITIES; DUPLICATED FORM; DUPLICATED API LOGIC; DUPLICATED STATE; INLINE TYPES; INLINE VALIDATION | Dynamic Zod schema, defaults, credential field rendering, password visibility, connection testing, submit mapping, mutation flow, and sheet UI. | Schema/default builders; form hook; payload mapper; field section; test result; actions; controlled sheet. Use shadcn Form/Sheet fields. | 6–7 |
| 576 | `app/workspace/data-pipelines/[id]/destinations/[destinationId]/page.tsx` | PAGE TOO LARGE; MIXED RESPONSIBILITIES; DUPLICATED FORM; DUPLICATED API LOGIC; DUPLICATED STATE; INLINE VALIDATION; INLINE BUSINESS LOGIC | Route/query/mutations, destination form state, assignments, table creation, connection testing, validation, and multi-section JSX. | Thin route; destination editor hook; defaults/types; basic fields; assignment section; destination-table section; review/actions. | 6–8 |
| 563 | `app/workspace/activity/page.tsx` | PAGE TOO LARGE; MIXED RESPONSIBILITIES; DUPLICATED TABLE; INLINE TYPES; INLINE BUSINESS LOGIC | URL filters/sorting/pagination, metrics, inline columns, filter toolbar, details sheet state, and table composition. | Thin page; activity query hook; filter constants; columns; toolbar; metrics; table container/detail controller. Continue shared `DataTable`. | 6–7 |
| 552 | `app/workspace/data-sources/[id]/query/view/page.tsx` | PAGE TOO LARGE; MIXED RESPONSIBILITIES; DUPLICATED API LOGIC; DUPLICATED STATE; INLINE BUSINESS LOGIC | Route/session transfer, schema navigation, query/table execution, result normalization, export handling, loading/error states, and obsolete mock helpers. | Thin route; results controller hook; session parser; response mapper; header; navigation pane; result panel. Remove unreachable mock helpers. | 6–7 |
| 540 | `app/workspace/connections/components/ConnectionList.tsx` | COMPONENT TOO LARGE; MIXED RESPONSIBILITIES; DUPLICATED TABLE; DUPLICATED STATE; INLINE BUSINESS LOGIC | Connection actions/mutations, permission visibility, inline columns and row menu, filtering controls, empty state, table, and delete confirmation. | Connection action hook; columns; row-actions menu; toolbar; table container; controlled delete dialog. Continue shared `DataTable`. | 5–6 |
| 534 | `app/workspace/connections/data/connectionFields.ts` | MIXED RESPONSIBILITIES; DUPLICATED FORM; INLINE TYPES; INLINE VALIDATION | Types and all connector field definitions in one array; overlaps the two connector registries and source constants. | Shared field types/builders; database fields; SaaS fields; registry composer. | 4–5 |
| 523 | `lib/api/hooks/use-destination-schemas.ts` | MIXED RESPONSIBILITIES; DUPLICATED API LOGIC; DUPLICATED STATE | Destination schema CRUD, validation, table existence/creation, preview, SQL-model validation, and discovery with repeated cache invalidation. | Keys/cache helpers; CRUD hooks; validation/discovery hooks; table/preview hooks; compatibility exports. | 4–5 |
| 519 | `config/connectors.ts` | MIXED RESPONSIBILITIES; DUPLICATED FORM; INLINE TYPES; INLINE VALIDATION | Connector types plus unused database schemas and active connector metadata; overlaps database registry and data-source constants. | Types/builders, database schemas, SaaS schemas, connector registry. Remove or isolate truly unused definitions after reference verification. | 4–5 |
| 517 | `config/database-registry.ts` | MIXED RESPONSIBILITIES; DUPLICATED FORM; INLINE TYPES | Field builders, per-database fields, registry entries, availability policy, and selectors; overlaps connection field registries. | Shared types/builders, database field groups, registry entries, availability utilities. | 4–5 |

## Cross-cutting duplication found

1. **Connector metadata and credential fields** are defined independently in `components/data-sources/constants.ts`, `config/connectors.ts`, `config/database-registry.ts`, and `app/workspace/connections/data/connectionFields.ts`. They use similar field shapes but different names and defaults.
2. **Table configuration is buried in pages** for dashboard, team, activity, connections, and pipeline operations. Four already use the shared TanStack foundation but still mix business actions and column definitions into page/container files.
3. **Pipeline operations and SaaS preview tables** render table structures locally instead of supplying typed feature columns to the shared TanStack table.
4. **Mutation invalidation and toast/error handling** repeat throughout pipeline, connection, schema, team, and settings hooks.
5. **Connection forms** repeat schema construction, default-value building, credential visibility, connection testing, and payload mapping.
6. **Status, date, count, duration, and cell-value formatters** appear in several dashboards, tables, run views, and previews.
7. **Controlled destructive actions** use both `ConfirmationModal` and local dialog markup. These can converge on an application-level `ConfirmActionDialog` backed by shadcn `AlertDialog` without changing callbacks.
8. **Organization synchronization** is duplicated between page/sidebar consumers and the workspace store.
9. **Page-local URL query state** is repeated for search/sorting/pagination and should be expressed through focused feature query-state hooks built on the existing `nuqs` helpers.

## Shared architecture to retain and extend

### Reusable shadcn/application components

Existing shadcn primitives will remain the accessibility foundation. The refactor should converge feature code on:

- `PageHeader`, `SectionHeader`, `SectionSurface`, `PageContainer`
- `StatusBadge` / existing typed status badges
- `SearchInput`, filter toolbars, and shared empty/loading/error states
- `ConfirmActionDialog` / `DeleteConfirmationDialog` backed by `AlertDialog`
- `FormSection`, `FormActions`, reusable RHF field adapters, and existing password/secret controls
- `ResponsiveTableContainer` through the shared `DataTable` content layer

### TanStack Table architecture

The existing `components/shared/data-table` system already supplies typed generics, client/server pagination, sorting, global filtering, visibility persistence, row selection, loading/fetching/error/empty states, responsive rendering, and custom cells. It should be extended only where gaps are proven:

```text
components/shared/data-table/
  data-table.tsx                 # state orchestration
  data-table-content.tsx         # semantic shadcn Table rendering
  data-table-toolbar.tsx
  data-table-search.tsx
  data-table-pagination.tsx
  data-table-column-header.tsx
  data-table-view-options.tsx
  data-table-faceted-filter.tsx  # add when a migrated feature needs it
  data-table-row-actions.tsx     # generic trigger/shell only
  data-table-empty-state.tsx
  data-table-types.ts
```

Feature-specific columns and permissions remain outside this directory.

### Shared forms

```text
components/shared/forms/
  text-field.tsx
  password-field.tsx
  number-field.tsx
  select-field.tsx
  checkbox-field.tsx
  switch-field.tsx
  textarea-field.tsx
  form-section.tsx
  form-actions.tsx

features/<feature>/forms/
features/<feature>/schemas/
features/<feature>/utils/*-payload.ts
features/<feature>/hooks/use-*-form.ts
```

The refactor will reuse the installed React Hook Form, Zod, and shadcn Form primitives. Payload and response contracts remain unchanged.

### Shared dialogs

```text
components/shared/dialogs/
  confirm-action-dialog.tsx
  delete-confirmation-dialog.tsx

features/<feature>/dialogs/
  one controlled dialog per selected record/action
```

Dialogs are mounted once per feature container rather than once per row.

## Proposed feature structure

The existing App Router paths remain unchanged. Feature implementations move behind thin route entries:

```text
apps/app/
  app/                              # route/layout entry points only
  components/
    ui/                             # shadcn primitives
    shared/
      data-table/
      dialogs/
      feedback/
      forms/
      layout/
  features/
    auth/
    onboarding/
    dashboard/
    connections/
    data-sources/
    pipelines/
      operations/
      transformations/
      destinations/
      runs/
    activity/
    team/
    settings/
    product-tour/
  hooks/                            # genuinely cross-feature hooks
  lib/api/                          # shared client plus compatibility exports
  lib/types/                        # truly shared domain types only
  config/                           # composed registries, not UI logic
```

## Refactoring order

1. Split configuration, types, services, and hooks while preserving public exports.
2. Split oversized compound UI primitives (`onboarding`, `sidebar`).
3. Extract shared dialog/form helpers and extend the existing TanStack table system only as needed.
4. Refactor dashboard, team, activity, connections, onboarding, data-source preview/query, settings, and product-tour features.
5. Refactor pipeline destination/transformation editors, then the operations page last because it consumes most of the extracted contracts.
6. Recount every production source file after each feature batch.
7. Run only formatting, lint, TypeScript type-check/build-time checking, and production build.

## Invariants and risks

- Routes, search parameters, API paths, request/response payloads, permissions, roles, plans, toasts, loading/error/empty states, and responsive behavior must remain stable.
- Pipeline source identifiers remain `schema.table`; DuckDB staging names continue to use the sanctioned `duckdbTableNameForStream` helper.
- `selected_streams` remains structured and delivery remains Upsert-only; no legacy processing paths or user-facing ETL labels may be introduced.
- Service and hook compatibility exports are retained during migration so downstream consumers do not require uncontrolled rewrites.
- Existing tests are neither run nor modified during this task.
