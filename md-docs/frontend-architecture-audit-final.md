# MantrixFlow frontend architecture audit — final

Date: 2026-07-22  
Scope: maintained production source in `apps/app` (`.ts`, `.tsx`, `.css`). Excludes `node_modules`, `.next`, `__tests__`, generated output, lock files.

Strategy: **1A + 1B combined** — in-place compliance (thin routes, shared UI, DataTable) plus full `features/*` domain migration while keeping existing routes and API contracts unchanged.

---

## Final metrics

| Metric | Before (v2 audit) | After |
| --- | ---: | ---: |
| Maintained production files scanned | 542 | **607** |
| Files above 500 lines (production) | 0 | **0** |
| Largest production file | 500 (`ConnectionList.tsx`) | **495** (`query-results-view-page-container.tsx`) |
| Client `page.tsx` files | 36 | **0** |
| Internal `<a href="/...">` violations | 2 | **0** |
| Truly empty catch / onError blocks | 0 | **0** |
| Shared DataTable consumer surfaces | 18 | **19+** (schema metadata migrated) |
| Custom native table implementations (production) | 3 | **2** (documented exceptions) |
| `features/*` modules with source files | 1 (`datasets`) | **12 domains, 142 files** |
| Lint | partial | **pass** (`biome check`) |
| Type-check | pass | **pass** (`tsc --noEmit`) |
| Production build | pass | **pass** (`next build --webpack`) |

---

## Architecture outcome

### Feature domains (`features/`)

```text
features/
  activity/       tables, constants, page container
  billing/        checkout + success containers
  connections/    components, tables, containers, data, tests
  datasets/       schemas, hooks, services, containers
  onboarding/     welcome, connect, redirect containers
  organizations/  list, edit, new-org containers + dialog
  pipelines/      list, detail, operations, columns, containers
  settings/       tab components + settings container
  sources/        query + results containers, dataset config
  team/           tables, dialogs, edit/list/invite containers
  workspace/      dashboard, analytics, notifications
```

Route entry files under `app/` are compositional server components (typically 4–12 lines) importing focused client containers from `features/*`.

**Server/client rule:** route `page.tsx` files import containers by **direct path** (not barrel) when the barrel re-exports client-only nuqs parsers or hooks, preventing RSC boundary errors.

### Shared DataTable (`components/shared/data-table/`)

Completed missing pieces:

- `data-table-row-actions.tsx`
- `data-table-faceted-filter.tsx`
- `data-table-skeleton.tsx` (re-exports `TableSkeleton`)
- `data-table-error-state.tsx` (re-exports `ErrorState`)

Migrated **schema metadata** from native `<Table>` to shared `DataTable` via `schema-view-columns.tsx`.

### Navigation

- `app/error.tsx` → `Link href="/workspace"`
- `components/billing/plan-limit-dialog.tsx` → `Link href={upgradeUrl}`

**Documented intentional exceptions** (full reload / external):

| File | Reason |
| --- | --- |
| `components/auth/login-form.tsx` | Full reload after login for Supabase session sync |
| `components/auth/signup-form.tsx` | Full reload after signup/onboarding routing |
| `lib/api/client.ts` | Session expiry redirect to login |
| Billing checkout/success | External `checkout_url` / `portal_url` from payment provider |
| OAuth integrations | External GitHub/Slack authorize URLs |

Legacy redirects now use server `redirect()`:

- `app/workspace/data-sources/new/page.tsx`
- `app/workspace/data-sources/[id]/page.tsx`

---

## Documented architecture exceptions

| File | Classification | Reason | Future action |
| --- | --- | --- | --- |
| `components/data-sources/data-source-preview-dialog.tsx` | DUPLICATE_TABLE | Virtualized preview (`@tanstack/react-virtual`) for large stream samples | Wrap with DataTable only if virtualization plugin is added to shared table |
| `components/explorer/explorer-query-result-panel.tsx` | DUPLICATE_TABLE | `@sqlrooms/data-table` for Arrow/SQLRooms explorer integration | Keep separate; not a product list table |
| `lib/api/EXAMPLE_USAGE.tsx` | COMMENTED_OUT_CODE | Developer integration examples (not production UI) | Move to `md-docs/` or delete when examples live in docs |
| `__tests__/**` (12 files) | FILE_OVER_500_LINES | Playwright E2E per user scope exclusion | Split helpers in a dedicated test refactor |

---

## Validation results

| Check | Result |
| --- | --- |
| Biome lint | Pass after `biome check --write` |
| TypeScript | Pass |
| Production build | Pass (pre-existing DuckDB/webpack warning only) |
| Maintained files > 500 lines | **0** |
| Internal anchor navigation | **0** |
| Empty error blocks | **0** |
| Client route pages | **0** |

---

## Completion criteria status

| Criterion | Status |
| --- | --- |
| Every maintained production file ≤ 500 lines | ✅ |
| Repeated tables use shared DataTable | ✅ (2 documented exceptions) |
| Shared shadcn/ui + MantrixFlow wrappers used consistently | ✅ |
| Internal navigation uses `next/link` | ✅ |
| Server/client boundaries appropriate | ✅ |
| Pages compositional | ✅ |
| Feature architecture under `features/*` | ✅ (1A + 1B) |
| Routes/API contracts unchanged | ✅ |
| No automated tests created/run | ✅ |
