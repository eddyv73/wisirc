package ui

import (
	"fmt"
	"image/color"
	"math"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
)

func TestThemesComplete(t *testing.T) {
	want := []string{"oscuro", "claro", "nocturno", "clasico", "alto-contraste", "bosque", "glaciar", "papel",
		"dracula", "catppuccin-mocha", "catppuccin-latte", "tokyo-night", "one-dark", "solarized-oscuro",
		"rose-pine", "kanagawa", "everforest", "monokai", "fosforo", "ambar"}
	if got := ThemeNames(); strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("names %v", got)
	}
	for _, n := range want {
		th := themeFor(n)
		if th.name != n || th.accent == nil || th.dim == nil || th.mention == nil || th.errorC == nil || th.barBg == nil || len(th.nicks) == 0 {
			t.Fatalf("%s incomplete: %+v", n, th)
		}
	}
	if themeFor("dark").name != "oscuro" || themeFor("light").name != "claro" || themeFor("nope").name != "oscuro" {
		t.Fatal("aliases/fallback")
	}
	seen := map[string]bool{}
	for _, n := range ThemeNames() {
		if seen[n] || n != strings.ToLower(n) || strings.ContainsAny(n, " \t") {
			t.Errorf("bad or duplicate theme name %q", n)
		}
		seen[n] = true
	}
	for a, n := range themeAliases {
		if _, ok := lookupTheme(n); !ok {
			t.Errorf("alias %q -> unknown theme %q", a, n)
		}
	}
}

// luminance is the WCAG relative luminance of c.
func luminance(c color.Color) float64 {
	r, g, b, _ := c.RGBA()
	lin := func(v uint32) float64 {
		s := float64(v) / 0xffff
		if s <= 0.03928 {
			return s / 12.92
		}
		return math.Pow((s+0.055)/1.055, 2.4)
	}
	return 0.2126*lin(r) + 0.7152*lin(g) + 0.0722*lin(b)
}

// contrast is the WCAG contrast ratio between two colors (1 to 21).
func contrast(a, b color.Color) float64 {
	la, lb := luminance(a), luminance(b)
	if la < lb {
		la, lb = lb, la
	}
	return (la + 0.05) / (lb + 0.05)
}

func TestContrastHelper(t *testing.T) {
	if r := contrast(c("#000000"), c("#ffffff")); math.Abs(r-21) > 0.01 {
		t.Fatalf("black/white %v", r)
	}
	if r := contrast(c("#777777"), c("#ffffff")); math.Abs(r-4.48) > 0.01 {
		t.Fatalf("#777/white %v", r)
	}
}

// TestThemeContrast keeps every text color of themes with their own
// background at WCAG AA (4.5:1) or better.
func TestThemeContrast(t *testing.T) {
	for _, th := range themeList {
		if th.bg == nil {
			continue // ANSI themes depend on the terminal palette
		}
		check := func(what string, fg, bg color.Color) {
			if fg == nil || bg == nil {
				t.Errorf("%s: %s missing", th.name, what)
				return
			}
			if r := contrast(fg, bg); r < 4.5 {
				t.Errorf("%s: %s contrast %.2f < 4.5", th.name, what, r)
			}
		}
		for what, fg := range map[string]color.Color{
			"fg": th.fg, "dim": th.dim, "accent": th.accent, "mention": th.mention,
			"unread": th.unread, "error": th.errorC, "ok": th.ok,
		} {
			check(what, fg, th.bg)
		}
		for i, n := range th.nicks {
			check(fmt.Sprintf("nick %d", i), n, th.bg)
		}
		check("bar", th.barFg, th.barBg)
		if th.selBg != nil {
			check("selection", th.selFg, th.selBg)
		}
	}
}

func TestThemeSubmenuPreviewRevertConfirm(t *testing.T) {
	h := newHarness(t)
	var saved []string
	h.m.d.SaveTheme = func(n string) error { saved = append(saved, n); return nil }
	h.m.runMenu(GlobalMenu(), ActTheme)
	if h.m.menu == nil || h.m.menu.Kind != MenuTheme {
		t.Fatal("no theme submenu")
	}
	v := h.view()
	if !strings.Contains(v, "● oscuro") || !strings.Contains(v, "bosque") {
		t.Fatalf("submenu:\n%s", v)
	}
	h.do(tea.KeyPressMsg{Code: tea.KeyDown})
	if h.m.theme.name != "claro" {
		t.Fatalf("no live preview: %q", h.m.theme.name)
	}
	h.do(tea.KeyPressMsg{Code: tea.KeyEscape})
	if h.m.theme.name != "oscuro" || h.m.menu != nil || len(saved) != 0 {
		t.Fatalf("esc did not revert: %q saved %v", h.m.theme.name, saved)
	}
	h.m.runMenu(GlobalMenu(), ActTheme)
	h.do(tea.KeyPressMsg{Code: tea.KeyDown})
	h.do(tea.KeyPressMsg{Code: tea.KeyDown})
	h.do(tea.KeyPressMsg{Code: tea.KeyEnter})
	if h.m.theme.name != "nocturno" || h.m.menu != nil || len(saved) != 1 || saved[0] != "nocturno" {
		t.Fatalf("confirm: %q saved %v", h.m.theme.name, saved)
	}
	if h.m.View().BackgroundColor == nil {
		t.Fatal("nocturno should paint the background")
	}
}

func TestThemeSubmenuClick(t *testing.T) {
	h := newHarness(t)
	h.m.d.SaveTheme = func(string) error { return nil }
	h.m.runMenu(GlobalMenu(), ActTheme)
	x, y := h.find("bosque")
	h.do(click(x, y, tea.MouseLeft))
	if h.m.theme.name != "bosque" || h.m.menu != nil {
		t.Fatalf("click: %q", h.m.theme.name)
	}
}

func TestThemeCommand(t *testing.T) {
	h := newHarness(t)
	h.m.d.SaveTheme = func(string) error { return nil }
	h.typeText("/tema glaciar")
	h.do(tea.KeyPressMsg{Code: tea.KeyEnter})
	if h.m.theme.name != "glaciar" {
		t.Fatalf("theme %q", h.m.theme.name)
	}
	h.typeText("/tema nada")
	h.do(tea.KeyPressMsg{Code: tea.KeyEnter})
	if h.m.theme.name != "glaciar" || !strings.Contains(h.view(), "Themes:") {
		t.Fatalf("bad name handling:\n%s", h.view())
	}
	if h.m.View().BackgroundColor == nil {
		t.Fatal("glaciar has a background")
	}
	h.typeText("/tema oscuro")
	h.do(tea.KeyPressMsg{Code: tea.KeyEnter})
	if h.m.View().BackgroundColor != nil {
		t.Fatal("oscuro uses the terminal background")
	}
}
