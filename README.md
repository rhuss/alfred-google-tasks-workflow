# Alfred Google Tasks Workflow

An Alfred 5 workflow for managing Google Tasks directly from your launcher. Create tasks with smart date parsing, view tasks grouped by timeframe, and manage them with quick actions.

## Features

- **Quick Task Creation** (`gt add`): Create tasks with natural language dates and list targeting
  - `gt add Buy groceries, tomorrow #Shopping`
  - `gt add Call dentist, friday`
  - `gt add Review PR, next week #Work`
- **Task Listing** (`gt list`): View tasks grouped by Overdue, Today, This Week, Later, No Date
  - Filter by list: `gt list #Work`
- **Task Actions**: Complete, delete, or open tasks from the action sub-menu
- **Browser Access** (`gt open`): Open Google Tasks web UI
- **OAuth 2.0 with PKCE**: Secure authentication with Google

## Prerequisites

1. **Go 1.26+** installed
2. **Alfred 5** with Powerpack license
3. **Google Cloud Project** with Tasks API enabled

## Google Cloud Setup

1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Create a new project (or select existing)
3. Enable the **Google Tasks API**
4. Go to **Credentials** > **Create Credentials** > **OAuth 2.0 Client ID**
   - Application type: **Desktop app**
   - Download the JSON file
5. Rename the downloaded file to `client_secret.json`
6. Place it in the workflow's data directory:
   `~/Library/Application Support/Alfred/Workflow Data/com.rhuss.gtasks/`

## Installation

### From Source

```bash
git clone https://github.com/rhuss/alfred-google-tasks-workflow.git
cd alfred-google-tasks-workflow
make build
make package
```

Then double-click the generated `.alfredworkflow` file to install.

## Usage

### First-Time Setup

1. Type `gt login` in Alfred
2. Complete the Google OAuth flow in your browser
3. You're ready to use the workflow

### Commands

| Command | Description |
|---------|-------------|
| `gt add <task>, <date> #List` | Create a task with optional date and list |
| `gt list` | List all tasks grouped by timeframe |
| `gt list #ListName` | List tasks from a specific list |
| `gt open` | Open Google Tasks in browser |
| `gt logout` | Remove stored credentials |

### Date Formats

- `today`, `tomorrow`
- Weekday names: `monday`, `tuesday`, etc. (next occurrence)
- `next week` (next Monday)
- ISO format: `2026-07-15`
- Short format: `07-15` (current or next year)

### List Targeting

Use `#ListName` at the end of your input to target a specific list. The list is created automatically if it doesn't exist.

- Hyphens and underscores are converted to spaces: `#my-project` becomes "my project"

## Build

```bash
make build    # Build the binary
make test     # Run tests
make clean    # Clean build artifacts
make package  # Create .alfredworkflow package
```

## License

MIT
