# Tasks: Idea Inbox Sync

**Input**: Design documents from `specs/004-idea-inbox-sync/`
**Prerequisites**: plan.md (required), spec.md (required), research.md, data-model.md

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Global Constraints

These apply to ALL tasks:
- **FR-009**: Never block, delay, or interfere with the task listing display. All sync errors logged to stderr and swallowed.
- **SC-003**: Idea sync adds less than 500ms of latency to list operations when the Ideas list is empty or does not exist.

## Interfaces

```go
// internal/ideas/inbox.go
func ReadSyncedTaskIDs(inboxPath string) (map[string]bool, error)
func AppendIdeaEntry(inboxPath string, entry IdeaEntry) error

type IdeaEntry struct {
    Title       string
    Date        string // YYYY-MM-DD, derived from task.Updated (RFC 3339)
    Account     string // empty in single-account mode
    TaskID      string
    Description string // from task.Notes, may be empty
}

// internal/ideas/sync.go
func SyncIdeas(client *tasks.Client, accountName string, listName string, inboxPath string) (int, error)
```

---

## Phase 1: Foundational (Inbox File I/O)

**Purpose**: Core inbox file operations that all user stories depend on

- [ ] T001 [P] Implement IdeaEntry struct, ReadSyncedTaskIDs, and AppendIdeaEntry in internal/ideas/inbox.go (ReadSyncedTaskIDs reads file and parses `- TaskID:` lines into a map for dedup; AppendIdeaEntry appends an entry with H3 heading, Date/Account/TaskID metadata as bullet-list fields, optional description from task notes; auto-create file with `# Idea Inbox` header if missing including parent directories; omit Account field when empty; Date is derived from task.Updated parsed as RFC 3339 and formatted as YYYY-MM-DD per FR-012)
- [ ] T002 [P] Implement unit tests for inbox file operations in internal/ideas/inbox_test.go (test ReadSyncedTaskIDs with existing entries, empty file, and missing file; test AppendIdeaEntry with and without description, with and without Account, auto-create with header, parent directory creation, existing content preservation)

**Checkpoint**: Inbox file I/O is independently testable with `go test ./internal/ideas/`

---

## Phase 2: User Story 1 - Automatic Idea Extraction (Priority: P1) 🎯 MVP

**Goal**: During `gt list`, fetch tasks from the Ideas list, write new ones to the Obsidian inbox, and delete synced tasks from Google Tasks.

**Independent Test**: Create a task in the Ideas list, run `gt list`, verify the task appears in the inbox file and is removed from Google Tasks.

### Implementation for User Story 1

- [ ] T003 [US1] Implement SyncIdeas function in internal/ideas/sync.go (signature: `func SyncIdeas(client *tasks.Client, accountName, listName, inboxPath string) (int, error)`; uses client.FindTaskListByName to find the Ideas list, returns 0 if not found; calls client.ListTasks, calls ReadSyncedTaskIDs for dedup, builds IdeaEntry from each task with Date from task.Updated, calls AppendIdeaEntry for new entries, calls client.DeleteTask after successful write; returns count of synced ideas)
- [ ] T004 [US1] Implement unit tests for SyncIdeas in internal/ideas/sync_test.go (test sync with new tasks, dedup skip, delete-after-write ordering, missing list returns 0, API error handling)
- [ ] T005 [US1] Add syncIdeasToInbox method to Workflow in internal/alfred/workflow.go (read IDEA_INBOX_PATH and IDEA_LIST_NAME from os.Getenv, early-return if either is empty; create tasks.Client using getAuthenticatedClient; call ideas.SyncIdeas(client, accountCtx.Name, listName, inboxPath); wrap entire method in deferred recover to swallow panics; log errors to stderr via fmt.Fprintf)
- [ ] T006 [US1] Wire syncIdeasToInbox into handleList in internal/alfred/workflow.go (call syncIdeasToInbox after requireAuth check but before fetchTasksForCurrentAccount, for single-account mode only)

**Checkpoint**: Single-account idea sync works end-to-end

---

## Phase 3: User Story 2 - Multi-Account Idea Collection (Priority: P1)

**Goal**: Scan Ideas list on ALL authenticated accounts during every list operation, regardless of which account is targeted.

**Independent Test**: Add ideas to two accounts, run `gt list`, verify both appear in inbox with correct account tags.

### Implementation for User Story 2

