#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
AI_CONTEXT="${ROOT_DIR}/ai-context.json"
EVIDENCE_FILE="${AGENT_EVIDENCE_FILE:-/tmp/knifer-go-agent-validation.json}"

python3 - "${ROOT_DIR}" "${AI_CONTEXT}" "${EVIDENCE_FILE}" <<'PY'
import json
import os
import sys

root_dir, ai_context, evidence_file = sys.argv[1], sys.argv[2], sys.argv[3]
errors = []
DIFF_FILTER = "ACDMRTUXB"


def add_error(message):
    errors.append(message)


def require_mapping(value, path):
    if not isinstance(value, dict):
        add_error(f"{path} must be an object")
        return {}
    return value


def require_string(value, path):
    if not isinstance(value, str) or not value.strip():
        add_error(f"{path} must be a non-empty string")
        return ""
    return value


def require_optional_string(value, path):
    if value is None:
        return ""
    if not isinstance(value, str):
        add_error(f"{path} must be a string")
        return ""
    return value


def require_string_list(value, path):
    if not isinstance(value, list):
        add_error(f"{path} must be a list")
        return []
    result = []
    for index, item in enumerate(value):
        item = require_string(item, f"{path}[{index}]")
        if item:
            result.append(item)
    return result


def load_json(path, label):
    try:
        with open(path, "r", encoding="utf-8") as f:
            return json.load(f)
    except FileNotFoundError:
        print(f"missing {label}: {path}", file=sys.stderr)
        sys.exit(1)
    except json.JSONDecodeError as exc:
        print(f"invalid {label}: {exc}", file=sys.stderr)
        sys.exit(1)


def is_doc_go_comment_only(path):
    if not path.endswith("/doc.go"):
        return False
    file_path = os.path.join(root_dir, path)
    try:
        with open(file_path, "r", encoding="utf-8") as f:
            lines = f.read().splitlines()
    except UnicodeDecodeError:
        return False
    in_block_comment = False
    seen_package = False
    for raw_line in lines:
        line = raw_line.strip()
        if not line:
            continue
        if in_block_comment:
            if "*/" in line:
                in_block_comment = False
                line = line.split("*/", 1)[1].strip()
                if not line:
                    continue
            else:
                continue
        if line.startswith("//"):
            continue
        if line.startswith("/*"):
            if "*/" not in line:
                in_block_comment = True
                continue
            line = line.split("*/", 1)[1].strip()
            if not line:
                continue
        if line.startswith("package "):
            seen_package = True
            continue
        return False
    return seen_package


context = require_mapping(load_json(ai_context, "ai-context.json"), "ai-context.json")
evidence = require_mapping(load_json(evidence_file, "Agent evidence"), "agent evidence")

project = require_mapping(context.get("project"), "ai-context.json.project")
commands = require_mapping(context.get("commands"), "ai-context.json.commands")
policies = require_mapping(context.get("change_type_policies"), "ai-context.json.change_type_policies")

schema_version = require_string(evidence.get("schema_version"), "schema_version")
if schema_version and schema_version != "1.0":
    add_error("schema_version must be 1.0")

if require_string(evidence.get("repository"), "repository") != project.get("name"):
    add_error("repository must match ai-context.json.project.name")
if require_string(evidence.get("module"), "module") != project.get("module"):
    add_error("module must match ai-context.json.project.module")

for key in ("generated_at", "branch", "commit"):
    require_string(evidence.get(key), key)

require_optional_string(evidence.get("change_base_ref"), "change_base_ref")
if require_string(evidence.get("diff_filter"), "diff_filter") != DIFF_FILTER:
    add_error(f"diff_filter must be {DIFF_FILTER}")

changed_files = require_string_list(evidence.get("changed_files"), "changed_files")
detected_policies = require_string_list(evidence.get("detected_change_policies"), "detected_change_policies")
required_commands = require_string_list(evidence.get("required_commands"), "required_commands")
security_sensitive_paths = require_string_list(evidence.get("security_sensitive_paths"), "security_sensitive_paths")

