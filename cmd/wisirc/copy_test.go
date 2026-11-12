package main

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/eddyv73/wisirc/internal/config"
)

func write(t *testing.T, path, data string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(data), 0o644); err != nil {
		t.Fatal(err)
	}
}

func mode(t *testing.T, path string) fs.FileMode {
	t.Helper()
	st, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	return st.Mode().Perm()
}

const wisircConfig = `[servers.libera]
  nickname = "alice"
  server = "irc.libera.chat"
  port = 6697
`

func TestCopyWisircDirectory(t *testing.T) {
	src := filepath.Join(t.TempDir(), "old")
	write(t, filepath.Join(src, "config.toml"), wisircConfig)
	write(t, filepath.Join(src, "servers.toml"), "[servers.znc]\n  nickname = \"alice\"\n  server = \"192.0.2.10\"\n  port = 7000\n")
	write(t, filepath.Join(src, "prefs.toml"), "lang = \"es\"\nnick = \"alice\"\n")
	dst := filepath.Join(t.TempDir(), "wisirc", "config.toml")

	msg, warns, err := copyConfig(src, dst)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(msg, "2 server") || len(warns) != 0 {
		t.Errorf("msg %q warns %v", msg, warns)
	}
	cfg, _, err := config.Load(dst)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := cfg.Servers["libera"]; !ok {
		t.Error("config.toml not copied")
	}
	if !cfg.Managed["znc"] {
		t.Error("servers.toml not copied")
	}
	if p := config.LoadPrefs(dst); p.Lang != "es" || p.Nick != "alice" {
		t.Errorf("prefs %+v", p)
	}
	for _, f := range []string{dst, config.ServersPath(dst), config.PrefsPath(dst)} {
		if m := mode(t, f); m != 0o600 {
			t.Errorf("%s is %v, want 0600", f, m)
		}
	}
	if m := mode(t, filepath.Dir(dst)); m != 0o700 {
		t.Errorf("dir is %v, want 0700", m)
	}
}

// a lone config.toml with only wisirc keys is copied as is
func TestCopyWisircFile(t *testing.T) {
	src := filepath.Join(t.TempDir(), "mine.toml")
	write(t, src, wisircConfig+"\n[theme]\n  name = \"papel\"\n")
	dst := filepath.Join(t.TempDir(), "config.toml")
	if _, _, err := copyConfig(src, dst); err != nil {
		t.Fatal(err)
	}
	got, _ := os.ReadFile(dst)
	want, _ := os.ReadFile(src)
	if string(got) != string(want) {
		t.Errorf("copied:\n%s", got)
	}
	if _, err := os.Stat(config.ServersPath(dst)); !errors.Is(err, fs.ErrNotExist) {
		t.Error("no servers.toml to copy, none should appear")
	}
}

func TestCopyHalloyConfig(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "halloy")
	write(t, filepath.Join(dir, "pw"), "s3cret\n")
	write(t, filepath.Join(dir, "config.toml"), `theme = "ferra"

[buffer.channel.topic]
enabled = true

[servers.liberachat]
nickname = "alice"
server = "irc.libera.chat"
channels = ["#halloy"]

[servers.znc]
nickname = "alice"
server = "192.0.2.10"
port = 7000
use_tls = false
password_file = "pw"
password_command = "pass show znc"
`)
	dst := filepath.Join(t.TempDir(), "config.toml")
	msg, warns, err := copyConfig(dir, dst)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(msg, "2 server") {
		t.Errorf("msg %q", msg)
	}
	if len(warns) != 1 || !strings.Contains(warns[0], "password_command") {
		t.Errorf("warns %v", warns)
	}
	cfg, cw, err := config.Load(dst)
	if err != nil || len(cw) != 0 {
		t.Fatalf("the new config should load cleanly: %v %v", err, cw)
	}
	l, z := cfg.Servers["liberachat"], cfg.Servers["znc"]
	if !l.TLS() || l.Port != 6697 || strings.Join(l.Channels, ",") != "#halloy" {
		t.Errorf("liberachat %+v", l)
	}
	if z.TLS() || z.Port != 7000 || z.Password != "s3cret" {
		t.Errorf("znc %+v", z)
	}
	if !cfg.Managed["znc"] {
		t.Error("Halloy servers belong in servers.toml")
	}
	if m := mode(t, config.ServersPath(dst)); m != 0o600 {
		t.Errorf("servers.toml is %v", m)
	}
}

func TestCopyNeverReplaces(t *testing.T) {
	src := filepath.Join(t.TempDir(), "config.toml")
	write(t, src, wisircConfig)
	for _, existing := range []string{"config.toml", "servers.toml", "prefs.toml"} {
		dir := t.TempDir()
		write(t, filepath.Join(dir, existing), "# mine\n")
		dst := filepath.Join(dir, "config.toml")
		if _, _, err := copyConfig(src, dst); err == nil || !strings.Contains(err.Error(), "already exists") {
			t.Errorf("%s present: err %v", existing, err)
		}
		if b, _ := os.ReadFile(filepath.Join(dir, existing)); string(b) != "# mine\n" {
			t.Errorf("%s was changed", existing)
		}
	}
}

func TestCopyBadSource(t *testing.T) {
	dst := filepath.Join(t.TempDir(), "config.toml")
	if _, _, err := copyConfig(filepath.Join(t.TempDir(), "missing.toml"), dst); err == nil {
		t.Error("a missing source should fail")
	}
	bad := filepath.Join(t.TempDir(), "config.toml")
	write(t, bad, "this is = = not toml")
	if _, _, err := copyConfig(bad, dst); err == nil {
		t.Error("a broken source should fail")
	}
	if _, err := os.Stat(dst); !errors.Is(err, fs.ErrNotExist) {
		t.Error("a failed copy must not leave a config behind")
	}
}

func TestCopyExpandsHome(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	write(t, filepath.Join(home, "old", "config.toml"), wisircConfig)
	dst := filepath.Join(t.TempDir(), "config.toml")
	if _, _, err := copyConfig("~/old", dst); err != nil {
		t.Fatal(err)
	}
}
