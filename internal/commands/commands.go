// Package commands turns input-line text into IRC lines and UI actions.
package commands

import (
	"strings"
	"unicode/utf8"

	"github.com/eddyv73/wisirc/internal/i18n"
	"github.com/eddyv73/wisirc/internal/irc"
	"github.com/eddyv73/wisirc/internal/state"
)

const MaxMessageBytes = 400

type Context struct {
	Target string
	Kind   state.BufKind
	Nick   string
}

type Kind int

const (
	Send Kind = iota
	Close
	Query
	Connect
	Quit
	Help
	Theme
	Lang
	Dashboard
	Error
)

// Action is what the UI should do. Lines are raw IRC lines for the active
// network; Echo holds own text to show locally, one per PRIVMSG line.
type Action struct {
	Kind     Kind
	Lines    []string
	Echo     []string
	IsAction bool
	Target   string
	Arg      string
	Err      string
}

func errorf(msg string) Action { return Action{Kind: Error, Err: msg} }

func usage(u string) Action { return errorf(i18n.T("Usage: %s", u)) }

func isChannel(s string) bool {
	return s != "" && strings.ContainsRune("#&+!", rune(s[0]))
}

func privmsg(target, text string, action bool) Action {
	a := Action{Kind: Send, Target: target, IsAction: action}
	limit := MaxMessageBytes
	if action {
		limit -= len("\x01ACTION \x01")
	}
	for _, part := range SplitMessage(text, limit) {
		body := part
		if action {
			body = "\x01ACTION " + part + "\x01"
		}
		a.Lines = append(a.Lines, "PRIVMSG "+target+" :"+body)
		a.Echo = append(a.Echo, part)
	}
	return a
}

func send(lines ...string) Action { return Action{Kind: Send, Lines: lines} }

// Parse interprets one line typed in the input box.
func Parse(input string, ctx Context) Action {
	inBuffer := ctx.Kind != state.KindStatus && ctx.Target != ""
	if !strings.HasPrefix(input, "/") || strings.HasPrefix(input, "//") {
		if strings.HasPrefix(input, "//") {
			input = input[1:]
		}
		if input == "" {
			return Action{Kind: Send}
		}
		if !inBuffer {
			return errorf(i18n.T("You are not in a channel"))
		}
		return privmsg(ctx.Target, input, false)
	}

	name, rest, _ := strings.Cut(input[1:], " ")
	name = strings.ToLower(name)
	rest = strings.TrimSpace(rest)
	args := strings.Fields(rest)
	// tail returns the text after the first n words.
	tail := func(n int) string {
		s := rest
		for i := 0; i < n; i++ {
			_, s, _ = strings.Cut(strings.TrimLeft(s, " "), " ")
		}
		return strings.TrimSpace(s)
	}
	needChan := func() (string, bool) { return ctx.Target, inBuffer && ctx.Kind == state.KindChannel }

	switch name {
	case "me":
		if rest == "" {
			return usage(i18n.T("/me <action>"))
		}
		if !inBuffer {
			return errorf(i18n.T("You are not in a channel"))
		}
		return privmsg(ctx.Target, rest, true)
	case "join", "j":
		if len(args) == 0 {
			return usage(i18n.T("/join #channel[,#other] [key]"))
		}
		chans := strings.Split(args[0], ",")
		for i, c := range chans {
			if !isChannel(c) {
				chans[i] = "#" + c
			}
		}
		params := []string{strings.Join(chans, ",")}
		if len(args) > 1 {
			params = append(params, args[1])
		}
		return send(irc.Encode("JOIN", params...))
	case "part", "leave":
		ch, ok := needChan()
		msg := rest
		if len(args) > 0 && isChannel(args[0]) {
			ch, ok, msg = args[0], true, tail(1)
		}
		if !ok {
			return usage(i18n.T("/part [#channel] [message]"))
		}
		if msg != "" {
			return send("PART " + ch + " :" + msg)
		}
		return send(irc.Encode("PART", ch))
	case "msg":
		if len(args) < 2 {
			return usage(i18n.T("/msg <nick|#channel> <text>"))
		}
		return privmsg(args[0], tail(1), false)
	case "query", "q":
		if len(args) == 0 {
			return usage(i18n.T("/query <nick> [text]"))
		}
		a := Action{Kind: Query, Target: args[0]}
		if text := tail(1); text != "" {
			p := privmsg(args[0], text, false)
			a.Lines, a.Echo = p.Lines, p.Echo
		}
		return a
	case "nick":
		if len(args) != 1 {
			return usage(i18n.T("/nick <new>"))
		}
		return send(irc.Encode("NICK", args[0]))
	case "whois", "wi":
		if len(args) == 0 {
			return usage("/whois <nick>")
		}
		return send(irc.Encode("WHOIS", args[0]))
	case "topic":
		ch, ok := needChan()
		text := rest
		if len(args) > 0 && isChannel(args[0]) {
			ch, ok, text = args[0], true, tail(1)
		}
		if !ok {
			return usage(i18n.T("/topic [#channel] [text]"))
		}
		if text == "" {
			return send(irc.Encode("TOPIC", ch))
		}
		return send("TOPIC " + ch + " :" + text)
	case "kick", "k":
		ch, ok := needChan()
		if !ok || len(args) == 0 {
			return usage(i18n.T("/kick <nick> [reason]"))
		}
		if reason := tail(1); reason != "" {
			return send("KICK " + ch + " " + args[0] + " :" + reason)
		}
		return send(irc.Encode("KICK", ch, args[0]))
	case "ban":
		ch, ok := needChan()
		if !ok || len(args) == 0 {
			return usage("/ban <nick>")
		}
		mask := args[0]
		if !strings.ContainsAny(mask, "!@") {
			mask += "!*@*"
		}
		return send(irc.Encode("MODE", ch, "+b", mask))
	case "mode":
		if len(args) == 0 {
			return usage(i18n.T("/mode [target] <modes> [args]"))
		}
		if !strings.HasPrefix(args[0], "+") && !strings.HasPrefix(args[0], "-") {
			return send("MODE " + rest)
		}
		target := ctx.Target
		if ctx.Kind != state.KindChannel {
			target = ctx.Nick
		}
		return send("MODE " + target + " " + rest)
	case "quit", "exit":
		return Action{Kind: Quit, Arg: rest}
	case "connect", "conectar", "server":
		// without arguments the UI opens the connect screen with examples
		return Action{Kind: Connect, Arg: rest}
	case "close", "wc":
		return Action{Kind: Close}
	case "quote", "raw":
		if rest == "" {
			return usage(i18n.T("/quote <IRC line>"))
		}
		return send(rest)
	case "help":
		return Action{Kind: Help}
	case "panel", "dashboard":
		return Action{Kind: Dashboard}
	case "lang", "idioma", "language":
		return Action{Kind: Lang, Arg: rest}
	case "tema", "theme":
		return Action{Kind: Theme, Arg: rest}
	}
	return errorf(i18n.T("Unknown command: /%s", name))
}

// SplitMessage splits text into parts of at most max bytes, cutting on rune
// boundaries and preferring the last space.
func SplitMessage(text string, max int) []string {
	var out []string
	for len(text) > max {
		cut := max
		for cut > 0 && !utf8.RuneStart(text[cut]) {
			cut--
		}
		if sp := strings.LastIndexByte(text[:cut+1], ' '); sp > 0 {
			out = append(out, text[:sp])
			text = text[sp+1:]
			continue
		}
		out = append(out, text[:cut])
		text = text[cut:]
	}
	if text != "" {
		out = append(out, text)
	}
	return out
}
