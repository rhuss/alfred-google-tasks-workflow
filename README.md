# Alfred Google Tasks Workflow

An Alfred 5 workflow for managing Google Tasks directly from your launcher. Create tasks with smart date parsing, view tasks grouped by timeframe, and manage them with quick actions.

## Features

- **Quick task creation** with natural language dates and list targeting
- **Task listing** grouped by Overdue, Today, This Week, Later, No Date
- **Task actions**: complete, delete, or open tasks from a sub-menu
- **List filtering**: view tasks from a specific list with `#ListName`
- **Auto-create lists**: reference a list that doesn't exist and it gets created
- **Secure OAuth 2.0** with PKCE, tokens stored locally

## Installation

### From GitHub Release

1. Download `alfred-google-tasks.alfredworkflow` from the [latest release](https://github.com/rhuss/alfred-google-tasks-workflow/releases/latest)
2. Double-click the file to install in Alfred

### From Source

Requires Go 1.26+.

```bash
git clone https://github.com/rhuss/alfred-google-tasks-workflow.git
cd alfred-google-tasks-workflow
make install
```

## Google Cloud Setup (Required)

You need your own Google Cloud credentials. This is a one-time setup that takes about 5 minutes.

### 1. Create a Google Cloud Project

1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Click **Select a project** at the top, then **New Project**
3. Name it something like "Alfred Tasks" and click **Create**

### 2. Enable the Tasks API

1. Go to **APIs & Services** > **Library**
2. Search for "Tasks API" (not "Google Tasks API")
3. Click **Enable**

### 3. Configure the OAuth Consent Screen

1. Go to **APIs & Services** > **OAuth consent screen**
2. If prompted for user type, select **External** and click **Create**
3. Fill in:
   - **App name**: Alfred Tasks (or anything you like)
   - **User support email**: your email
   - **Developer contact**: your email
4. Click **Save and Continue** through Scopes (no changes needed)
5. On **Test users**, click **Add Users** and add your Google email
6. Click **Save and Continue**, then **Back to Dashboard**

**Important**: Go to **OAuth consent screen** > **Publishing status** and click **Publish App**. This changes the status to "In production" and prevents your refresh token from expiring every 7 days. Verification is not required for personal use (you'll see a warning screen once during login, which is normal).

### 4. Create OAuth Credentials

1. Go to **APIs & Services** > **Credentials**
2. Click **Create Credentials** > **OAuth client ID**
3. Application type: **Desktop app**
4. Name: Alfred Tasks (or anything)
5. Click **Create**
6. Click **Download JSON** (the download button next to your new credential)
7. Rename the downloaded file to `client_secret.json`

### 5. Place the Credentials File

Copy `client_secret.json` to the workflow's data directory:

```bash
mkdir -p ~/Library/Application\ Support/Alfred/Workflow\ Data/com.rhuss.gtasks/
cp client_secret.json ~/Library/Application\ Support/Alfred/Workflow\ Data/com.rhuss.gtasks/
```

## Usage

### First-Time Login

Type `gt login` in Alfred. Your browser opens to Google's consent screen. Grant permission and you'll see "Authenticated! You can close this tab." Return to Alfred.

### Commands

| Command | Description | Example |
|---------|-------------|---------|
| `gt add <task>` | Create a task | `gt add Buy milk` |
| `gt add <task>, <date>` | Create with due date | `gt add Buy milk, tomorrow` |
| `gt add <task>, <date> #List` | Create in a specific list | `gt add Buy milk, friday #Shopping` |
| `gt list` | Show all tasks by timeframe | |
| `gt list #ListName` | Show tasks from one list | `gt list #Work` |
| `gt` | Same as `gt list` | |
| `gt open` | Open Google Tasks in browser | |
| `gt logout` | Disconnect your account | |

### Smart Date Parsing

Dates can appear at the beginning or end of the task title, optionally separated by a comma.

| Input | Due Date |
|-------|----------|
| `today` | Today |
| `tomorrow` | Tomorrow |
| `monday` through `sunday` | Next occurrence of that weekday |
| `next week` | Next Monday |
| `2026-07-15` | Exact date (ISO format) |
| `07-15` | July 15 of the current year (or next year if passed) |

**Examples**:
```
gt add Buy groceries, tomorrow          # "Buy groceries" due tomorrow
gt add Monday Review PR                 # "Review PR" due next Monday
gt add Submit report, 07-10 #Work       # "Submit report" due July 10, in Work list
gt add tomorrow, Call dentist           # "Call dentist" due tomorrow
```

### List Targeting

Add `#ListName` at the end to target a specific list. If the list doesn't exist, it gets created automatically.

- `#Shopping` targets the "Shopping" list
- `#My-Project` targets "My Project" (hyphens become spaces)
- `#home_tasks` targets "home tasks" (underscores become spaces)

Without a `#ListName`, tasks go to your default list (usually "My Tasks").

### Task Actions

When viewing the task list, press Enter on a task to see the action menu:

- **Complete Task**: marks the task as done
- **Delete Task**: permanently removes the task
- **Open in Browser**: opens Google Tasks web UI

## Troubleshooting

**"Setup Required" when running gt login**
Your `client_secret.json` is not in the right place. Check that it exists at:
`~/Library/Application Support/Alfred/Workflow Data/com.rhuss.gtasks/client_secret.json`

**"Not authenticated" on every command**
Run `gt login` first. If you've already logged in and it stopped working, your refresh token may have expired (this happens if your Google Cloud project is still in "Testing" mode). Publish your app to "In production" in the OAuth consent screen settings, then run `gt logout` followed by `gt login`.

**Google shows "This app isn't verified" warning**
This is normal for personal-use apps. Click **Advanced** > **Go to Alfred Tasks (unsafe)** > **Continue**. You only see this once.

**Tasks don't appear in the list**
The workflow only shows incomplete tasks. Completed tasks are hidden. Also check that you're looking at the right list (try `gt list` without a filter to see all lists).

**Build errors**
Make sure you have Go 1.26+ installed: `go version`. If using an older Go, update via `brew install go` or from [go.dev](https://go.dev/dl/).

## Development

```bash
make build      # Build the binary
make test       # Run all 67 tests
make clean      # Remove build artifacts
make package    # Create .alfredworkflow file
make install    # Build, package, and open in Alfred
```

## License

MIT
