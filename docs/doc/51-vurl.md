# vurl Quickstart

`vurl` provides URL building, parsing, encoding/decoding, query handling, normalization, resource opening, and safe HTTP(S) resource access helpers.

## Which helper should I use?

Use `vurl` when the task is about URL shape or safe resource access rather than a full HTTP request workflow. Use `vhttp` or `vresty` when you need request headers, methods, response status handling, downloads, or client configuration.

| Scenario | Recommended helper family | Related package |
| --- | --- | --- |
| Build a URL from host, path, query, and fragment | `NewHTTPURLBuilder`, `AppendQuery`, `BuildQuery` | Use before passing the URL to `vhttp` or `vresty`. |
| Encode or decode path, query, or fragment data | `Encode*`, `Decode*`, `EncodeQueryMap`, `DecodeQueryFirst` | Keeps escaping rules explicit before an HTTP request is built. |
| Parse, inspect, or normalize URL strings | `ParseHTTP`, `Host`, `Path`, `NormalizeUsingOptions` | Normalize first, then apply package-specific HTTP policy. |
| Safely open or inspect HTTP(S) resources | `OpenSafeWithOptions`, `ContentLengthSafeWithOptions` | Use `vhttp`/`vresty` for richer request and download flows. |
| Enforce URL trust policy | `WithAllowedSchemes`, `WithAllowedHosts`, `WithRejectPrivateHosts`, `WithLookupIP` | Mirrors the safe URL policy concepts used by `vhttp` and `vresty`. |

## URL safety checklist

- Treat URLs from users, configuration files, webhooks, and upstream services as untrusted.
- Allow only the schemes the caller needs. Most remote resource reads should use `WithAllowedSchemes("http", "https")`.
- Use `WithAllowedHosts` for known partner domains or fixed upstream hosts.
- Keep private-host rejection enabled for internet-facing inputs; disable it only for controlled tests or documented internal calls.
- Inject `WithLookupIP` in deterministic tests so host validation does not depend on external DNS.
- Set `WithTimeout` and `WithMaxBytes` for remote reads to avoid hanging calls and unbounded response bodies.
- Safe helpers validate redirects before following them; keep that behavior enabled for untrusted input.
- Use `vhttp` or `vresty` safe helpers when you need method-specific requests, headers, or download helpers after URL validation.

## When not to use vurl

- Use `vhttp` or `vresty` when the task is a full HTTP workflow with methods, headers, response status handling, downloads, retries, or client-level configuration.
- Use `vnet` when the task is IP math, CIDR/range handling, DNS records, TCP probes, TLS configuration, or multipart upload saving.
- Use `Open` and `OpenWithOptions` only for trusted resource locations. Use `OpenSafe` or `OpenSafeWithOptions` for user, configuration, webhook, or upstream URLs.

## Related packages

- Use `vhttp` or `vresty` when a validated URL should be used for an outbound HTTP request.
- Use `vnet` when policy depends on resolved IPs, private ranges, ports, or network interfaces.
- Use `vcodec` when URL-safe encoding or decoding is the main task rather than URL construction.

## Benchmarks and trade-offs

Measure URL parsing, normalization, query encoding, and safe resource access overhead locally:

```bash
go test -bench=. -benchmem -run=^$ ./internal/url ./vurl
```

Use benchmark output as a local baseline. Do not treat results as universal throughput claims; options such as safe host validation, DNS lookup injection, timeouts, and byte limits change the work performed by each helper.

## Build URLs

```go
package main

import (
	"fmt"

	"github.com/imajinyun/knifer-go/vurl"
)

func main() {
	raw := vurl.NewHTTPURLBuilder("example.com").
		SetPath("/search").
		AddQuery("q", "go knifer").
		SetFragment("top").
		Build()

	fmt.Println(raw)
}
```

## Encode, decode, and use query maps

```go
package main

import (
	"fmt"

	"github.com/imajinyun/knifer-go/vurl"
)

func main() {
	encoded := vurl.Encode("a b&c")
	decoded, err := vurl.Decode(encoded)
	if err != nil {
		panic(err)
	}

	fmt.Println(encoded)
	fmt.Println(decoded)
	fmt.Println(vurl.EncodeQueryMap(map[string]any{"q": "go", "page": 1}))
	fmt.Println(vurl.DecodeQueryFirst("q=go&page=1"))
}
```

## Parse and normalize URLs

```go
package main

import (
	"fmt"

	"github.com/imajinyun/knifer-go/vurl"
)

func main() {
	u, err := vurl.ParseHTTP("https://example.com/a b?q=go")
	if err != nil {
		panic(err)
	}

	fmt.Println(vurl.Host(u))
	fmt.Println(vurl.Path(u.String()))
	fmt.Println(vurl.NormalizeUsingOptions("example.com//a b", vurl.WithDefaultScheme("https"), vurl.WithEncodePath(true)))
	fmt.Println(vurl.IsWebURL(u.String()))
}
```

## Safely open HTTP(S) resources

```go
package main

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/imajinyun/knifer-go/vurl"
)

func main() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("ok"))
	}))
	defer server.Close()

	r, err := vurl.OpenSafeWithOptions(server.URL,
		vurl.WithTimeout(time.Second),
		vurl.WithAllowedSchemes("http", "https"),
		vurl.WithRejectPrivateHosts(false),
		vurl.WithMaxBytes(16),
	)
	if err != nil {
		panic(err)
	}
	defer r.Close()

	body, _ := io.ReadAll(r)
	fmt.Println(string(body))
}
```

## FAQ

### Does normalization make a URL safe?

No. Normalization changes URL shape; it does not decide whether a host, scheme, redirect, or response size is allowed. Apply `OpenSafeWithOptions` or caller-specific policy after normalization.

### Can `Open` read local files?

Yes. `Open` and `OpenWithOptions` support file URLs and plain paths, so they are for trusted locations only. Safe helpers disable local files by default.

### Why inject `WithLookupIP` in tests?

Safe host validation resolves hostnames to detect private, loopback, and link-local addresses. Injecting a resolver keeps tests hermetic and avoids depending on external DNS.

### Should I allow private hosts?

Only for documented internal calls. For internet-facing input, keep private-host rejection enabled and add `WithAllowedHosts` when only known upstream domains are expected.
