# Tasks: Move Task Between Accounts

**Input**: Design documents from `specs/003-move-task-between-accounts/`
**Prerequisites**: plan.md (required), spec.md (required), research.md, data-model.md

**Organization**: Tasks are grouped by user story to enable independent implementation and testing.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

---

## Phase 1: Foundational (API Methods)

**Purpose**: Add the Google Tasks API methods that the move feature depends on

- [x] T001 [P] Add `GetTask(listID, taskID string) (*tasks.Task, error)` method to internal/tasks/client.go that fetches a single task's full details via `Tasks.Get`
- [x] T002 [P] Add `GetTaskList(listID string) (*tasks.TaskList, error)` method to internal/tasks/client.go that fetches a task list's metadata (needed for list name resolution)

**Checkpoint**: New API methods available, existing tests still pass

---

## Phase 2: User Story 1 - Move a task to another account (Priority: P1)

**Goal**: User selects a task, sees "Move to {account}" in the action menu, and the task is moved to the target account's same-named list.

**Independent Test**: Select a task in multi-account mode, choose "Move to {account}", verify the task appears on the target account and is deleted from the source.

### Implementation for User Story 1

- [x] T003 [US1] Add `MoveTask(sourceListID, taskID string, targetClient *Client, targetListName string) (*tasks.Task, error)` method to internal/tasks/actions.go. Must: fetch source task via GetTask, resolve target list via ResolveTaskList (auto-create), insert task with Title/Due/Notes on target, delete from source. Define a `PartialMoveError` struct in internal/tasks/actions.go (fields: `CreatedTask *tasks.Task`, `DeleteErr error`) implementing the `error` interface. Return `PartialMoveError` when create succeeds but delete fails, so the caller can use `errors.As` to differentiate partial from full failure.
- [x] T004 [US1] Extend `RenderActionMenu` in internal/alfred/items.go to accept `accountConfig *auth.AccountConfig` as a fourth parameter (new signature: `RenderActionMenu(listID, taskID, accountName string, accountConfig *auth.AccountConfig)`). When `accountConfig` is non-nil and has 2+ accounts, iterate `accountConfig.AccountNames()`, skip the current `accountName`, resolve each target via `auth.ResolveAccount(accountConfig, targetName)`, check `auth.TokenExists(targetCtx.DataDir)`, and add a "Move to {targetName}" item with arg format `move:{targetName}|{listID}:{taskID}` and icon `icons/move.png`. Move entries only appear for tasks already shown in the list, which are always top-level tasks (FR-010 is satisfied implicitly).
- [x] T005 [US1] Update call site of `RenderActionMenu` in internal/alfred/workflow.go `handleAction` to pass the AccountConfig.
- [x] T006 [US1+US3] Add `move:{targetAccount}` case to `executeAction` in internal/alfred/workflow.go. Must: parse target account name from the action string (prefix before `|`), create source client from `w.AccountCtx`, resolve target account via `auth.ResolveAccount(w.AccountConfig, targetAccount)`, create target client via `tasks.NewClient`, fetch source list name via source client's `GetTaskList(listID)`, call source client's `MoveTask(listID, taskID, targetClient, listName)`. Handle all three outcomes: (1) success: show notification "Task moved to {targetAccount}"; (2) `PartialMoveError` (use `errors.As`): show warning "Task moved but could not delete original. You may have a duplicate."; (3) other error: show error "Failed to move task: {error}".

**Checkpoint**: Move feature fully functional for the main flow

---

## Phase 3: User Story 2 - Move preserves task properties (Priority: P1)

**Goal**: Task title, due date, and notes are preserved during the move.

*This is handled by the MoveTask implementation in T003. No additional tasks needed since T003 explicitly copies Title, Due, and Notes fields.*

**Checkpoint**: Covered by T003 implementation

---

## Phase 4: Tests & Verification

- [x] T007 Add unit tests for `GetTask`, `GetTaskList`, and `MoveTask` (including `PartialMoveError` path) in internal/tasks/client_test.go and internal/tasks/actions_test.go. Follow existing test patterns (table-driven tests with mock HTTP transport). After writing tests, run `go test ./...` and `go vet ./...` to verify all tests pass and no warnings exist.

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 1 (Foundational)**: No dependencies, can start immediately
- **Phase 2 (US1+US3)**: Depends on Phase 1 (T001, T002 must complete before T003)
- **Phase 3 (US2)**: Covered by Phase 2 (no separate tasks)
- **Phase 4 (Tests)**: Depends on all prior phases

### Task Dependencies

```
T001, T002 (parallel) → T003 → T004, T005 (parallel) → T006 → T007
```

### Parallel Opportunities

- T001 and T002 can run in parallel (different methods, same file but no conflicts)
- T004 and T005 can run in parallel (different files)

---

## Implementation Strategy

### MVP (User Stories 1 + 3)

1. Complete Phase 1: Add GetTask and GetTaskList
2. Complete Phase 2: Implement MoveTask (with PartialMoveError), extend action menu, wire up dispatcher with full error handling
3. **STOP and VALIDATE**: Test move between two accounts manually, including error paths

### Full Feature

4. Complete Phase 4: Unit tests and verification
