#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")/.."

python3 - <<'PY'
from __future__ import annotations

import json
import os
import re
import sys
from pathlib import Path
from typing import Callable

root = Path.cwd()
errors: list[str] = []


def add_error(message: str) -> None:
	errors.append(message)


def run_section(name: str, validators: list[Callable[[], None]]) -> None:
	print(f"governance maturity: running {name}")
	for validator in validators:
		validator()


def require_mapping(value: object, path: str) -> dict:
	if not isinstance(value, dict):
		add_error(f"{path} must be an object")
		return {}
	return value


def require_string_list(value: object, path: str) -> list[str]:
	if not isinstance(value, list):
		add_error(f"{path} must be a list")
		return []
	items: list[str] = []
	for index, item in enumerate(value):
		if not isinstance(item, str) or not item.strip():
			add_error(f"{path}[{index}] must be a non-empty string")
			continue
		items.append(item)
	return items


def extract_markdown_table(path: Path, heading: str) -> dict[str, int]:
	try:
		text = path.read_text(encoding="utf-8")
	except FileNotFoundError:
		add_error(f"{path.relative_to(root).as_posix()} must exist")
		return {}
	match = re.search(rf"^## {re.escape(heading)}\n(?P<body>.*?)(?=^## |\Z)", text, flags=re.MULTILINE | re.DOTALL)
	if not match:
		add_error(f"{path.relative_to(root).as_posix()} must contain ## {heading}")
		return {}
	values: dict[str, int] = {}
	for line in match.group("body").splitlines():
		if not line.startswith("|"):
			continue
		columns = [column.strip() for column in line.strip().strip("|").split("|")]
		if len(columns) != 2 or columns[0] in {"Metric", "---"} or set(columns[0]) <= {"-", ":"}:
			continue
		number = columns[1].replace(",", "")
		if not re.fullmatch(r"\d+", number):
			add_error(f"{path.relative_to(root).as_posix()} {heading} metric {columns[0]!r} must be an integer, got {columns[1]!r}")
			continue
		values[columns[0]] = int(number)
	return values


def extract_markdown_rows(path: Path, heading: str) -> list[dict[str, str]]:
	try:
		text = path.read_text(encoding="utf-8")
	except FileNotFoundError:
		add_error(f"{path.relative_to(root).as_posix()} must exist")
		return []
	match = re.search(rf"^## {re.escape(heading)}\n(?P<body>.*?)(?=^## |\Z)", text, flags=re.MULTILINE | re.DOTALL)
	if not match:
		add_error(f"{path.relative_to(root).as_posix()} must contain ## {heading}")
		return []
	table_lines = [line for line in match.group("body").splitlines() if line.startswith("|")]
	if len(table_lines) < 2:
		add_error(f"{path.relative_to(root).as_posix()} {heading} must contain a markdown table")
		return []
	headers = [column.strip() for column in table_lines[0].strip().strip("|").split("|")]
	rows: list[dict[str, str]] = []
	for line in table_lines[2:]:
		columns = [column.strip() for column in line.strip().strip("|").split("|")]
		if len(columns) != len(headers):
			add_error(f"{path.relative_to(root).as_posix()} {heading} row has {len(columns)} columns, want {len(headers)}")
			continue
		rows.append(dict(zip(headers, columns)))
	return rows


def extract_markdown_section(text: str, heading: str, level: int = 2) -> str:
	prefix = "#" * level
	next_prefix = "#" * level
	match = re.search(
		rf"^{re.escape(prefix + ' ' + heading)}\n(?P<body>.*?)(?=^{re.escape(next_prefix)} |\Z)",
		text,
		flags=re.MULTILINE | re.DOTALL,
	)
	if not match:
		return ""
	return match.group("body")


def extract_tools_markdown_golden_path(tools_md: str, package_name: str) -> list[str]:
	package_body = extract_markdown_section(tools_md, package_name, level=3)
	if not package_body:
		add_error(f"docs/api/tools.md must contain ### {package_name}")
		return []
	match = re.search(
		r"^Golden path API set:\n\n(?P<table>(?:\|.*\n)+)",
		package_body,
		flags=re.MULTILINE,
	)
	if not match:
		add_error(f"docs/api/tools.md {package_name} must contain Golden path API set table")
		return []
	rows: list[str] = []
	for line in match.group("table").splitlines()[2:]:
		columns = [column.strip() for column in line.strip().strip("|").split("|")]
		if columns and re.fullmatch(r"`[^`]+`", columns[0]):
			rows.append(columns[0].strip("`"))
	return rows


def extract_quickstart_golden_path(package_name: str) -> list[str]:
	matches = sorted((root / "docs/doc").glob(f"*-{package_name}.md"))
	if not matches:
		add_error(f"docs/doc must contain quickstart for {package_name}")
		return []
	text = matches[0].read_text(encoding="utf-8")
	body = extract_markdown_section(text, "Golden path APIs")
	if not body:
		add_error(f"{matches[0].relative_to(root).as_posix()} must contain ## Golden path APIs")
		return []
	return re.findall(r"^- `([^`]+)`\s*$", body, flags=re.MULTILINE)


def extract_backticked_facades(value: str) -> list[str]:
	return re.findall(r"`(v[A-Za-z0-9_]*)`", value)


def markdown_code_list(values: list[str]) -> str:
	return ", ".join(f"`{value}`" for value in values)


def generated_block(text: str, begin: str, end: str) -> str:
	pattern = re.compile(rf"<!-- {re.escape(begin)} -->\n(?P<body>.*?)\n<!-- {re.escape(end)} -->", flags=re.DOTALL)
	match = pattern.search(text)
	if not match:
		add_error(f"generated block {begin} / {end} must exist")
		return ""
	return match.group("body").strip()


def expected_facade_tiering_dependency_table(dependency_tiers: dict) -> str:
	core = require_string_list(dependency_tiers.get("core_facades"), "dependency_tiers.core_facades")
	heavy = require_string_list(dependency_tiers.get("heavy_extension_facades"), "dependency_tiers.heavy_extension_facades")
	providers = require_string_list(dependency_tiers.get("provider_contract_facades"), "dependency_tiers.provider_contract_facades")
	return "\n".join(
		[
			"| Tier | Facades | Import rule |",
			"| --- | --- | --- |",
			"| core facades | {facades} | Standard-library-first; third-party imports require explicit allowlist review. |".format(
				facades=markdown_code_list(core),
			),
			"| heavy extension facades | {facades} | Optional integrations stay inside their owning facade and matching `internal/*` package family. |".format(
				facades=markdown_code_list(heavy),
			),
			"| provider contract facades | {facades} | Public APIs expose provider interfaces and call contracts; concrete clients, credentials, dictionaries, and NLP engines stay outside core. |".format(
				facades=markdown_code_list(providers),
			),
		]
	)


def expected_facade_tiering_security_table(security_sensitive: list[str]) -> str:
	security_sensitive_set = set(security_sensitive)
	categories = [
		("Network and URL boundaries", ["vhttp", "vresty", "vurl", "vnet"]),
		("File, archive, and config boundaries", ["vfile", "vzip", "vconf"]),
		("Crypto, token, random, and identity boundaries", ["vcrypto", "vjwt", "vrand", "vid"]),
		("SQL and command boundaries", ["vdb", "vcli"]),
		("Provider contract boundaries", ["vai", "vftp", "vssh"]),
	]
	rows = [
		"| Category | Facades |",
		"| --- | --- |",
	]
	covered: set[str] = set()
	for category, facades in categories:
		covered.update(facades)
		filtered = [facade for facade in facades if facade in security_sensitive_set]
		if filtered:
			rows.append(f"| {category} | {markdown_code_list(filtered)} |")
	uncovered = sorted(security_sensitive_set - covered)
	if uncovered:
		add_error("facade tiering security overlay does not cover security-sensitive facade(s): " + ", ".join(uncovered))
	return "\n".join(rows)


def file_exists(reference: str) -> bool:
	path = reference.split(":", 1)[0]
	return (root / path).exists()


def reference_exists(reference: str) -> bool:
	path, separator, symbol = reference.partition(":")
	file_path = root / path
	if not file_path.exists():
		return False
	if not separator or not re.match(r"^[A-Za-z_]\w*$", symbol):
		return True
	try:
		text = file_path.read_text(encoding="utf-8")
	except UnicodeDecodeError:
		return False
	return re.search(rf"^func\s+{re.escape(symbol)}\s*\(", text, flags=re.MULTILINE) is not None


def references_function(reference: str) -> bool:
	_, separator, symbol = reference.partition(":")
	return bool(separator and re.match(r"^[A-Za-z_]\w*$", symbol))


def test_function_exists(reference: str) -> bool:
	path, separator, symbol = reference.partition(":")
	if not separator or not re.match(r"^[A-Za-z_]\w*$", symbol):
		return False
	file_path = root / path
	if not file_path.exists():
		return False
	try:
		text = file_path.read_text(encoding="utf-8")
	except UnicodeDecodeError:
		return False
	return re.search(rf"^func\s+{re.escape(symbol)}\s*\(\s*t\s+\*testing\.T\s*\)", text, flags=re.MULTILINE) is not None


def example_function_exists(package_name: str, example_name: str) -> bool:
	package_dir = root / package_name
	if not package_dir.is_dir():
		return False
	pattern = re.compile(rf"^func\s+{re.escape(example_name)}\s*\(\s*\)", flags=re.MULTILINE)
	for path in package_dir.glob("*_test.go"):
		try:
			text = path.read_text(encoding="utf-8")
		except UnicodeDecodeError:
			continue
		if pattern.search(text):
			return True
	return False


with open("ai-context.json", "r", encoding="utf-8") as f:
	ai_context = json.load(f)

with open("docs/api/tools.json", "r", encoding="utf-8") as f:
	tools = json.load(f)

with open("Makefile", "r", encoding="utf-8") as f:
	makefile = f.read()


def make_target_dependencies(target: str) -> list[str]:
	match = re.search(rf"^{re.escape(target)}:[ \t]*(.*)$", makefile, flags=re.MULTILINE)
	if not match:
		add_error(f"Makefile must define target {target}")
		return []
	return [item for item in match.group(1).split() if item and not item.startswith("$")]


def make_target_depends_on(target: str, dependency: str, seen: set[str] | None = None) -> bool:
	if seen is None:
		seen = set()
	if target in seen:
		return False
	seen.add(target)
	deps = make_target_dependencies(target)
	if dependency in deps:
		return True
	if any(make_target_depends_on(dep, dependency, seen) for dep in deps if re.match(r"^[A-Za-z0-9_.-]+$", dep)):
		return True
	recipe_match = re.search(rf"^{re.escape(target)}:.*\n(?P<body>(?:\t.*\n)*)", makefile, flags=re.MULTILINE)
	if not recipe_match:
		return False
	called_targets = re.findall(r"(?:\$\(MAKE\)|make)\s+([A-Za-z0-9_.-]+)", recipe_match.group("body"))
	return any(make_target_depends_on(called, dependency, seen) for called in called_targets)


commands = require_mapping(ai_context.get("commands"), "commands")
for command_name in ("governance_maturity_check", "bench_regression_check"):
	if command_name not in commands:
		add_error(f"commands.{command_name} must declare command side effects")
	elif commands[command_name].get("safe_for_agent_auto_run") is not True:
		add_error(f"commands.{command_name}.safe_for_agent_auto_run must be true")

public_facade_names = [item.get("package") for item in ai_context.get("public_facades", []) if isinstance(item, dict)]
public_facades = set(name for name in public_facade_names if isinstance(name, str) and name)
tool_packages = {pkg.get("name"): pkg for pkg in tools.get("packages", []) if isinstance(pkg, dict)}
if len(public_facade_names) != len(public_facades):
	add_error("public_facades must not contain duplicate package names")


def validate_roadmap_catalog_baseline() -> None:
	summary = require_mapping(tools.get("summary"), "docs/api/tools.json.summary")
	status_counts = require_mapping(summary.get("status_counts"), "docs/api/tools.json.summary.status_counts")
	synopsis_sources = require_mapping(summary.get("synopsis_sources"), "docs/api/tools.json.summary.synopsis_sources")
	expected = {
		"Public facade packages": summary.get("package_count"),
		"Public functions": summary.get("function_count"),
		"Functions with executable examples": summary.get("functions_with_examples"),
		"Context-aware functions": summary.get("context_aware_functions"),
		"Functions returning errors": summary.get("returns_error_functions"),
		"Recommended public functions": status_counts.get("recommended"),
		"Compatibility public functions": status_counts.get("compatibility"),
		"Empty function synopses": synopsis_sources.get("empty"),
		"Facade-sourced function synopses": synopsis_sources.get("facade"),
		"Internal-sourced function synopses": synopsis_sources.get("internal"),
	}
	actual = extract_markdown_table(root / "docs/superpowers/plans/49-roadmap.md", "Baseline")
	for metric, expected_value in expected.items():
		if not isinstance(expected_value, int) or isinstance(expected_value, bool):
			add_error(f"docs/api/tools.json.summary source for {metric} must be an integer")
			continue
		actual_value = actual.get(metric)
		if actual_value is None:
			add_error(f"docs/superpowers/plans/49-roadmap.md Baseline missing metric {metric}")
			continue
		if actual_value != expected_value:
			add_error(
				"docs/superpowers/plans/49-roadmap.md Baseline "
				f"{metric}={actual_value} must match docs/api/tools.json.summary value {expected_value}"
			)
	extra_metrics = sorted(set(actual) - set(expected))
	if extra_metrics:
		add_error("docs/superpowers/plans/49-roadmap.md Baseline includes unmanaged metric(s): " + ", ".join(extra_metrics))


def package_summary_int(package_name: str, field: str) -> int | None:
	pkg = tool_packages.get(package_name)
	if not pkg:
		add_error(f"docs/api/tools.json missing package {package_name}")
		return None
	summary = require_mapping(pkg.get("summary"), f"docs/api/tools.json.packages.{package_name}.summary")
	value = summary.get(field)
	if not isinstance(value, int) or isinstance(value, bool):
		add_error(f"docs/api/tools.json.packages.{package_name}.summary.{field} must be an integer")
		return None
	return value


def package_summary(package_name: str) -> dict:
	pkg = tool_packages.get(package_name)
	if not pkg:
		add_error(f"docs/api/tools.json missing package {package_name}")
		return {}
	return require_mapping(pkg.get("summary"), f"docs/api/tools.json.packages.{package_name}.summary")


def parse_int_cell(value: str, path: str) -> int | None:
	number = value.replace(",", "")
	if not re.fullmatch(r"\d+", number):
		add_error(f"{path} must be an integer, got {value!r}")
		return None
	return int(number)


def validate_roadmap_star_domain_scorecard() -> None:
	roadmap = root / "docs/superpowers/plans/49-roadmap.md"
	domains = {
		"Safe HTTP (`vhttp`, `vresty`, `vurl`)": ("vhttp", "vresty", "vurl"),
		"Safe crypto (`vcrypto`, `vrand`, `vjwt`)": ("vcrypto", "vrand", "vjwt"),
		"Daily JSON/file (`vjson`, `vfile`)": ("vjson", "vfile"),
	}
	rows = {
		row.get("Domain", ""): row
		for row in extract_markdown_rows(roadmap, "90-Day Star Domain Scorecard")
	}
	missing = sorted(set(domains) - set(rows))
	if missing:
		add_error("docs/superpowers/plans/49-roadmap.md scorecard missing domain row(s): " + ", ".join(missing))
	extra = sorted(set(rows) - set(domains))
	if extra:
		add_error("docs/superpowers/plans/49-roadmap.md scorecard includes unmanaged domain row(s): " + ", ".join(extra))
	for domain, packages in domains.items():
		row = rows.get(domain)
		if not row:
			continue
		function_count = 0
		example_count = 0
		for package_name in packages:
			package_functions = package_summary_int(package_name, "function_count")
			package_examples = package_summary_int(package_name, "functions_with_examples")
			if package_functions is not None:
				function_count += package_functions
			if package_examples is not None:
				example_count += package_examples
		actual_functions = parse_int_cell(row.get("Public functions", ""), f"{domain} Public functions")
		actual_examples = parse_int_cell(row.get("Examples", ""), f"{domain} Examples")
		if actual_functions is not None and actual_functions != function_count:
			add_error(f"{domain} Public functions={actual_functions} must match tools catalog value {function_count}")
		if actual_examples is not None and actual_examples != example_count:
			add_error(f"{domain} Examples={actual_examples} must match tools catalog value {example_count}")
		expected_ratio = "0.0%" if function_count == 0 else f"{example_count / function_count * 100:.1f}%"
		actual_ratio = row.get("Example ratio", "")
		if actual_ratio != expected_ratio:
			add_error(f"{domain} Example ratio={actual_ratio!r} must match tools catalog value {expected_ratio!r}")


def validate_safe_http_cookbook_governance() -> None:
	governance = require_mapping(ai_context.get("safe_http_cookbook_governance"), "safe_http_cookbook_governance")
	roadmap_path = governance.get("roadmap_path")
	if not isinstance(roadmap_path, str) or not roadmap_path.strip():
		add_error("safe_http_cookbook_governance.roadmap_path must be non-empty")
		roadmap_path = "docs/superpowers/plans/49-roadmap.md"
	cookbook_path = governance.get("cookbook_path")
	if not isinstance(cookbook_path, str) or not cookbook_path.strip():
		add_error("safe_http_cookbook_governance.cookbook_path must be non-empty")
		cookbook_path = "docs/doc/safe-http-cookbook.md"
	sprint = governance.get("sprint")
	if sprint != 23:
		add_error("safe_http_cookbook_governance.sprint must be 23")
	status = governance.get("status")
	if status not in {"active", "completed"}:
		add_error("safe_http_cookbook_governance.status must be active or completed")
	packages = require_string_list(governance.get("packages"), "safe_http_cookbook_governance.packages")
	expected_packages = ["vhttp", "vresty", "vurl"]
	if packages != expected_packages:
		add_error("safe_http_cookbook_governance.packages must be ordered as: " + ", ".join(expected_packages))
	required_scenarios = require_string_list(governance.get("required_scenarios"), "safe_http_cookbook_governance.required_scenarios")
	if len(required_scenarios) < 4:
		add_error("safe_http_cookbook_governance.required_scenarios must include at least four cookbook scenarios")
	required_checks = require_string_list(governance.get("required_checks"), "safe_http_cookbook_governance.required_checks")
	for check in ("docs-check", "ai-context-check", "governance-maturity-check", "agent-security-check"):
		if check not in required_checks:
			add_error(f"safe_http_cookbook_governance.required_checks must include {check}")
	for package_name in packages:
		if package_name not in public_facades:
			add_error(f"safe_http_cookbook_governance.packages references non-public facade {package_name}")
	cookbook_file = root / cookbook_path
	try:
		cookbook_text = cookbook_file.read_text(encoding="utf-8")
	except FileNotFoundError:
		add_error(f"{cookbook_path} must exist")
		cookbook_text = ""
	if cookbook_text:
		if not cookbook_text.startswith("# Safe HTTP Cookbook\n"):
			add_error(f"{cookbook_path} must start with '# Safe HTTP Cookbook'")
		for package_name in packages:
			if f"`{package_name}`" not in cookbook_text and f"/{package_name}" not in cookbook_text:
				add_error(f"{cookbook_path} must mention {package_name}")
		for scenario in required_scenarios:
			if scenario not in cookbook_text:
				add_error(f"{cookbook_path} missing required scenario {scenario!r}")
		for phrase in ("Trust Boundary Checklist", "WithAllowedHosts", "WithLookupIP", "WithMaxResponseBytes", "DownloadFileSafe"):
			if phrase not in cookbook_text:
				add_error(f"{cookbook_path} must include {phrase!r}")
		for check in required_checks:
			if f"make {check}" not in cookbook_text:
				add_error(f"{cookbook_path} validation section must mention make {check}")
	roadmap_rows = extract_markdown_rows(root / roadmap_path, "90-Day Star Domain Scorecard")
	safe_http_rows = [row for row in roadmap_rows if row.get("Domain") == "Safe HTTP (`vhttp`, `vresty`, `vurl`)"]
	if len(safe_http_rows) != 1:
		add_error(f"{roadmap_path} scorecard must contain exactly one Safe HTTP row")
	else:
		cookbook_status = safe_http_rows[0].get("Cookbook status", "")
		if cookbook_path not in cookbook_status or "Present" not in cookbook_status:
			add_error(f"{roadmap_path} Safe HTTP Cookbook status must be Present and reference {cookbook_path}")
	sprint_rows = extract_markdown_rows(root / roadmap_path, "Sprint order")
	sprint_23_rows = [row for row in sprint_rows if row.get("Sprint") == "23"]
	if len(sprint_23_rows) != 1:
		add_error(f"{roadmap_path} Sprint order must contain exactly one Sprint 23 row")
	else:
		sprint_23 = sprint_23_rows[0]
		expected_status = "Completed" if status == "completed" else "Active"
		if sprint_23.get("Status") != expected_status:
			add_error(f"{roadmap_path} Sprint 23 status must be {expected_status}")
		sprint_text = " ".join(sprint_23.values())
		missing_from_sprint = [package for package in packages if f"`{package}`" not in sprint_text]
		if missing_from_sprint:
			add_error(f"{roadmap_path} Sprint 23 row missing package(s): " + ", ".join(missing_from_sprint))


def validate_safe_crypto_cookbook_governance() -> None:
	governance = require_mapping(ai_context.get("safe_crypto_cookbook_governance"), "safe_crypto_cookbook_governance")
	roadmap_path = governance.get("roadmap_path")
	if not isinstance(roadmap_path, str) or not roadmap_path.strip():
		add_error("safe_crypto_cookbook_governance.roadmap_path must be non-empty")
		roadmap_path = "docs/superpowers/plans/49-roadmap.md"
	cookbook_path = governance.get("cookbook_path")
	if not isinstance(cookbook_path, str) or not cookbook_path.strip():
		add_error("safe_crypto_cookbook_governance.cookbook_path must be non-empty")
		cookbook_path = "docs/doc/safe-crypto-cookbook.md"
	sprint = governance.get("sprint")
	if sprint != 24:
		add_error("safe_crypto_cookbook_governance.sprint must be 24")
	status = governance.get("status")
	if status not in {"active", "completed"}:
		add_error("safe_crypto_cookbook_governance.status must be active or completed")
	packages = require_string_list(governance.get("packages"), "safe_crypto_cookbook_governance.packages")
	expected_packages = ["vcrypto", "vrand", "vjwt"]
	if packages != expected_packages:
		add_error("safe_crypto_cookbook_governance.packages must be ordered as: " + ", ".join(expected_packages))
	required_scenarios = require_string_list(governance.get("required_scenarios"), "safe_crypto_cookbook_governance.required_scenarios")
	if len(required_scenarios) < 4:
		add_error("safe_crypto_cookbook_governance.required_scenarios must include at least four cookbook scenarios")
	required_checks = require_string_list(governance.get("required_checks"), "safe_crypto_cookbook_governance.required_checks")
	for check in ("docs-check", "ai-context-check", "governance-maturity-check", "agent-security-check"):
		if check not in required_checks:
			add_error(f"safe_crypto_cookbook_governance.required_checks must include {check}")
	for package_name in packages:
		if package_name not in public_facades:
			add_error(f"safe_crypto_cookbook_governance.packages references non-public facade {package_name}")
	cookbook_file = root / cookbook_path
	try:
		cookbook_text = cookbook_file.read_text(encoding="utf-8")
	except FileNotFoundError:
		add_error(f"{cookbook_path} must exist")
		cookbook_text = ""
	if cookbook_text:
		if not cookbook_text.startswith("# Safe Crypto Cookbook\n"):
			add_error(f"{cookbook_path} must start with '# Safe Crypto Cookbook'")
		for package_name in packages:
			if f"`{package_name}`" not in cookbook_text and f"/{package_name}" not in cookbook_text:
				add_error(f"{cookbook_path} must mention {package_name}")
		for scenario in required_scenarios:
			if scenario not in cookbook_text:
				add_error(f"{cookbook_path} missing required scenario {scenario!r}")
		for phrase in (
			"Crypto Decision Matrix",
			"Secret Boundary Checklist",
			"SecureBytes",
			"HMACEqual",
			"AESSealGCM",
			"AESOpenGCM",
			"CreateTokenWithOptions",
			"ValidateWithOptions",
		):
			if phrase not in cookbook_text:
				add_error(f"{cookbook_path} must include {phrase!r}")
		for check in required_checks:
			if f"make {check}" not in cookbook_text:
				add_error(f"{cookbook_path} validation section must mention make {check}")
	roadmap_rows = extract_markdown_rows(root / roadmap_path, "90-Day Star Domain Scorecard")
	safe_crypto_rows = [row for row in roadmap_rows if row.get("Domain") == "Safe crypto (`vcrypto`, `vrand`, `vjwt`)"]
	if len(safe_crypto_rows) != 1:
		add_error(f"{roadmap_path} scorecard must contain exactly one Safe crypto row")
	else:
		row = safe_crypto_rows[0]
		for column in ("Comparison page status", "Cookbook status"):
			status_text = row.get(column, "")
			if cookbook_path not in status_text or "Present" not in status_text:
				add_error(f"{roadmap_path} Safe crypto {column} must be Present and reference {cookbook_path}")
	sprint_rows = extract_markdown_rows(root / roadmap_path, "Sprint order")
	sprint_24_rows = [row for row in sprint_rows if row.get("Sprint") == "24"]
	if len(sprint_24_rows) != 1:
		add_error(f"{roadmap_path} Sprint order must contain exactly one Sprint 24 row")
	else:
		sprint_24 = sprint_24_rows[0]
		expected_status = "Completed" if status == "completed" else "Active"
		if sprint_24.get("Status") != expected_status:
			add_error(f"{roadmap_path} Sprint 24 status must be {expected_status}")
		sprint_text = " ".join(sprint_24.values())
		missing_from_sprint = [package for package in packages if f"`{package}`" not in sprint_text]
		if missing_from_sprint:
			add_error(f"{roadmap_path} Sprint 24 row missing package(s): " + ", ".join(missing_from_sprint))


def validate_daily_json_file_faq_governance() -> None:
	governance = require_mapping(ai_context.get("daily_json_file_faq_governance"), "daily_json_file_faq_governance")
	roadmap_path = governance.get("roadmap_path")
	if not isinstance(roadmap_path, str) or not roadmap_path.strip():
		add_error("daily_json_file_faq_governance.roadmap_path must be non-empty")
		roadmap_path = "docs/superpowers/plans/49-roadmap.md"
	faq_path = governance.get("faq_path")
	if not isinstance(faq_path, str) or not faq_path.strip():
		add_error("daily_json_file_faq_governance.faq_path must be non-empty")
		faq_path = "docs/doc/daily-json-file-faq.md"
	sprint = governance.get("sprint")
	if sprint != 25:
		add_error("daily_json_file_faq_governance.sprint must be 25")
	status = governance.get("status")
	if status not in {"active", "completed"}:
		add_error("daily_json_file_faq_governance.status must be active or completed")
	packages = require_string_list(governance.get("packages"), "daily_json_file_faq_governance.packages")
	expected_packages = ["vjson", "vfile"]
	if packages != expected_packages:
		add_error("daily_json_file_faq_governance.packages must be ordered as: " + ", ".join(expected_packages))
	required_questions = require_string_list(
		governance.get("required_questions"),
		"daily_json_file_faq_governance.required_questions",
	)
	if len(required_questions) < 5:
		add_error("daily_json_file_faq_governance.required_questions must include at least five FAQ questions")
	required_checks = require_string_list(governance.get("required_checks"), "daily_json_file_faq_governance.required_checks")
	for check in ("docs-check", "ai-context-check", "governance-maturity-check"):
		if check not in required_checks:
			add_error(f"daily_json_file_faq_governance.required_checks must include {check}")
	for package_name in packages:
		if package_name not in public_facades:
			add_error(f"daily_json_file_faq_governance.packages references non-public facade {package_name}")
	faq_file = root / faq_path
	try:
		faq_text = faq_file.read_text(encoding="utf-8")
	except FileNotFoundError:
		add_error(f"{faq_path} must exist")
		faq_text = ""
	if faq_text:
		if not faq_text.startswith("# Daily JSON/File FAQ\n"):
			add_error(f"{faq_path} must start with '# Daily JSON/File FAQ'")
		for package_name in packages:
			if f"`{package_name}`" not in faq_text and f"/{package_name}" not in faq_text:
				add_error(f"{faq_path} must mention {package_name}")
		for question in required_questions:
			if f"### {question}" not in faq_text:
				add_error(f"{faq_path} missing required FAQ question {question!r}")
		for phrase in (
			"Decision Matrix",
			"encoding/json",
			"os",
			"io",
			"bounded reads",
			"explicit errors",
			"provider-backed",
			"untrusted JSON input",
			"untrusted file paths",
		):
			if phrase not in faq_text:
				add_error(f"{faq_path} must include {phrase!r}")
		for check in required_checks:
			if f"make {check}" not in faq_text:
				add_error(f"{faq_path} validation section must mention make {check}")
	roadmap_rows = extract_markdown_rows(root / roadmap_path, "90-Day Star Domain Scorecard")
	daily_rows = [row for row in roadmap_rows if row.get("Domain") == "Daily JSON/file (`vjson`, `vfile`)"]
	if len(daily_rows) != 1:
		add_error(f"{roadmap_path} scorecard must contain exactly one Daily JSON/file row")
	else:
		faq_status = daily_rows[0].get("FAQ status", "")
		if faq_path not in faq_status or "Present" not in faq_status:
			add_error(f"{roadmap_path} Daily JSON/file FAQ status must be Present and reference {faq_path}")
	sprint_rows = extract_markdown_rows(root / roadmap_path, "Sprint order")
	sprint_25_rows = [row for row in sprint_rows if row.get("Sprint") == "25"]
	if len(sprint_25_rows) != 1:
		add_error(f"{roadmap_path} Sprint order must contain exactly one Sprint 25 row")
	else:
		sprint_25 = sprint_25_rows[0]
		expected_status = "Completed" if status == "completed" else "Active"
		if sprint_25.get("Status") != expected_status:
			add_error(f"{roadmap_path} Sprint 25 status must be {expected_status}")
		sprint_text = " ".join(sprint_25.values())
		missing_from_sprint = [package for package in packages if f"`{package}`" not in sprint_text]
		if missing_from_sprint:
			add_error(f"{roadmap_path} Sprint 25 row missing package(s): " + ", ".join(missing_from_sprint))


