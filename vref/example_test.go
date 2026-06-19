package vref_test

import (
	"fmt"

	"github.com/imajinyun/go-knifer/vref"
)

func ExampleTypeOf() {
	t := vref.TypeOf("hello")
	fmt.Println(t.Name())
	// Output: string
}

func ExampleIsNil() {
	var s *string
	fmt.Println(vref.IsNil(s))
	fmt.Println(vref.IsNil("hello"))
	// Output:
	// true
	// false
}

func ExampleGetFieldValue() {
	type User struct {
		Name string
	}

	fmt.Println(vref.GetFieldValue(User{Name: "Alice"}, "Name"))
	// Output: Alice
}

func ExampleSetFieldValue() {
	type User struct {
		Name string
	}
	user := User{Name: "Alice"}

	err := vref.SetFieldValue(&user, "Name", "Bob")

	fmt.Println(user.Name)
	fmt.Println(err)
	// Output:
	// Bob
	// <nil>
}

func ExampleInvokeFunc() {
	result, err := vref.InvokeFunc(func(a, b int) int {
		return a + b
	}, 2, 3)

	fmt.Println(result)
	fmt.Println(err)
	// Output:
	// 5
	// <nil>
}
