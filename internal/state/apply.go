package state

import (
	"sort"
	"time"

	"strings"

	"github.com/eddyv73/wisirc/internal/i18n"
	"github.com/eddyv73/wisirc/internal/irc"
)

type applier struct {
	s     *State
	n     *Network
	ev    irc.Event
	app   []Appended
	notif *Notification
}

func (a *applier) add(ref BufRef, l Line) {
	if l.Time.IsZero() {
		l.Time = a.ev.Time
	}
	a.s.AddLine(ref, l)
	a.app = append(a.app, Appended{Ref: a.s.Buffer(ref).Ref, Line: l})
}

func (a *applier) status() BufRef { return BufRef{Network: a.n.Name} }

// activeOrStatus is where replies like WHOIS go.
func (a *applier) activeOrStatus() BufRef {
	if a.s.Active.Network == a.n.Name {
		return a.s.Active
	}
	return a.status()
}

func isChannel(name string) bool {
	return name != "" && strings.ContainsRune("#&+!", rune(name[0]))
}

// Apply updates state from an event. It returns a notification for mentions
// and private messages, and every line it appended (for logging).
func (s *State) Apply(ev irc.Event) (*Notification, []Appended) {
	a := &applier{s: s, n: s.AddNetwork(ev.Network), ev: ev}
	switch ev.Kind {
	case irc.EvConnecting:
		a.n.Connecting = true
		a.add(a.status(), Line{Kind: LineInfo, Text: ev.Text})
	case irc.EvConnected:
		a.n.Connected, a.n.Connecting = true, false
		a.n.ConnectedAt = a.ev.Time
		if a.n.ConnectedAt.IsZero() {
			a.n.ConnectedAt = time.Now()
		}
		a.n.Nick = ev.Msg.Param(0)
		a.add(a.status(), Line{Kind: LineInfo, Text: ev.Text})
	case irc.EvDisconnected, irc.EvError:
		if ev.Kind == irc.EvDisconnected {
			a.n.Connected, a.n.Connecting = false, false
			a.n.ConnectedAt = time.Time{}
			for _, b := range a.n.Buffers {
				if b.Kind == KindChannel {
					b.Joined = false
					b.Members = map[string]*Member{}
				}
			}
		}
		a.add(a.status(), Line{Kind: LineError, Text: ev.Text})
	case irc.EvLag:
		a.n.Lag = ev.Lag
		a.n.LagHistory = append(a.n.LagHistory, ev.Lag)
		if over := len(a.n.LagHistory) - MaxLagSamples; over > 0 {
			a.n.LagHistory = append(a.n.LagHistory[:0], a.n.LagHistory[over:]...)
		}
	case irc.EvMessage:
		a.message(ev.Msg)
	}
	return a.notif, a.app
}

func (a *applier) self(nick string) bool {
	return nick != "" && irc.Fold(nick) == irc.Fold(a.n.Nick)
}

