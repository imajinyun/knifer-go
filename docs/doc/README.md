# 📚 go-knifer Documentation Hub

> Detailed package navigation, architecture notes, safety defaults, and contribution workflow for `go-knifer`.

## 📑 Table of Contents

- [🧭 Quick navigation](#quick-navigation)
- [⭐ Start with star domains](#start-with-star-domains)
- [🧩 Package catalog](#package-catalog)
- [📝 Quickstart documents](#quickstart-documents)
- [🧭 Sprint direction](#sprint-direction)
- [🏗️ Architecture and package boundaries](#architecture-and-package-boundaries)
- [🔒 API compatibility, deprecation, and releases](#api-compatibility-deprecation-and-releases)
- [✅ Recommended API entry points](#recommended-api-entry-points)
- [📦 Build, test, and release workflow](#build-test-and-release-workflow)
- [🛡️ Governance](#governance)
- [🤝 Contributing](#contributing)

<a id="quick-navigation"></a>

## 🧭 Quick navigation

- 🏠 Root README: [`../../README.md`](../../README.md)
- 🌐 Online Go docs: [pkg.go.dev/github.com/imajinyun/go-knifer](https://pkg.go.dev/github.com/imajinyun/go-knifer)
- 🧾 Public API snapshot: [`../api/exports.txt`](../api/exports.txt)
- 🤖 Machine-readable tool catalog: [`../api/tools.json`](../api/tools.json)
- 📋 Readable tool catalog: [`../api/tools.md`](../api/tools.md)
- 🗺️ AI-oriented project map: [`../../llms.txt`](../../llms.txt)
- 🤖 Machine-readable AI/CLI metadata: [`../../ai-context.json`](../../ai-context.json)
- 🧯 Security policy: [`../../SECURITY.md`](../../SECURITY.md)
- 📝 Changelog: [`../../CHANGELOG.md`](../../CHANGELOG.md)

<a id="start-with-star-domains"></a>

## ⭐ Start with star domains

These domains are the quickest way to evaluate whether `go-knifer` fits a project. They combine recommended API entry points, executable examples, cookbook workflows, benchmark commands, and explicit safety boundaries.

| Need | Start here | Trust signals |
| --- | --- | --- |
| Safe HTTP and downloads | [`vhttp`](22-vhttp.md), [`vresty`](41-vresty.md), [`vurl`](51-vurl.md) | Helper selection, safe URL policy checklist, FAQ, benchmark commands, and stdlib/Resty boundary guidance. |
| Safe crypto workflows | [`vcrypto`](11-vcrypto.md), [`vrand`](38-vrand.md), [`vjwt`](28-vjwt.md) | Recommended cryptographic entry points, secret-handling FAQ, benchmark commands, and direct-stdlib boundary guidance. |
| Daily JSON and file workflows | [`vjson`](27-vjson.md), [`vfile`](17-vfile.md) | Cookbook examples for object/path/formatting/file I/O, filesystem safety guidance, and explicit error handling. |

Comparison entry points:

- Use [`vhttp`](22-vhttp.md) for standard-library-style HTTP helpers and [`vresty`](41-vresty.md) when Resty-style request chains improve readability.
- Use [`vurl`](51-vurl.md) for URL construction, normalization, query encoding, resource probing, or safe opening before an HTTP request is needed.
- Use [`vcrypto`](11-vcrypto.md) when recommended crypto workflows reduce misuse risk; use the Go standard library directly when a caller needs lower-level protocol control.
- Use [`vjson`](27-vjson.md) for common object/path/formatting/XML bridge flows; use `encoding/json` directly for streaming, tokenization, or full decoder control.
- Use [`vfile`](17-vfile.md) for bounded reads, provider-backed file-system tests, and explicit file errors; keep untrusted path handling visible at the call site.

<a id="package-catalog"></a>

## 🧩 Package catalog

The project follows an “internal implementation + public facade” layout: `internal/*` contains concrete implementations, while `v*` packages expose stable public APIs.

| Module | Import path | Description |
| --- | --- | --- |
| [`vai`](01-vai.md) | `github.com/imajinyun/go-knifer/vai` | AI adapter helpers: provider-injected chat and embeddings, request validation, defensive copies, deterministic examples, and redaction-safe diagnostics. |
| [`vbean`](02-vbean.md) | `github.com/imajinyun/go-knifer/vbean` | Bean/struct mapping helpers: struct/map conversion, copy properties, tag and alias matching, ignore-empty/zero options, and weak type conversion. |
| [`vblf`](03-vblf.md) | `github.com/imajinyun/go-knifer/vblf` | Bloom filters: bitmap/bitset/filter abstractions, string hash algorithms, option constructors, validation-returning `E` constructors, and provider-backed file initialization. |
| [`vbool`](04-vbool.md) | `github.com/imajinyun/go-knifer/vbool` | Boolean helpers: negate, bool-to-int, all/any checks. |
| [`vcache`](05-vcache.md) | `github.com/imajinyun/go-knifer/vcache` | Generic caches: FIFO, LFU, LRU, Timed, Weak, and NoCache; supports TTL, clocks, removal listeners, lazy loading, ticker/runner providers, and weak-cache finalizer providers. |
| [`vcli`](06-vcli.md) | `github.com/imajinyun/go-knifer/vcli` | CLI helpers: context-aware command execution, injected runners, typed flag parsing, subcommand routing, deterministic help rendering, and ANSI color controls. |
| [`vcodec`](07-vcodec.md) | `github.com/imajinyun/go-knifer/vcodec` | Encoding helpers: Base64, URL-safe Base64, raw URL-safe Base64, custom Base64 encoding providers, and Hex. |
| [`vconf`](08-vconf.md) | `github.com/imajinyun/go-knifer/vconf` | Grouped configuration reader for setting/properties-style text, YAML subset, and TOML parsing, with typed getters, schema validation, profile/remote/file loading, SSRF-checked remote loads, bounded reads, and clone support. |
| [`vconv`](09-vconv.md) | `github.com/imajinyun/go-knifer/vconv` | Permissive type conversion: string, int, int64, float64, bool, bytes, and default-value variants. |
| [`vcron`](10-vcron.md) | `github.com/imajinyun/go-knifer/vcron` | Cron expression parsing and task scheduling, configurable schedulers/options, provider injection, running-task metrics, `Wait`, and graceful `Shutdown(ctx)`. |
| [`vcrypto`](11-vcrypto.md) | `github.com/imajinyun/go-knifer/vcrypto` | Cryptography and digests: SHA-2, HMAC, PBKDF2-SHA256, parameter signing, random bytes, AES-GCM, RSA OAEP/PSS, PEM, and X.509 helpers. |
| [`vcsv`](12-vcsv.md) | `github.com/imajinyun/go-knifer/vcsv` | CSV helpers: reader/writer options, records-to-map conversion, map writing, struct tag export, and record callbacks. |
| [`vdate`](13-vdate.md) | `github.com/imajinyun/go-knifer/vdate` | Date/time helpers: common layouts, parse/format, begin/end of day/month/year, offsets, and comparisons. |
| [`vdb`](14-vdb.md) | `github.com/imajinyun/go-knifer/vdb` | Database helpers built on `database/sql`: SQL execution, named parameters, entities, conditions, query builders, transactions, pagination, metadata lookup, and injectable `sql.Open` providers. |
| [`vdfa`](15-vdfa.md) | `github.com/imajinyun/go-knifer/vdfa` | DFA word-tree matching: stop-rune filtering, first/all matches, dense/greedy modes, found-word metadata, matcher helpers, text replacement, and async initialization providers. |
| [`verr`](16-verr.md) | `github.com/imajinyun/go-knifer/verr` | Error helpers: panic recovery, aggregation, multierror matching, stack capture/formatting, injectable logging/stack/exit/timer/runner providers, and optional logrus/Sentry integration. |
| [`vfile`](17-vfile.md) | `github.com/imajinyun/go-knifer/vfile` | File and IO helpers: read/write/copy, lines, mkdir/touch/delete, filename helpers, quiet close, and provider-backed file-system operations. |
| [`vform`](18-vform.md) | `github.com/imajinyun/go-knifer/vform` | Form and input validation helpers: email, mobile, URL, IPv4/IPv6, ID card, Chinese text, number strings, and matcher providers. |
| [`vftp`](19-vftp.md) | `github.com/imajinyun/go-knifer/vftp` | FTP adapter helpers: provider-injected listing, in-memory download/upload contracts, request validation, transfer limits, and defensive copies. |
| [`vhan`](20-vhan.md) | `github.com/imajinyun/go-knifer/vhan` | Han text adapter helpers: provider-injected Chinese-to-pinyin conversion and initials extraction, request validation, input limits, and defensive copies. |
| [`vhash`](21-vhash.md) | `github.com/imajinyun/go-knifer/vhash` | Non-cryptographic hash helpers: additive, FNV, injectable 32-bit providers, and classic string hashes. |
| [`vhttp`](22-vhttp.md) | `github.com/imajinyun/go-knifer/vhttp` | Standard-library HTTP facade: chainable clients, global/isolated config, explicit-error shortcuts, classified HTTP errors, safe downloads, BasicAuth, HTML helpers, and provider-backed transports/factories. |
| [`vid`](23-vid.md) | `github.com/imajinyun/go-knifer/vid` | ID helpers: UUIDs, ObjectId, Snowflake, worker/datacenter derivation, NanoId generation, fallback random sources, and isolated Snowflake creation. |
| [`vident`](24-vident.md) | `github.com/imajinyun/go-knifer/vident` | Identity helpers: mainland China ID card conversion/validation, birthday/age/gender extraction, province/city/district parsing, masking, and HK/Macau/Taiwan card validation. |
| [`vimg`](25-vimg.md) | `github.com/imajinyun/go-knifer/vimg` | Image helpers: thumbnails, PNG/JPEG/GIF conversion, metadata, QR/barcode generation and decoding, QR logo/background options, and graphical captchas. |
| [`vjob`](26-vjob.md) | `github.com/imajinyun/go-knifer/vjob` | Sliceable job execution with typed adapters, context cancellation, and serialized merge callbacks. |
| [`vjson`](27-vjson.md) | `github.com/imajinyun/go-knifer/vjson` | Ordered JSON objects/arrays, parsing/formatting, path get/put, provider-backed marshal/unmarshal, configurable conversion, and XML/JSON adapters. |
| [`vjwt`](28-vjwt.md) | `github.com/imajinyun/go-knifer/vjwt` | JWT creation, parsing, signing, verification, time-claim validation, HMAC/RSA-PSS/ECDSA support, and unsigned-token rejection. |
| [`vlog`](29-vlog.md) | `github.com/imajinyun/go-knifer/vlog` | Logging facade: console/color loggers, log levels, global logger, static functions, per-call options, and isolated logger creation. |
| [`vmail`](30-vmail.md) | `github.com/imajinyun/go-knifer/vmail` | Mail helpers: RFC 5322 parsing, MIME message construction, text/HTML, inline files, attachments, quick send helpers, context-aware SMTP, mandatory TLS defaults, injection checks, and provider options. |
| [`vmap`](31-vmap.md) | `github.com/imajinyun/go-knifer/vmap` | Map helpers: construction, contains/get/find, sorted keys/values, map/filter/reject/partition, reduce/group/count, inverse, merge, set-style diffs, pick/omit, clone, and equality. |
| [`vmask`](32-vmask.md) | `github.com/imajinyun/go-knifer/vmask` | Masking helpers for names, IDs, phones, addresses, email, passwords, license plates, bank cards, IPs, passports, and credit codes. |
| [`vnet`](33-vnet.md) | `github.com/imajinyun/go-knifer/vnet` | Network helpers: IPv4/IPv6 conversion, CIDR/range/mask utilities, local ports, host/interface/MAC lookup, TLS config, dial/ping options, and multipart forms. |
| [`vnum`](34-vnum.md) | `github.com/imajinyun/go-knifer/vnum` | Numeric helpers: precise arithmetic, generic aggregation, rounding, parsing/formatting providers, random unique numbers, ranges, factorial/combinations, gcd/lcm, binary conversion, byte conversion, and expression calculation. |
| [`vobj`](35-vobj.md) | `github.com/imajinyun/go-knifer/vobj` | Object helpers: nil/empty checks, equality, defaults, clone/serialization, comparison, type inspection, and container utilities. |
| [`vpass`](36-vpass.md) | `github.com/imajinyun/go-knifer/vpass` | Password helpers: deterministic local scoring, strength buckets, strong/weak predicates, repeated/sequential-run detection, and common weak-password blocklist. |
| [`vpoi`](37-vpoi.md) | `github.com/imajinyun/go-knifer/vpoi` | Office document helpers: XLSX sheet listing, row reading/writing, multi-sheet writing, in-memory workbook creation, and injectable workbook/file-system providers. |
| [`vrand`](38-vrand.md) | `github.com/imajinyun/go-knifer/vrand` | Random helpers: integers, floats, booleans, bytes, strings, numeric strings, random element selection, deterministic seeding, and resettable pseudo-random source providers. |
| [`vref`](39-vref.md) | `github.com/imajinyun/go-knifer/vref` | Reflection helpers: field lookup/mutation, method discovery/invocation, constructor-style calls, nil-safe type/value utilities, classifier helpers, and explicit unsafe access options. |
| [`vregex`](40-vregex.md) | `github.com/imajinyun/go-knifer/vregex` | Regular-expression helpers: matching, group extraction, named groups, deletion, counting, index lookup, template/function replacement, escaping, and compiler/DOTALL options. |
| [`vresty`](41-vresty.md) | `github.com/imajinyun/go-knifer/vresty` | Resty v3 HTTP facade: chainable requests, JSON/form/multipart bodies, isolated/global config, request factories, resettable default clients, downloads, safe downloads, and response helpers. |
| [`vsem`](42-vsem.md) | `github.com/imajinyun/go-knifer/vsem` | Weighted, context-aware counting semaphore with FIFO fairness, try-acquire, close notifications, and in-use metrics. |
| [`vset`](43-vset.md) | `github.com/imajinyun/go-knifer/vset` | Generic and typed set utilities with add/remove/contains, set operations, and JSON/YAML encoding helpers. |
| [`vskt`](44-vskt.md) | `github.com/imajinyun/go-knifer/vskt` | TCP socket utilities: plain connections, NIO/AIO server/client helpers, protocol encoder/decoder interfaces, and configurable thread-pool/listener/connection/runner/IP-parser providers. |
| [`vslice`](45-vslice.md) | `github.com/imajinyun/go-knifer/vslice` | Slice helpers: contains/index, reverse, distinct, join, filter/map, sub-slice, concat, set-like operations, and paging. |
| [`vssh`](46-vssh.md) | `github.com/imajinyun/go-knifer/vssh` | SSH/SFTP adapter helpers: provider-injected command execution, SFTP-style listing, in-memory download/upload contracts, output and transfer limits, and defensive copies. |
| [`vstr`](47-vstr.md) | `github.com/imajinyun/go-knifer/vstr` | String/text helpers: blank checks, trimming, splitting, substrings, formatting, emoji helpers, naming conversion, Unicode escaping, Ant matching, text similarity, SimHash, HTML escaping, and rune checks. |
| [`vsys`](48-vsys.md) | `github.com/imajinyun/go-knifer/vsys` | System/runtime information: host, OS, user, Go runtime, process memory, goroutines, environment variables, resettable info cache, and injectable env/command/runtime providers. |
| [`vtok`](49-vtok.md) | `github.com/imajinyun/go-knifer/vtok` | Tokenization adapter helpers: provider-injected text tokenization and keyword extraction, request validation, input/token limits, and defensive copies. |
| [`vtpl`](50-vtpl.md) | `github.com/imajinyun/go-knifer/vtpl` | Template rendering helpers with `html/template`, `text/template`, engine-neutral adapters, context-first rendering, template name, FuncMap, delimiter, factory, and executor options. |
| [`vurl`](51-vurl.md) | `github.com/imajinyun/go-knifer/vurl` | URL/URI helpers: parse, normalize, resolve, query encode/decode, percent encoding providers, URL building, Data URI, scheme checks, file URL conversion, resource open/size helpers, and SSRF-oriented safe variants. |
| [`vver`](52-vver.md) | `github.com/imajinyun/go-knifer/vver` | Version helpers: comparison, greater/less predicates, expression matching, inclusive ranges, and custom expression delimiters. |
| [`vxml`](53-vxml.md) | `github.com/imajinyun/go-knifer/vxml` | XML helpers: parse/read/write/format, tree navigation, XPath-style lookup, escaping, map/bean conversion, transform options, and namespace utilities. |
| [`vzip`](54-vzip.md) | `github.com/imajinyun/go-knifer/vzip` | ZIP, gzip, and zlib helpers: archive creation/extraction, entry lookup, traversal, append, in-memory entries, stream compression, bounded extraction/decompression, traversal checks, and symlink escape checks. |

<a id="quickstart-documents"></a>

## 📝 Quickstart documents

Per-package quickstart examples live in the linked documents above so examples stay focused and easy to maintain by domain.

<a id="sprint-direction"></a>

## 🧭 Sprint direction

The current governance stream prioritizes scenario mindshare, deeper high-value modules, documentation and benchmark trust, and explicit ecosystem adapter lanes for AI, FTP, SSH/SFTP, pinyin, tokenization, multi-template engines, and CLI utilities.

Sprint state is maintained through the generated API snapshots, this documentation hub, local sprint plans under `docs/superpowers/plans/`, and recent commits. The former `49-roadmap.md` page is no longer part of the tracked documentation set; do not recreate it unless roadmap restoration is explicitly requested.

<a id="architecture-and-package-boundaries"></a>

## 🏗️ Architecture and package boundaries

`go-knifer` uses public `v*` packages as facade APIs and keeps concrete code in `internal/*`. Application code should import the `v*` packages; `internal/*` exists so implementations can evolve without exposing every helper as public API.

Facade rules:

- `internal/<domain>` owns implementation details and domain-specific tests.
- `v<domain>` exposes the stable public surface for that domain.
- Small utility packages may use hand-written thin facades; larger modules may keep generated `facade.go` files.
- Newly exported internal APIs should be reviewed before being exposed publicly.
- Facades may keep short names such as `vform`, `vmask`, `vsem`, `vskt`, `vblf`, and `vver`; their meaning is documented in the package catalog instead of changing established import paths.

API compatibility:

- The root package and top-level `v*` subpackages are the public compatibility boundary.
- Their exported API surface is recorded in [`../api/exports.txt`](../api/exports.txt), including function signatures, exported type definitions, struct fields, interface methods, and method sets.
- Standalone `internal/*` packages are intentionally excluded so implementation packages can be refactored without public API noise.
- `make api-check` regenerates a temporary snapshot and compares it with the checked-in file.
- When a public API change is intentional, run `UPDATE_API=1 make api-check` and review the snapshot diff together with the implementation change.
- API additions, removals, and renames should also be reflected in package `doc.go` comments, examples, and the changelog before tagging a release.

<a id="api-compatibility-deprecation-and-releases"></a>

## 🔒 API compatibility, deprecation, and releases

The compatibility policy applies to the public facade layer, not to implementation-only packages. Review [`../api/exports.txt`](../api/exports.txt) with each intentional public API change.

| Stability level | Applies to | Compatibility promise |
| --- | --- | --- |
| Stable | Exported names in `v*` facade packages and `docs/api/exports.txt` | No breaking change without a documented migration path and release note. |
| Internal | `internal/*` implementation packages | May change without public compatibility guarantees. |
| Experimental | Newly introduced provider contracts or adapter packages marked experimental in docs | May change before being promoted to Stable; migration notes are still required. |

A breaking change includes:

- Removing or renaming an exported facade API.
- Changing a public function signature.
- Changing exported type field semantics.
- Changing sentinel error matching behavior.
- Weakening a documented security default.
- Changing generated API snapshot content without release notes.

Deprecated APIs stay available for at least two minor releases. Every deprecation must name the replacement API, explain the migration, and appear in release notes before removal.

Release notes should keep user-visible changes grouped with this structure:

```markdown
## Added
## Changed
## Deprecated
## Removed
## Fixed
## Security
## Migration
```

Configurable APIs and provider injection:

- Many packages expose functional options through `WithXxx` helpers and `XxxWithOptions` variants.
- Configuration-heavy APIs may use explicit option structs such as `vconf.LoadOptions`.
- Existing fixed-argument APIs stay stable, while option-based variants add advanced control for callers that need it.
- Provider-style options let callers inject file-system functions, network/TLS dialers or readers, HTTP request/multipart factories, clocks, timers/tickers, random sources, DB openers, Excel workbook factories, loggers, stack capture functions, finalizers, environment lookups, command executors, Sentry/logrus hooks, and other process-global dependencies for deterministic tests and controlled runtime behavior.
- Package-level defaults remain explicit. For example, HTTP global defaults can be read as immutable snapshots, and isolated request constructors can build requests without reading package-level defaults.
- Configuration objects are mutable while being built or loaded, then should be treated as read-only snapshots after publication. Use clone helpers before applying runtime changes and publish the new pointer atomically instead of mutating a shared instance in place.

Domain boundary rules:

- `vhash` owns non-cryptographic hash helpers; `vcrypto` owns security-oriented digests, HMAC, encryption, and key/PEM operations.
- `vhttp` is the lightweight standard-library HTTP facade; `vresty` is the Resty-based chainable client facade.
- URL escaping, query building/parsing, and scheme checks live in `vurl` instead of being re-exported by HTTP facades.
- `vdb` owns SQL database helpers on top of `database/sql`; callers keep control of drivers and connection pools through `*sql.DB` and per-call options.
- `vdfa` owns DFA word-tree matching and text replacement; generic string helpers should not absorb dictionary-matching logic.
- `vid` owns generated identifiers; `vident` owns legal identity numbers and regional card parsing.
- `vcodec` owns encoding/decoding algorithms; `vurl` owns URL/URI parsing, normalization, resource open/size checks, and scheme semantics.
- `vjson` owns JSON objects, arrays, paths, and lightweight XML adapters; `vxml` owns XML parsing, tree navigation, formatting, namespace handling, and XML-specific map/bean conversion.
- `vbean` owns direct struct/map property mapping without serializing through JSON.
- `vobj` is a convenience object-level facade. New domain logic should still be implemented first in clear packages such as `vstr`, `vslice`, `vmap`, or `vref`, then wrapped by `vobj` only when useful.

### 🚦 Error contract

The root package `knifer` owns the cross-cutting error contract: the `ErrCode` classifier, unified `knifer.Error` type, `CodeCarrier` interface, `CodeOf` extractor, and `NewError` / `WrapError` / `Errorf` constructors.

Subpackages that opt in return `*knifer.Error` or add code-aware matching to their existing error types/sentinels, so callers can match or extract by code while keeping the cause chain:

```go
if errors.Is(err, knifer.ErrCodeInvalidInput) { /* ... */ }
if code, ok := knifer.CodeOf(err); ok { /* ... */ }
```

`vcrypto` is a reference integration: validation errors match both `knifer.ErrCodeInvalidInput` and existing `vcrypto` sentinels such as `ErrInvalidKey`, `ErrInvalidIV`, and `ErrInvalidCipherText`.

### 🔐 Security and safety defaults

Security-sensitive helpers expose only the currently recommended public API surface:

- `vcrypto` keeps SHA-2 digests, HMAC-SHA-256/384/512, PBKDF2-SHA-256, AES-GCM, RSA-OAEP encryption, and RSA-PSS signatures.
- JWT RSA signing is exposed through RSA-PSS alongside HMAC and ECDSA signers; unsigned JWT `alg=none` tokens are rejected.
- TLS helpers create configs with TLS 1.2+ as the minimum version; convenience APIs do not bypass certificate verification.
- HTTP and Resty downloads validate automatically discovered filenames before joining them under destination directories.
- Safe HTTP/URL helpers reject local/private/link-local/unspecified targets by default and re-check redirect targets.
- `vfile`, `vconf`, `vurl`, and `vzip` use bounded reads or extraction/decompression limits by default.
- ZIP extraction cleans entry names, checks traversal, and resolves destination parents to reject symlink escape.
- Bloom filter constructors ending in `E` return validation errors for invalid sizes or hash configuration instead of panicking.
- `vdb` condition builders validate operators against an allowlist.
- `vskt.AioSession` serializes reads that share the session buffer and keeps buffers available during close callbacks.
- `vftp` does not open network connections, read credentials, touch local filesystem paths, or log transfer data; callers inject providers and enforce deployment-specific FTP security at the boundary.
- `vssh` does not open network connections, execute shell commands, read credentials, parse keys, touch local filesystem paths, or log command output or transfer data; callers inject providers and enforce SSH/SFTP security at the boundary.
- `vhan` does not import dictionaries, tokenize text, open network connections, read credentials, touch local filesystem paths, or log input text; callers inject providers and own dictionary/polyphone behavior at the boundary.
- `vtok` does not import dictionaries, segment text, rank keywords, open network connections, read credentials, touch local filesystem paths, or log input text; callers inject providers and own tokenizer/ranking behavior at the boundary.

<a id="recommended-api-entry-points"></a>

## ✅ Recommended API entry points

Use these APIs for new code. Request helpers that can fail return errors explicitly instead of swallowing failures.

| Scenario | Recommended API |
| --- | --- |
| Build a trusted standard-library HTTP request | `vhttp.Get`, `vhttp.Post`, `vhttp.NewRequest` |
| Read a trusted HTTP response body and handle errors | `vhttp.GetStringE`, `vhttp.PostJSONE`, `vhttp.DownloadBytesE` |
| Access user-controlled or otherwise untrusted HTTP(S) URLs | `vhttp.GetStringSafeE`, `vhttp.PostJSONSafeE`, `vhttp.DownloadBytesSafeE`, `vurl.OpenSafe` |
| Use the Resty-backed HTTP facade | `vresty.Get`, `vresty.Post`, `vresty.GetStringE`, `vresty.PostJSONE` |
| Access untrusted URLs through Resty | `vresty.GetStringSafeE`, `vresty.PostJSONSafeE`, `vresty.DownloadBytesSafeE` |
| Download a user-controlled URL to a file | `vhttp.DownloadFileSafe` or `vresty.DownloadFileSafe` |
| Generate bytes for secrets, tokens, keys, nonces, or salts | `vrand.SecureBytes` |
| Create an LRU cache | `vcache.NewLRU` or `vcache.NewLRUWithTimeout` |
| Parse a cron expression | `vcron.NewPattern` or `vcron.MustNewPattern` |
| Load remote configuration from a trust boundary | `vconf.LoadRemoteSafe` or `vconf.LoadRemoteSafeWithOptions` |
| Use a provider-injected FTP contract without adding a network client dependency | `vftp.New`, `vftp.List`, `vftp.Download`, `vftp.Upload` |
| Use a provider-injected SSH/SFTP contract without adding a network client dependency | `vssh.New`, `vssh.Run`, `vssh.List`, `vssh.Download`, `vssh.Upload` |

<a id="build-test-and-release-workflow"></a>

## 📦 Build, test, and release workflow

Clone the source code:

```bash
git clone https://github.com/imajinyun/go-knifer.git
cd go-knifer
```

Run tests:

```bash
make test
```

Diagnose the local Go/tooling/Git environment without modifying files:

```bash
make doctor
```

Check that unrelated untracked Go files cannot pollute tests or commits:

```bash
make worktree-check
```

Optionally install local Git hooks so `make quick-check` runs before commit and `make full-check COVERAGE_FILE=/tmp/go-knifer-coverage.out` runs before push:

```bash
make install-hooks
```

Run the CI test-job gates locally. This verifies modules, vet, tidy/diff cleanliness, architecture rules, race/shuffle tests, coverage gates, and the exported API snapshot:

```bash
make ci-test
```

Validate the machine-readable AI metadata used by agents and CLI automation:

```bash
make ai-context-check
```

Run the same local safety checks used by CI before opening a PR:

```bash
make check
```

`make check` includes the `ci-test` class of checks plus `golangci-lint` and `govulncheck`.

Benchmark baselines:

```bash
make bench-smoke
make bench-core
make bench-facade
make bench-codec
make bench-core BENCHCOUNT=10 BENCHTIME=3s
```

Use `make bench-smoke` to verify benchmark health quickly after changing hot paths. Use `make bench-core`, `make bench-facade`, and `make bench-codec` for stable package groups. Treat single-run benchmark output as a baseline only; use repeated runs and `benchstat` before documenting an improvement or regression.

Refresh the API snapshot after an intentional exported API change:

```bash
UPDATE_API=1 make api-check
```

Refresh generated documentation artifacts after intentional facade, doc comment, or Example changes:

```bash
make docs-gen
make docs-check
```

Run repository `go:generate` directives after confirming generated output is expected:

```bash
make generate
```

GitHub Actions reuses the Makefile targets for module verification, vet, tidy checks, diff cleanliness, architecture checks, race/shuffle tests, coverage gates, and API compatibility checks. It also runs `golangci-lint`, `govulncheck`, and CodeQL. Dependabot is configured for Go modules and GitHub Actions updates.

Format code:

```bash
gofmt -w .
```

<a id="governance"></a>

## 🛡️ Governance

- Security reports: see [`../../SECURITY.md`](../../SECURITY.md). Please do not disclose suspected vulnerabilities in public issues.
- Release notes: see [`../../CHANGELOG.md`](../../CHANGELOG.md). User-visible changes should be recorded before tagging a release.
- Coverage gate: CI enforces the repository baseline with `bash bin/check_coverage.sh coverage.out`. Raise `COVERAGE_THRESHOLD` or `PACKAGE_COVERAGE_THRESHOLDS` only after adding tests that support the new gate.
- API gate: `make api-check` compares root-package and top-level `v*` API signatures, exported fields, interface methods, and method sets against [`../api/exports.txt`](../api/exports.txt). Commit the refreshed snapshot only for intentional public API changes.
- Generated documentation gate: `make docs-check` verifies generated documentation artifacts, including the machine-readable tool catalog at [`../api/tools.json`](../api/tools.json) and the readable catalog at [`../api/tools.md`](../api/tools.md). Regenerate with `make docs-gen` only when source docs, facade functions, or Examples intentionally change.
- AI metadata gate: `make ai-context-check` validates [`../../ai-context.json`](../../ai-context.json), including command side effects, facade inventory, coverage gates, and security-sensitive package references.
- Workflow gates: use `make doctor` for environment diagnostics, `make worktree-check` to block unrelated untracked Go files, `make quick-check` for fast local validation, `make security-check` for lint and vulnerability scanning, `make full-check COVERAGE_FILE=/tmp/go-knifer-coverage.out` for the full pre-push gate, and `make ci-test` for the GitHub Actions test-job gate. Optional Git hooks can be enabled with `make install-hooks` and disabled with `make uninstall-hooks`.
- Security suppressions: keep `.golangci.yml`, `#nosec`, and `//nolint:gosec` exceptions narrow and justified at the call site; prefer a regression test before broadening an exclusion.
- Benchmark baseline: use `make bench-core` for hot-path benchmark suites or `make bench-facade` for matching public facade packages. Treat the output as a baseline unless a separate `benchstat` comparison proves a performance change.

<a id="contributing"></a>

## 🤝 Contributing

If you find a bug or want to request a new utility, please open a GitHub Issue. It is recommended to include:

- Go version and operating system;
- `go-knifer` version or commit;
- Minimal reproducible code;
- Expected behavior and actual behavior;
- Related error logs or test output.

Pull requests are welcome. To keep the toolkit stable, please follow these principles where possible:

1. Add new capabilities to the appropriate `internal/*` implementation package first, then expose public APIs from the corresponding `v*` package;
2. Add necessary comments for new or modified public APIs;
3. Add unit tests for core logic and run `go test ./...` before submitting;
4. Keep code formatted with `gofmt`;
5. Avoid unnecessary third-party dependencies and prefer the standard library when possible.
