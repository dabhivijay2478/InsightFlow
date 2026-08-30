# MANTrixFlow — Analytics Page Implementation

This is the implementation spec for the workspace Analytics page and its
organization-scoped backend APIs.

It is intentionally **repo-aligned**, not prompt-literal. Use this document
when implementing the real feature in:

- `apps/app`
- `apps/server/main-server`

This spec is decision-complete for the v1 implementation. If a detail in the
original prompt conflicts with the current repo shape, follow this document.

---

## 1. Scope

The Analytics page is a workspace page at `/workspace/analytics`. It shows ELT
pipeline metrics for the current organization:

- rows delivered
- successful, partial, and failed run counts
- run duration trends
- per-pipeline performance
- recent failed runs
- usage against plan limits

It is **not** a generic admin dashboard. All metrics are pipeline-operations
metrics sourced from the existing Go API database tables.

v1 scope:

- shared period selector: `7d | 30d | 90d`
- org-scoped read-only analytics APIs
- CSV export for all runs in the selected period
- workspace page UI built from existing shadcn and chart primitives
- source/destination map endpoint implemented and typed, but not rendered on
  the initial page

Explicitly out of scope for v1:

- custom date ranges
- live or realtime streaming updates
- new role-gating beyond the repo's current org-member API access model
- schema migrations for `pipeline_runs.organization_id`
- branch-color-aware charts

---

## 2. Prompt-To-Repo Mapping

Use the following mapping whenever the original prompt uses generic names.

| Prompt concept | Implement in this repo |
| --- | --- |
| Frontend route `/analytics` | `/workspace/analytics` in `apps/app/app/workspace/analytics/page.tsx` |
| API prefix `/api/v1/analytics/*` | `/api/v1/organizations/:organizationId/analytics/*` |
| Legacy API parity | `/api/organizations/:organizationId/analytics/*` via the existing legacy org router |
| Generic Go `internal/router/router.go` | `apps/server/main-server/internal/server/routes.go` |
| Generic Go handlers package | `apps/server/main-server/internal/server/analytics_http.go` |
| Generic Go service package | `apps/server/main-server/internal/services/analytics/*` |
| `etl_pipeline_runs` | `pipeline_runs` |
| `etl_pipelines` | `pipelines` |
| `org_id` | `organization_id` |
| `phase_1_status` / `phase_2_status` / `phase_3_status` columns | derive from `pipeline_runs.run_metadata` |
| Generic `lib/api/hooks/useAnalytics.ts` | `apps/app/lib/api/hooks/use-analytics.ts` |
| Generic `lib/api/services/*` | `apps/app/lib/api/services/*` using `orgPath(organizationId)` |

Important repo-specific notes:

- The Go API returns JSON via `response.Success(...)`, so all JSON endpoints
  return the standard `{ meta, data }` envelope. Frontend hooks should type the
  `data` payload because `ApiClient` already unwraps it.
- `pipeline_runs.organization_id` already exists and is already indexed. Do not
  add the prompt's `org_id` migration.
- Partial delivery is not a top-level `pipeline_runs.status`. It is represented
  by callback metadata such as `run_metadata.raw_status = "partial_success"` and
  `run_metadata.delivery_status = "partial"`.

---

## 3. Page Contract

The page keeps the prompt's layout, but it is rendered under the workspace app:

```text
/workspace/analytics
│
├── Header: "Analytics" + period selector + Export CSV
├── Row 1: KPI cards (4)
├── Row 2: Rows over time + run status donut
├── Row 3: Pipeline run frequency + run duration trend
├── Row 4: Top pipelines + recent failed runs
└── Row 5: Usage vs plan
```

### UI rules

- Use existing `PageHeader`, shadcn `Card`, `Table`, `Progress`, `Skeleton`,
  and the existing `ChartContainer` / `ChartTooltip` wrappers from
  `apps/app/components/ui/chart.tsx`.
- Keep the page dark-themed and consistent with the rest of the workspace.
- Use these colors:
  - primary: `blue-500`
  - success: `green-500`
  - failed: `red-500`
  - warning / partial: `amber-500`
  - grid lines: `#27272a`
  - axis text: `#52525b`
