package state

import (
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/eddyv73/wisirc/internal/irc"
)

func msg(t *testing.T, line string) irc.Event {
	t.Helper()
	m, err := irc.Parse(line)
	if err != nil {
		t.Fatal(err)
	}
	return irc.Event{Network: "n", Kind: irc.EvMessage, Msg: m}
}

func connected(t *testing.T) *State {
	t.Helper()
	s := New([]string{"n"})
	ev := msg(t, ":srv 001 alice :hi")
	ev.Kind = irc.EvConnected
	s.Apply(ev)
	s.Apply(msg(t, ":alice!k@h JOIN #lobby"))
	return s
}

func TestChannelMessageAndUnread(t *testing.T) {
	s := connected(t)
	if s.Network("n").Nick != "alice" || !s.Network("n").Connected {
		t.Fatal("not connected")
	}
	n, app := s.Apply(msg(t, ":ana!a@h PRIVMSG #lobby :hola"))
	if n != nil {
		t.Fatal("unexpected notification")
	}
	b := s.Buffer(BufRef{"n", "#lobby"})
	last := b.Lines[len(b.Lines)-1]
	if last.Text != "hola" || last.Nick != "ana" || last.Kind != LineMsg {
		t.Fatalf("line %+v", last)
	}
	if b.Unread != 1 || b.Mentioned {
		t.Fatalf("unread=%d mentioned=%v", b.Unread, b.Mentioned)
	}
	if len(app) != 1 || app[0].Ref.Name != "#lobby" {
		t.Fatalf("appended %+v", app)
	}
}

func TestMentionNotifies(t *testing.T) {
	s := connected(t)
	n, _ := s.Apply(msg(t, ":ana!a@h PRIVMSG #lobby :alice ven"))
	if n == nil || n.From != "ana" || n.Text != "alice ven" {
		t.Fatalf("notification %+v", n)
	}
	b := s.Buffer(BufRef{"n", "#lobby"})
	if !b.Mentioned || !b.Lines[len(b.Lines)-1].Mention {
		t.Fatal("not marked")
	}
}

func TestOwnMessageNoMention(t *testing.T) {
	s := connected(t)
	n, _ := s.Apply(msg(t, ":alice!k@h PRIVMSG #lobby :alice soy yo"))
	if n != nil {
		t.Fatal("own message notified")
	}
	b := s.Buffer(BufRef{"n", "#lobby"})
	if !b.Lines[len(b.Lines)-1].Self {
		t.Fatal("not self")
	}
}

func TestQueryCreatesBuffer(t *testing.T) {
	s := connected(t)
	n, _ := s.Apply(msg(t, ":bob!b@h PRIVMSG alice :hi"))
	b := s.Buffer(BufRef{"n", "bob"})
	if b == nil || b.Kind != KindQuery {
		t.Fatalf("buffer %+v", b)
	}
	if n == nil {
		t.Fatal("query should notify")
	}
}

func TestCaseInsensitiveChannel(t *testing.T) {
	s := New([]string{"n"})
	ev := msg(t, ":srv 001 alice :hi")
	ev.Kind = irc.EvConnected
	s.Apply(ev)
	s.Apply(msg(t, ":alice!k@h JOIN #Lobby"))
	s.Apply(msg(t, ":ana!a@h PRIVMSG #lobby :x"))
	if got := len(s.Network("n").Buffers); got != 2 {
		t.Fatalf("buffers = %d", got)
	}
	b := s.Buffer(BufRef{"n", "#LOBBY"})
	if b == nil || b.Ref.Name != "#Lobby" {
		t.Fatalf("buffer %+v", b)
	}
}

func memberNames(ms []Member) (nicks, modes []string) {
	for _, m := range ms {
		nicks = append(nicks, m.Nick)
		modes = append(modes, m.Modes)
	}
	return
}

func TestNamesAndModes(t *testing.T) {
	s := connected(t)
	s.Apply(msg(t, ":srv 353 alice = #lobby :@alice +ana bob"))
	b := s.Buffer(BufRef{"n", "#lobby"})
	nicks, modes := memberNames(b.SortedMembers())
	if !reflect.DeepEqual(nicks, []string{"alice", "ana", "bob"}) || !reflect.DeepEqual(modes, []string{"@", "+", ""}) {
		t.Fatalf("%v %v", nicks, modes)
	}
	s.Apply(msg(t, ":alice!k@h MODE #lobby +o-v ana ana"))
	if m := b.Members[irc.Fold("ana")]; m == nil || m.Modes != "@" {
		t.Fatalf("ana %+v", m)
	}
	s.Apply(msg(t, ":ana!a@h NICK Ana2"))
	if m := b.Members[irc.Fold("Ana2")]; m == nil || m.Nick != "Ana2" || m.Modes != "@" {
		t.Fatalf("renamed %+v", m)
	}
	if b.Members["ana"] != nil {
		t.Fatal("old nick kept")
	}
	s.Apply(msg(t, ":bob!b@h QUIT :bye"))
	if b.Members["bob"] != nil {
		t.Fatal("bob not removed")
	}
	if b.Lines[len(b.Lines)-1].Kind != LineQuit {
		t.Fatal("no quit line")
	}
}

