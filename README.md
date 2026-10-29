<p align="center">
  <picture>
    <source media="(prefers-color-scheme: dark)" srcset="docs/brand/banner-dark.svg">
    <img src="docs/brand/banner-light.svg" alt="wisirc: IRC in your terminal" width="760">
  </picture>
</p>

<p align="center">
  <a href="https://github.com/eddyv73/wisirc/releases/latest"><img src="https://img.shields.io/github/v/release/eddyv73/wisirc?style=flat-square&color=3fb950&label=release" alt="latest release"></a>
  <a href="https://github.com/eddyv73/wisirc/actions/workflows/ci.yml"><img src="https://img.shields.io/github/actions/workflow/status/eddyv73/wisirc/ci.yml?branch=main&style=flat-square&label=tests" alt="tests"></a>
  <a href="https://github.com/eddyv73/wisirc/actions/workflows/install.yml"><img src="https://img.shields.io/github/actions/workflow/status/eddyv73/wisirc/install.yml?branch=main&style=flat-square&label=installer" alt="installer"></a>
  <img src="https://img.shields.io/github/go-mod/go-version/eddyv73/wisirc?style=flat-square&color=4fc3d9" alt="Go version">
  <img src="https://img.shields.io/badge/platform-macOS%20%7C%20Linux-6ea8ff?style=flat-square" alt="macOS and Linux">
  <a href="LICENSE"><img src="https://img.shields.io/github/license/eddyv73/wisirc?style=flat-square&color=b38cff" alt="MIT license"></a>
</p>

<p align="center">
  <a href="#install">Install</a> ·
  <a href="#features">Features</a> ·
  <a href="#screenshots">Screenshots</a> ·
  <a href="#documentation">Documentation</a> ·
  <a href="README.es.md">Español</a>
</p>

<p align="center">
  <img src="docs/screenshots/chat.png" alt="wisirc: servers and channels on the left, the chat in the middle, users on the right and a live status bar" width="900">
</p>

**wisirc** is an IRC client for the terminal that feels like a desktop app. Servers and channels live in a sidebar, the chat sits in the middle and the users on the right. Everything works with the mouse (clicks, wheel, context menus) and everything works from the keyboard, the irssi way. One static binary, no runtime, no config needed to start.

## Install

```sh
curl -fsSL https://raw.githubusercontent.com/eddyv73/wisirc/main/install.sh | sh
```

Then run `wisirc`. The installer picks the binary for your system (macOS or Linux, Intel or ARM), checks it against `SHA256SUMS`, puts it in `~/.local/bin` and tells you if that folder is missing from your `PATH`. Run it again to upgrade. Every release is installed this way on Linux and macOS by CI before you get it.

