package ui

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/eddyv73/wisirc/internal/irc"
)

// 4. events waiting in the channel are handled together
func TestWaitEventBatches(t *testing.T) {
	ch := make(chan irc.Event, 10)
	for i := 0; i < 5; i++ {
		ch <- irc.Event{Network: "n", Kind: irc.EvLag}
	}
	msg := waitEvent(ch)()
	batch, ok := msg.(eventsMsg)
	if !ok || len(batch) != 5 {
		t.Fatalf("got %T len %d", msg, len(batch))
	}
}

// 5. the animation stops while the terminal is not focused
func TestAnimationPausesWithoutFocus(t *testing.T) {
	h := newHarness(t)
	h.send(tea.BlurMsg{})
	if _, cmd := h.m.Update(animTickMsg{}); cmd != nil {
		t.Fatal("still ticking without focus")
	}
	if _, cmd := h.m.Update(tea.FocusMsg{}); cmd == nil {
		t.Fatal("focus did not restart the animation")
	}
	if _, cmd := h.m.Update(tea.FocusMsg{}); cmd != nil {
		t.Fatal("a second focus started a second ticker")
	}
}