- Every KPI, chart, and table must have a loading skeleton that matches the
  loaded layout dimensions.
- If the organization has no pipelines, render a clear empty state on the page
  instead of blank charts.

### Navigation

Re-enable the commented Analytics sidebar item in
`apps/app/components/workspace/workspace-sidebar.tsx` and place it literally
between the existing `Connections` and `Data Pipelines` items. Use `BarChart2`
and route it to `/workspace/analytics`.

---

## 4. Backend Implementation

### 4.1 Files To Add Or Change

| File | Change |
| --- | --- |
| `apps/server/main-server/internal/services/analytics/service.go` | new analytics service constructor and query methods |
| `apps/server/main-server/internal/services/analytics/types.go` | request-independent response DTOs and helper structs |
| `apps/server/main-server/internal/server/analytics_http.go` | Fiber handlers for all analytics routes |
| `apps/server/main-server/internal/server/state.go` | add `Analytics *analytics.Service` to `State` and initialize it in `NewState` |
| `apps/server/main-server/internal/server/routes.go` | mount `analytics` under `mountOrganizationScopedRoutes` |

Keep analytics in `internal/server`, not a separate handlers package, to match
the repo's current server layout.

### 4.2 Route Mounting

Inside `mountOrganizationScopedRoutes`, add:

```go
analytics := org.Group("/analytics")
analytics.Get("/overview", s.AnalyticsOverview)
analytics.Get("/rows-over-time", s.AnalyticsRowsOverTime)
analytics.Get("/pipeline-stats", s.AnalyticsPipelineStats)
analytics.Get("/run-duration-trend", s.AnalyticsRunDurationTrend)
analytics.Get("/failed-runs", s.AnalyticsFailedRuns)
analytics.Get("/usage", s.AnalyticsUsage)
analytics.Get("/source-destination-map", s.AnalyticsSourceDestinationMap)
analytics.Get("/export", s.AnalyticsExportCSV)
```

Because `mountOrganizationScopedRoutes` is already used by both:

- `/api/v1/organizations/:organizationId/...`
- `/api/organizations/:organizationId/...`

the analytics routes automatically get legacy parity without extra work.

### 4.3 Shared Backend Rules

#### Auth and tenancy

- All analytics routes stay under the existing authenticated org router:
  `AuthJWT()` + `RequireOrgMember()` + `MeterSDKAPICalls()`.
- Every analytics query must explicitly scope by `organization_id`. Do not rely
  on implicit RLS because the Go API uses a direct DB connection.
- When querying `pipeline_runs`, prefer direct `organization_id = ?`.
- When querying related pipeline metadata, join back to `pipelines` and keep
  the same org filter.

#### Period parsing

- Supported periods: `7d`, `30d`, `90d`
- Default period: `30d`
- Custom date ranges are not part of v1
- Period windows are **calendar-day aligned in UTC**

For a period of `N` days:

- `to = startOfDayUTC(now).AddDate(0, 0, 1)`
- `from = to.AddDate(0, 0, -(N))`

This produces a full-day inclusive window ending with the current UTC day.

Previous-period deltas use the immediately preceding window of the same length:

- `previousFrom = from.AddDate(0, 0, -N)`
- `previousTo = from`

#### Delta calculation

Use this rule everywhere a delta is returned:

```text
if previous == 0 and current == 0 => 0
if previous == 0 and current > 0 => 100
otherwise => ((current - previous) / previous) * 100
```

#### Limit parsing

- `pipeline-stats`: default `10`, max `50`
- `run-duration-trend`: default `30`, max `100`
- `failed-runs`: default `10`, max `50`

Reject invalid UUIDs and bad params with the repo's normal `response.Error` /
`response.SafeError` helpers.

### 4.4 Derived Metric Definitions

These definitions are the source of truth for backend and frontend.

