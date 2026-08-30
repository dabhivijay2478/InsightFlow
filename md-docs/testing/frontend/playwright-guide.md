# MantrixFlow Playwright End-to-End & Integration Testing Guide

This directory contains the Playwright end-to-end (E2E), accessibility, responsive, and backend integration test suites for the MantrixFlow frontend application.

## Prerequisites

- [Bun](https://bun.sh/) installed locally
- Node.js 20+

## Quick Start

1. Install Playwright browsers (if not already installed):
   ```bash
   bunx playwright install chromium firefox webkit
   ```

2. Copy the environment configuration template:
   ```bash
   cp .env.e2e.example .env.e2e
   ```

3. Run smoke test suite:
   ```bash
   bun run test:e2e:smoke
   ```

4. Run mocked regression suite:
   ```bash
   bun run test:e2e:mocked
   ```

5. Show HTML test report:
   ```bash
   bun run test:e2e:report
   ```

## Available Test Scripts

| Command | Description |
| :--- | :--- |
| `bun run test:e2e` | Run all E2E test suites |
| `bun run test:e2e:smoke` | Run critical `@smoke` tests |
| `bun run test:e2e:regression` | Run complete `@regression` suite |
| `bun run test:e2e:mocked` | Run API-intercepted deterministic frontend tests |
| `bun run test:e2e:integration` | Run real control-plane integration tests |
| `bun run test:e2e:ui` | Open interactive Playwright UI mode |
| `bun run test:e2e:headed` | Run tests in headed browser mode |
| `bun run test:e2e:debug` | Debug tests with Playwright Inspector |
| `bun run test:e2e:report` | Serve generated HTML test report |

## Running Real Integration Tests

To run real pipeline execution tests against a live Go control-plane API:

```bash
TEST_RUN_REAL_INTEGRATIONS=true bun run test:e2e:integration
```

## Structure

```text
tests/
  setup/            # Authentication setup project & storage state persistence
  fixtures/         # Custom Playwright fixtures (auth, API client, loggers)
  helpers/          # API client, network interceptors, cleanup utilities
  page-objects/     # Modular Page Object Model classes
  auth/             # Login, signup, password reset tests
  onboarding/       # Route-driven onboarding flow tests
  organizations/    # Organization management & OrganizationDialog tests
  workspace-shell/  # Sidebar, topbar, navigation & global search tests
  dashboard/        # KPIs, recent activity, recent migrations tests
  connections/      # Connection catalog, forms, secret masking tests
  data-sources/     # Schema discovery & preview tests
  sql-explorer/     # Query editor & execution tests
  pipelines/        # Pipeline creation, workspace, tabs, runs & integration tests
  team/             # Workspace members, role & status filter tests
  activity/         # Activity log list & date filter tests
  notifications/    # Notification list & detail sheet tests
  analytics/        # Analytics summary & chart tests
  settings/         # Profile, security & org settings tests
  billing/          # Usage meters & plan limit tests
  accessibility/    # Automated @axe-core/playwright scans
  responsive/       # Desktop, Tablet, Mobile viewport tests
```
