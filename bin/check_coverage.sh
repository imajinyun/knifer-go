#!/usr/bin/env bash
#
# check_coverage.sh enforces repository-wide and package-level coverage baselines.
# ai-context.json is the source of truth for default thresholds.
# Set COVERAGE_THRESHOLD or PACKAGE_COVERAGE_THRESHOLDS to override defaults locally.

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
AI_CONTEXT="${ROOT_DIR}/ai-context.json"
DIFF_FILTER="ACDMRTUXB"

coverage_file="${1:-coverage.out}"

changed_files_from_base() {
	local base_ref="${AGENT_CHANGE_BASE_REF:-}"
	if [ -z "${base_ref}" ] && [ -n "${GITHUB_BASE_REF:-}" ]; then
		base_ref="origin/${GITHUB_BASE_REF}"
	fi
	if [ -n "${base_ref}" ] && git -C "${ROOT_DIR}" rev-parse --verify --quiet "${base_ref}^{commit}" >/dev/null; then
		git -C "${ROOT_DIR}" diff --name-only --diff-filter="${DIFF_FILTER}" "${base_ref}...HEAD" --
	fi
}

changed_files_from_worktree() {
	git -C "${ROOT_DIR}" diff --name-only --diff-filter="${DIFF_FILTER}" HEAD --
	git -C "${ROOT_DIR}" diff --name-only --cached --diff-filter="${DIFF_FILTER}" --
	git -C "${ROOT_DIR}" ls-files --others --exclude-standard --
}

changed_files="$({
	changed_files_from_base
	changed_files_from_worktree
} | sort -u)"

coverage_config="$(
	python3 - "${AI_CONTEXT}" "${changed_files}" <<'PY'
import json
import os
import sys

ai_context = sys.argv[1]
changed_files = [line.strip().strip("/") for line in sys.argv[2].splitlines() if line.strip()]
root_dir = os.path.dirname(ai_context)

with open(ai_context, "r", encoding="utf-8") as f:
    data = json.load(f)

def has_statement_source(package_dir):
    path = os.path.join(root_dir, package_dir)
    if not os.path.isdir(path):
        return False
    return any(
        entry.endswith(".go") and not entry.endswith("_test.go") and entry != "doc.go"
        for entry in os.listdir(path)
    )

coverage_gates = data["coverage_gates"]
repository_threshold = coverage_gates["repository_threshold"]
security_sensitive_min_threshold = coverage_gates.get("security_sensitive_min_threshold", 0)
package_thresholds = " ".join(
    f"{package_path}={threshold:.1f}"
    for package_path, threshold in coverage_gates["package_thresholds"].items()
)
changed_package_thresholds = set()
module = data["project"]["module"]
facade_to_internal = {
    entry["package"]: entry["internal"].rstrip("/")
    for entry in data["public_facades"]
}
security_sensitive_paths = set()
changed_security_sensitive_paths = set()
security_prefix_to_package_dir = {}
for package in data["security_sensitive_packages"]:
    package_dir = package.rstrip("/")
    security_prefix_to_package_dir[package_dir + "/"] = package_dir
    if has_statement_source(package_dir):
        security_sensitive_paths.add(f"{module}/{package_dir}")
    internal = facade_to_internal.get(package)
    if internal and has_statement_source(internal):
        security_sensitive_paths.add(f"{module}/{internal}")
        security_prefix_to_package_dir[internal.rstrip("/") + "/"] = internal.rstrip("/")
security_sensitive_paths = " ".join(sorted(security_sensitive_paths))
for path in changed_files:
    if not path.endswith(".go") or path.endswith("/doc.go"):
        continue
    package_path = f"{module}/{os.path.dirname(path)}"
    if package_path in coverage_gates["package_thresholds"]:
        changed_package_thresholds.add(f"{package_path}={coverage_gates['package_thresholds'][package_path]:.1f}")
    for prefix, package_dir in security_prefix_to_package_dir.items():
        if path.startswith(prefix) and has_statement_source(package_dir):
            changed_security_sensitive_paths.add(f"{module}/{package_dir}")
changed_security_sensitive_paths = " ".join(sorted(changed_security_sensitive_paths))
print(f"{repository_threshold:.1f}|{package_thresholds}|{security_sensitive_paths}|{security_sensitive_min_threshold:.1f}|{changed_security_sensitive_paths}|{' '.join(sorted(changed_package_thresholds))}")
PY
)"

