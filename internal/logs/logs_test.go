package logs

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/eddyv73/wisirc/internal/state"
)

func at(day, hms string) time.Time {
	t, err := time.ParseInLocation("2006-01-02 15:04:05", day+" "+hms, time.Local)
	if err != nil {
		panic(err)
	}
	return t
}

func TestAppendAndTail(t *testing.T) {
	l := &Logger{Root: t.TempDir()}
	for i, txt := range []string{"uno", "dos", "tres"} {
		line := state.Line{Time: at("2026-10-27", "12:00:0"+string(rune('0'+i))), Nick: "ana", Text: txt, Kind: state.LineMsg}
		if err := l.Append("irc.example.net", "#Lobby", line); err != nil {
			t.Fatal(err)
		}
	}
	got, err := l.Tail("irc.example.net", "#Lobby", 2)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"12:00:01 <ana> dos", "12:00:02 <ana> tres"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %q", got)
	}
	if _, err := os.Stat(l.Path("irc.example.net", "#lobby", at("2026-10-27", "00:00:00"))); err != nil {
		t.Fatal(err)
	}
}

func TestTailAcrossDays(t *testing.T) {
	l := &Logger{Root: t.TempDir()}
	l.Append("n", "#c", state.Line{Time: at("2026-10-26", "10:00:00"), Nick: "a", Text: "1"})
	l.Append("n", "#c", state.Line{Time: at("2026-10-26", "11:00:00"), Nick: "a", Text: "2"})
	l.Append("n", "#c", state.Line{Time: at("2026-10-27", "09:00:00"), Nick: "a", Text: "3"})
	got, _ := l.Tail("n", "#c", 3)
	want := []string{"10:00:00 <a> 1", "11:00:00 <a> 2", "09:00:00 <a> 3"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %q", got)
	}
}

func TestTailMissing(t *testing.T) {
	l := &Logger{Root: t.TempDir()}
	got, err := l.Tail("n", "#nada", 10)
	if got != nil || err != nil {
		t.Fatalf("%v %v", got, err)
	}
}

func TestSafePath(t *testing.T) {
	root := t.TempDir()
	l := &Logger{Root: root}
	p := l.Path("../evil", "a/../b", at("2026-10-27", "00:00:00"))
	rel, err := filepath.Rel(root, p)
	if err != nil || strings.HasPrefix(rel, "..") {
		t.Fatalf("escaped root: %q", p)
	}
	if p2 := l.Path("n", "..", time.Now()); strings.Contains(filepath.Dir(p2), string(filepath.Separator)+".."+string(filepath.Separator)) {
		t.Fatalf("dotdot dir: %q", p2)
	}
}

func TestStatusDir(t *testing.T) {
	l := &Logger{Root: "/r"}
	if p := l.Path("n", "", at("2026-10-27", "00:00:00")); p != filepath.Join("/r", "n", "_status", "2026-10-27.log") {
		t.Fatalf("%q", p)
	}
}

func TestFormat(t *testing.T) {
	ts := at("2026-10-27", "12:01:02")
	cases := map[string]state.Line{
		"12:01:02 * ana baila":     {Time: ts, Nick: "ana", Text: "baila", Kind: state.LineAction},
		"12:01:02 <ana> hola":      {Time: ts, Nick: "ana", Text: "hola", Kind: state.LineMsg},
		"12:01:02 -ana- aviso":     {Time: ts, Nick: "ana", Text: "aviso", Kind: state.LineNotice},
		"12:01:02 -!- ana se unió": {Time: ts, Nick: "ana", Text: "ana se unió", Kind: state.LineJoin},
	}
	for want, line := range cases {
		if got := Format(line); got != want {
			t.Errorf("got %q want %q", got, want)
		}
	}
}
