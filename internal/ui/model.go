// Package ui is the Bubble Tea interface of wisirc.
package ui

import (
	"errors"
	"sort"
	"time"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"github.com/eddyv73/wisirc/internal/config"
	"github.com/eddyv73/wisirc/internal/i18n"
	"github.com/eddyv73/wisirc/internal/irc"
	"github.com/eddyv73/wisirc/internal/logs"
	"github.com/eddyv73/wisirc/internal/notify"
	"github.com/eddyv73/wisirc/internal/state"
)

type Sender interface {
	Send(line string) error
	CurrentNick() string
}

// Networks is the connection manager the UI drives.
type Networks interface {
	Get(name string) (Sender, bool)
	Connect(arg string) (name string, err error)
	Disconnect(name string)
	Reload(cfg *config.Config) (started []string)
	Names() []string
	Config() *config.Config
	IsAdHoc(name string) bool
	SaveAdHoc(name string) error
	SaveServer(old, name string, s config.Server) error
	DeleteServer(name string) error
	Import(servers map[string]config.Server) (imported, skipped []string, err error)
}

type Deps struct {
	Cfg     *config.Config
	CfgPath string
	Nets    Networks
	Events  <-chan irc.Event
	Logger  *logs.Logger
	Notify  func(title, body string) error
	Copy    func(text string) error // system clipboard; nil or failure falls back to OSC 52

	ThemeName string                  // chosen in the menu earlier (prefs.toml); wins over Cfg.Theme
	SaveTheme func(name string) error // remembers a theme chosen in the menu
	SaveLang  func(code string) error // remembers a language chosen in the menu

	ServerColors    map[string]string            // server → "#rrggbb" chosen in its Color… menu
	SaveServerColor func(name, hex string) error // remembers it ("" = automatic)

	// Halloy reads the servers of a Halloy config; nil when none was found.
	Halloy func() (map[string]config.Server, []string, error)
	Hints  []string // startup tips for the status buffer

	StartOnDashboard bool // open the dashboard (F2) at startup
	Warnings         []string
}

type eventMsg irc.Event

type configEditedMsg struct{ err error }

// eventsMsg carries every event that was waiting, so a flood is handled in
// a few updates instead of one update (and one render) per line.
type eventsMsg []irc.Event

// maxEventBatch bounds one batch so the screen still refreshes during floods.
const maxEventBatch = 512

func waitEvent(ch <-chan irc.Event) tea.Cmd {
	return func() tea.Msg {
		ev, ok := <-ch
		if !ok {
			return nil
		}
		batch := eventsMsg{ev}
		for len(batch) < maxEventBatch {
			select {
			case ev, ok := <-ch:
				if !ok {
					return batch
				}
				batch = append(batch, ev)
			default:
				return batch
			}
		}
		return batch
	}
}

type dragKind int

const (
	dragNone dragKind = iota
	dragSidebar
	dragNick
)

// textOverlay is a read-only full-chat panel (help, logs). With typing set,
// keys go to the input line while the panel stays open (connect screen).
type textOverlay struct {
	title  string
	lines  []string
	scroll int // lines from the bottom
	typing bool
}

type completion struct {
	before string
	cands  []string
	idx    int
	start  bool
}

type Model struct {
	d  Deps
	st *state.State

	w, h            int
	sidebarW, nickW int
	theme           theme
	focused         bool

	input   textinput.Model
	history []string
	histPos int
	comp    *completion

	menu    *Menu
	overlay *textOverlay
	drag    dragKind
	sel     *selection
	admin   *adminScreen
	dash    *dashScreen
	asked   map[string]bool // ad-hoc connections already offered for saving

	themeBefore string            // theme to restore when the theme menu is cancelled
	langBefore  string            // language to restore when the language menu is cancelled
	tints       map[string]string // server → chosen color; "" or missing = automatic
	tintBefore  string            // color to restore when the color menu is cancelled
	flash       string            // short notice in the status bar, cleared on the next input
	frame       int               // animation frame for the status bar
	ticking     bool              // an animation tick is scheduled
	barHits     []barHit          // clickable spans of the status bar from the last render

	scroll     map[string]int
	sideScroll int
	nickScroll int
	seen       map[string]bool // buffers whose log history was loaded
}

func refKey(r state.BufRef) string { return r.Network + "\x00" + irc.Fold(r.Name) }

