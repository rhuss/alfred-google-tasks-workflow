# Feature Specification: Alfred Google Tasks Workflow

**Feature Branch**: `001-google-tasks-workflow`
**Created**: 2026-07-06
**Status**: Draft
**Input**: User description: "Alfred Google Tasks Workflow - Go-based Alfred workflow for managing Google Tasks with OAuth, smart date parsing, and task list management"

## User Scenarios & Testing

### User Story 1 - First-Time Authentication (Priority: P1)

A user installs the Alfred workflow and needs to connect their Google account before they can manage tasks. They place their Google Cloud `client_secret.json` file in the workflow's data directory, then trigger the login command. The workflow opens their browser for Google's consent screen, and after granting permission, the browser shows a confirmation page. The user returns to Alfred and can immediately start managing tasks.

**Why this priority**: Without authentication, no other feature works. This is the foundation for all task operations.

**Independent Test**: Can be fully tested by running the login command and verifying that tokens are stored and the workflow can make authenticated API calls.

**Acceptance Scenarios**:

1. **Given** the user has placed `client_secret.json` in the workflow data directory, **When** they type `gt login` in Alfred, **Then** their default browser opens to Google's OAuth consent screen
2. **Given** the user grants permission in the browser, **When** Google redirects to the local callback URL, **Then** the workflow captures the authorization code, exchanges it for tokens, stores them, and shows a success page in the browser
3. **Given** the user has not placed `client_secret.json`, **When** they type `gt login`, **Then** the workflow shows an error message explaining how to set up credentials
4. **Given** the user denies permission in the browser, **When** the OAuth flow completes, **Then** the workflow handles the error gracefully and shows a message in Alfred
5. **Given** the user's access token has expired, **When** they use any `gt` command, **Then** the workflow automatically refreshes the token using the stored refresh token without user interaction

---

### User Story 2 - Quick Task Creation (Priority: P1)

A user wants to quickly capture a task from Alfred without leaving their current context. They type a task title with an optional natural language date and optional list tag. The task is created in Google Tasks with the correct due date and in the specified list.

**Why this priority**: Task creation is the primary use case. Users need fast, frictionless task capture.

**Independent Test**: Can be fully tested by creating tasks with various date and list combinations and verifying they appear correctly in Google Tasks.

**Acceptance Scenarios**:

1. **Given** the user is authenticated, **When** they type `gt add Buy groceries`, **Then** a task titled "Buy groceries" is created in the default task list with no due date
2. **Given** the user is authenticated, **When** they type `gt add Buy groceries, tomorrow`, **Then** a task titled "Buy groceries" is created with tomorrow's date as the due date
3. **Given** the user is authenticated, **When** they type `gt add Buy groceries, tomorrow #Shopping`, **Then** a task titled "Buy groceries" is created with tomorrow's date in the "Shopping" list
4. **Given** the user references a list that does not exist (e.g., `#NewList`), **When** the task is created, **Then** the workflow creates the new list first, then adds the task to it
5. **Given** the user types `gt add Monday Review PR`, **Then** a task titled "Review PR" is created with next Monday's date as the due date
6. **Given** the user types `gt add 2026-07-10 Submit report`, **Then** a task titled "Submit report" is created with 2026-07-10 as the due date
7. **Given** the user types `gt add Submit report, 07-10`, **Then** a task titled "Submit report" is created with the date interpreted as July 10 of the current year
8. **Given** the user is not authenticated, **When** they try to add a task, **Then** the workflow shows a message prompting them to run `gt login` first

---

### User Story 3 - Task Listing with Timeframe Grouping (Priority: P1)

A user wants to see their upcoming tasks at a glance, organized by urgency. They type the list command and see tasks grouped into Overdue, Today, This Week, and Later sections, with each task showing its title, due date, and which list it belongs to.

**Why this priority**: Viewing tasks is essential for task management. The timeframe grouping provides immediate actionable context.

**Independent Test**: Can be fully tested by creating tasks with various due dates and verifying the grouping and sorting in Alfred's results.

**Acceptance Scenarios**:

1. **Given** the user has tasks with various due dates, **When** they type `gt list`, **Then** tasks are displayed grouped by: Overdue (past due), Today, This Week (next 7 days), Later (beyond 7 days)
2. **Given** tasks exist in multiple lists, **When** the user types `gt list`, **Then** each task shows the list name as a subtitle
3. **Given** the user types `gt list #Work`, **Then** only tasks from the "Work" list are shown, still grouped by timeframe
4. **Given** there are no tasks matching the filter, **When** the user types `gt list`, **Then** the workflow shows a "No tasks found" message
5. **Given** multiple tasks exist within the same timeframe group, **When** displayed, **Then** they are sorted by due date (earliest first)
6. **Given** a task has no due date, **When** the user types `gt list`, **Then** the task appears in a "No Date" section at the end

