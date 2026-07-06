# Implementation Plan: Alfred Google Tasks Workflow

**Branch**: `001-google-tasks-workflow` | **Date**: 2026-07-06 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `specs/001-google-tasks-workflow/spec.md`

## Summary

Build an Alfred workflow for managing Google Tasks, implemented as a single compiled Go binary using the AwGo library. The workflow provides commands for OAuth authentication (`gt login`/`gt logout`), task creation with smart date parsing (`gt add`), task listing grouped by timeframe (`gt list`), task actions (complete/delete/open), and opening the Google Tasks web UI (`gt open`). OAuth uses the localhost loopback redirect pattern with PKCE.

## Technical Context

**Language/Version**: Go 1.26+
**Primary Dependencies**: AwGo (github.com/deanishe/awgo), google-api-go-client (google.golang.org/api/tasks/v1), golang.org/x/oauth2
**Storage**: JSON files in Alfred's workflow data directory (`~/Library/Application Support/Alfred/Workflow Data/{bundleID}`)
**Testing**: `go test ./...`
**Target Platform**: macOS (Alfred is macOS-only)
**Project Type**: desktop-app (Alfred workflow, single compiled binary)
**Performance Goals**: Task creation < 2s, task listing < 3s
**Constraints**: Zero runtime dependencies beyond Alfred 5, single binary distribution
**Scale/Scope**: Single-user, personal task management

## Constitution Check

*Constitution template not filled in for this project. No gates to evaluate.*

## Project Structure

### Documentation (this feature)

```text
specs/001-google-tasks-workflow/
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
└── tasks.md
```

### Source Code (repository root)

```text
cmd/
└── main.go              # Entry point, command routing

internal/
├── auth/
│   ├── oauth.go          # OAuth flow: PKCE, local server, token exchange
│   ├── token.go          # Token storage: load, save, refresh
│   └── credentials.go    # client_secret.json parsing
├── tasks/
│   ├── client.go         # Google Tasks API wrapper
│   ├── create.go         # Task creation logic
│   ├── list.go           # Task listing with timeframe grouping
│   └── actions.go        # Complete, delete task actions
├── dateparse/
│   └── dateparse.go      # Custom date parser (today, tomorrow, weekdays, ISO)
├── input/
│   └── parser.go         # Input parsing: extract title, date, list tag
└── alfred/
    ├── workflow.go       # AwGo workflow setup, command routing
    ├── items.go          # Alfred Script Filter item builders
    └── notifications.go  # macOS notification helpers

go.mod
go.sum
Makefile                  # Build, test, package targets
info.plist                # Alfred workflow manifest
icon.png                  # Workflow icon
```

**Structure Decision**: Single Go module with `cmd/` entry point and `internal/` packages. The `internal/` directory enforces that packages are not importable by external code. Packages are organized by domain responsibility: `auth` for OAuth, `tasks` for API operations, `dateparse` for date extraction, `input` for user input parsing, and `alfred` for Alfred-specific integrations.

## Global Constraints

These constraints apply to every task implicitly. Copied verbatim from the spec:

- **Language**: Go 1.26+ (current stable version)
- **Binary**: Single compiled binary with zero runtime dependencies beyond Alfred 5
- **Platform**: macOS only (Alfred is macOS-only)
- **Distribution**: Single `.alfredworkflow` zip file
- **Alfred**: Alfred 5 with Powerpack required
- **OAuth Scope**: `https://www.googleapis.com/auth/tasks` (read-write)
- **Storage Location**: `~/Library/Application Support/Alfred/Workflow Data/{bundleID}`
- **Performance**: Task creation < 2s, task listing < 3s

## Key Interfaces

Cross-task interfaces for implementers who read only their own task:

### `internal/auth/credentials.go`
```go
func LoadClientCredentials(dataDir string) (*oauth2.Config, error)
```

### `internal/auth/token.go`
```go
func LoadToken(dataDir string) (*oauth2.Token, error)
func SaveToken(dataDir string, token *oauth2.Token) error
func DeleteToken(dataDir string) error
func EnsureValidToken(dataDir string, config *oauth2.Config) (*oauth2.Token, error)
```

### `internal/auth/oauth.go`
```go
func RunOAuthFlow(config *oauth2.Config) (*oauth2.Token, error)
```

### `internal/tasks/client.go`
```go
func NewClient(token *oauth2.Token, config *oauth2.Config) (*Client, error)
func (c *Client) ListTaskLists() ([]*tasks.TaskList, error)
func (c *Client) CreateTaskList(title string) (*tasks.TaskList, error)
func (c *Client) ListTasks(listID string) ([]*tasks.Task, error)
func (c *Client) InsertTask(listID string, task *tasks.Task) (*tasks.Task, error)
func (c *Client) CompleteTask(listID, taskID string) error
func (c *Client) DeleteTask(listID, taskID string) error
```

### `internal/dateparse/dateparse.go`
```go
func Parse(input string, relativeTo time.Time) (time.Time, bool)
```

### `internal/input/parser.go`
```go
type ParsedInput struct {
    Title    string
    Date     *time.Time
    ListName string
}
func Parse(input string) ParsedInput
```

### `internal/alfred/notifications.go`
```go
func NotifySuccess(wf *aw.Workflow, message string)
func NotifyError(wf *aw.Workflow, message string)
```

## Complexity Tracking

No constitution violations to justify.
