# Regenerating the screenshots

Everything here uses only Go (to build wisirc), the Python 3 standard library,
tmux and headless Google Chrome. Nothing connects to a real IRC network.

| File | What it does |
|---|---|
| `fakeircd.py` | Fake IRC server on 127.0.0.1:16601–16603 (three "networks") with a scripted chat, server-time backlog, varied PING delays and a few live messages |
| `ansi2html.py` | Turns a `tmux capture-pane -e` dump into an HTML terminal window (truecolor/256/16 SGR, theme background read from `internal/ui/theme.go`) |
| `render.sh` | `ansi2html.py` + headless Chrome: measures the page, then screenshots it |
| `render-all.sh` | Renders every `docs/screenshots/*.png` from a folder of captures |

## Steps

```sh
T=/tmp/wisirc-shots; mkdir -p $T/cfg $T/home $T/data $T/out
GO111MODULE=on CGO_ENABLED=0 go build -o $T/wisirc ./cmd/wisirc
python3 -I docs/screenshots/tools/fakeircd.py &
```

`$T/cfg/config.toml`: three servers (`libera-net`, `darkstar`, `relaybox`) on
`127.0.0.1` ports 16601/16602/16603 with `use_tls = false`, nick `kestrel`;
`$T/cfg/servers.toml`: an `oftc` server with `autoconnect = false` (the
disconnected card); `$T/cfg/prefs.toml`:

```toml
theme = "nocturno"
lang = "en"
[server_colors]
"libera-net" = "#5fd7af"
"darkstar" = "#af87ff"
"relaybox" = "#ffaf5f"
```

Run wisirc inside a private tmux server, isolated from your real config:

```sh
printf 'set -g status off\n' > $T/tmux.conf
tmux -L wshots -f $T/tmux.conf new-session -d -s shot -x 132 -y 22 \
  "env HOME=$T/home XDG_DATA_HOME=$T/data XDG_CONFIG_HOME=$T/home/.config \
   TERM=xterm-256color COLORTERM=truecolor $T/wisirc --config $T/cfg/config.toml"
```

(`TERM=xterm-256color` matters: with `tmux-256color` the app falls back to
256 colors.) Wait about a minute on the dashboard so the lag graphs fill up,
then drive the app with `tmux -L wshots send-keys -t shot ...` (F1, F2,
Escape, Down, Enter, `/theme papel`, `/lang es`, `/connect`; a right click
is `printf '\033[<2;X;YM'` sent with `send-keys -l`) and save each scene with:

```sh
tmux -L wshots capture-pane -e -p -N -t shot > $T/out/<scene>.ansi
```

Sizes used: 132×22 for dashboard/servers/server-form, 132×30 for chat, menu,
themes, theme-papel and spanish, 132×32 for connect. Finally:

```sh
docs/screenshots/tools/render-all.sh $T/out
```

## Theme mosaic

`themes-grid.png` is the `#go` chat at 110×24 in six themes (`tokyo-night`,
`catppuccin-mocha`, `ambar`, `rose-pine`, `papel`, `fosforo`), each started
fresh with `theme = "<name>"` in `$T/cfg/prefs.toml` (wisirc reads prefs.toml
next to `--config`). Each capture goes through `ansi2html.py --pad 0 --from
transparent --to transparent`, is shot at 2x with
`--default-background-color=00000000`, and the six PNGs are laid out 3×2 on a
transparent page so the mosaic works on light and dark GitHub.
