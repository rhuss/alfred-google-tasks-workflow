# Brainstorm: Alfred Google Tasks Workflow

**Date:** 2026-07-06
**Status:** spec-created

## Problem Framing

There is no well-maintained Alfred workflow for Google Tasks. The only attempt (ntawileh/GTasks) has 0 stars, 1 commit, and uses deprecated OAuth patterns. Forum threads from 2014 and 2021 show people wanting this but nobody has built it well. This is an unfilled niche in the Alfred ecosystem.

The biggest challenge is handling Google OAuth smoothly within Alfred. Research shows the proven pattern is a localhost loopback redirect with a temporary HTTP server, as used by deanishe's alfred-gcal (229 stars, Go).

## Approaches Considered

### A: Single Go Binary with AwGo + External Date Library

One compiled Go binary handles everything: OAuth server, API calls, Alfred JSON output. Uses an external NLP date parsing library (e.g. olebedev/when or tj/go-naturaldate) for natural language dates.

- Pros: Zero runtime deps, fast startup, proven pattern, full NLP date parsing ("in 3 days", etc.)
- Cons: Go's NLP date libraries are less mature than Python's, unpredictable edge cases, unnecessary complexity for a finite set of date patterns

### B: Single Go Binary with AwGo + Custom Date Parser (CHOSEN)

Same as A, but with a focused hand-rolled date parser that only handles the specific patterns needed: "today", "tomorrow", "next week", weekday names, ISO dates (YYYY-MM-DD, MM-DD).

- Pros: Predictable behavior, easy to test, no dependency on immature NLP libraries, the pattern set is small and well-defined
- Cons: Won't understand "in 3 days" or "next Tuesday afternoon", but Google Tasks only stores dates (not times) so this is fine

### C: Go Binary with Separate Auth Helper

Split into two binaries: a main workflow binary for task operations, and a separate auth helper that only handles the OAuth flow.

- Pros: Clean separation, auth reusable for other Google workflows, easier to debug OAuth independently
- Cons: Two binaries to build/distribute, more complexity for v1, harder state coordination

## Decision

**Approach B: Single Go binary with AwGo and a custom date parser.**

A single binary keeps distribution simple. The custom date parser is more reliable than an NLP library for the finite set of patterns needed. The date set ("today", "tomorrow", "next week", weekday names, ISO dates) maps exactly to what Google Tasks supports (date-only, no times).

## Key Requirements

### OAuth
- User provides their own `client_secret.json` from Google Cloud Console ("Desktop app" type)
- Localhost loopback redirect (`http://127.0.0.1:<random_port>`) with PKCE (S256)
- Tokens stored as JSON in Alfred's workflow data directory (`alfred_workflow_data`)
- Google Cloud consent screen set to "In production" to avoid 7-day refresh token expiry
- Automatic access token refresh using stored refresh token
- Scope: `https://www.googleapis.com/auth/tasks` (read-write)

### Commands
- `gt login` - initiate OAuth flow (opens browser, localhost catches callback)
- `gt add <title>[, <date>] [#<list>]` - create task with optional smart date and list tag
- `gt list [#<list>]` - show tasks grouped by timeframe, optionally filtered by list
- `gt open` - open Google Tasks web UI in default browser

### Task Creation
- Smart date extraction from title text: "today", "tomorrow", "next week", weekday names ("Monday"..."Sunday" for next occurrence), ISO dates ("2026-07-06", "07-06")
- Date can appear at start or end of title, comma-separated
- `#ListName` tag specifies target task list; auto-creates list if it doesn't exist
- Default task list used when no tag provided

### Task Listing
- Grouped by timeframe sections: Overdue, Today, This Week, Later
- Each task shows title + due date + list name (as subtitle)
- Filtered by `#ListName` tag when provided; shows all lists otherwise
- Sorted by due date within each group

### Task Actions
- Selecting a task from the list shows a sub-menu with: Complete, Open in browser, Delete

### Technology
- Language: Go
- Alfred library: AwGo (github.com/deanishe/awgo)
- Google API: google-api-go-client (tasks/v1)
- OAuth: golang.org/x/oauth2 with PKCE
- Token storage: JSON file in alfred_workflow_data
- Credentials: User-provided client_secret.json
- Date parsing: Custom parser (no external NLP library)

### Prior Art Studied
- deanishe/alfred-gcal (Go, 229 stars) - OAuth localhost pattern, AwGo, multi-account
- azai91/alfred-drive-workflow (Ruby, 220 stars) - Keychain token storage
- fniephaus/alfred-gmail (Python, 78 stars) - Keychain via alfred-workflow library
- ntawileh/GTasks (Python, 0 stars) - only existing Google Tasks workflow, abandoned

## Open Questions

- Should `gt list` cache results locally for faster display, or always fetch from API? (caching adds complexity but improves responsiveness)
- What should the default task list be if the user has multiple lists? First list returned by API, or a configurable default?
- Should `gt logout` be a command to revoke tokens and clear stored credentials?
- Should tasks without a due date be shown in the list view? If so, under which timeframe group?
