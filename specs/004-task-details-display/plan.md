# Implementation Plan: Task Details Display

**Branch**: `004-task-details-display` | **Date**: 2026-07-13 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/004-task-details-display/spec.md`

## Summary

Add Large Type support (Cmd+L) to task list items so users can view a task's notes/description without leaving Alfred. The task's `Notes` field from the Google Tasks API response is passed through to AwGo's `.Largetype()` method on each rendered item, with "(no details)" as fallback for tasks without notes.

## Technical Context

**Language/Version**: Go 1.22+
**Primary Dependencies**: AwGo (github.com/deanishe/awgo v0.29.1), Google Tasks API (google.golang.org/api/tasks/v1)
**Storage**: N/A (read-only display of existing API data)
**Testing**: `go test ./...`
**Target Platform**: macOS (Alfred 5+)
**Project Type**: Alfred workflow (Go binary + Alfred plist)
**Performance Goals**: No impact; no additional API calls or I/O
**Constraints**: Must not break existing task list rendering or action menu
**Scale/Scope**: 2 render functions modified, ~5 lines of code changed

## Constitution Check

Constitution is not configured (template only). No gates to check.

## Project Structure

### Documentation (this feature)

```text
specs/004-task-details-display/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
└── tasks.md             # Phase 2 output (via /speckit-tasks)
```

### Source Code (repository root)

```text
internal/
├── alfred/
│   ├── items.go         # RenderGroupedTasks - add .Largetype() [MODIFY]
│   └── workflow.go      # renderGroupedTasksWithWarnings - add .Largetype() [MODIFY]
└── tasks/
    └── list.go          # TaskItem struct - Notes already available via Task.Notes [NO CHANGE]
```

**Structure Decision**: No new files or directories needed. This feature only modifies two existing render functions to pass through data that is already available.

## Implementation Approach

### Change 1: `internal/alfred/items.go` - `RenderGroupedTasks`

In the loop that creates items for each task group, add a `.Largetype()` call after the existing `.Var()` calls. Determine the Large Type text: if `item.Task.Notes` is non-empty, use it; otherwise use `"(no details)"`.

### Change 2: `internal/alfred/workflow.go` - `renderGroupedTasksWithWarnings`

Same change as above in the parallel render function used for multi-account merged listings.

### Helper Pattern

Extract a small helper function `largetypeText(notes string) string` to avoid duplicating the empty-check logic across both render sites. Place it in `items.go` alongside the other rendering helpers.

### Test Approach

Add unit tests for `largetypeText` to verify:
- Non-empty notes are passed through unchanged
- Empty string returns "(no details)"
- Whitespace-only notes return "(no details)"

The AwGo `.Largetype()` method itself is a library call and does not need testing. Integration testing (pressing Cmd+L in Alfred) is manual.
