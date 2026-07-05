#!/usr/bin/env bash
#
# check_api_freeze.sh validates v1 API freeze/deprecation governance metadata.

set -euo pipefail

cd "$(dirname "$0")/.."

AI_CONTEXT_FILE="${AI_CONTEXT_FILE:-ai-context.json}"
TOOLS_JSON_FILE="${TOOLS_JSON_FILE:-docs/api/tools.json}"

python3 - "${AI_CONTEXT_FILE}" "${TOOLS_JSON_FILE}" <<'PY'
from __future__ import annotations

import json
import sys

ai_context_path, tools_path = sys.argv[1], sys.argv[2]
errors: list[str] = []


def add_error(message: str) -> None:
    errors.append(message)


with open(tools_path, "r", encoding="utf-8") as f:
    tools = json.load(f)

with open(ai_context_path, "r", encoding="utf-8") as f:
    ai_context = json.load(f)

api_freeze = ai_context.get("api_freeze", {})
if not isinstance(api_freeze, dict):
    add_error("ai-context.json api_freeze must be an object")
    api_freeze = {}

if not api_freeze.get("decision_card_required", False):
    add_error("api_freeze.decision_card_required must be true")
if not api_freeze.get("replacement_required_for_deprecation", False):
    add_error("api_freeze.replacement_required_for_deprecation must be true")

allowed_statuses = set(api_freeze.get("allowed_statuses", []))
expected_statuses = {"recommended", "compatibility", "experimental", "deprecated"}
if allowed_statuses != expected_statuses:
    add_error("api_freeze.allowed_statuses must match recommended, compatibility, experimental, deprecated")

decision_cards = api_freeze.get("decision_cards", [])
if not isinstance(decision_cards, list):
    add_error("api_freeze.decision_cards must be a list")
    decision_cards = []
if api_freeze.get("decision_card_required", False) and len(decision_cards) < 5:
    add_error("api_freeze.decision_cards must contain at least five v1 decision cards")

public_facades = {item.get("package") for item in ai_context.get("public_facades", []) if isinstance(item, dict)}
expected_card_ids = {
    "v1-public-api-entry-budget",
    "v1-dynamic-contract-matrix",
    "v1-heavy-dependency-isolation",
    "v1-error-taxonomy",
    "v1-security-threat-model",
}
seen_card_ids: set[str] = set()
for index, item in enumerate(decision_cards):
    if not isinstance(item, dict):
        add_error(f"api_freeze.decision_cards[{index}] must be an object")
        continue
    card_id = item.get("id")
    if not isinstance(card_id, str) or not card_id:
        add_error(f"api_freeze.decision_cards[{index}].id must be non-empty")
    elif card_id in seen_card_ids:
        add_error(f"api_freeze.decision_cards duplicate id {card_id}")
    else:
        seen_card_ids.add(card_id)
    if item.get("status") not in expected_statuses:
        add_error(f"api_freeze.decision_cards[{index}].status must be an allowed API status")
    for field in ("decision", "rationale"):
        if not isinstance(item.get(field), str) or not item[field].strip():
            add_error(f"api_freeze.decision_cards[{index}].{field} must be non-empty")
    packages = item.get("packages")
    if not isinstance(packages, list) or not packages:
        add_error(f"api_freeze.decision_cards[{index}].packages must be non-empty")
    else:
        unknown_packages = sorted(package for package in packages if package != "all" and package not in public_facades)
        if unknown_packages:
            add_error(f"api_freeze.decision_cards[{index}].packages contains unknown facade(s): " + ", ".join(unknown_packages))
    validation = item.get("validation")
    if not isinstance(validation, list) or len(validation) < 2:
        add_error(f"api_freeze.decision_cards[{index}].validation must contain at least two validation entries")

missing_card_ids = sorted(expected_card_ids - seen_card_ids)
if missing_card_ids:
    add_error("api_freeze.decision_cards missing required v1 decision card(s): " + ", ".join(missing_card_ids))

cards_by_id = {
    item.get("id"): item
    for item in decision_cards
    if isinstance(item, dict) and isinstance(item.get("id"), str) and item.get("id")
}
api_status_decision_cards = api_freeze.get("api_status_decision_cards")
if not isinstance(api_status_decision_cards, dict):
    add_error("api_freeze.api_status_decision_cards must be an object")
    api_status_decision_cards = {}
mapping_statuses = set(api_status_decision_cards)
if mapping_statuses != expected_statuses:
    add_error("api_freeze.api_status_decision_cards must map recommended, compatibility, experimental, deprecated")
for status in sorted(expected_statuses):
    card_ids = api_status_decision_cards.get(status)
    if not isinstance(card_ids, list) or not card_ids:
        add_error(f"api_freeze.api_status_decision_cards.{status} must contain at least one decision card id")
        continue
    seen_status_card_ids: set[str] = set()
    for index, card_id in enumerate(card_ids):
        if not isinstance(card_id, str) or not card_id:
            add_error(f"api_freeze.api_status_decision_cards.{status}[{index}] must be a non-empty string")
            continue
        if card_id in seen_status_card_ids:
            add_error(f"api_freeze.api_status_decision_cards.{status} duplicates decision card {card_id}")
        seen_status_card_ids.add(card_id)
        card = cards_by_id.get(card_id)
        if card is None:
            add_error(f"api_freeze.api_status_decision_cards.{status} references unknown decision card {card_id}")
            continue
        if card.get("status") != status:
            add_error(f"api_freeze.api_status_decision_cards.{status} references {card_id} with status {card.get('status')!r}")

