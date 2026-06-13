# Changelog

All notable changes to this project are documented in this file.

This project follows [Semantic Versioning](https://semver.org/). Public
subpackage APIs are treated as the compatibility boundary.

## Unreleased

### Governance

- Added a repository security policy for private vulnerability reporting,
  supported versions, and security-sensitive package review areas.
- Added a coverage gate script so CI can enforce a measurable test baseline.
- Added facade coverage tests for `vurl`, `vzip`, `vdb`, `vjwt`, `vhttp`,
  `vnum`, `vtpl`, `vmap`, `vref`, `vobj`, `vmask`, `vresty`, `vconf`,
  `vcrypto`, and `vfile`.
- Added core HTTP client coverage tests for `internal/httpx/http` and
  `internal/httpx/resty`.
- Added package-level coverage gates for security-sensitive facade packages:
  `vhttp`, `vresty`, `vconf`, `vzip`, `vcrypto`, `vurl`, `vfile`,
  `internal/httpx/http`, and `internal/httpx/resty`.
- Added Makefile-driven stability gates so local validation and CI share the
  same module, vet, architecture, race/shuffle test, coverage, lint, and
  vulnerability check targets.
- Fixed package-level coverage accounting so race-mode coverage profiles count
  each statement as covered once instead of multiplying by execution count.
- Added `vref` reflection helper APIs for nil-safe `reflect.Value` checks,
  type classification, interface implementation checks, and exported field name
  discovery.
- Added `vref` object-level type predicate helpers for functions, iteratees,
  collections, slices, arrays, and maps.
- Added internal generic numeric constraints for shared implementation helpers
  and exposed generic `vnum` sum, average, min, max, and absolute-value APIs.
- Added direct coverage for `internal/httpx/internal/shared` so HTTP protocol
  helpers are validated before being wrapped by `vhttp` and `vresty`.
- Fixed quoted `Content-Disposition` filename parsing when parameters follow
  the filename token.
- Documented release notes in a changelog so user-visible changes can be
  reviewed before tagging.

### Quality targets

- Current coverage gate baseline: 74.2%.
- Current security-sensitive package gates: `vhttp` >= 75%, `vresty` >= 65%,
  `vconf` >= 75%, `vzip` >= 80%, `vcrypto` >= 70%, `vurl` >= 80%,
  `vfile` >= 85%, `internal/httpx/http` >= 75%, and
  `internal/httpx/resty` >= 75%, and `internal/httpx/internal/shared` >= 80%.
- Near-term target: 75% total statement coverage.
- Longer-term target: 80% total statement coverage, with priority on public
  facade packages and security-sensitive packages.

### Review focus

- Prioritize tests for `vhttp`, `vresty`, `vurl`, `vconf`, `vjwt`, `vzip`,
  `vcrypto`, `vdb`, and other packages that process untrusted input.
- Keep `v*` facade packages thin and preserve the `v* -> internal/*`
  dependency direction.
