package ui

import (
	"fmt"
	"image/color"
	"sort"
	"strconv"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/eddyv73/wisirc/internal/i18n"
	"github.com/eddyv73/wisirc/internal/state"
)

// dashPingFrames is how often (in animation frames) the dashboard measures
// lag while it is open: 25 × 200ms = every 5s instead of every 30s.
const dashPingFrames = 25

const (
	dashCardW = 34
	dashCardH = 6
)

// dashScreen is the dashboard: one card per server with live status, lag
// graph, uptime, channels and unread counts. It is drawn in the chat panel.
type dashScreen struct {
	sel   int
	cards []dashCard // layout of the last render, panel-relative
}

type dashCard struct {
	name       string
	x, y, w, h int
}

func (m *Model) toggleDashboard() {
	if m.dash != nil {
		m.dash = nil
		return
	}
	m.overlay, m.menu, m.admin = nil, nil, nil
	m.dash = &dashScreen{}
	for i, n := range m.dashNetworks() {
		if n == m.st.Active.Network {
			m.dash.sel = i
		}
	}
	m.pingAll()
}

// dashNetworks lists servers in sidebar order, then configured servers that
// are not running.
func (m *Model) dashNetworks() []string {
	var out []string
	seen := map[string]bool{}
	for _, n := range m.st.Networks {
		if n.Name != "" {
			out = append(out, n.Name)
			seen[n.Name] = true
		}
	}
	var rest []string
	if cfg := m.d.Nets.Config(); cfg != nil {
		for name := range cfg.Servers {
			if !seen[name] {
				rest = append(rest, name)
			}
		}
	}
	sort.Strings(rest)
	return append(out, rest...)
}

// pingAll measures lag on every connected server now.
func (m *Model) pingAll() {
	for _, n := range m.st.Networks {
		if !n.Connected {
			continue
		}
		if s, ok := m.d.Nets.Get(n.Name); ok {
			s.Send("PING wisirc-" + strconv.FormatInt(time.Now().UnixNano(), 10))
		}
	}
}

// dashTick runs every animation frame while the dashboard is open.
func (m *Model) dashTick() {
	if m.dash != nil && m.frame%dashPingFrames == 0 {
		m.pingAll()
	}
}

