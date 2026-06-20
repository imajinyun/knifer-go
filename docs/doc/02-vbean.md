# vbean Quickstart

`vbean` maps fields between structs and maps. Use `Copy` / `CopyProperties` for trusted Go-to-Go property copy, `Decode` / `DecodeResult` for weak string/numeric/bool input conversion with metadata, and `Merge` / `MergeResult` when multiple sources should update an existing destination from left to right.

## Convert a struct to a map

```go
package main

import (
	"fmt"

	"github.com/imajinyun/go-knifer/vbean"
)

type UserDTO struct {
	Name string `json:"name"`
	Age  string `json:"age"`
}

func main() {
	m, err := vbean.ToMap(UserDTO{Name: "alice", Age: "18"})
	if err != nil {
		panic(err)
	}
	fmt.Println(m["name"], m["age"])
}
```

## Fill a struct from a map

```go
package main

import (
	"fmt"

	"github.com/imajinyun/go-knifer/vbean"
)

type User struct {
	Name string `json:"full_name"`
	Age  int    `json:"age"`
}

func main() {
	var user User
	err := vbean.ToStruct(map[string]any{"FULL_NAME": "drew", "age": "21"}, &user,
		vbean.WithCaseInsensitive(true),
		vbean.WithWeaklyTyped(true),
	)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%s:%d\n", user.Name, user.Age)
}
```

## Use custom tags and skip zero values

```go
package main

import (
	"fmt"

	"github.com/imajinyun/go-knifer/vbean"
)

type Row struct {
	Name string `db:"user_name"`
	Age  int    `db:"age"`
}

func main() {
	m, err := vbean.ToMap(Row{Name: "casey", Age: 0},
		vbean.WithTagNames("db"),
		vbean.WithIgnoreZero(true),
	)
	if err != nil {
		panic(err)
	}
	fmt.Println(m) // age is skipped
}
```

## Decode weak input with metadata

Use `DecodeResult` when callers need to know which source fields matched the destination and which inputs were unused.

```go
package main

import (
	"fmt"

	"github.com/imajinyun/go-knifer/vbean"
)

type User struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func main() {
	var user User
	result, err := vbean.DecodeResult(map[string]any{"name": "Kai", "age": "34", "extra": true}, &user)
	if err != nil {
		panic(err)
	}

	fmt.Println(user)
	fmt.Println(result.Matched)
	fmt.Println(result.Unused)
}
```

## Copy fields while preserving existing non-empty values

```go
package main

import (
	"fmt"

	"github.com/imajinyun/go-knifer/vbean"
)

type User struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func main() {
	dst := User{Name: "existing", Age: 30}
	err := vbean.Copy(map[string]any{"name": "", "age": "22"}, &dst,
		vbean.WithIgnoreEmpty(true),
	)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%+v\n", dst)
}
```

## Merge multiple sources into an existing value

Use `Merge` when later sources should override earlier sources while preserving unmatched destination fields.

```go
package main

import (
	"fmt"

	"github.com/imajinyun/go-knifer/vbean"
)

type User struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func main() {
	user := User{Name: "existing", Age: 18}
	if err := vbean.Merge(&user, map[string]any{"name": "new"}, map[string]any{"age": "21"}); err != nil {
		panic(err)
	}

	fmt.Println(user)
}
```
