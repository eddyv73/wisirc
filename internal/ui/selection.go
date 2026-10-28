package ui

import (
	"strings"
	"unicode/utf8"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/eddyv73/wisirc/internal/i18n"
)

// selection is a mouse drag over the chat panel, in screen cells; both ends
// are inclusive and it runs like terminal text, row by row.
type selection struct {
	x0, y0, x1, y1 int
	moved, done    bool
}

// span returns the selected columns of panel row y (relative to the panel).
func (s *selection) span(c Rect, y int) (from, to int, ok bool) {
	ax, ay, bx, by := s.x0, s.y0, s.x1, s.y1
	if by < ay || (by == ay && bx < ax) {
		ax, ay, bx, by = bx, by, ax, ay
	}
	sy := y + c.Y
	if sy < ay || sy > by {
		return 0, 0, false
	}
	from, to = 0, c.W
	if sy == ay {
		from = ax - c.X
	}
	if sy == by {
		to = bx - c.X + 1
	}
	return max(from, 0), min(to, c.W), from < to
}

func (m *Model) handleRelease() tea.Cmd {
	m.drag = dragNone
	s := m.sel
	if s == nil || s.done {
		return nil
	}
	if !s.moved {
		m.sel = nil
		if m.overlay == nil {
			if nick, ok := m.chatNickAt(s.x0, s.y0); ok {
				menu := NickMenu(m.st.Active, nick, s.x0, s.y0)
				m.menu = &menu
			}
		}
		return nil
	}
	s.done = true
	text := m.selectionText()
	if text == "" {
		m.sel = nil
		return nil
	}
	m.flash = i18n.T("Copied (%d characters)", utf8.RuneCountInString(text))
	return m.copyCmd(text)
}

// selectionText returns the plain text under the selection, one line per row.
func (m *Model) selectionText() string {
	l := m.Layout()
	rows := m.renderChat(l)
	var out []string
	for y, row := range rows {
		if from, to, ok := m.sel.span(l.Chat, y); ok {
			out = append(out, strings.TrimRight(ansi.Cut(ansi.Strip(row), from, to), " "))
		}
	}
	return strings.TrimRight(strings.Join(out, "\n"), "\n")
}

// highlight draws the selection over the rendered chat rows.
func (m *Model) highlight(rows []string, c Rect) []string {
	if m.sel == nil || !m.sel.moved {
		return rows
	}
	for y, row := range rows {
		if from, to, ok := m.sel.span(c, y); ok {
			mid := ansi.Strip(ansi.Cut(row, from, to))
			rows[y] = ansi.Cut(row, 0, from) + "\x1b[0;7m" + mid + "\x1b[0m" + ansi.Cut(row, to, c.W)
		}
	}
	return rows
}

// copyCmd copies with the system tool, falling back to the terminal (OSC 52).
func (m *Model) copyCmd(text string) tea.Cmd {
	copyFn := m.d.Copy
	if copyFn == nil {
		return tea.SetClipboard(text)
	}
	return func() tea.Msg {
		if err := copyFn(text); err != nil {
			return tea.SetClipboard(text)()
		}
		return nil
	}
}
