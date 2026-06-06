// Package num provides numeric helpers.
package num

import (
	cryptorand "crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
	"math/big"
	"strconv"
	"strings"
	"unicode"
)

const defaultDivScale = 10

var factorials = [...]uint64{
	1, 1, 2, 6, 24, 120, 720, 5040, 40320, 362880, 3628800,
	39916800, 479001600, 6227020800, 87178291200, 1307674368000,
	20922789888000, 355687428096000, 6402373705728000,
	121645100408832000, 2432902008176640000,
}

type randomNumberConfig struct {
	randomReader io.Reader
}

// RandomNumberOption customizes random-number generation per call.
type RandomNumberOption func(*randomNumberConfig)

// WithRandomReader sets the random source used by Gen*WithOptions helpers.
func WithRandomReader(reader io.Reader) RandomNumberOption {
	return func(c *randomNumberConfig) {
		if reader != nil {
			c.randomReader = reader
		}
	}
}

func applyRandomNumberOptions(opts []RandomNumberOption) randomNumberConfig {
	cfg := randomNumberConfig{randomReader: cryptorand.Reader}
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}
	if cfg.randomReader == nil {
		cfg.randomReader = cryptorand.Reader
	}
	return cfg
}

// This file provides numeric helper functions for arithmetic, parsing, formatting,
// comparison, random number generation, and low-level conversions.

// NumberAdd performs high-precision addition to reduce floating-point rounding surprises.
func NumberAdd(a, b float64) float64 { return Add(a, b) }

// NumberSub performs high-precision subtraction.
func NumberSub(a, b float64) float64 { return Sub(a, b) }

// NumberMul performs high-precision multiplication.
func NumberMul(a, b float64) float64 { return Mul(a, b) }

// NumberDiv divides a by b and rounds to scale decimal places with HALF_UP semantics.
// A negative scale disables rounding; division by zero returns 0 for compatibility.
func NumberDiv(a, b float64, scale int) float64 {
	if b == 0 {
		return 0
	}
	if scale < 0 {
		return a / b
	}
	return Round(a/b, scale)
}

// Add returns the sum of all values using decimal strings as the intermediate form.
func Add(values ...float64) float64 {
	r := new(big.Rat)
	for _, v := range values {
		r.Add(r, ratFromFloat(v))
	}
	return ratToFloat(r)
}

// AddStr returns the exact decimal sum of all numeric strings.
func AddStr(values ...string) *big.Rat { return foldRat((*big.Rat).Add, values...) }

// Sub subtracts all following values from the first value.
func Sub(values ...float64) float64 {
	if len(values) == 0 {
		return 0
	}
	r := ratFromFloat(values[0])
	for _, v := range values[1:] {
		r.Sub(r, ratFromFloat(v))
	}
	return ratToFloat(r)
}

// SubStr subtracts all following numeric strings from the first value.
func SubStr(values ...string) *big.Rat {
	if len(values) == 0 {
		return new(big.Rat)
	}
	r := ToBigDecimal(values[0])
	for _, v := range values[1:] {
		if strings.TrimSpace(v) != "" {
			r.Sub(r, ToBigDecimal(v))
		}
	}
	return r
}

// Mul returns the product of all values. Empty input returns 0.
func Mul(values ...float64) float64 {
	if len(values) == 0 {
		return 0
	}
	r := big.NewRat(1, 1)
	for _, v := range values {
		r.Mul(r, ratFromFloat(v))
	}
	return ratToFloat(r)
}

// MulStr returns the exact decimal product of all numeric strings.
func MulStr(values ...string) *big.Rat {
	if len(values) == 0 {
		return new(big.Rat)
	}
	r := big.NewRat(1, 1)
	for _, v := range values {
		if strings.TrimSpace(v) == "" {
			return new(big.Rat)
		}
		r.Mul(r, ToBigDecimal(v))
	}
	return r
}

// Div divides a by b using the default scale of 10 decimal places.
func Div(a, b float64) float64 { return NumberDiv(a, b, defaultDivScale) }

// DivWithMode divides a by b and rounds using the requested mode.
func DivWithMode(a, b float64, scale int, mode RoundingMode) float64 {
	if b == 0 {
		return 0
	}
	return RoundMode(a/b, scale, mode)
}

// CeilDiv returns ceil(v1 / v2).
func CeilDiv(v1, v2 int) int { return int(math.Ceil(float64(v1) / float64(v2))) }

// RoundingMode describes decimal rounding behavior.
type RoundingMode int

