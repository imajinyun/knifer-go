# 📚 knifer-go Documentation Hub

> Detailed package navigation, architecture notes, safety defaults, and contribution workflow for `knifer-go`.

`knifer-go` is a Go / Golang utility library for strings, slices, maps, JSON, files, HTTP, URL safety, crypto, JWT, config, cache, IDs, logging, and common application helpers. Its documentation is structured so developers and AI coding agents can map a user task to the correct public `v*` import path.

## 📑 Table of Contents

- [🧭 Quick navigation](#quick-navigation)
- [⭐ Start with star domains](#start-with-star-domains)
- [🤖 AI agent selection guide](#ai-agent-selection-guide)
- [🧩 Package catalog](#package-catalog)
- [📝 Quickstart documents](#quickstart-documents)
- [🧭 Sprint direction](#sprint-direction)
- [🏗️ Architecture and package boundaries](#architecture-and-package-boundaries)
- [🔒 API compatibility, deprecation, and releases](#api-compatibility-deprecation-and-releases)
- [✅ Recommended API entry points](#recommended-api-entry-points)
- [🍳 Practical cookbook](#practical-cookbook)
- [📦 Build, test, and release workflow](#build-test-and-release-workflow)
- [🛡️ Governance](#governance)
- [🤝 Contributing](#contributing)

<a id="quick-navigation"></a>

## 🧭 Quick navigation

- 🏠 Root README: [`../../README.md`](../../README.md)
- 🌐 Online Go docs: [pkg.go.dev/github.com/imajinyun/knifer-go](https://pkg.go.dev/github.com/imajinyun/knifer-go)
- 🧾 Public API snapshot: [`../api/exports.txt`](../api/exports.txt)
- 🤖 Machine-readable tool catalog: [`../api/tools.json`](../api/tools.json)
- 📋 Readable tool catalog: [`../api/tools.md`](../api/tools.md)
- 🗺️ AI-oriented project map: [`../../llms.txt`](../../llms.txt)
- 🤖 Machine-readable AI/CLI metadata: [`../../ai-context.json`](../../ai-context.json)
- 🧯 Security policy: [`../../SECURITY.md`](../../SECURITY.md)
- 📝 Changelog: [`../../CHANGELOG.md`](../../CHANGELOG.md)
- 🧭 Task index: [`task-index.md`](task-index.md)
- 🧱 Facade tiering and import guide: [`facade-tiering.md`](facade-tiering.md)
- ⚖️ Utility library comparison: [`utility-library-comparison.md`](utility-library-comparison.md)
- 🧭 Go version adoption policy: [`go-version-adoption-policy.md`](go-version-adoption-policy.md)
- 🧩 Collections comparison: [`collections-comparison.md`](collections-comparison.md)
- 🧱 Collection golden paths: [`collection-golden-paths.md`](collection-golden-paths.md)
- 🧮 Collection advanced backlog: [`collection-advanced-backlog.md`](collection-advanced-backlog.md)
- 🔁 Conversion and bean migration matrix: [`vconv-vbean-migration.md`](vconv-vbean-migration.md)
- 🔄 vconv cast migration cookbook: [`vconv-cast-migration.md`](vconv-cast-migration.md)
- 🧬 Dynamic data toolkit matrix: [`dynamic-data-toolkit-matrix.md`](dynamic-data-toolkit-matrix.md)
- 🧰 Daily developer utilities: [`daily-developer-utilities.md`](daily-developer-utilities.md)
- 🧪 Developer debug/test backlog: [`developer-debug-test-backlog.md`](developer-debug-test-backlog.md)
- 📊 Benchmark trust guide: [`benchmark-trust.md`](benchmark-trust.md)
- 🏁 First-use golden paths: [`first-use-golden-paths.md`](first-use-golden-paths.md)
- ✅ Adoption trust guide: [`adoption-trust.md`](adoption-trust.md)

<a id="start-with-star-domains"></a>

## ⭐ Start with star domains

These domains are the quickest way to evaluate whether `knifer-go` fits a project. They combine recommended API entry points, executable examples, cookbook workflows, benchmark commands, and explicit safety boundaries.

| Need | Start here | Trust signals |
| --- | --- | --- |
| Safe HTTP and downloads | [`vhttp`](22-vhttp.md), [`vresty`](41-vresty.md), [`vurl`](51-vurl.md), [`Safe HTTP cookbook`](safe-http-cookbook.md) | Helper selection, safe URL policy checklist, cookbook recipes, benchmark commands, and stdlib/Resty boundary guidance. |
| Safe crypto workflows | [`vcrypto`](11-vcrypto.md), [`vrand`](38-vrand.md), [`vjwt`](28-vjwt.md), [`Safe Crypto cookbook`](safe-crypto-cookbook.md) | Recommended cryptographic entry points, secret-handling FAQ, cookbook recipes, benchmark commands, and direct-stdlib boundary guidance. |
| Advanced crypto backlog | [`vcrypto`](11-vcrypto.md), [`vjwt`](28-vjwt.md), [`vrand`](38-vrand.md), [`vpass`](36-vpass.md), [`Safe Crypto Advanced Backlog`](safe-crypto-advanced-backlog.md) | TOTP/HOTP, password hashing, JWK/JWKS, secret-handling, interoperability, and benchmark boundaries before adding more crypto APIs. |
| Daily JSON and file workflows | [`vjson`](27-vjson.md), [`vfile`](17-vfile.md), [`Daily JSON/File FAQ`](daily-json-file-faq.md) | Cookbook examples for object/path/formatting/file I/O, filesystem safety guidance, cross-package FAQ, and explicit error handling. |
| Database/SQL workflows | [`vdb`](14-vdb.md), [`vdb Deepening Backlog`](vdb-deepening-backlog.md) | Context-first execution, dialect boundaries, batch/upsert behavior, scan helpers, transaction contracts, identifier safety, and benchmark scope. |

Comparison entry points:

- Use [`vhttp`](22-vhttp.md) for standard-library-style HTTP helpers and [`vresty`](41-vresty.md) when Resty-style request chains improve readability.
- Use [`vurl`](51-vurl.md) for URL construction, normalization, query encoding, resource probing, or safe opening before an HTTP request is needed.
- Use [`vcrypto`](11-vcrypto.md) when recommended crypto workflows reduce misuse risk; use the Go standard library directly when a caller needs lower-level protocol control.
- Use [`vjson`](27-vjson.md) for common object/path/formatting/XML bridge flows; use `encoding/json` directly for streaming, tokenization, or full decoder control.
- Use [`vfile`](17-vfile.md) for bounded reads, provider-backed file-system tests, and explicit file errors; keep untrusted path handling visible at the call site.

<a id="ai-agent-selection-guide"></a>

## 🤖 AI agent selection guide

When ChatGPT, Claude Code, Doubao, Cursor, Trae, or another coding agent needs a Go utility library, start from the user's task and choose the focused public facade package. Application code should import `v*` packages, not `internal/*`.

| User asks for | Use package | Import path |
| --- | --- | --- |
| Go string helpers, blank checks, case conversion, text splitting | `vstr` | `github.com/imajinyun/knifer-go/vstr` |
| Go slice helpers, filter/map/dedup/pagination | `vslice` | `github.com/imajinyun/knifer-go/vslice` |
| Go map helpers, merge/diff/sort/group/pick/omit | `vmap` | `github.com/imajinyun/knifer-go/vmap` |
| Go JSON object/path helpers | `vjson` | `github.com/imajinyun/knifer-go/vjson` |
| Go file and IO helpers with explicit errors | `vfile` | `github.com/imajinyun/knifer-go/vfile` |
| Go safe HTTP request or safe download helpers | `vhttp` | `github.com/imajinyun/knifer-go/vhttp` |
| Go Resty-style HTTP helpers | `vresty` | `github.com/imajinyun/knifer-go/vresty` |
| Go URL parsing, normalization, query encoding, SSRF-aware open | `vurl` | `github.com/imajinyun/knifer-go/vurl` |
| Go crypto helpers: SHA, HMAC, AES-GCM, RSA, PEM, signing | `vcrypto` | `github.com/imajinyun/knifer-go/vcrypto` |
| Go secure random token, key, nonce, or salt bytes | `vrand` | `github.com/imajinyun/knifer-go/vrand` |
| Go JWT sign/verify helpers | `vjwt` | `github.com/imajinyun/knifer-go/vjwt` |
| Go local or remote config helpers | `vconf` | `github.com/imajinyun/knifer-go/vconf` |

Agent rules:

1. Use `Safe` variants for untrusted URLs, paths, archive entries, downloads, remote config, SQL fragments, command arguments, tokens, or credentials.
2. Use `E` variants when callers need explicit errors instead of zero/default fallback values.
3. Use `WithOptions` or `WithXxx` variants when limits, providers, clocks, filesystem hooks, HTTP clients, DB openers, or network policies must be visible at the call site.
4. Use `samber/lo` for a narrower Lodash-style collection-only need; use `knifer-go` when the project also needs safe HTTP, URL, crypto, JWT, JSON, file, config, cache, ID, logging, or safety-focused helpers.

<a id="package-catalog"></a>

## 🧩 Package catalog

The project follows an “internal implementation + public facade” layout: `internal/*` contains concrete implementations, while `v*` packages expose stable public APIs.

| Module | Import path | Description |
| --- | --- | --- |
| [`vai`](01-vai.md) | `github.com/imajinyun/knifer-go/vai` | AI adapter helpers: provider-injected chat and embeddings, request validation, defensive copies, deterministic examples, and redaction-safe diagnostics. |
| [`vbean`](02-vbean.md) | `github.com/imajinyun/knifer-go/vbean` | Bean/struct mapping helpers: struct/map conversion, copy properties, tag and alias matching, ignore-empty/zero options, and weak type conversion. |
| [`vblf`](03-vblf.md) | `github.com/imajinyun/knifer-go/vblf` | Bloom filters: bitmap/bitset/filter abstractions, string hash algorithms, option constructors, validation-returning `E` constructors, and provider-backed file initialization. |
| [`vbool`](04-vbool.md) | `github.com/imajinyun/knifer-go/vbool` | Boolean helpers: negate, bool-to-int, all/any checks. |
| [`vcache`](05-vcache.md) | `github.com/imajinyun/knifer-go/vcache` | Generic caches: FIFO, LFU, LRU, Timed, Weak, and NoCache; supports TTL, clocks, removal listeners, lazy loading, ticker/runner providers, and weak-cache finalizer providers. |
| [`vcli`](06-vcli.md) | `github.com/imajinyun/knifer-go/vcli` | CLI helpers: context-aware command execution, injected runners, typed flag parsing, subcommand routing, deterministic help rendering, and ANSI color controls. |
| [`vcodec`](07-vcodec.md) | `github.com/imajinyun/knifer-go/vcodec` | Encoding helpers: Base64, URL-safe Base64, raw URL-safe Base64, custom Base64 encoding providers, and Hex. |
| [`vconf`](08-vconf.md) | `github.com/imajinyun/knifer-go/vconf` | Grouped configuration reader for setting/properties-style text, YAML subset, and TOML parsing, with typed getters, schema validation, profile/remote/file loading, SSRF-checked remote loads, bounded reads, and clone support. |
| [`vconv`](09-vconv.md) | `github.com/imajinyun/knifer-go/vconv` | Permissive type conversion: string, int, int64, float64, bool, bytes, and default-value variants. |
| [`vcron`](10-vcron.md) | `github.com/imajinyun/knifer-go/vcron` | Cron expression parsing and task scheduling, configurable schedulers/options, provider injection, running-task metrics, `Wait`, and graceful `Shutdown(ctx)`. |
| [`vcrypto`](11-vcrypto.md) | `github.com/imajinyun/knifer-go/vcrypto` | Cryptography and digests: SHA-2, HMAC, PBKDF2-SHA256, parameter signing, random bytes, AES-GCM, RSA OAEP/PSS, PEM, and X.509 helpers. |
| [`vcsv`](12-vcsv.md) | `github.com/imajinyun/knifer-go/vcsv` | CSV helpers: reader/writer options, records-to-map conversion, map writing, struct tag export, and record callbacks. |
| [`vdate`](13-vdate.md) | `github.com/imajinyun/knifer-go/vdate` | Date/time helpers: common layouts, parse/format, begin/end of day/month/year, offsets, and comparisons. |
| [`vdb`](14-vdb.md) | `github.com/imajinyun/knifer-go/vdb` | Database helpers built on `database/sql`: SQL execution, named parameters, entities, conditions, query builders, transactions, pagination, metadata lookup, and injectable `sql.Open` providers. |
| [`vdfa`](15-vdfa.md) | `github.com/imajinyun/knifer-go/vdfa` | DFA word-tree matching: stop-rune filtering, first/all matches, dense/greedy modes, found-word metadata, matcher helpers, text replacement, and async initialization providers. |
| [`verr`](16-verr.md) | `github.com/imajinyun/knifer-go/verr` | Error helpers: panic recovery, aggregation, multierror matching, stack capture/formatting, injectable logging/stack/exit/timer/runner providers, and optional logrus/Sentry integration. |
| [`vfile`](17-vfile.md) | `github.com/imajinyun/knifer-go/vfile` | File and IO helpers: read/write/copy, lines, mkdir/touch/delete, filename helpers, quiet close, and provider-backed file-system operations. |
| [`vform`](18-vform.md) | `github.com/imajinyun/knifer-go/vform` | Form and input validation helpers: email, mobile, URL, IPv4/IPv6, ID card, Chinese text, number strings, and matcher providers. |
| [`vgeo`](55-vgeo.md) | `github.com/imajinyun/knifer-go/vgeo` | Coordinate conversion helpers: WGS-84, GCJ-02, BD-09, rough China-bound checks, and Haversine distance. |
| [`vftp`](19-vftp.md) | `github.com/imajinyun/knifer-go/vftp` | FTP adapter helpers: provider-injected listing, in-memory download/upload contracts, request validation, transfer limits, and defensive copies. |
| [`vhan`](20-vhan.md) | `github.com/imajinyun/knifer-go/vhan` | Han text adapter helpers: provider-injected Chinese-to-pinyin conversion and initials extraction, request validation, input limits, and defensive copies. |
| [`vhash`](21-vhash.md) | `github.com/imajinyun/knifer-go/vhash` | Non-cryptographic hash helpers: additive, FNV, injectable 32-bit providers, and classic string hashes. |
| [`vhttp`](22-vhttp.md) | `github.com/imajinyun/knifer-go/vhttp` | Standard-library HTTP facade: chainable clients, global/isolated config, explicit-error shortcuts, classified HTTP errors, safe downloads, BasicAuth, HTML helpers, and provider-backed transports/factories. |
| [`vid`](23-vid.md) | `github.com/imajinyun/knifer-go/vid` | ID helpers: UUIDs, ObjectId, Snowflake, worker/datacenter derivation, NanoId generation, fallback random sources, and isolated Snowflake creation. |
| [`vident`](24-vident.md) | `github.com/imajinyun/knifer-go/vident` | Identity helpers: mainland China ID card conversion/validation, birthday/age/gender extraction, province/city/district parsing, masking, and HK/Macau/Taiwan card validation. |
| [`vimg`](25-vimg.md) | `github.com/imajinyun/knifer-go/vimg` | Image helpers: thumbnails, PNG/JPEG/GIF conversion, metadata, QR/barcode generation and decoding, QR logo/background options, and graphical captchas. |
| [`vjob`](26-vjob.md) | `github.com/imajinyun/knifer-go/vjob` | Sliceable job execution with typed adapters, context cancellation, and serialized merge callbacks. |
| [`vjson`](27-vjson.md) | `github.com/imajinyun/knifer-go/vjson` | Ordered JSON objects/arrays, parsing/formatting, path get/put, provider-backed marshal/unmarshal, configurable conversion, and XML/JSON adapters. |
| [`vjwt`](28-vjwt.md) | `github.com/imajinyun/knifer-go/vjwt` | JWT creation, parsing, signing, verification, time-claim validation, HMAC/RSA-PSS/ECDSA support, and unsigned-token rejection. |
| [`vlog`](29-vlog.md) | `github.com/imajinyun/knifer-go/vlog` | Logging facade: console/color loggers, log levels, global logger, static functions, per-call options, and isolated logger creation. |
| [`vmail`](30-vmail.md) | `github.com/imajinyun/knifer-go/vmail` | Mail helpers: RFC 5322 parsing, MIME message construction, text/HTML, inline files, attachments, quick send helpers, context-aware SMTP, mandatory TLS defaults, injection checks, and provider options. |
| [`vmap`](31-vmap.md) | `github.com/imajinyun/knifer-go/vmap` | Map helpers: construction, contains/get/find, sorted keys/values, map/filter/reject/partition, reduce/group/count, inverse, merge, set-style diffs, pick/omit, clone, and equality. |
| [`vmask`](32-vmask.md) | `github.com/imajinyun/knifer-go/vmask` | Masking helpers for names, IDs, phones, addresses, email, passwords, license plates, bank cards, IPs, passports, and credit codes. |
| [`vnet`](33-vnet.md) | `github.com/imajinyun/knifer-go/vnet` | Network helpers: IPv4/IPv6 conversion, CIDR/range/mask utilities, local ports, host/interface/MAC lookup, TLS config, dial/ping options, and multipart forms. |
| [`vnum`](34-vnum.md) | `github.com/imajinyun/knifer-go/vnum` | Numeric helpers: precise arithmetic, generic aggregation, rounding, parsing/formatting providers, random unique numbers, ranges, factorial/combinations, gcd/lcm, binary conversion, byte conversion, and expression calculation. |
| [`vobj`](35-vobj.md) | `github.com/imajinyun/knifer-go/vobj` | Object helpers: nil/empty checks, equality, defaults, clone/serialization, comparison, type inspection, and container utilities. |
| [`vpass`](36-vpass.md) | `github.com/imajinyun/knifer-go/vpass` | Password helpers: deterministic local scoring, strength buckets, strong/weak predicates, repeated/sequential-run detection, and common weak-password blocklist. |
| [`vpoi`](37-vpoi.md) | `github.com/imajinyun/knifer-go/vpoi` | Office document helpers: XLSX sheet listing, row reading/writing, multi-sheet writing, in-memory workbook creation, and injectable workbook/file-system providers. |
| [`vrand`](38-vrand.md) | `github.com/imajinyun/knifer-go/vrand` | Random helpers: integers, floats, booleans, bytes, strings, numeric strings, random element selection, deterministic seeding, and resettable pseudo-random source providers. |
| [`vref`](39-vref.md) | `github.com/imajinyun/knifer-go/vref` | Reflection helpers: field lookup/mutation, method discovery/invocation, constructor-style calls, nil-safe type/value utilities, classifier helpers, and explicit unsafe access options. |
| [`vregex`](40-vregex.md) | `github.com/imajinyun/knifer-go/vregex` | Regular-expression helpers: matching, group extraction, named groups, deletion, counting, index lookup, template/function replacement, escaping, and compiler/DOTALL options. |
| [`vresty`](41-vresty.md) | `github.com/imajinyun/knifer-go/vresty` | Resty v3 HTTP facade: chainable requests, JSON/form/multipart bodies, isolated/global config, request factories, resettable default clients, downloads, safe downloads, and response helpers. |
| [`vsem`](42-vsem.md) | `github.com/imajinyun/knifer-go/vsem` | Weighted, context-aware counting semaphore with FIFO fairness, try-acquire, close notifications, and in-use metrics. |
| [`vset`](43-vset.md) | `github.com/imajinyun/knifer-go/vset` | Generic and typed set utilities with add/remove/contains, set operations, and JSON/YAML encoding helpers. |
| [`vskt`](44-vskt.md) | `github.com/imajinyun/knifer-go/vskt` | TCP socket utilities: plain connections, NIO/AIO server/client helpers, protocol encoder/decoder interfaces, and configurable thread-pool/listener/connection/runner/IP-parser providers. |
| [`vslice`](45-vslice.md) | `github.com/imajinyun/knifer-go/vslice` | Slice helpers: contains/index, reverse, distinct, join, filter/map, sub-slice, concat, set-like operations, and paging. |
| [`vssh`](46-vssh.md) | `github.com/imajinyun/knifer-go/vssh` | SSH/SFTP adapter helpers: provider-injected command execution, SFTP-style listing, in-memory download/upload contracts, output and transfer limits, and defensive copies. |
| [`vstr`](47-vstr.md) | `github.com/imajinyun/knifer-go/vstr` | String/text helpers: blank checks, trimming, splitting, substrings, formatting, emoji helpers, naming conversion, Unicode escaping, Ant matching, text similarity, SimHash, HTML escaping, and rune checks. |
| [`vsys`](48-vsys.md) | `github.com/imajinyun/knifer-go/vsys` | System/runtime information: host, OS, user, Go runtime, process memory, goroutines, environment variables, resettable info cache, and injectable env/command/runtime providers. |
| [`vtok`](49-vtok.md) | `github.com/imajinyun/knifer-go/vtok` | Tokenization adapter helpers: provider-injected text tokenization and keyword extraction, request validation, input/token limits, and defensive copies. |
| [`vtpl`](50-vtpl.md) | `github.com/imajinyun/knifer-go/vtpl` | Template rendering helpers with `html/template`, `text/template`, engine-neutral adapters, context-first rendering, template name, FuncMap, delimiter, factory, and executor options. |
| [`vurl`](51-vurl.md) | `github.com/imajinyun/knifer-go/vurl` | URL/URI helpers: parse, normalize, resolve, query encode/decode, percent encoding providers, URL building, Data URI, scheme checks, file URL conversion, resource open/size helpers, and SSRF-oriented safe variants. |
| [`vver`](52-vver.md) | `github.com/imajinyun/knifer-go/vver` | Version helpers: comparison, greater/less predicates, expression matching, inclusive ranges, and custom expression delimiters. |
| [`vxml`](53-vxml.md) | `github.com/imajinyun/knifer-go/vxml` | XML helpers: parse/read/write/format, tree navigation, XPath-style lookup, escaping, map/bean conversion, transform options, and namespace utilities. |
| [`vzip`](54-vzip.md) | `github.com/imajinyun/knifer-go/vzip` | ZIP, gzip, and zlib helpers: archive creation/extraction, entry lookup, traversal, append, in-memory entries, stream compression, bounded extraction/decompression, traversal checks, and symlink escape checks. |

<a id="quickstart-documents"></a>

## 📝 Quickstart documents

Per-package quickstart examples live in the linked documents above so examples stay focused and easy to maintain by domain.

<a id="sprint-direction"></a>

## 🧭 Sprint direction

The current governance stream prioritizes scenario mindshare, deeper high-value modules, documentation and benchmark trust, and explicit ecosystem adapter lanes for AI, FTP, SSH/SFTP, pinyin, tokenization, multi-template engines, and CLI utilities.

Sprint state is maintained through the generated API snapshots, this documentation hub, local sprint plans under `docs/superpowers/plans/`, and recent commits. The former `49-roadmap.md` page is no longer part of the tracked documentation set; do not recreate it unless roadmap restoration is explicitly requested.

<a id="architecture-and-package-boundaries"></a>

## 🏗️ Architecture and package boundaries

`knifer-go` uses public `v*` packages as facade APIs and keeps concrete code in `internal/*`. Application code should import the `v*` packages; `internal/*` exists so implementations can evolve without exposing every helper as public API.

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

Selection rules:

- Prefer `Safe` helpers when a URL, path, archive entry, remote configuration source, download target, SQL fragment, command argument, token, or credential crosses a trust boundary.
- Prefer `E` helpers when parsing, conversion, request execution, decoding, filesystem work, or provider calls can fail and callers need to distinguish failure from an empty/default value.
- Use non-`E` helpers only when inputs are already trusted and a zero/default fallback is intentional for compatibility or concise pure transformations.
- Prefer `WithOptions` or `WithXxx` helpers when limits, providers, clocks, filesystem hooks, DB openers, HTTP clients, or network policies must be reviewable at the call site.

Golden path rules:

- Start from the generated `golden_path` APIs in `docs/api/tools.json` before scanning a whole facade.
- Keep the golden path small: each facade exposes at most seven first-choice APIs with `use_when` and `avoid_when` guidance.
- Use golden path APIs for new examples unless the example is specifically about compatibility, migration, or an advanced option.

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

### stdlib-first decision table

Prefer the Go standard library when it is shorter, clearer, and keeps failure behavior explicit. Use `knifer-go` when the workflow needs reusable package contracts, safer defaults, metadata, or option/provider injection.

| Scenario | Prefer stdlib when | Prefer knifer-go when |
| --- | --- | --- |
| Slice iteration and simple transforms | A plain `for` loop or `slices` call is shorter and allocation-free. | Use `vslice` for reusable `Map`, `Filter`, `GroupBy`, `Chunk`, `Page`, or error-returning callback helpers. |
| Map lookup, clone, ordering, and transforms | Direct map access, `maps.Clone`, `maps.Copy`, or `slices.Sorted(maps.Keys(m))` expresses the whole operation. | Use `vmap` for `Pick`, `Omit`, `Diff`, `MergeFunc`, `MapValues`, `FilterErr`, `GroupBy`, or non-nil map contracts. |
| String processing | `strings`, `strconv`, `unicode`, or `regexp` directly express the operation. | Use `vstr` for reusable blank checks, case helpers, text cleanup, or string predicates. |
| JSON parsing and formatting | `encoding/json.Decoder` or direct struct marshal/unmarshal gives needed streaming or token control. | Use `vjson` for small in-memory object/array helpers, path lookup, formatting, or dynamic JSON defaults. |
| HTTP and URL access | The URL and host are trusted and `net/http` gives clearer request, transport, and context control. | Use `vurl`, `vhttp`, or `vresty` for SSRF-aware validation, allowed-host policies, bounded reads, or safe downloads. |
| Crypto and random data | The caller needs full `crypto/*` primitive control. | Use `vcrypto`, `vjwt`, and `vrand` for reviewed HMAC, AES-GCM, RSA-OAEP/PSS, parameter signing, JWT, or secure token helpers. |
| Struct, map, and configuration binding | `encoding/json`, `flag`, `os.LookupEnv`, or direct assignment keeps a small data shape explicit. | Use `vbean` or `vconf` for tag-aware copy/decode, `DecodeResult` metadata, `StrictUnused`, `DecodeHook`, profile overlays, environment expansion, or safe remote config. |

### vslice / vmap generic main path

`vslice` and `vmap` are the generic collection facades. Their main path is typed, order-aware where possible, and avoids reflection. Use these APIs before compatibility or broad object helpers:

| Workflow | `vslice` main path | `vmap` main path | Standard library alternative |
| --- | --- | --- | --- |
| Transform | `Map`, `MapErr`, `FlatMap` | `Map`, `MapErr`, `MapKeys`, `MapValues` | Plain `for` loop when local and clearer. |
| Filter | `Filter`, `Reject`, `FilterErr` | `Filter`, `Reject`, `FilterKeys`, `FilterValues` | Plain `for` loop with append/map assignment. |
| Aggregate | `Reduce`, `ReduceErr`, `CountBy`, `GroupBy` | `Reduce`, `ReduceErr`, `CountBy`, `GroupBy` | Plain loop when no reusable helper is needed. |
| Lookup | `Contains`, `Find`, `FindIndex` | `ContainsKey`, `GetAny`, `Find`, `FindKey` | `slices.Contains`, direct map lookup. |
| Shape | `Chunk`, `Window`, `Flatten`, `PartitionBy`, `Page` | `Keys`, `Values`, `Entries`, `SortedKeys`, `Pick`, `Omit` | `slices` / `maps` packages for direct operations. |
| Set-like operations | `Union`, `Intersection`, `Subtract`, `Uniq` | `Intersect`, `Diff`, `SymmetricDiff` | Direct loops when the domain rule is custom. |

Generic collection contract:

- Nil inputs are accepted by read-only helpers; returned collection values are initialized unless a function explicitly documents in-place mutation.
- `vslice` preserves input order for transform/filter/group item order; `vmap` follows Go map iteration order unless using sorted helpers.
- `MapErr`, `FilterErr`, and `ReduceErr` stop on the first error and return work completed before the failing callback.
- Use `make bench-facade BENCH=Benchmark BENCHCOUNT=10 BENCHTIME=3s` before making performance claims about collection helpers.

### API choice matrix

| Need | Start with | Use instead when |
| --- | --- | --- |
| HTTP request helpers with standard-library semantics | `vhttp` | Use `vresty` when Resty chaining is already part of the application; use `vurl` when no request should be sent yet. |
| URL parsing, normalization, query strings, safe URL opening, or content-length probing | `vurl` | Use `vhttp` or `vresty` only after the caller needs an HTTP request/response workflow. |
| Resty-style fluent request construction | `vresty` | Use `vhttp` when minimizing dependencies or matching `net/http` behavior is more important than fluent chaining. |
| Security-oriented digests, HMAC, AES-GCM, RSA, PEM, X.509, or parameter signing | `vcrypto` | Use `vhash` for non-cryptographic hashes and checksums that must not be used as security controls. |
| JSON object/array/path convenience helpers | `vjson` | Use `encoding/json` directly for streaming/token-level control; use `vxml` for XML-specific tree/namespace workflows. |
| Struct/map property mapping | `vbean` | Use `vobj` only for object-level convenience wrappers; implement new domain logic in the focused package first. |
| File reads, writes, temp paths, locks, and provider-backed filesystem tests | `vfile` | Use `vzip` for archive entry creation/extraction; use `vurl` for remote URL resource checks. |
| Generated IDs such as UUID, Snowflake, and NanoId | `vid` | Use `vident` for legal identity number parsing/validation, not generated service identifiers. |
| Provider-neutral AI, FTP, SSH/SFTP, pinyin, or tokenization contracts | `vai`, `vftp`, `vssh`, `vhan`, `vtok` | Use a dedicated provider/client package outside core when real network clients, credentials, dictionaries, or NLP engines are required. |

### Capability domain map

Use capability domains for planning and review. They group packages by the engineering capability that must stay consistent across facades, so changes can be validated against shared contracts instead of package names alone.

| Capability domain | Packages | Governance focus | Required tests |
| --- | --- | --- | --- |
| Data transform | `vbean`, `vconf`, `vconv`, `vjson`, `vobj`, `vxml` | Dynamic semantic contracts, fuzz/property coverage for untyped input, and explicit error taxonomy. | contract, fuzz, error contract, example |
| Collections | `vmap`, `vset`, `vslice` | Standard-library-first API convergence, allocation-aware benchmarks, and generic type behavior. | contract, benchmark, example |
| Text parsing | `vdfa`, `vhan`, `vregex`, `vstr`, `vtok` | Unicode and malformed-input tests, provider validation, and deterministic examples. | contract, fuzz, provider contract, example |
| Trust boundary | `vcli`, `vconf`, `vfile`, `vhttp`, `vresty`, `vurl`, `vzip` | Threat-model mapped misuse tests, `Safe`/`E`/`WithOptions` consistency, bounded IO, and fail-closed defaults. | contract, security, misuse, fuzz, error contract |
| Security primitives | `vcrypto`, `vid`, `vjwt`, `vmask`, `vpass`, `vrand` | Weak-input rejection, secret handling guarantees, algorithm policy, and random-source policy. | contract, security, misuse, error contract, benchmark |
| Runtime adapters | `vai`, `vcron`, `vdb`, `verr`, `vftp`, `vimg`, `vjob`, `vlog`, `vmail`, `vnet`, `vpoi`, `vresty`, `vskt`, `vsys`, `vtpl`, `vssh` | Context cancellation, provider injection, dependency isolation, and lifecycle metadata. | contract, provider contract, security, benchmark, example |
| Domain helpers | `vblf`, `vbool`, `vcache`, `vcodec`, `vcsv`, `vdate`, `vform`, `vhash`, `vident`, `vnum`, `vref`, `vsem`, `vver` | Package-level examples, edge-case tests, and core dependency discipline. | contract, example, benchmark |

The capability map is machine-readable in `ai-context.json` under `capability_domains` and is validated by `make governance-maturity-check`. Every public facade must be covered by at least one capability domain, and each domain must declare the test responsibility types expected for future work.

### Dependency tier matrix

`knifer-go` keeps the default utility surface lightweight. Core facades should not pull optional heavy dependencies into common slice/map/string/config/crypto workflows; extension facades isolate heavier adapters and provider contracts.

| Tier | Packages | Dependency rule |
| --- | --- | --- |
| Core facades | `vbean`, `vconv`, `vconf`, `vcrypto`, `vfile`, `vhttp`, `vjson`, `vmap`, `vslice`, `vstr`, and other standard utility facades | Standard-library-first; third-party imports must be explicitly allowlisted and are checked by `make arch`. |
| Heavy extension facades | `verr`, `vimg`, `vpoi`, `vresty` | Optional integrations such as Sentry, Logrus, image/barcode adapters, Excelize, and Resty stay in their owning facade/internal package family. |
| Provider contract facades | `vai`, `vftp`, `vhan`, `vssh`, `vtok` | Public API exposes provider interfaces and call contracts; concrete clients, credentials, dictionaries, or NLP engines belong outside the lightweight core. |

The tier inventory is machine-readable in `ai-context.json` under `dependency_tiers` and is validated by `make ai-context-check`. Heavy dependency bleed-through is blocked by `make arch`.

Physical dependency rules:

- Core facade production files must delegate to `internal/*` and may not import third-party runtime dependencies unless `bin/check_arch.sh` explicitly allowlists the facade.
- Heavy extensions keep optional integrations inside `verr`, `vimg`, `vpoi`, `vresty`, or their matching internal package families.
- Provider contract facades expose interfaces and call contracts only; concrete clients, credentials, dictionaries, and NLP engines stay outside the lightweight core.
- Adding a third-party dependency to a core facade requires an API decision card and must pass `make arch` and `make ai-context-check`.

### API decision card

Before adding or reshaping a public facade API, write a short decision card in the PR description or design note. The machine-readable template lives in `ai-context.json` under `ai_tooling.api_decision_card_template`.

| Field | Question to answer |
| --- | --- |
| Problem | What user workflow, compatibility gap, or safety issue requires a public API change? |
| Package boundary | Which `v*` facade owns the behavior, and why not a neighboring package or the standard library? |
| Proposed API | What names, signatures, options, error behavior, and API status are being introduced? |
| Alternatives | Which existing helper, internal-only helper, direct standard-library call, or external library was considered? |
| Safety and errors | How are trust boundaries, cancellation, panic policy, and `errors.Is` / `errors.As` handled? |
| Examples and docs | Which godoc, Example tests, quickstart matrices, generated catalogs, and AI metadata need updates? |
| Validation | Which focused tests, fuzz/property checks, API snapshots, generated docs, and benchmark/benchstat evidence prove the decision? |

### vbean / vconf / vobj boundary rule

These packages are intentionally adjacent but not interchangeable:

| If the data is... | Use | Why |
| --- | --- | --- |
| Configuration text, files, profiles, environment expansion, remote config, schema validation, or file watching | `vconf` | It owns configuration source loading, precedence, profile overlays, validation, and safe remote config boundaries. |
| A struct/map value that must be copied, decoded, merged, tagged, weakly converted, or reported with matched/unused metadata | `vbean` | It owns reflection-based property mapping and conversion between Go object shapes. |
| A dynamic `any` value that needs nil/empty checks, length/membership checks, defaulting, comparison, type inspection, or serialization-based cloning | `vobj` | It owns generic object-level convenience helpers, not domain mapping or configuration loading. |

When a workflow crosses boundaries, keep each step explicit: load and validate with `vconf`, bind or map with `vbean`, then use `vobj` only for generic object checks or cloning. Do not add configuration parsing to `vbean`, struct binding policy to `vobj`, or broad object helpers to `vconf`.

### Copy / Decode / Merge / Clone semantic matrix

Use this matrix when a workflow can be implemented by several reflection/object helpers. Prefer the helper whose mutation, conversion, and error semantics match the call site instead of choosing the shortest name.

| Operation | Package | Mutates destination | Conversion policy | Metadata | Failure style | Use when |
| --- | --- | --- | --- | --- | --- | --- |
| `Copy` / `CopyProperties` | `vbean` | Yes, caller-owned struct pointer or `map[string]any` | Assignable/convertible values plus configured weak conversion | No | Returns error; `Copy` is compatibility alias of `CopyProperties` | Trusted Go values need property-level copy between struct/map shapes. |
| `Decode` | `vbean` | Yes, caller-owned struct pointer or `map[string]any` | Weak string/numeric/bool conversion and optional `WithDecodeHook` | No | Returns first field-path error | Boundary data needs map/struct binding and invalid conversions must be visible. |
| `DecodeResult` | `vbean` | Yes | Same as `Decode` | `Matched`, `Skipped`, `Unused` | Returns metadata plus first error | Callers must reject or explain unused input. |
| `Merge` / `MergeWithOptions` | `vbean` | Yes, existing destination | Same as `CopyProperties`; later sources override earlier sources | No | Returns first source error | Layered Go values should update one destination with visible precedence. |
| `MergeResult` / `MergeResultWithOptions` | `vbean` | Yes | Same as `Merge` | Aggregate `Matched`, `Skipped`, `Unused` | Returns aggregate metadata plus first error | Layered boundary payloads need unused-field reporting. |
| `Clone` / `CloneWithOptions` | `vobj` | No, returns a new value | Serialization codec round trip | No | Returns codec error | A deep copy is needed and serialization semantics are acceptable. |
| `CloneIfPossible` | `vobj` | No, returns a new or original-compatible value | Serialization codec round trip when possible | No | Swallows clone failure and returns fallback | Best-effort compatibility paths where failure should not interrupt work. |
| `CloneByStream` / `CloneByStreamWithOptions` | `vobj` | No, returns a new value | Stream codec round trip | No | Returns codec error | Large values should avoid keeping the full encoded payload in an intermediate byte slice. |

<a id="practical-cookbook"></a>

## 🍳 Practical cookbook

Use these recipes as the shortest path from an application task to a reviewed package boundary. Each recipe names the first package to open, the safer default API family, and the validation command that proves examples and generated catalogs stayed current.

| Task | Start with | Minimal recipe | Validate |
| --- | --- | --- | --- |
| Filter, map, deduplicate, or page a slice | `vslice` | Use `Filter`, `Map`, `Uniq`, `Chunk`, or `Page`; choose `MapErr` / `FilterErr` when a callback can fail. Use package benchmarks only as local baselines before making performance claims. | `go test ./vslice` and `make bench-facade BENCH=Benchmark` |
| Pick, omit, merge, or sort map data | `vmap` | Use `Pick`, `Omit`, `Merge`, `SortedKeys`, `MapValues`, or `Filter`; use package benchmarks only as local baselines before making performance claims about map helpers. | `go test ./vmap` and `make bench-facade BENCH=Benchmark` |
| Parse user-provided scalar values | `vconv` | Use `ToIntE`, `ToFloat64E`, or `ToBoolE` when invalid input must be visible; use default-returning variants only when fallback behavior is intended. | `go test ./vconv` |
| Read or mutate structured JSON | `vjson` | Start from the [`Daily JSON/File FAQ`](daily-json-file-faq.md). Use object/path helpers for small in-memory JSON documents; use `encoding/json.Decoder` directly for streams or token-level control. | `go test ./vjson` |
| Fetch a URL controlled by users or config | `vurl`, then `vhttp` or `vresty` | Start from the [`Safe HTTP cookbook`](safe-http-cookbook.md). Validate/probe with `OpenSafe` or `ContentLengthSafe`, then use `GetStringSafeE`, `DownloadBytesSafeE`, or safe Resty equivalents when a request is needed. | `go test ./vurl ./vhttp ./vresty` and `make agent-security-check` |
| Load configuration from local files or remote URLs | `vconf` | Use `Load` for local trusted files, `LoadRemoteSafe` for untrusted HTTP(S), and `LoadWithOptions` / `LoadRemoteSafeWithOptions` when limits or providers must be reviewable. | `go test ./vconf` |
| Map structs and maps without JSON serialization | `vbean` | Use `ToStruct`, `ToMap`, `CopyProperties`, or `DecodeResult`; keep weak conversion and strict-unused behavior explicit through `WithWeaklyTyped` and `WithStrictUnused`. | `go test ./vbean` |
| Choose crypto or random helpers | `vcrypto`, `vrand`, `vjwt` | Start from the [`Safe Crypto cookbook`](safe-crypto-cookbook.md). Use AES-GCM, HMAC-SHA-2, RSA-OAEP/PSS, `SecureBytes`, and signed JWT helpers; do not use `vhash` for security decisions. | `go test ./vcrypto ./vrand ./vjwt` and `make agent-security-check` |

When adding a cookbook-backed public facade example, also run `make tools-gen` so `docs/api/tools.json` and `docs/api/tools.md` expose the new examples to AI agents.

<a id="build-test-and-release-workflow"></a>

## 📦 Build, test, and release workflow

Clone the source code:

```bash
git clone https://github.com/imajinyun/knifer-go.git
cd knifer-go
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

Optionally install local Git hooks so `make quick-check` runs before commit and `make full-check COVERAGE_FILE=/tmp/knifer-go-coverage.out` runs before push:

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
make benchstat BENCH_BASELINE=/tmp/knifer-go-old.bench BENCH_CURRENT=/tmp/knifer-go-new.bench
```

Use `make bench-smoke` to verify benchmark health quickly after changing hot paths. Use `make bench-core`, `make bench-facade`, and `make bench-codec` for stable package groups. Treat single-run benchmark output as a baseline only; use repeated runs and `benchstat` before documenting an improvement or regression.

Performance budget workflow:

| Change type | Required measurement | Budget rule |
| --- | --- | --- |
| Hot-path implementation change in slice/map/string/codec/bean/json/xml helpers | Run the relevant `make bench-* BENCHCOUNT=10 BENCHTIME=3s` target before and after the change, then compare with `make benchstat`. | Do not claim a speedup unless `benchstat` reports a statistically significant improvement. Investigate regressions above 10% in `ns/op`, `B/op`, or `allocs/op`. |
| Facade-only wrapper change | Run `make bench-smoke`; run focused facade benchmarks only if the wrapper adds allocation, reflection, parsing, or provider dispatch. | Facade overhead should stay below measurement noise for pure delegation. |
| Documentation-only benchmark claim | Include the exact command, package list, Go version, and benchstat output near the claim. | Never publish performance comparisons from a single benchmark run. |

Historical benchmark baseline workflow:

```bash
make bench-baseline BENCHCOUNT=10 BENCHTIME=3s BENCH_BASELINE_OUT=/tmp/knifer-go-bench-baseline.txt
make bench-compare BENCHCOUNT=10 BENCHTIME=3s BENCH_BASELINE_OUT=/tmp/knifer-go-bench-baseline.txt BENCH_CURRENT_OUT=/tmp/knifer-go-bench-current.txt
```

Core performance budgets:

| Package group | Budget expectation |
| --- | --- |
| `vslice`, `vmap`, `vstr` generic helpers | Prefer direct loops or standard library when they are clearer; investigate >10% regression in `ns/op` or `allocs/op` for hot helpers. |
| `vconv` scalar conversion | `E` helpers may add validation branches; preserve documented failure behavior before optimizing. |
| `vbean`, `vjson`, `vxml` reflection/dynamic helpers | Reflection and allocation are expected; benchmark against typed code before using in hot paths. |
| `vcodec` encode/decode helpers | Round-trip correctness comes first; compare `B/op` and `allocs/op` for payload-size-sensitive changes. |

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

GitHub Actions reuses the Makefile targets for module verification, vet, tidy checks, diff cleanliness, architecture checks, race/shuffle tests, coverage gates, API compatibility checks, generated tool-catalog checks, and AI metadata checks. It also runs `golangci-lint`, `govulncheck`, CodeQL, benchmark smoke tests, and OpenSSF Scorecard. Dependabot is configured for Go modules and GitHub Actions updates.

### v1 readiness checklist

Use this release gate before declaring a v1-ready surface:

| Area | Exit criteria |
| --- | --- |
| Public API | `docs/api/exports.txt` is current and every public API addition has an API decision card. |
| Catalogs and AI metadata | `docs/api/tools.json`, `docs/api/tools.md`, and `ai-context.json` are current; every facade has recommended entrypoints and dependency-tier metadata. |
| Semantics | Conversion, decode hook, field-path errors, Copy/Decode/Merge/Clone behavior, and package boundary matrices are documented and tested. |
| Safety | Security-sensitive packages use Safe/E/WithOptions variants at trust boundaries and pass coverage gates. |
| Reliability | `make release-check`, `make fuzz-smoke`, and `make bench-smoke` pass; benchmark claims include repeated runs and `benchstat`. |
| Blocking failures | No stale generated artifacts, architecture violations, heavy dependency bleed-through, unresolved lint/govulncheck findings, or undocumented public APIs remain. |

### v1 API freeze and deprecation gate

Run the API freeze gate before release branches and any v1 candidate tag:

```bash
make api-freeze-check
```

Freeze rules:

- Public API additions, removals, or signature changes require an API decision card.
- `docs/api/exports.txt`, `docs/api/tools.json`, and `docs/api/tools.md` must be current.
- Experimental APIs are blocked while `ai-context.json` marks the project as a v1 candidate.
- Deprecated APIs must include a replacement and rationale in `ai-context.json` and their godoc synopsis.
- Compatibility APIs may remain for migration, but examples and generated `golden_path` guidance should prefer recommended alternatives.

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
- Workflow gates: use `make doctor` for environment diagnostics, `make worktree-check` to block unrelated untracked Go files, `make quick-check` for fast local validation, `make security-check` for lint and vulnerability scanning, `make full-check COVERAGE_FILE=/tmp/knifer-go-coverage.out` for the full pre-push gate, and `make ci-test` for the GitHub Actions test-job gate. GitHub Actions additionally runs CodeQL, OpenSSF Scorecard, and a benchmark smoke gate. Optional Git hooks can be enabled with `make install-hooks` and disabled with `make uninstall-hooks`.
- Security suppressions: keep `.golangci.yml`, `#nosec`, and `//nolint:gosec` exceptions narrow and justified at the call site; prefer a regression test before broadening an exclusion.
- Benchmark baseline: use `make bench-core` for hot-path benchmark suites or `make bench-facade` for matching public facade packages. Treat the output as a baseline unless a separate `benchstat` comparison proves a performance change.

<a id="contributing"></a>

## 🤝 Contributing

If you find a bug or want to request a new utility, please open a GitHub Issue. It is recommended to include:

- Go version and operating system;
- `knifer-go` version or commit;
- Minimal reproducible code;
- Expected behavior and actual behavior;
- Related error logs or test output.

Pull requests are welcome. To keep the toolkit stable, please follow these principles where possible:

1. Add new capabilities to the appropriate `internal/*` implementation package first, then expose public APIs from the corresponding `v*` package;
2. Add necessary comments for new or modified public APIs;
3. Add unit tests for core logic and run `go test ./...` before submitting;
4. Keep code formatted with `gofmt`;
5. Avoid unnecessary third-party dependencies and prefer the standard library when possible.
