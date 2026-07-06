package alfred

import (
	"fmt"

	aw "github.com/deanishe/awgo"

	"github.com/rhuss/alfred-google-tasks-workflow/internal/tasks"
)

// iconForTimeframe returns an appropriate icon for the timeframe group.
func iconForTimeframe(tf tasks.Timeframe) *aw.Icon {
	switch tf {
	case tasks.TimeframeOverdue:
		return aw.IconError
	case tasks.TimeframeToday:
		return aw.IconFavorite
	case tasks.TimeframeThisWeek:
		return aw.IconClock
	case tasks.TimeframeLater:
		return aw.IconNote
	case tasks.TimeframeNoDate:
		return aw.IconInfo
	default:
		return aw.IconNote
	}
}

// RenderGroupedTasks converts grouped tasks into Alfred Script Filter items.
func (w *Workflow) RenderGroupedTasks(grouped tasks.GroupedTasks) {
	if len(grouped.Groups) == 0 {
		w.WF.NewItem("No tasks found").
			Subtitle("Use 'gt add' to create a new task").
			Icon(aw.IconInfo).
			Valid(false)
		w.WF.SendFeedback()
		return
	}

	for _, group := range grouped.Groups {
		icon := iconForTimeframe(group.Timeframe)

		for _, item := range group.Tasks {
			subtitle := buildSubtitle(group.Label, item)
			// arg format: listID:taskID for action handling
			arg := fmt.Sprintf("%s:%s", item.ListID, item.Task.Id)

			w.WF.NewItem(item.Task.Title).
				Subtitle(subtitle).
				Arg(arg).
				Icon(icon).
				Valid(true)
		}
	}

	w.WF.SendFeedback()
}

// buildSubtitle creates the subtitle string showing group, list name, and due date.
func buildSubtitle(groupLabel string, item tasks.TaskItem) string {
	subtitle := fmt.Sprintf("[%s] %s", groupLabel, item.ListName)

	if item.Task.Due != "" && len(item.Task.Due) >= 10 {
		subtitle += fmt.Sprintf(" - due %s", item.Task.Due[:10])
	}

	return subtitle
}

// RenderActionMenu shows the action sub-menu for a selected task.
// The taskArg format is "listID:taskID".
func (w *Workflow) RenderActionMenu(taskArg string) {
	w.WF.NewItem("Complete Task").
		Subtitle("Mark this task as done").
		Arg("complete:" + taskArg).
		Icon(aw.IconFavorite).
		Valid(true)

	w.WF.NewItem("Delete Task").
		Subtitle("Permanently delete this task").
		Arg("delete:" + taskArg).
		Icon(aw.IconTrash).
		Valid(true)

	w.WF.NewItem("Open in Browser").
		Subtitle("Open Google Tasks in your browser").
		Arg("open:" + taskArg).
		Icon(aw.IconWeb).
		Valid(true)

	w.WF.SendFeedback()
}
