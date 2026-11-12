package halloy

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/BurntSushi/toml"
	"github.com/eddyv73/wisirc/internal/config"
)

func ptr(b bool) *bool { return &b }

func exportServers() map[string]config.Server {
	return map[string]config.Server{
		"examplenet": {
			Nickname: "alice", Username: "alice", Realname: "Alice A.",
			Server: "irc.example.net", Port: 6697, Channels: []string{"#lobby", "#go"},
			SASL: &config.SASL{Plain: &config.SASLPlain{Username: "alice", Password: "hunter2"}},
		},
		"znc": {
			Nickname: "bob", Username: "bob/libera", Server: "192.0.2.10", Port: 7000,
			UseTLS: ptr(false), Password: "secreto", AutoConnect: ptr(false),
		},
		"lab": {
			Nickname: "bob", Server: "irc.example.org", // no port: TLS and 6697 by default
			DangerouslyAcceptInvalidCerts: true,
		},
	}
}

func TestExportWritesHalloyServers(t *testing.T) {
	out, err := Export(exportServers(), func(name string) bool { return name == "lab" })
	if err != nil {
		t.Fatal(err)
	}
	var f file
	md, err := toml.Decode(string(out), &f)
	if err != nil {
		t.Fatalf("export is not valid TOML: %v\n%s", err, out)
	}
	if len(md.Undecoded()) != 0 {
		// autoconnect is the only key Halloy has that the importer skips
		for _, k := range md.Undecoded() {
			if k[len(k)-1] != "autoconnect" {
				t.Errorf("unexpected key %v", k)
			}
		}
	}
	if len(f.Servers) != 3 {
		t.Fatalf("servers = %v", f.Servers)
	}
	ex := f.Servers["examplenet"]
	if ex.Server != "irc.example.net" || ex.Port != 6697 || ex.UseTLS == nil || !*ex.UseTLS ||
		ex.Nickname != "alice" || ex.Realname != "Alice A." || !reflect.DeepEqual(ex.Channels, []string{"#lobby", "#go"}) {
		t.Errorf("examplenet = %+v", ex)
	}
	if ex.SASL == nil || ex.SASL.Plain == nil || ex.SASL.Plain.Username != "alice" || ex.SASL.Plain.Password != "hunter2" {
		t.Errorf("examplenet sasl = %+v", ex.SASL)
	}
	znc := f.Servers["znc"]
	if znc.UseTLS == nil || *znc.UseTLS || znc.Port != 7000 || znc.Password != "secreto" || znc.Username != "bob/libera" {
		t.Errorf("znc = %+v", znc)
	}
	lab := f.Servers["lab"]
	if lab.Port != 6697 || lab.UseTLS == nil || !*lab.UseTLS || !lab.AcceptInvalid {
		t.Errorf("lab = %+v", lab)
	}

	var auto struct {
		Servers map[string]struct {
			Autoconnect *bool `toml:"autoconnect"`
		} `toml:"servers"`
	}
	if _, err := toml.Decode(string(out), &auto); err != nil {
		t.Fatal(err)
	}
	for name, want := range map[string]bool{"examplenet": true, "znc": false, "lab": false} {
		got := auto.Servers[name].Autoconnect
		if got == nil || *got != want {
			t.Errorf("%s autoconnect = %v, want %v", name, got, want)
		}
	}
	if !strings.HasPrefix(string(out), "#") || !strings.Contains(string(out), "password") {
		t.Errorf("export should start with a comment that warns about passwords:\n%s", out)
	}
}

// What Export writes, Load reads back as the same servers.
func TestExportRoundTrip(t *testing.T) {
	in := exportServers()
	out, err := Export(in, nil)
	if err != nil {
		t.Fatal(err)
	}
	p := filepath.Join(t.TempDir(), "config.toml")
	if err := os.WriteFile(p, out, 0o600); err != nil {
		t.Fatal(err)
	}
	back, warns, err := Load(p)
	if err != nil || len(warns) != 0 {
		t.Fatalf("Load: %v %v", err, warns)
	}
	for name, s := range in {
		b := back[name]
		if b.Addr() != s.Addr() || b.TLS() != s.TLS() || b.Nickname != s.Nickname || b.Password != s.Password ||
			!reflect.DeepEqual(b.SASL, s.SASL) || !reflect.DeepEqual(b.Channels, s.Channels) {
			t.Errorf("%s: got %+v, want %+v", name, b, s)
		}
	}
}

// The export is meant to be pasted at the end of Halloy's config, so it must
// not open [servers] or other tables that the config may already have.
func TestExportPastesIntoAHalloyConfig(t *testing.T) {
	out, err := Export(exportServers(), nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, l := range strings.Split(string(out), "\n") {
		l = strings.TrimSpace(l)
		if l == "[servers]" || strings.HasSuffix(l, ".sasl]") {
			t.Errorf("table header %q would clash or is empty:\n%s", l, out)
		}
	}
	halloyCfg := "[servers]\n\n[servers.mine]\nnickname = \"carol\"\nserver = \"irc.example.com\"\n\n[buffer]\nchannel_name_casing = \"lowercase\"\n"
	var f file
	if _, err := toml.Decode(halloyCfg+"\n"+string(out), &f); err != nil {
		t.Fatalf("Halloy config + export is not valid TOML: %v", err)
	}
	if len(f.Servers) != 4 {
		t.Errorf("servers = %v", f.Servers)
	}
}

func TestExportNeedsServers(t *testing.T) {
	if _, err := Export(nil, nil); err == nil {
		t.Fatal("exporting nothing should fail")
	}
}

func TestWriteNewIsPrivateAndNeverOverwrites(t *testing.T) {
	p := filepath.Join(t.TempDir(), "sub", "halloy.toml")
	if err := WriteNew(p, []byte("a")); err != nil {
		t.Fatal(err)
	}
	st, err := os.Stat(p)
	if err != nil {
		t.Fatal(err)
	}
	if st.Mode().Perm() != 0o600 {
		t.Errorf("mode = %v, want 0600 (it holds passwords)", st.Mode().Perm())
	}
	err = WriteNew(p, []byte("b"))
	if !errors.Is(err, fs.ErrExist) {
		t.Fatalf("second write: %v, want ErrExist", err)
	}
	if b, _ := os.ReadFile(p); string(b) != "a" {
		t.Errorf("file was overwritten: %q", b)
	}
}