func uptime(d time.Duration) string {
	switch {
	case d < time.Minute:
		return fmt.Sprintf("%ds", int(d.Seconds()))
	case d < time.Hour:
		return fmt.Sprintf("%dm", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh %dm", int(d.Hours()), int(d.Minutes())%60)
	}
	return fmt.Sprintf("%dd %dh", int(d.Hours())/24, int(d.Hours())%24)
}

func (m *Model) dashKey(msg tea.KeyPressMsg) tea.Cmd {
	names := m.dashNetworks()
	cols := max(len(m.dash.cards), 1)
	if len(m.dash.cards) > 0 {
		cols = 0
		for _, c := range m.dash.cards {
			if c.y == m.dash.cards[0].y {
				cols++
			}
		}
	}
	move := func(d int) {
		if len(names) > 0 {
			m.dash.sel = min(max(m.dash.sel+d, 0), len(names)-1)
		}
	}
	switch msg.String() {
	case "esc", "f2", "q":
		m.dash = nil
	case "ctrl+c":
		return m.quit("")
	case "left", "h":
		move(-1)
	case "right", "l", "tab":
		move(1)
	case "up", "k":
		move(-cols)
	case "down", "j":
		move(cols)
	case "enter":
		if m.dash.sel < len(names) {
			m.dashOpen(names[m.dash.sel])
		}
	case "c":
		if m.dash.sel < len(names) {
			name := names[m.dash.sel]
			if _, running := m.d.Nets.Get(name); running {
				m.d.Nets.Disconnect(name)
			} else if _, err := m.d.Nets.Connect(name); err == nil {
				m.st.AddNetwork(name)
			}
		}
	}
	return nil
}

// dashOpen leaves the dashboard for a server's status buffer.
func (m *Model) dashOpen(name string) {
	m.dash = nil
	if m.st.Network(name) == nil {
		m.connect(name)
		return
	}
	m.switchTo(state.BufRef{Network: name})
}

func (m *Model) dashCardAt(x, y int) (int, bool) {
	c := m.Layout().Chat
	px, py := x-c.X, y-c.Y
	for i, card := range m.dash.cards {
		if px >= card.x && px < card.x+card.w && py >= card.y && py < card.y+card.h {
			return i, true
		}
	}
	return 0, false
}

func (m *Model) dashClick(ms tea.Mouse) tea.Cmd {
	i, ok := m.dashCardAt(ms.X, ms.Y)
	if !ok {
		return nil
	}
	m.dash.sel = i
	name := m.dash.cards[i].name
	switch ms.Button {
	case tea.MouseLeft:
		m.dashOpen(name)
	case tea.MouseRight:
		n := m.st.Network(name)
		menu := ServerMenu(state.BufRef{Network: name}, n != nil && n.Connected, ms.X, ms.Y)
		m.menu = &menu
	}
	return nil
}

// cardBox draws a rounded (or double, when selected) box around lines.
func cardBox(title string, lines []string, w int, border color.Color, selected bool) []string {
	tl, tr, bl, br, hz, vt := "╭", "╮", "╰", "╯", "─", "│"
	if selected {
		tl, tr, bl, br, hz, vt = "╔", "╗", "╚", "╝", "═", "║"
	}
	bs := lipgloss.NewStyle().Foreground(border)
	inner := w - 2
	title = ansi.Truncate(sanitize(title), inner-4, "…")
	head := hz + " " + lipgloss.NewStyle().Bold(true).Render(title) + " "
	head += strings.Repeat(hz, max(inner-ansi.StringWidth(head), 0))
	out := []string{bs.Render(tl) + colorHead(head, bs) + bs.Render(tr)}
	for _, l := range lines {
		out = append(out, bs.Render(vt)+" "+fit(l, inner-2)+" "+bs.Render(vt))
	}
	out = append(out, bs.Render(bl+strings.Repeat(hz, inner)+br))
	return out
}

// colorHead colors the border runs of a card title row, keeping the title bold.
func colorHead(head string, bs lipgloss.Style) string {
	i := strings.Index(head, " ")
	j := strings.LastIndex(head, " ")
	if i < 0 || j <= i {
		return bs.Render(head)
	}
	return bs.Render(head[:i]) + head[i:j+1] + bs.Render(head[j+1:])
}

func (m *Model) dashCardLines(name string, w int) []string {
	th := m.theme
	dim := lipgloss.NewStyle().Foreground(th.dim)
	n := m.st.Network(name)
	var status, right string
	switch {
	case n != nil && n.Connected:
		status = lipgloss.NewStyle().Foreground(th.ok).Render("● " + i18n.T("connected"))
		if !n.ConnectedAt.IsZero() {
			right = dim.Render("⏱ " + uptime(time.Since(n.ConnectedAt)))
		}
	case n != nil && n.Connecting:
		status = lipgloss.NewStyle().Foreground(th.accent).Render(spinFrames[m.frame%len(spinFrames)] + " " + i18n.T("connecting…"))
	default:
		status = dim.Render("○ " + i18n.T("disconnected"))
	}
	if pad := w - ansi.StringWidth(status) - ansi.StringWidth(right); pad > 0 {
		status += strings.Repeat(" ", pad) + right
	}

	lag := dim.Render(i18n.T("no lag samples yet"))
	if n != nil && n.Connected && len(n.LagHistory) > 0 {
		c := lagColor(n.Lag)
		// the newest bar flickers with the pulse: the graph is live
		r := []rune(Sparkline(n.LagHistory))
		head, last := string(r[:len(r)-1]), string(r[len(r)-1])
		lastStyle := lipgloss.NewStyle().Foreground(c)
		if m.frame%4 < 2 {
			lastStyle = lastStyle.Bold(true).Foreground(th.accent)
		}
		lag = lipgloss.NewStyle().Foreground(c).Render(head) + lastStyle.Render(last) + "  " +
			lipgloss.NewStyle().Foreground(c).Render(pulseFrames[m.frame%len(pulseFrames)]+fmt.Sprintf(" %dms", n.Lag.Milliseconds()))
	} else if n != nil && n.Connected {
		lag = lipgloss.NewStyle().Foreground(th.dim).Render(pulseFrames[m.frame%len(pulseFrames)] + " " + i18n.T("measuring…"))
	}

	var chans []string
	unread, mentions := 0, 0
	if n != nil {
		for _, b := range n.Buffers {
			if b.Kind == state.KindChannel && b.Joined {
				chans = append(chans, sanitize(b.Ref.Name))
			}
			unread += b.Unread
			if b.Mentioned {
				mentions++
			}
		}
	}
	chanLine := dim.Render(i18n.T("no channels"))
	if len(chans) > 0 {
		chanLine = lipgloss.NewStyle().Foreground(th.accent).Render(ansi.Truncate(strings.Join(chans, " "), w, "…"))
	}
	var act []string
	if unread > 0 {
		act = append(act, lipgloss.NewStyle().Foreground(th.unread).Bold(true).Render(i18n.T("%d new", unread)))
	}
	if mentions > 0 {
		label := i18n.T("%d mentions", mentions)
		if mentions == 1 {
			label = i18n.T("1 mention")
		}
		act = append(act, lipgloss.NewStyle().Foreground(th.mention).Bold(true).Render(label))
	}
	actLine := dim.Render(i18n.T("all read"))
	if len(act) > 0 {
		actLine = strings.Join(act, dim.Render(" · "))
	}
	return []string{status, lag, chanLine, actLine}
}

// renderDashboard draws the cards in a grid that fits the panel width.
func (m *Model) renderDashboard(w, h int) []string {
	d := m.dash
	d.cards = nil
	names := m.dashNetworks()
	if len(names) > 0 {
		d.sel = min(max(d.sel, 0), len(names)-1)
	}
	out := make([]string, h)
	for i := range out {
		out[i] = strings.Repeat(" ", max(w, 0))
	}
	if h < 3 {
		return out
	}
	connected, total, lagSum, lagN := 0, len(names), time.Duration(0), 0
	for _, name := range names {
		if n := m.st.Network(name); n != nil && n.Connected {
			connected++
			if n.Lag > 0 {
				lagSum += n.Lag
				lagN++
			}
		}
	}
	summary := i18n.T("servers: %d · connected: %d", total, connected)
	if lagN > 0 {
		summary += " · " + i18n.T("average %dms", (lagSum/time.Duration(lagN)).Milliseconds())
	}
	live := lipgloss.NewStyle().Foreground(m.theme.ok).Render(pulseFrames[m.frame%len(pulseFrames)] + " LIVE")
	out[0] = fit(" "+live+"  "+lipgloss.NewStyle().Bold(true).Render(summary), w)
	out[1] = fit(" "+adminDim.Render(i18n.T("Enter/click go · right click or c connect/disconnect · arrows choose · F2/Esc close")), w)
	if len(names) == 0 {
		out[3] = fit(" "+adminDim.Render(i18n.T("No servers yet: add one in ≡ → Servers")), w)
		return out
	}

	cw := min(dashCardW, max(w-2, 12))
	cols := max((w-1)/(cw+1), 1)
	for i, name := range names {
		col, row := i%cols, i/cols
		x, y := 1+col*(cw+1), 3+row*(dashCardH+1)
		if y+dashCardH > h {
			break
		}
		d.cards = append(d.cards, dashCard{name: name, x: x, y: y, w: cw, h: dashCardH})
		tint := m.serverTint(name)
		if m.st.Network(name) == nil {
			tint = m.theme.dim
		}
		box := cardBox(name, m.dashCardLines(name, cw-4), cw, tint, i == d.sel)
		for j, line := range box {
			out[y+j] = Overlay(out[y+j], line, x, 0)
		}
	}
	for i := range out {
		out[i] = fit(out[i], w)
	}
	return out
}
