#!/usr/bin/env bash
#
# check_docs_quickstart.sh validates the human-authored facade quickstart
# structure. ai-context.json is the source of truth for public facade packages.

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
AI_CONTEXT="${ROOT_DIR}/ai-context.json"
DOC_DIR="${ROOT_DIR}/docs/doc"
DOC_INDEX="${DOC_DIR}/README.md"

python3 - "${ROOT_DIR}" "${AI_CONTEXT}" "${DOC_DIR}" "${DOC_INDEX}" <<'PY'
import json
import os
import re
import sys

root_dir, ai_context, doc_dir, doc_index = sys.argv[1:]
errors = []


def add_error(message):
    errors.append(message)


with open(ai_context, "r", encoding="utf-8") as f:
    data = json.load(f)

public_facades = data.get("public_facades", [])
if not isinstance(public_facades, list):
    add_error("ai-context.json public_facades must be a list")
    public_facades = []

try:
    with open(doc_index, "r", encoding="utf-8") as f:
        index_text = f.read()
except FileNotFoundError:
    index_text = ""
    add_error("docs/doc/README.md is missing")

doc_files = [
    name
    for name in os.listdir(doc_dir)
    if re.match(r"^\d{2}-v[a-z0-9]+\.md$", name)
]
docs_by_package = {}
for name in doc_files:
    package = re.sub(r"^\d{2}-", "", name).removesuffix(".md")
    docs_by_package.setdefault(package, []).append(name)

required_literal_sections = [
    "## Which helper should I use?",
    "## Related packages",
    "## Benchmarks and trade-offs",
    "## FAQ",
]

checklist_pattern = re.compile(r"^## .*checklist$|^## Safety notes$", re.MULTILINE)

for entry in public_facades:
    package = entry.get("package")
    if not isinstance(package, str) or not package:
        add_error(f"invalid public_facades entry: {entry!r}")
        continue

    matches = docs_by_package.get(package, [])
    if not matches:
        add_error(f"missing quickstart doc for {package}")
        continue
    if len(matches) > 1:
        add_error(f"multiple quickstart docs for {package}: {', '.join(sorted(matches))}")
        continue

    filename = matches[0]
    path = os.path.join(doc_dir, filename)
    with open(path, "r", encoding="utf-8") as f:
        text = f.read()

    title_patterns = [
        rf"^# {re.escape(package)} Quickstart\s*$",
        rf"^# {re.escape(package)}: .+\s*$",
    ]
    if not any(re.search(pattern, text, flags=re.MULTILINE) for pattern in title_patterns):
        add_error(f"{filename} must start with '# {package} Quickstart' or an approved adapter title")

    for section in required_literal_sections:
        if section not in text:
            add_error(f"{filename} is missing required section {section!r}")

    if not re.search(r"^## When not to use", text, flags=re.MULTILINE):
        add_error(f"{filename} is missing required section '## When not to use ...'")

    if not checklist_pattern.search(text):
        add_error(f"{filename} is missing a checklist section")

    if text.count("```") % 2 != 0:
        add_error(f"{filename} has unbalanced fenced code blocks")

    if index_text and f"]({filename})" not in index_text:
        add_error(f"docs/doc/README.md does not link to {filename}")

known_packages = {entry.get("package") for entry in public_facades if isinstance(entry.get("package"), str)}
extra_docs = sorted(set(docs_by_package) - known_packages)
if extra_docs:
    add_error("quickstart docs exist for unknown facade package(s): " + ", ".join(extra_docs))

if errors:
    for error in errors:
        print(f"docs quickstart check error: {error}", file=sys.stderr)
    sys.exit(1)

print(f"quickstart docs are valid ({len(public_facades)} public facades)")
PY
