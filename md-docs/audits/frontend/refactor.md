# Frontend refactor audit and migration plan

Date: 2026-07-21  
Scope: `apps/app` (Next.js 16 App Router, React 19, TypeScript, Tailwind CSS,
shadcn/ui, TanStack Query, TanStack Table)

This document is the pre-implementation audit and the controlling plan for the
incremental frontend refactor. The compatibility rule is strict: routes, query
parameters, API request/response contracts, permissions, plan gates, product
copy, interactions, and ELT pipeline contracts remain unchanged unless a later
change is explicitly documented and verified.

## Executive summary

- The frontend contains 405 TypeScript/TSX source files, 47 page routes, and 57
  existing shadcn UI primitives.
- 33 source files are at least 500 lines; five are at least 1,000 lines.
- The largest file, the pipeline operations page, is 3,492 lines and combines
  URL state, 20+ server queries/mutations, transformations, permissions,
  dialogs, tables, preview state, GitHub synchronization, and rendering.
- A reusable TanStack `DataTable` already exists and is used by eight product
  screens, but it is 969 lines and owns toolbar, persistence, sorting,
  filtering, selection, table rendering, loading/error/empty states, and
  pagination in one module.
- At least four product data tables and several data-preview/schema tables
  still implement sorting, filtering, pagination, and empty states manually.
- Dataset configuration is implemented twice in closely related 925-line and
  1,100-line modules. The common schema, column DnD items, mock column fetch,
  selection logic, and form layout should be extracted once while retaining
  the route-specific differences.
- Shared feedback, layout, confirmation, form-section, and table primitives
  already exist. The refactor should finish and normalize these systems rather
  than introduce competing replacements.
- The pre-refactor production build passes. Lint, direct `tsc`, and the generic
  `bun test` command have pre-existing failures described below.

## Inventory and baseline

### Source inventory

| Item | Count |
| --- | ---: |
| TypeScript/TSX files in `app`, `components`, `hooks`, `lib`, `config` | 405 |
| App Router page files | 47 |
| shadcn UI component files | 57 |
| Existing screens importing shared `DataTable` | 8 |
| Files directly calling `useReactTable` | 1 |

### Files at least 1,000 lines

| Lines | File | Primary issue |
| ---: | --- | --- |
| 3,492 | `app/workspace/data-pipelines/[id]/_components/pipeline-operations-page.tsx` | Entire pipeline operator console in one client component |
| 2,409 | `app/workspace/settings/page.tsx` | Profile, workspace, billing, Slack, GitHub, UI, and 20 local state values |
| 1,624 | `lib/api/services/data-sources.service.ts` | Multiple data-source domains and legacy/current APIs in one class |
| 1,193 | `components/data-sources/constants.ts` | Connector metadata and large static catalogs mixed in a component tree |
| 1,100 | `app/workspace/data-sources/[id]/query/dataset-config.tsx` | Dataset form, list, DnD, bulk selection, mock fetch, and table |

### Files from 500 through 999 lines

| Lines | File |
| ---: | --- |
| 975 | `lib/api/hooks/use-data-pipelines.ts` |
| 969 | `components/shared/data-table/data-table.tsx` |
| 925 | `app/workspace/datasets/[id]/page.tsx` |
| 892 | `components/ui/onboarding.tsx` |
| 843 | `app/workspace/page.tsx` |
| 831 | `app/workspace/connections/components/CredentialForm.tsx` |
| 824 | `components/data-sources/saas-source-preview.tsx` |
| 813 | `lib/actions/auth.ts` |
| 794 | `config/transform-examples.ts` |
| 784 | `app/workspace/team/page.tsx` |
| 757 | `lib/api/types/data-pipelines.ts` |
| 733 | `app/workspace/data-pipelines/[id]/transformations/[transformationId]/page.tsx` |
| 727 | `components/product-tour/product-tour-provider.tsx` |
| 726 | `components/ui/sidebar.tsx` |
| 668 | `lib/api/hooks/use-data-sources.ts` |
| 621 | `app/onboarding/welcome/page.tsx` |
| 589 | `components/workspace/workspace-sidebar.tsx` |
| 586 | `lib/api/services/data-pipelines.service.ts` |
| 578 | `components/data-sources/connection-sheet.tsx` |
| 576 | `app/workspace/data-pipelines/[id]/destinations/[destinationId]/page.tsx` |
| 563 | `app/workspace/activity/page.tsx` |
| 552 | `app/workspace/data-sources/[id]/query/view/page.tsx` |
| 542 | `app/workspace/data-sources/page.tsx` |
| 534 | `app/workspace/connections/data/connectionFields.ts` |
| 534 | `app/workspace/connections/components/ConnectionList.tsx` |
| 523 | `lib/api/hooks/use-destination-schemas.ts` |
| 519 | `config/connectors.ts` |
| 517 | `config/database-registry.ts` |

