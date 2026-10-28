package ui

import (
	"fmt"
	"strings"
	"sync"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/eddyv73/wisirc/internal/config"
	"github.com/eddyv73/wisirc/internal/irc"
	"github.com/eddyv73/wisirc/internal/logs"
	"github.com/eddyv73/wisirc/internal/state"
)

const netName = "examplenet"

type fakeSender struct {
	mu   sync.Mutex
	sent []string
}

func (f *fakeSender) Send(line string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.sent = append(f.sent, line)
	return nil
}
func (f *fakeSender) CurrentNick() string { return "alice" }

type fakeNets struct {
	senders  map[string]*fakeSender
	cfg      *config.Config
	adhoc    map[string]bool
	savedAdh []string
	connects []string
	none     bool // no networks at all (first run)
}

func (f *fakeNets) Get(name string) (Sender, bool) {
	s, ok := f.senders[name]
	return s, ok
}
func (f *fakeNets) Connect(arg string) (string, error) {
	f.connects = append(f.connects, arg)
	return arg, nil
}
func (f *fakeNets) Import(servers map[string]config.Server) (imported, skipped []string, err error) {
	for _, n := range []string{"examplenet", "znc", "znc-libera"} {
		s, ok := servers[n]
		if !ok {
			continue
		}
		if _, taken := f.cfg.Servers[n]; taken {
			skipped = append(skipped, n)
			continue
		}
		f.cfg.Servers[n], f.cfg.Managed[n] = s, true
		imported = append(imported, n)
	}
	return imported, skipped, nil
}
func (f *fakeNets) Disconnect(name string)             {}
func (f *fakeNets) Reload(cfg *config.Config) []string { return nil }
func (f *fakeNets) Names() []string {
	if f.none {
		return nil
	}
	return []string{netName}
}
func (f *fakeNets) sent() []string           { return f.senders[netName].sent }
func (f *fakeNets) Config() *config.Config   { return f.cfg }
func (f *fakeNets) IsAdHoc(name string) bool { return f.adhoc[name] }
func (f *fakeNets) SaveAdHoc(name string) error {
	f.savedAdh = append(f.savedAdh, name)
	delete(f.adhoc, name)
	return nil
}
func (f *fakeNets) SaveServer(old, name string, s config.Server) error {
	if _, taken := f.cfg.Servers[name]; taken && name != old {
		return fmt.Errorf("a server named %s already exists", name)
	}
	delete(f.cfg.Servers, old)
	delete(f.cfg.Managed, old)
	f.cfg.Servers[name] = s
	f.cfg.Managed[name] = true
	return nil
}
func (f *fakeNets) DeleteServer(name string) error {
	delete(f.cfg.Servers, name)
	delete(f.cfg.Managed, name)
	return nil
}

type notifyRec struct{ titles, bodies []string }

type harness struct {
	t      *testing.T
	m      *Model
	nets   *fakeNets
	notes  *notifyRec
	logger *logs.Logger
	copied []string
}

func newHarness(t *testing.T) *harness {
	t.Helper()
	rec := &notifyRec{}
	lg := &logs.Logger{Root: t.TempDir()}
	cfg := &config.Config{Servers: map[string]config.Server{netName: {Nickname: "alice", Server: "irc.example.net", Port: 6697}}, Managed: map[string]bool{}}
	nets := &fakeNets{senders: map[string]*fakeSender{netName: {}}, cfg: cfg, adhoc: map[string]bool{}}
	m := New(Deps{
		Cfg: cfg, Nets: nets, Events: make(chan irc.Event), Logger: lg,
		Notify: func(title, body string) error {
			rec.titles = append(rec.titles, title)
			rec.bodies = append(rec.bodies, body)
			return nil
		},
	})
	h := &harness{t: t, m: m, nets: nets, notes: rec, logger: lg}
	m.d.Copy = func(text string) error {
		h.copied = append(h.copied, text)
		return nil
	}
	h.send(tea.WindowSizeMsg{Width: 100, Height: 30})
	return h
}

func (h *harness) send(msg tea.Msg) {
	h.t.Helper()
	h.m.Update(msg)
}

func (h *harness) irc(line string) {
	h.t.Helper()
	m, err := irc.Parse(line)
	if err != nil {
		h.t.Fatal(err)
	}
	kind := irc.EvMessage
	if m.Command == "001" {
		kind = irc.EvConnected
	}
	h.send(eventMsg(irc.Event{Network: netName, Kind: kind, Msg: m}))
}

