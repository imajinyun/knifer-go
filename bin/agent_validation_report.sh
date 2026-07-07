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


def run_json(args):
    result = run(args)
    parsed = None
    parse_error = ""
    raw = result["stdout"].strip()
    if raw:
        try:
            parsed = json.loads(raw)
        except json.JSONDecodeError as exc:
            parse_error = str(exc)
    structured = {
        "status": result["status"],
        "cmd": result["cmd"],
        "exit_code": result["exit_code"],
        "stdout": result["stdout"],
        "stderr": result["stderr"],
        "json": parsed if isinstance(parsed, dict) else {},
    }
    if parse_error:
        structured["parse_error"] = parse_error
    return structured


def command_attestation(result, source):
    return {
        "status": result["status"],
        "source": source,
        "cmd": result["cmd"],
        "exit_code": result["exit_code"],
        "stdout": result.get("stdout", ""),
        "stderr": result.get("stderr", ""),
    }


def external_command_attestation(command, command_spec):
    env_name = "AGENT_ATTEST_" + command.upper()
    status = os.environ.get(env_name, "").strip()
    if not status:
        return None
    if status not in {"passed", "failed", "skipped", "covered_by_ci"}:
        return None
    attestation = {
        "status": status,
        "source": os.environ.get(env_name + "_SOURCE", "agent_run").strip() or "agent_run",
        "cmd": os.environ.get(env_name + "_CMD", command_spec.get("cmd", command)).strip() or command_spec.get("cmd", command),
    }
    exit_code = os.environ.get(env_name + "_EXIT_CODE", "").strip()
    if exit_code:
        try:
            attestation["exit_code"] = int(exit_code)
        except ValueError:
            pass
    elif status == "passed":
        attestation["exit_code"] = 0
    elif status == "failed":
        attestation["exit_code"] = 1
    if status in {"skipped", "covered_by_ci"}:
        reason = os.environ.get(env_name + "_REASON", "").strip()
        if reason:
            attestation["reason"] = reason
    if status == "covered_by_ci":
        ci_job = os.environ.get(env_name + "_CI_JOB", "").strip()
        if ci_job:
            attestation["ci_job"] = ci_job
    return attestation


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
    facade_path = next((package for package in facades if path.startswith(package + "/")), "")
    if path == "docs/api/exports.txt" or (facade_path and path.endswith(".go") and not path.endswith("_test.go")):
        detected_policies.add("public_api")
    elif facade_path and path.endswith("_test.go"):
        detected_policies.add("bug_fix")
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
structured_checks = {
    "change_policy_check": run_json(["go", "run", "./bin/changepolicycheck", "-root", root_dir, "-json"]),
    "ci_workflow_check": run_json(["go", "run", "./bin/ciworkflowcheck", "-root", root_dir, "-json"]),
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
    command_spec = data["commands"].get(command, {})
    external = external_command_attestation(command, command_spec)
    if external is not None:
        command_attestations[command] = external
        continue
    command_attestations.setdefault(command, {
        "status": "not_recorded",
        "source": "required_by_policy",
        "cmd": command_spec.get("cmd", command),
        "reason": "required command has not been attested in this evidence",
    })


def attestation_satisfies_command(command):
    if command in {"security_sensitive_diff"}:
        return True
    attestation = command_attestations.get(command, {})
    if command == "agent_evidence_check":
        return attestation.get("status") == "pending"
    return attestation.get("status") in {"passed", "covered_by_ci"}


def build_security_review():
    required = "security_sensitive" in detected_policies
    required_commands_for_review = []
    if required:
        for command in ("agent_full_check", "agent_security_check"):
            if command in required_commands:
                required_commands_for_review.append(command)
    review_attestations = {
        command: command_attestations.get(command, {
            "status": "not_recorded",
            "source": "required_by_policy",
            "cmd": data["commands"].get(command, {}).get("cmd", command),
            "reason": "required security review command has not been attested in this evidence",
        })
        for command in required_commands_for_review
    }
    ready = bool(required) and bool(security_sensitive_paths) and all(
        review_attestations.get(command, {}).get("status") in {"passed", "covered_by_ci"}
        for command in required_commands_for_review
    )
    if not required:
        status = "not_required"
        audit_conclusion = "No security-sensitive paths changed."
    elif ready:
        status = "ready"
        audit_conclusion = "Security-sensitive change has full and security validation attestations."
    else:
        status = "blocked"
        audit_conclusion = "Security-sensitive change is blocked until agent_full_check and agent_security_check are attested."
    return {
        "required": required,
        "security_review_required": required,
        "status": status,
        "paths": sorted(set(security_sensitive_paths)),
        "required_commands": required_commands_for_review,
        "command_attestations": review_attestations,
        "audit_conclusion": audit_conclusion,
    }


merge_blockers = []
for command in required_commands:
    if not attestation_satisfies_command(command):
        merge_blockers.append(command)
if checks["ai_context_check"]["status"] != "passed":
    merge_blockers.append("ai_context_check")
if checks["change_policy_check"]["status"] != "passed":
    merge_blockers.append("change_policy_check")
if "security_sensitive" in detected_policies:
    if command_attestations.get("agent_security_check", {}).get("status") not in {"passed", "covered_by_ci"}:
        merge_blockers.append("agent_security_check")
    if command_attestations.get("agent_full_check", {}).get("status") not in {"passed", "covered_by_ci"}:
        merge_blockers.append("agent_full_check")

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
    "security_review": build_security_review(),
    "checks": checks,
    "structured_checks": structured_checks,
    "merge_ready": len(merge_blockers) == 0,
    "merge_blockers": sorted(set(merge_blockers)),
    "worktree_status": git(["status", "--short"]),
}

output_dir = os.path.dirname(output_file)
if output_dir:
    os.makedirs(output_dir, exist_ok=True)
with open(output_file, "w", encoding="utf-8") as f:
    json.dump(report, f, indent=2, sort_keys=True)
    f.write("\n")


def summary_list(values):
    return ", ".join(values) if values else "none"


change_policy_json = structured_checks.get("change_policy_check", {}).get("json", {})
ci_workflow_json = structured_checks.get("ci_workflow_check", {}).get("json", {})
change_policy_rule_ids = change_policy_json.get("rule_ids", [])
change_policy_semantic_rule_ids = change_policy_json.get("semantic_rule_ids", [])
ci_workflow_findings = ci_workflow_json.get("findings", [])

print(f"agent validation evidence written to {output_file}")
print("detected policies: " + (", ".join(report["detected_change_policies"]) or "none"))
print("change policy rule ids: " + summary_list(change_policy_rule_ids))
print("change policy semantic rule ids: " + summary_list(change_policy_semantic_rule_ids))
print("ci workflow findings: " + str(len(ci_workflow_findings)))
print("required commands: " + (", ".join(required_commands) or "none"))
print("highest required command risk: " + highest_risk)
print("merge ready: " + ("true" if report["merge_ready"] else "false"))
if report["merge_blockers"]:
    print("merge blockers: " + ", ".join(report["merge_blockers"]))
PY
