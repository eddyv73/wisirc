package ui

import (
	"strings"
	"testing"
	"time"

	"github.com/eddyv73/wisirc/internal/irc"
)

func TestSparkline(t *testing.T) {
	ms := time.Millisecond
	if got := Sparkline([]time.Duration{10 * ms, 30 * ms, 100 * ms, 2 * time.Second}); got != "▁▂▄█" {
		t.Fatalf("got %q", got)
	}
	if got := Sparkline(nil); got != "" {
		t.Fatalf("empty: %q", got)
	}
}

func (h *harness) statusRow() string {
	rows := strings.Split(h.view(), "\n")
	return rows[len(rows)-1]
}

func TestStatusShowsLagGraph(t *testing.T) {
	h := newHarness(t)
	h.irc(":srv 001 alice :hi")
	for _, d := range []time.Duration{10, 30, 42} {
		h.send(eventMsg(irc.Event{Network: netName, Kind: irc.EvLag, Lag: d * time.Millisecond}))
	}
	st := h.statusRow()
	if !strings.Contains(st, "▁▂▂") || !strings.Contains(st, "42ms") {
		t.Fatalf("status %q", st)
	}
	before := st
	_, cmd := h.m.Update(animTickMsg{})
	if cmd == nil {
		t.Fatal("animation stopped ticking")
	}
	if h.statusRow() == before {
		t.Fatal("tick did not animate the status bar")
	}
}

func TestStatusConnectingSpinner(t *testing.T) {
	h := newHarness(t)
	h.send(eventMsg(irc.Event{Network: netName, Kind: irc.EvConnecting, Text: "Connecting"}))
	before := h.statusRow()
	if !strings.Contains(before, "connecting") {
		t.Fatalf("status %q", before)
	}
	h.m.Update(animTickMsg{})
	if h.statusRow() == before {
		t.Fatal("spinner did not move")
	}
}
