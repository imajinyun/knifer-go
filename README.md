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

Before, calculating a SHA-256 digest often meant writing repetitive boilerplate in business code:

```go
sum := sha256.Sum256([]byte("hello"))
text := hex.EncodeToString(sum[:])
```

Now, with `go-knifer`, you can call a utility function directly:

```go
text := vcrypto.SHA256Hex("hello")
```

This style of utility wrapping reduces repeated code, avoids hidden risks from copy-paste snippets, and keeps the same scenarios represented by consistent APIs across a team.

## 🧭 Find by scenario

Not sure which package to import? Start from what you want to do:

| I want to… | Use |
| --- | --- |
| Trim, split, case-convert, Unicode-escape, Ant-match, compare text similarity, or check blank strings | `vstr` |
| Filter / map / dedup / paginate a slice | `vslice` |
| Create, query, transform, merge, diff, or sort maps | `vmap` |
| Loosely convert `any` to int/float/bool/string | `vconv` |
| Precise arithmetic, rounding, or evaluate an expression | `vnum` |
| SHA/HMAC, AES-GCM/RSA-PSS, sign parameters | `vcrypto` |
| Non-cryptographic hashes (FNV, BKDR, …) | `vhash` |
| Encode/parse URLs, build/parse query strings, or open untrusted HTTP(S) resources safely | `vurl` |
| Base64 / Hex encode-decode | `vcodec` |
| Read/write CSV records, maps, or structs | `vcsv` |
| Produce image thumbnails, convert PNG/JPEG/GIF, read metadata, or generate graphical captchas | `vimg` |
| Build/parse JSON, path get/put, JSON↔XML | `vjson` |
| Load local or remote configuration, including SSRF-checked remote config | `vconf` |
| Parse, build, or navigate XML | `vxml` |
| Generate UUID / Snowflake / NanoId | `vid` |
| Validate or parse ID-card numbers | `vident` |
| Read/write files, paths, copy, mkdir | `vfile` |
| Format/parse dates, offsets, day ranges | `vdate` |
| Send HTTP requests (standard library) | `vhttp` |
| Send HTTP requests (Resty-based) | `vresty` |
| Validate form/input data such as email/mobile/IP | `vform` |
| Mask sensitive data | `vmask` |
| Score and classify password strength locally | `vpass` |
| JWT sign/verify | `vjwt` |
| Schedule cron tasks | `vcron` |
| Cache with FIFO/LRU/LFU/TTL | `vcache` |

For the full list, see the module matrix below.

## 🧩 Module

The project follows an “internal implementation + public facade” layout: `internal/*` contains concrete implementations, while `v*` packages expose stable public APIs.