func (h *harness) joinLobby() {
	h.irc(":srv 001 alice :hi")
	h.irc(":alice!k@h JOIN #lobby")
}

func (h *harness) typeText(s string) {
	for _, r := range s {
		h.send(tea.KeyPressMsg{Code: r, Text: string(r)})
	}
}

func (h *harness) view() string { return ansi.Strip(h.m.View().Content) }

func click(x, y int, b tea.MouseButton) tea.MouseClickMsg {
	return tea.MouseClickMsg{X: x, Y: y, Button: b}
}

func TestTypingSendsPrivmsg(t *testing.T) {
	h := newHarness(t)
	h.joinLobby()
	if h.m.State().Active.Name != "#lobby" {
		t.Fatalf("active %+v", h.m.State().Active)
	}
	h.typeText("hola")
	h.send(tea.KeyPressMsg{Code: tea.KeyEnter})
	sent := h.nets.sent()
	if len(sent) == 0 || sent[len(sent)-1] != "PRIVMSG #lobby :hola" {
		t.Fatalf("sent %q", sent)
	}
	b := h.m.State().Buffer(state.BufRef{Network: netName, Name: "#lobby"})
	last := b.Lines[len(b.Lines)-1]
	if !last.Self || last.Text != "hola" {
		t.Fatalf("last %+v", last)
	}
}

func TestClickSidebarSwitches(t *testing.T) {
	h := newHarness(t)
	h.joinLobby()
	h.irc(":alice!k@h JOIN #otro")
	h.send(click(5, 3, tea.MouseLeft))
	if got := h.m.State().Active.Name; got != "#otro" {
		t.Fatalf("active %q", got)
	}
	h.send(click(5, 1, tea.MouseLeft))
	if got := h.m.State().Active; got != (state.BufRef{Network: netName}) {
		t.Fatalf("active %+v", got)
	}
}

func TestRightClickNickOpensMenu(t *testing.T) {
	h := newHarness(t)
	h.joinLobby()
	h.irc(":srv 353 alice = #lobby :@alice +ana")
	l := h.m.Layout()
	h.send(click(l.Nicklist.X+1, 2, tea.MouseRight))
	if !strings.Contains(h.view(), "Private message") {
		t.Fatal("menu not shown")
	}
	r := h.m.menu.Rect(100, h.m.menuH())
	h.send(click(r.X+2, r.Y+2, tea.MouseLeft)) // "Whois", below the title
	sent := h.nets.sent()
	if sent[len(sent)-1] != "WHOIS ana" {
		t.Fatalf("sent %q", sent)
	}
	if h.m.menu != nil {
		t.Fatal("menu still open")
	}
}

func TestMenuIconToggles(t *testing.T) {
	h := newHarness(t)
	h.send(click(1, 0, tea.MouseLeft))
	if !strings.Contains(h.view(), "Edit config") {
		t.Fatal("global menu not shown")
	}
	h.send(click(50, 15, tea.MouseLeft))
	if strings.Contains(h.view(), "Edit config") {
		t.Fatal("menu not closed")
	}
}

func TestF1OpensMenuKeyboard(t *testing.T) {
	h := newHarness(t)
	h.joinLobby()
	h.send(tea.KeyPressMsg{Code: tea.KeyF1})
	if h.m.menu == nil || h.m.menu.Kind != MenuGlobal {
		t.Fatal("no menu")
	}
	h.send(tea.KeyPressMsg{Code: tea.KeyDown})
	h.send(tea.KeyPressMsg{Code: tea.KeyDown})
	h.send(tea.KeyPressMsg{Code: tea.KeyDown})
	h.send(tea.KeyPressMsg{Code: tea.KeyEnter}) // "Join channel…"
	if h.m.menu != nil || h.m.input.Value() != "/join #" {
		t.Fatalf("input %q", h.m.input.Value())
	}
}

func TestDragSidebar(t *testing.T) {
	h := newHarness(t)
	h.send(click(18, 5, tea.MouseLeft))
	h.send(tea.MouseMotionMsg{X: 25, Y: 5, Button: tea.MouseLeft})
	h.send(tea.MouseReleaseMsg{X: 25, Y: 5, Button: tea.MouseLeft})
	if w := h.m.Layout().Sidebar.W; w != 25 {
		t.Fatalf("sidebar %d", w)
	}
}

