# vsys Quickstart

`vsys` provides system information, Go runtime information, process metrics, environment reads, and system-info dump helpers, with option-based data provider injection.

## Read host, OS, and Go information

```go
package main

import (
	"fmt"
	"runtime"

	"github.com/imajinyun/go-knifer/vsys"
)

func main() {
	host := vsys.SystemHostInfoWithOptions(vsys.WithHostNameFunc(func() (string, error) {
		return "dev-host", nil
	}))
	osInfo := vsys.SystemOsInfo()
	goInfo := vsys.SystemGoInfoWithOptions(vsys.WithGoVersionFunc(runtime.Version))

	fmt.Println(host.Name)
	fmt.Println(osInfo.Name)
	fmt.Println(goInfo.Version != "")
}
```

## Read process and runtime metrics

```go
package main

import (
	"fmt"

	"github.com/imajinyun/go-knifer/vsys"
)

func main() {
	fmt.Println(vsys.CurrentPIDWithOptions(vsys.WithPIDFunc(func() int { return 1234 })))
	fmt.Println(vsys.TotalGoroutineCountWithOptions(vsys.WithProcessNumGoroutineFunc(func() int { return 8 })))
	fmt.Println(vsys.TotalMemory() >= 0)
	fmt.Println(vsys.FreeMemory() >= 0)
}
```

## Read environment variables

```go
package main

import (
	"fmt"

	"github.com/imajinyun/go-knifer/vsys"
)

func main() {
	lookup := func(key string) (string, bool) {
		values := map[string]string{"PORT": "8080", "DEBUG": "true"}
		v, ok := values[key]
		return v, ok
	}

	fmt.Println(vsys.EnvOrDefaultWithOptions("APP", "go-knifer", vsys.WithEnvLookupFunc(lookup)))
	fmt.Println(vsys.EnvIntWithOptions("PORT", 80, vsys.WithEnvLookupFunc(lookup)))
	fmt.Println(vsys.EnvBoolWithOptions("DEBUG", false, vsys.WithEnvLookupFunc(lookup)))
}
```

## Dump system information to a writer

```go
package main

import (
	"bytes"
	"fmt"

	"github.com/imajinyun/go-knifer/vsys"
)

func main() {
	var out bytes.Buffer
	vsys.DumpSystemInfoWithOptions(&out,
		vsys.WithDumpHostOptions(vsys.WithHostNameFunc(func() (string, error) { return "docs", nil })),
	)

	fmt.Println(out.Len() > 0)
}
```
