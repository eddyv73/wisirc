package state

import (
	"testing"

	"github.com/eddyv73/wisirc/internal/i18n"
)

func TestEventLinesFollowLanguage(t *testing.T) {
	s := connected(t)
	s.Apply(msg(t, ":ana!a@h JOIN #lobby"))
	b := s.Buffer(BufRef{"n", "#lobby"})
	if got := b.Lines[len(b.Lines)-1].Text; got != "ana joined #lobby" {
		t.Fatalf("en: %q", got)
	}

	i18n.SetLang("es")
	t.Cleanup(func() { i18n.SetLang("en") })
	s.Apply(msg(t, ":bob!b@h JOIN #lobby"))
	s.Apply(msg(t, ":bob!b@h PART #lobby :adiós"))
	lines := b.Lines[len(b.Lines)-2:]
	if lines[0].Text != "bob se unió a #lobby" || lines[1].Text != "bob salió de #lobby (adiós)" {
		t.Fatalf("es: %q / %q", lines[0].Text, lines[1].Text)
	}
}
