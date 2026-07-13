package tasks

import (
	"fmt"
	"os/exec"

	taskapi "google.golang.org/api/tasks/v1"
)

// PartialMoveError indicates that the task was created on the target account
// but the original could not be deleted from the source. The caller should
// warn the user about a potential duplicate.
type PartialMoveError struct {
	CreatedTask *taskapi.Task
	DeleteErr   error
}

func (e *PartialMoveError) Error() string {
	return fmt.Sprintf("task created on target but delete failed: %v", e.DeleteErr)
}

func (e *PartialMoveError) Unwrap() error {
	return e.DeleteErr
}

// MoveTask moves a task from this client's account to the target account.
// It fetches the source task details, resolves (or creates) the matching list
// on the target, creates the task there with Title/Due/Notes preserved, and
// deletes the original. Returns the newly created task.
//
// If the create succeeds but the delete fails, a *PartialMoveError is returned
// so the caller can distinguish partial from full failure.
func (c *Client) MoveTask(sourceListID, taskID string, targetClient *Client, targetListName string) (*taskapi.Task, error) {
	// 1. Fetch source task details
	sourceTask, err := c.GetTask(sourceListID, taskID)
	if err != nil {
		return nil, fmt.Errorf("fetching source task: %w", err)
	}

	// 2. Resolve target list (auto-creates if needed)
	targetList, err := targetClient.ResolveTaskList(targetListName)
	if err != nil {
		return nil, fmt.Errorf("resolving target list %q: %w", targetListName, err)
	}

	// 3. Create new task on target with preserved properties
	newTask := &taskapi.Task{
		Title: sourceTask.Title,
	}
	if sourceTask.Due != "" {
		newTask.Due = sourceTask.Due
	}
	if sourceTask.Notes != "" {
		newTask.Notes = sourceTask.Notes
	}

	created, err := targetClient.InsertTask(targetList.Id, newTask)
	if err != nil {
		return nil, fmt.Errorf("creating task on target: %w", err)
	}

	// 4. Delete original from source
	if err := c.DeleteTask(sourceListID, taskID); err != nil {
		return created, &PartialMoveError{
			CreatedTask: created,
			DeleteErr:   err,
		}
	}

	return created, nil
}

// CompleteTaskByID marks the specified task as completed.
func (c *Client) CompleteTaskByID(listID, taskID string) error {
	if err := c.CompleteTask(listID, taskID); err != nil {
		return fmt.Errorf("completing task: %w", err)
	}
	return nil
}

// DeleteTaskByID removes the specified task.
func (c *Client) DeleteTaskByID(listID, taskID string) error {
	if err := c.DeleteTask(listID, taskID); err != nil {
		return fmt.Errorf("deleting task: %w", err)
	}
	return nil
}

// OpenGoogleTasks opens the Google Tasks web UI in the default browser.
// If profileIndex >= 0, appends ?authuser=N to select the correct Google profile.
// Pass -1 to open without an authuser parameter (single-account mode).
func OpenGoogleTasks(profileIndex int) error {
	url := "https://tasks.google.com/"
	if profileIndex >= 0 {
		url = fmt.Sprintf("https://tasks.google.com/?authuser=%d", profileIndex)
	}
	cmd := exec.Command("open", url)
	return cmd.Run()
}
