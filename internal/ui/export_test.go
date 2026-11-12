package ui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/BurntSushi/toml"
	"github.com/eddyv73/wisirc/internal/config"
)

func (h *harness) run(line string) {
	h.t.Helper()
	h.typeText(line)
	h.send(tea.KeyPressMsg{Code: tea.KeyEnter})
}

// activeLines is what the active buffer shows, read from the buffer (not the
// screen) so long paths are not cut by the window width.
func (h *harness) activeLines() string {
	var b strings.Builder
	for _, l := range h.m.st.Buffer(h.m.st.Active).Lines {
		b.WriteString(l.Text + "\n")
	}
	return b.String()
}

func exportedServers(t *testing.T, path string) map[string]map[string]any {
	t.Helper()
	var f struct {
		Servers map[string]map[string]any `toml:"servers"`
	}
	if _, err := toml.DecodeFile(path, &f); err != nil {
		t.Fatalf("%s: %v", path, err)
	}
	return f.Servers
}

func TestExportHalloyWritesNextToConfig(t *testing.T) {
	h := newHarness(t)
	dir := t.TempDir()
	h.m.d.CfgPath = filepath.Join(dir, "config.toml")
	h.nets.cfg.Servers["znc"] = config.Server{Nickname: "bob", Server: "192.0.2.10", Port: 7000, Password: "secreto"}
	h.m.disabled["znc"] = true
	h.nets.adhoc["suelto"] = true // a /connect that was never saved is not exported

	h.run("/export halloy")

	want := filepath.Join(dir, "halloy-export.toml")
	got := exportedServers(t, want)
	if len(got) != 2 || got[netName]["server"] != "irc.example.net" || got["znc"]["password"] != "secreto" {
		t.Fatalf("exported %v", got)
	}
	if got["znc"]["autoconnect"] != false || got[netName]["autoconnect"] != true {
		t.Errorf("a disabled server should not autoconnect in Halloy: %v", got)
	}
	lines := h.activeLines()
	if !strings.Contains(lines, "Exported 2 server(s) to "+want) || !strings.Contains(lines, "passwords") {
		t.Errorf("buffer:\n%s", lines)
	}
}

func TestExportHalloyNeverOverwrites(t *testing.T) {
	h := newHarness(t)
	p := filepath.Join(t.TempDir(), "halloy.toml")
	if err := os.WriteFile(p, []byte("# my halloy\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	h.run("/export halloy " + p)
	if b, _ := os.ReadFile(p); string(b) != "# my halloy\n" {
		t.Fatalf("existing file was changed: %q", b)
	}
	if !strings.Contains(h.activeLines(), "already exists") {
		t.Errorf("buffer:\n%s", h.activeLines())
	}
}

func TestExportHalloyExpandsHome(t *testing.T) {
	h := newHarness(t)
	home := t.TempDir()
	t.Setenv("HOME", home)
	h.run("/export halloy ~/halloy/servers.toml")
	if got := exportedServers(t, filepath.Join(home, "halloy", "servers.toml")); len(got) != 1 {
		t.Fatalf("exported %v", got)
	}
}

func TestExportHalloyWithoutServers(t *testing.T) {
	h := newHarness(t)
	h.m.d.CfgPath = filepath.Join(t.TempDir(), "config.toml")
	h.nets.cfg.Servers = map[string]config.Server{}
	h.run("/export halloy")
	if _, err := os.Stat(filepath.Join(filepath.Dir(h.m.d.CfgPath), "halloy-export.toml")); err == nil {
		t.Error("nothing to export, but a file was written")
	}
	if !strings.Contains(h.activeLines(), "there are no servers to export") {
		t.Errorf("buffer:\n%s", h.activeLines())
	}
}
