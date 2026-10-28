package irc

import (
	"bufio"
	"bytes"
	"context"
	"crypto/tls"
	"encoding/base64"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/eddyv73/wisirc/internal/i18n"
)

// ErrNotConnected is returned by Send while there is no connection. Its text
// is translated when it is shown, so it follows the current language.
var ErrNotConnected error = errNotConnected{}

// MaxLineBytes caps one line from the server: 8191 bytes of IRCv3 tags plus
// the 512-byte message. Anything longer is cut so a hostile server cannot
// make wisirc hold megabytes per line.
const MaxLineBytes = 8191 + 512

type errNotConnected struct{}

func (errNotConnected) Error() string { return i18n.T("not connected") }

type Options struct {
	Name, Addr       string
	TLS, InsecureTLS bool
	Nick, User, Real string
	Pass             string
	SASLUser         string
	SASLPass         string
	Channels         []string
	Dial             func(ctx context.Context) (net.Conn, error)
	Backoff          func(attempt int) time.Duration
	PingEvery        time.Duration
}

var wantedCaps = []string{"sasl", "server-time", "multi-prefix"}

const pingPrefix = "wisirc-"

type Client struct {
	opts   Options
	events chan<- Event

	mu     sync.Mutex
	conn   net.Conn
	nick   string
	joined []string // channels we are in, kept across reconnects
	fatal  bool

	// per-session state, only touched by the reader goroutine
	registered bool
	lsCaps     []string
}

func New(opts Options, events chan<- Event) *Client {
	if opts.Backoff == nil {
		opts.Backoff = func(attempt int) time.Duration {
			if attempt > 6 {
				return 60 * time.Second
			}
			return min(time.Duration(1<<attempt)*time.Second, 60*time.Second)
		}
	}
	if opts.PingEvery == 0 {
		opts.PingEvery = 30 * time.Second
	}
	if opts.Dial == nil {
		opts.Dial = defaultDialer(opts)
	}
	return &Client{opts: opts, events: events, nick: opts.Nick}
}

func defaultDialer(opts Options) func(ctx context.Context) (net.Conn, error) {
	return func(ctx context.Context) (net.Conn, error) {
		d := &net.Dialer{Timeout: 20 * time.Second, KeepAlive: 60 * time.Second}
		if !opts.TLS {
			return d.DialContext(ctx, "tcp", opts.Addr)
		}
		host, _, _ := net.SplitHostPort(opts.Addr)
		td := &tls.Dialer{NetDialer: d, Config: &tls.Config{
			ServerName:         host,
			InsecureSkipVerify: opts.InsecureTLS,
		}}
		return td.DialContext(ctx, "tcp", opts.Addr)
	}
}

// CurrentNick returns our nickname on this network.
func (c *Client) CurrentNick() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.nick
}

// Send writes one raw line to the server.
func (c *Client) Send(line string) error {
	c.mu.Lock()
	conn := c.conn
	c.mu.Unlock()
	if conn == nil {
		return ErrNotConnected
	}
	conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
	_, err := conn.Write([]byte(line + "\r\n"))
	return err
}

func (c *Client) emit(ctx context.Context, ev Event) {
	ev.Network = c.opts.Name
	if ev.Time.IsZero() {
		ev.Time = time.Now()
	}
	select {
	case c.events <- ev:
	case <-ctx.Done():
	}
}

// emitFinal reports the last event after ctx is already cancelled.
func (c *Client) emitFinal(ev Event) {
	ev.Network = c.opts.Name
	ev.Time = time.Now()
	select {
	case c.events <- ev:
	case <-time.After(time.Second):
	}
}

