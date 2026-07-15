package ideas

import (
	"fmt"
	"os"
	"strings"

	taskapi "google.golang.org/api/tasks/v1"
)

// ParseListNames splits a comma-separated string of list names into a
// trimmed, non-empty slice. Returns nil if the input is empty.
func ParseListNames(csv string) []string {
	if strings.TrimSpace(csv) == "" {
		return nil
	}
	parts := strings.Split(csv, ",")
	var names []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			names = append(names, p)
		}
	}
	return names
}

// TasksClient defines the Google Tasks operations needed by the sync logic.
type TasksClient interface {
	FindTaskListByName(name string) (*taskapi.TaskList, error)
	ListTasks(listID string) ([]*taskapi.Task, error)
	DeleteTask(listID, taskID string) error
}

// SyncIdeas fetches tasks from the named Ideas lists, appends new ones
// to the inbox file (deduplicating by TaskID), and deletes synced tasks
// from Google Tasks. listNames is a slice of list names to check (e.g.,
// "Ideas", "Ideen", "Idea"). Returns the count of newly synced ideas.
func SyncIdeas(client TasksClient, accountName string, listNames []string, inboxPath string) (int, error) {
	if inboxPath == "" || len(listNames) == 0 {
		return 0, nil
	}

	if err := EnsureInboxFile(inboxPath); err != nil {
		return 0, fmt.Errorf("creating inbox file: %w", err)
	}

	existingIDs, err := ReadSyncedTaskIDs(inboxPath)
	if err != nil {
		return 0, fmt.Errorf("reading synced task IDs: %w", err)
	}

	synced := 0
	for _, listName := range listNames {
		n, err := syncFromList(client, accountName, listName, inboxPath, existingIDs)
		if err != nil {
			fmt.Fprintf(os.Stderr, "idea sync: list %q: %v\n", listName, err)
			continue
		}
		synced += n
	}

	return synced, nil
}

func syncFromList(client TasksClient, accountName, listName, inboxPath string, existingIDs map[string]bool) (int, error) {
	list, err := client.FindTaskListByName(listName)
	if err != nil {
		return 0, fmt.Errorf("finding ideas list: %w", err)
	}
	if list == nil {
		return 0, nil
	}

	apiTasks, err := client.ListTasks(list.Id)
	if err != nil {
		return 0, fmt.Errorf("listing ideas: %w", err)
	}

	synced := 0
	for _, task := range apiTasks {
		if strings.TrimSpace(task.Title) == "" {
			continue
		}
		if existingIDs[task.Id] {
			continue
		}

		entry := IdeaEntry{
			Title:       task.Title,
			Date:        FormatDate(task.Updated),
			Account:     accountName,
			TaskID:      task.Id,
			Description: task.Notes,
		}

		if err := AppendIdeaEntry(inboxPath, entry); err != nil {
			fmt.Fprintf(os.Stderr, "idea sync: failed to write entry %q: %v\n", task.Title, err)
			continue
		}

		existingIDs[task.Id] = true

		if err := client.DeleteTask(list.Id, task.Id); err != nil {
			fmt.Fprintf(os.Stderr, "idea sync: failed to delete task %q: %v\n", task.Title, err)
		}

		synced++
	}

	return synced, nil
}
