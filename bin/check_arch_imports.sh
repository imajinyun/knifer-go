#!/usr/bin/env bash
set -euo pipefail

if [ -n "${ARCH_CHECK_ROOT:-}" ]; then
	cd "${ARCH_CHECK_ROOT}"
else
	cd "$(dirname "$0")/.."
fi

echo "arch imports: resolving module"
MODULE="$(go list -m 2>/dev/null | grep 'knifer-go' | head -n1)"
if [ -z "${MODULE}" ]; then
	echo "ARCH IMPORT CHECK ERROR: cannot resolve module path via 'go list -m'" >&2
	exit 2
fi

fail=0

err() {
	echo "ARCH IMPORT VIOLATION: $*" >&2
	fail=1
}

HEAVY_ALLOWLIST_TSV="$(mktemp)"
trap 'rm -f "${HEAVY_ALLOWLIST_TSV}"' EXIT
python3 - "${HEAVY_ALLOWLIST_TSV}" <<'PY'
import json
import sys

out_path = sys.argv[1]
with open("ai-context.json", "r", encoding="utf-8") as f:
    data = json.load(f)

allowlist = data.get("dependency_tiers", {}).get("heavy_dependency_allowlist", {})
with open(out_path, "w", encoding="utf-8") as f:
    for import_path, prefixes in sorted(allowlist.items()):
        if not isinstance(import_path, str) or not isinstance(prefixes, list):
            continue
        for prefix in prefixes:
            if isinstance(prefix, str):
                f.write(import_path + "\t" + prefix + "\n")
PY

import_pattern_matches() {
	pattern="$1"
	import_path="$2"
	case "${pattern}" in
	*"*"*)
		case "${import_path}" in
		${pattern})
			return 0
			;;
		esac
		;;
	*)
		[ "${pattern}" = "${import_path}" ] && return 0
		;;
	esac
	return 1
}

allowed_heavy_external_import() {
	rel="$1"
	import_path="$2"
	while IFS="$(printf '\t')" read -r pattern prefix; do
		[ -z "${pattern}" ] && continue
		if import_pattern_matches "${pattern}" "${import_path}"; then
			case "${rel}" in
			"${prefix}" | "${prefix}"/*)
				return 0
				;;
			esac
		fi
	done <"${HEAVY_ALLOWLIST_TSV}"
	return 1
}

is_heavy_external_import() {
	import_path="$1"
	while IFS="$(printf '\t')" read -r pattern _prefix; do
		[ -z "${pattern}" ] && continue
		if import_pattern_matches "${pattern}" "${import_path}"; then
			return 0
		fi
	done <"${HEAVY_ALLOWLIST_TSV}"
	return 1
}

is_external_import() {
	first="${1%%/*}"
	case "${first}" in
	*.*)
		return 0
		;;
	esac
	return 1
}

echo "arch imports: scanning public facades"
for dir in v*/; do
	pkg="${dir%/}"
	if ! ls "${pkg}"/*.go >/dev/null 2>&1; then
		continue
	fi
	if [ ! -f "${pkg}/doc.go" ]; then
		err "${pkg}: missing doc.go"
	fi
	imports="$(go list -f '{{range .Imports}}{{println .}}{{end}}' "./${pkg}")"
	while IFS= read -r imp; do
		[ -z "${imp}" ] && continue
		case "${imp}" in
		"${MODULE}"/v*)
			err "${pkg}: imports another public package ${imp} (v* packages must not depend on each other)"
			;;
		"${MODULE}"/internal/*)
			rel="${imp#"${MODULE}"/}"
			if [ ! -d "${rel}" ]; then
				err "${pkg}: imports non-existent internal path ${imp}"
			fi
			;;
		"${MODULE}"*)
			;;
		*)
			if is_external_import "${imp}" && ! allowed_heavy_external_import "${pkg}" "${imp}"; then
				err "${pkg}: imports third-party dependency ${imp} (facade dependency surface must be allowlisted)"
			fi
			;;
		esac
	done <<<"${imports}"

	for file in "${pkg}"/*.go; do
		base="$(basename "${file}")"
		case "${base}" in
		doc.go|*_test.go)
			continue
			;;
		esac
		file_imports="$(go list -f '{{range .Imports}}{{println .}}{{end}}' "${file}")"
		file_internal_count=0
		while IFS= read -r imp; do
			[ -z "${imp}" ] && continue
			case "${imp}" in
			"${MODULE}"/internal/*)
				file_internal_count=$((file_internal_count + 1))
				;;
			esac
		done <<<"${file_imports}"
		if [ "${file_internal_count}" -eq 0 ]; then
			err "${file}: does not import any internal/ implementation (each facade source file must delegate to internal)"
		fi
	done
done

echo "arch imports: scanning internal packages"
while IFS= read -r pkg; do
	[ -z "${pkg}" ] && continue
	imports="$(go list -f '{{range .Imports}}{{println .}}{{end}}' "${pkg}")"
	while IFS= read -r imp; do
		[ -z "${imp}" ] && continue
		case "${imp}" in
		"${MODULE}"/v*)
			err "${pkg#${MODULE}/}: imports public facade ${imp} (internal packages must not depend on v* packages)"
			;;
		esac
	done <<<"${imports}"
done < <(go list ./internal/... 2>/dev/null)

echo "arch imports: scanning heavy dependency isolation"
while IFS= read -r pkg; do
	[ -z "${pkg}" ] && continue
	rel="${pkg#${MODULE}/}"
	imports="$(go list -f '{{range .Imports}}{{println .}}{{end}}' "${pkg}")"
	while IFS= read -r imp; do
		[ -z "${imp}" ] && continue
		if is_heavy_external_import "${imp}" && ! allowed_heavy_external_import "${rel}" "${imp}"; then
			err "${rel}: imports heavy optional dependency ${imp} outside its isolated package family"
		fi
	done <<<"${imports}"
done < <(go list ./internal/... ./v... 2>/dev/null)

if [ "${fail}" -ne 0 ]; then
	echo "Architecture import check failed. See violations above." >&2
	exit 1
fi

echo "architecture import governance passed"