Large static registries are lower-risk than large stateful client components,
but they still need domain-based files so feature code does not import an
unrelated catalog.

### Pre-refactor quality baseline

| Check | Baseline | Evidence |
| --- | --- | --- |
| `bun run build` | Passes with warnings | All 46 static pages generated; DuckDB dynamic dependency warning |
| `bun run lint` | Fails before refactor | Biome formatting only: two workflow JSON fixtures |
| `tsc --noEmit` | Fails before refactor | Three test files import undeclared `vitest`; their callback parameters then infer `any` |
| `bun test` | Wrong runner / fails | 43 unit assertions pass; Bun also loads 10 Playwright specs outside Playwright |
| Playwright | Not run in audit | Requires the local authenticated backend/ELT test environment and credentials |

Additional baseline warnings:

- `next@16.0.10` and React `19.2.1` are below the versions recommended by the
  current framework security guidance. Upgrade is a separate, controlled
  dependency change because it is not behavior-neutral.
- `baseline-browser-mapping` data is stale.
- `@duckdb/duckdb-wasm` emits a webpack dynamic-dependency warning through the
  SQLRooms explorer route.

## Duplicate-component report

### Tables

Shared `DataTable` consumers already include organizations, activity,
connections, pipelines, data sources, notifications, dashboard content, and
team members. The following implementations still render table mechanics
directly and should be migrated only after table parity tests exist:

- `components/data-sources/data-source-table.tsx`: manual search, status filter,
  five sortable headers, actions, and empty state.
- `components/data-sources/enhanced-result-viewer.tsx`: manual sorting,
  pagination, sticky header, horizontal scroll, dynamic cells, and empty state.
- `components/data-pipelines/data-preview-table.tsx`: dynamic preview columns,
  loading rows, badges, refresh, and horizontal scrolling.
- `app/workspace/data-pipelines/[id]/_components/pipeline-operations-page.tsx`:
  bespoke run, transformation, destination, source, and relationship tables
  plus a local `TablePager`.
- `app/workspace/data-sources/[id]/query/dataset-config.tsx`: dataset list table
  plus column-selection list behavior.
- `components/data-sources/data-source-preview-dialog.tsx`,
  `components/data-sources/saas-source-preview.tsx`, `schema-view.tsx`, and
  `incoming-data-tree-view.tsx`: specialized preview/schema tabular displays.
- `app/workspace/analytics/components/TopPipelinesTable.tsx` and
  `RecentFailedRunsTable.tsx`: repeated card/table/skeleton/empty-state shells.

Not every preview grid should expose every CRUD-table feature. All should share
the same TanStack rendering primitives, column headers, state shells, and
scroll/accessibility behavior while keeping feature logic outside the generic
table.

### Forms and validation

- `app/workspace/datasets/[id]/page.tsx` and
  `app/workspace/data-sources/[id]/query/dataset-config.tsx` duplicate the
  dataset Zod schema, `fetchColumns`, column icons, sortable/available column
  items, DnD and selection behavior, and most of the form UI.
