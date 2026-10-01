"""Validate the narrow ignored-runtime-artifact workspace write scope."""
import json
import os
from pathlib import PurePosixPath
import subprocess
import sys


def validate_runtime_artifacts(root, commands):
    errors = []
    for name, spec in commands.items():
        if not isinstance(spec, dict):
            errors.append(f"commands.{name} must be an object")
            continue
        scope = spec.get("workspace_write_scope")
        if scope is None:
            continue
        if scope != "runtime_artifacts":
            errors.append(f"commands.{name} has unknown workspace_write_scope")
            continue
        if spec.get("writes_workspace") is not True or spec.get("writes_git_config"):
            errors.append(f"commands.{name} runtime scope requires workspace writes and forbids Git config writes")
        if not spec.get("writes_files") or not spec.get("creates_artifacts"):
            errors.append(f"commands.{name} runtime scope must declare writes_files and creates_artifacts")
        for path in spec.get("writes_files", []) + spec.get("creates_artifacts", []):
            if not isinstance(path, str):
                errors.append(f"commands.{name} artifact path must be a string")
                continue
            if path in spec.get("creates_artifacts", []) and path not in spec.get("writes_files", []):
                # Existing profiles outside the workspace remain declared temp
                # artifacts; only workspace writes use the runtime scope.
                if os.path.realpath(path).startswith(("/private/tmp/", "/tmp/")):
                    continue
            parts = PurePosixPath(path).parts
            if not path.startswith(".aiflow/") or ".." in parts or "\\" in path:
                errors.append(f"commands.{name} artifact path escapes .aiflow: {path}")
                continue
            expected = os.path.join(os.path.realpath(root), ".aiflow")
            actual = os.path.realpath(os.path.join(root, path))
            if not actual.startswith(expected + os.sep):
                errors.append(f"commands.{name} artifact path resolves outside .aiflow: {path}")
                continue
            ignored = subprocess.run(["git", "-C", root, "check-ignore", "--no-index", "-q", "--", path], check=False)
            if ignored.returncode != 0:
                errors.append(f"commands.{name} artifact path is not ignored: {path}")
    fuzz = commands.get("fuzz_smoke", {})
    release = commands.get("release_check")
    if release is not None and fuzz.get("workspace_write_scope") == "runtime_artifacts":
        if release.get("workspace_write_scope") != "runtime_artifacts" or release.get("writes_workspace") is not True:
            errors.append("release_check must declare inherited fuzz artifact writes")
        if not set(fuzz.get("writes_files", [])).issubset(release.get("writes_files", [])):
            errors.append("release_check must include fuzz_smoke writes_files")
    return errors


if __name__ == "__main__":
    with open(os.path.join(sys.argv[1], "ai-context.json"), encoding="utf-8") as file:
        findings = validate_runtime_artifacts(sys.argv[1], json.load(file)["commands"])
    for finding in findings:
        print(finding, file=sys.stderr)
    sys.exit(bool(findings))
