#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")/.."

BENCH_ONLY=0
if [ "${1:-}" = "--bench-only" ]; then
	BENCH_ONLY=1
fi

python3 - "${BENCH_ONLY}" <<'PY'
from __future__ import annotations

import json
import os
import re
import sys
from pathlib import Path

bench_only = sys.argv[1] == "1"
root = Path.cwd()
errors: list[str] = []


def add_error(message: str) -> None:
	errors.append(message)


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
	return any(make_target_depends_on(dep, dependency, seen) for dep in deps if re.match(r"^[A-Za-z0-9_.-]+$", dep))

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


def validate_benchmark_regression() -> None:
	bench = require_mapping(ai_context.get("benchmark_regression"), "benchmark_regression")
	if bench.get("benchstat_required") is not True:
		add_error("benchmark_regression.benchstat_required must be true")
	for field in ("baseline_command", "compare_command"):
		value = bench.get(field)
		if not isinstance(value, str) or not value.startswith("make bench-"):
			add_error(f"benchmark_regression.{field} must start with a make bench-* target")
	thresholds = require_mapping(bench.get("thresholds"), "benchmark_regression.thresholds")
	for key in ("ns_per_op_regression_percent", "bytes_per_op_regression_percent", "allocs_per_op_regression_percent"):
		value = thresholds.get(key)
		if not isinstance(value, (int, float)) or isinstance(value, bool) or value <= 0:
			add_error(f"benchmark_regression.thresholds.{key} must be a positive number")
	minimum_count = thresholds.get("minimum_count")
	if not isinstance(minimum_count, (int, float)) or isinstance(minimum_count, bool) or minimum_count < 10:
		add_error("benchmark_regression.thresholds.minimum_count must be at least 10")
	tracked = require_string_list(bench.get("tracked_packages"), "benchmark_regression.tracked_packages")
	if len(tracked) < 5:
		add_error("benchmark_regression.tracked_packages must include representative core and facade packages")
	if not any(pkg.startswith("./internal/") for pkg in tracked):
		add_error("benchmark_regression.tracked_packages must include at least one internal package")
	if not any(pkg.startswith("./v") for pkg in tracked):
		add_error("benchmark_regression.tracked_packages must include at least one public facade package")
	for pkg in tracked:
		if pkg.startswith("./") and not (root / pkg[2:]).is_dir():
			add_error(f"benchmark_regression.tracked_packages references missing package directory {pkg}")
	for target in ("bench-baseline", "bench-compare", "bench-regression-check", "benchstat"):
		if not re.search(rf"^{re.escape(target)}:(?:\s|$)", makefile, flags=re.MULTILINE):
			add_error(f"Makefile must define benchmark target {target}")


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


def validate_local_governance_gates() -> None:
	for target in ("quick-check", "full-check", "ci-workflow-check", "release-check"):
		if not make_target_depends_on(target, "bench-regression-check"):
			add_error(f"Makefile target {target} must depend on bench-regression-check")