| Other ways | How |
|---|---|
| **Arch Linux / Garuda** | `makepkg -si` in [`packaging/aur`](packaging/aur), with man page and menu entry (AUR package coming soon) |
| **Go** | `go install github.com/eddyv73/wisirc/cmd/wisirc@latest` |
| **Binaries** | [latest release](https://github.com/eddyv73/wisirc/releases/latest): `darwin-arm64`, `darwin-amd64`, `linux-amd64`, `linux-arm64` |
| **From source** | `make build` (Go 1.27) |

<details>
<summary><b>Installer options</b></summary>

Options go after `sh -s --`:

```sh
curl -fsSL https://raw.githubusercontent.com/eddyv73/wisirc/main/install.sh | sh -s -- --version v0.3.1
curl -fsSL https://raw.githubusercontent.com/eddyv73/wisirc/main/install.sh | sh -s -- --add-to-path
curl -fsSL https://raw.githubusercontent.com/eddyv73/wisirc/main/install.sh | sh -s -- --source     # build it yourself (git + Go)
curl -fsSL https://raw.githubusercontent.com/eddyv73/wisirc/main/install.sh | sh -s -- --uninstall
```

| Option / variable | What it does |
|---|---|
| `--version vX.Y.Z` | install that release instead of the latest |
| `--source` | build from source: this checkout, or a fresh clone |
| `--add-to-path` | add the install folder to your shell profile (zsh, bash or fish), once |
| `--uninstall` | remove wisirc; your settings and logs stay |
| `WISIRC_PREFIX=/usr/local` | install into another place (`$WISIRC_PREFIX/bin`) |
| `WISIRC_REPO=owner/name` | install from a fork |
| `GITHUB_TOKEN` | for a private fork (or log in with `gh auth login`) |

From a clone, `./install.sh` takes the same options.

</details>

<details>
<summary><b>Manual install and building from source</b></summary>

Download the binary for your system from the release (or `dist/` after `make release`) and put it on your `PATH`:

| System | Binary |
|---|---|
| macOS Apple Silicon | `wisirc-darwin-arm64` |
| macOS Intel | `wisirc-darwin-amd64` |
| Linux x86-64 | `wisirc-linux-amd64` |
| Linux ARM64 | `wisirc-linux-arm64` |

```sh
install -m 0755 dist/wisirc-darwin-arm64 ~/.local/bin/wisirc
```

Building from source (requires Go 1.27):

```sh
make build            # ./wisirc
make install          # copies to ~/.local/bin (PREFIX=/usr/local for another destination)
make release VERSION=v0.1.0   # the 4 binaries in dist/
make test
```

</details>

## Features

<table>
<tr>
<td width="50%" valign="top">

**🖱️ Mouse and keyboard, both first class**<br>
Click channels, right-click servers, channels and nicks for their menus, drag borders to resize, select text to copy it. Or never touch the mouse.

</td>
<td width="50%" valign="top">

**📊 Live dashboard**<br>
One card per server with connection state, uptime, an animated lag graph, channels and unread messages. Opens at startup, F2 anytime.

</td>
</tr>
<tr>
<td valign="top">

**🎨 20 themes, previewed live**<br>
Dracula, Catppuccin, Tokyo Night, Rosé Pine, Kanagawa, CRT green and amber… all checked for readable contrast (WCAG AA). Each server can have its own color.

</td>
<td valign="top">

**🔐 Many servers, TLS and SASL**<br>
Connect to several networks at once, with SASL PLAIN, server and bouncer (ZNC) passwords and self-signed certificates when you allow them.

</td>
</tr>
<tr>
<td valign="top">

**🧭 No config file needed**<br>
`/connect irc.libera.chat -nick you -join #linux`, then save it with one click. Servers are managed from a screen with a form, and Halloy users import everything in one click.

</td>
<td valign="top">

**🌎 English and Spanish**<br>
The whole interface follows your system language, and you can switch it live from the menu or with `/lang`.

</td>
</tr>
<tr>
<td valign="top">

**🔔 Mentions that reach you**<br>
Highlighted lines, marked channels, a mention counter in the status bar and a desktop notification when you are not looking.

</td>
<td valign="top">

**🗂️ Logs and history**<br>
One log file per channel and day; opening a channel brings back its recent history, and the full log is one click away.

</td>
</tr>
</table>

## Quick start

1. Run `wisirc`. It opens on the dashboard; press **Esc** to go to the chat.
2. Press **F1** (or click **≡**) for the main menu: connect, servers, themes, language, help.
3. Connect with `/connect irc.libera.chat -nick yournick -join #linux`, and say **Yes** when wisirc offers to save it.
4. Switch channels with a click or **Alt+1…9**; right-click anything for its menu.

## Screenshots

<table>
<tr>
<td width="50%"><img src="docs/screenshots/dashboard.png" alt="Dashboard"><br><sub><b>Dashboard</b> (F2): live lag graphs, uptime, channels and unread counts per server</sub></td>
<td width="50%"><img src="docs/screenshots/menu.png" alt="Main menu"><br><sub><b>≡ menu</b> (F1 or click): everything is one click away</sub></td>
</tr>
<tr>
<td><img src="docs/screenshots/themes.png" alt="Theme picker"><br><sub><b>Themes</b>: arrows preview live (here catppuccin-mocha)</sub></td>
<td><img src="docs/screenshots/theme-papel.png" alt="papel theme"><br><sub>The light <b>papel</b> theme</sub></td>
</tr>
<tr>
<td><img src="docs/screenshots/servers.png" alt="Servers screen"><br><sub><b>≡ → Servers</b>: connect, edit, delete, import from Halloy</sub></td>
<td><img src="docs/screenshots/server-form.png" alt="Server form"><br><sub>Server form, BNC friendly (<code>user/network</code>)</sub></td>
</tr>
<tr>
<td><img src="docs/screenshots/connect.png" alt="Connect screen"><br><sub><code>/connect</code> with ready-to-adapt examples</sub></td>
<td><img src="docs/screenshots/spanish.png" alt="Spanish interface"><br><sub>Spanish interface (<code>/lang es</code>) and a nick's right-click menu</sub></td>
</tr>
</table>

A 38-second terminal recording is in [`docs/demo.cast`](docs/demo.cast): `asciinema play docs/demo.cast`.

## Documentation

`man wisirc` has the short version. Everything else, by topic:

<details>
<summary><b>⌨️ Keyboard shortcuts and commands</b></summary>

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

</details>

<details>
<summary><b>🖱️ Mouse</b></summary>

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

</details>

<details>
<summary><b>🔌 Connecting to a server</b></summary>

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

</details>

<details>
<summary><b>🗄️ Servers screen</b></summary>

≡ → **Servers** lists every server with its state (● connected, ◌ connecting, ○ disconnected) and where it comes from:

- `servers.toml`: saved by wisirc. Servers are created, edited and deleted from this screen with a form (name, host, port, TLS, invalid certificate, nick, `user/network` username for a BNC, password, SASL, channels and *Connect on startup*).
- `config.toml`: your hand-written file. *Edit* opens the editor; these servers cannot be deleted from the screen, so your comments are never lost.
- unsaved: connections made with `/connect`; save them with *Save*.

Buttons: **[+ New]**, **[Edit]**, **[Connect]** / **[Disconnect]**, **[Save]** (unsaved connections only), **[Delete]**, **[Close]** and, when a Halloy config is found, **[Import from Halloy]**.

Keys: ↑/↓ choose · Enter edit · `n` new · `c` connect/disconnect · `s` save · `d` delete (asks for confirmation) · `i` import from Halloy · Esc close. In the form: Tab/↓ next field, Space toggles checkboxes, Enter saves, Esc goes back.

If a name is in both files, `config.toml` wins. With `autoconnect = false` a server stays saved but does not connect on startup.

> In macOS Terminal.app or iTerm2, enable "Use Option as Meta key" so the `Alt` shortcuts work.

</details>

<details>
<summary><b>📥 Coming from Halloy</b></summary>

If you use [Halloy](https://halloy.chat), wisirc finds its config (`~/Library/Application Support/halloy/config.toml` on macOS, `~/.config/halloy/config.toml` on Linux). In ≡ → **Servers** an **[Import from Halloy]** button appears (or press `i`): it brings in every server you do not have yet, saves them to `servers.toml` and connects them. Names that already exist are left untouched.

Imported fields: nick, username, realname, host, port, TLS (on by default in Halloy), invalid certificates, channels, `password`, `password_file` and SASL PLAIN (also with `password_file`). For security, `password_command` is **not executed**; wisirc warns you so you can enter that password by hand. `nick_password` and SASL EXTERNAL are not supported yet and are reported as well.

The first time you open wisirc without a config, if there is a Halloy install on the machine, it starts straight on that screen.

</details>

<details>
<summary><b>⚙️ Configuration file</b></summary>

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

</details>

<details>
<summary><b>🎨 Themes</b></summary>

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

</details>

<details>
<summary><b>🌈 Server colors</b></summary>

Each server has its own color in the sidebar: the selected row is filled with it, and the server's ▾ and its channels are written in it, so every group is easy to tell apart. By default servers get shades of green in sidebar order. Right-click a server → **Color…** to pick another one (green, dark green, lime, jade, cyan, blue, purple, pink, red, orange, amber, yellow, gray) or **Automatic**; arrows preview it live, Enter keeps it, Esc reverts. The choice is stored per server in `prefs.toml`. Unread channels are bold in the server's color; mentions keep the theme's alert color.

</details>

<details>
<summary><b>📊 Dashboard</b></summary>

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

</details>

<details>
<summary><b>🌎 Language</b></summary>

The interface is available in English and Spanish. By default it follows the system locale (`LC_ALL`, then `LC_MESSAGES`, then `LANG`): for example `LANG=es_MX.UTF-8` gives Spanish, and anything else gives English.

To change it, open ≡ → **Language / Idioma**: moving with ↑/↓ previews the language live, Enter keeps it and Esc reverts to the previous one. You can also type `/lang en` or `/lang es` (aliases `/idioma` and `/language`). The choice is stored in `~/.config/wisirc/prefs.toml`.

</details>

<details>
<summary><b>🗂️ Logs and mentions</b></summary>

One file per day and channel is written to `~/.local/share/wisirc/logs/<server>/<channel>/<YYYY-MM-DD>.log` (respects `$XDG_DATA_HOME`). When you open a channel its last 200 lines are loaded as history, and **≡ → View logs** shows the full log of the current channel.

When someone writes your nick or sends you a private message, the line is highlighted and the channel is marked in the sidebar. If you are not looking at that channel, or the terminal does not have focus, you also get a system notification: on macOS through `osascript` and on Linux through `notify-send`. If neither is available, the terminal bell rings.

</details>

## Contributing

Issues and pull requests are welcome: see [CONTRIBUTING.md](CONTRIBUTING.md). What changed in each version is in [CHANGELOG.md](CHANGELOG.md).

## License

[MIT](LICENSE) © 2026 eddyv73
