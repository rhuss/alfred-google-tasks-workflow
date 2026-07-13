package ideas

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	taskapi "google.golang.org/api/tasks/v1"
)

type mockTasksClient struct {
	taskLists  map[string]*taskapi.TaskList
	tasks      map[string][]*taskapi.Task
	deleted    []deleteCall
	findErr    error
	listErr    error
	deleteErr  error
}

type deleteCall struct {
	listID string
	taskID string
}

func (m *mockTasksClient) FindTaskListByName(name string) (*taskapi.TaskList, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	return m.taskLists[name], nil
}

func (m *mockTasksClient) ListTasks(listID string) ([]*taskapi.Task, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	return m.tasks[listID], nil
}

func (m *mockTasksClient) DeleteTask(listID, taskID string) error {
	m.deleted = append(m.deleted, deleteCall{listID, taskID})
	if m.deleteErr != nil {
		return m.deleteErr
	}
	return nil
}

func TestSyncIdeas_NewTasks(t *testing.T) {
	dir := t.TempDir()
	inboxPath := filepath.Join(dir, "inbox.md")

	client := &mockTasksClient{
		taskLists: map[string]*taskapi.TaskList{
			"Ideas": {Id: "list1", Title: "Ideas"},
		},
		tasks: map[string][]*taskapi.Task{
			"list1": {
				{Id: "task1", Title: "Build a birdhouse", Updated: "2026-07-13T10:00:00Z", Notes: "Use cedar wood"},
				{Id: "task2", Title: "Learn origami", Updated: "2026-07-12T08:00:00Z"},
			},
		},
	}

	count, err := SyncIdeas(client, "personal", "Ideas", inboxPath)
	if err != nil {
		t.Fatalf("SyncIdeas() error = %v", err)
	}
	if count != 2 {
		t.Errorf("SyncIdeas() count = %d, want 2", count)
	}

	data, err := os.ReadFile(inboxPath)
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)

	expected := `# Idea Inbox

### Build a birdhouse
- Date: 2026-07-13
- Account: personal
- TaskID: task1

Use cedar wood

### Learn origami
- Date: 2026-07-12
- Account: personal
- TaskID: task2

`
	if content != expected {
		t.Errorf("inbox content mismatch.\ngot:\n%s\nwant:\n%s", content, expected)
	}

	if len(client.deleted) != 2 {
		t.Fatalf("expected 2 deletes, got %d", len(client.deleted))
	}
	if client.deleted[0].taskID != "task1" || client.deleted[1].taskID != "task2" {
		t.Errorf("unexpected delete order: %v", client.deleted)
	}
}

func TestSyncIdeas_DedupSkip(t *testing.T) {
	dir := t.TempDir()
	inboxPath := filepath.Join(dir, "inbox.md")

	existing := `# Idea Inbox

### Already synced
- Date: 2026-07-10
- TaskID: task1

`
	if err := os.WriteFile(inboxPath, []byte(existing), 0644); err != nil {
		t.Fatal(err)
	}

	client := &mockTasksClient{
		taskLists: map[string]*taskapi.TaskList{
			"Ideas": {Id: "list1", Title: "Ideas"},
		},
		tasks: map[string][]*taskapi.Task{
			"list1": {
				{Id: "task1", Title: "Already synced", Updated: "2026-07-10T10:00:00Z"},
				{Id: "task2", Title: "Brand new idea", Updated: "2026-07-13T10:00:00Z"},
			},
		},
	}

	count, err := SyncIdeas(client, "", "Ideas", inboxPath)
	if err != nil {
		t.Fatalf("SyncIdeas() error = %v", err)
	}
	if count != 1 {
		t.Errorf("SyncIdeas() count = %d, want 1 (task1 should be deduped)", count)
	}

	ids, err := ReadSyncedTaskIDs(inboxPath)
	if err != nil {
		t.Fatal(err)
	}
	if !ids["task1"] || !ids["task2"] {
		t.Error("expected both task1 and task2 in inbox after sync")
	}

	if len(client.deleted) != 1 {
		t.Fatalf("expected 1 delete (only new task), got %d", len(client.deleted))
	}
	if client.deleted[0].taskID != "task2" {
		t.Errorf("expected task2 to be deleted, got %s", client.deleted[0].taskID)
	}
}

func TestSyncIdeas_DeleteAfterWrite(t *testing.T) {
	dir := t.TempDir()
	inboxPath := filepath.Join(dir, "inbox.md")

	client := &mockTasksClient{
		taskLists: map[string]*taskapi.TaskList{
			"Ideas": {Id: "list1", Title: "Ideas"},
		},
		tasks: map[string][]*taskapi.Task{
			"list1": {
				{Id: "task1", Title: "Important idea", Updated: "2026-07-13T10:00:00Z"},
			},
		},
	}

	count, err := SyncIdeas(client, "work", "Ideas", inboxPath)
	if err != nil {
		t.Fatalf("SyncIdeas() error = %v", err)
	}
	if count != 1 {
		t.Errorf("SyncIdeas() count = %d, want 1", count)
	}

	ids, _ := ReadSyncedTaskIDs(inboxPath)
	if !ids["task1"] {
		t.Error("task1 should be written to inbox before delete")
	}

	if len(client.deleted) != 1 || client.deleted[0].taskID != "task1" {
		t.Error("task1 should be deleted after successful write")
	}
}

