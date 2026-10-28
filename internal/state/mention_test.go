package state

import "testing"

func TestIsMention(t *testing.T) {
	yes := []string{"hola alice", "ALICE: hey", "@alice!", "alice"}
	no := []string{"alicexx", "mialice", "", "kar la"}
	for _, s := range yes {
		if !IsMention("alice", s) {
			t.Errorf("want mention in %q", s)
		}
	}
	for _, s := range no {
		if IsMention("alice", s) {
			t.Errorf("unexpected mention in %q", s)
		}
	}
	if IsMention("", "hola") {
		t.Error("empty nick")
	}
}
