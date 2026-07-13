# Research: Move Task Between Accounts

**Date**: 2026-07-13

## Google Tasks API: Cross-Account Move

**Decision**: Implement as create-on-target + delete-from-source.
**Rationale**: Google Tasks API has no cross-account move operation. Each account is a separate API scope with its own OAuth token.
**Alternatives**: None viable. The API only supports operations within a single authenticated scope.

## Task Fields to Preserve

**Decision**: Copy `Title`, `Due`, and `Notes` fields.
**Rationale**: These are the user-visible fields that matter for task context. Other fields like `Id`, `Updated`, `SelfLink`, `Etag`, `Status` are account-specific metadata that should not be copied.
**Alternatives**: Could also copy `Links` (web links attached to a task), but these are rarely used and can be deferred to a later version.

## Action Argument Encoding

**Decision**: Use `move:{targetAccount}|{listID}:{taskID}` format.
**Rationale**: The existing action format is `{action}|{listID}:{taskID}`. Adding a compound action prefix `move:{targetAccount}` fits the existing pattern while carrying the target account name. The `executeAction` dispatcher already splits on `|` and handles action-specific logic.
**Alternatives**: Could use a separate Alfred variable for the target account, but the arg-based approach is self-contained and does not require additional wiring.

## Error Handling Strategy

**Decision**: Differentiate between create failure and delete failure.
**Rationale**: If create fails, the move is cleanly aborted (original untouched). If delete fails after successful create, the user ends up with a duplicate, which is better than losing the task, but warrants a warning.
**Alternatives**: Could attempt to roll back (delete the created task on failure to delete original), but this adds complexity and risk of losing the task entirely.
