# Feature Specification: Multi-Account Support

**Feature Branch**: `002-multi-account-support`
**Created**: 2026-07-06
**Status**: Draft
**Input**: User description: "Multi-account support for Alfred Google Tasks workflow with config-driven account registry, @account prefix syntax, and merged task listing"

## User Scenarios & Testing

### User Story 1 - Single-Account Backward Compatibility (Priority: P1)

A user who has been using the workflow with a single Google account upgrades to the new version. They have no `accounts.json` file. Everything works exactly as before with no configuration changes needed.

**Why this priority**: Existing users must not be disrupted. This is the foundation of trust for the upgrade path.

**Independent Test**: Can be fully tested by running all existing commands (`gt add`, `gt list`, `gt login`, `gt open`) without an `accounts.json` file and verifying identical behavior to the current version.

**Acceptance Scenarios**:

1. **Given** no `accounts.json` exists in the workflow data directory, **When** the user types any `gt` command, **Then** the workflow behaves identically to the current single-account version
2. **Given** no `accounts.json` exists, **When** the user types `gt login`, **Then** credentials and tokens are stored in the same locations as the current version
3. **Given** no `accounts.json` exists, **When** the user types `gt @work list`, **Then** the `@work` prefix is treated as part of the search query (not as an account selector)

---

### User Story 2 - Account Configuration Setup (Priority: P1)

A user with multiple Google accounts (personal and work) wants to set up multi-account access. They create an `accounts.json` file in the workflow data directory, defining each account with a symbolic name and its own credentials subdirectory. Once configured, the workflow recognizes both accounts.

**Why this priority**: Multi-account configuration is the prerequisite for all multi-account features. Without it, nothing else in this feature works.

**Independent Test**: Can be fully tested by creating an `accounts.json` with two accounts and verifying the workflow loads and validates the configuration correctly.

**Acceptance Scenarios**:

1. **Given** the user creates a valid `accounts.json` with two accounts, **When** they run any `gt` command, **Then** the workflow loads the configuration and recognizes both accounts
2. **Given** `accounts.json` exists but has invalid JSON syntax, **When** the user runs any `gt` command, **Then** the workflow shows a clear error message identifying the syntax problem
3. **Given** `accounts.json` references a credentials file that does not exist, **When** the user runs a command targeting that account, **Then** the workflow shows an error message identifying the missing credentials file and which account is affected
4. **Given** `accounts.json` does not specify a `default` field, **When** the user runs `gt add Buy milk`, **Then** the workflow uses the first account defined in the config as the default
5. **Given** `accounts.json` specifies a `default` that does not match any account name, **When** the user runs any `gt` command, **Then** the workflow shows an error message identifying the invalid default

---

### User Story 3 - Account-Targeted Task Creation with @ Prefix (Priority: P1)

A user with multiple accounts configured wants to add a task to a specific account. They use the `@accountname` prefix before or after the command to target that account.

**Why this priority**: The `@` prefix is the primary interaction model for multi-account usage. It must work reliably with task creation, the most common operation.

**Independent Test**: Can be fully tested by creating tasks with `@account` prefixes and verifying they appear in the correct Google Tasks account.

**Acceptance Scenarios**:

1. **Given** accounts "personal" and "work" are configured, **When** the user types `gt @work add Submit report, friday`, **Then** a task "Submit report" is created in the "work" account with Friday's date
2. **Given** accounts are configured with "personal" as default, **When** the user types `gt add Buy milk`, **Then** the task is created in the "personal" account
3. **Given** accounts are configured, **When** the user types `gt @nonexistent add Something`, **Then** the workflow shows an error listing the valid account names
4. **Given** accounts are configured, **When** the user types `gt @work add Buy groceries, tomorrow #Shopping`, **Then** the task is created in the "work" account with the correct date and list, and the `@work` prefix is fully stripped from the input before parsing

---

### User Story 4 - Multi-Account Task Listing (Priority: P1)

A user with multiple accounts wants to view tasks across accounts. Depending on the `list_default` setting, the default `gt list` command shows either merged tasks from all accounts or only the default account's tasks. The user can always target a specific account with the `@` prefix.

**Why this priority**: Viewing tasks across accounts is the core value proposition of multi-account support.

**Independent Test**: Can be fully tested by creating tasks in multiple accounts and verifying listing behavior with different `list_default` settings and `@` prefixes.

**Acceptance Scenarios**:

1. **Given** `list_default` is `"all"` and tasks exist in both accounts, **When** the user types `gt list`, **Then** tasks from all accounts are merged into a single list, each task's subtitle includes the account name in parentheses
2. **Given** `list_default` is `"default"` and "personal" is the default account, **When** the user types `gt list`, **Then** only tasks from the "personal" account are shown
3. **Given** any `list_default` setting, **When** the user types `gt @work list`, **Then** only tasks from the "work" account are shown
4. **Given** `list_default` is `"all"`, **When** tasks from multiple accounts are displayed, **Then** tasks are still grouped by timeframe (Overdue, Today, This Week, Later) with account name visible in each task's subtitle
5. **Given** `list_default` is `"all"` and one account has expired credentials, **When** the user types `gt list`, **Then** tasks from the authenticated account are shown and a warning indicates the other account needs re-authentication

