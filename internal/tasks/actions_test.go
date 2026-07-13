package tasks

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"

	taskapi "google.golang.org/api/tasks/v1"
)

func TestPartialMoveError(t *testing.T) {
	origErr := fmt.Errorf("delete failed: connection refused")
	task := &taskapi.Task{Id: "new-task-1", Title: "Test Task"}

	pme := &PartialMoveError{
		CreatedTask: task,
		DeleteErr:   origErr,
	}

	// Test Error() message
	if !strings.Contains(pme.Error(), "delete failed") {
		t.Errorf("Error() = %q, expected it to contain 'delete failed'", pme.Error())
	}

	// Test Unwrap() returns the original delete error
	if pme.Unwrap() != origErr {
		t.Errorf("Unwrap() returned different error")
	}

	// Test errors.As works correctly
	var target *PartialMoveError
	if !errors.As(pme, &target) {
		t.Error("errors.As failed to match PartialMoveError")
	}
	if target.CreatedTask.Id != "new-task-1" {
		t.Errorf("errors.As target CreatedTask.Id = %q, want %q", target.CreatedTask.Id, "new-task-1")
	}

	// Test that a regular error does NOT match PartialMoveError
	regularErr := fmt.Errorf("some other error")
	if errors.As(regularErr, &target) {
		t.Error("errors.As should not match regular error as PartialMoveError")
	}
}

