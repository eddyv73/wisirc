// Package config loads the wisirc TOML configuration.
package config

import (
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"

	"github.com/BurntSushi/toml"
	"github.com/eddyv73/wisirc/internal/i18n"
)

type Config struct {
	Servers       map[string]Server `toml:"servers"`
	Notifications Notifications     `toml:"notifications"`
	Theme         Theme             `toml:"theme"`

	Path    string          `toml:"-"` // config.toml this was loaded from
	Managed map[string]bool `toml:"-"` // servers that live in servers.toml
}

type Server struct {
	Nickname                      string   `toml:"nickname"`
	Username                      string   `toml:"username,omitempty"`
	Realname                      string   `toml:"realname,omitempty"`
	Server                        string   `toml:"server"`
	Port                          int      `toml:"port"`
	UseTLS                        *bool    `toml:"use_tls,omitempty"`
	Channels                      []string `toml:"channels,omitempty"`
	DangerouslyAcceptInvalidCerts bool     `toml:"dangerously_accept_invalid_certs,omitempty"`
	Password                      string   `toml:"password,omitempty"`
	SASL                          *SASL    `toml:"sasl,omitempty"`
	AutoConnect                   *bool    `toml:"autoconnect,omitempty"`
}

// Auto reports whether the server connects at startup (default true).
func (s Server) Auto() bool { return s.AutoConnect == nil || *s.AutoConnect }

type SASL struct {
	Plain *SASLPlain `toml:"plain"`
}

type SASLPlain struct {
	Username string `toml:"username"`
	Password string `toml:"password"`
}

type Notifications struct {
	Enabled *bool `toml:"enabled"`
}

// On reports whether notifications are enabled (default true).
func (n Notifications) On() bool { return n.Enabled == nil || *n.Enabled }

type Theme struct {
	Name string `toml:"name"`
}

// TLS reports whether the connection should use TLS.
func (s Server) TLS() bool {
	if s.UseTLS != nil {
		return *s.UseTLS
	}
	return s.Port == 6697 || s.Port == 0
}

// Real returns the realname, defaulting to the nickname.
func (s Server) Real() string {
	if s.Realname != "" {
		return s.Realname
	}
	return s.Nickname
}

// Addr returns host:port, defaulting the port by TLS mode.
func (s Server) Addr() string {
	port := s.Port
	if port == 0 {
		if s.TLS() {
			port = 6697
		} else {
			port = 6667
		}
	}
	return net.JoinHostPort(s.Server, strconv.Itoa(port))
}

// Load parses the config at path plus the servers.toml next to it. Unknown
// keys are returned as warnings; config.toml wins when a name is in both.
func Load(path string) (*Config, []string, error) {
	var cfg Config
	warnings, err := decode(path, &cfg)
	if err != nil {
		return nil, nil, err
	}
	if cfg.Servers == nil {
		cfg.Servers = map[string]Server{}
	}
	cfg.Path = path
	cfg.Managed = map[string]bool{}

	var managed Config
	mw, err := decode(ServersPath(path), &managed)
	if errors.Is(err, os.ErrNotExist) {
		return &cfg, warnings, nil
	}
	if err != nil {
		return nil, nil, err
	}
	for _, w := range mw {
		warnings = append(warnings, "servers.toml: "+w)
	}
	for name, s := range managed.Servers {
		if _, dup := cfg.Servers[name]; dup {
			warnings = append(warnings, "servers.toml: "+i18n.T("%s is also in config.toml (config.toml is used)", name))
			continue
		}
		cfg.Servers[name] = s
		cfg.Managed[name] = true
	}
	return &cfg, warnings, nil
}

// decode reads one TOML file into cfg and lists its unknown keys.
func decode(path string, cfg *Config) ([]string, error) {
	if _, err := os.Stat(path); err != nil {
		return nil, err
	}
	md, err := toml.DecodeFile(path, cfg)
	if err != nil {
		var pe toml.ParseError
		if errors.As(err, &pe) {
			return nil, fmt.Errorf("%s:%d: %s", path, pe.Position.Line, pe.Message)
		}
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	var warnings []string
	// Undecoded lists a table and every key under it; report only the outermost.
	var reported []toml.Key
	for _, k := range md.Undecoded() {
		if hasPrefixKey(k, reported) {
			continue
		}
		reported = append(reported, k)
		warnings = append(warnings, i18n.T("unknown key ignored: %s", k.String()))
	}
	return warnings, nil
}

func hasPrefixKey(k toml.Key, prefixes []toml.Key) bool {
	for _, p := range prefixes {
		if len(p) < len(k) {
			match := true
			for i := range p {
				if p[i] != k[i] {
					match = false
					break
				}
			}
			if match {
				return true
			}
		}
	}
	return false
}

// DefaultPath returns $XDG_CONFIG_HOME/wisirc/config.toml or ~/.config/wisirc/config.toml.
func DefaultPath() string {
	base := os.Getenv("XDG_CONFIG_HOME")
	if base == "" {
		home, _ := os.UserHomeDir()
		base = filepath.Join(home, ".config")
	}
	return filepath.Join(base, "wisirc", "config.toml")
}

// EnsureExample writes the example config if path does not exist.
func EnsureExample(path string) (bool, error) { return ensure(path, exampleConfig) }

// EnsureStarter writes a config without servers if path does not exist, for
// a first run that imports them instead (from Halloy).
func EnsureStarter(path string) (bool, error) { return ensure(path, starterConfig) }

func ensure(path, content string) (bool, error) {
	if _, err := os.Stat(path); err == nil {
		return false, nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return false, err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return false, err
	}
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		return false, err
	}
	return true, nil
}
