package state

import (
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/eddyv73/wisirc/internal/irc"
)

func isNickRune(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r) || strings.ContainsRune("_-[]\\^{}|`", r)
}

// IsMention reports whether nick appears in text as a whole word.
func IsMention(nick, text string) bool {
	if nick == "" {
		return false
	}
	n, t := irc.Fold(nick), irc.Fold(text)
	for i := 0; ; {
		j := strings.Index(t[i:], n)
		if j < 0 {
			return false
		}
		start, end := i+j, i+j+len(n)
		before, _ := utf8.DecodeLastRuneInString(t[:start])
		after, _ := utf8.DecodeRuneInString(t[end:])
		if (start == 0 || !isNickRune(before)) && (end == len(t) || !isNickRune(after)) {
			return true
		}
		i = start + 1
	}
}
