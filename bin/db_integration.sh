#!/usr/bin/env bash
set -euo pipefail
root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
mkdir -p "$root/.aiflow/db-integration"
cd "$root"
export GOWORK=off
export GOTOOLCHAIN=local
export KNIFER_DB_ENGINES="${DB_TEST_ENGINES:-sqlite,postgres,mysql}"
"${GO:-go}" -C integration/db test -mod=readonly -race -shuffle=on -count=1 -timeout=3m -json ./... \
  | tee "$root/.aiflow/db-integration/results.jsonl"
