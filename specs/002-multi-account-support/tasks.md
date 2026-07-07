# Tasks: Multi-Account Support

**Input**: Design documents from `specs/002-multi-account-support/`
**Prerequisites**: plan.md (required), spec.md (required), research.md, data-model.md

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Phase 1: Setup

**Purpose**: Project initialization for multi-account support

- [X] T001 Create account configuration types (AccountConfig, Account, AccountContext) in `internal/auth/accounts.go`

  **Interfaces (produced by this task, consumed by T002-T009+):**
  ```go
  type Account struct {
      Name         string // symbolic name (e.g., "personal", "work")
      Credentials  string // path to client_secret.json (relative to workflow data dir)
      TokenDir     string // directory for OAuth tokens (relative to workflow data dir)
      Keyword      string // optional Alfred keyword
      ProfileIndex int    // optional Google multi-login index
  }

  type AccountConfig struct {
      Default     string              // name of default account
      ListDefault string              // "default" or "all"
      Accounts    map[string]Account  // name -> Account
  }

  type AccountContext struct {
      Name            string // resolved account name
      DataDir         string // absolute path to token storage directory
      CredentialsPath string // absolute path to credentials file
      ProfileIndex    int    // Google authuser index
  }
  ```

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core account infrastructure that MUST be complete before any user story

- [X] T002 Implement `LoadAccountConfig(dataDir string)` to parse and validate `accounts.json` in `internal/auth/accounts.go`

  **Interfaces**: Consumes `AccountConfig`, `Account` from T001. Produces `func LoadAccountConfig(dataDir string) (*AccountConfig, error)`.

- [X] T003 Implement `DefaultContext(dataDir string)` to create an implicit single-account AccountContext when no `accounts.json` exists in `internal/auth/accounts.go`

  **Interfaces**: Consumes `AccountContext` from T001. Produces `func DefaultContext(dataDir string) *AccountContext`. Returns an AccountContext where `DataDir` = dataDir, `CredentialsPath` = filepath.Join(dataDir, "client_secret.json").

- [X] T004 Implement `ResolveAccount(config *AccountConfig, name string)` to resolve an account name to AccountContext in `internal/auth/accounts.go`

  **Interfaces**: Consumes `AccountConfig`, `AccountContext` from T001. Produces `func ResolveAccount(config *AccountConfig, name string) (*AccountContext, error)`.

- [X] T005 [P] Add `LoadClientCredentialsFrom(path string)` variant that loads credentials from an explicit path (for multi-account) in `internal/auth/credentials.go`

  **Interfaces**: Produces `func LoadClientCredentialsFrom(path string) (*oauth2.Config, error)`.

- [X] T006 [P] Write unit tests for account config loading, validation, and resolution in `internal/auth/accounts_test.go`

**Checkpoint**: Account configuration layer ready, user story implementation can begin

---

## Phase 3: User Story 1 - Single-Account Backward Compatibility (Priority: P1)

**Goal**: Existing single-account users experience zero behavior changes after upgrading

**Independent Test**: Run all commands without `accounts.json` and verify identical behavior to current version

- [X] T007 [US1] Refactor `Workflow` struct to hold an `AccountContext` field instead of using `DataDir` directly in `internal/alfred/workflow.go`

  **Interfaces**: Consumes `AccountContext` from T001. The `Workflow` struct becomes:
  ```go
  type Workflow struct {
      WF         *aw.Workflow
      AccountCtx *auth.AccountContext // replaces DataDir field
  }
  ```
- [X] T008 [US1] Update `NewWorkflow()` to call `DefaultContext()` when no `accounts.json` exists, preserving existing DataDir/credentials paths in `internal/alfred/workflow.go`
- [X] T009 [US1] Update all command handlers (`handleLogin`, `handleLogout`, `handleAdd`, `handleList`, `handleOpen`, `handleAction`) to use `AccountContext.DataDir` instead of `w.DataDir` in `internal/alfred/workflow.go`
- [X] T010 [US1] Verify `@` prefix in input is NOT parsed as account selector when `accounts.json` is absent in `internal/alfred/workflow.go`

---

## Phase 4: User Story 2 - Account Configuration Setup (Priority: P1)

**Goal**: Users can create `accounts.json` and the workflow recognizes multiple accounts

**Independent Test**: Create `accounts.json` with two accounts and run any command to verify config loads correctly

- [X] T011 [US2] Update `NewWorkflow()` to detect `accounts.json` and load `AccountConfig` when present in `internal/alfred/workflow.go`
- [X] T012 [US2] Add config validation error display as Alfred error items (invalid JSON, missing credentials, bad default) in `internal/alfred/workflow.go`
- [X] T013 [P] [US2] Write integration test for config loading with valid and invalid `accounts.json` files in `internal/auth/accounts_test.go`

---

## Phase 5: User Story 3 - Account-Targeted Task Creation with @ Prefix (Priority: P1)

**Goal**: Users can target specific accounts with `@accountname` prefix for task creation

**Independent Test**: Create tasks with `@account` prefix and verify they appear in the correct Google account

- [X] T014 [US3] Implement `@accountname` prefix extraction in `handleFilter()` as a pre-processing step before command routing in `internal/alfred/workflow.go`
- [X] T015 [US3] Strip `@accountname` from input before passing to command handlers (ensure `input.Parse()` receives clean input) in `internal/alfred/workflow.go`
- [X] T016 [US3] Add error handling for unknown account names (display valid account names as Alfred items) in `internal/alfred/workflow.go`
- [X] T017 [P] [US3] Write unit tests for `@` prefix extraction (valid accounts, unknown accounts, no prefix, single-account mode) in `internal/alfred/workflow_test.go`

---

## Phase 6: User Story 4 - Multi-Account Task Listing (Priority: P1)

