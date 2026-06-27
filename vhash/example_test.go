package vhash_test

import (
	"fmt"

	"github.com/imajinyun/knifer-go/vhash"
)

func ExampleAdditiveHash() {
	fmt.Println(vhash.AdditiveHash("abc", 31))
	// Output: 18
}

func ExampleJavaDefaultHash() {
	// Equivalent to Java String.hashCode.
	fmt.Println(vhash.JavaDefaultHash("a"))
	// Output: 97
}

func ExampleBkdrHash() {
	fmt.Println(vhash.BkdrHash("a"))
	// Output: 97
}

func ExampleDjbHash() {
	fmt.Println(vhash.DjbHash("a"))
	// Output: 177670
}

func ExampleHfHash() {
	fmt.Println(vhash.HfHash("abc"))
	// Output: 888
}

func ExampleNewConsistentHash() {
	ring := vhash.NewConsistentHash(vhash.WithVirtualNodes(8))
	ring.Add("cache-a")
	ring.Add("cache-b")
	ring.Add("cache-c")

	node, err := ring.Get("user:42")
	if err != nil {
		panic(err)
	}
	replicas, err := ring.GetN("user:42", 2)
	if err != nil {
		panic(err)
	}
	fmt.Println(node != "")
	fmt.Println(len(replicas))
	// Output:
	// true
	// 2
}

func ExampleWithVirtualNodes() {
	ring := vhash.NewConsistentHash(vhash.WithVirtualNodes(4))
	ring.Add("cache-a")
	ring.Add("cache-b")

	node, err := ring.Get("asset:logo")
	if err != nil {
		panic(err)
	}
	fmt.Println(node != "")
	// Output: true
}

func ExampleWithReplicaCount() {
	ring := vhash.NewConsistentHash(vhash.WithReplicaCount(4))
	ring.Add("cache-a")
	ring.Add("cache-b")

	nodes, err := ring.GetN("asset:logo", 2)
	if err != nil {
		panic(err)
	}
	fmt.Println(len(nodes))
	// Output: 2
}

func ExampleWithHashFunc() {
	hashFunc := func(data []byte) uint64 {
		var sum uint64
		for _, b := range data {
			sum = sum*131 + uint64(b)
		}
		return sum
	}
	ring := vhash.NewConsistentHash(
		vhash.WithVirtualNodes(2),
		vhash.WithHashFunc(hashFunc),
	)
	ring.Add("cache-a")
	ring.Add("cache-b")

	node, err := ring.Get("asset:logo")
	if err != nil {
		panic(err)
	}
	fmt.Println(node != "")
	// Output: true
}
