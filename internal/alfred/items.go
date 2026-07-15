package alfred

import (
	"crypto/md5"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	aw "github.com/deanishe/awgo"

	"github.com/rhuss/alfred-google-tasks-workflow/internal/auth"
	"github.com/rhuss/alfred-google-tasks-workflow/internal/tasks"
)

var (
	iconOverdue  = &aw.Icon{Value: "icons/overdue.png"}
	iconToday    = &aw.Icon{Value: "icons/today.png"}
	iconThisWeek = &aw.Icon{Value: "icons/thisweek.png"}
	iconLater    = &aw.Icon{Value: "icons/later.png"}
	iconNoDate   = &aw.Icon{Value: "icons/nodate.png"}
	iconComplete = &aw.Icon{Value: "icons/complete.png"}
	iconDelete   = &aw.Icon{Value: "icons/delete.png"}
	iconOpen     = &aw.Icon{Value: "icons/open.png"}
	iconMove     = &aw.Icon{Value: "icons/move.png"}
	iconIdea     = &aw.Icon{Value: "icons/idea.png"}
)

func iconForTimeframe(tf tasks.Timeframe) *aw.Icon {
	switch tf {
	case tasks.TimeframeOverdue:
		return iconOverdue
	case tasks.TimeframeToday:
		return iconToday
	case tasks.TimeframeThisWeek:
		return iconThisWeek
	case tasks.TimeframeLater:
		return iconLater
	case tasks.TimeframeNoDate:
		return iconNoDate
	default:
		return iconLater
	}
}

func (w *Workflow) RenderGroupedTasks(grouped tasks.GroupedTasks) {
	if len(grouped.Groups) == 0 {
		w.WF.NewItem("No tasks found").
			Subtitle("Use 'gt add' to create a new task").
			Icon(iconNoDate).
			Valid(false)
		w.WF.SendFeedback()
		return
	}

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

	w.WF.SendFeedback()
}

// quicklookFile writes the task notes to a temp file and returns the path.
// Returns empty string if notes are empty. Files are written to Alfred's
// cache directory using a hash-based name to avoid conflicts.
func (w *Workflow) quicklookFile(taskID, title, notes string) string {
	if strings.TrimSpace(notes) == "" {
		return ""
	}
	cacheDir := w.WF.CacheDir()
	qlDir := filepath.Join(cacheDir, "quicklook")
	_ = os.MkdirAll(qlDir, 0o755)

	hash := fmt.Sprintf("%x", md5.Sum([]byte(taskID)))
	path := filepath.Join(qlDir, hash+".txt")
	content := title + "\n" + strings.Repeat("=", len(title)) + "\n\n" + notes
	_ = os.WriteFile(path, []byte(content), 0o644)
	return path
}

func buildSubtitle(groupLabel string, item tasks.TaskItem) string {
	listPart := item.ListName
	if item.AccountName != "" {
		listPart += fmt.Sprintf(" (%s)", item.AccountName)
	}

	subtitle := fmt.Sprintf("[%s] %s", groupLabel, listPart)

	if item.Task.Due != "" && len(item.Task.Due) >= 10 {
		subtitle += fmt.Sprintf(" - due %s", item.Task.Due[:10])
	}

	if strings.TrimSpace(item.Task.Notes) != "" {
		subtitle += " *"
	}

	return subtitle
}

func (w *Workflow) RenderActionMenu(listID, taskID, accountName string, accountConfig *auth.AccountConfig) {
	actionRef := listID + ":" + taskID

	completeItem := w.WF.NewItem("Complete Task").
		Subtitle("Mark this task as done").
		Arg("complete|" + actionRef).
		Icon(iconComplete).
		Valid(true)

	deleteItem := w.WF.NewItem("Delete Task").
		Subtitle("Permanently delete this task").
		Arg("delete|" + actionRef).
		Icon(iconDelete).
		Valid(true)

	openItem := w.WF.NewItem("Open in Browser").
		Subtitle("Open Google Tasks in your browser").
		Arg("open|" + actionRef).
		Icon(iconOpen).
		Valid(true)

	if accountName != "" {
		completeItem.Var("accountName", accountName)
		deleteItem.Var("accountName", accountName)
		openItem.Var("accountName", accountName)
	}

	// Add "Move to Inbox" when idea inbox is configured
	if inboxPath := os.Getenv("IDEA_INBOX_PATH"); inboxPath != "" {
		inboxItem := w.WF.NewItem("Move to Inbox").
			Subtitle("Capture this task as an idea in Obsidian").
			Arg("inbox|" + actionRef).
			Icon(iconIdea).
			Valid(true)
		if accountName != "" {
			inboxItem.Var("accountName", accountName)
		}
	}

	// Add "Move to {target}" entries when multi-account mode is active
	if accountConfig != nil && len(accountConfig.Accounts) >= 2 {
		// When accountName is empty (single-account listing in multi-account mode),
		// fall back to the default account name for the self-exclusion check.
		effectiveAccount := accountName
		if effectiveAccount == "" {
			effectiveAccount = accountConfig.Default
			if effectiveAccount == "" {
				effectiveAccount = accountConfig.AccountNames()[0]
			}
		}
		for _, targetName := range accountConfig.AccountNames() {
			if targetName == effectiveAccount {
				continue
			}
			targetCtx, err := auth.ResolveAccount(accountConfig, targetName)
			if err != nil {
				continue
			}
			if !auth.TokenExists(targetCtx.DataDir) {
				continue
			}
			moveItem := w.WF.NewItem(fmt.Sprintf("Move to %s", targetName)).
				Subtitle(fmt.Sprintf("Move this task to %s account", targetName)).
				Arg(fmt.Sprintf("move:%s|%s", targetName, actionRef)).
				Icon(iconMove).
				Valid(true)
			moveItem.Var("accountName", effectiveAccount)
		}
	}

	w.WF.SendFeedback()
}
