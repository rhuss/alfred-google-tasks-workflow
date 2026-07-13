# Data Model: Move Task Between Accounts

**Date**: 2026-07-13

## Entities

### Task (Google Tasks API: tasks.Task)

Fields relevant to the move operation:

| Field | Type | Copied on Move | Notes |
|-------|------|----------------|-------|
| Title | string | Yes | Task description text |
| Due | string (RFC 3339) | Yes | Due date, may be empty |
| Notes | string | Yes | Additional notes, may be empty |
| Id | string | No | Account-specific, new ID assigned on target |
| Status | string | No | New task created as "needsAction" |

### TaskList (Google Tasks API: tasks.TaskList)

| Field | Type | Used | Notes |
|-------|------|------|-------|
| Id | string | Yes | Used to resolve source list, find/create target list |
| Title | string | Yes | Matched case-insensitively between accounts |

### Account (internal: auth.Account)

| Field | Type | Used | Notes |
|-------|------|------|-------|
| Name | string | Yes | Displayed in "Move to {name}" menu entry |
| Credentials | string | Yes | Path to OAuth credentials for target client |
| ProfileIndex | int | No | Not needed for API calls, only for browser URLs |

## Relationships

- A **Task** belongs to exactly one **TaskList** on one **Account**
- A move operation creates a new **Task** on the target **Account**'s matching **TaskList**
- The source **Task** is deleted after successful creation on the target
