package num

import (
	cryptorand "crypto/rand"
	"errors"
	"math/big"
	"reflect"
	"testing"
)

type errReader struct{}

func (errReader) Read([]byte) (int, error) { return 0, errors.New("forced random failure") }

type sequenceReader struct {
	next byte
}

func (r *sequenceReader) Read(p []byte) (int, error) {
	for i := range p {
		p[i] = r.next
		r.next++
	}
	return len(p), nil
}

func TestRangeFunc(t *testing.T) {
	r := Range(0, 5, 1)
	if len(r) != 5 || r[0] != 0 || r[4] != 4 {
		t.Fatalf("Range asc: %v", r)
	}
	r = Range(5, 0, -1)
	if len(r) != 5 || r[0] != 5 || r[4] != 1 {
		t.Fatalf("Range desc: %v", r)
	}
	r = Range(0, 10, 3)
	if len(r) != 4 || r[3] != 9 {
		t.Fatalf("Range step: %v", r)
	}
}

func TestRangeFactorialAndCombinatorics(t *testing.T) {
	randoms := GenerateRandomNumber(1, 10, 5)
	if len(randoms) != 5 {
		t.Fatalf("GenerateRandomNumber length: %v", randoms)
	}
	seen := map[int]struct{}{}
	for _, v := range randoms {
		if v < 1 || v >= 10 {
			t.Fatalf("GenerateRandomNumber value out of range: %v", randoms)
		}
		if _, ok := seen[v]; ok {
			t.Fatalf("GenerateRandomNumber duplicated value: %v", randoms)
		}
		seen[v] = struct{}{}
	}
	bySet := GenerateBySet(1, 10, 5)
	if len(bySet) != 5 {
		t.Fatalf("GenerateBySet length: %v", bySet)
	}
	if got := RangeClosed(1, 5, 2); !reflect.DeepEqual(got, []int{1, 3, 5}) {
		t.Fatalf("RangeClosed asc: %v", got)
	}
	if got := RangeClosed(5, 1, 2); !reflect.DeepEqual(got, []int{5, 3, 1}) {
		t.Fatalf("RangeClosed desc: %v", got)
	}
	if got := AppendRange(1, 3, 1, []int{0}); !reflect.DeepEqual(got, []int{0, 1, 2, 3}) {
		t.Fatalf("AppendRange: %v", got)
	}
	if got, err := Factorial(5); err != nil || got != 120 {
		t.Fatalf("Factorial: %d %v", got, err)
	}
	if got, err := FactorialRange(5, 2); err != nil || got != 60 {
		t.Fatalf("FactorialRange: %d %v", got, err)
	}
	if FactorialBig(big.NewInt(20)).String() != "2432902008176640000" {
		t.Fatal("FactorialBig failed")
	}
	if Sqrt(81) != 9 || ProcessMultiple(7, 5) != 21 || Divisor(24, 18) != 6 || Multiple(4, 6) != 12 {
		t.Fatal("math helpers failed")
	}
}

func TestRandomRangeAndFactorialEdges(t *testing.T) {
	if got := GenerateRandomNumber(10, 1, 3); len(got) != 0 {
		t.Fatalf("GenerateRandomNumber reversed bounds should be empty because default seed is empty: %v", got)
	}
	if got := GenerateRandomNumber(1, 3, 5); len(got) != 0 {
		t.Fatalf("GenerateRandomNumber oversize should be empty: %v", got)
	}
	if got := GenerateRandomNumberWithSeed(1, 10, 2, []int{7}); len(got) != 0 {
		t.Fatalf("GenerateRandomNumberWithSeed short seed should be empty: %v", got)
	}
	seed := []int{1, 2, 3, 4}
	got := GenerateRandomNumberWithSeed(1, 5, 2, seed)
	if len(got) != 2 || !reflect.DeepEqual(seed, []int{1, 2, 3, 4}) {
		t.Fatalf("GenerateRandomNumberWithSeed should not mutate seed: got=%v seed=%v", got, seed)
	}
	if got := GenerateBySet(5, 1, 0); len(got) != 0 {
		t.Fatalf("GenerateBySet zero size should be empty: %v", got)
	}
	if got := GenerateBySet(1, 2, 3); len(got) != 0 {
		t.Fatalf("GenerateBySet oversize should be empty: %v", got)
	}
	if got := Range(1, 5, 0); !reflect.DeepEqual(got, []int{1, 2, 3, 4}) {
		t.Fatalf("Range zero positive step: %v", got)
	}
	if got := Range(5, 1, 0); !reflect.DeepEqual(got, []int{5, 4, 3, 2}) {
		t.Fatalf("Range zero negative step: %v", got)
	}
	if got := RangeClosed(3, 3, 0); !reflect.DeepEqual(got, []int{3}) {
		t.Fatalf("RangeClosed equal endpoints: %v", got)
	}
	if got := RangeClosed(1, 5, -2); !reflect.DeepEqual(got, []int{1, 3, 5}) {
		t.Fatalf("RangeClosed should normalize step sign: %v", got)
	}
	if got, err := Factorial(0); err != nil || got != 1 {
		t.Fatalf("Factorial(0) = %d, %v", got, err)
	}
	if got, err := Factorial(21); err == nil || got != 0 {
		t.Fatalf("Factorial overflow should fail: %d, %v", got, err)
	}
	if got, err := FactorialRange(0, 10); err != nil || got != 1 {
		t.Fatalf("FactorialRange start zero = %d, %v", got, err)
	}
	if got, err := FactorialRange(2, 5); err != nil || got != 0 {
		t.Fatalf("FactorialRange start smaller should be zero: %d, %v", got, err)
	}
	if got, err := FactorialRange(21, 0); err == nil || got != 0 {
		t.Fatalf("FactorialRange overflow should fail: %d, %v", got, err)
	}
	if FactorialBig(nil).String() != "1" || FactorialBig(big.NewInt(-1)).String() != "1" {
		t.Fatal("FactorialBig nil/negative should be one")
	}
	if FactorialBigRange(big.NewInt(5), big.NewInt(2)).String() != "60" {
		t.Fatal("FactorialBigRange normal case failed")
	}
	if FactorialBigRange(nil, big.NewInt(0)).String() != "1" || FactorialBigRange(big.NewInt(2), big.NewInt(5)).String() != "1" {
		t.Fatal("FactorialBigRange guard cases failed")
	}
}

