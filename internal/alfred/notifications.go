package alfred

import (
	"context"
	"os/exec"
	"time"
)

// NotifySuccess shows a macOS notification for a successful operation.
func NotifySuccess(title, message string) {
	notify(title, message, "")
}

// NotifyError shows a macOS notification for an error.
func NotifyError(title, message string) {
	notify(title, message, "Basso")
}

// notify displays a macOS notification using osascript.
// sound can be empty for no sound, or a macOS sound name like "Basso", "Glass", etc.
func notify(title, message, sound string) {
	script := `display notification "` + escapeAppleScript(message) + `" with title "` + escapeAppleScript(title) + `"`
	if sound != "" {
		script += ` sound name "` + sound + `"`
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	go func() {
		defer cancel()
		_ = exec.CommandContext(ctx, "osascript", "-e", script).Run()
	}()
}

// escapeAppleScript escapes special characters for AppleScript strings.
func escapeAppleScript(s string) string {
	result := make([]byte, 0, len(s))
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '"':
			result = append(result, '\\', '"')
		case '\\':
			result = append(result, '\\', '\\')
		case '\n':
			result = append(result, '\\', 'n')
		default:
			result = append(result, s[i])
		}
	}
	return string(result)
}
