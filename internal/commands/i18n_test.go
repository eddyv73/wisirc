package commands

import (
	"testing"

	"github.com/eddyv73/wisirc/internal/i18n"
)

func TestErrorsFollowLanguage(t *testing.T) {
	i18n.SetLang("es")
	t.Cleanup(func() { i18n.SetLang("en") })
	if a := Parse("/msg bob", chanCtx); a.Err != "Uso: /msg <nick|#canal> <texto>" {
		t.Errorf("usage: %q", a.Err)
	}
	if a := Parse("/frob", chanCtx); a.Err != "Comando desconocido: /frob" {
		t.Errorf("unknown: %q", a.Err)
	}
	// Spanish and English command names both keep working
	if a := Parse("/conectar", chanCtx); a.Kind != Connect {
		t.Errorf("/conectar: %+v", a)
	}
}
