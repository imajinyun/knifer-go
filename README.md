# go-knifer

> 🍬 A set of Go tools that keep development sharp.

![go-knifer](./go-knifer.jpeg)

[![Go Reference](https://pkg.go.dev/badge/github.com/imajinyun/go-knifer.svg)](https://pkg.go.dev/github.com/imajinyun/go-knifer)
[![Go Version](https://img.shields.io/badge/go-%3E%3D1.20-00ADD8?logo=go)](https://go.dev/)

## 📚 Introduction

`go-knifer` is a practical utility toolkit for Go projects. Inspired by the role Hutool plays in the Java ecosystem, it collects frequently used capabilities—string helpers, collection utilities, encoding/decoding, cryptography, HTTP, JSON, cache, cron, JWT, logging, configuration, and system information—into reusable packages.

The root package `github.com/imajinyun/go-knifer` only acts as the module entry point. Actual APIs are split into multiple public `v*` packages by domain, so users can import only what they need without mixing unrelated utilities into business code.

## 🔪 Origin of the `go-knifer` name

`knifer` comes from “knife”: a handy little tool for solving common everyday problems in Go development. It does not try to replace the standard library. Instead, it lightly wraps standard-library features and common engineering practices to make code shorter, more consistent, and easier to maintain.

## ✨ How go-knifer changes the way we code

Before, calculating an MD5 digest often meant writing repetitive boilerplate in business code:

```go
sum := md5.Sum([]byte("hello"))
text := hex.EncodeToString(sum[:])
```

Now, with `go-knifer`, you can call a utility function directly:

```go
text := vbase.MD5Hex("hello")
```

This style of utility wrapping reduces repeated code, avoids hidden risks from copy-paste snippets, and keeps the same scenarios represented by consistent APIs across a team.

## 🧩 Module

The project follows an “internal implementation + public facade” layout: `internal/*` contains concrete implementations, while `v*` packages expose stable public APIs.

| Module | Import path | Description |
| --- | --- | --- |
| `vbase` | `github.com/imajinyun/go-knifer/vbase` | Core utilities: strings, random IDs, encoding/decoding, date/time, numbers, slices, maps, type conversion, file IO, regex, naming conversion, and more. |
| `vbf` | `github.com/imajinyun/go-knifer/vbf` | Bloom filters: bitmap/bitset/filter abstractions and multiple string hash algorithms. |
| `vcache` | `github.com/imajinyun/go-knifer/vcache` | Generic caches: FIFO, LFU, LRU, Timed, Weak, and NoCache; supports TTL, removal listeners, and lazy loading. |
| `vcaptcha` | `github.com/imajinyun/go-knifer/vcaptcha` | Image captcha generation: line, circle, shear, and GIF captchas, with random and math-expression generators. |
| `vcron` | `github.com/imajinyun/go-knifer/vcron` | Cron expression parsing and task scheduling, including both default and custom schedulers. |
| `vcrypto` | `github.com/imajinyun/go-knifer/vcrypto` | Cryptography and digests: MD5/SHA, HMAC, random bytes, AES-CBC/AES-GCM, RSA-OAEP, and RSA PEM encoding/decoding. |
| `vextra` | `github.com/imajinyun/go-knifer/vextra` | Extra utilities: gzip/zlib, zip/unzip, emoji helpers, Go template rendering, and common validation helpers. |
| `vhttp` | `github.com/imajinyun/go-knifer/vhttp` | Chainable HTTP client, downloads, global headers/timeouts, BasicAuth, User-Agent parsing, and a simple server helper. |
| `vjson` | `github.com/imajinyun/go-knifer/vjson` | Ordered JSON objects/arrays, JSON parsing and formatting, path-based get/put, bean/list conversion, and XML/JSON conversion. |
| `vjwt` | `github.com/imajinyun/go-knifer/vjwt` | JWT creation, parsing, signing, verification, and time-claim validation; supports HMAC, RSA, ECDSA, and none signers. |
| `vlog` | `github.com/imajinyun/go-knifer/vlog` | Logging facade: console/color console loggers, log levels, global logger, and static logging functions. |
| `vconf` | `github.com/imajinyun/go-knifer/vconf` | Grouped configuration reader for setting/properties-style text and a simple YAML subset, with typed getters. |
| `vset` | `github.com/imajinyun/go-knifer/vset` | Set utilities for string, int, int32, int64, uint, uint32, and uint64 values, with add/remove/contains and set operations. |
| `vsem` | `github.com/imajinyun/go-knifer/vsem` | Weighted, context-aware counting semaphore with FIFO fairness, try-acquire, close notifications, and in-use metrics. |
| `vskt` | `github.com/imajinyun/go-knifer/vskt` | TCP socket utilities: plain connections, NIO/AIO server/client helpers, and protocol encoder/decoder interfaces. |
| `vsys` | `github.com/imajinyun/go-knifer/vsys` | System and runtime information: host, OS, user, Go runtime, process memory, goroutines, environment variables, and more. |

## 🚀 Install

Go 1.20 or later is required.

```bash
go get github.com/imajinyun/go-knifer
```

Go will resolve the module according to the subpackages you actually import, for example:

```go
import (
 "github.com/imajinyun/go-knifer/vbase"
 "github.com/imajinyun/go-knifer/vhttp"
)
```

## 📝 Quick start

### Base utilities and JSON

```go
package main

import (
 "fmt"

 "github.com/imajinyun/go-knifer/vbase"
 "github.com/imajinyun/go-knifer/vjson"
)

func main() {
 name := vbase.DefaultIfBlank("", "go-knifer")

 obj := vjson.NewObject().
  Set("id", vbase.FastUUID()).
  Set("name", name).
  Set("tags", []string{"go", "tool"})

 fmt.Println(obj.GetString("name"))
 fmt.Println(obj.ToStringPretty())
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
 c.Put("answer", 42)

 value, ok := c.Get("answer")
 fmt.Println(value, ok)

 loaded, err := c.GetOrLoad("miss", func() (int, error) {
  return 100, nil
 })
 fmt.Println(loaded, err)
}
```

### Chainable HTTP request

```go
package main

import (
 "fmt"
 "time"

 "github.com/imajinyun/go-knifer/vhttp"
)

func main() {
 vhttp.SetGlobalTimeout(3 * time.Second)

 resp := vhttp.Get("https://example.com").
  Query("lang", "go").
  Header("X-Client", "go-knifer").
  FollowRedirects(true).
  Execute()

 if resp.Err() != nil {
  panic(resp.Err())
 }

 fmt.Println(resp.Status())
 fmt.Println(resp.ContentType())
 fmt.Println(resp.Body())
}
```

### Digest and JWT

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

## 📖 Doc

- Root package documentation: `doc.go`
- Public APIs: `doc.go` and facade files in each `v*` subpackage
- Test examples: `*_test.go` files under each module
- Online documentation: [pkg.go.dev/github.com/imajinyun/go-knifer](https://pkg.go.dev/github.com/imajinyun/go-knifer)

## 📦 Download & Build

Clone the source code:

```bash
git clone https://github.com/imajinyun/go-knifer.git
cd go-knifer
```

Run tests:

```bash
go test ./...
```

Format code:

```bash
gofmt -w .
```

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