| Metric | Definition |
| --- | --- |
| `rowsDelivered` | `SUM(pipeline_runs.rows_written)` |
| `partialRuns` | runs where `status = 'success'` and `run_metadata.raw_status = 'partial_success'` or `run_metadata.delivery_status = 'partial'` |
| `successfulRuns` | runs where `status = 'success'` and the run is **not** classified as partial |
| `failedRuns` | runs where `status = 'failed'` |
| `avgRunDurationSeconds` | average `duration_seconds` across terminal runs with non-null duration (`status IN ('success', 'failed')`) |
| `totalPipelines` | count of undeleted `pipelines` for the organization |
| `activePipelines` | count of undeleted `pipelines` with status in `idle`, `running`, `listing`, `listening`, `initializing` |
| `rowsLost` | `max(rows_failed, rows_read - rows_written, 0)` for failed-run rows |

#### Failed phase derivation

Because the repo does not store `phase_1_status` / `phase_2_status` /
`phase_3_status` columns, derive `errorPhase` from `run_metadata` in this
order:

1. `phase_3`
   if `delivery_status IN ('failed', 'partial')`
   or `delivery_failures` is non-empty
2. `phase_2`
   if `dbt_run_status = 'failed'`
   or `dbt_failed_models` is non-empty
   or `dbt_failed_tests` is non-empty
3. `phase_1`
   fallback for remaining failed runs

Do not emit `unknown` unless the record is malformed enough that no fallback is
possible.

### 4.5 Endpoint Contract

All JSON endpoints return `response.Success(...)`, so the HTTP payload is:

```json
{
  "meta": {
    "statusCode": 200,
    "message": "OK",
    "status": "success",
    "timestamp": "2026-04-20T12:00:00Z"
  },
  "data": {}
}
```

The shapes below describe the `data` object.

#### `GET /api/v1/organizations/:organizationId/analytics/overview`

Query params:

- `period=7d|30d|90d`

`data` shape:

```ts
interface AnalyticsOverview {
  period: { from: string; to: string; label: string };
  rowsDelivered: number;
  rowsDeliveredDeltaPct: number;
  successfulRuns: number;
  successfulRunsDeltaPct: number;
  partialRuns: number;
  failedRuns: number;
  failedRunsDeltaPct: number;
  avgRunDurationSeconds: number;
  avgRunDurationDeltaPct: number;
  totalPipelines: number;
  activePipelines: number;
}
```

Notes:

- `partialRuns` is included so the donut chart does not have to infer it.
- `successfulRuns` excludes partial runs.
- `period.label` should be `Last 7 days`, `Last 30 days`, or `Last 90 days`.

#### `GET /api/v1/organizations/:organizationId/analytics/rows-over-time`

Query params:

- `period=7d|30d|90d`

`data` shape:

```ts
interface AnalyticsRowsOverTimeResponse {
  data: AnalyticsRowsPoint[];
}

interface AnalyticsRowsPoint {
  date: string;
  rowsDelivered: number;
  successfulRuns: number;
  partialRuns: number;
  failedRuns: number;
}
```

Notes:

- Group by UTC day.
- Zero-fill missing dates so the chart always renders a continuous daily series.
- The chart plots `rowsDelivered`; tooltip can also surface successful,
  partial, and failed counts.

#### `GET /api/v1/organizations/:organizationId/analytics/pipeline-stats`

Query params:

- `period=7d|30d|90d`
- `limit=<int>`

`data` shape:

```ts
interface AnalyticsPipelineStatsResponse {
  pipelines: AnalyticsPipelineStat[];
}

interface AnalyticsPipelineStat {
  pipelineId: string;
  pipelineName: string;
  sourceConnectionName: string | null;
  sourceConnectorType: string | null;
  destinationConnectionName: string | null;
  destinationConnectorType: string | null;
  runCount: number;
  successCount: number;
  partialCount: number;
  failCount: number;
  rowsDelivered: number;
  avgDurationSeconds: number;
  successRatePct: number;
  lastRunAt: string | null;
  lastRunStatus: string | null;
}
```

Notes:

