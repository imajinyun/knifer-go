# vver Quickstart

`vver` provides version comparison and version-expression matching helpers for checking whether a current version satisfies single, multiple, or custom-delimiter expressions.

## Compare two versions

```go
package main

import (
	"fmt"

	"github.com/imajinyun/go-knifer/vver"
)

func main() {
	fmt.Println(vver.CompareVersion("1.0.0", "1.0.2") < 0)
	fmt.Println(vver.CompareVersion("1.2.0", "1.2.0") == 0)
	fmt.Println(vver.CompareVersion("2.0.0", "1.9.9") > 0)
}
```

## Use relational predicate helpers

```go
package main

import (
	"fmt"

	"github.com/imajinyun/go-knifer/vver"
)

func main() {
	fmt.Println(vver.IsGreaterThan("1.0.3", "1.0.2"))
	fmt.Println(vver.IsGreaterThanOrEqual("1.0.2", "1.0.2"))
	fmt.Println(vver.IsLessThan("1.0.1", "1.0.2"))
	fmt.Println(vver.IsLessThanOrEqual("1.0.2", "1.0.2"))
}
```

## Match version expressions

```go
package main

import (
	"fmt"

	"github.com/imajinyun/go-knifer/vver"
)

func main() {
	fmt.Println(vver.MatchEl("1.0.2", ">=1.0.0"))
	fmt.Println(vver.MatchEl("1.0.2", "1.0.1-1.1.0"))
	fmt.Println(vver.MatchElWithDelimiter("1.0.2", "<1.0.1,1.0.2", ","))
}
```

## Match any of multiple expressions

```go
package main

import (
	"fmt"

	"github.com/imajinyun/go-knifer/vver"
)

func main() {
	fmt.Println(vver.AnyMatch("1.0.2", "<1.0.1", "1.0.2"))
	fmt.Println(vver.AnyMatchSlice("1.0.2", []string{"<1.0.1", ">=1.0.0"}))
	fmt.Println(vver.MatchElWithDelimiterErr("1.0.2", ">=1.0.0", ";") == nil)
}
```
