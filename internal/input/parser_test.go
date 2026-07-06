package input

import (
	"testing"
	"time"
)

// refTime is a fixed reference time for deterministic tests: Wednesday, 2026-07-01
var refTime = time.Date(2026, 7, 1, 12, 0, 0, 0, time.UTC)

func TestParseTitleOnly(t *testing.T) {
	result := ParseWithTime("Buy groceries", refTime)
	if result.Title != "Buy groceries" {
		t.Errorf("expected title 'Buy groceries', got %q", result.Title)
	}
	if result.Date != nil {
		t.Error("expected no date")
	}
	if result.ListName != "" {
		t.Errorf("expected no list name, got %q", result.ListName)
	}
}

func TestParseTitleWithTrailingDate(t *testing.T) {
	result := ParseWithTime("Buy groceries, tomorrow", refTime)
	if result.Title != "Buy groceries" {
		t.Errorf("expected title 'Buy groceries', got %q", result.Title)
	}
	if result.Date == nil {
		t.Fatal("expected a date")
	}
	expected := time.Date(2026, 7, 2, 0, 0, 0, 0, time.UTC)
	if !result.Date.Equal(expected) {
		t.Errorf("expected date %v, got %v", expected, *result.Date)
	}
}

func TestParseTitleWithLeadingDate(t *testing.T) {
	result := ParseWithTime("tomorrow Buy groceries", refTime)
	if result.Title != "Buy groceries" {
		t.Errorf("expected title 'Buy groceries', got %q", result.Title)
	}
	if result.Date == nil {
		t.Fatal("expected a date")
	}
	expected := time.Date(2026, 7, 2, 0, 0, 0, 0, time.UTC)
	if !result.Date.Equal(expected) {
		t.Errorf("expected date %v, got %v", expected, *result.Date)
	}
}

func TestParseTitleWithListName(t *testing.T) {
	result := ParseWithTime("Buy groceries #Shopping", refTime)
	if result.Title != "Buy groceries" {
		t.Errorf("expected title 'Buy groceries', got %q", result.Title)
	}
	if result.ListName != "Shopping" {
		t.Errorf("expected list name 'Shopping', got %q", result.ListName)
	}
	if result.Date != nil {
		t.Error("expected no date")
	}
}

func TestParseTitleDateAndList(t *testing.T) {
	result := ParseWithTime("Buy groceries, friday #Shopping", refTime)
	if result.Title != "Buy groceries" {
		t.Errorf("expected title 'Buy groceries', got %q", result.Title)
	}
	if result.ListName != "Shopping" {
		t.Errorf("expected list name 'Shopping', got %q", result.ListName)
	}
	if result.Date == nil {
		t.Fatal("expected a date")
	}
	expected := time.Date(2026, 7, 3, 0, 0, 0, 0, time.UTC)
	if !result.Date.Equal(expected) {
		t.Errorf("expected date %v, got %v", expected, *result.Date)
	}
}

func TestParseLeadingDateWithList(t *testing.T) {
	result := ParseWithTime("friday Buy groceries #Shopping", refTime)
	if result.Title != "Buy groceries" {
		t.Errorf("expected title 'Buy groceries', got %q", result.Title)
	}
	if result.ListName != "Shopping" {
		t.Errorf("expected list name 'Shopping', got %q", result.ListName)
	}
	if result.Date == nil {
		t.Fatal("expected a date")
	}
	expected := time.Date(2026, 7, 3, 0, 0, 0, 0, time.UTC)
	if !result.Date.Equal(expected) {
		t.Errorf("expected date %v, got %v", expected, *result.Date)
	}
}

func TestParseNextWeekLeading(t *testing.T) {
	result := ParseWithTime("next week Submit report", refTime)
	if result.Title != "Submit report" {
		t.Errorf("expected title 'Submit report', got %q", result.Title)
	}
	if result.Date == nil {
		t.Fatal("expected a date")
	}
	// Next Monday from Wednesday Jul 1 is Jul 6
	expected := time.Date(2026, 7, 6, 0, 0, 0, 0, time.UTC)
	if !result.Date.Equal(expected) {
		t.Errorf("expected date %v, got %v", expected, *result.Date)
	}
}