**Goal**: Users can view tasks across all accounts in a merged list or target specific accounts

**Independent Test**: Create tasks in multiple accounts and verify listing with `list_default` settings

- [X] T018 [US4] Add `AccountName` field to `TaskItem` struct in `internal/tasks/list.go`

  **Interfaces (produced, consumed by T019-T022):**
  ```go
  type TaskItem struct {
      Task        *taskapi.Task
      ListName    string
      ListID      string
      AccountName string // NEW: empty in single-account mode, set in multi-account
  }
  ```
- [X] T019 [US4] Implement merged listing in `handleList()`: iterate accounts, fetch tasks per account, tag with account name, merge results in `internal/alfred/workflow.go`
- [X] T020 [US4] Update `RenderGroupedTasks()` to include account name in subtitle when in multi-account mode (e.g., "Work list (personal)") in `internal/alfred/items.go`
- [X] T021 [US4] Handle per-account auth failures gracefully: show tasks from authenticated accounts, append warning item for failed accounts in `internal/alfred/workflow.go`
- [X] T022 [P] [US4] Update list test expectations for account name in TaskItem in `internal/tasks/list_test.go`

---

## Phase 7: User Story 5 - Per-Account Authentication (Priority: P1)

**Goal**: Each account authenticates independently with its own credentials and token storage

**Independent Test**: Log in to each account separately and verify tokens stored in correct subdirectories

- [X] T023 [US5] Update `handleLogin()` to use `AccountContext.CredentialsPath` for loading credentials and `AccountContext.DataDir` for token storage in `internal/alfred/workflow.go`
- [X] T024 [US5] Update `handleLogout()` to delete only the targeted account's token using `AccountContext.DataDir` in `internal/alfred/workflow.go`
- [X] T025 [US5] Update `requireAuth()` to check auth status using `AccountContext.DataDir` and show account-specific login message (e.g., "Run 'gt @work login'") in `internal/alfred/workflow.go`
- [X] T026 [US5] Update `getAuthenticatedClient()` to use `AccountContext.CredentialsPath` and `AccountContext.DataDir` in `internal/alfred/workflow.go`

---

## Phase 8: User Story 6 - Account-Targeted Browser Opening (Priority: P2)

**Goal**: Users can open Google Tasks in the browser for a specific account's profile

**Independent Test**: Run `gt @work open` and verify browser URL includes correct `authuser` parameter

- [X] T027 [US6] Update `handleOpen()` and `OpenGoogleTasks()` to accept an `authuser` parameter and construct URL as `https://tasks.google.com/embed/?authuser=N` in `internal/tasks/actions.go` and `internal/alfred/workflow.go`

---

## Phase 9: User Story 7 - Per-Account Keywords (Priority: P3)

**Goal**: Users can configure dedicated Alfred keywords per account for faster access

**Independent Test**: Configure keyword `gtw` for work account and verify `gtw list` routes to work account

- [X] T028 [US7] Add keyword-to-account resolution in `NewWorkflow()` or a new initialization path that checks if the invoked keyword maps to a specific account in `internal/alfred/workflow.go`
- [X] T029 [US7] When a keyword-identified account is active, ignore any `@` prefix in input (keyword takes precedence) in `internal/alfred/workflow.go`
- [X] T030 [US7] Document in README.md that per-account keywords require manually adding Script Filter entries in Alfred's `info.plist` in `README.md`

---

## Phase 10: Polish & Cross-Cutting Concerns

**Purpose**: Final cleanup, documentation, and edge cases

- [X] T031 Update README.md with multi-account setup instructions, `accounts.json` schema, and usage examples in `README.md`
- [X] T032 Handle edge case: `accounts.json` deleted mid-session (fall back to single-account on next invocation) in `internal/alfred/workflow.go`
- [X] T033 [P] Add `showHelp()` updates with multi-account command examples (show `@account` syntax in help items) in `internal/alfred/workflow.go`

---

## Dependencies

```
T001 → T002, T003, T004, T005 (types before logic)
T002, T003, T004 → T006 (tests after implementation)
T006 → T007 (foundation before stories)
T007 → T008 → T009 → T010 (US1 sequential: refactor → init → handlers → verify)
T010 → T011 → T012 (US2 depends on US1 refactor)
T010 → T014 → T015 → T016 (US3 depends on US1 refactor)
T010 → T018 → T019 → T020 → T021 (US4 depends on US1 refactor)
T010 → T023 → T024 → T025 → T026 (US5 depends on US1 refactor)
T010 → T027 (US6 depends on US1 refactor)
T010 → T028 → T029 (US7 depends on US1 refactor)
US3, US4, US5 can run in parallel after US1 completes
US6, US7 can run in parallel after US1 completes
T030, T031, T032, T033 after all stories complete
```

## Parallel Execution Opportunities

After Phase 3 (US1) completes, the following stories can be implemented in parallel since they modify different functions or aspects of `workflow.go`:

- **US3** (@ prefix parsing) + **US5** (per-account auth) operate on different handler aspects
- **US6** (browser opening) + **US7** (keywords) are independent features
- **US4** (merged listing) is best done after US3 and US5 since it uses both prefix routing and per-account auth

## Implementation Strategy

**MVP**: Phase 1 + Phase 2 + Phase 3 (US1) + Phase 4 (US2) + Phase 5 (US3) + Phase 7 (US5)
This delivers: config loading, backward compatibility, `@` prefix routing, and per-account authentication. Users can manage tasks across accounts with the core `@` prefix workflow.

**Increment 2**: Phase 6 (US4) - merged task listing
**Increment 3**: Phase 8 (US6) + Phase 9 (US7) - browser opening and keywords
**Final**: Phase 10 - polish and documentation

**Total tasks**: 33
