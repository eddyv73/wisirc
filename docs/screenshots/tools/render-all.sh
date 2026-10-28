#!/bin/sh
# render-all.sh <dir-with-captures>
# Renders every README screenshot from its tmux capture (<scene>.ansi) into
# docs/screenshots/<scene>.png at 1.5x.
set -eu
here=$(cd "$(dirname "$0")" && pwd)
src=$1
dst=$(cd "$here/.." && pwd)
export SCALE=${SCALE:-1.5}
r() { "$here/render.sh" "$src/$1.ansi" "$dst/$1.png" "$2" ${3:+--from "$3" --to "$4"}; }
r dashboard   nocturno
r chat        nocturno
r menu        nocturno
r themes      catppuccin-mocha "#f5c2e7" "#89b4fa"
r theme-papel papel            "#fde2c4" "#c9d6ff"
r servers     nocturno
r server-form nocturno
r connect     nocturno
r spanish     nocturno         "#ffd6a5" "#c9d6ff"
