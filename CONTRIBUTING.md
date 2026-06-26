# Contributing to knifer-go

Thanks for contributing! knifer-go is a large, multi-domain Go utility library
(54 public facade subpackages). To keep it consistent and maintainable at this scale,
please follow the conventions below. Most of them are enforced by CI
(`go vet`, `golangci-lint`, and `bin/check_arch.sh`).

## Documentation map

- Start with [`README.md`](README.md) for project overview, installation, and package discovery.
- Use [`docs/doc/README.md`](docs/doc/README.md) as the documentation hub and package catalog.
- Use [`docs/api/exports.txt`](docs/api/exports.txt) and [`docs/api/tools.md`](docs/api/tools.md) for public API and tool-catalog review.
- Read [`SECURITY.md`](SECURITY.md) before changing security-sensitive facades or their `internal/*` implementations.
- Keep user-visible changes in [`CHANGELOG.md`](CHANGELOG.md).
- Agents should follow [`AGENTS.md`](AGENTS.md); Claude-specific deep workflow details live in [`CLAUDE.md`](CLAUDE.md).

## Project layout

```
<module-root>/
├── errors.go            # root package: cross-cutting error contract only
├── doc.go               # root package overview + domain-grouped navigation
├── v<domain>/           # public facade packages (v prefix), import these
│   ├── doc.go           # required: package doc
│   └── <domain>.go      # thin facade: forwards to internal/<domain>
└── internal/<domain>/   # implementation; not importable by external modules
```

- **Public packages** live in `v<domain>/` and are the only API users import.
- **Implementations** live in `internal/<domain>/` and may evolve freely.
- The **root package `knifer`** exposes no business APIs — only the error
  contract (see below).

## Core rules (enforced by CI)

1. **Facades are thin.** A `v<domain>` function only forwards to
   `internal/<domain>`; it must not contain business logic, loops, `panic`,
   or type assertions. Put all logic in `internal/`.
2. **Public packages never import each other.** No `v*` package may import
   another `v*` package. Shared logic goes down into `internal/` (e.g.
   `internal/common`, or a domain package such as `internal/str`).
3. **Every `v<domain>` has a `doc.go`** with a package comment.
4. **Every `v<domain>` imports at least one `internal/` package** that exists.
5. **Internal packages never import public facades.** The dependency direction
   is always `v* -> internal/*`, never `internal/* -> v*`.
6. **Production `panic` is exceptional.** New panics are only allowed for
   explicit `MustXxx` / `PanicXxx` APIs or documented compatibility cases.

`bin/check_arch.sh` verifies these architecture rules and runs in CI.

## Naming

- Public packages: `v` + domain, all lowercase, no underscores, ≤ 8 chars.
- Prefer the full domain name when short (`vhttp`, `vjson`, `vstr`).
- Abbreviations are allowed for long domains but **must be registered** in the
  table below, and the package `doc.go` must state the full name on its first
  line, e.g. `// Package vskt (socket) ...`.
- `internal/<domain>` uses the bare full domain name (no `v` prefix, no `impl`
  suffix). The facade import alias is `<fulldomain>impl` (e.g. `httpimpl`,
  `bloomfilterimpl`) — use the full name even for abbreviated facades.

### Abbreviation registry

| Public | Full domain | internal |
| --- | --- | --- |
| vblf | bloomfilter | internal/bloomfilter |
| vident | identity | internal/identity |
| vmask | data masking (desensitization) | internal/mask |
| verr | errx (extended error) | internal/errx |
| vsem | semaphore | internal/semaphore |
| vset | sets | internal/sets |
| vskt | socket | internal/socket |
| vsys | system | internal/system |
| vtpl | template | internal/template |
| vform | form/input validation | internal/validator |
| vver | version | internal/version |

## Package placement

Group functions by their **input domain, not their output type**. For example,
a `[]T -> map[K][]T` aggregation belongs with slice utilities (the input is a
slice), even though it returns a map. When a capability already has an owner
package, extend that package instead of duplicating it elsewhere (single source
of truth).

## API design

- Prefer friendly types: `string`, `[]byte`, standard-library types.
- Return `(result, error)` rather than `panic` for recoverable failures.
- IO/network functions take `context.Context` as the first parameter.
- Do not add synonym aliases for the same function.
- Keep exported identifiers stable; renaming a public symbol is a breaking
  change (see Versioning).

### New public API checklist

Before adding or changing a public API, answer these questions in the PR/MR:

1. **Domain owner:** Which `v<domain>` package owns this capability? If a clear
   owner already exists, extend it instead of creating a sibling package.
2. **Implementation boundary:** Is the real logic in `internal/<domain>` and is
   the `v*` layer a thin facade only?
3. **Root package safety:** Does the change keep the root `knifer` package free
   of business helpers? Root additions should be limited to cross-cutting
   contracts such as errors.
4. **Dependency direction:** Does the implementation avoid importing public
   `v*` packages? Shared code should move downward into `internal/*`.
