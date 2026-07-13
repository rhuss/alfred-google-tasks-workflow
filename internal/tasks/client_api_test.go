package tasks

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"google.golang.org/api/option"
	taskapi "google.golang.org/api/tasks/v1"
)

// newTestClient creates a Client backed by a mock HTTP server.
// The handler receives real Google Tasks API requests.
func newTestClient(t *testing.T, handler http.Handler) (*Client, *httptest.Server) {
	t.Helper()
	server := httptest.NewServer(handler)

	service, err := taskapi.NewService(context.Background(),
		option.WithHTTPClient(server.Client()),
		option.WithEndpoint(server.URL),
	)
	if err != nil {
		t.Fatalf("creating test service: %v", err)
	}

	return &Client{service: service}, server
}

func TestGetTask(t *testing.T) {
	tests := []struct {
		name       string
		listID     string
		taskID     string
		response   *taskapi.Task
		statusCode int
		wantErr    bool
	}{
		{
			name:   "success",
			listID: "list1",
			taskID: "task1",
			response: &taskapi.Task{
				Id:    "task1",
				Title: "Buy groceries",
				Due:   "2026-07-20T00:00:00Z",
				Notes: "Get milk and eggs",
			},
			statusCode: http.StatusOK,
		},
		{
			name:       "not found",
			listID:     "list1",
			taskID:     "nonexistent",
			statusCode: http.StatusNotFound,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				expectedPath := fmt.Sprintf("/tasks/v1/lists/%s/tasks/%s", tt.listID, tt.taskID)
				if r.URL.Path != expectedPath {
					t.Errorf("unexpected path: got %s, want %s", r.URL.Path, expectedPath)
				}
				if r.Method != http.MethodGet {
					t.Errorf("unexpected method: got %s, want GET", r.Method)
				}

				w.Header().Set("Content-Type", "application/json")
				if tt.statusCode != http.StatusOK {
					w.WriteHeader(tt.statusCode)
					json.NewEncoder(w).Encode(map[string]any{
						"error": map[string]any{
							"code":    tt.statusCode,
							"message": "not found",
						},
					})
					return
				}
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(tt.response)
			})

			client, server := newTestClient(t, handler)
			defer server.Close()

			task, err := client.GetTask(tt.listID, tt.taskID)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if task.Id != tt.response.Id {
				t.Errorf("task ID = %q, want %q", task.Id, tt.response.Id)
			}
			if task.Title != tt.response.Title {
				t.Errorf("task Title = %q, want %q", task.Title, tt.response.Title)
			}
			if task.Due != tt.response.Due {
				t.Errorf("task Due = %q, want %q", task.Due, tt.response.Due)
			}
			if task.Notes != tt.response.Notes {
				t.Errorf("task Notes = %q, want %q", task.Notes, tt.response.Notes)
			}
		})
	}
}

func TestGetTaskList(t *testing.T) {
	tests := []struct {
		name       string
		listID     string
		response   *taskapi.TaskList
		statusCode int
		wantErr    bool
	}{
		{
			name:   "success",
			listID: "list1",
			response: &taskapi.TaskList{
				Id:    "list1",
				Title: "Shopping",
			},
			statusCode: http.StatusOK,
		},
		{
			name:       "not found",
			listID:     "nonexistent",
			statusCode: http.StatusNotFound,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				expectedPath := fmt.Sprintf("/tasks/v1/users/@me/lists/%s", tt.listID)
				if r.URL.Path != expectedPath {
					t.Errorf("unexpected path: got %s, want %s", r.URL.Path, expectedPath)
				}

				w.Header().Set("Content-Type", "application/json")
				if tt.statusCode != http.StatusOK {
					w.WriteHeader(tt.statusCode)
					json.NewEncoder(w).Encode(map[string]any{
						"error": map[string]any{
							"code":    tt.statusCode,
							"message": "not found",
						},
					})
					return
				}
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(tt.response)
			})

			client, server := newTestClient(t, handler)
			defer server.Close()

			list, err := client.GetTaskList(tt.listID)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if list.Id != tt.response.Id {
				t.Errorf("list ID = %q, want %q", list.Id, tt.response.Id)
			}
			if list.Title != tt.response.Title {
				t.Errorf("list Title = %q, want %q", list.Title, tt.response.Title)
			}
		})
	}
}
