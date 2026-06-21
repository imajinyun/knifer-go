# vresty Quickstart

`vresty` is a resty-based HTTP client facade that provides chained requests, shortcut GET/POST helpers, response reads, downloads, safe URL validation, and global configuration.

## Which helper should I use?

| Scenario | Start with | Why |
| --- | --- | --- |
| Standard-library style request | `vhttp` `Get`, `Post`, request builders | Keeps dependencies light and behavior explicit. |
| Resty-based fluent request | `Get`, `Post`, `Put`, `Delete`, `NewRequest` | Uses Resty ergonomics while keeping go-knifer safety docs. |
| Untrusted URL | `GetSafe`, `PostSafe`, `NewSafeRequest`, `GetStringSafeE` | Applies URL validation before network access. |
| Simple response body as text | `GetStringE`, `PostStringE`, `PostJSONE`, `PostFormE` | Returns explicit errors instead of hiding failure in a response wrapper. |
| Bounded body reads | `WithMaxResponseBytes`, `WithMaxDecodeBytes` | Prevents unbounded memory growth when reading or decoding responses. |
| File download | `DownloadSafe`, `DownloadFileSafe`, `DownloadFileSafeWithOptions` | Keeps URL and filesystem risks visible. |
| Per-call HTTP client injection | `WithRestyClient`, `WithRestyClientFactory`, `NewIsolatedClient` | Keeps tests hermetic and avoids mutating global defaults. |
| Global application defaults | `SetGlobalTimeout`, `ConfigureGlobalConfig`, `WithScopedGlobalConfig` | Configure once at application startup or scope carefully in tests. |

## HTTP safety checklist

- Use safe helpers for untrusted URLs and pair them with `WithAllowedHosts`, `WithURLPolicy`, and `WithLookupIP` when SSRF matters.
- Set timeouts with `WithTimeout` or global configuration; avoid unbounded request waits.
- Bound response and decode sizes with `WithMaxResponseBytes`, `WithMaxDecodeBytes`, and safe download helpers.
- Use `NewIsolatedClient`, `WithRestyClient`, or `WithRestyClientFactory` in tests to avoid real network calls and global state leakage.
- Review redirect policy with `WithFollowRedirects` and `WithMaxRedirects`, especially for untrusted URLs.
- Treat `DownloadFile*` destinations as filesystem boundaries; configure overwrite, parent creation, permissions, and fake file providers in tests.
- Avoid logging request bodies, authorization headers, cookies, or downloaded data.

## When to use `vresty` instead of `vhttp`

`vresty` is the right facade when the codebase already depends on Resty or when fluent request construction is more readable than passing option lists through standard-library style helpers. If the request flow is simple and you do not need Resty features, `vhttp` is usually the smaller dependency surface.

| Decision point | Prefer `vresty` | Prefer `vhttp` |
| --- | --- | --- |
| Client model | Existing Resty client conventions, middleware, or fluent request style. | Standard-library-style helpers and minimal dependency expectations. |
| Request shape | Many per-request headers, query params, body mutations, or chained options. | Small GET/POST/download helpers with explicit option arguments. |
| Team familiarity | Callers already read Resty request chains. | Callers expect `net/http` semantics and explicit response closing. |
| Safety boundary | Use `GetSafe`, `PostSafe`, `DownloadSafe`, and `WithURLPolicy` while keeping Resty ergonomics. | Use the same safe policy concepts through `vhttp` helpers. |
| URL-only work | Use `vurl` first if no HTTP request should be sent yet. | Use `vurl` first if no HTTP request should be sent yet. |

The chained API keeps request construction close to execution:

```go
resp := vresty.Get("https://example.com/api").
	Header("X-Trace", "quickstart").
	Query("page", 1).
	Timeout(3 * time.Second).
	Execute()
```

For untrusted URLs, use the same shape with safe constructors and an explicit policy:

```go
resp := vresty.GetSafe("https://api.example.com/users",
	vresty.WithAllowedHosts("api.example.com"),
).Execute()
```

## Related packages

