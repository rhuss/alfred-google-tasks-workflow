# Tasks: Alfred Google Tasks Workflow

**Input**: Design documents from `specs/001-google-tasks-workflow/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, quickstart.md

**Tests**: Tests included as part of implementation tasks (Go convention: test files alongside source).

**Organization**: Tasks grouped by user story for independent implementation and testing.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization, Go module, Alfred workflow scaffold

- [x] T001 Initialize Go module with `go mod init github.com/rhuss/alfred-google-tasks-workflow` in go.mod
- [x] T002 Add dependencies: AwGo, google-api-go-client, golang.org/x/oauth2 in go.mod and run `go mod tidy`
- [x] T003 [P] Create Makefile with targets: build, test, clean, package (build binary, create .alfredworkflow zip)
- [x] T004 [P] Create Alfred workflow manifest in info.plist with keyword `gt`, bundle ID `com.rhuss.gtasks`, and script filter configuration for all commands (login, logout, add, list, open)
- [x] T005 [P] Create workflow icon in icon.png (placeholder)

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure that MUST be complete before user stories

**Critical**: OAuth and API client are prerequisites for all task operations

- [x] T006 Implement client credentials parsing: read and validate `client_secret.json` from Alfred workflow data directory in internal/auth/credentials.go
- [x] T007 Implement OAuth flow with PKCE: code_verifier generation, code_challenge (S256), temporary HTTP server on 127.0.0.1, browser open, authorization code capture, token exchange in internal/auth/oauth.go
- [x] T008 Implement token storage: load/save/delete token.json, automatic access token refresh using refresh token, expiry checking in internal/auth/token.go
- [x] T009 Implement Google Tasks API client wrapper: authenticated HTTP client creation, tasklists.list, tasklists.insert, tasks.list, tasks.insert, tasks.patch (complete), tasks.delete in internal/tasks/client.go
- [x] T010 Implement custom date parser with support for: "today", "tomorrow", "next week", weekday names (case-insensitive, next occurrence), ISO dates (YYYY-MM-DD), short dates (MM-DD, current/next year) in internal/dateparse/dateparse.go
- [x] T011 Implement input parser: extract title, optional date, optional #ListName tag from user input string, handling comma-separated and leading date positions in internal/input/parser.go
- [x] T012 [P] Implement Alfred workflow setup and command routing: initialize AwGo workflow, route subcommands (login, logout, add, list, open, actions) in internal/alfred/workflow.go
- [x] T013 [P] Implement Alfred notification helpers: macOS notification for task created/completed/deleted confirmations in internal/alfred/notifications.go
- [x] T014 Create main entry point: parse os.Args, delegate to workflow router in cmd/main.go
- [x] T015 Write unit tests for date parser covering all patterns, edge cases (past weekdays, year rollover for MM-DD, "next week" on Sunday) in internal/dateparse/dateparse_test.go
- [x] T016 [P] Write unit tests for input parser covering title-only, title+date, title+date+list, leading date, comma separation, list-only in internal/input/parser_test.go

**Checkpoint**: Foundation ready. OAuth flow, API client, date parser, and input parser all functional.

---

## Phase 3: User Story 1 - First-Time Authentication (Priority: P1) MVP

**Goal**: User can authenticate with Google and store credentials for subsequent use

**Independent Test**: Run `gt login`, complete browser OAuth flow, verify token.json is created and API calls succeed

- [x] T017 [US1] Implement `gt login` command handler: check for client_secret.json, start OAuth flow, display success/error in Alfred in internal/alfred/workflow.go
- [x] T018 [US1] Implement OAuth success HTML page: browser shows "Authenticated! You can close this tab." after successful callback in internal/auth/oauth.go
- [x] T019 [US1] Implement error handling for missing credentials, denied permission, OAuth timeout (3 min), and network errors in internal/auth/oauth.go
- [x] T020 [US1] Implement automatic token refresh: before each API call, check token expiry and refresh if needed, handle invalid_grant by prompting re-login in internal/auth/token.go
- [x] T021 [US1] Implement unauthenticated state detection: all `gt` commands check for valid token and show "Run gt login first" message if missing in internal/alfred/workflow.go

---

## Phase 4: User Story 2 - Quick Task Creation (Priority: P1)

**Goal**: User can create tasks with smart date parsing and list targeting from Alfred

**Independent Test**: Type `gt add Buy groceries, tomorrow #Shopping`, verify task appears in Google Tasks with correct date and list

- [x] T022 [US2] Implement `gt add` command handler: parse input, resolve list (create if needed), create task via API, show notification in internal/tasks/create.go
- [x] T023 [US2] Implement task list resolution: lookup list by name from cached tasklists.list response, create new list via tasklists.insert if not found in internal/tasks/client.go
- [x] T024 [US2] Implement date-to-RFC3339 conversion: convert parsed date to Google Tasks API format (midnight UTC) in internal/tasks/create.go
- [x] T025 [US2] Write integration test for task creation with various input combinations (title only, title+date, title+date+list, auto-create list) in internal/tasks/create_test.go

---

## Phase 5: User Story 3 - Task Listing with Timeframe Grouping (Priority: P1)

**Goal**: User can view tasks grouped by Overdue/Today/This Week/Later/No Date

**Independent Test**: Type `gt list`, verify tasks appear grouped by timeframe with correct sorting and subtitles

