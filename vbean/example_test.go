package vbean_test

import (
	"fmt"

	"github.com/imajinyun/knifer-go/vbean"
)

func ExampleToMap() {
	type User struct {
		Name string
		Age  int
	}

	u := User{Name: "Alice", Age: 30}
	m, _ := vbean.ToMap(u)
	fmt.Println(m["Name"], m["Age"])
	// Output: Alice 30
}

func ExampleToStruct() {
	m := map[string]any{"name": "Bob", "age": 25}

	type User struct {
		Name string
		Age  int
	}

	var u User
	_ = vbean.ToStruct(m, &u)
	fmt.Printf("%s is %d", u.Name, u.Age)
	// Output: Bob is 25
}

func ExampleCopyProperties() {
	type Source struct {
		Name string
		Age  int
	}
	type Target struct {
		Name string
		Age  int
	}

	var dst Target
	_ = vbean.CopyProperties(Source{Name: "Carol", Age: 28}, &dst)
	fmt.Println(dst.Name, dst.Age)
	// Output: Carol 28
}

func ExampleFillMap() {
	type User struct {
		Name string
		Age  int
	}

	dst := map[string]any{}
	err := vbean.FillMap(User{Name: "Dana", Age: 31}, dst)

	fmt.Println(dst["Name"], dst["Age"])
	fmt.Println(err)
	// Output:
	// Dana 31
	// <nil>
}

func ExampleToStruct_withOptions() {
	type User struct {
		Name string
		Age  int
	}

	var u User
	err := vbean.ToStruct(
		map[string]any{"NAME": "Drew", "AGE": "21"},
		&u,
		vbean.WithCaseInsensitive(true),
		vbean.WithWeaklyTyped(true),
	)

	fmt.Println(u.Name, u.Age)
	fmt.Println(err)
	// Output:
	// Drew 21
	// <nil>
}

func ExampleCopy() {
	type Source struct {
		Name string
	}
	type Target struct {
		Name string
	}

	var dst Target
	err := vbean.Copy(Source{Name: "Eve"}, &dst)

	fmt.Println(dst.Name)
	fmt.Println(err)
	// Output:
	// Eve
	// <nil>
}

func ExampleDecodeResult() {
	type User struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	var user User
	result, err := vbean.DecodeResult(map[string]any{"name": "Kai", "age": "34", "extra": true}, &user)

	fmt.Println(user.Name, user.Age)
	fmt.Println(result.Matched)
	fmt.Println(result.Unused)
	fmt.Println(err)
	// Output:
	// Kai 34
	// [age name]
	// [extra]
	// <nil>
}

func ExampleMerge() {
	type User struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	user := User{Name: "old", Age: 18}
	err := vbean.Merge(&user, map[string]any{"name": "new"}, map[string]any{"age": "21"})

	fmt.Println(user.Name, user.Age)
	fmt.Println(err)
	// Output:
	// new 21
	// <nil>
}
