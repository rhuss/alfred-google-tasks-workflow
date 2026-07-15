package alfred

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	aw "github.com/deanishe/awgo"

	"github.com/rhuss/alfred-google-tasks-workflow/internal/auth"
	"github.com/rhuss/alfred-google-tasks-workflow/internal/ideas"
	"github.com/rhuss/alfred-google-tasks-workflow/internal/input"
	"github.com/rhuss/alfred-google-tasks-workflow/internal/tasks"
	taskapi "google.golang.org/api/tasks/v1"
)

// Workflow wraps the AwGo workflow instance and provides command routing.
type Workflow struct {
	WF              *aw.Workflow
	AccountCtx      *auth.AccountContext
	AccountConfig   *auth.AccountConfig // nil in single-account mode
	configErr       error               // non-nil if accounts.json exists but is invalid
	accountTargeted bool                // true if user explicitly used @accountname
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

	w := &Workflow{
		WF: wf,
	}

	// Try to load multi-account configuration
	accountConfig, err := auth.LoadAccountConfig(dataDir)
	if err != nil {
		// Config exists but is invalid; store the error for later display.
		// Use DefaultContext so the workflow can still show error items.
		w.AccountCtx = auth.DefaultContext(dataDir)
		w.configErr = err
		return w
	}

	if accountConfig != nil {
		// Apply Alfred config overrides from environment variables
		if envDefault := os.Getenv("DEFAULT_ACCOUNT"); envDefault != "" {
			accountConfig.Default = envDefault
		}
		if envListDefault := os.Getenv("LIST_DEFAULT"); envListDefault != "" {
			accountConfig.ListDefault = envListDefault
		}

		w.AccountConfig = accountConfig

		// Quick-add: resolve account by name from QUICK_ADD_ACCOUNT env var
		if accountName := os.Getenv("QUICK_ADD_ACCOUNT"); accountName != "" {
			ctx, err := auth.ResolveAccount(accountConfig, accountName)
			if err != nil {
				w.AccountCtx = auth.DefaultContext(dataDir)
				w.configErr = err
				return w
			}
			w.AccountCtx = ctx
			w.accountTargeted = true
			return w
		}

		// Item var propagation: accountName set on Script Filter items
		// and propagated by Alfred to downstream Run Script actions
		if accountName := os.Getenv("accountName"); accountName != "" {
			ctx, err := auth.ResolveAccount(accountConfig, accountName)
			if err == nil {
				w.AccountCtx = ctx
				w.accountTargeted = true
				return w
			}
		}

		// Per-account keyword from accounts.json
		if keyword := os.Getenv("ACCOUNT_KEYWORD"); keyword != "" {
			if acct := accountConfig.FindAccountByKeyword(keyword); acct != nil {
				ctx, err := auth.ResolveAccount(accountConfig, acct.Name)
				if err != nil {
					w.AccountCtx = auth.DefaultContext(dataDir)
					w.configErr = err
					return w
				}
				w.AccountCtx = ctx
				w.accountTargeted = true
				return w
			}
			w.AccountCtx = auth.DefaultContext(dataDir)
			w.configErr = fmt.Errorf("keyword %q does not match any account in accounts.json", keyword)
			return w
		}

		// Resolve the default account
		ctx, err := auth.ResolveAccount(accountConfig, "")
		if err != nil {
			w.AccountCtx = auth.DefaultContext(dataDir)
			w.configErr = err
			return w
		}
		w.AccountCtx = ctx
	} else {
		// Single-account mode: no accounts.json found
		w.AccountCtx = auth.DefaultContext(dataDir)
	}

	return w
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
	// If accounts.json exists but is invalid, show the error to the user
	// instead of proceeding with any command.
	if w.configErr != nil {
		w.WF.NewItem("Account configuration error").
			Subtitle(w.configErr.Error()).
			Icon(aw.IconError).
			Valid(false)
		w.WF.NewItem("Fix accounts.json in workflow data directory").
			Subtitle("Delete or correct the file, then try again").
			Icon(aw.IconInfo).
			Valid(false)
		w.WF.SendFeedback()
		return
	}

	if len(args) == 0 {
		w.showHelp()
		return
	}

	// In multi-account mode, check if the first arg is an @accountname.
	// Skip if account was already resolved via keyword (accountTargeted).
	// Alfred's Script Filter may pass the entire query as a single arg
	// (e.g., "@work login"), so split on space to extract just the account name.
	if w.AccountConfig != nil && !w.accountTargeted && len(args) > 0 && strings.HasPrefix(args[0], "@") {
		firstArg := args[0]
		parts := strings.SplitN(firstArg[1:], " ", 2)
		accountName := parts[0]
		if accountName == "" {
			w.showAccountList()
			return
		}
		ctx, err := auth.ResolveAccount(w.AccountConfig, accountName)
		if err != nil {
			w.showUnknownAccountError(accountName)
			return
		}
		w.AccountCtx = ctx
		w.accountTargeted = true
		// Rebuild args: replace first arg with remainder after @account
		if len(parts) > 1 && parts[1] != "" {
			args = append([]string{parts[1]}, args[1:]...)
		} else {
			args = args[1:]
		}
		if len(args) == 0 {
			w.showHelp()
			return
		}
	}

	command := args[0]
	remaining := args[1:]

	switch command {
	case "login":
		w.showLoginPreview()
	case "logout":
		w.showCommandPreview("Logout", "Remove stored credentials", "cmd:logout")
	case "add":
		w.handleAddPreview(strings.Join(remaining, " "))
	case "create":
		w.handleAdd(remaining)
	case "list":
		w.handleList(remaining)
	case "open":
		w.showCommandPreview("Open Google Tasks", "Open in your browser", "cmd:open")
	case "action":
		w.handleAction(remaining)
	default:
		// Default behavior: treat as script filter query for task listing
		w.handleFilter(args)
	}
}

