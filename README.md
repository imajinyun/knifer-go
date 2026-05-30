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
| `vresty` | `github.com/imajinyun/go-knifer/vresty` | Resty v3 based HTTP facade: chainable requests, JSON/form/multipart bodies, global headers/timeouts, downloads, and lightweight response helpers. |
| `vjson` | `github.com/imajinyun/go-knifer/vjson` | Ordered JSON objects/arrays, JSON parsing and formatting, path-based get/put, bean/list conversion, and XML/JSON conversion. |
| `vjwt` | `github.com/imajinyun/go-knifer/vjwt` | JWT creation, parsing, signing, verification, and time-claim validation; supports HMAC, RSA, ECDSA, and none signers. |
| `vlog` | `github.com/imajinyun/go-knifer/vlog` | Logging facade: console/color console loggers, log levels, global logger, and static logging functions. |
| `verr` | `github.com/imajinyun/go-knifer/verr` | Error helpers: panic recovery, error aggregation, multierror matching, stack capture/formatting, and optional logrus/Sentry integration. |
| `vconf` | `github.com/imajinyun/go-knifer/vconf` | Grouped configuration reader for setting/properties-style text and a simple YAML subset, with typed getters. |
| `vset` | `github.com/imajinyun/go-knifer/vset` | Generic and typed set utilities with add/remove/contains, set operations, and JSON/YAML encoding helpers. |
| `vjob` | `github.com/imajinyun/go-knifer/vjob` | Sliceable job execution: separate job data from scheduling options, typed slice/map adapters, context cancellation, and serialized merge callbacks; no generic type-alias experiment is required. |
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

### Resty v3 HTTP facade

`vresty` provides a thin, chainable facade over `resty.dev/v3`. It keeps the
public API lightweight while supporting common HTTP operations such as query
parameters, headers, cookies, Basic/Bearer auth, JSON/form bodies, multipart
uploads, per-request timeout, TLS skip verification, redirect control, and
downloads.

```go
package main

import (
 "fmt"
 "time"

 "github.com/imajinyun/go-knifer/vresty"
)

func main() {
 vresty.SetGlobalTimeout(5 * time.Second)
 vresty.SetGlobalHeader("X-App", "go-knifer")

 resp := vresty.Post("https://api.example.com/users").
  Query("source", "demo").
  BearerAuth("token").
  BodyJSON(`{"name":"go-knifer"}`).
  Timeout(3 * time.Second).
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

Shortcuts are available for simple cases and downloads:

```go
body := vresty.GetString("https://example.com")
jsonBody := vresty.PostJSON("https://api.example.com/events", `{"event":"created"}`)
n, err := vresty.DownloadFile("https://example.com/report.csv", "./downloads")
_, _, _ = body, jsonBody, n
_ = err
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
