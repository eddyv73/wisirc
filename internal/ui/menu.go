package ui

import (
	"fmt"
	"image/color"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/eddyv73/wisirc/internal/i18n"
	"github.com/eddyv73/wisirc/internal/state"
)

type MenuAction string

const (
	ActWhois       MenuAction = "whois"
	ActQuery       MenuAction = "query"
	ActOp          MenuAction = "op"
	ActDeop        MenuAction = "deop"
	ActVoice       MenuAction = "voice"
	ActDevoice     MenuAction = "devoice"
	ActKick        MenuAction = "kick"
	ActBan         MenuAction = "ban"
	ActPart        MenuAction = "part"
	ActClose       MenuAction = "close"
	ActMarkRead    MenuAction = "markread"
	ActTopic       MenuAction = "topic"
	ActConnect     MenuAction = "connect"
	ActDisconnect  MenuAction = "disconnect"
	ActJoin        MenuAction = "join"
	ActLogs        MenuAction = "logs"
	ActTheme       MenuAction = "theme"
	ActConfig      MenuAction = "config"
	ActHelp        MenuAction = "help"
	ActQuit        MenuAction = "quit"
	ActServers     MenuAction = "servers"
	ActSaveAdHoc   MenuAction = "saveadhoc"
	ActDelServer   MenuAction = "delserver"
	ActDismiss     MenuAction = "dismiss"
	ActPickTheme   MenuAction = "picktheme"
	ActLang        MenuAction = "lang"
	ActPickLang    MenuAction = "picklang"
	ActDashboard   MenuAction = "dashboard"
	ActServerColor MenuAction = "servercolor"
	ActPickColor   MenuAction = "pickcolor"
	ActSkipSave    MenuAction = "skipsave"
)

type MenuItem struct {
	Label  string
	Action MenuAction
}

type MenuKind int

const (
	MenuGlobal MenuKind = iota
	MenuNick
	MenuChannel
	MenuServer
	MenuConfirm // yes/no question; Y/N (S/N in Spanish) answer it from the keyboard
	MenuTheme   // theme list; moving the cursor previews the theme
	MenuLang    // language list; moving the cursor previews the language
	MenuColor   // server color list; moving the cursor previews the color
)

// Menu is a floating popup. X, Y is the requested top-left corner.
type Menu struct {
	Kind     MenuKind
	Items    []MenuItem
	X, Y     int
	Selected int
	Ref      state.BufRef
	Nick     string
	Title    string
	Args     []string      // per-item value (theme names)
	Offset   int           // first visible item when the menu is taller than the screen
	Swatches []color.Color // optional color sample drawn before each item
}

// ConfirmMenu is a centered yes/no dialog; the first item is "yes".
func ConfirmMenu(title string, yes MenuItem, no MenuItem, ref state.BufRef, screenW, screenH int) Menu {
	m := Menu{Kind: MenuConfirm, Title: title, Ref: ref, Items: []MenuItem{yes, no}}
	r := m.Rect(screenW, screenH)
	m.X, m.Y = max((screenW-r.W)/2, 0), max((screenH-r.H)/2, 0)
	return m
}

func GlobalMenu() Menu {
	return Menu{Kind: MenuGlobal, X: 0, Y: 1, Items: []MenuItem{
		{i18n.T("Dashboard"), ActDashboard},
		{i18n.T("Connect to server"), ActConnect},
		{i18n.T("Servers"), ActServers},
		{i18n.T("Join channel…"), ActJoin},
		{i18n.T("View logs"), ActLogs},
		{i18n.T("Theme / colors"), ActTheme},
		{"Language / Idioma", ActLang},
		{i18n.T("Edit config"), ActConfig},
		{i18n.T("Help / shortcuts"), ActHelp},
		{i18n.T("Quit"), ActQuit},
	}}
}

func NickMenu(ref state.BufRef, nick string, x, y int) Menu {
	return Menu{Kind: MenuNick, Ref: ref, Nick: nick, Title: sanitize(nick), X: x, Y: y + 1, Items: []MenuItem{
		{i18n.T("Whois"), ActWhois},
		{i18n.T("Private message"), ActQuery},
		{i18n.T("Give op"), ActOp},
		{i18n.T("Take op"), ActDeop},
		{i18n.T("Give voice"), ActVoice},
		{i18n.T("Take voice"), ActDevoice},
		{i18n.T("Kick"), ActKick},
		{i18n.T("Ban"), ActBan},
	}}
}