unknown_policies = sorted(set(detected_policies) - set(policies))
if unknown_policies:
    add_error("detected_change_policies contains unknown policies: " + ", ".join(unknown_policies))

unknown_commands = sorted(set(required_commands) - set(commands))
if unknown_commands:
    add_error("required_commands contains unknown commands: " + ", ".join(unknown_commands))

expected_required_commands = []
for policy in sorted(detected_policies):
    policy_spec = require_mapping(policies.get(policy, {}), f"ai-context.json.change_type_policies.{policy}")
    for command in policy_spec.get("required_commands", []):
        if command not in expected_required_commands:
            expected_required_commands.append(command)
if required_commands != expected_required_commands:
    add_error(
        "required_commands must match detected policies; "
        f"got {required_commands}, want {expected_required_commands}"
    )

risk_rank = {"low": 1, "medium": 2, "high": 3, "forbidden_for_agent": 4}
highest_risk = "low"
for command in required_commands:
    command_spec = require_mapping(commands.get(command, {}), f"ai-context.json.commands.{command}")
    risk = command_spec.get("risk_level")
    if risk not in risk_rank:
        add_error(f"ai-context.json.commands.{command}.risk_level is invalid")
        continue
    if risk_rank[risk] > risk_rank[highest_risk]:
        highest_risk = risk
if require_string(evidence.get("highest_required_command_risk"), "highest_required_command_risk") != highest_risk:
    add_error(f"highest_required_command_risk must be {highest_risk}")

allowed_attestation_statuses = {"passed", "failed", "pending", "not_recorded", "skipped", "covered_by_ci"}
allowed_attestation_sources = {
    "embedded_check",
    "current_process",
    "post_generation",
    "required_by_policy",
    "agent_run",
    "ci_job",
    "manual_review",
}
command_attestations = require_mapping(evidence.get("command_attestations"), "command_attestations")
for command in required_commands:
    attestation = require_mapping(command_attestations.get(command), f"command_attestations.{command}")
    status = require_string(attestation.get("status"), f"command_attestations.{command}.status")
    if status and status not in allowed_attestation_statuses:
        add_error(
            f"command_attestations.{command}.status must be one of: "
            + ", ".join(sorted(allowed_attestation_statuses))
        )
    source = require_string(attestation.get("source"), f"command_attestations.{command}.source")
    if source and source not in allowed_attestation_sources:
        add_error(
            f"command_attestations.{command}.source must be one of: "
            + ", ".join(sorted(allowed_attestation_sources))
        )
    if status in {"pending", "not_recorded", "skipped"}:
        require_string(attestation.get("reason"), f"command_attestations.{command}.reason")
    if status in {"passed", "failed"}:
        require_string(attestation.get("cmd"), f"command_attestations.{command}.cmd")
        exit_code = attestation.get("exit_code")
        if not isinstance(exit_code, int) or isinstance(exit_code, bool):
            add_error(f"command_attestations.{command}.exit_code must be an integer")
        elif status == "passed" and exit_code != 0:
            add_error(f"command_attestations.{command}.exit_code must be 0 when status is passed")
        elif status == "failed" and exit_code == 0:
            add_error(f"command_attestations.{command}.exit_code must be non-zero when status is failed")
    if status == "covered_by_ci":
        require_string(attestation.get("ci_job"), f"command_attestations.{command}.ci_job")

agent_evidence_attestation = require_mapping(
    command_attestations.get("agent_evidence"),
    "command_attestations.agent_evidence",
)
if require_string(agent_evidence_attestation.get("status"), "command_attestations.agent_evidence.status") != "passed":
    add_error("command_attestations.agent_evidence.status must be passed")
if require_string(agent_evidence_attestation.get("source"), "command_attestations.agent_evidence.source") != "current_process":
    add_error("command_attestations.agent_evidence.source must be current_process")

agent_evidence_check_attestation = require_mapping(
    command_attestations.get("agent_evidence_check"),
    "command_attestations.agent_evidence_check",
)
if require_string(agent_evidence_check_attestation.get("status"), "command_attestations.agent_evidence_check.status") != "pending":
    add_error("command_attestations.agent_evidence_check.status must be pending")
