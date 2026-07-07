# Data Model: Multi-Account Support

## Entities

### AccountConfig (top-level configuration)

Represents the parsed `accounts.json` file.

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| Default | string | No | Symbolic name of the default account. If omitted, uses the first account in the map. |
| ListDefault | string | No | Controls `gt list` behavior: `"all"` (merge all accounts) or `"default"` (default account only). Defaults to `"default"`. |
| Accounts | map[string]Account | Yes | Map of symbolic name to account definition. At least one entry required. |

### Account (per-account definition)

Represents a single account entry within the `accounts` map.

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| Credentials | string | Yes | Path to `client_secret.json`, relative to DataDir or absolute. |
| Keyword | string | No | Optional dedicated Alfred keyword for this account (e.g., `"gtw"`). |
| Authuser | int | No | Google profile index for browser multi-login (0-based). Defaults to 0. |

### AccountContext (runtime resolved context)

Not persisted. Created at the start of each command invocation by resolving the target account from the configuration and user input.

| Field | Type | Description |
|-------|------|-------------|
| Name | string | Symbolic name of the resolved account (e.g., "work"). Empty string for implicit single-account mode. |
| DataDir | string | Resolved data directory for this account's tokens (e.g., `DataDir/work/`). |
| CredentialsPath | string | Resolved absolute path to the credentials file. |
| Authuser | int | Google profile index for browser URLs. |

## Relationships

```
AccountConfig
  ├── Default → points to one Account by name
  ├── ListDefault → controls listing behavior
  └── Accounts (1:N)
       └── Account
            └── resolves to → AccountContext (at runtime)
```

## File Layout (on disk)

### Single-account mode (no accounts.json)

```
<DataDir>/
├── client_secret.json
└── token.json
```

### Multi-account mode (accounts.json present)

```
<DataDir>/
├── accounts.json
├── personal/
│   ├── client_secret.json   (or shared via path reference)
│   └── token.json
└── work/
    ├── client_secret.json   (or shared via path reference)
    └── token.json
```

## Validation Rules

1. `accounts.json` must be valid JSON conforming to the AccountConfig schema
2. If `default` is specified, it must match a key in the `accounts` map
3. Each account's `credentials` path must resolve to an existing file
4. Account names must be non-empty strings containing only alphanumeric characters, hyphens, and underscores
5. If `keyword` is specified, it must be unique across all accounts
6. `list_default` must be either `"all"` or `"default"` (if specified)

## State Transitions

Accounts do not have lifecycle states. The relevant state is per-account authentication status:

| State | Condition | Transitions |
|-------|-----------|-------------|
| Unauthenticated | No `token.json` in account's data subdirectory | `gt @name login` -> Authenticated |
| Authenticated | Valid `token.json` exists | Token expiry -> Token Refresh (automatic) |
| Token Refresh | Access token expired, refresh token valid | Success -> Authenticated, Failure -> Expired |
| Expired | Refresh failed (invalid_grant) | `gt @name login` -> Authenticated, `gt @name logout` -> Unauthenticated |
