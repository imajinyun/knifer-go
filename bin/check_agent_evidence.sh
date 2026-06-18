#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
AI_CONTEXT="${ROOT_DIR}/ai-context.json"
EVIDENCE_FILE="${AGENT_EVIDENCE_FILE:-/tmp/go-knifer-agent-validation.json}"

python3 - "${ROOT_DIR}" "${AI_CONTEXT}" "${EVIDENCE_FILE}" <<'PY'
import json
import os
import sys

root_dir, ai_context, evidence_file = sys.argv[1], sys.argv[2], sys.argv[3]
errors = []


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

checks = require_mapping(evidence.get("checks"), "checks")
for check_name in ("ai_context_check", "change_policy_check", "security_sensitive_diff"):
    check = require_mapping(checks.get(check_name), f"checks.{check_name}")
    if require_string(check.get("status"), f"checks.{check_name}.status") != "passed":
        add_error(f"checks.{check_name}.status must be passed")
    if not isinstance(check.get("exit_code"), int) or isinstance(check.get("exit_code"), bool):
        add_error(f"checks.{check_name}.exit_code must be an integer")
    elif check.get("exit_code") != 0:
        add_error(f"checks.{check_name}.exit_code must be 0")
    require_string(check.get("cmd"), f"checks.{check_name}.cmd")

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

if not isinstance(evidence.get("worktree_status"), str):
    add_error("worktree_status must be a string")

if errors:
    for error in errors:
        print(f"agent evidence check error: {error}", file=sys.stderr)
    sys.exit(1)

display_path = os.path.relpath(evidence_file, root_dir) if evidence_file.startswith(root_dir + os.sep) else evidence_file
print(
    f"agent evidence is valid ({display_path}; "
    f"{len(detected_policies)} policies, {len(required_commands)} required commands)"
)
PY
