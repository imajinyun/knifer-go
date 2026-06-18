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
- [ ] I ran `make agent-check` or documented why it is not applicable.
- [ ] I ran `make worktree-check` or documented unrelated untracked files below.
- [ ] I ran `make security-sensitive-diff` or documented why it is not applicable.
- [ ] I ran `go test -race -shuffle=on -coverprofile=/tmp/go-knifer-coverage.out ./...` when the change is non-trivial.
- [ ] I ran `bash bin/check_coverage.sh /tmp/go-knifer-coverage.out` when a fresh coverage profile was generated.
- [ ] I ran `bash bin/check_arch.sh`.
- [ ] I ran `golangci-lint run ./...`.
- [ ] I updated `CHANGELOG.md` for user-visible changes.

## Validation

- Commands run:
  -
- Commands intentionally skipped and reason:
  -

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
