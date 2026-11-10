package main

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/charmbracelet/colorprofile"
	"github.com/eddyv73/wisirc/internal/config"
)

// 6. an explicit COLORTERM=truecolor wins even inside tmux
func TestColorOverride(t *testing.T) {
	env := func(m map[string]string) func(string) string { return func(k string) string { return m[k] } }
	if p, ok := colorOverride(env(map[string]string{"COLORTERM": "truecolor", "TERM": "tmux-256color"})); !ok || p != colorprofile.TrueColor {
		t.Fatalf("truecolor: %v %v", p, ok)
	}
	if p, ok := colorOverride(env(map[string]string{"COLORTERM": "24bit"})); !ok || p != colorprofile.TrueColor {
		t.Fatalf("24bit: %v %v", p, ok)
	}
	if _, ok := colorOverride(env(map[string]string{"TERM": "xterm-256color"})); ok {
		t.Fatal("no COLORTERM must leave detection alone")
	}
}

func TestVersionString(t *testing.T) {
	old := Version
	t.Cleanup(func() { Version = old })
	Version = "v1.2.3"
	if got := versionString(); got != "v1.2.3" {
		t.Fatalf("ldflags version: %q", got)
	}
	Version = "dev"
	if got := versionString(); got == "" {
		t.Fatal("empty version")
	}
}

func TestPrefsChoicesDoNotClobberEachOther(t *testing.T) {
	cfgPath := filepath.Join(t.TempDir(), "config.toml")
	f := prefsFile(cfgPath)
	steps := []error{
		f.theme("papel"),
		f.lang("es"),
		f.color("znc", "#ff8800"),
		f.color("libera", "#00ff88"),
		f.disabled([]string{"darkstar"}),
		f.color("libera", ""), // back to automatic
	}
	for i, err := range steps {
		if err != nil {
			t.Fatalf("step %d: %v", i, err)
		}
	}
	p := config.LoadPrefs(cfgPath)
	if p.Theme != "papel" || p.Lang != "es" || !reflect.DeepEqual(p.Disabled, []string{"darkstar"}) ||
		!reflect.DeepEqual(p.ServerColors, map[string]string{"znc": "#ff8800"}) {
		t.Fatalf("prefs %+v", p)
	}
	if err := f.disabled(nil); err != nil {
		t.Fatal(err)
	}
	if p := config.LoadPrefs(cfgPath); len(p.Disabled) != 0 || p.Theme != "papel" {
		t.Fatalf("after enabling all: %+v", p)
	}
	if _, err := os.Stat(filepath.Join(filepath.Dir(cfgPath), "config.toml")); !os.IsNotExist(err) {
		t.Fatal("saving prefs wrote config.toml")
	}
}
