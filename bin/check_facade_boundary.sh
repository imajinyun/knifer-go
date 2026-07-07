#!/usr/bin/env bash
set -euo pipefail

if [ -n "${ARCH_CHECK_ROOT:-}" ]; then
	ROOT_DIR="${ARCH_CHECK_ROOT}"
else
	ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
fi

echo "facade boundary: scanning docs, unsafe opt-in, and thin facade rules"
cd "$(dirname "${BASH_SOURCE[0]}")/.."
go run ./bin/facadeboundarycheck -root "${ROOT_DIR}"
