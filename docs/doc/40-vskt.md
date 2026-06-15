# vskt Quickstart

`vskt` provides TCP socket facades for connections, NIO/AIO clients and servers, socket configuration, session reads/writes, and socket error wrapping.

## Create socket configurations

```go
package main

import (
	"fmt"
	"time"

	"github.com/imajinyun/go-knifer/vskt"
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

	"github.com/imajinyun/go-knifer/vskt"
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

	"github.com/imajinyun/go-knifer/vskt"
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

	"github.com/imajinyun/go-knifer/vskt"
)

func main() {
	err := vskt.WrapSocketError(errors.New("dial refused"), "connect server")
	fmt.Println(err != nil)
	fmt.Println(vskt.NewSocketErrorMsg("closed") != nil)
}
```
