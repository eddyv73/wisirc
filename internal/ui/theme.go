package ui

import (
	"image/color"
	"math"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/eddyv73/wisirc/internal/i18n"
)

// theme is a color scheme. bg/fg nil means the terminal's own colors; selBg
// nil means selected rows use reverse video.
type theme struct {
	name               string
	bg, fg             color.Color
	accent, dim        color.Color
	mention, unread    color.Color
	errorC, ok         color.Color
	border, menuBorder color.Color
	barFg, barBg       color.Color
	selFg, selBg       color.Color
	nicks              []color.Color
}

func colors(specs ...string) []color.Color {
	out := make([]color.Color, len(specs))
	for i, s := range specs {
		out[i] = lipgloss.Color(s)
	}
	return out
}

var c = lipgloss.Color

var ansiNicks = colors("1", "2", "3", "4", "5", "6", "9", "10", "11", "12", "13", "14")

// themeList is the menu order; the first one is the default.
var themeList = []theme{
	{
		name: "oscuro", accent: c("6"), dim: c("8"), mention: c("11"), unread: c("15"),
		errorC: c("9"), ok: c("10"), border: c("8"), menuBorder: c("6"),
		barFg: c("15"), barBg: c("236"), nicks: ansiNicks,
	},
	{
		name: "claro", accent: c("4"), dim: c("244"), mention: c("1"), unread: c("0"),
		errorC: c("1"), ok: c("2"), border: c("7"), menuBorder: c("4"),
		barFg: c("0"), barBg: c("254"), nicks: colors("1", "2", "4", "5", "6", "3", "88", "22", "94", "18", "90", "30"),
	},
	{
		name: "nocturno", bg: c("#15171c"), fg: c("#c9ced8"), accent: c("#7cc4f0"), dim: c("#7a8394"),
		mention: c("#f2c062"), unread: c("#f4f6fa"), errorC: c("#f07a7a"), ok: c("#9bd7a0"),
		border: c("#2a2f3a"), menuBorder: c("#3a4150"), barFg: c("#aab2c0"), barBg: c("#1d2027"),
		selFg: c("#e0e6ee"), selBg: c("#2b3340"),
		nicks: colors("#7cc4f0", "#9bd7a0", "#f2c062", "#e79bd8", "#f0a07a", "#8fb8ff", "#79d2c0", "#d6b4ff", "#e8d27a", "#ff9aa8", "#a8d8f0", "#c0e090"),
	},
	{
		name: "clasico", accent: c("6"), dim: c("8"), mention: c("11"), unread: c("15"),
		errorC: c("9"), ok: c("6"), border: c("4"), menuBorder: c("6"),
		barFg: c("15"), barBg: c("4"), nicks: ansiNicks,
	},
	{
		name: "alto-contraste", bg: c("#000000"), fg: c("#ffffff"), accent: c("#ffff00"), dim: c("#c0c0c0"),
		mention: c("#ffff00"), unread: c("#00ffff"), errorC: c("#ff8080"), ok: c("#80ff80"),
		border: c("#ffffff"), menuBorder: c("#ffffff"), barFg: c("#000000"), barBg: c("#ffffff"),
		selFg: c("#000000"), selBg: c("#ffffff"),
		nicks: colors("#ffffff", "#ffff00", "#00ffff", "#ff80ff", "#80ff80", "#ffb060", "#a0c0ff", "#ff9090"),
	},
	{
		name: "bosque", bg: c("#282828"), fg: c("#ebdbb2"), accent: c("#83a598"), dim: c("#a89984"),
		mention: c("#fabd2f"), unread: c("#fbf1c7"), errorC: c("#fe7060"), ok: c("#b8bb26"),
		border: c("#504945"), menuBorder: c("#83a598"), barFg: c("#ebdbb2"), barBg: c("#3c3836"),
		selFg: c("#fbf1c7"), selBg: c("#504945"),
		nicks: colors("#83a598", "#b8bb26", "#fabd2f", "#d3869b", "#8ec07c", "#fe8019", "#a9b665", "#e9b143", "#7daea3", "#ea6962", "#d8a657", "#89b482"),
	},
	{
		// dim nudged from #8f9bb3 to #949fb8 to reach 4.5:1.
		name: "glaciar", bg: c("#2e3440"), fg: c("#d8dee9"), accent: c("#88c0d0"), dim: c("#949fb8"),
		mention: c("#ebcb8b"), unread: c("#eceff4"), errorC: c("#e08088"), ok: c("#a3be8c"),
		border: c("#434c5e"), menuBorder: c("#88c0d0"), barFg: c("#e5e9f0"), barBg: c("#3b4252"),
		selFg: c("#eceff4"), selBg: c("#434c5e"),
		nicks: colors("#88c0d0", "#a3be8c", "#ebcb8b", "#c4a0c0", "#e09a80", "#8fb0d0", "#8fbcbb", "#e0a0a8", "#c8d6a0", "#a8c8f0", "#d8b8e0", "#f0d0a0"),
	},
	{
		name: "papel", bg: c("#fdf6e3"), fg: c("#33434b"), accent: c("#1d63a8"), dim: c("#5f6d73"),
		mention: c("#a34d00"), unread: c("#002b36"), errorC: c("#b8261f"), ok: c("#2d7a2d"),
		border: c("#d6cdb5"), menuBorder: c("#1d63a8"), barFg: c("#33434b"), barBg: c("#eee8d5"),
		selFg: c("#002b36"), selBg: c("#e4dcc4"),
		nicks: colors("#1d63a8", "#2d7a2d", "#a34d00", "#8a3a9a", "#b8261f", "#0f6e6e", "#5a5aa8", "#7a6200", "#a03060", "#2a6f97", "#6b5b00", "#4a7a3a"),
	},
	// Popular schemes from other terminals and editors. Where an official
	// "comment" grey or accent was too faint for text (WCAG < 4.5), it is
	// lightened (or darkened on light backgrounds); see TestThemeContrast.
	{
		// Dracula; comment #6272a4 lightened to #8e9acc.
		name: "dracula", bg: c("#282a36"), fg: c("#f8f8f2"), accent: c("#bd93f9"), dim: c("#8e9acc"),
		mention: c("#f1fa8c"), unread: c("#ffffff"), errorC: c("#ff6e6e"), ok: c("#50fa7b"),
		border: c("#44475a"), menuBorder: c("#bd93f9"), barFg: c("#f8f8f2"), barBg: c("#21222c"),
		selFg: c("#f8f8f2"), selBg: c("#44475a"),
		nicks: colors("#8be9fd", "#50fa7b", "#ffb86c", "#ff79c6", "#bd93f9", "#f1fa8c", "#ff6e6e", "#a4ffff", "#69ff94", "#ff92df", "#d6acff", "#ffffa5"),
	},
	{
		name: "catppuccin-mocha", bg: c("#1e1e2e"), fg: c("#cdd6f4"), accent: c("#cba6f7"), dim: c("#9399b2"),
		mention: c("#f9e2af"), unread: c("#ffffff"), errorC: c("#f38ba8"), ok: c("#a6e3a1"),
		border: c("#45475a"), menuBorder: c("#cba6f7"), barFg: c("#cdd6f4"), barBg: c("#181825"),
		selFg: c("#cdd6f4"), selBg: c("#45475a"),
		nicks: colors("#89b4fa", "#a6e3a1", "#fab387", "#f5c2e7", "#94e2d5", "#cba6f7", "#f9e2af", "#eba0ac", "#89dceb", "#b4befe", "#f2cdcd", "#74c7ec"),
	},
	{
		// Catppuccin Latte; subtext0, blue, green, yellow, peach, teal, sky, pink and
		// lavender darkened to read on the light base.
		name: "catppuccin-latte", bg: c("#eff1f5"), fg: c("#4c4f69"), accent: c("#8839ef"), dim: c("#62657a"),
		mention: c("#a8590a"), unread: c("#1e1e2e"), errorC: c("#d20f39"), ok: c("#2f7a1f"),
		border: c("#bcc0cc"), menuBorder: c("#8839ef"), barFg: c("#4c4f69"), barBg: c("#dce0e8"),
		selFg: c("#4c4f69"), selBg: c("#ccd0da"),
		nicks: colors("#1a5ee0", "#2f7a1f", "#b34500", "#b0378f", "#0f7377", "#8839ef", "#8a5a00", "#c0233f", "#0369a1", "#4458d8", "#b04a4a", "#13788a"),
	},
	{
		// Tokyo Night; comment #565f89 lightened to #8189b5.
		name: "tokyo-night", bg: c("#1a1b26"), fg: c("#c0caf5"), accent: c("#7aa2f7"), dim: c("#8189b5"),
		mention: c("#e0af68"), unread: c("#ffffff"), errorC: c("#f7768e"), ok: c("#9ece6a"),
		border: c("#3b4261"), menuBorder: c("#7aa2f7"), barFg: c("#a9b1d6"), barBg: c("#16161e"),
		selFg: c("#c0caf5"), selBg: c("#33467c"),
		nicks: colors("#7aa2f7", "#9ece6a", "#ff9e64", "#bb9af7", "#7dcfff", "#e0af68", "#f7768e", "#73daca", "#2ac3de", "#b4f9f8", "#c0a4ff", "#ffc777"),
	},
	{
		// Atom One Dark; comment #5c6370 lightened to #9096a1, red to #e8767e.
		name: "one-dark", bg: c("#282c34"), fg: c("#abb2bf"), accent: c("#61afef"), dim: c("#9096a1"),
		mention: c("#e5c07b"), unread: c("#ffffff"), errorC: c("#e8767e"), ok: c("#98c379"),
		border: c("#3e4451"), menuBorder: c("#61afef"), barFg: c("#abb2bf"), barBg: c("#21252b"),
		selFg: c("#e6e9ef"), selBg: c("#3e4451"),
		nicks: colors("#61afef", "#98c379", "#d19a66", "#c678dd", "#56b6c2", "#e5c07b", "#e8767e", "#8cc8ff", "#b5e08a", "#e5a1f0", "#7fd4df", "#f0a080"),
	},
	{
		// Solarized dark; base01, blue, red, magenta and violet lightened
		// since the official accents are too faint on base03.
		name: "solarized-oscuro", bg: c("#002b36"), fg: c("#93a1a1"), accent: c("#3d9be0"), dim: c("#7f9496"),
		mention: c("#c89a10"), unread: c("#eee8d5"), errorC: c("#f0605c"), ok: c("#93a800"),
		border: c("#0b4a5a"), menuBorder: c("#3d9be0"), barFg: c("#93a1a1"), barBg: c("#073642"),
		selFg: c("#eee8d5"), selBg: c("#0b4a5a"),
		nicks: colors("#3d9be0", "#93a800", "#c89a10", "#e5609a", "#2aa198", "#e86a3a", "#9093e0", "#f0605c", "#5fb8b0", "#b5c040", "#7fb4e8", "#e0a060"),
	},
	{
		// Rosé Pine; muted #6e6a86 replaced by subtle #908caa, pine lightened.
		name: "rose-pine", bg: c("#191724"), fg: c("#e0def4"), accent: c("#c4a7e7"), dim: c("#908caa"),
		mention: c("#f6c177"), unread: c("#ffffff"), errorC: c("#eb6f92"), ok: c("#9ccfd8"),
		border: c("#403d52"), menuBorder: c("#ebbcba"), barFg: c("#e0def4"), barBg: c("#1f1d2e"),
		selFg: c("#e0def4"), selBg: c("#403d52"),
		nicks: colors("#ebbcba", "#9ccfd8", "#f6c177", "#c4a7e7", "#eb6f92", "#5fa8c8", "#f2d4c4", "#b8e0e6", "#ffd9a0", "#dccaf2", "#f39ab3", "#8fc4a8"),
	},
	{
		// Kanagawa (wave); fujiGray #727169 lightened to #969588.
		name: "kanagawa", bg: c("#1f1f28"), fg: c("#dcd7ba"), accent: c("#7e9cd8"), dim: c("#969588"),
		mention: c("#e6c384"), unread: c("#ffffff"), errorC: c("#ff5d62"), ok: c("#98bb6c"),
		border: c("#363646"), menuBorder: c("#7e9cd8"), barFg: c("#c8c093"), barBg: c("#16161d"),
		selFg: c("#dcd7ba"), selBg: c("#2d4f67"),
		nicks: colors("#7e9cd8", "#98bb6c", "#ffa066", "#d27e99", "#7aa89f", "#e6c384", "#e46876", "#957fb8", "#7fb4ca", "#a3d4d5", "#c0a36e", "#b8a0e0"),
	},
	{
		// Everforest (dark medium); grey0 replaced by grey2.
		name: "everforest", bg: c("#2d353b"), fg: c("#d3c6aa"), accent: c("#a7c080"), dim: c("#9da9a0"),
		mention: c("#dbbc7f"), unread: c("#fdf6e3"), errorC: c("#f08c8e"), ok: c("#a7c080"),
		border: c("#475258"), menuBorder: c("#a7c080"), barFg: c("#d3c6aa"), barBg: c("#232a2e"),
		selFg: c("#d3c6aa"), selBg: c("#543a48"),
		nicks: colors("#7fbbb3", "#a7c080", "#e69875", "#d699b6", "#83c092", "#dbbc7f", "#e67e80", "#9fd0c8", "#c0d8a0", "#e8b0c8", "#a0d4b0", "#f0c8a0"),
	},
	{
		// Monokai; comment #75715e lightened to #a09b85, pink lightened.
		name: "monokai", bg: c("#272822"), fg: c("#f8f8f2"), accent: c("#66d9ef"), dim: c("#a09b85"),
		mention: c("#e6db74"), unread: c("#ffffff"), errorC: c("#ff5c8f"), ok: c("#a6e22e"),
		border: c("#49483e"), menuBorder: c("#f92672"), barFg: c("#f8f8f2"), barBg: c("#3e3d32"),
		selFg: c("#f8f8f2"), selBg: c("#49483e"),
		nicks: colors("#66d9ef", "#a6e22e", "#fd971f", "#ff5c8f", "#ae81ff", "#e6db74", "#a1efe4", "#c4f060", "#ffb060", "#ff8ab0", "#c8a8ff", "#f0e8a0"),
	},
	{
		// Green-phosphor CRT, like an old VT terminal.
		name: "fosforo", bg: c("#050c05"), fg: c("#33ff66"), accent: c("#9cffb0"), dim: c("#1fae45"),
		mention: c("#e0ffe0"), unread: c("#c8ffd0"), errorC: c("#d0ff40"), ok: c("#66ff99"),
		border: c("#0f4a1c"), menuBorder: c("#33ff66"), barFg: c("#050c05"), barBg: c("#33ff66"),
		selFg: c("#050c05"), selBg: c("#33ff66"),
		nicks: colors("#33ff66", "#9cffb0", "#66ff99", "#00e060", "#b0ff80", "#4fd88a", "#7dffcc", "#d0ffd8"),
	},
	{
		// Amber monochrome CRT.
		name: "ambar", bg: c("#120b00"), fg: c("#ffb000"), accent: c("#ffd060"), dim: c("#c08000"),
		mention: c("#fff0c0"), unread: c("#ffe0a0"), errorC: c("#ff7a30"), ok: c("#ffcc33"),
		border: c("#4a3000"), menuBorder: c("#ffb000"), barFg: c("#120b00"), barBg: c("#ffb000"),
		selFg: c("#120b00"), selBg: c("#ffb000"),
		nicks: colors("#ffb000", "#ffd060", "#ff9030", "#ffe090", "#e0a020", "#ffc870", "#ff8040", "#f0d8a0"),
	},
}