- Reuse the join pattern already present in `pipeline_http.go` for source and
  destination connector metadata.
- Order by `rowsDelivered DESC` in SQL.
- The run-frequency chart uses the same response but sorts the first eight rows
  client-side by `runCount DESC`.
- Do not add branch-color plumbing in v1. The chart should use the shared
  primary color because the analytics contract is pipeline-level, not branch-level.

#### `GET /api/v1/organizations/:organizationId/analytics/run-duration-trend`

Query params:

- `pipelineId=<uuid>`
- `limit=<int>`

`data` shape:

```ts
interface AnalyticsRunDurationTrendResponse {
  pipelineId: string;
  pipelineName: string;
  runs: AnalyticsRunDurationPoint[];
}

interface AnalyticsRunDurationPoint {
  runNumber: number;
  durationSeconds: number;
  status: "success" | "failed";
  rowsDelivered: number;
  createdAt: string;
}
```

Notes:

- Fetch the most recent terminal runs for the selected pipeline.
- Query newest-first for efficiency, then reverse in code so the response is
  oldest-to-newest and `runNumber` increases left-to-right.
- Only include terminal runs with non-null `duration_seconds`.

#### `GET /api/v1/organizations/:organizationId/analytics/failed-runs`

Query params:

- `period=7d|30d|90d`
- `limit=<int>`

`data` shape:

```ts
interface AnalyticsFailedRunsResponse {
  failedRuns: AnalyticsFailedRun[];
}

interface AnalyticsFailedRun {
  runId: string;
  pipelineId: string;
  pipelineName: string;
  failedAt: string;
  errorPhase: "phase_1" | "phase_2" | "phase_3";
  errorMessage: string | null;
  rowsRead: number;
  rowsDelivered: number;
  rowsLost: number;
}
```

Notes:

- `rowsRead` and `rowsDelivered` come from existing `pipeline_runs` columns.
- `rowsLost` is derived from existing run counters, not from nonexistent staged
  row fields.
- Use `error_message` directly for preview text.

#### `GET /api/v1/organizations/:organizationId/analytics/usage`

Query params:

- none

`data` shape:

```ts
interface AnalyticsUsage {
  plan: string;
  billingMonth: string;
  usage: {
    pipelines: AnalyticsUsageMetric;
    rowsThisMonth: AnalyticsUsageMetric;
    apiCallsToday: AnalyticsUsageMetric;
    stagingDiskGbThisMonth: AnalyticsUsageMetric;
  };
  overage: {
    extraRows: number;
    extraRowsCostCents: number;
    extraApiCalls: number;
    extraApiCostCents: number;
  };
}

interface AnalyticsUsageMetric {
  used: number;
  limit: number | null;
  pct: number | null;
}
```

Notes:

- This endpoint is an adapter over `internal/services/usage`.
- Reuse existing billing math from `CurrentUsage`; do not recalculate plan
  overages independently.
- Map:
  - `CurrentUsage.pipelines` -> `usage.pipelines`
  - `CurrentUsage.rowsProcessed` -> `usage.rowsThisMonth`
  - `CurrentUsage.apiCallsToday` -> `usage.apiCallsToday`
- Add `stagingDiskGbThisMonth` by summing
  `run_metadata.staging_size_bytes` for the current billing month and converting
  to GiB with one decimal place.
- Use `config/plans.go` for the staging size limit (`StagingSizeLimitGB`).
- If the staging limit is unlimited, return `limit: null` and `pct: null`.

#### `GET /api/v1/organizations/:organizationId/analytics/source-destination-map`

Query params:

- none

`data` shape:

```ts
interface AnalyticsSourceDestinationMapResponse {
  connections: AnalyticsSourceDestConnection[];
}

interface AnalyticsSourceDestConnection {
  sourceName: string;
  sourceConnectorType: string;
  destinationName: string;
  destinationConnectorType: string;
  pipelineCount: number;
  totalRowsDelivered: number;
}
```

Notes:

