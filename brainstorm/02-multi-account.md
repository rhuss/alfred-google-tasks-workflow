# Brainstorm: Multi-Account Support

**Date:** 2026-07-06
**Status:** active

## Problem Framing

Users often have multiple Google accounts (personal, work) and want to manage tasks across both from Alfred. The current workflow supports only a single account. Adding multi-account support requires thoughtful UX so that switching between accounts is fast and intuitive without adding friction to the common single-account case.

## Approaches Considered

### A: Config-Driven Account Registry (CHOSEN)

A JSON config file (`accounts.json`) in the workflow data directory defines accounts with symbolic names, each with its own credentials and token subdirectory. Supports `@account` prefix syntax for targeting, separate Alfred keywords per account, and configurable default/listing behavior.

- Pros: Flexible, user-defined names, single config file, supports both `@` and keyword modes, backward compatible (no config = single account)
- Cons: More complex config setup, need to manage multiple credential directories

### B: Implicit Account Detection

No config file. Subdirectories in the workflow data dir auto-become accounts, named after the directory.

- Pros: Zero config
- Cons: No control over defaults, no symbolic names, directory naming is the only UX

### C: Alfred Workflow Variables Only

Multiple text fields in Alfred's Configure Workflow UI for each account.

- Pros: Native Alfred UI
- Cons: Fixed number of accounts, clunky with many fields, inflexible

## Decision

**Approach A: Config-driven account registry.** Provides the best balance of flexibility and UX. Single-account users are unaffected (no config file needed, existing behavior preserved).

## Key Requirements

### Account Configuration
- `accounts.json` in the workflow data directory defines all accounts
- Each account has a symbolic name (e.g., "personal", "work") and its own credentials subdirectory
- `default` field sets which account is used when no `@` selector is specified
- `list_default` field: `"all"` shows tasks from all accounts merged, `"default"` shows only the default account
- Backward compatible: no `accounts.json` = single-account mode (existing behavior unchanged)

### @Account Prefix Syntax
- `gt @work add Submit report, friday` targets the "work" account
- `gt @personal list` lists tasks from the "personal" account
- `gt @work login` authenticates the work account
- `gt add Buy milk` (no prefix) uses the default account
- `@` prefix is parsed and stripped before passing to the command handler

### Separate Keyword Support
- Each account can optionally define a keyword (e.g., `"keyword": "gtw"`)
- Requires additional Script Filter entries in info.plist
- When using a dedicated keyword, `@` syntax is ignored (the keyword already identifies the account)

### Task Listing
- `gt list` behavior depends on `list_default` setting
- When `list_default` is `"all"`: tasks from all accounts are merged, each task's subtitle includes the account name
- When `list_default` is `"default"`: only default account tasks shown
- `gt @work list` always filters to that specific account regardless of `list_default`

### Browser Opening
- `gt @work open` opens Google Tasks in the browser using the work account's Google profile
- Use Google's multi-login URL format to target the correct account

### Authentication
- `gt login` authenticates the default account (or single account if no config)
- `gt @work login` authenticates a specific account
- Each account stores its own `token.json` in its subdirectory
- Account selector (`prompt=select_account`) ensures user picks the right Google account

### Config File Structure
```json
{
  "default": "personal",
  "list_default": "all",
  "accounts": {
    "personal": {
      "credentials": "personal/client_secret.json"
    },
    "work": {
      "credentials": "work/client_secret.json",
      "keyword": "gtw"
    }
  }
}
```

## Open Questions

- Should `gt accounts` be a command that lists configured accounts and their auth status?
- How to handle the case where `accounts.json` exists but an account's credentials are missing (error per-account or skip silently)?
- Should the merged list view visually distinguish accounts with different colors or just text labels?
- For the separate keyword feature, should the workflow auto-generate the info.plist entries or require manual setup?
