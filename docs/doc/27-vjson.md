# vjson Quickstart

`vjson` provides JSON encoding, parsing, formatting, path reads/writes, object/array wrappers, and conversion helpers between JSON, structs, and XML.

Use `encoding/json` directly when you need full control over streaming, tokenization, or decoder settings. Use `vjson` when the common object, array, formatting, path lookup, or XML bridge helpers reduce boilerplate for your workflow.

## Which helper should I use?

Start with `encoding/json` when you need decoder-level control. Use `vjson` when the operation is a common encode/decode, object lookup, path update, formatting, or XML bridge.

| Need | Use | Notes |
| --- | --- | --- |
| Encode a value to compact JSON | `ToStr` | Supports package options such as date formatting for common value-to-string workflows. |
| Encode pretty JSON | `ToStrIndent`, `Format`, `FormatWithOptions` | Use for logs, fixtures, docs, or human review; do not treat formatting as validation by itself. |
| Decode into a struct | `ToBean` | The destination must be a pointer. Keep strict decoding requirements in `encoding/json` if unknown fields must fail. |
| Parse and inspect object fields | `ParseObj`, `JSONObject` getters | Useful when payload shape is partially dynamic but object access should stay readable. |
| Parse arrays | `ParseArray`, `JSONArray` helpers | Use when the top-level JSON value is an array or you need typed element helpers. |
| Read or update nested values | `GetByPath`, `GetByPathOr`, `PutByPath` | Keep path strings close to the schema they describe; missing paths should have explicit defaults. |
| Convert between XML and JSON | `XMLToJSON`, `ToXML` | Good for bridge code. Validate XML inputs separately when they cross a trust boundary. |
| Inject parsing behavior in tests | `ParseObjWithOptions`, `WithParseDecoderFactory` | Keeps tests deterministic without changing global JSON behavior. |

## JSON safety checklist

- Validate payload size before parsing when JSON comes from a network, file upload, queue, or other untrusted source.
- Use `encoding/json.Decoder` directly when you need streaming, `DisallowUnknownFields`, number preservation, token inspection, or multiple JSON values from one stream.
- Treat path lookups as schema-dependent. Prefer `GetByPathOr` when missing fields are acceptable, and fail explicitly when they are required.
- Do not log raw JSON that may contain secrets or personal data; redact before formatting for diagnostics.
- Validate XML inputs before XML/JSON conversion when the XML crosses a trust boundary.
- Keep custom decoder factories local to tests or narrow integration points so parsing behavior remains predictable.

## When not to use vjson

- Use `encoding/json.Decoder` directly for streaming, token-level parsing, multiple JSON values from one stream, `DisallowUnknownFields`, or `UseNumber` behavior.
- Use schema validation or typed request structs when the payload must satisfy strict business rules before use.
- Use specialized high-performance JSON libraries only after benchmarking and documenting their compatibility trade-offs.
- Avoid whole-document helpers for very large or unbounded payloads; stream and bound input size first.
- Use XML-specific parsers and validators when XML semantics, namespaces, attributes, or entity policy matter beyond a small JSON bridge.

## Related packages

- Use `vbean` when decoded JSON needs to be copied or bound into typed structs.
- Use `vconf` when JSON is one configuration format among YAML, TOML, files, or remote sources.
- Use `vxml` when XML semantics, namespaces, attributes, or streaming behavior matter beyond JSON bridging.

## Benchmarks and trade-offs

Use the JSON benchmark suite to measure helper overhead and conversion cost on your machine:

```bash
go test -bench=. -benchmem -run=^$ ./vjson
```

The suite covers object parsing, JSON encoding, path reads, and XML-to-JSON conversion. Treat the output as a local baseline rather than a universal performance claim. For streaming or very large payloads, benchmark the direct `encoding/json.Decoder` approach next to the `vjson` helper you plan to use.

## FAQ

### Does vjson replace encoding/json?

No. `vjson` wraps common workflows. Use `encoding/json` directly when you need streaming, strict decoder options, custom token handling, or maximum control over allocation and numeric behavior.

### Are path helpers schema validation?

No. Path helpers make nested access concise. They do not prove that the entire payload matches a schema; validate required fields and types at your boundary.