- [x] T026 [US3] Implement task fetching: retrieve tasks from all lists (or filtered list), aggregate, sort by due date in internal/tasks/list.go
- [x] T027 [US3] Implement timeframe grouping: classify tasks into Overdue, Today, This Week, Later, No Date groups based on due date vs current date in internal/tasks/list.go
- [x] T028 [US3] Implement Alfred Script Filter item builder: render grouped tasks as Alfred items with title, subtitle (list name + due date), arg (task ID + list ID), icon per group in internal/alfred/items.go
- [x] T029 [US3] Implement `gt list` command handler with optional #ListName filter: parse filter, fetch tasks, group, render items in internal/alfred/workflow.go
- [x] T030 [US3] Implement `gt` (no subcommand) as alias for `gt list` in internal/alfred/workflow.go
- [x] T031 [US3] Write unit tests for timeframe grouping logic with tasks across all groups in internal/tasks/list_test.go

---

## Phase 6: User Story 4 - Task Actions (Priority: P2)

**Goal**: User can complete, open, or delete a task from the Alfred sub-menu

**Independent Test**: Select a task from list, choose "Complete", verify task is marked done in Google Tasks

- [x] T032 [US4] Implement task action sub-menu: when user selects a task, show Script Filter with Complete/Open in Browser/Delete options in internal/alfred/items.go
- [x] T033 [US4] Implement complete action: patch task status to "completed" via API, show notification in internal/tasks/actions.go
- [x] T034 [US4] Implement delete action: delete task via API, show notification in internal/tasks/actions.go
- [x] T035 [US4] Implement open-in-browser action: open https://tasks.google.com/ in default browser via `open` command in internal/tasks/actions.go

---

## Phase 7: User Story 5 - Open Google Tasks Web UI (Priority: P3)

**Goal**: User can open Google Tasks in their browser

**Independent Test**: Type `gt open`, verify browser opens to tasks.google.com

- [x] T036 [US5] Implement `gt open` command handler: open https://tasks.google.com/ in default browser in internal/alfred/workflow.go

---

## Phase 8: User Story 6 - Account Logout (Priority: P3)

**Goal**: User can disconnect their Google account

**Independent Test**: Type `gt logout`, verify token.json is deleted and subsequent commands prompt for login

- [x] T037 [US6] Implement `gt logout` command handler: delete token.json, show confirmation notification in internal/alfred/workflow.go

---

## Phase 9: Polish & Cross-Cutting Concerns

**Purpose**: Final quality, documentation, packaging

- [x] T038 [P] Add README.md with setup instructions (Google Cloud project, credentials, installation), usage examples, and build instructions
- [x] T039 [P] Harden error handling in command handlers: (1) in internal/tasks/client.go, wrap all API calls with a helper that detects HTTP 429 (rate limit) and returns a user-friendly "Rate limit reached, try again in a moment" error; (2) in internal/tasks/client.go, detect network errors (timeout, DNS, connection refused) and return "No internet connection" message; (3) in internal/alfred/workflow.go, add a top-level error handler that converts all error types to Alfred items with appropriate icons and messages; (4) in internal/auth/token.go, handle invalid_grant errors during refresh by deleting the stored token and returning "Session expired, run gt login again"
- [x] T040 Verify end-to-end workflow: build binary, package .alfredworkflow, test all commands in Alfred. Validate NFR thresholds: (1) measure task creation round-trip time (target < 2s from Enter to notification), (2) measure task listing time (target < 3s from keystroke to displayed results), (3) verify single binary with `file` and `otool -L` (no dynamic deps beyond system libs), (4) verify first-use auth completes within 2 minutes (excluding Google Cloud setup)

---

## Dependencies

```
T001-T005 (Setup) → T006-T016 (Foundation) → T017-T021 (US1: Auth)
                                              ↓
                              T022-T025 (US2: Create) ← requires auth
                              T026-T031 (US3: List)   ← requires auth
                              T032-T035 (US4: Actions) ← requires list
                              T036 (US5: Open)        ← independent
                              T037 (US6: Logout)      ← requires auth
                                              ↓
                              T038-T040 (Polish)
```

**Story independence**: US2 (Create) and US3 (List) can be implemented in parallel after US1 (Auth). US4 (Actions) depends on US3 (List) for the sub-menu. US5 (Open) and US6 (Logout) are independent of each other.

## Parallel Execution Opportunities

- **Phase 1**: T003, T004, T005 are independent
- **Phase 2**: T012, T013 are independent of T006-T011; T015, T016 are independent of each other
- **Phase 4+5**: US2 and US3 can be implemented in parallel after US1 completes
- **Phase 9**: T038, T039 are independent

## Implementation Strategy

1. **MVP (Phases 1-5)**: Setup + Foundation + Auth + Create + List = functional workflow that can add and view tasks
2. **Complete (Phases 6-8)**: Actions + Open + Logout = full feature set
3. **Ship (Phase 9)**: Polish + README + packaging

Total tasks: 40
- Phase 1 (Setup): 5 tasks
- Phase 2 (Foundation): 11 tasks
- Phase 3 (US1 Auth): 5 tasks
- Phase 4 (US2 Create): 4 tasks
- Phase 5 (US3 List): 6 tasks
- Phase 6 (US4 Actions): 4 tasks
- Phase 7 (US5 Open): 1 task
- Phase 8 (US6 Logout): 1 task
- Phase 9 (Polish): 3 tasks