// showHelp displays available commands as Alfred items.
func (w *Workflow) showHelp() {
	prefix := "gt"
	if w.AccountConfig != nil && w.AccountCtx.Name != "" && w.accountTargeted {
		prefix = fmt.Sprintf("gt @%s", w.AccountCtx.Name)
	}

	w.WF.NewItem(prefix + " login").
		Subtitle("Authenticate with Google Tasks").
		Arg("cmd:login").
		Valid(true)
	w.WF.NewItem(prefix + " add <task>").
		Subtitle("Add a new task (supports dates and #ListName)").
		Valid(false)
	w.WF.NewItem(prefix + " list").
		Subtitle("List your tasks grouped by timeframe").
		Arg("cmd:list").
		Valid(true)
	w.WF.NewItem(prefix + " open").
		Subtitle("Open Google Tasks in your browser").
		Arg("cmd:open").
		Valid(true)
	w.WF.NewItem(prefix + " logout").
		Subtitle("Remove stored credentials").
		Arg("cmd:logout").
		Valid(true)

	// In multi-account mode, show the @account syntax hint
	if w.AccountConfig != nil && !w.accountTargeted {
		w.WF.NewItem("gt @<account> <command>").
			Subtitle("Target a specific account (e.g., gt @work list)").
			Icon(aw.IconAccount).
			Valid(false)
	}

	w.WF.SendFeedback()
}

// extractAccountPrefix checks for an @accountname prefix in the query string.
// In multi-account mode, if the first token starts with @, it is treated as an
// account selector. The account is resolved and set on the Workflow, and the
// cleaned query (without the @prefix) is returned.
// Returns the cleaned query and true if routing should continue, or false if
// an error was displayed and routing should stop.
func (w *Workflow) extractAccountPrefix(query string) (string, bool) {
	// Only process @ prefix in multi-account mode, and skip if account
	// was already resolved via keyword (accountTargeted).
	if w.AccountConfig == nil || w.accountTargeted {
		return query, true
	}

	query = strings.TrimSpace(query)
	if !strings.HasPrefix(query, "@") {
		return query, true
	}

	// Extract the account name (first token after @)
	rest := query[1:] // strip leading @
	parts := strings.SplitN(rest, " ", 2)
	accountName := parts[0]

	if accountName == "" {
		// Bare "@" with nothing after it: show available accounts
		w.showAccountList()
		return "", false
	}

	// Resolve the account
	ctx, err := auth.ResolveAccount(w.AccountConfig, accountName)
	if err != nil {
		// Unknown account name: show error with valid options
		w.showUnknownAccountError(accountName)
		return "", false
	}

	// Update the workflow's account context
	w.AccountCtx = ctx
	w.accountTargeted = true

	// Return the remaining query after the @accountname
	if len(parts) > 1 {
		return strings.TrimSpace(parts[1]), true
	}
	return "", true
}