- Authentication forms share labeled fields, server-action errors, pending
  buttons, and loading placeholders but use several one-off shells.
- Organization create/edit, team invite/edit, connection credentials, and
  data-source connection sheet repeat label/description/error/action layouts.
- `CredentialForm.tsx` uses an untyped record-shaped local form with manual
  validation/test state despite React Hook Form and Zod already being present.
- Validation is spread among `lib/validations/auth.ts`, action files, hook-side
  response checks, inline schemas, and manual conditional branches. API
  response validation and user-input validation need distinct schema folders.

### Dialogs, sheets, and confirmations

- Ten product modules use `ConfirmationModal`/`AlertDialog`; pipeline
  operations holds pipeline, destination, transformation, run cancellation,
  field, and GitHub confirmation state in the page.
- Team list and team detail independently implement remove and role-change
  confirmations and status badges.
- Connection drawer, connection sheet, destination drawer, data-preview
  dialog, activity detail sheet, and settings integrations repeat controlled
  open/close, pending action, header, description, and footer structure.
- The correct direction is one controlled dialog instance per feature with a
  selected entity, preserving all existing confirmation copy and pending/error
  behavior.

### Search, filters, and pagination

- Search controls are repeated across pipeline operations, activity,
  connections, data sources, team, pipeline list, schema navigation, table
  navigation, and explorer panels.
- URL-synchronized pagination/sort/search is already standardized with `nuqs`
  in organizations, activity, team, pipelines, and connections, but each page
  repeats state-to-API mapping and reset-to-first-page behavior.
- `DataTable` owns generic pagination while pipeline operations and result
  viewer implement separate pagers. `components/ui/pagination.tsx` is not the
  shared product-table pagination implementation.
- Filter definitions and renderers remain page-specific even when the toolbar
  layout and reset behavior are the same.

### Status badges

- Pipeline status has `components/pipeline/RunStatusBadge.tsx`,
  `SyncModeBadge.tsx`, `lib/constants/data-pipelines.ts`, a local
  `statusBadge()` in pipeline operations, and a separate
  `getStatusConfig()` in `pipeline-run-tracker.tsx`.
- Team list and team detail independently render role/status badges.
- Connection state badges are repeated in connection cards, list rows, data
  source cards, and legacy data-source table.
- Notification status configuration is repeated between the notifications page
  and notification sheet.
- Activity status is the strongest current pattern: a domain mapper in
  `lib/utils/activity-log-display.ts` and a shared badge component.

### Loading, empty, and error states

The codebase has reusable `LoadingState`, `ErrorState`, `EmptyState`, and
skeleton components, but direct `Loading...`, `No data available`, `No schemas
found`, and hand-built skeleton card/table shells remain in auth, analytics,
data-source, pipeline, and explorer modules. Shared feedback primitives need
variants for page, section, table, search-empty, permission, plan, connection,
and pipeline-run contexts.

### API request and response logic

- Pipeline operations imports service classes directly, creates ad-hoc
  `useQuery`/`useMutation` calls, and also consumes the feature hooks. It fetches
  the same pipeline through `usePipelineWithSchemas` and `usePipeline`, then
  manually merges and refetches both results.
- `lib/api/hooks/use-data-pipelines.ts` combines create/read/update/delete,
  run/pause/resume/cancel, validation/dry-run, run history, stats, and sync-state
  hooks in one 975-line file.
- `DataSourcesService` contains discovery, preview, schemas, connections,
  legacy compatibility, and mutations in one 1,624-line class.
- `app/workspace/data-sources/page.tsx` redirects to connections but still
  mounts legacy queries, transformations, dialogs, and mutation handlers while
  the redirect effect runs. That code is obsolete for normal navigation and
  should be removed only after route/redirect verification.
- Authentication and onboarding call service APIs from server actions and
  client forms with repeated response/error normalization.

### Validation and permission logic

