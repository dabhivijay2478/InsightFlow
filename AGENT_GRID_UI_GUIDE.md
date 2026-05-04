# Agent Grid View - Visual Guide

## Layout Structure

```
┌──────────────────────────────────────────────────────────────────┐
│                         AGENTS PAGE                              │
├──────────────────────────────────────────────────────────────────┤
│ 🤖 Agents                                                        │
│ Browse and manage all pipeline agents (5 total)                 │
├──────────────────────────────────────────────────────────────────┤
│                                                                  │
│ ┌─────────────────────────────────┬──────────┬─────────────────┐ │
│ │ 🔍 Search agents...             │ Status ▼ │ Sort By ▼      │ │
│ └─────────────────────────────────┴──────────┴─────────────────┘ │
│                                                                  │
├──────────────────────────────────────────────────────────────────┤
│                                                                  │
│  ┌─────────────────────┐  ┌─────────────────────┐              │
│  │ [Active] [Trigger]  │  │ [Active] [Trigger]  │              │
│  │ Agent Name          │  │ Agent Name          │              │
│  │ Pipeline: Users ETL │  │ Pipeline: Analytics │              │
│  │ Description...      │  │ Description...      │              │
│  │                     │  │                     │              │
│  │ • 5 dest • 3 src    │  │ • 2 dest • 4 src    │              │
│  │ ──────────────────  │  │ ──────────────────  │              │
│  │ Created: 05/02/2026 │  │ Created: 04/15/2026 │              │
│  │ Updated: 05/01/2026 │  │ Updated: 05/01/2026 │              │
│  │                     │  │                     │              │
│  │ [  Open Agent   ]   │  │ [  Open Agent   ]   │              │
│  └─────────────────────┘  └─────────────────────┘              │
│                                                                  │
│  ┌─────────────────────┐  ┌─────────────────────┐              │
│  │ [Inactive]          │  │ [Active]            │              │
│  │ Legacy Worker       │  │ Data Validator      │              │
│  │ Pipeline: Old Sync  │  │ Pipeline: Validation│              │
│  │ Description...      │  │ Description...      │              │
│  │                     │  │                     │              │
│  │ • 1 dest • 2 src    │  │ • 3 dest • 5 src    │              │
│  │ ──────────────────  │  │ ──────────────────  │              │
│  │ Created: 03/10/2026 │  │ Created: 05/02/2026 │              │
│  │ Updated: 03/10/2026 │  │ Updated: 05/02/2026 │              │
│  │                     │  │                     │              │
│  │ [  Open Agent   ]   │  │ [  Open Agent   ]   │              │
│  └─────────────────────┘  └─────────────────────┘              │
│                                                                  │
└──────────────────────────────────────────────────────────────────┘
```

## Filter Controls

### Search Input
- **Icon**: 🔍 Magnifying glass
- **Placeholder**: "Search agents by name, description or pipeline..."
- **Searches**: Agent name, description, pipeline name
- **Real-time**: Results update as you type

### Status Filter Dropdown
```
┌──────────────────┐
│ Status ▼         │
├──────────────────┤
│ All Statuses     │
│ ✓ Active         │
│ Inactive         │
└──────────────────┘
```

### Sort By Dropdown
```
┌──────────────────┐
│ Sort By ▼        │
├──────────────────┤
│ ✓ Name           │
│ Created          │
│ Updated          │
└──────────────────┘
```

## Agent Card Components

### Header Section
```
┌─────────────────────────────────┐
│ [Active] [Can Trigger Runs]    │  ← Status badges
└─────────────────────────────────┘
```

### Agent Info
```
┌─────────────────────────────────┐
│ Agent Name                      │  ← Font: semibold, sm
│ Pipeline: Pipeline Name         │  ← Font: xs, muted
└─────────────────────────────────┘
```

### Description
```
┌─────────────────────────────────┐
│ This agent helps with data...   │  ← Max 2 lines
└─────────────────────────────────┘
```

