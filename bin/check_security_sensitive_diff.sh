#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
AI_CONTEXT="${ROOT_DIR}/ai-context.json"
DIFF_FILTER="ACDMRTUXB"

if ! git -C "${ROOT_DIR}" rev-parse --is-inside-work-tree >/dev/null 2>&1; then
	echo "SECURITY DIFF CHECK ERROR: ${ROOT_DIR} is not inside a Git worktree" >&2
	exit 2
fi

changed_files_from_base() {
	local base_ref="${AGENT_CHANGE_BASE_REF:-}"
	if [ -z "${base_ref}" ] && [ -n "${GITHUB_BASE_REF:-}" ]; then
		base_ref="origin/${GITHUB_BASE_REF}"
	fi
	if [ -n "${base_ref}" ] && git -C "${ROOT_DIR}" rev-parse --verify --quiet "${base_ref}^{commit}" >/dev/null; then
		git -C "${ROOT_DIR}" diff --name-only --diff-filter="${DIFF_FILTER}" "${base_ref}...HEAD" --
	fi
}

changed_files_from_worktree() {
	git -C "${ROOT_DIR}" diff --name-only --diff-filter="${DIFF_FILTER}" HEAD --
	git -C "${ROOT_DIR}" diff --name-only --cached --diff-filter="${DIFF_FILTER}" --
	git -C "${ROOT_DIR}" ls-files --others --exclude-standard --
}

changed_files="$({
	changed_files_from_base
	changed_files_from_worktree
} | sort -u)"

if [ -z "${changed_files}" ]; then
	echo "security-sensitive diff check passed: no changed files under selected diff sources"
	exit 0
fi

matched_paths="$(
	python3 - "${AI_CONTEXT}" "${changed_files}" <<'PY'
import json
import sys

ai_context = sys.argv[1]
changed_files = [line.strip() for line in sys.argv[2].splitlines() if line.strip()]

with open(ai_context, "r", encoding="utf-8") as f:
    data = json.load(f)

facade_to_internal = {
    entry["package"]: entry["internal"].rstrip("/")
    for entry in data["public_facades"]
}

security_prefixes = set()
for package in data["security_sensitive_packages"]:
    security_prefixes.add(package.rstrip("/") + "/")
    internal = facade_to_internal.get(package)
    if internal:
        security_prefixes.add(internal.rstrip("/") + "/")

for path in changed_files:
    normalized = path.strip("/")
    if any(normalized.startswith(prefix) for prefix in security_prefixes):
        print(normalized)
PY
)"

if [ -z "${matched_paths}" ]; then
	echo "security-sensitive diff check passed: no security-sensitive paths changed"
	exit 0
fi

echo "SECURITY DIFF CHECK ERROR: security-sensitive files changed:" >&2
printf '%s\n' "${matched_paths}" | while IFS= read -r path; do echo "  - ${path}" >&2; done
echo "Run make agent-security-check and document security review evidence before merging." >&2
exit 1
