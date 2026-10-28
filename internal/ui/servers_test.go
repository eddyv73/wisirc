package ui

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/eddyv73/wisirc/internal/config"
	"github.com/eddyv73/wisirc/internal/irc"
)

func key(r rune) tea.KeyPressMsg { return tea.KeyPressMsg{Code: r, Text: string(r)} }

func (h *harness) connected(name string) {
	h.t.Helper()
	m, _ := irc.Parse(":srv 001 alice :hi")
	h.send(eventMsg(irc.Event{Network: name, Kind: irc.EvConnected, Msg: m}))
}

func TestAdHocConnectAsksToSave(t *testing.T) {
	h := newHarness(t)
	h.nets.adhoc["irc.example.org"] = true
	h.connected("irc.example.org")
	if h.m.menu == nil || h.m.menu.Kind != MenuConfirm || !strings.Contains(h.view(), "Save irc.example.org") {
		t.Fatalf("no save prompt:\n%s", h.view())
	}
	h.do(key('s'))
	if h.m.menu != nil || len(h.nets.savedAdh) != 1 || h.nets.savedAdh[0] != "irc.example.org" {
		t.Fatalf("menu %v saved %v", h.m.menu, h.nets.savedAdh)
	}
	// it only asks once per connection name
	h.connected("irc.example.org")
	if h.m.menu != nil {
		t.Fatal("asked twice")
	}
}

func TestSavePromptNo(t *testing.T) {
	h := newHarness(t)
	h.nets.adhoc["x"] = true
	h.connected("x")
	h.do(key('n'))
	if h.m.menu != nil || len(h.nets.savedAdh) != 0 {
		t.Fatal("saved after no")
	}
	if !strings.Contains(h.view(), "Servers") {
		t.Fatal("no hint about the Servers screen")
	}
}

func TestSavedServerDoesNotAsk(t *testing.T) {
	h := newHarness(t)
	h.connected(netName)
	if h.m.menu != nil {
		t.Fatal("asked to save a configured server")
	}
}

func TestServersScreenLists(t *testing.T) {
	h := newHarness(t)
	h.nets.cfg.Servers["znc"] = config.Server{Server: "192.0.2.10", Port: 7000, Password: "x"}
	h.nets.cfg.Managed["znc"] = true
	h.connected(netName)
	h.m.runMenu(GlobalMenu(), ActServers)
	v := h.view()
	for _, want := range []string{"Servers", netName, "irc.example.net:6697", "config.toml", "znc", "192.0.2.10:7000", "servers.toml", "[+ New]"} {
		if !strings.Contains(v, want) {
			t.Fatalf("missing %q:\n%s", want, v)
		}
	}
	h.do(tea.KeyPressMsg{Code: tea.KeyEscape})
	if h.m.admin != nil {
		t.Fatal("esc did not close")
	}
}

func TestServersScreenNewServer(t *testing.T) {
	h := newHarness(t)
	h.m.runMenu(GlobalMenu(), ActServers)
	x, y := h.find("[+ New]")
	h.do(click(x+1, y, tea.MouseLeft))
	h.do(tea.MouseReleaseMsg{X: x + 1, Y: y, Button: tea.MouseLeft})
	if h.m.admin == nil || h.m.admin.form == nil {
		t.Fatal("form not open")
	}
	tab := tea.KeyPressMsg{Code: tea.KeyTab}
	h.typeText("znc")
	h.do(tab)
	h.typeText("192.0.2.5")
	h.do(tab)
	h.m.admin.form.fields[fPort].input.SetValue("")
	h.typeText("7000")
	h.do(tab)
	h.do(key(' ')) // TLS off
	h.do(tab)
	h.do(tab)
	h.m.admin.form.fields[fNick].input.SetValue("")
	h.typeText("zephyr")
	h.do(tab)
	h.typeText("bob/libera")
	h.do(tab)
	h.typeText("secreto")
	if strings.Contains(h.view(), "secreto") {
		t.Fatal("password shown in clear")
	}
	h.do(tea.KeyPressMsg{Code: tea.KeyEnter})
	s, ok := h.nets.cfg.Servers["znc"]
	if !ok {
		t.Fatalf("not saved; screen:\n%s", h.view())
	}
	if s.Server != "192.0.2.5" || s.Port != 7000 || s.TLS() || s.Nickname != "zephyr" || s.Username != "bob/libera" || s.Password != "secreto" || !s.Auto() {
		t.Fatalf("saved %+v", s)
	}
	if h.m.admin.form != nil || !strings.Contains(h.view(), "znc") {
		t.Fatal("did not return to the list")
	}
}