IFS='|' read -r metadata_threshold metadata_package_thresholds metadata_security_sensitive_paths metadata_security_sensitive_min_threshold metadata_changed_security_sensitive_paths metadata_changed_package_thresholds <<<"${coverage_config}"
threshold="${COVERAGE_THRESHOLD:-${metadata_threshold}}"
package_thresholds="${PACKAGE_COVERAGE_THRESHOLDS:-${metadata_package_thresholds}}"
changed_package_thresholds="${CHANGED_PACKAGE_COVERAGE_THRESHOLDS:-${metadata_changed_package_thresholds}}"
security_sensitive_paths="${SECURITY_SENSITIVE_COVERAGE_PATHS:-${metadata_security_sensitive_paths}}"
security_sensitive_min_threshold="${SECURITY_SENSITIVE_MIN_COVERAGE_THRESHOLD:-${metadata_security_sensitive_min_threshold}}"
changed_security_sensitive_paths="${CHANGED_SECURITY_SENSITIVE_COVERAGE_PATHS:-${metadata_changed_security_sensitive_paths}}"
coverage_check_all_packages="${COVERAGE_CHECK_ALL_PACKAGES:-0}"

if [ ! -f "${coverage_file}" ]; then
	echo "COVERAGE CHECK ERROR: ${coverage_file} does not exist" >&2
	exit 2
fi

total="$(
	go tool cover -func="${coverage_file}" |
		awk '/^total:/ { gsub("%", "", $3); print $3 }'
)"

if [ -z "${total}" ]; then
	echo "COVERAGE CHECK ERROR: cannot read total coverage from ${coverage_file}" >&2
	exit 2
fi

awk -v total="${total}" -v threshold="${threshold}" '
BEGIN {
	if (total + 0 < threshold + 0) {
		printf "coverage %.1f%% is below required %.1f%%\n", total, threshold > "/dev/stderr"
		exit 1
	}
	printf "coverage %.1f%% meets required %.1f%%\n", total, threshold
}
'

if [ "${coverage_check_all_packages}" != "1" ]; then
	echo "package coverage thresholds skipped for unchanged packages; set COVERAGE_CHECK_ALL_PACKAGES=1 to enforce all package thresholds"
elif [ -z "${package_thresholds}" ]; then
	:
else
	for gate in ${package_thresholds}; do
		package_path="${gate%%=*}"
		package_threshold="${gate#*=}"
		package_total="$(
			awk -v pkg="${package_path}" '
			NR == 1 { next }
			{
				file = $1
				sub(/:.*/, "", file)
				if (file ~ "^" pkg "/[^/]+\\.go$") {
					statements += $2
					if ($3 > 0) {
						covered += $2
					}
				}
			}
			END {
				if (statements > 0) {
					printf "%.1f", covered * 100 / statements
				}
			}
			' "${coverage_file}"
		)"
		if [ -z "${package_total}" ]; then
			echo "COVERAGE CHECK ERROR: package ${package_path} has no coverage data" >&2
			exit 2
		fi
		awk -v package_path="${package_path}" -v total="${package_total}" -v threshold="${package_threshold}" '
		BEGIN {
			if (total + 0 < threshold + 0) {
				printf "%s coverage %.1f%% is below required %.1f%%\n", package_path, total, threshold > "/dev/stderr"
				exit 1
			}
			printf "%s coverage %.1f%% meets required %.1f%%\n", package_path, total, threshold
		}
		'
	done
fi

if [ -n "${changed_package_thresholds}" ]; then
	for gate in ${changed_package_thresholds}; do
		package_path="${gate%%=*}"
		package_threshold="${gate#*=}"
		package_total="$(
			awk -v pkg="${package_path}" '
			NR == 1 { next }
			{
				file = $1
				sub(/:.*/, "", file)
				if (file ~ "^" pkg "/[^/]+\\.go$") {
					statements += $2
					if ($3 > 0) {
						covered += $2
					}
				}
			}
			END {
				if (statements > 0) {
					printf "%.1f", covered * 100 / statements
				}
			}
			' "${coverage_file}"
		)"
		if [ -z "${package_total}" ]; then
			echo "COVERAGE CHECK ERROR: changed package ${package_path} has no coverage data" >&2
			exit 2
		fi
		awk -v package_path="${package_path}" -v total="${package_total}" -v threshold="${package_threshold}" '
		BEGIN {
			if (total + 0 < threshold + 0) {
				printf "changed package %s coverage %.1f%% is below required %.1f%%\n", package_path, total, threshold > "/dev/stderr"
				exit 1
			}
			printf "changed package %s coverage %.1f%% meets required %.1f%%\n", package_path, total, threshold
		}
		'
	done
fi

if [ -z "${security_sensitive_paths}" ]; then
	:
