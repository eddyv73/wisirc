package config

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

const userExample = `[servers.examplenet]
  nickname = "alice"
  username = "alice"
  server = "irc.example.net"
  port = 6697
  channels = ["#lobby"]
  dangerously_accept_invalid_certs = true

  [servers.examplenet.sasl.plain]
  username = "alice"
  password = "hunter2"
`

func writeTemp(t *testing.T, content string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "config.toml")
	if err := os.WriteFile(p, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestLoadUserExample(t *testing.T) {
	cfg, warnings, err := Load(writeTemp(t, userExample))
	if err != nil {
		t.Fatal(err)
	}
	s, ok := cfg.Servers["examplenet"]
	if !ok {
		t.Fatal("server missing")
	}
	if s.Nickname != "alice" || s.Username != "alice" || s.Server != "irc.example.net" || s.Port != 6697 {
		t.Fatalf("bad server: %+v", s)
	}
	if !reflect.DeepEqual(s.Channels, []string{"#lobby"}) {
		t.Fatalf("channels: %v", s.Channels)
	}
	if !s.DangerouslyAcceptInvalidCerts || !s.TLS() {
		t.Fatal("tls flags")
	}
	if s.SASL == nil || s.SASL.Plain == nil || s.SASL.Plain.Username != "alice" || s.SASL.Plain.Password != "hunter2" {
		t.Fatalf("sasl: %+v", s.SASL)
	}
	if s.Real() != "alice" {
		t.Fatalf("real: %q", s.Real())
	}
	if s.Addr() != "irc.example.net:6697" {
		t.Fatalf("addr: %q", s.Addr())
	}
	if len(warnings) != 0 {
		t.Fatalf("warnings: %v", warnings)
	}
	if !cfg.Notifications.On() {
		t.Fatal("notifications should default on")
	}
}

func TestLoadUnknownKeysWarn(t *testing.T) {
	content := strings.Replace(userExample, "port = 6697", "port = 6697\n  foo = 1", 1) + "\n[buffer]\nx = 1\n"
	_, warnings, err := Load(writeTemp(t, content))
	if err != nil {
		t.Fatal(err)
	}
	if len(warnings) != 2 {
		t.Fatalf("want 2 warnings, got %v", warnings)
	}
	joined := strings.Join(warnings, "\n")
	if !strings.Contains(joined, "servers.examplenet.foo") || !strings.Contains(joined, "buffer") {
		t.Fatalf("warnings: %v", warnings)
	}
}

func TestLoadSyntaxErrorHasLine(t *testing.T) {
	_, _, err := Load(writeTemp(t, "[servers.a]\nnickname = \"a\"\nport = \n"))
	if err == nil {
		t.Fatal("want error")
	}
	if !strings.Contains(err.Error(), ":3:") {
		t.Fatalf("error lacks line: %v", err)
	}
}

func TestTLSDefaults(t *testing.T) {
	s := Server{Server: "h", Port: 6667}
	if s.TLS() {
		t.Fatal("6667 should not be TLS")
	}
	on := true
	s.UseTLS = &on
	if !s.TLS() {
		t.Fatal("use_tls should force TLS")
	}
	s0 := Server{Server: "h", UseTLS: new(bool)}
	if !strings.HasSuffix(s0.Addr(), ":6667") {
		t.Fatalf("addr: %q", s0.Addr())
	}
}

func TestDefaultPathXDG(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	if got, want := DefaultPath(), filepath.Join(dir, "wisirc", "config.toml"); got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestEnsureExample(t *testing.T) {
	p := filepath.Join(t.TempDir(), "sub", "config.toml")
	created, err := EnsureExample(p)
	if err != nil || !created {
		t.Fatalf("created=%v err=%v", created, err)
	}
	cfg, _, err := Load(p)
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.Servers) < 1 {
		t.Fatal("example has no servers")
	}
	st, _ := os.Stat(p)
	if st.Mode().Perm() != 0o600 {
		t.Fatalf("mode %v", st.Mode().Perm())
	}
	before, _ := os.ReadFile(p)
	created, err = EnsureExample(p)
	if err != nil || created {
		t.Fatalf("second call created=%v err=%v", created, err)
	}
	after, _ := os.ReadFile(p)
	if string(before) != string(after) {
		t.Fatal("file changed")
	}
}
