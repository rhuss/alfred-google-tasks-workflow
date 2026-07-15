# Data Model: Idea Inbox Sync

## Entities

### IdeaEntry

Represents a single idea extracted from Google Tasks and written to the Obsidian inbox file.

| Field | Source | Format | Required |
|-------|--------|--------|----------|
| Title | `task.Title` | H3 heading text | Yes |
| Date | `task.Updated` (parsed from RFC 3339) | YYYY-MM-DD | Yes |
| Account | `AccountContext.Name` | Plain text | No (omitted in single-account mode) |
| TaskID | `task.Id` | Google Tasks ID string | Yes |
| Description | `task.Notes` | Plain text block | No (omitted when empty) |

### IdeaSyncConfig

Configuration read from Alfred workflow environment variables.

| Field | Env Var | Default | Required |
|-------|---------|---------|----------|
| InboxPath | `IDEA_INBOX_PATH` | (none) | Yes (feature disabled if unset) |
| ListName | `IDEA_LIST_NAME` | (none) | Yes (feature disabled if unset) |

## Relationships

- An IdeaEntry is derived from a Google Tasks `Task` resource
- An IdeaEntry belongs to one Account (identified by name)
- An IdeaEntry is written to exactly one Inbox File (at InboxPath)
- Multiple IdeaEntries from different accounts coexist in the same Inbox File

## State Transitions

```
Google Tasks "Ideas" List          Obsidian Inbox File
─────────────────────────          ───────────────────
  [task exists]
       │
       ▼
  gt list triggered
       │
       ├── TaskID found in inbox? ──YES──► skip (already synced)
       │
       NO
       │
       ▼
  append entry to inbox file
       │
       ▼
  delete task from Google Tasks
       │
       ▼
  [task removed, entry persists]
```