func TestNickCaseUpdate(t *testing.T) {
	s := connected(t)
	s.Apply(msg(t, ":srv 353 alice = #lobby :ana"))
	s.Apply(msg(t, ":ANA!a@h PART #lobby"))
	if s.Buffer(BufRef{"n", "#lobby"}).Members["ana"] != nil {
		t.Fatal("ana not removed")
	}
}

func TestMaxLines(t *testing.T) {
	s := connected(t)
	for i := 0; i <= MaxLines; i++ {
		s.Apply(msg(t, ":ana!a@h PRIVMSG #lobby :"+strconv.Itoa(i)))
	}
	b := s.Buffer(BufRef{"n", "#lobby"})
	if len(b.Lines) != MaxLines {
		t.Fatalf("len %d", len(b.Lines))
	}
	if b.Lines[len(b.Lines)-1].Text != strconv.Itoa(MaxLines) {
		t.Fatalf("last %q", b.Lines[len(b.Lines)-1].Text)
	}
}

func TestSetActiveClears(t *testing.T) {
	s := connected(t)
	s.Apply(msg(t, ":ana!a@h PRIVMSG #lobby :alice!"))
	ref := BufRef{"n", "#lobby"}
	s.SetActive(ref)
	b := s.Buffer(ref)
	if b.Unread != 0 || b.Mentioned || s.Active != ref {
		t.Fatal("not cleared")
	}
	s.Apply(msg(t, ":ana!a@h PRIVMSG #lobby :otra"))
	if b.Unread != 0 {
		t.Fatal("active buffer counted unread")
	}
}

func TestOrder(t *testing.T) {
	s := New([]string{"a", "b"})
	s.EnsureBuffer(BufRef{"a", "#x"}, KindChannel)
	want := []BufRef{{"a", ""}, {"a", "#x"}, {"b", ""}}
	if got := s.Order(); !reflect.DeepEqual(got, want) {
		t.Fatalf("%v", got)
	}
}

func TestAction(t *testing.T) {
	s := connected(t)
	s.Apply(msg(t, ":ana!a@h PRIVMSG #lobby :\x01ACTION baila\x01"))
	b := s.Buffer(BufRef{"n", "#lobby"})
	l := b.Lines[len(b.Lines)-1]
	if l.Kind != LineAction || l.Text != "baila" {
		t.Fatalf("%+v", l)
	}
}

func TestRemoveBuffer(t *testing.T) {
	s := connected(t)
	ref := BufRef{"n", "#lobby"}
	s.SetActive(ref)
	s.RemoveBuffer(ref)
	if s.Buffer(ref) != nil || s.Active != (BufRef{"n", ""}) {
		t.Fatal("not removed")
	}
	s.RemoveBuffer(BufRef{"n", ""})
	if s.Buffer(BufRef{"n", ""}) == nil {
		t.Fatal("status removed")
	}
}

func TestDisconnectLine(t *testing.T) {
	s := connected(t)
	s.Apply(irc.Event{Network: "n", Kind: irc.EvDisconnected, Text: "Disconnected: EOF"})
	if s.Network("n").Connected {
		t.Fatal("still connected")
	}
	st := s.Buffer(BufRef{"n", ""})
	if l := st.Lines[len(st.Lines)-1]; l.Kind != LineError || l.Text != "Disconnected: EOF" {
		t.Fatalf("%+v", l)
	}
}

func TestLagHistory(t *testing.T) {
	s := New(nil)
	for i := 1; i <= 15; i++ {
		s.Apply(irc.Event{Network: "n", Kind: irc.EvLag, Lag: time.Duration(i) * time.Millisecond})
	}
	n := s.Network("n")
	if n.Lag != 15*time.Millisecond || len(n.LagHistory) != MaxLagSamples {
		t.Fatalf("lag %v history %v", n.Lag, n.LagHistory)
	}
	if n.LagHistory[0] != 4*time.Millisecond || n.LagHistory[MaxLagSamples-1] != 15*time.Millisecond {
		t.Fatalf("history %v", n.LagHistory)
	}
}

func TestConnectingFlag(t *testing.T) {
	s := New(nil)
	s.Apply(irc.Event{Network: "n", Kind: irc.EvConnecting, Text: "Conectando"})
	if !s.Network("n").Connecting {
		t.Fatal("not connecting")
	}
	m, _ := irc.Parse(":srv 001 alice :hi")
	s.Apply(irc.Event{Network: "n", Kind: irc.EvConnected, Msg: m})
	if s.Network("n").Connecting {
		t.Fatal("still connecting after 001")
	}
	s.Apply(irc.Event{Network: "n", Kind: irc.EvConnecting})
	s.Apply(irc.Event{Network: "n", Kind: irc.EvDisconnected, Text: "x"})
	if s.Network("n").Connecting {
		t.Fatal("still connecting after disconnect")
	}
}

func TestConnectedSince(t *testing.T) {
	s := New(nil)
	at := time.Date(2026, 10, 28, 9, 0, 0, 0, time.UTC)
	m, _ := irc.Parse(":srv 001 alice :hi")
	s.Apply(irc.Event{Network: "n", Kind: irc.EvConnected, Msg: m, Time: at})
	if got := s.Network("n").ConnectedAt; !got.Equal(at) {
		t.Fatalf("connected at %v", got)
	}
	s.Apply(irc.Event{Network: "n", Kind: irc.EvDisconnected, Text: "x"})
	if !s.Network("n").ConnectedAt.IsZero() {
		t.Fatal("not cleared on disconnect")
	}
}
