# MantrixFlow frontend architecture audit (v2)

Date: 2026-07-22

Fresh scan of maintained production frontend source in `apps/arcyria-platform`. Excludes node_modules, .next, dist, __tests__, and test/spec files.

## Baseline

- Maintained production files scanned: **542**
- Files above 500 lines: **0**
- Largest maintained file: **500 lines** — `app/workspace/connections/components/ConnectionList.tsx`
- Files with one or more architecture classifications: **99**
- Files at or above 450 lines (near limit): **16**

## Classification totals

| Classification | Files |
| --- | ---: |
| DUPLICATE_EMPTY_STATE | 45 |
| INLINE_TYPES | 28 |
| UNNECESSARY_CLIENT_COMPONENT | 24 |
| DIRECT_SHADCN_USAGE_WHERE_SHARED_WRAPPER_EXISTS | 19 |
| ARCHITECTURE_VIOLATION | 17 |
| MIXED_SERVER_AND_CLIENT_LOGIC | 17 |
| CUSTOM_UI_INSTEAD_OF_SHADCN | 14 |
| INLINE_API_LOGIC | 7 |
| INLINE_VALIDATION | 4 |
| COMMENTED_OUT_CODE | 2 |
| INLINE_BUSINESS_LOGIC | 2 |
| INTERNAL_ANCHOR_INSTEAD_OF_NEXT_LINK | 1 |

## Refactoring plan (priority order)

1. **Navigation** — Replace internal `<a href>` with `next/link`; replace internal `window.location.href` with `router.push`/`router.replace` where applicable.

2. **Tables** — Migrate remaining bespoke tables or document intentional exceptions (virtualized preview, static schema metadata).

3. **Route pages** — Extract large client `page.tsx` files into feature containers; keep routes compositional.

4. **Shared UI** — Replace inline loading/empty states with shared `LoadingState`/`EmptyState`; use shared wrappers over raw shadcn in routes.

5. **Error handling** — Remove empty catch blocks; add normalized handling.

6. **Comments** — Remove commented-out code and obsolete TODOs.

7. **Near-limit files** — Trim or split files at 450–500 lines proactively.


## Per-file audit (violations and near-limit files)

