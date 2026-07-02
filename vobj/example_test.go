package vobj_test

import (
	"fmt"

	"github.com/imajinyun/knifer-go/vobj"
)

type exampleObject struct {
	Name   string
	Scores []int
}

func ExampleEqual() {
	fmt.Println(vobj.Equal(42, 42))
	fmt.Println(vobj.Equal(42, 43))
	// Output:
	// true
	// false
}

func ExampleIsNil() {
	fmt.Println(vobj.IsNil(nil))
	fmt.Println(vobj.IsNil(0))
	// Output:
	// true
	// false
}

func ExampleIsEmpty() {
	fmt.Println(vobj.IsEmpty(""))
	fmt.Println(vobj.IsEmpty([]int{1}))
	// Output:
	// true
	// false
}

func ExampleContains() {
	fmt.Println(vobj.Contains([]string{"go", "knifer"}, "go"))
	fmt.Println(vobj.Contains([]string{"go", "knifer"}, "java"))
	// Output:
	// true
	// false
}

func ExampleLength() {
	fmt.Println(vobj.Length("go"))
	fmt.Println(vobj.Length([]int{1, 2, 3}))
	// Output:
	// 2
	// 3
}

func ExampleDefaultIfNil() {
	value := 7
	fmt.Println(vobj.DefaultIfNil(&value, 42))
	fmt.Println(vobj.DefaultIfNil[int](nil, 42))
	// Output:
	// 7
	// 42
}

func ExampleDefaultIfNilFunc() {
	fmt.Println(vobj.DefaultIfNilFunc[int](nil, func() int { return 99 }))
	// Output: 99
}

func ExampleDefaultIfNilApply() {
	name := "go"
	fmt.Println(vobj.DefaultIfNilApply(&name, func(s string) int { return len(s) }, -1))
	fmt.Println(vobj.DefaultIfNilApply[string, int](nil, func(s string) int { return len(s) }, -1))
	// Output:
	// 2
	// -1
}

func ExampleSerialize() {
	data, err := vobj.Serialize(exampleObject{Name: "Ada", Scores: []int{3, 5}})
	if err != nil {
		fmt.Println(err)
		return
	}

	decoded, err := vobj.DeserializeTo[exampleObject](data)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(decoded.Name)
	fmt.Println(decoded.Scores)
	// Output:
	// Ada
	// [3 5]
}

func ExampleDeserialize() {
	data, err := vobj.Serialize(exampleObject{Name: "Lin", Scores: []int{8, 13}})
	if err != nil {
		fmt.Println(err)
		return
	}

	var decoded exampleObject
	if err := vobj.Deserialize(data, &decoded); err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(decoded.Name, decoded.Scores[1])
	// Output: Lin 13
}

func ExampleClone() {
	original := exampleObject{Name: "Ada", Scores: []int{1, 2}}
	cloned, err := vobj.Clone(original)
	if err != nil {
		fmt.Println(err)
		return
	}

	cloned.Scores[0] = 99
	fmt.Println(original.Scores)
	fmt.Println(cloned.Scores)
	// Output:
	// [1 2]
	// [99 2]
}

func ExampleCloneIfPossible() {
	type notSerializable struct {
		Fn func()
	}
	src := notSerializable{Fn: func() {}}
	cloned := vobj.CloneIfPossible(src)

	fmt.Println(cloned.Fn != nil)
	// Output: true
}

func ExampleCloneByStream() {
	cloned, err := vobj.CloneByStream(exampleObject{Name: "Ada", Scores: []int{1}})
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(cloned.Name, cloned.Scores)
	// Output: Ada [1]
}

func ExampleApply() {
	name := "knifer-go"
	length := vobj.Apply(&name, func(s string) int { return len(s) })
	missing := vobj.Apply[string, int](nil, func(s string) int { return len(s) })

	fmt.Println(length)
	fmt.Println(missing)
	// Output:
	// 9
	// 0
}

func ExampleAccept() {
	name := "go"
	out := ""
	vobj.Accept(&name, func(s string) { out = "hello " + s })

	fmt.Println(out)
	// Output: hello go
}

func ExampleCompare() {
	low := 1
	high := 2

	fmt.Println(vobj.Compare(&low, &high))
	fmt.Println(vobj.Compare[int](nil, &high))
	fmt.Println(vobj.CompareNull[int](nil, &high, false))
	// Output:
	// -1
	// 1
	// -1
}

func ExampleEmptyCount() {
	fmt.Println(vobj.EmptyCount(nil, "", []int{}, "go"))
	fmt.Println(vobj.HasEmpty("go", []int{1}))
	// Output:
	// 3
	// false
}

func ExampleTypeName() {
	fmt.Println(vobj.TypeName(exampleObject{}))
	// Output: vobj_test.exampleObject
}

func ExampleToString() {
	fmt.Println(vobj.ToString(nil))
	fmt.Println(vobj.ToString(42))
	// Output:
	// null
	// 42
}