---

### User Story 5 - Per-Account Authentication (Priority: P1)

A user with multiple accounts configured needs to authenticate each account independently. The login command accepts an `@` prefix to target a specific account, and each account stores its own tokens in a separate subdirectory.

**Why this priority**: Authentication is the gateway to all account operations. Each account must be independently authenticable.

**Independent Test**: Can be fully tested by logging into each account separately and verifying tokens are stored in the correct subdirectories.

**Acceptance Scenarios**:

1. **Given** accounts are configured, **When** the user types `gt login`, **Then** the default account's OAuth flow starts
2. **Given** accounts are configured, **When** the user types `gt @work login`, **Then** the "work" account's OAuth flow starts using that account's credentials file
3. **Given** the "work" account's OAuth flow completes, **When** tokens are stored, **Then** they are saved in the "work" account's subdirectory, not the default location
4. **Given** the "personal" account is authenticated but "work" is not, **When** the user types `gt @work add Something`, **Then** the workflow shows a message prompting them to run `gt @work login`
5. **Given** accounts are configured, **When** the user types `gt @work logout`, **Then** only the "work" account's tokens are removed

---

### User Story 6 - Account-Targeted Browser Opening (Priority: P2)

A user wants to open Google Tasks in the browser for a specific account. The workflow uses Google's multi-login URL format to open the correct account's task view.

**Why this priority**: Convenient but not essential. Users can navigate to Google Tasks manually.

**Independent Test**: Can be fully tested by running the open command with different `@` prefixes and verifying the correct URL is opened.

**Acceptance Scenarios**:

1. **Given** accounts are configured, **When** the user types `gt @work open`, **Then** the browser opens Google Tasks with the work account's profile selected
2. **Given** accounts are configured, **When** the user types `gt open`, **Then** the browser opens Google Tasks with the default account's profile
3. **Given** the targeted account has a Google profile index configured, **When** the browser opens, **Then** the URL includes the correct `authuser` parameter

---

### User Story 7 - Per-Account Keywords (Priority: P3)

A user wants a dedicated Alfred keyword for their work account so they can type `gtw list` instead of `gt @work list`. They configure an optional `keyword` field in the account's config entry.

**Why this priority**: This is a power-user convenience feature. The `@` prefix already provides full account targeting.

**Independent Test**: Can be fully tested by configuring a keyword for an account and verifying that keyword routes all commands to that account.

**Acceptance Scenarios**:

1. **Given** the "work" account has `"keyword": "gtw"` configured, **When** the user types `gtw list`, **Then** tasks from the "work" account are listed
2. **Given** the "work" account has a keyword configured, **When** the user types `gtw add Submit report`, **Then** the task is created in the "work" account
3. **Given** the "work" account has a keyword configured, **When** the user types `gtw @personal list`, **Then** the `@personal` prefix is ignored (keyword already identifies the account) and "work" account tasks are shown

---

### Edge Cases

- What happens when `accounts.json` is deleted while the workflow is running? The workflow MUST fall back to single-account mode on the next invocation.
- What happens when two accounts have overlapping task list names (e.g., both have a "Work" list)? In merged view, tasks show the account name to disambiguate.
- What happens when `@` appears as part of a task title (e.g., `gt add Email @john about meeting`)? The `@` prefix is only recognized immediately after the `gt` keyword or after `gt [command]`, not within task content. The pattern is `gt @account command args` where `@account` must be the token immediately after `gt`.
- What happens when the user types just `gt @work` with no command? The workflow shows the default action menu for the "work" account (same as typing `gt` but scoped to that account).

## Clarifications

### Session 2026-07-06

- Q: Can accounts share a single OAuth credentials file, or must each have its own? → A: Accounts can share a single credentials file. The `credentials` path in each account entry can point to the same file. Each account still gets its own token subdirectory.
- Q: How should merged listing handle API failures for one account (timeout vs auth failure)? → A: Show tasks from successful accounts and append a warning item for the failed account, distinguishing between "needs re-authentication" and "temporarily unavailable."
- Q: What is the maximum number of accounts supported? → A: No hard limit enforced, but tested and documented for up to 5 accounts. Performance degrades linearly with account count due to sequential API calls.

## Requirements

### Functional Requirements

