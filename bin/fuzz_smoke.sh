#!/usr/bin/env bash
set -euo pipefail

go_cmd="${GO:-go}"
fuzz_time="${FUZZTIME:-1s}"
if [ "$#" -eq 0 ]; then
	set -- ./...
fi

# Discover from compiled tests so build constraints and external test packages
# match the selected toolchain. A discovery failure must fail the gate.
listing="$("$go_cmd" test -list '^Fuzz' "$@")"
targets="$(printf '%s\n' "$listing" | awk '
	/^Fuzz[^[:space:]]*$/ { names[++count] = $0 }
	/^ok[[:space:]]/ {
		for (i = 1; i <= count; i++) print $2, names[i]
		count = 0
	}
')"
if [ -z "$targets" ]; then
	echo "fuzz smoke: no Fuzz targets found in selected packages" >&2
	exit 1
fi

count=0
while read -r package target; do
	echo "fuzz smoke: $package $target ($fuzz_time)"
	"$go_cmd" test -run '^$' -fuzz "^${target}$" -fuzztime="$fuzz_time" "$package"
	count=$((count + 1))
done <<< "$targets"
echo "fuzz smoke: passed $count target(s)"
