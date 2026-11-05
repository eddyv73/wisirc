// Package update finds newer wisirc releases on GitHub and installs them over
// the running binary, the same way install.sh does: the release binary for
// this system, checked against the release's SHA256SUMS.
package update

import (
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/eddyv73/wisirc/internal/i18n"
)

// Repo is where releases are published.
const Repo = "eddyv73/wisirc"

const maxBinary = 64 << 20

// ErrHomebrew means the binary belongs to Homebrew, which must upgrade it.
var ErrHomebrew = errors.New("installed with Homebrew")

// Updater talks to GitHub. The zero value is not usable; use New.
type Updater struct {
	API      string // GitHub API root
	Download string // GitHub web root (release downloads)
	Repo     string
	HTTP     *http.Client
	Exe      func() (string, error)       // the binary to replace
	Verify   func(path, tag string) error // checks the new binary runs
	Asset    string                       // release file for this system
}

// New returns an Updater for the public wisirc releases.
func New() *Updater {
	return &Updater{
		API:      "https://api.github.com",
		Download: "https://github.com",
		Repo:     Repo,
		HTTP:     &http.Client{Timeout: 2 * time.Minute},
		Exe:      os.Executable,
		Verify:   runsAs,
		Asset:    "wisirc-" + runtime.GOOS + "-" + runtime.GOARCH,
	}
}

// Latest returns the tag of the newest stable release.
func (u *Updater) Latest(ctx context.Context) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.API+"/repos/"+u.Repo+"/releases/latest", nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	resp, err := u.HTTP.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GitHub: %s", resp.Status)
	}
	var rel struct {
		Tag string `json:"tag_name"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&rel); err != nil {
		return "", err
	}
	if _, ok := parse(rel.Tag); !ok {
		return "", fmt.Errorf("GitHub: unexpected tag %q", rel.Tag)
	}
	return rel.Tag, nil
}

// Apply downloads tag for this system, checks it against SHA256SUMS and
// replaces the running binary. It returns the path it replaced; the new
// version runs the next time wisirc starts.
func (u *Updater) Apply(ctx context.Context, tag string) (string, error) {
	if _, ok := parse(tag); !ok {
		return "", fmt.Errorf("bad version %q", tag)
	}
	switch u.Asset {
	case "wisirc-darwin-arm64", "wisirc-darwin-amd64", "wisirc-linux-arm64", "wisirc-linux-amd64":
	default:
		return "", errors.New(i18n.T("There is no release binary for this system; build it from source"))
	}
	exe, err := u.Exe()
	if err != nil {
		return "", err
	}
	if real, err := filepath.EvalSymlinks(exe); err == nil {
		exe = real
	}
	if strings.Contains(exe, "/Cellar/") {
		return "", ErrHomebrew
	}
	base := u.Download + "/" + u.Repo + "/releases/download/" + tag + "/"
	sums, err := u.get(ctx, base+"SHA256SUMS", 1<<20)
	if err != nil {
		return "", err
	}
	want, ok := sumFor(sums, u.Asset)
	if !ok {
		return "", fmt.Errorf("SHA256SUMS has no %s", u.Asset)
	}
	bin, err := u.get(ctx, base+u.Asset, maxBinary)
	if err != nil {
		return "", err
	}
	if got := sha256.Sum256(bin); hex.EncodeToString(got[:]) != want {
		return "", errors.New(i18n.T("The download does not match SHA256SUMS; nothing was changed"))
	}
	tmp, err := os.CreateTemp(filepath.Dir(exe), ".wisirc-update-*")
	if err != nil {
		return "", errors.New(i18n.T("Cannot write next to %s: %s", exe, err.Error()))
	}
	defer os.Remove(tmp.Name())
	if _, err := tmp.Write(bin); err != nil {
		tmp.Close()
		return "", err
	}
	if err := tmp.Close(); err != nil {
		return "", err
	}
	if err := os.Chmod(tmp.Name(), 0o755); err != nil {
		return "", err
	}
	if err := u.Verify(tmp.Name(), tag); err != nil {
		return "", err
	}
	if err := os.Rename(tmp.Name(), exe); err != nil {
		return "", err
	}
	return exe, nil
}

func (u *Updater) get(ctx context.Context, url string, limit int64) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := u.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s: %s", url, resp.Status)
	}
	b, err := io.ReadAll(io.LimitReader(resp.Body, limit+1))
	if err != nil {
		return nil, err
	}
	if int64(len(b)) > limit {
		return nil, fmt.Errorf("%s: too large", url)
	}
	return b, nil
}

func sumFor(sums []byte, name string) (string, bool) {
	sc := bufio.NewScanner(strings.NewReader(string(sums)))
	for sc.Scan() {
		f := strings.Fields(sc.Text())
		if len(f) == 2 && strings.TrimPrefix(f[1], "*") == name && len(f[0]) == 64 {
			return strings.ToLower(f[0]), true
		}
	}
	return "", false
}

// runsAs checks that the downloaded binary starts and reports tag.
func runsAs(path, tag string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	out, err := exec.CommandContext(ctx, path, "--version").Output()
	if err != nil || !strings.Contains(string(out), tag) {
		return errors.New(i18n.T("The new binary did not start; nothing was changed"))
	}
	return nil
}

type version struct {
	core [3]int
	pre  string // "rc.1"; empty for a release
	dev  bool   // built after the tag (git describe: v0.3.0-3-gabc123)
}

var describe = regexp.MustCompile(`^\d+-g[0-9a-f]+(-dirty)?$`)

func parse(s string) (version, bool) {
	var v version
	s = strings.TrimPrefix(s, "v")
	core, pre, _ := strings.Cut(s, "-")
	parts := strings.Split(core, ".")
	if len(parts) != 3 {
		return v, false
	}
	for i, p := range parts {
		n, err := strconv.Atoi(p)
		if err != nil || n < 0 {
			return v, false
		}
		v.core[i] = n
	}
	if describe.MatchString(pre) {
		v.dev = true
	} else {
		v.pre = pre
	}
	return v, true
}

// Newer reports whether latest is a newer release than current. Builds that
// are not releases ("dev") never get offered updates.
func Newer(current, latest string) bool {
	c, ok1 := parse(current)
	l, ok2 := parse(latest)
	if !ok1 || !ok2 || l.pre != "" {
		return false
	}
	for i := range 3 {
		if l.core[i] != c.core[i] {
			return l.core[i] > c.core[i]
		}
	}
	// same numbers: a release candidate is older than its release
	return c.pre != "" && !c.dev
}
