package ui

import (
	"strings"

	"github.com/charmbracelet/x/ansi"
)

// Overlay draws top over base with its top-left corner at display column x,
// row y. Both may contain ANSI styling and wide characters.
func Overlay(base, top string, x, y int) string {
	rows := strings.Split(base, "\n")
	tops := strings.Split(top, "\n")
	for len(rows) < y+len(tops) {
		rows = append(rows, "")
	}
	for i, t := range tops {
		row := rows[y+i]
		rw, tw := ansi.StringWidth(row), ansi.StringWidth(t)

		left := ansi.Truncate(row, x, "")
		if lw := ansi.StringWidth(left); lw < x {
			left += strings.Repeat(" ", x-lw)
		}
		right := ""
		if rw > x+tw {
			right = ansi.TruncateLeft(row, x+tw, "")
			// a wide rune cut in half on the left edge is dropped; keep alignment
			if missing := rw - (x + tw) - ansi.StringWidth(right); missing > 0 {
				right = strings.Repeat(" ", missing) + right
			}
		}
		if strings.Contains(left, "\x1b") {
			left += "\x1b[0m"
		}
		rows[y+i] = left + t + right
	}
	return strings.Join(rows, "\n")
}
