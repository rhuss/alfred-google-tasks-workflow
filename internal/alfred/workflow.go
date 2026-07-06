package alfred

import (
	"fmt"
	"os"
	"strings"
	"time"

	aw "github.com/deanishe/awgo"

	"github.com/rhuss/alfred-google-tasks-workflow/internal/auth"
	"github.com/rhuss/alfred-google-tasks-workflow/internal/input"
	"github.com/rhuss/alfred-google-tasks-workflow/internal/tasks"
)

// Workflow wraps the AwGo workflow instance and provides command routing.
type Workflow struct {
	WF      *aw.Workflow
	DataDir string
}

// NewWorkflow creates a new Alfred workflow wrapper.
func NewWorkflow() *Workflow {
	wf := aw.New()
	dataDir := wf.DataDir()

	if credDir := os.Getenv("CREDENTIALS_DIR"); credDir != "" {
		expanded := credDir
		if len(expanded) > 0 && expanded[0] == '~' {
			if home, err := os.UserHomeDir(); err == nil {
				expanded = home + expanded[1:]
			}
		}
		dataDir = expanded
	}

	return &Workflow{
		WF:      wf,
		DataDir: dataDir,
	}
}

// Run starts the workflow and routes to the appropriate command handler.
// Args should be os.Args[1:] (the arguments after the binary name).
func (w *Workflow) Run(args []string) {
	w.WF.Run(func() {
		w.route(args)
	})
}

// route dispatches to the correct command handler based on arguments.
func (w *Workflow) route(args []string) {
	if len(args) == 0 {
		w.showHelp()
		return
	}

	command := args[0]
	remaining := args[1:]

	switch command {
	case "login":
		w.handleLogin()
	case "logout":
		w.handleLogout()
	case "add":
		w.handleAddPreview(strings.Join(remaining, " "))
	case "create":
		w.handleAdd(remaining)
	case "list":
		w.handleList(remaining)
	case "open":
		w.handleOpen()
	case "action":
		w.handleAction(remaining)
	default:
		// Default behavior: treat as script filter query for task listing
		w.handleFilter(args)
	}
}

// showHelp displays available commands as Alfred items.
func (w *Workflow) showHelp() {
	w.WF.NewItem("gt login").
		Subtitle("Authenticate with Google Tasks").
		Arg("login").
		Valid(true)
	w.WF.NewItem("gt add <task>").
		Subtitle("Add a new task (supports dates and #ListName)").
		Arg("add").
		Valid(false)
	w.WF.NewItem("gt list").
		Subtitle("List your tasks grouped by timeframe").
		Arg("list").
		Valid(true)
	w.WF.NewItem("gt open").
		Subtitle("Open Google Tasks in your browser").
		Arg("open").
		Valid(true)
	w.WF.NewItem("gt logout").
		Subtitle("Remove stored credentials").
		Arg("logout").
		Valid(true)
	w.WF.SendFeedback()
}

// handleFilter processes script filter queries (the main Alfred input).
func (w *Workflow) handleFilter(args []string) {
	query := ""
	if len(args) > 0 {
		query = args[0]
	}

	// Route based on query prefix
	switch {
	case query == "login":
		w.handleLogin()
	case query == "logout":
		w.handleLogout()
	case query == "open":
		w.handleOpen()
	case len(query) >= 4 && query[:4] == "add ":
		w.handleAddPreview(query[4:])
	case len(query) >= 5 && query[:5] == "list ":
		w.handleList([]string{strings.TrimSpace(query[5:])})
	case query == "list" || query == "":
		w.handleList(nil)
	default:
		w.handleList([]string{query})
	}
}

// handleLogin initiates the OAuth flow.
func (w *Workflow) handleLogin() {
	// Check if already authenticated
	if auth.TokenExists(w.DataDir) {
		w.WF.NewItem("Already authenticated").
			Subtitle("Run 'gt logout' first to re-authenticate").
			Icon(aw.IconInfo).
			Valid(false)
		w.WF.SendFeedback()
		return
	}

	// Load client credentials
	config, err := auth.LoadClientCredentials(w.DataDir)
	if err != nil {
		w.WF.NewItem("Setup Required").
			Subtitle("Place client_secret.json in workflow data directory").
			Icon(aw.IconError).
			Valid(false)
		w.WF.SendFeedback()
		fmt.Fprintf(os.Stderr, "credentials error: %v\n", err)
		return
	}

	// Run OAuth flow (opens browser, waits for callback)
	token, err := auth.RunOAuthFlow(config)
	if err != nil {
		NotifyError("Google Tasks", fmt.Sprintf("Login failed: %v", err))
		fmt.Fprintf(os.Stderr, "oauth error: %v\n", err)
		return
	}

	// Save the token
	if err := auth.SaveToken(w.DataDir, token); err != nil {
		NotifyError("Google Tasks", fmt.Sprintf("Failed to save token: %v", err))
		fmt.Fprintf(os.Stderr, "save token error: %v\n", err)
		return
	}

	NotifySuccess("Google Tasks", "Successfully authenticated!")
	w.WF.SendFeedback()
}

