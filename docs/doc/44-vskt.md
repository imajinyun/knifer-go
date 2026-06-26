# vskt Quickstart

`vskt` provides TCP socket facades for connections, NIO/AIO clients and servers, socket configuration, session reads/writes, and socket error wrapping.

## Which helper should I use?

| Need | Use | Notes |
| --- | --- | --- |
| Build socket configuration | `NewSocketConfig`, `NewSocketConfigWithOptions`, `WithReadTimeout`, `WithWriteTimeout`, `WithReadBufferSize`, `WithWriteBufferSize` | Timeout options are in milliseconds for config values. |
| Inject deterministic providers | `WithClock`, `WithRunner`, `WithListenerFactory`, `WithConnFactory`, `WithSocketIPParser`, `WithConnectDialer` | Use these in tests to avoid real network dependencies. |
| Open a TCP connection | `SocketConnectWithOptions`, `SocketConnectAddrWithOptions`, `ChannelDialWithOptions` | Always set a timeout or context for external peers. |
| Create NIO-style endpoints | `NewNioServer*`, `NewNioClient*`, `ChannelHandler`, `ChannelHandlerFunc` | Use for event-style TCP handling with explicit channel operations. |
| Create AIO-style endpoints | `NewAioServer*`, `NewAioClient*`, `NewAioSession`, `SimpleIoAction` | Use when callbacks should process async read results. |
| Encode or decode framed messages | `FuncEncoder`, `FuncDecoder`, `Protocol` | Keep framing and size limits in the protocol layer. |
| Inspect connections | `SocketIsConnected`, `SocketRemoteAddress`, `GetRemoteAddress` | `SocketIsConnected` only checks that the connection object is non-nil. |
| Wrap socket errors | `NewSocketError`, `NewSocketErrorMsg`, `NewSocketErrorf`, `WrapSocketError` | Preserve context while keeping socket failures recognizable. |

## Socket safety checklist

- Always configure dial, read, and write timeouts for peers outside deterministic tests.
- Close clients, sessions, listeners, and `net.Conn` values when ownership ends. Do not rely on garbage collection for socket cleanup.
- Treat `SocketIsConnected` as a lightweight nil check, not as proof that the peer is healthy or authenticated.
- Validate bind addresses and ports before starting servers. Binding to all interfaces exposes the listener beyond localhost.
- Bound message sizes in your protocol decoder. TCP is a stream, so application framing must defend against oversized or partial messages.
- Avoid launching unbounded goroutines through custom runners. Use a bounded worker pool or a test runner that executes deterministically.
- Do not read and write shared buffers concurrently unless the protocol or session type documents synchronization.

## When not to use vskt

- Use `vnet` for IP math, DNS, simple TCP probes, TLS config helpers, and multipart upload handling.
- Use `net/http`, `vhttp`, or `vresty` for HTTP protocols; raw sockets do not provide HTTP redirects, headers, status handling, or URL safety policy.
- Use the standard `net` package directly when you need UDP, Unix sockets, TLS handshakes, custom `net.Dialer` fields, or platform-specific socket options.

## Create socket configurations

```go
package main

import (
	"fmt"
	"time"

	"github.com/imajinyun/knifer-go/vskt"
)

func main() {
	cfg := vskt.NewSocketConfigWithOptions(
		vskt.WithReadTimeout(1_000),
		vskt.WithWriteTimeout(1_000),
		vskt.WithReadBufferSize(4*1024),
		vskt.WithWriteBufferSize(4*1024),
		vskt.WithClock(time.Now),
	)

	fmt.Println(cfg != nil)
}
```

## Connect to TCP addresses

```go
package main

import (
	"fmt"
	"time"

	"github.com/imajinyun/knifer-go/vskt"
)

func main() {
	conn, err := vskt.SocketConnectWithOptions(
		"127.0.0.1",
		8080,
		vskt.WithConnectNetwork("tcp"),
		vskt.WithConnectTimeout(2*time.Second),
	)
	if err != nil {
		fmt.Println("connect failed:", err)
		return
	}
	defer conn.Close()

	fmt.Println(vskt.SocketIsConnected(conn))
	fmt.Println(vskt.SocketRemoteAddress(conn))
}
```

## Create AIO sessions with net.Pipe

```go
package main

import (
	"bytes"
	"fmt"
	"net"

	"github.com/imajinyun/knifer-go/vskt"
)

func main() {
	client, server := net.Pipe()
	defer client.Close()
	defer server.Close()

	action := &vskt.SimpleIoAction{
		OnDoAction: func(session *vskt.AioSession, data *bytes.Buffer) {
			fmt.Println(data.String())
		},
	}
	session := vskt.NewAioSession(client, action, vskt.NewSocketConfig())

	_, _ = server.Write([]byte("ping"))
	session.Read()
}
```

## Wrap socket errors

```go
package main

import (
	"errors"
	"fmt"

	"github.com/imajinyun/knifer-go/vskt"
)

func main() {
	err := vskt.WrapSocketError(errors.New("dial refused"), "connect server")
	fmt.Println(err != nil)
	fmt.Println(vskt.NewSocketErrorMsg("closed") != nil)
}
```

## Related packages

- Use `vnet` when socket behavior depends on host, port, IP, or network-boundary checks.
- Use `vhttp` or `vresty` when the protocol is HTTP and request/response helpers are more appropriate.
- Use `vcli` when socket diagnostics are driven by external command execution in tools.

## Benchmarks and trade-offs

Socket throughput and latency depend on kernel buffers, scheduling, network path, and protocol framing, so the cookbook does not publish universal benchmark numbers. Use deterministic tests for facade behavior:

```bash
go test ./internal/socket ./vskt
```

When measuring production protocols, benchmark the complete framing and handler path with realistic payload sizes. Smaller buffers can reduce memory per connection but increase syscall pressure; larger buffers can improve throughput but raise per-connection memory cost.

## FAQ

### Does `SocketIsConnected` verify the remote peer is still alive?

No. It reports whether the connection value is non-nil. Use application heartbeats, deadlines, reads/writes, or protocol acknowledgements to detect dead peers.

### Should tests open real TCP ports?

Prefer provider injection or `net.Pipe` for unit tests. Real ports are useful for integration tests but can be flaky because port availability and timing depend on the host environment.

### Are AIO callbacks safe for heavy work?

Keep callbacks short or hand work to a bounded worker pool. Long-running callbacks can delay socket progress and unbounded goroutine runners can create resource pressure.

### Does vskt add authentication or encryption?

No. It provides raw TCP socket helpers. Add authentication, authorization, encryption, and message integrity at the protocol or transport layer required by your application.
