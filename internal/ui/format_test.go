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

// Terminals disagree on how wide these sequences are, so wisirc must only
// emit text whose width is the same whether measured by grapheme or by rune.
func TestStableWidth(t *testing.T) {
	for _, in := range []string{
		"Mainly Clear 🌤️ │ Rain 🌦️",
		"❤️ 1️⃣ ☀︎",
		"👍🏽 ok",
		"👨‍👩‍👧 family",
		"❤️‍🔥",
		"🇺🇸 🦃 plain",
	} {
		for name, got := range map[string]string{"RenderIRC": RenderIRC(in), "sanitize": sanitize(in)} {
			if g, w := ansi.StringWidth(got), ansi.StringWidthWc(got); g != w {
				t.Errorf("%s(%q) = %q: grapheme width %d, terminal width %d", name, in, got, g, w)
			}
		}
	}
}

func TestStableWidthKeepsJoinersInText(t *testing.T) {
	const persian = "می‌خواهم" // ZWNJ, untouched
	const hindi = "क्‍ष"       // ZWJ between letters changes the glyph
	if got := RenderIRC(persian); got != persian {
		t.Errorf("RenderIRC(%q) = %q", persian, got)
	}
	if got := RenderIRC(hindi); got != hindi {
		t.Errorf("RenderIRC(%q) = %q", hindi, got)
	}
	if got := sanitize("🏻"); got != "🏻" {
		t.Errorf("a lone skin tone was dropped: %q", got)
	}
}
