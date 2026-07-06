#!/usr/bin/env bash
#
# check_coverage.sh enforces repository-wide and package-level coverage baselines.
# ai-context.json is the source of truth for default thresholds.
# Set COVERAGE_THRESHOLD or PACKAGE_COVERAGE_THRESHOLDS to override defaults locally.

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
coverage_file="${1:-coverage.out}"

cd "${ROOT_DIR}"
go run ./bin/coveragecheck -root "${ROOT_DIR}" "${coverage_file}"
