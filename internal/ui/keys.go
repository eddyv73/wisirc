package ui

import (
	"sort"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/eddyv73/wisirc/internal/commands"
	"github.com/eddyv73/wisirc/internal/i18n"
	"github.com/eddyv73/wisirc/internal/irc"
	"github.com/eddyv73/wisirc/internal/state"
)

const maxHistory = 100

func (m *Model) handleKey(msg tea.KeyPressMsg) tea.Cmd {
	key := msg.String()
	m.flash, m.sel = "", nil
	if key != "tab" {
		m.comp = nil
	}
	if m.overlay != nil {
		return m.overlayKey(msg)
	}
	if key == "f2" && m.menu == nil {
		m.toggleDashboard()
		return nil
	}
	if m.menu != nil {
		if m.menu.Kind == MenuConfirm {
			switch key {
			case "s", "S", "y", "Y":
				m.menu.Selected = 0
				key = "enter"
			case "n", "N":
				m.menu.Selected = 1
				key = "enter"
			}
		}
		switch key {
		case "up", "shift+tab":
			m.menu.Move(-1)
			m.previewTheme()
			m.previewLang()
			m.previewColor()
		case "down", "tab":
			m.menu.Move(1)
			m.previewTheme()
			m.previewLang()
			m.previewColor()
		case "enter":
			menu := *m.menu
			m.menu = nil
			return m.runMenu(menu, menu.Items[menu.Selected].Action)
		case "esc", "f1":
			if m.menu.Kind == MenuTheme {
				m.revertTheme()
			}
			if m.menu.Kind == MenuLang {
				i18n.SetLang(m.langBefore)
			}
			if m.menu.Kind == MenuColor {
				m.revertColor(m.menu.Ref.Network)
			}
			m.menu = nil
		case "ctrl+c":
			return m.quit("")
		}
		return nil
	}
	if m.admin != nil {
		return m.adminKey(msg)
	}
	if m.dash != nil {
		return m.dashKey(msg)
	}

	switch key {
	case "ctrl+c":
		return m.quit("")
	case "f1":
		g := GlobalMenu()
		m.menu = &g
		return nil
	case "enter":
		return m.submit()
	case "tab":
		m.complete()
		return nil
	case "up":
		m.historyMove(-1)
		return nil
	case "down":
		m.historyMove(1)
		return nil
	case "pgup":
		m.scrollChat(max(m.Layout().Chat.H-1, 1))
		return nil
	case "pgdown":
		m.scrollChat(-max(m.Layout().Chat.H-1, 1))
		return nil
	case "ctrl+n", "alt+right":
		m.cycle(1)
		return nil
	case "ctrl+p", "alt+left":
		m.cycle(-1)
		return nil
	case "alt+a":
		m.nextActivity()
		return nil
	}
	if len(key) == 5 && strings.HasPrefix(key, "alt+") && key[4] >= '1' && key[4] <= '9' {
		if order := m.st.Order(); int(key[4]-'1') < len(order) {
			m.switchTo(order[key[4]-'1'])
		}
		return nil
	}
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return cmd
}

func (m *Model) overlayKey(msg tea.KeyPressMsg) tea.Cmd {
	key := msg.String()
	page := max(m.Layout().Chat.H-1, 1)
	if m.overlay.typing {
		switch key {
		case "esc":
			m.overlay = nil
			m.input.SetValue("")
			return nil
		case "enter":
			m.overlay = nil
			return m.submit()
		case "up", "down", "pgup", "pgdown", "ctrl+c":
		default:
			var cmd tea.Cmd
			m.input, cmd = m.input.Update(msg)
			return cmd
		}
	}
	switch key {
	case "esc", "q", "enter":
		m.overlay = nil
	case "up":
		m.overlay.scroll++
	case "down":
		m.overlay.scroll = max(m.overlay.scroll-1, 0)
	case "pgup":
		m.overlay.scroll += page
	case "pgdown":
		m.overlay.scroll = max(m.overlay.scroll-page, 0)
	case "ctrl+c":
		return m.quit("")
	}
	return nil
}

func (m *Model) cycle(delta int) {
	order := m.st.Order()
	if len(order) == 0 {
		return
	}
	cur := 0
	for i, r := range order {
		if refKey(r) == refKey(m.st.Active) {
			cur = i
		}
	}
	m.switchTo(order[((cur+delta)%len(order)+len(order))%len(order)])
}

func (m *Model) nextActivity() {
	order := m.st.Order()
	for _, want := range []func(*state.Buffer) bool{
		func(b *state.Buffer) bool { return b.Mentioned },
		func(b *state.Buffer) bool { return b.Unread > 0 },
	} {
		for _, r := range order {
			if b := m.st.Buffer(r); b != nil && refKey(r) != refKey(m.st.Active) && want(b) {
				m.switchTo(r)
				return
			}
		}
	}
}

