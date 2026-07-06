#!/usr/bin/env bash
#
# check_docs_quickstart.sh validates the human-authored facade quickstart
# structure. ai-context.json is the source of truth for public facade packages.

set -euo pipefail

if [ -n "${DOCS_QUICKSTART_ROOT:-}" ]; then
	ROOT_DIR="${DOCS_QUICKSTART_ROOT}"
else
	ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
fi
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

docs_quality_profiles = data.get("docs_quality_profiles", {})
profile_packages = docs_quality_profiles.get("packages", {}) if isinstance(docs_quality_profiles, dict) else {}
allowed_profiles = set(docs_quality_profiles.get("allowed_profiles", [])) if isinstance(docs_quality_profiles, dict) else set()

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
    "## Golden path APIs",
    "## Which helper should I use?",
    "## Related packages",
    "## Benchmarks and trade-offs",
    "## FAQ",
]

checklist_pattern = re.compile(r"^## .*checklist$|^## Safety notes$", re.MULTILINE)
go_fence_pattern = re.compile(r"```go\n(.*?)\n```", re.DOTALL)
related_package_pattern = re.compile(r"^- Use `v[a-z0-9]+` ", re.MULTILINE)
package_main_pattern = re.compile(r"^package\s+main\b", re.MULTILINE)
import_block_pattern = re.compile(r"import\s*\((.*?)\)", re.DOTALL)
single_import_pattern = re.compile(r"import\s+\"([^\"]+)\"")


def collect_imports(block):
    imports = set(single_import_pattern.findall(block))
    for import_block in import_block_pattern.findall(block):
        for line in import_block.splitlines():
            line = line.strip()
            match = re.search(r'"([^"]+)"', line)
            if match:
                imports.add(match.group(1))
    return imports


def require_phrase(filename, text, phrases, profile, purpose):
    lower_text = text.lower()
    if not any(phrase in lower_text for phrase in phrases):
        add_error(f"{filename} profile {profile} must document {purpose}")


known_packages = {entry.get("package") for entry in public_facades if isinstance(entry.get("package"), str)}
if not isinstance(docs_quality_profiles, dict):
    add_error("ai-context.json docs_quality_profiles must be an object")
if not allowed_profiles:
    add_error("ai-context.json docs_quality_profiles.allowed_profiles must be non-empty")
if not isinstance(profile_packages, dict):
    add_error("ai-context.json docs_quality_profiles.packages must be an object")
    profile_packages = {}

profile_package_names = set(profile_packages)
missing_profile_packages = sorted(known_packages - profile_package_names)
extra_profile_packages = sorted(profile_package_names - known_packages)
if missing_profile_packages:
    add_error("docs_quality_profiles.packages missing facade package(s): " + ", ".join(missing_profile_packages))
if extra_profile_packages:
    add_error("docs_quality_profiles.packages contains unknown facade package(s): " + ", ".join(extra_profile_packages))

for entry in public_facades:
    package = entry.get("package")
    if not isinstance(package, str) or not package:
        add_error(f"invalid public_facades entry: {entry!r}")
        continue
    profiles = profile_packages.get(package, [])
    if not isinstance(profiles, list) or not profiles:
        add_error(f"docs_quality_profiles.packages.{package} must contain at least one profile")
        profiles = ["error_returning"]
    profiles = [profile for profile in profiles if isinstance(profile, str)]
    unknown_profiles = sorted(set(profiles) - allowed_profiles)
    if unknown_profiles:
        add_error(f"docs_quality_profiles.packages.{package} contains unknown profile(s): " + ", ".join(unknown_profiles))
    if "error_returning" in profiles and "no_error_returning" in profiles:
        add_error(f"docs_quality_profiles.packages.{package} cannot combine error_returning and no_error_returning")

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

    if not related_package_pattern.search(text):
        add_error(f"{filename} must include at least one related-package bullet using 'Use `v...`'")

    if "Prefer" not in text and "Use" not in text:
        add_error(f"{filename} helper guidance must include explicit use/prefer wording")

    lower_text = text.lower()
    if "error_returning" in profiles and not any(term in lower_text for term in ("error", "errors.is", "panic(err)", "err != nil")):
        add_error(f"{filename} must document error behavior or explicitly show error handling")
    if "no_error_returning" in profiles and not any(term in lower_text for term in ("no error", "do not return errors", "does not return errors", "return errors?")):
        add_error(f"{filename} profile no_error_returning must explicitly state that facade helpers do not return errors")
    if "security_sensitive" in profiles:
        require_phrase(
            filename,
            text,
            ("untrusted", "trust boundary", "safety checklist", "safe ", "security", "credential", "secret"),
            "security_sensitive",
            "trust-boundary or safe-input guidance",
        )
    if "provider_contract" in profiles:
        require_phrase(
            filename,
            text,
            ("provider injection", "injected provider", "no built-in", "does not read", "does not open", "credential"),
            "provider_contract",
            "provider injection and no-default-provider boundaries",
        )
    if "heavy_extension" in profiles:
        require_phrase(
            filename,
            text,
            ("dependency", "trade-off", "optional", "directly", "benchmark"),
            "heavy_extension",
            "dependency or trade-off guidance",
        )

    if not re.search(r"^## When not to use", text, flags=re.MULTILINE):
        add_error(f"{filename} is missing required section '## When not to use ...'")

    if not checklist_pattern.search(text):
        add_error(f"{filename} is missing a checklist section")

    if text.count("```") % 2 != 0:
        add_error(f"{filename} has unbalanced fenced code blocks")

    go_blocks = go_fence_pattern.findall(text)
    runnable_blocks = [
        block
        for block in go_blocks
        if package_main_pattern.search(block)
    ]
    runnable_facade_blocks = [
        block
        for block in runnable_blocks
        if f"github.com/imajinyun/knifer-go/{package}" in collect_imports(block)
    ]
    if go_blocks and not runnable_facade_blocks:
        add_error(f"{filename} must include at least one runnable package main example that imports {package}")
    for block_index, block in enumerate(runnable_facade_blocks, start=1):
        if "func main()" not in block:
            add_error(f"{filename} runnable facade example {block_index} must define func main()")
    if runnable_facade_blocks and not any(
        "fmt.Println" in block or "fmt.Printf" in block or "panic(err)" in block
        for block in runnable_facade_blocks
    ):
        add_error(f"{filename} must include at least one runnable facade example with observable output or explicit error handling")

    if index_text and f"]({filename})" not in index_text:
        add_error(f"docs/doc/README.md does not link to {filename}")

extra_docs = sorted(set(docs_by_package) - known_packages)
if extra_docs:
    add_error("quickstart docs exist for unknown facade package(s): " + ", ".join(extra_docs))

if errors:
    for error in errors:
        print(f"docs quickstart check error: {error}", file=sys.stderr)
    sys.exit(1)

print(f"quickstart docs are valid ({len(public_facades)} public facades)")
PY
