package halloy

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/eddyv73/wisirc/internal/config"
	"github.com/eddyv73/wisirc/internal/i18n"
)

// exportServer is a Halloy [servers.<name>] table. Port, TLS and autoconnect
// are always written, so the file does not depend on either client's defaults.
type exportServer struct {
	Nickname      string      `toml:"nickname,omitempty"`
	Username      string      `toml:"username,omitempty"`
	Realname      string      `toml:"realname,omitempty"`
	Server        string      `toml:"server"`
	Port          int         `toml:"port"`
	UseTLS        bool        `toml:"use_tls"`
	AcceptInvalid bool        `toml:"dangerously_accept_invalid_certs,omitempty"`
	Password      string      `toml:"password,omitempty"`
	Channels      []string    `toml:"channels,omitempty"`
	Autoconnect   bool        `toml:"autoconnect"`
	SASL          *exportSASL `toml:"sasl,omitempty"`
}

type exportSASL struct {
	Plain exportPlain `toml:"plain"`
}

type exportPlain struct {
	Username string `toml:"username"`
	Password string `toml:"password"`
}

// header goes on top of every export. It stays in English like the rest of
// the file: it is a config file, not UI text.
const header = `# Halloy servers exported from wisirc.
# This file holds passwords in plain text: keep it private.
# Copy the [servers.*] tables into Halloy's config.toml.

`

// Export renders servers as Halloy [servers.<name>] tables. A server that
// is disabled (or has autoconnect = false) gets autoconnect = false.
func Export(servers map[string]config.Server, disabled func(name string) bool) ([]byte, error) {
	if len(servers) == 0 {
		return nil, errors.New(i18n.T("there are no servers to export"))
	}
	out := struct {
		Servers map[string]exportServer `toml:"servers"`
	}{Servers: map[string]exportServer{}}
	for name, s := range servers {
		port := s.Port
		if port == 0 {
			port = 6667
			if s.TLS() {
				port = 6697
			}
		}
		e := exportServer{
			Nickname: s.Nickname, Username: s.Username, Realname: s.Realname,
			Server: s.Server, Port: port, UseTLS: s.TLS(),
			AcceptInvalid: s.DangerouslyAcceptInvalidCerts, Password: s.Password,
			Channels:    s.Channels,
			Autoconnect: s.Auto() && (disabled == nil || !disabled(name)),
		}
		if s.SASL != nil && s.SASL.Plain != nil {
			e.SASL = &exportSASL{Plain: exportPlain{Username: s.SASL.Plain.Username, Password: s.SASL.Plain.Password}}
		}
		out.Servers[name] = e
	}
	var body bytes.Buffer
	enc := toml.NewEncoder(&body)
	enc.Indent = "" // flat, like Halloy's own docs; dropEmptyTables relies on it
	if err := enc.Encode(out); err != nil {
		return nil, err
	}
	return append([]byte(header), dropEmptyTables(body.Bytes())...), nil
}

// dropEmptyTables removes headers like [servers] or [servers.x.sasl] that
// only open the table below them. They are implied by [servers.x] and
// [servers.x.sasl.plain], and an explicit [servers] would clash with the
// one in the Halloy config this file gets pasted into.
func dropEmptyTables(b []byte) []byte {
	lines := strings.Split(string(b), "\n")
	var out []string
	for i, l := range lines {
		if strings.HasPrefix(l, "[") {
			next := ""
			for _, n := range lines[i+1:] {
				if strings.TrimSpace(n) != "" {
					next = n
					break
				}
			}
			if strings.HasPrefix(next, strings.TrimSuffix(l, "]")+".") {
				continue
			}
			if len(out) > 0 && out[len(out)-1] != "" {
				out = append(out, "") // a blank line before each table
			}
		}
		out = append(out, l)
	}
	return []byte(strings.Join(out, "\n"))
}

// WriteNew writes data to a new file readable only by its owner. It never
// replaces an existing file: the error then wraps fs.ErrExist.
func WriteNew(path string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return err
	}
	if _, err := f.Write(data); err != nil {
		f.Close()
		os.Remove(path)
		return err
	}
	return f.Close()
}
