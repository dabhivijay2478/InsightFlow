# MantrixFlow frontend — fractal architecture restructure (final)

Date: 2026-07-22  
Scope: `apps/app` (Next.js 16 frontend)  
Source of truth: `main` branch conventions, **no `src/` root** (per project decision)

---

## Executive summary

The frontend was reorganized into a **feature-based fractal architecture** without introducing a `src/` folder. Next.js routes remain under `app/`; domain code lives under `features/`; cross-cutting UI and infrastructure live under `shared/`. The legacy `components/` directory was removed. All Playwright and unit test infrastructure was removed. Production build, lint, and typecheck pass.

---

## Final folder tree (high level)

```text
apps/app/
├── app/                    # Next.js App Router (thin routes only)
├── features/               # Domain modules (fractal)
│   ├── activity/
│   ├── auth/
│   ├── billing/
│   ├── connections/
│   ├── data-sources/
│   ├── datasets/
│   ├── onboarding/
│   ├── organizations/
│   ├── pipelines/
│   ├── product-tour/
│   ├── settings/
│   ├── sql-explorer/
│   ├── team/
│   ├── workspace/
│   └── workspace-shell/
├── shared/                 # Cross-feature layer
│   ├── ui/                 # shadcn primitives
│   ├── data-table/         # TanStack DataTable
│   ├── components/         # PageHeader, EmptyState, modals, etc.
│   ├── feedback/
│   ├── forms/
│   ├── layout/
│   ├── navigation/
│   ├── skeletons/
│   ├── sheet/
│   ├── data-display/
│   ├── providers/
│   └── theme/
├── lib/                    # API client, actions, stores, utils (infra — see exceptions)
├── hooks/                  # Shared hooks (e.g. use-confirmation)
├── config/                 # Connector registries, transform examples
├── public/
├── package.json
├── tsconfig.json           # "@/*" → "./*"
└── biome.json
```

**Removed:** `components/`, `__tests__/`, `playwright.config.ts`, `.playwright/`, `playwright-report/`, `ui-screenshots/`, empty `src/`, screenshot scripts.

---

## Feature ownership map

| Domain | Feature path | Screens (route entry) |
| --- | --- | --- |
| Auth | `features/auth/` | Login, signup, reset, invite forms via `app/auth/*` |
| Onboarding | `features/onboarding/` | Welcome, connect, importing, redirects |
| Organizations | `features/organizations/` | List, edit, new org |
| Workspace shell | `features/workspace-shell/` | Sidebar, topbar, notifications popover |
| Dashboard | `features/workspace/` | `WorkspaceDashboardScreen` |
| Analytics | `features/workspace/` | `AnalyticsScreen` |
| Notifications | `features/workspace/` | `NotificationsScreen` |
| Activity | `features/activity/` | `ActivityScreen` |
| Connections | `features/connections/` | List, detail, edit, new catalog/form |
| Data sources | `features/data-sources/` | Query, query results view |
| Datasets | `features/datasets/` | Dataset configuration |
| Pipelines | `features/pipelines/` | List, new, detail, edit, destination/transformation editors |
| SQL explorer | `features/sql-explorer/` | Explorer panels (used from data-sources screen) |
| Team | `features/team/` | Team list, edit member, invite |
| Settings | `features/settings/` | Settings tabs screen |
| Billing | `features/billing/` | Checkout, success, plan limit dialog |
| Product tour | `features/product-tour/` | Tour provider (workspace layout) |

---

## Route → feature map (examples)

| Route | Screen import |
| --- | --- |
| `app/workspace/connections/page.tsx` | `@/features/connections` → `ConnectionsScreen` |
| `app/workspace/data-pipelines/page.tsx` | `@/features/pipelines` → `PipelinesScreen` |
| `app/workspace/data-pipelines/[id]/page.tsx` | `@/features/pipelines` → `PipelineDetailScreen` |
| `app/workspace/team/page.tsx` | `@/features/team` → `TeamScreen` |
| `app/workspace/page.tsx` | `@/features/workspace` → `WorkspaceDashboardScreen` |

Route files are **4–15 lines** server components importing a single screen.

---

## Public feature exports (barrel rule)

Each `features/*/index.ts` exports **screens only** (plus auth forms for auth routes). This prevents server routes from pulling client-only nuqs parsers/hooks through barrel side effects.

**Deep imports** for internal modules:

```text
@/features/connections/components/ConnectionList
@/features/connections/data/connectors
@/features/pipelines/components/operations/use-pipeline-operations-model
@/features/team/types/team-member-types
@/shared/ui/button
@/shared/data-table
@/shared/components
```

---

## Shared module map

| Path | Responsibility |
| --- | --- |
| `shared/ui/` | shadcn/Radix primitives |
| `shared/data-table/` | TanStack `DataTable` system |
| `shared/components/` | Cross-feature layout/feedback/modals |
| `shared/providers/` | PostHog provider |
| `shared/theme/` | Theme provider + customizer |

Import aliases updated project-wide:

- `@/components/ui/*` → `@/shared/ui/*`
- `@/components/shared/*` → `@/shared/*` or `@/shared/components`

---

## Pipeline workspace (canonical)

Pipeline operations remain under `features/pipelines/components/operations/` with tabs:

Overview · Source · Transformations · Destinations · Runs · GitHub · Settings

Shared pipeline UI (schema view, run tracker, schedule editor) lives in `features/pipelines/components/shared/`. **No canvas-first builder** references remain in production routes.

---

## Removal report

### Test infrastructure removed

