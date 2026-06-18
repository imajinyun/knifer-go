# go-knifer — AI Agent Guide

> Go utility library (48 public `v*` facade packages + `internal/*` implementations).

---

## Quick reference

### Project layout

```
./
├── doc.go              # root package: domain-grouped navigation, no business APIs
├── errors.go           # root error contract: ErrCode, Error, CodeCarrier, CodeOf
├── v<domain>/          # public facade — import these (thin: forwards to internal/*)
│   └── doc.go          # required: package doc with full domain name on first line
├── internal/<domain>/  # implementation — not importable by external modules
├── bin/                # validation scripts (api_snapshot, check_arch, check_coverage)
├── docs/
│   ├── api/exports.txt # exported API snapshot (19k+ lines, CI-enforced)
│   └── doc/            # 48 per-package quickstart docs (01-vbean.md .. 48-vzip.md)
└── .github/workflows/  # CI (go.yml) + release (release.yml) automation
```

### Core rules (enforced by CI)

- **Public facades never import each other.** Shared logic goes into `internal/*`.
- **Internal packages never import public facades.** Direction: `v* → internal/*`.
- **Production `panic` is exceptional.** Only in `MustXxx`/`PanicXxx` APIs or documented cases.
- **Facades are thin.** No business logic, loops, `panic`, or type assertions in `v*`.

### Package catalog (48 v* packages)

| Package | Domain | internal |
|---------|--------|----------|
| vstr | string/text | internal/str |
| vslice | slices | internal/slice |
| vmap | maps | internal/maps |
| vcache | caches (FIFO/LRU/LFU/TTL) | internal/cache |
| vcrypto | cryptography/digests | internal/crypto |
| vhttp | HTTP (stdlib) | internal/httpx |
| vresty | HTTP (Resty) | internal/httpx/resty |
| vjson | JSON | internal/json |
| vconf | configuration | internal/conf |
| vid | UUID/Snowflake/NanoId | internal/id |
| vjwt | JWT sign/verify | internal/jwt |
| vrand | randomness | internal/rand |
| vlog | logging | internal/log |
| verr | error handling/panic recovery | internal/errx |
| vfile | file/IO | internal/file |
| vurl | URL/URI | internal/url |
| vmask | data masking | internal/mask |
| vcodec | Base64/Hex | internal/codec |
| vhash | non-crypto hashes | internal/hash |
| vzip | archive/compression | internal/zip |
| vset | sets | internal/sets |
| vobj | object helpers | internal/obj |
| vconv | type conversion | internal/conv |
| vdate | date/time | internal/date |
| vref | reflection | internal/ref |
| vregex | regular expressions | internal/regex |
| vtpl | templates | internal/template |
| vbool | booleans | internal/boolean |
| vnum | numeric helpers | internal/num |
| vbean | struct/map mapping | internal/bean |
| vblf | bloom filter | internal/bloomfilter |
| vcsv | CSV | internal/csvx |
| vimg | images/captchas | internal/imgx |
| vxml | XML | internal/xml |
| vmail | email/SMTP | internal/mail |
| vskt | sockets | internal/socket |
| vnet | IP/port/network | internal/net |
| vdb | database/SQL | internal/db |
| vcron | cron scheduling | internal/cron |
| vjob | job orchestration | internal/job |
| vsem | semaphores | internal/semaphore |
| vdfa | DFA word-tree matching | internal/dfa |
| vpass | password strength | internal/pass |
| vident | identity numbers | internal/identity |
| vform | form/input validation | internal/validator |
| vver | version comparison | internal/version |
| vsys | system information | internal/system |
| vpoi | office documents (Excel) | internal/poi |
| vyaml | YAML | internal/yaml |

### Validation commands

| Command | Scope |
|---------|-------|
| `make test` | Run unit tests |
| `make test-race` / `make coverage-profile` | Race/shuffle tests with coverage |
| `make coverage-report COVERAGE_FILE=<file>` | Print function coverage |
| `make coverage-check COVERAGE_FILE=<file>` | Enforce coverage gates |
| `make quick-check` | Fast local: mod-verify → vet → arch → test → api-check → diff-whitespace |
| `make security-check` | Lint + govulncheck |
| `make full-check COVERAGE_FILE=/tmp/coverage.out` | Full pre-push: quick-check + race coverage + coverage gate + lint + vuln |
| `make ci-test` | CI test-job gate (mod-verify + vet + tidy-check + diff-check + arch + test-race + coverage-check + api-check) |
| `make check` | Alias for `full-check` |
| `UPDATE_API=1 make api-check` | Refresh API snapshot after intentional public API changes |
| `go generate ./...` | Refresh API snapshot via `//go:generate` in `doc.go` |
| `make govulncheck` | Vulnerability scan |
| `make bench-core` | Core benchmark baselines (`internal/slice`, `internal/maps`, `internal/str`, `internal/num`) |
| `make bench-facade` | Facade benchmark baselines (`vslice`, `vmap`, `vstr`, `vnum`) |
| `make bench-codec` | JSON/XML benchmark baselines |
| `make bench-smoke` | Short core benchmark smoke check |