var themeAliases = map[string]string{
	"dark": "oscuro", "light": "claro", "": "oscuro",
	"catppuccin": "catppuccin-mocha", "tokyonight": "tokyo-night", "onedark": "one-dark",
	"rosepine": "rose-pine", "rose-pine-main": "rose-pine", "solarized-dark": "solarized-oscuro",
	"phosphor": "fosforo", "fósforo": "fosforo", "matrix": "fosforo", "amber": "ambar", "ámbar": "ambar",
}

// ThemeNames lists the built-in themes in menu order.
func ThemeNames() []string {
	out := make([]string, len(themeList))
	for i, t := range themeList {
		out[i] = t.name
	}
	return out
}

func lookupTheme(name string) (theme, bool) {
	name = strings.ToLower(strings.TrimSpace(name))
	if a, ok := themeAliases[name]; ok {
		name = a
	}
	for _, t := range themeList {
		if t.name == name {
			return t, true
		}
	}
	return theme{}, false
}

// themeFor returns the named theme, or the default one.
func themeFor(name string) theme {
	if t, ok := lookupTheme(name); ok {
		return t
	}
	return themeList[0]
}

// Styles shared by renderers that are not methods of Model; setTheme keeps
// them in step with the active theme.
var (
	nickColors   = ansiNicks
	dimStyle     lipgloss.Style
	mentionStyle lipgloss.Style
	errorStyle   lipgloss.Style
	menuBox      lipgloss.Style
	menuItem     = lipgloss.NewStyle().Padding(0, 1)
	menuSelected lipgloss.Style
	adminDim     lipgloss.Style
	adminErr     lipgloss.Style
	adminOK      lipgloss.Style
	adminButton  lipgloss.Style
	adminSel     lipgloss.Style
)