func New(d Deps) *Model {
	names := d.Nets.Names()
	sort.Strings(names)
	ti := textinput.New()
	ti.Prompt = ""
	ti.CharLimit = 0
	ti.Focus()
	m := &Model{
		d: d, st: state.New(names),
		sidebarW: DefaultSidebarW, nickW: DefaultNickW,
		focused: true, ticking: true, input: ti,
		scroll: map[string]int{}, seen: map[string]bool{}, asked: map[string]bool{}, tints: map[string]string{},
	}
	name := d.ThemeName
	if name == "" && d.Cfg != nil {
		name = d.Cfg.Theme.Name
	}
	m.setTheme(themeFor(name))
	for name, hex := range d.ServerColors {
		m.tints[name] = hex
	}
	m.addHints(d.Hints)
	if len(names) == 0 && d.Halloy != nil {
		// first run next to Halloy: the import button is one click away
		m.showServers()
		m.adminNote(i18n.T("Found your Halloy config: press [Import from Halloy] or the i key"), false)
	} else if d.StartOnDashboard && len(names) > 0 {
		m.dash = &dashScreen{}
	}
	for _, w := range d.Warnings {
		m.st.AddLine(state.BufRef{Network: m.st.Active.Network}, state.Line{Kind: state.LineError, Text: "config: " + w})
	}
	if d.Cfg != nil {
		for _, n := range names {
			if s, ok := d.Cfg.Servers[n]; ok && s.DangerouslyAcceptInvalidCerts {
				m.st.AddLine(state.BufRef{Network: n}, state.Line{Kind: state.LineError, Text: i18n.T("warning: certificate verification disabled")})
			}
		}
	}
	return m
}

// State exposes the UI state (used by tests).
func (m *Model) State() *state.State { return m.st }

// Layout returns the current screen geometry.
func (m *Model) Layout() Layout {
	l := ComputeLayout(m.w, m.h, m.sidebarW, m.nickW)
	if (m.admin != nil || m.dash != nil) && l.Nicklist.W > 0 {
		// the Servers screen uses the nicklist's columns too
		extra := l.Nicklist.W + 1
		l.Chat.W += extra
		l.Header.W += extra
		l.Nicklist, l.NickBorderX = Rect{}, -1
	}
	return l
}

func (m *Model) Init() tea.Cmd {
	return tea.Batch(waitEvent(m.d.Events), textinput.Blink, animTick())
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.w, m.h = msg.Width, msg.Height
		return m, nil
	case tea.FocusMsg:
		m.focused = true
		if !m.ticking {
			// the animation was paused while the terminal had no focus
			m.ticking = true
			return m, animTick()
		}
		return m, nil
	case tea.BlurMsg:
		m.focused = false
		return m, nil
	case eventMsg:
		cmd := m.handleEvent(irc.Event(msg))
		return m, tea.Batch(cmd, waitEvent(m.d.Events))
	case eventsMsg:
		cmds := make([]tea.Cmd, 0, len(msg)+1)
		for _, ev := range msg {
			cmds = append(cmds, m.handleEvent(ev))
		}
		return m, tea.Batch(append(cmds, waitEvent(m.d.Events))...)
	case configEditedMsg:
		m.reloadConfig(msg.err)
		return m, nil
	case tea.KeyPressMsg:
		return m, m.handleKey(msg)
	case tea.MouseClickMsg:
		return m, m.handleClick(msg.Mouse())
	case tea.MouseMotionMsg:
		m.handleMotion(msg.Mouse())
		return m, nil
	case animTickMsg:
		if !m.focused {
			m.ticking = false // nobody is looking: stop redrawing until focus returns
			return m, nil
		}
		m.frame++
		m.dashTick()
		return m, animTick()
	case tea.PasteMsg:
		if m.admin != nil && m.admin.form != nil {
			return m, m.formInput(msg)
		}
	case tea.MouseReleaseMsg:
		return m, m.handleRelease()
	case tea.MouseWheelMsg:
		m.handleWheel(msg.Mouse())
		return m, nil
	}
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m *Model) handleEvent(ev irc.Event) tea.Cmd {
	n := m.st.Network(ev.Network)
	before := map[string]bool{}
	if n != nil {
		for _, b := range n.Buffers {
			before[refKey(b.Ref)] = true
		}
	}
	prevActive := m.st.Active
	notif, appended := m.st.Apply(ev)
	if ev.Kind == irc.EvConnected {
		m.offerSave(ev.Network)
	}

	// new buffers get their log history before this event's lines are logged
	for _, b := range m.st.Network(ev.Network).Buffers {
		if !before[refKey(b.Ref)] {
			m.loadHistory(b)
		}
	}
	for _, a := range appended {
		if m.d.Logger != nil {
			m.d.Logger.Append(a.Ref.Network, a.Ref.Name, a.Line)
		}
	}

	// self JOIN while looking at that network's status: jump to the channel
	if ev.Kind == irc.EvMessage && ev.Msg.Command == "JOIN" && prevActive.Name == "" && prevActive.Network == ev.Network {
		if nw := m.st.Network(ev.Network); nw != nil && irc.Fold(ev.Msg.Nick()) == irc.Fold(nw.Nick) {
			m.switchTo(state.BufRef{Network: ev.Network, Name: ev.Msg.Param(0)})
		}
	}
	// our own JOIN: ask the channel's modes for the status bar
	if ev.Kind == irc.EvMessage && ev.Msg.Command == "JOIN" {
		if nw := m.st.Network(ev.Network); nw != nil && irc.Fold(ev.Msg.Nick()) == irc.Fold(nw.Nick) {
			if s, ok := m.d.Nets.Get(ev.Network); ok {
				s.Send(irc.Encode("MODE", ev.Msg.Param(0)))
			}
		}
	}

	if notif == nil || m.d.Notify == nil || (m.d.Cfg != nil && !m.d.Cfg.Notifications.On()) {
		return nil
	}
	if m.focused && m.st.Active == notif.Ref {
		return nil
	}
	if err := m.d.Notify("wisirc: "+notif.From, notif.Text); errors.Is(err, notify.ErrUnavailable) {
		return tea.Raw("\a")
	}
	return nil
}