const (
	RoundHalfUp RoundingMode = iota
	RoundHalfEven
	RoundDown
)

// Round rounds v to scale decimal places with HALF_UP semantics.
func Round(v float64, scale int) float64 { return RoundMode(v, scale, RoundHalfUp) }

// RoundMode rounds v to scale decimal places using mode.
func RoundMode(v float64, scale int, mode RoundingMode) float64 {
	if scale < 0 {
		scale = 0
	}
	pow := math.Pow(10, float64(scale))
	scaled := v * pow
	var rounded float64
	switch mode {
	case RoundDown:
		rounded = math.Trunc(scaled)
	case RoundHalfEven:
		rounded = math.RoundToEven(scaled)
	default:
		if scaled >= 0 {
			rounded = math.Floor(scaled + 0.5)
		} else {
			rounded = math.Ceil(scaled - 0.5)
		}
	}
	return rounded / pow
}

// RoundStr returns Round formatted with fixed scale digits.
func RoundStr(v float64, scale int) string {
	return strconv.FormatFloat(Round(v, scale), 'f', maxInt(scale, 0), 64)
}

// RoundHalfEvenFloat rounds with banker rounding.
func RoundHalfEvenFloat(v float64, scale int) float64 { return RoundMode(v, scale, RoundHalfEven) }

// RoundDownFloat truncates extra decimal places.
func RoundDownFloat(v float64, scale int) float64 { return RoundMode(v, scale, RoundDown) }

// DecimalFormat formats v with common decimal patterns such as "0", "0.00", ",##0.00" and percent patterns.
func DecimalFormat(format string, v float64) string {
	if format == "" {
		return strconv.FormatFloat(v, 'f', 0, 64)
	}
	percent := strings.Contains(format, "%")
	if percent {
		v *= 100
	}
	decimals := 0
	if dot := strings.Index(format, "."); dot >= 0 {
		for _, r := range format[dot+1:] {
			if r == '0' || r == '#' {
				decimals++
			}
		}
	}
	out := strconv.FormatFloat(Round(v, decimals), 'f', decimals, 64)
	if strings.Contains(format, ",") {
		out = addThousands(out)
	}
	if percent {
		out += "%"
	}
	return out
}

// DecimalFormatMoney formats money with comma grouping and two decimal places.
func DecimalFormatMoney(v float64) string { return DecimalFormat(",##0.00", v) }

// FormatPercent formats number as a percentage with scale fraction digits.
func FormatPercent(number float64, scale int) string {
	return DecimalFormat("0."+strings.Repeat("0", maxInt(scale, 0))+"%", number)
}

// IsNumber reports whether s is a valid number, including hex and scientific notation.
func IsNumber(s string) bool {
	s = strings.TrimSpace(s)
	if s == "" {
		return false
	}
	if strings.HasPrefix(s, "+") || strings.HasPrefix(s, "-") {
		s = s[1:]
	}
	if strings.HasPrefix(strings.ToLower(s), "0x") {
		_, ok := new(big.Int).SetString(s[2:], 16)
		return ok
	}
	last := s[len(s)-1]
	if strings.ContainsRune("dDfFlL", rune(last)) {
		s = s[:len(s)-1]
	}
	_, err := strconv.ParseFloat(s, 64)
	return err == nil
}

// IsInteger reports whether s is a valid base-10 int.
func IsInteger(s string) bool {
	if strings.TrimSpace(s) == "" {
		return false
	}
	_, err := strconv.ParseInt(s, 10, 0)
	return err == nil
}

// IsLong reports whether s is a valid base-10 int64.
func IsLong(s string) bool {
	if strings.TrimSpace(s) == "" {
		return false
	}
	_, err := strconv.ParseInt(s, 10, 64)
	return err == nil
}

// IsDouble reports whether s is a valid floating-point value containing a decimal point.
func IsDouble(s string) bool {
	if strings.TrimSpace(s) == "" || !strings.Contains(s, ".") {
		return false
	}
	_, err := strconv.ParseFloat(s, 64)
	return err == nil
}