| Module | Import path | Description |
| --- | --- | --- |
| `vstr` | `github.com/imajinyun/go-knifer/vstr` | String and text helpers: blank/empty checks, trimming, splitting, substring helpers, formatting, provider-backed emoji helpers, naming conversion, defaults, Unicode escaping/unescaping, Ant-style path matching, rune-set Jaccard similarity, rune n-gram similarity, SimHash, 64-bit Hamming distance, HTML escaping, and rune checks (blank, letter, digit, ASCII, letter-or-digit). |
| `vslice` | `github.com/imajinyun/go-knifer/vslice` | Slice helpers: contains/index, reverse, distinct, join, filter/map, sub-slice, concat, set-like operations, and paging. |
| `vmap` | `github.com/imajinyun/go-knifer/vmap` | Map helpers: construction, empty checks, contains/get/find, keys/values and sorted views, map/filter/reject/partition, reduce/group/count, inverse, merge/merge-with-resolver, intersect/diff/symmetric diff, pick/omit, update/clone, and equality checks. |
| `vconv` | `github.com/imajinyun/go-knifer/vconv` | Permissive type conversion: string, int, int64, float64, bool, bytes, and default-value variants. |
| `vdate` | `github.com/imajinyun/go-knifer/vdate` | Date/time helpers: common layouts, parse/format, begin/end of day/month/year, offsets, and comparisons. |
| `vfile` | `github.com/imajinyun/go-knifer/vfile` | File and IO helpers: read/write/copy, lines, mkdir/touch/delete, filename helpers, quiet close, and provider-backed file-system operations. |
| `vcodec` | `github.com/imajinyun/go-knifer/vcodec` | Encoding helpers: Base64, URL-safe Base64, raw URL-safe Base64, custom Base64 encoding providers, and Hex. |
| `vcsv` | `github.com/imajinyun/go-knifer/vcsv` | CSV helpers: reader/writer options for delimiters, comments, field counts, lazy quotes, trimming, CRLF output, records-to-map conversion, map writing, struct tag export, and record callbacks. |
| `vimg` | `github.com/imajinyun/go-knifer/vimg` | Image helpers: proportional thumbnails, format conversion between PNG/JPEG/GIF, metadata introspection (width/height/format), and image captcha generation through line/circle/shear/GIF captcha types. |
| `vurl` | `github.com/imajinyun/go-knifer/vurl` | URL and URI helpers: parse, normalize, resolve relative URLs, query encode/decode, URL/path/fragment percent encoding with injectable query/path escape providers, URL building, Data URI building, scheme checks, file URL conversion, resource open/size helpers, and SSRF-oriented `OpenSafe` / `ContentLengthSafe` variants. |
| `vnet` | `github.com/imajinyun/go-knifer/vnet` | Network helpers: IPv4/IPv6 conversion, CIDR/range/mask utilities with injectable IP/CIDR/int parsers, local ports, host/interface/MAC lookup, TLS config, address/dial/ping provider options, and multipart form helpers. |
| `vobj` | `github.com/imajinyun/go-knifer/vobj` | Object helpers: nil/empty checks, equality, defaults, clone/serialization, comparison, type inspection, and container utilities. |
| `vver` | `github.com/imajinyun/go-knifer/vver` | Version helpers: version comparison, greater/less predicates, expression matching, inclusive ranges, and custom expression delimiters. |
| `vref` | `github.com/imajinyun/go-knifer/vref` | Reflection helpers: field lookup and mutation, method discovery and invocation, constructor-style function calls, nil-safe type/value utilities, object-level type predicates (`IsFunction`, `IsIteratee`, `IsCollection`, `IsSlice`, `IsArray`, `IsMap`), method classification, and explicit unsafe/unexported field-access options. |
| `vbean` | `github.com/imajinyun/go-knifer/vbean` | Bean/struct mapping helpers: struct/map conversion, copy properties, tag and alias matching, ignore-empty/zero options, and weak type conversion. |
| `vzip` | `github.com/imajinyun/go-knifer/vzip` | ZIP, gzip, and zlib helpers: archive creation/extraction, entry lookup, archive traversal, append, in-memory entries, stream compression, provider-backed archive file operations, bounded extraction/decompression defaults, path traversal checks, and symlink escape checks during extraction. |
| `vpoi` | `github.com/imajinyun/go-knifer/vpoi` | Office document helpers: lightweight Excel XLSX sheet listing, row reading/writing, multi-sheet writing, in-memory workbook creation, and injectable workbook/file-system providers. |
| `vmask` | `github.com/imajinyun/go-knifer/vmask` | Masking helpers: mask names, IDs, phones, addresses, email, passwords, license plates, bank cards, IPs, passports, and credit codes. |
| `vpass` | `github.com/imajinyun/go-knifer/vpass` | Password helpers: deterministic local scoring, strength buckets, strong/weak predicates, character-class signals, repeated/sequential-run detection, and a small common-weak-password blocklist. |
| `vnum` | `github.com/imajinyun/go-knifer/vnum` | Numeric helpers: precise arithmetic, generic sum/average/min/max/abs helpers, rounding modes, provider-backed parsing/formatting, number checks, random unique numbers, ranges, factorial/combinations, gcd/lcm, binary conversion, comparison, byte conversion, expression calculation, and odd/even checks. |
| `vrand` | `github.com/imajinyun/go-knifer/vrand` | Random helpers: integers, floats, booleans, bytes, strings, numeric strings, random element selection, deterministic seeding, and resettable package-level pseudo-random source providers. |
| `vid` | `github.com/imajinyun/go-knifer/vid` | ID helpers: random/simple/fast UUIDs, MongoDB-style ObjectId, Snowflake generators and singleton next-id helpers, worker/datacenter id derivation, NanoId generation, fallback random sources, isolated Snowflake creation, and resettable fallback PRNG providers/seeds. |
| `vident` | `github.com/imajinyun/go-knifer/vident` | Identity helpers: mainland China ID card 15/18-digit conversion, validation, check code, birthday/age/gender extraction with parsing options, province/city/district code parsing, masking, and Hong Kong/Macau/Taiwan card validation. |
| `vhash` | `github.com/imajinyun/go-knifer/vhash` | Non-cryptographic hash helpers: additive, FNV, injectable 32-bit hash providers, and a set of classic string hashes (RS, JS, PJW, ELF, BKDR, SDBM, DJB, AP, HF, HFIP, TianL, Java default). |
| `vform` | `github.com/imajinyun/go-knifer/vform` | Form and input validation helpers: email, mobile, URL, IPv4/IPv6, ID card, Chinese text, number string checks, and per-call matcher providers for rule-sensitive checks. |
| `vtpl` | `github.com/imajinyun/go-knifer/vtpl` | Go html/template rendering helpers with per-call template name, FuncMap, delimiter, factory, and executor options. |
| `vregex` | `github.com/imajinyun/go-knifer/vregex` | Regular-expression helpers: matching, group extraction, named groups, deletion, counting, index lookup, template/function replacement, escaping, and per-call compiler / DOTALL options. |
| `vbool` | `github.com/imajinyun/go-knifer/vbool` | Boolean helpers: negate, bool-to-int, all/any checks. |
| `vblf` | `github.com/imajinyun/go-knifer/vblf` | Bloom filters: bitmap/bitset/filter abstractions, multiple string hash algorithms, option-based constructors, `E` constructors that return validation errors instead of panicking, and provider-backed file initialization. |
| `vcache` | `github.com/imajinyun/go-knifer/vcache` | Generic caches: FIFO, LFU, LRU, Timed, Weak, and NoCache; supports TTL, clocks, removal listeners, lazy loading, ticker/runner providers, and weak-cache finalizer providers. Removal listeners run outside cache locks, so callbacks can safely re-enter the same cache. |
| `vcron` | `github.com/imajinyun/go-knifer/vcron` | Cron expression parsing and task scheduling, including default/custom schedulers, configurable cron options, ID random-reader/clock/sleeper/runner providers, isolated per-call default scheduler overrides, running-task metrics, `Wait`, and graceful `Shutdown(ctx)`. |
| `vcrypto` | `github.com/imajinyun/go-knifer/vcrypto` | Cryptography and digests: SHA-2, provider-backed digest helpers, HMAC, PBKDF2-SHA256, parameter signing, random bytes, AES-GCM with nonce/tag/block-factory options, RSA OAEP/PSS plus configurable data-signing options, PEM, and X.509 certificate helpers. |
| `vdb` | `github.com/imajinyun/go-knifer/vdb` | Database helpers built on database/sql: SQL execution, named parameters, entities, conditions, query builders, transactions, pagination, lightweight metadata lookup, and injectable `sql.Open` providers. |
| `vdfa` | `github.com/imajinyun/go-knifer/vdfa` | DFA word-tree matching: stop-rune filtering, first/all matches, dense and greedy match modes, found-word metadata, package-level matcher helpers, isolated matcher options, JSON marshal/unmarshal providers for `Any` helpers, text replacement, and resettable async runner providers for package-level initialization. |
| `vhttp` | `github.com/imajinyun/go-knifer/vhttp` | Chainable HTTP client, isolated/global-config request construction, create/get/post `WithOptions` helpers, explicit-error `E` shortcuts, code-classified HTTP errors, provider-backed transports/request factories/multipart writers/download saves, safe file downloads, BasicAuth, User-Agent parsing, provider-backed HTML cleaning/filtering, resettable transport/server starters, async server runner options, and simple server helpers. |
| `vresty` | `github.com/imajinyun/go-knifer/vresty` | Resty v3 based HTTP facade: chainable requests, JSON/form/multipart bodies, isolated/global-config request construction, create/get/post `WithOptions` helpers, per-request client factories, resettable default Resty client providers, downloads and safe file downloads, and lightweight response helpers. |
| `vjson` | `github.com/imajinyun/go-knifer/vjson` | Ordered JSON objects/arrays, JSON parsing and formatting, path-based get/put, provider-backed marshal/unmarshal, injectable scalar parse/format functions, configurable object/array/bean/list conversion, and XML/JSON conversion with parser/writer options. |
| `vxml` | `github.com/imajinyun/go-knifer/vxml` | XML helpers: parse/read/write/format, tree navigation, simple XPath-style lookup, escaping, map/bean conversion with parser/codec/scalar parser options, transform options, and namespace utilities. |
| `vjwt` | `github.com/imajinyun/go-knifer/vjwt` | JWT creation, parsing, signing, verification, and time-claim validation; supports HMAC, RSA-PSS, ECDSA, rejects unsigned `alg=none` tokens, and provides JSON marshal/unmarshal options. |
| `vlog` | `github.com/imajinyun/go-knifer/vlog` | Logging facade: console/color console loggers, injectable color factories, log levels, global logger, static logging functions, per-call logger options, and isolated logger creation. |
| `verr` | `github.com/imajinyun/go-knifer/verr` | Error helpers: panic recovery, error aggregation, multierror matching, collector construction options, stack capture/formatting, resettable log/stack caches, injectable logging/stack/exit/timer/runner providers, isolated logrus creation, and optional logrus/Sentry integration. |
| `vconf` | `github.com/imajinyun/go-knifer/vconf` | Grouped configuration reader for setting/properties-style text, a simple YAML subset, and TOML parsing, with typed getters, schema validation, profile/remote/file loading options, SSRF-checked `LoadRemoteSafe`, environment expansion providers, watch ticker/runner providers, bounded reads, read-only snapshot guidance, and deep-copy `Clone` support. |
| `vset` | `github.com/imajinyun/go-knifer/vset` | Generic and typed set utilities with add/remove/contains, set operations, and JSON/YAML encoding helpers. |
| `vjob` | `github.com/imajinyun/go-knifer/vjob` | Sliceable job execution: separate job data from scheduling options, typed slice/map adapters, context cancellation, and serialized merge callbacks; no generic type-alias experiment is required. |
| `vsem` | `github.com/imajinyun/go-knifer/vsem` | Weighted, context-aware counting semaphore with FIFO fairness, try-acquire, close notifications, and in-use metrics. |
| `vskt` | `github.com/imajinyun/go-knifer/vskt` | TCP socket utilities: plain connections, NIO/AIO server/client helpers, protocol encoder/decoder interfaces, and configurable thread-pool/listener/connection/runner/IP-parser providers. |
| `vsys` | `github.com/imajinyun/go-knifer/vsys` | System and runtime information: host, OS, user, Go runtime, process memory, goroutines, environment variables, resettable info cache, and injectable env/command/runtime providers. |

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