// showAccountList displays all configured account names as Alfred items.
func (w *Workflow) showAccountList() {
	names := w.AccountConfig.AccountNames()
	for _, name := range names {
		subtitle := "Use @" + name + " to target this account"
		if name == w.AccountConfig.Default || (w.AccountConfig.Default == "" && name == names[0]) {
			subtitle += " (default)"
		}
		w.WF.NewItem("@" + name).
			Subtitle(subtitle).
			Icon(aw.IconAccount).
			Valid(false)
	}
	w.WF.SendFeedback()
}

// showUnknownAccountError displays an error for an unrecognized @accountname
// and lists valid account names.
func (w *Workflow) showUnknownAccountError(name string) {
	w.WF.NewItem(fmt.Sprintf("Unknown account: %s", name)).
		Subtitle("Valid accounts are listed below").
		Icon(aw.IconError).
		Valid(false)
	for _, n := range w.AccountConfig.AccountNames() {
		w.WF.NewItem("@" + n).
			Subtitle("Use @" + n + " to target this account").
			Icon(aw.IconAccount).
			Valid(false)
	}
	w.WF.SendFeedback()
}

// handleFilter processes script filter queries (the main Alfred input).
func (w *Workflow) handleFilter(args []string) {
	query := ""
	if len(args) > 0 {
		query = args[0]
	}

	// In multi-account mode, extract @accountname prefix
	query, ok := w.extractAccountPrefix(query)
	if !ok {
		return // error was displayed
	}

	// Route based on query prefix.
	// Commands that perform actions (login, logout, open) are shown as
	// preview items with a cmd: prefix. Pressing Enter routes them through
	// the Conditional to the run-create Run Script, which executes them.
	// This prevents actions from running while the user is still typing.
	switch {
	case query == "login":
		w.showLoginPreview()
	case query == "logout":
		w.showCommandPreview("Logout", "Remove stored credentials", "cmd:logout")
	case query == "open":
		w.showCommandPreview("Open Google Tasks", "Open in your browser", "cmd:open")
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

func (w *Workflow) showLoginPreview() {
	if auth.TokenExists(w.AccountCtx.DataDir) {
		w.WF.NewItem("Already authenticated").
			Subtitle("Run 'gt logout' first to re-authenticate").
			Icon(aw.IconInfo).
			Valid(false)
	} else {
		w.WF.NewItem("Login to Google Tasks").
			Subtitle("Press Enter to start authentication").
			Arg("cmd:login").
			Icon(iconComplete).
			Valid(true)
	}
	w.WF.SendFeedback()
}

func (w *Workflow) showCommandPreview(title, subtitle, arg string) {
	w.WF.NewItem(title).
		Subtitle(subtitle).
		Arg(arg).
		Icon(iconComplete).
		Valid(true)
	w.WF.SendFeedback()
}

// handleLogin initiates the OAuth flow.
func (w *Workflow) handleLogin() {
	// Check if already authenticated
	if auth.TokenExists(w.AccountCtx.DataDir) {
		w.WF.NewItem("Already authenticated").
			Subtitle("Run 'gt logout' first to re-authenticate").
			Icon(aw.IconInfo).
			Valid(false)
		w.WF.SendFeedback()
		return
	}

	// Load client credentials
	config, err := auth.LoadClientCredentialsFrom(w.AccountCtx.CredentialsPath)
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
	if err := auth.SaveToken(w.AccountCtx.DataDir, token); err != nil {
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
	if !auth.TokenExists(w.AccountCtx.DataDir) {
		loginHint := "Press Enter to connect your Google account"
		if w.AccountConfig != nil && w.AccountCtx.Name != "" {
			loginHint = fmt.Sprintf("Press Enter to authenticate the %s account", w.AccountCtx.Name)
		}
		w.WF.NewItem("Not authenticated").
			Subtitle(loginHint).
			Arg("cmd:login").
			Icon(aw.IconError).
			Valid(true)
		w.WF.SendFeedback()
		return false
	}
	return true
}

// getAuthenticatedClient loads credentials and a valid token, then returns
// the config and token needed to create an API client. Returns nil values
// and sends error feedback if anything fails.
func (w *Workflow) getAuthenticatedClient() (*auth.OAuthConfig, error) {
	config, err := auth.LoadClientCredentialsFrom(w.AccountCtx.CredentialsPath)
	if err != nil {
		return nil, fmt.Errorf("loading credentials: %w", err)
	}

	token, err := auth.EnsureValidToken(w.AccountCtx.DataDir, config)
	if err != nil {
		return nil, fmt.Errorf("validating token: %w", err)
	}

	return &auth.OAuthConfig{Config: config, Token: token}, nil
}

// handleLogout removes stored credentials.
func (w *Workflow) handleLogout() {
	if err := auth.DeleteToken(w.AccountCtx.DataDir); err != nil {
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

	item := w.WF.NewItem(title).
		Subtitle(subtitle).
		Arg("create:" + rawInput).
		Icon(iconComplete).
		Valid(true)

	if w.AccountCtx != nil && w.AccountCtx.Name != "" {
		item.Var("accountName", w.AccountCtx.Name)
	}

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

	if w.addToIdeaInbox(client, created, listName) {
		NotifySuccess("Google Tasks", fmt.Sprintf("Idea captured: %s", created.Title))
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

	// Determine if we should list from all accounts (merged listing)
	if w.shouldListAllAccounts() {
		w.handleListAllAccounts(listFilter)
		return
	}

	// Single-account listing (default account, explicit @account, or single-account mode)
	if !w.requireAuth() {
		return
	}

	if w.AccountConfig != nil {
		w.syncIdeasAllAccounts()
	} else {
		w.syncIdeasToInbox()
	}

	items, err := w.fetchTasksForCurrentAccount(listFilter)
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

// shouldListAllAccounts returns true when the listing should merge tasks
// from all configured accounts. This happens when:
// - Multi-account mode is active (AccountConfig is loaded)
// - list_default is "all"
// - The user did not explicitly target an account with @prefix
func (w *Workflow) shouldListAllAccounts() bool {
	if w.AccountConfig == nil {
		return false
	}
	if w.accountTargeted {
		return false
	}
	return w.AccountConfig.ListDefault == "all"
}

// handleListAllAccounts fetches tasks from all configured accounts sequentially,
// tags each task with its account name, and renders the merged result.
func (w *Workflow) handleListAllAccounts(listFilter string) {
	w.syncIdeasAllAccounts()

	var allItems []tasks.TaskItem
	var warnings []string

	originalCtx := w.AccountCtx

	for _, name := range w.AccountConfig.AccountNames() {
		ctx, err := auth.ResolveAccount(w.AccountConfig, name)
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("%s: config error", name))
			continue
		}

		// Temporarily switch context to this account
		w.AccountCtx = ctx

		if !auth.TokenExists(ctx.DataDir) {
			warnings = append(warnings, fmt.Sprintf("%s: not authenticated", name))
			continue
		}

		items, err := w.fetchTasksForCurrentAccount(listFilter)
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("%s: %v", name, err))
			fmt.Fprintf(os.Stderr, "fetch tasks for %s: %v\n", name, err)
			continue
		}

		// Tag each item with the account name
		for i := range items {
			items[i].AccountName = name
		}

		allItems = append(allItems, items...)
	}

	// Restore original context
	w.AccountCtx = originalCtx

	if len(allItems) == 0 && len(warnings) == 0 {
		// No tasks and no warnings: show empty state via RenderGroupedTasks
		grouped := tasks.GroupTasksByTimeframe(allItems, time.Now())
		w.RenderGroupedTasks(grouped)
		return
	}

	if len(allItems) == 0 && len(warnings) > 0 {
		// Only warnings, no tasks: show warnings directly
		for _, warning := range warnings {
			w.WF.NewItem("Account warning").
				Subtitle(warning).
				Icon(aw.IconWarning).
				Valid(false)
		}
		w.WF.SendFeedback()
		return
	}

	// Has tasks: render them, then append any warnings.
	// RenderGroupedTasks calls SendFeedback, so we add warnings before it.
	grouped := tasks.GroupTasksByTimeframe(allItems, time.Now())
	w.renderGroupedTasksWithWarnings(grouped, warnings)
}

// renderGroupedTasksWithWarnings renders grouped tasks and appends warning items
// before calling SendFeedback. This avoids the double-SendFeedback issue that
// would occur if we called RenderGroupedTasks followed by additional items.
func (w *Workflow) renderGroupedTasksWithWarnings(grouped tasks.GroupedTasks, warnings []string) {
	for _, group := range grouped.Groups {
		icon := iconForTimeframe(group.Timeframe)

		for _, item := range group.Tasks {
			subtitle := buildSubtitle(group.Label, item)

			it := w.WF.NewItem(item.Task.Title).
				Subtitle(subtitle).
				Arg(item.Task.Title).
				Var("listID", item.ListID).
				Var("taskID", item.Task.Id).
				Icon(icon).
				Valid(true)

			if qlPath := w.quicklookFile(item.Task.Id, item.Task.Title, item.Task.Notes); qlPath != "" {
				it.Quicklook(qlPath)
			}

			if item.AccountName != "" {
				it.Var("accountName", item.AccountName)
			}
		}
	}

	for _, warning := range warnings {
		w.WF.NewItem("Account warning").
			Subtitle(warning).
			Icon(aw.IconWarning).
			Valid(false)
	}

	w.WF.SendFeedback()
}

// fetchTasksForCurrentAccount fetches tasks using the current AccountContext.
// Optionally filters by list name. Returns tagged TaskItems.
func (w *Workflow) fetchTasksForCurrentAccount(listFilter string) ([]tasks.TaskItem, error) {
	authConfig, err := w.getAuthenticatedClient()
	if err != nil {
		return nil, fmt.Errorf("auth: %w", err)
	}

	client, err := tasks.NewClient(authConfig.Token, authConfig.Config)
	if err != nil {
		return nil, fmt.Errorf("api client: %w", err)
	}

	if listFilter != "" {
		return client.FetchTasksFromList(listFilter)
	}
	return client.FetchAllTasks()
}

// addToIdeaInbox checks if a newly created task belongs to a configured idea
// list. If so, it writes the task directly to the inbox file and deletes it
// from Google Tasks. Returns true if the task was handled as an idea.
func (w *Workflow) addToIdeaInbox(client *tasks.Client, task *taskapi.Task, listName string) bool {
	inboxPath := os.Getenv("IDEA_INBOX_PATH")
	listNames := ideas.ParseListNames(os.Getenv("IDEA_LIST_NAME"))
	if inboxPath == "" || len(listNames) == 0 {
		return false
	}

	match := false
	lowerListName := strings.ToLower(listName)
	for _, name := range listNames {
		if strings.ToLower(name) == lowerListName {
			match = true
			break
		}
	}
	if !match {
		return false
	}

	accountName := ""
	if w.AccountConfig != nil && w.AccountCtx.Name != "" {
		accountName = w.AccountCtx.Name
	}

	entry := ideas.IdeaEntry{
		Title:       task.Title,
		Date:        ideas.FormatDate(task.Updated),
		Account:     accountName,
		TaskID:      task.Id,
		Description: task.Notes,
	}

	if err := ideas.AppendIdeaEntry(inboxPath, entry); err != nil {
		fmt.Fprintf(os.Stderr, "idea inbox: failed to write: %v\n", err)
		return false
	}

	list, err := client.FindTaskListByName(listName)
	if err == nil && list != nil {
		if delErr := client.DeleteTask(list.Id, task.Id); delErr != nil {
			fmt.Fprintf(os.Stderr, "idea inbox: failed to delete from tasks: %v\n", delErr)
		}
	}

	return true
}

// syncIdeasToInbox syncs ideas from the current account's Ideas list to the
// Obsidian inbox file. Silently returns on any error (FR-009).
func (w *Workflow) syncIdeasToInbox() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintf(os.Stderr, "idea sync panic: %v\n", r)
		}
	}()

	inboxPath := os.Getenv("IDEA_INBOX_PATH")
	listNames := ideas.ParseListNames(os.Getenv("IDEA_LIST_NAME"))
	if inboxPath == "" || len(listNames) == 0 {
		return
	}

	authConfig, err := w.getAuthenticatedClient()
	if err != nil {
		fmt.Fprintf(os.Stderr, "idea sync auth: %v\n", err)
		return
	}

	client, err := tasks.NewClient(authConfig.Token, authConfig.Config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "idea sync client: %v\n", err)
		return
	}

	accountName := ""
	if w.AccountConfig != nil && w.AccountCtx.Name != "" {
		accountName = w.AccountCtx.Name
	}

	count, err := ideas.SyncIdeas(client, accountName, listNames, inboxPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "idea sync: %v\n", err)
		return
	}
	if count > 0 {
		fmt.Fprintf(os.Stderr, "idea sync: synced %d ideas\n", count)
	}
}

