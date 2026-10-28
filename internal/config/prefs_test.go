package config

import (
	"os"
	"testing"
)

func TestPrefsRoundTrip(t *testing.T) {
	p := writeTemp(t, userExample)
	if got := LoadPrefs(p); got.Theme != "" {
		t.Fatalf("empty prefs %+v", got)
	}
	if err := SavePrefs(p, Prefs{Theme: "bosque"}); err != nil {
		t.Fatal(err)
	}
	if got := LoadPrefs(p); got.Theme != "bosque" {
		t.Fatalf("got %+v", got)
	}
	st, _ := os.Stat(PrefsPath(p))
	if st.Mode().Perm() != 0o600 {
		t.Fatalf("mode %v", st.Mode())
	}
}

func TestPrefsLang(t *testing.T) {
	p := writeTemp(t, userExample)
	if err := SavePrefs(p, Prefs{Theme: "bosque", Lang: "es"}); err != nil {
		t.Fatal(err)
	}
	if got := LoadPrefs(p); got.Lang != "es" || got.Theme != "bosque" {
		t.Fatalf("%+v", got)
	}
}

func TestPrefsServerColors(t *testing.T) {
	p := writeTemp(t, userExample)
	if err := SavePrefs(p, Prefs{ServerColors: map[string]string{"znc": "#5fafd7"}}); err != nil {
		t.Fatal(err)
	}
	if got := LoadPrefs(p).ServerColors["znc"]; got != "#5fafd7" {
		t.Fatalf("got %q", got)
	}
}
