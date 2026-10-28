package logs

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/eddyv73/wisirc/internal/state"
)

// 3. the logger keeps files open, caps how many, and loses nothing
func TestAppendKeepsFilesOpen(t *testing.T) {
	l := &Logger{Root: t.TempDir()}
	day := time.Date(2026, 10, 28, 10, 0, 0, 0, time.Local)
	for i := 0; i < 3; i++ {
		if err := l.Append("n", "#c", state.Line{Time: day, Nick: "a", Text: fmt.Sprint("m", i), Kind: state.LineMsg}); err != nil {
			t.Fatal(err)
		}
	}
	if l.openFiles() != 1 {
		t.Fatalf("%d open files", l.openFiles())
	}
	for i := 0; i < maxOpenLogs+10; i++ {
		l.Append("n", fmt.Sprint("#x", i), state.Line{Time: day, Text: "y"})
	}
	if l.openFiles() > maxOpenLogs {
		t.Fatalf("%d open files, cap %d", l.openFiles(), maxOpenLogs)
	}
	// the next day goes to its own file
	l.Append("n", "#c", state.Line{Time: day.Add(24 * time.Hour), Text: "tomorrow"})
	if err := l.Close(); err != nil {
		t.Fatal(err)
	}
	b, _ := os.ReadFile(l.Path("n", "#c", day))
	if got := strings.Count(string(b), "\n"); got != 3 || !strings.Contains(string(b), "m2") {
		t.Fatalf("today's file: %q", b)
	}
	if b, _ := os.ReadFile(l.Path("n", "#c", day.Add(24*time.Hour))); !strings.Contains(string(b), "tomorrow") {
		t.Fatalf("tomorrow's file: %q", b)
	}
	// still usable after Close
	if err := l.Append("n", "#c", state.Line{Time: day, Text: "again"}); err != nil {
		t.Fatal(err)
	}
	l.Close()
}
