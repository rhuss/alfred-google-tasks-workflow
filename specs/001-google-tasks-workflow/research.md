# Research: Alfred Google Tasks Workflow

**Date**: 2026-07-06
**Feature**: [spec.md](spec.md)

## R1: Google Tasks API v1 Integration

**Decision**: Use `google.golang.org/api/tasks/v1` Go client library.

**Rationale**: Official Google-maintained client for Go. Provides typed structs for Task, TaskList, and all CRUD operations. Auto-generated from the API discovery document, so it stays current with the API.

**Alternatives considered**:
- Raw HTTP calls to `https://www.googleapis.com/tasks/v1/`: More control but significantly more boilerplate for auth headers, JSON parsing, error handling. Not worth it when an official client exists.
- Third-party wrappers: None found that add meaningful value over the official client.

**Key API endpoints used**:
- `tasklists.list` - enumerate user's task lists
- `tasklists.insert` - create new task list
- `tasks.list` - list tasks in a task list (supports `showCompleted=false`, `dueMin`/`dueMax`)
- `tasks.insert` - create a task
- `tasks.patch` - update task (mark complete)
- `tasks.delete` - delete a task

**API constraints**:
- Due dates are RFC 3339 date-only strings (e.g., `2026-07-06T00:00:00.000Z`), always midnight UTC
- `tasks.list` returns max 100 tasks per page by default (pagination via `nextPageToken`)
- Rate limit: 50,000 requests per day per project (more than sufficient for personal use)

## R2: OAuth 2.0 with PKCE for Desktop Apps

**Decision**: Use `golang.org/x/oauth2` with manual PKCE extension and a temporary `net/http` server on `127.0.0.1`.

**Rationale**: The `golang.org/x/oauth2` package handles token exchange and refresh but does not natively support PKCE. PKCE parameters (code_verifier, code_challenge) need to be added manually to the authorization URL and token exchange request. A local HTTP server on a random port captures the redirect.

**Alternatives considered**:
- `golang.org/x/oauth2` without PKCE: Works but less secure. Google strongly recommends PKCE for desktop apps.
- `google-auth-oauthlib` (Python): Would require Python runtime, contradicts Go single-binary requirement.

**Implementation details**:
- Generate `code_verifier`: 43-128 character random string from `[A-Za-z0-9-._~]`
- Generate `code_challenge`: Base64URL(SHA256(code_verifier))
- Redirect URI: `http://127.0.0.1:<random_port>` (prefer 127.0.0.1 over localhost for firewall compatibility)
- Server timeout: 3 minutes, then clean shutdown
- Token exchange includes both `client_secret` and `code_verifier` (Google requires both for Desktop app clients)

## R3: AwGo Library for Alfred Integration

**Decision**: Use `github.com/deanishe/awgo` v0.29+ for Alfred workflow integration.

**Rationale**: Most mature Go library for Alfred workflows. Provides Script Filter JSON generation, caching, updates, environment variable access, and logging. Used by the most popular Go-based Alfred workflow (alfred-gcal, 229 stars).

**Alternatives considered**:
- `github.com/jason0x43/go-alfred`: Less feature-rich, fewer users, less documentation.
- Raw JSON output to stdout: Possible but significantly more boilerplate for item rendering, icons, modifiers, and error display.

**Key AwGo features used**:
- `aw.Workflow` - main workflow struct with Run/Warn/Fatal helpers
- `aw.Item` - Script Filter result items with title, subtitle, arg, icon
- `aw.NewItem().Subtitle().Arg().Valid()` - fluent item builder
- Environment variables: `alfred_workflow_data` for token/config storage

**Compatibility note**: AwGo was last updated in 2021 but works with Alfred 5. The API is stable and no breaking changes are expected.

## R4: Custom Date Parser Design

**Decision**: Hand-rolled parser with a finite set of recognized patterns, no external NLP library.

**Rationale**: The required date patterns are small and well-defined. A custom parser provides 100% predictable behavior, easy testing, and zero dependencies. Google Tasks only supports date-only (no times), so NLP overkill is unnecessary.

**Patterns to support**:
| Input | Interpretation |
|-------|---------------|
| `today` | Current date |
| `tomorrow` | Current date + 1 day |
| `next week` | Next Monday |
| `monday`...`sunday` | Next occurrence of that weekday (case-insensitive) |
| `YYYY-MM-DD` | Exact ISO date |
| `MM-DD` | Month-day of current year (or next year if date has passed) |

**Extraction rules**:
1. Check for `#ListName` tag first, remove it from input
2. Split on last comma to check for trailing date
3. If no comma, check for leading date token(s)
4. If no date found, entire input is the task title (no due date)
5. Remaining text after date and list extraction becomes the task title

**Edge cases**:
- Input is only a date (no title): treat as error, title is required
- Date-like words in title (e.g., "Monday meeting"): when "Monday" appears at the start with additional words, it is interpreted as a date. To prevent this, user can quote or use comma: "Monday meeting" vs "Monday, meeting" (latter creates task "meeting" due Monday)

## R5: Token Storage Strategy

**Decision**: Store tokens as a JSON file (`token.json`) in Alfred's workflow data directory.

**Rationale**: Simple, debuggable, and consistent with the most popular Alfred workflow using Google APIs (deanishe/alfred-gcal). The workflow data directory is per-workflow and persists across Alfred updates.

**Alternatives considered**:
- macOS Keychain via `security` CLI: More secure but adds subprocess overhead and complexity. Tokens are not highly sensitive (they are scoped to personal tasks and revocable). The trade-off favors simplicity.
- Encrypted file: Adds key management complexity without meaningful security benefit since the user's macOS account already protects the file.

**File layout in `alfred_workflow_data`**:
```
~/Library/Application Support/Alfred/Workflow Data/{bundleID}/
├── client_secret.json    # User-provided OAuth credentials
└── token.json            # Stored access + refresh tokens
```

## R6: Alfred Workflow Packaging

**Decision**: Ship as a `.alfredworkflow` file (zip archive) containing the compiled binary, `info.plist`, and `icon.png`.

**Rationale**: Standard Alfred workflow distribution format. Users double-click to install. The compiled Go binary means zero runtime dependencies.

**Build process**:
- `go build -o gtasks ./cmd/` produces the binary
- `Makefile` target packages binary + info.plist + icon.png into `.alfredworkflow`
- Cross-compilation not needed (macOS-only, but build for both amd64 and arm64 via universal binary or separate downloads)
