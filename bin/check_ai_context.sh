#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
AI_CONTEXT="${ROOT_DIR}/ai-context.json"

python3 - "${ROOT_DIR}" "${AI_CONTEXT}" <<'PY'
import json
import os
import re
import sys

root_dir, ai_context = sys.argv[1], sys.argv[2]
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


def require_bool(value, path):
    if not isinstance(value, bool):
        add_error(f"{path} must be a boolean")
        return False
    return value


def require_string_list(value, path):
    if not isinstance(value, list):
        add_error(f"{path} must be a list")
        return []
    for index, item in enumerate(value):
        require_string(item, f"{path}[{index}]")
    return value


def require_number(value, path):
    if not isinstance(value, (int, float)) or isinstance(value, bool):
        add_error(f"{path} must be a number")
        return 0.0
    return float(value)


def require_enum(value, path, allowed):
    value = require_string(value, path)
    if value and value not in allowed:
        add_error(f"{path} must be one of: {', '.join(sorted(allowed))}")
        return ""
    return value


try:
    with open(ai_context, "r", encoding="utf-8") as f:
        data = json.load(f)
except FileNotFoundError:
    print("missing ai-context.json", file=sys.stderr)
    sys.exit(1)
except json.JSONDecodeError as exc:
    print(f"invalid ai-context.json: {exc}", file=sys.stderr)
    sys.exit(1)

data = require_mapping(data, "ai-context.json")

require_string(data.get("schema_version"), "schema_version")

project = require_mapping(data.get("project"), "project")
for key in ("name", "module", "language", "go_version", "layout"):
    require_string(project.get(key), f"project.{key}")

commands = require_mapping(data.get("commands"), "commands")
command_name_pattern = re.compile(r"^[a-z][a-z0-9_]*$")
allowed_risk_levels = {"low", "medium", "high", "forbidden_for_agent"}
command_names = set(commands)
for name, spec in sorted(commands.items()):
    if not command_name_pattern.match(name):
        add_error(f"commands.{name} must use snake_case")
    spec = require_mapping(spec, f"commands.{name}")
    cmd = require_string(spec.get("cmd"), f"commands.{name}.cmd")
    safe_for_agent_auto_run = require_bool(
        spec.get("safe_for_agent_auto_run"),
        f"commands.{name}.safe_for_agent_auto_run",
    )
    writes_workspace = require_bool(spec.get("writes_workspace"), f"commands.{name}.writes_workspace")
    writes_git_config = require_bool(spec.get("writes_git_config"), f"commands.{name}.writes_git_config")
    network_required = require_bool(spec.get("network_required"), f"commands.{name}.network_required")
    risk_level = require_enum(spec.get("risk_level"), f"commands.{name}.risk_level", allowed_risk_levels)
    require_string(spec.get("notes"), f"commands.{name}.notes")

    requires_user_consent = spec.get("requires_user_consent", False)
    if requires_user_consent is not False:
        requires_user_consent = require_bool(requires_user_consent, f"commands.{name}.requires_user_consent")

    if not safe_for_agent_auto_run and not requires_user_consent:
        add_error(f"commands.{name} is not safe for auto-run and must require user consent")
    if safe_for_agent_auto_run and requires_user_consent:
        add_error(f"commands.{name} cannot be both auto-runnable and consent-required")
    if writes_git_config and not requires_user_consent:
        add_error(f"commands.{name} writes Git config and must require user consent")
    if writes_git_config and safe_for_agent_auto_run:
        add_error(f"commands.{name} writes Git config and must not be auto-runnable")
    if writes_workspace and safe_for_agent_auto_run:
        add_error(f"commands.{name} writes workspace files and must not be auto-runnable")
    if safe_for_agent_auto_run and risk_level in {"high", "forbidden_for_agent"}:
        add_error(f"commands.{name} is auto-runnable and must not be high or forbidden_for_agent risk")
    if requires_user_consent and risk_level == "low":
        add_error(f"commands.{name} requires user consent and must not be low risk")
    if writes_workspace and risk_level == "low":
        add_error(f"commands.{name} writes workspace files and must not be low risk")
    if writes_git_config and risk_level not in {"high", "forbidden_for_agent"}:
        add_error(f"commands.{name} writes Git config and must be high or forbidden_for_agent risk")
    if network_required and risk_level == "low":
        add_error(f"commands.{name} requires network and must be at least medium risk")
    if writes_workspace and not spec.get("writes_files"):
        add_error(f"commands.{name} writes workspace files and must declare writes_files")
    if "writes_files" in spec:
        require_string_list(spec["writes_files"], f"commands.{name}.writes_files")
    if "creates_artifacts" in spec:
        require_string_list(spec["creates_artifacts"], f"commands.{name}.creates_artifacts")
    if "reads_artifacts" in spec:
        require_string_list(spec["reads_artifacts"], f"commands.{name}.reads_artifacts")
    if cmd.startswith("make "):
        target = cmd.split()[1].split("=", 1)[0]
        if target == "check":
            target = "full-check"
        makefile = os.path.join(root_dir, "Makefile")
        with open(makefile, "r", encoding="utf-8") as f:
            makefile_text = f.read()
        if not re.search(rf"^{re.escape(target)}:(?:\s|$)", makefile_text, flags=re.MULTILINE):
            add_error(f"commands.{name}.cmd references missing Makefile target {target!r}")

