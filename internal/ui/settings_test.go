package ui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/eddyv73/wisirc/internal/state"
)

// These tests edit config.toml the way ≡ → Edit config does (write the
// file, then the editor exits) and check that each setting takes effect
// without restarting wisirc.

func (h *harness) editConfig(body string) {
	h.t.Helper()
	if h.m.d.CfgPath == "" {
		h.m.d.CfgPath = filepath.Join(h.t.TempDir(), "config.toml")
	}
	if err := os.WriteFile(h.m.d.CfgPath, []byte(body), 0o600); err != nil {
		h.t.Fatal(err)
	}
	h.send(configEditedMsg{})
}

const lobbyServer = `
[servers.` + netName + `]
  nickname = "alice"
  server = "irc.example.net"
`

func (h *harness) netLines(name string) string {
	var b strings.Builder
	for _, l := range h.m.st.Buffer(state.BufRef{Network: name}).Lines {
		b.WriteString(l.Text + "\n")
	}
	return b.String()
}

func TestReloadReportsWhatChanged(t *testing.T) {
	h := newHarness(t)
	h.nets.reloaded = Reloaded{Started: []string{"nueva"}, Restarted: []string{"znc"}, Stopped: []string{"vieja"}}
	h.editConfig(lobbyServer + `
[servers.nueva]
  nickname = "alice"
  server = "nueva.example"
[servers.znc]
  nickname = "alice"
  server = "192.0.2.9"
  dangerously_accept_invalid_certs = true
`)
	all := h.view() + h.netLines(h.m.st.Active.Network)
	for _, want := range []string{"Connecting to nueva", "Reconnecting to znc with the new settings", "vieja is no longer in the config", "Config reloaded"} {
		if !strings.Contains(all, want) {
			t.Errorf("missing %q in:\n%s", want, all)
		}
	}
	if !strings.Contains(h.netLines("znc"), "certificate verification disabled") {
		t.Errorf("no insecure warning for znc: %q", h.netLines("znc"))
	}
	if h.m.d.Cfg.Servers["znc"].Server != "192.0.2.9" {
		t.Errorf("config not swapped: %+v", h.m.d.Cfg.Servers)
	}
}

func TestReloadKeepsDisabledServersOff(t *testing.T) {
	h := newHarness(t)
	h.m.disabled["apagado"] = true
	h.editConfig(lobbyServer)
	if h.nets.skip == nil || !h.nets.skip("apagado") || h.nets.skip(netName) {
		t.Fatal("Reload not told which servers are disabled")
	}
}

func TestReloadAppliesThemeUnlessChosenInMenu(t *testing.T) {
	h := newHarness(t)
	h.editConfig(lobbyServer + "[theme]\n  name = \"papel\"\n")
	if h.m.theme.name != "papel" {
		t.Fatalf("theme %q after editing [theme]", h.m.theme.name)
	}
	h.m.chooseTheme("ambar") // the menu choice lives in prefs.toml and wins
	h.editConfig(lobbyServer + "[theme]\n  name = \"fosforo\"\n")
	if h.m.theme.name != "ambar" {
		t.Fatalf("[theme] overrode the menu choice: %q", h.m.theme.name)
	}
}

func TestReloadNotificationsSwitch(t *testing.T) {
	h := newHarness(t)
	h.joinLobby()
	h.irc(":alice!k@h JOIN #otro")
	h.send(click(5, 2, 1)) // back to #lobby

	h.editConfig(lobbyServer + "[notifications]\n  enabled = false\n")
	h.irc(":ana!a@h PRIVMSG #otro :alice uno")
	if len(h.notes.titles) != 0 {
		t.Fatalf("notified with notifications off: %v", h.notes.titles)
	}
	h.editConfig(lobbyServer + "[notifications]\n  enabled = true\n")
	h.irc(":ana!a@h PRIVMSG #otro :alice dos")
	if len(h.notes.titles) != 1 {
		t.Fatalf("not notified after turning them back on: %v", h.notes.titles)
	}
}

func TestReloadUpdatesSwitch(t *testing.T) {
	h := newHarness(t)
	h.withUpdater("v0.4.0", "v0.4.1")
	h.send(h.m.checkUpdate(false)())
	if !strings.Contains(h.statusRow(), "v0.4.1") {
		t.Fatalf("no notice: %q", h.statusRow())
	}
	h.editConfig(lobbyServer + "[updates]\n  check = false\n")
	if strings.Contains(h.statusRow(), "v0.4.1") || h.m.globalMenu().Items[0].Action == ActUpdate {
		t.Fatalf("still announcing after check = false: %q", h.statusRow())
	}
	if h.m.checkUpdate(false) != nil {
		t.Fatal("still checking with check = false")
	}
	h.editConfig(lobbyServer)
	if h.m.d.Updates == nil {
		t.Fatal("check not back on after removing check = false")
	}
}

func TestReloadBadConfigKeepsCurrent(t *testing.T) {
	h := newHarness(t)
	before := h.m.d.Cfg
	h.editConfig("[servers.roto\n")
	if h.m.d.Cfg != before {
		t.Fatal("a broken config replaced the working one")
	}
	if v := h.view(); !strings.Contains(v, "config.toml") && !strings.Contains(v, "line") {
		t.Fatalf("no error shown:\n%s", v)
	}
}