freeze_checks = api_freeze.get("freeze_checks", [])
if not isinstance(freeze_checks, list) or len(freeze_checks) < 4:
    add_error("api_freeze.freeze_checks must document at least four freeze checks")
else:
    freeze_text = " ".join(str(item).lower() for item in freeze_checks)
    for term in ("decision card", "replacement", "snapshot", "tools catalog"):
        if term not in freeze_text:
            add_error(f"api_freeze.freeze_checks must mention {term!r}")

deprecated_functions: list[str] = []
experimental_functions: list[str] = []
must_functions: list[str] = []
for package in tools.get("packages", []):
    package_name = package.get("name", "")
    for fn in package.get("functions", []):
        status = fn.get("status")
        name = f"{package_name}.{fn.get('name')}"
        if status not in expected_statuses:
            add_error(f"{name} has unknown API status {status!r}")
            continue
        covering_cards = []
        for card_id in api_status_decision_cards.get(status, []):
            card = cards_by_id.get(card_id)
            if not card:
                continue
            packages = card.get("packages", [])
            if "all" in packages or package_name in packages:
                covering_cards.append(card_id)
        if not covering_cards:
            add_error(f"{name} status {status!r} is not covered by api_freeze.api_status_decision_cards")
        if status == "deprecated":
            deprecated_functions.append(name)
            synopsis = fn.get("synopsis", "")
            if "Deprecated:" not in synopsis or "Use " not in synopsis:
                add_error(f"{name} is deprecated but synopsis must include 'Deprecated:' and a replacement using 'Use '")
        if status == "experimental":
            experimental_functions.append(name)
        if isinstance(fn.get("name"), str) and fn["name"].startswith("Must"):
            must_functions.append(name)

if api_freeze.get("v1_candidate", False) and experimental_functions:
    add_error("api_freeze.v1_candidate forbids experimental APIs: " + ", ".join(experimental_functions))

declared_deprecations = api_freeze.get("deprecations", [])
if not isinstance(declared_deprecations, list):
    add_error("api_freeze.deprecations must be a list")
    declared_deprecations = []
declared_deprecated_names = set()
for index, item in enumerate(declared_deprecations):
    if not isinstance(item, dict):
        add_error(f"api_freeze.deprecations[{index}] must be an object")
        continue
    name = item.get("name")
    replacement = item.get("replacement")
    rationale = item.get("rationale")
    if not isinstance(name, str) or not name:
        add_error(f"api_freeze.deprecations[{index}].name must be a non-empty string")
        continue
    declared_deprecated_names.add(name)
    if not isinstance(replacement, str) or not replacement:
        add_error(f"api_freeze.deprecations[{index}].replacement must be a non-empty string")
    if not isinstance(rationale, str) or not rationale:
        add_error(f"api_freeze.deprecations[{index}].rationale must be a non-empty string")

missing_deprecation_entries = sorted(set(deprecated_functions) - declared_deprecated_names)
if missing_deprecation_entries:
    add_error("api_freeze.deprecations missing deprecated function(s): " + ", ".join(missing_deprecation_entries))

must_api_inventory = api_freeze.get("must_api_inventory", [])
if not isinstance(must_api_inventory, list):
    add_error("api_freeze.must_api_inventory must be a list")
    must_api_inventory = []
must_inventory_names: set[str] = set()
for index, item in enumerate(must_api_inventory):
    if not isinstance(item, dict):
        add_error(f"api_freeze.must_api_inventory[{index}] must be an object")
        continue
    name = item.get("name")
    replacement = item.get("replacement")
    rationale = item.get("rationale")
    doc_path = item.get("doc_path")
    if not isinstance(name, str) or not name:
        add_error(f"api_freeze.must_api_inventory[{index}].name must be a non-empty string")
        continue
    must_inventory_names.add(name)
    for field, value in (("replacement", replacement), ("rationale", rationale), ("doc_path", doc_path)):
        if not isinstance(value, str) or not value.strip():
            add_error(f"api_freeze.must_api_inventory[{index}].{field} must be a non-empty string")
    if isinstance(doc_path, str) and doc_path:
        try:
            doc_text = open(doc_path, "r", encoding="utf-8").read()
        except FileNotFoundError:
            add_error(f"api_freeze.must_api_inventory[{index}].doc_path does not exist: {doc_path}")
        else:
            function_name = name.rsplit(".", 1)[-1]
            if function_name not in doc_text:
                add_error(f"{doc_path} must document {name}")
            if isinstance(replacement, str) and replacement:
                replacement_tokens = [token.strip("` ,.") for token in replacement.replace(" or ", " ").replace(",", " ").split()]
                if not any(token and token in doc_text for token in replacement_tokens):
                    add_error(f"{doc_path} must mention replacement guidance for {name}: {replacement}")

missing_must_entries = sorted(set(must_functions) - must_inventory_names)
if missing_must_entries:
    add_error("api_freeze.must_api_inventory missing Must API(s): " + ", ".join(missing_must_entries))
stale_must_entries = sorted(must_inventory_names - set(must_functions))
if stale_must_entries:
    add_error("api_freeze.must_api_inventory includes non-Must or missing API(s): " + ", ".join(stale_must_entries))

if errors:
    for error in errors:
        print(f"api-freeze check error: {error}", file=sys.stderr)
    sys.exit(1)

print(
    "api freeze metadata is valid "
    f"({len(deprecated_functions)} deprecated, {len(experimental_functions)} experimental APIs)"
)
PY
