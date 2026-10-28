package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadMergesServersFile(t *testing.T) {
	p := writeTemp(t, userExample)
	os.WriteFile(ServersPath(p), []byte(`[servers.znc]
  nickname = "zephyr"
  server = "192.0.2.10"
  port = 7000
  use_tls = false
  autoconnect = false

[servers.examplenet]
  nickname = "otra"
  server = "x"
`), 0o600)
	cfg, warns, err := Load(p)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Path != p {
		t.Fatalf("path %q", cfg.Path)
	}
	z, ok := cfg.Servers["znc"]
	if !ok || z.Server != "192.0.2.10" || z.Auto() || !cfg.Managed["znc"] {
		t.Fatalf("znc %+v managed %v", z, cfg.Managed)
	}
	if cfg.Servers["examplenet"].Nickname != "alice" || cfg.Managed["examplenet"] {
		t.Fatal("config.toml must win over servers.toml")
	}
	if len(warns) != 1 || !strings.Contains(warns[0], "examplenet") {
		t.Fatalf("warns %q", warns)
	}
	if !cfg.Servers["examplenet"].Auto() {
		t.Fatal("autoconnect must default to true")
	}
}

func TestLoadWithoutServersFile(t *testing.T) {
	cfg, _, err := Load(writeTemp(t, userExample))
	if err != nil || len(cfg.Managed) != 0 || len(cfg.Servers) != 1 {
		t.Fatalf("%+v %v", cfg, err)
	}
}

func TestLoadBrokenServersFileFails(t *testing.T) {
	p := writeTemp(t, userExample)
	os.WriteFile(ServersPath(p), []byte("[servers.x\n"), 0o600)
	if _, _, err := Load(p); err == nil || !strings.Contains(err.Error(), "servers.toml:") {
		t.Fatalf("err %v", err)
	}
}

func TestSaveServersRoundTrip(t *testing.T) {
	p := writeTemp(t, userExample)
	off := false
	err := SaveServers(p, map[string]Server{"znc": {
		Nickname: "zephyr", Username: "bob/libera", Server: "192.0.2.10", Port: 7000,
		Password: "secreto", UseTLS: &off, AutoConnect: &off, Channels: []string{"#a"},
		SASL: &SASL{Plain: &SASLPlain{Username: "u", Password: "p"}},
	}})
	if err != nil {
		t.Fatal(err)
	}
	st, _ := os.Stat(ServersPath(p))
	if st.Mode().Perm() != 0o600 {
		t.Fatalf("mode %v", st.Mode())
	}
	raw, _ := os.ReadFile(ServersPath(p))
	if !strings.HasPrefix(string(raw), "#") {
		t.Fatalf("missing header:\n%s", raw)
	}
	cfg, warns, err := Load(p)
	if err != nil || len(warns) != 0 {
		t.Fatalf("%v %v", warns, err)
	}
	z := cfg.Servers["znc"]
	if z.Username != "bob/libera" || z.Password != "secreto" || z.TLS() || z.Auto() || z.SASL.Plain.Password != "p" || z.Channels[0] != "#a" {
		t.Fatalf("round trip %+v", z)
	}
	if left, _ := filepath.Glob(filepath.Join(filepath.Dir(p), "*.tmp*")); len(left) != 0 {
		t.Fatalf("temp files left: %v", left)
	}
}

func TestEnsureStarterHasNoServers(t *testing.T) {
	p := filepath.Join(t.TempDir(), "wisirc", "config.toml")
	created, err := EnsureStarter(p)
	if err != nil || !created {
		t.Fatalf("%v %v", created, err)
	}
	cfg, warns, err := Load(p)
	if err != nil || len(warns) != 0 || len(cfg.Servers) != 0 {
		t.Fatalf("%+v %v %v", cfg, warns, err)
	}
	if again, _ := EnsureStarter(p); again {
		t.Fatal("overwrote an existing config")
	}
}