func init() { applyStyles(themeList[0]) }

func selectedStyle(t theme) lipgloss.Style {
	if t.selBg == nil {
		return lipgloss.NewStyle().Reverse(true)
	}
	return lipgloss.NewStyle().Foreground(t.selFg).Background(t.selBg)
}

func applyStyles(t theme) {
	lightIRC = lightTheme(t)
	nickColors = t.nicks
	dimStyle = lipgloss.NewStyle().Foreground(t.dim)
	mentionStyle = lipgloss.NewStyle().Bold(true).Foreground(t.mention)
	errorStyle = lipgloss.NewStyle().Foreground(t.errorC)
	menuBox = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(t.menuBorder)
	menuSelected = menuItem.Inherit(selectedStyle(t))
	adminDim = dimStyle
	adminErr = errorStyle
	adminOK = lipgloss.NewStyle().Foreground(t.ok)
	adminButton = lipgloss.NewStyle().Foreground(t.accent).Bold(true)
	adminSel = selectedStyle(t)
	lagGood, lagBad = t.ok, t.errorC
}

func (m *Model) setTheme(t theme) {
	m.theme = t
	applyStyles(t)
}

// ThemeMenu lists the themes next to the global menu, marking the current one.
func ThemeMenu(current string, x, y int) Menu {
	menu := Menu{Kind: MenuTheme, X: x, Y: y, Title: i18n.T("Theme")}
	for i, n := range ThemeNames() {
		label := "  " + n
		if n == current {
			label = "● " + n
			menu.Selected = i
		}
		menu.Items = append(menu.Items, MenuItem{label, ActPickTheme})
		menu.Args = append(menu.Args, n)
	}
	return menu
}

