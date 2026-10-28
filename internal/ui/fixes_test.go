package ui

import (
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
)

// 1. the status bar's lag color stays readable on every theme's bar
func TestStatusLagReadableOnEveryBar(t *testing.T) {
	for _, th := range themeList {
		for _, d := range []time.Duration{10 * time.Millisecond, 300 * time.Millisecond, 2 * time.Second} {
			col := barLagColor(th, d)
			if th.barBg == nil || !isHex(th.barBg) || !isHex(col) {
				continue
			}
			if r := contrastRatio(col, th.barBg); r < 3 {
				t.Errorf("%s %v: lag color contrast %.2f on the bar", th.name, d, r)
			}
		}
	}
}

// 2. mIRC colors that vanish on light backgrounds are darkened there
func TestIRCColorsOnLightTheme(t *testing.T) {
	t.Cleanup(func() { applyStyles(themeList[0]) })
	applyStyles(themeFor("papel"))
	if out := RenderIRC("\x0308yellow\x0f \x0309green"); strings.Contains(out, "38;5;11") || strings.Contains(out, "38;5;10") {
		t.Fatalf("light colors kept on a light theme: %q", out)
	}
	if out := RenderIRC("\x0308,01on black"); !strings.Contains(out, "38;5;11") {
		t.Fatalf("explicit background must keep the sender's color: %q", out)
	}
	applyStyles(themeFor("oscuro"))
	if out := RenderIRC("\x0308yellow"); !strings.Contains(out, "38;5;11") {
		t.Fatalf("dark theme changed colors: %q", out)
	}
}

func TestServerGlyphReadableOnLightTheme(t *testing.T) {
	h := newHarness(t)
	t.Cleanup(func() { h.m.setTheme(themeList[0]) })
	h.joinLobby()
	h.m.setTheme(themeFor("papel"))
	h.m.tints[netName] = "#f4e04d" // yellow: invisible on paper unless darkened
	row := sidebarRow(h, netName)
	if !strings.Contains(row, fgSeq(h.m.channelTint(netName))) {
		t.Fatalf("glyph not darkened: %q", row)
	}
}

// 3. menus never cover the input line or the status bar
func TestMenusStayAboveInput(t *testing.T) {
	h := newHarness(t)
	h.send(tea.WindowSizeMsg{Width: 100, Height: 18})
	h.m.openThemeMenu()
	rows := strings.Split(h.view(), "\n")
	if !strings.Contains(rows[len(rows)-2], ">") || !strings.Contains(rows[len(rows)-1], "disconnected") {
		t.Fatalf("menu covers input/status:\n%s", h.view())
	}
	r := h.m.menu.Rect(h.m.w, h.m.menuH())
	if r.Y+r.H > h.m.h-2 {
		t.Fatalf("rect %+v reaches the input row", r)
	}
}

// 4. the connect screen tells how saving works today
func TestConnectTextMentionsSaving(t *testing.T) {
	text := strings.Join(connectText(), "\n")
	if strings.Contains(text, "Edit config") || !strings.Contains(text, "Servers") {
		t.Fatalf("outdated hint:\n%s", text)
	}
}

// 5. context menus say whose they are and open below the clicked row
func TestNickMenuTitledAndBelow(t *testing.T) {
	h := newHarness(t)
	h.joinLobby()
	h.irc(":srv 353 alice = #lobby :alice @ana")
	h.irc(":srv 366 alice #lobby :end")
	x, y := h.find("ana")
	h.do(click(x, y, tea.MouseRight))
	if h.m.menu == nil || h.m.menu.Kind != MenuNick || h.m.menu.Title != "ana" {
		t.Fatalf("menu %+v", h.m.menu)
	}
	if r := h.m.menu.Rect(h.m.w, h.m.menuH()); r.Y <= y {
		t.Fatalf("menu covers the nick: rect %+v click y %d", r, y)
	}
	if v := h.view(); !strings.Contains(strings.Split(v, "\n")[y], "ana") {
		t.Fatalf("nick hidden:\n%s", v)
	}
}
