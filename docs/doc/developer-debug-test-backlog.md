# Developer Debug/Test Utilities Backlog

Use this page to track possible developer-facing debug and test utility
facades. This is a backlog, not an API promise. The current public path remains
`vcli`, `vsys`, `vfile`, `vlog`, package-local test helpers, and the Go
standard library.

## Current Entry Points

| Need | Use now | Boundary |
| --- | --- | --- |
| command execution in tests | `vcli` with injected runners | Avoid shell concatenation and host PATH dependencies. |
| system and runtime evidence | `vsys` | Review environment-derived values before publishing logs or fixtures. |
| diagnostic file reads/writes | `vfile` | Keep path policy and size limits visible. |
| diagnostic logging | `vlog` | Use isolated loggers and injected outputs for deterministic tests. |
| assertions | Go `testing` package or project-local helpers | Do not add broad assertion helpers unless they improve repeated project workflows. |
| object dumps | `vsys.DumpSystemInfo`, `vjson`, `vobj`, or caller-owned formatting | Avoid dumping secrets, tokens, environment values, or large payloads by default. |

## Candidate Lanes

| Candidate | Status | Possible scope | Non-goals |
| --- | --- | --- | --- |
| `vtest` | planned only | small test helpers for fixtures, temporary providers, golden file policy, and assertion wrappers that preserve standard `testing` semantics | Replacing `testing`, `testify`, fuzzing, race detection, or integration-test frameworks. |
| `vdump` | planned only | safe object/system dump helpers with redaction hooks, size limits, and deterministic formatting for diagnostic evidence | Long-running collectors, resident background processes, broad logging framework behavior, or secret-leaking dumps. |

## Adoption Rules

- vtest is a planned lane, not a current public facade.
- vdump is a planned lane, not a current public facade.
- Do not document `vtest` or `vdump` as available API until they appear in
  `docs/api/tools.json` and `ai-context.json`.
- Prefer standard Go testing first.
- Prefer package-local helpers when only one package needs the behavior.
- Debug dumps must have redaction hooks and size limits before becoming public.
- No resident background utility process.
- No broad assertion framework replacement.

Machine-readable boundaries:

- test helpers
- object dumps
- system dumps
- redaction hooks
- size limits
- golden file policy
- replacing testing
- replacing testify
- resident background process
- secret-leaking dumps
- broad assertion framework replacement

## Open Questions

1. Which repeated test-helper workflows exist across three or more facades?
2. Which diagnostic dump outputs are safe to include in issue reports or CI artifacts?
3. Should redaction policies reuse `vmask`, explicit callbacks, or both?
4. Should golden-file helpers live in `vtest`, or stay package-local until release workflows need them?
5. Which workflows need examples before any public API is proposed?

## API Decision Backlog v2

This section answers the current decision questions before any `vtest` or
`vdump` package is created.

| Question | Current answer | Decision |
| --- | --- | --- |
| Which workflows repeat across three or more facades? | Provider injection tests, temporary filesystem fixtures, deterministic output capture, and generated-artifact drift checks appear across `vcli`, `vsys`, `vfile`, `vconf`, `vhttp`, `vresty`, `vlog`, and `vzip`. | Keep these as candidates for `vtest`, but require a concrete API decision card before implementation. |
| Which dump outputs are safe for issue reports or CI artifacts? | System/runtime summaries, bounded file metadata, redacted config keys, and explicit diagnostic logs can be safe when size limits and redaction hooks are present. | Keep these as candidates for `vdump`, with redaction hooks and size limits as mandatory design inputs. |
| Should redaction reuse `vmask`, callbacks, or both? | `vmask` covers common masking formats, while callers still need domain-specific redaction. | Use explicit callbacks first; allow `vmask` helpers as optional building blocks. |
| Should golden-file helpers become public? | Golden-file policy is useful across generated docs, API snapshots, and examples, but each package currently owns its local fixtures. | Keep golden-file helpers package-local until at least three facades need the same public workflow. |
| Would this duplicate `testing`, `testify`, `fmt`, or `slog`? | Broad assertions duplicate `testing` / `testify`; formatting-only dumps duplicate `fmt`; logging abstractions duplicate `vlog` / `slog`. | Reject broad assertion framework replacement and broad logging replacement. Prefer tiny fixtures, redaction, and deterministic evidence helpers only. |

## Candidate API Cards

| Candidate | Minimum API card scope | Required evidence |
| --- | --- | --- |
| `vtest` fixture helpers | temporary directory/file fixtures, provider injection helpers, and golden file comparison policy | examples across at least three facades, `go test` coverage, and docs catalog entry |
| `vtest` assertion helpers | only narrow helpers that preserve standard `testing.T` semantics | proof they do not replace `testing` or `testify` |
| `vdump` object dump helpers | deterministic formatting, redaction hooks, size limits, and caller-owned writers | secret redaction tests and size-limit tests |
| `vdump` system dump helpers | wrappers around existing `vsys` evidence with redaction and bounded output | examples showing no raw environment secrets by default |

Machine-readable decision v2:

- repeated workflows across three or more facades
- provider injection tests
- temporary filesystem fixtures
- deterministic output capture
- generated-artifact drift checks
- safe issue report dumps require redaction hooks and size limits
- use explicit callbacks first for redaction
- vmask helpers are optional building blocks
- keep golden-file helpers package-local until three facades need them
- reject broad assertion framework replacement
- reject broad logging replacement
- vtest fixture helpers
- vtest assertion helpers
- vdump object dump helpers
- vdump system dump helpers

## Exit Criteria Before Implementation

- A public API decision card names the exact workflow, package boundary, and
  alternatives.
- The API does not duplicate `testing`, `testify`, `slog`, `fmt`, or
  package-local helpers without a repeated workflow.
- Security-sensitive output has redaction tests.
- Generated docs and `ai-context.json` list the new facade only after the public
  package exists.
- `make docs-check`, `make ai-context-check`, `make governance-maturity-check`,
  and `make agent-check` pass after the implementation.