- Aggregate across the organization's active pipelines.
- `totalRowsDelivered` is all-time in v1 because the endpoint is not period-bound.
- This endpoint is backend-ready for future lineage UI, but the initial page
  does not need to render it.

#### `GET /api/v1/organizations/:organizationId/analytics/export`

Query params:

- `period=7d|30d|90d`

Response:

- raw CSV
- `Content-Type: text/csv; charset=utf-8`
- `Content-Disposition: attachment; filename="analytics-<period>.csv"`

CSV columns:

```text
runId,pipelineName,status,rowsDelivered,durationSeconds,createdAt
```

Notes:

- Export **all** runs in the selected period, not the page-limited failed-run
  or chart subsets.
- Join `pipeline_runs` to `pipelines` for `pipelineName`.
- Do not wrap the CSV response in `response.Success`.

### 4.6 Backend Implementation Notes

- Reuse `pipeline_http.go` join patterns for connector metadata instead of
  inventing new aliases.
- Reuse `usage.CurrentUsage(...)` and `config.GetPlan(...)` instead of copying
  billing logic into analytics.
- Keep response JSON camelCase even though DB columns are snake_case.
- Keep analytics read-only; no writes beyond existing retry behavior triggered
  by the frontend via the normal run endpoint.

---

## 5. Frontend Implementation

### 5.1 Files To Add Or Change

| File | Change |
| --- | --- |
| `apps/app/lib/api/types/analytics.ts` | add analytics TS contracts |
| `apps/app/lib/api/services/analytics.service.ts` | add analytics service methods |
| `apps/app/lib/api/hooks/use-analytics.ts` | add analytics query keys and hooks |
| `apps/app/lib/api/index.ts` | export analytics types, service, hooks |
| `apps/app/lib/analytics/formatters.ts` | add analytics-specific format helpers |
| `apps/app/app/workspace/analytics/page.tsx` | replace mock page with real implementation |
| `apps/app/app/workspace/analytics/components/*` | add route-local analytics components |
| `apps/app/components/workspace/workspace-sidebar.tsx` | re-enable Analytics nav item |

### 5.2 Public TypeScript Surface

Create `apps/app/lib/api/types/analytics.ts` with:

```ts
export type AnalyticsPeriod = "7d" | "30d" | "90d";

export interface AnalyticsOverview { ... }
export interface AnalyticsRowsPoint { ... }
export interface AnalyticsPipelineStat { ... }
export interface AnalyticsRunDurationPoint { ... }
export interface AnalyticsFailedRun { ... }
export interface AnalyticsUsage { ... }
export interface AnalyticsSourceDestConnection { ... }
```

Use the backend `data` shapes from section 4.5 exactly.

### 5.3 Service And Hook Contract

Create `apps/app/lib/api/services/analytics.service.ts` with:

```ts
class AnalyticsService {
  static getOverview(organizationId: string, period: AnalyticsPeriod)
  static getRowsOverTime(organizationId: string, period: AnalyticsPeriod)
  static getPipelineStats(
    organizationId: string,
    period: AnalyticsPeriod,
    limit = 10,
  )
  static getRunDurationTrend(
    organizationId: string,
    pipelineId: string,
    limit = 30,
  )
  static getFailedRuns(
    organizationId: string,
    period: AnalyticsPeriod,
    limit = 10,
  )
  static getUsage(organizationId: string)
  static getSourceDestinationMap(organizationId: string)
  static exportCSV(
    organizationId: string,
    period: AnalyticsPeriod,
  ): Promise<Blob>
}
```

Service routing:

- use `orgPath(organizationId)` for JSON endpoints
- endpoint suffixes are `/analytics/overview`, `/analytics/rows-over-time`, and
  so on

CSV export implementation rule:

- do **not** use `ApiClient.get<T>()` because it assumes JSON
- call `fetch(getApiUrl(...), await createFetchOptions(...))`
- return `await response.blob()`

Create `apps/app/lib/api/hooks/use-analytics.ts` with:

