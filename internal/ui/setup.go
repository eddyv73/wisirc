package ui

import (
	"os"
	"regexp"
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/eddyv73/wisirc/internal/config"
	"github.com/eddyv73/wisirc/internal/i18n"
)

// The first-run setup: four short steps shown in the chat panel the first
// time wisirc runs (no config.toml yet). Esc skips it and saves nothing.
const (
	stepLang = iota
	stepYou
	stepTheme
	stepServer
	stepCount
)

// First-server choices of the last step.
const (
	firstLibera = "libera"
	firstHalloy = "halloy"
	firstOwn    = "own"
	firstNone   = "none"
)

type setupScreen struct {
	step int
	sel  int // cursor in the list of the language, theme and server steps
	// what each list step had chosen, so Back returns to it
	lang, themeName, server string

	nick, real textinput.Model
	focus      int // 0 nick, 1 realname
	err        string

	langBefore, themeBefore string // restored when the setup is skipped
	hits                    []adminHit
}

// validNick follows RFC 2812's nickname rule, with a generous length.
var validNick = regexp.MustCompile(`^[A-Za-z\[\]\\` + "`" + `_^{|}][A-Za-z0-9\[\]\\` + "`" + `_^{|}-]{0,29}$`)

// suggestNick turns the system user into a nick, or "" when it cannot.
func suggestNick(user string) string {
	var b strings.Builder
	for _, r := range user {
		if r < 0x80 && validNick.MatchString("a"+string(r)) {
			b.WriteRune(r)
		}
	}
	n := strings.TrimLeft(b.String(), "0123456789-")
	if len(n) > 30 {
		n = n[:30]
	}
	if !validNick.MatchString(n) {
		return ""
	}
	return n
}

func newSetup(d Deps, themeName string) *setupScreen {
	field := func(value, placeholder string, limit int) textinput.Model {
		ti := textinput.New()
		ti.Prompt = ""
		ti.CharLimit = limit
		ti.Placeholder = placeholder
		ti.SetValue(value)
		return ti
	}
	nick := d.Nick
	if nick == "" {
		nick = suggestNick(os.Getenv("USER"))
	}
	if nick == "" {
		nick = suggestNick(os.Getenv("USERNAME"))
	}
	s := &setupScreen{
		lang: i18n.Lang(), themeName: themeName, server: firstLibera,
		nick: field(nick, "alice", 30), real: field(d.Realname, "Alice Liddell", 64),
		langBefore: i18n.Lang(), themeBefore: themeName,
	}
	if d.Halloy != nil {
		s.server = firstHalloy
	}
	s.enter(stepLang, d)
	return s
}

// options lists the codes of a list step and their labels.
func (s *setupScreen) options(d Deps) (codes, labels []string) {
	switch s.step {
	case stepLang:
		for _, l := range i18n.Languages() {
			codes, labels = append(codes, l.Code), append(labels, l.Name)
		}
	case stepTheme:
		codes = ThemeNames()
		labels = codes
	case stepServer:
		if d.Halloy != nil {
			codes, labels = append(codes, firstHalloy), append(labels, i18n.T("Import from Halloy (found on this computer)"))
		}
		codes = append(codes, firstLibera, firstOwn, firstNone)
		labels = append(labels,
			i18n.T("Libera.Chat, joining #wisirc"),
			i18n.T("My own server (address, ZNC, password…)"),
			i18n.T("None for now"))
	}
	return codes, labels
}

// enter shows a step with its cursor on what was chosen there.
func (s *setupScreen) enter(step int, d Deps) {
	s.step, s.err, s.sel = step, "", 0
	chosen := map[int]string{stepLang: s.lang, stepTheme: s.themeName, stepServer: s.server}[step]
	codes, _ := s.options(d)
	for i, c := range codes {
		if c == chosen {
			s.sel = i
		}
	}
	s.setFocus(0)
}

func (s *setupScreen) setFocus(i int) {
	s.focus = i
	s.nick.Blur()
	s.real.Blur()
	if s.step != stepYou {
		return
	}
	if i == 0 {
		s.nick.Focus()
	} else {
		s.real.Focus()
	}
}

// pick moves the cursor of a list step and previews the choice.
func (m *Model) setupPick(i int) {
	s := m.setup
	codes, _ := s.options(m.d)
	if len(codes) == 0 {
		return
	}
	s.sel = (i + len(codes)) % len(codes)
	switch s.step {
	case stepLang:
		s.lang = codes[s.sel]
		i18n.SetLang(s.lang)
	case stepTheme:
		s.themeName = codes[s.sel]
		m.setTheme(themeFor(s.themeName))
	case stepServer:
		s.server = codes[s.sel]
	}
}

