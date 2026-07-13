# Research: Task Details Display

**Date**: 2026-07-13
**Feature**: 004-task-details-display

## AwGo Largetype API

**Decision**: Use `item.Largetype(string)` method from AwGo v0.29.1
**Rationale**: Confirmed in source at `feedback.go`. Method signature is `Largetype(s string) *Item`, returns the item for chaining. Sets the text shown in Alfred's Large Text window when user presses Cmd+L.
**Alternatives considered**: Quick Look (`.Quicklookurl()`) was rejected during brainstorming as overkill for plain text notes display.

## Google Tasks API Notes Field

**Decision**: Use `Task.Notes` from the existing list API response
**Rationale**: The `tasks.list` endpoint returns full task objects including the `Notes` field by default. No field selection, additional API calls, or response parsing changes needed. The `TaskItem` struct already wraps `*tasks.Task`, so `item.Task.Notes` is directly accessible.
**Alternatives considered**: Fetching task details individually via `tasks.get` was rejected as unnecessary since list already includes notes.

## Empty Notes Handling

**Decision**: Use "(no details)" as fallback when notes are empty
**Rationale**: Pressing Cmd+L on a task with no notes should show meaningful text rather than an empty overlay. The fallback should be parenthesized to visually distinguish it from actual note content.
**Alternatives considered**: Showing the task title as fallback was rejected during brainstorming (metadata already visible in subtitle). Showing nothing was rejected (poor UX with blank overlay).
