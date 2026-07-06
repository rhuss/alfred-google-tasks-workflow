package alfred

import (
	"fmt"
	"strings"

	aw "github.com/deanishe/awgo"

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
			arg := fmt.Sprintf("%s:%s:%s", item.ListID, item.Task.Id, item.Task.Title)

			w.WF.NewItem(item.Task.Title).
				Subtitle(subtitle).
				Arg(arg).
				Icon(icon).
				Valid(true)
		}
	}

	w.WF.SendFeedback()
}

func buildSubtitle(groupLabel string, item tasks.TaskItem) string {
	subtitle := fmt.Sprintf("[%s] %s", groupLabel, item.ListName)

	if item.Task.Due != "" && len(item.Task.Due) >= 10 {
		subtitle += fmt.Sprintf(" - due %s", item.Task.Due[:10])
	}

	return subtitle
}

func (w *Workflow) RenderActionMenu(taskArg string) {
	// taskArg format: listID:taskID:title
	// Extract listID:taskID for action args
	parts := strings.SplitN(taskArg, ":", 3)
	actionRef := taskArg
	if len(parts) >= 2 {
		actionRef = parts[0] + ":" + parts[1]
	}

	w.WF.NewItem("Complete Task").
		Subtitle("Mark this task as done").
		Arg("complete|" + actionRef).
		Icon(iconComplete).
		Valid(true)

	w.WF.NewItem("Delete Task").
		Subtitle("Permanently delete this task").
		Arg("delete|" + actionRef).
		Icon(iconDelete).
		Valid(true)

	w.WF.NewItem("Open in Browser").
		Subtitle("Open Google Tasks in your browser").
		Arg("open|" + actionRef).
		Icon(iconOpen).
		Valid(true)

	w.WF.SendFeedback()
}
