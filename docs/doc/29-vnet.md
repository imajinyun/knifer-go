# vnet Quickstart

`vnet` provides helpers for network addresses, IP/CIDR handling, DNS, connection probes, ports, multipart upload saving, and TLS-related operations.

## Convert and validate IPs

```go
package main

import (
	"fmt"

	"github.com/imajinyun/go-knifer/vnet"
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

	"github.com/imajinyun/go-knifer/vnet"
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

	"github.com/imajinyun/go-knifer/vnet"
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

	"github.com/imajinyun/go-knifer/vnet"
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
