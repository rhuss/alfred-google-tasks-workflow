# Implementation Plan: Multi-Account Support

**Branch**: `002-multi-account-support` | **Date**: 2026-07-06 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `specs/002-multi-account-support/spec.md`

## Summary

Add multi-account Google Tasks support via a config-driven account registry (`accounts.json`). Users target accounts with an `@accountname` prefix (e.g., `gt @work add Report`). The feature preserves full backward compatibility: without `accounts.json`, the workflow behaves identically to the current version. Key technical approach: introduce an `AccountContext` abstraction that unifies single- and multi-account modes, extracted in the routing layer before command dispatch.

## Technical Context

**Language/Version**: Go 1.26.4
**Primary Dependencies**: awgo v0.29.1 (Alfred framework), golang.org/x/oauth2, google.golang.org/api (Tasks v1)
**Storage**: JSON files on disk (credentials, tokens, account config)
**Testing**: `go test ./...` via `make test`
**Target Platform**: macOS (Alfred 5 workflow)
**Project Type**: CLI binary (Alfred Script Filter)
**Performance Goals**: All commands complete within 1 second for up to 5 accounts in merged listing
**Constraints**: Short-lived process (invoked per keystroke), no persistent state beyond files
**Scale/Scope**: 2-5 Google accounts per user

## Constitution Check

*Constitution is not customized (template only). No gates to evaluate.*

## Project Structure

### Documentation (this feature)

```text
specs/002-multi-account-support/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
└── tasks.md             # Phase 2 output (created by /speckit.tasks)
```

### Source Code (repository root)

```text
cmd/
└── main.go                    # Entry point (unchanged)

internal/
├── auth/
│   ├── accounts.go            # NEW: AccountConfig, Account, AccountContext types and loading
│   ├── config.go              # OAuthConfig struct (unchanged)
│   ├── credentials.go         # LoadClientCredentials (unchanged) + NEW LoadClientCredentialsFrom(path)
│   ├── exec.go                # OAuth flow execution (unchanged)
│   ├── oauth.go               # OAuth helpers (unchanged)
│   └── token.go               # Token CRUD (unchanged, accepts dataDir)
├── alfred/
│   ├── items.go               # Alfred item rendering (minor changes for account name in subtitle)
│   ├── notifications.go       # Notification helpers (unchanged)
│   └── workflow.go            # Command routing (major changes: @prefix extraction, AccountContext threading)
├── input/
│   └── parser.go              # Input parsing (unchanged, receives cleaned input)
├── tasks/
│   ├── actions.go             # Task actions (unchanged)
│   ├── client.go              # Tasks API client (unchanged)
│   ├── create.go              # Task creation (unchanged)
│   ├── create_test.go         # Creation tests (unchanged)
│   ├── list.go                # Task listing (minor changes: account name in TaskItem)
│   └── list_test.go           # Listing tests (updated for account name)
└── dateparse/
    ├── dateparse.go           # Date parsing (unchanged)
    └── dateparse_test.go      # Date tests (unchanged)
```

**Structure Decision**: The existing flat package structure is maintained. The only new file is `internal/auth/accounts.go` for account configuration logic. All other changes are modifications to existing files, primarily `workflow.go` (routing) and `items.go`/`list.go` (display).

## Global Constraints

These apply to every task implicitly:

- **Go version**: 1.26.4
- **Alfred framework**: awgo v0.29.1
- **Platform**: macOS (Alfred 5 workflow)
- **Process model**: Short-lived (invoked per keystroke), no persistent state beyond files
- **Performance**: All commands complete within 1 second for up to 5 accounts in merged listing
- **Backward compatibility**: Without `accounts.json`, all behavior must be identical to the current single-account version

## Implementation Strategy

### Layer 1: Account Configuration (internal/auth/accounts.go)

New file containing:
- `AccountConfig`, `Account`, `AccountContext` struct types
- `LoadAccountConfig(dataDir string)` loads and validates `accounts.json`
- `ResolveAccount(config *AccountConfig, accountName string)` resolves an account name to an `AccountContext`
- `DefaultContext(dataDir string)` creates an implicit single-account `AccountContext` for backward compatibility

The key design principle: when `accounts.json` does not exist, `DefaultContext()` returns an `AccountContext` that points to the same paths as the current code. This makes multi-account a superset of single-account with zero behavioral changes.

### Layer 2: Input Prefix Extraction (internal/alfred/workflow.go)

Modify `handleFilter()` and `route()` to:
1. Check if `accounts.json` exists (determines mode)
2. If multi-account mode, extract `@accountname` from the first token
3. Resolve the account to an `AccountContext`
4. Pass the `AccountContext` to all command handlers (replacing the current `w.DataDir` usage)

The `@` prefix is ONLY recognized in multi-account mode. In single-account mode, `@` is treated as a regular character (preserving backward compatibility per User Story 1, Scenario 3).

### Layer 3: Command Handler Updates (internal/alfred/workflow.go)

The `Workflow` struct gains an `AccountContext` field (set during initialization or after `@` prefix extraction). Each handler (`handleLogin`, `handleLogout`, `handleAdd`, `handleList`, `handleOpen`, `handleAction`) reads from `w.AccountCtx` instead of using `w.DataDir` directly. Handler method signatures remain unchanged; only the data source shifts from `w.DataDir` to `w.AccountCtx.DataDir`.

### Layer 4: Merged Listing (internal/alfred/workflow.go + items.go)

When `list_default` is `"all"` and no `@` prefix targets a specific account:
1. Iterate over all accounts in config
2. For each authenticated account, create an API client and fetch tasks
3. Tag each `TaskItem` with the account name
4. Merge all task items, then apply the existing timeframe grouping
5. Display account name in subtitle (e.g., "Work list (personal)")

For accounts with auth failures, append a warning item instead of failing entirely.

### Layer 5: Browser Opening (internal/tasks/ or inline)

Modify `handleOpen()` to use the `AccountContext.Authuser` field when constructing the Google Tasks URL. The URL format: `https://tasks.google.com/embed/?authuser=N`.

## Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| `@` conflicts with task titles | Low | Low | Only recognize `@` as first token, not within content |
| Merged listing too slow with 4+ accounts | Low | Medium | Sequential queries; optimize to concurrent later if needed |
| Google OAuth confusion (wrong account selected in consent screen) | Medium | Medium | Documentation guidance; `prompt=select_account` parameter |
| Breaking existing single-account users | Low | Critical | Extensive backward compatibility tests; no-config = no-change |
