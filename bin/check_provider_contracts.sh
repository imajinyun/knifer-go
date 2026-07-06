#!/usr/bin/env bash
set -euo pipefail

if [ -n "${PROVIDER_CONTRACT_ROOT:-}" ]; then
	cd "${PROVIDER_CONTRACT_ROOT}"
else
	cd "$(dirname "$0")/.."
fi

python3 - <<'PY'
from __future__ import annotations

import json
import pathlib
import re
import sys

root = pathlib.Path.cwd()
violations: list[str] = []

with (root / "ai-context.json").open(encoding="utf-8") as f:
	ai_context = json.load(f)

facade_to_internal = {
	entry["package"]: entry["internal"].rstrip("/")
	for entry in ai_context.get("public_facades", [])
	if isinstance(entry, dict) and "package" in entry and "internal" in entry
}
providers = ai_context.get("dependency_tiers", {}).get("provider_contract_facades", [])
if not isinstance(providers, list):
	violations.append("ai-context.json dependency_tiers.provider_contract_facades must be a list")
	providers = []

for facade in providers:
	internal = facade_to_internal.get(facade)
	if not internal:
		violations.append(f"provider contract facade {facade}: missing public_facades internal mapping")
		continue
	paths: list[pathlib.Path] = []
	for directory in (root / facade, root / internal):
		if not directory.is_dir():
			violations.append(f"provider contract facade {facade}: missing directory {directory.relative_to(root)}")
			continue
		paths.extend(path for path in directory.glob("*.go") if not path.name.endswith("_test.go"))
	combined = "\n".join(path.read_text() for path in paths)
	if not re.search(r"type\s+\w*Provider\s+interface\s*{", combined):
		violations.append(f"provider contract facade {facade}: must define a Provider interface contract")
	for path in paths:
		rel = path.relative_to(root).as_posix()
		text = path.read_text()
		for forbidden_import in ('"net/http"', '"resty.dev/', '"google.golang.org/grpc', '"golang.org/x/oauth2'):
			if forbidden_import in text:
				violations.append(f"{rel}: provider contract packages must not import concrete provider/network SDK dependency {forbidden_import}")
		for forbidden_call in ("os.Getenv", "os.ReadFile", "http.NewRequest", "http.Client", "net.Dial", "grpc.Dial"):
			if forbidden_call in text:
				violations.append(f"{rel}: provider contract packages must not read credentials, touch local files, or open network connections directly ({forbidden_call})")

if violations:
	for violation in violations:
		print(f"PROVIDER CONTRACT VIOLATION: {violation}", file=sys.stderr)
	sys.exit(1)

print(f"provider contract governance is valid ({len(providers)} facades)")
PY
