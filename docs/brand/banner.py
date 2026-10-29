# Generates docs/brand/banner-{dark,light}.svg: python3 -I docs/brand/banner.py docs/brand
import sys
MONO = "ui-monospace, SFMono-Regular, Menlo, Consolas, 'Liberation Mono', monospace"
SANS = "-apple-system, BlinkMacSystemFont, 'Segoe UI', 'Noto Sans', Helvetica, Arial, sans-serif"
P = {
 "dark": dict(bg1="#0d1117", bg2="#161b22", border="#30363d", bar="#161b22", text="#e6edf3", muted="#8b949e",
              g=["#3fb950", "#4fc3d9", "#b38cff"], panel="#0d1117", s=["#5fd7af", "#b38cff", "#ffaf5f"], chan="#ff8fc8", glow="0.35"),
 "light": dict(bg1="#ffffff", bg2="#f6f8fa", border="#d0d7de", bar="#f6f8fa", text="#1f2328", muted="#656d76",
               g=["#1a7f37", "#0969da", "#8250df"], panel="#ffffff", s=["#1a7f37", "#8250df", "#bc4c00"], chan="#bf3989", glow="0.18"),
}
CMD = "/join #wisirc"
CW = 10.2  # glyph advance at 17px
TX = round(56 + 12*CW, 1)  # after "[kestrel] > "

def bars(x0, y0, color, seed):
    out = []
    import random
    r = random.Random(seed)
    for i in range(12):
        hs = [r.randint(4, 22) for _ in range(5)]
        hs.append(hs[0])
        vals_h = ";".join(str(h) for h in hs)
        vals_y = ";".join(str(y0 - h) for h in hs)
        out.append(f'<rect x="{x0 + i*9}" y="{y0-hs[0]}" width="6" height="{hs[0]}" rx="1.5" fill="{color}">'
                   f'<animate attributeName="height" values="{vals_h}" dur="3.2s" begin="-{i*0.27:.2f}s" repeatCount="indefinite"/>'
                   f'<animate attributeName="y" values="{vals_y}" dur="3.2s" begin="-{i*0.27:.2f}s" repeatCount="indefinite"/></rect>')
    return "".join(out)

def svg(mode):
    c = P[mode]
    w, h = 960, 300
    typed = ";".join(str(round(i*CW, 1)) for i in range(len(CMD)+1))
    keyt = ";".join(f"{i/ (len(CMD)) * 0.3:.3f}" for i in range(len(CMD)+1))
    servers = [("libera", c["s"][0], "38ms", 0.0), ("darkstar", c["s"][1], "71ms", 0.6), ("relaybox", c["s"][2], "112ms", 1.2)]
    rows = []
    for i, (name, col, ms, d) in enumerate(servers):
        y = 132 + i*44
        rows.append(f'''
  <circle cx="634" cy="{y-5}" r="5" fill="{col}"><animate attributeName="opacity" values="1;.35;1" dur="2.4s" begin="{d}s" repeatCount="indefinite"/></circle>
  <text x="650" y="{y}" font-family="{MONO}" font-size="15" fill="{c['text']}">{name}</text>
  <g opacity=".9">{bars(752, y+2, col, i+7)}</g>
  <text x="904" y="{y}" text-anchor="end" font-family="{MONO}" font-size="13" fill="{c['muted']}">{ms}</text>''')
    return f'''<svg xmlns="http://www.w3.org/2000/svg" width="{w}" height="{h}" viewBox="0 0 {w} {h}" role="img" aria-label="wisirc: IRC in your terminal">
  <defs>
    <linearGradient id="bg" x1="0" y1="0" x2="1" y2="1"><stop offset="0" stop-color="{c['bg1']}"/><stop offset="1" stop-color="{c['bg2']}"/></linearGradient>
    <linearGradient id="word" x1="0" y1="0" x2="1" y2="0" spreadMethod="reflect">
      <stop offset="0" stop-color="{c['g'][0]}"/><stop offset=".5" stop-color="{c['g'][1]}"/><stop offset="1" stop-color="{c['g'][2]}"/>
      <animateTransform attributeName="gradientTransform" type="translate" values="0 0;1 0;0 0" dur="9s" repeatCount="indefinite"/>
    </linearGradient>
    <radialGradient id="halo" cx=".22" cy=".45" r=".5"><stop offset="0" stop-color="{c['g'][1]}" stop-opacity="{c['glow']}"/><stop offset="1" stop-color="{c['g'][1]}" stop-opacity="0"/></radialGradient>
    <clipPath id="type"><rect x="{TX}" y="214" height="32" width="0">
      <animate attributeName="width" values="{typed}" keyTimes="{keyt}" calcMode="discrete" dur="6s" repeatCount="indefinite"/>
    </rect></clipPath>
  </defs>
  <rect x="1" y="1" width="{w-2}" height="{h-2}" rx="16" fill="url(#bg)" stroke="{c['border']}" stroke-width="2"/>
  <rect x="1" y="1" width="{w-2}" height="{h-2}" rx="16" fill="url(#halo)"/>
  <circle cx="30" cy="27" r="6" fill="#ff5f57"/><circle cx="50" cy="27" r="6" fill="#febc2e"/><circle cx="70" cy="27" r="6" fill="#28c840"/>
  <text x="{w/2:.0f}" y="32" text-anchor="middle" font-family="{SANS}" font-size="13" font-weight="600" fill="{c['muted']}">wisirc</text>
  <line x1="2" y1="52" x2="{w-2}" y2="52" stroke="{c['border']}"/>

  <text x="52" y="146" font-family="{MONO}" font-size="82" font-weight="800" letter-spacing="-1" fill="url(#word)">wisirc</text>
  <text x="56" y="186" font-family="{SANS}" font-size="20" fill="{c['text']}">IRC in your terminal, with the feel of a desktop app.</text>

  <text x="56" y="236" font-family="{MONO}" font-size="17" fill="{c['s'][0]}">[kestrel] &gt;</text>
  <g clip-path="url(#type)"><text x="{TX}" y="236" font-family="{MONO}" font-size="17" fill="{c['text']}">/join <tspan fill="{c['chan']}">#wisirc</tspan></text></g>
  <rect x="{TX}" y="221" width="9" height="19" fill="{c['text']}">
    <animate attributeName="x" values="{';'.join(str(round(TX+i*CW,1)) for i in range(len(CMD)+1))}" keyTimes="{keyt}" calcMode="discrete" dur="6s" repeatCount="indefinite"/>
    <animate attributeName="opacity" values="1;0;1" keyTimes="0;.5;1" calcMode="discrete" dur="1s" repeatCount="indefinite"/>
  </rect>

  <rect x="606" y="78" width="322" height="194" rx="12" fill="{c['panel']}" stroke="{c['border']}"/>
  <text x="626" y="104" font-family="{MONO}" font-size="12" font-weight="700" fill="{c['s'][0]}">● LIVE<animate attributeName="opacity" values="1;.4;1" dur="1.6s" repeatCount="indefinite"/></text>
  <text x="904" y="104" text-anchor="end" font-family="{MONO}" font-size="12" fill="{c['muted']}">3 servers · lag</text>
  {"".join(rows)}
</svg>
'''
for m in ("dark", "light"):
    open(f"{sys.argv[1]}/banner-{m}.svg", "w").write(svg(m))