def validate_api_convergence() -> None:
	api_convergence = require_mapping(ai_context.get("api_convergence"), "api_convergence")
	max_golden = api_convergence.get("max_golden_path_per_facade")
	if not isinstance(max_golden, (int, float)) or isinstance(max_golden, bool) or int(max_golden) != 5:
		add_error("api_convergence.max_golden_path_per_facade must be 5")
	max_golden = 5
	required = set(require_string_list(api_convergence.get("required_classifications"), "api_convergence.required_classifications"))
	for name in ("primary", "advanced", "compatibility", "avoid"):
		if name not in required:
			add_error(f"api_convergence.required_classifications must include {name}")
	facades = require_mapping(api_convergence.get("facades"), "api_convergence.facades")
	missing = sorted(public_facades - set(facades))
	if missing:
		add_error("api_convergence.facades missing public facade(s): " + ", ".join(missing))
	extra = sorted(set(facades) - public_facades)
	if extra:
		add_error("api_convergence.facades includes non-public facade(s): " + ", ".join(extra))
	for package_name in sorted(public_facades):
		entry = require_mapping(facades.get(package_name), f"api_convergence.facades.{package_name}")
		pkg = tool_packages.get(package_name)
		if not pkg:
			add_error(f"docs/api/tools.json missing package {package_name}")
			continue
		function_names = {fn.get("name") for fn in pkg.get("functions", []) if isinstance(fn, dict)}
		compatibility_functions = {fn.get("name") for fn in pkg.get("functions", []) if isinstance(fn, dict) and fn.get("status") == "compatibility"}
		golden = [item.get("name") for item in pkg.get("golden_path", []) if isinstance(item, dict)]
		if not golden:
			add_error(f"{package_name} must expose at least one golden_path entry")
		if len(golden) > max_golden:
			add_error(f"{package_name} golden_path has {len(golden)} entries; max is {max_golden}")
		primary = require_string_list(entry.get("primary"), f"api_convergence.facades.{package_name}.primary")
		if not primary or len(primary) > max_golden:
			add_error(f"api_convergence.facades.{package_name}.primary must contain 1-{max_golden} APIs")
		if set(primary) != set(golden):
			add_error(f"api_convergence.facades.{package_name}.primary must match docs/api/tools.json golden_path")
		bucket_values: dict[str, set[str]] = {}
		for bucket in ("primary", "advanced", "compatibility", "avoid"):
			values = require_string_list(entry.get(bucket), f"api_convergence.facades.{package_name}.{bucket}")
			bucket_values[bucket] = set(values)
			if len(values) != len(set(values)):
				add_error(f"api_convergence.facades.{package_name}.{bucket} must not contain duplicates")
			for fn_name in values:
				if fn_name not in function_names:
					add_error(f"api_convergence.facades.{package_name}.{bucket} references unknown API {fn_name}")
		for left, right in (("primary", "advanced"), ("primary", "avoid"), ("advanced", "compatibility"), ("advanced", "avoid"), ("compatibility", "avoid")):
			overlap = sorted(bucket_values[left] & bucket_values[right])
			if overlap:
				add_error(f"api_convergence.facades.{package_name}.{left} and {right} overlap: " + ", ".join(overlap))
		for fn_name in require_string_list(entry.get("compatibility"), f"api_convergence.facades.{package_name}.compatibility"):
			if fn_name not in compatibility_functions:
				add_error(f"api_convergence.facades.{package_name}.compatibility includes non-compatibility API {fn_name}")
		if not isinstance(entry.get("decision"), str) or not entry["decision"].strip():
			add_error(f"api_convergence.facades.{package_name}.decision must be non-empty")


def validate_lifecycle() -> None:
	lifecycle = require_mapping(ai_context.get("package_lifecycle"), "package_lifecycle")
	allowed = set(require_string_list(lifecycle.get("allowed_grades"), "package_lifecycle.allowed_grades"))
	for grade in ("core", "stable", "maintenance", "adapter", "heavy", "candidate-for-split", "candidate-for-deprecation"):
		if grade not in allowed:
			add_error(f"package_lifecycle.allowed_grades must include {grade}")
	packages = require_mapping(lifecycle.get("packages"), "package_lifecycle.packages")
	missing = sorted(public_facades - set(packages))
	if missing:
		add_error("package_lifecycle.packages missing public facade(s): " + ", ".join(missing))
	extra = sorted(set(packages) - public_facades)
	if extra:
		add_error("package_lifecycle.packages includes non-public facade(s): " + ", ".join(extra))
	dependency_tiers = require_mapping(ai_context.get("dependency_tiers"), "dependency_tiers")
	heavy = set(require_string_list(dependency_tiers.get("heavy_extension_facades"), "dependency_tiers.heavy_extension_facades"))
	adapters = set(require_string_list(dependency_tiers.get("provider_contract_facades"), "dependency_tiers.provider_contract_facades"))
	core = set(require_string_list(dependency_tiers.get("core_facades"), "dependency_tiers.core_facades"))
	for tier_name, tier_values in (("heavy_extension_facades", heavy), ("provider_contract_facades", adapters), ("core_facades", core)):
		unknown = sorted(tier_values - public_facades)
		if unknown:
			add_error(f"dependency_tiers.{tier_name} includes non-public facade(s): " + ", ".join(unknown))
	if heavy & adapters or heavy & core or adapters & core:
		add_error("dependency_tiers facade sets must be mutually exclusive")
	for package_name, entry_value in sorted(packages.items()):
		entry = require_mapping(entry_value, f"package_lifecycle.packages.{package_name}")
		grade = entry.get("grade")
		if grade not in allowed:
			add_error(f"package_lifecycle.packages.{package_name}.grade must be an allowed lifecycle grade")
		if not isinstance(entry.get("rationale"), str) or not entry["rationale"].strip():
			add_error(f"package_lifecycle.packages.{package_name}.rationale must be non-empty")
		if package_name in heavy and grade != "heavy":
			add_error(f"package_lifecycle.packages.{package_name}.grade must be heavy")
		if package_name in adapters and grade != "adapter":
			add_error(f"package_lifecycle.packages.{package_name}.grade must be adapter")
		if package_name in core and grade not in {"core", "stable", "maintenance", "candidate-for-split", "candidate-for-deprecation"}:
			add_error(f"package_lifecycle.packages.{package_name}.grade must remain core-compatible")


