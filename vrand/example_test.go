package vrand_test

import (
	"encoding/hex"
	"fmt"
	mathrand "math/rand"
	"strings"

	"github.com/imajinyun/knifer-go/vrand"
)

func ExampleSecureBytes() {
	b, err := vrand.SecureBytes(16)
	if err != nil {
		panic(err)
	}
	fmt.Println(len(b))
	// Output: 16
}

func ExampleSecureBytesWithOptions() {
	b, err := vrand.SecureBytesWithOptions(8, vrand.WithRandomReader(strings.NewReader("nonce123")))
	if err != nil {
		panic(err)
	}
	fmt.Println(hex.EncodeToString(b))
	// Output: 6e6f6e6365313233
}

func ExampleIntRange() {
	n := vrand.IntRange(10, 20)
	fmt.Println(n >= 10 && n < 20)
	// Output: true
}

func ExampleString() {
	s := vrand.String(8)
	fmt.Println(len(s))
	// Output: 8
}

func ExampleStringFrom() {
	fmt.Println(vrand.StringFrom("A", 4))
	// Output: AAAA
}

func ExampleEle() {
	fmt.Println(vrand.Ele([]string{"only"}))
	// Output: only
}

func ExampleBytesWithOptions() {
	b, err := vrand.BytesWithOptions(4, vrand.WithRandomReader(strings.NewReader("data")), vrand.WithStrictCryptoRandom())
	if err != nil {
		panic(err)
	}
	fmt.Println(string(b))
	// Output: data
}

func ExampleIntWithOptions() {
	source := mathrand.New(mathrand.NewSource(1))
	fmt.Println(vrand.IntWithOptions(100, vrand.WithRandomSource(source)))
	// Output: 81
}

func ExampleEleWithOptions() {
	source := mathrand.New(mathrand.NewSource(1))
	fmt.Println(vrand.EleWithOptions([]string{"a", "b", "c"}, vrand.WithRandomSource(source)))
	// Output: c
}

func ExampleWeightedPick() {
	item, err := vrand.WeightedPick(
		[]string{"cold", "hot"},
		[]float64{0, 10},
		vrand.WithWeightedRandSource(mathrand.New(mathrand.NewSource(1))),
	)
	if err != nil {
		panic(err)
	}
	fmt.Println(item)
	// Output: hot
}

func ExampleWeightedPickN() {
	items, err := vrand.WeightedPickN(
		[]string{"a", "b", "c"},
		[]float64{0, 1, 0},
		3,
		vrand.WithWeightedRandSource(mathrand.New(mathrand.NewSource(1))),
	)
	if err != nil {
		panic(err)
	}
	fmt.Println(items)
	// Output: [b b b]
}

func ExampleWeightedPickUniqueN() {
	items, err := vrand.WeightedPickUniqueN(
		[]string{"red", "green", "blue"},
		[]float64{1, 1, 1},
		2,
		vrand.WithWeightedRandSource(mathrand.New(mathrand.NewSource(2))),
	)
	if err != nil {
		panic(err)
	}
	fmt.Println(len(items), items[0] != items[1])
	// Output: 2 true
}
