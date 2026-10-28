#!/bin/sh
# Records docs/demo.cast: wisirc against the fake servers in docs/screenshots/tools,
# driven by tmux. Needs Go, Python 3, tmux and asciinema 3. Never touches your config.
set -e
R=$(cd "$(dirname "$0")/../../.." && pwd); T=$(mktemp -d)
mkdir -p $T/cfg $T/home $T/data
(cd $R && GO111MODULE=on CGO_ENABLED=0 go build -o $T/wisirc ./cmd/wisirc)
cat > $T/cfg/config.toml <<'C'
[servers.libera-net]
  nickname = "kestrel"
  server = "127.0.0.1"
  port = 16601
  use_tls = false
  channels = ["#go", "#linux", "#homelab"]

[servers.darkstar]
  nickname = "kestrel"
  server = "127.0.0.1"
  port = 16602
  use_tls = false
  channels = ["#lobby", "#code"]

[servers.relaybox]
  nickname = "kestrel"
  server = "127.0.0.1"
  port = 16603
  use_tls = false
  channels = ["#selfhosted", "#irc"]
C
cat > $T/cfg/prefs.toml <<'C'
# Preferences chosen inside wisirc (theme, language, server colors). They win over config.toml.

theme = "nocturno"
lang = "en"

[server_colors]
  libera-net = "#5fd7af"
  darkstar = "#af87ff"
  relaybox = "#ffaf5f"
C
python3 -I $R/docs/screenshots/tools/fakeircd.py >/dev/null 2>&1 & SRV=$!
sleep 1
printf 'set -g status off\nset -g window-size latest\nset -g default-terminal xterm-256color\n' > $T/tmux.conf
tmux -L wdemo -f $T/tmux.conf new-session -d -s demo -x 120 -y 32 "env HOME=$T/home XDG_DATA_HOME=$T/data XDG_CONFIG_HOME=$T/home/.config TERM=xterm-256color COLORTERM=truecolor LANG=en_US.UTF-8 $T/wisirc --config $T/cfg/config.toml"
(
k() { tmux -L wdemo send-keys -t demo "$@"; }
t() { tmux -L wdemo send-keys -t demo -l "$1"; }
sleep 9                              # dashboard: lag graphs animate
k Escape; sleep 3                    # into the chat
k M-2; sleep 2.5                     # another channel
k M-1; sleep 1.5
k F1; sleep 2.5; k Escape; sleep 0.8 # the ≡ menu
t "/theme"; k Enter; sleep 1.2       # theme submenu with live preview
for i in 1 2 3 4 5 6; do k Down; sleep 0.9; done
k Escape; sleep 1.2                  # back to the saved theme
t "/lang es"; k Enter; sleep 2.5     # Spanish UI
k F2; sleep 5                        # dashboard in Spanish
k F2; sleep 1.5
t "/lang en"; k Enter; sleep 1
t "/quit"; k Enter
) &
asciinema rec --headless --window-size 120x32 -i 1.5 -q --overwrite -c "tmux -L wdemo attach -t demo" $R/docs/demo.cast
tmux -L wdemo kill-server 2>/dev/null || true; kill $SRV; rm -rf $T
