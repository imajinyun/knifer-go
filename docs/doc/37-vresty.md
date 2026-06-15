# vresty Quickstart

`vresty` is a resty-based HTTP client facade that provides chained requests, shortcut GET/POST helpers, response reads, downloads, safe URL validation, and global configuration.

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
