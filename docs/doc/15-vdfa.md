# vdfa Quickstart

`vdfa` provides word-tree based sensitive-word matching and filtering, with support for package-level dictionaries, local matchers, JSON object filtering, and custom character filters.

## Initialize the default dictionary and filter text

```go
package main

import (
	"fmt"

	"github.com/imajinyun/go-knifer/vdfa"
)

func main() {
	vdfa.Init([]string{"badword", "secret"})

	fmt.Println(vdfa.Contains("a badword appears"))
	fmt.Println(vdfa.Filter("keep the secret"))
}
```

## Use an independent WordTree

```go
package main

import (
	"fmt"

	"github.com/imajinyun/go-knifer/vdfa"
)

func main() {
	tree := vdfa.NewWordTree().AddWords("foo", "foobar")

	found, ok := tree.MatchWord("say foobar now")
	if ok {
		fmt.Println(found.Word, found.Start, found.End)
	}
	fmt.Println(tree.MatchAll("foo and foobar"))
}
```

## Customize character filtering rules

```go
package main

import (
	"fmt"

	"github.com/imajinyun/go-knifer/vdfa"
)

func main() {
	tree := vdfa.NewWordTreeWithOptions(
		vdfa.WithCharFilter(func(r rune) bool { return r != '-' }),
	).AddWord("t-io")

	fmt.Println(tree.IsMatch("tio"))
	fmt.Println(vdfa.ContainsWithOptions("a local word", vdfa.WithMatcherWords([]string{"local"})))
}
```

## Filter JSON content in structs

```go
package main

import (
	"fmt"

	"github.com/imajinyun/go-knifer/vdfa"
)

type Payload struct {
	Text string `json:"text"`
}

func main() {
	vdfa.Init([]string{"secret"})

	got, err := vdfa.FilterAny(Payload{Text: "a secret"}, true, nil)
	if err != nil {
		panic(err)
	}
	fmt.Println(got.Text)
}
```