### When should I use the XML bridge?

Use it for small bridge workflows where callers need JSON-shaped access to XML data or must serialize a JSON object back to XML. Keep XML-specific validation and trust-boundary checks outside the bridge call.

## Cookbook

### Encode a struct

```go
type User struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

body, err := vjson.ToStr(User{Name: "knifer-go", Age: 5})
if err != nil {
	panic(err)
}
fmt.Println(body)
```

### Decode into a struct

```go
type User struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

var user User
if err := vjson.ToBean(`{"name":"knifer-go","age":5}`, &user); err != nil {
	panic(err)
}
fmt.Println(user.Name, user.Age)
```

### Parse into an object and read by path

```go
obj, err := vjson.ParseObj(`{"user":{"name":"knifer-go"}}`)
if err != nil {
	panic(err)
}
fmt.Println(vjson.GetByPath(obj, "user.name"))
fmt.Println(vjson.GetByPathOr(obj, "user.email", "missing"))
```

### Format JSON for humans

```go
pretty := vjson.FormatWithOptions(`{"name":"knifer-go"}`, vjson.WithFormatIndentWidth(2))
fmt.Println(pretty)
```

### Convert between XML and JSON

```go
obj, err := vjson.XMLToJSON(`<user><name>knifer-go</name></user>`)
if err != nil {
	panic(err)
}
xmlText, err := vjson.ToXML(obj.GetJSONObject("user"), "user")
if err != nil {
	panic(err)
}
fmt.Println(vjson.GetByPath(obj, "user.name"))
fmt.Println(xmlText)
```

### Inject custom parsing behavior for tests

```go
obj, err := vjson.ParseObjWithOptions(`{"n":"ignored"}`, vjson.WithParseDecoderFactory(func(io.Reader) *json.Decoder {
	return json.NewDecoder(strings.NewReader(`{"n":7}`))
}))
if err != nil {
	panic(err)
}
fmt.Println(obj.GetInt("n"))
```

## Encode and format JSON

```go
package main

import (
	"fmt"
	"time"

	"github.com/imajinyun/knifer-go/vjson"
)

func main() {
	data := map[string]any{
		"name": "knifer-go",
		"date": time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
	}

	compact, err := vjson.ToStr(data, vjson.WithDateFormat("2006-01-02"))
	if err != nil {
		panic(err)
	}
	pretty, err := vjson.ToStrIndent(data, 2)
	if err != nil {
		panic(err)
	}

	fmt.Println(compact)
	fmt.Println(pretty)
}
```

## Parse objects and read typed fields

```go
package main

import (
	"fmt"

	"github.com/imajinyun/knifer-go/vjson"
)

func main() {
	obj, err := vjson.ParseObj(`{"user":{"name":"alice"},"age":30,"active":true}`)
	if err != nil {
		panic(err)
	}

	user := obj.GetJSONObject("user")
	fmt.Println(user.GetString("name"))
	fmt.Println(obj.GetInt("age"), obj.GetBool("active"))
}
```

## Read and write with path expressions

```go
package main

import (
	"fmt"

	"github.com/imajinyun/knifer-go/vjson"
)

func main() {
	root, err := vjson.Parse(`{"user":{"name":"alice","roles":["admin"]}}`)
	if err != nil {
		panic(err)
	}

	fmt.Println(vjson.GetByPath(root, "user.name"))
	if err := vjson.PutByPath(root, "user.city", "Shanghai"); err != nil {
		panic(err)
	}
	fmt.Println(vjson.GetByPathOr(root, "user.city", "unknown"))
}
```

## Convert between JSON, structs, and XML

```go
package main

import (
	"fmt"

	"github.com/imajinyun/knifer-go/vjson"
)

type User struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func main() {
	var user User
	if err := vjson.ToBean(`{"name":"alice","age":30}`, &user); err != nil {
		panic(err)
	}
	fmt.Println(user.Name, user.Age)

	obj, err := vjson.XMLToJSON(`<user><name>bob</name></user>`)
	if err != nil {
		panic(err)
	}
	xmlText, err := vjson.ToXML(obj, "root")
	if err != nil {
		panic(err)
	}
	fmt.Println(xmlText)
}
```