func TestWheelScrolls(t *testing.T) {
	h := newHarness(t)
	h.joinLobby()
	for i := 0; i < 100; i++ {
		h.irc(fmt.Sprintf(":ana!a@h PRIVMSG #lobby :line %d", i))
	}
	if !strings.Contains(h.view(), "line 99") {
		t.Fatal("last line not visible")
	}
	h.send(tea.MouseWheelMsg{X: 40, Y: 10, Button: tea.MouseWheelUp})
	if strings.Contains(h.view(), "line 99") {
		t.Fatal("did not scroll")
	}
	h.send(tea.MouseWheelMsg{X: 40, Y: 10, Button: tea.MouseWheelDown})
	if !strings.Contains(h.view(), "line 99") {
		t.Fatal("did not scroll back")
	}
}

func TestMentionNotifies(t *testing.T) {
	h := newHarness(t)
	h.joinLobby()
	h.irc(":alice!k@h JOIN #otro")
	h.send(click(5, 2, tea.MouseLeft)) // back to #lobby
	h.irc(":ana!a@h PRIVMSG #otro :alice hola")
	if len(h.notes.titles) != 1 || h.notes.titles[0] != "wisirc: ana" || h.notes.bodies[0] != "alice hola" {
		t.Fatalf("notes %+v", h.notes)
	}
	got, _ := h.logger.Tail(netName, "#otro", 1)
	if len(got) != 1 || !strings.HasSuffix(got[0], "<ana> alice hola") {
		t.Fatalf("log %q", got)
	}
}

func TestNoNotifyWhenActiveAndFocused(t *testing.T) {
	h := newHarness(t)
	h.joinLobby()
	h.irc(":ana!a@h PRIVMSG #lobby :alice hola")
	if len(h.notes.titles) != 0 {
		t.Fatal("notified for visible buffer")
	}
	h.send(tea.BlurMsg{})
	h.irc(":ana!a@h PRIVMSG #lobby :alice otra vez")
	if len(h.notes.titles) != 1 {
		t.Fatal("not notified while unfocused")
	}
}

func TestAltNumberSwitch(t *testing.T) {
	h := newHarness(t)
	h.joinLobby()
	h.irc(":alice!k@h JOIN #otro")
	h.send(tea.KeyPressMsg{Code: '3', Mod: tea.ModAlt})
	if got := h.m.State().Active; got != h.m.State().Order()[2] {
		t.Fatalf("active %+v", got)
	}
	h.send(tea.KeyPressMsg{Code: 'p', Mod: tea.ModCtrl})
	if got := h.m.State().Active; got != h.m.State().Order()[1] {
		t.Fatalf("ctrl+p active %+v", got)
	}
}

func TestTabCompletion(t *testing.T) {
	h := newHarness(t)
	h.joinLobby()
	h.irc(":srv 353 alice = #lobby :ana anabel bob")
	h.typeText("an")
	h.send(tea.KeyPressMsg{Code: tea.KeyTab})
	if got := h.m.input.Value(); got != "ana: " {
		t.Fatalf("first tab %q", got)
	}
	h.send(tea.KeyPressMsg{Code: tea.KeyTab})
	if got := h.m.input.Value(); got != "anabel: " {
		t.Fatalf("second tab %q", got)
	}
}

func TestHistoryLoadedOnOpen(t *testing.T) {
	h := newHarness(t)
	for _, txt := range []string{"a", "b", "c"} {
		h.logger.Append(netName, "#otro", state.Line{Nick: "ana", Text: txt, Kind: state.LineMsg})
	}
	h.joinLobby()
	h.irc(":alice!k@h JOIN #otro")
	b := h.m.State().Buffer(state.BufRef{Network: netName, Name: "#otro"})
	if len(b.Lines) < 5 {
		t.Fatalf("lines %+v", b.Lines)
	}
	for i := 0; i < 3; i++ {
		if b.Lines[i].Kind != state.LineHistory {
			t.Fatalf("line %d kind %v", i, b.Lines[i].Kind)
		}
	}
	if b.Lines[3].Text != "── history ──" {
		t.Fatalf("separator %q", b.Lines[3].Text)
	}
}