### API design rules

- Prefer `(result, error)` over `panic` for recoverable failures.
- IO/network functions take `context.Context` as first parameter.
- No synonym aliases for the same function.
- Renaming a public symbol is a breaking change (SemVer).
- Security-sensitive code: `vhttp`, `vresty`, `vurl`, `vconf`, `vzip`, `vfile`, `vcrypto`, `vjwt`, `vrand`, `vid`, `vdb`. See `SECURITY.md`.

### Governance constraints

- **Coverage**: Keep total coverage above the threshold in `bin/check_coverage.sh`.
- **Architecture**: 8 rules enforced by `bin/check_arch.sh` — doc.go existence, no v*-to-v* imports, per-file internal/ imports, no internal→v* imports, package comments, panic policy, facade boundary policy, dependency allowlist.
- **API snapshot**: `docs/api/exports.txt` is CI-enforced. Run `UPDATE_API=1 make api-check` after intentional public API changes.
- **Panic**: Production code must not introduce new `panic()` calls unless in a `MustXxx`/`PanicXxx` function.

---

## Workflows

### General change validation and delivery

When the user asks to implement, rename, refactor, document, or otherwise modify repository files, handle the change end-to-end without stopping after edits:

1. **Inspect** the existing code and documentation first so the change follows current package boundaries, naming, and README/API snapshot conventions.

2. **Apply** only the requested logical change. Do not include unrelated local files, generated experiments, secrets, or user-owned untracked files.

3. **Format** touched Go files with `gofmt -w` before validation.

4. **Validate** focused tests first, then broaden to repository-level gates when feasible:
   - `go test -v -gcflags="all=-l -N" ./<changed-package>` for affected Go packages.
   - `go test -v -gcflags="all=-l -N" ./...` for repository-wide regressions.
   - `go vet ./...` after code or public API changes.
   - `bash bin/check_arch.sh` after production code changes.
   - `bash bin/check_api_compat.sh`; if the public API change is intentional, run `UPDATE_API=1 bash bin/check_api_compat.sh` and re-run the check. Public facade additions must update `docs/api/exports.txt` in the same logical change.
   - `golangci-lint run ./...` after non-trivial Go code or test changes when the tool is available.
   - For coverage gates, first generate a fresh profile, then pass that exact file to the checker, e.g. `go test -race -shuffle=on -coverprofile=/tmp/go-knifer-coverage.out ./...` followed by `bash bin/check_coverage.sh /tmp/go-knifer-coverage.out`. Do not rely on an implicit or stale `coverage.out`.
   - `git diff --check` before committing.
   - Prefer the named workflow targets when they match the change scope: `make quick-check` for fast local validation, `make security-check` for lint/vulnerability gates, `make full-check` or `make check` for the full pre-push gate, and `make ci-test` for the CI test-job gate.

5. If validation fails, **fix the cause** and re-run the failing command before reporting completion.

6. Before committing, **re-check the final staged logical change**:
   - Run `git status --porcelain=v1 -b` and review `git diff --stat` / `git diff --staged --stat` so the commit contains only the requested files.
   - Ensure the latest validation was run after the final edit/API snapshot update, not before it.
   - For non-trivial Go changes, the pre-commit validation set should include: focused package tests, `go test -v -gcflags="all=-l -N" ./...`, `go vet ./...`, `bash bin/check_arch.sh`, `bash bin/check_api_compat.sh`, `golangci-lint run ./...`, `go test -race -shuffle=on -coverprofile=/tmp/go-knifer-coverage.out ./...`, `bash bin/check_coverage.sh /tmp/go-knifer-coverage.out`, and `git diff --check`. `make full-check COVERAGE_FILE=/tmp/go-knifer-coverage.out` is the preferred aggregate target when a full local gate is feasible.
   - If a public API snapshot was intentionally refreshed, run the API check once to observe the stale snapshot, then `UPDATE_API=1 bash bin/check_api_compat.sh`, then re-run `bash bin/check_api_compat.sh` and include `docs/api/exports.txt` in the same logical commit.

