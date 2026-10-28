package notify

import (
	"errors"
	"reflect"
	"strings"
	"testing"
)

func found(string) (string, error)   { return "/usr/bin/x", nil }
func missing(string) (string, error) { return "", errors.New("not found") }

func TestDarwinEscapes(t *testing.T) {
	name, args, ok := Command("darwin", `a"b`, `c\d`, missing)
	if !ok || name != "osascript" || args[0] != "-e" {
		t.Fatalf("%v %v %v", name, args, ok)
	}
	if want := `display notification "c\\d" with title "a\"b"`; args[1] != want {
		t.Fatalf("got %s want %s", args[1], want)
	}
}

func TestLinuxNotifySend(t *testing.T) {
	name, args, ok := Command("linux", "t", "b", found)
	if !ok || name != "notify-send" || !reflect.DeepEqual(args, []string{"t", "b"}) {
		t.Fatalf("%v %v %v", name, args, ok)
	}
}

func TestLinuxMissing(t *testing.T) {
	if _, _, ok := Command("linux", "t", "b", missing); ok {
		t.Fatal("want !ok")
	}
}

func TestBodyTruncated(t *testing.T) {
	_, args, _ := Command("linux", "t", strings.Repeat("á", 300), found)
	if n := len([]rune(args[1])); n != 200 {
		t.Fatalf("len %d", n)
	}
}
