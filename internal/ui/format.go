package ui

import (
	"hash/fnv"
	"image/color"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/eddyv73/wisirc/internal/irc"
)

// mIRC color index → xterm-256 index.
var mircToXterm = [16]int{15, 0, 4, 2, 9, 1, 5, 3, 11, 10, 6, 14, 12, 13, 8, 7}

// lightIRC is set on light themes: pale mIRC colors written on the theme's
// own background are swapped for darker relatives (white→black,
// yellow→orange, light green→green, light cyan→cyan, light blue→blue,
// pink→purple, light grey→grey). A sender-chosen background keeps them.
var (
	lightIRC  bool
	lightMIRC = [16]int{1, 1, 2, 3, 4, 5, 6, 7, 7, 3, 10, 10, 2, 6, 14, 14}
)

type ircStyle struct {
	bold, italic, underline, reverse bool
	fg, bg                           int // -1 = default
}

func (s ircStyle) plain() bool {
	return s == ircStyle{fg: -1, bg: -1}
}

func (s ircStyle) sgr() string {
	codes := []string{"0"}
	if s.bold {
		codes = append(codes, "1")
	}
	if s.italic {
		codes = append(codes, "3")
	}
	if s.underline {
		codes = append(codes, "4")
	}
	if s.reverse {
		codes = append(codes, "7")
	}
	if s.fg >= 0 {
		fg := s.fg
		if lightIRC && s.bg < 0 {
			fg = lightMIRC[fg]
		}
		codes = append(codes, "38;5;"+strconv.Itoa(mircToXterm[fg]))
	}
	if s.bg >= 0 {
		codes = append(codes, "48;5;"+strconv.Itoa(mircToXterm[s.bg]))
	}
	return "\x1b[" + strings.Join(codes, ";") + "m"
}

// readColor parses up to two digits at s[i:]; returns value (or -1) and new index.
func readColor(s string, i int) (int, int, bool) {
	j := i
	for j < len(s) && j-i < 2 && s[j] >= '0' && s[j] <= '9' {
		j++
	}
	if j == i {
		return -1, i, false
	}
	n, _ := strconv.Atoi(s[i:j])
	if n > 15 {
		n = -1
	}
	return n, j, true
}

// RenderIRC converts mIRC formatting codes to ANSI and drops other control
// characters, so text from the network can never drive the terminal.
func RenderIRC(text string) string {
	var b strings.Builder
	st := ircStyle{fg: -1, bg: -1}
	styled := false
	apply := func(n ircStyle) {
		if n != st {
			st = n
			b.WriteString(st.sgr())
			styled = true
		}
	}
	for i := 0; i < len(text); {
		c := text[i]
		n := st
		switch c {
		case 0x02:
			n.bold = !n.bold
		case 0x1D:
			n.italic = !n.italic
		case 0x1F:
			n.underline = !n.underline
		case 0x16:
			n.reverse = !n.reverse
		case 0x0F:
			n = ircStyle{fg: -1, bg: -1}
		case 0x03:
			fg, j, ok := readColor(text, i+1)
			if !ok {
				n.fg, n.bg = -1, -1
				i++
				apply(n)
				continue
			}
			n.fg = fg
			if j+1 < len(text) && text[j] == ',' && text[j+1] >= '0' && text[j+1] <= '9' {
				n.bg, j, _ = readColor(text, j+1)
			}
			apply(n)
			i = j
			continue
		default:
			if c < 0x20 || c == 0x7F {
				i++
				continue
			}
			r, size := utf8.DecodeRuneInString(text[i:])
			if isC1(r) {
				i += size
				continue
			}
			b.WriteString(text[i : i+size])
			i += size
			continue
		}
		apply(n)
		i++
	}
	if styled && !st.plain() {
		b.WriteString("\x1b[0m")
	}
	return b.String()
}

// isC1 reports Unicode C1 control characters (U+0080–U+009F); some
// terminals interpret them like ESC sequences (U+009B is CSI).
func isC1(r rune) bool { return r >= 0x80 && r <= 0x9F }

// sanitize removes control characters (used for nicks and labels).
func sanitize(s string) string {
	return strings.Map(func(r rune) rune {
		if r < 0x20 || r == 0x7F || isC1(r) {
			return -1
		}
		return r
	}, s)
}

// NickColor picks a stable color for a nickname from the active theme.
func NickColor(nick string) color.Color {
	h := fnv.New32a()
	h.Write([]byte(irc.Fold(nick)))
	return nickColors[h.Sum32()%uint32(len(nickColors))]
}
