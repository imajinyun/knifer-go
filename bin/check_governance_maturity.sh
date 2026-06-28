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
