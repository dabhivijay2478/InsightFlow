# AI Agent / Twenty experience audit

Date: 2026-08-24

## Compared revisions

- Twenty reference: `twentyhq/twenty` main at `bb4469747870856ac558b885ffdf0733677390fc`.
- MantrixFlow frontend baseline: `cd488791db6bcec6c1f2cdb4e41874bc109dcebc`.
- MantrixFlow Go baseline: `1bc2c1c126f2a5b847699b7059e6904f05b9e77e`.
- MantrixFlow ELT baseline: `021d4dccecf8003141a783ea9b8b8a8029990914`.

The reference was inspected from a fresh sparse checkout of current Twenty main. The implementation is original MantrixFlow code and reuses the existing internal Oria runtime, authorization, context, action-confirmation, and thread APIs. Product-facing chat copy now uses **AI Agent**; stable Oria API, event, and persistence identifiers remain unchanged for compatibility.

## Reference findings

Twenty uses a single chat surface for its full-page and side-panel entry points. Its important UX patterns are a narrow centered transcript, compact product header, thread-aware drafts, a bottom composer with context/model/file controls, queued follow-up messages, drag-and-drop attachment feedback, explicit streaming controls, tool/reasoning state, and a latest-message affordance. Thread and composer state live above the view so switching surfaces does not create a second chat implementation.

## Implemented parity

| Capability | Status | MantrixFlow implementation |
| --- | --- | --- |
| Shared page/panel chat surface | READY | `OriaChatSurface` renders both routes and the existing drawer. |
| Branded sidebar header | READY | The existing MantrixFlow logo and project name remain in the first row; a smaller Home/Chat switch and New chat action sit below it. |
| Sidebar mode behavior | READY | Home renders workspace search/navigation; Chat renders history without navigating or resetting the central page. Agent routes initially select Chat. |
| Responsive sidebar | READY | The switch, New chat, history, and persistent footer controls support expanded, collapsed, mobile-drawer layouts. |
| Compact chat header | READY | Full-page duplicate history/New chat controls were removed; panel-only history/New chat actions remain available. |
| Centered responsive transcript | READY | One 768px-class content column with mobile-safe padding and overflow. |
| Empty state and contextual starters | READY | Focus-aware AI Agent welcome and suggested prompts. |
| Per-thread composer drafts | READY | Drafts are retained by thread key in the internal chat context provider. |
| Send, stop, and queued follow-ups | READY | Enter/Shift+Enter behavior, stop while streaming, and removable queue items. |
| Model mode selector | READY | Exposes only supported `Auto` and `Deep analysis` modes. |
| Context preview | READY | Displays the sanitized focused workspace reference and supports removal. |
| Text-file attachments | READY | Local text extraction with type, empty-file, and 32 KB limits; no credentials or browser storage. |
| Drag-and-drop state | READY | Full-surface dropzone with visible overlay. |
| Streaming/tool/reasoning display | READY | Existing streamed messages plus restrained operational reasoning and tool status. |
| Action confirmation | READY | Existing pending-action cards remain above the composer and keep server authorization. |
| Scroll recovery | READY | Floating `Latest` action uses assistant-ui scroll state. |
| Thread history | READY | Search, loading/error/empty states, date groups, active highlights, timestamps, rename, archive, restore, delete, and URL-backed selection. |
| Accessibility | READY | Named controls, keyboard composer behavior, focus-safe shadcn primitives, and semantic status output. |

## Deliberate architecture boundaries

- The AI Agent currently executes the model stream in the Next.js orchestrator after the Go service prepares and finalizes the run. This rewrite does not move execution to Go because that would change backend contracts outside the requested chat-layout scope.
- The present transport deduplicates sequenced SSE events, but the backend does not expose Twenty-equivalent stream catch-up/replay. Reconnect parity is therefore PARTIAL.
- Attachments are deliberately limited to bounded text formats supported by the current request pipeline. There is no new object-storage upload contract.
- The selector exposes response modes rather than invented provider model IDs. Workspace policy remains the authority for the actual model.
- Twenty-equivalent structured question widgets are not introduced because the current internal runtime has no corresponding typed question event contract.

## Security and tenancy review

- Existing authenticated Oria routes, organization scoping, authorization, run preparation, and action confirmation are unchanged.
- Focused context is sanitized in the browser and re-resolved by the server; the UI does not claim a client reference is authorization.
- Attachments are converted into bounded request text and are subject to the same server-side message validation. No secrets, cookies, browser storage, credentials, or unrestricted file URLs are persisted.
- Message rendering continues through the existing sanitized response renderer.

## Verification record

- Protected-route browser check: `/agents` responded and redirected to `/auth/login?next=%2Fagents` in the available signed-out session. The page had meaningful content and no Next.js error overlay. Authenticated AI Agent interaction was unavailable without a signed-in session.
- Changed-file Biome check: PASS (31 files).
- TypeScript `tsc --noEmit`: PASS.
- Next.js production build: PASS (84 pages); existing DuckDB dynamic-dependency and stale `baseline-browser-mapping` warnings remain.
- Full-repository Biome check: BLOCKED by 23 pre-existing errors and 15 warnings in untouched Oria server, intelligence, automation, developer, and enterprise files. No diagnostics are present in the rewritten files.
- Changed-file length audit: PASS; the largest touched file is `features/workspace-shell/components/workspace-sidebar.tsx` at 408 lines.
- Repository-wide length audit still finds two untouched files above the 500-line policy: `features/team/components/team-screen.tsx` (546) and `features/ai-copilot/server/agent/orchestrator.ts` (526).

## Overall status

The requested Twenty-style AI Agent page, branded sidebar modes, history navigation, and shared chat interaction structure are READY. Full parity with Twenty's backend replay/catch-up, structured questions, and stored uploads is PARTIAL by design and would require separately authorized API and persistence work.
