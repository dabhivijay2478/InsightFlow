# Quick Start Guide - Agent Grid View

## How to Use

### Access the Grid View
1. Navigate to `/workspace/agents`
2. You'll see the agent grid view by default

### Features to Try

#### 🔍 Search
- Type in the search box to find agents by:
  - Agent name
  - Description
  - Pipeline name
- Results update in real-time

#### 🏷️ Filter by Status
- Click the "Status" dropdown
- Select: All Statuses, Active, or Inactive
- Grid updates immediately

#### 📊 Sort Results
- Click the "Sort By" dropdown
- Choose: Name, Created, or Updated
- Cards reorganize instantly

#### 📲 Open an Agent
- Click any agent card or the "Open Agent" button
- Takes you to the agent chat view for that pipeline
- Shows: Chat interface, context panel, publisher panel

#### ← Go Back to Grid
- From chat view, navigate to `/workspace/agents`
- Returns to the grid view

## Component Structure

```
page.tsx (async server component)
  ↓
  ├─ AgentGridClient (new - card grid view)
  │   ├─ Header (title + agent count)
  │   ├─ Filters (search + dropdown controls)
  │   ├─ Grid Layout (responsive cards)
  │   └─ AgentCard (individual agent card component)
  │
  └─ AgentsPlatformClient (existing - chat view)
```

## API Calls Made

### When Page Loads
1. `GET /api/v1/orgs/{orgId}/pipelines` - Fetch all pipelines
2. `GET /api/v1/orgs/{orgId}/pipelines/{pipelineId}/agent` - For each pipeline
   - Makes parallel requests for all pipelines
   - Handles failures gracefully

## State Management

The component manages:
- **searchQuery** - User's search text
- **statusFilter** - Selected status filter
- **sortBy** - Selected sort field
- **agents** - All loaded agents with pipeline names
- **isLoading** - Loading state while fetching agents
- **pipelinesQuery** - TanStack Query for pipelines

## Performance Considerations

✅ **Optimized**
- Parallel agent fetching (Promise.all)
- Memoized filtering and sorting
- Lazy loading with scroll area
- Graceful error handling

⚠️ **Potential Improvements**
- Pagination for large agent lists (100+)
- Server-side filtering
- Caching agent data
- Virtual scrolling for huge lists

## Styling Details

### Colors
- Active Badge: Primary color
- Inactive Badge: Secondary color
- Hover Effects: Primary/50 border, shadow
- Gradient Overlay: Primary/5% on hover

### Typography
- Title: 24px, bold
- Card Title: 14px, semibold
- Metadata: 12px, light
- Stats: 12px, regular

### Spacing
- Page padding: 24px (p-6)
- Card gap: 16px (gap-4)
- Internal card padding: 16px (p-4)
- Section spacing: 16px (gap-4)

## Testing Scenarios

### Test 1: Empty State
- Create an organization with no agents
- Should show: "No agents yet. Create one from a pipeline."

### Test 2: Search Functionality
- Create agent named "UserValidator"
- Search for "valid" → should appear
- Search for "xyz" → should disappear

### Test 3: Status Filter
- Create both active and inactive agents
- Filter by "Active" → only show active agents
- Filter by "All Statuses" → show all

### Test 4: Sort
- Create agents with different dates
- Sort by "Updated" → newest first
- Sort by "Name" → alphabetical order

### Test 5: Navigation
- Click "Open Agent" button
- Should navigate to `/workspace/agents?pipelineId={id}&view=chat`
- Back button or clicking agents link returns to grid

## Troubleshooting

### Q: Agents not showing?
- A: Check that pipelines exist and have agents created
- Check browser console for API errors
- Verify organization ID is loaded

### Q: Search not working?
- A: Ensure searchQuery state is updating
- Check that agent names/descriptions contain search text
- Try with exact matches first

### Q: Cards not responsive?
- A: Resize browser window
- Check that TailwindCSS is properly configured
- Verify `md:` and `lg:` breakpoints are active

### Q: Filters not updating?
- A: State updates should be immediate
- Check browser DevTools > React Dev Tools
- Verify Select component is properly connected

## Future Enhancements

```typescript
// Possible additions:

// 1. Bulk Actions
const [selectedAgents, setSelectedAgents] = useState<string[]>([]);
// Enable/disable multiple agents at once

// 2. Quick Stats
// Total agents, Active/Inactive count in header

// 3. Agent Comparison
// Side-by-side view of multiple agents

// 4. Favorites
// Pin agents to top of grid

// 5. Export
// Download agent list as CSV/JSON

// 6. Webhooks Info
// Quick access to webhook URLs

// 7. Performance Metrics
// Agent usage stats, last run time

// 8. Inline Editing
// Edit agent details directly from card
```

## File Locations Reference

- **Grid Component**: `apps/app/app/workspace/agents/AgentGridClient.tsx`
- **Page Router**: `apps/app/app/workspace/agents/page.tsx`
- **Hooks**: `apps/app/lib/api/hooks/use-agents.ts`
- **Documentation**: 
  - `AGENT_GRID_IMPLEMENTATION.md` - Technical overview
  - `AGENT_GRID_UI_GUIDE.md` - Design guide

## Commands to Run

```bash
# Navigate to app directory
cd apps/app

# Run development server
bun run dev

# Run linter
bun run lint

# Build for production
bun run build

# Navigate to agents page
# http://localhost:3000/workspace/agents
```

## Support

For questions or issues:
1. Check the documentation files
2. Review component code in AgentGridClient.tsx
3. Check browser console for errors
4. Verify API responses in Network tab
