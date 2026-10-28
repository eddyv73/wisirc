package ui

import (
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/eddyv73/wisirc/internal/config"
	"github.com/eddyv73/wisirc/internal/i18n"
	"github.com/eddyv73/wisirc/internal/irc"
	"github.com/eddyv73/wisirc/internal/state"
)

func dashHarness(t *testing.T) *harness {
	h := newHarness(t)
	h.joinLobby()
	for _, d := range []time.Duration{20, 40, 42} {
		h.send(eventMsg(irc.Event{Network: netName, Kind: irc.EvLag, Lag: d * time.Millisecond}))
	}
	h.irc(":ana!a@h PRIVMSG alice :hola")
	h.nets.cfg.Servers["libera"] = config.Server{Server: "irc.libera.chat", Port: 6697}
	return h
}

func TestDashboardCards(t *testing.T) {
	h := dashHarness(t)
	h.do(tea.KeyPressMsg{Code: tea.KeyF2})
	if h.m.dash == nil {
		t.Fatal("F2 did not open the dashboard")
	}
	v := h.view()
	for _, want := range []string{"Dashboard", netName, "connected", "▁▂▂", "42ms", "#lobby", "1 new", "libera", "disconnected", "⏱"} {
		if !strings.Contains(v, want) {
			t.Fatalf("missing %q:\n%s", want, v)
		}
	}
	h.do(tea.KeyPressMsg{Code: tea.KeyF2})
	if h.m.dash != nil {
		t.Fatal("F2 did not close it")
	}
}

func TestDashboardFromMenuAndCommand(t *testing.T) {
	h := dashHarness(t)
	if labels(GlobalMenu())[0] != "Dashboard" {
		t.Fatalf("menu %v", labels(GlobalMenu()))
	}
	h.m.runMenu(GlobalMenu(), ActDashboard)
	if h.m.dash == nil {
		t.Fatal("menu did not open it")
	}
	h.do(tea.KeyPressMsg{Code: tea.KeyEscape})
	h.typeText("/panel")
	h.do(tea.KeyPressMsg{Code: tea.KeyEnter})
	if h.m.dash == nil {
		t.Fatal("/panel did not open it")
	}
}

func TestDashboardPingsOftenWhileOpen(t *testing.T) {
	h := dashHarness(t)
	h.do(tea.KeyPressMsg{Code: tea.KeyF2})
	before := len(h.nets.sent())
	for i := 0; i < dashPingFrames; i++ {
		h.m.Update(animTickMsg{})
	}
	var pings int
	for _, l := range h.nets.sent()[before:] {
		if strings.HasPrefix(l, "PING wisirc-") {
			pings++
		}
	}
	if pings != 1 {
		t.Fatalf("pings %d in %v", pings, h.nets.sent()[before:])
	}
}

func TestDashboardClickGoesToServer(t *testing.T) {
	h := dashHarness(t)
	h.m.switchTo(state.BufRef{Network: netName, Name: "#lobby"})
	h.do(tea.KeyPressMsg{Code: tea.KeyF2})
	x, y := h.find("═ " + netName)
	h.do(click(x+2, y+2, tea.MouseLeft))
	h.do(tea.MouseReleaseMsg{X: x + 2, Y: y + 2, Button: tea.MouseLeft})
	if h.m.dash != nil || h.m.State().Active != (state.BufRef{Network: netName}) {
		t.Fatalf("dash %v active %+v", h.m.dash != nil, h.m.State().Active)
	}
}

func TestDashboardRightClickServerMenu(t *testing.T) {
	h := dashHarness(t)
	h.do(tea.KeyPressMsg{Code: tea.KeyF2})
	x, y := h.find("─ libera")
	h.do(click(x, y+1, tea.MouseRight))
	if h.m.menu == nil || h.m.menu.Kind != MenuServer || h.m.menu.Ref.Network != "libera" {
		t.Fatalf("menu %+v", h.m.menu)
	}
}

func TestDashboardSpanish(t *testing.T) {
	t.Cleanup(func() { i18n.SetLang("en") })
	h := dashHarness(t)
	i18n.SetLang("es")
	h.do(tea.KeyPressMsg{Code: tea.KeyF2})
	if v := h.view(); !strings.Contains(v, "Panel") || !strings.Contains(v, "conectado") {
		t.Fatalf("not Spanish:\n%s", v)
	}
}

func TestUptime(t *testing.T) {
	for d, want := range map[time.Duration]string{
		40 * time.Second:             "40s",
		12 * time.Minute:             "12m",
		2*time.Hour + 14*time.Minute: "2h 14m",
		50*time.Hour + 3*time.Minute: "2d 2h",
	} {
		if got := uptime(d); got != want {
			t.Errorf("%v: %q want %q", d, got, want)
		}
	}
}

func TestStartOnDashboard(t *testing.T) {
	cfg := &config.Config{Servers: map[string]config.Server{netName: {Server: "x"}}, Managed: map[string]bool{}}
	nets := &fakeNets{senders: map[string]*fakeSender{netName: {}}, cfg: cfg, adhoc: map[string]bool{}}
	m := New(Deps{Cfg: cfg, Nets: nets, Events: make(chan irc.Event), StartOnDashboard: true})
	if m.dash == nil {
		t.Fatal("did not start on the dashboard")
	}
	// a first run with nothing configured but Halloy around still opens Servers
	empty := &config.Config{Servers: map[string]config.Server{}, Managed: map[string]bool{}}
	none := &fakeNets{senders: map[string]*fakeSender{}, cfg: empty, adhoc: map[string]bool{}, none: true}
	m = New(Deps{Cfg: empty, Nets: none, Events: make(chan irc.Event), StartOnDashboard: true,
		Halloy: func() (map[string]config.Server, []string, error) { return nil, nil, nil }})
	if m.dash != nil || m.admin == nil {
		t.Fatal("first run with Halloy should open Servers, not the dashboard")
	}
}
