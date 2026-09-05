# MantrixFlow frontend refactor report

Date: 2026-07-22  
Scope: maintained production source in `apps/arcyria-platform` with `.ts`, `.tsx`, `.js`, `.jsx`, `.css`, `.scss`, and `.less` extensions. Generated output, dependencies, locks, vendor assets, and test/spec files are excluded. Tests were neither created nor run.

## Outcome

- Maintained production source files scanned before: **436**
- Maintained production source files scanned after: **540**
- Maintained production source lines after: **75,488**
- Files above 500 lines before: **28**
- Files above 500 lines after: **0**
- Largest maintained production source file after: **500 lines** — `app/workspace/connections/components/ConnectionList.tsx`
- New focused source modules: **102** — 40 `.tsx` component/view modules and 62 `.ts` hook/service/type/config modules
- Duplicate implementations removed: **6 directly countable implementations** — five bespoke table renderers and one raw interactive button implementation
- Tables migrated to the shared TanStack Table foundation: **5** — SaaS source preview, pipeline source streams, transformations, destinations, and runs
- Custom interactive UI implementations replaced with shadcn/ui: **1 directly countable raw button implementation**; extracted dialogs, sheets, selects, switches, inputs, tabs, badges, alerts, and tooltips continue to use the existing shadcn primitives

The existing routes, navigation, URL query parameters, API request/response contracts, mutation payloads, permissions, organization roles, billing/plan rules, source discovery, previews, pipeline execution, run monitoring, GitHub/Slack integration flows, and responsive mobile card layouts were retained. No intentional business-rule or workflow changes were made.

## Original files above 500 lines

| Original lines | File | Refactoring result |
| ---: | --- | --- |
| 3,496 | `app/workspace/data-pipelines/[id]/_components/pipeline-operations-page.tsx` | Thin page plus typed model/context, query/source/GitHub/editor hooks, derivation utilities, seven tab modules, GitHub panels, dialogs, shared UI helpers, and four TanStack column configurations. |
| 1,886 | `app/workspace/settings/page.tsx` | 192-line page; organization, billing, integrations, Slack/GitHub drawers, callback URL helpers, badges, and integration controller extracted. |
| 1,624 | `lib/api/services/data-sources.service.ts` | Compatibility facade over CRUD, connection, preview, schema, legacy connection, operations, and discovery-normalizer services. |
| 1,193 | `components/data-sources/constants.ts` | Schema types plus database, storage, and SaaS schema catalogs. |
| 975 | `lib/api/hooks/use-data-pipelines.ts` | Compatibility export over query, CRUD, execution, validation, run, and sync hooks. |
| 892 | `components/ui/onboarding.tsx` | Compatibility export over flow, step indicator, choice group, carousel, and tips components. |
| 843 | `app/workspace/page.tsx` | Dashboard helpers, header/KPIs, quick actions, and recent sections extracted. |
| 824 | `components/data-sources/saas-source-preview.tsx` | Resource catalog and typed shared DataTable extracted. |
| 813 | `lib/actions/auth.ts` | Compatibility export over one server-action module per authentication workflow and shared result helpers. |
| 794 | `config/transform-examples.ts` | Typed SQL, commerce, CRM, and GitHub example registries. |
| 783 | `app/workspace/team/page.tsx` | Team member types/mappers and TanStack columns extracted; page is 499 lines. |
| 757 | `lib/api/types/data-pipelines.ts` | Compatibility export over common, config, core, entity, result, and run types. |
| 733 | `app/workspace/data-pipelines/[id]/transformations/[transformationId]/page.tsx` | Controller separated from the transformation editor view. |
| 727 | `components/product-tour/product-tour-provider.tsx` | Product-tour configuration extracted from provider orchestration. |
| 726 | `components/ui/sidebar.tsx` | Compatibility export over context, shell, layout, and menu primitives. |
| 668 | `lib/api/hooks/use-data-sources.ts` | Compatibility export over connection, discovery, query, sync, and monitoring hooks. |
| 618 | `app/onboarding/welcome/page.tsx` | Visible onboarding steps extracted. |
| 589 | `components/workspace/workspace-sidebar.tsx` | Organization switcher extracted. |
| 586 | `lib/api/services/data-pipelines.service.ts` | Pipeline configuration service separated while preserving the public facade. |
| 578 | `components/data-sources/connection-sheet.tsx` | Form schema/types/defaults and password-visibility state extracted. |
| 576 | `app/workspace/data-pipelines/[id]/destinations/[destinationId]/page.tsx` | Destination target card extracted. |
| 563 | `app/workspace/activity/page.tsx` | Typed activity columns extracted. |
| 552 | `app/workspace/data-sources/[id]/query/view/page.tsx` | Unreachable mock query helpers removed; production API flow retained. |
| 540 | `app/workspace/connections/components/ConnectionList.tsx` | Header and utilities extracted; interactive raw markup standardized on shadcn Button. |
| 534 | `app/workspace/connections/data/connectionFields.ts` | Shared field types plus database and SaaS field registries. |
| 523 | `lib/api/hooks/use-destination-schemas.ts` | Compatibility export over CRUD, table, and validation hooks. |
| 519 | `config/connectors.ts` | Shared connector types extracted; registry remains a focused compatibility catalog. |
| 517 | `config/database-registry.ts` | Shared connector types extracted; registry remains below the hard limit. |