def validate_star_domain_no_missing_governance() -> None:
	governance = require_mapping(ai_context.get("star_domain_no_missing_governance"), "star_domain_no_missing_governance")
	roadmap_path = governance.get("roadmap_path")
	if not isinstance(roadmap_path, str) or not roadmap_path.strip():
		add_error("star_domain_no_missing_governance.roadmap_path must be non-empty")
		roadmap_path = "docs/superpowers/plans/49-roadmap.md"
	sprint = governance.get("sprint")
	if sprint != 26:
		add_error("star_domain_no_missing_governance.sprint must be 26")
	status = governance.get("status")
	if status not in {"active", "completed"}:
		add_error("star_domain_no_missing_governance.status must be active or completed")
	domains = require_string_list(governance.get("domains"), "star_domain_no_missing_governance.domains")
	expected_domains = [
		"Safe HTTP (`vhttp`, `vresty`, `vurl`)",
		"Safe crypto (`vcrypto`, `vrand`, `vjwt`)",
		"Daily JSON/file (`vjson`, `vfile`)",
	]
	if domains != expected_domains:
		add_error("star_domain_no_missing_governance.domains must be ordered as: " + ", ".join(expected_domains))
	status_columns = require_string_list(
		governance.get("status_columns"),
		"star_domain_no_missing_governance.status_columns",
	)
	expected_status_columns = [
		"Recommended API docs status",
		"FAQ status",
		"Comparison page status",
		"Cookbook status",
	]
	if status_columns != expected_status_columns:
		add_error("star_domain_no_missing_governance.status_columns must be ordered as: " + ", ".join(expected_status_columns))
	required_checks = require_string_list(governance.get("required_checks"), "star_domain_no_missing_governance.required_checks")
	for check in ("docs-check", "ai-context-check", "governance-maturity-check"):
		if check not in required_checks:
			add_error(f"star_domain_no_missing_governance.required_checks must include {check}")

	scorecard_rows = {
		row.get("Domain", ""): row
		for row in extract_markdown_rows(root / roadmap_path, "90-Day Star Domain Scorecard")
	}
	for domain in domains:
		row = scorecard_rows.get(domain)
		if row is None:
			add_error(f"{roadmap_path} scorecard missing governed domain {domain}")
			continue
		for column in status_columns:
			value = row.get(column)
			if value is None:
				add_error(f"{roadmap_path} {domain} missing governed status column {column}")
				continue
			if "Missing" in value:
				add_error(f"{roadmap_path} {domain} {column} must not contain Missing")
			if "Present" not in value:
				add_error(f"{roadmap_path} {domain} {column} must contain Present")

	sprint_rows = extract_markdown_rows(root / roadmap_path, "Sprint order")
	sprint_26_rows = [row for row in sprint_rows if row.get("Sprint") == "26"]
	if len(sprint_26_rows) != 1:
		add_error(f"{roadmap_path} Sprint order must contain exactly one Sprint 26 row")
	else:
		sprint_26 = sprint_26_rows[0]
		expected_status = "Completed" if status == "completed" else "Active"
		if sprint_26.get("Status") != expected_status:
			add_error(f"{roadmap_path} Sprint 26 status must be {expected_status}")
		sprint_text = " ".join(sprint_26.values())
		for required_phrase in ("star-domain", "Missing", "governance"):
			if required_phrase not in sprint_text:
				add_error(f"{roadmap_path} Sprint 26 row must mention {required_phrase!r}")


def validate_vdb_deepening_governance() -> None:
	governance = require_mapping(ai_context.get("vdb_deepening_governance"), "vdb_deepening_governance")
	roadmap_path = governance.get("roadmap_path")
	if not isinstance(roadmap_path, str) or not roadmap_path.strip():
		add_error("vdb_deepening_governance.roadmap_path must be non-empty")
		roadmap_path = "docs/superpowers/plans/49-roadmap.md"
	doc_path = governance.get("doc_path")
	if not isinstance(doc_path, str) or not doc_path.strip():
		add_error("vdb_deepening_governance.doc_path must be non-empty")
		doc_path = "docs/doc/vdb-deepening-backlog.md"
	quickstart_path = governance.get("quickstart_path")
	if not isinstance(quickstart_path, str) or not quickstart_path.strip():
		add_error("vdb_deepening_governance.quickstart_path must be non-empty")
		quickstart_path = "docs/doc/14-vdb.md"
	sprint = governance.get("sprint")
	if sprint != 27:
		add_error("vdb_deepening_governance.sprint must be 27")
	status = governance.get("status")
	if status not in {"active", "completed"}:
		add_error("vdb_deepening_governance.status must be active or completed")
	packages = require_string_list(governance.get("packages"), "vdb_deepening_governance.packages")
	if packages != ["vdb"]:
		add_error("vdb_deepening_governance.packages must be ordered as: vdb")
	internal_packages = require_string_list(governance.get("internal_packages"), "vdb_deepening_governance.internal_packages")
	if internal_packages != ["internal/db"]:
		add_error("vdb_deepening_governance.internal_packages must be ordered as: internal/db")
	for package_name in packages:
		if package_name not in public_facades:
			add_error(f"vdb_deepening_governance.packages references non-public facade {package_name}")
	required_lanes = require_string_list(governance.get("required_lanes"), "vdb_deepening_governance.required_lanes")
	expected_lanes = [
		"Context-first execution",
		"Dialect depth",
		"Batch operations",
		"Upsert semantics",
		"Scan helpers",
		"Transaction behavior",
		"Identifier safety",
		"Benchmark scope",
	]
	if required_lanes != expected_lanes:
		add_error("vdb_deepening_governance.required_lanes must be ordered as: " + ", ".join(expected_lanes))
	required_tests = require_string_list(governance.get("required_tests"), "vdb_deepening_governance.required_tests")
	expected_tests = [
		"internal/db/session_exec_test.go",
		"internal/db/builder_write_test.go",
		"internal/db/db_sql_helpers_test.go",
		"vdb/session_exec_test.go",
		"vdb/error_contract_test.go",
	]
	if required_tests != expected_tests:
		add_error("vdb_deepening_governance.required_tests must be ordered as: " + ", ".join(expected_tests))
	for test_path in required_tests:
		if not (root / test_path).exists():
			add_error(f"vdb_deepening_governance.required_tests references missing file {test_path}")
	required_checks = require_string_list(governance.get("required_checks"), "vdb_deepening_governance.required_checks")
	for check in ("docs-check", "ai-context-check", "governance-maturity-check", "agent-security-check"):
		if check not in required_checks:
			add_error(f"vdb_deepening_governance.required_checks must include {check}")

	doc_file = root / doc_path
	try:
		doc_text = doc_file.read_text(encoding="utf-8")
	except FileNotFoundError:
		add_error(f"{doc_path} must exist")
		doc_text = ""
	if doc_text:
		if not doc_text.startswith("# vdb Deepening Backlog\n"):
			add_error(f"{doc_path} must start with '# vdb Deepening Backlog'")
		for phrase in ("database/sql", "parameterized", "Non-Goals", "Required Evidence", "Validation"):
			if phrase not in doc_text:
				add_error(f"{doc_path} must include {phrase!r}")
		for lane in required_lanes:
			if lane not in doc_text:
				add_error(f"{doc_path} missing required lane {lane!r}")
		for test_path in required_tests:
			if f"`{test_path}`" not in doc_text:
				add_error(f"{doc_path} must mention required test {test_path}")
		for check in required_checks:
			if f"make {check}" not in doc_text:
				add_error(f"{doc_path} validation section must mention make {check}")
		for non_goal in ("Do not turn `vdb` into an ORM", "Do not add driver dependencies", "Do not own migrations"):
			if non_goal not in doc_text:
				add_error(f"{doc_path} must include non-goal {non_goal!r}")
	quickstart_file = root / quickstart_path
	try:
		quickstart_text = quickstart_file.read_text(encoding="utf-8")
	except FileNotFoundError:
		add_error(f"{quickstart_path} must exist")
		quickstart_text = ""
	if quickstart_text:
		readme_text = (root / "docs/doc/README.md").read_text(encoding="utf-8")
		doc_link = Path(doc_path).name
		if doc_path not in readme_text and doc_link not in readme_text:
			add_error(f"docs/doc/README.md must link {doc_path}")

	sprint_rows = extract_markdown_rows(root / roadmap_path, "Sprint order")
	sprint_27_rows = [row for row in sprint_rows if row.get("Sprint") == "27"]
	if len(sprint_27_rows) != 1:
		add_error(f"{roadmap_path} Sprint order must contain exactly one Sprint 27 row")
	else:
		sprint_27 = sprint_27_rows[0]
		expected_status = "Completed" if status == "completed" else "Active"
		if sprint_27.get("Status") != expected_status:
			add_error(f"{roadmap_path} Sprint 27 status must be {expected_status}")
		sprint_text = " ".join(sprint_27.values())
		for required_phrase in ("`vdb`", "deepening backlog", "context-first", "dialect"):
			if required_phrase not in sprint_text:
				add_error(f"{roadmap_path} Sprint 27 row must mention {required_phrase!r}")


def validate_vdb_execution_evidence_governance() -> None:
	governance = require_mapping(ai_context.get("vdb_execution_evidence_governance"), "vdb_execution_evidence_governance")
	roadmap_path = governance.get("roadmap_path")
	if not isinstance(roadmap_path, str) or not roadmap_path.strip():
		add_error("vdb_execution_evidence_governance.roadmap_path must be non-empty")
		roadmap_path = "docs/superpowers/plans/49-roadmap.md"
	sprint = governance.get("sprint")
	if sprint != 28:
		add_error("vdb_execution_evidence_governance.sprint must be 28")
	status = governance.get("status")
	if status not in {"active", "completed"}:
		add_error("vdb_execution_evidence_governance.status must be active or completed")
	packages = require_string_list(governance.get("packages"), "vdb_execution_evidence_governance.packages")
	if packages != ["vdb"]:
		add_error("vdb_execution_evidence_governance.packages must be ordered as: vdb")
	internal_packages = require_string_list(governance.get("internal_packages"), "vdb_execution_evidence_governance.internal_packages")
	if internal_packages != ["internal/db"]:
		add_error("vdb_execution_evidence_governance.internal_packages must be ordered as: internal/db")
	required_contracts = require_string_list(governance.get("required_contracts"), "vdb_execution_evidence_governance.required_contracts")
	expected_contracts = [
		"ExecBatch partial failure",
		"Upsert dialect behavior",
		"Tx rollback and commit errors",
		"Scan edge cases",
		"Identifier safety",
	]
	if required_contracts != expected_contracts:
		add_error("vdb_execution_evidence_governance.required_contracts must be ordered as: " + ", ".join(expected_contracts))
	required_test_functions = require_string_list(
		governance.get("required_test_functions"),
		"vdb_execution_evidence_governance.required_test_functions",
	)
	expected_test_functions = [
		"internal/db/db_exec_test.go:TestDBExecBatchPartialFailureReturnsCompletedResults",
		"internal/db/builder_write_test.go:TestUpsertSQLDialectVariants",
		"internal/db/db_exec_test.go:TestDBTxRollbackErrorJoinsCause",
		"internal/db/db_sql_helpers_test.go:TestScanRowsNormalizesBytesAndReportsIteratorErrors",
		"internal/db/builder_select_test.go:TestSQLBuilderRejectsUnsafeIdentifiers",
	]
	if required_test_functions != expected_test_functions:
		add_error("vdb_execution_evidence_governance.required_test_functions must be ordered as: " + ", ".join(expected_test_functions))
	for reference in required_test_functions:
		if not test_function_exists(reference):
			add_error(f"vdb_execution_evidence_governance.required_test_functions references missing test {reference}")
	required_checks = require_string_list(governance.get("required_checks"), "vdb_execution_evidence_governance.required_checks")
	for check in ("go test ./internal/db ./vdb", "docs-check", "ai-context-check", "governance-maturity-check", "agent-security-check"):
		if check not in required_checks:
			add_error(f"vdb_execution_evidence_governance.required_checks must include {check}")

	sprint_rows = extract_markdown_rows(root / roadmap_path, "Sprint order")
	sprint_28_rows = [row for row in sprint_rows if row.get("Sprint") == "28"]
	if len(sprint_28_rows) != 1:
		add_error(f"{roadmap_path} Sprint order must contain exactly one Sprint 28 row")
	else:
		sprint_28 = sprint_28_rows[0]
		expected_status = "Completed" if status == "completed" else "Active"
		if sprint_28.get("Status") != expected_status:
			add_error(f"{roadmap_path} Sprint 28 status must be {expected_status}")
		sprint_text = " ".join(sprint_28.values())
		for required_phrase in ("`vdb`", "ExecBatch", "Upsert", "Tx", "scan", "identifier safety"):
			if required_phrase not in sprint_text:
				add_error(f"{roadmap_path} Sprint 28 row must mention {required_phrase!r}")


def validate_vdb_example_depth_governance() -> None:
	governance = require_mapping(ai_context.get("vdb_example_depth_governance"), "vdb_example_depth_governance")
	roadmap_path = governance.get("roadmap_path")
	if not isinstance(roadmap_path, str) or not roadmap_path.strip():
		add_error("vdb_example_depth_governance.roadmap_path must be non-empty")
		roadmap_path = "docs/superpowers/plans/49-roadmap.md"
	sprint = governance.get("sprint")
	if sprint != 29:
		add_error("vdb_example_depth_governance.sprint must be 29")
	status = governance.get("status")
	if status not in {"active", "completed"}:
		add_error("vdb_example_depth_governance.status must be active or completed")
	package_name = governance.get("package")
	if package_name != "vdb":
		add_error("vdb_example_depth_governance.package must be vdb")
	baseline_examples = governance.get("baseline_examples")
	if baseline_examples != 10:
		add_error("vdb_example_depth_governance.baseline_examples must be 10")
	target_examples = governance.get("target_examples")
	if not isinstance(target_examples, int) or isinstance(target_examples, bool) or target_examples < 20:
		add_error("vdb_example_depth_governance.target_examples must be at least 20")
	required_examples = require_string_list(governance.get("required_examples"), "vdb_example_depth_governance.required_examples")
	expected_examples = [
		"ExampleDB_ExecBatch",
		"ExampleDB_Upsert",
		"ExampleDB_Tx",
		"ExampleScanRows",
		"ExampleScanOne",
		"ExampleNewPageResult",
		"ExampleNormalizeDialect",
		"ExampleWrapperForDialect",
		"ExampleRaw",
		"ExampleNewWrapper",
	]
	if required_examples != expected_examples:
		add_error("vdb_example_depth_governance.required_examples must be ordered as: " + ", ".join(expected_examples))
	for example_name in required_examples:
		if not example_function_exists("vdb", example_name):
			add_error(f"vdb_example_depth_governance.required_examples references missing example {example_name}")
	summary = package_summary("vdb")
	example_count = summary.get("functions_with_examples")
	if not isinstance(example_count, int) or isinstance(example_count, bool):
		add_error("docs/api/tools.json.packages.vdb.summary.functions_with_examples must be an integer")
	elif isinstance(target_examples, int) and example_count < target_examples:
		add_error(f"vdb examples must be at least {target_examples}, got {example_count}")
	required_checks = require_string_list(governance.get("required_checks"), "vdb_example_depth_governance.required_checks")
	for check in ("tools-gen", "docs-check", "ai-context-check", "governance-maturity-check"):
		if check not in required_checks:
			add_error(f"vdb_example_depth_governance.required_checks must include {check}")

	sprint_rows = extract_markdown_rows(root / roadmap_path, "Sprint order")
	sprint_29_rows = [row for row in sprint_rows if row.get("Sprint") == "29"]
	if len(sprint_29_rows) != 1:
		add_error(f"{roadmap_path} Sprint order must contain exactly one Sprint 29 row")
	else:
		sprint_29 = sprint_29_rows[0]
		expected_status = "Completed" if status == "completed" else "Active"
		if sprint_29.get("Status") != expected_status:
			add_error(f"{roadmap_path} Sprint 29 status must be {expected_status}")
		sprint_text = " ".join(sprint_29.values())
		for required_phrase in ("`vdb`", "example", "20+", "ScanRows", "WrapperForDialect"):
			if required_phrase not in sprint_text:
				add_error(f"{roadmap_path} Sprint 29 row must mention {required_phrase!r}")


def validate_safe_crypto_advanced_backlog_governance() -> None:
	governance = require_mapping(
		ai_context.get("safe_crypto_advanced_backlog_governance"),
		"safe_crypto_advanced_backlog_governance",
	)
	roadmap_path = governance.get("roadmap_path")
	if not isinstance(roadmap_path, str) or not roadmap_path.strip():
		add_error("safe_crypto_advanced_backlog_governance.roadmap_path must be non-empty")
		roadmap_path = "docs/superpowers/plans/49-roadmap.md"
	doc_path = governance.get("doc_path")
	if not isinstance(doc_path, str) or not doc_path.strip():
		add_error("safe_crypto_advanced_backlog_governance.doc_path must be non-empty")
		doc_path = "docs/doc/safe-crypto-advanced-backlog.md"
	quickstart_path = governance.get("quickstart_path")
	if not isinstance(quickstart_path, str) or not quickstart_path.strip():
		add_error("safe_crypto_advanced_backlog_governance.quickstart_path must be non-empty")
		quickstart_path = "docs/doc/11-vcrypto.md"
	cookbook_path = governance.get("cookbook_path")
	if not isinstance(cookbook_path, str) or not cookbook_path.strip():
		add_error("safe_crypto_advanced_backlog_governance.cookbook_path must be non-empty")
		cookbook_path = "docs/doc/safe-crypto-cookbook.md"
	sprint = governance.get("sprint")
	if sprint != 30:
		add_error("safe_crypto_advanced_backlog_governance.sprint must be 30")
	status = governance.get("status")
	if status not in {"active", "completed"}:
		add_error("safe_crypto_advanced_backlog_governance.status must be active or completed")
	packages = require_string_list(governance.get("packages"), "safe_crypto_advanced_backlog_governance.packages")
	expected_packages = ["vcrypto", "vjwt", "vrand", "vpass"]
	if packages != expected_packages:
		add_error("safe_crypto_advanced_backlog_governance.packages must be ordered as: " + ", ".join(expected_packages))
	for package_name in packages:
		if package_name not in public_facades:
			add_error(f"safe_crypto_advanced_backlog_governance.packages references non-public facade {package_name}")
	required_lanes = require_string_list(governance.get("required_lanes"), "safe_crypto_advanced_backlog_governance.required_lanes")
	expected_lanes = [
		"TOTP and HOTP",
		"Password hashing",
		"JWK and JWKS",
		"Secret handling",
		"Interoperability boundaries",
		"Benchmark scope",
	]
	if required_lanes != expected_lanes:
		add_error("safe_crypto_advanced_backlog_governance.required_lanes must be ordered as: " + ", ".join(expected_lanes))
	non_goals = require_string_list(governance.get("non_goals"), "safe_crypto_advanced_backlog_governance.non_goals")
	expected_non_goals = [
		"No OAuth or OIDC provider implementation",
		"No password storage service",
		"No key management service",
		"No custom cryptographic primitive",
	]
	if non_goals != expected_non_goals:
		add_error("safe_crypto_advanced_backlog_governance.non_goals must be ordered as: " + ", ".join(expected_non_goals))
	required_checks = require_string_list(governance.get("required_checks"), "safe_crypto_advanced_backlog_governance.required_checks")
	for check in ("docs-check", "ai-context-check", "governance-maturity-check", "agent-security-check"):
		if check not in required_checks:
			add_error(f"safe_crypto_advanced_backlog_governance.required_checks must include {check}")

	doc_file = root / doc_path
	try:
		doc_text = doc_file.read_text(encoding="utf-8")
	except FileNotFoundError:
		add_error(f"{doc_path} must exist")
		doc_text = ""
	if doc_text:
		if not doc_text.startswith("# Safe Crypto Advanced Backlog\n"):
			add_error(f"{doc_path} must start with '# Safe Crypto Advanced Backlog'")
		for phrase in ("Required Evidence", "Validation", "RFC-compatible", "Argon2id", "JWK/JWKS", "No custom cryptographic primitive"):
			if phrase not in doc_text:
				add_error(f"{doc_path} must include {phrase!r}")
		for package_name in packages:
			if f"`{package_name}`" not in doc_text:
				add_error(f"{doc_path} must mention {package_name}")
		for lane in required_lanes:
			if lane not in doc_text:
				add_error(f"{doc_path} missing required lane {lane!r}")
		for non_goal in non_goals:
			if non_goal not in doc_text:
				add_error(f"{doc_path} missing non-goal {non_goal!r}")
		for check in required_checks:
			if f"make {check}" not in doc_text:
				add_error(f"{doc_path} validation section must mention make {check}")
	for path in (quickstart_path, cookbook_path):
		if not (root / path).exists():
			add_error(f"{path} must exist")
	readme_text = (root / "docs/doc/README.md").read_text(encoding="utf-8")
	doc_link = Path(doc_path).name
	if doc_path not in readme_text and doc_link not in readme_text:
		add_error(f"docs/doc/README.md must link {doc_path}")

	sprint_rows = extract_markdown_rows(root / roadmap_path, "Sprint order")
	sprint_30_rows = [row for row in sprint_rows if row.get("Sprint") == "30"]
	if len(sprint_30_rows) != 1:
		add_error(f"{roadmap_path} Sprint order must contain exactly one Sprint 30 row")
	else:
		sprint_30 = sprint_30_rows[0]
		expected_status = "Completed" if status == "completed" else "Active"
		if sprint_30.get("Status") != expected_status:
			add_error(f"{roadmap_path} Sprint 30 status must be {expected_status}")
		sprint_text = " ".join(sprint_30.values())
		for required_phrase in ("TOTP/HOTP", "password hashing", "JWK/JWKS", "benchmark"):
			if required_phrase not in sprint_text:
				add_error(f"{roadmap_path} Sprint 30 row must mention {required_phrase!r}")


def validate_safe_crypto_otp_governance() -> None:
	governance = require_mapping(ai_context.get("safe_crypto_otp_governance"), "safe_crypto_otp_governance")
	roadmap_path = governance.get("roadmap_path")
	if not isinstance(roadmap_path, str) or not roadmap_path.strip():
		add_error("safe_crypto_otp_governance.roadmap_path must be non-empty")
		roadmap_path = "docs/superpowers/plans/49-roadmap.md"
	quickstart_path = governance.get("quickstart_path")
	if not isinstance(quickstart_path, str) or not quickstart_path.strip():
		add_error("safe_crypto_otp_governance.quickstart_path must be non-empty")
		quickstart_path = "docs/doc/11-vcrypto.md"
	backlog_path = governance.get("backlog_path")
	if not isinstance(backlog_path, str) or not backlog_path.strip():
		add_error("safe_crypto_otp_governance.backlog_path must be non-empty")
		backlog_path = "docs/doc/safe-crypto-advanced-backlog.md"
	sprint = governance.get("sprint")
	if sprint != 31:
		add_error("safe_crypto_otp_governance.sprint must be 31")
	status = governance.get("status")
	if status not in {"active", "completed"}:
		add_error("safe_crypto_otp_governance.status must be active or completed")
	packages = require_string_list(governance.get("packages"), "safe_crypto_otp_governance.packages")
	if packages != ["vcrypto"]:
		add_error("safe_crypto_otp_governance.packages must be ordered as: vcrypto")
	internal_packages = require_string_list(governance.get("internal_packages"), "safe_crypto_otp_governance.internal_packages")
	if internal_packages != ["internal/crypto"]:
		add_error("safe_crypto_otp_governance.internal_packages must be ordered as: internal/crypto")
	required_functions = require_string_list(governance.get("required_functions"), "safe_crypto_otp_governance.required_functions")
	expected_functions = [
		"HOTP",
		"HOTPVerify",
		"TOTP",
		"TOTPNow",
		"TOTPVerify",
		"TOTPVerifyNow",
		"OTPAuthURL",
		"GenerateOTPSecret",
		"OTPSecretBase32",
		"ParseOTPSecretBase32",
	]
	if required_functions != expected_functions:
		add_error("safe_crypto_otp_governance.required_functions must be ordered as: " + ", ".join(expected_functions))
	for function_name in required_functions:
		if not reference_exists(f"vcrypto/otp.go:{function_name}"):
			add_error(f"safe_crypto_otp_governance.required_functions references missing facade function {function_name}")
	required_test_functions = require_string_list(governance.get("required_test_functions"), "safe_crypto_otp_governance.required_test_functions")
	expected_test_functions = [
		"internal/crypto/otp_test.go:TestHOTPRFC4226Vectors",
		"internal/crypto/otp_test.go:TestTOTPRFC6238Vectors",
		"internal/crypto/otp_test.go:TestTOTPVerifyWindowAndClock",
		"internal/crypto/otp_test.go:TestOTPSecretBase32AndAuthURL",
		"internal/crypto/otp_test.go:TestOTPErrors",
		"vcrypto/otp_test.go:TestFacadeHOTPTOTP",
	]
	if required_test_functions != expected_test_functions:
		add_error("safe_crypto_otp_governance.required_test_functions must be ordered as: " + ", ".join(expected_test_functions))
	for reference in required_test_functions:
		if not test_function_exists(reference):
			add_error(f"safe_crypto_otp_governance.required_test_functions references missing test {reference}")
	required_examples = require_string_list(governance.get("required_examples"), "safe_crypto_otp_governance.required_examples")
	expected_examples = ["ExampleHOTP", "ExampleTOTP", "ExampleTOTPVerifyNow", "ExampleOTPAuthURL"]
	if required_examples != expected_examples:
		add_error("safe_crypto_otp_governance.required_examples must be ordered as: " + ", ".join(expected_examples))
	for example_name in required_examples:
		if not example_function_exists("vcrypto", example_name):
			add_error(f"safe_crypto_otp_governance.required_examples references missing example {example_name}")
	required_checks = require_string_list(governance.get("required_checks"), "safe_crypto_otp_governance.required_checks")
	for check in ("go test ./internal/crypto ./vcrypto", "tools-gen", "docs-check", "ai-context-check", "governance-maturity-check", "agent-security-check"):
		if check not in required_checks:
			add_error(f"safe_crypto_otp_governance.required_checks must include {check}")
	for path in (quickstart_path, backlog_path):
		if not (root / path).exists():
			add_error(f"{path} must exist")
	quickstart_text = (root / quickstart_path).read_text(encoding="utf-8") if (root / quickstart_path).exists() else ""
	for phrase in ("one-time passwords", "HOTP", "TOTP", "OTPAuthURL", "WithOTPClock", "WithTOTPWindow"):
		if quickstart_text and phrase not in quickstart_text:
			add_error(f"{quickstart_path} must include {phrase!r}")
	backlog_text = (root / backlog_path).read_text(encoding="utf-8") if (root / backlog_path).exists() else ""
	for phrase in ("TOTP and HOTP | Completed", "safe_crypto_otp_governance", "RFC vectors", "clock/window tests"):
		if backlog_text and phrase not in backlog_text:
			add_error(f"{backlog_path} must include {phrase!r}")

	sprint_rows = extract_markdown_rows(root / roadmap_path, "Sprint order")
	sprint_31_rows = [row for row in sprint_rows if row.get("Sprint") == "31"]
	if len(sprint_31_rows) != 1:
		add_error(f"{roadmap_path} Sprint order must contain exactly one Sprint 31 row")
	else:
		sprint_31 = sprint_31_rows[0]
		expected_status = "Completed" if status == "completed" else "Active"
		if sprint_31.get("Status") != expected_status:
			add_error(f"{roadmap_path} Sprint 31 status must be {expected_status}")
		sprint_text = " ".join(sprint_31.values())
		for required_phrase in ("HOTP/TOTP", "Base32", "otpauth", "RFC vectors"):
			if required_phrase not in sprint_text:
				add_error(f"{roadmap_path} Sprint 31 row must mention {required_phrase!r}")


def validate_safe_crypto_password_hashing_governance() -> None:
	governance = require_mapping(
		ai_context.get("safe_crypto_password_hashing_governance"),
		"safe_crypto_password_hashing_governance",
	)
	roadmap_path = governance.get("roadmap_path")
	if not isinstance(roadmap_path, str) or not roadmap_path.strip():
		add_error("safe_crypto_password_hashing_governance.roadmap_path must be non-empty")
		roadmap_path = "docs/superpowers/plans/49-roadmap.md"
	backlog_path = governance.get("backlog_path")
	if not isinstance(backlog_path, str) or not backlog_path.strip():
		add_error("safe_crypto_password_hashing_governance.backlog_path must be non-empty")
		backlog_path = "docs/doc/safe-crypto-advanced-backlog.md"
	quickstart_path = governance.get("quickstart_path")
	if not isinstance(quickstart_path, str) or not quickstart_path.strip():
		add_error("safe_crypto_password_hashing_governance.quickstart_path must be non-empty")
		quickstart_path = "docs/doc/11-vcrypto.md"
	password_strength_path = governance.get("password_strength_path")
	if not isinstance(password_strength_path, str) or not password_strength_path.strip():
		add_error("safe_crypto_password_hashing_governance.password_strength_path must be non-empty")
		password_strength_path = "docs/doc/36-vpass.md"
	sprint = governance.get("sprint")
	if sprint != 32:
		add_error("safe_crypto_password_hashing_governance.sprint must be 32")
	status = governance.get("status")
	if status not in {"active", "completed"}:
		add_error("safe_crypto_password_hashing_governance.status must be active or completed")
	packages = require_string_list(governance.get("packages"), "safe_crypto_password_hashing_governance.packages")
	expected_packages = ["vcrypto", "vpass", "vrand"]
	if packages != expected_packages:
		add_error("safe_crypto_password_hashing_governance.packages must be ordered as: " + ", ".join(expected_packages))
	for package_name in packages:
		if package_name not in public_facades:
			add_error(f"safe_crypto_password_hashing_governance.packages references non-public facade {package_name}")
	algorithm_policy = require_string_list(governance.get("algorithm_policy"), "safe_crypto_password_hashing_governance.algorithm_policy")
	expected_algorithm_policy = [
		"Argon2id-style encoded hash envelope",
		"versioned parameters",
		"explicit salt source",
		"bounded test cost",
	]
	if algorithm_policy != expected_algorithm_policy:
		add_error("safe_crypto_password_hashing_governance.algorithm_policy must be ordered as: " + ", ".join(expected_algorithm_policy))
	required_contracts = require_string_list(governance.get("required_contracts"), "safe_crypto_password_hashing_governance.required_contracts")
	expected_contracts = [
		"encoded hash round trip",
		"mismatch returns false",
		"malformed hash returns explicit error",
		"invalid parameters are rejected",
		"raw passwords are not logged or stored",
	]
	if required_contracts != expected_contracts:
		add_error("safe_crypto_password_hashing_governance.required_contracts must be ordered as: " + ", ".join(expected_contracts))
	non_goals = require_string_list(governance.get("non_goals"), "safe_crypto_password_hashing_governance.non_goals")
	expected_non_goals = [
		"No password storage service",
		"No account lifecycle or reset flow",
		"No breached-password corpus check",
		"No MFA or recovery policy",
	]
	if non_goals != expected_non_goals:
		add_error("safe_crypto_password_hashing_governance.non_goals must be ordered as: " + ", ".join(expected_non_goals))
	required_checks = require_string_list(governance.get("required_checks"), "safe_crypto_password_hashing_governance.required_checks")
	for check in ("docs-check", "ai-context-check", "governance-maturity-check", "agent-security-check"):
		if check not in required_checks:
			add_error(f"safe_crypto_password_hashing_governance.required_checks must include {check}")
	for path in (backlog_path, quickstart_path, password_strength_path):
		if not (root / path).exists():
			add_error(f"{path} must exist")
	backlog_text = (root / backlog_path).read_text(encoding="utf-8") if (root / backlog_path).exists() else ""
	for phrase in (
		"safe_crypto_password_hashing_governance",
		"Argon2id-style",
		"malformed-hash",
		"mismatch",
		"bounded test costs",
	):
		if backlog_text and phrase not in backlog_text:
			add_error(f"{backlog_path} must include {phrase!r}")
	if backlog_text and "Password hashing |" not in backlog_text:
		add_error(f"{backlog_path} must include the Password hashing landed evidence row")
	for non_goal in non_goals:
		if backlog_text and non_goal not in backlog_text:
			add_error(f"{backlog_path} must include non-goal {non_goal!r}")
	quickstart_text = (root / quickstart_path).read_text(encoding="utf-8") if (root / quickstart_path).exists() else ""
	for phrase in ("password storage", "Argon2id", "PBKDF2"):
		if quickstart_text and phrase not in quickstart_text:
			add_error(f"{quickstart_path} must include {phrase!r}")
	strength_text = (root / password_strength_path).read_text(encoding="utf-8") if (root / password_strength_path).exists() else ""
	for phrase in ("password hashing", "storage", "breached-password", "MFA"):
		if strength_text and phrase not in strength_text:
			add_error(f"{password_strength_path} must include {phrase!r}")

	sprint_rows = extract_markdown_rows(root / roadmap_path, "Sprint order")
	sprint_32_rows = [row for row in sprint_rows if row.get("Sprint") == "32"]
	if len(sprint_32_rows) != 1:
		add_error(f"{roadmap_path} Sprint order must contain exactly one Sprint 32 row")
	else:
		sprint_32 = sprint_32_rows[0]
		expected_status = "Completed" if status == "completed" else "Active"
		if sprint_32.get("Status") != expected_status:
			add_error(f"{roadmap_path} Sprint 32 status must be {expected_status}")
		sprint_text = " ".join(sprint_32.values())
		for required_phrase in ("password hashing", "Argon2id", "malformed-hash", "mismatch"):
			if required_phrase not in sprint_text:
				add_error(f"{roadmap_path} Sprint 32 row must mention {required_phrase!r}")


