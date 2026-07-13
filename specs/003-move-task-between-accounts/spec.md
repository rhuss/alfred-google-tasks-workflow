# Feature Specification: Move Task Between Accounts

**Feature Branch**: `003-move-task-between-accounts`
**Created**: 2026-07-13
**Status**: Draft
**Input**: Brainstorm #03: move-task-between-accounts

## User Scenarios & Testing

### User Story 1 - Move a task to another account (Priority: P1)

A user with multiple Google accounts (e.g., "work" and "personal") selects a task from the task list and sees "Move to {account}" entries in the action menu for each other configured account. Pressing Enter on a move entry recreates the task on the target account's same-named list and deletes the original.

**Why this priority**: Core feature. Without this, the user has no way to move tasks between accounts from the workflow.

**Independent Test**: Can be tested by selecting any task in multi-account mode and verifying the move entries appear and execute correctly.

**Acceptance Scenarios**:

1. **Given** a user with accounts "work" and "personal" selects a task on "work", **When** the action menu appears, **Then** it shows "Move to personal" (but not "Move to work").
2. **Given** a user selects "Move to personal" for a task in list "Shopping" on "work", **When** the move executes, **Then** the task is created in list "Shopping" on "personal" (auto-created if needed) and deleted from "work".
3. **Given** a user selects "Move to personal", **When** the move completes successfully, **Then** a notification shows "Task moved to personal".

---

### User Story 2 - Move preserves task properties (Priority: P1)

When a task is moved between accounts, its title, due date, and notes are preserved on the target account.

**Why this priority**: Without property preservation, the move loses important task context.

**Independent Test**: Move a task with a due date and notes, verify all fields are present on the target.

**Acceptance Scenarios**:

1. **Given** a task with title "Buy groceries", due date "2026-07-20", and notes "Get milk and eggs", **When** moved to another account, **Then** the new task has identical title, due date, and notes.
2. **Given** a task with no due date and no notes, **When** moved, **Then** only the title is set on the new task.

---

### User Story 3 - Error handling during move (Priority: P2)

If the move operation partially fails (create succeeds but delete fails), the user is warned about the potential duplicate.

**Why this priority**: Important for data integrity, but partial failures are rare in practice.

**Independent Test**: Simulate a delete failure after successful create and verify the warning notification.

**Acceptance Scenarios**:

1. **Given** the create on the target account succeeds but the delete on the source fails, **When** the move completes, **Then** a warning notification shows "Task moved but could not delete original. You may have a duplicate."
2. **Given** the create on the target account fails, **When** the move operation runs, **Then** an error notification shows "Failed to move task" and the original task is untouched.
3. **Given** the target account is not authenticated, **When** the action menu is rendered, **Then** the "Move to {account}" entry for that account is not shown (or shows as disabled with a hint).

---

### Edge Cases

- What happens when the user has only one account configured? No move entries are shown (FR-008).
- What happens when the target account is not authenticated? The move entry is hidden from the menu (FR-009).
- What happens when three or more accounts are configured? One move entry per other account is shown (e.g., for accounts A, B, C and a task on A, show "Move to B" and "Move to C").
- What happens when the task's source list name contains special characters? The list name is used as-is for lookup/creation on the target account.
- What happens when the task has subtasks? Only the top-level task is moved; subtasks remain on the source account (FR-010). No warning is shown for this case in v1.

## Requirements

### Functional Requirements

- **FR-001**: The action menu MUST show "Move to {account}" entries for each configured account except the task's current account, when in multi-account mode.
- **FR-002**: The move operation MUST create the task on the target account in a list with the same name as the source list, auto-creating the list if it does not exist (using `ResolveTaskList`).
- **FR-003**: The move operation MUST preserve the task's `Title`, `Due`, and `Notes` fields (as defined in the Google Tasks API `tasks.Task` struct). The move MUST use `InsertTask` directly (not `CreateTaskFromInput`) so all fields can be set explicitly.
- **FR-004**: After successful create on the target, the system MUST delete the original task from the source account.
- **FR-005**: The system MUST show a success notification after a successful move.
- **FR-006**: If create succeeds but delete fails, the system MUST show a warning about a potential duplicate.
- **FR-007**: If create fails, the system MUST show an error and leave the original task untouched.
- **FR-008**: Move entries MUST NOT appear in single-account mode.
- **FR-009**: Move entries MUST NOT appear for unauthenticated target accounts. The authentication check (`auth.TokenExists`) MUST happen at action menu render time so that unauthenticated accounts are excluded from the menu rather than failing at execution time.
- **FR-010**: Subtasks MUST NOT be moved; only top-level tasks are eligible for the move action.

### Implementation Details

- **Action argument format**: Move entries MUST use the argument format `move:{targetAccount}|{listID}:{taskID}` to encode the target account name alongside the standard action reference. The `executeAction` dispatcher MUST be extended to handle the `move:{targetAccount}` action prefix.
- **Task detail retrieval**: Before creating the task on the target account, the move handler MUST fetch the source task's full details via the Google Tasks API `Tasks.Get` call (using listID and taskID) to obtain the title, due date, and notes. The action menu only passes listID and taskID through Alfred item variables, not the full task data.
- **Source list name retrieval**: The move handler MUST also resolve the source list's name from the listID (via `Tasklists.Get` or by iterating `ListTaskLists`) so it can find or create the matching list on the target account.

### Key Entities

- **Task**: The Google Tasks item being moved (title, due date, notes, list membership).
- **Account**: A configured Google account with its own OAuth credentials and task lists.
- **Task List**: A named collection of tasks within an account. Lists may differ between accounts.

## Success Criteria

### Measurable Outcomes

- **SC-001**: Users can move a task between accounts in a single action (select task, choose "Move to {account}").
- **SC-002**: Task properties (title, due date, notes) are fully preserved after a move.
- **SC-003**: Partial failures (create succeeds, delete fails) result in a clear warning to the user.
- **SC-004**: All existing action menu functionality (Complete, Delete, Open) continues to work unchanged.

## Assumptions

- Subtasks are not moved in this version (only the top-level task is moved). This is tracked as FR-010.
- The Google Tasks API `Tasks.Insert` supports setting `Title`, `Due`, and `Notes` fields in a single call. This is confirmed by the existing `InsertTask` method in `client.go`.
- Users realistically have 2-3 accounts, so the action menu will not be overwhelmed with move entries.
- The target account's list name matching is case-insensitive, using the existing `FindTaskListByName` method which already normalizes with `strings.ToLower`.
