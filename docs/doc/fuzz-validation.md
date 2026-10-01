# Fuzz Validation and Failure Artifacts

`make fuzz-smoke` discovers every compiled Fuzz target and runs each for one
second. `make fuzz-extended` uses 30 seconds per target and runs daily and on
manual dispatch through `.github/workflows/fuzz.yml`. Extended runs are separate
from the pull-request admission job.

Both commands use four fuzz workers, two seconds of minimization per failure,
and a ten-minute whole-run deadline by default. Each target also has a test
timeout. Override budgets explicitly when needed:

```bash
make fuzz-smoke FUZZTIME=2s FUZZ_PARALLEL=2
make fuzz-extended FUZZ_EXTENDED_TIME=60s FUZZ_RUN_TIMEOUT=15m
make fuzz-smoke FUZZ_PKGS=./internal/json
```

## Isolation

The runner compiles instrumented test binaries once per package, copies package
`testdata` into an isolated working directory and executes there. Fuzz targets
should keep relative fixture files under `testdata`; other source-directory
assumptions are not part of the runner contract. The native Go failing corpus
path is relative to that working directory, so a failure never writes directly
into the package's curated source corpus.

Every invocation creates a unique `.aiflow/fuzz/run-*` directory. Temporary
binaries and exploration caches are removed after execution. The persistent
artifacts include:

- `summary.json`: target states, exact Go version, commit/tree, dirty-worktree
  indicator, budgets, binary checksums and captured corpus paths/checksums.
- `discovery.log`: discovery, package lookup and build diagnostics.
- `targets/<index>-<name>/fuzz.log`: native output, including failure and replay
  details.
- `targets/<index>-<name>/testdata/fuzz/<name>/`: original copied corpus and any
  newly generated failing inputs.

Workspace artifact overrides must stay under `.aiflow/`; an explicitly selected
artifact directory outside the workspace is also supported. A failed target
does not prevent later targets from running unless the whole-run deadline or
cancellation stops the run. The command fails if any target fails, times out,
cannot compile, or remains unexecuted. Reports are updated during execution so
an interrupted run is visibly incomplete.

## Reproducing and retaining a failure

Download the run's artifact and inspect the failed target's log and corpus.
Use the recorded commit and Go version. For a dirty local run, also recover the
local changes; the recorded commit alone is not an exact source snapshot.

Review the selected corpus file, then deliberately copy it into the corresponding
package's `testdata/fuzz/<FuzzName>/` directory. Run the native command printed in
the log, such as `go test ./internal/json -run '^FuzzName/<input-hash>$'`.
Commit a useful regression input together with its fix after review. The runner
never promotes or deletes curated source corpus automatically.

PR smoke, scheduled/manual exploration and release jobs upload `.aiflow/fuzz/`
artifacts with `if: always()`, including when fuzzing fails. Artifacts are retained
for 14 days. A cancelled or hard-killed runner may have only a partial summary;
such a run cannot satisfy admission.

## Side-effect contract

Fuzz commands and release-check declare `writes_workspace=true`, their exact
artifact paths, and `workspace_write_scope=runtime_artifacts`. The metadata
checker only permits this automatic-write scope for ignored `.aiflow/` paths;
source paths, traversal and escaping symlinks are rejected. Existing temporary
coverage profiles outside the workspace remain separately declared artifacts.
T2 admission output follows the same runtime-artifact contract.
