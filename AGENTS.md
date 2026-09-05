# Agents (MantrixFlow)

**Authoritative agent onboarding and specialized review personas:**

[`.cursor/rules/agents-and-orchestration.mdc`](.cursor/rules/agents-and-orchestration.mdc)

**Stack map, local dev commands, and repo conventions:**

[`.cursor/rules/mantrixflow-repo.mdc`](.cursor/rules/mantrixflow-repo.mdc)

**Backend services (scoped rules — apply when editing those trees):**

- [`.cursor/rules/go-main-server.mdc`](.cursor/rules/go-main-server.mdc) — `apps/server/arcyria-server` (Fiber, dispatch, callbacks, queue)
- [`.cursor/rules/python-elt-server.mdc`](.cursor/rules/python-elt-server.mdc) — `apps/server/arcyria-elt` (FastAPI, DuckDB-staged runner)

**Supabase RLS workflow and schema map:**

[`.cursor/rules/supabase-rls.mdc`](.cursor/rules/supabase-rls.mdc)

**ELT (strict invariants + flow diagram):**

- [`.cursor/rules/strict-elt-invariants.mdc`](.cursor/rules/strict-elt-invariants.mdc) — 12 strict ELT invariants
- [`.cursor/rules/elt-flow-diagram.mdc`](.cursor/rules/elt-flow-diagram.mdc) — 5-phase flow

**Human-readable pipeline guide:** [`md-docs/architecture/elt/strict-pipeline-guide.md`](./md-docs/architecture/elt/strict-pipeline-guide.md)

**Oria AI SDK runtime setup:** [`md-docs/ai/oria/agent-setup.md`](./md-docs/ai/oria/agent-setup.md)

**Frontend structure / TS conventions:** [`.cursor/rules/nextjs-typescript-stack.mdc`](.cursor/rules/nextjs-typescript-stack.mdc)



# MantrixFlow Frontend Codex Rules

These rules apply to every frontend change made in this repository.

## 1. File-size rule

Every maintained frontend source file must contain no more than 500 lines.

Applies to:

```text
.ts
.tsx
.js
.jsx
.css
.scss
.less
```

Generated code, lock files, build output, and third-party vendor files are excluded.

Do not avoid the rule by creating meaningless fragments. Split code by responsibility.

## 2. Page rules

Next.js `page.tsx` files must remain small and compositional.

A page may:

* Read route parameters
* Read search parameters
* Load page-level server data
* Render a page container
* Handle route-level missing states

A page must not contain:

* Full feature implementations
* Large forms
* Table engines
* Column definitions
* API client implementations
* Large dialogs
* Large business-rule functions

## 3. Server/client rules

Components are server components by default.

Use `"use client"` only when required for:

* State
* Effects
* Browser APIs
* Event handlers
* Client context
* Interactive libraries

Do not turn an entire page or layout into a client component because one child is interactive.

Extract the interactive section into a focused client component.

## 4. Navigation rules

Use `next/link` for all internal visible links.

Use:

```tsx
<Link href="/pipelines">Pipelines</Link>
```

Do not use:

```tsx
<a href="/pipelines">Pipelines</a>
```

Raw anchors are allowed only for external links, downloads, `mailto:`, or `tel:`.

Do not use `window.location.href` for normal internal navigation.

## 5. Table rules

All feature tables must use the shared TanStack Table-based `DataTable`.

Do not create new:

* Custom table engines
* Custom pagination systems
* Custom sorting systems
* Repeated filter toolbars
* Repeated column visibility controls
* Repeated row-selection implementations
* Repeated table loading states

Feature tables may define their own columns, filters, and actions, but must reuse the shared table system.

## 6. shadcn/ui rules

Use shared shadcn/ui components for reusable interactive patterns.

Before creating a component, check:

```text
components/ui/
components/shared/
components/forms/
components/dialogs/
components/feedback/
components/data-table/
```

Do not create custom versions of:

* Buttons
* Inputs
* Selects
* Checkboxes
* Switches
* Dialogs
* Alert dialogs
* Dropdown menus
* Tooltips
* Tabs
* Badges
* Cards
* Skeletons
* Alerts
* Popovers
* Sheets
* Drawers
* Toasts
* Pagination controls

Use semantic HTML for document structure when appropriate.

## 7. Shared component rule

Before implementing a reusable pattern:

1. Search the repository for an existing implementation.
2. Reuse it when behavior matches.
3. Extend it carefully when the new requirement is generally useful.
4. Create a new component only when the behavior is meaningfully different.

