// Package app wires configured servers to IRC clients for the UI.
package app

import (
	"context"
	"errors"
	"net"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/eddyv73/wisirc/internal/config"
	"github.com/eddyv73/wisirc/internal/i18n"
	"github.com/eddyv73/wisirc/internal/irc"
	"github.com/eddyv73/wisirc/internal/ui"
)

// DialFactory builds a dialer for a client; nil means the real TCP/TLS dialer.
type DialFactory func(irc.Options) func(context.Context) (net.Conn, error)

type entry struct {
	client *irc.Client
	cancel context.CancelFunc
}

// Manager owns one irc.Client per network and implements ui.Networks.
type Manager struct {
	ctx    context.Context
	dial   DialFactory
	events chan irc.Event

	mu      sync.Mutex
	cfg     *config.Config
	clients map[string]*entry
	adhoc   map[string]config.Server // started with /connect options, not saved
}

var _ ui.Networks = (*Manager)(nil)

func NewManager(ctx context.Context, cfg *config.Config, dial DialFactory) *Manager {
	if cfg == nil {
		cfg = &config.Config{}
	}
	return &Manager{ctx: ctx, dial: dial, cfg: cfg, events: make(chan irc.Event, 256), clients: map[string]*entry{}, adhoc: map[string]config.Server{}}
}

func (m *Manager) Events() <-chan irc.Event { return m.events }

func optionsFor(name string, s config.Server) irc.Options {
	o := irc.Options{
		Name: name, Addr: s.Addr(), TLS: s.TLS(), InsecureTLS: s.DangerouslyAcceptInvalidCerts,
		Nick: s.Nickname, User: s.Username, Real: s.Real(), Pass: s.Password, Channels: s.Channels,
	}
	if o.User == "" {
		o.User = s.Nickname
	}
	if s.SASL != nil && s.SASL.Plain != nil {
		o.SASLUser, o.SASLPass = s.SASL.Plain.Username, s.SASL.Plain.Password
	}
	return o
}

// Start launches a client for the named server (no-op if already running).
func (m *Manager) Start(name string, s config.Server) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.clients[name]; ok {
		return
	}
	opts := optionsFor(name, s)
	if m.dial != nil {
		opts.Dial = m.dial(opts)
	}
	ctx, cancel := context.WithCancel(m.ctx)
	c := irc.New(opts, m.events)
	e := &entry{client: c, cancel: cancel}
	m.clients[name] = e
	go func() {
		c.Run(ctx)
		// a client that gave up (e.g. SASL failure) can be started again
		m.mu.Lock()
		if m.clients[name] == e {
			delete(m.clients, name)
		}
		m.mu.Unlock()
		cancel()
	}()
}

func (m *Manager) Get(name string) (ui.Sender, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	e, ok := m.clients[name]
	if !ok {
		return nil, false
	}
	return e.client, true
}

func (m *Manager) Names() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	names := make([]string, 0, len(m.clients))
	for n := range m.clients {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}

func (m *Manager) Disconnect(name string) {
	m.mu.Lock()
	e, ok := m.clients[name]
	delete(m.clients, name)
	m.mu.Unlock()
	if !ok {
		return
	}
	e.client.Send("QUIT :wisirc")
	e.cancel()
}

// Reload starts autoconnect servers that are new in cfg and leaves the
// others alone.
func (m *Manager) Reload(cfg *config.Config) []string {
	m.mu.Lock()
	old := m.cfg
	m.cfg = cfg
	var fresh []string
	for name, s := range cfg.Servers {
		_, running := m.clients[name]
		_, known := old.Servers[name]
		if !running && !known && s.Auto() {
			fresh = append(fresh, name)
		}
	}
	m.mu.Unlock()
	sort.Strings(fresh)
	for _, name := range fresh {
		m.Start(name, cfg.Servers[name])
	}
	return fresh
}

// Connect starts a configured server by name, or an ad-hoc server:
//
//	host [port] [-tls|-notls] [-insecure] [-nick N] [-user U] [-pass P]
//	           [-sasl user:password] [-join #a,#b] [-name name]
func (m *Manager) Connect(arg string) (string, error) {
	fields := strings.Fields(arg)
	if len(fields) == 0 {
		return "", errors.New(i18n.T("Usage: /connect <name|host> [port] [options]"))
	}
	m.mu.Lock()
	s, configured := m.cfg.Servers[fields[0]]
	m.mu.Unlock()
	if configured && len(fields) == 1 {
		m.Start(fields[0], s)
		return fields[0], nil
	}
	srv, name, err := m.adHoc(fields)
	if err != nil {
		return "", err
	}
	m.mu.Lock()
	if _, saved := m.cfg.Servers[name]; !saved {
		m.adhoc[name] = srv
	}
	m.mu.Unlock()
	m.Start(name, srv)
	return name, nil
}

// Config returns the configuration the manager currently uses.
func (m *Manager) Config() *config.Config {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.cfg
}

// IsAdHoc reports whether name was connected with /connect and not saved.
func (m *Manager) IsAdHoc(name string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	_, ok := m.adhoc[name]
	return ok
}

// SaveAdHoc stores an ad-hoc connection in servers.toml.
func (m *Manager) SaveAdHoc(name string) error {
	m.mu.Lock()
	s, ok := m.adhoc[name]
	m.mu.Unlock()
	if !ok {
		return errors.New(i18n.T("%s is not a new connection", name))
	}
	return m.SaveServer("", name, s)
}

