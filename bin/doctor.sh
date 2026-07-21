#!/usr/bin/env bash
set -uo pipefail

root="${DOCTOR_ROOT:-$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)}"
go_cmd="${GO:-go}"
git_cmd="${GIT:-git}"
golangci_lint_cmd="${GOLANGCI_LINT:-golangci-lint}"
tmp_dir="$(mktemp -d "${TMPDIR:-/tmp}/knifer-go-doctor.XXXXXX")"
trap 'rm -rf "${tmp_dir}"' EXIT

failures=0

run_required() {
	local name="$1"
	local success_message="$2"
	local show_stdout="$3"
	shift 3

	local file_key
	file_key="$(printf '%s' "${name}" | tr -c '[:alnum:]_.-' '_')"
	local stdout_file="${tmp_dir}/${file_key}.stdout"
	local stderr_file="${tmp_dir}/${file_key}.stderr"
	local exit_code=0

	if "$@" >"${stdout_file}" 2>"${stderr_file}"; then
		exit_code=0
	else
		exit_code=$?
	fi

	if [ -s "${stdout_file}" ] && { [ "${show_stdout}" = "true" ] || [ "${exit_code}" -ne 0 ] || [ -s "${stderr_file}" ]; }; then
		cat "${stdout_file}"
	fi
	if [ -s "${stderr_file}" ]; then
		cat "${stderr_file}" >&2
	fi

	if [ "${exit_code}" -ne 0 ]; then
		echo "DOCTOR ERROR: ${name} exited with status ${exit_code}" >&2
		failures=$((failures + 1))
		return
	fi
	if [ -s "${stderr_file}" ]; then
		echo "DOCTOR ERROR: ${name} wrote to stderr" >&2
		failures=$((failures + 1))
		return
	fi
	if [ -n "${success_message}" ]; then
		echo "${success_message}"
	fi
}

echo "== go =="
run_required "go version" "" true "${go_cmd}" version

echo "== python3 =="
if command -v python3 >/dev/null 2>&1; then
	python3 --version
else
	echo "python3 not found"
fi

echo "== golangci-lint =="
if command -v "${golangci_lint_cmd}" >/dev/null 2>&1; then
	"${golangci_lint_cmd}" version
else
	echo "${golangci_lint_cmd} not found"
fi

echo "== govulncheck =="
govuln_stdout="${tmp_dir}/govulncheck.stdout"
govuln_stderr="${tmp_dir}/govulncheck.stderr"
if "${go_cmd}" tool govulncheck -version >"${govuln_stdout}" 2>"${govuln_stderr}"; then
	cat "${govuln_stdout}"
	if [ -s "${govuln_stderr}" ]; then
		cat "${govuln_stderr}" >&2
		echo "DOCTOR ERROR: go tool govulncheck -version wrote to stderr" >&2
		failures=$((failures + 1))
	fi
else
	cat "${govuln_stderr}" >&2
	echo "govulncheck tool not available; run: go get -tool golang.org/x/vuln/cmd/govulncheck"
fi

echo "== git status =="
run_required "git status" "" true "${git_cmd}" -C "${root}" status --short --branch

echo "== module =="
run_required "go list -m" "" true env GOWORK=off "${go_cmd}" -C "${root}" list -m

echo "== module cache =="
run_required "go module cache check" "" true env GO="${go_cmd}" bash "${root}/bin/check_go_module_cache.sh"

echo "== package list =="
run_required "go list ./..." "go list ./... OK" false "${go_cmd}" -C "${root}" list -buildvcs=false ./...

if [ "${failures}" -ne 0 ]; then
	echo "doctor found ${failures} required diagnostic failure(s)" >&2
	exit 1
fi

echo "doctor diagnostics passed"
