# Agent Instructions

## General change validation and delivery workflow

When the user asks to implement, rename, refactor, document, or otherwise modify repository files, handle the change end-to-end without stopping after edits:

1. Inspect the existing code and documentation first so the change follows current package boundaries, naming, and README/API snapshot conventions.
2. Apply only the requested logical change. Do not include unrelated local files, generated experiments, secrets, or user-owned untracked files.
3. Format touched Go files with `gofmt -w` before validation.
4. Run focused validation for the changed package or area first, then broaden to repository-level gates when feasible:
   - `go test -v -gcflags="all=-l -N" ./<changed-package>` for affected Go packages.
   - `go test -v -gcflags="all=-l -N" ./...` for repository-wide regressions.
   - `go vet ./...` after code or public API changes.
   - `bash bin/check_arch.sh` after production code changes.
   - `bash bin/check_api_compat.sh`; if the public API change is intentional, run `UPDATE_API=1 bash bin/check_api_compat.sh` and re-run the check. Public facade additions must update `docs/api/exports.txt` in the same logical change.
   - `golangci-lint run ./...` after non-trivial Go code or test changes when the tool is available.
   - For coverage gates, first generate a fresh profile, then pass that exact file to the checker, e.g. `go test -race -shuffle=on -coverprofile=/tmp/go-knifer-coverage.out ./...` followed by `bash bin/check_coverage.sh /tmp/go-knifer-coverage.out`. Do not rely on an implicit or stale `coverage.out`.
   - `git diff --check` before committing.
5. If validation fails, fix the cause and re-run the failing command before reporting completion.
6. Before committing, re-check the final staged logical change:
   - Run `git status --porcelain=v1 -b` and review `git diff --stat` / `git diff --staged --stat` so the commit contains only the requested files.
   - Ensure the latest validation was run after the final edit/API snapshot update, not before it.
   - For non-trivial Go changes, the pre-commit validation set should include: focused package tests, `go test -v -gcflags="all=-l -N" ./...`, `go vet ./...`, `bash bin/check_arch.sh`, `bash bin/check_api_compat.sh`, `golangci-lint run ./...`, `go test -race -shuffle=on -coverprofile=/tmp/go-knifer-coverage.out ./...`, `bash bin/check_coverage.sh /tmp/go-knifer-coverage.out`, and `git diff --check`.
   - If a public API snapshot was intentionally refreshed, run the API check once to observe the stale snapshot, then `UPDATE_API=1 bash bin/check_api_compat.sh`, then re-run `bash bin/check_api_compat.sh` and include `docs/api/exports.txt` in the same logical commit.
7. Generate a conventional commit message from the actual diff, preferring concise messages such as `feat: ...`, `fix: ...`, `docs: ...`, `refactor: ...`, or `test: ...`.
8. Stage only files belonging to the requested logical change, commit them, and push the branch to the configured remote when the user asks to commit/push or when the workflow explicitly requires it.
9. After pushing, run `git status --porcelain=v1 -b` to confirm the branch is clean/in sync, then report the commit hash, pushed branch, validation commands, and any intentionally excluded local files.

## Package test governance workflow

When the user asks to continue package test governance, package-level test cleanup, or uses equivalent Chinese wording such as “继续推进包测试治理”, treat it as an autonomous rolling workflow:

1. Continue round by round without asking for confirmation.
2. In each round, split one aggregated tracked Go `*_test.go` file into smaller test files organized by source-file responsibility.
3. Preserve existing test behavior unless a real defect is proven by code evidence.
4. After each completed round, write a conventional commit, commit only that round's changes, push to the remote, then automatically start the next round.

### Required unit-test generation protocol

For every round, follow the `bits-unit-test-gen` workflow strictly:

1. Step1: run the skill preparation script and record `BITS_TMP_ROOT` as `TMP_ROOT`.
2. Step2: load language knowledge and project context.
3. Step3: determine `TARGETS` before editing tests.
4. Step4: produce `BUG_MAP` before generating or moving test cases.
5. Step5: report each generate → verify → fix loop with command, failure type, and result.
6. Step6: report scope, defect analysis, generated test cases, validation result, commit, and push status.

When the round only reorganizes existing tests, `BUG_MAP` may be empty, but record a short candidate-filtering summary.

### Target selection rules

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

### Validation rules

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

### Commit and push rules

Commit only the current round's staged split files. Do not include unrelated local changes.

Use commit messages in this style:

```text
test: split <package-or-area> tests by responsibility
```

After a successful commit, run `git push`. If validation fails, fix the round before committing.

### Repository-specific gates

- Keep total coverage above the repository gate reported by `bin/check_coverage.sh`.
- `bin/check_arch.sh` must pass; production code must not introduce forbidden panic patterns.
- Do not modify production code during package test governance unless a verified test-governance blocker requires it and the final report calls it out.