else
	missing_security_sensitive_paths=""
	below_threshold_security_sensitive_paths=""
	security_sensitive_count=0
	for package_path in ${security_sensitive_paths}; do
		package_total="$(
			awk -v pkg="${package_path}" '
			NR == 1 { next }
			{
				file = $1
				sub(/:.*/, "", file)
				if (file ~ "^" pkg "/[^/]+\\.go$") {
					statements += $2
					if ($3 > 0) {
						covered += $2
					}
				}
			}
			END {
				if (statements > 0) {
					printf "%.1f", covered * 100 / statements
				}
			}
			' "${coverage_file}"
		)"
		if [ -z "${package_total}" ]; then
			missing_security_sensitive_paths="${missing_security_sensitive_paths}${package_path}
"
		else
			security_sensitive_count=$((security_sensitive_count + 1))
			if awk -v total="${package_total}" -v threshold="${security_sensitive_min_threshold}" 'BEGIN { exit !(threshold + 0 > 0 && total + 0 < threshold + 0) }'; then
				below_threshold_security_sensitive_paths="${below_threshold_security_sensitive_paths}${package_path} coverage ${package_total}% is below required ${security_sensitive_min_threshold}%
"
			fi
		fi
	done

	if [ -n "${missing_security_sensitive_paths}" ]; then
		echo "COVERAGE CHECK ERROR: security-sensitive package(s) have no coverage data:" >&2
		printf '%s' "${missing_security_sensitive_paths}" | while IFS= read -r package_path; do
			[ -z "${package_path}" ] || echo "  - ${package_path}" >&2
		done
		exit 2
	fi

	if [ -n "${below_threshold_security_sensitive_paths}" ]; then
		echo "COVERAGE CHECK ERROR: security-sensitive package(s) are below coverage threshold:" >&2
		printf '%s' "${below_threshold_security_sensitive_paths}" | while IFS= read -r message; do
			[ -z "${message}" ] || echo "  - ${message}" >&2
		done
		exit 1
	fi

	if awk -v threshold="${security_sensitive_min_threshold}" 'BEGIN { exit !(threshold + 0 > 0) }'; then
		echo "security-sensitive coverage data present for ${security_sensitive_count} package path(s), all at or above ${security_sensitive_min_threshold}%"
	else
		echo "security-sensitive coverage data present for ${security_sensitive_count} package path(s)"
	fi
fi

if [ -z "${changed_security_sensitive_paths}" ]; then
	exit 0
fi

missing_changed_security_sensitive_paths=""
below_threshold_changed_security_sensitive_paths=""
changed_security_sensitive_count=0
for package_path in ${changed_security_sensitive_paths}; do
	package_total="$(
		awk -v pkg="${package_path}" '
		NR == 1 { next }
		{
			file = $1
			sub(/:.*/, "", file)
			if (file ~ "^" pkg "/[^/]+\\.go$") {
				statements += $2
				if ($3 > 0) {
					covered += $2
				}
			}
		}
		END {
			if (statements > 0) {
				printf "%.1f", covered * 100 / statements
			}
		}
		' "${coverage_file}"
	)"
	if [ -z "${package_total}" ]; then
		missing_changed_security_sensitive_paths="${missing_changed_security_sensitive_paths}${package_path}
"
	else
		changed_security_sensitive_count=$((changed_security_sensitive_count + 1))
		if awk -v total="${package_total}" -v threshold="${security_sensitive_min_threshold}" 'BEGIN { exit !(threshold + 0 > 0 && total + 0 < threshold + 0) }'; then
			below_threshold_changed_security_sensitive_paths="${below_threshold_changed_security_sensitive_paths}${package_path} coverage ${package_total}% is below required ${security_sensitive_min_threshold}%
"
		fi
	fi
done

if [ -n "${missing_changed_security_sensitive_paths}" ]; then
	echo "COVERAGE CHECK ERROR: changed security-sensitive package(s) have no coverage data:" >&2
	printf '%s' "${missing_changed_security_sensitive_paths}" | while IFS= read -r package_path; do
		[ -z "${package_path}" ] || echo "  - ${package_path}" >&2
	done
	exit 2
fi

if [ -n "${below_threshold_changed_security_sensitive_paths}" ]; then
	echo "COVERAGE CHECK ERROR: changed security-sensitive package(s) are below coverage threshold:" >&2
	printf '%s' "${below_threshold_changed_security_sensitive_paths}" | while IFS= read -r message; do
		[ -z "${message}" ] || echo "  - ${message}" >&2
	done
	exit 1
fi

if awk -v threshold="${security_sensitive_min_threshold}" 'BEGIN { exit !(threshold + 0 > 0) }'; then
	echo "changed security-sensitive coverage data present for ${changed_security_sensitive_count} package path(s), all at or above ${security_sensitive_min_threshold}%"
else
	echo "changed security-sensitive coverage data present for ${changed_security_sensitive_count} package path(s)"
fi