- Owner/role restrictions are repeated in settings, team list, and team detail.
  The checks protect role editing, ownership transfer, removal, Slack settings,
  and some GitHub actions; they must be centralized as pure selectors without
  weakening UI or backend enforcement.
- Plan access is distributed among workspace layout, billing hooks, create
  organization dialogs, sidebar/topbar UI, and plan-limit events/dialogs.
- Pipeline readiness is computed inline from streams, destinations,
  transformations, connection status, and published revisions in the 3,492-line
  page. It should become a tested domain selector while preserving strict ELT
  invariants and exact disabled states.
- Dataset schemas are duplicated, while connection credentials rely heavily on
  manual required-field checks and DTO builder tests.

## Excessive state and mixed responsibilities

| Component | Evidence | Extraction boundary |
| --- | --- | --- |
| Settings page | 20 `useState` calls and many billing/Slack/GitHub/profile/workspace mutations | Route shell + profile/workspace/billing/integration panels and hooks |
| Pipeline operations | Multiple URL filters, 20+ queries/mutations, local forms, dialogs, previews, GitHub, derived readiness, five table views | Feature route controller + tab modules + domain hooks/selectors/dialog controller |
| Dataset configuration (two modules) | Form state, DnD state, async mock columns, dataset CRUD, bulk selection, table | Shared dataset feature components/schema/hook; route adapters |
| Credential form | Record form state, field metadata, test/create/update flows, setup guides, conditional field renderer | Zod schema factory, RHF hook, service mutations, field components, guides |
| Team page | Query mapping, permission checks, column definitions, three confirmation flows, edit dialog | Team model mapper, access selectors, columns, action-dialog controller |
| SaaS preview | Discovery, resource selection, preview query, badges, table, loading/empty/error UI | Preview hook + resource navigator + generic preview table |

## Dead, unused, legacy, and inconsistent code

Confirmed by `tsc --noUnusedLocals --noUnusedParameters` (after the undeclared
test dependency errors):

- `_executeQuery` and `_fetchTableData` in the query view page.
- `_workspaceSlug` in settings.
- `_hashType` in accept-invite.
- `_databases` in connection wizard.
- `_router` in the legacy data-source table.
- Eleven unused connector Zod schema constants in `config/connectors.ts`.
- Unused onboarding/team action results and placeholder IDs.
- `lib/api/EXAMPLE_USAGE.tsx` is documentation-like dead executable source.

Likely obsolete or transitional, requiring behavior verification before removal:

- `app/workspace/data-pipelines/mockListData.ts` has no observed import.
- `app/workspace/connections/data/mockConnections.ts` has no observed import.
- `lib/utils/mock-data-service.ts` appears isolated.
- Dataset/query/onboarding modules still generate mock tables or rows.
- Data sources route redirects to connections but retains a full legacy page.
- `DataSourceService` and `DataSourcesService`, plus singular/plural hooks and
  types, create confusing boundaries.
- Imports alternate between `@/components/shared`, deep paths, and direct file
  paths. File naming alternates between PascalCase and kebab-case inside the
  same feature.

Legacy pipeline query compatibility (`legacyDestination`, legacy stream parsing)
is intentional and must not be removed during structural cleanup.

## Target architecture

The app currently uses `apps/app` without a `src` directory. Moving every file
under `src` would create route/import churn without product value, so the
refactor keeps the existing root and introduces clear feature boundaries:

```text
apps/app/
  app/                         # route shells; URL behavior stays here
  components/
    ui/                        # owned shadcn primitives only
    shared/
      data-table/
      dialogs/
      feedback/
      forms/
      layout/
      navigation/
  features/
    connections/
      components/
      hooks/
      schemas/
      services/
      types/
      utils/
    pipelines/
      components/
      hooks/
      schemas/
      services/
      types/
      utils/
    sources/
    datasets/
    runs/
    workspace/
    billing/
    agents/
  hooks/                       # truly cross-feature client hooks
  lib/
    api/                       # API client and compatibility re-exports
    constants/
    nuqs/
    stores/
    utils/
  __tests__/
    unit/
    integration/
    playwright/
```

