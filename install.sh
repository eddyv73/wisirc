#!/bin/sh
# wisirc installer for macOS and Linux.
#
#   curl -fsSL https://raw.githubusercontent.com/eddyv73/wisirc/main/install.sh | sh
#   curl -fsSL https://raw.githubusercontent.com/eddyv73/wisirc/main/install.sh | sh -s -- --version v0.3.1
#
# Run ./install.sh --help (or pipe it to "sh -s -- --help") for every option.
# Everything runs inside main(), called on the last line, so a download that
# is cut halfway never executes a partial script.

set -eu

REPO="${WISIRC_REPO:-eddyv73/wisirc}"
PREFIX="${WISIRC_PREFIX:-$HOME/.local}"
BIN_DIR="$PREFIX/bin"
VERSION="latest"
MODE="release"
ADD_TO_PATH=0
DESKTOP_FILE="${XDG_DATA_HOME:-$HOME/.local/share}/applications/wisirc.desktop"

if [ -t 1 ]; then
	B='\033[1m' C='\033[1;36m' Y='\033[1;33m' R='\033[1;31m' G='\033[1;32m' N='\033[0m'
else
	B='' C='' Y='' R='' G='' N=''
fi
say() { printf '%b==>%b %s\n' "$C" "$N" "$*"; }
ok() { printf '%b✓%b %s\n' "$G" "$N" "$*"; }
warn() { printf '%bwarning:%b %s\n' "$Y" "$N" "$*" >&2; }
die() { printf '%berror:%b %s\n' "$R" "$N" "$*" >&2; exit 1; }
have() { command -v "$1" >/dev/null 2>&1; }

usage() {
	cat <<EOF
wisirc installer for macOS and Linux

  curl -fsSL https://raw.githubusercontent.com/$REPO/main/install.sh | sh
  curl -fsSL https://raw.githubusercontent.com/$REPO/main/install.sh | sh -s -- [options]
  ./install.sh [options]

Options:
  --version vX.Y.Z   install that release instead of the latest
  --source           build from source (this checkout, or a fresh clone); needs git and Go
  --add-to-path      add the install folder to your shell profile if it is missing
  --uninstall        remove wisirc (your settings and logs are kept)
  -h, --help         show this help

Environment:
  WISIRC_PREFIX      install into \$WISIRC_PREFIX/bin (default: ~/.local)
  WISIRC_REPO        install from another GitHub repository (default: $REPO)
  GITHUB_TOKEN       token for a private repository (or log in with "gh auth login")

The binary is checked against the release's SHA256SUMS before it is installed.
EOF
}

parse_args() {
	while [ $# -gt 0 ]; do
		case "$1" in
		--version)
			[ $# -ge 2 ] || die "--version needs a tag, e.g. v0.3.1"
			VERSION="$2"
			shift
			;;
		--source) MODE="source" ;;
		--add-to-path) ADD_TO_PATH=1 ;;
		--uninstall) MODE="uninstall" ;;
		-h | --help)
			usage
			exit 0
			;;
		*) die "unknown option: $1 (see --help)" ;;
		esac
		shift
	done
}

detect_platform() {
	case "$(uname -s)" in
	Darwin) OS="darwin" ;;
	Linux) OS="linux" ;;
	*) die "unsupported system: $(uname -s) (wisirc runs on macOS and Linux)" ;;
	esac
	case "$(uname -m)" in
	x86_64 | amd64) ARCH="amd64" ;;
	arm64 | aarch64) ARCH="arm64" ;;
	*) die "unsupported CPU: $(uname -m)" ;;
	esac
	ASSET="wisirc-$OS-$ARCH"
}

sha256() {
	if have sha256sum; then
		sha256sum "$1" | cut -d' ' -f1
	else
		shasum -a 256 "$1" | cut -d' ' -f1
	fi
}