- Use `vhttp` for dependency-light HTTP helpers built around `net/http` rather than Resty clients.
- Use `vurl` for URL parsing, normalization, and SSRF-oriented policy before issuing requests.
- Use `vjson` when request or response payload fixtures need JSON formatting and path inspection.

## Benchmarks and trade-offs

Use the HTTP benchmark suite to measure the convenience and safety overhead on your machine:

```bash
go test -bench=. -benchmem -run=^$ ./internal/httpx/... ./vhttp ./vresty
```

The suite uses `httptest.Server` and temporary files only. It covers simple GET requests, JSON response decode, bounded body reads, safe URL validation, and safe file downloads. Treat the output as a local baseline rather than a universal performance claim.

`vhttp` does not replace `net/http`; it provides repeatable convenience helpers and safe entry points for common request, response, and download flows.

`vresty` does not replace Resty; it keeps Resty ergonomics while documenting go-knifer's safety boundaries and generated examples.

Safe APIs may add validation overhead. Use the benchmark commands in this document to measure the trade-off on your workload.

## When not to use vresty

- Use `vhttp` or `net/http` directly when the dependency surface should stay standard-library-only.
- Use Resty directly when your code depends on advanced Resty features that the facade does not expose.
- Use `vurl` when the task is only URL construction, normalization, validation, or safe opening without an HTTP request workflow.
- Use a dedicated downloader or streaming pipeline for very large files, resume support, checksums, or backpressure.
- Avoid global configuration mutation in reusable libraries; pass request/client options instead.

## FAQ

### Why not use only `net/http`?

Use `net/http` directly when you need full control. Use `vhttp` when the common request, bounded read, or safe download path matches your use case and you want less boilerplate.

### How do I choose `vhttp` vs `vresty`?

Choose `vhttp` for lightweight standard-library style helpers. Choose `vresty` when your codebase already uses Resty or needs Resty's fluent request/client model.

### Are safe APIs free?

No. Safe APIs perform validation before work that can touch untrusted network or filesystem boundaries. Measure with the documented benchmark commands.

### How does `vresty` relate to `vurl`?

Use `vurl` for URL construction, normalization, query encoding, content-length probing, and safe resource opening. Use `vresty` once the caller needs an HTTP request/response workflow with Resty-style chaining.

## Send simple GET requests

```go
package main

import (
	"fmt"

	"github.com/imajinyun/go-knifer/vresty"
)

func main() {
	body, err := vresty.GetStringE("https://example.com")
	if err != nil {
		panic(err)
	}
	fmt.Println(body)
}
```

## Set request parameters with chained requests

```go
package main

import (
	"fmt"
	"time"

	"github.com/imajinyun/go-knifer/vresty"
)

func main() {
	resp := vresty.Get("https://example.com/api").
		Header("X-Trace", "quickstart").
		Query("page", 1).
		Timeout(3 * time.Second).
		Execute()
	if resp.Err() != nil {
		panic(resp.Err())
	}
	fmt.Println(resp.Status())
	fmt.Println(resp.Body())
}
```

## Submit JSON or forms

```go
package main

import (
	"fmt"

	"github.com/imajinyun/go-knifer/vresty"
)

func main() {
	body, err := vresty.PostJSONE("https://example.com/users", `{"name":"alice"}`)
	if err != nil {
		panic(err)
	}
	fmt.Println(body)

	body, err = vresty.PostFormE("https://example.com/login", map[string]any{"user": "alice"})
	if err != nil {
		panic(err)
	}
	fmt.Println(body)
}
```

## Safe requests, downloads, and global configuration

```go
package main

import (
	"bytes"
	"fmt"
	"time"

	"github.com/imajinyun/go-knifer/vresty"
)

func main() {
	vresty.SetGlobalTimeout(5 * time.Second)
	defer vresty.ResetGlobalConfig()

	body, err := vresty.GetStringSafeE("https://example.com", vresty.WithAllowedHosts("example.com"))
	if err != nil {
		panic(err)
	}
	fmt.Println(len(body))

	var buf bytes.Buffer
	n, err := vresty.DownloadSafe("https://example.com", &buf, vresty.WithAllowedHosts("example.com"))
	if err != nil {
		panic(err)
	}
	fmt.Println(n)
}
```