- [ ] T007 [US2] Implement syncIdeasAllAccounts method on Workflow in internal/alfred/workflow.go (read IDEA_INBOX_PATH and IDEA_LIST_NAME from env, early-return if either unset; iterate AccountConfig.AccountNames(); for each: call auth.ResolveAccount, check auth.TokenExists, load credentials via auth.LoadClientCredentialsFrom and auth.EnsureValidToken, create tasks.Client, call ideas.SyncIdeas with the account name; skip unauthenticated or erroring accounts silently; wrap in deferred recover)
- [ ] T008 [US2] Wire syncIdeasAllAccounts into handleList and handleListAllAccounts in internal/alfred/workflow.go (in multi-account mode: call syncIdeasAllAccounts before any task fetching in both handleList and handleListAllAccounts; in single-account mode: keep syncIdeasToInbox from T005; the sync always runs for all accounts regardless of @account targeting or list_default setting)

**Checkpoint**: Multi-account idea sync works, ideas from all accounts appear in the same inbox file

---

## Phase 4: User Story 3 - Feature Disabled When Not Configured (Priority: P2)

**Goal**: Zero overhead when IDEA_INBOX_PATH or IDEA_LIST_NAME is not set.

**Independent Test**: Unset IDEA_INBOX_PATH, run `gt list`, verify no file operations and no errors.

### Implementation for User Story 3

- [ ] T009 [US3] Add guard clause tests in internal/ideas/sync_test.go (verify SyncIdeas returns 0 immediately when inboxPath is empty; verify no API calls or file I/O occur when called with empty path)
- [ ] T010 [US3] Add workflow-level guard tests in internal/alfred/workflow_test.go (verify syncIdeasToInbox is a no-op when IDEA_INBOX_PATH env var is unset; verify syncIdeasAllAccounts is a no-op when env vars are unset)

**Checkpoint**: Feature is fully opt-in with zero impact when disabled

---

## Phase 5: User Story 4 - Inbox File Auto-Creation (Priority: P3)

**Goal**: Auto-create the inbox file with header when it doesn't exist.

**Independent Test**: Set IDEA_INBOX_PATH to non-existent path, create an idea, run `gt list`, verify file is created.

### Implementation for User Story 4

- [ ] T011 [US4] Add auto-creation edge case tests in internal/ideas/inbox_test.go (test file creation when parent directory is deeply nested, test that header is `# Idea Inbox\n\n`, test that subsequent appends preserve the header)

**Checkpoint**: Auto-creation is implemented in T001; this phase adds edge case test coverage

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Documentation and final validation

- [ ] T012 Update README.md with Idea Inbox Sync configuration section (IDEA_INBOX_PATH and IDEA_LIST_NAME variables, example setup, Gemini voice command tip for Android)

---

## Dependencies & Execution Order

### Phase Dependencies

- **Foundational (Phase 1)**: No dependencies
- **User Story 1 (Phase 2)**: Depends on Phase 1 (inbox I/O)
- **User Story 2 (Phase 3)**: Depends on Phase 2 (single-account sync)
- **User Story 3 (Phase 4)**: Can run after Phase 2 (tests only)
- **User Story 4 (Phase 5)**: Can run after Phase 1 (tests only, auto-create already in T001)
- **Polish (Phase 6)**: Depends on all user stories

### Parallel Opportunities

- T001 and T002 can run in parallel (implementation and tests in same package, different files)
- T009 and T011 can run in parallel (independent test additions)

### Within Each User Story

- Implementation before integration into workflow
- Unit tests alongside implementation
- Commit after each task

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Foundational inbox I/O (T001-T002)
2. Complete Phase 2: Single-account sync (T003-T006)
3. **STOP and VALIDATE**: Test with single account
4. Proceed to multi-account

### Incremental Delivery

1. Foundational -> inbox I/O works
2. Add US1 -> single-account sync works (MVP)
3. Add US2 -> multi-account sync works
4. Add US3 -> guard clause tests confirm opt-in behavior
5. Add US4 -> auto-creation edge case tests
6. Polish -> README updated

---

## Notes

- [P] tasks = different files, no dependencies
- All sync errors must be silently swallowed (FR-009)
- The `tasks.Client` already provides FindTaskListByName, ListTasks, and DeleteTask
- Inbox file format uses H3 headings with bullet-list metadata (Date, Account, TaskID)
- Account field is omitted in single-account mode (FR-005)
- Date field derived from task.Updated (RFC 3339) formatted as YYYY-MM-DD (FR-012)
