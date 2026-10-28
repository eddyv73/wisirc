// Package notify shows desktop notifications on macOS and Linux.
package notify

import (
	"errors"
	"os/exec"
	"runtime"
	"strings"
)

var ErrUnavailable = errors.New("no notifier available")

const maxBody = 200

var appleEscaper = strings.NewReplacer(`\`, `\\`, `"`, `\"`)

// Command returns the program and argv that show a notification on goos.
func Command(goos, title, body string, lookPath func(string) (string, error)) (string, []string, bool) {
	if r := []rune(body); len(r) > maxBody {
		body = string(r[:maxBody])
	}
	switch goos {
	case "darwin":
		script := `display notification "` + appleEscaper.Replace(body) + `" with title "` + appleEscaper.Replace(title) + `"`
		return "osascript", []string{"-e", script}, true
	case "linux", "freebsd", "openbsd", "netbsd":
		if _, err := lookPath("notify-send"); err == nil {
			return "notify-send", []string{title, body}, true
		}
	}
	return "", nil, false
}

// Send shows a notification without blocking; ErrUnavailable means fall back to BEL.
func Send(title, body string) error {
	name, args, ok := Command(runtime.GOOS, title, body, exec.LookPath)
	if !ok {
		return ErrUnavailable
	}
	cmd := exec.Command(name, args...)
	if err := cmd.Start(); err != nil {
		return err
	}
	go cmd.Wait()
	return nil
}
