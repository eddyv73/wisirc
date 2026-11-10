package app

import (
	"reflect"
	"testing"

	"github.com/eddyv73/wisirc/internal/config"
)

// These tests change each server setting the way the user would (the
// Servers form, ≡ → Edit config) and check that the running connections
// follow.

func cfgWith(servers map[string]config.Server) *config.Config {
	return &config.Config{Servers: servers, Managed: map[string]bool{}}
}

func noneDisabled(string) bool { return false }

func TestFormEditsEachConnectionSetting(t *testing.T) {
	on, off := true, false
	base := config.Server{Nickname: "pepe", Server: "192.0.2.1", Port: 7000}
	edits := map[string]func(*config.Server){
		"server":   func(s *config.Server) { s.Server = "192.0.2.9" },
		"port":     func(s *config.Server) { s.Port = 7001 },
		"tls":      func(s *config.Server) { s.UseTLS = &on },
		"insecure": func(s *config.Server) { s.DangerouslyAcceptInvalidCerts = true },
		"nickname": func(s *config.Server) { s.Nickname = "pepa" },
		"username": func(s *config.Server) { s.Username = "pepe/libera" },
		"realname": func(s *config.Server) { s.Realname = "Pepe" },
		"password": func(s *config.Server) { s.Password = "secreto" },
		"sasl": func(s *config.Server) {
			s.SASL = &config.SASL{Plain: &config.SASLPlain{Username: "pepe", Password: "x"}}
		},
	}
	kept := map[string]func(*config.Server){
		"channels":    func(s *config.Server) { s.Channels = []string{"#a"} },
		"autoconnect": func(s *config.Server) { s.AutoConnect = &off },
	}
	run := func(field string, edit func(*config.Server), wantDials int) {
		t.Run(field, func(t *testing.T) {
			m, d := newTestManager(t, tempCfg(t))
			if err := m.SaveServer("", "znc", base); err != nil {
				t.Fatal(err)
			}
			m.Start("znc", base)
			s := base
			edit(&s)
			if err := m.SaveServer("znc", "znc", s); err != nil {
				t.Fatal(err)
			}
			if got := len(d.addrs()); got != wantDials {
				t.Fatalf("connections made: %d, want %d", got, wantDials)
			}
			if got := m.Names(); !reflect.DeepEqual(got, []string{"znc"}) {
				t.Fatalf("clients %v", got)
			}
			again, _, err := config.Load(m.Config().Path)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(again.Servers["znc"], s) {
				t.Fatalf("saved %+v, want %+v", again.Servers["znc"], s)
			}
		})
	}
	for f, e := range edits {
		run(f, e, 2)
	}
	for f, e := range kept {
		run(f, e, 1)
	}
}

func TestReloadAppliesEditedConfig(t *testing.T) {
	old := map[string]config.Server{"a": srv("a.example"), "b": srv("b.example"), "c": srv("c.example")}
	m, d := newTestManager(t, cfgWith(old))
	for _, n := range []string{"a", "b", "c"} {
		m.Start(n, old[n])
	}
	if _, err := m.Connect("192.0.2.5 7000 -notls -name adhoc"); err != nil {
		t.Fatal(err)
	}
	moved := srv("a2.example")
	chans := srv("b.example")
	chans.Channels = []string{"#nuevo"}
	got := m.Reload(cfgWith(map[string]config.Server{
		"a": moved, // new address: reconnect
		"b": chans, // channels only: keep
		// c removed: disconnect
		"d": srv("d.example"), // new: start
	}), noneDisabled)
	want := Reloaded{Started: []string{"d"}, Restarted: []string{"a"}, Stopped: []string{"c"}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("reload %+v, want %+v", got, want)
	}
	if names := m.Names(); !reflect.DeepEqual(names, []string{"a", "adhoc", "b", "d"}) {
		t.Fatalf("clients %v", names)
	}
	addrs := d.addrs()
	if addrs[len(addrs)-2] != "a2.example:6697" && addrs[len(addrs)-1] != "a2.example:6697" {
		t.Fatalf("a not redialed at its new address: %v", addrs)
	}
}

func TestReloadLeavesDisabledAndStoppedServersAlone(t *testing.T) {
	m, d := newTestManager(t, cfgWith(map[string]config.Server{"a": srv("a.example")}))
	// a is configured but not connected (disconnected by hand or disabled)
	got := m.Reload(cfgWith(map[string]config.Server{
		"a": srv("a2.example"),
		"z": srv("z.example"), // new, but disabled
	}), func(n string) bool { return n == "z" })
	if !reflect.DeepEqual(got, Reloaded{}) {
		t.Fatalf("reload %+v", got)
	}
	if len(d.addrs()) != 0 || len(m.Names()) != 0 {
		t.Fatalf("connected %v %v", d.addrs(), m.Names())
	}
}

func TestDeleteEditRevertsRunningConnection(t *testing.T) {
	cfg := tempCfg(t) // a = alice@a:6697 from config.toml
	m, d := newTestManager(t, cfg)
	m.Start("a", cfg.Servers["a"])
	s := cfg.Servers["a"]
	s.Server = "a-nuevo"
	if err := m.SaveServer("a", "a", s); err != nil {
		t.Fatal(err)
	}
	if err := m.DeleteServer("a"); err != nil { // Undo edit
		t.Fatal(err)
	}
	addrs := d.addrs()
	if len(addrs) != 3 || addrs[1] != "a-nuevo:6697" || addrs[2] != "a:6697" {
		t.Fatalf("dialed %v", addrs)
	}
}
