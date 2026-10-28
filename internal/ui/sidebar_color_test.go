package ui

import (
	"fmt"
	"image/color"
	"strings"
	"testing"

	"github.com/eddyv73/wisirc/internal/irc"
	"github.com/eddyv73/wisirc/internal/state"
)

func bgSeq(c color.Color) string {
	r, g, b, _ := c.RGBA()
	return fmt.Sprintf("48;2;%d;%d;%d", r>>8, g>>8, b>>8)
}

func TestServerColorsDiffer(t *testing.T) {
	seen := map[string]bool{}
	for i := 0; i < len(serverColors); i++ {
		s := bgSeq(serverColor(i))
		if seen[s] {
			t.Fatalf("color %d repeats", i)
		}
		seen[s] = true
	}
	if bgSeq(serverColor(len(serverColors))) != bgSeq(serverColor(0)) {
		t.Fatal("palette should wrap around")
	}
}

func TestActiveRowUsesItsServerColor(t *testing.T) {
	h := newHarness(t)
	h.joinLobby()
	m2, _ := irc.Parse(":srv 001 alice :hi")
	h.send(eventMsg(irc.Event{Network: "otra", Kind: irc.EvConnected, Msg: m2}))
	j, _ := irc.Parse(":alice!k@h JOIN #b")
	h.send(eventMsg(irc.Event{Network: "otra", Kind: irc.EvMessage, Msg: j}))

	h.m.switchTo(state.BufRef{Network: netName, Name: "#lobby"})
	side := strings.Join(h.m.renderSidebar(h.m.Layout()), "\n")
	if !strings.Contains(side, bgSeq(serverColor(0))) {
		t.Fatalf("first server's row is not in its color:\n%q", side)
	}
	h.m.switchTo(state.BufRef{Network: "otra", Name: "#b"})
	side = strings.Join(h.m.renderSidebar(h.m.Layout()), "\n")
	if !strings.Contains(side, bgSeq(serverColor(1))) || strings.Contains(side, bgSeq(serverColor(0))) {
		t.Fatalf("second server's row is not in its own color:\n%q", side)
	}
}