// setupNext checks the step and moves on; on the last step it finishes.
func (m *Model) setupNext() tea.Cmd {
	s := m.setup
	if s.step == stepYou {
		nick := strings.TrimSpace(s.nick.Value())
		if !validNick.MatchString(nick) {
			s.err = i18n.T("A nick is one word: letters, digits and - _ [ ] { } | ^ `, not starting with a digit")
			s.setFocus(0)
			return nil
		}
		s.nick.SetValue(nick)
	}
	if s.step+1 < stepCount {
		s.enter(s.step+1, m.d)
		return nil
	}
	return m.finishSetup()
}

func (m *Model) setupBack() {
	if s := m.setup; s.step > 0 {
		s.enter(s.step-1, m.d)
	}
}

// skipSetup closes the setup without saving, undoing its previews.
func (m *Model) skipSetup() {
	s := m.setup
	i18n.SetLang(s.langBefore)
	m.setTheme(themeFor(s.themeBefore))
	m.setup = nil
	m.info(i18n.T("Setup skipped. Add a server any time: ≡ → Servers, or /connect irc.libera.chat"))
}

// finishSetup saves the answers and starts the chosen first server.
func (m *Model) finishSetup() tea.Cmd {
	s := m.setup
	nick, real := s.nick.Value(), strings.TrimSpace(s.real.Value())
	m.setup = nil
	save := func(what string, err error) {
		if err != nil {
			m.errorLine(i18n.T("Could not save the %s: %s", what, err.Error()))
		}
	}
	if m.d.SaveLang != nil {
		save(i18n.T("language"), m.d.SaveLang(s.lang))
	}
	m.d.ThemeName = s.themeName
	if m.d.SaveTheme != nil {
		save(i18n.T("theme"), m.d.SaveTheme(s.themeName))
	}
	m.d.Nick, m.d.Realname = nick, real
	if m.d.SaveIdentity != nil {
		save(i18n.T("nick"), m.d.SaveIdentity(nick, real))
	}
	m.info(i18n.T("Welcome, %s! Change language, theme and servers any time from ≡ (F1).", nick))

	switch s.server {
	case firstLibera:
		srv := config.Server{
			Nickname: nick, Username: nick, Realname: real,
			Server: "irc.libera.chat", Port: 6697, Channels: []string{"#wisirc"},
		}
		if err := m.d.Nets.SaveServer("", "libera", srv); err != nil {
			m.errorLine(err.Error())
			return nil
		}
		m.connect("libera")
	case firstHalloy:
		m.showServers()
		m.importHalloy()
	case firstOwn:
		m.showServers()
		m.admin.form = newServerForm("", m.newServer())
		m.admin.form.setFocus(fName)
	default:
		m.info(i18n.T("Add a server any time: ≡ → Servers, or /connect irc.libera.chat"))
	}
	return nil
}

func (m *Model) setupKey(msg tea.KeyPressMsg) tea.Cmd {
	s := m.setup
	switch msg.String() {
	case "ctrl+c":
		return m.quit("")
	case "esc":
		m.skipSetup()
		return nil
	case "enter":
		return m.setupNext()
	case "shift+tab":
		if s.step == stepYou && s.focus == 1 {
			s.setFocus(0)
			return nil
		}
		m.setupBack()
		return nil
	}
	if s.step == stepYou {
		switch msg.String() {
		case "tab", "down":
			s.setFocus(1)
			return nil
		case "up":
			s.setFocus(0)
			return nil
		}
		return m.setupInput(msg)
	}
	switch msg.String() {
	case "up", "k":
		m.setupPick(s.sel - 1)
	case "down", "j", "tab":
		m.setupPick(s.sel + 1)
	case "left", "backspace":
		m.setupBack()
	case "right":
		return m.setupNext()
	}
	return nil
}

// setupInput sends typing (and pastes) to the focused field of step 2.
func (m *Model) setupInput(msg tea.Msg) tea.Cmd {
	s := m.setup
	if s.step != stepYou {
		return nil
	}
	s.err = ""
	var cmd tea.Cmd
	if s.focus == 0 {
		s.nick, cmd = s.nick.Update(msg)
	} else {
		s.real, cmd = s.real.Update(msg)
	}
	return cmd
}

