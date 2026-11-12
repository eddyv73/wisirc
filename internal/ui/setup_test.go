package ui

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/eddyv73/wisirc/internal/config"
	"github.com/eddyv73/wisirc/internal/i18n"
)

// saved records what the setup stored in prefs.toml.
type saved struct {
	lang, theme, nick, real string
	identityCalls           int
}

// newSetupHarness is a first run: no config before, no servers yet.
func newSetupHarness(t *testing.T, edit func(*Deps)) (*harness, *saved) {
	t.Helper()
	t.Cleanup(func() { i18n.SetLang("en") })
	t.Setenv("USER", "carol")
	rec := &saved{}
	h := newHarnessWith(t, func(d *Deps, nets *fakeNets) {
		nets.none = true
		nets.cfg.Servers = map[string]config.Server{}
		d.Setup = true
		d.SaveLang = func(code string) error { rec.lang = code; return nil }
		d.SaveTheme = func(name string) error { rec.theme = name; return nil }
		d.SaveIdentity = func(nick, real string) error {
			rec.nick, rec.real = nick, real
			rec.identityCalls++
			return nil
		}
		if edit != nil {
			edit(d)
		}
	})
	return h, rec
}

func (h *harness) key(code rune) {
	h.t.Helper()
	h.do(tea.KeyPressMsg{Code: code})
}

func TestSetupWalksThroughAndConnectsLibera(t *testing.T) {
	h, rec := newSetupHarness(t, nil)
	if !strings.Contains(h.view(), "Welcome to wisirc") {
		t.Fatalf("first run should open the setup:\n%s", h.view())
	}

	// 1. language: moving the cursor shows the screen in that language
	h.key(tea.KeyDown)
	if i18n.Lang() != "es" || !strings.Contains(h.view(), "Te damos la bienvenida a wisirc") {
		t.Fatalf("language preview: %s\n%s", i18n.Lang(), h.view())
	}
	h.key(tea.KeyEnter)

	// 2. nick (suggested from the system user) and realname
	if got := h.m.setup.nick.Value(); got != "carol" {
		t.Errorf("suggested nick %q, want the system user", got)
	}
	h.m.setup.nick.SetValue("")
	h.typeText("alice")
	h.key(tea.KeyTab)
	h.typeText("Alice A.")
	h.key(tea.KeyEnter)

	// 3. theme, previewed while moving
	before := h.m.theme.name
	h.key(tea.KeyDown)
	chosen := h.m.theme.name
	if chosen == before {
		t.Fatal("moving in the theme list should preview the theme")
	}
	h.key(tea.KeyEnter)

	// 4. first server: Libera.Chat is the default without Halloy
	if !strings.Contains(h.view(), "Libera.Chat") {
		t.Fatalf("server step:\n%s", h.view())
	}
	h.key(tea.KeyEnter)

	if h.m.setup != nil {
		t.Fatal("setup should close when it finishes")
	}
	if rec.lang != "es" || rec.theme != chosen || rec.nick != "alice" || rec.real != "Alice A." {
		t.Errorf("saved %+v", rec)
	}
	s, ok := h.nets.cfg.Servers["libera"]
	if !ok || s.Nickname != "alice" || s.Username != "alice" || s.Realname != "Alice A." ||
		s.Server != "irc.libera.chat" || s.Port != 6697 || !s.TLS() || strings.Join(s.Channels, ",") != "#wisirc" {
		t.Fatalf("libera = %+v (ok %v)", s, ok)
	}
	if strings.Join(h.nets.connects, ",") != "libera" {
		t.Errorf("connects %v", h.nets.connects)
	}
	if h.m.d.Nick != "alice" {
		t.Errorf("identity not kept for later: %q", h.m.d.Nick)
	}
}

func TestSetupWithTheMouse(t *testing.T) {
	h, rec := newSetupHarness(t, nil)
	x, y := h.find("Español")
	h.do(click(x+1, y, tea.MouseLeft))
	if i18n.Lang() != "es" {
		t.Fatalf("clicking a language should pick it:\n%s", h.view())
	}
	for _, step := range []string{"[Siguiente →]", "[Siguiente →]", "[Siguiente →]"} {
		x, y = h.find(step)
		h.do(click(x+1, y, tea.MouseLeft))
	}
	x, y = h.find("Ninguno por ahora")
	h.do(click(x+1, y, tea.MouseLeft))
	x, y = h.find("[Terminar]")
	h.do(click(x+1, y, tea.MouseLeft))
	if h.m.setup != nil || rec.lang != "es" || rec.nick != "carol" {
		t.Fatalf("setup %v saved %+v", h.m.setup != nil, rec)
	}
	if len(h.nets.cfg.Servers) != 0 || len(h.nets.connects) != 0 {
		t.Errorf("no server was asked for: %v %v", h.nets.cfg.Servers, h.nets.connects)
	}
	if !strings.Contains(h.shown(), "/connect") {
		t.Errorf("should say how to add a server later:\n%s", h.shown())
	}
}

