package ideas

import (
	"fmt"
	"os"
	"strings"

	taskapi "google.golang.org/api/tasks/v1"
)

// TasksClient defines the Google Tasks operations needed by the sync logic.
type TasksClient interface {
	FindTaskListByName(name string) (*taskapi.TaskList, error)
	ListTasks(listID string) ([]*taskapi.Task, error)
	DeleteTask(listID, taskID string) error
}

// SyncIdeas fetches tasks from the named Ideas list, appends new ones
// to the inbox file (deduplicating by TaskID), and deletes synced tasks
// from Google Tasks. Returns the count of newly synced ideas.
func SyncIdeas(client TasksClient, accountName, listName, inboxPath string) (int, error) {
	if inboxPath == "" || listName == "" {
		return 0, nil
	}

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

	if len(apiTasks) == 0 {
		return 0, nil
	}

	existingIDs, err := ReadSyncedTaskIDs(inboxPath)
	if err != nil {
		return 0, fmt.Errorf("reading synced task IDs: %w", err)
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

		if err := client.DeleteTask(list.Id, task.Id); err != nil {
			fmt.Fprintf(os.Stderr, "idea sync: failed to delete task %q: %v\n", task.Title, err)
		}

		synced++
	}

	return synced, nil
}
