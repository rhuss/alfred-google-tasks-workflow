# Quickstart: Multi-Account Support

## What This Feature Does

Adds multi-account Google Tasks support to the Alfred workflow, allowing users to manage tasks across multiple Google accounts using an `@account` prefix syntax and a JSON configuration file.

## How to Test Manually

### Single-Account Mode (backward compatibility)

1. Build: `make build`
2. Ensure NO `accounts.json` exists in the workflow data directory
3. Run: `./build/gtasks "login"` (should work as before)
4. Run: `./build/gtasks "add Buy milk, tomorrow"` (should create task in default account)
5. Run: `./build/gtasks "list"` (should list tasks from default account)

### Multi-Account Mode

1. Create `accounts.json` in the workflow data directory:
   ```json
   {
     "default": "personal",
     "list_default": "all",
     "accounts": {
       "personal": {
         "credentials": "client_secret.json"
       },
       "work": {
         "credentials": "work/client_secret.json",
         "authuser": 1
       }
     }
   }
   ```

2. Place credentials files:
   - `<DataDir>/client_secret.json` (personal, shared or separate)
   - `<DataDir>/work/client_secret.json` (work, if using separate credentials)

3. Authenticate each account:
   - `./build/gtasks "login"` (authenticates default/personal)
   - `./build/gtasks "@work login"` (authenticates work)

4. Test commands:
   - `./build/gtasks "@work add Submit report, friday"` (add to work account)
   - `./build/gtasks "list"` (merged list from all accounts)
   - `./build/gtasks "@work list"` (work account only)
   - `./build/gtasks "@work open"` (open browser for work account)

### Running Tests

```bash
make test
```

## Key Files to Understand

| File | Purpose |
|------|---------|
| `internal/auth/accounts.go` | Account configuration loading, validation, resolution |
| `internal/alfred/workflow.go` | Command routing with `@account` prefix extraction |
| `internal/input/parser.go` | Input parsing (unchanged, receives cleaned input) |
| `internal/tasks/client.go` | Google Tasks API client (unchanged, receives account context) |

## Architecture Notes

- `AccountContext` is the abstraction that unifies single-account and multi-account modes
- In single-account mode (no `accounts.json`), an implicit `AccountContext` is created using the existing `DataDir` paths
- The `@account` prefix is extracted in the routing layer before any command handler runs
- All existing auth functions work unchanged because they accept a `dataDir` parameter
