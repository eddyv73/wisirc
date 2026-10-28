package ui

import (
	"image/color"

	"github.com/eddyv73/wisirc/internal/i18n"
	"github.com/eddyv73/wisirc/internal/state"
)

// serverColorChoices are the colors offered in a server's Color… menu. All
// keep the sidebar's dark selection text readable (contrast >= 4.5).
var serverColorChoices = []struct{ hex, name string }{
	{"#3fb950", "Green"},
	{"#2e8b57", "Dark green"},
	{"#87d75f", "Lime"},
	{"#00af87", "Jade"},
	{"#4fc3d9", "Cyan"},
	{"#6ea8ff", "Blue"},
	{"#b38cff", "Purple"},
	{"#ff8fc8", "Pink"},
	{"#ff6b6b", "Red"},
	{"#ff9f43", "Orange"},
	{"#ffc857", "Amber"},
	{"#f4e04d", "Yellow"},
	{"#a0a7b4", "Gray"},
}

// colorName translates a choice name (kept as literals for the catalog test).
func colorName(name string) string {
	switch name {
	case "Green":
		return i18n.T("Green")
	case "Dark green":
		return i18n.T("Dark green")
	case "Lime":
		return i18n.T("Lime")
	case "Jade":
		return i18n.T("Jade")
	case "Cyan":
		return i18n.T("Cyan")
	case "Blue":
		return i18n.T("Blue")
	case "Purple":
		return i18n.T("Purple")
	case "Pink":
		return i18n.T("Pink")
	case "Red":
		return i18n.T("Red")
	case "Orange":
		return i18n.T("Orange")
	case "Amber":
		return i18n.T("Amber")
	case "Yellow":
		return i18n.T("Yellow")
	case "Gray":
		return i18n.T("Gray")
	}
	return name
}

// serverTint is the server's chosen color, or its sidebar-order shade.
func (m *Model) serverTint(name string) color.Color {
	if hex := m.tints[name]; hex != "" {
		return c(hex)
	}
	return serverColor(serverIndex(m.st, name))
}

// channelTint is the text color of a server's channels: its tint, darkened on
// light themes so it stays readable.
func (m *Model) channelTint(name string) color.Color {
	t := m.serverTint(name)
	if !lightTheme(m.theme) {
		return t
	}
	r, g, b, _ := t.RGBA()
	return color.RGBA{uint8(r >> 8 * 45 / 100), uint8(g >> 8 * 45 / 100), uint8(b >> 8 * 45 / 100), 255}
}

func lightTheme(t theme) bool {
	if t.bg == nil {
		return t.name == "claro"
	}
	r, g, b, _ := t.bg.RGBA()
	return (299*(r>>8)+587*(g>>8)+114*(b>>8))/1000 > 128
}

func (m *Model) openColorMenu(name string, x, y int) {
	menu := Menu{Kind: MenuColor, Ref: state.BufRef{Network: name}, X: x, Y: y, Title: i18n.T("Color of %s", name)}
	current := m.tints[name]
	add := func(label, hex string, sw color.Color) {
		mark := "  "
		if hex == current {
			mark = "● "
			menu.Selected = len(menu.Items)
		}
		menu.Items = append(menu.Items, MenuItem{mark + label, ActPickColor})
		menu.Args = append(menu.Args, hex)
		menu.Swatches = append(menu.Swatches, sw)
	}
	add(i18n.T("Automatic"), "", serverColor(serverIndex(m.st, name)))
	for _, ch := range serverColorChoices {
		add(colorName(ch.name), ch.hex, c(ch.hex))
	}
	m.tintBefore = current
	m.menu = &menu
}

// previewColor tints the server with the color under the cursor.
func (m *Model) previewColor() {
	if m.menu != nil && m.menu.Kind == MenuColor {
		m.tints[m.menu.Ref.Network] = m.menu.Args[m.menu.Selected]
	}
}

func (m *Model) revertColor(name string) { m.tints[name] = m.tintBefore }

// chooseServerColor applies and remembers a server's color ("" = automatic).
func (m *Model) chooseServerColor(name, hex string) {
	m.tints[name] = hex
	if m.d.SaveServerColor != nil {
		if err := m.d.SaveServerColor(name, hex); err != nil {
			m.errorLine(i18n.T("Could not save the color: %s", err.Error()))
		}
	}
}
