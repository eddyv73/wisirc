package app

import (
	"bufio"
	"context"
	"net"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/eddyv73/wisirc/internal/config"
	"github.com/eddyv73/wisirc/internal/irc"
)

type dialRec struct {
	mu   sync.Mutex
	opts []irc.Options
}

func (d *dialRec) factory(o irc.Options) func(context.Context) (net.Conn, error) {
	d.mu.Lock()
	d.opts = append(d.opts, o)
	d.mu.Unlock()
	return func(ctx context.Context) (net.Conn, error) {
		c, s := net.Pipe()
		go func() { <-ctx.Done(); s.Close() }()
		go func() { // drain whatever the client writes
			buf := make([]byte, 512)
			for {
				if _, err := s.Read(buf); err != nil {
					return
				}
			}
		}()
		return c, nil
	}
}

func newTestManager(t *testing.T, cfg *config.Config) (*Manager, *dialRec) {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	d := &dialRec{}
	m := NewManager(ctx, cfg, d.factory)
	go func() { // keep the event channel flowing
		for {
			select {
			case <-m.Events():
			case <-ctx.Done():
				return
			}
		}
	}()
	return m, d
}

func srv(host string) config.Server {
	return config.Server{Nickname: "alice", Username: "alice", Server: host, Port: 6697}
}

func TestReloadStartsOnlyNew(t *testing.T) {
	cfg := &config.Config{Servers: map[string]config.Server{"a": srv("a.example")}}
	m, _ := newTestManager(t, cfg)
	m.Start("a", cfg.Servers["a"])
	cfg2 := &config.Config{Servers: map[string]config.Server{"a": srv("a.example"), "b": srv("b.example")}}
	if got := m.Reload(cfg2); !reflect.DeepEqual(got, []string{"b"}) {
		t.Fatalf("started %v", got)
	}
	if got := m.Names(); !reflect.DeepEqual(got, []string{"a", "b"}) {
		t.Fatalf("names %v", got)
	}
}

func TestConnectAdHoc(t *testing.T) {
	cfg := &config.Config{Servers: map[string]config.Server{}}
	m, d := newTestManager(t, cfg)
	name, err := m.Connect("irc.example.org 6697")
	if err != nil || name != "irc.example.org" {
		t.Fatalf("%q %v", name, err)
	}
	o := d.opts[len(d.opts)-1]
	if o.Addr != "irc.example.org:6697" || !o.TLS || o.Nick != "wisirc" {
		t.Fatalf("opts %+v", o)
	}
	if _, ok := m.Get("irc.example.org"); !ok {
		t.Fatal("not registered")
	}
}

func TestConnectAdHocPlainPort(t *testing.T) {
	cfg := &config.Config{Servers: map[string]config.Server{"x": srv("x")}}
	m, d := newTestManager(t, cfg)
	m.Connect("irc.example.org 6667")
	o := d.opts[len(d.opts)-1]
	if o.TLS || o.Nick != "alice" {
		t.Fatalf("opts %+v", o)
	}
}

func TestConnectConfigured(t *testing.T) {
	cfg := &config.Config{Servers: map[string]config.Server{"libera": srv("irc.libera.chat")}}
	m, _ := newTestManager(t, cfg)
	if name, err := m.Connect("libera"); err != nil || name != "libera" {
		t.Fatalf("%q %v", name, err)
	}
	if _, ok := m.Get("libera"); !ok {
		t.Fatal("not started")
	}
}

func TestConnectEmpty(t *testing.T) {
	m, _ := newTestManager(t, &config.Config{})
	if _, err := m.Connect(""); err == nil || !strings.Contains(err.Error(), "Usage:") {
		t.Fatalf("err %v", err)
	}
}

func TestDisconnectRemoves(t *testing.T) {
	cfg := &config.Config{Servers: map[string]config.Server{"a": srv("a")}}
	m, _ := newTestManager(t, cfg)
	m.Start("a", cfg.Servers["a"])
	m.Disconnect("a")
	if _, ok := m.Get("a"); ok {
		t.Fatal("still registered")
	}
}