// syncIdeasAllAccounts syncs ideas from ALL authenticated accounts' Ideas lists
// to the Obsidian inbox file. Used in multi-account mode so that ideas from every
// account are captured regardless of which account is targeted. Silently skips
// unauthenticated or erroring accounts (FR-009).
func (w *Workflow) syncIdeasAllAccounts() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintf(os.Stderr, "idea sync panic: %v\n", r)
		}
	}()

	inboxPath := os.Getenv("IDEA_INBOX_PATH")
	listNames := ideas.ParseListNames(os.Getenv("IDEA_LIST_NAME"))
	if inboxPath == "" || len(listNames) == 0 {
		return
	}

	for _, name := range w.AccountConfig.AccountNames() {
		ctx, err := auth.ResolveAccount(w.AccountConfig, name)
		if err != nil {
			fmt.Fprintf(os.Stderr, "idea sync: resolve %s: %v\n", name, err)
			continue
		}

		if !auth.TokenExists(ctx.DataDir) {
			continue
		}

		config, err := auth.LoadClientCredentialsFrom(ctx.CredentialsPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "idea sync: credentials %s: %v\n", name, err)
			continue
		}

		token, err := auth.EnsureValidToken(ctx.DataDir, config)
		if err != nil {
			fmt.Fprintf(os.Stderr, "idea sync: token %s: %v\n", name, err)
			continue
		}

		client, err := tasks.NewClient(token, config)
		if err != nil {
			fmt.Fprintf(os.Stderr, "idea sync: client %s: %v\n", name, err)
			continue
		}

		count, err := ideas.SyncIdeas(client, name, listNames, inboxPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "idea sync: %s: %v\n", name, err)
			continue
		}
		if count > 0 {
			fmt.Fprintf(os.Stderr, "idea sync: synced %d ideas from %s\n", count, name)
		}
	}
}

