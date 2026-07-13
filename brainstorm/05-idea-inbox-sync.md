# Brainstorm: Idea Inbox Sync (Google Tasks to Obsidian)

**Date:** 2026-07-13
**Status:** active

## Problem Framing

Ideas happen on the go. The fastest way to capture them on Android is via Gemini voice ("Add a task [title] to my Ideas list"). But Google Tasks is not where ideas should live long-term. Obsidian is the knowledge base where ideas get processed, expanded, and connected.

The gap: there's no bridge from Google Tasks to Obsidian. Ideas accumulate in Google Tasks and get forgotten. The workflow should automatically extract ideas from a designated Google Tasks list and append them to an Obsidian inbox file, so they flow into the existing knowledge management system without manual effort.

## Approaches Considered

### A: Inline sync during list fetch (chosen)
Hook into the existing `handleList` / `handleListAllAccounts` flow. After fetching tasks for display, also check the Ideas list on each account. New ideas get appended to the Obsidian file and deleted from Google Tasks.

- Pros: zero latency overhead when Ideas list is already in scope, no new commands, no new entry points
- Cons: list operations do a write side effect (file I/O + delete API call), which is unusual for a "list" operation

### B: Background sync via Alfred's Run Script
Add a separate Run Script action in Alfred that triggers after the Script Filter completes. The idea sync runs asynchronously.

- Pros: list rendering stays fast and side-effect-free, errors in sync don't break listing
- Cons: requires Alfred workflow plumbing, two separate execution paths

### C: Periodic sync via Alfred's cron-like mechanism
Use Alfred's "Run Script" with a hotkey or periodic trigger.

- Pros: completely decoupled, predictable timing
- Cons: contradicts the automatic requirement, ideas could sit unsynced

## Decision

**Approach A: inline sync during list fetch.** Simplest implementation, matches the "zero friction" goal. The side effect (appending to a file + deleting a task) is fast. Error handling is straightforward: if the file path isn't set, skip silently.

## Key Requirements

- **Trigger**: automatic side effect during every `gt list` operation (including merged multi-account listing)
- **Account scope**: always scans the Ideas list on all authenticated accounts, regardless of which account the user targeted with `@`
- **Inbox file path**: configurable via Alfred workflow variable (`IDEA_INBOX_PATH`). Feature entirely disabled if not set.
- **Ideas list name**: configurable via a single global Alfred variable (`IDEA_LIST_NAME`). If the list doesn't exist on an account, skip silently.
- **Entry format**: H3 heading per idea, with Date, Account, and TaskID metadata. Description (from task notes) below if present, omitted if empty.
- **Deduplication**: by TaskID in the metadata lines of the existing inbox file
- **Post-sync**: delete the task from Google Tasks after successfully writing to the inbox file
- **Error handling**: silent skip on missing config, missing list, or unauthenticated accounts. Never block or break the list display.
- **Voice capture**: designed for Gemini on Android ("Add a task [title] to my Ideas list"). Title only, no notes expected.

### Inbox file format

```markdown
# Idea Inbox

### Build a CLI for X
- Date: 2026-07-13
- Account: work
- TaskID: abc123xyz

Some description text from task notes

### Refactor auth flow
- Date: 2026-07-10
- Account: personal
- TaskID: def456uvw
```

## Open Questions

- Should there be a notification when ideas are synced (e.g., "2 ideas synced to inbox"), or should it be completely silent?
- If the inbox file doesn't exist yet, should the workflow create it with the `# Idea Inbox` header, or require the user to create it first?
