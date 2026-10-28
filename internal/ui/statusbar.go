package ui

import (
	"fmt"
	"image/color"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/eddyv73/wisirc/internal/i18n"
	"github.com/eddyv73/wisirc/internal/state"
)

// barHit is a clickable span of the status bar, in columns.
type barHit struct {
	x0, x1 int
	act    string // "network", "buffer", "mention"
	ref    state.BufRef
}

// barSeg is a piece of the status bar with the hits inside it, relative to
// the segment start.
type barSeg struct {
	text string
	hits []barHit
}

func (s *barSeg) add(text string, act string, ref state.BufRef) {
	w := ansi.StringWidth(s.text)
	s.text += text
	if act != "" {
		s.hits = append(s.hits, barHit{x0: w, x1: w + ansi.StringWidth(text), act: act, ref: ref})
	}
}

func (s barSeg) width() int { return ansi.StringWidth(s.text) }

// renderStatusBar lays out, by priority: server dots and the active server
// (always), activity chips and the mention count (then), and on the right
// nick/channel modes and the clock (first to go when narrow).
func (m *Model) renderStatusBar(w int) string {
	th := m.theme
	base := lipgloss.NewStyle().Foreground(th.barFg).Background(th.barBg)
	paint := func(c color.Color, s string) string { return base.Foreground(c).Render(s) }
	sep := base.Render(" · ")

	// left: one dot per server, then the active server's state and lag
	var left barSeg
	left.add(base.Render(" "), "", state.BufRef{})
	first := true
	for _, n := range m.st.Networks {
		if n.Name == "" {
			continue
		}
		if !first {
			left.add(base.Render(" "), "", state.BufRef{})
		}
		first = false
		glyph, c := "○", th.dim
		switch {
		case n.Connected:
			glyph, c = "●", m.serverTint(n.Name)
		case n.Connecting:
			glyph, c = []string{"◌", "○"}[m.frame/2%2], m.serverTint(n.Name)
		}
		left.add(paint(c, glyph), "network", state.BufRef{Network: n.Name})
	}
	if n := m.st.Network(m.st.Active.Network); n != nil && n.Name != "" {
		left.add(base.Render("  "), "", state.BufRef{})
		left.add(base.Bold(true).Render(sanitize(n.Name)), "network", state.BufRef{Network: n.Name})
		if cfg := m.d.Nets.Config(); cfg != nil {
			if s, ok := cfg.Servers[n.Name]; ok {
				if s.TLS() {
					left.add(base.Render(" ")+paint(th.ok, "tls"), "", state.BufRef{})
				} else {
					left.add(base.Render(" ")+paint(th.mention, i18n.T("no tls")), "", state.BufRef{})
				}
			}
		}
		switch {
		case n.Connected:
			lag := paint(th.dim, pulseFrames[m.frame%len(pulseFrames)])
			if len(n.LagHistory) > 0 {
				c := barLagColor(th, n.Lag)
				lag = paint(c, Sparkline(n.LagHistory)) + base.Render(" ") +
					paint(c, pulseFrames[m.frame%len(pulseFrames)]+fmt.Sprintf(" %dms", n.Lag.Milliseconds()))
			}
			left.add(sep+lag, "", state.BufRef{})
		case n.Connecting:
			left.add(sep+paint(th.accent, spinFrames[m.frame%len(spinFrames)])+base.Render(" "+i18n.T("connecting…")), "", state.BufRef{})
		default:
			left.add(sep+base.Render(i18n.T("disconnected")), "", state.BufRef{})
		}
	}

	// middle: unread chips (mentions first) and the mention count
	var chips []barSeg
	mentions := 0
	var unread []*state.Buffer
	for _, r := range m.st.Order() {
		b := m.st.Buffer(r)
		if b == nil || refKey(r) == refKey(m.st.Active) {
			continue
		}
		if b.Mentioned {
			mentions++
		}
		if b.Unread > 0 || b.Mentioned {
			unread = append(unread, b)
		}
	}
	for pass := 0; pass < 2; pass++ {
		for _, b := range unread {
			if b.Mentioned != (pass == 0) {
				continue
			}
			name := b.Ref.Name
			if name == "" {
				name = b.Ref.Network
			}
			st := base.Foreground(m.channelTint(b.Ref.Network)).Bold(true)
			if b.Mentioned {
				st = base.Foreground(th.mention).Bold(true)
			}
			var c barSeg
			c.add(st.Render(fmt.Sprintf("%s %d", sanitize(name), b.Unread)), "buffer", b.Ref)
			chips = append(chips, c)
		}
	}
	var mention barSeg
	if mentions > 0 {
		mention.add(base.Foreground(th.mention).Bold(true).Render(fmt.Sprintf("!%d", mentions)), "mention", state.BufRef{})
	}
	var flash barSeg
	if m.flash != "" {
		flash.add(base.Render(sanitize(m.flash)), "", state.BufRef{})
	}

	// right: nick(+modes) · #chan(+modes) users · clock
	var right barSeg
	if n := m.st.Network(m.st.Active.Network); n != nil && n.Connected && n.Nick != "" {
		nick := sanitize(n.Nick)
		if n.UserModes != "" {
			nick += "(+" + n.UserModes + ")"
		}
		right.add(base.Render(nick), "", state.BufRef{})
		if b := m.activeBuffer(); b != nil && b.Kind == state.KindChannel {
			ch := sanitize(b.Ref.Name)
			if b.Modes != "" {
				ch += "(+" + b.Modes + ")"
			}
			right.add(sep+base.Render(fmt.Sprintf("%s %d", ch, len(b.Members))), "", state.BufRef{})
		}
		right.add(sep, "", state.BufRef{})
	}
	right.add(base.Render(time.Now().Format("15:04")+" "), "", state.BufRef{})

	// fit by priority: trim chips to three (+N), then drop the right side,
	// then trim the remaining chips
	divider := base.Render(" │ ")
	const dividerW = 3
	middle := func(n int) barSeg {
		var s barSeg
		parts := append(append([]barSeg{}, flash), chips[:n]...)
		if hidden := len(chips) - n; hidden > 0 {
			var more barSeg
			more.add(base.Foreground(th.dim).Render(fmt.Sprintf("+%d", hidden)), "mention", state.BufRef{})
			parts = append(parts, more)
		}
		if mention.text != "" {
			parts = append(parts, mention)
		}
		for _, p := range parts {
			if p.text == "" {
				continue
			}
			if s.text != "" {
				s.add(base.Render("  "), "", state.BufRef{})
			}
			off := s.width()
			s.text += p.text
			for _, h := range p.hits {
				h.x0, h.x1 = h.x0+off, h.x1+off
				s.hits = append(s.hits, h)
			}
		}
		return s
	}
	n := len(chips)
	mid := middle(n)
	showRight := true
	total := func() int {
		t := left.width()
		if mid.text != "" {
			t += dividerW + mid.width()
		}
		if showRight {
			t += 1 + right.width()
		}
		return t
	}
	for total() > w && n > 3 {
		n--
		mid = middle(n)
	}
	if total() > w {
		showRight = false
	}
	for total() > w && n > 0 {
		n--
		mid = middle(n)
	}

	m.barHits = append(m.barHits[:0], left.hits...)
	out := left.text
	if mid.text != "" {
		off := ansi.StringWidth(out) + dividerW
		out += divider + mid.text
		for _, h := range mid.hits {
			h.x0, h.x1 = h.x0+off, h.x1+off
			m.barHits = append(m.barHits, h)
		}
	}
	if showRight {
		if pad := w - ansi.StringWidth(out) - right.width(); pad > 0 {
			out += base.Render(strings.Repeat(" ", pad))
		}
		out += right.text
	}
	if ow := ansi.StringWidth(out); ow > w {
		out = ansi.Truncate(out, w, "")
	} else if ow < w {
		out += base.Render(strings.Repeat(" ", w-ow))
	}
	return out
}

// statusClick handles a click on the status bar.
func (m *Model) statusClick(x int) tea.Cmd {
	for _, h := range m.barHits {
		if x < h.x0 || x >= h.x1 {
			continue
		}
		switch h.act {
		case "network":
			m.dash = nil
			m.switchTo(h.ref)
		case "buffer":
			m.dash = nil
			m.switchTo(h.ref)
		case "mention":
			m.dash = nil
			m.nextActivity()
		}
		return nil
	}
	return nil
}
