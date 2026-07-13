# Tasks: Task Details Display

**Input**: Design documents from `/specs/004-task-details-display/`
**Prerequisites**: plan.md (required), spec.md (required for user stories)

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2)
- Include exact file paths in descriptions

## Phase 1: Setup

**Purpose**: Add helper function for Large Type text resolution

- [x] T001 Add `largetypeText` helper function in internal/alfred/items.go that returns notes text or "(no details)" fallback for empty/whitespace-only input

---

## Phase 2: User Story 1 - View task notes via Large Type (Priority: P1)

**Goal**: Users can press Cmd+L on any task in the single-account list view to see task notes in Large Type

**Independent Test**: List tasks with `gt list`, press Cmd+L on a task with notes, verify notes appear. Press Cmd+L on a task without notes, verify "(no details)" appears.

### Tests for User Story 1

- [x] T002 [US1] Add unit tests for `largetypeText` in internal/alfred/items_test.go covering: non-empty notes passed through, empty string returns "(no details)", whitespace-only returns "(no details)"

### Implementation for User Story 1

- [x] T003 [US1] Add `.Largetype(largetypeText(item.Task.Notes))` call to each task item in `RenderGroupedTasks` in internal/alfred/items.go

**Checkpoint**: Single-account task listing now supports Large Type for notes

---

## Phase 3: User Story 2 - View task notes in multi-account merged listing (Priority: P1)

**Goal**: Users can press Cmd+L on any task in the merged multi-account list view to see task notes in Large Type

**Independent Test**: Configure two accounts with list_default "all", list tasks, press Cmd+L on tasks from each account, verify correct notes appear.

### Implementation for User Story 2

- [x] T004 [US2] Add `.Largetype(largetypeText(item.Task.Notes))` call to each task item in `renderGroupedTasksWithWarnings` in internal/alfred/workflow.go

**Checkpoint**: Both single-account and multi-account task listings support Large Type for notes

---

## Phase 4: Polish & Cross-Cutting Concerns

**Purpose**: Verify all tests pass and feature is complete

- [x] T005 Run `go test ./...` to verify all tests pass
- [x] T006 Run `go build ./...` to verify clean compilation

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 1 (Setup)**: No dependencies, start immediately
- **Phase 2 (US1)**: Depends on T001 (helper function)
- **Phase 3 (US2)**: Depends on T001 (helper function), can run in parallel with Phase 2
- **Phase 4 (Polish)**: Depends on Phases 2 and 3

### User Story Dependencies

- **User Story 1 (P1)**: Depends only on T001 (helper function)
- **User Story 2 (P1)**: Depends only on T001 (helper function), independent of US1

### Parallel Opportunities

- T003 and T004 modify different files (items.go vs workflow.go) and can run in parallel after T001

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete T001: Helper function
2. Complete T002: Tests
3. Complete T003: Single-account Large Type
4. **STOP and VALIDATE**: Test Cmd+L in Alfred
5. Continue to User Story 2

### Incremental Delivery

1. T001 (helper) -> T002 (tests) -> T003 (single-account) -> Validate
2. T004 (multi-account) -> Validate
3. T005-T006 (polish) -> Done

---

## Notes

- Total tasks: 6
- User Story 1: 2 tasks (T002, T003)
- User Story 2: 1 task (T004)
- Setup: 1 task (T001)
- Polish: 2 tasks (T005, T006)
- Parallel opportunities: T003 and T004 can run concurrently
- This is a minimal feature with no new dependencies, API calls, or data model changes
