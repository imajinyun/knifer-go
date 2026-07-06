#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")/.."

echo "facade boundary: scanning docs, unsafe opt-in, and thin facade rules"
python3 - <<'PY'
from __future__ import annotations

import pathlib
import re
import sys

root = pathlib.Path.cwd()
violations: list[str] = []

package_comment = re.compile(r"(?m)^//\s+Package\s+\w+")
unsafe_ref_access = re.compile(r"fieldAccessConfig\{\s*unsafeAccess:\s*true\s*\}")
facade_logic = re.compile(r"^(?:if|for|switch|select|defer|go)\b|:=")

allowed_facade_logic_paths = {
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


for path in [root / "doc.go"] + sorted(root.glob("v*/doc.go")):
	if not path.exists():
		continue
	if not package_comment.search(path.read_text()):
		rel = path.relative_to(root)
		violations.append(f"{rel}: doc.go must contain a package comment starting with 'Package <name>'")

ref_path = root / "internal/ref/ref.go"
if ref_path.exists():
	lines = ref_path.read_text().splitlines()
	for i, line in enumerate(lines, start=1):
		if unsafe_ref_access.search(line) and "call(" not in enclosing_func(lines, i - 1):
			violations.append(f"internal/ref/ref.go:{i}: unsafe field access must require explicit FieldAccessOption opt-in")

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

if violations:
	for violation in violations:
		print(f"FACADE BOUNDARY VIOLATION: {violation}", file=sys.stderr)
	sys.exit(1)

print("facade boundary governance is valid")
PY