def validate_safe_crypto_argon2id_governance() -> None:
	governance = require_mapping(ai_context.get("safe_crypto_argon2id_governance"), "safe_crypto_argon2id_governance")
	roadmap_path = governance.get("roadmap_path")
	if not isinstance(roadmap_path, str) or not roadmap_path.strip():
		add_error("safe_crypto_argon2id_governance.roadmap_path must be non-empty")
		roadmap_path = "docs/superpowers/plans/49-roadmap.md"
	quickstart_path = governance.get("quickstart_path")
	if not isinstance(quickstart_path, str) or not quickstart_path.strip():
		add_error("safe_crypto_argon2id_governance.quickstart_path must be non-empty")
		quickstart_path = "docs/doc/11-vcrypto.md"
	backlog_path = governance.get("backlog_path")
	if not isinstance(backlog_path, str) or not backlog_path.strip():
		add_error("safe_crypto_argon2id_governance.backlog_path must be non-empty")
		backlog_path = "docs/doc/safe-crypto-advanced-backlog.md"
	sprint = governance.get("sprint")
	if sprint != 33:
		add_error("safe_crypto_argon2id_governance.sprint must be 33")
	status = governance.get("status")
	if status not in {"active", "completed"}:
		add_error("safe_crypto_argon2id_governance.status must be active or completed")
	packages = require_string_list(governance.get("packages"), "safe_crypto_argon2id_governance.packages")
	if packages != ["vcrypto"]:
		add_error("safe_crypto_argon2id_governance.packages must be ordered as: vcrypto")
	internal_packages = require_string_list(governance.get("internal_packages"), "safe_crypto_argon2id_governance.internal_packages")
	if internal_packages != ["internal/crypto"]:
		add_error("safe_crypto_argon2id_governance.internal_packages must be ordered as: internal/crypto")
	required_functions = require_string_list(governance.get("required_functions"), "safe_crypto_argon2id_governance.required_functions")
	expected_functions = [
		"HashPasswordArgon2id",
		"VerifyPasswordArgon2id",
		"ParsePasswordHash",
		"WithArgon2idMemory",
		"WithArgon2idIterations",
		"WithArgon2idParallelism",
		"WithArgon2idSaltLength",
		"WithArgon2idKeyLength",
		"WithPasswordHashRandomOptions",
	]
	if required_functions != expected_functions:
		add_error("safe_crypto_argon2id_governance.required_functions must be ordered as: " + ", ".join(expected_functions))
	for function_name in required_functions:
		if not reference_exists(f"vcrypto/password.go:{function_name}"):
			add_error(f"safe_crypto_argon2id_governance.required_functions references missing facade function {function_name}")
	required_test_functions = require_string_list(governance.get("required_test_functions"), "safe_crypto_argon2id_governance.required_test_functions")
	expected_test_functions = [
		"internal/crypto/password_test.go:TestHashPasswordArgon2idRoundTrip",
		"internal/crypto/password_test.go:TestParsePasswordHash",
		"internal/crypto/password_test.go:TestPasswordHashErrors",
		"vcrypto/password_test.go:TestFacadePasswordHashArgon2id",
		"vcrypto/password_test.go:TestFacadePasswordHashErrors",
	]
	if required_test_functions != expected_test_functions:
		add_error("safe_crypto_argon2id_governance.required_test_functions must be ordered as: " + ", ".join(expected_test_functions))
	for reference in required_test_functions:
		if not test_function_exists(reference):
			add_error(f"safe_crypto_argon2id_governance.required_test_functions references missing test {reference}")
	required_examples = require_string_list(governance.get("required_examples"), "safe_crypto_argon2id_governance.required_examples")
	expected_examples = ["ExampleHashPasswordArgon2id", "ExampleVerifyPasswordArgon2id", "ExampleParsePasswordHash"]
	if required_examples != expected_examples:
		add_error("safe_crypto_argon2id_governance.required_examples must be ordered as: " + ", ".join(expected_examples))
	for example_name in required_examples:
		if not example_function_exists("vcrypto", example_name):
			add_error(f"safe_crypto_argon2id_governance.required_examples references missing example {example_name}")
	required_checks = require_string_list(governance.get("required_checks"), "safe_crypto_argon2id_governance.required_checks")
	for check in ("go test ./internal/crypto ./vcrypto", "tools-gen", "docs-check", "ai-context-check", "governance-maturity-check", "agent-security-check"):
		if check not in required_checks:
			add_error(f"safe_crypto_argon2id_governance.required_checks must include {check}")
	for path in (quickstart_path, backlog_path):
		if not (root / path).exists():
			add_error(f"{path} must exist")
	quickstart_text = (root / quickstart_path).read_text(encoding="utf-8") if (root / quickstart_path).exists() else ""
	for phrase in ("HashPasswordArgon2id", "VerifyPasswordArgon2id", "ParsePasswordHash", "ErrInvalidPasswordHash"):
		if quickstart_text and phrase not in quickstart_text:
			add_error(f"{quickstart_path} must include {phrase!r}")
	backlog_text = (root / backlog_path).read_text(encoding="utf-8") if (root / backlog_path).exists() else ""
	for phrase in ("Password hashing | Completed", "safe_crypto_argon2id_governance", "encoded hash implementation", "malformed-hash errors"):
		if backlog_text and phrase not in backlog_text:
			add_error(f"{backlog_path} must include {phrase!r}")

	sprint_rows = extract_markdown_rows(root / roadmap_path, "Sprint order")
	sprint_33_rows = [row for row in sprint_rows if row.get("Sprint") == "33"]
	if len(sprint_33_rows) != 1:
		add_error(f"{roadmap_path} Sprint order must contain exactly one Sprint 33 row")
	else:
		sprint_33 = sprint_33_rows[0]
		expected_status = "Completed" if status == "completed" else "Active"
		if sprint_33.get("Status") != expected_status:
			add_error(f"{roadmap_path} Sprint 33 status must be {expected_status}")
		sprint_text = " ".join(sprint_33.values())
		for required_phrase in ("Argon2id", "encoded", "mismatch", "malformed-hash"):
			if required_phrase not in sprint_text:
				add_error(f"{roadmap_path} Sprint 33 row must mention {required_phrase!r}")


def validate_safe_crypto_jwk_jwks_governance() -> None:
	governance = require_mapping(ai_context.get("safe_crypto_jwk_jwks_governance"), "safe_crypto_jwk_jwks_governance")
	roadmap_path = governance.get("roadmap_path")
	if not isinstance(roadmap_path, str) or not roadmap_path.strip():
		add_error("safe_crypto_jwk_jwks_governance.roadmap_path must be non-empty")
		roadmap_path = "docs/superpowers/plans/49-roadmap.md"
	backlog_path = governance.get("backlog_path")
	if not isinstance(backlog_path, str) or not backlog_path.strip():
		add_error("safe_crypto_jwk_jwks_governance.backlog_path must be non-empty")
		backlog_path = "docs/doc/safe-crypto-advanced-backlog.md"
	quickstart_path = governance.get("quickstart_path")
	if not isinstance(quickstart_path, str) or not quickstart_path.strip():
		add_error("safe_crypto_jwk_jwks_governance.quickstart_path must be non-empty")
		quickstart_path = "docs/doc/11-vcrypto.md"
	jwt_path = governance.get("jwt_path")
	if not isinstance(jwt_path, str) or not jwt_path.strip():
		add_error("safe_crypto_jwk_jwks_governance.jwt_path must be non-empty")
		jwt_path = "docs/doc/28-vjwt.md"
	sprint = governance.get("sprint")
	if sprint != 34:
		add_error("safe_crypto_jwk_jwks_governance.sprint must be 34")
	status = governance.get("status")
	if status not in {"active", "completed"}:
		add_error("safe_crypto_jwk_jwks_governance.status must be active or completed")
	packages = require_string_list(governance.get("packages"), "safe_crypto_jwk_jwks_governance.packages")
	expected_packages = ["vcrypto", "vjwt"]
	if packages != expected_packages:
		add_error("safe_crypto_jwk_jwks_governance.packages must be ordered as: " + ", ".join(expected_packages))
	for package_name in packages:
		if package_name not in public_facades:
			add_error(f"safe_crypto_jwk_jwks_governance.packages references non-public facade {package_name}")
	scope = require_string_list(governance.get("scope"), "safe_crypto_jwk_jwks_governance.scope")
	expected_scope = [
		"local key material helpers only",
		"RSA-first JWK/JWKS support",
		"EC and OKP explicitly deferred unless implemented with fixtures",
		"select public keys by kid",
		"malformed key errors are explicit",
		"no network discovery",
	]
	if scope != expected_scope:
		add_error("safe_crypto_jwk_jwks_governance.scope must be ordered as: " + ", ".join(expected_scope))
	required_contracts = require_string_list(governance.get("required_contracts"), "safe_crypto_jwk_jwks_governance.required_contracts")
	expected_contracts = [
		"RSA public JWK parse and export",
		"JWKS select by kid",
		"unknown kid returns explicit error",
		"malformed key returns explicit error",
		"no remote discovery or cache",
	]
	if required_contracts != expected_contracts:
		add_error("safe_crypto_jwk_jwks_governance.required_contracts must be ordered as: " + ", ".join(expected_contracts))
	non_goals = require_string_list(governance.get("non_goals"), "safe_crypto_jwk_jwks_governance.non_goals")
	expected_non_goals = [
		"No OAuth or OIDC discovery",
		"No remote JWKS fetch",
		"No JWKS cache or refresh loop",
		"No key rotation daemon",
		"No token validation policy inside vcrypto",
	]
	if non_goals != expected_non_goals:
		add_error("safe_crypto_jwk_jwks_governance.non_goals must be ordered as: " + ", ".join(expected_non_goals))
	required_checks = require_string_list(governance.get("required_checks"), "safe_crypto_jwk_jwks_governance.required_checks")
	for check in ("docs-check", "ai-context-check", "governance-maturity-check", "agent-security-check"):
		if check not in required_checks:
			add_error(f"safe_crypto_jwk_jwks_governance.required_checks must include {check}")
	for path in (backlog_path, quickstart_path, jwt_path):
		if not (root / path).exists():
			add_error(f"{path} must exist")
	backlog_text = (root / backlog_path).read_text(encoding="utf-8") if (root / backlog_path).exists() else ""
	for phrase in (
		"safe_crypto_jwk_jwks_governance",
		"local key material",
		"unknown-`kid` behavior",
		"malformed-key errors",
		"no network discovery",
	):
		if backlog_text and phrase not in backlog_text:
			add_error(f"{backlog_path} must include {phrase!r}")
	if backlog_text and "JWK and JWKS |" not in backlog_text:
		add_error(f"{backlog_path} must include the JWK and JWKS landed evidence row")
	for non_goal in non_goals:
		if backlog_text and non_goal not in backlog_text:
			add_error(f"{backlog_path} must include non-goal {non_goal!r}")
	jwt_text = (root / jwt_path).read_text(encoding="utf-8") if (root / jwt_path).exists() else ""
	for phrase in ("JWKS rotation", "kid", "OAuth/OIDC"):
		if jwt_text and phrase not in jwt_text:
			add_error(f"{jwt_path} must include {phrase!r}")

	sprint_rows = extract_markdown_rows(root / roadmap_path, "Sprint order")
	sprint_34_rows = [row for row in sprint_rows if row.get("Sprint") == "34"]
	if len(sprint_34_rows) != 1:
		add_error(f"{roadmap_path} Sprint order must contain exactly one Sprint 34 row")
	else:
		sprint_34 = sprint_34_rows[0]
		expected_status = "Completed" if status == "completed" else "Active"
		if sprint_34.get("Status") != expected_status:
			add_error(f"{roadmap_path} Sprint 34 status must be {expected_status}")
		sprint_text = " ".join(sprint_34.values())
		for required_phrase in ("JWK/JWKS", "local key material", "unknown-`kid`", "no network discovery"):
			if required_phrase not in sprint_text:
				add_error(f"{roadmap_path} Sprint 34 row must mention {required_phrase!r}")


def validate_safe_crypto_jwk_jwks_implementation_governance() -> None:
	governance = require_mapping(
		ai_context.get("safe_crypto_jwk_jwks_implementation_governance"),
		"safe_crypto_jwk_jwks_implementation_governance",
	)
	roadmap_path = governance.get("roadmap_path")
	if not isinstance(roadmap_path, str) or not roadmap_path.strip():
		add_error("safe_crypto_jwk_jwks_implementation_governance.roadmap_path must be non-empty")
		roadmap_path = "docs/superpowers/plans/49-roadmap.md"
	quickstart_path = governance.get("quickstart_path")
	if not isinstance(quickstart_path, str) or not quickstart_path.strip():
		add_error("safe_crypto_jwk_jwks_implementation_governance.quickstart_path must be non-empty")
		quickstart_path = "docs/doc/11-vcrypto.md"
	backlog_path = governance.get("backlog_path")
	if not isinstance(backlog_path, str) or not backlog_path.strip():
		add_error("safe_crypto_jwk_jwks_implementation_governance.backlog_path must be non-empty")
		backlog_path = "docs/doc/safe-crypto-advanced-backlog.md"
	sprint = governance.get("sprint")
	if sprint != 35:
		add_error("safe_crypto_jwk_jwks_implementation_governance.sprint must be 35")
	status = governance.get("status")
	if status not in {"active", "completed"}:
		add_error("safe_crypto_jwk_jwks_implementation_governance.status must be active or completed")
	packages = require_string_list(governance.get("packages"), "safe_crypto_jwk_jwks_implementation_governance.packages")
	if packages != ["vcrypto"]:
		add_error("safe_crypto_jwk_jwks_implementation_governance.packages must be ordered as: vcrypto")
	internal_packages = require_string_list(governance.get("internal_packages"), "safe_crypto_jwk_jwks_implementation_governance.internal_packages")
	if internal_packages != ["internal/crypto"]:
		add_error("safe_crypto_jwk_jwks_implementation_governance.internal_packages must be ordered as: internal/crypto")
	required_functions = require_string_list(governance.get("required_functions"), "safe_crypto_jwk_jwks_implementation_governance.required_functions")
	expected_functions = [
		"RSAPublicKeyToJWK",
		"RSAPrivateKeyToJWK",
		"JWKToRSAPublicKey",
		"JWKToRSAPrivateKey",
		"MarshalJWK",
		"ParseJWK",
		"MarshalJWKS",
		"ParseJWKS",
		"SelectJWKByKeyID",
	]
	if required_functions != expected_functions:
		add_error("safe_crypto_jwk_jwks_implementation_governance.required_functions must be ordered as: " + ", ".join(expected_functions))
	for function_name in required_functions:
		if not reference_exists(f"vcrypto/jwk.go:{function_name}"):
			add_error(f"safe_crypto_jwk_jwks_implementation_governance.required_functions references missing facade function {function_name}")
	required_test_functions = require_string_list(
		governance.get("required_test_functions"),
		"safe_crypto_jwk_jwks_implementation_governance.required_test_functions",
	)
	expected_test_functions = [
		"internal/crypto/jwk_test.go:TestRSAJWKPublicRoundTrip",
		"internal/crypto/jwk_test.go:TestRSAJWKPrivateRoundTrip",
		"internal/crypto/jwk_test.go:TestJWKSSelectByKeyID",
		"internal/crypto/jwk_test.go:TestJWKMalformedErrors",
		"vcrypto/jwk_test.go:TestFacadeRSAJWKAndJWKS",
		"vcrypto/jwk_test.go:TestFacadeJWKErrors",
	]
	if required_test_functions != expected_test_functions:
		add_error("safe_crypto_jwk_jwks_implementation_governance.required_test_functions must be ordered as: " + ", ".join(expected_test_functions))
	for reference in required_test_functions:
		if not test_function_exists(reference):
			add_error(f"safe_crypto_jwk_jwks_implementation_governance.required_test_functions references missing test {reference}")
	required_examples = require_string_list(governance.get("required_examples"), "safe_crypto_jwk_jwks_implementation_governance.required_examples")
	expected_examples = ["ExampleRSAPublicKeyToJWK", "ExampleJWKToRSAPublicKey", "ExampleSelectJWKByKeyID"]
	if required_examples != expected_examples:
		add_error("safe_crypto_jwk_jwks_implementation_governance.required_examples must be ordered as: " + ", ".join(expected_examples))
	for example_name in required_examples:
		if not example_function_exists("vcrypto", example_name):
			add_error(f"safe_crypto_jwk_jwks_implementation_governance.required_examples references missing example {example_name}")
	required_checks = require_string_list(governance.get("required_checks"), "safe_crypto_jwk_jwks_implementation_governance.required_checks")
	for check in ("go test ./internal/crypto ./vcrypto", "tools-gen", "docs-check", "ai-context-check", "governance-maturity-check", "agent-security-check"):
		if check not in required_checks:
			add_error(f"safe_crypto_jwk_jwks_implementation_governance.required_checks must include {check}")
	for path in (quickstart_path, backlog_path):
		if not (root / path).exists():
			add_error(f"{path} must exist")
	quickstart_text = (root / quickstart_path).read_text(encoding="utf-8") if (root / quickstart_path).exists() else ""
	for phrase in ("RSAPublicKeyToJWK", "JWKToRSAPublicKey", "MarshalJWKS", "SelectJWKByKeyID", "remote JWKS"):
		if quickstart_text and phrase not in quickstart_text:
			add_error(f"{quickstart_path} must include {phrase!r}")
	backlog_text = (root / backlog_path).read_text(encoding="utf-8") if (root / backlog_path).exists() else ""
	for phrase in ("JWK and JWKS | Completed", "safe_crypto_jwk_jwks_implementation_governance", "RSA public/private JWK round trips", "unknown-`kid` behavior"):
		if backlog_text and phrase not in backlog_text:
			add_error(f"{backlog_path} must include {phrase!r}")

	sprint_rows = extract_markdown_rows(root / roadmap_path, "Sprint order")
	sprint_35_rows = [row for row in sprint_rows if row.get("Sprint") == "35"]
	if len(sprint_35_rows) != 1:
		add_error(f"{roadmap_path} Sprint order must contain exactly one Sprint 35 row")
	else:
		sprint_35 = sprint_35_rows[0]
		expected_status = "Completed" if status == "completed" else "Active"
		if sprint_35.get("Status") != expected_status:
			add_error(f"{roadmap_path} Sprint 35 status must be {expected_status}")
		sprint_text = " ".join(sprint_35.values())
		for required_phrase in ("RSA JWK/JWKS", "`kid`", "malformed-key", "no network discovery"):
			if required_phrase not in sprint_text:
				add_error(f"{roadmap_path} Sprint 35 row must mention {required_phrase!r}")


def validate_safe_crypto_secret_handling_governance() -> None:
	governance = require_mapping(ai_context.get("safe_crypto_secret_handling_governance"), "safe_crypto_secret_handling_governance")
	roadmap_path = governance.get("roadmap_path")
	if not isinstance(roadmap_path, str) or not roadmap_path.strip():
		add_error("safe_crypto_secret_handling_governance.roadmap_path must be non-empty")
		roadmap_path = "docs/superpowers/plans/49-roadmap.md"
	quickstart_path = governance.get("quickstart_path")
	if not isinstance(quickstart_path, str) or not quickstart_path.strip():
		add_error("safe_crypto_secret_handling_governance.quickstart_path must be non-empty")
		quickstart_path = "docs/doc/11-vcrypto.md"
	backlog_path = governance.get("backlog_path")
	if not isinstance(backlog_path, str) or not backlog_path.strip():
		add_error("safe_crypto_secret_handling_governance.backlog_path must be non-empty")
		backlog_path = "docs/doc/safe-crypto-advanced-backlog.md"
	sprint = governance.get("sprint")
	if sprint != 36:
		add_error("safe_crypto_secret_handling_governance.sprint must be 36")
	status = governance.get("status")
	if status not in {"active", "completed"}:
		add_error("safe_crypto_secret_handling_governance.status must be active or completed")
	packages = require_string_list(governance.get("packages"), "safe_crypto_secret_handling_governance.packages")
	expected_packages = ["vcrypto", "vrand", "vjwt", "vpass"]
	if packages != expected_packages:
		add_error("safe_crypto_secret_handling_governance.packages must be ordered as: " + ", ".join(expected_packages))
	for package_name in packages:
		if package_name not in public_facades:
			add_error(f"safe_crypto_secret_handling_governance.packages references non-public facade {package_name}")
	required_boundaries = require_string_list(governance.get("required_boundaries"), "safe_crypto_secret_handling_governance.required_boundaries")
	expected_boundaries = [
		"fixed secrets are demo-only fixtures",
		"production secret material uses secure random or key management",
		"deterministic readers are test-only injection points",
		"raw secrets are not logged",
		"private keys and encoded hashes are treated as secrets",
	]
	if required_boundaries != expected_boundaries:
		add_error("safe_crypto_secret_handling_governance.required_boundaries must be ordered as: " + ", ".join(expected_boundaries))
	required_doc_phrases = require_string_list(governance.get("required_doc_phrases"), "safe_crypto_secret_handling_governance.required_doc_phrases")
	expected_doc_phrases = [
		"demo-only fixtures",
		"vrand.SecureBytes",
		"vcrypto.RandomBytes",
		"deterministic readers",
		"not production defaults",
	]
	if required_doc_phrases != expected_doc_phrases:
		add_error("safe_crypto_secret_handling_governance.required_doc_phrases must be ordered as: " + ", ".join(expected_doc_phrases))
	required_checks = require_string_list(governance.get("required_checks"), "safe_crypto_secret_handling_governance.required_checks")
	for check in ("docs-check", "ai-context-check", "governance-maturity-check", "agent-security-check"):
		if check not in required_checks:
			add_error(f"safe_crypto_secret_handling_governance.required_checks must include {check}")
	for path in (quickstart_path, backlog_path):
		if not (root / path).exists():
			add_error(f"{path} must exist")
	quickstart_text = (root / quickstart_path).read_text(encoding="utf-8") if (root / quickstart_path).exists() else ""
	backlog_text = (root / backlog_path).read_text(encoding="utf-8") if (root / backlog_path).exists() else ""
	for phrase in required_doc_phrases:
		if quickstart_text and phrase not in quickstart_text:
			add_error(f"{quickstart_path} must include {phrase!r}")
	for phrase in ("Secret handling | Governance completed", "safe_crypto_secret_handling_governance", "demo-secret labeling", "random-source injection"):
		if backlog_text and phrase not in backlog_text:
			add_error(f"{backlog_path} must include {phrase!r}")

	sprint_rows = extract_markdown_rows(root / roadmap_path, "Sprint order")
	sprint_36_rows = [row for row in sprint_rows if row.get("Sprint") == "36"]
	if len(sprint_36_rows) != 1:
		add_error(f"{roadmap_path} Sprint order must contain exactly one Sprint 36 row")
	else:
		sprint_36 = sprint_36_rows[0]
		expected_status = "Completed" if status == "completed" else "Active"
		if sprint_36.get("Status") != expected_status:
			add_error(f"{roadmap_path} Sprint 36 status must be {expected_status}")
		sprint_text = " ".join(sprint_36.values())
		for required_phrase in ("demo secrets", "deterministic fixtures", "random-source injection", "secret-handling"):
			if required_phrase not in sprint_text:
				add_error(f"{roadmap_path} Sprint 36 row must mention {required_phrase!r}")


def validate_safe_crypto_interoperability_governance() -> None:
	governance = require_mapping(ai_context.get("safe_crypto_interoperability_governance"), "safe_crypto_interoperability_governance")
	roadmap_path = governance.get("roadmap_path")
	if not isinstance(roadmap_path, str) or not roadmap_path.strip():
		add_error("safe_crypto_interoperability_governance.roadmap_path must be non-empty")
		roadmap_path = "docs/superpowers/plans/49-roadmap.md"
	quickstart_path = governance.get("quickstart_path")
	if not isinstance(quickstart_path, str) or not quickstart_path.strip():
		add_error("safe_crypto_interoperability_governance.quickstart_path must be non-empty")
		quickstart_path = "docs/doc/11-vcrypto.md"
	backlog_path = governance.get("backlog_path")
	if not isinstance(backlog_path, str) or not backlog_path.strip():
		add_error("safe_crypto_interoperability_governance.backlog_path must be non-empty")
		backlog_path = "docs/doc/safe-crypto-advanced-backlog.md"
	sprint = governance.get("sprint")
	if sprint != 37:
		add_error("safe_crypto_interoperability_governance.sprint must be 37")
	status = governance.get("status")
	if status not in {"active", "completed"}:
		add_error("safe_crypto_interoperability_governance.status must be active or completed")
	packages = require_string_list(governance.get("packages"), "safe_crypto_interoperability_governance.packages")
	expected_packages = ["vcrypto", "vjwt"]
	if packages != expected_packages:
		add_error("safe_crypto_interoperability_governance.packages must be ordered as: " + ", ".join(expected_packages))
	for package_name in packages:
		if package_name not in public_facades:
			add_error(f"safe_crypto_interoperability_governance.packages references non-public facade {package_name}")
	required_boundaries = require_string_list(governance.get("required_boundaries"), "safe_crypto_interoperability_governance.required_boundaries")
	expected_boundaries = [
		"interoperability-only helpers are explicitly documented",
		"SM4-ECB is legacy and non-default",
		"SM2 UID policy is explicit at call sites",
		"RSA OAEP and PSS options are interoperability choices",
		"PEM and JWK exchange are key material helpers",
		"new designs prefer authenticated encryption",
	]
	if required_boundaries != expected_boundaries:
		add_error("safe_crypto_interoperability_governance.required_boundaries must be ordered as: " + ", ".join(expected_boundaries))
	required_doc_phrases = require_string_list(governance.get("required_doc_phrases"), "safe_crypto_interoperability_governance.required_doc_phrases")
	expected_doc_phrases = [
		"interoperability-only",
		"SM4-ECB",
		"SM2 UID policy",
		"RSA-OAEP/PSS",
		"PEM/JWK",
		"not the default recommendation",
	]
	if required_doc_phrases != expected_doc_phrases:
		add_error("safe_crypto_interoperability_governance.required_doc_phrases must be ordered as: " + ", ".join(expected_doc_phrases))
	required_checks = require_string_list(governance.get("required_checks"), "safe_crypto_interoperability_governance.required_checks")
	for check in ("docs-check", "ai-context-check", "governance-maturity-check", "agent-security-check"):
		if check not in required_checks:
			add_error(f"safe_crypto_interoperability_governance.required_checks must include {check}")
	for path in (quickstart_path, backlog_path):
		if not (root / path).exists():
			add_error(f"{path} must exist")
	quickstart_text = (root / quickstart_path).read_text(encoding="utf-8") if (root / quickstart_path).exists() else ""
	backlog_text = (root / backlog_path).read_text(encoding="utf-8") if (root / backlog_path).exists() else ""
	for phrase in required_doc_phrases:
		if quickstart_text and phrase not in quickstart_text:
			add_error(f"{quickstart_path} must include {phrase!r}")
	for phrase in ("Interoperability boundaries | Governance completed", "safe_crypto_interoperability_governance", "legacy-mode warnings", "SM4-ECB non-default"):
		if backlog_text and phrase not in backlog_text:
			add_error(f"{backlog_path} must include {phrase!r}")

	sprint_rows = extract_markdown_rows(root / roadmap_path, "Sprint order")
	sprint_37_rows = [row for row in sprint_rows if row.get("Sprint") == "37"]
	if len(sprint_37_rows) != 1:
		add_error(f"{roadmap_path} Sprint order must contain exactly one Sprint 37 row")
	else:
		sprint_37 = sprint_37_rows[0]
		expected_status = "Completed" if status == "completed" else "Active"
		if sprint_37.get("Status") != expected_status:
			add_error(f"{roadmap_path} Sprint 37 status must be {expected_status}")
		sprint_text = " ".join(sprint_37.values())
		for required_phrase in ("interoperability-only", "SM4-ECB", "SM2 UID", "non-default"):
			if required_phrase not in sprint_text:
				add_error(f"{roadmap_path} Sprint 37 row must mention {required_phrase!r}")