7. **Commit**: Generate a conventional commit message from the actual diff, preferring concise messages such as `feat: ...`, `fix: ...`, `docs: ...`, `refactor: ...`, or `test: ...`. Stage only files belonging to the requested logical change, commit them.

8. **Push** the branch to the configured remote when the user asks to commit/push or when the workflow explicitly requires it.

9. After pushing, run `git status --porcelain=v1 -b` to confirm the branch is clean/in sync, then **report** the commit hash, pushed branch, validation commands, and any intentionally excluded local files.

### Collection modernization

When changing `internal/slice`, `vslice`, `internal/maps`, or `vmap`, keep the package focused on typed convenience helpers that complement the Go standard library rather than reimplementing it:

1. Prefer Go 1.21+ standard packages for overlapping behavior before adding or keeping custom loops:
   - Use `slices.Contains`, `slices.Index`, `slices.IndexFunc`, `slices.Reverse`, `slices.Clone`, `slices.Concat`, `slices.DeleteFunc`, `slices.Sorted`, and `slices.SortedFunc` where they preserve the existing API contract.
   - Use `maps.Clone`, `maps.Copy`, `maps.Equal`, `maps.EqualFunc`, `maps.Keys`, `maps.Values`, and `maps.All` where they preserve the existing API contract.
   - Use `cmp.Or` for simple zero-value fallback helpers when that exactly matches the function semantics.

2. Keep custom implementations for behavior the standard library does not cover directly, especially order-preserving `Uniq`/`UniqBy`, `GroupBy`, `CountBy`, `KeyBy`, `Chunk`, `PartitionBy`, set-style helpers, typed pair conversions, and facade compatibility functions.

3. For Go 1.23+ iterator support, expose adapters instead of inventing callback-only APIs:
   - Slice packages should provide value and indexed adapters backed by `slices.Values` and `slices.All`.
   - Map packages should provide key-value, key-only, and value-only adapters backed by `maps.All`, `maps.Keys`, and `maps.Values`.

4. Preserve existing nil/empty return contracts. If a public helper historically returns a non-nil empty slice or map, wrap standard-library results as needed so callers do not observe a regression.

5. Any new public facade function in `vslice` or `vmap` must update `docs/api/exports.txt` via the API compatibility workflow and include focused tests in both the internal package and facade package.

6. Validate collection changes with focused tests first (`./internal/slice ./vslice` or `./internal/maps ./vmap`), then run the normal repository gates from the general workflow before reporting completion.

### Governance audit

When the user asks to continue general governance, generate next governance tasks, or uses equivalent Chinese wording such as "继续治理", "生成下一步治理任务", or "更新工作流", treat it as an autonomous quality pass over security, public facade usability, coverage stability, and benchmark baselines:

1. Start with safety checks and repository context:
   - Run `git status --porcelain=v1 -b` and avoid mixing unrelated local files into the current logical change.
   - Run `govulncheck ./...` before changing dependencies or security-sensitive code. If vulnerabilities are reported, classify reachable findings separately from dependency-only findings and do not upgrade dependencies blindly.

2. Review security suppressions and random/entropy boundaries:
   - Search `#nosec` and confirm each suppression has a narrow reason tied to the operation, especially `G304`, `G115`, `G404`, `G110`, `G103`, and `G204`.
   - Keep `math/rand` confined to non-security helpers, deterministic tests, or documented compatibility fallbacks. Security-sensitive bytes must use fail-closed crypto-random helpers.

3. Audit public facade usability after internal improvements:
   - Compare recent internal or facade changes with `docs/api/exports.txt` and quickstart/example coverage.
   - Add executable `ExampleXxx` tests for reader-facing behavior, especially iterator adapters in `vslice`/`vmap` and security-sensitive examples in `vrand`, `vid`, and `verr`.
   - Keep examples deterministic; sort map-derived output before printing.

4. Use coverage data to choose small, stable test improvements:
   - Generate a fresh profile with `go test -coverprofile=/tmp/go-knifer-coverage-audit.out ./...` and inspect `go tool cover -func=/tmp/go-knifer-coverage-audit.out`.
   - Prefer low-risk tests for public facade options, error branches, nil/empty boundaries, deterministic fallback behavior, and serialization round trips.
   - Do not test implementation details or add flaky timing/network dependencies only to increase coverage.

