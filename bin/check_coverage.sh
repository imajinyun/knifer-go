#!/usr/bin/env bash
#
# check_coverage.sh enforces repository-wide and package-level coverage baselines.
# ai-context.json is the source of truth for default thresholds.
# Set COVERAGE_THRESHOLD or PACKAGE_COVERAGE_THRESHOLDS to override defaults locally.

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
AI_CONTEXT="${ROOT_DIR}/ai-context.json"

coverage_file="${1:-coverage.out}"
coverage_config="$(
	python3 - "${AI_CONTEXT}" <<'PY'
import json
import os
import sys

ai_context = sys.argv[1]
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
module = data["project"]["module"]
facade_to_internal = {
    entry["package"]: entry["internal"].rstrip("/")
    for entry in data["public_facades"]
}
security_sensitive_paths = set()
for package in data["security_sensitive_packages"]:
    package_dir = package.rstrip("/")
    if has_statement_source(package_dir):
        security_sensitive_paths.add(f"{module}/{package_dir}")
    internal = facade_to_internal.get(package)
    if internal and has_statement_source(internal):
        security_sensitive_paths.add(f"{module}/{internal}")
security_sensitive_paths = " ".join(sorted(security_sensitive_paths))
print(f"{repository_threshold:.1f}|{package_thresholds}|{security_sensitive_paths}|{security_sensitive_min_threshold:.1f}")
PY
)"

IFS='|' read -r metadata_threshold metadata_package_thresholds metadata_security_sensitive_paths metadata_security_sensitive_min_threshold <<<"${coverage_config}"
threshold="${COVERAGE_THRESHOLD:-${metadata_threshold}}"
package_thresholds="${PACKAGE_COVERAGE_THRESHOLDS:-${metadata_package_thresholds}}"
security_sensitive_paths="${SECURITY_SENSITIVE_COVERAGE_PATHS:-${metadata_security_sensitive_paths}}"
security_sensitive_min_threshold="${SECURITY_SENSITIVE_MIN_COVERAGE_THRESHOLD:-${metadata_security_sensitive_min_threshold}}"

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

if [ -z "${package_thresholds}" ]; then
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

if [ -z "${security_sensitive_paths}" ]; then
	exit 0
fi

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
