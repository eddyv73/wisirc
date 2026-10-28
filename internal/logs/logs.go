// Package logs writes per-buffer daily log files and reads them back.
package logs

import (
	"bufio"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/eddyv73/wisirc/internal/irc"
	"github.com/eddyv73/wisirc/internal/state"
)

// maxOpenLogs caps how many log files stay open at once.
const maxOpenLogs = 32

// Logger appends to per-buffer daily files, keeping recently used files open
// so busy channels do not pay an open/close per line.
type Logger struct {
	Root string

	mu    sync.Mutex
	files map[string]*os.File
	order []string // open paths, least recently opened first
}

// DefaultRoot returns $XDG_DATA_HOME/wisirc/logs or ~/.local/share/wisirc/logs.
func DefaultRoot() string {
	base := os.Getenv("XDG_DATA_HOME")
	if base == "" {
		home, _ := os.UserHomeDir()
		base = filepath.Join(home, ".local", "share")
	}
	return filepath.Join(base, "wisirc", "logs")
}

var unsafeChars = strings.NewReplacer("/", "_", `\`, "_", "\x00", "_")

func safe(s string) string {
	s = unsafeChars.Replace(s)
	if s == "" || strings.HasPrefix(s, ".") {
		s = "_" + s
	}
	return s
}

func (l *Logger) dir(network, target string) string {
	name := "_status"
	if target != "" {
		name = safe(irc.Fold(target))
	}
	return filepath.Join(l.Root, safe(network), name)
}

func (l *Logger) Path(network, target string, day time.Time) string {
	return filepath.Join(l.dir(network, target), day.Local().Format("2006-01-02")+".log")
}

func (l *Logger) Append(network, target string, line state.Line) error {
	if line.Time.IsZero() {
		line.Time = time.Now()
	}
	p := l.Path(network, target, line.Time)
	l.mu.Lock()
	defer l.mu.Unlock()
	f, err := l.open(p)
	if err != nil {
		return err
	}
	_, err = f.WriteString(Format(line) + "\n")
	return err
}

// open returns the cached handle for p, opening it (and closing the oldest
// handle when the cache is full) if needed. Callers hold l.mu.
func (l *Logger) open(p string) (*os.File, error) {
	if f, ok := l.files[p]; ok {
		return f, nil
	}
	if err := os.MkdirAll(filepath.Dir(p), 0o700); err != nil {
		return nil, err
	}
	f, err := os.OpenFile(p, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return nil, err
	}
	if l.files == nil {
		l.files = map[string]*os.File{}
	}
	if len(l.order) >= maxOpenLogs {
		old := l.order[0]
		l.order = l.order[1:]
		l.files[old].Close()
		delete(l.files, old)
	}
	l.files[p] = f
	l.order = append(l.order, p)
	return f, nil
}

func (l *Logger) openFiles() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return len(l.files)
}

// Close closes every open log file; the logger stays usable afterwards.
func (l *Logger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	var first error
	for _, f := range l.files {
		if err := f.Close(); err != nil && first == nil {
			first = err
		}
	}
	l.files, l.order = nil, nil
	return first
}

// Tail returns the last n lines across the newest day files, oldest first.
func (l *Logger) Tail(network, target string, n int) ([]string, error) {
	files, err := filepath.Glob(filepath.Join(l.dir(network, target), "*.log"))
	if err != nil || len(files) == 0 {
		return nil, err
	}
	sort.Sort(sort.Reverse(sort.StringSlice(files)))
	var out []string
	for _, f := range files {
		lines, err := readLines(f)
		if err != nil {
			return out, err
		}
		out = append(lines, out...)
		if len(out) >= n {
			return out[len(out)-n:], nil
		}
	}
	return out, nil
}

func readLines(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var lines []string
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 64*1024), 1024*1024)
	for sc.Scan() {
		lines = append(lines, sc.Text())
	}
	return lines, sc.Err()
}

// Format renders a line irssi-style.
func Format(l state.Line) string {
	ts := l.Time.Local().Format("15:04:05")
	text := strings.ReplaceAll(l.Text, "\n", " ")
	switch l.Kind {
	case state.LineMsg:
		return ts + " <" + l.Nick + "> " + text
	case state.LineAction:
		return ts + " * " + l.Nick + " " + text
	case state.LineNotice:
		return ts + " -" + l.Nick + "- " + text
	default:
		return ts + " -!- " + text
	}
}
