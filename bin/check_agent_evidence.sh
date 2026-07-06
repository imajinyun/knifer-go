#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
AI_CONTEXT="${ROOT_DIR}/ai-context.json"
EVIDENCE_FILE="${AGENT_EVIDENCE_FILE:-/tmp/knifer-go-agent-validation.json}"

cd "${ROOT_DIR}"
go run ./bin/agentevidencecheck -root "${ROOT_DIR}" -ai-context "${AI_CONTEXT}" -evidence "${EVIDENCE_FILE}"