## 🚀 Install

Go 1.25 or later is required.

```bash
go get github.com/imajinyun/go-knifer
```

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

Go will resolve the module according to the subpackages you actually import, for example:

```go
import (
  "github.com/imajinyun/go-knifer/vstr"
  "github.com/imajinyun/go-knifer/vhttp"
)
```

## 📝 Quick start

### Domain utilities and JSON

```go
package main

import (
  "fmt"

  "github.com/imajinyun/go-knifer/vid"
  "github.com/imajinyun/go-knifer/vjson"
  "github.com/imajinyun/go-knifer/vstr"
)

func main() {
  name := vstr.DefaultIfBlank("", "go-knifer")

  obj := vjson.NewObject().
    Set("id", vid.FastUUID()).
    Set("name", name).
    Set("tags", []string{"go", "tool"})

  fmt.Println(obj.GetString("name"))
  fmt.Println(obj.ToStringPretty())
}
```

### Per-call providers and `WithOptions` variants

Runtime-sensitive helpers keep their simple APIs, and expose `WithOptions`
variants for callers that need deterministic tests, aliases, stricter escaping,
or controlled dependencies. Nil providers fall back to the standard-library
implementation used by the plain API.

```go
package main

import (
  "fmt"
  "net"
  "net/url"
  "strconv"
  "strings"

  "github.com/imajinyun/go-knifer/vjson"
  "github.com/imajinyun/go-knifer/vnet"
  "github.com/imajinyun/go-knifer/vnum"
  "github.com/imajinyun/go-knifer/vurl"
  "github.com/imajinyun/go-knifer/vxml"
)

func main() {
  parseIP := func(s string) net.IP {
    if s == "loopback" {
      return net.ParseIP("127.0.0.1")
    }
    return net.ParseIP(s)
  }
  ipLong, _ := vnet.IPv4ToLongWithOptions("loopback", vnet.WithIPParser(parseIP))
  inRange := vnet.IsInRangeWithOptions("loopback", "127.0.0.0/8", vnet.WithIPParser(parseIP))

  strictQuery := vurl.EncodeQueryWithOptions("a b+c", vurl.WithQueryEscapeFunc(func(s string) string {
    return strings.ReplaceAll(url.QueryEscape(s), "+", "%20")
  }))

  total, _ := vnum.CalculateWithOptions("5 + 2", vnum.WithParseFloatFunc(func(s string, bitSize int) (float64, error) {
    if s == "5" {
      return 5, nil
    }
    return strconv.ParseFloat(s, bitSize)
  }))

  cfg := vjson.NewConfig()
  cfg.ParseIntFunc = func(s string, base, bitSize int) (int64, error) {
    if s == "answer" {
      return 42, nil
    }
    return strconv.ParseInt(s, base, bitSize)
  }
  obj := vjson.NewObjectWithConfig(cfg).Set("count", "answer")

  xmlMap, _ := vxml.XMLToMapWithOptions(`<root><count>answer</count></root>`,
    vxml.WithScalarIntParser(cfg.ParseIntFunc))

  fmt.Println(ipLong, inRange, strictQuery, total)
  fmt.Println(obj.GetInt64("count"), xmlMap["root"])
}
```