func TestNarrowWindowRenders(t *testing.T) {
	h := newHarness(t)
	h.joinLobby()
	h.send(tea.WindowSizeMsg{Width: 30, Height: 6})
	if n := len(strings.Split(h.m.View().Content, "\n")); n != 6 {
		t.Fatalf("rows %d", n)
	}
	h.send(tea.WindowSizeMsg{Width: 3, Height: 1})
	h.m.View()
}

func TestQuitSendsQuit(t *testing.T) {
	h := newHarness(t)
	h.joinLobby()
	h.typeText("/quit chao")
	_, cmd := h.m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("no quit cmd")
	}
	sent := h.nets.sent()
	if sent[len(sent)-1] != "QUIT :chao" {
		t.Fatalf("sent %q", sent)
	}
}

func TestViewSettings(t *testing.T) {
	h := newHarness(t)
	v := h.m.View()
	if !v.AltScreen || v.MouseMode != tea.MouseModeCellMotion {
		t.Fatal("view flags")
	}
}

func TestMenuIconVisibleAndRowsAligned(t *testing.T) {
	h := newHarness(t)
	h.joinLobby()
	rows := strings.Split(h.m.View().Content, "\n")
	if got := ansi.Strip(ansi.Cut(rows[0], 0, 3)); got != " ≡ " {
		t.Fatalf("icon cell %q", got)
	}
	for i, r := range rows {
		if w := ansi.StringWidth(r); w != 100 {
			t.Fatalf("row %d width %d", i, w)
		}
		if i < 29 && ansi.Strip(ansi.Cut(r, 18, 19)) != "│" {
			t.Fatalf("row %d border misaligned: %q", i, ansi.Strip(r))
		}
	}
}

func TestHistoryRendersIRCCodes(t *testing.T) {
	h := newHarness(t)
	h.logger.Append(netName, "#otro", state.Line{Nick: "ana", Text: "hola \x0304rojo\x0f fin", Kind: state.LineMsg})
	h.joinLobby()
	h.irc(":alice!k@h JOIN #otro")
	h.send(click(5, 3, tea.MouseLeft))
	v := h.view()
	if strings.Contains(v, "04rojo") || !strings.Contains(v, "hola rojo fin") {
		t.Fatalf("history codes not rendered:\n%s", v)
	}
	b := h.m.State().Buffer(state.BufRef{Network: netName, Name: "#otro"})
	if b.Lines[1].Text != "── history ──" || b.Lines[1].Time.IsZero() {
		t.Fatalf("separator %+v", b.Lines[1])
	}
}

func TestConnectScreenShowsExamples(t *testing.T) {
	h := newHarness(t)
	h.typeText("/connect")
	h.send(tea.KeyPressMsg{Code: tea.KeyEnter})
	v := h.view()
	for _, want := range []string{"Connect to server", "-user", "-pass", "-sasl", "BNC"} {
		if !strings.Contains(v, want) {
			t.Fatalf("missing %q in connect screen:\n%s", want, v)
		}
	}
	if got := h.m.input.Value(); got != "/connect " {
		t.Fatalf("input %q", got)
	}
	// typing goes to the input while the examples stay visible
	h.typeText("irc.example.org")
	if got := h.m.input.Value(); got != "/connect irc.example.org" {
		t.Fatalf("input %q", got)
	}
	if !strings.Contains(h.view(), "-sasl") {
		t.Fatal("examples closed while typing")
	}
	h.send(tea.KeyPressMsg{Code: tea.KeyEnter})
	if h.m.overlay != nil {
		t.Fatal("connect screen still open after enter")
	}
	if got := h.m.State().Active.Network; got != "irc.example.org" {
		t.Fatalf("active %q", got)
	}
}

func TestConnectMenuOpensScreen(t *testing.T) {
	h := newHarness(t)
	h.m.runMenu(GlobalMenu(), ActConnect)
	if h.m.overlay == nil || !strings.Contains(h.view(), "-pass") {
		t.Fatal("menu did not open the connect screen")
	}
	h.send(tea.KeyPressMsg{Code: tea.KeyEscape})
	if h.m.overlay != nil || h.m.input.Value() != "" {
		t.Fatalf("esc: overlay %v input %q", h.m.overlay != nil, h.m.input.Value())
	}
}

func TestConnectScreenWrapsLongExamples(t *testing.T) {
	h := newHarness(t)
	h.m.showConnect()
	if v := h.view(); !strings.Contains(v, "znc-oftc") || !strings.Contains(v, "myZncPass") {
		t.Fatalf("long example cut off:\n%s", v)
	}
}

