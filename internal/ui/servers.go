package ui

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/eddyv73/wisirc/internal/config"
	"github.com/eddyv73/wisirc/internal/i18n"
	"github.com/eddyv73/wisirc/internal/state"
)

// adminScreen is the "Servers" panel: a list of servers and, while
// creating or editing one, its form. It is drawn in the chat panel.
type adminScreen struct {
	sel  int
	form *serverForm
	msg  string
	err  bool
	hits []adminHit // clickable spots from the last render, panel-relative
}

type adminHit struct {
	y, x0, x1 int
	act       string // "row", "field", "new", "edit", "toggle", "delete", "save", "close", "submit", "cancel"
	index     int
}

type adminEntry struct {
	Name   string
	Srv    config.Server
	Origin string // "config.toml", "servers.toml" or originUnsaved
}

// originUnsaved marks a running /connect connection that is not saved.
const originUnsaved = "unsaved"

// Form field indexes.
const (
	fName = iota
	fHost
	fPort
	fTLS
	fInsecure
	fNick
	fUser
	fPass
	fSASLUser
	fSASLPass
	fChannels
	fAuto
	fCount
)

type formField struct {
	toggle bool
	on     bool
	input  textinput.Model
}

type serverForm struct {
	old    string // name being edited; "" creates a new server
	fields []formField
	focus  int // fCount = Save, fCount+1 = Cancel
}

func newServerForm(old string, s config.Server) *serverForm {
	f := &serverForm{old: old}
	text := func(value string, secret bool) formField {
		ti := textinput.New()
		ti.Prompt = ""
		ti.CharLimit = 256
		ti.SetValue(value)
		if secret {
			ti.EchoMode = textinput.EchoPassword
		}
		return formField{input: ti}
	}
	toggle := func(on bool) formField { return formField{toggle: true, on: on} }
	port := ""
	if s.Port != 0 {
		port = strconv.Itoa(s.Port)
	}
	var saslUser, saslPass string
	if s.SASL != nil && s.SASL.Plain != nil {
		saslUser, saslPass = s.SASL.Plain.Username, s.SASL.Plain.Password
	}
	f.fields = []formField{
		fName:     text(old, false),
		fHost:     text(s.Server, false),
		fPort:     text(port, false),
		fTLS:      toggle(s.TLS()),
		fInsecure: toggle(s.DangerouslyAcceptInvalidCerts),
		fNick:     text(s.Nickname, false),
		fUser:     text(s.Username, false),
		fPass:     text(s.Password, true),
		fSASLUser: text(saslUser, false),
		fSASLPass: text(saslPass, true),
		fChannels: text(strings.Join(s.Channels, ","), false),
		fAuto:     toggle(s.Auto()),
	}
	f.setFocus(0)
	return f
}

// fieldText returns the label and hint of form field i in the active language.
func fieldText(i int) (label, hint string) {
	switch i {
	case fName:
		return i18n.T("Name"), i18n.T("as shown in the sidebar")
	case fHost:
		return i18n.T("Server"), i18n.T("host or IP")
	case fPort:
		return i18n.T("Port"), i18n.T("empty = 6697")
	case fTLS:
		return "TLS", i18n.T("encryption; usually on 6697")
	case fInsecure:
		return i18n.T("Accept invalid cert."), i18n.T("only for self-signed certificates")
	case fNick:
		return "Nick", ""
	case fUser:
		return i18n.T("User"), i18n.T("on a BNC/ZNC: user/network")
	case fPass:
		return i18n.T("Server/BNC password"), i18n.T("PASS; on ZNC the user's password")
	case fSASLUser:
		return i18n.T("SASL user"), i18n.T("NickServ login")
	case fSASLPass:
		return i18n.T("SASL password"), ""
	case fChannels:
		return i18n.T("Channels"), "#a,#b"
	case fAuto:
		return i18n.T("Connect on startup"), ""
	}
	return "", ""
}

func (f *serverForm) setFocus(i int) {
	n := fCount + 2
	f.focus = (i%n + n) % n
	for j := range f.fields {
		if j == f.focus && !f.fields[j].toggle {
			f.fields[j].input.Focus()
		} else {
			f.fields[j].input.Blur()
		}
	}
}

