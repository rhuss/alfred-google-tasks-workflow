package dateparse

import (
	"strings"
	"time"
)

// weekdays maps lowercase weekday names (English and German) to time.Weekday values.
var weekdays = map[string]time.Weekday{
	"monday":     time.Monday,
	"tuesday":    time.Tuesday,
	"wednesday":  time.Wednesday,
	"thursday":   time.Thursday,
	"friday":     time.Friday,
	"saturday":   time.Saturday,
	"sunday":     time.Sunday,
	"montag":     time.Monday,
	"dienstag":   time.Tuesday,
	"mittwoch":   time.Wednesday,
	"donnerstag": time.Thursday,
	"freitag":    time.Friday,
	"samstag":    time.Saturday,
	"sonntag":    time.Sunday,
}

// Parse attempts to parse a date string relative to the given reference time.
// It returns the parsed date and true if successful, or zero time and false if
// the input is not a recognized date pattern.
//
// Supported patterns (English and German):
//   - "today" / "heute"                        -> reference date
//   - "tomorrow" / "morgen"                    -> reference date + 1 day
//   - "uebermorgen" / "übermorgen"             -> reference date + 2 days
//   - "next week" / "naechste woche"           -> next Monday from reference date
//   - weekday names (EN: monday.. / DE: montag..) -> next occurrence (case-insensitive)
//   - "YYYY-MM-DD"                             -> exact ISO date
//   - "MM-DD"                                  -> month-day of current year, or next year if past
func Parse(input string, relativeTo time.Time) (time.Time, bool) {
	normalized := strings.TrimSpace(strings.ToLower(input))

	if normalized == "" {
		return time.Time{}, false
	}

	// Strip time component from relativeTo, work with date only
	today := time.Date(relativeTo.Year(), relativeTo.Month(), relativeTo.Day(),
		0, 0, 0, 0, relativeTo.Location())

	// Check keyword dates (English and German)
	switch normalized {
	case "today", "heute":
		return today, true
	case "tomorrow", "morgen":
		return today.AddDate(0, 0, 1), true
	case "next week", "naechste woche", "nächste woche":
		return nextWeekday(today, time.Monday), true
	case "uebermorgen", "übermorgen":
		return today.AddDate(0, 0, 2), true
	}

	// Check weekday names
	if wd, ok := weekdays[normalized]; ok {
		return nextWeekday(today, wd), true
	}

	// Check ISO date (YYYY-MM-DD)
	if t, err := time.Parse("2006-01-02", normalized); err == nil {
		return t, true
	}

	// Check short date (MM-DD)
	if len(normalized) == 5 && normalized[2] == '-' {
		if t, err := time.Parse("01-02", normalized); err == nil {
			// Set to current year
			result := time.Date(today.Year(), t.Month(), t.Day(),
				0, 0, 0, 0, today.Location())
			// If the date has already passed this year, use next year
			if result.Before(today) {
				result = result.AddDate(1, 0, 0)
			}
			return result, true
		}
	}

	return time.Time{}, false
}

// nextWeekday returns the date of the next occurrence of the given weekday
// after (not including) the reference date.
func nextWeekday(from time.Time, target time.Weekday) time.Time {
	current := from.Weekday()
	daysUntil := int(target) - int(current)
	if daysUntil <= 0 {
		daysUntil += 7
	}
	return from.AddDate(0, 0, daysUntil)
}
