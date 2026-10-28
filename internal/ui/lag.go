package ui

import (
	"image/color"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

const animEvery = 200 * time.Millisecond

type animTickMsg struct{}

func animTick() tea.Cmd {
	return tea.Tick(animEvery, func(time.Time) tea.Msg { return animTickMsg{} })
}

var (
	sparkBlocks = []rune("▁▂▃▄▅▆▇█")
	// lag at or above each step raises the bar one block
	sparkSteps               = []time.Duration{25, 50, 100, 200, 400, 800, 1600}
	pulseFrames              = []string{"·", "•", "●", "•"}
	spinFrames               = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	lagGood      color.Color = lipgloss.Color("10")
	lagSlow      color.Color = lipgloss.Color("11")
	lagBad       color.Color = lipgloss.Color("9")
	lagSlowAfter             = 150 * time.Millisecond
	lagBadAfter              = 500 * time.Millisecond
)

// Sparkline draws lag samples on a fixed scale, one block per sample.
func Sparkline(samples []time.Duration) string {
	var b strings.Builder
	for _, s := range samples {
		lvl := 0
		for _, step := range sparkSteps {
			if s >= step*time.Millisecond {
				lvl++
			}
		}
		b.WriteRune(sparkBlocks[lvl])
	}
	return b.String()
}

func lagColor(d time.Duration) color.Color {
	switch {
	case d >= lagBadAfter:
		return lagBad
	case d >= lagSlowAfter:
		return lagSlow
	}
	return lagGood
}

// barLagColor is the lag color for the status bar of theme t, falling back
// to the bar's own text color when it would not stand out from the bar.
func barLagColor(t theme, d time.Duration) color.Color {
	c := t.ok
	switch {
	case d >= lagBadAfter:
		c = t.errorC
	case d >= lagSlowAfter:
		c = lagSlow
	}
	if t.barBg != nil && contrastRatio(c, t.barBg) < 3 {
		return t.barFg
	}
	return c
}