func TestOptionsFromConfig(t *testing.T) {
	s := config.Server{
		Nickname: "alice", Username: "alice", Server: "irc.example.net", Port: 6697,
		Channels: []string{"#lobby"}, DangerouslyAcceptInvalidCerts: true,
		SASL: &config.SASL{Plain: &config.SASLPlain{Username: "alice", Password: "hunter2"}},
	}
	o := optionsFor("examplenet", s)
	want := irc.Options{
		Name: "examplenet", Addr: "irc.example.net:6697", TLS: true, InsecureTLS: true,
		Nick: "alice", User: "alice", Real: "alice", SASLUser: "alice", SASLPass: "hunter2",
		Channels: []string{"#lobby"},
	}
	if !reflect.DeepEqual(o, want) {
		t.Fatalf("got %+v", o)
	}
}

func TestSASLFailureAllowsReconnect(t *testing.T) {
	cfg := &config.Config{Servers: map[string]config.Server{"a": {
		Nickname: "alice", Username: "alice", Server: "a", Port: 6697,
		SASL: &config.SASL{Plain: &config.SASLPlain{Username: "alice", Password: "bad"}},
	}}}
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	dial := func(irc.Options) func(context.Context) (net.Conn, error) {
		return func(ctx context.Context) (net.Conn, error) {
			c, s := net.Pipe()
			go func() {
				r := bufio.NewReader(s)
				for {
					line, err := r.ReadString('\n')
					if err != nil {
						return
					}
					switch {
					case strings.HasPrefix(line, "CAP LS"):
						go s.Write([]byte(":srv CAP * LS :sasl\r\n"))
					case strings.HasPrefix(line, "CAP REQ"):
						go s.Write([]byte(":srv CAP alice ACK :sasl\r\n"))
					case strings.HasPrefix(line, "AUTHENTICATE PLAIN"):
						go s.Write([]byte("AUTHENTICATE +\r\n"))
					case strings.HasPrefix(line, "AUTHENTICATE "):
						go s.Write([]byte(":srv 904 alice :bad password\r\n"))
					}
				}
			}()
			return c, nil
		}
	}
	m := NewManager(ctx, cfg, dial)
	go func() {
		for {
			select {
			case <-m.Events():
			case <-ctx.Done():
				return
			}
		}
	}()
	m.Start("a", cfg.Servers["a"])
	deadline := time.Now().Add(2 * time.Second)
	for {
		if _, ok := m.Get("a"); !ok {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("client entry kept after fatal SASL failure")
		}
		time.Sleep(10 * time.Millisecond)
	}
	if _, err := m.Connect("a"); err != nil {
		t.Fatal(err)
	}
	if _, ok := m.Get("a"); !ok {
		t.Fatal("Connect did not restart the network")
	}
}

func TestConnectAdHocBNC(t *testing.T) {
	m, d := newTestManager(t, &config.Config{})
	name, err := m.Connect("192.0.2.5 7000 -notls -nick pepe -user pepe/libera -pass secreto")
	if err != nil || name != "192.0.2.5" {
		t.Fatalf("%q %v", name, err)
	}
	o := d.opts[len(d.opts)-1]
	if o.Addr != "192.0.2.5:7000" || o.TLS || o.Nick != "pepe" || o.User != "pepe/libera" || o.Pass != "secreto" {
		t.Fatalf("opts %+v", o)
	}
}

func TestConnectAdHocSASLAndJoin(t *testing.T) {
	m, d := newTestManager(t, &config.Config{})
	name, err := m.Connect("irc.example.org -sasl ana:clave -insecure -join #a,#b -name casa")
	if err != nil || name != "casa" {
		t.Fatalf("%q %v", name, err)
	}
	o := d.opts[len(d.opts)-1]
	if o.Addr != "irc.example.org:6697" || !o.TLS || !o.InsecureTLS || o.SASLUser != "ana" || o.SASLPass != "clave" ||
		!reflect.DeepEqual(o.Channels, []string{"#a", "#b"}) {
		t.Fatalf("opts %+v", o)
	}
	if _, ok := m.Get("casa"); !ok {
		t.Fatal("not registered under -name")
	}
}

func TestConnectAdHocTLSOnOtherPort(t *testing.T) {
	m, d := newTestManager(t, &config.Config{})
	if _, err := m.Connect("irc.example.org 7000 -tls"); err != nil {
		t.Fatal(err)
	}
	if o := d.opts[len(d.opts)-1]; !o.TLS {
		t.Fatalf("opts %+v", o)
	}
}

func TestConnectAdHocBadOptions(t *testing.T) {
	m, _ := newTestManager(t, &config.Config{})
	for in, want := range map[string]string{
		"h -frob":        "Unknown option: -frob",
		"h -pass":        "Missing value for -pass",
		"h -sasl soloyo": "Usage: -sasl user:password",
		"h 6667 7000":    "Extra argument: 7000",
		"h 99999":        "Invalid port: 99999",
	} {
		if _, err := m.Connect(in); err == nil || err.Error() != want {
			t.Errorf("%q: err %v, want %q", in, err, want)
		}
	}
}