def validate_safe_crypto_benchmark_scope_governance() -> None:
	governance = require_mapping(ai_context.get("safe_crypto_benchmark_scope_governance"), "safe_crypto_benchmark_scope_governance")
	roadmap_path = governance.get("roadmap_path")
	if not isinstance(roadmap_path, str) or not roadmap_path.strip():
		add_error("safe_crypto_benchmark_scope_governance.roadmap_path must be non-empty")
		roadmap_path = "docs/superpowers/plans/49-roadmap.md"
	quickstart_path = governance.get("quickstart_path")
	if not isinstance(quickstart_path, str) or not quickstart_path.strip():
		add_error("safe_crypto_benchmark_scope_governance.quickstart_path must be non-empty")
		quickstart_path = "docs/doc/11-vcrypto.md"
	backlog_path = governance.get("backlog_path")
	if not isinstance(backlog_path, str) or not backlog_path.strip():
		add_error("safe_crypto_benchmark_scope_governance.backlog_path must be non-empty")
		backlog_path = "docs/doc/safe-crypto-advanced-backlog.md"
	sprint = governance.get("sprint")
	if sprint != 38:
		add_error("safe_crypto_benchmark_scope_governance.sprint must be 38")
	status = governance.get("status")
	if status not in {"active", "completed"}:
		add_error("safe_crypto_benchmark_scope_governance.status must be active or completed")
	packages = require_string_list(governance.get("packages"), "safe_crypto_benchmark_scope_governance.packages")
	if packages != ["vcrypto", "vrand"]:
		add_error("safe_crypto_benchmark_scope_governance.packages must be ordered as: vcrypto, vrand")
	allowlist = require_string_list(governance.get("quick_benchmark_allowlist"), "safe_crypto_benchmark_scope_governance.quick_benchmark_allowlist")
	expected_allowlist = [
		"BenchmarkSHA256Digest",
		"BenchmarkHMACSHA256Signing",
		"BenchmarkAESGCMEncrypt",
		"BenchmarkAESGCMDecrypt",
		"BenchmarkAESSealGCM",
		"BenchmarkHMACSHA256Hex",
		"BenchmarkSecureBytes",
	]
	if allowlist != expected_allowlist:
		add_error("safe_crypto_benchmark_scope_governance.quick_benchmark_allowlist must be ordered as: " + ", ".join(expected_allowlist))
	for benchmark_name in allowlist:
		found = any(
			reference_exists(f"{path}:{benchmark_name}")
			for path in ("vcrypto/crypto_benchmark_test.go", "vcrypto/aes_random_test.go", "vcrypto/digest_hmac_test.go", "vrand/rand_benchmark_test.go")
		)
		if not found:
			add_error(f"safe_crypto_benchmark_scope_governance.quick_benchmark_allowlist references missing benchmark {benchmark_name}")
	excluded = require_string_list(governance.get("excluded_from_quick_gates"), "safe_crypto_benchmark_scope_governance.excluded_from_quick_gates")
	expected_excluded = [
		"production-strength password hashing",
		"remote key discovery",
		"key rotation loops",
		"network-bound JWKS refresh",
	]
	if excluded != expected_excluded:
		add_error("safe_crypto_benchmark_scope_governance.excluded_from_quick_gates must be ordered as: " + ", ".join(expected_excluded))
	boundaries = require_string_list(governance.get("required_boundaries"), "safe_crypto_benchmark_scope_governance.required_boundaries")
	expected_boundaries = [
		"benchmark paths must be deterministic",
		"benchmark paths must be bounded for CI smoke",
		"password hashing benchmarks require explicit opt-in evidence",
		"benchmark output is evidence not a universal performance claim",
	]
	if boundaries != expected_boundaries:
		add_error("safe_crypto_benchmark_scope_governance.required_boundaries must be ordered as: " + ", ".join(expected_boundaries))
	required_checks = require_string_list(governance.get("required_checks"), "safe_crypto_benchmark_scope_governance.required_checks")
	for check in ("docs-check", "ai-context-check", "governance-maturity-check", "bench-regression-check"):
		if check not in required_checks:
			add_error(f"safe_crypto_benchmark_scope_governance.required_checks must include {check}")
	for path in (quickstart_path, backlog_path):
		if not (root / path).exists():
			add_error(f"{path} must exist")
	quickstart_text = (root / quickstart_path).read_text(encoding="utf-8") if (root / quickstart_path).exists() else ""
	backlog_text = (root / backlog_path).read_text(encoding="utf-8") if (root / backlog_path).exists() else ""
	for phrase in ("deterministic, bounded hot paths", "production-strength password hashing", "explicit opt-in evidence"):
		if quickstart_text and phrase not in quickstart_text:
			add_error(f"{quickstart_path} must include {phrase!r}")
	for phrase in ("Benchmark scope | Governance completed", "safe_crypto_benchmark_scope_governance", "deterministic crypto benchmark allowlists", "production-strength password hashing"):
		if backlog_text and phrase not in backlog_text:
			add_error(f"{backlog_path} must include {phrase!r}")

	sprint_rows = extract_markdown_rows(root / roadmap_path, "Sprint order")
	sprint_38_rows = [row for row in sprint_rows if row.get("Sprint") == "38"]
	if len(sprint_38_rows) != 1:
		add_error(f"{roadmap_path} Sprint order must contain exactly one Sprint 38 row")
	else:
		sprint_38 = sprint_38_rows[0]
		expected_status = "Completed" if status == "completed" else "Active"
		if sprint_38.get("Status") != expected_status:
			add_error(f"{roadmap_path} Sprint 38 status must be {expected_status}")
		sprint_text = " ".join(sprint_38.values())
		for required_phrase in ("deterministic", "password hashing", "quick gates", "secure-random"):
			if required_phrase not in sprint_text:
				add_error(f"{roadmap_path} Sprint 38 row must mention {required_phrase!r}")


def validate_utility_library_comparison_governance() -> None:
	governance = require_mapping(ai_context.get("utility_library_comparison_governance"), "utility_library_comparison_governance")
	roadmap_path = governance.get("roadmap_path")
	if not isinstance(roadmap_path, str) or not roadmap_path.strip():
		add_error("utility_library_comparison_governance.roadmap_path must be non-empty")
		roadmap_path = "docs/superpowers/plans/49-roadmap.md"
	doc_path = governance.get("doc_path")
	if not isinstance(doc_path, str) or not doc_path.strip():
		add_error("utility_library_comparison_governance.doc_path must be non-empty")
		doc_path = "docs/doc/utility-library-comparison.md"
	readme_path = governance.get("readme_path")
	if readme_path != "README.md":
		add_error("utility_library_comparison_governance.readme_path must be README.md")
		readme_path = "README.md"
	sprint = governance.get("sprint")
	if sprint != 39:
		add_error("utility_library_comparison_governance.sprint must be 39")
	status = governance.get("status")
	if status not in {"active", "completed"}:
		add_error("utility_library_comparison_governance.status must be active or completed")
	competitors = require_string_list(governance.get("competitors"), "utility_library_comparison_governance.competitors")
	expected_competitors = ["samber/lo", "duke-git/lancet", "thoas/go-funk", "gookit/goutil", "spf13/cast"]
	if competitors != expected_competitors:
		add_error("utility_library_comparison_governance.competitors must be ordered as: " + ", ".join(expected_competitors))
	required_boundaries = require_string_list(governance.get("required_boundaries"), "utility_library_comparison_governance.required_boundaries")
	expected_boundaries = [
		"stdlib first for short local code",
		"specialist libraries for narrow domains",
		"knifer-go for cross-domain workflows",
		"knifer-go for safety defaults and governance gates",
	]
	if required_boundaries != expected_boundaries:
		add_error("utility_library_comparison_governance.required_boundaries must be ordered as: " + ", ".join(expected_boundaries))
	required_checks = require_string_list(governance.get("required_checks"), "utility_library_comparison_governance.required_checks")
	for check in ("docs-check", "ai-context-check", "governance-maturity-check"):
		if check not in required_checks:
			add_error(f"utility_library_comparison_governance.required_checks must include {check}")
	for path in (doc_path, readme_path):
		if not (root / path).exists():
			add_error(f"{path} must exist")
	doc_text = (root / doc_path).read_text(encoding="utf-8") if (root / doc_path).exists() else ""
	readme_text = (root / readme_path).read_text(encoding="utf-8") if (root / readme_path).exists() else ""
	for competitor in competitors:
		if doc_text and competitor not in doc_text:
			add_error(f"{doc_path} must mention {competitor}")
		if readme_text and competitor not in readme_text:
			add_error(f"{readme_path} comparison table must mention {competitor}")
	for phrase in ("standard library", "specialist library", "cross-domain", "governance gates"):
		if doc_text and phrase not in doc_text:
			add_error(f"{doc_path} must include {phrase!r}")
	if readme_text and "utility-library-comparison.md" not in readme_text:
		add_error("README.md must link docs/doc/utility-library-comparison.md")

	sprint_rows = extract_markdown_rows(root / roadmap_path, "Sprint order")
	sprint_39_rows = [row for row in sprint_rows if row.get("Sprint") == "39"]
	if len(sprint_39_rows) != 1:
		add_error(f"{roadmap_path} Sprint order must contain exactly one Sprint 39 row")
	else:
		sprint_39 = sprint_39_rows[0]
		expected_status = "Completed" if status == "completed" else "Active"
		if sprint_39.get("Status") != expected_status:
			add_error(f"{roadmap_path} Sprint 39 status must be {expected_status}")
		sprint_text = " ".join(sprint_39.values())
		for required_phrase in ("samber/lo", "duke-git/lancet", "gookit/goutil", "spf13/cast"):
			if required_phrase not in sprint_text:
				add_error(f"{roadmap_path} Sprint 39 row must mention {required_phrase!r}")


def validate_utility_top5_comparison_governance_v2() -> None:
	governance = require_mapping(ai_context.get("utility_top5_comparison_governance_v2"), "utility_top5_comparison_governance_v2")
	roadmap_path = governance.get("roadmap_path")
	if not isinstance(roadmap_path, str) or not roadmap_path.strip():
		add_error("utility_top5_comparison_governance_v2.roadmap_path must be non-empty")
		roadmap_path = "docs/superpowers/plans/49-roadmap.md"
	doc_path = governance.get("doc_path")
	if not isinstance(doc_path, str) or not doc_path.strip():
		add_error("utility_top5_comparison_governance_v2.doc_path must be non-empty")
		doc_path = "docs/doc/utility-library-comparison.md"
	if governance.get("sprint") != 58:
		add_error("utility_top5_comparison_governance_v2.sprint must be 58")
	status = governance.get("status")
	if status not in {"active", "completed"}:
		add_error("utility_top5_comparison_governance_v2.status must be active or completed")
	if governance.get("last_checked") != "2026-07-02":
		add_error("utility_top5_comparison_governance_v2.last_checked must be 2026-07-02")
	top5 = require_string_list(governance.get("top5"), "utility_top5_comparison_governance_v2.top5")
	expected_top5 = ["samber/lo", "duke-git/lancet", "thoas/go-funk", "spf13/cast", "gookit/goutil"]
	if top5 != expected_top5:
		add_error("utility_top5_comparison_governance_v2.top5 must be ordered as: " + ", ".join(expected_top5))
	sections = require_string_list(governance.get("required_sections"), "utility_top5_comparison_governance_v2.required_sections")
	expected_sections = ["GitHub Top 5 Utility Libraries", "Comparison Matrix", "Decision Rules", "Gap Summary", "TODO Lanes", "Refresh Workflow", "Sources"]
	if sections != expected_sections:
		add_error("utility_top5_comparison_governance_v2.required_sections must be ordered as: " + ", ".join(expected_sections))
	paths = require_string_list(governance.get("required_paths"), "utility_top5_comparison_governance_v2.required_paths")
	expected_paths = [
		"collection-golden-paths.md",
		"collections-comparison.md",
		"vconv-cast-migration.md",
		"dynamic-data-toolkit-matrix.md",
		"daily-developer-utilities.md",
		"developer-debug-test-backlog.md",
		"facade-tiering.md",
		"benchmark-trust.md",
	]
	if paths != expected_paths:
		add_error("utility_top5_comparison_governance_v2.required_paths must be ordered as: " + ", ".join(expected_paths))
	required_checks = require_string_list(governance.get("required_checks"), "utility_top5_comparison_governance_v2.required_checks")
	for check in ("docs-check", "ai-context-check", "governance-maturity-check"):
		if check not in required_checks:
			add_error(f"utility_top5_comparison_governance_v2.required_checks must include {check}")
	if not (root / doc_path).exists():
		add_error(f"{doc_path} must exist")
	doc_text = (root / doc_path).read_text(encoding="utf-8") if (root / doc_path).exists() else ""
	for phrase in top5 + sections + paths + ["Last checked: 2026-07-02", "GitHub API", "Stars", "Last pushed"]:
		if doc_text and phrase not in doc_text:
			add_error(f"{doc_path} must include {phrase!r}")
	sprint_rows = extract_markdown_rows(root / roadmap_path, "Sprint order")
	sprint_58_rows = [row for row in sprint_rows if row.get("Sprint") == "58"]
	if len(sprint_58_rows) != 1:
		add_error(f"{roadmap_path} Sprint order must contain exactly one Sprint 58 row")
	else:
		expected_status = "Completed" if status == "completed" else "Active"
		if sprint_58_rows[0].get("Status") != expected_status:
			add_error(f"{roadmap_path} Sprint 58 status must be {expected_status}")
		sprint_text = " ".join(sprint_58_rows[0].values())
		for phrase in ("utility-library-comparison", "top5", "sources", "TODO"):
			if phrase not in sprint_text:
				add_error(f"{roadmap_path} Sprint 58 row must mention {phrase!r}")
	roadmap_text = (root / roadmap_path).read_text(encoding="utf-8") if (root / roadmap_path).exists() else ""
	if "utility_top5_comparison_governance_v2" not in roadmap_text:
		add_error(f"{roadmap_path} must mention utility_top5_comparison_governance_v2")


def validate_utility_top5_refresh_workflow_governance() -> None:
	governance = require_mapping(ai_context.get("utility_top5_refresh_workflow_governance"), "utility_top5_refresh_workflow_governance")
	roadmap_path = governance.get("roadmap_path")
	if not isinstance(roadmap_path, str) or not roadmap_path.strip():
		add_error("utility_top5_refresh_workflow_governance.roadmap_path must be non-empty")
		roadmap_path = "docs/superpowers/plans/49-roadmap.md"
	doc_path = governance.get("doc_path")
	if not isinstance(doc_path, str) or not doc_path.strip():
		add_error("utility_top5_refresh_workflow_governance.doc_path must be non-empty")
		doc_path = "docs/doc/utility-library-comparison.md"
	script_path = governance.get("script_path")
	if not isinstance(script_path, str) or not script_path.strip():
		add_error("utility_top5_refresh_workflow_governance.script_path must be non-empty")
		script_path = "bin/update_utility_comparison.py"
	make_target = governance.get("make_target")
	if make_target != "utility-comparison-refresh":
		add_error("utility_top5_refresh_workflow_governance.make_target must be utility-comparison-refresh")
		make_target = "utility-comparison-refresh"
	command_name = governance.get("command")
	if command_name != "utility_comparison_refresh":
		add_error("utility_top5_refresh_workflow_governance.command must be utility_comparison_refresh")
	if governance.get("sprint") != 64:
		add_error("utility_top5_refresh_workflow_governance.sprint must be 64")
	status = governance.get("status")
	if status not in {"active", "completed"}:
		add_error("utility_top5_refresh_workflow_governance.status must be active or completed")
	scope = require_string_list(governance.get("refresh_scope"), "utility_top5_refresh_workflow_governance.refresh_scope")
	expected_scope = ["samber/lo", "duke-git/lancet", "thoas/go-funk", "spf13/cast", "gookit/goutil"]
	if scope != expected_scope:
		add_error("utility_top5_refresh_workflow_governance.refresh_scope must be ordered as: " + ", ".join(expected_scope))
	if governance.get("explicit_only") is not True:
		add_error("utility_top5_refresh_workflow_governance.explicit_only must be true")
	if governance.get("network_required") is not True:
		add_error("utility_top5_refresh_workflow_governance.network_required must be true")
	default_gate_exclusions = require_string_list(governance.get("default_gate_exclusions"), "utility_top5_refresh_workflow_governance.default_gate_exclusions")
	expected_exclusions = ["docs-check", "quick-check", "agent-check", "ci-test"]
	if default_gate_exclusions != expected_exclusions:
		add_error("utility_top5_refresh_workflow_governance.default_gate_exclusions must be ordered as: " + ", ".join(expected_exclusions))
	updated_files = require_string_list(governance.get("updated_files"), "utility_top5_refresh_workflow_governance.updated_files")
	expected_updated_files = ["docs/doc/utility-library-comparison.md", "ai-context.json"]
	if updated_files != expected_updated_files:
		add_error("utility_top5_refresh_workflow_governance.updated_files must be ordered as: " + ", ".join(expected_updated_files))
	required_boundaries = require_string_list(governance.get("required_boundaries"), "utility_top5_refresh_workflow_governance.required_boundaries")
	expected_boundaries = [
		"do not run from ordinary docs-check",
		"do not depend on network in ordinary gates",
		"refresh writes comparison docs and AI metadata together",
		"GitHub API sources stay visible",
	]
	if required_boundaries != expected_boundaries:
		add_error("utility_top5_refresh_workflow_governance.required_boundaries must be ordered as: " + ", ".join(expected_boundaries))
	required_checks = require_string_list(governance.get("required_checks"), "utility_top5_refresh_workflow_governance.required_checks")
	for check in ("docs-check", "ai-context-check", "governance-maturity-check"):
		if check not in required_checks:
			add_error(f"utility_top5_refresh_workflow_governance.required_checks must include {check}")
	for path in (doc_path, script_path):
		if not (root / path).exists():
			add_error(f"{path} must exist")
	command = require_mapping(commands.get("utility_comparison_refresh"), "commands.utility_comparison_refresh")
	if command.get("cmd") != "make utility-comparison-refresh":
		add_error("commands.utility_comparison_refresh.cmd must be make utility-comparison-refresh")
	if command.get("safe_for_agent_auto_run") is not False:
		add_error("commands.utility_comparison_refresh.safe_for_agent_auto_run must be false")
	if command.get("requires_user_consent") is not True:
		add_error("commands.utility_comparison_refresh.requires_user_consent must be true")
	if command.get("writes_workspace") is not True:
		add_error("commands.utility_comparison_refresh.writes_workspace must be true")
	if command.get("network_required") is not True:
		add_error("commands.utility_comparison_refresh.network_required must be true")
	for path in updated_files:
		if path not in command.get("writes_files", []):
			add_error(f"commands.utility_comparison_refresh.writes_files must include {path}")
	if not re.search(rf"^{re.escape(make_target)}:(?:\s|$)", makefile, flags=re.MULTILINE):
		add_error(f"Makefile must define target {make_target}")
	else:
		target_match = re.search(rf"^{re.escape(make_target)}:(?:[^\n]*)\n(?P<body>(?:\t.*\n)+)", makefile, flags=re.MULTILINE)
		target_body = target_match.group("body") if target_match else ""
		if script_path not in target_body or "--write" not in target_body:
			add_error(f"Makefile target {make_target} must run {script_path} --write")
	for gate in default_gate_exclusions:
		if make_target_depends_on(gate, make_target):
			add_error(f"Makefile target {gate} must not depend on network refresh target {make_target}")
	script_text = (root / script_path).read_text(encoding="utf-8") if (root / script_path).exists() else ""
	for phrase in scope + ["api.github.com/repos", "--write", "Last checked", "utility_top5_comparison_governance_v2", "urllib.request"]:
		if script_text and phrase not in script_text:
			add_error(f"{script_path} must include {phrase!r}")
	doc_text = (root / doc_path).read_text(encoding="utf-8") if (root / doc_path).exists() else ""
	for phrase in scope + ["Last checked:", "GitHub API", "Keep this top5 comparison governed by current GitHub metadata"]:
		if doc_text and phrase not in doc_text:
			add_error(f"{doc_path} must include {phrase!r}")
	sprint_rows = extract_markdown_rows(root / roadmap_path, "Sprint order")
	sprint_64_rows = [row for row in sprint_rows if row.get("Sprint") == "64"]
	if len(sprint_64_rows) != 1:
		add_error(f"{roadmap_path} Sprint order must contain exactly one Sprint 64 row")
	else:
		expected_status = "Completed" if status == "completed" else "Active"
		if sprint_64_rows[0].get("Status") != expected_status:
			add_error(f"{roadmap_path} Sprint 64 status must be {expected_status}")
		sprint_text = " ".join(sprint_64_rows[0].values())
		for phrase in ("utility-comparison-refresh", "GitHub API", "explicit", "network"):
			if phrase not in sprint_text:
				add_error(f"{roadmap_path} Sprint 64 row must mention {phrase!r}")
	roadmap_text = (root / roadmap_path).read_text(encoding="utf-8") if (root / roadmap_path).exists() else ""
	if "utility_top5_refresh_workflow_governance" not in roadmap_text:
		add_error(f"{roadmap_path} must mention utility_top5_refresh_workflow_governance")


def validate_safe_crypto_advanced_closeout_governance() -> None:
	governance = require_mapping(ai_context.get("safe_crypto_advanced_closeout_governance"), "safe_crypto_advanced_closeout_governance")
	roadmap_path = governance.get("roadmap_path")
	if not isinstance(roadmap_path, str) or not roadmap_path.strip():
		add_error("safe_crypto_advanced_closeout_governance.roadmap_path must be non-empty")
		roadmap_path = "docs/superpowers/plans/49-roadmap.md"
	backlog_path = governance.get("backlog_path")
	if not isinstance(backlog_path, str) or not backlog_path.strip():
		add_error("safe_crypto_advanced_closeout_governance.backlog_path must be non-empty")
		backlog_path = "docs/doc/safe-crypto-advanced-backlog.md"
	if governance.get("sprint") != 40:
		add_error("safe_crypto_advanced_closeout_governance.sprint must be 40")
	status = governance.get("status")
	if status not in {"active", "completed"}:
		add_error("safe_crypto_advanced_closeout_governance.status must be active or completed")
	if governance.get("capability_row") != "Crypto":
		add_error("safe_crypto_advanced_closeout_governance.capability_row must be Crypto")
	required_landed = require_string_list(governance.get("required_landed_governance"), "safe_crypto_advanced_closeout_governance.required_landed_governance")
	expected_landed = [
		"safe_crypto_otp_governance",
		"safe_crypto_password_hashing_governance",
		"safe_crypto_argon2id_governance",
		"safe_crypto_jwk_jwks_governance",
		"safe_crypto_jwk_jwks_implementation_governance",
		"safe_crypto_secret_handling_governance",
		"safe_crypto_interoperability_governance",
		"safe_crypto_benchmark_scope_governance",
	]
	if required_landed != expected_landed:
		add_error("safe_crypto_advanced_closeout_governance.required_landed_governance must be ordered as: " + ", ".join(expected_landed))
	for name in required_landed:
		entry = require_mapping(ai_context.get(name), name)
		if entry.get("status") != "completed":
			add_error(f"{name}.status must be completed for safe crypto closeout")
	forbidden = require_string_list(governance.get("forbidden_roadmap_phrases"), "safe_crypto_advanced_closeout_governance.forbidden_roadmap_phrases")
	for check in ("docs-check", "ai-context-check", "governance-maturity-check"):
		if check not in require_string_list(governance.get("required_checks"), "safe_crypto_advanced_closeout_governance.required_checks"):
			add_error(f"safe_crypto_advanced_closeout_governance.required_checks must include {check}")
	for path in (roadmap_path, backlog_path):
		if not (root / path).exists():
			add_error(f"{path} must exist")
	backlog_text = (root / backlog_path).read_text(encoding="utf-8") if (root / backlog_path).exists() else ""
	capability_rows = extract_markdown_rows(root / roadmap_path, "Capability matrix")
	crypto_rows = [row for row in capability_rows if row.get("Area") == "Crypto"]
	if len(crypto_rows) != 1:
		add_error(f"{roadmap_path} Capability matrix must contain exactly one Crypto row")
	else:
		row_text = " ".join(crypto_rows[0].values())
		for phrase in forbidden:
			if phrase in row_text:
				add_error(f"{roadmap_path} Crypto row must not contain stale phrase {phrase!r}")
		for phrase in ("Advanced safe-crypto depth is completed", "safe_crypto_advanced_closeout_governance", "not broad crypto gap closure"):
			if phrase not in row_text:
				add_error(f"{roadmap_path} Crypto row must mention {phrase!r}")
	for phrase in ("TOTP and HOTP | Completed", "Password hashing | Completed", "JWK and JWKS | Completed", "Secret handling | Governance completed", "Interoperability boundaries | Governance completed", "Benchmark scope | Governance completed"):
		if backlog_text and phrase not in backlog_text:
			add_error(f"{backlog_path} must include landed lane {phrase!r}")
	sprint_rows = extract_markdown_rows(root / roadmap_path, "Sprint order")
	sprint_40_rows = [row for row in sprint_rows if row.get("Sprint") == "40"]
	if len(sprint_40_rows) != 1:
		add_error(f"{roadmap_path} Sprint order must contain exactly one Sprint 40 row")
	else:
		expected_status = "Completed" if status == "completed" else "Active"
		if sprint_40_rows[0].get("Status") != expected_status:
			add_error(f"{roadmap_path} Sprint 40 status must be {expected_status}")
		sprint_40_text = " ".join(sprint_40_rows[0].values())
		for phrase in forbidden:
			if phrase in sprint_40_text:
				add_error(f"{roadmap_path} Sprint 40 row must not contain stale phrase {phrase!r}")


def validate_go_version_adoption_governance() -> None:
	governance = require_mapping(ai_context.get("go_version_adoption_governance"), "go_version_adoption_governance")
	roadmap_path = governance.get("roadmap_path")
	if not isinstance(roadmap_path, str) or not roadmap_path.strip():
		add_error("go_version_adoption_governance.roadmap_path must be non-empty")
		roadmap_path = "docs/superpowers/plans/49-roadmap.md"
	doc_path = governance.get("doc_path")
	if not isinstance(doc_path, str) or not doc_path.strip():
		add_error("go_version_adoption_governance.doc_path must be non-empty")
		doc_path = "docs/doc/go-version-adoption-policy.md"
	readme_path = governance.get("readme_path")
	if readme_path != "README.md":
		add_error("go_version_adoption_governance.readme_path must be README.md")
		readme_path = "README.md"
	if governance.get("sprint") != 41:
		add_error("go_version_adoption_governance.sprint must be 41")
	status = governance.get("status")
	if status not in {"active", "completed"}:
		add_error("go_version_adoption_governance.status must be active or completed")
	if governance.get("minimum_go_version") != "1.25":
		add_error("go_version_adoption_governance.minimum_go_version must be 1.25")
	ci_versions = require_string_list(governance.get("ci_versions"), "go_version_adoption_governance.ci_versions")
	if ci_versions != ["1.25.11", "1.26"]:
		add_error("go_version_adoption_governance.ci_versions must be ordered as: 1.25.11, 1.26")
	if governance.get("release_go_version") != "1.25.11":
		add_error("go_version_adoption_governance.release_go_version must be 1.25.11")
	if governance.get("downgrade_status") != "not supported today":
		add_error("go_version_adoption_governance.downgrade_status must be not supported today")
	required_rationale = require_string_list(governance.get("required_rationale"), "go_version_adoption_governance.required_rationale")
	expected_rationale = [
		"go.mod declares go 1.25.0",
		"benchmarks use testing.B.Loop",
		"ci verifies go 1.25.11 and 1.26",
		"release workflow pins go 1.25.11",
	]
	if required_rationale != expected_rationale:
		add_error("go_version_adoption_governance.required_rationale must be ordered as: " + ", ".join(expected_rationale))
	required_checks = require_string_list(governance.get("required_checks"), "go_version_adoption_governance.required_checks")
	for check in ("docs-check", "ai-context-check", "ci-workflow-check", "governance-maturity-check"):
		if check not in required_checks:
			add_error(f"go_version_adoption_governance.required_checks must include {check}")
	for path in (roadmap_path, doc_path, readme_path):
		if not (root / path).exists():
			add_error(f"{path} must exist")
	go_mod_text = (root / "go.mod").read_text(encoding="utf-8")
	if "go 1.25.0" not in go_mod_text:
		add_error("go.mod must declare go 1.25.0")
	readme_text = (root / readme_path).read_text(encoding="utf-8") if (root / readme_path).exists() else ""
	if "go-version-adoption-policy.md" not in readme_text:
		add_error("README.md must link docs/doc/go-version-adoption-policy.md")
	doc_text = (root / doc_path).read_text(encoding="utf-8") if (root / doc_path).exists() else ""
	for phrase in ("Go 1.25", "Go 1.25.11", "Go 1.26", "Go 1.23/1.24 downgrade", "testing.B.Loop"):
		if doc_text and phrase not in doc_text:
			add_error(f"{doc_path} must include {phrase!r}")
	workflow_text = (root / ".github/workflows/go.yml").read_text(encoding="utf-8")
	for version in ("1.25.11", "1.26"):
		if version not in workflow_text:
			add_error(f".github/workflows/go.yml must include Go {version}")
	release_text = (root / ".github/workflows/release.yml").read_text(encoding="utf-8")
	if "1.25.11" not in release_text:
		add_error(".github/workflows/release.yml must pin Go 1.25.11")
	sprint_rows = extract_markdown_rows(root / roadmap_path, "Sprint order")
	sprint_41_rows = [row for row in sprint_rows if row.get("Sprint") == "41"]
	if len(sprint_41_rows) != 1:
		add_error(f"{roadmap_path} Sprint order must contain exactly one Sprint 41 row")
	else:
		expected_status = "Completed" if status == "completed" else "Active"
		if sprint_41_rows[0].get("Status") != expected_status:
			add_error(f"{roadmap_path} Sprint 41 status must be {expected_status}")


