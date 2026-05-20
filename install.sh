#!/usr/bin/env bash
set -euo pipefail

INSTALL_DIR="${INSTALL_DIR:-$HOME/.local/bin}"
REPO="github.com/niiithish/lazyports"
GIT_URL="https://github.com/niiithish/lazyports.git"

ensure_go() {
	if command -v go >/dev/null 2>&1; then
		return 0
	fi
	if [[ -x "$HOME/.local/go/bin/go" ]]; then
		export PATH="$HOME/.local/go/bin:$PATH"
		return 0
	fi
	echo "Go is required to build lazyports." >&2
	echo "Install Go from https://go.dev/dl/ then re-run this script." >&2
	exit 1
}

finish() {
	mkdir -p "$INSTALL_DIR"
	echo "Installed lazyports to $INSTALL_DIR/lazyports"
	echo "Run: lazyports"
	if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
		echo "Add this to your shell rc: export PATH=\"$INSTALL_DIR:\$PATH\""
	fi
}

install_from_source_dir() {
	local root="$1"
	cd "$root"
	if [[ -x "$root/build.sh" ]]; then
		"$root/build.sh"
	else
		ensure_go
		go build -o lazyports .
	fi
	mkdir -p "$INSTALL_DIR"
	install -m 755 lazyports "$INSTALL_DIR/lazyports"
	finish
}

install_remote() {
	ensure_go
	if GOBIN="$INSTALL_DIR" go install "${REPO}@latest"; then
		finish
		return 0
	fi

	if ! command -v git >/dev/null 2>&1; then
		echo "go install failed and git is not available for a source fallback." >&2
		exit 1
	fi

	local tmp
	tmp="$(mktemp -d)"
	trap 'rm -rf "$tmp"' EXIT
	git clone --depth 1 "$GIT_URL" "$tmp/lazyports"
	install_from_source_dir "$tmp/lazyports"
}

script_path="${BASH_SOURCE[0]:-}"
if [[ -n "$script_path" && "$script_path" != "bash" && -f "$script_path" ]]; then
	root="$(cd "$(dirname "$script_path")" && pwd)"
	if [[ -f "$root/go.mod" ]]; then
		install_from_source_dir "$root"
		exit 0
	fi
fi

install_remote
