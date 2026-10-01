#!/usr/bin/env bash
set -euo pipefail

# The runner executes compiled fuzz tests in isolated artifact directories so
# native failing corpus never lands in package source directories.
exec "${GO:-go}" run ./bin/fuzzrun \
  -go "${GO:-go}" \
  -artifacts "${FUZZ_ARTIFACT_DIR:-.aiflow/fuzz}" \
  -time "${FUZZTIME:-1s}" \
  -minimize "${FUZZ_MINIMIZE_TIME:-2s}" \
  -timeout "${FUZZ_RUN_TIMEOUT:-10m}" \
  -parallel "${FUZZ_PARALLEL:-4}" "$@"
