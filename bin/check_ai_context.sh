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


def require_non_negative_integer(value, path):
    if not isinstance(value, int) or isinstance(value, bool):
        add_error(f"{path} must be a non-negative integer")
        return 0
    if value < 0:
        add_error(f"{path} must be a non-negative integer")
        return 0
    return value


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

go_mod_path = os.path.join(root_dir, "go.mod")
go_mod_text = ""
if not os.path.exists(go_mod_path):
    add_error("go.mod is missing")
else:
    with open(go_mod_path, "r", encoding="utf-8") as f:
        go_mod_text = f.read()
    match = re.search(r"^go\s+(\d+\.\d+)(?:\.\d+)?\s*$", go_mod_text, flags=re.MULTILINE)
    if not match:
        add_error("go.mod must declare a Go language version")
    else:
        module_go_minor = match.group(1)
        expected_project_go = f">={module_go_minor}"
        if project.get("go_version") != expected_project_go:
            add_error(
                f"project.go_version must be {expected_project_go!r} to match go.mod language version {module_go_minor!r}"
            )

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
        if go_1_25_patch and f'go-version: ["{go_1_25_patch}", "1.26"]' not in workflow_text:
            add_error(
                f"ci_workflows.github_actions.{name} test matrix must include minimum patch {go_1_25_patch!r} and next Go minor '1.26'"
            )
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

ai_tooling = require_mapping(data.get("ai_tooling"), "ai_tooling")
api_catalog = require_mapping(ai_tooling.get("api_catalog"), "ai_tooling.api_catalog")
api_catalog_path = require_string(api_catalog.get("path"), "ai_tooling.api_catalog.path")
api_catalog_schema = require_string(api_catalog.get("schema"), "ai_tooling.api_catalog.schema")
recommended_profile_schema = require_string(
    api_catalog.get("recommended_profile_schema"),
    "ai_tooling.api_catalog.recommended_profile_schema",
)
api_catalog_regenerate_command = require_string(
    api_catalog.get("regenerate_command"),
    "ai_tooling.api_catalog.regenerate_command",
)
api_catalog_check_command = require_string(api_catalog.get("check_command"), "ai_tooling.api_catalog.check_command")
for command_name, path in (
    (api_catalog_regenerate_command, "ai_tooling.api_catalog.regenerate_command"),
    (api_catalog_check_command, "ai_tooling.api_catalog.check_command"),
):
    if command_name and command_name not in command_names:
        add_error(f"{path} references unknown command {command_name!r}")
if api_catalog_regenerate_command:
    generator = commands.get(api_catalog_regenerate_command, {})
    if generator and not generator.get("requires_user_consent", False):
        add_error("ai_tooling.api_catalog.regenerate_command must require user consent")
if api_catalog_check_command:
    checker = commands.get(api_catalog_check_command, {})
    if checker and checker.get("writes_workspace", False):
        add_error("ai_tooling.api_catalog.check_command must not write workspace files")

tools_catalog_data = None
if api_catalog_path:
    absolute_api_catalog_path = os.path.join(root_dir, api_catalog_path)
    if not os.path.exists(absolute_api_catalog_path):
        add_error(f"ai_tooling.api_catalog.path references missing file {api_catalog_path!r}")
    else:
        try:
            with open(absolute_api_catalog_path, "r", encoding="utf-8") as f:
                tools_catalog_data = json.load(f)
        except json.JSONDecodeError as exc:
            add_error(f"invalid tools catalog {api_catalog_path!r}: {exc}")

declared_api_metrics = {
    "package_count": require_non_negative_integer(
        api_catalog.get("package_count"),
        "ai_tooling.api_catalog.package_count",
    ),
    "function_count": require_non_negative_integer(
        api_catalog.get("function_count"),
        "ai_tooling.api_catalog.function_count",
    ),
    "functions_with_examples": require_non_negative_integer(
        api_catalog.get("functions_with_examples"),
        "ai_tooling.api_catalog.functions_with_examples",
    ),
    "context_aware_functions": require_non_negative_integer(
        api_catalog.get("context_aware_functions"),
        "ai_tooling.api_catalog.context_aware_functions",
    ),
    "returns_error_functions": require_non_negative_integer(
        api_catalog.get("returns_error_functions"),
        "ai_tooling.api_catalog.returns_error_functions",
    ),
}
status_counts = require_mapping(api_catalog.get("status_counts"), "ai_tooling.api_catalog.status_counts")
synopsis_sources = require_mapping(api_catalog.get("synopsis_sources"), "ai_tooling.api_catalog.synopsis_sources")
for key, value in sorted(status_counts.items()):
    require_non_negative_integer(value, f"ai_tooling.api_catalog.status_counts.{key}")
