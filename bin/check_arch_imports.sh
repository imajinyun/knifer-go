#!/usr/bin/env bash
set -euo pipefail

if [ -n "${ARCH_CHECK_ROOT:-}" ]; then
	ROOT_DIR="${ARCH_CHECK_ROOT}"
else
	ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
fi

cd "$(dirname "${BASH_SOURCE[0]}")/.."
go run ./bin/archimportscheck -root "${ROOT_DIR}"