- `analyticsKeys.all`
- `analyticsKeys.overview(orgId, period)`
- `analyticsKeys.rowsOverTime(orgId, period)`
- `analyticsKeys.pipelineStats(orgId, period, limit)`
- `analyticsKeys.runDurationTrend(orgId, pipelineId, limit)`
- `analyticsKeys.failedRuns(orgId, period, limit)`
- `analyticsKeys.usage(orgId)`
- `analyticsKeys.sourceDestinationMap(orgId)`

Hook names:

- `useAnalyticsOverview`
- `useAnalyticsRowsOverTime`
- `useAnalyticsPipelineStats`
- `useAnalyticsRunDurationTrend`
- `useAnalyticsFailedRuns`
- `useAnalyticsUsage`
- `useAnalyticsSourceDestinationMap`

Hook rules:

- `staleTime: 5 * 60 * 1000` for all analytics queries
- `enabled: !!organizationId` for org-wide queries
- `enabled: !!organizationId && !!pipelineId` for the duration-trend hook

### 5.4 Page Composition

The page reads `currentOrganization?.id` from `useWorkspaceStore()` and owns the
shared `period` state.

Page-level state:

- `period: AnalyticsPeriod`
- `selectedPipelineId: string | null`
- `isExporting: boolean`

Selection behavior:

- default `selectedPipelineId` to the first pipeline returned by
  `useAnalyticsPipelineStats`
- if there are no pipelines, the duration trend card renders an empty state
- if the selected pipeline disappears after refetch, fall back to the new first
  pipeline

### 5.5 Route-Local Components

Create these files under `apps/app/app/workspace/analytics/components`:

| File | Responsibility |
| --- | --- |
| `PeriodSelector.tsx` | segmented `7d / 30d / 90d` control |
| `KpiCard.tsx` | reusable KPI card for the four top metrics |
| `RowsOverTimeChart.tsx` | area chart for `rowsDelivered` |
| `RunStatusDonut.tsx` | donut chart for success / partial / failed |
| `PipelineRunFrequencyChart.tsx` | horizontal bar chart using `runCount` |
| `RunDurationTrendChart.tsx` | line chart for selected pipeline run durations |
| `TopPipelinesTable.tsx` | top pipelines table |
| `RecentFailedRunsTable.tsx` | failed runs table with retry action |
| `UsageProgressBars.tsx` | plan usage bars and upgrade CTA |

Component-specific rules:

#### `PeriodSelector.tsx`

- purely controlled component
- values: `7d`, `30d`, `90d`
- page owns the selected state

#### `KpiCard.tsx`

- props should include label, value, deltaPct, and a tone mode
- delta tone rules:
  - rows delivered: up is good
  - successful runs: up is good
  - failed runs: down is good
  - avg duration: down is good, up is warning

#### `RowsOverTimeChart.tsx`

- use `AreaChart` inside `ChartContainer`
- plot `rowsDelivered`
- tooltip shows date, rows delivered, success, partial, and failed counts

#### `RunStatusDonut.tsx`

- donut segments:
  - success = `overview.successfulRuns`
  - partial = `overview.partialRuns`
  - failed = `overview.failedRuns`
- center label shows total runs for the period

#### `PipelineRunFrequencyChart.tsx`

- take the pipeline stats response, sort a derived copy by `runCount DESC`
- render at most 8 pipelines
- use a single primary bar color in v1
- include a link to `/workspace/data-pipelines`

#### `RunDurationTrendChart.tsx`

- render a pipeline selector above the chart
- point color by run status:
  - success -> green
  - failed -> red
- x-axis is `runNumber`
- y-axis is seconds

#### `TopPipelinesTable.tsx`

- columns:
  - pipeline
  - source type
  - runs
  - success rate
  - rows delivered
  - avg duration
  - last run
- clicking a row routes to `/workspace/data-pipelines/{id}`

#### `RecentFailedRunsTable.tsx`

- columns:
  - pipeline
  - failed at
  - phase
  - error preview
  - rows lost
  - action
- phase badge colors:
  - phase 1 -> blue
  - phase 2 -> purple
  - phase 3 -> orange
