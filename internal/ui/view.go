package ui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/eddyv73/wisirc/internal/i18n"
	"github.com/eddyv73/wisirc/internal/state"
)

// fit truncates or pads s (which may contain ANSI) to exactly w columns.
func fit(s string, w int) string {
	if w <= 0 {
		return ""
	}
	if ansi.StringWidth(s) > w {
		s = ansi.Truncate(s, w, "…")
	}
	if pad := w - ansi.StringWidth(s); pad > 0 {
		s += strings.Repeat(" ", pad)
	}
	return s
}

func (m *Model) View() tea.View {
	v := tea.NewView(m.render())
	v.AltScreen = true
	v.MouseMode = tea.MouseModeCellMotion
	v.ReportFocus = true
	v.WindowTitle = "wisirc"
	v.BackgroundColor, v.ForegroundColor = m.theme.bg, m.theme.fg
	return v
}

func (m *Model) render() string {
	if m.w <= 0 || m.h <= 0 {
		return ""
	}
	l := m.Layout()
	th := m.theme
	border := lipgloss.NewStyle().Foreground(th.border).Render("│")

	side := m.renderSidebar(l)
	header := m.renderHeader(l)
	chat := m.highlight(m.renderChat(l), l.Chat)
	nicks := m.renderNicklist(l)
	input := m.renderInput(l)

	rows := make([]string, m.h)
	for y := 0; y < m.h; y++ {
		var b strings.Builder
		switch {
		case y == l.Status.Y && l.Status.W > 0:
			rows[y] = m.renderStatus(l)
			continue
		case l.Input.H > 0 && y == l.Input.Y:
			if l.Sidebar.W > 0 {
				b.WriteString(side[y] + border)
			}
			b.WriteString(input)
			rows[y] = fit(b.String(), m.w)
			continue
		}
		if l.Sidebar.W > 0 {
			b.WriteString(side[y] + border)
		} else if y == 0 && l.MenuIcon.W > 0 {
			b.WriteString(m.menuIcon(l.MenuIcon.W))
		}
		switch {
		case y == 0 && l.Header.W > 0:
			b.WriteString(header)
		case y >= l.Chat.Y && y < l.Chat.Y+l.Chat.H:
			b.WriteString(chat[y-l.Chat.Y])
		default:
			b.WriteString(strings.Repeat(" ", max(l.Chat.W, 0)))
		}
		if l.Nicklist.W > 0 && y < l.Nicklist.H {
			b.WriteString(border + nicks[y])
		}
		rows[y] = fit(b.String(), m.w)
	}
	out := strings.Join(rows, "\n")
	if m.menu != nil {
		m.menu.Scroll(m.menuH())
		r := m.menu.Rect(m.w, m.menuH())
		out = Overlay(out, m.menu.Render(m.menuH()), r.X, r.Y)
	}
	return out
}

func (m *Model) menuIcon(w int) string {
	style := lipgloss.NewStyle().Foreground(m.theme.accent).Bold(true)
	if m.menu != nil && m.menu.Kind == MenuGlobal {
		style = style.Reverse(true)
	}
	return fit(style.Render(" ≡ "), w)
}

func (m *Model) renderSidebar(l Layout) []string {
	out := make([]string, max(l.Sidebar.H, l.H))
	if l.Sidebar.W == 0 {
		return out
	}
	w := l.Sidebar.W
	out[0] = m.menuIcon(3) + fit(lipgloss.NewStyle().Foreground(m.theme.dim).Render("wisirc"), w-3)
	rows := SidebarRows(m.st)
	for y := 1; y < l.Sidebar.H; y++ {
		i := y - 1 + m.sideScroll
		if i >= len(rows) {
			out[y] = strings.Repeat(" ", w)
			continue
		}
		r := rows[i]
		style := lipgloss.NewStyle()
		if r.IsNetwork {
			style = style.Bold(true)
			if n := m.st.Network(r.Ref.Network); n != nil && !n.Connected {
				style = style.Foreground(m.theme.dim)
			}
		} else if b := m.st.Buffer(r.Ref); b != nil {
			// channels share their server's tint so each group reads as one
			switch {
			case b.Mentioned:
				style = style.Foreground(m.theme.mention).Bold(true)
			case b.Unread > 0:
				style = style.Foreground(m.channelTint(r.Ref.Network)).Bold(true)
			case b.Kind == state.KindChannel && !b.Joined:
				style = style.Foreground(m.theme.dim)
			default:
				style = style.Foreground(m.channelTint(r.Ref.Network))
			}
		}
		active := refKey(r.Ref) == refKey(m.st.Active)
		label := fit(r.Label, w)
		tint := m.serverTint(r.Ref.Network)
		switch {
		case active:
			// the selected row takes its server's own shade
			out[y] = style.Foreground(serverText).Background(tint).Render(label)
		case r.IsNetwork:
			glyph, rest := ansi.Cut(label, 0, 2), ansi.Cut(label, 2, w)
			out[y] = lipgloss.NewStyle().Foreground(m.channelTint(r.Ref.Network)).Render(glyph) + style.Render(rest)
		default:
			out[y] = style.Render(label)
		}
	}
	return out
}