for key, value in sorted(synopsis_sources.items()):
    require_non_negative_integer(value, f"ai_tooling.api_catalog.synopsis_sources.{key}")

if tools_catalog_data:
    tools_summary = require_mapping(tools_catalog_data.get("summary"), f"{api_catalog_path}.summary")
    if api_catalog_schema and tools_catalog_data.get("schema") != api_catalog_schema:
        add_error(
            "ai_tooling.api_catalog.schema must match "
            f"{api_catalog_path}.schema {tools_catalog_data.get('schema')!r}"
        )
    for key, expected in declared_api_metrics.items():
        actual = tools_summary.get(key)
        if actual != expected:
            add_error(f"ai_tooling.api_catalog.{key} must match {api_catalog_path}.summary.{key} ({actual!r})")
    for key, expected in sorted(status_counts.items()):
        actual = tools_summary.get("status_counts", {}).get(key)
        if actual != expected:
            add_error(
                f"ai_tooling.api_catalog.status_counts.{key} must match "
                f"{api_catalog_path}.summary.status_counts.{key} ({actual!r})"
            )
    for key, expected in sorted(synopsis_sources.items()):
        actual = tools_summary.get("synopsis_sources", {}).get(key)
        if actual != expected:
            add_error(
                f"ai_tooling.api_catalog.synopsis_sources.{key} must match "
                f"{api_catalog_path}.summary.synopsis_sources.{key} ({actual!r})"
            )
    allowed_recommended_profiles = {"day-one", "safe", "error", "options", "compatibility"}
    if recommended_profile_schema:
        for profile in sorted(allowed_recommended_profiles):
            if profile not in recommended_profile_schema:
                add_error(f"ai_tooling.api_catalog.recommended_profile_schema must mention {profile!r}")
        for term in ("golden_path", "use_when", "avoid_when"):
            if term not in recommended_profile_schema:
                add_error(f"ai_tooling.api_catalog.recommended_profile_schema must mention {term!r}")
    for package_index, package in enumerate(tools_catalog_data.get("packages", [])):
        package = require_mapping(package, f"{api_catalog_path}.packages[{package_index}]")
        package_name = require_string(package.get("name"), f"{api_catalog_path}.packages[{package_index}].name")
        functions = require_mapping(
            {fn.get("name"): fn for fn in package.get("functions", []) if isinstance(fn, dict)},
            f"{api_catalog_path}.packages[{package_index}].functions_by_name",
        )
        golden_path = package.get("golden_path")
        if not isinstance(golden_path, list) or not golden_path:
            add_error(f"{api_catalog_path}.packages[{package_index}] {package_name!r} must declare golden_path")
        elif len(golden_path) > 7:
            add_error(f"{api_catalog_path}.packages[{package_index}] {package_name!r} golden_path must contain at most 7 APIs")
        else:
            seen_golden_names = set()
            for golden_index, golden_entrypoint in enumerate(golden_path):
                golden_entrypoint = require_mapping(
                    golden_entrypoint,
                    f"{api_catalog_path}.packages[{package_index}].golden_path[{golden_index}]",
                )
                golden_name = require_string(
                    golden_entrypoint.get("name"),
                    f"{api_catalog_path}.packages[{package_index}].golden_path[{golden_index}].name",
                )
                require_string(
                    golden_entrypoint.get("use_when"),
                    f"{api_catalog_path}.packages[{package_index}].golden_path[{golden_index}].use_when",
                )
                require_string(
                    golden_entrypoint.get("avoid_when"),
                    f"{api_catalog_path}.packages[{package_index}].golden_path[{golden_index}].avoid_when",
                )
                if golden_name in seen_golden_names:
                    add_error(f"{api_catalog_path}.{package_name}.golden_path repeats function {golden_name!r}")
                seen_golden_names.add(golden_name)
                if golden_name not in functions:
                    add_error(f"{api_catalog_path}.{package_name}.golden_path contains unknown function {golden_name!r}")
        entrypoints = package.get("recommended_entrypoints")
        if not isinstance(entrypoints, list) or not entrypoints:
            add_error(f"{api_catalog_path}.packages[{package_index}] {package_name!r} must declare recommended_entrypoints")
            continue
        seen_profiles = set()
        for entrypoint_index, entrypoint in enumerate(entrypoints):
            entrypoint = require_mapping(
                entrypoint,
                f"{api_catalog_path}.packages[{package_index}].recommended_entrypoints[{entrypoint_index}]",
            )
            name = require_string(
                entrypoint.get("name"),
                f"{api_catalog_path}.packages[{package_index}].recommended_entrypoints[{entrypoint_index}].name",
            )
            profile = require_string(
                entrypoint.get("profile"),
                f"{api_catalog_path}.packages[{package_index}].recommended_entrypoints[{entrypoint_index}].profile",
            )
            require_string(
                entrypoint.get("rationale"),
                f"{api_catalog_path}.packages[{package_index}].recommended_entrypoints[{entrypoint_index}].rationale",
            )
            if name not in functions:
                add_error(f"{api_catalog_path}.{package_name}.recommended_entrypoints contains unknown function {name!r}")
            if profile not in allowed_recommended_profiles:
                add_error(f"{api_catalog_path}.{package_name}.{name} has unknown recommended profile {profile!r}")
            if profile in seen_profiles:
                add_error(f"{api_catalog_path}.{package_name} repeats recommended profile {profile!r}")
            seen_profiles.add(profile)

