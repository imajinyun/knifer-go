# Changelog

All notable changes to this project are documented in this file.

This project follows [Semantic Versioning](https://semver.org/). Public
subpackage APIs are treated as the compatibility boundary.

## Unreleased

### Documentation

- Standardized all 54 facade quickstarts with helper selection guidance,
  safety and correctness checklists, when-not-to-use boundaries, related package
  guidance, benchmark and trade-off notes, and FAQs.

### Governance

- Defined public API stability levels for stable `v*` facades, internal
  implementation packages, and experimental provider or adapter contracts.
- Documented breaking-change rules, the two-minor-release deprecation window,
  and the required release note sections for user-visible changes and
  migrations.
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
- Strengthened coverage governance so security-sensitive packages declared in
  `ai-context.json` must have coverage profile data.
- Hardened database mutation guards, upsert conflict validation, secure random
  byte failure handling, and zip extraction destination safety tests.
- Added executable examples for database named parameters and updates, random
  option helpers, and ZIP file/filter archive helpers.
- Added a release readiness gate and strengthened release automation with tag
  format, changelog, full validation, and protected environment checks.
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
- Updated object equality helpers so `time.Time` values compare by instant while
  preserving cross-numeric value comparison.
- Added `vcsv` CSV helpers for configurable record reading/writing, map
  conversion, struct tag export, and per-record callbacks.
- Added `vimg` image helpers: proportional thumbnails, lossless format
  conversion between PNG/JPEG/GIF, metadata introspection (width/height/format),
  and image captcha generation through the unified `internal/imgx` implementation
  package.
- Added `vpass` password strength helpers for deterministic local password
  scoring, strength classification, and rule-level analysis.
- Added `vstr` text helpers for Unicode escaping/unescaping and Ant-style path
  matching, plus rune-set Jaccard similarity, rune n-gram similarity, SimHash,
  and 64-bit Hamming distance without introducing a separate text package.
- Scoped the exported API compatibility snapshot to the module root and
  top-level `v*` facades, keeping `internal/*` refactors out of the public API
  gate.
- Expanded the API compatibility snapshot from exported symbol names to
  function signatures, exported type definitions, exported struct fields,
  interface methods, and method sets so breaking public API shape changes are
  detected before release.
- Expanded `vmail` with account-based quick send helpers, SMTP envelope sender
  control, lazy reader/file attachments, and RFC 2231-compliant attachment
  filename parameter rendering.
- Added internal generic numeric constraints for shared implementation helpers
  and exposed generic `vnum` sum, average, min, max, and absolute-value APIs.
- Added direct coverage for `internal/httpx/internal/shared` so HTTP protocol
  helpers are validated before being wrapped by `vhttp` and `vresty`.
- Fixed quoted `Content-Disposition` filename parsing when parameters follow
  the filename token.
- Documented release notes in a changelog so user-visible changes can be
  reviewed before tagging.

### Quality targets

- Current coverage gate baseline: 75.2%.
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
