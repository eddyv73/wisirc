package irc

import (
	"bufio"
	"context"
	"errors"
	"net"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

type fakeServer struct {
	t     *testing.T
	conns chan net.Conn
	dials atomic.Int32
}

func newFakeServer(t *testing.T) *fakeServer {
	return &fakeServer{t: t, conns: make(chan net.Conn, 4)}
}

func (f *fakeServer) dial(ctx context.Context) (net.Conn, error) {
	c, s := net.Pipe()
	f.dials.Add(1)
	f.conns <- s
	return c, nil
}

type peer struct {
	t *testing.T
	c net.Conn
	r *bufio.Reader
}

func (f *fakeServer) accept() *peer {
	f.t.Helper()
	select {
	case c := <-f.conns:
		return &peer{t: f.t, c: c, r: bufio.NewReader(c)}
	case <-time.After(2 * time.Second):
		f.t.Fatal("no connection")
		return nil
	}
}

func (p *peer) expect(want string) {
	p.t.Helper()
	p.c.SetReadDeadline(time.Now().Add(2 * time.Second))
	line, err := p.r.ReadString('\n')
	if err != nil {
		p.t.Fatalf("expecting %q: %v", want, err)
	}
	if got := strings.TrimRight(line, "\r\n"); got != want {
		p.t.Fatalf("got %q want %q", got, want)
	}
}

func (p *peer) expectPrefix(prefix string) {
	p.t.Helper()
	p.c.SetReadDeadline(time.Now().Add(2 * time.Second))
	line, err := p.r.ReadString('\n')
	if err != nil || !strings.HasPrefix(line, prefix) {
		p.t.Fatalf("got %q (%v) want prefix %q", line, err, prefix)
	}
}

func (p *peer) send(line string) {
	p.t.Helper()
	p.c.SetWriteDeadline(time.Now().Add(2 * time.Second))
	if _, err := p.c.Write([]byte(line + "\r\n")); err != nil {
		p.t.Fatalf("send %q: %v", line, err)
	}
}

func (p *peer) register() {
	p.t.Helper()
	p.expect("CAP LS 302")
	p.expect("NICK alice")
	p.expect("USER alice 0 * :alice")
}

func startClient(t *testing.T, f *fakeServer, mod func(*Options)) (*Client, chan Event, context.CancelFunc, chan struct{}) {
	t.Helper()
	opts := Options{
		Name: "test", Addr: "srv:6697", Nick: "alice", User: "alice", Real: "alice",
		Channels:  []string{"#lobby"},
		Dial:      f.dial,
		Backoff:   func(int) time.Duration { return time.Millisecond },
		PingEvery: time.Hour,
	}
	if mod != nil {
		mod(&opts)
	}
	events := make(chan Event, 1000)
	c := New(opts, events)
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() { c.Run(ctx); close(done) }()
	t.Cleanup(func() { cancel(); <-done })
	return c, events, cancel, done
}

func waitEvent(t *testing.T, events chan Event, pred func(Event) bool) Event {
	t.Helper()
	deadline := time.After(2 * time.Second)
	for {
		select {
		case ev := <-events:
			if pred(ev) {
				return ev
			}
		case <-deadline:
			t.Fatal("event not received")
		}
	}
}

func TestRegisterWithSASL(t *testing.T) {
	f := newFakeServer(t)
	_, events, _, _ := startClient(t, f, func(o *Options) { o.SASLUser, o.SASLPass = "alice", "hunter2" })
	p := f.accept()
	p.register()
	p.send(":srv CAP * LS :sasl server-time")
	p.expect("CAP REQ :sasl server-time")
	p.send(":srv CAP alice ACK :sasl server-time")
	p.expect("AUTHENTICATE PLAIN")
	p.send("AUTHENTICATE +")
	p.expect("AUTHENTICATE YWxpY2UAYWxpY2UAaHVudGVyMg==") // alice\0alice\0hunter2
	p.send(":srv 903 alice :ok")
	p.expect("CAP END")
	p.send(":srv 001 alice :hi")
	p.expect("JOIN #lobby")
	waitEvent(t, events, func(e Event) bool { return e.Kind == EvConnected })
}

func TestSASLFailureStopsReconnecting(t *testing.T) {
	f := newFakeServer(t)
	_, events, _, done := startClient(t, f, func(o *Options) { o.SASLUser, o.SASLPass = "alice", "bad" })
	p := f.accept()
	p.register()
	p.send(":srv CAP * LS :sasl")
	p.expect("CAP REQ :sasl")
	p.send(":srv CAP alice ACK :sasl")
	p.expect("AUTHENTICATE PLAIN")
	p.send("AUTHENTICATE +")
	p.r.ReadString('\n') // payload
	p.send(":srv 904 alice :bad password")
	p.expect("QUIT")
	ev := waitEvent(t, events, func(e Event) bool { return e.Kind == EvError })
	if !strings.Contains(ev.Text, "SASL failed") {
		t.Fatalf("text: %q", ev.Text)
	}
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("Run did not return")
	}
	if n := f.dials.Load(); n != 1 {
		t.Fatalf("dials = %d", n)
	}
}

