package main

import (
	"runtime/debug"
	"strings"

	"context"
	"flag"
	"fmt"
	"github.com/charmbracelet/colorprofile"
	"os"
	"sort"

	tea "charm.land/bubbletea/v2"
	"github.com/eddyv73/wisirc/internal/app"
	"github.com/eddyv73/wisirc/internal/clipboard"
	"github.com/eddyv73/wisirc/internal/config"
	"github.com/eddyv73/wisirc/internal/halloy"
	"github.com/eddyv73/wisirc/internal/i18n"
	"github.com/eddyv73/wisirc/internal/logs"
	"github.com/eddyv73/wisirc/internal/notify"
	"github.com/eddyv73/wisirc/internal/ui"
	"github.com/eddyv73/wisirc/internal/update"
)

// Version is set at build time with -ldflags "-X main.Version=...".
var Version = "dev"

func main() {
	// the system locale until the saved choice is read (it needs -config)
	i18n.SetLang(i18n.Detect(os.Getenv))
	cfgPath := flag.String("config", config.DefaultPath(), i18n.T("path to the config file"))
	showVersion := flag.Bool("version", false, i18n.T("print the version"))
	flag.Parse()

	if *showVersion {
		fmt.Printf("wisirc %s\n", versionString())
		return
	}

	halloyPath, haveHalloy := halloy.Find()
	// language: the one chosen in the menu, else the system locale
	saved := config.LoadPrefs(*cfgPath)
	lang := saved.Lang
	if lang == "" {
		lang = i18n.Detect(os.Getenv)
	}
	i18n.SetLang(lang)

	// the first run (no config.toml yet) writes a config without servers and
	// opens the setup, which asks for the language, nick, theme and first server
	firstRun, err := config.EnsureStarter(*cfgPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "wisirc:", err)
		os.Exit(1)
	}

	cfg, warnings, err := config.Load(*cfgPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "wisirc:", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mgr := app.NewManager(ctx, cfg, nil)
	mgr.SetIdentity(saved.Nick, saved.Realname)
	disabled := map[string]bool{}
	for _, n := range config.LoadPrefs(*cfgPath).Disabled {
		disabled[n] = true
	}
	names := make([]string, 0, len(cfg.Servers))
	for n := range cfg.Servers {
		names = append(names, n)
	}
	sort.Strings(names)
	for _, n := range names {
		if cfg.Servers[n].Auto() && !disabled[n] {
			mgr.Start(n, cfg.Servers[n])
		}
	}

	// over SSH the system clipboard is the remote one; let the terminal copy (OSC 52)
	copyFn := clipboard.Copy
	if os.Getenv("SSH_TTY") != "" || os.Getenv("SSH_CONNECTION") != "" {
		copyFn = nil
	}
	prefs := prefsFile(*cfgPath)
	saveIdentity := func(nick, realname string) error {
		mgr.SetIdentity(nick, realname)
		return prefs.identity(nick, realname)
	}
	logger := &logs.Logger{Root: logs.DefaultRoot()}
	defer logger.Close()
	model := ui.New(ui.Deps{
		Cfg:     cfg,
		CfgPath: *cfgPath,
		Nets:    mgr,
		Events:  mgr.Events(),
		Logger:  logger,
		Notify:  notify.Send,
		Copy:    copyFn,

		Halloy:           halloyLoader(halloyPath, haveHalloy),
		Hints:            halloyHints(halloyPath, haveHalloy, cfg),
		ThemeName:        config.LoadPrefs(*cfgPath).Theme,
		StartOnDashboard: true,
		Setup:            firstRun,
		Nick:             saved.Nick,
		Realname:         saved.Realname,
		SaveIdentity:     saveIdentity,
		ServerColors:     config.LoadPrefs(*cfgPath).ServerColors,
		SaveServerColor:  prefs.color,
		Version:          versionString(),
		Updates:          updater(cfg),
		Disabled:         config.LoadPrefs(*cfgPath).Disabled,
		SaveDisabled:     prefs.disabled,
		SaveLang:         prefs.lang,
		SaveTheme:        prefs.theme,
		Warnings:         warnings,
	})
	var opts []tea.ProgramOption
	if p, ok := colorOverride(os.Getenv); ok {
		opts = append(opts, tea.WithColorProfile(p))
	}
	if _, err := tea.NewProgram(model, opts...).Run(); err != nil {
		fmt.Fprintln(os.Stderr, "wisirc:", err)
		os.Exit(1)
	}
}

// prefsFile saves the choices made in the UI to prefs.toml next to the
// config. Each save rereads the file so it only changes its own key.
type prefsFile string

func (f prefsFile) change(edit func(*config.Prefs)) error {
	p := config.LoadPrefs(string(f))
	edit(&p)
	return config.SavePrefs(string(f), p)
}

func (f prefsFile) theme(name string) error {
	return f.change(func(p *config.Prefs) { p.Theme = name })
}

func (f prefsFile) lang(code string) error {
	return f.change(func(p *config.Prefs) { p.Lang = code })
}

func (f prefsFile) identity(nick, realname string) error {
	return f.change(func(p *config.Prefs) { p.Nick, p.Realname = nick, realname })
}

func (f prefsFile) disabled(names []string) error {
	return f.change(func(p *config.Prefs) { p.Disabled = names })
}

func (f prefsFile) color(name, hex string) error {
	return f.change(func(p *config.Prefs) {
		if p.ServerColors == nil {
			p.ServerColors = map[string]string{}
		}
		if hex == "" {
			delete(p.ServerColors, name)
		} else {
			p.ServerColors[name] = hex
		}
	})
}

func halloyLoader(path string, ok bool) func() (map[string]config.Server, []string, error) {
	if !ok {
		return nil
	}
	return func() (map[string]config.Server, []string, error) { return halloy.Load(path) }
}

// halloyHints tells the user about Halloy servers that are not in wisirc yet.
func halloyHints(path string, ok bool, cfg *config.Config) []string {
	if !ok {
		return nil
	}
	servers, _, err := halloy.Load(path)
	if err != nil {
		return nil
	}
	n := 0
	for name := range servers {
		if _, have := cfg.Servers[name]; !have {
			n++
		}
	}
	if n == 0 {
		return nil
	}
	return []string{i18n.T("Found %d new server(s) in your Halloy config: ≡ → Servers → [Import from Halloy] to bring them in with one click.", n)}
}

// colorOverride honors an explicit COLORTERM=truecolor|24bit. The detector
// ignores COLORTERM inside tmux and asks the tmux server instead, which picks
// the wrong server for tmux -L sockets; tmux downsamples by itself if needed.
func colorOverride(getenv func(string) string) (colorprofile.Profile, bool) {
	switch strings.ToLower(getenv("COLORTERM")) {
	case "truecolor", "24bit":
		return colorprofile.TrueColor, true
	}
	return 0, false
}

// updater checks GitHub for new releases unless [updates] check = false.
func updater(cfg *config.Config) ui.Updater {
	if !cfg.Updates.On() {
		return nil
	}
	return update.New()
}

// versionString is the -ldflags version, or the module version recorded by
// "go install ...@vX.Y.Z", or "dev".
func versionString() string {
	if Version != "dev" {
		return Version
	}
	if bi, ok := debug.ReadBuildInfo(); ok && bi.Main.Version != "" && bi.Main.Version != "(devel)" {
		return bi.Main.Version
	}
	return Version
}
