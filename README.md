<p align="center">
  <picture>
    <source media="(prefers-color-scheme: dark)" srcset="docs/brand/banner-dark.svg">
    <img src="docs/brand/banner-light.svg" alt="wisirc: IRC in your terminal" width="820">
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

<p align="center"><a href="#install">Install</a> · <a href="#features">Features</a> · <a href="#themes">Themes</a> · <a href="#tour">Tour</a> · <a href="docs/guide.md">Guide</a> · <a href="README.es.md">Español</a></p>

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

## Quick start

1. Run `wisirc`. It opens on the dashboard; press **Esc** to go to the chat.
2. Press **F1** (or click **≡**) for the main menu: connect, servers, themes, language, help.
3. Connect with `/connect irc.libera.chat -nick yournick -join #linux`, and say **Yes** when wisirc offers to save it.
4. Switch channels with a click or **Alt+1…9**; right-click anything for its menu.

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

## Themes

20 built-in themes, previewed live as you move through the list (≡ → <b>Themes</b> or <code>/theme</code>). All of them keep text at WCAG AA contrast, and each server can have its own color.

<p align="center">
  <img src="docs/screenshots/themes-grid.png" alt="The same chat in six themes: tokyo-night, catppuccin-mocha, ambar, rose-pine, papel and fosforo" width="900">
</p>

## Tour

<table>
<tr>
<td width="50%" valign="top"><img src="docs/screenshots/dashboard.png" alt="Dashboard"><p align="center"><sub><b>Dashboard</b> · live lag, uptime and unread messages per server (F2)</sub></p></td>
<td width="50%" valign="top"><img src="docs/screenshots/menu.png" alt="Main menu"><p align="center"><sub><b>≡ menu</b> · everything one click away (F1)</sub></p></td>
</tr>
<tr>
<td width="50%" valign="top"><img src="docs/screenshots/servers.png" alt="Servers screen"><p align="center"><sub><b>Servers</b> · connect, edit, delete or import from Halloy</sub></p></td>
<td width="50%" valign="top"><img src="docs/screenshots/spanish.png" alt="Spanish interface"><p align="center"><sub><b>Español</b> · the whole UI in Spanish, and a nick's right-click menu</sub></p></td>
</tr>
</table>

A 38-second recording is in <a href="docs/demo.cast"><code>docs/demo.cast</code></a> (<code>asciinema play docs/demo.cast</code>).

## Documentation

The <a href="docs/guide.md"><b>guide</b></a> covers everything by topic, and <code>man wisirc</code> has the short version.

<table>
<tr><td width="33%"><a href="docs/guide.md#keyboard-shortcuts-and-commands">⌨️ Keyboard shortcuts and commands</a></td><td width="33%"><a href="docs/guide.md#mouse">🖱️ Mouse</a></td><td width="33%"><a href="docs/guide.md#connecting-to-a-server">🔌 Connecting to a server</a></td></tr>
<tr><td width="33%"><a href="docs/guide.md#servers-screen">🗄️ Servers screen</a></td><td width="33%"><a href="docs/guide.md#coming-from-halloy">📥 Coming from Halloy</a></td><td width="33%"><a href="docs/guide.md#configuration-file">⚙️ Configuration file</a></td></tr>
<tr><td width="33%"><a href="docs/guide.md#themes">🎨 Themes</a></td><td width="33%"><a href="docs/guide.md#server-colors">🌈 Server colors</a></td><td width="33%"><a href="docs/guide.md#dashboard">📊 Dashboard</a></td></tr>
<tr><td width="33%"><a href="docs/guide.md#language">🌎 Language</a></td><td width="33%"><a href="docs/guide.md#logs-and-mentions">🗂️ Logs and mentions</a></td><td width="33%"></td></tr>
</table>

## Contributing

Issues and pull requests are welcome: see [CONTRIBUTING.md](CONTRIBUTING.md). What changed in each version is in [CHANGELOG.md](CHANGELOG.md).

## License

[MIT](LICENSE) © 2026 eddyv73