The same pattern is used by socket helpers as configuration options, for example
`vskt.NewSocketConfigWithOptions(vskt.WithSocketIPParser(parseIP))` or
`vskt.NewNioClientWithOptions(host, port, vskt.WithSocketIPParser(parseIP))`.

### Form and input validation helpers

`vform` provides a short public entry point for common form and input checks. It
keeps the frequently used boolean validators together while delegating the
actual domain logic to the corresponding internal packages.

```go
package main

import (
  "fmt"

  "github.com/imajinyun/go-knifer/vform"
)

func main() {
  fmt.Println(vform.IsEmail("a@b.com"))
  fmt.Println(vform.IsMobile("13812345678"))
  fmt.Println(vform.IsURL("https://example.com"))
  fmt.Println(vform.IsIPv4("127.0.0.1"))
  fmt.Println(vform.IsIPv6("2001:db8::1"))
  fmt.Println(vform.IsIDCard("11010519491231002X"))
  fmt.Println(vform.IsChinese("你好"))
  fmt.Println(vform.IsNumberStr("-3.14"))
}
```

### LRU cache and lazy loading

```go
package main

import (
  "fmt"
  "time"

  "github.com/imajinyun/go-knifer/vcache"
)

func main() {
  c := vcache.NewLRUWithTimeout[string, int](3, 5*time.Minute)
  c.SetListener(vcache.CacheListenerFunc[string, int](func(key string, value int) {
    // Listener callbacks run after internal locks are released, so re-entering
    // the same cache from a callback is safe.
    fmt.Println("removed", key, value, "remaining", c.Size())
  }))

  c.Put("answer", 42)

  value, ok := c.Get("answer")
  fmt.Println(value, ok)

  loaded, err := c.GetOrLoad("miss", func() (int, error) {
    return 100, nil
  })
  fmt.Println(loaded, err)
}
```

Removal listeners are called after cache mutations release their internal lock.
This avoids self-deadlocks and lets listener code safely call back into the same
cache for cleanup, metrics, reload, or follow-up writes.

### Configuration snapshots

`vconf` objects are intentionally simple mutable maps while being built or
loaded. After a config is shared with application code, treat it as an immutable
snapshot. For runtime updates, clone or reload, mutate the new copy, then publish
the new pointer. Local and remote config loading is bounded by
`vconf.DefaultMaxBytes` by default; pass `LoadOptions{MaxBytes: n}` to use a
stricter limit, or a negative value only when the source is already bounded by
another layer.

```go
package main

import (
  "fmt"
  "sync/atomic"

  "github.com/imajinyun/go-knifer/vconf"
)

func main() {
  cfg, _ := vconf.Parse("app.name=go-knifer\n")

  loaded, _ := vconf.LoadRemoteSafeWithOptions("https://example.com/app.yaml", vconf.LoadOptions{
    MaxBytes:            1 << 20,
    RemoteAllowedHosts: []string{"example.com"},
  })

  var current atomic.Pointer[vconf.Conf]
  current.Store(cfg)

  next := current.Load().Clone()
  next.Set("app.name", "go-knifer-next")
  current.Store(next)

  _ = loaded
  fmt.Println(current.Load().Get("app.name"))
}
```

`vconf` also supports schema validation and struct binding for typed
configuration contracts. Use them after loading and before publishing a config
snapshot. Prefer `LoadRemoteSafe` for remote configuration unless the URL is a
trusted constant and another layer already enforces host and redirect policy.

