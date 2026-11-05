package ui

import (
	"sort"

	"github.com/eddyv73/wisirc/internal/config"
	"github.com/eddyv73/wisirc/internal/i18n"
)

// A disabled server does not connect when wisirc starts; the list is kept in
// prefs.toml. Connecting one by hand enables it again.

// toggleServer disables a server (stopping it) or enables and connects it,
// and returns the notice to show.
func (m *Model) toggleServer(name string) string {
	if cfg := m.d.Nets.Config(); cfg == nil || !hasServer(cfg.Servers, name) {
		return i18n.T("%s is not saved: save it first", name)
	}
	if m.disabled[name] {
		m.startServer(name)
		return i18n.T("%s enabled: connecting…", name)
	}
	m.disabled[name] = true
	m.saveDisabled()
	if m.running(name) {
		m.d.Nets.Disconnect(name)
	}
	return i18n.T("%s disabled: it won't connect at startup", name)
}

// startServer connects a configured server without leaving the current view.
func (m *Model) startServer(name string) {
	m.enable(name)
	if _, err := m.d.Nets.Connect(name); err == nil {
		m.st.AddNetwork(name)
	}
}

// enable clears a server's disabled mark, if it has one.
func (m *Model) enable(name string) {
	if m.disabled[name] {
		delete(m.disabled, name)
		m.saveDisabled()
	}
}

func (m *Model) saveDisabled() {
	if m.d.SaveDisabled == nil {
		return
	}
	names := make([]string, 0, len(m.disabled))
	for n := range m.disabled {
		names = append(names, n)
	}
	sort.Strings(names)
	if err := m.d.SaveDisabled(names); err != nil {
		m.errorLine(i18n.T("Could not save the disabled servers: %s", err.Error()))
	}
}

// editServer opens the Servers screen with name's form.
func (m *Model) editServer(name string) {
	m.dash = nil
	m.showServers()
	for i, e := range m.adminEntries() {
		if e.Name == name {
			m.admin.sel = i
			m.adminAction("edit")
			return
		}
	}
}

func hasServer(servers map[string]config.Server, name string) bool {
	_, ok := servers[name]
	return ok
}
