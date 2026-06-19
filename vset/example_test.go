package vset_test

import (
	"fmt"

	"github.com/imajinyun/go-knifer/vset"
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
