# Implementation Plan: Idea Inbox Sync

**Branch**: `004-idea-inbox-sync` | **Date**: 2026-07-13 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `specs/004-idea-inbox-sync/spec.md`

## Summary

Add automatic synchronization of tasks from a configurable Google Tasks "Ideas" list to an Obsidian markdown inbox file. The sync runs as a silent side effect during every `gt list` operation, scanning all authenticated accounts, appending new ideas (deduplicated by TaskID), and deleting synced tasks from Google Tasks. The feature is opt-in via two Alfred workflow variables (IDEA_INBOX_PATH and IDEA_LIST_NAME) and disabled when either is unset.

## Technical Context

**Language/Version**: Go 1.26.4
**Primary Dependencies**: awgo (Alfred workflow), google.golang.org/api/tasks/v1, golang.org/x/oauth2
**Storage**: Local filesystem (Obsidian markdown file)
**Testing**: `go test ./...` with table-driven tests
**Target Platform**: macOS (Alfred workflow)
**Project Type**: CLI / Alfred workflow
**Performance Goals**: Idea sync adds less than 500ms latency to list operations
**Constraints**: Never block or interfere with task listing display
**Scale/Scope**: Single user, handful of ideas per sync cycle

## Constitution Check

No project constitution defined. No gates to check.

## Project Structure

### Documentation (this feature)

```text
specs/004-idea-inbox-sync/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── spec.md              # Feature specification
├── checklists/          # Quality checklists
└── tasks.md             # Phase 2 output (created by /speckit-tasks)
```

### Source Code (repository root)

```text
cmd/
└── main.go                       # Entry point (unchanged)

internal/
├── alfred/
│   └── workflow.go               # MODIFIED: add syncIdeasToInbox calls
├── ideas/
│   ├── sync.go                   # NEW: core sync logic (fetch, dedup, append, delete)
│   ├── sync_test.go              # NEW: unit tests for sync logic
│   ├── inbox.go                  # NEW: inbox file I/O (read existing IDs, append entries, create file)
│   └── inbox_test.go             # NEW: unit tests for inbox file operations
└── tasks/
    └── client.go                 # UNCHANGED (reuse existing API methods)
```

**Structure Decision**: New `internal/ideas/` package encapsulates all idea sync logic. This keeps the `tasks` package focused on Google Tasks API operations and the `alfred` package focused on Alfred UI rendering. The workflow.go file is the integration point that wires the two together.

## Design Decisions

### Package Separation

The sync logic lives in `internal/ideas/` rather than in `internal/tasks/` or `internal/alfred/` because:
- It crosses concerns: file I/O (Obsidian) + API calls (Google Tasks) + dedup logic
- The `tasks` package is a thin API client wrapper and should stay that way
- The `alfred` package is for UI rendering, not business logic
- A dedicated package makes the sync logic independently testable

### Sync Entry Point

A new `syncIdeasToInbox` method on `Workflow` is called from both `handleList` and `handleListAllAccounts`. It:
1. Reads config from env vars (IDEA_INBOX_PATH, IDEA_LIST_NAME)
2. Early-returns if either is unset
3. Iterates ALL accounts (regardless of current targeting)
4. For each account: finds the Ideas list, fetches tasks, deduplicates, appends, deletes
5. All errors are logged to stderr and swallowed

### Dedup Strategy

Read the inbox file once at the start of sync, extract all `- TaskID: xxx` values into a set. Check each task's ID against this set before appending. This is O(n) where n is the number of existing entries, which is negligible for the expected scale.
