// Package state holds networks, buffers and nicklists, updated from IRC events.
package state

import (
	"sort"
	"strings"
	"time"

	"github.com/eddyv73/wisirc/internal/irc"
)

const MaxLines = 5000

type BufKind int

const (
	KindStatus BufKind = iota
	KindChannel
	KindQuery
)

type LineKind int

const (
	LineMsg LineKind = iota
	LineAction
	LineNotice
	LineJoin
	LinePart
	LineQuit
	LineNick
	LineKick
	LineMode
	LineTopic
	LineInfo
	LineError
	LineHistory
)

type Line struct {
	Time    time.Time
	Nick    string
	Text    string
	Kind    LineKind
	Mention bool
	Self    bool
}

// Member is a channel user; Modes holds prefix chars, highest first.
type Member struct {
	Nick  string
	Modes string
}

// BufRef identifies a buffer. Name "" is the network's status buffer.
type BufRef struct {
	Network string
	Name    string
}

type Buffer struct {
	Ref       BufRef
	Kind      BufKind
	Topic     string
	Lines     []Line
	Members   map[string]*Member // key: irc.Fold(nick)
	Unread    int
	Mentioned bool
	Joined    bool
	Modes     string // channel flag modes, sorted, e.g. "nt" (no list or member modes)
}

const prefixRank = "~&@%+"

func rank(modes string) int {
	if modes == "" {
		return len(prefixRank)
	}
	return strings.IndexByte(prefixRank, modes[0])
}

// SortedMembers orders by highest prefix, then case-folded nick.
func (b *Buffer) SortedMembers() []Member {
	out := make([]Member, 0, len(b.Members))
	for _, m := range b.Members {
		out = append(out, *m)
	}
	sort.Slice(out, func(i, j int) bool {
		ri, rj := rank(out[i].Modes), rank(out[j].Modes)
		if ri != rj {
			return ri < rj
		}
		return irc.Fold(out[i].Nick) < irc.Fold(out[j].Nick)
	})
	return out
}

// MaxQueries caps private-message windows per network; messages from new
// nicks beyond it go to the status buffer, so a spam flood cannot open
// thousands of windows.
const MaxQueries = 100

// MaxLagSamples is how many recent lag measurements a network keeps.
const MaxLagSamples = 12

type Network struct {
	Name        string
	Nick        string
	Connected   bool
	Connecting  bool
	Lag         time.Duration
	LagHistory  []time.Duration // oldest first, at most MaxLagSamples
	ConnectedAt time.Time       // zero while not connected
	UserModes   string          // our own user modes, sorted, e.g. "i"
	Buffers     []*Buffer       // [0] is the status buffer
	Collapsed   bool
}

type Notification struct {
	Ref  BufRef
	From string
	Text string
}

type Appended struct {
	Ref  BufRef
	Line Line
}

type State struct {
	Networks []*Network
	Active   BufRef
}

func New(networks []string) *State {
	s := &State{}
	for _, n := range networks {
		s.AddNetwork(n)
	}
	if len(s.Networks) > 0 {
		s.Active = BufRef{Network: s.Networks[0].Name}
	}
	return s
}

// AddNetwork returns the named network, creating it with a status buffer.
func (s *State) AddNetwork(name string) *Network {
	if n := s.Network(name); n != nil {
		return n
	}
	n := &Network{Name: name}
	n.Buffers = []*Buffer{{Ref: BufRef{Network: name}, Kind: KindStatus, Members: map[string]*Member{}}}
	s.Networks = append(s.Networks, n)
	if s.Active.Network == "" {
		s.Active = BufRef{Network: name}
	}
	return n
}

func (s *State) Network(name string) *Network {
	for _, n := range s.Networks {
		if n.Name == name {
			return n
		}
	}
	return nil
}

// Buffer finds a buffer, matching the name case-insensitively.
func (s *State) Buffer(ref BufRef) *Buffer {
	n := s.Network(ref.Network)
	if n == nil {
		return nil
	}
	key := irc.Fold(ref.Name)
	for _, b := range n.Buffers {
		if irc.Fold(b.Ref.Name) == key {
			return b
		}
	}
	return nil
}

func (s *State) EnsureBuffer(ref BufRef, kind BufKind) *Buffer {
	if b := s.Buffer(ref); b != nil {
		return b
	}
	n := s.AddNetwork(ref.Network)
	if ref.Name == "" {
		return n.Buffers[0]
	}
	b := &Buffer{Ref: ref, Kind: kind, Members: map[string]*Member{}}
	n.Buffers = append(n.Buffers, b)
	return b
}

// RemoveBuffer deletes a channel/query buffer; status buffers stay.
func (s *State) RemoveBuffer(ref BufRef) {
	if ref.Name == "" {
		return
	}
	n := s.Network(ref.Network)
	if n == nil {
		return
	}
	key := irc.Fold(ref.Name)
	for i, b := range n.Buffers {
		if irc.Fold(b.Ref.Name) == key {
			n.Buffers = append(n.Buffers[:i], n.Buffers[i+1:]...)
			break
		}
	}
	if s.Active.Network == ref.Network && irc.Fold(s.Active.Name) == key {
		s.Active = BufRef{Network: ref.Network}
	}
}

func (s *State) SetActive(ref BufRef) {
	b := s.Buffer(ref)
	if b == nil {
		return
	}
	s.Active = b.Ref
	b.Unread = 0
	b.Mentioned = false
}

func (s *State) isActive(ref BufRef) bool {
	return s.Active.Network == ref.Network && irc.Fold(s.Active.Name) == irc.Fold(ref.Name)
}

// AddLine appends a line, trims scrollback and updates activity markers.
func (s *State) AddLine(ref BufRef, l Line) {
	b := s.EnsureBuffer(ref, KindChannel)
	if l.Time.IsZero() {
		l.Time = time.Now()
	}
	b.Lines = append(b.Lines, l)
	if over := len(b.Lines) - MaxLines; over > 0 {
		// reslicing drops the oldest lines without copying; append moves the
		// buffer to a fresh array only when its capacity runs out
		b.Lines = b.Lines[over:]
	}
	if s.isActive(b.Ref) || l.Self {
		return
	}
	switch l.Kind {
	case LineMsg, LineAction, LineNotice:
		b.Unread++
		if l.Mention {
			b.Mentioned = true
		}
	}
}

// Order is the flattened buffer order used for keyboard navigation.
func (s *State) Order() []BufRef {
	var out []BufRef
	for _, n := range s.Networks {
		for _, b := range n.Buffers {
			out = append(out, b.Ref)
		}
	}
	return out
}
