# vregex Quickstart

`vregex` provides regex matching, finding, capture-group extraction, replacement, deletion, template-variable extraction, and special-character escaping helpers.

## Match, find, and count

```go
package main

import (
	"fmt"

	"github.com/imajinyun/go-knifer/vregex"
)

func main() {
	fmt.Println(vregex.Match(`\d+`, "abc123"))
	fmt.Println(vregex.IsMatch(`\d+`, "123"))
	fmt.Println(vregex.FindAll(`\d+`, "a1b22"))
	fmt.Println(vregex.Count(`\d+`, "a1b22"))
}
```

## Extract capture groups and named groups

```go
package main

import (
	"fmt"

	"github.com/imajinyun/go-knifer/vregex"
)

func main() {
	fmt.Println(vregex.GetGroup1(`user:(\w+)`, "user:alice"))
	fmt.Println(vregex.Get(`(a)(b)`, "ab", 2))
	fmt.Println(vregex.GetByName(`(?<word>\w+)-(?<num>\d+)`, "abc-123", "num"))
	fmt.Println(vregex.FindAllGroup(`x(\d+)`, "x1x22", 1))
}
```

## Replace and delete matches

```go
package main

import (
	"fmt"
	"strings"

	"github.com/imajinyun/go-knifer/vregex"
)

func main() {
	fmt.Println(vregex.Replace(`\d+`, "a1b22", "#"))
	fmt.Println(vregex.ReplaceFirst(`\d+`, "a1b22", "#"))
	fmt.Println(vregex.ReplaceAllFunc("a1b22", `\d+`, func(m vregex.MatchResult) string {
		return strings.Repeat("*", len(m.Text))
	}))
	fmt.Println(vregex.DelAll(`\d+`, "a1b22"))
}
```

## Template extraction, DotAll, and escaping

```go
package main

import (
	"fmt"

	"github.com/imajinyun/go-knifer/vregex"
)

func main() {
	fmt.Println(vregex.ExtractMulti(`(\w+)-(\d+)`, "abc-123", "$2:$1"))
	fmt.Println(vregex.TemplateVars("$2:$1"))
	fmt.Println(vregex.FindWithOptions(`a.*c`, "a\nb\nc", vregex.WithDotAll(true)))
	fmt.Println(vregex.Escape("a+b*c?"))
}
```
