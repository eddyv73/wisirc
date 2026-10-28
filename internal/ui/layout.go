package ui

const (
	DefaultSidebarW = 18
	DefaultNickW    = 14
)

type Rect struct{ X, Y, W, H int }

func (r Rect) Contains(x, y int) bool {
	return r.W > 0 && r.H > 0 && x >= r.X && x < r.X+r.W && y >= r.Y && y < r.Y+r.H
}

type Zone int

const (
	ZoneNone Zone = iota
	ZoneMenuIcon
	ZoneSidebar
	ZoneSidebarBorder
	ZoneHeader
	ZoneChat
	ZoneNickBorder
	ZoneNicklist
	ZoneInput
	ZoneStatus
)

// Layout is the screen geometry. Border X values are -1 when hidden.
type Layout struct {
	W, H                                                     int
	MenuIcon, Sidebar, Header, Chat, Nicklist, Input, Status Rect
	SidebarBorderX, NickBorderX                              int
}

func clamp(v, lo, hi int) int {
	if v > hi {
		v = hi
	}
	if v < lo {
		v = lo
	}
	return v
}

// ComputeLayout places the panels for a w×h screen.
func ComputeLayout(w, h, sidebarW, nickW int) Layout {
	w, h = max(w, 0), max(h, 0)
	l := Layout{W: w, H: h, SidebarBorderX: -1, NickBorderX: -1}
	if w == 0 || h == 0 {
		return l
	}
	l.Status = Rect{0, h - 1, w, 1}
	body := h - 1 // rows above the status bar

	mainX := 0
	if w >= 40 {
		sw := clamp(sidebarW, 12, w/3)
		l.Sidebar = Rect{0, 0, sw, body}
		l.SidebarBorderX = sw
		mainX = sw + 1
	}
	right := w
	if w >= 60 {
		nw := clamp(nickW, 10, w/4)
		l.NickBorderX = w - nw - 1
		right = l.NickBorderX
		l.Nicklist = Rect{w - nw, 0, nw, max(body-1, 0)}
	}

	l.MenuIcon = Rect{0, 0, min(3, w), min(1, body)}
	headerX := mainX
	if l.Sidebar.W == 0 {
		headerX = l.MenuIcon.W
	}
	if body >= 2 {
		l.Input = Rect{mainX, body - 1, w - mainX, 1}
		l.Header = Rect{headerX, 0, max(right-headerX, 0), 1}
	} else if body == 1 {
		l.Input = Rect{mainX, 0, w - mainX, 1}
		l.MenuIcon = Rect{}
	}
	if h >= 5 {
		l.Chat = Rect{mainX, 1, max(right-mainX, 0), body - 2}
	} else {
		l.Chat = Rect{mainX, 1, max(right-mainX, 0), 0}
	}
	return l
}

// Hit returns the zone under screen cell (x, y).
func (l Layout) Hit(x, y int) Zone {
	switch {
	case x < 0 || y < 0 || x >= l.W || y >= l.H:
		return ZoneNone
	case l.MenuIcon.Contains(x, y):
		return ZoneMenuIcon
	case l.Status.Contains(x, y):
		return ZoneStatus
	case l.Input.Contains(x, y):
		return ZoneInput
	case x == l.SidebarBorderX:
		return ZoneSidebarBorder
	case x == l.NickBorderX:
		return ZoneNickBorder
	case l.Sidebar.Contains(x, y):
		return ZoneSidebar
	case l.Header.Contains(x, y):
		return ZoneHeader
	case l.Chat.Contains(x, y):
		return ZoneChat
	case l.Nicklist.Contains(x, y):
		return ZoneNicklist
	}
	return ZoneNone
}
