# vjson Quickstart

`vjson` provides JSON encoding, parsing, formatting, path reads/writes, object/array wrappers, and conversion helpers between JSON, structs, and XML.

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