// loadHistory prepends the tail of the on-disk log to a newly created buffer.
func (m *Model) loadHistory(b *state.Buffer) {
	key := refKey(b.Ref)
	if b.Kind == state.KindStatus || m.seen[key] || m.d.Logger == nil {
		return
	}
	m.seen[key] = true
	hist, err := m.d.Logger.Tail(b.Ref.Network, b.Ref.Name, 200)
	if err != nil || len(hist) == 0 {
		return
	}
	lines := make([]state.Line, 0, len(hist)+1+len(b.Lines))
	for _, h := range hist {
		lines = append(lines, state.Line{Kind: state.LineHistory, Text: h})
	}
	lines = append(lines, state.Line{Kind: state.LineInfo, Text: i18n.T("── history ──"), Time: time.Now()})
	b.Lines = append(lines, b.Lines...)
}

// ensureBuffer creates (and loads history for) a buffer opened from the UI.
func (m *Model) ensureBuffer(ref state.BufRef, kind state.BufKind) *state.Buffer {
	existed := m.st.Buffer(ref) != nil
	b := m.st.EnsureBuffer(ref, kind)
	if !existed {
		m.loadHistory(b)
	}
	return b
}

func (m *Model) switchTo(ref state.BufRef) {
	if m.st.Buffer(ref) == nil {
		return
	}
	m.st.SetActive(ref)
	m.nickScroll = 0
	m.comp = nil
}

func (m *Model) activeBuffer() *state.Buffer { return m.st.Buffer(m.st.Active) }

func (m *Model) info(text string) {
	m.st.AddLine(m.st.Active, state.Line{Kind: state.LineInfo, Text: text})
}

func (m *Model) errorLine(text string) {
	m.st.AddLine(m.st.Active, state.Line{Kind: state.LineError, Text: text})
}

func (m *Model) reloadConfig(execErr error) {
	if execErr != nil {
		m.errorLine("editor: " + execErr.Error())
	}
	cfg, warns, err := config.Load(m.d.CfgPath)
	if err != nil {
		m.errorLine(err.Error())
		return
	}
	m.d.Cfg = cfg
	for _, w := range warns {
		m.errorLine("config: " + w)
	}
	for _, name := range m.d.Nets.Reload(cfg) {
		m.st.AddNetwork(name)
		m.info(i18n.T("Connecting to %s", name))
	}
	m.info(i18n.T("Config reloaded"))
}

// offerSave asks once whether a connection made with /connect should be kept.
func (m *Model) offerSave(name string) {
	if m.asked[name] || !m.d.Nets.IsAdHoc(name) {
		return
	}
	m.asked[name] = true
	if m.menu != nil {
		m.info(i18n.T("To save %s use ≡ → Servers", name))
		return
	}
	menu := ConfirmMenu(i18n.T("Save %s permanently?", name),
		MenuItem{i18n.T("Yes, save  (Y)"), ActSaveAdHoc}, MenuItem{i18n.T("No  (N)"), ActSkipSave},
		state.BufRef{Network: name}, m.w, m.h)
	m.menu = &menu
}

// addHints shows startup tips in the status buffer.
func (m *Model) addHints(hints []string) {
	for _, h := range hints {
		m.st.AddLine(state.BufRef{Network: m.st.Active.Network}, state.Line{Kind: state.LineInfo, Text: h})
	}
}