func TestMoveTask_Success(t *testing.T) {
	sourceTask := &taskapi.Task{
		Id:    "task1",
		Title: "Buy groceries",
		Due:   "2026-07-20T00:00:00Z",
		Notes: "Get milk and eggs",
	}

	// Track what the target receives
	var insertedTask *taskapi.Task

	// Source mock: serves GetTask and DeleteTask
	sourceHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodGet && strings.Contains(r.URL.Path, "/tasks/task1"):
			// GetTask
			json.NewEncoder(w).Encode(sourceTask)
		case r.Method == http.MethodDelete && strings.Contains(r.URL.Path, "/tasks/task1"):
			// DeleteTask
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Errorf("source: unexpected request %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	})

	// Target mock: serves ListTaskLists (for FindTaskListByName) and InsertTask
	targetHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodGet && strings.HasSuffix(r.URL.Path, "/users/@me/lists"):
			// ListTaskLists (called by FindTaskListByName via ResolveTaskList)
			json.NewEncoder(w).Encode(&taskapi.TaskLists{
				Items: []*taskapi.TaskList{
					{Id: "target-list-1", Title: "Shopping"},
				},
			})
		case r.Method == http.MethodPost && strings.Contains(r.URL.Path, "/lists/target-list-1/tasks"):
			// InsertTask
			var task taskapi.Task
			json.NewDecoder(r.Body).Decode(&task)
			insertedTask = &task
			task.Id = "new-task-on-target"
			json.NewEncoder(w).Encode(&task)
		default:
			t.Errorf("target: unexpected request %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	})

	sourceClient, sourceServer := newTestClient(t, sourceHandler)
	defer sourceServer.Close()

	targetClient, targetServer := newTestClient(t, targetHandler)
	defer targetServer.Close()

	created, err := sourceClient.MoveTask("list1", "task1", targetClient, "Shopping")
	if err != nil {
		t.Fatalf("MoveTask returned error: %v", err)
	}

	if created.Id != "new-task-on-target" {
		t.Errorf("created task ID = %q, want %q", created.Id, "new-task-on-target")
	}

	// Verify the inserted task had the right fields
	if insertedTask == nil {
		t.Fatal("no task was inserted on target")
	}
	if insertedTask.Title != "Buy groceries" {
		t.Errorf("inserted Title = %q, want %q", insertedTask.Title, "Buy groceries")
	}
	if insertedTask.Due != "2026-07-20T00:00:00Z" {
		t.Errorf("inserted Due = %q, want %q", insertedTask.Due, "2026-07-20T00:00:00Z")
	}
	if insertedTask.Notes != "Get milk and eggs" {
		t.Errorf("inserted Notes = %q, want %q", insertedTask.Notes, "Get milk and eggs")
	}
}

func TestMoveTask_PartialFailure(t *testing.T) {
	sourceTask := &taskapi.Task{
		Id:    "task1",
		Title: "Test task",
	}

	// Source mock: GetTask succeeds, DeleteTask fails
	sourceHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodGet && strings.Contains(r.URL.Path, "/tasks/task1"):
			json.NewEncoder(w).Encode(sourceTask)
		case r.Method == http.MethodDelete && strings.Contains(r.URL.Path, "/tasks/task1"):
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]any{
				"error": map[string]any{
					"code":    500,
					"message": "internal error",
				},
			})
		default:
			t.Errorf("source: unexpected request %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	})

	// Target mock: succeeds
	targetHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodGet && strings.HasSuffix(r.URL.Path, "/users/@me/lists"):
			json.NewEncoder(w).Encode(&taskapi.TaskLists{
				Items: []*taskapi.TaskList{
					{Id: "target-list-1", Title: "Work"},
				},
			})
		case r.Method == http.MethodPost && strings.Contains(r.URL.Path, "/lists/target-list-1/tasks"):
			var task taskapi.Task
			json.NewDecoder(r.Body).Decode(&task)
			task.Id = "created-on-target"
			json.NewEncoder(w).Encode(&task)
		default:
			t.Errorf("target: unexpected request %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	})

	sourceClient, sourceServer := newTestClient(t, sourceHandler)
	defer sourceServer.Close()

	targetClient, targetServer := newTestClient(t, targetHandler)
	defer targetServer.Close()

	created, err := sourceClient.MoveTask("list1", "task1", targetClient, "Work")
	if err == nil {
		t.Fatal("expected error for partial failure, got nil")
	}

	// The created task should still be returned
	if created == nil {
		t.Fatal("expected created task to be returned on partial failure")
	}
	if created.Id != "created-on-target" {
		t.Errorf("created task ID = %q, want %q", created.Id, "created-on-target")
	}

	// Should be a PartialMoveError
	var partialErr *PartialMoveError
	if !errors.As(err, &partialErr) {
		t.Fatalf("expected PartialMoveError, got %T: %v", err, err)
	}
	if partialErr.CreatedTask.Id != "created-on-target" {
		t.Errorf("partial error CreatedTask.Id = %q, want %q", partialErr.CreatedTask.Id, "created-on-target")
	}
}

func TestMoveTask_GetTaskFails(t *testing.T) {
	// Source mock: GetTask fails
	sourceHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]any{
				"code":    404,
				"message": "not found",
			},
		})
	})

	targetHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("target should not be called when GetTask fails")
	})

	sourceClient, sourceServer := newTestClient(t, sourceHandler)
	defer sourceServer.Close()

	targetClient, targetServer := newTestClient(t, targetHandler)
	defer targetServer.Close()

	created, err := sourceClient.MoveTask("list1", "task1", targetClient, "Work")
	if err == nil {
		t.Fatal("expected error when GetTask fails, got nil")
	}
	if created != nil {
		t.Error("expected nil created task when GetTask fails")
	}

	// Should NOT be a PartialMoveError
	if _, ok := errors.AsType[*PartialMoveError](err); ok {
		t.Error("should not be PartialMoveError when GetTask fails")
	}
}

func TestMoveTask_InsertFails(t *testing.T) {
	sourceTask := &taskapi.Task{
		Id:    "task1",
		Title: "Test task",
	}

	// Source mock: GetTask succeeds
	sourceHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodGet && strings.Contains(r.URL.Path, "/tasks/task1") {
			json.NewEncoder(w).Encode(sourceTask)
			return
		}
		// DeleteTask should NOT be called
		if r.Method == http.MethodDelete {
			t.Error("DeleteTask should not be called when InsertTask fails")
		}
		w.WriteHeader(http.StatusNotFound)
	})

	// Target mock: ListTaskLists succeeds, InsertTask fails
	targetHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodGet && strings.HasSuffix(r.URL.Path, "/users/@me/lists"):
			json.NewEncoder(w).Encode(&taskapi.TaskLists{
				Items: []*taskapi.TaskList{
					{Id: "target-list-1", Title: "Work"},
				},
			})
		case r.Method == http.MethodPost:
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(map[string]any{
				"error": map[string]any{
					"code":    403,
					"message": "forbidden",
				},
			})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	})

	sourceClient, sourceServer := newTestClient(t, sourceHandler)
	defer sourceServer.Close()

	targetClient, targetServer := newTestClient(t, targetHandler)
	defer targetServer.Close()

	created, err := sourceClient.MoveTask("list1", "task1", targetClient, "Work")
	if err == nil {
		t.Fatal("expected error when InsertTask fails, got nil")
	}
	if created != nil {
		t.Error("expected nil created task when InsertTask fails")
	}

	// Should NOT be a PartialMoveError
	if _, ok := errors.AsType[*PartialMoveError](err); ok {
		t.Error("should not be PartialMoveError when InsertTask fails")
	}
}

