#!/usr/bin/env python3
"""Refresh the public utility-library comparison from GitHub metadata.

This script is intentionally not part of docs-check or agent-check. It performs
network I/O and writes documentation plus AI metadata, so it should only be run
through the explicit opt-in Makefile target.
"""

from __future__ import annotations

import argparse
import json
import os
import re
import sys
import urllib.error
import urllib.request
from dataclasses import dataclass
from datetime import date
from pathlib import Path


REPOS = [
	"samber/lo",
	"duke-git/lancet",
	"thoas/go-funk",
	"spf13/cast",
	"gookit/goutil",
]

SCOPES = {
	"samber/lo": "Lodash-style generic collection helpers.",
	"duke-git/lancet": "Broad Go utility toolkit with many helper domains.",
	"thoas/go-funk": "Reflection-heavy functional helpers for map/find/filter-style workflows.",
	"spf13/cast": "Type conversion helpers used by configuration-heavy projects.",
	"gookit/goutil": "Daily developer utilities across strings, arrays, maps, env, filesystem, system, CLI, test/assert, and more.",
}


@dataclass(frozen=True)
class RepoMetadata:
	full_name: str
	stars: int
	scope: str
	pushed_at: str
	license_id: str


def github_json(repo: str, token: str | None) -> dict[str, object]:
	request = urllib.request.Request(
		f"https://api.github.com/repos/{repo}",
		headers={
			"Accept": "application/vnd.github+json",
			"User-Agent": "knifer-go-utility-comparison-refresh",
		},
	)
	if token:
		request.add_header("Authorization", f"Bearer {token}")
	try:
		with urllib.request.urlopen(request, timeout=20) as response:
			return json.loads(response.read().decode("utf-8"))
	except urllib.error.HTTPError as exc:
		message = exc.read().decode("utf-8", errors="replace")
		raise RuntimeError(f"GitHub API request for {repo} failed: HTTP {exc.code}: {message}") from exc
	except urllib.error.URLError as exc:
		raise RuntimeError(f"GitHub API request for {repo} failed: {exc.reason}") from exc


def collect_metadata(token: str | None) -> list[RepoMetadata]:
	records: list[RepoMetadata] = []
	for repo in REPOS:
		payload = github_json(repo, token)
		full_name = str(payload.get("full_name") or repo)
		stars = int(payload.get("stargazers_count") or 0)
		pushed_at = str(payload.get("pushed_at") or "")[:10]
		license_payload = payload.get("license")
		license_id = ""
		if isinstance(license_payload, dict):
			license_id = str(license_payload.get("spdx_id") or license_payload.get("key") or "")
		records.append(
			RepoMetadata(
				full_name=full_name,
				stars=stars,
				scope=SCOPES[repo],
				pushed_at=pushed_at,
				license_id=license_id,
			)
		)
	return sorted(records, key=lambda item: (-item.stars, item.full_name))


def format_table(records: list[RepoMetadata]) -> str:
	lines = [
		"| Rank | Library | Stars | Scope | Last pushed | License |",
		"| --- | --- | ---: | --- | --- | --- |",
	]
	for index, record in enumerate(records, start=1):
		lines.append(
			"| {rank} | `{library}` | {stars} | {scope} | {pushed_at} | {license_id} |".format(
				rank=index,
				library=record.full_name,
				stars=f"{record.stars:,}",
				scope=record.scope,
				pushed_at=record.pushed_at,
				license_id=record.license_id,
			)
		)
	return "\n".join(lines)


def top5_section(records: list[RepoMetadata]) -> str:
	return "\n".join(
		[
			"## GitHub Top 5 Utility Libraries",
			"",
			'This list uses the project comparison scope "Go utility libraries", not web',
			"frameworks, ORMs, CLI frameworks, or test frameworks.",
			"",
			format_table(records),
		]
	)


def source_section(records: list[RepoMetadata]) -> str:
	lines = ["## Sources", ""]
	lines.extend(f"- GitHub API: `https://api.github.com/repos/{record.full_name}`" for record in records)
	return "\n".join(lines)


def replace_section(text: str, heading: str, replacement: str) -> str:
	pattern = re.compile(
		rf"^## {re.escape(heading)}\n(?P<body>.*?)(?=^## |\Z)",
		flags=re.MULTILINE | re.DOTALL,
	)
	if not pattern.search(text):
		raise RuntimeError(f"missing markdown section: ## {heading}")
	return pattern.sub(replacement.rstrip() + "\n\n", text, count=1)


def update_markdown(path: Path, records: list[RepoMetadata], checked_date: str) -> str:
	text = path.read_text(encoding="utf-8")
	text, count = re.subn(r"Last checked: \d{4}-\d{2}-\d{2}\.", f"Last checked: {checked_date}.", text, count=1)
	if count != 1:
		raise RuntimeError(f"{path} must contain exactly one Last checked line")
	text = replace_section(text, "GitHub Top 5 Utility Libraries", top5_section(records))
	text = replace_section(text, "Sources", source_section(records))
	return text.rstrip() + "\n"


def update_ai_context(path: Path, records: list[RepoMetadata], checked_date: str) -> str:
	text = path.read_text(encoding="utf-8")
	start_marker = '  "utility_top5_comparison_governance_v2": {'
	end_marker = '  "safe_crypto_advanced_closeout_governance": {'
	start = text.find(start_marker)
	end = text.find(end_marker, start)
	if start < 0 or end < 0:
		raise RuntimeError("ai-context.json is missing utility_top5_comparison_governance_v2 block")
	block = text[start:end]
	block, last_checked_count = re.subn(
		r'"last_checked": "[^"]+"',
		f'"last_checked": "{checked_date}"',
		block,
		count=1,
	)
	top5_json = json.dumps([record.full_name for record in records], ensure_ascii=True)
	block, top5_count = re.subn(
		r'"top5": \[[^\]]+\]',
		f'"top5": {top5_json}',
		block,
		count=1,
	)
	if last_checked_count != 1 or top5_count != 1:
		raise RuntimeError("ai-context.json top5 governance block has unexpected shape")
	return text[:start] + block + text[end:]


def parse_args() -> argparse.Namespace:
	parser = argparse.ArgumentParser(description=__doc__)
	parser.add_argument(
		"--write",
		action="store_true",
		help="write docs/doc/utility-library-comparison.md and ai-context.json",
	)
	parser.add_argument(
		"--date",
		default=date.today().isoformat(),
		help="last-checked date to write; defaults to today's local date",
	)
	return parser.parse_args()


def main() -> int:
	args = parse_args()
	root = Path(__file__).resolve().parents[1]
	records = collect_metadata(os.environ.get("GITHUB_TOKEN"))
	markdown_path = root / "docs/doc/utility-library-comparison.md"
	ai_context_path = root / "ai-context.json"
	markdown = update_markdown(markdown_path, records, args.date)
	ai_context = update_ai_context(ai_context_path, records, args.date)
	if args.write:
		markdown_path.write_text(markdown, encoding="utf-8")
		ai_context_path.write_text(ai_context, encoding="utf-8")
	else:
		print(format_table(records))
		print()
		print("Dry run only. Pass --write to update documentation and ai-context.json.", file=sys.stderr)
	return 0


if __name__ == "__main__":
	raise SystemExit(main())
