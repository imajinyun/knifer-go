# vstr Quickstart

`vstr` provides string helpers for blank checks, trimming, substring extraction, splitting, naming-style conversion, emoji/HTML handling, and text similarity calculation.

## Blank checks and defaults

```go
package main

import (
	"fmt"

	"github.com/imajinyun/go-knifer/vstr"
)

func main() {
	fmt.Println(vstr.IsBlank("  "))
	fmt.Println(vstr.HasBlank("name", " "))
	fmt.Println(vstr.DefaultIfBlank("\t", "unknown"))
	fmt.Println(vstr.Trim("  go  "))
}
```

## Substrings, splitting, and padding

```go
package main

import (
	"fmt"

	"github.com/imajinyun/go-knifer/vstr"
)

func main() {
	fmt.Println(vstr.Sub("\u4f60\u597d\u4e16\u754c", 1, 3))
	fmt.Println(vstr.SubBefore("a/b/c", "/", false))
	fmt.Println(vstr.SubAfter("a/b/c", "/", true))
	fmt.Println(vstr.SplitTrim(" a, b ,c ", ","))
	fmt.Println(vstr.PadLeft("7", 3, '0'))
}
```

## Naming styles and prefixes/suffixes

```go
package main

import (
	"fmt"

	"github.com/imajinyun/go-knifer/vstr"
)

func main() {
	fmt.Println(vstr.ToCamelCase("hello_world"))
	fmt.Println(vstr.ToPascalCase("hello-world"))
	fmt.Println(vstr.ToUnderlineCase("HelloWorld"))
	fmt.Println(vstr.AddPrefixIfNot("api", "/"))
	fmt.Println(vstr.RemoveSuffix("main.go", ".go"))
}
```

## Emoji, HTML, and text similarity

```go
package main

import (
	"fmt"

	"github.com/imajinyun/go-knifer/vstr"
)

func main() {
	fmt.Println(vstr.ContainsEmoji("go 🚀"))
	fmt.Println(vstr.RemoveEmoji("go 🚀"))
	fmt.Println(vstr.EscapeHTML("<b>go</b>"))
	fmt.Println(vstr.LevenshteinDistance("kitten", "sitting"))
	fmt.Println(vstr.JaccardSimilarity("go fast", "go faster") > 0)
}
```
