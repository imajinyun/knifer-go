#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")/.."

echo "panic policy: scanning production Go files"
python3 - <<'PY'
from __future__ import annotations

import pathlib
import re
import sys

root = pathlib.Path.cwd()
violations: list[str] = []

allowed_panic_paths = {
	"internal/bloomfilter/bitset_bloomfilter.go",
	"internal/bloomfilter/filter.go",
	"internal/cron/pattern.go",
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


def enclosing_func(lines: list[str], idx: int) -> str:
	for j in range(idx, -1, -1):
		line = lines[j].strip()
		if line.startswith("func "):
			return line
	return ""


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

if violations:
	for violation in violations:
		print(f"PANIC POLICY VIOLATION: {violation}", file=sys.stderr)
	sys.exit(1)

print("panic policy is valid")
PY