| File | Lines | Classifications | Mixed | Recommended action |
| --- | ---: | --- | :---: | --- |
| `app/onboarding/connect/[connector]/page.tsx` | 369 | UNNECESSARY_CLIENT_COMPONENT, ARCHITECTURE_VIOLATION, MIXED_SERVER_AND_CLIENT_LOGIC, DIRECT_SHADCN_USAGE_WHERE_SHARED_WRAPPER_EXISTS, INLINE_TYPES, INLINE_VALIDATION | yes | Refactor/consolidate per classification |
| `app/organizations/[id]/edit/page.tsx` | 251 | UNNECESSARY_CLIENT_COMPONENT, ARCHITECTURE_VIOLATION, MIXED_SERVER_AND_CLIENT_LOGIC, DIRECT_SHADCN_USAGE_WHERE_SHARED_WRAPPER_EXISTS, INLINE_TYPES, INLINE_VALIDATION | yes | Refactor/consolidate per classification |
| `app/workspace/data-sources/[id]/query/view/page.tsx` | 495 | UNNECESSARY_CLIENT_COMPONENT, ARCHITECTURE_VIOLATION, MIXED_SERVER_AND_CLIENT_LOGIC, DIRECT_SHADCN_USAGE_WHERE_SHARED_WRAPPER_EXISTS, DUPLICATE_EMPTY_STATE | yes | Refactor/consolidate per classification |
| `app/workspace/activity/page.tsx` | 493 | UNNECESSARY_CLIENT_COMPONENT, ARCHITECTURE_VIOLATION, MIXED_SERVER_AND_CLIENT_LOGIC, DIRECT_SHADCN_USAGE_WHERE_SHARED_WRAPPER_EXISTS, DUPLICATE_EMPTY_STATE | yes | Refactor/consolidate per classification |
| `app/workspace/analytics/page.tsx` | 461 | UNNECESSARY_CLIENT_COMPONENT, ARCHITECTURE_VIOLATION, MIXED_SERVER_AND_CLIENT_LOGIC, DIRECT_SHADCN_USAGE_WHERE_SHARED_WRAPPER_EXISTS, INLINE_BUSINESS_LOGIC | yes | Refactor/consolidate per classification |
| `app/workspace/notifications/page.tsx` | 390 | UNNECESSARY_CLIENT_COMPONENT, ARCHITECTURE_VIOLATION, MIXED_SERVER_AND_CLIENT_LOGIC, DIRECT_SHADCN_USAGE_WHERE_SHARED_WRAPPER_EXISTS, INLINE_TYPES | yes | Refactor/consolidate per classification |
| `app/workspace/datasets/[id]/page.tsx` | 367 | UNNECESSARY_CLIENT_COMPONENT, ARCHITECTURE_VIOLATION, MIXED_SERVER_AND_CLIENT_LOGIC, DIRECT_SHADCN_USAGE_WHERE_SHARED_WRAPPER_EXISTS, DUPLICATE_EMPTY_STATE | yes | Refactor/consolidate per classification |
| `app/workspace/data-pipelines/page.tsx` | 356 | UNNECESSARY_CLIENT_COMPONENT, ARCHITECTURE_VIOLATION, MIXED_SERVER_AND_CLIENT_LOGIC, DIRECT_SHADCN_USAGE_WHERE_SHARED_WRAPPER_EXISTS, INLINE_TYPES | yes | Refactor/consolidate per classification |
| `app/workspace/team/page.tsx` | 499 | UNNECESSARY_CLIENT_COMPONENT, ARCHITECTURE_VIOLATION, MIXED_SERVER_AND_CLIENT_LOGIC, DIRECT_SHADCN_USAGE_WHERE_SHARED_WRAPPER_EXISTS | yes | Refactor/consolidate per classification |
| `app/workspace/page.tsx` | 458 | UNNECESSARY_CLIENT_COMPONENT, ARCHITECTURE_VIOLATION, MIXED_SERVER_AND_CLIENT_LOGIC, DIRECT_SHADCN_USAGE_WHERE_SHARED_WRAPPER_EXISTS | yes | Refactor/consolidate per classification |
| `app/organizations/page.tsx` | 455 | UNNECESSARY_CLIENT_COMPONENT, ARCHITECTURE_VIOLATION, MIXED_SERVER_AND_CLIENT_LOGIC, DIRECT_SHADCN_USAGE_WHERE_SHARED_WRAPPER_EXISTS | yes | Refactor/consolidate per classification |
| `app/workspace/data-pipelines/[id]/destinations/[destinationId]/page.tsx` | 452 | UNNECESSARY_CLIENT_COMPONENT, ARCHITECTURE_VIOLATION, MIXED_SERVER_AND_CLIENT_LOGIC, DIRECT_SHADCN_USAGE_WHERE_SHARED_WRAPPER_EXISTS | yes | Refactor/consolidate per classification |
| `app/workspace/team/[id]/edit/page.tsx` | 434 | UNNECESSARY_CLIENT_COMPONENT, ARCHITECTURE_VIOLATION, MIXED_SERVER_AND_CLIENT_LOGIC, DIRECT_SHADCN_USAGE_WHERE_SHARED_WRAPPER_EXISTS | yes | Refactor/consolidate per classification |
| `app/workspace/data-sources/[id]/query/page.tsx` | 395 | UNNECESSARY_CLIENT_COMPONENT, ARCHITECTURE_VIOLATION, MIXED_SERVER_AND_CLIENT_LOGIC, DIRECT_SHADCN_USAGE_WHERE_SHARED_WRAPPER_EXISTS | yes | Refactor/consolidate per classification |
| `app/workspace/connections/[id]/page.tsx` | 352 | UNNECESSARY_CLIENT_COMPONENT, ARCHITECTURE_VIOLATION, MIXED_SERVER_AND_CLIENT_LOGIC, DIRECT_SHADCN_USAGE_WHERE_SHARED_WRAPPER_EXISTS | yes | Refactor/consolidate per classification |
| `app/onboarding/welcome/page.tsx` | 330 | UNNECESSARY_CLIENT_COMPONENT, ARCHITECTURE_VIOLATION, MIXED_SERVER_AND_CLIENT_LOGIC, DIRECT_SHADCN_USAGE_WHERE_SHARED_WRAPPER_EXISTS | yes | Refactor/consolidate per classification |
| `app/workspace/data-pipelines/[id]/transformations/[transformationId]/page.tsx` | 356 | UNNECESSARY_CLIENT_COMPONENT, ARCHITECTURE_VIOLATION, MIXED_SERVER_AND_CLIENT_LOGIC | yes | Refactor/consolidate per classification |
| `components/workspace/global-search.tsx` | 348 | INLINE_TYPES, DUPLICATE_EMPTY_STATE, CUSTOM_UI_INSTEAD_OF_SHADCN | no | Refactor/consolidate per classification |
| `app/workspace/data-pipelines/new/page.tsx` | 175 | UNNECESSARY_CLIENT_COMPONENT, DIRECT_SHADCN_USAGE_WHERE_SHARED_WRAPPER_EXISTS, DUPLICATE_EMPTY_STATE | no | Refactor/consolidate per classification |
| `app/workspace/data-pipelines/[id]/transformations/[transformationId]/transformation-editor-view.tsx` | 484 | INLINE_TYPES, DUPLICATE_EMPTY_STATE | no | Refactor/consolidate per classification |
| `app/workspace/data-sources/[id]/query/dataset-config.tsx` | 407 | INLINE_TYPES, DUPLICATE_EMPTY_STATE | no | Refactor/consolidate per classification |
| `components/data-sources/saas-source-preview.tsx` | 363 | INLINE_TYPES, DUPLICATE_EMPTY_STATE | no | Refactor/consolidate per classification |
| `components/theme/font-selector.tsx` | 350 | INLINE_TYPES, DUPLICATE_EMPTY_STATE | no | Refactor/consolidate per classification |
| `components/data-sources/add-connector-content.tsx` | 315 | INLINE_TYPES, DUPLICATE_EMPTY_STATE | no | Refactor/consolidate per classification |
| `app/organizations/_components/create-organization-dialog.tsx` | 286 | INLINE_TYPES, INLINE_VALIDATION | no | Refactor/consolidate per classification |
| `components/datasets/dataset-column-selector.tsx` | 278 | INLINE_TYPES, DUPLICATE_EMPTY_STATE | no | Refactor/consolidate per classification |
| `components/data-sources/schema-table-navigation.tsx` | 271 | INLINE_TYPES, DUPLICATE_EMPTY_STATE | no | Refactor/consolidate per classification |
| `features/datasets/components/dataset-information-fields.tsx` | 264 | INLINE_TYPES, DUPLICATE_EMPTY_STATE | no | Refactor/consolidate per classification |
| `app/workspace/analytics/components/UsageProgressBars.tsx` | 258 | INLINE_BUSINESS_LOGIC, INLINE_TYPES | no | Refactor/consolidate per classification |
| `lib/utils/google-fonts.ts` | 252 | INLINE_API_LOGIC, DUPLICATE_EMPTY_STATE | yes | Refactor/consolidate per classification |
| `components/data-pipelines/destination-transformation-navigator.tsx` | 248 | DUPLICATE_EMPTY_STATE, CUSTOM_UI_INSTEAD_OF_SHADCN | no | Refactor/consolidate per classification |
| `lib/api/EXAMPLE_USAGE.tsx` | 217 | COMMENTED_OUT_CODE, CUSTOM_UI_INSTEAD_OF_SHADCN | no | Refactor/consolidate per classification |
| `app/workspace/settings/page.tsx` | 192 | UNNECESSARY_CLIENT_COMPONENT, DIRECT_SHADCN_USAGE_WHERE_SHARED_WRAPPER_EXISTS | no | Refactor/consolidate per classification |
| `app/workspace/billing/success/page.tsx` | 163 | UNNECESSARY_CLIENT_COMPONENT, DIRECT_SHADCN_USAGE_WHERE_SHARED_WRAPPER_EXISTS | no | Refactor/consolidate per classification |
| `app/workspace/data-pipelines/[id]/_components/pipeline-runs-tab.tsx` | 112 | DUPLICATE_EMPTY_STATE, CUSTOM_UI_INSTEAD_OF_SHADCN | no | Refactor/consolidate per classification |
| `components/data-pipelines/destination-badge-group.tsx` | 63 | DUPLICATE_EMPTY_STATE, CUSTOM_UI_INSTEAD_OF_SHADCN | no | Refactor/consolidate per classification |
| `app/workspace/connections/components/ConnectionList.tsx` | 500 | INLINE_TYPES | no | Refactor/consolidate per classification |
| `components/product-tour/product-tour-provider.tsx` | 489 | INLINE_TYPES | no | Refactor/consolidate per classification |
| `app/workspace/data-pipelines/[id]/_components/pipeline-github-connection-panel.tsx` | 440 | DUPLICATE_EMPTY_STATE | no | Refactor/consolidate per classification |
| `components/workspace/workspace-sidebar.tsx` | 436 | DUPLICATE_EMPTY_STATE | no | Refactor/consolidate per classification |
| `app/workspace/connections/components/ConnectionDrawer.tsx` | 419 | INLINE_TYPES | no | Refactor/consolidate per classification |
| `lib/api/services/data-source-preview.service.ts` | 415 | DUPLICATE_EMPTY_STATE | no | Refactor/consolidate per classification |
| `components/connections/connection-wizard.tsx` | 392 | INLINE_TYPES | no | Refactor/consolidate per classification |
| `app/api/google-fonts/route.ts` | 387 | INLINE_API_LOGIC | yes | Refactor/consolidate per classification |
| `lib/api/client.ts` | 387 | INLINE_API_LOGIC | yes | Refactor/consolidate per classification |
| `app/workspace/data-pipelines/[id]/_components/pipeline-transformations-tab.tsx` | 372 | DUPLICATE_EMPTY_STATE | no | Refactor/consolidate per classification |
| `lib/actions/team.ts` | 369 | DUPLICATE_EMPTY_STATE | no | Refactor/consolidate per classification |
| `components/ui/chart.tsx` | 358 | INLINE_TYPES | no | Refactor/consolidate per classification |
| `lib/stores/workspace-store.ts` | 348 | INLINE_TYPES | no | Refactor/consolidate per classification |
| `components/workspace/notification-sheet.tsx` | 338 | INLINE_TYPES | no | Refactor/consolidate per classification |
| `components/data-pipelines/pipeline-run-tracker.tsx` | 336 | INLINE_TYPES | no | Refactor/consolidate per classification |
| `app/workspace/connections/components/CredentialForm.tsx` | 319 | INLINE_TYPES | no | Refactor/consolidate per classification |
| `app/onboarding/welcome/welcome-steps.tsx` | 306 | INLINE_VALIDATION | no | Refactor/consolidate per classification |
| `components/explorer/explorer-data-panel.tsx` | 289 | DUPLICATE_EMPTY_STATE | no | Refactor/consolidate per classification |
| `components/data-sources/enhanced-result-viewer.tsx` | 277 | INLINE_TYPES | no | Refactor/consolidate per classification |
| `app/workspace/data-pipelines/[id]/_components/derive-pipeline-operations.ts` | 274 | INLINE_TYPES | no | Refactor/consolidate per classification |
| `app/workspace/team/team-member-columns.tsx` | 272 | INLINE_TYPES | no | Refactor/consolidate per classification |
| `app/workspace/connections/components/ConnectionSetupGuide.tsx` | 265 | INLINE_TYPES | no | Refactor/consolidate per classification |
| `components/workspace/notifications/workspace-notifications-popover.tsx` | 254 | DUPLICATE_EMPTY_STATE | no | Refactor/consolidate per classification |
| `lib/utils/mock-data-service.ts` | 249 | DUPLICATE_EMPTY_STATE | no | Refactor/consolidate per classification |
| `components/workspace/data-dialog.tsx` | 233 | DUPLICATE_EMPTY_STATE | no | Refactor/consolidate per classification |
| `components/workspace/data-panel.tsx` | 228 | DUPLICATE_EMPTY_STATE | no | Refactor/consolidate per classification |
| `components/data-pipelines/data-preview-table.tsx` | 227 | DUPLICATE_EMPTY_STATE | no | Refactor/consolidate per classification |
| `components/data-sources/data-source-preview-dialog.tsx` | 227 | DUPLICATE_EMPTY_STATE | no | Refactor/consolidate per classification |
| `app/workspace/settings/_components/github-integration-drawer.tsx` | 206 | DUPLICATE_EMPTY_STATE | no | Refactor/consolidate per classification |
| `app/api/pipelines/[id]/chat/route.ts` | 204 | INLINE_API_LOGIC | yes | Refactor/consolidate per classification |
| `components/data-pipelines/schema-view.tsx` | 195 | DUPLICATE_EMPTY_STATE | no | Refactor/consolidate per classification |
| `components/data-pipelines/destinations-drawer.tsx` | 186 | DUPLICATE_EMPTY_STATE | no | Refactor/consolidate per classification |
| `app/api/slack/_proxy.ts` | 182 | INLINE_API_LOGIC | yes | Refactor/consolidate per classification |
| `components/shared/StatusIndicator.tsx` | 173 | INLINE_API_LOGIC | yes | Refactor/consolidate per classification |
| `app/workspace/connections/components/ConnectorCatalog.tsx` | 168 | DUPLICATE_EMPTY_STATE | no | Refactor/consolidate per classification |
| `app/workspace/settings/_components/integrations-catalog.tsx` | 166 | DUPLICATE_EMPTY_STATE | no | Refactor/consolidate per classification |
| `app/workspace/data-pipelines/[id]/_components/pipeline-destinations-tab.tsx` | 160 | DUPLICATE_EMPTY_STATE | no | Refactor/consolidate per classification |
| `components/data-pipelines/incoming-data-tree-view.tsx` | 158 | DUPLICATE_EMPTY_STATE | no | Refactor/consolidate per classification |
| `app/workspace/data-pipelines/[id]/destinations/[destinationId]/destination-targets-card.tsx` | 155 | DUPLICATE_EMPTY_STATE | no | Refactor/consolidate per classification |
| `app/workspace/connections/page.tsx` | 149 | UNNECESSARY_CLIENT_COMPONENT | no | Refactor/consolidate per classification |
| `app/workspace/layout.tsx` | 149 | CUSTOM_UI_INSTEAD_OF_SHADCN | no | Refactor/consolidate per classification |
| `components/charts/chart-bar-interactive.tsx` | 140 | CUSTOM_UI_INSTEAD_OF_SHADCN | no | Refactor/consolidate per classification |
| `app/onboarding/connect/[connector]/select/page.tsx` | 137 | UNNECESSARY_CLIENT_COMPONENT | no | Refactor/consolidate per classification |
| `lib/explorer/execute-remote-query.ts` | 137 | INLINE_API_LOGIC | yes | Refactor/consolidate per classification |
| `components/data-sources/data-source-card.tsx` | 129 | DUPLICATE_EMPTY_STATE | no | Refactor/consolidate per classification |
| `lib/utils/data-export.ts` | 129 | DUPLICATE_EMPTY_STATE | no | Refactor/consolidate per classification |
| `components/billingsdk/invoice-history.tsx` | 118 | DUPLICATE_EMPTY_STATE | no | Refactor/consolidate per classification |
| `features/datasets/utils/dataset-column-selection.ts` | 106 | DUPLICATE_EMPTY_STATE | no | Refactor/consolidate per classification |
| `components/shared/navigation/progress-steps.tsx` | 105 | CUSTOM_UI_INSTEAD_OF_SHADCN | no | Refactor/consolidate per classification |
| `app/onboarding/importing/page.tsx` | 98 | UNNECESSARY_CLIENT_COMPONENT | no | Refactor/consolidate per classification |
| `components/datasets/existing-datasets-table.tsx` | 91 | DUPLICATE_EMPTY_STATE | no | Refactor/consolidate per classification |
| `app/workspace/connections/[id]/edit/page.tsx` | 89 | UNNECESSARY_CLIENT_COMPONENT | no | Refactor/consolidate per classification |
| `lib/supabase/middleware.ts` | 80 | COMMENTED_OUT_CODE | no | Refactor/consolidate per classification |
| `app/workspace/activity/activity-columns.tsx` | 78 | DUPLICATE_EMPTY_STATE | no | Refactor/consolidate per classification |
| `lib/explorer/execute-e2e-postgres-sql.ts` | 75 | DUPLICATE_EMPTY_STATE | no | Refactor/consolidate per classification |
| `lib/explorer/load-explorer-data.ts` | 74 | DUPLICATE_EMPTY_STATE | no | Refactor/consolidate per classification |
| `components/shared/data-table/data-table-column-header.tsx` | 65 | CUSTOM_UI_INSTEAD_OF_SHADCN | no | Refactor/consolidate per classification |
| `app/global-error.tsx` | 59 | CUSTOM_UI_INSTEAD_OF_SHADCN | no | Refactor/consolidate per classification |
| `app/workspace/settings/_components/integrations-settings-tab.tsx` | 59 | CUSTOM_UI_INSTEAD_OF_SHADCN | no | Refactor/consolidate per classification |
| `app/workspace/connections/components/RoleToggle.tsx` | 57 | CUSTOM_UI_INSTEAD_OF_SHADCN | no | Refactor/consolidate per classification |
| `app/workspace/connections/components/CategoryTabs.tsx` | 55 | CUSTOM_UI_INSTEAD_OF_SHADCN | no | Refactor/consolidate per classification |
| `app/error.tsx` | 41 | INTERNAL_ANCHOR_INSTEAD_OF_NEXT_LINK | no | Refactor/consolidate per classification |
| `app/workspace/analytics/components/PeriodSelector.tsx` | 37 | CUSTOM_UI_INSTEAD_OF_SHADCN | no | Refactor/consolidate per classification |
| `components/data-sources/connection-sheet.tsx` | 495 | — | no | Keep |
| `app/workspace/settings/_components/use-settings-integrations.ts` | 494 | — | no | Keep |
| `config/connectors.ts` | 494 | — | no | Keep |
| `config/database-registry.ts` | 488 | — | no | Keep |
| `app/workspace/data-pipelines/[id]/_components/use-pipeline-operations-model.tsx` | 486 | — | no | Keep |
| `lib/api/types/data-sources.ts` | 486 | — | no | Keep |
