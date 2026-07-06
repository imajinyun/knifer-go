# knifer-go — AI Agent Guide

> Go utility library (54 public `v*` facade packages + `internal/*` implementations).

`AGENTS.md` is the cross-agent concise entrypoint. This file is the Claude-specific detailed guide and should carry the deeper workflow, package-boundary, and validation playbooks. Keep shared rules aligned, but avoid duplicating long policy text across both files.

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
├── bin/                # validation scripts (api_snapshot, toolsgen, check_arch, check_coverage)
├── docs/
│   ├── api/exports.txt # exported API snapshot (19k+ lines, CI-enforced)
│   ├── api/tools.json  # machine-readable facade function catalog for AI/tooling
│   ├── api/tools.md    # generated human-readable facade function catalog
│   └── doc/            # 54 per-package quickstart docs (01-vai.md .. 54-vzip.md)
├── AGENTS.md           # cross-agent workflow, validation, and generated-doc rules
└── .github/workflows/  # CI (go.yml) + release (release.yml) automation
```

### Core rules (enforced by CI)

- **Public facades never import each other.** Shared logic goes into `internal/*`.
- **Internal packages never import public facades.** Direction: `v* → internal/*`.
- **Production `panic` is exceptional.** Only in `MustXxx`/`PanicXxx` APIs or documented cases.
- **Facades are thin.** No business logic, loops, `panic`, or type assertions in `v*`.

### Package catalog (54 v* packages)

| Package | Domain | internal |
|---------|--------|----------|
| vai | AI adapters | internal/ai |
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
| vcli | CLI helpers | internal/cli |
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
| vftp | FTP adapters | internal/ftp |
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
| vhan | Han text romanization adapters | internal/pinyin |
| vpoi | office documents (Excel) | internal/poi |
| vssh | SSH/SFTP adapters | internal/ssh |
| vtok | tokenization adapters | internal/tokenize |

### Validation commands

| Command | Scope |
|---------|-------|
| `make test` | Run unit tests |
| `make test-race` / `make coverage-profile` | Race/shuffle tests with coverage |
| `make coverage-report COVERAGE_FILE=<file>` | Print function coverage |
| `make coverage-check COVERAGE_FILE=<file>` | Enforce coverage gates |
| `make doctor` | Diagnose local Go/tooling/Git environment without modifying files |
| `make worktree-check` | Fail when unrelated untracked Go files may pollute tests or commits |
| `make change-policy-check` | Detect change policies from staged, unstaged, and untracked local diff |
| `make security-sensitive-diff` | Detect changes under security-sensitive public facades and their internal implementations |
| `make provider-contract-check` | Verify provider contract facades do not embed concrete providers, credentials, or default network/file access |
| `make arch-imports-check` | Verify facade/internal import direction and heavy dependency isolation |
| `make panic-policy-check` | Verify production panic usage stays inside Must/Panic or documented compatibility paths |
| `make facade-boundary-check` | Verify package docs, unsafe opt-in, and thin-facade boundaries |
| `make agent-evidence` | Emit `/tmp/knifer-go-agent-validation.json` with detected policies, required commands, command attestations, and structured security review evidence |
| `make agent-evidence-check` | Validate Agent evidence schema, policy references, command references, command attestations, security review evidence, risk, and embedded check status |
| `make quick-check` | Fast local: mod-verify → vet → arch → test → api-check → tools-check → ai-context-check → diff-whitespace |
| `make security-check` | Lint + govulncheck |
| `make full-check COVERAGE_FILE=/tmp/coverage.out` | Full pre-push: quick-check + race coverage + coverage gate + lint + vuln |
| `make agent-check` | Default AI/Agent-safe validation gate; delegates to `quick-check` |
| `make agent-full-check COVERAGE_FILE=/tmp/knifer-go-coverage.out` | Full AI/Agent validation gate with coverage, lint, and vulnerability scan |
| `make agent-security-check` | AI/Agent security validation gate |
| `make ci-agent-governance` | CI Agent governance gate; detects change policies and emits validation evidence |
| `make install-hooks` / `make uninstall-hooks` | Enable or disable optional local Git hooks for pre-commit/pre-push validation |
| `make tools-check` | Verify `docs/api/tools.json` matches public facade functions, doc comments, and Example tests |
| `make tools-gen` | Refresh `docs/api/tools.json`; ask first because generated files may change |
| `make docs-check` | Verify generated documentation artifacts are current |
| `make docs-gen` | Refresh generated documentation artifacts; ask first because generated files may change |
| `make ai-context-check` | Validate machine-readable AI metadata, command side effects, facade inventory, and coverage gates |
| `make ci-workflow-check` | Validate GitHub Actions workflow invariants declared in `ai-context.json` |
| `make generate` | Run repository go:generate directives; ask first because generated files may change |
| `make ci-test` | CI test-job gate (mod-verify + vet + tidy-check + diff-check + arch + test-race + coverage-check + api-check + tools-check) |
| `make check` | Alias for `full-check` |
| `UPDATE_API=1 make api-check` | Refresh API snapshot after intentional public API changes |
| `make tools-gen` | Refresh tools catalog after intentional facade, doc comment, or Example changes |
| `go generate ./...` | Lower-level equivalent of `make generate` |
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
- Security-sensitive code: `vhttp`, `vresty`, `vurl`, `vconf`, `vzip`, `vfile`, `vcrypto`, `vjwt`, `vrand`, `vid`, `vdb`, `vcli`, `vai`, `vftp`, `vssh`. See `SECURITY.md`.

### Governance constraints

- **Coverage**: Keep total coverage above the threshold in `ai-context.json`; `bin/check_coverage.sh` reads `ai-context.json` as the default source of truth.
- **Architecture**: `make arch` composes focused gates for provider contracts, import direction, heavy dependency isolation, panic policy, package docs, unsafe reflection opt-in, and thin facade boundaries.
- **API snapshot**: `docs/api/exports.txt` is CI-enforced. Run `UPDATE_API=1 make api-check` after intentional public API changes.
- **Tools catalog**: `docs/api/tools.json` and `docs/api/tools.md` are CI-enforced by `make tools-check`. Run `make tools-gen` after intentional facade, doc comment, or Example changes.
- **Generated docs**: `make docs-check` guards generated documentation artifacts. Run `make docs-gen` after intentional generated-doc changes, then re-run `make docs-check`.
- **AI metadata**: `ai-context.json` is CI-enforced by `make ai-context-check`; update command side-effect metadata, `risk_level`, facades, security-sensitive package lists, or coverage gates when governance inputs change.
- **Change policies**: `ai-context.json.change_type_policies` maps PR change types to required Agent validation commands; keep PR template change types aligned with those policy keys.
- **Agent evidence**: `make agent-evidence` writes `/tmp/knifer-go-agent-validation.json`; use it to summarize detected policies, required commands, command attestations, security-sensitive paths, structured `security_review`, and governance check results.
- **Agent evidence validation**: `make agent-evidence-check` validates the generated evidence JSON against `ai-context.json`; all change policies must require both evidence generation and evidence validation, every required command must have an attestation entry, and security-sensitive changes must carry a validated `security_review` conclusion.
- **CI Agent governance**: GitHub Actions runs `make ci-agent-governance` with `AGENT_CHANGE_BASE_REF` so policy detection is based on the PR or push diff, then uploads the Agent evidence JSON artifact.
- **CI workflow invariants**: `ai-context.json.ci_workflows` declares required GitHub Actions jobs, Agent governance commands, environment variables, and artifacts; `make ci-workflow-check` validates them.
- **Security-sensitive diff**: `make security-sensitive-diff` checks staged, unstaged, and untracked paths against `ai-context.json.security_sensitive_packages` and their mapped `internal/*` implementations.
- **Provider contracts**: `make provider-contract-check` validates `vai`, `vftp`, `vhan`, `vssh`, and `vtok` provider-contract boundaries independently from the broader architecture gate.
- **API freeze decisions**: `make api-freeze-check` validates that every generated API status is covered by `api_freeze.api_status_decision_cards`, and each mapped decision card has the same status and covers the package.
- **aiflow layout**: `aiflow.yaml` is committed at the repository root. `.aiflow/` is ignored and reserved for generated runtime evidence, traces, reports, scratch files, and temporary state. `make aiflow-layout-check` enforces this boundary.
- **Panic**: Production code must not introduce new `panic()` calls unless in a `MustXxx`/`PanicXxx` function.

---

## Workflows

### General change validation and delivery

When the user asks to implement, rename, refactor, document, or otherwise modify repository files, handle the change end-to-end without stopping after edits:

1. **Inspect** the existing code and documentation first so the change follows current package boundaries, naming, and README/API snapshot conventions.

2. **Apply** only the requested logical change. Do not include unrelated local files, generated experiments, secrets, or user-owned untracked files.

   Run `make worktree-check` before broad validation. If unrelated untracked Go files exist, do not stage, edit, delete, or rely on them unless the user explicitly includes them in scope. Set `SKIP_WORKTREE_CHECK=1` only when those files are intentionally excluded and report them.

3. **Format** touched Go files with `gofmt -w` before validation. Go source must keep real tab indentation, not space-indented visual alignment. For editor display, follow `.editorconfig`: `.go` files use `indent_style = tab`, `indent_size = tab`, and `tab_width = 4` so one literal tab renders as one tab stop instead of appearing like eight spaces. If `golangci-lint` reports `gofmt`, run `gofmt -w` on the reported files, then re-run the focused package tests and lint gate.

4. **Validate** focused tests first, then broaden to repository-level gates when feasible:
   - `go test -v -gcflags="all=-l -N" ./<changed-package>` for affected Go packages.
   - `go test -v -gcflags="all=-l -N" ./...` for repository-wide regressions.
   - `go vet ./...` after code or public API changes.
   - `bash bin/check_arch.sh` after production code changes.
   - `make change-policy-check` to detect which `ai-context.json.change_type_policies` apply to the local diff.
   - `make security-sensitive-diff` after production changes to identify whether security validation is required.
   - `bash bin/check_api_compat.sh`; if the public API change is intentional, run `UPDATE_API=1 bash bin/check_api_compat.sh` and re-run the check. Public facade additions must update `docs/api/exports.txt` in the same logical change.
   - `make tools-check`; if facade functions, doc comments, or Example tests changed intentionally, run `make tools-gen` and re-run `make tools-check`. Keep `docs/api/tools.json` and `docs/api/tools.md` in the same logical change.
   - `make docs-check`; if generated documentation artifacts changed intentionally, run `make docs-gen` and re-run `make docs-check`. Keep generated artifacts in the same logical change.
   - `golangci-lint run ./...` after non-trivial Go code or test changes when the tool is available. Because lint analyzes untracked Go files too, either include intentional untracked Go files in the logical change and format/fix them, or exclude/stash unrelated ones before broad lint.
   - For coverage gates, first generate a fresh profile, then pass that exact file to the checker, e.g. `go test -race -shuffle=on -coverprofile=/tmp/knifer-go-coverage.out ./...` followed by `bash bin/check_coverage.sh /tmp/knifer-go-coverage.out`. Do not rely on an implicit or stale `coverage.out`.
   - `git diff --check` before committing.
   - `make agent-evidence` after governance checks to generate `/tmp/knifer-go-agent-validation.json` for PR evidence.
   - `make agent-evidence-check` after generating Agent evidence to verify policy, command, attestation, risk, and embedded check consistency.
   - Prefer the named workflow targets when they match the change scope: `make agent-check` for the default AI/Agent-safe gate, `make agent-security-check` for lint/vulnerability gates, `make agent-full-check COVERAGE_FILE=/tmp/knifer-go-coverage.out` for the full AI/Agent gate, and `make ci-test` for the CI test-job gate. For security-sensitive changes, use the `security_sensitive` policy in `ai-context.json.change_type_policies`.

5. If validation fails, **fix the cause** and re-run the failing command before reporting completion.

6. Before committing, **re-check the final staged logical change**:
   - Run `git status --porcelain=v1 -b` and review `git diff --stat` / `git diff --staged --stat` so the commit contains only the requested files.
   - Ensure the latest validation was run after the final edit/API snapshot update, not before it. If lint required formatting or simplification fixes, re-run lint after those final edits.
   - For non-trivial Go changes, the pre-commit validation set should include: focused package tests, `go test -v -gcflags="all=-l -N" ./...`, `go vet ./...`, `bash bin/check_arch.sh`, `bash bin/check_api_compat.sh`, `make tools-check`, `golangci-lint run ./...`, `go test -race -shuffle=on -coverprofile=/tmp/knifer-go-coverage.out ./...`, `bash bin/check_coverage.sh /tmp/knifer-go-coverage.out`, and `git diff --check`. `make agent-full-check COVERAGE_FILE=/tmp/knifer-go-coverage.out` is the preferred AI/Agent aggregate target when a full local gate is feasible.
   - If a public API snapshot was intentionally refreshed, run the API check once to observe the stale snapshot, then `UPDATE_API=1 bash bin/check_api_compat.sh`, then re-run `bash bin/check_api_compat.sh` and include `docs/api/exports.txt` in the same logical commit.
   - If the tools catalog was intentionally refreshed, run `make tools-check` once to observe drift, then `make tools-gen`, then re-run `make tools-check` and include `docs/api/tools.json` and `docs/api/tools.md` in the same logical commit.
   - If generated docs were intentionally refreshed, run `make docs-check` once to observe drift, then `make docs-gen`, then re-run `make docs-check` and include generated artifacts in the same logical commit.

7. **Commit**: Generate a conventional commit message from the actual diff, preferring concise messages such as `feat: ...`, `fix: ...`, `docs: ...`, `refactor: ...`, or `test: ...`. Stage only files belonging to the requested logical change, commit them.

8. **Local hooks**: `make install-hooks` and `make uninstall-hooks` modify local Git config. Do not run them unless the user explicitly asks to enable or disable hooks.

9. **Push** the branch to the configured remote when the user asks to commit/push or when the workflow explicitly requires it.

10. After pushing, run `git status --porcelain=v1 -b` to confirm the branch is clean/in sync, then **report** the commit hash, pushed branch, validation commands, and any intentionally excluded local files.

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

5. Any new public facade function in `vslice` or `vmap` must update `docs/api/exports.txt` via the API compatibility workflow, update `docs/api/tools.json` and `docs/api/tools.md`, and include focused tests in both the internal package and facade package.

6. Validate collection changes with focused tests first (`./internal/slice ./vslice` or `./internal/maps ./vmap`), then run the normal repository gates from the general workflow before reporting completion.

### Roadmap sprint workflow

When the user asks to update the workflow, continue the roadmap, or uses equivalent Chinese wording such as "更新工作流", treat it as a roadmap-state synchronization task before starting more feature work:

1. Read `docs/api/tools.json`, `docs/doc/README.md`, existing sprint plans under `docs/superpowers/plans/`, and the latest commits to determine the current baseline, completed sprint, active sprint, and next sprint. If `docs/doc/49-roadmap.md` exists, treat it as supplemental state; do not recreate it unless the user explicitly asks for roadmap restoration.
2. Update only existing workflow, documentation hub, or sprint-plan files unless the user explicitly asks for a new planning document. Keep Sprint status, baseline metrics, scenario guidance, and validation commands aligned with the actual repository state.
3. If a sprint was just completed, record the commit hash when available in the relevant sprint plan or documentation hub, and set the next sprint as `Current` in the active planning document. Do not mark a future sprint as completed without committed code or documentation evidence.
4. For the next sprint, add the concrete execution loop: scope, first package/files to inspect, tests/examples/benchmarks/docs expected, and validation gates. Prefer TDD for API behavior and deterministic examples for reader-facing APIs.
5. If the workflow update changes only Markdown or agent instructions, validate with `git diff --check` and `make docs-check`. If it also changes generated artifacts or Go code, follow the broader validation workflow for those touched areas.
6. Report the files updated, current sprint pointer, validation commands, and whether a commit was created. Do not create a commit unless the user explicitly requested committing in this turn.

### Governance audit

When the user asks to continue general governance, generate next governance tasks, or uses equivalent Chinese wording such as "继续治理" or "生成下一步治理任务", treat it as an autonomous quality pass over security, public facade usability, coverage stability, and benchmark baselines:

1. Start with safety checks and repository context:
   - Run `git status --porcelain=v1 -b` and avoid mixing unrelated local files into the current logical change.
   - Run `govulncheck ./...` before changing dependencies or security-sensitive code. If vulnerabilities are reported, classify reachable findings separately from dependency-only findings and do not upgrade dependencies blindly.

2. Review security suppressions and random/entropy boundaries:
   - Search `#nosec` and confirm each suppression has a narrow reason tied to the operation, especially `G304`, `G115`, `G404`, `G110`, `G103`, and `G204`.
   - Keep `math/rand` confined to non-security helpers, deterministic tests, or documented compatibility fallbacks. Security-sensitive bytes must use fail-closed crypto-random helpers.

3. Audit public facade usability after internal improvements:
   - Compare recent internal or facade changes with `docs/api/exports.txt`, `docs/api/tools.json`, `docs/api/tools.md`, and quickstart/example coverage.
   - Add executable `ExampleXxx` tests for reader-facing behavior, especially iterator adapters in `vslice`/`vmap` and security-sensitive examples in `vrand`, `vid`, and `verr`.
   - Keep examples deterministic; sort map-derived output before printing.

4. Use coverage data to choose small, stable test improvements:
   - Generate a fresh profile with `go test -coverprofile=/tmp/knifer-go-coverage-audit.out ./...` and inspect `go tool cover -func=/tmp/knifer-go-coverage-audit.out`.
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
   - `go test -race -shuffle=on -coverprofile=/tmp/knifer-go-coverage.out ./...` followed by `bash bin/check_coverage.sh /tmp/knifer-go-coverage.out`.
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
   - `go test -race -shuffle=on -coverprofile=/tmp/knifer-go-coverage.out ./...`
   - `bash bin/check_coverage.sh /tmp/knifer-go-coverage.out`
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