change_type_policies = require_mapping(data.get("change_type_policies"), "change_type_policies")
for name, policy in sorted(change_type_policies.items()):
    if not command_name_pattern.match(name):
        add_error(f"change_type_policies.{name} must use snake_case")
    policy = require_mapping(policy, f"change_type_policies.{name}")
    required_commands = require_string_list(
        policy.get("required_commands"),
        f"change_type_policies.{name}.required_commands",
    )
    requires_user_consent_commands = require_string_list(
        policy.get("requires_user_consent_commands", []),
        f"change_type_policies.{name}.requires_user_consent_commands",
    )
    coverage_required = require_bool(
        policy.get("coverage_required"),
        f"change_type_policies.{name}.coverage_required",
    )
    security_review_required = require_bool(
        policy.get("security_review_required"),
        f"change_type_policies.{name}.security_review_required",
    )
    require_string(policy.get("notes"), f"change_type_policies.{name}.notes")

    for command_name in required_commands + requires_user_consent_commands:
        if command_name not in command_names:
            add_error(f"change_type_policies.{name} references unknown command {command_name!r}")
    for command_name in requires_user_consent_commands:
        command = commands.get(command_name, {})
        if not command.get("requires_user_consent", False):
            add_error(f"change_type_policies.{name} consent command {command_name!r} must require user consent")
    if coverage_required and "coverage_check" not in required_commands and "agent_full_check" not in required_commands:
        add_error(f"change_type_policies.{name} requires coverage but does not require coverage_check or agent_full_check")
    if security_review_required and "agent_security_check" not in required_commands and "security_check" not in required_commands:
        add_error(f"change_type_policies.{name} requires security review but does not require a security check command")
    if name == "public_api" and "api_check" not in required_commands:
        add_error("change_type_policies.public_api must require api_check")
    if name == "security_sensitive" and not security_review_required:
        add_error("change_type_policies.security_sensitive must require security review")

git_hooks = require_mapping(data.get("git_hooks"), "git_hooks")
require_bool(git_hooks.get("optional"), "git_hooks.optional")
for key in ("install", "uninstall", "pre_commit", "pre_push"):
    require_string(git_hooks.get(key), f"git_hooks.{key}")

architecture_rules = require_string_list(data.get("architecture_rules"), "architecture_rules")
if len(architecture_rules) < 5:
    add_error("architecture_rules should document the enforced architecture policy")

coverage_gates = require_mapping(data.get("coverage_gates"), "coverage_gates")
repository_threshold = require_number(coverage_gates.get("repository_threshold"), "coverage_gates.repository_threshold")
package_thresholds = require_mapping(coverage_gates.get("package_thresholds"), "coverage_gates.package_thresholds")

module_path = project.get("module")
if repository_threshold <= 0 or repository_threshold > 100:
    add_error("coverage_gates.repository_threshold must be greater than 0 and no more than 100")
for package_path, threshold in sorted(package_thresholds.items()):
    require_string(package_path, f"coverage_gates.package_thresholds.{package_path}")
    threshold = require_number(threshold, f"coverage_gates.package_thresholds.{package_path}")
    if module_path and not package_path.startswith(module_path + "/"):
        add_error(f"coverage_gates.package_thresholds.{package_path} must start with project.module")
    if threshold <= 0 or threshold > 100:
        add_error(f"coverage_gates.package_thresholds.{package_path} must be greater than 0 and no more than 100")

public_facades = data.get("public_facades")
if not isinstance(public_facades, list):
    add_error("public_facades must be a list")
    public_facades = []

declared_facades = set()
for index, entry in enumerate(public_facades):
    entry = require_mapping(entry, f"public_facades[{index}]")
    package = require_string(entry.get("package"), f"public_facades[{index}].package")
    require_string(entry.get("domain"), f"public_facades[{index}].domain")
    require_string(entry.get("internal"), f"public_facades[{index}].internal")
    if package in declared_facades:
        add_error(f"public_facades contains duplicate package {package}")
    declared_facades.add(package)

actual_facades = {
    entry
    for entry in os.listdir(root_dir)
    if entry.startswith("v")
    and os.path.isdir(os.path.join(root_dir, entry))
    and os.path.exists(os.path.join(root_dir, entry, "doc.go"))
}

missing_facades = sorted(actual_facades - declared_facades)
stale_facades = sorted(declared_facades - actual_facades)
if missing_facades:
    add_error("public_facades is missing package(s): " + ", ".join(missing_facades))
if stale_facades:
    add_error("public_facades contains stale package(s): " + ", ".join(stale_facades))

security_sensitive = set(require_string_list(data.get("security_sensitive_packages"), "security_sensitive_packages"))
unknown_security_sensitive = sorted(security_sensitive - declared_facades)
if unknown_security_sensitive:
    add_error("security_sensitive_packages contains unknown package(s): " + ", ".join(unknown_security_sensitive))

if errors:
    for error in errors:
        print(f"ai-context check error: {error}", file=sys.stderr)
    sys.exit(1)

print(f"ai-context.json is valid ({len(commands)} commands, {len(declared_facades)} public facades)")
PY