def validate_collections_comparison_governance() -> None:
	governance = require_mapping(ai_context.get("collections_comparison_governance"), "collections_comparison_governance")
	roadmap_path = governance.get("roadmap_path")
	if not isinstance(roadmap_path, str) or not roadmap_path.strip():
		add_error("collections_comparison_governance.roadmap_path must be non-empty")
		roadmap_path = "docs/superpowers/plans/49-roadmap.md"
	doc_path = governance.get("doc_path")
	if not isinstance(doc_path, str) or not doc_path.strip():
		add_error("collections_comparison_governance.doc_path must be non-empty")
		doc_path = "docs/doc/collections-comparison.md"
	if governance.get("sprint") != 42:
		add_error("collections_comparison_governance.sprint must be 42")
	status = governance.get("status")
	if status not in {"active", "completed"}:
		add_error("collections_comparison_governance.status must be active or completed")
	packages = require_string_list(governance.get("packages"), "collections_comparison_governance.packages")
	expected_packages = ["vslice", "vmap", "vset"]
	if packages != expected_packages:
		add_error("collections_comparison_governance.packages must be ordered as: " + ", ".join(expected_packages))
	competitors = require_string_list(governance.get("competitors"), "collections_comparison_governance.competitors")
	expected_competitors = ["samber/lo", "duke-git/lancet", "stdlib slices/maps"]
	if competitors != expected_competitors:
		add_error("collections_comparison_governance.competitors must be ordered as: " + ", ".join(expected_competitors))
	workflows = require_string_list(governance.get("required_workflows"), "collections_comparison_governance.required_workflows")
	expected_workflows = ["map", "filter", "reduce", "group", "partition", "window", "chunk", "set-like helpers"]
	if workflows != expected_workflows:
		add_error("collections_comparison_governance.required_workflows must be ordered as: " + ", ".join(expected_workflows))
	boundaries = require_string_list(governance.get("required_boundaries"), "collections_comparison_governance.required_boundaries")
	expected_boundaries = [
		"standard library first for local loops",
		"samber/lo for collection-only lodash-style helpers",
		"lancet for broad helper coverage",
		"knifer-go for cross-domain facade workflows",
		"error-returning helpers for fallible callbacks",
	]
	if boundaries != expected_boundaries:
		add_error("collections_comparison_governance.required_boundaries must be ordered as: " + ", ".join(expected_boundaries))
	required_checks = require_string_list(governance.get("required_checks"), "collections_comparison_governance.required_checks")
	for check in ("docs-check", "ai-context-check", "governance-maturity-check"):
		if check not in required_checks:
			add_error(f"collections_comparison_governance.required_checks must include {check}")
	if not (root / doc_path).exists():
		add_error(f"{doc_path} must exist")
	doc_text = (root / doc_path).read_text(encoding="utf-8") if (root / doc_path).exists() else ""
	for phrase in packages + competitors + workflows:
		if doc_text and phrase not in doc_text:
			add_error(f"{doc_path} must include {phrase!r}")
	for phrase in ("standard library", "fallible callbacks", "cross-domain", "Do not copy every helper"):
		if doc_text and phrase not in doc_text:
			add_error(f"{doc_path} must include {phrase!r}")
	readme_text = (root / "docs/doc/README.md").read_text(encoding="utf-8")
	if "collections-comparison.md" not in readme_text:
		add_error("docs/doc/README.md must link docs/doc/collections-comparison.md")
	sprint_rows = extract_markdown_rows(root / roadmap_path, "Sprint order")
	sprint_42_rows = [row for row in sprint_rows if row.get("Sprint") == "42"]
	if len(sprint_42_rows) != 1:
		add_error(f"{roadmap_path} Sprint order must contain exactly one Sprint 42 row")
	else:
		expected_status = "Completed" if status == "completed" else "Active"
		if sprint_42_rows[0].get("Status") != expected_status:
			add_error(f"{roadmap_path} Sprint 42 status must be {expected_status}")


def validate_collection_mindshare_pack_governance() -> None:
	governance = require_mapping(ai_context.get("collection_mindshare_pack_governance"), "collection_mindshare_pack_governance")
	roadmap_path = governance.get("roadmap_path")
	if not isinstance(roadmap_path, str) or not roadmap_path.strip():
		add_error("collection_mindshare_pack_governance.roadmap_path must be non-empty")
		roadmap_path = "docs/superpowers/plans/49-roadmap.md"
	doc_path = governance.get("doc_path")
	if not isinstance(doc_path, str) or not doc_path.strip():
		add_error("collection_mindshare_pack_governance.doc_path must be non-empty")
		doc_path = "docs/doc/collection-golden-paths.md"
	comparison_path = governance.get("comparison_path")
	if comparison_path != "docs/doc/collections-comparison.md":
		add_error("collection_mindshare_pack_governance.comparison_path must be docs/doc/collections-comparison.md")
		comparison_path = "docs/doc/collections-comparison.md"
	readme_path = governance.get("readme_path")
	if readme_path != "README.md":
		add_error("collection_mindshare_pack_governance.readme_path must be README.md")
		readme_path = "README.md"
	if governance.get("sprint") != 50:
		add_error("collection_mindshare_pack_governance.sprint must be 50")
	status = governance.get("status")
	if status not in {"active", "completed"}:
		add_error("collection_mindshare_pack_governance.status must be active or completed")
	packages = require_string_list(governance.get("packages"), "collection_mindshare_pack_governance.packages")
	expected_packages = ["vslice", "vmap", "vset"]
	if packages != expected_packages:
		add_error("collection_mindshare_pack_governance.packages must be ordered as: " + ", ".join(expected_packages))
	competitors = require_string_list(governance.get("competitors"), "collection_mindshare_pack_governance.competitors")
	expected_competitors = ["stdlib slices/maps", "samber/lo", "duke-git/lancet"]
	if competitors != expected_competitors:
		add_error("collection_mindshare_pack_governance.competitors must be ordered as: " + ", ".join(expected_competitors))
	tasks = require_string_list(governance.get("required_tasks"), "collection_mindshare_pack_governance.required_tasks")
	expected_tasks = ["map", "filter", "reduce", "group", "chunk", "window", "set", "zip", "partition", "find", "contains"]
	if tasks != expected_tasks:
		add_error("collection_mindshare_pack_governance.required_tasks must be ordered as: " + ", ".join(expected_tasks))
	baseline = require_mapping(governance.get("baseline_examples"), "collection_mindshare_pack_governance.baseline_examples")
	expected_baseline = {
		"vslice": (43, 43, 100.0),
		"vmap": (65, 65, 100.0),
		"vset": (10, 10, 100.0),
	}
	for facade in expected_packages:
		entry = require_mapping(baseline.get(facade), f"collection_mindshare_pack_governance.baseline_examples.{facade}")
		expected_function_count, expected_examples, expected_ratio = expected_baseline[facade]
		if entry.get("function_count") != expected_function_count:
			add_error(f"collection_mindshare_pack_governance.baseline_examples.{facade}.function_count must be {expected_function_count}")
		if entry.get("functions_with_examples") != expected_examples:
			add_error(f"collection_mindshare_pack_governance.baseline_examples.{facade}.functions_with_examples must be {expected_examples}")
		if entry.get("example_coverage_percent") != expected_ratio:
			add_error(f"collection_mindshare_pack_governance.baseline_examples.{facade}.example_coverage_percent must be {expected_ratio}")
		pkg = tool_packages.get(facade)
		if not pkg:
			add_error(f"docs/api/tools.json missing package {facade}")
			continue
		summary = require_mapping(pkg.get("summary"), f"docs/api/tools.json.packages.{facade}.summary")
		if summary.get("function_count") != expected_function_count:
			add_error(f"{facade} function count changed from governed baseline {expected_function_count} to {summary.get('function_count')}; update Sprint 50 governance deliberately")
		example_count = summary.get("functions_with_examples")
		if not isinstance(example_count, int) or isinstance(example_count, bool):
			add_error(f"docs/api/tools.json.packages.{facade}.summary.functions_with_examples must be an integer")
		elif example_count < expected_examples:
			add_error(f"{facade} examples regressed from Sprint 50 baseline {expected_examples} to {example_count}")
	boundaries = require_string_list(governance.get("required_boundaries"), "collection_mindshare_pack_governance.required_boundaries")
	expected_boundaries = [
		"workflow-first collection entry point",
		"standard library first for local loops",
		"samber/lo for collection-only lodash-style helpers",
		"lancet for broad helper coverage",
		"knifer-go for cross-domain facade workflows",
		"do not copy every helper from lo or lancet",
	]
	if boundaries != expected_boundaries:
		add_error("collection_mindshare_pack_governance.required_boundaries must be ordered as: " + ", ".join(expected_boundaries))
	required_checks = require_string_list(governance.get("required_checks"), "collection_mindshare_pack_governance.required_checks")
	for check in ("docs-check", "ai-context-check", "governance-maturity-check"):
		if check not in required_checks:
			add_error(f"collection_mindshare_pack_governance.required_checks must include {check}")
	if not (root / doc_path).exists():
		add_error(f"{doc_path} must exist")
	if not (root / comparison_path).exists():
		add_error(f"{comparison_path} must exist")
	doc_text = (root / doc_path).read_text(encoding="utf-8") if (root / doc_path).exists() else ""
	for phrase in packages + competitors + tasks + boundaries + ["Task Index", "Shortest `knifer-go` path", "collections-comparison.md"]:
		if doc_text and phrase not in doc_text:
			add_error(f"{doc_path} must include {phrase!r}")
	readme_text = (root / readme_path).read_text(encoding="utf-8") if (root / readme_path).exists() else ""
	if "collection-golden-paths.md" not in readme_text:
		add_error("README.md must link docs/doc/collection-golden-paths.md")
	doc_index_text = (root / "docs/doc/README.md").read_text(encoding="utf-8")
	if "collection-golden-paths.md" not in doc_index_text:
		add_error("docs/doc/README.md must link docs/doc/collection-golden-paths.md")
	sprint_rows = extract_markdown_rows(root / roadmap_path, "Sprint order")
	sprint_50_rows = [row for row in sprint_rows if row.get("Sprint") == "50"]
	if len(sprint_50_rows) != 1:
		add_error(f"{roadmap_path} Sprint order must contain exactly one Sprint 50 row")
	else:
		expected_status = "Completed" if status == "completed" else "Active"
		if sprint_50_rows[0].get("Status") != expected_status:
			add_error(f"{roadmap_path} Sprint 50 status must be {expected_status}")
		sprint_text = " ".join(sprint_50_rows[0].values())
		for phrase in ("collection-golden-paths", "vslice", "vmap", "vset", "samber/lo", "lancet"):
			if phrase not in sprint_text:
				add_error(f"{roadmap_path} Sprint 50 row must mention {phrase!r}")
	roadmap_text = (root / roadmap_path).read_text(encoding="utf-8") if (root / roadmap_path).exists() else ""
	if "collection_mindshare_pack_governance" not in roadmap_text:
		add_error(f"{roadmap_path} must mention collection_mindshare_pack_governance")


def validate_collection_advanced_backlog_governance() -> None:
	governance = require_mapping(ai_context.get("collection_advanced_backlog_governance"), "collection_advanced_backlog_governance")
	roadmap_path = governance.get("roadmap_path")
	if not isinstance(roadmap_path, str) or not roadmap_path.strip():
		add_error("collection_advanced_backlog_governance.roadmap_path must be non-empty")
		roadmap_path = "docs/superpowers/plans/49-roadmap.md"
	doc_path = governance.get("doc_path")
	if not isinstance(doc_path, str) or not doc_path.strip():
		add_error("collection_advanced_backlog_governance.doc_path must be non-empty")
		doc_path = "docs/doc/collection-advanced-backlog.md"
	readme_path = governance.get("readme_path")
	if readme_path != "README.md":
		add_error("collection_advanced_backlog_governance.readme_path must be README.md")
		readme_path = "README.md"
	if governance.get("sprint") != 60:
		add_error("collection_advanced_backlog_governance.sprint must be 60")
	status = governance.get("status")
	if status not in {"active", "completed"}:
		add_error("collection_advanced_backlog_governance.status must be active or completed")
	packages = require_string_list(governance.get("packages"), "collection_advanced_backlog_governance.packages")
	expected_packages = ["vslice", "vmap", "vset", "vjob"]
	if packages != expected_packages:
		add_error("collection_advanced_backlog_governance.packages must be ordered as: " + ", ".join(expected_packages))
	lanes = require_string_list(governance.get("candidate_lanes"), "collection_advanced_backlog_governance.candidate_lanes")
	expected_lanes = ["slice partition by predicate", "zip N", "cartesian product", "channel helpers", "parallel transforms", "iterator-first helpers"]
	if lanes != expected_lanes:
		add_error("collection_advanced_backlog_governance.candidate_lanes must be ordered as: " + ", ".join(expected_lanes))
	boundaries = require_string_list(governance.get("required_boundaries"), "collection_advanced_backlog_governance.required_boundaries")
	expected_boundaries = [
		"do not copy every helper from lo or lancet",
		"require an API decision card before implementation",
		"require executable examples before public API",
		"require benchmark evidence before allocation-heavy helpers",
		"keep error and cancellation contracts explicit",
	]
	if boundaries != expected_boundaries:
		add_error("collection_advanced_backlog_governance.required_boundaries must be ordered as: " + ", ".join(expected_boundaries))
	required_checks = require_string_list(governance.get("required_checks"), "collection_advanced_backlog_governance.required_checks")
	for check in ("docs-check", "ai-context-check", "governance-maturity-check"):
		if check not in required_checks:
			add_error(f"collection_advanced_backlog_governance.required_checks must include {check}")
	if not (root / doc_path).exists():
		add_error(f"{doc_path} must exist")
	doc_text = (root / doc_path).read_text(encoding="utf-8") if (root / doc_path).exists() else ""
	for phrase in packages + lanes + boundaries + ["Collection Advanced Backlog", "Candidate Lanes", "Required API Decision Card Questions"]:
		if doc_text and phrase not in doc_text:
			add_error(f"{doc_path} must include {phrase!r}")
	readme_text = (root / readme_path).read_text(encoding="utf-8") if (root / readme_path).exists() else ""
	if "collection-advanced-backlog.md" not in readme_text:
		add_error("README.md must link docs/doc/collection-advanced-backlog.md")
	doc_index_text = (root / "docs/doc/README.md").read_text(encoding="utf-8")
	if "collection-advanced-backlog.md" not in doc_index_text:
		add_error("docs/doc/README.md must link docs/doc/collection-advanced-backlog.md")
	sprint_rows = extract_markdown_rows(root / roadmap_path, "Sprint order")
	sprint_60_rows = [row for row in sprint_rows if row.get("Sprint") == "60"]
	if len(sprint_60_rows) != 1:
		add_error(f"{roadmap_path} Sprint order must contain exactly one Sprint 60 row")
	else:
		expected_status = "Completed" if status == "completed" else "Active"
		if sprint_60_rows[0].get("Status") != expected_status:
			add_error(f"{roadmap_path} Sprint 60 status must be {expected_status}")
		sprint_text = " ".join(sprint_60_rows[0].values())
		for phrase in ("collection-advanced-backlog", "partition", "zip", "parallel", "iterator"):
			if phrase not in sprint_text:
				add_error(f"{roadmap_path} Sprint 60 row must mention {phrase!r}")
	roadmap_text = (root / roadmap_path).read_text(encoding="utf-8") if (root / roadmap_path).exists() else ""
	if "collection_advanced_backlog_governance" not in roadmap_text:
		add_error(f"{roadmap_path} must mention collection_advanced_backlog_governance")


def validate_vconv_vbean_migration_governance() -> None:
	governance = require_mapping(ai_context.get("vconv_vbean_migration_governance"), "vconv_vbean_migration_governance")
	roadmap_path = governance.get("roadmap_path")
	if not isinstance(roadmap_path, str) or not roadmap_path.strip():
		add_error("vconv_vbean_migration_governance.roadmap_path must be non-empty")
		roadmap_path = "docs/superpowers/plans/49-roadmap.md"
	doc_path = governance.get("doc_path")
	if not isinstance(doc_path, str) or not doc_path.strip():
		add_error("vconv_vbean_migration_governance.doc_path must be non-empty")
		doc_path = "docs/doc/vconv-vbean-migration.md"
	if governance.get("sprint") != 43:
		add_error("vconv_vbean_migration_governance.sprint must be 43")
	status = governance.get("status")
	if status not in {"active", "completed"}:
		add_error("vconv_vbean_migration_governance.status must be active or completed")
	packages = require_string_list(governance.get("packages"), "vconv_vbean_migration_governance.packages")
	expected_packages = ["vconv", "vbean", "vconf"]
	if packages != expected_packages:
		add_error("vconv_vbean_migration_governance.packages must be ordered as: " + ", ".join(expected_packages))
	competitors = require_string_list(governance.get("competitors"), "vconv_vbean_migration_governance.competitors")
	expected_competitors = ["spf13/cast", "jinzhu/copier", "mitchellh/mapstructure", "mergo"]
	if competitors != expected_competitors:
		add_error("vconv_vbean_migration_governance.competitors must be ordered as: " + ", ".join(expected_competitors))
	workflows = require_string_list(governance.get("required_workflows"), "vconv_vbean_migration_governance.required_workflows")
	expected_workflows = ["strict conversion", "weak conversion", "copy", "decode", "merge", "unused metadata"]
	if workflows != expected_workflows:
		add_error("vconv_vbean_migration_governance.required_workflows must be ordered as: " + ", ".join(expected_workflows))
	required_checks = require_string_list(governance.get("required_checks"), "vconv_vbean_migration_governance.required_checks")
	for check in ("docs-check", "ai-context-check", "governance-maturity-check"):
		if check not in required_checks:
			add_error(f"vconv_vbean_migration_governance.required_checks must include {check}")
	if not (root / doc_path).exists():
		add_error(f"{doc_path} must exist")
	doc_text = (root / doc_path).read_text(encoding="utf-8") if (root / doc_path).exists() else ""
	for phrase in packages + competitors + workflows:
		if doc_text and phrase not in doc_text:
			add_error(f"{doc_path} must include {phrase!r}")
	for phrase in ("strict conversion", "weak conversion", "unused-key metadata", "copying every API"):
		if doc_text and phrase not in doc_text:
			add_error(f"{doc_path} must include {phrase!r}")
	readme_text = (root / "docs/doc/README.md").read_text(encoding="utf-8")
	if "vconv-vbean-migration.md" not in readme_text:
		add_error("docs/doc/README.md must link docs/doc/vconv-vbean-migration.md")
	sprint_rows = extract_markdown_rows(root / roadmap_path, "Sprint order")
	sprint_43_rows = [row for row in sprint_rows if row.get("Sprint") == "43"]
	if len(sprint_43_rows) != 1:
		add_error(f"{roadmap_path} Sprint order must contain exactly one Sprint 43 row")
	else:
		expected_status = "Completed" if status == "completed" else "Active"
		if sprint_43_rows[0].get("Status") != expected_status:
			add_error(f"{roadmap_path} Sprint 43 status must be {expected_status}")
	roadmap_text = (root / roadmap_path).read_text(encoding="utf-8") if (root / roadmap_path).exists() else ""
	if "vconv_vbean_migration_governance" not in roadmap_text:
		add_error(f"{roadmap_path} must mention vconv_vbean_migration_governance")


def validate_vconv_cast_migration_governance() -> None:
	governance = require_mapping(ai_context.get("vconv_cast_migration_governance"), "vconv_cast_migration_governance")
	roadmap_path = governance.get("roadmap_path")
	if not isinstance(roadmap_path, str) or not roadmap_path.strip():
		add_error("vconv_cast_migration_governance.roadmap_path must be non-empty")
		roadmap_path = "docs/superpowers/plans/49-roadmap.md"
	doc_path = governance.get("doc_path")
	if not isinstance(doc_path, str) or not doc_path.strip():
		add_error("vconv_cast_migration_governance.doc_path must be non-empty")
		doc_path = "docs/doc/vconv-cast-migration.md"
	readme_path = governance.get("readme_path")
	if readme_path != "README.md":
		add_error("vconv_cast_migration_governance.readme_path must be README.md")
		readme_path = "README.md"
	if governance.get("sprint") != 52:
		add_error("vconv_cast_migration_governance.sprint must be 52")
	status = governance.get("status")
	if status not in {"active", "completed"}:
		add_error("vconv_cast_migration_governance.status must be active or completed")
	if governance.get("package") != "vconv":
		add_error("vconv_cast_migration_governance.package must be vconv")
	if governance.get("competitor") != "spf13/cast":
		add_error("vconv_cast_migration_governance.competitor must be spf13/cast")
	workflows = require_string_list(governance.get("required_workflows"), "vconv_cast_migration_governance.required_workflows")
	expected_workflows = ["strict conversion", "weak conversion", "default fallback", "custom parser policy", "slice/map conversion", "duration/time conversion", "overflow handling"]
	if workflows != expected_workflows:
		add_error("vconv_cast_migration_governance.required_workflows must be ordered as: " + ", ".join(expected_workflows))
	boundaries = require_string_list(governance.get("required_boundaries"), "vconv_cast_migration_governance.required_boundaries")
	expected_boundaries = ["vconv is scalar-first", "E helpers at trust boundaries", "do not move collection conversion into vconv"]
	if boundaries != expected_boundaries:
		add_error("vconv_cast_migration_governance.required_boundaries must be ordered as: " + ", ".join(expected_boundaries))
	required_checks = require_string_list(governance.get("required_checks"), "vconv_cast_migration_governance.required_checks")
	for check in ("docs-check", "ai-context-check", "governance-maturity-check"):
		if check not in required_checks:
			add_error(f"vconv_cast_migration_governance.required_checks must include {check}")
	if not (root / doc_path).exists():
		add_error(f"{doc_path} must exist")
	doc_text = (root / doc_path).read_text(encoding="utf-8") if (root / doc_path).exists() else ""
	for phrase in workflows + boundaries + ["spf13/cast", "vconv", "Migration Rules", "Machine-Readable Boundaries"]:
		if doc_text and phrase not in doc_text:
			add_error(f"{doc_path} must include {phrase!r}")
	readme_text = (root / readme_path).read_text(encoding="utf-8") if (root / readme_path).exists() else ""
	if "vconv-cast-migration.md" not in readme_text:
		add_error("README.md must link docs/doc/vconv-cast-migration.md")
	doc_index_text = (root / "docs/doc/README.md").read_text(encoding="utf-8")
	if "vconv-cast-migration.md" not in doc_index_text:
		add_error("docs/doc/README.md must link docs/doc/vconv-cast-migration.md")
	sprint_rows = extract_markdown_rows(root / roadmap_path, "Sprint order")
	sprint_52_rows = [row for row in sprint_rows if row.get("Sprint") == "52"]
	if len(sprint_52_rows) != 1:
		add_error(f"{roadmap_path} Sprint order must contain exactly one Sprint 52 row")
	else:
		expected_status = "Completed" if status == "completed" else "Active"
		if sprint_52_rows[0].get("Status") != expected_status:
			add_error(f"{roadmap_path} Sprint 52 status must be {expected_status}")
		sprint_text = " ".join(sprint_52_rows[0].values())
		for phrase in ("vconv-cast-migration", "spf13/cast", "strict conversion", "weak conversion", "overflow"):
			if phrase not in sprint_text:
				add_error(f"{roadmap_path} Sprint 52 row must mention {phrase!r}")
	roadmap_text = (root / roadmap_path).read_text(encoding="utf-8") if (root / roadmap_path).exists() else ""
	if "vconv_cast_migration_governance" not in roadmap_text:
		add_error(f"{roadmap_path} must mention vconv_cast_migration_governance")


def validate_vconv_cast_migration_examples_governance() -> None:
	governance = require_mapping(ai_context.get("vconv_cast_migration_examples_governance"), "vconv_cast_migration_examples_governance")
	roadmap_path = governance.get("roadmap_path")
	if not isinstance(roadmap_path, str) or not roadmap_path.strip():
		add_error("vconv_cast_migration_examples_governance.roadmap_path must be non-empty")
		roadmap_path = "docs/superpowers/plans/49-roadmap.md"
	doc_path = governance.get("doc_path")
	if not isinstance(doc_path, str) or not doc_path.strip():
		add_error("vconv_cast_migration_examples_governance.doc_path must be non-empty")
		doc_path = "docs/doc/vconv-cast-migration.md"
	example_path = governance.get("example_path")
	if example_path != "vconv/example_test.go":
		add_error("vconv_cast_migration_examples_governance.example_path must be vconv/example_test.go")
		example_path = "vconv/example_test.go"
	if governance.get("sprint") != 61:
		add_error("vconv_cast_migration_examples_governance.sprint must be 61")
	status = governance.get("status")
	if status not in {"active", "completed"}:
		add_error("vconv_cast_migration_examples_governance.status must be active or completed")
	if governance.get("package") != "vconv":
		add_error("vconv_cast_migration_examples_governance.package must be vconv")
	if governance.get("competitor") != "spf13/cast":
		add_error("vconv_cast_migration_examples_governance.competitor must be spf13/cast")
	examples = require_string_list(governance.get("required_examples"), "vconv_cast_migration_examples_governance.required_examples")
	expected_examples = [
		"Example_castMigration_strictConversion",
		"Example_castMigration_weakConversion",
		"Example_castMigration_defaultFallback",
		"Example_castMigration_customParserPolicy",
		"Example_castMigration_sliceMapBoundary",
		"Example_castMigration_durationTimeBoundary",
		"Example_castMigration_overflowHandling",
	]
	if examples != expected_examples:
		add_error("vconv_cast_migration_examples_governance.required_examples must be ordered as: " + ", ".join(expected_examples))
	workflows = require_string_list(governance.get("required_workflows"), "vconv_cast_migration_examples_governance.required_workflows")
	expected_workflows = ["strict conversion", "weak conversion", "default fallback", "custom parser policy", "slice/map conversion", "duration/time conversion", "overflow handling"]
	if workflows != expected_workflows:
		add_error("vconv_cast_migration_examples_governance.required_workflows must be ordered as: " + ", ".join(expected_workflows))
	required_checks = require_string_list(governance.get("required_checks"), "vconv_cast_migration_examples_governance.required_checks")
	for check in ("go test ./vconv", "tools-gen", "docs-check", "ai-context-check", "governance-maturity-check"):
		if check not in required_checks:
			add_error(f"vconv_cast_migration_examples_governance.required_checks must include {check}")
	if not (root / doc_path).exists():
		add_error(f"{doc_path} must exist")
	if not (root / example_path).exists():
		add_error(f"{example_path} must exist")
	doc_text = (root / doc_path).read_text(encoding="utf-8") if (root / doc_path).exists() else ""
	example_text = (root / example_path).read_text(encoding="utf-8") if (root / example_path).exists() else ""
	for phrase in workflows + ["spf13/cast", "vconv"]:
		if doc_text and phrase not in doc_text:
			add_error(f"{doc_path} must include {phrase!r}")
	for example in examples:
		if example_text and f"func {example}()" not in example_text:
			add_error(f"{example_path} must include {example}")
	sprint_rows = extract_markdown_rows(root / roadmap_path, "Sprint order")
	sprint_61_rows = [row for row in sprint_rows if row.get("Sprint") == "61"]
	if len(sprint_61_rows) != 1:
		add_error(f"{roadmap_path} Sprint order must contain exactly one Sprint 61 row")
	else:
		expected_status = "Completed" if status == "completed" else "Active"
		if sprint_61_rows[0].get("Status") != expected_status:
			add_error(f"{roadmap_path} Sprint 61 status must be {expected_status}")
		sprint_text = " ".join(sprint_61_rows[0].values())
		for phrase in ("vconv", "cast", "examples", "strict", "overflow"):
			if phrase not in sprint_text:
				add_error(f"{roadmap_path} Sprint 61 row must mention {phrase!r}")
	roadmap_text = (root / roadmap_path).read_text(encoding="utf-8") if (root / roadmap_path).exists() else ""
	if "vconv_cast_migration_examples_governance" not in roadmap_text:
		add_error(f"{roadmap_path} must mention vconv_cast_migration_examples_governance")


def validate_dynamic_data_toolkit_matrix_governance() -> None:
	governance = require_mapping(ai_context.get("dynamic_data_toolkit_matrix_governance"), "dynamic_data_toolkit_matrix_governance")
	roadmap_path = governance.get("roadmap_path")
	if not isinstance(roadmap_path, str) or not roadmap_path.strip():
		add_error("dynamic_data_toolkit_matrix_governance.roadmap_path must be non-empty")
		roadmap_path = "docs/superpowers/plans/49-roadmap.md"
	doc_path = governance.get("doc_path")
	if not isinstance(doc_path, str) or not doc_path.strip():
		add_error("dynamic_data_toolkit_matrix_governance.doc_path must be non-empty")
		doc_path = "docs/doc/dynamic-data-toolkit-matrix.md"
	readme_path = governance.get("readme_path")
	if readme_path != "README.md":
		add_error("dynamic_data_toolkit_matrix_governance.readme_path must be README.md")
		readme_path = "README.md"
	if governance.get("sprint") != 53:
		add_error("dynamic_data_toolkit_matrix_governance.sprint must be 53")
	status = governance.get("status")
	if status not in {"active", "completed"}:
		add_error("dynamic_data_toolkit_matrix_governance.status must be active or completed")
	packages = require_string_list(governance.get("packages"), "dynamic_data_toolkit_matrix_governance.packages")
	expected_packages = ["vconf", "vbean", "vjson", "vobj", "vref", "vconv"]
	if packages != expected_packages:
		add_error("dynamic_data_toolkit_matrix_governance.packages must be ordered as: " + ", ".join(expected_packages))
	competitors = require_string_list(governance.get("competitors"), "dynamic_data_toolkit_matrix_governance.competitors")
	expected_competitors = ["thoas/go-funk", "mitchellh/mapstructure", "jinzhu/copier", "spf13/cast"]
	if competitors != expected_competitors:
		add_error("dynamic_data_toolkit_matrix_governance.competitors must be ordered as: " + ", ".join(expected_competitors))
	workflows = require_string_list(governance.get("required_workflows"), "dynamic_data_toolkit_matrix_governance.required_workflows")
	expected_workflows = ["configuration loading", "map/struct decode", "struct copy", "JSON object path", "dynamic object checks", "reflection field access", "scalar conversion after lookup"]
	if workflows != expected_workflows:
		add_error("dynamic_data_toolkit_matrix_governance.required_workflows must be ordered as: " + ", ".join(expected_workflows))
	boundaries = require_string_list(governance.get("required_boundaries"), "dynamic_data_toolkit_matrix_governance.required_boundaries")
	expected_boundaries = ["typed Go code first", "vconf before vbean for configuration input", "vbean before vobj for mapping metadata", "vref only at dynamic adapter boundaries", "vconv after dynamic lookup", "avoid reflection-heavy hot paths"]
	if boundaries != expected_boundaries:
		add_error("dynamic_data_toolkit_matrix_governance.required_boundaries must be ordered as: " + ", ".join(expected_boundaries))
	required_checks = require_string_list(governance.get("required_checks"), "dynamic_data_toolkit_matrix_governance.required_checks")
	for check in ("docs-check", "ai-context-check", "governance-maturity-check"):
		if check not in required_checks:
			add_error(f"dynamic_data_toolkit_matrix_governance.required_checks must include {check}")
	if not (root / doc_path).exists():
		add_error(f"{doc_path} must exist")
	doc_text = (root / doc_path).read_text(encoding="utf-8") if (root / doc_path).exists() else ""
	for phrase in packages + competitors + workflows + boundaries + ["Dynamic Data Toolkit Matrix", "Machine-Readable Boundaries"]:
		if doc_text and phrase not in doc_text:
			add_error(f"{doc_path} must include {phrase!r}")
	readme_text = (root / readme_path).read_text(encoding="utf-8") if (root / readme_path).exists() else ""
	if "dynamic-data-toolkit-matrix.md" not in readme_text:
		add_error("README.md must link docs/doc/dynamic-data-toolkit-matrix.md")
	doc_index_text = (root / "docs/doc/README.md").read_text(encoding="utf-8")
	if "dynamic-data-toolkit-matrix.md" not in doc_index_text:
		add_error("docs/doc/README.md must link docs/doc/dynamic-data-toolkit-matrix.md")
	sprint_rows = extract_markdown_rows(root / roadmap_path, "Sprint order")
	sprint_53_rows = [row for row in sprint_rows if row.get("Sprint") == "53"]
	if len(sprint_53_rows) != 1:
		add_error(f"{roadmap_path} Sprint order must contain exactly one Sprint 53 row")
	else:
		expected_status = "Completed" if status == "completed" else "Active"
		if sprint_53_rows[0].get("Status") != expected_status:
			add_error(f"{roadmap_path} Sprint 53 status must be {expected_status}")
		sprint_text = " ".join(sprint_53_rows[0].values())
		for phrase in ("dynamic-data-toolkit", "vconf", "vbean", "vjson", "vobj", "vref", "vconv"):
			if phrase not in sprint_text:
				add_error(f"{roadmap_path} Sprint 53 row must mention {phrase!r}")
	roadmap_text = (root / roadmap_path).read_text(encoding="utf-8") if (root / roadmap_path).exists() else ""
	if "dynamic_data_toolkit_matrix_governance" not in roadmap_text:
		add_error(f"{roadmap_path} must mention dynamic_data_toolkit_matrix_governance")


