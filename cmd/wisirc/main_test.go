package main

import (
	"testing"

	"github.com/charmbracelet/colorprofile"
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