// requireAuth checks if the user is authenticated and shows an error item if not.
// Returns true if authenticated, false if not.
func (w *Workflow) requireAuth() bool {
	if !auth.TokenExists(w.DataDir) {
		w.WF.NewItem("Not authenticated").
			Subtitle("Run 'gt login' to connect your Google account").
			Icon(aw.IconError).
			Valid(false)
		w.WF.SendFeedback()
		return false
	}
	return true
}

// getAuthenticatedClient loads credentials and a valid token, then returns
// the config and token needed to create an API client. Returns nil values
// and sends error feedback if anything fails.
func (w *Workflow) getAuthenticatedClient() (*auth.OAuthConfig, error) {
	config, err := auth.LoadClientCredentials(w.DataDir)
	if err != nil {
		return nil, fmt.Errorf("loading credentials: %w", err)
	}

	token, err := auth.EnsureValidToken(w.DataDir, config)
	if err != nil {
		return nil, fmt.Errorf("validating token: %w", err)
	}

	return &auth.OAuthConfig{Config: config, Token: token}, nil
}

// handleLogout removes stored credentials.
func (w *Workflow) handleLogout() {
	if err := auth.DeleteToken(w.DataDir); err != nil {
		NotifyError("Google Tasks", err.Error())
		fmt.Fprintf(os.Stderr, "logout error: %v\n", err)
		return
	}
	NotifySuccess("Google Tasks", "Logged out successfully")
	w.WF.SendFeedback()
}

// handleAddPreview shows a preview of the task to be created.
// The actual creation happens when the user presses Enter (via the "create" command).
func (w *Workflow) handleAddPreview(rawInput string) {
	rawInput = strings.TrimSpace(rawInput)
	if rawInput == "" {
		w.WF.NewItem("Type a task description").
			Subtitle("Example: Buy groceries, tomorrow #Shopping").
			Icon(iconComplete).
			Valid(false)
		w.WF.SendFeedback()
		return
	}

	parsed := input.ParseWithTime(rawInput, time.Now())

	title := parsed.Title
	if title == "" {
		title = rawInput
	}

	subtitle := "Press Enter to create"
	if parsed.Date != nil {
		subtitle += fmt.Sprintf(" (due %s)", parsed.Date.Format("2006-01-02"))
	}
	if parsed.ListName != "" {
		subtitle += fmt.Sprintf(" in \"%s\"", parsed.ListName)
	}

	w.WF.NewItem(title).
		Subtitle(subtitle).
		Arg("create:" + rawInput).
		Icon(iconComplete).
		Valid(true)

	w.WF.SendFeedback()
}

// handleAdd creates a new task from the input.
func (w *Workflow) handleAdd(args []string) {
	if !w.requireAuth() {
		return
	}

	if len(args) == 0 || strings.TrimSpace(strings.Join(args, " ")) == "" {
		NotifyError("Google Tasks", "No task description provided")
		return
	}

	rawInput := strings.Join(args, " ")

	authConfig, err := w.getAuthenticatedClient()
	if err != nil {
		NotifyError("Google Tasks", fmt.Sprintf("Auth error: %v", err))
		fmt.Fprintf(os.Stderr, "auth error: %v\n", err)
		return
	}

	client, err := tasks.NewClient(authConfig.Token, authConfig.Config)
	if err != nil {
		NotifyError("Google Tasks", fmt.Sprintf("API client error: %v", err))
		fmt.Fprintf(os.Stderr, "client error: %v\n", err)
		return
	}

	created, listName, err := client.CreateTaskFromInput(rawInput)
	if err != nil {
		NotifyError("Google Tasks", fmt.Sprintf("Failed to create task: %v", err))
		fmt.Fprintf(os.Stderr, "create task error: %v\n", err)
		return
	}

	subtitle := fmt.Sprintf("Added to %s", listName)
	if len(created.Due) >= 10 {
		subtitle += fmt.Sprintf(" (due %s)", created.Due[:10])
	}
	NotifySuccess("Google Tasks", fmt.Sprintf("Created: %s\n%s", created.Title, subtitle))
}

