package vnum

import (
	"io"

	numimpl "github.com/imajinyun/knifer-go/internal/num"
)

func WithRandomReader(reader io.Reader) RandomNumberOption {
	return numimpl.WithRandomReader(reader)
}

func GenRandomNumber(begin, end, size int) []int {
	return numimpl.GenRandomNumber(begin, end, size)
}

func GenRandomNumberWithOptions(begin, end, size int, opts ...RandomNumberOption) []int {
	return numimpl.GenRandomNumberWithOptions(begin, end, size, opts...)
}

func GenRandomNumberWithSeed(begin, end, size int, seed []int) []int {
	return numimpl.GenRandomNumberWithSeed(begin, end, size, seed)
}

func GenRandomNumberWithSeedWithOptions(begin, end, size int, seed []int, opts ...RandomNumberOption) []int {
	return numimpl.GenRandomNumberWithSeedWithOptions(begin, end, size, seed, opts...)
}

func GenBySet(begin, end, size int) []int { return numimpl.GenBySet(begin, end, size) }

func GenBySetWithOptions(begin, end, size int, opts ...RandomNumberOption) []int {
	return numimpl.GenBySetWithOptions(begin, end, size, opts...)
}
