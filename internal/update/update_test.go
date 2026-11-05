package update

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewer(t *testing.T) {
	for _, c := range []struct {
		cur, latest string
		want        bool
	}{
		{"v0.4.0", "v0.4.1", true},
		{"v0.4.0", "v0.5.0", true},
		{"v0.4.0", "v1.0.0", true},
		{"v0.4.1", "v0.4.0", false},
		{"v0.4.0", "v0.4.0", false},
		{"v0.4.0-rc.1", "v0.4.0", true},
		{"v0.3.0-3-g2d5a96d", "v0.3.0", false}, // built after the tag
		{"v0.3.0-3-g2d5a96d", "v0.3.9", true},
		{"v0.4.0", "v0.4.1-rc.1", false}, // never offer release candidates
		{"dev", "v0.4.1", false},
		{"v0.4.0", "garbage", false},
	} {
		if got := Newer(c.cur, c.latest); got != c.want {
			t.Errorf("Newer(%q, %q) = %v", c.cur, c.latest, got)
		}
	}
}

// fakeGitHub serves the latest-release API and one release's downloads.
func fakeGitHub(t *testing.T, tag string, bin []byte, sums string) *Updater {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/repos/o/r/releases/latest", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"tag_name":"` + tag + `","name":"x"}`))
	})
	mux.HandleFunc("/o/r/releases/download/"+tag+"/SHA256SUMS", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(sums))
	})
	mux.HandleFunc("/o/r/releases/download/"+tag+"/wisirc-linux-amd64", func(w http.ResponseWriter, r *http.Request) {
		w.Write(bin)
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	exe := filepath.Join(t.TempDir(), "wisirc")
	os.WriteFile(exe, []byte("old"), 0o755)
	return &Updater{
		API: srv.URL, Download: srv.URL, Repo: "o/r", HTTP: srv.Client(),
		Exe:    func() (string, error) { return exe, nil },
		Verify: func(path, tag string) error { return nil },
		Asset:  "wisirc-linux-amd64",
	}
}

func sum(b []byte) string {
	s := sha256.Sum256(b)
	return hex.EncodeToString(s[:])
}

func TestLatest(t *testing.T) {
	u := fakeGitHub(t, "v0.5.0", nil, "")
	tag, err := u.Latest(context.Background())
	if err != nil || tag != "v0.5.0" {
		t.Fatalf("%q %v", tag, err)
	}
}

func TestApplyReplacesBinary(t *testing.T) {
	bin := []byte("new binary")
	u := fakeGitHub(t, "v0.5.0", bin, sum([]byte("other"))+"  wisirc-darwin-arm64\n"+sum(bin)+"  wisirc-linux-amd64\n")
	path, err := u.Apply(context.Background(), "v0.5.0")
	if err != nil {
		t.Fatal(err)
	}
	got, _ := os.ReadFile(path)
	if string(got) != "new binary" {
		t.Fatalf("binary %q", got)
	}
	if fi, _ := os.Stat(path); fi.Mode().Perm() != 0o755 {
		t.Fatalf("mode %v", fi.Mode())
	}
	left, _ := filepath.Glob(filepath.Join(filepath.Dir(path), ".wisirc-update-*"))
	if len(left) != 0 {
		t.Fatalf("temp files left: %v", left)
	}
}

func TestApplyRejectsBadChecksum(t *testing.T) {
	u := fakeGitHub(t, "v0.5.0", []byte("tampered"), sum([]byte("real"))+"  wisirc-linux-amd64\n")
	if _, err := u.Apply(context.Background(), "v0.5.0"); err == nil || !strings.Contains(err.Error(), "SHA256SUMS") {
		t.Fatalf("err %v", err)
	}
	exe, _ := u.Exe()
	if got, _ := os.ReadFile(exe); string(got) != "old" {
		t.Fatal("binary changed after a bad checksum")
	}
}

func TestApplyKeepsOldWhenNewDoesNotRun(t *testing.T) {
	bin := []byte("broken")
	u := fakeGitHub(t, "v0.5.0", bin, sum(bin)+"  wisirc-linux-amd64\n")
	u.Verify = func(string, string) error { return errors.New("does not start") }
	if _, err := u.Apply(context.Background(), "v0.5.0"); err == nil {
		t.Fatal("applied a binary that does not run")
	}
	exe, _ := u.Exe()
	if got, _ := os.ReadFile(exe); string(got) != "old" {
		t.Fatal("binary changed")
	}
}

func TestApplyRefusesHomebrew(t *testing.T) {
	u := fakeGitHub(t, "v0.5.0", nil, "")
	u.Exe = func() (string, error) { return "/opt/homebrew/Cellar/wisirc/0.4.0/bin/wisirc", nil }
	if _, err := u.Apply(context.Background(), "v0.5.0"); !errors.Is(err, ErrHomebrew) {
		t.Fatalf("err %v", err)
	}
}

func TestApplyUnsupportedSystem(t *testing.T) {
	u := fakeGitHub(t, "v0.5.0", nil, "")
	u.Asset = "wisirc-windows-amd64"
	if _, err := u.Apply(context.Background(), "v0.5.0"); err == nil {
		t.Fatal("no error for a system without binaries")
	}
}
