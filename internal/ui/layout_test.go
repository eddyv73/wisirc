package ui

import "testing"

func TestLayoutNormal(t *testing.T) {
	l := ComputeLayout(100, 30, 18, 14)
	checks := map[string][2]Rect{
		"sidebar":  {l.Sidebar, {0, 0, 18, 29}},
		"header":   {l.Header, {19, 0, 66, 1}},
		"chat":     {l.Chat, {19, 1, 66, 27}},
		"nicklist": {l.Nicklist, {86, 0, 14, 28}},
		"input":    {l.Input, {19, 28, 81, 1}},
		"status":   {l.Status, {0, 29, 100, 1}},
		"menu":     {l.MenuIcon, {0, 0, 3, 1}},
	}
	for name, c := range checks {
		if c[0] != c[1] {
			t.Errorf("%s = %+v want %+v", name, c[0], c[1])
		}
	}
	if l.SidebarBorderX != 18 || l.NickBorderX != 85 {
		t.Errorf("borders %d %d", l.SidebarBorderX, l.NickBorderX)
	}
}

func TestHitZones(t *testing.T) {
	l := ComputeLayout(100, 30, 18, 14)
	cases := []struct {
		x, y int
		want Zone
	}{
		{1, 0, ZoneMenuIcon}, {5, 5, ZoneSidebar}, {18, 5, ZoneSidebarBorder}, {40, 10, ZoneChat},
		{40, 0, ZoneHeader}, {85, 3, ZoneNickBorder}, {90, 3, ZoneNicklist}, {40, 28, ZoneInput},
		{40, 29, ZoneStatus}, {-1, 0, ZoneNone}, {100, 0, ZoneNone},
	}
	for _, c := range cases {
		if got := l.Hit(c.x, c.y); got != c.want {
			t.Errorf("Hit(%d,%d) = %v want %v", c.x, c.y, got, c.want)
		}
	}
}

func TestLayoutNarrow(t *testing.T) {
	l := ComputeLayout(50, 20, 18, 14)
	if l.Nicklist.W != 0 || l.NickBorderX != -1 {
		t.Fatalf("nicklist %+v border %d", l.Nicklist, l.NickBorderX)
	}
	l = ComputeLayout(30, 20, 18, 14)
	if l.Sidebar.W != 0 || l.SidebarBorderX != -1 {
		t.Fatalf("sidebar %+v", l.Sidebar)
	}
	if l.Hit(1, 0) != ZoneMenuIcon {
		t.Fatal("menu icon missing")
	}
	if l.Header.X != 3 {
		t.Fatalf("header %+v", l.Header)
	}
}

func TestLayoutTiny(t *testing.T) {
	for w := 0; w <= 20; w++ {
		for h := 0; h <= 6; h++ {
			l := ComputeLayout(w, h, 18, 14)
			for _, r := range []Rect{l.MenuIcon, l.Sidebar, l.Header, l.Chat, l.Nicklist, l.Input, l.Status} {
				if r.W < 0 || r.H < 0 {
					t.Fatalf("w=%d h=%d negative rect %+v", w, h, r)
				}
			}
			for x := -1; x <= w; x++ {
				for y := -1; y <= h; y++ {
					l.Hit(x, y)
				}
			}
		}
	}
}

func TestClamp(t *testing.T) {
	l := ComputeLayout(100, 30, 80, 80)
	if l.Sidebar.W != 33 || l.Nicklist.W != 25 {
		t.Fatalf("sidebar %d nick %d", l.Sidebar.W, l.Nicklist.W)
	}
	l = ComputeLayout(100, 30, 1, 1)
	if l.Sidebar.W != 12 || l.Nicklist.W != 10 {
		t.Fatalf("sidebar %d nick %d", l.Sidebar.W, l.Nicklist.W)
	}
}
