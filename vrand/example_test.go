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

func ExampleInt() {
	fmt.Println(vrand.Int(1))
	// Output: 0
}

func ExampleLong() {
	n := vrand.Long()
	fmt.Println(n >= 0)
	// Output: true
}

func ExampleFloat() {
	n := vrand.Float()
	fmt.Println(n >= 0 && n < 1)
	// Output: true
}

func ExampleBool() {
	_ = vrand.Bool()
	fmt.Println("ok")
	// Output: ok
}

func ExampleString() {
	s := vrand.String(8)
	fmt.Println(len(s))
	// Output: 8
}

func ExampleNumbers() {
	s := vrand.Numbers(6)
	fmt.Println(len(s))
	// Output: 6
}

func ExampleStringUpper() {
	s := vrand.StringUpper(8)
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

func ExampleIntRangeWithOptions() {
	source := mathrand.New(mathrand.NewSource(1))
	fmt.Println(vrand.IntRangeWithOptions(10, 20, vrand.WithRandomSource(source)))
	// Output: 11
}

func ExampleLongWithOptions() {
	source := mathrand.New(mathrand.NewSource(1))
	fmt.Println(vrand.LongWithOptions(vrand.WithRandomSource(source)))
	// Output: 5577006791947779410
}

func ExampleFloatWithOptions() {
	source := mathrand.New(mathrand.NewSource(1))
	fmt.Printf("%.3f\n", vrand.FloatWithOptions(vrand.WithRandomSource(source)))
	// Output: 0.605
}

func ExampleBoolWithOptions() {
	source := mathrand.New(mathrand.NewSource(1))
	fmt.Println(vrand.BoolWithOptions(vrand.WithRandomSource(source)))
	// Output: false
}

func ExampleStringWithOptions() {
	source := mathrand.New(mathrand.NewSource(1))
	fmt.Println(vrand.StringWithOptions(5, vrand.WithRandomSource(source)))
	// Output: fplln
}

func ExampleNumbersWithOptions() {
	source := mathrand.New(mathrand.NewSource(1))
	fmt.Println(vrand.NumbersWithOptions(5, vrand.WithRandomSource(source)))
	// Output: 17791
}

func ExampleStringUpperWithOptions() {
	source := mathrand.New(mathrand.NewSource(1))
	fmt.Println(vrand.StringUpperWithOptions(5, vrand.WithRandomSource(source)))
	// Output: BpLnf
}

func ExampleStringFromWithOptions() {
	source := mathrand.New(mathrand.NewSource(1))
	fmt.Println(vrand.StringFromWithOptions("abc", 5, vrand.WithRandomSource(source)))
	// Output: caccb
}

func ExampleEleWithOptions() {
	source := mathrand.New(mathrand.NewSource(1))
	fmt.Println(vrand.EleWithOptions([]string{"a", "b", "c"}, vrand.WithRandomSource(source)))
	// Output: c
}

func ExampleWithRandomSource() {
	source := mathrand.New(mathrand.NewSource(1))
	fmt.Println(vrand.IntWithOptions(10, vrand.WithRandomSource(source)))
	// Output: 1
}

func ExampleWithRandomReader() {
	b, err := vrand.SecureBytesWithOptions(4, vrand.WithRandomReader(strings.NewReader("salt")))
	if err != nil {
		panic(err)
	}
	fmt.Println(string(b))
	// Output: salt
}

func ExampleWithStrictCryptoRandom() {
	_, err := vrand.BytesWithOptions(4, vrand.WithRandomReader(strings.NewReader("x")), vrand.WithStrictCryptoRandom())
	fmt.Println(err != nil)
	// Output: true
}

func ExampleSetSeed() {
	vrand.SetSeed(42)
	defer vrand.ResetDefaultRandomSource()

	first := vrand.Int(100)
	vrand.SetSeed(42)
	fmt.Println(first == vrand.Int(100))
	// Output: true
}

func ExampleConfigureDefaultRandomSourceProvider() {
	vrand.ConfigureDefaultRandomSourceProvider(func() *mathrand.Rand {
		return mathrand.New(mathrand.NewSource(9))
	})
	defer vrand.ResetDefaultRandomSource()

	first := vrand.Int(1000)
	vrand.ConfigureDefaultRandomSourceProvider(func() *mathrand.Rand {
		return mathrand.New(mathrand.NewSource(9))
	})
	fmt.Println(first == vrand.Int(1000))
	// Output: true
}

func ExampleResetDefaultRandomSource() {
	vrand.SetSeed(42)
	vrand.ResetDefaultRandomSource()
	fmt.Println("reset")
	// Output: reset
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

func ExampleWithWeightedRandSource() {
	source := mathrand.New(mathrand.NewSource(1))
	item, err := vrand.WeightedPick([]string{"a", "b"}, []float64{0, 1}, vrand.WithWeightedRandSource(source))
	if err != nil {
		panic(err)
	}
	fmt.Println(item)
	// Output: b
}

func ExampleWithWeightedPrecision() {
	_, err := vrand.WeightedPick([]string{"tiny"}, []float64{1e-15}, vrand.WithWeightedPrecision(1e-12))
	fmt.Println(err != nil)
	// Output: true
}
