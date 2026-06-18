#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
AI_CONTEXT="${ROOT_DIR}/ai-context.json"

changed_files="$({
	git -C "${ROOT_DIR}" diff --name-only --diff-filter=ACMRTUXB HEAD --
	git -C "${ROOT_DIR}" diff --name-only --cached --diff-filter=ACMRTUXB --
	git -C "${ROOT_DIR}" ls-files --others --exclude-standard --
} | sort -u)"

python3 - "${AI_CONTEXT}" "${changed_files}" <<'PY'
import json
import sys

ai_context = sys.argv[1]
changed_files = [line.strip().strip("/") for line in sys.argv[2].splitlines() if line.strip()]

with open(ai_context, "r", encoding="utf-8") as f:
    data = json.load(f)

policies = data["change_type_policies"]
facades = {entry["package"]: entry["internal"].rstrip("/") for entry in data["public_facades"]}

security_prefixes = set()
for package in data["security_sensitive_packages"]:
    security_prefixes.add(package.rstrip("/") + "/")
    internal = facades.get(package)
    if internal:
        security_prefixes.add(internal.rstrip("/") + "/")

detected = set()
matched = {name: [] for name in policies}

for path in changed_files:
    if path in {"go.mod", "go.sum"}:
        detected.add("dependency_change")
        matched["dependency_change"].append(path)

    if path == "ai-context.json" or path == "Makefile" or path.startswith(".github/") or path.startswith("bin/check_") or path.startswith("bin/agent_"):
        detected.add("ci_governance")
        matched["ci_governance"].append(path)

    if path == "docs/api/exports.txt":
        detected.add("public_api")
        matched["public_api"].append(path)

    if path.endswith(".md") or path in {"CLAUDE.md", "llms.txt"} or path.startswith("docs/"):
        detected.add("documentation")
        matched["documentation"].append(path)

    if any(path.startswith(prefix) for prefix in security_prefixes):
        detected.add("security_sensitive")
        matched["security_sensitive"].append(path)

    if any(path.startswith(package + "/") for package in facades):
        detected.add("public_api")
        matched["public_api"].append(path)

    if path.startswith("internal/") and not any(path.startswith(prefix) for prefix in security_prefixes):
        if path.endswith("_test.go"):
            detected.add("bug_fix")
            matched["bug_fix"].append(path)
        else:
            detected.add("internal_refactor")
            matched["internal_refactor"].append(path)

if not changed_files:
    print("change policy check passed: no changed files")
    sys.exit(0)

if not detected:
    detected.add("bug_fix")
    matched["bug_fix"].extend(changed_files)

unknown = sorted(detected - set(policies))
if unknown:
    print("CHANGE POLICY CHECK ERROR: detected unknown policies: " + ", ".join(unknown), file=sys.stderr)
    sys.exit(1)

required_commands = []
for policy in sorted(detected):
    for command in policies[policy].get("required_commands", []):
        if command not in required_commands:
            required_commands.append(command)

print("change policy check passed")
print("detected policies: " + ", ".join(sorted(detected)))
print("required commands: " + ", ".join(required_commands))
for policy in sorted(detected):
    paths = sorted(set(matched.get(policy, [])))
    if paths:
        print(f"{policy} paths:")
        for path in paths:
            print(f"  - {path}")
PY
