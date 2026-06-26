# vhttp Quickstart

`vhttp` provides chained HTTP requests, shortcut GET/POST/download helpers, response saving, global/client configuration, safe URL validation, and simple HTTP server wrappers.

## Recommended HTTP entry points

| Scenario | Recommended package | Recommended API family | Why |
| --- | --- | --- | --- |
| Standard-library style request | `vhttp` | `Get`, `Post`, request builders | Keeps dependencies light and behavior explicit. |
| Resty-based fluent client | `vresty` | client/request helpers | Uses Resty ergonomics while keeping knifer-go safety docs. |
| Untrusted URL | `vhttp`/`vresty` plus safe APIs | `Safe`/`E` variants | Applies validation and explicit errors before network access. |
| File download | `vhttp`/`vresty` safe download helpers | `DownloadFileSafe` family | Keeps path and transfer risks visible. |

## Which helper should I use?

Start with the smallest helper that expresses the trust boundary. Prefer `E` or chained response APIs when the caller must inspect errors, and prefer `Safe` APIs whenever the URL comes from configuration, user input, a webhook, or another service.

| Need | Use | Notes |
| --- | --- | --- |
| One-off GET body as a string | `GetStringE`, `GetWithTimeoutE` | Use for trusted URLs when a full `net/http` request would be boilerplate. |
| Chained request with headers, query, or timeout | `Get(...).Header(...).Query(...).Execute()` | The response object keeps status, headers, body helpers, and error state together. |
| JSON or form POST | `PostJSONE`, `PostFormE`, `Post(...).BodyString(...).Execute()` | Choose shortcut helpers for simple payloads and chained requests for custom headers or timeout policy. |
| Bounded response copy | `DownloadWithOptions`, `DownloadBytesE` | Set `WithMaxResponseBytes` when the remote body size is not tightly controlled. |
| Untrusted URL fetch | `GetStringSafeE`, `GetSafe`, `PostSafe` | Combine with `WithAllowedHosts`, `WithURLPolicy`, and resolver injection in tests. |
| Untrusted file download | `DownloadSafe`, `DownloadFileSafe` | Keep URL validation and destination-path decisions explicit at the call site. |
| URL parsing, normalization, or resource probing before HTTP | `vurl` | Use `vurl` when you need URL-only helpers without issuing a full request through `vhttp`. |

## Safe URL policy checklist

- Use `Safe` helpers for any URL that is not a compile-time constant owned by the application.
- Restrict schemes with `WithURLPolicy` or `WithAllowedSchemes`; most HTTP clients should allow only `http` and `https`.
- Prefer `WithAllowedHosts` for partner APIs and fixed upstreams. Host allow-lists are easier to audit than broad resolver rules.
- Keep `RejectPrivate` enabled for internet-facing inputs unless the caller is intentionally reaching a private test server or internal endpoint.
- Inject `WithLookupIP` in examples and tests when you need deterministic host classification without depending on external DNS.
- Keep response-size limits visible with `WithMaxResponseBytes` for downloads or other unbounded bodies.

## Benchmarks and trade-offs

Use the HTTP benchmark suite to measure the convenience and safety overhead on your machine:

```bash
go test -bench=. -benchmem -run=^$ ./internal/httpx/... ./vhttp ./vresty
```

The suite uses `httptest.Server` and temporary files only. It covers simple GET requests, JSON response decode, bounded body reads, safe URL validation, and safe file downloads. Treat the output as a local baseline rather than a universal performance claim.

`vhttp` does not replace `net/http`; it provides repeatable convenience helpers and safe entry points for common request, response, and download flows.

`vresty` does not replace Resty; it keeps Resty ergonomics while documenting knifer-go's safety boundaries and generated examples.

Safe APIs may add validation overhead. Use the benchmark commands in this document to measure the trade-off on your workload.

## Related packages

- Use `vresty` when the project standard is Resty clients, middleware, or fluent request configuration.
- Use `vurl` for URL parsing, normalization, query construction, and safety checks without sending a request.
- Use `vnet` when request policy depends on IP, DNS, port, or private-network classification helpers.

## When not to use vhttp

- Use `net/http` directly when you need full transport tuning, custom redirect behavior, streaming request/response bodies, connection pooling details, or middleware integration.
- Use `vresty` when the project already depends on Resty or fluent Resty request/client behavior is the expected abstraction.
- Use `vurl` when the task is URL normalization, validation, query construction, or resource probing without sending an HTTP request.
- Avoid non-safe shortcut helpers for URLs from users, config, webhooks, queues, or service discovery; use `Safe` APIs and explicit URL policy instead.
- Use a dedicated downloader for very large files, resume support, checksums, bandwidth limiting, or backpressure.

## FAQ

### Why not use only `net/http`?

Use `net/http` directly when you need full control. Use `vhttp` when the common request, bounded read, or safe download path matches your use case and you want less boilerplate.

### How do I choose `vhttp` vs `vresty`?

Choose `vhttp` for lightweight standard-library style helpers. Choose `vresty` when your codebase already uses Resty or needs Resty's fluent request/client model.

### Are safe APIs free?

No. Safe APIs perform validation before work that can touch untrusted network or filesystem boundaries. Measure with the documented benchmark commands.

## Send chained requests

```go
package main

import (
	"fmt"
	"time"

	"github.com/imajinyun/knifer-go/vhttp"
)

func main() {
	resp := vhttp.Get("https://example.com",
		vhttp.WithTimeout(5*time.Second),
		vhttp.WithHeader("Accept", "text/html"),
	).
		Query("page", 1).
		Execute()
	defer resp.Close()

	if err := resp.Err(); err != nil {
		panic(err)
	}
	fmt.Println(resp.Status(), resp.ContentType())
}
```

## Submit data with shortcut helpers

```go
package main

import (
	"fmt"
	"time"

	"github.com/imajinyun/knifer-go/vhttp"
)

func main() {
	body, err := vhttp.PostJSONE(
		"https://example.com/api",
		`{"name":"alice"}`,
	)
	if err != nil {
		panic(err)
	}
	fmt.Println(body)

	body, err = vhttp.GetWithTimeoutE("https://example.com", 3*time.Second)
	if err != nil {
		panic(err)
	}
	fmt.Println(len(body))
}
```

## Download response content

```go
package main

import (
	"bytes"
	"fmt"

	"github.com/imajinyun/knifer-go/vhttp"
)

func main() {
	data, err := vhttp.DownloadBytesE("https://example.com/file.txt")
	if err != nil {
		panic(err)
	}
	fmt.Println(len(data))

	var buf bytes.Buffer
	n, err := vhttp.DownloadWithOptions("https://example.com/file.txt", &buf, vhttp.WithMaxResponseBytes(1<<20))
	if err != nil {
		panic(err)
	}
	fmt.Println(n, buf.Len())
}
```

## Create a simple HTTP service

```go
package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/imajinyun/knifer-go/vhttp"
)

func main() {
	server := vhttp.NewSimpleServerWithOptions(8080, vhttp.WithReadHeaderTimeout(5*time.Second))
	server.AddAction("/health", func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, "ok")
	})
	server.SetRoot("./public")

	// errCh := server.StartAsync()
	// _ = server.Stop(5 * time.Second)
	_ = server
}
```
