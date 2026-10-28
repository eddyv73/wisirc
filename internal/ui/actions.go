package ui

import (
	"os"
	"os/exec"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/eddyv73/wisirc/internal/i18n"
	"github.com/eddyv73/wisirc/internal/irc"
	"github.com/eddyv73/wisirc/internal/state"
)

func (m *Model) sendOn(network string, line string) {
	s, ok := m.d.Nets.Get(network)
	if !ok {
		m.errorLine(i18n.T("not connected"))
		return
	}
	if err := s.Send(line); err != nil {
		m.errorLine(i18n.T("could not send: %s", err.Error()))
	}
}

func (m *Model) prefill(text string) {
	m.input.SetValue(text)
	m.input.CursorEnd()
}

// runMenu executes a menu item chosen by mouse or keyboard.
func (m *Model) runMenu(menu Menu, act MenuAction) tea.Cmd {
	ref := menu.Ref
	if menu.Kind == MenuGlobal {
		ref = m.st.Active
	}
	nick := menu.Nick
	mode := func(flag string) { m.sendOn(ref.Network, irc.Encode("MODE", ref.Name, flag, nick)) }
	switch act {
	case ActWhois:
		m.sendOn(ref.Network, irc.Encode("WHOIS", nick))
	case ActQuery:
		q := state.BufRef{Network: ref.Network, Name: nick}
		m.ensureBuffer(q, state.KindQuery)
		m.switchTo(q)
	case ActOp:
		mode("+o")
	case ActDeop:
		mode("-o")
	case ActVoice:
		mode("+v")
	case ActDevoice:
		mode("-v")
	case ActKick:
		m.sendOn(ref.Network, irc.Encode("KICK", ref.Name, nick))
	case ActBan:
		m.sendOn(ref.Network, irc.Encode("MODE", ref.Name, "+b", nick+"!*@*"))
	case ActPart:
		m.sendOn(ref.Network, irc.Encode("PART", ref.Name))
	case ActClose:
		m.closeBuffer(ref)
	case ActMarkRead:
		if b := m.st.Buffer(ref); b != nil {
			b.Unread, b.Mentioned = 0, false
		}
	case ActTopic:
		if b := m.st.Buffer(ref); b != nil {
			topic := b.Topic
			if topic == "" {
				topic = i18n.T("(no topic)")
			}
			m.st.AddLine(ref, state.Line{Kind: state.LineTopic, Text: i18n.T("Topic of %s: %s", b.Ref.Name, topic)})
		}
	case ActConnect:
		if menu.Kind == MenuServer {
			m.connect(ref.Network)
		} else {
			m.showConnect()
		}
	case ActDisconnect:
		m.d.Nets.Disconnect(ref.Network)
	case ActJoin:
		if menu.Kind == MenuServer {
			m.switchTo(state.BufRef{Network: ref.Network})
		}
		m.prefill("/join #")
	case ActLogs:
		m.showLogs(ref)
	case ActTheme:
		m.openThemeMenu()
	case ActServerColor:
		m.openColorMenu(ref.Network, menu.X, menu.Y)
	case ActPickColor:
		m.chooseServerColor(ref.Network, menu.Args[menu.Selected])
	case ActDashboard:
		m.toggleDashboard()
	case ActLang:
		m.openLangMenu()
	case ActPickLang:
		m.chooseLang(menu.Args[menu.Selected])
	case ActPickTheme:
		m.chooseTheme(menu.Args[menu.Selected])
	case ActConfig:
		return m.editConfig()
	case ActHelp:
		m.showHelp()
	case ActQuit:
		return m.quit("")
	case ActServers:
		m.showServers()
	case ActSaveAdHoc:
		if err := m.d.Nets.SaveAdHoc(ref.Network); err != nil {
			m.errorLine(err.Error())
		} else {
			m.info(i18n.T("%s saved to servers.toml: it will connect when wisirc starts", ref.Network))
		}
	case ActSkipSave:
		m.info(i18n.T("Not saved. You can save it later in ≡ → Servers."))
	case ActDelServer:
		m.deleteServer(ref.Network)
	}
	return nil
}

func (m *Model) showLogs(ref state.BufRef) {
	if m.d.Logger == nil {
		return
	}
	lines, err := m.d.Logger.Tail(ref.Network, ref.Name, 2000)
	if err != nil {
		m.errorLine("logs: " + err.Error())
		return
	}
	name := ref.Name
	if name == "" {
		name = ref.Network
	}
	if len(lines) == 0 {
		lines = []string{i18n.T("(no logs yet)")}
	}
	m.overlay = &textOverlay{title: i18n.T("Logs of %s (Esc to close)", name), lines: lines}
}

func (m *Model) editConfig() tea.Cmd {
	editor := strings.Fields(os.Getenv("VISUAL"))
	if len(editor) == 0 {
		editor = strings.Fields(os.Getenv("EDITOR"))
	}
	if len(editor) == 0 {
		editor = []string{"vi"}
	}
	cmd := exec.Command(editor[0], append(editor[1:], m.d.CfgPath)...)
	return tea.ExecProcess(cmd, func(err error) tea.Msg { return configEditedMsg{err: err} })
}