// IsDigits reports whether s contains only unsigned ASCII digits.
func IsDigits(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

// IsPrimes reports whether n is a prime number.
func IsPrimes(n int) bool {
	if n <= 1 {
		return false
	}
	if n <= 3 {
		return true
	}
	if n%2 == 0 || n%3 == 0 {
		return false
	}
	for i := 5; i <= n/i; i += 6 {
		if n%i == 0 || n%(i+2) == 0 {
			return false
		}
	}
	return true
}

// GenerateRandomNumber returns size unique random integers in [begin, end).
func GenerateRandomNumber(begin, end, size int) []int {
	return GenRandomNumberWithOptions(begin, end, size)
}

// GenRandomNumberWithOptions returns size unique random integers in [begin, end) with per-call options.
func GenRandomNumberWithOptions(begin, end, size int, opts ...RandomNumberOption) []int {
	return GenRandomNumberWithSeedWithOptions(begin, end, size, Range(begin, end, 1), opts...)
}

// GenerateRandomNumberWithSeed picks size unique values from seed.
func GenerateRandomNumberWithSeed(begin, end, size int, seed []int) []int {
	return GenRandomNumberWithSeedWithOptions(begin, end, size, seed)
}

// GenRandomNumberWithSeedWithOptions picks size unique values from seed with per-call options.
func GenRandomNumberWithSeedWithOptions(begin, end, size int, seed []int, opts ...RandomNumberOption) []int {
	if begin > end {
		begin, end = end, begin
	}
	if size < 0 || end-begin < size || len(seed) < size {
		return []int{}
	}
	cfg := applyRandomNumberOptions(opts)
	pool := append([]int(nil), seed...)
	out := make([]int, size)
	for i := 0; i < size; i++ {
		j := secureIntnWithReader(cfg.randomReader, len(pool)-i)
		out[i] = pool[j]
		pool[j] = pool[len(pool)-1-i]
	}
	return out
}

// GenerateBySet returns size unique random integers in [begin, end).
func GenerateBySet(begin, end, size int) []int {
	return GenBySetWithOptions(begin, end, size)
}

// GenBySetWithOptions returns size unique random integers in [begin, end) with per-call options.
func GenBySetWithOptions(begin, end, size int, opts ...RandomNumberOption) []int {
	if begin > end {
		begin, end = end, begin
	}
	if size < 0 || end-begin < size {
		return []int{}
	}
	cfg := applyRandomNumberOptions(opts)
	set := make(map[int]struct{}, size)
	for len(set) < size {
		set[begin+secureIntnWithReader(cfg.randomReader, end-begin)] = struct{}{}
	}
	out := make([]int, 0, size)
	for v := range set {
		out = append(out, v)
	}
	return out
}

// Range returns a half-open integer sequence [start, end) using step.
// A zero step is normalized to 1 or -1 based on the direction.
func Range(start, end, step int) []int {
	if step == 0 {
		if end >= start {
			step = 1
		} else {
			step = -1
		}
	}
	out := make([]int, 0)
	if step > 0 {
		for i := start; i < end; i += step {
			out = append(out, i)
		}
	} else {
		for i := start; i > end; i += step {
			out = append(out, i)
		}
	}
	return out
}

// RangeClosed returns an inclusive integer sequence.
func RangeClosed(start, stop, step int) []int {
	if start == stop {
		return []int{start}
	}
	if step == 0 {
		step = 1
	}
	if start < stop {
		step = absInt(step)
	} else {
		step = -absInt(step)
	}
	out := make([]int, 0, absInt(stop-start)/absInt(step)+1)
	for i := start; ; i += step {
		out = append(out, i)
		if (step > 0 && i+step > stop) || (step < 0 && i+step < stop) {
			break
		}
	}
	return out
}

// AppendRange appends an inclusive range to values and returns the result.
func AppendRange(start, stop, step int, values []int) []int {
	return append(values, RangeClosed(start, stop, step)...)
}

// Factorial returns n! for 0 <= n <= 20.
func Factorial(n uint64) (uint64, error) {
	if n > 20 {
		return 0, fmt.Errorf("factorial overflows uint64: %d", n)
	}
	return factorials[n], nil
}

// FactorialRange returns start * (start-1) * ... * (end+1).
func FactorialRange(start, end uint64) (uint64, error) {
	if start == 0 || start == end {
		return 1, nil
	}
	if start < end {
		return 0, nil
	}
	result := uint64(1)
	for i := start; i > end; i-- {
		if result > math.MaxUint64/i {
			return 0, fmt.Errorf("factorial range overflows uint64")
		}
		result *= i
	}
	return result, nil
}

// FactorialBig returns n! as a big integer.
func FactorialBig(n *big.Int) *big.Int {
	if n == nil || n.Sign() <= 0 {
		return big.NewInt(1)
	}
	return FactorialBigRange(n, big.NewInt(0))
}

// FactorialBigRange returns start * (start-1) * ... * (end+1) as a big integer.
func FactorialBigRange(start, end *big.Int) *big.Int {
	if start == nil || end == nil || start.Sign() < 0 || end.Sign() < 0 || start.Cmp(end) <= 0 {
		return big.NewInt(1)
	}
	result := big.NewInt(1)
	for i := new(big.Int).Set(start); i.Cmp(end) > 0; i.Sub(i, big.NewInt(1)) {
		result.Mul(result, i)
	}
	return result
}

// Sqrt returns the integer square root of x.
func Sqrt(x uint64) uint64 { return uint64(math.Sqrt(float64(x))) }

// ProcessMultiple returns the combination count C(selectNum, minNum).
func ProcessMultiple(selectNum, minNum int) int {
	if minNum < 0 || selectNum < minNum {
		return 0
	}
	return mathSubNode(selectNum, minNum) / mathNode(selectNum-minNum)
}

// Divisor returns the greatest common divisor.
func Divisor(m, n int) int {
	m, n = absInt(m), absInt(n)
	if n == 0 {
		return m
	}
	for m%n != 0 {
		m, n = n, m%n
	}
	return n
}

// Multiple returns the least common multiple.
func Multiple(m, n int) int {
	if m == 0 || n == 0 {
		return 0
	}
	return absInt(m / Divisor(m, n) * n)
}

// GetBinaryStr returns the binary representation of common numeric values.
func GetBinaryStr(number any) string {
	switch v := number.(type) {
	case int:
		return strconv.FormatInt(int64(v), 2)
	case int8:
		return fmt.Sprintf("%08b", int64(v)&0xff)
	case int16:
		return fmt.Sprintf("%016b", int64(v)&0xffff)
	case int32:
		return strconv.FormatInt(int64(v), 2)
	case int64:
		return strconv.FormatInt(v, 2)
	case uint, uint8, uint16, uint32, uint64:
		return fmt.Sprintf("%b", v)
	case float32:
		return fmt.Sprintf("%032b", math.Float32bits(v))
	case float64:
		return fmt.Sprintf("%064b", math.Float64bits(v))
	case *big.Int:
		if v == nil {
			return ""
		}
		return v.Text(2)
	default:
		return ""
	}
}

// BinaryToInt parses a binary string into int.
func BinaryToInt(binaryStr string) (int, error) {
	v, err := strconv.ParseInt(binaryStr, 2, 0)
	return int(v), err
}

// BinaryToLong parses a binary string into int64.
func BinaryToLong(binaryStr string) (int64, error) { return strconv.ParseInt(binaryStr, 2, 64) }

// Compare returns -1, 0, or 1 according to x and y ordering.
func Compare[T Ordered](x, y T) int {
	if x < y {
		return -1
	}
	if x > y {
		return 1
	}
	return 0
}

// IsGreater reports whether a > b.
func IsGreater[T Ordered](a, b T) bool { return a > b }

// IsGreaterOrEqual reports whether a >= b.
func IsGreaterOrEqual[T Ordered](a, b T) bool { return a >= b }

// IsLess reports whether a < b.
func IsLess[T Ordered](a, b T) bool { return a < b }

// IsLessOrEqual reports whether a <= b.
func IsLessOrEqual[T Ordered](a, b T) bool { return a <= b }

// IsIn reports whether value is within [minInclude, maxInclude].
func IsIn[T Ordered](value, minInclude, maxInclude T) bool {
	return value >= minInclude && value <= maxInclude
}

// Equals compares two floats using a fixed 1e-9 tolerance.
func Equals(a, b float64) bool { return math.Abs(a-b) < 1e-9 }

// EqualsExact compares two floats by IEEE-754 bits.
func EqualsExact(a, b float64) bool { return math.Float64bits(a) == math.Float64bits(b) }

// EqualsFloat32Exact compares two float32 values by IEEE-754 bits.
func EqualsFloat32Exact(a, b float32) bool { return math.Float32bits(a) == math.Float32bits(b) }

// EqualsInt64 compares int64 values.
func EqualsInt64(a, b int64) bool { return a == b }

// EqualsBigDecimal compares decimal values by numeric value.
func EqualsBigDecimal(a, b *big.Rat) bool {
	if a == b {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return a.Cmp(b) == 0
}

// EqualsChar compares two runes with optional case-insensitive mode.
func EqualsChar(c1, c2 rune, ignoreCase bool) bool {
	if ignoreCase {
		return unicode.ToLower(c1) == unicode.ToLower(c2)
	}
	return c1 == c2
}

// Min returns the minimum value, or the zero value when no values are provided.
func Min[T Ordered](nums ...T) T {
	if len(nums) == 0 {
		var zero T
		return zero
	}
	m := nums[0]
	for _, v := range nums[1:] {
		if v < m {
			m = v
		}
	}
	return m
}

// Max returns the maximum value, or the zero value when no values are provided.
func Max[T Ordered](nums ...T) T {
	if len(nums) == 0 {
		var zero T
		return zero
	}
	m := nums[0]
	for _, v := range nums[1:] {
		if v > m {
			m = v
		}
	}
	return m
}

// Sum returns the sum of all values.
func Sum[T Number](nums ...T) T {
	var s T
	for _, v := range nums {
		s += v
	}
	return s
}

// Avg returns the arithmetic mean as float64, or 0 for empty input.
func Avg[T Number](nums ...T) float64 {
	if len(nums) == 0 {
		return 0
	}
	var s float64
	for _, v := range nums {
		s += float64(v)
	}
	return s / float64(len(nums))
}

// ToStr converts a float64 to string and strips trailing fractional zeros.
func ToStr(number float64) string { return ToStrStrip(number, true) }

// ToStrDefault converts a pointer to string or returns defaultValue when nil.
func ToStrDefault(number *float64, defaultValue string) string {
	if number == nil {
		return defaultValue
	}
	return ToStr(*number)
}

// ToStrStrip converts number to string and optionally strips trailing zeros.
func ToStrStrip(number float64, strip bool) string {
	if !IsValid(number) {
		return ""
	}
	s := strconv.FormatFloat(number, 'f', -1, 64)
	if strip {
		s = stripTrailingZeros(s)
	}
	return s
}

// ToBigDecimal parses a decimal string into a rational number. Blank input returns 0.
func ToBigDecimal(numberStr string) *big.Rat {
	s := strings.TrimSpace(strings.ReplaceAll(numberStr, ",", ""))
	if s == "" {
		return new(big.Rat)
	}
	r := new(big.Rat)
	if _, ok := r.SetString(s); ok {
		return r
	}
	f, _ := strconv.ParseFloat(s, 64)
	return ratFromFloat(f)
}

// ToBigInteger parses an integer string. Blank input returns 0.
func ToBigInteger(number string) *big.Int {
	s := strings.TrimSpace(number)
	if s == "" {
		return big.NewInt(0)
	}
	if i, ok := new(big.Int).SetString(s, 10); ok {
		return i
	}
	return big.NewInt(ParseLong(s))
}

// Count returns how many parts of size part are needed for total.
func Count(total, part int) int {
	if total == 0 || part == 0 {
		return 0
	}
	return (total-1)/part + 1
}

// Null2Zero returns 0 when decimal is nil.
func Null2Zero(decimal *big.Rat) *big.Rat { return NullBigDecimalToZero(decimal) }

// Zero2One returns 1 when value is 0, otherwise value.
func Zero2One(value int) int {
	if value == 0 {
		return 1
	}
	return value
}

// NullToZero returns the zero value when number is nil.
func NullToZero[T Number](number *T) T {
	if number == nil {
		var zero T
		return zero
	}
	return *number
}

// NullBigIntToZero returns 0 when number is nil.
func NullBigIntToZero(number *big.Int) *big.Int {
	if number == nil {
		return big.NewInt(0)
	}
	return number
}

// NullBigDecimalToZero returns 0 when number is nil.
func NullBigDecimalToZero(number *big.Rat) *big.Rat {
	if number == nil {
		return new(big.Rat)
	}
	return number
}

// NewBigInteger creates a BigInteger from decimal, hex (0x/#), or octal strings.
func NewBigInteger(str string) (*big.Int, bool) {
	s := strings.TrimSpace(str)
	if s == "" {
		return nil, false
	}
	negate := false
	if strings.HasPrefix(s, "-") {
		negate = true
		s = s[1:]
	}
	radix := 10
	switch {
	case strings.HasPrefix(strings.ToLower(s), "0x"):
		radix, s = 16, s[2:]
	case strings.HasPrefix(s, "#"):
		radix, s = 16, s[1:]
	case strings.HasPrefix(s, "0") && len(s) > 1:
		radix, s = 8, s[1:]
	}
	i, ok := new(big.Int).SetString(s, radix)
	if !ok {
		return nil, false
	}
	if negate {
		i.Neg(i)
	}
	return i, true
}

// IsBeside reports whether two integers differ by 1.
func IsBeside[T ~int | ~int64](number1, number2 T) bool {
	if number1 > number2 {
		return number1-number2 == 1
	}
	return number2-number1 == 1
}

// PartValue splits total into partCount parts and rounds up when there is a remainder.
func PartValue(total, partCount int) int { return PartValueWithMode(total, partCount, true) }

// PartValueWithMode splits total into partCount parts.
func PartValueWithMode(total, partCount int, plusOneWhenHasRem bool) int {
	if partCount == 0 {
		return 0
	}
	part := total / partCount
	if plusOneWhenHasRem && total%partCount > 0 {
		part++
	}
	return part
}

// Pow raises number to n. Negative exponents are rounded to two decimal places.
func Pow(number float64, n int) float64 { return PowWithMode(number, n, 2, RoundHalfUp) }

// PowWithMode raises number to n and applies scale/mode for negative exponents.
func PowWithMode(number float64, n, scale int, mode RoundingMode) float64 {
	if n < 0 {
		return RoundMode(1/math.Pow(number, float64(-n)), scale, mode)
	}
	return math.Pow(number, float64(n))
}

// IsPowerOfTwo reports whether n is a positive power of two.
func IsPowerOfTwo(n int64) bool { return n > 0 && (n&(n-1)) == 0 }

// ParseInt parses an int with tolerant handling for blank, hex, and decimal fractions.
func ParseInt(number string) int {
	s := strings.TrimSpace(number)
	if s == "" || strings.HasPrefix(s, ".") {
		return 0
	}
	if strings.HasPrefix(strings.ToLower(s), "0x") {
		v, _ := strconv.ParseInt(s[2:], 16, 0)
		return int(v)
	}
	if strings.ContainsAny(s, "eE") {
		return 0
	}
	if i, err := strconv.Atoi(s); err == nil {
		return i
	}
	f, _ := strconv.ParseFloat(strings.ReplaceAll(s, ",", ""), 64)
	return int(f)
}

// ParseLong parses an int64 with tolerant handling for blank, hex, and decimal fractions.
func ParseLong(number string) int64 {
	s := strings.TrimSpace(number)
	if s == "" || strings.HasPrefix(s, ".") {
		return 0
	}
	if strings.HasPrefix(strings.ToLower(s), "0x") {
		v, _ := strconv.ParseInt(s[2:], 16, 64)
		return v
	}
	if i, err := strconv.ParseInt(s, 10, 64); err == nil {
		return i
	}
	f, _ := strconv.ParseFloat(strings.ReplaceAll(s, ",", ""), 64)
	return int64(f)
}

// ParseFloat parses a float32. Blank input returns 0.
func ParseFloat(number string) float32 { return float32(ParseDouble(number)) }

// ParseDouble parses a float64. Blank input returns 0.
func ParseDouble(number string) float64 {
	s := strings.TrimSpace(strings.ReplaceAll(number, ",", ""))
	if s == "" {
		return 0
	}
	f, _ := strconv.ParseFloat(s, 64)
	return f
}

// ParseNumber parses a numeric string as float64.
func ParseNumber(numberStr string) (float64, error) {
	s := strings.TrimSpace(strings.ReplaceAll(numberStr, ",", ""))
	if strings.HasPrefix(strings.ToLower(s), "0x") {
		v, err := strconv.ParseInt(s[2:], 16, 64)
		return float64(v), err
	}
	s = strings.TrimPrefix(s, "+")
	return strconv.ParseFloat(s, 64)
}

// ParseIntDefault parses an int or returns defaultValue on failure.
func ParseIntDefault(numberStr string, defaultValue int) int {
	if strings.TrimSpace(numberStr) == "" {
		return defaultValue
	}
	if !IsNumber(numberStr) && !strings.Contains(numberStr, ",") {
		return defaultValue
	}
	return ParseInt(numberStr)
}

// ParseLongDefault parses an int64 or returns defaultValue on failure.
func ParseLongDefault(numberStr string, defaultValue int64) int64 {
	if strings.TrimSpace(numberStr) == "" {
		return defaultValue
	}
	if !IsNumber(numberStr) && !strings.Contains(numberStr, ",") {
		return defaultValue
	}
	return ParseLong(numberStr)
}

// ParseFloatDefault parses a float32 or returns defaultValue on failure.
func ParseFloatDefault(numberStr string, defaultValue float32) float32 {
	if strings.TrimSpace(numberStr) == "" {
		return defaultValue
	}
	if f, err := strconv.ParseFloat(strings.ReplaceAll(numberStr, ",", ""), 32); err == nil {
		return float32(f)
	}
	return defaultValue
}

// ParseDoubleDefault parses a float64 or returns defaultValue on failure.
func ParseDoubleDefault(numberStr string, defaultValue float64) float64 {
	if strings.TrimSpace(numberStr) == "" {
		return defaultValue
	}
	if f, err := strconv.ParseFloat(strings.ReplaceAll(numberStr, ",", ""), 64); err == nil {
		return f
	}
	return defaultValue
}

// ToBytes converts int32 to big-endian bytes.
func ToBytes(value int32) []byte {
	out := make([]byte, 4)
	binary.BigEndian.PutUint32(out, uint32(value)) // #nosec G115 -- preserve the exact two's-complement bit pattern.
	return out
}

// ToInt converts four big-endian bytes to int32.
func ToInt(bytes []byte) int32 {
	if len(bytes) < 4 {
		return 0
	}
	return int32(binary.BigEndian.Uint32(bytes[:4])) // #nosec G115 -- preserve the exact two's-complement bit pattern.
}

// ToUnsignedByteArray returns the unsigned big-endian byte representation of value.
func ToUnsignedByteArray(value *big.Int) []byte {
	if value == nil {
		return nil
	}
	return value.Bytes()
}

// ToUnsignedByteArrayLen returns value padded to length bytes.
func ToUnsignedByteArrayLen(length int, value *big.Int) ([]byte, error) {
	bytes := ToUnsignedByteArray(value)
	if len(bytes) > length {
		return nil, errors.New("standard length exceeded for value")
	}
	out := make([]byte, length)
	copy(out[length-len(bytes):], bytes)
	return out, nil
}

// FromUnsignedByteArray converts unsigned big-endian bytes to a big integer.
func FromUnsignedByteArray(buf []byte) *big.Int { return new(big.Int).SetBytes(buf) }

// FromUnsignedByteArrayRange converts a sub-slice of unsigned big-endian bytes to a big integer.
func FromUnsignedByteArrayRange(buf []byte, off, length int) *big.Int {
	if off < 0 || length < 0 || off+length > len(buf) {
		return big.NewInt(0)
	}
	return new(big.Int).SetBytes(buf[off : off+length])
}

// IsValidNumber reports whether number is a finite float64.
func IsValidNumber(number any) bool {
	switch v := number.(type) {
	case float64:
		return IsValid(v)
	case float32:
		return IsValidFloat32(v)
	case nil:
		return false
	default:
		return true
	}
}

// IsValid reports whether number is neither NaN nor infinite.
func IsValid(number float64) bool { return !math.IsNaN(number) && !math.IsInf(number, 0) }

// IsValidFloat32 reports whether number is neither NaN nor infinite.
func IsValidFloat32(number float32) bool {
	return !math.IsNaN(float64(number)) && !math.IsInf(float64(number), 0)
}

// Calculate evaluates a simple arithmetic expression supporting +, -, *, /, %, and parentheses.
func Calculate(expression string) (float64, error) { return evalExpression(expression) }

// ToDouble converts numeric values to float64 while preserving float32 textual precision.
func ToDouble(value any) float64 {
	switch v := value.(type) {
	case float32:
		f, _ := strconv.ParseFloat(strconv.FormatFloat(float64(v), 'f', -1, 32), 64)
		return f
	case float64:
		return v
	case int:
		return float64(v)
	case int64:
		return float64(v)
	case uint64:
		return float64(v)
	default:
		return 0
	}
}

// IsOdd reports whether num is odd.
func IsOdd(num int) bool { return num&1 == 1 }

// IsEven reports whether num is even.
func IsEven(num int) bool { return !IsOdd(num) }

// Generic numeric constraints.

// Number is the set of supported numeric types.
type Number interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 |
		~float32 | ~float64
}

// Ordered is the set of supported ordered types.
type Ordered interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 |
		~float32 | ~float64 | ~string
}

