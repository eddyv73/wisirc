// Package clipboard copies text with the system clipboard tool.
package clipboard

import (
	"errors"
	"os/exec"
	"runtime"
	"strings"
)

var ErrUnavailable = errors.New("no system clipboard")

// Command returns the program and argv that read stdin into the clipboard.
func Command(goos string, lookPath func(string) (string, error)) (string, []string, bool) {
	if goos == "darwin" {
		return "pbcopy", nil, true
	}
	for _, c := range []struct {
		name string
		args []string
	}{
		{"wl-copy", nil},
		{"xclip", []string{"-selection", "clipboard"}},
		{"xsel", []string{"--clipboard", "--input"}},
	} {
		if _, err := lookPath(c.name); err == nil {
			return c.name, c.args, true
		}
	}
	return "", nil, false
}

// Copy puts text on the system clipboard.
func Copy(text string) error {
	name, args, ok := Command(runtime.GOOS, exec.LookPath)
	if !ok {
		return ErrUnavailable
	}
	cmd := exec.Command(name, args...)
	cmd.Stdin = strings.NewReader(text)
	return cmd.Run()
}
