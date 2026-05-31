package vnum

import (
	"math/big"

	numimpl "github.com/imajinyun/go-knifer/internal/num"
)

type (
	Number       = numimpl.Number
	Ordered      = numimpl.Ordered
	RoundingMode = numimpl.RoundingMode
)

const (
	RoundHalfUp   = numimpl.RoundHalfUp
	RoundHalfEven = numimpl.RoundHalfEven
	RoundDown     = numimpl.RoundDown
)

func Add(values ...float64) float64    { return numimpl.Add(values...) }
func AddStr(values ...string) *big.Rat { return numimpl.AddStr(values...) }
func Sub(values ...float64) float64    { return numimpl.Sub(values...) }
func SubStr(values ...string) *big.Rat { return numimpl.SubStr(values...) }
func Mul(values ...float64) float64    { return numimpl.Mul(values...) }
func MulStr(values ...string) *big.Rat { return numimpl.MulStr(values...) }
func Div(a, b float64, scale ...int) float64 {
	if len(scale) == 0 {
		return numimpl.Div(a, b)
	}
	return numimpl.NumberDiv(a, b, scale[0])
}
func DivWithMode(a, b float64, scale int, mode RoundingMode) float64 {
	return numimpl.DivWithMode(a, b, scale, mode)
}
func CeilDiv(v1, v2 int) int             { return numimpl.CeilDiv(v1, v2) }
func Round(v float64, scale int) float64 { return numimpl.Round(v, scale) }
func RoundMode(v float64, scale int, mode RoundingMode) float64 {
	return numimpl.RoundMode(v, scale, mode)
}
func RoundStr(v float64, scale int) string            { return numimpl.RoundStr(v, scale) }
func RoundHalfEvenFloat(v float64, scale int) float64 { return numimpl.RoundHalfEvenFloat(v, scale) }
func RoundDownFloat(v float64, scale int) float64     { return numimpl.RoundDownFloat(v, scale) }
func DecimalFormat(format string, v float64) string   { return numimpl.DecimalFormat(format, v) }
func DecimalFormatMoney(v float64) string             { return numimpl.DecimalFormatMoney(v) }
func FormatPercent(number float64, scale int) string  { return numimpl.FormatPercent(number, scale) }
func IsNumber(s string) bool                          { return numimpl.IsNumber(s) }
func IsInteger(s string) bool                         { return numimpl.IsInteger(s) }
func IsLong(s string) bool                            { return numimpl.IsLong(s) }
func IsDouble(s string) bool                          { return numimpl.IsDouble(s) }
func IsDigits(s string) bool                          { return numimpl.IsDigits(s) }
func IsPrimes(n int) bool                             { return numimpl.IsPrimes(n) }
func GenerateRandomNumber(begin, end, size int) []int {
	return numimpl.GenerateRandomNumber(begin, end, size)
}
func GenerateRandomNumberWithSeed(begin, end, size int, seed []int) []int {
	return numimpl.GenerateRandomNumberWithSeed(begin, end, size, seed)
}
func GenerateBySet(begin, end, size int) []int { return numimpl.GenerateBySet(begin, end, size) }
func Range(start, end, step int) []int         { return numimpl.Range(start, end, step) }
func RangeClosed(start, stop, step int) []int  { return numimpl.RangeClosed(start, stop, step) }
func AppendRange(start, stop, step int, values []int) []int {
	return numimpl.AppendRange(start, stop, step, values)
}
func Factorial(n uint64) (uint64, error)               { return numimpl.Factorial(n) }
func FactorialRange(start, end uint64) (uint64, error) { return numimpl.FactorialRange(start, end) }
func FactorialBig(n *big.Int) *big.Int                 { return numimpl.FactorialBig(n) }
func FactorialBigRange(start, end *big.Int) *big.Int   { return numimpl.FactorialBigRange(start, end) }
func Sqrt(x uint64) uint64                             { return numimpl.Sqrt(x) }
func ProcessMultiple(selectNum, minNum int) int        { return numimpl.ProcessMultiple(selectNum, minNum) }
func Divisor(m, n int) int                             { return numimpl.Divisor(m, n) }
func Multiple(m, n int) int                            { return numimpl.Multiple(m, n) }
func GetBinaryStr(number any) string                   { return numimpl.GetBinaryStr(number) }
func BinaryToInt(binaryStr string) (int, error)        { return numimpl.BinaryToInt(binaryStr) }
func BinaryToLong(binaryStr string) (int64, error)     { return numimpl.BinaryToLong(binaryStr) }
func Compare[T Ordered](x, y T) int                    { return numimpl.Compare(x, y) }
func IsGreater[T Ordered](a, b T) bool                 { return numimpl.IsGreater(a, b) }
func IsGreaterOrEqual[T Ordered](a, b T) bool          { return numimpl.IsGreaterOrEqual(a, b) }
func IsLess[T Ordered](a, b T) bool                    { return numimpl.IsLess(a, b) }
func IsLessOrEqual[T Ordered](a, b T) bool             { return numimpl.IsLessOrEqual(a, b) }
func IsIn[T Ordered](value, minInclude, maxInclude T) bool {
	return numimpl.IsIn(value, minInclude, maxInclude)
}
func Equals(a, b float64) bool                     { return numimpl.Equals(a, b) }
func EqualsExact(a, b float64) bool                { return numimpl.EqualsExact(a, b) }
func EqualsFloat32Exact(a, b float32) bool         { return numimpl.EqualsFloat32Exact(a, b) }
func EqualsInt64(a, b int64) bool                  { return numimpl.EqualsInt64(a, b) }
func EqualsBigDecimal(a, b *big.Rat) bool          { return numimpl.EqualsBigDecimal(a, b) }
func EqualsChar(c1, c2 rune, ignoreCase bool) bool { return numimpl.EqualsChar(c1, c2, ignoreCase) }
func Min[T Ordered](nums ...T) T                   { return numimpl.Min(nums...) }
func Max[T Ordered](nums ...T) T                   { return numimpl.Max(nums...) }
func Sum[T Number](nums ...T) T                    { return numimpl.Sum(nums...) }
func Avg[T Number](nums ...T) float64              { return numimpl.Avg(nums...) }
func ToStr(number float64) string                  { return numimpl.ToStr(number) }
func ToStrDefault(number *float64, defaultValue string) string {
	return numimpl.ToStrDefault(number, defaultValue)
}
func ToStrStrip(number float64, stripTrailingZeros bool) string {
	return numimpl.ToStrStrip(number, stripTrailingZeros)
}
func ToBigDecimal(numberStr string) *big.Rat            { return numimpl.ToBigDecimal(numberStr) }
func ToBigInteger(number string) *big.Int               { return numimpl.ToBigInteger(number) }
func Count(total, part int) int                         { return numimpl.Count(total, part) }
func Null2Zero(decimal *big.Rat) *big.Rat               { return numimpl.Null2Zero(decimal) }
func Zero2One(value int) int                            { return numimpl.Zero2One(value) }
func NullToZero[T Number](number *T) T                  { return numimpl.NullToZero(number) }
func NullBigIntToZero(number *big.Int) *big.Int         { return numimpl.NullBigIntToZero(number) }
func NullBigDecimalToZero(number *big.Rat) *big.Rat     { return numimpl.NullBigDecimalToZero(number) }
func NewBigInteger(str string) (*big.Int, bool)         { return numimpl.NewBigInteger(str) }
func IsBeside[T ~int | ~int64](number1, number2 T) bool { return numimpl.IsBeside(number1, number2) }
func PartValue(total, partCount int) int                { return numimpl.PartValue(total, partCount) }
func PartValueWithMode(total, partCount int, plusOneWhenHasRem bool) int {
	return numimpl.PartValueWithMode(total, partCount, plusOneWhenHasRem)
}
func Pow(number float64, n int) float64 { return numimpl.Pow(number, n) }
func PowWithMode(number float64, n, scale int, mode RoundingMode) float64 {
	return numimpl.PowWithMode(number, n, scale, mode)
}
func IsPowerOfTwo(n int64) bool                     { return numimpl.IsPowerOfTwo(n) }
func ParseInt(number string) int                    { return numimpl.ParseInt(number) }
func ParseLong(number string) int64                 { return numimpl.ParseLong(number) }
func ParseFloat(number string) float32              { return numimpl.ParseFloat(number) }
func ParseDouble(number string) float64             { return numimpl.ParseDouble(number) }
func ParseNumber(numberStr string) (float64, error) { return numimpl.ParseNumber(numberStr) }
func ParseIntDefault(numberStr string, defaultValue int) int {
	return numimpl.ParseIntDefault(numberStr, defaultValue)
}
func ParseLongDefault(numberStr string, defaultValue int64) int64 {
	return numimpl.ParseLongDefault(numberStr, defaultValue)
}
func ParseFloatDefault(numberStr string, defaultValue float32) float32 {
	return numimpl.ParseFloatDefault(numberStr, defaultValue)
}
func ParseDoubleDefault(numberStr string, defaultValue float64) float64 {
	return numimpl.ParseDoubleDefault(numberStr, defaultValue)
}
func ToBytes(value int32) []byte                { return numimpl.ToBytes(value) }
func ToInt(bytes []byte) int32                  { return numimpl.ToInt(bytes) }
func ToUnsignedByteArray(value *big.Int) []byte { return numimpl.ToUnsignedByteArray(value) }
func ToUnsignedByteArrayLen(length int, value *big.Int) ([]byte, error) {
	return numimpl.ToUnsignedByteArrayLen(length, value)
}
func FromUnsignedByteArray(buf []byte) *big.Int { return numimpl.FromUnsignedByteArray(buf) }
func FromUnsignedByteArrayRange(buf []byte, off, length int) *big.Int {
	return numimpl.FromUnsignedByteArrayRange(buf, off, length)
}
func IsValidNumber(number any) bool                { return numimpl.IsValidNumber(number) }
func IsValid(number float64) bool                  { return numimpl.IsValid(number) }
func IsValidFloat32(number float32) bool           { return numimpl.IsValidFloat32(number) }
func Calculate(expression string) (float64, error) { return numimpl.Calculate(expression) }
func ToDouble(value any) float64                   { return numimpl.ToDouble(value) }
func IsOdd(num int) bool                           { return numimpl.IsOdd(num) }
func IsEven(num int) bool                          { return numimpl.IsEven(num) }
