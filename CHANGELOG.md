# Changelog

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
