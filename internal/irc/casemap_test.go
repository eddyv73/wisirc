package irc

import "testing"

func TestFold(t *testing.T) {
	if got := Fold("#Lobby"); got != "#lobby" {
		t.Fatalf("got %q", got)
	}
	if got := Fold("Nick[]\\~"); got != "nick{}|^" {
		t.Fatalf("got %q", got)
	}
}
