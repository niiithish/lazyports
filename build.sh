#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "${ROOT}"

"${ROOT}/scripts/go" build -o lazyports .
echo "Built ./lazyports"