func (m *Model) scrollChat(delta int) {
	key := refKey(m.st.Active)
	l := m.Layout()
	s := max(m.scroll[key]+delta, 0)
	if b := m.activeBuffer(); b != nil {
		_, s = tailRows(b.Lines, l.Chat.W, l.Chat.H, s)
	}
	m.scroll[key] = s
}

func (m *Model) historyMove(delta int) {
	if len(m.history) == 0 {
		return
	}
	m.histPos = min(max(m.histPos+delta, 0), len(m.history))
	if m.histPos == len(m.history) {
		m.input.SetValue("")
	} else {
		m.input.SetValue(m.history[m.histPos])
	}
	m.input.CursorEnd()
}

func (m *Model) complete() {
	if m.comp == nil {
		val := m.input.Value()
		before, word := "", val
		if i := strings.LastIndexByte(val, ' '); i >= 0 {
			before, word = val[:i+1], val[i+1:]
		}
		b := m.activeBuffer()
		if word == "" || b == nil {
			return
		}
		var cands []string
		for _, mem := range b.Members {
			if strings.HasPrefix(irc.Fold(mem.Nick), irc.Fold(word)) {
				cands = append(cands, mem.Nick)
			}
		}
		if len(cands) == 0 {
			return
		}
		sort.Slice(cands, func(i, j int) bool { return irc.Fold(cands[i]) < irc.Fold(cands[j]) })
		m.comp = &completion{before: before, cands: cands, idx: -1, start: before == ""}
	}
	m.comp.idx = (m.comp.idx + 1) % len(m.comp.cands)
	suffix := " "
	if m.comp.start {
		suffix = ": "
	}
	m.input.SetValue(m.comp.before + m.comp.cands[m.comp.idx] + suffix)
	m.input.CursorEnd()
}

func (m *Model) submit() tea.Cmd {
	text := m.input.Value()
	m.input.SetValue("")
	if strings.TrimSpace(text) == "" {
		return nil
	}
	m.history = append(m.history, text)
	if len(m.history) > maxHistory {
		m.history = m.history[len(m.history)-maxHistory:]
	}
	m.histPos = len(m.history)
	m.scroll[refKey(m.st.Active)] = 0

	b := m.activeBuffer()
	ctx := commands.Context{}
	if b != nil {
		ctx.Target, ctx.Kind = b.Ref.Name, b.Kind
	}
	if n := m.st.Network(m.st.Active.Network); n != nil {
		ctx.Nick = n.Nick
	}
	a := commands.Parse(text, ctx)
	switch a.Kind {
	case commands.Error:
		m.errorLine(a.Err)
	case commands.Send:
		m.sendAction(a)
	case commands.Query:
		ref := state.BufRef{Network: m.st.Active.Network, Name: a.Target}
		m.ensureBuffer(ref, state.KindQuery)
		m.switchTo(ref)
		m.sendAction(a)
	case commands.Close:
		m.closeBuffer(m.st.Active)
	case commands.Connect:
		if a.Arg == "" {
			m.showConnect()
		} else {
			m.connect(a.Arg)
		}
	case commands.Help:
		m.showHelp()
	case commands.Dashboard:
		if m.dash == nil {
			m.toggleDashboard()
		}
	case commands.Lang:
		if a.Arg == "" {
			m.openLangMenu()
		} else {
			m.chooseLang(a.Arg)
		}
	case commands.Theme:
		if a.Arg == "" {
			m.openThemeMenu()
		} else {
			m.chooseTheme(a.Arg)
		}
	case commands.Quit:
		return m.quit(a.Arg)
	}
	return nil
}

// sendAction sends raw lines on the active network and echoes own messages.
func (m *Model) sendAction(a commands.Action) {
	if len(a.Lines) == 0 {
		return
	}
	s, ok := m.d.Nets.Get(m.st.Active.Network)
	if !ok {
		m.errorLine(i18n.T("not connected"))
		return
	}
	for _, line := range a.Lines {
		if err := s.Send(line); err != nil {
			m.errorLine(i18n.T("could not send: %s", err.Error()))
			return
		}
	}
	if len(a.Echo) == 0 {
		return
	}
	kind := state.KindQuery
	if strings.ContainsRune("#&+!", rune(a.Target[0])) {
		kind = state.KindChannel
	}
	ref := m.ensureBuffer(state.BufRef{Network: m.st.Active.Network, Name: a.Target}, kind).Ref
	nick := s.CurrentNick()
	lk := state.LineMsg
	if a.IsAction {
		lk = state.LineAction
	}
	for _, e := range a.Echo {
		line := state.Line{Nick: nick, Text: e, Kind: lk, Self: true}
		m.st.AddLine(ref, line)
		if m.d.Logger != nil {
			b := m.st.Buffer(ref)
			m.d.Logger.Append(ref.Network, ref.Name, b.Lines[len(b.Lines)-1])
		}
	}
}

