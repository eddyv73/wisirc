package ui

import (
	"context"
	"errors"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/eddyv73/wisirc/internal/i18n"
	"github.com/eddyv73/wisirc/internal/state"
	"github.com/eddyv73/wisirc/internal/update"
)

// Updater finds and installs wisirc releases (internal/update).
type Updater interface {
	Latest(ctx context.Context) (string, error)
	Apply(ctx context.Context, tag string) (string, error)
}

// updateEvery is how often a running wisirc looks for a new release.
const updateEvery = 24 * time.Hour

type updateState struct {
	tag      string // newer release available ("" = none known)
	applying bool
	brew     bool   // installed with Homebrew: brew upgrades it
	done     string // release installed, used after restarting
}

type updateTickMsg struct{}

type updateCheckedMsg struct {
	tag    string
	err    error
	manual bool // from /update: install it if newer
}

type updateAppliedMsg struct {
	tag, path string
	err       error
}

// checkUpdate asks GitHub for the latest release in the background.
func (m *Model) checkUpdate(manual bool) tea.Cmd {
	u := m.d.Updates
	if u == nil {
		return nil
	}
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()
		tag, err := u.Latest(ctx)
		return updateCheckedMsg{tag: tag, err: err, manual: manual}
	}
}

func (m *Model) updateChecked(msg updateCheckedMsg) tea.Cmd {
	var next tea.Cmd
	if !msg.manual {
		next = tea.Tick(updateEvery, func(time.Time) tea.Msg { return updateTickMsg{} })
	}
	if msg.err != nil {
		if msg.manual {
			m.errorLine(i18n.T("Could not check for updates: %s", msg.err.Error()))
		}
		return next
	}
	if update.Newer(m.d.Version, msg.tag) && msg.tag != m.update.done {
		m.update.tag = msg.tag
		if msg.manual {
			return m.applyUpdate()
		}
		return next
	}
	if msg.manual {
		m.info(i18n.T("wisirc %s is up to date (latest release: %s)", m.d.Version, msg.tag))
	}
	return next
}

// startUpdate is /update: install the newer release, checking first if
// none is known yet.
func (m *Model) startUpdate() tea.Cmd {
	switch {
	case m.d.Updates == nil:
		m.info(i18n.T("Update checks are off ([updates] check = false in config.toml)"))
		return nil
	case m.update.applying:
		return nil
	case m.update.done != "":
		m.info(i18n.T("wisirc %s is installed: reopen wisirc to use it", m.update.done))
		return nil
	case m.update.tag != "":
		return m.applyUpdate()
	}
	m.info(i18n.T("Checking for updates…"))
	return m.checkUpdate(true)
}

func (m *Model) applyUpdate() tea.Cmd {
	u, tag := m.d.Updates, m.update.tag
	m.update.applying = true
	m.info(i18n.T("Downloading wisirc %s…", tag))
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
		defer cancel()
		path, err := u.Apply(ctx, tag)
		return updateAppliedMsg{tag: tag, path: path, err: err}
	}
}

func (m *Model) updateApplied(msg updateAppliedMsg) {
	m.update.applying = false
	switch {
	case errors.Is(msg.err, update.ErrHomebrew):
		m.update.brew = true
		m.errorLine(i18n.T("wisirc was installed with Homebrew: update it with brew upgrade wisirc"))
	case msg.err != nil:
		m.errorLine(i18n.T("Could not update: %s", msg.err.Error()))
	default:
		m.update.tag, m.update.done = "", msg.tag
		m.info(i18n.T("wisirc %s installed in %s: reopen wisirc to use it", msg.tag, msg.path))
	}
}

// updateSeg is the status bar notice for a newer release.
func (m *Model) updateSeg(base lipgloss.Style) barSeg {
	var s barSeg
	switch {
	case m.update.applying:
		s.add(base.Foreground(m.theme.accent).Render("⬇ "+m.update.tag), "", state.BufRef{})
	case m.update.tag != "" && m.update.brew:
		s.add(base.Foreground(m.theme.ok).Bold(true).Render(i18n.T("⬆ %s available · brew upgrade wisirc", m.update.tag)), "", state.BufRef{})
	case m.update.tag != "":
		s.add(base.Foreground(m.theme.ok).Bold(true).Render(i18n.T("⬆ %s available · /update", m.update.tag)), "update", state.BufRef{})
	case m.update.done != "":
		s.add(base.Foreground(m.theme.ok).Render(i18n.T("✓ %s: reopen wisirc", m.update.done)), "", state.BufRef{})
	}
	return s
}

// globalMenu is the ≡ menu, with an Update entry when a release is waiting.
func (m *Model) globalMenu() Menu {
	g := GlobalMenu()
	if m.update.tag != "" && !m.update.applying && !m.update.brew {
		g.Items = append([]MenuItem{{i18n.T("Update to %s", m.update.tag), ActUpdate}}, g.Items...)
	}
	return g
}