```go
schema := vconf.Schema{Fields: []vconf.FieldRule{
  {Group: "server", Key: "port", Type: vconf.TypeInt, Required: true},
}}
if err := cfg.ValidateSchema(schema); err != nil {
  panic(err)
}
```

### Chainable HTTP request

Use the chainable API when you need full request/response control. For one-line
helpers, prefer the `E` suffix variants in new code when errors matter:
`GetStringE`, `GetWithTimeoutE`, `GetWithParamsE`, `PostFormE`, `PostJSONE`,
`PostStringE`, `DownloadStringE`, and `DownloadBytesE` return `(value, error)`
instead of silently converting failures to empty values.

```go
package main

import (
  "errors"
  "fmt"
  "time"

  knifer "github.com/imajinyun/go-knifer"
  "github.com/imajinyun/go-knifer/vhttp"
)

func main() {
  resp := vhttp.Get("https://example.com",
    vhttp.WithTimeout(3*time.Second),
    vhttp.WithHeader("X-Client", "go-knifer"),
    vhttp.WithFollowRedirects(true),
  ).
    Query("lang", "go").
    Execute()

  if resp.Err() != nil {
    panic(resp.Err())
  }

  fmt.Println(resp.Status())
  fmt.Println(resp.ContentType())
  fmt.Println(resp.Body())

  body, err := vhttp.GetStringEWithOptions("https://example.com/ping",
    vhttp.WithTimeout(500*time.Millisecond),
  )
  if errors.Is(err, knifer.ErrCodeTimeout) {
    fmt.Println("request timed out")
    return
  }
  if err != nil {
    panic(err)
  }
  fmt.Println(body)
}
```

Request defaults can still be configured globally when needed, but new code
should prefer per-call options to keep request behavior explicit and avoid
cross-request state coupling. Available options include `WithTimeout`,
`WithHeader`, `WithHeaders`, `WithFollowRedirects`, `WithMaxRedirects`,
`WithTransport`, `WithClient`, `WithCookieJar`, and
`WithUserAgent`. TLS behavior is configured by passing an explicit
`*tls.Config` with `WithTLSConfig`; the facade does not expose a helper that
disables certificate verification.

HTTP errors are code-classified for routing and retry logic: malformed URLs and
request construction problems match `knifer.ErrCodeInvalidInput`, timeouts match
`knifer.ErrCodeTimeout`, redirect/body limit cases match
`knifer.ErrCodeUnsupported`, and remaining transport/read failures match
`knifer.ErrCodeInternal`. Use `vhttp.NewErrorWithCode` or
`vhttp.ErrorfWithCode` when wrapping custom HTTP-layer failures with an explicit
code.

### Resty v3 HTTP facade

`vresty` provides a thin, chainable facade over `resty.dev/v3`. It keeps the
public API lightweight while supporting common HTTP operations such as query
parameters, headers, cookies, Basic/Bearer auth, JSON/form bodies, multipart
uploads, per-call options, TLS configuration, redirect control, and
downloads.

```go
package main

import (
  "fmt"
  "time"

  "github.com/imajinyun/go-knifer/vresty"
)

func main() {
  resp := vresty.Post("https://api.example.com/users",
    vresty.WithTimeout(3*time.Second),
    vresty.WithHeader("X-App", "go-knifer"),
    vresty.WithUserAgent("go-knifer-demo/1.0"),
  ).
    Query("source", "demo").
    BearerAuth("token").
    BodyJSON(`{"name":"go-knifer"}`).
    Execute()

  if resp.Err() != nil {
    panic(resp.Err())
  }
  if !resp.IsOK() {
    panic(fmt.Sprintf("unexpected status: %d", resp.Status()))
  }

  fmt.Println(resp.ContentType())
  fmt.Println(resp.Body())
}
```

Like `vhttp`, `vresty` supports construction-time request options so each call
can override defaults independently: `WithTimeout`, `WithHeader`, `WithHeaders`,
`WithFollowRedirects`, `WithMaxRedirects`, `WithTLSConfig`, `WithRestyClient`,
`WithUserAgent`, and `WithCookieDisabled`.

Shortcuts are available for simple cases and downloads:

```go
body, err := vresty.GetStringE("https://example.com")
if err != nil {
  panic(err)
}
jsonBody, err := vresty.PostJSONE("https://api.example.com/events", `{"event":"created"}`)
if err != nil {
  panic(err)
}
n, err := vresty.DownloadFileSafe("https://example.com/report.csv", "./downloads")
_, _, _ = body, jsonBody, n
_ = err
```

When a download destination is a directory, the filename inferred from the
response is validated before being joined under that directory. Provide an
explicit file path when you need a fixed output name.

### Cron scheduling and shutdown

`vcron` supports both package-level helpers and explicit scheduler instances.
For long-running services, prefer an explicit scheduler so lifecycle is clear:
`RunningCount` reports in-flight executions, `Wait` blocks until they finish, and
`Shutdown(ctx)` stops the timer loop and waits up to the provided context.

```go
package main

import (
  "context"
  "fmt"
  "time"

  "github.com/imajinyun/go-knifer/vcron"
)

func main() {
  s := vcron.NewSchedulerWithOptions(vcron.WithMatchSecond(true))
  _, _ = s.ScheduleFunc("* * * * * *", func() {
    time.Sleep(100 * time.Millisecond)
    fmt.Println("tick")
  })

  if err := s.Start(); err != nil {
    panic(err)
  }

  time.Sleep(1500 * time.Millisecond)
  fmt.Println("running:", s.RunningCount())

  ctx, cancel := context.WithTimeout(context.Background(), time.Second)
  defer cancel()
  if err := s.Shutdown(ctx, true); err != nil {
    panic(err)
  }
}
```

