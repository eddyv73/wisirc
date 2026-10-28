#!/usr/bin/env python3
"""Turn a `tmux capture-pane -e -p` dump into a pretty HTML "terminal window".

    python3 -I ansi2html.py scene.ansi scene.html --theme nocturno [--title wisirc]

The theme's default background/foreground are read from internal/ui/theme.go
(the terminal's OSC 10/11 colors are not part of a capture). Standard library
only. Wide characters (emoji, CJK) get a 2-cell box and every other non-ASCII
glyph a 1-cell box, so columns stay aligned whatever font fallback happens.
"""
import argparse
import html
import os
import re
import sys
import unicodedata

HERE = os.path.dirname(os.path.abspath(__file__))
THEME_GO = os.path.join(HERE, "..", "..", "..", "internal", "ui", "theme.go")

BASE16 = [
    "#1d1f21", "#cc6666", "#b5bd68", "#f0c674", "#81a2be", "#b294bb", "#8abeb7", "#c5c8c6",
    "#666666", "#d54e53", "#b9ca4a", "#e7c547", "#7aa6da", "#c397d8", "#70c0b1", "#eaeaea",
]


def xterm256(n):
    if n < 16:
        return BASE16[n]
    if n < 232:
        n -= 16
        steps = [0, 95, 135, 175, 215, 255]
        return "#%02x%02x%02x" % (steps[n // 36], steps[(n // 6) % 6], steps[n % 6])
    v = 8 + (n - 232) * 10
    return "#%02x%02x%02x" % (v, v, v)


def theme_colors(name):
    """(bg, fg) for a theme, from theme.go; None means the terminal default."""
    try:
        src = open(THEME_GO, encoding="utf-8").read()
    except OSError:
        return None, None
    m = re.search(r'name:\s*"%s",\s*bg:\s*c\("(#[0-9a-fA-F]{6})"\),\s*fg:\s*c\("(#[0-9a-fA-F]{6})"\)' % re.escape(name), src)
    if m:
        return m.group(1), m.group(2)
    return None, None


class Style:
    __slots__ = ("fg", "bg", "bold", "faint", "italic", "under", "rev")

    def __init__(self):
        self.reset()

    def reset(self):
        self.fg = self.bg = None
        self.bold = self.faint = self.italic = self.under = self.rev = False

    def key(self):
        return (self.fg, self.bg, self.bold, self.faint, self.italic, self.under, self.rev)


def apply_sgr(st, params):
    p = [int(x) if x.isdigit() else 0 for x in params.split(";")] if params else [0]
    i = 0
    while i < len(p):
        n = p[i]
        if n == 0:
            st.reset()
        elif n == 1:
            st.bold = True
        elif n == 2:
            st.faint = True
        elif n == 3:
            st.italic = True
        elif n == 4:
            st.under = True
        elif n == 7:
            st.rev = True
        elif n == 22:
            st.bold = st.faint = False
        elif n == 23:
            st.italic = False
        elif n == 24:
            st.under = False
        elif n == 27:
            st.rev = False
        elif 30 <= n <= 37:
            st.fg = xterm256(n - 30)
        elif 90 <= n <= 97:
            st.fg = xterm256(n - 90 + 8)
        elif n == 39:
            st.fg = None
        elif 40 <= n <= 47:
            st.bg = xterm256(n - 40)
        elif 100 <= n <= 107:
            st.bg = xterm256(n - 100 + 8)
        elif n == 49:
            st.bg = None
        elif n in (38, 48) and i + 1 < len(p):
            if p[i + 1] == 5 and i + 2 < len(p):
                col = xterm256(p[i + 2])
                i += 2
            elif p[i + 1] == 2 and i + 4 < len(p):
                col = "#%02x%02x%02x" % tuple(p[i + 2:i + 5])
                i += 4
            else:
                col = None
            if n == 38:
                st.fg = col
            else:
                st.bg = col
        i += 1


ESC = re.compile(r"\x1b\[([0-9;:]*)([A-Za-z])|\x1b\][^\x07\x1b]*(?:\x07|\x1b\\)|\x1b.")


def cell(ch):
    if ord(ch) < 128:
        return html.escape(ch)
    if 0x2581 <= ord(ch) <= 0x2588:
        # lower blocks (sparklines): draw them so they really fill k/8 of the row
        pct = (ord(ch) - 0x2580) * 12.5
        return '<span class="w1 blk" style="background:linear-gradient(to top,currentColor %g%%,transparent %g%%)"></span>' % (pct, pct)
    w = 2 if unicodedata.east_asian_width(ch) in ("W", "F") else 1
    cls = "w2" if w == 2 else "w1"
    if 0x2500 <= ord(ch) <= 0x257F:
        cls += " bx"
    return '<span class="%s">%s</span>' % (cls, html.escape(ch))


def convert(text, dfg, dbg, rows=None):
    lines = text.split("\n")
    if lines and lines[-1] == "":
        lines.pop()
    if rows:
        lines = (lines + [""] * rows)[:rows]
    out = []
    st = Style()
    for line in lines:
        runs, buf, cur = [], [], st.key()
        pos = 0
        for m in ESC.finditer(line):
            if m.start() > pos:
                buf.append(line[pos:m.start()])
            pos = m.end()
            if m.group(2) == "m":
                if buf:
                    runs.append((cur, "".join(buf)))
                    buf = []
                apply_sgr(st, m.group(1).replace(":", ";"))
                cur = st.key()
        if pos < len(line):
            buf.append(line[pos:])
        if buf:
            runs.append((cur, "".join(buf)))
        parts = []
        for (fg, bg, bold, faint, italic, under, rev), s in runs:
            if rev:
                fg, bg = (bg or dbg), (fg or dfg)
            css = []
            if fg:
                css.append("color:" + fg)
            if bg:
                css.append("background:" + bg)
            if bold:
                css.append("font-weight:700")
            if faint:
                css.append("opacity:.6")
            if italic:
                css.append("font-style:italic")
            if under:
                css.append("text-decoration:underline")
            body = "".join(cell(c) for c in s)
            parts.append('<span style="%s">%s</span>' % (";".join(css), body) if css else body)
        out.append("".join(parts))
    return "\n".join(out)


def mix(a, b, t):
    a = [int(a[i:i + 2], 16) for i in (1, 3, 5)]
    b = [int(b[i:i + 2], 16) for i in (1, 3, 5)]
    return "#%02x%02x%02x" % tuple(round(x + (y - x) * t) for x, y in zip(a, b))


PAGE = """<!doctype html>
<html><head><meta charset="utf-8"><title>{title}</title>
<style>
html,body{{margin:0;padding:0}}
body{{background:{backdrop};display:inline-block;padding:{pad}px}}
.win{{border-radius:12px;overflow:hidden;background:{bg};
  box-shadow:0 0 0 1px rgba(0,0,0,.35),0 1px 0 1px rgba(255,255,255,.06) inset,0 24px 60px rgba(10,12,30,.45),0 6px 18px rgba(10,12,30,.25)}}
.bar{{height:34px;display:flex;align-items:center;position:relative;
  background:{barbg};border-bottom:1px solid rgba(0,0,0,.35)}}
.dots{{display:flex;gap:8px;padding-left:14px}}
.dots i{{width:12px;height:12px;border-radius:50%;display:block;box-shadow:inset 0 0 0 .5px rgba(0,0,0,.25)}}
.t{{position:absolute;left:0;right:0;text-align:center;font:600 13px -apple-system,BlinkMacSystemFont,"Helvetica Neue",sans-serif;
  color:{titlefg};letter-spacing:.2px;pointer-events:none}}
pre{{margin:0;padding:12px 16px 14px;color:{fg};background:{bg};
  font-family:Menlo,"DejaVu Sans Mono for Powerline","Literation Mono Powerline",monospace;
  font-size:{size}px;line-height:{lh};font-variant-ligatures:none;font-kerning:none;
  -webkit-font-smoothing:antialiased;white-space:pre}}
.w1,.w2{{display:inline-block;text-align:center;overflow:visible;vertical-align:top}}
.w1{{width:1ch}} .w2{{width:2ch}}
.blk{{height:{lh}em;vertical-align:top}}
.bx{{transform:scaleY({bxscale});transform-origin:50% 55%}}
</style></head><body><div class="win"><div class="bar"><div class="dots">
<i style="background:#ff5f57"></i><i style="background:#febc2e"></i><i style="background:#28c840"></i></div>
<div class="t">{title}</div></div><pre>{body}</pre></div>
<script>
var b=document.body.getBoundingClientRect();document.title=Math.ceil(b.width)+"x"+Math.ceil(b.height);
</script></body></html>
"""


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("src")
    ap.add_argument("dst")
    ap.add_argument("--theme", default="nocturno")
    ap.add_argument("--title", default="wisirc")
    ap.add_argument("--rows", type=int, default=0)
    ap.add_argument("--size", type=float, default=14)
    ap.add_argument("--lh", type=float, default=1.22)
    ap.add_argument("--pad", type=int, default=44)
    ap.add_argument("--from", dest="g1", default=None, help="backdrop gradient start color")
    ap.add_argument("--to", dest="g2", default=None, help="backdrop gradient end color")
    a = ap.parse_args()

    bg, fg = theme_colors(a.theme)
    bg, fg = bg or "#1c1c1c", fg or "#d0d0d0"
    text = open(a.src, encoding="utf-8", errors="replace").read()
    body = convert(text, fg, bg, a.rows or None)
    g1 = a.g1 or "#c9d6ff"
    g2 = a.g2 or "#e2c9ff"
    page = PAGE.format(
        title=html.escape(a.title), body=body, bg=bg, fg=fg, size=a.size, lh=a.lh, pad=a.pad,
        bxscale=round(a.lh / 1.12, 3),
        barbg=mix(bg, "#ffffff", 0.07), titlefg=mix(bg, fg, 0.7),
        backdrop="linear-gradient(135deg,%s 0%%,%s 100%%)" % (g1, g2),
    )
    open(a.dst, "w", encoding="utf-8").write(page)


if __name__ == "__main__":
    sys.exit(main())