def validate_capability_domains() -> None:
	domains = require_mapping(ai_context.get("capability_domains"), "capability_domains")
	expected_domains = {
		"data_transform",
		"collections",
		"text_parsing",
		"trust_boundary",
		"security_primitives",
		"runtime_adapters",
		"domain_helpers",
	}
	missing_domains = sorted(expected_domains - set(domains))
	if missing_domains:
		add_error("capability_domains missing required domain(s): " + ", ".join(missing_domains))
	extra_domains = sorted(set(domains) - expected_domains)
	if extra_domains:
		add_error("capability_domains includes unknown domain(s): " + ", ".join(extra_domains))
	covered_packages: set[str] = set()
	allowed_test_types = {"benchmark", "contract", "error_contract", "example", "fuzz", "misuse", "provider_contract", "security"}
	required_test_matrix = {
		"data_transform": {"contract", "fuzz", "error_contract"},
		"collections": {"contract", "benchmark"},
		"text_parsing": {"contract", "fuzz", "provider_contract"},
		"trust_boundary": {"contract", "security", "misuse", "fuzz", "error_contract"},
		"security_primitives": {"contract", "security", "misuse", "error_contract"},
		"runtime_adapters": {"contract", "provider_contract"},
		"domain_helpers": {"contract", "example"},
	}
	for domain_name, domain_value in sorted(domains.items()):
		domain = require_mapping(domain_value, f"capability_domains.{domain_name}")
		purpose = domain.get("purpose")
		if not isinstance(purpose, str) or not purpose.strip():
			add_error(f"capability_domains.{domain_name}.purpose must be non-empty")
		packages = require_string_list(domain.get("packages"), f"capability_domains.{domain_name}.packages")
		if len(packages) < 2:
			add_error(f"capability_domains.{domain_name}.packages must include at least 2 facades")
		unknown = sorted(set(packages) - public_facades)
		if unknown:
			add_error(f"capability_domains.{domain_name}.packages includes non-public facade(s): " + ", ".join(unknown))
		covered_packages.update(packages)
		if len(require_string_list(domain.get("required_focus"), f"capability_domains.{domain_name}.required_focus")) < 2:
			add_error(f"capability_domains.{domain_name}.required_focus must include at least 2 focus areas")
		required_tests = set(require_string_list(domain.get("required_tests"), f"capability_domains.{domain_name}.required_tests"))
		if not required_tests:
			add_error(f"capability_domains.{domain_name}.required_tests must be non-empty")
		unknown_tests = sorted(required_tests - allowed_test_types)
		if unknown_tests:
			add_error(f"capability_domains.{domain_name}.required_tests includes unknown test type(s): " + ", ".join(unknown_tests))
		missing_tests = sorted(required_test_matrix.get(domain_name, set()) - required_tests)
		if missing_tests:
			add_error(f"capability_domains.{domain_name}.required_tests missing required test type(s): " + ", ".join(missing_tests))
	missing_packages = sorted(public_facades - covered_packages)
	if missing_packages:
		add_error("capability_domains do not cover public facade(s): " + ", ".join(missing_packages))

	security_sensitive = set(require_string_list(ai_context.get("security_sensitive_packages"), "security_sensitive_packages"))
	trust_boundary = set(require_string_list(require_mapping(domains.get("trust_boundary"), "capability_domains.trust_boundary").get("packages"), "capability_domains.trust_boundary.packages"))
	security_primitives = set(require_string_list(require_mapping(domains.get("security_primitives"), "capability_domains.security_primitives").get("packages"), "capability_domains.security_primitives.packages"))
	runtime_adapters = set(require_string_list(require_mapping(domains.get("runtime_adapters"), "capability_domains.runtime_adapters").get("packages"), "capability_domains.runtime_adapters.packages"))
	covered_sensitive = trust_boundary | security_primitives | runtime_adapters
	missing_sensitive = sorted(security_sensitive - covered_sensitive)
	if missing_sensitive:
		add_error("security-sensitive facades must be represented by trust_boundary, security_primitives, or runtime_adapters: " + ", ".join(missing_sensitive))