func TestNickInUse(t *testing.T) {
	f := newFakeServer(t)
	c, events, _, _ := startClient(t, f, nil)
	p := f.accept()
	p.register()
	p.send(":srv 433 * alice :Nickname is already in use")
	p.expect("NICK alice_")
	p.send(":srv 001 alice_ :hi")
	p.expect("JOIN #lobby")
	waitEvent(t, events, func(e Event) bool { return e.Kind == EvConnected })
	if c.CurrentNick() != "alice_" {
		t.Fatalf("nick %q", c.CurrentNick())
	}
}

func TestPingPong(t *testing.T) {
	f := newFakeServer(t)
	startClient(t, f, nil)
	p := f.accept()
	p.register()
	p.send("PING :abc")
	p.expect("PONG abc")
}

func TestReconnectRejoins(t *testing.T) {
	f := newFakeServer(t)
	_, events, _, _ := startClient(t, f, nil)
	p := f.accept()
	p.register()
	p.send(":srv CAP * LS :")
	p.expect("CAP END")
	p.send(":srv 001 alice :hi")
	p.expect("JOIN #lobby")
	p.expectPrefix("PING wisirc-")
	p.send(":alice!k@h JOIN #lobby")
	p.send(":alice!k@h JOIN #otro")
	waitEvent(t, events, func(e Event) bool { return e.Kind == EvMessage && e.Msg.Param(0) == "#otro" })
	p.c.Close()
	waitEvent(t, events, func(e Event) bool { return e.Kind == EvDisconnected })

	p2 := f.accept()
	p2.register()
	p2.send(":srv CAP * LS :")
	p2.expect("CAP END")
	p2.send(":srv 001 alice :hi")
	p2.expect("JOIN #lobby,#otro")
}

func TestSendWhenDisconnected(t *testing.T) {
	c := New(Options{Nick: "alice"}, make(chan Event, 1))
	if err := c.Send("PRIVMSG #a :x"); !errors.Is(err, ErrNotConnected) {
		t.Fatalf("err = %v", err)
	}
}

func TestServerTimeTag(t *testing.T) {
	f := newFakeServer(t)
	_, events, _, _ := startClient(t, f, nil)
	p := f.accept()
	p.register()
	p.send("@time=2026-10-27T01:02:03Z :a!b@c PRIVMSG #lobby :x")
	ev := waitEvent(t, events, func(e Event) bool { return e.Kind == EvMessage && e.Msg.Command == "PRIVMSG" })
	if !ev.Time.Equal(time.Date(2026, 10, 27, 1, 2, 3, 0, time.UTC)) {
		t.Fatalf("time %v", ev.Time)
	}
}

func TestCancelEmitsDisconnected(t *testing.T) {
	f := newFakeServer(t)
	_, events, cancel, _ := startClient(t, f, nil)
	p := f.accept()
	p.register()
	cancel()
	ev := waitEvent(t, events, func(e Event) bool { return e.Kind == EvDisconnected })
	if ev.Text != "Disconnected" {
		t.Fatalf("text %q", ev.Text)
	}
}

func TestPingTimeoutReconnects(t *testing.T) {
	f := newFakeServer(t)
	_, events, _, _ := startClient(t, f, func(o *Options) { o.PingEvery = 50 * time.Millisecond })
	p := f.accept()
	go func() { // drain client writes, never answer
		for {
			if _, err := p.r.ReadString('\n'); err != nil {
				return
			}
		}
	}()
	ev := waitEvent(t, events, func(e Event) bool { return e.Kind == EvDisconnected })
	if !strings.Contains(ev.Text, "ping timeout") {
		t.Fatalf("text %q", ev.Text)
	}
	f.accept() // it redials
}