The full responsibility/classification audit, duplication findings, estimates, and target structure are in `md-docs/frontend-file-size-audit.md` and were written before source refactoring began.

## Resulting feature architecture

```text
app/workspace/
  dashboard-*.tsx
  activity/activity-columns.tsx
  connections/components + data
  data-pipelines/[id]/_components/
    pipeline-*-tab.tsx
    pipeline-*-columns.tsx
    pipeline-operations-{context,dialogs,query,ui,view}.tsx
    use-pipeline-*.ts
    derive-pipeline-operations.ts
  settings/_components/
  team/{team-member-columns,team-member-types}.ts(x)

components/
  data-sources/        # connection schemas, preview resources/table, form support
  shared/data-table/   # shared typed TanStack foundation
  ui/                  # separated shadcn compound primitives
  workspace/           # shell and organization switcher

lib/
  actions/             # one server-action workflow per module
  api/hooks/           # focused feature query/mutation hooks
  api/services/        # focused typed service layers with compatibility facades
  api/types/           # domain-specific type modules with compatibility exports
```

## Shared UI, table, form, and dialog architecture

The refactor retains the existing shadcn primitives (`Button`, `Input`, `Label`, `Select`, `Switch`, `Dialog`, `AlertDialog`-backed confirmation, `Sheet`, `Tabs`, `Badge`, `Alert`, `ScrollArea`, and semantic `Table`) and application wrappers (`PageHeader`, loading/error/empty states, `ConfirmationModal`, and settings panels). Product-specific wrappers are used only where they add stable behavior.

The shared `DataTable<TData, TValue>` remains responsible for typed columns, row IDs, loading/error/empty states, client/server pagination, filtering, sorting, visibility, selection, row actions, persistence, and responsive horizontal overflow. Feature files now provide only row-specific columns and callbacks. Mobile pipeline card layouts remain alongside the desktop TanStack tables to preserve responsive behavior.

Connection form schema/default metadata and password visibility were separated from rendering. Existing React Hook Form/Zod/shadcn form behavior was retained rather than changing validation or payload mapping. Settings organization and billing forms now live in focused tab components. Large Slack/GitHub sheets and all pipeline operation dialogs are controlled once at feature scope rather than being coupled to table rows.

## Extracted hooks, services, types, and schemas

- Pipeline operations: model, editor state, source operations, paginated queries, GitHub operations, derived selectors, URL tab parser, and typed columns.
- Data pipelines: query, CRUD mutation, execution, validation, run, and sync hooks; configuration service; common/config/core/entity/result/run types.
- Data sources: connection, discovery, query, sync, and monitoring hooks; CRUD, connection, preview, schema, legacy, operations, and normalization services.
- Destination schemas: CRUD, table, and validation hooks.
- Settings: integration controller, callback URL helpers, integration badges, billing and organization controllers.
- Forms/config: connector types, connection schema families, connection field families, transform-example families, connection sheet types/default builders.
- Authentication: shared action result/validation support plus login, signup, forgot, reset, change-password, and invite actions.

## Duplication removed

- Five hand-rendered desktop tables now share the TanStack DataTable state/rendering system.
- Connector, connection-field, transform-example, pipeline-type, hook, and service monoliths now compose focused modules through compatibility exports.
- Pipeline status/pagination helpers and settings integration badge/callback logic have one feature-owned implementation.
- Large page-local JSX for dashboard, settings, pipeline operations, transformation editing, destination editing, onboarding, product tour, and sidebar organization selection is now responsibility-owned.
- Controlled dialogs and drawers are separated from row/page markup; no per-row pipeline operation dialogs were introduced.

## Final file-size report

- Files above 500 lines: **none**
- Largest files: `ConnectionList.tsx` 500; `team/page.tsx` 499; `connection-sheet.tsx` 495; query result view 495; `connectors.ts` 494; settings integration hook 494.
- The hard limit is measured after formatter output, not before formatting.

## Static validation

| Check | Result |
| --- | --- |
| Production-source formatter | Passed — 546 paths checked, no fixes required. |
| Production-source Biome lint/check | Passed — 546 paths checked, no diagnostics. |
| TypeScript `tsc --noEmit` | Passed. |
| Next.js production build | Passed; all 46 static pages generated and route collection completed. |
| Tests | Not created and not run, as required. |

The repository-wide `bun run lint` also scans tests and is blocked by one pre-existing formatting difference in `__tests__/playwright/workflow/pages/pipelines.page.ts` around line 461. That test file was intentionally not modified. Production-source lint passes cleanly.

## Pre-existing warnings and remaining technical debt

- Build warning: `@duckdb/duckdb-wasm/dist/duckdb-node.cjs` uses a dynamic dependency expression through SQLRooms. The build succeeds.
- Tooling warning: `baseline-browser-mapping` data is older than two months. No dependency update was made because dependency upgrades were outside scope.
- Several files are close to the hard limit. `ConnectionList.tsx` at exactly 500 lines is compliant but remains the first candidate for a future action-controller extraction.
- Connector metadata still has multiple product-specific views because their runtime shapes and availability policies differ. The shared types/schema-family extraction reduces coupling without changing contracts; deeper normalization should be a separately tested change.
- This was intentionally static-only work. Functional regression, accessibility interaction, and browser workflow testing remain for the separately authorized testing phase.
