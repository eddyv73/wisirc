// Package irc implements the IRC wire protocol and a reconnecting client.
package irc

import (
	"errors"
	"strings"
	"time"
	"unicode/utf8"
)

type Message struct {
	Tags    map[string]string
	Prefix  string
	Command string
	Params  []string
}

var errEmpty = errors.New("irc: empty message")

var tagUnescaper = strings.NewReplacer(`\:`, ";", `\s`, " ", `\\`, `\`, `\r`, "\r", `\n`, "\n")

// Parse parses one IRC line (with or without the trailing CRLF).
func Parse(line string) (Message, error) {
	line = strings.TrimRight(line, "\r\n")
	var m Message
	if strings.HasPrefix(line, "@") {
		raw, rest, _ := strings.Cut(line[1:], " ")
		m.Tags = map[string]string{}
		for _, kv := range strings.Split(raw, ";") {
			if kv == "" {
				continue
			}
			k, v, _ := strings.Cut(kv, "=")
			m.Tags[k] = tagUnescaper.Replace(v)
		}
		line = strings.TrimLeft(rest, " ")
	}
	if strings.HasPrefix(line, ":") {
		p, rest, _ := strings.Cut(line[1:], " ")
		m.Prefix = p
		line = strings.TrimLeft(rest, " ")
	}
	for line != "" {
		if strings.HasPrefix(line, ":") {
			m.Params = append(m.Params, line[1:])
			break
		}
		var tok string
		tok, line, _ = strings.Cut(line, " ")
		line = strings.TrimLeft(line, " ")
		if m.Command == "" {
			m.Command = strings.ToUpper(tok)
		} else {
			m.Params = append(m.Params, tok)
		}
	}
	if m.Command == "" {
		return Message{}, errEmpty
	}
	return m, nil
}

// Nick returns the nickname part of the prefix.
func (m Message) Nick() string {
	n, _, _ := strings.Cut(m.Prefix, "!")
	return n
}

// Param returns the i-th parameter or "".
func (m Message) Param(i int) string {
	if i < 0 || i >= len(m.Params) {
		return ""
	}
	return m.Params[i]
}

// Trailing returns the last parameter or "".
func (m Message) Trailing() string { return m.Param(len(m.Params) - 1) }

// Time returns the IRCv3 server-time tag, if present and valid.
func (m Message) Time() (time.Time, bool) {
	v, ok := m.Tags["time"]
	if !ok {
		return time.Time{}, false
	}
	t, err := time.Parse(time.RFC3339Nano, v)
	if err != nil {
		return time.Time{}, false
	}
	return t, true
}

// Encode builds an IRC line without CRLF.
func Encode(cmd string, params ...string) string {
	var b strings.Builder
	b.WriteString(cmd)
	for i, p := range params {
		b.WriteByte(' ')
		if i == len(params)-1 && (p == "" || strings.Contains(p, " ") || strings.HasPrefix(p, ":")) {
			b.WriteByte(':')
		}
		b.WriteString(p)
	}
	return b.String()
}

// DecodeLine returns b as UTF-8, falling back to latin-1 for invalid input.
func DecodeLine(b []byte) string {
	if utf8.Valid(b) {
		return string(b)
	}
	r := make([]rune, len(b))
	for i, c := range b {
		r[i] = rune(c)
	}
	return string(r)
}
