#!/usr/bin/env python3
"""Generate the facade-tiering dependency view from ai-context.json."""

from __future__ import annotations

import json
import re
from pathlib import Path


DEPENDENCY_BEGIN = "<!-- BEGIN GENERATED DEPENDENCY TIERS -->"
DEPENDENCY_END = "<!-- END GENERATED DEPENDENCY TIERS -->"
SECURITY_BEGIN = "<!-- BEGIN GENERATED SECURITY OVERLAY -->"
SECURITY_END = "<!-- END GENERATED SECURITY OVERLAY -->"

SECURITY_CATEGORIES = [
	(
		"Network and URL boundaries",
		["vhttp", "vresty", "vurl", "vnet"],
	),
	(
		"File, archive, and config boundaries",
		["vfile", "vzip", "vconf"],
	),
	(
		"Crypto, token, random, and identity boundaries",
		["vcrypto", "vjwt", "vrand", "vid"],
	),
	(
		"SQL and command boundaries",
		["vdb", "vcli"],
	),
	(
		"Provider contract boundaries",
		["vai", "vftp", "vssh"],
	),
]


def as_code_list(values: list[str]) -> str:
	return ", ".join(f"`{value}`" for value in values)


def dependency_tiers_table(tiers: dict[str, object]) -> str:
	core = tiers.get("core_facades", [])
	heavy = tiers.get("heavy_extension_facades", [])
	providers = tiers.get("provider_contract_facades", [])
	if not all(isinstance(values, list) for values in (core, heavy, providers)):
		raise RuntimeError("dependency_tiers facade lists must be arrays")
	rows = [
		"| Tier | Facades | Import rule |",
		"| --- | --- | --- |",
		"| core facades | {facades} | Standard-library-first; third-party imports require explicit allowlist review. |".format(
			facades=as_code_list([str(value) for value in core]),
		),
		"| heavy extension facades | {facades} | Optional integrations stay inside their owning facade and matching `internal/*` package family. |".format(
			facades=as_code_list([str(value) for value in heavy]),
		),
		"| provider contract facades | {facades} | Public APIs expose provider interfaces and call contracts; concrete clients, credentials, dictionaries, and NLP engines stay outside core. |".format(
			facades=as_code_list([str(value) for value in providers]),
		),
	]
	return "\n".join(rows)


def security_overlay_table(security_sensitive: list[str]) -> str:
	security_sensitive_set = set(security_sensitive)
	rows = [
		"| Category | Facades |",
		"| --- | --- |",
	]
	for category, facades in SECURITY_CATEGORIES:
		filtered = [facade for facade in facades if facade in security_sensitive_set]
		if filtered:
			rows.append(f"| {category} | {as_code_list(filtered)} |")
	covered = {facade for _, facades in SECURITY_CATEGORIES for facade in facades}
	uncovered = sorted(security_sensitive_set - covered)
	if uncovered:
		raise RuntimeError("security overlay does not cover: " + ", ".join(uncovered))
	return "\n".join(rows)


def replace_generated_block(text: str, begin: str, end: str, generated: str) -> str:
	pattern = re.compile(rf"{re.escape(begin)}\n.*?\n{re.escape(end)}", flags=re.DOTALL)
	replacement = f"{begin}\n{generated.rstrip()}\n{end}"
	text, count = pattern.subn(replacement, text, count=1)
	if count != 1:
		raise RuntimeError(f"expected exactly one generated block between {begin} and {end}")
	return text


def main() -> int:
	root = Path(__file__).resolve().parents[1]
	ai_context = json.loads((root / "ai-context.json").read_text(encoding="utf-8"))
	doc_path = root / "docs/doc/facade-tiering.md"
	doc = doc_path.read_text(encoding="utf-8")
	doc = replace_generated_block(
		doc,
		DEPENDENCY_BEGIN,
		DEPENDENCY_END,
		dependency_tiers_table(ai_context["dependency_tiers"]),
	)
	doc = replace_generated_block(
		doc,
		SECURITY_BEGIN,
		SECURITY_END,
		security_overlay_table([str(value) for value in ai_context["security_sensitive_packages"]]),
	)
	doc_path.write_text(doc.rstrip() + "\n", encoding="utf-8")
	return 0


if __name__ == "__main__":
	raise SystemExit(main())
