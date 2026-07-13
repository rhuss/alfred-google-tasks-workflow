# Brainstorm: Move Task Between Accounts

**Date:** 2026-07-13
**Status:** active

## Problem Framing

When using multiple Google accounts (e.g., work and personal), tasks sometimes need to be moved from one account to another. Currently there is no way to do this from the workflow. The user would have to manually recreate the task on the target account and delete it from the source.

## Approaches Considered

### A: Action menu entries (Chosen)

Add "Move to {account}" items directly in the existing action menu (`RenderActionMenu`). In multi-account mode, append one entry per other account (excluding the current one). The move action performs create-on-target + delete-from-source.

- Pros: Simple UX (one click), consistent with existing action pattern, no new commands or Alfred wiring needed
- Cons: Action menu grows with number of accounts (but realistically 2-3 accounts max)

### B: Sub-menu approach

Add a single "Move to..." entry that opens a second-level menu listing target accounts.

- Pros: Cleaner action menu when many accounts exist
- Cons: Extra interaction step, more complex Alfred wiring (nested script filters), overkill for 2-3 accounts

### C: Command-based move

Add a `gt move` command that takes `@target` as argument.

- Pros: Composable with other commands
- Cons: Poor UX (user has to type task references), breaks the existing select-then-act flow

## Decision

Approach A: direct action menu entries. The action menu currently shows Complete, Delete, Open in Browser. In multi-account mode, it will additionally show "Move to {account}" for each configured account except the one the task belongs to.

## Key Requirements

- Move operation is create-on-target + delete-from-source (Google Tasks API has no cross-account move)
- Target list: same-named list on the target account, auto-created if it doesn't exist
- Only show move entries in multi-account mode (2+ accounts configured)
- Exclude the task's current account from the move targets
- Preserve task title, due date, and notes during the move
- Show success/failure notification after the move

## Open Questions

- Should the move preserve subtasks (if any), or only move the top-level task?
- Error handling: if create succeeds but delete fails, should the user be warned about the duplicate?