func TestRandomGenerationWithOptions(t *testing.T) {
	seed := []int{10, 20, 30, 40}
	got := GenRandomNumberWithSeedWithOptions(0, 4, 3, seed, WithRandomReader(&sequenceReader{}))
	if !reflect.DeepEqual(got, []int{10, 20, 40}) {
		t.Fatalf("GenRandomNumberWithSeedWithOptions deterministic = %v", got)
	}
	if !reflect.DeepEqual(seed, []int{10, 20, 30, 40}) {
		t.Fatalf("GenRandomNumberWithSeedWithOptions should not mutate seed: %v", seed)
	}

	got = GenRandomNumberWithOptions(0, 5, 3, WithRandomReader(&sequenceReader{}))
	if !reflect.DeepEqual(got, []int{0, 1, 2}) {
		t.Fatalf("GenRandomNumberWithOptions deterministic = %v", got)
	}

	got = GenBySetWithOptions(0, 5, 3, WithRandomReader(&sequenceReader{}))
	if len(got) != 3 {
		t.Fatalf("GenBySetWithOptions length = %v", got)
	}
	seen := map[int]bool{}
	for _, v := range got {
		seen[v] = true
	}
	for _, want := range []int{0, 1, 2} {
		if !seen[want] {
			t.Fatalf("GenBySetWithOptions missing %d in %v", want, got)
		}
	}

	if got := GenRandomNumberWithOptions(0, 5, 2, WithRandomReader(errReader{})); !reflect.DeepEqual(got, []int{0, 4}) {
		t.Fatalf("random failure should preserve fallback index behavior: %v", got)
	}
}

func TestCombinatoricsGcdLcmAndBinaryEdges(t *testing.T) {
	if ProcessMultiple(3, 5) != 0 || ProcessMultiple(5, -1) != 0 || ProcessMultiple(5, 5) != 1 || ProcessMultiple(5, 2) != 10 {
		t.Fatal("ProcessMultiple edge cases failed")
	}
	if Divisor(-24, 18) != 6 || Divisor(7, 0) != 7 || Multiple(-4, 6) != 12 || Multiple(0, 6) != 0 {
		t.Fatal("gcd/lcm edge cases failed")
	}
	binaryCases := map[any]string{
		int(5):            "101",
		int8(-2):          "11111110",
		int16(-2):         "1111111111111110",
		int32(-2):         "-10",
		int64(9):          "1001",
		uint8(7):          "111",
		float64(1):        "0011111111110000000000000000000000000000000000000000000000000000",
		big.NewInt(10):    "1010",
		(*big.Int)(nil):   "",
		complex64(1 + 2i): "",
	}
	for input, want := range binaryCases {
		if got := GetBinaryStr(input); got != want {
			t.Fatalf("GetBinaryStr(%T %[1]v) = %q, want %q", input, got, want)
		}
	}
	if _, err := BinaryToInt("102"); err == nil {
		t.Fatal("BinaryToInt should reject invalid binary strings")
	}
	if _, err := BinaryToLong("2"); err == nil {
		t.Fatal("BinaryToLong should reject invalid binary strings")
	}
}

func TestRemainingInternalHelperEdges(t *testing.T) {
	if IsLong("   ") {
		t.Fatal("IsLong blank should be false")
	}
	if got := RangeClosed(5, 1, 0); !reflect.DeepEqual(got, []int{5, 4, 3, 2, 1}) {
		t.Fatalf("RangeClosed descending zero step: %v", got)
	}
	if Max(1, 3, 2) != 3 {
		t.Fatal("Max should update when a later value is larger")
	}
	if ParseInt("123") != 123 || ParseLong("123") != 123 {
		t.Fatal("ParseInt/ParseLong direct integer branch failed")
	}
	if stripTrailingZeros("12") != "12" || stripTrailingZeros("1.2300") != "1.23" || stripTrailingZeros("1.2300e2") != "1.2300e2" {
		t.Fatal("stripTrailingZeros cases failed")
	}
	if addThousands("123") != "123" || addThousands("1234") != "1,234" {
		t.Fatal("addThousands integer cases failed")
	}
	if got, err := Calculate("1+"); err == nil || got != 0 {
		t.Fatalf("Calculate trailing plus should fail: %v %v", got, err)
	}
	if got, err := Calculate("1*"); err == nil || got != 0 {
		t.Fatalf("Calculate trailing multiply should fail: %v %v", got, err)
	}
	oldReader := cryptorand.Reader
	cryptorand.Reader = errReader{}
	t.Cleanup(func() { cryptorand.Reader = oldReader })
	if got := secureIntn(10); got != 0 {
		t.Fatalf("secureIntn should return 0 when crypto random fails: %d", got)
	}
}
