# Implementation Plan: Move Task Between Accounts

**Branch**: `003-move-task-between-accounts` | **Date**: 2026-07-13 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `specs/003-move-task-between-accounts/spec.md`

## Summary

Add "Move to {account}" entries to the task action menu in multi-account mode. The move operation creates the task on the target account (same-named list, auto-created) and deletes the original. Requires extending the action menu rendering, adding a `GetTask` method to the client, adding a `MoveTask` method, and handling the new `move:{targetAccount}` action format in the dispatcher.

## Technical Context

**Language/Version**: Go 1.22+
**Primary Dependencies**: deanishe/awgo (Alfred workflow library), google.golang.org/api/tasks/v1
**Storage**: N/A (uses Google Tasks API as backend)
**Testing**: `go test ./...`
**Target Platform**: macOS (Alfred workflow)
**Project Type**: CLI (Alfred workflow binary)
**Performance Goals**: N/A (single-user, interactive)
**Constraints**: Google Tasks API rate limits (standard quotas)
**Scale/Scope**: 2-3 accounts, <100 tasks typically

## Constitution Check

*GATE: Constitution is not configured (template only). No gates to enforce.*

## Project Structure

### Documentation (this feature)

```text
specs/003-move-task-between-accounts/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
└── tasks.md             # Phase 2 output (via /speckit-tasks)
```

### Source Code (repository root)

```text
cmd/
└── main.go              # Entry point (no changes needed)

internal/
├── alfred/
│   ├── items.go         # RenderActionMenu (add move entries)
│   └── workflow.go      # executeAction (add move: case)
├── auth/
│   └── accounts.go      # AccountConfig, ResolveAccount (existing, used as-is)
└── tasks/
    ├── actions.go       # Add MoveTask method
    └── client.go        # Add GetTask, GetTaskList methods
```

**Structure Decision**: Existing Go project structure with `internal/` packages. Changes touch 3 files (items.go, workflow.go, client.go) and add logic to 1 file (actions.go).

## Implementation Approach

### 1. Add GetTask and GetTaskList to client.go

The move handler needs to fetch the full task details (title, due, notes) and the source list name. Add two new methods to `Client`:

- `GetTask(listID, taskID string) (*tasks.Task, error)`: Calls `Tasks.Get(listID, taskID).Do()`
- `GetTaskList(listID string) (*tasks.TaskList, error)`: Calls `Tasklists.Get(listID).Do()`

### 2. Add MoveTask to actions.go

New method on `Client` that orchestrates the move:

```
MoveTask(sourceListID, taskID string, targetClient *Client, targetListName string) (*tasks.Task, error)
```

Steps:
1. Fetch source task via `GetTask(sourceListID, taskID)`
2. Resolve target list via `targetClient.ResolveTaskList(targetListName)` (auto-creates if needed)
3. Create new task via `targetClient.InsertTask(targetListID, newTask)` with Title, Due, Notes copied
4. Delete source task via `DeleteTask(sourceListID, taskID)`
5. If create succeeds but delete fails, return the created task and a `PartialMoveError` (wrapping the delete error) so the caller can show a warning while still reporting success

### 3. Extend RenderActionMenu in items.go

When `AccountConfig` is non-nil and has 2+ accounts:
1. Get the current task's account name from the `accountName` parameter
2. Iterate `AccountConfig.AccountNames()`
3. For each account != current account, check `auth.TokenExists` for the target
4. Add "Move to {account}" item with arg format `move:{targetAccount}|{listID}:{taskID}`
5. Set `accountName` var on the item (source account for the move handler)

### 4. Handle move action in workflow.go executeAction

Add a `move:` prefix case in the action dispatcher:
1. Parse `move:{targetAccount}` from the action string
2. Create source client (current account context)
3. Resolve target account via `auth.ResolveAccount`
4. Create target client
5. Fetch source list name via `GetTaskList`
6. Call `MoveTask` on the source client
7. Show appropriate notification (success, warning for partial failure, error)

## Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| Create succeeds, delete fails | Low | Medium | Warn user about potential duplicate (FR-006) |
| Target account token expired | Low | Low | Token auto-refresh via existing `EnsureValidToken` |
| Rate limiting on rapid moves | Very Low | Low | Standard API quotas are generous for single-user |
