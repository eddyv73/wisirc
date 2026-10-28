package ui

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/eddyv73/wisirc/internal/i18n"
)

func TestLanguageSubmenu(t *testing.T) {
	t.Cleanup(func() { i18n.SetLang("en") })
	h := newHarness(t)
	var saved []string
	h.m.d.SaveLang = func(c string) error { saved = append(saved, c); return nil }
	if !strings.Contains(strings.Join(labels(GlobalMenu()), ","), "Language / Idioma") {
		t.Fatal("global menu has no bilingual language entry")
	}
	h.m.runMenu(GlobalMenu(), ActLang)
	if h.m.menu == nil || h.m.menu.Kind != MenuLang {
		t.Fatal("no language submenu")
	}
	if v := h.view(); !strings.Contains(v, "● English") || !strings.Contains(v, "Español") {
		t.Fatalf("submenu:\n%s", v)
	}
	h.do(tea.KeyPressMsg{Code: tea.KeyDown})
	if i18n.Lang() != "es" {
		t.Fatal("no live preview")
	}
	h.do(tea.KeyPressMsg{Code: tea.KeyEscape})
	if i18n.Lang() != "en" || len(saved) != 0 {
		t.Fatalf("esc: %s saved %v", i18n.Lang(), saved)
	}
	h.m.runMenu(GlobalMenu(), ActLang)
	h.do(tea.KeyPressMsg{Code: tea.KeyDown})
	h.do(tea.KeyPressMsg{Code: tea.KeyEnter})
	if i18n.Lang() != "es" || len(saved) != 1 || saved[0] != "es" {
		t.Fatalf("enter: %s saved %v", i18n.Lang(), saved)
	}
}

func TestLangCommand(t *testing.T) {
	t.Cleanup(func() { i18n.SetLang("en") })
	h := newHarness(t)
	h.m.d.SaveLang = func(string) error { return nil }
	h.typeText("/lang es")
	h.do(tea.KeyPressMsg{Code: tea.KeyEnter})
	if i18n.Lang() != "es" {
		t.Fatal("not switched")
	}
	h.typeText("/lang xx")
	h.do(tea.KeyPressMsg{Code: tea.KeyEnter})
	if i18n.Lang() != "es" || !strings.Contains(h.view(), "en, es") {
		t.Fatalf("bad code:\n%s", h.view())
	}
}

func TestSpanishUI(t *testing.T) {
	i18n.SetLang("es")
	t.Cleanup(func() { i18n.SetLang("en") })
	h := newHarness(t)
	if v := h.view(); !strings.Contains(v, "desconectado") || !strings.Contains(v, "estado") {
		t.Fatalf("status bar not in Spanish:\n%s", v)
	}
	g := GlobalMenu()
	h.m.menu = &g
	v := h.view()
	for _, want := range []string{"Conectar servidor", "Servidores", "Tema / colores", "Language / Idioma", "Salir"} {
		if !strings.Contains(v, want) {
			t.Fatalf("missing %q in global menu:\n%s", want, v)
		}
	}
	// the language can change live: the same menu built again is English
	i18n.SetLang("en")
	if got := labels(GlobalMenu())[1]; got != "Connect to server" {
		t.Fatalf("after switching back: %q", got)
	}
}

func TestConfirmAcceptsYesInBothLanguages(t *testing.T) {
	for _, r := range []rune{'y', 'Y', 's', 'S'} {
		h := newHarness(t)
		h.nets.adhoc["x"] = true
		h.connected("x")
		h.do(key(r))
		if h.m.menu != nil || len(h.nets.savedAdh) != 1 {
			t.Fatalf("%c did not save: %v", r, h.nets.savedAdh)
		}
	}
}
