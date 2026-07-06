package tasks

import (
	"fmt"
	"os/exec"
)

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
func OpenGoogleTasks() error {
	cmd := exec.Command("open", "https://tasks.google.com/")
	return cmd.Run()
}
