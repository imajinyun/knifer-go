# vjson Quickstart

`vjson` provides JSON encoding, parsing, formatting, path reads/writes, object/array wrappers, and conversion helpers between JSON, structs, and XML.

Use `encoding/json` directly when you need full control over streaming, tokenization, or decoder settings. Use `vjson` when the common object, array, formatting, path lookup, or XML bridge helpers reduce boilerplate for your workflow.

## Cookbook

### Encode a struct

```go
type User struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

body, err := vjson.ToStr(User{Name: "go-knifer", Age: 5})
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
if err := vjson.ToBean(`{"name":"go-knifer","age":5}`, &user); err != nil {
	panic(err)
}
fmt.Println(user.Name, user.Age)
```

### Parse into an object and read by path

```go
obj, err := vjson.ParseObj(`{"user":{"name":"go-knifer"}}`)
if err != nil {
	panic(err)
}
fmt.Println(vjson.GetByPath(obj, "user.name"))
fmt.Println(vjson.GetByPathOr(obj, "user.email", "missing"))
```

### Format JSON for humans

```go
pretty := vjson.FormatWithOptions(`{"name":"go-knifer"}`, vjson.WithFormatIndentWidth(2))
fmt.Println(pretty)
```

### Convert between XML and JSON

```go
obj, err := vjson.XMLToJSON(`<user><name>go-knifer</name></user>`)
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

	"github.com/imajinyun/go-knifer/vjson"
)

func main() {
	data := map[string]any{
		"name": "go-knifer",
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

	"github.com/imajinyun/go-knifer/vjson"
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

	"github.com/imajinyun/go-knifer/vjson"
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

	"github.com/imajinyun/go-knifer/vjson"
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