func (f *serverForm) value(i int) string { return strings.TrimSpace(f.fields[i].input.Value()) }

// server builds the config entry, or explains what is missing.
func (f *serverForm) server() (string, config.Server, error) {
	name := f.value(fName)
	var s config.Server
	switch {
	case name == "":
		return "", s, errors.New(i18n.T("The name is missing"))
	case strings.ContainsAny(name, " \t"):
		return "", s, errors.New(i18n.T("The name cannot contain spaces"))
	case f.value(fHost) == "":
		return "", s, errors.New(i18n.T("The server is missing (host or IP)"))
	}
	s.Server = f.value(fHost)
	s.Port = 6697
	if p := f.value(fPort); p != "" {
		n, err := strconv.Atoi(p)
		if err != nil || n <= 0 || n > 65535 {
			return "", s, errors.New(i18n.T("Invalid port: %s", p))
		}
		s.Port = n
	}
	tlsOn := f.fields[fTLS].on
	s.UseTLS = &tlsOn
	s.DangerouslyAcceptInvalidCerts = f.fields[fInsecure].on
	s.Nickname = f.value(fNick)
	s.Username = f.value(fUser)
	s.Password = f.fields[fPass].input.Value()
	if u := f.value(fSASLUser); u != "" {
		s.SASL = &config.SASL{Plain: &config.SASLPlain{Username: u, Password: f.fields[fSASLPass].input.Value()}}
	}
	for _, ch := range strings.Split(f.value(fChannels), ",") {
		if ch = strings.TrimSpace(ch); ch != "" {
			s.Channels = append(s.Channels, ch)
		}
	}
	if !f.fields[fAuto].on {
		off := false
		s.AutoConnect = &off
	}
	return name, s, nil
}

