# Coverage Evidence

`make coverage-profile` runs the complete root-module race/shuffle suite with
`-coverpkg=./...`, so shared implementations receive coverage from their callers.
The producer merges identical file/range blocks once and combines execution with
logical OR. It writes a normalized statement profile and `<profile>.meta.json`.
Execution frequencies are not preserved by this statement-coverage normalization.

The sidecar records the actual Go version, command, test scope, test exit status,
commit, Git tree, effective source hash, module hashes (including nested modules),
profile hash, timestamp and CI run/attempt. Source snapshots before and after the
test run must agree. Failed or partial tests cannot create passing coverage
evidence, even if the available profile exceeds the numeric threshold.

```sh
make coverage-profile COVERAGE_FILE=/tmp/knifer-coverage.out
make coverage-check COVERAGE_FILE=/tmp/knifer-coverage.out
go run ./bin/coveragecheck -root . -json /tmp/knifer-coverage.out
make coverage-report COVERAGE_FILE=/tmp/knifer-coverage.out
```

`coverage-check` rejects missing/stale execution evidence, changed source or
dependency inputs, modified profile bytes, incomplete scope and failed tests.
It also rejects malformed profiles and conflicting statement counts for the
same block. CI uploads the profile and metadata together. Recreate coverage
after committing or changing source; a prior uncommitted run is not evidence for
a new commit, even if changes are small.

The root policy explicitly enrolls `internal/httpboundary` and
`internal/httpx/internal/shared` in security coverage. Both missing data and
coverage below the security threshold fail the gate. Tests include resolver
errors, empty/mixed address lists and nil-context propagation.

Reports separate weighted statement coverage for the runtime library, internal
implementations, facades and `governance_go_test`. The last group only measures
code instrumented in the Go test processes. Governance commands built/spawned by
fixtures without instrumentation are not automatically covered; low values in
that group do not mean those commands have no tests.

The roadmap links this execution evidence instead of copying percentages into
static scorecards. Use the recorded commit/source/profile hashes whenever
publishing coverage numbers. Database integration results remain a separate
required CI result from `integration/db`; root `./...` coverage does not traverse
that standalone module.
