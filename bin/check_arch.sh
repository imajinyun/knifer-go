#!/usr/bin/env bash
#
# check_arch.sh composes the project architecture gates.
#
# Rule ownership lives in focused sub-gates so failures point at the boundary
# that actually drifted:
#   - check_provider_contracts.sh
#   - check_arch_imports.sh
#   - check_panic_policy.sh
#   - check_facade_boundary.sh

set -euo pipefail

cd "$(dirname "$0")/.."

run_gate() {
	name="$1"
	shift
	echo "arch: running ${name}"
	"$@"
}

run_gate "provider contract gate" bash bin/check_provider_contracts.sh
run_gate "module resolve, facade import scan, internal import scan, heavy dependency scan" bash bin/check_arch_imports.sh
run_gate "panic policy scan" bash bin/check_panic_policy.sh
run_gate "facade boundary scan" bash bin/check_facade_boundary.sh

echo "Architecture check passed."