func (m *Model) openThemeMenu() {
	g := GlobalMenu()
	r := g.Rect(m.w, m.menuH())
	y := r.Y + 1
	for i, it := range g.Items {
		if it.Action == ActTheme {
			y += i
		}
	}
	menu := ThemeMenu(m.theme.name, r.X+r.W, y)
	m.themeBefore = m.theme.name
	m.menu = &menu
}

// previewTheme shows the theme under the cursor of the theme menu.
func (m *Model) previewTheme() {
	if m.menu != nil && m.menu.Kind == MenuTheme {
		m.setTheme(themeFor(m.menu.Args[m.menu.Selected]))
	}
}

func (m *Model) revertTheme() {
	m.setTheme(themeFor(m.themeBefore))
}

// chooseTheme applies and remembers a theme by name.
func (m *Model) chooseTheme(name string) {
	t, ok := lookupTheme(name)
	if !ok {
		m.errorLine(i18n.T("Unknown theme: %s. Themes: %s", name, strings.Join(ThemeNames(), ", ")))
		return
	}
	m.setTheme(t)
	if m.d.SaveTheme != nil {
		if err := m.d.SaveTheme(t.name); err != nil {
			m.errorLine(i18n.T("Could not save the theme: %s", err.Error()))
			return
		}
	}
	m.info(i18n.T("Theme: %s", t.name))
}

// contrastRatio is the WCAG contrast between two colors (1 to 21).
func contrastRatio(a, b color.Color) float64 {
	la, lb := relLuminance(a), relLuminance(b)
	if la < lb {
		la, lb = lb, la
	}
	return (la + 0.05) / (lb + 0.05)
}

func relLuminance(c color.Color) float64 {
	r, g, b, _ := c.RGBA()
	ch := func(v uint32) float64 {
		x := float64(v>>8) / 255
		if x <= 0.03928 {
			return x / 12.92
		}
		return math.Pow((x+0.055)/1.055, 2.4)
	}
	return 0.2126*ch(r) + 0.7152*ch(g) + 0.0722*ch(b)
}

// isHex reports whether c is an exact RGB color (not a terminal palette index).
func isHex(c color.Color) bool {
	switch c.(type) {
	case ansi.BasicColor, ansi.IndexedColor:
		return false
	}
	return c != nil
}