5. **Error behavior:** For recoverable failures, does the API return `error` and
   participate in the root error contract where appropriate (`ErrCode`,
   `CodeCarrier`, `CodeOf`)?
6. **Panic behavior:** If the API can panic, is it named `MustXxx`/`PanicXxx` or
   explicitly documented as a compatibility panic?
7. **Type exposure:** Are concrete stateful structs wrapped by the facade when
   future compatibility requires control over fields/methods? Prefer aliases
   only for stable interfaces, function types, and option/value types.
8. **Documentation:** Does every new public package have `doc.go`, and does
   every new exported identifier have a useful doc comment starting with its
   name?
9. **Tests/examples:** Are internal behavior, facade compatibility, error
   matching, and at least one common usage path covered by tests or examples?
10. **Dependencies:** Does the API avoid introducing heavyweight dependencies
    into otherwise lightweight packages?

## Error contract

The root package owns a thin, dependency-free error contract in `errors.go`:

- `ErrCode` — a stable classifier that is itself an `error`, so it can be the
  target of `errors.Is` directly. Base codes: `ErrCodeInvalidInput`,
  `ErrCodeNotFound`, `ErrCodeUnsupported`, `ErrCodeTimeout`, `ErrCodeInternal`.
- `Error{Code, Message, Cause}` with `Error`/`Unwrap`/`Is`, plus
  `NewError` / `WrapError` / `Errorf`.
- `CodeCarrier` and `CodeOf` — a small extraction contract for existing custom
  errors and sentinels that need to expose a stable code without changing their
  concrete type.

Packages that return rich errors should participate so callers can write
`errors.Is(err, knifer.ErrCodeXxx)` while preserving the cause chain. Custom
error types (e.g. `JWTError`, `JSONError`, `HTTPError`) add a `Code` field,
`ErrorCode`, and an `Is` method that matches an `ErrCode`; wrap underlying
causes with their `Cause` field. Logging uses the existing `vlog` package — do
not add a second logging abstraction.

## Tests & examples

- Black-box facade tests use `package v<domain>_test`.
- Add an `example_test.go` with runnable `ExampleXxx` functions for new public
  packages. Use deterministic `// Output:` assertions where possible; for
  random/IO/time-based results, assert on a stable property (length, boolean).
- Run the full suite before pushing: `go test ./...`.
- Public facade and security-sensitive packages should include tests for common
  usage, invalid input, and error classification.
- Coverage is checked from `coverage.out`; keep the repository baseline passing.
  Security-sensitive packages declared in `ai-context.json` must have fresh
  coverage profile data and meet the configured security-sensitive minimum even
  when an older package-specific gate is lower.
- Raise coverage thresholds only after adding tests that support the new gate.
- Prefer `make check` for local stability validation so vet, architecture,
  race/shuffle tests, coverage, lint, and vulnerability checks stay in one
  documented path.

## Governance checks

- Use `make ci-test` to run the CI test-job gates locally: module verification,
  vet, tidy/diff checks, architecture checks, race/shuffle tests, coverage gates,
  and API compatibility.
- Use `make check` before pushing non-trivial changes. It includes the CI test
  gates plus `golangci-lint` and `govulncheck`.
- Run `make govulncheck` before changing dependencies or security-sensitive code.
  Treat reachable findings separately from dependency-only findings.
- Public facade examples should be executable `ExampleXxx` tests when possible.
  Sort map-derived output and assert deterministic properties for random, time,
  or IO-heavy examples.
- Generate coverage from a fresh profile before enforcing gates:
  `go test -race -shuffle=on -coverprofile=coverage.out ./...` followed by
  `make coverage-check`. Do not rely on a stale `coverage.out`.
- Use `make bench-core` to run benchmark baselines for hot-path slice, map,
  string, and numeric helpers, or `make bench-facade` for the matching public
  facade packages. Benchmark output is a baseline; only claim a performance
  change after a repeated `benchstat` comparison.

## Before you push

```bash
make check
```

## Versioning

- Follow [SemVer](https://semver.org/). The subpackage is the unit of API
  stability.
- Removing or renaming an exported symbol is a breaking change (major bump).
- `v2+` must change the module path (`.../knifer-go/v2`).
- Add user-visible changes to `CHANGELOG.md` before tagging a release.

## Security

- Read `SECURITY.md` before touching SSRF, archive extraction, cryptography,
  JWT, randomness, file IO, network, or database helpers.
- Do not report suspected vulnerabilities in public issues.
- Keep `.golangci.yml` security exclusions narrow and documented; prefer a
  regression test over a broader `gosec` suppression.
- New `#nosec` or `//nolint:gosec` suppressions must include the specific rule
  reason at the call site and should be paired with a boundary or regression
  test when the code touches untrusted input.

## Linter exceptions

`.golangci.yml` documents intentional exceptions, e.g. `SA5012` (a staticcheck
crash on generic variadic forwarding) and `ST1003` (initialisms in stable
public API names are kept on purpose). Add a comment explaining any new
exception you introduce.
