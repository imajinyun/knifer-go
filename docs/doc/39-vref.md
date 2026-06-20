# vref Quickstart

`vref` provides `reflect`-based helpers for type checks, field reads/writes, method lookup, dynamic construction, and function calls.

## Get type and value information

```go
package main

import (
	"fmt"
	"reflect"

	"github.com/imajinyun/go-knifer/vref"
)

func main() {
	type User struct{ Name string }
	u := &User{Name: "alice"}

	typ := vref.TypeOf(u)
	fmt.Println(typ.Kind() == reflect.Pointer)
	fmt.Println(vref.IndirectType(typ).Name())
	fmt.Println(vref.IsNil((*User)(nil)))
}
```

## Read and modify struct fields

```go
package main

import (
	"fmt"

	"github.com/imajinyun/go-knifer/vref"
)

type User struct {
	Name string `json:"name"`
	age  int
}

func main() {
	u := &User{Name: "alice", age: 18}

	fmt.Println(vref.HasField(u, "name"))
	fmt.Println(vref.GetFieldValue(u, "name"))

	if err := vref.SetFieldValue(u, "Name", "bob"); err != nil {
		panic(err)
	}
	fmt.Println(u.Name)

	fmt.Println(vref.GetFieldValueWithOptions(u, "age", vref.WithUnsafeAccess(true)))
}
```

## Find and call methods

```go
package main

import (
	"fmt"
	"reflect"

	"github.com/imajinyun/go-knifer/vref"
)

type Counter struct{}

func (Counter) Add(a, b int) int { return a + b }

func main() {
	c := Counter{}
	method, ok := vref.GetMethod(c, false, "Add", reflect.TypeOf(1), reflect.TypeOf(2))
	if !ok {
		panic("method not found")
	}
	got, err := vref.InvokeMethod(c, method, 2, 3)
	if err != nil {
		panic(err)
	}
	fmt.Println(got)

	got, err = vref.Invoke(c, "Add", 4, 5)
	if err != nil {
		panic(err)
	}
	fmt.Println(got)
}
```

## Dynamically construct values and call functions

```go
package main

import (
	"fmt"

	"github.com/imajinyun/go-knifer/vref"
)

type User struct{ Name string }

func NewUser(name string) User { return User{Name: name} }

func main() {
	created, err := vref.NewInstance(NewUser, "alice")
	if err != nil {
		panic(err)
	}
	fmt.Println(created.(User).Name)

	got, err := vref.InvokeFunc(func(a, b int) int { return a + b }, 6, 7)
	if err != nil {
		panic(err)
	}
	fmt.Println(got)
}
```
