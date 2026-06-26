# vnet Quickstart

`vnet` provides helpers for network addresses, IP/CIDR handling, DNS, connection probes, ports, multipart upload saving, and TLS-related operations.

## Which helper should I use?

| Need | Use | Notes |
| --- | --- | --- |
| Parse or format IP addresses | `IsIP`, `IsIPv4`, `IsIPv6`, `IPv4ToLong`, `LongToIPv4`, `IPv6ToBigInt`, `BigIntToIPv6` | Prefer error-returning helpers when invalid input must be reported. |
| Work with CIDR, ranges, or masks | `BeginIP`, `EndIP`, `CountByMaskBit`, `CountByIPRange`, `ListIPCIDR`, `ListIPRange`, `IsInRange` | Large ranges can produce large result slices; count before listing. |
| Build or inspect socket addresses | `CreateAddress`, `BuildInetSocketAddress`, `AddressOption` helpers | Validate ports before exposing listener or dial targets. |
| Resolve DNS or IDN names | `IDNToASCII`, `GetIPByHostWithOptions`, `GetDNSInfoWithOptions` | Use `WithResolver`, `WithResolveContext`, and `WithResolveTimeout` for deterministic or bounded calls. |
| Probe TCP reachability | `PingWithOptions`, `IsOpenWithOptions`, `ConnectWithOptions`, `NetCatWithOptions` | Always set timeouts and validate untrusted hosts before dialing. |
| Find local ports for tests | `GetUsableLocalPort*`, `NewLocalPortGenerator` | Treat returned ports as hints; another process can bind the port before you use it. |
| Save multipart uploads | `ParseMultipartForm`, `SaveUploadedFile`, `UploadFileName`, `UploadFileSize` | Enforce size limits, extension policy, destination roots, and overwrite behavior at the application boundary. |
| Configure TLS | `CreateTLSConfig`, `NewTLSConfigBuilder`, `AddRootCA*`, `TLSVersion` | Keep TLS 1.2+ minimums for new clients and servers. |

## Network boundary safety checklist

- Treat hostnames, IPs, ports, cookie headers, and uploaded file names from users or upstream systems as untrusted input.
- Bound every DNS lookup and TCP dial with a context or timeout. Avoid defaulting to unbounded network operations in request paths.
- Validate whether private, loopback, link-local, or metadata-service addresses are allowed before dialing user-controlled hosts.
- Do not use `Ping` or `IsOpen` as an authorization, health, or security decision by itself; they only indicate whether a TCP connection can be attempted.
- Count IP ranges before listing them. Expanding broad CIDR ranges can allocate large slices and create denial-of-service risk.
- Treat usable-port helpers as race-prone test utilities. Prefer binding to port `0` and reading the assigned listener address when possible.
- Never trust multipart filenames or content types. Choose the destination path yourself, keep writes inside an allowed root, and disable overwrite unless replacement is intended.
- Do not lower TLS protocol versions for new systems. Use legacy TLS constants only for compatibility with explicitly documented legacy peers.

## When not to use vnet

- Use `vurl`, `vhttp`, or `vresty` safe URL helpers for URL-level SSRF policy, HTTP status handling, headers, redirects, or response body limits.
- Use the standard `net`, `net/netip`, or `net/http` packages directly when you need low-level socket control or protocol-specific behavior not exposed by the facade.
- Use an application upload pipeline when you need antivirus scanning, content sniffing, quarantine, object storage, or tenant-specific storage policy.

## Convert and validate IPs

```go
package main

import (
	"fmt"

	"github.com/imajinyun/knifer-go/vnet"
)

func main() {
	longIP, err := vnet.IPv4ToLong("192.168.1.10")
	if err != nil {
		panic(err)
	}

	fmt.Println(longIP)
	fmt.Println(vnet.LongToIPv4(longIP))
	fmt.Println(vnet.IsIPv4("192.168.1.10"), vnet.IsIPv6("::1"))
}
```

## CIDR, ranges, and wildcards

