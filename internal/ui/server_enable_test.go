package ui

import (
	"strings"
	"testing"

	"github.com/eddyv73/wisirc/internal/config"
	"github.com/eddyv73/wisirc/internal/state"
)

func (h *harness) savedDisabled() *[]string {
	var saved []string
	h.m.d.SaveDisabled = func(names []string) error {
		saved = append([]string(nil), names...)
		return nil
	}
	return &saved
}

func TestEditConfigServerOpensForm(t *testing.T) {
	h := newHarness(t) // examplenet lives in config.toml
	h.m.showServers()
	h.do(key('e'))
	if h.m.admin.form == nil || h.m.admin.form.old != netName {
		t.Fatalf("no form for a config.toml server:\n%s", h.view())
	}
	if !strings.Contains(h.view(), "servers.toml") {
		t.Fatalf("form does not say where the edit goes:\n%s", h.view())
	}
}

func TestDisableAndEnableFromServers(t *testing.T) {
	h := newHarness(t)
	saved := h.savedDisabled()
	h.m.showServers()
	h.do(key('x'))
	if !h.m.disabled[netName] || len(*saved) != 1 || (*saved)[0] != netName {
		t.Fatalf("not disabled: %v saved %v", h.m.disabled, *saved)
	}
	if len(h.nets.discons) != 1 {
		t.Fatalf("running server not disconnected: %v", h.nets.discons)
	}
	if v := h.view(); !strings.Contains(v, "⊘") || !strings.Contains(v, "[Enable]") {
		t.Fatalf("no disabled mark or [Enable]:\n%s", v)
	}
	h.do(key('x'))
	if h.m.disabled[netName] || len(*saved) != 0 {
		t.Fatalf("not enabled: %v saved %v", h.m.disabled, *saved)
	}
	if len(h.nets.connects) != 1 || h.nets.connects[0] != netName {
		t.Fatalf("enabling did not connect: %v", h.nets.connects)
	}
}

func TestConnectEnablesDisabledServer(t *testing.T) {
	h := newHarness(t)
	saved := h.savedDisabled()
	h.m.disabled[netName] = true
	h.m.connect(netName)
	if h.m.disabled[netName] || len(*saved) != 0 {
		t.Fatalf("connect kept it disabled: %v", h.m.disabled)
	}
}

func TestServerMenuEditAndDisable(t *testing.T) {
	h := newHarness(t)
	h.savedDisabled()
	ref := state.BufRef{Network: netName}
	menu := ServerMenu(ref, true, false, 0, 0)
	if !hasItem(menu, ActEditServer, "Edit") {
		t.Fatalf("no Edit… in %+v", menu.Items)
	}
	if got := ServerMenu(ref, false, true, 0, 0); !hasItem(got, ActToggleServer, "Enable") {
		t.Fatalf("disabled server menu: %+v", got.Items)
	}
	if !hasItem(menu, ActToggleServer, "Disable") {
		t.Fatalf("enabled server menu: %+v", menu.Items)
	}
	h.m.toggleDashboard()
	h.m.runMenu(menu, ActEditServer)
	if h.m.dash != nil || h.m.admin == nil || h.m.admin.form == nil || h.m.admin.form.old != netName {
		t.Fatal("Edit… did not open the form")
	}
}

func hasItem(m Menu, act MenuAction, label string) bool {
	for _, it := range m.Items {
		if it.Action == act && strings.Contains(it.Label, label) {
			return true
		}
	}
	return false
}

func TestDashboardDisableAndEdit(t *testing.T) {
	h := newHarness(t)
	h.savedDisabled()
	h.m.toggleDashboard()
	h.do(key('x'))
	if !h.m.disabled[netName] || !strings.Contains(h.view(), "disabled") {
		t.Fatalf("dashboard x did not disable:\n%s", h.view())
	}
	h.do(key('e'))
	if h.m.dash != nil || h.m.admin == nil || h.m.admin.form == nil {
		t.Fatal("dashboard e did not open the form")
	}
}

func TestRevertEditedConfigServer(t *testing.T) {
	h := newHarness(t)
	h.nets.cfg.Managed[netName] = true
	h.nets.cfg.Shadowed = map[string]config.Server{netName: h.nets.cfg.Servers[netName]}
	h.m.showServers()
	if !strings.Contains(h.view(), "✎") {
		t.Fatalf("edited server not marked:\n%s", h.view())
	}
	h.do(key('d'))
	if h.m.menu == nil || !strings.Contains(h.view(), "config.toml") {
		t.Fatalf("no revert prompt:\n%s", h.view())
	}
}