func (m *Model) quit(msg string) tea.Cmd {
	if msg == "" {
		msg = "wisirc"
	}
	for _, n := range m.d.Nets.Names() {
		if s, ok := m.d.Nets.Get(n); ok {
			s.Send("QUIT :" + msg)
		}
	}
	return tea.Quit
}

func (m *Model) closeBuffer(ref state.BufRef) {
	b := m.st.Buffer(ref)
	if b == nil || b.Kind == state.KindStatus {
		m.errorLine(i18n.T("The status buffer cannot be closed"))
		return
	}
	if b.Kind == state.KindChannel && b.Joined {
		if s, ok := m.d.Nets.Get(ref.Network); ok {
			s.Send(irc.Encode("PART", b.Ref.Name))
		}
	}
	m.st.RemoveBuffer(ref)
	delete(m.seen, refKey(ref))
}

func (m *Model) connect(arg string) {
	name, err := m.d.Nets.Connect(arg)
	if err != nil {
		m.errorLine(i18n.T("%s — type /connect to see examples", err.Error()))
		return
	}
	m.st.AddNetwork(name)
	m.switchTo(state.BufRef{Network: name})
}

// helpText is built on each call so it follows the active language.
func helpText() []string {
	return []string{
		i18n.T("wisirc — help"),
		"",
		i18n.T("Keyboard"),
		i18n.T("  F1                 main menu (≡)"),
		i18n.T("  Alt+1..9           go to buffer N"),
		i18n.T("  Ctrl+N / Ctrl+P    next / previous (also Alt+→ / Alt+←)"),
		i18n.T("  Alt+A              next buffer with activity (mentions first)"),
		i18n.T("  PgUp / PgDn        scroll the history"),
		i18n.T("  ↑ / ↓              history of what you typed"),
		i18n.T("  Tab                complete nick"),
		i18n.T("  Esc                close menu or panel"),
		"",
		"Mouse",
		i18n.T("  Click on ≡         main menu"),
		i18n.T("  Click on the bar   switch channel; ▾/▸ collapses a server"),
		i18n.T("  Right click        context menu (server, channel, nick)"),
		i18n.T("  Click on a nick    nick menu"),
		i18n.T("  Wheel              scroll history and lists"),
		i18n.T("  Drag borders       resize the panels"),
		"",
		i18n.T("Commands"),
		"  /join /part /msg /query /me /nick /whois /topic /kick /ban",
		"  /mode /quit /connect /close /quote /help",
	}
}

func (m *Model) showHelp() {
	m.overlay = &textOverlay{title: i18n.T("Help / shortcuts (Esc to close)"), lines: helpText()}
}

// connectText is built on each call so it follows the active language.
func connectText() []string {
	return []string{
		i18n.T("Type below and press Enter · Esc cancels · ↑/↓ scrolls this help"),
		"",
		i18n.T("Examples (sample data: replace it with yours)"),
		i18n.T("  BNC (ZNC) on your network, no TLS — user is user/network:"),
		i18n.T("    /connect 192.0.2.20 7000 -notls -nick bob -user bob/libera -pass myZncPass"),
		i18n.T("  BNC with TLS and a self-signed certificate:"),
		i18n.T("    /connect bnc.mydomain.net 6697 -insecure -user bob/oftc -pass myZncPass -name znc-oftc"),
		i18n.T("  Server with a password, unencrypted:"),
		i18n.T("    /connect irc.example.org 6667 -nick bob -pass serverPassword"),
		i18n.T("  Public network with SASL login (NickServ):"),
		i18n.T("    /connect irc.libera.chat -nick bob -sasl bob:myPass123"),
		i18n.T("  Public network, joining channels:"),
		i18n.T("    /connect irc.libera.chat -nick bob -join #linux,#go-nuts"),
		i18n.T("  Server saved in the config:"),
		"    /connect libera",
		"",
		i18n.T("Syntax: /connect <host> [port] [options]  ·  /connect <name>"),
		i18n.T("  -nick <nick>          your nick"),
		i18n.T("  -user <user>          user (on a BNC: user/network)"),
		i18n.T("  -pass <password>      server or BNC password"),
		i18n.T("  -sasl <user:password> SASL login"),
		i18n.T("  -join <#a,#b>         channels to join on connect"),
		i18n.T("  -tls / -notls         force TLS (by default only on port 6697)"),
		i18n.T("  -insecure             accept invalid certificates"),
		i18n.T("  -name <name>          name in the sidebar"),
		"",
		i18n.T("When it connects wisirc asks whether to save it; saved servers are managed in ≡ → Servers."),
	}
}

// showConnect opens the connect screen: examples above, typing below.
func (m *Model) showConnect() {
	m.overlay = &textOverlay{title: i18n.T("Connect to server (Enter to connect, Esc to cancel)"), lines: connectText(), typing: true, scroll: 1 << 30}
	m.prefill("/connect ")
}
