#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
INSTALL_DIR="${INSTALL_DIR:-$HOME/.local/bin}"

cd "$ROOT"

if [[ -x "$ROOT/build.sh" ]]; then
  "$ROOT/build.sh"
else
  if ! command -v go >/dev/null 2>&1; then
    if [[ -x "$HOME/.local/go/bin/go" ]]; then
      export PATH="$HOME/.local/go/bin:$PATH"
    else
      echo "Go is required. Install from https://go.dev/dl/ or run: curl -fsSL https://go.dev/dl/go1.23.4.linux-amd64.tar.gz | sudo tar -C /usr/local -xzf -" >&2
      exit 1
    fi
  fi
  go build -o lazyports .
fi

mkdir -p "$INSTALL_DIR"
install -m 755 lazyports "$INSTALL_DIR/lazyports"

echo "Installed lazyports to $INSTALL_DIR/lazyports"
echo "Run: lazyports"
if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
  echo "Add this to your shell rc: export PATH=\"$INSTALL_DIR:\$PATH\""
fi