if require_string(agent_evidence_check_attestation.get("source"), "command_attestations.agent_evidence_check.source") != "post_generation":
    add_error("command_attestations.agent_evidence_check.source must be post_generation")
require_string(agent_evidence_check_attestation.get("reason"), "command_attestations.agent_evidence_check.reason")

checks = require_mapping(evidence.get("checks"), "checks")
for check_name in ("ai_context_check", "change_policy_check"):
    check = require_mapping(checks.get(check_name), f"checks.{check_name}")
    if require_string(check.get("status"), f"checks.{check_name}.status") != "passed":
        add_error(f"checks.{check_name}.status must be passed")
    if not isinstance(check.get("exit_code"), int) or isinstance(check.get("exit_code"), bool):
        add_error(f"checks.{check_name}.exit_code must be an integer")
    elif check.get("exit_code") != 0:
        add_error(f"checks.{check_name}.exit_code must be 0")
    require_string(check.get("cmd"), f"checks.{check_name}.cmd")
    attestation = require_mapping(command_attestations.get(check_name), f"command_attestations.{check_name}")
    if attestation.get("status") != check.get("status"):
        add_error(f"command_attestations.{check_name}.status must match checks.{check_name}.status")
    if attestation.get("exit_code") != check.get("exit_code"):
        add_error(f"command_attestations.{check_name}.exit_code must match checks.{check_name}.exit_code")
    if attestation.get("cmd") != check.get("cmd"):
        add_error(f"command_attestations.{check_name}.cmd must match checks.{check_name}.cmd")
    if attestation.get("source") != "embedded_check":
        add_error(f"command_attestations.{check_name}.source must be embedded_check")

if security_sensitive_paths and "security_sensitive" not in detected_policies:
    add_error("security_sensitive_paths requires detected security_sensitive policy")

facades = {
    entry["package"]: entry["internal"].rstrip("/")
    for entry in context.get("public_facades", [])
    if isinstance(entry, dict) and "package" in entry and "internal" in entry
}
security_prefixes = set()
for package in context.get("security_sensitive_packages", []):
    if isinstance(package, str):
        security_prefixes.add(package.rstrip("/") + "/")
        internal = facades.get(package)
        if internal:
            security_prefixes.add(internal.rstrip("/") + "/")
expected_security_sensitive_paths = sorted(
    path for path in changed_files if any(path.startswith(prefix) for prefix in security_prefixes)
)
if sorted(security_sensitive_paths) != expected_security_sensitive_paths:
    add_error(
        "security_sensitive_paths must match changed security-sensitive paths; "
        f"got {sorted(security_sensitive_paths)}, want {expected_security_sensitive_paths}"
    )

security_review = require_mapping(evidence.get("security_review"), "security_review")
security_review_required = security_review.get("security_review_required")
if not isinstance(security_review_required, bool):
    add_error("security_review.security_review_required must be a boolean")
review_required = security_review.get("required")
if not isinstance(review_required, bool):
    add_error("security_review.required must be a boolean")
expected_review_required = "security_sensitive" in detected_policies
if isinstance(review_required, bool) and review_required != expected_review_required:
    add_error(f"security_review.required must be {str(expected_review_required).lower()}")
if isinstance(security_review_required, bool) and security_review_required != expected_review_required:
    add_error(f"security_review.security_review_required must be {str(expected_review_required).lower()}")
if expected_review_required and not policies.get("security_sensitive", {}).get("security_review_required"):
    add_error("security_sensitive policy must require security review when security_review is required")

review_paths = require_string_list(security_review.get("paths"), "security_review.paths")
if sorted(review_paths) != expected_security_sensitive_paths:
    add_error(
        "security_review.paths must match changed security-sensitive paths; "
        f"got {sorted(review_paths)}, want {expected_security_sensitive_paths}"
    )

expected_review_commands = []
if expected_review_required:
    for command in ("agent_full_check", "agent_security_check"):
        if command in required_commands:
            expected_review_commands.append(command)