- truncate the error preview to ~60 chars and show the full message in a tooltip
- retry uses `DataPipelinesService.runPipeline(organizationId, pipelineId)`

#### `UsageProgressBars.tsx`

- render:
  - pipelines
  - rows this month
  - api calls today
  - staging disk this month
- progress colors:
  - `< 70%` -> blue
  - `70% to 90%` -> amber
  - `> 90%` -> red with pulse
- when a metric exceeds `90%`, show `Upgrade ->` linked to
  `/workspace/settings?tab=billing`

### 5.6 Formatters

Create `apps/app/lib/analytics/formatters.ts` with:

- `formatRows(n)`
- `formatDuration(seconds)`
- `formatDelta(pct, inverse = false)`
- `formatDate(isoDate, period)`

Formatting rules:

- `formatRows(45231) -> "45.2K"`
- `formatRows(1230000) -> "1.2M"`
- `formatRows(45000000) -> "45M"`
- `formatDuration(134) -> "2m 14s"`
- `formatDuration(45) -> "45s"`
- `formatDuration(3661) -> "1h 1m 1s"`
- `formatDate`:
  - `7d` -> `Mon 10`
  - `30d` -> `Jun 10`
  - `90d` -> `Jun 2025`

### 5.7 Query Invalidation And Actions

When retrying a failed run from the table:

1. call `DataPipelinesService.runPipeline(organizationId, pipelineId)`
2. on success, invalidate:
   - `analyticsKeys.all`
   - `dataPipelinesKeys.runs(organizationId, pipelineId, ...)` for visible run data
   - `dataPipelinesKeys.stats(organizationId, pipelineId)`

This keeps analytics and pipeline detail UIs in sync.

### 5.8 Frontend Implementation Notes

- The page should use the existing `PageHeader` action slot for the period
  selector and export button.
- Keep the existing `ChartContainer` wrapper instead of importing raw Recharts
  tooltips or responsive containers directly.
- Reuse existing `Skeleton` and `TableSkeleton` utilities when it helps, but do
  not collapse all analytics loading states into a single generic dashboard
  skeleton; each panel should match its final size.

---

## 6. Testing

### 6.1 Backend Unit And Integration Coverage

Add or update tests for:

- period parsing defaults to `30d`
- invalid periods fall back or reject consistently
- limit parsing caps oversized values
- delta calculation when previous = `0`
- overview counts separate success, partial, and failed correctly
- failed phase derivation from `run_metadata`
- every endpoint rejects cross-org access by filter shape
- rows-over-time zero-fills missing days
- CSV export includes all runs in the selected period

### 6.2 Frontend Coverage

Add tests for:

- period selector changes all period-bound query keys
- duration chart defaults to the first pipeline and resets cleanly
- empty state renders when there are no pipelines
- every chart and table has a matching-dimension skeleton
- failed-run retry invalidates analytics and pipeline-run queries
- usage CTA appears only above the `90%` threshold

### 6.3 Manual Verification Checklist

These steps should also exist in `md-docs/testing-local.md`:

1. Open `/workspace/analytics` with a populated organization.
2. Switch `7d`, `30d`, and `90d`; verify every card and chart updates.
3. Export CSV; verify the file includes all runs in the selected period.
4. Confirm usage bars change color near 70% and 90%.
5. Confirm the failed-runs table truncates long messages but preserves the full
   tooltip.
6. Retry a failed pipeline from the table and verify the page refetches.
7. Verify rows and pipeline data stay organization-scoped when switching orgs.
8. Confirm source and destination connector labels match the pipeline list joins.

---

## 7. Assumptions And Defaults

- `custom` periods are intentionally excluded from v1.
- `pipeline_runs.organization_id` already exists; no migration is part of this
  work.
- The source/destination map endpoint is required for backend completeness, but
  the initial page does not render a lineage widget.
- Analytics uses 5-minute cache freshness plus manual refresh; it does not use
  Supabase Realtime.
- The backend keeps using the repo's current org-member auth model. If product
  requirements later restrict analytics to owners/admins only, treat that as a
  separate authorization change.