func TestSyncIdeas_MissingListReturnsZero(t *testing.T) {
	dir := t.TempDir()
	inboxPath := filepath.Join(dir, "inbox.md")

	client := &mockTasksClient{
		taskLists: map[string]*taskapi.TaskList{},
	}

	count, err := SyncIdeas(client, "", "Ideas", inboxPath)
	if err != nil {
		t.Fatalf("SyncIdeas() error = %v", err)
	}
	if count != 0 {
		t.Errorf("SyncIdeas() count = %d, want 0 for missing list", count)
	}

	if _, err := os.Stat(inboxPath); !os.IsNotExist(err) {
		t.Error("inbox file should not be created when list is missing")
	}
}

func TestSyncIdeas_FindListError(t *testing.T) {
	dir := t.TempDir()
	inboxPath := filepath.Join(dir, "inbox.md")

	client := &mockTasksClient{
		findErr: fmt.Errorf("network error"),
	}

	_, err := SyncIdeas(client, "", "Ideas", inboxPath)
	if err == nil {
		t.Fatal("expected error from FindTaskListByName failure")
	}
}

func TestSyncIdeas_ListTasksError(t *testing.T) {
	dir := t.TempDir()
	inboxPath := filepath.Join(dir, "inbox.md")

	client := &mockTasksClient{
		taskLists: map[string]*taskapi.TaskList{
			"Ideas": {Id: "list1", Title: "Ideas"},
		},
		listErr: fmt.Errorf("api error"),
	}

	_, err := SyncIdeas(client, "", "Ideas", inboxPath)
	if err == nil {
		t.Fatal("expected error from ListTasks failure")
	}
}

func TestSyncIdeas_DeleteErrorContinues(t *testing.T) {
	dir := t.TempDir()
	inboxPath := filepath.Join(dir, "inbox.md")

	client := &mockTasksClient{
		taskLists: map[string]*taskapi.TaskList{
			"Ideas": {Id: "list1", Title: "Ideas"},
		},
		tasks: map[string][]*taskapi.Task{
			"list1": {
				{Id: "task1", Title: "Idea one", Updated: "2026-07-13T10:00:00Z"},
				{Id: "task2", Title: "Idea two", Updated: "2026-07-13T10:00:00Z"},
			},
		},
		deleteErr: fmt.Errorf("delete failed"),
	}

	count, err := SyncIdeas(client, "", "Ideas", inboxPath)
	if err != nil {
		t.Fatalf("SyncIdeas() should not return error on delete failure, got %v", err)
	}
	if count != 2 {
		t.Errorf("SyncIdeas() count = %d, want 2 (both should be counted despite delete failure)", count)
	}

	ids, _ := ReadSyncedTaskIDs(inboxPath)
	if !ids["task1"] || !ids["task2"] {
		t.Error("both tasks should be written to inbox despite delete failure")
	}
}

func TestSyncIdeas_SkipsEmptyTitles(t *testing.T) {
	dir := t.TempDir()
	inboxPath := filepath.Join(dir, "inbox.md")

	client := &mockTasksClient{
		taskLists: map[string]*taskapi.TaskList{
			"Ideas": {Id: "list1", Title: "Ideas"},
		},
		tasks: map[string][]*taskapi.Task{
			"list1": {
				{Id: "task1", Title: "", Updated: "2026-07-13T10:00:00Z"},
				{Id: "task2", Title: "   ", Updated: "2026-07-13T10:00:00Z"},
				{Id: "task3", Title: "Valid idea", Updated: "2026-07-13T10:00:00Z"},
			},
		},
	}

	count, err := SyncIdeas(client, "", "Ideas", inboxPath)
	if err != nil {
		t.Fatalf("SyncIdeas() error = %v", err)
	}
	if count != 1 {
		t.Errorf("SyncIdeas() count = %d, want 1 (empty titles skipped)", count)
	}
}

func TestSyncIdeas_EmptyInboxPath(t *testing.T) {
	client := &mockTasksClient{
		findErr: fmt.Errorf("should not be called"),
	}

	count, err := SyncIdeas(client, "", "Ideas", "")
	if err != nil {
		t.Fatalf("SyncIdeas() error = %v", err)
	}
	if count != 0 {
		t.Errorf("SyncIdeas() count = %d, want 0", count)
	}
}

func TestSyncIdeas_EmptyListName(t *testing.T) {
	client := &mockTasksClient{
		findErr: fmt.Errorf("should not be called"),
	}

	count, err := SyncIdeas(client, "", "", "/some/path.md")
	if err != nil {
		t.Fatalf("SyncIdeas() error = %v", err)
	}
	if count != 0 {
		t.Errorf("SyncIdeas() count = %d, want 0", count)
	}
}

func TestSyncIdeas_BothEmpty(t *testing.T) {
	client := &mockTasksClient{
		findErr: fmt.Errorf("should not be called"),
	}

	count, err := SyncIdeas(client, "", "", "")
	if err != nil {
		t.Fatalf("SyncIdeas() error = %v", err)
	}
	if count != 0 {
		t.Errorf("SyncIdeas() count = %d, want 0", count)
	}
}

func TestSyncIdeas_EmptyTaskList(t *testing.T) {
	dir := t.TempDir()
	inboxPath := filepath.Join(dir, "inbox.md")

	client := &mockTasksClient{
		taskLists: map[string]*taskapi.TaskList{
			"Ideas": {Id: "list1", Title: "Ideas"},
		},
		tasks: map[string][]*taskapi.Task{
			"list1": {},
		},
	}

	count, err := SyncIdeas(client, "", "Ideas", inboxPath)
	if err != nil {
		t.Fatalf("SyncIdeas() error = %v", err)
	}
	if count != 0 {
		t.Errorf("SyncIdeas() count = %d, want 0", count)
	}
}
