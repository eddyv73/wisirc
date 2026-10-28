package ui

import (
	"strings"

	"github.com/eddyv73/wisirc/internal/i18n"
)

// LangMenu lists the languages by their own name, marking the active one.
func LangMenu(x, y int) Menu {
	menu := Menu{Kind: MenuLang, X: x, Y: y, Title: "Language / Idioma"}
	for i, l := range i18n.Languages() {
		label := "  " + l.Name
		if l.Code == i18n.Lang() {
			label = "● " + l.Name
			menu.Selected = i
		}
		menu.Items = append(menu.Items, MenuItem{label, ActPickLang})
		menu.Args = append(menu.Args, l.Code)
	}
	return menu
}

func (m *Model) openLangMenu() {
	g := GlobalMenu()
	r := g.Rect(m.w, m.menuH())
	y := r.Y + 1
	for i, it := range g.Items {
		if it.Action == ActLang {
			y += i
		}
	}
	menu := LangMenu(r.X+r.W, y)
	m.langBefore = i18n.Lang()
	m.menu = &menu
}

// previewLang switches the UI to the language under the cursor.
func (m *Model) previewLang() {
	if m.menu != nil && m.menu.Kind == MenuLang {
		i18n.SetLang(m.menu.Args[m.menu.Selected])
	}
}

// chooseLang applies and remembers a language by code.
func (m *Model) chooseLang(code string) {
	code = strings.ToLower(strings.TrimSpace(code))
	if !i18n.SetLang(code) {
		var codes []string
		for _, l := range i18n.Languages() {
			codes = append(codes, l.Code)
		}
		m.errorLine(i18n.T("Unknown language: %s. Languages: %s", code, strings.Join(codes, ", ")))
		return
	}
	if m.d.SaveLang != nil {
		if err := m.d.SaveLang(code); err != nil {
			m.errorLine(i18n.T("Could not save the language: %s", err.Error()))
			return
		}
	}
	m.info(i18n.T("Language: English"))
}