### URL and URI helpers

`vurl` centralizes URL parsing, normalization, query string handling, percent
encoding, URL building, scheme checks, Data URI building, and file URL
conversion. For user-provided resource locations, use the safe resource helpers
instead of `Open`: they reject local files, non-HTTP schemes, private network
targets, and unsafe redirects by default.

```go
package main

import (
  "fmt"
  "io"

  "github.com/imajinyun/go-knifer/vurl"
)

func main() {
  normalized := vurl.Normalize(`example.com\docs/a b`, true, true)
  completed, _ := vurl.Complete("https://example.com/base/", "next?id=1")
  query := vurl.BuildQuery(map[string]any{"lang": "go", "page": 1})
  dataURI := vurl.DataURIBase64("text/plain", "aGVsbG8=")
  built := vurl.NewHTTPURLBuilder("example.com").AddPathSegment("a b").AddQuery("q", "go").Build()
  reader, err := vurl.OpenSafeWithOptions("https://example.com/config.yaml",
    vurl.WithAllowedHosts("example.com"),
  )
  if err != nil {
    panic(err)
  }
  defer reader.Close()
  _, _ = io.Copy(io.Discard, reader)

  fmt.Println(normalized)
  fmt.Println(completed)
  fmt.Println(query)
  fmt.Println(vurl.IsWebURL(completed))
  fmt.Println(dataURI)
  fmt.Println(built)
}
```

`OpenWithOptions` and `ContentLengthWithOptions` remain available for trusted
local files or controlled resources. Use `OpenSafeWithOptions` and
`ContentLengthSafeWithOptions` when the location crosses a trust boundary.

### Network and IP helpers

`vnet` provides network helpers for IPv4/IPv6 conversion, CIDR and mask
calculation, IP range expansion, local port probing, host/interface/MAC lookup,
TLS client config creation, and multipart form helpers.

```go
package main

import (
  "fmt"

  "github.com/imajinyun/go-knifer/vnet"
)

func main() {
  ipLong, _ := vnet.IPv4ToLong("127.0.0.1")
  begin, _ := vnet.BeginIP("192.168.1.9", 24)
  end, _ := vnet.EndIP("192.168.1.9", 24)
  tlsConfig := vnet.CreateTLSConfig()

  fmt.Println(ipLong, vnet.LongToIPv4(ipLong))
  fmt.Println(begin, end, vnet.IsInRange("192.168.1.8", "192.168.1.0/24"))
  fmt.Println(vnet.HideIPPart("192.168.1.8"))
  fmt.Println(tlsConfig.MinVersion)
}
```

`CreateTLSConfig` creates a client TLS config with TLS 1.2 or newer as the
minimum version. For custom trust roots, use `NewTLSConfigBuilder` and add root
CA PEM data from bytes, readers, or files.

### Object helpers

`vobj` provides nil-safe object helpers for common data handling: equality,
emptiness checks, default values, clone/serialization helpers, comparison, and
type inspection.

```go
package main

import (
  "fmt"

  "github.com/imajinyun/go-knifer/vobj"
)

type Profile struct {
  Name string
  Tags []string
}

func main() {
  name := "go-knifer"
  profile := Profile{Name: name, Tags: []string{"go", "tool"}}

  cloned := vobj.CloneIfPossible(profile)
  fmt.Println(vobj.Equal(1, int64(1)))
  fmt.Println(vobj.IsEmpty([]string{}))
  fmt.Println(vobj.DefaultIfNil(&name, "default"))
  fmt.Println(vobj.Contains(cloned.Tags, "go"))
  fmt.Println(vobj.TypeName(profile))
}
```

### Map helpers

`vmap` provides generic helpers for common map operations. It returns non-nil
maps for constructors and pure helpers, keeps input maps unmodified unless the
function is explicitly in-place (`Clear`, `Update`), and supports both
last-write-wins merging, custom conflict resolution, in-place merge variants,
and map-to-slice transformations.

```go
package main

import (
  "fmt"

  "github.com/imajinyun/go-knifer/vmap"
)

func main() {
  base := vmap.Of[string, int]("a", 1, "b", 2)
  merged := vmap.Merge(base, map[string]int{"b": 20, "c": 3})
  evens := vmap.FilterValues(merged, func(v int) bool { return v%2 == 0 })
  pairs := vmap.ToSlice(merged, func(k string, v int) string { return fmt.Sprintf("%s=%d", k, v) })
  grouped := vmap.GroupBy([]string{"go", "git", "java"}, func(s string) byte { return s[0] })

  fmt.Println(vmap.SortedKeys(merged))
  fmt.Println(evens)
  fmt.Println(pairs)
  fmt.Println(grouped['g'])
}
```

### Database helpers

`vdb` provides SQL helpers on top of `database/sql`: named parameters,
condition builders, entity-based insert/update/delete, pagination,
transactions, and lightweight metadata lookup. Drivers and connection pools stay
under caller control. Condition helpers validate operators against a fixed
allowlist, so prefer the typed builders (`Eq`, `Ne`, `Gt`, `Gte`, `Lt`, `Lte`,
`Like`, `In`, `Between`, `IsNull`, `IsNotNull`, `AndGroup`, `OrGroup`) instead
of assembling ad-hoc operator strings.

