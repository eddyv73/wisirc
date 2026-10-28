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

	// with a Halloy config around, the first run starts right away and offers
	// to import its servers instead of asking to edit an example
	halloyPath, haveHalloy := halloy.Find()
	ensure := config.EnsureExample
	if haveHalloy {
		ensure = config.EnsureStarter
	}
	// language: the one chosen in the menu, else the system locale
	lang := config.LoadPrefs(*cfgPath).Lang
	if lang == "" {
		lang = i18n.Detect(os.Getenv)
	}
	i18n.SetLang(lang)

	created, err := ensure(*cfgPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "wisirc:", err)
		os.Exit(1)
	}
	if created && !haveHalloy {
		fmt.Println(i18n.T("Created an example config at %s. Edit it and run wisirc again.", *cfgPath))
		return
	}

	cfg, warnings, err := config.Load(*cfgPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "wisirc:", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mgr := app.NewManager(ctx, cfg, nil)
	names := make([]string, 0, len(cfg.Servers))
	for n := range cfg.Servers {
		names = append(names, n)
	}
	sort.Strings(names)
	for _, n := range names {
		if cfg.Servers[n].Auto() {
			mgr.Start(n, cfg.Servers[n])
		}
	}

	// over SSH the system clipboard is the remote one; let the terminal copy (OSC 52)
	copyFn := clipboard.Copy
	if os.Getenv("SSH_TTY") != "" || os.Getenv("SSH_CONNECTION") != "" {
		copyFn = nil
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
		ServerColors:     config.LoadPrefs(*cfgPath).ServerColors,
		SaveServerColor: func(name, hex string) error {
			p := config.LoadPrefs(*cfgPath)
			if p.ServerColors == nil {
				p.ServerColors = map[string]string{}
			}
			if hex == "" {
				delete(p.ServerColors, name)
			} else {
				p.ServerColors[name] = hex
			}
			return config.SavePrefs(*cfgPath, p)
		},
		SaveLang: func(code string) error {
			p := config.LoadPrefs(*cfgPath)
			p.Lang = code
			return config.SavePrefs(*cfgPath, p)
		},
		SaveTheme: func(name string) error {
			p := config.LoadPrefs(*cfgPath)
			p.Theme = name
			return config.SavePrefs(*cfgPath, p)
		},
		Warnings: warnings,
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
