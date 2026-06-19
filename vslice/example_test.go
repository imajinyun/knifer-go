package vslice_test

import (
	"fmt"

	"github.com/imajinyun/go-knifer/vslice"
)

func ExampleDistinct() {
	fmt.Println(vslice.Distinct([]int{1, 2, 2, 3, 3, 3}))
	// Output: [1 2 3]
}

func ExampleContains() {
	fmt.Println(vslice.Contains([]string{"go", "rust"}, "go"))
	// Output: true
}

func ExampleMap() {
	doubled := vslice.Map([]int{1, 2, 3}, func(n int) int { return n * 2 })
	fmt.Println(doubled)
	// Output: [2 4 6]
}

func ExampleFilter() {
	even := vslice.Filter([]int{1, 2, 3, 4}, func(n int) bool { return n%2 == 0 })
	fmt.Println(even)
	// Output: [2 4]
}

func ExampleMapErr() {
	lengths, err := vslice.MapErr([]string{"go", "knifer"}, func(s string) (int, error) {
		return len(s), nil
	})
	fmt.Println(lengths, err)
	// Output: [2 6] <nil>
}

func ExampleFilterErr() {
	odd, err := vslice.FilterErr([]int{1, 2, 3}, func(n int) (bool, error) {
		return n%2 == 1, nil
	})
	fmt.Println(odd, err)
	// Output: [1 3] <nil>
}

func ExampleReduceErr() {
	sum, err := vslice.ReduceErr([]int{1, 2, 3}, 0, func(acc, n int) (int, error) {
		return acc + n, nil
	})
	fmt.Println(sum, err)
	// Output: 6 <nil>
}

func ExampleWindow() {
	fmt.Println(vslice.Window([]int{1, 2, 3, 4}, 3))
	// Output: [[1 2 3] [2 3 4]]
}

func ExampleSliding() {
	fmt.Println(vslice.Sliding([]int{1, 2, 3, 4, 5}, 2, 2))
	// Output: [[1 2] [3 4]]
}

func ExampleZip2() {
	pairs := vslice.Zip2([]string{"a", "b"}, []int{1, 2, 3})
	fmt.Println(pairs)
	// Output: [{a 1} {b 2}]
}

func ExampleUnzip2() {
	left, right := vslice.Unzip2([]vslice.Pair[string, int]{{First: "a", Second: 1}, {First: "b", Second: 2}})
	fmt.Println(left, right)
	// Output: [a b] [1 2]
}

func ExampleUnion() {
	fmt.Println(vslice.Union([]int{1, 2}, []int{2, 3}))
	// Output: [1 2 3]
}

func ExampleIter() {
	for value := range vslice.Iter([]string{"go", "knifer"}) {
		fmt.Println(value)
	}
	// Output:
	// go
	// knifer
}

func ExampleIterIndexed() {
	for index, value := range vslice.IterIndexed([]string{"go", "knifer"}) {
		fmt.Printf("%d:%s\n", index, value)
	}
	// Output:
	// 0:go
	// 1:knifer
}