func TestParseNextWeekTrailing(t *testing.T) {
	result := ParseWithTime("Submit report, next week", refTime)
	if result.Title != "Submit report" {
		t.Errorf("expected title 'Submit report', got %q", result.Title)
	}
	if result.Date == nil {
		t.Fatal("expected a date")
	}
	expected := time.Date(2026, 7, 6, 0, 0, 0, 0, time.UTC)
	if !result.Date.Equal(expected) {
		t.Errorf("expected date %v, got %v", expected, *result.Date)
	}
}

func TestParseISODateTrailing(t *testing.T) {
	result := ParseWithTime("Submit report, 2026-07-15", refTime)
	if result.Title != "Submit report" {
		t.Errorf("expected title 'Submit report', got %q", result.Title)
	}
	if result.Date == nil {
		t.Fatal("expected a date")
	}
	expected := time.Date(2026, 7, 15, 0, 0, 0, 0, time.UTC)
	if !result.Date.Equal(expected) {
		t.Errorf("expected date %v, got %v", expected, *result.Date)
	}
}

func TestParseListOnlyReturnsNoTitle(t *testing.T) {
	result := ParseWithTime("#Work", refTime)
	if result.Title != "" {
		t.Errorf("expected empty title, got %q", result.Title)
	}
	if result.ListName != "Work" {
		t.Errorf("expected list name 'Work', got %q", result.ListName)
	}
}

func TestParseListNameWithHyphens(t *testing.T) {
	result := ParseWithTime("Fix bug #my-project", refTime)
	if result.Title != "Fix bug" {
		t.Errorf("expected title 'Fix bug', got %q", result.Title)
	}
	if result.ListName != "my project" {
		t.Errorf("expected list name 'my project', got %q", result.ListName)
	}
}

func TestParseListNameWithUnderscores(t *testing.T) {
	result := ParseWithTime("Fix bug #my_project", refTime)
	if result.Title != "Fix bug" {
		t.Errorf("expected title 'Fix bug', got %q", result.Title)
	}
	if result.ListName != "my project" {
		t.Errorf("expected list name 'my project', got %q", result.ListName)
	}
}

func TestParseEmptyInput(t *testing.T) {
	result := ParseWithTime("", refTime)
	if result.Title != "" {
		t.Errorf("expected empty title, got %q", result.Title)
	}
	if result.Date != nil {
		t.Error("expected no date")
	}
	if result.ListName != "" {
		t.Errorf("expected no list name, got %q", result.ListName)
	}
}

func TestParseWhitespaceInput(t *testing.T) {
	result := ParseWithTime("   ", refTime)
	if result.Title != "" {
		t.Errorf("expected empty title, got %q", result.Title)
	}
}

func TestParseCommaWithoutDate(t *testing.T) {
	// "Hello, world" should keep full text as title since "world" is not a date
	result := ParseWithTime("Hello, world", refTime)
	if result.Title != "Hello, world" {
		t.Errorf("expected title 'Hello, world', got %q", result.Title)
	}
	if result.Date != nil {
		t.Error("expected no date")
	}
}

func TestParseTodayLeading(t *testing.T) {
	result := ParseWithTime("today Call dentist", refTime)
	if result.Title != "Call dentist" {
		t.Errorf("expected title 'Call dentist', got %q", result.Title)
	}
	if result.Date == nil {
		t.Fatal("expected a date")
	}
	expected := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	if !result.Date.Equal(expected) {
		t.Errorf("expected date %v, got %v", expected, *result.Date)
	}
}

func TestParsePublicFunctionUsesNow(t *testing.T) {
	// Just verify Parse() returns a result (uses time.Now internally)
	result := Parse("Buy groceries")
	if result.Title != "Buy groceries" {
		t.Errorf("expected title 'Buy groceries', got %q", result.Title)
	}
}