- **FR-001**: System MUST load account configuration from `accounts.json` in the workflow data directory when the file exists
- **FR-002**: System MUST fall back to single-account behavior when `accounts.json` does not exist, preserving full backward compatibility
- **FR-003**: System MUST parse `@accountname` prefix from user input and route the command to the specified account
- **FR-004**: System MUST strip the `@accountname` prefix from input before passing it to the command handler
- **FR-005**: System MUST use the `default` account when no `@` prefix is specified and `accounts.json` exists
- **FR-006**: System MUST store each account's OAuth tokens in a separate subdirectory within the workflow data directory
- **FR-007**: System MUST support a `list_default` setting with values `"all"` (merge tasks from all accounts) and `"default"` (show only default account's tasks). When `list_default` is omitted from `accounts.json`, it MUST default to `"default"`
- **FR-008**: System MUST display the account name in each task's subtitle when showing merged results from multiple accounts
- **FR-009**: System MUST validate `accounts.json` on load and provide clear error messages for invalid configuration (bad JSON, missing credentials, invalid default)
- **FR-010**: System MUST support per-account `keyword` configuration for dedicated Alfred Script Filter entries
- **FR-011**: System MUST use Google's multi-login URL format (`authuser` parameter) when opening the browser for a specific account
- **FR-012**: System MUST authenticate each account independently, using each account's own credentials file and token storage
- **FR-013**: System MUST support per-account logout by removing only the targeted account's stored tokens without affecting other accounts

### Key Entities

- **Account**: A named Google account configuration with a symbolic name, credentials path, token storage location, and optional keyword. Has one default flag per configuration.
- **AccountConfig**: The top-level configuration containing the default account name, list_default behavior, and a map of account name to account definition.
- **AccountContext**: The resolved runtime context for a command, including the target account's API client, credentials, and token paths.

### Configuration Schema

The `accounts.json` file uses the following structure:

```json
{
  "default": "personal",
  "list_default": "default",
  "accounts": {
    "personal": {
      "credentials": "credentials/client_secret.json",
      "token_dir": "tokens/personal"
    },
    "work": {
      "credentials": "credentials/client_secret.json",
      "token_dir": "tokens/work",
      "keyword": "gtw",
      "profile_index": 1
    }
  }
}
```

| Field | Type | Required | Description |
|---|---|---|---|
| `default` | string | No | Name of the default account. If omitted, the first account in the map is used. |
| `list_default` | string | No | `"default"` (show only default account) or `"all"` (merge all accounts). Defaults to `"default"`. |
| `accounts` | map | Yes | Map of account name to account configuration. |
| `accounts.<name>.credentials` | string | Yes | Path to the OAuth credentials file, relative to workflow data directory. |
| `accounts.<name>.token_dir` | string | Yes | Directory for storing OAuth tokens, relative to workflow data directory. |
| `accounts.<name>.keyword` | string | No | Dedicated Alfred keyword for this account (P3 feature). |
| `accounts.<name>.profile_index` | integer | No | Google multi-login profile index for browser `authuser` parameter. |

## Success Criteria

### Measurable Outcomes

- **SC-001**: Users with a single account experience zero behavior changes after upgrading (100% backward compatibility)
- **SC-002**: Users can switch between accounts in under 2 seconds using the `@` prefix, with no additional keystrokes beyond the account name
- **SC-003**: Merged task listing from 2 accounts completes and displays within the same timeframe as the current single-account listing (within 1 second)
- **SC-004**: Each account's authentication is fully independent, allowing users to log in, log out, and re-authenticate any account without affecting other accounts
- **SC-005**: Configuration errors (invalid JSON, missing credentials, bad defaults) produce actionable error messages that identify the specific problem and suggest a corrective action (e.g., "Account 'work' references missing credentials file at path/to/file. Create the file or update accounts.json.")

## Assumptions

- Users may share a single Google Cloud OAuth credentials file (`client_secret.json`) across multiple accounts or use separate files per account; each account's `credentials` path in `accounts.json` can reference the same file
- Performance scales linearly with account count due to sequential API calls; the feature is tested and documented for up to 5 accounts
- The Alfred workflow data directory is writable and has sufficient space for additional token files (one per account)
- Users understand Google's multi-account system and can select the correct Google account during the OAuth consent flow
- The `@` character is not commonly used as the first character in task titles, making it safe to use as the account selector prefix
- The separate keyword feature (P3) requires users to manually add Script Filter entries in Alfred's workflow configuration, as programmatic modification of `info.plist` is fragile and version-dependent
- Google Tasks API rate limits are sufficient for querying multiple accounts in sequence for merged listing

## Out of Scope

- **Cross-account task moving**: Moving or copying tasks between accounts is not supported
- **Per-account task list selection**: Selecting different default task lists per account (all accounts use the same list selection logic)
- **Parallel API calls**: API calls to multiple accounts are sequential; concurrent requests are a potential future optimization
- **Programmatic Alfred keyword setup**: Automatically modifying `info.plist` to register per-account keywords; users must add Script Filter entries manually
- **Account discovery**: Automatically detecting Google accounts on the system; accounts must be manually configured in `accounts.json`
