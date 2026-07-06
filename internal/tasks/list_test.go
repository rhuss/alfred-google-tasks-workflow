package tasks

import (
	"testing"
	"time"

	taskapi "google.golang.org/api/tasks/v1"
)

// refNow is a fixed reference time: Wednesday, 2026-07-01 at noon UTC
var refNow = time.Date(2026, 7, 1, 12, 0, 0, 0, time.UTC)

func TestClassifyTimeframe(t *testing.T) {
	tests := []struct {
		name     string
		dueStr   string
		expected Timeframe
	}{
		{
			name:     "no date",
			dueStr:   "",
			expected: TimeframeNoDate,
		},
		{
			name:     "overdue yesterday",
			dueStr:   "2026-06-30T00:00:00Z",
			expected: TimeframeOverdue,
		},
		{
			name:     "overdue last week",
			dueStr:   "2026-06-24T00:00:00Z",
			expected: TimeframeOverdue,
		},
		{
			name:     "today",
			dueStr:   "2026-07-01T00:00:00Z",
			expected: TimeframeToday,
		},
		{
			name:     "tomorrow (this week)",
			dueStr:   "2026-07-02T00:00:00Z",
			expected: TimeframeThisWeek,
		},
		{
			name:     "6 days from now (this week)",
			dueStr:   "2026-07-07T00:00:00Z",
			expected: TimeframeThisWeek,
		},
		{
			name:     "7 days from now (later)",
			dueStr:   "2026-07-08T00:00:00Z",
			expected: TimeframeLater,
		},
		{
			name:     "next month (later)",
			dueStr:   "2026-08-15T00:00:00Z",
			expected: TimeframeLater,
		},
		{
			name:     "invalid date string",
			dueStr:   "not-a-date",
			expected: TimeframeNoDate,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ClassifyTimeframe(tt.dueStr, refNow)
			if result != tt.expected {
				t.Errorf("ClassifyTimeframe(%q) = %d (%s), want %d (%s)",
					tt.dueStr, result, TimeframeLabel(result), tt.expected, TimeframeLabel(tt.expected))
			}
		})
	}
}

func TestGroupTasksByTimeframe(t *testing.T) {
	items := []TaskItem{
		{Task: &taskapi.Task{Title: "Overdue task", Due: "2026-06-29T00:00:00Z"}, ListName: "Work", ListID: "list1"},
		{Task: &taskapi.Task{Title: "Today task", Due: "2026-07-01T00:00:00Z"}, ListName: "Work", ListID: "list1"},
		{Task: &taskapi.Task{Title: "This week task", Due: "2026-07-04T00:00:00Z"}, ListName: "Personal", ListID: "list2"},
		{Task: &taskapi.Task{Title: "Later task", Due: "2026-08-01T00:00:00Z"}, ListName: "Work", ListID: "list1"},
		{Task: &taskapi.Task{Title: "No date task", Due: ""}, ListName: "Personal", ListID: "list2"},
	}

	grouped := GroupTasksByTimeframe(items, refNow)

	if len(grouped.Groups) != 5 {
		t.Fatalf("expected 5 groups, got %d", len(grouped.Groups))
	}

	// Verify group ordering
	expectedOrder := []Timeframe{TimeframeOverdue, TimeframeToday, TimeframeThisWeek, TimeframeLater, TimeframeNoDate}
	for i, group := range grouped.Groups {
		if group.Timeframe != expectedOrder[i] {
			t.Errorf("group %d: expected timeframe %s, got %s",
				i, TimeframeLabel(expectedOrder[i]), TimeframeLabel(group.Timeframe))
		}
	}

	// Verify each group has the correct number of tasks
	for _, group := range grouped.Groups {
		if len(group.Tasks) != 1 {
			t.Errorf("group %s: expected 1 task, got %d", group.Label, len(group.Tasks))
		}
	}
}

func TestGroupTasksByTimeframeEmpty(t *testing.T) {
	grouped := GroupTasksByTimeframe(nil, refNow)
	if len(grouped.Groups) != 0 {
		t.Errorf("expected 0 groups for nil input, got %d", len(grouped.Groups))
	}
}

func TestGroupTasksByTimeframeSorting(t *testing.T) {
	items := []TaskItem{
		{Task: &taskapi.Task{Title: "Later task", Due: "2026-07-05T00:00:00Z"}, ListName: "Work", ListID: "list1"},
		{Task: &taskapi.Task{Title: "Earlier task", Due: "2026-07-03T00:00:00Z"}, ListName: "Work", ListID: "list1"},
		{Task: &taskapi.Task{Title: "Middle task", Due: "2026-07-04T00:00:00Z"}, ListName: "Work", ListID: "list1"},
	}

	grouped := GroupTasksByTimeframe(items, refNow)

	if len(grouped.Groups) != 1 {
		t.Fatalf("expected 1 group (This Week), got %d", len(grouped.Groups))
	}

	group := grouped.Groups[0]
	if len(group.Tasks) != 3 {
		t.Fatalf("expected 3 tasks, got %d", len(group.Tasks))
	}

	// Verify sorted by due date
	if group.Tasks[0].Task.Title != "Earlier task" {
		t.Errorf("expected first task 'Earlier task', got %q", group.Tasks[0].Task.Title)
	}
	if group.Tasks[1].Task.Title != "Middle task" {
		t.Errorf("expected second task 'Middle task', got %q", group.Tasks[1].Task.Title)
	}
	if group.Tasks[2].Task.Title != "Later task" {
		t.Errorf("expected third task 'Later task', got %q", group.Tasks[2].Task.Title)
	}
}

func TestGroupTasksByTimeframeNoDatesAlphabetical(t *testing.T) {
	items := []TaskItem{
		{Task: &taskapi.Task{Title: "Charlie", Due: ""}, ListName: "Work", ListID: "list1"},
		{Task: &taskapi.Task{Title: "Alpha", Due: ""}, ListName: "Work", ListID: "list1"},
		{Task: &taskapi.Task{Title: "Bravo", Due: ""}, ListName: "Work", ListID: "list1"},
	}

	grouped := GroupTasksByTimeframe(items, refNow)

	if len(grouped.Groups) != 1 {
		t.Fatalf("expected 1 group (No Date), got %d", len(grouped.Groups))
	}

	group := grouped.Groups[0]
	if group.Tasks[0].Task.Title != "Alpha" {
		t.Errorf("expected first task 'Alpha', got %q", group.Tasks[0].Task.Title)
	}
	if group.Tasks[1].Task.Title != "Bravo" {
		t.Errorf("expected second task 'Bravo', got %q", group.Tasks[1].Task.Title)
	}
	if group.Tasks[2].Task.Title != "Charlie" {
		t.Errorf("expected third task 'Charlie', got %q", group.Tasks[2].Task.Title)
	}
}

func TestTimeframeLabel(t *testing.T) {
	tests := []struct {
		tf       Timeframe
		expected string
	}{
		{TimeframeOverdue, "Overdue"},
		{TimeframeToday, "Today"},
		{TimeframeThisWeek, "This Week"},
		{TimeframeLater, "Later"},
		{TimeframeNoDate, "No Date"},
		{Timeframe(99), "Unknown"},
	}

	for _, tt := range tests {
		result := TimeframeLabel(tt.tf)
		if result != tt.expected {
			t.Errorf("TimeframeLabel(%d) = %q, want %q", tt.tf, result, tt.expected)
		}
	}
}
