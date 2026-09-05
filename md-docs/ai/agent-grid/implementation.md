# Agent Grid View Implementation

## Overview
Added a new **Agent Grid View** to the workspace agents page with filtering and search capabilities.

## Changes Made

### 1. **AgentGridClient.tsx** (New Component)
- Displays all agents across pipelines in a responsive card grid
- Location: `/apps/arcyria-platform/app/workspace/agents/AgentGridClient.tsx`
- Features:
  - **Responsive Grid Layout**: 1 column (mobile) → 2 columns (tablet) → 3 columns (desktop)
  - **Agent Cards** with:
    - Active/Inactive status badge
    - Agent name and associated pipeline
    - Description (clamped to 2 lines)
    - Statistics (destination & source table counts)
    - Permissions info (public queries, allowed domains)
    - Created/Updated dates
    - "Open Agent" button

### 2. **Filtering & Search**
The grid includes three filter/sort controls:

#### Search Box
- Searches agent name, description, or pipeline name
- Real-time filtering

#### Status Filter (Dropdown)
- All Statuses
- Active
- Inactive

#### Sort By (Dropdown)
- Name (A-Z)
- Created (newest first)
- Updated (newest first)

### 3. **Updated Agents Page** (`page.tsx`)
- Added `view` query parameter support
- Routes:
  - `/workspace/agents` → Shows grid view (default)
  - `/workspace/agents?view=grid` → Explicit grid view
  - `/workspace/agents?pipelineId={id}&view=chat` → Pipeline agent chat view
- Integrated with existing `AgentsPlatformClient` for backward compatibility

### 4. **New Hook** (`use-agents.ts`)
- Created agent-specific React Query hooks
- Provides standardized query key structure
- Ready for future expansion (get all agents, etc.)

## UI Features

### Agent Card Design
```
┌─────────────────────────────────────┐
│ [Active] [Can Trigger Runs]         │
│ Agent Name                          │
│ Pipeline: Pipeline Name             │
│ Description text...                 │
│ • 5 dest tables • 3 src tables      │
│ [Public Queries] [2 domains]        │
│ ─────────────────────────────────── │
│ Created: 05/02/2026                 │
│ Updated: 05/02/2026                 │
│ ┌─────────────────────────────────┐ │
│ │     Open Agent Button           │ │
│ └─────────────────────────────────┘ │
└─────────────────────────────────────┘
```

### Hover Effects
- Subtle shadow and border color change on hover
- Gradient background effect
- Button styling changes on hover
- Smooth CSS transitions

## User Flow

1. **Visit Agents Page**
   - `/workspace/agents` shows grid of all agents

2. **Search & Filter**
   - Use search box to find agents
   - Filter by status (active/inactive)
   - Sort by name, created date, or updated date

3. **Open Agent**
   - Click "Open Agent" button
   - Redirects to `/workspace/agents?pipelineId={id}&view=chat`
   - Shows agent chat interface for that specific pipeline

4. **Back to Grid**
   - Navigate to `/workspace/agents` to return to grid view

## API Integration

The component:
- Fetches all pipelines using existing `usePipelines` hook
- For each pipeline, makes API call to get its agent
- Gracefully handles pipelines with no agents
- Shows loading state while fetching
- Displays empty state when no agents exist

## Responsive Design

### Mobile (< 768px)
- 1 column grid
- Full-width cards
- Compact spacing

### Tablet (768px - 1024px)
- 2 column grid
- Filters in column layout

### Desktop (> 1024px)
- 3 column grid
- Filters in row layout with search on left

## Future Enhancements

Potential additions:
- Bulk actions (enable/disable multiple agents)
- Agent creation from grid view
- Favorites/pinning agents
- Agent performance metrics
- Webhook history quick view
- Export agents list
- Agent comparison view
