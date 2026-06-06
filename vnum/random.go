package vnum

import (
	"io"

	numimpl "github.com/imajinyun/go-knifer/internal/num"
)

func WithRandomReader(reader io.Reader) RandomNumberOption {
	return numimpl.WithRandomReader(reader)
}

func GenerateRandomNumber(begin, end, size int) []int {
	return numimpl.GenerateRandomNumber(begin, end, size)
}

func GenRandomNumberWithOptions(begin, end, size int, opts ...RandomNumberOption) []int {
	return numimpl.GenRandomNumberWithOptions(begin, end, size, opts...)
}

func GenerateRandomNumberWithSeed(begin, end, size int, seed []int) []int {
	return numimpl.GenerateRandomNumberWithSeed(begin, end, size, seed)
}

func GenRandomNumberWithSeedWithOptions(begin, end, size int, seed []int, opts ...RandomNumberOption) []int {
	return numimpl.GenRandomNumberWithSeedWithOptions(begin, end, size, seed, opts...)
}

func GenerateBySet(begin, end, size int) []int { return numimpl.GenerateBySet(begin, end, size) }

func GenBySetWithOptions(begin, end, size int, opts ...RandomNumberOption) []int {
	return numimpl.GenBySetWithOptions(begin, end, size, opts...)
}