// handleOpen opens Google Tasks in the browser.
func (w *Workflow) handleOpen() {
	if err := tasks.OpenGoogleTasks(w.AccountCtx.ProfileIndex); err != nil {
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

	// In multi-account mode, re-resolve the account context from the
	// accountName variable that was propagated through item vars.
	// This must run before any action dispatch (including create:).
	w.resolveAccountFromEnv()

	// Handle cmd: prefixed actions (login, logout, open) from the Run Script.
	if strings.HasPrefix(actionArg, "cmd:") {
		switch actionArg[4:] {
		case "login":
			w.handleLogin()
		case "logout":
			w.handleLogout()
		case "open":
			w.handleOpen()
		default:
			NotifyError("Google Tasks", fmt.Sprintf("Unknown command: %s", actionArg[4:]))
		}
		return
	}

	// Task creation is routed to the run-create Run Script via the Conditional.
	// This fallback handles it if the routing doesn't match.
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
	w.RenderActionMenu(listID, taskID, os.Getenv("accountName"), w.AccountConfig)
}

// resolveAccountFromEnv checks for an accountName environment variable
// (set via Alfred item vars on merged-list tasks) and re-resolves the
// account context so that action handlers authenticate against the
// correct account, not the default.
func (w *Workflow) resolveAccountFromEnv() {
	if w.AccountConfig == nil {
		return
	}
	accountName := os.Getenv("accountName")
	if accountName == "" {
		return
	}
	ctx, err := auth.ResolveAccount(w.AccountConfig, accountName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "resolving account %q from env: %v\n", accountName, err)
		return
	}
	w.AccountCtx = ctx
}