review_commands = require_string_list(security_review.get("required_commands"), "security_review.required_commands")
if review_commands != expected_review_commands:
    add_error(
        "security_review.required_commands must match security validation commands; "
        f"got {review_commands}, want {expected_review_commands}"
    )

review_attestations = require_mapping(security_review.get("command_attestations"), "security_review.command_attestations")
for command in review_commands:
    review_attestation = require_mapping(
        review_attestations.get(command),
        f"security_review.command_attestations.{command}",
    )
    top_attestation = require_mapping(command_attestations.get(command), f"command_attestations.{command}")
    for key in ("status", "source", "cmd"):
        if review_attestation.get(key) != top_attestation.get(key):
            add_error(f"security_review.command_attestations.{command}.{key} must match command_attestations.{command}.{key}")
    if "exit_code" in top_attestation and review_attestation.get("exit_code") != top_attestation.get("exit_code"):
        add_error(f"security_review.command_attestations.{command}.exit_code must match command_attestations.{command}.exit_code")
    if top_attestation.get("status") in {"skipped", "not_recorded"}:
        require_string(review_attestation.get("reason"), f"security_review.command_attestations.{command}.reason")
    if top_attestation.get("status") == "covered_by_ci":
        require_string(review_attestation.get("ci_job"), f"security_review.command_attestations.{command}.ci_job")

review_status = require_string(security_review.get("status"), "security_review.status")
if review_status and review_status not in {"not_required", "blocked", "ready"}:
    add_error("security_review.status must be one of: blocked, not_required, ready")
review_ready = expected_review_required and bool(expected_security_sensitive_paths) and all(
    command_attestations.get(command, {}).get("status") in {"passed", "covered_by_ci"}
    for command in expected_review_commands
)
expected_review_status = "not_required"
if expected_review_required:
    expected_review_status = "ready" if review_ready else "blocked"
if review_status and review_status != expected_review_status:
    add_error(f"security_review.status must be {expected_review_status}")
audit_conclusion = require_string(security_review.get("audit_conclusion"), "security_review.audit_conclusion")
if audit_conclusion:
    if expected_review_status == "ready" and "validation attestations" not in audit_conclusion:
        add_error("security_review.audit_conclusion must describe validation attestations when ready")
    if expected_review_status == "blocked" and "blocked" not in audit_conclusion.lower():
        add_error("security_review.audit_conclusion must explain the blocked security review")

security_sensitive_check = require_mapping(checks.get("security_sensitive_diff"), "checks.security_sensitive_diff")
security_sensitive_attestation = require_mapping(
    command_attestations.get("security_sensitive_diff"),
    "command_attestations.security_sensitive_diff",
)
security_sensitive_status = require_string(
    security_sensitive_check.get("status"),
    "checks.security_sensitive_diff.status",
)
security_sensitive_exit_code = security_sensitive_check.get("exit_code")
if not isinstance(security_sensitive_exit_code, int) or isinstance(security_sensitive_exit_code, bool):
    add_error("checks.security_sensitive_diff.exit_code must be an integer")
require_string(security_sensitive_check.get("cmd"), "checks.security_sensitive_diff.cmd")
if security_sensitive_attestation.get("status") != security_sensitive_status:
    add_error("command_attestations.security_sensitive_diff.status must match checks.security_sensitive_diff.status")
if security_sensitive_attestation.get("exit_code") != security_sensitive_exit_code:
    add_error("command_attestations.security_sensitive_diff.exit_code must match checks.security_sensitive_diff.exit_code")
if security_sensitive_attestation.get("cmd") != security_sensitive_check.get("cmd"):
    add_error("command_attestations.security_sensitive_diff.cmd must match checks.security_sensitive_diff.cmd")
if security_sensitive_attestation.get("source") != "embedded_check":
    add_error("command_attestations.security_sensitive_diff.source must be embedded_check")
