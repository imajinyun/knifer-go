package vnum

import (
	"math/big"

	numimpl "github.com/imajinyun/knifer-go/internal/num"
)

func GetBinaryStr(number any) string { return numimpl.GetBinaryStr(number) }

func GetBinaryStrWithOptions(number any, opts ...FormatOption) string {
	return numimpl.GetBinaryStrWithOptions(number, opts...)
}

func BinaryToInt(binaryStr string) (int, error) { return numimpl.BinaryToInt(binaryStr) }

func BinaryToIntWithOptions(binaryStr string, opts ...ParseOption) (int, error) {
	return numimpl.BinaryToIntWithOptions(binaryStr, opts...)
}

func BinaryToLong(binaryStr string) (int64, error) { return numimpl.BinaryToLong(binaryStr) }

func BinaryToLongWithOptions(binaryStr string, opts ...ParseOption) (int64, error) {
	return numimpl.BinaryToLongWithOptions(binaryStr, opts...)
}

func ToBytes(value int32) []byte { return numimpl.ToBytes(value) }

func ToInt(bytes []byte) int32 { return numimpl.ToInt(bytes) }

func ToUnsignedByteArray(value *big.Int) []byte { return numimpl.ToUnsignedByteArray(value) }

func ToUnsignedByteArrayLen(length int, value *big.Int) ([]byte, error) {
	return numimpl.ToUnsignedByteArrayLen(length, value)
}

func FromUnsignedByteArray(buf []byte) *big.Int { return numimpl.FromUnsignedByteArray(buf) }

func FromUnsignedByteArrayRange(buf []byte, off, length int) *big.Int {
	return numimpl.FromUnsignedByteArrayRange(buf, off, length)
}
