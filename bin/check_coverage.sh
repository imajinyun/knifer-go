#!/usr/bin/env bash
#
# check_coverage.sh enforces repository-wide and package-level coverage baselines.
# Set COVERAGE_THRESHOLD or PACKAGE_COVERAGE_THRESHOLDS to tune required percentages.

set -euo pipefail

coverage_file="${1:-coverage.out}"
threshold="${COVERAGE_THRESHOLD:-75.2}"
package_thresholds="${PACKAGE_COVERAGE_THRESHOLDS:-github.com/imajinyun/go-knifer/vhttp=75.0 github.com/imajinyun/go-knifer/vresty=65.0 github.com/imajinyun/go-knifer/vconf=75.0 github.com/imajinyun/go-knifer/vzip=80.0 github.com/imajinyun/go-knifer/vcrypto=70.0 github.com/imajinyun/go-knifer/vurl=80.0 github.com/imajinyun/go-knifer/vfile=85.0 github.com/imajinyun/go-knifer/vset=80.0 github.com/imajinyun/go-knifer/vdate=85.0 github.com/imajinyun/go-knifer/vform=90.0 github.com/imajinyun/go-knifer/internal/db=60.0 github.com/imajinyun/go-knifer/internal/obj=90.0 github.com/imajinyun/go-knifer/internal/validator=100.0 github.com/imajinyun/go-knifer/internal/bean=75.0 github.com/imajinyun/go-knifer/internal/net=80.0 github.com/imajinyun/go-knifer/internal/url=80.0 github.com/imajinyun/go-knifer/internal/template=95.0 github.com/imajinyun/go-knifer/internal/httpx/http=75.0 github.com/imajinyun/go-knifer/internal/httpx/resty=75.0 github.com/imajinyun/go-knifer/internal/httpx/internal/shared=80.0}"

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
	exit 0
fi

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
