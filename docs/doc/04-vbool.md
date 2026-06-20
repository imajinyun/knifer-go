# vbool Quickstart

`vbool` provides lightweight boolean helpers for negation, integer conversion, and batch AND/OR checks.

## Negate a bool

```go
package main

import (
	"fmt"

	"github.com/imajinyun/go-knifer/vbool"
)

func main() {
	fmt.Println(vbool.Negate(true))
}
```

## Convert bool to int

```go
package main

import (
	"fmt"

	"github.com/imajinyun/go-knifer/vbool"
)

func main() {
	fmt.Println(vbool.ToInt(true))
	fmt.Println(vbool.ToInt(false))
}
```

## Batch logical AND

```go
package main

import (
	"fmt"

	"github.com/imajinyun/go-knifer/vbool"
)

func main() {
	fmt.Println(vbool.And(true, true, false))
}
```

## Batch logical OR

```go
package main

import (
	"fmt"

	"github.com/imajinyun/go-knifer/vbool"
)

func main() {
	fmt.Println(vbool.Or(false, false, true))
}
```
