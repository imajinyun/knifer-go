#!/usr/bin/env bash
#
# check_arch.sh enforces knifer-go's architectural conventions in CI.
#
# Rules checked:
#   1. Every public v* package directory has a doc.go.
#   2. Public v* packages do not import each other (production code only).
#   3. Every public v* production source file imports at least one internal/
#      implementation path, and every imported internal/ path actually exists.
#   4. internal/* implementation packages do not import public v* facades.
#   5. Public package doc.go files contain a package comment.
#   6. New production panics are blocked unless they are known compatibility
#      panics or are in explicit Must/Panic-style APIs.
#   7. Public v* facades stay thin: implementation logic is blocked outside
#      known compatibility shims.
#   8. Public v* production imports do not add new third-party dependencies
#      unless explicitly allowed for that facade.
#   9. Heavy optional integrations stay isolated to their owning facade/internal
#      package families so core utilities do not accidentally pull them in.
#
# It relies on the Go toolchain (go list) for accurate import analysis instead
# of fragile text matching, so it transparently handles abbreviated package
# names (vblf -> internal/bloomfilter), pluralized ones (vmap -> internal/maps),
# and subtrees (vhttp -> internal/httpx/http).
#
# Exit code is non-zero when any rule is violated.

set -euo pipefail

cd "$(dirname "$0")/.."

# Resolve this module's path. In some environments (e.g. a go.work workspace)
# `go list -m` prints multiple modules; pick the one for this directory.
MODULE="$(go list -m 2>/dev/null | grep 'knifer-go' | head -n1)"
if [ -z "${MODULE}" ]; then
	echo "ARCH CHECK ERROR: cannot resolve module path via 'go list -m'" >&2
	exit 2
fi
fail=0

err() {
	echo "ARCH VIOLATION: $*" >&2
	fail=1
}

allowed_facade_external_import() {
	case "$1:$2" in
		verr:github.com/getsentry/sentry-go | \
	verr:github.com/sirupsen/logrus | \
	vpoi:github.com/xuri/excelize/v2 | \
	vresty:resty.dev/v3)
		return 0
		;;
	esac
	return 1
}

allowed_heavy_external_import() {
	case "$1:$2" in
		internal/errx:github.com/getsentry/sentry-go | \
		internal/errx:github.com/sirupsen/logrus | \
		verr:github.com/getsentry/sentry-go | \
		verr:github.com/sirupsen/logrus | \
		internal/httpx/resty:resty.dev/v3 | \
		vresty:resty.dev/v3 | \
		internal/poi:github.com/xuri/excelize/v2 | \
		vpoi:github.com/xuri/excelize/v2)
			return 0
			;;
	esac
	case "$1:$2" in
		internal/imgx:github.com/makiuchi-d/gozxing*)
			return 0
			;;
	esac
	return 1
}

is_heavy_external_import() {
	case "$1" in
		github.com/getsentry/sentry-go | \
		github.com/sirupsen/logrus | \
		github.com/xuri/excelize/v2 | \
		resty.dev/v3 | \
		github.com/makiuchi-d/gozxing*)
			return 0
			;;
	esac
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

# Collect public package directories (top-level v* dirs containing .go files).
for dir in v*/; do
	pkg="${dir%/}"
	# Skip directories without Go files.
	if ! ls "${pkg}"/*.go >/dev/null 2>&1; then
		continue
	fi

	# Rule 1: doc.go must exist.
	if [ ! -f "${pkg}/doc.go" ]; then
		err "${pkg}: missing doc.go"
	fi

	# Gather this package's production (non-test) imports via the Go toolchain.
	imports="$(go list -f '{{range .Imports}}{{println .}}{{end}}' "./${pkg}")"

	# Rule 2: must not import another public v* package.
	while IFS= read -r imp; do
		[ -z "${imp}" ] && continue
		case "${imp}" in
		"${MODULE}"/v*)
			err "${pkg}: imports another public package ${imp} (v* packages must not depend on each other)"
			;;
		esac
	done <<<"${imports}"

	# Rule 3: package-level imports must refer only to existing internal implementations.
	internal_count=0
	while IFS= read -r imp; do
		[ -z "${imp}" ] && continue
		case "${imp}" in
		"${MODULE}"/internal/*)
			internal_count=$((internal_count + 1))
			rel="${imp#"${MODULE}"/}"
			if [ ! -d "${rel}" ]; then
				err "${pkg}: imports non-existent internal path ${imp}"
			fi
			;;
		"${MODULE}"*)
			;;
		*)
			if is_external_import "${imp}" && ! allowed_facade_external_import "${pkg}" "${imp}"; then
				err "${pkg}: imports third-party dependency ${imp} (facade dependency surface must be allowlisted)"
			fi
			;;
		esac
	done <<<"${imports}"


	# Rule 3 (per-file): every public facade source file must delegate directly to
	# internal/. This catches accidental standalone logic in a v* file even when a
	# sibling file still imports internal/ and keeps the package-level check green.
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

# Rule 4: internal implementation packages must not depend on public facades.
# This keeps the dependency direction one-way: v* -> internal/*, never back up.
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

# Rule 9: heavyweight optional integrations must not bleed into core packages.
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

# Rules 5 and 6 use lightweight source checks. They intentionally complement
# go vet/golangci-lint by encoding project-specific architecture policy.
python3 - <<'PY' || fail=1
from __future__ import annotations

import pathlib
import re
import sys

root = pathlib.Path.cwd()
violations: list[str] = []

package_comment = re.compile(r"(?m)^//\s+Package\s+\w+")


def check_package_docs() -> None:
	files = [root / "doc.go"] + sorted(root.glob("v*/doc.go"))
	for path in files:
		if not path.exists():
			continue
		if not package_comment.search(path.read_text()):
			rel = path.relative_to(root)
			violations.append(f"{rel}: doc.go must contain a package comment starting with 'Package <name>'")