def validate_task_index_governance() -> None:
	governance = require_mapping(ai_context.get("task_index_governance"), "task_index_governance")
	roadmap_path = governance.get("roadmap_path")
	if not isinstance(roadmap_path, str) or not roadmap_path.strip():
		add_error("task_index_governance.roadmap_path must be non-empty")
		roadmap_path = "docs/superpowers/plans/49-roadmap.md"
	doc_path = governance.get("doc_path")
	if not isinstance(doc_path, str) or not doc_path.strip():
		add_error("task_index_governance.doc_path must be non-empty")
		doc_path = "docs/doc/task-index.md"
	readme_path = governance.get("readme_path")
	if readme_path != "README.md":
		add_error("task_index_governance.readme_path must be README.md")
		readme_path = "README.md"
	if governance.get("sprint") != 54:
		add_error("task_index_governance.sprint must be 54")
	status = governance.get("status")
	if status not in {"active", "completed"}:
		add_error("task_index_governance.status must be active or completed")
	day_one_tasks = require_string_list(governance.get("day_one_tasks"), "task_index_governance.day_one_tasks")
	expected_day_one_tasks = ["string cleanup", "slice transformation", "map transformation", "JSON path and formatting", "file IO", "safe HTTP", "crypto", "configuration", "database", "CLI command execution"]
	if day_one_tasks != expected_day_one_tasks:
		add_error("task_index_governance.day_one_tasks must be ordered as: " + ", ".join(expected_day_one_tasks))
	star_domains = require_string_list(governance.get("star_domains"), "task_index_governance.star_domains")
	expected_star_domains = ["Safe HTTP", "Safe Crypto", "Daily JSON/File"]
	if star_domains != expected_star_domains:
		add_error("task_index_governance.star_domains must be ordered as: " + ", ".join(expected_star_domains))
	daily_domains = require_string_list(governance.get("daily_domains"), "task_index_governance.daily_domains")
	expected_daily_domains = ["Daily developer utilities", "Collection workflows", "Dynamic data workflows"]
	if daily_domains != expected_daily_domains:
		add_error("task_index_governance.daily_domains must be ordered as: " + ", ".join(expected_daily_domains))
	default_facades = require_string_list(governance.get("default_facades"), "task_index_governance.default_facades")
	expected_default_facades = ["vstr", "vslice", "vmap", "vjson", "vfile", "vhttp", "vcrypto", "vconf", "vdb", "vcli"]
	if default_facades != expected_default_facades:
		add_error("task_index_governance.default_facades must be ordered as: " + ", ".join(expected_default_facades))
	boundaries = require_string_list(governance.get("required_boundaries"), "task_index_governance.required_boundaries")
	expected_boundaries = ["choose one default facade first", "related facades only when workflow crosses package boundaries", "Safe/E/WithOptions flows at trust boundaries", "do not import internal packages"]
	if boundaries != expected_boundaries:
		add_error("task_index_governance.required_boundaries must be ordered as: " + ", ".join(expected_boundaries))
	required_checks = require_string_list(governance.get("required_checks"), "task_index_governance.required_checks")
	for check in ("docs-check", "ai-context-check", "governance-maturity-check"):
		if check not in required_checks:
			add_error(f"task_index_governance.required_checks must include {check}")
	if not (root / doc_path).exists():
		add_error(f"{doc_path} must exist")
	doc_text = (root / doc_path).read_text(encoding="utf-8") if (root / doc_path).exists() else ""
	for phrase in day_one_tasks + star_domains + daily_domains + default_facades + boundaries + ["Task Index", "Day-One Tasks"]:
		if doc_text and phrase not in doc_text:
			add_error(f"{doc_path} must include {phrase!r}")
	readme_text = (root / readme_path).read_text(encoding="utf-8") if (root / readme_path).exists() else ""
	if "task-index.md" not in readme_text:
		add_error("README.md must link docs/doc/task-index.md")
	doc_index_text = (root / "docs/doc/README.md").read_text(encoding="utf-8")
	if "task-index.md" not in doc_index_text:
		add_error("docs/doc/README.md must link docs/doc/task-index.md")
	sprint_rows = extract_markdown_rows(root / roadmap_path, "Sprint order")
	sprint_54_rows = [row for row in sprint_rows if row.get("Sprint") == "54"]
	if len(sprint_54_rows) != 1:
		add_error(f"{roadmap_path} Sprint order must contain exactly one Sprint 54 row")
	else:
		expected_status = "Completed" if status == "completed" else "Active"
		if sprint_54_rows[0].get("Status") != expected_status:
			add_error(f"{roadmap_path} Sprint 54 status must be {expected_status}")
		sprint_text = " ".join(sprint_54_rows[0].values())
		for phrase in ("task-index", "day-one", "star-domain", "daily"):
			if phrase not in sprint_text:
				add_error(f"{roadmap_path} Sprint 54 row must mention {phrase!r}")
	roadmap_text = (root / roadmap_path).read_text(encoding="utf-8") if (root / roadmap_path).exists() else ""
	if "task_index_governance" not in roadmap_text:
		add_error(f"{roadmap_path} must mention task_index_governance")


def validate_task_index_auto_check_governance() -> None:
	governance = require_mapping(ai_context.get("task_index_auto_check_governance"), "task_index_auto_check_governance")
	roadmap_path = governance.get("roadmap_path")
	if not isinstance(roadmap_path, str) or not roadmap_path.strip():
		add_error("task_index_auto_check_governance.roadmap_path must be non-empty")
		roadmap_path = "docs/superpowers/plans/49-roadmap.md"
	doc_path = governance.get("doc_path")
	if not isinstance(doc_path, str) or not doc_path.strip():
		add_error("task_index_auto_check_governance.doc_path must be non-empty")
		doc_path = "docs/doc/task-index.md"
	if governance.get("sprint") != 63:
		add_error("task_index_auto_check_governance.sprint must be 63")
	status = governance.get("status")
	if status not in {"active", "completed"}:
		add_error("task_index_auto_check_governance.status must be active or completed")
	if governance.get("source_governance") != "task_index_governance":
		add_error("task_index_auto_check_governance.source_governance must be task_index_governance")
	if governance.get("tool_catalog_path") != "docs/api/tools.json":
		add_error("task_index_auto_check_governance.tool_catalog_path must be docs/api/tools.json")
	if governance.get("public_facades_source") != "public_facades":
		add_error("task_index_auto_check_governance.public_facades_source must be public_facades")
	if governance.get("star_domain_source") != "task_index_governance.star_domains":
		add_error("task_index_auto_check_governance.star_domain_source must be task_index_governance.star_domains")
	required_tables = require_string_list(governance.get("required_tables"), "task_index_auto_check_governance.required_tables")
	expected_tables = ["Day-One Tasks", "Star Domains", "Daily Domains"]
	if required_tables != expected_tables:
		add_error("task_index_auto_check_governance.required_tables must be ordered as: " + ", ".join(expected_tables))
	required_cross_checks = require_string_list(governance.get("required_cross_checks"), "task_index_auto_check_governance.required_cross_checks")
	expected_cross_checks = [
		"default facades exist in docs/api/tools.json",
		"default facades are public facades",
		"related facades are public facades",
		"star domain rows match ai-context star domains",
		"day-one default facades match task_index_governance",
	]
	if required_cross_checks != expected_cross_checks:
		add_error("task_index_auto_check_governance.required_cross_checks must be ordered as: " + ", ".join(expected_cross_checks))
	required_checks = require_string_list(governance.get("required_checks"), "task_index_auto_check_governance.required_checks")
	for check in ("docs-check", "ai-context-check", "governance-maturity-check"):
		if check not in required_checks:
			add_error(f"task_index_auto_check_governance.required_checks must include {check}")

	doc_file = root / doc_path
	if not doc_file.exists():
		add_error(f"{doc_path} must exist")
	task_index = require_mapping(ai_context.get("task_index_governance"), "task_index_governance")
	expected_day_one_defaults = require_string_list(task_index.get("default_facades"), "task_index_governance.default_facades")
	expected_star_domains = require_string_list(task_index.get("star_domains"), "task_index_governance.star_domains")
	if doc_file.exists():
		day_one_rows = extract_markdown_rows(doc_file, "Day-One Tasks")
		star_domain_rows = extract_markdown_rows(doc_file, "Star Domains")
		daily_domain_rows = extract_markdown_rows(doc_file, "Daily Domains")
	else:
		day_one_rows = []
		star_domain_rows = []
		daily_domain_rows = []
	for heading, rows in (
		("Day-One Tasks", day_one_rows),
		("Star Domains", star_domain_rows),
		("Daily Domains", daily_domain_rows),
	):
		if not rows:
			add_error(f"{doc_path} {heading} table must have at least one data row")
		for index, row in enumerate(rows, start=1):
			default_facades = extract_backticked_facades(row.get("Default facade", ""))
			if len(default_facades) != 1:
				add_error(f"{doc_path} {heading} row {index} must contain exactly one backticked Default facade")
				continue
			default_facade = default_facades[0]
			if default_facade not in public_facades:
				add_error(f"{doc_path} {heading} row {index} default facade {default_facade} is not in ai-context.public_facades")
			if default_facade not in tool_packages:
				add_error(f"{doc_path} {heading} row {index} default facade {default_facade} is missing from docs/api/tools.json")
			for related_facade in extract_backticked_facades(row.get("Related facades", "")):
				if related_facade not in public_facades:
					add_error(f"{doc_path} {heading} row {index} related facade {related_facade} is not in ai-context.public_facades")
	day_one_defaults = [extract_backticked_facades(row.get("Default facade", ""))[0] for row in day_one_rows if len(extract_backticked_facades(row.get("Default facade", ""))) == 1]
	if day_one_defaults != expected_day_one_defaults:
		add_error(f"{doc_path} Day-One Tasks default facades must match task_index_governance.default_facades")
	star_domain_names = [row.get("Domain", "") for row in star_domain_rows]
	if star_domain_names != expected_star_domains:
		add_error(f"{doc_path} Star Domains rows must match task_index_governance.star_domains")
	for row in star_domain_rows + daily_domain_rows:
		start_here = row.get("Start here", "")
		for link in re.findall(r"\]\(([^)]+)\)", start_here):
			if link.startswith(("http://", "https://", "#")):
				continue
			if not (doc_file.parent / link).exists():
				add_error(f"{doc_path} references missing Start here link {link}")
	sprint_rows = extract_markdown_rows(root / roadmap_path, "Sprint order")
	sprint_63_rows = [row for row in sprint_rows if row.get("Sprint") == "63"]
	if len(sprint_63_rows) != 1:
		add_error(f"{roadmap_path} Sprint order must contain exactly one Sprint 63 row")
	else:
		expected_status = "Completed" if status == "completed" else "Active"
		if sprint_63_rows[0].get("Status") != expected_status:
			add_error(f"{roadmap_path} Sprint 63 status must be {expected_status}")
		sprint_text = " ".join(sprint_63_rows[0].values())
		for phrase in ("task-index", "tools.json", "public facades", "star domains"):
			if phrase not in sprint_text:
				add_error(f"{roadmap_path} Sprint 63 row must mention {phrase!r}")
	roadmap_text = (root / roadmap_path).read_text(encoding="utf-8") if (root / roadmap_path).exists() else ""
	if "task_index_auto_check_governance" not in roadmap_text:
		add_error(f"{roadmap_path} must mention task_index_auto_check_governance")


def validate_facade_tiering_import_governance() -> None:
	governance = require_mapping(ai_context.get("facade_tiering_import_governance"), "facade_tiering_import_governance")
	roadmap_path = governance.get("roadmap_path")
	if not isinstance(roadmap_path, str) or not roadmap_path.strip():
		add_error("facade_tiering_import_governance.roadmap_path must be non-empty")
		roadmap_path = "docs/superpowers/plans/49-roadmap.md"
	doc_path = governance.get("doc_path")
	if not isinstance(doc_path, str) or not doc_path.strip():
		add_error("facade_tiering_import_governance.doc_path must be non-empty")
		doc_path = "docs/doc/facade-tiering.md"
	readme_path = governance.get("readme_path")
	if readme_path != "README.md":
		add_error("facade_tiering_import_governance.readme_path must be README.md")
		readme_path = "README.md"
	if governance.get("sprint") != 57:
		add_error("facade_tiering_import_governance.sprint must be 57")
	status = governance.get("status")
	if status not in {"active", "completed"}:
		add_error("facade_tiering_import_governance.status must be active or completed")
	day_one_defaults = require_string_list(governance.get("day_one_defaults"), "facade_tiering_import_governance.day_one_defaults")
	expected_day_one_defaults = ["vstr", "vslice", "vmap", "vjson", "vfile", "vhttp", "vcrypto", "vconf", "vdb", "vcli"]
	if day_one_defaults != expected_day_one_defaults:
		add_error("facade_tiering_import_governance.day_one_defaults must be ordered as: " + ", ".join(expected_day_one_defaults))
	if governance.get("dependency_tiers_source") != "dependency_tiers":
		add_error("facade_tiering_import_governance.dependency_tiers_source must be dependency_tiers")
	if governance.get("security_sensitive_source") != "security_sensitive_packages":
		add_error("facade_tiering_import_governance.security_sensitive_source must be security_sensitive_packages")
	required_tiers = require_string_list(governance.get("required_tiers"), "facade_tiering_import_governance.required_tiers")
	expected_tiers = ["core facades", "heavy extension facades", "provider contract facades", "security-sensitive overlay"]
	if required_tiers != expected_tiers:
		add_error("facade_tiering_import_governance.required_tiers must be ordered as: " + ", ".join(expected_tiers))
	boundaries = require_string_list(governance.get("required_boundaries"), "facade_tiering_import_governance.required_boundaries")
	expected_boundaries = [
		"import public v* facade packages",
		"do not import internal packages",
		"heavy dependencies require allowlist review",
		"provider contracts stay provider-neutral",
		"Safe/E/WithOptions flows in security-sensitive facades",
	]
	if boundaries != expected_boundaries:
		add_error("facade_tiering_import_governance.required_boundaries must be ordered as: " + ", ".join(expected_boundaries))
	required_checks = require_string_list(governance.get("required_checks"), "facade_tiering_import_governance.required_checks")
	for check in ("docs-check", "ai-context-check", "governance-maturity-check"):
		if check not in required_checks:
			add_error(f"facade_tiering_import_governance.required_checks must include {check}")
	dependency_tiers = require_mapping(ai_context.get("dependency_tiers"), "dependency_tiers")
	for tier_key in ("core_facades", "heavy_extension_facades", "provider_contract_facades"):
		tier_values = require_string_list(dependency_tiers.get(tier_key), f"dependency_tiers.{tier_key}")
		if not tier_values:
			add_error(f"dependency_tiers.{tier_key} must not be empty")
	security_sensitive = require_string_list(ai_context.get("security_sensitive_packages"), "security_sensitive_packages")
	if not security_sensitive:
		add_error("security_sensitive_packages must not be empty")
	if not (root / doc_path).exists():
		add_error(f"{doc_path} must exist")
	doc_text = (root / doc_path).read_text(encoding="utf-8") if (root / doc_path).exists() else ""
	for phrase in day_one_defaults + required_tiers + boundaries + ["dependency_tiers", "security_sensitive_packages", "Facade Tiering and Import Guide"]:
		if doc_text and phrase not in doc_text:
			add_error(f"{doc_path} must include {phrase!r}")
	for facade in dependency_tiers.get("heavy_extension_facades", []):
		if isinstance(facade, str) and doc_text and f"`{facade}`" not in doc_text:
			add_error(f"{doc_path} must include heavy extension facade {facade!r}")
	for facade in dependency_tiers.get("provider_contract_facades", []):
		if isinstance(facade, str) and doc_text and f"`{facade}`" not in doc_text:
			add_error(f"{doc_path} must include provider contract facade {facade!r}")
	readme_text = (root / readme_path).read_text(encoding="utf-8") if (root / readme_path).exists() else ""
	if "facade-tiering.md" not in readme_text:
		add_error("README.md must link docs/doc/facade-tiering.md")
	doc_index_text = (root / "docs/doc/README.md").read_text(encoding="utf-8")
	if "facade-tiering.md" not in doc_index_text:
		add_error("docs/doc/README.md must link docs/doc/facade-tiering.md")
	sprint_rows = extract_markdown_rows(root / roadmap_path, "Sprint order")
	sprint_57_rows = [row for row in sprint_rows if row.get("Sprint") == "57"]
	if len(sprint_57_rows) != 1:
		add_error(f"{roadmap_path} Sprint order must contain exactly one Sprint 57 row")
	else:
		expected_status = "Completed" if status == "completed" else "Active"
		if sprint_57_rows[0].get("Status") != expected_status:
			add_error(f"{roadmap_path} Sprint 57 status must be {expected_status}")
		sprint_text = " ".join(sprint_57_rows[0].values())
		for phrase in ("facade-tiering", "core", "heavy extension", "provider contract", "security-sensitive"):
			if phrase not in sprint_text:
				add_error(f"{roadmap_path} Sprint 57 row must mention {phrase!r}")
	roadmap_text = (root / roadmap_path).read_text(encoding="utf-8") if (root / roadmap_path).exists() else ""
	if "facade_tiering_import_governance" not in roadmap_text:
		add_error(f"{roadmap_path} must mention facade_tiering_import_governance")


def validate_facade_tiering_generated_view_governance() -> None:
	governance = require_mapping(ai_context.get("facade_tiering_generated_view_governance"), "facade_tiering_generated_view_governance")
	roadmap_path = governance.get("roadmap_path")
	if not isinstance(roadmap_path, str) or not roadmap_path.strip():
		add_error("facade_tiering_generated_view_governance.roadmap_path must be non-empty")
		roadmap_path = "docs/superpowers/plans/49-roadmap.md"
	doc_path = governance.get("doc_path")
	if not isinstance(doc_path, str) or not doc_path.strip():
		add_error("facade_tiering_generated_view_governance.doc_path must be non-empty")
		doc_path = "docs/doc/facade-tiering.md"
	script_path = governance.get("script_path")
	if not isinstance(script_path, str) or not script_path.strip():
		add_error("facade_tiering_generated_view_governance.script_path must be non-empty")
		script_path = "bin/generate_facade_tiering.py"
	if governance.get("make_target") != "facade-tiering-gen":
		add_error("facade_tiering_generated_view_governance.make_target must be facade-tiering-gen")
	if governance.get("command") != "facade_tiering_gen":
		add_error("facade_tiering_generated_view_governance.command must be facade_tiering_gen")
	if governance.get("sprint") != 65:
		add_error("facade_tiering_generated_view_governance.sprint must be 65")
	status = governance.get("status")
	if status not in {"active", "completed"}:
		add_error("facade_tiering_generated_view_governance.status must be active or completed")
	if governance.get("dependency_tiers_source") != "dependency_tiers":
		add_error("facade_tiering_generated_view_governance.dependency_tiers_source must be dependency_tiers")
	if governance.get("security_sensitive_source") != "security_sensitive_packages":
		add_error("facade_tiering_generated_view_governance.security_sensitive_source must be security_sensitive_packages")
	generated_markers = require_string_list(governance.get("generated_markers"), "facade_tiering_generated_view_governance.generated_markers")
	expected_markers = ["BEGIN GENERATED DEPENDENCY TIERS", "END GENERATED DEPENDENCY TIERS", "BEGIN GENERATED SECURITY OVERLAY", "END GENERATED SECURITY OVERLAY"]
	if generated_markers != expected_markers:
		add_error("facade_tiering_generated_view_governance.generated_markers must be ordered as: " + ", ".join(expected_markers))
	required_tiers = require_string_list(governance.get("required_tiers"), "facade_tiering_generated_view_governance.required_tiers")
	expected_tiers = ["core facades", "heavy extension facades", "provider contract facades", "security-sensitive overlay"]
	if required_tiers != expected_tiers:
		add_error("facade_tiering_generated_view_governance.required_tiers must be ordered as: " + ", ".join(expected_tiers))
	required_boundaries = require_string_list(governance.get("required_boundaries"), "facade_tiering_generated_view_governance.required_boundaries")
	expected_boundaries = [
		"dependency_tiers is the single source of truth",
		"security_sensitive_packages is the overlay source",
		"generated view must not be hand-edited between markers",
		"docs-check must stay network-free",
	]
	if required_boundaries != expected_boundaries:
		add_error("facade_tiering_generated_view_governance.required_boundaries must be ordered as: " + ", ".join(expected_boundaries))
	required_checks = require_string_list(governance.get("required_checks"), "facade_tiering_generated_view_governance.required_checks")
	for check in ("facade-tiering-gen", "docs-check", "ai-context-check", "governance-maturity-check"):
		if check not in required_checks:
			add_error(f"facade_tiering_generated_view_governance.required_checks must include {check}")
	for path in (doc_path, script_path):
		if not (root / path).exists():
			add_error(f"{path} must exist")
	command = require_mapping(commands.get("facade_tiering_gen"), "commands.facade_tiering_gen")
	if command.get("cmd") != "make facade-tiering-gen":
		add_error("commands.facade_tiering_gen.cmd must be make facade-tiering-gen")
	if command.get("safe_for_agent_auto_run") is not False:
		add_error("commands.facade_tiering_gen.safe_for_agent_auto_run must be false")
	if command.get("requires_user_consent") is not True:
		add_error("commands.facade_tiering_gen.requires_user_consent must be true")
	if command.get("writes_workspace") is not True:
		add_error("commands.facade_tiering_gen.writes_workspace must be true")
	if command.get("network_required") is not False:
		add_error("commands.facade_tiering_gen.network_required must be false")
	if doc_path not in command.get("writes_files", []):
		add_error(f"commands.facade_tiering_gen.writes_files must include {doc_path}")
	if not make_target_depends_on("docs-gen", "facade-tiering-gen"):
		add_error("Makefile target docs-gen must depend on facade-tiering-gen")
	if make_target_depends_on("docs-check", "facade-tiering-gen"):
		add_error("Makefile target docs-check must not write facade-tiering generated view")
	script_text = (root / script_path).read_text(encoding="utf-8") if (root / script_path).exists() else ""
	for phrase in ("dependency_tiers", "security_sensitive_packages", "BEGIN GENERATED DEPENDENCY TIERS", "BEGIN GENERATED SECURITY OVERLAY"):
		if script_text and phrase not in script_text:
			add_error(f"{script_path} must include {phrase!r}")
	doc_text = (root / doc_path).read_text(encoding="utf-8") if (root / doc_path).exists() else ""
	for marker in generated_markers:
		if doc_text and f"<!-- {marker} -->" not in doc_text:
			add_error(f"{doc_path} must include generated marker {marker!r}")
	dependency_tiers = require_mapping(ai_context.get("dependency_tiers"), "dependency_tiers")
	security_sensitive = require_string_list(ai_context.get("security_sensitive_packages"), "security_sensitive_packages")
	expected_dependency = expected_facade_tiering_dependency_table(dependency_tiers)
	actual_dependency = generated_block(doc_text, "BEGIN GENERATED DEPENDENCY TIERS", "END GENERATED DEPENDENCY TIERS") if doc_text else ""
	if actual_dependency and actual_dependency != expected_dependency:
		add_error(f"{doc_path} generated dependency tiers must match ai-context.json dependency_tiers")
	expected_security = expected_facade_tiering_security_table(security_sensitive)
	actual_security = generated_block(doc_text, "BEGIN GENERATED SECURITY OVERLAY", "END GENERATED SECURITY OVERLAY") if doc_text else ""
	if actual_security and actual_security != expected_security:
		add_error(f"{doc_path} generated security overlay must match ai-context.json security_sensitive_packages")
	sprint_rows = extract_markdown_rows(root / roadmap_path, "Sprint order")
	sprint_65_rows = [row for row in sprint_rows if row.get("Sprint") == "65"]
	if len(sprint_65_rows) != 1:
		add_error(f"{roadmap_path} Sprint order must contain exactly one Sprint 65 row")
	else:
		expected_status = "Completed" if status == "completed" else "Active"
		if sprint_65_rows[0].get("Status") != expected_status:
			add_error(f"{roadmap_path} Sprint 65 status must be {expected_status}")
		sprint_text = " ".join(sprint_65_rows[0].values())
		for phrase in ("facade-tiering", "generated", "dependency_tiers", "security_sensitive_packages"):
			if phrase not in sprint_text:
				add_error(f"{roadmap_path} Sprint 65 row must mention {phrase!r}")
	roadmap_text = (root / roadmap_path).read_text(encoding="utf-8") if (root / roadmap_path).exists() else ""
	if "facade_tiering_generated_view_governance" not in roadmap_text:
		add_error(f"{roadmap_path} must mention facade_tiering_generated_view_governance")


def validate_daily_developer_toolkit_governance() -> None:
	governance = require_mapping(ai_context.get("daily_developer_toolkit_governance"), "daily_developer_toolkit_governance")
	roadmap_path = governance.get("roadmap_path")
	if not isinstance(roadmap_path, str) or not roadmap_path.strip():
		add_error("daily_developer_toolkit_governance.roadmap_path must be non-empty")
		roadmap_path = "docs/superpowers/plans/49-roadmap.md"
	doc_path = governance.get("doc_path")
	if not isinstance(doc_path, str) or not doc_path.strip():
		add_error("daily_developer_toolkit_governance.doc_path must be non-empty")
		doc_path = "docs/doc/daily-developer-utilities.md"
	readme_path = governance.get("readme_path")
	if readme_path != "README.md":
		add_error("daily_developer_toolkit_governance.readme_path must be README.md")
		readme_path = "README.md"
	if governance.get("sprint") != 44:
		add_error("daily_developer_toolkit_governance.sprint must be 44")
	status = governance.get("status")
	if status not in {"active", "completed"}:
		add_error("daily_developer_toolkit_governance.status must be active or completed")
	packages = require_string_list(governance.get("packages"), "daily_developer_toolkit_governance.packages")
	expected_packages = ["vcli", "vsys", "vfile", "vnet", "vjob", "vlog"]
	if packages != expected_packages:
		add_error("daily_developer_toolkit_governance.packages must be ordered as: " + ", ".join(expected_packages))
	if governance.get("competitor") != "gookit/goutil":
		add_error("daily_developer_toolkit_governance.competitor must be gookit/goutil")
	planned_lanes = require_string_list(governance.get("planned_lanes"), "daily_developer_toolkit_governance.planned_lanes")
	if planned_lanes != ["vtest"]:
		add_error("daily_developer_toolkit_governance.planned_lanes must be vtest")
	workflows = require_string_list(governance.get("required_workflows"), "daily_developer_toolkit_governance.required_workflows")
	expected_workflows = [
		"CLI commands and terminal output",
		"System and runtime inspection",
		"File and IO tasks",
		"Network diagnostics",
		"Local job orchestration",
		"Logging while scripting",
	]
	if workflows != expected_workflows:
		add_error("daily_developer_toolkit_governance.required_workflows must be ordered as: " + ", ".join(expected_workflows))
	required_checks = require_string_list(governance.get("required_checks"), "daily_developer_toolkit_governance.required_checks")
	for check in ("docs-check", "ai-context-check", "governance-maturity-check"):
		if check not in required_checks:
			add_error(f"daily_developer_toolkit_governance.required_checks must include {check}")
	if not (root / doc_path).exists():
		add_error(f"{doc_path} must exist")
	doc_text = (root / doc_path).read_text(encoding="utf-8") if (root / doc_path).exists() else ""
	for phrase in packages + workflows + ["gookit/goutil", "vtest is a planned lane", "not a current public facade"]:
		if doc_text and phrase not in doc_text:
			add_error(f"{doc_path} must include {phrase!r}")
	readme_text = (root / readme_path).read_text(encoding="utf-8") if (root / readme_path).exists() else ""
	if "daily-developer-utilities.md" not in readme_text:
		add_error("README.md must link docs/doc/daily-developer-utilities.md")
	doc_index_text = (root / "docs/doc/README.md").read_text(encoding="utf-8")
	if "daily-developer-utilities.md" not in doc_index_text:
		add_error("docs/doc/README.md must link docs/doc/daily-developer-utilities.md")
	sprint_rows = extract_markdown_rows(root / roadmap_path, "Sprint order")
	sprint_44_rows = [row for row in sprint_rows if row.get("Sprint") == "44"]
	if len(sprint_44_rows) != 1:
		add_error(f"{roadmap_path} Sprint order must contain exactly one Sprint 44 row")
	else:
		expected_status = "Completed" if status == "completed" else "Active"
		if sprint_44_rows[0].get("Status") != expected_status:
			add_error(f"{roadmap_path} Sprint 44 status must be {expected_status}")
	roadmap_text = (root / roadmap_path).read_text(encoding="utf-8") if (root / roadmap_path).exists() else ""
	if "daily_developer_toolkit_governance" not in roadmap_text:
		add_error(f"{roadmap_path} must mention daily_developer_toolkit_governance")


