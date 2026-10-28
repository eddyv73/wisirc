package clipboard

import (
	"errors"
	"reflect"
	"testing"
)

func only(names ...string) func(string) (string, error) {
	return func(n string) (string, error) {
		for _, x := range names {
			if x == n {
				return "/usr/bin/" + n, nil
			}
		}
		return "", errors.New("not found")
	}
}

func TestCommand(t *testing.T) {
	cases := []struct {
		goos  string
		found []string
		name  string
		args  []string
		ok    bool
	}{
		{"darwin", nil, "pbcopy", nil, true},
		{"linux", []string{"wl-copy", "xclip"}, "wl-copy", nil, true},
		{"linux", []string{"xclip", "xsel"}, "xclip", []string{"-selection", "clipboard"}, true},
		{"linux", []string{"xsel"}, "xsel", []string{"--clipboard", "--input"}, true},
		{"linux", nil, "", nil, false},
	}
	for _, c := range cases {
		name, args, ok := Command(c.goos, only(c.found...))
		if name != c.name || !reflect.DeepEqual(args, c.args) || ok != c.ok {
			t.Errorf("%s %v: got %q %q %v", c.goos, c.found, name, args, ok)
		}
	}
}
