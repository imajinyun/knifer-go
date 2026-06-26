package vslice_test

import (
	"fmt"

	"github.com/imajinyun/knifer-go/vslice"
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

func ExampleAssociate() {
	lengths := vslice.Associate([]string{"go", "rust"}, func(s string) (string, int) {
		return s, len(s)
	})
	fmt.Println(lengths["go"], lengths["rust"])
	// Output: 2 4
}

func ExampleChunk() {
	fmt.Println(vslice.Chunk([]int{1, 2, 3, 4, 5}, 2))
	// Output: [[1 2] [3 4] [5]]
}

func ExampleCompact() {
	fmt.Println(vslice.Compact([]int{0, 1, 0, 2, 3}))
	// Output: [1 2 3]
}

func ExampleConcat() {
	fmt.Println(vslice.Concat([]string{"go"}, []string{"knifer", "tools"}))
	// Output: [go knifer tools]
}

func ExampleCountBy() {
	counts := vslice.CountBy([]string{"go", "js", "rust", "java"}, func(s string) int {
		return len(s)
	})
	fmt.Println(counts[2], counts[4])
	// Output: 2 2
}

func ExampleFilterMap() {
	names := vslice.FilterMap([]int{1, 2, 3, 4}, func(n int) (string, bool) {
		if n%2 == 0 {
			return fmt.Sprintf("n%d", n), true
		}
		return "", false
	})
	fmt.Println(names)
	// Output: [n2 n4]
}

func ExampleFind() {
	value, ok := vslice.Find([]int{1, 2, 3, 4}, func(n int) bool { return n > 2 })
	fmt.Println(value, ok)
	// Output: 3 true
}

func ExampleFindIndex() {
	fmt.Println(vslice.FindIndex([]string{"go", "rust", "java"}, func(s string) bool { return len(s) == 4 }))
	// Output: 1
}

func ExampleFlatMap() {
	fmt.Println(vslice.FlatMap([]int{1, 2, 3}, func(n int) []int { return []int{n, -n} }))
	// Output: [1 -1 2 -2 3 -3]
}

func ExampleFlatten() {
	fmt.Println(vslice.Flatten([][]string{{"go", "rust"}, {}, {"java"}}))
	// Output: [go rust java]
}

func ExampleForEach() {
	seen := []int{}
	vslice.ForEach([]int{1, 2, 3}, func(n int) { seen = append(seen, n*10) })
	fmt.Println(seen)
	// Output: [10 20 30]
}

func ExampleGroupBy() {
	groups := vslice.GroupBy([]string{"go", "js", "rust", "java"}, func(s string) int {
		return len(s)
	})
	fmt.Println(groups[2], groups[4])
	// Output: [go js] [rust java]
}

func ExampleIndexOf() {
	fmt.Println(vslice.IndexOf([]string{"go", "rust", "go"}, "rust"))
	// Output: 1
}

func ExampleIntersection() {
	fmt.Println(vslice.Intersection([]int{1, 2, 2, 3}, []int{2, 3, 4}))
	// Output: [2 3]
}

func ExampleIsEmpty() {
	fmt.Println(vslice.IsEmpty([]int{}), vslice.IsEmpty([]int{1}))
	// Output: true false
}

func ExampleIsNotEmpty() {
	fmt.Println(vslice.IsNotEmpty([]int{1}), vslice.IsNotEmpty([]int{}))
	// Output: true false
}

func ExampleJoin() {
	fmt.Println(vslice.Join([]int{1, 2, 3}, ":"))
	// Output: 1:2:3
}

func ExampleKeyBy() {
	byLength := vslice.KeyBy([]string{"go", "js", "rust", "java"}, func(s string) int {
		return len(s)
	})
	fmt.Println(byLength[2], byLength[4])
	// Output: js java
}

func ExampleLastIndexOf() {
	fmt.Println(vslice.LastIndexOf([]string{"go", "rust", "go"}, "go"))
	// Output: 2
}

func ExamplePage() {
	fmt.Println(vslice.Page([]int{1, 2, 3, 4, 5}, 2, 2))
	// Output: [3 4]
}

func ExamplePartitionBy() {
	fmt.Println(vslice.PartitionBy([]int{1, 3, 2, 4, 5}, func(n int) bool { return n%2 == 0 }))
	// Output: [[1 3] [2 4] [5]]
}

func ExampleReduce() {
	sum := vslice.Reduce([]int{1, 2, 3}, 0, func(acc, n int) int { return acc + n })
	fmt.Println(sum)
	// Output: 6
}

func ExampleReject() {
	fmt.Println(vslice.Reject([]int{1, 2, 3, 4}, func(n int) bool { return n%2 == 0 }))
	// Output: [1 3]
}

func ExampleReverse() {
	values := []int{1, 2, 3}
	fmt.Println(vslice.Reverse(values))
	// Output: [3 2 1]
}

func ExampleSliceToMap() {
	lengths := vslice.SliceToMap([]string{"go", "rust"}, func(s string) (string, int) {
		return s, len(s)
	})
	fmt.Println(lengths["go"], lengths["rust"])
	// Output: 2 4
}

func ExampleSub() {
	fmt.Println(vslice.Sub([]string{"a", "b", "c", "d"}, 1, 3))
	// Output: [b c]
}

func ExampleSubtract() {
	fmt.Println(vslice.Subtract([]int{1, 2, 3, 2}, []int{2}))
	// Output: [1 3]
}

func ExampleUniq() {
	fmt.Println(vslice.Uniq([]string{"go", "go", "rust", "go"}))
	// Output: [go rust]
}

func ExampleUniqBy() {
	fmt.Println(vslice.UniqBy([]string{"go", "js", "rust", "java"}, func(s string) int { return len(s) }))
	// Output: [go rust]
}
