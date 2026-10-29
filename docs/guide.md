<a id="top"></a>

# wisirc guide

Everything wisirc does, by topic. Back to the [README](../README.md) · [en español](guide.es.md) · `man wisirc` has the short version.

[⌨️ Keyboard shortcuts and commands](#keyboard-shortcuts-and-commands) ·
[🖱️ Mouse](#mouse) ·
[🔌 Connecting to a server](#connecting-to-a-server) ·
[🗄️ Servers screen](#servers-screen) ·
[📥 Coming from Halloy](#coming-from-halloy) ·
[⚙️ Configuration file](#configuration-file) ·
[🎨 Themes](#themes) ·
[🌈 Server colors](#server-colors) ·
[📊 Dashboard](#dashboard) ·
[🌎 Language](#language) ·
[🗂️ Logs and mentions](#logs-and-mentions)

<a id="keyboard-shortcuts-and-commands"></a>

## ⌨️ Keyboard shortcuts and commands

| Key | Action |
|---|---|
| `F1` | Main menu |
| `Alt+1..9` | Go to buffer N |
| `Ctrl+N` / `Ctrl+P` (or `Alt+→` / `Alt+←`) | Next / previous buffer |
| `Alt+A` | Next buffer with activity (mentions first) |
| `PgUp` / `PgDn` | Scroll the history |
| `↑` / `↓` | History of what you typed |
| `Tab` | Complete nick |
| `Esc` | Close menu or panel |
| `Ctrl+C` | Quit |

Commands: `/join`, `/part`, `/msg`, `/query`, `/me`, `/nick`, `/whois`, `/topic`, `/kick`, `/ban`, `/mode`, `/quit`, `/connect`, `/close`, `/quote`, `/theme`, `/lang`, `/help`. To send text that starts with `/`, type `//`.

<p align="right"><a href="#top">↑</a></p>

<a id="mouse"></a>

## 🖱️ Mouse

| Action | Result |
|---|---|
| Click **≡** | Main menu (connect, servers, join a channel, logs, theme, language, config, help, quit) |
| Click a channel in the sidebar | Switches to that channel |
| Click ▾ / ▸ | Collapses or expands a server |
| Right-click a server | Connect/Disconnect, Join channel…, Color… |
| Right-click a channel | Leave (part), Close, Mark as read, View topic |
| Click (or right-click) a nick | Whois, Private message, Op/Voice, Kick, Ban |
| Drag over the chat or a help panel | Selects and copies to the clipboard (pbcopy, wl-copy/xclip/xsel; over SSH, OSC 52) |
| Wheel | Scrolls the history, the sidebar or the nicklist |
| Drag a `│` border | Changes the width of the sidebar or the nicklist |

<p align="right"><a href="#top">↑</a></p>

<a id="connecting-to-a-server"></a>

## 🔌 Connecting to a server

`/connect` (or `/conectar`) with no arguments, or ≡ → *Connect to server*, opens a screen with examples. Syntax:

```
/connect <name>                            server saved in the config
/connect <host> [port] [options]
```

| Option | Purpose |
|---|---|
| `-nick <nick>` | your nick |
| `-user <user>` | username (on a BNC/ZNC: `user/network`) |
| `-pass <password>` | server or BNC password |
| `-sasl <user:password>` | SASL login |
| `-join <#a,#b>` | channels to join on connect |
| `-tls` / `-notls` | force TLS on or off (by default only on port 6697) |
| `-insecure` | accept invalid certificates |
| `-name <name>` | name shown in the sidebar |

Examples:

```
/connect 192.0.2.20 7000 -notls -nick pepe -user pepe/libera -pass myZncPassword
/connect irc.libera.chat -nick pepe -sasl pepe:myPassword123 -join #linux
/connect irc.example.org 6667 -nick pepe -pass serverPassword
```

When the server answers, wisirc asks **Save &lt;name&gt; permanently?** (Y/N or click). If you say yes, it is saved to `~/.config/wisirc/servers.toml` and connects automatically every time you open wisirc.

<p align="right"><a href="#top">↑</a></p>

<a id="servers-screen"></a>

## 🗄️ Servers screen

≡ → **Servers** lists every server with its state (● connected, ◌ connecting, ○ disconnected) and where it comes from:

- `servers.toml`: saved by wisirc. Servers are created, edited and deleted from this screen with a form (name, host, port, TLS, invalid certificate, nick, `user/network` username for a BNC, password, SASL, channels and *Connect on startup*).
- `config.toml`: your hand-written file. *Edit* opens the editor; these servers cannot be deleted from the screen, so your comments are never lost.
- unsaved: connections made with `/connect`; save them with *Save*.

Buttons: **[+ New]**, **[Edit]**, **[Connect]** / **[Disconnect]**, **[Save]** (unsaved connections only), **[Delete]**, **[Close]** and, when a Halloy config is found, **[Import from Halloy]**.

Keys: ↑/↓ choose · Enter edit · `n` new · `c` connect/disconnect · `s` save · `d` delete (asks for confirmation) · `i` import from Halloy · Esc close. In the form: Tab/↓ next field, Space toggles checkboxes, Enter saves, Esc goes back.

If a name is in both files, `config.toml` wins. With `autoconnect = false` a server stays saved but does not connect on startup.

> In macOS Terminal.app or iTerm2, enable "Use Option as Meta key" so the `Alt` shortcuts work.

<p align="right"><a href="#top">↑</a></p>

<a id="coming-from-halloy"></a>

## 📥 Coming from Halloy

If you use [Halloy](https://halloy.chat), wisirc finds its config (`~/Library/Application Support/halloy/config.toml` on macOS, `~/.config/halloy/config.toml` on Linux). In ≡ → **Servers** an **[Import from Halloy]** button appears (or press `i`): it brings in every server you do not have yet, saves them to `servers.toml` and connects them. Names that already exist are left untouched.

Imported fields: nick, username, realname, host, port, TLS (on by default in Halloy), invalid certificates, channels, `password`, `password_file` and SASL PLAIN (also with `password_file`). For security, `password_command` is **not executed**; wisirc warns you so you can enter that password by hand. `nick_password` and SASL EXTERNAL are not supported yet and are reported as well.

The first time you open wisirc without a config, if there is a Halloy install on the machine, it starts straight on that screen.

<p align="right"><a href="#top">↑</a></p>

<a id="configuration-file"></a>

## ⚙️ Configuration file

The file lives at `~/.config/wisirc/config.toml` (respects `$XDG_CONFIG_HOME`). To use a different one: `wisirc --config <path>`. An example is created the first time you run wisirc.

```toml
[servers.examplenet]
  nickname = "alice"
  username = "alice"
  server = "irc.example.net"
  port = 6697
  channels = ["#lobby"]
  dangerously_accept_invalid_certs = true   # only for self-signed certificates

  [servers.examplenet.sasl.plain]
  username = "alice"
  password = "hunter2"

[notifications]
  enabled = true

[theme]
  name = "oscuro"   # oscuro, claro, nocturno, clasico, alto-contraste, bosque, glaciar, papel, ...
```

Per-server fields: `nickname`, `username`, `realname`, `server`, `port`, `use_tls` (on by default on port 6697), `channels`, `dangerously_accept_invalid_certs`, `password` (server PASS), `sasl.plain` and `autoconnect`.

<p align="right"><a href="#top">↑</a></p>

<a id="themes"></a>

## 🎨 Themes

≡ → **Theme / colors** opens a submenu with the themes; the current one is marked with ●. Moving with ↑/↓ applies the theme live; Enter keeps it, Esc goes back to the previous one. With the mouse, clicking a theme applies it. You can also use `/theme <name>` (or `/tema`), and `/theme` on its own opens the submenu.

Theme names are identifiers and stay the same in both languages (`dark` and `light` are accepted as aliases of `oscuro` and `claro`).

| Theme | Style |
|---|---|
| `oscuro` | your terminal's colors (the classic one) |
| `claro` | for light-background terminals |
| `nocturno` | charcoal with a cyan accent, herdr style |
| `clasico` | an irssi tribute, blue status bar |
| `alto-contraste` | pure black and white, accessible |
| `bosque` | warm dark |
| `glaciar` | cool dark |
| `papel` | paper-like light |
| `dracula` | violet dark with neon accents |
| `catppuccin-mocha` | pastel dark, mauve and blue |
| `catppuccin-latte` | pastel light |
| `tokyo-night` | navy blue with electric accents |
| `one-dark` | Atom-style charcoal gray |
| `solarized-oscuro` | petrol blue, Solarized tones |
| `rose-pine` | night purple, rose and gold |
| `kanagawa` | Japanese ink, warm tones |
| `everforest` | soft forest green |
| `monokai` | dark olive, vivid colors |
| `fosforo` | green CRT monitor |
| `ambar` | amber CRT monitor |

The colors of the well-known schemes were lightened where needed so that all text has a contrast of at least 4.5:1 (WCAG AA); a test checks this. Themes with their own background change the terminal background while wisirc is open and restore it on exit. The theme chosen from the menu is stored in `~/.config/wisirc/prefs.toml` and takes precedence over `[theme]` in the config.

Keys that wisirc does not know (for example, from a halloy config) are ignored and shown as a warning. wisirc never modifies your config. If you change it from **≡ → Edit config**, `$EDITOR` opens, and when you close it the file is reloaded and any new servers are connected.

> The password is stored in plain text. Protect the file with `chmod 600 ~/.config/wisirc/config.toml`.

<p align="right"><a href="#top">↑</a></p>

<a id="server-colors"></a>

## 🌈 Server colors

Each server has its own color in the sidebar: the selected row is filled with it, and the server's ▾ and its channels are written in it, so every group is easy to tell apart. By default servers get shades of green in sidebar order. Right-click a server → **Color…** to pick another one (green, dark green, lime, jade, cyan, blue, purple, pink, red, orange, amber, yellow, gray) or **Automatic**; arrows preview it live, Enter keeps it, Esc reverts. The choice is stored per server in `prefs.toml`. Unread channels are bold in the server's color; mentions keep the theme's alert color.

<p align="right"><a href="#top">↑</a></p>

<a id="dashboard"></a>

## 📊 Dashboard

wisirc starts on the dashboard; Esc or F2 takes you to the chat. **F2** (or ≡ → **Dashboard**, or `/panel` / `/dashboard`) opens it again: a live overview of all your servers, one card each:

```
 ● LIVE  servers: 2 · connected: 2 · average 120ms
 ╔═ znc ══════════════════════════╗ ╭─ libera ───────────────────────╮
 ║ ● connected            ⏱ 2h 14m ║ │ ○ disconnected                 │
 ║ ▂▃▂▄▂▃▂  ● 42ms                ║ │ no lag samples yet             │
 ║ #lobby #code                   ║ │ no channels                    │
 ║ 3 new · 1 mention              ║ │ all read                       │
 ╚════════════════════════════════╝ ╰────────────────────────────────╯
```

Each card shows the connection state, uptime, an animated lag graph (green/yellow/red), joined channels and unread messages and mentions; card borders use the server's sidebar color. While the dashboard is open, lag is measured every 5 seconds instead of every 30. Click a card (or Enter) to go to that server; right click (or `c`) connects or disconnects it; arrows move between cards; F2 or Esc closes it.

<p align="right"><a href="#top">↑</a></p>

<a id="language"></a>

## 🌎 Language

The interface is available in English and Spanish. By default it follows the system locale (`LC_ALL`, then `LC_MESSAGES`, then `LANG`): for example `LANG=es_MX.UTF-8` gives Spanish, and anything else gives English.

To change it, open ≡ → **Language / Idioma**: moving with ↑/↓ previews the language live, Enter keeps it and Esc reverts to the previous one. You can also type `/lang en` or `/lang es` (aliases `/idioma` and `/language`). The choice is stored in `~/.config/wisirc/prefs.toml`.

<p align="right"><a href="#top">↑</a></p>

<a id="logs-and-mentions"></a>

## 🗂️ Logs and mentions

One file per day and channel is written to `~/.local/share/wisirc/logs/<server>/<channel>/<YYYY-MM-DD>.log` (respects `$XDG_DATA_HOME`). When you open a channel its last 200 lines are loaded as history, and **≡ → View logs** shows the full log of the current channel.

When someone writes your nick or sends you a private message, the line is highlighted and the channel is marked in the sidebar. If you are not looking at that channel, or the terminal does not have focus, you also get a system notification: on macOS through `osascript` and on Linux through `notify-send`. If neither is available, the terminal bell rings.

<p align="right"><a href="#top">↑</a></p>