def validate_dependency_isolation() -> None:
	dependency_tiers = require_mapping(ai_context.get("dependency_tiers"), "dependency_tiers")
	heavy_dependency_allowlist = require_mapping(dependency_tiers.get("heavy_dependency_allowlist"), "dependency_tiers.heavy_dependency_allowlist")
	heavy_imports = {
		"github.com/getsentry/sentry-go": {"internal/errx", "verr"},
		"github.com/makiuchi-d/gozxing*": {"internal/imgx", "vimg"},
		"github.com/sirupsen/logrus": {"internal/errx", "verr"},
		"github.com/xuri/excelize/v2": {"internal/poi", "vpoi"},
		"resty.dev/v3": {"internal/httpx/resty", "vresty"},
	}
	allowlist_from_context = {
		import_path: set(require_string_list(prefixes, f"dependency_tiers.heavy_dependency_allowlist.{import_path}"))
		for import_path, prefixes in heavy_dependency_allowlist.items()
	}
	if allowlist_from_context != heavy_imports:
		add_error("dependency_tiers.heavy_dependency_allowlist must match governance heavy dependency isolation policy")
	for path in root.rglob("*.go"):
		if path.name.endswith("_test.go") or "/.git/" in path.as_posix():
			continue
		rel = path.relative_to(root).as_posix()
		text = path.read_text(encoding="utf-8")
		for import_path, allowed_prefixes in heavy_imports.items():
			needle = import_path.rstrip("*")
			if f'"{needle}' not in text:
				continue
			if not any(rel.startswith(prefix + "/") or rel == prefix + ".go" for prefix in allowed_prefixes):
				add_error(f"{rel} imports heavy dependency {import_path} outside isolated facade/internal package")


def validate_error_model() -> None:
	error_model = require_mapping(ai_context.get("error_model"), "error_model")
	taxonomy = error_model.get("taxonomy")
	if not isinstance(taxonomy, list):
		add_error("error_model.taxonomy must be a list")
		taxonomy = []
	codes = set()
	for index, item in enumerate(taxonomy):
		entry = require_mapping(item, f"error_model.taxonomy[{index}]")
		for key in ("category", "code", "use_when"):
			if not isinstance(entry.get(key), str) or not entry[key].strip():
				add_error(f"error_model.taxonomy[{index}].{key} must be non-empty")
		if isinstance(entry.get("code"), str):
			codes.add(entry["code"])
	expected_codes = {"GK_INVALID_INPUT", "GK_NOT_FOUND", "GK_UNSUPPORTED", "GK_UNSAFE_RESOURCE", "GK_TIMEOUT", "GK_PROVIDER_FAILURE", "GK_INTERNAL"}
	if codes != expected_codes:
		add_error("error_model.taxonomy must cover exactly: " + ", ".join(sorted(expected_codes)))
	errors_go = (root / "errors.go").read_text(encoding="utf-8")
	for constant_name in ("ErrCodeInvalidInput", "ErrCodeNotFound", "ErrCodeUnsupported", "ErrCodeUnsafeResource", "ErrCodeTimeout", "ErrCodeProviderFailure", "ErrCodeInternal"):
		if constant_name not in errors_go:
			add_error(f"errors.go must define {constant_name}")
	for reference in require_string_list(error_model.get("contract_tests"), "error_model.contract_tests"):
		if not reference_exists(reference):
			add_error(f"error_model.contract_tests references missing file {reference}")


