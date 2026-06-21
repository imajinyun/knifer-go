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


def validate_json_schema(schema, instance, path="ai-context.json"):
    schema_type = schema.get("type")
    if schema_type == "object":
        if not isinstance(instance, dict):
            add_error(f"{path} must be an object")
            return
        for key in schema.get("required", []):
            if key not in instance:
                add_error(f"{path}.{key} is required by ai-context.schema.json")
        properties = schema.get("properties", {})
        additional = schema.get("additionalProperties", True)
        for key, value in instance.items():
            child_path = f"{path}.{key}"
            if key in properties:
                validate_json_schema(resolve_schema(properties[key]), value, child_path)
            elif isinstance(additional, dict):
                validate_json_schema(resolve_schema(additional), value, child_path)
            elif additional is False:
                add_error(f"{child_path} is not allowed by ai-context.schema.json")
    elif schema_type == "array":
        if not isinstance(instance, list):
            add_error(f"{path} must be a list")
            return
        minimum = schema.get("minItems")
        if minimum is not None and len(instance) < minimum:
            add_error(f"{path} must contain at least {minimum} item(s)")
        if schema.get("uniqueItems") and len(instance) != len(set(instance)):
            add_error(f"{path} must contain unique items")
        items_schema = resolve_schema(schema.get("items", {}))
        for index, item in enumerate(instance):
            validate_json_schema(items_schema, item, f"{path}[{index}]")
    elif schema_type == "string":
        if not isinstance(instance, str):
            add_error(f"{path} must be a string")
        elif schema.get("minLength", 0) and len(instance) < schema["minLength"]:
            add_error(f"{path} must be a non-empty string")
    elif schema_type == "boolean":
        if not isinstance(instance, bool):
            add_error(f"{path} must be a boolean")
    elif schema_type == "number":
        if not isinstance(instance, (int, float)) or isinstance(instance, bool):
            add_error(f"{path} must be a number")
            return
        if "exclusiveMinimum" in schema and not instance > schema["exclusiveMinimum"]:
            add_error(f"{path} must be greater than {schema['exclusiveMinimum']}")
        if "maximum" in schema and not instance <= schema["maximum"]:
            add_error(f"{path} must be no more than {schema['maximum']}")
    if "enum" in schema and instance not in schema["enum"]:
        add_error(f"{path} must be one of: {', '.join(schema['enum'])}")


schema_data = None


def resolve_schema(schema):
    ref = schema.get("$ref") if isinstance(schema, dict) else None
    if not ref:
        return schema
    if not ref.startswith("#/$defs/"):
        add_error(f"unsupported schema reference {ref!r}")
        return {}
    return schema_data.get("$defs", {}).get(ref.removeprefix("#/$defs/"), {})


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

schema_ref = require_string(data.get("$schema"), "$schema")
if schema_ref:
    schema_path = os.path.normpath(os.path.join(root_dir, schema_ref))
    if not os.path.exists(schema_path):
        add_error(f"$schema references missing file {schema_ref!r}")
    else:
        try:
            with open(schema_path, "r", encoding="utf-8") as f:
                schema_data = json.load(f)
            validate_json_schema(schema_data, data)
        except json.JSONDecodeError as exc:
            add_error(f"invalid schema file {schema_ref!r}: {exc}")

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
    if "change_policy_check" not in required_commands:
        add_error(f"change_type_policies.{name} must require change_policy_check")
    if "agent_evidence" not in required_commands:
        add_error(f"change_type_policies.{name} must require agent_evidence")
    if "agent_evidence_check" not in required_commands:
        add_error(f"change_type_policies.{name} must require agent_evidence_check")

git_hooks = require_mapping(data.get("git_hooks"), "git_hooks")
require_bool(git_hooks.get("optional"), "git_hooks.optional")
for key in ("install", "uninstall", "pre_commit", "pre_push"):
    require_string(git_hooks.get(key), f"git_hooks.{key}")

