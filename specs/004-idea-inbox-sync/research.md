# Research: Idea Inbox Sync

## Google Tasks API - Task Timestamps

**Decision**: Use the `Updated` field from the Google Tasks API `Task` struct for the Date metadata.

**Rationale**: The Google Tasks API v1 `Task` resource has an `Updated` field (RFC 3339 timestamp) but no explicit `Created` field. For newly created tasks that haven't been modified, `Updated` equals the creation time. This is the closest available proxy for creation date.

**Alternatives considered**:
- Using the task ID (contains no date information)
- Storing creation date separately (unnecessary complexity)

## Existing Codebase Integration Points

**Decision**: Hook the idea sync into `handleList` (single-account) and `handleListAllAccounts` (multi-account) in `internal/alfred/workflow.go`. The sync runs after task fetching but before rendering.

**Rationale**: Both list paths already have access to authenticated clients and account contexts. The multi-account path already iterates over all accounts, so adding idea sync there is natural. For single-account mode, the sync adds one extra API call (to find and fetch the Ideas list).

**Key observation**: In `handleListAllAccounts`, the workflow already iterates over `AccountConfig.AccountNames()` and creates clients for each account. The idea sync must also scan all accounts, so it needs its own iteration independent of the list display path, because:
1. `handleList` (single-account targeted via `@account`) should still sync ALL accounts
2. The sync must happen even when listing a specific `#ListName` filter

**Decision**: Create a separate `syncIdeasToInbox` method on Workflow that always iterates all accounts. Call it from both `handleList` and `handleListAllAccounts` before rendering.

## File I/O Strategy

**Decision**: Use simple file append with a read-scan-for-dedup approach.

**Rationale**: The inbox file is small (hundreds of entries at most). Reading the entire file to scan for existing TaskIDs, then appending new entries, is fast and simple. No need for a database or index file.

**Alternatives considered**:
- Separate tracking file for synced IDs (adds complexity, risks getting out of sync)
- SQLite database (massive overkill for this use case)

## Error Isolation

**Decision**: Wrap the entire idea sync in a recover/catch block that logs to stderr and never propagates errors to the caller.

**Rationale**: FR-009 requires that idea sync never blocks or interferes with task listing. Any error (file I/O failure, API error, parse error) must be silently absorbed. The listing must always complete normally.
