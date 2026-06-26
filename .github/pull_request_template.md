## Summary

- 

## Change type

- [ ] Bug fix
- [ ] New public API
- [ ] Internal refactor
- [ ] Documentation
- [ ] CI / governance
- [ ] Security-sensitive code
- [ ] Dependency change

## Checklist

- [ ] I kept `v*` packages as thin facades over `internal/*`.
- [ ] I avoided new dependencies in public facades, or updated the architecture allowlist with justification.
- [ ] I added or updated tests for changed behavior.
- [ ] I formatted touched Go files with `gofmt -w` before running lint.
- [ ] I ran `make agent-check` or documented why it is not applicable.
- [ ] I ran `make worktree-check` or documented unrelated untracked files below.
- [ ] I ran `make change-policy-check` and applied the detected policy below.
- [ ] I ran `make security-sensitive-diff` or documented why it is not applicable.
- [ ] I ran `make agent-evidence` and reviewed `/tmp/knifer-go-agent-validation.json`.
- [ ] I ran `go test -race -shuffle=on -coverprofile=/tmp/knifer-go-coverage.out ./...` when the change is non-trivial.
- [ ] I ran `bash bin/check_coverage.sh /tmp/knifer-go-coverage.out` when a fresh coverage profile was generated.
- [ ] I ran `bash bin/check_arch.sh`.
- [ ] I ran `golangci-lint run ./...`.
- [ ] I updated `CHANGELOG.md` for user-visible changes.
- [ ] Public facade API changed and `UPDATE_API=1 make api-check` was run.
- [ ] Examples were added or updated for reader-facing APIs.
- [ ] Package quickstart docs were updated for behavior changes.
- [ ] `make docs-check` and `make tools-check` pass.
- [ ] Security-sensitive behavior was reviewed for safe defaults and secret leakage.
- [ ] Benchmarks were added or updated for performance-sensitive changes.

## Validation

- Commands run:
  -
- Commands intentionally skipped and reason:
  -
- Formatting/lint notes:
  -
- Agent evidence:
  - `/tmp/knifer-go-agent-validation.json` generated: yes / no
  - Detected change policies:
  - Required commands:

## API and coverage impact

- Public API changed: yes / no
- `docs/api/exports.txt` updated: yes / no / not applicable
- Change policy applied: bug_fix / public_api / internal_refactor / documentation / ci_governance / security_sensitive / dependency_change
- Coverage impact:
  -

## Reviewer focus

-

## Intentionally excluded files

-

## AI / Agent assistance

- [ ] This PR was not AI-assisted.
- [ ] This PR was AI-assisted and the generated changes were reviewed.
- Agent/tools used:
  -
- Highest Agent command risk level used: low / medium / high / forbidden_for_agent / not applicable
- Commands run by Agent:
  -
- Commands intentionally skipped by Agent and reason:
  -
- User-consent commands executed by Agent:
  -
- Unrelated local/untracked files intentionally excluded:
  -
- [ ] No local hook or Git config changes are included unless explicitly requested.

## Security review

- [ ] This change does not touch security-sensitive code.
- [ ] This change touches security-sensitive code and includes regression tests plus `make agent-security-check` evidence.
- [ ] Any security linter suppression is narrow and documented.
