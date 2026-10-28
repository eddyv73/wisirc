package ui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
)

func TestRenderIRC(t *testing.T) {
	out := RenderIRC("\x02hola\x02 \x0304rojo\x0f fin")
	if s := ansi.Strip(out); s != "hola rojo fin" {
		t.Fatalf("stripped %q", s)
	}
	if !strings.Contains(out, "\x1b[") {
		t.Fatal("no styling")
	}
	if s := ansi.Strip(RenderIRC("\x0399")); s != "" {
		t.Fatalf("invalid color: %q", s)
	}
	if got := RenderIRC("texto plano"); got != "texto plano" {
		t.Fatalf("plain changed: %q", got)
	}
}

func TestRenderIRCDropsTerminalEscapes(t *testing.T) {
	out := RenderIRC("a\x1b[2Jb\x07c")
	if strings.Contains(out, "\x1b[2J") || strings.Contains(out, "\x07") {
		t.Fatalf("escape passed through: %q", out)
	}
	if s := ansi.Strip(out); s != "a[2Jbc" {
		t.Fatalf("got %q", s)
	}
}

func TestNickColorStable(t *testing.T) {
	if NickColor("Alice") != NickColor("alice") {
		t.Fatal("case-sensitive color")
	}
}

func TestStripsC1Controls(t *testing.T) {
	if out := RenderIRC("a\u009b2Jb\u0085c"); strings.ContainsAny(out, "\u009b\u0085") {
		t.Fatalf("C1 passed through: %q", out)
	}
	if out := sanitize("ni\u009bck"); out != "nick" {
		t.Fatalf("sanitize %q", out)
	}
}