func (w *Workflow) executeAction(action, listID, taskID string) {

	// "open" action doesn't need auth
	if action == "open" {
		if err := tasks.OpenGoogleTasks(w.AccountCtx.ProfileIndex); err != nil {
			NotifyError("Google Tasks", fmt.Sprintf("Failed to open browser: %v", err))
		}
		return
	}

	// Handle move:{targetAccount} action
	if strings.HasPrefix(action, "move:") {
		w.executeMoveAction(action, listID, taskID)
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

// executeMoveAction handles the "move:{targetAccount}" action by creating the
// task on the target account and deleting it from the source.
func (w *Workflow) executeMoveAction(action, listID, taskID string) {
	targetAccount := strings.TrimPrefix(action, "move:")

	if w.AccountConfig == nil {
		NotifyError("Google Tasks", "Move requires multi-account mode")
		return
	}

	if !w.requireAuth() {
		return
	}

	// Create source client
	sourceAuth, err := w.getAuthenticatedClient()
	if err != nil {
		NotifyError("Google Tasks", fmt.Sprintf("Source auth error: %v", err))
		fmt.Fprintf(os.Stderr, "move source auth error: %v\n", err)
		return
	}

	sourceClient, err := tasks.NewClient(sourceAuth.Token, sourceAuth.Config)
	if err != nil {
		NotifyError("Google Tasks", fmt.Sprintf("Source client error: %v", err))
		fmt.Fprintf(os.Stderr, "move source client error: %v\n", err)
		return
	}

	// Resolve target account
	targetCtx, err := auth.ResolveAccount(w.AccountConfig, targetAccount)
	if err != nil {
		NotifyError("Google Tasks", fmt.Sprintf("Unknown target account: %s", targetAccount))
		fmt.Fprintf(os.Stderr, "move resolve target error: %v\n", err)
		return
	}

	// Create target client
	targetConfig, err := auth.LoadClientCredentialsFrom(targetCtx.CredentialsPath)
	if err != nil {
		NotifyError("Google Tasks", fmt.Sprintf("Target credentials error: %v", err))
		fmt.Fprintf(os.Stderr, "move target credentials error: %v\n", err)
		return
	}

	targetToken, err := auth.EnsureValidToken(targetCtx.DataDir, targetConfig)
	if err != nil {
		NotifyError("Google Tasks", fmt.Sprintf("Target auth error: %v", err))
		fmt.Fprintf(os.Stderr, "move target token error: %v\n", err)
		return
	}

	targetClient, err := tasks.NewClient(targetToken, targetConfig)
	if err != nil {
		NotifyError("Google Tasks", fmt.Sprintf("Target client error: %v", err))
		fmt.Fprintf(os.Stderr, "move target client error: %v\n", err)
		return
	}

	// Fetch the source list name so the target list can be matched by name
	sourceList, err := sourceClient.GetTaskList(listID)
	if err != nil {
		NotifyError("Google Tasks", fmt.Sprintf("Failed to get source list: %v", err))
		fmt.Fprintf(os.Stderr, "move get source list error: %v\n", err)
		return
	}

	// Execute the move
	_, err = sourceClient.MoveTask(listID, taskID, targetClient, sourceList.Title)
	if err != nil {
		if _, ok := errors.AsType[*tasks.PartialMoveError](err); ok {
			NotifyError("Google Tasks", "Task moved but could not delete original. You may have a duplicate.")
			fmt.Fprintf(os.Stderr, "move partial error: %v\n", err)
			return
		}
		NotifyError("Google Tasks", fmt.Sprintf("Failed to move task: %v", err))
		fmt.Fprintf(os.Stderr, "move error: %v\n", err)
		return
	}

	NotifySuccess("Google Tasks", fmt.Sprintf("Task moved to %s", targetAccount))
}