Route-local components may remain under `app/**/_components` when they are not
reused and exist only to assemble that route. Feature business rules must not
live in route pages or generic UI components.

## Reusable component architecture

### Shared shadcn/product primitives

Retain existing shadcn source components. Add or normalize product wrappers only
where consistent behavior exists:

- `SearchInput`
- `PageHeader` (existing; extend through slots, not copies)
- `FormFieldWrapper`, `FormSection`, `FormActions` (existing)
- `AppDialog`
- `ConfirmActionDialog` / `DeleteConfirmationDialog` (evolve existing
  `ConfirmationModal` without breaking callers)
- `StatusBadge` with domain adapters
- `PermissionGuard` and `PlanFeatureGuard`
- `PageLoading`, `SectionLoading`, `TableLoading`
- `PageError`, `InlineError`
- `EmptyState`, `NoSearchResults`, `PermissionDenied`, `PlanUpgradeRequired`
- `ConnectionError`, `PipelineRunError`

### TanStack Table system

Decompose the current `DataTable` while initially retaining its public props:

```text
components/shared/data-table/
  data-table.tsx               # orchestration and compatibility facade
  data-table-types.ts
  data-table-state.ts          # controlled/uncontrolled adapters
  data-table-storage.ts        # versioned visibility persistence
  data-table-toolbar.tsx
  data-table-search.tsx
  data-table-column-header.tsx
  data-table-view-options.tsx
  data-table-faceted-filter.tsx
  data-table-pagination.tsx
  data-table-row-actions.tsx
  data-table-content.tsx
  data-table-skeleton.tsx
  data-table-empty-state.tsx
```

The table facade owns TanStack wiring and rendering mechanics. Feature tables
provide columns, data, controlled server/client state, filters, actions,
permissions, and copy. Column ordering, sticky headers, responsive horizontal
scrolling, bulk actions, URL state, and custom cell renderers are opt-in props;
no pipeline/team/connection rules enter the generic component.

### Form architecture

```text
features/<feature>/schemas/<entity>.schema.ts
features/<feature>/utils/<entity>-form-mappers.ts
features/<feature>/hooks/use-<entity>-form.ts
features/<feature>/components/<Entity>Form.tsx
components/shared/forms/fields/
```

Shared fields initially cover text, password, number, select, checkbox, switch,
and date controls. Cron, SQL, JSON, and connection credentials remain focused
wrappers around their existing editors. Schemas validate user input; service
response mappers validate/narrow API data separately. Submit mutations stay in
feature hooks and API contracts remain byte-for-byte compatible at the boundary.

### Dialog architecture

Each feature gets a small dialog controller state such as
`{ type, entity } | null`; a table renders no per-row dialog instances. Generic
confirm/delete wrappers handle focus, accessible description, pending state,
cancel, and error retention. Feature dialogs retain domain-specific fields,
copy, permission checks, and mutation hooks.

### API and query architecture

- Keep the existing authenticated `ApiClient` and base URL behavior.
- Split large services by resource while preserving methods through temporary
  compatibility exports.
- Split query keys from query/mutation hooks and colocate hooks by feature.
- Add missing feature hooks for the ad-hoc pipeline operation queries.
- Normalize query invalidation in one feature-level helper.
- Pass abort signals when supported by the existing client; do not invent
  retries/timeouts that alter mutation semantics.
- Never call the Python ELT service directly from product frontend flows.

## Incremental refactoring plan by feature

Every phase starts with a behavior inventory, preserves public exports during
migration, runs targeted tests/typecheck/lint/build, and removes the old path
only after its consumers pass.

### Phase 0 — baseline and safety net

1. Separate unit and Playwright commands; make unit test types resolvable.
2. Fix the two formatting-only lint failures.
3. Add characterization tests for table state/pagination, pipeline readiness,
   connection DTO mapping, dataset mapping, roles, and status adapters.
