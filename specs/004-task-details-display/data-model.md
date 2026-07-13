# Data Model: Task Details Display

**Date**: 2026-07-13
**Feature**: 004-task-details-display

## Entities

### Task (existing, no changes)

The Google Tasks API `Task` object already contains the `Notes` field. No data model changes are required.

| Field | Type | Source | Used By |
|-------|------|--------|---------|
| Title | string | API response | Alfred item title |
| Due | string (RFC3339) | API response | Subtitle, timeframe grouping |
| Notes | string | API response | **Large Type display (NEW USAGE)** |
| Id | string | API response | Action menu reference |
| Status | string | API response | List filtering |

### TaskItem (existing wrapper, no changes)

| Field | Type | Description |
|-------|------|-------------|
| Task | *tasks.Task | Full API task object (includes Notes) |
| ListName | string | Name of the task list |
| ListID | string | ID of the task list |
| AccountName | string | Account name (multi-account mode) |

No new entities, fields, or relationships are introduced by this feature.
