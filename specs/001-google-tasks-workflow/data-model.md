# Data Model: Alfred Google Tasks Workflow

**Date**: 2026-07-06
**Feature**: [spec.md](spec.md)

## Entities

### Task

A single to-do item managed through the Google Tasks API.

| Field | Type | Description |
|-------|------|-------------|
| ID | string | Google Tasks API resource ID |
| Title | string | Task title text (required, non-empty) |
| DueDate | date (optional) | Due date as YYYY-MM-DD, stored in API as RFC 3339 midnight UTC |
| Status | enum | `needsAction` or `completed` |
| ListID | string | ID of the parent TaskList |
| ListName | string | Display name of the parent TaskList (denormalized for display) |
| SelfLink | string | Google Tasks API self link URL |

**Validation rules**:
- Title must be non-empty after trimming whitespace
- DueDate, if present, must be a valid date (not in the distant past, though the API accepts any date)
- Status transitions: `needsAction` -> `completed` (via Complete action), or deletion

### TaskList

A named collection of tasks.

| Field | Type | Description |
|-------|------|-------------|
| ID | string | Google Tasks API resource ID |
| Title | string | List display name (e.g., "My Tasks", "Shopping") |

**Validation rules**:
- Title must be non-empty
- Auto-creation: when a `#ListName` tag references a non-existent list, the workflow creates it via `tasklists.insert`

### OAuthToken

Stored locally as `token.json` in the workflow data directory.

| Field | Type | Description |
|-------|------|-------------|
| AccessToken | string | Short-lived bearer token (expires in ~1 hour) |
| RefreshToken | string | Long-lived token for obtaining new access tokens |
| TokenType | string | Always "Bearer" |
| Expiry | datetime | When the access token expires (RFC 3339) |

**Lifecycle**:
- Created after successful OAuth flow (login)
- AccessToken refreshed automatically when expired (using RefreshToken)
- Entire file deleted on logout
- RefreshToken does not expire unless user revokes access or token is inactive for 6 months

### ClientCredentials

User-provided `client_secret.json` from Google Cloud Console.

| Field | Type | Description |
|-------|------|-------------|
| ClientID | string | OAuth 2.0 client identifier |
| ClientSecret | string | OAuth 2.0 client secret (not truly secret for desktop apps) |
| AuthURI | string | Authorization endpoint URL |
| TokenURI | string | Token exchange endpoint URL |
| RedirectURIs | []string | Registered redirect URIs |

**Source**: Parsed from the `installed` key in the JSON file downloaded from Google Cloud Console.

## Relationships

```
TaskList 1──* Task       (a list contains zero or more tasks)
OAuthToken ──> API       (token authenticates all API requests)
ClientCredentials ──> OAuth Flow  (credentials initiate the auth flow)
```

## Timeframe Grouping (Computed, Not Stored)

When displaying tasks, each task is assigned to a computed timeframe group based on its DueDate relative to the current date:

| Group | Condition |
|-------|-----------|
| Overdue | DueDate < today |
| Today | DueDate == today |
| This Week | DueDate > today AND DueDate <= today + 7 days |
| Later | DueDate > today + 7 days |
| No Date | DueDate is nil |

Tasks are sorted by DueDate ascending within each group. Groups are displayed in the order listed above.

## Input Parsing Model (Computed, Not Stored)

User input to `gt add` is parsed into structured components:

```
Input: "Buy groceries, tomorrow #Shopping"
  -> Title:    "Buy groceries"
  -> DueDate:  tomorrow's date
  -> ListTag:  "Shopping"

Input: "Monday Review PR"
  -> Title:    "Review PR"
  -> DueDate:  next Monday's date
  -> ListTag:  nil (use default list)
```

Parsing order: extract `#ListTag` first, then check for date at boundaries (comma-separated suffix or leading token), remainder is the title.
