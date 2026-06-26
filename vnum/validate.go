package vnum

import numimpl "github.com/imajinyun/knifer-go/internal/num"

func IsNumber(s string) bool { return numimpl.IsNumber(s) }

func IsNumberWithOptions(s string, opts ...ParseOption) bool {
	return numimpl.IsNumberWithOptions(s, opts...)
}

func IsInteger(s string) bool { return numimpl.IsInteger(s) }

func IsIntegerWithOptions(s string, opts ...ParseOption) bool {
	return numimpl.IsIntegerWithOptions(s, opts...)
}

func IsLong(s string) bool { return numimpl.IsLong(s) }

func IsLongWithOptions(s string, opts ...ParseOption) bool {
	return numimpl.IsLongWithOptions(s, opts...)
}

func IsDouble(s string) bool { return numimpl.IsDouble(s) }

func IsDoubleWithOptions(s string, opts ...ParseOption) bool {
	return numimpl.IsDoubleWithOptions(s, opts...)
}

func IsDigits(s string) bool { return numimpl.IsDigits(s) }

func IsPrimes(n int) bool { return numimpl.IsPrimes(n) }

func IsValidNumber(number any) bool { return numimpl.IsValidNumber(number) }

func IsValid(number float64) bool { return numimpl.IsValid(number) }

func IsValidFloat32(number float32) bool { return numimpl.IsValidFloat32(number) }

func IsOdd(num int) bool { return numimpl.IsOdd(num) }

func IsEven(num int) bool { return numimpl.IsEven(num) }

func IsPowerOfTwo(n int64) bool { return numimpl.IsPowerOfTwo(n) }
