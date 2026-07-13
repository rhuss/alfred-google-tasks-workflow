# Feature Specification: Idea Inbox Sync

**Feature Branch**: `004-idea-inbox-sync`
**Created**: 2026-07-13
**Status**: Draft
**Input**: Brainstorm document: brainstorm/05-idea-inbox-sync.md

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Automatic Idea Extraction During Task Listing (Priority: P1)

A user captures ideas throughout the day using Gemini voice on Android ("Add a task [title] to my Ideas list"). Later, when the user opens Alfred and runs `gt list` to view their tasks, the workflow silently checks the configured Ideas list on all authenticated Google Tasks accounts. Any new ideas found are appended to the user's Obsidian inbox file and deleted from Google Tasks. The user sees their normal task list without any interruption or extra steps. The ideas are now in Obsidian, ready to be processed.

**Why this priority**: This is the core value proposition. Without automatic extraction, ideas stay trapped in Google Tasks and the feature has no purpose.

**Independent Test**: Can be tested by creating a task in the Ideas list, running `gt list`, and verifying the task appears in the Obsidian inbox file and is removed from Google Tasks.

**Acceptance Scenarios**:

1. **Given** IDEA_INBOX_PATH and IDEA_LIST_NAME are configured, and an Ideas list exists on the user's Google Tasks account with 2 tasks, **When** the user runs `gt list`, **Then** both tasks are appended to the Obsidian inbox file with correct H3 headings, Date, Account, and TaskID metadata, and both tasks are deleted from Google Tasks.

2. **Given** IDEA_INBOX_PATH and IDEA_LIST_NAME are configured, and the Ideas list on one account has 1 new task and the inbox file already contains a previously synced task, **When** the user runs `gt list`, **Then** only the new task is appended (the existing task is skipped by dedup), and only the new task is deleted from Google Tasks.

3. **Given** IDEA_INBOX_PATH and IDEA_LIST_NAME are configured, and the Ideas list is empty on all accounts, **When** the user runs `gt list`, **Then** the task listing displays normally with no changes to the inbox file.

---

### User Story 2 - Multi-Account Idea Collection (Priority: P1)

A user has both a personal and work Google account configured. Ideas are captured on whichever account is convenient at the moment. During `gt list`, the workflow scans the Ideas list on all authenticated accounts, regardless of which account the user targeted with `@`. Ideas from both accounts are written to the same inbox file, each tagged with their source account name.

**Why this priority**: Multi-account support is essential because ideas are captured on whichever device/account is at hand. Without this, ideas on non-default accounts would be missed.

**Independent Test**: Can be tested by adding an idea to the Ideas list on each of two accounts, running `gt list`, and verifying both ideas appear in the inbox file with correct account tags.

**Acceptance Scenarios**:

1. **Given** two accounts (work, personal) are configured and authenticated, each with tasks in their Ideas list, **When** the user runs `gt list` (even with `@work`), **Then** ideas from both accounts are synced to the inbox file, each with the correct Account metadata.

2. **Given** one account is authenticated and another is not, **When** the user runs `gt list`, **Then** ideas from the authenticated account are synced, and the unauthenticated account is silently skipped.

---

### User Story 3 - Feature Disabled When Not Configured (Priority: P2)

A user who does not use Obsidian or does not want idea syncing simply leaves the IDEA_INBOX_PATH variable unset. The workflow behaves identically to the current behavior with zero overhead.

**Why this priority**: The feature must be opt-in and non-disruptive to existing users.

**Independent Test**: Can be tested by unsetting IDEA_INBOX_PATH, running `gt list`, and verifying no file operations occur and no errors are shown.

**Acceptance Scenarios**:

1. **Given** IDEA_INBOX_PATH is not set, **When** the user runs `gt list`, **Then** no idea sync occurs, no file operations happen, and the task listing displays normally.

2. **Given** IDEA_INBOX_PATH is set but IDEA_LIST_NAME is not set, **When** the user runs `gt list`, **Then** no idea sync occurs (both variables are required).

---

### User Story 4 - Inbox File Auto-Creation (Priority: P3)

A user configures IDEA_INBOX_PATH but the target file does not exist yet. The workflow creates the file with a `# Idea Inbox` header on the first sync, so the user does not need to manually create it.

**Why this priority**: Minor convenience that reduces setup friction.

**Independent Test**: Can be tested by setting IDEA_INBOX_PATH to a non-existent file, creating an idea task, running `gt list`, and verifying the file is created with the header and the idea is appended.

