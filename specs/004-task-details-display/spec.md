# Feature Specification: Task Details Display

**Feature Branch**: `004-task-details-display`  
**Created**: 2026-07-13  
**Status**: Draft  
**Input**: User description: "Add task details display via Alfred's Large Type (Cmd+L) on the main task list view"

## User Scenarios & Testing

### User Story 1 - View task notes via Large Type (Priority: P1)

A user browsing their tasks in Alfred wants to quickly peek at a task's notes/description without opening Google Tasks in a browser. They press Cmd+L on any task in the list, and Alfred displays the task's notes in a Large Type overlay.

**Why this priority**: This is the core and only feature. Viewing task details without leaving Alfred removes a major friction point in the current workflow.

**Independent Test**: Can be tested by listing tasks with `gt list`, pressing Cmd+L on a task that has notes, and verifying the notes appear in Large Type.

**Acceptance Scenarios**:

1. **Given** a task with notes exists in the user's task list, **When** the user presses Cmd+L on that task in Alfred, **Then** the task's notes text is displayed in Alfred's Large Type overlay.
2. **Given** a task without notes exists in the user's task list, **When** the user presses Cmd+L on that task in Alfred, **Then** the text "(no details)" is displayed in the Large Type overlay.

---

### User Story 2 - View task notes in multi-account merged listing (Priority: P1)

A user with multiple Google accounts configured and list_default set to "all" browses their merged task list. They press Cmd+L on any task (from any account) and see that task's notes in Large Type.

**Why this priority**: Multi-account users use the same task list view, so Large Type must work identically for merged listings.

**Independent Test**: Configure two accounts with list_default "all", create a task with notes on each account, list tasks, and press Cmd+L on tasks from each account.

**Acceptance Scenarios**:

1. **Given** a merged task listing from multiple accounts, **When** the user presses Cmd+L on a task from any account, **Then** the correct notes for that specific task are displayed.

---

### Edge Cases

- What happens when a task has very long notes (thousands of characters)? Alfred's Large Type displays the full text; truncation is handled by Alfred's own rendering.
- What happens when notes contain special characters, unicode, or emoji? They are passed through as-is to Large Type.

## Requirements

### Functional Requirements

- **FR-001**: The workflow MUST set the Large Type content on every task item rendered in the main task list view.
- **FR-002**: The workflow MUST use the task's Notes field value as the Large Type content.
- **FR-003**: When a task has no notes (empty or missing Notes field), the workflow MUST display "(no details)" as the Large Type fallback text.
- **FR-004**: Large Type MUST work for tasks in both single-account and multi-account (merged) listing modes.
- **FR-005**: The action sub-menu MUST NOT be modified by this feature.

### Error Handling

No new error handling is required for this feature. The notes data is already present in the existing task list API response, so no additional API calls or failure modes are introduced. If the notes field is empty or missing from a task, FR-003 defines the fallback behavior.

### Out of Scope

- Modifying the action sub-menu (Cmd+Enter) for tasks
- Editing or updating task notes from within Alfred
- Displaying notes in search result views or other non-list views
- Formatting or transforming note content (HTML stripping, markdown rendering)

## Success Criteria

### Measurable Outcomes

- **SC-001**: Users can view any task's notes by pressing Cmd+L, without opening a browser or leaving Alfred.
- **SC-002**: 100% of tasks in the list view have Large Type content set (either notes or fallback text).
- **SC-003**: No additional API calls are introduced by this feature (notes data is already available in the existing list response).

## Assumptions

### Data & Platform
- The Google Tasks API List endpoint already returns the Notes field for each task in its response. No additional API calls or field selection changes are needed.
- Alfred's Large Type feature (Cmd+L) is available in all supported Alfred versions (Alfred 5+).

### Implementation
- The AwGo library's `.Largetype()` method on items is the correct API for setting Large Type content. This should be verified during implementation.
