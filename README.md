# go-knifer

> 🍬 A set of Go tools that keep development sharp.

![go-knifer](./go-knifer.jpeg)

[![Go Reference](https://pkg.go.dev/badge/github.com/imajinyun/go-knifer.svg)](https://pkg.go.dev/github.com/imajinyun/go-knifer)
[![Go Version](https://img.shields.io/badge/go-%3E%3D1.25-00ADD8?logo=go)](https://go.dev/)
[![CI](https://github.com/imajinyun/go-knifer/actions/workflows/go.yml/badge.svg)](https://github.com/imajinyun/go-knifer/actions/workflows/go.yml)
[![License](https://img.shields.io/github/license/imajinyun/go-knifer)](./LICENSE)

## 📚 Introduction

`go-knifer` is a practical utility toolkit for Go projects. It collects frequently used capabilities—string helpers, collection utilities, encoding/decoding, cryptography, HTTP, JSON, cache, cron, JWT, logging, configuration, and system information—into reusable packages.

The root package `github.com/imajinyun/go-knifer` only acts as the module entry point. Actual APIs are split into multiple public `v*` packages by domain, so users can import only what they need without mixing unrelated utilities into business code.

## 🔪 Origin of the `go-knifer` name

`knifer` comes from “knife”: a handy little tool for solving common everyday problems in Go development. It does not try to replace the standard library. Instead, it lightly wraps standard-library features and common engineering practices to make code shorter, more consistent, and easier to maintain.

## ✨ How go-knifer changes the way we code

`go-knifer` moves frequent, repetitive, and copy-paste-prone utility logic into focused `v*` subpackages. Application code imports the facade it needs by domain and uses consistent APIs for the same scenarios across a team.

## 🚀 Install

Go 1.25 or later is required.

```bash
go get github.com/imajinyun/go-knifer
```

## 🧭 Find by scenario

Not sure which package to import? Start from what you want to do:

| I want to… | Use |
| --- | --- |
| Cache with FIFO/LRU/LFU/TTL | [`vcache`](docs/doc/04-vcache.md) |
| Base64 / Hex encode-decode | [`vcodec`](docs/doc/05-vcodec.md) |
| Load local or remote configuration, including SSRF-checked remote config | [`vconf`](docs/doc/06-vconf.md) |
| Loosely convert `any` to int/float/bool/string | [`vconv`](docs/doc/07-vconv.md) |
| Schedule cron tasks | [`vcron`](docs/doc/08-vcron.md) |
| SHA/HMAC, AES-GCM/RSA-PSS, sign parameters | [`vcrypto`](docs/doc/09-vcrypto.md) |
| Read/write CSV records, maps, or structs | [`vcsv`](docs/doc/10-vcsv.md) |
| Format/parse dates, offsets, day ranges | [`vdate`](docs/doc/11-vdate.md) |
| Read/write files, paths, copy, mkdir | [`vfile`](docs/doc/15-vfile.md) |
| Validate form/input data such as email/mobile/IP | [`vform`](docs/doc/16-vform.md) |
| Non-cryptographic hashes (FNV, BKDR, …) | [`vhash`](docs/doc/17-vhash.md) |
| Send HTTP requests (standard library) | [`vhttp`](docs/doc/18-vhttp.md) |
| Generate UUID / Snowflake / NanoId | [`vid`](docs/doc/19-vid.md) |
| Validate or parse ID-card numbers | [`vident`](docs/doc/20-vident.md) |
| Produce image thumbnails, convert PNG/JPEG/GIF, read metadata, or generate graphical captchas | [`vimg`](docs/doc/21-vimg.md) |
| Build/parse JSON, path get/put, JSON↔XML | [`vjson`](docs/doc/23-vjson.md) |
| JWT sign/verify | [`vjwt`](docs/doc/24-vjwt.md) |
| Compose and send email with text, HTML, inline files, or attachments | [`vmail`](docs/doc/26-vmail.md) |
| Create, query, transform, merge, diff, or sort maps | [`vmap`](docs/doc/27-vmap.md) |
| Mask sensitive data | [`vmask`](docs/doc/28-vmask.md) |
| Precise arithmetic, rounding, or evaluate an expression | [`vnum`](docs/doc/30-vnum.md) |
| Score and classify password strength locally | [`vpass`](docs/doc/32-vpass.md) |
| Send HTTP requests (Resty-based) | [`vresty`](docs/doc/37-vresty.md) |
| Filter / map / dedup / paginate a slice | [`vslice`](docs/doc/41-vslice.md) |
| Trim, split, case-convert, Unicode-escape, Ant-match, compare text similarity, or check blank strings | [`vstr`](docs/doc/42-vstr.md) |
| Encode/parse URLs, build/parse query strings, or open untrusted HTTP(S) resources safely | [`vurl`](docs/doc/45-vurl.md) |
| Parse, build, or navigate XML | [`vxml`](docs/doc/47-vxml.md) |

For the full list, see the module matrix below.

## 🧩 Module

The project follows an “internal implementation + public facade” layout: `internal/*` contains concrete implementations, while `v*` packages expose stable public APIs.

| Module | Import path | Description |
| --- | --- | --- |
| [`vbean`](docs/doc/01-vbean.md) | `github.com/imajinyun/go-knifer/vbean` | Bean/struct mapping helpers: struct/map conversion, copy properties, tag and alias matching, ignore-empty/zero options, and weak type conversion. |
| [`vblf`](docs/doc/02-vblf.md) | `github.com/imajinyun/go-knifer/vblf` | Bloom filters: bitmap/bitset/filter abstractions, multiple string hash algorithms, option-based constructors, `E` constructors that return validation errors instead of panicking, and provider-backed file initialization. |
| [`vbool`](docs/doc/03-vbool.md) | `github.com/imajinyun/go-knifer/vbool` | Boolean helpers: negate, bool-to-int, all/any checks. |
| [`vcache`](docs/doc/04-vcache.md) | `github.com/imajinyun/go-knifer/vcache` | Generic caches: FIFO, LFU, LRU, Timed, Weak, and NoCache; supports TTL, clocks, removal listeners, lazy loading, ticker/runner providers, and weak-cache finalizer providers. Removal listeners run outside cache locks, so callbacks can safely re-enter the same cache. |
| [`vcodec`](docs/doc/05-vcodec.md) | `github.com/imajinyun/go-knifer/vcodec` | Encoding helpers: Base64, URL-safe Base64, raw URL-safe Base64, custom Base64 encoding providers, and Hex. |
| [`vconf`](docs/doc/06-vconf.md) | `github.com/imajinyun/go-knifer/vconf` | Grouped configuration reader for setting/properties-style text, a simple YAML subset, and TOML parsing, with typed getters, schema validation, profile/remote/file loading options, SSRF-checked `LoadRemoteSafe`, environment expansion providers, watch ticker/runner providers, bounded reads, read-only snapshot guidance, and deep-copy `Clone` support. |
| [`vconv`](docs/doc/07-vconv.md) | `github.com/imajinyun/go-knifer/vconv` | Permissive type conversion: string, int, int64, float64, bool, bytes, and default-value variants. |
| [`vcron`](docs/doc/08-vcron.md) | `github.com/imajinyun/go-knifer/vcron` | Cron expression parsing and task scheduling, including default/custom schedulers, configurable cron options, ID random-reader/clock/sleeper/runner providers, isolated per-call default scheduler overrides, running-task metrics, `Wait`, and graceful `Shutdown(ctx)`. |
| [`vcrypto`](docs/doc/09-vcrypto.md) | `github.com/imajinyun/go-knifer/vcrypto` | Cryptography and digests: SHA-2, provider-backed digest helpers, HMAC, PBKDF2-SHA256, parameter signing, random bytes, AES-GCM with nonce/tag/block-factory options, RSA OAEP/PSS plus configurable data-signing options, PEM, and X.509 certificate helpers. |
| [`vcsv`](docs/doc/10-vcsv.md) | `github.com/imajinyun/go-knifer/vcsv` | CSV helpers: reader/writer options for delimiters, comments, field counts, lazy quotes, trimming, CRLF output, records-to-map conversion, map writing, struct tag export, and record callbacks. |
| [`vdate`](docs/doc/11-vdate.md) | `github.com/imajinyun/go-knifer/vdate` | Date/time helpers: common layouts, parse/format, begin/end of day/month/year, offsets, and comparisons. |
| [`vdb`](docs/doc/12-vdb.md) | `github.com/imajinyun/go-knifer/vdb` | Database helpers built on database/sql: SQL execution, named parameters, entities, conditions, query builders, transactions, pagination, lightweight metadata lookup, and injectable `sql.Open` providers. |
| [`vdfa`](docs/doc/13-vdfa.md) | `github.com/imajinyun/go-knifer/vdfa` | DFA word-tree matching: stop-rune filtering, first/all matches, dense and greedy match modes, found-word metadata, package-level matcher helpers, isolated matcher options, JSON marshal/unmarshal providers for `Any` helpers, text replacement, and resettable async runner providers for package-level initialization. |
| [`verr`](docs/doc/14-verr.md) | `github.com/imajinyun/go-knifer/verr` | Error helpers: panic recovery, error aggregation, multierror matching, collector construction options, stack capture/formatting, resettable log/stack caches, injectable logging/stack/exit/timer/runner providers, isolated logrus creation, and optional logrus/Sentry integration. |
| [`vfile`](docs/doc/15-vfile.md) | `github.com/imajinyun/go-knifer/vfile` | File and IO helpers: read/write/copy, lines, mkdir/touch/delete, filename helpers, quiet close, and provider-backed file-system operations. |
| [`vform`](docs/doc/16-vform.md) | `github.com/imajinyun/go-knifer/vform` | Form and input validation helpers: email, mobile, URL, IPv4/IPv6, ID card, Chinese text, number string checks, and per-call matcher providers for rule-sensitive checks. |
| [`vhash`](docs/doc/17-vhash.md) | `github.com/imajinyun/go-knifer/vhash` | Non-cryptographic hash helpers: additive, FNV, injectable 32-bit hash providers, and a set of classic string hashes (RS, JS, PJW, ELF, BKDR, SDBM, DJB, AP, HF, HFIP, TianL, Java default). |
| [`vhttp`](docs/doc/18-vhttp.md) | `github.com/imajinyun/go-knifer/vhttp` | Chainable HTTP client, isolated/global-config request construction, create/get/post `WithOptions` helpers, explicit-error `E` shortcuts, code-classified HTTP errors, provider-backed transports/request factories/multipart writers/download saves, safe file downloads, BasicAuth, User-Agent parsing, provider-backed HTML cleaning/filtering, resettable transport/server starters, async server runner options, and simple server helpers. |
| [`vid`](docs/doc/19-vid.md) | `github.com/imajinyun/go-knifer/vid` | ID helpers: random/simple/fast UUIDs, MongoDB-style ObjectId, Snowflake generators and singleton next-id helpers, worker/datacenter id derivation, NanoId generation, fallback random sources, isolated Snowflake creation, and resettable fallback PRNG providers/seeds. |
| [`vident`](docs/doc/20-vident.md) | `github.com/imajinyun/go-knifer/vident` | Identity helpers: mainland China ID card 15/18-digit conversion, validation, check code, birthday/age/gender extraction with parsing options, province/city/district code parsing, masking, and Hong Kong/Macau/Taiwan card validation. |
| [`vimg`](docs/doc/21-vimg.md) | `github.com/imajinyun/go-knifer/vimg` | Image helpers: proportional thumbnails, format conversion between PNG/JPEG/GIF, metadata introspection (width/height/format), and image captcha generation through line/circle/shear/GIF captcha types. |
| [`vjob`](docs/doc/22-vjob.md) | `github.com/imajinyun/go-knifer/vjob` | Sliceable job execution: separate job data from scheduling options, typed slice/map adapters, context cancellation, and serialized merge callbacks; no generic type-alias experiment is required. |
| [`vjson`](docs/doc/23-vjson.md) | `github.com/imajinyun/go-knifer/vjson` | Ordered JSON objects/arrays, JSON parsing and formatting, path-based get/put, provider-backed marshal/unmarshal, injectable scalar parse/format functions, configurable object/array/bean/list conversion, and XML/JSON conversion with parser/writer options. |
| [`vjwt`](docs/doc/24-vjwt.md) | `github.com/imajinyun/go-knifer/vjwt` | JWT creation, parsing, signing, verification, and time-claim validation; supports HMAC, RSA-PSS, ECDSA, rejects unsigned `alg=none` tokens, and provides JSON marshal/unmarshal options. |
| [`vlog`](docs/doc/25-vlog.md) | `github.com/imajinyun/go-knifer/vlog` | Logging facade: console/color console loggers, injectable color factories, log levels, global logger, static logging functions, per-call logger options, and isolated logger creation. |
| [`vmail`](docs/doc/26-vmail.md) | `github.com/imajinyun/go-knifer/vmail` | Mail helpers: RFC 5322 address parsing, fluent message construction, MIME mixed/related/alternative rendering, text/HTML bodies, inline files, attachments, context-aware SMTP sending, mandatory TLS defaults, CRLF injection checks, attachment size limits, and injectable senders/dialers/boundary generators. |
| [`vmap`](docs/doc/27-vmap.md) | `github.com/imajinyun/go-knifer/vmap` | Map helpers: construction, empty checks, contains/get/find, keys/values and sorted views, map/filter/reject/partition, reduce/group/count, inverse, merge/merge-with-resolver, intersect/diff/symmetric diff, pick/omit, update/clone, and equality checks. |
| [`vmask`](docs/doc/28-vmask.md) | `github.com/imajinyun/go-knifer/vmask` | Masking helpers: mask names, IDs, phones, addresses, email, passwords, license plates, bank cards, IPs, passports, and credit codes. |
| [`vnet`](docs/doc/29-vnet.md) | `github.com/imajinyun/go-knifer/vnet` | Network helpers: IPv4/IPv6 conversion, CIDR/range/mask utilities with injectable IP/CIDR/int parsers, local ports, host/interface/MAC lookup, TLS config, address/dial/ping provider options, and multipart form helpers. |
| [`vnum`](docs/doc/30-vnum.md) | `github.com/imajinyun/go-knifer/vnum` | Numeric helpers: precise arithmetic, generic sum/average/min/max/abs helpers, rounding modes, provider-backed parsing/formatting, number checks, random unique numbers, ranges, factorial/combinations, gcd/lcm, binary conversion, comparison, byte conversion, expression calculation, and odd/even checks. |
| [`vobj`](docs/doc/31-vobj.md) | `github.com/imajinyun/go-knifer/vobj` | Object helpers: nil/empty checks, equality, defaults, clone/serialization, comparison, type inspection, and container utilities. |
| [`vpass`](docs/doc/32-vpass.md) | `github.com/imajinyun/go-knifer/vpass` | Password helpers: deterministic local scoring, strength buckets, strong/weak predicates, character-class signals, repeated/sequential-run detection, and a small common-weak-password blocklist. |
| [`vpoi`](docs/doc/33-vpoi.md) | `github.com/imajinyun/go-knifer/vpoi` | Office document helpers: lightweight Excel XLSX sheet listing, row reading/writing, multi-sheet writing, in-memory workbook creation, and injectable workbook/file-system providers. |
| [`vrand`](docs/doc/34-vrand.md) | `github.com/imajinyun/go-knifer/vrand` | Random helpers: integers, floats, booleans, bytes, strings, numeric strings, random element selection, deterministic seeding, and resettable package-level pseudo-random source providers. |
| [`vref`](docs/doc/35-vref.md) | `github.com/imajinyun/go-knifer/vref` | Reflection helpers: field lookup and mutation, method discovery and invocation, constructor-style function calls, nil-safe type/value utilities, object-level type predicates (`IsFunction`, `IsIteratee`, `IsCollection`, `IsSlice`, `IsArray`, `IsMap`), method classification, and explicit unsafe/unexported field-access options. |
| [`vregex`](docs/doc/36-vregex.md) | `github.com/imajinyun/go-knifer/vregex` | Regular-expression helpers: matching, group extraction, named groups, deletion, counting, index lookup, template/function replacement, escaping, and per-call compiler / DOTALL options. |
| [`vresty`](docs/doc/37-vresty.md) | `github.com/imajinyun/go-knifer/vresty` | Resty v3 based HTTP facade: chainable requests, JSON/form/multipart bodies, isolated/global-config request construction, create/get/post `WithOptions` helpers, per-request client factories, resettable default Resty client providers, downloads and safe file downloads, and lightweight response helpers. |
| [`vsem`](docs/doc/38-vsem.md) | `github.com/imajinyun/go-knifer/vsem` | Weighted, context-aware counting semaphore with FIFO fairness, try-acquire, close notifications, and in-use metrics. |
| [`vset`](docs/doc/39-vset.md) | `github.com/imajinyun/go-knifer/vset` | Generic and typed set utilities with add/remove/contains, set operations, and JSON/YAML encoding helpers. |
| [`vskt`](docs/doc/40-vskt.md) | `github.com/imajinyun/go-knifer/vskt` | TCP socket utilities: plain connections, NIO/AIO server/client helpers, protocol encoder/decoder interfaces, and configurable thread-pool/listener/connection/runner/IP-parser providers. |
| [`vslice`](docs/doc/41-vslice.md) | `github.com/imajinyun/go-knifer/vslice` | Slice helpers: contains/index, reverse, distinct, join, filter/map, sub-slice, concat, set-like operations, and paging. |
| [`vstr`](docs/doc/42-vstr.md) | `github.com/imajinyun/go-knifer/vstr` | String and text helpers: blank/empty checks, trimming, splitting, substring helpers, formatting, provider-backed emoji helpers, naming conversion, defaults, Unicode escaping/unescaping, Ant-style path matching, rune-set Jaccard similarity, rune n-gram similarity, SimHash, 64-bit Hamming distance, HTML escaping, and rune checks (blank, letter, digit, ASCII, letter-or-digit). |
| [`vsys`](docs/doc/43-vsys.md) | `github.com/imajinyun/go-knifer/vsys` | System and runtime information: host, OS, user, Go runtime, process memory, goroutines, environment variables, resettable info cache, and injectable env/command/runtime providers. |
| [`vtpl`](docs/doc/44-vtpl.md) | `github.com/imajinyun/go-knifer/vtpl` | Go html/template rendering helpers with per-call template name, FuncMap, delimiter, factory, and executor options. |
| [`vurl`](docs/doc/45-vurl.md) | `github.com/imajinyun/go-knifer/vurl` | URL and URI helpers: parse, normalize, resolve relative URLs, query encode/decode, URL/path/fragment percent encoding with injectable query/path escape providers, URL building, Data URI building, scheme checks, file URL conversion, resource open/size helpers, and SSRF-oriented `OpenSafe` / `ContentLengthSafe` variants. |
| [`vver`](docs/doc/46-vver.md) | `github.com/imajinyun/go-knifer/vver` | Version helpers: version comparison, greater/less predicates, expression matching, inclusive ranges, and custom expression delimiters. |
| [`vxml`](docs/doc/47-vxml.md) | `github.com/imajinyun/go-knifer/vxml` | XML helpers: parse/read/write/format, tree navigation, simple XPath-style lookup, escaping, map/bean conversion with parser/codec/scalar parser options, transform options, and namespace utilities. |
| [`vzip`](docs/doc/48-vzip.md) | `github.com/imajinyun/go-knifer/vzip` | ZIP, gzip, and zlib helpers: archive creation/extraction, entry lookup, archive traversal, append, in-memory entries, stream compression, provider-backed archive file operations, bounded extraction/decompression defaults, path traversal checks, and symlink escape checks during extraction. |

## 🧭 Architecture and package boundaries

`go-knifer` uses public `v*` packages as facade APIs and keeps concrete code in
`internal/*`. Application code should import the `v*` packages; `internal/*`
exists so implementations can evolve without exposing every helper as public API.

Facade rules:

- `internal/<domain>` owns implementation details and domain-specific tests.
- `v<domain>` exposes the stable public surface for that domain.
- Small utility packages may use hand-written thin facades; larger modules may
  keep generated `facade.go` files. In either case, newly exported internal APIs
  should be reviewed before being exposed publicly.
- Facades may keep short names such as `vform`, `vmask`, `vsem`, `vskt`,
  `vblf`, and `vver`; their meaning is documented in the module table above
  instead of changing established import paths.

API compatibility:

- Public subpackages are the compatibility boundary. The current exported API
  surface is recorded in `docs/api/exports.txt`.
- `make api-check` regenerates a temporary snapshot and compares it with the
  checked-in file. When a public API change is intentional, run
  `UPDATE_API=1 make api-check` and review the snapshot diff together with the
  implementation change.
- API additions, removals, and renames should also be reflected in package
  `doc.go` comments, examples, and the changelog before tagging a release.

Configurable APIs and provider injection:

- Many packages expose functional options through `WithXxx` helpers and
  `XxxWithOptions` variants; configuration-heavy APIs may use explicit option
  structs such as `vconf.LoadOptions`. Existing fixed-argument APIs stay stable,
  while option-based variants add advanced control for callers that need it.
- The option pattern is available across runtime-sensitive helpers such as
  bloom filters, cache, captcha, config loading/watching, cron, crypto, DB,
  date/time, DFA, errors, files, HTTP/Resty, IDs, identity, JSON/JWT, logging,
  network, numbers, POI, random, socket, system, URL, XML, and ZIP helpers.
- Provider-style options let callers inject file-system functions, network/TLS
  dialers or readers, HTTP request/multipart factories, clocks, timers/tickers,
  random sources, DB openers, Excel workbook factories, loggers, stack capture
  functions, finalizers, environment lookups, command executors, Sentry/logrus
  hooks, and other process-global dependencies for deterministic tests and
  controlled runtime behavior.
- Scalar and parser-heavy helpers also expose provider options at the call site:
  `vnet` IP/CIDR/int parsers, `vskt` host-to-IP parsing, `vjson` string/int/
  float/bool parse and int/float format providers, `vxml` XML-to-map scalar
  parsers, `vnum` expression/double parse/format providers, and `vurl` query
  or path escaping providers. The plain APIs keep their standard-library
  behavior and delegate to the option-based variants internally.
- Package-level defaults remain explicit. For example, HTTP global defaults can
  be read as an immutable snapshot via `vhttp.SnapshotGlobalConfig`, and
  `vhttp.NewIsolatedRequest` can build a request without reading package-level
  defaults. Per-call options should not mutate hidden global state.
- Configuration objects are mutable while being built or loaded, then should be
  treated as read-only snapshots after publication. Use `Conf.Clone()` to create
  an independent copy before applying runtime changes, and publish the new
  pointer atomically instead of mutating a shared instance in place.

Provider coverage highlights:

| Area | Examples |
| --- | --- |
| HTTP / Resty | `vhttp.NewIsolatedRequest`, `vhttp.NewRequestWithConfig`, `vhttp.Get`, `vhttp.Post`, `vhttp.GetSafe`, `vhttp.PostSafe`, `vhttp.GetStringE`, `vhttp.GetStringSafeE`, `vhttp.GetWithTimeoutE`, `vhttp.PostJSONE`, `vhttp.PostJSONSafeE`, `vhttp.DownloadBytesE`, `vhttp.DownloadBytesSafeE`, `vhttp.DownloadFile`, `vhttp.DownloadFileSafe`, `vhttp.DownloadFileSafeWithOptions`, `vhttp.NewErrorWithCode`, `vhttp.WithTransportProvider`, `vhttp.WithRequestFactory`, `vhttp.WithMultipartWriterFactory`, `vhttp.ResetDefaultTransport`, `vhttp.WithListenAndServeFunc`, `vhttp.WithAsyncRunner`, `vhttp.CreateServerWithOptions`, `vhttp.CleanHTMLWithOptions`, `vhttp.FilterHTMLTagWithOptions`, `vhttp.WithHTMLFilterCompileFunc`, `vresty.NewIsolatedRequest`, `vresty.WithGlobalConfig`, `vresty.WithRestyClientFactory`, `vresty.ConfigureDefaultRestyClientProvider`, `vresty.ResetDefaultRestyClientProvider`, `vresty.Get`, `vresty.Post`, `vresty.GetSafe`, `vresty.PostSafe`, `vresty.GetStringE`, `vresty.GetStringSafeE`, `vresty.GetWithTimeoutE`, `vresty.PostJSONE`, `vresty.PostJSONSafeE`, `vresty.DownloadBytesE`, `vresty.DownloadBytesSafeE`, `vresty.DownloadFile`, `vresty.DownloadFileSafe`, `vresty.DownloadFileSafeWithOptions` |
| File / config / archive / POI | `vfile` provider options, `vconf.LoadWithOptions`, `vconf.LoadRemoteSafeWithOptions`, `vconf.WatchWithOptions`, `vconf.WatchOptions.Runner`, `(*vconf.Conf).Clone`, `vzip.WithMaxBytes`, `vzip` provider options, `vpoi.WithOpenFileFunc`, `vpoi.WithNewFileFunc`, `vpoi.WithSaveAsFunc` |
| Cron / DFA / ID / identity / random | `vcron.WithDefaultSchedulerOptions`, `vcron.NewConfigWithOptions`, `vcron.WithIDRandomReader`, `vcron.WithRunner`, `vcron.CronScheduleWithOptions`, `(*vcron.Scheduler).RunningCount`, `(*vcron.Scheduler).Wait`, `vcron.CronShutdown`, `vdfa.WithMatcherWords`, `vdfa.WithJSONMarshal`, `vdfa.WithJSONUnmarshal`, `vdfa.ContainsWithOptions`, `vdfa.ConfigureAsyncRunner`, `vdfa.ResetAsyncRunner`, `vid.NewIsolatedSnowflake`, `vid.CreateSnowflakeWithOptions`, `vid.WithSnowflakeCache`, `vid.WithFallbackRandomSource`, `vid.ConfigureDefaultFallbackRandomSourceProvider`, `vid.ResetDefaultFallbackRandomSource`, `vid.SetFallbackRandomSeed`, `vrand.SecureBytes`, `vrand.ConfigureDefaultRandomSourceProvider`, `vrand.ResetDefaultRandomSource`, `vrand.SetSeed`, `vident.BirthDateWithOptions` |
| Encoding / image / JSON / XML / JWT / hash | `vcodec.Base64EncodeWithEncoding`, `vcodec.Base64DecodeWithEncoding`, `vcodec.Base64RawURLEncode`, `vcodec.Base64RawURLDecode`, `vimg.Thumbnail`, `vimg.ConvertFormat`, `vimg.Info`, `vimg.NewLineCaptcha`, `vimg.NewCircleCaptcha`, `vimg.NewShearCaptcha`, `vimg.NewGifCaptcha`, `vhash.Hash32`, `vjson.WithMarshalFunc`, `vjson.WithUnmarshalFunc`, `vjson.WithParseUnmarshalFunc`, `vjson.WithBeanUnmarshalFunc`, `vjson.WithSprintFunc`, `vjson.WithParseIntFunc`, `vjson.WithParseFloatFunc`, `vjson.WithParseBoolFunc`, `vjson.WithFormatIntFunc`, `vjson.WithFormatFloatFunc`, `vjson.ParseObjWithOptions`, `vjson.ParseArrayWithOptions`, `vjson.ToBeanWithOptions`, `vjson.ToListWithOptions`, `vjson.XMLToJSONWithOptions`, `vjson.ToXMLWithOptions`, `vxml.WithScalarIntParser`, `vxml.WithScalarFloatParser`, `vxml.XMLToMapWithOptions`, `vxml.XMLNodeToMapWithOptions`, `vxml.XMLToMapIntoWithOptions`, `vxml.XMLNodeToMapIntoWithOptions`, `vxml.XMLToBeanWithOptions`, `vxml.XMLNodeToBeanWithOptions`, `vxml.TransformWithOptions`, `vxml.FormatWithOptions`, `vjwt.WithJSONMarshalFunc`, `vjwt.WithJSONUnmarshalFunc`, `vjwt.ParseTokenWithOptions`, `vjwt.WithTokenJSONOptions` |
| Crypto / password / template / regex / validation / strings | `vcrypto.Digest`, `vcrypto.DigestHex`, `vcrypto.WithGCMBlockFactory`, `vcrypto.AESSealGCMWithOptions`, `vcrypto.AESEncryptGCMWithOptions`, `vcrypto.SignWithRSAOptions`, `vcrypto.VerifyWithRSAOptions`, `vpass.Analyze`, `vpass.Score`, `vpass.StrengthOf`, `vpass.IsStrong`, `vpass.IsWeak`, `vtpl.RenderWithOptions`, `vtpl.WithFuncMap`, `vtpl.WithTemplateFactory`, `vregex.WithCompileFunc`, `vregex.WithDotAll`, `vregex.MatchWithOptions`, `vregex.ReplaceAllFuncWithOptions`, `vform.IsEmailWithOptions`, `vform.WithMobileMatcher`, `vstr.ContainsEmojiWithOptions`, `vstr.RemoveEmojiWithOptions`, `vstr.JaccardSimilarity`, `vstr.NGramSimilarity`, `vstr.SimHash`, `vstr.HammingDistance64` |
| DB / network / number / URL / system / reflection / socket | `vdb.WithSQLOpenFunc`, `vnet.WithConnectDialer`, `vnet.WithPingDialer`, `vnet.WithAddressNetwork`, `vnet.WithTCPAddrResolver`, `vnet.WithUploadOpenSource`, `vnet.WithIPParser`, `vnet.WithCIDRParser`, `vnet.WithIPIntParser`, `vnet.WithWildcardIPParser`, `vnet.WithWildcardIntParser`, `vnet.IPv4ToLongWithOptions`, `vnet.IsInRangeWithOptions`, `vnum.WithParseFloatFunc`, `vnum.WithDoubleParseFloatFunc`, `vnum.WithDoubleFormatFloatFunc`, `vnum.CalculateWithOptions`, `vnum.ToDoubleWithOptions`, `vurl.WithQueryEscapeFunc`, `vurl.WithPathEscapeFunc`, `vurl.EncodeQueryWithOptions`, `vurl.EncodePathSegmentWithOptions`, `vurl.FormURLEncodeWithOptions`, `vurl.OpenSafeWithOptions`, `vurl.WithAllowedSchemes`, `vurl.WithAllowedHosts`, `vurl.WithRejectPrivateHosts`, `vurl.WithAllowLocalFiles`, `vsys.WithGoEnvOutputFunc`, `vsys.WithGoRootEnvLookupFunc`, `vsys.WithOSEnvLookupFunc`, `vsys.WithEnvLookupFunc`, `vsys.ResetInfoCache`, `vref.WithUnsafeAccess`, `vskt.WithThreadPoolSizeFunc`, `vskt.WithRunner`, `vskt.WithSocketIPParser` |
| Errors / cache / logging / runtime | `verr.NewCollectorWithOptions`, `verr.WithCollectorLogFunc`, `verr.WithCollectorRunner`, `verr.WithCollectorContext`, `verr.WithCollectorLevel`, `verr.WithCollectorTimerFactory`, `verr.WithCollectorStackCaptureOptions`, `verr.WithLogFunc`, `verr.WithCollectorStackOptions`, `verr.WithDebugStackFunc`, `verr.WithCallersFunc`, `verr.WithFuncForPCFunc`, `verr.WithStackFrameCache`, `verr.ResetStackFrameCache`, `verr.ResetDefaultLogFunc`, `verr.NewIsolatedLogrusWithOptions`, `verr.MustExitWithOptions`, `vcache.WithClock`, `vcache.WithTickerFactory`, `vcache.WithRunner`, `vcache.WithWeakFinalizerFunc`, `vcache.WithWeakFinalizerEnabled`, `vlog.WithLogColorFactory`, `vlog.NewIsolatedLogger`, `vlog.LoggerWithOptions`, `vlog.InfoWithOptions` |

Domain boundary rules:

- `vhash` is for non-cryptographic hash helpers such as additive/FNV (bucketing,
  bloom filters); `vcrypto` owns security-oriented SHA-2 digests,
  HMAC, encryption, and key/PEM operations.
- `vhttp` is the lightweight standard-library HTTP facade; `vresty` is the
  Resty-based chainable client facade. Neither re-exports URL helpers: URL
  escaping, query building/parsing, and scheme checks (`IsHTTP`/`IsHTTPS`,
  `EncodeQueryMap`, `DecodeQuery`, etc.) live solely in `vurl`.
- `vdb` owns SQL database helpers on top of `database/sql`; callers keep control
  of drivers and connection pools through `*sql.DB` and per-call options.
- `vdfa` owns DFA word-tree matching, stop-rune filtering, dense/greedy match
  modes, found-word metadata, and text replacement. Generic string helpers
  should not absorb dictionary-matching logic.
- `vid` owns generated identifiers such as UUID, Snowflake, ObjectId, and
  NanoId; `vident` owns legal identity numbers and regional card parsing such
  as mainland China ID cards and Hong Kong/Macau/Taiwan card numbers.
- `vcodec` owns encoding/decoding algorithms such as Base64 and Hex; `vurl`
  owns URL escaping, URL/URI parsing, normalization, resource open/size checks,
  and scheme semantics.
- `vjson` owns JSON objects, arrays, paths, and lightweight XML adapters;
  `vxml` owns XML parsing, tree navigation, formatting, namespace handling, and
  XML-specific map/bean conversion.
- `vbean` owns direct struct/map property mapping, copy-properties, tag/alias
  matching, and weak type conversion without serializing through JSON.
- `vobj` is a convenience object-level facade. New domain logic should still be
  implemented first in clear packages such as `vstr`, `vslice`, `vmap`, or `vref`,
  then wrapped by `vobj` only when a broad object helper is useful.

Database helpers belong to `internal/db` and are exposed through `vdb`; DFA text
matching belongs to `internal/dfa` and is exposed through `vdfa`; office-document
helpers belong to `internal/poi` and are exposed through `vpoi`. Cross-domain
input validators belong to `internal/validator` and are exposed through
`vform`; domain-specific parsing and richer operations still stay in their
domain packages such as `vident`, `vnet`, and `vurl`.

### Error contract

The root package `knifer` owns the cross-cutting error contract: the `ErrCode`
classifier (`knifer.ErrCodeInvalidInput`, `ErrCodeNotFound`, `ErrCodeTimeout`,
…), the unified `knifer.Error` type, the `CodeCarrier` interface, the `CodeOf`
extractor, and the `NewError` / `WrapError` / `Errorf` constructors.
Subpackages that opt in return `*knifer.Error` or add code-aware matching to
their existing error types/sentinels, so callers can match or extract by code
while keeping the chain:

```go
if errors.Is(err, knifer.ErrCodeInvalidInput) { /* ... */ }
if code, ok := knifer.CodeOf(err); ok { /* ... */ }
```

`vcrypto` is a reference integration: validation errors match both
`knifer.ErrCodeInvalidInput` and the existing `vcrypto.ErrInvalidKey` /
`ErrInvalidIV` / `ErrInvalidCipherText` sentinels.

The `vjwt`, `vjson`, `vcron`, `vjob`, `vpoi`, `vcodec`, `vdate`, `vbean`,
`vsem`, `verr`, and `vhttp`/`vresty` errors also participate: their errors
match `knifer.ErrCodeInvalidInput` (vjwt, vjson, vcron, vjob, vpoi empty sheet
name, vcodec decode failures, vdate parse failures, vbean mapping/conversion
failures, vsem invalid weights, and invalid HTTP request input),
`knifer.ErrCodeTimeout` (HTTP timeouts/deadlines), `knifer.ErrCodeNotFound`
(vpoi no sheet, vblf missing initialization file), `knifer.ErrCodeUnsupported`
(vsem closed semaphore, HTTP redirect/body limit cases), or
`knifer.ErrCodeInternal` (remaining vhttp/vresty transport/read errors, vskt,
vblf read errors, and recovered panics from verr) while preserving their own
error types, sentinels, and cause chains.

### Security and safety defaults

Security-sensitive helpers expose only the currently recommended public API
surface. `vcrypto` keeps SHA-2 digests, HMAC-SHA-256/384/512, PBKDF2-SHA-256,
AES-GCM, RSA-OAEP encryption, and RSA-PSS signatures. JWT RSA
signing is exposed through RSA-PSS (`JWTAlgPS256`, `JWTAlgPS384`,
`JWTAlgPS512`, `NewRSAPSSSigner`, and `PS256` / `PS384` / `PS512` helpers),
alongside HMAC and ECDSA signers. Unsigned JWT `alg=none` tokens are rejected
and no public none signer is exposed.

Network and IO helpers prefer bounded, explicit behavior:

- TLS helpers create configs with TLS 1.2+ as the minimum version via
  `vnet.CreateTLSConfig()`. HTTP clients accept explicit `*tls.Config` values
  through `WithTLSConfig`; certificate verification is not bypassed by a
  convenience API.
- HTTP and Resty downloads validate automatically discovered filenames before
  joining them under the destination directory, preventing path traversal when
  callers pass a directory target. When the source URL is untrusted, use
  `vhttp.DownloadFileSafe` / `DownloadFileSafeWithOptions` or
  `vresty.DownloadFileSafe` / `DownloadFileSafeWithOptions`; these combine the
  safe request URL policy with the same destination-save validation.
- `vfile` read helpers use `vfile.DefaultMaxBytes` by default. Use
  `vfile.WithMaxBytes(n)` to tighten a read and `vfile.WithUnlimitedRead()` only
  when the caller has already bounded the source elsewhere.
- `vconf` local and remote loads use `vconf.DefaultMaxBytes` unless
  `LoadOptions.MaxBytes` is set. A negative value explicitly disables the
  config read limit.
- Use `vconf.LoadRemoteSafe` or `LoadRemoteSafeWithOptions` for remote
  configuration from untrusted or user-controlled URLs. The safe variant only
  accepts HTTP(S), rejects localhost/private/link-local/unspecified targets by
  default, and validates redirect targets with the same policy. Set
  `LoadOptions.RemoteAllowedHosts` when a private test server or known internal
  host is intentionally allowed.
- Use `vurl.OpenSafe`, `OpenSafeWithOptions`, `ContentLengthSafe`, or
  `ContentLengthSafeWithOptions` when opening remote resources from untrusted
  input. Safe resource helpers default to HTTP(S), reject local files and plain
  filesystem paths, reject private network targets, check HTTP status, apply a
  timeout, and re-check redirects. Use `WithAllowedHosts` to pin trusted hosts;
  host allowlists narrow the accepted host names but do not bypass private-host
  rejection. Only relax `WithRejectPrivateHosts` or `WithAllowLocalFiles` when
  the caller has already established a narrower trust boundary.
- `vzip` extraction and decompression helpers are bounded by default to reduce
  zip-bomb risk. ZIP entry names are cleaned and checked before writing, and
  extraction resolves destination parents with `filepath.EvalSymlinks` to reject
  entries that would escape through a symlink. Use `vzip.WithEvalSymlinks` only
  when tests or virtual filesystems need to replace that resolver. Use
  `vzip.WithMaxBytes(n)` or `UnzipToLimit` / `UnzipReaderToLimit` for a stricter
  budget; pass a negative max-byte value only when another layer already
  enforces a trusted size limit.
- Bloom filter constructors ending in `E`, such as
  `vblf.NewBitMapBloomFilterE`, `vblf.NewBitSetBloomFilterE`, and
  `vblf.NewFuncFilterE`, return validation errors for invalid sizes or hash
  configuration instead of panicking. The non-`E` constructors are retained for
  compatibility with existing callers.
- `vdb` condition builders validate operators against an allowlist; prefer
  helpers such as `Eq`, `Like`, `In`, `Between`, `IsNull`, and `IsNotNull`
  instead of interpolating raw SQL fragments.
- `vskt.AioSession` serializes reads that share the session buffer and keeps
  buffers available during close callbacks, so lifecycle hooks can inspect the
  last received data safely.
- JWT `alg=none` is always rejected; production code should use HMAC, RSA-PSS,
  or ECDSA signers.

## ✅ Recommended API entry points

Use these APIs for new code. Request helpers that can fail return errors
explicitly instead of swallowing failures.

| Scenario | Recommended API |
| --- | --- |
| Build a trusted standard-library HTTP request | `vhttp.Get`, `vhttp.Post`, `vhttp.NewRequest` |
| Read a trusted HTTP response body and handle errors | `vhttp.GetStringE`, `vhttp.PostJSONE`, `vhttp.DownloadBytesE` |
| Access user-controlled or otherwise untrusted HTTP(S) URLs | `vhttp.GetStringSafeE`, `vhttp.PostJSONSafeE`, `vhttp.DownloadBytesSafeE` |
| Use the Resty-backed HTTP facade | `vresty.Get`, `vresty.Post`, `vresty.GetStringE`, `vresty.PostJSONE` |
| Access untrusted URLs through Resty | `vresty.GetStringSafeE`, `vresty.PostJSONSafeE`, `vresty.DownloadBytesSafeE` |
| Download a user-controlled URL to a file | `vhttp.DownloadFileSafe` or `vresty.DownloadFileSafe` |
| Generate bytes for secrets, tokens, keys, nonces, or salts | `vrand.SecureBytes` |
| Create an LRU cache | `vcache.NewLRU` or `vcache.NewLRUWithTimeout` |
| Parse a cron expression | `vcron.NewPattern` or `vcron.MustNewPattern` |
| Load remote configuration from a trust boundary | `vconf.LoadRemoteSafe` or `vconf.LoadRemoteSafeWithOptions` |

## 📝 Quickstart documents

README keeps module navigation only. Per-package Quickstart examples live in the linked documents from the module matrix above, so examples can stay focused and easy to maintain by domain.

## 📖 Doc

- Root package documentation: `doc.go`
- Public APIs: `doc.go` and facade files in each `v*` subpackage
- Quickstart examples: linked `docs/doc/*.md` files from the module matrix
- Exported API snapshot: `docs/api/exports.txt`
- AI-oriented project map: `llms.txt`
- Online documentation: [pkg.go.dev/github.com/imajinyun/go-knifer](https://pkg.go.dev/github.com/imajinyun/go-knifer)

## 📦 Download & Build

Clone the source code:

```bash
git clone https://github.com/imajinyun/go-knifer.git
cd go-knifer
```

Run tests:

```bash
make test
```

Run the CI test-job gates locally. This verifies modules, vet, tidy/diff
cleanliness, architecture rules, race/shuffle tests, coverage gates, and the
exported API snapshot:

```bash
make ci-test
```

Run the same local safety checks used by CI before opening a PR:

```bash
make check
```

`make check` includes the `ci-test` class of checks plus `golangci-lint` and
`govulncheck`.

Refresh the API snapshot after an intentional exported API change:

```bash
UPDATE_API=1 make api-check
```

GitHub Actions reuses the Makefile targets for module verification, vet, tidy
checks, diff cleanliness, architecture checks, race/shuffle tests, coverage
gates, and API compatibility checks. It also runs golangci-lint, govulncheck,
and CodeQL. Dependabot is configured for Go modules and GitHub Actions updates.

Format code:

```bash
gofmt -w .
```

## 🛡️ Governance

- Security reports: see [SECURITY.md](./SECURITY.md). Please do not disclose
  suspected vulnerabilities in public issues.
- Release notes: see [CHANGELOG.md](./CHANGELOG.md). User-visible changes should
  be recorded before tagging a release.
- Coverage gate: CI enforces the repository baseline with
  `bash bin/check_coverage.sh coverage.out`. The current repository threshold is
  75.2%, with package gates for security-sensitive facades such as `vhttp`,
  `vresty`, `vconf`, `vzip`, `vcrypto`, `vurl`, and `vfile`, plus the core HTTP
  implementation packages. Raise `COVERAGE_THRESHOLD` or
  `PACKAGE_COVERAGE_THRESHOLDS` only after adding tests that support the new
  gate.
- API gate: `make api-check` compares exported symbols against
  `docs/api/exports.txt`. Commit the refreshed snapshot only for intentional API
  changes.
- Stability gate: use `make check` locally before pushing so vet, architecture,
  race/shuffle tests, coverage, API compatibility, lint, and vulnerability
  checks stay aligned with CI.

## 🤝 Provide feedback or suggestions on bugs

If you find a bug or want to request a new utility, please open a GitHub Issue. It is recommended to include:

- Go version and operating system;
- `go-knifer` version or commit;
- Minimal reproducible code;
- Expected behavior and actual behavior;
- Related error logs or test output.

## ✅ Principles of PR (pull request)

Pull requests are welcome. To keep the toolkit stable, please follow these principles where possible:

1. Add new capabilities to the appropriate `internal/*` implementation package first, then expose public APIs from the corresponding `v*` package;
2. Add necessary comments for new or modified public APIs;
3. Add unit tests for core logic and run `go test ./...` before submitting;
4. Keep code formatted with `gofmt`;
5. Avoid unnecessary third-party dependencies and prefer the standard library when possible.

## ⭐ Star go-knifer

If this project helps you reduce repeated code, please consider giving it a Star. Your feedback and contributions will help make it a sharper Go utility toolkit.