def validate_daily_utility_cookbook_v2_governance() -> None:
	governance = require_mapping(ai_context.get("daily_utility_cookbook_v2_governance"), "daily_utility_cookbook_v2_governance")
	roadmap_path = governance.get("roadmap_path")
	if not isinstance(roadmap_path, str) or not roadmap_path.strip():
		add_error("daily_utility_cookbook_v2_governance.roadmap_path must be non-empty")
		roadmap_path = "docs/superpowers/plans/49-roadmap.md"
	doc_path = governance.get("doc_path")
	if not isinstance(doc_path, str) or not doc_path.strip():
		add_error("daily_utility_cookbook_v2_governance.doc_path must be non-empty")
		doc_path = "docs/doc/daily-developer-utilities.md"
	if governance.get("sprint") != 51:
		add_error("daily_utility_cookbook_v2_governance.sprint must be 51")
	status = governance.get("status")
	if status not in {"active", "completed"}:
		add_error("daily_utility_cookbook_v2_governance.status must be active or completed")
	packages = require_string_list(governance.get("packages"), "daily_utility_cookbook_v2_governance.packages")
	expected_packages = ["vcli", "vsys", "vfile", "vnet", "vjob", "vlog", "vconf", "vhttp"]
	if packages != expected_packages:
		add_error("daily_utility_cookbook_v2_governance.packages must be ordered as: " + ", ".join(expected_packages))
	if governance.get("competitor") != "gookit/goutil":
		add_error("daily_utility_cookbook_v2_governance.competitor must be gookit/goutil")
	workflows = require_string_list(governance.get("required_workflows"), "daily_utility_cookbook_v2_governance.required_workflows")
	expected_workflows = [
		"env-driven command execution",
		"config-backed file workflow",
		"network diagnostics report",
		"CLI support bundle",
		"local batch job runner",
		"filesystem cleanup preview",
		"lightweight service smoke script",
	]
	if workflows != expected_workflows:
		add_error("daily_utility_cookbook_v2_governance.required_workflows must be ordered as: " + ", ".join(expected_workflows))
	planned_lanes = require_string_list(governance.get("planned_lanes"), "daily_utility_cookbook_v2_governance.planned_lanes")
	if planned_lanes != ["vtest", "vdump"]:
		add_error("daily_utility_cookbook_v2_governance.planned_lanes must be vtest, vdump")
	boundaries = require_string_list(governance.get("required_boundaries"), "daily_utility_cookbook_v2_governance.required_boundaries")
	expected_boundaries = [
		"daily utilities should stay beside safety-focused facades",
		"no resident background utility process",
		"Safe/E/WithOptions flows for trust boundaries",
	]
	if boundaries != expected_boundaries:
		add_error("daily_utility_cookbook_v2_governance.required_boundaries must be ordered as: " + ", ".join(expected_boundaries))
	required_checks = require_string_list(governance.get("required_checks"), "daily_utility_cookbook_v2_governance.required_checks")
	for check in ("docs-check", "ai-context-check", "governance-maturity-check"):
		if check not in required_checks:
			add_error(f"daily_utility_cookbook_v2_governance.required_checks must include {check}")
	if not (root / doc_path).exists():
		add_error(f"{doc_path} must exist")
	doc_text = (root / doc_path).read_text(encoding="utf-8") if (root / doc_path).exists() else ""
	for phrase in packages + workflows + planned_lanes + boundaries + ["Cookbook", "gookit/goutil"]:
		if doc_text and phrase not in doc_text:
			add_error(f"{doc_path} must include {phrase!r}")
	sprint_rows = extract_markdown_rows(root / roadmap_path, "Sprint order")
	sprint_51_rows = [row for row in sprint_rows if row.get("Sprint") == "51"]
	if len(sprint_51_rows) != 1:
		add_error(f"{roadmap_path} Sprint order must contain exactly one Sprint 51 row")
	else:
		expected_status = "Completed" if status == "completed" else "Active"
		if sprint_51_rows[0].get("Status") != expected_status:
			add_error(f"{roadmap_path} Sprint 51 status must be {expected_status}")
		sprint_text = " ".join(sprint_51_rows[0].values())
		for phrase in ("daily-developer-utilities", "vcli", "vsys", "vfile", "vnet", "vjob", "vlog", "vconf"):
			if phrase not in sprint_text:
				add_error(f"{roadmap_path} Sprint 51 row must mention {phrase!r}")
	roadmap_text = (root / roadmap_path).read_text(encoding="utf-8") if (root / roadmap_path).exists() else ""
	if "daily_utility_cookbook_v2_governance" not in roadmap_text:
		add_error(f"{roadmap_path} must mention daily_utility_cookbook_v2_governance")


def validate_developer_debug_test_backlog_governance() -> None:
	governance = require_mapping(ai_context.get("developer_debug_test_backlog_governance"), "developer_debug_test_backlog_governance")
	roadmap_path = governance.get("roadmap_path")
	if not isinstance(roadmap_path, str) or not roadmap_path.strip():
		add_error("developer_debug_test_backlog_governance.roadmap_path must be non-empty")
		roadmap_path = "docs/superpowers/plans/49-roadmap.md"
	doc_path = governance.get("doc_path")
	if not isinstance(doc_path, str) or not doc_path.strip():
		add_error("developer_debug_test_backlog_governance.doc_path must be non-empty")
		doc_path = "docs/doc/developer-debug-test-backlog.md"
	readme_path = governance.get("readme_path")
	if readme_path != "README.md":
		add_error("developer_debug_test_backlog_governance.readme_path must be README.md")
		readme_path = "README.md"
	if governance.get("sprint") != 56:
		add_error("developer_debug_test_backlog_governance.sprint must be 56")
	status = governance.get("status")
	if status not in {"active", "completed"}:
		add_error("developer_debug_test_backlog_governance.status must be active or completed")
	current_facades = require_string_list(governance.get("current_facades"), "developer_debug_test_backlog_governance.current_facades")
	expected_current_facades = ["vcli", "vsys", "vfile", "vlog"]
	if current_facades != expected_current_facades:
		add_error("developer_debug_test_backlog_governance.current_facades must be ordered as: " + ", ".join(expected_current_facades))
	planned_lanes = require_string_list(governance.get("planned_lanes"), "developer_debug_test_backlog_governance.planned_lanes")
	if planned_lanes != ["vtest", "vdump"]:
		add_error("developer_debug_test_backlog_governance.planned_lanes must be vtest, vdump")
	candidate_scopes = require_string_list(governance.get("candidate_scopes"), "developer_debug_test_backlog_governance.candidate_scopes")
	expected_candidate_scopes = ["test helpers", "object dumps", "system dumps", "redaction hooks", "size limits", "golden file policy"]
	if candidate_scopes != expected_candidate_scopes:
		add_error("developer_debug_test_backlog_governance.candidate_scopes must be ordered as: " + ", ".join(expected_candidate_scopes))
	non_goals = require_string_list(governance.get("non_goals"), "developer_debug_test_backlog_governance.non_goals")
	expected_non_goals = ["replacing testing", "replacing testify", "resident background process", "secret-leaking dumps", "broad assertion framework replacement"]
	if non_goals != expected_non_goals:
		add_error("developer_debug_test_backlog_governance.non_goals must be ordered as: " + ", ".join(expected_non_goals))
	required_checks = require_string_list(governance.get("required_checks"), "developer_debug_test_backlog_governance.required_checks")
	for check in ("docs-check", "ai-context-check", "governance-maturity-check"):
		if check not in required_checks:
			add_error(f"developer_debug_test_backlog_governance.required_checks must include {check}")
	public_facades_value = ai_context.get("public_facades")
	if not isinstance(public_facades_value, list):
		add_error("public_facades must be a list")
		public_facades_value = []
	public_facade_names = {entry.get("package") for entry in public_facades_value if isinstance(entry, dict)}
	for planned in planned_lanes:
		if planned in public_facade_names:
			add_error(f"{planned} must not be listed as a current public facade while marked planned")
		if planned in tool_packages:
			add_error(f"docs/api/tools.json must not include planned facade {planned}")
	if not (root / doc_path).exists():
		add_error(f"{doc_path} must exist")
	doc_text = (root / doc_path).read_text(encoding="utf-8") if (root / doc_path).exists() else ""
	for phrase in current_facades + planned_lanes + candidate_scopes + non_goals + ["planned only", "not a current public facade", "Do not document `vtest` or `vdump` as available API"]:
		if doc_text and phrase not in doc_text:
			add_error(f"{doc_path} must include {phrase!r}")
	readme_text = (root / readme_path).read_text(encoding="utf-8") if (root / readme_path).exists() else ""
	if "developer-debug-test-backlog.md" not in readme_text:
		add_error("README.md must link docs/doc/developer-debug-test-backlog.md")
	doc_index_text = (root / "docs/doc/README.md").read_text(encoding="utf-8")
	if "developer-debug-test-backlog.md" not in doc_index_text:
		add_error("docs/doc/README.md must link docs/doc/developer-debug-test-backlog.md")
	sprint_rows = extract_markdown_rows(root / roadmap_path, "Sprint order")
	sprint_56_rows = [row for row in sprint_rows if row.get("Sprint") == "56"]
	if len(sprint_56_rows) != 1:
		add_error(f"{roadmap_path} Sprint order must contain exactly one Sprint 56 row")
	else:
		expected_status = "Completed" if status == "completed" else "Active"
		if sprint_56_rows[0].get("Status") != expected_status:
			add_error(f"{roadmap_path} Sprint 56 status must be {expected_status}")
		sprint_text = " ".join(sprint_56_rows[0].values())
		for phrase in ("developer-debug-test-backlog", "vtest", "vdump", "planned"):
			if phrase not in sprint_text:
				add_error(f"{roadmap_path} Sprint 56 row must mention {phrase!r}")
	roadmap_text = (root / roadmap_path).read_text(encoding="utf-8") if (root / roadmap_path).exists() else ""
	if "developer_debug_test_backlog_governance" not in roadmap_text:
		add_error(f"{roadmap_path} must mention developer_debug_test_backlog_governance")


def validate_developer_debug_test_api_decision_governance() -> None:
	governance = require_mapping(ai_context.get("developer_debug_test_api_decision_governance"), "developer_debug_test_api_decision_governance")
	roadmap_path = governance.get("roadmap_path")
	if not isinstance(roadmap_path, str) or not roadmap_path.strip():
		add_error("developer_debug_test_api_decision_governance.roadmap_path must be non-empty")
		roadmap_path = "docs/superpowers/plans/49-roadmap.md"
	doc_path = governance.get("doc_path")
	if not isinstance(doc_path, str) or not doc_path.strip():
		add_error("developer_debug_test_api_decision_governance.doc_path must be non-empty")
		doc_path = "docs/doc/developer-debug-test-backlog.md"
	if governance.get("sprint") != 62:
		add_error("developer_debug_test_api_decision_governance.sprint must be 62")
	status = governance.get("status")
	if status not in {"active", "completed"}:
		add_error("developer_debug_test_api_decision_governance.status must be active or completed")
	planned_lanes = require_string_list(governance.get("planned_lanes"), "developer_debug_test_api_decision_governance.planned_lanes")
	if planned_lanes != ["vtest", "vdump"]:
		add_error("developer_debug_test_api_decision_governance.planned_lanes must be vtest, vdump")
	decisions = require_string_list(governance.get("required_decisions"), "developer_debug_test_api_decision_governance.required_decisions")
	expected_decisions = [
		"repeated workflows across three or more facades",
		"safe issue report dumps require redaction hooks and size limits",
		"use explicit callbacks first for redaction",
		"keep golden-file helpers package-local until three facades need them",
		"reject broad assertion framework replacement",
		"reject broad logging replacement",
	]
	if decisions != expected_decisions:
		add_error("developer_debug_test_api_decision_governance.required_decisions must be ordered as: " + ", ".join(expected_decisions))
	candidate_cards = require_string_list(governance.get("candidate_cards"), "developer_debug_test_api_decision_governance.candidate_cards")
	expected_cards = ["vtest fixture helpers", "vtest assertion helpers", "vdump object dump helpers", "vdump system dump helpers"]
	if candidate_cards != expected_cards:
		add_error("developer_debug_test_api_decision_governance.candidate_cards must be ordered as: " + ", ".join(expected_cards))
	required_checks = require_string_list(governance.get("required_checks"), "developer_debug_test_api_decision_governance.required_checks")
	for check in ("docs-check", "ai-context-check", "governance-maturity-check"):
		if check not in required_checks:
			add_error(f"developer_debug_test_api_decision_governance.required_checks must include {check}")
	if not (root / doc_path).exists():
		add_error(f"{doc_path} must exist")
	doc_text = (root / doc_path).read_text(encoding="utf-8") if (root / doc_path).exists() else ""
	for phrase in planned_lanes + decisions + candidate_cards + ["API Decision Backlog v2", "Candidate API Cards"]:
		if doc_text and phrase not in doc_text:
			add_error(f"{doc_path} must include {phrase!r}")
	public_facades_value = ai_context.get("public_facades")
	if not isinstance(public_facades_value, list):
		add_error("public_facades must be a list")
		public_facades_value = []
	public_facade_names = {entry.get("package") for entry in public_facades_value if isinstance(entry, dict)}
	for planned in planned_lanes:
		if planned in public_facade_names:
			add_error(f"{planned} must not be listed as a current public facade while marked planned")
		if planned in tool_packages:
			add_error(f"docs/api/tools.json must not include planned facade {planned}")
	sprint_rows = extract_markdown_rows(root / roadmap_path, "Sprint order")
	sprint_62_rows = [row for row in sprint_rows if row.get("Sprint") == "62"]
	if len(sprint_62_rows) != 1:
		add_error(f"{roadmap_path} Sprint order must contain exactly one Sprint 62 row")
	else:
		expected_status = "Completed" if status == "completed" else "Active"
		if sprint_62_rows[0].get("Status") != expected_status:
			add_error(f"{roadmap_path} Sprint 62 status must be {expected_status}")
		sprint_text = " ".join(sprint_62_rows[0].values())
		for phrase in ("developer-debug-test", "vtest", "vdump", "API decision"):
			if phrase not in sprint_text:
				add_error(f"{roadmap_path} Sprint 62 row must mention {phrase!r}")
	roadmap_text = (root / roadmap_path).read_text(encoding="utf-8") if (root / roadmap_path).exists() else ""
	if "developer_debug_test_api_decision_governance" not in roadmap_text:
		add_error(f"{roadmap_path} must mention developer_debug_test_api_decision_governance")


def validate_benchmark_trust_governance() -> None:
	governance = require_mapping(ai_context.get("benchmark_trust_governance"), "benchmark_trust_governance")
	roadmap_path = governance.get("roadmap_path")
	if not isinstance(roadmap_path, str) or not roadmap_path.strip():
		add_error("benchmark_trust_governance.roadmap_path must be non-empty")
		roadmap_path = "docs/superpowers/plans/49-roadmap.md"
	doc_path = governance.get("doc_path")
	if not isinstance(doc_path, str) or not doc_path.strip():
		add_error("benchmark_trust_governance.doc_path must be non-empty")
		doc_path = "docs/doc/benchmark-trust.md"
	readme_path = governance.get("readme_path")
	if readme_path != "README.md":
		add_error("benchmark_trust_governance.readme_path must be README.md")
		readme_path = "README.md"
	if governance.get("sprint") != 45:
		add_error("benchmark_trust_governance.sprint must be 45")
	status = governance.get("status")
	if status not in {"active", "completed"}:
		add_error("benchmark_trust_governance.status must be active or completed")
	quick_gates = require_string_list(governance.get("quick_gates"), "benchmark_trust_governance.quick_gates")
	expected_quick_gates = ["make bench-smoke", "make bench-regression-check"]
	if quick_gates != expected_quick_gates:
		add_error("benchmark_trust_governance.quick_gates must be ordered as: " + ", ".join(expected_quick_gates))
	manual_opt_in = require_string_list(governance.get("manual_opt_in"), "benchmark_trust_governance.manual_opt_in")
	expected_manual_opt_in = ["make bench-core", "make bench-facade", "make bench-codec", "make bench-baseline", "make bench-compare"]
	if manual_opt_in != expected_manual_opt_in:
		add_error("benchmark_trust_governance.manual_opt_in must be ordered as: " + ", ".join(expected_manual_opt_in))
	boundaries = require_string_list(governance.get("required_boundaries"), "benchmark_trust_governance.required_boundaries")
	expected_boundaries = [
		"benchmark output is local baseline evidence",
		"performance claims require repeated runs and benchstat",
		"quick gates are deterministic and bounded",
		"long or workload-specific benchmarks are manual opt-in",
		"do not publish universal performance claims",
	]
	if boundaries != expected_boundaries:
		add_error("benchmark_trust_governance.required_boundaries must be ordered as: " + ", ".join(expected_boundaries))
	required_checks = require_string_list(governance.get("required_checks"), "benchmark_trust_governance.required_checks")
	for check in ("docs-check", "ai-context-check", "governance-maturity-check", "bench-regression-check"):
		if check not in required_checks:
			add_error(f"benchmark_trust_governance.required_checks must include {check}")
	if not (root / doc_path).exists():
		add_error(f"{doc_path} must exist")
	doc_text = (root / doc_path).read_text(encoding="utf-8") if (root / doc_path).exists() else ""
	for phrase in quick_gates + manual_opt_in + boundaries + ["Quick Gates", "Manual Opt-In Evidence"]:
		if doc_text and phrase not in doc_text:
			add_error(f"{doc_path} must include {phrase!r}")
	readme_text = (root / readme_path).read_text(encoding="utf-8") if (root / readme_path).exists() else ""
	if "benchmark-trust.md" not in readme_text:
		add_error("README.md must link docs/doc/benchmark-trust.md")
	doc_index_text = (root / "docs/doc/README.md").read_text(encoding="utf-8")
	if "benchmark-trust.md" not in doc_index_text:
		add_error("docs/doc/README.md must link docs/doc/benchmark-trust.md")
	sprint_rows = extract_markdown_rows(root / roadmap_path, "Sprint order")
	sprint_45_rows = [row for row in sprint_rows if row.get("Sprint") == "45"]
	if len(sprint_45_rows) != 1:
		add_error(f"{roadmap_path} Sprint order must contain exactly one Sprint 45 row")
	else:
		expected_status = "Completed" if status == "completed" else "Active"
		if sprint_45_rows[0].get("Status") != expected_status:
			add_error(f"{roadmap_path} Sprint 45 status must be {expected_status}")
	roadmap_text = (root / roadmap_path).read_text(encoding="utf-8") if (root / roadmap_path).exists() else ""
	if "benchmark_trust_governance" not in roadmap_text:
		add_error(f"{roadmap_path} must mention benchmark_trust_governance")


def validate_collections_benchmark_trust_governance() -> None:
	governance = require_mapping(ai_context.get("collections_benchmark_trust_governance"), "collections_benchmark_trust_governance")
	roadmap_path = governance.get("roadmap_path")
	if not isinstance(roadmap_path, str) or not roadmap_path.strip():
		add_error("collections_benchmark_trust_governance.roadmap_path must be non-empty")
		roadmap_path = "docs/superpowers/plans/49-roadmap.md"
	doc_path = governance.get("doc_path")
	if not isinstance(doc_path, str) or not doc_path.strip():
		add_error("collections_benchmark_trust_governance.doc_path must be non-empty")
		doc_path = "docs/doc/benchmark-trust.md"
	if governance.get("sprint") != 55:
		add_error("collections_benchmark_trust_governance.sprint must be 55")
	status = governance.get("status")
	if status not in {"active", "completed"}:
		add_error("collections_benchmark_trust_governance.status must be active or completed")
	packages = require_string_list(governance.get("packages"), "collections_benchmark_trust_governance.packages")
	expected_packages = ["vslice", "vmap", "vset"]
	if packages != expected_packages:
		add_error("collections_benchmark_trust_governance.packages must be ordered as: " + ", ".join(expected_packages))
	scopes = require_string_list(governance.get("benchmark_scopes"), "collections_benchmark_trust_governance.benchmark_scopes")
	expected_scopes = ["slice facade helpers", "map facade helpers", "internal slice/map helpers", "set helpers"]
	if scopes != expected_scopes:
		add_error("collections_benchmark_trust_governance.benchmark_scopes must be ordered as: " + ", ".join(expected_scopes))
	commands = require_string_list(governance.get("commands"), "collections_benchmark_trust_governance.commands")
	expected_commands = [
		"make bench-facade BENCH=Benchmark BENCHCOUNT=10 BENCHTIME=3s",
		"make bench-core BENCH=Benchmark BENCHCOUNT=10 BENCHTIME=3s",
		"go test -bench=Benchmark -benchmem -benchtime=100ms -count=1 -run=^$ ./vset",
	]
	if commands != expected_commands:
		add_error("collections_benchmark_trust_governance.commands must be ordered as: " + ", ".join(expected_commands))
	boundaries = require_string_list(governance.get("required_boundaries"), "collections_benchmark_trust_governance.required_boundaries")
	expected_boundaries = [
		"collection benchmark output is local baseline evidence",
		"compare against direct stdlib loops",
		"run repeated benchmarks before and after a change",
		"use benchstat before documenting improvement or regression",
		"publish collection benchmark output as evidence not marketing",
		"vset benchmark suite covers contains union intersect sub members json marshal json unmarshal",
	]
	if boundaries != expected_boundaries:
		add_error("collections_benchmark_trust_governance.required_boundaries must be ordered as: " + ", ".join(expected_boundaries))
	required_checks = require_string_list(governance.get("required_checks"), "collections_benchmark_trust_governance.required_checks")
	for check in ("docs-check", "ai-context-check", "governance-maturity-check", "bench-regression-check"):
		if check not in required_checks:
			add_error(f"collections_benchmark_trust_governance.required_checks must include {check}")
	if not (root / doc_path).exists():
		add_error(f"{doc_path} must exist")
	doc_text = (root / doc_path).read_text(encoding="utf-8") if (root / doc_path).exists() else ""
	for phrase in packages + scopes + commands + boundaries + ["Collection Benchmarks", "samber/lo", "duke-git/lancet"]:
		if doc_text and phrase not in doc_text:
			add_error(f"{doc_path} must include {phrase!r}")
	sprint_rows = extract_markdown_rows(root / roadmap_path, "Sprint order")
	sprint_55_rows = [row for row in sprint_rows if row.get("Sprint") == "55"]
	if len(sprint_55_rows) != 1:
		add_error(f"{roadmap_path} Sprint order must contain exactly one Sprint 55 row")
	else:
		expected_status = "Completed" if status == "completed" else "Active"
		if sprint_55_rows[0].get("Status") != expected_status:
			add_error(f"{roadmap_path} Sprint 55 status must be {expected_status}")
		sprint_text = " ".join(sprint_55_rows[0].values())
		for phrase in ("collection", "benchmark", "vslice", "vmap", "vset", "benchstat"):
			if phrase not in sprint_text:
				add_error(f"{roadmap_path} Sprint 55 row must mention {phrase!r}")
	roadmap_text = (root / roadmap_path).read_text(encoding="utf-8") if (root / roadmap_path).exists() else ""
	if "collections_benchmark_trust_governance" not in roadmap_text:
		add_error(f"{roadmap_path} must mention collections_benchmark_trust_governance")


def validate_first_use_golden_paths_governance() -> None:
	governance = require_mapping(ai_context.get("first_use_golden_paths_governance"), "first_use_golden_paths_governance")
	roadmap_path = governance.get("roadmap_path")
	if not isinstance(roadmap_path, str) or not roadmap_path.strip():
		add_error("first_use_golden_paths_governance.roadmap_path must be non-empty")
		roadmap_path = "docs/superpowers/plans/49-roadmap.md"
	doc_path = governance.get("doc_path")
	if not isinstance(doc_path, str) or not doc_path.strip():
		add_error("first_use_golden_paths_governance.doc_path must be non-empty")
		doc_path = "docs/doc/first-use-golden-paths.md"
	readme_path = governance.get("readme_path")
	if readme_path != "README.md":
		add_error("first_use_golden_paths_governance.readme_path must be README.md")
		readme_path = "README.md"
	if governance.get("sprint") != 46:
		add_error("first_use_golden_paths_governance.sprint must be 46")
	status = governance.get("status")
	if status not in {"active", "completed"}:
		add_error("first_use_golden_paths_governance.status must be active or completed")
	tasks = require_string_list(governance.get("tasks"), "first_use_golden_paths_governance.tasks")
	expected_tasks = ["string", "slice", "map", "json", "file", "http", "crypto", "config", "db", "cli"]
	if tasks != expected_tasks:
		add_error("first_use_golden_paths_governance.tasks must be ordered as: " + ", ".join(expected_tasks))
	facades = require_string_list(governance.get("facades"), "first_use_golden_paths_governance.facades")
	expected_facades = ["vstr", "vslice", "vmap", "vjson", "vfile", "vhttp", "vcrypto", "vconf", "vdb", "vcli"]
	if facades != expected_facades:
		add_error("first_use_golden_paths_governance.facades must be ordered as: " + ", ".join(expected_facades))
	boundaries = require_string_list(governance.get("required_boundaries"), "first_use_golden_paths_governance.required_boundaries")
	expected_boundaries = [
		"one recommended facade per task",
		"shortest example",
		"10 tasks in 10 minutes",
		"explicit error-returning flows before defaults",
		"Safe context-aware or WithOptions flows for trust boundaries",
	]
	if boundaries != expected_boundaries:
		add_error("first_use_golden_paths_governance.required_boundaries must be ordered as: " + ", ".join(expected_boundaries))
	required_checks = require_string_list(governance.get("required_checks"), "first_use_golden_paths_governance.required_checks")
	for check in ("docs-check", "ai-context-check", "governance-maturity-check"):
		if check not in required_checks:
			add_error(f"first_use_golden_paths_governance.required_checks must include {check}")
	if not (root / doc_path).exists():
		add_error(f"{doc_path} must exist")
	doc_text = (root / doc_path).read_text(encoding="utf-8") if (root / doc_path).exists() else ""
	for phrase in tasks + facades + boundaries + ["10 Tasks In 10 Minutes", "Recommended facade", "Shortest example"]:
		if doc_text and phrase not in doc_text:
			add_error(f"{doc_path} must include {phrase!r}")
	for task, facade in zip(tasks, facades):
		row = f"| {task} | `{facade}` |"
		if doc_text and row not in doc_text:
			add_error(f"{doc_path} must include one golden path row {row!r}")
	readme_text = (root / readme_path).read_text(encoding="utf-8") if (root / readme_path).exists() else ""
	if "first-use-golden-paths.md" not in readme_text:
		add_error("README.md must link docs/doc/first-use-golden-paths.md")
	doc_index_text = (root / "docs/doc/README.md").read_text(encoding="utf-8")
	if "first-use-golden-paths.md" not in doc_index_text:
		add_error("docs/doc/README.md must link docs/doc/first-use-golden-paths.md")
	sprint_rows = extract_markdown_rows(root / roadmap_path, "Sprint order")
	sprint_46_rows = [row for row in sprint_rows if row.get("Sprint") == "46"]
	if len(sprint_46_rows) != 1:
		add_error(f"{roadmap_path} Sprint order must contain exactly one Sprint 46 row")
	else:
		expected_status = "Completed" if status == "completed" else "Active"
		if sprint_46_rows[0].get("Status") != expected_status:
			add_error(f"{roadmap_path} Sprint 46 status must be {expected_status}")
	roadmap_text = (root / roadmap_path).read_text(encoding="utf-8") if (root / roadmap_path).exists() else ""
	if "first_use_golden_paths_governance" not in roadmap_text:
		add_error(f"{roadmap_path} must mention first_use_golden_paths_governance")


def validate_weak_facade_example_density_governance() -> None:
	governance = require_mapping(ai_context.get("weak_facade_example_density_governance"), "weak_facade_example_density_governance")
	roadmap_path = governance.get("roadmap_path")
	if not isinstance(roadmap_path, str) or not roadmap_path.strip():
		add_error("weak_facade_example_density_governance.roadmap_path must be non-empty")
		roadmap_path = "docs/superpowers/plans/49-roadmap.md"
	if governance.get("sprint") != 47:
		add_error("weak_facade_example_density_governance.sprint must be 47")
	status = governance.get("status")
	if status not in {"active", "completed"}:
		add_error("weak_facade_example_density_governance.status must be active or completed")
	if governance.get("selection_rule") != "common facades with example coverage below 25 percent":
		add_error("weak_facade_example_density_governance.selection_rule must describe below-25-percent common facades")
	target_facades = require_string_list(governance.get("target_facades"), "weak_facade_example_density_governance.target_facades")
	expected_targets = ["vlog", "vmail"]
	if target_facades != expected_targets:
		add_error("weak_facade_example_density_governance.target_facades must be ordered as: " + ", ".join(expected_targets))
	baseline = require_mapping(governance.get("baseline"), "weak_facade_example_density_governance.baseline")
	target_examples = require_mapping(governance.get("target_examples"), "weak_facade_example_density_governance.target_examples")
	expected_baseline = {
		"vlog": (54, 8, 14.8),
		"vmail": (51, 7, 13.7),
	}
	for facade in target_facades:
		entry = require_mapping(baseline.get(facade), f"weak_facade_example_density_governance.baseline.{facade}")
		expected_function_count, expected_examples, expected_ratio = expected_baseline[facade]
		if entry.get("function_count") != expected_function_count:
			add_error(f"weak_facade_example_density_governance.baseline.{facade}.function_count must be {expected_function_count}")
		if entry.get("functions_with_examples") != expected_examples:
			add_error(f"weak_facade_example_density_governance.baseline.{facade}.functions_with_examples must be {expected_examples}")
		if entry.get("example_coverage_percent") != expected_ratio:
			add_error(f"weak_facade_example_density_governance.baseline.{facade}.example_coverage_percent must be {expected_ratio}")
		target = target_examples.get(facade)
		if target != 12:
			add_error(f"weak_facade_example_density_governance.target_examples.{facade} must be 12")
		pkg = tool_packages.get(facade)
		if not pkg:
			add_error(f"docs/api/tools.json missing package {facade}")
			continue
		summary = require_mapping(pkg.get("summary"), f"docs/api/tools.json.packages.{facade}.summary")
		function_count = summary.get("function_count")
		example_count = summary.get("functions_with_examples")
		if function_count != expected_function_count:
			add_error(f"{facade} function count changed from governed baseline {expected_function_count} to {function_count}; update Sprint 47 governance deliberately")
		if not isinstance(example_count, int) or isinstance(example_count, bool):
			add_error(f"docs/api/tools.json.packages.{facade}.summary.functions_with_examples must be an integer")
		elif example_count < target:
			add_error(f"{facade} examples must be at least Sprint 47 target {target}; got {example_count}")
	if governance.get("ratchet_policy") != "raise selected weak facades in small increments instead of completing every API at once":
		add_error("weak_facade_example_density_governance.ratchet_policy must preserve small-increment ratchet wording")
	required_checks = require_string_list(governance.get("required_checks"), "weak_facade_example_density_governance.required_checks")
	for check in ("go test ./vlog ./vmail", "tools-gen", "docs-check", "ai-context-check", "governance-maturity-check"):
		if check not in required_checks:
			add_error(f"weak_facade_example_density_governance.required_checks must include {check}")
	sprint_rows = extract_markdown_rows(root / roadmap_path, "Sprint order")
	sprint_47_rows = [row for row in sprint_rows if row.get("Sprint") == "47"]
	if len(sprint_47_rows) != 1:
		add_error(f"{roadmap_path} Sprint order must contain exactly one Sprint 47 row")
	else:
		expected_status = "Completed" if status == "completed" else "Active"
		if sprint_47_rows[0].get("Status") != expected_status:
			add_error(f"{roadmap_path} Sprint 47 status must be {expected_status}")
		sprint_text = " ".join(sprint_47_rows[0].values())
		for phrase in ("vlog", "vmail", "12 examples", "ratchet"):
			if phrase not in sprint_text:
				add_error(f"{roadmap_path} Sprint 47 row must mention {phrase!r}")
	roadmap_text = (root / roadmap_path).read_text(encoding="utf-8") if (root / roadmap_path).exists() else ""
	if "weak_facade_example_density_governance" not in roadmap_text:
		add_error(f"{roadmap_path} must mention weak_facade_example_density_governance")


