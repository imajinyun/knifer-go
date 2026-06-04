# go-knifer

> 🍬 A set of Go tools that keep development sharp.

![go-knifer](./go-knifer.jpeg)

[![Go Reference](https://pkg.go.dev/badge/github.com/imajinyun/go-knifer.svg)](https://pkg.go.dev/github.com/imajinyun/go-knifer)
[![Go Version](https://img.shields.io/badge/go-%3E%3D1.20-00ADD8?logo=go)](https://go.dev/)

## 📚 Introduction

`go-knifer` is a practical utility toolkit for Go projects. It collects frequently used capabilities—string helpers, collection utilities, encoding/decoding, cryptography, HTTP, JSON, cache, cron, JWT, logging, configuration, and system information—into reusable packages.

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
text := vcrypto.MD5Hex("hello")
```

This style of utility wrapping reduces repeated code, avoids hidden risks from copy-paste snippets, and keeps the same scenarios represented by consistent APIs across a team.

## 🧭 Find by scenario

Not sure which package to import? Start from what you want to do:

| I want to… | Use |
| --- | --- |
| Trim, split, case-convert, or check blank strings | `vstr` |
| Filter / map / dedup / paginate a slice | `vslice` |
| Create, query, transform, merge, diff, or sort maps | `vmap` |
| Loosely convert `any` to int/float/bool/string | `vconv` |
| Precise arithmetic, rounding, or evaluate an expression | `vnum` |
| MD5/SHA/HMAC, AES/RSA, sign parameters | `vcrypto` |
| Non-cryptographic hashes (FNV, BKDR, …) | `vhash` |
| Encode/parse URLs, build/parse query strings | `vurl` |
| Base64 / Hex encode-decode | `vcodec` |
| Build/parse JSON, path get/put, JSON↔XML | `vjson` |
| Parse, build, or navigate XML | `vxml` |
| Generate UUID / Snowflake / NanoId | `vid` |
| Validate or parse ID-card numbers | `vident` |
| Read/write files, paths, copy, mkdir | `vfile` |
| Format/parse dates, offsets, day ranges | `vdate` |
| Send HTTP requests (standard library) | `vhttp` |
| Send HTTP requests (Resty-based) | `vresty` |
| Validate email/mobile/IP, etc. | `vvalidator` |
| Mask sensitive data | `vmask` |
| JWT sign/verify | `vjwt` |
| Schedule cron tasks | `vcron` |
| Cache with FIFO/LRU/LFU/TTL | `vcache` |

For the full list, see the module matrix below.

## 🧩 Module

The project follows an “internal implementation + public facade” layout: `internal/*` contains concrete implementations, while `v*` packages expose stable public APIs.

| Module | Import path | Description |
| --- | --- | --- |
| `vstr` | `github.com/imajinyun/go-knifer/vstr` | String helpers: blank/empty checks, trimming, splitting, substring helpers, formatting, emoji helpers, naming conversion, defaults, HTML escaping, and rune checks (blank, letter, digit, ASCII, letter-or-digit). |
| `vslice` | `github.com/imajinyun/go-knifer/vslice` | Slice helpers: contains/index, reverse, distinct, join, filter/map, sub-slice, concat, set-like operations, and paging. |
| `vmap` | `github.com/imajinyun/go-knifer/vmap` | Map helpers: construction, empty checks, contains/get/find, keys/values and sorted views, map/filter/reject/partition, reduce/group/count, inverse, merge/merge-with-resolver, intersect/diff/symmetric diff, pick/omit, update/clone, and equality checks. |
| `vconv` | `github.com/imajinyun/go-knifer/vconv` | Permissive type conversion: string, int, int64, float64, bool, bytes, and default-value variants. |
| `vdate` | `github.com/imajinyun/go-knifer/vdate` | Date/time helpers: common layouts, parse/format, begin/end of day/month/year, offsets, and comparisons. |
| `vfile` | `github.com/imajinyun/go-knifer/vfile` | File and IO helpers: read/write/copy, lines, mkdir/touch/delete, filename helpers, and quiet close. |
| `vcodec` | `github.com/imajinyun/go-knifer/vcodec` | Encoding helpers: Base64, URL-safe Base64, and Hex. |
| `vurl` | `github.com/imajinyun/go-knifer/vurl` | URL and URI helpers: parse, normalize, resolve relative URLs, query encode/decode, URL/path/fragment percent encoding, URL building, Data URI building, scheme checks, and file URL conversion. |
| `vnet` | `github.com/imajinyun/go-knifer/vnet` | Network helpers: IPv4/IPv6 conversion, CIDR/range/mask utilities, local ports, host/interface/MAC lookup, TLS config, and multipart form helpers. |
| `vobj` | `github.com/imajinyun/go-knifer/vobj` | Object helpers: nil/empty checks, equality, defaults, clone/serialization, comparison, type inspection, and container utilities. |
| `vver` | `github.com/imajinyun/go-knifer/vver` | Version helpers: version comparison, greater/less predicates, expression matching, inclusive ranges, and custom expression delimiters. |
| `vref` | `github.com/imajinyun/go-knifer/vref` | Reflection helpers: field lookup and mutation, method discovery and invocation, constructor-style function calls, type/value utilities, and method classification. |
| `vbean` | `github.com/imajinyun/go-knifer/vbean` | Bean/struct mapping helpers: struct/map conversion, copy properties, tag and alias matching, ignore-empty/zero options, and weak type conversion. |
| `vzip` | `github.com/imajinyun/go-knifer/vzip` | ZIP, gzip, and zlib helpers: archive creation/extraction, entry lookup, archive traversal, append, in-memory entries, and stream compression. |
| `vpoi` | `github.com/imajinyun/go-knifer/vpoi` | Office document helpers: lightweight Excel XLSX sheet listing, row reading, row writing, multi-sheet writing, and in-memory workbook creation. |
| `vmask` | `github.com/imajinyun/go-knifer/vmask` | Masking helpers: mask names, IDs, phones, addresses, email, passwords, license plates, bank cards, IPs, passports, and credit codes. |
| `vnum` | `github.com/imajinyun/go-knifer/vnum` | Numeric helpers: precise arithmetic, rounding modes, formatting, number checks, random unique numbers, ranges, factorial/combinations, gcd/lcm, binary conversion, comparison, parsing, byte conversion, expression calculation, and odd/even checks. |
| `vrand` | `github.com/imajinyun/go-knifer/vrand` | Random helpers: integers, floats, booleans, bytes, strings, numeric strings, and random element selection. |
| `vid` | `github.com/imajinyun/go-knifer/vid` | ID helpers: random/simple/fast UUIDs, MongoDB-style ObjectId, Snowflake generators and singleton next-id helpers, worker/datacenter id derivation, and NanoId generation. |
| `vident` | `github.com/imajinyun/go-knifer/vident` | Identity helpers: mainland China ID card 15/18-digit conversion, validation, check code, birthday/age/gender extraction, province/city/district code parsing, masking, and Hong Kong/Macau/Taiwan card validation. |
| `vhash` | `github.com/imajinyun/go-knifer/vhash` | Non-cryptographic hash helpers: additive, FNV, and a set of classic string hashes (RS, JS, PJW, ELF, BKDR, SDBM, DJB, AP, HF, HFIP, TianL, Java default). |
| `vvalidator` | `github.com/imajinyun/go-knifer/vvalidator` | Validation helpers: email, mobile, URL, IPv4, Chinese text, and number string checks. |
| `vtpl` | `github.com/imajinyun/go-knifer/vtpl` | Go html/template rendering helpers. |
| `vregex` | `github.com/imajinyun/go-knifer/vregex` | Regular-expression helpers: matching, group extraction, named groups, deletion, counting, index lookup, template/function replacement, and escaping. |
| `vbool` | `github.com/imajinyun/go-knifer/vbool` | Boolean helpers: negate, bool-to-int, all/any checks. |
| `vblf` | `github.com/imajinyun/go-knifer/vblf` | Bloom filters: bitmap/bitset/filter abstractions and multiple string hash algorithms. |
| `vcache` | `github.com/imajinyun/go-knifer/vcache` | Generic caches: FIFO, LFU, LRU, Timed, Weak, and NoCache; supports TTL, removal listeners, and lazy loading. |
| `vcaptcha` | `github.com/imajinyun/go-knifer/vcaptcha` | Image captcha generation: line, circle, shear, and GIF captchas, with random and math-expression generators. |
| `vcron` | `github.com/imajinyun/go-knifer/vcron` | Cron expression parsing and task scheduling, including both default and custom schedulers. |
| `vcrypto` | `github.com/imajinyun/go-knifer/vcrypto` | Cryptography and digests: MD5/SHA, HMAC, PBKDF2, parameter signing, random bytes, AES CBC/ECB/CTR/CFB/OFB/GCM, DES/3DES, RC4, Vigenere, XXTEA, RSA OAEP/PKCS#1/PSS, PEM, and X.509 certificate helpers. |
| `vdb` | `github.com/imajinyun/go-knifer/vdb` | Database helpers built on database/sql: SQL execution, named parameters, entities, conditions, query builders, transactions, pagination, and lightweight metadata lookup. |
| `vdfa` | `github.com/imajinyun/go-knifer/vdfa` | DFA word-tree matching: stop-rune filtering, first/all matches, dense and greedy match modes, found-word metadata, package-level matcher helpers, and text replacement. |
| `vhttp` | `github.com/imajinyun/go-knifer/vhttp` | Chainable HTTP client, downloads, per-call request options, BasicAuth, User-Agent parsing, and a simple server helper. |
| `vresty` | `github.com/imajinyun/go-knifer/vresty` | Resty v3 based HTTP facade: chainable requests, JSON/form/multipart bodies, per-call request options, downloads, and lightweight response helpers. |
| `vjson` | `github.com/imajinyun/go-knifer/vjson` | Ordered JSON objects/arrays, JSON parsing and formatting, path-based get/put, bean/list conversion, and XML/JSON conversion. |
| `vxml` | `github.com/imajinyun/go-knifer/vxml` | XML helpers: parse/read/write/format, tree navigation, simple XPath-style lookup, escaping, map/bean conversion, and namespace utilities. |
| `vjwt` | `github.com/imajinyun/go-knifer/vjwt` | JWT creation, parsing, signing, verification, and time-claim validation; supports HMAC, RSA, ECDSA, and none signers. |
| `vlog` | `github.com/imajinyun/go-knifer/vlog` | Logging facade: console/color console loggers, log levels, global logger, and static logging functions. |
| `verr` | `github.com/imajinyun/go-knifer/verr` | Error helpers: panic recovery, error aggregation, multierror matching, stack capture/formatting, and optional logrus/Sentry integration. |
| `vconf` | `github.com/imajinyun/go-knifer/vconf` | Grouped configuration reader for setting/properties-style text and a simple YAML subset, with typed getters. |
| `vset` | `github.com/imajinyun/go-knifer/vset` | Generic and typed set utilities with add/remove/contains, set operations, and JSON/YAML encoding helpers. |
| `vjob` | `github.com/imajinyun/go-knifer/vjob` | Sliceable job execution: separate job data from scheduling options, typed slice/map adapters, context cancellation, and serialized merge callbacks; no generic type-alias experiment is required. |
| `vsem` | `github.com/imajinyun/go-knifer/vsem` | Weighted, context-aware counting semaphore with FIFO fairness, try-acquire, close notifications, and in-use metrics. |
| `vskt` | `github.com/imajinyun/go-knifer/vskt` | TCP socket utilities: plain connections, NIO/AIO server/client helpers, and protocol encoder/decoder interfaces. |
| `vsys` | `github.com/imajinyun/go-knifer/vsys` | System and runtime information: host, OS, user, Go runtime, process memory, goroutines, environment variables, and more. |

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
- Facades may keep short names such as `vmask`, `vsem`, `vskt`, `vblf`,
  and `vver`; their meaning is documented in the module table above instead of
  changing established import paths.

Domain boundary rules:

- `vhash` is for non-cryptographic hash helpers such as additive/FNV (bucketing,
  bloom filters); `vcrypto` owns all security-oriented digests (MD5/SHA family),
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
  owns URL escaping, URL/URI parsing, normalization, resource, and scheme
  semantics.
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
helpers belong to `internal/poi` and are exposed through `vpoi`.

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

The `vjwt`, `vjson`, `vcron`, `vjob`, `vpoi`, and `vhttp`/`vresty` errors also
participate: their errors match `knifer.ErrCodeInvalidInput` (vjwt, vjson,
vcron, vjob, vpoi empty sheet name), `knifer.ErrCodeNotFound` (vpoi no sheet),
or `knifer.ErrCodeInternal` (vhttp/vresty and vskt) while preserving their own
error types, sentinels, and cause chains.

## 🚀 Install

Go 1.20 or later is required.

```bash
go get github.com/imajinyun/go-knifer
```

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
}
```

Request defaults can still be configured globally when needed, but new code
should prefer per-call options to keep request behavior explicit and avoid
cross-request state coupling. Available options include `WithTimeout`,
`WithHeader`, `WithHeaders`, `WithFollowRedirects`, `WithMaxRedirects`,
`WithSkipTLSVerify`, `WithTransport`, `WithClient`, `WithCookieJar`, and
`WithUserAgent`.

### Resty v3 HTTP facade

`vresty` provides a thin, chainable facade over `resty.dev/v3`. It keeps the
public API lightweight while supporting common HTTP operations such as query
parameters, headers, cookies, Basic/Bearer auth, JSON/form bodies, multipart
uploads, per-call options, TLS skip verification, redirect control, and
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
`WithFollowRedirects`, `WithMaxRedirects`, `WithSkipTLSVerify`,
`WithRestyClient`, `WithUserAgent`, and `WithCookieDisabled`.

Shortcuts are available for simple cases and downloads:

```go
body := vresty.GetString("https://example.com")
jsonBody := vresty.PostJSON("https://api.example.com/events", `{"event":"created"}`)
n, err := vresty.DownloadFile("https://example.com/report.csv", "./downloads")
_, _, _ = body, jsonBody, n
_ = err
```

### URL and URI helpers

`vurl` centralizes URL parsing, normalization, query string handling, percent
encoding, URL building, scheme checks, Data URI building, and file URL
conversion.

```go
package main

import (
  "fmt"

  "github.com/imajinyun/go-knifer/vurl"
)

func main() {
  normalized := vurl.Normalize(`example.com\docs/a b`, true, true)
  completed, _ := vurl.Complete("https://example.com/base/", "next?id=1")
  query := vurl.BuildQuery(map[string]any{"lang": "go", "page": 1})
  dataURI := vurl.DataURIBase64("text/plain", "aGVsbG8=")
  built := vurl.NewHTTPURLBuilder("example.com").AddPathSegment("a b").AddQuery("q", "go").Build()

  fmt.Println(normalized)
  fmt.Println(completed)
  fmt.Println(query)
  fmt.Println(vurl.IsWebURL(completed))
  fmt.Println(dataURI)
  fmt.Println(built)
}
```

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

  fmt.Println(ipLong, vnet.LongToIPv4(ipLong))
  fmt.Println(begin, end, vnet.IsInRange("192.168.1.8", "192.168.1.0/24"))
  fmt.Println(vnet.HideIPPart("192.168.1.8"))
}
```

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
last-write-wins merging and custom conflict resolution.

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
  grouped := vmap.GroupBy([]string{"go", "git", "java"}, func(s string) byte { return s[0] })

  fmt.Println(vmap.SortedKeys(merged))
  fmt.Println(evens)
  fmt.Println(grouped['g'])
}
```

### Database helpers

`vdb` provides SQL helpers on top of `database/sql`: named parameters,
condition builders, entity-based insert/update/delete, pagination,
transactions, and lightweight metadata lookup. Drivers and connection pools stay
under caller control.

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

```go
package main

import (
  "fmt"

  "github.com/imajinyun/go-knifer/vzip"
)

func main() {
  _ = vzip.ZipEntries("demo.zip", vzip.EntryData{Name: "hello.txt", Data: []byte("hello")})
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
  cipherText, err := vcrypto.AESEncryptCBC([]byte("secret message"), aesKey, iv)
  if err != nil {
    panic(err)
  }
  plain, err := vcrypto.AESDecryptCBC(cipherText, aesKey, iv)
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
