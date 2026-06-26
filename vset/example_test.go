package vset_test

import (
	"fmt"

	"github.com/imajinyun/knifer-go/vset"
)

func ExampleNewString() {
	s := vset.NewString("a", "b", "a")
	fmt.Println(len(s.Members()))
	fmt.Println(s.Contains("a"))
	fmt.Println(s.Contains("c"))
	// Output:
	// 2
	// true
	// false
}

func ExampleNewInt() {
	s := vset.NewInt(1, 2, 2, 3)
	fmt.Println(len(s.Members()))
	fmt.Println(s.Contains(2))
	fmt.Println(s.Contains(4))
	// Output:
	// 3
	// true
	// false
}

func ExampleNew() {
	s := vset.New("red", "blue", "red")

	fmt.Println(len(s.Members()))
	fmt.Println(s.Contains("blue"))
	// Output:
	// 2
	// true
}

func ExampleNew_addRemove() {
	s := vset.New("a")
	s.Add("b")
	s.Remove("a")

	fmt.Println(s.Contains("a"))
	fmt.Println(s.Contains("b"))
	// Output:
	// false
	// true
}

func ExampleSet_Union() {
	left := vset.New(1, 2)
	right := vset.New(2, 3)

	union := left.Union(right)
	intersection := left.Intersect(right)

	fmt.Println(len(union.Members()))
	fmt.Println(intersection.Contains(2))
	fmt.Println(intersection.Contains(1))
	// Output:
	// 3
	// true
	// false
}

func ExampleNewInt32() {
	s := vset.NewInt32(1, 2, 2)

	fmt.Println(len(s.Members()))
	fmt.Println(s.Contains(2))
	// Output:
	// 2
	// true
}

func ExampleNewUint() {
	s := vset.NewUint(1, 2, 2)

	fmt.Println(len(s.Members()))
	fmt.Println(s.Contains(2))
	// Output:
	// 2
	// true
}

func ExampleNewInt64() {
	s := vset.NewInt64(1, 2, 2)

	fmt.Println(len(s.Members()))
	fmt.Println(s.Contains(1))
	// Output:
	// 2
	// true
}

func ExampleNewUint32() {
	s := vset.NewUint32(1, 2, 2)

	fmt.Println(len(s.Members()))
	fmt.Println(s.Contains(2))
	// Output:
	// 2
	// true
}

func ExampleNewUint64() {
	s := vset.NewUint64(1, 2, 2)

	fmt.Println(len(s.Members()))
	fmt.Println(s.Contains(3))
	// Output:
	// 2
	// false
}

func ExampleWithSetMarshalFunc() {
	s := vset.New("red")
	data, err := s.MarshalJSONWithOptions(vset.WithSetMarshalFunc(func(any) ([]byte, error) {
		return []byte(`["custom"]`), nil
	}))

	fmt.Println(string(data))
	fmt.Println(err)
	// Output:
	// ["custom"]
	// <nil>
}

func ExampleWithSetUnmarshalFunc() {
	var s vset.Set[string]
	err := s.UnmarshalJSONWithOptions([]byte(`ignored`), vset.WithSetUnmarshalFunc(func(_ []byte, v any) error {
		items := v.(*[]string)
		*items = []string{"red", "red", "blue"}
		return nil
	}))

	fmt.Println(err)
	fmt.Println(len(s.Members()))
	fmt.Println(s.Contains("blue"))
	// Output:
	// <nil>
	// 2
	// true
}