```go
package main

import (
  "context"
  "database/sql"
  "fmt"

  "github.com/imajinyun/go-knifer/vdb"
)

func main() {
  var raw *sql.DB // normally opened by your selected SQL driver
  db := vdb.Use(raw, vdb.WithDialect(vdb.DialectPostgres))

  sqlText, args, _ := vdb.NewBuilder(vdb.WithDialect(vdb.DialectPostgres)).
    Select("id", "name").
    From("users").
    Where(vdb.Eq("status", "active")).
    OrderBy(vdb.Desc("id")).
    Page(vdb.NewPage(1, 20)).
    SQL()

  named, _ := vdb.ParseNamed(
    "select * from users where id = :id",
    map[string]any{"id": 1},
    vdb.DialectPostgres,
  )

  _ = db
  _ = context.Background()
  fmt.Println(sqlText, args, named.SQL, named.Params)
}
```

### Bean and struct mapping

`vbean` copies properties directly between structs and maps without a JSON
round-trip. It supports tag/alias matching, weak type conversion, and per-call
options such as ignoring empty or zero source values.

```go
package main

import (
  "fmt"

  "github.com/imajinyun/go-knifer/vbean"
)

type UserDTO struct {
  Name  string `bean:"name,alias=full_name|displayName"`
  Age   string `bean:"age"`
  Admin string `bean:"admin"`
}

type User struct {
  Name  string `json:"full_name"`
  Age   int    `json:"age"`
  Admin bool   `json:"admin"`
}

func main() {
  src := UserDTO{Name: "alice", Age: "42", Admin: "yes"}

  var dst User
  _ = vbean.CopyProperties(src, &dst, vbean.WithIgnoreEmpty(true))

  m, _ := vbean.ToMap(dst)
  fmt.Println(dst.Age, dst.Admin)
  fmt.Println(m["full_name"])
}
```

### Serialization helpers

`vobj` provides gob-based serialization helpers for byte encoding, typed
deserialization, deep cloning, interface type registration, and optional decoded
object graph validation.

```go
package main

import (
  "fmt"

  "github.com/imajinyun/go-knifer/vobj"
)

type Profile struct {
  Name string
  Tags []string
}

func main() {
  profile := Profile{Name: "go-knifer", Tags: []string{"go", "tool"}}

  data, _ := vobj.Serialize(profile)
  decoded, _ := vobj.DeserializeTo[Profile](data, Profile{})
  cloned := vobj.CloneIfPossible(profile)

  fmt.Println(decoded.Name)
  fmt.Println(cloned.Tags)
}
```

### Version helpers

`vver` provides version comparison and expression matching. Expressions support
comparison operators (`>`, `>=`, `<`, `<=`, `≥`, `≤`), inclusive ranges such as
`1.0.0-1.5.0`, open ranges such as `1.0.0-`, and multiple alternatives with a
custom delimiter.

```go
package main

import (
  "fmt"

  "github.com/imajinyun/go-knifer/vver"
)

func main() {
  fmt.Println(vver.CompareVersion("1.0.0", "1.0.2"))
  fmt.Println(vver.IsGreaterThan("1.13.0", "1.12.1c"))
  fmt.Println(vver.MatchEl("1.0.2", ">=1.0.0;1.2.0"))
  fmt.Println(vver.MatchElWithDelimiter("1.0.2", "<1.0.1,1.0.2-1.1.1", ","))
}
```

### ZIP, gzip, and zlib helpers

`vzip` provides archive creation/extraction, entry lookup, archive traversal,
append operations, in-memory entries, and byte/string compression helpers.
Extraction and decompression are bounded by default; pass `WithMaxBytes` or the
`Limit` helpers to set an explicit budget for the archive you expect.

```go
package main

import (
  "fmt"

  "github.com/imajinyun/go-knifer/vzip"
)

func main() {
  _ = vzip.ZipEntries("demo.zip", vzip.EntryData{Name: "hello.txt", Data: []byte("hello")})
  _ = vzip.UnzipToWithOptions("demo.zip", "./out", vzip.WithMaxBytes(64<<20))
  data, _ := vzip.GetBytes("demo.zip", "hello.txt")
  gz, _ := vzip.GzipString(string(data))
  text, _ := vzip.UnGzipString(gz)

  fmt.Println(text)
}
```

### Masking helpers

`vmask` provides built-in masking rules for common sensitive fields such as names,
identity numbers, phones, addresses, email addresses, passwords, license plates,
bank cards, IP addresses, passports, and credit codes.

```go
package main

import (
  "fmt"

  "github.com/imajinyun/go-knifer/vmask"
)

func main() {
  fmt.Println(vmask.MobilePhone("18049531999"))
  fmt.Println(vmask.Email("duandazhi-jack@gmail.com.cn"))
  fmt.Println(vmask.BankCard("11011111222233333256"))
  fmt.Println(vmask.Masked("PJ1234567", vmask.PassportType))
}
```

### Regular-expression helpers

`vregex` provides safe regular-expression helpers for whole-string matching,
substring lookup, capture groups, named groups, deletion, counting, index lookup,
template/function replacement, and escaping regex metacharacters.

