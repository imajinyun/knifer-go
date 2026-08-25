# Changelog

All notable changes to this project are documented in this file.

This project follows [Semantic Versioning](https://semver.org/). Public
subpackage APIs are treated as the compatibility boundary.

## Unreleased

### Added

- Added `vgeo` coordinate conversion helpers for WGS-84, GCJ-02, BD-09,
  BD-09 Mercator, coarse China-bound checks, and Haversine distance.
- Added provider-contract facades `vai`, `vftp`, `vssh`, `vhan`, and `vtok`
  so callers inject chat, FTP, SSH/SFTP, pinyin, and tokenization providers
  without pulling concrete clients into the core module.
- Added `vcrypto` TOTP/HOTP helpers with injected clocks, window verification,
  Base32 secrets, and `otpauth` URL formatting.
- Added `vcrypto` Argon2id-style encoded password hashing and verification
  with bounded test costs and malformed-hash errors.
- Added local RSA JWK/JWKS helpers for parse/marshal and `kid` selection,
  without remote JWKS discovery or key rotation.
- Added SM2/SM3/SM4 helpers for interoperability-only workflows, with
  SM4-ECB documented as non-default.
- Added `vcsv` helpers for configurable record reading/writing, map
  conversion, struct tag export, and per-record callbacks.
- Added `vimg` helpers for proportional thumbnails, PNG/JPEG/GIF conversion,
  metadata introspection, QR/barcode workflows, and graphical captchas.
- Added `vpass` password strength helpers for deterministic local scoring,
  strength buckets, and common weak-password checks.
- Added `vstr` Unicode escape helpers, Ant-style path matching, rune-set
  Jaccard similarity, rune n-gram similarity, SimHash, and Hamming distance.
- Added `vref` nil-safe reflection helpers for type classification, interface
  checks, exported field discovery, and object-level predicates.
- Added generic `vnum` sum, average, min, max, and absolute-value APIs.
- Expanded `vmail` with account-based quick send helpers, SMTP envelope
  sender control, lazy attachments, and RFC 2231 filename parameters.
- Added a repository security policy for private vulnerability reporting and
  security-sensitive package review areas.
- Added OpenSSF Scorecard to CI and the README trust signals.

### Changed

- Standardized all 55 facade quickstarts with helper selection guidance,
  safety checklists, when-not-to-use boundaries, related packages,
  benchmark notes, and FAQs.
- Added golden-path Example tests for `verr` and `vid`.
- Added golden-path Example tests for `vpoi` and `vcsv`.
- Aligned public facade counts and catalogs to 55 packages, including `vgeo`.
- Defined public API stability levels for stable `v*` facades, internal
  packages, and experimental provider or adapter contracts.
- Documented breaking-change rules, the two-minor-release deprecation window,
  and the required release-note sections.
- Scoped the exported API snapshot to the module root and top-level `v*`
  facades, then expanded it from symbol names to signatures, types, fields,
  interface methods, and method sets.
- Updated object equality helpers so `time.Time` values compare by instant
  while preserving cross-numeric value comparison.
- Made an isolated Go build cache under `/tmp` the default for Agent,
  governance, API, and documentation Make targets, with explicit `GOCACHE`
  overrides still honored.
- Added Makefile-driven stability gates so local validation and CI share
  module, vet, architecture, race/shuffle, coverage, lint, and vulnerability
  targets.
- Set the repository coverage gate to 75.2% statement coverage, with an 80%
  minimum for security-sensitive packages that contain statements.

### Deprecated

- None.

### Removed

- None.

### Fixed

- Fixed quoted `Content-Disposition` filename parsing when parameters follow
  the filename token.
- Fixed package-level coverage accounting so race-mode profiles count each
  statement once instead of multiplying by execution count.
- Hardened `make doctor` so required diagnostics preserve stderr and fail on
  stderr output or a non-zero status.
- Hardened database mutation guards, upsert conflict validation, secure
  random byte failure handling, and zip extraction destination safety tests.

### Security

- Safe HTTP/URL helpers continue to reject local, private, link-local, and
  unspecified targets by default and re-check redirect targets.
- `vfile`, `vconf`, `vurl`, and `vzip` keep bounded reads or
  extraction/decompression limits by default; ZIP extraction rejects
  traversal and symlink escape.
- Security-sensitive byte helpers fail closed on entropy errors instead of
  falling back to pseudo-random bytes.
- JWT helpers reject unsigned `alg=none` tokens; HMAC strict constructors
  reject weak keys.
- Provider-contract facades do not open network connections, read
  credentials, or touch local filesystem paths; callers inject providers.
- Pinned CI and release Go 1.25 patch to 1.25.13 so `govulncheck` uses
  the standard-library fixes in `net/url`, `html/template`, `crypto/tls`,
  `net/http`, and `encoding/xml`.

### Migration

- Prefer `Safe`, `E`, and `WithOptions` helpers at trust boundaries. Keep
  zero/default fallbacks for trusted inputs and compatibility call sites.
- `MustXxx` APIs remain available as compatibility helpers. New code should
  use `vcron.NewPattern`, `vjwt.NewHMACSigner` or `NewHMACSignerStrict`,
  `vobj.DeserializeTo`, and return errors instead of `verr.MustExit`.
- Experimental APIs are blocked while the module is a v1 candidate.
- `vtest` and `vdump` are not public facades; keep using `testing`, `vcli`,
  `vsys`, `vfile`, and `vlog`.
- Do not import `internal/*` from application code.
