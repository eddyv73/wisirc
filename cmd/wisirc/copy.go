package main

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/eddyv73/wisirc/internal/config"
	"github.com/eddyv73/wisirc/internal/halloy"
	"github.com/eddyv73/wisirc/internal/i18n"
)

// copyConfig clones the config at src so wisirc starts with it at dst. src is
// a wisirc config.toml (its servers.toml and prefs.toml come along), a Halloy
// config.toml (its servers go to servers.toml), or the directory holding
// either. It never replaces anything: dst must not have a config yet.
// It returns what to tell the user and the Halloy import warnings.
func copyConfig(src, dst string) (string, []string, error) {
	src = expandHome(src)
	st, err := os.Stat(src)
	if err != nil {
		return "", nil, err
	}
	if st.IsDir() {
		src = filepath.Join(src, "config.toml")
	}
	targets := []string{dst, config.ServersPath(dst), config.PrefsPath(dst)}
	for _, p := range targets {
		if _, err := os.Stat(p); err == nil {
			return "", nil, errors.New(i18n.T("%s already exists; wisirc never replaces a config: move it away or choose another one with --config", p))
		} else if !errors.Is(err, fs.ErrNotExist) {
			return "", nil, err
		}
	}

	cfg, warns, loadErr := config.Load(src)
	_, serversErr := os.Stat(config.ServersPath(src))
	_, prefsErr := os.Stat(config.PrefsPath(src))
	// a Halloy config has keys wisirc does not know (or a theme it cannot
	// read); a wisirc one has none, or has wisirc's files next to it
	isWisirc := loadErr == nil && (len(warns) == 0 || serversErr == nil || prefsErr == nil)
	if !isWisirc {
		servers, hw, err := halloy.Load(src)
		if err == nil && len(servers) > 0 {
			// config.toml goes last: if anything fails, the next run starts clean
			if err := config.SaveServers(dst, servers); err != nil {
				return "", nil, err
			}
			if _, err := config.EnsureStarter(dst); err != nil {
				return "", nil, err
			}
			return i18n.T("Copied %d server(s) from the Halloy config %s", len(servers), src), hw, nil
		}
		if loadErr != nil {
			return "", nil, loadErr
		}
	}

	for _, pair := range [][2]string{
		{config.ServersPath(src), config.ServersPath(dst)},
		{config.PrefsPath(src), config.PrefsPath(dst)},
		{src, dst}, // last, for the same reason
	} {
		if err := copyFile(pair[0], pair[1]); err != nil && !(pair[0] != src && errors.Is(err, fs.ErrNotExist)) {
			return "", nil, err
		}
	}
	return i18n.T("Copied the wisirc config from %s (%d server(s))", src, len(cfg.Servers)), nil, nil
}

// copyFile copies src to a new file dst readable only by its owner.
func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0o700); err != nil {
		return err
	}
	f, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return err
	}
	if _, err := f.Write(data); err != nil {
		f.Close()
		os.Remove(dst)
		return err
	}
	return f.Close()
}

func expandHome(p string) string {
	if p == "~" || strings.HasPrefix(p, "~/") {
		if home, err := os.UserHomeDir(); err == nil {
			return filepath.Join(home, p[1:])
		}
	}
	return p
}
