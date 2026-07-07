# Research: Multi-Account Support

## R1: Account Prefix Parsing Strategy

**Decision**: Parse `@accountname` as the first token after the Alfred keyword in the script filter query string. The input arrives as a single string argument to `./gtasks`, so parsing happens in the Go binary's `route()` or `handleFilter()` method before any command dispatch.

**Rationale**: The `@` prefix must be extracted before command routing because it affects which credentials/tokens are used. The existing `handleFilter()` method already does query prefix matching (`"login"`, `"add "`, `"list "`). Adding `@account` extraction as a pre-processing step before the switch statement is the cleanest approach.

**Alternatives considered**:
- Alfred-level parsing (via workflow variables): rejected because Alfred's Script Filter passes the entire query as one argument, and we need the Go binary to resolve the account context for API calls.
- Regex-based parsing: unnecessarily complex. Simple `strings.HasPrefix(token, "@")` on the first token is sufficient.

## R2: Configuration File Location and Loading

**Decision**: Load `accounts.json` from `Workflow.DataDir` (the Alfred workflow data directory, overridable via `CREDENTIALS_DIR` env var). Parse it into a typed Go struct on each invocation.

**Rationale**: The workflow data directory is the established location for credentials and tokens. Adding `accounts.json` there is consistent. Since Alfred workflows are short-lived processes (invoked per keystroke), there's no need for caching or file-watching.

**Alternatives considered**:
- Alfred workflow variables: rejected per brainstorm analysis (inflexible, fixed number of accounts).
- YAML config: rejected because the project has no YAML dependency and JSON is sufficient.

## R3: Token Storage Layout for Multiple Accounts

**Decision**: Each account stores its tokens in a subdirectory of `DataDir` named after the account's symbolic name. For example, `DataDir/work/token.json`. The credentials file path is specified in `accounts.json` and can be relative to `DataDir` or absolute.

**Rationale**: The existing `auth.LoadToken()`, `auth.SaveToken()`, and `auth.TokenExists()` functions all accept a `dataDir` parameter. By passing `DataDir/accountName` instead of `DataDir`, multi-account token isolation works with zero changes to the token management functions.

**Alternatives considered**:
- Prefixed token filenames (e.g., `token_work.json`): rejected because it requires modifying the auth functions' file naming logic.
- Separate directories outside workflow data: rejected because it breaks the self-contained data directory model.

## R4: Google Multi-Login URL Format

**Decision**: Use `https://tasks.google.com/embed/?authuser=N` where N is the Google profile index (0-based). The profile index is an optional field in `accounts.json` (`"authuser": 1`).

**Rationale**: Google's multi-login system uses the `authuser` query parameter to select the active account in browser sessions. This is well-established across Google services.

**Alternatives considered**:
- Email-based selection (`authuser=user@gmail.com`): works for some Google services but is less reliable than numeric index.
- No multi-login support: rejected because the brainstorm explicitly requires per-account browser opening.

## R5: Merged Listing Performance

**Decision**: Query accounts sequentially (not concurrently) for the initial implementation. Google Tasks API calls typically complete in 200-500ms each. With 2-3 accounts, total time stays well under 2 seconds.

**Rationale**: Sequential queries are simpler to implement and debug. Concurrent queries would require goroutine management, error aggregation, and careful context cancellation, adding complexity without a meaningful UX improvement for the expected 2-3 account case.

**Alternatives considered**:
- Concurrent queries with goroutines: viable optimization for a future version if users report slow merged listings with 4+ accounts, but premature for initial implementation.

## R6: Backward Compatibility Mechanism

**Decision**: Check for `accounts.json` existence at startup. If absent, construct a single implicit `AccountContext` using the existing `DataDir` paths (same `client_secret.json` and `token.json` locations). All downstream code operates on `AccountContext` regardless of mode.

**Rationale**: This avoids if/else branching throughout the codebase. The single-account path becomes a special case of multi-account with one implicit account, handled entirely at the configuration loading layer.

**Alternatives considered**:
- Feature flag/toggle: unnecessary complexity since the presence of `accounts.json` is the natural toggle.
- Separate code paths: rejected to avoid duplication and divergence.
