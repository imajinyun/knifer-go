#!/usr/bin/env bash
#
# check_docs_quickstart.sh validates the human-authored facade quickstart
# structure. ai-context.json is the source of truth for public facade packages.

set -euo pipefail

if [ -n "${DOCS_QUICKSTART_ROOT:-}" ]; then
	ROOT_DIR="${DOCS_QUICKSTART_ROOT}"
else
	ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
fi

cd "$(dirname "${BASH_SOURCE[0]}")/.."
go run ./bin/docsquickstartcheck -root "${ROOT_DIR}"
