package vref_test

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	"github.com/imajinyun/knifer-go/vref"
)

type exampleUser struct {
	Name string
}

func (u exampleUser) Greet(prefix string) string {
	return prefix + ", " + u.Name
}

func (u exampleUser) GetName() string { return u.Name }

func (u exampleUser) String() string { return u.Name }

type exampleProfile struct {
	ID   int    `json:"id"`
	Name string `ref:"display_name"`
	note string
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

func ExampleIndirectType() {
	t := vref.IndirectType(reflect.TypeOf(&exampleUser{}))
	fmt.Println(t.Name())
	// Output: exampleUser
}

func ExampleValueOf() {
	v := vref.ValueOf(42)
	fmt.Println(v.Kind())
	fmt.Println(v.Int())
	// Output:
	// int
	// 42
}

func ExampleIndirectValue() {
	name := "knifer-go"
	v := vref.IndirectValue(reflect.ValueOf(&name))
	fmt.Println(v.String())
	// Output: knifer-go
}

func ExampleIsFunction() {
	fmt.Println(vref.IsFunction(func() {}))
	fmt.Println(vref.IsFunction("not a function"))
	// Output:
	// true
	// false
}

func ExampleIsCollection() {
	fmt.Println(vref.IsCollection([]string{"a", "b"}))
	fmt.Println(vref.IsCollection(map[string]int{"a": 1}))
	// Output:
	// true
	// false
}

func ExampleIsMap() {
	fmt.Println(vref.IsMap(map[string]int{"a": 1}))
	fmt.Println(vref.IsMap([]int{1, 2}))
	// Output:
	// true
	// false
}

func ExampleImplementsError() {
	t := reflect.TypeOf(errors.New("boom"))
	fmt.Println(vref.ImplementsError(t))
	// Output: true
}

func ExampleImplementsContext() {
	t := reflect.TypeOf(context.Background())
	fmt.Println(vref.ImplementsContext(t))
	// Output: true
}

func ExampleHasField() {
	fmt.Println(vref.HasField(exampleProfile{}, "display_name"))
	fmt.Println(vref.HasField(exampleProfile{}, "missing"))
	// Output:
	// true
	// false
}

func ExampleGetFieldName() {
	field, _ := reflect.TypeOf(exampleProfile{}).FieldByName("ID")
	fmt.Println(vref.GetFieldName(field))
	// Output: id
}

func ExampleGetFieldsValue() {
	profile := exampleProfile{ID: 7, Name: "Alice"}
	values := vref.GetFieldsValue(profile, func(field reflect.StructField) bool {
		return field.IsExported()
	})
	fmt.Println(values)
	// Output: [7 Alice]
}

func ExampleGetPublicMethodNames() {
	fmt.Println(vref.GetPublicMethodNames(exampleUser{}))
	// Output: [GetName Greet String]
}

func ExampleGetMethodByNameIgnoreCase() {
	method, ok := vref.GetMethodByNameIgnoreCase(exampleUser{}, "greet")
	fmt.Println(method.Name)
	fmt.Println(ok)
	// Output:
	// Greet
	// true
}

func ExampleIsGetterOrSetter() {
	method, _ := vref.GetMethodByName(exampleUser{}, "GetName")
	fmt.Println(vref.IsGetterOrSetter(method, false))
	// Output: true
}

func ExampleNewInstance() {
	value, err := vref.NewInstance(exampleProfile{})
	fmt.Printf("%T\n", value)
	fmt.Println(err)
	// Output:
	// vref_test.exampleProfile
	// <nil>
}

func ExampleInvokeRaw() {
	result, err := vref.InvokeRaw(func(name string) string {
		return "hello, " + name
	}, "Alice")
	fmt.Println(result)
	fmt.Println(err)
	// Output:
	// hello, Alice
	// <nil>
}

func ExampleSetFieldValueWithOptions() {
	profile := exampleProfile{note: "draft"}
	err := vref.SetFieldValueWithOptions(&profile, "note", "published", vref.WithUnsafeAccess(true))
	fmt.Println(profile.note)
	fmt.Println(err)
	// Output:
	// published
	// <nil>
}