func TestMoveTask_NoDueNoNotes(t *testing.T) {
	// Task with only a title (no due date, no notes)
	sourceTask := &taskapi.Task{
		Id:    "task1",
		Title: "Simple task",
	}

	var insertedTask *taskapi.Task

	sourceHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodGet && strings.Contains(r.URL.Path, "/tasks/task1"):
			json.NewEncoder(w).Encode(sourceTask)
		case r.Method == http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	})

	targetHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodGet && strings.HasSuffix(r.URL.Path, "/users/@me/lists"):
			json.NewEncoder(w).Encode(&taskapi.TaskLists{
				Items: []*taskapi.TaskList{
					{Id: "target-list-1", Title: "Inbox"},
				},
			})
		case r.Method == http.MethodPost:
			var task taskapi.Task
			json.NewDecoder(r.Body).Decode(&task)
			insertedTask = &task
			task.Id = "new-task"
			json.NewEncoder(w).Encode(&task)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	})

	sourceClient, sourceServer := newTestClient(t, sourceHandler)
	defer sourceServer.Close()

	targetClient, targetServer := newTestClient(t, targetHandler)
	defer targetServer.Close()

	_, err := sourceClient.MoveTask("list1", "task1", targetClient, "Inbox")
	if err != nil {
		t.Fatalf("MoveTask returned error: %v", err)
	}

	if insertedTask == nil {
		t.Fatal("no task was inserted on target")
	}
	if insertedTask.Title != "Simple task" {
		t.Errorf("inserted Title = %q, want %q", insertedTask.Title, "Simple task")
	}
	if insertedTask.Due != "" {
		t.Errorf("inserted Due = %q, want empty string", insertedTask.Due)
	}
	if insertedTask.Notes != "" {
		t.Errorf("inserted Notes = %q, want empty string", insertedTask.Notes)
	}
}

func TestMoveTask_AutoCreatesList(t *testing.T) {
	sourceTask := &taskapi.Task{
		Id:    "task1",
		Title: "Test task",
	}

	var createdListTitle string

	sourceHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodGet && strings.Contains(r.URL.Path, "/tasks/task1"):
			json.NewEncoder(w).Encode(sourceTask)
		case r.Method == http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	})

	// Target mock: no existing lists, so ResolveTaskList will auto-create
	targetHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodGet && strings.HasSuffix(r.URL.Path, "/users/@me/lists"):
			// Return empty list so FindTaskListByName returns nil
			json.NewEncoder(w).Encode(&taskapi.TaskLists{
				Items: []*taskapi.TaskList{},
			})
		case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/users/@me/lists"):
			// CreateTaskList
			var list taskapi.TaskList
			json.NewDecoder(r.Body).Decode(&list)
			createdListTitle = list.Title
			list.Id = "new-list-id"
			json.NewEncoder(w).Encode(&list)
		case r.Method == http.MethodPost && strings.Contains(r.URL.Path, "/lists/new-list-id/tasks"):
			// InsertTask
			var task taskapi.Task
			json.NewDecoder(r.Body).Decode(&task)
			task.Id = "new-task"
			json.NewEncoder(w).Encode(&task)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	})

	sourceClient, sourceServer := newTestClient(t, sourceHandler)
	defer sourceServer.Close()

	targetClient, targetServer := newTestClient(t, targetHandler)
	defer targetServer.Close()

	_, err := sourceClient.MoveTask("list1", "task1", targetClient, "New List")
	if err != nil {
		t.Fatalf("MoveTask returned error: %v", err)
	}

	if createdListTitle != "New List" {
		t.Errorf("auto-created list title = %q, want %q", createdListTitle, "New List")
	}
}
