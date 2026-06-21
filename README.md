# 🔪 go-knifer

> 🔪 A Swiss Army knife for Go development, keeping your daily coding sharp.
>
> 🧰 Batteries-included utility toolkit: string, slice, map, crypto, HTTP, cache, ID generation, logging, config, and more. Import only what you need via `v*` domain packages.

![go-knifer](./go-knifer.jpeg)

[![Go Reference](https://pkg.go.dev/badge/github.com/imajinyun/go-knifer.svg)](https://pkg.go.dev/github.com/imajinyun/go-knifer)
[![Go Version](https://img.shields.io/badge/go-%3E%3D1.25-00ADD8?logo=go)](https://go.dev/)
[![CI](https://github.com/imajinyun/go-knifer/actions/workflows/go.yml/badge.svg)](https://github.com/imajinyun/go-knifer/actions/workflows/go.yml)
[![License](https://img.shields.io/github/license/imajinyun/go-knifer)](./LICENSE)

## 📑 Table of Contents

- [📚 Introduction](#introduction)
- [✨ Why go-knifer](#why-go-knifer)
- [🚀 Install](#install)
- [🧭 Find by scenario](#find-by-scenario)
- [🧩 Package catalog](#package-catalog)
- [🏗️ Architecture](#architecture)
- [🔒 API compatibility policy](#api-compatibility-policy)
- [✅ Recommended APIs](#recommended-apis)
- [📖 Documentation](#documentation)
- [📦 Build and test](#build-and-test)
- [🛡️ Governance](#governance)
- [🤝 Contributing](#contributing)
- [⭐ Star go-knifer](#star-go-knifer)

<a id="introduction"></a>

## 📚 Introduction

`go-knifer` is a practical utility toolkit for Go projects. It collects frequently used capabilities—string helpers, collection utilities, encoding/decoding, cryptography, HTTP, JSON, cache, cron, JWT, logging, configuration, and system information—into reusable packages.

The root package `github.com/imajinyun/go-knifer` is only the module entry point. Actual APIs live in public `v*` facade packages so applications can import only the domain they need.

<a id="why-go-knifer"></a>

## ✨ Why go-knifer

`knifer` comes from “knife”: a handy little tool for solving common everyday problems in Go development.

- 🧰 **Focused facades**: import `vstr`, `vslice`, `vhttp`, `vcrypto`, and other domain packages directly.
- 🧪 **Testable options**: many APIs provide `WithXxx` options and provider injection for deterministic tests.
- 🛡️ **Safe defaults**: security-sensitive helpers prefer explicit errors, bounded reads, SSRF-aware URL access, and path traversal checks.
- 📚 **Domain docs**: detailed quickstarts live under [`docs/doc`](./docs/doc/README.md), keeping this README easy to scan.

<a id="install"></a>

## 🚀 Install

Go 1.25 or later is required.

```bash
go get github.com/imajinyun/go-knifer
```

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

👉 See the [full documentation index](./docs/doc/README.md#package-catalog) for every package.

<a id="package-catalog"></a>

## 🧩 Package catalog

`go-knifer` follows an “internal implementation + public facade” layout: `internal/*` contains concrete implementations, while `v*` packages expose stable public APIs.

- 📦 Full module matrix: [`docs/doc/README.md#package-catalog`](./docs/doc/README.md#package-catalog)
- 🔎 Per-package quickstarts: [`docs/doc/*.md`](./docs/doc/README.md#quickstart-documents)
- 🧾 Exported API snapshot: [`docs/api/exports.txt`](./docs/api/exports.txt)

<a id="architecture"></a>

## 🏗️ Architecture

Application code should import public `v*` packages. `internal/*` packages are implementation details and can evolve without exposing every helper as public API.

For domain boundary rules, provider-injection patterns, API compatibility policy, error contracts, and safety defaults, see [Architecture and package boundaries](./docs/doc/README.md#architecture-and-package-boundaries).

<a id="api-compatibility-policy"></a>

## 🔒 API compatibility policy

`go-knifer` treats top-level `v*` facade packages as the public API boundary. The generated API snapshot in [`docs/api/exports.txt`](./docs/api/exports.txt) is reviewed with public API changes so upgrade risk is visible before release.

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
- 🌐 Online Go docs: [pkg.go.dev/github.com/imajinyun/go-knifer](https://pkg.go.dev/github.com/imajinyun/go-knifer)
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
git clone https://github.com/imajinyun/go-knifer.git
cd go-knifer
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
- Coverage/API/workflow gate details: see [Governance](./docs/doc/README.md#governance).

<a id="contributing"></a>

## 🤝 Contributing

Pull requests are welcome. Please add new capabilities to the appropriate `internal/*` implementation package first, expose public APIs from the corresponding `v*` package, add comments/tests, run local checks, and keep code formatted with `gofmt`.

For issue templates, PR principles, and gate expectations, see [Contributing](./docs/doc/README.md#contributing).

<a id="star-go-knifer"></a>

## ⭐ Star go-knifer

If this project helps you reduce repeated code, please consider giving it a Star. Your feedback and contributions will help make it a sharper Go utility toolkit.
