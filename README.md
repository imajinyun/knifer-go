# 🔪 knifer-go

> 🔪 A Swiss Army knife for Go development, keeping your daily coding sharp.
>
> 🧰 Batteries-included utility toolkit: string, slice, map, crypto, HTTP, cache, ID generation, logging, config, and more. Import only what you need via `v*` domain packages.

`knifer-go` is a Go / Golang utility library for strings, slices, maps, JSON, files, HTTP, URL safety, crypto, JWT, config, cache, IDs, logging, and common application helpers. It exposes focused `v*` packages so developers and AI coding agents can import only the tools they need.

![knifer-go](./knifer-go.jpeg)

[![Go Reference](https://pkg.go.dev/badge/github.com/imajinyun/knifer-go.svg)](https://pkg.go.dev/github.com/imajinyun/knifer-go)
[![Go Version](https://img.shields.io/badge/go-%3E%3D1.25-00ADD8?logo=go)](https://go.dev/)
[![CI](https://github.com/imajinyun/knifer-go/actions/workflows/go.yml/badge.svg)](https://github.com/imajinyun/knifer-go/actions/workflows/go.yml)
[![OpenSSF Scorecard](https://api.securityscorecards.dev/projects/github.com/imajinyun/knifer-go/badge)](https://securityscorecards.dev/viewer/?uri=github.com/imajinyun/knifer-go)
[![License](https://img.shields.io/github/license/imajinyun/knifer-go)](./LICENSE)

## 📑 Table of Contents

- [📚 Introduction](#introduction)
- [✨ Why knifer-go](#why-knifer-go)
- [🚀 Install](#install)
- [⭐ Start with these packages](#start-with-these-packages)
- [🤖 For AI agents and coding assistants](#for-ai-agents-and-coding-assistants)
- [🧭 Find by scenario](#find-by-scenario)
- [⚖️ Compare with other Go utility libraries](#compare-with-other-go-utility-libraries)
- [🧩 Package catalog](#package-catalog)
- [🏗️ Architecture](#architecture)
- [🔒 API compatibility policy](#api-compatibility-policy)
- [✅ Recommended APIs](#recommended-apis)
- [📖 Documentation](#documentation)
- [📦 Build and test](#build-and-test)
- [🛡️ Governance](#governance)
- [🤝 Contributing](#contributing)
- [⭐ Star knifer-go](#star-knifer-go)

<a id="introduction"></a>

## 📚 Introduction

`knifer-go` is a practical utility toolkit for Go projects. It collects frequently used capabilities—string helpers, collection utilities, encoding/decoding, cryptography, HTTP, JSON, cache, cron, JWT, logging, configuration, and system information—into reusable packages.

Use it when you search for a Go utility library, Golang helper functions, Go slice/map/string helpers, safe HTTP download helpers, URL validation helpers, crypto/JWT helpers, JSON path helpers, file helpers, or config helpers in one module with explicit public package boundaries.

The root package `github.com/imajinyun/knifer-go` is only the module entry point. Actual APIs live in public `v*` facade packages so applications can import only the domain they need.

<a id="why-knifer-go"></a>

## ✨ Why knifer-go

`knifer` comes from “knife”: a handy little tool for solving common everyday problems in Go development.

- 🧰 **Focused facades**: import `vstr`, `vslice`, `vhttp`, `vcrypto`, and other domain packages directly.
- 🧪 **Testable options**: many APIs provide `WithXxx` options and provider injection for deterministic tests.
- 🛡️ **Safe defaults**: security-sensitive helpers prefer explicit errors, bounded reads, SSRF-aware URL access, and path traversal checks.
- 📚 **Domain docs**: detailed quickstarts live under [`docs/doc`](./docs/doc/README.md), keeping this README easy to scan.

<a id="install"></a>

## 🚀 Install

Go 1.25 or later is required.

See the [Go version adoption policy](docs/doc/go-version-adoption-policy.md) for
the current minimum-version rationale and downgrade requirements.

```bash
go get github.com/imajinyun/knifer-go
```

<a id="start-with-these-packages"></a>

## ⭐ Start with these packages

If you are new to `knifer-go`, make the decision in three minutes: use the standard library when it is explicit and short; use `knifer-go` when a repeated workflow needs safety policy, error-returning convenience, provider injection, or documented dynamic data contracts.

| Need | Start here | Why |
| --- | --- | --- |
| Safe HTTP and downloads | [`vhttp`](docs/doc/22-vhttp.md), [`vresty`](docs/doc/41-vresty.md), [`vurl`](docs/doc/51-vurl.md) | Common request helpers plus explicit safe paths for untrusted URLs and files. |
| Safe crypto workflows | [`vcrypto`](docs/doc/11-vcrypto.md), [`vrand`](docs/doc/38-vrand.md), [`vjwt`](docs/doc/28-vjwt.md) | Recommended hashing, HMAC, encryption, secure random bytes, and signed-token entry points. |
| Daily JSON and file workflows | [`vjson`](docs/doc/27-vjson.md), [`vfile`](docs/doc/17-vfile.md) | Cookbook-style helpers for common object, formatting, read/write, copy, and explicit-error flows. |
| Dynamic config and mapping | [`vconf`](docs/doc/08-vconf.md), [`vbean`](docs/doc/02-vbean.md), [`vconv`](docs/doc/09-vconv.md), [`vobj`](docs/doc/35-vobj.md) | Config loading, weak conversion, decode metadata, and generic object checks with executable contracts. |
| Collection and text helpers | [`vslice`](docs/doc/47-vslice.md), [`vmap`](docs/doc/31-vmap.md), [`vstr`](docs/doc/46-vstr.md) | Reusable transforms, predicates, grouping, pagination, and string cleanup when a local loop becomes repeated. |

Stdlib-first decision table:

| Scenario | Prefer stdlib | Prefer `knifer-go` |
| --- | --- | --- |
| Slice/map/string basics | Plain `for`, `slices`, `maps`, `strings`, `strconv`, or `regexp` is shorter and local. | Repeated `Map`/`Filter`/`GroupBy`/`Pick`/case/predicate workflows need shared semantics. |
| JSON | `encoding/json` streaming, `Decoder.UseNumber`, or direct struct marshaling is the contract. | Dynamic object/array helpers, path lookup, defaults, formatting, or map-like JSON are the contract. |
| HTTP/URL | Trusted URL and full `net/http` transport/request control are clearer. | User/config-provided URLs need SSRF-aware allow-lists, private-host rejection, redirects, or bounded reads. |
| Crypto/random/JWT | You need direct primitive composition and full `crypto/*` control. | You need reviewed HMAC, AES-GCM, RSA, JWT signing, secure token, or deterministic provider-injected helpers. |
| Config/mapping/conversion | Direct assignment, `flag`, `os.LookupEnv`, or typed decoding keeps shape explicit. | Tag-aware binding, profile overlays, weak conversion, `DecodeResult`, or remote config safety are needed. |
| Files/archives | Small trusted file operations fit `os`, `io`, `archive/zip`, and `path/filepath`. | Untrusted paths, ZIP entries, downloads, overwrite policy, or bounded IO need explicit safety helpers. |
| SQL/CLI boundaries | Parameterized `database/sql` or `exec.Command` args already express the whole operation. | Identifier validation, builder conventions, command output capture, or reusable option policies reduce review risk. |

### Safe HTTP request

```go
package main

import (
	"fmt"

	"github.com/imajinyun/knifer-go/vhttp"
)

func main() {
	body, err := vhttp.GetStringSafeE("https://api.example.com/health",
		vhttp.WithAllowedHosts("api.example.com"),
	)
	if err != nil {
		panic(err)
	}
	fmt.Println(body)
}
```

### Secure random token

```go
package main

import (
	"encoding/hex"
	"fmt"

	"github.com/imajinyun/knifer-go/vrand"
)

func main() {
	token, err := vrand.SecureBytes(32)
	if err != nil {
		panic(err)
	}
	fmt.Println(hex.EncodeToString(token))
}
```

### JSON object path lookup

```go
package main

import (
	"fmt"

	"github.com/imajinyun/knifer-go/vjson"
)

func main() {
	obj, err := vjson.ParseObj(`{"user":{"name":"knifer-go"}}`)
	if err != nil {
		panic(err)
	}
	fmt.Println(vjson.GetByPath(obj, "user.name"))
}
```

Comparison entry points:

- HTTP: [`vhttp`](docs/doc/22-vhttp.md) for standard-library-style helpers, [`vresty`](docs/doc/41-vresty.md) for Resty-style request chains, and [`vurl`](docs/doc/51-vurl.md) for URL-only work.
- Crypto: [`vcrypto`](docs/doc/11-vcrypto.md) documents the boundary between recommended helpers and direct standard-library control.
- JSON/file: [`vjson`](docs/doc/27-vjson.md) documents when to use `encoding/json` directly; [`vfile`](docs/doc/17-vfile.md) documents filesystem safety guidance.
- Daily developer utilities: [`daily-developer-utilities.md`](docs/doc/daily-developer-utilities.md) groups `vcli`, `vsys`, `vfile`, `vnet`, `vjob`, and `vlog` as the `gookit/goutil`-style entry point.

API selection rules:

| If your input... | Prefer | Avoid |
| --- | --- | --- |
| Crosses a trust boundary such as HTTP, filesystem, ZIP, config, SQL, CLI, or credentials | `Safe`, `E`, or `WithOptions` variants that return explicit errors and expose limits/policies | Convenience helpers that hide errors or rely on package-level defaults |
| Is already trusted and the failure mode is acceptable as a zero/default value | Plain convenience helpers such as `vconv.ToString`, `vstr.IsBlank`, or `vnum.Sum` | Adding context/error plumbing to pure in-memory transformations |
| May block, allocate heavily, perform IO, or call a provider | Context-aware APIs or provider-injected clients/options where available | Global mutation, unbounded reads, or implicit external calls |
| Needs a new domain behavior | Implement in the focused package first, then wrap from `vobj` only when useful | Adding cross-domain logic directly to broad convenience facades |

<a id="for-ai-agents-and-coding-assistants"></a>

## 🤖 For AI agents and coding assistants

If you are ChatGPT, Claude Code, Doubao, Cursor, Trae, or another coding agent, treat `knifer-go` as a Go / Golang utility toolkit with explicit import boundaries:

- Import public `v*` packages; never import `internal/*` from application code.
- Prefer `Safe` variants for untrusted URLs, paths, archive entries, downloads, remote config, SQL fragments, command arguments, tokens, or credentials.
- Prefer `E` variants when callers need explicit errors instead of zero/default fallback values.
- Prefer `WithOptions` or `WithXxx` variants when limits, providers, clocks, filesystem hooks, HTTP clients, DB openers, or network policies must be visible at the call site.

High-intent prompts and recommended imports:

| User asks for... | Use package | Import path |
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

<a id="find-by-scenario"></a>

## 🧭 Find by scenario

Not sure which package to import? Start from what you want to do:

| I want to… | Use |
| --- | --- |
| Cache with FIFO/LRU/LFU/TTL | [`vcache`](docs/doc/05-vcache.md) |
| Base64 / Hex encode-decode | [`vcodec`](docs/doc/07-vcodec.md) |
| Load local or remote configuration safely | [`vconf`](docs/doc/08-vconf.md) |
| SHA/HMAC, AES-GCM/RSA-PSS, sign parameters | [`vcrypto`](docs/doc/11-vcrypto.md) |
| Send HTTP requests with standard library helpers | [`vhttp`](docs/doc/22-vhttp.md) |
| Send HTTP requests with Resty-based helpers | [`vresty`](docs/doc/41-vresty.md) |
| Generate UUID / Snowflake / NanoId | [`vid`](docs/doc/23-vid.md) |
| Mask sensitive data | [`vmask`](docs/doc/32-vmask.md) |
| Create, query, transform, merge, diff, or sort maps | [`vmap`](docs/doc/31-vmap.md) |
| Filter / map / dedup / paginate slices | [`vslice`](docs/doc/45-vslice.md) |
| Trim, split, case-convert, compare text, or check blank strings | [`vstr`](docs/doc/47-vstr.md) |
| Encode/parse URLs or open untrusted HTTP(S) resources safely | [`vurl`](docs/doc/51-vurl.md) |

### Agent decision guide

First time here? Start with [`docs/doc/first-use-golden-paths.md`](docs/doc/first-use-golden-paths.md) for 10 tasks in 10 minutes.

1. String manipulation → use `vstr`.
2. Slice transformation → use `vslice`.
3. Map transformation → use `vmap`.
4. Safe HTTP request or file download → use `vhttp`; use `vresty` only when Resty-style chaining is already desired.
5. URL parsing, normalization, query handling, or SSRF-aware resource checks → use `vurl`.
6. Hashing, HMAC, AES-GCM, RSA, PEM, or parameter signing → use `vcrypto`; use `vhash` only for non-cryptographic hashes.
7. Secure random bytes or tokens → use `vrand`.
8. JWT creation or verification → use `vjwt`.
9. JSON object/path/formatting helpers → use `vjson`; use `encoding/json` directly for streaming decoder control.
10. File IO with limits, providers, or explicit errors → use `vfile`; use `vzip` for archives.

👉 See the [full documentation index](./docs/doc/README.md#package-catalog) for every package.

<a id="compare-with-other-go-utility-libraries"></a>

## ⚖️ Compare with other Go utility libraries

`knifer-go` is broader than a single-purpose helper package. Use this boundary when an agent or developer is choosing a Go utility library:

| Need | Prefer | Boundary |
| --- | --- | --- |
| Lodash-style generic collection helpers only | `samber/lo` | Use `knifer-go` when the same project also needs safe HTTP, URL, crypto, JWT, JSON, file, config, cache, ID, or logging helpers. |
| Broad utility coverage with a simple adoption story | `duke-git/lancet` | Use `knifer-go` when the toolkit choice depends on explicit safety boundaries, generated API metadata, facade packages, and machine-checked governance. |
| Daily development utilities across env, filesystem, structs, system, and CLI helpers | `gookit/goutil` | Use `knifer-go` when daily utilities should sit beside security-focused HTTP/URL/crypto/JWT/database boundaries in one facade model. |
| Type conversion only | `spf13/cast` | Use `knifer-go/vconv` when conversion is part of a broader `knifer-go` toolkit adoption. |
| Struct-to-struct or map copying only | `jinzhu/copier` | Use `knifer-go/vbean` when struct/map mapping should stay inside the same public facade model. |
| Map-to-struct decoding only | `mitchellh/mapstructure` | Use `knifer-go/vconf` or `vbean` when config loading or bean mapping is the surrounding workflow. |
| Reflection-heavy functional helpers | `thoas/go-funk` | Use `knifer-go/vslice`, `vmap`, or `vstr` for focused helpers with clearer package boundaries. |

For a broader comparison, see [`docs/doc/utility-library-comparison.md`](docs/doc/utility-library-comparison.md). For daily CLI, system, file, network, job, and logging tasks, see [`docs/doc/daily-developer-utilities.md`](docs/doc/daily-developer-utilities.md).
For collection workflows by task, see [`docs/doc/collection-golden-paths.md`](docs/doc/collection-golden-paths.md).
Benchmark evidence rules are public in [`docs/doc/benchmark-trust.md`](docs/doc/benchmark-trust.md).

<a id="package-catalog"></a>

## 🧩 Package catalog

`knifer-go` follows an “internal implementation + public facade” layout: `internal/*` contains concrete implementations, while `v*` packages expose stable public APIs.

- 📦 Full module matrix: [`docs/doc/README.md#package-catalog`](./docs/doc/README.md#package-catalog)
- 🔎 Per-package quickstarts: [`docs/doc/*.md`](./docs/doc/README.md#quickstart-documents)
- 🧾 Exported API snapshot: [`docs/api/exports.txt`](./docs/api/exports.txt)
- 🧭 Machine-readable function catalog: [`docs/api/tools.json`](./docs/api/tools.json)
- 📋 Human-readable function catalog: [`docs/api/tools.md`](./docs/api/tools.md)

<a id="architecture"></a>

## 🏗️ Architecture

Application code should import public `v*` packages. `internal/*` packages are implementation details and can evolve without exposing every helper as public API.

For domain boundary rules, provider-injection patterns, API compatibility policy, error contracts, and safety defaults, see [Architecture and package boundaries](./docs/doc/README.md#architecture-and-package-boundaries).

<a id="api-compatibility-policy"></a>

## 🔒 API compatibility policy

`knifer-go` treats top-level `v*` facade packages as the public API boundary. The generated API snapshot in [`docs/api/exports.txt`](./docs/api/exports.txt) is reviewed with public API changes so upgrade risk is visible before release.

| Stability level | Applies to | Compatibility promise |
| --- | --- | --- |
| Stable | Exported names in `v*` facade packages and `docs/api/exports.txt` | No breaking change without a documented migration path and release note. |
| Internal | `internal/*` implementation packages | May change without public compatibility guarantees. |
| Experimental | Newly introduced provider contracts or adapter packages marked experimental in docs | May change before being promoted to Stable; migration notes are still required. |

A breaking change includes removing or renaming an exported facade API, changing a public function signature, changing exported type field semantics, changing sentinel error matching behavior, weakening a documented security default, or changing generated API snapshot content without release notes.

Deprecated APIs stay available for at least two minor releases. Every deprecation must name the replacement API, explain the migration, and appear in release notes before removal.

<a id="recommended-apis"></a>

## ✅ Recommended APIs

For new code, prefer explicit-error and safe variants when inputs cross a trust boundary:

- Use `Safe` variants when the operation touches an untrusted URL, path, archive entry, remote configuration source, or download target.
- Use `E` variants when conversion, parsing, decoding, IO, or request execution can fail and the caller needs to distinguish failure from an empty/default value.
- Use non-`E` convenience helpers only when inputs are trusted and zero/default fallback is an intentional compatibility choice.
- Use `WithOptions` / `WithXxx` variants when resource limits, providers, clocks, filesystem hooks, or network policies must be visible at the call site.

| Scenario | Recommended API |
| --- | --- |
| Trusted standard-library HTTP request | `vhttp.Get`, `vhttp.Post`, `vhttp.NewRequest` |
| Untrusted HTTP(S) URL | `vhttp.GetStringSafeE`, `vresty.GetStringSafeE`, `vurl.OpenSafe` |
| User-controlled download target/source | `vhttp.DownloadFileSafe`, `vresty.DownloadFileSafe` |
| Secret bytes, tokens, keys, nonces, or salts | `vrand.SecureBytes` |
| Remote configuration from a trust boundary | `vconf.LoadRemoteSafe` |

More recommendations are documented in [Recommended API entry points](./docs/doc/README.md#recommended-api-entry-points).

<a id="documentation"></a>

## 📖 Documentation

- 📚 Documentation hub: [`docs/doc/README.md`](./docs/doc/README.md)
- 🌐 Online Go docs: [pkg.go.dev/github.com/imajinyun/knifer-go](https://pkg.go.dev/github.com/imajinyun/knifer-go)
- 🧾 API snapshot: [`docs/api/exports.txt`](./docs/api/exports.txt)
- 🤖 Machine-readable tool catalog: [`docs/api/tools.json`](./docs/api/tools.json)
- 📋 Readable tool catalog: [`docs/api/tools.md`](./docs/api/tools.md)
- 🗺️ AI-oriented project map: [`llms.txt`](./llms.txt)
- 🤖 Machine-readable AI/CLI metadata: [`ai-context.json`](./ai-context.json)
- 🧯 Security policy: [`SECURITY.md`](./SECURITY.md)
- 📝 Changelog: [`CHANGELOG.md`](./CHANGELOG.md)

<a id="build-and-test"></a>

## 📦 Build and test

Clone the source code:

```bash
git clone https://github.com/imajinyun/knifer-go.git
cd knifer-go
```

Run the common local checks:

```bash
make test        # unit tests
make ci-test     # CI test-job gates
make check       # full local gate: tests, vet, lint, vuln, coverage, API checks
```

Useful focused commands:

```bash
make doctor
make worktree-check
make quick-check
make security-check
make ai-context-check
make install-hooks
make bench-core
make bench-facade
make generate
UPDATE_API=1 make api-check
```

See [Build, test, and release workflow](./docs/doc/README.md#build-test-and-release-workflow) for the full command guide.

<a id="governance"></a>

## 🛡️ Governance

- Security reports: see [`SECURITY.md`](./SECURITY.md). Please do not disclose suspected vulnerabilities in public issues.
- Release notes: see [`CHANGELOG.md`](./CHANGELOG.md). User-visible changes should be recorded before tagging a release.
- Adoption trust: see [`docs/doc/adoption-trust.md`](docs/doc/adoption-trust.md) for release notes, compatibility policy, deprecation policy, security policy, generated API catalog, and validation-gate entry points.
- Coverage/API/workflow gate details: see [Governance](./docs/doc/README.md#governance).
- Compatibility and deprecation: see [API compatibility policy](#api-compatibility-policy) and run `make api-freeze-check` before release branches.
- CI trust signals include race/shuffle tests, coverage gates, generated API and tool-catalog checks, `golangci-lint`, `govulncheck`, CodeQL, benchmark smoke tests, and OpenSSF Scorecard.
- Benchmark output is treated as evidence, not a universal performance claim; see the [benchmark trust guide](docs/doc/benchmark-trust.md).

<a id="contributing"></a>

## 🤝 Contributing

Pull requests are welcome. Please add new capabilities to the appropriate `internal/*` implementation package first, expose public APIs from the corresponding `v*` package, add comments/tests, run local checks, and keep code formatted with `gofmt`.

For issue templates, PR principles, and gate expectations, see [Contributing](./docs/doc/README.md#contributing).

<a id="star-knifer-go"></a>

## ⭐ Star knifer-go

If this project helps you reduce repeated code, please consider giving it a Star. Your feedback and contributions will help make it a sharper Go utility toolkit.