```go
package main

import (
  "fmt"

  "github.com/imajinyun/go-knifer/vregex"
)

func main() {
  text := "date=2026-05-31; score=100"

  fmt.Println(vregex.GetByName(`(?<year>\d{4})-(?<month>\d{2})-(?<day>\d{2})`, text, "year"))
  fmt.Println(vregex.ExtractMulti(`score=(\d+)`, text, "score:$1"))
  fmt.Println(vregex.DelFirst(`\d+`, text))
  fmt.Println(vregex.ReplaceAllFunc(text, `\d+`, func(m vregex.MatchResult) string {
    return "[" + m.Text + "]"
  }))
  fmt.Println(vregex.Escape("a+b(c)"))
}
```

### Digest and JWT

`vcrypto` intentionally exposes the safer cryptographic helpers: SHA-2 digests,
HMAC-SHA-256/384/512, PBKDF2-SHA-256, AES-GCM, RSA-OAEP, and
RSA-PSS. The examples below use AES-GCM for authenticated encryption and HMAC
JWT signing for symmetric service tokens.

```go
package main

import (
  "fmt"
  "time"

  "github.com/imajinyun/go-knifer/vcrypto"
  "github.com/imajinyun/go-knifer/vjwt"
)

func main() {
  fmt.Println(vcrypto.SHA256Hex("hello"))
  fmt.Println(vcrypto.HMACSHA256Hex([]byte("key"), []byte("hello")))

  aesKey := []byte("1234567890123456")
  iv := []byte("abcdefghijklmnop")
  cipherText, err := vcrypto.AESEncryptGCM([]byte("secret message"), aesKey, iv[:12], nil)
  if err != nil {
    panic(err)
  }
  plain, err := vcrypto.AESDecryptGCM(cipherText, aesKey, iv[:12], nil)
  if err != nil {
    panic(err)
  }
  fmt.Println(string(plain))

  key := []byte("secret")
  token, err := vjwt.NewJWT().
    SetSubject("user-1").
    SetPayload("role", "admin").
    SetExpiresAt(time.Now().Add(time.Hour)).
    SetKey(key).
    Sign()
  if err != nil {
    panic(err)
  }

  jwt, err := vjwt.ParseJWT(token)
  if err != nil {
    panic(err)
  }

  fmt.Println(jwt.SetKey(key).Verify())
}
```

### Generic sets

`vset` exposes a generic `Set[T]` plus common typed constructors. The generic
facade is implemented as a regular generic type instead of a generic type alias,
so it works with the default Go toolchain and `go vet` without enabling
`GOEXPERIMENT=aliastypeparams`.

```go
package main

import (
  "encoding/json"
  "fmt"

  "github.com/imajinyun/go-knifer/vset"
)

func main() {
  tags := vset.New("go", "tool")
  tags.Add("sdk")

  other := vset.New("tool", "cli")
  fmt.Println(tags.Contains("go"))
  fmt.Println(tags.Union(other).Members())
  fmt.Println(tags.Intersect(other).Members())

  data, _ := json.Marshal(tags)
  var decoded vset.Set[string]
  _ = json.Unmarshal(data, &decoded)
  fmt.Println(decoded.Equal(tags))

  ids := vset.NewInt(1, 2, 3)
  ids.Remove(2)
  fmt.Println(ids.Members())
}
```

### Sliceable job execution

`vjob` separates the job contract from scheduling options. A job only needs to
implement `Len` and range-based `Run`; `Options` controls batch size and maximum
concurrency. The zero value is valid: `Run` processes the whole job as one serial
shard, while `RunWith` accepts explicit scheduling options. Returned `Merge`
callbacks are replayed serially after shards succeed, which lets each worker
build local results concurrently and merge them safely afterwards. `Batch[T]` is
a facade wrapper around the internal implementation, not a generic type alias,
so `go vet` works without extra experiment flags.

```go
package main

import (
  "context"
  "fmt"
  "sync"

  "github.com/imajinyun/go-knifer/vjob"
)

func main() {
  values := []int{1, 2, 3, 4}
  var (
    mu  sync.Mutex
    sum int
  )

  job := vjob.NewBatch(func(ctx context.Context, batch []int) (vjob.Merge, error) {
    local := 0
    for _, v := range batch {
      local += v
    }
    return func() error {
      mu.Lock()
      defer mu.Unlock()
      sum += local
      return nil
    }, nil
  }, values)

  if err := vjob.RunWith(context.Background(), job, vjob.Options{BatchSize: 2, MaxConcurrency: 2}); err != nil {
    panic(err)
  }
  fmt.Println(sum)
}
```

Reusable jobs can embed `vjob.Options` and pass their own configuration to
`RunWith`:

```go
type UserImportJob struct {
  vjob.Options
  users []User
}

func (j *UserImportJob) Len() int { return len(j.users) }

func (j *UserImportJob) Run(ctx context.Context, start, end int) (vjob.Merge, error) {
  batch := j.users[start:end]
  return func() error {
    return saveUsers(batch)
  }, nil
}

err := vjob.RunWith(ctx, job, job.Options)
```

### Error recovery and stack helpers

```go
package main

import (
  "fmt"

  "github.com/imajinyun/go-knifer/verr"
)

func main() {
  err := verr.Recover(func() error {
    panic("boom")
  }, "run risky job")
  if err != nil {
    fmt.Println(err)
    fmt.Println(verr.GetStack(err))
  }

  collector := verr.NewCollector()
  collector.GoRun(func() error { return fmt.Errorf("task failed") }, "async task")
  if err := collector.Error(); err != nil {
    fmt.Println(err)
  }
}
```

## 📖 Doc

- Root package documentation: `doc.go`
- Public APIs: `doc.go` and facade files in each `v*` subpackage
- Test examples: `*_test.go` files under each module
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