// Run connects and keeps reconnecting until ctx is done or SASL fails.
// Every exit reports EvDisconnected so the UI never shows a stale connection.
func (c *Client) Run(ctx context.Context) {
	defer c.emitFinal(Event{Kind: EvDisconnected, Text: i18n.T("Disconnected")})
	attempt := 0
	for ctx.Err() == nil {
		c.emit(ctx, Event{Kind: EvConnecting, Text: i18n.T("Connecting to %s…", c.opts.Addr)})
		conn, err := c.opts.Dial(ctx)
		if err == nil {
			registered := c.session(ctx, conn)
			c.mu.Lock()
			fatal := c.fatal
			c.mu.Unlock()
			if fatal {
				return
			}
			if registered {
				attempt = 0
			}
		} else {
			c.emit(ctx, Event{Kind: EvDisconnected, Text: i18n.T("Could not connect: %s", err.Error())})
		}
		select {
		case <-time.After(c.opts.Backoff(attempt)):
		case <-ctx.Done():
			return
		}
		attempt++
	}
}

func (c *Client) session(ctx context.Context, conn net.Conn) bool {
	c.mu.Lock()
	c.conn = conn
	c.nick = c.opts.Nick
	c.mu.Unlock()
	c.registered = false
	c.lsCaps = nil

	done := make(chan struct{})
	defer func() {
		close(done)
		conn.Close()
		c.mu.Lock()
		c.conn = nil
		c.mu.Unlock()
	}()
	go func() {
		ticker := time.NewTicker(c.opts.PingEvery)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				conn.Close()
				return
			case <-done:
				return
			case <-ticker.C:
				c.ping()
			}
		}
	}()

	c.Send("CAP LS 302")
	if c.opts.Pass != "" {
		c.Send(Encode("PASS", c.opts.Pass))
	}
	c.Send(Encode("NICK", c.opts.Nick))
	c.Send("USER " + c.opts.User + " 0 * :" + c.opts.Real)

	r := bufio.NewReaderSize(conn, 16*1024)
	for {
		// any line from the server (including PONG) pushes the deadline forward
		conn.SetReadDeadline(time.Now().Add(2 * c.opts.PingEvery))
		raw, err := readLine(r, MaxLineBytes)
		if len(raw) > 0 {
			if msg, perr := Parse(DecodeLine(bytes.TrimRight(raw, "\r\n"))); perr == nil {
				c.handle(ctx, msg)
			}
		}
		if err != nil {
			c.mu.Lock()
			fatal := c.fatal
			c.mu.Unlock()
			if !fatal && ctx.Err() == nil {
				text := i18n.T("Disconnected: %s", err.Error())
				var ne net.Error
				if errors.As(err, &ne) && ne.Timeout() {
					text = i18n.T("Disconnected: ping timeout")
				}
				c.emit(ctx, Event{Kind: EvDisconnected, Text: text})
			}
			return c.registered
		}
	}
}

func (c *Client) ping() {
	c.Send("PING " + pingPrefix + strconv.FormatInt(time.Now().UnixNano(), 10))
}

func (c *Client) isSelf(nick string) bool {
	return Fold(nick) == Fold(c.CurrentNick())
}

