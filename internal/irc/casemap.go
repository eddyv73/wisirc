package irc

import "strings"

var rfc1459 = strings.NewReplacer("[", "{", "]", "}", `\`, "|", "~", "^")

// Fold lowercases s using RFC1459 casemapping.
func Fold(s string) string {
	return rfc1459.Replace(strings.ToLower(s))
}
