package commands

import (
	"reflect"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/eddyv73/wisirc/internal/state"
)

var chanCtx = Context{Target: "#lobby", Kind: state.KindChannel, Nick: "alice"}

func TestParse(t *testing.T) {
	cases := []struct {
		in    string
		ctx   Context
		kind  Kind
		lines []string
		check func(t *testing.T, a Action)
	}{
		{in: "hola", kind: Send, lines: []string{"PRIVMSG #lobby :hola"}, check: func(t *testing.T, a Action) {
			if !reflect.DeepEqual(a.Echo, []string{"hola"}) || a.Target != "#lobby" {
				t.Errorf("echo %v target %q", a.Echo, a.Target)
			}
		}},
		{in: "hola", ctx: Context{Kind: state.KindStatus, Nick: "alice"}, kind: Error, check: func(t *testing.T, a Action) {
			if a.Err != "You are not in a channel" {
				t.Errorf("err %q", a.Err)
			}
		}},
		{in: "/me baila", kind: Send, lines: []string{"PRIVMSG #lobby :\x01ACTION baila\x01"}, check: func(t *testing.T, a Action) {
			if !a.IsAction || !reflect.DeepEqual(a.Echo, []string{"baila"}) {
				t.Errorf("action %v echo %v", a.IsAction, a.Echo)
			}
		}},
		{in: "/JOIN lobby", kind: Send, lines: []string{"JOIN #lobby"}},
		{in: "/join #a,b key", kind: Send, lines: []string{"JOIN #a,#b key"}},
		{in: "/part", kind: Send, lines: []string{"PART #lobby"}},
		{in: "/part #otro bye", kind: Send, lines: []string{"PART #otro :bye"}},
		{in: "/msg bob hola que tal", kind: Send, lines: []string{"PRIVMSG bob :hola que tal"}, check: func(t *testing.T, a Action) {
			if a.Target != "bob" {
				t.Errorf("target %q", a.Target)
			}
		}},
		{in: "/query bob", kind: Query, check: func(t *testing.T, a Action) {
			if a.Target != "bob" || len(a.Lines) != 0 {
				t.Errorf("%+v", a)
			}
		}},
		{in: "/query bob hey", kind: Query, lines: []string{"PRIVMSG bob :hey"}},
		{in: "/nick alice2", kind: Send, lines: []string{"NICK alice2"}},
		{in: "/whois ana", kind: Send, lines: []string{"WHOIS ana"}},
		{in: "/ban ana", kind: Send, lines: []string{"MODE #lobby +b ana!*@*"}},
		{in: "/kick ana fuera de aquí", kind: Send, lines: []string{"KICK #lobby ana :fuera de aquí"}},
		{in: "/kick ana", kind: Send, lines: []string{"KICK #lobby ana"}},
		{in: "/topic", kind: Send, lines: []string{"TOPIC #lobby"}},
		{in: "/topic nuevo tema", kind: Send, lines: []string{"TOPIC #lobby :nuevo tema"}},
		{in: "/mode +o ana", kind: Send, lines: []string{"MODE #lobby +o ana"}},
		{in: "/mode #otro +m", kind: Send, lines: []string{"MODE #otro +m"}},
		{in: "/quit adiós", kind: Quit, check: func(t *testing.T, a Action) {
			if a.Arg != "adiós" {
				t.Errorf("arg %q", a.Arg)
			}
		}},
		{in: "/conectar znc", kind: Connect, check: func(t *testing.T, a Action) {
			if a.Arg != "znc" {
				t.Errorf("arg %q", a.Arg)
			}
		}},
		{in: "/connect libera 6697", kind: Connect, check: func(t *testing.T, a Action) {
			if a.Arg != "libera 6697" {
				t.Errorf("arg %q", a.Arg)
			}
		}},
		{in: "/close", kind: Close},
		{in: "/help", kind: Help},
		{in: "/quote PRIVMSG x :y", kind: Send, lines: []string{"PRIVMSG x :y"}},
		{in: "/frob", kind: Error, check: func(t *testing.T, a Action) {
			if a.Err != "Unknown command: /frob" {
				t.Errorf("err %q", a.Err)
			}
		}},
		{in: "//slash", kind: Send, lines: []string{"PRIVMSG #lobby :/slash"}},
		{in: "/msg bob", kind: Error, check: func(t *testing.T, a Action) {
			if !strings.HasPrefix(a.Err, "Usage:") {
				t.Errorf("err %q", a.Err)
			}
		}},
		{in: "/part", ctx: Context{Kind: state.KindStatus, Nick: "alice"}, kind: Error},
	}
	for _, c := range cases {
		ctx := c.ctx
		if ctx == (Context{}) {
			ctx = chanCtx
		}
		a := Parse(c.in, ctx)
		if a.Kind != c.kind {
			t.Errorf("%q: kind %v want %v (%+v)", c.in, a.Kind, c.kind, a)
			continue
		}
		if c.lines != nil && !reflect.DeepEqual(a.Lines, c.lines) {
			t.Errorf("%q: lines %q want %q", c.in, a.Lines, c.lines)
		}
		if c.check != nil {
			c.check(t, a)
		}
	}
}

func TestSplitMessage(t *testing.T) {
	long := strings.Repeat("á", 1000)
	parts := SplitMessage(long, 400)
	if len(parts) != 5 {
		t.Fatalf("parts %d", len(parts))
	}
	for _, p := range parts {
		if len(p) > 400 || !utf8.ValidString(p) {
			t.Fatalf("bad part len %d", len(p))
		}
	}
	if strings.Join(parts, "") != long {
		t.Fatal("content changed")
	}
	if got := SplitMessage("aaa bbb ccc", 8); !reflect.DeepEqual(got, []string{"aaa bbb", "ccc"}) {
		t.Fatalf("got %q", got)
	}
	if got := SplitMessage("", 8); len(got) != 0 {
		t.Fatalf("empty: %q", got)
	}
}

func TestLongMessageSplitsIntoLines(t *testing.T) {
	a := Parse(strings.Repeat("x", 900), chanCtx)
	if len(a.Lines) != 3 || len(a.Echo) != 3 {
		t.Fatalf("lines %d echo %d", len(a.Lines), len(a.Echo))
	}
}
