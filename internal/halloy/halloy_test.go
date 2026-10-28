package halloy

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

const sample = `# Halloy config.
[servers.examplenet]
  nickname = "alice"
  username = "alice"
  server = "irc.example.net"
  port = 6697
  channels = ["#lobby"]
  dangerously_accept_invalid_certs = true

  [servers.examplenet.sasl.plain]
  username = "alice"
  password = "hunter2"

[servers.znc]
nickname = "zephyr"
username = "bob/libera"
server = "192.0.2.10"
port = 7000
password = "secreto"
use_tls = false

[servers.libera]
nickname = "pepe"
server = "irc.libera.chat"
password_file = "libera.pass"
nick_password = "x"

[servers.cmd]
nickname = "pepe"
server = "irc.example.org"
port = 6667
password_command = "pass show irc"
[servers.cmd.sasl.plain]
username = "pepe"
password_file = "/no/existe"

[buffer]
something = 1
`

func write(t *testing.T) string {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "libera.pass"), []byte("clave-libera\n"), 0o600)
	p := filepath.Join(dir, "config.toml")
	os.WriteFile(p, []byte(sample), 0o600)
	return p
}

func TestLoadMapsServers(t *testing.T) {
	servers, warns, err := Load(write(t))
	if err != nil {
		t.Fatal(err)
	}
	if got := sortedKeys(servers); !reflect.DeepEqual(got, []string{"cmd", "examplenet", "libera", "znc"}) {
		t.Fatalf("names %v", got)
	}
	o := servers["examplenet"]
	if o.Server != "irc.example.net" || o.Port != 6697 || !o.TLS() || !o.DangerouslyAcceptInvalidCerts ||
		o.SASL.Plain.Password != "hunter2" || o.Channels[0] != "#lobby" {
		t.Fatalf("irc.example.net %+v", o)
	}
	z := servers["znc"]
	if z.TLS() || z.Port != 7000 || z.Password != "secreto" || z.Username != "bob/libera" {
		t.Fatalf("znc %+v", z)
	}
	// halloy defaults to TLS even on unusual ports, and reads password_file
	l := servers["libera"]
	if !l.TLS() || l.Port != 6697 || l.Password != "clave-libera" {
		t.Fatalf("libera %+v", l)
	}
	c := servers["cmd"]
	if !c.TLS() || c.Password != "" || c.SASL != nil {
		t.Fatalf("cmd %+v", c)
	}
	joined := strings.Join(warns, "\n")
	for _, want := range []string{"libera: nick_password", "cmd: password_command", "cmd: sasl.plain.password_file"} {
		if !strings.Contains(joined, want) {
			t.Fatalf("missing warning %q in:\n%s", want, joined)
		}
	}
}

func TestLoadMissing(t *testing.T) {
	if _, _, err := Load(filepath.Join(t.TempDir(), "nope.toml")); err == nil {
		t.Fatal("no error")
	}
}

func TestPaths(t *testing.T) {
	mac := Paths("darwin", "/Users/u", "")
	if mac[0] != "/Users/u/Library/Application Support/halloy/config.toml" || mac[1] != "/Users/u/.config/halloy/config.toml" {
		t.Fatalf("mac %v", mac)
	}
	lin := Paths("linux", "/home/u", "/xdg")
	want := []string{"/xdg/halloy/config.toml", "/home/u/.config/halloy/config.toml", "/home/u/.var/app/org.squidowl.halloy/config/config.toml"}
	if !reflect.DeepEqual(lin, want) {
		t.Fatalf("linux %v", lin)
	}
}
