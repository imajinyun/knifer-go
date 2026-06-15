# vrand Quickstart

`vrand` provides random integer, float, bool, string, slice element, and byte generators, with support for per-call random source injection.

## Generate numbers and booleans

```go
package main

import (
	"fmt"

	"github.com/imajinyun/go-knifer/vrand"
)

func main() {
	fmt.Println(vrand.Int(10))
	fmt.Println(vrand.IntRange(10, 20))
	fmt.Println(vrand.Long())
	fmt.Println(vrand.Float())
	fmt.Println(vrand.Bool())
}
```

## Generate random strings

```go
package main

import (
	"fmt"

	"github.com/imajinyun/go-knifer/vrand"
)

func main() {
	fmt.Println(vrand.String(8))
	fmt.Println(vrand.Numbers(6))
	fmt.Println(vrand.StringUpper(8))
	fmt.Println(vrand.StringFrom("ABC", 4))
}
```

## Choose random elements from slices

```go
package main

import (
	"fmt"

	"github.com/imajinyun/go-knifer/vrand"
)

func main() {
	items := []string{"go", "knifer", "tool"}
	fmt.Println(vrand.Ele(items))
}
```

## Generate secure random bytes and reproducible results

```go
package main

import (
	"fmt"
	mathrand "math/rand"

	"github.com/imajinyun/go-knifer/vrand"
)

func main() {
	b, err := vrand.SecureBytes(16)
	if err != nil {
		panic(err)
	}
	fmt.Println(len(b))

	source := mathrand.New(mathrand.NewSource(1))
	fmt.Println(vrand.IntWithOptions(100, vrand.WithRandomSource(source)))
}
```
