// Package i18n translates user-visible text. Messages are written in English
// in the code and looked up in a per-language catalog; a missing
// translation falls back to the English text.
package i18n

import (
	"fmt"
	"strings"
	"sync/atomic"
)

// Language is a supported UI language.
type Language struct {
	Code string // "en", "es"
	Name string // native name, shown in the language menu
}

var languages = []Language{{"en", "English"}, {"es", "Español"}}

// catalogs maps language code → English message → translation.
var catalogs = map[string]map[string]string{"es": {}}

var current atomic.Value

func init() { current.Store("en") }

// add merges translations into a catalog (one call per catalog file).
func add(lang string, m map[string]string) {
	for k, v := range m {
		catalogs[lang][k] = v
	}
}

// Languages lists the supported languages, English first.
func Languages() []Language { return append([]Language(nil), languages...) }

// SetLang switches the language; unknown codes are ignored.
func SetLang(code string) bool {
	for _, l := range languages {
		if l.Code == code {
			current.Store(code)
			return true
		}
	}
	return false
}

// Lang returns the active language code.
func Lang() string { return current.Load().(string) }

// T translates msg and formats it with args like fmt.Sprintf.
func T(msg string, args ...any) string {
	if tr, ok := catalogs[Lang()][msg]; ok && tr != "" {
		msg = tr
	}
	if len(args) > 0 {
		return fmt.Sprintf(msg, args...)
	}
	return msg
}

// Detect picks a supported language from the POSIX locale variables.
func Detect(getenv func(string) string) string {
	for _, k := range []string{"LC_ALL", "LC_MESSAGES", "LANG"} {
		v := getenv(k)
		if v == "" {
			continue
		}
		code := strings.ToLower(v)
		if i := strings.IndexAny(code, "_.-@"); i >= 0 {
			code = code[:i]
		}
		for _, l := range languages {
			if l.Code == code {
				return code
			}
		}
		return "en"
	}
	return "en"
}