human_catalog = require_mapping(ai_tooling.get("human_catalog"), "ai_tooling.human_catalog")
human_catalog_path = require_string(human_catalog.get("path"), "ai_tooling.human_catalog.path")
human_catalog_check_command = require_string(human_catalog.get("check_command"), "ai_tooling.human_catalog.check_command")
if human_catalog_path and not os.path.exists(os.path.join(root_dir, human_catalog_path)):
    add_error(f"ai_tooling.human_catalog.path references missing file {human_catalog_path!r}")
if human_catalog_check_command and human_catalog_check_command not in command_names:
    add_error(f"ai_tooling.human_catalog.check_command references unknown command {human_catalog_check_command!r}")

agent_import_rules = require_string_list(ai_tooling.get("agent_import_rules"), "ai_tooling.agent_import_rules")
selection_rules = require_string_list(ai_tooling.get("selection_rules"), "ai_tooling.selection_rules")
metadata_refresh_triggers = require_string_list(
    ai_tooling.get("metadata_refresh_triggers"),
    "ai_tooling.metadata_refresh_triggers",
)
api_decision_card_template = require_string_list(
    ai_tooling.get("api_decision_card_template"),
    "ai_tooling.api_decision_card_template",
)
if len(agent_import_rules) < 3:
    add_error("ai_tooling.agent_import_rules should include import boundary, selection, and safety guidance")
if len(selection_rules) < 5:
    add_error("ai_tooling.selection_rules should document common package-routing decisions")
metadata_refresh_text = " ".join(metadata_refresh_triggers)
if "tools_update" not in metadata_refresh_text and "make tools-gen" not in metadata_refresh_text:
    add_error("ai_tooling.metadata_refresh_triggers must mention tools_update or make tools-gen")
if "ai-context-check" not in metadata_refresh_text:
    add_error("ai_tooling.metadata_refresh_triggers must mention make ai-context-check")
decision_card_text = " ".join(api_decision_card_template).lower()
for required_term in ("problem", "package", "proposed api", "alternatives", "safety", "validation"):
    if required_term not in decision_card_text:
        add_error(f"ai_tooling.api_decision_card_template must mention {required_term!r}")

top_entrypoints = ai_tooling.get("top_entrypoints")
if not isinstance(top_entrypoints, list):
    add_error("ai_tooling.top_entrypoints must be a list")
    top_entrypoints = []
if len(top_entrypoints) < 5:
    add_error("ai_tooling.top_entrypoints should cover the main package-selection intents")
for index, entrypoint in enumerate(top_entrypoints):
    entrypoint = require_mapping(entrypoint, f"ai_tooling.top_entrypoints[{index}]")
    require_string(entrypoint.get("intent"), f"ai_tooling.top_entrypoints[{index}].intent")
    require_string_list(entrypoint.get("packages"), f"ai_tooling.top_entrypoints[{index}].packages")