func ChannelMenu(ref state.BufRef, x, y int) Menu {
	return Menu{Kind: MenuChannel, Ref: ref, Title: sanitize(ref.Name), X: x, Y: y + 1, Items: []MenuItem{
		{i18n.T("Leave (part)"), ActPart},
		{i18n.T("Close"), ActClose},
		{i18n.T("Mark as read"), ActMarkRead},
		{i18n.T("Show topic"), ActTopic},
	}}
}

func ServerMenu(ref state.BufRef, connected bool, x, y int) Menu {
	first := MenuItem{i18n.T("Connect"), ActConnect}
	if connected {
		first = MenuItem{i18n.T("Disconnect"), ActDisconnect}
	}
	return Menu{Kind: MenuServer, Ref: ref, Title: sanitize(ref.Network), X: x, Y: y + 1, Items: []MenuItem{first, {i18n.T("Join channel…"), ActJoin}, {i18n.T("Color…"), ActServerColor}}}
}

func (m Menu) swatchW() int {
	if len(m.Swatches) == 0 {
		return 0
	}
	return 3
}

func (m Menu) titleRows() int {
	if m.Title == "" {
		return 0
	}
	return 1
}

// visible is how many items fit on a screen screenH rows tall.
func (m Menu) visible(screenH int) int {
	return min(len(m.Items), max(screenH-2-m.titleRows(), 1))
}

func (m Menu) clipped(screenH int) bool { return m.visible(screenH) < len(m.Items) }

// Scroll moves the visible window so the selected item is on screen.
func (m *Menu) Scroll(screenH int) {
	v := m.visible(screenH)
	if m.Selected < m.Offset {
		m.Offset = m.Selected
	}
	if m.Selected >= m.Offset+v {
		m.Offset = m.Selected - v + 1
	}
	m.Offset = min(max(m.Offset, 0), len(m.Items)-v)
}

func (m Menu) title(screenH int) string {
	if m.Title != "" && m.clipped(screenH) {
		return fmt.Sprintf("%s  %d/%d", m.Title, m.Selected+1, len(m.Items))
	}
	return m.Title
}

func (m Menu) innerWidth(screenH int) int {
	w := ansi.StringWidth(m.title(screenH))
	for _, it := range m.Items {
		w = max(w, ansi.StringWidth(it.Label)+m.swatchW())
	}
	return w
}

// Rect is the on-screen box, shifted to stay inside a screenW×screenH screen.
func (m Menu) Rect(screenW, screenH int) Rect {
	r := Rect{X: m.X, Y: m.Y, W: m.innerWidth(screenH) + 4, H: m.visible(screenH) + m.titleRows() + 2}
	if r.X+r.W > screenW {
		r.X = screenW - r.W
	}
	if r.Y+r.H > screenH {
		r.Y = screenH - r.H
	}
	r.X, r.Y = max(r.X, 0), max(r.Y, 0)
	return r
}

func (m Menu) Render(screenH int) string {
	w := m.innerWidth(screenH)
	pad := func(s string) string { return s + strings.Repeat(" ", w-ansi.StringWidth(s)) }
	var rows []string
	if t := m.title(screenH); t != "" {
		rows = append(rows, menuItem.Bold(true).Render(pad(t)))
	}
	for i := m.Offset; i < m.Offset+m.visible(screenH); i++ {
		st := menuItem
		if i == m.Selected {
			st = menuSelected
		}
		if m.swatchW() == 0 {
			rows = append(rows, st.Render(pad(m.Items[i].Label)))
			continue
		}
		// the sample sits outside the item style so selection cannot hide it
		sample := "  "
		if sw := m.Swatches[i]; sw != nil {
			sample = lipgloss.NewStyle().Foreground(sw).Render("██")
		}
		label := m.Items[i].Label + strings.Repeat(" ", w-m.swatchW()-ansi.StringWidth(m.Items[i].Label))
		rows = append(rows, " "+sample+st.Render(label))
	}
	return menuBox.Render(strings.Join(rows, "\n"))
}

// ItemAt returns the item index under screen cell (x, y), or -1.
func (m Menu) ItemAt(x, y, screenW, screenH int) int {
	r := m.Rect(screenW, screenH)
	if x <= r.X || x >= r.X+r.W-1 || y <= r.Y || y >= r.Y+r.H-1 {
		return -1
	}
	if i := y - r.Y - 1 - m.titleRows(); i >= 0 {
		return m.Offset + i
	}
	return -1
}

func (m *Menu) Move(delta int) {
	n := len(m.Items)
	if n == 0 {
		return
	}
	m.Selected = ((m.Selected+delta)%n + n) % n
}

// menuH is the screen height menus may use: everything above the input line
// and the status bar, which stay visible.
func (m *Model) menuH() int { return max(m.h-2, 3) }