ci_workflows = require_mapping(data.get("ci_workflows"), "ci_workflows")
tool_versions = require_mapping(ci_workflows.get("tool_versions"), "ci_workflows.tool_versions")
go_1_25_patch = require_string(tool_versions.get("go_1_25_patch"), "ci_workflows.tool_versions.go_1_25_patch")
golangci_lint_version = require_string(tool_versions.get("golangci_lint"), "ci_workflows.tool_versions.golangci_lint")
github_actions = require_mapping(ci_workflows.get("github_actions"), "ci_workflows.github_actions")
for name, workflow in sorted(github_actions.items()):
    if not command_name_pattern.match(name):
        add_error(f"ci_workflows.github_actions.{name} must use snake_case")
    workflow = require_mapping(workflow, f"ci_workflows.github_actions.{name}")
    workflow_path = require_string(workflow.get("path"), f"ci_workflows.github_actions.{name}.path")
    required_jobs = require_string_list(
        workflow.get("required_jobs"),
        f"ci_workflows.github_actions.{name}.required_jobs",
    )
    workflow_text = ""
    if workflow_path:
        absolute_workflow_path = os.path.join(root_dir, workflow_path)
        if not os.path.exists(absolute_workflow_path):
            add_error(f"ci_workflows.github_actions.{name}.path references missing workflow {workflow_path!r}")
        else:
            with open(absolute_workflow_path, "r", encoding="utf-8") as f:
                workflow_text = f.read()
    for job in required_jobs:
        if workflow_text and not re.search(rf"^  {re.escape(job)}:\s*$", workflow_text, flags=re.MULTILINE):
            add_error(f"ci_workflows.github_actions.{name} is missing required job {job!r}")

    agent_governance = require_mapping(
        workflow.get("agent_governance"),
        f"ci_workflows.github_actions.{name}.agent_governance",
    )
    required_commands = require_string_list(
        agent_governance.get("required_commands"),
        f"ci_workflows.github_actions.{name}.agent_governance.required_commands",
    )
    required_env = require_string_list(
        agent_governance.get("required_env"),
        f"ci_workflows.github_actions.{name}.agent_governance.required_env",
    )
    required_artifacts = require_string_list(
        agent_governance.get("required_artifacts"),
        f"ci_workflows.github_actions.{name}.agent_governance.required_artifacts",
    )
    for required_text in required_commands + required_env + required_artifacts:
        if workflow_text and required_text not in workflow_text:
            add_error(f"ci_workflows.github_actions.{name} workflow must contain {required_text!r}")
    if workflow_text and go_1_25_patch and name in {"go", "release"}:
        if "GO_1_25_PATCH_VERSION" not in workflow_text or go_1_25_patch not in workflow_text:
            add_error(f"ci_workflows.github_actions.{name} must use declared Go patch version {go_1_25_patch!r}")
    if workflow_text and golangci_lint_version and name == "go":
        if "GOLANGCI_LINT_VERSION" not in workflow_text or golangci_lint_version not in workflow_text:
            add_error(f"ci_workflows.github_actions.{name} must use declared golangci-lint version {golangci_lint_version!r}")
    if workflow_text and name == "go":
        duplicate_steps = ["make race-test", "make shuffle-test", "make mod-check"]
        for duplicate_step in duplicate_steps:
            if duplicate_step in workflow_text:
                add_error(f"ci_workflows.github_actions.{name} should not duplicate ci-test sub-step {duplicate_step!r}")

architecture_rules = require_string_list(data.get("architecture_rules"), "architecture_rules")
if len(architecture_rules) < 5:
    add_error("architecture_rules should document the enforced architecture policy")