func TestServersScreenValidates(t *testing.T) {
	h := newHarness(t)
	h.m.runMenu(GlobalMenu(), ActServers)
	h.do(key('n'))
	h.do(tea.KeyPressMsg{Code: tea.KeyEnter})
	if h.m.admin.form == nil || !strings.Contains(h.view(), "name") {
		t.Fatalf("empty form accepted:\n%s", h.view())
	}
}

func TestServersScreenDeleteConfirms(t *testing.T) {
	h := newHarness(t)
	h.nets.cfg.Servers["zz"] = config.Server{Server: "z"}
	h.nets.cfg.Managed["zz"] = true
	h.m.runMenu(GlobalMenu(), ActServers)
	h.do(tea.KeyPressMsg{Code: tea.KeyDown})
	h.do(key('d'))
	if h.m.menu == nil || h.m.menu.Kind != MenuConfirm {
		t.Fatal("no confirmation")
	}
	h.do(key('s'))
	if _, ok := h.nets.cfg.Servers["zz"]; ok {
		t.Fatal("not deleted")
	}
}

func TestServersScreenConfigTomlReadOnly(t *testing.T) {
	h := newHarness(t)
	h.m.runMenu(GlobalMenu(), ActServers)
	h.do(key('d'))
	if h.m.menu != nil {
		t.Fatal("offered to delete a config.toml server")
	}
	if !strings.Contains(h.view(), "config.toml") {
		t.Fatal("no explanation")
	}
}

func TestImportHalloyOneClick(t *testing.T) {
	h := newHarness(t)
	h.m.d.Halloy = func() (map[string]config.Server, []string, error) {
		return map[string]config.Server{
			netName:      {Server: "irc.example.net"},
			"znc":        {Server: "192.0.2.10", Port: 7000},
			"znc-libera": {Server: "192.0.2.10", Port: 7000},
		}, []string{"znc-libera: password_command is not run, for safety"}, nil
	}
	h.m.runMenu(GlobalMenu(), ActServers)
	x, y := h.find("[Import from Halloy]")
	h.do(click(x+1, y, tea.MouseLeft))
	if _, ok := h.nets.cfg.Servers["znc-libera"]; !ok {
		t.Fatalf("not imported:\n%s", h.view())
	}
	if strings.Join(h.nets.connects, ",") != "znc,znc-libera" {
		t.Fatalf("connects %v", h.nets.connects)
	}
	v := h.view()
	for _, want := range []string{"Imported from Halloy: znc, znc-libera", "already present: " + netName, "password_command"} {
		if !strings.Contains(v, want) {
			t.Fatalf("missing %q:\n%s", want, v)
		}
	}
}

func TestImportHalloyUnavailable(t *testing.T) {
	h := newHarness(t)
	h.m.runMenu(GlobalMenu(), ActServers)
	if strings.Contains(h.view(), "Halloy") {
		t.Fatal("button shown without a Halloy config")
	}
	h.do(key('i'))
	if !strings.Contains(h.view(), "No Halloy config found") {
		t.Fatalf("no explanation:\n%s", h.view())
	}
}

func TestStartupHints(t *testing.T) {
	h := newHarness(t)
	h.m.addHints([]string{"Found 2 servers in your Halloy config"})
	if !strings.Contains(h.view(), "Found 2 servers") {
		t.Fatal("hint not shown")
	}
}

func TestFirstRunWithHalloyOpensServers(t *testing.T) {
	cfg := &config.Config{Servers: map[string]config.Server{}, Managed: map[string]bool{}}
	nets := &fakeNets{senders: map[string]*fakeSender{}, cfg: cfg, adhoc: map[string]bool{}, none: true}
	m := New(Deps{
		Cfg: cfg, Nets: nets, Events: make(chan irc.Event),
		Halloy: func() (map[string]config.Server, []string, error) { return nil, nil, nil },
		Hints:  []string{"Found 2 new server(s) in your Halloy config"},
	})
	m.Update(tea.WindowSizeMsg{Width: 100, Height: 24})
	if m.admin == nil {
		t.Fatal("first run with Halloy should open the Servers screen")
	}
	v := ansi.Strip(m.View().Content)
	for _, bad := range []string{"▾ \n", " — status", " : disconnected"} {
		if strings.Contains(v, bad) {
			t.Fatalf("empty network leaks %q:\n%s", bad, v)
		}
	}
	if !strings.Contains(v, "[Import from Halloy]") || !strings.Contains(v, "Found your Halloy config") {
		t.Fatalf("screen:\n%s", v)
	}
}
