package irc

import (
	"reflect"
	"testing"
	"time"
)

func TestParseFull(t *testing.T) {
	m, err := Parse("@time=2026-10-27T01:02:03.000Z;msgid=a\\sb :alice!k@host PRIVMSG #lobby :hola qué tal\r\n")
	if err != nil {
		t.Fatal(err)
	}
	if m.Tags["time"] != "2026-10-27T01:02:03.000Z" || m.Tags["msgid"] != "a b" {
		t.Fatalf("tags: %v", m.Tags)
	}
	if m.Nick() != "alice" || m.Command != "PRIVMSG" {
		t.Fatalf("nick/cmd: %q %q", m.Nick(), m.Command)
	}
	if !reflect.DeepEqual(m.Params, []string{"#lobby", "hola qué tal"}) {
		t.Fatalf("params: %q", m.Params)
	}
	ts, ok := m.Time()
	if !ok || !ts.Equal(time.Date(2026, 10, 27, 1, 2, 3, 0, time.UTC)) {
		t.Fatalf("time: %v %v", ts, ok)
	}
}

func TestParseNoPrefixNumeric(t *testing.T) {
	m, err := Parse("001 alice :Welcome")
	if err != nil || m.Command != "001" || m.Param(0) != "alice" || m.Param(5) != "" {
		t.Fatalf("%+v %v", m, err)
	}
}

func TestParseLowercaseCommand(t *testing.T) {
	m, err := Parse("ping :x")
	if err != nil || m.Command != "PING" || m.Trailing() != "x" {
		t.Fatalf("%+v %v", m, err)
	}
}

func TestParseEmpty(t *testing.T) {
	for _, in := range []string{"", ":prefix", "\r\n", "@a=b"} {
		if _, err := Parse(in); err == nil {
			t.Errorf("Parse(%q) want error", in)
		}
	}
}

func TestEncode(t *testing.T) {
	cases := map[string]string{
		Encode("PRIVMSG", "#a", "hola mundo"): "PRIVMSG #a :hola mundo",
		Encode("NICK", "alice"):               "NICK alice",
		Encode("PRIVMSG", "#a", ":)"):         "PRIVMSG #a ::)",
		Encode("TOPIC", "#a", ""):             "TOPIC #a :",
	}
	for got, want := range cases {
		if got != want {
			t.Errorf("got %q want %q", got, want)
		}
	}
}

func TestDecodeLineLatin1(t *testing.T) {
	if got := DecodeLine([]byte{'c', 'a', 'f', 0xE9}); got != "café" {
		t.Fatalf("got %q", got)
	}
	if got := DecodeLine([]byte("niño 😀")); got != "niño 😀" {
		t.Fatalf("got %q", got)
	}
}
