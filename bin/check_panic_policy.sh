#!/usr/bin/env bash
set -euo pipefail

if [ -n "${ARCH_CHECK_ROOT:-}" ]; then
	ROOT_DIR="${ARCH_CHECK_ROOT}"
else
	ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
fi

echo "panic policy: scanning production Go files"
cd "$(dirname "${BASH_SOURCE[0]}")/.."
go run ./bin/panicpolicycheck -root "${ROOT_DIR}"