**Acceptance Scenarios**:

1. **Given** IDEA_INBOX_PATH points to a file that does not exist, and the Ideas list has tasks, **When** the user runs `gt list`, **Then** the file is created with `# Idea Inbox` as the first line, followed by the synced ideas.

2. **Given** IDEA_INBOX_PATH points to a file that does not exist, and the parent directory does not exist, **When** the user runs `gt list`, **Then** the parent directories are created and the file is created with the header and synced ideas.

---

### Edge Cases

- What happens when the inbox file exists but has been manually edited with custom content above the ideas? The workflow appends to the end of the file, preserving all existing content.
- What happens when a task is created in Google Tasks but deleted before the next `gt list`? Nothing happens since the task no longer exists in the Ideas list.
- What happens when the same task title exists on two different accounts? Both are synced because deduplication is by TaskID (unique per task), not by title.
- What happens when the Obsidian inbox file write succeeds but the Google Tasks delete fails? The task remains in Google Tasks and will be detected as a duplicate on the next sync (dedup by TaskID), so no data loss occurs. The error is logged to stderr but does not block the task listing.
- What happens when the Ideas list exists but contains only completed tasks? Completed tasks are not returned by the Google Tasks API (default behavior), so they are ignored.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST read `IDEA_INBOX_PATH` and `IDEA_LIST_NAME` from Alfred workflow environment variables on every list operation.
- **FR-002**: System MUST skip idea sync entirely if either `IDEA_INBOX_PATH` or `IDEA_LIST_NAME` is not set, with zero performance impact on the listing.
- **FR-003**: System MUST scan the configured Ideas list on ALL authenticated accounts during every `gt list` operation, regardless of account targeting.
- **FR-004**: System MUST deduplicate ideas by matching TaskID values against `- TaskID:` lines in the existing inbox file.
- **FR-005**: System MUST append new ideas to the inbox file using H3 headings with Date, Account, and TaskID metadata fields.
- **FR-006**: System MUST include the task notes as description text below the metadata when notes are present, and omit the description section when notes are empty.
- **FR-007**: System MUST delete each task from Google Tasks only after it has been successfully written to the inbox file.
- **FR-008**: System MUST create the inbox file with a `# Idea Inbox` header if it does not exist, including creating parent directories as needed.
- **FR-009**: System MUST NOT block, delay, or interfere with the task listing display if any part of the idea sync fails.
- **FR-010**: System MUST silently skip accounts where the Ideas list does not exist.
- **FR-011**: System MUST silently skip accounts that are not authenticated.
- **FR-012**: System MUST use the task's `updated` timestamp from the Google Tasks API for the Date metadata field, formatted as YYYY-MM-DD.

### Key Entities

- **Idea Entry**: A single idea written to the inbox file, consisting of a title (H3 heading), metadata (Date, Account, TaskID), and optional description (from task notes).
- **Ideas List**: A Google Tasks list whose name matches the configured `IDEA_LIST_NAME` variable. May or may not exist on each account.
- **Inbox File**: A markdown file at the path specified by `IDEA_INBOX_PATH` that accumulates all synced ideas.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Ideas captured via voice on Android appear in the Obsidian inbox file within one `gt list` invocation, with no manual steps required.
- **SC-002**: The Ideas list on Google Tasks is empty after a successful sync, keeping the capture surface clean for new ideas.
- **SC-003**: Task listing performance is not noticeably degraded when the Ideas list is empty or does not exist (the common case).
- **SC-004**: Zero data loss: every idea written to Google Tasks eventually appears in the Obsidian inbox, even if sync fails partway through (dedup ensures retry on next listing).
- **SC-005**: Existing users who do not configure the feature experience no behavioral changes.

## Clarifications

### Session 2026-07-13

- Q: Which Google Tasks API field to use for the Date metadata? → A: Use the `updated` timestamp (for newly created tasks this equals creation time; the API does not expose a separate `created` field). Format as YYYY-MM-DD.

## Assumptions

- The Google Tasks API returns task creation/update timestamps that can be used for the Date metadata field.
- The user's Obsidian vault is on the local filesystem accessible by the Alfred workflow.
- The `IDEA_INBOX_PATH` points to a writable location (the workflow has filesystem write permissions to that path).
- The `IDEA_LIST_NAME` is consistent across accounts (same list name on all accounts where ideas are captured).
- Completed and hidden tasks in the Ideas list are not returned by the API and are therefore not synced.
- Single-account mode (no accounts.json) also supports idea sync, using the single configured account.
