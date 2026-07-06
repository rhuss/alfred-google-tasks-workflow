# Quickstart: Alfred Google Tasks Workflow

## Prerequisites

- macOS with Alfred 5 and Powerpack
- Go 1.26+ installed
- A Google Cloud project with the Tasks API enabled

## Setup Google Cloud Credentials

1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Create a new project (or select an existing one)
3. Enable the Google Tasks API: APIs & Services > Library > search "Tasks API" > Enable
4. Configure OAuth consent screen: APIs & Services > OAuth consent screen
   - Choose "External" user type
   - Fill in app name and support email
   - Add scope: `https://www.googleapis.com/auth/tasks`
   - Add yourself as a test user
   - Set publishing status to "In production" (avoids 7-day token expiry)
5. Create OAuth credentials: APIs & Services > Credentials > Create Credentials > OAuth client ID
   - Application type: "Desktop app"
   - Download the `client_secret.json` file

## Build the Workflow

```bash
git clone https://github.com/rhuss/alfred-google-tasks-workflow.git
cd alfred-google-tasks-workflow
make build
```

## Install

1. Double-click the generated `.alfredworkflow` file, or run `make install`
2. Copy `client_secret.json` to the workflow data directory:
   ```bash
   cp client_secret.json "~/Library/Application Support/Alfred/Workflow Data/com.rhuss.gtasks/"
   ```
3. In Alfred, type `gt login` to authenticate with Google

## Usage

```
gt add Buy groceries, tomorrow #Shopping    # Create task with date and list
gt add Review PR, Monday                    # Create task due next Monday
gt list                                     # Show all tasks by timeframe
gt list #Work                               # Show only Work list tasks
gt open                                     # Open Google Tasks in browser
gt logout                                   # Disconnect Google account
```

## Development

```bash
make test      # Run all tests
make build     # Build the binary
make package   # Create .alfredworkflow file
make clean     # Remove build artifacts
```
