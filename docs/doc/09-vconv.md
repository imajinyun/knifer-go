# vconv Quickstart

`vconv` provides loose type conversion helpers that convert common inputs to string, int, int64, float64, bool, and []byte, with defaults and custom parse/format options.

## Convert to numbers

```go
package main

import (
	"fmt"

	"github.com/imajinyun/go-knifer/vconv"
)

func main() {
	fmt.Println(vconv.ToInt("42"))
	fmt.Println(vconv.ToIntDefault("bad", 7))
	fmt.Println(vconv.ToFloat64("3.14"))
}
```

## Convert to bool

```go
package main

import (
	"fmt"
	"strings"

	"github.com/imajinyun/go-knifer/vconv"
)

func main() {
	fmt.Println(vconv.ToBool("true"))
	fmt.Println(vconv.ToBoolWithOptions("YES", vconv.WithBoolParser(func(s string) (bool, error) {
		return strings.EqualFold(s, "yes"), nil
	})))
}
```

## Convert to strings

```go
package main

import (
	"fmt"
	"strconv"

	"github.com/imajinyun/go-knifer/vconv"
)

func main() {
	fmt.Println(vconv.ToString(123))
	fmt.Println(vconv.ToStringDefault(nil, "fallback"))
	fmt.Println(vconv.ToStringWithOptions(true, vconv.WithFormatBoolFunc(strconv.FormatBool)))
}
```

## Convert to byte slices

```go
package main

import (
	"fmt"

	"github.com/imajinyun/go-knifer/vconv"
)

func main() {
	b := vconv.ToBytes("hello")
	fmt.Println(string(b))
}
```