stdlib_first_decisions = ai_tooling.get("stdlib_first_decisions")
if not isinstance(stdlib_first_decisions, list):
    add_error("ai_tooling.stdlib_first_decisions must be a list")
    stdlib_first_decisions = []
if len(stdlib_first_decisions) < 5:
    add_error("ai_tooling.stdlib_first_decisions should cover core package-vs-stdlib decisions")
for index, decision in enumerate(stdlib_first_decisions):
    decision = require_mapping(decision, f"ai_tooling.stdlib_first_decisions[{index}]")
    require_string(decision.get("scenario"), f"ai_tooling.stdlib_first_decisions[{index}].scenario")
    prefer_stdlib_when = require_string(
        decision.get("prefer_stdlib_when"),
        f"ai_tooling.stdlib_first_decisions[{index}].prefer_stdlib_when",
    )
    prefer_go_knifer_when = require_string(
        decision.get("prefer_go_knifer_when"),
        f"ai_tooling.stdlib_first_decisions[{index}].prefer_go_knifer_when",
    )
    packages = require_string_list(
        decision.get("packages"),
        f"ai_tooling.stdlib_first_decisions[{index}].packages",
    )
    if "stdlib" not in prefer_stdlib_when.lower() and "standard" not in prefer_stdlib_when.lower():
        add_error(f"ai_tooling.stdlib_first_decisions[{index}].prefer_stdlib_when must mention stdlib or standard library")
    if "go_knifer" not in prefer_go_knifer_when.lower() and "knifer-go" not in prefer_go_knifer_when.lower():
        add_error(f"ai_tooling.stdlib_first_decisions[{index}].prefer_go_knifer_when must mention knifer-go")

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

dependency_tiers = require_mapping(data.get("dependency_tiers"), "dependency_tiers")
dependency_tier_names = ("core_facades", "heavy_extension_facades", "provider_contract_facades")
declared_tier_packages = set()
for tier_name in dependency_tier_names:
    packages = require_string_list(dependency_tiers.get(tier_name), f"dependency_tiers.{tier_name}")
    duplicate_packages = sorted({package for package in packages if packages.count(package) > 1})
    if duplicate_packages:
        add_error(f"dependency_tiers.{tier_name} contains duplicate package(s): " + ", ".join(duplicate_packages))
    for package in packages:
        if package in declared_tier_packages:
            add_error(f"dependency_tiers contains package {package!r} in more than one tier")
        declared_tier_packages.add(package)
require_string_list(dependency_tiers.get("notes"), "dependency_tiers.notes")

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

unknown_tier_packages = sorted(declared_tier_packages - declared_facades)
missing_tier_packages = sorted(declared_facades - declared_tier_packages)
if unknown_tier_packages:
    add_error("dependency_tiers contains unknown package(s): " + ", ".join(unknown_tier_packages))
if missing_tier_packages:
    add_error("dependency_tiers is missing package(s): " + ", ".join(missing_tier_packages))

for index, entrypoint in enumerate(top_entrypoints):
    if not isinstance(entrypoint, dict):
        continue
    for package in entrypoint.get("packages", []):
        if package not in declared_facades:
            add_error(f"ai_tooling.top_entrypoints[{index}].packages contains unknown package {package!r}")

for index, decision in enumerate(stdlib_first_decisions):
    if not isinstance(decision, dict):
        continue
    for package in decision.get("packages", []):
        if package not in declared_facades:
            add_error(f"ai_tooling.stdlib_first_decisions[{index}].packages contains unknown package {package!r}")

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

v1_readiness = require_mapping(data.get("v1_readiness"), "v1_readiness")
v1_checklist = require_string_list(v1_readiness.get("checklist"), "v1_readiness.checklist")
v1_check_commands = require_string_list(v1_readiness.get("check_commands"), "v1_readiness.check_commands")
v1_blocking_exit_criteria = require_string_list(
    v1_readiness.get("blocking_exit_criteria"),
    "v1_readiness.blocking_exit_criteria",
)
if len(v1_checklist) < 5:
    add_error("v1_readiness.checklist must include at least five readiness items")
if len(v1_blocking_exit_criteria) < 3:
    add_error("v1_readiness.blocking_exit_criteria must include at least three blocking criteria")
