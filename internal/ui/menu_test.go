package ui

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
	"github.com/eddyv73/wisirc/internal/state"
)

func labels(m Menu) []string {
	var out []string
	for _, it := range m.Items {
		out = append(out, it.Label)
	}
	return out
}

func TestGlobalMenuLabels(t *testing.T) {
	want := []string{"Dashboard", "Connect to server", "Servers", "Join channel…", "View logs", "Theme / colors", "Language / Idioma", "Edit config", "Help / shortcuts", "Quit"}
	if got := labels(GlobalMenu()); !reflect.DeepEqual(got, want) {
		t.Fatalf("got %q", got)
	}
}

func TestContextMenus(t *testing.T) {
	ref := state.BufRef{Network: "n", Name: "#lobby"}
	if got := labels(NickMenu(ref, "ana", 0, 0)); got[0] != "Whois" || got[1] != "Private message" || len(got) != 8 {
		t.Fatalf("nick %q", got)
	}
	if got := labels(ChannelMenu(ref, 0, 0)); !reflect.DeepEqual(got, []string{"Leave (part)", "Close", "Mark as read", "Show topic"}) {
		t.Fatalf("chan %q", got)
	}
	if got := ServerMenu(ref, true, 0, 0).Items[0].Action; got != ActDisconnect {
		t.Fatalf("server connected first = %v", got)
	}
	if got := ServerMenu(ref, false, 0, 0).Items[0].Action; got != ActConnect {
		t.Fatalf("server disconnected first = %v", got)
	}
}

func TestMenuRectClamped(t *testing.T) {
	m := NickMenu(state.BufRef{}, "ana", 95, 28)
	r := m.Rect(100, 30)
	if r.X < 0 || r.Y < 0 || r.X+r.W > 100 || r.Y+r.H > 30 {
		t.Fatalf("rect %+v", r)
	}
	if r.H != len(m.Items)+3 { // borders + the nick as title
		t.Fatalf("height %d", r.H)
	}
	rendered := strings.Split(m.Render(24), "\n")
	if len(rendered) != r.H || ansi.StringWidth(rendered[0]) != r.W {
		t.Fatalf("render %d×%d vs rect %+v", ansi.StringWidth(rendered[0]), len(rendered), r)
	}
}

func TestItemAt(t *testing.T) {
	m := NickMenu(state.BufRef{}, "ana", 10, 5)
	r := m.Rect(100, 30)
	if got := m.ItemAt(r.X+2, r.Y+1, 100, 30); got != -1 {
		t.Fatalf("title row %d", got)
	}
	if got := m.ItemAt(r.X+2, r.Y+2, 100, 30); got != 0 {
		t.Fatalf("first item %d", got)
	}
	if got := m.ItemAt(r.X+2, r.Y+4, 100, 30); got != 2 {
		t.Fatalf("third item %d", got)
	}
	if got := m.ItemAt(r.X, r.Y, 100, 30); got != -1 {
		t.Fatalf("border %d", got)
	}
	if got := m.ItemAt(r.X+r.W+1, r.Y+1, 100, 30); got != -1 {
		t.Fatalf("outside %d", got)
	}
}

func TestMoveWraps(t *testing.T) {
	m := GlobalMenu()
	m.Move(-1)
	if m.Selected != len(m.Items)-1 {
		t.Fatalf("selected %d", m.Selected)
	}
	m.Move(1)
	if m.Selected != 0 {
		t.Fatalf("selected %d", m.Selected)
	}
}

func TestMenuScrollsWhenTallerThanScreen(t *testing.T) {
	m := ThemeMenu("oscuro", 0, 0)
	n := len(m.Items)
	m.Selected = n - 1
	m.Scroll(12)
	r := m.Rect(80, 12)
	if r.H != 12 {
		t.Fatalf("height %d", r.H)
	}
	out := ansi.Strip(m.Render(12))
	if strings.Count(out, "\n")+1 != 12 || !strings.Contains(out, m.Args[n-1]) || strings.Contains(out, " oscuro ") {
		t.Fatalf("window wrong:\n%s", out)
	}
	if !strings.Contains(out, fmt.Sprintf("%d/%d", n, n)) {
		t.Fatalf("no position indicator:\n%s", out)
	}
	if got := m.ItemAt(2, r.Y+r.H-2, 80, 12); got != n-1 {
		t.Fatalf("ItemAt last row = %d", got)
	}
	m.Selected = 0
	m.Scroll(12)
	if out := ansi.Strip(m.Render(12)); !strings.Contains(out, "oscuro") {
		t.Fatalf("did not scroll back:\n%s", out)
	}
}
