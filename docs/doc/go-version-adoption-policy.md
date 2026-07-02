# Go Version Adoption Policy

`knifer-go` currently targets Go 1.25 or later. The CI matrix verifies Go
1.25.11 and the next Go line, and release builds use the pinned Go 1.25 patch
version declared in GitHub Actions.

## Current Decision

| Decision | Status | Rationale |
| --- | --- | --- |
| Minimum supported Go version | Go 1.25 | The module uses Go 1.25 syntax and tooling expectations, including `testing.B.Loop` in benchmark code. |
| CI compatibility check | Go 1.25.11 and Go 1.26 | CI verifies the supported minimum and the next Go line without claiming older versions are supported. |
| Release toolchain | Go 1.25.11 | Release workflow pins the patch version for repeatable release gates. |
| Go 1.23/1.24 downgrade | Not supported today | Downgrade would require replacing Go 1.25 benchmark loops and revalidating generated docs, snapshots, lint, race, and release gates. |

## Compatibility Boundaries

- Public API compatibility is about `v*` facade signatures and behavior, not
  claiming support for older Go toolchains.
- Iterator-style helpers can expose Go 1.23 range adapters where useful, but
  that does not lower the module's minimum toolchain.
- New benchmark code may use Go 1.25 `b.Loop`; this keeps benchmark style
  consistent with the current toolchain policy.
- A future downgrade proposal must include a focused branch, generated API
  snapshots, full test/race/coverage gates, lint, `govulncheck`, docs checks,
  and release workflow updates.

## Validation

Run these checks after changing Go version policy, workflows, Makefile gates, or
toolchain metadata:

```bash
make ai-context-check
make ci-workflow-check
make governance-maturity-check
make docs-check
```
