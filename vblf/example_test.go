package vblf_test

import (
	"fmt"
	"io"
	"strings"

	"github.com/imajinyun/knifer-go/vblf"
)

func ExampleNewBitSetBloomFilter() {
	bf := vblf.NewBitSetBloomFilter(1000, 100, 3)
	bf.Add("hello")
	bf.Add("world")

	fmt.Println(bf.Contains("hello"))
	fmt.Println(bf.Contains("go"))
	// Output:
	// true
	// false
}

func ExampleNewIntMap() {
	m := vblf.NewIntMap(1000)
	m.Add(42)

	fmt.Println(m.Contains(42))
	fmt.Println(m.Contains(100))
	// Output:
	// true
	// false
}

func ExampleNewBitMapBloomFilter() {
	bf := vblf.NewBitMapBloomFilter(5)
	bf.Add("hello")

	fmt.Println(bf.Contains("hello"))
	// Output: true
}

func ExampleNewFuncFilter() {
	filter := vblf.NewFuncFilter(1000, func(s string) int64 { return int64(len(s)) })
	filter.Add("hello")

	fmt.Println(filter.Contains("hello"))
	// Output: true
}

func ExampleNewFuncFilterWithMachineNumE() {
	filter, err := vblf.NewFuncFilterWithMachineNumE(1000, vblf.BloomMachine64, func(s string) int64 {
		return int64(vblf.JavaDefaultHash(s))
	})
	if err != nil {
		panic(err)
	}

	fmt.Println(filter.Add("hello"))
	fmt.Println(filter.Add("hello"))
	fmt.Println(filter.Contains("hello"))
	// Output:
	// true
	// false
	// true
}

func ExampleNewDefaultFilter() {
	filter := vblf.NewDefaultFilter(1000)
	filter.Add("knifer-go")

	fmt.Println(filter.Contains("knifer-go"))
	// Output: true
}

func ExampleNewLongMap() {
	m := vblf.NewLongMap(100)
	m.Add(7)

	fmt.Println(m.Contains(7))
	fmt.Println(m.Contains(99))
	// Output:
	// true
	// false
}

func ExampleNewBitMapBloomFilterWithOptions() {
	bf := vblf.NewBitMapBloomFilterWithOptions(vblf.WithBitMapSize(5))
	bf.Add("go")

	fmt.Println(bf.Contains("go"))
	// Output: true
}

func ExampleNewBitMapBloomFilterWithFilters() {
	filter := vblf.NewDefaultBloomFilter(1000)
	bf := vblf.NewBitMapBloomFilterWithFilters(1, filter)
	bf.Add("go")

	fmt.Println(bf.Contains("go"))
	fmt.Println(bf.Contains("rust"))
	// Output:
	// true
	// false
}

func ExampleNewBitMapBloomFilterE() {
	bf, err := vblf.NewBitMapBloomFilterE(5)
	fmt.Println(bf != nil)
	fmt.Println(err)
	// Output:
	// true
	// <nil>
}

func ExampleNewBitSetBloomFilterWithOptions() {
	bf := vblf.NewBitSetBloomFilterWithOptions(
		vblf.WithBitSetCapacity(100),
		vblf.WithExpectedElements(10),
		vblf.WithHashFunctionNumber(3),
	)
	bf.Add("go")

	fmt.Println(bf.Contains("go"))
	// Output: true
}

func ExampleNewBitSetBloomFilterE() {
	bf, err := vblf.NewBitSetBloomFilterE(100, 10, 3)
	fmt.Println(bf != nil)
	fmt.Println(err)
	// Output:
	// true
	// <nil>
}

func ExampleBitSetBloomFilter_FalsePositiveProbability() {
	bf := vblf.NewBitSetBloomFilter(100, 10, 3)

	fmt.Printf("%.4f\n", bf.FalsePositiveProbability())
	// Output: 0.0009
}

func ExampleInitFromReader() {
	bf := vblf.NewBitSetBloomFilter(100, 10, 3)
	err := vblf.InitFromReader(bf, strings.NewReader("alpha\nbeta\n"))

	fmt.Println(bf.Contains("alpha"))
	fmt.Println(bf.Contains("beta"))
	fmt.Println(err)
	// Output:
	// true
	// true
	// <nil>
}

func ExampleInitFromFileWithOptions() {
	bf := vblf.NewBitSetBloomFilter(100, 10, 3)
	err := vblf.InitFromFileWithOptions(
		bf,
		"ignored.txt",
		vblf.WithOpenFile(func(string) (io.ReadCloser, error) {
			return io.NopCloser(strings.NewReader("gamma\n")), nil
		}),
	)

	fmt.Println(bf.Contains("gamma"))
	fmt.Println(err)
	// Output:
	// true
	// <nil>
}

func ExampleNewFuncFilterE() {
	filter, err := vblf.NewFuncFilterE(1000, func(s string) int64 {
		return int64(len(s))
	})
	filter.Add("go")

	fmt.Println(filter.Contains("go"))
	fmt.Println(err)
	// Output:
	// true
	// <nil>
}

func ExampleNewFuncFilterWithOptions() {
	filter := vblf.NewFuncFilterWithOptions(
		vblf.WithMaxValue(1000),
		vblf.WithMachineNum(vblf.BloomMachine64),
		vblf.WithHashFunc(func(s string) int64 { return int64(len(s)) }),
	)
	filter.Add("go")

	fmt.Println(filter.Contains("go"))
	// Output: true
}

func ExampleBloomFNVHash() {
	goHash := vblf.BloomFNVHash("go")
	rustHash := vblf.BloomFNVHash("rust")

	fmt.Println(goHash != 0)
	fmt.Println(goHash == rustHash)
	// Output:
	// true
	// false
}
