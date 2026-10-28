package ui

import (
	tea "charm.land/bubbletea/v2"
	"github.com/eddyv73/wisirc/internal/i18n"
	"github.com/eddyv73/wisirc/internal/state"
)

func (m *Model) sidebarRowAt(y int) (SidebarRow, bool) {
	rows := SidebarRows(m.st)
	i := y - 1 + m.sideScroll
	if y < 1 || i < 0 || i >= len(rows) {
		return SidebarRow{}, false
	}
	return rows[i], true
}

func (m *Model) nickAt(y int) (state.Member, bool) {
	b := m.activeBuffer()
	if b == nil || b.Kind != state.KindChannel {
		return state.Member{}, false
	}
	members := b.SortedMembers()
	i := y - 1 + m.nickScroll
	if y < 1 || i < 0 || i >= len(members) {
		return state.Member{}, false
	}
	return members[i], true
}

// chatNickAt returns the nick drawn at screen cell (x, y) in the chat panel.
func (m *Model) chatNickAt(x, y int) (string, bool) {
	l := m.Layout()
	b := m.activeBuffer()
	if b == nil {
		return "", false
	}
	rows, _ := tailRows(b.Lines, l.Chat.W, l.Chat.H, m.scroll[refKey(m.st.Active)])
	// rows are bottom-aligned in the panel
	i := y - l.Chat.Y - (l.Chat.H - len(rows))
	if i < 0 || i >= len(rows) {
		return "", false
	}
	r := rows[i]
	col := x - l.Chat.X
	if r.Nick == "" || col < r.NickStart || col >= r.NickEnd {
		return "", false
	}
	return r.Nick, true
}

func (m *Model) handleClick(ms tea.Mouse) tea.Cmd {
	m.flash, m.sel = "", nil
	if m.menu != nil {
		menu := *m.menu
		m.menu = nil
		if i := menu.ItemAt(ms.X, ms.Y, m.w, m.menuH()); i >= 0 {
			menu.Selected = i
			return m.runMenu(menu, menu.Items[i].Action)
		}
		if menu.Kind == MenuTheme {
			m.revertTheme()
		}
		if menu.Kind == MenuLang {
			i18n.SetLang(m.langBefore)
		}
		if menu.Kind == MenuColor {
			m.revertColor(menu.Ref.Network)
		}
		return nil
	}
	l := m.Layout()
	zone := l.Hit(ms.X, ms.Y)
	if m.dash != nil {
		switch zone {
		case ZoneChat:
			return m.dashClick(ms)
		case ZoneHeader:
			return nil
		}
		m.dash = nil
	}
	if m.admin != nil {
		switch zone {
		case ZoneChat:
			if ms.Button == tea.MouseLeft {
				return m.adminClick(ms.X, ms.Y)
			}
			return nil
		case ZoneHeader:
			return nil
		}
		m.admin = nil
	}
	if zone == ZoneChat && ms.Button == tea.MouseLeft {
		// a drag selects text; a plain click on a nick opens its menu on release
		m.sel = &selection{x0: ms.X, y0: ms.Y, x1: ms.X, y1: ms.Y}
		return nil
	}
	if m.overlay != nil && zone != ZoneMenuIcon {
		if zone != ZoneChat && zone != ZoneHeader {
			m.overlay = nil
		}
		return nil
	}
	left, right := ms.Button == tea.MouseLeft, ms.Button == tea.MouseRight
	if !left && !right {
		return nil
	}
	switch zone {
	case ZoneMenuIcon:
		if left {
			g := GlobalMenu()
			m.menu = &g
		}
	case ZoneSidebarBorder:
		if left {
			m.drag = dragSidebar
		}
	case ZoneNickBorder:
		if left {
			m.drag = dragNick
		}
	case ZoneStatus:
		if left {
			return m.statusClick(ms.X)
		}
	case ZoneSidebar:
		if ms.Y == 0 {
			// the "wisirc" title next to ≡ opens the dashboard
			if left && m.dash == nil {
				m.toggleDashboard()
			}
			return nil
		}
		row, ok := m.sidebarRowAt(ms.Y)
		if !ok {
			return nil
		}
		if right {
			if row.IsNetwork {
				n := m.st.Network(row.Ref.Network)
				menu := ServerMenu(row.Ref, n != nil && n.Connected, ms.X, ms.Y)
				m.menu = &menu
			} else {
				menu := ChannelMenu(row.Ref, ms.X, ms.Y)
				m.menu = &menu
			}
			return nil
		}
		if row.IsNetwork && ms.X <= 1 {
			if n := m.st.Network(row.Ref.Network); n != nil {
				n.Collapsed = !n.Collapsed
			}
			return nil
		}
		m.switchTo(row.Ref)
	case ZoneNicklist:
		if mem, ok := m.nickAt(ms.Y); ok {
			menu := NickMenu(m.st.Active, mem.Nick, ms.X, ms.Y)
			m.menu = &menu
		}
	case ZoneChat:
		if nick, ok := m.chatNickAt(ms.X, ms.Y); ok {
			menu := NickMenu(m.st.Active, nick, ms.X, ms.Y)
			m.menu = &menu
		}
	}
	return nil
}

func (m *Model) handleMotion(ms tea.Mouse) {
	if s := m.sel; s != nil && !s.done {
		c := m.Layout().Chat
		s.x1 = min(max(ms.X, c.X), c.X+c.W-1)
		s.y1 = min(max(ms.Y, c.Y), c.Y+c.H-1)
		s.moved = s.moved || s.x1 != s.x0 || s.y1 != s.y0
		return
	}
	switch m.drag {
	case dragSidebar:
		m.sidebarW = ComputeLayout(m.w, m.h, ms.X, m.nickW).Sidebar.W
	case dragNick:
		m.nickW = ComputeLayout(m.w, m.h, m.sidebarW, m.w-ms.X-1).Nicklist.W
	}
}

func (m *Model) handleWheel(ms tea.Mouse) {
	delta := 0
	switch ms.Button {
	case tea.MouseWheelUp:
		delta = WheelStep
	case tea.MouseWheelDown:
		delta = -WheelStep
	default:
		return
	}
	m.sel = nil
	if m.admin != nil || m.dash != nil {
		return
	}
	if m.overlay != nil {
		m.overlay.scroll = max(m.overlay.scroll+delta, 0)
		return
	}
	l := m.Layout()
	switch l.Hit(ms.X, ms.Y) {
	case ZoneChat, ZoneHeader:
		m.scrollChat(delta)
	case ZoneSidebar:
		// wheel up shows earlier rows
		m.sideScroll = min(max(m.sideScroll-delta, 0), max(len(SidebarRows(m.st))-(l.Sidebar.H-1), 0))
	case ZoneNicklist:
		if b := m.activeBuffer(); b != nil {
			m.nickScroll = min(max(m.nickScroll-delta, 0), max(len(b.Members)-(l.Nicklist.H-1), 0))
		}
	}
}