Do not create feature-specific copies of shared components.

Do not build overly generic components with many unrelated boolean props.

## 8. Error-handling rules

Empty error handling is prohibited.

Do not write:

```ts
catch {}
catch (error) {}
.catch(() => {})
onError: () => {}
```

Every error must be:

* Displayed
* Normalized
* Logged
* Reported
* Rethrown
* Or explicitly and safely ignored with a comment

Never log secrets, credentials, tokens, connection strings, or sensitive customer data.

## 9. Comment rules

Do not keep commented-out code.

Remove:

* Old JSX implementations
* Old functions
* Commented imports
* Blank comments
* Decorative separator comments
* Comments that repeat the code
* Obsolete TODOs and FIXMEs

Keep comments only for non-obvious business rules, security behavior, framework limitations, API inconsistencies, and necessary workarounds.

Use Git history instead of preserving old code in comments.

## 10. Form rules

Forms should use:

* React Hook Form
* Zod
* shadcn Form components
* Shared field components
* Feature-level schemas
* Typed default values
* Payload mappers

Do not combine a complete form, validation schema, API calls, mapping logic, and dialog implementation in one large file.

## 11. Dialog rules

Use shadcn `Dialog` or `AlertDialog`.

Large dialogs must live in separate files.

Do not render one dialog for every table row.

Use one controlled dialog and store the selected record.

## 12. API rules

UI components must not contain large API implementations.

API requests belong in:

```text
services/
features/<feature>/services/
```

Server-state behavior belongs in focused hooks or the existing query architecture.

Do not duplicate API clients or error normalization.

Do not change API contracts without explicit instructions.

## 13. Business-logic rules

Business rules belong in:

```text
hooks/
services/
utils/
domain helpers/
selectors/
mappers/
```

Presentation components should focus on rendering and user interaction.

Do not move all complexity into a single oversized hook.

Hooks must remain focused and under 500 lines.

## 14. TypeScript rules

Avoid:

```text
any
Duplicated interfaces
Large inline object types
Unexplained type assertions
Repeated string-status values
```

Use:

* Feature-level types
* Shared domain types
* Typed constants
* Discriminated unions
* Runtime validation at external boundaries

Do not weaken types only to make compilation pass.

## 15. Import-boundary rules

Allowed direction:

```text
app → features → shared components
features → shared hooks/services/types
shared modules → lower-level shared modules
```

Disallowed:

```text
shared component → feature implementation
feature A → private internals of feature B
circular feature dependencies
```

Avoid giant barrel files that expose every private module.

## 16. Styling rules

Use the existing Tailwind and design-token system.

Avoid repeating long identical class lists across many files.

Use a shared component or variant helper when the style represents one reusable UI pattern.

Do not create new custom CSS when Tailwind or the existing component system already supports the requirement.

## 17. Accessibility rules

Preserve:

* Semantic elements
* Labels
* Keyboard navigation
* Focus behavior
* Accessible names
* Dialog descriptions
* Table headers
* Error descriptions
* Screen-reader content
* Disabled states

Do not remove shadcn accessibility behavior while wrapping components.

## 18. Performance rules

Avoid:

* Duplicate API requests
* Duplicate source-preview requests
* Recreated large column definitions
* One dialog per row
* Large derived arrays on every render
* Unnecessary effects
* Duplicated server state in local state
* Unnecessary client components

Use memoization only where it solves a real issue.

## 19. Testing scope

For frontend restructuring tasks, do not create or run:

* Playwright tests
* E2E tests
* Unit tests
* Integration tests

Testing is handled separately unless the user explicitly changes the scope.

Allowed checks:

```text
formatter
lint
TypeScript type-check
production build
file-length audit
architecture audit
```

## 20. Pre-completion audit

Before completing any frontend refactoring task, verify:

```text
No maintained source file exceeds 500 lines
No new duplicate table implementation exists
All internal links use next/link
No empty catch blocks exist
No commented-out implementation remains
Shared shadcn components are reused
Feature tables use shared DataTable
Page files remain small
Client components are justified
API logic is separated from UI
Business logic is separated from presentation
Permissions and plan restrictions remain intact
```

## 21. Final reporting

Every major refactoring response must report:

* Files changed
* Components extracted
* Duplicate implementations removed
* Tables migrated
* Internal links corrected
* Empty error blocks corrected
* Comments and dead code removed
* Largest resulting file
* Files remaining above 500 lines
* Architecture exceptions
* Lint result
* Type-check result
* Build result

Do not claim completion when architecture violations remain undocumented.