func foldRat(op func(*big.Rat, *big.Rat, *big.Rat) *big.Rat, values ...string) *big.Rat {
	if len(values) == 0 {
		return new(big.Rat)
	}
	r := new(big.Rat)
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			op(r, r, ToBigDecimal(v))
		}
	}
	return r
}

func ratFromFloat(v float64) *big.Rat {
	r := new(big.Rat)
	r.SetString(strconv.FormatFloat(v, 'f', -1, 64))
	return r
}

func ratToFloat(r *big.Rat) float64 {
	f, _ := r.Float64()
	return f
}

func stripTrailingZeros(s string) string {
	if strings.Contains(s, ".") && !strings.ContainsAny(s, "eE") {
		for strings.HasSuffix(s, "0") {
			s = strings.TrimSuffix(s, "0")
		}
		s = strings.TrimSuffix(s, ".")
	}
	return s
}

func addThousands(s string) string {
	parts := strings.SplitN(s, ".", 2)
	intPart := parts[0]
	sign := ""
	if strings.HasPrefix(intPart, "-") {
		sign, intPart = "-", intPart[1:]
	}
	for i := len(intPart) - 3; i > 0; i -= 3 {
		intPart = intPart[:i] + "," + intPart[i:]
	}
	if len(parts) == 2 {
		return sign + intPart + "." + parts[1]
	}
	return sign + intPart
}