func (m *Model) setupClick(x, y int) tea.Cmd {
	s := m.setup
	c := m.Layout().Chat
	px, py := x-c.X, y-c.Y
	for _, h := range s.hits {
		if h.y != py || px < h.x0 || px >= h.x1 {
			continue
		}
		switch h.act {
		case "option":
			m.setupPick(h.index)
		case "field":
			s.setFocus(h.index)
		case "next":
			return m.setupNext()
		case "back":
			m.setupBack()
		case "skip":
			m.skipSetup()
		}
		return nil
	}
	return nil
}

// moreLine marks hidden options above or below a scrolled list; it stays a
// blank line when there are none, so the list does not jump.
func moreLine(n int, text string) string {
	if n <= 0 {
		return ""
	}
	return adminDim.Render(text)
}

// renderSetup draws the current step into the chat panel.
func (m *Model) renderSetup(w, h int) []string {
	s := m.setup
	s.hits = nil
	var lines []string
	add := func(line string) { lines = append(lines, line) }
	bold := lipgloss.NewStyle().Bold(true)
	accent := lipgloss.NewStyle().Bold(true).Foreground(m.theme.accent)

	add("")
	add(" " + accent.Render(i18n.T("Welcome to wisirc")) + adminDim.Render(i18n.T("  ·  step %d of %d", s.step+1, stepCount)))
	add("")
	titles := []string{
		i18n.T("Language"),
		i18n.T("Who are you on IRC?"),
		i18n.T("Pick a look"),
		i18n.T("Where do you want to start?"),
	}
	hints := []string{
		i18n.T("↑/↓ to choose; the screen changes as you move."),
		i18n.T("Your nick is how people see you. The real name is optional and public."),
		i18n.T("↑/↓ to preview each theme."),
		i18n.T("You can add more servers later from ≡ → Servers."),
	}
	add(" " + bold.Render(titles[s.step]))
	add(adminDim.Render(" " + hints[s.step]))
	add("")

	if s.step == stepYou {
		row := func(i int, label string, in textinput.Model) {
			prefix := "   "
			if s.focus == i {
				prefix = " ▸ "
			}
			in.SetWidth(max(min(w-17, 40), 1)) // without a width the placeholder shows one letter
			text := prefix + fit(label, 12) + " " + in.View()
			s.hits = append(s.hits, adminHit{y: len(lines), x0: 0, x1: w, act: "field", index: i})
			add(text)
		}
		row(0, i18n.T("Nick"), s.nick)
		row(1, i18n.T("Real name"), s.real)
	} else {
		_, labels := s.options(m.d)
		// long lists (the themes) scroll around the cursor so the buttons
		// below always stay on screen
		room := max(h-len(lines)-6, 3)
		first, last := 0, len(labels)
		if len(labels) > room {
			room -= 2 // the "more" markers
			first = min(max(s.sel-room/2, 0), len(labels)-room)
			last = first + room
			add(moreLine(first, i18n.T("   ↑ %d more", first)))
		}
		for i := first; i < last; i++ {
			label := labels[i]
			mark := "( )"
			if i == s.sel {
				mark = "(•)"
			}
			text := "   " + mark + " " + label
			if i == s.sel {
				text = adminSel.Render(fit(text, min(w, ansi.StringWidth(text)+2)))
			}
			s.hits = append(s.hits, adminHit{y: len(lines), x0: 0, x1: w, act: "option", index: i})
			add(text)
		}
		if last < len(labels) || first > 0 {
			add(moreLine(len(labels)-last, i18n.T("   ↓ %d more", len(labels)-last)))
		}
	}
	add("")
	if s.err != "" {
		add(adminErr.Render(" " + s.err))
		add("")
	}

	next := i18n.T("Next →")
	if s.step == stepCount-1 {
		next = i18n.T("Finish")
	}
	btns := [][2]string{{next, "next"}}
	if s.step > 0 {
		btns = append([][2]string{{i18n.T("← Back"), "back"}}, btns...)
	}
	btns = append(btns, [2]string{i18n.T("Skip setup"), "skip"})
	bar := &adminScreen{}
	row := bar.buttons(len(lines), btns)
	s.hits = append(s.hits, bar.hits...)
	add(row)
	add("")
	add(adminDim.Render(i18n.T(" Enter next · Shift+Tab back · Esc skip")))

	for i := range lines {
		lines[i] = fit(lines[i], w)
	}
	for len(lines) < h {
		lines = append(lines, strings.Repeat(" ", max(w, 0)))
	}
	return lines[:max(h, 0)]
}
