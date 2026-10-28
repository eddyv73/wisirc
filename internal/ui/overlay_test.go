package ui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
)

func TestOverlayPlain(t *testing.T) {
	if got := Overlay("aaaaa\nbbbbb\nccccc", "XY\nZW", 1, 1); got != "aaaaa\nbXYbb\ncZWcc" {
		t.Fatalf("got %q", got)
	}
}

func TestOverlayANSIAndWide(t *testing.T) {
	base := "\x1b[31m日本語abc\x1b[0m"
	got := Overlay(base, "XX", 2, 0)
	if s := ansi.Strip(got); s != "日XX語abc" {
		t.Fatalf("stripped %q", s)
	}
	if w := ansi.StringWidth(got); w != 9 {
		t.Fatalf("width %d", w)
	}
}

func TestOverlayBeyondBase(t *testing.T) {
	got := strings.Split(Overlay("ab", "Z", 5, 2), "\n")
	if len(got) != 3 || got[2] != "     Z" {
		t.Fatalf("got %q", got)
	}
}