def validate_dynamic_semantic_contracts() -> None:
	contracts = require_mapping(ai_context.get("dynamic_semantic_contracts"), "dynamic_semantic_contracts")
	required_domains = set(require_string_list(contracts.get("required_domains"), "dynamic_semantic_contracts.required_domains"))
	expected_domains = {"vbean_decode_copy_merge", "vjson_dynamic", "vobj_dynamic", "vconf_dynamic"}
	if required_domains != expected_domains:
		add_error("dynamic_semantic_contracts.required_domains must cover exactly: " + ", ".join(sorted(expected_domains)))
	domains = require_mapping(contracts.get("domains"), "dynamic_semantic_contracts.domains")
	missing = sorted(required_domains - set(domains))
	if missing:
		add_error("dynamic_semantic_contracts.domains missing required domain(s): " + ", ".join(missing))
	extra = sorted(set(domains) - required_domains)
	if extra:
		add_error("dynamic_semantic_contracts.domains includes unknown domain(s): " + ", ".join(extra))
	covered_packages: set[str] = set()
	for domain_name, domain_value in sorted(domains.items()):
		domain = require_mapping(domain_value, f"dynamic_semantic_contracts.domains.{domain_name}")
		packages = require_string_list(domain.get("packages"), f"dynamic_semantic_contracts.domains.{domain_name}.packages")
		covered_packages.update(packages)
		for package in packages:
			if not (root / package).is_dir():
				add_error(f"dynamic_semantic_contracts.domains.{domain_name}.packages references missing directory {package}")
		if len(require_string_list(domain.get("guarantees"), f"dynamic_semantic_contracts.domains.{domain_name}.guarantees")) < 3:
			add_error(f"dynamic_semantic_contracts.domains.{domain_name}.guarantees must contain at least 3 semantic guarantees")
		for field in ("contract_tests", "fuzz_tests"):
			references = require_string_list(domain.get(field), f"dynamic_semantic_contracts.domains.{domain_name}.{field}")
			if field == "contract_tests" and not references:
				add_error(f"dynamic_semantic_contracts.domains.{domain_name}.{field} must be non-empty")
			for reference in references:
				if not references_function(reference):
					add_error(f"dynamic_semantic_contracts.domains.{domain_name}.{field} must reference explicit test functions, got {reference}")
				if not reference_exists(reference):
					add_error(f"dynamic_semantic_contracts.domains.{domain_name}.{field} references missing file or function {reference}")
	expected_packages = {"internal/bean", "vjson", "vobj", "vconf"}
	if covered_packages != expected_packages:
		add_error("dynamic_semantic_contracts must cover exactly package directories: " + ", ".join(sorted(expected_packages)))