// adminEntries lists configured servers plus running unsaved connections.
func (m *Model) adminEntries() []adminEntry {
	cfg := m.d.Nets.Config()
	var out []adminEntry
	if cfg != nil {
		for name, s := range cfg.Servers {
			origin := "config.toml"
			if cfg.Managed[name] {
				origin = "servers.toml"
			}
			out = append(out, adminEntry{Name: name, Srv: s, Origin: origin})
		}
	}
	for _, name := range m.d.Nets.Names() {
		if m.d.Nets.IsAdHoc(name) {
			out = append(out, adminEntry{Name: name, Origin: originUnsaved})
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

func (m *Model) showServers() {
	m.overlay, m.menu = nil, nil
	m.admin = &adminScreen{}
}

func (m *Model) adminSelected() (adminEntry, bool) {
	entries := m.adminEntries()
	if len(entries) == 0 {
		return adminEntry{}, false
	}
	m.admin.sel = min(max(m.admin.sel, 0), len(entries)-1)
	return entries[m.admin.sel], true
}

func (m *Model) adminNote(text string, isErr bool) {
	m.admin.msg, m.admin.err = text, isErr
}

func (m *Model) netStatus(name string) string {
	n := m.st.Network(name)
	switch {
	case n == nil:
		return "○"
	case n.Connected:
		return "●"
	case n.Connecting:
		return "◌"
	}
	return "○"
}

func (m *Model) running(name string) bool {
	_, ok := m.d.Nets.Get(name)
	return ok
}

// adminAction runs a list-view action on the selected server.
func (m *Model) adminAction(act string) tea.Cmd {
	if act == "new" {
		nick := ""
		if cfg := m.d.Nets.Config(); cfg != nil {
			for _, s := range cfg.Servers {
				nick = s.Nickname
				break
			}
		}
		m.admin.form = newServerForm("", config.Server{Nickname: nick, Port: 6697})
		m.admin.msg = ""
		return nil
	}
	if act == "close" {
		m.admin = nil
		return nil
	}
	if act == "halloy" {
		m.importHalloy()
		return nil
	}
	e, ok := m.adminSelected()
	if !ok {
		return nil
	}
	ref := state.BufRef{Network: e.Name}
	switch act {
	case "edit":
		switch e.Origin {
		case "config.toml":
			m.adminNote(i18n.T("%s is in config.toml: opening the editor", e.Name), false)
			return m.editConfig()
		case originUnsaved:
			m.adminNote(i18n.T("Use [Save] to save %s", e.Name), false)
		default:
			m.admin.form = newServerForm(e.Name, e.Srv)
			m.admin.msg = ""
		}
	case "toggle":
		if m.running(e.Name) {
			m.d.Nets.Disconnect(e.Name)
			m.adminNote(i18n.T("Disconnected from %s", e.Name), false)
		} else {
			m.connect(e.Name)
			m.adminNote(i18n.T("Connecting to %s…", e.Name), false)
		}
	case "save":
		if e.Origin != originUnsaved {
			return nil
		}
		if err := m.d.Nets.SaveAdHoc(e.Name); err != nil {
			m.adminNote(err.Error(), true)
		} else {
			m.adminNote(i18n.T("%s saved to servers.toml", e.Name), false)
		}
	case "delete":
		switch e.Origin {
		case "config.toml":
			m.adminNote(i18n.T("%s is in config.toml: delete it with ≡ → Edit config", e.Name), true)
		case originUnsaved:
			m.d.Nets.Disconnect(e.Name)
			m.adminNote(i18n.T("Disconnected from %s", e.Name), false)
		default:
			menu := ConfirmMenu(i18n.T("Delete %s from servers.toml?", e.Name),
				MenuItem{i18n.T("Yes, delete  (Y)"), ActDelServer}, MenuItem{i18n.T("Cancel  (N)"), ActDismiss}, ref, m.w, m.h)
			m.menu = &menu
		}
	}
	return nil
}

func (m *Model) deleteServer(name string) {
	if err := m.d.Nets.DeleteServer(name); err != nil {
		m.errorLine(err.Error())
		if m.admin != nil {
			m.adminNote(err.Error(), true)
		}
		return
	}
	if m.admin != nil {
		m.adminNote(i18n.T("%s deleted", name), false)
	}
}

func (m *Model) submitForm() {
	f := m.admin.form
	name, s, err := f.server()
	if err == nil {
		err = m.d.Nets.SaveServer(f.old, name, s)
	}
	if err != nil {
		m.adminNote(err.Error(), true)
		return
	}
	note := i18n.T("%s saved to servers.toml", name)
	if m.running(name) || (f.old != "" && m.running(f.old)) {
		note += i18n.T(" · applies on reconnect")
	}
	m.admin.form = nil
	m.adminNote(note, false)
	for i, e := range m.adminEntries() {
		if e.Name == name {
			m.admin.sel = i
		}
	}
}

func (m *Model) adminKey(msg tea.KeyPressMsg) tea.Cmd {
	key := msg.String()
	if key == "ctrl+c" {
		return m.quit("")
	}
	if f := m.admin.form; f != nil {
		switch key {
		case "esc":
			m.admin.form = nil
			m.admin.msg = ""
		case "tab", "down":
			f.setFocus(f.focus + 1)
		case "shift+tab", "up":
			f.setFocus(f.focus - 1)
		case "enter":
			if f.focus == fCount+1 {
				m.admin.form = nil
				return nil
			}
			m.submitForm()
		case "space", " ":
			if f.focus < fCount && f.fields[f.focus].toggle {
				f.fields[f.focus].on = !f.fields[f.focus].on
				return nil
			}
			return m.formInput(msg)
		default:
			return m.formInput(msg)
		}
		return nil
	}
	switch key {
	case "esc", "q":
		m.admin = nil
	case "up", "k":
		m.admin.sel--
		m.adminSelected()
	case "down", "j":
		m.admin.sel++
		m.adminSelected()
	case "n", "+":
		return m.adminAction("new")
	case "enter", "e":
		return m.adminAction("edit")
	case "c":
		return m.adminAction("toggle")
	case "d", "delete", "backspace":
		return m.adminAction("delete")
	case "g", "s":
		return m.adminAction("save")
	case "i":
		return m.adminAction("halloy")
	}
	return nil
}

// formInput passes a message to the focused text field.
func (m *Model) formInput(msg tea.Msg) tea.Cmd {
	f := m.admin.form
	if f == nil || f.focus >= fCount || f.fields[f.focus].toggle {
		return nil
	}
	var cmd tea.Cmd
	f.fields[f.focus].input, cmd = f.fields[f.focus].input.Update(msg)
	return cmd
}

func (m *Model) adminClick(x, y int) tea.Cmd {
	c := m.Layout().Chat
	px, py := x-c.X, y-c.Y
	for _, h := range m.admin.hits {
		if h.y != py || px < h.x0 || px >= h.x1 {
			continue
		}
		f := m.admin.form
		switch h.act {
		case "row":
			m.admin.sel = h.index
		case "field":
			f.setFocus(h.index)
			if f.fields[h.index].toggle {
				f.fields[h.index].on = !f.fields[h.index].on
			}
		case "submit":
			m.submitForm()
		case "cancel":
			m.admin.form = nil
			m.admin.msg = ""
		default:
			return m.adminAction(h.act)
		}
		return nil
	}
	return nil
}

// buttons lays out [label] buttons on one row and records their hits.
func (a *adminScreen) buttons(y int, items [][2]string) string {
	var b strings.Builder
	x := 1
	b.WriteString(" ")
	for i, it := range items {
		if i > 0 {
			b.WriteString("  ")
			x += 2
		}
		label := "[" + it[0] + "]"
		w := ansi.StringWidth(label)
		a.hits = append(a.hits, adminHit{y: y, x0: x, x1: x + w, act: it[1]})
		b.WriteString(adminButton.Render(label))
		x += w
	}
	return b.String()
}

// renderAdmin draws the Servers screen into the chat panel.
func (m *Model) renderAdmin(w, h int) []string {
	a := m.admin
	a.hits = nil
	var lines []string
	add := func(s string) { lines = append(lines, s) }
	if f := a.form; f != nil {
		title := i18n.T("New server")
		if f.old != "" {
			title = i18n.T("Edit %s", f.old)
		}
		add(" " + lipgloss.NewStyle().Bold(true).Render(title) + adminDim.Render(i18n.T("  ·  Tab/↓ next · Space toggles · Enter saves · Esc goes back")))
		add("")
		const labelW = 24
		inputW := max(min(w-labelW-32, 32), 8)
		for i := range f.fields {
			fl := &f.fields[i]
			flLabel, flHint := fieldText(i)
			marker := "  "
			if i == f.focus {
				marker = "▸ "
			}
			var val string
			if fl.toggle {
				box := "[ ]"
				if fl.on {
					box = "[x]"
				}
				val = box
			} else {
				fl.input.SetWidth(inputW)
				val = "▏" + fl.input.View()
				if pad := inputW + 1 - ansi.StringWidth(val); pad > 0 {
					val += strings.Repeat(" ", pad)
				}
			}
			label := flLabel + strings.Repeat(" ", max(labelW-ansi.StringWidth(flLabel), 1))
			row := " " + marker + label + val
			if flHint != "" {
				row += "  " + adminDim.Render(flHint)
			}
			a.hits = append(a.hits, adminHit{y: len(lines), x0: 0, x1: w, act: "field", index: i})
			add(row)
		}
		add("")
		save, cancel := i18n.T("Save"), i18n.T("Cancel")
		if f.focus == fCount {
			save = "▸ " + save
		}
		if f.focus == fCount+1 {
			cancel = "▸ " + cancel
		}
		add(a.buttons(len(lines), [][2]string{{save, "submit"}, {cancel, "cancel"}}))
	} else {
		add(adminDim.Render(i18n.T(" ↑/↓ choose · Enter edit · n new · c connect · d delete · Esc close")))
		add("")
		entries := m.adminEntries()
		if len(entries) == 0 {
			add(adminDim.Render(i18n.T(" No servers. Create one with [+ New].")))
		}
		if a.sel >= len(entries) {
			a.sel = max(len(entries)-1, 0)
		}
		nameW, addrW := 6, 6
		for _, e := range entries {
			nameW = max(nameW, ansi.StringWidth(e.Name))
			if e.Origin != originUnsaved {
				addrW = max(addrW, ansi.StringWidth(e.Srv.Addr()))
			}
		}
		for i, e := range entries {
			addr := ""
			var flags []string
			if e.Origin != originUnsaved {
				addr = e.Srv.Addr()
				if e.Srv.TLS() {
					flags = append(flags, "TLS")
				}
				if e.Srv.SASL != nil && e.Srv.SASL.Plain != nil {
					flags = append(flags, "SASL")
				}
				if e.Srv.Password != "" {
					flags = append(flags, "PASS")
				}
				if !e.Srv.Auto() {
					flags = append(flags, "manual")
				}
			}
			origin := e.Origin
			if origin == originUnsaved {
				origin = i18n.T("unsaved")
			}
			row := fmt.Sprintf(" %s %-*s  %-*s  %-15s %s", m.netStatus(e.Name), nameW, sanitize(e.Name), addrW, sanitize(addr), strings.Join(flags, " "), origin)
			row = fit(row, w)
			if i == a.sel {
				row = adminSel.Render(row)
			}
			a.hits = append(a.hits, adminHit{y: len(lines), x0: 0, x1: w, act: "row", index: i})
			add(row)
		}
		add("")
		toggle := i18n.T("Connect")
		var extra [][2]string
		if e, ok := m.adminSelected(); ok {
			if m.running(e.Name) {
				toggle = i18n.T("Disconnect")
			}
			if e.Origin == originUnsaved {
				extra = append(extra, [2]string{i18n.T("Save"), "save"})
			}
		}
		btns := append([][2]string{{i18n.T("+ New"), "new"}, {i18n.T("Edit"), "edit"}, {toggle, "toggle"}}, extra...)
		btns = append(btns, [2]string{i18n.T("Delete"), "delete"}, [2]string{i18n.T("Close"), "close"})
		add(a.buttons(len(lines), btns))
		if m.d.Halloy != nil {
			add(a.buttons(len(lines), [][2]string{{i18n.T("Import from Halloy"), "halloy"}}) + adminDim.Render(i18n.T("  brings your Halloy servers over and connects them")))
		}
		add("")
		add(adminDim.Render(i18n.T(" ● connected  ◌ connecting  ○ disconnected")))
		add(adminDim.Render(i18n.T(" servers.toml is written by wisirc · config.toml is edited by hand")))
	}
	if a.msg != "" {
		add("")
		st := adminOK
		if a.err {
			st = adminErr
		}
		for _, line := range strings.Split(a.msg, "\n") {
			add(" " + st.Render(sanitize(line)))
		}
	}
	out := make([]string, h)
	for i := range out {
		if i < len(lines) {
			out[i] = fit(lines[i], w)
		} else {
			out[i] = strings.Repeat(" ", max(w, 0))
		}
	}
	return out
}

// importHalloy copies the Halloy servers wisirc does not have yet and
// connects them: the one-click migration.
func (m *Model) importHalloy() {
	if m.d.Halloy == nil {
		m.adminNote(i18n.T("No Halloy config found on this computer"), true)
		return
	}
	servers, warns, err := m.d.Halloy()
	if err != nil {
		m.adminNote("Halloy: "+err.Error(), true)
		return
	}
	imported, skipped, err := m.d.Nets.Import(servers)
	if err != nil {
		m.adminNote(err.Error(), true)
		return
	}
	var msg []string
	if len(imported) > 0 {
		msg = append(msg, i18n.T("Imported from Halloy: %s", strings.Join(imported, ", ")))
	} else {
		msg = append(msg, i18n.T("No new servers in Halloy"))
	}
	if len(skipped) > 0 {
		msg[0] += i18n.T(" · already present: %s", strings.Join(skipped, ", "))
	}
	for _, w := range warns {
		msg = append(msg, "⚠ "+w)
	}
	for _, name := range imported {
		if s := m.d.Nets.Config().Servers[name]; s.Auto() {
			if _, err := m.d.Nets.Connect(name); err == nil {
				m.st.AddNetwork(name)
			}
		}
	}
	m.adminNote(strings.Join(msg, "\n"), false)
}