func absInt(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func mathSubNode(selectNum, minNum int) int {
	if selectNum == minNum {
		return 1
	}
	return selectNum * mathSubNode(selectNum-1, minNum)
}

func mathNode(selectNum int) int {
	if selectNum == 0 {
		return 1
	}
	return selectNum * mathNode(selectNum-1)
}

func evalExpression(expr string) (float64, error) {
	p := expressionParser{s: expr}
	v, err := p.parseExpression()
	if err != nil {
		return 0, err
	}
	p.skipSpaces()
	if p.pos != len(p.s) {
		return 0, fmt.Errorf("unexpected token at %d", p.pos)
	}
	return v, nil
}

type expressionParser struct {
	s   string
	pos int
}

func (p *expressionParser) parseExpression() (float64, error) {
	v, err := p.parseTerm()
	if err != nil {
		return 0, err
	}
	for {
		p.skipSpaces()
		switch {
		case p.match('+'):
			r, err := p.parseTerm()
			if err != nil {
				return 0, err
			}
			v += r
		case p.match('-'):
			r, err := p.parseTerm()
			if err != nil {
				return 0, err
			}
			v -= r
		default:
			return v, nil
		}
	}
}

func (p *expressionParser) parseTerm() (float64, error) {
	v, err := p.parseFactor()
	if err != nil {
		return 0, err
	}
	for {
		p.skipSpaces()
		switch {
		case p.match('*'):
			r, err := p.parseFactor()
			if err != nil {
				return 0, err
			}
			v *= r
		case p.match('/'):
			r, err := p.parseFactor()
			if err != nil {
				return 0, err
			}
			v /= r
		case p.match('%'):
			r, err := p.parseFactor()
			if err != nil {
				return 0, err
			}
			v = math.Mod(v, r)
		default:
			return v, nil
		}
	}
}

func secureIntn(max int) int {
	return secureIntnWithReader(cryptorand.Reader, max)
}

func secureIntnWithReader(reader io.Reader, max int) int {
	if max <= 0 {
		return 0
	}
	if reader == nil {
		reader = cryptorand.Reader
	}
	n, err := cryptorand.Int(reader, big.NewInt(int64(max)))
	if err != nil {
		return 0
	}
	return int(n.Int64())
}

func (p *expressionParser) parseFactor() (float64, error) {
	p.skipSpaces()
	if p.match('+') {
		return p.parseFactor()
	}
	if p.match('-') {
		v, err := p.parseFactor()
		return -v, err
	}
	if p.match('(') {
		v, err := p.parseExpression()
		if err != nil {
			return 0, err
		}
		if !p.match(')') {
			return 0, errors.New("missing closing parenthesis")
		}
		return v, nil
	}
	start := p.pos
	for p.pos < len(p.s) && (unicode.IsDigit(rune(p.s[p.pos])) || p.s[p.pos] == '.') {
		p.pos++
	}
	if start == p.pos {
		return 0, fmt.Errorf("expected number at %d", p.pos)
	}
	return strconv.ParseFloat(p.s[start:p.pos], 64)
}

func (p *expressionParser) skipSpaces() {
	for p.pos < len(p.s) && unicode.IsSpace(rune(p.s[p.pos])) {
		p.pos++
	}
}

func (p *expressionParser) match(ch byte) bool {
	p.skipSpaces()
	if p.pos < len(p.s) && p.s[p.pos] == ch {
		p.pos++
		return true
	}
	return false
}
