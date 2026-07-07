#!/usr/bin/env bash
set -euo pipefail

if [ -n "${PROVIDER_CONTRACT_ROOT:-}" ]; then
	ROOT_DIR="${PROVIDER_CONTRACT_ROOT}"
else
	ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
fi

cd "$(dirname "${BASH_SOURCE[0]}")/.."
go run ./bin/providercontractcheck -root "${ROOT_DIR}"