// handleList displays tasks grouped by timeframe.
// Supports optional #ListName filter in args.
func (w *Workflow) handleList(args []string) {
	if !w.requireAuth() {
		return
	}

	authConfig, err := w.getAuthenticatedClient()
	if err != nil {
		w.WF.NewItem("Authentication error").
			Subtitle(err.Error()).
			Icon(aw.IconError).
			Valid(false)
		w.WF.SendFeedback()
		fmt.Fprintf(os.Stderr, "auth error: %v\n", err)
		return
	}

	client, err := tasks.NewClient(authConfig.Token, authConfig.Config)
	if err != nil {
		w.WF.NewItem("API client error").
			Subtitle(err.Error()).
			Icon(aw.IconError).
			Valid(false)
		w.WF.SendFeedback()
		fmt.Fprintf(os.Stderr, "client error: %v\n", err)
		return
	}

	// Check for #ListName filter in args
	listFilter := ""
	if len(args) > 0 {
		query := strings.Join(args, " ")
		if strings.HasPrefix(query, "#") && len(query) > 1 {
			listFilter = strings.TrimSpace(query[1:])
			listFilter = strings.ReplaceAll(listFilter, "-", " ")
			listFilter = strings.ReplaceAll(listFilter, "_", " ")
		}
	}

	var items []tasks.TaskItem
	if listFilter != "" {
		items, err = client.FetchTasksFromList(listFilter)
	} else {
		items, err = client.FetchAllTasks()
	}
	if err != nil {
		w.WF.NewItem("Failed to fetch tasks").
			Subtitle(err.Error()).
			Icon(aw.IconError).
			Valid(false)
		w.WF.SendFeedback()
		fmt.Fprintf(os.Stderr, "fetch tasks error: %v\n", err)
		return
	}

	grouped := tasks.GroupTasksByTimeframe(items, time.Now())
	w.RenderGroupedTasks(grouped)
}

// handleOpen opens Google Tasks in the browser.
func (w *Workflow) handleOpen() {
	if err := tasks.OpenGoogleTasks(); err != nil {
		NotifyError("Google Tasks", fmt.Sprintf("Failed to open browser: %v", err))
		fmt.Fprintf(os.Stderr, "open error: %v\n", err)
		return
	}
}

// handleAction processes task actions (complete, delete, open).
// Arg format: "action:listID:taskID" or just "listID:taskID" for sub-menu.
func (w *Workflow) handleAction(args []string) {
	if len(args) == 0 {
		return
	}

	actionArg := args[0]

	// Handle task creation (from add preview)
	if strings.HasPrefix(actionArg, "create:") {
		w.handleAdd([]string{actionArg[7:]})
		return
	}

	// Handle action execution (format: action|listID:taskID)
	if strings.Contains(actionArg, "|") {
		pipeParts := strings.SplitN(actionArg, "|", 2)
		action := pipeParts[0]
		idParts := strings.SplitN(pipeParts[1], ":", 2)
		if len(idParts) != 2 {
			NotifyError("Google Tasks", "Invalid action format")
			return
		}
		listID := idParts[0]
		taskID := idParts[1]
		w.executeAction(action, listID, taskID)
		return
	}

	// No pipe separator: this is a task selection from the list
	// Read listID and taskID from Alfred workflow variables (set on items)
	listID := os.Getenv("listID")
	taskID := os.Getenv("taskID")
	if listID == "" || taskID == "" {
		NotifyError("Google Tasks", "Could not identify task")
		return
	}
	w.RenderActionMenu(listID, taskID)
}

func (w *Workflow) executeAction(action, listID, taskID string) {

	// "open" action doesn't need auth
	if action == "open" {
		if err := tasks.OpenGoogleTasks(); err != nil {
			NotifyError("Google Tasks", fmt.Sprintf("Failed to open browser: %v", err))
		}
		return
	}

	if !w.requireAuth() {
		return
	}

	authConfig, err := w.getAuthenticatedClient()
	if err != nil {
		NotifyError("Google Tasks", fmt.Sprintf("Auth error: %v", err))
		fmt.Fprintf(os.Stderr, "auth error: %v\n", err)
		return
	}

	client, err := tasks.NewClient(authConfig.Token, authConfig.Config)
	if err != nil {
		NotifyError("Google Tasks", fmt.Sprintf("API client error: %v", err))
		fmt.Fprintf(os.Stderr, "client error: %v\n", err)
		return
	}

	switch action {
	case "complete":
		if err := client.CompleteTaskByID(listID, taskID); err != nil {
			NotifyError("Google Tasks", fmt.Sprintf("Failed to complete task: %v", err))
			fmt.Fprintf(os.Stderr, "complete error: %v\n", err)
			return
		}
		NotifySuccess("Google Tasks", "Task completed!")

	case "delete":
		if err := client.DeleteTaskByID(listID, taskID); err != nil {
			NotifyError("Google Tasks", fmt.Sprintf("Failed to delete task: %v", err))
			fmt.Fprintf(os.Stderr, "delete error: %v\n", err)
			return
		}
		NotifySuccess("Google Tasks", "Task deleted")

	default:
		NotifyError("Google Tasks", fmt.Sprintf("Unknown action: %s", action))
	}
}
