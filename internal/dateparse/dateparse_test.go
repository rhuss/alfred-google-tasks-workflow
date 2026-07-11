package dateparse

import (
	"testing"
	"time"
)

// refTime is a fixed reference time for deterministic tests: Wednesday, 2026-07-01
var refTime = time.Date(2026, 7, 1, 12, 0, 0, 0, time.UTC)

func TestParseToday(t *testing.T) {
	date, ok := Parse("today", refTime)
	if !ok {
		t.Fatal("expected 'today' to parse")
	}
	expected := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	if !date.Equal(expected) {
		t.Errorf("expected %v, got %v", expected, date)
	}
}

func TestParseTodayCaseInsensitive(t *testing.T) {
	date, ok := Parse("TODAY", refTime)
	if !ok {
		t.Fatal("expected 'TODAY' to parse")
	}
	expected := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	if !date.Equal(expected) {
		t.Errorf("expected %v, got %v", expected, date)
	}
}

func TestParseTomorrow(t *testing.T) {
	date, ok := Parse("tomorrow", refTime)
	if !ok {
		t.Fatal("expected 'tomorrow' to parse")
	}
	expected := time.Date(2026, 7, 2, 0, 0, 0, 0, time.UTC)
	if !date.Equal(expected) {
		t.Errorf("expected %v, got %v", expected, date)
	}
}

func TestParseNextWeek(t *testing.T) {
	// refTime is Wednesday Jul 1; next Monday is Jul 6
	date, ok := Parse("next week", refTime)
	if !ok {
		t.Fatal("expected 'next week' to parse")
	}
	expected := time.Date(2026, 7, 6, 0, 0, 0, 0, time.UTC)
	if !date.Equal(expected) {
		t.Errorf("expected %v, got %v", expected, date)
	}
}

func TestParseNextWeekOnSunday(t *testing.T) {
	// Sunday Jul 5; next Monday is Jul 6
	sunday := time.Date(2026, 7, 5, 12, 0, 0, 0, time.UTC)
	date, ok := Parse("next week", sunday)
	if !ok {
		t.Fatal("expected 'next week' to parse")
	}
	expected := time.Date(2026, 7, 6, 0, 0, 0, 0, time.UTC)
	if !date.Equal(expected) {
		t.Errorf("expected %v, got %v", expected, date)
	}
}

func TestParseNextWeekOnMonday(t *testing.T) {
	// Monday Jul 6; next Monday is Jul 13 (7 days, not same day)
	monday := time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC)
	date, ok := Parse("next week", monday)
	if !ok {
		t.Fatal("expected 'next week' to parse")
	}
	expected := time.Date(2026, 7, 13, 0, 0, 0, 0, time.UTC)
	if !date.Equal(expected) {
		t.Errorf("expected %v, got %v", expected, date)
	}
}

func TestParseGermanKeywords(t *testing.T) {
	tests := []struct {
		input    string
		expected time.Time
	}{
		{"heute", time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)},
		{"Heute", time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)},
		{"morgen", time.Date(2026, 7, 2, 0, 0, 0, 0, time.UTC)},
		{"Morgen", time.Date(2026, 7, 2, 0, 0, 0, 0, time.UTC)},
		{"uebermorgen", time.Date(2026, 7, 3, 0, 0, 0, 0, time.UTC)},
		{"Uebermorgen", time.Date(2026, 7, 3, 0, 0, 0, 0, time.UTC)},
		{"übermorgen", time.Date(2026, 7, 3, 0, 0, 0, 0, time.UTC)},
		{"naechste woche", time.Date(2026, 7, 6, 0, 0, 0, 0, time.UTC)},
		{"nächste woche", time.Date(2026, 7, 6, 0, 0, 0, 0, time.UTC)},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			date, ok := Parse(tt.input, refTime)
			if !ok {
				t.Fatalf("expected %q to parse", tt.input)
			}
			if !date.Equal(tt.expected) {
				t.Errorf("expected %v, got %v", tt.expected, date)
			}
		})
	}
}