generated_artifacts = require_mapping(data.get("generated_artifacts"), "generated_artifacts")
for name, artifact in sorted(generated_artifacts.items()):
    if not command_name_pattern.match(name):
        add_error(f"generated_artifacts.{name} must use snake_case")
    artifact = require_mapping(artifact, f"generated_artifacts.{name}")
    path = require_string(artifact.get("path"), f"generated_artifacts.{name}.path")
    generator_command = require_string(
        artifact.get("generator_command"),
        f"generated_artifacts.{name}.generator_command",
    )
    check_command = require_string(artifact.get("check_command"), f"generated_artifacts.{name}.check_command")
    if path and not os.path.exists(os.path.join(root_dir, path)):
        add_error(f"generated_artifacts.{name}.path references missing file {path!r}")
    if generator_command and generator_command not in command_names:
        add_error(f"generated_artifacts.{name}.generator_command references unknown command {generator_command!r}")
    if check_command and check_command not in command_names:
        add_error(f"generated_artifacts.{name}.check_command references unknown command {check_command!r}")
    generator = commands.get(generator_command, {})
    if generator and not generator.get("requires_user_consent", False):
        add_error(f"generated_artifacts.{name}.generator_command must require user consent")
    if path and generator and path not in generator.get("writes_files", []):
        add_error(f"generated_artifacts.{name}.path must be declared in commands.{generator_command}.writes_files")
    checker = commands.get(check_command, {})
    if checker and checker.get("writes_workspace", False):
        add_error(f"generated_artifacts.{name}.check_command must not write workspace files")

coverage_gates = require_mapping(data.get("coverage_gates"), "coverage_gates")
repository_threshold = require_number(coverage_gates.get("repository_threshold"), "coverage_gates.repository_threshold")
security_sensitive_min_threshold = coverage_gates.get("security_sensitive_min_threshold")
if security_sensitive_min_threshold is not None:
    security_sensitive_min_threshold = require_number(
        security_sensitive_min_threshold,
        "coverage_gates.security_sensitive_min_threshold",
    )
package_thresholds = require_mapping(coverage_gates.get("package_thresholds"), "coverage_gates.package_thresholds")

module_path = project.get("module")
if repository_threshold <= 0 or repository_threshold > 100:
    add_error("coverage_gates.repository_threshold must be greater than 0 and no more than 100")
if security_sensitive_min_threshold is not None and (
    security_sensitive_min_threshold <= 0 or security_sensitive_min_threshold > 100
):
    add_error("coverage_gates.security_sensitive_min_threshold must be greater than 0 and no more than 100")
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

security_domains = require_mapping(data.get("security_domains", {}), "security_domains")
declared_security_domain_packages = set()
for name, domain in sorted(security_domains.items()):
    if not command_name_pattern.match(name):
        add_error(f"security_domains.{name} must use snake_case")
    domain = require_mapping(domain, f"security_domains.{name}")
    packages = set(require_string_list(domain.get("packages"), f"security_domains.{name}.packages"))
    threats = require_string_list(domain.get("threats"), f"security_domains.{name}.threats")
    review_focus = require_string_list(
        domain.get("required_review_focus"),
        f"security_domains.{name}.required_review_focus",
    )
    declared_security_domain_packages.update(packages)
    unknown_packages = sorted(packages - declared_facades)
    if unknown_packages:
        add_error(f"security_domains.{name}.packages contains unknown package(s): " + ", ".join(unknown_packages))
    if not threats:
        add_error(f"security_domains.{name}.threats must not be empty")
    if not review_focus:
        add_error(f"security_domains.{name}.required_review_focus must not be empty")

missing_domain_packages = sorted(security_sensitive - declared_security_domain_packages)
if security_sensitive and missing_domain_packages:
    add_error("security_domains does not classify security-sensitive package(s): " + ", ".join(missing_domain_packages))

if errors:
    for error in errors:
        print(f"ai-context check error: {error}", file=sys.stderr)
    sys.exit(1)

print(f"ai-context.json is valid ({len(commands)} commands, {len(declared_facades)} public facades)")
PY
