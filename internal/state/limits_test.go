package state

import (
	"fmt"
	"strings"
	"testing"
)

// 2. adding a line to a full buffer does not copy the whole buffer
func TestAddLineAtCapIsCheap(t *testing.T) {
	s := New([]string{"n"})
	ref := BufRef{Network: "n", Name: "#c"}
	for i := 0; i < MaxLines+10; i++ {
		s.AddLine(ref, Line{Text: fmt.Sprint(i)})
	}
	b := s.Buffer(ref)
	if len(b.Lines) != MaxLines || b.Lines[0].Text != "10" {
		t.Fatalf("len %d first %q", len(b.Lines), b.Lines[0].Text)
	}
	allocs := testing.AllocsPerRun(2000, func() { s.AddLine(ref, Line{Text: "x"}) })
	if allocs > 0.1 {
		t.Fatalf("%.2f allocations per line at the cap", allocs)
	}
}

// 6. a private-message flood cannot open unlimited windows
func TestQueryWindowLimit(t *testing.T) {
	s := connected(t)
	for i := 0; i < MaxQueries+20; i++ {
		s.Apply(msg(t, fmt.Sprintf(":spam%d!u@h PRIVMSG alice :buy now %d", i, i)))
	}
	queries := 0
	for _, b := range s.Network("n").Buffers {
		if b.Kind == KindQuery {
			queries++
		}
	}
	if queries != MaxQueries {
		t.Fatalf("%d query windows", queries)
	}
	status := s.Network("n").Buffers[0]
	last := status.Lines[len(status.Lines)-1]
	if last.Nick != fmt.Sprintf("spam%d", MaxQueries+19) || !strings.Contains(last.Text, "buy now") {
		t.Fatalf("overflow not in the status buffer: %+v", last)
	}
	// an existing window keeps receiving
	s.Apply(msg(t, ":spam0!u@h PRIVMSG alice :hola"))
	if b := s.Buffer(BufRef{Network: "n", Name: "spam0"}); b == nil || b.Lines[len(b.Lines)-1].Text != "hola" {
		t.Fatal("existing query lost its message")
	}
}