```go
package main

import (
	"fmt"

	"github.com/imajinyun/knifer-go/vnet"
)

func main() {
	begin, err := vnet.BeginIP("192.168.1.10", 24)
	if err != nil {
		panic(err)
	}
	end, err := vnet.EndIP("192.168.1.10", 24)
	if err != nil {
		panic(err)
	}
	count, err := vnet.CountByIPRange(begin, end)
	if err != nil {
		panic(err)
	}

	fmt.Println(begin, end, count)
	fmt.Println(vnet.IsInRange("192.168.1.20", "192.168.1.0/24"))
	fmt.Println(vnet.MatchesWildcard("192.168.1.*", "192.168.1.20"))
}
```

## Address, DNS, and cookie helpers

```go
package main

import (
	"fmt"
	"time"

	"github.com/imajinyun/knifer-go/vnet"
)

func main() {
	addr, err := vnet.BuildInetSocketAddress("127.0.0.1", 8080)
	if err != nil {
		panic(err)
	}
	fmt.Println(addr.String())

	punycode, err := vnet.IDNToASCII("\u4f8b\u5b50.\u6d4b\u8bd5")
	if err != nil {
		panic(err)
	}
	fmt.Println(punycode)
	fmt.Println(vnet.ParseCookies("sid=abc; theme=dark"))
	fmt.Println(vnet.PingWithOptions("127.0.0.1", vnet.WithPingTimeout(50*time.Millisecond), vnet.WithPingPorts(80, 443)))
}
```

## Connections and multipart upload saving

```go
package main

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"time"

	"github.com/imajinyun/knifer-go/vnet"
)

func main() {
	conn, err := vnet.Connect("127.0.0.1", 80, 100*time.Millisecond)
	if err == nil {
		defer conn.Close()
		fmt.Println(vnet.IsConnected(conn))
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("avatar", "a.txt")
	if err != nil {
		panic(err)
	}
	_, _ = io.WriteString(part, "hello")
	_ = writer.Close()

	req, err := http.NewRequest(http.MethodPost, "/upload", body)
	if err != nil {
		panic(err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	form, err := vnet.ParseMultipartForm(req, vnet.NewUploadSetting())
	if err != nil {
		panic(err)
	}
	file := form.GetFile("avatar")
	fmt.Println(vnet.UploadFileName(file))
	fmt.Println(vnet.SaveUploadedFile(file, "tmp/upload.bin", vnet.WithUploadOverwrite(true)) != nil)
}
```

## Related packages

- Use `vurl` when network checks start from URLs rather than host, IP, or port values.
- Use `vhttp` or `vresty` when network policy is part of outbound HTTP requests.
- Use `vssh` or `vftp` when connectivity checks are tied to transfer protocols.

## Benchmarks and trade-offs

Network behavior depends on DNS, kernel state, local listeners, and remote peers, so cookbook examples prefer deterministic provider injection over universal throughput claims. Use the package tests as the local regression gate:

```bash
go test ./internal/net ./vnet
```

For performance-sensitive code, benchmark the specific operation in your environment: IP math is CPU-bound, DNS and dial helpers are latency-bound, multipart saving is I/O-bound, and TLS setup depends on certificate parsing plus handshake behavior outside these helpers.

## FAQ

### Can I dial a hostname supplied by a user?

Only after applying an application allow-list or private-address rejection policy. `vnet` exposes low-level dial helpers; use `vurl`, `vhttp`, or `vresty` safe helpers when the input is a URL crossing a trust boundary.

### Is a port returned by `GetUsableLocalPort` guaranteed to remain available?

No. It is available at the time of the check, but another process can bind it before the caller does. For tests and servers, binding to port `0` is safer when supported.

### Are uploaded filenames safe to reuse as destination paths?

No. Treat multipart filenames as metadata only. Choose the destination path from trusted application state and keep writes under an allowed directory.

### Should I use TLS 1.0 or TLS 1.1 constants?

Only for documented legacy interoperability. New systems should use `CreateTLSConfig` or a builder configuration that keeps TLS 1.2 or newer as the minimum version.
