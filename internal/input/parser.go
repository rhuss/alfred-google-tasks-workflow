package input

import (
	"regexp"
	"strings"
	"time"

	"github.com/rhuss/alfred-google-tasks-workflow/internal/dateparse"
)

// ParsedInput holds the extracted components from user input.
type ParsedInput struct {
	Title    string
	Date     *time.Time
	ListName string
}

// listTagRegex matches #ListName tags (alphanumeric, hyphens, underscores).
var listTagRegex = regexp.MustCompile(`\s*#([A-Za-z0-9_-]+)\s*$`)

// Parse extracts a title, optional date, and optional #ListName from user input.
//
// Parsing order:
//  1. Extract #ListName tag (if present at end of input)
//  2. Check for date after last comma (comma-separated trailing date)
//  3. Check for date at the beginning of input (leading date tokens)
//  4. Remaining text becomes the title
func Parse(input string) ParsedInput {
	return ParseWithTime(input, time.Now())
}

// ParseWithTime is like Parse but uses the given time as reference for date parsing.
func ParseWithTime(input string, now time.Time) ParsedInput {
	result := ParsedInput{}
	remaining := strings.TrimSpace(input)

	if remaining == "" {
		return result
	}

	// Step 1: Extract #ListName tag from end of input
	if match := listTagRegex.FindStringSubmatchIndex(remaining); match != nil {
		result.ListName = remaining[match[2]:match[3]]
		// Replace hyphens and underscores with spaces for the actual list name
		result.ListName = strings.ReplaceAll(result.ListName, "-", " ")
		result.ListName = strings.ReplaceAll(result.ListName, "_", " ")
		remaining = strings.TrimSpace(remaining[:match[0]])
	}

	if remaining == "" {
		return result
	}

	// Step 2: Check for trailing date after last comma
	if lastComma := strings.LastIndex(remaining, ","); lastComma >= 0 {
		beforeComma := strings.TrimSpace(remaining[:lastComma])
		afterComma := strings.TrimSpace(remaining[lastComma+1:])

		if afterComma != "" {
			if date, ok := dateparse.Parse(afterComma, now); ok {
				result.Date = &date
				result.Title = beforeComma
				return result
			}
		}
	}

	// Step 3: Check for leading date token(s)
	// Try "next week" (two-word date phrase) first
	words := strings.Fields(remaining)
	if len(words) >= 3 {
		twoWordDate := strings.Join(words[:2], " ")
		if date, ok := dateparse.Parse(twoWordDate, now); ok {
			result.Date = &date
			result.Title = strings.TrimSpace(strings.Join(words[2:], " "))
			return result
		}
	}

	// Try single leading word as date
	if len(words) >= 2 {
		if date, ok := dateparse.Parse(words[0], now); ok {
			result.Date = &date
			result.Title = strings.TrimSpace(strings.Join(words[1:], " "))
			return result
		}
	}

	// Step 4: No date found, entire input is the title
	result.Title = remaining
	return result
}