func (m *Model) renderHeader(l Layout) string {
	if l.Header.W == 0 {
		return ""
	}
	b := m.activeBuffer()
	title := ""
	if b != nil {
		switch b.Kind {
		case state.KindStatus:
			title = i18n.T("%s — status", b.Ref.Network)
			if b.Ref.Network == "" {
				title = i18n.T("wisirc — no servers: ≡ → Servers")
			}
		case state.KindChannel:
			title = b.Ref.Name
			if b.Topic != "" {
				title += " — " + ansi.Strip(RenderIRC(b.Topic))
			}
		default:
			title = i18n.T("%s — private", b.Ref.Name)
		}
	}
	if m.overlay != nil {
		title = m.overlay.title
	}
	if m.admin != nil {
		title = i18n.T("Servers")
	}
	if m.dash != nil {
		title = i18n.T("Dashboard")
	}
	style := lipgloss.NewStyle().Bold(true).Foreground(m.theme.accent)
	return style.Render(fit(" "+sanitize(title), l.Header.W))
}

func (m *Model) renderChat(l Layout) []string {
	out := make([]string, max(l.Chat.H, 0))
	if l.Chat.H <= 0 {
		return out
	}
	blank := strings.Repeat(" ", max(l.Chat.W, 0))
	for i := range out {
		out[i] = blank
	}
	if m.dash != nil {
		return m.renderDashboard(l.Chat.W, l.Chat.H)
	}
	if m.admin != nil {
		return m.renderAdmin(l.Chat.W, l.Chat.H)
	}
	if m.overlay != nil {
		var lines []string
		for _, line := range m.overlay.lines {
			lines = append(lines, wrapOverlayLine(sanitize(line), l.Chat.W)...)
		}
		end := max(len(lines)-m.overlay.scroll, 0)
		start := max(end-l.Chat.H, 0)
		if start == 0 {
			end = min(len(lines), l.Chat.H)
			m.overlay.scroll = max(len(lines)-end, 0)
		}
		for i, line := range lines[start:end] {
			out[i] = fit(line, l.Chat.W)
		}
		return out
	}
	b := m.activeBuffer()
	if b == nil {
		return out
	}
	key := refKey(b.Ref)
	rows, s := tailRows(b.Lines, l.Chat.W, l.Chat.H, m.scroll[key])
	m.scroll[key] = s
	offset := l.Chat.H - len(rows)
	for i, r := range rows {
		out[offset+i] = fit(r.Text, l.Chat.W)
	}
	if s > 0 && l.Chat.W > 12 {
		out[0] = Overlay(out[0], lipgloss.NewStyle().Reverse(true).Render(fmt.Sprintf(" ↑ %d ", s)), l.Chat.W-len(fmt.Sprintf(" ↑ %d ", s)), 0)
	}
	return out
}

func (m *Model) renderNicklist(l Layout) []string {
	out := make([]string, max(l.Nicklist.H, 0))
	w := l.Nicklist.W
	for i := range out {
		out[i] = strings.Repeat(" ", max(w, 0))
	}
	if w == 0 || len(out) == 0 {
		return out
	}
	b := m.activeBuffer()
	if b == nil || b.Kind != state.KindChannel {
		return out
	}
	members := b.SortedMembers()
	out[0] = lipgloss.NewStyle().Foreground(m.theme.dim).Render(fit(i18n.T(" %d users", len(members)), w))
	for y := 1; y < len(out); y++ {
		i := y - 1 + m.nickScroll
		if i >= len(members) {
			break
		}
		mem := members[i]
		mode := " "
		if mem.Modes != "" {
			mode = mem.Modes[:1]
		}
		nick := lipgloss.NewStyle().Foreground(NickColor(mem.Nick)).Render(sanitize(mem.Nick))
		out[y] = fit(lipgloss.NewStyle().Foreground(m.theme.accent).Render(mode)+nick, w)
	}
	return out
}

func (m *Model) renderInput(l Layout) string {
	if l.Input.W <= 0 {
		return ""
	}
	nick := "?"
	if n := m.st.Network(m.st.Active.Network); n != nil && n.Nick != "" {
		nick = n.Nick
	}
	prompt := lipgloss.NewStyle().Foreground(m.theme.accent).Render("[" + sanitize(nick) + "] > ")
	pw := ansi.StringWidth(prompt)
	m.input.SetWidth(max(l.Input.W-pw-1, 1))
	return fit(prompt+m.input.View(), l.Input.W)
}

func (m *Model) renderStatus(l Layout) string {
	return m.renderStatusBar(l.Status.W)
}

// wrapOverlayLine wraps a panel line to width, keeping its indentation on
// continuation rows so examples stay readable in narrow windows.
func wrapOverlayLine(line string, width int) []string {
	if width <= 0 || ansi.StringWidth(line) <= width {
		return []string{line}
	}
	indent := line[:len(line)-len(strings.TrimLeft(line, " "))]
	if len(indent)+2 >= width {
		indent = ""
	}
	rows := strings.Split(ansi.Wrap(strings.TrimLeft(line, " "), width-len(indent)-2, ""), "\n")
	for i := range rows {
		if i == 0 {
			rows[i] = indent + rows[i]
		} else {
			rows[i] = indent + "  " + rows[i]
		}
	}
	return rows
}