# http_get URL OUT [TOKEN ACCEPT]: download with curl or wget, optionally
# authenticated against the GitHub API.
http_get() {
	if have curl; then
		if [ $# -eq 4 ]; then
			curl -fsSL -H "Authorization: Bearer $3" -H "Accept: $4" -o "$2" "$1"
		else
			curl -fsSL -o "$2" "$1"
		fi
	elif have wget; then
		if [ $# -eq 4 ]; then
			wget -q --header="Authorization: Bearer $3" --header="Accept: $4" -O "$2" "$1"
		else
			wget -q -O "$2" "$1"
		fi
	else
		die "need curl or wget"
	fi
}

# fetch_with_token NAME: one release asset through the GitHub API.
fetch_with_token() {
	if [ "$VERSION" = "latest" ]; then
		rel="https://api.github.com/repos/$REPO/releases/latest"
	else
		rel="https://api.github.com/repos/$REPO/releases/tags/$VERSION"
	fi
	if [ ! -f "$TMP/release.json" ]; then
		http_get "$rel" "$TMP/release.json" "$TOKEN" "application/vnd.github+json" ||
			die "could not read the release (is the token allowed to see $REPO?)"
	fi
	# find the asset id that belongs to NAME without needing jq
	id=$(tr ',' '\n' <"$TMP/release.json" | grep -E '"(id|name)"' |
		awk -v want="\"name\": \"$1\"" '/"id"/ { id = $2 } index($0, want) { print id; exit }')
	[ -n "$id" ] || die "release has no asset named $1"
	http_get "https://api.github.com/repos/$REPO/releases/assets/$id" "$TMP/$1" \
		"$TOKEN" "application/octet-stream"
}

fetch_public() {
	if [ "$VERSION" = "latest" ]; then
		http_get "https://github.com/$REPO/releases/latest/download/$1" "$TMP/$1"
	else
		http_get "https://github.com/$REPO/releases/download/$VERSION/$1" "$TMP/$1"
	fi
}

download() {
	TOKEN="${GITHUB_TOKEN:-${GH_TOKEN:-}}"
	if fetch_public "$ASSET" 2>/dev/null && fetch_public SHA256SUMS 2>/dev/null; then
		say "downloaded $ASSET"
	elif have gh && gh auth status >/dev/null 2>&1 && {
		tag=""
		[ "$VERSION" = "latest" ] || tag="$VERSION"
		# shellcheck disable=SC2086 # an empty $tag means "latest"
		gh release download $tag --repo "$REPO" --dir "$TMP" --clobber --pattern "$ASSET" --pattern SHA256SUMS 2>/dev/null
	}; then
		say "downloaded $ASSET with gh"
	elif [ -n "$TOKEN" ]; then
		fetch_with_token "$ASSET"
		fetch_with_token SHA256SUMS
		say "downloaded $ASSET with GITHUB_TOKEN"
	else
		die "could not download $ASSET from $REPO ($VERSION). If the repository is private, log in with 'gh auth login' or set GITHUB_TOKEN; or install with --source"
	fi
	want=$(grep " $ASSET\$" "$TMP/SHA256SUMS" | cut -d' ' -f1)
	[ -n "$want" ] || die "SHA256SUMS has no entry for $ASSET"
	[ "$(sha256 "$TMP/$ASSET")" = "$want" ] || die "checksum mismatch for $ASSET, not installing"
	ok "checksum verified"
	mv "$TMP/$ASSET" "$TMP/wisirc"
}

build_from_source() {
	have go || die "--source needs Go 1.27 or newer (https://go.dev/dl)"
	# this checkout when the script runs from one, otherwise a fresh clone
	src=""
	case "$0" in */install.sh | install.sh) src="$(cd "$(dirname "$0")" && pwd)" ;; esac
	if [ -z "$src" ] || [ ! -f "$src/go.mod" ]; then
		have git || die "--source needs git to fetch the code"
		src="$TMP/src"
		ref=""
		[ "$VERSION" = "latest" ] || ref="--branch $VERSION"
		say "cloning $REPO"
		# shellcheck disable=SC2086 # empty $ref means the default branch
		git clone -q --depth 1 $ref "https://github.com/$REPO.git" "$src" 2>/dev/null ||
			{ have gh && gh repo clone "$REPO" "$src" -- -q --depth 1 $ref 2>/dev/null; } ||
			die "could not clone $REPO"
	fi
	say "building from $src"
	ver=$(git -C "$src" describe --tags --always --dirty 2>/dev/null || echo dev)
	(cd "$src" && GO111MODULE=on CGO_ENABLED=0 go build -trimpath \
		-ldflags "-s -w -X main.Version=$ver" -o "$TMP/wisirc" ./cmd/wisirc)
	ok "built wisirc $ver"
}

install_desktop_entry() {
	[ "$OS" = "linux" ] || return 0
	[ -n "${XDG_CURRENT_DESKTOP:-}${DISPLAY:-}${WAYLAND_DISPLAY:-}" ] || return 0
	term=""
	for t in gnome-terminal konsole xfce4-terminal kgx kitty alacritty foot wezterm xterm; do
		if have "$t"; then
			term="$t"
			break
		fi
	done
	[ -n "$term" ] || return 0
	case "$term" in
	gnome-terminal) exec_line="gnome-terminal --title=wisirc --geometry=140x40 -- $BIN_DIR/wisirc" ;;
	wezterm) exec_line="wezterm start -- $BIN_DIR/wisirc" ;;
	*) exec_line="$term -e $BIN_DIR/wisirc" ;;
	esac
	mkdir -p "$(dirname "$DESKTOP_FILE")"
	cat >"$DESKTOP_FILE" <<EOF