---

### User Story 4 - Task Actions (Priority: P2)

A user viewing their task list wants to take action on a specific task. They select a task and see a sub-menu with options to complete it, open it in the browser, or delete it.

**Why this priority**: Acting on tasks (completing, deleting) is important but secondary to creating and viewing tasks.

**Independent Test**: Can be fully tested by selecting a task from the list and exercising each action option.

**Acceptance Scenarios**:

1. **Given** the user is viewing the task list, **When** they select a task (press Enter), **Then** a sub-menu appears with options: "Complete", "Open in Browser", "Delete"
2. **Given** the user selects "Complete" from the sub-menu, **When** the action executes, **Then** the task is marked as completed in Google Tasks and a confirmation notification is shown
3. **Given** the user selects "Open in Browser", **When** the action executes, **Then** the Google Tasks web UI opens in the default browser
4. **Given** the user selects "Delete", **When** the action executes, **Then** the task is deleted from Google Tasks and a confirmation notification is shown

---

### User Story 5 - Open Google Tasks Web UI (Priority: P3)

A user wants to switch to the full Google Tasks interface for more complex task management. They type the open command and the web UI launches in their browser.

**Why this priority**: This is a convenience feature. Users can always navigate to Google Tasks manually.

**Independent Test**: Can be fully tested by running the command and verifying the browser opens to the correct URL.

**Acceptance Scenarios**:

1. **Given** the user types `gt open`, **When** the command executes, **Then** the default browser opens to `https://tasks.google.com/`

---

### User Story 6 - Account Logout (Priority: P3)

A user wants to disconnect their Google account from the workflow. They type the logout command, stored tokens are deleted, and they receive a confirmation.

**Why this priority**: Account management is a hygiene feature. Users rarely need it but it must exist.

**Independent Test**: Can be fully tested by running logout and verifying tokens are removed and subsequent commands prompt for re-authentication.

**Acceptance Scenarios**:

1. **Given** the user is authenticated, **When** they type `gt logout`, **Then** stored OAuth tokens are deleted and a confirmation notification is shown
2. **Given** the user is not authenticated, **When** they type `gt logout`, **Then** the workflow shows a message indicating no account is connected

---

### Edge Cases

- What happens when the local OAuth server times out waiting for the callback? The workflow should show a timeout error after 3 minutes and clean up the server.
- What happens when the user's internet connection drops during an API call? The workflow should show a connection error message.
- What happens when the Google Tasks API rate limit is hit? The workflow should show a rate limit message and suggest trying again later.
- What happens when the user has no task lists at all? The workflow should handle this gracefully and create a default list on first task creation.
- What happens when a `#ListName` tag contains spaces? The workflow should support multi-word list names using `#My-List` or `#My_List` syntax with hyphens or underscores replacing spaces.
- What happens when the date text is ambiguous (e.g., "next week" on a Sunday)? The workflow should consistently interpret "next week" as next Monday.

## Clarifications

### Session 2026-07-06

- Q: Should task listing cache API results for faster display? → A: No caching in v1. Always fetch fresh from API on each `gt list` invocation. Keeps implementation simple and ensures data is always current.
- Q: What happens when user types `gt` with no subcommand? → A: Equivalent to `gt list`. Viewing tasks is the most common action and provides immediate value.
- Q: Should a `gt logout` command be included? → A: Yes. `gt logout` deletes stored tokens from the workflow data directory and shows a confirmation notification. Users need a way to disconnect their account.
- Q: How should action confirmations be displayed? → A: Via macOS notifications using Alfred's built-in notification system (brief, non-blocking). Applies to task creation, completion, and deletion confirmations.

## Requirements

### Functional Requirements

