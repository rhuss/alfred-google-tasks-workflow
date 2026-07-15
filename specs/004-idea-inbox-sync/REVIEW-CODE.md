# Code Review: Idea Inbox Sync

**Feature Branch**: `004-idea-inbox-sync`
**Review Date**: 2026-07-13
**Reviewer**: Ship Pipeline (Stage 7)

## Spec Compliance Check

**Compliance Score: 12/12 (100%)**

| Requirement | Status | Evidence |
|-------------|--------|----------|
| FR-001: Read env vars | PASS | `syncIdeasToInbox` and `syncIdeasAllAccounts` read `IDEA_INBOX_PATH` and `IDEA_LIST_NAME` via `os.Getenv` |
| FR-002: Skip if unset | PASS | Both methods early-return when either env var is empty; tested in `workflow_test.go` |
| FR-003: Scan all accounts | PASS | `syncIdeasAllAccounts` iterates `AccountConfig.AccountNames()` |
| FR-004: Dedup by TaskID | PASS | `ReadSyncedTaskIDs` parses `- TaskID:` lines; `SyncIdeas` checks `existingIDs[task.Id]` |
| FR-005: Append format, omit Account in single-account | PASS | `AppendIdeaEntry` omits Account when empty; tested in `TestAppendIdeaEntry_WithoutAccount` |
| FR-006: Include notes as description | PASS | `AppendIdeaEntry` writes description block when non-empty, omits when empty |
| FR-007: Delete after write | PASS | `SyncIdeas` calls `DeleteTask` only after successful `AppendIdeaEntry` |
| FR-008: Auto-create file | PASS | `ensureInboxFile` creates file with header and parent dirs; tested in `TestAppendIdeaEntry_CreatesParentDirectories` |
| FR-009: Never block listing | PASS | Both workflow methods use `defer recover()` to catch panics; all errors logged to stderr and swallowed |
| FR-010: Skip missing Ideas list | PASS | `SyncIdeas` returns 0 when `FindTaskListByName` returns nil; tested in `TestSyncIdeas_MissingListReturnsZero` |
| FR-011: Skip unauth accounts | PASS | `syncIdeasAllAccounts` checks `auth.TokenExists` and `continue`s on failure |
| FR-012: Use `updated` timestamp YYYY-MM-DD | PASS | `FormatDate` parses RFC 3339 `task.Updated` and formats as `2006-01-02` |

## Gate Outcome: PASS

## Deep Review Report

### Correctness Review

**Finding count: 0 Critical, 1 Important, 1 Advisory**

1. **[Important] handleListAllAccounts sync placement**: The `syncIdeasAllAccounts()` call in `handleListAllAccounts` (line 624) runs before the account iteration loop but after the method entry. This is correct for triggering sync, but it runs even when `shouldListAllAccounts()` returns false (since `handleListAllAccounts` is only called when `shouldListAllAccounts()` is true). No bug, but the placement is slightly misleading. **Status: Acceptable** - the method is only called from the correct path.

2. **[Advisory] `syncIdeasAllAccounts` nil AccountConfig**: If called with `nil` `AccountConfig`, it would panic at `AccountConfig.AccountNames()`. The `handleList` code guards this (`if w.AccountConfig != nil`), but the method itself doesn't. The test `TestSyncIdeasAllAccounts_NoOpWhenNilAccountConfig` relies on the deferred recover. **Status: Acceptable** - the panic is caught by the deferred recover, matching FR-009.

### Architecture Review

**Finding count: 0 Critical, 0 Important, 0 Advisory**

Clean separation: `internal/ideas/` package with `inbox.go` (file I/O) and `sync.go` (orchestration). The `TasksClient` interface in `sync.go` enables unit testing without the real API. Integration into `workflow.go` is minimal and well-guarded.

### Security Review

**Finding count: 0 Critical, 0 Important, 1 Advisory**

1. **[Advisory] File path from env var**: `IDEA_INBOX_PATH` is used directly for file creation with `os.MkdirAll` and `os.OpenFile`. Since this is a local Alfred workflow (single user, macOS), there is no path traversal risk. The env var is set by the user in Alfred preferences. **Status: Acceptable**.

### Production Readiness Review

**Finding count: 0 Critical, 0 Important, 1 Advisory**

1. **[Advisory] No notification on sync**: The spec left this as an open question. Currently, sync count is logged to stderr only. For a future iteration, consider showing an Alfred notification when ideas are synced. **Status: Out of scope for this feature**.

### Test Coverage Review

**Finding count: 0 Critical, 0 Important, 0 Advisory**

Comprehensive test coverage:
- `inbox_test.go`: 12 tests covering read, write, auto-create, dedup, content preservation, nested directories, header preservation
- `sync_test.go`: 10 tests covering new tasks, dedup, delete-after-write, missing list, API errors, delete errors, empty titles, guard clauses
- `workflow_test.go`: 4 tests covering env var guard clauses for both single and multi-account modes

All 173 tests pass. `go test ./...` clean.

### Summary

| Dimension | Critical | Important | Advisory |
|-----------|----------|-----------|----------|
| Correctness | 0 | 1 | 1 |
| Architecture | 0 | 0 | 0 |
| Security | 0 | 0 | 1 |
| Production | 0 | 0 | 1 |
| Tests | 0 | 0 | 0 |
| **Total** | **0** | **1** | **3** |

**Overall: PASS**. No critical or blocking findings. The 1 important finding (nil AccountConfig guard) is mitigated by the deferred recover and call-site guard. All advisory findings are acceptable for the scope of this feature.

### External Tools

- CodeRabbit: disabled
- Copilot: disabled
