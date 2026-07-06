#!/usr/bin/env bash
#
# check_release_notes.sh validates release-note readiness before tags are
# published. CHANGELOG.md is the source of truth for human-authored notes.

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
CHANGELOG_FILE="${CHANGELOG_FILE:-${ROOT_DIR}/CHANGELOG.md}"
GOVERNANCE_RELEASE_TEMPLATE_FILE="${GOVERNANCE_RELEASE_TEMPLATE_FILE:-${ROOT_DIR}/docs/doc/adoption-trust.md}"
RELEASE_VERSION="${RELEASE_VERSION:-${1:-}}"

python3 - "${CHANGELOG_FILE}" "${GOVERNANCE_RELEASE_TEMPLATE_FILE}" "${RELEASE_VERSION}" <<'PY'
import os
import re
import sys

changelog_file, governance_template_file, release_version = sys.argv[1], sys.argv[2], sys.argv[3].strip()
errors = []


def add_error(message):
    errors.append(message)


try:
    with open(changelog_file, "r", encoding="utf-8") as f:
        text = f.read()
except FileNotFoundError:
    print(f"release notes check error: {changelog_file} does not exist", file=sys.stderr)
    sys.exit(1)

try:
    with open(governance_template_file, "r", encoding="utf-8") as f:
        governance_template = f.read()
except FileNotFoundError:
    add_error(f"governance release summary template file {governance_template_file} does not exist")
    governance_template = ""

if not re.search(r"^## Unreleased\s*$", text, flags=re.MULTILINE):
    add_error("CHANGELOG.md must contain a '## Unreleased' section")

for token in (
    "## Governance Validation Contracts",
    "Contract changed:",
    "User impact:",
    "Required action:",
    "Validation evidence:",
    "Compatibility note:",
):
    if governance_template and token not in governance_template:
        add_error(f"governance release summary template must include {token!r}")

version_headings = re.findall(r"^## \[?([0-9]+\.[0-9]+\.[0-9]+)\]?(?:[\s-]|$).*", text, flags=re.MULTILINE)
if len(version_headings) != len(set(version_headings)):
    add_error("CHANGELOG.md contains duplicate release version headings")

if release_version:
    if release_version.startswith("v"):
        add_error("RELEASE_VERSION must not include the leading 'v'")
        release_version = release_version[1:]
    if not re.fullmatch(r"[0-9]+\.[0-9]+\.[0-9]+", release_version):
        add_error(f"RELEASE_VERSION must be MAJOR.MINOR.PATCH, got {release_version!r}")
    heading_pattern = rf"^## \[?{re.escape(release_version)}\]?(?:[\s-]|$)"
    if not re.search(heading_pattern, text, flags=re.MULTILINE):
        add_error(f"CHANGELOG.md is missing release notes for {release_version}")
    unreleased_match = re.search(r"^## Unreleased\s*$(.*?)(?=^##\s+|\Z)", text, flags=re.MULTILINE | re.DOTALL)
    if unreleased_match and re.search(r"^###\s+|^-\s+", unreleased_match.group(1), flags=re.MULTILINE):
        add_error("CHANGELOG.md still has entries under Unreleased; move release notes into the version section before tagging")

if errors:
    for error in errors:
        print(f"release notes check error: {error}", file=sys.stderr)
    sys.exit(1)

if release_version:
    print(f"release notes are valid for {release_version}")
else:
    print("release notes structure is valid")
PY
