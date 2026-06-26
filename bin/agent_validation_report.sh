#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
AI_CONTEXT="${ROOT_DIR}/ai-context.json"
OUTPUT_FILE="${AGENT_EVIDENCE_FILE:-/tmp/knifer-go-agent-validation.json}"

python3 - "${ROOT_DIR}" "${AI_CONTEXT}" "${OUTPUT_FILE}" <<'PY'
import json
import os
import subprocess
import sys
from datetime import datetime, timezone

root_dir, ai_context, output_file = sys.argv[1], sys.argv[2], sys.argv[3]
DIFF_FILTER = "ACDMRTUXB"


def git(args):
    result = subprocess.run(
        ["git", "-C", root_dir, *args],
        check=True,
        text=True,
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
    )
    return result.stdout.strip()


def run(args):
    result = subprocess.run(
        args,
        cwd=root_dir,
        text=True,
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
    )
    return {
        "cmd": " ".join(args),
        "status": "passed" if result.returncode == 0 else "failed",
        "exit_code": result.returncode,
        "stdout": result.stdout.strip(),
        "stderr": result.stderr.strip(),
    }


def command_attestation(result, source):
    return {
        "status": result["status"],
        "source": source,
        "cmd": result["cmd"],
        "exit_code": result["exit_code"],
        "stdout": result.get("stdout", ""),
        "stderr": result.get("stderr", ""),
    }


def change_base_ref():
    base_ref = os.environ.get("AGENT_CHANGE_BASE_REF")
    if not base_ref and os.environ.get("GITHUB_BASE_REF"):
        base_ref = "origin/" + os.environ["GITHUB_BASE_REF"]
    return base_ref or ""


def changed_files():
    files = set()
    base_ref = change_base_ref()
    if base_ref:
        result = subprocess.run(
            ["git", "-C", root_dir, "rev-parse", "--verify", "--quiet", base_ref + "^{commit}"],
            text=True,
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
        )
        if result.returncode == 0:
            output = git(["diff", "--name-only", "--diff-filter=" + DIFF_FILTER, base_ref + "...HEAD", "--"])
            for line in output.splitlines():
                line = line.strip().strip("/")
                if line:
                    files.add(line)
    for args in (
        ["diff", "--name-only", "--diff-filter=" + DIFF_FILTER, "HEAD", "--"],
        ["diff", "--name-only", "--cached", "--diff-filter=" + DIFF_FILTER, "--"],
        ["ls-files", "--others", "--exclude-standard", "--"],
    ):
        output = git(args)
        for line in output.splitlines():
            line = line.strip().strip("/")
            if line:
                files.add(line)
    return sorted(files)


with open(ai_context, "r", encoding="utf-8") as f:
    data = json.load(f)

files = changed_files()
facades = {entry["package"]: entry["internal"].rstrip("/") for entry in data["public_facades"]}
security_prefixes = set()
for package in data["security_sensitive_packages"]:
    security_prefixes.add(package.rstrip("/") + "/")
    internal = facades.get(package)
    if internal:
        security_prefixes.add(internal.rstrip("/") + "/")

detected_policies = set()
security_sensitive_paths = []
for path in files:
    if path in {"go.mod", "go.sum"}:
        detected_policies.add("dependency_change")
    if path == "ai-context.json" or path == "Makefile" or path.startswith(".github/") or path.startswith("bin/check_") or path.startswith("bin/agent_"):
        detected_policies.add("ci_governance")
    if path == "docs/api/exports.txt" or any(path.startswith(package + "/") for package in facades):
        detected_policies.add("public_api")
    if path.endswith(".md") or path in {"CLAUDE.md", "llms.txt"} or path.startswith("docs/"):
        detected_policies.add("documentation")
    if any(path.startswith(prefix) for prefix in security_prefixes):
        detected_policies.add("security_sensitive")
        security_sensitive_paths.append(path)
    if path.startswith("internal/") and not any(path.startswith(prefix) for prefix in security_prefixes):
        detected_policies.add("bug_fix" if path.endswith("_test.go") else "internal_refactor")

if files and not detected_policies:
    detected_policies.add("bug_fix")

required_commands = []
for policy in sorted(detected_policies):
    for command in data["change_type_policies"].get(policy, {}).get("required_commands", []):
        if command not in required_commands:
            required_commands.append(command)

command_risks = {
    name: data["commands"][name]["risk_level"]
    for name in required_commands
    if name in data["commands"]
}
risk_rank = {"low": 1, "medium": 2, "high": 3, "forbidden_for_agent": 4}
highest_risk = "low"
for risk in command_risks.values():
    if risk_rank[risk] > risk_rank[highest_risk]:
        highest_risk = risk

checks = {
    "ai_context_check": run(["bash", "bin/check_ai_context.sh"]),
    "security_sensitive_diff": run(["bash", "bin/check_security_sensitive_diff.sh"]),
    "change_policy_check": run(["bash", "bin/check_change_policy.sh"]),
}

command_attestations = {
    "ai_context_check": command_attestation(checks["ai_context_check"], "embedded_check"),
    "change_policy_check": command_attestation(checks["change_policy_check"], "embedded_check"),
    "security_sensitive_diff": command_attestation(checks["security_sensitive_diff"], "embedded_check"),
    "agent_evidence": {
        "status": "passed",
        "source": "current_process",
        "cmd": data["commands"]["agent_evidence"]["cmd"],
        "exit_code": 0,
    },
    "agent_evidence_check": {
        "status": "pending",
        "source": "post_generation",
        "cmd": data["commands"]["agent_evidence_check"]["cmd"],
        "reason": "validated by make agent-evidence-check after evidence generation",
    },
}
for command in required_commands:
    command_attestations.setdefault(command, {
        "status": "not_recorded",
        "source": "required_by_policy",
        "cmd": data["commands"].get(command, {}).get("cmd", command),
        "reason": "required command has not been attested in this evidence",
    })

report = {
    "schema_version": "1.0",
    "generated_at": datetime.now(timezone.utc).isoformat(),
    "repository": data["project"]["name"],
    "module": data["project"]["module"],
    "branch": git(["branch", "--show-current"]),
    "commit": git(["rev-parse", "HEAD"]),
    "change_base_ref": change_base_ref(),
    "diff_filter": DIFF_FILTER,
    "changed_files": files,
    "detected_change_policies": sorted(detected_policies),
    "required_commands": required_commands,
    "command_attestations": command_attestations,
    "highest_required_command_risk": highest_risk,
    "security_sensitive_paths": sorted(set(security_sensitive_paths)),
    "checks": checks,
    "worktree_status": git(["status", "--short"]),
}

output_dir = os.path.dirname(output_file)
if output_dir:
    os.makedirs(output_dir, exist_ok=True)
with open(output_file, "w", encoding="utf-8") as f:
    json.dump(report, f, indent=2, sort_keys=True)
    f.write("\n")

print(f"agent validation evidence written to {output_file}")
print("detected policies: " + (", ".join(report["detected_change_policies"]) or "none"))
print("required commands: " + (", ".join(required_commands) or "none"))
print("highest required command risk: " + highest_risk)
PY
