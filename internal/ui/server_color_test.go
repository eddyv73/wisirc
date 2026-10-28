package ui

import (
	"fmt"
	"image/color"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/eddyv73/wisirc/internal/state"
)

func fgSeq(c color.Color) string {
	r, g, b, _ := c.RGBA()
	return fmt.Sprintf("38;2;%d;%d;%d", r>>8, g>>8, b>>8)
}

func sidebarRow(h *harness, label string) string {
	for _, row := range h.m.renderSidebar(h.m.Layout()) {
		if strings.Contains(row, label) {
			return row
		}
	}
	h.t.Fatalf("no sidebar row %q", label)
	return ""
}

func TestChannelsUseServerTint(t *testing.T) {
	h := newHarness(t)
	h.joinLobby()
	h.irc(":alice!k@h JOIN #otro")
	h.m.switchTo(state.BufRef{Network: netName, Name: "#otro"})
	row := sidebarRow(h, "#lobby")
	if !strings.Contains(row, fgSeq(h.m.channelTint(netName))) {
		t.Fatalf("channel not tinted: %q", row)
	}
}

func TestServerColorMenu(t *testing.T) {
	h := newHarness(t)
	h.joinLobby()
	var saved [][2]string
	h.m.d.SaveServerColor = func(n, c string) error { saved = append(saved, [2]string{n, c}); return nil }
	auto := h.m.serverTint(netName)

	h.do(click(3, 1, tea.MouseRight)) // the server row
	if h.m.menu == nil || h.m.menu.Kind != MenuServer {
		t.Fatal("no server menu")
	}
	if !strings.Contains(strings.Join(labels(*h.m.menu), ","), "Color…") {
		t.Fatalf("server menu %v", labels(*h.m.menu))
	}
	h.m.runMenu(*h.m.menu, ActServerColor)
	if h.m.menu == nil || h.m.menu.Kind != MenuColor || h.m.menu.Args[0] != "" {
		t.Fatalf("color menu %+v", h.m.menu)
	}
	h.do(tea.KeyPressMsg{Code: tea.KeyDown})
	picked := h.m.menu.Args[h.m.menu.Selected]
	if bgSeq(h.m.serverTint(netName)) != bgSeq(c(picked)) {
		t.Fatal("no live preview")
	}
	h.do(tea.KeyPressMsg{Code: tea.KeyEscape})
	if bgSeq(h.m.serverTint(netName)) != bgSeq(auto) || len(saved) != 0 {
		t.Fatal("esc did not revert")
	}
	h.m.runMenu(ServerMenu(state.BufRef{Network: netName}, true, 3, 1), ActServerColor)
	h.do(tea.KeyPressMsg{Code: tea.KeyDown})
	h.do(tea.KeyPressMsg{Code: tea.KeyDown})
	h.do(tea.KeyPressMsg{Code: tea.KeyEnter})
	want := serverColorChoices[1].hex
	if len(saved) != 1 || saved[0] != [2]string{netName, want} || bgSeq(h.m.serverTint(netName)) != bgSeq(c(want)) {
		t.Fatalf("saved %v", saved)
	}
	// Automático goes back to the sidebar-order shade
	h.m.runMenu(ServerMenu(state.BufRef{Network: netName}, true, 3, 1), ActServerColor)
	h.m.menu.Selected = 0
	h.do(tea.KeyPressMsg{Code: tea.KeyEnter})
	if bgSeq(h.m.serverTint(netName)) != bgSeq(auto) || saved[len(saved)-1][1] != "" {
		t.Fatalf("auto: saved %v", saved)
	}
}

func TestServerColorChoicesReadable(t *testing.T) {
	for _, ch := range serverColorChoices {
		if r := contrast(c(ch.hex), c("#0b0f0c")); r < 4.5 {
			t.Errorf("%s %s: dark text contrast %.2f", ch.name, ch.hex, r)
		}
	}
	if len(serverColorChoices) < 10 {
		t.Fatalf("only %d colors", len(serverColorChoices))
	}
}
