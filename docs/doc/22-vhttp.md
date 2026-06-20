# vhttp Quickstart

`vhttp` provides chained HTTP requests, shortcut GET/POST/download helpers, response saving, global/client configuration, safe URL validation, and simple HTTP server wrappers.

## Send chained requests

```go
package main

import (
	"fmt"
	"time"

	"github.com/imajinyun/go-knifer/vhttp"
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

	"github.com/imajinyun/go-knifer/vhttp"
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

	"github.com/imajinyun/go-knifer/vhttp"
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

	"github.com/imajinyun/go-knifer/vhttp"
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
