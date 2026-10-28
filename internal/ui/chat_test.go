package ui

import (
	"testing"
	"time"

	"github.com/charmbracelet/x/ansi"
	"github.com/eddyv73/wisirc/internal/state"
)

func TestWrapLinesWide(t *testing.T) {
	ts := time.Date(2026, 10, 27, 12, 1, 0, 0, time.Local)
	lines := []state.Line{{Time: ts, Nick: "ana", Kind: state.LineMsg, Text: "日本語 emoji 😀 texto bastante largo para envolver varias veces 日本語日本語日本語"}}
	rows := WrapLines(lines, 20)
	if len(rows) < 3 {
		t.Fatalf("rows %d", len(rows))
	}
	for i, r := range rows {
		if w := ansi.StringWidth(r.Text); w > 20 {
			t.Fatalf("row %d width %d: %q", i, w, ansi.Strip(r.Text))
		}
	}
	if rows[0].NickStart != 6 || rows[0].NickEnd != 11 || rows[0].Nick != "ana" {
		t.Fatalf("nick span %d..%d %q", rows[0].NickStart, rows[0].NickEnd, rows[0].Nick)
	}
	if rows[1].NickStart != -1 {
		t.Fatal("continuation row has nick span")
	}
	if got := ansi.Strip(rows[0].Text)[:11]; got != "12:01 <ana>" {
		t.Fatalf("prefix %q", got)
	}
}

func TestWrapLinesWideNick(t *testing.T) {
	lines := []state.Line{{Time: time.Now(), Nick: "日本", Kind: state.LineMsg, Text: "hola"}}
	rows := WrapLines(lines, 40)
	if rows[0].NickStart != 6 || rows[0].NickEnd != 12 {
		t.Fatalf("span %d..%d", rows[0].NickStart, rows[0].NickEnd)
	}
}
