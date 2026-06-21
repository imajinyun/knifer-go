# vxml Quickstart

`vxml` provides lightweight XML reads/writes, node construction, XPath queries, map/struct conversion, cleanup, formatting, and SAX reading helpers.

## Which helper should I use?

| Goal | Start with | Notes |
| --- | --- | --- |
| Parse XML from a string | `ParseXML` | Use parse options for strictness, namespaces, entity maps, charset readers, and byte limits. |
| Read XML from unknown input | `ReadXML`, `ReadXMLFile`, `ReadXMLBytes`, `ReadXMLReader` | Prefer the source-specific helper when the input type is known; inject `WithOpenFile` in tests. |
| Stream large XML | `ReadBySAX`, `ReadBySAXWithOptions`, `ReadBySAXFileWithOptions` | Processes tokens without building a full document tree. |
| Navigate document nodes | `GetRootElement`, `GetElement`, `GetElements`, `ElementText` | Best for shallow or known document shapes. |
| Query with XPath | `GetElementByXPath`, `GetNodeListByXPath`, `GetNodeByXPath`, `GetByXPath` | Use when paths are dynamic or direct child traversal would be verbose. |
| Build a document tree | `CreateXML`, `CreateXMLWithRoot`, `CreateXMLWithRootNS`, `AppendChild`, `AppendText` | Keeps construction explicit and avoids string concatenation bugs. |
| Serialize XML | `MarshalString`, `WriteTo`, `WriteFile` | Use `WithPretty`, `WithOmitDeclaration`, `WithCharset`, and file options for output contracts. |
| Convert XML to maps or structs | `XMLToMap`, `XMLToMapInto`, `XMLToBean`, `XMLNodeToBean` | Use `Into` helpers to reuse caller-owned result maps. |
| Marshal maps or structs | `MarshalMap`, `MarshalBean`, `TransformWith` | Set `WithRootName` and `WithNamespace` when the output schema requires a stable root. |
| Sanitize or escape content | `CleanInvalid`, `CleanComment`, `Escape`, `Unescape` | Cleaning is a pre-processing step; it does not validate a schema. |

## XML safety checklist

- Bound untrusted input with `WithMaxBytes` before parsing or SAX reading.
- Prefer `WithStrict(true)` when malformed XML should fail fast instead of being normalized.
- Control entity handling with `WithEntity`; do not accept arbitrary entity expansion from untrusted sources.
- Use `WithCharsetReader` for non-UTF-8 inputs instead of assuming the default decoder can read every declared charset.
- Inject `WithOpenFile` and `WithOpenWriteFile` in tests to avoid touching the real filesystem.
- Use `AppendChild`, `AppendText`, `Escape`, and marshal helpers instead of hand-concatenating XML strings.
- Review `WithOverwrite`, `WithCreateParents`, `WithFilePerm`, and `WithDirPerm` before writing files from user-provided paths.

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

## When not to use vxml

- Use `encoding/xml` directly when you only need standard struct marshal/unmarshal behavior and no facade options.
- Use a schema validator when XML must be checked against XSD/DTD rules; `vxml` focuses on parsing, building, traversal, and conversion helpers.
- Use SAX helpers instead of tree parsing for very large documents or streams that do not need random node access.
- Avoid XPath with untrusted user-supplied expressions unless the expression language is constrained by the caller.

## Related packages

- Use `vjson` when XML data must be bridged into JSON-shaped access or serialized back from JSON objects.
- Use `vfile` when XML input/output paths need filesystem policy or temporary-file handling.
- Use `vbean` when parsed XML data should be bound into typed structs after validation.

## Benchmarks and trade-offs

- Tree parsing makes traversal and modification simple, but it holds the whole document in memory. SAX reading reduces memory pressure for large inputs.
- Pretty formatting improves readability but adds indentation work and larger output. Keep compact output for high-volume wire formats.
- Map conversion is flexible for dynamic shapes, while struct conversion gives stronger field contracts when the schema is known.
- File write options improve safety and testability but add setup; direct `WriteTo` is simpler when the caller already owns the destination writer.
- XPath is expressive but more expensive than direct child lookups for known shallow paths.

## FAQ

### Should I parse into a tree or use SAX?

Use tree parsing when you need random access, mutation, XPath, or conversion. Use SAX when the document is large and each token can be handled in one pass.

### How do I make file-based XML tests hermetic?

Inject `WithOpenFile` for reads and `WithOpenWriteFile`/`WithMkdirAll` for writes. This lets tests use in-memory fakes instead of creating real files.

### Does `CleanInvalid` make arbitrary XML safe?

No. It removes configured invalid characters. You still need byte limits, strict parsing, entity policy, and schema/domain validation for untrusted input.

### When should I set `WithRootName`?

Set it when marshaling maps or structs that must produce a specific document root for an external schema or API contract.
