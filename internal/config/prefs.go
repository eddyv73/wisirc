package config

import (
	"bytes"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// Prefs are choices made from inside wisirc (not hand-edited config).
type Prefs struct {
	Theme string `toml:"theme,omitempty"`
	Lang  string `toml:"lang,omitempty"`

	// Nick and Realname are the identity from the first-run setup: the
	// default for /connect and for new servers.
	Nick     string `toml:"nick,omitempty"`
	Realname string `toml:"realname,omitempty"`

	// ServerColors maps a server name to the "#rrggbb" picked for it.
	ServerColors map[string]string `toml:"server_colors,omitempty"`

	// Disabled lists servers that do not connect when wisirc starts.
	Disabled []string `toml:"disabled_servers,omitempty"`
}

// PrefsPath returns the prefs.toml that sits next to the config file.
func PrefsPath(cfgPath string) string {
	return filepath.Join(filepath.Dir(cfgPath), "prefs.toml")
}

// LoadPrefs reads prefs.toml; a missing or broken file gives empty prefs.
func LoadPrefs(cfgPath string) Prefs {
	var p Prefs
	if _, err := toml.DecodeFile(PrefsPath(cfgPath), &p); err != nil {
		return Prefs{}
	}
	return p
}

// SavePrefs replaces prefs.toml, atomically and mode 0600.
func SavePrefs(cfgPath string, p Prefs) error {
	var buf bytes.Buffer
	buf.WriteString("# Preferences chosen inside wisirc (identity, theme, language, server colors, disabled servers). They win over config.toml.\n\n")
	if err := toml.NewEncoder(&buf).Encode(p); err != nil {
		return err
	}
	return writeAtomic(PrefsPath(cfgPath), buf.Bytes())
}
