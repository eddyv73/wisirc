package config

import (
	"bytes"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

const serversHeader = `# Servers saved from wisirc (≡ → Servers, or when connecting with /connect).
# You can edit it by hand; if a name is also in config.toml, config.toml wins.

`

// ServersPath returns the servers.toml that sits next to the config file.
func ServersPath(cfgPath string) string {
	return filepath.Join(filepath.Dir(cfgPath), "servers.toml")
}

// SaveServers replaces servers.toml with servers, atomically and mode 0600.
func SaveServers(cfgPath string, servers map[string]Server) error {
	var buf bytes.Buffer
	buf.WriteString(serversHeader)
	if len(servers) > 0 {
		if err := toml.NewEncoder(&buf).Encode(struct {
			Servers map[string]Server `toml:"servers"`
		}{servers}); err != nil {
			return err
		}
	}
	return writeAtomic(ServersPath(cfgPath), buf.Bytes())
}

// writeAtomic replaces path with data through a temp file, mode 0600.
func writeAtomic(path string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), filepath.Base(path)+".tmp*")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	if err := tmp.Chmod(0o600); err != nil {
		tmp.Close()
		return err
	}
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), path)
}