5. Establish benchmark baselines before performance work:
   - Add benchmarks only for stable hot-path helpers with deterministic inputs, such as `internal/slice`, `internal/maps`, `internal/str`, and `internal/num`.
   - Cover empty, small, medium, and large input sizes where meaningful.
   - Use `b.Loop()` for new benchmarks because the module targets Go 1.25. Keep benchmark results out of assertions; they are baselines, not optimization claims.
   - Run `go test -bench=. -run=^$ ./<target packages>` and report that the benchmark suite runs, not that a performance change was proven.

6. Validate the final governance change with the normal repository gates:
   - `gofmt -w` on touched Go files.
   - Focused tests for changed packages.
   - `go test -v -gcflags="all=-l -N" ./...`.
   - `go vet ./...`.
   - `bash bin/check_arch.sh`.
   - `bash bin/check_api_compat.sh` when public exports may have changed.
   - `golangci-lint run ./...` when available.
   - `go test -race -shuffle=on -coverprofile=/tmp/go-knifer-coverage.out ./...` followed by `bash bin/check_coverage.sh /tmp/go-knifer-coverage.out`.
   - `git diff --check`.

### Package test governance

When the user asks to continue package test governance, package-level test cleanup, or uses equivalent Chinese wording such as "继续推进包测试治理", treat it as an autonomous rolling workflow:

1. Continue round by round without asking for confirmation.
2. In each round, split one aggregated tracked Go `*_test.go` file into smaller test files organized by source-file responsibility.
3. Preserve existing test behavior unless a real defect is proven by code evidence.
4. After each completed round, write a conventional commit, commit only that round's changes, push to the remote, then automatically start the next round.

#### Required protocol

For every round:

1. **Step1**: run the skill preparation script and record `BITS_TMP_ROOT` as `TMP_ROOT`.
2. **Step2**: load language knowledge and project context.
3. **Step3**: determine `TARGETS` before editing tests.
4. **Step4**: produce `BUG_MAP` before generating or moving test cases.
5. **Step5**: report each generate → verify → fix loop with command, failure type, and result.
6. **Step6**: report scope, defect analysis, generated test cases, validation result, commit, and push status.

When the round only reorganizes existing tests, `BUG_MAP` may be empty, but record a short candidate-filtering summary.

#### Target selection rules

Prefer the next largest tracked `*_test.go` file that aggregates multiple source responsibilities.

Exclude:
- Untracked or intentionally excluded directories: `internal/csvx`, `vcsv`, `internal/image`, `vimage`, `internal/imgx`, `vimg`, `internal/pass`, `vpass`, `docs`.
- Files with unrelated local user changes, especially `internal/str/str_test.go` and `vstr/str_test.go` unless the user explicitly asks to handle them.
- Generated code, vendor/build outputs, and trivial single-responsibility tests.

Prefer split names that reflect source responsibility, for example:
- `*_helpers_test.go` for shared fixtures or local test helpers.
- `*_options_test.go` for option/provider behavior.
- `*_lifecycle_test.go` for start/stop/reset behavior.
- `*_safe_*_test.go` for safety/security boundary cases.

#### Package test governance validation rules

Run validation in this order:

1. `gofmt -w` on all touched Go files.
2. Package test for the changed package, e.g. `go test -v -gcflags="all=-l -N" ./internal/foo`.
3. Stage only the current round's files.
4. Create a clean detached worktree from `HEAD`, apply only the staged diff, then run:
   - `go test -v -gcflags="all=-l -N" ./...`
   - `go vet ./...`
   - `bash bin/check_arch.sh`
   - `go test -race -shuffle=on -coverprofile=/tmp/go-knifer-coverage.out ./...`
   - `bash bin/check_coverage.sh /tmp/go-knifer-coverage.out`
   - `bash bin/check_api_compat.sh` when public exports may have changed.
   - `golangci-lint run ./...` when production code or non-trivial tests changed and the tool is available.
5. Run `utree flush` before the final report.

The clean-worktree validation is mandatory because this repository may contain unrelated local modifications or untracked experimental packages.

#### Package test governance commit and push rules

Commit only the current round's staged split files. Do not include unrelated local changes.

Use commit messages in this style:
```text
test: split <package-or-area> tests by responsibility
```

After a successful commit, run `git push`. If validation fails, fix the round before committing.

#### Repository-specific gates

- Keep total coverage above the repository gate reported by `bin/check_coverage.sh`.
- `bin/check_arch.sh` must pass; production code must not introduce forbidden panic patterns.
- Do not modify production code during package test governance unless a verified test-governance blocker requires it and the final report calls it out.