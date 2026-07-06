package auth

import "os/exec"

// execCommand runs an external command. Extracted for testability.
func execCommand(name string, args ...string) error {
	return exec.Command(name, args...).Start()
}