func tempCfg(t *testing.T) *config.Config {
	t.Helper()
	p := filepath.Join(t.TempDir(), "config.toml")
	os.WriteFile(p, []byte("[servers.a]\n  nickname = \"alice\"\n  server = \"a\"\n"), 0o600)
	cfg, _, err := config.Load(p)
	if err != nil {
		t.Fatal(err)
	}
	return cfg
}

func TestSaveAdHocPersists(t *testing.T) {
	cfg := tempCfg(t)
	m, _ := newTestManager(t, cfg)
	name, err := m.Connect("192.0.2.5 7000 -notls -user pepe/libera -pass secreto -name znc")
	if err != nil || !m.IsAdHoc(name) {
		t.Fatalf("%q %v adhoc=%v", name, err, m.IsAdHoc(name))
	}
	if err := m.SaveAdHoc("znc"); err != nil {
		t.Fatal(err)
	}
	if m.IsAdHoc("znc") {
		t.Fatal("still ad-hoc after save")
	}
	again, _, err := config.Load(cfg.Path)
	if err != nil {
		t.Fatal(err)
	}
	z := again.Servers["znc"]
	if z.Server != "192.0.2.5" || z.Port != 7000 || z.TLS() || z.Username != "pepe/libera" || z.Password != "secreto" || !again.Managed["znc"] {
		t.Fatalf("saved %+v", z)
	}
	if _, ok := m.Config().Servers["znc"]; !ok {
		t.Fatal("manager config not updated")
	}
}

func TestSaveServerEditRenameDelete(t *testing.T) {
	cfg := tempCfg(t)
	m, _ := newTestManager(t, cfg)
	s := config.Server{Nickname: "pepe", Server: "irc.example.org", Port: 6697}
	if err := m.SaveServer("", "ex", s); err != nil {
		t.Fatal(err)
	}
	if err := m.SaveServer("", "ex", s); err == nil {
		t.Fatal("duplicate name accepted")
	}
	if err := m.SaveServer("", "a", s); err == nil {
		t.Fatal("name from config.toml accepted")
	}
	if err := m.SaveServer("a", "a", s); err == nil || !strings.Contains(err.Error(), "config.toml") {
		t.Fatalf("editing a config.toml server: %v", err)
	}
	if err := m.SaveServer("ex", "ex2", s); err != nil {
		t.Fatal(err)
	}
	again, _, _ := config.Load(cfg.Path)
	if _, old := again.Servers["ex"]; old || !again.Managed["ex2"] {
		t.Fatalf("rename: %v", again.Managed)
	}
	if err := m.DeleteServer("a"); err == nil {
		t.Fatal("deleted a config.toml server")
	}
	if err := m.DeleteServer("ex2"); err != nil {
		t.Fatal(err)
	}
	again, _, _ = config.Load(cfg.Path)
	if len(again.Managed) != 0 || len(m.Config().Servers) != 1 {
		t.Fatalf("after delete %v %v", again.Managed, m.Config().Servers)
	}
}

func TestReloadSkipsNoAutoconnect(t *testing.T) {
	m, _ := newTestManager(t, &config.Config{Servers: map[string]config.Server{}})
	off := false
	got := m.Reload(&config.Config{Servers: map[string]config.Server{"x": {Server: "x", AutoConnect: &off}}})
	if len(got) != 0 {
		t.Fatalf("started %v", got)
	}
}

func TestImportSkipsExisting(t *testing.T) {
	cfg := tempCfg(t) // has "a"
	m, _ := newTestManager(t, cfg)
	imported, skipped, err := m.Import(map[string]config.Server{
		"a": {Server: "otro"},
		"b": {Server: "b.example", Port: 6697},
		"c": {Server: "c.example", Port: 6667},
	})
	if err != nil || !reflect.DeepEqual(imported, []string{"b", "c"}) || !reflect.DeepEqual(skipped, []string{"a"}) {
		t.Fatalf("%v %v %v", imported, skipped, err)
	}
	again, _, _ := config.Load(cfg.Path)
	if again.Servers["a"].Server != "a" || !again.Managed["b"] || !again.Managed["c"] {
		t.Fatalf("saved %+v", again.Servers)
	}
}
