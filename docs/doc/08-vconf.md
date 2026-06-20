# vconf Quickstart

`vconf` reads, parses, and manages grouped configuration, with support for setting/properties, YAML, TOML, profile overrides, environment expansion, schema validation, and file watching.

## Parse TOML and read grouped values

```go
package main

import (
	"fmt"

	"github.com/imajinyun/go-knifer/vconf"
)

func main() {
	c, err := vconf.ParseTOML(`
name = "demo"
[server]
port = 8080
debug = true
`)
	if err != nil {
		panic(err)
	}

	fmt.Println(c.Get("name"))
	fmt.Println(c.GetIntByGroup("server", "port", 0))
	fmt.Println(c.GetBoolByGroup("server", "debug", false))
}
```

## Expand environment variables

```go
package main

import (
	"fmt"

	"github.com/imajinyun/go-knifer/vconf"
)

func main() {
	c, err := vconf.Parse("base=http://${ENV:HOST}\n")
	if err != nil {
		panic(err)
	}

	value := c.GetExpandedWithOptions("base", vconf.WithEnvLookup(func(name string) string {
		if name == "HOST" {
			return "localhost:8080"
		}
		return ""
	}))
	fmt.Println(value)
}
```

## Bind to a struct

```go
package main

import (
	"fmt"

	"github.com/imajinyun/go-knifer/vconf"
)

type Server struct {
	Port  int      `conf:"port"`
	Debug bool     `conf:"debug"`
	Tags  []string `conf:"tags"`
}

func main() {
	c, err := vconf.ParseTOML(`
[server]
port = 8080
debug = true
tags = ["api", "admin"]
`)
	if err != nil {
		panic(err)
	}

	var server Server
	if err := c.BindGroup("server", &server); err != nil {
		panic(err)
	}
	fmt.Printf("%+v\n", server)
}
```

## Apply profile overrides

```go
package main

import (
	"fmt"

	"github.com/imajinyun/go-knifer/vconf"
)

func main() {
	c, err := vconf.ParseTOML(`
[server]
port = 8080
[profile.prod.server]
port = 9090
`)
	if err != nil {
		panic(err)
	}

	prod := c.ApplyProfile("prod")
	fmt.Println(prod.GetByGroup("server", "port"))
}
```