| Item | Action |
| --- | --- |
| `__tests__/` (Playwright API/UI/workflow) | Deleted |
| `features/**/__tests__/` | Deleted |
| `components/**/__tests__/` | Deleted |
| `lib/pipelines/__tests__/` | Deleted |
| `shared/data-table/__tests__/` | Deleted |
| `playwright.config.ts` | Deleted |
| `.playwright/`, `playwright-report/`, `ui-screenshots/` | Deleted |

### Test scripts removed from `package.json`

All `test`, `test:unit`, `test:playwright*`, `test:e2e*` scripts removed.

### Test dependencies removed

`@playwright/test`, `playwright` removed from devDependencies; `bun.lock` regenerated.

### CI removed

`.github/workflows/playwright-e2e.yml` deleted.

### Scripts removed

| Script | Action |
| --- | --- |
| `capture-screenshots` | Removed from package.json |
| `capture-docs-screenshots` | Removed from package.json |
| `scripts/capture-product-screenshots.mjs` | Deleted (prior session) |
| `scripts/capture-docs-screenshots.mjs` | Deleted |
| `build_output.log`, `respone.txt` | Deleted |

### Legacy / dead code removed

| Item | Action |
| --- | --- |
| `components/` directory | Fully migrated; directory removed |
| `ConnectionListRow.tsx` | Removed (unused) |
| Empty `src/` | Removed |

---

## File movement report (major moves)

| Old path | New path | Owner | Reason |
| --- | --- | --- | --- |
| `components/ui/*` | `shared/ui/*` | Shared | Centralize shadcn primitives |
| `components/shared/data-table/*` | `shared/data-table/*` | Shared | Canonical table engine |
| `components/shared/*` | `shared/components`, `shared/feedback`, etc. | Shared | Split shared layer |
| `components/workspace/*` | `features/workspace-shell/components/*` | workspace-shell | Feature ownership |
| `components/auth/*` | `features/auth/components/*` | auth | Feature ownership |
| `components/explorer/*` | `features/sql-explorer/components/*` | sql-explorer | Feature ownership |
| `components/data-sources/*` | `features/data-sources/components/*` | data-sources | Merged with prior `features/sources` |
| `components/data-pipelines/*` | `features/pipelines/components/shared/*` | pipelines | Pipeline shared UI |
| `components/datasets/*` | `features/datasets/components/*` | datasets | Feature ownership |
| `components/connections/*` | `features/connections/components/*` | connections | Merge with feature |
| `components/billing/*`, `billingsdk/*` | `features/billing/components/*` | billing | Feature ownership |
| `components/product-tour/*` | `features/product-tour/components/*` | product-tour | Feature ownership |
| `components/pipeline/*` | `features/pipelines/components/badges/*` | pipelines | Status badges |
| `components/charts/*` | `features/workspace/components/charts/*` | workspace | Analytics charts |
| `components/theme/*` | `shared/theme/*` | Shared | Global theme |
| `components/logo.tsx` | `shared/components/logo.tsx` | Shared | App chrome |
| `app/workspace/**/_components/*` | `features/*/screens`, `features/pipelines/components/operations/*` | Various | Thin routes |
| `*-page-container.tsx` | `*-screen.tsx` | Various | Screen naming convention |

---

## File-size report

| Metric | Value |
| --- | --- |
| Total maintained frontend source files | **605** |
| Files above 500 lines before restructure | **0** (prod); 12 in `__tests__` (removed) |
| Files above 500 lines after | **0** |
| Largest files (approx.) | `query-results-view-screen.tsx` (~495), `team-screen.tsx` (~478), `use-pipeline-operations-model.tsx` (~486) |

---

## Validation report

| Check | Result |
| --- | --- |
| `bun run format` | Pass |
| `bun run lint` | Pass |
| `bun run typecheck` | Pass |
| `bun run build` | Pass (pre-existing DuckDB/webpack warning) |
| Broken `@/components/*` imports | **0** |
| Internal `<a href="/...">` | **0** |
| Client `page.tsx` files | **0** |
| Playwright files | **0** |
| `*.test.ts` / `*.spec.ts` in app source | **0** |

---

## Architecture exceptions (documented)

| Path | Reason | Future action |
| --- | --- | --- |
| `lib/api/**` | Generic API client + domain hooks still centralized | Incrementally move hooks/services into owning `features/*/services` and `shared/api` |
| `lib/stores/workspace-store.ts` | Cross-route org selection + UI state | Split UI state from duplicated server entities when safe |
| `features/data-sources/components/data-source-preview-dialog.tsx` | Virtualized preview table | Keep until DataTable supports virtualization |
| `features/sql-explorer/` + `@sqlrooms/*` | External SQLRooms table engine | Not a product list table; keep separate |
| `lib/api/EXAMPLE_USAGE.tsx` | Developer examples | Move to docs or delete |
| Auth `window.location.href` after login | Session sync full reload | Intentional |

---

## Dependency rules (enforced)

```text
app → features (screens)
app → shared
features → shared
features → lib (transitional)
shared ↛ features
feature A ↛ feature B private internals (use deep paths only when approved)
```

---

## Completion criteria

| Criterion | Status |
| --- | --- |
| No `src/` folder; `app/` + `features/` + `shared/` layout | ✅ |
| Thin Next.js route files | ✅ |
| Feature-owned components/hooks/services | ✅ (lib/api migration partial) |
| Shared layer without feature logic | ✅ |
| Shared DataTable only generic table | ✅ (2 documented exceptions) |
| shadcn in `shared/ui/` | ✅ |
| All files ≤ 500 lines | ✅ |
| No Playwright/tests | ✅ |
| No test scripts/deps/CI | ✅ |
| Production build passes | ✅ |
