package ui

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

// exportName is the file /export halloy writes next to the config.
const exportName = "halloy-export.toml"

// exportHalloy writes the saved servers (config.toml and servers.toml, not
// unsaved /connect ones) as a Halloy config. It never replaces a file.
func (m *Model) exportHalloy(path string) {
	path = m.exportPath(path)
	var servers map[string]config.Server
	if cfg := m.d.Nets.Config(); cfg != nil {
		servers = cfg.Servers
	}
	data, err := halloy.Export(servers, func(name string) bool { return m.disabled[name] })
	if err != nil {
		m.errorLine(err.Error())
		return
	}
	if err := halloy.WriteNew(path, data); err != nil {
		if errors.Is(err, fs.ErrExist) {
			m.errorLine(i18n.T("%s already exists; pick another file: /export halloy <file>", path))
		} else {
			m.errorLine(err.Error())
		}
		return
	}
	m.info(i18n.T("Exported %d server(s) to %s", len(servers), path))
	m.info(i18n.T("It holds your passwords in plain text: keep it private. Copy its [servers.*] into Halloy's config.toml."))
}

// exportPath resolves the file for /export: the default one next to the
// config when empty, with a leading ~ meaning the home folder.
func (m *Model) exportPath(path string) string {
	if path == "" {
		cfg := m.d.CfgPath
		if cfg == "" {
			cfg = config.DefaultPath()
		}
		return filepath.Join(filepath.Dir(cfg), exportName)
	}
	if path == "~" || strings.HasPrefix(path, "~/") {
		if home, err := os.UserHomeDir(); err == nil {
			path = filepath.Join(home, strings.TrimPrefix(path, "~"))
		}
	}
	return path
}
