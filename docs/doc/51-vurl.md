# vurl Quickstart

`vurl` provides URL building, parsing, encoding/decoding, query handling, normalization, resource opening, and safe HTTP(S) resource access helpers.

## Build URLs

```go
package main

import (
	"fmt"

	"github.com/imajinyun/go-knifer/vurl"
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

	"github.com/imajinyun/go-knifer/vurl"
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

	"github.com/imajinyun/go-knifer/vurl"
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

	"github.com/imajinyun/go-knifer/vurl"
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