func (a *applier) message(m irc.Message) {
	from := m.Nick()
	fromServer := m.Prefix == "" || !strings.Contains(m.Prefix, "!")
	switch m.Command {
	case "PRIVMSG", "NOTICE":
		a.privmsg(m, from, fromServer)
	case "JOIN":
		ch := m.Param(0)
		if a.self(from) {
			b := a.s.EnsureBuffer(BufRef{a.n.Name, ch}, KindChannel)
			b.Joined = true
			b.Members = map[string]*Member{}
		}
		if b := a.s.Buffer(BufRef{a.n.Name, ch}); b != nil {
			b.Members[irc.Fold(from)] = &Member{Nick: from}
			a.add(b.Ref, Line{Kind: LineJoin, Nick: from, Text: i18n.T("%s joined %s", from, b.Ref.Name)})
		}
	case "PART":
		ch := m.Param(0)
		if b := a.s.Buffer(BufRef{a.n.Name, ch}); b != nil {
			delete(b.Members, irc.Fold(from))
			text := i18n.T("%s left %s", from, b.Ref.Name)
			if r := m.Param(1); r != "" {
				text += " (" + r + ")"
			}
			if a.self(from) {
				b.Joined = false
				b.Members = map[string]*Member{}
			}
			a.add(b.Ref, Line{Kind: LinePart, Nick: from, Text: text})
		}
	case "KICK":
		ch, victim := m.Param(0), m.Param(1)
		if b := a.s.Buffer(BufRef{a.n.Name, ch}); b != nil {
			delete(b.Members, irc.Fold(victim))
			text := i18n.T("%s kicked %s from %s", from, victim, b.Ref.Name)
			if r := m.Param(2); r != "" {
				text += " (" + r + ")"
			}
			if a.self(victim) {
				b.Joined = false
				b.Members = map[string]*Member{}
			}
			a.add(b.Ref, Line{Kind: LineKick, Nick: from, Text: text, Mention: a.self(victim)})
		}
	case "QUIT":
		text := i18n.T("%s quit", from)
		if r := m.Param(0); r != "" {
			text += " (" + r + ")"
		}
		key := irc.Fold(from)
		for _, b := range a.n.Buffers {
			_, inChan := b.Members[key]
			if inChan || (b.Kind == KindQuery && irc.Fold(b.Ref.Name) == key) {
				delete(b.Members, key)
				a.add(b.Ref, Line{Kind: LineQuit, Nick: from, Text: text})
			}
		}
	case "NICK":
		nw := m.Param(0)
		if a.self(from) {
			a.n.Nick = nw
			a.add(a.status(), Line{Kind: LineNick, Text: i18n.T("You are now %s", nw)})
		}
		oldKey := irc.Fold(from)
		for _, b := range a.n.Buffers {
			if mem, ok := b.Members[oldKey]; ok {
				delete(b.Members, oldKey)
				mem.Nick = nw
				b.Members[irc.Fold(nw)] = mem
				a.add(b.Ref, Line{Kind: LineNick, Nick: from, Text: i18n.T("%s is now %s", from, nw)})
			} else if b.Kind == KindQuery && irc.Fold(b.Ref.Name) == oldKey {
				wasActive := a.s.isActive(b.Ref)
				b.Ref.Name = nw
				if wasActive {
					a.s.Active = b.Ref
				}
				a.add(b.Ref, Line{Kind: LineNick, Nick: from, Text: i18n.T("%s is now %s", from, nw)})
			}
		}
	case "MODE":
		target := m.Param(0)
		text := i18n.T("%s set mode: %s", from, strings.Join(m.Params[1:], " "))
		if b := a.s.Buffer(BufRef{a.n.Name, target}); b != nil && isChannel(target) {
			applyModes(b, m.Params[1:])
			a.add(b.Ref, Line{Kind: LineMode, Nick: from, Text: text})
		} else if a.self(target) {
			a.n.UserModes = toggleModes(a.n.UserModes, m.Param(1))
			a.add(a.status(), Line{Kind: LineMode, Nick: from, Text: text})
		} else {
			a.add(a.status(), Line{Kind: LineMode, Nick: from, Text: text})
		}
	case "324": // RPL_CHANNELMODEIS: me #chan +modes [params]
		if b := a.s.Buffer(BufRef{a.n.Name, m.Param(1)}); b != nil {
			b.Modes = ""
			applyModes(b, m.Params[2:])
		}
	case "221": // RPL_UMODEIS: me +modes
		a.n.UserModes = toggleModes("", m.Param(1))
	case "TOPIC":
		if b := a.s.Buffer(BufRef{a.n.Name, m.Param(0)}); b != nil {
			b.Topic = m.Param(1)
			a.add(b.Ref, Line{Kind: LineTopic, Nick: from, Text: i18n.T("%s changed the topic: %s", from, b.Topic)})
		}
	case "332":
		if b := a.s.Buffer(BufRef{a.n.Name, m.Param(1)}); b != nil {
			b.Topic = m.Param(2)
			a.add(b.Ref, Line{Kind: LineTopic, Text: i18n.T("Topic: %s", b.Topic)})
		}
	case "333", "366":
	case "353":
		if b := a.s.Buffer(BufRef{a.n.Name, m.Param(2)}); b != nil {
			for _, tok := range strings.Fields(m.Param(3)) {
				i := 0
				for i < len(tok) && strings.IndexByte(prefixRank, tok[i]) >= 0 {
					i++
				}
				nick, _, _ := strings.Cut(tok[i:], "!")
				b.Members[irc.Fold(nick)] = &Member{Nick: nick, Modes: sortModes(tok[:i])}
			}
		}
	case "ERROR":
		a.add(a.status(), Line{Kind: LineError, Text: m.Trailing()})
	default:
		a.numeric(m)
	}
}