[Desktop Entry]
Type=Application
Name=wisirc
GenericName=IRC Client
Comment=Terminal IRC client with mouse, menus and a live dashboard
Exec=$exec_line
Icon=utilities-terminal
Terminal=false
Categories=Network;Chat;IRCClient;
Keywords=irc;chat;
EOF
	if have update-desktop-database; then
		update-desktop-database "$(dirname "$DESKTOP_FILE")" 2>/dev/null || true
	fi
	ok "added wisirc to the applications menu ($term)"
}

# profile_line prints the profile file and the line that puts BIN_DIR on PATH.
profile_line() {
	case "$(basename "${SHELL:-sh}")" in
	zsh) echo "$HOME/.zshrc|export PATH=\"$BIN_DIR:\$PATH\"" ;;
	bash)
		if [ "$OS" = "darwin" ]; then
			echo "$HOME/.bash_profile|export PATH=\"$BIN_DIR:\$PATH\""
		else
			echo "$HOME/.bashrc|export PATH=\"$BIN_DIR:\$PATH\""
		fi
		;;
	fish) echo "${XDG_CONFIG_HOME:-$HOME/.config}/fish/config.fish|fish_add_path $BIN_DIR" ;;
	*) echo "$HOME/.profile|export PATH=\"$BIN_DIR:\$PATH\"" ;;
	esac
}

check_path() {
	case ":$PATH:" in
	*":$BIN_DIR:"*) return 0 ;;
	esac
	file="$(profile_line | cut -d'|' -f1)"
	line="$(profile_line | cut -d'|' -f2-)"
	if [ "$ADD_TO_PATH" = 1 ]; then
		if [ -f "$file" ] && grep -qF "$line" "$file"; then
			ok "$file already adds $BIN_DIR to PATH; open a new terminal"
		else
			mkdir -p "$(dirname "$file")"
			printf '\n# added by the wisirc installer\n%s\n' "$line" >>"$file"
			ok "added $BIN_DIR to PATH in $file; open a new terminal"
		fi
		return 0
	fi
	warn "$BIN_DIR is not in your PATH. Add it with:"
	printf "    echo '%s' >> %s\n" "$line" "$file"
	printf "  or run the installer again with --add-to-path\n"
}

uninstall() {
	rm -f "$BIN_DIR/wisirc" "$DESKTOP_FILE"
	ok "removed $BIN_DIR/wisirc"
	say "your settings stay in ~/.config/wisirc and logs in ~/.local/share/wisirc"
}

main() {
	parse_args "$@"
	detect_platform
	if [ "$MODE" = "uninstall" ]; then
		uninstall
		return 0
	fi

	TMP="$(mktemp -d)"
	trap 'rm -rf "$TMP"' EXIT INT TERM

	printf '%bwisirc installer%b · %s/%s → %s\n' "$B" "$N" "$OS" "$ARCH" "$BIN_DIR"
	before=""
	if [ -x "$BIN_DIR/wisirc" ]; then
		before="$("$BIN_DIR/wisirc" --version 2>/dev/null | awk '{print $2}')"
	fi

	if [ "$MODE" = "source" ]; then build_from_source; else download; fi

	mkdir -p "$BIN_DIR"
	install -m 0755 "$TMP/wisirc" "$BIN_DIR/wisirc"
	if [ "$OS" = "darwin" ]; then
		xattr -d com.apple.quarantine "$BIN_DIR/wisirc" 2>/dev/null || true
	fi
	after="$("$BIN_DIR/wisirc" --version | awk '{print $2}')"
	if [ -n "$before" ] && [ "$before" != "$after" ]; then
		ok "upgraded wisirc $before → $after"
	else
		ok "installed wisirc $after to $BIN_DIR/wisirc"
	fi

	install_desktop_entry
	check_path
	printf '\n  Start it with: %bwisirc%b   (config: ~/.config/wisirc/config.toml)\n' "$B" "$N"
}

main "$@"
