#!/usr/bin/env bash
#
# check_api_freeze.sh validates v1 API freeze/deprecation governance metadata.

set -euo pipefail

cd "$(dirname "$0")/.."

AI_CONTEXT_FILE="${AI_CONTEXT_FILE:-ai-context.json}"
TOOLS_JSON_FILE="${TOOLS_JSON_FILE:-docs/api/tools.json}"

go run ./bin/apifreezecheck -ai-context "${AI_CONTEXT_FILE}" -tools "${TOOLS_JSON_FILE}"
