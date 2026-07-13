# Alfred Google Tasks Workflow

An [Alfred 5](https://www.alfredapp.com/) workflow for managing Google Tasks without leaving your launcher. Create tasks with natural dates, view them grouped by timeframe, and manage multiple Google accounts.

## Features

- **Quick task creation** with natural language dates (`tomorrow`, `friday`, `next week`) and list targeting (`#Shopping`)
- **Task listing** grouped by Overdue, Today, This Week, Later, No Date
- **Task actions**: complete, delete, or open in browser from a sub-menu
- **Multi-account support**: manage tasks across personal and work Google accounts
- **Quick-add keywords**: configurable shortcuts (`gta`, `gtp`) for adding tasks to specific accounts
- **Hotkey trigger**: assign a keyboard shortcut for instant task creation
- **Auto-create lists**: reference a list that doesn't exist and it gets created
- **Secure OAuth 2.0** with PKCE, tokens stored locally

## Installation

### From GitHub Release

1. Download `alfred-google-tasks.alfredworkflow` from the [latest release](https://github.com/rhuss/alfred-google-tasks-workflow/releases/latest)
2. Double-click the file to install in Alfred

### From Source

Requires Go 1.22+.

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

> **Note:** Go to **OAuth consent screen** > **Publishing status** and click **Publish App** to change the status to "In production". This prevents your refresh token from expiring every 7 days. Verification is not required for personal use. You'll see a warning screen once during login, which is normal.

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

Type `gt login` in Alfred. Your browser opens to Google's consent screen. Grant permission and you'll see "Authenticated! You can close this tab."

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

**Examples:**
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

## Multi-Account Support

Manage tasks across multiple Google accounts (e.g., personal and work). This is optional. Without any configuration, the workflow behaves as a single-account setup.

### Setup

1. Create separate Google Cloud credentials for each account (follow the Google Cloud Setup above for each).

2. Place each credentials file in the workflow data directory with a unique name:
   ```bash
   DATA_DIR=~/Library/Application\ Support/Alfred/Workflow\ Data/com.rhuss.gtasks
   cp personal-credentials.json "$DATA_DIR/personal.json"
   cp work-credentials.json "$DATA_DIR/work.json"
   ```

3. Create `accounts.json` in the same directory:
   ```json
   {
     "default": "personal",
     "list_default": "all",
     "accounts": {
       "personal": {
         "credentials": "personal.json"
       },
       "work": {
         "credentials": "work.json",
         "authuser": 1
       }
     }
   }
   ```

### accounts.json Reference

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `default` | string | Yes | Name of the default account |
| `list_default` | string | No | `"default"` (show default account only) or `"all"` (merge all accounts). Default: `"default"` |
| `accounts.<name>.credentials` | string | Yes | Path to credentials file (relative to data directory) |
| `accounts.<name>.authuser` | int | No | Google multi-login index for browser opening (0-based) |
| `accounts.<name>.keyword` | string | No | Dedicated Alfred keyword for this account |

### Using Multiple Accounts

**Target a specific account** with the `@` prefix:

| Command | Description |
|---------|-------------|
| `gt @work list` | List tasks from work account |
| `gt @personal add Buy milk` | Add task to personal account |
| `gt @work open` | Open Google Tasks for work profile |
| `gt @work login` | Authenticate the work account |

**Merged listing**: When `list_default` is `"all"`, typing `gt list` shows tasks from all authenticated accounts. Each task's subtitle shows its account name in parentheses (e.g., `[Today] Inbox (work)`).

**Default account**: Commands without `@` prefix use the default account (unless `list_default` is `"all"` for the `list` command).

**Account selector**: Type `gt @` to see all configured accounts.

### Quick-Add Keywords

The workflow includes two configurable quick-add keywords for fast task creation on specific accounts. Instead of typing `gt @work add Buy milk`, you type `gta Buy milk`.

Configure them in Alfred's workflow settings (click the `[x]` button in the workflow editor):

| Setting | Default | Description |
|---------|---------|-------------|
| Quick Add 1 Keyword | `gta` | Alfred keyword for quick-adding to account 1 |
| Quick Add 1 Account | *(empty)* | Account name from `accounts.json` (e.g., `work`) |
| Quick Add 2 Keyword | `gtp` | Alfred keyword for quick-adding to account 2 |
| Quick Add 2 Account | *(empty)* | Account name from `accounts.json` (e.g., `personal`) |

Quick-add keywords support the full syntax: dates, `#ListName`, everything that `gt add` supports.

### Hotkey

The workflow includes a Hotkey trigger node connected to Quick Add 1. To set it up:

1. Open Alfred Preferences > Workflows > Google Tasks
2. Double-click the Hotkey node (empty white box, bottom-left)
3. Record your preferred keyboard shortcut

Pressing the shortcut opens Alfred with the quick-add keyword, ready for you to type a task.

## Workflow Settings

All settings are accessible via the `[x]` button in Alfred's workflow editor.

| Setting | Description |
|---------|-------------|
| Credentials Directory | Custom path for `accounts.json` and `client_secret.json`. Defaults to `~/Library/Application Support/Alfred/Workflow Data/com.rhuss.gtasks/` |
| Default Account | Override the default account from `accounts.json` |
| Task Listing Mode | "Default account only" or "All accounts (merged)" |
| Quick Add 1 Keyword | Alfred keyword for quick-add slot 1 (default: `gta`) |
| Quick Add 1 Account | Account name for quick-add slot 1 |
| Quick Add 2 Keyword | Alfred keyword for quick-add slot 2 (default: `gtp`) |
| Quick Add 2 Account | Account name for quick-add slot 2 |

## Idea Inbox Sync (Obsidian)

Automatically sync tasks from a designated Google Tasks list to an Obsidian markdown file. Useful for capturing ideas via voice (e.g., Google Gemini on Android) and routing them into your Obsidian inbox for later processing.

The sync runs silently during every `gt list` operation. New ideas are appended to the inbox file and deleted from Google Tasks. Already-synced ideas are skipped (deduplicated by TaskID). The feature scans all authenticated accounts, regardless of which account is targeted.

### Setup

Set two environment variables in Alfred's workflow configuration (`[x]` button):

| Variable | Description | Example |
|----------|-------------|---------|
| `IDEA_INBOX_PATH` | Absolute path to the Obsidian inbox markdown file | `~/Obsidian/Vault/00-Inbox/ideas.md` |
| `IDEA_LIST_NAME` | Name of the Google Tasks list to sync from | `Ideas` |

Both variables must be set for the feature to activate. If either is missing, the workflow behaves normally with zero overhead.

The inbox file and any parent directories are created automatically on first sync.

### Inbox File Format

Each synced idea is appended as an H3 entry with metadata:

```markdown
# Idea Inbox

### Buy noise-cancelling headphones
- Date: 2026-07-13
- Account: personal
- TaskID: dHJhbnNpdC0xMjM

Check reviews on Wirecutter first. Budget around 300 EUR.

### Learn origami
- Date: 2026-07-12
- Account: work
- TaskID: dHJhbnNpdC00NTY
```

The `Account` field is omitted in single-account mode.

### Voice Capture with Gemini (Android)

On Android, you can use Google Gemini to quickly capture ideas by voice:

> "Add a task *Build a standing desk* to my Ideas list"

The task lands in your Google Tasks "Ideas" list instantly. Next time you run `gt list` on your Mac, the idea is synced to your Obsidian inbox and removed from Google Tasks.

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
Make sure you have Go 1.22+ installed: `go version`. If using an older Go, update via `brew install go` or from [go.dev](https://go.dev/dl/).

## Development

```bash
make build                    # Build the binary
make test                     # Run tests
make clean                    # Remove build artifacts
make package                  # Create .alfredworkflow file
make install                  # Build, package, and open in Alfred
make release VERSION=v1.2.0   # Tag, push, and create GitHub release
```

## License

MIT
