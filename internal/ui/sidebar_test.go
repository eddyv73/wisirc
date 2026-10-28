package ui

import (
	"reflect"
	"testing"

	"github.com/eddyv73/wisirc/internal/state"
)

func TestSidebarRows(t *testing.T) {
	s := state.New([]string{"a", "b"})
	s.EnsureBuffer(state.BufRef{Network: "a", Name: "#x"}, state.KindChannel).Unread = 1
	s.EnsureBuffer(state.BufRef{Network: "a", Name: "#z"}, state.KindChannel)
	s.EnsureBuffer(state.BufRef{Network: "b", Name: "#y"}, state.KindChannel)
	s.Network("b").Collapsed = true
	rows := SidebarRows(s)
	var got []string
	for _, r := range rows {
		got = append(got, r.Label)
	}
	want := []string{"▾ a", "  ● #x", "    #z", "▸ b"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %q", got)
	}
	if !rows[0].IsNetwork || rows[0].Ref != (state.BufRef{Network: "a"}) || rows[1].Ref.Name != "#x" {
		t.Fatalf("refs %+v", rows)
	}
}
