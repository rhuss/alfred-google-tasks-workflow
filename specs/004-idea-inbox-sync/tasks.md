# Tasks: Idea Inbox Sync

**Input**: Design documents from `specs/004-idea-inbox-sync/`
**Prerequisites**: plan.md (required), spec.md (required), research.md, data-model.md

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Phase 1: Setup

**Purpose**: Create the new ideas package structure

- [ ] T001 Create internal/ideas/ package directory

---

## Phase 2: Foundational (Inbox File I/O)

**Purpose**: Core inbox file operations that all user stories depend on

- [ ] T002 [P] Implement inbox file reader and TaskID extractor in internal/ideas/inbox.go (read file, parse `- TaskID:` lines into a set for dedup)
- [ ] T003 [P] Implement inbox file writer in internal/ideas/inbox.go (append entries with H3 heading, Date/Account/TaskID metadata, optional description; auto-create file with `# Idea Inbox` header if missing, including parent directories)
- [ ] T004 [P] Implement unit tests for inbox file operations in internal/ideas/inbox_test.go (test read/dedup, append, auto-create, existing content preservation)

**Checkpoint**: Inbox file I/O is independently testable with `go test ./internal/ideas/`

---

## Phase 3: User Story 1 - Automatic Idea Extraction (Priority: P1) 🎯 MVP

**Goal**: During `gt list`, fetch tasks from the Ideas list, write new ones to the Obsidian inbox, and delete synced tasks from Google Tasks.

**Independent Test**: Create a task in the Ideas list, run `gt list`, verify the task appears in the inbox file and is removed from Google Tasks.

### Implementation for User Story 1

- [ ] T005 [US1] Implement SyncIdeas function in internal/ideas/sync.go (accepts a tasks.Client, account name, list name, and inbox path; fetches Ideas list tasks, deduplicates against existing inbox entries, appends new entries, deletes synced tasks; returns count of synced ideas; all errors logged to stderr and swallowed)
- [ ] T006 [US1] Implement unit tests for SyncIdeas in internal/ideas/sync_test.go (test sync with new tasks, dedup skip, delete-after-write, missing list skip, error handling)
- [ ] T007 [US1] Add syncIdeasToInbox method to Workflow in internal/alfred/workflow.go (read IDEA_INBOX_PATH and IDEA_LIST_NAME from env, early-return if either unset, create client for current account, call ideas.SyncIdeas, wrap in error recovery)
- [ ] T008 [US1] Wire syncIdeasToInbox into handleList in internal/alfred/workflow.go (call before rendering, after auth check, for single-account mode)

**Checkpoint**: Single-account idea sync works end-to-end

---

## Phase 4: User Story 2 - Multi-Account Idea Collection (Priority: P1)

**Goal**: Scan Ideas list on ALL authenticated accounts during every list operation, regardless of which account is targeted.

**Independent Test**: Add ideas to two accounts, run `gt list`, verify both appear in inbox with correct account tags.

### Implementation for User Story 2

- [ ] T009 [US2] Implement syncIdeasAllAccounts method on Workflow in internal/alfred/workflow.go (iterate all accounts from AccountConfig, create client for each authenticated account, call ideas.SyncIdeas for each, skip unauthenticated accounts silently)
- [ ] T010 [US2] Wire syncIdeasAllAccounts into handleList and handleListAllAccounts in internal/alfred/workflow.go (replace single-account sync call; in multi-account mode always sync all accounts; in single-account mode fall back to single sync from T007)

**Checkpoint**: Multi-account idea sync works, ideas from all accounts appear in the same inbox file

---

## Phase 5: User Story 3 - Feature Disabled When Not Configured (Priority: P2)

**Goal**: Zero overhead when IDEA_INBOX_PATH or IDEA_LIST_NAME is not set.

**Independent Test**: Unset IDEA_INBOX_PATH, run `gt list`, verify no file operations and no errors.

### Implementation for User Story 3

- [ ] T011 [US3] Add guard clause tests in internal/ideas/sync_test.go (verify SyncIdeas returns immediately when inbox path is empty; verify no API calls or file I/O occur)
- [ ] T012 [US3] Add workflow-level guard tests in internal/alfred/workflow_test.go (verify syncIdeasToInbox is a no-op when env vars are unset)

**Checkpoint**: Feature is fully opt-in with zero impact when disabled

---

## Phase 6: User Story 4 - Inbox File Auto-Creation (Priority: P3)

**Goal**: Auto-create the inbox file with header when it doesn't exist.

**Independent Test**: Set IDEA_INBOX_PATH to non-existent path, create an idea, run `gt list`, verify file is created.

### Implementation for User Story 4

- [ ] T013 [US4] Add auto-creation tests in internal/ideas/inbox_test.go (test file creation with header, parent directory creation, append after auto-create)

**Checkpoint**: Auto-creation is already implemented in T003; this phase adds test coverage for the edge cases

---

## Phase 7: Polish & Cross-Cutting Concerns

**Purpose**: Documentation and final validation

- [ ] T014 Update README.md with Idea Inbox Sync configuration section (IDEA_INBOX_PATH and IDEA_LIST_NAME variables, example setup, Gemini voice command tip)

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies
- **Foundational (Phase 2)**: Depends on Phase 1
- **User Story 1 (Phase 3)**: Depends on Phase 2 (inbox I/O)
- **User Story 2 (Phase 4)**: Depends on Phase 3 (single-account sync)
- **User Story 3 (Phase 5)**: Can run after Phase 3 (tests only)
- **User Story 4 (Phase 6)**: Can run after Phase 2 (tests only, auto-create already in T003)
- **Polish (Phase 7)**: Depends on all user stories

### Parallel Opportunities

- T002, T003, T004 can run in parallel (different functions in same package, no dependencies)
- T011 and T013 can run in parallel (independent test files)

### Within Each User Story

- Implementation before integration into workflow
- Unit tests alongside implementation
- Commit after each task

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup (T001)
2. Complete Phase 2: Foundational inbox I/O (T002-T004)
3. Complete Phase 3: Single-account sync (T005-T008)
4. **STOP and VALIDATE**: Test with single account
5. Proceed to multi-account

### Incremental Delivery

1. Setup + Foundational -> inbox I/O works
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