func (c *Client) handle(ctx context.Context, m Message) {
	ev := Event{Kind: EvMessage, Msg: m}
	if t, ok := m.Time(); ok {
		ev.Time = t
	}
	switch m.Command {
	case "PING":
		c.Send(Encode("PONG", m.Params...))
	case "PONG":
		if tok := m.Trailing(); strings.HasPrefix(tok, pingPrefix) {
			if ns, err := strconv.ParseInt(tok[len(pingPrefix):], 10, 64); err == nil {
				c.emit(ctx, Event{Kind: EvLag, Lag: time.Since(time.Unix(0, ns))})
			}
			return
		}
	case "CAP":
		c.handleCap(m)
	case "AUTHENTICATE":
		if m.Param(0) == "+" {
			payload := c.opts.SASLUser + "\x00" + c.opts.SASLUser + "\x00" + c.opts.SASLPass
			c.Send("AUTHENTICATE " + base64.StdEncoding.EncodeToString([]byte(payload)))
		}
	case "903", "906", "907":
		c.Send("CAP END")
	case "904", "905":
		c.mu.Lock()
		c.fatal = true
		c.mu.Unlock()
		c.emit(ctx, Event{Kind: EvError, Text: i18n.T("SASL failed: %s", m.Trailing())})
		c.Send("QUIT")
		c.mu.Lock()
		if c.conn != nil {
			c.conn.Close()
		}
		c.mu.Unlock()
		return
	case "433":
		if !c.registered {
			c.mu.Lock()
			c.nick += "_"
			n := c.nick
			c.mu.Unlock()
			c.Send(Encode("NICK", n))
		}
	case "001":
		c.registered = true
		c.mu.Lock()
		c.nick = m.Param(0)
		chans := mergeChannels(c.opts.Channels, c.joined)
		c.joined = nil
		c.mu.Unlock()
		c.emit(ctx, Event{Kind: EvConnected, Msg: m, Text: i18n.T("Connected to %s", c.opts.Addr)})
		if len(chans) > 0 {
			c.Send(Encode("JOIN", strings.Join(chans, ",")))
		}
		c.ping()
	case "JOIN":
		if c.isSelf(m.Nick()) {
			c.mu.Lock()
			c.joined = mergeChannels(c.joined, []string{m.Param(0)})
			c.mu.Unlock()
		}
	case "PART":
		if c.isSelf(m.Nick()) {
			c.removeJoined(m.Param(0))
		}
	case "KICK":
		if c.isSelf(m.Param(1)) {
			c.removeJoined(m.Param(0))
		}
	case "NICK":
		if c.isSelf(m.Nick()) {
			c.mu.Lock()
			c.nick = m.Param(0)
			c.mu.Unlock()
		}
	}
	c.emit(ctx, ev)
}

func (c *Client) handleCap(m Message) {
	sub := strings.ToUpper(m.Param(1))
	switch sub {
	case "LS":
		more := m.Param(2) == "*" && len(m.Params) > 3
		for _, tok := range strings.Fields(m.Trailing()) {
			name, _, _ := strings.Cut(tok, "=")
			c.lsCaps = append(c.lsCaps, name)
		}
		if more {
			return
		}
		var req []string
		for _, offered := range c.lsCaps {
			for _, w := range wantedCaps {
				if offered == w && (w != "sasl" || c.opts.SASLUser != "") {
					req = append(req, w)
				}
			}
		}
		if len(req) == 0 {
			c.Send("CAP END")
			return
		}
		c.Send("CAP REQ :" + strings.Join(req, " "))
	case "ACK":
		acked := strings.Fields(m.Trailing())
		for _, a := range acked {
			if a == "sasl" && c.opts.SASLUser != "" {
				c.Send("AUTHENTICATE PLAIN")
				return
			}
		}
		c.Send("CAP END")
	case "NAK":
		c.Send("CAP END")
	}
}

func (c *Client) removeJoined(ch string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	out := c.joined[:0]
	for _, j := range c.joined {
		if Fold(j) != Fold(ch) {
			out = append(out, j)
		}
	}
	c.joined = out
}

func mergeChannels(a, b []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, list := range [][]string{a, b} {
		for _, ch := range list {
			if ch == "" || seen[Fold(ch)] {
				continue
			}
			seen[Fold(ch)] = true
			out = append(out, ch)
		}
	}
	return out
}

func (o Options) String() string { return fmt.Sprintf("%s(%s)", o.Name, o.Addr) }

// readLine reads up to the next newline but keeps at most max bytes of it,
// discarding the rest of an overlong line (cut on a UTF-8 boundary).
func readLine(r *bufio.Reader, max int) ([]byte, error) {
	var line []byte
	cut := false
	for {
		chunk, err := r.ReadSlice('\n')
		if room := max - len(line); room > 0 {
			if len(chunk) > room {
				chunk, cut = chunk[:room], true
			}
			line = append(line, chunk...)
		} else if len(chunk) > 0 {
			cut = true
		}
		if err == bufio.ErrBufferFull {
			continue
		}
		if cut {
			for i := 0; i < utf8.UTFMax && len(line) > 0 && !utf8.Valid(line); i++ {
				line = line[:len(line)-1]
			}
		}
		return line, err
	}
}