func TestSetupBackKeepsAnswers(t *testing.T) {
	h, _ := newSetupHarness(t, nil)
	h.key(tea.KeyEnter) // language
	h.m.setup.nick.SetValue("")
	h.typeText("alice")
	h.key(tea.KeyEnter)
	x, y := h.find("[← Back]")
	h.do(click(x+1, y, tea.MouseLeft))
	if h.m.setup.step != stepYou || h.m.setup.nick.Value() != "alice" {
		t.Fatalf("step %d nick %q", h.m.setup.step, h.m.setup.nick.Value())
	}
}

func TestSetupEscSkipsAndSavesNothing(t *testing.T) {
	h, rec := newSetupHarness(t, nil)
	theme := h.m.theme.name
	h.key(tea.KeyDown) // previews Spanish
	h.key(tea.KeyEscape)
	if h.m.setup != nil {
		t.Fatal("Esc should skip the setup")
	}
	if i18n.Lang() != "en" || h.m.theme.name != theme {
		t.Errorf("previews should be undone: lang %s theme %s", i18n.Lang(), h.m.theme.name)
	}
	if *rec != (saved{}) {
		t.Errorf("nothing should be saved: %+v", rec)
	}
	if !strings.Contains(h.shown(), "/connect") {
		t.Errorf("should say how to add a server later:\n%s", h.shown())
	}
}

func TestSetupRejectsABadNick(t *testing.T) {
	h, _ := newSetupHarness(t, nil)
	h.key(tea.KeyEnter)
	for _, bad := range []string{"", "two words", "9lives", "#chan"} {
		h.m.setup.nick.SetValue(bad)
		h.key(tea.KeyEnter)
		if h.m.setup.step != stepYou {
			t.Errorf("%q was accepted", bad)
		}
		if h.m.setup.err == "" {
			t.Errorf("%q: no error shown", bad)
		}
	}
	h.m.setup.nick.SetValue("al[ice]_-|^")
	h.key(tea.KeyEnter)
	if h.m.setup.step != stepTheme {
		t.Errorf("a valid IRC nick was refused: %s", h.m.setup.err)
	}
}

func TestSetupOffersHalloyFirst(t *testing.T) {
	h, rec := newSetupHarness(t, func(d *Deps) {
		d.Halloy = func() (map[string]config.Server, []string, error) {
			return map[string]config.Server{"znc": {Server: "192.0.2.10", Port: 7000}}, nil, nil
		}
	})
	for range 3 {
		h.key(tea.KeyEnter)
	}
	if !strings.Contains(h.view(), "Import from Halloy") {
		t.Fatalf("server step:\n%s", h.view())
	}
	h.key(tea.KeyEnter) // the Halloy option is the default
	if _, ok := h.nets.cfg.Servers["znc"]; !ok {
		t.Fatalf("Halloy servers not imported:\n%s", h.view())
	}
	if h.m.admin == nil {
		t.Error("the Servers screen should show what was imported")
	}
	if rec.identityCalls != 1 {
		t.Errorf("identity saved %d times", rec.identityCalls)
	}
}

func TestSetupOwnServerOpensTheFormWithTheNick(t *testing.T) {
	h, _ := newSetupHarness(t, nil)
	h.key(tea.KeyEnter)
	h.m.setup.nick.SetValue("alice")
	h.key(tea.KeyEnter)
	h.key(tea.KeyEnter)
	x, y := h.find("My own server")
	h.do(click(x+1, y, tea.MouseLeft))
	h.key(tea.KeyEnter)
	if h.m.admin == nil || h.m.admin.form == nil {
		t.Fatalf("the new-server form should open:\n%s", h.view())
	}
	if got := h.m.admin.form.value(fNick); got != "alice" {
		t.Errorf("form nick %q", got)
	}
}

// After the setup, every new server starts with the identity filled in.
func TestNewServerFormUsesIdentity(t *testing.T) {
	h := newHarnessWith(t, func(d *Deps, _ *fakeNets) { d.Nick = "alice" })
	h.m.runMenu(GlobalMenu(), ActServers)
	h.do(tea.KeyPressMsg{Code: 'n', Text: "n"})
	if h.m.admin.form == nil {
		t.Fatal("no form")
	}
	if got := h.m.admin.form.value(fNick); got != "alice" {
		t.Errorf("form nick %q", got)
	}
	if got := h.m.admin.form.value(fUser); got != "alice" {
		t.Errorf("form user %q", got)
	}
}

func TestNoSetupWhenTheConfigExists(t *testing.T) {
	h := newHarness(t)
	if h.m.setup != nil || strings.Contains(h.view(), "Welcome to wisirc") {
		t.Fatal("setup should only run the first time")
	}
}

// shown is the text of the active buffer, read from the buffer and not the
// screen so it does not depend on the window width.
func (h *harness) shown() string {
	var b strings.Builder
	for _, l := range h.m.st.Buffer(h.m.st.Active).Lines {
		b.WriteString(l.Text + "\n")
	}
	return b.String()
}