allowed_panic_paths = {
	# Compatibility panics from constructors or dynamic adapters. Prefer adding
	# error-returning APIs for new call sites instead of extending this list.
	"internal/bloomfilter/bitset_bloomfilter.go",
	"internal/bloomfilter/filter.go",
	"internal/cron/pattern.go",
	# DB.Tx intentionally rolls back and rethrows user callback panics to preserve
	# standard transaction-boundary panic semantics.
	"internal/db/db.go",
	"internal/errx/exit.go",
	"internal/job/map.go",
	"internal/jwt/jwt.go",
	"internal/jwt/signer.go",
	"internal/jwt/signer_util.go",
	"internal/maps/maps.go",
	"internal/obj/serialize.go",
	"internal/semaphore/semaphore.go",
}

allowed_panic_funcs = re.compile(r"^(?:func\s+(?:\([^)]*\)\s*)?(?:Must|Panic)\w*\b)")
unsafe_ref_access = re.compile(r"fieldAccessConfig\{\s*unsafeAccess:\s*true\s*\}")
facade_logic = re.compile(r"^(?:if|for|switch|select|defer|go)\b|:=")

allowed_facade_logic_paths = {
	# Existing compatibility shims that predate the thin-facade policy. Keep new
	# facade files as aliases/delegates and move implementation details into
	# internal/* instead of extending this list.
	# TODO(v2): migrate these remaining compatibility shims into internal/* so the
	# public v* surface is uniformly aliases, type aliases, and delegates.
	"vcache/cache.go",
	"vjob/job.go",
	"vnum/arith.go",
	"vrand/rand.go",
	"vset/set.go",
	"vskt/socket.go",
	"vxml/element.go",
}


def enclosing_func(lines: list[str], idx: int) -> str:
	for j in range(idx, -1, -1):
		line = lines[j].strip()
		if line.startswith("func "):
			return line
	return ""


def check_panic_policy() -> None:
	for path in sorted(root.glob("**/*.go")):
		if path.name.endswith("_test.go") or "/.git/" in str(path):
			continue
		rel = path.relative_to(root).as_posix()
		lines = path.read_text().splitlines()
		for i, line in enumerate(lines):
			if "panic(" not in line:
				continue
			fn = enclosing_func(lines, i)
			if allowed_panic_funcs.match(fn):
				continue
			if rel in allowed_panic_paths:
				continue
			violations.append(f"{rel}:{i + 1}: production panic is not allowed outside known compatibility or Must/Panic-style APIs")


def check_ref_unsafe_opt_in() -> None:
	path = root / "internal/ref/ref.go"
	if not path.exists():
		return
	for i, line in enumerate(path.read_text().splitlines(), start=1):
		if unsafe_ref_access.search(line) and "call(" not in enclosing_func(path.read_text().splitlines(), i - 1):
			violations.append(f"internal/ref/ref.go:{i}: unsafe field access must require explicit FieldAccessOption opt-in")


def check_facade_boundary_policy() -> None:
	for path in sorted(root.glob("v*/*.go")):
		if path.name == "doc.go" or path.name.endswith("_test.go"):
			continue
		rel = path.relative_to(root).as_posix()
		if rel in allowed_facade_logic_paths:
			continue
		for i, raw in enumerate(path.read_text().splitlines(), start=1):
			line = raw.strip()
			if not line or line.startswith("//"):
				continue
			if facade_logic.search(line):
				violations.append(
					f"{rel}:{i}: facade packages should not contain implementation control flow or local state; move logic to internal/*"
				)


check_package_docs()
check_panic_policy()
check_ref_unsafe_opt_in()
check_facade_boundary_policy()

for violation in violations:
	print(f"ARCH VIOLATION: {violation}", file=sys.stderr)

sys.exit(1 if violations else 0)
PY

if [ "${fail}" -ne 0 ]; then
	echo "" >&2
	echo "Architecture check failed. See violations above." >&2
	exit 1
fi

echo "Architecture check passed."