def validate_weak_facade_example_density_governance_2() -> None:
	governance = require_mapping(ai_context.get("weak_facade_example_density_governance_2"), "weak_facade_example_density_governance_2")
	roadmap_path = governance.get("roadmap_path")
	if not isinstance(roadmap_path, str) or not roadmap_path.strip():
		add_error("weak_facade_example_density_governance_2.roadmap_path must be non-empty")
		roadmap_path = "docs/superpowers/plans/49-roadmap.md"
	if governance.get("sprint") != 49:
		add_error("weak_facade_example_density_governance_2.sprint must be 49")
	status = governance.get("status")
	if status not in {"active", "completed"}:
		add_error("weak_facade_example_density_governance_2.status must be active or completed")
	if governance.get("selection_rule") != "common facades with example coverage below 25 percent":
		add_error("weak_facade_example_density_governance_2.selection_rule must describe below-25-percent common facades")
	target_facades = require_string_list(governance.get("target_facades"), "weak_facade_example_density_governance_2.target_facades")
	expected_targets = ["vcron", "vcache"]
	if target_facades != expected_targets:
		add_error("weak_facade_example_density_governance_2.target_facades must be ordered as: " + ", ".join(expected_targets))
	baseline = require_mapping(governance.get("baseline"), "weak_facade_example_density_governance_2.baseline")
	target_examples = require_mapping(governance.get("target_examples"), "weak_facade_example_density_governance_2.target_examples")
	expected_baseline = {
		"vcron": (51, 8, 15.7),
		"vcache": (24, 5, 20.8),
	}
	for facade in target_facades:
		entry = require_mapping(baseline.get(facade), f"weak_facade_example_density_governance_2.baseline.{facade}")
		expected_function_count, expected_examples, expected_ratio = expected_baseline[facade]
		if entry.get("function_count") != expected_function_count:
			add_error(f"weak_facade_example_density_governance_2.baseline.{facade}.function_count must be {expected_function_count}")
		if entry.get("functions_with_examples") != expected_examples:
			add_error(f"weak_facade_example_density_governance_2.baseline.{facade}.functions_with_examples must be {expected_examples}")
		if entry.get("example_coverage_percent") != expected_ratio:
			add_error(f"weak_facade_example_density_governance_2.baseline.{facade}.example_coverage_percent must be {expected_ratio}")
		target = target_examples.get(facade)
		if target != 12:
			add_error(f"weak_facade_example_density_governance_2.target_examples.{facade} must be 12")
		pkg = tool_packages.get(facade)
		if not pkg:
			add_error(f"docs/api/tools.json missing package {facade}")
			continue
		summary = require_mapping(pkg.get("summary"), f"docs/api/tools.json.packages.{facade}.summary")
		function_count = summary.get("function_count")
		example_count = summary.get("functions_with_examples")
		if function_count != expected_function_count:
			add_error(f"{facade} function count changed from governed baseline {expected_function_count} to {function_count}; update Sprint 49 governance deliberately")
		if not isinstance(example_count, int) or isinstance(example_count, bool):
			add_error(f"docs/api/tools.json.packages.{facade}.summary.functions_with_examples must be an integer")
		elif example_count < target:
			add_error(f"{facade} examples must be at least Sprint 49 target {target}; got {example_count}")
	if governance.get("ratchet_policy") != "raise selected weak facades in small increments instead of completing every API at once":
		add_error("weak_facade_example_density_governance_2.ratchet_policy must preserve small-increment ratchet wording")
	required_checks = require_string_list(governance.get("required_checks"), "weak_facade_example_density_governance_2.required_checks")
	for check in ("go test ./vcron ./vcache", "tools-gen", "docs-check", "ai-context-check", "governance-maturity-check"):
		if check not in required_checks:
			add_error(f"weak_facade_example_density_governance_2.required_checks must include {check}")
	sprint_rows = extract_markdown_rows(root / roadmap_path, "Sprint order")
	sprint_49_rows = [row for row in sprint_rows if row.get("Sprint") == "49"]
	if len(sprint_49_rows) != 1:
		add_error(f"{roadmap_path} Sprint order must contain exactly one Sprint 49 row")
	else:
		expected_status = "Completed" if status == "completed" else "Active"
		if sprint_49_rows[0].get("Status") != expected_status:
			add_error(f"{roadmap_path} Sprint 49 status must be {expected_status}")
		sprint_text = " ".join(sprint_49_rows[0].values())
		for phrase in ("vcron", "vcache", "12 examples", "ratchet"):
			if phrase not in sprint_text:
				add_error(f"{roadmap_path} Sprint 49 row must mention {phrase!r}")
	roadmap_text = (root / roadmap_path).read_text(encoding="utf-8") if (root / roadmap_path).exists() else ""
	if "weak_facade_example_density_governance_2" not in roadmap_text:
		add_error(f"{roadmap_path} must mention weak_facade_example_density_governance_2")


def validate_weak_facade_example_density_governance_3() -> None:
	governance = require_mapping(ai_context.get("weak_facade_example_density_governance_3"), "weak_facade_example_density_governance_3")
	roadmap_path = governance.get("roadmap_path")
	if not isinstance(roadmap_path, str) or not roadmap_path.strip():
		add_error("weak_facade_example_density_governance_3.roadmap_path must be non-empty")
		roadmap_path = "docs/superpowers/plans/49-roadmap.md"
	if governance.get("sprint") != 59:
		add_error("weak_facade_example_density_governance_3.sprint must be 59")
	status = governance.get("status")
	if status not in {"active", "completed"}:
		add_error("weak_facade_example_density_governance_3.status must be active or completed")
	if governance.get("selection_rule") != "common facades with example coverage below 25 percent":
		add_error("weak_facade_example_density_governance_3.selection_rule must describe below-25-percent common facades")
	target_facades = require_string_list(governance.get("target_facades"), "weak_facade_example_density_governance_3.target_facades")
	expected_targets = ["vconf", "vobj"]
	if target_facades != expected_targets:
		add_error("weak_facade_example_density_governance_3.target_facades must be ordered as: " + ", ".join(expected_targets))
	baseline = require_mapping(governance.get("baseline"), "weak_facade_example_density_governance_3.baseline")
	target_examples = require_mapping(governance.get("target_examples"), "weak_facade_example_density_governance_3.target_examples")
	expected_baseline = {
		"vconf": (39, 8, 20.5),
		"vobj": (49, 11, 22.4),
	}
	expected_targets_by_facade = {
		"vconf": 15,
		"vobj": 18,
	}
	for facade in target_facades:
		entry = require_mapping(baseline.get(facade), f"weak_facade_example_density_governance_3.baseline.{facade}")
		expected_function_count, expected_examples, expected_ratio = expected_baseline[facade]
		if entry.get("function_count") != expected_function_count:
			add_error(f"weak_facade_example_density_governance_3.baseline.{facade}.function_count must be {expected_function_count}")
		if entry.get("functions_with_examples") != expected_examples:
			add_error(f"weak_facade_example_density_governance_3.baseline.{facade}.functions_with_examples must be {expected_examples}")
		if entry.get("example_coverage_percent") != expected_ratio:
			add_error(f"weak_facade_example_density_governance_3.baseline.{facade}.example_coverage_percent must be {expected_ratio}")
		target = target_examples.get(facade)
		expected_target = expected_targets_by_facade[facade]
		if target != expected_target:
			add_error(f"weak_facade_example_density_governance_3.target_examples.{facade} must be {expected_target}")
		pkg = tool_packages.get(facade)
		if not pkg:
			add_error(f"docs/api/tools.json missing package {facade}")
			continue
		summary = require_mapping(pkg.get("summary"), f"docs/api/tools.json.packages.{facade}.summary")
		function_count = summary.get("function_count")
		example_count = summary.get("functions_with_examples")
		if function_count != expected_function_count:
			add_error(f"{facade} function count changed from governed baseline {expected_function_count} to {function_count}; update Sprint 59 governance deliberately")
		if not isinstance(example_count, int) or isinstance(example_count, bool):
			add_error(f"docs/api/tools.json.packages.{facade}.summary.functions_with_examples must be an integer")
		elif example_count < expected_target:
			add_error(f"{facade} examples must be at least Sprint 59 target {expected_target}; got {example_count}")
	if governance.get("ratchet_policy") != "raise selected weak facades in small increments instead of completing every API at once":
		add_error("weak_facade_example_density_governance_3.ratchet_policy must preserve small-increment ratchet wording")
	required_checks = require_string_list(governance.get("required_checks"), "weak_facade_example_density_governance_3.required_checks")
	for check in ("go test ./vconf ./vobj", "tools-gen", "docs-check", "ai-context-check", "governance-maturity-check", "agent-security-check"):
		if check not in required_checks:
			add_error(f"weak_facade_example_density_governance_3.required_checks must include {check}")
	sprint_rows = extract_markdown_rows(root / roadmap_path, "Sprint order")
	sprint_59_rows = [row for row in sprint_rows if row.get("Sprint") == "59"]
	if len(sprint_59_rows) != 1:
		add_error(f"{roadmap_path} Sprint order must contain exactly one Sprint 59 row")
	else:
		expected_status = "Completed" if status == "completed" else "Active"
		if sprint_59_rows[0].get("Status") != expected_status:
			add_error(f"{roadmap_path} Sprint 59 status must be {expected_status}")
		sprint_text = " ".join(sprint_59_rows[0].values())
		for phrase in ("vconf", "vobj", "15 examples", "18 examples", "ratchet"):
			if phrase not in sprint_text:
				add_error(f"{roadmap_path} Sprint 59 row must mention {phrase!r}")
	roadmap_text = (root / roadmap_path).read_text(encoding="utf-8") if (root / roadmap_path).exists() else ""
	if "weak_facade_example_density_governance_3" not in roadmap_text:
		add_error(f"{roadmap_path} must mention weak_facade_example_density_governance_3")


def validate_adoption_trust_governance() -> None:
	governance = require_mapping(ai_context.get("adoption_trust_governance"), "adoption_trust_governance")
	roadmap_path = governance.get("roadmap_path")
	if not isinstance(roadmap_path, str) or not roadmap_path.strip():
		add_error("adoption_trust_governance.roadmap_path must be non-empty")
		roadmap_path = "docs/superpowers/plans/49-roadmap.md"
	doc_path = governance.get("doc_path")
	if not isinstance(doc_path, str) or not doc_path.strip():
		add_error("adoption_trust_governance.doc_path must be non-empty")
		doc_path = "docs/doc/adoption-trust.md"
	readme_path = governance.get("readme_path")
	if readme_path != "README.md":
		add_error("adoption_trust_governance.readme_path must be README.md")
		readme_path = "README.md"
	if governance.get("sprint") != 48:
		add_error("adoption_trust_governance.sprint must be 48")
	status = governance.get("status")
	if status not in {"active", "completed"}:
		add_error("adoption_trust_governance.status must be active or completed")
	entrypoints = require_string_list(governance.get("required_entrypoints"), "adoption_trust_governance.required_entrypoints")
	expected_entrypoints = ["release notes", "compatibility policy", "deprecation policy", "security policy", "generated API catalog", "validation gates", "why trust this library"]
	if entrypoints != expected_entrypoints:
		add_error("adoption_trust_governance.required_entrypoints must be ordered as: " + ", ".join(expected_entrypoints))
	paths = require_string_list(governance.get("required_paths"), "adoption_trust_governance.required_paths")
	expected_paths = ["CHANGELOG.md", "SECURITY.md", "docs/api/exports.txt", "docs/api/tools.json", "docs/api/tools.md", "docs/doc/benchmark-trust.md"]
	if paths != expected_paths:
		add_error("adoption_trust_governance.required_paths must be ordered as: " + ", ".join(expected_paths))
	commands = require_string_list(governance.get("required_commands"), "adoption_trust_governance.required_commands")
	expected_commands = ["make agent-check", "make agent-full-check", "make release-check", "make api-freeze-check", "make release-notes-check"]
	if commands != expected_commands:
		add_error("adoption_trust_governance.required_commands must be ordered as: " + ", ".join(expected_commands))
	required_checks = require_string_list(governance.get("required_checks"), "adoption_trust_governance.required_checks")
	for check in ("docs-check", "ai-context-check", "governance-maturity-check", "release-notes-check", "api-freeze-check"):
		if check not in required_checks:
			add_error(f"adoption_trust_governance.required_checks must include {check}")
	if not (root / doc_path).exists():
		add_error(f"{doc_path} must exist")
	doc_text = (root / doc_path).read_text(encoding="utf-8") if (root / doc_path).exists() else ""
	for phrase in entrypoints + paths + commands + ["Why Trust This Library", "Adoption Checklist"]:
		if doc_text and phrase not in doc_text:
			add_error(f"{doc_path} must include {phrase!r}")
	readme_text = (root / readme_path).read_text(encoding="utf-8") if (root / readme_path).exists() else ""
	for phrase in ("adoption-trust.md", "CHANGELOG.md", "SECURITY.md", "API compatibility policy", "deprecation", "api-freeze-check"):
		if readme_text and phrase not in readme_text:
			add_error(f"README.md must include {phrase!r}")
	doc_index_text = (root / "docs/doc/README.md").read_text(encoding="utf-8")
	if "adoption-trust.md" not in doc_index_text:
		add_error("docs/doc/README.md must link docs/doc/adoption-trust.md")
	makefile_text = (root / "Makefile").read_text(encoding="utf-8")
	for target in ("release-notes-check", "api-freeze-check", "release-check", "agent-check", "agent-full-check"):
		if target not in makefile_text:
			add_error(f"Makefile must define {target}")
	sprint_rows = extract_markdown_rows(root / roadmap_path, "Sprint order")
	sprint_48_rows = [row for row in sprint_rows if row.get("Sprint") == "48"]
	if len(sprint_48_rows) != 1:
		add_error(f"{roadmap_path} Sprint order must contain exactly one Sprint 48 row")
	else:
		expected_status = "Completed" if status == "completed" else "Active"
		if sprint_48_rows[0].get("Status") != expected_status:
			add_error(f"{roadmap_path} Sprint 48 status must be {expected_status}")
		sprint_text = " ".join(sprint_48_rows[0].values())
		for phrase in ("adoption trust", "release notes", "compatibility", "deprecation", "security policy"):
			if phrase not in sprint_text:
				add_error(f"{roadmap_path} Sprint 48 row must mention {phrase!r}")
	roadmap_text = (root / roadmap_path).read_text(encoding="utf-8") if (root / roadmap_path).exists() else ""
	if "adoption_trust_governance" not in roadmap_text:
		add_error(f"{roadmap_path} must mention adoption_trust_governance")


def validate_docs_pkg_discovery_polish_governance() -> None:
	governance = require_mapping(ai_context.get("docs_pkg_discovery_polish_governance"), "docs_pkg_discovery_polish_governance")
	roadmap_path = governance.get("roadmap_path")
	if not isinstance(roadmap_path, str) or not roadmap_path.strip():
		add_error("docs_pkg_discovery_polish_governance.roadmap_path must be non-empty")
		roadmap_path = "docs/superpowers/plans/49-roadmap.md"
	readme_path = governance.get("readme_path")
	if readme_path != "README.md":
		add_error("docs_pkg_discovery_polish_governance.readme_path must be README.md")
		readme_path = "README.md"
	docs_hub_path = governance.get("docs_hub_path")
	if docs_hub_path != "docs/doc/README.md":
		add_error("docs_pkg_discovery_polish_governance.docs_hub_path must be docs/doc/README.md")
		docs_hub_path = "docs/doc/README.md"
	if governance.get("sprint") != 66:
		add_error("docs_pkg_discovery_polish_governance.sprint must be 66")
	status = governance.get("status")
	if status not in {"active", "completed"}:
		add_error("docs_pkg_discovery_polish_governance.status must be active or completed")
	top_facades = require_string_list(governance.get("top_facades"), "docs_pkg_discovery_polish_governance.top_facades")
	expected_top_facades = ["vhttp", "vcrypto", "vjson", "vfile"]
	if top_facades != expected_top_facades:
		add_error("docs_pkg_discovery_polish_governance.top_facades must be ordered as: " + ", ".join(expected_top_facades))
	readme_links = require_string_list(governance.get("required_readme_links"), "docs_pkg_discovery_polish_governance.required_readme_links")
	expected_readme_links = ["docs/doc/first-use-golden-paths.md", "docs/doc/task-index.md", "docs/doc/README.md", "pkg.go.dev/github.com/imajinyun/knifer-go"]
	if readme_links != expected_readme_links:
		add_error("docs_pkg_discovery_polish_governance.required_readme_links must be ordered as: " + ", ".join(expected_readme_links))
	docs_hub_links = require_string_list(governance.get("required_docs_hub_links"), "docs_pkg_discovery_polish_governance.required_docs_hub_links")
	expected_docs_hub_links = ["first-use-golden-paths.md", "task-index.md", "facade-tiering.md", "pkg.go.dev/github.com/imajinyun/knifer-go"]
	if docs_hub_links != expected_docs_hub_links:
		add_error("docs_pkg_discovery_polish_governance.required_docs_hub_links must be ordered as: " + ", ".join(expected_docs_hub_links))
	pkg_comment_links = require_string_list(governance.get("required_pkg_comment_links"), "docs_pkg_discovery_polish_governance.required_pkg_comment_links")
	expected_pkg_comment_links = [
		"docs/doc/22-vhttp.md",
		"docs/doc/safe-http-cookbook.md",
		"docs/doc/11-vcrypto.md",
		"docs/doc/safe-crypto-cookbook.md",
		"docs/doc/27-vjson.md",
		"docs/doc/17-vfile.md",
		"docs/doc/daily-json-file-faq.md",
	]
	if pkg_comment_links != expected_pkg_comment_links:
		add_error("docs_pkg_discovery_polish_governance.required_pkg_comment_links must be ordered as: " + ", ".join(expected_pkg_comment_links))
	boundaries = require_string_list(governance.get("required_boundaries"), "docs_pkg_discovery_polish_governance.required_boundaries")
	expected_boundaries = [
		"README first screen points to first-use paths",
		"docs hub separates new-user and task lookup entry points",
		"top facade package comments include pkg.go.dev-visible first links",
		"discovery polish does not add public APIs",
	]
	if boundaries != expected_boundaries:
		add_error("docs_pkg_discovery_polish_governance.required_boundaries must be ordered as: " + ", ".join(expected_boundaries))
	required_checks = require_string_list(governance.get("required_checks"), "docs_pkg_discovery_polish_governance.required_checks")
	for check in ("tools-gen", "docs-check", "ai-context-check", "governance-maturity-check"):
		if check not in required_checks:
			add_error(f"docs_pkg_discovery_polish_governance.required_checks must include {check}")
	readme_text = (root / readme_path).read_text(encoding="utf-8") if (root / readme_path).exists() else ""
	readme_first_screen = "\n".join(readme_text.splitlines()[:40])
	for phrase in ["New here"] + readme_links:
		if readme_first_screen and phrase not in readme_first_screen:
			add_error(f"README.md first screen must include {phrase!r}")
	docs_hub_text = (root / docs_hub_path).read_text(encoding="utf-8") if (root / docs_hub_path).exists() else ""
	for phrase in ["Start here", "New users", "Task-to-facade lookup"] + docs_hub_links:
		if docs_hub_text and phrase not in docs_hub_text:
			add_error(f"{docs_hub_path} must include {phrase!r}")
	facade_required_links = {
		"vhttp": ["docs/doc/22-vhttp.md", "docs/doc/safe-http-cookbook.md"],
		"vcrypto": ["docs/doc/11-vcrypto.md", "docs/doc/safe-crypto-cookbook.md"],
		"vjson": ["docs/doc/27-vjson.md", "docs/doc/daily-json-file-faq.md"],
		"vfile": ["docs/doc/17-vfile.md", "docs/doc/daily-json-file-faq.md"],
	}
	for facade in top_facades:
		doc_go_path = root / facade / "doc.go"
		if not doc_go_path.exists():
			add_error(f"{facade}/doc.go must exist")
			continue
		doc_go_text = doc_go_path.read_text(encoding="utf-8")
		for phrase in ["Start here", "https://github.com/imajinyun/knifer-go/blob/main/"] + facade_required_links.get(facade, []):
			if phrase not in doc_go_text:
				add_error(f"{facade}/doc.go must include {phrase!r}")
	sprint_rows = extract_markdown_rows(root / roadmap_path, "Sprint order")
	sprint_66_rows = [row for row in sprint_rows if row.get("Sprint") == "66"]
	if len(sprint_66_rows) != 1:
		add_error(f"{roadmap_path} Sprint order must contain exactly one Sprint 66 row")
	else:
		expected_status = "Completed" if status == "completed" else "Active"
		if sprint_66_rows[0].get("Status") != expected_status:
			add_error(f"{roadmap_path} Sprint 66 status must be {expected_status}")
		sprint_text = " ".join(sprint_66_rows[0].values())
		for phrase in ("pkg.go.dev", "README", "docs hub", "package comments"):
			if phrase not in sprint_text:
				add_error(f"{roadmap_path} Sprint 66 row must mention {phrase!r}")
	roadmap_text = (root / roadmap_path).read_text(encoding="utf-8") if (root / roadmap_path).exists() else ""
	if "docs_pkg_discovery_polish_governance" not in roadmap_text:
		add_error(f"{roadmap_path} must mention docs_pkg_discovery_polish_governance")


def validate_example_depth_governance() -> None:
	governance = require_mapping(ai_context.get("example_depth_governance"), "example_depth_governance")
	sprint = governance.get("sprint")
	if sprint != 22:
		add_error("example_depth_governance.sprint must be 22")
	status = governance.get("status")
	if status != "completed":
		add_error("example_depth_governance.status must be completed")
	roadmap_path = governance.get("roadmap_path")
	if not isinstance(roadmap_path, str) or not roadmap_path.strip():
		add_error("example_depth_governance.roadmap_path must be non-empty")
		roadmap_path = "docs/superpowers/plans/49-roadmap.md"
	roadmap_lane = governance.get("roadmap_lane")
	if roadmap_lane != "Examples":
		add_error("example_depth_governance.roadmap_lane must be Examples")
	target_facades = require_string_list(governance.get("target_facades"), "example_depth_governance.target_facades")
	expected_targets = ["vhttp", "vnet", "vnum", "vresty", "vzip"]
	if target_facades != expected_targets:
		add_error("example_depth_governance.target_facades must be ordered as: " + ", ".join(expected_targets))
	if governance.get("non_regression") is not True:
		add_error("example_depth_governance.non_regression must be true")
	if governance.get("first_implementation_facade") != "vnum":
		add_error("example_depth_governance.first_implementation_facade must be vnum")
	target_after_first_pass = require_mapping(governance.get("target_after_first_pass"), "example_depth_governance.target_after_first_pass")
	if target_after_first_pass.get("vnum") != 52:
		add_error("example_depth_governance.target_after_first_pass.vnum must be 52")
	baseline = require_mapping(governance.get("baseline"), "example_depth_governance.baseline")
	missing_baseline = sorted(set(target_facades) - set(baseline))
	if missing_baseline:
		add_error("example_depth_governance.baseline missing facade(s): " + ", ".join(missing_baseline))
	extra_baseline = sorted(set(baseline) - set(target_facades))
	if extra_baseline:
		add_error("example_depth_governance.baseline includes unmanaged facade(s): " + ", ".join(extra_baseline))
	for facade in target_facades:
		entry = require_mapping(baseline.get(facade), f"example_depth_governance.baseline.{facade}")
		function_baseline = entry.get("function_count")
		example_baseline = entry.get("functions_with_examples")
		ratio_baseline = entry.get("example_coverage_percent")
		pkg = tool_packages.get(facade)
		if not pkg:
			add_error(f"docs/api/tools.json missing package {facade}")
			continue
		summary = require_mapping(pkg.get("summary"), f"docs/api/tools.json.packages.{facade}.summary")
		function_count = summary.get("function_count")
		example_count = summary.get("functions_with_examples")
		if not isinstance(function_count, int) or isinstance(function_count, bool):
			add_error(f"docs/api/tools.json.packages.{facade}.summary.function_count must be an integer")
			continue
		if not isinstance(example_count, int) or isinstance(example_count, bool):
			add_error(f"docs/api/tools.json.packages.{facade}.summary.functions_with_examples must be an integer")
			continue
		if function_baseline != function_count:
			add_error(f"example_depth_governance.baseline.{facade}.function_count={function_baseline} must match tools catalog value {function_count}")
		if not isinstance(example_baseline, int) or isinstance(example_baseline, bool):
			add_error(f"example_depth_governance.baseline.{facade}.functions_with_examples must be an integer")
		elif example_count < example_baseline:
			add_error(f"{facade} examples regressed from baseline {example_baseline} to {example_count}")
		expected_ratio = round(example_baseline / function_count * 100, 1) if isinstance(example_baseline, int) and function_count else 0.0
		if ratio_baseline != expected_ratio:
			add_error(f"example_depth_governance.baseline.{facade}.example_coverage_percent={ratio_baseline} must be {expected_ratio}")
	roadmap_rows = extract_markdown_rows(root / roadmap_path, "Capability matrix")
	example_rows = [row for row in roadmap_rows if row.get("Area") == roadmap_lane]
	if len(example_rows) != 1:
		add_error(f"{roadmap_path} Capability matrix must contain exactly one {roadmap_lane} lane")
		return
	lane_text = " ".join(example_rows[0].values())
	missing_from_lane = [facade for facade in target_facades if f"`{facade}`" not in lane_text]
	if missing_from_lane:
		add_error(f"{roadmap_path} Examples lane missing facade(s): " + ", ".join(missing_from_lane))
	sprint_rows = extract_markdown_rows(root / roadmap_path, "Sprint order")
	sprint_22_rows = [row for row in sprint_rows if row.get("Sprint") == "22"]
	if len(sprint_22_rows) != 1:
		add_error(f"{roadmap_path} Sprint order must contain exactly one Sprint 22 row")
		return
	sprint_22 = sprint_22_rows[0]
	if sprint_22.get("Status") != "Completed":
		add_error(f"{roadmap_path} Sprint 22 status must be Completed")
	sprint_text = " ".join(sprint_22.values())
	missing_from_sprint = [facade for facade in target_facades if f"`{facade}`" not in sprint_text]
	if missing_from_sprint:
		add_error(f"{roadmap_path} Sprint 22 row missing facade(s): " + ", ".join(missing_from_sprint))


run_section("roadmap_catalog", [
	validate_roadmap_catalog_baseline,
	validate_roadmap_star_domain_scorecard,
])
run_section("star_domains", [
	validate_safe_http_cookbook_governance,
	validate_daily_json_file_faq_governance,
	validate_star_domain_no_missing_governance,
	validate_vdb_deepening_governance,
	validate_vdb_execution_evidence_governance,
	validate_vdb_example_depth_governance,
])
run_section("safe_crypto", [
	validate_safe_crypto_cookbook_governance,
	validate_safe_crypto_advanced_backlog_governance,
	validate_safe_crypto_otp_governance,
	validate_safe_crypto_password_hashing_governance,
	validate_safe_crypto_argon2id_governance,
	validate_safe_crypto_jwk_jwks_governance,
	validate_safe_crypto_jwk_jwks_implementation_governance,
	validate_safe_crypto_secret_handling_governance,
	validate_safe_crypto_interoperability_governance,
	validate_safe_crypto_benchmark_scope_governance,
	validate_safe_crypto_advanced_closeout_governance,
])
run_section("utility_comparison", [
	validate_utility_library_comparison_governance,
	validate_utility_top5_comparison_governance_v2,
	validate_utility_top5_refresh_workflow_governance,
	validate_go_version_adoption_governance,
])
run_section("collections", [
	validate_collections_comparison_governance,
	validate_collection_mindshare_pack_governance,
	validate_collection_advanced_backlog_governance,
	validate_collections_benchmark_trust_governance,
])
run_section("migration_domains", [
	validate_vconv_vbean_migration_governance,
	validate_vconv_cast_migration_governance,
	validate_vconv_cast_migration_examples_governance,
	validate_dynamic_data_toolkit_matrix_governance,
])
run_section("facade_tiering", [
	validate_task_index_governance,
	validate_task_index_auto_check_governance,
	validate_facade_tiering_import_governance,
	validate_facade_tiering_generated_view_governance,
])
run_section("daily_toolkit", [
	validate_daily_developer_toolkit_governance,
	validate_daily_utility_cookbook_v2_governance,
	validate_developer_debug_test_backlog_governance,
	validate_developer_debug_test_api_decision_governance,
])
run_section("trust_and_examples", [
	validate_benchmark_trust_governance,
	validate_first_use_golden_paths_governance,
	validate_weak_facade_example_density_governance,
	validate_weak_facade_example_density_governance_2,
	validate_weak_facade_example_density_governance_3,
	validate_adoption_trust_governance,
	validate_docs_pkg_discovery_polish_governance,
	validate_example_depth_governance,
])

if errors:
	for error in errors:
		print(f"governance maturity check error: {error}", file=sys.stderr)
	sys.exit(1)

print("governance maturity metadata is valid")
PY