def validate_threat_model() -> None:
	threat_model = require_mapping(ai_context.get("threat_model"), "threat_model")
	methodology = threat_model.get("methodology")
	if not isinstance(methodology, str) or "STRIDE" not in methodology or "DREAD" not in methodology:
		add_error("threat_model.methodology must mention STRIDE and DREAD")
	for reference in require_string_list(threat_model.get("misuse_tests"), "threat_model.misuse_tests"):
		if not references_function(reference):
			add_error(f"threat_model.misuse_tests must reference explicit test functions, got {reference}")
		if not reference_exists(reference):
			add_error(f"threat_model.misuse_tests references missing file or function {reference}")
	domains = require_mapping(threat_model.get("domains"), "threat_model.domains")
	covered_packages: set[str] = set()
	for domain_name, domain_value in sorted(domains.items()):
		domain = require_mapping(domain_value, f"threat_model.domains.{domain_name}")
		packages = require_string_list(domain.get("packages"), f"threat_model.domains.{domain_name}.packages")
		covered_packages.update(packages)
		threats = require_string_list(domain.get("threats"), f"threat_model.domains.{domain_name}.threats")
		if not threats:
			add_error(f"threat_model.domains.{domain_name}.threats must be non-empty")
		contract_tests = require_mapping(domain.get("contract_tests"), f"threat_model.domains.{domain_name}.contract_tests")
		missing_threat_contracts = sorted(set(threats) - set(contract_tests))
		if missing_threat_contracts:
			add_error(f"threat_model.domains.{domain_name}.contract_tests missing threat(s): " + ", ".join(missing_threat_contracts))
		for threat, references_value in sorted(contract_tests.items()):
			if threat not in threats:
				add_error(f"threat_model.domains.{domain_name}.contract_tests declares unknown threat {threat}")
			references = require_string_list(references_value, f"threat_model.domains.{domain_name}.contract_tests.{threat}")
			minimum_references = 2
			if domain_name in {"network_clients", "crypto_identity_randomness", "data_and_cli_boundaries"}:
				minimum_references = 3
			if len(references) < minimum_references:
				add_error(f"threat_model.domains.{domain_name}.contract_tests.{threat} must reference at least {minimum_references} tests")
			if not any(references_function(reference) for reference in references):
				add_error(f"threat_model.domains.{domain_name}.contract_tests.{threat} must reference at least one explicit test function")
			for reference in references:
				if not reference_exists(reference):
					add_error(f"threat_model.domains.{domain_name}.contract_tests.{threat} references missing file or function {reference}")
		for reference in require_string_list(domain.get("misuse_tests"), f"threat_model.domains.{domain_name}.misuse_tests"):
			if not reference_exists(reference):
				add_error(f"threat_model.domains.{domain_name}.misuse_tests references missing file {reference}")
	security_sensitive = set(require_string_list(ai_context.get("security_sensitive_packages"), "security_sensitive_packages"))
	unknown_sensitive = sorted(security_sensitive - public_facades)
	if unknown_sensitive:
		add_error("security_sensitive_packages includes non-public facade(s): " + ", ".join(unknown_sensitive))
	missing = sorted(security_sensitive - covered_packages)
	if missing:
		add_error("threat_model.domains do not cover security-sensitive package(s): " + ", ".join(missing))
	unexpected = sorted(covered_packages - security_sensitive)
	if unexpected:
		add_error("threat_model.domains cover non-security-sensitive package(s): " + ", ".join(unexpected))


validate_benchmark_regression()
if not bench_only:
	validate_local_governance_gates()
	validate_roadmap_catalog_baseline()
	validate_roadmap_star_domain_scorecard()
	validate_safe_http_cookbook_governance()
	validate_safe_crypto_cookbook_governance()
	validate_daily_json_file_faq_governance()
	validate_star_domain_no_missing_governance()
	validate_vdb_deepening_governance()
	validate_vdb_execution_evidence_governance()
	validate_vdb_example_depth_governance()
	validate_safe_crypto_advanced_backlog_governance()
	validate_safe_crypto_otp_governance()
	validate_safe_crypto_password_hashing_governance()
	validate_safe_crypto_argon2id_governance()
	validate_safe_crypto_jwk_jwks_governance()
	validate_safe_crypto_jwk_jwks_implementation_governance()
	validate_safe_crypto_secret_handling_governance()
	validate_safe_crypto_interoperability_governance()
	validate_safe_crypto_benchmark_scope_governance()
	validate_utility_library_comparison_governance()
	validate_safe_crypto_advanced_closeout_governance()
	validate_go_version_adoption_governance()
	validate_collections_comparison_governance()
	validate_vconv_vbean_migration_governance()
	validate_daily_developer_toolkit_governance()
	validate_benchmark_trust_governance()
	validate_first_use_golden_paths_governance()
	validate_weak_facade_example_density_governance()
	validate_weak_facade_example_density_governance_2()
	validate_adoption_trust_governance()
	validate_example_depth_governance()
	validate_api_convergence()
	validate_lifecycle()
	validate_capability_domains()
	validate_dependency_isolation()
	validate_error_model()
	validate_dynamic_semantic_contracts()
	validate_threat_model()

if errors:
	for error in errors:
		print(f"governance maturity check error: {error}", file=sys.stderr)
	sys.exit(1)

if bench_only:
	print("benchmark regression metadata is valid")
else:
	print("governance maturity metadata is valid")
PY