4. Record critical routes, query parameters, API calls, and screenshots.

### Phase 1 — shared primitives and DataTable

1. Extract DataTable types, visibility persistence, column header, toolbar,
   content, and pagination without changing current consumers.
2. Add tests for controlled/uncontrolled pagination, manual filtering/sorting,
   selection, loading/error/empty states, fixed/visible columns, and storage.
3. Add `SearchInput`, status adapters, and missing feedback variants.
4. Migrate analytics tables and the legacy data-source table first; they are
   lower risk than the pipeline operator.

### Phase 2 — connections and sources

1. Extract connection form schema factory, defaults, DTO mappers, field
   components, test/create/update hook, and setup guides.
2. Keep all current credential fields and never log or expose secrets.
3. Consolidate connection/data-source list actions and status badges.
4. Split discovery/preview/schema services and hooks behind compatibility
   exports.
5. Verify create/test/edit/delete, source/destination roles, private database
   guidance, SaaS connectors, search, pagination, permissions, and errors.

### Phase 3 — datasets and explorer

1. Extract the duplicated dataset schema, DnD column selector, field list,
   form, defaults, and mappers.
2. Preserve route-specific source types (`custom_query` versus `saved_query`)
   through explicit adapters rather than merging their business rules.
3. Replace mock data only when a real existing API path is confirmed; structural
   refactoring alone must preserve current behavior.
4. Consolidate result/preview tables and query loading/error/empty states.

### Phase 4 — workspace, team, activity, notifications, organizations

1. Extract team access selectors, API-to-row mapper, columns, and controlled
   action dialogs.
2. Consolidate role/status badges while retaining owner restrictions.
3. Extract common URL table-state hooks and filter toolbar layout.
4. Remove redirected legacy data-source UI only after direct-navigation and
   redirect query-parameter verification.

### Phase 5 — settings, billing, integrations, onboarding, auth

1. Split settings into route shell plus profile, workspace, billing, Slack,
   GitHub, and appearance panels with focused hooks.
2. Preserve owner-only actions, plan state, checkout/portal URLs, OAuth return
   parameters, pending-link claims, and notification behavior.
3. Normalize auth/onboarding form shells and validation without changing server
   actions or callback URLs.
4. Defer large static onboarding/sidebar visual primitives unless they contain
   duplicated business behavior.

### Phase 6 — pipelines and runs

1. Add pure, tested selectors for pipeline normalization, readiness, status,
   transformation counts, and visible source rows.
2. Split API hooks into pipeline, transformation, destination, run, schedule,
   sync-state, and GitHub modules while preserving query keys.
3. Split the operations page by tab and add one dialog controller.
4. Migrate runs, sources, transformations, and destinations to the shared table
   primitives with their existing URL-synced parameters.
5. Keep `selected_streams` as strict `SourceStreamConfig[]`, use only the
   sanctioned DuckDB table-name helper, retain `schema.table` destinations,
   preserve upsert-only behavior, and retain legacy redirect compatibility.
6. Verify create, schema discovery, preview, SQL validation, mapping, run,
   realtime status, retry/cancel/pause/resume, edit/delete, GitHub sync, errors,
   plan gates, and permissions.

### Phase 7 — cleanup and full verification

1. Remove compatibility re-exports and duplicated modules only after the final
   consumer migrates.
2. Run unused-export analysis, strict TypeScript, Biome, unit, integration,
   build, and the complete Playwright matrix.
3. Compare route inventory, URL/query behavior, API request snapshots,
   screenshots, responsive breakpoints, keyboard/focus behavior, and accessible
   names with the baseline.
4. Produce before/after line counts, removed-code totals, verification results,
   and remaining debt.

## Per-feature verification matrix

For every migrated feature record Pass/Fail/Not applicable for:

