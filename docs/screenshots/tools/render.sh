#!/bin/sh
# render.sh <scene.ansi> <out.png> <theme> [extra ansi2html args...]
# Converts a tmux capture to HTML, measures it in headless Chrome and takes
# a 2x screenshot that fits the window exactly.
set -eu
CHROME=${CHROME:-"/Applications/Google Chrome.app/Contents/MacOS/Google Chrome"}
SCALE=${SCALE:-2}
here=$(cd "$(dirname "$0")" && pwd)
src=$1 out=$2 theme=$3
shift 3
html="${src%.ansi}.html"
python3 -I "$here/ansi2html.py" "$src" "$html" --theme "$theme" "$@"
size=$("$CHROME" --headless=new --disable-gpu --hide-scrollbars --window-size=3000,2000 \
	--dump-dom "file://$html" 2>/dev/null | sed -n 's:.*<title>\([0-9]*x[0-9]*\)</title>.*:\1:p')
w=${size%x*} h=${size#*x}
"$CHROME" --headless=new --disable-gpu --hide-scrollbars --force-device-scale-factor="$SCALE" \
	--window-size="$w,$h" --screenshot="$out" "file://$html" >/dev/null 2>&1
echo "$out ${w}x${h} @${SCALE}x"
