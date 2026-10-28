package ui

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/eddyv73/wisirc/internal/state"
)

const WheelStep = 3

// ChatRow is one rendered screen row. NickStart/NickEnd are display columns
// of the clickable nick, or -1.
type ChatRow struct {
	Text      string
	Nick      string
	NickStart int
	NickEnd   int
}

// wrapLine renders one buffer line into one or more rows of at most width columns.
func wrapLine(l state.Line, width int) []ChatRow {
	ts := l.Time.Local().Format("15:04") + " "
	nick := sanitize(l.Nick)
	nickStyle := lipgloss.NewStyle().Foreground(NickColor(l.Nick))
	var plainPrefix, prefix string
	start, end := -1, -1
	body := RenderIRC(l.Text)
	switch l.Kind {
	case state.LineMsg:
		plainPrefix = ts + "<" + nick + "> "
		start, end = ansi.StringWidth(ts), ansi.StringWidth(ts)+ansi.StringWidth(nick)+2
		prefix = dimStyle.Render(ts) + nickStyle.Render("<"+nick+">") + " "
	case state.LineAction:
		plainPrefix = ts + "* " + nick + " "
		start = ansi.StringWidth(ts) + 2
		end = start + ansi.StringWidth(nick)
		prefix = dimStyle.Render(ts+"* ") + nickStyle.Render(nick) + " "
	case state.LineNotice:
		plainPrefix = ts + "-" + nick + "- "
		start, end = ansi.StringWidth(ts), ansi.StringWidth(ts)+ansi.StringWidth(nick)+2
		prefix = dimStyle.Render(ts) + nickStyle.Render("-"+nick+"-") + " "
	case state.LineHistory:
		plainPrefix, prefix = "", ""
		body = dimStyle.Render(ansi.Strip(RenderIRC(l.Text)))
	case state.LineError:
		plainPrefix = ts + "-!- "
		prefix = dimStyle.Render(plainPrefix)
		body = errorStyle.Render(ansi.Strip(body))
	default:
		plainPrefix = ts + "-!- "
		prefix = dimStyle.Render(plainPrefix)
		body = dimStyle.Render(ansi.Strip(body))
	}
	if l.Mention {
		body = mentionStyle.Render(ansi.Strip(body))
	}
	pw := ansi.StringWidth(plainPrefix)
	if width <= 0 {
		return nil
	}
	avail := width - pw
	indent := pw
	if avail < 10 {
		// too narrow to hang-indent: wrap the whole line
		parts := strings.Split(ansi.Wrap(prefix+body, width, ""), "\n")
		rows := make([]ChatRow, len(parts))
		for i, p := range parts {
			rows[i] = ChatRow{Text: p, NickStart: -1, NickEnd: -1}
		}
		if end <= width {
			rows[0].Nick, rows[0].NickStart, rows[0].NickEnd = l.Nick, start, end
		}
		if start < 0 {
			rows[0].NickStart, rows[0].NickEnd = -1, -1
		}
		return rows
	}
	parts := strings.Split(ansi.Wrap(body, avail, ""), "\n")
	rows := make([]ChatRow, len(parts))
	for i, p := range parts {
		if i == 0 {
			rows[i] = ChatRow{Text: prefix + p, Nick: l.Nick, NickStart: start, NickEnd: end}
			if start < 0 {
				rows[i].Nick = ""
			}
		} else {
			rows[i] = ChatRow{Text: strings.Repeat(" ", indent) + p, NickStart: -1, NickEnd: -1}
		}
	}
	return rows
}

// WrapLines renders all lines to rows of at most width columns.
func WrapLines(lines []state.Line, width int) []ChatRow {
	var rows []ChatRow
	for _, l := range lines {
		rows = append(rows, wrapLine(l, width)...)
	}
	return rows
}

// tailRows renders only as many trailing lines as needed to fill height rows
// scrolled up by scroll. It returns the visible rows (top to bottom, at most
// height) and the scroll value clamped to the available content.
func tailRows(lines []state.Line, width, height, scroll int) ([]ChatRow, int) {
	if height <= 0 || width <= 0 {
		return nil, 0
	}
	need := height + scroll
	var rows []ChatRow
	i := len(lines) - 1
	for ; i >= 0 && len(rows) < need; i-- {
		rows = append(wrapLine(lines[i], width), rows...)
	}
	if i < 0 && len(rows) < need {
		scroll = max(len(rows)-height, 0)
	}
	endIdx := len(rows) - scroll
	startIdx := max(endIdx-height, 0)
	return rows[startIdx:endIdx], scroll
}
