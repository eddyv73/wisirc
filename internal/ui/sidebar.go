package ui

import (
	"image/color"

	"charm.land/lipgloss/v2"
	"github.com/eddyv73/wisirc/internal/state"
)

type SidebarRow struct {
	Ref       state.BufRef
	IsNetwork bool
	Label     string
}

// SidebarRows flattens networks and their buffers into sidebar lines.
func SidebarRows(s *state.State) []SidebarRow {
	var rows []SidebarRow
	for _, n := range s.Networks {
		if n.Name == "" {
			continue // home buffer of a first run without servers
		}
		glyph := "▾ "
		if n.Collapsed {
			glyph = "▸ "
		}
		rows = append(rows, SidebarRow{Ref: state.BufRef{Network: n.Name}, IsNetwork: true, Label: glyph + sanitize(n.Name)})
		if n.Collapsed {
			continue
		}
		for _, b := range n.Buffers[1:] {
			mark := "    "
			if b.Unread > 0 || b.Mentioned {
				mark = "  ● "
			}
			rows = append(rows, SidebarRow{Ref: b.Ref, Label: mark + sanitize(b.Ref.Name)})
		}
	}
	return rows
}

// serverColors tints each server in the sidebar, in sidebar order: shades of
// green that all keep dark text readable on top (contrast >= 4.5).
var serverColors = colors("#3fb950", "#2e8b57", "#87d75f", "#00af87", "#5faf5f", "#a8e6a3", "#4caf7d", "#9ccc65")

func serverColor(i int) color.Color { return serverColors[i%len(serverColors)] }

var serverText = lipgloss.Color("#0b0f0c")

// serverIndex is the position of a network among the named networks.
func serverIndex(s *state.State, name string) int {
	i := 0
	for _, n := range s.Networks {
		if n.Name == name {
			return i
		}
		if n.Name != "" {
			i++
		}
	}
	return 0
}