// do sends msg and runs the commands it returns, including batches.
func (h *harness) do(msg tea.Msg) {
	h.t.Helper()
	_, cmd := h.m.Update(msg)
	run(cmd)
}

func run(cmd tea.Cmd) {
	if cmd == nil {
		return
	}
	if batch, ok := cmd().(tea.BatchMsg); ok {
		for _, c := range batch {
			run(c)
		}
	}
}

// find returns the screen cell where text starts.
func (h *harness) find(text string) (int, int) {
	h.t.Helper()
	for y, line := range strings.Split(h.view(), "\n") {
		if i := strings.Index(line, text); i >= 0 {
			return ansi.StringWidth(line[:i]), y
		}
	}
	h.t.Fatalf("%q not on screen:\n%s", text, h.view())
	return 0, 0
}

func TestDragSelectsAndCopies(t *testing.T) {
	h := newHarness(t)
	h.joinLobby()
	h.irc(":ana!a@h PRIVMSG #lobby :hola mundo cruel")
	x, y := h.find("hola mundo")
	h.do(tea.MouseClickMsg{X: x, Y: y, Button: tea.MouseLeft})
	h.do(tea.MouseMotionMsg{X: x + 5, Y: y, Button: tea.MouseLeft})
	h.do(tea.MouseMotionMsg{X: x + 9, Y: y, Button: tea.MouseLeft})
	h.do(tea.MouseReleaseMsg{X: x + 9, Y: y, Button: tea.MouseLeft})
	if len(h.copied) != 1 || h.copied[0] != "hola mundo" {
		t.Fatalf("copied %q", h.copied)
	}
	if h.m.menu != nil {
		t.Fatal("drag opened a menu")
	}
	if !strings.Contains(h.view(), "Copied") {
		t.Fatal("no copy feedback in status bar")
	}
}

func TestDragAcrossRowsCopiesLines(t *testing.T) {
	h := newHarness(t)
	h.joinLobby()
	h.irc(":ana!a@h PRIVMSG #lobby :primera linea")
	h.irc(":ana!a@h PRIVMSG #lobby :segunda linea")
	x0, y0 := h.find("primera")
	x1, y1 := h.find("segunda")
	h.do(tea.MouseClickMsg{X: x0, Y: y0, Button: tea.MouseLeft})
	h.do(tea.MouseMotionMsg{X: x1 + 6, Y: y1, Button: tea.MouseLeft})
	h.do(tea.MouseReleaseMsg{X: x1 + 6, Y: y1, Button: tea.MouseLeft})
	if len(h.copied) != 1 || !strings.HasPrefix(h.copied[0], "primera linea\n") || !strings.HasSuffix(h.copied[0], "segunda") {
		t.Fatalf("copied %q", h.copied)
	}
}

func TestLeftClickChatNickOpensMenuOnRelease(t *testing.T) {
	h := newHarness(t)
	h.joinLobby()
	h.irc(":ana!a@h PRIVMSG #lobby :hola")
	x, y := h.find("ana")
	h.do(tea.MouseClickMsg{X: x, Y: y, Button: tea.MouseLeft})
	h.do(tea.MouseReleaseMsg{X: x, Y: y, Button: tea.MouseLeft})
	if h.m.menu == nil || h.m.menu.Nick != "ana" {
		t.Fatalf("menu %+v", h.m.menu)
	}
	if len(h.copied) != 0 {
		t.Fatalf("click copied %q", h.copied)
	}
}

func TestSelectInConnectScreen(t *testing.T) {
	h := newHarness(t)
	h.m.showConnect()
	x, y := h.find("/connect irc.libera.chat -nick bob -sasl")
	end := x + len("/connect irc.libera.chat -nick bob -sasl bob:myPass123")
	h.do(tea.MouseClickMsg{X: x, Y: y, Button: tea.MouseLeft})
	h.do(tea.MouseMotionMsg{X: end - 1, Y: y, Button: tea.MouseLeft})
	h.do(tea.MouseReleaseMsg{X: end - 1, Y: y, Button: tea.MouseLeft})
	if len(h.copied) != 1 || h.copied[0] != "/connect irc.libera.chat -nick bob -sasl bob:myPass123" {
		t.Fatalf("copied %q", h.copied)
	}
	if h.m.overlay == nil {
		t.Fatal("selecting closed the connect screen")
	}
}
