package tasks

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/oauth2"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"
	tasks "google.golang.org/api/tasks/v1"
)

// Client wraps the Google Tasks API service.
type Client struct {
	service *tasks.Service
}

// NewClient creates an authenticated Google Tasks API client.
func NewClient(token *oauth2.Token, config *oauth2.Config) (*Client, error) {
	ctx := context.Background()
	tokenSource := config.TokenSource(ctx, token)
	httpClient := oauth2.NewClient(ctx, tokenSource)

	service, err := tasks.NewService(ctx, option.WithHTTPClient(httpClient))
	if err != nil {
		return nil, fmt.Errorf("creating tasks service: %w", err)
	}

	return &Client{service: service}, nil
}

// ListTaskLists returns all task lists for the authenticated user.
func (c *Client) ListTaskLists() ([]*tasks.TaskList, error) {
	var allLists []*tasks.TaskList
	pageToken := ""

	for {
		call := c.service.Tasklists.List().MaxResults(100)
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}

		resp, err := call.Do()
		if err != nil {
			return nil, wrapAPIError(err)
		}

		allLists = append(allLists, resp.Items...)

		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}

	return allLists, nil
}

// CreateTaskList creates a new task list with the given title.
func (c *Client) CreateTaskList(title string) (*tasks.TaskList, error) {
	taskList := &tasks.TaskList{Title: title}

	created, err := c.service.Tasklists.Insert(taskList).Do()
	if err != nil {
		return nil, wrapAPIError(err)
	}

	return created, nil
}

// ListTasks returns all incomplete tasks from a task list.
func (c *Client) ListTasks(listID string) ([]*tasks.Task, error) {
	var allTasks []*tasks.Task
	pageToken := ""

	for {
		call := c.service.Tasks.List(listID).
			MaxResults(100).
			ShowCompleted(false).
			ShowHidden(false)
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}

		resp, err := call.Do()
		if err != nil {
			return nil, wrapAPIError(err)
		}

		allTasks = append(allTasks, resp.Items...)

		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}

	return allTasks, nil
}

// InsertTask creates a new task in the specified task list.
func (c *Client) InsertTask(listID string, task *tasks.Task) (*tasks.Task, error) {
	created, err := c.service.Tasks.Insert(listID, task).Do()
	if err != nil {
		return nil, wrapAPIError(err)
	}

	return created, nil
}

// CompleteTask marks a task as completed.
func (c *Client) CompleteTask(listID, taskID string) error {
	task := &tasks.Task{
		Id:     taskID,
		Status: "completed",
	}

	_, err := c.service.Tasks.Patch(listID, taskID, task).Do()
	if err != nil {
		return wrapAPIError(err)
	}

	return nil
}

// DeleteTask removes a task from a task list.
func (c *Client) DeleteTask(listID, taskID string) error {
	err := c.service.Tasks.Delete(listID, taskID).Do()
	if err != nil {
		return wrapAPIError(err)
	}

	return nil
}

// FindTaskListByName searches for a task list with the given name (case-insensitive).
// Returns nil if no matching list is found.
func (c *Client) FindTaskListByName(name string) (*tasks.TaskList, error) {
	lists, err := c.ListTaskLists()
	if err != nil {
		return nil, err
	}

	normalizedName := strings.ToLower(strings.TrimSpace(name))
	for _, list := range lists {
		if strings.ToLower(list.Title) == normalizedName {
			return list, nil
		}
	}

	return nil, nil
}

// ResolveTaskList finds a task list by name, creating it if it does not exist.
func (c *Client) ResolveTaskList(name string) (*tasks.TaskList, error) {
	list, err := c.FindTaskListByName(name)
	if err != nil {
		return nil, err
	}

	if list != nil {
		return list, nil
	}

	// List not found, create it
	return c.CreateTaskList(name)
}

// GetDefaultTaskList returns the user's first (default) task list.
func (c *Client) GetDefaultTaskList() (*tasks.TaskList, error) {
	lists, err := c.ListTaskLists()
	if err != nil {
		return nil, err
	}

	if len(lists) == 0 {
		// No lists exist, create a default one
		return c.CreateTaskList("My Tasks")
	}

	return lists[0], nil
}

// wrapAPIError converts Google API errors into user-friendly messages.
func wrapAPIError(err error) error {
	if err == nil {
		return nil
	}

	// Check for Google API errors using proper type assertion
	var apiErr *googleapi.Error
	if ok := isGoogleAPIError(err, &apiErr); ok {
		switch apiErr.Code {
		case 429:
			return fmt.Errorf("rate limit reached, try again in a moment")
		case 401, 403:
			return fmt.Errorf("authentication error: run 'gt login' again")
		case 404:
			return fmt.Errorf("not found: the requested resource does not exist")
		default:
			return fmt.Errorf("Google Tasks API error (HTTP %d): %s", apiErr.Code, apiErr.Message)
		}
	}

	// Check for network errors by string matching (no structured type available)
	errStr := err.Error()
	if strings.Contains(errStr, "no such host") ||
		strings.Contains(errStr, "connection refused") ||
		strings.Contains(errStr, "network is unreachable") ||
		strings.Contains(errStr, "timeout") ||
		strings.Contains(errStr, "dial tcp") {
		return fmt.Errorf("no internet connection: %w", err)
	}

	return fmt.Errorf("Google Tasks API error: %w", err)
}

// isGoogleAPIError checks if an error is a Google API error using errors.As.
func isGoogleAPIError(err error, target **googleapi.Error) bool {
	return errors.As(err, target)
}
