# vxml Quickstart

`vxml` provides lightweight XML reads/writes, node construction, XPath queries, map/struct conversion, cleanup, formatting, and SAX reading helpers.

## Parse XML and read nodes

```go
package main

import (
	"fmt"
	"strings"

	"github.com/imajinyun/go-knifer/vxml"
)

func main() {
	doc, err := vxml.ParseXML(`<root><name>go</name><item>1</item><item>2</item></root>`)
	if err != nil {
		panic(err)
	}

	root := vxml.GetRootElement(doc)
	fmt.Println(vxml.ElementText(root, "name"))
	fmt.Println(len(vxml.GetElements(root, "item")))
	fmt.Println(strings.TrimSpace(vxml.GetElementByXPath("/root/item", doc).Text))
}
```

## Build and serialize XML

```go
package main

import (
	"fmt"

	"github.com/imajinyun/go-knifer/vxml"
)

func main() {
	doc := vxml.CreateXMLWithRoot("user")
	root := vxml.GetRootElement(doc)
	vxml.AppendText(vxml.AppendChild(root, "name"), "alice")
	vxml.AppendText(vxml.AppendChild(root, "age"), 30)

	out, err := vxml.MarshalString(doc, vxml.WithOmitDeclaration(true), vxml.WithPretty())
	if err != nil {
		panic(err)
	}
	fmt.Println(out)
}
```

## Convert between XML, maps, and structs

```go
package main

import (
	"fmt"

	"github.com/imajinyun/go-knifer/vxml"
)

type User struct {
	Name string `json:"name"`
}

func main() {
	m, err := vxml.XMLToMap(`<user><name>alice</name></user>`)
	if err != nil {
		panic(err)
	}
	fmt.Println(m["user"] != nil)

	out, err := vxml.MarshalMap(map[string]any{"name": "bob"}, vxml.WithRootName("user"), vxml.WithOmitDeclaration(true))
	if err != nil {
		panic(err)
	}
	fmt.Println(out)
}
```

## Clean, format, and read with SAX

```go
package main

import (
	"encoding/xml"
	"fmt"
	"strings"

	"github.com/imajinyun/go-knifer/vxml"
)

func main() {
	clean := vxml.CleanComment(`<root><!--skip--><a>1</a></root>`)
	formatted, err := vxml.Format(clean)
	if err != nil {
		panic(err)
	}
	fmt.Println(strings.Contains(formatted, "<a>"))

	count := 0
	err = vxml.ReadBySAX(strings.NewReader(clean), func(token xml.Token) error {
		if _, ok := token.(xml.StartElement); ok {
			count++
		}
		return nil
	})
	if err != nil {
		panic(err)
	}
	fmt.Println(count)
}
```
