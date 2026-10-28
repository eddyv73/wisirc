package ui

import (
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/eddyv73/wisirc/internal/config"
	"github.com/eddyv73/wisirc/internal/irc"
	"github.com/eddyv73/wisirc/internal/state"
)

func barHarness(t *testing.T) *harness {
	h := newHarness(t)
	h.joinLobby()
	h.irc(":srv 353 alice = #lobby :alice @ana")
	h.irc(":srv 324 alice #lobby +nt")
	h.irc(":alice MODE alice :+i")
	h.irc(":alice!k@h JOIN #b")
	h.m.switchTo(state.BufRef{Network: netName, Name: "#lobby"})
	h.send(eventMsg(irc.Event{Network: netName, Kind: irc.EvLag, Lag: 42 * time.Millisecond}))
	h.irc(":bob!b@h PRIVMSG #b :hola")
	h.irc(":bob!b@h PRIVMSG #b :alice: mira")
	h.nets.cfg.Servers["plain"] = config.Server{Server: "x", Port: 6667}
	m2, _ := irc.Parse(":srv 001 alice :hi")
	h.send(eventMsg(irc.Event{Network: "plain", Kind: irc.EvConnected, Msg: m2}))
	h.send(eventMsg(irc.Event{Network: "offline", Kind: irc.EvDisconnected, Text: "x"}))
	h.m.switchTo(state.BufRef{Network: netName, Name: "#lobby"})
	return h
}

func TestStatusBarSegments(t *testing.T) {
	h := barHarness(t)
	st := h.statusRow()
	for _, want := range []string{"● ● ○", netName + " tls", "42ms", "#b 2", "!1", "alice(+i)", "#lobby(+nt) 2", time.Now().Format("15:04")} {
		if !strings.Contains(st, want) {
			t.Fatalf("missing %q in %q", want, st)
		}
	}
	h.m.switchTo(state.BufRef{Network: "plain"})
	if st := h.statusRow(); !strings.Contains(st, "plain no tls") {
		t.Fatalf("plain connection not flagged: %q", st)
	}
}

func (h *harness) clickStatus(text string) {
	h.t.Helper()
	st := h.statusRow()
	i := strings.Index(st, text)
	if i < 0 {
		h.t.Fatalf("%q not in status %q", text, st)
	}
	x := len([]rune(st[:i]))
	h.do(click(x, h.m.h-1, tea.MouseLeft))
}

func TestStatusBarClicks(t *testing.T) {
	h := barHarness(t)
	h.clickStatus("#b 2")
	if got := h.m.State().Active; got.Name != "#b" {
		t.Fatalf("chip click → %+v", got)
	}
	h.m.switchTo(state.BufRef{Network: netName, Name: "#lobby"})
	h.irc(":bob!b@h PRIVMSG #b :alice: otra vez")
	h.clickStatus("!1")
	if got := h.m.State().Active; got.Name != "#b" {
		t.Fatalf("mention click → %+v", got)
	}
	st := h.statusRow()
	dots := strings.Index(st, "● ● ○")
	h.do(click(len([]rune(st[:dots]))+2, h.m.h-1, tea.MouseLeft)) // second dot: "plain"
	if got := h.m.State().Active; got != (state.BufRef{Network: "plain"}) {
		t.Fatalf("dot click → %+v", got)
	}
}

func TestStatusBarNarrow(t *testing.T) {
	h := barHarness(t)
	h.send(tea.WindowSizeMsg{Width: 60, Height: 20})
	st := h.statusRow()
	if w := len([]rune(st)); w > 60 {
		t.Fatalf("status %d columns wide: %q", w, st)
	}
	if !strings.Contains(st, "● ● ○") || !strings.Contains(st, netName) {
		t.Fatalf("essentials dropped: %q", st)
	}
	if strings.Contains(st, time.Now().Format("15:04")) {
		t.Fatalf("right side should go first: %q", st)
	}
}

func TestJoinAsksChannelModes(t *testing.T) {
	h := newHarness(t)
	h.joinLobby()
	found := false
	for _, l := range h.nets.sent() {
		if l == "MODE #lobby" {
			found = true
		}
	}
	if !found {
		t.Fatalf("sent %q", h.nets.sent())
	}
}

func TestClickWisircOpensDashboard(t *testing.T) {
	h := newHarness(t)
	x, y := h.find("wisirc")
	h.do(click(x+1, y, tea.MouseLeft))
	if h.m.dash == nil {
		t.Fatal("clicking the wisirc title did not open the dashboard")
	}
	h.do(click(x+1, y, tea.MouseLeft))
	if h.m.dash == nil {
		t.Fatal("a second click should keep the dashboard open")
	}
}

func TestStatusBarManyChipsKeepRight(t *testing.T) {
	h := barHarness(t)
	for i := 0; i < 8; i++ {
		ch := "#c" + string(rune('a'+i))
		h.irc(":alice!k@h JOIN " + ch)
		h.irc(":bob!b@h PRIVMSG " + ch + " :hey")
	}
	h.m.switchTo(state.BufRef{Network: netName, Name: "#lobby"})
	h.send(tea.WindowSizeMsg{Width: 110, Height: 20})
	st := h.statusRow()
	if !strings.Contains(st, time.Now().Format("15:04")) || !strings.Contains(st, "alice(+i)") {
		t.Fatalf("right side dropped for chips: %q", st)
	}
	if !strings.Contains(st, "+") || strings.Count(st, " 1") > 4 {
		t.Fatalf("no +N overflow: %q", st)
	}
	h.clickStatus("+")
	if h.m.State().Active.Name == "#lobby" {
		t.Fatal("+N did not jump to activity")
	}
}