func (a *applier) privmsg(m irc.Message, from string, fromServer bool) {
	target, text := m.Param(0), m.Param(1)
	kind := LineMsg
	if m.Command == "NOTICE" {
		kind = LineNotice
	}
	if strings.HasPrefix(text, "\x01") {
		inner := strings.Trim(text, "\x01")
		rest, ok := strings.CutPrefix(inner, "ACTION")
		if !ok || m.Command != "PRIVMSG" {
			return // other CTCP is ignored
		}
		kind, text = LineAction, strings.TrimPrefix(rest, " ")
	}
	self := a.self(from)
	var ref BufRef
	isQuery := false
	switch {
	case fromServer || !a.n.Connected && kind == LineNotice:
		ref = a.status()
	case isChannel(target):
		ref = a.s.EnsureBuffer(BufRef{a.n.Name, target}, KindChannel).Ref
	case kind == LineNotice:
		ref = a.status()
		if b := a.s.Buffer(BufRef{a.n.Name, from}); b != nil {
			ref = b.Ref
		}
	default:
		name := from
		if self {
			name = target
		}
		ref = BufRef{a.n.Name, name}
		if a.s.Buffer(ref) == nil && a.queries() >= MaxQueries {
			ref = a.status()
		} else {
			ref = a.s.EnsureBuffer(ref, KindQuery).Ref
		}
		isQuery = true
	}
	mention := !self && !fromServer && (IsMention(a.n.Nick, text) || (isQuery && kind != LineNotice))
	a.add(ref, Line{Kind: kind, Nick: from, Text: text, Mention: mention, Self: self})
	if mention {
		a.notif = &Notification{Ref: ref, From: from, Text: text}
	}
}

func (a *applier) numeric(m irc.Message) {
	if len(m.Command) != 3 || m.Command[0] < '0' || m.Command[0] > '9' {
		return
	}
	text := strings.Join(m.Params[min(1, len(m.Params)):], " ")
	switch {
	case isWhois(m.Command):
		a.add(a.activeOrStatus(), Line{Kind: LineInfo, Text: text})
	case m.Command[0] == '4' || m.Command[0] == '5':
		a.add(a.activeOrStatus(), Line{Kind: LineError, Text: text})
	default:
		a.add(a.status(), Line{Kind: LineInfo, Text: text})
	}
}

func isWhois(code string) bool {
	switch code {
	case "311", "312", "313", "314", "317", "318", "319", "320", "330", "338", "369", "378", "671":
		return true
	}
	return false
}

var modePrefix = map[byte]byte{'q': '~', 'a': '&', 'o': '@', 'h': '%', 'v': '+'}

func sortModes(m string) string {
	var b strings.Builder
	for i := 0; i < len(prefixRank); i++ {
		if strings.IndexByte(m, prefixRank[i]) >= 0 {
			b.WriteByte(prefixRank[i])
		}
	}
	return b.String()
}

// applyModes updates member prefixes from "MODE #c +o-v nick nick" args.
func applyModes(b *Buffer, args []string) {
	if len(args) == 0 {
		return
	}
	params := args[1:]
	adding := true
	next := func() string {
		if len(params) == 0 {
			return ""
		}
		p := params[0]
		params = params[1:]
		return p
	}
	for i := 0; i < len(args[0]); i++ {
		c := args[0][i]
		switch c {
		case '+':
			adding = true
		case '-':
			adding = false
		case 'b', 'e', 'I':
			next()
		case 'k', 'l':
			if adding || c == 'k' {
				next()
			}
			b.Modes = toggleModes(b.Modes, string([]byte{sign(adding), c}))
		default:
			pfx, ok := modePrefix[c]
			if !ok {
				b.Modes = toggleModes(b.Modes, string([]byte{sign(adding), c}))
				continue
			}
			mem := b.Members[irc.Fold(next())]
			if mem == nil {
				continue
			}
			if adding {
				mem.Modes = sortModes(mem.Modes + string(pfx))
			} else {
				mem.Modes = strings.ReplaceAll(mem.Modes, string(pfx), "")
			}
		}
	}
}

func (a *applier) queries() int {
	n := 0
	for _, b := range a.n.Buffers {
		if b.Kind == KindQuery {
			n++
		}
	}
	return n
}

func sign(adding bool) byte {
	if adding {
		return '+'
	}
	return '-'
}

// toggleModes applies a "+ab-c" change to a sorted set of mode letters.
func toggleModes(have, change string) string {
	set := map[rune]bool{}
	for _, r := range have {
		set[r] = true
	}
	adding := true
	for _, r := range change {
		switch r {
		case '+':
			adding = true
		case '-':
			adding = false
		default:
			set[r] = adding
		}
	}
	out := make([]rune, 0, len(set))
	for r, on := range set {
		if on {
			out = append(out, r)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return string(out)
}
