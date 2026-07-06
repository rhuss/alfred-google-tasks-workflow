package tasks

import (
	"sort"
	"time"

	taskapi "google.golang.org/api/tasks/v1"
)

// Timeframe represents a grouping category for tasks by due date.
type Timeframe int

const (
	TimeframeOverdue Timeframe = iota
	TimeframeToday
	TimeframeThisWeek
	TimeframeLater
	TimeframeNoDate
)

// TimeframeLabel returns the display label for a timeframe group.
func TimeframeLabel(tf Timeframe) string {
	switch tf {
	case TimeframeOverdue:
		return "Overdue"
	case TimeframeToday:
		return "Today"
	case TimeframeThisWeek:
		return "This Week"
	case TimeframeLater:
		return "Later"
	case TimeframeNoDate:
		return "No Date"
	default:
		return "Unknown"
	}
}

// TaskItem is an enriched task that includes the list name for display.
type TaskItem struct {
	Task     *taskapi.Task
	ListName string
	ListID   string
}

// GroupedTasks holds tasks organized by timeframe.
type GroupedTasks struct {
	Groups []TaskGroup
}

// TaskGroup is a single timeframe group with its tasks.
type TaskGroup struct {
	Timeframe Timeframe
	Label     string
	Tasks     []TaskItem
}

// FetchAllTasks retrieves tasks from all task lists and returns them as TaskItems.
func (c *Client) FetchAllTasks() ([]TaskItem, error) {
	lists, err := c.ListTaskLists()
	if err != nil {
		return nil, err
	}

	var allItems []TaskItem
	for _, list := range lists {
		tasks, err := c.ListTasks(list.Id)
		if err != nil {
			return nil, err
		}
		for _, task := range tasks {
			allItems = append(allItems, TaskItem{
				Task:     task,
				ListName: list.Title,
				ListID:   list.Id,
			})
		}
	}

	return allItems, nil
}

// FetchTasksFromList retrieves tasks from a specific named list.
func (c *Client) FetchTasksFromList(listName string) ([]TaskItem, error) {
	list, err := c.FindTaskListByName(listName)
	if err != nil {
		return nil, err
	}
	if list == nil {
		return nil, nil
	}

	tasks, err := c.ListTasks(list.Id)
	if err != nil {
		return nil, err
	}

	var items []TaskItem
	for _, task := range tasks {
		items = append(items, TaskItem{
			Task:     task,
			ListName: list.Title,
			ListID:   list.Id,
		})
	}

	return items, nil
}

// ClassifyTimeframe determines which timeframe group a task belongs to.
func ClassifyTimeframe(dueStr string, now time.Time) Timeframe {
	if dueStr == "" {
		return TimeframeNoDate
	}

	due, err := time.Parse(time.RFC3339, dueStr)
	if err != nil {
		return TimeframeNoDate
	}

	// Normalize to date-only comparisons
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	dueDate := time.Date(due.Year(), due.Month(), due.Day(), 0, 0, 0, 0, time.UTC)

	if dueDate.Before(today) {
		return TimeframeOverdue
	}
	if dueDate.Equal(today) {
		return TimeframeToday
	}

	// "This Week" means within the next 7 days (including today)
	endOfWeek := today.AddDate(0, 0, 7)
	if dueDate.Before(endOfWeek) {
		return TimeframeThisWeek
	}

	return TimeframeLater
}

// GroupTasksByTimeframe organizes tasks into timeframe groups.
func GroupTasksByTimeframe(items []TaskItem, now time.Time) GroupedTasks {
	groups := make(map[Timeframe][]TaskItem)

	for _, item := range items {
		tf := ClassifyTimeframe(item.Task.Due, now)
		groups[tf] = append(groups[tf], item)
	}

	// Sort tasks within each group by due date
	for tf := range groups {
		sortTaskItems(groups[tf])
	}

	// Build ordered result
	var result []TaskGroup
	order := []Timeframe{TimeframeOverdue, TimeframeToday, TimeframeThisWeek, TimeframeLater, TimeframeNoDate}

	for _, tf := range order {
		if tasks, ok := groups[tf]; ok && len(tasks) > 0 {
			result = append(result, TaskGroup{
				Timeframe: tf,
				Label:     TimeframeLabel(tf),
				Tasks:     tasks,
			})
		}
	}

	return GroupedTasks{Groups: result}
}

// sortTaskItems sorts task items by due date (no-date items at the end).
func sortTaskItems(items []TaskItem) {
	sort.Slice(items, func(i, j int) bool {
		di := items[i].Task.Due
		dj := items[j].Task.Due

		// No-date tasks go after dated tasks
		if di == "" && dj == "" {
			return items[i].Task.Title < items[j].Task.Title
		}
		if di == "" {
			return false
		}
		if dj == "" {
			return true
		}

		return di < dj
	})
}