### Statistics
```
┌─────────────────────────────────┐
│ • 5 dest tables  • 3 src tables │  ← Bullet-separated
└─────────────────────────────────┘
```

### Permissions (Optional)
```
┌─────────────────────────────────┐
│ [Public Queries] [2 domains]   │  ← Only if applicable
└─────────────────────────────────┘
```

### Metadata
```
┌─────────────────────────────────┐
│ Created: 05/02/2026            │
│ Updated: 05/01/2026            │
└─────────────────────────────────┘
```

### Action Button
```
┌─────────────────────────────────┐
│      [  Open Agent  ]           │  ← Full width
└─────────────────────────────────┘
```

## Responsive Breakpoints

### Mobile (< 768px)
- 1 column
- Filters stack vertically
- Search box full width
- Card padding: 16px

### Tablet (768px - 1024px)
- 2 columns
- Filters: search on top, dropdowns below
- Card spacing: 16px

### Desktop (> 1024px)
- 3 columns
- Filters: search on left, dropdowns on right
- Max card height: auto
- Card spacing: 16px

## Interaction States

### Card Hover
```
Border: primary/50 (subtle color change)
Shadow: md (light shadow)
Title: color → primary
Content opacity: slight increase
Button: bg-primary, text-primary-foreground
Background: subtle gradient overlay
```

### Button Hover
```
Before: variant="outline"
After: bg-primary, text-primary-foreground
Transition: smooth color change
```

## Search & Filter Examples

### Example 1: Search by Agent Name
```
Search: "validation"
Result: Shows only agents with "validation" in name
```

### Example 2: Find Inactive Agents
```
Status: "Inactive"
Result: Shows only agents where isActive = false
```

### Example 3: Sort by Creation Date
```
Sort By: "Created"
Result: Agents sorted by creation date (newest first)
```

### Example 4: Combined Filters
```
Search: "user"
Status: "Active"
Sort By: "Updated"
Result: Active agents matching "user", sorted by update date
```

## Empty States

### No Agents Yet
```
┌────────────────────────────────┐
│                                │
│        🤖 (40% opacity)        │
│                                │
│  No agents yet.                │
│  Create one from a pipeline.   │
│                                │
└────────────────────────────────┘
```

### No Matches for Filters
```
┌────────────────────────────────┐
│                                │
│        🤖 (40% opacity)        │
│                                │
│  No agents match your filters. │
│                                │
└────────────────────────────────┘
```

### Loading State
```
┌────────────────────────────────┐
│                                │
│    ⟳ Loading agents...        │  ← Spinning icon
│                                │
└────────────────────────────────┘
```

## Navigation Flow

```
                    ┌──────────────────┐
                    │  Agents Page     │
                    │  (Grid View)     │
                    └────────┬─────────┘
                             │
                    ┌────────▼────────┐
                    │  Search/Filter  │
                    └────────┬────────┘
                             │
        ┌────────────────────┼────────────────────┐
        │                    │                    │
        ▼                    ▼                    ▼
   [Agent Card 1]    [Agent Card 2]    [Agent Card 3]
        │                    │                    │
        └────────────────────┼────────────────────┘
                             │
                    ┌────────▼────────┐
                    │  Click Card     │
                    │  "Open Agent"   │
                    └────────┬────────┘
                             │
                    ┌────────▼────────┐
                    │  Agent Chat     │
                    │  View           │
                    │  (Pipeline View)│
                    └─────────────────┘
```

## Color Scheme

- **Primary**: Active badges, text on hover, button background
- **Muted**: Descriptive text, icons
- **Secondary**: Inactive badges
- **Background/Card**: Elevated surfaces
- **Border**: Subtle dividers
- **Primary/5%**: Hover gradient overlay

## Typography

- **Header**: 24px, bold
- **Subheader**: 14px, semibold
- **Card Title**: 14px, semibold
- **Description**: 12px, regular
- **Metadata**: 12px, light
- **Badge**: 12px/10px, semibold
