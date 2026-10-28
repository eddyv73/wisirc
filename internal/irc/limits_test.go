package irc

import (
	"strings"
	"testing"
)

// 1. a huge line from the server is cut, and the next line still parses
func TestOverlongLineIsTruncated(t *testing.T) {
	f := newFakeServer(t)
	_, events, _, _ := startClient(t, f, nil)
	p := f.accept()
	p.register()
	go p.send(":big!u@h PRIVMSG #lobby :" + strings.Repeat("A", 1<<20))
	ev := waitEvent(t, events, func(e Event) bool { return e.Kind == EvMessage && e.Msg.Command == "PRIVMSG" })
	if n := len(ev.Msg.Trailing()); n == 0 || n > MaxLineBytes {
		t.Fatalf("trailing kept %d bytes", n)
	}
	p.send(":ana!u@h PRIVMSG #lobby :after")
	ev = waitEvent(t, events, func(e Event) bool { return e.Kind == EvMessage && e.Msg.Command == "PRIVMSG" })
	if ev.Msg.Trailing() != "after" || ev.Msg.Nick() != "ana" {
		t.Fatalf("next line broken: %+v", ev.Msg)
	}
}
