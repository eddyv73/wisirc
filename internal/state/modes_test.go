package state

import "testing"

func TestChannelAndUserModes(t *testing.T) {
	s := connected(t)
	s.Apply(msg(t, ":alice!k@h JOIN #c"))
	b := s.Buffer(BufRef{Network: "n", Name: "#c"})
	s.Apply(msg(t, ":srv 324 alice #c +nt"))
	if b.Modes != "nt" {
		t.Fatalf("324: %q", b.Modes)
	}
	s.Apply(msg(t, ":op!o@h MODE #c +m-t"))
	if b.Modes != "mn" {
		t.Fatalf("+m-t: %q", b.Modes)
	}
	s.Apply(msg(t, ":op!o@h MODE #c +kl secret 20"))
	if b.Modes != "klmn" {
		t.Fatalf("+kl: %q", b.Modes)
	}
	s.Apply(msg(t, ":op!o@h MODE #c +o alice"))
	if b.Modes != "klmn" {
		t.Fatalf("member mode leaked into channel modes: %q", b.Modes)
	}
	s.Apply(msg(t, ":srv 324 alice #c +s"))
	if b.Modes != "s" {
		t.Fatalf("324 must replace: %q", b.Modes)
	}
	n := s.Network("n")
	s.Apply(msg(t, ":alice MODE alice :+iw"))
	if n.UserModes != "iw" {
		t.Fatalf("user modes %q", n.UserModes)
	}
	s.Apply(msg(t, ":alice MODE alice :-w"))
	if n.UserModes != "i" {
		t.Fatalf("user modes %q", n.UserModes)
	}
	s.Apply(msg(t, ":srv 221 alice +ix"))
	if n.UserModes != "ix" {
		t.Fatalf("221: %q", n.UserModes)
	}
}
