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