- Create, read, update, delete
- Search, filters, sort, pagination, selection, bulk actions
- Dialogs, sheets, notifications, and keyboard/focus behavior
- Permissions, owner restrictions, and plan restrictions
- Loading, empty, search-empty, error, retry, and disabled states
- API URL, method, query parameters, request payload, and response mapping
- Mobile, tablet, laptop, desktop, and large desktop layout
- Unit, component, integration, production build, and Playwright coverage

## Definition of done

The refactor is complete only when all feature phases are migrated, the strict
ELT invariants review is clean, routes/API/query behavior are unchanged, all
supported checks pass, critical authenticated Playwright flows pass against the
full local stack, and the final report contains the requested file list, line
counts, duplication delta, functional verification evidence, test results, and
remaining technical debt.

## Implementation checkpoint — 2026-07-21

This checkpoint is an incremental milestone, not the final definition of done.

### Completed slices

- Split the shared TanStack table facade into focused types, storage, search,
  toolbar, view-options, column-header, content, empty-state, and pagination
  modules while preserving its existing public props.
- Added stable row IDs and optional pagination to support controlled feature
  selection without index-based row identity.
- Migrated analytics top-pipeline and recent-failure tables, the query result
  viewer, pipeline data preview, and embedded dataset list to the shared table.
- Removed the confirmed-unused legacy `data-source-table.tsx` and its barrel
  export after a repository-wide consumer search.
- Split connection credential rendering and setup guidance from the controller;
  moved edit hydration and DTO mapping to tested helpers.
- Consolidated both dataset routes around one Zod schema, RHF field component,
  column-discovery service boundary, DnD selector, tested selection model, and
  embedded TanStack table.
- Split settings billing configuration, integrations catalog, profile,
  notification, and security views into responsibility-focused modules.
- Added a dedicated unit-test command so Bun no longer discovers Playwright
  specs as unit tests, and restored direct TypeScript test resolution.

### Before/after line counts for migrated large files

| File | Before | Checkpoint | Delta |
| --- | ---: | ---: | ---: |
| `components/shared/data-table/data-table.tsx` | 969 | 405 | -564 |
| `app/workspace/connections/components/CredentialForm.tsx` | 831 | 319 | -512 |
| `app/workspace/datasets/[id]/page.tsx` | 925 | 367 | -558 |
| `app/workspace/data-sources/[id]/query/dataset-config.tsx` | 1,100 | 407 | -693 |
| `app/workspace/settings/page.tsx` | 2,409 | 1,886 | -523 |
| `components/data-sources/enhanced-result-viewer.tsx` | 367 | 277 | -90 |
| `components/data-pipelines/data-preview-table.tsx` | 255 | 227 | -28 |

New focused modules hold the extracted behavior, so these deltas represent
reduced responsibility and duplicate code rather than simple deletion.

### Verification at this checkpoint

- TypeScript: pass (`tsc --noEmit`).
- Biome: pass across 489 files.
- Unit tests: 55 pass, 0 fail, 63 expectations.
- Next.js production build: pass; all 46 existing routes remain in the build.
- Known unchanged build warnings: DuckDB WASM's expression-based dynamic
  dependency and stale `baseline-browser-mapping` data.
- Authenticated Playwright flows were not run at this checkpoint because the
  full local frontend, Go API, ELT service, database, and test credentials were
  not started as part of this slice.

### Remaining critical work

- Split the 3,492-line pipeline operations controller by tab, hook, service,
  selector, and controlled dialog, then migrate its manual tables.
- Finish settings organization, billing, Slack, and GitHub panel extraction.
- Migrate the remaining feature tables listed by the audit; `ui/table.tsx`,
  skeleton tables, and semantic detail grids are intentionally not migration
  targets.
- Refactor the remaining files above 500 lines (workspace, team, activity,
  onboarding, source preview, connection sheet, transformation, destination,
  and query-view flows).
- Complete API/service consolidation, form-schema extraction, permission/plan
  guards, unused-export cleanup, component/integration coverage, authenticated
  Playwright coverage, responsive screenshots, and accessibility verification.
