package vref_test

import (
	"fmt"

	"github.com/imajinyun/go-knifer/vref"
)

type exampleUser struct {
	Name string
}

func (u exampleUser) Greet(prefix string) string {
	return prefix + ", " + u.Name
}

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
	fmt.Println(vref.GetFieldValue(exampleUser{Name: "Alice"}, "Name"))
	// Output: Alice
}

func ExampleSetFieldValue() {
	user := exampleUser{Name: "Alice"}

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

func ExampleInvoke() {
	result, err := vref.Invoke(&exampleUser{Name: "Alice"}, "Greet", "hello")

	fmt.Println(result)
	fmt.Println(err)
	// Output:
	// hello, Alice
	// <nil>
}

func ExampleGetPublicFieldNames() {
	fmt.Println(vref.GetPublicFieldNames(exampleUser{}))
	// Output: [Name]
}
