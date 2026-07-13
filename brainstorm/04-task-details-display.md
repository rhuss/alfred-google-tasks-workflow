# Brainstorm: Task Details Display

**Date:** 2026-07-13
**Status:** active

## Problem Framing

When browsing tasks in Alfred, users can see the task title, timeframe group, list name, and due date in the subtitle. However, there is no way to see the task's notes/description without opening Google Tasks in the browser. Users want a quick way to peek at task details without leaving Alfred.

## Approaches Considered

### A: Direct Largetype on task list items (Chosen)
Set `.Largetype(item.Task.Notes)` on each task item in the main list view. When the user presses Cmd+L on any task, Alfred displays the notes text in a large overlay. Use "(no details)" as fallback for tasks without notes.

- Pros: Minimal code change (one line per render site). No extra API calls since the List endpoint already returns the Notes field. No architecture change.
- Cons: Plain text only, no formatting. Empty notes need a fallback to avoid a blank overlay.

### B: Formatted fallback with title context
Same as A, but prefix the notes with the task title and metadata for context. Always shows something meaningful when Cmd+L is pressed.

- Pros: Better UX for tasks without notes.
- Cons: Slightly more logic. Adds information that's already visible in the subtitle.

### C: Quick Look via temp HTML file
Generate an HTML file in Alfred's cache directory with formatted notes content. Use `.Quicklookurl()` for a rich preview via Shift/Cmd+Y.

- Pros: Rich formatting possible (markdown rendering, styled text).
- Cons: Overkill for this requirement. Adds file I/O, cache management, and complexity. Better suited for a future native editor feature.

## Decision

Approach A: Direct Largetype. The simplest solution that solves the problem. The Google Tasks API already returns notes in the list response, so the data is available without additional API calls. A single `.Largetype()` call on each task item in the render functions is all that's needed.

## Key Requirements

- Set `.Largetype()` on all task items in `RenderGroupedTasks` (items.go)
- Set `.Largetype()` on all task items in `renderGroupedTasksWithWarnings` (workflow.go)
- Use "(no details)" fallback text when `item.Task.Notes` is empty
- Only applies to the main task list view, not the action sub-menu

## Open Questions

- Should the future native task editor (separate brainstorm) reuse any of this display logic, or will it be fully independent?