security_sensitive_stdout = security_sensitive_check.get("stdout", "")
security_sensitive_stderr = security_sensitive_check.get("stderr", "")
if not expected_security_sensitive_paths:
    if security_sensitive_status != "passed":
        add_error("checks.security_sensitive_diff.status must be passed when no security-sensitive paths changed")
    if isinstance(security_sensitive_exit_code, int) and security_sensitive_exit_code != 0:
        add_error("checks.security_sensitive_diff.exit_code must be 0 when no security-sensitive paths changed")
else:
    combined_security_output = "\n".join(
        value for value in (security_sensitive_stdout, security_sensitive_stderr) if isinstance(value, str)
    )
    if "no changed files" in combined_security_output:
        add_error("checks.security_sensitive_diff output conflicts with security_sensitive_paths")
    for path in expected_security_sensitive_paths:
        if path not in combined_security_output:
            add_error(f"checks.security_sensitive_diff output must mention changed security-sensitive path {path!r}")
    documentation_only_security_diff = all(
        path.endswith("/example_test.go") or is_doc_go_comment_only(path)
        for path in expected_security_sensitive_paths
    )
    if documentation_only_security_diff:
        if security_sensitive_status != "passed":
            add_error("checks.security_sensitive_diff.status may be passed for security-sensitive example/doc-only diffs")
        if isinstance(security_sensitive_exit_code, int) and security_sensitive_exit_code != 0:
            add_error("checks.security_sensitive_diff.exit_code must be 0 for security-sensitive example/doc-only diffs")
        if "example/doc-only diff" not in combined_security_output:
            add_error("checks.security_sensitive_diff output must explain security-sensitive example/doc-only diff")
    else:
        if security_sensitive_status != "failed":
            add_error("checks.security_sensitive_diff.status must be failed when security-sensitive non-example paths changed")
        if isinstance(security_sensitive_exit_code, int) and security_sensitive_exit_code == 0:
            add_error("checks.security_sensitive_diff.exit_code must be non-zero when security-sensitive non-example paths changed")

def attestation_ready(command: str) -> bool:
    if command == "security_sensitive_diff":
        return True
    attestation = command_attestations.get(command, {})
    if command == "agent_evidence_check":
        return attestation.get("status") == "pending"
    return attestation.get("status") in {"passed", "covered_by_ci"}


expected_merge_blockers = []
for command in required_commands:
    if not attestation_ready(command):
        expected_merge_blockers.append(command)
if require_mapping(checks.get("ai_context_check"), "checks.ai_context_check").get("status") != "passed":
    expected_merge_blockers.append("ai_context_check")
if require_mapping(checks.get("change_policy_check"), "checks.change_policy_check").get("status") != "passed":
    expected_merge_blockers.append("change_policy_check")
if "security_sensitive" in detected_policies:
    if command_attestations.get("agent_security_check", {}).get("status") not in {"passed", "covered_by_ci"}:
        expected_merge_blockers.append("agent_security_check")
    if command_attestations.get("agent_full_check", {}).get("status") not in {"passed", "covered_by_ci"}:
        expected_merge_blockers.append("agent_full_check")
expected_merge_blockers = sorted(set(expected_merge_blockers))

merge_ready = evidence.get("merge_ready")
if not isinstance(merge_ready, bool):
    add_error("merge_ready must be a boolean")
elif merge_ready != (len(expected_merge_blockers) == 0):
    add_error(f"merge_ready must be {str(len(expected_merge_blockers) == 0).lower()}")

merge_blockers = require_string_list(evidence.get("merge_blockers"), "merge_blockers")
if sorted(merge_blockers) != expected_merge_blockers:
    add_error(f"merge_blockers must be {expected_merge_blockers}, got {sorted(merge_blockers)}")

if not isinstance(evidence.get("worktree_status"), str):
    add_error("worktree_status must be a string")

if errors:
    for error in errors:
        print(f"agent evidence check error: {error}", file=sys.stderr)
    sys.exit(1)

display_path = os.path.relpath(evidence_file, root_dir) if evidence_file.startswith(root_dir + os.sep) else evidence_file
print(
    f"agent evidence is valid ({display_path}; "
    f"{len(detected_policies)} policies, {len(required_commands)} required commands, "
    f"merge_ready={str(evidence.get('merge_ready')).lower()})"
)
PY