// SaveServer creates (old == "") or replaces the servers.toml entry old,
// stored as name. Servers from config.toml are read-only here.
func (m *Manager) SaveServer(old, name string, s config.Server) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if name == "" {
		return errors.New(i18n.T("The server needs a name"))
	}
	if old != "" && !m.cfg.Managed[old] {
		if _, ok := m.cfg.Servers[old]; ok {
			return errors.New(i18n.T("%s is in config.toml: edit it with ≡ → Edit config", old))
		}
	}
	if _, taken := m.cfg.Servers[name]; taken && name != old {
		return errors.New(i18n.T("A server named %s already exists", name))
	}
	managed := m.managedLocked()
	delete(managed, old)
	managed[name] = s
	if err := m.writeLocked(managed); err != nil {
		return err
	}
	delete(m.cfg.Servers, old)
	delete(m.cfg.Managed, old)
	delete(m.adhoc, old)
	delete(m.adhoc, name)
	m.cfg.Servers[name] = s
	m.cfg.Managed[name] = true
	return nil
}

// Import saves every server whose name is free into servers.toml in one
// write; names that already exist are skipped, never overwritten.
func (m *Manager) Import(servers map[string]config.Server) (imported, skipped []string, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	managed := m.managedLocked()
	names := make([]string, 0, len(servers))
	for n := range servers {
		names = append(names, n)
	}
	sort.Strings(names)
	for _, n := range names {
		if _, taken := m.cfg.Servers[n]; taken {
			skipped = append(skipped, n)
			continue
		}
		managed[n] = servers[n]
		imported = append(imported, n)
	}
	if len(imported) == 0 {
		return nil, skipped, nil
	}
	if err := m.writeLocked(managed); err != nil {
		return nil, nil, err
	}
	for _, n := range imported {
		m.cfg.Servers[n] = servers[n]
		m.cfg.Managed[n] = true
		delete(m.adhoc, n)
	}
	return imported, skipped, nil
}

// DeleteServer removes a servers.toml entry and disconnects it.
func (m *Manager) DeleteServer(name string) error {
	m.mu.Lock()
	if !m.cfg.Managed[name] {
		m.mu.Unlock()
		if _, ok := m.cfg.Servers[name]; ok {
			return errors.New(i18n.T("%s is in config.toml: delete it with ≡ → Edit config", name))
		}
		return errors.New(i18n.T("There is no server %s", name))
	}
	managed := m.managedLocked()
	delete(managed, name)
	if err := m.writeLocked(managed); err != nil {
		m.mu.Unlock()
		return err
	}
	delete(m.cfg.Servers, name)
	delete(m.cfg.Managed, name)
	m.mu.Unlock()
	m.Disconnect(name)
	return nil
}

func (m *Manager) managedLocked() map[string]config.Server {
	out := map[string]config.Server{}
	for n := range m.cfg.Managed {
		out[n] = m.cfg.Servers[n]
	}
	return out
}

func (m *Manager) writeLocked(managed map[string]config.Server) error {
	if m.cfg.Path == "" {
		return errors.New(i18n.T("I don't know where the config is"))
	}
	if m.cfg.Managed == nil {
		m.cfg.Managed = map[string]bool{}
	}
	if err := config.SaveServers(m.cfg.Path, managed); err != nil {
		return errors.New(i18n.T("Could not save servers.toml: %s", err.Error()))
	}
	return nil
}

// adHoc builds a server from "/connect host [port] [options]" fields.
func (m *Manager) adHoc(fields []string) (config.Server, string, error) {
	srv := config.Server{Server: fields[0], Port: 6697}
	name := fields[0]
	var tlsOn *bool
	portSet := false
	for i := 1; i < len(fields); i++ {
		f := fields[i]
		if !strings.HasPrefix(f, "-") {
			if portSet {
				return srv, "", errors.New(i18n.T("Extra argument: %s", f))
			}
			p, err := strconv.Atoi(f)
			if err != nil || p <= 0 || p > 65535 {
				return srv, "", errors.New(i18n.T("Invalid port: %s", f))
			}
			srv.Port, portSet = p, true
			continue
		}
		switch f {
		case "-tls", "-notls":
			on := f == "-tls"
			tlsOn = &on
			continue
		case "-insecure":
			srv.DangerouslyAcceptInvalidCerts = true
			continue
		case "-nick", "-user", "-pass", "-sasl", "-join", "-name":
		default:
			return srv, "", errors.New(i18n.T("Unknown option: %s", f))
		}
		if i+1 >= len(fields) {
			return srv, "", errors.New(i18n.T("Missing value for %s", f))
		}
		i++
		v := fields[i]
		switch f {
		case "-nick":
			srv.Nickname = v
		case "-user":
			srv.Username = v
		case "-pass":
			srv.Password = v
		case "-sasl":
			user, pass, ok := strings.Cut(v, ":")
			if !ok || user == "" {
				return srv, "", errors.New(i18n.T("Usage: -sasl user:password"))
			}
			srv.SASL = &config.SASL{Plain: &config.SASLPlain{Username: user, Password: pass}}
		case "-join":
			srv.Channels = strings.Split(v, ",")
		case "-name":
			name = v
		}
	}
	if srv.Nickname == "" {
		srv.Nickname = m.defaultNick()
	}
	if srv.Username == "" {
		srv.Username = srv.Nickname
	}
	if tlsOn == nil {
		on := srv.Port == 6697
		tlsOn = &on
	}
	srv.UseTLS = tlsOn
	return srv, name, nil
}

// defaultNick is the nickname of the first configured server, or "wisirc".
func (m *Manager) defaultNick() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	names := make([]string, 0, len(m.cfg.Servers))
	for n := range m.cfg.Servers {
		names = append(names, n)
	}
	sort.Strings(names)
	if len(names) > 0 && m.cfg.Servers[names[0]].Nickname != "" {
		return m.cfg.Servers[names[0]].Nickname
	}
	return "wisirc"
}
