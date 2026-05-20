#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
GO="${ROOT}/scripts/go"
VERSION="${GO_VERSION:-1.24.2}"
ARCH="${GO_ARCH:-linux-amd64}"
DEST="${HOME}/.local/go"
TARBALL="go${VERSION}.${ARCH}.tar.gz"
URL="https://go.dev/dl/${TARBALL}"

if [[ -x "${DEST}/bin/go" ]]; then
  echo "Go already installed at ${DEST}/bin/go"
  "${DEST}/bin/go" version
  exit 0
fi

if command -v go >/dev/null 2>&1; then
  echo "Go already available: $(command -v go)"
  go version
  exit 0
fi

command -v curl >/dev/null 2>&1 || { echo "curl is required" >&2; exit 1; }

TMP="$(mktemp -d)"
trap 'rm -rf "${TMP}"' EXIT

echo "Downloading Go ${VERSION}..."
curl -fsSL "${URL}" -o "${TMP}/${TARBALL}"
mkdir -p "${HOME}/.local"
rm -rf "${DEST}"
tar -C "${HOME}/.local" -xzf "${TMP}/${TARBALL}"

echo "Installed Go to ${DEST}/bin/go"
"${DEST}/bin/go" version
echo
echo "Add this to your shell rc (~/.bashrc):"
echo '  export PATH="$HOME/.local/go/bin:$PATH"'
