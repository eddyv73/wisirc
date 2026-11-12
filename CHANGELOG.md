# Changelog

## v0.4.3

- Fixed: ≡ → Edit config did nothing (the screen just flickered) when
  neither `$VISUAL` nor `$EDITOR` was set and `vi` was not installed. wisirc
  now opens the first of `nano`, `vim`, `vi` or `nvim` it finds, and if
  there is no editor (or the one in `$EDITOR` is missing) the status bar
  says so instead of hiding the error behind the dashboard.

## v0.4.2

- `wisirc --config-copy <path>` starts with a copy of a config you already
  have: a wisirc config directory or `config.toml` (with its `servers.toml`
  and `prefs.toml`), or a Halloy config, whose servers are imported. It
  never replaces an existing config, the copies are readable only by you,
  and the first-run setup is skipped.

## v0.4.1

Everything from v0.4.1-rc.1 to rc.5, now stable:

- **First-run setup:** the first time wisirc runs it asks for the language,
  your nick and real name, a theme (previewed as you move) and a first
  server: Libera.Chat, import from Halloy, your own server or none for now.
  Esc skips it. The nick and real name are kept in `prefs.toml` and fill in
  every new server and `/connect`. The first run no longer writes an example
  `config.toml` with Libera.Chat.
- **`/export halloy [file]`** (`/exportar`) writes your saved servers as
  Halloy `[servers.*]` tables, passwords and SASL included. Disabled servers
  get `autoconnect = false`; the file is readable only by you and is never
  overwritten.
- **Fixed: saved changes now apply right away.** Editing a server's
  address, port, TLS, nick or credentials in the Servers screen reconnects
  it with the new settings (before, it kept retrying the old address until
  a restart). Editing `config.toml` from ≡ → Edit config applies every
  change without restarting: changed servers reconnect, removed ones
  disconnect, and `[theme]`, `[notifications]` and `[updates]` take effect.
- **Updates:** wisirc tells you when a new release is out (status bar and ≡
  menu) and `/update` installs it, checked against `SHA256SUMS`. Homebrew
  installs are told to `brew upgrade`. `[updates] check = false` turns it
  off.

## v0.4.1-rc.5

- The first time wisirc runs, a short setup asks for the language, your nick
  and real name, a theme (previewed as you move) and a first server:
  Libera.Chat, import from Halloy, your own server or none for now. Esc
  skips it. The nick and real name are kept in `prefs.toml` and fill in
  every new server and `/connect`.
- The first run no longer writes an example `config.toml` with Libera.Chat:
  it starts empty and the setup adds the server you pick.
- `/export halloy [file]` (`/exportar`) writes your saved servers as Halloy
  `[servers.*]` tables, passwords and SASL included, ready to paste into
  Halloy's `config.toml`. Disabled servers get `autoconnect = false`. The
  file is created readable only by you and is never overwritten (default
  `halloy-export.toml` next to `config.toml`).

## v0.4.1-rc.4

- Editing `config.toml` from ≡ → Edit config now applies every change
  without restarting: servers with a new address, port, TLS, nick or
  credentials reconnect, removed servers disconnect, new servers stay off if
  you disabled them, and `[theme]` (unless a theme was chosen in the menu),
  `[notifications]` and `[updates] check = false` take effect at once.
- Tests that change each setting and check what it does. (v0.4.1-rc.3 was
  tagged but not released: one of these tests depended on the screen width.)

## v0.4.1-rc.2

- Fixed: editing a connected (or retrying) server's address, port, TLS,
  nick or credentials now reconnects it with the new settings. Before, it
  kept retrying the old address until wisirc was restarted. Changing only
  the channels leaves the connection alone.

## v0.4.1-rc.1

- wisirc tells you when a new release is out (status bar and ≡ menu) and
  `/update` installs it: the release binary, checked against `SHA256SUMS`,
  replaces the running one and is used the next time wisirc starts.
  Homebrew installs are told to `brew upgrade`. `[updates] check = false`
  turns the check off.

## v0.4.0

Same as v0.4.0-rc.1, now stable:


- Servers from `config.toml` are edited with the same form; the changes go
  to `servers.toml` (which now wins for the same name) and **Undo edit**
  brings the `config.toml` version back. `config.toml` is never written.
- Servers can be disabled and enabled (Servers screen, right-click menu,
  dashboard `x`): a disabled server is disconnected, stops retrying and does
  not connect on startup. Kept in `prefs.toml`.
- Dashboard: `e` edits the selected server; the right-click menu shows
  **Disconnect** while a server is still retrying.

## v0.3.9

- Emoji no longer push the nick list and borders out of place: variation
  selectors, skin tones and emoji joiners are dropped on screen, so a row
  has the same width in every terminal. Logs keep the original text.

## v0.3.8

- README redesigned as a landing page: animated banner, a mosaic of six
  themes, a short tour, and the full reference moved to `docs/guide.md`
  (`docs/guide.es.md` in Spanish).

## v0.3.7

- New README: banner, feature grid, the reference folded into sections;
  the installer diagram is gone.

## v0.3.6

- The install check waits for GitHub to point "latest" at a new release
  before testing it; CI actions updated to Node 24.

## v0.3.5

- Releases are built and published by GitHub Actions when a version tag is
  pushed, and installed on Linux and macOS right after to check them.

## v0.3.4

- The installer is tested on every release, weekly and when it changes, on
  Linux (x86-64 and ARM) and macOS: install, upgrade, pinning a version,
  `--add-to-path`, `--uninstall` and `--source`.
- The README shows the install steps as a diagram.

## v0.3.3

- MIT license, Arch Linux (AUR) package, man page, CI on Linux and macOS.
- Go module path is now `github.com/eddyv73/wisirc`, so
  `go install github.com/eddyv73/wisirc/cmd/wisirc@latest` works and reports
  its version.
- Tests, docs and screenshots use neutral example data.

## v0.3.2

- One-line installer: `curl -fsSL …/install.sh | sh`, with `--version`,
  `--source`, `--add-to-path` and `--uninstall`; safe against truncated
  downloads; reports upgrades.

## v0.3.1

- Protection against floods and hostile servers: 8 KiB line cap, no copies
  when a buffer is full, log files kept open, batched events, animation paused
  without focus, at most 100 private windows per network.
- Richer status bar: a dot per server, tls badge, clickable unread chips,
  mention counter, nick and channel modes, clock.
- Clicking the "wisirc" title opens the dashboard.

## v0.3.0

- Live dashboard (F2, opens at startup) with animated lag graphs.
- English and Spanish UI with a language selector.
- Per-server colors in the sidebar; right-click a server → Color….
- One-click import of Halloy servers.
- 20 themes, all checked for contrast.

## v0.2.0

- `/connect` options (`-user`, `-pass`, `-sasl`, `-notls`, …) and a connect
  screen with examples; save prompt and Servers screen.
- Mouse selection copies to the clipboard; animated lag graph; themes.

## v0.1.0

- First release: multiple servers with TLS and SASL, sidebar, nick list,
  mouse menus, logs and history, notifications on mentions.