func TestParseWeekdayNames(t *testing.T) {
	tests := []struct {
		input    string
		expected time.Time
	}{
		// refTime is Wednesday Jul 1
		{"monday", time.Date(2026, 7, 6, 0, 0, 0, 0, time.UTC)},    // next Monday
		{"tuesday", time.Date(2026, 7, 7, 0, 0, 0, 0, time.UTC)},   // next Tuesday
		{"wednesday", time.Date(2026, 7, 8, 0, 0, 0, 0, time.UTC)},  // next Wednesday (not today)
		{"thursday", time.Date(2026, 7, 2, 0, 0, 0, 0, time.UTC)},   // tomorrow (Thursday)
		{"friday", time.Date(2026, 7, 3, 0, 0, 0, 0, time.UTC)},     // this Friday
		{"saturday", time.Date(2026, 7, 4, 0, 0, 0, 0, time.UTC)},   // this Saturday
		{"sunday", time.Date(2026, 7, 5, 0, 0, 0, 0, time.UTC)},     // this Sunday
		{"Friday", time.Date(2026, 7, 3, 0, 0, 0, 0, time.UTC)},     // case insensitive
		{"MONDAY", time.Date(2026, 7, 6, 0, 0, 0, 0, time.UTC)},     // case insensitive
		{"Montag", time.Date(2026, 7, 6, 0, 0, 0, 0, time.UTC)},     // German Monday
		{"dienstag", time.Date(2026, 7, 7, 0, 0, 0, 0, time.UTC)},   // German Tuesday
		{"Mittwoch", time.Date(2026, 7, 8, 0, 0, 0, 0, time.UTC)},   // German Wednesday
		{"donnerstag", time.Date(2026, 7, 2, 0, 0, 0, 0, time.UTC)}, // German Thursday
		{"Freitag", time.Date(2026, 7, 3, 0, 0, 0, 0, time.UTC)},    // German Friday
		{"samstag", time.Date(2026, 7, 4, 0, 0, 0, 0, time.UTC)},    // German Saturday
		{"Sonntag", time.Date(2026, 7, 5, 0, 0, 0, 0, time.UTC)},    // German Sunday
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			date, ok := Parse(tt.input, refTime)
			if !ok {
				t.Fatalf("expected %q to parse", tt.input)
			}
			if !date.Equal(tt.expected) {
				t.Errorf("expected %v, got %v", tt.expected, date)
			}
		})
	}
}

func TestParseISODate(t *testing.T) {
	date, ok := Parse("2026-12-25", refTime)
	if !ok {
		t.Fatal("expected ISO date to parse")
	}
	expected := time.Date(2026, 12, 25, 0, 0, 0, 0, time.UTC)
	if !date.Equal(expected) {
		t.Errorf("expected %v, got %v", expected, date)
	}
}

func TestParseShortDateFuture(t *testing.T) {
	// 08-15 is in the future relative to Jul 1
	date, ok := Parse("08-15", refTime)
	if !ok {
		t.Fatal("expected short date to parse")
	}
	expected := time.Date(2026, 8, 15, 0, 0, 0, 0, time.UTC)
	if !date.Equal(expected) {
		t.Errorf("expected %v, got %v", expected, date)
	}
}

func TestParseShortDatePastRollsToNextYear(t *testing.T) {
	// 01-15 is in the past relative to Jul 1, should roll to 2027
	date, ok := Parse("01-15", refTime)
	if !ok {
		t.Fatal("expected short date to parse")
	}
	expected := time.Date(2027, 1, 15, 0, 0, 0, 0, time.UTC)
	if !date.Equal(expected) {
		t.Errorf("expected %v, got %v", expected, date)
	}
}

func TestParseEmptyString(t *testing.T) {
	_, ok := Parse("", refTime)
	if ok {
		t.Error("expected empty string to not parse")
	}
}

func TestParseWhitespace(t *testing.T) {
	date, ok := Parse("  today  ", refTime)
	if !ok {
		t.Fatal("expected whitespace-padded 'today' to parse")
	}
	expected := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	if !date.Equal(expected) {
		t.Errorf("expected %v, got %v", expected, date)
	}
}

func TestParseInvalidInput(t *testing.T) {
	invalids := []string{
		"not a date",
		"yesterday",
		"2026-13-01",
		"2026",
		"hello world",
		"next month",
	}
	for _, input := range invalids {
		t.Run(input, func(t *testing.T) {
			_, ok := Parse(input, refTime)
			if ok {
				t.Errorf("expected %q to not parse", input)
			}
		})
	}
}

func TestParseStripsTimeComponent(t *testing.T) {
	// Even with a reference time that has hours, the result should be midnight
	ref := time.Date(2026, 7, 1, 15, 30, 45, 0, time.UTC)
	date, ok := Parse("today", ref)
	if !ok {
		t.Fatal("expected 'today' to parse")
	}
	if date.Hour() != 0 || date.Minute() != 0 || date.Second() != 0 {
		t.Errorf("expected midnight, got %v", date)
	}
}
