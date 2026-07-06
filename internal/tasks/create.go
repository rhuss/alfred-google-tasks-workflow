package tasks

import (
	"fmt"
	"time"

	"github.com/rhuss/alfred-google-tasks-workflow/internal/input"
	taskapi "google.golang.org/api/tasks/v1"
)

// CreateTaskFromInput parses user input and creates a task via the Google Tasks API.
// Returns the created task and the list name it was added to.
func (c *Client) CreateTaskFromInput(rawInput string) (*taskapi.Task, string, error) {
	return c.CreateTaskFromInputWithTime(rawInput, time.Now())
}

// CreateTaskFromInputWithTime is like CreateTaskFromInput but uses the given
// time as reference for date parsing.
func (c *Client) CreateTaskFromInputWithTime(rawInput string, now time.Time) (*taskapi.Task, string, error) {
	parsed := input.ParseWithTime(rawInput, now)

	if parsed.Title == "" {
		return nil, "", fmt.Errorf("task title cannot be empty")
	}

	// Resolve the target task list
	var taskList *taskapi.TaskList
	var err error

	if parsed.ListName != "" {
		taskList, err = c.ResolveTaskList(parsed.ListName)
	} else {
		taskList, err = c.GetDefaultTaskList()
	}
	if err != nil {
		return nil, "", fmt.Errorf("resolving task list: %w", err)
	}

	// Build the task
	task := &taskapi.Task{
		Title: parsed.Title,
	}

	// Convert date to RFC 3339 format if present
	if parsed.Date != nil {
		task.Due = DateToRFC3339(*parsed.Date)
	}

	// Create the task
	created, err := c.InsertTask(taskList.Id, task)
	if err != nil {
		return nil, "", fmt.Errorf("creating task: %w", err)
	}

	return created, taskList.Title, nil
}

// DateToRFC3339 converts a time.Time to the RFC 3339 format used by Google Tasks API.
// Google Tasks uses midnight UTC for due dates.
func DateToRFC3339(t time.Time) string {
	// Normalize to midnight UTC
	utcMidnight := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
	return utcMidnight.Format(time.RFC3339)
}