- **FR-001**: System MUST authenticate users via Google OAuth 2.0 using the localhost loopback redirect pattern with PKCE (S256)
- **FR-002**: System MUST support user-provided `client_secret.json` credentials placed in the workflow's data directory
- **FR-003**: System MUST store OAuth tokens (access token and refresh token) as a JSON file in Alfred's workflow data directory
- **FR-004**: System MUST automatically refresh expired access tokens using the stored refresh token without user interaction
- **FR-005**: System MUST create Google Tasks with a title and optional due date via the Google Tasks API
- **FR-006**: System MUST parse natural language dates from task input: "today", "tomorrow", "next week", weekday names ("Monday" through "Sunday"), ISO dates ("YYYY-MM-DD"), and short dates ("MM-DD")
- **FR-007**: System MUST extract dates from either the beginning or end of the task title, with comma separation supported. If no recognized date pattern is matched, the entire input MUST be treated as the task title with no due date
- **FR-008**: System MUST support `#ListName` syntax to specify a target task list, creating the list if it does not exist
- **FR-009**: System MUST list tasks grouped by timeframe sections: Overdue, Today, This Week (next 7 days), Later, and No Date
- **FR-010**: System MUST display each task with its title, due date, and parent list name
- **FR-011**: System MUST sort tasks by due date within each timeframe group (earliest first)
- **FR-012**: System MUST filter task listing by `#ListName` tag when provided
- **FR-013**: System MUST provide a sub-menu when selecting a task with options: Complete, Open in Browser, Delete
- **FR-014**: System MUST open the Google Tasks web UI in the default browser via the `gt open` command
- **FR-015**: System MUST show clear error messages when credentials are missing, authentication fails, or API calls fail
- **FR-016**: System MUST use the `https://www.googleapis.com/auth/tasks` OAuth scope for read-write access
- **FR-017**: System MUST treat `gt` with no subcommand as equivalent to `gt list`
- **FR-018**: System MUST provide a `gt logout` command that deletes stored OAuth tokens and shows a confirmation notification
- **FR-019**: System MUST display action confirmations (task created, completed, deleted) as macOS notifications via Alfred's notification system
- **FR-020**: System MUST fetch task data fresh from the Google Tasks API on each listing request (no local caching in v1)

### Key Entities

- **Task**: A single to-do item with a title, optional due date, completion status, and parent list. Corresponds to a Google Tasks API task resource.
- **Task List**: A named collection of tasks. Corresponds to a Google Tasks API tasklist resource. Users reference lists via `#ListName` syntax.
- **OAuth Token**: A pair of access token (short-lived, 1 hour) and refresh token (long-lived) used to authenticate API requests. Stored locally as JSON.
- **Client Credentials**: The `client_secret.json` file containing the OAuth client ID and secret, provided by the user from their Google Cloud Console project.

## Success Criteria

### Measurable Outcomes

- **SC-001**: Users can authenticate with Google and start managing tasks within 2 minutes of first launch (excluding Google Cloud project setup)
- **SC-002**: Task creation completes within 2 seconds of pressing Enter in Alfred
- **SC-003**: Task listing loads and displays within 3 seconds of typing the list command
- **SC-004**: Natural date parsing correctly interprets all supported date formats ("today", "tomorrow", "next week", weekday names, ISO dates, short dates) with 100% accuracy for the defined pattern set
- **SC-005**: The workflow operates as a single compiled binary with zero runtime dependencies beyond Alfred itself
- **SC-006**: Token refresh happens transparently, so users authenticate once and never need to re-authenticate unless they revoke access

## Dependencies

- **Google Tasks API v1**: REST API for task and task list CRUD operations
- **Alfred 5 with Powerpack**: macOS launcher providing workflow execution, Script Filter UI, and notification system
- **Go compiler**: Current stable version, for building the single-binary workflow
- **macOS**: Target platform (Alfred is macOS-only)

## Out of Scope (v1)

- Subtask creation or hierarchy management
- Recurring/repeating tasks
- Task editing or renaming after creation
- Multi-account support (only one Google account at a time)
- Offline mode or local caching
- Task notes or descriptions
- Drag-and-drop reordering

## Assumptions

- Users have Alfred 5 with the Powerpack installed (required for workflow functionality)
- Users can create a Google Cloud project and configure OAuth credentials (documented in workflow README)
- Users have a stable internet connection for API calls (offline mode is out of scope)
- The Google Tasks API remains stable and backward-compatible with the v1 API
- macOS is the only supported platform (Alfred is macOS-only)
- The workflow targets the current stable version of Go for compilation
- Tasks without a due date are shown in the listing under a "No Date" section
- "Next week" consistently means next Monday regardless of the current day
- The default task list (when no `#ListName` is provided) is the user's first task list as returned by the Google Tasks API (typically "My Tasks")
- Multi-word list names use hyphens or underscores in the `#ListName` syntax (e.g., `#My-List`)