for command_name in v1_check_commands:
    if command_name not in command_names:
        add_error(f"v1_readiness.check_commands references unknown command {command_name!r}")

api_freeze = require_mapping(data.get("api_freeze"), "api_freeze")
allowed_statuses = set(require_string_list(api_freeze.get("allowed_statuses"), "api_freeze.allowed_statuses"))
if allowed_statuses != {"recommended", "compatibility", "experimental", "deprecated"}:
    add_error("api_freeze.allowed_statuses must contain recommended, compatibility, experimental, deprecated")
if api_freeze.get("decision_card_required") is not True:
    add_error("api_freeze.decision_card_required must be true")
if api_freeze.get("replacement_required_for_deprecation") is not True:
    add_error("api_freeze.replacement_required_for_deprecation must be true")
freeze_checks = require_string_list(api_freeze.get("freeze_checks"), "api_freeze.freeze_checks")
freeze_checks_text = " ".join(freeze_checks).lower()
for required_term in ("decision card", "replacement", "snapshot", "tools catalog"):
    if required_term not in freeze_checks_text:
        add_error(f"api_freeze.freeze_checks must mention {required_term!r}")
deprecations = api_freeze.get("deprecations")
if not isinstance(deprecations, list):
    add_error("api_freeze.deprecations must be a list")
    deprecations = []
for index, deprecation in enumerate(deprecations):
    deprecation = require_mapping(deprecation, f"api_freeze.deprecations[{index}]")
    require_string(deprecation.get("name"), f"api_freeze.deprecations[{index}].name")
    require_string(deprecation.get("replacement"), f"api_freeze.deprecations[{index}].replacement")
    require_string(deprecation.get("rationale"), f"api_freeze.deprecations[{index}].rationale")
must_api_inventory = api_freeze.get("must_api_inventory")
if not isinstance(must_api_inventory, list):
    add_error("api_freeze.must_api_inventory must be a list")
    must_api_inventory = []
for index, entry in enumerate(must_api_inventory):
    entry = require_mapping(entry, f"api_freeze.must_api_inventory[{index}]")
    require_string(entry.get("name"), f"api_freeze.must_api_inventory[{index}].name")
    require_string(entry.get("replacement"), f"api_freeze.must_api_inventory[{index}].replacement")
    require_string(entry.get("rationale"), f"api_freeze.must_api_inventory[{index}].rationale")
    require_string(entry.get("doc_path"), f"api_freeze.must_api_inventory[{index}].doc_path")
decision_cards = api_freeze.get("decision_cards")
if not isinstance(decision_cards, list):
    add_error("api_freeze.decision_cards must be a list")
    decision_cards = []
decision_card_statuses = {}
for index, card in enumerate(decision_cards):
    card = require_mapping(card, f"api_freeze.decision_cards[{index}]")
    card_id = require_string(card.get("id"), f"api_freeze.decision_cards[{index}].id")
    card_status = require_string(card.get("status"), f"api_freeze.decision_cards[{index}].status")
    if card_id:
        decision_card_statuses[card_id] = card_status
api_status_decision_cards = api_freeze.get("api_status_decision_cards")
if not isinstance(api_status_decision_cards, dict):
    add_error("api_freeze.api_status_decision_cards must be an object")
    api_status_decision_cards = {}
if set(api_status_decision_cards) != allowed_statuses:
    add_error("api_freeze.api_status_decision_cards must map every allowed API status")
for status, card_ids in api_status_decision_cards.items():
    card_ids = require_string_list(card_ids, f"api_freeze.api_status_decision_cards.{status}")
    if not card_ids:
        add_error(f"api_freeze.api_status_decision_cards.{status} must not be empty")
    for card_id in card_ids:
        card_status = decision_card_statuses.get(card_id)
        if card_status is None:
            add_error(f"api_freeze.api_status_decision_cards.{status} references unknown decision card {card_id!r}")
        elif card_status != status:
            add_error(f"api_freeze.api_status_decision_cards.{status} references {card_id!r} with status {card_status!r}")

if errors:
    for error in errors:
        print(f"ai-context check error: {error}", file=sys.stderr)
    sys.exit(1)

print(f"ai-context.json is valid ({len(commands)} commands, {len(declared_facades)} public facades)")
PY
