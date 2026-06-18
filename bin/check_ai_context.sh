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
    require_bool(spec.get("network_required"), f"commands.{name}.network_required")
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

git_hooks = require_mapping(data.get("git_hooks"), "git_hooks")
require_bool(git_hooks.get("optional"), "git_hooks.optional")
for key in ("install", "uninstall", "pre_commit", "pre_push"):
    require_string(git_hooks.get(key), f"git_hooks.{key}")

architecture_rules = require_string_list(data.get("architecture_rules"), "architecture_rules")
if len(architecture_rules) < 5:
    add_error("architecture_rules should document the enforced architecture policy")

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
