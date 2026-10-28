// Package halloy reads the servers from a Halloy IRC client config so they
// can be imported into wisirc.
package halloy

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/eddyv73/wisirc/internal/config"
	"github.com/eddyv73/wisirc/internal/i18n"
)

type file struct {
	Servers map[string]server `toml:"servers"`
}

type server struct {
	Nickname        string   `toml:"nickname"`
	Username        string   `toml:"username"`
	Realname        string   `toml:"realname"`
	Server          string   `toml:"server"`
	Port            int      `toml:"port"`
	UseTLS          *bool    `toml:"use_tls"`
	Channels        []string `toml:"channels"`
	AcceptInvalid   bool     `toml:"dangerously_accept_invalid_certs"`
	Password        string   `toml:"password"`
	PasswordFile    string   `toml:"password_file"`
	PasswordCommand string   `toml:"password_command"`
	NickPassword    string   `toml:"nick_password"`
	SASL            *struct {
		Plain *struct {
			Username        string `toml:"username"`
			Password        string `toml:"password"`
			PasswordFile    string `toml:"password_file"`
			PasswordCommand string `toml:"password_command"`
		} `toml:"plain"`
		External *struct{} `toml:"external"`
	} `toml:"sasl"`
}

// Paths lists where Halloy keeps its config on goos, most likely first.
func Paths(goos, home, xdg string) []string {
	switch goos {
	case "darwin":
		return []string{
			filepath.Join(home, "Library", "Application Support", "halloy", "config.toml"),
			filepath.Join(home, ".config", "halloy", "config.toml"),
		}
	case "windows":
		return []string{filepath.Join(os.Getenv("AppData"), "halloy", "config.toml")}
	}
	var out []string
	if xdg != "" {
		out = append(out, filepath.Join(xdg, "halloy", "config.toml"))
	}
	return append(out,
		filepath.Join(home, ".config", "halloy", "config.toml"),
		filepath.Join(home, ".var", "app", "org.squidowl.halloy", "config", "config.toml"))
}

// Find returns the first Halloy config that exists on this machine.
func Find() (string, bool) {
	home, _ := os.UserHomeDir()
	for _, p := range Paths(runtime.GOOS, home, os.Getenv("XDG_CONFIG_HOME")) {
		if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
			return p, true
		}
	}
	return "", false
}

// Load converts the servers in a Halloy config. Settings wisirc cannot use
// are reported as warnings; password_command is never executed.
func Load(path string) (map[string]config.Server, []string, error) {
	var f file
	if _, err := toml.DecodeFile(path, &f); err != nil {
		return nil, nil, err
	}
	dir := filepath.Dir(path)
	readSecret := func(p string) (string, error) {
		if !filepath.IsAbs(p) {
			p = filepath.Join(dir, p)
		}
		b, err := os.ReadFile(p)
		return strings.TrimRight(string(b), "\r\n"), err
	}
	out := map[string]config.Server{}
	var warns []string
	warn := func(name, msg string) { warns = append(warns, name+": "+msg) }
	for _, name := range sortedKeys(f.Servers) {
		h := f.Servers[name]
		if h.Server == "" {
			warn(name, i18n.T("no server, not imported"))
			continue
		}
		tlsOn := h.UseTLS == nil || *h.UseTLS // Halloy uses TLS unless told otherwise
		s := config.Server{
			Nickname: h.Nickname, Username: h.Username, Realname: h.Realname,
			Server: h.Server, Port: h.Port, UseTLS: &tlsOn, Channels: h.Channels,
			DangerouslyAcceptInvalidCerts: h.AcceptInvalid, Password: h.Password,
		}
		if s.Port == 0 {
			s.Port = 6697
			if !tlsOn {
				s.Port = 6667
			}
		}
		if s.Password == "" && h.PasswordFile != "" {
			if pw, err := readSecret(h.PasswordFile); err == nil {
				s.Password = pw
			} else {
				warn(name, i18n.T("password_file could not be read; set the password in ≡ → Servers"))
			}
		}
		if h.PasswordCommand != "" {
			warn(name, i18n.T("password_command is not run for security; set the password in ≡ → Servers"))
		}
		if h.NickPassword != "" {
			warn(name, i18n.T("nick_password is not supported; use SASL or /msg NickServ identify"))
		}
		if h.SASL != nil && h.SASL.External != nil {
			warn(name, i18n.T("sasl.external (certificate) is not supported"))
		}
		if h.SASL != nil && h.SASL.Plain != nil {
			p := h.SASL.Plain
			pw := p.Password
			if pw == "" && p.PasswordFile != "" {
				var err error
				if pw, err = readSecret(p.PasswordFile); err != nil {
					warn(name, i18n.T("sasl.plain.password_file could not be read; set the password in ≡ → Servers"))
				}
			}
			if p.PasswordCommand != "" {
				warn(name, i18n.T("sasl.plain.password_command is not run for security; set the password in ≡ → Servers"))
			}
			if pw != "" {
				s.SASL = &config.SASL{Plain: &config.SASLPlain{Username: p.Username, Password: pw}}
			}
		}
		out[name] = s
	}
	if len(out) == 0 && len(warns) == 0 {
		return nil, nil, errors.New(i18n.T("the Halloy config has no servers"))
	}
	return out, warns, nil
}

func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
