#!/usr/bin/env bash
set -euo pipefail

if [ -n "${CI_WORKFLOW_ROOT:-}" ]; then
	ROOT_DIR="${CI_WORKFLOW_ROOT}"
else
	ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
fi

cd "$(dirname "${BASH_SOURCE[0]}")/.."
go run ./bin/ciworkflowcheck -root "${ROOT_DIR}"
